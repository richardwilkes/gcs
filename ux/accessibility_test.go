// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/equipmentsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/namegen"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/study"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/traitsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

// TestEveryControlHasAnAccessibleName opens the sheet, each editor seeded with every kind of feature and prerequisite,
// the settings views, the calculators, and representative libraries and file editors, and fails for every control in
// them that an assistive technology would be handed with no name. A field, popup or color well is named by the label
// laid out before it, by the label it points at with Accessibility.LabeledBy, or by an Accessibility.Name of its own;
// an icon-only button by its tooltip. A control with none of those is announced as "edit text" or "pop up button" and
// nothing more.
//
// The tree is walked rather than the panels, so that what a table describes without a panel apiece -- its column
// headers and the cells of its rows -- is checked along with everything else.
func TestEveryControlHasAnAccessibleName(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	audit := &axNameAudit{AXNameAudit: uxtest.NewAXNameAudit(t, screen, wnd), t: t, screen: screen}
	audit.Check("workspace", wnd.Content())

	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var entity *gurps.Entity
	var eq *gurps.Equipment
	screen.Do(func() {
		entity = sheet.Entity()
		// Give the sheet's lists a row apiece so that their cells are described.
		tr := gurps.NewTrait(entity, nil, false)
		tr.Name = "Audit Trait"
		entity.Traits = append(entity.Traits, tr)
		sk := gurps.NewSkill(entity, nil, false)
		sk.Name = "Audit Skill"
		entity.Skills = append(entity.Skills, sk)
		eq = gurps.NewEquipment(entity, nil, false)
		eq.Name = "Audit Equipment"
		eq.BaseValue = "10"
		entity.CarriedEquipment = append(entity.CarriedEquipment, eq)
		entity.Recalculate()
		sheet.Rebuild(true)
	})
	audit.Check("character sheet", sheet)
	screen.Do(func() { sheet.toggleLayoutEditing() })
	audit.Check("character sheet layout editing", sheet)
	screen.Do(func() { sheet.toggleLayoutEditing() })
	var buy *buyDockable
	screen.Do(func() {
		buy = showBuyDockable(sheet)
		buy.addItem(eq.Clone(gurps.LibraryFile{}, entity, nil, gurps.Duplicate))
		buy.Sync()
	})
	audit.Check("buy dockable", buy)

	// The dialogs run a nested modal loop, so each is posted rather than run through Do, checked in its own window,
	// and canceled.
	checkDialog := func(view string, open func()) {
		if !screen.Post(open) {
			t.Fatalf("%s: the dialog could not be opened", view)
		}
		screen.Sync()
		dialogWnd, dialog := modalDialog(t, screen, wnd)
		dialogAudit := uxtest.NewAXNameAudit(t, screen, dialogWnd)
		dialogAudit.Check(view, dialogWnd.Content())
		dialogAudit.Report()
		screen.Do(func() { dialog.StopModal(unison.ModalResponseCancel) })
		screen.Sync()
	}
	checkDialog("purchase dialog", func() { confirmPurchaseWithDialog(buy, &buy.plan) })
	screen.Do(func() { buy.discardAndClose() })
	checkDialog("sell dialog", func() { showSellDialog(sheet, []*gurps.Equipment{eq}) })
	audit.CheckOpened("shopping record", func() {
		entity.AddShoppingRecord(&gurps.ShoppingRecordEntry{
			When:      jio.Now(),
			Name:      "Audit Sale",
			Kind:      gurps.ShoppingSold,
			Items:     []gurps.ShoppingRecordItem{{Name: "Audit Equipment", Quantity: fxp.One, Value: fxp.Ten}},
			Money:     []gurps.ShoppingRecordItem{{Name: "Audit Coin", Quantity: fxp.Five, Value: fxp.One}},
			Exchanged: fxp.Five,
		})
		showShoppingRecord(sheet)
	})

	// Every feature and prerequisite type has a row of its own in the editors, so each editor is seeded with one of
	// every kind it can hold.
	allFeatures := func(owner fmt.Stringer, forEquipmentModifier bool) gurps.Features {
		var list gurps.Features
		fp := newFeaturesPanel(entity, owner, &gurps.Features{}, forEquipmentModifier)
		for _, ft := range feature.Types {
			if ft == feature.Unknown {
				continue
			}
			if f := fp.createFeatureForType(ft); f != nil {
				list = append(list, f)
			}
		}
		seedEveryFeatureControl(list)
		return list
	}
	allPrereqs := func(types []prereq.Type, ownerIsSpell bool) *gurps.PrereqList {
		root := gurps.NewPrereqList()
		pp := newPrereqPanel(entity, &root, types, ownerIsSpell)
		for _, pt := range types {
			if pr := pp.createPrereqForType(pt, root); pr != nil {
				root.Prereqs = append(root.Prereqs, pr)
			}
		}
		seedEveryPrereqControl(root)
		return root
	}
	studies := func() []*gurps.Study {
		return []*gurps.Study{{Type: study.Self, Hours: fxp.Ten, Note: "audit"}}
	}

	audit.checkRows("trait editor", func() {
		tr := gurps.NewTrait(entity, nil, false)
		tr.Name = "Audit Trait (@style@, @lang@, @Weapon: Sword, Axe, etc.@)"
		tr.Features = allFeatures(tr, false)
		// One nameable with every chip in use, one code with no nameable, one marker that defines itself, and a
		// modifier code linked to the trait's.
		tr.Nameables = map[string]nameable.Detail{"style": {
			Label: "Style", Tooltip: "Which style", Options: []string{"Judo", "Karate"}, FreeForm: true, NotBlank: true,
		}}
		modifier := gurps.NewTraitModifier(entity, nil, false)
		modifier.Name = "Practiced @art@"
		modifier.Nameables = map[string]nameable.Detail{"art": {Label: "Art", FreeForm: true}}
		tr.Modifiers = []*gurps.TraitModifier{modifier}
		tr.NameableLinks = map[tid.TID]map[string]string{modifier.TID: {"art": "style"}}
		tr.Prereq = allPrereqs(prereq.TypesForNonEquipment, false)
		tr.Study = studies()
		tr.Weapons = []*gurps.Weapon{gurps.NewWeapon(tr, true), gurps.NewWeapon(tr, false)}
		EditTrait(sheet, tr)
	})
	audit.checkRows("trait modifier editor", func() {
		m := gurps.NewTraitModifier(entity, nil, false)
		m.Features = allFeatures(m, false)
		EditTraitModifier(sheet, m)
	})
	audit.checkRows("skill editor", func() {
		s := gurps.NewSkill(entity, nil, false)
		s.Prereq = allPrereqs(prereq.TypesForNonEquipment, false)
		s.Features = allFeatures(s, false)
		s.Study = studies()
		s.Defaults = []*gurps.SkillDefault{{DefaultType: gurps.DexterityID}, {
			DefaultType:    gurps.SkillID,
			Name:           criteria.Text{Compare: criteria.IsText, Qualifier: "Audit"},
			Specialization: criteria.Text{Compare: criteria.IsText, Qualifier: "Audit"},
			Tags:           criteria.Text{Compare: criteria.IsText, Qualifier: "Audit"},
			WhenTL:         criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.Three},
		}, {DefaultType: gurps.ParryID}}
		EditSkill(sheet, s)
	})
	audit.CheckOpened("technique editor", func() {
		EditSkill(sheet, gurps.NewTechnique(entity, nil, "Audit Skill"))
	})
	audit.checkRows("spell editor", func() {
		s := gurps.NewSpell(entity, nil, false)
		s.Prereq = allPrereqs(prereq.TypesForNonEquipment, true)
		s.Study = studies()
		EditSpell(sheet, s)
	})
	audit.checkRows("equipment editor", func() {
		e := gurps.NewEquipment(entity, nil, false)
		e.Features = allFeatures(e, false)
		e.Prereq = allPrereqs(prereq.TypesForEquipment, false)
		EditEquipment(sheet, e, true)
	})
	audit.checkRows("equipment modifier editor", func() {
		m := gurps.NewEquipmentModifier(entity, nil, false)
		m.Features = allFeatures(m, true)
		EditEquipmentModifier(sheet, m)
	})
	audit.CheckOpened("note editor", func() { EditNote(sheet, gurps.NewNote(entity, nil, false)) })
	audit.checkRows("melee weapon editor", func() {
		w := gurps.NewWeapon(gurps.NewTrait(entity, nil, false), true)
		w.Defaults = []*gurps.SkillDefault{{DefaultType: gurps.SkillID}}
		EditWeapon(sheet, w)
	})
	// With no defaults, so that the empty defaults panel's placeholder is checked.
	audit.CheckOpened("ranged weapon editor", func() {
		EditWeapon(sheet, gurps.NewWeapon(gurps.NewTrait(entity, nil, false), false))
	})
	audit.CheckOpened("points editor", func() {
		entity.PointsRecord = append(entity.PointsRecord, &gurps.PointsRecord{
			When:   jio.Now(),
			Points: fxp.Five,
			Reason: "audit",
		})
		displayPointsEditor(sheet, entity)
	})

	audit.CheckOpened("sheet settings", func() {
		// A sell money entry has a row, with its remove button.
		coin := gurps.NewEquipment(nil, nil, false)
		coin.Name = "Audit Coin"
		coin.BaseValue = "1"
		coin.Tags = []string{gurps.MoneyTag}
		entity.SheetSettings.SellMoney = []*gurps.Equipment{coin}
		// A skill level cap has a row, with its sentence and remove button.
		levelCap := gurps.NewSkillLevelCap()
		levelCap.Cap.Children = gurps.LevelCaps{gurps.NewFixedLevelCap(levelCap.Cap, fxp.Ten)}
		entity.SheetSettings.SkillLevelCaps = gurps.SkillLevelCaps{levelCap}
		ShowSheetSettings(sheet)
	})
	// The attribute rows start out collapsed to their headers, so they are checked again expanded, with their fields
	// and the features of their thresholds.
	if d, isOne := audit.Open(func() { ShowAttributeSettings(sheet) }).(*attributeSettingsDockable); isOne {
		audit.Check("attribute settings", d)
		screen.Do(d.toggleAllExpanded)
		audit.Check("attribute settings, expanded", d)
	}
	audit.CheckOpened("body settings", func() { ShowBodySettings(sheet) })
	audit.CheckOpened("general settings", ShowGeneralSettings)
	audit.CheckOpened("color settings", ShowColorSettings)
	audit.CheckOpened("font settings", ShowFontSettings)
	audit.CheckOpened("menu key settings", ShowMenuKeySettings)
	audit.CheckOpened("page reference mappings", ShowPageRefMappings)
	audit.CheckOpened("saved filters", ShowListFilterSettings)
	audit.CheckOpened("library settings", func() { ShowLibrarySettings(gurps.GlobalSettings().Libraries.User()) })

	audit.CheckOpened("character template", func() { newCharacterTemplateAction.Execute(nil) })
	audit.CheckOpened("loot sheet", func() { newLootSheetAction.Execute(nil) })
	audit.CheckOpened("traits library", func() { newTraitsLibraryAction.Execute(nil) })
	audit.CheckOpened("equipment library", func() { newEquipmentLibraryAction.Execute(nil) })
	audit.CheckOpened("markdown file", func() { newMarkdownFileAction.Execute(nil) })

	if d, isOne := audit.Open(func() { newAncestryAction.Execute(nil) }).(*ancestryEditorDockable); isOne {
		screen.Do(func() {
			d.model.CommonOptions.HairOptions = append(d.model.CommonOptions.HairOptions,
				&gurps.WeightedStringOption{Weight: 1, Value: "Brown"})
			d.model.GenderOptions = append(d.model.GenderOptions, &gurps.WeightedAncestryOptions{
				Weight: 1,
				Value:  &gurps.AncestryOptions{Name: "Audit"},
			})
			d.sync()
		})
		audit.Check("ancestry editor", d)
	} else {
		t.Error("the ancestry editor did not open")
	}
	if d, isOne := audit.Open(func() { newNameGeneratorAction.Execute(nil) }).(*nameGeneratorEditorDockable); isOne {
		screen.Do(func() {
			d.model.Type = namegen.Compound
			d.model.Compound = []*gurps.NameGenerator{
				{Type: namegen.Simple, Entries: []*gurps.WeightedStringOption{{Weight: 1, Value: "Audit"}}},
			}
			d.sync()
		})
		audit.Check("name generator editor", d)
	} else {
		t.Error("the name generator editor did not open")
	}

	audit.Report()
}

