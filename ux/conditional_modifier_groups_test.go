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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestSheetGroupedReactionsDiscloseAndPersist verifies that reactions filed under a group appear on the sheet as a
// container row holding them, that closing the container hides them and records the state in the global settings, and
// that the state survives a rebuild, which regenerates the rows from scratch.
func TestSheetGroupedReactionsDiscloseAndPersist(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Warrior"
	for _, one := range []struct {
		situation string
		group     string
	}{
		{situation: "from foes", group: "Combat"},
		{situation: "from allies", group: "Combat"},
		{situation: "from everyone", group: ""},
	} {
		bonus := gurps.NewReactionBonus()
		bonus.Situation = one.situation
		bonus.Group = one.group
		bonus.Amount = fxp.One
		trait.Features = append(trait.Features, bonus)
	}
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)

	c.True(listAttachedToSheet(sheet, sheet.Reactions), "the Reactions list must be on the page")
	table := sheet.Reactions.Table
	c.Equal(2, table.RootRowCount(), "a group container and an ungrouped entry")
	c.Equal(4, sheet.Reactions.RowCount(), "with the group open, its members are shown beneath it")
	var group *Node[*gurps.ConditionalModifier]
	for _, row := range table.RootRows() {
		if row.CanHaveChildren() {
			group = row
		}
	}
	c.NotNil(group, "the group row can have children")
	if group == nil {
		return
	}
	closedKey := "n:" + string(group.Data().ID())
	t.Cleanup(func() { gurps.SetClosedState(closedKey, false) })
	c.True(group.IsOpen(), "the group starts out open")
	c.False(gurps.IsClosed(closedKey))

	open, exists := sheet.Reactions.FirstDisclosureState()
	c.True(exists, "the list reports having something to disclose")
	c.True(open)
	sheet.Reactions.SetDisclosureState(false)
	c.True(gurps.IsClosed(closedKey), "closing the group records it in the global settings by ID")
	c.Equal(2, sheet.Reactions.RowCount(), "with the group closed, its members are hidden")

	sheet.Rebuild(true)
	c.True(listAttachedToSheet(sheet, sheet.Reactions), "the Reactions list is still on the page")
	c.Equal(2, sheet.Reactions.Table.RootRowCount(), "the rows are rebuilt")
	c.Equal(2, sheet.Reactions.RowCount(), "and the group stays closed, since its ID is derived from its name")
	open, exists = sheet.Reactions.FirstDisclosureState()
	c.True(exists)
	c.False(open)

	sheet.Reactions.SetDisclosureState(true)
	c.False(gurps.IsClosed(closedKey), "opening the group clears the recorded state")
	c.Equal(4, sheet.Reactions.RowCount())
}
