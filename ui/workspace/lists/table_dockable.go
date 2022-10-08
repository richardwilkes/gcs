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

package lists

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/crc"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	gsettings "github.com/richardwilkes/gcs/v5/model/gurps/settings"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/widget/ntable"
	"github.com/richardwilkes/gcs/v5/ui/workspace"
	"github.com/richardwilkes/gcs/v5/ui/workspace/editors"
	"github.com/richardwilkes/gcs/v5/ui/workspace/sheet"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
)

var (
	_ workspace.FileBackedDockable = &TableDockable[*gurps.Trait]{}
	_ unison.UndoManagerProvider   = &TableDockable[*gurps.Trait]{}
	_ widget.ModifiableRoot        = &TableDockable[*gurps.Trait]{}
	_ widget.Rebuildable           = &TableDockable[*gurps.Trait]{}
	_ widget.DockableKind          = &TableDockable[*gurps.Trait]{}
	_ unison.TabCloser             = &TableDockable[*gurps.Trait]{}
)

// TableDockable holds the view for a file that contains a (potentially hierarchical) list of data.
type TableDockable[T gurps.NodeTypes] struct {
	unison.Panel
	path              string
	extension         string
	undoMgr           *unison.UndoManager
	provider          ntable.TableProvider[T]
	saver             func(path string) error
	canCreateIDs      map[int]bool
	hierarchyButton   *unison.Button
	sizeToFitButton   *unison.Button
	scale             int
	scaleField        *widget.PercentageField
	backButton        *unison.Button
	forwardButton     *unison.Button
	filterPopup       *unison.PopupMenu[string]
	searchField       *unison.Field
	matchesLabel      *unison.Label
	scroll            *unison.ScrollPanel
	tableHeader       *unison.TableHeader[*ntable.Node[T]]
	table             *unison.Table[*ntable.Node[T]]
	crc               uint64
	searchResult      []*ntable.Node[T]
	searchIndex       int
	needsSaveAsPrompt bool
}

// NewTableDockable creates a new TableDockable for list data files.
func NewTableDockable[T gurps.NodeTypes](filePath, extension string, provider ntable.TableProvider[T], saver func(path string) error, canCreateIDs ...int) *TableDockable[T] {
	header, table := ntable.NewNodeTable[T](provider, nil)
	d := &TableDockable[T]{
		path:              filePath,
		extension:         extension,
		undoMgr:           unison.NewUndoManager(200, func(err error) { jot.Error(err) }),
		provider:          provider,
		saver:             saver,
		canCreateIDs:      make(map[int]bool),
		scroll:            unison.NewScrollPanel(),
		tableHeader:       header,
		table:             table,
		scale:             settings.Global().General.InitialListUIScale,
		needsSaveAsPrompt: true,
	}
	d.Self = d
	d.SetLayout(&unison.FlexLayout{Columns: 1})

	for _, id := range canCreateIDs {
		d.canCreateIDs[id] = true
	}

	d.table.SyncToModel()
	d.table.SizeColumnsToFit(true)
	ntable.InstallTableDropSupport(d.table, d.provider)

	d.scroll.SetColumnHeader(d.tableHeader)
	d.scroll.SetContent(d.table, unison.FillBehavior, unison.FillBehavior)
	d.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
		VGrab:  true,
	})

	d.AddChild(d.createToolbar())
	d.AddChild(d.scroll)

	d.applyScale()

	d.InstallCmdHandlers(constants.OpenEditorItemID,
		func(_ any) bool { return d.table.HasSelection() },
		func(_ any) { d.provider.OpenEditor(d, d.table) })
	d.InstallCmdHandlers(constants.OpenOnePageReferenceItemID,
		func(_ any) bool { return editors.CanOpenPageRef(d.table) },
		func(_ any) { editors.OpenPageRef(d.table) })
	d.InstallCmdHandlers(constants.OpenEachPageReferenceItemID,
		func(_ any) bool { return editors.CanOpenPageRef(d.table) },
		func(_ any) { editors.OpenEachPageRef(d.table) })
	d.InstallCmdHandlers(constants.SaveItemID,
		func(_ any) bool { return d.Modified() },
		func(_ any) { d.save(false) })
	d.InstallCmdHandlers(constants.SaveAsItemID, unison.AlwaysEnabled, func(_ any) { d.save(true) })
	d.InstallCmdHandlers(unison.DeleteItemID,
		func(_ any) bool { return d.table.HasSelection() },
		func(_ any) { ntable.DeleteSelection(d.table) })
	d.InstallCmdHandlers(constants.DuplicateItemID,
		func(_ any) bool { return d.table.HasSelection() },
		func(_ any) { ntable.DuplicateSelection(d.table) })
	d.InstallCmdHandlers(constants.CopyToSheetItemID, d.canCopySelectionToSheet, d.copySelectionToSheet)
	d.InstallCmdHandlers(constants.CopyToTemplateItemID, d.canCopySelectionToTemplate, d.copySelectionToTemplate)
	for _, id := range canCreateIDs {
		variant := ntable.ItemVariant(-1)
		switch {
		case id > constants.FirstNonContainerMarker && id < constants.LastNonContainerMarker:
			variant = ntable.NoItemVariant
		case id > constants.FirstContainerMarker && id < constants.LastContainerMarker:
			variant = ntable.ContainerItemVariant
		case id > constants.FirstAlternateNonContainerMarker && id < constants.LastAlternateNonContainerMarker:
			variant = ntable.AlternateItemVariant
		}
		if variant != -1 {
			d.InstallCmdHandlers(id, unison.AlwaysEnabled,
				func(_ any) { d.provider.CreateItem(d, d.table, variant) })
		}
	}

	d.crc = d.crc64()
	return d
}

