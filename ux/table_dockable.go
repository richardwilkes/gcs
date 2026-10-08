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
	// Installed here rather than in NewNodeTable, since only library lists can be a source: modifiers applied from the
	// list inside a trait or equipment editor to the item being edited would be lost when the editor applies its
	// changes, which replace the item's modifiers wholesale.
	switch t := any(d.table).(type) {
	case *unison.Table[*Node[*gurps.TraitModifier]]:
		installApplyModifierHandler(d, t, traitModifierTargetKind())
		installModifierChoiceConversionHandlers[*gurps.TraitModifier, *gurps.TraitModifierEditData](d, t)
	case *unison.Table[*Node[*gurps.EquipmentModifier]]:
		installApplyModifierHandler(d, t, equipmentModifierTargetKind())
		installModifierChoiceConversionHandlers[*gurps.EquipmentModifier, *gurps.EquipmentModifierEditData](d, t)
	}
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
		case id > FirstGroupContainerMarker && id < LastGroupContainerMarker:
			variant = GroupContainerItemVariant
		case id > FirstModifierChoiceMarker && id < LastModifierChoiceMarker:
			variant = ChoiceContainerItemVariant
		case id > FirstAlternateNonContainerMarker && id < LastAlternateNonContainerMarker:
			variant = AlternateItemVariant
		}
		if variant != -1 {
			// Creating an item inserts a row, which unison.Table.ApplyHierarchicalFilter forbids while a filter is
			// applied, so the command is disabled then, as Delete, Duplicate and the move commands are. Left on, the
			// insert would clear the filter, open an editor on the new row and then have the rebuild that follows
			// re-filter the row out from under that editor.
			d.InstallCmdHandlers(id,
				func(_ any) bool { return !d.table.IsFiltered() },
				func(_ any) { d.provider.CreateItem(d, d.table, variant) })
		}
	}
	return d
}

func (d *TableDockable[T]) createToolbar() *unison.Panel {
	// The hierarchy and note buttons are kept so they can be disabled while a filter is applied (see toggleHierarchy
	// and toggleNotes).
	d.hierarchyButton = unison.NewSVGButton(svg.Hierarchy)
	d.hierarchyButton.Tooltip = newWrappedTooltip(i18n.Text("Opens/closes all hierarchical rows"))
	d.hierarchyButton.ClickCallback = d.toggleHierarchy

	d.noteToggleButton = unison.NewSVGButton(svg.NotesToggle)
	d.noteToggleButton.Tooltip = newWrappedTooltip(i18n.Text("Opens/closes all embedded notes"))
	d.noteToggleButton.ClickCallback = d.toggleNotes

	sizeToFitButton := unison.NewSVGButton(svg.SizeToFit)
	sizeToFitButton.Tooltip = newWrappedTooltip(i18n.Text("Sets the width of each column to fit its contents"))
	sizeToFitButton.ClickCallback = d.sizeToFit

	// The quick filter is always in play: what is typed here narrows the list on top of a saved filter when one is in
	// force, and is all that filters the list otherwise.
	d.filterField = NewSearchField(i18n.Text("Quick Filter"), func(_, _ *unison.FieldState) { d.applyFilter() })

	toolbar := NewToolbar()
	toolbar.AddChild(NewDefaultInfoPop())
	AddUIScaleField(toolbar, func() int { return gurps.GlobalSettings().General.InitialListUIScale },
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
	FinishToolbarLayout(toolbar)
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

// toggleHierarchy opens or closes every container in the table. It does nothing while a filter is applied, when its
// button is disabled: the filtered view shows every container it keeps as open whatever its own open state, so the
// toggle would have nothing to show for it until the filter was cleared.
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
	// The filter stays in force across a rebuild, and applying it syncs the table itself, so sync here only when it
	// didn't. Syncing first regardless would measure every row of the stale filtered set only to throw that away.
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

// chooseFilter puts the given saved filter in force, or drops the current one when f is nil. The quick filter's text is
// left alone, since it narrows the list on top of the saved filter.
func (d *TableDockable[T]) chooseFilter(f *gurps.ListFilter) {
	d.selectedFilter = f
	d.applyFilter()
}

// listFiltersChanged implements listFilterObserver.
func (d *TableDockable[T]) listFiltersChanged(key string, source *listFilterPopup) {
	if d.savedFilters != nil && d.savedFilters != source && d.savedFilters.spec.key == key {
		d.savedFilters.refresh()
	}
}

// applyFilter applies the saved filter in force and the quick filter's text together, showing a row only when it passes
// both, and reports whether the table was synced to its model, which unison.Table.ApplyHierarchicalFilter does whenever
// it is given a filter or has one to clear. The hierarchy is kept so a matching row is seen beneath the containers that
// hold it; a container shown only for that reason is dimmed (see Node.cellData).
func (d *TableDockable[T]) applyFilter() (synced bool) {
	if d.filterField == nil {
		return false
	}
	// The table's filter is told which rows to drop, so each of these rejects the rows that fail its filter.
	var saved, quick func(row *Node[T]) bool
	if d.selectedFilter != nil {
		// The fields are looked up once here rather than once per row.
		m := gurps.NewListFilterMatcher(d.selectedFilter, d.provider.FilterFields())
		saved = func(row *Node[T]) bool { return !m(row.Data()) }
	}
	if text := strings.ToLower(strings.TrimSpace(d.filterField.GetFieldState().Text)); text != "" {
		// Match looks at every column, the tags column included.
		quick = func(row *Node[T]) bool { return !row.Match(text) }
	}
	var f func(row *Node[T]) bool
	switch {
	case saved != nil && quick != nil:
		f = func(row *Node[T]) bool { return saved(row) || quick(row) }
	case saved != nil:
		f = saved
	case quick != nil:
		f = quick
	case !d.table.IsFiltered():
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
