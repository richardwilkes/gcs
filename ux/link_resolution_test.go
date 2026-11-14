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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestLinkedFile verifies that a link to a file finds it with or without its ".md" extension and with the "#anchor"
// that follows its name split off, which a relative link between markdown files relies on, while a file whose own name
// holds a '#' is still found whole, and a target naming no file is left to OpenPageReference. A target ending in a
// folder rather than a name is not given the extension, which would find the page beside that folder.
func TestLinkedFile(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	c.NoError(os.MkdirAll(filepath.Join(dir, "Notes", "Sub"), 0o750))
	for _, name := range []string{"Other Page.md", "Notes.md", "Notes#1.md"} {
		c.NoError(os.WriteFile(filepath.Join(dir, name), []byte("# Title\n"), 0o600))
	}
	otherPage := filepath.Join(dir, "Other Page.md")
	for _, tc := range []struct {
		name       string
		from       string // the working directory, relative to dir
		target     string
		wantFile   string
		wantAnchor string
		wantOK     bool
	}{
		{name: "anchor split off", target: "Other%20Page#section", wantFile: otherPage, wantAnchor: "section", wantOK: true},
		{name: "extension given", target: "Other%20Page.md#section", wantFile: otherPage, wantAnchor: "section", wantOK: true},
		{name: "plain", target: "Other%20Page", wantFile: otherPage, wantOK: true},
		{name: "name holding a hash", target: "Notes#1", wantFile: filepath.Join(dir, "Notes#1.md"), wantOK: true},
		{name: "page beside a folder of its name", target: "Notes", wantFile: filepath.Join(dir, "Notes.md"), wantOK: true},
		{name: "missing", target: "Missing#section"},
		{name: "empty", from: "Notes", target: ""},
		{name: "working folder at an anchor", from: "Notes", target: "./#x"},
		{name: "folder above", from: filepath.Join("Notes", "Sub"), target: ".."},
		{name: "folder with a trailing slash", target: "Notes/"},
	} {
		t.Run(tc.name, func(_ *testing.T) {
			filePath, anchor, ok := linkedFile(filepath.Join(dir, tc.from), tc.target)
			c.Equal(tc.wantOK, ok, tc.name)
			c.Equal(tc.wantFile, filePath, tc.name)
			c.Equal(tc.wantAnchor, anchor, tc.name)
		})
	}
}

// TestLinksOpenWhatTheyName verifies that following a link, in a markdown file or the User Guide, opens what its target
// names rather than asking for a PDF to map a key to: an md: reference is unescaped once, as a page reference holding
// it is, so that an escaped '#' or '%' is part of the file's name, a mailto: link is handed to the system through
// unison, an empty link or an anchor, which the markdown hands over only when it has no heading of that name, opens
// nothing, and a link to a file that does not exist says so, whether or not its name ends in a page number.
func TestLinksOpenWhatTheyName(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &unison.DefaultMarkdownTheme.AltLinkPrefixes, []string{"md:"})
	dir := filepath.Join(gurps.GlobalSettings().Libraries.User().Path(false), "Markdown")
	c.NoError(os.MkdirAll(dir, 0o750))
	for _, name := range []string{"Notes.md", "Notes#1.md", "A B.md", "A%20B.md"} {
		c.NoError(os.WriteFile(filepath.Join(dir, name), []byte("# Title\n"), 0o600))
	}
	for _, tc := range []struct {
		name      string
		target    string
		inGuide   bool
		wantFile  string
		wantURLs  []string
		wantAsked string // the text of the dialog the link opens, if any
	}{
		{name: "md: reference with an escaped '#'", target: "md:Notes%231", wantFile: "Notes#1.md"},
		{name: "md: reference with an escaped '%'", target: "md:A%2520B", wantFile: "A%20B.md"},
		{name: "mailto: link", target: "mailto:x@y.com", wantURLs: []string{"mailto:x@y.com"}},
		{name: "mailto: link in the User Guide", target: "mailto:x@y.com", inGuide: true, wantURLs: []string{"mailto:x@y.com"}},
		{name: "anchor with no heading", target: "#step-1"},
		{
			name:      "missing file whose name ends in a number",
			target:    "Missing%202",
			wantAsked: "Unable to open Missing 2 There is no file of that name.",
		},
		{
			name:      "missing file whose name has no page number",
			target:    "Missing.md",
			wantAsked: "Unable to open Missing.md There is no file of that name.",
		},
		{name: "empty", target: ""},
	} {
		// A prompt runs a nested modal loop, so the link is followed through Post and any dialog it opens is canceled.
		c.True(screen.Post(func() {
			if tc.inGuide {
				newUserGuideDockable().handleLink(nil, tc.target)
			} else {
				HandleLink(nil, tc.target)
			}
		}), tc.name)
		screen.Sync()
		var dialog *unison.Dialog
		var asked, opened []string
		var views []uxtest.Closable
		screen.Do(func() {
			for _, w := range unison.Windows() {
				if one, ok := w.ClientData()[unison.DialogClientDataKey].(*unison.Dialog); ok {
					dialog = one
					for _, label := range uxtest.PanelsOfType[*unison.Label](w.Content()) {
						if text := label.String(); text != "" {
							asked = append(asked, text)
						}
					}
				}
			}
			for _, one := range AllDockables() {
				if md, ok := one.AsPanel().Self.(*MarkdownDockable); ok {
					opened = append(opened, filepath.Base(md.BackingFilePath()))
					views = append(views, md)
				}
			}
		})
		if dialog != nil {
			screen.Do(func() { dialog.StopModal(unison.ModalResponseCancel) })
			screen.Sync()
		}
		c.Equal(tc.wantAsked, strings.Join(asked, " "), "%s: what a dialog asked", tc.name)
		c.Equal(tc.wantFile, strings.Join(opened, ", "), "%s: the files opened", tc.name)
		c.Equal(tc.wantURLs, screen.OpenedURLs(), "%s: the URLs handed to the system", tc.name)
		for _, view := range views {
			uxtest.CloseEditorWithoutPrompt(t, screen, view)
		}
	}
}
