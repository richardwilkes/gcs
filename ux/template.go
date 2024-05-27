// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
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
	"os"
	"path/filepath"
	"slices"

	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/entity"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
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
	crc               uint64
	content           *templateContent
	Traits            *PageList[*gurps.Trait]
	Skills            *PageList[*gurps.Skill]
	Spells            *PageList[*gurps.Spell]
	Equipment         *PageList[*gurps.Equipment]
	Notes             *PageList[*gurps.Note]
	dragReroutePanel  *unison.Panel
	scale             int
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
	d := &Template{
		path:              filePath,
		undoMgr:           unison.NewUndoManager(200, func(err error) { errs.Log(err) }),
		scroll:            unison.NewScrollPanel(),
		template:          template,
		scale:             gurps.GlobalSettings().General.InitialSheetUIScale,
		crc:               template.CRC64(),
		needsSaveAsPrompt: true,
	}
	d.Self = d
	d.targetMgr = NewTargetMgr(d)
	d.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  align.Fill,
		VAlign:  align.Fill,
	})

	d.MouseDownCallback = func(_ unison.Point, _, _ int, _ unison.Modifiers) bool {
		d.RequestFocus()
		return false
	}
	d.DataDragOverCallback = func(_ unison.Point, data map[string]any) bool {
		d.dragReroutePanel = nil
		for _, key := range dropKeys {
			if _, ok := data[key]; ok {
				if d.dragReroutePanel = d.keyToPanel(key); d.dragReroutePanel != nil {
					d.dragReroutePanel.DataDragOverCallback(unison.Point{Y: 100000000}, data)
					return true
				}
				break
			}
		}
		return false
	}
	d.DataDragExitCallback = func() {
		if d.dragReroutePanel != nil {
			d.dragReroutePanel.DataDragExitCallback()
			d.dragReroutePanel = nil
		}
	}
	d.DataDragDropCallback = func(_ unison.Point, data map[string]any) {
		if d.dragReroutePanel != nil {
			d.dragReroutePanel.DataDragDropCallback(unison.Point{Y: 10000000}, data)
			d.dragReroutePanel = nil
		}
	}
	d.DrawOverCallback = func(gc *unison.Canvas, _ unison.Rect) {
		if d.dragReroutePanel != nil {
			r := d.RectFromRoot(d.dragReroutePanel.RectToRoot(d.dragReroutePanel.ContentRect(true)))
			paint := unison.ThemeWarning.Paint(gc, r, paintstyle.Fill)
			paint.SetColorFilter(unison.Alpha30Filter())
			gc.DrawRect(r, paint)
		}
	}

	d.scroll.SetContent(d.createContent(), behavior.Unmodified, behavior.Unmodified)
	d.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})

	helpButton := unison.NewSVGButton(svg.Help)
	helpButton.Tooltip = newWrappedTooltip(i18n.Text("Help"))
	helpButton.ClickCallback = func() { HandleLink(nil, "md:Help/Interface/Character Template") }

	addUserButton := unison.NewSVGButton(svg.Stamper)
	addUserButton.Tooltip = newWrappedTooltip(applyTemplateAction.Title)
	addUserButton.ClickCallback = func() {
		if CanApplyTemplate() {
			d.applyTemplate(nil)
		}
	}

	d.toolbar = unison.NewPanel()
	d.AddChild(d.toolbar)
	d.toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
	d.toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	d.toolbar.AddChild(NewDefaultInfoPop())
	d.toolbar.AddChild(helpButton)
	d.toolbar.AddChild(
		NewScaleField(
			gurps.InitialUIScaleMin,
			gurps.InitialUIScaleMax,
			func() int { return gurps.GlobalSettings().General.InitialSheetUIScale },
			func() int { return d.scale },
			func(scale int) { d.scale = scale },
			nil,
			false,
			d.scroll,
		),
	)
	d.toolbar.AddChild(addUserButton)
	installSearchTracker(d.toolbar, func() {
		d.Traits.Table.ClearSelection()
		d.Skills.Table.ClearSelection()
		d.Spells.Table.ClearSelection()
		d.Equipment.Table.ClearSelection()
		d.Notes.Table.ClearSelection()
	}, func(refList *[]*searchRef, text string, namesOnly bool) {
		searchSheetTable(refList, text, namesOnly, d.Traits)
		searchSheetTable(refList, text, namesOnly, d.Skills)
		searchSheetTable(refList, text, namesOnly, d.Spells)
		searchSheetTable(refList, text, namesOnly, d.Equipment)
		searchSheetTable(refList, text, namesOnly, d.Notes)
	})
	d.toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(d.toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})

	d.AddChild(d.scroll)

	d.InstallCmdHandlers(SaveItemID, func(_ any) bool { return d.Modified() }, func(_ any) { d.save(false) })
	d.InstallCmdHandlers(SaveAsItemID, unison.AlwaysEnabled, func(_ any) { d.save(true) })
	d.installNewItemCmdHandlers(NewTraitItemID, NewTraitContainerItemID, d.Traits)
	d.installNewItemCmdHandlers(NewSkillItemID, NewSkillContainerItemID, d.Skills)
	d.installNewItemCmdHandlers(NewTechniqueItemID, -1, d.Skills)
	d.installNewItemCmdHandlers(NewSpellItemID, NewSpellContainerItemID, d.Spells)
	d.installNewItemCmdHandlers(NewRitualMagicSpellItemID, -1, d.Spells)
	d.installNewItemCmdHandlers(NewCarriedEquipmentItemID,
		NewCarriedEquipmentContainerItemID, d.Equipment)
	d.installNewItemCmdHandlers(NewNoteItemID, NewNoteContainerItemID, d.Notes)
	d.InstallCmdHandlers(AddNaturalAttacksItemID, unison.AlwaysEnabled, func(_ any) {
		InsertItems[*gurps.Trait](d, d.Traits.Table, d.template.TraitList, d.template.SetTraitList,
			func(_ *unison.Table[*Node[*gurps.Trait]]) []*Node[*gurps.Trait] {
				return d.Traits.provider.RootRows()
			}, gurps.NewNaturalAttacks(nil, nil))
	})
	d.InstallCmdHandlers(ApplyTemplateItemID, d.canApplyTemplate, d.applyTemplate)
	d.InstallCmdHandlers(NewSheetFromTemplateItemID, unison.AlwaysEnabled, d.newSheetFromTemplate)

	return d
}

