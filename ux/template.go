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
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/check"
)

var (
	_ FileBackedDockable         = &Template{}
	_ ExportDockable             = &Template{}
	_ unison.UndoManagerProvider = &Template{}
	_ ModifiableRoot             = &Template{}
	_ Rebuildable                = &Template{}
	_ unison.TabCloser           = &Template{}
	_ KeyedDockable              = &Template{}
)

// Template holds the view for a GURPS character template.
type Template struct {
	pageView
	template  *gurps.Template
	content   *templateContent
	Traits    *PageList[*gurps.Trait]
	Skills    *PageList[*gurps.Skill]
	Spells    *PageList[*gurps.Spell]
	Equipment *PageList[*gurps.Equipment]
	Notes     *PageList[*gurps.Note]
	lastBody  *gurps.Body
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
	return openDockableFromFile(filePath, gurps.NewTemplateFromFile, NewTemplate)
}

// NewTemplate creates a new unison.Dockable for GURPS template files.
func NewTemplate(filePath string, template *gurps.Template) *Template {
	t := &Template{
		template: template,
		lastBody: template.BodyType,
	}
	if t.lastBody == nil {
		t.lastBody = gurps.FactoryBody()
	}
	t.initPageDockable(t, filePath, gurps.TemplatesExt, template.Save, template)
	t.finishPageDockable(t, t.createContent())

	installListItemCmdHandlers(t, listItemCreators{
		traits:           func() itemCreator { return t.Traits },
		skills:           func() itemCreator { return t.Skills },
		spells:           func() itemCreator { return t.Spells },
		carriedEquipment: func() itemCreator { return t.Equipment },
		notes:            func() itemCreator { return t.Notes },
	})
	installTraitListCmdHandlers(t, t.template, nil, func() *PageList[*gurps.Trait] { return t.Traits })
	t.InstallCmdHandlers(ApplyTemplateItemID, t.canApplyTemplate, t.applyTemplate)
	t.InstallCmdHandlers(NewSheetFromTemplateItemID, unison.AlwaysEnabled, t.newSheetFromTemplate)
	InstallExportCmdHandlers(t)

	prepareForPage(t.template)
	return t
}

// PageInfoProvider returns the page info provider for this template.
func (t *Template) PageInfoProvider() gurps.PageInfoProvider {
	return t.template
}

func (t *Template) createToolbar() {
	t.toolbar = newToolbar()
	t.AddChild(t.toolbar)
	t.toolbar.AddChild(NewDefaultInfoPop())

	addHelpButton(t.toolbar, "md:User%20Guide/Character%20Templates")

	addUIScaleField(t.toolbar, func() int { return gurps.GlobalSettings().General.InitialSheetUIScale },
		func() int { return t.scale }, func(scale int) { t.scale = scale }, true, t.scroll)

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

	t.searchTracker = installListSearchTracker(t.toolbar, t.lists)

	finishToolbarLayout(t.toolbar)
}

