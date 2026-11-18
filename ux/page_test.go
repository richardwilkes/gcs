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
	"github.com/richardwilkes/gcs/v5/model/paper"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
)

// TestPageNumberingCountsOnlyPages verifies that a page's number and count come only from the pages sharing its
// parent, so that anything else beside them -- the layout editor's overlay, say -- neither shifts the numbers nor
// swaps the footer's halves around, and that a page with no parent is still the first of one.
func TestPageNumberingCountsOnlyPages(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	lone := NewPage(entity)
	number, count := lone.pageNumbering()
	c.Equal(1, number, "a page with no parent is the first")
	c.Equal(1, count, "of one")

	parent := unison.NewPanel()
	pages := []*Page{NewPage(entity), NewPage(entity), NewPage(entity)}
	parent.AddChild(unison.NewPanel())
	for _, page := range pages {
		parent.AddChild(page)
		parent.AddChild(unison.NewPanel())
	}
	for i, page := range pages {
		number, count = page.pageNumbering()
		c.Equal(i+1, number, "page %d must be numbered by its place among the pages alone", i+1)
		c.Equal(len(pages), count, "the count must be the number of pages, not of children")
	}
}

// TestPageBorderFollowsThePageSettings verifies that a page laid out after its margins were changed takes each new
// margin as its border on that edge, with room for the footer still kept below the bottom one.
func TestPageBorderFollowsThePageSettings(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	page := NewPage(entity)
	before := page.Border().Insets()
	settings := entity.PageSettings()
	footer := before.Bottom - settings.BottomMargin.Pixels()
	inches := func(length float64) paper.Length { return paper.Length{Length: length, Units: paper.Inch} }
	settings.TopMargin = inches(1)
	settings.LeftMargin = inches(0.75)
	settings.BottomMargin = inches(0.5)
	settings.RightMargin = inches(1.25)
	want := geom.Insets{
		Top:    inches(1).Pixels(),
		Left:   inches(0.75).Pixels(),
		Bottom: inches(0.5).Pixels() + footer,
		Right:  inches(1.25).Pixels(),
	}
	c.NotEqual(want, before, "the margins must start out otherwise, or the check below means nothing")
	page.Sizes(geom.Size{})
	c.Equal(want, page.Border().Insets(), "the border must follow the margins the page is laid out with")
}
