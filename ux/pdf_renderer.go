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
	"slices"
	"sync"
	"time"

	"github.com/richardwilkes/pdfview"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
)

// maxPDFPageCacheBytes is a soft cap on the number of pixel bytes retained for rendered pages. Pages are always
// rendered at the maximum zoom, since zooming is done by scaling the already-rendered image rather than by re-
// rendering, so a US Letter page costs roughly 71MB on a Retina display. That works out to about 7 pages within the
// cap there and about 28 on a 1x display. Pages that are currently wanted are never evicted, so the cap may be
// exceeded when a great many pages are visible at once.
const maxPDFPageCacheBytes = uint64(512 << 20)

// maxPDFTextPages is a cap on the number of pages whose extracted text is retained. Text is tiny next to a page image
// -- a densely typeset page runs to a few hundred KB rather than the tens of MB its image costs -- but a long document
// paged through from end to end would still accumulate all of it, so the cap keeps that bounded. It is generous enough
// that it never gets in the way of a selection, since a selection can only span the pages a single drag manages to
// cross.
const maxPDFTextPages = 64

// PDFTableOfContents holds a table of contents entry.
type PDFTableOfContents struct {
	Title      string
	PageNumber int
	Children   []*PDFTableOfContents
}

// PDFPage holds a rendered PDFRenderer page.
type PDFPage struct {
	Error      error
	PageNumber int
	Image      *unison.Image
	Links      []*PDFLink
}

// PDFLink holds a single link on a page. If PageNumber if >= 0, then this is an internal link and the URI will be
// empty.
type PDFLink struct {
	Bounds     geom.Rect
	PageNumber int
	URI        string
}

// pdfCacheEntry holds a rendered page along with the information needed to manage its residency in the cache. What is
// rendered depends only on the page itself, so an entry never goes out of date while it is resident: it is either the
// image for that page or it isn't there at all.
type pdfCacheEntry struct {
	page     *PDFPage // Image is nil when Error is not nil
	bytes    uint64   // The number of pixel bytes the image holds; 0 for an error entry
	lastUsed uint64   // The value of the renderer's monotonic use tick when this entry was last asked for
}

// pdfTextEntry holds one page's extracted text along with the information needed to manage its residency in the text
// cache. A nil text means the page couldn't be read at all, which is the one failure pdfview reports from an
// extraction; a page that merely has no text on it comes back as a real, empty TextPage. The entry is kept for an
// unreadable page anyway so that the failure is answered from the cache rather than being rediscovered on every mouse
// move across the page.
//
// The entry also memoizes the last answer SearchMatches gave for the page, since drawing asks for it again on every
// frame while the search text sits unchanged. An empty search means nothing has been searched for yet, which needs no
// flag of its own: SearchMatches answers an empty needle without ever consulting the entry, so the string stored here
// is always one somebody actually searched for.
type pdfTextEntry struct {
	text     *pdfview.TextPage // nil when the page couldn't be read
	search   string            // The search text the matches below were computed for; empty until one has been
	matches  []geom.Rect       // The matches of search on this page, in logical space
	lastUsed uint64            // The value of the renderer's monotonic text use tick when this entry was last asked for
}

// PDFRenderer holds a PDFRenderer page renderer.
type PDFRenderer struct {
	// The fields from here through textExtractedCallback are set during construction and are immutable afterwards, so
	// they may be read without holding the lock.
	scaleAdjust           geom.Point
	doc                   *pdfview.Document
	pageCount             int
	dpi                   int
	pageSizes             []geom.Size
	pageRenderedCallback  func(pageNumber int)
	textExtractedCallback func(pageNumber int)

	// The fields below are mutable and must only be accessed while holding the lock.
	lock        sync.Mutex
	want        []int        // The pages that should be rendered, in priority order
	wantSet     map[int]bool // The same pages as want, for quick membership tests
	cache       map[int]*pdfCacheEntry
	cacheBytes  uint64
	useTick     uint64
	pending     map[int]time.Time // Wanted, but not yet rendered: the time first asked for, which drives the overlay
	textCache   map[int]*pdfTextEntry
	textWant    []int // The pages whose text should be extracted, in the order it was asked for
	textUseTick uint64
	toc         []*PDFTableOfContents
	tocLoaded   bool
	closed      bool

	// ppi is immutable, like the fields at the top of the struct, but is placed here because it fits within the
	// padding the flags above would otherwise leave behind.
	ppi float32
}

