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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/gurpstest"
	"github.com/richardwilkes/gcs/v5/model/paper"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// newContinuousTestSheet returns a sheet for a new character shown as one continuous page, for the tests that look at
// the page as a whole.
func newContinuousTestSheet(t *testing.T) *Sheet {
	t.Helper()
	uxtest.SwapForTest(t, &gurps.GlobalSettings().General.InitialSheetPaginated, false)
	return newTestSheetForTemplate(t)
}

// plainTraits returns the given number of traits for the entity, none of them a container.
func plainTraits(entity *gurps.Entity, count int) []*gurps.Trait {
	traits := make([]*gurps.Trait, count)
	for i := range traits {
		traits[i] = gurps.NewTrait(entity, nil, false)
		traits[i].Name = fmt.Sprintf("Trait %d", i+1)
	}
	return traits
}

// newPaginatedTestSheetWithTraits returns a sheet shown as pages, outside of any window, whose character has the
// given number of plain traits.
func newPaginatedTestSheetWithTraits(t *testing.T, count int) *Sheet {
	t.Helper()
	uxtest.SwapForTest(t, &gurps.GlobalSettings().General.InitialSheetPaginated, true)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.Traits = plainTraits(entity, count)
	entity.Recalculate()
	sheet.Rebuild(true)
	return sheet
}

// openPaginatedSheet opens a new character sheet shown as pages in the headless workspace, its traits replaced by
// those traits returns.
func openPaginatedSheet(t *testing.T, screen *unison.HeadlessScreen, traits func(entity *gurps.Entity) []*gurps.Trait) *Sheet {
	t.Helper()
	uxtest.SwapForTest(t, &gurps.GlobalSettings().General.InitialSheetPaginated, true)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	screen.Do(func() {
		entity := sheet.Entity()
		entity.Traits = traits(entity)
		entity.Recalculate()
		sheet.Rebuild(true)
		sheet.markUnmodified()
	})
	t.Cleanup(func() { screen.Do(sheet.markUnmodified) })
	return sheet
}

// pageBands returns the bands on the sheet's pages, in order. Filling the pages builds anew each band that holds more
// than one block, so the same bands mean the pages weren't filled again.
func pageBands(sheet *Sheet) []*unison.Panel {
	var bands []*unison.Panel
	for _, page := range sheet.grid.pages {
		bands = append(bands, page.Children()...)
	}
	return bands
}

// TestPaginatedSheetWithNoEmptyListBreaksPagesAsTheExporterDoes verifies that a sheet shown as pages fills pages of the
// paper's size with the sheet's own lists, drawing every row of each exactly once, and breaks them where the exporter
// does for the same character, both as it opens and once laid out again after the paper turned, which reaches the sheet
// as a rebuild that lays nothing out anew. The character has no empty list that can be added to, since the sheet shows
// one and the exporter leaves it off. Leaving a list whole, handing a continuation out twice, sizing a page by its
// content or filling the pages for a change of paper only when a band is next laid out at the new size fails it.
func TestPaginatedSheetWithNoEmptyListBreaksPagesAsTheExporterDoes(t *testing.T) {
	c := check.New(t)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().General.InitialSheetPaginated, true)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	uxtest.SwapForTest(t, &Workspace.DocumentDock, NewDocumentDock())
	size := gurpstest.DefaultLargeEntitySize.Scaled(0.1)
	sheet := NewSheet("test"+gurps.SheetExt, gurpstest.NewLargeEntity(&size))
	for _, one := range []struct {
		name   string
		change func()
	}{
		{name: "as opened"},
		{name: "after the paper turned", change: func() {
			sheet.Entity().SheetSettings.Page.Orientation = paper.Landscape
			sheet.Rebuild(false)
			sheet.grid.ValidateLayout()
		}},
	} {
		if one.change != nil {
			one.change()
		}
		pages := sheet.grid.pages
		c.True(len(pages) > 2, "%s: the character must take several pages, not %d", one.name, len(pages))
		exporter := newPageExporter(sheet.Entity())
		c.Equal(exporter.PageSize(), pages[0].paperSize(), "%s: the paper", one.name)
		for i, page := range pages {
			c.Equal(exporter.PageSize(), page.FrameRect().Size, "%s: page %d must be the size of the paper", one.name,
				i+1)
		}
		c.Equal(len(exporter.pages), len(pages), "%s: the sheet must have as many pages as the exporter", one.name)
		for key, list := range sheet.splittableLists() {
			c.True(list.RowCount() != 0 || isDerivedListKey(key), "%s: the %s list must not be empty", one.name, key)
			ranges := drawnRowRanges(pages, key)
			next := 0
			for i, span := range ranges {
				c.Equal(next, span[0], "%s: the %s rows of part %d must pick up where the part before left off",
					one.name, key, i+1)
				next = span[1]
			}
			c.Equal(list.RowCount(), next, "%s: every %s row must be drawn", one.name, key)
			c.Equal(drawnRowRanges(exporter.pages, key), ranges, "%s: the %s rows must break where the exporter's do",
				one.name, key)
		}
	}
}

