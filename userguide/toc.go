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
	"html"
	"io/fs"
	"regexp"
	"strings"
	"sync"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/unison"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// TOCEntry is an entry in the User Guide's table of contents: a page, or a group of entries.
type TOCEntry struct {
	// Title is the text the entry is listed as.
	Title string
	// Page is the page the entry opens, or "" for a group.
	Page string
	// Children holds a group's entries, in order.
	Children []*TOCEntry
}

var (
	guideTOC = sync.OnceValues(func() ([]*TOCEntry, error) {
		data, err := fs.ReadFile(FS(), sidebarFile)
		if err != nil {
			return nil, errs.Wrap(err)
		}
		return parseTOC(data, func(target string) (string, bool) {
			page, anchor, ok := ResolveLink("", target)
			return page, ok && page != "" && anchor == ""
		})
	})
	tocWikiLink = regexp.MustCompile(`^\[\[([^\[\]|]+)\]\]$`)
	tocSummary  = regexp.MustCompile(`^<summary>([^<>]+)</summary>$`)
	tocLink     = regexp.MustCompile(`^( {2})?- \[([^\[\]]+)\]\(([^()\s]+)\)$`)
	tocSubGroup = regexp.MustCompile(`^- \*\*([^*]+)\*\*$`)
)

// TOC returns the User Guide's table of contents, read from _sidebar.md on first use. The entries are shared, so they
// must not be modified.
func TOC() ([]*TOCEntry, error) {
	return guideTOC()
}

// tocPages returns the pages the entries open, in reading order.
func tocPages(entries []*TOCEntry) []string {
	list := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Page != "" {
			list = append(list, entry.Page)
		}
		list = append(list, tocPages(entry.Children)...)
	}
	return list
}

// tocRunOn is what the wiki carries on into the next line of _sidebar.md when no blank line comes between them.
type tocRunOn uint8

const (
	runOnNothing   tocRunOn = iota
	runOnHTML               // HTML, which takes in every line up to a blank one
	runOnParagraph          // the paragraph of a "[[Page]]" or list item, which takes in a "[[Page]]" line
)

type tocParser struct {
	resolve      func(target string) (page string, ok bool)
	md           parser.Parser
	group        *TOCEntry // the open <details> group
	sub          *TOCEntry // the open sub-group of group
	entries      []*TOCEntry
	groupLine    int
	subLine      int
	runOn        tocRunOn
	needSummary  bool
	headingFound bool
	entryFound   bool
}

// parseTOC parses the table of contents in _sidebar.md, which also serves the GitHub wiki the guide is published as.
// Only these line shapes are accepted, so that the tree built from them is the one the wiki shows:
//
//   - blank lines, one of which must separate an entry or sub-group from a <details>, <summary> or </details> line
//     above it, which the wiki would otherwise show it as part of, and a "[[Page]]" from an entry or sub-group above
//     it, which the wiki would otherwise join it to;
//   - a single "## Heading" ahead of everything else, for the wiki, which is ignored;
//   - "[[Page]]", an entry titled with the page's name;
//   - "<details>" or "<details open>", followed by "<summary>Label</summary>", which opens a group, and "</details>",
//     which closes it. Groups do not nest, and neither they nor sub-groups may be empty. "open" only matters to the
//     wiki;
//   - "- [Label](Target)", an entry, which also closes any open sub-group;
//   - "- **Label**", which opens a sub-group of the open group, and "  - [Label](Target)", an entry in it.
//
// A label is taken as the wiki shows it: with runs of white space collapsed, character references resolved and, but
// for the HTML of <summary>, backslash escapes resolved and other inline Markdown reduced to its text. Targets are
// resolved to pages by resolve. Anything else, including an entry or sub-group the wiki does not show as a link or bold
// text, is an error that names its line.
func parseTOC(data []byte, resolve func(target string) (page string, ok bool)) ([]*TOCEntry, error) {
	p := tocParser{resolve: resolve, md: unison.NewMarkdownParser()}
	lineNum := 0
	for line := range strings.Lines(string(data)) {
		lineNum++
		line = strings.TrimRight(line, " \t\r\n")
		if line == "" {
			p.runOn = runOnNothing
			continue
		}
		if err := p.parseLine(lineNum, line); err != nil {
			return nil, err
		}
	}
	if p.group != nil {
		return nil, tocError(p.groupLine, "<details> is never closed", "")
	}
	return p.entries, nil
}

