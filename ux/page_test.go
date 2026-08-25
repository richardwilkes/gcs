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
)

// TestPageNumberingCountsOnlyPages verifies that a page's number and the count its footer reports come from the pages
// that share its parent alone, so that anything else put beside them -- the layout editor's overlay, say -- neither
// shifts the numbers nor swaps the footer's halves around, and that a page with no parent still calls itself the first
// of one.
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
