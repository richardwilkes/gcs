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
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// listFilterSettingsFixture is the Saved Filters view open over two trait filters and one skill filter.
type listFilterSettingsFixture struct {
	screen *unison.HeadlessScreen
	view   *listFilterSettingsDockable
	alpha  *gurps.ListFilter
	beta   *gurps.ListFilter
	gamma  *gurps.ListFilter
}

func newListFilterSettingsFixture(t *testing.T, c check.Checker) *listFilterSettingsFixture {
	t.Helper()
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	f := &listFilterSettingsFixture{
		screen: screen,
		alpha:  seedListFilter("Alpha", "tags", "a"),
		beta:   seedListFilter("Beta", "tags", "b"),
		gamma:  gurps.NewListFilter("Gamma"),
	}
	f.gamma.Root.Children = append(f.gamma.Root.Children, gurps.NewFilterCondition(f.gamma.Root, "tags"))
	gurps.GlobalSettings().AddListFilter("skl", f.gamma)
	screen.Do(ShowListFilterSettings)
	f.view = uxtest.SoleEditor[*listFilterSettingsDockable](t, screen, func(d unison.Dockable) bool {
		_, isOne := d.AsPanel().Self.(*listFilterSettingsDockable)
		return isOne
	})
	return f
}

// savedNames returns the names of the saved filters for the list type with the key.
func savedNames(key string) []string {
	return listFilterNames(gurps.GlobalSettings().ListFiltersFor(key))
}

// TestListFilterSettingsCommands verifies each toolbar command, that each is one undo step, that undoing an edit or a
// deletion puts back the same filter rather than a copy, and that the selection survives the rebuilds.
func TestListFilterSettingsCommands(t *testing.T) {
	c := check.New(t)
	f := newListFilterSettingsFixture(t, c)
	f.screen.Do(func() {
		v := f.view
		c.Equal([]string{"Alpha", "Beta"}, savedNames("adq"), "precondition: the trait filters")

		v.selectFilters([]*gurps.ListFilter{f.alpha})
		v.duplicateFilters()
		c.Equal([]string{"Alpha", "Alpha (2)", "Beta"}, savedNames("adq"), "a copy is named after its source")
		c.Equal("Undo Duplicate", v.undoMgr.UndoTitle())
		v.undoMgr.Undo()
		c.Equal([]string{"Alpha", "Beta"}, savedNames("adq"), "undo removes the copy")
		v.undoMgr.Redo()
		c.Equal([]string{"Alpha", "Alpha (2)", "Beta"}, savedNames("adq"), "redo brings the copy back")
		v.undoMgr.Undo()

		uxtest.SwapForTest(t, &showFilterEditor,
			func(_, _ string, _ *gurps.ListFilter, _ []filterFieldInfo, _ *gurps.ListFilter) bool { return true })
		v.selectFilters([]*gurps.ListFilter{f.alpha})
		v.editFilter()
		c.Equal("Redo Duplicate", v.undoMgr.RedoTitle(), "an edit that changes nothing is no undo step")
		uxtest.SwapForTest(t, &showFilterEditor,
			func(_, _ string, filter *gurps.ListFilter, _ []filterFieldInfo, _ *gurps.ListFilter) bool {
				filter.Name = "Zeta"
				return true
			})
		v.editFilter()
		c.Equal("Zeta", f.alpha.Name, "the edit lands in the saved filter itself")
		c.Equal([]*gurps.ListFilter{f.alpha}, v.selectedFilters(), "the edited filter stays selected")
		v.undoMgr.Undo()
		c.Equal("Alpha", f.alpha.Name, "undoing the edit restores the same filter")
		c.Equal([]*gurps.ListFilter{f.alpha}, v.selectedFilters(), "the filter stays selected through the undo")

		v.selectFilters([]*gurps.ListFilter{f.alpha, f.gamma})
		c.False(v.mergeButton.Enabled(), "filters of different list types can't be merged")
		v.selectFilters([]*gurps.ListFilter{f.alpha, f.beta})
		c.True(v.mergeButton.Enabled(), "filters of one list type can be merged")
		uxtest.SwapForTest(t, &showMergeFilters,
			func(_ string, _ []*gurps.ListFilter) (string, bool, bool) { return "Either", false, true })
		v.mergeFilters()
		c.Equal([]string{"Alpha", "Beta", "Either"}, savedNames("adq"), "the merge adds a filter and keeps the sources")
		merged := gurps.GlobalSettings().ListFiltersFor("adq")[2]
		c.False(merged.Root.All, "the merge requires any one source, as asked")
		c.Equal(2, len(merged.Root.Children), "each source's lone condition is folded into the merged filter")
		v.undoMgr.Undo()

		uxtest.SwapForTest(t, &confirmDeletion, func(string) bool { return true })
		v.selectFilters([]*gurps.ListFilter{f.alpha, f.gamma})
		v.deleteFilters()
		c.Equal([]string{"Beta"}, savedNames("adq"), "the selected trait filter is deleted")
		c.Equal(0, len(savedNames("skl")), "the selected skill filter is deleted")
		v.undoMgr.Undo()
		c.True(gurps.GlobalSettings().ListFiltersFor("adq")[0] == f.alpha, "undoing the deletion restores the original")
		c.True(gurps.GlobalSettings().ListFiltersFor("skl")[0] == f.gamma, "for every list type")
	})
}

