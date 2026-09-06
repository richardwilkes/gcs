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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/mod"
)

type searchRef struct {
	table any
	row   any
}

// matchStepper holds the controls and state the search toolbars share for stepping through matches: the back and
// forward buttons, the search field, the label showing the position within the matches, the matches themselves and the
// index of the current one. A searchIndex of -1 means no match is current yet, which the label shows as "- of N".
type matchStepper[T any] struct {
	backButton    *unison.Button
	forwardButton *unison.Button
	searchField   *unison.Field
	matchesLabel  *unison.Label
	showMatch     func(match T)
	searchResult  []T
	searchIndex   int
}

// setupControls creates the controls: the back and forward buttons, disabled until there is somewhere to step, the
// search field, whose RETURN and SHIFT-RETURN step to the next and previous match, and the matches label.
// searchModified is called when the search field's text changes and showMatch is called to select and reveal the
// current match. Nothing is added to a parent; see addControlsTo and installJumpToSearchHandlers.
func (m *matchStepper[T]) setupControls(searchModified func(before, after *unison.FieldState), showMatch func(match T)) {
	m.showMatch = showMatch

	m.backButton = unison.NewSVGButton(svg.Back)
	m.backButton.Tooltip = newWrappedTooltip(i18n.Text("Previous Match"))
	m.backButton.ClickCallback = m.previousMatch
	m.backButton.SetEnabled(false)

	m.forwardButton = unison.NewSVGButton(svg.Forward)
	m.forwardButton.Tooltip = newWrappedTooltip(i18n.Text("Next Match"))
	m.forwardButton.ClickCallback = m.nextMatch
	m.forwardButton.SetEnabled(false)

	searchText := i18n.Text("Search")
	m.searchField = NewSearchField(searchText, searchModified)
	m.searchField.Tooltip = newWrappedTooltipWithSecondaryText(searchText, i18n.Text("Press RETURN to select the next match\nPress SHIFT-RETURN to select the previous match"))
	m.searchField.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, repeat bool) bool {
		if keyCode == unison.KeyReturn || keyCode == unison.KeyNumPadEnter {
			if mods.ShiftDown() {
				m.previousMatch()
			} else {
				m.nextMatch()
			}
			return true
		}
		return m.searchField.DefaultKeyDown(keyCode, mods, repeat)
	}

	m.matchesLabel = unison.NewLabel()
	m.matchesLabel.SetTitle("-")
	m.matchesLabel.Tooltip = newWrappedTooltip(i18n.Text("Number of matches found"))
}

// addControlsTo adds the controls to the given panel, in order.
func (m *matchStepper[T]) addControlsTo(panel *unison.Panel) {
	panel.AddChild(m.backButton)
	panel.AddChild(m.forwardButton)
	panel.AddChild(m.searchField)
	panel.AddChild(m.matchesLabel)
}

// installJumpToSearchHandlers installs the handlers for the jump-to-search command on the given panel, so that the
// command moves the focus to the search field.
func (m *matchStepper[T]) installJumpToSearchHandlers(panel *unison.Panel) {
	panel.InstallCmdHandlers(JumpToSearchFilterItemID,
		func(any) bool { return !m.searchField.Focused() },
		func(any) { m.searchField.RequestFocus() })
}

func (m *matchStepper[T]) previousMatch() {
	if m.searchIndex > 0 {
		m.searchIndex--
		m.adjustForMatch()
	}
}

func (m *matchStepper[T]) nextMatch() {
	if m.searchIndex < len(m.searchResult)-1 {
		m.searchIndex++
		m.adjustForMatch()
	}
}

// adjustForMatch updates the controls and, when a match is current, shows it. Only direct user actions (typing in the
// search field, stepping between matches) should call this; asynchronous paths that merely refresh the results must
// use updateMatchControls instead, since grabbing the selection while the user may have moved on to other rows would
// discard their place.
func (m *matchStepper[T]) adjustForMatch() {
	m.updateMatchControls()
	if m.searchIndex >= 0 && m.searchIndex < len(m.searchResult) {
		m.showMatch(m.searchResult[m.searchIndex])
	}
}

// updateMatchControls updates the back and forward buttons and the matches label for the current search results
// without touching any selection or scroll position.
func (m *matchStepper[T]) updateMatchControls() {
	m.backButton.SetEnabled(m.searchIndex > 0)
	m.forwardButton.SetEnabled(len(m.searchResult) != 0 && m.searchIndex != len(m.searchResult)-1)
	switch {
	case len(m.searchResult) == 0:
		m.matchesLabel.SetTitle("-")
	case m.searchIndex < 0:
		m.matchesLabel.SetTitle(fmt.Sprintf(i18n.Text("- of %d"), len(m.searchResult)))
	default:
		m.matchesLabel.SetTitle(fmt.Sprintf(i18n.Text("%d of %d"), m.searchIndex+1, len(m.searchResult)))
	}
	m.matchesLabel.Parent().MarkForLayoutAndRedraw()
}

// SearchTracker provides controls for searching.
type SearchTracker struct {
	matchStepper[*searchRef]
	clearTableSelections func()
	findMatches          func(refList *[]*searchRef, text string, namesOnly bool)
	namesOnlyCheckBox    *unison.CheckBox
}

