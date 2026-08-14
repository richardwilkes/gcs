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
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/richardwilkes/toolbox/v2/check"
)

// serveBytes stands up a server handing back the given payload, and returns an Asset describing it truthfully.
func serveBytes(t *testing.T, payload []byte) (*httptest.Server, Asset) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write(payload); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(srv.Close)
	sum := sha256.Sum256(payload)
	return srv, Asset{
		Name:   "gcs-5.46.0-linux-amd64.tgz",
		URL:    srv.URL,
		SHA256: hex.EncodeToString(sum[:]),
		Size:   int64(len(payload)),
	}
}

// TestDownload verifies the happy path, including that progress is reported and ends at the full size.
func TestDownload(t *testing.T) {
	c := check.New(t)
	payload := []byte("a plausible archive of some length")
	_, asset := serveBytes(t, payload)
	dst := filepath.Join(t.TempDir(), "asset.tgz")

	var last int64
	var calls int
	c.NoError(Download(t.Context(), nil, asset, dst, func(read int64) {
		calls++
		last = read
	}))

	got, err := os.ReadFile(dst)
	c.NoError(err)
	c.Equal(payload, got)
	c.True(calls > 0, "progress must be reported")
	c.Equal(int64(len(payload)), last, "progress must finish at the full size")
}

// TestDownloadRejectsATamperedPayload is the check that matters: a body that is not what GitHub published must never
// reach the disk as if it were.
func TestDownloadRejectsATamperedPayload(t *testing.T) {
	c := check.New(t)
	_, asset := serveBytes(t, []byte("the real archive"))
	asset.SHA256 = hex.EncodeToString(make([]byte, sha256.Size))
	dst := filepath.Join(t.TempDir(), "asset.tgz")

	c.HasError(Download(t.Context(), nil, asset, dst, nil))
	c.False(exists(dst), "a failed integrity check must not leave the file behind")
}

// TestDownloadRejectsAWrongLength verifies both directions of the length check. A short body is a truncated transfer; a
// long one means the response is not the asset that was described.
func TestDownloadRejectsAWrongLength(t *testing.T) {
	c := check.New(t)
	payload := []byte("0123456789")

	for name, size := range map[string]int64{
		"claims more than it sends": int64(len(payload)) + 5,
		"claims less than it sends": int64(len(payload)) - 5,
	} {
		_, asset := serveBytes(t, payload)
		asset.Size = size
		dst := filepath.Join(t.TempDir(), "asset.tgz")
		c.HasError(Download(t.Context(), nil, asset, dst, nil), name)
		c.False(exists(dst), "%s left a file behind", name)
	}
}

// TestDownloadRejectsAnErrorResponse verifies that an HTTP error is reported as one, rather than the error page being
// written out and handed to the extractor as though it were an archive.
func TestDownloadRejectsAnErrorResponse(t *testing.T) {
	c := check.New(t)
	for _, code := range []int{http.StatusNotFound, http.StatusForbidden, http.StatusInternalServerError} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(code)
			if _, err := w.Write([]byte("<html>nope</html>")); err != nil {
				t.Error(err)
			}
		}))
		asset := Asset{Name: "asset.tgz", URL: srv.URL, SHA256: "abc", Size: 17}
		dst := filepath.Join(t.TempDir(), "asset.tgz")
		err := Download(t.Context(), srv.Client(), asset, dst, nil)
		srv.Close()
		c.HasError(err, "status %d", code)
		c.Contains(err.Error(), strconv.Itoa(code), "status %d", code)
		c.False(exists(dst), "status %d left a file behind", code)
	}
}

// TestDownloadRefusesWithoutAChecksum verifies that an unverifiable asset is never fetched at all. Preflight already
// refuses these, so reaching here would mean a caller bypassed it.
func TestDownloadRefusesWithoutAChecksum(t *testing.T) {
	c := check.New(t)
	_, asset := serveBytes(t, []byte("content"))
	asset.SHA256 = ""
	dst := filepath.Join(t.TempDir(), "asset.tgz")
	c.HasError(Download(t.Context(), nil, asset, dst, nil))
	c.False(exists(dst))
}

// TestDownloadStopsWhenCanceled verifies that the Cancel button actually stops the transfer and cleans up, rather than
// leaving a partial file that a later run might find.
func TestDownloadStopsWhenCanceled(t *testing.T) {
	c := check.New(t)
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("the beginning of the file")); err != nil {
			return
		}
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()
	defer close(release)

	ctx, cancel := context.WithCancel(t.Context())
	asset := Asset{Name: "asset.tgz", URL: srv.URL, SHA256: "abc", Size: 1 << 20}
	dst := filepath.Join(t.TempDir(), "asset.tgz")

	done := make(chan error, 1)
	go func() {
		done <- Download(ctx, srv.Client(), asset, dst, func(int64) { cancel() })
	}()

	select {
	case err := <-done:
		c.HasError(err)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for the canceled download to return")
	}
	c.False(exists(dst), "a canceled download must not leave a partial file behind")
}

// TestDownloadRefusesAnExistingDestination verifies that the destination is created exclusively, so two updates racing
// in the same staging directory cannot interleave into one file.
func TestDownloadRefusesAnExistingDestination(t *testing.T) {
	c := check.New(t)
	_, asset := serveBytes(t, []byte("content"))
	dst := filepath.Join(t.TempDir(), "asset.tgz")
	write(t, dst, "already here")
	c.HasError(Download(t.Context(), nil, asset, dst, nil))
	c.Equal("already here", read(t, dst), "the existing file must be left alone")
}
