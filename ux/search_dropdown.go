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

	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// maxSearchMatches is the most rows a search dropdown offers; the walk that finds them stops once it has that many,
// since the list that shows them builds a cell for every row whenever it is sized.
const maxSearchMatches = 100

// searchMatch is a row a search found: its label, a title shown dimmed after it (the list or folder it is in), a key
// that names the same row across a rebuild of the table holding it, an optional icon drawn ahead of the label, and a
// function that discloses, selects and scrolls to it.
type searchMatch struct {
	show  func()
	icon  unison.Drawable
	name  string
	title string
	key   string
}

func (m *searchMatch) String() string {
	return m.name
}

// searchDropdown is a search field whose matches are offered in a dropdown as the user types; choosing one reveals it.
// The rows are only looked for while the dropdown can be seen, since the walk that finds them can be costly: on a sheet
// it has every row of a page list derive its children and the data of its cells afresh as it is read (see rowReuse),
// and in the Library Explorer it reads the files the deep search has not cached yet.
type searchDropdown struct {
	field       *unison.Field
	suggestions *unison.FieldSuggestions[*searchMatch]
	// findMatches returns the rows matching the lower-cased text. A field so that tests can count the walks.
	findMatches func(text string) []*searchMatch
	// choose reveals the chosen match.
	choose func(match *searchMatch)
}

// newSearchDropdown creates the field and its dropdown. Nothing is added to a parent and no command handler is
// installed; see installJumpToField.
func newSearchDropdown(findMatches func(text string) []*searchMatch, choose func(match *searchMatch)) *searchDropdown {
	s := &searchDropdown{findMatches: findMatches, choose: choose}
	searchText := i18n.Text("Search")
	s.field = NewSearchField(searchText, s.textModified)
	// A toolbar moves the field to a row of its own rather than squeeze it narrower than this much text; see
	// flowFillLayout.
	s.field.SetMinimumTextWidthUsing(strings.Repeat("0", 10))
	s.field.Tooltip = newWrappedTooltipWithSecondaryText(searchText,
		i18n.Text("Lists the matching rows as you type\nPress RETURN to select the highlighted row"))
	s.suggestions = unison.NewFieldSuggestions[*searchMatch](s.field)
	s.suggestions.List().Factory = &searchMatchCellFactory{}
	s.suggestions.ChosenCallback = func(_ int, match *searchMatch) { s.choose(match) }
	s.field.KeyDownCallback = s.keyDown
	// Text put in the field while it lacked the focus, by the Library Explorer or the user, is searched for once it
	// has it, which is also when the rows can first be shown.
	s.field.GainedFocusCallback = func() {
		s.field.DefaultFocusGained()
		s.showMatches()
	}
	return s
}

// installJumpToField makes the jump-to-search command on the given panel move the focus to the field.
func installJumpToField(panel *unison.Panel, field *unison.Field) {
	panel.InstallCmdHandlers(JumpToSearchFilterItemID,
		func(any) bool { return !field.Focused() },
		func(any) { field.RequestFocus() })
}

// showMatches finds the rows matching the field's text and offers them, returning true if there were any. Nothing
// matches an empty search, so nothing is walked for one.
func (s *searchDropdown) showMatches() bool {
	text := strings.ToLower(s.field.Text())
	if text == "" {
		s.suggestions.SetRows(nil)
		return false
	}
	found := s.findMatches(text)
	s.suggestions.SetRows(found)
	return len(found) != 0
}

// textModified offers the rows matching the new text. They are only looked for while the field has the focus, since
// they can only be shown then (see unison.FieldSuggestions.Show); see newSearchDropdown for text set at other times.
// The table selections are left alone: a row is only selected by choosing it.
func (s *searchDropdown) textModified(_, after *unison.FieldState) {
	if after.Text == "" || !s.field.Focused() {
		s.suggestions.SetRows(nil)
		return
	}
	s.suggestions.SetRows(s.findMatches(strings.ToLower(after.Text)))
}

// keyDown gives the field's keys their meaning for the search. While the rows are hidden, RETURN or a bare Down offers
// them again, Down with the first highlighted; they are looked for afresh, since they may have changed. While they
// show with none highlighted, RETURN chooses the first, where it would otherwise merely hide them. Everything else is
// left to the field, whose default handling gives the rows first crack at the keys that drive them.
func (s *searchDropdown) keyDown(keyCode unison.KeyCode, mods mod.Modifiers, repeat bool) bool {
	isReturn := keyCode == unison.KeyReturn || keyCode == unison.KeyNumPadEnter
	isDown := keyCode == unison.KeyDown && mods&mod.NonSticky == 0
	switch {
	case !s.suggestions.Showing():
		if (isReturn || isDown) && s.field.Text() != "" {
			if s.showMatches() && isDown {
				s.suggestions.Select(0)
			}
			return true
		}
	case isReturn && s.suggestions.Selected() < 0:
		s.suggestions.Select(0)
	}
	return s.field.DefaultKeyDown(keyCode, mods, repeat)
}

// Refresh brings the rows on offer back into line with the table after a change to it, which may have removed or
// renamed the rows they name. Only rows that are showing are found again, with the highlight kept on the row with the
// same key where it is still among them; hidden ones are dropped, to be found afresh when next asked for, so that a
// change walks the table only while the user is looking at the matches.
func (s *searchDropdown) Refresh() {
	if !s.suggestions.Showing() {
		if s.suggestions.List().Count() != 0 {
			s.suggestions.SetRows(nil)
		}
		return
	}
	var highlighted string
	if i := s.suggestions.Selected(); i >= 0 {
		highlighted = s.suggestions.List().DataAtIndex(i).key
	}
	found := s.findMatches(strings.ToLower(s.field.Text()))
	s.suggestions.SetRows(found)
	if highlighted == "" {
		return
	}
	for i, match := range found {
		if match.key == highlighted {
			s.suggestions.Select(i)
			return
		}
	}
}

// searchMatchCellFactory draws the dropdown's rows as the default factory does, with the match's icon, if it has one,
// ahead of its name and its title dimmed after it.
type searchMatchCellFactory struct {
	unison.DefaultCellFactory
}

// CreateCell implements unison.CellFactory.
func (f *searchMatchCellFactory) CreateCell(owner unison.Paneler, element any, row int, foreground, background unison.Ink, selected, focused bool) unison.Paneler {
	panel := f.DefaultCellFactory.CreateCell(owner, element, row, foreground, background, selected, focused)
	match, ok := element.(*searchMatch)
	label, ok2 := panel.(*unison.Label)
	if !ok || !ok2 {
		return panel
	}
	label.Drawable = match.icon
	label.Text = unison.NewText(match.name, &unison.TextDecoration{Font: unison.FieldFont, OnBackgroundInk: foreground})
	if match.title != "" {
		label.Text.AddString("  "+match.title, &unison.TextDecoration{
			Font:            fonts.FieldSecondary,
			OnBackgroundInk: &unison.ColorFilteredInk{OriginalInk: foreground, ColorFilter: unison.Alpha50Filter()},
		})
		label.Accessibility.Name = i18n.Text("%s, %s", match.name, match.title)
	} else {
		label.Accessibility.Name = match.name
	}
	return label
}
