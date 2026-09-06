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
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/toolbox/v2/xslices"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

// SyncToLibraryData syncs GCS sheet, template, and loot files found in the given paths with their source libraries.
func SyncToLibraryData(paths ...string) error {
	var err error
	paths, err = xfilepath.UniquePaths(paths...)
	if err != nil {
		return err
	}
	pathSet := make(map[string]struct{})
	f := convertWalker(pathSet, xslices.Set(slices.Collect(maps.Keys(librarySyncers))))
	for _, p := range paths {
		_ = filepath.WalkDir(p, f) //nolint:errcheck // We want to continue on even if there was an error
	}
	list := slices.SortedFunc(maps.Keys(pathSet), func(a, b string) int { return xstrings.NaturalCmp(a, b, true) })
	for _, p := range list {
		fmt.Printf(i18n.Text("Processing %s\n"), p)
		load, ok := librarySyncers[strings.ToLower(filepath.Ext(p))]
		if !ok {
			continue // The walker only collects the extensions in the table, so this should never happen
		}
		var data librarySyncable
		if data, err = load(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
			return err
		}
		data.EnsureAttachments()
		data.SourceMatcher().PrepareHashes(data)
		data.SyncWithLibrarySources()
		if err = data.Save(p); err != nil {
			return err
		}
	}
	if len(list) == 1 {
		fmt.Println(i18n.Text("Processed 1 file"))
	} else {
		fmt.Printf(i18n.Text("Processed %d files\n"), len(list))
	}
	return nil
}

// librarySyncable is implemented by the file types whose contents can be synced with their source libraries.
type librarySyncable interface {
	ListProvider
	EnsureAttachments()
	SourceMatcher() *SrcMatcher
	SyncWithLibrarySources()
	Save(filePath string) error
}

// librarySyncers maps each GCS file extension, in lowercase, whose contents can be synced with their source libraries
// to the function that loads a file of that type. The walker only collects files whose extensions are in this set.
var librarySyncers = map[string]func(fs.FS, string) (librarySyncable, error){
	LootExt:      loadLibrarySyncable(NewLootFromFile),
	SheetExt:     loadLibrarySyncable(NewEntityFromFile),
	TemplatesExt: loadLibrarySyncable(NewTemplateFromFile),
}

// loadLibrarySyncable adapts a typed file loader to one that returns the loaded data as a librarySyncable.
func loadLibrarySyncable[T librarySyncable](load func(fs.FS, string) (T, error)) func(fs.FS, string) (librarySyncable, error) {
	return func(fileSystem fs.FS, filePath string) (librarySyncable, error) {
		data, err := load(fileSystem, filePath)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
}
