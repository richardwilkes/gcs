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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	unisoncheck "github.com/richardwilkes/unison/enums/check"
)

// sourceSetRowState is what a test expects of a row of a source set's checklist.
type sourceSetRowState struct {
	state   unisoncheck.Enum
	enabled bool
}

// TestSourceSetChecklist verifies the rules of a source set's checklist: checking a folder checks the rows beneath it,
// which can't then be changed, and drops their own entries; a folder with only some of what is beneath it checked is
// mixed; unchecking an entry naming nothing takes it out of the set while its row stays to be checked again; changing
// the type keeps the library and folder entries, drops the file entries and shows the files of the new type; and the
// entries the checklist yields are normalized.
func TestSourceSetChecklist(t *testing.T) {
	c := check.New(t)
	master, user := uxtest.UseTestLibraries(t, c)
	skills := testNavigatorFolder("Skills", "Skills/A.skl", "Skills/B.skl")
	skills.entries = append(skills.entries, testNavigatorFolder("Skills/Deep", "Skills/Deep/C.skl"))
	listing := &navigatorListing{libraries: []libraryListing{
		{library: master, entries: []*navigatorEntry{{path: "Basic.skl"}}},
		{library: user, entries: []*navigatorEntry{
			skills,
			testNavigatorFolder("Spells", "Spells/S.spl"),
			{path: "Top.skl"},
		}},
	}}
	m := master.Key()
	u := user.Key()
	// Normalized entries are sorted by library key, which puts the User Library's first.
	c.True(u < m, "the User Library's key sorts ahead of the Master Library's")
	entry := func(key, p string) gurps.LibraryFile { return gurps.LibraryFile{Library: key, Path: p} }
	on := sourceSetRowState{state: unisoncheck.On, enabled: true}
	implied := sourceSetRowState{state: unisoncheck.On}
	off := sourceSetRowState{state: unisoncheck.Off, enabled: true}
	mixed := sourceSetRowState{state: unisoncheck.Mixed, enabled: true}
	for _, one := range []struct {
		name    string
		entries []gurps.LibraryFile
		ext     string              // The extension the checklist is switched to, if any, ahead of the toggles
		toggles []gurps.LibraryFile // The rows toggled, by entry, in order
		want    []gurps.LibraryFile
		rows    map[gurps.LibraryFile]sourceSetRowState
		absent  []gurps.LibraryFile // Entries with no row
	}{
		{
			name:    "checking a folder implies what is beneath it",
			entries: []gurps.LibraryFile{entry(u, "Skills/A.skl"), entry(u, "Skills/Deep/C.skl")},
			toggles: []gurps.LibraryFile{entry(u, "Skills")},
			want:    []gurps.LibraryFile{entry(u, "Skills")},
			rows: map[gurps.LibraryFile]sourceSetRowState{
				entry(u, ""):                  mixed,
				entry(u, "Skills"):            on,
				entry(u, "Skills/A.skl"):      implied,
				entry(u, "Skills/Deep"):       implied,
				entry(u, "Skills/Deep/C.skl"): implied,
				entry(u, "Top.skl"):           off,
			},
			absent: []gurps.LibraryFile{entry(u, "Spells"), entry(u, "Spells/S.spl")},
		},
		{
			name:    "a partly checked folder is mixed",
			entries: []gurps.LibraryFile{entry(u, "Skills/Deep/C.skl")},
			want:    []gurps.LibraryFile{entry(u, "Skills/Deep/C.skl")},
			rows: map[gurps.LibraryFile]sourceSetRowState{
				entry(m, ""):                  off,
				entry(u, ""):                  mixed,
				entry(u, "Skills"):            mixed,
				entry(u, "Skills/A.skl"):      off,
				entry(u, "Skills/Deep"):       mixed,
				entry(u, "Skills/Deep/C.skl"): on,
			},
		},
		{
			name:    "unchecking an entry naming nothing removes it",
			entries: []gurps.LibraryFile{entry(u, "Gone.skl"), entry(u, "Top.skl")},
			toggles: []gurps.LibraryFile{entry(u, "Gone.skl")},
			want:    []gurps.LibraryFile{entry(u, "Top.skl")},
			rows: map[gurps.LibraryFile]sourceSetRowState{
				entry(u, "Gone.skl"): off,
				entry(u, "Top.skl"):  on,
			},
		},
		{
			name: "changing the type keeps folders and drops files",
			entries: []gurps.LibraryFile{
				entry(u, "Skills"), entry(u, "Top.skl"), entry(m, ""), entry(u, "Gone"), entry(u, "Lost.skl"),
			},
			ext:  gurps.SpellsExt,
			want: []gurps.LibraryFile{entry(u, "Gone"), entry(u, "Skills"), entry(m, "")},
			rows: map[gurps.LibraryFile]sourceSetRowState{
				entry(m, ""):             on,
				entry(u, "Gone"):         on,
				entry(u, "Skills"):       on,
				entry(u, "Spells"):       off,
				entry(u, "Spells/S.spl"): off,
			},
			absent: []gurps.LibraryFile{entry(m, "Basic.skl"), entry(u, "Top.skl"), entry(u, "Skills/A.skl")},
		},
		{
			name:    "the entries are normalized",
			entries: []gurps.LibraryFile{entry(u, "Top.skl"), entry(u, `Skills\A.skl`), entry(u, "Top.skl")},
			toggles: []gurps.LibraryFile{entry(m, "Basic.skl")},
			want:    []gurps.LibraryFile{entry(u, "Skills/A.skl"), entry(u, "Top.skl"), entry(m, "Basic.skl")},
		},
	} {
		ext := gurps.SkillsExt
		list := newSourceSetChecklist(listing, gurps.GlobalSettings().Libraries, ext, one.entries)
		if one.ext != "" {
			list.setExtension(one.ext)
		}
		rowFor := func(f gurps.LibraryFile) int {
			for i, row := range list.rows {
				if row.entry == f {
					return i
				}
			}
			return -1
		}
		for _, toggled := range one.toggles {
			i := rowFor(toggled)
			c.NotEqual(-1, i, "%s: %v has a row to toggle", one.name, toggled)
			list.toggle(i)
		}
		c.Equal(one.want, list.result(), one.name)
		for f, want := range one.rows {
			if i := rowFor(f); i == -1 {
				t.Errorf("%s: %v has no row", one.name, f)
			} else {
				c.Equal(want, sourceSetRowState{state: list.rows[i].state, enabled: list.rows[i].enabled},
					"%s: the row of %v", one.name, f)
			}
		}
		for _, f := range one.absent {
			c.Equal(-1, rowFor(f), "%s: %v has no row", one.name, f)
		}
	}
}

