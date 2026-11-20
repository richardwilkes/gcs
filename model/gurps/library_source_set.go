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
	"cmp"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/richardwilkes/toolbox/v2/xstrings"
)

// LibrarySourceSet is a named list of the library files a library filter view draws its rows from. Each entry's Path is
// "" for the whole library, or the library-relative, slash-separated path of a folder or file.
type LibrarySourceSet struct {
	Name    string        `json:"name"`
	Entries []LibraryFile `json:"entries,omitempty"`
}

// LibraryFilterSettings holds the settings of the library filter views.
type LibraryFilterSettings struct {
	// Sets holds the source sets of each library list type, keyed as the saved filters are (see
	// ListFilterKeyForExtension).
	Sets map[string][]*LibrarySourceSet `json:"sets,omitempty"`
	// LastKey is the key of the list type a new view starts on.
	LastKey string `json:"last_key,omitzero"`
	// LastSets holds, by list type key, the name of the source set last used. A type with no name, or one no set has,
	// last used all its files.
	LastSets map[string]string `json:"last_sets,omitempty"`
}

// IsZero reports whether json's omitzero option should omit these settings.
func (l LibraryFilterSettings) IsZero() bool {
	return len(l.Sets) == 0 && l.LastKey == "" && len(l.LastSets) == 0
}

// Covers returns true if l refers to the whole of f's library, to f itself, or to a folder f lies within.
func (l LibraryFile) Covers(f LibraryFile) bool {
	return l.Library == f.Library && (l.Path == "" || l.Path == f.Path || strings.HasPrefix(f.Path, l.Path+"/"))
}

// Covers returns true if any of the set's entries covers f. A nil set stands for all files and covers everything.
func (s *LibrarySourceSet) Covers(f LibraryFile) bool {
	return s == nil || slices.ContainsFunc(s.Entries, func(entry LibraryFile) bool { return entry.Covers(f) })
}

// Clone returns a deep copy of the set, or nil if it is nil.
func (s *LibrarySourceSet) Clone() *LibrarySourceSet {
	if s == nil {
		return nil
	}
	return &LibrarySourceSet{Name: s.Name, Entries: slices.Clone(s.Entries)}
}

// NormalizeLibrarySourceEntries returns a new list of the entries with each path made slash-separated (see
// filepath.ToSlash) and clean, and "" for a whole library. Entries without a library, with an absolute path or with one
// that leads out of its library are dropped, as are duplicates and entries another one covers. The rest are sorted by
// library, then naturally by path.
func NormalizeLibrarySourceEntries(entries []LibraryFile) []LibraryFile {
	var unique []LibraryFile
	for _, entry := range entries {
		entry.Path = path.Clean(filepath.ToSlash(entry.Path))
		if entry.Path == "." {
			entry.Path = ""
		}
		if entry.Library == "" || path.IsAbs(entry.Path) || filepath.VolumeName(entry.Path) != "" || entry.Path == ".." ||
			strings.HasPrefix(entry.Path, "../") || slices.Contains(unique, entry) {
			continue
		}
		unique = append(unique, entry)
	}
	var result []LibraryFile
	for _, entry := range unique {
		if !slices.ContainsFunc(unique, func(other LibraryFile) bool { return other != entry && other.Covers(entry) }) {
			result = append(result, entry)
		}
	}
	slices.SortFunc(result, func(a, b LibraryFile) int {
		return cmp.Or(strings.Compare(a.Library, b.Library), xstrings.NaturalCmp(a.Path, b.Path, true))
	})
	return result
}

// RekeyLibraryEntries changes the library of each of the entries in the library oldKey to newKey.
func RekeyLibraryEntries(entries []LibraryFile, oldKey, newKey string) {
	for i := range entries {
		if entries[i].Library == oldKey {
			entries[i].Library = newKey
		}
	}
}

// RenameLibraryEntries moves each of the entries in the library that is the file or folder at oldPath, or lies within
// it, to the same place at newPath, and returns true if it moved any. Both paths are library-relative and
// slash-separated.
func RenameLibraryEntries(entries []LibraryFile, library, oldPath, newPath string) bool {
	moved := false
	for i := range entries {
		entry := &entries[i]
		if entry.Library != library {
			continue
		}
		if entry.Path == oldPath {
			entry.Path = newPath
		} else if after, ok := strings.CutPrefix(entry.Path, oldPath+"/"); ok {
			entry.Path = newPath + "/" + after
		} else {
			continue
		}
		moved = true
	}
	return moved
}

// LibrarySourceSetsFor returns the source sets for the list type with the given key, sorted by name. The slice is a
// copy, but the sets in it are the live ones.
func (s *Settings) LibrarySourceSetsFor(key string) []*LibrarySourceSet {
	return slices.Clone(s.LibraryFilter.Sets[key])
}

// SetLibrarySourceSetsFor replaces the source sets for the list type with the given key. Nil entries are dropped, the
// rest are sorted by name, and the key is removed when nothing remains.
func (s *Settings) SetLibrarySourceSetsFor(key string, sets []*LibrarySourceSet) {
	sets = slices.DeleteFunc(slices.Clone(sets), func(set *LibrarySourceSet) bool { return set == nil })
	if len(sets) == 0 {
		delete(s.LibraryFilter.Sets, key)
		return
	}
	slices.SortStableFunc(sets, func(a, b *LibrarySourceSet) int { return xstrings.NaturalCmp(a.Name, b.Name, true) })
	if s.LibraryFilter.Sets == nil {
		s.LibraryFilter.Sets = make(map[string][]*LibrarySourceSet)
	}
	s.LibraryFilter.Sets[key] = sets
}