// NewPDFRenderer creates a new PDFRenderer page renderer. Everything it does is potentially slow: reading the whole
// file into memory, parsing it, and then asking every page for the size it will render at. A file that the OS has to
// fetch back from cloud storage first can make the read alone take minutes, so this is meant to be called from a
// background goroutine rather than from the UI thread.
//
// That is precisely why ppi and scaleAdjust are parameters rather than being looked up in here. They derive from
// gurps.GlobalSettings().General.MonitorPPI() and primaryDisplayScale(), both of which end up calling
// unison.PrimaryDisplay(), which is only safe to touch from the main/UI thread on some platforms. The caller must
// capture both values on the UI thread and hand them in. Do not "simplify" this by moving those lookups back inside
// this function.
//
// pageRenderedCallback is invoked on the rendering goroutine each time a page becomes available and
// textExtractedCallback is invoked on that same goroutine each time a page's text becomes available, so neither may
// touch the UI directly.
func NewPDFRenderer(filePath string, ppi float32, scaleAdjust geom.Point, pageRenderedCallback, textExtractedCallback func(pageNumber int)) (*PDFRenderer, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	var doc *pdfview.Document
	if doc, err = pdfview.New(data, 0); err != nil {
		return nil, errs.Wrap(err)
	}
	p := &PDFRenderer{
		ppi:                   ppi,
		scaleAdjust:           scaleAdjust,
		doc:                   doc,
		pageCount:             doc.PageCount(),
		pageRenderedCallback:  pageRenderedCallback,
		textExtractedCallback: textExtractedCallback,
		wantSet:               make(map[int]bool),
		cache:                 make(map[int]*pdfCacheEntry),
		pending:               make(map[int]time.Time),
		textCache:             make(map[int]*pdfTextEntry),
	}

	// We always render the PDF at the largest scale
	p.dpi = pdfRenderDPI(p.ppi, p.scaleAdjust)

	// Precompute the logical size of every page. PageRenderSize reports the exact pixel dimensions RenderPage will
	// produce at our dpi, so multiplying by scaleAdjust (the scale the images are created with) yields precisely what
	// the eventual Image.LogicalSize() will be, without having to render anything.
	p.pageSizes = make([]geom.Size, p.pageCount)
	for i := range p.pageSizes {
		var w, h int
		if w, h, err = doc.PageRenderSize(i, p.dpi); err != nil {
			// The page can't be loaded, which means RenderPage will fail at the same point later on and the slot will
			// only ever show an error, so just reserve a US Letter-sized slot for it.
			p.pageSizes[i] = pdfUSLetterLogicalSize(p.ppi, p.scaleAdjust)
			continue
		}
		p.pageSizes[i] = geom.NewSize(float32(w)*p.scaleAdjust.X, float32(h)*p.scaleAdjust.Y)
	}
	return p, nil
}

// pdfRenderDPI returns the resolution pages are rendered at on a display with the given pixels-per-inch and image
// scale adjustment. Pages are always rendered at the maximum zoom, since zooming scales the already-rendered image
// rather than re-rendering it, and the adjustment compensates for the extra pixels a high-density display's images
// carry, so that the result is the same physical size everywhere.
func pdfRenderDPI(ppi float32, scaleAdjust geom.Point) int {
	return int(((maxPDFDockableScale * ppi) / 100) / min(scaleAdjust.X, scaleAdjust.Y))
}

// pdfUSLetterLogicalSize returns the logical size a US Letter page has once rendered at pdfRenderDPI, which is what
// PageLogicalSize would report for such a page. It stands in wherever a real page size isn't available: for a page
// that can't be loaded, and for a document that hasn't been loaded yet.
func pdfUSLetterLogicalSize(ppi float32, scaleAdjust geom.Point) geom.Size {
	dpi := float32(pdfRenderDPI(ppi, scaleAdjust))
	return geom.NewSize(8.5*dpi*scaleAdjust.X, 11*dpi*scaleAdjust.Y)
}

// PageCount returns the total page count.
func (p *PDFRenderer) PageCount() int {
	return p.pageCount
}

// PageLogicalSize returns the logical size of the given 0-based page. This matches the rendered page's
// Image.LogicalSize() exactly and is available before any rendering has been done, so it may be used to lay out the
// space a page will occupy while it is still unrendered.
func (p *PDFRenderer) PageLogicalSize(pageNumber int) geom.Size {
	return p.pageSizes[pageNumber]
}

// TOC returns the table of contents, or nil if it hasn't been loaded yet.
func (p *PDFRenderer) TOC() []*PDFTableOfContents {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.toc
}

