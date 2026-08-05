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
	"runtime"
	"slices"
)

// pdfRenderable is everything the render queue needs from a document renderer: a single call that renders whichever
// page that renderer currently considers the most important. *PDFRenderer is the only implementation the application
// uses; the interface exists so that the queue's ordering rules can be exercised without a real document.
type pdfRenderable interface {
	renderNext()
}

// pdfQueueParams is a message to the queue's process() goroutine. Everything the queue tracks -- the backlog and the
// sticky priority -- belongs to that goroutine and is never touched from anywhere else, so these messages are the only
// way to alter the queue's state.
type pdfQueueParams struct {
	pdf pdfRenderable
	// enqueue asks for a backlog entry for the renderer, i.e. "this document has a page that still needs rendering".
	enqueue bool
	// continuation marks an enqueue the renderer made for itself upon finishing a page, rather than one that came from
	// something the user did. See addToBacklog for how the two are placed differently.
	continuation bool
	// makePriority makes the renderer the queue's sticky priority, displacing whichever renderer held it before.
	makePriority bool
	// remove discards the renderer's backlog entry and relinquishes the priority if it holds it. The fields above,
	// other than pdf, are ignored when this is set.
	remove bool
	// idle is the worker reporting that it has finished the render it was handed and is ready for another. It carries
	// no renderer. Routing this through the same channel the submissions arrive on is what makes the ordering airtight:
	// a renderer submits its continuation before its render call returns, and therefore before this message is sent, so
	// the continuation has always been folded into the backlog by the time the next dispatch is chosen.
	idle bool
}

type pdfQueueData struct {
	in       chan *pdfQueueParams
	work     chan pdfRenderable
	backlog  []pdfRenderable
	priority pdfRenderable
}

var pdfQueue *pdfQueueData

// The queue is created here rather than in the variable's initializer because the renderer resubmits itself as it
// finishes each page, which makes the queue's initialization appear to be cyclic to the compiler.
func init() {
	pdfQueue = newPDFQueue()
}

// newPDFQueue creates a render queue along with the goroutines that drive it. Note that only one page is ever rendered
// at a time. pdfview is pure Go and serializes rendering within a single document internally, so this isn't needed for
// correctness -- the single worker is kept for simplicity and to bound the memory churn that rendering many large pages
// at once would cause.
func newPDFQueue() *pdfQueueData {
	q := &pdfQueueData{
		in:   make(chan *pdfQueueParams, runtime.NumCPU()*2),
		work: make(chan pdfRenderable),
	}
	go q.worker()
	go q.process()
	return q
}

// submitUserSignal queues a renderer up in response to something the user did: opening a document, scrolling, resizing
// or changing the search text. Such a submission goes to the front of the dispatch order, since it concerns a document
// someone is looking at right now. No page number is passed along, since the renderer picks the highest priority page
// that still needs to be rendered at the moment the work is actually started, which means the newest visibility
// information always wins.
func (q *pdfQueueData) submitUserSignal(pdf pdfRenderable) {
	q.in <- &pdfQueueParams{pdf: pdf, enqueue: true}
}

// submitContinuation queues a renderer up for its next page, which is what it does for itself each time it finishes
// one. These go to the back of the dispatch order, so that documents which are still catching up take turns rather than
// one of them monopolizing the worker.
func (q *pdfQueueData) submitContinuation(pdf pdfRenderable) {
	q.in <- &pdfQueueParams{pdf: pdf, enqueue: true, continuation: true}
}

// submitPriority makes the renderer the queue's sticky priority, which it holds until another renderer claims it or the
// renderer is removed. While it is held, that renderer's backlog entry -- including the continuations it submits for
// itself -- is dispatched ahead of everything else, so the focused document renders everything it wants before the
// documents in the background resume. Pass true for enqueue if the renderer has a page that needs rendering right now;
// a renderer that has nothing to do still claims the priority, so that it wins the instant scrolling or searching asks
// it for something new.
func (q *pdfQueueData) submitPriority(pdf pdfRenderable, enqueue bool) {
	q.in <- &pdfQueueParams{pdf: pdf, enqueue: enqueue, makePriority: true}
}

// submitRemoval drops the renderer from the queue, discarding any backlog entry it has and relinquishing the priority
// if it holds it. This is done when the renderer is closed, both so that no further rendering is attempted for it and
// so that the queue doesn't keep it -- and the sizable document buffer it holds onto -- alive.
//
// The send can't deadlock, even though it is made from the UI thread while the worker may be in the middle of a render:
// process() only ever waits on q.in and on the handoff to the worker, never on anything a renderer does, so it always
// comes back around to drain q.in. The worst case is that this waits for room in the buffer.
func (q *pdfQueueData) submitRemoval(pdf pdfRenderable) {
	q.in <- &pdfQueueParams{pdf: pdf, remove: true}
}

// process owns the backlog, the priority and the knowledge of whether the worker is busy. It hands the worker something
// to do whenever it is idle and there is something to do, and otherwise waits for the next message.
func (q *pdfQueueData) process() {
	idle := true
	for {
		if idle && len(q.backlog) != 0 {
			i := q.dispatchIndex()
			// The worker reports itself idle before returning to its receive, so at worst this waits for the moment it
			// takes to get back there.
			q.work <- q.backlog[i]
			q.backlog = slices.Delete(q.backlog, i, i+1)
			idle = false
		}
		p := <-q.in
		if p.idle {
			idle = true
			continue
		}
		q.handle(p)
	}
}

// handle applies one message to the queue's state.
func (q *pdfQueueData) handle(p *pdfQueueParams) {
	if p.remove {
		if q.priority == p.pdf {
			q.priority = nil
		}
		if i := slices.Index(q.backlog, p.pdf); i >= 0 {
			q.backlog = slices.Delete(q.backlog, i, i+1)
		}
		return
	}
	if p.makePriority {
		q.priority = p.pdf
	}
	if p.enqueue {
		q.addToBacklog(p)
	}
}

// dispatchIndex returns the index of the backlog entry that should be handed to the worker next, which is the priority
// renderer's entry if it has one and the front of the backlog otherwise. The backlog must not be empty.
func (q *pdfQueueData) dispatchIndex() int {
	if q.priority != nil {
		if i := slices.Index(q.backlog, q.priority); i >= 0 {
			return i
		}
	}
	return 0
}

// addToBacklog adds an entry for the submission's renderer, or repositions the one that is already present. A user
// signal goes to the front, since it means someone is looking at that document right now, while a continuation goes to
// the back and leaves an entry that is already present where it is. That is what keeps a document which is still
// catching up from starving one that was just opened: the continuations take turns among the busy documents, while
// anything the user touches jumps ahead of all of them.
func (q *pdfQueueData) addToBacklog(p *pdfQueueParams) {
	if i := slices.Index(q.backlog, p.pdf); i >= 0 {
		if p.continuation {
			return
		}
		copy(q.backlog[1:i+1], q.backlog[:i])
		q.backlog[0] = p.pdf
		return
	}
	if p.continuation {
		q.backlog = append(q.backlog, p.pdf)
	} else {
		q.backlog = slices.Insert(q.backlog, 0, p.pdf)
	}
}

func (q *pdfQueueData) worker() {
	for pdf := range q.work {
		pdf.renderNext()
		q.in <- &pdfQueueParams{idle: true}
	}
}
