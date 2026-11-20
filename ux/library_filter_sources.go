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
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
)

// libraryFilterFile is one of the library files a library filter view draws its rows from.
type libraryFilterFile struct {
	gurps.LibraryFile
	libTitle string // The title of the file's library
	fullPath string // Where the file is on disk, named as the library's watch names it
	display  string // What the File column shows for the file's rows
	tooltip  string // The tooltip of the File column for the file's rows
}

// libraryFilterCacheKey identifies a file's parse in a library filter view's cache. The library and path alone would
// not do, since a library may have been moved on disk, nor would the path on disk, since a library rooted within
// another's folder lists the same file a second time, and each listing gets rows of its own.
type libraryFilterCacheKey struct {
	gurps.LibraryFile
	fullPath string
}

func (f *libraryFilterFile) cacheKey() libraryFilterCacheKey {
	return libraryFilterCacheKey{LibraryFile: f.LibraryFile, fullPath: f.fullPath}
}

// resolveLibraryFilterFiles returns the files with the extension that the set covers (all of them for a nil set), in
// listing order by library and natural path order within one, and the set's entries that name nothing in the listing.
// The extension is compared without regard to case. Libraries no longer registered are skipped. A nil listing yields
// nothing.
func resolveLibraryFilterFiles(listing *navigatorListing, libs *library.Libraries, ext string,
	set *gurps.LibrarySourceSet,
) (files []libraryFilterFile, missing []gurps.LibraryFile) {
	if listing == nil {
		return nil, nil
	}
	var found map[gurps.LibraryFile]bool
	if set != nil {
		found = make(map[gurps.LibraryFile]bool, len(set.Entries))
	}
	for _, one := range listing.libraries {
		lib := one.library
		key := lib.Key()
		if libs.Lookup(key) != lib {
			continue
		}
		if found != nil {
			found[gurps.LibraryFile{Library: key}] = true
		}
		title := lib.Data().Title
		root := lib.Path(false)
		start := len(files)
		var walk func(entries []*navigatorEntry)
		walk = func(entries []*navigatorEntry) {
			for _, entry := range entries {
				f := gurps.LibraryFile{Library: key, Path: entry.path}
				if found != nil {
					found[f] = true
				}
				if entry.isDir {
					walk(entry.entries)
				} else if strings.EqualFold(path.Ext(entry.path), ext) && set.Covers(f) {
					files = append(files, libraryFilterFile{
						LibraryFile: f,
						libTitle:    title,
						fullPath:    filepath.Join(root, filepath.FromSlash(entry.path)),
					})
				}
			}
		}
		walk(one.entries)
		slices.SortStableFunc(files[start:], func(a, b libraryFilterFile) int {
			return xstrings.NaturalCmp(a.Path, b.Path, true)
		})
	}
	if set != nil {
		for _, entry := range set.Entries {
			if !found[entry] {
				missing = append(missing, entry)
			}
		}
	}
	nameLibraryFilterFiles(files)
	return files, missing
}

// nameLibraryFilterFiles sets what the File column shows for each of the files: its name without the extension, or,
// where that is shared with another, its path within its library, which is put after its library's title where even
// that is shared. The column's tooltip names the file's library and its path within it, then its path on disk.
func nameLibraryFilterFiles(files []libraryFilterFile) {
	trimmed := func(p string) string { return strings.TrimSuffix(p, path.Ext(p)) }
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = trimmed(path.Base(f.Path))
	}
	widen := func(name func(f libraryFilterFile) string) {
		counts := make(map[string]int, len(names))
		for _, one := range names {
			counts[one]++
		}
		for i, one := range names {
			if counts[one] > 1 {
				names[i] = name(files[i])
			}
		}
	}
	widen(func(f libraryFilterFile) string { return trimmed(f.Path) })
	widen(func(f libraryFilterFile) string { return f.libTitle + ": " + trimmed(f.Path) })
	for i := range files {
		f := &files[i]
		f.display = names[i]
		f.tooltip = f.libTitle + ": " + f.Path + "\n" + f.fullPath
	}
}