// SetWantedPages sets the pages that should be rendered, in priority order. Pages that have already been rendered are
// left as-is, while the others are queued up for rendering. Nothing here depends on what is being searched for: a
// rendered page is the page itself and nothing more, with the search hits drawn on top of the image by the view rather
// than into it, so typing in the search field never disturbs what has been rendered.
func (p *PDFRenderer) SetWantedPages(pages []int) {
	p.lock.Lock()
	if p.closed || slices.Equal(pages, p.want) {
		p.lock.Unlock()
		return
	}
	p.want = slices.Clone(pages)
	p.wantSet = make(map[int]bool, len(pages))
	for _, pageNumber := range pages {
		p.wantSet[pageNumber] = true
	}
	// Pages that are no longer wanted lose their pending stamp, so that coming back to one later starts its overlay
	// timer over rather than showing the overlay immediately.
	for pageNumber := range p.pending {
		if !p.wantSet[pageNumber] {
			delete(p.pending, pageNumber)
		}
	}
	now := time.Now()
	needsRender := false
	for _, pageNumber := range pages {
		if _, exists := p.cache[pageNumber]; exists {
			continue
		}
		needsRender = true
		// A page that is already pending keeps the stamp it has, so that its overlay timer runs from the moment it was
		// first asked for rather than being restarted by a change to the wanted set that has nothing to do with it.
		if _, alreadyPending := p.pending[pageNumber]; !alreadyPending {
			p.pending[pageNumber] = now
		}
	}
	p.lock.Unlock()
	if needsRender {
		// This is a user signal -- the document was just opened, scrolled or resized -- so it jumps ahead of the
		// documents that are merely working their way through what they still owe.
		pdfQueue.submitUserSignal(p)
	}
}

// CachedPage returns the rendered page for the given 0-based page number, or nil if it hasn't been rendered yet. The
// returned page must not be retained across UI events, since the image it holds may be disposed of once the page has
// been evicted from the cache.
func (p *PDFRenderer) CachedPage(pageNumber int) *PDFPage {
	p.lock.Lock()
	defer p.lock.Unlock()
	entry, exists := p.cache[pageNumber]
	if !exists {
		return nil
	}
	p.useTick++
	entry.lastUsed = p.useTick
	return entry.page
}

// PendingSince returns the time at which the given 0-based page was first asked for, along with true, if the page is
// wanted but hasn't been rendered yet. That time is what the "Rendering page N…" overlay's timer runs from, so that a
// page which arrives promptly never brings the overlay up at all.
func (p *PDFRenderer) PendingSince(pageNumber int) (requested time.Time, pending bool) {
	p.lock.Lock()
	defer p.lock.Unlock()
	requested, pending = p.pending[pageNumber]
	return requested, pending
}

// RequestText asks for the text of the given 0-based page to be extracted, if that hasn't already been done. Nothing
// happens for a page whose text is already available, so this is cheap to call repeatedly. The extraction itself runs
// on the render queue's worker, since it costs about what the text portion of a render of the same page costs and so
// has no business happening on the UI thread; a later call to one of the accessors below picks up the result.
//
// The request is submitted to the queue as a user signal rather than as a continuation. Text is only ever asked for
// because someone is pointing at, clicking on or dragging across a page right now, so it should not have to wait
// behind the pages that documents in the background are still working their way through.
func (p *PDFRenderer) RequestText(pageNumber int) {
	p.lock.Lock()
	queue := p.requestText(pageNumber)
	p.lock.Unlock()
	if queue {
		pdfQueue.submitUserSignal(p)
	}
}

// TextLength returns the number of characters on the given 0-based page. known is false only when the page's text
// hasn't been extracted yet, in which case the extraction is requested as a side effect and a subsequent call, once
// the text extracted callback has fired, will answer. A page that couldn't be read, and a page that simply holds no
// text at all, are both known with a length of 0: the attempt was made and there is nothing more to wait for.
func (p *PDFRenderer) TextLength(pageNumber int) (length int, known bool) {
	p.lock.Lock()
	var text *pdfview.TextPage
	text, known = p.textFor(pageNumber)
	queue := false
	if !known {
		queue = p.requestText(pageNumber)
	}
	p.lock.Unlock()
	if queue {
		pdfQueue.submitUserSignal(p)
	}
	return text.Len(), known
}

// TextIndexAt returns the insertion index nearest the given point, which is where a click at that point should put the
// caret. The point is in the page's logical space, the same space PDFPage.Links and the rectangles TextHighlights and
// SearchMatches report are in, and is converted to the pixel space the extracted text works in here so that callers
// never have to think about the difference. ok carries the same meaning it does for TextLength: false only while the
// page's text hasn't been extracted yet, and the extraction is requested as a side effect.
func (p *PDFRenderer) TextIndexAt(pageNumber int, pt geom.Point) (index int, ok bool) {
	text, ok := p.textForRequesting(pageNumber)
	return text.IndexAt(p.pagePointFromPoint(pt)), ok
}

// TextWordAt returns the half-open range of the word containing the given index, which is what a double-click on that
// index should select. An index that isn't within a word selects just that index's character. ok carries the same
// meaning it does for TextLength: false only while the page's text hasn't been extracted yet, and the extraction is
// requested as a side effect.
func (p *PDFRenderer) TextWordAt(pageNumber, index int) (start, end int, ok bool) {
	text, ok := p.textForRequesting(pageNumber)
	start, end = text.WordAt(index)
	return start, end, ok
}

