// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestTraverseVisitsDepthFirst verifies that Traverse visits each node before its children, and a node's descendants
// before its next sibling, which decides the node a walk that stops early finds.
func TestTraverseVisitsDepthFirst(t *testing.T) {
	c := check.New(t)
	newTrait := func(name string, children ...*Trait) *Trait {
		trait := NewTrait(nil, nil, len(children) != 0)
		trait.Name = name
		trait.Children = children
		return trait
	}
	list := []*Trait{newTrait("A", newTrait("A1", newTrait("A1a")), newTrait("A2")), newTrait("B")}
	var names []string
	Traverse(func(trait *Trait) bool {
		names = append(names, trait.Name)
		return false
	}, false, false, list...)
	c.Equal([]string{"A", "A1", "A1a", "A2", "B"}, names)
}

// TestCountNodes verifies that CountNodes descends into containers, honors the enabled-only flag the way Traverse does,
// and applies the optional filter.
func TestCountNodes(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	group := NewTrait(e, nil, true)
	child := NewTrait(e, group, false)
	group.Children = append(group.Children, child)
	disabled := NewTrait(e, nil, true)
	disabled.SetDisabled(true)
	hidden := NewTrait(e, disabled, false)
	disabled.Children = append(disabled.Children, hidden)
	list := []*Trait{group, disabled}
	c.Equal(4, CountNodes(list, false, nil), "a nil filter counts every node, containers included")
	c.Equal(2, CountNodes(list, true, nil), "a disabled node and its descendants are skipped")
	c.Equal(2, CountNodes(list, false, func(t *Trait) bool { return !t.Container() }), "the filter picks the nodes")
	c.Equal(0, CountNodes[*Trait](nil, false, nil))
}
