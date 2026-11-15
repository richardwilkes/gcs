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
	"fmt"
	"html"
	"io/fs"
	"path"
	"regexp"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// attrPattern matches an attribute of an HTML tag as CommonMark defines it, capturing its name and any value.
const attrPattern = `\s+([a-zA-Z_:][\w.:-]*)(?:\s*=\s*([^\s"'=<>` + "`" + `]+|'[^']*'|"[^"]*"))?`

// imgTagRE matches an <img> tag, capturing its attributes, and attrRE each of those.
var (
	imgTagRE = regexp.MustCompile(`<(?i:img)((?:` + attrPattern + `)*)\s*/?>`)
	attrRE   = regexp.MustCompile(attrPattern)
)

// Every link, image and anchor in the guide must resolve with its exact case in the embedded files, which are
// case-sensitive on every platform, as the viewer resolves it. Links are found in each page's parsed document, so
// examples in code spans and fenced code are never taken for links.
func TestEveryGuideLinkResolves(t *testing.T) {
	fsys := FS()
	md := unison.NewMarkdownParser()
	type parsedPage struct {
		doc ast.Node
		ids map[string]bool
		src []byte
	}
	parsed := make(map[string]*parsedPage)
	for _, page := range pages() {
		src, err := fs.ReadFile(fsys, page)
		if err != nil {
			t.Fatal(err)
		}
		pp := &parsedPage{doc: md.Parse(text.NewReader(src)), ids: make(map[string]bool), src: src}
		walk(t, pp.doc, func(n ast.Node) {
			if heading, ok := n.(*ast.Heading); ok {
				pp.ids[headingID(heading)] = true
			}
		})
		parsed[page] = pp
	}
	var failures []string
	fail := func(page, target, format string, args ...any) {
		failures = append(failures, fmt.Sprintf("%s: %q: ", page, target)+fmt.Sprintf(format, args...))
	}
	isExternal := func(target string) bool {
		lower := strings.ToLower(target)
		return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") ||
			strings.HasPrefix(lower, "mailto:")
	}
	checkAnchor := func(page, target, resolved, anchor string) {
		if anchor != "" && !parsed[resolved].ids[anchor] {
			fail(page, target, "%s has no heading with the ID %q", resolved, anchor)
		}
	}
	checkLink := func(page, target string) {
		if isExternal(target) {
			return
		}
		if ref, isRef := strings.CutPrefix(target, "md:"); isRef {
			rest, found := strings.CutPrefix(ref, "User%20Guide/")
			if !found {
				if rest, found = strings.CutPrefix(ref, RefFolder+"/"); !found {
					fail(page, target, "md: links must name a page of the User Guide")
					return
				}
			}
			resolved, anchor, ok := ResolveLink("", rest)
			if !ok {
				fail(page, target, "names no page")
				return
			}
			checkAnchor(page, target, resolved, anchor)
			return
		}
		resolved, anchor, ok := ResolveLink(page, target)
		if !ok {
			// A page reference, such as B10 or PY83:21, is opened in a PDF.
			if last, _ := utf8.DecodeLastRuneInString(target); !strings.Contains(target, ":") && !unicode.IsDigit(last) {
				fail(page, target, "names no page")
			}
			return
		}
		checkAnchor(page, target, resolved, anchor)
	}
	checkImage := func(page, target string) {
		if isExternal(target) {
			return
		}
		dir := path.Dir(page)
		if dir == "." {
			dir = ""
		}
		revised, err := unison.ReviseFSTarget(dir, target, nil)
		if err != nil {
			fail(page, target, "%v", err)
			return
		}
		if info, statErr := fs.Stat(fsys, revised); statErr != nil || info.IsDir() {
			fail(page, target, "image %q is not in the guide", revised)
		}
	}
	for _, page := range pages() {
		pp := parsed[page]
		walk(t, pp.doc, func(n ast.Node) {
			switch node := n.(type) {
			case *ast.Link:
				checkLink(page, string(markdownUnescape(node.Destination)))
			case *ast.AutoLink:
				if node.AutoLinkType == ast.AutoLinkURL {
					checkLink(page, string(node.URL(pp.src)))
				}
			case *ast.Image:
				checkImage(page, string(markdownUnescape(node.Destination)))
			default:
				for _, attrs := range imgTags(n, pp.src) {
					checkImage(page, attrs["src"])
				}
			}
		})
	}
	if len(failures) != 0 {
		t.Errorf("%d targets do not resolve with their exact case:\n%s", len(failures), strings.Join(failures, "\n"))
	}
}

