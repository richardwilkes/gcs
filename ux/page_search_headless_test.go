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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// pageSearchFixture is a character sheet open in a headless workspace, with the traits the test asked for and the focus
// in the sheet's content rather than its search field.
type pageSearchFixture struct {
	screen *unison.HeadlessScreen
	wnd    *unison.Window
	sheet  *Sheet
	search *pageSearch
}

// startPageSearch opens a new character sheet in a headless workspace and gives it the traits seed builds for its
// entity. The sheet is left unmodified, so that nothing prompts to save it when the session ends.
func startPageSearch(t *testing.T, c check.Checker, seed func(entity *gurps.Entity) []*gurps.Trait) *pageSearchFixture {
	t.Helper()
	f := &pageSearchFixture{}
	f.screen, f.wnd = uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, f.screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	f.sheet = sheet
	f.search = sheet.search
	f.screen.Do(func() {
		sheet.Entity().Traits = seed(sheet.Entity())
		sheet.Rebuild(true)
		sheet.markUnmodified()
	})
	return f
}

// offered returns the window showing the matches, or nil while none shows.
func (f *pageSearchFixture) offered() *unison.Window {
	return f.screen.SuggestionsWindow(f.wnd)
}

// rows returns how many matches are on offer, showing or not.
func (f *pageSearchFixture) rows() int {
	var count int
	f.screen.Do(func() { count = f.search.suggestions.List().Count() })
	return count
}

// highlighted returns the index of the highlighted match, or -1.
func (f *pageSearchFixture) highlighted() int {
	index := -2
	f.screen.Do(func() { index = f.search.suggestions.Selected() })
	return index
}

// selectedTraits returns the traits selected in the traits table.
func (f *pageSearchFixture) selectedTraits() []*gurps.Trait {
	var traits []*gurps.Trait
	f.screen.Do(func() {
		for _, row := range f.sheet.Traits.Table.SelectedRows(false) {
			traits = append(traits, row.Data())
		}
	})
	return traits
}

// fieldFocused returns true if the search field holds the keyboard focus.
func (f *pageSearchFixture) fieldFocused() bool {
	var focused bool
	f.screen.Do(func() { focused = f.wnd.Focus() == f.search.field.AsPanel() })
	return focused
}

// countWalks has the search count the walks of the lists it makes from now on.
func (f *pageSearchFixture) countWalks() *int {
	walks := 0
	f.screen.Do(func() {
		findMatches := f.search.findMatches
		f.search.findMatches = func(text string) []*pageSearchMatch {
			walks++
			return findMatches(text)
		}
	})
	return &walks
}

// Typing in the search field offers the matching rows. Choosing one, with RETURN or a click, discloses it, selects it
// alone and leaves the focus in the field with its text intact. RETURN with the matches hidden offers them again, and
// with none highlighted chooses the first; Escape hides them; text nothing matches offers nothing, as does clearing
// the field.
func TestPageSearchChoosesARow(t *testing.T) {
	c := check.New(t)
	var container, trait, other *gurps.Trait
	f := startPageSearch(t, c, func(entity *gurps.Entity) []*gurps.Trait {
		container = gurps.NewTrait(entity, nil, true)
		container.Name = "Container"
		container.SetOpen(false)
		trait = gurps.NewTrait(entity, container, false)
		trait.Name = "Findable"
		container.Children = []*gurps.Trait{trait}
		other = gurps.NewTrait(entity, nil, false)
		other.Name = "Other"
		return []*gurps.Trait{container, other}
	})
	c.Nil(f.offered(), "nothing is on offer before anything is typed")

	f.screen.Click(f.screen.PanelCenter(f.search.field))
	f.screen.Type("find")
	c.NotNil(f.offered(), "typing offers the matches")
	c.Equal(1, f.rows(), "one row matches")
	c.Equal(-1, f.highlighted(), "none is highlighted to start with")
	c.Equal(0, len(f.selectedTraits()), "typing selects nothing")

	f.screen.KeyPress(unison.KeyDown, mod.None)
	c.Equal(0, f.highlighted(), "Down highlights the first row")
	uxtest.CaptureScreen(t, c, f.screen, "page_search")
	f.screen.KeyPress(unison.KeyReturn, mod.None)
	c.Nil(f.offered(), "choosing hides the matches")
	selected := f.selectedTraits()
	c.Equal(1, len(selected), "the chosen row alone is selected")
	c.True(len(selected) == 1 && selected[0] == trait, "and it is the trait")
	var open bool
	var text string
	f.screen.Do(func() {
		open = container.IsOpen()
		text = f.search.field.Text()
	})
	c.True(open, "the container holding the trait was opened")
	c.True(f.fieldFocused(), "the focus stays in the field")
	c.Equal("find", text, "and the text is kept")

	// RETURN offers the matches again, and with none highlighted chooses the first.
	f.screen.Do(func() { f.sheet.Traits.Table.ClearSelection() })
	f.screen.KeyPress(unison.KeyReturn, mod.None)
	c.NotNil(f.offered(), "RETURN with the matches hidden offers them again")
	c.Equal(-1, f.highlighted(), "with none highlighted")
	f.screen.KeyPress(unison.KeyReturn, mod.None)
	c.Nil(f.offered(), "RETURN with none highlighted chooses the first")
	selected = f.selectedTraits()
	c.True(len(selected) == 1 && selected[0] == trait, "which is the trait")

	// Escape hides them, and nothing is offered for text nothing matches.
	f.screen.KeyPress(unison.KeyReturn, mod.None)
	c.NotNil(f.offered())
	f.screen.KeyPress(unison.KeyEscape, mod.None)
	c.Nil(f.offered(), "Escape hides the matches")
	f.screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
	f.screen.Type("zzz")
	c.Nil(f.offered(), "nothing is offered for text nothing matches")
	c.Equal(0, f.rows())

	// Clicking a row chooses it, and the clear button takes the matches away with the text.
	f.screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
	f.screen.Type("other")
	popup := f.offered()
	c.NotNil(popup, "the other trait is offered")
	var list *unison.Panel
	f.screen.Do(func() { list = f.search.suggestions.List().AsPanel() })
	f.screen.Click(f.screen.PanelCenter(list))
	c.Nil(f.offered(), "a click on the row chooses it")
	selected = f.selectedTraits()
	c.True(len(selected) == 1 && selected[0] == other, "and selects the other trait")
	c.True(f.fieldFocused(), "with the focus still in the field")
	f.screen.KeyPress(unison.KeyReturn, mod.None)
	c.NotNil(f.offered())
	clearButton, ok := uxtest.FirstPanelOfType[*unison.Button](f.search.field.AsPanel())
	if !ok {
		t.Fatal("the search field must have a clear button")
	}
	f.screen.Click(f.screen.PanelCenter(clearButton))
	f.screen.Do(func() { text = f.search.field.Text() })
	c.Equal("", text, "the clear button empties the field")
	c.Nil(f.offered(), "and takes the matches away")
	c.Equal(0, f.rows())
}

