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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/uti"
)

// TestPrepareProfileForContentCache verifies that every profile field placed into the deep search content cache is
// lowercased, since the search text is lowercased before the comparison is made.
func TestPrepareProfileForContentCache(t *testing.T) {
	c := check.New(t)
	content := prepareProfileForContentCache(&gurps.Profile{
		Name:         "Conan",
		Age:          "Thirty",
		Birthday:     "January 1",
		Eyes:         "Blue",
		Hair:         "Black",
		Skin:         "Tan",
		Handedness:   "Right",
		Gender:       "Male",
		PlayerName:   "Robert",
		Title:        "King Of Aquilonia",
		Organization: "The Black Dragons",
		Religion:     "Crom",
	})
	for _, one := range []string{
		"conan",
		"thirty",
		"january 1",
		"blue",
		"black",
		"tan",
		"right",
		"male",
		"robert",
		"king of aquilonia",
		"the black dragons",
		"crom",
	} {
		c.True(strings.Contains(content, one), "expected %q in the cached content, got %q", one, content)
	}
}

// TestDeepSearchOfSheetProfileIsCaseInsensitive verifies that a deep search of a character sheet matches profile
// fields regardless of case. The search text is lowercased before comparison, so a character named "Conan" must be
// found when searching for "conan", even though the file name doesn't match.
func TestDeepSearchOfSheetProfileIsCaseInsensitive(t *testing.T) {
	c := check.New(t)
	RegisterGCSFileTypes() // The deep search resolves the file's extension through the file type registry.
	dir := t.TempDir()
	entity := gurps.NewEntity()
	entity.Profile.Name = "Conan"
	entity.Profile.PlayerName = "Robert"
	entity.Profile.Organization = "The Black Dragons"
	fileName := "sheet1" + gurps.SheetExt
	c.NoError(entity.Save(filepath.Join(dir, fileName)))

	lib := gurps.NewLibrary("Test", "test", "", "test", dir)
	n := &Navigator{
		searchIndex: -1,
		deepSearch:  map[string]bool{gurps.SheetExt: true},
	}
	node := NewFileNode(lib, fileName, nil)
	for _, one := range []struct {
		name string
		text string
	}{
		{name: "character name", text: "conan"},
		{name: "partial character name", text: "cona"},
		{name: "player name", text: "robert"},
		{name: "organization", text: "the black dragons"},
	} {
		n.searchResult = nil
		n.search(one.text, []*NavigatorNode{node})
		c.Equal(1, len(n.searchResult), one.name)
	}

	// Something that isn't present must still not match.
	n.searchResult = nil
	n.search("belit", []*NavigatorNode{node})
	c.Equal(0, len(n.searchResult))
}

// TestDeepSearchCachesMarkdownContent verifies that the markdown branch of the deep search content loader populates the
// content cache like every other branch does. Without it, each keystroke in the search field re-reads and re-lowercases
// every markdown file in every library.
func TestDeepSearchCachesMarkdownContent(t *testing.T) {
	c := check.New(t)
	RegisterGCSFileTypes() // The deep search resolves the file's extension through the file type registry.
	ext := uti.Markdown.Extensions[0]
	dir := t.TempDir()
	fileName := "notes" + ext
	p := filepath.Join(dir, fileName)
	c.NoError(os.WriteFile(p, []byte("# Conan The Barbarian\n"), 0o600))

	lib := gurps.NewLibrary("Test", "test", "", "test", dir)
	n := &Navigator{
		searchIndex: -1,
		deepSearch:  map[string]bool{ext: true},
	}
	node := NewFileNode(lib, fileName, nil)
	n.search("barbarian", []*NavigatorNode{node})
	c.Equal(1, len(n.searchResult))

	// The lowercased content must now be cached under the file's full path.
	cached, ok := n.contentCache[p]
	c.True(ok, "expected the markdown content to be cached for %q", p)
	c.Equal("# conan the barbarian\n", cached)

	// Removing the file must not change the outcome of a subsequent search, proving the cache is what gets consulted.
	c.NoError(os.Remove(p))
	n.searchResult = nil
	n.search("barbarian", []*NavigatorNode{node})
	c.Equal(1, len(n.searchResult))
}

