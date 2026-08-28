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

func TestExtractPartsNoPlaceholders(t *testing.T) {
	c := check.New(t)
	got := nameable.ExtractParts("abc def", '@', '@')
	c.Equal([]nameable.Part{{Value: "abc def", Placeholder: false}}, got)
}

func TestExtractPartsEmptyInput(t *testing.T) {
	c := check.New(t)
	c.Equal(0, len(nameable.ExtractParts("", '@', '@')))
}

func TestExtractPartsSinglePlaceholderWithSurroundingText(t *testing.T) {
	c := check.New(t)
	got := nameable.ExtractParts("abc @Foo@ def", '@', '@')
	c.Equal([]nameable.Part{
		{Value: "abc ", Placeholder: false},
		{Value: "Foo", Placeholder: true},
		{Value: " def", Placeholder: false},
	}, got)
}

func TestExtractPartsWholeStringIsPlaceholder(t *testing.T) {
	c := check.New(t)
	got := nameable.ExtractParts("@Foo@", '@', '@')
	c.Equal([]nameable.Part{{Value: "Foo", Placeholder: true}}, got)
}

func TestExtractPartsLeadingPlaceholder(t *testing.T) {
	c := check.New(t)
	got := nameable.ExtractParts("@Foo@ trailing", '@', '@')
	c.Equal([]nameable.Part{
		{Value: "Foo", Placeholder: true},
		{Value: " trailing", Placeholder: false},
	}, got)
}

func TestExtractPartsMultiplePlaceholdersSeparatedByPlainText(t *testing.T) {
	c := check.New(t)
	// Regression test: closing a placeholder used to leave isPlaceholder set to true instead of resetting it to
	// false (the "Start new non-placeholder part" comment gave away the intent), so the plain-text run between two
	// placeholders was itself misidentified as a placeholder. That broke every multi-marker string; fixed now.
	got := nameable.ExtractParts("@A@ and @B@", '@', '@')
	c.Equal([]nameable.Part{
		{Value: "A", Placeholder: true},
		{Value: " and ", Placeholder: false},
		{Value: "B", Placeholder: true},
	}, got)
}

func TestExtractPartsEscapedDelimitersYieldNoPlaceholder(t *testing.T) {
	c := check.New(t)
	got := nameable.ExtractParts(`abc \@Foo\@ def`, '@', '@')
	c.Equal([]nameable.Part{{Value: `abc \@Foo\@ def`, Placeholder: false}}, got)
}

func TestExtractPartsDoubledEscapeBeforeOpenStillOpensPlaceholder(t *testing.T) {
	c := check.New(t)
	// Two escape runes cancel out (one literal backslash), so the following '@' is not escaped and still opens a
	// placeholder.
	got := nameable.ExtractParts(`abc \\@Foo@`, '@', '@')
	c.Equal([]nameable.Part{
		{Value: `abc \\`, Placeholder: false},
		{Value: "Foo", Placeholder: true},
	}, got)
}

func TestExtractPartsTripleEscapeBeforeOpenEscapesIt(t *testing.T) {
	c := check.New(t)
	// An odd number of escape runes leaves the final one live, so the '@' it precedes is escaped and no placeholder
	// is opened. Since open == close here, the trailing, unescaped '@' is then read as a fresh open attempt (there's
	// no separate "close" semantics while not already inside a placeholder) that never finds a match before end of
	// input -- so it round-trips as its own trailing literal part rather than merging into the first one, but
	// nothing is lost.
	got := nameable.ExtractParts(`abc \\\@Foo@`, '@', '@')
	c.Equal([]nameable.Part{
		{Value: `abc \\\@Foo`, Placeholder: false},
		{Value: "@", Placeholder: false},
	}, got)
}

func TestExtractPartsEmptyPlaceholderTreatedAsLiteralText(t *testing.T) {
	c := check.New(t)
	// An open immediately followed by a close (nothing between) is not a placeholder at all -- it's folded back
	// into plain text.
	got := nameable.ExtractParts("a@@b", '@', '@')
	c.Equal([]nameable.Part{
		{Value: "a", Placeholder: false},
		{Value: "@@b", Placeholder: false},
	}, got)
}

func TestExtractPartsUnterminatedPlaceholderPreservesOpenDelimiter(t *testing.T) {
	c := check.New(t)
	// Regression test: an opened-but-never-closed placeholder used to be flushed as plain text at end of input
	// with its open delimiter rune silently dropped (already consumed when the placeholder started, and never
	// written back out), so "abc @Foo" lost the '@' entirely instead of round-tripping as literal text. Fixed now:
	// the trailing flush backs up over the dropped delimiter before slicing.
	got := nameable.ExtractParts("abc @Foo", '@', '@')
	c.Equal([]nameable.Part{
		{Value: "abc ", Placeholder: false},
		{Value: "@Foo", Placeholder: false},
	}, got)
}

