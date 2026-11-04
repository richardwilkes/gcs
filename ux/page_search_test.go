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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// matchLabels returns the matches as "name | title" strings.
func matchLabels(matches []*searchMatch) []string {
	labels := make([]string, len(matches))
	for i, m := range matches {
		labels[i] = m.name + " | " + m.title
	}
	return labels
}

// newFindableTrait returns a trait of the given name owned by the entity.
func newFindableTrait(entity *gurps.Entity, name string) *gurps.Trait {
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = name
	return trait
}

// The matches are labeled with what the row's hierarchy column shows -- not column 0, which on a sheet is the equipped
// box for carried equipment -- cut to its first line, since a note's String() is its whole text, and carry the title of
// the list they were found in, in block order. A template's and a loot sheet's equipment lists are headed plainly
// "Equipment", as their blocks are.
func TestPageSearchMatchesCarryTheirListTitles(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.Traits = []*gurps.Trait{newFindableTrait(entity, "Findable Trait")}
	skill := gurps.NewSkill(entity, nil, false)
	skill.Name = "Findable Skill"
	entity.Skills = []*gurps.Skill{skill}
	sword := gurps.NewEquipment(entity, nil, false)
	sword.Name = "Findable Sword"
	entity.CarriedEquipment = []*gurps.Equipment{sword}
	note := gurps.NewNote(entity, nil, false)
	note.MarkDown = "# Findable note\nwith a second line"
	entity.Notes = []*gurps.Note{note}
	sheet.Rebuild(true)
	c.Equal([]string{
		"Findable Trait | Traits",
		"Findable Skill | Skills",
		"Findable Sword | Carried Equipment",
		"Findable note | Notes",
	}, matchLabels(sheet.search.findMatches("findable")))

	data := gurps.NewTemplate()
	gear := gurps.NewEquipment(nil, nil, false)
	gear.Name = "Findable Gear"
	data.Equipment = []*gurps.Equipment{gear}
	template := newTestTemplateDockable("Search", data)
	c.Equal([]string{"Findable Gear | Equipment"}, matchLabels(template.search.findMatches("findable")))

	loot := newTestLootSheet(t)
	lootGear := gurps.NewEquipment(loot.loot, nil, false)
	lootGear.Name = "Findable Loot"
	loot.loot.Equipment = []*gurps.Equipment{lootGear}
	lootNote := gurps.NewNote(loot.loot, nil, false)
	lootNote.MarkDown = "Findable loot note"
	loot.loot.Notes = []*gurps.Note{lootNote}
	loot.Rebuild(true)
	c.Equal([]string{"Findable Loot | Equipment", "Findable loot note | Notes"},
		matchLabels(loot.search.findMatches("findable")))
}

// A long label is cut short, and a weapon, whose description column shows its owner, adds its usage so that one
// weapon's attacks can be told apart, even when its owner's name is cut short.
func TestPageSearchLabels(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	blade := newFindableTrait(entity, "Findable Blade")
	swing := gurps.NewWeapon(blade, true)
	swing.Usage = "Swung"
	thrust := gurps.NewWeapon(blade, true)
	thrust.Usage = "Thrust"
	blade.Weapons = []*gurps.Weapon{swing, thrust}
	longName := "Findable " + strings.Repeat("y", maxPageSearchLabelLength)
	cutName := "Findable " + strings.Repeat("y", maxPageSearchLabelLength-len("Findable ")) + "…"
	longBlade := newFindableTrait(entity, longName)
	longSwing := gurps.NewWeapon(longBlade, true)
	longSwing.Usage = "Swung"
	longThrust := gurps.NewWeapon(longBlade, true)
	longThrust.Usage = "Thrust"
	longBlade.Weapons = []*gurps.Weapon{longSwing, longThrust}
	entity.Traits = []*gurps.Trait{blade, longBlade}
	note := gurps.NewNote(entity, nil, false)
	note.MarkDown = "Findable " + strings.Repeat("x", maxPageSearchLabelLength)
	entity.Notes = []*gurps.Note{note}
	sheet.Rebuild(true)
	c.Equal([]string{
		"Findable Blade (Swung) | Melee Attacks",
		"Findable Blade (Thrust) | Melee Attacks",
		cutName + " (Swung) | Melee Attacks",
		cutName + " (Thrust) | Melee Attacks",
		"Findable Blade | Traits",
		cutName + " | Traits",
		"Findable " + strings.Repeat("x", maxPageSearchLabelLength-len("Findable ")) + "… | Notes",
	}, matchLabels(sheet.search.findMatches("findable")))
}

