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

package ux

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/crc"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	gsettings "github.com/richardwilkes/gcs/v5/model/gurps/settings"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
)

var (
	_ FileBackedDockable         = &TableDockable[*gurps.Trait]{}
	_ unison.UndoManagerProvider = &TableDockable[*gurps.Trait]{}
	_ ModifiableRoot             = &TableDockable[*gurps.Trait]{}
	_ Rebuildable                = &TableDockable[*gurps.Trait]{}
	_ unison.TabCloser           = &TableDockable[*gurps.Trait]{}
)

// TableDockable holds the view for a file that contains a (potentially hierarchical) list of data.
type TableDockable[T gurps.NodeTypes] struct {
	unison.Panel
	path              string
	extension         string
	undoMgr           *unison.UndoManager
	provider          TableProvider[T]
	saver             func(path string) error
	canCreateIDs      map[int]bool
	hierarchyButton   *unison.Button
	sizeToFitButton   *unison.Button
	filterPopup       *unison.PopupMenu[string]
	filterField       *unison.Field
	scroll            *unison.ScrollPanel
	tableHeader       *unison.TableHeader[*Node[T]]
	table             *unison.Table[*Node[T]]
	crc               uint64
	scale             int
	needsSaveAsPrompt bool
}

// NewTableDockable creates a new TableDockable for list data files.
func NewTableDockable[T gurps.NodeTypes](filePath, extension string, provider TableProvider[T], saver func(path string) error, canCreateIDs ...int) *TableDockable[T] {
	header, table := NewNodeTable[T](provider, nil)
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
	InstallTableDropSupport(d.table, d.provider)

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

	d.InstallCmdHandlers(OpenEditorItemID,
		func(_ any) bool { return d.table.HasSelection() },
		func(_ any) { d.provider.OpenEditor(d, d.table) })
	d.InstallCmdHandlers(OpenOnePageReferenceItemID,
		func(_ any) bool { return CanOpenPageRef(d.table) },
		func(_ any) { OpenPageRef(d.table) })
	d.InstallCmdHandlers(OpenEachPageReferenceItemID,
		func(_ any) bool { return CanOpenPageRef(d.table) },
		func(_ any) { OpenEachPageRef(d.table) })
	d.InstallCmdHandlers(SaveItemID,
		func(_ any) bool { return d.Modified() },
		func(_ any) { d.save(false) })
	d.InstallCmdHandlers(SaveAsItemID, unison.AlwaysEnabled, func(_ any) { d.save(true) })
	d.InstallCmdHandlers(unison.DeleteItemID,
		func(_ any) bool { return !d.table.IsFiltered() && d.table.HasSelection() },
		func(_ any) { DeleteSelection(d.table) })
	d.InstallCmdHandlers(DuplicateItemID,
		func(_ any) bool { return !d.table.IsFiltered() && d.table.HasSelection() },
		func(_ any) { DuplicateSelection(d.table) })
	for _, id := range canCreateIDs {
		variant := ItemVariant(-1)
		switch {
		case id > FirstNonContainerMarker && id < LastNonContainerMarker:
			variant = NoItemVariant
		case id > FirstContainerMarker && id < LastContainerMarker:
			variant = ContainerItemVariant
		case id > FirstAlternateNonContainerMarker && id < LastAlternateNonContainerMarker:
			variant = AlternateItemVariant
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
	d.hierarchyButton = unison.NewSVGButton(svg.Hierarchy)
	d.hierarchyButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Opens/closes all hierarchical rows"))
	d.hierarchyButton.ClickCallback = d.toggleHierarchy

	d.sizeToFitButton = unison.NewSVGButton(svg.SizeToFit)
	d.sizeToFitButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Sets the width of each column to fit its contents"))
	d.sizeToFitButton.ClickCallback = d.sizeToFit

	d.filterField = unison.NewField()
	filter := i18n.Text("Content Filter")
	d.filterField.Watermark = filter
	d.filterField.Tooltip = unison.NewTooltipWithText(filter)
	d.filterField.ModifiedCallback = d.applyFilter
	d.filterField.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
		HGrab:  true,
	})

	d.filterPopup = unison.NewPopupMenu[string]()
	d.filterPopup.Tooltip = unison.NewTooltipWithText(i18n.Text("Tag Filter"))
	d.filterPopup.AddItem(i18n.Text("Any Tag"))
	for _, tag := range d.provider.AllTags() {
		if d.filterPopup.ItemCount() == 1 {
			d.filterPopup.AddSeparator()
		}
		d.filterPopup.AddItem(tag)
	}
	d.filterPopup.SelectionCallback = func(index int, tag string) {
		d.applyFilter()
	}
	d.filterPopup.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})

	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
	toolbar.AddChild(NewDefaultInfoPop())
	toolbar.AddChild(NewScaleField(gsettings.InitialUIScaleMin, gsettings.InitialUIScaleMax,
		func() int { return settings.Global().General.InitialListUIScale }, func() int { return d.scale },
		func(scale int) { d.scale = scale }, d.scroll, nil, false))
	toolbar.AddChild(d.hierarchyButton)
	toolbar.AddChild(d.sizeToFitButton)
	toolbar.AddChild(d.filterField)
	toolbar.AddChild(d.filterPopup)
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
	return ListDockableKind
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
	return MayAttemptCloseOfGroup(d)
}

// AttemptClose implements unison.TabCloser
func (d *TableDockable[T]) AttemptClose() bool {
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
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.Close(d)
	}
	return true
}

func (d *TableDockable[T]) save(forceSaveAs bool) bool {
	success := false
	if forceSaveAs || d.needsSaveAsPrompt {
		success = SaveDockableAs(d, d.extension, d.saver, func(path string) {
			d.crc = d.crc64()
			d.path = path
		})
	} else {
		success = SaveDockable(d, d.saver, func() { d.crc = d.crc64() })
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
			setTableDockableRowOpen(row, open)
		}
	}
	d.table.SyncToModel()
}

func setTableDockableRowOpen[T gurps.NodeTypes](row *Node[T], open bool) {
	row.SetOpen(open)
	for _, child := range row.Children() {
		if child.CanHaveChildren() {
			setTableDockableRowOpen(child, open)
		}
	}
}

func (d *TableDockable[T]) sizeToFit() {
	d.table.SizeColumnsToFit(true)
	d.table.MarkForRedraw()
}

func (d *TableDockable[T]) applyFilter() {
	tag := ""
	if index := d.filterPopup.SelectedIndex(); index > 0 {
		if item, ok := d.filterPopup.ItemAt(index); ok {
			tag = item
		}
	}
	text := strings.TrimSpace(d.filterField.Text())
	if tag == "" && text == "" {
		d.table.ApplyFilter(nil)
	} else {
		d.table.ApplyFilter(func(row *Node[T]) bool {
			return !(row.HasTag(tag) && row.PartialMatchExceptTag(text))
		})
	}
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