// TestCheckLibrarySourceSet verifies when the source set editor accepts a set: its name, which may be the one the set
// being edited has already, can be neither blank nor that of another set of its type, ignoring case and surrounding
// whitespace, nor what the source set popup calls the type's set of every file or an unsaved set; and it needs an
// entry. What is wrong is said, the name first, except for a blank name, which the name field's watermark covers.
func TestCheckLibrarySourceSet(t *testing.T) {
	c := check.New(t)
	skills, ok := libraryFilterKindFor(librarySkillsKey)
	c.True(ok, "skills are a list type")
	settings := &gurps.Settings{}
	ours := &gurps.LibrarySourceSet{Name: "Our Set"}
	settings.AddLibrarySourceSet(librarySkillsKey, ours)
	settings.AddLibrarySourceSet(gurps.ListFilterKeyForExtension(gurps.SpellsExt), &gurps.LibrarySourceSet{Name: "Magic"})
	const inUse = "Another set of this type already has this name."
	const empty = "Check at least one library, folder or file."
	for _, one := range []struct {
		name      string
		setName   string
		entries   int
		except    *gurps.LibrarySourceSet
		nameValid bool
		accept    bool
		message   string
	}{
		{name: "a free name", setName: "Mine", entries: 1, nameValid: true, accept: true},
		{name: "no entries", setName: "Mine", nameValid: true, message: empty},
		{name: "a blank name", setName: "  ", entries: 1},
		{name: "another set's name", setName: " our set ", entries: 1, message: inUse},
		{name: "another set's name and no entries", setName: "Our Set", message: inUse},
		{name: "the edited set's own name", setName: "our set", entries: 1, except: ours, nameValid: true, accept: true},
		{name: "a set's name of another type", setName: "Magic", entries: 1, nameValid: true, accept: true},
		{name: "every file's title", setName: "all skills", entries: 1, message: inUse},
		{name: "an unsaved set's title", setName: " Selected Skills ", entries: 1, message: inUse},
		{name: "every file's title of another type", setName: "All Spells", entries: 1, nameValid: true, accept: true},
	} {
		nameValid, accept, message := checkLibrarySourceSet(settings, skills, one.setName, one.entries, one.except)
		c.Equal(one.nameValid, nameValid, "%s: the name is valid", one.name)
		c.Equal(one.accept, accept, "%s: the set can be accepted", one.name)
		c.Equal(one.message, message, "%s: the message", one.name)
	}
}
