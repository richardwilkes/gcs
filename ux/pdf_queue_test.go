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
	"time"
)

// pdfQueueHandshakeTimeout is how long these tests wait on a handshake with the queue before giving up. It is far
// longer than anything that should ever be needed, since it exists only so that a broken queue fails the test rather
// than hanging it.
const pdfQueueHandshakeTimeout = 10 * time.Second

// fakePDFRenderer stands in for a *PDFRenderer so that the queue's ordering can be driven one step at a time. Each call
// to renderNext announces itself on started and then waits to be released, which keeps the single worker occupied --
// and therefore the queue's state stable -- for as long as the test needs it to be.
type fakePDFRenderer struct {
	q       *pdfQueueData
	started chan string
	release chan bool
	name    string
}

func newFakePDFRenderer(q *pdfQueueData, name string, started chan string) *fakePDFRenderer {
	return &fakePDFRenderer{
		q:       q,
		started: started,
		release: make(chan bool),
		name:    name,
	}
}

func (f *fakePDFRenderer) renderNext() {
	f.started <- f.name
	if <-f.release {
		// This mirrors what PDFRenderer.recordPage does: the continuation is submitted before the render call returns.
		f.q.submitContinuation(f)
	}
}

// releaseRender lets a render that is in progress complete. Pass true for more to have the renderer submit a
// continuation for itself, which is what a real one does when it still has pages left to render.
func (f *fakePDFRenderer) releaseRender(t *testing.T, more bool) {
	t.Helper()
	select {
	case f.release <- more:
	case <-time.After(pdfQueueHandshakeTimeout):
		t.Fatalf("timed out releasing the render of %s", f.name)
	}
}

// expectRender waits for the worker to pick up the next render and verifies that it is the expected one. A mismatch is
// fatal rather than merely a failure, since every step that follows assumes the expected renderer is the one sitting in
// renderNext waiting to be released.
func expectRender(t *testing.T, started chan string, want string) {
	t.Helper()
	select {
	case name := <-started:
		if name != want {
			t.Fatalf("expected the queue to dispatch %s, but it dispatched %s", want, name)
		}
	case <-time.After(pdfQueueHandshakeTimeout):
		t.Fatalf("timed out waiting for the queue to dispatch %s", want)
	}
}

// TestPDFQueueUserSignalBeatsContinuations verifies that a document which is still catching up doesn't starve one the
// user just opened. The busy renderer resubmits itself after every page it finishes, which is what used to keep it
// perpetually at the head of the queue.
func TestPDFQueueUserSignalBeatsContinuations(t *testing.T) {
	q := newPDFQueue()
	started := make(chan string)
	busy := newFakePDFRenderer(q, "busy", started)
	newcomer := newFakePDFRenderer(q, "newcomer", started)

	// The busy document was opened first, so it has the worker.
	q.submitUserSignal(busy)
	expectRender(t, started, "busy")

	// While it is rendering, a second document is opened. That submission is made before the busy renderer is released,
	// so the continuation the busy renderer makes for itself arrives after it.
	q.submitUserSignal(newcomer)
	busy.releaseRender(t, true)

	// The newcomer wins, rather than the busy document rolling straight on to its next page.
	expectRender(t, started, "newcomer")

	// The busy document's continuation is still queued up, though, so it resumes as soon as the newcomer is done.
	newcomer.releaseRender(t, false)
	expectRender(t, started, "busy")
	busy.releaseRender(t, false)
}

// TestPDFQueueStickyPriority verifies that the renderer which claimed the priority keeps winning dispatch, including
// for the continuations it makes for itself, until it has nothing left to render.
func TestPDFQueueStickyPriority(t *testing.T) {
	q := newPDFQueue()
	started := make(chan string)
	focused := newFakePDFRenderer(q, "focused", started)
	background := newFakePDFRenderer(q, "background", started)

	// Get the background document rendering, so that the worker is busy while the rest of the queue's state is set up.
	q.submitUserSignal(background)
	expectRender(t, started, "background")

	// The focused document claims the priority and has work to do. The background document then asks for more, which
	// puts it at the front of the backlog, ahead of the focused document's entry.
	q.submitPriority(focused, true)
	q.submitUserSignal(background)

	// Position in the backlog doesn't matter: the priority document is dispatched first...
	background.releaseRender(t, false)
	expectRender(t, started, "focused")

	// ...and it keeps winning for as long as it resubmits itself, even though the background document has been waiting
	// the entire time.
	focused.releaseRender(t, true)
	expectRender(t, started, "focused")

	// Once the focused document has nothing left to render, the background document finally gets its turn.
	focused.releaseRender(t, false)
	expectRender(t, started, "background")
	background.releaseRender(t, false)
}

// TestPDFQueueRemoval verifies that removing a renderer, which is what closing one does, discards its pending entry and
// relinquishes the priority it was holding.
func TestPDFQueueRemoval(t *testing.T) {
	q := newPDFQueue()
	started := make(chan string)
	closing := newFakePDFRenderer(q, "closing", started)
	other := newFakePDFRenderer(q, "other", started)
	third := newFakePDFRenderer(q, "third", started)

	// The document that is about to be closed is rendering and holds the priority.
	q.submitUserSignal(closing)
	expectRender(t, started, "closing")
	q.submitPriority(closing, false)

	// At the point it is removed it has a pending entry of its own, sitting ahead of another document's entry in the
	// backlog.
	q.submitUserSignal(other)
	q.submitUserSignal(closing)
	q.submitRemoval(closing)

	// Its pending entry is gone, so the other document is what gets dispatched.
	closing.releaseRender(t, false)
	expectRender(t, started, "other")

	// The priority doesn't refer to it any longer either: a fresh submission from it is dispatched in backlog order
	// rather than jumping ahead of the document submitted after it.
	q.submitUserSignal(closing)
	q.submitUserSignal(third)
	other.releaseRender(t, false)
	expectRender(t, started, "third")
	third.releaseRender(t, false)
	expectRender(t, started, "closing")
	closing.releaseRender(t, false)
}
