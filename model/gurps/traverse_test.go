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

// TestCountNodes verifies that countNodes descends into containers, honors the enabled-only flag the way Traverse does,
// and applies the optional filter.
func TestCountNodes(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	group := NewTrait(e, nil, true)
	child := NewTrait(e, group, false)
	group.Children = append(group.Children, child)
	disabled := NewTrait(e, nil, true)
	disabled.Disabled = true
	hidden := NewTrait(e, disabled, false)
	disabled.Children = append(disabled.Children, hidden)
	list := []*Trait{group, disabled}
	c.Equal(4, countNodes(list, false, nil), "a nil filter counts every node, containers included")
	c.Equal(2, countNodes(list, true, nil), "a disabled node and its descendants are skipped")
	c.Equal(2, countNodes(list, false, func(t *Trait) bool { return !t.Container() }), "the filter picks the nodes")
	c.Equal(0, countNodes[*Trait](nil, false, nil))
}
