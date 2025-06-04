// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

var (
	_ FileBackedDockable         = &Template{}
	_ unison.UndoManagerProvider = &Template{}
	_ ModifiableRoot             = &Template{}
	_ Rebuildable                = &Template{}
	_ unison.TabCloser           = &Template{}
	_ KeyedDockable              = &Template{}
)

// Template holds the view for a GURPS character template.
type Template struct {
	unison.Panel
	path              string
	targetMgr         *TargetMgr
	undoMgr           *unison.UndoManager
	toolbar           *unison.Panel
	scroll            *unison.ScrollPanel
	template          *gurps.Template
	hash              uint64
	content           *templateContent
	Traits            *PageList[*gurps.Trait]
	Skills            *PageList[*gurps.Skill]
	Spells            *PageList[*gurps.Spell]
	Equipment         *PageList[*gurps.Equipment]
	Notes             *PageList[*gurps.Note]
	dragReroutePanel  *unison.Panel
	lastBody          *gurps.Body
	searchTracker     *SearchTracker
	scale             int
	awaitingUpdate    bool
	needsSaveAsPrompt bool
}

// OpenTemplates returns the currently open templates.
func OpenTemplates(exclude *Template) []*Template {
	var templates []*Template
	for _, one := range AllDockables() {
		if template, ok := one.(*Template); ok && template != exclude {
			templates = append(templates, template)
		}
	}
	return templates
}

// NewTemplateFromFile loads a GURPS template file and creates a new unison.Dockable for it.
func NewTemplateFromFile(filePath string) (unison.Dockable, error) {
	template, err := gurps.NewTemplateFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	t := NewTemplate(filePath, template)
	t.needsSaveAsPrompt = false
	return t, nil
}

// NewTemplate creates a new unison.Dockable for GURPS template files.
func NewTemplate(filePath string, template *gurps.Template) *Template {
	t := &Template{
		path:              filePath,
		undoMgr:           unison.NewUndoManager(200, func(err error) { errs.Log(err) }),
		scroll:            unison.NewScrollPanel(),
		template:          template,
		lastBody:          template.BodyType,
		scale:             gurps.GlobalSettings().General.InitialSheetUIScale,
		hash:              gurps.Hash64(template),
		needsSaveAsPrompt: true,
	}
	if t.lastBody == nil {
		t.lastBody = gurps.FactoryBody()
	}
	t.Self = t
	t.targetMgr = NewTargetMgr(t)
	t.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  align.Fill,
		VAlign:  align.Fill,
	})

	t.MouseDownCallback = func(_ unison.Point, _, _ int, _ unison.Modifiers) bool {
		t.RequestFocus()
		return false
	}
	t.DataDragOverCallback = func(_ unison.Point, data map[string]any) bool {
		t.dragReroutePanel = nil
		for _, key := range dropKeys {
			if _, ok := data[key]; ok {
				if t.dragReroutePanel = t.keyToPanel(key); t.dragReroutePanel != nil {
					t.dragReroutePanel.DataDragOverCallback(unison.Point{Y: 100000000}, data)
					return true
				}
				break
			}
		}
		return false
	}
	t.DataDragExitCallback = func() {
		if t.dragReroutePanel != nil {
			t.dragReroutePanel.DataDragExitCallback()
			t.dragReroutePanel = nil
		}
	}
	t.DataDragDropCallback = func(_ unison.Point, data map[string]any) {
		if t.dragReroutePanel != nil {
			t.dragReroutePanel.DataDragDropCallback(unison.Point{Y: 10000000}, data)
			t.dragReroutePanel = nil
		}
	}
	t.DrawOverCallback = func(gc *unison.Canvas, _ unison.Rect) {
		if t.dragReroutePanel != nil {
			r := t.RectFromRoot(t.dragReroutePanel.RectToRoot(t.dragReroutePanel.ContentRect(true)))
			paint := unison.ThemeWarning.Paint(gc, r, paintstyle.Fill)
			paint.SetColorFilter(unison.Alpha30Filter())
			gc.DrawRect(r, paint)
		}
	}

	t.scroll.SetContent(t.createContent(), behavior.Unmodified, behavior.Unmodified)
	t.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	t.createToolbar()
	t.AddChild(t.scroll)

	t.InstallCmdHandlers(SaveItemID, func(_ any) bool { return t.Modified() }, func(_ any) { t.save(false) })
	t.InstallCmdHandlers(SaveAsItemID, unison.AlwaysEnabled, func(_ any) { t.save(true) })
	t.installNewItemCmdHandlers(NewTraitItemID, NewTraitContainerItemID, t.Traits)
	t.installNewItemCmdHandlers(NewSkillItemID, NewSkillContainerItemID, t.Skills)
	t.installNewItemCmdHandlers(NewTechniqueItemID, -1, t.Skills)
	t.installNewItemCmdHandlers(NewSpellItemID, NewSpellContainerItemID, t.Spells)
	t.installNewItemCmdHandlers(NewRitualMagicSpellItemID, -1, t.Spells)
	t.installNewItemCmdHandlers(NewCarriedEquipmentItemID, NewCarriedEquipmentContainerItemID, t.Equipment)
	t.installNewItemCmdHandlers(NewNoteItemID, NewNoteContainerItemID, t.Notes)
	t.InstallCmdHandlers(AddNaturalAttacksItemID, unison.AlwaysEnabled, func(_ any) {
		InsertItems(t, t.Traits.Table, t.template.TraitList, t.template.SetTraitList,
			func(_ *unison.Table[*Node[*gurps.Trait]]) []*Node[*gurps.Trait] {
				return t.Traits.provider.RootRows()
			}, gurps.NewNaturalAttacks(nil, nil))
	})
	t.InstallCmdHandlers(ApplyTemplateItemID, t.canApplyTemplate, t.applyTemplate)
	t.InstallCmdHandlers(NewSheetFromTemplateItemID, unison.AlwaysEnabled, t.newSheetFromTemplate)

	t.template.EnsureAttachments()
	t.template.SourceMatcher().PrepareHashes(t.template)
	return t
}

