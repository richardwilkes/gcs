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
	"image"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/geom"
)

// pdfTextSampleString is the one line of text pdfTextSamplePDF draws.
const pdfTextSampleString = "Hello World"

// pdfTextSamplePDF is a one-page 200x100 pt document that draws pdfTextSampleString in 24 pt Helvetica, a standard-14
// font that needs nothing embedded. It is spelled out here rather than being carried as a test file so that what the
// assertions below rely on is visible right next to them. No xref is supplied (startxref 0), so the engine rebuilds
// one, which is why the object offsets don't have to be maintained by hand. The content stream's /Length is exact.
const pdfTextSamplePDF = `%PDF-1.7
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 100] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>
endobj
4 0 obj
<< /Length 42 >>
stream
BT
/F1 24 Tf
20 50 Td
(Hello World) Tj
ET
endstream
endobj
5 0 obj
<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>
endobj
trailer
<< /Root 1 0 R /Size 6 >>
startxref
0
%%EOF
`

// newPDFTextSampleRenderer writes pdfTextSamplePDF to a temporary file and opens a renderer over it, closing the
// renderer when the test ends. A scale adjustment of 1 is used so that the logical space the renderer's public API
// speaks in and the pixel space the extracted text works in are the same, which lets the assertions talk about
// coordinates without having to undo a scaling first. The callbacks do nothing: these tests drive the extraction
// themselves rather than waiting to be told about it.
func newPDFTextSampleRenderer(t *testing.T) *PDFRenderer {
	t.Helper()
	return newPDFTextSampleRendererAtScale(t, geom.NewPoint(1, 1))
}

// newPDFTextSampleRendererAtScale is newPDFTextSampleRenderer with the image scale adjustment spelled out. That
// adjustment is what a display's pixel density produces -- 1 on an ordinary display, 0.5 on a Retina one -- and it is
// the whole of the difference between the pixel space the extracted text works in and the logical space the renderer's
// public API speaks in, so a test that wants to see that conversion actually happen asks for a scale that isn't 1.
func newPDFTextSampleRendererAtScale(t *testing.T, scaleAdjust geom.Point) *PDFRenderer {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sample.pdf")
	if err := os.WriteFile(path, []byte(pdfTextSamplePDF), 0o600); err != nil {
		t.Fatal(err)
	}
	pdf, err := NewPDFRenderer(path, 72, scaleAdjust, func(_ int) {}, func(_ int) {})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pdf.Close)
	return pdf
}

// pdfBlockSharedQueue occupies the shared render queue's single worker for the remainder of the test, so that work a
// renderer hands to the queue stays queued instead of being done behind the test's back. The accessors that answer a
// live interaction -- SearchMatches among them -- ask for a page's text extraction through the queue rather than
// leaving it to the caller, and a test that wants to see the request sitting there, and to decide for itself when it
// gets answered, can't have the worker racing it to the answer.
//
// Call this before opening the renderer under test. Cleanups run in reverse order, so that has the renderer closed --
// which is what takes it back out of the queue -- while the worker is still parked here, rather than the worker being
// let loose on a renderer that is in the middle of being closed.
func pdfBlockSharedQueue(t *testing.T) {
	t.Helper()
	started := make(chan string)
	blocker := newFakePDFRenderer(pdfQueue, "blocker", started)
	pdfQueue.submitUserSignal(blocker)
	expectRender(t, started, "blocker")
	t.Cleanup(func() { blocker.releaseRender(t, false) })
}

// TestPDFRendererPagePointFromPointRoundTrip pins that pagePointFromPoint is the exact inverse of the scaling
// rectFromPageRect applies. Hit-testing a click against the extracted text runs a coordinate the other way through the
// same conversion the highlights and search hits come out of, so any disagreement between the two would show up as a
// caret that lands a pixel away from where the selection it belongs to is drawn.
func TestPDFRendererPagePointFromPointRoundTrip(t *testing.T) {
	// A half-scale adjustment is what a Retina display produces, and it is the case a rounding mistake shows up in:
	// every odd pixel coordinate lands on a half in logical space.
	p := &PDFRenderer{scaleAdjust: geom.NewPoint(0.5, 0.5)}
	for _, r := range []image.Rectangle{
		image.Rect(0, 0, 1, 1),
		image.Rect(1, 1, 2, 2),
		image.Rect(3, 7, 40, 90),
		image.Rect(101, 250, 640, 481),
		image.Rect(1274, 1650, 1275, 1651),
	} {
		if got := p.pagePointFromPoint(p.rectFromPageRect(r).Point); got != r.Min {
			t.Errorf("%v round-tripped to %v, want %v", r, got, r.Min)
		}
	}
}

