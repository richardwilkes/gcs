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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// listFilterControls is the part of a list dockable that filters its table through the quick filter, the tag filter and
// the saved filters, and opens and closes the table's containers and notes, all from the dockable's toolbar. A dockable
// embeds it by value and calls initListFilterControls before anything else.
type listFilterControls[T gurps.Node[T]] struct {
	table *unison.Table[*Node[T]]
	// provider supplies the filter key and fields and the rows the saved filters and the tag filter work from.
	provider TableProvider[T]
	// The hierarchy and note buttons are kept so they can be disabled while a filter is applied (see toggleHierarchy
	// and toggleNotes).
	hierarchyButton  *unison.Button
	noteToggleButton *unison.Button
	filterField      *unison.Field
	savedFilters     *listFilterPopup
	selectedFilter   *gurps.ListFilter
	tagFilter        *listTagFilterPopup
	// quickMatch tests a row against the quick filter's lowercased text; nil means (*Node[T]).Match.
	quickMatch func(row *Node[T], text string) bool
	// filterApplied, when set, is called after applyFilter changed what the table's filter is, with whether one is now
	// in force and how many rows passed it.
	filterApplied func(filtered bool, matched int)
}

// initListFilterControls sets the controls up for the table and its provider. quickMatch and filterApplied may be nil.
func (f *listFilterControls[T]) initListFilterControls(table *unison.Table[*Node[T]], provider TableProvider[T],
	quickMatch func(row *Node[T], text string) bool, filterApplied func(filtered bool, matched int),
) {
	f.table = table
	f.provider = provider
	f.quickMatch = quickMatch
	f.filterApplied = filterApplied
}

// addListFilterControls creates the hierarchy and note buttons, the quick filter and, when the provider has a filter
// key, the tag filter and saved filter popups, and adds them to the toolbar in that order, with the given panels placed
// just ahead of the quick filter.
func (f *listFilterControls[T]) addListFilterControls(toolbar *unison.Panel, beforeQuickFilter ...unison.Paneler) {
	f.hierarchyButton = unison.NewSVGButton(svg.Hierarchy)
	f.hierarchyButton.Tooltip = newWrappedTooltip(i18n.Text("Opens/closes all hierarchical rows"))
	f.hierarchyButton.ClickCallback = f.toggleHierarchy

	f.noteToggleButton = unison.NewSVGButton(svg.NotesToggle)
	f.noteToggleButton.Tooltip = newWrappedTooltip(i18n.Text("Opens/closes all embedded notes"))
	f.noteToggleButton.ClickCallback = f.toggleNotes

	// The quick filter is always in play: what is typed here narrows what the saved filter and the chosen tags let
	// through when either is in force, and is all that filters the list otherwise.
	f.filterField = NewToolbarSearchField(i18n.Text("Quick Filter"), func(_, _ *unison.FieldState) { f.applyFilter() })

	toolbar.AddChild(f.hierarchyButton)
	toolbar.AddChild(f.noteToggleButton)
	for _, one := range beforeQuickFilter {
		toolbar.AddChild(one)
	}
	toolbar.AddChild(f.filterField)
	// The weapon and conditional modifier providers have no filter key, since their lists are never shown in a list
	// dockable, so they get no saved filter popup.
	if key := f.provider.FilterKey(); key != "" {
		f.savedFilters = newListFilterPopup(listFilterPopupSpec{
			key:     key,
			fields:  filterFieldInfos(f.provider.FilterFields(), func() []T { return f.provider.RootData() }),
			current: func() *gurps.ListFilter { return f.selectedFilter },
			choose:  f.chooseFilter,
		})
		f.tagFilter = newListTagFilterPopup(listTagFilterPopupSpec{
			tags: func() []string {
				return gurps.ListFilterTags(f.selectedFilter, f.provider.FilterFields(), f.provider.RootData())
			},
			changed: func() { f.applyFilter() },
			save:    f.saveTagFilter,
		})
		toolbar.AddChild(f.tagFilter.popup)
		toolbar.AddChild(f.savedFilters.popup)
	}
}

