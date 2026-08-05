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
	"github.com/richardwilkes/unison"
)

// maxPDFPageCacheBytes is a soft cap on the number of pixel bytes retained for rendered pages. Pages are always
// rendered at the maximum zoom, since zooming is done by scaling the already-rendered image rather than by re-
// rendering, so a US Letter page costs roughly 71MB on a Retina display. That works out to about 7 pages within the
// cap there and about 28 on a 1x display. Pages that are currently wanted are never evicted, so the cap may be
// exceeded when a great many pages are visible at once.
const maxPDFPageCacheBytes = uint64(512 << 20)

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
	Matches    []geom.Rect
}

// PDFLink holds a single link on a page. If PageNumber if >= 0, then this is an internal link and the URI will be
// empty.
type PDFLink struct {
	Bounds     geom.Rect
	PageNumber int
	URI        string
}

// pdfCacheEntry holds a rendered page along with the information needed to manage its residency in the cache. An entry
// whose generation differs from the renderer's current generation is stale: it was rendered against an older search
// text, so its highlights are out of date, but it remains perfectly drawable and is kept on screen as a placeholder
// until the re-render for the current search arrives and replaces it.
type pdfCacheEntry struct {
	page       *PDFPage // Image is nil when Error is not nil
	bytes      uint64   // The number of pixel bytes the image holds; 0 for an error entry
	lastUsed   uint64   // The value of the renderer's monotonic use tick when this entry was last asked for
	generation int      // The renderer generation this page was rendered under
}

