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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// TestSideBarYieldsToNarrowView verifies that a shown side bar takes no more than 40% of a view too narrow for its
// preferred width, leaving the rest to the content, and returns to its preferred width once the view widens again.
func TestSideBarYieldsToNarrowView(t *testing.T) {
	c := check.New(t)
	container := unison.NewPanel()
	container.SetLayout(&unison.FlexLayout{Columns: 1, HAlign: align.Fill, VAlign: align.Fill})
	content := unison.NewPanel()
	content.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, VAlign: align.Fill, HGrab: true, VGrab: true})
	container.AddChild(content)
	s := newSideBar(unison.NewPanel())
	s.toggle(container, content)
	layOut := func(width float32) (side, divider, rest float32) {
		container.SetFrameRect(geom.NewRect(0, 0, width, 400))
		container.ValidateLayout()
		return s.scroll.FrameRect().Width, s.divider.FrameRect().Width, content.FrameRect().Width
	}

	const narrow = 300
	side, divider, rest := layOut(narrow)
	c.True(side <= narrow*0.4, "the side bar is %v wide in a %v wide view", side, narrow)
	c.Equal(narrow-side-divider, rest, "the content has the rest of the view")

	side, _, _ = layOut(1000)
	c.Equal(float32(sideBarInitialWidth), side, "the side bar is back at its preferred width in a wide view")
}
