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
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

// NamedFileRef holds a reference to a file.
type NamedFileRef struct {
	// Name is the file's name with its extension removed.
	Name string
	// FileSystem is the file system that holds the file.
	FileSystem fs.FS
	// FilePath is the slash-separated path of the file, relative to FileSystem.
	FilePath string
	// DiskPath is the absolute path of the file on disk, or "" when the file is embedded and has no disk location.
	DiskPath string
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
func ScanForNamedFileSets(builtIn fs.FS, builtInDir string, omitDuplicateNames bool, libraries *Libraries, extensions ...string) []*NamedFileSet {
	set := make(map[string]bool)
	list := make([]*NamedFileSet, 0)
	for _, lib := range libraries.List() {
		libPath := lib.Path()
		if refs := scanForNamedFileSets(os.DirFS(libPath), SettingsDirName, libPath, extensions, omitDuplicateNames, set); len(refs) != 0 {
			list = append(list, &NamedFileSet{
				Name: lib.Data().Title,
				List: refs,
			})
		}
	}
	if builtIn != nil {
		if refs := scanForNamedFileSets(builtIn, builtInDir, "", extensions, omitDuplicateNames, set); len(refs) != 0 {
			list = append(list, &NamedFileSet{
				Name: i18n.Text("Built-in"),
				List: refs,
			})
		}
	}
	return list
}

// scanForNamedFileSets collects the files under dirPath within fileSystem that have one of the extensions. When
// rootDiskPath is not empty, fileSystem is rooted at that directory on disk and each reference records where the file
// lives there; when it is empty, the files are embedded and have no disk location.
func scanForNamedFileSets(fileSystem fs.FS, dirPath, rootDiskPath string, extensions []string, omitDuplicateNames bool, set map[string]bool) []*NamedFileRef {
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
		if !d.IsDir() && extMap[strings.ToLower(path.Ext(name))] {
			shortName := xfilepath.TrimExtension(name)
			if shortLowerName := strings.ToLower(shortName); !omitDuplicateNames || !set[shortLowerName] {
				set[shortLowerName] = true
				ref := &NamedFileRef{
					Name:       shortName,
					FileSystem: fileSystem,
					FilePath:   p,
				}
				if rootDiskPath != "" {
					ref.DiskPath = filepath.Join(rootDiskPath, filepath.FromSlash(p))
				}
				list = append(list, ref)
			}
		}
		return nil
	})
	slices.SortFunc(list, func(a, b *NamedFileRef) int {
		if a.Name == b.Name {
			return xstrings.NaturalCmp(a.FilePath, b.FilePath, true)
		}
		return xstrings.NaturalCmp(a.Name, b.Name, true)
	})
	return list
}
