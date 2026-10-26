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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
)

// maxPageSearchLabelLength is the most runes a page search match's label shows of the row's text.
const maxPageSearchLabelLength = 80

// searchableList is a list the page search looks through, with the title shown beside its matches.
type searchableList struct {
	list  sheetList
	title string
}

// installPageSearch adds the search field of a sheet, template or loot sheet to the toolbar and makes the
// jump-to-search command on the toolbar's parent focus it. The matches are the rows of the lists on the page; choosing
// one selects it alone, deselecting the rows of every list first, while the focus stays in the field. The lists are
// fetched afresh each time rather than captured, since a rebuild can replace any of them (see syncOrRebuildList).
func installPageSearch(toolbar *unison.Panel, lists func() []searchableList) *searchDropdown {
	s := newSearchDropdown(func(text string) []*searchMatch {
		var found []*searchMatch
		for _, one := range lists() {
			if len(found) > maxSearchMatches {
				break
			}
			one.list.search(&found, text, one.title)
		}
		return found
	}, func(match *searchMatch) {
		for _, one := range lists() {
			one.list.clearSelection()
		}
		match.show()
	})
	toolbar.AddChild(s.field)
	installJumpToField(toolbar.Parent(), s.field)
	return s
}

// searchTextReceiver is a dockable that can be handed search text to show the matches of, as the Library Explorer does
// when it opens a file that matched on its contents rather than its name.
type searchTextReceiver interface {
	receiveSearchText(text string)
}

// receiveSearchText puts the text in the search field and gives it the focus, offering the matches. Each of those
// offers them only when it changes something, so a field that already held both, with its matches hidden by Escape or
// by its window losing the focus, has them offered here.
func (pv *pageView) receiveSearchText(text string) {
	field := pv.search.field
	if !field.Focused() {
		// The text is searched for once the field gains the focus; see newSearchDropdown.
		field.SetText(text)
		field.RequestFocus()
		return
	}
	before := field.Text()
	field.SetText(text)
	if field.Text() == before && !pv.search.suggestions.Showing() {
		pv.search.showMatches()
	}
}

// search adds the rows of the list that match the text to the matches, but only when the list is on the page. A list
// the layout doesn't show has no parent (see Sheet.buildLayout), and a match found in one could only be shown by
// scrolling a table nobody is looking at into view. A list that doesn't exist has nothing to search.
func (p *PageList[T]) search(matches *[]*searchMatch, text, title string) {
	if p == nil || p.Parent() == nil {
		return
	}
	for _, row := range p.Table.RootRows() {
		searchPageListRows(matches, text, title, p.Table, row)
	}
}

// searchPageListRows adds the row and those beneath it that match the text to the matches, stopping once there is one
// more than maxSearchMatches of them (see searchMatchRows).
func searchPageListRows[T gurps.Node[T]](matches *[]*searchMatch, text, title string, table *unison.Table[*Node[T]], row *Node[T]) {
	if len(*matches) > maxSearchMatches {
		return
	}
	if row.Match(text) {
		*matches = append(*matches, &searchMatch{
			show:  func() { revealRow(table, row) },
			name:  row.searchLabel(),
			title: title,
			key:   string(row.ID()),
		})
	}
	if row.CanHaveChildren() {
		for _, child := range row.Children() {
			searchPageListRows(matches, text, title, table, child)
		}
	}
}

// revealRow discloses the row's ancestors, selects the row alone and scrolls it into view. Disclosing may have grown the
// table, and with it the page, but the scroll panel can only scroll as far as the page reached at its last layout, so
// the layout is validated from the scroll panel down before the scroll is asked for. For a row the table no longer
// holds, the selection is only cleared.
func revealRow[T gurps.Node[T]](table *unison.Table[*Node[T]], row *Node[T]) {
	table.DiscloseRow(row, false)
	table.ClearSelection()
	if index := table.RowToIndex(row); index >= 0 {
		table.SelectByIndex(index)
		if scroller := unison.Ancestor[*unison.ScrollPanel](table); scroller != nil {
			scroller.ValidateLayout()
		}
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