// seedEveryPrereqControl turns on every optional criterion of the prerequisites in the root, and adds a group with a
// tech level and nothing in it, so that an audit sees every control the panel can show.
func seedEveryPrereqControl(root *gurps.PrereqList) {
	on := criteria.Text{Compare: criteria.IsText}
	for _, one := range root.Prereqs {
		switch pr := one.(type) {
		case *gurps.TraitPrereq:
			pr.LevelCriteria.Qualifier = fxp.One
			pr.PointsCriteria.Compare = criteria.AtLeastNumber
			pr.TagsCriteria = on
			pr.NotesCriteria = on
			pr.ModifierNameCriteria = on
			pr.ModifierNotesCriteria = on
			pr.ModifierTagsCriteria = on
		case *gurps.SkillPrereq:
			pr.SpecializationCriteria = on
			pr.OptionalSpecializationCriteria = on
			pr.PointsCriteria.Compare = criteria.AtLeastNumber
			pr.TagsCriteria = on
		case *gurps.SpellPrereq:
			if !pr.SamePowerSource {
				pr.PowerSourceCriteria = on
			}
			pr.LevelCriteria.Compare = criteria.AtLeastNumber
		case *gurps.AttributePrereq:
			pr.CombinedWith = gurps.DexterityID
		case *gurps.EquippedEquipmentPrereq:
			pr.TagsCriteria = on
			pr.ModifierNameCriteria = on
			pr.ModifierNotesCriteria = on
			pr.ModifierTagsCriteria = on
		default:
		}
	}
	group := gurps.NewPrereqList()
	group.All = false
	group.WhenTL.Compare = criteria.AtMostNumber
	group.Parent = root
	root.Prereqs = append(root.Prereqs, group)
}