// A row of the dropdown shows the match's name followed by the title of its list, and is named for both to an
// assistive technology. A match with an icon, as the Library Explorer's have, shows it, and one without a title, as a
// root row of the Library Explorer is, is named by its name alone.
func TestSearchMatchCellShowsTheTitleAndIcon(t *testing.T) {
	c := check.New(t)
	list := unison.NewList[*searchMatch]()
	match := &searchMatch{name: "Findable", title: "Traits", show: func() {}}
	label, ok := (&searchMatchCellFactory{}).CreateCell(list, match, 0, unison.Black, unison.White, false, false).(*unison.Label)
	if !ok {
		t.Fatal("the cell must be a label")
	}
	c.Equal("Findable  Traits", label.String())
	c.Equal("Findable, Traits", label.Accessibility.Name)
	c.Nil(label.Drawable, "a match without an icon shows none")

	icon := &unison.DrawableSVG{SVG: unison.DocumentSVG}
	match = &searchMatch{name: "Library", icon: icon, show: func() {}}
	label, ok = (&searchMatchCellFactory{}).CreateCell(list, match, 0, unison.Black, unison.White, false, false).(*unison.Label)
	if !ok {
		t.Fatal("the cell must be a label")
	}
	c.Equal("Library", label.String())
	c.Equal("Library", label.Accessibility.Name)
	c.True(label.Drawable == icon, "a match with an icon shows it")
}

// A row that only carries a note, as the one saying the matches were cut short is, is set apart from the matches: it
// is drawn dimmed in the italic secondary font, under a line, and shows no icon.
func TestSearchMatchCellSetsANoteApart(t *testing.T) {
	c := check.New(t)
	list := unison.NewList[*searchMatch]()
	factory := &searchMatchCellFactory{}
	match, ok := factory.CreateCell(list, &searchMatch{name: "Findable", show: func() {}}, 0, unison.Black, unison.White,
		false, false).(*unison.Label)
	if !ok {
		t.Fatal("the cell must be a label")
	}
	note, ok := factory.CreateCell(list, &searchMatch{name: "Narrow the search"}, 0, unison.Black, unison.White, false,
		false).(*unison.Label)
	if !ok {
		t.Fatal("the cell must be a label")
	}
	c.Equal("Narrow the search", note.String())
	c.Equal("Narrow the search", note.Accessibility.Name)
	c.Nil(note.Drawable)
	want := noteFont()
	c.True(note.Font.Face() == want.Face() && note.Font.Size() == want.Size(), "a note is drawn in the note font")
	var isDimmed bool
	note.Text.AdjustDecorations(func(d *unison.TextDecoration) { _, isDimmed = d.OnBackgroundInk.(*unison.ColorFilteredInk) })
	c.True(isDimmed, "a note is dimmed")
	c.Equal(match.Border().Insets().Top+1, note.Border().Insets().Top, "a note sits under a line")
	c.Equal(match.Border().Insets().Bottom, note.Border().Insets().Bottom)
}

// The search field switches to the focus border while it holds the focus, as other fields do, and still searches for
// the text it holds when it gains the focus.
func TestSearchFieldShowsTheFocusBorder(t *testing.T) {
	c := check.New(t)
	walks := 0
	s := newSearchDropdown(func(string) []*searchMatch {
		walks++
		return nil
	}, func(*searchMatch) {})
	unfocused := s.field.Border()
	s.field.SetText("findable")
	s.field.GainedFocusCallback()
	c.True(s.field.Border() != unfocused, "gaining the focus shows the focus border")
	c.Equal(1, walks, "and searches for the text")
	s.field.LostFocusCallback()
	c.True(s.field.Border() == unfocused, "losing the focus puts the unfocused border back")
}

// The lists are only walked while the matches can be seen: not for an empty search, not for text put in the field while
// it lacks the focus, which can't show them, and not by the refresh every update performs while they are hidden. The
// walk has every row of every list derive its children and the data of its cells afresh, which was most of what an
// update of a sheet with large lists cost while nobody was searching. Nothing is selected until a match is chosen.
func TestPageSearchDoesNotWalkWhileHidden(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.Traits = []*gurps.Trait{newFindableTrait(entity, "Findable")}
	sheet.Rebuild(true)
	search := sheet.search
	walks := 0
	findMatches := search.findMatches
	search.findMatches = func(text string) []*searchMatch {
		walks++
		return findMatches(text)
	}

	search.Refresh()
	c.Equal(0, walks, "an empty search must not walk the lists")

	search.field.SetText("findable")
	c.Equal(0, walks, "text set while the field lacks the focus must not walk the lists")
	c.Equal(0, search.suggestions.List().Count(), "nothing is on offer")
	c.False(search.suggestions.Showing())

	search.Refresh()
	c.Equal(0, walks, "a refresh while the matches are hidden must not walk the lists")
	c.Equal(0, sheet.Traits.Table.SelectionCount(), "nothing is selected until a match is chosen")

	c.Equal(1, len(findMatches("findable")), "the walk itself finds the trait")
	c.Equal(0, len(findMatches("")), "but nothing for an empty search")
}

