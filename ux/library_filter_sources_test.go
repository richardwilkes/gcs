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
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// testNavigatorFolder returns the listing entry of the folder at the path, holding the files at the other paths, all
// slash-separated and relative to the library.
func testNavigatorFolder(folder string, files ...string) *navigatorEntry {
	entry := &navigatorEntry{path: folder, isDir: true}
	for _, one := range files {
		entry.entries = append(entry.entries, &navigatorEntry{path: one})
	}
	return entry
}

// TestResolveLibraryFilterFiles verifies that a source set resolves against the Library Explorer's listing to the
// files of the list type it covers: every one for a nil set, otherwise those named by a file entry, beneath a folder
// entry (a folder whose name merely starts the same is not beneath it) or in a whole-library entry, in library order
// and then natural path order, with the extension compared without regard to case; that a library no longer registered
// is passed over; that the entries naming nothing are reported; and that the File column shows a file's name, widened
// to its path within its library and then to its library's title as far as needed to tell it from the others.
func TestResolveLibraryFilterFiles(t *testing.T) {
	c := check.New(t)
	master, user := uxtest.UseTestLibraries(t, c)
	gone := library.NewLibrary("Gone", "someone", "", "gone", filepath.Join(t.TempDir(), "gone"))
	// Listed out of natural order, as a listing that puts folders first would list them.
	skills := testNavigatorFolder("Skills", "Skills/Spell 10.skl", "Skills/Spell 2.skl", "Skills/Basic Set.skl")
	skills.entries = append(skills.entries, testNavigatorFolder("Skills/Deeper", "Skills/Deeper/Deep.skl"))
	listing := &navigatorListing{libraries: []libraryListing{
		{library: master, entries: []*navigatorEntry{
			skills,
			testNavigatorFolder("SkillsExtra", "SkillsExtra/Extra.skl"),
			{path: "Traits.adq"},
			{path: "Upper.SKL"},
		}},
		{library: gone, entries: []*navigatorEntry{{path: "Gone.skl"}}},
		{library: user, entries: []*navigatorEntry{
			{path: "Basic Set.skl"},
			testNavigatorFolder("Skills", "Skills/Basic Set.skl"),
		}},
	}}

	m := master.Key()
	u := user.Key()
	entry := func(key, p string) gurps.LibraryFile { return gurps.LibraryFile{Library: key, Path: p} }
	allSkills := []gurps.LibraryFile{
		entry(m, "Skills/Basic Set.skl"),
		entry(m, "Skills/Deeper/Deep.skl"),
		entry(m, "Skills/Spell 2.skl"),
		entry(m, "Skills/Spell 10.skl"),
		entry(m, "SkillsExtra/Extra.skl"),
		entry(m, "Upper.SKL"),
		entry(u, "Basic Set.skl"),
		entry(u, "Skills/Basic Set.skl"),
	}
	for _, tc := range []struct {
		name        string
		set         *gurps.LibrarySourceSet
		wantFiles   []gurps.LibraryFile
		wantMissing []gurps.LibraryFile
	}{
		{name: "every file of the type", wantFiles: allSkills},
		{
			name:      "a file",
			set:       &gurps.LibrarySourceSet{Entries: []gurps.LibraryFile{entry(m, "Skills/Spell 2.skl")}},
			wantFiles: []gurps.LibraryFile{entry(m, "Skills/Spell 2.skl")},
		},
		{
			name: "a folder",
			set:  &gurps.LibrarySourceSet{Entries: []gurps.LibraryFile{entry(m, "Skills")}},
			wantFiles: []gurps.LibraryFile{
				entry(m, "Skills/Basic Set.skl"),
				entry(m, "Skills/Deeper/Deep.skl"),
				entry(m, "Skills/Spell 2.skl"),
				entry(m, "Skills/Spell 10.skl"),
			},
		},
		{
			name:      "a whole library",
			set:       &gurps.LibrarySourceSet{Entries: []gurps.LibraryFile{entry(u, "")}},
			wantFiles: []gurps.LibraryFile{entry(u, "Basic Set.skl"), entry(u, "Skills/Basic Set.skl")},
		},
		{
			name: "entries that name nothing",
			set: &gurps.LibrarySourceSet{Entries: []gurps.LibraryFile{
				entry(m, "Skills/Missing.skl"),
				entry(m, "Upper.SKL"),
				entry(gone.Key(), ""),
				entry(u, "Skill"),
			}},
			wantFiles:   []gurps.LibraryFile{entry(m, "Upper.SKL")},
			wantMissing: []gurps.LibraryFile{entry(m, "Skills/Missing.skl"), entry(gone.Key(), ""), entry(u, "Skill")},
		},
	} {
		files, missing := resolveLibraryFilterFiles(listing, gurps.GlobalSettings().Libraries, gurps.SkillsExt, tc.set)
		got := make([]gurps.LibraryFile, len(files))
		for i, f := range files {
			got[i] = f.LibraryFile
		}
		c.Equal(tc.wantFiles, got, tc.name)
		c.Equal(tc.wantMissing, missing, tc.name)
	}

	files, _ := resolveLibraryFilterFiles(listing, gurps.GlobalSettings().Libraries, gurps.SkillsExt, nil)
	displays := make([]string, len(files))
	for i, f := range files {
		displays[i] = f.display
	}
	c.Equal([]string{
		master.Data().Title + ": Skills/Basic Set",
		"Deep",
		"Spell 2",
		"Spell 10",
		"Extra",
		"Upper",
		"Basic Set",
		user.Data().Title + ": Skills/Basic Set",
	}, displays, "the names shown in the File column")
	c.Equal(filepath.Join(master.Path(false), "Skills", "Deeper", "Deep.skl"), files[1].fullPath)
	c.Equal(master.Data().Title, files[1].libTitle)

	files, missing := resolveLibraryFilterFiles(nil, gurps.GlobalSettings().Libraries, gurps.SkillsExt, nil)
	c.Equal(0, len(files), "there are no files to resolve to before the libraries have been listed")
	c.Equal(0, len(missing))
}

