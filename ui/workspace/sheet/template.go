/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package sheet

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/gid"
	gsettings "github.com/richardwilkes/gcs/v5/model/gurps/settings"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/model/theme"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/widget/ntable"
	"github.com/richardwilkes/gcs/v5/ui/workspace"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/slices"
)

var (
	_ workspace.FileBackedDockable = &Template{}
	_ unison.UndoManagerProvider   = &Template{}
	_ widget.ModifiableRoot        = &Template{}
	_ widget.Rebuildable           = &Template{}
	_ widget.DockableKind          = &Template{}
	_ unison.TabCloser             = &Template{}
)

// Template holds the view for a GURPS character template.
type Template struct {
	unison.Panel
	path              string
	targetMgr         *widget.TargetMgr
	undoMgr           *unison.UndoManager
	toolbar           *unison.Panel
	scroll            *unison.ScrollPanel
	template          *gurps.Template
	crc               uint64
	scale             int
	content           *templateContent
	scaleField        *widget.PercentageField
	Traits            *PageList[*gurps.Trait]
	Skills            *PageList[*gurps.Skill]
	Spells            *PageList[*gurps.Spell]
	Equipment         *PageList[*gurps.Equipment]
	Notes             *PageList[*gurps.Note]
	dragReroutePanel  *unison.Panel
	needsSaveAsPrompt bool
}

// OpenTemplates returns the currently open templates.
func OpenTemplates() []*Template {
	var templates []*Template
	ws := workspace.FromWindowOrAny(unison.ActiveWindow())
	ws.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
		for _, one := range dc.Dockables() {
			if template, ok := one.(*Template); ok {
				templates = append(templates, template)
			}
		}
		return false
	})
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
		undoMgr:           unison.NewUndoManager(200, func(err error) { jot.Error(err) }),
		scroll:            unison.NewScrollPanel(),
		template:          template,
		scale:             settings.Global().General.InitialSheetUIScale,
		crc:               template.CRC64(),
		needsSaveAsPrompt: true,
	}
	d.Self = d
	d.targetMgr = widget.NewTargetMgr(d)
	d.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  unison.FillAlignment,
		VAlign:  unison.FillAlignment,
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
	d.DrawOverCallback = func(gc *unison.Canvas, rect unison.Rect) {
		if d.dragReroutePanel != nil {
			r := d.RectFromRoot(d.dragReroutePanel.RectToRoot(d.dragReroutePanel.ContentRect(true)))
			paint := unison.DropAreaColor.Paint(gc, r, unison.Fill)
			paint.SetColorFilter(unison.Alpha30Filter())
			gc.DrawRect(r, paint)
		}
	}

	d.scroll.SetContent(d.createContent(), unison.UnmodifiedBehavior, unison.UnmodifiedBehavior)
	d.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
		VGrab:  true,
	})
	d.scroll.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, theme.PageVoidColor.Paint(gc, rect, unison.Fill))
	}

	scaleTitle := i18n.Text("Scale")
	d.scaleField = widget.NewPercentageField(nil, "", scaleTitle,
		func() int { return d.scale },
		func(v int) {
			d.scale = v
			d.applyScale()
		}, gsettings.InitialUIScaleMin, gsettings.InitialUIScaleMax, false, false)
	d.scaleField.SetMarksModified(false)
	d.scaleField.Tooltip = unison.NewTooltipWithText(scaleTitle)

	d.toolbar = unison.NewPanel()
	d.toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
	d.toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	d.toolbar.AddChild(d.scaleField)
	d.toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(d.toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})

	d.AddChild(d.toolbar)
	d.AddChild(d.scroll)

	d.applyScale()

	d.InstallCmdHandlers(constants.SaveItemID, func(_ any) bool { return d.Modified() }, func(_ any) { d.save(false) })
	d.InstallCmdHandlers(constants.SaveAsItemID, unison.AlwaysEnabled, func(_ any) { d.save(true) })
	d.installNewItemCmdHandlers(constants.NewTraitItemID, constants.NewTraitContainerItemID, d.Traits)
	d.installNewItemCmdHandlers(constants.NewSkillItemID, constants.NewSkillContainerItemID, d.Skills)
	d.installNewItemCmdHandlers(constants.NewTechniqueItemID, -1, d.Skills)
	d.installNewItemCmdHandlers(constants.NewSpellItemID, constants.NewSpellContainerItemID, d.Spells)
	d.installNewItemCmdHandlers(constants.NewRitualMagicSpellItemID, -1, d.Spells)
	d.installNewItemCmdHandlers(constants.NewCarriedEquipmentItemID,
		constants.NewCarriedEquipmentContainerItemID, d.Equipment)
	d.installNewItemCmdHandlers(constants.NewNoteItemID, constants.NewNoteContainerItemID, d.Notes)
	d.InstallCmdHandlers(constants.AddNaturalAttacksItemID, unison.AlwaysEnabled, func(_ any) {
		ntable.InsertItems[*gurps.Trait](d, d.Traits.Table, d.template.TraitList, d.template.SetTraitList,
			func(_ *unison.Table[*ntable.Node[*gurps.Trait]]) []*ntable.Node[*gurps.Trait] {
				return d.Traits.provider.RootRows()
			}, gurps.NewNaturalAttacks(nil, nil))
	})
	d.InstallCmdHandlers(constants.ApplyTemplateItemID, d.canApplyTemplate, d.applyTemplate)

	return d
}