// TextLineAt returns the half-open range of the line containing the given index, which is what a triple-click on that
// index should select. A line here is a run of characters the page's own geometry keeps on one baseline, not a
// paragraph. ok carries the same meaning it does for TextLength: false only while the page's text hasn't been
// extracted yet, and the extraction is requested as a side effect.
func (p *PDFRenderer) TextLineAt(pageNumber, index int) (start, end int, ok bool) {
	text, ok := p.textForRequesting(pageNumber)
	start, end = text.LineAt(index)
	return start, end, ok
}

// TextHighlights returns the rectangles that paint the selection [start, end) on the given 0-based page, in the page's
// logical space -- the same space SearchMatches reports its hits in, so a selection and a search hit are drawn the same
// way. A range covering no characters, a range whose characters paint nothing (the letters a ligature spells beyond its
// first have no width of their own, and pdfview drops their rectangles), and a page whose text isn't available, all
// return nil.
//
// Unlike the accessors above, this does not request an extraction for a page that hasn't had one. It is called from
// drawing, once for every visible page on every frame, and a repaint is not a reason to put a page's text at the front
// of the render queue. The pages a selection spans are pulled in by the caller asking TextLength for each of them,
// which happens when the selection changes rather than when the view repaints.
func (p *PDFRenderer) TextHighlights(pageNumber, start, end int) []geom.Rect {
	p.lock.Lock()
	text, _ := p.textFor(pageNumber)
	p.lock.Unlock()
	return p.convertMatches(text.Highlights(start, end))
}

// TextRange returns the text of the characters in [start, end) on the given 0-based page, which is what a copy of that
// selection should put on the clipboard. An empty string is returned for a range covering no characters and for a page
// whose text isn't available. Like TextHighlights, this doesn't request an extraction: a copy can only be asked for
// once a selection exists, and making the selection is what brought the text in.
func (p *PDFRenderer) TextRange(pageNumber, start, end int) string {
	p.lock.Lock()
	text, _ := p.textFor(pageNumber)
	p.lock.Unlock()
	return text.Text(start, end)
}

// pdfMaxSearchHits caps the number of matches reported for a single page. It is the cap the render-time search was
// made with back when the hits came out of the rendering, kept so that a needle matching a large fraction of a dense
// page's glyphs can't turn drawing that page into thousands of rectangles.
const pdfMaxSearchHits = 100

// SearchMatches returns the rectangles, in logical space, of the matches of search on the given 0-based page -- the
// same space TextHighlights reports a selection in, so a search hit and a selection are drawn the same way. At most
// pdfMaxSearchHits of them are reported. An empty search, a search holding nothing but whitespace, a page with no match
// on it and a page that couldn't be read all return nil. The returned slice is the renderer's own and must not be
// modified.
//
// ok carries the same meaning it does for TextLength: false only while the page's text hasn't been extracted yet, and
// the extraction is requested as a side effect. Unlike TextHighlights, which is also called from drawing and
// deliberately doesn't ask for one, this does: the hits are wanted for exactly the pages being drawn, and nothing else
// brings the text of those pages in, so there is nobody else to wait for. requestText drops a request for a page that
// has already been asked for, which is what makes calling this from every frame cheap. An empty search is the one case
// that asks for nothing, since merely drawing a document nobody is searching has no business extracting the text of
// every page that scrolls past.
func (p *PDFRenderer) SearchMatches(pageNumber int, search string) (matches []geom.Rect, ok bool) {
	if search == "" {
		return nil, true
	}
	text, ok := p.textForRequesting(pageNumber)
	if !ok {
		return nil, false
	}
	cached := false
	if matches, cached = p.cachedSearchMatches(pageNumber, search); cached {
		return matches, true
	}
	// The matcher is run without the lock held. A TextPage is immutable once extracted and safe to use concurrently,
	// and a needle that matches often on a dense page is enough work that doing it with every other page's accessors
	// blocked behind it would be felt. A nil text -- a page that couldn't be read -- searches to nothing here, which
	// caches the emptiness along with everything else.
	matches = p.convertMatches(text.Search(search, pdfMaxSearchHits))
	p.recordSearchMatches(pageNumber, search, matches)
	return matches, true
}

// cachedSearchMatches returns the matches already computed for the given 0-based page, if its entry is still in the
// text cache and holds an answer for this very search text. Drawing asks for the matches of every visible page on
// every frame while the answer only changes when the search text does, so the matcher runs once per page per search
// rather than once per frame.
func (p *PDFRenderer) cachedSearchMatches(pageNumber int, search string) (matches []geom.Rect, cached bool) {
	p.lock.Lock()
	defer p.lock.Unlock()
	entry, exists := p.textCache[pageNumber]
	if !exists || entry.search != search {
		return nil, false
	}
	return entry.matches, true
}