// DockKey implements KeyedDockable.
func (t *Template) DockKey() string {
	return filePrefix + t.path
}

func (t *Template) createToolbar() {
	t.toolbar = unison.NewPanel()
	t.AddChild(t.toolbar)
	t.toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
	t.toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	t.toolbar.AddChild(NewDefaultInfoPop())

	helpButton := unison.NewSVGButton(svg.Help)
	helpButton.Tooltip = newWrappedTooltip(i18n.Text("Help"))
	helpButton.ClickCallback = func() { HandleLink(nil, "md:Help/Interface/Character Template") }
	t.toolbar.AddChild(helpButton)

	t.toolbar.AddChild(
		NewScaleField(
			gurps.InitialUIScaleMin,
			gurps.InitialUIScaleMax,
			func() int { return gurps.GlobalSettings().General.InitialSheetUIScale },
			func() int { return t.scale },
			func(scale int) { t.scale = scale },
			nil,
			false,
			t.scroll,
		),
	)

	hierarchyButton := unison.NewSVGButton(svg.Hierarchy)
	hierarchyButton.Tooltip = newWrappedTooltip(i18n.Text("Opens/closes all hierarchical rows"))
	hierarchyButton.ClickCallback = t.toggleHierarchy
	t.toolbar.AddChild(hierarchyButton)

	noteToggleButton := unison.NewSVGButton(svg.NotesToggle)
	noteToggleButton.Tooltip = newWrappedTooltip(i18n.Text("Opens/closes all embedded notes"))
	noteToggleButton.ClickCallback = t.toggleNotes
	t.toolbar.AddChild(noteToggleButton)

	applyTemplateButton := unison.NewSVGButton(svg.Stamper)
	applyTemplateButton.Tooltip = newWrappedTooltip(applyTemplateAction.Title)
	applyTemplateButton.ClickCallback = func() {
		if CanApplyTemplate() {
			t.applyTemplate(nil)
		}
	}
	t.toolbar.AddChild(applyTemplateButton)

	syncSourceButton := unison.NewSVGButton(svg.DownToBracket)
	syncSourceButton.Tooltip = newWrappedTooltip(i18n.Text("Sync with all sources in this sheet"))
	syncSourceButton.ClickCallback = func() { t.syncWithAllSources() }
	t.toolbar.AddChild(syncSourceButton)

	t.searchTracker = InstallSearchTracker(t.toolbar, func() {
		t.Traits.Table.ClearSelection()
		t.Skills.Table.ClearSelection()
		t.Spells.Table.ClearSelection()
		t.Equipment.Table.ClearSelection()
		t.Notes.Table.ClearSelection()
	}, func(refList *[]*searchRef, text string, namesOnly bool) {
		searchSheetTable(refList, text, namesOnly, t.Traits)
		searchSheetTable(refList, text, namesOnly, t.Skills)
		searchSheetTable(refList, text, namesOnly, t.Spells)
		searchSheetTable(refList, text, namesOnly, t.Equipment)
		searchSheetTable(refList, text, namesOnly, t.Notes)
	})

	t.toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(t.toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})
}

