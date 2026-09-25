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
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestBrowserRequestsAreRecordedHeadless verifies that the places GCS hands a URL to the browser go through
// unison.OpenBrowser, so that a headless session records the request for OpenedURLs() rather than launching the
// host's browser from under a test, and that no error dialog results from the recorded request.
func TestBrowserRequestsAreRecordedHeadless(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	c.Nil(screen.OpenedURLs(), "nothing has asked for the browser yet")

	var cancel bool
	c.True(screen.Do(func() { cancel = OpenPageReference("https://example.com/ref", "", nil) }))
	c.False(cancel, "a web page reference never asks to cancel further processing")
	c.Equal([]string{"https://example.com/ref"}, screen.OpenedURLs(),
		"a page reference that is a URL is handed to the browser through unison")

	c.True(screen.Do(func() { showWebPage("https://example.com/help") }))
	c.Equal([]string{"https://example.com/help"}, screen.OpenedURLs(),
		"the Help menu's web pages are handed to the browser through unison")
	c.Nil(screen.OpenedURLs(), "and each request is handed back only once")
}