// TestPDFRendererTextUnextracted pins that a page nobody has asked about yet reports its length as unknown rather than
// as zero. The distinction is what lets the dockable tell "there is no text here" apart from "the text isn't ready
// yet", and therefore whether it should try again once the text extracted callback fires.
func TestPDFRendererTextUnextracted(t *testing.T) {
	pdf := newPDFTextSampleRenderer(t)
	if length, known := pdf.TextLength(0); known {
		t.Errorf("the length of page 0 is reported as known (%d) before anything has been extracted", length)
	}
}

// TestPDFRendererTextExtraction runs a real document all the way through the text path: the request, the extraction on
// what would be the queue's worker, and then every accessor the selection UI uses. The extraction is driven directly
// rather than through the queue so that the test doesn't depend on the shared queue's timing; requestText is the same
// bookkeeping RequestText does under the lock, minus the submission that would hand the work to the worker.
func TestPDFRendererTextExtraction(t *testing.T) {
	pdf := newPDFTextSampleRenderer(t)
	if pdf.extractNextText() {
		t.Fatal("extractNextText did work when nothing had been asked for")
	}
	pdf.lock.Lock()
	queued := pdf.requestText(0)
	pdf.lock.Unlock()
	if !queued {
		t.Fatal("requestText didn't take the request for page 0")
	}
	if !pdf.extractNextText() {
		t.Fatal("extractNextText did nothing with page 0 asked for")
	}
	if pdf.extractNextText() {
		t.Fatal("extractNextText did work a second time for a page that was already extracted")
	}

	length, known := pdf.TextLength(0)
	if !known {
		t.Fatal("the length of page 0 is still unknown after it was extracted")
	}
	if length != len(pdfTextSampleString) {
		t.Errorf("page 0 holds %d characters, want %d", length, len(pdfTextSampleString))
	}
	if text := pdf.TextRange(0, 0, length); !strings.Contains(text, pdfTextSampleString) {
		t.Errorf("page 0 reads %q, want it to contain %q", text, pdfTextSampleString)
	}

	// A whole-page selection is one line here, so it paints as a single rectangle that has to cover something.
	hits := pdf.TextHighlights(0, 0, length)
	if len(hits) == 0 {
		t.Fatal("selecting all of page 0 produced no highlights")
	}
	for i, hit := range hits {
		if hit.Empty() {
			t.Errorf("highlight %d of the whole-page selection is empty: %v", i, hit)
		}
	}

	// Clicking in the middle of the rectangle the first character paints in has to put the caret at one end of that
	// character, since there is nothing else there to land on.
	first := pdf.TextHighlights(0, 0, 1)
	if len(first) != 1 {
		t.Fatalf("the first character painted as %d rectangles, want 1: %v", len(first), first)
	}
	index, ok := pdf.TextIndexAt(0, first[0].Center())
	if !ok {
		t.Fatal("TextIndexAt reported page 0 as unavailable after it was extracted")
	}
	if index > 1 {
		t.Errorf("clicking the middle of the first character put the caret at %d, want 0 or 1", index)
	}

	// The fixture's line is a pair of words, so the word at the caret is the first of them and the line is both.
	start, end, ok := pdf.TextWordAt(0, 0)
	if !ok {
		t.Fatal("TextWordAt reported page 0 as unavailable after it was extracted")
	}
	if word := pdf.TextRange(0, start, end); word != "Hello" {
		t.Errorf("the word at the start of page 0 is %q, want %q", word, "Hello")
	}
	if start, end, ok = pdf.TextLineAt(0, 0); !ok {
		t.Fatal("TextLineAt reported page 0 as unavailable after it was extracted")
	}
	if line := pdf.TextRange(0, start, end); line != pdfTextSampleString {
		t.Errorf("the line at the start of page 0 is %q, want %q", line, pdfTextSampleString)
	}
}

// pdfTextRequestState reports what the renderer's text bookkeeping holds: the pages whose text has been asked for but
// not yet extracted, and how many pages the text cache holds. Both live under the lock, which is where they are read
// from here.
func pdfTextRequestState(p *PDFRenderer) (want []int, cached int) {
	p.lock.Lock()
	defer p.lock.Unlock()
	return slices.Clone(p.textWant), len(p.textCache)
}

