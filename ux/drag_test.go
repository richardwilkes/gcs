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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
	"github.com/richardwilkes/unison/enums/mod"
)

// rerouteTarget is a panel that records the drag callbacks rerouted to it.
type rerouteTarget struct {
	*unison.Panel
	updates []geom.Point
	exits   int
	drops   []drag.Info
	op      drag.Op
	handled bool
}

func newRerouteTarget(op drag.Op, handled bool) *rerouteTarget {
	rt := &rerouteTarget{Panel: unison.NewPanel(), op: op, handled: handled}
	rt.DragUpdatedCallback = func(_ drag.Info, where geom.Point, _ mod.Modifiers) drag.Op {
		rt.updates = append(rt.updates, where)
		return rt.op
	}
	rt.DragExitedCallback = func() { rt.exits++ }
	rt.DropCallback = func(di drag.Info, _ geom.Point, _ mod.Modifiers) bool {
		rt.drops = append(rt.drops, di)
		return rt.handled
	}
	return rt
}

// TestInstallDropReroutingForwardsToTheResolvedPanel verifies that the root panel only accepts the supplied drag data
// types, that each drag update resolves the payload's data type to a target panel afresh and forwards the update (with
// the drop point pushed below any rows) there, and that exit and drop are forwarded to the panel the last update
// resolved and then forgotten, so that a stale target never receives a later event.
func TestInstallDropReroutingForwardsToTheResolvedPanel(t *testing.T) {
	c := check.New(t)
	root := unison.NewPanel()
	skills := newRerouteTarget(drag.Copy, false)
	traits := newRerouteTarget(drag.Move, true)
	var resolved []*uti.DataType
	installDropRerouting(root, []*uti.DataType{skillDragKey, traitDragKey}, func(key *uti.DataType) *unison.Panel {
		resolved = append(resolved, key)
		switch key {
		case skillDragKey:
			return skills.Panel
		case traitDragKey:
			return traits.Panel
		default:
			return nil
		}
	})
	skillDrag := &fakeDragInfo{types: []string{skillDragKey.UTI}}
	traitDrag := &fakeDragInfo{types: []string{traitDragKey.UTI}}
	noteDrag := &fakeDragInfo{types: []string{noteDragKey.UTI}}
	where := geom.Point{X: 5, Y: 7}

	c.True(root.CanAcceptDropCallback(skillDrag), "a listed data type is accepted")
	c.True(root.CanAcceptDropCallback(traitDrag), "every listed data type is accepted")
	c.False(root.CanAcceptDropCallback(noteDrag), "an unlisted data type is declined")

	// Entering forwards to the resolved panel and reports its drag op.
	c.Equal(drag.Copy, root.DragEnteredCallback(skillDrag, where, mod.None), "enter reports the target's op")
	c.Equal([]*uti.DataType{skillDragKey}, resolved, "enter resolves the payload's data type")
	c.Equal(1, len(skills.updates), "enter is forwarded to the skills list")
	c.Equal(0, len(traits.updates), "enter is not forwarded to the traits list")
	c.True(skills.updates[0].Y > 1000000, "the forwarded point is far below any row")
	c.Equal(float32(0), skills.updates[0].X, "the forwarded point ignores the root's drop point")

	// Each update re-resolves the target, so a payload of a different type moves the reroute to another list.
	c.Equal(drag.Move, root.DragUpdatedCallback(traitDrag, where, mod.None), "update reports the new target's op")
	c.Equal([]*uti.DataType{skillDragKey, traitDragKey}, resolved, "update resolves afresh")
	c.Equal(1, len(traits.updates), "update is forwarded to the traits list")
	c.Equal(1, len(skills.updates), "update is no longer forwarded to the skills list")

	// The drop goes to the panel the last update resolved, and only there.
	c.True(root.DropCallback(traitDrag, where, mod.None), "drop reports whether the target handled it")
	c.Equal([]drag.Info{traitDrag}, traits.drops, "drop is forwarded to the traits list")
	c.Equal(0, len(skills.drops), "drop is not forwarded to the skills list")
	c.Equal(0, skills.exits+traits.exits, "a drop does not exit the target")

	// Once dropped, the target is forgotten: another drop without an update in between has nowhere to go.
	c.False(root.DropCallback(traitDrag, where, mod.None), "a drop with no target is not handled")
	c.Equal(1, len(traits.drops), "a drop with no target is not forwarded")

	// Exiting is forwarded to the current target exactly once.
	c.Equal(drag.Copy, root.DragEnteredCallback(skillDrag, where, mod.None), "re-enter")
	root.DragExitedCallback()
	c.Equal(1, skills.exits, "exit is forwarded to the current target")
	root.DragExitedCallback()
	c.Equal(1, skills.exits, "a second exit has no target to forward to")
	c.Equal(0, traits.exits, "exit is not forwarded to a stale target")
	c.False(root.DropCallback(skillDrag, where, mod.None), "a drop after an exit has no target")
	c.Equal(0, len(skills.drops), "a drop after an exit is not forwarded")

	// The target's answer to a drop is what the root reports.
	c.Equal(drag.Copy, root.DragEnteredCallback(skillDrag, where, mod.None), "re-enter for the drop")
	c.False(root.DropCallback(skillDrag, where, mod.None), "the skills list declines the drop")
	c.Equal(1, len(skills.drops), "the declined drop was still forwarded to the skills list")
}