// finishListFilterToolbar finishes the toolbar's layout. While there are filter popups, among the controls or the
// extra ones the caller added to the toolbar, they turn compact when the toolbar is too narrow for them, and
// extraChanged, when set, is called after they do.
func (f *listFilterControls[T]) finishListFilterToolbar(toolbar *unison.Panel, extra []*filterPopupButton,
	extraChanged func(),
) {
	layout := FinishToolbarLayout(toolbar)
	var buttons []*filterPopupButton
	if f.savedFilters != nil {
		buttons = append(buttons, f.tagFilter.button, f.savedFilters.button)
	}
	buttons = append(buttons, extra...)
	if len(buttons) == 0 {
		return
	}
	toolbar.SetLayout(&compactingToolbarLayout{
		FlexLayout: layout,
		buttons:    buttons,
		changed: func() {
			if f.savedFilters != nil {
				f.savedFilters.syncTooltip()
			}
			if extraChanged != nil {
				extraChanged()
			}
		},
	})
}

// FirstDisclosureState implements hierarchyDiscloser.
func (f *listFilterControls[T]) FirstDisclosureState() (open, exists bool) {
	return firstTableDisclosureState(f.table)
}

// SetDisclosureState implements hierarchyDiscloser.
func (f *listFilterControls[T]) SetDisclosureState(open bool) {
	setTableDisclosureState(f.table, open)
}

// FirstNoteState implements noteDiscloser.
func (f *listFilterControls[T]) FirstNoteState() int {
	return firstTableNoteState(f.table)
}

// ApplyNoteState implements noteDiscloser.
func (f *listFilterControls[T]) ApplyNoteState(closed bool) {
	applyTableNoteState(f.table, closed)
}

// toggleHierarchy opens or closes every container in the table. It does nothing while a filter is applied, when its
// button is disabled: the filtered view shows every container it keeps as open whatever its own open state, so the
// toggle would have nothing to show for it until the filter was cleared.
func (f *listFilterControls[T]) toggleHierarchy() {
	if f.table.IsFiltered() {
		return
	}
	toggleHierarchy(f)
}

// toggleNotes shows or hides every note in the table. It is gated on the filter the same way toggleHierarchy is, since
// it would reach the notes of the rows the filter is hiding as well as those in view.
func (f *listFilterControls[T]) toggleNotes() {
	if f.table.IsFiltered() {
		return
	}
	if toggleNotes(f) {
		f.table.SyncToModel()
	}
}

// syncSharedOpenStates implements openStateSharer.
func (f *listFilterControls[T]) syncSharedOpenStates(origin *unison.Panel, ids map[tid.TID]bool) {
	if f.table.AsPanel() == origin {
		return
	}
	shown := false
	gurps.Traverse(func(row T) bool {
		shown = ids[row.ID()]
		return shown
	}, false, false, f.provider.RootData()...)
	if shown {
		f.table.SyncToModel()
	}
}

// rebuildKeepingView builds the table's rows anew, keeping the filter in force, the selection and the scroll position.
func (f *listFilterControls[T]) rebuildKeepingView(scroll *unison.ScrollPanel) {
	invalidateDisplayData()
	gurps.DiscardGlobalResolveCache()
	h, v := scroll.Position()
	sel := f.table.CopySelectionMap()
	f.dropUnofferedTags()
	// The filter stays in force across a rebuild, and applying it syncs the table itself, so sync here only when it
	// didn't. Syncing first regardless would measure every row of the stale filtered set only to throw that away. The
	// table is told to measure the rows the filter shows, which it otherwise keeps the heights of from one filter to
	// the next, since what they show may have changed.
	f.table.DiscardRowHeights()
	if !f.applyFilter() {
		f.table.SyncToModel()
	}
	f.table.SetSelectionMap(sel)
	scroll.SetPosition(h, v)
}

