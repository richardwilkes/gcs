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
	"bytes"
	"io/fs"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/unison"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// snippetRunes is about how long the text of a BodyMatch is.
const snippetRunes = 80

// MatchKind identifies the part of a page a Match was found in. The kinds are declared in the order Search ranks them.
type MatchKind uint8

// The kinds of Match.
const (
	// TitleMatch is a match in a page's title.
	TitleMatch MatchKind = iota
	// HeadingMatch is a match in the text of a heading.
	HeadingMatch
	// BodyMatch is a match in the text of a section, which runs from one heading to the next.
	BodyMatch
)

// Match is a place in the User Guide that Search found.
type Match struct {
	// Page is the page the match is on.
	Page string
	// Anchor is the ID of the heading the match is under, or "" for a match in the title or above the first heading.
	Anchor string
	// Text is what matched: the page's title, the heading's text or, for a BodyMatch, about 80 characters of the
	// section's text around the match, with "…" where it was cut.
	Text string
	// Section is the text of the heading the match is under, or "" for a match in the title or above the first
	// heading.
	Section string
	// Kind is the part of the page the match was found in.
	Kind MatchKind
}

var builtIndex = sync.OnceValue(func() *searchIndex {
	var order []string // without a table of contents, pages are taken in name order
	if entries, err := TOC(); err == nil {
		order = tocPages(entries)
	}
	return buildIndex(FS(), order)
})

type searchIndex struct {
	byPage map[string]*pageIndex
	pages  []*pageIndex // in the order matches are listed
}

type pageIndex struct {
	name       string
	title      string
	lowerTitle string
	headings   []headingIndex
	sections   []sectionIndex
}

type headingIndex struct {
	text  string
	lower string
	id    string
}

type sectionIndex struct {
	anchor  string
	heading string
	text    string
	lower   string
}

type place struct {
	page   string
	anchor string
}

// Search returns the places in the User Guide whose text contains lowerText, which must already be lowercase: first the
// pages whose titles contain it, then the headings, then the sections. Within each of those, pages are taken in the
// order of the table of contents, followed by any it does not list in name order, and places on a page in document
// order. A place is listed once, as its first match; a page whose title matches is not listed again for the text above
// its first heading. Empty text matches nothing. At most limit matches are returned, when limit is greater than zero.
// The first call builds the index searched, parsing every page.
func Search(lowerText string, limit int) []Match {
	return builtIndex().search(lowerText, limit)
}

func (idx *searchIndex) search(lowerText string, limit int) []Match {
	if lowerText == "" {
		return nil
	}
	var matches []Match
	seen := make(map[place]bool)
	// add records m unless its place is already listed, and returns false once the limit is reached.
	add := func(m Match) bool {
		key := place{page: m.Page, anchor: m.Anchor}
		if !seen[key] {
			seen[key] = true
			matches = append(matches, m)
		}
		return limit < 1 || len(matches) < limit
	}
	for _, p := range idx.pages {
		if strings.Contains(p.lowerTitle, lowerText) && !add(Match{Page: p.name, Text: p.title, Kind: TitleMatch}) {
			return matches
		}
	}
	for _, p := range idx.pages {
		for _, h := range p.headings {
			if strings.Contains(h.lower, lowerText) && !add(Match{
				Page:    p.name,
				Anchor:  h.id,
				Text:    h.text,
				Section: h.text,
				Kind:    HeadingMatch,
			}) {
				return matches
			}
		}
	}
	for _, p := range idx.pages {
		for _, s := range p.sections {
			if i := strings.Index(s.lower, lowerText); i >= 0 && !add(Match{
				Page:    p.name,
				Anchor:  s.anchor,
				Text:    snippet(s.text, s.lower, i, len(lowerText)),
				Section: s.heading,
				Kind:    BodyMatch,
			}) {
				return matches
			}
		}
	}
	return matches
}

