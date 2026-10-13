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
	"slices"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ unison.UndoManagerProvider = &listFilterSettingsDockable{}
	_ listFilterObserver         = &listFilterSettingsDockable{}
)

// Seams so the commands can be exercised without a window.
var (
	showMergeFilters  = showMergeListFiltersDialog
	showImportFilters = showImportListFiltersDialog
	confirmDeletion   = func(question string) bool {
		return unison.QuestionDialog(question, i18n.Text("You can undo this from the Edit menu.")) ==
			unison.ModalResponseOK
	}
)

// listFilterSettingsDockable is the view that manages the saved filters of every list type at once. Every change made
// through it is an undo step, and so is a change made through a list's saved filter popup while it is open.
type listFilterSettingsDockable struct {
	SettingsDockable
	undoMgr         *unison.UndoManager
	table           *unison.Table[*listFilterRow]
	rows            listFilterRows
	editButton      *unison.Button
	duplicateButton *unison.Button
	mergeButton     *unison.Button
	deleteButton    *unison.Button
	// snapshot holds the saved filters as of the last change this view knows of, which is what a change made through a
	// popup is recorded against.
	snapshot listFilterSnapshot
}

// ShowListFilterSettings shows the Saved Filters view.
func ShowListFilterSettings() {
	if ActivateDockableOfType[*listFilterSettingsDockable](nil) {
		return
	}
	d := &listFilterSettingsDockable{
		undoMgr:  unison.NewUndoManager(100, func(err error) { errs.Log(err) }),
		snapshot: takeListFilterSnapshot(),
	}
	d.RefLoader = d.load
	d.SelectionSaver = d.saveSelected
	d.HasSelection = func() bool { return len(d.selectedFilters()) != 0 }
	d.initSettings(d, &settingsSpec{
		title: i18n.Text("Saved Filters"),
		ext:   gurps.ListFiltersExt,
		saver: func(filePath string) error {
			return gurps.SaveListFilters(gurps.GlobalSettings().ListFilters, filePath)
		},
		addToStartToolbar: d.addToStartToolbar,
		initContent:       d.initContent,
	})
}

// UndoManager implements unison.UndoManagerProvider.
func (d *listFilterSettingsDockable) UndoManager() *unison.UndoManager {
	return d.undoMgr
}

func (d *listFilterSettingsDockable) addToStartToolbar(toolbar *unison.Panel) {
	addButton := func(icon *unison.SVG, tip string, click func()) *unison.Button {
		b := unison.NewSVGButton(icon)
		b.Tooltip = newWrappedTooltip(tip)
		b.ClickCallback = click
		toolbar.AddChild(b)
		return b
	}
	addHelpButton(toolbar, "md:User%20Guide/Saved%20Filters")
	toolbar.AddChild(NewToolbarSeparator())
	var newButton *unison.Button
	newButton = addButton(unison.CircledAddSVG, i18n.Text("New Filter…"), func() { d.newFilter(newButton) })
	d.editButton = addButton(svg.Edit, i18n.Text("Edit Filter…"), d.editFilter)
	d.duplicateButton = addButton(svg.Copy, i18n.Text("Duplicate"), d.duplicateFilters)
	d.mergeButton = addButton(svg.Merge, i18n.Text("Merge…"), d.mergeFilters)
	d.deleteButton = addButton(unison.TrashSVG, i18n.Text("Delete…"), d.deleteFilters)
}

func (d *listFilterSettingsDockable) initContent(content *unison.Panel) {
	content.SetLayout(&unison.FlexLayout{Columns: 1})
	d.table = unison.NewTable(&unison.SimpleTableModel[*listFilterRow]{})
	d.table.Accessibility.Name = i18n.Text("Saved Filters")
	d.table.Columns = []unison.ColumnInfo{{ID: 0}, {ID: 1}}
	d.table.ShowFirstColumnDivider = false
	d.table.ShowLastColumnDivider = false
	d.table.SelectionChangedCallback = d.updateButtons
	d.table.DoubleClickCallback = d.editFilter
	d.table.ContextMenuCallback = func(_ geom.Point) unison.Menu {
		f := unison.DefaultMenuFactory()
		cm := f.NewMenu(unison.PopupMenuTemporaryBaseID|unison.ContextMenuIDFlag, "", nil)
		id := 1
		for _, b := range []*unison.Button{d.editButton, d.duplicateButton, d.mergeButton, d.deleteButton} {
			cm.InsertItem(-1, newContextMenuItemFromButton(f, &id, b))
		}
		return cm
	}
	d.table.FrameChangeCallback = func() { d.table.SizeColumnsToFitWithExcessIn(1) }
	// The table asks for no width of its own, which would otherwise be the width of its columns as last sized, so that
	// it takes the width it is given and the summaries wrap to fit it, however narrow the view becomes.
	d.table.SetSizer(func(hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
		minSize, prefSize, maxSize = d.table.DefaultSizes(hint)
		minSize.Width = 0
		prefSize.Width = 0
		return minSize, prefSize, maxSize
	})
	d.table.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	content.AddChild(d.table)
	d.sync()
}

