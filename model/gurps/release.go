// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"net/http"
	"slices"
	"strings"

	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xio"
)

// Release holds information about a single release of a GitHub repo.
type Release struct {
	Version     string
	Notes       string
	ZipFileURL  string
	CheckFailed bool
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
func LoadReleases(ctx context.Context, client *http.Client, githubAccountName, accessToken, repoName, currentVersion string, filter func(version, notes string) bool) ([]Release, error) {
	if githubAccountName == "" || repoName == "" {
		return nil, nil
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
	}
	if err = json.NewDecoder(rsp.Body).Decode(&releases); err != nil {
		return nil, errs.NewWithCause("unable to decode response from GitHub API "+uri, err)
	}
	for _, one := range releases {
		if strings.HasPrefix(one.TagName, "v") {
			if version := strings.TrimSpace(one.TagName[1:]); version != "" &&
				(currentVersion == version || txt.NaturalLess(currentVersion, version, true)) {
				if filter == nil || !filter(version, one.Body) {
					versions = append(versions, Release{
						Version:    version,
						Notes:      one.Body,
						ZipFileURL: one.ZipBallURL,
					})
				}
			}
		}
	}
	slices.SortFunc(versions, func(a, b Release) int { return txt.NaturalCmp(b.Version, a.Version, true) })
	if len(versions) > 1 && versions[len(versions)-1].Version == currentVersion {
		versions = versions[:len(versions)-1]
	}
	return versions, nil
}
