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
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/eqcontainer"
	"github.com/richardwilkes/gcs/v5/model/jio"
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
	dull := withTags(newEquipmentItem("Dull Coin", "5", "0"), "Money")
	dull.SetDisabled(true)
	sack := newContainerItem("Sack", "0")
	sack.SetDisabled(true)
	hidden := withTags(newEquipmentItem("Hidden Coin", "5", "0"), "Money")
	carry(e, nil, sword, eagle, purse, dull, sack)
	carry(e, purse, penny, token, pouch)
	carry(e, pouch, crown)
	carry(e, sack, hidden)
	e.OtherEquipment = append(e.OtherEquipment, stray)
	money := e.MoneyItems()
	c.Equal(3, len(money), "money that is disabled, or in a disabled container, isn't to be spent")
	for i, expected := range []*Equipment{eagle, penny, crown} {
		c.True(expected == money[i], "expected %v at %d, got %v", expected, i, money[i])
	}
}

func TestRemoveSellMoney(t *testing.T) {
	c := check.New(t)
	var s SheetSettings
	c.True(s.AddSellMoney(denomination("Silver Eagle", "20"), LibraryFile{}))
	c.True(s.AddSellMoney(denomination("Copper Penny", "1"), LibraryFile{}))
	silver, copper := s.SellMoney[0], s.SellMoney[1]
	c.True(s.RemoveSellMoney(silver.TID))
	c.False(s.RemoveSellMoney(silver.TID))
	c.False(s.RemoveSellMoney(""))
	c.Equal(SellMoneyList{copper}, s.SellMoney)
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
	eqp.Weapons = []*Weapon{NewWeapon(eqp, true)}
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
	same.Equipped = !base.Equipped
	same.ThirdParty = nil
	c.NotEqual(base.TID, same.TID)
	c.True(base.MergeMatches(same), "only the ID, quantity, source, equipped state and third-party data differ")
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
		{name: "weapon data", change: func(eqp *Equipment) { eqp.Weapons[0].Usage = "Thrust" }},
		{name: "weapon hidden", change: func(eqp *Equipment) { eqp.Weapons[0].Hide = true }},
		{name: "features", change: func(eqp *Equipment) { eqp.Features = append(eqp.Features, NewDRBonus()) }},
		{name: "weight ignored for skills", change: func(eqp *Equipment) { eqp.WeightIgnoredForSkills = true }},
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

// TestMergeMatchesCoversEveryField fails when the equipment, equipment modifier or weapon data gains or loses a field,
// so that MergeMatches, equipmentModifiersMatch and weaponsMatch are revisited to compare the new field or deliberately
// ignore it.
func TestMergeMatchesCoversEveryField(t *testing.T) {
	tests := []struct {
		typ      reflect.Type
		expected []string
	}{
		{
			typ: reflect.TypeFor[WeaponData](),
			expected: []string{
				"TID", "SubVersion", "Damage", "Strength", "Usage", "UsageNotes", "Reach", "Parry", "Block", "Accuracy",
				"Range", "RateOfFire", "Shots", "Bulk", "Recoil", "Defaults", "Hide",
			},
		},
		{
			typ: reflect.TypeFor[EquipmentData](),
			expected: []string{
				"TID", "Source", "ContainerType", "Name", "PageRef", "PageRefHighlight", "LocalNotes", "TechLevel",
				"LegalityClass", "Tags", "BaseValue", "BaseWeight", "MaxUses", "Prereq", "Weapons", "Features",
				"WeightIgnoredForSkills", "TemplatePicker", "VTTNotes", "Replacements", "Modifiers", "RatedST", "Quantity",
				"Level", "Uses", "Equipped", "Disabled", "SwitchedOn", "Preconfigured", "PickSeparately", "ThirdParty",
				"Children", "parent",
			},
		},
		{
			typ: reflect.TypeFor[EquipmentModifierData](),
			expected: []string{
				"TID", "Source", "Name", "PageRef", "PageRefHighlight", "LocalNotes", "Tags", "VTTNotes", "Replacements",
				"CostType", "CostIsPerLevel", "CostIsPerPound", "WeightType", "WeightIsPerLevel", "ShowNotesOnWeapon",
				"TechLevel", "CostAmount", "WeightAmount", "Features", "HideNotes", "Disabled", "Choice", "ThirdParty",
				"Children", "parent",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.typ.Name(), func(t *testing.T) {
			check.New(t).Equal(tt.expected, flattenedFields(tt.typ),
				"%s's fields have changed: revisit MergeMatches, equipmentModifiersMatch and weaponsMatch "+
					"(model/gurps/trade.go) so that each new field is either compared or deliberately ignored, then update "+
					"this list", tt.typ.Name())
		})
	}
}

func TestTradePriceAndQuantity(t *testing.T) {
	c := check.New(t)
	knife := withQuantity(newEquipmentItem("Small Knife", "30", "1 lb"), "3")
	c.Equal(fxp.FromInteger(30), TradePrice(knife, fxp.Two, 50))
	c.Equal(fxp.FromStringForced("31.2"), TradePrice(knife, fxp.Two, 52))
	c.Equal(fxp.FromInteger(0), TradePrice(knife, 0, 100))
	c.Equal(fxp.FromInteger(0), TradePrice(knife, fxp.NegOne, 100))
	c.Equal(fxp.FromInteger(0), TradePrice(nil, fxp.One, 100))
	c.Equal(fxp.Three, knife.TradeQuantity())

	// Each backpack holds 2 torches and 3 ropes, so one is worth 5+2*10+3*1.
	backpack := withQuantity(newContainerItem("Backpack", "5"), "2")
	torch := withQuantity(newEquipmentItem("Torch", "10", "1 lb"), "2")
	rope := withQuantity(newEquipmentItem("Rope", "1", "1 lb"), "3")
	torch.SetParent(backpack)
	rope.SetParent(backpack)
	backpack.Children = []*Equipment{torch, rope}
	c.Equal(fxp.FromInteger(28), TradePrice(backpack, fxp.One, 100))
	c.Equal(fxp.FromInteger(56), TradePrice(backpack, fxp.Two, 100))
	c.Equal(fxp.FromInteger(14), TradePrice(backpack, fxp.One, 50))
	c.Equal(fxp.Two, backpack.TradeQuantity())

	group := NewEquipmentGroup(nil, nil)
	gem := withQuantity(newEquipmentItem("Gem", "100", "0"), "2")
	gem.SetParent(group)
	group.Children = []*Equipment{gem}
	c.Equal(fxp.One, group.TradeQuantity())
	c.Equal(fxp.FromInteger(200), TradePrice(group, fxp.One, 100))
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
	c.True(rope.Equipped, "merging into an equipped item leaves it equipped")
	c.Equal(4, len(e.CarriedEquipment))
	addedLantern := e.CarriedEquipment[2]
	c.Equal("Lantern", addedLantern.Name)
	c.Equal(fxp.One, addedLantern.Quantity, "an item is bought at least once")
	c.False(addedLantern.Equipped, "a bought item is carried unequipped")
	c.Equal(Source{LibraryFile: library, TID: lantern.TID}, addedLantern.Source, "a bought item keeps its source")
	c.True(addedLantern.DataOwner() == DataOwner(e))
	addedPennies := e.CarriedEquipment[3]
	c.Equal("Copper Penny", addedPennies.Name)
	c.Equal(fxp.FromInteger(6), addedPennies.Quantity, "change with nothing to merge into is carried")
	c.True(addedPennies.Equipped, "change is carried equipped")

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

// TestCommitPurchaseMergesWhateverTheEquippedState checks that what is bought, and the change, merges into a carried
// match whether or not either is equipped, and that the match keeps its equipped state.
func TestCommitPurchaseMergesWhateverTheEquippedState(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	bolts := withQuantity(newEquipmentItem("Crossbow Bolt", "2", "0.06 lb"), "20")
	bolts.Equipped = false
	arrows := withQuantity(newEquipmentItem("Arrow", "2", "0.1 lb"), "10")
	copper := denomination("Copper Penny", "1")
	pennies := withQuantity(copper.Clone(LibraryFile{}, nil, nil, Duplicate), "5")
	pennies.Equipped = false
	carry(e, nil, bolts, arrows, pennies)
	boughtBolts := bolts.Clone(LibraryFile{}, e, nil, Reference)
	boughtBolts.Equipped = true
	boughtArrows := arrows.Clone(LibraryFile{}, e, nil, Reference)
	boughtArrows.Equipped = false

	e.CommitPurchase([]PurchaseItem{
		{Equipment: boughtBolts, Quantity: fxp.FromInteger(20)},
		{Equipment: boughtArrows, Quantity: fxp.Ten},
	}, nil, []Payout{{Money: copper, Quantity: 3}}, true, "Ammunition", mustTime(t, "2026-10-15T12:00:00Z"))

	c.Equal([]string{"Crossbow Bolt ×40", "Arrow ×20", "Copper Penny ×8"}, describeCarried(e))
	c.False(bolts.Equipped, "an unequipped match stays unequipped")
	c.True(arrows.Equipped, "an equipped match stays equipped")
	c.False(pennies.Equipped, "unequipped money stays unequipped")
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

// TestSellMoneyItems checks that each sell money entry is handed out as a copy owned by the entity, with a quantity of
// one and its modifiers, keeping the entry's Source and in no list, and that the entry and the item it was copied from
// are left as they were. Entries that can't be money, which only a direct assignment can hold, are skipped.
func TestSellMoneyItems(t *testing.T) {
	c := check.New(t)
	libFile := LibraryFile{Library: "Test Library", Path: "Money/Coins" + EquipmentExt}
	silver := withQuantity(withTags(newEquipmentItem("Silver Eagle", "20", "0.01 lb"), MoneyTag), "50")
	silver.AddModifiers(NewEquipmentModifier(nil, nil, false))
	e := NewEntity()
	c.True(e.SheetSettings.AddSellMoney(silver, libFile))
	entry := e.SheetSettings.SellMoney[0]
	src := Source{LibraryFile: libFile, TID: silver.TID}
	c.Equal(src, entry.Source, "the entry points back at the item it was copied from")
	e.SheetSettings.SellMoney = append(SellMoneyList{
		nil,
		withTags(newContainerItem("Purse", "2"), MoneyTag),
		newEquipmentItem("Gold Crown", "100", "0"),
	}, e.SheetSettings.SellMoney...)
	items := e.SellMoneyItems()
	c.Equal(1, len(items), "nil entries, containers and items not tagged as money are skipped")
	money := items[0]
	c.Equal("Silver Eagle", money.Name)
	c.Equal(src, money.Source, "the copy keeps the entry's Source")
	c.NotEqual(entry.TID, money.TID)
	c.Equal(fxp.One, money.Quantity)
	c.Equal(fxp.FromInteger(20), money.AdjustedValue())
	c.True(money.DataOwner() == DataOwner(e))
	c.Nil(money.Parent())
	c.Equal(0, len(e.CarriedEquipment), "the copies aren't put in any list")
	c.Equal(1, len(money.Modifiers))
	c.Nil(entry.DataOwner(), "the entry has no owner")
	c.Equal(fxp.One, entry.Quantity, "the entry holds one of the coin, however many were copied")
	c.Equal(fxp.FromInteger(50), silver.Quantity, "the item copied from is untouched")
}

// TestTradesLeaveTheSellMoneyEntriesAlone checks that paying out and returning change in the sell money, which marks
// what is handed out as equipped and carries it, changes nothing about the entries the settings hold, since clones of
// the settings share them.
func TestTradesLeaveTheSellMoneyEntriesAlone(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	penny := denomination("Copper Penny", "1")
	penny.Equipped = false
	c.True(e.SheetSettings.AddSellMoney(penny, LibraryFile{}))
	entry := e.SheetSettings.SellMoney[0]
	c.False(entry.Equipped)
	knife := withQuantity(newEquipmentItem("Small Knife", "30", "1 lb"), "2")
	carry(e, nil, knife)
	when := mustTime(t, "2026-10-16T09:00:00Z")
	payouts, paid := PayoutFor(fxp.FromInteger(15), e.SellMoneyItems())
	c.Equal(fxp.FromInteger(15), paid)
	e.CommitSale([]SaleItem{{Equipment: knife, Quantity: fxp.One}}, payouts, false, "", when)
	c.Equal(2, len(e.CarriedEquipment))
	c.True(e.CarriedEquipment[1].Equipped, "the pennies received are equipped")
	c.Equal(fxp.FromInteger(15), e.CarriedEquipment[1].Quantity)
	change, returned := e.ChangeFor(fxp.Three, e.SellMoneyItems(), nil)
	c.Equal(fxp.Three, returned)
	e.CommitPurchase(nil, nil, change, false, "", when)
	c.Equal(3, len(e.CarriedEquipment))
	c.True(e.CarriedEquipment[2].Equipped, "the change is equipped")
	c.False(entry.Equipped, "the entry isn't marked equipped by being paid out")
	c.Equal(fxp.One, entry.Quantity)
	c.Nil(entry.DataOwner())
	c.Equal(SellMoneyList{entry}, e.SheetSettings.SellMoney, "the entry is still the one listed")
}

// TestTradeQuantityCountsEveryCopyOfAContainer checks that equipment inside a container is counted once for each copy of
// the container, as its extended value is, and that a sale of it takes it out of every copy.
func TestTradeQuantityCountsEveryCopyOfAContainer(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	backpack := withQuantity(newContainerItem("Backpack", "5"), "2")
	torches := withQuantity(newEquipmentItem("Torch", "10", "1 lb"), "2")
	pouch := withQuantity(newContainerItem("Pouch", "1"), "3")
	gems := NewEquipmentGroup(nil, nil)
	gems.Name = "Gems"
	ruby := newEquipmentItem("Ruby", "250", "0")
	carry(e, nil, backpack)
	carry(e, backpack, torches, pouch, gems)
	carry(e, gems, ruby)
	c.Equal(fxp.Two, backpack.TradeQuantity())
	c.Equal(fxp.Four, torches.TradeQuantity(), "two torches come with each of the two backpacks")
	c.Equal(fxp.Six, pouch.TradeQuantity(), "a container inside another counts the same way")
	c.Equal(fxp.Two, gems.TradeQuantity(), "a group comes once with each backpack")
	c.Equal(fxp.Two, ruby.TradeQuantity())
	c.Equal(fxp.FromInteger(40), TradePrice(torches, torches.TradeQuantity(), 100))
	when := mustTime(t, "2026-10-13T09:40:00Z")

	wealth := e.WealthCarried()
	entry := e.CommitSale([]SaleItem{{Equipment: torches, Quantity: fxp.Ten}}, nil, true, "", when)
	c.Equal([]ShoppingRecordItem{{Name: "Torch", Quantity: fxp.Four, Value: fxp.Ten}}, entry.Items,
		"selling all of the torches sells those in every backpack")
	c.False(slices.Contains(backpack.Children, torches), "sold in full, the torches are gone")
	c.Equal(wealth-fxp.FromInteger(40), e.WealthCarried(), "the wealth falls by what was sold")

	wealth = e.WealthCarried()
	entry = e.CommitSale([]SaleItem{{Equipment: pouch, Quantity: fxp.Two}}, nil, true, "", when)
	c.Equal([]ShoppingRecordItem{{Name: "Pouch", Quantity: fxp.Two, Value: fxp.One}}, entry.Items)
	c.Equal(fxp.Two, pouch.Quantity, "selling two of six pouches takes one out of each backpack")
	c.Equal(wealth-fxp.Two, e.WealthCarried())

	wealth = e.WealthCarried()
	entry = e.CommitSale([]SaleItem{{Equipment: gems, Quantity: fxp.One}}, nil, true, "", when)
	c.Equal([]ShoppingRecordItem{{Name: "Gems", Quantity: fxp.Two, Value: fxp.FromInteger(250)}}, entry.Items,
		"a group is sold in full, once for each backpack")
	c.False(slices.Contains(backpack.Children, gems))
	c.Equal(wealth-fxp.FromInteger(500), e.WealthCarried())
}

// TestCommitPurchaseSpendsFromEveryCopyOfAContainer checks that money inside a container can be spent as many times as
// the container has copies, since each copy holds its own, and that spending takes it out of every copy.
func TestCommitPurchaseSpendsFromEveryCopyOfAContainer(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	silver := denomination("Silver Eagle", "20")
	pouch := withQuantity(newContainerItem("Pouch", "0"), "2")
	eagles := withQuantity(silver.Clone(LibraryFile{}, nil, nil, Duplicate), "5")
	carry(e, nil, pouch)
	carry(e, pouch, eagles)
	when := mustTime(t, "2026-10-13T09:40:00Z")
	c.Equal(fxp.Ten, eagles.TradeQuantity())
	c.Equal(fxp.FromInteger(200), e.WealthCarried())

	payments, paid := PaymentFor(fxp.FromInteger(120), e.MoneyItems())
	c.Equal([]Payout{{Money: eagles, Quantity: 6}}, payments, "the eagles in both pouches can be spent")
	c.Equal(fxp.FromInteger(120), paid)

	entry := e.CommitPurchase(nil, payments, nil, true, "", when)
	c.Equal([]ShoppingRecordItem{{Name: "Silver Eagle", Quantity: fxp.FromInteger(-6), Value: fxp.Twenty}}, entry.Money)
	c.Equal(fxp.FromInteger(120), entry.Exchanged)
	c.Equal(fxp.Two, eagles.Quantity, "three of the six come out of each pouch")
	c.Equal(fxp.FromInteger(80), e.WealthCarried(), "the wealth falls by what was spent")

	entry = e.CommitPurchase(nil, []Payout{{Money: eagles, Quantity: 7}}, nil, true, "", when)
	c.Equal([]ShoppingRecordItem{{Name: "Silver Eagle", Quantity: fxp.FromInteger(-4), Value: fxp.Twenty}}, entry.Money,
		"no more than is in every pouch together is spent")
	c.Equal(0, len(pouch.Children), "spent in full, the eagles are gone")
	c.Equal(fxp.Int(0), e.WealthCarried())
}

// TestMergeSkipsMatchesInsideCopiedContainers checks that money and items received are never merged into a match held
// by a container with more than one copy, since every copy would gain them, but go to a match held by none or by single
// containers, and are otherwise carried anew.
func TestMergeSkipsMatchesInsideCopiedContainers(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	silver := denomination("Silver Eagle", "20")
	pouch := withQuantity(newContainerItem("Pouch", "0"), "2")
	eagles := withQuantity(silver.Clone(LibraryFile{}, nil, nil, Duplicate), "5")
	carry(e, nil, pouch)
	carry(e, pouch, eagles)
	when := mustTime(t, "2026-10-13T09:40:00Z")

	entry := e.CommitSale(nil, []Payout{{Money: silver, Quantity: 10}}, true, "", when)
	c.Equal(fxp.FromInteger(200), entry.Exchanged)
	c.Equal(fxp.Five, eagles.Quantity, "the eagles in the pouches are left alone")
	c.Equal(fxp.FromInteger(400), e.WealthCarried(), "the wealth rises by what was received")
	c.Equal(2, len(e.CarriedEquipment))
	if len(e.CarriedEquipment) == 2 {
		added := e.CarriedEquipment[1]
		c.Equal("Silver Eagle", added.Name)
		c.Equal(fxp.Ten, added.Quantity, "the eagles received are carried anew")
	}

	purse := newContainerItem("Purse", "0")
	loose := withQuantity(silver.Clone(LibraryFile{}, nil, nil, Duplicate), "3")
	e.CarriedEquipment = e.CarriedEquipment[:1]
	carry(e, nil, purse)
	carry(e, purse, loose)
	e.CommitPurchase(nil, nil, []Payout{{Money: silver, Quantity: 2}}, true, "", when)
	c.Equal(fxp.Five, eagles.Quantity)
	c.Equal(fxp.Five, loose.Quantity, "a match in a container with one copy takes the merge")
	c.Equal(2, len(e.CarriedEquipment))
}

// TestChangeForReturnsNoMoreThanWasPaid checks that change drawn from the money paid hands back no more of each item than
// was paid with, however many of it are carried, while the sell money denominations are drawn on without limit, and that
// payments of matching money add up.
func TestChangeForReturnsNoMoreThanWasPaid(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	copper := denomination("Copper Penny", "1")
	tens := withQuantity(denomination("Ten", "10"), "3")
	pennies := withQuantity(copper.Clone(LibraryFile{}, nil, nil, Duplicate), "11")
	more := withQuantity(copper.Clone(LibraryFile{}, nil, nil, Duplicate), "11")
	carry(e, nil, tens, pennies, more)

	change, paid := e.ChangeFor(fxp.Eight, nil, []Payout{{Money: tens, Quantity: 1}, {Money: pennies, Quantity: 1}})
	c.Equal(fxp.One, paid, "only the one penny paid can be handed back")
	c.Equal(1, len(change))
	if len(change) == 1 {
		c.Equal("Copper Penny", change[0].Money.Name)
		c.Equal(1, change[0].Quantity)
	}

	change, paid = e.ChangeFor(fxp.Eight, nil, []Payout{
		{Money: tens, Quantity: 1},
		{Money: pennies, Quantity: 2},
		{Money: more, Quantity: 3},
	})
	c.Equal(fxp.Five, paid, "payments of matching money are handed back together")
	c.Equal(1, len(change))
	if len(change) == 1 {
		c.Equal(5, change[0].Quantity)
	}

	change, paid = e.ChangeFor(fxp.Eight, []*Equipment{copper}, []Payout{{Money: tens, Quantity: 1}})
	c.Equal(fxp.Eight, paid, "sell money is drawn on without limit")
	c.Equal([]Payout{{Money: copper, Quantity: 8}}, change)

	// An overpayment too large for the exact search still hands back no more than was paid.
	change, paid = e.ChangeFor(fxp.FromInteger(2_500_000), nil, []Payout{{Money: pennies, Quantity: 5}})
	c.Equal(fxp.Five, paid)
	c.Equal(1, len(change))
	if len(change) == 1 {
		c.Equal(5, change[0].Quantity)
	}
}

// describeCarried describes the entity's carried equipment as the name and quantity of each item, in order, with what
// each container holds in brackets after it.
func describeCarried(e *Entity) []string {
	var describe func(list []*Equipment) []string
	describe = func(list []*Equipment) []string {
		out := make([]string, 0, len(list))
		for _, eqp := range list {
			s := eqp.Name + " ×" + eqp.Quantity.String()
			if len(eqp.Children) != 0 {
				s += "[" + strings.Join(describe(eqp.Children), ", ") + "]"
			}
			out = append(out, s)
		}
		return out
	}
	return describe(e.CarriedEquipment)
}

// TestPartialTradesSplitCopiedContainers checks that selling or spending part of an item held by a container with
// several copies, in an amount the copies can't share without fractions, splits the container apart: the copies left
// alike keep it and the rest get a copy of it, inserted after it, so that no copy is left holding a fraction, and that
// the copies trade on as separate items.
func TestPartialTradesSplitCopiedContainers(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	box := withQuantity(newContainerItem("Box", "0"), "3")
	gem := newEquipmentItem("Gem", "100", "0")
	carry(e, nil, box)
	carry(e, box, gem)
	when := mustTime(t, "2026-10-13T09:40:00Z")

	entry := e.CommitSale([]SaleItem{{Equipment: gem, Quantity: fxp.One}}, nil, true, "", when)
	c.Equal([]ShoppingRecordItem{{Name: "Gem", Quantity: fxp.One, Value: fxp.Hundred}}, entry.Items)
	c.Equal([]string{"Box ×2[Gem ×1]", "Box ×1"}, describeCarried(e),
		"the gem comes out of one box, which is split off empty after the two still holding one")
	c.Equal(fxp.Two, box.Quantity, "the boxes still holding a gem keep the original")
	c.True(gem.Parent() == box)
	empty := e.CarriedEquipment[1]
	c.NotEqual(box.TID, empty.TID, "the box split off is a new copy")
	c.True(empty.DataOwner() == DataOwner(e))
	c.Nil(empty.Parent())
	c.Equal(fxp.Two, gem.TradeQuantity())
	c.Equal(fxp.FromInteger(200), e.WealthCarried())
	e.CommitSale([]SaleItem{{Equipment: gem, Quantity: fxp.Two}}, nil, true, "", when)
	c.Equal([]string{"Box ×2", "Box ×1"}, describeCarried(e), "selling the two gems left removes them")
	c.Equal(fxp.Int(0), e.WealthCarried())

	silver := denomination("Silver Eagle", "20")
	pouch := withQuantity(newContainerItem("Pouch", "0"), "3")
	eagles := withQuantity(silver.Clone(LibraryFile{}, nil, nil, Duplicate), "5")
	e.CarriedEquipment = nil
	carry(e, nil, pouch)
	carry(e, pouch, eagles)
	entry = e.CommitPurchase(nil, []Payout{{Money: eagles, Quantity: 2}}, nil, true, "", when)
	c.Equal([]ShoppingRecordItem{{Name: "Silver Eagle", Quantity: -fxp.Two, Value: fxp.Twenty}}, entry.Money)
	c.Equal([]string{"Pouch ×1[Silver Eagle ×5]", "Pouch ×2[Silver Eagle ×4]"}, describeCarried(e),
		"the two eagles come out of two pouches, one each, which are split off after the one left untouched")
	c.Equal(fxp.One, pouch.Quantity)
	c.Equal(fxp.Five, eagles.Quantity)
	c.Equal(fxp.FromInteger(260), e.WealthCarried())
	payments, paid := PaymentFor(fxp.FromInteger(260), e.MoneyItems())
	c.Equal(fxp.FromInteger(260), paid, "every eagle left can still be spent")
	e.CommitPurchase(nil, payments, nil, true, "", when)
	c.Equal([]string{"Pouch ×1", "Pouch ×2"}, describeCarried(e), "spending the thirteen eagles removes them")
	c.Equal(fxp.Int(0), e.WealthCarried())
}

// TestPartialTradesSplitNestedContainers checks that an item held by containers within containers is shared out over
// every copy of them, as evenly as whole numbers allow, and that only the containers whose copies are left holding
// different amounts are split apart, at whichever depth that happens, which SaleSplits predicts without changing
// anything.
func TestPartialTradesSplitNestedContainers(t *testing.T) {
	c := check.New(t)
	when := mustTime(t, "2026-10-13T09:40:00Z")
	setup := func() (e *Entity, backpack, quiver, arrows *Equipment) {
		e = NewEntity()
		backpack = withQuantity(newContainerItem("Backpack", "10"), "2")
		quiver = withQuantity(newContainerItem("Quiver", "2"), "3")
		arrows = withQuantity(newEquipmentItem("Arrows", "1", "0"), "5")
		carry(e, nil, backpack)
		carry(e, backpack, quiver)
		carry(e, quiver, arrows)
		c.Equal(fxp.FromInteger(30), arrows.TradeQuantity())
		c.Equal(fxp.FromInteger(62), e.WealthCarried())
		return e, backpack, quiver, arrows
	}

	e, backpack, quiver, arrows := setup()
	items := []SaleItem{{Equipment: arrows, Quantity: fxp.Seven}}
	splits := e.SaleSplits(items)
	c.Equal(1, len(splits))
	if len(splits) == 1 {
		c.True(splits[0] == backpack, "the backpack is split, as one holds quivers left with four arrows and three")
	}
	c.Equal([]string{"Backpack ×2[Quiver ×3[Arrows ×5]]"}, describeCarried(e), "predicting the splits changes nothing")
	e.CommitSale(items, nil, true, "", when)
	c.Equal([]string{
		"Backpack ×1[Quiver ×3[Arrows ×4]]",
		"Backpack ×1[Quiver ×2[Arrows ×4], Quiver ×1[Arrows ×3]]",
	}, describeCarried(e), "five quivers are left with four arrows and one with three")
	c.Equal(fxp.One, backpack.Quantity)
	c.Equal(fxp.Three, quiver.Quantity, "the quivers in the backpack that keeps the original aren't split")
	c.Equal(fxp.Four, arrows.Quantity)
	c.Equal(fxp.FromInteger(12), arrows.TradeQuantity())
	c.Equal(fxp.FromInteger(55), e.WealthCarried(), "the wealth falls by the seven arrows sold")

	e, backpack, quiver, arrows = setup()
	items = []SaleItem{{Equipment: arrows, Quantity: fxp.FromInteger(29)}}
	splits = e.SaleSplits(items)
	c.Equal(2, len(splits))
	if len(splits) == 2 {
		c.True(splits[0] == backpack && splits[1] == quiver,
			"both the backpack and the quiver within the one that keeps the original are split")
	}
	e.CommitSale(items, nil, true, "", when)
	c.Equal([]string{
		"Backpack ×1[Quiver ×1[Arrows ×1], Quiver ×2]",
		"Backpack ×1[Quiver ×3]",
	}, describeCarried(e), "the one arrow left stays in the original quiver")
	c.Equal(fxp.One, quiver.Quantity)
	c.Equal(fxp.One, arrows.Quantity)
	c.True(arrows.Parent() == quiver && quiver.Parent() == backpack)
	c.Equal(fxp.FromInteger(33), e.WealthCarried())
}

// TestSaleOfSeveralItemsInACopiedContainerSplitsItOnce checks that selling parts of several items held by the same
// container splits it apart just once, into the copies left alike in all of them, each item being shared out over the
// copies as it was before the sale.
func TestSaleOfSeveralItemsInACopiedContainerSplitsItOnce(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	backpack := withQuantity(newContainerItem("Backpack", "10"), "3")
	arrows := withQuantity(newEquipmentItem("Arrows", "1", "0"), "5")
	torches := withQuantity(newEquipmentItem("Torch", "2", "1 lb"), "2")
	carry(e, nil, backpack)
	carry(e, backpack, arrows, torches)
	when := mustTime(t, "2026-10-13T09:40:00Z")

	entry := e.CommitSale([]SaleItem{
		{Equipment: arrows, Quantity: fxp.Seven},
		{Equipment: torches, Quantity: fxp.Three},
	}, nil, true, "", when)
	c.Equal([]ShoppingRecordItem{
		{Name: "Arrows", Quantity: fxp.Seven, Value: fxp.One},
		{Name: "Torch", Quantity: fxp.Three, Value: fxp.Two},
	}, entry.Items)
	c.Equal([]string{"Backpack ×2[Arrows ×3, Torch ×1]", "Backpack ×1[Arrows ×2, Torch ×1]"}, describeCarried(e),
		"the torches share out evenly, so only the arrows split the backpacks")
	c.Equal(fxp.FromInteger(30+8+6), e.WealthCarried())
}

// TestPartialTradesOfFractionsEmptyCopiesOneAtATime checks that when what is left of an item can't be shared out over
// the copies of its containers in whole numbers, because what is sold or what each copy holds is a fraction, the copies
// are emptied one at a time, so that the only copy left with an amount other than what it held is the one the remainder
// is in, while an amount the copies can share evenly in whole numbers needs no split.
func TestPartialTradesOfFractionsEmptyCopiesOneAtATime(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sack := withQuantity(newContainerItem("Sack", "0"), "3")
	flour := withQuantity(newEquipmentItem("Flour", "1", "1 lb"), "0.5")
	box := withQuantity(newContainerItem("Box", "0"), "3")
	arrows := withQuantity(newEquipmentItem("Arrows", "1", "0"), "5")
	bag := withQuantity(newContainerItem("Bag", "0"), "2")
	grain := withQuantity(newEquipmentItem("Grain", "1", "1 lb"), "1.5")
	carry(e, nil, sack, box, bag)
	carry(e, sack, flour)
	carry(e, box, arrows)
	carry(e, bag, grain)
	when := mustTime(t, "2026-10-13T09:40:00Z")

	e.CommitSale([]SaleItem{
		{Equipment: flour, Quantity: fxp.Half},
		{Equipment: arrows, Quantity: fxp.FromStringForced("0.25")},
		{Equipment: grain, Quantity: fxp.One},
	}, nil, true, "", when)
	c.Equal([]string{
		"Sack ×2[Flour ×0.5]", "Sack ×1",
		"Box ×2[Arrows ×5]", "Box ×1[Arrows ×4.75]",
		"Bag ×2[Grain ×1]",
	}, describeCarried(e))
}

// TestPartialTradesInContainersWithFractionalQuantitiesShareEvenly checks that an item held by a container whose
// quantity isn't a whole number, which can't be split apart, has each copy keep an even share of what is left, rounded
// up to fxp's precision so that no whole unit is lost, even when the copies of the containers holding it multiply out
// to a whole number.
func TestPartialTradesInContainersWithFractionalQuantitiesShareEvenly(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	crate := withQuantity(newContainerItem("Crate", "0"), "1.5")
	gems := withQuantity(newEquipmentItem("Gem", "100", "0"), "2")
	carry(e, nil, crate)
	carry(e, crate, gems)
	c.Equal(fxp.Three, gems.TradeQuantity())
	when := mustTime(t, "2026-10-13T09:40:00Z")

	c.Equal(0, len(e.SaleSplits([]SaleItem{{Equipment: gems, Quantity: fxp.One}})))
	e.CommitSale([]SaleItem{{Equipment: gems, Quantity: fxp.One}}, nil, true, "", when)
	c.Equal([]string{"Crate ×1.5[Gem ×1.3334]"}, describeCarried(e))
	c.True(gems.TradeQuantity() >= fxp.Two, "both gems left can still be sold")
	e.CommitSale([]SaleItem{{Equipment: gems, Quantity: fxp.Two}}, nil, true, "", when)
	c.Equal([]string{"Crate ×1.5"}, describeCarried(e), "selling the two gems removes them rather than leaving dust")

	pouch := withQuantity(newContainerItem("Pouch", "0"), "2")
	rubies := withQuantity(newEquipmentItem("Ruby", "100", "0"), "2")
	carry(e, crate, pouch)
	carry(e, pouch, rubies)
	c.Equal(fxp.Six, rubies.TradeQuantity())
	c.Equal(0, len(e.SaleSplits([]SaleItem{{Equipment: rubies, Quantity: fxp.One}})))
	e.CommitSale([]SaleItem{{Equipment: rubies, Quantity: fxp.One}}, nil, true, "", when)
	c.Equal([]string{"Crate ×1.5[Pouch ×2[Ruby ×1.6667]]"}, describeCarried(e),
		"the pouches, held by a crate and a half, can't be split apart either")
}

// TestCommitSaleOfAContainerTakesPrecedenceOverItsContents checks that when a sale lists both a container and
// something it holds, in either order, only the container is sold, since selling it sells what it holds.
func TestCommitSaleOfAContainerTakesPrecedenceOverItsContents(t *testing.T) {
	c := check.New(t)
	when := mustTime(t, "2026-10-13T09:40:00Z")
	for _, contentsFirst := range []bool{true, false} {
		e := NewEntity()
		backpack := withQuantity(newContainerItem("Backpack", "10"), "3")
		rope := withQuantity(newEquipmentItem("Rope", "5", "1 lb"), "2")
		carry(e, nil, backpack)
		carry(e, backpack, rope)
		items := []SaleItem{{Equipment: backpack, Quantity: fxp.One}, {Equipment: rope, Quantity: fxp.One}}
		if contentsFirst {
			slices.Reverse(items)
		}
		entry := e.CommitSale(items, nil, true, "", when)
		c.Equal([]ShoppingRecordItem{{Name: "Backpack", Quantity: fxp.One, Value: fxp.Twenty}}, entry.Items)
		c.Equal([]string{"Backpack ×2[Rope ×2]"}, describeCarried(e))
	}
}

// TestCommitSalePayoutsMergeAsMatchedBeforeTheSale checks that a payout merges into the carried money it matched
// before the sale's items were taken out, so that money held by a container with several copies, which a merge passes
// over, isn't merged into when the sale splits the container and leaves the money in a single copy.
func TestCommitSalePayoutsMergeAsMatchedBeforeTheSale(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	silver := denomination("Silver Eagle", "20")
	box := withQuantity(newContainerItem("Box", "0"), "3")
	gem := newEquipmentItem("Gem", "100", "0")
	eagles := withQuantity(silver.Clone(LibraryFile{}, nil, nil, Duplicate), "2")
	carry(e, nil, box)
	carry(e, box, gem, eagles)
	when := mustTime(t, "2026-10-13T09:40:00Z")

	e.CommitSale([]SaleItem{{Equipment: gem, Quantity: fxp.Two}}, []Payout{{Money: silver, Quantity: 10}}, true, "",
		when)
	c.Equal([]string{"Box ×1[Gem ×1, Silver Eagle ×2]", "Box ×2[Silver Eagle ×2]", "Silver Eagle ×10"},
		describeCarried(e), "the eagles paid are carried anew rather than merged into those in the box left with the gem")
}

// TestChangeForNeverReturnsLessThanTheSellMoneyAlone checks that drawing on the money paid, which may have to settle
// for an approximate search where the sell money alone could search exactly, never hands back less change than the sell
// money alone would.
func TestChangeForNeverReturnsLessThanTheSellMoneyAlone(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	twenty := denomination("Twenty", "20")
	fifteen := denomination("Fifteen", "15")
	one := withQuantity(denomination("One", "1"), "1")
	carry(e, nil, one)
	over := fxp.FromInteger(4_000_011)
	_, alone := PayoutFor(over, []*Equipment{twenty, fifteen})
	c.Equal(fxp.FromInteger(4_000_010), alone)
	change, paid := e.ChangeFor(over, []*Equipment{twenty, fifteen}, []Payout{{Money: one, Quantity: 1}})
	c.Equal(alone, paid, "the sell money's own change is kept when the money paid can't better it")
	c.Equal([]Payout{{Money: twenty, Quantity: 199_999}, {Money: fifteen, Quantity: 2}}, change)
}

// TestCommitSaleSkipsDisabledEquipment checks that a sale skips what is disabled, and what is held by a disabled
// container, since either counts as if it weren't on the sheet, while a container holding something disabled sells
// without the value of that.
func TestCommitSaleSkipsDisabledEquipment(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sword := newEquipmentItem("Sword", "500", "3 lb")
	sword.SetDisabled(true)
	sack := newContainerItem("Sack", "1")
	sack.SetDisabled(true)
	torch := newEquipmentItem("Torch", "3", "1 lb")
	rope := newEquipmentItem("Rope", "5", "1 lb")
	kit := newContainerItem("Kit", "10")
	bandage := newEquipmentItem("Bandage", "2", "0")
	bandage.SetDisabled(true)
	carry(e, nil, sword, sack, rope, kit)
	carry(e, sack, torch)
	carry(e, kit, bandage)

	c.Nil(e.SaleSplits([]SaleItem{{Equipment: sword, Quantity: fxp.One}}))
	entry := e.CommitSale([]SaleItem{
		{Equipment: sword, Quantity: fxp.One},
		{Equipment: torch, Quantity: fxp.One},
		{Equipment: rope, Quantity: fxp.One},
		{Equipment: kit, Quantity: fxp.One},
	}, nil, true, "Clear out", mustTime(t, "2026-10-14T10:00:00Z"))

	c.Equal([]string{"Sword ×1", "Sack ×1[Torch ×1]"}, describeCarried(e),
		"the disabled sword and the torch in the disabled sack aren't sold")
	c.Equal([]ShoppingRecordItem{
		{Name: "Rope", Quantity: fxp.One, Value: fxp.Five},
		{Name: "Kit", Quantity: fxp.One, Value: fxp.FromInteger(10)},
	}, entry.Items, "the kit sells without the value of its disabled bandage")
}

// TestCommitPurchaseEnablesWhatIsBought checks that a purchase enables what it buys, along with anything it contains,
// before pricing, matching and carrying it, since it may have been copied from a sheet where it was disabled, carries
// it unequipped, and that nothing merges into a carried match held by a disabled container, which counts as if it
// weren't on the sheet.
func TestCommitPurchaseEnablesWhatIsBought(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	rope := newEquipmentItem("Rope", "5", "1 lb")
	sack := newContainerItem("Sack", "1")
	sack.SetDisabled(true)
	torch := newEquipmentItem("Torch", "3", "1 lb")
	carry(e, nil, rope, sack)
	carry(e, sack, torch)

	library := LibraryFile{Library: "Lib", Path: "Gear.eqp"}
	boughtRope := rope.Clone(library, e, nil, Reference)
	boughtRope.SetDisabled(true)
	boughtTorch := torch.Clone(library, e, nil, Reference)
	kit := newContainerItem("Kit", "10")
	kit.SetDisabled(true)
	bandage := newEquipmentItem("Bandage", "2", "0")
	bandage.SetDisabled(true)
	carry(e, kit, bandage)
	boughtKit := kit.Clone(library, e, nil, Reference)
	c.True(boughtKit.IsDisabled() && boughtKit.Children[0].IsDisabled(), "a copy for a sheet keeps the flags")

	entry := e.CommitPurchase([]PurchaseItem{
		{Equipment: boughtRope, Quantity: fxp.Two},
		{Equipment: boughtTorch, Quantity: fxp.One},
		{Equipment: boughtKit, Quantity: fxp.One},
	}, nil, nil, true, "Restock", mustTime(t, "2026-10-14T10:00:00Z"))

	c.Equal([]string{"Rope ×3", "Sack ×1[Torch ×1]", "Torch ×1", "Kit ×1[Bandage ×1]"}, describeCarried(e),
		"the rope merges into the enabled rope once enabled itself; the torch doesn't merge into the disabled sack")
	Traverse(func(eqp *Equipment) bool {
		c.True(eqp.Enabled(), "%s is carried enabled", eqp.Name)
		c.False(eqp.Equipped, "%s is carried unequipped", eqp.Name)
		return false
	}, false, false, e.CarriedEquipment[2:]...)
	c.Equal([]ShoppingRecordItem{
		{Name: "Rope", Quantity: fxp.Two, Value: fxp.Five},
		{Name: "Torch", Quantity: fxp.One, Value: fxp.Three},
		{Name: "Kit", Quantity: fxp.One, Value: fxp.FromInteger(12)},
	}, entry.Items, "the kit is priced with its bandage enabled")
}

// TestTradeTakeKeepsToWhatCopiesThatCantBeSplitGiveUp checks that where the containers holding an item can't be split
// apart, a trade takes only what leaving each copy an even share, rounded up to fxp's precision, gives up: a trade of
// less than one of fxp's smallest units per copy takes nothing, and a sale of it changes nothing and records nothing,
// rather than paying out for an item still carried in full.
func TestTradeTakeKeepsToWhatCopiesThatCantBeSplitGiveUp(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	box := withQuantity(newContainerItem("Box", "0"), "2.5")
	gem := withQuantity(newEquipmentItem("Gem", "10000", "0"), "1")
	carry(e, nil, box)
	carry(e, box, gem)
	c.Equal(fxp.TwoAndAHalf, gem.TradeQuantity())
	c.Equal(fxp.Int(0), gem.TradeTake(fxp.FromStringForced("0.0001")), "less than one unit per copy takes nothing")
	c.Equal(fxp.FromStringForced("0.0003"), gem.TradeTake(fxp.FromStringForced("0.0003")))
	c.Equal(fxp.FromStringForced("0.0003"), gem.TradeTake(fxp.FromStringForced("0.0004")), "a little less is taken")
	c.Equal(fxp.One, gem.TradeTake(fxp.One))
	c.Equal(fxp.TwoAndAHalf, gem.TradeTake(fxp.Three), "no more than there is")
	c.Equal(fxp.Int(0), gem.TradeTake(0))
	knife := withQuantity(newEquipmentItem("Knife", "30", "0"), "10")
	carry(e, nil, knife)
	c.Equal(fxp.FromStringForced("0.0001"), knife.TradeTake(fxp.FromStringForced("0.0001")),
		"an item held by no container is taken as asked")
	group := NewEquipment(nil, nil, true)
	group.Name = "Group"
	group.ContainerType = eqcontainer.Group
	carry(e, box, group)
	c.Equal(fxp.TwoAndAHalf, group.TradeTake(fxp.One), "a group is taken in full")

	entry := e.CommitSale([]SaleItem{{Equipment: gem, Quantity: fxp.FromStringForced("0.0001")}}, nil, true, "", jio.Now())
	c.Equal(0, len(entry.Items), "a sale that takes nothing records nothing")
	c.Equal(fxp.One, gem.Quantity)
	entry = e.CommitSale([]SaleItem{{Equipment: gem, Quantity: fxp.FromStringForced("0.0004")}}, nil, true, "", jio.Now())
	c.Equal([]ShoppingRecordItem{{Name: "Gem", Quantity: fxp.FromStringForced("0.0003"), Value: fxp.TenThousand}},
		entry.Items, "the sale records what it takes")
	c.Equal(fxp.FromStringForced("0.9999"), gem.Quantity)
}
