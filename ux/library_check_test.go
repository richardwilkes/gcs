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
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/v2/check"
)

// releasesAPIRedirect sends requests to a test server rather than to the GitHub API host the release checks ask.
type releasesAPIRedirect struct {
	host string
}

// RoundTrip implements http.RoundTripper.
func (r *releasesAPIRedirect) RoundTrip(req *http.Request) (*http.Response, error) {
	redirected := req.Clone(req.Context())
	redirected.URL.Scheme = "http"
	redirected.URL.Host = r.host
	redirected.Host = ""
	return http.DefaultTransport.RoundTrip(redirected)
}

// checkLibraryAgainst makes the library's release check against a stand-in for the GitHub releases API that answers
// with the given body.
func checkLibraryAgainst(t *testing.T, lib *library.Library, body string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, body) }))
	t.Cleanup(srv.Close)
	client := &http.Client{Transport: &releasesAPIRedirect{host: srv.Listener.Addr().String()}}
	lib.CheckForAvailableUpgrade(t.Context(), client)
}

// TestLibraryUpdateButtonsEnabled verifies when the Library Explorer's update and release-notes buttons are usable: for
// a library with a repository before any check has found a release and after a check that couldn't reach it, since with
// the periodic checks set to Never nothing else fills in a library's releases, and after a check that found a release,
// but not after one that found none, nor for a library with no repository.
func TestLibraryUpdateButtonsEnabled(t *testing.T) {
	c := check.New(t)
	local := library.NewLibrary("Local", "", "", "local", t.TempDir())
	c.False(libraryUpdateButtonsEnabled(local), "a library without a repository has nothing to offer")

	lib := library.NewLibrary("Test", "someone", "", "repo", t.TempDir())
	c.True(libraryUpdateButtonsEnabled(lib), "an unchecked library with a repository must be offered a check")

	// A check that fails -- the context is already canceled, so no request is ever made -- must not disable the
	// buttons, since trying again is the only way to find out what the repository has.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	lib.CheckForAvailableUpgrade(ctx, &http.Client{})
	c.True(libraryUpdateButtonsEnabled(lib), "a failed check must leave the buttons usable")

	offered := library.NewLibrary("Offered", "someone", "", "offered", t.TempDir())
	checkLibraryAgainst(t, offered, fmt.Sprintf(`[{"tag_name":"v%d.1.0","body":"notes"}]`, jio.CurrentDataVersion))
	c.False(offered.NeedsUpgradeCheck(), "precondition: the check must have found the release")
	c.True(libraryUpdateButtonsEnabled(offered), "a check that found a release must leave the buttons usable")

	none := library.NewLibrary("None", "someone", "", "none", t.TempDir())
	checkLibraryAgainst(t, none, "[]")
	c.False(none.NeedsUpgradeCheck(), "precondition: the check must have reached the repository")
	c.False(libraryUpdateButtonsEnabled(none), "a check that found no release leaves nothing to offer")
}

// TestCheckLibraryReleasesSkipsLibrariesWithNothingToCheck verifies that libraries with no check to make are passed
// straight through, without the window that reports on a check being opened -- which, without a UI, it can't be.
func TestCheckLibraryReleasesSkipsLibrariesWithNothingToCheck(t *testing.T) {
	c := check.New(t)
	local := library.NewLibrary("Local", "", "", "local", t.TempDir())
	c.True(checkLibraryReleases(nil))
	c.True(checkLibraryReleases([]*library.Library{local}))
}
