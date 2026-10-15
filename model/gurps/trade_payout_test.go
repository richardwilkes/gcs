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
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

func librarySourced(eqp *Equipment) *Equipment {
	lib := LibraryFile{Library: "Basic Set", Path: "Basic Set Equipment.eqp"}
	eqp.Source = Source{LibraryFile: lib, TID: eqp.TID}
	return eqp
}

func TestPayoutForIsExact(t *testing.T) {
	silver := newEquipmentItem("Silver Eagle", "20", "0")
	copper := newEquipmentItem("Copper Penny", "1", "0")
	fifteen := newEquipmentItem("Fifteen", "15", "0")
	tests := []struct {
		name     string
		money    []*Equipment
		target   string
		expected []Payout
		paid     string
	}{
		{
			name:     "exact with both",
			money:    []*Equipment{copper, silver},
			target:   "45",
			expected: []Payout{{Money: silver, Quantity: 2}, {Money: copper, Quantity: 5}},
			paid:     "45",
		},
		{
			name:     "fewest items with largest first",
			money:    []*Equipment{copper, silver},
			target:   "31",
			expected: []Payout{{Money: silver, Quantity: 1}, {Money: copper, Quantity: 11}},
			paid:     "31",
		},
		{
			name:     "closest without exceeding",
			money:    []*Equipment{copper, silver},
			target:   "239.20",
			expected: []Payout{{Money: silver, Quantity: 11}, {Money: copper, Quantity: 19}},
			paid:     "239",
		},
		{
			name:     "non-canonical denominations reach an exact amount",
			money:    []*Equipment{silver, fifteen},
			target:   "30",
			expected: []Payout{{Money: fifteen, Quantity: 2}},
			paid:     "30",
		},
		{
			name:     "non-canonical denominations closest without exceeding",
			money:    []*Equipment{silver, fifteen},
			target:   "31",
			expected: []Payout{{Money: fifteen, Quantity: 2}},
			paid:     "30",
		},
		{
			name:     "three denominations prefer fewest items over largest first",
			money:    []*Equipment{copper, silver, fifteen},
			target:   "31",
			expected: []Payout{{Money: fifteen, Quantity: 2}, {Money: copper, Quantity: 1}},
			paid:     "31",
		},
		{
			name:     "less than one item",
			money:    []*Equipment{copper, silver},
			target:   "0.50",
			expected: nil,
			paid:     "0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := check.New(t)
			payouts, paid := PayoutFor(fxp.FromStringForced(tt.target), tt.money)
			tc.Equal(tt.expected, payouts)
			tc.Equal(fxp.FromStringForced(tt.paid), paid)
		})
	}
}

func TestPayoutForIgnoresNilAndWorthlessItems(t *testing.T) {
	c := check.New(t)
	silver := newEquipmentItem("Silver Eagle", "20", "0")
	worthless := newEquipmentItem("Trinket", "0", "0")
	payouts, paid := PayoutFor(fxp.FromInteger(40), []*Equipment{nil, worthless, silver})
	c.Equal([]Payout{{Money: silver, Quantity: 2}}, payouts)
	c.Equal(fxp.FromInteger(40), paid)
}

// TestAddSellMoneyRejectsWhatCanNotBeMoney checks that only library equipment tagged as money (see MoneyTag) that isn't
// a container and isn't already listed can be sell money, since buying spends only money so tagged, and sell money that
// buying can't spend would be dead weight.
func TestAddSellMoneyRejectsWhatCanNotBeMoney(t *testing.T) {
	c := check.New(t)
	var s SheetSettings
	silver := librarySourced(denomination("Silver Eagle", "20"))
	unsourced := denomination("Homebrew Coin", "5")
	untagged := librarySourced(newEquipmentItem("Gold Crown", "100", "0"))
	container := NewEquipment(nil, nil, true)
	container.Source = silver.Source
	container.Tags = []string{MoneyTag}
	c.True(s.AddSellMoney(silver))
	c.False(s.AddSellMoney(silver))
	c.False(s.AddSellMoney(unsourced))
	c.False(s.AddSellMoney(untagged), "an item that isn't tagged as money can't be spent and so isn't sell money")
	c.False(s.AddSellMoney(container))
	c.False(s.AddSellMoney(nil))
	c.Equal([]Source{silver.Source}, s.SellMoney)
}

