// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestSettingsDockableDoLoadPrefersRefLoader verifies that doLoad hands the whole file reference to RefLoader when one
// is set, falls back to Loader with the reference's file system and path when there is no RefLoader, and reports a
// loader's failure through the workspace's error handler under the dockable's title.
func TestSettingsDockableDoLoadPrefersRefLoader(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	ref := &gurps.NamedFileRef{
		Name:       "Elf",
		FileSystem: os.DirFS(dir),
		FilePath:   "Elf.ancestry",
		DiskPath:   filepath.Join(dir, "Elf.ancestry"),
	}
	var loaderCalls, refLoaderCalls int
	var gotFS fs.FS
	var gotPath string
	var gotRef *gurps.NamedFileRef
	d := &SettingsDockable{
		TabTitle: "Ancestry",
		Loader: func(fileSystem fs.FS, filePath string) error {
			loaderCalls++
			gotFS = fileSystem
			gotPath = filePath
			return nil
		},
		RefLoader: func(ref *gurps.NamedFileRef) error {
			refLoaderCalls++
			gotRef = ref
			return nil
		},
	}

	// With both set, RefLoader wins and receives the reference itself.
	d.doLoad(ref)
	c.Equal(1, refLoaderCalls, "RefLoader is preferred")
	c.Equal(0, loaderCalls, "Loader is not called when RefLoader is set")
	c.True(gotRef == ref, "RefLoader receives the same reference")

	// With only Loader set, it receives the reference's file system and path.
	d.RefLoader = nil
	d.doLoad(ref)
	c.Equal(1, loaderCalls, "Loader is used when there is no RefLoader")
	c.Equal(ref.FileSystem, gotFS)
	c.Equal(ref.FilePath, gotPath)

	// A failing loader is reported through the workspace's error handler.
	savedHandler := Workspace.ErrorHandler
	t.Cleanup(func() { Workspace.ErrorHandler = savedHandler })
	var gotMsg string
	var gotErr error
	Workspace.ErrorHandler = func(msg string, err error) {
		gotMsg = msg
		gotErr = err
	}
	boom := errors.New("boom")
	d.RefLoader = func(_ *gurps.NamedFileRef) error { return boom }
	d.doLoad(ref)
	c.Equal("Unable to load Ancestry", gotMsg)
	c.True(errors.Is(gotErr, boom), "the loader's error is passed through")
	c.Equal(1, loaderCalls, "Loader is not called when RefLoader fails")
}

// TestSettingsDockableCanLoad verifies that a dockable can load with either form of loader, and not without one.
func TestSettingsDockableCanLoad(t *testing.T) {
	c := check.New(t)
	d := &SettingsDockable{}
	c.False(d.canLoad(), "no loader")
	d.Loader = func(_ fs.FS, _ string) error { return nil }
	c.True(d.canLoad(), "Loader only")
	d.RefLoader = func(_ *gurps.NamedFileRef) error { return nil }
	c.True(d.canLoad(), "both loaders")
	d.Loader = nil
	c.True(d.canLoad(), "RefLoader only")
}
