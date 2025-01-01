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
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	xfs "github.com/richardwilkes/toolbox/xio/fs"
)

// NamedFileRef holds a reference to a file.
type NamedFileRef struct {
	Name       string
	FileSystem fs.FS
	FilePath   string
}

func (n *NamedFileRef) String() string {
	return n.Name
}

// NamedFileSet holds a named list of file references.
type NamedFileSet struct {
	Name string
	List []*NamedFileRef
}

// ScanForNamedFileSets scans for settings files of a particular type.
func ScanForNamedFileSets(builtIn fs.FS, builtInDir string, omitDuplicateNames bool, libraries Libraries, extensions ...string) []*NamedFileSet {
	set := make(map[string]bool)
	list := make([]*NamedFileSet, 0)
	for _, lib := range libraries.List() {
		if refs := scanForNamedFileSets(os.DirFS(lib.Path()), "Settings", extensions, omitDuplicateNames, set); len(refs) != 0 {
			list = append(list, &NamedFileSet{
				Name: lib.Title,
				List: refs,
			})
		}
	}
	if builtIn != nil {
		if refs := scanForNamedFileSets(builtIn, builtInDir, extensions, omitDuplicateNames, set); len(refs) != 0 {
			list = append(list, &NamedFileSet{
				Name: i18n.Text("Built-in"),
				List: refs,
			})
		}
	}
	return list
}

func scanForNamedFileSets(fileSystem fs.FS, dirPath string, extensions []string, omitDuplicateNames bool, set map[string]bool) []*NamedFileRef {
	extMap := make(map[string]bool, len(extensions))
	for _, ext := range extensions {
		extMap[strings.ToLower(ext)] = true
	}
	list := make([]*NamedFileRef, 0)
	_ = fs.WalkDir(fileSystem, dirPath, func(p string, d fs.DirEntry, err error) error { //nolint:errcheck // Intentionally ignored the error result
		if err != nil {
			return nil
		}
		name := d.Name()
		if strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !d.IsDir() && extMap[path.Ext(name)] {
			shortName := xfs.TrimExtension(name)
			if shortLowerName := strings.ToLower(shortName); !omitDuplicateNames || !set[shortLowerName] {
				set[shortLowerName] = true
				list = append(list, &NamedFileRef{
					Name:       shortName,
					FileSystem: fileSystem,
					FilePath:   p,
				})
			}
		}
		return nil
	})
	slices.SortFunc(list, func(a, b *NamedFileRef) int {
		if a.Name == b.Name {
			return txt.NaturalCmp(a.FilePath, b.FilePath, true)
		}
		return txt.NaturalCmp(a.Name, b.Name, true)
	})
	return list
}