// TestPaginatedSheetFillsThePagesAgainWhenARowGrows verifies that a row of a list growing taller has the pages filled
// again once the sheet is laid out, so that no band runs past the bottom of its page, although the list still has as
// many rows. Filling the pages again only when a list's number of rows changes fails it.
func TestPaginatedSheetFillsThePagesAgainWhenARowGrows(t *testing.T) {
	c := check.New(t)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().SheetSettings().NotesDisplay, display.Inline)
	sheet := newPaginatedTestSheetWithTraits(t, 150)
	entity := sheet.Entity()
	entity.Traits[5].LocalNotes = strings.Repeat("A line of notes\n", 30)
	sheet.MarkModified(sheet)
	sheet.grid.ValidateLayout()
	for i, page := range sheet.grid.pages {
		bottom := page.paperSize().Height - page.insets().Bottom
		for _, band := range page.Children() {
			c.True(band.FrameRect().Bottom() <= bottom+0.5, "a band of page %d ends at %v, past the printable bottom of %v",
				i+1, band.FrameRect().Bottom(), bottom)
		}
	}
	ranges := drawnRowRanges(sheet.grid.pages, gurps.BlockTraitsKey)
	c.Equal(len(entity.Traits), ranges[len(ranges)-1][1], "every row must be drawn")
}

// TestPaginatedSheetShowsARowAddedAtTheEndOfAList verifies that a row added at the end of a list going on over several
// pages is drawn once the sheet is laid out again, although adding it changes the frame of no view, since each keeps
// the range of rows it was given. Filling the pages again only when a frame changes fails it.
func TestPaginatedSheetShowsARowAddedAtTheEndOfAList(t *testing.T) {
	c := check.New(t)
	sheet := newPaginatedTestSheetWithTraits(t, 150)
	c.True(len(drawnRowRanges(sheet.grid.pages, gurps.BlockTraitsKey)) > 1, "the traits must go on over several pages")
	entity := sheet.Entity()
	added := gurps.NewTrait(entity, nil, false)
	added.Name = "Trait 151"
	entity.Traits = append(entity.Traits, added)
	sheet.MarkModified(sheet)
	sheet.grid.ValidateLayout()
	ranges := drawnRowRanges(sheet.grid.pages, gurps.BlockTraitsKey)
	c.Equal(len(entity.Traits), ranges[len(ranges)-1][1], "the added row must be drawn")
}

// TestPaginatedSheetKeepsAnEmptyListBesideTheStartOfALongOne verifies that an empty list sharing a band with a list
// going on over several pages is shown once, on the page the band starts on, rather than handed out afresh for each
// page the band goes on to. Taking a list with no rows for one not yet placed fails it.
func TestPaginatedSheetKeepsAnEmptyListBesideTheStartOfALongOne(t *testing.T) {
	c := check.New(t)
	sheet := newPaginatedTestSheetWithTraits(t, 150)
	pages := sheet.grid.pages
	c.True(len(drawnRowRanges(pages, gurps.BlockTraitsKey)) > 1, "the traits must go on over several pages")
	c.Equal(0, sheet.Skills.RowCount(), "the skills list must be empty")
	c.True(unison.AncestorIs(sheet.Skills.Table, pages[0]), "the skills list must be on the first page")
	count := 0
	for _, page := range pages {
		count += len(blockPanels(page.AsPanel(), gurps.BlockSkillsKey))
	}
	c.Equal(1, count, "the skills list must be shown once")
}

