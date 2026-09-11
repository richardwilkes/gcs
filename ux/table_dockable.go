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
)

var (
	_ FileBackedDockable         = &TableDockable[*gurps.Trait]{}
	_ unison.UndoManagerProvider = &TableDockable[*gurps.Trait]{}
	_ ModifiableRoot             = &TableDockable[*gurps.Trait]{}
	_ Rebuildable                = &TableDockable[*gurps.Trait]{}
	_ unison.TabCloser           = &TableDockable[*gurps.Trait]{}
	_ KeyedDockable              = &TableDockable[*gurps.Trait]{}
	_ gurps.Hashable             = &TableDockable[*gurps.Trait]{}
	_ listFilterObserver         = &TableDockable[*gurps.Trait]{}
)

// TableDockable holds the view for a file that contains a (potentially hierarchical) list of data.
type TableDockable[T gurps.Node[T]] struct {
	fileBackedPanel
	undoMgr          *unison.UndoManager
	provider         TableProvider[T]
	hierarchyButton  *unison.Button
	noteToggleButton *unison.Button
	filterField      *unison.Field
	filterBanner     *filterBanner
	savedFilters     *listFilterPopup
	selectedFilter   *gurps.ListFilter
	scroll           *unison.ScrollPanel
	tableHeader      *unison.TableHeader[*Node[T]]
	table            *unison.Table[*Node[T]]
	scale            int
	choosingFilter   bool
}

// NewTableDockable creates a new TableDockable for list data files.
func NewTableDockable[T gurps.Node[T]](filePath, extension string, provider TableProvider[T], saver func(path string) error, canCreateIDs ...int) *TableDockable[T] {
	header, table := NewNodeTable(provider, nil)
	d := &TableDockable[T]{
		undoMgr:      unison.NewUndoManager(200, func(err error) { errs.Log(err) }),
		provider:     provider,
		filterBanner: newFilterBanner(),
		scroll:       unison.NewScrollPanel(),
		tableHeader:  header,
		table:        table,
		scale:        gurps.GlobalSettings().General.InitialListUIScale,
	}
	d.Self = d
	d.initFileEditor(d, filePath, extension, saver, d)
	d.SetLayout(&unison.FlexLayout{Columns: 1})

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
		func(any) bool { return d.filterField.Enabled() && !d.filterField.Focused() },
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
			// Creating an item inserts a row, which unison.Table.ApplyHierarchicalFilter says must not be done while a
			// filter is applied, so the command is turned off whenever a filter is hiding part of the list, as Delete,
			// Duplicate and the move commands are. Left on, the insert would clear the filter, open an editor on the
			// new row and then have the rebuild that follows re-filter the row out from under that editor.
			d.InstallCmdHandlers(id,
				func(_ any) bool { return !d.table.IsFiltered() },
				func(_ any) { d.provider.CreateItem(d, d.table, variant) })
		}
	}
	return d
}

func (d *TableDockable[T]) createToolbar() *unison.Panel {
	// The hierarchy and note buttons are kept so that they can be turned off while a filter is applied, when the table
	// shows a flat list of the rows that passed and a toggle over them would silently change just those rows.
	d.hierarchyButton = unison.NewSVGButton(svg.Hierarchy)
	d.hierarchyButton.Tooltip = newWrappedTooltip(i18n.Text("Opens/closes all hierarchical rows"))
	d.hierarchyButton.ClickCallback = d.toggleHierarchy

	d.noteToggleButton = unison.NewSVGButton(svg.NotesToggle)
	d.noteToggleButton.Tooltip = newWrappedTooltip(i18n.Text("Opens/closes all embedded notes"))
	d.noteToggleButton.ClickCallback = d.toggleNotes

	sizeToFitButton := unison.NewSVGButton(svg.SizeToFit)
	sizeToFitButton.Tooltip = newWrappedTooltip(i18n.Text("Sets the width of each column to fit its contents"))
	sizeToFitButton.ClickCallback = d.sizeToFit

	// The field is named the way the saved filter popup's entry for it is, since that entry is what hands it the list.
	d.filterField = NewSearchField(i18n.Text("Quick Filter"), func(_, _ *unison.FieldState) {
		if !d.choosingFilter {
			d.applyFilter()
		}
	})

	toolbar := newToolbar()
	toolbar.AddChild(NewDefaultInfoPop())
	addUIScaleField(toolbar, func() int { return gurps.GlobalSettings().General.InitialListUIScale },
		func() int { return d.scale }, func(scale int) { d.scale = scale }, false, d.scroll)
	toolbar.AddChild(d.hierarchyButton)
	toolbar.AddChild(d.noteToggleButton)
	toolbar.AddChild(sizeToFitButton)
	toolbar.AddChild(d.filterField)
	// The weapon and conditional modifier providers have no filter key, since their lists are never shown in a list
	// dockable, so they get no saved filter popup.
	if key := d.provider.FilterKey(); key != "" {
		d.savedFilters = newListFilterPopup(listFilterPopupSpec{
			key:     key,
			fields:  filterFieldInfos(d.provider.FilterFields()),
			current: func() *gurps.ListFilter { return d.selectedFilter },
			choose:  d.chooseFilter,
		})
		toolbar.AddChild(d.savedFilters.popup)
	}
	finishToolbarLayout(toolbar)
	return toolbar
}