func (t *Template) keyToPanel(key string) *unison.Panel {
	var p unison.Paneler
	switch key {
	case equipmentDragKey:
		p = t.Equipment.Table
	case gurps.SkillID:
		p = t.Skills.Table
	case gurps.SpellID:
		p = t.Spells.Table
	case traitDragKey:
		p = t.Traits.Table
	case noteDragKey:
		p = t.Notes.Table
	default:
		return nil
	}
	return p.AsPanel()
}

// CanApplyTemplate returns true if a template can be applied.
func CanApplyTemplate() bool {
	return len(OpenSheets(nil)) > 0
}

func (t *Template) canApplyTemplate(_ any) bool {
	return CanApplyTemplate()
}

// NewSheetFromTemplate loads the specified template file and creates a new character sheet from it.
func NewSheetFromTemplate(filePath string) {
	d, err := NewTemplateFromFile(filePath)
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to load template"), err)
		return
	}
	if t, ok := d.(*Template); ok {
		t.newSheetFromTemplate(nil)
	}
}

func (t *Template) newSheetFromTemplate(_ any) {
	e := gurps.NewEntity()
	sheet := NewSheet(e.Profile.Name+gurps.SheetExt, e)
	DisplayNewDockable(sheet)
	if t.applyTemplateToSheet(sheet, true) {
		sheet.undoMgr.Clear()
		sheet.hash = 0
	}
	sheet.SetBackingFilePath(e.Profile.Name + gurps.SheetExt)
}

// ApplyTemplate loads the specified template file and applies it to a sheet.
func ApplyTemplate(filePath string) {
	t, err := NewTemplateFromFile(filePath)
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to load template"), err)
		return
	}
	if CanApplyTemplate() {
		if t, ok := t.(*Template); ok {
			t.applyTemplate(nil)
		}
	}
}

func (t *Template) applyTemplate(suppressRandomizePromptAsBool any) {
	//nolint:errcheck // The default of false on failure is acceptable
	suppressRandomizePrompt, _ := suppressRandomizePromptAsBool.(bool)
	for _, sheet := range PromptForDestination(OpenSheets(nil)) {
		t.applyTemplateToSheet(sheet, suppressRandomizePrompt)
	}
}

func (t *Template) applyTemplateToSheet(sheet *Sheet, suppressRandomizePrompt bool) bool {
	var undo *unison.UndoEdit[*ApplyTemplateUndoEditData]
	mgr := unison.UndoManagerFor(sheet)
	if mgr != nil {
		if beforeData, err := NewApplyTemplateUndoEditData(sheet); err != nil {
			errs.Log(err)
			mgr = nil
		} else {
			undo = &unison.UndoEdit[*ApplyTemplateUndoEditData]{
				ID:         unison.NextUndoID(),
				EditName:   i18n.Text("Apply Template"),
				UndoFunc:   func(e *unison.UndoEdit[*ApplyTemplateUndoEditData]) { e.BeforeData.Apply() },
				RedoFunc:   func(e *unison.UndoEdit[*ApplyTemplateUndoEditData]) { e.AfterData.Apply() },
				AbsorbFunc: func(_ *unison.UndoEdit[*ApplyTemplateUndoEditData], _ unison.Undoable) bool { return false },
				BeforeData: beforeData,
			}
		}
	}
	e := sheet.Entity()
	if t.template.BodyType != nil {
		e.SheetSettings.BodyType = t.template.BodyType.Clone(e, nil)
	}
	templateAncestries := gurps.ActiveAncestries(ExtractNodeDataFromList(t.Traits.Table.RootRows()))
	if len(templateAncestries) != 0 {
		entityAncestries := gurps.ActiveAncestries(e.Traits)
		if len(entityAncestries) != 0 {
			if unison.YesNoDialog(fmt.Sprintf(i18n.Text(`The template contains an Ancestry (%s).
Disable your character's existing Ancestry (%s)?`),
				templateAncestries[0].Name, entityAncestries[0].Name), "") == unison.ModalResponseOK {
				for _, one := range gurps.ActiveAncestryTraits(e.Traits) {
					one.Disabled = true
				}
			}
		}
	}
	traits := cloneRows(sheet.Traits.Table, t.Traits.Table.RootRows())
	skills := cloneRows(sheet.Skills.Table, t.Skills.Table.RootRows())
	spells := cloneRows(sheet.Spells.Table, t.Spells.Table.RootRows())
	equipment := cloneRows(sheet.CarriedEquipment.Table, t.Equipment.Table.RootRows())
	notes := cloneRows(sheet.Notes.Table, t.Notes.Table.RootRows())
	var abort bool
	if traits, abort = processPickerRows(traits); abort {
		return false
	}
	if skills, abort = processPickerRows(skills); abort {
		return false
	}
	if spells, abort = processPickerRows(spells); abort {
		return false
	}
	appendRows(sheet.Traits.Table, traits)
	appendRows(sheet.Skills.Table, skills)
	appendRows(sheet.Spells.Table, spells)
	appendRows(sheet.CarriedEquipment.Table, equipment)
	appendRows(sheet.Notes.Table, notes)
	sheet.Rebuild(true)
	ProcessModifiersForSelection(sheet.Traits.Table)
	ProcessModifiersForSelection(sheet.Skills.Table)
	ProcessModifiersForSelection(sheet.Spells.Table)
	ProcessModifiersForSelection(sheet.CarriedEquipment.Table)
	ProcessModifiersForSelection(sheet.Notes.Table)
	ProcessNameablesForSelection(sheet.Traits.Table)
	ProcessNameablesForSelection(sheet.Skills.Table)
	ProcessNameablesForSelection(sheet.Spells.Table)
	ProcessNameablesForSelection(sheet.CarriedEquipment.Table)
	ProcessNameablesForSelection(sheet.Notes.Table)
	if len(templateAncestries) != 0 && gurps.GlobalSettings().General.AutoFillProfile {
		randomize := true
		if !suppressRandomizePrompt {
			randomize = unison.YesNoDialog(i18n.Text("Would you like to apply the initial randomization again?"), "") == unison.ModalResponseOK
		}
		if randomize {
			e.Profile.ApplyRandomizers(e)
			updateRandomizedProfileFieldsWithoutUndo(sheet)
			sheet.Rebuild(true)
		}
	}
	if mgr != nil && undo != nil {
		var err error
		if undo.AfterData, err = NewApplyTemplateUndoEditData(sheet); err != nil {
			errs.Log(err)
		} else {
			mgr.Add(undo)
		}
	}
	sheet.Window().ToFront()
	sheet.RequestFocus()
	return true
}