// seedEveryFeatureControl turns on every optional criterion of the features and gives each grant an item, so that an
// audit sees every control the panel can show.
func seedEveryFeatureControl(list gurps.Features) {
	on := criteria.Text{Compare: criteria.IsText}
	for _, one := range list {
		one.SetSwitchable(true)
		switch f := one.(type) {
		case *gurps.AttributeBonus:
			f.Attribute = gurps.StrengthID
			f.Limitation = stlimit.StrikingOnly
		case *gurps.ConditionalModifierBonus:
			f.Group = "Audit"
		case *gurps.ReactionBonus:
			f.Group = "Audit"
		case *gurps.DRBonus:
			f.Specialization = "crushing"
		case *gurps.SkillBonus:
			f.SpecializationCriteria = on
			f.OptionalSpecializationCriteria = on
			f.TagsCriteria = on
		case *gurps.SkillPointBonus:
			f.SpecializationCriteria = on
			f.OptionalSpecializationCriteria = on
			f.TagsCriteria = on
		case *gurps.SpellBonus:
			f.TagsCriteria = on
		case *gurps.SpellPointBonus:
			f.TagsCriteria = on
		case *gurps.TraitBonus:
			f.TagsCriteria = on
		case *gurps.EquipmentMaxUsesBonus:
			f.SelectionType = equipmentsel.EquipmentWithName
			f.TagsCriteria = on
		case *gurps.TraitMaxLevelBonus:
			f.SelectionType = traitsel.TraitWithName
			f.TagsCriteria = on
		case *gurps.WeaponBonus:
			f.SpecializationCriteria = on
			f.UsageCriteria = on
			f.DamageTypeCriteria = on
			f.TagsCriteria = on
			f.RelativeLevelCriteria = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.One}
		case *gurps.SelectorOverride:
			f.SelectionType = wsel.WithRequiredSkill
			f.SpecializationCriteria = on
			f.UsageCriteria = on
			f.DamageTypeCriteria = on
			f.TagsCriteria = on
			f.RelativeLevelCriteria = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.One}
		case *gurps.Grant:
			switch f.Type {
			case feature.GrantSkill:
				f.SetItem(newTestSkill("Audit Granted Skill", fxp.One, nil), gurps.LibraryFile{})
			case feature.GrantSpell:
				f.SetItem(newTestSpell("Audit Granted Spell", fxp.One, nil), gurps.LibraryFile{})
			default:
				f.SetItem(newGrantTestTrait("Audit Granted Trait", fxp.One), gurps.LibraryFile{})
			}
			f.CheckPrereqs = true
		default:
		}
	}
}