// testNavigatorRow returns a Library Explorer row of the kind for the slash-separated path within the library, beneath
// the parent, holding the children. It is built by hand, since the constructors record a folder's path in the global
// settings.
func testNavigatorRow(kind byte, lib *library.Library, p string, parent *NavigatorNode,
	children ...*NavigatorNode,
) *NavigatorNode {
	row := &NavigatorNode{id: tid.MustNewTID(kind), path: p, library: lib, parent: parent, children: children}
	for _, child := range children {
		child.parent = row
	}
	return row
}

// TestLibraryFilterSelection verifies which list types the Library Explorer offers to open a library filter on for the
// selected rows, and the entries they stand for: selected files decide the type, and are offered only when they are
// all list files of one type, alongside the selected folders; libraries and folders alone offer every type with a file
// beneath them, and nothing when there is none; a library stands for all of it; and a row shown under Favorites is the
// same entry as under its library, the Favorites row itself standing for nothing.
func TestLibraryFilterSelection(t *testing.T) {
	c := check.New(t)
	_, user := uxtest.UseTestLibraries(t, c)
	file := func(p string) *NavigatorNode { return testNavigatorRow(kinds.NavigatorFile, user, p, nil) }
	folder := func(p string, children ...*NavigatorNode) *NavigatorNode {
		return testNavigatorRow(kinds.NavigatorDirectory, user, p, nil, children...)
	}
	skill := file("Skills/Basic.skl")
	skills := folder("Skills", skill)
	mixed := folder("Mixed", file("Mixed/Fire.spl"), file("Mixed/Fast.ADQ"),
		folder("Mixed/Deeper", file("Mixed/Deeper/Sword.eqp")))
	docs := folder("Docs", file("Docs/Read Me.md"))
	traits := file("Traits.adq")
	readMe := file("Read Me.md")
	lib := testNavigatorRow(kinds.NavigatorLibrary, user, "", nil, skills, mixed, docs, traits, readMe)
	favoriteSkill := file("Skills/Basic.skl")
	favorites := testNavigatorRow(kinds.NavigatorFavorites, nil, "", nil, favoriteSkill)

	u := user.Key()
	traitsKey := gurps.ListFilterKeyForExtension(gurps.TraitsExt)
	spellsKey := gurps.ListFilterKeyForExtension(gurps.SpellsExt)
	equipmentKey := gurps.ListFilterKeyForExtension(gurps.EquipmentExt)
	for _, one := range []struct {
		name        string
		rows        []*NavigatorNode
		wantKeys    []string
		wantEntries []gurps.LibraryFile
	}{
		{
			name:        "files of one list type",
			rows:        []*NavigatorNode{skill},
			wantKeys:    []string{librarySkillsKey},
			wantEntries: []gurps.LibraryFile{{Library: u, Path: "Skills/Basic.skl"}},
		},
		{name: "a file that is no list file", rows: []*NavigatorNode{skill, readMe}},
		{name: "files of two list types", rows: []*NavigatorNode{skill, traits}},
		{
			name:        "a folder holding several list types offers each",
			rows:        []*NavigatorNode{mixed},
			wantKeys:    []string{traitsKey, spellsKey, equipmentKey},
			wantEntries: []gurps.LibraryFile{{Library: u, Path: "Mixed"}},
		},
		{name: "a folder holding no list file offers nothing", rows: []*NavigatorNode{docs}},
		{
			name:        "a selected file decides the type for the folders selected with it",
			rows:        []*NavigatorNode{mixed, skill},
			wantKeys:    []string{librarySkillsKey},
			wantEntries: []gurps.LibraryFile{{Library: u, Path: "Mixed"}, {Library: u, Path: "Skills/Basic.skl"}},
		},
		{
			name:        "a library stands for all of it",
			rows:        []*NavigatorNode{lib},
			wantKeys:    []string{traitsKey, librarySkillsKey, spellsKey, equipmentKey},
			wantEntries: []gurps.LibraryFile{{Library: u}},
		},
		{
			name:        "a file under Favorites and under its library is one entry",
			rows:        []*NavigatorNode{favoriteSkill, skill},
			wantKeys:    []string{librarySkillsKey},
			wantEntries: []gurps.LibraryFile{{Library: u, Path: "Skills/Basic.skl"}},
		},
		{
			name:        "the Favorites row stands for nothing",
			rows:        []*NavigatorNode{favorites, mixed},
			wantKeys:    []string{traitsKey, spellsKey, equipmentKey},
			wantEntries: []gurps.LibraryFile{{Library: u, Path: "Mixed"}},
		},
	} {
		keys, entries := libraryFilterSelection(one.rows)
		c.Equal(one.wantKeys, keys, one.name)
		c.Equal(one.wantEntries, entries, one.name)
	}
}

// TestLibraryFilterMenuItems verifies what the Library Explorer's context menu offers for opening a library filter
// view: one item that can be chosen for a single list type, and for several, a header that can't be, followed by an
// item for each type, indented beneath it.
func TestLibraryFilterMenuItems(t *testing.T) {
	c := check.New(t)
	traitsKey := gurps.ListFilterKeyForExtension(gurps.TraitsExt)
	spellsKey := gurps.ListFilterKeyForExtension(gurps.SpellsExt)
	c.Equal([]libraryFilterMenuItem{{title: "Open in Library Filter", key: librarySkillsKey}},
		libraryFilterMenuItems([]string{librarySkillsKey}), "one type")
	c.Equal([]libraryFilterMenuItem{
		{title: "Open in Library Filter:"},
		{title: "    Traits", key: traitsKey},
		{title: "    Spells", key: spellsKey},
	}, libraryFilterMenuItems([]string{traitsKey, spellsKey}), "several types")
}