func (d *TableDockable[T]) createToolbar() *unison.Panel {
	scaleTitle := i18n.Text("Scale")
	d.scaleField = widget.NewPercentageField(nil, "", scaleTitle,
		func() int { return d.scale },
		func(v int) {
			d.scale = v
			d.applyScale()
		}, gsettings.InitialUIScaleMin, gsettings.InitialUIScaleMax, false, false)
	d.scaleField.Tooltip = unison.NewTooltipWithText(scaleTitle)

	d.hierarchyButton = unison.NewSVGButton(res.HierarchySVG)
	d.hierarchyButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Opens/closes all hierarchical rows"))
	d.hierarchyButton.ClickCallback = d.toggleHierarchy

	d.sizeToFitButton = unison.NewSVGButton(res.SizeToFitSVG)
	d.sizeToFitButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Sets the width of each column to fit its contents"))
	d.sizeToFitButton.ClickCallback = d.sizeToFit

	d.backButton = unison.NewSVGButton(res.BackSVG)
	d.backButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Previous Match"))
	d.backButton.ClickCallback = d.previousMatch
	d.backButton.SetEnabled(false)

	d.forwardButton = unison.NewSVGButton(res.ForwardSVG)
	d.forwardButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Next Match"))
	d.forwardButton.ClickCallback = d.nextMatch
	d.forwardButton.SetEnabled(false)

	d.filterPopup = unison.NewPopupMenu[string]()
	d.filterPopup.Tooltip = unison.NewTooltipWithText(i18n.Text("Tag Filter"))
	d.filterPopup.AddItem("")
	for _, tag := range d.provider.Tags() {
		d.filterPopup.AddItem(tag)
	}
	d.filterPopup.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})

	d.searchField = unison.NewField()
	search := i18n.Text("Search")
	d.searchField.Watermark = search
	d.searchField.Tooltip = unison.NewTooltipWithText(search)
	d.searchField.ModifiedCallback = d.searchModified
	d.searchField.KeyDownCallback = func(keyCode unison.KeyCode, mod unison.Modifiers, repeat bool) bool {
		if keyCode == unison.KeyReturn || keyCode == unison.KeyNumPadEnter {
			if mod.ShiftDown() {
				d.previousMatch()
			} else {
				d.nextMatch()
			}
			return true
		}
		return d.searchField.DefaultKeyDown(keyCode, mod, repeat)
	}
	d.searchField.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
		HGrab:  true,
	})

	d.matchesLabel = unison.NewLabel()
	d.matchesLabel.Text = "-"
	d.matchesLabel.Tooltip = unison.NewTooltipWithText(i18n.Text("Number of matches found"))

	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
	toolbar.AddChild(d.scaleField)
	toolbar.AddChild(d.hierarchyButton)
	toolbar.AddChild(d.sizeToFitButton)
	toolbar.AddChild(d.filterPopup)
	toolbar.AddChild(d.backButton)
	toolbar.AddChild(d.forwardButton)
	toolbar.AddChild(d.searchField)
	toolbar.AddChild(d.matchesLabel)
	toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})
	return toolbar
}