func moneyItem(name, value, quantity string) *Equipment {
	eqp := newEquipmentItem(name, value, "0")
	eqp.Quantity = fxp.FromStringForced(quantity)
	return eqp
}

func TestPaymentFor(t *testing.T) {
	silver := moneyItem("Silver Eagle", "20", "5")
	copper := moneyItem("Copper Penny", "1", "20")
	fifteen := moneyItem("Fifteen", "15", "2")
	lone := moneyItem("Lone Eagle", "20", "1")
	threePennies := moneyItem("Copper Penny", "1", "3")
	halfPennies := moneyItem("Copper Penny", "1", "2.5")
	fiver := moneyItem("Fiver", "5", "1")
	sliver := moneyItem("Sliver", "0.5", "0.75")
	worthless := moneyItem("Trinket", "0", "10")
	tests := []struct {
		name     string
		money    []*Equipment
		target   string
		expected []Payout
		paid     string
	}{
		{
			name:     "exact cover",
			money:    []*Equipment{copper, silver},
			target:   "45",
			expected: []Payout{{Money: silver, Quantity: 2}, {Money: copper, Quantity: 5}},
			paid:     "45",
		},
		{
			name:     "smallest overpayment",
			money:    []*Equipment{threePennies, lone},
			target:   "15",
			expected: []Payout{{Money: lone, Quantity: 1}},
			paid:     "20",
		},
		{
			name:     "overpays a fraction",
			money:    []*Equipment{copper, silver},
			target:   "39.20",
			expected: []Payout{{Money: silver, Quantity: 2}},
			paid:     "40",
		},
		{
			name:     "fewest items for the smallest total",
			money:    []*Equipment{copper, lone, fifteen},
			target:   "31",
			expected: []Payout{{Money: fifteen, Quantity: 2}, {Money: copper, Quantity: 1}},
			paid:     "31",
		},
		{
			name:     "smallest total before fewest items",
			money:    []*Equipment{copper, silver},
			target:   "19",
			expected: []Payout{{Money: copper, Quantity: 19}},
			paid:     "19",
		},
		{
			name:     "larger items on a tie",
			money:    []*Equipment{moneyItem("Four", "4", "5"), moneyItem("Five", "5", "5"), moneyItem("Six", "6", "5")},
			target:   "10",
			expected: nil, // filled in below
			paid:     "10",
		},
		{
			name:     "quantities limit the items used",
			money:    []*Equipment{threePennies, fiver},
			target:   "5",
			expected: []Payout{{Money: fiver, Quantity: 1}},
			paid:     "5",
		},
		{
			name:     "quantities limit the items used to overpay",
			money:    []*Equipment{threePennies, lone},
			target:   "4",
			expected: []Payout{{Money: lone, Quantity: 1}},
			paid:     "20",
		},
		{
			name:   "too few to reach",
			money:  []*Equipment{threePennies},
			target: "5",
			paid:   "0",
		},
		{
			name:     "fractional quantities are floored",
			money:    []*Equipment{halfPennies, fiver},
			target:   "3",
			expected: []Payout{{Money: fiver, Quantity: 1}},
			paid:     "5",
		},
		{
			name:   "fractional quantities are floored and can't reach",
			money:  []*Equipment{halfPennies},
			target: "3",
			paid:   "0",
		},
		{
			name:   "less than one of an item is ignored",
			money:  []*Equipment{sliver},
			target: "0.25",
			paid:   "0",
		},
		{
			name:   "nil and worthless items are ignored",
			money:  []*Equipment{nil, worthless},
			target: "1",
			paid:   "0",
		},
		{
			name:   "zero target",
			money:  []*Equipment{copper},
			target: "0",
			paid:   "0",
		},
		{
			name:   "negative target",
			money:  []*Equipment{copper},
			target: "-5",
			paid:   "0",
		},
		{
			name:   "no money",
			target: "5",
			paid:   "0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := check.New(t)
			expected := tt.expected
			if tt.name == "larger items on a tie" {
				// 6+4 and 5+5 both pay 10 with two items; the one with the larger item wins.
				expected = []Payout{{Money: tt.money[2], Quantity: 1}, {Money: tt.money[0], Quantity: 1}}
			}
			payments, paid := PaymentFor(fxp.FromStringForced(tt.target), tt.money)
			tc.Equal(expected, payments)
			tc.Equal(fxp.FromStringForced(tt.paid), paid)
		})
	}
}

