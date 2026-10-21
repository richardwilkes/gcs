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
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// matchStepper holds the controls and state a search toolbar uses for stepping through matches. A searchIndex of -1
// means no match is current yet, which the label shows as "- of N".
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
// searchModified is called when the search field's text changes and showMatch to select and reveal the current match.
// Nothing is added to a parent; see addControlsTo and installJumpToSearchHandlers.
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

func (m *matchStepper[T]) addControlsTo(panel *unison.Panel) {
	panel.AddChild(m.backButton)
	panel.AddChild(m.forwardButton)
	panel.AddChild(m.searchField)
	panel.AddChild(m.matchesLabel)
}

// installJumpToSearchHandlers makes the jump-to-search command on the given panel move the focus to the search field.
func (m *matchStepper[T]) installJumpToSearchHandlers(panel *unison.Panel) {
	installJumpToField(panel, m.searchField)
}

// installJumpToField makes the jump-to-search command on the given panel move the focus to the field.
func installJumpToField(panel *unison.Panel, field *unison.Field) {
	panel.InstallCmdHandlers(JumpToSearchFilterItemID,
		func(any) bool { return !field.Focused() },
		func(any) { field.RequestFocus() })
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
// search field, stepping between matches) may call this; paths that merely refresh the results must use
// updateMatchControls instead, since grabbing the selection after the user has moved on would discard their place.
func (m *matchStepper[T]) adjustForMatch() {
	m.updateMatchControls()
	if m.searchIndex >= 0 && m.searchIndex < len(m.searchResult) {
		m.showMatch(m.searchResult[m.searchIndex])
	}
}

// updateMatchControls updates the buttons and the matches label for the current results without touching any selection
// or scroll position.
func (m *matchStepper[T]) updateMatchControls() {
	m.backButton.SetEnabled(m.searchIndex > 0)
	m.forwardButton.SetEnabled(len(m.searchResult) != 0 && m.searchIndex != len(m.searchResult)-1)
	switch {
	case len(m.searchResult) == 0:
		m.matchesLabel.SetTitle("-")
	case m.searchIndex < 0:
		m.matchesLabel.SetTitle(i18n.Text("- of %d", len(m.searchResult)))
	default:
		m.matchesLabel.SetTitle(i18n.Text("%d of %d", m.searchIndex+1, len(m.searchResult)))
	}
	m.matchesLabel.Parent().MarkForLayoutAndRedraw()
}
