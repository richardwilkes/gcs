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
	"github.com/richardwilkes/unison"
)

// TestMarkdownDockableNotModifiedWhenOpened verifies that opening markdown, whether read from a file or given as
// content, never reports the dockable as modified, whatever its line endings, while an edit is still reported. The
// content is normalized to LF when loaded, so the copy retained for comparison must be too, or a file with CRLF (or CR)
// line endings would prompt to save on close though nothing was edited.
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
		fromFile, err := NewMarkdownDockable(p, true, false)
		c.NoError(err, tc.name)
		for _, opened := range []struct {
			how      string
			dockable unison.Dockable
		}{
			{how: "file", dockable: fromFile},
			{how: "content", dockable: NewMarkdownDockableWithContent(tc.name, tc.data, true, false)},
		} {
			d, ok := opened.dockable.(*MarkdownDockable)
			c.True(ok, "%s %s: expected a markdown dockable", tc.name, opened.how)
			if !ok {
				continue
			}
			c.False(d.Modified(), "%s %s: a freshly opened markdown dockable must not be reported as modified", tc.name,
				opened.how)
			d.content += "More text.\n"
			c.True(d.Modified(), "%s %s: an edited markdown dockable must be reported as modified", tc.name, opened.how)
		}
	}
}

// TestMarkdownDockableResolvesRelativePathsAgainstItsFile verifies that a markdown dockable opened from a file resolves
// the relative links and images in its content against the file's own folder, while content that is not in a file
// resolves them against the current directory. The working directory is captured when the dockable is built, before
// its content is set and its images are first retrieved, so capturing it before the backing file path is known sends
// every relative image to the wrong folder.
func TestMarkdownDockableResolvesRelativePathsAgainstItsFile(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "notes.md")
	c.NoError(os.WriteFile(p, []byte("![icon](images/icon.svg)\n"), 0o600))
	for _, allowEditing := range []bool{true, false} {
		opened, err := NewMarkdownDockable(p, allowEditing, false)
		c.NoError(err, "allowEditing=%v", allowEditing)
		d, ok := opened.(*MarkdownDockable)
		c.True(ok, "allowEditing=%v: expected a markdown dockable", allowEditing)
		if !ok {
			continue
		}
		c.Equal(dir, d.markdown.ClientData()[WorkingDirKey],
			"allowEditing=%v: a markdown file must resolve relative paths against its own folder", allowEditing)
	}
	d, ok := NewMarkdownDockableWithContent("Notes", "![icon](images/icon.svg)\n", false, false).(*MarkdownDockable)
	c.True(ok, "expected a markdown dockable")
	if ok {
		c.Equal(".", d.markdown.ClientData()[WorkingDirKey],
			"markdown content that is not in a file must resolve relative paths against the current directory")
	}
}
