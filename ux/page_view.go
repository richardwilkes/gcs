// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import "github.com/richardwilkes/unison"

// pageView is what the dockables that show a page of lists -- the character sheet, the template and the loot sheet --
// have in common around the page itself: the toolbar above it, the scroll panel it sits in, the target manager that
// finds its fields by reference key, and the search over its lists. Bringing such a page back into line with its
// model, whether by marking the dockable as modified or by rebuilding it, works with all four, so it is shared here.
type pageView struct {
	targetMgr     *TargetMgr
	toolbar       *unison.Panel
	scroll        *unison.ScrollPanel
	searchTracker *SearchTracker
	// awaitingUpdate is set while the dockable is being marked as modified, and MarkModified does nothing at all while
	// it is set (see ownerRecalculates for one consequence of that).
	awaitingUpdate bool
}

// viewState is the user's place on the page: the scroll position and the field holding the keyboard focus.
type viewState struct {
	scrollH  float32
	scrollV  float32
	focusRef *FocusRef
}

// captureViewState records the user's place on the page, so that restoreViewState can put it back once the page has
// been brought back into line with its model. That disturbs both parts of it: the tables are re-synced or built anew,
// which takes the focus away from a field in one of them, and the page is re-laid out, which moves the scroll
// position.
func (pv *pageView) captureViewState() viewState {
	h, v := pv.scroll.Position()
	return viewState{scrollH: h, scrollV: v, focusRef: pv.targetMgr.CurrentFocusRef()}
}

// restoreViewState puts the user's place on the page back. The focus goes to the field with the reference key of the
// one that held it -- a table built anew shares its key with the one it replaced -- or, when no such field remains,
// to the first focusable part of the toolbar or the page. The scroll position goes back last, since focusing a field
// scrolls it into view.
func (pv *pageView) restoreViewState(state viewState) {
	pv.targetMgr.ReacquireFocus(state.focusRef, pv.toolbar, pv.scroll.Content())
	pv.scroll.SetPosition(state.scrollH, state.scrollV)
}

// resync completes bringing the dockable back into line with its model, once whatever the dockable had to do first --
// rebuild its lists, recalculate its entity -- has been done and with the user's place captured before that: every
// panel is synced, the title is brought up to date, the search results, which name rows that may no longer exist, are
// found again, and the user's place is put back.
func (pv *pageView) resync(d unison.Dockable, state viewState) {
	DeepSync(d)
	UpdateTitleForDockable(d)
	// If this is marking the dockable as modified, that update is over once every panel has been synced. What remains
	// only puts the presentation back, and a change reported from within it -- a field committing its edit as the
	// focus leaves it -- is a fresh edit that calls for an update of its own rather than one to be dropped.
	pv.awaitingUpdate = false
	pv.searchTracker.Refresh()
	pv.restoreViewState(state)
}

// markModified is the body of MarkModified for a dockable with a page view: the user's place is captured, whatever the
// dockable has to do first is done -- which is what the prepare function, if there is one, is for -- and the page is
// brought back into line with the model. A call that arrives while the update is already under way is dropped.
func (pv *pageView) markModified(d unison.Dockable, prepare func()) {
	if pv.awaitingUpdate {
		return
	}
	pv.awaitingUpdate = true
	state := pv.captureViewState()
	if prepare != nil {
		prepare()
	}
	pv.resync(d, state)
}