// Entity implements gurps.EntityProvider
func (d *TableDockable[T]) Entity() *gurps.Entity {
	return nil
}

// UndoManager implements undo.Provider
func (d *TableDockable[T]) UndoManager() *unison.UndoManager {
	return d.undoMgr
}

// DockableKind implements widget.DockableKind
func (d *TableDockable[T]) DockableKind() string {
	return widget.ListDockableKind
}

func (d *TableDockable[T]) applyScale() {
	s := float32(d.scale) / 100
	d.tableHeader.SetScale(s)
	d.table.SetScale(s)
	d.scroll.Sync()
}

// TitleIcon implements workspace.FileBackedDockable
func (d *TableDockable[T]) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  library.FileInfoFor(d.path).SVG,
		Size: suggestedSize,
	}
}

// Title implements workspace.FileBackedDockable
func (d *TableDockable[T]) Title() string {
	return fs.BaseName(d.path)
}

func (d *TableDockable[T]) String() string {
	return d.Title()
}

// Tooltip implements workspace.FileBackedDockable
func (d *TableDockable[T]) Tooltip() string {
	return d.path
}

// BackingFilePath implements workspace.FileBackedDockable
func (d *TableDockable[T]) BackingFilePath() string {
	return d.path
}

// SetBackingFilePath implements workspace.FileBackedDockable
func (d *TableDockable[T]) SetBackingFilePath(p string) {
	d.path = p
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.UpdateTitle(d)
	}
}

// Modified implements workspace.FileBackedDockable
func (d *TableDockable[T]) Modified() bool {
	return d.crc != d.crc64()
}

// MarkModified implements widget.ModifiableRoot.
func (d *TableDockable[T]) MarkModified(_ unison.Paneler) {
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.UpdateTitle(d)
	}
}

// MayAttemptClose implements unison.TabCloser
func (d *TableDockable[T]) MayAttemptClose() bool {
	return workspace.MayAttemptCloseOfGroup(d)
}

// AttemptClose implements unison.TabCloser
func (d *TableDockable[T]) AttemptClose() bool {
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

func (d *TableDockable[T]) save(forceSaveAs bool) bool {
	success := false
	if forceSaveAs || d.needsSaveAsPrompt {
		success = workspace.SaveDockableAs(d, d.extension, d.saver, func(path string) {
			d.crc = d.crc64()
			d.path = path
		})
	} else {
		success = workspace.SaveDockable(d, d.saver, func() { d.crc = d.crc64() })
	}
	if success {
		d.needsSaveAsPrompt = false
	}
	return success
}

func (d *TableDockable[T]) toggleHierarchy() {
	first := true
	open := false
	for _, row := range d.table.RootRows() {
		if row.CanHaveChildren() {
			if first {
				first = false
				open = !row.IsOpen()
			}
			setRowOpen(row, open)
		}
	}
	d.table.SyncToModel()
}

func setRowOpen[T gurps.NodeTypes](row *ntable.Node[T], open bool) {
	row.SetOpen(open)
	for _, child := range row.Children() {
		if child.CanHaveChildren() {
			setRowOpen(child, open)
		}
	}
}

func (d *TableDockable[T]) sizeToFit() {
	d.table.SizeColumnsToFit(true)
	d.table.MarkForRedraw()
}

func (d *TableDockable[T]) searchModified() {
	d.searchIndex = 0
	d.searchResult = nil
	text := strings.ToLower(d.searchField.Text())
	for _, row := range d.table.RootRows() {
		d.search(text, row)
	}
	d.adjustForMatch()
}

func (d *TableDockable[T]) search(text string, row *ntable.Node[T]) {
	if row.Match(text) {
		d.searchResult = append(d.searchResult, row)
	}
	if row.CanHaveChildren() {
		for _, child := range row.Children() {
			d.search(text, child)
		}
	}
}

func (d *TableDockable[T]) previousMatch() {
	if d.searchIndex > 0 {
		d.searchIndex--
		d.adjustForMatch()
	}
}

func (d *TableDockable[T]) nextMatch() {
	if d.searchIndex < len(d.searchResult)-1 {
		d.searchIndex++
		d.adjustForMatch()
	}
}

func (d *TableDockable[T]) adjustForMatch() {
	d.backButton.SetEnabled(d.searchIndex != 0)
	d.forwardButton.SetEnabled(len(d.searchResult) != 0 && d.searchIndex != len(d.searchResult)-1)
	if len(d.searchResult) != 0 {
		d.matchesLabel.Text = fmt.Sprintf(i18n.Text("%d of %d"), d.searchIndex+1, len(d.searchResult))
		row := d.searchResult[d.searchIndex]
		d.table.DiscloseRow(row, false)
		d.table.ClearSelection()
		rowIndex := d.table.RowToIndex(row)
		d.table.SelectByIndex(rowIndex)
		d.table.ScrollRowIntoView(rowIndex)
	} else {
		d.matchesLabel.Text = "-"
	}
	d.matchesLabel.Parent().MarkForLayoutAndRedraw()
}

// Rebuild implements widget.Rebuildable.
func (d *TableDockable[T]) Rebuild(_ bool) {
	h, v := d.scroll.Position()
	sel := d.table.CopySelectionMap()
	d.table.SyncToModel()
	d.table.SetSelectionMap(sel)
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.UpdateTitle(d)
	}
	d.scroll.SetPosition(h, v)
}