// recordSearchMatches memoizes what the matcher produced for the given 0-based page. The entry is looked up again
// rather than being carried in from the caller, since the page may have been evicted from the text cache while the
// matcher ran without the lock; an answer for a page that is no longer there is simply dropped.
func (p *PDFRenderer) recordSearchMatches(pageNumber int, search string, matches []geom.Rect) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if entry, exists := p.textCache[pageNumber]; exists {
		entry.search = search
		entry.matches = matches
	}
}

// textForRequesting returns the cached text for the given 0-based page, requesting an extraction if there isn't one
// yet. It is the shared body of the accessors that answer a live interaction, all of which want the same thing: the
// text if it is there, and the work started if it isn't. The returned text is nil whenever ok is false, and may also be
// nil when ok is true, since an unreadable page is cached as a nil text.
func (p *PDFRenderer) textForRequesting(pageNumber int) (text *pdfview.TextPage, ok bool) {
	p.lock.Lock()
	text, ok = p.textFor(pageNumber)
	queue := false
	if !ok {
		queue = p.requestText(pageNumber)
	}
	p.lock.Unlock()
	if queue {
		pdfQueue.submitUserSignal(p)
	}
	return text, ok
}

// textFor returns the cached text for the given 0-based page, stamping the entry as the most recently used one so that
// the pages being worked with are the last to be evicted. exists distinguishes "no extraction has been done" from
// "an extraction was done", the latter covering both a page with text, a page without any (an empty TextPage) and a
// page that couldn't be read (a nil text). The lock must be held when calling this.
func (p *PDFRenderer) textFor(pageNumber int) (text *pdfview.TextPage, exists bool) {
	entry, ok := p.textCache[pageNumber]
	if !ok {
		return nil, false
	}
	p.textUseTick++
	entry.lastUsed = p.textUseTick
	return entry.text, true
}

// requestText records that the given 0-based page's text should be extracted, returning true if the queue needs to be
// told about it. A page that is already extracted or already asked for adds nothing, so the queue only hears about
// pages that genuinely represent new work; every append made here is paired with exactly one queue submission, and the
// chain of continuations the worker submits for itself carries the rest. The lock must be held when calling this, and
// the queue submission must be made without it.
func (p *PDFRenderer) requestText(pageNumber int) bool {
	if p.closed || pageNumber < 0 || pageNumber >= p.pageCount {
		return false
	}
	if _, exists := p.textCache[pageNumber]; exists {
		return false
	}
	if slices.Contains(p.textWant, pageNumber) {
		return false
	}
	p.textWant = append(p.textWant, pageNumber)
	return true
}

// RequestRenderPriority makes this PDFRenderer the render queue's priority document, which is done when its view gains
// the focus. It keeps that status until another document claims it, so everything this one wants is rendered before the
// queue returns to the documents in the background. The claim is made even when nothing needs rendering at the moment,
// so that scrolling within the focused document, or a search or selection that needs a page's text, is served
// immediately.
func (p *PDFRenderer) RequestRenderPriority() {
	p.lock.Lock()
	closed := p.closed
	work := false
	if !closed {
		work = p.hasWork()
	}
	p.lock.Unlock()
	if !closed {
		pdfQueue.submitPriority(p, work)
	}
}

// Close releases the underlying document along with everything that has been rendered from it. It must be called on
// the UI thread and the PDFRenderer must not be used afterwards.
func (p *PDFRenderer) Close() {
	p.lock.Lock()
	p.closed = true
	p.want = nil
	clear(p.wantSet)
	p.discardCachedPages()
	p.textWant = nil
	clear(p.textCache)
	p.lock.Unlock()
	// Anything the queue is still holding for this renderer is dropped, so that neither a backlog entry nor the
	// priority keeps the renderer -- and the document buffer it holds onto -- alive. See submitRemoval for why sending
	// this from the UI thread while the worker may be in the middle of a render can't deadlock.
	pdfQueue.submitRemoval(p)
	p.doc.Release()
}

