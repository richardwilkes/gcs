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
	"bytes"
	"context"
	"hash"
	"strings"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/check"
)

var (
	_ FileBackedDockable         = &TableDockable[*gurps.Trait]{}
	_ unison.UndoManagerProvider = &TableDockable[*gurps.Trait]{}
	_ ModifiableRoot             = &TableDockable[*gurps.Trait]{}
	_ Rebuildable                = &TableDockable[*gurps.Trait]{}
	_ unison.TabCloser           = &TableDockable[*gurps.Trait]{}
	_ KeyedDockable              = &TableDockable[*gurps.Trait]{}
	_ TagProvider                = &TableDockable[*gurps.Trait]{}
	_ gurps.Hashable             = &TableDockable[*gurps.Trait]{}
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
	filterField       *unison.Field
	namesOnlyCheckBox *unison.CheckBox
	scroll            *unison.ScrollPanel
	tableHeader       *unison.TableHeader[*Node[T]]
	table             *unison.Table[*Node[T]]
	hash              uint64
	scale             int
	needsSaveAsPrompt bool
}

// NewTableDockable creates a new TableDockable for list data files.
func NewTableDockable[T gurps.NodeTypes](filePath, extension string, provider TableProvider[T], saver func(path string) error, canCreateIDs ...int) *TableDockable[T] {
	header, table := NewNodeTable(provider, nil)
	d := &TableDockable[T]{
		path:              filePath,
		extension:         extension,
		undoMgr:           unison.NewUndoManager(200, func(err error) { errs.Log(err) }),
		provider:          provider,
		saver:             saver,
		canCreateIDs:      make(map[int]bool),
		scroll:            unison.NewScrollPanel(),
		tableHeader:       header,
		table:             table,
		scale:             gurps.GlobalSettings().General.InitialListUIScale,
		needsSaveAsPrompt: true,
	}
	d.Self = d
	d.SetLayout(&unison.FlexLayout{Columns: 1})

	for _, id := range canCreateIDs {
		d.canCreateIDs[id] = true
	}

	d.table.SyncToModel()
	d.table.SizeColumnsToFit(true)
	if columnSizing, ok := gurps.GlobalSettings().ColumnSizing[filePath]; ok {
		needSync := false
		for id, width := range columnSizing {
			if id != -1 {
				if i := d.table.ColumnIndexForID(id); i != -1 {
					if d.table.Columns[i].Current != width {
						d.table.Columns[i].Current = width
						needSync = true
					}
				}
			}
		}
		if needSync {
			d.table.SyncToModel()
		}
	}

	InstallTableDropSupport(d.table, d.provider)

	d.scroll.SetColumnHeader(d.tableHeader)
	d.scroll.SetContent(d.table, behavior.Fill, behavior.Fill)
	d.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
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
		func(_ any) bool { return HasSelectionAndNotFiltered(d.table) },
		func(_ any) { DeleteSelection(d.table, true) })
	d.InstallCmdHandlers(DuplicateItemID,
		func(_ any) bool { return HasSelectionAndNotFiltered(d.table) },
		func(_ any) { DuplicateSelection(d.table) })
	table.InstallCmdHandlers(SyncWithSourceItemID,
		func(_ any) bool { return HasSelectionAndNotFiltered(d.table) },
		func(_ any) { SyncWithSourceForSelection(d.table) })
	table.InstallCmdHandlers(ClearSourceItemID,
		func(_ any) bool { return HasSelectionAndNotFiltered(d.table) },
		func(_ any) { ClearSourceFromSelection(d.table) })
	d.InstallCmdHandlers(JumpToSearchFilterItemID,
		func(any) bool { return !d.filterField.Focused() },
		func(any) { d.filterField.RequestFocus() })
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
	d.hash = gurps.Hash64(d)
	return d
}

// DockKey implements KeyedDockable
func (d *TableDockable[T]) DockKey() string {
	return filePrefix + d.path
}

func (d *TableDockable[T]) createToolbar() *unison.Panel {
	hierarchyButton := unison.NewSVGButton(svg.Hierarchy)
	hierarchyButton.Tooltip = newWrappedTooltip(i18n.Text("Opens/closes all hierarchical rows"))
	hierarchyButton.ClickCallback = d.toggleHierarchy

	noteToggleButton := unison.NewSVGButton(svg.NotesToggle)
	noteToggleButton.Tooltip = newWrappedTooltip(i18n.Text("Opens/closes all embedded notes"))
	noteToggleButton.ClickCallback = d.toggleNotes

	sizeToFitButton := unison.NewSVGButton(svg.SizeToFit)
	sizeToFitButton.Tooltip = newWrappedTooltip(i18n.Text("Sets the width of each column to fit its contents"))
	sizeToFitButton.ClickCallback = d.sizeToFit

	filterPopup := NewTagFilterPopup(d)

	d.filterField = NewSearchField(i18n.Text("Content Filter"), func(_, _ *unison.FieldState) {
		d.ApplyFilter(SelectedTags(filterPopup))
	})

	d.namesOnlyCheckBox = unison.NewCheckBox()
	d.namesOnlyCheckBox.SetTitle(i18n.Text("Names Only"))
	d.namesOnlyCheckBox.ClickCallback = func() { d.ApplyFilter(SelectedTags(filterPopup)) }

	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
	toolbar.AddChild(NewDefaultInfoPop())
	toolbar.AddChild(
		NewScaleField(
			gurps.InitialUIScaleMin,
			gurps.InitialUIScaleMax,
			func() int { return gurps.GlobalSettings().General.InitialListUIScale },
			func() int { return d.scale },
			func(scale int) { d.scale = scale },
			nil,
			false,
			d.scroll,
		),
	)
	toolbar.AddChild(hierarchyButton)
	toolbar.AddChild(noteToggleButton)
	toolbar.AddChild(sizeToFitButton)
	toolbar.AddChild(d.filterField)
	toolbar.AddChild(d.namesOnlyCheckBox)
	toolbar.AddChild(filterPopup)
	toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
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
		SVG:  gurps.FileInfoFor(d.path).SVG,
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
	UpdateTitleForDockable(d)
}

