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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// A dockable opened from a file is tied to that file: it is not modified, it goes by the file's name, and saving it
// writes back to the file without needing to be told where.
func TestOpenDockableFromFile(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Alertness"
	p := filepath.Join(t.TempDir(), "library"+gurps.TraitsExt)
	c.NoError(gurps.SaveTraits([]*gurps.Trait{trait}, p))

	dockable, err := NewTraitTableDockableFromFile(p)
	c.NoError(err)
	d, ok := dockable.(*TableDockable[*gurps.Trait])
	c.True(ok)
	c.False(d.needsSaveAsPrompt, "a dockable opened from a file must not need a Save As prompt")
	c.False(d.Modified(), "a freshly opened file must not be reported as modified")
	c.Equal(p, d.BackingFilePath())
	c.Equal("library", d.Title())
	c.Equal(p, d.Tooltip())
	c.Equal(filePrefix+p, d.DockKey())

	d.provider.RootData()[0].Name = "Acute Vision"
	c.True(d.Modified(), "a change to the content must be reported as modified")
	c.True(d.save(false), "saving to the file the content came from must succeed")
	c.False(d.Modified(), "the content must not be reported as modified once it has been saved")
	traits := loadSavedFile(t, c, p, gurps.NewTraitsFromFile)
	c.Equal(1, len(traits))
	c.Equal("Acute Vision", traits[0].Name, "the save must have written the changed content to the file")

	_, err = NewTraitTableDockableFromFile(filepath.Join(t.TempDir(), "missing"+gurps.TraitsExt))
	c.HasError(err, "opening a file that does not exist must fail")
}

// A sheet that has never been saved takes its file name from the name of its content, and only from its path once it
// has been saved.
func TestUnsavedSheetsGoByTheirNames(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(func() { registerActions() })

	entity := gurps.NewEntity()
	entity.Profile.Name = " Bob "
	sheet := NewSheet("untitled"+gurps.SheetExt, entity)
	c.True(sheet.needsSaveAsPrompt, "a new sheet must need a Save As prompt")
	c.Equal("Bob"+gurps.SheetExt, sheet.BackingFilePath())
	c.Equal("Bob", sheet.Title())
	c.Equal("Bob"+gurps.SheetExt, sheet.Tooltip())
	entity.Profile.Name = ""
	c.Equal("Unnamed Character"+gurps.SheetExt, sheet.BackingFilePath())
	sheet.markLoadedFromFile()
	c.Equal("untitled"+gurps.SheetExt, sheet.BackingFilePath(), "a saved sheet must go by its path")
	c.Equal("untitled", sheet.Title())

	loot := gurps.NewLoot()
	loot.Name = "Dragon Hoard"
	lootSheet := NewLootSheet("untitled"+gurps.LootExt, loot)
	c.Equal("Dragon Hoard"+gurps.LootExt, lootSheet.BackingFilePath())
	c.Equal("Dragon Hoard", lootSheet.Title())
	loot.Name = ""
	c.Equal("Unnamed Loot"+gurps.LootExt, lootSheet.BackingFilePath())
	lootSheet.markLoadedFromFile()
	c.Equal("untitled"+gurps.LootExt, lootSheet.BackingFilePath(), "a saved loot sheet must go by its path")
}

// A dockable that only views a file is tied to it by name and key, but never has anything to save.
func TestFileViewerIsNeverModified(t *testing.T) {
	c := check.New(t)
	p := filepath.Join(t.TempDir(), "picture.svg")
	c.NoError(os.WriteFile(p, []byte(sampleSVG), 0o600))
	dockable, err := NewImageDockable(p)
	c.NoError(err)
	c.False(dockable.Modified())
	c.Equal("picture", dockable.Title())
	c.Equal(p, dockable.Tooltip())
	keyed, ok := dockable.(KeyedDockable)
	c.True(ok, "an image dockable must have a dock key, so that it can be restored with the workspace")
	c.Equal(filePrefix+p, keyed.DockKey())

	readOnly := NewMarkdownDockableWithContent("Read Me", "# Hello\n", false, false)
	c.False(readOnly.Modified())
	c.Equal("Read Me", readOnly.Title())
	c.Equal("", readOnly.Tooltip(), "content that is not in a file has no path to show")
}

// Saving an edited markdown file writes the edited content back to it and clears the modified state, which is what the
// hash over its content is for.
func TestMarkdownDockableSaves(t *testing.T) {
	c := check.New(t)
	p := filepath.Join(t.TempDir(), "notes.md")
	c.NoError(os.WriteFile(p, []byte("# Title\n"), 0o600))
	dockable, err := NewMarkdownDockable(p, true, false)
	c.NoError(err)
	d, ok := dockable.(*MarkdownDockable)
	c.True(ok)
	d.content += "\nMore text.\n"
	c.True(d.Modified())
	c.True(d.save(false))
	c.False(d.Modified(), "the content must not be reported as modified once it has been saved")
	data, err := os.ReadFile(p)
	c.NoError(err)
	c.Equal("# Title\n\nMore text.\n", string(data))
}