// Entity implements EntityPanel. A list file has no entity, so nil is always returned.
func (d *TableDockable[T]) Entity() *gurps.Entity {
	return nil
}

// UndoManager implements unison.UndoManagerProvider.
func (d *TableDockable[T]) UndoManager() *unison.UndoManager {
	return d.undoMgr
}

// MarkModified implements ModifiableRoot.
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

// toggleHierarchy opens or closes every container in the table. Like the other row-affecting commands, it does nothing
// while a filter is applied, and its button is turned off along with the filter being applied. The filtered view shows
// every container it keeps as open whatever the container's own open state, so the toggle would change the open states
// without anything to show for it until the filter was cleared.
func (d *TableDockable[T]) toggleHierarchy() {
	if d.table.IsFiltered() {
		return
	}
	toggleHierarchy(d)
	d.table.SyncToModel()
}

// toggleNotes shows or hides every note in the table. It is gated on the filter the same way toggleHierarchy is, since
// it would reach the notes of the rows the filter is hiding as well as those in view.
func (d *TableDockable[T]) toggleNotes() {
	if d.table.IsFiltered() {
		return
	}
	if toggleNotes(d) {
		d.table.SyncToModel()
	}
}

func (d *TableDockable[T]) sizeToFit() {
	d.table.SizeColumnsToFit(true)
	d.table.MarkForRedraw()
}

// Rebuild implements Rebuildable.
func (d *TableDockable[T]) Rebuild(_ bool) {
	gurps.DiscardGlobalResolveCache()
	h, v := d.scroll.Position()
	sel := d.table.CopySelectionMap()
	// The rows have to be built afresh from the model and put through the filter, which is in force across a rebuild.
	// Applying the filter syncs the table itself, over the rows that pass, so the sync is only done here when there is
	// no filter to apply and none to clear, which is the one case where applying it leaves the table alone. Syncing
	// first regardless would measure every row over the stale set of filtered rows only to throw that away.
	if !d.applyFilter() {
		d.table.SyncToModel()
	}
	d.table.SetSelectionMap(sel)
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

// chooseFilter puts the given saved filter in force, or hands the list back to the quick filter when it is nil. The
// quick filter's field is only usable while no saved filter is in force, since the two would otherwise disagree about
// which rows to show.
func (d *TableDockable[T]) chooseFilter(f *gurps.ListFilter) {
	d.selectedFilter = f
	if f != nil {
		// Emptying the field fires its ModifiedCallback whenever it held text, which would put the rows through the
		// saved filter once here and once more below, so the callback is told to leave the filtering to this method.
		d.choosingFilter = true
		d.filterField.SetText("")
		d.choosingFilter = false
	}
	adjustFieldBlank(d.filterField, f != nil)
	d.applyFilter()
}

// listFiltersChanged implements listFilterObserver.
func (d *TableDockable[T]) listFiltersChanged(key string, source *listFilterPopup) {
	if d.savedFilters != nil && d.savedFilters != source && d.savedFilters.spec.key == key {
		d.savedFilters.refresh()
	}
}

// applyFilter applies the current filtering and reports whether the table was synced to its model as part of that,
// which unison.Table.ApplyHierarchicalFilter does whenever it is given a filter or has one to clear. The hierarchy is
// kept so that a matching row is seen in context, beneath the containers that hold it. A container shown only for that
// reason is dimmed (see Node.cellData), so the rows that actually matched stand out from those that are just context.
func (d *TableDockable[T]) applyFilter() (synced bool) {
	if d.filterField == nil {
		return false
	}
	var f func(row *Node[T]) bool
	if d.selectedFilter != nil {
		// The fields are looked up once here rather than once per row.
		m := gurps.NewListFilterMatcher(d.selectedFilter, d.provider.FilterFields())
		f = func(row *Node[T]) bool { return !m(row.Data()) }
	} else if text := strings.ToLower(strings.TrimSpace(d.filterField.GetFieldState().Text)); text != "" {
		// Match looks at every column, the tags column included, now that the tag popup that once did the tag
		// filtering is gone.
		f = func(row *Node[T]) bool { return !row.Match(text) }
	}
	if f == nil && !d.table.IsFiltered() {
		return false
	}
	d.table.ApplyHierarchicalFilter(f)
	filtered := d.table.IsFiltered()
	d.hierarchyButton.SetEnabled(!filtered)
	d.noteToggleButton.SetEnabled(!filtered)
	d.showFilterBanner(filtered)
	return true
}

// showFilterBanner puts the banner between the toolbar and the table, or takes it away again, and has the dockable
// laid out afresh when that changes anything. The banner is added and removed rather than hidden, since a hidden panel
// still takes up its space in a FlexLayout.
func (d *TableDockable[T]) showFilterBanner(show bool) {
	if show == (d.filterBanner.Parent() != nil) {
		return
	}
	if show {
		d.AddChildAtIndex(d.filterBanner, d.IndexOfChild(d.scroll))
	} else {
		d.filterBanner.RemoveFromParent()
	}
	d.MarkForLayoutAndRedraw()
}