// The walk stops once it has one more match than the dropdown offers, so that the dropdown can tell the matches were
// cut short and say so in a row of its own, which chooses nothing; with no more than it offers, it says nothing.
func TestPageSearchCapsTheMatches(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	for i := range maxSearchMatches {
		entity.Traits = append(entity.Traits, newFindableTrait(entity, fmt.Sprintf("Findable %d", i)))
	}
	sheet.Rebuild(true)
	found := sheet.search.findMatches("findable")
	c.Equal(maxSearchMatches, len(found))
	rows := searchMatchRows(found)
	c.Equal(maxSearchMatches, len(rows), "as many matches as the dropdown offers are all shown")
	c.NotNil(rows[len(rows)-1].show, "with no notice")

	for i := range 50 {
		entity.Traits = append(entity.Traits, newFindableTrait(entity, fmt.Sprintf("Findable %d", maxSearchMatches+i)))
	}
	sheet.Rebuild(true)
	found = sheet.search.findMatches("findable")
	c.Equal(maxSearchMatches+1, len(found), "the walk stops one past what the dropdown offers")
	rows = searchMatchRows(found)
	c.Equal(maxSearchMatches+1, len(rows))
	c.Equal(found[:maxSearchMatches], rows[:maxSearchMatches], "the first are offered")
	c.Nil(rows[maxSearchMatches].show, "followed by a notice that chooses nothing")
	c.Contains(rows[maxSearchMatches].name, "Narrow the search")
}

// pageSearchFixture is a character sheet open in a headless workspace, with the traits the test asked for and the focus
// in the sheet's content rather than its search field.
type pageSearchFixture struct {
	screen *unison.HeadlessScreen
	wnd    *unison.Window
	sheet  *Sheet
	search *searchDropdown
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
		f.search.findMatches = func(text string) []*searchMatch {
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
// handing over the text a file matched on -- is searched for once the field has the focus. Text handed over to a field
// that already holds it and the focus, with the matches hidden, offers them again.
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

	f.screen.KeyPress(unison.KeyEscape, mod.None)
	c.Nil(f.offered())
	f.screen.Do(func() { f.sheet.receiveSearchText("able") })
	c.NotNil(f.offered(), "handing over the text the focused field holds offers the matches again")
}

// Choosing a match inside a closed container scrolls the row into view even when disclosing the container grows the
// page: the scroll panel's range is only as large as the page was at its last layout, so the page has to be laid out
// again before the scroll is asked for, or it stops short of a row beneath the old bottom of the page.
func TestPageSearchScrollsDisclosedRowIntoView(t *testing.T) {
	c := check.New(t)
	f := startPageSearch(t, c, func(_ *gurps.Entity) []*gurps.Trait { return nil })
	var last *gurps.Note
	f.screen.Do(func() {
		entity := f.sheet.Entity()
		container := gurps.NewNote(entity, nil, true)
		container.MarkDown = "Container"
		container.SetOpen(false)
		for i := range 40 {
			child := gurps.NewNote(entity, container, false)
			child.MarkDown = fmt.Sprintf("Child %d", i)
			container.Children = append(container.Children, child)
		}
		last = container.Children[len(container.Children)-1]
		last.MarkDown = "Findable"
		entity.Notes = []*gurps.Note{container}
		f.sheet.Rebuild(true)
		f.sheet.markUnmodified()
	})

	f.screen.Click(f.screen.PanelCenter(f.search.field))
	f.screen.Type("findable")
	c.Equal(1, f.rows(), "the note is offered")
	f.screen.KeyPress(unison.KeyReturn, mod.None)
	index := -1
	var rowRect, viewRect geom.Rect
	f.screen.Do(func() {
		table := f.sheet.Notes.Table
		index = rowIndexForData(table, last)
		if index >= 0 {
			view := table.ViewForRow(index)
			rowRect = view.RectToRoot(view.RowFrame(index))
		}
		view := f.sheet.scroll.ContentView()
		viewRect = view.RectToRoot(view.ContentRect(false))
	})
	c.True(index >= 0, "the chosen note is shown")
	// The scroll and the conversion to root coordinates each scale the row through the page view's scale, so the two can
	// differ by a rounding error; half a pixel covers that without letting a row that is really cut off through.
	viewRect = viewRect.Inset(geom.NewUniformInsets(-0.5))
	c.True(viewRect.Contains(rowRect), "the chosen row %v is scrolled into the view %v", rowRect, viewRect)
}