// chooseFilter puts the given saved filter in force, or drops the current one when lf is nil. The quick filter's text
// is left alone, since it narrows the list on top of the saved filter. The chosen tags are dropped when another saved
// filter is chosen, and otherwise kept as long as the filter still offers them.
func (f *listFilterControls[T]) chooseFilter(lf *gurps.ListFilter) {
	if f.tagFilter != nil && lf != f.selectedFilter {
		f.tagFilter.clear()
	}
	f.selectedFilter = lf
	f.dropUnofferedTags()
	f.applyFilter()
}

// clearFilters empties the quick filter and drops the chosen tags and the saved filter, applying the result once.
func (f *listFilterControls[T]) clearFilters() {
	if f.tagFilter != nil {
		f.tagFilter.clear()
	}
	f.selectedFilter = nil
	if f.savedFilters != nil {
		f.savedFilters.rebuildItems()
	}
	if f.filterField.Text() != "" {
		f.filterField.SetText("") // Its modified callback applies the filter.
	} else {
		f.applyFilter()
	}
}

// dropUnofferedTags drops the chosen tags the list no longer offers. With no tag chosen there is nothing to drop, and
// the tag filter gathers the tags afresh when its menu is opened, so the list isn't gathered from here for nothing.
func (f *listFilterControls[T]) dropUnofferedTags() {
	if f.tagFilter != nil && len(f.tagFilter.selected) != 0 {
		f.tagFilter.rebuildItems()
	}
}

// saveTagFilter puts up the editor for a new saved filter that requires what the saved filter in force does and the
// chosen tags as well.
func (f *listFilterControls[T]) saveTagFilter() {
	f.savedFilters.newFilter(gurps.NewCombinedListFilter(f.selectedFilter, f.tagFilter.chosenTags()))
}

// listFiltersChanged implements listFilterObserver.
func (f *listFilterControls[T]) listFiltersChanged(key string, source *listFilterPopup) {
	if f.savedFilters != nil && f.savedFilters != source && f.savedFilters.spec.key == key {
		f.savedFilters.refresh()
	}
}

// applyFilter applies the saved filter in force, requiring the chosen tags as well, and the quick filter's text
// together, showing a row only when it passes both, and reports whether the table was synced to its model, which
// unison.Table.ApplyHierarchicalFilter does whenever it is given a filter or has one to clear. The hierarchy is kept so
// a matching row is seen beneath the containers that hold it; a container shown only for that reason is dimmed (see
// Node.cellData).
func (f *listFilterControls[T]) applyFilter() (synced bool) {
	if f.filterField == nil {
		return false
	}
	// The table's filter is told which rows to drop, so each of these rejects the rows that fail its filter.
	var saved, quick func(row *Node[T]) bool
	var tags []string
	if f.tagFilter != nil {
		tags = f.tagFilter.chosenTags()
	}
	if f.selectedFilter != nil || len(tags) != 0 {
		// The fields are looked up once here rather than once per row.
		m := gurps.NewListFilterMatcher(gurps.NewCombinedListFilter(f.selectedFilter, tags), f.provider.FilterFields())
		saved = func(row *Node[T]) bool { return !m(row.Data()) }
	}
	if text := strings.ToLower(strings.TrimSpace(f.filterField.GetFieldState().Text)); text != "" {
		match := f.quickMatch
		if match == nil {
			match = (*Node[T]).Match // Looks at every column, the tags column included.
		}
		quick = func(row *Node[T]) bool { return !match(row, text) }
	}
	var drop func(row *Node[T]) bool
	switch {
	case saved != nil && quick != nil:
		drop = func(row *Node[T]) bool { return saved(row) || quick(row) }
	case saved != nil:
		drop = saved
	case quick != nil:
		drop = quick
	case !f.table.IsFiltered():
		return false
	}
	matched := 0
	if drop != nil {
		// The table puts each row through the filter exactly once.
		rejects := drop
		drop = func(row *Node[T]) bool {
			if rejects(row) {
				return true
			}
			matched++
			return false
		}
	}
	f.table.ApplyHierarchicalFilter(drop)
	filtered := f.table.IsFiltered()
	f.hierarchyButton.SetEnabled(!filtered)
	f.noteToggleButton.SetEnabled(!filtered)
	if f.filterApplied != nil {
		f.filterApplied(filtered, matched)
	}
	return true
}