// axNameAudit is uxtest.AXNameAudit with checkRows, which knows the panels of ux's editors.
type axNameAudit struct {
	*uxtest.AXNameAudit
	t      *testing.T
	screen *unison.HeadlessScreen
}

// checkRows opens a dockable with fn and checks the controls in it, with its prerequisites, defaults and features
// panels collapsed as they start out, then again with them expanded and each of their rows open in turn, since a closed
// row shows only its sentence.
func (a *axNameAudit) checkRows(view string, fn func()) {
	a.t.Helper()
	d := a.Open(fn)
	if d == nil {
		return
	}
	a.Check(view, d)
	// A prerequisites, defaults or features panel with anything in it starts out collapsed, showing a paragraph in
	// place of its rows, and its title bar says so.
	var collapsed []*sectionToggle
	a.screen.Do(func() {
		for _, p := range uxtest.PanelsOfType[*prereqPanel](d.AsPanel()) {
			if len(p.tree().Prereqs) != 0 {
				collapsed = append(collapsed, p.collapse)
			}
		}
		for _, p := range uxtest.PanelsOfType[*featuresPanel](d.AsPanel()) {
			if len(*p.features) != 0 {
				collapsed = append(collapsed, p.collapse)
			}
		}
		for _, p := range uxtest.PanelsOfType[*defaultsPanel](d.AsPanel()) {
			if len(*p.defaults) != 0 {
				collapsed = append(collapsed, p.collapse)
			}
		}
	})
	for _, toggle := range collapsed {
		if node := a.screen.AccessibilityNodeFor(toggle); node == nil || !node.Expandable || node.Expanded {
			a.t.Errorf("%s: the collapsed %s panel's title bar isn't described as collapsed", view, toggle.border.Title)
		}
	}
	a.screen.Do(func() {
		for _, toggle := range collapsed {
			toggle.toggle()
		}
	})
	a.Check(view+", expanded", d)
	var names []string
	var toggles []func()
	a.screen.Do(func() {
		for _, p := range uxtest.PanelsOfType[*prereqPanel](d.AsPanel()) {
			for i, one := range p.tree().Prereqs {
				if one.PrereqType() != prereq.List && one.PrereqType() != prereq.Unknown {
					path := childPath(treeRootPath, i)
					names = append(names, "prerequisite "+path)
					toggles = append(toggles, func() { p.toggle(path) })
				}
			}
		}
		for _, p := range uxtest.PanelsOfType[*nameablesPanel](d.AsPanel()) {
			p.HasInSelfOrDescendants(func(one *unison.Panel) bool {
				if b, ok := one.Self.(*sentenceButton); ok && b.onClick != nil {
					path := strings.TrimSuffix(b.RefKey, keySentence)
					names = append(names, "nameable "+path)
					toggles = append(toggles, func() { p.toggle(path) })
				}
				return false
			})
		}
		for _, p := range uxtest.PanelsOfType[*featuresPanel](d.AsPanel()) {
			for i, one := range *p.features {
				if one.FeatureType() != feature.Unknown {
					path := strconv.Itoa(i)
					names = append(names, "feature "+path)
					toggles = append(toggles, func() { p.toggle(path) })
				}
			}
		}
		for _, p := range uxtest.PanelsOfType[*defaultsPanel](d.AsPanel()) {
			for i := range *p.defaults {
				path := strconv.Itoa(i)
				names = append(names, "default "+path)
				toggles = append(toggles, func() { p.toggle(path) })
			}
		}
	})
	if len(toggles) == 0 {
		a.t.Errorf("%s: no rows to open", view)
		return
	}
	for i, toggle := range toggles {
		a.screen.Do(toggle)
		a.Check(view+", "+names[i]+" open", d)
	}
}

type traitRow = *Node[*gurps.Trait]

// plainHeader is a column header not built around a label, which GCS never builds but unison allows.
type plainHeader struct {
	unison.Panel
	state unison.SortState
}

func newPlainHeader() *plainHeader {
	h := &plainHeader{}
	h.Self = h
	return h
}

func (h *plainHeader) SortState() unison.SortState { return h.state }

func (h *plainHeader) SetSortState(state unison.SortState) { h.state = state }

func (h *plainHeader) Less() func(a, b string) bool { return nil }

// columnHeaderTitle must name a header as unison's axColumnHeaderName does.
func TestColumnHeaderTitle(t *testing.T) {
	c := check.New(t)
	title := columnHeaderTitle[traitRow]
	c.Equal("Points", title(NewTableColumnHeader[*gurps.Trait]("Points", "The points", nil)))
	c.Equal("Trait", title(NewPageTableColumnHeader[*gurps.Trait]("Trait", "", nil)))
	c.Equal("Enabled", title(NewEditorListSVGHeader[*gurps.Trait](svg.Weight, "Enabled", nil, false)))
	c.Equal("Enabled", title(NewEditorListSVGHeader[*gurps.Trait](svg.Weight, "Enabled", nil, true)))
	c.Equal("", title(NewTableColumnHeader[*gurps.Trait]("", "", nil)), "a header with nothing to say has no title")
	named := NewTableColumnHeader[*gurps.Trait]("Pts", "", nil)
	named.Accessibility.Name = "Points"
	c.Equal("Points", title(named), "the name wins over the text")

	plain := newPlainHeader()
	c.Equal("", title(plain))
	plain.Tooltip = unison.NewTooltipWithText("The tip")
	c.Equal("The tip", title(plain))
	plain.Tooltip.Accessibility.Name = ""
	c.Equal("The tip", title(plain), "a tooltip without a name is read from its labels")
	within := unison.NewLabel()
	within.SetTitle("Within")
	plain.AddChild(within)
	c.Equal("Within", title(plain), "the labels within the header win over its tooltip")
	second := unison.NewLabel()
	second.SetTitle("the header")
	plain.AddChild(second)
	c.Equal("Within the header", title(plain), "the labels are run together with a space between them")
	second.Hidden = true
	c.Equal("Within", title(plain), "a hidden label says nothing")
	plain.Accessibility.Name = "Named"
	c.Equal("Named", title(plain))

	data := gurps.HeaderData{
		Title: gurps.HeaderBookmark, Name: "Page Reference", Detail: gurps.PageRefTooltip(),
		TitleIsImageKey: true,
	}
	c.Equal("Page Reference", title(headerFromData[*gurps.Trait](data, false)))
	c.Equal("Page Reference", title(headerFromData[*gurps.Trait](data, true)))
	data = gurps.HeaderData{
		Title: gurps.HeaderStackedCoins, Name: "Extended Value", Detail: "The value of all",
		TitleIsImageKey: true,
	}
	c.Equal("Extended Value", columnHeaderTitle[*Node[*gurps.Equipment]](headerFromData[*gurps.Equipment](data, false)))
	data = gurps.HeaderData{Title: "Pts", Name: "Points", Detail: "Points"}
	c.Equal("Points", title(headerFromData[*gurps.Trait](data, false)))
	c.Equal("Points", title(headerFromData[*gurps.Trait](data, true)))
	data = gurps.HeaderData{Title: "Trait", Detail: "The trait"}
	c.Equal("Trait", title(headerFromData[*gurps.Trait](data, false)), "a title that is a word is read as drawn")
}