func (d *Template) keyToPanel(key string) *unison.Panel {
	var p unison.Paneler
	switch key {
	case gid.Equipment:
		p = d.Equipment.Table
	case gid.Skill:
		p = d.Skills.Table
	case gid.Spell:
		p = d.Spells.Table
	case gid.Trait:
		p = d.Traits.Table
	case gid.Note:
		p = d.Notes.Table
	default:
		return nil
	}
	return p.AsPanel()
}

func (d *Template) canApplyTemplate(_ any) bool {
	return len(OpenSheets()) > 0
}

func (d *Template) applyTemplate(_ any) {
	for _, sheet := range workspace.PromptForDestination(OpenSheets()) {
		var undo *unison.UndoEdit[*ApplyTemplateUndoEditData]
		mgr := unison.UndoManagerFor(sheet)
		if mgr != nil {
			if beforeData, err := NewApplyTemplateUndoEditData(sheet); err != nil {
				jot.Warn(err)
				mgr = nil
			} else {
				undo = &unison.UndoEdit[*ApplyTemplateUndoEditData]{
					ID:         unison.NextUndoID(),
					EditName:   i18n.Text("Apply Template"),
					UndoFunc:   func(e *unison.UndoEdit[*ApplyTemplateUndoEditData]) { e.BeforeData.Apply() },
					RedoFunc:   func(e *unison.UndoEdit[*ApplyTemplateUndoEditData]) { e.AfterData.Apply() },
					AbsorbFunc: func(e *unison.UndoEdit[*ApplyTemplateUndoEditData], other unison.Undoable) bool { return false },
					BeforeData: beforeData,
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
			return
		}
		if skills, abort = processPickerRows(skills); abort {
			return
		}
		if spells, abort = processPickerRows(spells); abort {
			return
		}
		appendRows(sheet.Traits.Table, traits)
		appendRows(sheet.Skills.Table, skills)
		appendRows(sheet.Spells.Table, spells)
		appendRows(sheet.CarriedEquipment.Table, equipment)
		appendRows(sheet.Notes.Table, notes)
		sheet.Rebuild(true)
		ntable.ProcessModifiersForSelection(sheet.Traits.Table)
		ntable.ProcessModifiersForSelection(sheet.Skills.Table)
		ntable.ProcessModifiersForSelection(sheet.Spells.Table)
		ntable.ProcessModifiersForSelection(sheet.CarriedEquipment.Table)
		ntable.ProcessModifiersForSelection(sheet.Notes.Table)
		ntable.ProcessNameablesForSelection(sheet.Traits.Table)
		ntable.ProcessNameablesForSelection(sheet.Skills.Table)
		ntable.ProcessNameablesForSelection(sheet.Spells.Table)
		ntable.ProcessNameablesForSelection(sheet.CarriedEquipment.Table)
		ntable.ProcessNameablesForSelection(sheet.Notes.Table)
		if mgr != nil && undo != nil {
			var err error
			if undo.AfterData, err = NewApplyTemplateUndoEditData(sheet); err != nil {
				jot.Warn(err)
			} else {
				mgr.Add(undo)
			}
		}
	}
}

func cloneRows[T gurps.NodeTypes](table *unison.Table[*ntable.Node[T]], rows []*ntable.Node[T]) []*ntable.Node[T] {
	rows = slices.Clone(rows)
	for j, row := range rows {
		rows[j] = row.CloneForTarget(table, nil)
	}
	return rows
}

func appendRows[T gurps.NodeTypes](table *unison.Table[*ntable.Node[T]], rows []*ntable.Node[T]) {
	table.SetRootRows(append(slices.Clone(table.RootRows()), rows...))
	selMap := make(map[uuid.UUID]bool, len(rows))
	for _, row := range rows {
		selMap[row.UUID()] = true
	}
	table.SetSelectionMap(selMap)
	if provider, ok := table.ClientData()[ntable.TableProviderClientKey]; ok {
		var tableProvider ntable.TableProvider[T]
		if tableProvider, ok = provider.(ntable.TableProvider[T]); ok {
			tableProvider.ProcessDropData(nil, table)
		}
	}
}

func processPickerRows[T gurps.NodeTypes](rows []*ntable.Node[T]) (revised []*ntable.Node[T], abort bool) {
	for _, one := range ntable.ExtractNodeDataFromList(rows) {
		result, cancel := processPickerRow(one)
		if cancel {
			return nil, true
		}
		for _, replacement := range result {
			revised = append(revised, ntable.NewNodeLike(rows[0], replacement))
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
			result, cancel := processPickerRow(child)
			if cancel {
				return nil, true
			}
			rowChildren = append(rowChildren, result...)
		}
		n.SetChildren(rowChildren)
		ntable.SetParents(rowChildren, row)
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
			if box.State == unison.OnCheckState {
				switch tp.Type {
				case gurps.CountTemplatePickerType:
					total += fxp.One
				case gurps.PointsTemplatePickerType:
					total += rawPoints(children[i])
				}
			}
		}
		dialog.Button(unison.ModalResponseOK).SetEnabled(tp.Qualifier.Matches(total))
	}
	for _, child := range children {
		checkBox := unison.NewCheckBox()
		checkBox.Text = fmt.Sprintf("%v", child)
		if tp.Type == gurps.PointsTemplatePickerType {
			points := rawPoints(child)
			pointsLabel := i18n.Text("points")
			if points == fxp.One {
				pointsLabel = i18n.Text("point")
			}
			checkBox.Text += fmt.Sprintf(" [%s %s]", points.Comma(), pointsLabel)
		}
		checkBox.ClickCallback = callback
		list.AddChild(checkBox)
		boxes = append(boxes, checkBox)
	}

	scroll := unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.NewUniformInsets(1), false))
	scroll.SetContent(list, unison.FillBehavior, unison.FillBehavior)
	scroll.BackgroundInk = unison.ContentColor
	scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
		VGrab:  true,
	})

	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   unison.FillAlignment,
		VAlign:   unison.FillAlignment,
	})
	label := unison.NewLabel()
	label.Text = fmt.Sprintf("%v", row)
	panel.AddChild(label)
	label = unison.NewLabel()
	label.Text = tp.Description()
	panel.AddChild(label)
	panel.AddChild(scroll)

	var err error
	dialog, err = unison.NewDialog(unison.DefaultDialogTheme.QuestionIcon,
		unison.DefaultDialogTheme.QuestionIconInk, panel,
		[]*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfo()})
	if err != nil {
		jot.Error(err)
		return nil, true
	}
	dialog.Button(unison.ModalResponseOK).SetEnabled(false)
	if dialog.RunModal() == unison.ModalResponseCancel {
		return nil, true
	}

	rowChildren := make([]T, 0, len(children))
	for i, box := range boxes {
		if box.State == unison.OnCheckState {
			result, cancel := processPickerRow(children[i])
			if cancel {
				return nil, true
			}
			rowChildren = append(rowChildren, result...)
		}
	}
	return rowChildren, false
}

