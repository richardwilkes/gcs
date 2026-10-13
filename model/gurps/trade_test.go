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
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// carry puts the equipment into the container, or into the entity's carried equipment when container is nil.
func carry(e *Entity, container *Equipment, items ...*Equipment) {
	for _, one := range items {
		one.SetDataOwner(e)
		if container == nil {
			one.SetParent(nil)
			e.CarriedEquipment = append(e.CarriedEquipment, one)
		} else {
			one.SetParent(container)
			container.Children = append(container.Children, one)
		}
	}
}

func newContainerItem(name, value string) *Equipment {
	eqp := NewEquipment(nil, nil, true)
	eqp.Name = name
	eqp.BaseValue = value
	return eqp
}

func withQuantity(eqp *Equipment, quantity string) *Equipment {
	eqp.Quantity = fxp.FromStringForced(quantity)
	return eqp
}

func withTags(eqp *Equipment, tags ...string) *Equipment {
	eqp.Tags = tags
	return eqp
}

func TestMoneyItems(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sword := newEquipmentItem("Sword", "500", "3 lb")
	eagle := withTags(newEquipmentItem("Silver Eagle", "20", "0"), "money")
	purse := withTags(newContainerItem("Purse", "1"), "Money")
	penny := withTags(newEquipmentItem("Copper Penny", "1", "0"), "Currency: Money")
	token := withTags(newEquipmentItem("Token", "0", "0"), "Money")
	pouch := newContainerItem("Pouch", "0")
	crown := withTags(newEquipmentItem("Crown", "100", "0"), "MONEY")
	stray := withTags(newEquipmentItem("Stray Coin", "5", "0"), "Money")
	carry(e, nil, sword, eagle, purse)
	carry(e, purse, penny, token, pouch)
	carry(e, pouch, crown)
	e.OtherEquipment = append(e.OtherEquipment, stray)
	money := e.MoneyItems()
	c.Equal(3, len(money))
	for i, expected := range []*Equipment{eagle, penny, crown} {
		c.True(expected == money[i], "expected %v at %d, got %v", expected, i, money[i])
	}
}

func TestRemoveSellMoney(t *testing.T) {
	c := check.New(t)
	var s SheetSettings
	silver := librarySourced(newEquipmentItem("Silver Eagle", "20", "0"))
	copper := librarySourced(newEquipmentItem("Copper Penny", "1", "0"))
	c.True(s.AddSellMoney(silver))
	c.True(s.AddSellMoney(copper))
	c.True(s.RemoveSellMoney(silver.Source))
	c.False(s.RemoveSellMoney(silver.Source))
	c.False(s.RemoveSellMoney(Source{}))
	c.Equal([]Source{copper.Source}, s.SellMoney)
}

// mergeBase returns a non-container with every field MergeMatches compares set, along with modifiers that nest.
func mergeBase() *Equipment {
	eqp := newEquipmentItem("Lantern", "@Price@", "2 lb")
	eqp.PageRef = "B288"
	eqp.PageRefHighlight = "Lantern"
	eqp.LocalNotes = "Lights a room"
	eqp.TechLevel = "3"
	eqp.Tags = []string{"Light"}
	eqp.MaxUses = 10
	eqp.Uses = 4
	eqp.Level = fxp.Two
	eqp.RatedST = fxp.Ten
	eqp.VTTNotes = "lantern"
	eqp.Replacements = map[string]string{"Price": "20"}
	eqp.SwitchedOn = true
	eqp.Source = Source{Library: "Lib", Path: "Gear.eqp", TID: eqp.TID}
	eqp.ThirdParty = map[string]any{"a": 1}
	hooded := NewEquipmentModifier(nil, nil, false)
	hooded.Name = "Hooded"
	hooded.CostAmount = "+5"
	hooded.VTTNotes = "hood"
	group := NewEquipmentModifier(nil, nil, true)
	group.Name = "Fuel"
	oil := NewEquipmentModifier(nil, group, false)
	oil.Name = "Oil"
	oil.Disabled = true
	group.Children = []*EquipmentModifier{oil}
	eqp.SetModifiers([]*EquipmentModifier{hooded, group})
	return eqp
}