// An update of the sheet finds the matches again while they are showing, keeping the highlight on the same row, hides
// them when none is left, and leaves the lists alone while they are hidden.
func TestPageSearchRefreshWhileShowing(t *testing.T) {
	c := check.New(t)
	f := startPageSearch(t, c, func(entity *gurps.Entity) []*gurps.Trait {
		first := gurps.NewTrait(entity, nil, false)
		first.Name = "Findable"
		second := gurps.NewTrait(entity, nil, false)
		second.Name = "Findable Too"
		return []*gurps.Trait{first, second}
	})
	f.screen.Click(f.screen.PanelCenter(f.search.field))
	f.screen.Type("find")
	c.NotNil(f.offered())
	c.Equal(2, f.rows())
	f.screen.KeyPress(unison.KeyDown, mod.None)
	f.screen.KeyPress(unison.KeyDown, mod.None)
	c.Equal(1, f.highlighted(), "the second row is highlighted")

	walks := f.countWalks()
	f.screen.Do(func() { f.sheet.MarkModified(nil) })
	c.Equal(1, *walks, "an update walks the lists while the matches are showing")
	c.NotNil(f.offered(), "the matches stay on offer")
	c.Equal(2, f.rows())
	c.Equal(1, f.highlighted(), "with the highlight kept on the same row")

	f.screen.Do(func() {
		entity := f.sheet.Entity()
		entity.Traits = entity.Traits[1:]
		f.sheet.MarkModified(nil)
	})
	c.Equal(2, *walks)
	c.NotNil(f.offered(), "the matches stay on offer while any remains")
	c.Equal(1, f.rows(), "the row that was removed is gone")
	c.Equal(0, f.highlighted(), "and the highlight follows its row")

	f.screen.Do(func() {
		f.sheet.Entity().Traits = nil
		f.sheet.MarkModified(nil)
	})
	c.Equal(3, *walks)
	c.Nil(f.offered(), "the matches are hidden once none is left")

	f.screen.Do(func() {
		f.sheet.Entity().Traits = []*gurps.Trait{newFindableTrait(f.sheet.Entity(), "Findable Again")}
		f.sheet.MarkModified(nil)
	})
	c.Equal(3, *walks, "an update with the matches hidden leaves the lists alone")
	c.Nil(f.offered(), "and offers nothing")
	f.screen.Do(func() { f.sheet.markUnmodified() })
}

// Text put in the field while it lacks the focus -- by the jump-to-search command's target, or by the Library Explorer
// handing over the text a file matched on -- is searched for once the field has the focus.
func TestPageSearchShowsMatchesOnFocus(t *testing.T) {
	c := check.New(t)
	f := startPageSearch(t, c, func(entity *gurps.Entity) []*gurps.Trait {
		return []*gurps.Trait{newFindableTrait(entity, "Findable")}
	})
	c.False(f.fieldFocused(), "the focus starts in the sheet's content")
	f.screen.Do(func() { f.search.field.SetText("find") })
	c.Nil(f.offered(), "text set while the field lacks the focus offers nothing")
	c.Equal(0, f.rows())

	f.screen.Do(func() { f.sheet.AsPanel().PerformCmd(nil, JumpToSearchFilterItemID) })
	c.True(f.fieldFocused(), "the jump-to-search command focuses the field")
	c.NotNil(f.offered(), "which offers the matches of the text it holds")
	c.Equal(1, f.rows())

	f.screen.KeyPress(unison.KeyEscape, mod.None)
	c.Nil(f.offered())
	f.screen.Do(func() { f.sheet.Traits.Table.RequestFocus() })
	c.False(f.fieldFocused())
	f.screen.Do(func() { f.sheet.receiveSearchText("able") })
	c.True(f.fieldFocused(), "handing over search text focuses the field")
	c.NotNil(f.offered(), "and offers the matches")
	var text string
	f.screen.Do(func() { text = f.search.field.Text() })
	c.Equal("able", text)
}
