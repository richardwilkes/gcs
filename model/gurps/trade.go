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

// MoneyItems returns the enabled carried equipment, at any depth, that is tagged as money (see MoneyTag) and has a
// positive value, in the order it is listed. Containers are never included, nor is anything disabled or held by a
// disabled container, which counts as if it weren't on the sheet.
func (e *Entity) MoneyItems() []*Equipment {
	var list []*Equipment
	Traverse(func(eqp *Equipment) bool {
		if HasTag(MoneyTag, eqp.Tags) && eqp.AdjustedValue() > 0 {
			list = append(list, eqp)
		}
		return false
	}, true, true, e.CarriedEquipment...)
	return list
}

// SellMoneyItems returns a copy of the library equipment each of the sheet's sell money sources refers to, with a
// quantity of one, as the denominations to give PayoutFor. Each copy is owned by this entity, isn't in any list and has
// the source it came from as its Source, unless the library's item has a Source of its own, from another library, which
// the copy keeps instead (see SourcedID.AdjustSource). Sources that don't refer to non-container equipment tagged as
// money in the library (see SheetSettings.CanAddSellMoney) are skipped, since buying spends only money so tagged.
func (e *Entity) SellMoneyItems() []*Equipment {
	sm := e.SourceMatcher()
	list := make([]*Equipment, 0, len(e.SheetSettings.SellMoney))
	for _, src := range e.SheetSettings.SellMoney {
		if lib := sm.libraryEquipment(src); lib != nil && HasTag(MoneyTag, lib.Tags) {
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
func TradePrice(eqp *Equipment, qty fxp.Int, percent int) fxp.Int {
	if eqp == nil || qty <= 0 {
		return 0
	}
	return eqp.ExtendedValueOfJustOne().Mul(qty).Mul(fxp.FromInteger(percent)).Div(fxp.Hundred)
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

// TradeTake returns how much of this equipment a trade asked to take qty of it takes: qty, though no more than its
// TradeQuantity, and all of a group, which has no quantity of its own. Where the containers holding it can't be split
// apart (see tradeTakes), each of their copies keeps an even share of the rest, rounded up to fxp's precision, so the
// trade takes what that leaves over, which can be a little less than qty and is nothing when qty is less than one of
// fxp's smallest units per copy.
func (e *Equipment) TradeTake(qty fxp.Int) fxp.Int {
	most := e.TradeQuantity()
	switch {
	case qty <= 0 || most <= 0:
		return 0
	case e.IsGroup() || qty >= most:
		return most
	case heldByWholeCopies(e):
		return qty
	}
	copies := e.containerCopies()
	return most - fractionalShare(most-qty, copies).Mul(copies)
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
// and when in the shopping record and returns the entry. Each item's quantity is kept to what a sale of it takes (see
// Equipment.TradeTake): no more than its TradeQuantity, all of a group and only what the copies of the containers
// holding it give up. Selling all of an item removes it, along with anything it contains, and selling less leaves
// the copies of the containers holding it with the rest, splitting the containers apart when the copies can't share it
// without fractions (see tradeTakes). Items that are not carried or not enabled are skipped, as are those held by
// another item in the list, since selling a container sells what it holds. Each payout's Money is a denomination from
// SellMoneyItems: when merge is true, its quantity is added to the first carried item it MergeMatches, if any, as
// matched before the items are taken out (see MergeTarget), and otherwise a copy of it is carried, enabled and equipped
// (see addToCarried).
func (e *Entity) CommitSale(items []SaleItem, payouts []Payout, merge bool, name string, when jio.Time) *ShoppingRecordEntry {
	entry := &ShoppingRecordEntry{When: when, Name: name, Kind: ShoppingSold}
	takes := e.saleTakes(items)
	for _, take := range takes.list {
		entry.Items = append(entry.Items, ShoppingRecordItem{
			Name:     take.eqp.String(),
			Quantity: take.quantity,
			Value:    take.eqp.ExtendedValueOfJustOne(),
		})
	}
	targets := e.mergeTargets(payouts, merge, takes.removed())
	takes.apply()
	entry.Exchanged = e.receiveMoney(entry, payouts, targets)
	e.Recalculate()
	e.AddShoppingRecord(entry)
	return entry
}

// SaleSplits returns the carried containers that a CommitSale of the items would split apart, leaving each with fewer
// copies and the rest to new copies of it (see tradeTakes). Nothing is changed.
func (e *Entity) SaleSplits(items []SaleItem) []*Equipment {
	return e.saleTakes(items).splits()
}

// PaymentSplits returns the carried containers that a CommitPurchase spending the payments would split apart, leaving
// each with fewer copies and the rest to new copies of it (see tradeTakes). Nothing is changed.
func (e *Entity) PaymentSplits(payments []Payout) []*Equipment {
	return e.paymentTakes(payments).splits()
}

// saleTakes returns what a sale of the items takes out of the carried equipment (see CommitSale).
func (e *Entity) saleTakes(items []SaleItem) *tradeTakes {
	takes := e.newTradeTakes()
	for _, item := range items {
		takes.add(item.Equipment, item.Quantity)
	}
	return takes
}

// paymentTakes returns what spending the payments takes out of the carried equipment (see CommitPurchase).
func (e *Entity) paymentTakes(payments []Payout) *tradeTakes {
	takes := e.newTradeTakes()
	for _, payment := range payments {
		if payment.Money != nil && payment.Quantity > 0 {
			takes.add(e.carriedByID(payment.Money.TID), fxp.FromInteger(payment.Quantity))
		}
	}
	return takes
}

// CommitPurchase spends the payments from the carried equipment and carries the items and the change, then records the
// purchase under name and when in the shopping record and returns the entry. Each payment's Money is a carried money
// item, such as one from MoneyItems, found again by its ID, since the pointer may be stale. No more is spent than its
// TradeQuantity; spending less leaves the copies of the containers holding it with the rest, splitting the containers
// apart when the copies can't share it without fractions (see tradeTakes), and a money item with none left is removed.
// Each item is bought at least once, and a group just once. It is enabled and equipped first, along with anything it
// contains, whatever it was copied from, so that it is priced, matched and carried as something in use. Each change
// payout's Money is a denomination from ChangeFor: one of the sheet's sell money, or a copy of money paid. When merge
// is true, each item and change payout adds its quantity to the first carried item it MergeMatches, if any (see
// MergeTarget); otherwise a copy of it is carried (see addToCarried).
func (e *Entity) CommitPurchase(items []PurchaseItem, payments, change []Payout, merge bool, name string, when jio.Time) *ShoppingRecordEntry {
	entry := &ShoppingRecordEntry{When: when, Name: name, Kind: ShoppingBought}
	takes := e.paymentTakes(payments)
	var paid fxp.Int
	for _, take := range takes.list {
		value := take.eqp.AdjustedValue()
		entry.Money = append(entry.Money, ShoppingRecordItem{Name: take.eqp.String(), Quantity: -take.quantity, Value: value})
		paid = fxp.SaturatingAdd(paid, value.Mul(take.quantity))
	}
	takes.apply()
	for _, item := range items {
		eqp := item.Equipment
		if eqp == nil {
			continue
		}
		enableAndEquip(eqp)
		qty := fxp.One
		if !eqp.IsGroup() {
			qty = max(fxp.One, item.Quantity)
		}
		entry.Items = append(entry.Items, ShoppingRecordItem{
			Name:     eqp.String(),
			Quantity: qty,
			Value:    eqp.ExtendedValueOfJustOne(),
		})
		var target *Equipment
		if merge {
			target = e.MergeTarget(eqp, nil)
		}
		e.addToCarried(eqp, qty, target)
	}
	entry.Exchanged = (paid - e.receiveMoney(entry, change, e.mergeTargets(change, merge, nil))).Abs()
	e.Recalculate()
	e.AddShoppingRecord(entry)
	return entry
}

// ChangeFor returns the change for an overpayment of over, and what it comes to: the money items from sellMoney, the
// sheet's sell money denominations (see SellMoneyItems), that PayoutFor picks. When those can't return all of it, the
// money given in payments is drawn on as well, since whoever is paid can hand some of it back, though no more of it
// than was paid: a copy of each payment's money that none of the denominations MergeMatches is added to them, owned by
// this entity, with a quantity of one and in no list, and the quantities of the payments whose money it MergeMatches
// are the most of it that is returned. Drawing on the money paid never returns less than the sell money alone, even
// where the larger search has to settle for an approximate result.
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
	if within, withinPaid := payoutWithin(over, coins); withinPaid > paid {
		return within, withinPaid
	}
	return change, paid
}

// receiveMoney adds the money to the carried equipment, each payout going to its entry in targets or, when that is nil,
// being carried anew (see addToCarried), recording each in the entry, and returns its total value.
func (e *Entity) receiveMoney(entry *ShoppingRecordEntry, money []Payout, targets []*Equipment) fxp.Int {
	var total fxp.Int
	for i, one := range money {
		if one.Money == nil || one.Quantity <= 0 {
			continue
		}
		qty := fxp.FromInteger(one.Quantity)
		value := one.Money.AdjustedValue()
		entry.Money = append(entry.Money, ShoppingRecordItem{Name: one.Money.String(), Quantity: qty, Value: value})
		total = fxp.SaturatingAdd(total, value.Mul(qty))
		e.addToCarried(one.Money, qty, targets[i])
	}
	return total
}

// PayoutMergeTargets returns the carried item each payout's money would merge into were the payouts received now with
// merging on, as CommitSale and CommitPurchase match them (see MergeTarget), passing over the IDs in except, or nil for
// each that would be carried anew. The money is matched as it will be when carried, enabled and equipped, but is left
// as it is, since the caller may go on comparing it with other copies of it.
func (e *Entity) PayoutMergeTargets(payouts []Payout, except map[tid.TID]bool) []*Equipment {
	copies := make([]Payout, len(payouts))
	for i, payout := range payouts {
		if payout.Money != nil {
			copies[i] = Payout{Money: payout.Money.Clone(LibraryFile{}, e, nil, Duplicate), Quantity: payout.Quantity}
		}
	}
	return e.mergeTargets(copies, true, except)
}

// mergeTargets returns the MergeTarget of each payout's money, passing over the IDs in except, or nil for each when
// merge is false. The money is marked enabled and equipped first, as it will be when carried, since MergeMatches
// compares both.
func (e *Entity) mergeTargets(payouts []Payout, merge bool, except map[tid.TID]bool) []*Equipment {
	targets := make([]*Equipment, len(payouts))
	if merge {
		for i, payout := range payouts {
			if payout.Money != nil && payout.Quantity > 0 {
				enableAndEquip(payout.Money)
				targets[i] = e.MergeTarget(payout.Money, except)
			}
		}
	}
	return targets
}

// MergeTarget returns the carried item that a sale or purchase merging the equipment adds its quantity to: the first
// enabled one, at any depth, that it MergeMatches and that isn't held by a container with more than one copy, since
// every copy would gain the quantity (see TradeQuantity). Items whose IDs are in except, such as those the trade removes
// first, are passed over, as is anything held by a disabled container, which counts as if it weren't on the sheet.
// Returns nil if there is none.
func (e *Entity) MergeTarget(eqp *Equipment, except map[tid.TID]bool) *Equipment {
	var match *Equipment
	Traverse(func(one *Equipment) bool {
		if !except[one.TID] && one.MergeMatches(eqp) && one.containerCopies() == fxp.One {
			match = one
			return true
		}
		return false
	}, true, true, e.CarriedEquipment...)
	return match
}

// enableAndEquip marks the equipment and anything it contains as enabled and equipped, as carried equipment normally
// is. Something bought or received may have been copied from a sheet where it was disabled.
func enableAndEquip(eqp *Equipment) {
	Traverse(func(one *Equipment) bool {
		one.SetDisabled(false)
		one.Equipped = true
		return false
	}, false, false, eqp)
}

// addToCarried marks the equipment and anything it contains as enabled and equipped (see enableAndEquip), then adds qty
// to the target, a carried item it MergeMatches, or, when that is nil, carries a copy of the equipment with a quantity
// of qty.
func (e *Entity) addToCarried(eqp *Equipment, qty fxp.Int, target *Equipment) {
	enableAndEquip(eqp)
	if target != nil {
		target.Quantity += qty
		return
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

// tradeTake is how much of a carried item's TradeQuantity a trade takes.
type tradeTake struct {
	eqp      *Equipment
	most     fxp.Int
	quantity fxp.Int
}

// tradeTakes collects what a trade takes out of the carried equipment, so that it can all be applied at once. Taking
// all of an item removes it, along with anything it contains. Taking less leaves the copies of the containers holding
// it with the rest, shared out as tradeShares describes, and where that leaves some copies of a container holding
// different amounts than the others, the container is split apart: the first run of copies left alike keeps it, and
// each further run gets a copy of it, inserted after it in the list holding it. Containers whose quantities aren't
// whole numbers can't be split apart, so each copy of those keeps an even share of the rest, rounded up to fxp's
// precision so that no whole unit is lost, unless less than one of fxp's smallest units per copy is left, which removes
// the item instead.
type tradeTakes struct {
	entity *Entity
	list   []*tradeTake
	byEqp  map[*Equipment]*tradeTake
}

func (e *Entity) newTradeTakes() *tradeTakes {
	return &tradeTakes{entity: e, byEqp: make(map[*Equipment]*tradeTake)}
}

// add records taking qty more of the carried equipment, as much of it as a trade of that much takes (see TradeTake), so
// never more than is left of it, all of a group and only what the copies of the containers holding it can give up.
// Nothing is recorded for equipment that isn't carried, isn't enabled, is held by equipment already recorded or of
// which a trade of qty would take nothing, and recording a container drops what was recorded for anything it holds,
// since taking a container takes what it holds.
func (t *tradeTakes) add(eqp *Equipment, qty fxp.Int) {
	if eqp == nil || qty <= 0 || !eqp.Enabled() || !t.entity.carries(eqp) {
		return
	}
	take := t.byEqp[eqp]
	if take == nil {
		for parent := eqp.Parent(); parent != nil; parent = parent.Parent() {
			if t.byEqp[parent] != nil {
				return
			}
		}
		if eqp.TradeTake(qty) <= 0 {
			return
		}
		if eqp.HasChildren() {
			t.list = slices.DeleteFunc(t.list, func(other *tradeTake) bool {
				if !heldBy(other.eqp, eqp) {
					return false
				}
				delete(t.byEqp, other.eqp)
				return true
			})
		}
		take = &tradeTake{eqp: eqp, most: eqp.TradeQuantity()}
		t.list = append(t.list, take)
		t.byEqp[eqp] = take
	}
	take.quantity = eqp.TradeTake(take.quantity + qty)
}

// heldBy returns true if the container holds the equipment, at any depth.
func heldBy(eqp, container *Equipment) bool {
	for parent := eqp.Parent(); parent != nil; parent = parent.Parent() {
		if parent == container {
			return true
		}
	}
	return false
}

// removed returns the IDs of the carried equipment the takes remove: each item taken in full and anything it contains.
func (t *tradeTakes) removed() map[tid.TID]bool {
	ids := make(map[tid.TID]bool)
	for _, take := range t.list {
		if take.leaves() == nil {
			Traverse(func(one *Equipment) bool {
				ids[one.TID] = true
				return false
			}, false, false, take.eqp)
		}
	}
	return ids
}

// apply makes the takes, changing the carried equipment as tradeTakes describes.
func (t *tradeTakes) apply() {
	tops, held := t.held()
	for _, take := range t.list {
		if shares := take.leaves(); len(shares) == 1 {
			t.entity.leave(take.eqp, shares[0].quantity)
		} else if shares == nil {
			t.entity.removeCarried(take.eqp)
		}
	}
	for _, top := range tops {
		t.entity.splitHeld(top, held[top])
	}
}

// splits returns the containers that apply would split apart, in no particular order. Nothing is changed.
func (t *tradeTakes) splits() []*Equipment {
	var list []*Equipment
	tops, held := t.held()
	for _, top := range tops {
		wouldSplit(top, held[top], func(c *Equipment) { list = append(list, c) })
	}
	return list
}

// held returns the takes that leave the copies of the containers holding their items unalike (see tradeShares), by the
// carried container at the top of each, in the order first seen.
func (t *tradeTakes) held() (tops []*Equipment, held map[*Equipment][]heldTake) {
	held = make(map[*Equipment][]heldTake)
	for _, take := range t.list {
		if shares := take.leaves(); len(shares) > 1 {
			top := take.eqp
			for top.Parent() != nil {
				top = top.Parent()
			}
			if _, seen := held[top]; !seen {
				tops = append(tops, top)
			}
			held[top] = append(held[top], heldTake{eqp: take.eqp, shares: shares})
		}
	}
	return tops, held
}

// leaves returns what the take leaves the item at in each copy of the containers holding it, in order, or nil when it
// takes all of it. A single share stands for every copy, including the one copy of an item held by no container, and
// is all there can be when the copies can't be split apart (see tradeTakes); otherwise see tradeShares.
func (take *tradeTake) leaves() []tradeShare {
	if take.quantity >= take.most {
		return nil
	}
	left := take.most - take.quantity
	copies := take.eqp.containerCopies()
	if !heldByWholeCopies(take.eqp) {
		each := fractionalShare(left, copies)
		if each <= 0 {
			return nil
		}
		return []tradeShare{{quantity: each, copies: 1}}
	}
	return tradeShares(take.eqp.Quantity, left, copies.AsInteger[int64]())
}

// fractionalShare returns what each of the copies of the containers holding an item is left holding when they can't be
// split apart (see heldByWholeCopies) and left of the item remains between them: an even share, rounded up to fxp's
// precision so that no whole unit is lost, or nothing when less than one of fxp's smallest units per copy is left.
func fractionalShare(left, copies fxp.Int) fxp.Int {
	each := left.Div(copies)
	if each <= 0 {
		return 0
	}
	if each.Mul(copies) < left {
		each++
	}
	return each
}

// heldByWholeCopies returns true if each container holding the equipment has a whole number of copies, one or more,
// which is what lets them be split apart.
func heldByWholeCopies(eqp *Equipment) bool {
	for parent := eqp.Parent(); parent != nil; parent = parent.Parent() {
		if parent.Quantity < fxp.One || parent.Quantity.Floor() != parent.Quantity {
			return false
		}
	}
	return true
}

// tradeShare is a run of the copies of the containers holding an item that a trade leaves holding the same quantity
// of it.
type tradeShare struct {
	quantity fxp.Int
	copies   int64
}

// tradeShares returns what each of the n copies of the containers holding an item is left with when they hold each of
// it apiece and left of it between them. A whole amount is shared out as evenly as whole numbers allow, the copies that
// get one more coming first, so that part of something countable is never left as a fraction. Otherwise, and when
// evening out would give a copy more than it held, the copies are emptied one at a time: those left holding what they
// held come first, then the one left holding the remainder, if any, then those left holding none.
func tradeShares(each, left fxp.Int, n int64) []tradeShare {
	if left.Floor() == left {
		whole := left.AsInteger[int64]()
		lo := whole / n
		more := whole - lo*n
		if more == 0 {
			return []tradeShare{{quantity: fxp.FromInteger(lo), copies: n}}
		}
		if hi := fxp.FromInteger(lo + 1); hi <= each {
			return []tradeShare{{quantity: hi, copies: more}, {quantity: fxp.FromInteger(lo), copies: n - more}}
		}
	}
	full := min(left.Div(each).Floor(), fxp.FromInteger(n-1))
	rest := n - full.AsInteger[int64]()
	shares := make([]tradeShare, 0, 3)
	if full > 0 {
		shares = append(shares, tradeShare{quantity: each, copies: full.AsInteger[int64]()})
	}
	if remainder := left - each.Mul(full); remainder > 0 {
		shares = append(shares, tradeShare{quantity: remainder, copies: 1})
		rest--
	}
	if rest > 0 {
		shares = append(shares, tradeShare{quantity: 0, copies: rest})
	}
	return shares
}

// heldTake is a take from an item held by the container being split apart, with what each copy of the container is
// left holding of it (see tradeShares): one entry for each copy of the container times those of each container between
// it and the item.
type heldTake struct {
	eqp    *Equipment
	shares []tradeShare
}

// copyGroup is a run of copies of a container that the takes it holds leave alike.
type copyGroup struct {
	copies int64
	takes  []heldTake // each take's shares for one copy of the run
}

// groupCopies splits the n copies of a container into runs that the takes it holds leave alike.
func groupCopies(n int64, takes []heldTake) []copyGroup {
	per := make([]int64, len(takes))
	for i, take := range takes {
		per[i] = countShares(take.shares) / n
	}
	var groups []copyGroup
	for i := range n {
		one := make([]heldTake, len(takes))
		for j, take := range takes {
			one[j] = heldTake{eqp: take.eqp, shares: sliceShares(take.shares, i*per[j], (i+1)*per[j])}
		}
		if last := len(groups) - 1; last >= 0 && sameTakes(groups[last].takes, one) {
			groups[last].copies++
		} else {
			groups = append(groups, copyGroup{copies: 1, takes: one})
		}
	}
	return groups
}

// countShares returns how many copies the shares cover.
func countShares(shares []tradeShare) int64 {
	var count int64
	for _, share := range shares {
		count += share.copies
	}
	return count
}

// sliceShares returns the part of the shares covering the copies from from up to, but not including, to.
func sliceShares(shares []tradeShare, from, to int64) []tradeShare {
	var part []tradeShare
	var start int64
	for _, share := range shares {
		end := start + share.copies
		if lo, hi := max(start, from), min(end, to); lo < hi {
			part = append(part, tradeShare{quantity: share.quantity, copies: hi - lo})
		}
		if start = end; start >= to {
			break
		}
	}
	return part
}

func sameTakes(a, b []heldTake) bool {
	if len(a) != len(b) {
		return false
	}
	for i, one := range a {
		if one.eqp != b[i].eqp || !slices.Equal(one.shares, b[i].shares) {
			return false
		}
	}
	return true
}

// splitHeld leaves each copy of the container holding its share of the takes, which it holds at some depth, splitting
// the container apart where its copies are left unalike (see tradeTakes).
func (e *Entity) splitHeld(c *Equipment, takes []heldTake) {
	groups := groupCopies(c.Quantity.AsInteger[int64](), takes)
	holders := make([]*Equipment, len(groups))
	holders[0] = c
	if len(groups) > 1 {
		c.Quantity = fxp.FromInteger(groups[0].copies)
		for i, group := range groups[1:] {
			holder := c.Clone(LibraryFile{}, e, c.Parent(), Duplicate)
			holder.Quantity = fxp.FromInteger(group.copies)
			copies := make(map[*Equipment]*Equipment)
			mapCopies(c, holder, copies)
			for j := range group.takes {
				group.takes[j].eqp = copies[group.takes[j].eqp]
			}
			holders[i+1] = holder
		}
		e.insertAfter(c, holders[1:])
	}
	for i, group := range groups {
		own, containers, below := childTakes(holders[i], group.takes)
		for _, take := range own {
			e.leave(take.eqp, take.shares[0].quantity)
		}
		for _, child := range containers {
			e.splitHeld(child, below[child])
		}
	}
}

// wouldSplit calls split with the container if splitHeld would split it apart, then goes on to the containers below it
// in the copies that keep it, the only ones that exist to be edited.
func wouldSplit(c *Equipment, takes []heldTake, split func(*Equipment)) {
	groups := groupCopies(c.Quantity.AsInteger[int64](), takes)
	if len(groups) > 1 {
		split(c)
	}
	_, containers, below := childTakes(c, groups[0].takes)
	for _, child := range containers {
		wouldSplit(child, below[child], split)
	}
}

// childTakes sorts the takes the container holds into those of its own children and those held by containers among its
// children, the latter by container, in the order first seen.
func childTakes(c *Equipment, takes []heldTake) (own []heldTake, containers []*Equipment, below map[*Equipment][]heldTake) {
	below = make(map[*Equipment][]heldTake)
	for _, take := range takes {
		child := take.eqp
		for child.Parent() != c {
			child = child.Parent()
		}
		if child == take.eqp {
			own = append(own, take)
			continue
		}
		if _, seen := below[child]; !seen {
			containers = append(containers, child)
		}
		below[child] = append(below[child], take)
	}
	return own, containers, below
}

// mapCopies maps the equipment and everything it holds to the same in its copy.
func mapCopies(eqp, cp *Equipment, m map[*Equipment]*Equipment) {
	m[eqp] = cp
	for i, child := range eqp.Children {
		mapCopies(child, cp.Children[i], m)
	}
}

// insertAfter puts the equipment, whose parent must already be eqp's, into the list holding eqp, right after it.
func (e *Entity) insertAfter(eqp *Equipment, list []*Equipment) {
	if parent := eqp.Parent(); parent != nil {
		parent.Children = slices.Insert(parent.Children, slices.Index(parent.Children, eqp)+1, list...)
	} else {
		e.CarriedEquipment = slices.Insert(e.CarriedEquipment, slices.Index(e.CarriedEquipment, eqp)+1, list...)
	}
}

// leave sets the carried equipment's quantity, removing it instead when nothing is left.
func (e *Entity) leave(eqp *Equipment, qty fxp.Int) {
	if qty <= 0 {
		e.removeCarried(eqp)
	} else {
		eqp.Quantity = qty
	}
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
