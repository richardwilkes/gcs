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
	"image"
	"image/png"
	"os"
	"path"
	"path/filepath"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestMain(m *testing.M) {
	ux.RegisterExternalFileTypes()
	ux.RegisterGCSFileTypes()
	os.Exit(m.Run())
}

// TestGenerateDocIcons verifies that the document icons generated are exactly the ones the packaging configuration
// names, and that each is a usable image of the same size as the document image it is built on top of. It also pins
// down that generation needs no UI: unison is never started and no window exists, which works only because the SVGs
// are rendered by canvas' CPU rasterizer.
func TestGenerateDocIcons(t *testing.T) {
	c := check.New(t)
	docCfg, _, err := image.DecodeConfig(bytes.NewReader(docImgBytes))
	c.NoError(err)
	dir := t.TempDir()
	c.NoError(generateDocIcons(dir))
	var want []string
	for _, one := range buildConfig().FileInfo {
		if one.Icon != "" {
			want = append(want, path.Base(one.Icon))
		}
	}
	slices.Sort(want)
	c.NotEqual(0, len(want), "the configuration names no document icons")
	var entries []os.DirEntry
	entries, err = os.ReadDir(dir)
	c.NoError(err)
	got := make([]string, len(entries))
	for i, one := range entries {
		got[i] = one.Name()
	}
	c.Equal(want, got, "the generated icons are not the ones the configuration names")
	for _, name := range got {
		var data []byte
		data, err = os.ReadFile(filepath.Join(dir, name))
		c.NoError(err)
		var cfg image.Config
		cfg, err = png.DecodeConfig(bytes.NewReader(data))
		c.NoError(err, name+" is not a usable PNG")
		c.Equal(image.Pt(docCfg.Width, docCfg.Height), image.Pt(cfg.Width, cfg.Height),
			name+" is not the same size as the document image")
	}
}

// TestRemoveGeneratedDocIconsReportsUnreadableDir verifies that a missing icons directory is reported rather than
// panicking on the nil directory entry that fs.WalkDir hands the walk function when it can't read the root.
func TestRemoveGeneratedDocIconsReportsUnreadableDir(t *testing.T) {
	c := check.New(t)
	c.HasError(removeGeneratedDocIcons(filepath.Join(t.TempDir(), "does_not_exist")))
}

func TestRemoveGeneratedDocIconsRemovesOnlyDocIcons(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	c.NoError(os.Mkdir(filepath.Join(dir, "sub"), 0o755))
	for _, one := range []string{"app.png", "gcs" + docIconSuffix, filepath.Join("sub", "adq"+docIconSuffix)} {
		c.NoError(os.WriteFile(filepath.Join(dir, one), []byte("x"), 0o600))
	}
	c.NoError(removeGeneratedDocIcons(dir))
	_, err := os.Stat(filepath.Join(dir, "app.png"))
	c.NoError(err, "app.png should not have been removed")
	for _, one := range []string{"gcs" + docIconSuffix, filepath.Join("sub", "adq"+docIconSuffix)} {
		_, err = os.Stat(filepath.Join(dir, one))
		c.True(os.IsNotExist(err), one+" should have been removed")
	}
}
