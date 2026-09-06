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

// TestNewListQuestionPanel verifies the shape of the panel the list question dialogs share: the header label comes
// first, any extra header labels follow it in the order given, and the list itself sits last, inside a scroll panel
// that is the only child asked to take up the dialog's spare room. The extra headers are how the modifier prompt names
// the row being asked about, so they have to land between the header and the list rather than anywhere else.
func TestNewListQuestionPanel(t *testing.T) {
	c := check.New(t)
	list := unison.NewPanel()
	extra := unison.NewLabel()
	extra.SetTitle("Extra")
	panel := newListQuestionPanel("Header", list, extra)
	children := panel.Children()
	c.Equal(3, len(children), "the panel must hold the header, the extra header and the scroll panel")

	header, ok := children[0].Self.(*unison.Label)
	c.True(ok, "the first child must be the header label")
	c.Equal("Header", header.String())
	c.Equal(extra.AsPanel(), children[1], "the extra header must follow the header label")

	scroll, ok := children[2].Self.(*unison.ScrollPanel)
	c.True(ok, "the list must be wrapped in a scroll panel")
	c.Equal(list, scroll.Content(), "the scroll panel must hold the list")
	c.NotNil(scroll.Border(), "the scroll panel must be outlined")
	layout, ok := scroll.LayoutData().(*unison.FlexLayoutData)
	c.True(ok, "the scroll panel must carry flex layout data")
	c.True(layout.HGrab && layout.VGrab, "the scroll panel must take up the dialog's spare room")
	for _, child := range children[:2] {
		c.Nil(child.LayoutData(), "the headers must not compete with the scroll panel for room")
	}

	c.Equal(2, len(newListQuestionPanel("Header", unison.NewPanel()).Children()),
		"with no extra headers, only the header and the scroll panel remain")
}