// TestPaginatedSheetSetsOutAsManyPagesAcrossAsFit verifies that a sheet shown as pages sets them out in rows of as many
// as fit the width of the view at the sheet's scale, following a change of scale and a narrower view alike, without
// filling the pages again for either. Setting out a fixed number across, or filling the pages whenever they are laid
// out, fails it.
func TestPaginatedSheetSetsOutAsManyPagesAcrossAsFit(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet := openPaginatedSheet(t, screen, func(entity *gurps.Entity) []*gurps.Trait { return plainTraits(entity, 300) })
	var bands []*unison.Panel
	var fullWidth float32
	screen.Do(func() {
		bands = pageBands(sheet)
		fullWidth = sheet.scroll.FrameRect().Width
	})
	c.True(slices.ContainsFunc(bands, func(band *unison.Panel) bool { return bandKindOf(band) == rowBandPanel }),
		"a band must hold more than one block, so that filling the pages again would replace it")
	acrossFor := make(map[string]int)
	for _, one := range []struct {
		name   string
		scale  int
		narrow bool
	}{
		{name: "100%", scale: Scale100ItemID},
		{name: "50%", scale: Scale50ItemID},
		{name: "50% narrow", scale: Scale50ItemID, narrow: true},
	} {
		var across, want, count int
		var same bool
		screen.Do(func() {
			sheet.AsPanel().PerformCmd(nil, one.scale)
			r := sheet.scroll.FrameRect()
			r.Width = fullWidth
			if one.narrow {
				r.Width = 300
			}
			sheet.scroll.SetFrameRect(r)
			sheet.scroll.ValidateLayout()
			pages := sheet.grid.pages
			count = len(pages)
			for _, page := range pages {
				if page.FrameRect().Y == pages[0].FrameRect().Y {
					across++
				}
			}
			avail := sheet.scroll.ContentView().ContentRect(false).Width / sheet.grid.Scale().X
			width := pages[0].FrameRect().Width
			want = 1
			for want < count && float32(want+1)*width+float32(want)*pageGridGap <= avail {
				want++
			}
			same = slices.Equal(bands, pageBands(sheet))
		})
		c.Equal(want, across, "%s: as many pages across as fit", one.name)
		c.True(same, "%s: setting the pages out anew must not fill them again", one.name)
		c.True(count > across, "%s: the pages must take more than one row", one.name)
		acrossFor[one.name] = across
	}
	c.True(acrossFor["50%"] > acrossFor["100%"], "a smaller scale must fit more pages across")
	c.Equal(1, acrossFor["50% narrow"], "a view narrower than two pages must show one across")
}

// TestPagesFilledWhileTheGridIsSizedAreSetOut verifies that pages filled when the scroll panel asks for the grid's
// size, rather than when the grid is laid out, are set out all the same as the scroll panel is laid out, rather than
// left on top of one another at the grid's origin, where filling them puts them. Filling them from LayoutSizes without
// having the grid laid out afterwards fails it.
func TestPagesFilledWhileTheGridIsSizedAreSetOut(t *testing.T) {
	c := check.New(t)
	sheet := newPaginatedTestSheetWithTraits(t, 300)
	sheet.scroll.SetFrameRect(geom.NewRect(0, 0, 2000, 1000))
	sheet.scroll.ValidateLayout()
	c.True(len(sheet.grid.pages) > 2, "the character must take several pages")
	// Nothing today has the pages filled without marking the grid for layout, so a list is made to seem to have
	// changed its number of rows behind the grid's back.
	sheet.grid.rowCounts[gurps.BlockNotesKey] = 99
	sheet.scroll.MarkForLayoutAndRedraw()
	sheet.scroll.ValidateLayout()
	c.False(sheet.grid.needsFilling(), "the pages must have been filled")
	atOrigin := 0
	for _, page := range sheet.grid.pages {
		if page.FrameRect().Point == (geom.Point{}) {
			atOrigin++
		}
	}
	c.Equal(1, atOrigin, "only the first of the pages may be at the grid's origin")
}

