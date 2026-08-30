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
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/unison/enums/imgfmt"
)

// TestImageFileTypesGroupWithEachOther verifies that every image file type, SVG included, lists all of the other image
// extensions in its GroupWith set. unison's imgfmt enum has no SVG member, so the SVG extensions have to be added to
// the shared group explicitly. Without them, DockContainerHoldsExtension/LocateDockContainerForExtension can never
// match a dock container holding SVG files, so a second SVG opens in a new dock instead of stacking with the first.
func TestImageFileTypesGroupWithEachOther(t *testing.T) {
	c := check.New(t)
	RegisterExternalFileTypes()
	imageExts := make([]string, 0, 16)
	for _, ext := range append(imgfmt.AllReadableExtensions(), uti.SVG.Extensions...) {
		if !slices.Contains(imageExts, ext) {
			imageExts = append(imageExts, ext)
		}
	}
	for _, ext := range uti.SVG.Extensions {
		c.True(slices.Contains(imageExts, ext), ext+" should be one of the image extensions")
	}
	for _, ext := range imageExts {
		fi := gurps.FileInfoFor("image" + ext)
		c.True(fi.IsImage, ext+" should be registered as an image file type")
		for _, other := range imageExts {
			c.True(slices.Contains(fi.GroupWith, other), ext+" should group with "+other)
		}
	}
}

// TestRegisterKnownFileTypesRegistersOnce verifies that repeated calls leave the registry as the first call built it.
// The deep search content cache's worker goroutines read the registry without synchronization, so a second
// registration -- which the tests would otherwise perform whenever one of them needs the registry -- must not rewrite
// it underneath a build still running from an earlier test.
func TestRegisterKnownFileTypesRegistersOnce(t *testing.T) {
	c := check.New(t)
	RegisterKnownFileTypes()
	known := slices.Clone(gurps.KnownFileTypes)
	c.True(len(known) != 0, "the first call must populate the registry")
	sheet := gurps.FileInfoFor("x" + gurps.SheetExt)
	RegisterKnownFileTypes()
	c.Equal(len(known), len(gurps.KnownFileTypes), "a repeated call must not append to the known file types")
	for i, fi := range gurps.KnownFileTypes {
		c.True(known[i] == fi, "a repeated call must not replace entry %d", i)
	}
	c.True(sheet == gurps.FileInfoFor("x"+gurps.SheetExt), "a repeated call must not replace the registry entries")
}
