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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xio"
)

// An access token travels as a bearer token and its absence sends no authorization at all, since GitHub answers a
// malformed header with a 401 that reads like a bad token. A failed request hands back no response, so a caller can't
// be left holding a body it must close.
func TestGitHubGetPresentsTokenAsBearer(t *testing.T) {
	c := check.New(t)
	var status atomic.Int64
	status.Store(http.StatusOK)
	var authorization atomic.Pointer[string]
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		authorization.Store(&h)
		w.WriteHeader(int(status.Load()))
		fmt.Fprint(w, "ok")
	}))
	defer srv.Close()

	for _, token := range []string{"", "secret"} {
		rsp, err := gitHubGet(t.Context(), srv.Client(), srv.URL, token)
		c.NoError(err, "token %q", token)
		if err != nil {
			continue
		}
		data, err := io.ReadAll(rsp.Body)
		xio.CloseIgnoringErrors(rsp.Body)
		c.NoError(err, "token %q", token)
		c.Equal("ok", string(data), "token %q", token)
		want := ""
		if token != "" {
			want = "Bearer " + token
		}
		c.Equal(want, *authorization.Load(), "token %q", token)
	}

	status.Store(http.StatusUnauthorized)
	rsp, err := gitHubGet(t.Context(), srv.Client(), srv.URL, "secret")
	c.HasError(err)
	c.Nil(rsp)
}

// A library update downloads the release's source archive, so the archive's URL has to survive the hand-off from the
// GitHub listing along with the version and notes, newest release first. With nothing on offer the result is nil, which
// is what tells the caller there is no update.
func TestLoadReleasesMapsTheGitHubListing(t *testing.T) {
	c := check.New(t)
	body := `[
		{"tag_name": "v1.5.0", "body": "Older", "zipball_url": "https://example.com/zipball/1.5.0"},
		{"tag_name": "v2.0.0", "body": "Newer", "zipball_url": "https://example.com/zipball/2.0.0"}
	]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
	defer srv.Close()
	client := &http.Client{Transport: &redirectingTransport{host: srv.Listener.Addr().String()}}

	releases, err := LoadReleases(t.Context(), client, "someone", "", "repo", "1.0.0", nil, false)
	c.NoError(err)
	c.Equal([]Release{
		{Version: "2.0.0", Notes: "Newer", ZipFileURL: "https://example.com/zipball/2.0.0"},
		{Version: "1.5.0", Notes: "Older", ZipFileURL: "https://example.com/zipball/1.5.0"},
	}, releases)

	releases, err = LoadReleases(t.Context(), client, "someone", "", "repo", "3.0.0", nil, false)
	c.NoError(err)
	c.Nil(releases)
}