func TestExtractPartsControlCharBreaksInProgressPlaceholder(t *testing.T) {
	c := check.New(t)
	// A control rune (here, newline) aborts an in-progress placeholder scan, and the abandoned '@'..text is folded
	// back into a single plain-text run once the following delimiter is (re)opened. That reopened '@' then never
	// finds a close before end of input, so it survives as a preserved, unterminated literal too.
	got := nameable.ExtractParts("@Foo\nBar@ baz", '@', '@')
	c.Equal([]nameable.Part{
		{Value: "Foo\nBar", Placeholder: false},
		{Value: "@ baz", Placeholder: false},
	}, got)
}

func TestExtractPartsDistinctOpenAndCloseRunes(t *testing.T) {
	c := check.New(t)
	got := nameable.ExtractParts("a(b)c", '(', ')')
	c.Equal([]nameable.Part{
		{Value: "a", Placeholder: false},
		{Value: "b", Placeholder: true},
		{Value: "c", Placeholder: false},
	}, got)
}

func TestExtractPartsOpenRuneInsideOpenPlaceholderIsTreatedAsRestart(t *testing.T) {
	c := check.New(t)
	// A second, distinct open rune encountered while already inside a placeholder is treated as a misidentified
	// placeholder start: the first open and everything collected since are folded back into plain text, and
	// scanning restarts from the second open.
	got := nameable.ExtractParts("a(b(c)d", '(', ')')
	c.Equal([]nameable.Part{
		{Value: "a", Placeholder: false},
		{Value: "(b", Placeholder: false},
		{Value: "c", Placeholder: true},
		{Value: "d", Placeholder: false},
	}, got)
}

// ---- Mid-placeholder redirect (the "open rune inside an open placeholder" branch) ----

func TestExtractPartsRestartWithNothingBetweenTheTwoOpens(t *testing.T) {
	c := check.New(t)
	// The redirect fires immediately: the first open contributed nothing (start == i for it), so the folded-back
	// non-placeholder part is exactly the one-rune-wide open delimiter itself, never empty.
	got := nameable.ExtractParts("((x)", '(', ')')
	c.Equal([]nameable.Part{
		{Value: "(", Placeholder: false},
		{Value: "x", Placeholder: true},
	}, got)
}

func TestExtractPartsRestartChainsThroughMultipleOpens(t *testing.T) {
	c := check.New(t)
	// Each additional open re-triggers the redirect, folding back one more single-rune non-placeholder part; only
	// the last open in the run survives to actually open the placeholder.
	got := nameable.ExtractParts("(((x)", '(', ')')
	c.Equal([]nameable.Part{
		{Value: "(", Placeholder: false},
		{Value: "(", Placeholder: false},
		{Value: "x", Placeholder: true},
	}, got)
}

func TestExtractPartsRestartImmediatelyAfterARealClose(t *testing.T) {
	c := check.New(t)
	// This is not the redirect branch (isPlaceholder is false again right after a real close), but it's the
	// adjacent case worth pinning next to it: an unterminated placeholder attempt starting right where the prior
	// one closed collapses to plain text -- open delimiter included -- once nothing closes it before end of input.
	got := nameable.ExtractParts("(a)(b", '(', ')')
	c.Equal([]nameable.Part{
		{Value: "a", Placeholder: true},
		{Value: "(b", Placeholder: false},
	}, got)
}

func TestExtractPartsEscapedOpenInsidePlaceholderDoesNotRestart(t *testing.T) {
	c := check.New(t)
	// An escaped second open is never seen as `open`, so the redirect branch does not fire and the placeholder
	// keeps accumulating through it, escape rune included (unescaping is UnescapeRunes' job, not this one's).
	got := nameable.ExtractParts(`(a\(b)`, '(', ')')
	c.Equal([]nameable.Part{{Value: `a\(b`, Placeholder: true}}, got)
}

func TestExtractPartsRestartWithNoClosePriorToEndOfInput(t *testing.T) {
	c := check.New(t)
	// The redirect folds the first open back to plain text as usual, and the second, restarted placeholder attempt
	// is itself left unterminated at end of input -- but (per the unterminated-placeholder fix) it round-trips as
	// literal text rather than vanishing, giving two single-rune non-placeholder parts.
	got := nameable.ExtractParts("((", '(', ')')
	c.Equal([]nameable.Part{
		{Value: "(", Placeholder: false},
		{Value: "(", Placeholder: false},
	}, got)
}