// Raster images are <img> tags with a width in logical pixels, so that a 2x screenshot is shown at the size it had on
// screen.
func TestRasterImagesAreImgTagsWithWidth(t *testing.T) {
	fsys := FS()
	md := unison.NewMarkdownParser()
	positiveInt := regexp.MustCompile(`^[1-9]\d*$`)
	isSVG := func(target string) bool { return strings.EqualFold(path.Ext(target), ".svg") }
	var failures []string
	for _, page := range pages() {
		src, err := fs.ReadFile(fsys, page)
		if err != nil {
			t.Fatal(err)
		}
		walk(t, md.Parse(text.NewReader(src)), func(n ast.Node) {
			if img, ok := n.(*ast.Image); ok {
				if target := string(markdownUnescape(img.Destination)); !isSVG(target) {
					failures = append(failures, fmt.Sprintf(`%s: %q: write it as <img src=%q alt="…" width="…">`, page,
						target, target))
				}
				return
			}
			for _, attrs := range imgTags(n, src) {
				if isSVG(attrs["src"]) {
					continue
				}
				if width, ok := attrs["width"]; !ok {
					failures = append(failures, fmt.Sprintf("%s: %q: has no width", page, attrs["src"]))
				} else if !positiveInt.MatchString(width) {
					failures = append(failures, fmt.Sprintf("%s: %q: width %q is not a positive whole number", page,
						attrs["src"], width))
				}
			}
		})
	}
	if len(failures) != 0 {
		t.Errorf("%d raster images must be <img> tags with a width in logical pixels, half the pixel width of a 2x "+
			"capture:\n%s", len(failures), strings.Join(failures, "\n"))
	}
}

// Links are resolved relative to the folder of the page they are on, with their exact case, and page references and
// links that use a scheme are left for the caller.
func TestResolveLink(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		from, target string
		page, anchor string
		ok           bool
	}{
		{from: "Traits.md", target: "Skills", page: "Skills.md", ok: true},
		{from: "Home.md", target: "Character%20Sheet%20Overview", page: "Character Sheet Overview.md", ok: true},
		{from: "Home.md", target: "Traits.md", page: "Traits.md", ok: true},
		{from: "Home.md", target: "./Traits", page: "Traits.md", ok: true},
		{from: "Scripting/Dice.md", target: "Entity", page: "Scripting/Entity.md", ok: true},
		{from: "Scripting/Dice.md", target: "../Scripting%20Guide", page: "Scripting Guide.md", ok: true},
		{from: "Home.md", target: "Scripting/Entity", page: "Scripting/Entity.md", ok: true},
		{from: "Home.md", target: "Traits#editing-traits", page: "Traits.md", anchor: "editing-traits", ok: true},
		{from: "Traits.md", target: "#editing-traits", page: "Traits.md", anchor: "editing-traits", ok: true},
		{from: "Home.md", target: ""},
		{from: "Home.md", target: "https://gurpscharactersheet.com"},
		{from: "Home.md", target: "md:User%20Guide/Traits"},
		{from: "Home.md", target: "B10"},
		{from: "Home.md", target: "PY83:21"},
		{from: "Home.md", target: "traits"},
		{from: "Home.md", target: "/Traits"},
		{from: "Scripting/Dice.md", target: "../../Traits"},
		{from: "Home.md", target: "_sidebar"},
	} {
		page, anchor, ok := ResolveLink(tc.from, tc.target)
		c.Equal(tc.ok, ok, "ok for %q from %q", tc.target, tc.from)
		c.Equal(tc.page, page, "page for %q from %q", tc.target, tc.from)
		c.Equal(tc.anchor, anchor, "anchor for %q from %q", tc.target, tc.from)
	}
}

// md: references to the guide match its folder and pages ignoring case, as hand-typed references may not match it,
// while references to other folders are left to the libraries.
func TestPageForRef(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		refPath string
		page    string
		inGuide bool
	}{
		{refPath: "User Guide/Traits.md", page: "Traits.md", inGuide: true},
		{refPath: "User Guide/Traits", page: "Traits.md", inGuide: true},
		{refPath: "user guide/TRAITS.MD", page: "Traits.md", inGuide: true},
		{refPath: "User Guide/Body Type.md", page: "Body type.md", inGuide: true},
		{refPath: "User Guide/scripting/dice.md", page: "Scripting/Dice.md", inGuide: true},
		{refPath: "User Guide/No Such Page.md", inGuide: true},
		{refPath: "User Guide/_sidebar.md", inGuide: true},
		{refPath: "Other/Traits.md"},
	} {
		page, inGuide := PageForRef(tc.refPath)
		c.Equal(tc.inGuide, inGuide, "inGuide for %q", tc.refPath)
		c.Equal(tc.page, page, "page for %q", tc.refPath)
	}
}

// walk calls visit for each node of doc, parents first.
func walk(t *testing.T, doc ast.Node, visit func(n ast.Node)) {
	t.Helper()
	if err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			visit(n)
		}
		return ast.WalkContinue, nil
	}); err != nil {
		t.Fatal(err)
	}
}

// imgTags returns the attributes of each <img> tag in n when it is inline or block HTML, keyed by their lowercase
// names, with character references resolved.
func imgTags(n ast.Node, src []byte) []map[string]string {
	var lines *text.Segments
	switch node := n.(type) {
	case *ast.RawHTML:
		lines = node.Segments
	case *ast.HTMLBlock:
		lines = node.Lines()
	default:
		return nil
	}
	var markup strings.Builder
	for i := range lines.Len() {
		segment := lines.At(i)
		markup.Write(segment.Value(src))
	}
	matches := imgTagRE.FindAllStringSubmatch(markup.String(), -1)
	tags := make([]map[string]string, 0, len(matches))
	for _, tag := range matches {
		attrs := make(map[string]string)
		for _, attr := range attrRE.FindAllStringSubmatch(tag[1], -1) {
			value := attr[2]
			if value != "" && (value[0] == '"' || value[0] == '\'') {
				value = value[1 : len(value)-1]
			}
			attrs[strings.ToLower(attr[1])] = html.UnescapeString(value)
		}
		tags = append(tags, attrs)
	}
	return tags
}