// TestListFilterSettingsListTypeSelection verifies that selecting a list type's row selects every filter of that type,
// that the Delete key deletes the selection, and that the table asks for some width of its own.
func TestListFilterSettingsListTypeSelection(t *testing.T) {
	c := check.New(t)
	f := newListFilterSettingsFixture(t, c)
	var asked string
	uxtest.SwapForTest(t, &confirmDeletion, func(question string) bool {
		asked = question
		return true
	})
	f.screen.Do(func() {
		v := f.view
		_, pref, _ := v.table.Sizes(geom.Size{})
		c.True(pref.Width > v.table.Columns[0].Current, "the table asks for room for the summaries")

		v.table.SetSelectionMap(map[tid.TID]bool{v.rows.ids["adq"]: true, v.rows.ids[f.alpha]: true})
		c.Equal([]*gurps.ListFilter{f.alpha, f.beta}, v.selectedFilters(),
			"a selected list type brings in each of its filters once")
		c.True(v.HasSelection(), "Export Selected is offered for a selected list type")

		c.True(v.table.KeyDownCallback(unison.KeyDelete, 0, false), "the Delete key is handled")
		c.Equal("Delete the 2 selected filters?", asked, "the Delete key asks to delete the selection")
		c.Equal(0, len(savedNames("adq")), "every filter of the list type is deleted")
		c.Equal([]string{"Gamma"}, savedNames("skl"), "other list types are left alone")
	})
}

// TestListFilterSettingsRecordsPopupChanges verifies that a change made through a list's saved filter popup while the
// view is open becomes an undo step of the view's.
func TestListFilterSettingsRecordsPopupChanges(t *testing.T) {
	c := check.New(t)
	f := newListFilterSettingsFixture(t, c)
	list := openListFilterTraitDockable(t, f.screen)
	uxtest.SwapForTest(t, &showFilterEditor,
		func(_, _ string, filter *gurps.ListFilter, _ []filterFieldInfo, _ *gurps.ListFilter) bool {
			filter.Name = "Bravo"
			return true
		})
	f.screen.Do(func() {
		c.False(f.view.undoMgr.CanUndo(), "precondition: nothing to undo")
		list.savedFilters.selectFilter(f.beta)
		list.savedFilters.editFilter()
		c.Equal([]string{"Alpha", "Bravo"}, savedNames("adq"), "the popup renamed the filter")
		c.Equal("Undo Change Saved Filters", f.view.undoMgr.UndoTitle())
		f.view.undoMgr.Undo()
		c.Equal([]string{"Alpha", "Beta"}, savedNames("adq"), "undo here reverts the popup's change")
		c.True(list.selectedFilter == f.beta, "the list keeps the filter in force")
	})
}

// TestListFilterSettingsExportImport verifies that Export Selected writes only the selection and Export All every
// filter, and that an imported filter whose name is in use is renamed.
func TestListFilterSettingsExportImport(t *testing.T) {
	c := check.New(t)
	f := newListFilterSettingsFixture(t, c)
	one := filepath.Join(t.TempDir(), "one"+gurps.ListFiltersExt)
	all := filepath.Join(t.TempDir(), "all"+gurps.ListFiltersExt)
	var offered []importedListFilter
	uxtest.SwapForTest(t, &showImportFilters,
		func(_ string, set map[string][]*gurps.ListFilter) []importedListFilter {
			offered = nil
			for key, filters := range set {
				for _, lf := range filters {
					offered = append(offered, importedListFilter{key: key, filter: lf})
				}
			}
			return offered
		})
	f.screen.Do(func() {
		v := f.view
		c.False(v.HasSelection(), "precondition: nothing is selected")
		v.selectFilters([]*gurps.ListFilter{f.beta})
		c.True(v.HasSelection(), "Export Selected is offered once a filter is selected")
		c.NoError(v.saveSelected(one))
		c.NoError(v.Saver(all))

		c.NoError(v.load(diskFileRef(all)))
		c.Equal(3, len(offered), "Export All writes every filter, whatever is selected")
		v.undoMgr.Undo()
		c.NoError(v.load(diskFileRef(one)))
		c.Equal(1, len(offered), "Export Selected writes only the selection")
		c.Equal([]string{"Alpha", "Beta", "Beta (2)"}, savedNames("adq"), "a name in use gets a number")
		c.Equal("Undo Import", v.undoMgr.UndoTitle())
	})
}

// TestSummarizeListFilter verifies the summary of a filter's conditions in the Saved Filters view.
func TestSummarizeListFilter(t *testing.T) {
	c := check.New(t)
	fields := filterFieldInfos(gurps.TraitFilterFields(), nil)
	f := gurps.NewListFilter("x")
	c.Equal("No conditions", summarizeListFilter(f, fields))
	cond := gurps.NewFilterCondition(f.Root, "tags")
	cond.Text.Compare = criteria.IsText
	cond.Text.Qualifier = "Mental"
	f.Root.Children = append(f.Root.Children, cond, gurps.NewFilterGroup(f.Root))
	c.Equal("Must have tags where at least one is Mental; All of (No conditions)", summarizeListFilter(f, fields))
	f.Root.All = false
	c.Equal("Any of (Must have tags where at least one is Mental; All of (No conditions))",
		summarizeListFilter(f, fields))
}
