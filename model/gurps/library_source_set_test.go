// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"runtime"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestLibraryFileCovers verifies that an entry covers the whole of its library when its path is empty, and otherwise
// the file at its path or what lies within the folder at its path, and that a source set covers what any of its entries
// does, or everything when it is nil.
func TestLibraryFileCovers(t *testing.T) {
	c := check.New(t)
	file := func(library, p string) gurps.LibraryFile { return gurps.LibraryFile{Library: library, Path: p} }
	for _, one := range []struct {
		name  string
		entry gurps.LibraryFile
		f     gurps.LibraryFile
		want  bool
	}{
		{"whole library", file("lib", ""), file("lib", "Magic/Fire.spl"), true},
		{"exact file", file("lib", "Magic/Fire.spl"), file("lib", "Magic/Fire.spl"), true},
		{"folder", file("lib", "Magic"), file("lib", "Magic/Fire/Basic.spl"), true},
		{"folder name prefix", file("lib", "Magic"), file("lib", "Magic Items/Wand.spl"), false},
		{"different library", file("lib", ""), file("other", "Magic/Fire.spl"), false},
	} {
		c.Equal(one.want, one.entry.Covers(one.f), one.name)
	}

	var all *gurps.LibrarySourceSet
	c.True(all.Covers(file("any", "x.spl")), "a nil set stands for all files")
	set := &gurps.LibrarySourceSet{Name: "Some", Entries: []gurps.LibraryFile{file("lib", "Magic"), file("other", "")}}
	c.True(set.Covers(file("other", "x.spl")), "a set covers what any of its entries covers")
	c.False(set.Covers(file("lib", "x.spl")), "a set covers nothing none of its entries covers")
}

