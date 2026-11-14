// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package userguide

import (
	"embed"
	"io/fs"
	"net/url"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/richardwilkes/toolbox/v2/errs"
)

// HomePage is the page the User Guide opens at.
const HomePage = "Home.md"

// RefFolder is the folder that md: page references name the User Guide by, as in "md:User%20Guide/Traits".
const RefFolder = "User Guide"

const (
	mdExt       = ".md"
	sidebarFile = "_sidebar.md"
)

// The explicit globs take in _sidebar.md, which a directory pattern would skip, and keep stray dot files out.
//
//go:embed content/*.md content/Scripting/*.md content/images
var embedded embed.FS

var (
	guideFS = sync.OnceValue(func() fs.FS {
		sub, err := fs.Sub(embedded, "content")
		if err != nil {
			panic(err) // fs.Sub only fails for an invalid directory name
		}
		return sub
	})
	guidePages = sync.OnceValue(func() *pageSet { return newPageSet(listPages(FS())) })
)

type pageSet struct {
	exists map[string]bool
	names  []string
}

func newPageSet(names []string) *pageSet {
	set := &pageSet{
		exists: make(map[string]bool, len(names)),
		names:  names,
	}
	for _, name := range names {
		set.exists[name] = true
	}
	return set
}

// FS returns the User Guide's files: its pages, as slash paths with their ".md" extension, such as "Home.md" and
// "Scripting/Dice.md", the images they show and the table of contents, _sidebar.md.
func FS() fs.FS {
	return guideFS()
}

// Content returns the Markdown of a page.
func Content(page string) ([]byte, error) {
	if !guidePages().exists[page] {
		return nil, errs.Wrap(&fs.PathError{Op: "open", Path: page, Err: fs.ErrNotExist})
	}
	data, err := fs.ReadFile(FS(), page)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	return data, nil
}

// Title returns the text of a page's first level-1 heading or, when it has none, its file name less the ".md"
// extension. The first call builds the search index (see Search).
func Title(page string) string {
	if p, ok := builtIndex().byPage[page]; ok {
		return p.title
	}
	return fileTitle(page)
}

func fileTitle(page string) string {
	return strings.TrimSuffix(path.Base(page), mdExt)
}

// ResolveLink returns the page and anchor that a link target on the page from leads to. A target of only "#anchor"
// stays on from. Any other target must be a URL-escaped path relative to from's folder, with or without the ".md"
// extension, that names a page with its exact case. ok is false for everything else: URLs, md: links, page references
// such as "B10" and "PY83:21", and paths that lead out of the guide or to no page. The anchor is returned unescaped.
func ResolveLink(from, target string) (page, anchor string, ok bool) {
	if target == "" {
		return "", "", false
	}
	target, rawAnchor, _ := strings.Cut(target, "#")
	var err error
	if anchor, err = url.PathUnescape(rawAnchor); err != nil {
		return "", "", false
	}
	if target == "" {
		return from, anchor, true
	}
	if first, _, _ := strings.Cut(target, "/"); first == "" || strings.Contains(first, ":") {
		return "", "", false
	}
	var p string
	if p, err = url.PathUnescape(target); err != nil {
		return "", "", false
	}
	p = path.Join(path.Dir(from), p)
	if !fs.ValidPath(p) {
		return "", "", false
	}
	if !strings.HasSuffix(p, mdExt) {
		p += mdExt
	}
	if page, ok = lookup(p, false); !ok {
		return "", "", false
	}
	return page, anchor, true
}

// PageForRef returns the page an md: page reference names, given the reference's path unescaped and with slashes,
// such as "User Guide/Traits.md". inGuide is true when the path's first segment is RefFolder, ignoring case. page is the
// page the rest of the path names, also ignoring case, or "" when the guide has no such page.
func PageForRef(refPath string) (page string, inGuide bool) {
	folder, rest, _ := strings.Cut(refPath, "/")
	if !strings.EqualFold(folder, RefFolder) {
		return "", false
	}
	page, _ = lookup(rest, true)
	return page, true
}

// pages returns the guide's pages in name order.
func pages() []string {
	return guidePages().names
}

// lookup returns the page named name, or name with ".md" appended, matching case exactly or, when foldCase is true,
// ignoring it.
func lookup(name string, foldCase bool) (page string, ok bool) {
	set := guidePages()
	withExt := name + mdExt
	for _, candidate := range []string{name, withExt} {
		if set.exists[candidate] {
			return candidate, true
		}
	}
	if foldCase {
		for _, candidate := range set.names {
			if strings.EqualFold(candidate, name) || strings.EqualFold(candidate, withExt) {
				return candidate, true
			}
		}
	}
	return "", false
}

// listPages returns the Markdown files in fsys in name order, less those whose names begin with '_', which the wiki the
// guide is also published as reserves for its sidebar, header and footer.
func listPages(fsys fs.FS) []string {
	var names []string
	if err := fs.WalkDir(fsys, ".", func(name string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() && path.Ext(name) == mdExt && !strings.HasPrefix(d.Name(), "_") {
			names = append(names, name)
		}
		return nil
	}); err != nil {
		errs.Log(err)
	}
	slices.Sort(names)
	return names
}
