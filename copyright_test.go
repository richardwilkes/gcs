// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package main

import (
	"io/fs"
	"os"
	"path"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// copyrightMarker is a line every source file's MPL 2.0 header contains. The year range in the first line varies, so
// this line is matched instead.
const copyrightMarker = "This Source Code Form is subject to the terms of the Mozilla Public"

// TestGoSourcesHaveCopyrightHeader looks for the header within the first 512 bytes rather than at the very start, since
// generated files carry a "Code generated" line ahead of it.
func TestGoSourcesHaveCopyrightHeader(t *testing.T) {
	c := check.New(t)
	root, err := os.OpenRoot(".")
	c.NoError(err)
	defer func() { c.NoError(root.Close()) }()
	fileSystem := root.FS()
	c.NoError(fs.WalkDir(fileSystem, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if p != "." && name[0] == '.' {
				return fs.SkipDir
			}
			return nil
		}
		if path.Ext(name) != ".go" {
			return nil
		}
		var data []byte
		if data, err = fs.ReadFile(fileSystem, p); err != nil {
			return err
		}
		c.Contains(string(data[:min(len(data), 512)]), copyrightMarker, p)
		return nil
	}))
}