// ---- Every path that appends a Part is either explicitly guarded against a zero-width span or is provably at
// least `open`'s (or the delimiter's) byte width -- so none of them should ever be able to produce a Part with an
// empty Value. These pin that invariant across the specific inputs most likely to hit each guard. ----

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

func TestExtractPartsLoneUnterminatedOpenRoundTripsAsLiteralText(t *testing.T) {
	c := check.New(t)
	// Regression test: the extreme case of the unterminated-placeholder data loss fixed above -- when the dangling
	// open is the entire input, it used to vanish completely (zero parts returned for a one-character input).
	// Fixed now: it round-trips as a single literal, non-placeholder part.
	c.Equal([]nameable.Part{{Value: "@", Placeholder: false}}, nameable.ExtractParts("@", '@', '@'))
}

func TestExtractPartsPanicsWhenOpenEqualsEscape(t *testing.T) {
	c := check.New(t)
	c.Panics(func() { nameable.ExtractParts("x", nameable.EscapeRune, '@') })
}

func TestExtractPartsPanicsWhenCloseEqualsEscape(t *testing.T) {
	c := check.New(t)
	c.Panics(func() { nameable.ExtractParts("x", '@', nameable.EscapeRune) })
}

func TestExtractSegmentsNoDelimiters(t *testing.T) {
	c := check.New(t)
	c.Equal([]string{"ABC"}, nameable.ExtractSegments("ABC", '|'))
}

func TestExtractSegmentsEmptyInput(t *testing.T) {
	c := check.New(t)
	c.Equal(0, len(nameable.ExtractSegments("", '|')))
}

func TestExtractSegmentsBasicSplit(t *testing.T) {
	c := check.New(t)
	c.Equal([]string{"A", "B", "C"}, nameable.ExtractSegments("A|B|C", '|'))
}

func TestExtractSegmentsLeadingDelimiterYieldsEmptyFirstSegment(t *testing.T) {
	c := check.New(t)
	c.Equal([]string{"", "A", "B"}, nameable.ExtractSegments("|A|B", '|'))
}

func TestExtractSegmentsTrailingDelimiterDropsEmptyFinalSegment(t *testing.T) {
	c := check.New(t)
	// Unlike a leading delimiter, a trailing one produces no corresponding empty segment: the loop only appends a
	// final segment when start < len(src).
	c.Equal([]string{"A", "B"}, nameable.ExtractSegments("A|B|", '|'))
}

func TestExtractSegmentsConsecutiveDelimitersYieldEmptySegment(t *testing.T) {
	c := check.New(t)
	c.Equal([]string{"A", "", "B"}, nameable.ExtractSegments("A||B", '|'))
}

func TestExtractSegmentsEscapedDelimiterIsNotASplitPoint(t *testing.T) {
	c := check.New(t)
	// The escape is left in place in the returned segment -- ExtractSegments only decides where to split, it
	// doesn't unescape; that's UnescapeRunes' job.
	got := nameable.ExtractSegments(`A\|B|C`, '|')
	c.Equal([]string{`A\|B`, "C"}, got)
}

func TestExtractSegmentsDoubledEscapeBeforeDelimiterStillSplits(t *testing.T) {
	c := check.New(t)
	got := nameable.ExtractSegments(`A\\|B`, '|')
	c.Equal([]string{`A\\`, "B"}, got)
}

func TestExtractSegmentsTrailingUnpairedEscapeIsKeptLiteral(t *testing.T) {
	c := check.New(t)
	got := nameable.ExtractSegments(`A|B\`, '|')
	c.Equal([]string{"A", `B\`}, got)
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

func TestEscapeRunesUsesReservedLetterForReservedRune(t *testing.T) {
	c := check.New(t)
	// A rune with a reserved letter (here, an actual newline) is escaped using that letter instead of the literal
	// control byte, so the escaped text stays a single visible line.
	c.Equal(`line one\nline two`, nameable.EscapeRunes("line one\nline two", '|', '\n'))
}

func TestUnescapeRunesTrailingDanglingEscapeIsPreserved(t *testing.T) {
	c := check.New(t)
	// An escape rune with nothing after it (the loop ends while still escaping) is written back out as-is instead of
	// being silently dropped.
	c.Equal(`abc\`, nameable.UnescapeRunes(`abc\`, '|'))
}

func TestEscapeRunesAndUnescapeRunesRoundTripReservedNewline(t *testing.T) {
	c := check.New(t)
	s := "line one\nline two"
	c.Equal(s, nameable.UnescapeRunes(nameable.EscapeRunes(s, '|', '\n'), '|', '\n'))
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
