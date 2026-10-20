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
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// TestMatchStepper checks the stepping the search toolbars share: the buttons enable only when there is a match to step
// to in their direction, the label shows the position -- or "-" in place of one when no match is current yet, and alone
// when there are no matches -- stepping shows the match it lands on while a step past either end does nothing, and
// RETURN and SHIFT-RETURN in the search field step the same way the buttons do.
func TestMatchStepper(t *testing.T) {
	c := check.New(t)
	var shown []string
	var m matchStepper[string]
	m.setupControls(func(_, _ *unison.FieldState) {}, func(match string) { shown = append(shown, match) })
	m.addControlsTo(unison.NewPanel()) // The label's refresh marks its parent for layout.
	c.Equal("-", m.matchesLabel.String())
	c.False(m.backButton.Enabled())
	c.False(m.forwardButton.Enabled())

	m.searchResult = []string{"a", "b", "c"}
	m.searchIndex = -1
	m.adjustForMatch()
	c.Equal("- of 3", m.matchesLabel.String())
	c.False(m.backButton.Enabled(), "with no match current, there is nothing to step back to")
	c.True(m.forwardButton.Enabled())
	c.Equal(0, len(shown), "with no match current, nothing is shown")

	m.previousMatch()
	c.Equal(-1, m.searchIndex, "stepping back from no match stays put")
	c.Equal(0, len(shown))

	m.nextMatch()
	c.Equal(0, m.searchIndex)
	c.Equal("1 of 3", m.matchesLabel.String())
	c.Equal([]string{"a"}, shown)
	c.False(m.backButton.Enabled())
	c.True(m.forwardButton.Enabled())

	c.True(m.searchField.KeyDownCallback(unison.KeyReturn, mod.None, false), "RETURN is consumed")
	c.True(m.searchField.KeyDownCallback(unison.KeyNumPadEnter, mod.None, false), "the numeric keypad's ENTER, too")
	c.Equal(2, m.searchIndex, "RETURN steps forward")
	c.Equal("3 of 3", m.matchesLabel.String())
	c.Equal([]string{"a", "b", "c"}, shown)
	c.True(m.backButton.Enabled())
	c.False(m.forwardButton.Enabled(), "on the last match, there is nothing to step forward to")

	m.nextMatch()
	c.Equal(2, m.searchIndex, "stepping forward from the last match stays put")
	c.Equal(3, len(shown))

	c.True(m.searchField.KeyDownCallback(unison.KeyReturn, mod.Shift, false))
	c.Equal(1, m.searchIndex, "SHIFT-RETURN steps back")
	c.Equal("2 of 3", m.matchesLabel.String())
	c.Equal("b", shown[len(shown)-1])
	c.True(m.backButton.Enabled())
	c.True(m.forwardButton.Enabled())

	// A refresh of the controls alone must not show anything.
	shown = nil
	m.updateMatchControls()
	c.Equal(0, len(shown))

	m.searchResult = nil
	m.searchIndex = 0
	m.adjustForMatch()
	c.Equal("-", m.matchesLabel.String())
	c.False(m.backButton.Enabled())
	c.False(m.forwardButton.Enabled())
	c.Equal(0, len(shown), "with no matches, nothing is shown")
}

// An empty search matches nothing, so refreshing it must not walk the lists: the walk has every row of every list
// derive its children and the data of its cells afresh, which was most of what an update of a sheet with large lists
// cost while nobody was searching. The controls must nevertheless end up as a search that found nothing leaves them.
func TestSearchRefreshSkipsTheListsWhileTheFieldIsEmpty(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	trait := gurps.NewTrait(sheet.Entity(), nil, false)
	trait.Name = "Findable"
	sheet.Entity().Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)
	tracker := sheet.searchTracker
	walks := 0
	findMatches := tracker.findMatches
	tracker.findMatches = func(refList *[]*searchRef, text string, namesOnly bool) {
		walks++
		findMatches(refList, text, namesOnly)
	}
	clears := 0
	clearTableSelections := tracker.clearTableSelections
	tracker.clearTableSelections = func() {
		clears++
		clearTableSelections()
	}

	tracker.Refresh()
	c.Equal(0, walks, "an empty search must not walk the lists")
	c.Equal("-", tracker.matchesLabel.String(), "an empty search has no matches to report")
	c.False(tracker.backButton.Enabled(), "there is no previous match to step to")
	c.False(tracker.forwardButton.Enabled(), "there is no next match to step to")

	tracker.searchField.SetText("findable")
	c.Equal(1, walks, "typing into the field searches the lists")
	c.Equal(1, len(tracker.searchResult), "the trait must be found")
	c.Equal("1 of 1", tracker.matchesLabel.String())
	tracker.Refresh()
	c.Equal(2, walks, "a refresh with text in the field searches the lists again")
	c.Equal(1, len(tracker.searchResult), "the trait must still be found")

	clears = 0
	tracker.searchField.SetText("")
	c.Equal(2, walks, "clearing the field must not walk the lists")
	c.Equal(1, clears, "clearing the field still deselects the last match")
	c.Equal(0, len(tracker.searchResult), "an empty search finds nothing")
	c.Equal("-", tracker.matchesLabel.String())
	c.False(tracker.backButton.Enabled())
	c.False(tracker.forwardButton.Enabled())
}