// TestInstallDropReroutingWithoutATarget verifies that a payload the resolver has no panel for is declined without
// forwarding anything, that a target lacking exit or drop callbacks is tolerated, and that a data type the root does
// not list is never resolved even if the resolver would have a panel for it.
func TestInstallDropReroutingWithoutATarget(t *testing.T) {
	c := check.New(t)
	root := unison.NewPanel()
	bare := unison.NewPanel()
	bare.DragUpdatedCallback = func(_ drag.Info, _ geom.Point, _ mod.Modifiers) drag.Op { return drag.Copy }
	var resolved []*uti.DataType
	installDropRerouting(root, []*uti.DataType{skillDragKey, traitDragKey}, func(key *uti.DataType) *unison.Panel {
		resolved = append(resolved, key)
		if key == skillDragKey {
			return bare
		}
		return nil
	})
	where := geom.Point{X: 5, Y: 7}

	// No panel for the payload: the drag is declined and nothing is left to receive the exit or drop.
	traitDrag := &fakeDragInfo{types: []string{traitDragKey.UTI}}
	c.Equal(drag.None, root.DragEnteredCallback(traitDrag, where, mod.None), "unresolved payload is declined")
	c.Equal([]*uti.DataType{traitDragKey}, resolved, "the resolver was consulted")
	root.DragExitedCallback()
	c.False(root.DropCallback(traitDrag, where, mod.None), "unresolved payload is not dropped")

	// A payload the root does not list is never resolved, even when the resolver could answer for it.
	resolved = nil
	noteDrag := &fakeDragInfo{types: []string{noteDragKey.UTI}}
	c.Equal(drag.None, root.DragEnteredCallback(noteDrag, where, mod.None), "unlisted payload is declined")
	c.Equal(0, len(resolved), "unlisted payload is not resolved")

	// Only the first listed data type present in the payload is considered.
	both := &fakeDragInfo{types: []string{traitDragKey.UTI, skillDragKey.UTI}}
	c.Equal(drag.Copy, root.DragEnteredCallback(both, where, mod.None), "the first listed type wins")
	c.Equal([]*uti.DataType{skillDragKey}, resolved, "only the first listed type is resolved")

	// A target without exit and drop callbacks must not be called through nil.
	skillDrag := &fakeDragInfo{types: []string{skillDragKey.UTI}}
	c.Equal(drag.Copy, root.DragEnteredCallback(skillDrag, where, mod.None), "enter the bare target")
	c.False(root.DropCallback(skillDrag, where, mod.None), "a target without a drop callback declines")
	c.Equal(drag.Copy, root.DragEnteredCallback(skillDrag, where, mod.None), "re-enter the bare target")
	root.DragExitedCallback()
	c.False(root.DropCallback(skillDrag, where, mod.None), "exit cleared the bare target")
}