func (t *Template) keyToPanel(key *uti.DataType) *unison.Panel {
	var p unison.Paneler
	switch key {
	case equipmentDragKey:
		p = t.Equipment.Table
	case skillDragKey:
		p = t.Skills.Table
	case spellDragKey:
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

// templateRows holds the rows cloned from a template for insertion into a sheet.
type templateRows struct {
	traits    []*Node[*gurps.Trait]
	skills    []*Node[*gurps.Skill]
	spells    []*Node[*gurps.Spell]
	equipment []*Node[*gurps.Equipment]
	notes     []*Node[*gurps.Note]
}

func (t *Template) applyTemplateToSheet(sheet *Sheet, suppressRandomizePrompt bool) bool {
	return t.applyTemplateToSheetWithPickers(sheet, suppressRandomizePrompt, processTemplatePickers)
}

// applyTemplateToSheetWithPickers applies the template to the sheet, using processPickers to resolve any template
// pickers the template's rows contain. The picker processing is passed in so that headless tests, which have no way to
// respond to the dialogs it would otherwise present, can substitute their own.
func (t *Template) applyTemplateToSheetWithPickers(sheet *Sheet, suppressRandomizePrompt bool, processPickers func(rows *templateRows) bool) bool {
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
	// Nothing from here until the pickers have been dealt with may modify the sheet: canceling a picker abandons the
	// entire operation, which must leave the sheet exactly as it was. That includes the Ancestry question below, which
	// is asked here to preserve the order the questions are presented in, but not acted upon until the operation is
	// known to be going through.
	disableExistingAncestries := false
	templateAncestries := gurps.ActiveAncestries(ExtractNodeDataFromList(t.Traits.Table.RootRows()))
	if len(templateAncestries) != 0 {
		entityAncestries := gurps.ActiveAncestries(e.Traits)
		if len(entityAncestries) != 0 {
			disableExistingAncestries = unison.YesNoDialog(fmt.Sprintf(i18n.Text(`The template contains an Ancestry (%s).
Disable your character's existing Ancestry (%s)?`),
				templateAncestries[0].Name, entityAncestries[0].Name), "") == unison.ModalResponseOK
		}
	}
	rows := &templateRows{
		traits:    cloneRows(sheet.Traits.Table, t.Traits.Table.RootRows()),
		skills:    cloneRows(sheet.Skills.Table, t.Skills.Table.RootRows()),
		spells:    cloneRows(sheet.Spells.Table, t.Spells.Table.RootRows()),
		equipment: cloneRows(sheet.CarriedEquipment.Table, t.Equipment.Table.RootRows()),
		notes:     cloneRows(sheet.Notes.Table, t.Notes.Table.RootRows()),
	}
	if !processPickers(rows) {
		return false // A picker was canceled, so the sheet has been left untouched.
	}
	// The sheet is modified from this point on.
	if t.template.BodyType != nil {
		e.SheetSettings.BodyType = t.template.BodyType.Clone(e, nil)
	}
	if disableExistingAncestries {
		for _, one := range gurps.ActiveAncestryTraits(e.Traits) {
			one.Disabled = true
		}
	}
	// Skills and spells merge points with identical existing rows during appendRows, and the merge match includes the
	// nameable replacements. Since they have no modifiers to toggle, resolve their nameables up front so the merge
	// compares against the final replacements; otherwise re-applying a template would compare empty replacements
	// against the already-resolved existing rows and add duplicates instead of merging.
	ProcessNameables(sheet.Skills.Table, ExtractNodeDataFromList(rows.skills))
	ProcessNameables(sheet.Spells.Table, ExtractNodeDataFromList(rows.spells))
	appendRows(sheet.Traits.Table, rows.traits)
	appendRows(sheet.Skills.Table, rows.skills)
	appendRows(sheet.Spells.Table, rows.spells)
	appendRows(sheet.CarriedEquipment.Table, rows.equipment)
	appendRows(sheet.Notes.Table, rows.notes)
	rebuildAsModified(sheet, true)
	ProcessModifiersForSelection(sheet.Traits.Table)
	ProcessModifiersForSelection(sheet.CarriedEquipment.Table)
	ProcessNameablesForSelection(sheet.Traits.Table)
	ProcessNameablesForSelection(sheet.CarriedEquipment.Table)
	ProcessNameablesForSelection(sheet.Notes.Table)
	maybeClearPreconfiguredFlag(sheet.Traits.Table, sheet.Traits.Table.RootRows())
	maybeClearPreconfiguredFlag(sheet.Skills.Table, sheet.Skills.Table.RootRows())
	maybeClearPreconfiguredFlag(sheet.Spells.Table, sheet.Spells.Table.RootRows())
	maybeClearPreconfiguredFlag(sheet.CarriedEquipment.Table, sheet.CarriedEquipment.Table.RootRows())
	maybeClearPreconfiguredFlag(sheet.Notes.Table, sheet.Notes.Table.RootRows())

	if len(templateAncestries) != 0 && gurps.GlobalSettings().General.AutoFillProfile {
		randomize := true
		if !suppressRandomizePrompt {
			randomize = unison.YesNoDialog(i18n.Text("Would you like to apply the initial randomization again?"), "") == unison.ModalResponseOK
		}
		if randomize {
			e.Profile.ApplyRandomizers(e)
			updateRandomizedProfileFieldsWithoutUndo(sheet)
			rebuildAsModified(sheet, true)
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
			f.SetText(f.Format(value))
			sheet.undoMgr = saved
		}
	}
}

func updateWeightField(sheet *Sheet, refKey string, value fxp.Weight) {
	if panel := sheet.targetMgr.Find(refKey); panel != nil {
		if f, ok := panel.Self.(*WeightField); ok {
			saved := sheet.undoMgr
			sheet.undoMgr = nil
			f.SetText(f.Format(value))
			sheet.undoMgr = saved
		}
	}
}

func cloneRows[T gurps.Node[T]](table *unison.Table[*Node[T]], rows []*Node[T]) []*Node[T] {
	rows = slices.Clone(rows)
	for j, row := range rows {
		rows[j] = row.CloneForTarget(table, nil)
	}
	return rows
}

func appendRows[T gurps.Node[T]](table *unison.Table[*Node[T]], rows []*Node[T]) {
	selMap := make(map[tid.TID]bool)
	orig := slices.Clone(table.RootRows())
	switch t := any(table).(type) {
	case *unison.Table[*Node[*gurps.Skill]]:
		rows = mergeRowsFor(t, orig, rows, selMap)
	case *unison.Table[*Node[*gurps.Spell]]:
		rows = mergeRowsFor(t, orig, rows, selMap)
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

// Entity implements EntityPanel. A template has no entity, so nil is always returned.
func (t *Template) Entity() *gurps.Entity {
	return nil
}

// MarkModified implements ModifiableRoot.
func (t *Template) MarkModified(_ unison.Paneler) {
	t.markModified(t, nil)
}

func (t *Template) createContent() unison.Paneler {
	t.content = newTemplateContent()
	t.createLists()
	return t.content
}

// createLists (re)creates the page's lists from the default block layout, which is the one templates follow. A list is
// only built anew when the columns it has to show no longer match the ones it has, since a table's columns are fixed
// at creation; otherwise the existing list is kept and synced (see syncOrRebuildList). Anything that captured a list
// has to allow for it being replaced -- see installNewItemCmdHandlers. Taking the page apart takes the focus away from
// whichever table held it and moves the scroll position; Rebuild puts both back.
func (t *Template) createLists() {
	// The lists are detached first, so that a list the layout doesn't place is left without a parent. Removing the
	// content's children only detaches the bands, leaving a list nested inside one still pointing at a band nobody can
	// see, and a list's parent is how the search tells one that is on the page from one that isn't.
	for _, list := range t.lists() {
		if !xreflect.IsNil(list) {
			list.AsPanel().RemoveFromParent()
		}
	}
	t.content.RemoveAllChildren()
	layout := gurps.GlobalSettings().Sheet.Layout.Filtered(gurps.IsTemplateBlockKey)
	for _, band := range buildLayoutBands(layout.Root, t.layoutLeaf) {
		t.content.AddChild(band)
	}
	// A list the layout doesn't show is still brought into being, since the rebuild's selection tracking and the
	// dockable's other machinery reach for all five of them without asking whether they are on the page.
	for _, key := range gurps.AllBlockKeys {
		if gurps.IsTemplateBlockKey(key) && !layout.Contains(key) {
			t.layoutLeaf(key)
		}
	}

	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
	})
	panel.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: unison.StdVSpacing}))
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
}

