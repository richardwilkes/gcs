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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

const (
	// maxPageSearchMatches is the most rows the search offers; the walk stops once it has that many, since the list
	// that shows them builds a cell for every row whenever it is sized.
	maxPageSearchMatches = 100
	// maxPageSearchLabelLength is the most runes a match's label shows of the row's text.
	maxPageSearchLabelLength = 80
)

// searchableList is a list the page search looks through, with the title shown beside its matches.
type searchableList struct {
	list  sheetList
	title string
}

// pageSearchMatch is a row the page search found: its label, the title of the list it is in, its ID, and a function
// that discloses, selects and scrolls to it in its table.
type pageSearchMatch struct {
	show  func()
	name  string
	title string
	id    tid.TID
}

func (m *pageSearchMatch) String() string {
	return m.name
}

// pageSearch is the search field in the toolbar of a sheet, template or loot sheet. The rows matching its text are
// offered in a dropdown as the user types; choosing one selects it on the page. The lists are only walked while the
// dropdown can be seen, since the walk has every row of a page list derive its children and the data of its cells
// afresh as it is read (see rowReuse), a cost a sheet with large lists can ill afford on every update.
type pageSearch struct {
	field       *unison.Field
	suggestions *unison.FieldSuggestions[*pageSearchMatch]
	lists       func() []searchableList
	// findMatches returns the rows matching the lower-cased text. A field so that tests can count the walks.
	findMatches func(text string) []*pageSearchMatch
}

// installPageSearch adds the search field to the toolbar and makes the jump-to-search command on the toolbar's parent
// focus it. The lists are fetched afresh each time rather than captured, since a rebuild can replace any of them (see
// syncOrRebuildList).
func installPageSearch(toolbar *unison.Panel, lists func() []searchableList) *pageSearch {
	s := &pageSearch{lists: lists}
	s.findMatches = s.matches
	searchText := i18n.Text("Search")
	s.field = NewSearchField(searchText, s.textModified)
	s.field.Tooltip = newWrappedTooltipWithSecondaryText(searchText,
		i18n.Text("Lists the matching rows as you type\nPress RETURN to select the highlighted row"))
	s.suggestions = unison.NewFieldSuggestions[*pageSearchMatch](s.field)
	s.suggestions.List().Factory = &pageSearchCellFactory{}
	s.suggestions.ChosenCallback = s.choose
	s.field.KeyDownCallback = s.keyDown
	// Text put in the field while it lacked the focus, by the Library Explorer or the user, is searched for once it
	// has it, which is also when the rows can first be shown.
	s.field.GainedFocusCallback = func() {
		s.field.DefaultFocusGained()
		s.showMatches()
	}
	toolbar.AddChild(s.field)
	installJumpToField(toolbar.Parent(), s.field)
	return s
}

// matches returns the rows matching the lower-cased text across the lists on the page, up to maxPageSearchMatches.
func (s *pageSearch) matches(text string) []*pageSearchMatch {
	var found []*pageSearchMatch
	for _, one := range s.lists() {
		if len(found) >= maxPageSearchMatches {
			break
		}
		one.list.search(&found, text, one.title)
	}
	return found
}

// showMatches finds the rows matching the field's text and offers them, returning true if there were any. Nothing
// matches an empty search, so the lists aren't walked for one.
func (s *pageSearch) showMatches() bool {
	text := strings.ToLower(s.field.Text())
	if text == "" {
		s.suggestions.SetRows(nil)
		return false
	}
	found := s.findMatches(text)
	s.suggestions.SetRows(found)
	return len(found) != 0
}

// textModified offers the rows matching the new text. The lists are only walked while the field has the focus, since
// the rows can only be shown then (see unison.FieldSuggestions.Show); see installPageSearch for text set at other
// times. The table selections are left alone: a row is only selected by choosing it.
func (s *pageSearch) textModified(_, after *unison.FieldState) {
	if after.Text == "" || !s.field.Focused() {
		s.suggestions.SetRows(nil)
		return
	}
	s.suggestions.SetRows(s.findMatches(strings.ToLower(after.Text)))
}

// choose selects the chosen row on the page, deselecting everything else first. The focus stays in the field.
func (s *pageSearch) choose(_ int, match *pageSearchMatch) {
	for _, one := range s.lists() {
		one.list.clearSelection()
	}
	match.show()
}

