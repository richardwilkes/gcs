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
	"bytes"
	"os"
	"path"
	"path/filepath"
	"slices"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xos"
)

// repoRoot is the repository root, relative to this package's directory, which is where the test runs.
const repoRoot = "../.."

// TestGeneratedFilesAreUpToDate regenerates each enum and compares it against the committed file. A failure means the
// template or the enum definitions changed without "go generate ./cmd/enumgen/main.go" being run. Only the path is
// reported, since regenerating and reading the resulting diff shows what changed.
func TestGeneratedFilesAreUpToDate(t *testing.T) {
	for _, one := range allEnums {
		relPath := path.Join(one.Pkg, one.Name+genSuffix)
		generated, err := generateEnumSource(one)
		if err != nil {
			t.Errorf("%s: %v", relPath, err)
			continue
		}
		var committed []byte
		if committed, err = os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(relPath))); err != nil {
			t.Errorf("%s: %v", relPath, err)
			continue
		}
		if !bytes.Equal(committed, generated) {
			t.Errorf("%s is out of date; run 'go generate ./cmd/enumgen/main.go'", relPath)
		}
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
	c.NoError(walkGenFiles(repoRoot, func(relPath string) error {
		c.True(expected[relPath], relPath+" has no matching entry in allEnums")
		return nil
	}))
}

// TestEnumInfoAccessors covers what the enums in allEnums don't reach: an enum with no values, and a value that isn't
// localized but has an alt.
func TestEnumInfoAccessors(t *testing.T) {
	c := check.New(t)
	empty := &enumInfo{}
	c.Nil(empty.Default())
	c.Nil(empty.First())
	c.Nil(empty.Last())
	c.False(empty.HasAlt())
	c.False(empty.HasOldKeys())
	c.False(empty.NeedI18N())

	needI18N := func(v *enumValue) bool { return (&enumInfo{Values: []*enumValue{v}}).NeedI18N() }
	c.True(needI18N(&enumValue{Key: "k", Alt: "alt", NoLocalize: true}))
	c.False(needI18N(&enumValue{Key: "k", Alt: "alt", NoLocalize: true, NoLocalizeAlt: true}))
}

// TestDefaultSelection verifies how DefaultKey, DefaultLast and DefaultUnknown pick the default and order the
// constants, while Values keeps declaration order and RealValues leaves out the default only with DefaultUnknown. The
// two combinations the enums in allEnums use, no flags and DefaultLast with DefaultUnknown, are left to
// TestGeneratedFilesAreUpToDate.
func TestDefaultSelection(t *testing.T) {
	low := &enumValue{Key: "low"}
	normal := &enumValue{Key: "normal"}
	high := &enumValue{Key: "high"}
	higher := &enumValue{Key: "higher"}
	values := []*enumValue{low, normal, high, higher}
	normalFirst := []*enumValue{normal, low, high, higher}
	normalLast := []*enumValue{low, high, higher, normal}
	withoutNormal := []*enumValue{low, high, higher}
	for _, tc := range []struct {
		name    string
		key     string
		last    bool
		unknown bool
		consts  []*enumValue
		def     *enumValue
		listed  []*enumValue
	}{
		{name: "unknown", unknown: true, consts: values, def: low, listed: []*enumValue{normal, high, higher}},
		{name: "last", last: true, consts: values, def: higher, listed: values},
		{name: "key", key: "normal", consts: normalFirst, def: normal, listed: values},
		{name: "key already first", key: "low", consts: values, def: low, listed: values},
		{name: "key, last", key: "normal", last: true, consts: normalLast, def: normal, listed: values},
		{name: "key, unknown", key: "normal", unknown: true, consts: normalFirst, def: normal, listed: withoutNormal},
		{
			name: "key, last, unknown", key: "normal", last: true, unknown: true,
			consts: normalLast, def: normal, listed: withoutNormal,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			info := &enumInfo{
				Values:         slices.Clone(values),
				DefaultKey:     tc.key,
				DefaultLast:    tc.last,
				DefaultUnknown: tc.unknown,
			}
			c.Equal(tc.consts, info.ConstValues())
			c.Equal(tc.def, info.Default())
			c.Equal(tc.consts[0], info.First())
			c.Equal(tc.consts[len(tc.consts)-1], info.Last())
			c.Equal(tc.listed, info.RealValues())
			c.Equal(values, info.Values)
		})
	}
}

// TestDefaultKey verifies that the generated source declares the constant of the value DefaultKey names first, while
// the slice of values and the groups keep declaration order, and that a DefaultKey which names no value is refused.
func TestDefaultKey(t *testing.T) {
	c := check.New(t)
	info := &enumInfo{
		Pkg:        "test/level",
		Name:       "level",
		Desc:       "is a level",
		DefaultKey: "normal",
		Values: []*enumValue{
			{Key: "low", Groups: []string{"g"}},
			{Key: "normal", Groups: []string{"g"}},
			{Key: "high"},
			{Key: "higher", Groups: []string{"g"}},
		},
	}
	c.Equal(map[string][]string{"g": {"Low", "Normal", "Higher"}}, info.Groups())

	src, err := generateEnumSource(info)
	c.NoError(err)
	c.Contains(string(src), "Normal Level = iota\n\tLow\n")
	c.Contains(string(src), "var Levels = []Level{\n\tLow,\n\tNormal,\n\tHigh,\n\tHigher,\n}")

	info.DefaultKey = "missing"
	_, err = generateEnumSource(info)
	c.HasError(err)
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
