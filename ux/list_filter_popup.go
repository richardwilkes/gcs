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
	"slices"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// listFilterQuickIndex is the index of the "Quick Filter" entry, which is always the first item in the popup.
const listFilterQuickIndex = 0

// Seams so the commands can be exercised without a window.
var (
	showFilterEditor      = showListFilterDialog
	confirmFilterDeletion = func(name string) bool {
		return unison.QuestionDialog(fmt.Sprintf(i18n.Text("Delete the filter %q?"), name),
			i18n.Text("This cannot be undone.")) == unison.ModalResponseOK
	}
)

// listFilterObserver is implemented by a dockable that shows a saved filter popup, so that it can be told when the
// saved filters of a list type were changed through the popup of another dockable.
type listFilterObserver interface {
	// listFiltersChanged is called after the saved filters for the list type with the given key were changed through
	// source, which is nil when the change didn't come from a popup.
	listFiltersChanged(key string, source *listFilterPopup)
}

// notifyListFiltersChanged tells every open dockable that the saved filters for the list type with the given key have
// changed, so that the ones showing that list type can bring their popups into line. The popup the change was made
// through is passed as source, since it has already done so.
func notifyListFiltersChanged(key string, source *listFilterPopup) {
	for _, d := range AllDockables() {
		if observer, ok := d.(listFilterObserver); ok {
			observer.listFiltersChanged(key, source)
		}
	}
}

// listFilterPopupSpec is what the saved filter popup needs from the dockable that shows it.
type listFilterPopupSpec struct {
	// key is the list type key the saved filters are stored under, such as "adq".
	key string
	// fields are the fields a filter for this list type may test.
	fields []filterFieldInfo
	// current returns the filter in force, which is nil when the quick filter has the list.
	current func() *gurps.ListFilter
	// choose puts a filter in force. A nil filter hands the list back to the quick filter.
	choose func(f *gurps.ListFilter)
}

// listFilterPopup is the toolbar popup that chooses among a list type's saved filters and creates, edits and deletes
// them. The popup's items are rebuilt from the global settings each time it is about to be shown, since another
// dockable showing the same list type may have changed them in the meantime.
type listFilterPopup struct {
	popup *unison.PopupMenu[string]
	spec  listFilterPopupSpec
	// filters runs parallel to the popup's items, holding the filter each item stands for. It is nil for the quick
	// filter, the separators and the three commands, since separators occupy an index of their own.
	filters     []*gurps.ListFilter
	newIndex    int
	editIndex   int
	deleteIndex int
	// rebuilding suppresses the selection callback while the items are being replaced, since emptying the popup and
	// filling it again moves the selection without the user having chosen anything.
	rebuilding bool
}

// newListFilterPopup creates the saved filter popup for the list type the spec names.
func newListFilterPopup(spec listFilterPopupSpec) *listFilterPopup {
	p := &listFilterPopup{
		popup: unison.NewPopupMenu[string](),
		spec:  spec,
	}
	p.popup.WillShowMenuCallback = func(_ *unison.PopupMenu[string]) {
		if p.rebuildItems() {
			// The filter that was in force is gone, so the list falls back to the quick filter.
			p.spec.choose(nil)
		}
	}
	p.popup.ChoiceMadeCallback = func(popup *unison.PopupMenu[string], index int, _ string) {
		// The commands act and leave the selection where it was; only a filter, or the quick filter, becomes the
		// selection.
		switch index {
		case p.newIndex:
			p.newFilter()
		case p.editIndex:
			p.editFilter()
		case p.deleteIndex:
			p.deleteFilter()
		default:
			popup.SelectIndex(index)
		}
	}
	p.popup.SelectionChangedCallback = func(popup *unison.PopupMenu[string]) {
		if p.rebuilding {
			return
		}
		i := popup.SelectedIndex()
		p.enableCommands(i != listFilterQuickIndex)
		p.spec.choose(p.filterAt(i))
	}
	p.popup.Tooltip = newWrappedTooltipWithSecondaryText(i18n.Text("Saved Filters"),
		i18n.Text("Choose a saved filter, or create, edit and delete them"))
	p.popup.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	// Fill the popup in now so that it shows "Quick Filter" before it has ever been opened.
	p.rebuildItems()
	return p
}

