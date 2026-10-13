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
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// payoutScale converts a fixed-point amount into whole units of its smallest decimal place. fxp keeps four places.
const payoutScale = 10000

// maxPayoutCells limits the size of the table PayoutFor builds. Beyond it, PayoutFor falls back to largest value first.
const maxPayoutCells = 2_000_000

// unreachable marks an amount no combination of money items can make.
const unreachable = -1

// Payout is a quantity of one money item given in payment.
type Payout struct {
	Money    *Equipment
	Quantity int
}

type payoutCoin struct {
	eqp   *Equipment
	value fxp.Int
	units int64
}

// PayoutFor returns the money items, drawn from money in any quantity, whose total value is as close to target as
// possible without exceeding it, along with that total. Where more than one combination reaches that total, the one
// with the fewest items is used. Items without a positive value are ignored.
func PayoutFor(target fxp.Int, money []*Equipment) (payouts []Payout, paid fxp.Int) {
	zero := fxp.FromInteger(0)
	coins := payoutCoins(money)
	if len(coins) == 0 || target <= zero {
		return nil, zero
	}
	scale := fxp.FromInteger(payoutScale)
	var g int64
	for i := range coins {
		coins[i].units = coins[i].value.Mul(scale).AsInteger[int64]()
		g = gcd(g, coins[i].units)
	}
	limit := target.Mul(scale).AsInteger[int64]() / g
	if limit > maxPayoutCells {
		return largestFirstPayout(target, coins)
	}
	for i := range coins {
		coins[i].units /= g
	}
	// count[s] is the fewest items summing to s units, and last[s] is the index of the item added last to reach it.
	count := make([]int32, limit+1)
	last := make([]int32, limit+1)
	best := int64(0)
	for s := int64(1); s <= limit; s++ {
		count[s] = unreachable
		for i, c := range coins {
			if c.units > s || count[s-c.units] == unreachable {
				continue
			}
			if n := count[s-c.units] + 1; count[s] == unreachable || n < count[s] {
				count[s] = n
				last[s] = int32(i)
			}
		}
		if count[s] != unreachable {
			best = s
		}
	}
	return buildPayout(coins, best, last)
}

// payoutCoins returns the money items that have a positive value, largest value first.
func payoutCoins(money []*Equipment) []payoutCoin {
	zero := fxp.FromInteger(0)
	coins := make([]payoutCoin, 0, len(money))
	for _, m := range money {
		if m == nil {
			continue
		}
		if v := m.AdjustedValue(); v > zero {
			coins = append(coins, payoutCoin{eqp: m, value: v})
		}
	}
	slices.SortFunc(coins, func(a, b payoutCoin) int {
		switch {
		case a.value > b.value:
			return -1
		case a.value < b.value:
			return 1
		default:
			return 0
		}
	})
	return coins
}

// buildPayout turns the item chosen for each amount into the payout for best units.
func buildPayout(coins []payoutCoin, best int64, last []int32) (payouts []Payout, paid fxp.Int) {
	quantities := make([]int, len(coins))
	for s := best; s > 0; s -= coins[last[s]].units {
		quantities[last[s]]++
	}
	for i, q := range quantities {
		if q > 0 {
			payouts = append(payouts, Payout{Money: coins[i].eqp, Quantity: q})
			paid = paid.Add(coins[i].value.Mul(fxp.FromInteger(q)))
		}
	}
	return payouts, paid
}

// largestFirstPayout gives the target using the largest items first, then fills what remains. It is used only when the
// target is too large for the exact search.
func largestFirstPayout(target fxp.Int, coins []payoutCoin) (payouts []Payout, paid fxp.Int) {
	zero := fxp.FromInteger(0)
	paid = zero
	remaining := target
	for _, c := range coins {
		if remaining < c.value {
			continue
		}
		n := remaining.Div(c.value).Floor().AsInteger[int]()
		for n > 0 && c.value.Mul(fxp.FromInteger(n)) > remaining {
			n--
		}
		if n > 0 {
			spent := c.value.Mul(fxp.FromInteger(n))
			payouts = append(payouts, Payout{Money: c.eqp, Quantity: n})
			paid = paid.Add(spent)
			remaining = remaining.Sub(spent)
		}
	}
	return payouts, paid
}

func gcd(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	if a < 0 {
		return -a
	}
	return a
}
