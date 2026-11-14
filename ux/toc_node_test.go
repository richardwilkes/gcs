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
	"github.com/richardwilkes/toolbox/v2/tid"
)

// TestNewTOCGivesEachBookmarkALastingIDOfItsOwn verifies that each bookmark of a PDF's table of contents gets an ID of
// its own, since the rows are selected and opened by ID, even one with the same title and page as another, as a part's
// bookmark and that of the chapter it opens with may have; and that a bookmark gets the same ID each time the table of
// contents is built, since whether its row is open is recorded by ID.
func TestNewTOCGivesEachBookmarkALastingIDOfItsOwn(t *testing.T) {
	c := check.New(t)
	const pdfPath = "/does/not/exist/duplicate-bookmarks.pdf"
	settings := gurps.GlobalSettings()
	t.Cleanup(func() { delete(settings.PDFs, pdfPath) })
	d := &PDFDockable{path: pdfPath}
	toc := []*PDFTableOfContents{
		{Title: "Skills", PageNumber: 40, Children: []*PDFTableOfContents{
			{Title: "Skills", PageNumber: 40},
			{Title: "Skill List", PageNumber: 41},
		}},
		{Title: "Skills", PageNumber: 40},
		{Title: "Skills", PageNumber: 41},
	}
	// ids builds the table of contents and returns the IDs of its rows, each parent's ahead of its children's.
	ids := func() []tid.TID {
		var all []tid.TID
		var collect func(nodes []*tocNode)
		collect = func(nodes []*tocNode) {
			for _, node := range nodes {
				all = append(all, node.ID())
				collect(node.children)
			}
		}
		collect(newTOC(d, nil, toc))
		return all
	}
	first := ids()
	c.Equal(5, len(first), "every bookmark must have a row")
	seen := make(map[tid.TID]int)
	for i, id := range first {
		if j, dup := seen[id]; dup {
			c.True(false, "row %d must have an ID of its own, but shares row %d's", i, j)
		}
		seen[id] = i
	}
	c.Equal(first, ids(), "building the table of contents again must give each bookmark the same ID")
}