// renderNext performs one piece of this renderer's outstanding work: a text extraction if any has been asked for, and
// otherwise a render of the highest priority page that isn't in the cache yet. It is called on the PDF queue's worker
// goroutine. Neither the document, the image creation, nor the queue submission may be touched while the lock is held.
//
// Text comes first because it is only ever asked for by an interaction that is waiting on the answer -- a click that
// wants a caret, a drag that wants a selection -- while a render is filling in something the user can already see a
// placeholder for. An extraction is also far cheaper than a render, so letting it go first delays the pages by very
// little. The continuation the extraction submits for itself is what brings the worker straight back around to them.
func (p *PDFRenderer) renderNext() {
	if p.extractNextText() {
		return
	}
	p.lock.Lock()
	if p.closed {
		p.lock.Unlock()
		return
	}
	pageNumber, ok := p.nextWantedPage()
	p.lock.Unlock()
	if !ok {
		return
	}

	p.loadTOC()

	// No search is asked of the render. The hits are found in the page's extracted text by SearchMatches and drawn on
	// top of the image by the view, so what comes back here is the page and nothing else, and it stays good no matter
	// what is typed into the search field afterwards.
	rendered, err := p.doc.RenderPage(pageNumber, p.dpi, 0, "")
	if err != nil {
		// Cache the failure so that a page which can't be rendered doesn't get retried in a tight loop.
		p.recordPage(&PDFPage{Error: err, PageNumber: pageNumber}, 0)
		return
	}
	links := p.convertLinks(rendered.Links)
	width := rendered.Image.Rect.Dx()
	height := rendered.Image.Rect.Dy()

	// Creating the image copies a large chunk of memory, so don't bother if the result is already unwanted.
	p.lock.Lock()
	unwanted := p.isUnwanted(pageNumber)
	p.lock.Unlock()
	if unwanted {
		return
	}

	var img *unison.Image
	if img, err = retainPDFImageFromPixels(width, height, rendered.Image.Pix, p.scaleAdjust); err != nil {
		p.recordPage(&PDFPage{Error: err, PageNumber: pageNumber}, 0)
		return
	}
	if !p.recordPage(&PDFPage{
		PageNumber: pageNumber,
		Image:      img,
		Links:      links,
	}, uint64(width)*uint64(height)*4) {
		releasePDFImage(img)
	}
}

// recordPage adds a rendered page to the cache, unless it is no longer wanted, in which case false is returned and the
// caller is responsible for releasing any image the page holds. On success, the page rendered callback is fired and
// another render is queued up if any wanted page still hasn't been rendered.
//
// The insertion is never a replacement, so no image reference is ever dropped here: only a page that isn't in the cache
// is ever chosen for rendering, and the queue's single worker is the only thing that adds to the cache, so a page can't
// have gained an entry between being chosen and arriving here. Nothing takes a page's entry away and leaves it wanted,
// either -- eviction skips wanted pages, and closing discards the lot along with everything else -- so a page is
// rendered exactly once per stay in the cache.
func (p *PDFRenderer) recordPage(page *PDFPage, bytes uint64) bool {
	p.lock.Lock()
	if p.isUnwanted(page.PageNumber) {
		p.lock.Unlock()
		return false
	}
	p.useTick++
	p.cache[page.PageNumber] = &pdfCacheEntry{
		page:     page,
		bytes:    bytes,
		lastUsed: p.useTick,
	}
	p.cacheBytes += bytes
	delete(p.pending, page.PageNumber)
	p.evictCachedPages()
	p.lock.Unlock()

	p.pageRenderedCallback(page.PageNumber)

	p.lock.Lock()
	more := false
	if !p.closed {
		more = p.hasWork()
	}
	p.lock.Unlock()
	if more {
		// A continuation, rather than a user signal: this renderer is simply asking for its next page, so it goes to
		// the back of the line and lets the other documents that are catching up have a turn.
		pdfQueue.submitContinuation(p)
	}
	return true
}

// extractNextText extracts the text of the first page whose text has been asked for but hasn't been extracted yet,
// returning true if it did so. It is called on the PDF queue's worker goroutine and is the only place the document is
// ever asked for a page's text. As with rendering, the document must not be touched while the lock is held, so the
// page is chosen under the lock and the extraction itself is done without it.
func (p *PDFRenderer) extractNextText() bool {
	p.lock.Lock()
	if p.closed {
		p.lock.Unlock()
		return false
	}
	pageNumber, ok := p.nextWantedText()
	p.lock.Unlock()
	if !ok {
		return false
	}
	// The text is asked for at the one dpi everything is rendered at, so it is labeled for the very images the pointer
	// is over and never needs re-labeling with AtDPI: zooming scales the rendered image rather than re-rendering it, and
	// the text's pixel space scales along with it. An error means the page itself couldn't be read -- a page with
	// nothing on it to select is not an error, and comes back as an empty TextPage -- and such a page is cached as a
	// nil text rather than being left absent, so that the failure isn't rediscovered every time the pointer crosses it.
	// The render of the same page fails at the same point and reports it, so there is nothing more to say about it here.
	text, err := p.doc.TextPage(pageNumber, p.dpi)
	if err != nil {
		text = nil
	}
	p.recordText(pageNumber, text)
	return true
}

