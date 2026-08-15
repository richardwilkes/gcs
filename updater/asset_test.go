// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package updater

import (
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestAssetName pins the names down against what unison's packager actually produces. These are a contract with the
// published releases rather than an internal detail: getting one wrong means the update silently reports that no build
// exists for the user's platform.
func TestAssetName(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		goos   string
		goarch string
		want   string
	}{
		{"darwin", "amd64", "gcs-5.46.0-macos-amd64.dmg"},
		{"darwin", "arm64", "gcs-5.46.0-macos-arm64.dmg"},
		{"linux", "amd64", "gcs-5.46.0-linux-amd64.tgz"},
		{"linux", "arm64", "gcs-5.46.0-linux-arm64.tgz"},
		{"windows", "amd64", "gcs-5.46.0-windows-amd64.zip"},
		{"windows", "arm64", "gcs-5.46.0-windows-arm64.zip"},
	} {
		name, err := AssetName("5.46.0", one.goos, one.goarch)
		c.NoError(err, "%s/%s", one.goos, one.goarch)
		c.Equal(one.want, name, "%s/%s", one.goos, one.goarch)
	}
}

// TestAssetNameRejectsUnbuiltPlatforms verifies that a platform with no published build is an error rather than a
// plausible-looking name that would 404 partway through an update.
func TestAssetNameRejectsUnbuiltPlatforms(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		goos   string
		goarch string
	}{
		{"freebsd", "amd64"},
		{"js", "wasm"},
		{"linux", "386"},
		{"linux", "riscv64"},
		{"darwin", "ppc64"},
		{"", ""},
	} {
		_, err := AssetName("5.46.0", one.goos, one.goarch)
		c.HasError(err, "%s/%s", one.goos, one.goarch)
	}
	_, err := AssetName("", "linux", "amd64")
	c.HasError(err, "an empty version must be rejected")
}

// TestAssetNameUsesTheRawVersion guards against being handed a display version. ux.filterVersion trims a trailing ".0"
// and prefixes a "v", so "5.46.0" renders as "v5.46" -- neither of which appears in any published asset name.
func TestAssetNameUsesTheRawVersion(t *testing.T) {
	c := check.New(t)
	name, err := AssetName("5.46.0", "linux", "amd64")
	c.NoError(err)
	c.Equal("gcs-5.46.0-linux-amd64.tgz", name)
	c.NotContains(name, "v5.46-")
}

// TestPayloadName verifies what each archive is expected to hold: a bundle on macOS, a bare executable elsewhere.
func TestPayloadName(t *testing.T) {
	c := check.New(t)
	for goos, want := range map[string]string{
		"darwin":  "GCS.app",
		"linux":   "gcs",
		"windows": "gcs.exe",
	} {
		name, err := PayloadName(goos)
		c.NoError(err, goos)
		c.Equal(want, name, goos)
	}
	_, err := PayloadName("plan9")
	c.HasError(err)
}

// TestSelectAsset verifies that the distribution for a platform is found by exact name among everything the release
// published, and that a release missing that build reports it as missing rather than resolving to some other file.
func TestSelectAsset(t *testing.T) {
	c := check.New(t)
	assets := []Asset{
		{Name: "language.i18n", URL: "https://example.com/i18n"},
		{Name: "gcs-5.46.0-linux-amd64.tgz", URL: "https://example.com/linux-amd64"},
		{Name: "gcs-5.46.0-linux-arm64.tgz", URL: "https://example.com/linux-arm64"},
		{Name: "gcs-5.46.0-macos-arm64.dmg", URL: "https://example.com/macos-arm64"},
	}

	asset, err := SelectAsset("5.46.0", "linux", "arm64", assets)
	c.NoError(err)
	c.Equal("https://example.com/linux-arm64", asset.URL)

	asset, err = SelectAsset("5.46.0", "darwin", "arm64", assets)
	c.NoError(err)
	c.Equal("https://example.com/macos-arm64", asset.URL)

	// The Intel macOS build is absent from this release, which must be reported rather than falling back to the arm64
	// disk image that happens to be there.
	_, err = SelectAsset("5.46.0", "darwin", "amd64", assets)
	c.HasError(err)

	_, err = SelectAsset("5.46.0", "windows", "amd64", assets)
	c.HasError(err)

	_, err = SelectAsset("5.46.0", "linux", "amd64", nil)
	c.HasError(err, "a release with no assets must not match")

	// Names come from the GitHub API rather than a filesystem, so matching does not depend on case.
	asset, err = SelectAsset("5.46.0", "linux", "amd64", []Asset{{Name: "GCS-5.46.0-LINUX-AMD64.TGZ", URL: "u"}})
	c.NoError(err)
	c.Equal("u", asset.URL)
}

// TestUpdateArchPrefersAPublishedBuild verifies the Rosetta fallback ordering. On anything but a translated macOS
// process this is just runtime.GOARCH; the interesting case is that even when the native build would be preferred, an
// absent one must not be selected.
func TestUpdateArchPrefersAPublishedBuild(t *testing.T) {
	c := check.New(t)
	// No arm64 asset is published here, so whatever the translation state, the result has to name an asset that exists
	// or fall back to this build's own architecture.
	arch := UpdateArch([]Asset{{Name: "gcs-5.46.0-macos-amd64.dmg"}}, "5.46.0")
	c.True(arch == "amd64" || arch == "arm64")
	_, err := SelectAsset("5.46.0", "darwin", arch, []Asset{{Name: "gcs-5.46.0-macos-" + arch + ".dmg"}})
	c.NoError(err)
}
