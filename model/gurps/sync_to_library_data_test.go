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
	"strings"
	"testing"
	"time"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestLibrarySyncersCoverSyncedFileTypes verifies that the sheet, template and loot file types each have a loader in
// the table SyncToLibraryData dispatches through, keyed the way the dispatch looks them up.
func TestLibrarySyncersCoverSyncedFileTypes(t *testing.T) {
	c := check.New(t)
	for _, ext := range []string{SheetExt, TemplatesExt, LootExt} {
		_, exists := librarySyncers[ext]
		c.True(exists, "%s has a loader entry", ext)
	}
	for ext := range librarySyncers {
		c.Equal(strings.ToLower(ext), ext, "%s is keyed in lowercase, which is what the dispatch looks up", ext)
	}
}

// TestSyncToLibraryDataProcessesEachFileType verifies that SyncToLibraryData loads and writes back every sheet,
// template and loot file it finds, regardless of the case of the extension, leaves other files alone, and reports a
// file it cannot load rather than overwriting it.
func TestSyncToLibraryDataProcessesEachFileType(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	sheetPath := filepath.Join(dir, "sheet"+SheetExt)
	c.NoError(newPopulatedEntity().Save(sheetPath))
	upperSheetPath := filepath.Join(dir, "upper"+strings.ToUpper(SheetExt))
	c.NoError(NewEntity().Save(upperSheetPath))
	templatePath := filepath.Join(dir, "template"+TemplatesExt)
	c.NoError(NewTemplate().Save(templatePath))
	lootPath := filepath.Join(dir, "loot"+LootExt)
	c.NoError(NewLoot().Save(lootPath))
	notesPath := filepath.Join(dir, "notes.txt")
	c.NoError(os.WriteFile(notesPath, []byte("{}"), 0o600))

	// Push every file's modification time into the past so that a rewrite is observable.
	past := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	synced := []string{sheetPath, upperSheetPath, templatePath, lootPath}
	for _, p := range append(synced, notesPath) {
		c.NoError(os.Chtimes(p, past, past))
	}

	c.NoError(SyncToLibraryData(dir))

	for _, p := range synced {
		info, err := os.Stat(p)
		c.NoError(err)
		c.True(info.ModTime().After(past), "%s is written back out", filepath.Base(p))
	}
	info, err := os.Stat(notesPath)
	c.NoError(err)
	c.Equal(past, info.ModTime().UTC(), "a file of an unrelated type is left untouched")

	broken := []byte("not json")
	c.NoError(os.WriteFile(templatePath, broken, 0o600))
	c.HasError(SyncToLibraryData(dir), "a file that fails to load is reported")
	data, err := os.ReadFile(templatePath)
	c.NoError(err)
	c.Equal(broken, data, "a file that fails to load is not overwritten")
}
