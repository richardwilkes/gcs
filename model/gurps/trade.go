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
	"maps"
	"math"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// MoneyTag is the tag that marks carried equipment as money that buying can pay with.
const MoneyTag = "Money"

// SaleItem is a quantity of a piece of carried equipment to sell.
type SaleItem struct {
	Equipment *Equipment
	Quantity  fxp.Int
}

// PurchaseItem is a quantity of a piece of equipment to buy. Equipment is a clone owned by the entity buying it and not
// yet in any list.
type PurchaseItem struct {
	Equipment *Equipment
	Quantity  fxp.Int
}

// MoneyItems returns the carried equipment, at any depth, that is tagged as money (see MoneyTag) and has a positive
// value, in the order it is listed. Containers are never included.
func (e *Entity) MoneyItems() []*Equipment {
	var list []*Equipment
	Traverse(func(eqp *Equipment) bool {
		if HasTag(MoneyTag, eqp.Tags) && eqp.AdjustedValue() > 0 {
			list = append(list, eqp)
		}
		return false
	}, false, true, e.CarriedEquipment...)
	return list
}

// SellMoneyItems returns a copy of the library equipment each of the sheet's sell money sources refers to, with a
// quantity of one, as the denominations to give PayoutFor. Each copy is owned by this entity, isn't in any list and has
// the source it came from as its Source, unless the library's item has a Source of its own, from another library, which
// the copy keeps instead (see SourcedID.AdjustSource). Sources that don't refer to non-container equipment in the
// library are skipped.
func (e *Entity) SellMoneyItems() []*Equipment {
	sm := e.SourceMatcher()
	list := make([]*Equipment, 0, len(e.SheetSettings.SellMoney))
	for _, src := range e.SheetSettings.SellMoney {
		if lib := sm.libraryEquipment(src); lib != nil {
			money := lib.Clone(src.LibraryFile, e, nil, Reference)
			money.Quantity = fxp.One
			list = append(list, money)
		}
	}
	return list
}

// MergeMatches returns true if other may be merged into this equipment by adding its quantity to this one's: neither is
// a container, and they match in everything but their IDs, quantities, sources, open states and third-party data.
func (e *Equipment) MergeMatches(other *Equipment) bool {
	if e == nil || other == nil || e.Container() || other.Container() {
		return false
	}
	return Hash64(e) == Hash64(other) &&
		e.Equipped == other.Equipped &&
		e.Level == other.Level &&
		e.Uses == other.Uses &&
		e.RatedST == other.RatedST &&
		e.VTTNotes == other.VTTNotes &&
		e.ItemSwitch == other.ItemSwitch &&
		e.Preconfigured == other.Preconfigured &&
		e.PickSeparately == other.PickSeparately &&
		maps.Equal(e.Replacements, other.Replacements) &&
		weaponsMatch(e.Weapons, other.Weapons) &&
		equipmentModifiersMatch(e.Modifiers, other.Modifiers)
}

// weaponsMatch returns true if the weapons match in what WeaponData.Hash leaves out: whether each is hidden. The rest of
// their data, and their number, is compared through the hash of the equipment holding them.
func weaponsMatch(a, b []*Weapon) bool {
	if len(a) != len(b) {
		return false
	}
	for i, one := range a {
		if one.Hide != b[i].Hide {
			return false
		}
	}
	return true
}

func equipmentModifiersMatch(a, b []*EquipmentModifier) bool {
	if len(a) != len(b) {
		return false
	}
	for i, one := range a {
		other := b[i]
		if Hash64(one) != Hash64(other) ||
			one.Disabled != other.Disabled ||
			one.VTTNotes != other.VTTNotes ||
			!maps.Equal(one.Replacements, other.Replacements) ||
			!equipmentModifiersMatch(one.Children, other.Children) {
			return false
		}
	}
	return true
}

// TradePrice returns the price of qty of the equipment, including what each one contains, at percent of its full
// value. Returns 0 for nil equipment or a quantity that isn't positive.
func TradePrice(eqp *Equipment, qty, percent fxp.Int) fxp.Int {
	if eqp == nil || qty <= 0 {
		return 0
	}
	return eqp.ExtendedValueOfJustOne().Mul(qty).Mul(percent).Div(fxp.Hundred)
}

// TradeQuantity returns how many of this equipment there are to trade: its quantity times that of each container
// holding it, since every copy of a container holds its own contents, which is how the extended value counts them. A
// group, which has no quantity of its own, comes once for each copy of the containers holding it.
func (e *Equipment) TradeQuantity() fxp.Int {
	copies := e.containerCopies()
	if e.IsGroup() {
		return copies
	}
	return e.Quantity.Mul(copies)
}