// TestSettingsLibrarySourceSetHelpers verifies the settings-level bookkeeping for source sets: the ordering a user
// sees, the copies that keep callers from reaching into the stored list, lookups by identity and name, the pruning and
// normalizing that EnsureValidity does, rekeying a library, renaming a file or folder, and the round trip through a
// settings file.
func TestSettingsLibrarySourceSetHelpers(t *testing.T) {
	c := check.New(t)
	const key = "spl"
	s := &gurps.Settings{}

	// Sets are ordered by name, ignoring case, and numbers within a name compare as numbers.
	for _, name := range []string{"b", "item10", "A", "C", "item9"} {
		s.AddLibrarySourceSet(key, &gurps.LibrarySourceSet{Name: name})
	}
	c.Equal([]string{"A", "b", "C", "item9", "item10"}, sourceSetNames(s.LibrarySourceSetsFor(key)),
		"source sets are sorted naturally, without regard to case")

	// The returned slice is a copy, so a caller can't reorder or shorten what is stored.
	list := s.LibrarySourceSetsFor(key)
	list[0] = &gurps.LibrarySourceSet{Name: "intruder"}
	c.Equal([]string{"A", "b", "C", "item9", "item10"}, sourceSetNames(s.LibrarySourceSetsFor(key)),
		"the returned slice is a copy of the stored one")

	// A set renamed in place goes to its place when the list is sorted again.
	s.LibrarySourceSetsFor(key)[1].Name = "zzz"
	s.ResortLibrarySourceSets(key)
	c.Equal([]string{"A", "C", "item9", "item10", "zzz"}, sourceSetNames(s.LibrarySourceSetsFor(key)),
		"re-sorting puts the renamed set in its place")

	// Sets are found by identity and by name.
	first := s.LibrarySourceSetsFor(key)[0]
	found, ok := s.LibrarySourceSetKey(first)
	c.True(ok, "a stored set is found")
	c.Equal(key, found, "a stored set is found under its key")
	_, ok = s.LibrarySourceSetKey(first.Clone())
	c.False(ok, "a copy of a stored set isn't stored")
	c.True(s.LibrarySourceSetNamed(key, "  a ") == first, "a set is found by name regardless of case and whitespace")
	c.Nil(s.LibrarySourceSetNamed("adq", "A"), "names are tracked per list type")
	c.True(s.LibrarySourceSetNameInUse(key, " a ", nil), "a name in use is detected regardless of case and whitespace")
	c.False(s.LibrarySourceSetNameInUse(key, "A", first), "the set named as the exception doesn't count")
	c.Equal("a (2)", s.UniqueLibrarySourceSetName(key, "a"), "a name in use gets a number")

	// Removal is by pointer identity.
	s.RemoveLibrarySourceSet(key, first.Clone())
	c.Equal([]string{"A", "C", "item9", "item10", "zzz"}, sourceSetNames(s.LibrarySourceSetsFor(key)),
		"removing a set that isn't stored changes nothing, even when it bears a stored set's name")
	s.RemoveLibrarySourceSet(key, first)
	c.Equal([]string{"C", "item9", "item10", "zzz"}, sourceSetNames(s.LibrarySourceSetsFor(key)),
		"removing drops just that set")

	// EnsureValidity prunes what can't be used, keeps what it doesn't recognize and normalizes the entries.
	keep := &gurps.LibrarySourceSet{Name: " Keep ", Entries: []gurps.LibraryFile{
		{Library: "lib", Path: "Magic/Fire/Basic.spl"},
		{Library: "lib", Path: "Spells/../Magic/Fire"},
		{Library: "lib", Path: "../Outside.spl"},
		{Library: "lib", Path: "/Absolute.spl"},
		{Library: "", Path: "Orphan.spl"},
		{Library: "lib", Path: "Item10.spl"},
		{Library: "lib", Path: "item9.spl"},
		{Library: "lib", Path: "Magic Items"},
		{Library: "other", Path: "./"},
		{Library: "other", Path: "Anything.spl"},
	}}
	future := &gurps.LibrarySourceSet{
		Name:    "From a newer GCS",
		Entries: []gurps.LibraryFile{{Library: "other", Path: "New.spl"}},
	}
	s.LibraryFilter.Sets = map[string][]*gurps.LibrarySourceSet{
		key:      {nil, keep, {Name: "  "}, {Name: "keep"}},
		"adq":    {nil, {Name: ""}},
		"skl":    {{Name: "izmir"}, {Name: "İzmir"}},
		"future": {future},
	}
	s.EnsureValidity()
	c.Equal([]string{"Keep"}, sourceSetNames(s.LibrarySourceSetsFor(key)),
		"nil, unnamed and duplicate-named sets are dropped")
	c.Equal([]string{"izmir", "İzmir"}, sourceSetNames(s.LibrarySourceSetsFor("skl")),
		"names are duplicates only when LibrarySourceSetNameInUse would say so")
	c.Equal([]gurps.LibraryFile{
		{Library: "lib", Path: "item9.spl"},
		{Library: "lib", Path: "Item10.spl"},
		{Library: "lib", Path: "Magic Items"},
		{Library: "lib", Path: "Magic/Fire"},
		{Library: "other", Path: ""},
	}, keep.Entries, "entries are cleaned, sorted, and dropped when unusable, repeated or covered by another")
	odd := `Odd\Name.spl`
	if runtime.GOOS == "windows" {
		odd = "Odd/Name.spl"
	}
	c.Equal([]gurps.LibraryFile{{Library: "lib", Path: odd}},
		gurps.NormalizeLibrarySourceEntries([]gurps.LibraryFile{{Library: "lib", Path: `Odd\Name.spl`}}),
		"a backslash separates the names in a path only on Windows, and is part of a name elsewhere")
	drive := []gurps.LibraryFile{{Library: "lib", Path: "C:x.skl"}, {Library: "lib", Path: `C:\Lib\x.skl`}}
	want := drive
	if runtime.GOOS == "windows" {
		want = nil
	}
	c.Equal(want, gurps.NormalizeLibrarySourceEntries(drive),
		"a drive letter makes a path lead out of its library only on Windows, and is part of a name elsewhere")
	_, exists := s.LibraryFilter.Sets["adq"]
	c.False(exists, "a key left with nothing usable is deleted")
	c.Equal([]string{"From a newer GCS"}, sourceSetNames(s.LibrarySourceSetsFor("future")),
		"an unrecognized key is kept, since a newer version of GCS may have written it")
	c.NotNil(s.LibraryFilter.LastSets, "the last sets used can be recorded")

	// Rekeying a library reaches every set and merges the entries that now share a library.
	s.RekeyLibrarySourceSets("other", "lib")
	c.Equal([]gurps.LibraryFile{{Library: "lib", Path: ""}}, keep.Entries,
		"entries now covered by a whole-library entry are dropped")
	c.Equal([]gurps.LibraryFile{{Library: "lib", Path: "New.spl"}}, future.Entries,
		"the sets of every list type are rekeyed")

	// Renaming a file or folder reaches every set, moving the entries for it and within it, but not those of another
	// library or of a folder whose name begins with its own, and normalizes the entries it moved.
	moved := &gurps.Settings{}
	folder := &gurps.LibrarySourceSet{Name: "Folder", Entries: []gurps.LibraryFile{
		{Library: "lib", Path: "Arcane"},
		{Library: "lib", Path: "Magic Items"},
		{Library: "lib", Path: "Magic/Fire/Basic.spl"},
		{Library: "other", Path: "Magic"},
	}}
	exact := &gurps.LibrarySourceSet{Name: "Exact", Entries: []gurps.LibraryFile{{Library: "lib", Path: "Magic"}}}
	moved.AddLibrarySourceSet(key, folder)
	moved.AddLibrarySourceSet("skl", exact)
	c.True(moved.RenameLibrarySourcePaths("lib", "Magic", "Arcane"), "a rename that moves entries says so")
	c.Equal([]gurps.LibraryFile{
		{Library: "lib", Path: "Arcane"},
		{Library: "lib", Path: "Magic Items"},
		{Library: "other", Path: "Magic"},
	}, folder.Entries, "an entry within the folder moves with it, and is dropped once another entry covers it")
	c.Equal([]gurps.LibraryFile{{Library: "lib", Path: "Arcane"}}, exact.Entries,
		"the sets of every list type follow the rename")
	c.False(moved.RenameLibrarySourcePaths("lib", "Magic", "Arcane"), "a rename that moves no entry says so")

	// The sets and the last ones used are written to, and read back from, a settings file.
	source := &gurps.Settings{}
	source.AddLibrarySourceSet(key, &gurps.LibrarySourceSet{
		Name:    "Fantasy",
		Entries: []gurps.LibraryFile{{Library: "lib", Path: "Magic"}},
	})
	source.LibraryFilter.LastKey = key
	source.LibraryFilter.LastSets = map[string]string{key: "Fantasy"}
	data, err := jio.Marshal(source)
	c.NoError(err, "settings holding source sets should marshal")
	c.Contains(string(data), `"library_filter":{"sets":{"spl":[{"name":"Fantasy","entries":[{"library":"lib",`+
		`"path":"Magic"}]}]},"last_key":"spl","last_sets":{"spl":"Fantasy"}}`, "the source sets are written")
	var loaded gurps.Settings
	c.NoError(jio.Unmarshal(data, &loaded), "settings holding source sets should unmarshal")
	loaded.EnsureValidity()
	c.Equal(source.LibraryFilter, loaded.LibraryFilter, "the source sets read back are the ones written")
}

// sourceSetNames returns the names of the given sets, so that ordering can be compared directly.
func sourceSetNames(sets []*gurps.LibrarySourceSet) []string {
	names := make([]string, len(sets))
	for i, set := range sets {
		names[i] = set.Name
	}
	return names
}