// sync rebuilds the table from the saved filters, keeping the selection of the rows that remain.
func (d *listFilterSettingsDockable) sync() {
	sel := d.table.CopySelectionMap()
	d.table.SetRootRows(d.rows.build())
	d.table.SetSelectionMap(sel)
	d.table.SizeColumnsToFitWithExcessIn(1)
	d.updateButtons()
}

// selectFilters makes the rows of the given filters the selection.
func (d *listFilterSettingsDockable) selectFilters(filters []*gurps.ListFilter) {
	d.table.SetSelectionMap(d.rows.selectionFor(filters))
}

// selectedFilters returns the filters whose rows are selected, in the order shown, leaving out the list type rows.
func (d *listFilterSettingsDockable) selectedFilters() []*gurps.ListFilter {
	var filters []*gurps.ListFilter
	for _, row := range d.table.SelectedRows(false) {
		if row.filter != nil {
			filters = append(filters, row.filter)
		}
	}
	return filters
}

func (d *listFilterSettingsDockable) updateButtons() {
	filters := d.selectedFilters()
	d.editButton.SetEnabled(len(filters) == 1)
	d.duplicateButton.SetEnabled(len(filters) != 0)
	d.mergeButton.SetEnabled(d.canMerge(filters))
	d.deleteButton.SetEnabled(len(filters) != 0)
}

// canMerge reports whether the filters can be merged: two or more for the same list type, not all of them empty.
func (d *listFilterSettingsDockable) canMerge(filters []*gurps.ListFilter) bool {
	if len(filters) < 2 {
		return false
	}
	key := d.rows.kindOf(filters[0]).key
	return !slices.ContainsFunc(filters, func(f *gurps.ListFilter) bool { return d.rows.kindOf(f).key != key }) &&
		slices.ContainsFunc(filters, func(f *gurps.ListFilter) bool { return !f.Root.Empty() })
}

// newFilter asks which list type the filter is for, then puts up the editor for it and, if it is accepted, saves it.
func (d *listFilterSettingsDockable) newFilter(button *unison.Button) {
	f := unison.DefaultMenuFactory()
	m := f.NewMenu(unison.PopupMenuTemporaryBaseID, "", nil)
	for i, kind := range listFilterKinds() {
		m.InsertItem(-1, f.NewItem(unison.PopupMenuTemporaryBaseID+i+1, kind.title, unison.KeyBinding{}, nil,
			func(_ unison.MenuItem) {
				filter := gurps.NewListFilter("")
				if showFilterEditor(i18n.Text("New Filter"), kind.key, filter, kind.fields(), nil) {
					d.change(i18n.Text("New Filter"), func() { gurps.GlobalSettings().AddListFilter(kind.key, filter) })
					d.selectFilters([]*gurps.ListFilter{filter})
				}
			}))
	}
	m.Popup(button.RectToRoot(button.ContentRect(true)), 0)
}

// editFilter puts up the editor for the selected filter. A clone is edited so that a cancel leaves the saved filter
// untouched; an accepted edit is copied into the saved filter, so every list with it in force keeps it.
func (d *listFilterSettingsDockable) editFilter() {
	filters := d.selectedFilters()
	if len(filters) != 1 {
		return
	}
	current := filters[0]
	kind := d.rows.kindOf(current)
	clone := current.Clone()
	if showFilterEditor(i18n.Text("Edit Filter"), kind.key, clone, kind.fields(), current) &&
		gurps.Hash64(clone) != gurps.Hash64(current) {
		d.change(i18n.Text("Edit Filter"), func() {
			*current = *clone
			gurps.GlobalSettings().ResortListFilters(kind.key)
		})
	}
}

// duplicateFilters adds a copy of each selected filter, named after it, and selects the copies.
func (d *listFilterSettingsDockable) duplicateFilters() {
	var copies []*gurps.ListFilter
	d.change(i18n.Text("Duplicate"), func() {
		s := gurps.GlobalSettings()
		for _, f := range d.selectedFilters() {
			key := d.rows.kindOf(f).key
			one := f.Clone()
			one.Name = s.UniqueListFilterName(key, f.Name)
			s.AddListFilter(key, one)
			copies = append(copies, one)
		}
	})
	d.selectFilters(copies)
}

// mergeFilters asks for the name of a new filter that holds the selected ones and whether it requires all of them or
// any one, then adds it. The sources are kept.
func (d *listFilterSettingsDockable) mergeFilters() {
	filters := d.selectedFilters()
	if !d.canMerge(filters) {
		return
	}
	key := d.rows.kindOf(filters[0]).key
	if name, all, ok := showMergeFilters(key, filters); ok {
		merged := gurps.NewMergedListFilter(name, all, filters)
		d.change(i18n.Text("Merge"), func() { gurps.GlobalSettings().AddListFilter(key, merged) })
		d.selectFilters([]*gurps.ListFilter{merged})
	}
}