// TestPaymentForEqualItemsUseTheFirstListed verifies that two stacks of the same value are drawn on in the order given.
func TestPaymentForEqualItemsUseTheFirstListed(t *testing.T) {
	c := check.New(t)
	purse := moneyItem("Copper Penny", "1", "3")
	pouch := moneyItem("Copper Penny", "1", "10")
	payments, paid := PaymentFor(fxp.FromInteger(5), []*Equipment{purse, pouch})
	c.Equal([]Payout{{Money: purse, Quantity: 3}, {Money: pouch, Quantity: 2}}, payments)
	c.Equal(fxp.FromInteger(5), paid)
}

// TestPaymentForFallsBackForLargeTargets verifies the largest-first search used when the exact one's table would be too
// large: it still covers the target where the items can, leaving out the smaller items the total can spare.
func TestPaymentForFallsBackForLargeTargets(t *testing.T) {
	tests := []struct {
		name     string
		money    []*Equipment
		target   string
		expected []int
		paid     string
	}{
		{
			name:     "covers the remainder with the smallest item",
			money:    []*Equipment{moneyItem("Gold", "1000", "3000"), moneyItem("Copper", "1", "5000000")},
			target:   "3000500.5",
			expected: []int{3000, 501},
			paid:     "3000501",
		},
		{
			name:     "overshoots with a larger item and drops the smaller ones",
			money:    []*Equipment{moneyItem("Crown", "1000000", "5"), moneyItem("Copper", "1", "10")},
			target:   "2500000",
			expected: []int{3, 0},
			paid:     "3000000",
		},
		{
			name:     "is exact where it can be",
			money:    []*Equipment{moneyItem("Crown", "1000000", "5"), moneyItem("Copper", "1", "3000000")},
			target:   "2500000",
			expected: []int{2, 500000},
			paid:     "2500000",
		},
		{
			// The exact search would pay 40000000 with two of the 20000000 items.
			name: "need not find the smallest total",
			money: []*Equipment{
				moneyItem("Big", "25000000", "1"), moneyItem("Medium", "20000000", "2"),
				moneyItem("Copper", "1", "1"),
			},
			target:   "40000000",
			expected: []int{1, 1, 0},
			paid:     "45000000",
		},
		{
			name:     "can't reach",
			money:    []*Equipment{moneyItem("Crown", "1000000", "2"), moneyItem("Copper", "1", "10")},
			target:   "2500000",
			expected: nil,
			paid:     "0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := check.New(t)
			var expected []Payout
			for i, q := range tt.expected {
				if q > 0 {
					expected = append(expected, Payout{Money: tt.money[i], Quantity: q})
				}
			}
			payments, paid := PaymentFor(fxp.FromStringForced(tt.target), tt.money)
			tc.Equal(expected, payments)
			tc.Equal(fxp.FromStringForced(tt.paid), paid)
		})
	}
}

