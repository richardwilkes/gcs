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
)

// TestSplitMarkdownPageRef verifies that a markdown page reference is split into its file path and anchor correctly. The
// key regression this guards against is a reference containing a '#' anchor: the anchor must be removed before the ".md"
// extension is appended, otherwise the path becomes "File#Section.md" and never resolves to a real file.
func TestSplitMarkdownPageRef(t *testing.T) {
	c := check.New(t)
	const (
		homeMD = "Home.md"
		anchor = "New"
	)
	for _, tc := range []struct {
		name       string
		ref        string
		wantPath   string
		wantAnchor string
	}{
		{name: "no extension, no anchor", ref: "Home", wantPath: homeMD, wantAnchor: ""},
		{name: "extension, no anchor", ref: homeMD, wantPath: homeMD, wantAnchor: ""},
		{name: "no extension, with anchor", ref: "Home#" + anchor, wantPath: homeMD, wantAnchor: anchor},
		{name: "extension, with anchor", ref: "Home.md#" + anchor, wantPath: homeMD, wantAnchor: anchor},
		{name: "path with subdir and anchor", ref: "Guide/Intro#Getting Started", wantPath: "Guide/Intro.md", wantAnchor: "Getting Started"},
		{name: "uppercase extension preserved", ref: "Home.MD#" + anchor, wantPath: "Home.MD", wantAnchor: anchor},
		{name: "empty ref", ref: "", wantPath: "", wantAnchor: ""},
		{name: "anchor only", ref: "#" + anchor, wantPath: "", wantAnchor: anchor},
		{name: "trailing hash yields empty anchor", ref: "Home#", wantPath: homeMD, wantAnchor: ""},
		// URL-encoded spaces in the path are decoded so encoded and non-encoded references resolve to the same file.
		{name: "encoded spaces in path", ref: "User%20Guide/Scripting%20Guide#code", wantPath: "User Guide/Scripting Guide.md", wantAnchor: "code"},
		{name: "encoded spaces without anchor", ref: "User%20Guide/Home", wantPath: "User Guide/Home.md", wantAnchor: ""},
	} {
		t.Run(tc.name, func(_ *testing.T) {
			path, anchor := splitMarkdownPageRef(tc.ref)
			c.Equal(tc.wantPath, path, tc.name)
			c.Equal(tc.wantAnchor, anchor, tc.name)
		})
	}
}