func (d *Template) keyToPanel(key string) *unison.Panel {
	var p unison.Paneler
	switch key {
	case equipmentDragKey:
		p = d.Equipment.Table
	case gurps.SkillID:
		p = d.Skills.Table
	case gurps.SpellID:
		p = d.Spells.Table
	case traitDragKey:
		p = d.Traits.Table
	case noteDragKey:
		p = d.Notes.Table
	default:
		return nil
	}
	return p.AsPanel()
}

// CanApplyTemplate returns true if a template can be applied.
func CanApplyTemplate() bool {
	return len(OpenSheets(nil)) > 0
}

func (d *Template) canApplyTemplate(_ any) bool {
	return CanApplyTemplate()
}

// NewSheetFromTemplate loads the specified template file and creates a new character sheet from it.
func NewSheetFromTemplate(filePath string) {
	d, err := NewTemplateFromFile(filePath)
	if err != nil {
		unison.ErrorDialogWithError(i18n.Text("Unable to load template"), err)
		return
	}
	if t, ok := d.(*Template); ok {
		t.newSheetFromTemplate(nil)
	}
}

func (d *Template) newSheetFromTemplate(_ any) {
	e := gurps.NewEntity(entity.PC)
	sheet := NewSheet(e.Profile.Name+gurps.SheetExt, e)
	DisplayNewDockable(sheet)
	if d.applyTemplateToSheet(sheet, true) {
		sheet.undoMgr.Clear()
		sheet.crc = 0
	}
	sheet.SetBackingFilePath(e.Profile.Name + gurps.SheetExt)
}

// ApplyTemplate loads the specified template file and applies it to a sheet.
func ApplyTemplate(filePath string) {
	d, err := NewTemplateFromFile(filePath)
	if err != nil {
		unison.ErrorDialogWithError(i18n.Text("Unable to load template"), err)
		return
	}
	if CanApplyTemplate() {
		if t, ok := d.(*Template); ok {
			t.applyTemplate(nil)
		}
	}
}

func (d *Template) applyTemplate(suppressRandomizePromptAsBool any) {
	suppressRandomizePrompt, _ := suppressRandomizePromptAsBool.(bool) //nolint:errcheck // The default of false on failure is acceptable
	for _, sheet := range PromptForDestination(OpenSheets(nil)) {
		d.applyTemplateToSheet(sheet, suppressRandomizePrompt)
	}
}

