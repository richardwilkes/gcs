// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package library

import (
	"context"
	"net/http"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xio"
	"github.com/richardwilkes/unison/updater"
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

// gitHubGet issues a GET for the url, presenting the access token as a bearer token when one is given, and hands back
// the response only if it succeeded. The caller owns the body of a response that is returned; one that failed its
// status check has already been drained and closed.
func gitHubGet(ctx context.Context, client *http.Client, url, accessToken string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, errs.NewWithCause("unable to create request for "+url, err)
	}
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	var rsp *http.Response
	if rsp, err = client.Do(req); err != nil {
		return nil, errs.NewWithCause("unable to connect to "+url, err)
	}
	if rsp.StatusCode < 200 || rsp.StatusCode > 299 {
		xio.DiscardAndCloseIgnoringErrors(rsp.Body)
		return nil, errs.New("unexpected response code from " + url + " -> " + rsp.Status)
	}
	return rsp, nil
}

// LoadReleases loads the list of releases available from a given GitHub repo that are at least currentVersion, newest
// first, leaving out those filter, if not nil, returns true for. With useLatest, the repo's latest commit stands in for
// its releases.
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
	found, err := updater.FetchReleases(ctx, client, githubAccountName, repoName, accessToken, currentVersion, filter)
	if err != nil || len(found) == 0 {
		return nil, err
	}
	versions := make([]Release, len(found))
	for i := range found {
		versions[i] = Release{
			Version:    found[i].Version,
			Notes:      found[i].Notes,
			ZipFileURL: found[i].ZipBallURL,
		}
	}
	return versions, nil
}
