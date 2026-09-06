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
	"hash"
	"strings"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
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
type TableDockable[T gurps.Node[T]] struct {
	fileBackedPanel
	undoMgr           *unison.UndoManager
	provider          TableProvider[T]
	canCreateIDs      map[int]bool
	filterField       *unison.Field
	namesOnlyCheckBox *unison.CheckBox
	scroll            *unison.ScrollPanel
	tableHeader       *unison.TableHeader[*Node[T]]
	table             *unison.Table[*Node[T]]
	scale             int
}

// NewTableDockable creates a new TableDockable for list data files.
func NewTableDockable[T gurps.Node[T]](filePath, extension string, provider TableProvider[T], saver func(path string) error, canCreateIDs ...int) *TableDockable[T] {
	header, table := NewNodeTable(provider, nil)
	d := &TableDockable[T]{
		undoMgr:      unison.NewUndoManager(200, func(err error) { errs.Log(err) }),
		provider:     provider,
		canCreateIDs: make(map[int]bool),
		scroll:       unison.NewScrollPanel(),
		tableHeader:  header,
		table:        table,
		scale:        gurps.GlobalSettings().General.InitialListUIScale,
	}
	d.Self = d
	d.initFileEditor(d, filePath, extension, saver, d)
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

	installStandardTableCmdHandlers(d, d.table, d.provider, func() Rebuildable { return d }, true)
	d.InstallCmdHandlers(SaveItemID,
		func(_ any) bool { return d.Modified() },
		func(_ any) { d.save(false) })
	d.InstallCmdHandlers(SaveAsItemID, unison.AlwaysEnabled, func(_ any) { d.save(true) })
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
	return d
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

	toolbar := newToolbar()
	toolbar.AddChild(NewDefaultInfoPop())
	addUIScaleField(toolbar, func() int { return gurps.GlobalSettings().General.InitialListUIScale },
		func() int { return d.scale }, func(scale int) { d.scale = scale }, false, d.scroll)
	toolbar.AddChild(hierarchyButton)
	toolbar.AddChild(noteToggleButton)
	toolbar.AddChild(sizeToFitButton)
	toolbar.AddChild(d.filterField)
	toolbar.AddChild(d.namesOnlyCheckBox)
	toolbar.AddChild(filterPopup)
	finishToolbarLayout(toolbar)
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

// MarkModified implements widget.ModifiableRoot.
func (d *TableDockable[T]) MarkModified(_ unison.Paneler) {
	UpdateTitleForDockable(d)
}

// AttemptClose implements unison.TabCloser. The column widths are preserved as the table goes away, so that the file
// can be opened again with the same ones.
func (d *TableDockable[T]) AttemptClose() bool {
	if AttemptSaveForDockable(d) {
		d.preserveColumns()
		return AttemptCloseForDockable(d)
	}
	return false
}

func (d *TableDockable[T]) preserveColumns() {
	m := make(map[int]float32, len(d.table.Columns))
	m[-1] = gurps.ToColumnCutoff(time.Now().Unix())
	for _, col := range d.table.Columns {
		m[col.ID] = col.Current
	}
	gurps.GlobalSettings().ColumnSizing[d.BackingFilePath()] = m
}

// FirstDisclosureState implements hierarchyDiscloser.
func (d *TableDockable[T]) FirstDisclosureState() (open, exists bool) {
	return firstTableDisclosureState(d.table)
}

// SetDisclosureState implements hierarchyDiscloser.
func (d *TableDockable[T]) SetDisclosureState(open bool) {
	setTableDisclosureState(d.table, open)
}

// FirstNoteState implements noteDiscloser.
func (d *TableDockable[T]) FirstNoteState() int {
	return firstTableNoteState(d.table)
}

// ApplyNoteState implements noteDiscloser.
func (d *TableDockable[T]) ApplyNoteState(closed bool) {
	applyTableNoteState(d.table, closed)
}

func (d *TableDockable[T]) toggleHierarchy() {
	toggleHierarchy(d)
	d.table.SyncToModel()
}

func (d *TableDockable[T]) toggleNotes() {
	if toggleNotes(d) {
		d.table.SyncToModel()
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
	syncTablePreservingSelection(d.table)
	UpdateTitleForDockable(d)
	d.scroll.SetPosition(h, v)
}

// Hash writes this object's contents into the hasher.
func (d *TableDockable[T]) Hash(h hash.Hash) {
	rows := d.provider.RootRows()
	data := make([]any, 0, len(rows))
	for _, row := range rows {
		data = append(data, row.Data())
	}
	gurps.HashJSON(h, data)
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
					match = strings.Contains(strings.ToLower(row.data.String()), text)
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