// InstallSearchTracker creates a search tracker in the given toolbar.
func InstallSearchTracker(toolbar *unison.Panel, clearTableSelections func(), findMatches func(refList *[]*searchRef, text string, namesOnly bool)) *SearchTracker {
	s := &SearchTracker{
		clearTableSelections: clearTableSelections,
		findMatches:          findMatches,
	}
	s.setupControls(s.searchModified, func(ref *searchRef) {
		clearTableSelections()
		showSearchRef(ref)
	})

	s.namesOnlyCheckBox = unison.NewCheckBox()
	s.namesOnlyCheckBox.SetTitle(i18n.Text("Names Only"))
	s.namesOnlyCheckBox.ClickCallback = func() { s.doSearch(s.searchField.Text()) }

	s.addControlsTo(toolbar)
	toolbar.AddChild(s.namesOnlyCheckBox)
	s.installJumpToSearchHandlers(toolbar.Parent())
	return s
}

// Refresh the search state.
func (s *SearchTracker) Refresh() {
	s.searchResult = nil
	s.findMatches(&s.searchResult, strings.ToLower(s.searchField.Text()), s.namesOnlyCheckBox.State == check.On)
	s.searchIndex = max(min(s.searchIndex, len(s.searchResult)-1), 0)
	s.updateMatchControls()
}

func (s *SearchTracker) searchModified(_, after *unison.FieldState) {
	s.doSearch(after.Text)
}

// doSearch runs a fresh search and shows its first match. The selections are cleared even when nothing matches, so
// that clearing the search field deselects the last match.
func (s *SearchTracker) doSearch(text string) {
	s.searchIndex = 0
	s.searchResult = nil
	s.findMatches(&s.searchResult, strings.ToLower(text), s.namesOnlyCheckBox.State == check.On)
	s.clearTableSelections()
	s.adjustForMatch()
}

// search adds the rows of the list that match the text to the reference list, but only when the list is on the page.
// A list the layout doesn't show has no parent (see Sheet.buildLayout), and a match found in one could only be shown
// by scrolling a table nobody is looking at into view. A list that doesn't exist has nothing to search.
func (p *PageList[T]) search(refList *[]*searchRef, text string, namesOnly bool) {
	if p == nil || p.Parent() == nil {
		return
	}
	for _, row := range p.Table.RootRows() {
		searchSheetTableRows(refList, text, namesOnly, p.Table, row)
	}
}

// installListSearchTracker installs a search tracker on the toolbar that searches the lists the function returns and
// clears their selections when the search is cleared. The lists are fetched afresh each time rather than captured,
// since a rebuild can replace any of them (see syncOrRebuildList).
func installListSearchTracker(toolbar *unison.Panel, lists func() []sheetList) *SearchTracker {
	return InstallSearchTracker(toolbar, func() {
		for _, list := range lists() {
			list.clearSelection()
		}
	}, func(refList *[]*searchRef, text string, namesOnly bool) {
		for _, list := range lists() {
			list.search(refList, text, namesOnly)
		}
	})
}

func searchSheetTableRows[T gurps.Node[T]](refList *[]*searchRef, text string, namesOnly bool, table *unison.Table[*Node[T]], row *Node[T]) {
	if text != "" {
		if namesOnly {
			if strings.Contains(strings.ToLower(row.data.String()), text) {
				*refList = append(*refList, &searchRef{
					table: table,
					row:   row,
				})
			}
		} else {
			if row.Match(text) {
				*refList = append(*refList, &searchRef{
					table: table,
					row:   row,
				})
			}
		}
	}
	if row.CanHaveChildren() {
		for _, child := range row.Children() {
			searchSheetTableRows(refList, text, namesOnly, table, child)
		}
	}
}

func showSearchRef(ref *searchRef) {
	switch table := ref.table.(type) {
	case *unison.Table[*Node[*gurps.ConditionalModifier]]:
		if row, ok := ref.row.(*Node[*gurps.ConditionalModifier]); ok {
			showSearchResolvedRef(table, row)
		}
	case *unison.Table[*Node[*gurps.Weapon]]:
		if row, ok := ref.row.(*Node[*gurps.Weapon]); ok {
			showSearchResolvedRef(table, row)
		}
	case *unison.Table[*Node[*gurps.Trait]]:
		if row, ok := ref.row.(*Node[*gurps.Trait]); ok {
			showSearchResolvedRef(table, row)
		}
	case *unison.Table[*Node[*gurps.Skill]]:
		if row, ok := ref.row.(*Node[*gurps.Skill]); ok {
			showSearchResolvedRef(table, row)
		}
	case *unison.Table[*Node[*gurps.Spell]]:
		if row, ok := ref.row.(*Node[*gurps.Spell]); ok {
			showSearchResolvedRef(table, row)
		}
	case *unison.Table[*Node[*gurps.Equipment]]:
		if row, ok := ref.row.(*Node[*gurps.Equipment]); ok {
			showSearchResolvedRef(table, row)
		}
	case *unison.Table[*Node[*gurps.Note]]:
		if row, ok := ref.row.(*Node[*gurps.Note]); ok {
			showSearchResolvedRef(table, row)
		}
	}
}

func showSearchResolvedRef[T gurps.Node[T]](table *unison.Table[*Node[T]], row *Node[T]) {
	table.DiscloseRow(row, false)
	table.ClearSelection()
	rowIndex := table.RowToIndex(row)
	table.SelectByIndex(rowIndex)
	table.ScrollRowIntoView(rowIndex)
}