// TestARowOnALaterPageTakesClicksAndTheKeyboard verifies that a row a list shows on a later page than its first is
// selected by a click, which puts the focus on the view of the list showing it, and that the keyboard's cell cursor
// then lands on that row. A view that took no part in its table's selection and focus fails it.
func TestARowOnALaterPageTakesClicksAndTheKeyboard(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet := openPaginatedSheet(t, screen, func(entity *gurps.Entity) []*gurps.Trait { return plainTraits(entity, 150) })
	table := sheet.Traits.Table
	var view *unison.Table[*Node[*gurps.Trait]]
	var row int
	screen.Do(func() {
		views := table.Views()
		view = views[len(views)-1]
		start, endBefore := view.CurrentDrawRowRange()
		row = (start + endBefore) / 2
	})
	c.True(view != table, "the traits must go on over several pages")
	screen.Click(cellScreenPoint(t, screen, wnd, view, row, gurps.TraitDescriptionColumn))
	var selected int
	var focused bool
	screen.Do(func() {
		selected = table.FirstSelectedRowIndex()
		focused = wnd.Focus() == view.AsPanel()
	})
	c.Equal(row, selected, "the click must select the row")
	c.True(focused, "the click must focus the view showing the row")
	screen.KeyPress(unison.KeyRight, mod.None)
	var lead, column int
	screen.Do(func() {
		lead = table.LeadRowIndex()
		column = table.LeadColumnIndex()
		focused = wnd.Focus() == view.AsPanel()
	})
	c.Equal(row, lead, "the cell cursor must be on the row")
	c.Equal(0, column, "the cell cursor must be on the first column")
	c.True(focused, "the view showing the row must keep the focus")
}

// TestDroppingARowOnALaterPageFocusesTheViewShowingIt verifies that a row dragged to another spot within a list's
// continuation on a later page leaves the keyboard focus on the view showing it once the drop is done, although the
// drop leaves the table with no lead row. Focusing the view showing the lead row puts the focus on the list's first
// page instead.
func TestDroppingARowOnALaterPageFocusesTheViewShowingIt(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet := openPaginatedSheet(t, screen, func(entity *gurps.Entity) []*gurps.Trait { return plainTraits(entity, 150) })
	table := sheet.Traits.Table
	var view *unison.Table[*Node[*gurps.Trait]]
	var row int
	var dragged *gurps.Trait
	screen.Do(func() {
		views := table.Views()
		view = views[len(views)-1]
		start, _ := view.CurrentDrawRowRange()
		row = start + 1
		dragged = table.RowFromIndex(row).Data()
	})
	c.True(view != table, "the traits must go on over several pages")
	var from, to geom.Point
	screen.Do(func() {
		col := view.ColumnIndexForID(gurps.TraitDescriptionColumn)
		view.ScrollRectIntoView(view.CellFrame(row, col).Union(view.CellFrame(row+4, col)))
		view.ValidateScrollRoot()
		from = screenPoint(wnd, view.PointToRoot(view.CellFrame(row, col).Center()))
		to = screenPoint(wnd, view.PointToRoot(view.CellFrame(row+4, col).Center()))
	})
	screen.Drag(from, to, 10)
	var selected, lead int
	var onLaterPage, focused bool
	screen.Do(func() {
		selected = table.FirstSelectedRowIndex()
		lead = table.LeadRowIndex()
		shownBy := table.ViewForRow(selected)
		onLaterPage = shownBy != table
		focused = wnd.Focus() == shownBy.AsPanel()
	})
	c.NotEqual(row, selected, "the row must have been moved")
	c.Equal(dragged, sheet.Entity().Traits[selected], "the row moved must be the one selected")
	c.Equal(-1, lead, "the drop must leave no lead row")
	c.True(onLaterPage, "the row must still be on a later page")
	c.True(focused, "the focus must be on the view showing the row")
}

