/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package pdf

import (
	"runtime"
)

type queueParams struct {
	pdf             *PDF
	params          params
	onlyIfInBacklog bool
}

type queueData struct {
	in      chan *queueParams
	work    chan *queueParams
	backlog []*queueParams
}

var queue = newQueue()

func newQueue() *queueData {
	q := &queueData{
		in:   make(chan *queueParams, runtime.NumCPU()*2),
		work: make(chan *queueParams),
	}
	go q.worker()
	go q.process()
	return q
}

// submit a pdf for rendering. This intentionally prioritizes the most recent submission over older ones, on the theory
// that the newest is the one attempting to be viewed right this moment. Note that we also only ever render one page at
// a time. This is a limitation of the underlying C library. In theory, rendering pages from different PDFs at the same
// time should be possible, but I get sporadic crashes if I do that.
func submit(pdf *PDF, onlyIfInBacklog bool) {
	queue.in <- &queueParams{
		pdf:             pdf,
		params:          *pdf.lastRequest,
		onlyIfInBacklog: onlyIfInBacklog,
	}
}

func (q *queueData) process() {
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

func (q *queueData) addToBacklog(p *queueParams) {
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

func (q *queueData) worker() {
	for p := range q.work {
		p.pdf.render(&p.params)
	}
}