// TestToggleFavoritesKeysOnFullPath verifies that the de-duplication done while toggling favorites is keyed on the full
// path on disk rather than the library-relative path. Two libraries can each hold the same relative path, and favorites
// are tracked per library, so both must be toggled.
func TestToggleFavoritesKeysOnFullPath(t *testing.T) {
	c := check.New(t)
	relPath := filepath.Join("Notes", "foo.not")
	lib1 := gurps.NewLibrary("One", "one", "", "one", t.TempDir())
	lib2 := gurps.NewLibrary("Two", "two", "", "two", t.TempDir())
	rows := []*NavigatorNode{
		NewFileNode(lib1, relPath, nil),
		NewFileNode(lib2, relPath, nil),
	}
	c.True(toggleFavorites(rows))
	c.Equal([]string{relPath}, lib1.Favorites())
	c.Equal([]string{relPath}, lib2.Favorites())

	// A repeated reference to the same file within one call must only toggle once, leaving it a favorite rather than
	// toggling it back off.
	lib3 := gurps.NewLibrary("Three", "three", "", "three", t.TempDir())
	c.True(toggleFavorites([]*NavigatorNode{
		NewFileNode(lib3, relPath, nil),
		NewFileNode(lib3, relPath, nil),
	}))
	c.Equal([]string{relPath}, lib3.Favorites())

	// A selection holding nothing but library nodes toggles nothing.
	c.False(toggleFavorites([]*NavigatorNode{NewLibraryNode(nil, lib1)}))
	c.Equal([]string{relPath}, lib1.Favorites())
}

// TestTrimmedNameIsValid verifies that the name validation shared by the rename and new-folder dialogs both trims the
// name before judging it and reports back the trimmed name, so that callers build the path that was actually validated.
func TestTrimmedNameIsValid(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name    string
		in      string
		trimmed string
		valid   bool
	}{
		{name: "plain", in: "Notes", trimmed: "Notes", valid: true},
		{name: "surrounding whitespace is trimmed off", in: "  Notes  ", trimmed: "Notes", valid: true},
		{name: "empty", in: "", trimmed: "", valid: false},
		{name: "whitespace only", in: "   ", trimmed: "", valid: false},
		{name: "leading dot", in: ".hidden", trimmed: ".hidden", valid: false},
		{name: "leading dot after trimming", in: "  .hidden", trimmed: ".hidden", valid: false},
		{name: "path separator", in: "a/b", trimmed: "a/b", valid: false},
		{name: "path separator revealed by trimming", in: " a/b ", trimmed: "a/b", valid: false},
		{name: "reserved windows name", in: "con", trimmed: "con", valid: false},
		{name: "reserved windows name with whitespace", in: " CON ", trimmed: "CON", valid: false},
	} {
		trimmed, valid := trimmedNameIsValid(one.in)
		c.Equal(one.trimmed, trimmed, one.name)
		c.Equal(one.valid, valid, one.name)
	}
}

// TestPathsBuiltFromTheValidatedName verifies that the paths the rename and new-folder dialogs build are derived from
// the same trimmed name that trimmedNameIsValid judges, so a name entered with surrounding whitespace can't validate as
// one path and then create or rename to another.
func TestPathsBuiltFromTheValidatedName(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	for _, one := range []struct {
		name string
		in   string
		want string
	}{
		{name: "plain", in: "Notes", want: "Notes"},
		{name: "leading whitespace", in: "   Notes", want: "Notes"},
		{name: "trailing whitespace", in: "Notes   ", want: "Notes"},
		{name: "surrounding whitespace", in: "  Notes  ", want: "Notes"},
		{name: "interior whitespace is preserved", in: "  My Notes  ", want: "My Notes"},
	} {
		trimmed, valid := trimmedNameIsValid(one.in)
		c.True(valid, one.name)
		c.Equal(filepath.Join(dir, trimmed), newFolderPath(dir, one.in), one.name)
		c.Equal(filepath.Join(dir, one.want), newFolderPath(dir, one.in), one.name)

		oldPath := filepath.Join(dir, "old"+gurps.NotesExt)
		c.Equal(filepath.Join(dir, trimmed+gurps.NotesExt), renamedPath(oldPath, one.in), one.name)
		c.Equal(filepath.Join(dir, one.want+gurps.NotesExt), renamedPath(oldPath, one.in), one.name)
	}
}

// TestNewFolderRejectsCollisionWithWhitespacePaddedName verifies that the existence check the new-folder dialog makes
// and the directory it would create refer to the same path, so padding a name with whitespace can't slip a duplicate
// past validation.
func TestNewFolderRejectsCollisionWithWhitespacePaddedName(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	c.NoError(os.Mkdir(filepath.Join(dir, "Notes"), 0o750))

	// The name is padded, but the collision check must still find the existing directory.
	p := newFolderPath(dir, "  Notes  ")
	_, err := os.Stat(p)
	c.NoError(err, "the padded name must resolve to the existing directory")

	// And creating it must be what fails, rather than quietly producing a second, differently-named directory.
	c.HasError(os.Mkdir(p, 0o750))
	entries, err := os.ReadDir(dir)
	c.NoError(err)
	c.Equal(1, len(entries))
	c.Equal("Notes", entries[0].Name())
}
