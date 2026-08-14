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
	"context"
	"encoding/json/v2"
	"net/http"
	"slices"
	"strings"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xio"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

// ReleaseAsset holds information about a single file attached to a GitHub release. These are the packaged
// distributions the release workflow uploads, as opposed to the source archives GitHub generates on its own.
type ReleaseAsset struct {
	Name   string
	URL    string
	Digest string
	Size   int64
}

// SHA256 returns the hex-encoded SHA-256 of the asset, or an empty string if GitHub didn't supply one. The digest
// field was only added to the API in 2025, so assets from older releases don't have it.
func (a *ReleaseAsset) SHA256() string {
	return strings.TrimPrefix(a.Digest, "sha256:")
}

// Release holds information about a single release of a GitHub repo.
type Release struct {
	Version     string
	Notes       string
	ZipFileURL  string
	Assets      []ReleaseAsset
	CheckFailed bool
}

// Asset returns the asset with the given name, if present. Matching is case-insensitive, since the names come from
// the API rather than the local filesystem.
func (r *Release) Asset(name string) (ReleaseAsset, bool) {
	for _, one := range r.Assets {
		if strings.EqualFold(one.Name, name) {
			return one, true
		}
	}
	return ReleaseAsset{}, false
}

// HasUpdate returns true if there is an update available.
func (r *Release) HasUpdate() bool {
	return !r.CheckFailed && r.Version != ""
}

// HasReleaseNotes returns true if there are release notes available.
func (r *Release) HasReleaseNotes() bool {
	return !r.CheckFailed && r.Notes != ""
}

// LoadReleases loads the list of releases available from a given GitHub repo.
func LoadReleases(ctx context.Context, client *http.Client, githubAccountName, accessToken, repoName, currentVersion string, filter func(version, notes string) bool, useLatest bool) ([]Release, error) {
	if githubAccountName == "" || repoName == "" {
		return nil, nil
	}
	if useLatest {
		commit, err := discoverLatestCommit(ctx, "https://github.com/"+githubAccountName+"/"+repoName+".git",
			accessToken)
		if err != nil {
			return nil, err
		}
		return []Release{
			{
				Version: commit,
				Notes:   i18n.Text("Latest commit from repository. No release notes available. No data version compatibility guarantees."),
			},
		}, nil
	}
	var versions []Release
	uri := "https://api.github.com/repos/" + githubAccountName + "/" + repoName + "/releases"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, http.NoBody)
	if err != nil {
		return nil, errs.NewWithCause("unable to create GitHub API request "+uri, err)
	}
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	var rsp *http.Response
	if rsp, err = client.Do(req); err != nil {
		return nil, errs.NewWithCause("GitHub API request failed "+uri, err)
	}
	defer xio.DiscardAndCloseIgnoringErrors(rsp.Body)
	if rsp.StatusCode < 200 || rsp.StatusCode > 299 {
		return nil, errs.New("unexpected response code from GitHub API " + uri + " -> " + rsp.Status)
	}
	var releases []struct {
		TagName    string `json:"tag_name"`
		Body       string `json:"body"`
		ZipBallURL string `json:"zipball_url"`
		Assets     []struct {
			Name   string `json:"name"`
			URL    string `json:"browser_download_url"`
			Digest string `json:"digest"`
			Size   int64  `json:"size"`
		} `json:"assets"`
	}
	if err = json.UnmarshalRead(rsp.Body, &releases); err != nil {
		return nil, errs.NewWithCause("unable to decode response from GitHub API "+uri, err)
	}
	for _, one := range releases {
		if strings.HasPrefix(one.TagName, "v") {
			if version := strings.TrimSpace(one.TagName[1:]); version != "" &&
				(currentVersion == version || xstrings.NaturalLess(currentVersion, version, true)) {
				if filter == nil || !filter(version, one.Body) {
					var assets []ReleaseAsset
					if len(one.Assets) != 0 {
						assets = make([]ReleaseAsset, len(one.Assets))
						for i, asset := range one.Assets {
							assets[i] = ReleaseAsset{
								Name:   asset.Name,
								URL:    asset.URL,
								Digest: asset.Digest,
								Size:   asset.Size,
							}
						}
					}
					versions = append(versions, Release{
						Version:    version,
						Notes:      one.Body,
						ZipFileURL: one.ZipBallURL,
						Assets:     assets,
					})
				}
			}
		}
	}
	slices.SortFunc(versions, func(a, b Release) int { return xstrings.NaturalCmp(b.Version, a.Version, true) })
	if len(versions) > 1 && versions[len(versions)-1].Version == currentVersion {
		versions = versions[:len(versions)-1]
	}
	return versions, nil
}