// deleteFilters deletes the selected filters, once the user has confirmed it.
func (d *listFilterSettingsDockable) deleteFilters() {
	filters := d.selectedFilters()
	question := i18n.Text("Delete the %d selected filters?", len(filters))
	if len(filters) == 1 {
		question = i18n.Text("Delete the filter '%s'?", filters[0].Name)
	}
	if confirmDeletion(question) {
		d.change(i18n.Text("Delete"), func() {
			for _, f := range filters {
				gurps.GlobalSettings().RemoveListFilter(d.rows.kindOf(f).key, f)
			}
		})
	}
}

// load offers the filters in the file for import, and adds the ones the user keeps.
func (d *listFilterSettingsDockable) load(ref *library.NamedFileRef) error {
	set, err := gurps.NewListFiltersFromFile(ref.FileSystem, ref.FilePath)
	if err != nil {
		return err
	}
	source := ref.DiskPath
	if source == "" {
		source = ref.Name
	}
	chosen := showImportFilters(source, set)
	if len(chosen) == 0 {
		return nil
	}
	added := make([]*gurps.ListFilter, len(chosen))
	d.change(i18n.Text("Import"), func() {
		s := gurps.GlobalSettings()
		for i, one := range chosen {
			one.filter.Name = s.UniqueListFilterName(one.key, one.filter.Name)
			s.AddListFilter(one.key, one.filter)
			added[i] = one.filter
		}
	})
	d.selectFilters(added)
	return nil
}

// saveSelected writes the selected filters to the file.
func (d *listFilterSettingsDockable) saveSelected(filePath string) error {
	set := make(map[string][]*gurps.ListFilter)
	for _, f := range d.selectedFilters() {
		key := d.rows.kindOf(f).key
		set[key] = append(set[key], f)
	}
	return gurps.SaveListFilters(set, filePath)
}

// change makes a change to the saved filters as one undo step named title, then saves them and brings every list
// showing them into line.
func (d *listFilterSettingsDockable) change(title string, mutate func()) {
	before := takeListFilterSnapshot()
	mutate()
	d.record(title, before)
	d.filtersChanged()
}

// record posts an undo step named title that goes from before to the saved filters as they now stand.
func (d *listFilterSettingsDockable) record(title string, before listFilterSnapshot) {
	d.snapshot = takeListFilterSnapshot()
	restore := func(snapshot listFilterSnapshot) {
		snapshot.restore()
		d.snapshot = snapshot
		d.filtersChanged()
	}
	d.undoMgr.Add(&unison.UndoEdit[listFilterSnapshot]{
		ID:         unison.NextUndoID(),
		EditName:   title,
		UndoFunc:   func(e *unison.UndoEdit[listFilterSnapshot]) { restore(e.BeforeData) },
		RedoFunc:   func(e *unison.UndoEdit[listFilterSnapshot]) { restore(e.AfterData) },
		AbsorbFunc: func(_ *unison.UndoEdit[listFilterSnapshot], _ unison.Undoable) bool { return false },
		BeforeData: before,
		AfterData:  d.snapshot,
	})
}

// filtersChanged saves the settings, rebuilds the table and tells every list showing saved filters to catch up. The
// lists are told with no source, which is how listFiltersChanged knows not to record the change a second time.
func (d *listFilterSettingsDockable) filtersChanged() {
	saveGlobalSettings()
	d.sync()
	for _, key := range gurps.ListFilterKeys() {
		notifyListFiltersChanged(key, nil)
	}
}

// listFiltersChanged implements listFilterObserver. A change made through a list's popup becomes an undo step here, so
// that undoing here can't silently revert it.
func (d *listFilterSettingsDockable) listFiltersChanged(_ string, source *listFilterPopup) {
	if source != nil {
		d.record(i18n.Text("Change Saved Filters"), d.snapshot)
		d.sync()
	}
}

// listFilterSnapshot holds every saved filter as it stood at one moment, by list type. Each filter is kept with a copy
// of its contents, so that restoring the snapshot puts the same filter back rather than a copy of it, and every list
// with that filter in force keeps it.
type listFilterSnapshot map[string][]savedListFilter

type savedListFilter struct {
	live     *gurps.ListFilter
	contents *gurps.ListFilter
}

// takeListFilterSnapshot captures the saved filters as they now stand.
func takeListFilterSnapshot() listFilterSnapshot {
	snapshot := make(listFilterSnapshot)
	for key, filters := range gurps.GlobalSettings().ListFilters {
		for _, f := range filters {
			snapshot[key] = append(snapshot[key], savedListFilter{live: f, contents: f.Clone()})
		}
	}
	return snapshot
}

// restore makes the saved filters what the snapshot holds, leaving the snapshot as it was so it can be restored again.
func (s listFilterSnapshot) restore() {
	set := make(map[string][]*gurps.ListFilter, len(s))
	for key, saved := range s {
		for _, one := range saved {
			*one.live = *one.contents.Clone()
			set[key] = append(set[key], one.live)
		}
	}
	gurps.GlobalSettings().ListFilters = set
}
