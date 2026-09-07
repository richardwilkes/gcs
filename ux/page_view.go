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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
)

// pageView is what the dockables that show a page of lists -- the character sheet, the template and the loot sheet --
// have in common: the file-backed panel at their root, the undo manager and the UI scale of the document they show,
// and, around the page itself, the toolbar above it, the scroll panel it sits in, the target manager that finds its
// fields by reference key, and the search over its lists. Putting such a dockable together (see initPageDockable and
// finishPageDockable) and bringing its page back into line with its model, whether by marking the dockable as
// modified or by rebuilding it, work with all of those, so they are shared here.
type pageView struct {
	fileBackedPanel
	undoMgr       *unison.UndoManager
	targetMgr     *TargetMgr
	toolbar       *unison.Panel
	scroll        *unison.ScrollPanel
	searchTracker *SearchTracker
	scale         int
	// awaitingUpdate is set while the dockable is being marked as modified, and MarkModified does nothing at all while
	// it is set (see ownerRecalculates for one consequence of that).
	awaitingUpdate bool
}

// pageDockable is a dockable built on a pageView, as the shared scaffolding sees it: the file-backed dockable itself,
// plus the parts the scaffolding has to call back into, which each dockable defines for itself.
type pageDockable interface {
	FileBackedDockable
	// keyToPanel returns the list a drop of the given kind of item is rerouted to (see installDropRerouting).
	keyToPanel(key *uti.DataType) *unison.Panel
	// createToolbar builds the toolbar and adds it to the dockable, above where the page's scroll panel will go.
	createToolbar()
}

// initPageDockable is the first half of putting a page dockable together, to be called before any of the page is
// built: it gives the dockable its undo manager, its initial UI scale, its scroll panel, its file-backed panel -- which
// hashes the content, so that must be complete by now -- the target manager rooted at it, its layout, and the
// rerouting of drops of list items to the list that takes them. The dockable is passed in explicitly, since it is the
// concrete value, not this embedded part of it, that the panel's Self, the file-backed panel and the target manager
// must refer to.
func (pv *pageView) initPageDockable(d pageDockable, filePath, extension string, saver func(filePath string) error, hashable gurps.Hashable) {
	pv.undoMgr = unison.NewUndoManager(200, func(err error) { errs.Log(err) })
	pv.scale = gurps.GlobalSettings().General.InitialSheetUIScale
	pv.scroll = unison.NewScrollPanel()
	pv.Self = d
	pv.initFileEditor(d, filePath, extension, saver, hashable)
	pv.targetMgr = NewTargetMgr(d)
	pv.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  align.Fill,
		VAlign:  align.Fill,
	})
	installDropRerouting(pv.AsPanel(), dropKeys, d.keyToPanel)
}

// finishPageDockable is the second half of putting a page dockable together, to be called once the page's content has
// been built: the content goes into the scroll panel, the toolbar is created and added above it, the scroll panel is
// added below that, and the Save and Save As commands are wired up.
func (pv *pageView) finishPageDockable(d pageDockable, content unison.Paneler) {
	pv.scroll.SetContent(content, behavior.Unmodified, behavior.Unmodified)
	pv.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	d.createToolbar()
	pv.AddChild(pv.scroll)
	pv.InstallCmdHandlers(SaveItemID, func(_ any) bool { return d.Modified() }, func(_ any) { pv.save(false) })
	pv.InstallCmdHandlers(SaveAsItemID, unison.AlwaysEnabled, func(_ any) { pv.save(true) })
}

// UndoManager implements unison.UndoManagerProvider.
func (pv *pageView) UndoManager() *unison.UndoManager {
	return pv.undoMgr
}

// attachedListProvider is a model that owns the items on its lists and matches them against the library sources
// they came from: a template or a loot sheet. An entity is one too, but readies itself as part of recalculating.
type attachedListProvider interface {
	gurps.ListProvider
	EnsureAttachments()
	SourceMatcher() *gurps.SrcMatcher
}

// prepareForPage readies a model to be shown on a page, whether for the first time or again after a rebuild: its items
// are attached to it, and the hashes its source matcher compares against the library sources are brought up to date.
func prepareForPage(model attachedListProvider) {
	model.EnsureAttachments()
	model.SourceMatcher().PrepareHashes(model)
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
