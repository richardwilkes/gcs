// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package datasync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestSyncToLibraryDataProcessesEachFileType verifies that SyncToLibraryData syncs every sheet, template and loot file
// it finds with its library sources and writes it back out, regardless of the case of the extension, leaves other
// files alone, and reports a file it cannot load rather than overwriting it.
func TestSyncToLibraryDataProcessesEachFileType(t *testing.T) {
	c := check.New(t)
	libFile := gurps.LibraryFile{Library: "Sync Test", Path: "Test" + gurps.TraitsExt}
	lib := library.NewLibrary(libFile.Library, "", "", libFile.Library, t.TempDir())
	libs := gurps.GlobalSettings().Libraries
	libs.Store(libFile.Library, lib)
	t.Cleanup(func() { libs.Remove(libFile.Library) })
	libTrait := gurps.NewTrait(nil, nil, false)
	libTrait.Name = "Claws"
	c.NoError(gurps.SaveTraits([]*gurps.Trait{libTrait}, filepath.Join(lib.Path(false), libFile.Path)))

	dir := t.TempDir()
	sheetPath := filepath.Join(dir, "sheet"+gurps.SheetExt)
	sheet := gurps.NewEntity()
	stale := libTrait.Clone(libFile, sheet, nil, gurps.Reference)
	stale.Name = "Claws (old)"
	sheet.Traits = []*gurps.Trait{stale}
	c.NoError(sheet.Save(sheetPath))
	upperSheetPath := filepath.Join(dir, "upper"+strings.ToUpper(gurps.SheetExt))
	c.NoError(gurps.NewEntity().Save(upperSheetPath))
	templatePath := filepath.Join(dir, "template"+gurps.TemplatesExt)
	c.NoError(gurps.NewTemplate().Save(templatePath))
	lootPath := filepath.Join(dir, "loot"+gurps.LootExt)
	c.NoError(gurps.NewLoot().Save(lootPath))
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
	sheet, err = gurps.NewEntityFromFile(os.DirFS(dir), filepath.Base(sheetPath))
	c.NoError(err)
	names := make([]string, len(sheet.Traits))
	for i, one := range sheet.Traits {
		names[i] = one.Name
	}
	c.Equal([]string{libTrait.Name}, names, "a trait that differs from its library source is brought back in line")

	broken := []byte("not json")
	c.NoError(os.WriteFile(templatePath, broken, 0o600))
	c.HasError(SyncToLibraryData(dir), "a file that fails to load is reported")
	data, err := os.ReadFile(templatePath)
	c.NoError(err)
	c.Equal(broken, data, "a file that fails to load is not overwritten")
}
