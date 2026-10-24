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

// TestExtractPartsSinglePlaceholderWithSurroundingText pins how ExtractParts splits text into literal and placeholder
// parts. It yields no empty literal around a placeholder, an escaped delimiter delimits nothing and keeps its escape,
// and an empty, unclosed or control-rune-broken placeholder becomes literal text that keeps its open delimiter. An open
// inside a placeholder turns the earlier open and what follows it into literal text and restarts the placeholder.
func TestExtractPartsSinglePlaceholderWithSurroundingText(t *testing.T) {
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
		{src: `abc \@Foo\@ def`, begin: '@', end: '@', want: []nameable.Part{lit(`abc \@Foo\@ def`)}},
		// Paired escapes cancel out, so the '@' after them still opens a placeholder.
		{src: `abc \\@Foo@`, begin: '@', end: '@', want: []nameable.Part{lit(`abc \\`), ph("Foo")}},
		// An odd run escapes the first '@', so the second is an unclosed open that becomes its own literal part.
		{src: `abc \\\@Foo@`, begin: '@', end: '@', want: []nameable.Part{lit(`abc \\\@Foo`), lit("@")}},
		// An open immediately followed by a close is folded back into literal text.
		{src: "a@@b", begin: '@', end: '@', want: []nameable.Part{lit("a"), lit("@@b")}},
		// An unclosed open is flushed as literal text with its delimiter, even when it is the whole input.
		{src: "abc @Foo", begin: '@', end: '@', want: []nameable.Part{lit("abc "), lit("@Foo")}},
		{src: "@", begin: '@', end: '@', want: []nameable.Part{lit("@")}},
		// A control rune aborts the placeholder, keeping its open delimiter; the next '@' then opens one that never
		// closes.
		{src: "@Foo\nBar@ baz", begin: '@', end: '@', want: []nameable.Part{lit("@Foo\nBar"), lit("@ baz")}},
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

// ---- Every path that appends a Part is either explicitly guarded against a zero-width span or is provably at least a
// delimiter's byte width -- so none of them should ever be able to produce a Part with an empty Value. These pin that
// invariant across the specific inputs most likely to hit each guard. ----

func TestExtractPartsNeverProducesAnEmptyValuePart(t *testing.T) {
	for _, src := range []string{
		"@A@@B@",  // two placeholders with zero characters between their close and the next open
		"a@@@@b",  // two empty placeholders back to back, surrounded by text
		"@@",      // a lone empty placeholder, nothing else in the string
		"@@abc",   // an empty placeholder at the very start
		"abc@@",   // an empty placeholder at the very end
		"@@@",     // an empty placeholder followed by a dangling, unterminated open
		"@@@@",    // two empty placeholders, nothing surrounding either
		"@@@Foo@", // an empty placeholder immediately followed by a real one
		"@Foo@@@", // a real placeholder immediately followed by an empty one
		"((",      // adjacent opens (distinct pair) with the second left dangling
		"(((x)",   // a chain of redirects immediately followed by a real placeholder
		"(a)(b",   // a real placeholder immediately followed by a dangling open
		"@",       // a single, lone open with nothing after it at all
		"abc@",    // a dangling open at the very end, following plain text
		"\n@Foo@", // a control rune immediately followed by an open
		"@\nFoo@", // a control rune immediately inside an open placeholder
	} {
		for _, part := range nameable.ExtractParts(src, '@', '@') {
			if part.Value == "" {
				t.Errorf("ExtractParts(%q, ...) produced an empty-Value part (placeholder=%v)", src, part.Placeholder)
			}
		}
		for _, part := range nameable.ExtractParts(src, '(', ')') {
			if part.Value == "" {
				t.Errorf("ExtractParts(%q, '(', ')', ...) produced an empty-Value part (placeholder=%v)", src, part.Placeholder)
			}
		}
	}
}

// TestExtractPartsPanicsWhenOpenEqualsEscape pins that ExtractParts panics when either delimiter is the escape rune.
func TestExtractPartsPanicsWhenOpenEqualsEscape(t *testing.T) {
	c := check.New(t)
	c.Panics(func() { nameable.ExtractParts("x", nameable.EscapeRune, '@') }, "begin is the escape rune")
	c.Panics(func() { nameable.ExtractParts("x", '@', nameable.EscapeRune) }, "end is the escape rune")
}

// TestExtractSegmentsBasicSplit pins where ExtractSegments splits: a leading or doubled delimiter yields an empty
// segment but a trailing one does not, and an escaped delimiter is no split point and keeps its escape.
func TestExtractSegmentsBasicSplit(t *testing.T) {
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

func TestExtractSegmentsPanicsWhenDelimiterEqualsEscape(t *testing.T) {
	c := check.New(t)
	c.Panics(func() { nameable.ExtractSegments("x", nameable.EscapeRune) })
}

func TestEscapeRunesNoCharsIsANoOp(t *testing.T) {
	c := check.New(t)
	c.Equal(`a|b`, nameable.EscapeRunes(`a|b`))
}

func TestUnescapeRunesNoCharsIsANoOp(t *testing.T) {
	c := check.New(t)
	c.Equal(`a\|b`, nameable.UnescapeRunes(`a\|b`))
}

func TestUnescapeRunesTrailingDanglingEscapeIsPreserved(t *testing.T) {
	c := check.New(t)
	// An escape rune with nothing after it (the loop ends while still escaping) is written back out as-is instead of
	// being silently dropped.
	c.Equal(`abc\`, nameable.UnescapeRunes(`abc\`, '|'))
}

// TestEscapeRunesAndUnescapeRunesRoundTripReservedNewline pins that EscapeRunes writes a newline as its reserved
// letter, keeping the escaped text on one line, and that UnescapeRunes turns it back into the newline.
func TestEscapeRunesAndUnescapeRunesRoundTripReservedNewline(t *testing.T) {
	c := check.New(t)
	s := "line one\nline two"
	escaped := nameable.EscapeRunes(s, '|', '\n')
	c.Equal(`line one\nline two`, escaped)
	c.Equal(s, nameable.UnescapeRunes(escaped, '|', '\n'))
}

func TestEscapeRunesPanicsOnReservedCollision(t *testing.T) {
	c := check.New(t)
	// '\n' and its reserved letter 'n' can't both be requested at once -- an escaped 'n' and an escaped newline
	// would then be indistinguishable when unescaping.
	c.Panics(func() { nameable.EscapeRunes("x", '\n', 'n') })
}

func TestUnescapeRunesPanicsOnReservedCollision(t *testing.T) {
	c := check.New(t)
	c.Panics(func() { nameable.UnescapeRunes("x", '\n', 'n') })
}