// snippet returns about snippetRunes runes of str around the match of length matchLen at byte offset at of lower, which
// is str lowercased, cut at spaces where it can be, with "…" in place of what was cut.
func snippet(str, lower string, at, matchLen int) string {
	runes := []rune(str)
	if len(runes) <= snippetRunes {
		return str
	}
	// strings.ToLower maps rune for rune, so rune offsets into lower are rune offsets into str.
	hitStart := utf8.RuneCountInString(lower[:at])
	hitEnd := hitStart + utf8.RuneCountInString(lower[at:at+matchLen])
	width := max(snippetRunes, hitEnd-hitStart)
	start := max(hitStart-(width-(hitEnd-hitStart))/2, 0)
	end := min(start+width, len(runes))
	start = max(end-width, 0)
	if start > 0 {
		if i := slices.Index(runes[start:hitStart], ' '); i >= 0 {
			start += i + 1
		}
	}
	if end < len(runes) {
		if i := lastSpace(runes[hitEnd:end]); i >= 0 {
			end = hitEnd + i
		}
	}
	result := string(runes[start:end])
	if start > 0 {
		result = "…" + result
	}
	if end < len(runes) {
		result += "…"
	}
	return result
}

func lastSpace(runes []rune) int {
	for i, r := range slices.Backward(runes) {
		if r == ' ' {
			return i
		}
	}
	return -1
}

// buildIndex indexes the pages in fsys for Search, ordering them as order lists them, then the rest in name order.
func buildIndex(fsys fs.FS, order []string) *searchIndex {
	names := listPages(fsys)
	idx := &searchIndex{byPage: make(map[string]*pageIndex, len(names))}
	md := unison.NewMarkdownParser()
	for _, name := range names {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			errs.Log(err, "page", name)
			continue
		}
		idx.byPage[name] = indexPage(md, name, data)
	}
	idx.pages = make([]*pageIndex, 0, len(idx.byPage))
	placed := make(map[string]bool, len(idx.byPage))
	for _, name := range slices.Concat(order, names) {
		if p, ok := idx.byPage[name]; ok && !placed[name] {
			placed[name] = true
			idx.pages = append(idx.pages, p)
		}
	}
	return idx
}

type pageBuilder struct {
	page    *pageIndex
	src     []byte
	anchor  string
	heading string
	body    strings.Builder
	titled  bool
}

func indexPage(md parser.Parser, name string, src []byte) *pageIndex {
	b := pageBuilder{
		page: &pageIndex{name: name},
		src:  src,
	}
	b.blocks(md.Parse(text.NewReader(src)))
	b.finishSection()
	if b.page.title == "" {
		b.page.title = fileTitle(name)
	}
	b.page.lowerTitle = strings.ToLower(b.page.title)
	return b.page
}

// blocks adds the text of parent's block children to the page, starting a section at each heading but the first
// level-1 heading, which is the page's title.
func (b *pageBuilder) blocks(parent ast.Node) {
	for n := parent.FirstChild(); n != nil; n = n.NextSibling() {
		switch node := n.(type) {
		case *ast.Heading:
			str := inlineText(node, b.src)
			if node.Level == 1 && !b.titled {
				b.titled = true
				b.page.title = str
				continue
			}
			b.finishSection()
			b.anchor = headingID(node)
			b.heading = str
			b.page.headings = append(b.page.headings, headingIndex{
				text:  str,
				lower: strings.ToLower(str),
				id:    b.anchor,
			})
		case *ast.FencedCodeBlock, *ast.CodeBlock:
			lines := n.Lines()
			for i := range lines.Len() {
				segment := lines.At(i)
				b.body.Write(segment.Value(b.src))
			}
			b.body.WriteByte(' ')
		case *ast.HTMLBlock:
			// Has no text: it is shown only when it holds nothing but <img> tags.
		default:
			if child := n.FirstChild(); child == nil || child.Type() != ast.TypeInline {
				b.blocks(n)
				continue
			}
			// A paragraph, a tight list item's text or a table cell.
			str := inlineText(n, b.src)
			if _, isQuote := parent.(*ast.Blockquote); isQuote && n == parent.FirstChild() {
				str = trimAlertMarker(str)
			}
			b.body.WriteString(str)
			b.body.WriteByte(' ')
		}
	}
}

func (b *pageBuilder) finishSection() {
	if str := collapseSpace(b.body.String()); str != "" {
		b.page.sections = append(b.page.sections, sectionIndex{
			anchor:  b.anchor,
			heading: b.heading,
			text:    str,
			lower:   strings.ToLower(str),
		})
	}
	b.body.Reset()
}

// headingID returns the ID a heading's anchor is known by, which unison.NewMarkdownParser gives every heading.
func headingID(heading *ast.Heading) string {
	if id, ok := heading.AttributeString("id"); ok {
		if b, isBytes := id.([]byte); isBytes {
			return string(b)
		}
	}
	return ""
}