func updateRandomizedProfileFieldsWithoutUndo(sheet *Sheet) {
	e := sheet.Entity()
	updateStringField(sheet, identityPanelNameFieldRefKey, e.Profile.Name)
	updateStringField(sheet, descriptionPanelAgeFieldRefKey, e.Profile.Age)
	updateStringField(sheet, descriptionPanelBirthdayFieldRefKey, e.Profile.Birthday)
	updateStringField(sheet, descriptionPanelEyesFieldRefKey, e.Profile.Eyes)
	updateStringField(sheet, descriptionPanelHairFieldRefKey, e.Profile.Hair)
	updateStringField(sheet, descriptionPanelSkinFieldRefKey, e.Profile.Skin)
	updateStringField(sheet, descriptionPanelHandednessFieldRefKey, e.Profile.Handedness)
	updateStringField(sheet, descriptionPanelGenderFieldRefKey, e.Profile.Gender)
	updateLengthField(sheet, descriptionPanelHeightFieldRefKey, e.Profile.Height)
	updateWeightField(sheet, descriptionPanelWeightFieldRefKey, e.Profile.Weight)
}

func updateStringField(sheet *Sheet, refKey, value string) {
	if panel := sheet.targetMgr.Find(refKey); panel != nil {
		if f, ok := panel.Self.(*StringField); ok {
			saved := sheet.undoMgr
			sheet.undoMgr = nil
			f.SetText(value)
			sheet.undoMgr = saved
		}
	}
}

func updateLengthField(sheet *Sheet, refKey string, value fxp.Length) {
	if panel := sheet.targetMgr.Find(refKey); panel != nil {
		if f, ok := panel.Self.(*LengthField); ok {
			saved := sheet.undoMgr
			sheet.undoMgr = nil
			f.SetText(value.String())
			sheet.undoMgr = saved
		}
	}
}

func updateWeightField(sheet *Sheet, refKey string, value fxp.Weight) {
	if panel := sheet.targetMgr.Find(refKey); panel != nil {
		if f, ok := panel.Self.(*WeightField); ok {
			saved := sheet.undoMgr
			sheet.undoMgr = nil
			f.SetText(value.String())
			sheet.undoMgr = saved
		}
	}
}

func cloneRows[T gurps.NodeTypes](table *unison.Table[*Node[T]], rows []*Node[T]) []*Node[T] {
	rows = slices.Clone(rows)
	for j, row := range rows {
		rows[j] = row.CloneForTarget(table, nil)
	}
	return rows
}

