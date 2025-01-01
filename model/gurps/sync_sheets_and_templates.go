// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"os"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/toolbox/collection"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xio/fs"
)

// SyncSheetsAndTemplates syncs GCS sheet and template files found in the given paths with their source libraries.
func SyncSheetsAndTemplates(paths ...string) error {
	var err error
	paths, err = fs.UniquePaths(paths...)
	if err != nil {
		return err
	}
	pathSet := collection.NewSet[string]()
	f := convertWalker(pathSet, collection.NewSet(SheetExt, TemplatesExt))
	for _, p := range paths {
		_ = filepath.WalkDir(p, f) //nolint:errcheck // We want to continue on even if there was an error
	}
	list := pathSet.Values()
	txt.SortStringsNaturalAscending(list)
	for _, p := range list {
		fmt.Printf(i18n.Text("Processing %s\n"), p)
		switch strings.ToLower(filepath.Ext(p)) {
		case TemplatesExt:
			var tmpl *Template
			if tmpl, err = NewTemplateFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			tmpl.SyncWithLibrarySources()
			tmpl.EnsureAttachments()
			if err = tmpl.Save(p); err != nil {
				return err
			}
		case SheetExt:
			var entity *Entity
			if entity, err = NewEntityFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
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