// TestPaymentForMatchesExhaustiveSearch checks the exact search against every combination of a few small items.
func TestPaymentForMatchesExhaustiveSearch(t *testing.T) {
	c := check.New(t)
	rng := rand.New(rand.NewPCG(1, 2)) //nolint:gosec // A fixed seed keeps the cases reproducible.
	for range 500 {
		money := make([]*Equipment, 1+rng.IntN(4))
		values := make([]int, len(money))
		limits := make([]int, len(money))
		for i := range money {
			values[i] = 1 + rng.IntN(12)
			limits[i] = rng.IntN(5)
			money[i] = moneyItem("Coin", fxp.FromInteger(values[i]).String(), fxp.FromInteger(limits[i]).String())
		}
		target := 1 + rng.IntN(40)
		bestTotal, bestCount := -1, 0
		quantities := make([]int, len(money))
		var search func(i, total, count int)
		search = func(i, total, count int) {
			if i == len(money) {
				if total >= target && (bestTotal == -1 || total < bestTotal || (total == bestTotal && count < bestCount)) {
					bestTotal, bestCount = total, count
				}
				return
			}
			for q := 0; q <= limits[i]; q++ {
				quantities[i] = q
				search(i+1, total+q*values[i], count+q)
			}
		}
		search(0, 0, 0)
		payments, paid := PaymentFor(fxp.FromInteger(target), money)
		if bestTotal == -1 {
			c.Nil(payments, "values %v, limits %v, target %d", values, limits, target)
			c.Equal(fxp.FromInteger(0), paid)
			continue
		}
		count := 0
		total := fxp.FromInteger(0)
		for _, one := range payments {
			i := slices.Index(money, one.Money)
			c.True(one.Quantity <= limits[i], "values %v, limits %v, target %d", values, limits, target)
			count += one.Quantity
			total += fxp.FromInteger(values[i] * one.Quantity)
		}
		c.Equal(fxp.FromInteger(bestTotal), paid, "values %v, limits %v, target %d", values, limits, target)
		c.Equal(paid, total)
		c.Equal(bestCount, count, "values %v, limits %v, target %d", values, limits, target)
	}
}

// TestPayoutsHandleHugeAmounts checks that amounts and values large enough to saturate a fixed-point multiplication are
// still turned into units exactly, so huge targets are reached and huge coins are counted.
func TestPayoutsHandleHugeAmounts(t *testing.T) {
	c := check.New(t)
	bond := moneyItem("Bond", "1000000000", "1000")
	payments, paid := PaymentFor(fxp.FromInteger(200_000_000_000), []*Equipment{bond})
	c.Equal([]Payout{{Money: bond, Quantity: 200}}, payments)
	c.Equal(fxp.FromInteger(200_000_000_000), paid)

	payouts, paid := PayoutFor(fxp.FromInteger(500_000_000_000), []*Equipment{bond})
	c.Equal([]Payout{{Money: bond, Quantity: 500}}, payouts)
	c.Equal(fxp.FromInteger(500_000_000_000), paid)

	penny := moneyItem("Copper Penny", "1", "1")
	payments, paid = PaymentFor(fxp.FromInteger(300_000_000_000), []*Equipment{bond, penny})
	c.Equal([]Payout{{Money: bond, Quantity: 300}}, payments)
	c.Equal(fxp.FromInteger(300_000_000_000), paid)
	payments, paid = PaymentFor(fxp.FromInteger(2_000_000_000_000), []*Equipment{bond})
	c.Nil(payments, "the bonds can't reach two trillion")
	c.Equal(fxp.Int(0), paid)

	e := NewEntity()
	change, paid := e.ChangeFor(fxp.FromInteger(400_000_000_000), nil, []Payout{{Money: bond, Quantity: 500}})
	c.Equal(1, len(change))
	if len(change) == 1 {
		c.Equal(400, change[0].Quantity)
	}
	c.Equal(fxp.FromInteger(400_000_000_000), paid)
}

