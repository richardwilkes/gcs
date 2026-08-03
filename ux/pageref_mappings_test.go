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
	for _, tc := range []struct {
		name       string
		ref        string
		wantPath   string
		wantAnchor string
	}{
		{name: "no extension, no anchor", ref: "Home", wantPath: "Home.md", wantAnchor: ""},
		{name: "extension, no anchor", ref: "Home.md", wantPath: "Home.md", wantAnchor: ""},
		{name: "no extension, with anchor", ref: "Home#New", wantPath: "Home.md", wantAnchor: "New"},
		{name: "extension, with anchor", ref: "Home.md#New", wantPath: "Home.md", wantAnchor: "New"},
		{name: "path with subdir and anchor", ref: "Guide/Intro#Getting Started", wantPath: "Guide/Intro.md", wantAnchor: "Getting Started"},
		{name: "uppercase extension preserved", ref: "Home.MD#New", wantPath: "Home.MD", wantAnchor: "New"},
		{name: "empty ref", ref: "", wantPath: "", wantAnchor: ""},
		{name: "anchor only", ref: "#New", wantPath: "", wantAnchor: "New"},
		{name: "trailing hash yields empty anchor", ref: "Home#", wantPath: "Home.md", wantAnchor: ""},
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
