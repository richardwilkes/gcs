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
)

type pdfQueueParams struct {
	pdf             *PDFRenderer
	onlyIfInBacklog bool
}

type pdfQueueData struct {
	in      chan *pdfQueueParams
	work    chan *pdfQueueParams
	backlog []*pdfQueueParams
}

var pdfQueue *pdfQueueData

// The queue is created here rather than in the variable's initializer because the renderer resubmits itself as it
// finishes each page, which makes the queue's initialization appear to be cyclic to the compiler.
func init() {
	pdfQueue = newPDFQueue()
}

func newPDFQueue() *pdfQueueData {
	q := &pdfQueueData{
		in:   make(chan *pdfQueueParams, runtime.NumCPU()*2),
		work: make(chan *pdfQueueParams),
	}
	go q.worker()
	go q.process()
	return q
}

// submitPDF a pdf for rendering. This intentionally prioritizes the most recent submission over older ones, on the
// theory that the newest is the one attempting to be viewed right this moment. No page number is passed along, since
// the renderer picks the highest priority page that still needs to be rendered at the moment the work is actually
// started, which means the newest visibility information always wins. Note that we also only ever render one page at a
// time. pdfview is pure Go and serializes rendering within a single document internally, so this isn't needed for
// correctness -- the single worker is kept for simplicity and to bound the memory churn that rendering many large
// pages at once would cause.
func submitPDF(pdf *PDFRenderer, onlyIfInBacklog bool) {
	pdfQueue.in <- &pdfQueueParams{
		pdf:             pdf,
		onlyIfInBacklog: onlyIfInBacklog,
	}
}

func (q *pdfQueueData) process() {
	for {
		if len(q.backlog) != 0 {
			select {
			case p := <-q.in:
				q.addToBacklog(p)
			case q.work <- q.backlog[len(q.backlog)-1]:
				q.backlog[len(q.backlog)-1] = nil
				q.backlog = q.backlog[:len(q.backlog)-1]
			}
		} else {
			p := <-q.in
			select {
			case q.work <- p:
			default:
				q.addToBacklog(p)
			}
		}
	}
}

func (q *pdfQueueData) addToBacklog(p *pdfQueueParams) {
	for i, b := range q.backlog {
		if b.pdf == p.pdf {
			copy(q.backlog[i:], q.backlog[i+1:])
			q.backlog[len(q.backlog)-1] = p
			return
		}
	}
	if !p.onlyIfInBacklog {
		q.backlog = append(q.backlog, p)
	}
}

func (q *pdfQueueData) worker() {
	for p := range q.work {
		p.pdf.renderNext()
	}
}
