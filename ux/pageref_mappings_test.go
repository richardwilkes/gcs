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
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
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

// TestAsPageRefMappingsDockable verifies that the Page Reference Mappings view is found from any of the Dockable values
// that refer to it, including the embedded SettingsDockable that SettingsDockable.Setup hands to the placement code. A
// direct type assertion to *pageRefMappingsDockable misses that inner layer, which is the regression this guards
// against: RefreshPageRefMappingsView silently found nothing to refresh, leaving the view showing the previous PDF for
// a mapping until it was closed and reopened.
func TestAsPageRefMappingsDockable(t *testing.T) {
	c := check.New(t)
	d := &pageRefMappingsDockable{}
	d.Self = d
	c.Equal(d, asPageRefMappingsDockable(d), "view docked in the workspace")
	c.Equal(d, asPageRefMappingsDockable(&d.SettingsDockable), "view in a window of its own")
	other := &SettingsDockable{}
	other.Self = other
	c.Nil(asPageRefMappingsDockable(other), "some other settings view")
}

// TestPageRefMappingsSyncReflectsChangedPath verifies that syncing the Page Reference Mappings view rebuilds its rows
// from the current settings, so that pointing a key at a different PDF is reflected in the row's file name label. This
// is the update askUserForPageRefPath asks for after the user picks a new PDF for a mapping.
func TestPageRefMappingsSyncReflectsChangedPath(t *testing.T) {
	c := check.New(t)
	global := gurps.GlobalSettings()
	saved := global.PageRefs
	t.Cleanup(func() { global.PageRefs = saved })
	global.PageRefs = gurps.PageRefs{}
	global.PageRefs.Set(&gurps.PageRef{ID: "B", Path: filepath.Join("pdfs", "Basic Set.pdf"), Offset: 2})

	d := &pageRefMappingsDockable{}
	d.Self = d
	content := unison.NewPanel()
	d.initContent(content)
	c.Equal("Basic Set.pdf", pageRefMappingRowName(content, 0))

	global.PageRefs.Set(&gurps.PageRef{ID: "B", Path: filepath.Join("pdfs", "Basic Set 4th Edition.pdf"), Offset: 2})
	d.sync()
	c.Equal("Basic Set 4th Edition.pdf", pageRefMappingRowName(content, 0))
}

// pageRefMappingRowName returns the file name displayed by the given row of the Page Reference Mappings view, or an
// empty string if that row isn't present.
func pageRefMappingRowName(content *unison.Panel, row int) string {
	children := content.Children()
	i := row*5 + 4 // Each row is a trash button, ID label, offset field, edit button and file name label
	if i >= len(children) {
		return ""
	}
	if label, ok := children[i].Self.(*unison.Label); ok {
		return label.String()
	}
	return ""
}