// pdfExtractSamplePage extracts the sample document's only page, the same way the tests above do it: the request is
// recorded under the lock and the extraction is driven directly, so that neither depends on the shared queue.
func pdfExtractSamplePage(t *testing.T, pdf *PDFRenderer) {
	t.Helper()
	pdf.lock.Lock()
	pdf.requestText(0)
	pdf.lock.Unlock()
	if !pdf.extractNextText() {
		t.Fatal("the extraction of page 0 didn't happen")
	}
}

// TestPDFRendererSearchMatchesEmptySearch pins that drawing a document nobody is searching costs nothing. Every visible
// page asks for its matches on every frame, so an empty search has to be answered on the spot, without pulling the text
// of each page that scrolls past into memory to prove there is nothing to find.
func TestPDFRendererSearchMatchesEmptySearch(t *testing.T) {
	pdfBlockSharedQueue(t)
	pdf := newPDFTextSampleRenderer(t)
	matches, ok := pdf.SearchMatches(0, "")
	if !ok {
		t.Error("an empty search on page 0 has no answer yet")
	}
	if matches != nil {
		t.Errorf("an empty search on page 0 matched %v", matches)
	}
	want, cached := pdfTextRequestState(pdf)
	if len(want) != 0 {
		t.Errorf("an empty search queued the extraction of %v", want)
	}
	if cached != 0 {
		t.Errorf("an empty search left %d pages in the text cache", cached)
	}
}

// TestPDFRendererSearchMatches runs the whole of the search path against a real document: the ask that arrives before
// there is anything to search, the extraction it puts in the queue, and then the hits themselves. The rectangles are
// compared against what the page's text reports directly, which is the same answer a render-time search used to bake
// into the image -- the point of moving the search out of the rendering being that the answer doesn't change.
func TestPDFRendererSearchMatches(t *testing.T) {
	pdfBlockSharedQueue(t)
	pdf := newPDFTextSampleRenderer(t)

	// Nothing has been extracted, so there is no answer yet -- and asking is what puts the extraction in the queue.
	if matches, ok := pdf.SearchMatches(0, "Hello"); ok || matches != nil {
		t.Errorf("searching page 0 answered %v ok=%v before its text was extracted", matches, ok)
	}
	want, _ := pdfTextRequestState(pdf)
	if !slices.Contains(want, 0) {
		t.Fatalf("searching page 0 didn't queue its text extraction; %v is queued", want)
	}
	if !pdf.extractNextText() {
		t.Fatal("extractNextText did nothing with the page the search asked for")
	}

	matches, ok := pdf.SearchMatches(0, "Hello")
	if !ok {
		t.Fatal("searching page 0 still has no answer after its text was extracted")
	}
	if len(matches) != 1 {
		t.Fatalf("searching page 0 for %q found %d matches, want 1: %v", "Hello", len(matches), matches)
	}
	text, err := pdf.doc.TextPage(0, pdf.dpi)
	if err != nil {
		t.Fatal(err)
	}
	if expected := pdf.convertMatches(text.Search("Hello", pdfMaxSearchHits)); !reflect.DeepEqual(matches, expected) {
		t.Errorf("searching page 0 for %q found %v, want %v", "Hello", matches, expected)
	}
}

// TestPDFRendererSearchMatchesNeedles pins what each kind of needle finds on the fixture's one line. The whitespace-only
// needle is the one worth having: the search field is a text field like any other, so a space typed into it must not
// light up every gap on the page.
func TestPDFRendererSearchMatchesNeedles(t *testing.T) {
	pdfBlockSharedQueue(t)
	pdf := newPDFTextSampleRenderer(t)
	pdfExtractSamplePage(t, pdf)
	for _, tc := range []struct {
		name   string
		search string
		want   int
	}{
		{name: "the first word", search: "Hello", want: 1},
		{name: "the second word", search: "World", want: 1},
		{name: "the whole line", search: pdfTextSampleString, want: 1},
		{name: "a letter that appears twice", search: "o", want: 2},
		{name: "text that isn't there", search: "xyz", want: 0},
		{name: "nothing but whitespace", search: "   ", want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			matches, ok := pdf.SearchMatches(0, tc.search)
			if !ok {
				t.Fatalf("searching page 0 for %q has no answer after its text was extracted", tc.search)
			}
			if len(matches) != tc.want {
				t.Errorf("searching page 0 for %q found %d matches, want %d: %v", tc.search, len(matches), tc.want,
					matches)
			}
			if tc.want == 0 && matches != nil {
				t.Errorf("searching page 0 for %q produced %v rather than nothing at all", tc.search, matches)
			}
		})
	}
}