// libraryFilterSelection returns the list types a library filter could show for the selected rows and the source
// entries they stand for. Selected files decide the type, and must all be list files of one type, or nothing is
// offered; with none selected, every list type with a file beneath the selected libraries and folders is offered, as
// the Library Explorer last listed them. A library stands for all of it, a folder or file for its path within its
// library; a row shown under Favorites stands for the same thing it does under its library, and the Favorites row
// itself for nothing.
func libraryFilterSelection(rows []*NavigatorNode) (keys []string, entries []gurps.LibraryFile) {
	kinds := libraryFilterKinds()
	keysByExt := make(map[string]string, len(kinds))
	for _, kind := range kinds {
		keysByExt[strings.ToLower(kind.ext)] = kind.key
	}
	keyFor := func(row *NavigatorNode) string { return keysByExt[strings.ToLower(path.Ext(row.path))] }
	var fileKey string
	var containers []*NavigatorNode
	for _, row := range rows {
		switch {
		case row.IsFavorites():
		case row.IsFile():
			key := keyFor(row)
			if key == "" || (fileKey != "" && key != fileKey) {
				return nil, nil
			}
			fileKey = key
			entries = append(entries, gurps.LibraryFile{Library: row.library.Key(), Path: row.path})
		default:
			containers = append(containers, row)
			entries = append(entries, gurps.LibraryFile{Library: row.library.Key(), Path: row.path})
		}
	}
	if len(entries) == 0 {
		return nil, nil
	}
	if fileKey != "" {
		return []string{fileKey}, gurps.NormalizeLibrarySourceEntries(entries)
	}
	found := make(map[string]bool, len(kinds))
	// walk returns true once every list type has been found, since nothing more can be.
	var walk func(list []*NavigatorNode) bool
	walk = func(list []*NavigatorNode) bool {
		for _, row := range list {
			if row.IsFile() {
				if key := keyFor(row); key != "" && !found[key] {
					found[key] = true
					if len(found) == len(kinds) {
						return true
					}
				}
			} else if walk(row.children) {
				return true
			}
		}
		return false
	}
	walk(containers)
	for _, kind := range kinds {
		if found[kind.key] {
			keys = append(keys, kind.key)
		}
	}
	if len(keys) == 0 {
		return nil, nil
	}
	return keys, gurps.NormalizeLibrarySourceEntries(entries)
}

// libraryFilterMenuItem is an item the Library Explorer's context menu offers for opening a library filter view: its
// title and the key of the list type it opens a view on, which is empty for a header, which can't be chosen.
type libraryFilterMenuItem struct {
	title string
	key   string
}

// libraryFilterMenuItems returns the items that open a library filter view on one of the list types with the keys: one
// item for one type, or for several, a header followed by an item per type, indented beneath it.
func libraryFilterMenuItems(keys []string) []libraryFilterMenuItem {
	if len(keys) == 1 {
		return []libraryFilterMenuItem{{title: i18n.Text("Open in Library Filter"), key: keys[0]}}
	}
	items := []libraryFilterMenuItem{{title: i18n.Text("Open in Library Filter:")}}
	for _, key := range keys {
		if kind, ok := libraryFilterKindFor(key); ok {
			items = append(items, libraryFilterMenuItem{title: "    " + kind.title, key: key})
		}
	}
	return items
}

// appendLibraryFilterMenuItems appends to the Library Explorer's context menu the items that open a library filter view
// on the selected rows as an unsaved set (see libraryFilterMenuItems), followed by a separator, when there is a list
// type to offer (see libraryFilterSelection).
func appendLibraryFilterMenuItems(f unison.MenuFactory, cm unison.Menu, id *int, sel []*NavigatorNode) {
	keys, entries := libraryFilterSelection(sel)
	if len(keys) == 0 {
		return
	}
	for _, item := range libraryFilterMenuItems(keys) {
		var handler func(unison.MenuItem)
		var enabled func(unison.MenuItem) bool
		if key := item.key; key == "" {
			enabled = func(unison.MenuItem) bool { return false }
		} else {
			handler = func(unison.MenuItem) {
				openLibraryFilter(key, &gurps.LibrarySourceSet{Entries: slices.Clone(entries)}, true)
			}
		}
		cm.InsertItem(-1, f.NewItem(unison.PopupMenuTemporaryBaseID+*id, item.title, unison.KeyBinding{}, enabled,
			handler))
		*id++
	}
	cm.InsertSeparator(-1, true)
}
