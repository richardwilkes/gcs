/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

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

type searchTracker struct {
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

func installSearchTracker(toolbar *unison.Panel, clearTableSelections func(), findMatches func(refList *[]*searchRef, text string, namesOnly bool)) {
	s := &searchTracker{
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
	s.namesOnlyCheckBox.Text = i18n.Text("Names Only")
	s.namesOnlyCheckBox.ClickCallback = func() { s.doSearch(s.searchField.Text()) }

	s.matchesLabel = unison.NewLabel()
	s.matchesLabel.Text = i18n.Text("0 of 0")
	s.matchesLabel.Tooltip = newWrappedTooltip(i18n.Text("Number of matches found"))

	toolbar.AddChild(s.backButton)
	toolbar.AddChild(s.forwardButton)
	toolbar.AddChild(s.searchField)
	toolbar.AddChild(s.matchesLabel)
	toolbar.AddChild(s.namesOnlyCheckBox)

	toolbar.Parent().InstallCmdHandlers(JumpToSearchFilterItemID,
		func(any) bool { return !s.searchField.Focused() },
		func(any) { s.searchField.RequestFocus() })
}

func (s *searchTracker) searchModified(_, after *unison.FieldState) {
	s.doSearch(after.Text)
}

func (s *searchTracker) doSearch(text string) {
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

func (s *searchTracker) previousMatch() {
	if s.searchIndex > 0 {
		s.searchIndex--
		s.adjustForMatch()
	}
}

func (s *searchTracker) nextMatch() {
	if s.searchIndex < len(s.searchResult)-1 {
		s.searchIndex++
		s.adjustForMatch()
	}
}

func (s *searchTracker) adjustForMatch() {
	s.clearTableSelections()
	s.backButton.SetEnabled(s.searchIndex != 0)
	s.forwardButton.SetEnabled(len(s.searchResult) != 0 && s.searchIndex != len(s.searchResult)-1)
	if len(s.searchResult) != 0 {
		s.matchesLabel.Text = fmt.Sprintf(i18n.Text("%d of %d"), s.searchIndex+1, len(s.searchResult))
		showSearchRef(s.searchResult[s.searchIndex])
	} else {
		s.matchesLabel.Text = i18n.Text("0 of 0")
	}
	s.matchesLabel.Parent().MarkForLayoutAndRedraw()
}

func showSearchRef(ref *searchRef) {
	switch table := ref.table.(type) {
	case *unison.Table[*Node[*gurps.Trait]]:
		showSearchResolvedRef(table, ref.row.(*Node[*gurps.Trait]))
	case *unison.Table[*Node[*gurps.Skill]]:
		showSearchResolvedRef(table, ref.row.(*Node[*gurps.Skill]))
	case *unison.Table[*Node[*gurps.Spell]]:
		showSearchResolvedRef(table, ref.row.(*Node[*gurps.Spell]))
	case *unison.Table[*Node[*gurps.Equipment]]:
		showSearchResolvedRef(table, ref.row.(*Node[*gurps.Equipment]))
	case *unison.Table[*Node[*gurps.Note]]:
		showSearchResolvedRef(table, ref.row.(*Node[*gurps.Note]))
	}
}

func showSearchResolvedRef[T gurps.NodeTypes](table *unison.Table[*Node[T]], row *Node[T]) {
	table.DiscloseRow(row, false)
	table.ClearSelection()
	rowIndex := table.RowToIndex(row)
	table.SelectByIndex(rowIndex)
	table.ScrollRowIntoView(rowIndex)
}
