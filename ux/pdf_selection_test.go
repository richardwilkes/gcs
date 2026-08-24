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
	"testing"

	"github.com/richardwilkes/toolbox/v2/geom"
)

// TestPDFTextPosBefore pins the ordering of selection endpoints. Everything the selection does with its two endpoints
// -- deciding which of them the range starts at, which pages lie between them, and which way a drag ran -- rests on
// this comparison, and a selection made by dragging backwards through the text is the case that depends on it.
func TestPDFTextPosBefore(t *testing.T) {
	for _, tc := range []struct {
		name string
		a    pdfTextPos
		b    pdfTextPos
		want bool
	}{
		{name: "earlier index on the same page", a: pdfTextPos{page: 3, index: 10}, b: pdfTextPos{page: 3, index: 11}, want: true},
		{name: "later index on the same page", a: pdfTextPos{page: 3, index: 11}, b: pdfTextPos{page: 3, index: 10}, want: false},
		{name: "the same position", a: pdfTextPos{page: 3, index: 10}, b: pdfTextPos{page: 3, index: 10}, want: false},
		{name: "an earlier page", a: pdfTextPos{page: 2, index: 900}, b: pdfTextPos{page: 3, index: 0}, want: true},
		{name: "a later page", a: pdfTextPos{page: 4, index: 0}, b: pdfTextPos{page: 3, index: 900}, want: false},
		{name: "the start of the document", a: pdfTextPos{}, b: pdfTextPos{page: 0, index: 1}, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.a.before(tc.b); got != tc.want {
				t.Errorf("%+v.before(%+v) is %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

// TestPDFSelectionRangeFor pins what each page contributes to a selection. The interesting part is the middle of a
// multi-page selection: those pages contribute all of their text, so their length has to be known, and until it is
// there is no answer to give -- which is what tells the drawing and the copy to wait for the extraction rather than
// treating the page as empty.
func TestPDFSelectionRangeFor(t *testing.T) {
	// A length that is known for every page but the one nobody has extracted yet. The pages are deliberately given
	// different lengths, so a range that came from the wrong page would show up as the wrong number.
	lengths := map[int]int{0: 100, 1: 200, 2: 300, 4: 500}
	lengthOf := func(pageNumber int) (int, bool) {
		length, known := lengths[pageNumber]
		return length, known
	}
	for _, tc := range []struct {
		name      string
		lo        pdfTextPos
		hi        pdfTextPos
		page      int
		wantStart int
		wantEnd   int
		wantOK    bool
	}{
		{
			name: "wholly within one page",
			lo:   pdfTextPos{page: 1, index: 20}, hi: pdfTextPos{page: 1, index: 60},
			page: 1, wantStart: 20, wantEnd: 60, wantOK: true,
		},
		{
			name: "a zero-width selection contributes nothing",
			lo:   pdfTextPos{page: 1, index: 20}, hi: pdfTextPos{page: 1, index: 20},
			page: 1, wantStart: 20, wantEnd: 20, wantOK: true,
		},
		{
			name: "the first page of a multi-page selection runs to its end",
			lo:   pdfTextPos{page: 1, index: 20}, hi: pdfTextPos{page: 2, index: 60},
			page: 1, wantStart: 20, wantEnd: 200, wantOK: true,
		},
		{
			name: "the last page of a multi-page selection starts at its beginning",
			lo:   pdfTextPos{page: 1, index: 20}, hi: pdfTextPos{page: 2, index: 60},
			page: 2, wantStart: 0, wantEnd: 60, wantOK: true,
		},
		{
			name: "a page between the endpoints contributes all of itself",
			lo:   pdfTextPos{page: 0, index: 20}, hi: pdfTextPos{page: 2, index: 60},
			page: 1, wantStart: 0, wantEnd: 200, wantOK: true,
		},
		{
			name: "a page before the selection contributes nothing",
			lo:   pdfTextPos{page: 1, index: 20}, hi: pdfTextPos{page: 2, index: 60},
			page: 0, wantStart: 0, wantEnd: 0, wantOK: true,
		},
		{
			name: "a page after the selection contributes nothing",
			lo:   pdfTextPos{page: 1, index: 20}, hi: pdfTextPos{page: 2, index: 60},
			page: 4, wantStart: 0, wantEnd: 0, wantOK: true,
		},
		{
			name: "an unextracted page between the endpoints has no answer yet",
			lo:   pdfTextPos{page: 1, index: 20}, hi: pdfTextPos{page: 4, index: 60},
			page: 3, wantStart: 0, wantEnd: 0, wantOK: false,
		},
		{
			name: "an unextracted first page of a multi-page selection has no answer yet",
			lo:   pdfTextPos{page: 3, index: 20}, hi: pdfTextPos{page: 4, index: 60},
			page: 3, wantStart: 0, wantEnd: 0, wantOK: false,
		},
		{
			name: "an unextracted page outside the selection is still answered",
			lo:   pdfTextPos{page: 0, index: 20}, hi: pdfTextPos{page: 1, index: 60},
			page: 3, wantStart: 0, wantEnd: 0, wantOK: true,
		},
		{
			name: "an unextracted last page of a multi-page selection is answered from the endpoint",
			lo:   pdfTextPos{page: 1, index: 20}, hi: pdfTextPos{page: 3, index: 60},
			page: 3, wantStart: 0, wantEnd: 60, wantOK: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start, end, ok := pdfSelectionRangeFor(tc.lo, tc.hi, tc.page, lengthOf)
			if start != tc.wantStart || end != tc.wantEnd || ok != tc.wantOK {
				t.Errorf("page %d of the selection %+v..%+v is [%d,%d) ok=%v, want [%d,%d) ok=%v", tc.page, tc.lo,
					tc.hi, start, end, ok, tc.wantStart, tc.wantEnd, tc.wantOK)
			}
		})
	}
}

// TestPDFSelectionRangeForOrderedByCaller pins that the caller is the one responsible for putting the endpoints in
// document order, which is what orderedSelection does before this is reached. Handed a backwards pair, the function
// reports the whole document as being outside the selection rather than inventing a range, so a mistake there can't
// quietly produce a plausible-looking but wrong selection.
func TestPDFSelectionRangeForOrderedByCaller(t *testing.T) {
	lengthOf := func(_ int) (int, bool) {
		t.Helper()
		t.Error("the length of a page outside the selection was asked for")
		return 0, true
	}
	lo := pdfTextPos{page: 2, index: 60}
	hi := pdfTextPos{page: 1, index: 20}
	for _, pageNumber := range []int{0, 1, 2, 3} {
		start, end, ok := pdfSelectionRangeFor(lo, hi, pageNumber, lengthOf)
		if start != 0 || end != 0 || !ok {
			t.Errorf("page %d of the reversed selection is [%d,%d) ok=%v, want an empty, known range", pageNumber,
				start, end, ok)
		}
	}
}

// TestPDFTextPosSpan pins how a selection is derived from its anchor unit and the unit beneath the pointer: the span
// runs from the earlier of the two starts to the later of the two ends, whichever side of the anchor the pointer is
// on. This is what makes a drag begun with a double click grow and shrink by whole words, and what keeps the anchor's
// own word selected when the pointer comes back inside it.
func TestPDFTextPosSpan(t *testing.T) {
	word := func(page, start, end int) [2]pdfTextPos {
		return [2]pdfTextPos{{page: page, index: start}, {page: page, index: end}}
	}
	for _, tc := range []struct {
		name string
		a    [2]pdfTextPos
		b    [2]pdfTextPos
		want [2]pdfTextPos
	}{
		{name: "the pointer's word after the anchor's", a: word(1, 10, 15), b: word(1, 30, 36), want: word(1, 10, 36)},
		{name: "the pointer's word before the anchor's", a: word(1, 30, 36), b: word(1, 10, 15), want: word(1, 10, 36)},
		{name: "the pointer back inside the anchor's word", a: word(1, 10, 15), b: word(1, 12, 12), want: word(1, 10, 15)},
		{name: "overlapping words", a: word(1, 10, 15), b: word(1, 13, 20), want: word(1, 10, 20)},
		{
			name: "the pointer's word on a later page",
			a:    word(1, 10, 15),
			b:    word(3, 4, 9),
			want: [2]pdfTextPos{{page: 1, index: 10}, {page: 3, index: 9}},
		},
		{
			name: "the pointer's word on an earlier page",
			a:    word(3, 4, 9),
			b:    word(1, 10, 15),
			want: [2]pdfTextPos{{page: 1, index: 10}, {page: 3, index: 9}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lo, hi := pdfTextPosSpan(tc.a[0], tc.a[1], tc.b[0], tc.b[1])
			if lo != tc.want[0] || hi != tc.want[1] {
				t.Errorf("the span of %+v and %+v is %+v..%+v, want %+v..%+v", tc.a, tc.b, lo, hi, tc.want[0],
					tc.want[1])
			}
		})
	}
}

// TestPDFSearchHitBar pins the shape of the bar that marks a search hit: it spans the hit's full width, hugs the hit's
// bottom edge, and is a fraction of the hit's height rather than a fixed thickness, so that it scales with the text it
// underlines rather than swamping small type or vanishing under large.
func TestPDFSearchHitBar(t *testing.T) {
	const tolerance = 1e-4
	for _, hit := range []geom.Rect{
		geom.NewRect(10, 20, 100, 40),
		geom.NewRect(0, 0, 3, 7),
		geom.NewRect(512.5, 1030.25, 250.75, 21.5),
	} {
		bar := pdfSearchHitBar(hit)
		if bar.X != hit.X || bar.Width != hit.Width {
			t.Errorf("the bar for %v spans [%v,%v), want [%v,%v)", hit, bar.X, bar.Right(), hit.X, hit.Right())
		}
		if d := bar.Bottom() - hit.Bottom(); d > tolerance || d < -tolerance {
			t.Errorf("the bar for %v ends at %v, want the hit's bottom edge at %v", hit, bar.Bottom(), hit.Bottom())
		}
		if want := hit.Height * pdfSearchHitBarFraction; bar.Height != want {
			t.Errorf("the bar for %v is %v tall, want %v", hit, bar.Height, want)
		}
		if bar.Y < hit.Y {
			t.Errorf("the bar for %v starts above the hit, at %v", hit, bar.Y)
		}
	}
}

// TestPDFAutoScrollAmount pins the auto-scroll step. A pointer inside the viewport must produce exactly zero, since a
// nonzero step is what keeps the tick loop running, and a pointer well outside it must be capped, so that dragging to
// the far edge of the screen doesn't send the document past faster than the selection can follow.
func TestPDFAutoScrollAmount(t *testing.T) {
	const lo, hi = float32(100), float32(500)
	for _, tc := range []struct {
		name string
		pos  float32
		want float32
	}{
		{name: "well inside", pos: 300, want: 0},
		{name: "on the near edge", pos: lo, want: 0},
		{name: "on the far edge", pos: hi, want: 0},
		{name: "just past the near edge", pos: lo - 5, want: -5},
		{name: "just past the far edge", pos: hi + 5, want: 5},
		{name: "far past the near edge", pos: lo - 1000, want: -pdfAutoScrollMax},
		{name: "far past the far edge", pos: hi + 1000, want: pdfAutoScrollMax},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := pdfAutoScrollAmount(tc.pos, lo, hi); got != tc.want {
				t.Errorf("a pointer at %v within [%v,%v] scrolls by %v, want %v", tc.pos, lo, hi, got, tc.want)
			}
		})
	}
}

// TestPDFSelectionRangeAgainstRealText runs the range arithmetic against a real extracted page rather than a stubbed
// length, which is what pins that the two halves agree: the lengths pdfSelectionRangeFor asks the renderer for are
// indices into the same text TextRange later slices with the range it produced. The fixture is the one-line document
// the text tests use, so the whole document is a single page's worth of characters.
func TestPDFSelectionRangeAgainstRealText(t *testing.T) {
	pdf := newPDFTextSampleRenderer(t)

	// Before anything has been extracted, a selection running off the end of the page can't say how much of it is
	// covered, which is what holds a copy back until the text lands.
	lo := pdfTextPos{page: 0, index: 0}
	hi := pdfTextPos{page: 1, index: 0}
	if _, _, ok := pdfSelectionRangeFor(lo, hi, 0, pdf.TextLength); ok {
		t.Error("page 0 answered for its length before anything had been extracted")
	}

	// TextLength asked for the extraction as a side effect, so the worker has something to do now.
	if !pdf.extractNextText() {
		t.Fatal("asking for the length of page 0 didn't queue its extraction")
	}

	length, known := pdf.TextLength(0)
	if !known {
		t.Fatal("the length of page 0 is still unknown after it was extracted")
	}
	if length != len(pdfTextSampleString) {
		t.Fatalf("page 0 holds %d characters, want %d", length, len(pdfTextSampleString))
	}

	// A selection that runs off the end of the page now contributes the rest of the page.
	start, end, ok := pdfSelectionRangeFor(lo, hi, 0, pdf.TextLength)
	if !ok {
		t.Fatal("page 0 still has no length after it was extracted")
	}
	if start != 0 || end != length {
		t.Errorf("a selection from the start of page 0 onward covers [%d,%d), want [0,%d)", start, end, length)
	}
	if text := pdf.TextRange(0, start, end); text != pdfTextSampleString {
		t.Errorf("the whole of page 0 reads %q, want %q", text, pdfTextSampleString)
	}

	// A selection wholly within the page is handed straight through, and the text it names is the substring it should
	// be. "World" is the second word of the fixture's one line.
	const wordStart = len("Hello ")
	lo = pdfTextPos{page: 0, index: wordStart}
	hi = pdfTextPos{page: 0, index: length}
	if start, end, ok = pdfSelectionRangeFor(lo, hi, 0, pdf.TextLength); !ok {
		t.Fatal("a selection within page 0 has no range")
	}
	if start != wordStart || end != length {
		t.Errorf("a selection of the last word of page 0 covers [%d,%d), want [%d,%d)", start, end, wordStart, length)
	}
	if text, want := pdf.TextRange(0, start, end), pdfTextSampleString[wordStart:]; text != want {
		t.Errorf("the last word of page 0 reads %q, want %q", text, want)
	}
}