func appendRows[T gurps.NodeTypes](table *unison.Table[*Node[T]], rows []*Node[T]) {
	selMap := make(map[tid.TID]bool)
	orig := slices.Clone(table.RootRows())
	switch t := any(table).(type) {
	case *unison.Table[*Node[*gurps.Skill]]:
		if skillNodes, ok2 := any(orig).([]*Node[*gurps.Skill]); ok2 {
			if rowNodes, ok3 := any(rows).([]*Node[*gurps.Skill]); ok3 {
				if newRows, ok4 := any(mergeSkillRows(t, skillNodes, rowNodes, selMap)).([]*Node[T]); ok4 {
					rows = newRows
				}
			}
		}
	case *unison.Table[*Node[*gurps.Spell]]:
		if spellNodes, ok2 := any(orig).([]*Node[*gurps.Spell]); ok2 {
			if rowNodes, ok3 := any(rows).([]*Node[*gurps.Spell]); ok3 {
				if newRows, ok4 := any(mergeSpellRows(t, spellNodes, rowNodes, selMap)).([]*Node[T]); ok4 {
					rows = newRows
				}
			}
		}
	}
	table.SetRootRows(append(orig, rows...))
	for _, row := range rows {
		selMap[row.ID()] = true
	}
	table.SetSelectionMap(selMap)
	if provider, ok := table.ClientData()[TableProviderClientKey]; ok {
		var tableProvider TableProvider[T]
		if tableProvider, ok = provider.(TableProvider[T]); ok {
			tableProvider.ProcessDropData(nil, table)
		}
	}
}

func mergeSkillRows(skillTable *unison.Table[*Node[*gurps.Skill]], skillNodes, rows []*Node[*gurps.Skill], selMap map[tid.TID]bool) []*Node[*gurps.Skill] {
	skillMap := make(map[uint64]*gurps.Skill)
	gurps.Traverse(func(skill *gurps.Skill) bool {
		skillMap[gurps.Hash64(skill)] = skill
		return false
	}, true, true, ExtractNodeDataFromList(skillNodes)...)
	pruneMap := make(map[*gurps.Skill]bool)
	rowSkills := ExtractNodeDataFromList(rows)
	gurps.Traverse(func(skill *gurps.Skill) bool {
		if s, ok := skillMap[gurps.Hash64(skill)]; ok &&
			maps.Equal(s.Replacements, skill.Replacements) &&
			((s.TechLevel == nil && skill.TechLevel == nil) ||
				(s.TechLevel != nil && skill.TechLevel != nil && *s.TechLevel == *skill.TechLevel)) {
			pruneMap[skill] = true
			s.Points += skill.Points
			selMap[s.ID()] = true
		}
		return false
	}, true, true, rowSkills...)
	for skill := range pruneMap {
		parent := skill.Parent()
		if parent == nil {
			rowSkills = slices.DeleteFunc(rowSkills, func(s *gurps.Skill) bool {
				return s == skill
			})
		} else {
			parent.Children = slices.DeleteFunc(parent.Children, func(s *gurps.Skill) bool {
				return s == skill
			})
		}
	}
	replacements := make([]*Node[*gurps.Skill], 0, len(rowSkills))
	for _, skill := range rowSkills {
		replacements = append(replacements, NewNode(skillTable, nil, skill, true))
	}
	return replacements
}

func mergeSpellRows(spellTable *unison.Table[*Node[*gurps.Spell]], spellNodes, rows []*Node[*gurps.Spell], selMap map[tid.TID]bool) []*Node[*gurps.Spell] {
	spellMap := make(map[uint64]*gurps.Spell)
	gurps.Traverse(func(spell *gurps.Spell) bool {
		spellMap[gurps.Hash64(spell)] = spell
		return false
	}, true, true, ExtractNodeDataFromList(spellNodes)...)
	pruneMap := make(map[*gurps.Spell]bool)
	rowSpells := ExtractNodeDataFromList(rows)
	gurps.Traverse(func(spell *gurps.Spell) bool {
		if s, ok := spellMap[gurps.Hash64(spell)]; ok &&
			maps.Equal(s.Replacements, spell.Replacements) &&
			((s.TechLevel == nil && spell.TechLevel == nil) ||
				(s.TechLevel != nil && spell.TechLevel != nil && *s.TechLevel == *spell.TechLevel)) {
			pruneMap[spell] = true
			s.Points += spell.Points
			selMap[s.ID()] = true
		}
		return false
	}, true, true, rowSpells...)
	for spell := range pruneMap {
		parent := spell.Parent()
		if parent == nil {
			rowSpells = slices.DeleteFunc(rowSpells, func(s *gurps.Spell) bool {
				return s == spell
			})
		} else {
			parent.Children = slices.DeleteFunc(parent.Children, func(s *gurps.Spell) bool {
				return s == spell
			})
		}
	}
	replacements := make([]*Node[*gurps.Spell], 0, len(rowSpells))
	for _, spell := range rowSpells {
		replacements = append(replacements, NewNode(spellTable, nil, spell, true))
	}
	return replacements
}