// TestDisclosingARowOnALaterPageFillsThePagesAgain verifies that opening a container a list shows on a later page by
// its disclosure triangle fills the pages again with the rows it discloses, drawing each of them, and leaves the
// keyboard focus on the sheet. Filling the pages only when the sheet is rebuilt fails it.
func TestDisclosingARowOnALaterPageFillsThePagesAgain(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	var bag *gurps.Trait
	sheet := openPaginatedSheet(t, screen, func(entity *gurps.Entity) []*gurps.Trait {
		traits := plainTraits(entity, 150)
		bag = gurps.NewTrait(entity, nil, true)
		bag.Name = "Bag"
		bag.SetOpen(false)
		bag.Children = plainTraits(entity, 80)
		for _, child := range bag.Children {
			child.SetParent(bag)
		}
		return append(traits, bag)
	})
	table := sheet.Traits.Table
	var view *unison.Table[*Node[*gurps.Trait]]
	var row, pagesBefore int
	screen.Do(func() {
		row = table.LastRowIndex()
		view = table.ViewForRow(row)
		pagesBefore = len(sheet.grid.pages)
	})
	c.True(view != table, "the container must be on a later page than the list's first")
	var pt geom.Point
	screen.Do(func() {
		view.ScrollRowIntoView(row)
		view.ValidateScrollRoot()
		// The triangle sits in the hierarchy indent to the left of what the cell frame covers (see
		// unison.Table.CellFrame), centered in the row's first line.
		frame := view.CellFrame(row, view.ColumnIndexForID(view.HierarchyColumnID))
		indent := view.CurrentHierarchyIndent()
		size := min(indent, view.MinimumRowHeight) - 4
		center := geom.NewPoint(frame.X-indent-view.Padding.Left+2+size/2, frame.Y+view.MinimumRowHeight/2)
		pt = screenPoint(wnd, view.PointToRoot(center))
	})
	screen.Click(pt)
	var open, onSheet bool
	var rows, pagesAfter int
	var ranges [][2]int
	screen.Do(func() {
		open = bag.IsOpen()
		rows = table.LastRowIndex() + 1
		ranges = drawnRowRanges(sheet.grid.pages, gurps.BlockTraitsKey)
		pagesAfter = len(sheet.grid.pages)
		focus := wnd.Focus()
		onSheet = focus != nil && unison.AncestorIsOrSelf(focus, sheet)
	})
	c.True(open, "the click must open the container")
	c.Equal(231, rows, "the container's rows must be disclosed")
	c.Equal(rows, ranges[len(ranges)-1][1], "every disclosed row must be drawn")
	c.True(pagesAfter > pagesBefore, "the disclosed rows must take more pages")
	c.True(onSheet, "the focus must stay on the sheet")
}

// TestRevealingARowOnALaterPageScrollsToIt verifies that revealing a row inside a closed container a list shows on a
// later page, as choosing a search match does, scrolls the page holding it into view. Scrolling before the pages are
// filled again for the disclosed rows, or within the list's first view, fails it.
func TestRevealingARowOnALaterPageScrollsToIt(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	var target *gurps.Trait
	sheet := openPaginatedSheet(t, screen, func(entity *gurps.Entity) []*gurps.Trait {
		traits := plainTraits(entity, 150)
		bag := gurps.NewTrait(entity, nil, true)
		bag.Name = "Bag"
		bag.SetOpen(false)
		bag.Children = plainTraits(entity, 80)
		for _, child := range bag.Children {
			child.SetParent(bag)
		}
		target = bag.Children[len(bag.Children)-1]
		return append(traits, bag)
	})
	table := sheet.Traits.Table
	var before, after float32
	var index int
	var inView, laterPage bool
	screen.Do(func() {
		sheet.scroll.SetPosition(0, 0)
		_, before = sheet.scroll.Position()
		node := table.RootRows()[len(table.RootRows())-1].Children()[79]
		revealRow(table, node)
	})
	screen.Do(func() {
		_, after = sheet.scroll.Position()
		index = rowIndexForData(table, target)
		if index >= 0 {
			inView = rowInView(table, index)
			laterPage = table.ViewForRow(index) != table
		}
	})
	c.True(index >= 0, "the row must be disclosed")
	c.True(laterPage, "the row must be on a later page than the list's first")
	c.True(after > before, "the view must scroll down to the row")
	c.True(inView, "the row must be in view")
}

