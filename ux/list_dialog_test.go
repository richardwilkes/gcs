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

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestNewListQuestionPanel verifies the shape of the panel the list question dialogs share: the operation label, if
// any, then the header, then any extra headers in order, and last the list's scroll panel, the only child that takes up
// spare room. The extra headers are how the modifier prompt names the row being asked about, so they must land between
// the header and the list.
func TestNewListQuestionPanel(t *testing.T) {
	c := check.New(t)
	list := unison.NewPanel()
	extra := unison.NewLabel()
	extra.SetTitle("Extra")
	panel, scroll := newListQuestionPanel(promptOperation{}, "Header", list, extra)
	children := panel.Children()
	c.Equal(3, len(children), "the panel must hold the header, the extra header and the scroll panel")

	header, ok := children[0].Self.(*unison.Label)
	c.True(ok, "the first child must be the header label")
	c.Equal("Header", header.String())
	c.Equal(extra.AsPanel(), children[1], "the extra header must follow the header label")

	c.Equal(scroll.AsPanel(), children[2], "the list must be wrapped in the scroll panel returned with the panel")
	c.Equal(list, scroll.Content(), "the scroll panel must hold the list")
	c.NotNil(scroll.Border(), "the scroll panel must be outlined")
	layout, ok := scroll.LayoutData().(*unison.FlexLayoutData)
	c.True(ok, "the scroll panel must carry flex layout data")
	c.True(layout.HGrab && layout.VGrab, "the scroll panel must take up the dialog's spare room")
	for _, child := range children[:2] {
		c.Nil(child.LayoutData(), "the headers must not compete with the scroll panel for room")
	}

	panel, _ = newListQuestionPanel(promptOperation{}, "Header", unison.NewPanel())
	c.Equal(2, len(panel.Children()), "with no extra headers, only the header and the scroll panel remain")

	panel, _ = newListQuestionPanel(promptOperation{description: "Operation"}, "Header", unison.NewPanel())
	children = panel.Children()
	c.Equal(3, len(children), "an operation must add a label of its own")
	operation, ok := children[0].Self.(*unison.Label)
	c.True(ok, "the operation must come first")
	c.Equal("Operation", operation.String())
	header, ok = children[1].Self.(*unison.Label)
	c.True(ok, "the header must follow the operation")
	c.Equal("Header", header.String())
}