// keyDown gives the field's keys their meaning for the search. While the rows are hidden, RETURN or a bare Down offers
// them again, Down with the first highlighted; the lists are walked afresh, since they may have changed. While they
// show with none highlighted, RETURN chooses the first, where it would otherwise merely hide them. Everything else is
// left to the field, whose default handling gives the rows first crack at the keys that drive them.
func (s *pageSearch) keyDown(keyCode unison.KeyCode, mods mod.Modifiers, repeat bool) bool {
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

// Refresh brings the rows on offer back into line with the lists after an update, which may have removed or renamed
// the rows they name. Only rows that are showing are found again, with the highlight kept on the same row where it
// is still among them; hidden ones are dropped, to be found afresh when next asked for, so that an update walks the
// lists only while the user is looking at the matches.
func (s *pageSearch) Refresh() {
	if !s.suggestions.Showing() {
		if s.suggestions.List().Count() != 0 {
			s.suggestions.SetRows(nil)
		}
		return
	}
	var highlighted tid.TID
	if i := s.suggestions.Selected(); i >= 0 {
		highlighted = s.suggestions.List().DataAtIndex(i).id
	}
	found := s.findMatches(strings.ToLower(s.field.Text()))
	s.suggestions.SetRows(found)
	if highlighted == "" {
		return
	}
	for i, match := range found {
		if match.id == highlighted {
			s.suggestions.Select(i)
			return
		}
	}
}

// searchTextReceiver is a dockable that can be handed search text to show the matches of, as the Library Explorer does
// when it opens a file that matched on its contents rather than its name.
type searchTextReceiver interface {
	receiveSearchText(text string)
}

// receiveSearchText puts the text in the search field and gives it the focus, which shows the matches.
func (pv *pageView) receiveSearchText(text string) {
	pv.search.field.SetText(text)
	pv.search.field.RequestFocus()
}

// search adds the rows of the list that match the text to the matches, but only when the list is on the page. A list
// the layout doesn't show has no parent (see Sheet.buildLayout), and a match found in one could only be shown by
// scrolling a table nobody is looking at into view. A list that doesn't exist has nothing to search.
func (p *PageList[T]) search(matches *[]*pageSearchMatch, text, title string) {
	if p == nil || p.Parent() == nil {
		return
	}
	for _, row := range p.Table.RootRows() {
		searchPageListRows(matches, text, title, p.Table, row)
	}
}

// searchPageListRows adds the row and those beneath it that match the text to the matches, stopping once there are
// maxPageSearchMatches of them.
func searchPageListRows[T gurps.Node[T]](matches *[]*pageSearchMatch, text, title string, table *unison.Table[*Node[T]], row *Node[T]) {
	if len(*matches) >= maxPageSearchMatches {
		return
	}
	if row.Match(text) {
		*matches = append(*matches, &pageSearchMatch{
			show:  func() { revealRow(table, row) },
			name:  row.searchLabel(),
			title: title,
			id:    row.ID(),
		})
	}
	if row.CanHaveChildren() {
		for _, child := range row.Children() {
			searchPageListRows(matches, text, title, table, child)
		}
	}
}

// revealRow discloses the row's ancestors, selects the row alone and scrolls it into view. A row the table no longer
// holds is left alone.
func revealRow[T gurps.Node[T]](table *unison.Table[*Node[T]], row *Node[T]) {
	table.DiscloseRow(row, false)
	table.ClearSelection()
	if index := table.RowToIndex(row); index >= 0 {
		table.SelectByIndex(index)
		table.ScrollRowIntoView(index)
	}
}

// searchLabel returns the text the page search labels the row with: the first line of what the row's hierarchy column
// (its first column, when it has none) shows, cut to maxPageSearchLabelLength runes. A weapon adds its usage after the
// cut, since the column shows the weapon's owner, which several weapons may share. The row's String() would not do: a
// note's is its whole text and a weapon's is its owner's name.
func (n *Node[T]) searchLabel() string {
	col := max(n.table.ColumnIndexForID(n.table.HierarchyColumnID), 0)
	data := n.cellData(col, true)
	label := firstNonBlankLine(data.Primary)
	if data.Type == cell.Markdown {
		label = strings.TrimLeft(label, "#> ")
	}
	label = xstrings.Truncate(label, maxPageSearchLabelLength, true)
	if w, ok := any(n.data).(*gurps.Weapon); ok {
		if usage := w.UsageWithReplacements(); usage != "" {
			label = i18n.Text("%s (%s)", label, usage)
		}
	}
	return label
}

// firstNonBlankLine returns the first line of s with anything on it, trimmed of surrounding space.
func firstNonBlankLine(s string) string {
	for line := range strings.Lines(s) {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return ""
}

// pageSearchCellFactory draws the search's rows as the default factory does, with the title of the list the row is in
// dimmed after its name.
type pageSearchCellFactory struct {
	unison.DefaultCellFactory
}

// CreateCell implements unison.CellFactory.
func (f *pageSearchCellFactory) CreateCell(owner unison.Paneler, element any, row int, foreground, background unison.Ink, selected, focused bool) unison.Paneler {
	panel := f.DefaultCellFactory.CreateCell(owner, element, row, foreground, background, selected, focused)
	match, ok := element.(*pageSearchMatch)
	label, ok2 := panel.(*unison.Label)
	if !ok || !ok2 {
		return panel
	}
	label.Text = unison.NewText(match.name, &unison.TextDecoration{Font: unison.FieldFont, OnBackgroundInk: foreground})
	label.Text.AddString("  "+match.title, &unison.TextDecoration{
		Font:            fonts.FieldSecondary,
		OnBackgroundInk: &unison.ColorFilteredInk{OriginalInk: foreground, ColorFilter: unison.Alpha50Filter()},
	})
	label.Accessibility.Name = i18n.Text("%s, %s", match.name, match.title)
	return label
}