// TestDropOnAPaginatedSheetGoesToTheEndOfTheList verifies that an item dropped anywhere on a sheet whose list goes on
// over several pages is rerouted to the view of the list showing its last row, so that it lands at the end of the
// list. Rerouting it to the list's first view, which ends on an earlier page, would insert it there instead.
func TestDropOnAPaginatedSheetGoesToTheEndOfTheList(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet := openPaginatedSheet(t, screen, func(entity *gurps.Entity) []*gurps.Trait { return plainTraits(entity, 150) })
	table := sheet.Traits.Table
	var target, last *unison.Panel
	screen.Do(func() {
		target = sheet.keyToPanel(traitDragKey)
		last = table.ViewForRow(table.LastRowIndex()).AsPanel()
	})
	c.True(last != table.AsPanel(), "the list's last row must be on a later page than its first")
	c.True(target == last, "the drop must go to the view showing the list's last row")
}

// TestEditingTheLayoutShowsAPaginatedSheetContinuously verifies that editing the block layout of a sheet shown as pages
// shows it as one continuous page, which the editor's overlay covers, for as long as the editing lasts, and that the
// pages come back once it is over, all without changing the user's choice of pages. Leaving the pages up while the
// layout is edited, or not putting them back after, fails it.
func TestEditingTheLayoutShowsAPaginatedSheetContinuously(t *testing.T) {
	c := check.New(t)
	sheet := newPaginatedTestSheetWithTraits(t, 150)
	before := len(sheet.grid.pages)
	c.True(before > 1, "the character must take several pages")
	sheet.toggleLayoutEditing()
	c.True(sheet.paginated, "editing the layout must leave the choice of pages alone")
	c.Equal(1, len(sheet.grid.pages), "the layout must be edited on one continuous page")
	c.False(sheet.page.Force, "the continuous page must be as tall as its content")
	c.Equal(1, len(sheet.Traits.Table.Views()), "no list may go on past the continuous page")
	c.Equal(sheet.page.FrameRect(), sheet.layoutEditor.overlay.FrameRect(), "the overlay must cover the page")
	sheet.toggleLayoutEditing()
	c.True(sheet.paginated, "leaving the editor must leave the choice of pages alone")
	c.Equal(before, len(sheet.grid.pages), "leaving the editor must show the pages again")
	c.True(sheet.page.Force, "a page must be the size of the paper again")
}

// TestTogglingPagesKeepsTheFocus verifies that switching a sheet between pages and one continuous page leaves the
// keyboard focus where it was, whether on a field or on a list going on over several pages, where it stays on the view
// showing the selected row, and that the continuous page is the only one, with no list going on past it. Rebuilding
// without putting the focus back, putting it on another view of the list, or leaving pages or continuations behind,
// fails it.
func TestTogglingPagesKeepsTheFocus(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet := openPaginatedSheet(t, screen, func(entity *gurps.Entity) []*gurps.Trait { return plainTraits(entity, 150) })
	field := func() *unison.Panel { return sheet.AsPanel().FindRefKey(identityPanelNameFieldRefKey) }
	for _, one := range []struct {
		name   string
		focus  func() *unison.Panel
		holder func() *unison.Panel
	}{
		{name: "field", focus: field, holder: field},
		{
			name: "list",
			focus: func() *unison.Panel {
				table := sheet.Traits.Table
				views := table.Views()
				view := views[len(views)-1]
				start, _ := view.CurrentDrawRowRange()
				table.SelectByIndex(start)
				return view.AsPanel()
			},
			holder: func() *unison.Panel {
				table := sheet.Traits.Table
				return table.ViewForRow(table.FirstSelectedRowIndex()).AsPanel()
			},
		},
	} {
		var ref *FocusRef
		screen.Do(func() {
			if !sheet.paginated {
				sheet.togglePaginated()
			}
			if target := one.focus(); target != nil {
				target.RequestFocus()
			}
			ref = sheet.targetMgr.CurrentFocusRef()
		})
		if ref == nil || ref.Key == "" {
			t.Fatalf("%s: the focus must be on the sheet, on a panel with a reference key", one.name)
		}
		for _, paginated := range []bool{false, true} {
			var after *FocusRef
			var pages, views int
			var onHolder bool
			screen.Do(func() {
				sheet.togglePaginated()
				after = sheet.targetMgr.CurrentFocusRef()
				pages = len(sheet.grid.pages)
				views = len(sheet.Traits.Table.Views())
				onHolder = wnd.Focus() == one.holder()
			})
			c.Equal(paginated, sheet.paginated, "%s: the mode must have switched", one.name)
			c.NotNil(after, "%s: the focus must stay on the sheet (paginated %v)", one.name, paginated)
			if after != nil {
				c.Equal(ref.Key, after.Key, "%s: the focus must stay where it was (paginated %v)", one.name, paginated)
			}
			c.True(onHolder, "%s: the focus must be on the panel showing what it was on (paginated %v)", one.name,
				paginated)
			if paginated {
				c.True(pages > 1, "%s: the pages must come back", one.name)
			} else {
				c.Equal(1, pages, "%s: the sheet must show one continuous page", one.name)
				c.Equal(1, views, "%s: no list may go on past the continuous page", one.name)
			}
		}
	}
}

