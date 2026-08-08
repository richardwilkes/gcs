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
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestMarkdownDockableNotModifiedWhenOpened verifies that opening a markdown file never reports the dockable as
// modified, regardless of the line endings the file was stored with. The content is normalized to LF when loaded, so
// the copy retained for comparison must be normalized as well; otherwise a file saved with CRLF (or CR) line endings
// would be flagged as dirty the instant it was opened, marking the tab modified and prompting to save changes on close
// even though nothing was edited.
func TestMarkdownDockableNotModifiedWhenOpened(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name string
		data string
	}{
		{name: "LF", data: "# Title\n\nSome text.\n"},
		{name: "CRLF", data: "# Title\r\n\r\nSome text.\r\n"},
		{name: "CR", data: "# Title\r\rSome text.\r"},
		{name: "mixed", data: "# Title\r\n\nSome text.\r"},
		{name: "empty", data: ""},
	} {
		p := filepath.Join(t.TempDir(), "notes.md")
		c.NoError(os.WriteFile(p, []byte(tc.data), 0o600), tc.name)
		d, err := newMarkdownDockable(p, "", true, false)
		c.NoError(err, tc.name)
		c.False(d.Modified(), "%s: a freshly opened markdown file must not be reported as modified", tc.name)
	}
}

// TestMarkdownDockableWithContentNotModifiedWhenOpened verifies the same for content-only markdown dockables, which
// take their text directly rather than reading it from a file.
func TestMarkdownDockableWithContentNotModifiedWhenOpened(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name    string
		content string
	}{
		{name: "LF", content: "# Title\n\nSome text.\n"},
		{name: "CRLF", content: "# Title\r\n\r\nSome text.\r\n"},
		{name: "CR", content: "# Title\r\rSome text.\r"},
	} {
		dockable, err := NewMarkdownDockableWithContent(tc.name, tc.content, true, false)
		c.NoError(err, tc.name)
		d, ok := dockable.(*MarkdownDockable)
		c.True(ok, tc.name)
		c.False(d.Modified(), "%s: a freshly opened markdown dockable must not be reported as modified", tc.name)
	}
}

// TestMarkdownDockableModifiedAfterEdit verifies that the normalization of the retained original does not defeat the
// modified check itself: an actual edit must still be reported.
func TestMarkdownDockableModifiedAfterEdit(t *testing.T) {
	c := check.New(t)
	p := filepath.Join(t.TempDir(), "notes.md")
	c.NoError(os.WriteFile(p, []byte("# Title\r\n\r\nSome text.\r\n"), 0o600))
	d, err := newMarkdownDockable(p, "", true, false)
	c.NoError(err)
	c.False(d.Modified())
	d.content += "More text.\n"
	c.True(d.Modified(), "an edited markdown file must be reported as modified")
}
