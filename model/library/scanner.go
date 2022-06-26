/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package library

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
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
func ScanForNamedFileSets(builtIn fs.FS, builtInDir, extension string, omitDuplicateNames bool, libraries Libraries) []*NamedFileSet {
	set := make(map[string]bool)
	list := make([]*NamedFileSet, 0)
	for _, lib := range libraries.List() {
		if refs := scanForNamedFileSets(os.DirFS(lib.Path()), "Settings", extension, omitDuplicateNames, set); len(refs) != 0 {
			list = append(list, &NamedFileSet{
				Name: lib.Title,
				List: refs,
			})
		}
	}
	if builtIn != nil {
		if refs := scanForNamedFileSets(builtIn, builtInDir, extension, omitDuplicateNames, set); len(refs) != 0 {
			list = append(list, &NamedFileSet{
				Name: i18n.Text("Built-in"),
				List: refs,
			})
		}
	}
	return list
}

func scanForNamedFileSets(fileSystem fs.FS, dirPath, extension string, omitDuplicateNames bool, set map[string]bool) []*NamedFileRef {
	entries, err := fs.ReadDir(fileSystem, dirPath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			jot.Error(errs.Wrap(err))
		}
		return nil
	}
	list := make([]*NamedFileRef, 0)
	for _, entry := range entries {
		name := entry.Name()
		if strings.EqualFold(path.Ext(name), extension) {
			shortName := xfs.TrimExtension(name)
			if shortLowerName := strings.ToLower(shortName); !omitDuplicateNames || !set[shortLowerName] {
				set[shortLowerName] = true
				list = append(list, &NamedFileRef{
					Name:       shortName,
					FileSystem: fileSystem,
					FilePath:   path.Join(dirPath, name),
				})
			}
		}
	}
	return list
}
