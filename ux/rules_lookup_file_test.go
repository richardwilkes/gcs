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
	"errors"
	"testing"
	"time"

	"github.com/richardwilkes/toolbox/v2/check"
)

// rulesLookupHandshakeTimeout is how long these tests wait for the background goroutine before giving up. It is far
// longer than anything that should ever be needed, since it exists only so that a broken hand-off fails the test rather
// than hanging it.
const rulesLookupHandshakeTimeout = 10 * time.Second

// awaitRulesLookupResult receives the download's result, failing rather than hanging if it never arrives.
func awaitRulesLookupResult(t *testing.T, resultChan <-chan rulesLookupResult) rulesLookupResult {
	t.Helper()
	select {
	case result := <-resultChan:
		return result
	case <-time.After(rulesLookupHandshakeTimeout):
		t.Fatal("timed out waiting for the rules lookup download result")
		return rulesLookupResult{}
	}
}

// TestRunRulesLookupDownloadReportsFailure verifies that a failed download is reported as a failure. The UI thread
// receives the result only after its modal loop has been stopped, which is the moment the old code could still be
// racing with the goroutine that produced it and read back a nil error -- which would have written out an empty notes
// file instead of reporting the problem.
func TestRunRulesLookupDownloadReportsFailure(t *testing.T) {
	c := check.New(t)
	want := errors.New("download failed")
	resultChan := make(chan rulesLookupResult, 1)
	finished := make(chan struct{})
	go runRulesLookupDownload(resultChan, func() (map[string][]*rule, error) { return nil, want },
		func() { close(finished) })
	select {
	case <-finished:
	case <-time.After(rulesLookupHandshakeTimeout):
		t.Fatal("timed out waiting for the rules lookup download to finish")
	}
	result := awaitRulesLookupResult(t, resultChan)
	c.Equal(want, result.err)
	c.Equal(0, len(result.rules), "a failed download must not yield any rules")
}

// TestRunRulesLookupDownloadReportsSuccess verifies the same hand-off for a download that worked, including that the
// rules themselves make it across.
func TestRunRulesLookupDownloadReportsSuccess(t *testing.T) {
	c := check.New(t)
	want := map[string][]*rule{"B": {{Rule: "Dodge", Book: "B", Page: "374"}}}
	resultChan := make(chan rulesLookupResult, 1)
	go runRulesLookupDownload(resultChan, func() (map[string][]*rule, error) { return want, nil }, func() {})
	result := awaitRulesLookupResult(t, resultChan)
	c.NoError(result.err)
	c.Equal(want, result.rules)
}

// TestRunRulesLookupDownloadDeliversResultBeforeFinishing is the invariant the fix rests on: finish is what ultimately
// stops the modal loop, and the UI thread reads the result as soon as that loop exits, so the result has to already be
// in the channel by the time finish runs. The receive here is non-blocking on purpose -- an empty channel means the UI
// thread could have been released with nothing to read.
func TestRunRulesLookupDownloadDeliversResultBeforeFinishing(t *testing.T) {
	c := check.New(t)
	want := errors.New("download failed")
	resultChan := make(chan rulesLookupResult, 1)
	var (
		delivered bool
		got       rulesLookupResult
	)
	done := make(chan struct{})
	go runRulesLookupDownload(resultChan, func() (map[string][]*rule, error) { return nil, want }, func() {
		select {
		case got = <-resultChan:
			delivered = true
		default:
		}
		close(done)
	})
	select {
	case <-done:
	case <-time.After(rulesLookupHandshakeTimeout):
		t.Fatal("timed out waiting for the rules lookup download to finish")
	}
	c.True(delivered, "the result must be in the channel before the modal loop can be stopped")
	c.Equal(want, got.err)
}

// TestRunRulesLookupDownloadDoesNotBlockWithoutAReceiver verifies that the goroutine completes even though nothing is
// receiving yet. The UI thread cannot receive until its modal loop has been stopped, and the loop is only stopped by
// finish, so a send that blocked would deadlock the application.
func TestRunRulesLookupDownloadDoesNotBlockWithoutAReceiver(t *testing.T) {
	c := check.New(t)
	resultChan := make(chan rulesLookupResult, 1)
	finished := make(chan struct{})
	go runRulesLookupDownload(resultChan, func() (map[string][]*rule, error) { return nil, nil },
		func() { close(finished) })
	select {
	case <-finished:
	case <-time.After(rulesLookupHandshakeTimeout):
		t.Fatal("the rules lookup download goroutine blocked before finishing")
	}
	c.NoError(awaitRulesLookupResult(t, resultChan).err)
}

// TestParseRulesLookupData verifies that the downloaded data is grouped by book, that a UTF-8 BOM is tolerated and that
// malformed data is reported as an error rather than silently yielding an empty set of rules.
func TestParseRulesLookupData(t *testing.T) {
	c := check.New(t)

	t.Run("groups by book", func(_ *testing.T) {
		rules, err := parseRulesLookupData([]byte(`[
			{"Rule":"Dodge","Category":["Combat"],"Book":"B","Page":"374","Link":"B374"},
			{"Rule":"Parry","Category":["Combat"],"Book":"B","Page":"376","Link":"B376"},
			{"Rule":"Bless","Category":["Spells"],"Book":"M","Page":"37","Link":"M37"}
		]`))
		c.NoError(err)
		c.Equal(2, len(rules))
		c.Equal(2, len(rules["B"]))
		c.Equal(1, len(rules["M"]))
		c.Equal("Dodge", rules["B"][0].Rule)
		c.Equal("B374", rules["B"][0].Link)
		c.Equal([]string{"Spells"}, rules["M"][0].Category)
	})

	t.Run("strips a byte order mark", func(_ *testing.T) {
		rules, err := parseRulesLookupData(append([]byte("\xef\xbb\xbf"),
			[]byte(`[{"Rule":"Dodge","Book":"B","Page":"374","Link":"B374"}]`)...))
		c.NoError(err)
		c.Equal(1, len(rules["B"]))
	})

	t.Run("reports malformed data", func(_ *testing.T) {
		rules, err := parseRulesLookupData([]byte("not json"))
		c.HasError(err)
		c.Equal(0, len(rules), "a parse failure must not yield any rules")
	})
}