func (d *Template) applyTemplateToSheet(sheet *Sheet, suppressRandomizePrompt bool) bool {
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
	templateAncestries := gurps.ActiveAncestries(ExtractNodeDataFromList(d.Traits.Table.RootRows()))
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
	traits := cloneRows(sheet.Traits.Table, d.Traits.Table.RootRows())
	skills := cloneRows(sheet.Skills.Table, d.Skills.Table.RootRows())
	spells := cloneRows(sheet.Spells.Table, d.Spells.Table.RootRows())
	equipment := cloneRows(sheet.CarriedEquipment.Table, d.Equipment.Table.RootRows())
	notes := cloneRows(sheet.Notes.Table, d.Notes.Table.RootRows())
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
	table.SetRootRows(append(slices.Clone(table.RootRows()), rows...))
	selMap := make(map[uuid.UUID]bool, len(rows))
	for _, row := range rows {
		selMap[row.UUID()] = true
	}
	table.SetSelectionMap(selMap)
	if provider, ok := table.ClientData()[TableProviderClientKey]; ok {
		var tableProvider TableProvider[T]
		if tableProvider, ok = provider.(TableProvider[T]); ok {
			tableProvider.ProcessDropData(nil, table)
		}
	}
}

func processPickerRows[T gurps.NodeTypes](rows []*Node[T]) (revised []*Node[T], abort bool) {
	for _, one := range ExtractNodeDataFromList(rows) {
		result, cancel := processPickerRow(one)
		if cancel {
			return nil, true
		}
		for _, replacement := range result {
			revised = append(revised, NewNodeLike(rows[0], replacement))
		}
	}
	return revised, false
}

func processPickerRow[T gurps.NodeTypes](row T) (revised []T, abort bool) {
	n := gurps.AsNode[T](row)
	if !n.Container() {
		return []T{row}, false
	}
	children := n.NodeChildren()
	tpp, ok := n.(gurps.TemplatePickerProvider)
	if !ok || tpp.TemplatePickerData().ShouldOmit() {
		rowChildren := make([]T, 0, len(children))
		for _, child := range children {
			var result []T
			result, abort = processPickerRow(child)
			if abort {
				return nil, true
			}
			rowChildren = append(rowChildren, result...)
		}
		n.SetChildren(rowChildren)
		SetParents(rowChildren, row)
		return []T{row}, false
	}
	tp := tpp.TemplatePickerData()

	list := unison.NewPanel()
	list.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(unison.StdHSpacing)))
	list.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})

	boxes := make([]*unison.CheckBox, 0, len(children))
	var dialog *unison.Dialog
	callback := func() {
		var total fxp.Int
		for i, box := range boxes {
			if box.State == check.On {
				switch tp.Type {
				case picker.NotApplicable:
				case picker.Count:
					total += fxp.One
				case picker.Points:
					total += rawPoints(children[i])
				}
			}
		}
		dialog.Button(unison.ModalResponseOK).SetEnabled(tp.Qualifier.Matches(total))
	}
	for _, child := range children {
		checkBox := unison.NewCheckBox()
		title := fmt.Sprintf("%v", child)
		if tp.Type == picker.Points {
			points := rawPoints(child)
			pointsLabel := i18n.Text("points")
			if points == fxp.One {
				pointsLabel = i18n.Text("point")
			}
			title += fmt.Sprintf(" [%s %s]", points.Comma(), pointsLabel)
		}
		checkBox.SetTitle(title)
		checkBox.ClickCallback = callback
		boxes = append(boxes, checkBox)
		list.AddChild(checkBox)
	}

	scroll := unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0, unison.NewUniformInsets(1), false))
	scroll.SetContent(list, behavior.Fill, behavior.Fill)
	scroll.BackgroundInk = unison.ThemeSurface
	scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})

	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	label := unison.NewLabel()
	label.SetTitle(fmt.Sprintf("%v", row))
	panel.AddChild(label)
	if notesCapable, hasNotes := any(row).(interface{ Notes() string }); hasNotes {
		if notes := notesCapable.Notes(); notes != "" {
			label = unison.NewLabel()
			label.Font = gurps.FieldSecondaryFont
			label.SetTitle(notes)
			panel.AddChild(label)
		}
	}
	label = unison.NewLabel()
	label.SetTitle(tp.Description())
	label.SetBorder(unison.NewEmptyBorder(unison.Insets{Top: unison.StdVSpacing * 2}))
	panel.AddChild(label)
	panel.AddChild(scroll)

	var err error
	dialog, err = unison.NewDialog(unison.DefaultDialogTheme.QuestionIcon,
		unison.DefaultDialogTheme.QuestionIconInk, panel,
		[]*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfo()})
	if err != nil {
		errs.Log(err)
		return nil, true
	}
	callback()
	if dialog.RunModal() == unison.ModalResponseCancel {
		return nil, true
	}

	rowChildren := make([]T, 0, len(children))
	for i, box := range boxes {
		if box.State == check.On {
			var result []T
			result, abort = processPickerRow(children[i])
			if abort {
				return nil, true
			}
			rowChildren = append(rowChildren, result...)
		}
	}
	SetParents(rowChildren, n.Parent())
	return rowChildren, false
}