// The header row is never a tab stop, so the columns are announced as the focus arrives on a list at row level.
// Arriving at cell level is a return from a cell, so nothing is said. A click puts the cursor at row level, so it
// announces the columns whatever level the list was left at.
func TestTableAnnouncesItsColumnsOnEntry(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var table *unison.Table[traitRow]
	var header *unison.TableHeader[traitRow]
	var field *StringField
	screen.Do(func() {
		entity := sheet.Entity()
		tr := gurps.NewTrait(entity, nil, false)
		tr.Name = "Alpha"
		tr.PageRef = "B34"
		entity.Traits = append(entity.Traits, tr)
		entity.Recalculate()
		sheet.Rebuild(true)
		table, _ = uxtest.FirstPanelOfType[*unison.Table[traitRow]](sheet.AsPanel())
		header, _ = uxtest.FirstPanelOfType[*unison.TableHeader[traitRow]](sheet.AsPanel())
		field, _ = uxtest.FirstPanelOfType[*StringField](sheet.AsPanel())
	})
	if table == nil || header == nil || field == nil {
		t.Fatal("the character sheet must show the traits list, its header and a text field")
	}
	const expected = "Columns: Trait, Points, Page Reference, Library Source"
	c.Equal(expected, columnsAnnouncement(header),
		"the columns are named by their titles, the abbreviation by what it stands for and the icons by their short names")
	leadColumn := func() (col int) {
		screen.Do(func() { col = table.LeadColumnIndex() })
		return col
	}

	// Asking for the tree turns accessibility on, without which nothing is announced.
	tree := screen.AccessibilityTree(wnd)
	if tree == nil {
		t.Fatal("the window must be described")
	}
	tableNode := screen.AccessibilityNodeFor(table)
	if tableNode == nil {
		t.Fatal("the traits list must be described")
	}
	var alpha *accessibility.Node
	for _, id := range tableNode.Children {
		if n := tree.Node(id); n != nil && n.Role == role.Row && strings.HasPrefix(n.Name, "Alpha") {
			alpha = n
		}
	}
	if alpha == nil {
		t.Fatal("the Alpha row must be described")
	}
	c.True(strings.Contains(alpha.Name, "Page Reference B34"), "the row reads the reference after its column's name: %q", alpha.Name)
	c.False(strings.Contains(alpha.Name, gurps.PageRefTooltip()), "the row never reads the header's tooltip: %q", alpha.Name)
	screen.Announcements()
	focus := func(p unison.Paneler) { screen.Do(func() { p.AsPanel().RequestFocus() }) }
	focus(field)
	c.Equal(0, len(screen.Announcements()), "nothing is announced for a text field")
	focus(table)
	c.Equal([]string{expected}, screen.Announcements(), "the columns are announced as the focus arrives on the list")

	screen.Do(func() { table.SetLeadCell(0, 1) })
	screen.Announcements()
	focus(field)
	focus(table)
	c.Equal(0, len(screen.Announcements()), "nothing is announced when the focus returns at cell level")
	c.Equal(1, leadColumn(), "the cursor stays on the cell")

	focus(field)
	screen.Announcements()
	screen.Click(screen.PanelPoint(table, geom.NewPoint(4, 4)))
	c.Equal(-1, leadColumn(), "a click puts the cursor at row level")
	c.Equal([]string{expected}, screen.Announcements(), "the columns are announced as a click brings the focus in")

	screen.Do(func() { table.SetLeadCell(0, -1) })
	screen.Announcements()
	focus(field)
	focus(table)
	c.Equal([]string{expected}, screen.Announcements())
}

