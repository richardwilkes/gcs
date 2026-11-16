// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package nameable_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestExtractParts pins how ExtractParts splits text into literal and placeholder parts, none of them empty. It yields
// no empty literal around a placeholder, an escaped delimiter delimits nothing and keeps its escape, and an empty,
// unclosed or control-rune-broken placeholder becomes literal text that keeps its open delimiter. An open inside a
// placeholder turns the earlier open and what follows it into literal text and restarts the placeholder.
func TestExtractParts(t *testing.T) {
	c := check.New(t)
	lit := func(s string) nameable.Part { return nameable.Part{Value: s} }
	ph := func(s string) nameable.Part { return nameable.Part{Value: s, Placeholder: true} }
	for _, tc := range []struct {
		src        string
		begin, end rune
		want       []nameable.Part
	}{
		{src: "abc @Foo@ def", begin: '@', end: '@', want: []nameable.Part{lit("abc "), ph("Foo"), lit(" def")}},
		{src: "abc def", begin: '@', end: '@', want: []nameable.Part{lit("abc def")}},
		{src: "", begin: '@', end: '@', want: nil},
		{src: "@Foo@", begin: '@', end: '@', want: []nameable.Part{ph("Foo")}},
		{src: "@Foo@ trailing", begin: '@', end: '@', want: []nameable.Part{ph("Foo"), lit(" trailing")}},
		// A close must end the placeholder, or the text between two placeholders would be taken for a third.
		{src: "@A@ and @B@", begin: '@', end: '@', want: []nameable.Part{ph("A"), lit(" and "), ph("B")}},
		{src: "@A@@B@", begin: '@', end: '@', want: []nameable.Part{ph("A"), ph("B")}},
		{src: `abc \@Foo\@ def`, begin: '@', end: '@', want: []nameable.Part{lit(`abc \@Foo\@ def`)}},
		// Paired escapes cancel out, so the '@' after them still opens a placeholder.
		{src: `abc \\@Foo@`, begin: '@', end: '@', want: []nameable.Part{lit(`abc \\`), ph("Foo")}},
		// An odd run escapes the first '@', so the second is an unclosed open that becomes its own literal part.
		{src: `abc \\\@Foo@`, begin: '@', end: '@', want: []nameable.Part{lit(`abc \\\@Foo`), lit("@")}},
		// An open immediately followed by a close is folded back into literal text, wherever it is, however many there
		// are and whatever follows it.
		{src: "a@@b", begin: '@', end: '@', want: []nameable.Part{lit("a"), lit("@@b")}},
		{src: "@@", begin: '@', end: '@', want: []nameable.Part{lit("@@")}},
		{src: "@@abc", begin: '@', end: '@', want: []nameable.Part{lit("@@abc")}},
		{src: "abc@@", begin: '@', end: '@', want: []nameable.Part{lit("abc"), lit("@@")}},
		{src: "a@@@@b", begin: '@', end: '@', want: []nameable.Part{lit("a"), lit("@@"), lit("@@b")}},
		{src: "@@@@", begin: '@', end: '@', want: []nameable.Part{lit("@@"), lit("@@")}},
		{src: "@@@", begin: '@', end: '@', want: []nameable.Part{lit("@@"), lit("@")}},
		{src: "@@@Foo@", begin: '@', end: '@', want: []nameable.Part{lit("@@"), ph("Foo")}},
		{src: "@Foo@@@", begin: '@', end: '@', want: []nameable.Part{ph("Foo"), lit("@@")}},
		// An unclosed open is flushed as literal text with its delimiter, even when it is the whole input.
		{src: "abc @Foo", begin: '@', end: '@', want: []nameable.Part{lit("abc "), lit("@Foo")}},
		{src: "abc@", begin: '@', end: '@', want: []nameable.Part{lit("abc"), lit("@")}},
		{src: "@", begin: '@', end: '@', want: []nameable.Part{lit("@")}},
		// A control rune aborts the placeholder, keeping its open delimiter; the next '@' then opens one that never
		// closes.
		{src: "@Foo\nBar@ baz", begin: '@', end: '@', want: []nameable.Part{lit("@Foo\nBar"), lit("@ baz")}},
		{src: "@\nFoo@", begin: '@', end: '@', want: []nameable.Part{lit("@\nFoo"), lit("@")}},
		// A control rune outside a placeholder is plain literal text.
		{src: "\n@Foo@", begin: '@', end: '@', want: []nameable.Part{lit("\n"), ph("Foo")}},
		{src: "a(b)c", begin: '(', end: ')', want: []nameable.Part{lit("a"), ph("b"), lit("c")}},
		// An open inside a placeholder restarts it; the earlier open and what follows it become literal text.
		{src: "a(b(c)d", begin: '(', end: ')', want: []nameable.Part{lit("a"), lit("(b"), ph("c"), lit("d")}},
		{src: "(((x)", begin: '(', end: ')', want: []nameable.Part{lit("("), lit("("), ph("x")}},
		{src: "((", begin: '(', end: ')', want: []nameable.Part{lit("("), lit("(")}},
		// An open right after a close is literal text when nothing closes it.
		{src: "(a)(b", begin: '(', end: ')', want: []nameable.Part{ph("a"), lit("(b")}},
		// An escaped open inside a placeholder does not restart it.
		{src: `(a\(b)`, begin: '(', end: ')', want: []nameable.Part{ph(`a\(b`)}},
	} {
		got := nameable.ExtractParts(tc.src, tc.begin, tc.end)
		c.Equal(tc.want, got, "ExtractParts(%q, %q, %q)", tc.src, tc.begin, tc.end)
	}
}