func rawPoints(child any) fxp.Int {
	switch nc := child.(type) {
	case *gurps.Skill:
		if nc.Container() && nc.TemplatePicker != nil && nc.TemplatePicker.Type == picker.Points &&
			nc.TemplatePicker.Qualifier.Compare == gurps.EqualsNumber {
			return nc.TemplatePicker.Qualifier.Qualifier
		}
		return nc.RawPoints()
	case *gurps.Spell:
		if nc.Container() && nc.TemplatePicker != nil && nc.TemplatePicker.Type == picker.Points &&
			nc.TemplatePicker.Qualifier.Compare == gurps.EqualsNumber {
			return nc.TemplatePicker.Qualifier.Qualifier
		}
		return nc.RawPoints()
	case *gurps.Trait:
		if nc.Container() && nc.TemplatePicker != nil && nc.TemplatePicker.Type == picker.Points &&
			nc.TemplatePicker.Qualifier.Compare == gurps.EqualsNumber {
			return nc.TemplatePicker.Qualifier.Qualifier
		}
		return nc.AdjustedPoints()
	default:
		return 0
	}
}

func (d *Template) installNewItemCmdHandlers(itemID, containerID int, creator itemCreator) {
	variant := NoItemVariant
	if containerID == -1 {
		variant = AlternateItemVariant
	} else {
		d.InstallCmdHandlers(containerID, unison.AlwaysEnabled,
			func(_ any) { creator.CreateItem(d, ContainerItemVariant) })
	}
	d.InstallCmdHandlers(itemID, unison.AlwaysEnabled, func(_ any) { creator.CreateItem(d, variant) })
}

// Entity implements gurps.EntityProvider
func (d *Template) Entity() *gurps.Entity {
	return nil
}

// DockableKind implements widget.DockableKind
func (d *Template) DockableKind() string {
	return TemplateDockableKind
}

// UndoManager implements undo.Provider
func (d *Template) UndoManager() *unison.UndoManager {
	return d.undoMgr
}

// TitleIcon implements workspace.FileBackedDockable
func (d *Template) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  gurps.FileInfoFor(d.path).SVG,
		Size: suggestedSize,
	}
}

// Title implements workspace.FileBackedDockable
func (d *Template) Title() string {
	return fs.BaseName(d.path)
}

func (d *Template) String() string {
	return d.Title()
}

// Tooltip implements workspace.FileBackedDockable
func (d *Template) Tooltip() string {
	return d.path
}

// BackingFilePath implements workspace.FileBackedDockable
func (d *Template) BackingFilePath() string {
	return d.path
}

// SetBackingFilePath implements workspace.FileBackedDockable
func (d *Template) SetBackingFilePath(p string) {
	d.path = p
	UpdateTitleForDockable(d)
}

// Modified implements workspace.FileBackedDockable
func (d *Template) Modified() bool {
	return d.crc != d.template.CRC64()
}

// MarkModified implements widget.ModifiableRoot.
func (d *Template) MarkModified(_ unison.Paneler) {
	UpdateTitleForDockable(d)
}

// MayAttemptClose implements unison.TabCloser
func (d *Template) MayAttemptClose() bool {
	return MayAttemptCloseOfGroup(d)
}

// AttemptClose implements unison.TabCloser
func (d *Template) AttemptClose() bool {
	if !CloseGroup(d) {
		return false
	}
	if d.Modified() {
		switch unison.YesNoCancelDialog(fmt.Sprintf(i18n.Text("Save changes made to\n%s?"), d.Title()), "") {
		case unison.ModalResponseDiscard:
		case unison.ModalResponseOK:
			if !d.save(false) {
				return false
			}
		case unison.ModalResponseCancel:
			return false
		}
	}
	return AttemptCloseForDockable(d)
}

func (d *Template) createContent() unison.Paneler {
	d.content = newTemplateContent()
	d.createLists()
	return d.content
}

