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
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/uti"
)

const sampleSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 16" width="48" height="32">` +
	`<rect x="0" y="0" width="24" height="16"/></svg>`

func gzipped(t *testing.T, data string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)
	if _, err := gz.Write([]byte(data)); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

// TestIsSVGPath verifies that every extension registered for SVG content routes to the SVG branch of
// NewImageDockable. ".svgz" is one of uti.SVG's extensions, so it is advertised as openable; sending it to the raster
// branch instead just fails with "unable to decode image data".
func TestIsSVGPath(t *testing.T) {
	c := check.New(t)
	for _, ext := range uti.SVG.Extensions {
		c.True(isSVGPath("image"+ext), ext+" should be recognized as SVG")
		c.True(isSVGPath("IMAGE"+strings.ToUpper(ext)), strings.ToUpper(ext)+" should be recognized as SVG")
	}
	for _, ext := range []string{".png", ".jpg", ".pdf", ".not-svg", ""} {
		c.False(isSVGPath("image"+ext), ext+" should not be recognized as SVG")
	}
	// The extension has to match in full; a name merely ending in the letters must not be treated as SVG.
	c.False(isSVGPath("image.notsvg"), "a longer extension ending in 'svg' should not be recognized as SVG")
	c.False(isSVGPath("svg"), "a bare name should not be recognized as SVG")
}

// TestLoadSVGFromFile verifies that both plain and gzip-compressed SVG content loads. ".svgz" files hold gzipped SVG
// data, which the SVG parser cannot consume directly, so it must be decompressed first.
func TestLoadSVGFromFile(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	for _, tc := range []struct {
		name string
		file string
		data []byte
	}{
		{name: "plain svg", file: "image.svg", data: []byte(sampleSVG)},
		{name: "compressed svgz", file: "image.svgz", data: gzipped(t, sampleSVG)},
		// Compressed content stored with the plain extension is common enough to be worth handling, too.
		{name: "compressed svg", file: "compressed.svg", data: gzipped(t, sampleSVG)},
	} {
		p := filepath.Join(dir, tc.file)
		c.NoError(os.WriteFile(p, tc.data, 0o600), tc.name)
		vector, err := loadSVGFromFile(p)
		c.NoError(err, tc.name)
		c.Equal(float32(24), vector.Size().Width, tc.name)
		c.Equal(float32(16), vector.Size().Height, tc.name)
		c.Equal(float32(48), vector.SuggestedSize().Width, tc.name)
		c.Equal(float32(32), vector.SuggestedSize().Height, tc.name)
	}
}

// TestLoadSVGFromFileErrors verifies that unreadable or malformed content produces an error rather than a panic.
func TestLoadSVGFromFileErrors(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()

	_, err := loadSVGFromFile(filepath.Join(dir, "does-not-exist.svgz"))
	c.HasError(err, "a missing file should produce an error")

	empty := filepath.Join(dir, "empty.svgz")
	c.NoError(os.WriteFile(empty, nil, 0o600))
	_, err = loadSVGFromFile(empty)
	c.HasError(err, "an empty file should produce an error")

	truncated := filepath.Join(dir, "truncated.svgz")
	full := gzipped(t, sampleSVG)
	c.NoError(os.WriteFile(truncated, full[:len(full)/2], 0o600))
	_, err = loadSVGFromFile(truncated)
	c.HasError(err, "truncated gzip data should produce an error")

	notSVG := filepath.Join(dir, "not-svg.svgz")
	c.NoError(os.WriteFile(notSVG, gzipped(t, "this is not svg data"), 0o600))
	_, err = loadSVGFromFile(notSVG)
	c.HasError(err, "gzipped non-SVG data should produce an error")
}