func TestMergeMatches(t *testing.T) {
	c := check.New(t)
	base := mergeBase()
	same := base.Clone(LibraryFile{}, nil, nil, Duplicate)
	same.Quantity = fxp.Ten
	same.Source = Source{}
	same.ThirdParty = nil
	c.NotEqual(base.TID, same.TID)
	c.True(base.MergeMatches(same), "only the ID, quantity, source and third-party data differ")
	c.True(same.MergeMatches(base))
	c.False(base.MergeMatches(nil))
	c.False((*Equipment)(nil).MergeMatches(base))

	tests := []struct {
		name   string
		change func(eqp *Equipment)
	}{
		{name: "name", change: func(eqp *Equipment) { eqp.Name = "Lamp" }},
		{name: "page reference", change: func(eqp *Equipment) { eqp.PageRef = "B289" }},
		{name: "page reference highlight", change: func(eqp *Equipment) { eqp.PageRefHighlight = "Lamp" }},
		{name: "local notes", change: func(eqp *Equipment) { eqp.LocalNotes = "" }},
		{name: "tech level", change: func(eqp *Equipment) { eqp.TechLevel = "4" }},
		{name: "legality class", change: func(eqp *Equipment) { eqp.LegalityClass = "3" }},
		{name: "tags", change: func(eqp *Equipment) { eqp.Tags = append(eqp.Tags, "Fragile") }},
		{name: "base value", change: func(eqp *Equipment) { eqp.BaseValue = "20" }},
		{name: "base weight", change: func(eqp *Equipment) { eqp.BaseWeight = "3 lb" }},
		{name: "max uses", change: func(eqp *Equipment) { eqp.MaxUses = 11 }},
		{name: "prereqs", change: func(eqp *Equipment) {
			eqp.Prereq = NewPrereqList()
			eqp.Prereq.Prereqs = append(eqp.Prereq.Prereqs, NewTraitPrereq())
		}},
		{name: "weapons", change: func(eqp *Equipment) { eqp.Weapons = append(eqp.Weapons, NewWeapon(eqp, true)) }},
		{name: "features", change: func(eqp *Equipment) { eqp.Features = append(eqp.Features, NewDRBonus()) }},
		{name: "weight ignored for skills", change: func(eqp *Equipment) { eqp.WeightIgnoredForSkills = true }},
		{name: "equipped", change: func(eqp *Equipment) { eqp.Equipped = false }},
		{name: "level", change: func(eqp *Equipment) { eqp.Level = fxp.Three }},
		{name: "uses", change: func(eqp *Equipment) { eqp.Uses = 3 }},
		{name: "rated strength", change: func(eqp *Equipment) { eqp.RatedST = fxp.Eleven }},
		{name: "vtt notes", change: func(eqp *Equipment) { eqp.VTTNotes = "lamp" }},
		{name: "switched on", change: func(eqp *Equipment) { eqp.SwitchedOn = false }},
		{name: "preconfigured", change: func(eqp *Equipment) { eqp.Preconfigured = true }},
		{name: "pick separately", change: func(eqp *Equipment) { eqp.PickSeparately = true }},
		{name: "replacement value", change: func(eqp *Equipment) { eqp.Replacements["Price"] = "25" }},
		{name: "replacement added", change: func(eqp *Equipment) { eqp.Replacements["Size"] = "Large" }},
		{name: "replacements removed", change: func(eqp *Equipment) { eqp.Replacements = nil }},
		{name: "modifier added", change: func(eqp *Equipment) {
			eqp.AddModifiers(NewEquipmentModifier(nil, nil, false))
		}},
		{name: "modifier removed", change: func(eqp *Equipment) { eqp.Modifiers = eqp.Modifiers[:1] }},
		{name: "modifier order", change: func(eqp *Equipment) { slices.Reverse(eqp.Modifiers) }},
		{name: "modifier sync data", change: func(eqp *Equipment) { eqp.Modifiers[0].CostAmount = "+6" }},
		{name: "modifier disabled", change: func(eqp *Equipment) { eqp.Modifiers[0].Disabled = true }},
		{name: "modifier vtt notes", change: func(eqp *Equipment) { eqp.Modifiers[0].VTTNotes = "" }},
		{name: "modifier replacements", change: func(eqp *Equipment) {
			eqp.Modifiers[0].Replacements = map[string]string{"x": "y"}
		}},
		{name: "nested modifier sync data", change: func(eqp *Equipment) { eqp.Modifiers[1].Children[0].Name = "Gas" }},
		{name: "nested modifier disabled", change: func(eqp *Equipment) { eqp.Modifiers[1].Children[0].Disabled = false }},
		{name: "nested modifier removed", change: func(eqp *Equipment) { eqp.Modifiers[1].Children = nil }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := check.New(t)
			other := base.Clone(LibraryFile{}, nil, nil, Duplicate)
			tc.True(base.MergeMatches(other), "a fresh clone matches")
			tt.change(other)
			tc.False(base.MergeMatches(other))
			tc.False(other.MergeMatches(base))
		})
	}
}