func (p *tocParser) parseLine(lineNum int, line string) error {
	if p.needSummary {
		m := tocSummary.FindStringSubmatch(line)
		if m == nil {
			return tocError(lineNum, "expected <summary>Label</summary> after <details>", line)
		}
		p.group.Title = html.UnescapeString(collapseSpace(m[1]))
		p.needSummary = false
		p.runOn = runOnHTML
		return nil
	}
	if strings.HasPrefix(line, "## ") {
		if p.headingFound || p.entryFound {
			return tocError(lineNum, "a heading is only allowed once, ahead of every entry", line)
		}
		p.headingFound = true
		return nil
	}
	p.entryFound = true
	switch line {
	case "<details>", "<details open>":
		if p.group != nil {
			return tocError(lineNum, "<details> may not be nested", line)
		}
		p.group = &TOCEntry{}
		p.groupLine = lineNum
		p.needSummary = true
		p.runOn = runOnHTML
		return nil
	case "</details>":
		if p.group == nil {
			return tocError(lineNum, "</details> without <details>", line)
		}
		if err := p.closeSub(); err != nil {
			return err
		}
		if len(p.group.Children) == 0 {
			return tocError(p.groupLine, "group has no entries", "")
		}
		p.entries = append(p.entries, p.group)
		p.group = nil
		p.runOn = runOnHTML
		return nil
	}
	if m := tocWikiLink.FindStringSubmatch(line); m != nil {
		if err := p.startMarkdown(lineNum, line, true); err != nil {
			return err
		}
		return p.addEntry(lineNum, line, m[1], m[1], false)
	}
	if m := tocLink.FindStringSubmatch(line); m != nil {
		if err := p.startMarkdown(lineNum, line, false); err != nil {
			return err
		}
		title, ok := p.shownLabel(line, true)
		if !ok {
			return tocError(lineNum, "the wiki does not show this as a link", line)
		}
		return p.addEntry(lineNum, line, title, m[3], m[1] != "")
	}
	if tocSubGroup.MatchString(line) {
		if p.group == nil {
			return tocError(lineNum, "a sub-group must be inside <details>", line)
		}
		if err := p.startMarkdown(lineNum, line, false); err != nil {
			return err
		}
		title, ok := p.shownLabel(line, false)
		if !ok {
			return tocError(lineNum, "the wiki does not show this as bold text", line)
		}
		if err := p.closeSub(); err != nil {
			return err
		}
		p.sub = &TOCEntry{Title: title}
		p.subLine = lineNum
		p.group.Children = append(p.group.Children, p.sub)
		return nil
	}
	return tocError(lineNum, "unrecognized line", line)
}

// startMarkdown checks that the wiki does not take a Markdown line into what it made of the line above, and records
// that the wiki carries the line's paragraph on into a "[[Page]]" line directly below it.
func (p *tocParser) startMarkdown(lineNum int, line string, isPage bool) error {
	switch {
	case p.runOn == runOnHTML:
		return tocError(lineNum, "without a blank line above, the wiki shows this as part of the HTML above it", line)
	case p.runOn == runOnParagraph && isPage:
		return tocError(lineNum, "without a blank line above, the wiki joins this to the entry above it", line)
	}
	p.runOn = runOnParagraph
	return nil
}

// shownLabel returns the text the wiki shows for the label of an entry's or sub-group's list item, and false when it
// does not show the item as a link or, when link is false, as bold text.
func (p *tocParser) shownLabel(line string, link bool) (string, bool) {
	src := []byte(line)
	block := p.md.Parse(text.NewReader(src))
	for block.FirstChild() != nil && block.FirstChild().Type() != ast.TypeInline {
		block = block.FirstChild()
	}
	label := block.FirstChild()
	if label == nil || label.NextSibling() != nil {
		return "", false
	}
	switch node := label.(type) {
	case *ast.Link:
		if !link {
			return "", false
		}
	case *ast.Emphasis:
		if link || node.Level != 2 {
			return "", false
		}
	default:
		return "", false
	}
	return inlineText(label, src), true
}

func (p *tocParser) addEntry(lineNum int, line, title, target string, inSub bool) error {
	page, ok := p.resolve(target)
	if !ok {
		return tocError(lineNum, "target names no page of the guide", line)
	}
	entry := &TOCEntry{Title: strings.TrimSpace(title), Page: page}
	switch {
	case inSub:
		if p.sub == nil {
			return tocError(lineNum, "an indented entry must follow a sub-group", line)
		}
		p.sub.Children = append(p.sub.Children, entry)
	case p.group != nil:
		if err := p.closeSub(); err != nil {
			return err
		}
		p.group.Children = append(p.group.Children, entry)
	default:
		p.entries = append(p.entries, entry)
	}
	return nil
}

func (p *tocParser) closeSub() error {
	if p.sub != nil && len(p.sub.Children) == 0 {
		return tocError(p.subLine, "sub-group has no entries", "")
	}
	p.sub = nil
	return nil
}

func tocError(lineNum int, msg, line string) error {
	if line == "" {
		return errs.Newf("%s:%d: %s", sidebarFile, lineNum, msg)
	}
	return errs.Newf("%s:%d: %s: %q", sidebarFile, lineNum, msg, line)
}