func (d *TableDockable[T]) crc64() uint64 {
	var buffer bytes.Buffer
	rows := d.provider.RootRows()
	data := make([]any, 0, len(rows))
	for _, row := range rows {
		data = append(data, row.Data())
	}
	if err := jio.Save(context.Background(), &buffer, data); err != nil {
		return 0
	}
	return crc.Bytes(0, buffer.Bytes())
}

func (d *TableDockable[T]) canCopySelectionToSheet(_ any) bool {
	return d.table.HasSelection() && len(sheet.OpenSheets()) > 0
}

func (d *TableDockable[T]) copySelectionToSheet(_ any) {
	if d.table.HasSelection() {
		if sheets := workspace.PromptForDestination(sheet.OpenSheets()); len(sheets) > 0 {
			sel := d.table.SelectedRows(true)
			for _, s := range sheets {
				var table *unison.Table[*ntable.Node[T]]
				var postProcessor func(rows []*ntable.Node[T])
				switch any(sel[0].Data()).(type) {
				case *gurps.Trait:
					table = d.convertTable(s.Traits.Table)
				case *gurps.Skill:
					table = d.convertTable(s.Skills.Table)
				case *gurps.Spell:
					table = d.convertTable(s.Spells.Table)
				case *gurps.Equipment:
					table = d.convertTable(s.CarriedEquipment.Table)
					postProcessor = func(rows []*ntable.Node[T]) {
						if erows, ok := interface{}(rows).([]*ntable.Node[*gurps.Equipment]); ok {
							for _, row := range erows {
								gurps.Traverse(func(e *gurps.Equipment) bool {
									e.Equipped = true
									return false
								}, false, false, row.Data())
							}
						}
					}
				case *gurps.Note:
					table = d.convertTable(s.Notes.Table)
				default:
					continue
				}
				if table != nil {
					ntable.CopyRowsTo(table, sel, postProcessor)
					ntable.ProcessModifiersForSelection(table)
					ntable.ProcessNameablesForSelection(table)
				}
			}
		}
	}
}

func (d *TableDockable[T]) convertTable(table any) *unison.Table[*ntable.Node[T]] {
	// This is here just to get around limitations in the way Go generics behave
	if t, ok := table.(*unison.Table[*ntable.Node[T]]); ok {
		return t
	}
	return nil
}

func (d *TableDockable[T]) copySelectionToTemplate(_ any) {
	if d.table.HasSelection() {
		if templates := workspace.PromptForDestination(sheet.OpenTemplates()); len(templates) > 0 {
			sel := d.table.SelectedRows(true)
			for _, t := range templates {
				switch any(sel[0].Data()).(type) {
				case *gurps.Trait:
					ntable.CopyRowsTo(d.convertTable(t.Traits.Table), sel, nil)
				case *gurps.Skill:
					ntable.CopyRowsTo(d.convertTable(t.Skills.Table), sel, nil)
				case *gurps.Spell:
					ntable.CopyRowsTo(d.convertTable(t.Spells.Table), sel, nil)
				case *gurps.Equipment:
					ntable.CopyRowsTo(d.convertTable(t.Equipment.Table), sel, nil)
				case *gurps.Note:
					ntable.CopyRowsTo(d.convertTable(t.Notes.Table), sel, nil)
				}
			}
		}
	}
}

func (d *TableDockable[T]) canCopySelectionToTemplate(_ any) bool {
	return d.table.HasSelection() && len(sheet.OpenTemplates()) > 0
}
