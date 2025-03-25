// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"path/filepath"
	"strings"

	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
)

// WorkingDirKey the client data key for setting the working directory for markdown and link resolution.
const WorkingDirKey = "working_dir"

// WorkingDirProvider extracts a working dir for the given panel, if possible, otherwise returns ".".
func WorkingDirProvider(p unison.Paneler) string {
	if toolbox.IsNil(p) {
		return "."
	}
	if d := unison.AncestorOrSelf[FileBackedDockable](p); !toolbox.IsNil(d) {
		if filePath := d.BackingFilePath(); filePath != "" && !strings.HasPrefix(filePath, markdownContentOnlyPrefix) {
			return filepath.Dir(filePath)
		}
	}
	panel := p.AsPanel()
	for panel != nil {
		if data, ok := panel.ClientData()[WorkingDirKey]; ok {
			if s, ok2 := data.(string); ok2 {
				return s
			}
		}
		panel = panel.Parent()
	}
	return "."
}

// HandleLink will try to open http, https, and md links, as well as resolve page references.
func HandleLink(p unison.Paneler, target string) {
	revised, err := unison.ReviseTarget(WorkingDirProvider(p), target, unison.DefaultMarkdownTheme.AltLinkPrefixes)
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to open ")+target, err)
		return
	}
	if !unison.HasURLPrefix(revised) && !unison.HasAnyPrefix(unison.DefaultMarkdownTheme.AltLinkPrefixes, revised) {
		if fs.FileIsReadable(revised) {
			OpenFile(revised, 0)
			return
		}
		revised = target
	}
	OpenPageReference(revised, "", nil)
}
