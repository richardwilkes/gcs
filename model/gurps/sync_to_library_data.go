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
	f := convertWalker(pathSet, xslices.Set([]string{SheetExt, TemplatesExt, LootExt}))
	for _, p := range paths {
		_ = filepath.WalkDir(p, f) //nolint:errcheck // We want to continue on even if there was an error
	}
	list := slices.SortedFunc(maps.Keys(pathSet), func(a, b string) int { return xstrings.NaturalCmp(a, b, true) })
	for _, p := range list {
		fmt.Printf(i18n.Text("Processing %s\n"), p)
		switch strings.ToLower(filepath.Ext(p)) {
		case TemplatesExt:
			var tmpl *Template
			if tmpl, err = NewTemplateFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			tmpl.EnsureAttachments()
			tmpl.SourceMatcher().PrepareHashes(tmpl)
			tmpl.SyncWithLibrarySources()
			if err = tmpl.Save(p); err != nil {
				return err
			}
		case LootExt:
			var loot *Loot
			if loot, err = NewLootFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			loot.EnsureAttachments()
			loot.SourceMatcher().PrepareHashes(loot)
			loot.SyncWithLibrarySources()
			if err = loot.Save(p); err != nil {
				return err
			}
		case SheetExt:
			var entity *Entity
			if entity, err = NewEntityFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			entity.ensureAttachments()
			entity.SourceMatcher().PrepareHashes(entity)
			entity.SyncWithLibrarySources()
			entity.Recalculate()
			if err = entity.Save(p); err != nil {
				return err
			}
		}
	}
	if len(list) == 1 {
		fmt.Println(i18n.Text("Processed 1 file"))
	} else {
		fmt.Printf(i18n.Text("Processed %d files\n"), len(list))
	}
	return nil
}