// containerCopies returns how many copies there are of the containers holding this equipment: the product of their
// quantities, or one for equipment held by none.
func (e *Equipment) containerCopies() fxp.Int {
	copies := fxp.One
	for parent := e.Parent(); parent != nil; parent = parent.Parent() {
		copies = copies.Mul(parent.Quantity)
	}
	return copies
}

// CommitSale sells the items from the carried equipment and adds the payouts to it, then records the sale under name
// and when in the shopping record and returns the entry. Each item's quantity is limited to its TradeQuantity; selling
// all of an item removes it, along with anything it contains, and selling less reduces its quantity by what each copy
// of the containers holding it gives up. A group is sold in full. Items that are not carried, such as the contents of a
// container sold earlier in the list, are skipped. Each payout's Money is a denomination from SellMoneyItems: when
// merge is true, its quantity is added to the first carried item it MergeMatches, if any (see addToCarried), and
// otherwise a copy of it is carried, equipped.
func (e *Entity) CommitSale(items []SaleItem, payouts []Payout, merge bool, name string, when jio.Time) *ShoppingRecordEntry {
	entry := &ShoppingRecordEntry{When: when, Name: name, Kind: ShoppingSold}
	for _, item := range items {
		eqp := item.Equipment
		if eqp == nil || !e.carries(eqp) {
			continue
		}
		most := eqp.TradeQuantity()
		qty := min(item.Quantity, most)
		if qty <= 0 {
			continue
		}
		if eqp.IsGroup() {
			qty = most
		}
		entry.Items = append(entry.Items, ShoppingRecordItem{
			Name:     eqp.String(),
			Quantity: qty,
			Value:    eqp.ExtendedValueOfJustOne(),
		})
		if qty >= most {
			e.removeCarried(eqp)
		} else {
			eqp.Quantity -= qty.Div(eqp.containerCopies())
		}
	}
	entry.Exchanged = e.receiveMoney(entry, payouts, merge)
	e.Recalculate()
	e.AddShoppingRecord(entry)
	return entry
}

// CommitPurchase spends the payments from the carried equipment and carries the items and the change, then records the
// purchase under name and when in the shopping record and returns the entry. Each payment's Money is a carried money
// item, such as one from MoneyItems, found again by its ID, since the pointer may be stale; no more is spent than its
// TradeQuantity, spending less takes from every copy of the containers holding it, and a money item with none left is
// removed. Each item is bought at least once, and a group just once. Each change payout's Money is a denomination from
// ChangeFor: one of the sheet's sell money, or a copy of money paid. When merge is true, each item and change payout
// adds its quantity to the first carried item it MergeMatches, if any (see addToCarried); otherwise a copy of it is
// carried, equipped.
func (e *Entity) CommitPurchase(items []PurchaseItem, payments, change []Payout, merge bool, name string, when jio.Time) *ShoppingRecordEntry {
	entry := &ShoppingRecordEntry{When: when, Name: name, Kind: ShoppingBought}
	var paid fxp.Int
	for _, payment := range payments {
		if payment.Money == nil || payment.Quantity <= 0 {
			continue
		}
		money := e.carriedByID(payment.Money.TID)
		if money == nil {
			continue
		}
		most := money.TradeQuantity()
		qty := min(fxp.FromInteger(payment.Quantity), most)
		if qty <= 0 {
			continue
		}
		value := money.AdjustedValue()
		entry.Money = append(entry.Money, ShoppingRecordItem{Name: money.String(), Quantity: -qty, Value: value})
		paid += value.Mul(qty)
		if qty >= most {
			e.removeCarried(money)
		} else {
			money.Quantity -= qty.Div(money.containerCopies())
		}
	}
	for _, item := range items {
		eqp := item.Equipment
		if eqp == nil {
			continue
		}
		qty := fxp.One
		if !eqp.IsGroup() {
			qty = max(fxp.One, item.Quantity)
		}
		entry.Items = append(entry.Items, ShoppingRecordItem{
			Name:     eqp.String(),
			Quantity: qty,
			Value:    eqp.ExtendedValueOfJustOne(),
		})
		e.addToCarried(eqp, qty, merge)
	}
	entry.Exchanged = (paid - e.receiveMoney(entry, change, merge)).Abs()
	e.Recalculate()
	e.AddShoppingRecord(entry)
	return entry
}

