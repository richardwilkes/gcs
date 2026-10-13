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

func TestAddSellMoneyRejectsContainersUnsourcedAndDuplicates(t *testing.T) {
	c := check.New(t)
	var s SheetSettings
	silver := librarySourced(newEquipmentItem("Silver Eagle", "20", "0"))
	unsourced := newEquipmentItem("Homebrew Coin", "5", "0")
	container := NewEquipment(nil, nil, true)
	container.Source = silver.Source
	c.True(s.AddSellMoney(silver))
	c.False(s.AddSellMoney(silver))
	c.False(s.AddSellMoney(unsourced))
	c.False(s.AddSellMoney(container))
	c.False(s.AddSellMoney(nil))
	c.Equal([]Source{silver.Source}, s.SellMoney)
}