// TestPDFRendererSearchMatchesCached pins that the matcher runs once per page per search rather than once per frame.
// Drawing asks for the matches of every visible page on every frame, so an answer that was recomputed each time would
// be doing the work of the whole visible document over and over while the search text sits still.
func TestPDFRendererSearchMatchesCached(t *testing.T) {
	pdfBlockSharedQueue(t)
	pdf := newPDFTextSampleRenderer(t)
	pdfExtractSamplePage(t, pdf)

	first, ok := pdf.SearchMatches(0, "Hello")
	if !ok || len(first) != 1 {
		t.Fatalf("searching page 0 for %q found %v ok=%v, want one match", "Hello", first, ok)
	}
	second, _ := pdf.SearchMatches(0, "Hello")
	if len(second) != 1 || &second[0] != &first[0] {
		t.Errorf("asking for the same search again produced %v rather than the answer already computed", second)
	}

	// A different search is a different answer, so it is computed and takes the memoized answer's place...
	other, _ := pdf.SearchMatches(0, "World")
	if len(other) != 1 {
		t.Fatalf("searching page 0 for %q found %v, want one match", "World", other)
	}
	if other[0] == first[0] {
		t.Errorf("searching for %q found the same rectangle %v that %q did", "World", other[0], "Hello")
	}

	// ...which means coming back to the first search recomputes it, rather than the two answers being confused for
	// each other.
	again, _ := pdf.SearchMatches(0, "Hello")
	if !reflect.DeepEqual(again, first) {
		t.Errorf("searching page 0 for %q a second time found %v, want %v", "Hello", again, first)
	}
	if &again[0] == &first[0] {
		t.Error("the answer for a search that was displaced by another was handed back rather than recomputed")
	}
}

// TestPDFRendererSearchMatchesScaled pins that the hits come back in the logical space the view draws in rather than in
// the pixel space the extracted text works in. On a Retina display the two differ by a factor of two, so a conversion
// that was skipped would put every underline at twice the offset and twice the size of the text it belongs to.
func TestPDFRendererSearchMatchesScaled(t *testing.T) {
	scaleAdjust := geom.NewPoint(0.5, 0.5)
	pdfBlockSharedQueue(t)
	pdf := newPDFTextSampleRendererAtScale(t, scaleAdjust)
	pdfExtractSamplePage(t, pdf)

	matches, ok := pdf.SearchMatches(0, "World")
	if !ok {
		t.Fatal("searching page 0 has no answer after its text was extracted")
	}
	text, err := pdf.doc.TextPage(0, pdf.dpi)
	if err != nil {
		t.Fatal(err)
	}
	pixels := text.Search("World", pdfMaxSearchHits)
	if len(matches) != len(pixels) {
		t.Fatalf("searching page 0 found %d matches, want the %d the page's text reports", len(matches), len(pixels))
	}
	for i, hit := range matches {
		want := geom.NewRect(float32(pixels[i].Min.X)*scaleAdjust.X, float32(pixels[i].Min.Y)*scaleAdjust.Y,
			float32(pixels[i].Dx())*scaleAdjust.X, float32(pixels[i].Dy())*scaleAdjust.Y)
		if hit != want {
			t.Errorf("match %d is %v, want %v", i, hit, want)
		}
	}
	if size := pdf.PageLogicalSize(0); !matches[0].In(geom.Rect{Size: size}) {
		t.Errorf("the match at %v doesn't lie within the %v the page occupies", matches[0], size)
	}
}

