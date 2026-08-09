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
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xyaml"
	"github.com/richardwilkes/unison/cmd/upack/packager"
)

// repoRoot is the top-level directory of the repository, relative to this package's directory, which is where the test
// runs.
const repoRoot = "../.."

func TestMain(m *testing.M) {
	ux.RegisterExternalFileTypes()
	ux.RegisterGCSFileTypes()
	os.Exit(m.Run())
}

// TestConfigOnlyReferencesIconsThatAreGenerated verifies that the packaging configuration never points at a document
// icon that no one creates. Only the file types GCS owns have an icon generated for them, so those are the only
// entries permitted to name one.
func TestConfigOnlyReferencesIconsThatAreGenerated(t *testing.T) {
	c := check.New(t)
	cfg := buildConfig()
	c.NotEqual(0, len(cfg.FileInfo), "no file types were registered")
	owners := 0
	for _, one := range cfg.FileInfo {
		if one.Rank == "Owner" {
			owners++
			c.NotEqual("", one.Icon, one.Name+" is an owned file type, but has no icon")
		} else {
			c.Equal("", one.Icon, one.Name+" is not an owned file type, so it has no icon generated for it")
		}
	}
	c.NotEqual(0, owners, "no owned file types were registered")
}

// TestCommittedPackagingIconsExist verifies that every icon named by the committed packaging.yml is actually present
// in the repository. A dangling path here breaks packaging for anything that consumes it.
func TestCommittedPackagingIconsExist(t *testing.T) {
	c := check.New(t)
	var cfg packager.Config
	c.NoError(xyaml.Load(filepath.Join(repoRoot, "packaging.yml"), &cfg))
	c.NotEqual("", cfg.AppIcon)
	checkFileExists(c, cfg.AppIcon)
	for _, one := range cfg.FileInfo {
		if one.Icon != "" {
			checkFileExists(c, one.Icon)
		}
	}
}

// TestGenerateDocIcons verifies that a document icon is produced for every file type GCS owns, and that each one is a
// usable image of the same size as the document image they are built on top of. Being an ordinary test, it also pins
// down that the generation needs no UI: it runs on a test goroutine with unison never started and no window in
// existence, which is only possible because the SVGs are rendered by canvas' CPU rasterizer.
func TestGenerateDocIcons(t *testing.T) {
	c := check.New(t)
	docImg, _, err := image.Decode(bytes.NewBuffer(docImgBytes))
	c.NoError(err)
	dir := t.TempDir()
	c.NoError(generateDocIcons(dir))
	generated := 0
	for _, fi := range gurps.KnownFileTypes {
		if !fi.IsGCSData {
			continue
		}
		generated++
		name := docIconName(fi)
		var f *os.File
		f, err = os.Open(filepath.Join(dir, name))
		c.NoError(err, name+" was not generated")
		var img image.Image
		img, err = png.Decode(f)
		c.NoError(err, name+" is not a usable PNG")
		c.NoError(f.Close())
		c.Equal(docImg.Bounds(), img.Bounds(), name+" is not the same size as the document image")
	}
	c.NotEqual(0, generated, "no owned file types were registered")
	var entries []os.DirEntry
	entries, err = os.ReadDir(dir)
	c.NoError(err)
	c.Equal(generated, len(entries), "icons were generated for file types GCS doesn't own")
}

// TestRemoveGeneratedDocIconsReportsUnreadableDir verifies that a missing icons directory is reported rather than
// panicking on the nil directory entry that fs.WalkDir hands the walk function when it can't read the root.
func TestRemoveGeneratedDocIconsReportsUnreadableDir(t *testing.T) {
	c := check.New(t)
	c.HasError(removeGeneratedDocIcons(filepath.Join(t.TempDir(), "does_not_exist")))
}

// TestRemoveGeneratedDocIconsRemovesOnlyDocIcons verifies that the cleanup pass removes the generated document icons
// and leaves everything else alone.
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

func checkFileExists(c check.Checker, relPath string) {
	_, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(relPath)))
	c.NoError(err, relPath+" is referenced by packaging.yml but does not exist")
}