// layoutLeaf returns the panel to show for the block with the given key, or nil if a template has no such block.
func (t *Template) layoutLeaf(key string) unison.Paneler {
	switch key {
	case gurps.BlockTraitsKey:
		return syncOrRebuildList(&t.Traits, func() *PageList[*gurps.Trait] { return NewTraitsPageList(t, t.template) })
	case gurps.BlockSkillsKey:
		return syncOrRebuildList(&t.Skills, func() *PageList[*gurps.Skill] { return NewSkillsPageList(t, t.template) })
	case gurps.BlockSpellsKey:
		return syncOrRebuildList(&t.Spells, func() *PageList[*gurps.Spell] { return NewSpellsPageList(t, t.template) })
	case gurps.BlockEquipmentKey:
		return syncOrRebuildList(&t.Equipment, func() *PageList[*gurps.Equipment] {
			return NewCarriedEquipmentPageList(t, t.template)
		})
	case gurps.BlockNotesKey:
		return syncOrRebuildList(&t.Notes, func() *PageList[*gurps.Note] { return NewNotesPageList(t, t.template) })
	default:
		return nil
	}
}

// list returns the template's list for the given block key, or nil if the key isn't one of the five blocks a template
// can show. A list the template hasn't built yet -- only possible while the template is first being put together --
// comes back as a nil *PageList inside the interface, which xreflect.IsNil sees through.
func (t *Template) list(key string) sheetList {
	switch key {
	case gurps.BlockTraitsKey:
		return t.Traits
	case gurps.BlockSkillsKey:
		return t.Skills
	case gurps.BlockSpellsKey:
		return t.Spells
	case gurps.BlockEquipmentKey:
		return t.Equipment
	case gurps.BlockNotesKey:
		return t.Notes
	default:
		return nil
	}
}

// lists returns the template's five lists, in the canonical block order (see gurps.AllBlockKeys). See list for what
// comes back for a list the template hasn't built yet.
func (t *Template) lists() []sheetList {
	return listsForKeys(t.list, gurps.IsTemplateBlockKey)
}

// SheetSettingsUpdated implements gurps.SheetSettingsResponder.
func (t *Template) SheetSettingsUpdated(e *gurps.Entity, fullRebuild bool) {
	if e == nil {
		t.Rebuild(fullRebuild)
	}
}

// Rebuild implements Rebuildable.
func (t *Template) Rebuild(full bool) {
	gurps.DiscardGlobalResolveCache()
	prepareForPage(t.template)
	state := t.captureViewState()
	if full {
		defer preserveSelections(t.lists)()
		t.createLists()
	}
	t.resync(t, state)
}

func (t *Template) syncWithAllSources() {
	syncWithAllSources(t, t.template, t.Traits, t.Skills, t.Spells, t.Equipment, t.Notes)
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
	rebuildAsModified(t, true)
}

func (t *Template) toggleHierarchy() {
	toggleHierarchy(t.lists()...)
	t.Rebuild(true)
}

func (t *Template) toggleNotes() {
	if toggleNotes(t.lists()...) {
		t.Rebuild(true)
	}
}