// TestFocusReturningToAListGoesToTheViewShowingTheRow verifies that the focus held by a view of a list on a page later
// than the list's first goes back to the view showing the row the person is on once the sheet is rebuilt, or once an
// editor opened on the row is closed, either of which finds the focus again by a reference key the views share. Giving
// it to the first view found, the primary, fails it.
func TestFocusReturningToAListGoesToTheViewShowingTheRow(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet := openPaginatedSheet(t, screen, func(entity *gurps.Entity) []*gurps.Trait { return plainTraits(entity, 150) })
	table := sheet.Traits.Table
	for _, one := range []struct {
		name   string
		action func(view *unison.Table[*Node[*gurps.Trait]])
	}{
		{name: "the sheet rebuilt", action: func(_ *unison.Table[*Node[*gurps.Trait]]) {
			screen.Do(func() { sheet.Rebuild(true) })
		}},
		{name: "the row's editor closed", action: func(view *unison.Table[*Node[*gurps.Trait]]) {
			screen.Do(func() { view.PerformCmd(nil, OpenEditorItemID) })
			uxtest.CloseEditorWithoutPrompt(t, screen,
				uxtest.SoleEditor[*editor[*gurps.Trait, *gurps.TraitEditData]](t, screen, func(d unison.Dockable) bool {
					_, ok := d.AsPanel().Self.(*editor[*gurps.Trait, *gurps.TraitEditData])
					return ok
				}))
		}},
	} {
		var view *unison.Table[*Node[*gurps.Trait]]
		var row int
		screen.Do(func() {
			views := table.Views()
			view = views[len(views)-1]
			start, endBefore := view.CurrentDrawRowRange()
			row = (start + endBefore) / 2
		})
		c.True(view != table, "%s: the traits must go on over several pages", one.name)
		screen.Click(cellScreenPoint(t, screen, wnd, view, row, gurps.TraitDescriptionColumn))
		one.action(view)
		var focus, want *unison.Panel
		screen.Do(func() {
			focus = wnd.Focus()
			want = table.ViewForRow(row).AsPanel()
		})
		c.True(focus == want, "%s: the focus must go back to the view showing the row", one.name)
		c.True(focus != table.AsPanel(), "%s: which is not the primary view", one.name)
	}
}

// TestLeavingTheLayoutEditorKeepsThePlaceOnTheSheet verifies that a sheet shown as pages and scrolled well down is left
// about as far down once its layout has been edited, rather than scrolled to wherever the editor hands the focus back.
// Bringing the focus into view as the pages come back fails it.
func TestLeavingTheLayoutEditorKeepsThePlaceOnTheSheet(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet := openPaginatedSheet(t, screen, func(entity *gurps.Entity) []*gurps.Trait { return plainTraits(entity, 300) })
	var fraction float32
	screen.Do(func() {
		sheet.scroll.SetPosition(0, sheet.scroll.Bar(false).MaxValue())
		sheet.toggleLayoutEditing()
		sheet.toggleLayoutEditing()
		fraction = sheet.scrollFraction()
	})
	c.True(fraction > 0.5, "the sheet must still be scrolled well down, not %v of the way", fraction)
}