// inlineText returns the text that n's inline children show, with runs of whitespace collapsed to a single space and a
// space where a <br> or <hr> tag breaks the line. Link text is included; image alt text is not.
func inlineText(n ast.Node, src []byte) string {
	var buf strings.Builder
	appendInlineText(&buf, n, src)
	return collapseSpace(buf.String())
}

func appendInlineText(buf *strings.Builder, n ast.Node, src []byte) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch node := c.(type) {
		case *ast.Text:
			if node.IsRaw() {
				buf.Write(node.Value(src))
			} else {
				buf.Write(markdownUnescape(node.Value(src)))
			}
			if node.SoftLineBreak() || node.HardLineBreak() {
				buf.WriteByte(' ')
			}
		case *ast.String:
			buf.Write(markdownUnescape(node.Value))
		case *ast.AutoLink:
			buf.Write(node.Label(src))
		case *ast.RawHTML:
			if breaksLine(node.Segments.Value(src)) {
				buf.WriteByte(' ')
			}
		case *ast.Image:
			// Not shown as text.
		default:
			appendInlineText(buf, c, src)
		}
	}
}

// breaksLine reports whether an inline HTML fragment is a <br> or <hr> tag, which unison.Markdown draws as a line
// break. Only an open tag's name can follow the '<' directly, so a closing tag, comment or the like never matches.
func breaksLine(fragment []byte) bool {
	name, ok := bytes.CutPrefix(fragment, []byte{'<'})
	if !ok {
		return false
	}
	if end := bytes.IndexFunc(name, func(r rune) bool {
		return r != '-' && (r < '0' || r > '9') && (r < 'a' || r > 'z') && (r < 'A' || r > 'Z')
	}); end >= 0 {
		name = name[:end]
	}
	return bytes.EqualFold(name, []byte("br")) || bytes.EqualFold(name, []byte("hr"))
}

// markdownUnescape resolves the backslash escapes and character references in Markdown text, as unison.Markdown does
// for text and link destinations. CommonMark resolves them in a single pass, so an escaped '&' begins no reference and
// the text a reference stands for is taken as written.
func markdownUnescape(b []byte) []byte {
	i := bytes.IndexAny(b, `\&`)
	if i < 0 {
		return b
	}
	out := make([]byte, i, len(b))
	copy(out, b)
	for i < len(b) {
		c := b[i]
		if c == '\\' && i+1 < len(b) && util.IsPunct(b[i+1]) {
			out = append(out, b[i+1])
			i += 2
			continue
		}
		if c == '&' {
			if resolved, length := markdownReference(b[i:]); length != 0 {
				out = append(out, resolved...)
				i += length
				continue
			}
		}
		out = append(out, c)
		i++
	}
	return out
}

// markdownReference returns what the character reference that b starts with stands for and the reference's length, or
// a length of 0 when b starts with none.
func markdownReference(b []byte) (resolved []byte, length int) {
	if len(b) < 3 || b[0] != '&' {
		return nil, 0
	}
	if b[1] != '#' {
		end, ok := util.ReadWhile(b, [2]int{1, len(b)}, util.IsAlphaNumeric)
		if !ok || end == len(b) || b[end] != ';' {
			return nil, 0
		}
		entity, found := util.LookUpHTML5EntityByName(string(b[1:end]))
		if !found {
			return nil, 0
		}
		return entity.Characters, end + 1
	}
	start, base, maxDigits, isDigit := 2, 10, 7, util.IsNumeric
	if b[2] == 'x' || b[2] == 'X' {
		start, base, maxDigits, isDigit = 3, 16, 6, util.IsHexDecimal
	}
	end, ok := util.ReadWhile(b, [2]int{start, len(b)}, isDigit)
	if !ok || end-start > maxDigits || end == len(b) || b[end] != ';' {
		return nil, 0
	}
	v, err := strconv.ParseUint(string(b[start:end]), base, 32)
	if err != nil {
		return nil, 0
	}
	return utf8.AppendRune(nil, util.ToValidRune(rune(v))), end + 1
}

// trimAlertMarker removes the "[!NOTE]" or similar that makes a quote an alert, which unison.Markdown shows as an icon.
func trimAlertMarker(str string) string {
	for _, marker := range []string{"[!NOTE]", "[!TIP]", "[!IMPORTANT]", "[!WARNING]", "[!CAUTION]"} {
		if rest, found := strings.CutPrefix(str, marker); found {
			return strings.TrimSpace(rest)
		}
	}
	return str
}

func collapseSpace(str string) string {
	return strings.Join(strings.Fields(str), " ")
}