// The buttons a sheet keeps out of the Tab order (the description's randomize buttons and the buttons that show and
// hide an attribute section or a hit location's sub-table) are tab stops that Space presses while a screen reader is
// listening and the setting that puts static text and disabled controls in the Tab order is on. Otherwise they take no
// focus, and one that is a tab stop when the screen reader goes drops out as the focus leaves it.
func TestSheetButtonsJoinTheTabOrderForReading(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	setFocusForReading := uxtest.FocusForReadingSetter(t, screen, wnd)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var buttons []*unison.Button
	var randomize *unison.Button
	var gender *StringField
	var inDescription, inBody, inAttributes int
	screen.Do(func() {
		// A new sheet has neither a hit location with a sub-table nor a section among its attributes.
		entity := sheet.Entity()
		settings := gurps.SheetSettingsFor(entity)
		body := settings.BodyType
		subTable := &gurps.Body{Roll: gurps.Roller.Parse("1d")}
		subTable.SetOwningLocation(body.Locations[0])
		body.Locations[0].SubTable = subTable
		loc := gurps.NewHitLocation(entity, "")
		loc.LocID = "within"
		loc.ChoiceName = "Within"
		loc.TableName = "Within"
		loc.Slots = 6
		subTable.AddLocation(loc)
		body.Update(entity)
		settings.Attributes.Set["section"] = &gurps.AttributeDef{
			DefID: "section",
			Type:  attribute.SecondarySeparator,
			Name:  "Section",
			Order: len(settings.Attributes.Set) + 1,
		}
		entity.Recalculate()
		sheet.Rebuild(true)
		desc, _ := uxtest.FirstPanelOfType[*DescriptionPanel](sheet.AsPanel())
		if desc == nil {
			return
		}
		for _, one := range uxtest.PanelsOfType[*unison.Button](desc.AsPanel()) {
			buttons = append(buttons, one)
			if randomize == nil {
				randomize = one
			}
		}
		for _, one := range uxtest.PanelsOfType[*StringField](desc.AsPanel()) {
			if one.RefKey == descriptionPanelGenderFieldRefKey {
				gender = one
			}
		}
		inDescription = len(buttons)
		if body, found := uxtest.FirstPanelOfType[*BodyPanel](sheet.AsPanel()); found {
			buttons = append(buttons, uxtest.PanelsOfType[*unison.Button](body.AsPanel())...)
		}
		inBody = len(buttons) - inDescription
		for _, one := range uxtest.PanelsOfType[*AttrPanel](sheet.AsPanel()) {
			buttons = append(buttons, uxtest.PanelsOfType[*unison.Button](one.AsPanel())...)
		}
		inAttributes = len(buttons) - inDescription - inBody
	})
	if randomize == nil || gender == nil {
		t.Fatal("the character sheet must show the description, with a Gender field and its randomize button")
	}
	c.Equal(9, inDescription, "the description must have its randomize buttons")
	c.Equal(1, inBody, "the hit location with a sub-table must have the button that shows and hides it")
	c.Equal(1, inAttributes, "the section of the attributes must have the button that shows and hides it")
	focusable := func(p unison.Paneler) (result bool) {
		screen.Do(func() { result = p.AsPanel().Focusable() })
		return result
	}
	focus := func() (focus *unison.Panel) {
		screen.Do(func() { focus = wnd.Focus() })
		return focus
	}
	none := func(why string) {
		t.Helper()
		for _, one := range buttons {
			c.False(focusable(one), "%s, but a button took the focus", why)
		}
	}
	none("a button kept out of the Tab order takes no focus while no screen reader is listening")

	setFocusForReading(false)
	none("a button kept out of the Tab order takes no focus while the setting is off")
	node := screen.AccessibilityNodeFor(randomize)
	if node == nil {
		t.Fatal("the randomize button must be described")
	}
	c.Equal(role.Button, node.Role)
	c.False(node.Focusable, "the button is described as it is")
	c.False(node.Actions.Has(accessibility.Focus))
	c.True(node.Actions.Has(accessibility.Press), "a screen reader can press the button even so")

	setFocusForReading(true)
	for _, one := range buttons {
		c.True(focusable(one), "a button kept out of the Tab order is a tab stop while the setting is on")
	}
	node = screen.AccessibilityNodeFor(randomize)
	if node == nil {
		t.Fatal("the randomize button must be described")
	}
	c.True(node.Focusable, "the button is described as able to take the focus")
	c.True(node.Actions.Has(accessibility.Focus))
	c.Equal("Randomize the gender using the current ancestry", node.Name)

	// Tab goes from the tab stop ahead of the button to the button, and Space presses it.
	var ahead *unison.Panel
	screen.Do(func() {
		all := uxtest.PanelsMatching(sheet.scroll.Content().AsPanel(), (*unison.Panel).Focusable)
		for i, one := range all {
			if one == randomize.AsPanel() && i > 0 {
				ahead = all[i-1]
			}
		}
	})
	if ahead == nil {
		t.Fatal("the randomize button must follow another tab stop")
	}
	screen.Do(ahead.RequestFocus)
	c.Equal(ahead, focus())
	screen.KeyPress(unison.KeyTab, mod.None)
	c.Equal(randomize.AsPanel(), focus(), "Tab reaches the randomize button")
	screen.Do(func() {
		gender.SetText("")
		sheet.Entity().Profile.Gender = ""
	})
	screen.KeyPress(unison.KeySpace, mod.None)
	var text string
	screen.Do(func() { text = gender.Text() })
	c.NotEqual("", text, "Space on the randomize button randomizes the gender")
	c.Equal(gender.AsPanel(), focus(), "and puts the focus on the field, where a screen reader reads what it now holds")

	// A screen reader's Focus action on the button moves the focus there.
	c.True(screen.PerformAccessibilityAction(accessibility.ActionRequest{Node: node.ID, Action: accessibility.Focus}))
	c.Equal(randomize.AsPanel(), focus())

	setFocusForReading(false)
	none("a button kept out of the Tab order takes no focus once the setting is off again")

	// Nothing describes the window once the screen reader goes, so the button stays a tab stop until it loses the
	// focus.
	setFocusForReading(true)
	screen.Do(randomize.RequestFocus)
	c.Equal(randomize.AsPanel(), focus())
	t.Cleanup(func() { unison.SetAccessibilityEnabled(true) })
	screen.Do(func() { unison.SetAccessibilityEnabled(false) })
	var active bool
	screen.Do(func() { active = unison.IsAccessibilityActive() })
	c.False(active, "the screen reader must have gone")
	screen.Do(gender.RequestFocus)
	c.Equal(gender.AsPanel(), focus())
	c.False(focusable(randomize), "a button the focus has left is no tab stop once the screen reader has gone")
}