// recordText adds a page's extracted text to the text cache and then, exactly as recordPage does for a rendered page,
// fires the callback outside the lock and queues up another piece of work if anything is still outstanding.
func (p *PDFRenderer) recordText(pageNumber int, text *pdfview.TextPage) {
	p.lock.Lock()
	if p.closed {
		p.lock.Unlock()
		return
	}
	p.storeText(pageNumber, text)
	p.lock.Unlock()

	p.textExtractedCallback(pageNumber)

	p.lock.Lock()
	more := false
	if !p.closed {
		more = p.hasWork()
	}
	p.lock.Unlock()
	if more {
		pdfQueue.submitContinuation(p)
	}
}

// storeText inserts the entry for a page's extracted text, drops the request that asked for it, and evicts the least
// recently used entries until the cache is back within its cap. A page that is still waiting on an extraction is never
// chosen as the victim, which today can't come up -- a page joins the cache and leaves textWant in the same breath --
// but keeps the eviction from ever throwing away something that is about to be needed. The lock must be held when
// calling this.
func (p *PDFRenderer) storeText(pageNumber int, text *pdfview.TextPage) {
	p.textUseTick++
	p.textCache[pageNumber] = &pdfTextEntry{text: text, lastUsed: p.textUseTick}
	if i := slices.Index(p.textWant, pageNumber); i >= 0 {
		p.textWant = slices.Delete(p.textWant, i, i+1)
	}
	for len(p.textCache) > maxPDFTextPages {
		victim := -1
		var oldest uint64
		for candidate, entry := range p.textCache {
			if slices.Contains(p.textWant, candidate) {
				continue
			}
			if victim == -1 || entry.lastUsed < oldest {
				victim = candidate
				oldest = entry.lastUsed
			}
		}
		if victim == -1 {
			// Everything left is being waited on, so the cap has to be exceeded for the moment.
			return
		}
		delete(p.textCache, victim)
	}
}

// hasWork returns true if this renderer has anything left to do, whether that is a page to render or a page's text to
// extract. The lock must be held when calling this.
func (p *PDFRenderer) hasWork() bool {
	if len(p.textWant) != 0 {
		return true
	}
	_, exists := p.nextWantedPage()
	return exists
}

// nextWantedText returns the first page whose text has been asked for but hasn't been extracted yet. The page is left
// in textWant, so that a request arriving for it while the extraction is in flight doesn't ask for it a second time;
// recordText is what takes it out. Any page that has been extracted since it was asked for is dropped along the way
// rather than being rescanned forever. The lock must be held when calling this.
func (p *PDFRenderer) nextWantedText() (pageNumber int, exists bool) {
	for i, one := range p.textWant {
		if _, cached := p.textCache[one]; !cached {
			p.textWant = slices.Delete(p.textWant, 0, i)
			return one, true
		}
	}
	p.textWant = p.textWant[:0]
	return 0, false
}

// isUnwanted returns true if a render of the given page is no longer of any use, either because the renderer has been
// closed or because the page scrolled out of the wanted set while it was being rendered. The lock must be held when
// calling this.
func (p *PDFRenderer) isUnwanted(pageNumber int) bool {
	return p.closed || !p.wantSet[pageNumber]
}

// nextWantedPage returns the highest priority page that hasn't been rendered yet, which is simply the first wanted page
// with nothing in the cache. The lock must be held when calling this.
func (p *PDFRenderer) nextWantedPage() (pageNumber int, exists bool) {
	for _, one := range p.want {
		if _, cached := p.cache[one]; !cached {
			return one, true
		}
	}
	return 0, false
}

// discardCachedPages releases every rendered page and clears the pending state. This is only done when the renderer is
// closed, since nothing else invalidates a rendered page. The lock must be held when calling this.
func (p *PDFRenderer) discardCachedPages() {
	for _, entry := range p.cache {
		if entry.page != nil {
			releasePDFImage(entry.page.Image)
		}
	}
	clear(p.cache)
	clear(p.pending)
	p.cacheBytes = 0
}

// evictCachedPages drops the least recently used pages that aren't currently wanted until the cache is back under the
// byte cap. The lock must be held when calling this.
func (p *PDFRenderer) evictCachedPages() {
	for p.cacheBytes > maxPDFPageCacheBytes {
		victim := -1
		var oldest uint64
		for pageNumber, entry := range p.cache {
			if p.wantSet[pageNumber] {
				continue
			}
			if victim == -1 || entry.lastUsed < oldest {
				victim = pageNumber
				oldest = entry.lastUsed
			}
		}
		if victim == -1 {
			// Everything left is wanted, so the cap has to be exceeded for the moment.
			return
		}
		entry := p.cache[victim]
		delete(p.cache, victim)
		p.cacheBytes -= entry.bytes
		if entry.page != nil {
			releasePDFImage(entry.page.Image)
		}
	}
}

