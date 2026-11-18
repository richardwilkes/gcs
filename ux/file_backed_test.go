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

// A dockable opened from a file is tied to that file: it is not modified, it goes by the file's name, saving it writes
// back to the file without needing to be told where, and it follows the file when the file is renamed.
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

	renamed := filepath.Join(filepath.Dir(p), "renamed"+gurps.TraitsExt)
	d.SetBackingFilePath(renamed)
	c.Equal(renamed, d.BackingFilePath(), "the dockable follows its file to its new path")
	c.Equal("renamed", d.Title())
}

// A sheet, loot sheet or template that has never been saved takes its file name from the name of its content, and only
// from its path once it has been saved. A name holding path separators or other characters a file name cannot have is
// sanitized first, as the file editors do, since the title and the Save As suggestion take the base name of that path
// and would otherwise keep only the part after the last separator.
func TestUnsavedSheetsGoByTheirNames(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	sanitized := sanitizeFileName("Warrior/Mage")
	c.NotEqual("Warrior/Mage", sanitized, "the sanitizer must change a name holding a path separator")

	entity := gurps.NewEntity()
	entity.Profile.Name = " Bob "
	sheet := NewSheet("untitled"+gurps.SheetExt, entity)
	c.True(sheet.needsSaveAsPrompt, "a new sheet must need a Save As prompt")
	c.Equal("Bob"+gurps.SheetExt, sheet.BackingFilePath())
	c.Equal("Bob", sheet.Title())
	c.Equal("Bob"+gurps.SheetExt, sheet.Tooltip())
	entity.Profile.Name = "Warrior/Mage"
	c.Equal(sanitized+gurps.SheetExt, sheet.BackingFilePath())
	c.Equal(sanitized, sheet.Title())
	c.Equal(sanitized+gurps.SheetExt, sheet.Tooltip())
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
	loot.Name = "Warrior/Mage"
	c.Equal(sanitized+gurps.LootExt, lootSheet.BackingFilePath())
	c.Equal(sanitized, lootSheet.Title())
	loot.Name = ""
	c.Equal("Unnamed Loot"+gurps.LootExt, lootSheet.BackingFilePath())
	lootSheet.markLoadedFromFile()
	c.Equal("untitled"+gurps.LootExt, lootSheet.BackingFilePath(), "a saved loot sheet must go by its path")

	data := gurps.NewTemplate()
	data.Name = "Knight"
	template := NewTemplate("untitled"+gurps.TemplatesExt, data)
	c.Equal("Knight"+gurps.TemplatesExt, template.BackingFilePath())
	c.Equal("Knight", template.Title())
	data.Name = "Warrior/Mage"
	c.Equal(sanitized+gurps.TemplatesExt, template.BackingFilePath())
	c.Equal(sanitized, template.Title())
	data.Name = ""
	c.Equal("Unnamed Template"+gurps.TemplatesExt, template.BackingFilePath())
	template.markLoadedFromFile()
	c.Equal("untitled"+gurps.TemplatesExt, template.BackingFilePath(), "a saved template must go by its path")
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

	md := filepath.Join(filepath.Dir(p), "notes.md")
	c.NoError(os.WriteFile(md, []byte("# Notes\n"), 0o600))
	readOnlyFile, err := NewMarkdownDockable(md, false, false)
	c.NoError(err)
	c.False(readOnlyFile.Modified())
	c.Equal("notes", readOnlyFile.Title())
	c.Equal(md, readOnlyFile.Tooltip())

	readOnly := NewMarkdownDockableWithContent("Read Me", "# Hello\n", false, false)
	c.False(readOnly.Modified())
	c.Equal("Read Me", readOnly.Title())
	c.Equal("", readOnly.Tooltip(), "content that is not in a file has no path to show")
}

// TestSanitizeFileName pins how a name is made into a file name: each character a platform takes for a path separator
// becomes an underscore, a name that would make a hidden file gets one ahead of its dot, and a name that is empty or
// names a folder is replaced by a word.
func TestSanitizeFileName(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name string
		want string
	}{
		{name: "Warrior Mage", want: "Warrior Mage"},
		{name: `a:b/c\d`, want: "a_b_c_d"},
		{name: ".hidden", want: "_.hidden"},
		{name: "", want: "unnamed"},
		{name: ".", want: "dot"},
		{name: "..", want: "dotdot"},
	} {
		c.Equal(tc.want, sanitizeFileName(tc.name), "%q", tc.name)
	}
}