func rawPoints(child any) fxp.Int {
	switch nc := child.(type) {
	case *gurps.Skill:
		if nc.Container() && nc.TemplatePicker != nil && nc.TemplatePicker.Type == picker.Points &&
			nc.TemplatePicker.Qualifier.Compare == criteria.EqualsNumber {
			return nc.TemplatePicker.Qualifier.Qualifier
		}
		return nc.RawPoints()
	case *gurps.Spell:
		if nc.Container() && nc.TemplatePicker != nil && nc.TemplatePicker.Type == picker.Points &&
			nc.TemplatePicker.Qualifier.Compare == criteria.EqualsNumber {
			return nc.TemplatePicker.Qualifier.Qualifier
		}
		return nc.RawPoints()
	case *gurps.Trait:
		if nc.Container() && nc.TemplatePicker != nil && nc.TemplatePicker.Type == picker.Points &&
			nc.TemplatePicker.Qualifier.Compare == criteria.EqualsNumber {
			return nc.TemplatePicker.Qualifier.Qualifier
		}
		return nc.AdjustedPoints()
	default:
		return 0
	}
}

func (t *Template) installNewItemCmdHandlers(itemID, containerID int, creator itemCreator) {
	variant := NoItemVariant
	if containerID == -1 {
		variant = AlternateItemVariant
	} else {
		t.InstallCmdHandlers(containerID, unison.AlwaysEnabled,
			func(_ any) { creator.CreateItem(t, ContainerItemVariant) })
	}
	t.InstallCmdHandlers(itemID, unison.AlwaysEnabled, func(_ any) { creator.CreateItem(t, variant) })
}

// Entity implements gurps.EntityProvider
func (t *Template) Entity() *gurps.Entity {
	return nil
}

// DockableKind implements widget.DockableKind
func (t *Template) DockableKind() string {
	return TemplateDockableKind
}

// UndoManager implements undo.Provider
func (t *Template) UndoManager() *unison.UndoManager {
	return t.undoMgr
}

// TitleIcon implements workspace.FileBackedDockable
func (t *Template) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  gurps.FileInfoFor(t.path).SVG,
		Size: suggestedSize,
	}
}

// Title implements workspace.FileBackedDockable
func (t *Template) Title() string {
	return fs.BaseName(t.path)
}

func (t *Template) String() string {
	return t.Title()
}

// Tooltip implements workspace.FileBackedDockable
func (t *Template) Tooltip() string {
	return t.path
}

// BackingFilePath implements workspace.FileBackedDockable
func (t *Template) BackingFilePath() string {
	return t.path
}

// SetBackingFilePath implements workspace.FileBackedDockable
func (t *Template) SetBackingFilePath(p string) {
	t.path = p
	UpdateTitleForDockable(t)
}

// Modified implements workspace.FileBackedDockable
func (t *Template) Modified() bool {
	return t.hash != gurps.Hash64(t.template)
}

// MarkModified implements widget.ModifiableRoot.
func (t *Template) MarkModified(_ unison.Paneler) {
	if !t.awaitingUpdate {
		t.awaitingUpdate = true
		h, v := t.scroll.Position()
		focusRefKey := t.targetMgr.CurrentFocusRef()
		DeepSync(t)
		UpdateTitleForDockable(t)
		t.awaitingUpdate = false
		t.searchTracker.Refresh()
		t.targetMgr.ReacquireFocus(focusRefKey, t.toolbar, t.scroll.Content())
		t.scroll.SetPosition(h, v)
	}
}

// MayAttemptClose implements unison.TabCloser
func (t *Template) MayAttemptClose() bool {
	return MayAttemptCloseOfGroup(t)
}

// AttemptClose implements unison.TabCloser
func (t *Template) AttemptClose() bool {
	if AttemptSaveForDockable(t) {
		return AttemptCloseForDockable(t)
	}
	return false
}

func (t *Template) createContent() unison.Paneler {
	t.content = newTemplateContent()
	t.createLists()
	return t.content
}

func (t *Template) save(forceSaveAs bool) bool {
	success := false
	if forceSaveAs || t.needsSaveAsPrompt {
		success = SaveDockableAs(t, gurps.TemplatesExt, t.template.Save, func(path string) {
			t.hash = gurps.Hash64(t.template)
			t.path = path
		})
	} else {
		success = SaveDockable(t, t.template.Save, func() { t.hash = gurps.Hash64(t.template) })
	}
	if success {
		t.needsSaveAsPrompt = false
	}
	return success
}

