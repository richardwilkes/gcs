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
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// The table of contents parses and lists every page, so none is reachable only by a link.
func TestEveryPageIsInTheTOC(t *testing.T) {
	c := check.New(t)
	entries, err := TOC()
	c.NoError(err)
	listed := make(map[string]bool)
	for _, page := range tocPages(entries) {
		listed[page] = true
	}
	for _, page := range pages() {
		c.True(listed[page], "%s is not listed in %s", page, sidebarFile)
	}
}

// Only the line shapes the parser models are accepted, so that the tree it builds is the one the wiki shows from the
// same file.
func TestParseTOCRejectsUnknownLines(t *testing.T) {
	c := check.New(t)
	resolve := func(target string) (string, bool) {
		switch target {
		case "A", "B", "Sub/C":
			return target + mdExt, true
		default:
			return "", false
		}
	}
	for _, tc := range []struct {
		name string
		data string
		want string // the tree, when accepted
		line int    // the line the error names, when rejected
	}{
		{
			name: "accepted",
			data: `## Browse

[[A]]

<details open>
<summary>Group</summary>

- [Bee](B)
- **Sub**
  - [See](Sub/C)
- [Ay](A)

</details>
`,
			want: `A A.md
Group
  Bee B.md
  Sub
    See Sub/C.md
  Ay A.md
`,
		},
		{
			name: "titles as the wiki shows them",
			data: "<details>\n<summary>Q &amp;  A</summary>\n\n- [Q\\&A](B)\n- [\\&amp; &#38;amp;](A)\n- **S&#117;b**\n  - [See](Sub/C)\n</details>\n",
			want: "Q & A\n  Q&A B.md\n  &amp; &amp; A.md\n  Sub\n    See Sub/C.md\n",
		},
		{
			name: "inline Markdown in labels",
			data: "<details>\n<summary>G</summary>\n\n- [*Q*](B)\n- **a _b_ c**\n  - [`Q`](Sub/C)\n</details>\n",
			want: "G\n  Q B.md\n  a b c\n    Q Sub/C.md\n",
		},
		{name: "escaped bracket ending a label", data: "- [a\\](B)\n", line: 1},
		{
			name: "spaces inside the bold of a sub-group",
			data: "<details>\n<summary>G</summary>\n\n- ** S **\n  - [B](B)\n</details>\n",
			line: 4,
		},
		{name: "paragraph", data: "[[A]]\n\nSome text.\n", line: 3},
		{name: "indent-4 item", data: "<details>\n<summary>G</summary>\n\n- **S**\n    - [B](B)\n</details>\n", line: 5},
		{
			name: "nested details",
			data: "<details>\n<summary>G</summary>\n\n- [B](B)\n<details>\n<summary>H</summary>\n\n- [A](A)\n</details>\n</details>\n",
			line: 5,
		},
		{name: "missing summary", data: "<details>\n\n- [B](B)\n</details>\n", line: 3},
		{name: "indent-2 item outside a sub-group", data: "<details>\n<summary>G</summary>\n\n  - [B](B)\n</details>\n", line: 4},
		{name: "unresolvable target", data: "<details>\n<summary>G</summary>\n\n- [Z](Zed)\n</details>\n", line: 4},
		{name: "empty group", data: "[[A]]\n<details>\n<summary>G</summary>\n</details>\n", line: 2},
		{name: "empty sub-group", data: "<details>\n<summary>G</summary>\n\n- **S**\n- [B](B)\n</details>\n", line: 4},
		{name: "unclosed group", data: "<details>\n<summary>G</summary>\n\n- [B](B)\n", line: 1},
		{name: "heading after an entry", data: "[[A]]\n## Browse\n", line: 2},
		{name: "entry run into the HTML above", data: "<details>\n<summary>G</summary>\n- [B](B)\n\n</details>\n", line: 3},
		{name: "page run into the entry above", data: "- [B](B)\n[[A]]\n", line: 2},
	} {
		entries, err := parseTOC([]byte(tc.data), resolve)
		if tc.line == 0 {
			c.NoError(err, tc.name)
			c.Equal(tc.want, renderTOC(entries, ""), tc.name)
			continue
		}
		c.HasError(err, tc.name)
		if err != nil {
			msg, _, _ := strings.Cut(err.Error(), "\n") // the rest is the stack trace
			c.HasPrefix(msg, fmt.Sprintf("%s:%d: ", sidebarFile, tc.line), tc.name)
		}
	}
}

func renderTOC(entries []*TOCEntry, indent string) string {
	var buf strings.Builder
	for _, entry := range entries {
		buf.WriteString(indent + entry.Title)
		if entry.Page != "" {
			buf.WriteString(" " + entry.Page)
		}
		buf.WriteByte('\n')
		buf.WriteString(renderTOC(entry.Children, indent+"  "))
	}
	return buf.String()
}
