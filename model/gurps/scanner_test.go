// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestScanForNamedFileSetsIgnoresExtensionCase verifies that files are matched without regard to the case of their
// extension. The extension map is built from lowercased extensions, but the file's own extension was looked up as-is,
// so anything with an upper or mixed-case extension was silently skipped from the settings, ancestry, calendar and
// name-generator scans.
func TestScanForNamedFileSetsIgnoresExtensionCase(t *testing.T) {
	c := check.New(t)
	fileSystem := fstest.MapFS{
		"Settings/Human.ancestry": {Data: []byte("{}")},
		"Settings/Elf.ANCESTRY":   {Data: []byte("{}")},
		"Settings/Dwarf.Ancestry": {Data: []byte("{}")},
		"Settings/Notes.txt":      {Data: []byte("{}")},
	}
	refs := scanForNamedFileSets(fileSystem, "Settings", []string{AncestryExt}, true, make(map[string]bool))
	names := make([]string, 0, len(refs))
	for _, ref := range refs {
		names = append(names, ref.Name)
	}
	c.Equal([]string{"Dwarf", "Elf", "Human"}, names, "extensions match without regard to case")
}