// rebuildItems replaces the popup's items with the saved filters as they now stand, keeping the filter in force
// selected. It reports whether that filter is no longer among them, which happens when another dockable showing the
// same list type deleted it, in which case the selection has fallen back to the quick filter.
func (p *listFilterPopup) rebuildItems() (lostCurrent bool) {
	defer func(saved bool) { p.rebuilding = saved }(p.rebuilding)
	p.rebuilding = true
	current := p.spec.current()
	p.popup.RemoveAllItems()
	p.filters = nil
	addItem := func(title string, f *gurps.ListFilter) int {
		index := len(p.filters)
		p.popup.AddItem(title)
		p.filters = append(p.filters, f)
		return index
	}
	addSeparator := func() {
		p.popup.AddSeparator()
		p.filters = append(p.filters, nil)
	}
	addItem(i18n.Text("Quick Filter"), nil)
	index := listFilterQuickIndex
	if saved := gurps.GlobalSettings().ListFiltersFor(p.spec.key); len(saved) != 0 {
		addSeparator()
		for _, f := range saved {
			// The filter in force is found by identity rather than by name, since a saved filter may bear the same
			// name as one of the popup's own entries.
			if i := addItem(f.Name, f); f == current {
				index = i
			}
		}
	}
	addSeparator()
	p.newIndex = addItem(i18n.Text("New Filter…"), nil)
	p.editIndex = addItem(i18n.Text("Edit Filter…"), nil)
	p.deleteIndex = addItem(i18n.Text("Delete Filter…"), nil)
	p.popup.SelectIndex(index)
	p.enableCommands(index != listFilterQuickIndex)
	// The popup is as wide as its widest item, so adding, removing or renaming a filter can change its size, and the
	// toolbar it sits in has to be laid out again to make room. Marking the popup alone wouldn't do it: a parent that
	// is laid out again doesn't revisit its children, so the whole chain up to the window is marked.
	p.popup.MarkForLayoutRecursivelyUpward()
	p.popup.MarkForRedraw()
	return current != nil && index == listFilterQuickIndex
}

// filterAt returns the filter the item at the given index stands for, or nil when the item is the quick filter, a
// separator, a command or out of range.
func (p *listFilterPopup) filterAt(index int) *gurps.ListFilter {
	if index < 0 || index >= len(p.filters) {
		return nil
	}
	return p.filters[index]
}

// enableCommands turns the Edit and Delete commands on or off. They only apply to a saved filter, so they are off
// while the quick filter has the list.
func (p *listFilterPopup) enableCommands(enabled bool) {
	p.popup.SetItemEnabledAt(p.editIndex, enabled)
	p.popup.SetItemEnabledAt(p.deleteIndex, enabled)
}

// selectFilter rebuilds the popup around the given filter, selects it, and puts it in force. Pass nil to hand the list
// back to the quick filter.
func (p *listFilterPopup) selectFilter(f *gurps.ListFilter) {
	func() {
		defer func(saved bool) { p.rebuilding = saved }(p.rebuilding)
		p.rebuilding = true
		p.rebuildItems()
		index := listFilterQuickIndex
		if f != nil {
			if i := slices.Index(p.filters, f); i != -1 {
				index = i
			}
		}
		p.popup.SelectIndex(index)
		p.enableCommands(index != listFilterQuickIndex)
	}()
	// The filter is put in force unconditionally: editing a filter without renaming it doesn't move the selection, so
	// the selection callback would not fire, yet the contents changed and the list has to be filtered again.
	p.spec.choose(f)
}

// refresh brings the popup into line with the saved filters after they were changed elsewhere: the items are rebuilt,
// and the filter in force is either put through again, since its contents may have changed, or dropped in favor of
// the quick filter when it no longer exists.
func (p *listFilterPopup) refresh() {
	if p.rebuildItems() {
		p.spec.choose(nil)
	} else if current := p.spec.current(); current != nil {
		p.spec.choose(current)
	}
}

// newFilter puts up the editor for a new filter and, if it is accepted, saves it and puts it in force.
func (p *listFilterPopup) newFilter() {
	f := gurps.NewListFilter("")
	if showFilterEditor(i18n.Text("New Filter"), p.spec.key, f, p.spec.fields, nil) {
		gurps.GlobalSettings().AddListFilter(p.spec.key, f)
		p.filtersChanged(f)
	}
}

// editFilter puts up the editor for the filter in force. A clone is edited so that a cancel leaves the saved filter
// untouched. When the edit is accepted, the clone's contents are copied into the saved filter rather than replacing
// it, so that every other dockable with the same filter in force keeps it, and sees the change.
func (p *listFilterPopup) editFilter() {
	current := p.spec.current()
	if current == nil {
		return
	}
	clone := current.Clone()
	if showFilterEditor(i18n.Text("Edit Filter"), p.spec.key, clone, p.spec.fields, current) {
		*current = *clone
		gurps.GlobalSettings().ResortListFilters(p.spec.key) // The name may have changed.
		p.filtersChanged(current)
	}
}

// deleteFilter removes the filter in force, once the user has confirmed it, and hands the list back to the quick
// filter.
func (p *listFilterPopup) deleteFilter() {
	current := p.spec.current()
	if current == nil {
		return
	}
	if confirmFilterDeletion(current.Name) {
		gurps.GlobalSettings().RemoveListFilter(p.spec.key, current)
		p.filtersChanged(nil)
	}
}

// filtersChanged finishes a change made to the saved filters through this popup: the settings are saved, the given
// filter is put in force here, and every other dockable showing the same list type is told to catch up.
func (p *listFilterPopup) filtersChanged(f *gurps.ListFilter) {
	saveGlobalSettings()
	p.selectFilter(f)
	notifyListFiltersChanged(p.spec.key, p)
}