// TestExtractPartsPanicsOnEscapeDelimiter pins that ExtractParts panics when either delimiter is the escape rune.
func TestExtractPartsPanicsOnEscapeDelimiter(t *testing.T) {
	c := check.New(t)
	c.Panics(func() { nameable.ExtractParts("x", nameable.EscapeRune, '@') }, "begin is the escape rune")
	c.Panics(func() { nameable.ExtractParts("x", '@', nameable.EscapeRune) }, "end is the escape rune")
}

// TestExtractSegments pins where ExtractSegments splits: a leading or doubled delimiter yields an empty segment but a
// trailing one does not, and an escaped delimiter is no split point and keeps its escape.
func TestExtractSegments(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		src  string
		want []string
	}{
		{src: "A|B|C", want: []string{"A", "B", "C"}},
		{src: "ABC", want: []string{"ABC"}},
		{src: "", want: nil},
		{src: "|A|B", want: []string{"", "A", "B"}},
		{src: "A|B|", want: []string{"A", "B"}},
		{src: "A||B", want: []string{"A", "", "B"}},
		// ExtractSegments only decides where to split; unescaping is UnescapeRunes' job.
		{src: `A\|B|C`, want: []string{`A\|B`, "C"}},
		// Paired escapes cancel out, so the delimiter after them still splits.
		{src: `A\\|B`, want: []string{`A\\`, "B"}},
		{src: `A|B\`, want: []string{"A", `B\`}},
	} {
		c.Equal(tc.want, nameable.ExtractSegments(tc.src, '|'), "ExtractSegments(%q)", tc.src)
	}
}

func TestExtractSegmentsPanicsOnEscapeDelimiter(t *testing.T) {
	c := check.New(t)
	c.Panics(func() { nameable.ExtractSegments("x", nameable.EscapeRune) })
}

// EscapeRunes returns its input when given no runes to escape, and writes a newline as its reserved letter, keeping the
// escaped text on one line.
func TestEscapeRunes(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		in    string
		chars []rune
		want  string
	}{
		{in: `a|b`, chars: nil, want: `a|b`},
		{in: "line one\nline two", chars: []rune{'|', '\n'}, want: `line one\nline two`},
	} {
		c.Equal(tc.want, nameable.EscapeRunes(tc.in, tc.chars...), "EscapeRunes(%q, %q)", tc.in, tc.chars)
	}
}

// UnescapeRunes returns its input when given no runes to unescape, turns the reserved letter for a newline back into
// the newline, and writes an escape rune that ends the input back out rather than dropping it.
func TestUnescapeRunes(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		in    string
		chars []rune
		want  string
	}{
		{in: `a\|b`, chars: nil, want: `a\|b`},
		{in: `line one\nline two`, chars: []rune{'|', '\n'}, want: "line one\nline two"},
		{in: `abc\`, chars: []rune{'|'}, want: `abc\`},
	} {
		c.Equal(tc.want, nameable.UnescapeRunes(tc.in, tc.chars...), "UnescapeRunes(%q, %q)", tc.in, tc.chars)
	}
}

// '\n' and its reserved letter 'n' can't both be requested at once -- an escaped 'n' and an escaped newline would then
// be indistinguishable when unescaping.
func TestEscapeRunesAndUnescapeRunesPanicOnReservedCollision(t *testing.T) {
	c := check.New(t)
	c.Panics(func() { nameable.EscapeRunes("x", '\n', 'n') }, "EscapeRunes")
	c.Panics(func() { nameable.UnescapeRunes("x", '\n', 'n') }, "UnescapeRunes")
}
