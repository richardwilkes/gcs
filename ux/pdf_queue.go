// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	params          pdfParams
	onlyIfInBacklog bool
}

type pdfQueueData struct {
	in      chan *pdfQueueParams
	work    chan *pdfQueueParams
	backlog []*pdfQueueParams
}

var pdfQueue = newPDFQueue()

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
// theory that the newest is the one attempting to be viewed right this moment. Note that we also only ever render one
// page at a time. This is a limitation of the underlying C library. In theory, rendering pages from different PDFs at
// the same time should be possible, but I get sporadic crashes if I do that.
func submitPDF(pdf *PDFRenderer, onlyIfInBacklog bool) {
	pdfQueue.in <- &pdfQueueParams{
		pdf:             pdf,
		params:          *pdf.lastRequest,
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
		p.pdf.render(&p.params)
	}
}
