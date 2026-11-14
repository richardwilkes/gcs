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
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

// WorkingDirKey is the client data key for setting the working directory for markdown and link resolution.
const WorkingDirKey = "working_dir"

// WorkingDirProvider extracts a working dir for the given panel, if possible, otherwise returns ".".
func WorkingDirProvider(p unison.Paneler) string {
	if xreflect.IsNil(p) {
		return "."
	}
	if d := unison.AncestorOrSelf[FileBackedDockable](p); !xreflect.IsNil(d) {
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

// HandleLink follows a link. A target that is empty or only an "#anchor" opens nothing. Any other that is neither a URL
// nor an md: reference opens the file it names, if any (a markdown file at the "#anchor" that follows its name; see
// linkedFile), or else, when it can only name a file (see namesOnlyAFile), reports that there is none. The rest are
// handed, as written, to OpenPageReference.
func HandleLink(p unison.Paneler, target string) {
	if target == "" || strings.HasPrefix(target, "#") {
		// An empty target names the document holding the link, and the markdown hands over an anchor link only when it
		// has no heading the anchor names.
		return
	}
	workingDir := WorkingDirProvider(p)
	revised, err := unison.ReviseTarget(workingDir, target, unison.DefaultMarkdownTheme.AltLinkPrefixes)
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to open ")+target, err)
		return
	}
	if !unison.HasURLPrefix(revised) && !unison.HasAnyPrefix(unison.DefaultMarkdownTheme.AltLinkPrefixes, revised) {
		if filePath, anchor, ok := linkedFile(workingDir, target); ok {
			if gurps.FileInfoFor(filePath).UTI == uti.Markdown {
				openMarkdownFileAtAnchor(filePath, anchor)
			} else {
				OpenFile(filePath, gurps.PageInfo{})
			}
			return
		}
		if name, ok := namesOnlyAFile(target); ok {
			unison.ErrorDialogWithMessage(i18n.Text("Unable to open ")+name, i18n.Text("There is no file of that name."))
			return
		}
	}
	// Not revised, which ReviseTarget has unescaped, since OpenPageReference unescapes an md: reference itself.
	OpenPageReference(target, "", nil)
}

// namesOnlyAFile returns the link target unescaped, and true if OpenPageReference would take it for a page of a PDF
// though it lacks the key or the page of one (see splitPDFPageRef), or holds whitespace or a path separator, which no
// page reference does.
func namesOnlyAFile(target string) (name string, ok bool) {
	if unison.HasURLPrefix(target) || isMailtoLink(target) || strings.HasPrefix(target, "md:") {
		return target, false
	}
	name, err := url.PathUnescape(target)
	if err != nil {
		return target, false
	}
	if key, label, _ := splitPDFPageRef(target); key == "" || label == "" {
		return name, true
	}
	return name, strings.ContainsFunc(name, func(r rune) bool {
		return unicode.IsSpace(r) || (r < utf8.RuneSelf && os.IsPathSeparator(uint8(r)))
	})
}

// linkedFile returns the readable file a relative or absolute link target names, as given or, when it ends in a name
// rather than a folder, with ".md" appended, and the "#anchor" that followed it. ok is false for URLs, md: links and
// targets that name no readable file.
func linkedFile(workingDir, target string) (filePath, anchor string, ok bool) {
	// The whole target is tried first, so that a file whose name holds a '#' is still found.
	if filePath, ok = readableLinkTarget(workingDir, target); ok {
		return filePath, "", true
	}
	if before, after, found := strings.Cut(target, "#"); found && before != "" {
		if filePath, ok = readableLinkTarget(workingDir, before); ok {
			return filePath, after, true
		}
	}
	return "", "", false
}

func readableLinkTarget(workingDir, target string) (filePath string, ok bool) {
	altLinkPrefixes := unison.DefaultMarkdownTheme.AltLinkPrefixes
	revised, err := unison.ReviseTarget(workingDir, target, altLinkPrefixes)
	if err != nil || unison.HasURLPrefix(revised) || unison.HasAnyPrefix(altLinkPrefixes, revised) {
		return "", false
	}
	candidates := []string{revised}
	// A target ending in a folder, such as "", "." or "../", is not given the extension, which would make it name the
	// page beside that folder.
	if name, unescapeErr := url.PathUnescape(target); unescapeErr == nil && !endsInFolder(name) {
		candidates = append(candidates, revised+uti.Markdown.Extensions[0])
	}
	for _, one := range candidates {
		if xos.FileIsReadable(one) {
			return one, true
		}
	}
	return "", false
}

// endsInFolder returns true if the unescaped link target ends in a folder rather than a name: it is empty or ends in a
// separator, "." or "..".
func endsInFolder(target string) bool {
	if target == "" || os.IsPathSeparator(target[len(target)-1]) {
		return true
	}
	base := filepath.Base(target)
	return base == "." || base == ".."
}
