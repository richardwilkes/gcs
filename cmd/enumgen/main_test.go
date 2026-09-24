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
	"github.com/richardwilkes/toolbox/v2/xos"
)

// repoRoot is the repository root, relative to this package's directory, which is where the test runs.
const repoRoot = "../.."

// TestGeneratedFilesAreUpToDate regenerates each enum and compares it against the committed file. A failure means the
// template or the enum definitions changed without "go generate ./cmd/enumgen/main.go" being run.
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
// vanish on the next run.
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
			// Skip dot-prefixed directories. The walk root arrives as ".", so exclude it.
			if p != "." && strings.HasPrefix(d.Name(), ".") {
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

// TestEnumInfoAccessors covers the branches of the value-picking accessors that the enums in allEnums don't reach: an
// enum with no values, and DefaultUnknown without DefaultLast.
func TestEnumInfoAccessors(t *testing.T) {
	c := check.New(t)
	a := &enumValue{Key: "a"}
	b := &enumValue{Key: "b", Alt: "alt", NoLocalize: true, NoLocalizeAlt: true}
	z := &enumValue{Key: "z", OldKeys: []string{"old"}, NoLocalize: true}

	empty := &enumInfo{}
	c.Nil(empty.Default())
	c.Nil(empty.First())
	c.Nil(empty.Last())
	c.Nil(empty.RealValues())
	c.False(empty.HasAlt())
	c.False(empty.HasOldKeys())
	c.False(empty.NeedI18N())

	info := &enumInfo{Values: []*enumValue{a, b, z}}
	c.Equal(a, info.Default())
	c.Equal(a, info.First())
	c.Equal(z, info.Last())
	c.Equal([]*enumValue{a, b, z}, info.RealValues())
	c.True(info.HasAlt())
	c.True(info.HasOldKeys())
	c.True(info.NeedI18N())

	info.DefaultUnknown = true
	c.Equal(a, info.Default())
	c.Equal([]*enumValue{b, z}, info.RealValues())

	info.DefaultLast = true
	c.Equal(z, info.Default())
	c.Equal([]*enumValue{a, b}, info.RealValues())

	info.DefaultUnknown = false
	c.Equal(z, info.Default())
	c.Equal([]*enumValue{a, b, z}, info.RealValues())

	// Neither value needs localizing: b has both flags set and z has no alt.
	c.False((&enumInfo{Values: []*enumValue{b, z}}).NeedI18N())
}

// TestFindRepoRoot verifies that the root is recognized by the module its go.mod declares, whatever the directory is
// called, and that a directory holding some other module, or none, is refused.
func TestFindRepoRoot(t *testing.T) {
	c := check.New(t)
	root := filepath.Join(t.TempDir(), "any-name")
	sub := filepath.Join(root, "cmd", "enumgen")
	c.NoError(os.MkdirAll(sub, 0o750))
	c.NoError(os.WriteFile(filepath.Join(root, "go.mod"), []byte("module "+modulePath+"\n\ngo 1.27.0\n"), 0o640))

	found, err := findRepoRoot(root)
	c.NoError(err)
	c.Equal(root, found)

	t.Chdir(sub)
	found, err = findRepoRoot("")
	c.NoError(err)
	c.Equal(root, found)

	other := t.TempDir()
	c.NoError(os.WriteFile(filepath.Join(other, "go.mod"), []byte("module example.com/other\n"), 0o640))
	_, err = findRepoRoot(other)
	c.HasError(err)
	_, err = findRepoRoot(t.TempDir())
	c.HasError(err)
}

// TestRemoveExistingGenFilesLeavesOtherCheckoutsAlone verifies that only this tree's generated files are removed, and
// not those of a worktree or a nested module kept inside it.
func TestRemoveExistingGenFilesLeavesOtherCheckoutsAlone(t *testing.T) {
	c := check.New(t)
	root := t.TempDir()
	write := func(rel string) string {
		p := filepath.Join(root, filepath.FromSlash(rel))
		c.NoError(os.MkdirAll(filepath.Dir(p), 0o750))
		c.NoError(os.WriteFile(p, nil, 0o640))
		return p
	}
	ours := write("model/one" + genSuffix)
	worktree := write(".claude/worktrees/w/model/one" + genSuffix)
	write("nested/go.mod")
	nested := write("nested/model/one" + genSuffix)

	removeExistingGenFiles(root)
	c.False(xos.FileExists(ours))
	c.True(xos.FileExists(worktree))
	c.True(xos.FileExists(nested))
}
