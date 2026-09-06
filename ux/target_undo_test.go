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

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	checkenum "github.com/richardwilkes/unison/enums/check"
)

// newTestCheckBox returns a CheckBox backed by the given state, registered with the target manager under key and added
// to root.
func newTestCheckBox(root *popupUndoRoot, targetMgr *TargetMgr, key string, state *checkenum.Enum) *CheckBox {
	cb := NewCheckBox(targetMgr, key, "Flag",
		func() checkenum.Enum { return *state },
		func(v checkenum.Enum) { *state = v })
	root.AddChild(cb)
	return cb
}

// toggle simulates the user clicking the checkbox, which flips its state and then runs its click callback.
func toggle(cb *CheckBox) {
	if cb.State == checkenum.On {
		cb.State = checkenum.Off
	} else {
		cb.State = checkenum.On
	}
	cb.ClickCallback()
}

// TestTargetUndoRoutesToReplacementWidget verifies that an edit recorded by recordTargetUndo is applied to whichever
// widget holds the target key when it is undone or redone, rather than to the widget that recorded it. The editors
// rebuild their widgets whenever the model changes, so the widget that recorded the edit is usually gone by the time
// the edit is undone.
func TestTargetUndoRoutesToReplacementWidget(t *testing.T) {
	c := check.New(t)
	root := newPopupUndoRoot()
	targetMgr := NewTargetMgr(root)
	state := checkenum.Off
	original := newTestCheckBox(root, targetMgr, "flag", &state)
	c.Equal("flag", original.RefKey, "the target key must become the RefKey the target manager looks up")

	toggle(original)
	c.Equal(checkenum.On, state, "clicking must apply the new state")
	c.True(root.mgr.CanUndo(), "clicking must record an undo edit")

	// Simulate a rebuild: the original widget goes away and a replacement with the same key takes its place.
	root.RemoveChild(original)
	replacement := newTestCheckBox(root, targetMgr, "flag", &state)
	c.Equal(checkenum.On, replacement.State, "the replacement must start from the current state")

	root.mgr.Undo()
	c.Equal(checkenum.Off, state, "undo must restore the prior state")
	c.Equal(checkenum.Off, replacement.State, "undo must be applied to the replacement widget")
	c.Equal(checkenum.On, original.State, "undo must leave the discarded widget alone")

	root.mgr.Redo()
	c.Equal(checkenum.On, state, "redo must reapply the state")
	c.Equal(checkenum.On, replacement.State, "redo must be applied to the replacement widget")
}

// TestTargetUndoFallsBackToRecordingWidget verifies that, when no widget holds the target key, the edit is applied to
// the widget that recorded it.
func TestTargetUndoFallsBackToRecordingWidget(t *testing.T) {
	c := check.New(t)
	root := newPopupUndoRoot()
	targetMgr := NewTargetMgr(root)
	state := checkenum.Off
	cb := newTestCheckBox(root, targetMgr, "flag", &state)

	toggle(cb)
	root.RemoveChild(cb)
	root.mgr.Undo()
	c.Equal(checkenum.Off, state, "undo must restore the prior state")
	c.Equal(checkenum.Off, cb.State, "with no replacement to find, undo must be applied to the recording widget")
}

// TestRecordTargetUndoSkipsNoUndoID verifies that nothing is recorded for unison.NoUndoID, which is what a field
// reports while it has no editing session to attribute changes to, and that recording with no undo manager in reach is
// a no-op rather than a crash.
func TestRecordTargetUndoSkipsNoUndoID(t *testing.T) {
	c := check.New(t)
	root := newPopupUndoRoot()
	panel := unison.NewPanel()
	root.AddChild(panel)
	applied := 0
	apply := func(_ *unison.Panel, _ int) { applied++ }

	recordTargetUndo(panel, nil, "", "Change", unison.NoUndoID, 1, 2, apply)
	c.False(root.mgr.CanUndo(), "NoUndoID must record nothing")

	orphan := unison.NewPanel()
	recordTargetUndo(orphan, nil, "", "Change", unison.NextUndoID(), 1, 2, apply)
	c.Equal(0, applied, "a widget with no undo manager in reach must record nothing")

	recordTargetUndo(panel, nil, "", "Change", unison.NextUndoID(), 1, 2, apply)
	c.True(root.mgr.CanUndo(), "a real undo ID must record an edit")
	root.mgr.Undo()
	c.Equal(1, applied, "undoing the edit must run apply")
}

// TestSetTargetRefKey verifies that the RefKey is only set when both a target manager and a key are supplied.
func TestSetTargetRefKey(t *testing.T) {
	c := check.New(t)
	targetMgr := NewTargetMgr(unison.NewPanel())

	panel := unison.NewPanel()
	setTargetRefKey(panel, nil, "key")
	c.Equal("", panel.RefKey, "no target manager means no RefKey")

	setTargetRefKey(panel, targetMgr, "")
	c.Equal("", panel.RefKey, "no key means no RefKey")

	setTargetRefKey(panel, targetMgr, "key")
	c.Equal("key", panel.RefKey, "both present must set the RefKey")
}