// loadTOC returns the table of contents, extracting it from the document if that hasn't already been done. Extracting
// it walks the entire outline tree, so it is done only once per document.
func (p *PDFRenderer) loadTOC() []*PDFTableOfContents {
	p.lock.Lock()
	if p.tocLoaded {
		defer p.lock.Unlock()
		return p.toc
	}
	p.lock.Unlock()
	toc := p.convertTOCEntries(p.doc.TableOfContents(p.dpi))
	p.lock.Lock()
	defer p.lock.Unlock()
	if !p.tocLoaded {
		p.toc = toc
		p.tocLoaded = true
	}
	return p.toc
}

func (p *PDFRenderer) convertTOCEntries(entries []*pdfview.TOCEntry) []*PDFTableOfContents {
	if len(entries) == 0 {
		return nil
	}
	toc := make([]*PDFTableOfContents, len(entries))
	for i, entry := range entries {
		toc[i] = &PDFTableOfContents{
			Title:      entry.Title,
			PageNumber: entry.PageNumber,
			Children:   p.convertTOCEntries(entry.Children),
		}
	}
	return toc
}

func (p *PDFRenderer) convertLinks(pageLinks []*pdfview.PageLink) []*PDFLink {
	if len(pageLinks) == 0 {
		return nil
	}
	links := make([]*PDFLink, len(pageLinks))
	for i, link := range pageLinks {
		links[i] = &PDFLink{
			Bounds:     p.rectFromPageRect(link.Bounds),
			PageNumber: link.PageNumber,
			URI:        link.URI,
		}
	}
	return links
}

func (p *PDFRenderer) convertMatches(hits []image.Rectangle) []geom.Rect {
	if len(hits) == 0 {
		return nil
	}
	matches := make([]geom.Rect, len(hits))
	for i, hit := range hits {
		matches[i] = p.rectFromPageRect(hit)
	}
	return matches
}

func (p *PDFRenderer) rectFromPageRect(r image.Rectangle) geom.Rect {
	return geom.NewRect(float32(r.Min.X)*p.scaleAdjust.X, float32(r.Min.Y)*p.scaleAdjust.Y,
		float32(r.Dx())*p.scaleAdjust.X, float32(r.Dy())*p.scaleAdjust.Y)
}

// pagePointFromPoint converts a point in a page's logical space back into the pixel space of the rendered page, which
// is the space the extracted text hit-tests in. It is the exact inverse of the scaling rectFromPageRect applies, so
// that a point taken from a rectangle that came out of there lands back on the pixel it started from.
func (p *PDFRenderer) pagePointFromPoint(pt geom.Point) image.Point {
	return image.Pt(int(xmath.Round(pt.X/p.scaleAdjust.X)), int(xmath.Round(pt.Y/p.scaleAdjust.Y)))
}

// unison hands back a single, shared *unison.Image for any two sets of pixels with the same content hash, so pages
// whose rendered pixels are identical -- two blank pages, even ones from different documents -- end up sharing one
// image. Disposing such an image while another holder still draws it panics, so references to the images we create for
// PDF pages are counted here and only the last holder disposes the image.
var (
	pdfImageRefsLock sync.Mutex
	pdfImageRefs     = make(map[*unison.Image]int)
)

// retainPDFImageFromPixels creates an image from the given pixels and takes a reference to it. Pass the result to
// releasePDFImage when done with it. The lock is held across the creation because unison's content-hash dedup may
// return an image that another renderer is concurrently releasing; serializing creation against the disposal in
// releasePDFImage is what closes that race.
func retainPDFImageFromPixels(width, height int, pixels []byte, scale geom.Point) (*unison.Image, error) {
	pdfImageRefsLock.Lock()
	defer pdfImageRefsLock.Unlock()
	img, err := unison.NewImageFromPixels(width, height, pixels, scale)
	if err != nil {
		return nil, err
	}
	pdfImageRefs[img]++
	return img, nil
}

// releasePDFImage drops a reference taken by retainPDFImageFromPixels, disposing the image once the last reference goes
// away. Disposal is deferred to the UI thread because drawing only happens there, which guarantees the disposal can't
// interleave with a draw that is still using the image. A retain may have resurrected the image by the time the task
// runs, so the count is rechecked before disposing.
func releasePDFImage(img *unison.Image) {
	if img == nil {
		return
	}
	pdfImageRefsLock.Lock()
	defer pdfImageRefsLock.Unlock()
	count, exists := pdfImageRefs[img]
	if !exists {
		return
	}
	count--
	pdfImageRefs[img] = count
	if count > 0 {
		return
	}
	unison.InvokeTask(func() {
		pdfImageRefsLock.Lock()
		defer pdfImageRefsLock.Unlock()
		if c, ok := pdfImageRefs[img]; ok && c == 0 {
			delete(pdfImageRefs, img)
			img.Dispose()
		}
	})
}