// AddLibrarySourceSet adds a source set for the list type with the given key.
func (s *Settings) AddLibrarySourceSet(key string, set *LibrarySourceSet) {
	s.SetLibrarySourceSetsFor(key, append(s.LibrarySourceSetsFor(key), set))
}

// ResortLibrarySourceSets puts the source sets for the list type with the given key back into name order, which is
// needed after a set has been renamed in place.
func (s *Settings) ResortLibrarySourceSets(key string) {
	s.SetLibrarySourceSetsFor(key, s.LibraryFilter.Sets[key])
}

// RemoveLibrarySourceSet removes set from the source sets for the list type with the given key.
func (s *Settings) RemoveLibrarySourceSet(key string, set *LibrarySourceSet) {
	sets := s.LibrarySourceSetsFor(key)
	if i := slices.Index(sets, set); i != -1 {
		s.SetLibrarySourceSetsFor(key, slices.Delete(sets, i, i+1))
	}
}

// LibrarySourceSetKey returns the key of the list type whose source sets hold set, or false if none does.
func (s *Settings) LibrarySourceSetKey(set *LibrarySourceSet) (string, bool) {
	if set != nil {
		for key, sets := range s.LibraryFilter.Sets {
			if slices.Contains(sets, set) {
				return key, true
			}
		}
	}
	return "", false
}

// LibrarySourceSetNamed returns the source set for the list type with the given key that has the name, ignoring case
// and surrounding whitespace, or nil if there is none.
func (s *Settings) LibrarySourceSetNamed(key, name string) *LibrarySourceSet {
	return s.findLibrarySourceSet(key, name, nil)
}

// LibrarySourceSetNameInUse returns true if a source set for the list type with the given key already has the name,
// ignoring case and surrounding whitespace. The set passed as except, if any, doesn't count, so that a set being edited
// may keep its own name.
func (s *Settings) LibrarySourceSetNameInUse(key, name string, except *LibrarySourceSet) bool {
	return s.findLibrarySourceSet(key, name, except) != nil
}

func (s *Settings) findLibrarySourceSet(key, name string, except *LibrarySourceSet) *LibrarySourceSet {
	name = strings.TrimSpace(name)
	for _, set := range s.LibraryFilter.Sets[key] {
		if set != nil && set != except && strings.EqualFold(strings.TrimSpace(set.Name), name) {
			return set
		}
	}
	return nil
}

// UniqueLibrarySourceSetName returns name if neither a source set for the list type with the given key nor any of the
// reserved names has it, ignoring case and surrounding whitespace, and otherwise the name with the lowest number after
// it, as in "Name (2)", that none has. A number the name already ends with is replaced.
func (s *Settings) UniqueLibrarySourceSetName(key, name string, reserved ...string) string {
	return uniqueNumberedName(name, func(one string) bool {
		return s.LibrarySourceSetNameInUse(key, one, nil) ||
			slices.ContainsFunc(reserved, func(r string) bool { return strings.EqualFold(strings.TrimSpace(r), one) })
	})
}

// RekeyLibrarySourceSets changes the library of each entry of every source set in the library oldKey to newKey, and
// normalizes the entries of each set (see NormalizeLibrarySourceEntries), since newKey may already have been there.
func (s *Settings) RekeyLibrarySourceSets(oldKey, newKey string) {
	for _, sets := range s.LibraryFilter.Sets {
		for _, set := range sets {
			if set != nil {
				RekeyLibraryEntries(set.Entries, oldKey, newKey)
				set.Entries = NormalizeLibrarySourceEntries(set.Entries)
			}
		}
	}
}

// RenameLibrarySourcePaths does what RenameLibraryEntries does to the entries of every source set, normalizing the
// entries of each set it changes (see NormalizeLibrarySourceEntries), since newPath may already have been there, and
// returns true if it changed any.
func (s *Settings) RenameLibrarySourcePaths(library, oldPath, newPath string) bool {
	changed := false
	for _, sets := range s.LibraryFilter.Sets {
		for _, set := range sets {
			if set != nil && RenameLibraryEntries(set.Entries, library, oldPath, newPath) {
				set.Entries = NormalizeLibrarySourceEntries(set.Entries)
				changed = true
			}
		}
	}
	return changed
}

// ensureLibrarySourceSetsValidity drops the source sets that can't be used -- nil entries, sets without a name, and
// later sets that duplicate an earlier one's name -- trims the names of the rest, normalizes their entries and keeps
// each list sorted. Keys that aren't recognized are kept, since a newer version of GCS may have written them. The map
// of the last sets used is made ready for use; a name in it that no set has is kept, standing for all files.
func (s *Settings) ensureLibrarySourceSetsValidity() {
	for key, sets := range s.LibraryFilter.Sets {
		var seen []string
		sets = slices.DeleteFunc(sets, func(set *LibrarySourceSet) bool {
			if set == nil {
				return true
			}
			set.Name = strings.TrimSpace(set.Name)
			sameName := func(name string) bool { return strings.EqualFold(name, set.Name) }
			if set.Name == "" || slices.ContainsFunc(seen, sameName) {
				return true
			}
			seen = append(seen, set.Name)
			set.Entries = NormalizeLibrarySourceEntries(set.Entries)
			return false
		})
		s.SetLibrarySourceSetsFor(key, sets)
	}
	if s.LibraryFilter.LastSets == nil {
		s.LibraryFilter.LastSets = make(map[string]string)
	}
}