// TestPaymentForSetsApartCoinsThatPayTheTargetAlone checks that a money item worth the target or more, which can only
// ever be the whole payment, doesn't force the approximate search on the rest, and that it is still paid when it is the
// smallest payment or ties the smallest with fewer items.
func TestPaymentForSetsApartCoinsThatPayTheTargetAlone(t *testing.T) {
	c := check.New(t)
	gem := moneyItem("Gem", "1000000", "1")
	twentyFive := moneyItem("Twenty-Five", "25", "1")
	tens := moneyItem("Ten", "10", "3")
	cent := moneyItem("Cent", "0.01", "1")
	payments, paid := PaymentFor(fxp.FromInteger(30), []*Equipment{gem, twentyFive, tens, cent})
	c.Equal([]Payout{{Money: tens, Quantity: 3}}, payments, "the gem doesn't keep the rest from being searched exactly")
	c.Equal(fxp.FromInteger(30), paid)

	thirtyOne := moneyItem("Thirty-One", "31", "1")
	payments, paid = PaymentFor(fxp.FromInteger(30), []*Equipment{thirtyOne, twentyFive, twentyFive})
	c.Equal([]Payout{{Money: thirtyOne, Quantity: 1}}, payments, "a coin that overpays by less than the rest is paid")
	c.Equal(fxp.FromInteger(31), paid)

	thirty := moneyItem("Thirty", "30", "1")
	payments, paid = PaymentFor(fxp.FromInteger(30), []*Equipment{tens, thirty})
	c.Equal([]Payout{{Money: thirty, Quantity: 1}}, payments, "one coin beats three that pay the same")
	c.Equal(fxp.FromInteger(30), paid)

	payments, paid = PaymentFor(fxp.FromInteger(30), []*Equipment{gem, cent})
	c.Equal([]Payout{{Money: gem, Quantity: 1}}, payments, "the gem pays when the rest can't reach the target")
	c.Equal(fxp.FromInteger(1_000_000), paid)
}

// TestPayoutSearchesTakeTheMaximumTarget checks that a target at fxp's maximum, which a saturated total comes to, is
// searched without the sizes of the tables wrapping around: the payment reports that the money falls short, the payout
// pays out as far as the money reaches, and so does the change, even when the money's unit is fxp's smallest.
func TestPayoutSearchesTakeTheMaximumTarget(t *testing.T) {
	c := check.New(t)
	silver := moneyItem("Silver Eagle", "20", "5")
	copper := moneyItem("Copper Penny", "1", "20")
	sliver := moneyItem("Sliver", "0.0001", "1")
	payments, paid := PaymentFor(fxp.Max, []*Equipment{silver})
	c.Nil(payments)
	c.Equal(fxp.Int(0), paid)
	payments, paid = PaymentFor(fxp.Max, []*Equipment{silver, copper})
	c.Nil(payments)
	c.Equal(fxp.Int(0), paid)
	payouts, paid := PayoutFor(fxp.Max, []*Equipment{silver, copper})
	c.Equal(2, len(payouts))
	c.True(paid > 0 && paid <= fxp.Max)
	e := NewEntity()
	change, paid := e.ChangeFor(fxp.Max, []*Equipment{sliver}, []Payout{{Money: copper, Quantity: 3}})
	c.True(len(change) >= 1, "the slivers are paid out as far as they reach")
	c.True(paid > 0 && paid <= fxp.Max)
}

// TestPaymentForSingleCoinUsesTheFirstListedOfEqualValue checks that where one coin pays the target by itself and
// several equal-valued coins could, the first listed is spent, as the exact search spends the first listed on a tie.
func TestPaymentForSingleCoinUsesTheFirstListedOfEqualValue(t *testing.T) {
	c := check.New(t)
	purse := moneyItem("Purse Gold", "100", "1")
	pouch := moneyItem("Pouch Gold", "100", "1")
	payments, paid := PaymentFor(fxp.Fifty, []*Equipment{purse, pouch})
	c.Equal([]Payout{{Money: purse, Quantity: 1}}, payments)
	c.Equal(fxp.Hundred, paid)
	crown := moneyItem("Crown", "500", "1")
	payments, paid = PaymentFor(fxp.Fifty, []*Equipment{crown, purse, pouch})
	c.Equal([]Payout{{Money: purse, Quantity: 1}}, payments, "the least of the coins that pay the target alone is spent")
	c.Equal(fxp.Hundred, paid)
}