func (t *Template) createLists() {
	h, v := t.scroll.Position()
	var refocusOnKey string
	var refocusOn unison.Paneler
	if wnd := t.Window(); wnd != nil {
		if focus := wnd.Focus(); focus != nil {
			// For page lists, the focus will be the table, so we need to look up a level
			if focus = focus.Parent(); focus != nil {
				switch focus.Self {
				case t.Traits:
					refocusOnKey = gurps.BlockLayoutTraitsKey
				case t.Skills:
					refocusOnKey = gurps.BlockLayoutSkillsKey
				case t.Spells:
					refocusOnKey = gurps.BlockLayoutSpellsKey
				case t.Equipment:
					refocusOnKey = gurps.BlockLayoutEquipmentKey
				case t.Notes:
					refocusOnKey = gurps.BlockLayoutNotesKey
				}
			}
		}
	}
	t.content.RemoveAllChildren()
	for _, col := range gurps.GlobalSettings().Sheet.BlockLayout.ByRow() {
		rowPanel := unison.NewPanel()
		for _, c := range col {
			switch c {
			case gurps.BlockLayoutTraitsKey:
				if t.Traits == nil {
					t.Traits = NewTraitsPageList(t, t.template)
				} else {
					t.Traits.Sync()
				}
				rowPanel.AddChild(t.Traits)
				if c == refocusOnKey {
					refocusOn = t.Traits.Table
				}
			case gurps.BlockLayoutSkillsKey:
				if t.Skills == nil {
					t.Skills = NewSkillsPageList(t, t.template)
				} else {
					t.Skills.Sync()
				}
				rowPanel.AddChild(t.Skills)
				if c == refocusOnKey {
					refocusOn = t.Skills.Table
				}
			case gurps.BlockLayoutSpellsKey:
				if t.Spells == nil {
					t.Spells = NewSpellsPageList(t, t.template)
				} else {
					t.Spells.Sync()
				}
				rowPanel.AddChild(t.Spells)
				if c == refocusOnKey {
					refocusOn = t.Spells.Table
				}
			case gurps.BlockLayoutEquipmentKey:
				if t.Equipment == nil {
					t.Equipment = NewCarriedEquipmentPageList(t, t.template)
				} else {
					t.Equipment.Sync()
				}
				rowPanel.AddChild(t.Equipment)
				if c == refocusOnKey {
					refocusOn = t.Equipment.Table
				}
			case gurps.BlockLayoutNotesKey:
				if t.Notes == nil {
					t.Notes = NewNotesPageList(t, t.template)
				} else {
					t.Notes.Sync()
				}
				rowPanel.AddChild(t.Notes)
				if c == refocusOnKey {
					refocusOn = t.Notes.Table
				}
			}
		}
		if len(rowPanel.Children()) != 0 {
			rowPanel.SetLayout(&unison.FlexLayout{
				Columns:      len(rowPanel.Children()),
				HSpacing:     1,
				HAlign:       align.Fill,
				EqualColumns: true,
			})
			rowPanel.SetLayoutData(&unison.FlexLayoutData{
				HAlign: align.Fill,
				HGrab:  true,
			})
			t.content.AddChild(rowPanel)
		}
	}

	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
	})
	panel.SetBorder(unison.NewEmptyBorder(unison.Insets{Top: unison.StdVSpacing}))
	button := unison.NewButton()
	button.Font = fonts.PageFieldPrimary
	button.SetTitle(t.lastBody.Name)
	button.ClickCallback = func() { ShowBodySettings(t) }
	if t.template.BodyType == nil {
		button.SetEnabled(false)
	}
	box := NewCheckBox(nil, "", "", func() check.Enum {
		return check.FromBool(t.template.BodyType != nil)
	}, func(state check.Enum) {
		if state == check.On {
			if t.lastBody == nil {
				t.lastBody = gurps.FactoryBody()
			}
			t.template.BodyType = t.lastBody
		} else {
			t.template.BodyType = nil
		}
		button.SetEnabled(state == check.On)
	})
	box.Font = fonts.PageFieldPrimary
	box.SetTitle(i18n.Text("Set Body Type to"))
	panel.AddChild(box)
	panel.AddChild(button)
	t.content.AddChild(panel)

	t.content.ApplyPreferredSize()
	if refocusOn != nil {
		refocusOn.AsPanel().RequestFocus()
	}
	t.scroll.SetPosition(h, v)
}

// SheetSettingsUpdated implements gurps.SheetSettingsResponder.
func (t *Template) SheetSettingsUpdated(e *gurps.Entity, blockLayout bool) {
	if e == nil {
		t.Rebuild(blockLayout)
	}
}