// Modified implements workspace.FileBackedDockable
func (d *TableDockable[T]) Modified() bool {
	return d.hash != gurps.Hash64(d)
}

// MarkModified implements widget.ModifiableRoot.
func (d *TableDockable[T]) MarkModified(_ unison.Paneler) {
	UpdateTitleForDockable(d)
}

// MayAttemptClose implements unison.TabCloser
func (d *TableDockable[T]) MayAttemptClose() bool {
	return MayAttemptCloseOfGroup(d)
}

// AttemptClose implements unison.TabCloser
func (d *TableDockable[T]) AttemptClose() bool {
	if AttemptSaveForDockable(d) {
		d.preserveColumns()
		return AttemptCloseForDockable(d)
	}
	return false
}

func (d *TableDockable[T]) save(forceSaveAs bool) bool {
	success := false
	if forceSaveAs || d.needsSaveAsPrompt {
		success = SaveDockableAs(d, d.extension, d.saver, func(path string) {
			d.hash = gurps.Hash64(d)
			d.path = path
		})
	} else {
		success = SaveDockable(d, d.saver, func() { d.hash = gurps.Hash64(d) })
	}
	if success {
		d.needsSaveAsPrompt = false
	}
	return success
}

func (d *TableDockable[T]) preserveColumns() {
	m := make(map[int]float32, len(d.table.Columns))
	m[-1] = gurps.ToColumnCutoff(time.Now().Unix())
	for _, col := range d.table.Columns {
		m[col.ID] = col.Current
	}
	gurps.GlobalSettings().ColumnSizing[d.BackingFilePath()] = m
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

func (d *TableDockable[T]) toggleNotes() {
	state := 0
	for _, row := range d.table.RootRows() {
		discoverNoteState(row, &state)
		if state != 0 {
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
	for _, row := range d.table.RootRows() {
		applyNoteState(row, closed)
	}
	d.table.SyncToModel()
}

func discoverNoteState[T gurps.NodeTypes](n *Node[T], state *int) {
	for i := range n.table.Columns {
		var data gurps.CellData
		n.dataAsNode.CellData(n.table.Columns[i].ID, &data)
		if data.Type == cell.Text && data.Secondary != "" {
			if gurps.IsClosed("N:" + string(n.ID())) {
				*state = -1
			} else {
				*state = 1
			}
			return
		}
	}
	if n.CanHaveChildren() {
		for _, child := range n.Children() {
			discoverNoteState(child, state)
			if *state != 0 {
				return
			}
		}
	}
}

func applyNoteState[T gurps.NodeTypes](n *Node[T], closed bool) {
	for i := range n.table.Columns {
		var data gurps.CellData
		n.dataAsNode.CellData(n.table.Columns[i].ID, &data)
		if data.Type == cell.Text && data.Secondary != "" {
			id := "N:" + string(n.ID())
			if gurps.IsClosed(id) != closed {
				gurps.SetClosedState(id, closed)
			}
		}
	}
	if n.CanHaveChildren() {
		for _, child := range n.Children() {
			applyNoteState(child, closed)
		}
	}
}

func (d *TableDockable[T]) sizeToFit() {
	d.table.SizeColumnsToFit(true)
	d.table.MarkForRedraw()
}

// Rebuild implements widget.Rebuildable.
func (d *TableDockable[T]) Rebuild(_ bool) {
	gurps.DiscardGlobalResolveCache()
	h, v := d.scroll.Position()
	sel := d.table.CopySelectionMap()
	d.table.SyncToModel()
	d.table.SetSelectionMap(sel)
	UpdateTitleForDockable(d)
	d.scroll.SetPosition(h, v)
}

// Hash writes this object's contents into the hasher.
func (d *TableDockable[T]) Hash(h hash.Hash) {
	var buffer bytes.Buffer
	rows := d.provider.RootRows()
	data := make([]any, 0, len(rows))
	for _, row := range rows {
		data = append(data, row.Data())
	}
	if err := jio.Save(context.Background(), &buffer, data); err != nil {
		errs.Log(err)
		return
	}
	_, _ = h.Write(buffer.Bytes())
}

// AllTags returns all tags currently present in the data.
func (d *TableDockable[T]) AllTags() []string {
	return d.provider.AllTags()
}

// ApplyFilter applies the current filtering, if any.
func (d *TableDockable[T]) ApplyFilter(tags []string) {
	if d.filterField != nil {
		text := strings.ToLower(strings.TrimSpace(d.filterField.GetFieldState().Text))
		var f func(row *Node[T]) bool
		if len(tags) != 0 || text != "" {
			f = func(row *Node[T]) bool {
				match := false
				if d.namesOnlyCheckBox.State == check.On {
					match = strings.Contains(strings.ToLower(row.dataAsNode.String()), text)
				} else {
					match = row.PartialMatchExceptTag(text)
				}
				if match {
					for _, tag := range tags {
						if !row.HasTag(tag) {
							return true
						}
					}
					return false
				}
				return true
			}
		}
		d.table.ApplyFilter(f)
	}
}
