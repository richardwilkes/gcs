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

// TestEquipmentDragMovesWithinASheet verifies that equipment dragged within a sheet, from one of its equipment lists to
// the other included, is moved there, while equipment dragged to another sheet is copied.
func TestEquipmentDragMovesWithinASheet(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	other := newTestSheetForTemplate(t)
	carried := sheet.CarriedEquipment.Table
	provider, ok := carried.ClientData()[TableProviderClientKey].(TableProvider[*gurps.Equipment])
	c.True(ok, "the table must have its provider recorded")
	c.True(provider.DropShouldMoveData(carried, carried), "a drag within a list must move")
	c.True(provider.DropShouldMoveData(carried, sheet.OtherEquipment.Table), "a drag between a sheet's lists must move")
	c.False(provider.DropShouldMoveData(carried, other.CarriedEquipment.Table), "a drag between sheets must copy")
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

func TestEquipmentProviderUsesTheCarriedOrOtherList(t *testing.T) {
	c := check.New(t)
	lists := &listsForTest{
		carried: []*gurps.Equipment{gurps.NewEquipment(nil, nil, false), gurps.NewEquipment(nil, nil, false)},
		other:   []*gurps.Equipment{gurps.NewEquipment(nil, nil, false)},
	}
	carried := NewEquipmentProvider(lists, true, true)
	other := NewEquipmentProvider(lists, false, true)
	c.Equal(lists.carried, carried.RootData())
	c.Equal(lists.other, other.RootData())
	c.Equal(gurps.BlockEquipmentKey, carried.RefKey())
	c.Equal(gurps.BlockOtherEquipmentKey, other.RefKey())
	other.SetRootData(nil)
	c.Equal(0, len(lists.other), "the other provider must write to the other list")
	c.Equal(2, len(lists.carried), "the other provider must leave the carried list alone")
	carried.SetRootData(nil)
	c.Equal(0, len(lists.carried), "the carried provider must write to the carried list")
}

// TestEquipmentProviderFollowsTLAndLCColumnSettings verifies that the equipment list's tech level and legality class
// columns follow the sheet's settings on a page and are always present off one.
func TestEquipmentProviderFollowsTLAndLCColumnSettings(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	onPage := NewEquipmentProvider(entity, true, true)
	offPage := NewEquipmentProvider(entity, true, false)
	for _, hide := range []bool{false, true} {
		entity.SheetSettings.HideTLColumn = hide
		entity.SheetSettings.HideLCColumn = !hide
		c.Equal(!hide, slices.Contains(onPage.ColumnIDs(), gurps.EquipmentTLColumn),
			"on a page the TL column must follow the sheet's setting (hide=%v)", hide)
		c.Equal(hide, slices.Contains(onPage.ColumnIDs(), gurps.EquipmentLCColumn),
			"on a page the LC column must follow the sheet's setting (hide=%v)", !hide)
		c.True(slices.Contains(offPage.ColumnIDs(), gurps.EquipmentTLColumn), "off a page the TL column must be present")
		c.True(slices.Contains(offPage.ColumnIDs(), gurps.EquipmentLCColumn), "off a page the LC column must be present")
	}
}
