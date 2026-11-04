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
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// labelTitles returns the titles of every label found anywhere beneath the given panel.
func labelTitles(p *unison.Panel) []string {
	labels := uxtest.PanelsOfType[*unison.Label](p)
	titles := make([]string, 0, len(labels))
	for _, label := range labels {
		titles = append(titles, label.String())
	}
	return titles
}

// TestModifierContainerEditorsShowOnlyWhatAContainerUses verifies that a modifier group's editor leaves out every
// field that only a modifier in its own right uses, and that a choice's editor adds only what the choice asks for.
func TestModifierContainerEditorsShowOnlyWhatAContainerUses(t *testing.T) {
	c := check.New(t)
	leafOnly := []string{"Cost", "Level", "Total", "Tech Level", "Cost Modifier", "Weight Modifier", "VTT Notes"}
	checkNone := func(titles []string) {
		c.Helper()
		for _, title := range leafOnly {
			c.False(slices.Contains(titles, title), "a container's editor must not offer %q", title)
		}
	}

	_, content := buildEditorContent(nil, gurps.NewTraitModifier(nil, nil, true), initTraitModifierEditor)
	titles := labelTitles(content)
	checkNone(titles)
	c.True(slices.Contains(titles, "Notes"))
	c.False(slices.Contains(titles, "Choice"), "a group asks for no choice")
	c.Equal(0, len(checkBoxesTitled(content, "Enabled")), "a container is always enabled")
	c.Equal(0, len(checkBoxesTitled(content, "Show notes in owner's notes")))
	c.Equal(0, len(checkBoxesTitled(content, "Show notes in attack usage")))
	c.Equal(0, len(checkBoxesTitled(content, "Show in title notes")))

	choice := gurps.NewEquipmentModifierChoice(nil, nil)
	e, content := buildEditorContent(nil, choice, initEquipmentModifierEditor)
	titles = labelTitles(content)
	checkNone(titles)
	c.True(slices.Contains(titles, "Choice"), "a choice says what it asks for")
	popups := uxtest.PanelsOfType[*unison.PopupMenu[string]](content)
	c.Equal(1, len(popups))
	popups[0].SelectIndex(1)
	c.False(e.editorData.IsMandatoryChoice(), "the second item makes the choice optional")
	c.True(e.editorData.IsChoice(), "the choice can't be taken out of use here")
	c.True(choice.IsMandatoryChoice(), "nothing reaches the modifier until the edit is applied")
}
