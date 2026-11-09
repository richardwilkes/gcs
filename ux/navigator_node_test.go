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
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/unison"
)

// newLibraryWithVersionOnDisk returns a library whose release file records the given version.
func newLibraryWithVersionOnDisk(t *testing.T, version string) *library.Library {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "release.txt"), []byte(version+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return library.NewLibrary("Test", "someone", "", "repo", dir)
}

// TestNavigatorNodeTagsLibraryNeedingNewerApp verifies that the row of a library written for a newer version of the app
// says so with a tag beside its title, which goes on showing the version on disk.
func TestNavigatorNodeTagsLibraryNeedingNewerApp(t *testing.T) {
	c := check.New(t)
	RegisterKnownFileTypes() // The row looks its folder icon up in the registry
	version := strconv.Itoa(jio.CurrentDataVersion+1) + ".0.0"
	node := NewLibraryNode(nil, newLibraryWithVersionOnDisk(t, version), nil)
	cell := node.ColumnCell(0, 0, unison.Black, unison.White, false, false, false).AsPanel()
	tag, ok := uxtest.FirstPanelOfType[*unison.Tag](cell)
	c.True(ok, "the row must carry a tag")
	if ok {
		c.Equal(i18n.Text("Requires a newer version of %s", xos.AppName), tag.Text.String())
	}
	c.Equal([]string{"Test " + filterVersion(version)}, uxtest.LabelTexts(cell))
}