// The hit locations' notes header is only an icon, so it is a label named Notes, described by its tooltip, and a tab
// stop with static text in the Tab order, as the Roll, Location and DR headers are.
func TestBodyNotesHeaderIsReadAsTheOtherHeadersAre(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	setFocusForReading := uxtest.FocusForReadingSetter(t, screen, wnd)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var headers []*unison.Label
	screen.Do(func() {
		body, found := uxtest.FirstPanelOfType[*BodyPanel](sheet.AsPanel())
		if !found {
			return
		}
		// The first seven rows are the four headers and the empty panels in the separator columns between them.
		if rows := blockRows(body.AsPanel()); len(rows) > 7 {
			for _, child := range rows[:7] {
				if label, isLabel := child.Self.(*unison.Label); isLabel {
					headers = append(headers, label)
				}
			}
		}
	})
	if len(headers) != 4 {
		t.Fatalf("the body table must have four headers, but has %d", len(headers))
	}
	focusable := func(p unison.Paneler) (result bool) {
		screen.Do(func() { result = p.AsPanel().Focusable() })
		return result
	}
	setFocusForReading(true)
	names := make([]string, 0, len(headers))
	for _, one := range headers {
		node := screen.AccessibilityNodeFor(one)
		if node == nil {
			t.Fatalf("the %q header must be described", one.String())
		}
		names = append(names, node.Name)
		c.Equal(role.Label, node.Role, "the %q header is a label", node.Name)
		c.False(node.Ignored, "the %q header must be described", node.Name)
		c.True(node.Focusable, "the %q header is described as a tab stop", node.Name)
		c.True(focusable(one), "the %q header is a tab stop with static text in the Tab order", node.Name)
	}
	c.Equal([]string{"Roll", "Location", "DR", "Notes"}, names)
	notes := screen.AccessibilityNodeFor(headers[3])
	c.Equal("Notes for the hit location", notes.Description, "the tooltip says the rest")

	setFocusForReading(false)
	for _, one := range headers {
		c.False(focusable(one), "the %q header is no tab stop while the setting is off", one.String())
	}
}

// A label that follows a field, such as its units, is the field's description rather than an element of its own, so it
// is no tab stop with static text in the Tab order. (The calculators' link-aware labels are covered by the same test
// in ux/calculators.)
func TestTrailingLabelsDescribeTheirFields(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	setFocusForReading := uxtest.FocusForReadingSetter(t, screen, wnd)
	setFocusForReading(true)
	focusable := func(p unison.Paneler) (result bool) {
		screen.Do(func() { result = p.AsPanel().Focusable() })
		return result
	}
	screen.Do(ShowGeneralSettings)
	var settings *generalSettingsDockable
	screen.Do(func() { settings, _ = uxtest.FirstPanelOfType[*generalSettingsDockable](wnd.Content()) })
	if settings == nil {
		t.Fatal("the General Settings must be showing")
	}
	if screen.AccessibilityTree(wnd) == nil {
		t.Fatal("the window must be described")
	}
	for _, one := range []struct {
		field *unison.Panel
		name  string
		hint  string
	}{
		{field: settings.tooltipDelayField.AsPanel(), name: "Tooltip Delay", hint: "seconds"},
		{field: settings.tooltipDismissalField.AsPanel(), name: "Tooltip Dismissal", hint: "seconds"},
		{field: settings.cursorSizeField.AsPanel(), name: "Cursor Size", hint: "points"},
		{field: settings.exportResolutionField.AsPanel(), name: "Image Export Resolution", hint: "ppi"},
	} {
		node := screen.AccessibilityNodeFor(one.field)
		if node == nil {
			t.Fatalf("the %s field must be described", one.name)
		}
		c.Equal(one.name, node.Name)
		c.Equal(one.hint, node.Description, "the %s field is described by the label that follows it", one.name)
		var units *unison.Label
		screen.Do(func() {
			siblings := one.field.Parent().Children()
			if len(siblings) == 2 {
				if found, isLabel := siblings[1].Self.(*unison.Label); isLabel {
					units = found
				}
			}
		})
		if units == nil {
			t.Fatalf("the %s field must be followed by its units", one.name)
		}
		c.Equal(one.hint, units.String())
		c.False(focusable(units), "the units of the %s field are heard with it, so they are no tab stop", one.name)
		if label := screen.AccessibilityNodeFor(units); label != nil {
			c.True(label.Ignored, "the units of the %s field are not described, but are %+v", one.name, label)
		}
	}
}

