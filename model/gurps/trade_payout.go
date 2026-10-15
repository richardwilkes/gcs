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

// maxPayoutCells limits the size of the table PayoutFor and PaymentFor build. Beyond it, they fall back to largest
// value first.
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
	// limit is the most of the item PaymentFor and payoutWithin may use.
	limit int64
}

// PayoutFor returns the money items, drawn from money in any quantity, whose total value is as close to target as
// possible without exceeding it, along with that total. Where more than one combination reaches that total, the one
// with the fewest items is used. Items without a positive value are ignored. A target too large for the exact search
// (see maxPayoutCells) is paid with the largest items first instead, which may fall further short of it and use more
// items (see largestFirstPayout).
func PayoutFor(target fxp.Int, money []*Equipment) (payouts []Payout, paid fxp.Int) {
	zero := fxp.FromInteger(0)
	coins := payoutCoins(money)
	if len(coins) == 0 || target <= zero {
		return nil, zero
	}
	g := toUnits(coins)
	limit := units(target) / g
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

// PaymentFor returns the money items, drawn from money, whose total value is the smallest that is at least target,
// along with that total. Each item may be used no more often than the whole units of its TradeQuantity. Where more than
// one combination reaches that total, the one with the fewest items is used. Items without a positive value or without a
// whole unit to trade are ignored. Returns nil and zero when target isn't positive or no combination reaches it. A
// target too large for the exact search (see maxPayoutCells) is paid with the largest items first instead, which may
// pay more than the smallest total and use more items (see largestFirstPayment).
func PaymentFor(target fxp.Int, money []*Equipment) (payments []Payout, paid fxp.Int) {
	if target <= 0 {
		return nil, 0
	}
	all := payoutCoins(money)
	coins := all[:0]
	for _, c := range all {
		if c.limit = c.eqp.TradeQuantity().Floor().AsInteger[int64](); c.limit > 0 {
			coins = append(coins, c)
		}
	}
	if len(coins) == 0 {
		return nil, 0
	}
	g := toUnits(coins)
	goal := ceilDiv(units(target), g)
	var available int64
	for i := range coins {
		coins[i].units /= g
		// More of an item than reaches the goal by itself is never needed, since one could be left out.
		coins[i].limit = min(coins[i].limit, ceilDiv(goal, coins[i].units))
		available += coins[i].limit * coins[i].units
	}
	if available < goal {
		return nil, 0
	}
	// An item worth the goal or more pays it by itself, and paying anything along with it only pays more, so the least
	// of them is the only one that can be part of the smallest payment. They are set apart so that the table searching
	// the rest is bounded by the largest of the rest rather than by them.
	quantities := make([]int64, len(coins))
	first := 0
	for first < len(coins) && coins[first].units >= goal {
		first++
	}
	rest := coins[first:]
	var restAvailable int64
	for _, c := range rest {
		restAvailable += c.limit * c.units
	}
	if restAvailable >= goal {
		// The smallest total that reaches the goal exceeds it by less than the largest item, or any one item could be
		// left out. The table holds a count of each item for every total, so its size is that range times the items.
		var found []int64
		// The size is compared before it is formed, since a goal near the top of int64 would wrap it.
		if goal > maxPayoutCells/int64(len(rest))-rest[0].units {
			found = largestFirstPayment(rest, goal)
		} else {
			found = exactPayment(rest, goal, goal+rest[0].units)
		}
		var total int64
		for i, q := range found {
			total += q * rest[i].units
		}
		if first == 0 || total < coins[first-1].units {
			copy(quantities[first:], found)
			return payoutsFrom(coins, quantities)
		}
	}
	// Of the coins that pay the goal alone, the least is spent, and the first listed of those worth the same, as the
	// exact search spends the first listed on a tie.
	least := first - 1
	for least > 0 && coins[least-1].units == coins[least].units {
		least--
	}
	quantities[least] = 1
	return payoutsFrom(coins, quantities)
}

// payoutWithin is PayoutFor for coins each of which may be used no more often than its limit. The coins must be
// largest value first, as payoutCoins returns them, and are changed in place.
func payoutWithin(target fxp.Int, coins []payoutCoin) (payouts []Payout, paid fxp.Int) {
	if len(coins) == 0 || target <= 0 {
		return nil, 0
	}
	g := toUnits(coins)
	goal := units(target) / g
	for i := range coins {
		coins[i].units /= g
		coins[i].limit = min(coins[i].limit, goal/coins[i].units)
	}
	if goal > maxPayoutCells/int64(len(coins))-1 {
		return payoutsFrom(coins, largestFirstBoundedPayout(coins, goal))
	}
	return payoutsFrom(coins, exactPayout(coins, goal))
}

// payoutCoins returns the money items that have a positive value, largest value first. Items of the same value keep
// the order they had.
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
	slices.SortStableFunc(coins, func(a, b payoutCoin) int {
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

// units returns the amount in whole units of fxp's smallest decimal place, which is how an fxp.Int is stored, so no
// multiplication that could saturate is needed.
func units(amount fxp.Int) int64 {
	return int64(amount)
}

// toUnits sets the units of each coin to its value in whole units of fxp's smallest decimal place and returns their
// greatest common divisor, of which every amount the coins can make is a multiple.
func toUnits(coins []payoutCoin) int64 {
	var g int64
	for i := range coins {
		coins[i].units = units(coins[i].value)
		g = gcd(g, coins[i].units)
	}
	return g
}

// buildPayout turns the item chosen for each amount into the payout for best units.
func buildPayout(coins []payoutCoin, best int64, last []int32) (payouts []Payout, paid fxp.Int) {
	quantities := make([]int64, len(coins))
	for s := best; s > 0; s -= coins[last[s]].units {
		quantities[last[s]]++
	}
	return payoutsFrom(coins, quantities)
}

// payoutsFrom returns a payout for each coin with a positive quantity, along with their total value.
func payoutsFrom(coins []payoutCoin, quantities []int64) (payouts []Payout, paid fxp.Int) {
	for i, q := range quantities {
		if q > 0 {
			payouts = append(payouts, Payout{Money: coins[i].eqp, Quantity: int(q)})
			paid = fxp.SaturatingAdd(paid, coins[i].value.Mul(fxp.FromInteger(q)))
		}
	}
	return payouts, paid
}

// exactPayment returns the quantity of each coin, up to its limit, for the smallest total of at least goal units, using
// the fewest coins for that total and, where those can be made up more than one way, as many of the larger coins as
// they can (see boundedCounts). The total must be below size.
func exactPayment(coins []payoutCoin, goal, size int64) []int64 {
	count, used := boundedCounts(coins, size)
	best := goal
	for count[best] == unreachable {
		best++
	}
	return unwindCoins(coins, used, best)
}

// exactPayout returns the quantity of each coin, up to its limit, for the largest total of at most target units, using
// the fewest coins for that total and, where those can be made up more than one way, as many of the larger coins as
// they can (see boundedCounts). Every quantity is zero when no coin fits.
func exactPayout(coins []payoutCoin, target int64) []int64 {
	count, used := boundedCounts(coins, target+1)
	best := target
	for count[best] == unreachable {
		best--
	}
	return unwindCoins(coins, used, best)
}

// boundedCounts returns, for every total of units below size, the fewest coins summing to it, each used no more often
// than its limit, or unreachable, along with how many of each coin that takes.
func boundedCounts(coins []payoutCoin, size int64) (count []int32, used [][]int32) {
	// count[s] is the fewest coins summing to s units from the coins taken so far, smallest first, and used[i][s] is
	// how many of coin i that takes. Taking coin i, count[s] becomes the least of count[s-k*units]+k for k up to its
	// limit, which is the least of count[r+t*units]-t over a sliding window of t, plus j, for s = r+j*units. A window
	// that keeps the earliest of equal candidates takes as many of the coin as it can on a tie.
	type candidate struct {
		j, key int64
	}
	count = make([]int32, size)
	for s := range count {
		count[s] = unreachable
	}
	count[0] = 0
	used = make([][]int32, len(coins))
	var window []candidate
	for i := len(coins) - 1; i >= 0; i-- {
		c := coins[i]
		used[i] = make([]int32, size)
		for r := range min(c.units, size) {
			window = window[:0]
			head := 0
			for j, s := int64(0), r; s < size; j, s = j+1, s+c.units {
				if count[s] != unreachable {
					key := int64(count[s]) - j
					for len(window) > head && window[len(window)-1].key > key {
						window = window[:len(window)-1]
					}
					window = append(window, candidate{j: j, key: key})
				}
				for len(window) > head && window[head].j < j-c.limit {
					head++
				}
				if len(window) > head {
					count[s] = int32(window[head].key + j)
					used[i][s] = int32(j - window[head].j)
				}
			}
		}
	}
	return count, used
}

// unwindCoins returns the quantity of each coin that boundedCounts found for the total. The largest coin was taken
// last, so the coins are unwound largest first.
func unwindCoins(coins []payoutCoin, used [][]int32, total int64) []int64 {
	quantities := make([]int64, len(coins))
	for i := range coins {
		quantities[i] = int64(used[i][total])
		total -= quantities[i] * coins[i].units
	}
	return quantities
}

// largestFirstBoundedPayout returns the quantity of each coin, up to its limit, for a total of at most target units,
// taking as many of each coin as fit within what remains, largest first. It is used only when target is too large for
// the exact search, and need not find the closest total nor the fewest coins.
func largestFirstBoundedPayout(coins []payoutCoin, target int64) []int64 {
	quantities := make([]int64, len(coins))
	remaining := target
	for i, c := range coins {
		quantities[i] = min(c.limit, remaining/c.units)
		remaining -= quantities[i] * c.units
	}
	return quantities
}

// largestFirstPayment returns the quantity of each coin, up to its limit, for a total of at least goal units. It takes
// as many of each coin as fit within what remains, largest first, then one more of the smallest coin with any left to
// cover a shortfall, and finally leaves out as many of the smallest coins as the total can spare. It is used only when
// goal is too large for the exact search, and need not find the smallest total nor the fewest coins: given one 25 and
// two 20s, it pays 40 with 25+20 rather than 20+20. The coins must be able to reach goal.
func largestFirstPayment(coins []payoutCoin, goal int64) []int64 {
	quantities := make([]int64, len(coins))
	remaining := goal
	for i, c := range coins {
		quantities[i] = min(c.limit, remaining/c.units)
		remaining -= quantities[i] * c.units
	}
	if remaining > 0 {
		// Every coin with any left is larger than what remains, so one more of the smallest of them covers it.
		for i := len(coins) - 1; i >= 0; i-- {
			if quantities[i] < coins[i].limit {
				quantities[i]++
				remaining -= coins[i].units
				break
			}
		}
		for i := len(coins) - 1; i >= 0 && remaining < 0; i-- {
			spare := min(quantities[i], -remaining/coins[i].units)
			quantities[i] -= spare
			remaining += spare * coins[i].units
		}
	}
	return quantities
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

// ceilDiv returns a divided by b, rounded up. Both must be positive. The remainder is tested rather than b-1 added to
// a, which would wrap for an a near the top of int64, such as the units of a saturated fxp.Int.
func ceilDiv(a, b int64) int64 {
	q := a / b
	if a%b != 0 {
		q++
	}
	return q
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
