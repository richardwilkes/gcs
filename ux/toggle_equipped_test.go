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
	"github.com/richardwilkes/toolbox/v2/tid"
)

// TestToggleEquippedSkipsChoiceContainers verifies that a choice can't be unequipped from the list, since its editor
// offers no way to equip it again.
func TestToggleEquippedSkipsChoiceContainers(t *testing.T) {
	c := check.New(t)
	choice := gurps.NewEquipmentChoiceContainer(nil, nil)
	item := gurps.NewEquipment(nil, nil, false)
	template := newTestTemplateWithEquipment(choice, item)
	table := template.Equipment.Table

	table.SetSelectionMap(map[tid.TID]bool{choice.ID(): true})
	c.False(canToggleEquipped(table), "a choice alone must not offer to be unequipped")

	table.SetSelectionMap(map[tid.TID]bool{choice.ID(): true, item.ID(): true})
	toggleEquipped(template, table)
	c.True(choice.Equipped, "the choice must be left equipped")
	c.False(item.Equipped, "the item must be unequipped")
}
