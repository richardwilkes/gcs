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
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestEquipmentContextMenuLeavesOutTheListName verifies that an equipment list's context menu offers to add to that list
// without naming it, naming what it holds only in its first item, while still invoking the commands for that list, and
// that only the carried list offers a choice.
func TestEquipmentContextMenuLeavesOutTheListName(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	titles := func(carried bool) map[string]int {
		provider, ok := NewEquipmentProvider(gurps.NewTemplate(), carried, true).(*equipmentProvider)
		c.True(ok)
		m := make(map[string]int)
		for _, one := range provider.ContextMenuItems() {
			m[one.Title] = one.ID
		}
		return m
	}
	carried := titles(true)
	c.Equal(NewCarriedEquipmentItemID, carried["New Equipment"])
	c.Equal(NewCarriedEquipmentContainerItemID, carried["New Container"])
	c.Equal(NewCarriedEquipmentGroupItemID, carried["New Group"])
	c.Equal(NewEquipmentChoiceContainerItemID, carried["New Choice"])

	other := titles(false)
	c.Equal(NewOtherEquipmentItemID, other["New Equipment"])
	c.Equal(NewOtherEquipmentContainerItemID, other["New Container"])
	c.Equal(NewOtherEquipmentGroupItemID, other["New Group"])
	_, hasChoice := other["New Choice"]
	c.False(hasChoice, "the other equipment list never holds choices")
}

// TestEquipmentPageHeaderSyncRefreshesTotalsTooltip verifies that syncing a page's equipment list refreshes the
// header's tooltip along with the totals in its title. The tooltip is where the exact totals are offered once the
// display formats round them, and the formats can change after the headers were built -- which is exactly what
// changing them in the sheet settings does -- so refreshing only the title used to leave the tooltip missing, or stale.
func TestEquipmentPageHeaderSyncRefreshesTotalsTooltip(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	e := gurps.NewEquipment(entity, nil, false)
	e.Name = "Heavy Thing"
	e.BaseWeight = exactWeightText
	entity.CarriedEquipment = []*gurps.Equipment{e}
	provider := NewEquipmentProvider(entity, true, true)
	table := newProviderTable(provider)
	ids := provider.ColumnIDs()
	table.Columns = make([]unison.ColumnInfo, len(ids))
	for i, id := range ids {
		table.Columns[i].ID = id
	}
	headers := provider.Headers()
	header, ok := headers[slices.Index(ids, gurps.EquipmentDescriptionColumn)].(*PageTableColumnHeader[*gurps.Equipment])
	c.True(ok, "the page's description column must have a page header")
	c.Equal("Carried Equipment (7.5127 lb; $0)", header.Text.String())
	c.Equal("", header.TooltipText(), "with no display preference, there are no exact totals to offer")
	c.Nil(header.Tooltip)

	entity.SheetSettings.EquipmentWeightFormat = fxp.NumberFormat{Places: fxp.TwoPlaces}
	provider.SyncHeader(headers)
	c.Equal("Carried Equipment (7.51 lb; $0)", header.Text.String())
	c.Equal("7.5127 lb; $0", header.TooltipText(), "syncing must offer the exact totals once the formats round them")
	c.NotNil(header.Tooltip)

	entity.SheetSettings.EquipmentWeightFormat = fxp.NumberFormat{}
	provider.SyncHeader(headers)
	c.Equal("Carried Equipment (7.5127 lb; $0)", header.Text.String())
	c.Equal("", header.TooltipText(), "syncing must remove the tooltip once there is nothing to add")
	c.Nil(header.Tooltip)
}
