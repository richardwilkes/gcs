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
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// repoRoot is where the generated files live, relative to this package's directory, which is where the test runs.
const repoRoot = "../.."

// TestGeneratedFilesAreUpToDate regenerates each enum from the template and compares it against the file committed to
// the repository. A failure means either the template or the enum definitions changed without the generated code being
// regenerated ("go generate ./cmd/enumgen/main.go").
func TestGeneratedFilesAreUpToDate(t *testing.T) {
	c := check.New(t)
	for _, one := range allEnums {
		relPath := path.Join(one.Pkg, one.Name+genSuffix)
		generated, err := generateEnumSource(one)
		c.NoError(err, relPath)
		var committed []byte
		committed, err = os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(relPath)))
		c.NoError(err, relPath)
		c.Equal(string(committed), string(generated),
			relPath+" is out of date; run 'go generate ./cmd/enumgen/main.go'")
	}
}

// TestNoOrphanedGeneratedFiles verifies that every committed *_gen.go file corresponds to an entry in allEnums. The
// generator removes all of them before regenerating, so one that no longer has a definition behind it would simply
// vanish on the next run -- and template special-cases written for such a file can never fire.
func TestNoOrphanedGeneratedFiles(t *testing.T) {
	c := check.New(t)
	expected := make(map[string]bool, len(allEnums))
	for _, one := range allEnums {
		expected[path.Join(one.Pkg, one.Name+genSuffix)] = true
	}
	root, err := filepath.Abs(repoRoot)
	c.NoError(err)
	found := 0
	c.NoError(fs.WalkDir(os.DirFS(root), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), genSuffix) {
			found++
			c.True(expected[p], p+" has no matching entry in allEnums")
		}
		return nil
	}))
	c.Equal(len(expected), found, "the number of generated files on disk doesn't match the number of enums")
}