// PDFRenderer holds a PDFRenderer page renderer.
type PDFRenderer struct {
	// The fields from here through pageRenderedCallback are set during construction and are immutable afterwards, so
	// they may be read without holding the lock.
	scaleAdjust          geom.Point
	doc                  *pdfview.Document
	pageCount            int
	dpi                  int
	pageSizes            []geom.Size
	pageRenderedCallback func(pageNumber int)

	// The fields below are mutable and must only be accessed while holding the lock.
	lock       sync.Mutex
	search     string
	generation int          // Bumped when search text changes: in-flight renders are discarded, cache entries go stale
	want       []int        // The pages that should be rendered, in priority order
	wantSet    map[int]bool // The same pages as want, for quick membership tests
	cache      map[int]*pdfCacheEntry
	cacheBytes uint64
	useTick    uint64
	pending    map[int]time.Time // Wanted, but not yet rendered: the time first asked for, which drives the overlay
	toc        []*PDFTableOfContents
	tocLoaded  bool
	closed     bool

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
// The callback is invoked on the rendering goroutine each time a page becomes available, so it must not touch the UI
// directly.
func NewPDFRenderer(filePath string, ppi float32, scaleAdjust geom.Point, pageRenderedCallback func(pageNumber int)) (*PDFRenderer, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	var doc *pdfview.Document
	if doc, err = pdfview.New(data, 0); err != nil {
		return nil, errs.Wrap(err)
	}
	p := &PDFRenderer{
		ppi:                  ppi,
		scaleAdjust:          scaleAdjust,
		doc:                  doc,
		pageCount:            doc.PageCount(),
		pageRenderedCallback: pageRenderedCallback,
		wantSet:              make(map[int]bool),
		cache:                make(map[int]*pdfCacheEntry),
		pending:              make(map[int]time.Time),
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

// SetWantedPages sets the pages that should be rendered, in priority order, along with the text to search for within
// them. Pages that have already been rendered with the current search text are left as-is, while the others are queued
// up for rendering. Changing the search text doesn't throw away what has already been rendered, even though the search
// hits are drawn into the pages: the existing images stay in the cache and remain on screen as placeholders, still
// showing the previous search's highlights, and each is swapped out in place as its re-render arrives.
func (p *PDFRenderer) SetWantedPages(pages []int, search string) {
	p.lock.Lock()
	if p.closed || (search == p.search && slices.Equal(pages, p.want)) {
		p.lock.Unlock()
		return
	}
	searchChanged := search != p.search
	if searchChanged {
		p.search = search
		p.generation++
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
		if entry, exists := p.cache[pageNumber]; exists && entry.generation == p.generation {
			continue
		}
		needsRender = true
		// A page that is only pending because the search text changed is re-stamped on each change, rather than
		// keeping the stamp it already had. That restarts the overlay's timer with every keystroke, so the "Rendering
		// page N…" overlay doesn't flicker into view on top of the placeholder images that are still on screen while
		// the user is typing.
		if _, alreadyPending := p.pending[pageNumber]; searchChanged || !alreadyPending {
			p.pending[pageNumber] = now
		}
	}
	p.lock.Unlock()
	if needsRender {
		// This is a user signal -- the document was just opened, scrolled, resized or searched -- so it jumps ahead of
		// the documents that are merely working their way through what they still owe.
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
// wanted but hasn't been rendered for the current search text yet. A page that has a stale placeholder in the cache is
// therefore pending, too, since what is on screen for it doesn't reflect the current search.
func (p *PDFRenderer) PendingSince(pageNumber int) (requested time.Time, pending bool) {
	p.lock.Lock()
	defer p.lock.Unlock()
	requested, pending = p.pending[pageNumber]
	return requested, pending
}

// RequestRenderPriority makes this PDFRenderer the render queue's priority document, which is done when its view gains
// the focus. It keeps that status until another document claims it, so everything this one wants is rendered before the
// queue returns to the documents in the background. The claim is made even when nothing needs rendering at the moment,
// so that scrolling or searching within the focused document is served immediately.
func (p *PDFRenderer) RequestRenderPriority() {
	p.lock.Lock()
	closed := p.closed
	needsRender := false
	if !closed {
		_, needsRender = p.nextWantedPage()
	}
	p.lock.Unlock()
	if !closed {
		pdfQueue.submitPriority(p, needsRender)
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
	p.lock.Unlock()
	// Anything the queue is still holding for this renderer is dropped, so that neither a backlog entry nor the
	// priority keeps the renderer -- and the document buffer it holds onto -- alive. See submitRemoval for why sending
	// this from the UI thread while the worker may be in the middle of a render can't deadlock.
	pdfQueue.submitRemoval(p)
	p.doc.Release()
}

// renderNext renders the highest priority page that has been asked for but isn't in the cache yet. It is called on the
// PDF queue's worker goroutine. Neither the document, the image creation, nor the queue submission may be touched
// while the lock is held.
func (p *PDFRenderer) renderNext() {
	p.lock.Lock()
	if p.closed {
		p.lock.Unlock()
		return
	}
	pageNumber, ok := p.nextWantedPage()
	if !ok {
		p.lock.Unlock()
		return
	}
	generation := p.generation
	search := p.search
	p.lock.Unlock()

	p.loadTOC()

	rendered, err := p.doc.RenderPage(pageNumber, p.dpi, 100, search)
	if err != nil {
		// Cache the failure so that a page which can't be rendered doesn't get retried in a tight loop.
		p.recordPage(&PDFPage{Error: err, PageNumber: pageNumber}, 0, generation)
		return
	}
	links := p.convertLinks(rendered.Links)
	matches := p.convertMatches(rendered.SearchHits)
	width := rendered.Image.Rect.Dx()
	height := rendered.Image.Rect.Dy()

	// Creating the image copies a large chunk of memory, so don't bother if the result is already unwanted.
	p.lock.Lock()
	stale := p.isStale(pageNumber, generation)
	p.lock.Unlock()
	if stale {
		return
	}

	var img *unison.Image
	if img, err = retainPDFImageFromPixels(width, height, rendered.Image.Pix, p.scaleAdjust); err != nil {
		p.recordPage(&PDFPage{Error: err, PageNumber: pageNumber}, 0, generation)
		return
	}
	if !p.recordPage(&PDFPage{
		PageNumber: pageNumber,
		Image:      img,
		Links:      links,
		Matches:    matches,
	}, uint64(width)*uint64(height)*4, generation) {
		releasePDFImage(img)
	}
}

// recordPage adds a rendered page to the cache, unless it is no longer relevant, in which case false is returned and
// the caller is responsible for releasing any image the page holds. An entry that is already present for the page is
// replaced, which is the normal path when the search text has changed and the placeholder rendered under the previous
// search is being superseded. On success, the page rendered callback is fired and another render is queued up if any
// wanted page still hasn't been rendered.
func (p *PDFRenderer) recordPage(page *PDFPage, bytes uint64, generation int) bool {
	p.lock.Lock()
	if p.isStale(page.PageNumber, generation) {
		p.lock.Unlock()
		return false
	}
	if old, exists := p.cache[page.PageNumber]; exists {
		// An entry is being replaced, which is what happens when a placeholder left over from an earlier search is
		// superseded. The cache's reference to the old image is dropped before the new entry goes in, so a page never
		// holds onto more than one image's worth of pixels. The disposal that reference drop may trigger is deferred to
		// the UI thread, and drawing asks the cache for the page it is about to draw, so no draw can be left holding an
		// image that has gone away.
		p.cacheBytes -= old.bytes
		if old.page != nil {
			releasePDFImage(old.page.Image)
		}
	}
	p.useTick++
	p.cache[page.PageNumber] = &pdfCacheEntry{
		page:       page,
		bytes:      bytes,
		lastUsed:   p.useTick,
		generation: generation,
	}
	p.cacheBytes += bytes
	delete(p.pending, page.PageNumber)
	p.evictCachedPages()
	p.lock.Unlock()

	p.pageRenderedCallback(page.PageNumber)

	p.lock.Lock()
	more := false
	if !p.closed {
		_, more = p.nextWantedPage()
	}
	p.lock.Unlock()
	if more {
		// A continuation, rather than a user signal: this renderer is simply asking for its next page, so it goes to
		// the back of the line and lets the other documents that are catching up have a turn.
		pdfQueue.submitContinuation(p)
	}
	return true
}

// isStale returns true if a render of the given page from the given generation is no longer of any use. The lock must
// be held when calling this.
func (p *PDFRenderer) isStale(pageNumber, generation int) bool {
	return p.closed || generation != p.generation || !p.wantSet[pageNumber]
}

// nextWantedPage returns the highest priority page that hasn't been rendered for the current generation yet, which is
// either one that has nothing in the cache at all or one whose cached entry is a stale placeholder from an earlier
// search. The lock must be held when calling this.
func (p *PDFRenderer) nextWantedPage() (pageNumber int, exists bool) {
	for _, one := range p.want {
		if entry, cached := p.cache[one]; !cached || entry.generation != p.generation {
			return one, true
		}
	}
	return 0, false
}

// discardCachedPages releases every rendered page and clears the pending state. This is only done when the renderer is
// closed; a search change keeps what has been rendered around as placeholders. The lock must be held when calling this.
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
