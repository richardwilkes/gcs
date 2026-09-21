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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestShouldClearPreconfiguredFlag verifies that only the instance documents clear the flag. It used to be cleared
// everywhere but a template, which meant a drop into a library list threw it away: a library holds items that have
// been resolved once precisely so that they can be dropped onto a sheet without being asked about again, and that is
// the pattern the old test destroyed.
func TestShouldClearPreconfiguredFlag(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	loot := newTestLootSheet(t)
	template := newTestTemplateWithBodyType("")

	c.False(allowPreconfiguredFlag(sheet.Traits.Table), "a sheet must clear the flag")
	c.False(allowPreconfiguredFlag(loot.Equipment.Table), "a loot sheet must clear the flag")
	c.True(allowPreconfiguredFlag(template.Traits.Table), "a template must keep the flag")
	c.True(allowPreconfiguredFlag(newLibraryStyleTraitsTable()), "a library list must keep the flag")
	c.True(allowPreconfiguredFlag(nil), "nothing at all must not be cleared")
}

// TestClearPreconfiguredFlagLeavesALibraryListAlone verifies the same rule through the clearing function itself, since
// that is what the drop path calls. The flag on an item dropped into a library list is authored data.
func TestClearPreconfiguredFlagLeavesALibraryListAlone(t *testing.T) {
	c := check.New(t)
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Resolved Variant"
	trait.SetPreconfigured(true)
	table := newLibraryStyleTraitsTable(trait)

	c.False(maybeClearPreconfiguredFlag(table, table.RootRows()), "nothing must be reported as changed")
	c.True(trait.IsPreconfigured(), "an item dropped into a library list must keep its flag")
}

// TestClearPreconfiguredFlagReachesEverythingBeneathTheRows verifies that the clear descends. A selection holds only
// the shallowest rows of each branch, and only those that are showing, so a closed container arrives as a single row
// standing in for everything inside it -- rows the clear could not see at all before it traversed. A container that is
// itself preconfigured is cleared and descended into rather than treated as a settled subtree: its flag speaks for its
// own selections and substitutions, not its children's.
func TestClearPreconfiguredFlagReachesEverythingBeneathTheRows(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	newTrait := func(name string, parent *gurps.Trait, container bool) *gurps.Trait {
		trait := gurps.NewTrait(entity, parent, container)
		trait.Name = name
		trait.SetPreconfigured(true)
		return trait
	}
	outer := newTrait("Outer", nil, true)
	middle := newTrait("Middle", outer, true)
	leaf := newTrait("Leaf", middle, false)
	middle.Children = []*gurps.Trait{leaf}
	outer.Children = []*gurps.Trait{middle}
	outer.SetOpen(false)
	middle.SetOpen(false)
	entity.Traits = []*gurps.Trait{outer}
	sheet.Rebuild(true)

	table := sheet.Traits.Table
	selectTraits(table, outer)
	rows := table.SelectedRows(true)
	c.Equal(1, len(rows), "the closed container must stand alone in the selection, with nothing inside it")

	c.True(maybeClearPreconfiguredFlag(table, nil), "clearing must report that something changed")
	c.False(outer.IsPreconfigured(), "the container itself must be cleared")
	c.False(middle.IsPreconfigured(), "a container nested one deep must be cleared")
	c.False(leaf.IsPreconfigured(), "a leaf nested two deep must be cleared")

	c.False(maybeClearPreconfiguredFlag(table, nil), "clearing again must report no change")
}
