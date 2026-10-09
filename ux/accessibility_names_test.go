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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/equipmentsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/namegen"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/study"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/traitsel"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
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
	screen.Do(func() {
		entity = sheet.Entity()
		// Give the sheet's lists a row apiece so that their cells are described.
		tr := gurps.NewTrait(entity, nil, false)
		tr.Name = "Audit Trait"
		entity.Traits = append(entity.Traits, tr)
		sk := gurps.NewSkill(entity, nil, false)
		sk.Name = "Audit Skill"
		entity.Skills = append(entity.Skills, sk)
		eq := gurps.NewEquipment(entity, nil, false)
		eq.Name = "Audit Equipment"
		entity.CarriedEquipment = append(entity.CarriedEquipment, eq)
		entity.Recalculate()
		sheet.Rebuild(true)
	})
	audit.Check("character sheet", sheet)
	screen.Do(func() { sheet.toggleLayoutEditing() })
	audit.Check("character sheet layout editing", sheet)
	screen.Do(func() { sheet.toggleLayoutEditing() })

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
		tr.Name = "Audit Trait"
		tr.Features = allFeatures(tr, false)
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

	audit.CheckOpened("sheet settings", func() { ShowSheetSettings(sheet) })
	audit.CheckOpened("attribute settings", func() { ShowAttributeSettings(sheet) })
	audit.CheckOpened("body settings", func() { ShowBodySettings(sheet) })
	audit.CheckOpened("general settings", ShowGeneralSettings)
	audit.CheckOpened("color settings", ShowColorSettings)
	audit.CheckOpened("font settings", ShowFontSettings)
	audit.CheckOpened("menu key settings", ShowMenuKeySettings)
	audit.CheckOpened("page reference mappings", ShowPageRefMappings)
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
			pr.NotesCriteria = on
		case *gurps.SkillPrereq:
			pr.SpecializationCriteria = on
			pr.OptionalSpecializationCriteria = on
		case *gurps.SpellPrereq:
			if !pr.SamePowerSource {
				pr.PowerSourceCriteria = on
			}
		case *gurps.AttributePrereq:
			pr.CombinedWith = gurps.DexterityID
		case *gurps.EquippedEquipmentPrereq:
			pr.TagsCriteria = on
		default:
		}
	}
	group := gurps.NewPrereqList()
	group.All = false
	group.WhenTL.Compare = criteria.AtMostNumber
	group.Parent = root
	root.Prereqs = append(root.Prereqs, group)
}

// seedEveryFeatureControl turns on every optional criterion of the features, so that an audit sees every control the
// panel can show.
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
			f.TagsCriteria = on
			f.RelativeLevelCriteria = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.One}
		case *gurps.SelectorOverride:
			f.UsageCriteria = on
			f.TagsCriteria = on
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