// With a screen reader listening and the setting that puts static text in the Tab order on, the read-only fields (a hit
// location's penalty and DR, and the General Settings path rows) are tab stops, as the labels beside them are. With the
// setting off none of them is, while the notes field and the portrait, being controls, always are. The space shown for
// a location without a roll is never a tab stop, and a read-only field without text is left out of the description.
//
// On macOS and Windows the block headings follow the setting too, but a headless session follows Linux's rules, under
// which unison makes a heading a tab stop whenever a screen reader is listening, since that is how Orca reaches it.
func TestReadOnlyFieldsJoinTheTabOrderForScreenReaders(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	setFocusable := uxtest.FocusForReadingSetter(t, screen, wnd)
	focusable := func(p unison.Paneler) (result bool) {
		screen.Do(func() { result = p.AsPanel().Focusable() })
		return result
	}

	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var body *BodyPanel
	var portrait *PortraitPanel
	var headings []*unison.Panel
	var readOnly []*NonEditablePageField
	var labels, blank []*unison.Label
	var notes []*StringField
	screen.Do(func() {
		portrait, _ = uxtest.FirstPanelOfType[*PortraitPanel](sheet.AsPanel())
		headings = uxtest.PanelsMatching(sheet.AsPanel(), func(p *unison.Panel) bool {
			return p.Accessibility.Role == role.Heading
		})
		body, ok = uxtest.FirstPanelOfType[*BodyPanel](sheet.AsPanel())
		if !ok {
			return
		}
		readOnly = uxtest.PanelsOfType[*NonEditablePageField](body.AsPanel())
		// The notes header has no text; TestBodyNotesHeaderIsReadAsTheOtherHeadersAre covers it.
		for _, one := range uxtest.PanelsOfType[*unison.Label](body.AsPanel()) {
			switch one.String() {
			case "":
			case " ":
				blank = append(blank, one)
			default:
				labels = append(labels, one)
			}
		}
		notes = uxtest.PanelsOfType[*StringField](body.AsPanel())
	})
	if body == nil {
		t.Fatal("the character sheet must show the body table")
	}
	c.NotNil(portrait, "the character sheet must show the portrait")
	c.True(len(headings) >= 3, "the character sheet's blocks must each carry a heading")
	// Each heading is its block's first child and holds the title strip, which the border leaves to the content, so the
	// block's rows must be laid out below it.
	screen.Do(func() {
		for _, heading := range headings {
			block := heading.Parent()
			children := block.Children()
			c.True(len(children) > 1 && children[0] == heading, "the %q heading comes first in its block", heading.Accessibility.Name)
			c.Equal(block.Accessibility.Name, heading.Accessibility.Name, "the block is named as its heading is")
			bottom := heading.FrameRect().Y + heading.FrameRect().Height
			for _, child := range children[1:] {
				c.True(child.FrameRect().Y >= bottom, "in %q, a %T at %v is laid out under the title strip",
					heading.Accessibility.Name, child.Self, child.FrameRect())
			}
		}
	})
	c.True(len(readOnly) >= 2, "the body table must show a penalty and a DR for each hit location")
	c.True(len(labels) >= 2, "the body table must show a roll and a location for each hit location")
	c.True(len(blank) >= 1, "the default body must have a location without a roll, shown as a space")
	c.True(len(notes) >= 1, "the body table must show a notes field for each hit location")
	for _, one := range readOnly {
		c.NotEqual("", one.Text.String(), "each penalty and DR must have text, or there would be nothing to read")
	}

	// With the setting off, only the controls and, under Linux's rules, the headings take the focus.
	setFocusable(false)
	for _, one := range readOnly {
		c.False(focusable(one), "%q must not be a tab stop while the setting is off", one.Text.String())
	}
	for _, one := range labels {
		c.False(focusable(one), "%q must not be a tab stop while the setting is off", one.String())
	}
	c.True(focusable(notes[0]), "the notes field is a control, so it is always a tab stop")
	c.True(focusable(portrait), "the portrait is a control, so it is always a tab stop")
	for _, one := range headings {
		c.True(focusable(one), "the %q heading is a tab stop under Linux's rules whatever the setting", one.Accessibility.Name)
	}

	setFocusable(true)
	for _, one := range headings {
		c.True(focusable(one), "the %q heading must be a tab stop while the setting is on", one.Accessibility.Name)
	}
	for _, one := range readOnly {
		c.True(focusable(one), "%q must be a tab stop while the setting is on", one.Text.String())
	}
	for _, one := range labels {
		c.True(focusable(one), "%q must be a tab stop while the setting is on", one.String())
	}
	for _, one := range blank {
		c.False(focusable(one), "a location without a roll must not offer a blank tab stop")
	}
	c.True(focusable(notes[0]))

	screen.Do(ShowGeneralSettings)
	var paths, empty []*PathField
	screen.Do(func() {
		// The log path is empty here, since only main sets it.
		for _, one := range uxtest.PanelsOfType[*PathField](wnd.Content()) {
			if one.path != "" {
				paths = append(paths, one)
			} else {
				empty = append(empty, one)
			}
		}
	})
	c.True(len(paths) >= 1, "the General Settings must show its path rows")
	c.True(len(empty) >= 1, "the log path must be empty here, or there is nothing to check the description of")
	if screen.AccessibilityTree(wnd) == nil {
		t.Fatal("the window must be described")
	}
	for _, one := range paths {
		c.True(focusable(one), "%q must be a tab stop while the setting is on", one.path)
		node := screen.AccessibilityNodeFor(one)
		if node == nil {
			t.Fatalf("%q must be described", one.path)
		}
		c.False(node.Ignored, "%q must be described", one.path)
		c.Equal(one.path, node.Name, "the whole path is the name, however much of it is showing")
	}
	for _, one := range empty {
		c.False(focusable(one), "a field without text must not be a tab stop")
		if node := screen.AccessibilityNodeFor(one); node != nil {
			c.True(node.Ignored, "a field without text must be left out of the description, but is %+v", node)
		}
	}

	setFocusable(false)
	for _, one := range readOnly {
		c.False(focusable(one), "%q must not be a tab stop once the setting is off again", one.Text.String())
	}
	for _, one := range headings {
		c.True(focusable(one), "the %q heading is a tab stop under Linux's rules whatever the setting", one.Accessibility.Name)
	}
	c.True(focusable(portrait))
	for _, one := range paths {
		c.False(focusable(one), "%q must not be a tab stop once the setting is off again", one.path)
	}
}