func rawPoints(child any) fxp.Int {
	switch nc := child.(type) {
	case *gurps.Skill:
		return nc.RawPoints()
	case *gurps.Spell:
		return nc.RawPoints()
	case *gurps.Trait:
		return nc.AdjustedPoints()
	default:
		return 0
	}
}

func (d *Template) installNewItemCmdHandlers(itemID, containerID int, creator itemCreator) {
	variant := ntable.NoItemVariant
	if containerID == -1 {
		variant = ntable.AlternateItemVariant
	} else {
		d.InstallCmdHandlers(containerID, unison.AlwaysEnabled,
			func(_ any) { creator.CreateItem(d, ntable.ContainerItemVariant) })
	}
	d.InstallCmdHandlers(itemID, unison.AlwaysEnabled, func(_ any) { creator.CreateItem(d, variant) })
}

// Entity implements gurps.EntityProvider
func (d *Template) Entity() *gurps.Entity {
	return nil
}

// DockableKind implements widget.DockableKind
func (d *Template) DockableKind() string {
	return widget.TemplateDockableKind
}

func (d *Template) applyScale() {
	d.scroll.Content().AsPanel().SetScale(float32(d.scale) / 100)
	d.scroll.Sync()
}

// UndoManager implements undo.Provider
func (d *Template) UndoManager() *unison.UndoManager {
	return d.undoMgr
}