// TestPDFRendererSearchLeavesRenderedPagesAlone is the regression pin for what this whole arrangement is for: typing in
// the search field must not disturb what has been rendered. The hits used to come out of the rendering, so every
// keystroke marked every visible page stale and put the "Rendering page N…" overlay back up over pages that were
// already on screen. The page is rendered here for real -- the fixture is small enough that a real image is no trouble
// -- and then searched, one keystroke's worth of search text at a time.
func TestPDFRendererSearchLeavesRenderedPagesAlone(t *testing.T) {
	pdfBlockSharedQueue(t)
	pdf := newPDFTextSampleRenderer(t)
	pdf.SetWantedPages([]int{0})
	pdf.renderNext()
	page := pdf.CachedPage(0)
	if page == nil || page.Image == nil {
		t.Fatalf("page 0 didn't render: %+v", page)
	}
	if when, pending := pdf.PendingSince(0); pending {
		t.Fatalf("page 0 is still pending after it was rendered, since %v", when)
	}

	for _, search := range []string{"H", "He", "Hel", "Hell", "Hello", "Hello ", "Hello W"} {
		if _, ok := pdf.SearchMatches(0, search); !ok {
			// Only the first keystroke has any extraction to wait for; the rest are answered from the text it brought
			// in, which is what makes searching as you type free of the queue entirely.
			if search != "H" {
				t.Fatalf("searching for %q had to wait for a text extraction of its own", search)
			}
			if !pdf.extractNextText() {
				t.Fatal("extractNextText did nothing with the page the search asked for")
			}
			if _, ok = pdf.SearchMatches(0, search); !ok {
				t.Fatalf("searching for %q still has no answer after the extraction", search)
			}
		}
		switch again := pdf.CachedPage(0); {
		case again == nil:
			t.Fatalf("searching for %q threw away the rendered page", search)
		case again != page || again.Image != page.Image:
			t.Fatalf("searching for %q replaced the rendered page", search)
		}
		if when, pending := pdf.PendingSince(0); pending {
			t.Fatalf("searching for %q put page 0 back to pending, since %v", search, when)
		}
	}

	// The same holds for the wanted set being handed over again unchanged, which is what the view does on every scroll
	// and every resize: nothing is re-rendered and nothing goes back to pending.
	pdf.SetWantedPages([]int{0})
	if again := pdf.CachedPage(0); again != page {
		t.Error("asking for the same wanted pages again replaced the rendered page")
	}
	if _, pending := pdf.PendingSince(0); pending {
		t.Error("asking for the same wanted pages again put page 0 back to pending")
	}
}

// TestPDFRendererTextCacheHoldsCap pins that the text cache stays within its cap no matter how many pages are paged
// through. storeText is exercised directly rather than through recordText, since recordText also fires the callback
// and talks to the shared render queue, neither of which a bare renderer has any business doing. A nil text stands in
// for the extracted text, which the eviction bookkeeping never looks at.
func TestPDFRendererTextCacheHoldsCap(t *testing.T) {
	p := &PDFRenderer{textCache: make(map[int]*pdfTextEntry)}
	for pageNumber := range maxPDFTextPages * 3 {
		p.storeText(pageNumber, nil)
		if len(p.textCache) > maxPDFTextPages {
			t.Fatalf("the text cache holds %d entries after %d pages, want at most %d", len(p.textCache),
				pageNumber+1, maxPDFTextPages)
		}
	}
	// What survives is the most recent run of pages, since nothing has been asked for again along the way.
	for pageNumber := maxPDFTextPages * 2; pageNumber < maxPDFTextPages*3; pageNumber++ {
		if _, exists := p.textCache[pageNumber]; !exists {
			t.Errorf("page %d was evicted even though it is among the most recently stored", pageNumber)
		}
	}
}

// TestPDFRendererTextCacheEvictsLeastRecentlyUsed pins that it is the least recently used page that goes, rather than
// the one stored the longest ago. A page being read from is a page the selection is still working with, so touching it
// has to move it out of harm's way.
func TestPDFRendererTextCacheEvictsLeastRecentlyUsed(t *testing.T) {
	p := &PDFRenderer{textCache: make(map[int]*pdfTextEntry)}
	for pageNumber := range maxPDFTextPages {
		p.storeText(pageNumber, nil)
	}
	// Page 0 is the oldest entry and would be the next to go, but reading it makes it the newest.
	if _, exists := p.textFor(0); !exists {
		t.Fatal("page 0 isn't in the cache it was just stored in")
	}
	p.storeText(maxPDFTextPages, nil)
	if _, exists := p.textCache[0]; !exists {
		t.Error("page 0 was evicted even though it was the most recently used")
	}
	if _, exists := p.textCache[1]; exists {
		t.Error("page 1 survived even though it was the least recently used")
	}
}
