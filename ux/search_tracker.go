// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/check"
)

type searchRef struct {
	table any
	row   any
}

// SearchTracker provides controls for searching.
type SearchTracker struct {
	clearTableSelections func()
	findMatches          func(refList *[]*searchRef, text string, namesOnly bool)
	backButton           *unison.Button
	forwardButton        *unison.Button
	matchesLabel         *unison.Label
	searchField          *unison.Field
	namesOnlyCheckBox    *unison.CheckBox
	searchResult         []*searchRef
	searchIndex          int
}

// InstallSearchTracker creates a search tracker in the given toolbar.
func InstallSearchTracker(toolbar *unison.Panel, clearTableSelections func(), findMatches func(refList *[]*searchRef, text string, namesOnly bool)) *SearchTracker {
	s := &SearchTracker{
		clearTableSelections: clearTableSelections,
		findMatches:          findMatches,
	}
	s.backButton = unison.NewSVGButton(svg.Back)
	s.backButton.Tooltip = newWrappedTooltip(i18n.Text("Previous Match"))
	s.backButton.ClickCallback = s.previousMatch
	s.backButton.SetEnabled(false)

	s.forwardButton = unison.NewSVGButton(svg.Forward)
	s.forwardButton.Tooltip = newWrappedTooltip(i18n.Text("Next Match"))
	s.forwardButton.ClickCallback = s.nextMatch
	s.forwardButton.SetEnabled(false)

	searchText := i18n.Text("Search")
	s.searchField = NewSearchField(searchText, s.searchModified)
	s.searchField.Tooltip = newWrappedTooltipWithSecondaryText(searchText, i18n.Text("Press RETURN to select the next match\nPress SHIFT-RETURN to select the previous match"))
	s.searchField.KeyDownCallback = func(keyCode unison.KeyCode, mod unison.Modifiers, repeat bool) bool {
		if keyCode == unison.KeyReturn || keyCode == unison.KeyNumPadEnter {
			if mod.ShiftDown() {
				s.previousMatch()
			} else {
				s.nextMatch()
			}
			return true
		}
		return s.searchField.DefaultKeyDown(keyCode, mod, repeat)
	}

	s.namesOnlyCheckBox = unison.NewCheckBox()
	s.namesOnlyCheckBox.SetTitle(i18n.Text("Names Only"))
	s.namesOnlyCheckBox.ClickCallback = func() { s.doSearch(s.searchField.Text()) }

	s.matchesLabel = unison.NewLabel()
	s.matchesLabel.SetTitle(i18n.Text("0 of 0"))
	s.matchesLabel.Tooltip = newWrappedTooltip(i18n.Text("Number of matches found"))

	toolbar.AddChild(s.backButton)
	toolbar.AddChild(s.forwardButton)
	toolbar.AddChild(s.searchField)
	toolbar.AddChild(s.matchesLabel)
	toolbar.AddChild(s.namesOnlyCheckBox)

	toolbar.Parent().InstallCmdHandlers(JumpToSearchFilterItemID,
		func(any) bool { return !s.searchField.Focused() },
		func(any) { s.searchField.RequestFocus() })
	return s
}

// Refresh the search state.
func (s *SearchTracker) Refresh() {
	s.searchResult = nil
	s.findMatches(&s.searchResult, strings.ToLower(s.searchField.Text()), s.namesOnlyCheckBox.State == check.On)
	s.searchIndex = max(min(s.searchIndex, len(s.searchResult)-1), 0)
	s.adjustButtonsAndLabels()
}

func (s *SearchTracker) searchModified(_, after *unison.FieldState) {
	s.doSearch(after.Text)
}

func (s *SearchTracker) doSearch(text string) {
	s.searchIndex = 0
	s.searchResult = nil
	s.findMatches(&s.searchResult, strings.ToLower(text), s.namesOnlyCheckBox.State == check.On)
	s.adjustForMatch()
}

func searchSheetTable[T gurps.NodeTypes](refList *[]*searchRef, text string, namesOnly bool, pageList *PageList[T]) {
	for _, row := range pageList.Table.RootRows() {
		searchSheetTableRows(refList, text, namesOnly, pageList.Table, row)
	}
}

func searchSheetTableRows[T gurps.NodeTypes](refList *[]*searchRef, text string, namesOnly bool, table *unison.Table[*Node[T]], row *Node[T]) {
	if text != "" {
		if namesOnly {
			if strings.Contains(strings.ToLower(row.dataAsNode.String()), text) {
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

func (s *SearchTracker) previousMatch() {
	if s.searchIndex > 0 {
		s.searchIndex--
		s.adjustForMatch()
	}
}

func (s *SearchTracker) nextMatch() {
	if s.searchIndex < len(s.searchResult)-1 {
		s.searchIndex++
		s.adjustForMatch()
	}
}

func (s *SearchTracker) adjustForMatch() {
	s.clearTableSelections()
	s.adjustButtonsAndLabels()
	if len(s.searchResult) != 0 {
		showSearchRef(s.searchResult[s.searchIndex])
	}
}

func (s *SearchTracker) adjustButtonsAndLabels() {
	s.backButton.SetEnabled(s.searchIndex != 0)
	s.forwardButton.SetEnabled(len(s.searchResult) != 0 && s.searchIndex != len(s.searchResult)-1)
	if len(s.searchResult) != 0 {
		s.matchesLabel.SetTitle(fmt.Sprintf(i18n.Text("%d of %d"), s.searchIndex+1, len(s.searchResult)))
	} else {
		s.matchesLabel.SetTitle(i18n.Text("0 of 0"))
	}
	s.matchesLabel.Parent().MarkForLayoutAndRedraw()
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

func showSearchResolvedRef[T gurps.NodeTypes](table *unison.Table[*Node[T]], row *Node[T]) {
	table.DiscloseRow(row, false)
	table.ClearSelection()
	rowIndex := table.RowToIndex(row)
	table.SelectByIndex(rowIndex)
	table.ScrollRowIntoView(rowIndex)
}