// TitleIcon implements workspace.FileBackedDockable
func (d *Template) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  library.FileInfoFor(d.path).SVG,
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
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.UpdateTitle(d)
	}
}

// Modified implements workspace.FileBackedDockable
func (d *Template) Modified() bool {
	return d.crc != d.template.CRC64()
}

// MarkModified implements widget.ModifiableRoot.
func (d *Template) MarkModified(_ unison.Paneler) {
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.UpdateTitle(d)
	}
}

// MayAttemptClose implements unison.TabCloser
func (d *Template) MayAttemptClose() bool {
	return workspace.MayAttemptCloseOfGroup(d)
}

// AttemptClose implements unison.TabCloser
func (d *Template) AttemptClose() bool {
	if !workspace.CloseGroup(d) {
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
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.Close(d)
	}
	return true
}

func (d *Template) createContent() unison.Paneler {
	d.content = newTemplateContent()
	d.createLists()
	return d.content
}

func (d *Template) save(forceSaveAs bool) bool {
	success := false
	if forceSaveAs || d.needsSaveAsPrompt {
		success = workspace.SaveDockableAs(d, library.TemplatesExt, d.template.Save, func(path string) {
			d.crc = d.template.CRC64()
			d.path = path
		})
	} else {
		success = workspace.SaveDockable(d, d.template.Save, func() { d.crc = d.template.CRC64() })
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
	for _, col := range settings.Global().Sheet.BlockLayout.ByRow() {
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
				HAlign:       unison.FillAlignment,
				EqualColumns: true,
			})
			rowPanel.SetLayoutData(&unison.FlexLayoutData{
				HAlign: unison.FillAlignment,
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
func (d *Template) SheetSettingsUpdated(entity *gurps.Entity, blockLayout bool) {
	if entity == nil {
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
	widget.DeepSync(d)
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.UpdateTitle(d)
	}
	d.targetMgr.ReacquireFocus(focusRefKey, d.toolbar, d.scroll.Content())
	d.scroll.SetPosition(h, v)
}