func (d *Template) save(forceSaveAs bool) bool {
	success := false
	if forceSaveAs || d.needsSaveAsPrompt {
		success = SaveDockableAs(d, gurps.TemplatesExt, d.template.Save, func(path string) {
			d.crc = d.template.CRC64()
			d.path = path
		})
	} else {
		success = SaveDockable(d, d.template.Save, func() { d.crc = d.template.CRC64() })
	}
	if success {
		d.needsSaveAsPrompt = false
	}
	return success
}

func (d *Template) createLists() {
	h, v := d.scroll.Position()
	var refocusOnKey string
	var refocusOn unison.Paneler
	if wnd := d.Window(); wnd != nil {
		if focus := wnd.Focus(); focus != nil {
			// For page lists, the focus will be the table, so we need to look up a level
			if focus = focus.Parent(); focus != nil {
				switch focus.Self {
				case d.Traits:
					refocusOnKey = gurps.BlockLayoutTraitsKey
				case d.Skills:
					refocusOnKey = gurps.BlockLayoutSkillsKey
				case d.Spells:
					refocusOnKey = gurps.BlockLayoutSpellsKey
				case d.Equipment:
					refocusOnKey = gurps.BlockLayoutEquipmentKey
				case d.Notes:
					refocusOnKey = gurps.BlockLayoutNotesKey
				}
			}
		}
	}
	d.content.RemoveAllChildren()
	for _, col := range gurps.GlobalSettings().Sheet.BlockLayout.ByRow() {
		rowPanel := unison.NewPanel()
		for _, c := range col {
			switch c {
			case gurps.BlockLayoutTraitsKey:
				if d.Traits == nil {
					d.Traits = NewTraitsPageList(d, d.template)
				} else {
					d.Traits.Sync()
				}
				rowPanel.AddChild(d.Traits)
				if c == refocusOnKey {
					refocusOn = d.Traits.Table
				}
			case gurps.BlockLayoutSkillsKey:
				if d.Skills == nil {
					d.Skills = NewSkillsPageList(d, d.template)
				} else {
					d.Skills.Sync()
				}
				rowPanel.AddChild(d.Skills)
				if c == refocusOnKey {
					refocusOn = d.Skills.Table
				}
			case gurps.BlockLayoutSpellsKey:
				if d.Spells == nil {
					d.Spells = NewSpellsPageList(d, d.template)
				} else {
					d.Spells.Sync()
				}
				rowPanel.AddChild(d.Spells)
				if c == refocusOnKey {
					refocusOn = d.Spells.Table
				}
			case gurps.BlockLayoutEquipmentKey:
				if d.Equipment == nil {
					d.Equipment = NewCarriedEquipmentPageList(d, d.template)
				} else {
					d.Equipment.Sync()
				}
				rowPanel.AddChild(d.Equipment)
				if c == refocusOnKey {
					refocusOn = d.Equipment.Table
				}
			case gurps.BlockLayoutNotesKey:
				if d.Notes == nil {
					d.Notes = NewNotesPageList(d, d.template)
				} else {
					d.Notes.Sync()
				}
				rowPanel.AddChild(d.Notes)
				if c == refocusOnKey {
					refocusOn = d.Notes.Table
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
			d.content.AddChild(rowPanel)
		}
	}
	d.content.ApplyPreferredSize()
	if refocusOn != nil {
		refocusOn.AsPanel().RequestFocus()
	}
	d.scroll.SetPosition(h, v)
}

// SheetSettingsUpdated implements gurps.SheetSettingsResponder.
func (d *Template) SheetSettingsUpdated(e *gurps.Entity, blockLayout bool) {
	if e == nil {
		d.Rebuild(blockLayout)
	}
}

// Rebuild implements widget.Rebuildable.
func (d *Template) Rebuild(full bool) {
	h, v := d.scroll.Position()
	focusRefKey := d.targetMgr.CurrentFocusRef()
	if full {
		traitsSelMap := d.Traits.RecordSelection()
		skillsSelMap := d.Skills.RecordSelection()
		spellsSelMap := d.Spells.RecordSelection()
		equipmentSelMap := d.Equipment.RecordSelection()
		notesSelMap := d.Notes.RecordSelection()
		defer func() {
			d.Traits.ApplySelection(traitsSelMap)
			d.Skills.ApplySelection(skillsSelMap)
			d.Spells.ApplySelection(spellsSelMap)
			d.Equipment.ApplySelection(equipmentSelMap)
			d.Notes.ApplySelection(notesSelMap)
		}()
		d.createLists()
	}
	DeepSync(d)
	UpdateTitleForDockable(d)
	d.targetMgr.ReacquireFocus(focusRefKey, d.toolbar, d.scroll.Content())
	d.scroll.SetPosition(h, v)
}