func TestMergeMatchesNeverMatchesContainers(t *testing.T) {
	c := check.New(t)
	bag := newContainerItem("Bag", "5")
	twin := bag.Clone(LibraryFile{}, nil, nil, Duplicate)
	item := newEquipmentItem("Bag", "5", "0")
	c.False(bag.MergeMatches(twin), "identical containers don't match")
	c.False(bag.MergeMatches(item))
	c.False(item.MergeMatches(bag))
	group := NewEquipmentGroup(nil, nil)
	c.False(group.MergeMatches(group.Clone(LibraryFile{}, nil, nil, Duplicate)))
}

// flattenedFields returns the names of the fields of the struct type, with those of embedded structs in their place.
func flattenedFields(typ reflect.Type) []string {
	var names []string
	for field := range typ.Fields() {
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			names = append(names, flattenedFields(field.Type)...)
		} else {
			names = append(names, field.Name)
		}
	}
	return names
}

// TestMergeMatchesCoversEveryField fails when the equipment or equipment modifier data gains or loses a field, so that
// MergeMatches and equipmentModifiersMatch are revisited to compare the new field or deliberately ignore it.
func TestMergeMatchesCoversEveryField(t *testing.T) {
	tests := []struct {
		typ      reflect.Type
		expected []string
	}{
		{
			typ: reflect.TypeFor[EquipmentData](),
			expected: []string{
				"TID", "Source", "ContainerType", "Name", "PageRef", "PageRefHighlight", "LocalNotes", "TechLevel",
				"LegalityClass", "Tags", "BaseValue", "BaseWeight", "MaxUses", "Prereq", "Weapons", "Features",
				"WeightIgnoredForSkills", "TemplatePicker", "VTTNotes", "Replacements", "Modifiers", "RatedST", "Quantity",
				"Level", "Uses", "Equipped", "SwitchedOn", "Preconfigured", "PickSeparately", "ThirdParty", "Children",
				"parent",
			},
		},
		{
			typ: reflect.TypeFor[EquipmentModifierData](),
			expected: []string{
				"TID", "Source", "Name", "PageRef", "PageRefHighlight", "LocalNotes", "Tags", "VTTNotes", "Replacements",
				"CostType", "CostIsPerLevel", "CostIsPerPound", "WeightType", "WeightIsPerLevel", "ShowNotesOnWeapon",
				"TechLevel", "CostAmount", "WeightAmount", "Features", "Disabled", "Choice", "ThirdParty", "Children",
				"parent",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.typ.Name(), func(t *testing.T) {
			check.New(t).Equal(tt.expected, flattenedFields(tt.typ),
				"%s's fields have changed: revisit MergeMatches and equipmentModifiersMatch (model/gurps/trade.go) so that "+
					"each new field is either compared or deliberately ignored, then update this list", tt.typ.Name())
		})
	}
}

func TestTradePriceAndQuantity(t *testing.T) {
	c := check.New(t)
	knife := withQuantity(newEquipmentItem("Small Knife", "30", "1 lb"), "3")
	c.Equal(fxp.FromInteger(30), TradePrice(knife, fxp.Two, fxp.Fifty))
	c.Equal(fxp.FromStringForced("31.2"), TradePrice(knife, fxp.Two, fxp.FromInteger(52)))
	c.Equal(fxp.FromInteger(0), TradePrice(knife, 0, fxp.Hundred))
	c.Equal(fxp.FromInteger(0), TradePrice(knife, fxp.NegOne, fxp.Hundred))
	c.Equal(fxp.FromInteger(0), TradePrice(nil, fxp.One, fxp.Hundred))
	c.Equal(fxp.Three, knife.TradeQuantity())

	// Each backpack holds 2 torches and 3 ropes, so one is worth 5+2*10+3*1.
	backpack := withQuantity(newContainerItem("Backpack", "5"), "2")
	torch := withQuantity(newEquipmentItem("Torch", "10", "1 lb"), "2")
	rope := withQuantity(newEquipmentItem("Rope", "1", "1 lb"), "3")
	torch.SetParent(backpack)
	rope.SetParent(backpack)
	backpack.Children = []*Equipment{torch, rope}
	c.Equal(fxp.FromInteger(28), TradePrice(backpack, fxp.One, fxp.Hundred))
	c.Equal(fxp.FromInteger(56), TradePrice(backpack, fxp.Two, fxp.Hundred))
	c.Equal(fxp.FromInteger(14), TradePrice(backpack, fxp.One, fxp.Fifty))
	c.Equal(fxp.Two, backpack.TradeQuantity())

	group := NewEquipmentGroup(nil, nil)
	gem := withQuantity(newEquipmentItem("Gem", "100", "0"), "2")
	gem.SetParent(group)
	group.Children = []*Equipment{gem}
	c.Equal(fxp.One, group.TradeQuantity())
	c.Equal(fxp.FromInteger(200), TradePrice(group, fxp.One, fxp.Hundred))
}

// denomination returns a money item as SellMoneyItems would: a quantity of one, in no list.
func denomination(name, value string) *Equipment {
	return withTags(newEquipmentItem(name, value, "0"), MoneyTag)
}

func TestCommitSaleOfPartOfAnItemMergesPayouts(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	knives := withQuantity(newEquipmentItem("Small Knife", "30", "1 lb"), "3")
	silver := denomination("Silver Eagle", "20")
	copper := denomination("Copper Penny", "1")
	pennies := withQuantity(copper.Clone(LibraryFile{}, nil, nil, Duplicate), "5")
	carry(e, nil, knives, pennies)
	when := mustTime(t, "2026-10-13T09:40:00Z")

	entry := e.CommitSale([]SaleItem{{Equipment: knives, Quantity: fxp.Two}},
		[]Payout{{Money: silver, Quantity: 1}, {Money: copper, Quantity: 11}}, true, "Knives", when)

	c.Equal(fxp.One, knives.Quantity, "selling part of an item reduces its quantity")
	c.Equal(fxp.FromInteger(16), pennies.Quantity, "payouts merge into the carried item they match")
	c.Equal(3, len(e.CarriedEquipment))
	added := e.CarriedEquipment[2]
	c.True(added != silver, "a copy of the payout is carried")
	c.NotEqual(silver.TID, added.TID)
	c.Equal("Silver Eagle", added.Name)
	c.Equal(fxp.One, added.Quantity)
	c.True(added.Equipped, "a payout is carried equipped")
	c.True(added.DataOwner() == DataOwner(e))

	c.True(tid.IsKindAndValid(entry.ID, kinds.ShoppingRecord), "the entry is given an ID")
	c.Equal(ShoppingSold, entry.Kind)
	c.Equal("Knives", entry.Name)
	c.Equal(when, entry.When)
	c.Equal([]ShoppingRecordItem{{Name: "Small Knife", Quantity: fxp.Two, Value: fxp.FromInteger(30)}}, entry.Items)
	c.Equal([]ShoppingRecordItem{
		{Name: "Silver Eagle", Quantity: fxp.One, Value: fxp.FromInteger(20)},
		{Name: "Copper Penny", Quantity: fxp.FromInteger(11), Value: fxp.One},
	}, entry.Money)
	c.Equal(fxp.FromInteger(31), entry.Exchanged)
	c.Equal([]*ShoppingRecordEntry{entry}, e.ShoppingRecord)
	c.True(e.ShoppingRecord[0] == entry)
}

func TestCommitSaleWithoutMergeCarriesCopies(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	copper := denomination("Copper Penny", "1")
	copper.Equipped = false
	pennies := withQuantity(copper.Clone(LibraryFile{}, nil, nil, Duplicate), "5")
	pennies.Equipped = true
	carry(e, nil, pennies)
	entry := e.CommitSale(nil, []Payout{{Money: copper, Quantity: 4}, {Money: copper, Quantity: 0}, {Quantity: 2}}, false,
		"", mustTime(t, "2026-10-13T09:40:00Z"))
	c.Equal(fxp.FromInteger(5), pennies.Quantity, "nothing merges")
	c.Equal(2, len(e.CarriedEquipment))
	added := e.CarriedEquipment[1]
	c.Equal(fxp.FromInteger(4), added.Quantity)
	c.True(added.Equipped, "a payout is carried equipped")
	c.Nil(entry.Items)
	c.Equal([]ShoppingRecordItem{{Name: "Copper Penny", Quantity: fxp.FromInteger(4), Value: fxp.One}}, entry.Money)
	c.Equal(fxp.FromInteger(4), entry.Exchanged)
}

func TestCommitSaleRemovesWhatIsSoldInFull(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sword := newEquipmentItem("Sword", "500", "3 lb")
	backpack := newContainerItem("Backpack", "60")
	rope := newEquipmentItem("Rope", "5", "1 lb")
	torches := withQuantity(newEquipmentItem("Torch", "3", "1 lb"), "5")
	chest := newContainerItem("Chest", "100")
	bar := newEquipmentItem("Gold Bar", "1000", "1 lb")
	group := NewEquipmentGroup(nil, nil)
	group.Name = "Gems"
	ruby := newEquipmentItem("Ruby", "250", "0")
	shield := newEquipmentItem("Shield", "60", "15 lb")
	carry(e, nil, sword, backpack, chest, group)
	carry(e, backpack, rope, torches)
	carry(e, chest, bar)
	carry(e, group, ruby)
	shield.SetDataOwner(e)
	e.OtherEquipment = append(e.OtherEquipment, shield)

	entry := e.CommitSale([]SaleItem{
		{Equipment: sword, Quantity: fxp.Five},
		{Equipment: rope, Quantity: fxp.One},
		{Equipment: torches, Quantity: fxp.Two},
		{Equipment: chest, Quantity: fxp.One},
		{Equipment: bar, Quantity: fxp.One},
		{Equipment: group, Quantity: fxp.Three},
		{Equipment: backpack, Quantity: 0},
		{Equipment: shield, Quantity: fxp.One},
		{Quantity: fxp.One},
	}, nil, true, "Clear out", mustTime(t, "2026-10-13T09:40:00Z"))

	c.Equal(1, len(e.CarriedEquipment))
	c.True(e.CarriedEquipment[0] == backpack, "only the backpack, which wasn't sold, is left")
	c.Equal(1, len(backpack.Children))
	c.True(backpack.Children[0] == torches)
	c.Nil(rope.Parent(), "a child sold in full leaves its container")
	c.Equal(fxp.Three, torches.Quantity)
	c.Equal(1, len(chest.Children), "a container sold takes its contents with it")
	c.Equal(1, len(e.OtherEquipment))
	c.Equal([]ShoppingRecordItem{
		{Name: "Sword", Quantity: fxp.One, Value: fxp.FromInteger(500)},
		{Name: "Rope", Quantity: fxp.One, Value: fxp.Five},
		{Name: "Torch", Quantity: fxp.Two, Value: fxp.Three},
		{Name: "Chest", Quantity: fxp.One, Value: fxp.FromInteger(1100)},
		{Name: "Gems", Quantity: fxp.One, Value: fxp.FromInteger(250)},
	}, entry.Items, "the contents of a container sold earlier aren't sold again, nor is equipment not carried")
	c.Nil(entry.Money)
	c.Equal(fxp.FromInteger(0), entry.Exchanged)
}

func TestCommitPurchase(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	silver := denomination("Silver Eagle", "20")
	copper := denomination("Copper Penny", "1")
	eagles := withQuantity(silver.Clone(LibraryFile{}, nil, nil, Duplicate), "7")
	pennies := withQuantity(copper.Clone(LibraryFile{}, nil, nil, Duplicate), "11")
	rope := newEquipmentItem("Rope", "5", "1 lb")
	pouch := newContainerItem("Pouch", "1")
	carry(e, nil, rope, pouch)
	carry(e, pouch, eagles, pennies)
	staleEagles := withQuantity(eagles.Clone(LibraryFile{}, nil, nil, Copy), "100")

	library := LibraryFile{Library: "Lib", Path: "Gear.eqp"}
	boughtRope := rope.Clone(library, e, nil, Reference)
	lantern := newEquipmentItem("Lantern", "15", "2 lb")
	lantern.Equipped = false
	boughtLantern := lantern.Clone(library, e, nil, Reference)
	when := mustTime(t, "2026-10-13T09:40:00Z")

	entry := e.CommitPurchase([]PurchaseItem{
		{Equipment: boughtRope, Quantity: fxp.Two},
		{Equipment: boughtLantern, Quantity: 0},
	}, []Payout{
		{Money: staleEagles, Quantity: 2},
		{Money: pennies, Quantity: 11},
	}, []Payout{
		{Money: silver, Quantity: 1},
		{Money: copper, Quantity: 6},
	}, true, "Supplies", when)

	c.Equal(fxp.FromInteger(6), eagles.Quantity, "a payment is found by its ID, and change merges into it")
	c.Equal(fxp.FromInteger(100), staleEagles.Quantity, "the stale copy isn't touched")
	c.Equal(1, len(pouch.Children), "money spent in full is removed")
	c.True(pouch.Children[0] == eagles)
	c.Equal(fxp.Three, rope.Quantity, "a bought item merges into the carried item it matches")
	c.Equal(4, len(e.CarriedEquipment))
	addedLantern := e.CarriedEquipment[2]
	c.Equal("Lantern", addedLantern.Name)
	c.Equal(fxp.One, addedLantern.Quantity, "an item is bought at least once")
	c.True(addedLantern.Equipped, "a bought item is carried equipped")
	c.Equal(Source{LibraryFile: library, TID: lantern.TID}, addedLantern.Source, "a bought item keeps its source")
	c.True(addedLantern.DataOwner() == DataOwner(e))
	addedPennies := e.CarriedEquipment[3]
	c.Equal("Copper Penny", addedPennies.Name)
	c.Equal(fxp.FromInteger(6), addedPennies.Quantity, "change with nothing to merge into is carried")
	c.True(addedPennies.Equipped)

	c.True(tid.IsKindAndValid(entry.ID, kinds.ShoppingRecord))
	c.Equal(ShoppingBought, entry.Kind)
	c.Equal("Supplies", entry.Name)
	c.Equal(when, entry.When)
	c.Equal([]ShoppingRecordItem{
		{Name: "Rope", Quantity: fxp.Two, Value: fxp.Five},
		{Name: "Lantern", Quantity: fxp.One, Value: fxp.FromInteger(15)},
	}, entry.Items)
	c.Equal([]ShoppingRecordItem{
		{Name: "Silver Eagle", Quantity: fxp.FromInteger(-2), Value: fxp.FromInteger(20)},
		{Name: "Copper Penny", Quantity: fxp.FromInteger(-11), Value: fxp.One},
		{Name: "Silver Eagle", Quantity: fxp.One, Value: fxp.FromInteger(20)},
		{Name: "Copper Penny", Quantity: fxp.FromInteger(6), Value: fxp.One},
	}, entry.Money)
	c.Equal(fxp.FromInteger(25), entry.Exchanged, "the money exchanged is what was paid less the change")
	c.Equal([]*ShoppingRecordEntry{entry}, e.ShoppingRecord)
}

func TestChangeFor(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	silver := denomination("Silver Eagle", "20")
	copper := denomination("Copper Penny", "1")
	eagles := withQuantity(silver.Clone(LibraryFile{}, nil, nil, Duplicate), "7")
	pennies := withQuantity(copper.Clone(LibraryFile{}, nil, nil, Duplicate), "11")
	carry(e, nil, eagles, pennies)
	paid := []Payout{{Money: eagles, Quantity: 3}, {Money: pennies, Quantity: 10}}
	over := fxp.FromInteger(25)

	change, total := e.ChangeFor(over, []*Equipment{copper}, paid)
	c.Equal([]Payout{{Money: copper, Quantity: 25}}, change, "sell money that can return the change is used alone")
	c.Equal(over, total)

	change, total = e.ChangeFor(over, []*Equipment{silver}, paid)
	c.Equal(over, total, "the money paid is drawn on when the sell money falls short")
	c.Equal(2, len(change))
	c.True(change[0].Money == silver, "a sell money denomination the money paid matches stands in for it")
	c.Equal(1, change[0].Quantity)
	given := change[1].Money
	c.Equal("Copper Penny", given.Name)
	c.True(given != pennies, "the money paid is copied rather than used as it is")
	c.Equal(fxp.One, given.Quantity)
	c.True(given.DataOwner() == DataOwner(e))
	c.Equal(5, change[1].Quantity)

	change, total = e.ChangeFor(over, nil, paid)
	c.Equal(over, total, "without sell money, the change comes from the money paid")
	c.Equal(2, len(change))
	c.Equal("Silver Eagle", change[0].Money.Name)
	c.Equal(1, change[0].Quantity)
	c.Equal("Copper Penny", change[1].Money.Name)
	c.Equal(5, change[1].Quantity)

	change, total = e.ChangeFor(fxp.FromInteger(15), []*Equipment{silver}, paid[:1])
	c.Nil(change, "nothing that can return the change leaves it unreturned")
	c.Equal(fxp.Int(0), total)

	change, total = e.ChangeFor(fxp.FromInteger(15), nil, []Payout{{Money: pennies}, {Quantity: 2}})
	c.Nil(change, "payments that give nothing are ignored")
	c.Equal(fxp.Int(0), total)
}

func TestCommitPurchaseWithoutMerge(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	pennies := withQuantity(denomination("Copper Penny", "1"), "3")
	rope := newEquipmentItem("Rope", "5", "1 lb")
	carry(e, nil, pennies, rope)
	group := NewEquipmentGroup(e, nil)
	group.Name = "Kit"
	unknown := denomination("Copper Penny", "1")

	entry := e.CommitPurchase([]PurchaseItem{
		{Equipment: rope.Clone(LibraryFile{}, e, nil, Duplicate), Quantity: fxp.Two},
		{Equipment: group, Quantity: fxp.Three},
	}, []Payout{
		{Money: pennies, Quantity: 5},
		{Money: unknown, Quantity: 1},
	}, nil, false, "", mustTime(t, "2026-10-13T09:40:00Z"))

	c.Equal(3, len(e.CarriedEquipment), "spent money is removed, and items are carried even with a match")
	c.True(e.CarriedEquipment[0] == rope)
	c.Equal(fxp.One, rope.Quantity)
	c.Equal("Rope", e.CarriedEquipment[1].Name)
	c.Equal(fxp.Two, e.CarriedEquipment[1].Quantity)
	c.Equal("Kit", e.CarriedEquipment[2].Name)
	c.Equal(fxp.One, e.CarriedEquipment[2].Quantity, "a group is bought just once")
	c.Equal([]ShoppingRecordItem{
		{Name: "Rope", Quantity: fxp.Two, Value: fxp.Five},
		{Name: "Kit", Quantity: fxp.One, Value: 0},
	}, entry.Items)
	c.Equal([]ShoppingRecordItem{{Name: "Copper Penny", Quantity: fxp.FromInteger(-3), Value: fxp.One}}, entry.Money,
		"no more than is carried is spent, and money that isn't carried is skipped")
	c.Equal(fxp.Three, entry.Exchanged)

	refund := e.CommitPurchase(nil, nil, []Payout{{Money: denomination("Copper Penny", "1"), Quantity: 2}}, false, "",
		mustTime(t, "2026-10-13T09:41:00Z"))
	c.Equal(fxp.Two, refund.Exchanged, "the money exchanged is never negative")
	c.Equal([]*ShoppingRecordEntry{refund, entry}, e.ShoppingRecord)
}

// writeMoneyLibrary saves a library equipment file holding a non-container with a modifier, and a container.
func writeMoneyLibrary(t *testing.T) (libFile LibraryFile, silver, purse *Equipment) {
	t.Helper()
	libFile = LibraryFile{Library: "Test Library", Path: "Money/Coins" + EquipmentExt}
	silver = withQuantity(withTags(newEquipmentItem("Silver Eagle", "20", "0.01 lb"), MoneyTag), "50")
	silver.AddModifiers(NewEquipmentModifier(nil, nil, false))
	purse = newContainerItem("Purse", "2")
	p := filepath.Join(useTestLibrary(t, libFile.Library), filepath.FromSlash(libFile.Path))
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := SaveEquipment([]*Equipment{silver, purse}, p); err != nil {
		t.Fatal(err)
	}
	return libFile, silver, purse
}

func TestEquipmentForSource(t *testing.T) {
	c := check.New(t)
	isolateUnownedSrcMatcher(t)
	registerTestFileTypes(t, EquipmentExt)
	libFile, silver, purse := writeMoneyLibrary(t)

	found := EquipmentForSource(Source{LibraryFile: libFile, TID: silver.TID})
	c.NotNil(found)
	c.Equal("Silver Eagle", found.Name)
	c.Equal(silver.TID, found.TID)
	c.True(found == EquipmentForSource(Source{LibraryFile: libFile, TID: silver.TID}), "the library's cached copy")
	c.Nil(EquipmentForSource(Source{LibraryFile: libFile, TID: purse.TID}), "a container isn't returned")
	c.Nil(EquipmentForSource(Source{LibraryFile: libFile, TID: silver.Modifiers[0].TID}), "a modifier isn't returned")
	c.Nil(EquipmentForSource(Source{LibraryFile: libFile, TID: tid.MustNewTID(kinds.Equipment)}))
	c.Nil(EquipmentForSource(Source{Library: "Nowhere", Path: "x.eqp", TID: silver.TID}))
	c.Nil(EquipmentForSource(Source{}))

	done := BatchSourceMatches()
	inBatch := EquipmentForSource(Source{LibraryFile: libFile, TID: silver.TID})
	done()
	c.NotNil(inBatch, "a batch of matches finds it too")
	c.Equal("Silver Eagle", inBatch.Name)

	var sm *SrcMatcher
	c.Nil(sm.libraryEquipment(Source{LibraryFile: libFile, TID: silver.TID}), "a nil matcher finds nothing")
}

func TestSellMoneyItems(t *testing.T) {
	c := check.New(t)
	isolateUnownedSrcMatcher(t)
	registerTestFileTypes(t, EquipmentExt)
	libFile, silver, purse := writeMoneyLibrary(t)
	src := Source{LibraryFile: libFile, TID: silver.TID}

	e := NewEntity()
	e.SheetSettings.SellMoney = []Source{
		{LibraryFile: libFile, TID: tid.MustNewTID(kinds.Equipment)},
		src,
		{LibraryFile: libFile, TID: purse.TID},
		{LibraryFile: libFile, TID: silver.Modifiers[0].TID},
	}
	items := e.SellMoneyItems()
	c.Equal(1, len(items), "missing items, containers and modifiers are skipped")
	money := items[0]
	c.Equal("Silver Eagle", money.Name)
	c.Equal(src, money.Source)
	c.NotEqual(silver.TID, money.TID)
	c.Equal(fxp.One, money.Quantity)
	c.Equal(fxp.FromInteger(20), money.AdjustedValue())
	c.True(money.DataOwner() == DataOwner(e))
	c.Nil(money.Parent())
	c.Equal(0, len(e.CarriedEquipment), "the copies aren't put in any list")
	c.Equal(1, len(money.Modifiers))
	c.Equal(fxp.FromInteger(50), e.SourceMatcher().libraryEquipment(src).Quantity, "the library's copy is untouched")
}
