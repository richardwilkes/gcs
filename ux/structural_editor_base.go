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
	"github.com/richardwilkes/unison"
)

// structuralEditorBase is the part shared by every dockable that holds a whole model and changes it in undoable steps:
// the target manager its widgets register with, the undo manager, the model along with the hash of it as of when the
// dockable last matched what it edits, and the content rebuild that follows every step. Each step captures the whole
// model before and after, so that a structural change is as undoable as a typed one. With the row drag state it
// embeds, it supplies everything structuralEditor asks of a dockable other than being a panel.
type structuralEditorBase[T gurps.Hashable] struct {
	rowDragState
	targetMgr *TargetMgr
	undoMgr   *unison.UndoManager
	model     T
	// originalHash is the hash of the model as it was loaded, last saved or last applied, which is what modelModified()
	// compares against. It is deliberately not part of the undo state: undoing past a save must leave the editor
	// showing as modified, since what it then holds is not what its file or sheet holds.
	originalHash uint64
	toolbar      *unison.Panel
	// clone returns a deep copy of a model, which is what the undo edits hold and what the model is replaced with when
	// one of them is applied.
	clone func(T) T
	// buildContent fills the content panel from the model. It is called for the initial build and again by every
	// sync().
	buildContent func()
}

// initEditor readies the base for the dockable that embeds it. The editor is the dockable itself, as its Self: it is
// what the rows name as their editor in a drag payload, so that one can be recognized as this editor's own, and what
// the target manager is rooted at.
func (s *structuralEditorBase[T]) initEditor(editor rowDragEditor, clone func(T) T, buildContent func()) {
	s.editor = editor
	s.targetMgr = NewTargetMgr(editor)
	s.undoMgr = unison.NewUndoManager(100, func(err error) { errs.Log(err) })
	s.clone = clone
	s.buildContent = buildContent
}

// setModel makes the model the one the editor holds and records its hash as the state modelModified compares against.
func (s *structuralEditorBase[T]) setModel(model T) {
	s.model = model
	s.originalHash = gurps.Hash64(model)
}

// initContent readies the content panel, wires the row drag handling to it and builds the content into it.
func (s *structuralEditorBase[T]) initContent(content *unison.Panel) {
	content.SetBorder(nil)
	content.SetLayout(&unison.FlexLayout{Columns: 1})
	s.install(s.editor, content)
	s.buildContent()
}

func (s *structuralEditorBase[T]) UndoManager() *unison.UndoManager {
	return s.undoMgr
}

// targetManager implements structuralEditor.
func (s *structuralEditorBase[T]) targetManager() *TargetMgr {
	return s.targetMgr
}

// modelModified reports whether the model differs from the state it was last loaded, saved or applied in.
func (s *structuralEditorBase[T]) modelModified() bool {
	return s.originalHash != gurps.Hash64(s.model)
}

// prepareUndo starts an undo edit that captures the model as it stands.
func (s *structuralEditorBase[T]) prepareUndo(title string) *unison.UndoEdit[T] {
	return &unison.UndoEdit[T]{
		ID:         unison.NextUndoID(),
		EditName:   title,
		UndoFunc:   func(e *unison.UndoEdit[T]) { s.applyModel(e.BeforeData) },
		RedoFunc:   func(e *unison.UndoEdit[T]) { s.applyModel(e.AfterData) },
		AbsorbFunc: func(_ *unison.UndoEdit[T], _ unison.Undoable) bool { return false },
		BeforeData: s.clone(s.model),
	}
}

// finishAndPostUndo captures the model as it now stands into the undo edit and posts it.
func (s *structuralEditorBase[T]) finishAndPostUndo(undo *unison.UndoEdit[T]) {
	undo.AfterData = s.clone(s.model)
	s.UndoManager().Add(undo)
}

// applyModel replaces the model with a copy of the given one and rebuilds the content from it; undo and redo use it.
func (s *structuralEditorBase[T]) applyModel(model T) {
	s.model = s.clone(model)
	s.sync()
}

// reorderRows implements rowDragEditor. The move is applied inside a single undo edit named title; a move that reports
// no change posts no edit and leaves the content as it is.
func (s *structuralEditorBase[T]) reorderRows(title string, move func() bool) {
	undo := s.prepareUndo(title)
	if move() {
		s.finishAndPostUndo(undo)
		s.sync()
	}
}

// editStructure implements structuralEditor.
func (s *structuralEditorBase[T]) editStructure(title string, mutate func(), focusKey string) {
	undo := s.prepareUndo(title)
	mutate()
	s.finishAndPostUndo(undo)
	s.sync()
	if focusKey != "" {
		s.focusOn(focusKey)
	}
}

// sync rebuilds the content from the model, preserving the keyboard focus and the scroll position across the rebuild.
func (s *structuralEditorBase[T]) sync() {
	focusRefKey := s.targetMgr.CurrentFocusRef()
	scrollRoot := s.content.ScrollRoot()
	h, v := scrollRoot.Position()
	s.content.RemoveAllChildren()
	s.buildContent()
	p := s.editor.AsPanel()
	p.MarkForLayoutRecursively()
	p.MarkForRedraw()
	p.ValidateLayout()
	if root, ok := p.Self.(ModifiableRoot); ok {
		root.MarkModified(nil)
	}
	s.targetMgr.ReacquireFocus(focusRefKey, s.toolbar, s.content)
	scrollRoot.SetPosition(h, v)
}

// focusOn gives the keyboard focus to the widget with the given reference key and scrolls it into view. Panels call it
// after sync() has returned, since sync() restores the scroll position as its last act and would otherwise undo the
// scroll.
func (s *structuralEditorBase[T]) focusOn(refKey string) {
	if f := s.targetMgr.Find(refKey); f != nil {
		f.RequestFocus()
		f.ScrollIntoView()
	}
}
