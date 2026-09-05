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
	"os"
	"path/filepath"
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
	refs := scanForNamedFileSets(fileSystem, "Settings", "", []string{AncestryExt}, true, make(map[string]bool))
	names := make([]string, 0, len(refs))
	for _, ref := range refs {
		names = append(names, ref.Name)
	}
	c.Equal([]string{"Dwarf", "Elf", "Human"}, names, "extensions match without regard to case")
}

// TestScanForNamedFileSetsRecordsDiskPath verifies that a file found in a library carries the absolute path of the file
// on disk, so that an editor can save back to where the file came from, while a built-in file, which is embedded in the
// application and has no disk location, leaves it empty.
func TestScanForNamedFileSetsRecordsDiskPath(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	libs := NewLibraries()
	c.NoError(libs.Master().SetPath(filepath.Join(dir, "master")))
	userPath := filepath.Join(dir, "user")
	c.NoError(libs.User().SetPath(userPath))
	ancestriesDir := filepath.Join(userPath, SettingsDirName, AncestriesDirName)
	c.NoError(os.MkdirAll(ancestriesDir, 0o750))
	c.NoError(os.WriteFile(filepath.Join(ancestriesDir, "Elf.ancestry"), []byte(`{"version":5,"name":"Elf"}`), 0o640))
	var userRef, builtInRef *NamedFileRef
	for _, set := range ScanForNamedFileSets(embeddedFS, "embedded_data", true, libs, AncestryExt) {
		for _, ref := range set.List {
			switch {
			case set.Name == libs.User().Data().Title && ref.Name == "Elf":
				userRef = ref
			case set.Name == "Built-in" && ref.Name == DefaultAncestry:
				builtInRef = ref
			}
		}
	}
	c.NotNil(userRef, "the user library's ancestry was found")
	if userRef != nil {
		c.Equal("Elf", userRef.Name)
		c.Equal("Settings/Ancestries/Elf.ancestry", userRef.FilePath)
		c.Equal(filepath.Join(userPath, SettingsDirName, AncestriesDirName, "Elf.ancestry"), userRef.DiskPath)
	}
	c.NotNil(builtInRef, "the built-in ancestry was found")
	if builtInRef != nil {
		c.Equal("", builtInRef.DiskPath, "embedded files have no disk location")
	}
}
