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
	"maps"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
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
	fileBackedPanel
	targetMgr      *TargetMgr
	undoMgr        *unison.UndoManager
	toolbar        *unison.Panel
	scroll         *unison.ScrollPanel
	template       *gurps.Template
	content        *templateContent
	Traits         *PageList[*gurps.Trait]
	Skills         *PageList[*gurps.Skill]
	Spells         *PageList[*gurps.Spell]
	Equipment      *PageList[*gurps.Equipment]
	Notes          *PageList[*gurps.Note]
	lastBody       *gurps.Body
	searchTracker  *SearchTracker
	scale          int
	awaitingUpdate bool
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
		undoMgr:  unison.NewUndoManager(200, func(err error) { errs.Log(err) }),
		scroll:   unison.NewScrollPanel(),
		template: template,
		lastBody: template.BodyType,
		scale:    gurps.GlobalSettings().General.InitialSheetUIScale,
	}
	if t.lastBody == nil {
		t.lastBody = gurps.FactoryBody()
	}
	t.Self = t
	t.initFileEditor(t, filePath, gurps.TemplatesExt, template.Save, template)
	t.targetMgr = NewTargetMgr(t)
	t.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  align.Fill,
		VAlign:  align.Fill,
	})

	installDropRerouting(t.AsPanel(), dropKeys, t.keyToPanel)

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
	t.installNewItemCmdHandlers(NewTraitItemID, NewTraitContainerItemID, func() itemCreator { return t.Traits })
	t.installNewItemCmdHandlers(NewSkillItemID, NewSkillContainerItemID, func() itemCreator { return t.Skills })
	t.installNewItemCmdHandlers(NewTechniqueItemID, -1, func() itemCreator { return t.Skills })
	t.installNewItemCmdHandlers(NewSpellItemID, NewSpellContainerItemID, func() itemCreator { return t.Spells })
	t.installNewItemCmdHandlers(NewRitualMagicSpellItemID, -1, func() itemCreator { return t.Spells })
	t.installNewItemCmdHandlers(NewCarriedEquipmentItemID, NewCarriedEquipmentContainerItemID,
		func() itemCreator { return t.Equipment })
	t.installNewItemCmdHandlers(NewNoteItemID, NewNoteContainerItemID, func() itemCreator { return t.Notes })
	t.InstallCmdHandlers(AddNaturalAttacksItemID, unison.AlwaysEnabled, func(_ any) {
		InsertItems(t, t.Traits.Table, t.template.TraitList, t.template.SetTraitList,
			func(_ *unison.Table[*Node[*gurps.Trait]]) []*Node[*gurps.Trait] {
				return t.Traits.provider.RootRows()
			}, gurps.NewNaturalAttacks(nil, nil))
	})
	t.InstallCmdHandlers(OrganizeTraitsItemID, unison.AlwaysEnabled, func(_ any) { organizeTraits(t, t.Traits.Table) })
	t.InstallCmdHandlers(ApplyTemplateItemID, t.canApplyTemplate, t.applyTemplate)
	t.InstallCmdHandlers(NewSheetFromTemplateItemID, unison.AlwaysEnabled, t.newSheetFromTemplate)
	InstallExportCmdHandlers(t)

	t.template.EnsureAttachments()
	t.template.SourceMatcher().PrepareHashes(t.template)
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

