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

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// popupUndoRoot stands in for a SettingsDockable in the tests below: it supplies the undo manager that
// unison.UndoManagerFor locates and, like SettingsDockable.MarkModified, performs a DeepSync of its subtree whenever
// something within it is marked modified.
type popupUndoRoot struct {
	unison.Panel
	mgr *unison.UndoManager
}

func newPopupUndoRoot() *popupUndoRoot {
	r := &popupUndoRoot{mgr: unison.NewUndoManager(100, func(err error) { panic(err) })}
	r.Self = r
	return r
}

func (r *popupUndoRoot) UndoManager() *unison.UndoManager { return r.mgr }

// MarkModified implements ModifiableRoot.
func (r *popupUndoRoot) MarkModified(_ unison.Paneler) {
	DeepSync(r)
}

func newTestPopup(root *popupUndoRoot, value *attribute.Placement) *Popup[attribute.Placement] {
	p := NewPopup(nil, "", "Placement",
		func() attribute.Placement { return *value },
		func(v attribute.Placement) { *value = v },
		attribute.Automatic, attribute.Primary, attribute.Secondary)
	root.AddChild(p)
	return p
}

func selectedValue(p *Popup[attribute.Placement]) attribute.Placement {
	value, _ := p.Selected()
	return value
}

// TestPopupUndoRedo verifies that undoing and then redoing a Popup change works. Applying the undo drives the popup's
// selection back, which re-enters SelectionChangedCallback; if that re-entry adds another edit with the same undo ID
// while the manager is still processing the current one, the edit's AfterData is overwritten and the redo tail is
// released, making Redo a no-op.
func TestPopupUndoRedo(t *testing.T) {
	c := check.New(t)
	root := newPopupUndoRoot()
	value := attribute.Automatic
	p := newTestPopup(root, &value)

	p.Select(attribute.Secondary)
	c.Equal(attribute.Secondary, value, "selecting an item must apply it")
	c.True(root.mgr.CanUndo(), "changing the selection must add an undo edit")

	root.mgr.Undo()
	c.Equal(attribute.Automatic, value, "undo must restore the prior value")
	c.Equal(attribute.Automatic, selectedValue(p), "undo must restore the popup's selection")
	c.True(root.mgr.CanRedo(), "undo must leave the edit redoable")

	root.mgr.Redo()
	c.Equal(attribute.Secondary, value, "redo must reapply the value")
	c.Equal(attribute.Secondary, selectedValue(p), "redo must reapply the popup's selection")
	c.True(root.mgr.CanUndo(), "redo must leave the edit undoable again")
}

// TestPopupUndoPreservesNewerEdits verifies that undoing a Popup change does not destroy the edits made after it. The
// re-entrant Add() triggered by the undo would release the entire redo tail, discarding unrelated newer edits.
func TestPopupUndoPreservesNewerEdits(t *testing.T) {
	c := check.New(t)
	root := newPopupUndoRoot()
	first := attribute.Automatic
	second := attribute.Automatic
	p1 := newTestPopup(root, &first)
	p2 := newTestPopup(root, &second)

	p1.Select(attribute.Primary)
	p2.Select(attribute.Secondary)

	root.mgr.Undo() // undo the second popup's change
	c.Equal(attribute.Automatic, second, "undo must restore the second popup's prior value")
	c.Equal(attribute.Primary, first, "undoing the second popup must not disturb the first")

	root.mgr.Undo() // undo the first popup's change
	c.Equal(attribute.Automatic, first, "the older edit must still be undoable")

	root.mgr.Redo()
	c.Equal(attribute.Primary, first, "the older edit must still be redoable")
	root.mgr.Redo()
	c.Equal(attribute.Secondary, second, "the newer edit must survive undoing the older one")
}

// TestPopupUndoAddsOneEdit verifies that a single selection change produces exactly one undoable edit and that undoing
// it does not leave a second one behind.
func TestPopupUndoAddsOneEdit(t *testing.T) {
	c := check.New(t)
	root := newPopupUndoRoot()
	value := attribute.Automatic
	p := newTestPopup(root, &value)

	p.Select(attribute.Primary)
	root.mgr.Undo()
	c.False(root.mgr.CanUndo(), "a single selection change must produce exactly one undoable edit")
	c.Equal(attribute.Automatic, value, "undo must restore the prior value")
}