// ChangeFor returns the change for an overpayment of over, and what it comes to: the money items from sellMoney, the
// sheet's sell money denominations (see SellMoneyItems), that PayoutFor picks. When those can't return all of it, the
// money given in payments is drawn on as well, since whoever is paid can hand some of it back, though no more of it
// than was paid: a copy of each payment's money that none of the denominations MergeMatches is added to them, owned by
// this entity, with a quantity of one and in no list, and the quantities of the payments whose money it MergeMatches
// are the most of it that is returned.
func (e *Entity) ChangeFor(over fxp.Int, sellMoney []*Equipment, payments []Payout) (change []Payout, paid fxp.Int) {
	change, paid = PayoutFor(over, sellMoney)
	if paid >= over {
		return change, paid
	}
	money := slices.Clone(sellMoney)
	limits := make(map[*Equipment]int64)
	for _, payment := range payments {
		if payment.Money == nil || payment.Quantity <= 0 {
			continue
		}
		if i := slices.IndexFunc(money, payment.Money.MergeMatches); i != -1 {
			if _, limited := limits[money[i]]; limited {
				limits[money[i]] += int64(payment.Quantity)
			}
			continue
		}
		given := payment.Money.Clone(LibraryFile{}, e, nil, Duplicate)
		given.Quantity = fxp.One
		money = append(money, given)
		limits[given] = int64(payment.Quantity)
	}
	if len(money) == len(sellMoney) {
		return change, paid
	}
	coins := payoutCoins(money)
	for i := range coins {
		if limit, limited := limits[coins[i].eqp]; limited {
			coins[i].limit = limit
		} else {
			coins[i].limit = math.MaxInt64
		}
	}
	return payoutWithin(over, coins)
}

// receiveMoney adds the money to the carried equipment (see addToCarried), recording each in the entry, and returns its
// total value.
func (e *Entity) receiveMoney(entry *ShoppingRecordEntry, money []Payout, merge bool) fxp.Int {
	var total fxp.Int
	for _, one := range money {
		if one.Money == nil || one.Quantity <= 0 {
			continue
		}
		qty := fxp.FromInteger(one.Quantity)
		value := one.Money.AdjustedValue()
		entry.Money = append(entry.Money, ShoppingRecordItem{Name: one.Money.String(), Quantity: qty, Value: value})
		total += value.Mul(qty)
		e.addToCarried(one.Money, qty, merge)
	}
	return total
}

// addToCarried marks the equipment and anything it contains as equipped, as carried equipment normally is, then, when
// merge is true, adds qty to the first carried item it MergeMatches that isn't held by a container with more than one
// copy, since every copy would gain the quantity (see TradeQuantity). Otherwise, or when there is none, it carries a
// copy of the equipment with a quantity of qty.
func (e *Entity) addToCarried(eqp *Equipment, qty fxp.Int, merge bool) {
	Traverse(func(one *Equipment) bool {
		one.Equipped = true
		return false
	}, false, false, eqp)
	if merge {
		var match *Equipment
		Traverse(func(one *Equipment) bool {
			if one.MergeMatches(eqp) && one.containerCopies() == fxp.One {
				match = one
				return true
			}
			return false
		}, false, true, e.CarriedEquipment...)
		if match != nil {
			match.Quantity += qty
			return
		}
	}
	clone := eqp.Clone(LibraryFile{}, e, nil, Duplicate)
	clone.Quantity = qty
	e.CarriedEquipment = append(e.CarriedEquipment, clone)
}

// carries returns true if the equipment is among the carried equipment, at any depth.
func (e *Entity) carries(eqp *Equipment) bool {
	for parent := eqp.Parent(); parent != nil; parent = parent.Parent() {
		if !slices.Contains(parent.Children, eqp) {
			return false
		}
		eqp = parent
	}
	return slices.Contains(e.CarriedEquipment, eqp)
}

// carriedByID returns the carried non-container with the ID, at any depth, or nil if there is none.
func (e *Entity) carriedByID(id tid.TID) *Equipment {
	var found *Equipment
	Traverse(func(eqp *Equipment) bool {
		if eqp.TID == id {
			found = eqp
			return true
		}
		return false
	}, false, true, e.CarriedEquipment...)
	return found
}

// removeCarried removes the carried equipment, along with anything it contains.
func (e *Entity) removeCarried(eqp *Equipment) {
	isEqp := func(one *Equipment) bool { return one == eqp }
	if parent := eqp.Parent(); parent != nil {
		parent.Children = slices.DeleteFunc(parent.Children, isEqp)
		eqp.SetParent(nil)
	} else {
		e.CarriedEquipment = slices.DeleteFunc(e.CarriedEquipment, isEqp)
	}
}