// processTemplatePickers presents the template picker dialog for each row that has one, replacing the rows with the
// resulting choices. It returns false if the user canceled one of them, in which case the rows must be discarded.
func processTemplatePickers(rows *templateRows) bool {
	var abort bool
	if rows.traits, abort = processPickerRows(rows.traits); abort {
		return false
	}
	if rows.skills, abort = processPickerRows(rows.skills); abort {
		return false
	}
	if rows.spells, abort = processPickerRows(rows.spells); abort {
		return false
	}
	return true
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
	// entire operation, which must leave the sheet exactly as it was. That includes the answer to the Ancestry question
	// below, which is asked here to preserve the order the questions are presented in, but not acted upon until the
	// operation is known to be going through.
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
	// nameable replacements. Since skills and spells have no modifiers to toggle, resolve their nameables up front so
	// that the merge compares against the final replacements; otherwise re-applying a template would compare empty
	// replacements against the already-resolved existing rows and add duplicates instead of merging.
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
	clearPreconfiguredFlag(sheet.Traits.Table, sheet.Traits.Table.RootRows())
	clearPreconfiguredFlag(sheet.Skills.Table, sheet.Skills.Table.RootRows())
	clearPreconfiguredFlag(sheet.Spells.Table, sheet.Spells.Table.RootRows())
	clearPreconfiguredFlag(sheet.CarriedEquipment.Table, sheet.CarriedEquipment.Table.RootRows())
	clearPreconfiguredFlag(sheet.Notes.Table, sheet.Notes.Table.RootRows())

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

// entityTechLevel returns the tech level of the entity that owns the table, or an empty string if there is none. This
// is the value substituted for an empty tech level when merging rows, matching what the table providers do on drop.
func entityTechLevel[T gurps.Node[T]](table *unison.Table[*Node[T]]) string {
	if dataOwnerProvider := table.Ancestor[gurps.DataOwnerProvider](); !xreflect.IsNil(dataOwnerProvider) {
		if dataOwner := dataOwnerProvider.DataOwner(); !xreflect.IsNil(dataOwner) {
			if entity := dataOwner.OwningEntity(); entity != nil {
				return entity.Profile.TechLevel
			}
		}
	}
	return ""
}

// resolveEmptyTechLevel replaces an empty (but present) tech level with defaultTechLevel. This is the substitution
// performed when a row is dropped onto a sheet and when a template's rows are merged into one, so both must agree.
func resolveEmptyTechLevel(item gurps.TechLevelProvider, defaultTechLevel string) {
	if item.RequiresTL() && item.TL() == "" {
		item.SetTL(defaultTechLevel)
	}
}

// sameTechLevel reports whether two rows' tech levels match for merge purposes: both absent, or both present and equal.
func sameTechLevel(a, b gurps.TechLevelProvider) bool {
	if a.RequiresTL() != b.RequiresTL() {
		return false
	}
	return !a.RequiresTL() || a.TL() == b.TL()
}

// pointsMergeable is the set of row types whose identical rows are merged by folding their points together: skills and
// spells.
type pointsMergeable[T gurps.Node[T]] interface {
	gurps.Node[T]
	gurps.TechLevelProvider
	RawPoints() fxp.Int
	SetRawPoints(points fxp.Int) bool
	NameableReplacements() map[string]string
}

// mergeRowsFor merges the incoming rows into the existing ones when the table holds a mergeable row type (see
// mergeRows). It exists to bridge from a caller generic over any node type, which only knows the row type once it has
// switched on the table's concrete type, to mergeRows, which needs that concrete type; the conversions cannot fail
// once the switch has matched, but rows are returned untouched should one somehow not hold up.
func mergeRowsFor[T pointsMergeable[T], U gurps.Node[U]](table *unison.Table[*Node[T]], existing, rows []*Node[U], selMap map[tid.TID]bool) []*Node[U] {
	if existingNodes, ok := any(existing).([]*Node[T]); ok {
		if rowNodes, ok2 := any(rows).([]*Node[T]); ok2 {
			if merged, ok3 := any(mergeRows(table, existingNodes, rowNodes, selMap)).([]*Node[U]); ok3 {
				return merged
			}
		}
	}
	return rows
}

// mergeRows folds the points of the incoming rows into matching existing rows (see mergePoints) and returns fresh
// nodes for the incoming rows that survived, ready to be added to the table.
func mergeRows[T pointsMergeable[T]](table *unison.Table[*Node[T]], existing, rows []*Node[T], selMap map[tid.TID]bool) []*Node[T] {
	surviving := mergePoints(ExtractNodeDataFromList(existing), ExtractNodeDataFromList(rows), entityTechLevel(table),
		selMap)
	replacements := make([]*Node[T], 0, len(surviving))
	for _, item := range surviving {
		replacements = append(replacements, NewNode(table, nil, item, true))
	}
	return replacements
}

// mergePoints folds the points of each incoming row into a matching row, returning the incoming rows that had no match
// (and should therefore be added as new rows). A match requires an identical hash, the same nameable replacements, and
// the same tech level. Since neither the tech level nor the replacements are part of the hash, several rows can share
// a hash, so all candidates for a hash are considered rather than just one. An incoming row can match either an
// existing row or an earlier incoming row, so a template that itself contains two identical entries collapses them
// into one just as it merges into what is already on the sheet.
//
// An incoming row with an empty (but non-nil) tech level has it resolved to defaultTechLevel first, mirroring the
// substitution performed on drop by the skills and spells providers. Without this, a template applied a second time
// would compare the incoming empty tech level against the already-resolved tech level of the existing row, fail to
// match, and add a duplicate row instead of merging.
//
// Folding the points through SetRawPoints recomputes the level of the row merged into, which the caller's subsequent
// rebuild does again; the extra pass is harmless and keeps this free of knowledge about how each type stores its points.
func mergePoints[T pointsMergeable[T]](existing, incoming []T, defaultTechLevel string, selMap map[tid.TID]bool) []T {
	byHash := make(map[uint64][]T)
	gurps.Traverse(func(item T) bool {
		hash := gurps.Hash64(item)
		byHash[hash] = append(byHash[hash], item)
		return false
	}, true, true, existing...)
	pruneMap := make(map[T]bool)
	gurps.Traverse(func(item T) bool {
		resolveEmptyTechLevel(item, defaultTechLevel)
		hash := gurps.Hash64(item)
		matched := false
		for _, candidate := range byHash[hash] {
			if !maps.Equal(candidate.NameableReplacements(), item.NameableReplacements()) ||
				!sameTechLevel(candidate, item) {
				continue
			}
			pruneMap[item] = true
			candidate.SetRawPoints(candidate.RawPoints() + item.RawPoints())
			selMap[candidate.ID()] = true
			matched = true
			break
		}
		if !matched {
			// Register this surviving incoming row so that any later identical incoming row merges into it.
			byHash[hash] = append(byHash[hash], item)
		}
		return false
	}, true, true, incoming...)
	for item := range pruneMap {
		isItem := func(other T) bool { return other == item }
		if parent := item.Parent(); xreflect.IsNil(parent) {
			incoming = slices.DeleteFunc(incoming, isItem)
		} else {
			parent.SetChildren(slices.DeleteFunc(parent.NodeChildren(), isItem))
		}
	}
	return incoming
}

// MergeAddedRows folds the points of the newly-added, currently-selected top-level rows (and any rows nested within
// them, such as the contents of an added container) into identical skill or spell rows already present in the sheet,
// removing the now-redundant new rows. It applies the same merge used when a template is applied, so that dragging or
// copying a skill or spell that already exists on the sheet adds to its points rather than creating a duplicate. It
// must be called only after tech levels and nameables have been resolved on the new rows, since the match includes
// both. Only skills and spells are affected; other row types are left untouched.
func MergeAddedRows[T gurps.Node[T]](table *unison.Table[*Node[T]]) {
	switch t := any(table).(type) {
	case *unison.Table[*Node[*gurps.Skill]]:
		mergeNewlySelectedRows(t)
	case *unison.Table[*Node[*gurps.Spell]]:
		mergeNewlySelectedRows(t)
	}
}

func mergeNewlySelectedRows[T pointsMergeable[T]](table *unison.Table[*Node[T]]) {
	sel := table.CopySelectionMap()
	if len(sel) == 0 {
		return
	}
	roots := table.RootRows()
	existing := make([]T, 0, len(roots))
	incoming := make([]T, 0, len(roots))
	for _, node := range roots {
		if sel[node.ID()] {
			incoming = append(incoming, node.Data())
		} else {
			existing = append(existing, node.Data())
		}
	}
	if len(existing) == 0 || len(incoming) == 0 {
		return
	}
	newSel := make(map[tid.TID]bool)
	surviving := mergePoints(existing, incoming, entityTechLevel(table), newSel)
	if len(newSel) == 0 {
		return // Nothing merged, so leave the table untouched.
	}
	// Note that comparing len(surviving) to len(incoming) is not a valid way to detect that nothing merged: a merged
	// row nested inside an added container is pruned from the container's data without changing the top-level count,
	// and the view must still be refreshed or the pruned row remains visible until the next rebuild.
	survivingSet := make(map[T]bool, len(surviving))
	for _, data := range surviving {
		survivingSet[data] = true
	}
	newRoots := make([]*Node[T], 0, len(roots))
	for _, node := range roots {
		if sel[node.ID()] {
			if !survivingSet[node.Data()] {
				continue // This newly-added row was merged into an existing one, so drop it.
			}
			newSel[node.ID()] = true // Keep the surviving new rows selected alongside the rows they merged into.
			// Rows nested inside this row may have been pruned by the merge, so discard any cached child nodes to
			// force them to be rebuilt from the updated data.
			node.RefreshChildren()
		}
		newRoots = append(newRoots, node)
	}
	table.SetRootRows(newRoots)
	table.SetSelectionMap(newSel)
	rebuildAsModified(table.AncestorOrSelf[Rebuildable](), true)
}

func rawPoints[T gurps.Node[T]](child T) fxp.Int {
	if xreflect.IsNil(child) {
		return 0
	}
	if child.Container() {
		if pickable, ok := any(child).(gurps.TemplatePickerProvider); ok {
			if _, tp := pickable.TemplatePickerData(); tp.Type == picker.Points {
				if tp.Qualifier.Compare == criteria.EqualsNumber {
					return tp.Qualifier.Qualifier
				}
			}
		}
	}
	// Covers skills and spells
	if rp, ok := any(child).(interface{ RawPoints() fxp.Int }); ok {
		return rp.RawPoints()
	}
	// Covers traits
	if rp, ok := any(child).(interface{ AdjustedPoints() fxp.Int }); ok {
		return rp.AdjustedPoints()
	}
	// Fallback
	return 0
}

// installNewItemCmdHandlers installs the handlers for the "New ..." commands that add an item to one of the template's
// lists. As on a character sheet, the list is looked up through the getter each time a command is invoked rather than
// captured here, since a list whose set of columns has to change can only do so by being replaced outright (a table's
// columns are fixed at creation -- see PageList.needReconstruction). A template's lists never grow the switch column,
// which is reserved for character sheets (see showSwitchColumn), but the equipment list's TL and LC columns follow the
// global sheet settings, which the user can change while the template is open. See Sheet.installNewItemCmdHandlers for
// what goes wrong when a captured list is left orphaned by such a replacement.
func (t *Template) installNewItemCmdHandlers(itemID, containerID int, creator func() itemCreator) {
	variant := NoItemVariant
	if containerID == -1 {
		variant = AlternateItemVariant
	} else {
		t.InstallCmdHandlers(containerID, unison.AlwaysEnabled,
			func(_ any) { creator().CreateItem(t, ContainerItemVariant) })
	}
	t.InstallCmdHandlers(itemID, unison.AlwaysEnabled, func(_ any) { creator().CreateItem(t, variant) })
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

func (t *Template) createContent() unison.Paneler {
	t.content = newTemplateContent()
	t.createLists()
	return t.content
}

// createLists (re)creates the page's lists from the default block layout, which is the one templates follow. A list is
// only built anew when the columns it has to show no longer match the ones it has, since a table's columns are fixed
// at creation; otherwise the existing list is kept and synced (see syncOrRebuildList). Anything that captured a list
// has to allow for it being replaced -- see installNewItemCmdHandlers. Taking the page apart takes the focus away from
// whichever table held it; Rebuild puts it back by the table's reference key, which a replacement table shares.
func (t *Template) createLists() {
	h, v := t.scroll.Position()
	// The lists are detached first, so that a list the layout doesn't place is left without a parent. Removing the
	// content's children only detaches the bands, which would leave a list nested inside one still pointing at a band
	// nobody can see, and a list's parent is how the search tells one that is on the page from one that isn't.
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
	t.scroll.SetPosition(h, v)
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
// can show. A list the template hasn't built yet -- which is only the case while the template is first being put
// together -- comes back as a nil *PageList inside the interface, which xreflect.IsNil sees through.
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

// Rebuild implements widget.Rebuildable.
func (t *Template) Rebuild(full bool) {
	gurps.DiscardGlobalResolveCache()
	t.template.EnsureAttachments()
	t.template.SourceMatcher().PrepareHashes(t.template)
	h, v := t.scroll.Position()
	focusRefKey := t.targetMgr.CurrentFocusRef()
	if full {
		defer preserveSelections(t.lists)()
		t.createLists()
	}
	DeepSync(t)
	UpdateTitleForDockable(t)
	t.searchTracker.Refresh()
	t.targetMgr.ReacquireFocus(focusRefKey, t.toolbar, t.scroll.Content())
	t.scroll.SetPosition(h, v)
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
	tables := t.lists()
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
	tables := t.lists()
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