// Rebuild implements widget.Rebuildable.
func (t *Template) Rebuild(full bool) {
	gurps.DiscardGlobalResolveCache()
	t.template.EnsureAttachments()
	t.template.SourceMatcher().PrepareHashes(t.template)
	h, v := t.scroll.Position()
	focusRefKey := t.targetMgr.CurrentFocusRef()
	if full {
		traitsSelMap := t.Traits.RecordSelection()
		skillsSelMap := t.Skills.RecordSelection()
		spellsSelMap := t.Spells.RecordSelection()
		equipmentSelMap := t.Equipment.RecordSelection()
		notesSelMap := t.Notes.RecordSelection()
		defer func() {
			t.Traits.ApplySelection(traitsSelMap)
			t.Skills.ApplySelection(skillsSelMap)
			t.Spells.ApplySelection(spellsSelMap)
			t.Equipment.ApplySelection(equipmentSelMap)
			t.Notes.ApplySelection(notesSelMap)
		}()
		t.createLists()
	}
	DeepSync(t)
	UpdateTitleForDockable(t)
	t.searchTracker.Refresh()
	t.targetMgr.ReacquireFocus(focusRefKey, t.toolbar, t.scroll.Content())
	t.scroll.SetPosition(h, v)
}

type templateTablesUndoData struct {
	traits    *TableUndoEditData[*gurps.Trait]
	skills    *TableUndoEditData[*gurps.Skill]
	spells    *TableUndoEditData[*gurps.Spell]
	equipment *TableUndoEditData[*gurps.Equipment]
	notes     *TableUndoEditData[*gurps.Note]
}

func newTemplateTablesUndoData(t *Template) *templateTablesUndoData {
	return &templateTablesUndoData{
		traits:    NewTableUndoEditData(t.Traits.Table),
		skills:    NewTableUndoEditData(t.Skills.Table),
		spells:    NewTableUndoEditData(t.Spells.Table),
		equipment: NewTableUndoEditData(t.Equipment.Table),
		notes:     NewTableUndoEditData(t.Notes.Table),
	}
}

func (t *templateTablesUndoData) Apply() {
	t.traits.Apply()
	t.skills.Apply()
	t.spells.Apply()
	t.equipment.Apply()
	t.notes.Apply()
}

func (t *Template) syncWithAllSources() {
	var undo *unison.UndoEdit[*templateTablesUndoData]
	mgr := unison.UndoManagerFor(t)
	if mgr != nil {
		undo = &unison.UndoEdit[*templateTablesUndoData]{
			ID:         unison.NextUndoID(),
			EditName:   syncWithSourceAction.Title,
			UndoFunc:   func(e *unison.UndoEdit[*templateTablesUndoData]) { e.BeforeData.Apply() },
			RedoFunc:   func(e *unison.UndoEdit[*templateTablesUndoData]) { e.AfterData.Apply() },
			AbsorbFunc: func(_ *unison.UndoEdit[*templateTablesUndoData], _ unison.Undoable) bool { return false },
			BeforeData: newTemplateTablesUndoData(t),
		}
	}
	t.template.SyncWithLibrarySources()
	t.Traits.Table.SyncToModel()
	t.Skills.Table.SyncToModel()
	t.Spells.Table.SyncToModel()
	t.Equipment.Table.SyncToModel()
	t.Notes.Table.SyncToModel()
	if mgr != nil && undo != nil {
		undo.AfterData = newTemplateTablesUndoData(t)
		mgr.Add(undo)
	}
	t.Rebuild(true)
}

// BodySettingsTitle implements BodySettingsOwner.
func (t *Template) BodySettingsTitle() string {
	return fmt.Sprintf(i18n.Text("Body Type: %s"), t.Title())
}

// BodySettings implements BodySettingsOwner.
func (t *Template) BodySettings(forReset bool) *gurps.Body {
	if forReset {
		return gurps.GlobalSettings().Sheet.BodyType
	}
	return t.lastBody
}

// SetBodySettings implements BodySettingsOwner.
func (t *Template) SetBodySettings(body *gurps.Body) {
	t.lastBody = body
	t.template.BodyType = body
	t.Rebuild(true)
}

func (t *Template) disclosureTables() []disclosureTables {
	return []disclosureTables{
		t.Traits,
		t.Skills,
		t.Spells,
		t.Equipment,
		t.Notes,
	}
}

func (t *Template) toggleHierarchy() {
	tables := t.disclosureTables()
	var open, exists bool
	for _, table := range tables {
		if open, exists = table.FirstDisclosureState(); exists {
			break
		}
	}
	open = !open
	for _, table := range tables {
		table.SetDisclosureState(open)
	}
	t.Rebuild(true)
}

func (t *Template) toggleNotes() {
	tables := t.disclosureTables()
	state := 0
	for _, table := range tables {
		if state = table.FirstNoteState(); state != 0 {
			break
		}
	}
	if state == 0 {
		return
	}
	var closed bool
	if state == 1 {
		closed = true
	}
	for _, table := range tables {
		table.ApplyNoteState(closed)
	}
	t.Rebuild(true)
}
