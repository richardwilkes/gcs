// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// loadReleasesFrom runs LoadReleases against a test server serving the given GitHub API response body, rather than
// against the real api.github.com host baked into it.
func loadReleasesFrom(t *testing.T, body string) ([]Release, error) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
	t.Cleanup(srv.Close)
	client := &http.Client{Transport: &redirectingTransport{host: srv.Listener.Addr().String()}}
	return LoadReleases(t.Context(), client, "someone", "", "repo", "1.0.0", nil, false)
}

// TestLoadReleasesCapturesAssets verifies that the packaged distributions attached to a release are carried through.
// The app updater downloads one of these directly, so an asset's name, URL, size and digest all have to survive the
// decode; before this, only the source zipball URL did.
func TestLoadReleasesCapturesAssets(t *testing.T) {
	c := check.New(t)
	releases, err := loadReleasesFrom(t, `[{
		"tag_name": "v2.0.0",
		"body": "Notes",
		"zipball_url": "https://example.com/zipball",
		"assets": [
			{
				"name": "gcs-2.0.0-macos-arm64.dmg",
				"browser_download_url": "https://example.com/gcs-2.0.0-macos-arm64.dmg",
				"digest": "sha256:abc123",
				"size": 29234206
			},
			{
				"name": "language.i18n",
				"browser_download_url": "https://example.com/language.i18n",
				"digest": "sha256:def456",
				"size": 94365
			}
		]
	}]`)
	c.NoError(err)
	c.Equal(1, len(releases))
	c.Equal("https://example.com/zipball", releases[0].ZipFileURL, "the library updater's field must be untouched")
	c.Equal(2, len(releases[0].Assets))

	asset := releases[0].Assets[0]
	c.Equal("gcs-2.0.0-macos-arm64.dmg", asset.Name)
	c.Equal("https://example.com/gcs-2.0.0-macos-arm64.dmg", asset.URL)
	c.Equal("sha256:abc123", asset.Digest)
	c.Equal(int64(29234206), asset.Size)
	c.Equal("abc123", asset.SHA256(), "the sha256: prefix must be stripped")
}

// TestLoadReleasesToleratesMissingAssets verifies that a release with no assets, or with assets that predate GitHub's
// digest field, still loads. Every GCS release before v5.36.0 has assets carrying no digest, and the "latest commit"
// path synthesizes a release with no assets at all, so neither case may be an error here -- refusing to update on a
// missing digest is the updater's job, not the loader's.
func TestLoadReleasesToleratesMissingAssets(t *testing.T) {
	c := check.New(t)
	releases, err := loadReleasesFrom(t, `[
		{"tag_name": "v2.0.0", "body": "", "zipball_url": "", "assets": []},
		{"tag_name": "v1.5.0", "body": "", "zipball_url": ""},
		{"tag_name": "v1.2.0", "body": "", "zipball_url": "", "assets": [
			{"name": "old.tgz", "browser_download_url": "https://example.com/old.tgz", "size": 10}
		]}
	]`)
	c.NoError(err)
	c.Equal(3, len(releases))
	c.Nil(releases[0].Assets, "an empty assets array yields no assets")
	c.Nil(releases[1].Assets, "an absent assets key yields no assets")
	c.Equal(1, len(releases[2].Assets))
	c.Equal("", releases[2].Assets[0].Digest)
	c.Equal("", releases[2].Assets[0].SHA256(), "no digest must not produce a bogus hash")
}

// TestReleaseAssetLookup verifies the by-name lookup the updater uses to find the distribution built for the running
// platform. Names come from the API rather than the filesystem, so the match is case-insensitive.
func TestReleaseAssetLookup(t *testing.T) {
	c := check.New(t)
	release := Release{Assets: []ReleaseAsset{
		{Name: "gcs-2.0.0-linux-amd64.tgz", URL: "https://example.com/linux"},
		{Name: "gcs-2.0.0-windows-amd64.zip", URL: "https://example.com/windows"},
	}}

	asset, ok := release.Asset("gcs-2.0.0-windows-amd64.zip")
	c.True(ok)
	c.Equal("https://example.com/windows", asset.URL)

	asset, ok = release.Asset("GCS-2.0.0-Linux-AMD64.TGZ")
	c.True(ok, "matching must be case-insensitive")
	c.Equal("https://example.com/linux", asset.URL)

	_, ok = release.Asset("gcs-2.0.0-macos-arm64.dmg")
	c.False(ok)

	_, ok = (&Release{}).Asset("anything")
	c.False(ok, "a release with no assets must not match")
}
