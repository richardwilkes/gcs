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
	"fmt"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestPointsRangeForCountPicker verifies the range of a container that asks for a number of its children to be
// picked, for each way that number can be constrained. The cases are grouped by what the container offers: choices
// that cost points, choices that return them, and both at once.
func TestPointsRangeForCountPicker(t *testing.T) {
	c := check.New(t)

	// Advantages: every choice costs points.
	checkPickerCases(c, picker.Count, []pickerCase{
		// Pick exactly 1 of [10, 10]: every way of choosing costs the same, so there is nothing left to decide.
		{
			"a count picker whose children all cost the same is worth one of them",
			criteria.EqualsNumber, 1,
			[]int{10, 10},
			10, 10, true,
		},

		// Pick exactly 2 of [5, 10, 20, 40]: the 2 cheapest are 5 + 10, the 2 costliest are 40 + 20.
		{
			"picking an exact count spans the cheapest and costliest ways to make it",
			criteria.EqualsNumber, 2,
			[]int{5, 10, 20, 40},
			15, 60, false,
		},

		// Pick any number of [5, 10, 20]: taking nothing is cheapest, taking everything is costliest.
		{
			"an unconstrained count over advantages spans nothing to everything",
			criteria.AnyNumber, 0,
			[]int{5, 10, 20},
			0, 35, false,
		},

		// Pick at least 2 of [5, 10, 20]: the cheapest pair is 5 + 10, and nothing stops everything being taken.
		{
			"a minimum count over advantages starts at the cheapest way to make it",
			criteria.AtLeastNumber, 2,
			[]int{5, 10, 20},
			15, 35, false,
		},

		// Pick at most 2 of [5, 10, 20]: nothing need be taken, and the 2 costliest are 20 + 10.
		{
			"a maximum count over advantages limits how much can be taken",
			criteria.AtMostNumber, 2,
			[]int{5, 10, 20},
			0, 30, false,
		},

		// A count larger than the number of children cannot ask for more than there is.
		{
			"a count exceeding the child count takes every child",
			criteria.EqualsNumber, 9,
			[]int{10, 20},
			30, 30, true,
		},

		// A container that presents a choice among nothing costs nothing.
		{
			"a picker with no children costs nothing",
			criteria.EqualsNumber, 1,
			nil,
			0, 0, true,
		},
	})

	// Disadvantages: every choice returns points.
	checkPickerCases(c, picker.Count, []pickerCase{
		// Pick exactly 1 of [-10, -10]: every way of choosing returns the same, so there is nothing left to decide.
		{
			"a count picker whose disadvantages all return the same is worth one of them",
			criteria.EqualsNumber, 1,
			[]int{-10, -10},
			-10, -10, true,
		},

		// Pick exactly 2 of [-5, -10, -20, -40]: the 2 cheapest are -40 + -20, the 2 costliest are -5 + -10.
		{
			"picking an exact count of disadvantages spans the cheapest and costliest ways to make it",
			criteria.EqualsNumber, 2,
			[]int{-5, -10, -20, -40},
			-60, -15, false,
		},

		// Pick any number of [-5, -10, -20]: taking everything is cheapest, taking nothing is costliest.
		{
			"an unconstrained count over disadvantages spans everything to nothing",
			criteria.AnyNumber, 0,
			[]int{-5, -10, -20},
			-35, 0, false,
		},

		// Pick at least 2 of [-5, -10, -20]: everything may be taken at the low end, and the 2 costliest are -5 + -10.
		{
			"a minimum count of disadvantages has to be taken even though taking none would cost more",
			criteria.AtLeastNumber, 2,
			[]int{-5, -10, -20},
			-35, -15, false,
		},

		// Pick at most 2 of [-5, -10, -20]: the 2 cheapest are -20 + -10, and nothing need be taken at the high end.
		{
			"a maximum count limits how many disadvantages can be taken",
			criteria.AtMostNumber, 2,
			[]int{-5, -10, -20},
			-30, 0, false,
		},

		// A count larger than the number of children cannot ask for more than there is.
		{
			"a count exceeding the child count takes every disadvantage",
			criteria.EqualsNumber, 9,
			[]int{-10, -20},
			-30, -30, true,
		},
	})

	// Both at once, which is the everyday shape of a "pick 1": the racial "Choose One" lists offer variants that are
	// not all on the same side of 0, and the unconstrained "optional traits" lists do the same. Under any other count
	// it is rare -- 15 of the 1,126 pickers in the master library mix the two signs at all, and outside those two
	// shapes they are single digits. It is also where the two ends stop being the obvious ones: each is reached by a
	// selection that takes what helps it and no more, except where the count forces something unwelcome along.
	checkPickerCases(c, picker.Count, []pickerCase{
		// Pick any number of [5, 10, -20]: taking nothing but the disadvantage is cheapest, taking both advantages is
		// costliest.
		{
			"an unconstrained count takes every disadvantage at the low end and every advantage at the high end",
			criteria.AnyNumber, 0,
			[]int{5, 10, -20},
			-20, 15, false,
		},

		// Pick exactly 1 of [10, 20, -5]: one choice is one choice, so each end is simply the cheapest and costliest
		// child.
		{
			"an exact count of one spans the cheapest and costliest child",
			criteria.EqualsNumber, 1,
			[]int{10, 20, -5},
			-5, 20, false,
		},

		// Pick exactly 2 of [10, 20, 40, -5]: the cheapest pair pairs the disadvantage with the cheapest advantage, the
		// costliest pair takes the two costliest advantages.
		{
			"an exact count spans the cheapest and costliest pairs",
			criteria.EqualsNumber, 2,
			[]int{10, 20, 40, -5},
			5, 60, false,
		},

		// Pick exactly 2 of [5, 10, -20]: the disadvantage lands in the cheapest pair and is left out of the costliest.
		{
			"an exact count leaves the disadvantage out of the costliest way to make it",
			criteria.EqualsNumber, 2,
			[]int{5, 10, -20},
			-15, 15, false,
		},

		// Pick exactly 2 of [5, -5]: there is only one pair to take, so the choice is settled even though the children
		// differ.
		{
			"an exact count matching the child count is settled whatever the children cost",
			criteria.EqualsNumber, 2,
			[]int{5, -5},
			0, 0, true,
		},

		// Pick exactly 3 of [5, 10, -20]: likewise, taking everything is the only way to make the count.
		{
			"an exact count of every child is worth the total of them",
			criteria.EqualsNumber, 3,
			[]int{5, 10, -20},
			-5, -5, true,
		},

		// A count larger than the number of children cannot ask for more than there is.
		{
			"a count exceeding the child count takes every child",
			criteria.EqualsNumber, 9,
			[]int{10, -5},
			5, 5, true,
		},

		// Pick at least 1 of [10, -5]: the disadvantage alone is cheapest, and the advantage alone is costliest, since
		// beyond the one required nothing unwelcome need be taken at either end.
		{
			"beyond the required count, only what helps each end is taken",
			criteria.AtLeastNumber, 1,
			[]int{10, -5},
			-5, 10, false,
		},

		// Pick at least 2 of [10, 20, -5]: the cheapest pair is -5 + 10, with nothing left worth taking; the costliest
		// pair is 20 + 10, again with nothing left worth taking.
		{
			"a minimum count has to be taken even when what is left would be better",
			criteria.AtLeastNumber, 2,
			[]int{10, 20, -5},
			5, 30, false,
		},

		// Pick at least 2 of [10, 20, 40, -5]: past the required two, every remaining advantage is still worth taking at
		// the high end.
		{
			"a minimum count does not cap what may be taken beyond it",
			criteria.AtLeastNumber, 2,
			[]int{10, 20, 40, -5},
			5, 70, false,
		},

		// Pick at least 1 of [20, -10, -10]: only one child need be taken, but nothing stops the other disadvantage
		// being taken too, so the low end is both of them rather than the single cheapest child.
		{
			"a minimum count is a floor on how many are taken, not a ceiling",
			criteria.AtLeastNumber, 1,
			[]int{20, -10, -10},
			-20, 20, false,
		},

		// Pick at least 2 of [10, -5, -20]: two must be taken, so the high end cannot be the advantage alone -- the
		// cheaper of the two disadvantages comes along with it.
		{
			"a minimum count drags a disadvantage into the costliest selection when there are too few advantages",
			criteria.AtLeastNumber, 2,
			[]int{10, -5, -20},
			-25, 5, false,
		},

		// Pick at least 2 of [5, -5]: both must be taken, which settles a picker whose children differ.
		{
			"a minimum count matching the child count is settled",
			criteria.AtLeastNumber, 2,
			[]int{5, -5},
			0, 0, true,
		},

		// Pick at least 3 of [5, 10, -5, -10]: three of the four must be taken, so the costliest selection gives up the
		// cheaper disadvantage but must still carry the other.
		{
			"a minimum count short of every child still forces one of the disadvantages",
			criteria.AtLeastNumber, 3,
			[]int{5, 10, -5, -10},
			-10, 10, false,
		},

		// Pick at most 1 of [10, -5]: nothing need be taken, but neither end gains by taking nothing.
		{
			"a maximum count of one spans the cheapest and costliest child",
			criteria.AtMostNumber, 1,
			[]int{10, -5},
			-5, 10, false,
		},

		// Pick at most 2 of [10, 20, 40, -5]: the -5 alone is cheapest, and the 2 costliest are 40 + 20.
		{
			"a maximum count limits how much can be taken at either end",
			criteria.AtMostNumber, 2,
			[]int{10, 20, 40, -5},
			-5, 60, false,
		},

		// Pick at most 2 of [5, 10, -20]: the cap is above what either end wants to take, so it binds neither.
		{
			"a maximum count allowing more than is worth taking binds neither end",
			criteria.AtMostNumber, 2,
			[]int{5, 10, -20},
			-20, 15, false,
		},

		// Pick at most 0 of [10, -5]: nothing may be taken, so nothing is what it costs.
		{
			"a maximum count of none costs nothing",
			criteria.AtMostNumber, 0,
			[]int{10, -5},
			0, 0, true,
		},

		// The cases above keep the count close to the number of children, where nearly everything on offer is forced.
		// These two lists are 3x and 2x the count they are picked from, so most of what is on offer is left behind and
		// each end is free to take only what suits it.

		// Any number of [40, 20, 10, -5, -10, -30]: every disadvantage at the low end, every advantage at the high.
		{
			"an unconstrained count over a long list still splits it by sign",
			criteria.AnyNumber, 0,
			[]int{40, 20, 10, -5, -10, -30},
			-45, 70, false,
		},

		// Pick exactly 2 of [40, 20, 10, -5, -10, -30]: the 2 costliest disadvantages, or the 2 costliest advantages,
		// with the other 4 children untouched at either end.
		{
			"an exact count far short of the child count takes only the extremes",
			criteria.EqualsNumber, 2,
			[]int{40, 20, 10, -5, -10, -30},
			-40, 60, false,
		},

		// Pick at least 2 of [40, 20, 10, -5, -10, -30]: the 2 required are already covered by the advantages at the
		// high end and by the disadvantages at the low end, so the count binds neither and both ends take all of
		// what suits them.
		{
			"a minimum count a list this long already meets binds neither end",
			criteria.AtLeastNumber, 2,
			[]int{40, 20, 10, -5, -10, -30},
			-45, 70, false,
		},

		// Pick at most 2 of [40, 20, 10, -5, -10, -30]: here the cap does bind, since both ends would take 3 of the
		// 6 children if allowed to.
		{
			"a maximum count well short of the child count binds both ends",
			criteria.AtMostNumber, 2,
			[]int{40, 20, 10, -5, -10, -30},
			-40, 60, false,
		},

		// Any number of [30, 20, 10, 5, -5, -15]: the same split by sign, over a list with 4 advantages and 2
		// disadvantages.
		{
			"an unconstrained count is unaffected by how the list is made up",
			criteria.AnyNumber, 0,
			[]int{30, 20, 10, 5, -5, -15},
			-20, 65, false,
		},

		// Pick exactly 3 of [30, 20, 10, 5, -5, -15]: the cheapest 3 are both disadvantages plus the cheapest
		// advantage, since there is no third disadvantage to take; the costliest 3 are the 3 costliest advantages.
		{
			"an exact count reaches past the disadvantages when there are too few of them",
			criteria.EqualsNumber, 3,
			[]int{30, 20, 10, 5, -5, -15},
			-15, 60, false,
		},

		// Pick at least 3 of [30, 20, 10, 5, -5, -15]: the low end must still reach for that third child, while the
		// high end has 4 advantages to take and so is not touched by the count at all.
		{
			"a minimum count can bind one end of a long list and not the other",
			criteria.AtLeastNumber, 3,
			[]int{30, 20, 10, 5, -5, -15},
			-15, 65, false,
		},

		// Pick at most 3 of [30, 20, 10, 5, -5, -15]: the low end is under the cap with 2 disadvantages, so only the
		// high end is held back by it.
		{
			"a maximum count can bind one end of a long list and not the other",
			criteria.AtMostNumber, 3,
			[]int{30, 20, 10, 5, -5, -15},
			-20, 60, false,
		},

		// A count of 0 is the edge of each comparison rather than a count at all: two of them forbid picking anything,
		// while the third forbids nothing.

		// Pick exactly 0 of [10, 20, -5, -30]: the only selection that makes the count is the empty one, so the
		// children cannot be reached at either end.
		{
			"an exact count of none costs nothing whatever is on offer",
			criteria.EqualsNumber, 0,
			[]int{10, 20, -5, -30},
			0, 0, true,
		},

		// Pick at most 0 of [10, 20, -5, -30]: likewise, nothing may be taken.
		{
			"a maximum count of none costs nothing whatever is on offer",
			criteria.AtMostNumber, 0,
			[]int{10, 20, -5, -30},
			0, 0, true,
		},

		// Pick at least 0 of [10, 20, -5, -30]: every selection satisfies it, including the empty one, so it is the
		// unconstrained picker -- every disadvantage at the low end, every advantage at the high.
		{
			"a minimum count of none constrains nothing",
			criteria.AtLeastNumber, 0,
			[]int{10, 20, -5, -30},
			-35, 30, false,
		},
	})
}

// TestPointsRangeForPointsPicker verifies the range of a container that asks for an amount of points to be picked. A
// points picker measures the very quantity it constrains, so the qualifier bounds the container's total directly:
// the qualifier binds the end it names, taking nothing bounds the other, and the end the picker names no limit for
// stays open however much the children could have accounted for.
//
// Which side of nothing the children sit on is what decides whether the qualifier binds at all, so the groups below
// are the three answers to that -- advantages, disadvantages, and both at once. Nothing here assumes the amount a
// picker asks for is positive: a disadvantage package asking for "at most -20 points" is as legitimate as one asking
// for "at least 20".
func TestPointsRangeForPointsPicker(t *testing.T) {
	c := check.New(t)

	// Advantages: every choice costs points.

	checkPickerCases(c, picker.Points, []pickerCase{
		// Pick exactly 20 points worth: whatever is chosen, 20 points is what it costs.
		{
			"an exact points picker is worth what it asks for",
			criteria.EqualsNumber, 20,
			[]int{5, 10, 20, 40},
			20, 20, true,
		},

		// Pick at most 20 points from [10, 20, 40]: the cap is the most that may be spent even though there is more
		// on offer, and taking nothing is always allowed.
		{
			"a maximum points picker is capped by what it allows",
			criteria.AtMostNumber, 20,
			[]int{10, 20, 40},
			0, 20, false,
		},

		// Pick at most 100 points from children that only add up to 30: the cap is still what bounds it. What the
		// children happen to add up to does not pull the end in, since the points of a skill or spell picked from
		// them may be assigned while picking, and a leveled trait's cost can be raised there too.
		{
			"a maximum points picker allowing more than there is is bounded by what it allows anyway",
			criteria.AtMostNumber, 100,
			[]int{10, 20},
			0, 100, false,
		},

		// Pick at most 20 points from children that carry none: children that can only cost nothing take no side,
		// and a picker with no side to take is worth nothing rather than what its qualifier names. Nothing on offer
		// can be spent, so the cap bounds an end nothing reaches toward. Every comparison reaches the same place over
		// such children -- the "at least" case below included.
		{
			"a maximum points picker over children carrying no points costs nothing",
			criteria.AtMostNumber, 20,
			[]int{0, 0, 0},
			0, 0, true,
		},

		// A qualifier of 0 is the edge of each comparison: two of them forbid spending anything, while the third
		// forbids nothing.
		{
			"an exact points picker asking for nothing costs nothing",
			criteria.EqualsNumber, 0,
			[]int{10, 20},
			0, 0, true,
		},
		{
			"a maximum points picker allowing nothing costs nothing",
			criteria.AtMostNumber, 0,
			[]int{10, 20},
			0, 0, true,
		},
	})

	// The comparisons that name no ceiling leave none, since nothing an advantage costs brings the high end back
	// within a limit.
	checkOpenRangeCases(c, picker.Points, []openRangeCase{
		// Pick at least 20 points from [10, 20, 40]: 20 is the least that satisfies it, and nothing caps the rest.
		{
			"a minimum points picker spans what it asks for upward without limit",
			criteria.AtLeastNumber, 20,
			[]int{10, 20, 40},
			"20+",
		},

		// Any amount from [10, 20, 40]: the qualifier names neither end, so only taking nothing bounds it.
		{
			"an unconstrained points picker over advantages spans nothing upward without limit",
			criteria.AnyNumber, 0,
			[]int{10, 20, 40},
			"0+",
		},

		// A minimum of 0 is the comparison of the three that forbids nothing, and reaches the same place.
		{
			"a minimum points picker asking for nothing constrains nothing",
			criteria.AtLeastNumber, 0,
			[]int{10, 20},
			"0+",
		},
	})

	// Pick at least 100 points from children whose own costs only add up to 50: those costs set no ceiling on what the
	// container is worth, since the points of a skill or spell picked from it may be assigned while picking. This is the
	// shape of 13 of the 23 "at least" pickers in the master library, whose children carry no points at all.
	noCeiling := newPickerRange(
		picker.Points, criteria.AtLeastNumber,
		100,
		[]int{10, 40},
	)
	c.NotNil(noCeiling.Min, "a minimum points picker its children cannot bound has a lower limit")
	c.Equal(fxp.FromInteger(100), *noCeiling.Min, "it costs at least what it asks for")
	c.Nil(noCeiling.Max, "a minimum points picker its children cannot bound has no upper limit")
	c.False(noCeiling.IsSettled(), "a cost with no upper limit is never settled")
	c.Equal("100+", noCeiling.String(), "it renders as its minimum and a plus")

	// A container of 0-point skills is the other common form of that picker, and it parts ways with the case above:
	// children that can only cost nothing leave the picker no side to take, so the qualifier binds neither end and
	// the container is worth nothing until those children carry points of their own.
	skills := NewSkill(nil, nil, true)
	skills.TemplatePicker.Type = picker.Points
	skills.TemplatePicker.Qualifier.Compare = criteria.AtLeastNumber
	skills.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(16)
	for range 3 {
		child := NewSkill(nil, skills, false)
		child.Points = 0
		skills.Children = append(skills.Children, child)
	}
	c.Equal("0", skills.PointsRange(nil).String(),
		"a points picker over skills carrying no points costs nothing")
	c.Equal(fxp.Int(0), skills.AdjustedPoints(nil), "and its total agrees with the settled cost")

	// Disadvantages: every choice returns points instead, which is the shape of the disadvantage packages templates
	// use. Every end is the mirror of the advantage group: what is cheapest is taking the most.

	checkPickerCases(c, picker.Points, []pickerCase{
		// "Pick -50 points worth" is worth what it asks for, just as "pick 20 points worth" is.
		{
			"an exact points picker asking for disadvantages is worth what it asks for",
			criteria.EqualsNumber, -50,
			[]int{-10, -20, -40},
			-50, -50, true,
		},

		// Pick at least -20 points from [-10, -20, -40]: taking nothing satisfies it, and so does the -10 or the
		// -20, but nothing further down does.
		{
			"a minimum points picker asking for disadvantages is floored by what it asks for",
			criteria.AtLeastNumber, -20,
			[]int{-10, -20, -40},
			-20, 0, false,
		},

		// Pick at least -100 points from them: the qualifier is the floor here too, even though taking every
		// disadvantage on offer only returns 70. It is the mirror of the 100-point cap in the advantage group.
		{
			"a minimum below what the children can reach is the floor anyway",
			criteria.AtLeastNumber, -100,
			[]int{-10, -20, -40},
			-100, 0, false,
		},

		// An exact 0 is met only by picking nothing, whatever is on offer.
		{
			"an exact points picker asking for nothing costs nothing even over disadvantages",
			criteria.EqualsNumber, 0,
			[]int{-10, -20, -40},
			0, 0, true,
		},
	})

	// The comparisons that name no floor leave none, the mirror of the ceilings above.
	checkOpenRangeCases(c, picker.Points, []openRangeCase{
		// Pick at most -20 points from [-10, -20, -40]: the -20 alone is the least that still satisfies the cap, and
		// taking nothing would not, so the cap is what bounds the high end, not 0.
		{
			"a maximum points picker asking for disadvantages is bounded by its cap, not by zero",
			criteria.AtMostNumber, -20,
			[]int{-10, -20, -40},
			"≤-20",
		},

		// Any amount from [-10, -20, -40]: only taking nothing bounds it, which over disadvantages is the high end.
		{
			"an unconstrained points picker over disadvantages spans nothing downward without limit",
			criteria.AnyNumber, 0,
			[]int{-10, -20, -40},
			"≤0",
		},

		// A cap of 0 is met by every disadvantage there is, which is where it parts ways with the exact 0 above:
		// "spending nothing" and "taking no disadvantages" are not the same thing.
		{
			"a maximum of nothing still allows every disadvantage to be taken",
			criteria.AtMostNumber, 0,
			[]int{-10, -20, -40},
			"≤0",
		},
	})

	// The would-be mirror of the "no ceiling" case, and where it stops being one. A picker has a side to take only
	// when its children do, and children that can only cost nothing take neither side. A cap below nothing over them
	// therefore binds no end at all, and the picker settles at nothing -- exactly where the same children under
	// "at least 16" above settle, since which end the qualifier names stops mattering once no end is bound.
	noFloor := newPickerRange(
		picker.Points, criteria.AtMostNumber,
		-50,
		[]int{0, 0, 0},
	)
	c.Equal(fxp.Int(0), *noFloor.Min, "a cap below nothing over children carrying none costs nothing")
	c.Equal(fxp.Int(0), *noFloor.Max, "at both ends, since nothing on offer says the cap applies")
	c.True(noFloor.IsSettled(), "a cost with nothing left to decide is settled")
	c.Equal("0", noFloor.String(), "it renders as a single cost")

	// Both at once, which for a points picker is rarer still -- 2 of the 1,126 pickers in the master library, both of
	// them disadvantage packages offering something to spend the returned points on. Children reaching away from 0 in
	// both directions leave the picker no side to take, so an exact qualifier is the only one that still says what
	// the container is worth; every other comparison names an end the children do not agree on.

	checkPickerCases(c, picker.Points, []pickerCase{
		// Pick exactly 10 points worth: an exact qualifier is what it costs, whichever side the children come from.
		{
			"an exact points picker over a mixed list is still worth what it asks for",
			criteria.EqualsNumber, 10,
			[]int{20, 10, -5, -15},
			10, 10, true,
		},

		// Pick exactly 0: met by picking nothing, whatever is on offer at either side.
		{
			"an exact points picker asking for nothing costs nothing over a mixed list too",
			criteria.EqualsNumber, 0,
			[]int{20, 10, -5, -15},
			0, 0, true,
		},
	})

	// Every other comparison over [20, 10, -5, -15], whichever end it names and whichever side its qualifier falls
	// on, together with the one list that reaches both ways by a single child.
	checkOpenRangeCases(c, picker.Points, []openRangeCase{
		{
			"an unconstrained points picker over a mixed list is open at both ends",
			criteria.AnyNumber, 0,
			[]int{20, 10, -5, -15},
			"—",
		},
		{
			"a minimum over a mixed list is open at both ends",
			criteria.AtLeastNumber, 10,
			[]int{20, 10, -5, -15},
			"—",
		},
		{
			"a maximum over a mixed list is open at both ends",
			criteria.AtMostNumber, 10,
			[]int{20, 10, -5, -15},
			"—",
		},
		{
			"a minimum below nothing over a mixed list is open at both ends",
			criteria.AtLeastNumber, -30,
			[]int{20, 10, -5, -15},
			"—",
		},
		{
			"a maximum below nothing over a mixed list is open at both ends",
			criteria.AtMostNumber, -30,
			[]int{20, 10, -5, -15},
			"—",
		},
		{
			"a single disadvantage on offer is enough to leave a maximum open at both ends",
			criteria.AtMostNumber, 40,
			[]int{10, 20, 40, -5},
			"—",
		},
	})

	// A qualifier on the far side of nothing from every child is an authoring error, not a constraint to work around,
	// even where every pick would happen to meet it. It is left open at both ends, so it stands out rather than passing
	// for the unconstrained picker it most resembles.
	checkOpenRangeCases(c, picker.Points, []openRangeCase{
		{
			"a minimum below nothing over advantages is open at both ends",
			criteria.AtLeastNumber, -20,
			[]int{10, 20, 40},
			"—",
		},
		{
			"a maximum below nothing over advantages is open at both ends",
			criteria.AtMostNumber, -20,
			[]int{10, 20, 40},
			"—",
		},
		{
			"a minimum above nothing over disadvantages is open at both ends",
			criteria.AtLeastNumber, 20,
			[]int{-10, -20, -40},
			"—",
		},
		{
			"a maximum above nothing over disadvantages is open at both ends",
			criteria.AtMostNumber, 20,
			[]int{-10, -20, -40},
			"—",
		},
	})
}

// TestPointsRangeForPickerWithOpenEnds verifies the ranges the tables above cannot express: those with no limit at
// one end, or at neither. A points picker asks for an amount without saying how much beyond it may be spent, so a
// minimum over advantages leaves its upper end open and a maximum over disadvantages leaves its lower end open --
// those are the halves. A count picker inherits whichever ends the children it offers leave open, so presenting both
// halves at once is the only way to reach a range with no limit at either end. When an item node gains a range of
// its own, it takes the place of these inner containers and nothing else here changes.
func TestPointsRangeForPickerWithOpenEnds(t *testing.T) {
	c := check.New(t)

	// "Pick at least 1 point" over advantages: no ceiling, since nothing the children can cost brings the high end
	// back within a limit. The children have to cost something for the qualifier to bind at all -- children that can
	// only cost nothing leave the picker no side to take and settle it at nothing instead (see
	// TestPointsRangeForPointsPicker).
	noCeiling := func() *Trait {
		return newPickerContainer(
			picker.Points, criteria.AtLeastNumber,
			1,
			[]int{5, 10},
		)
	}

	// "Pick at most -1 point" over disadvantages: no floor, since nothing the children can return brings the low end
	// back within a limit. The children have to sit on the side the picker asks for here as well.
	noFloor := func() *Trait {
		return newPickerContainer(
			picker.Points, criteria.AtMostNumber,
			-1,
			[]int{-10, -20, -30},
		)
	}

	// A settled inner picker to pair each of them with: "pick 20 points worth" is worth 20 whichever way it is met.
	settledAt20 := func() *Trait {
		return newPickerContainer(
			picker.Points, criteria.EqualsNumber,
			20,
			[]int{5, 20},
		)
	}

	// And a settled disadvantage package to sit opposite an open end: "pick -20 points worth" is worth -20 whichever
	// way it is met.
	settledAtMinus20 := func() *Trait {
		return newPickerContainer(
			picker.Points, criteria.EqualsNumber,
			-20,
			[]int{0, 0},
		)
	}

	// Two more open children, further from 0 than the first pair, to offer alongside them.
	noCeilingAt30 := func() *Trait {
		return newPickerContainer(
			picker.Points, criteria.AtLeastNumber,
			30,
			[]int{5, 10},
		)
	}
	noFloorAt30 := func() *Trait {
		return newPickerContainer(
			picker.Points, criteria.AtMostNumber,
			-30,
			[]int{-10, -20, -30},
		)
	}

	checkOpenPickerCases(c, []openPickerCase{
		// An open end survives being picked from: the outer picker can only pass on what the child leaves open.
		{
			"a picker offering only a child with no upper limit has none either",
			criteria.EqualsNumber, 1,
			[]*Trait{noCeiling()},
			"1+", false,
		},
		{
			"a picker offering only a child with no lower limit has none either",
			criteria.EqualsNumber, 1,
			[]*Trait{noFloor()},
			"≤-1", false,
		},

		// Paired with a settled child, the open end stays open and the other end is whatever the two reach. Picking
		// 1 of [16+, 20] costs at least 16, since that is the cheaper of the two to start from.
		{
			"a child with no upper limit leaves the picker that offers it without one",
			criteria.EqualsNumber, 1,
			[]*Trait{noCeiling(), settledAt20()},
			"1+", false,
		},
		{
			"taking both adds the settled cost to the lower limit and leaves the upper open",
			criteria.EqualsNumber, 2,
			[]*Trait{noCeiling(), settledAt20()},
			"21+", false,
		},
		{
			"a child with no lower limit leaves the picker that offers it without one",
			criteria.EqualsNumber, 1,
			[]*Trait{noFloor(), settledAt20()},
			"≤20", false,
		},
		{
			"taking both adds the settled cost to the upper limit and leaves the lower open",
			criteria.EqualsNumber, 2,
			[]*Trait{noFloor(), settledAt20()},
			"≤19", false,
		},
		{
			"a minimum count passes on the open end just as an exact one does",
			criteria.AtLeastNumber, 1,
			[]*Trait{noCeiling(), settledAt20()},
			"1+", false,
		},
		{
			"a maximum count passes it on too",
			criteria.AtMostNumber, 1,
			[]*Trait{noFloor(), settledAt20()},
			"≤20", false,
		},

		// An open end says nothing about where the other one sits, so each open end is checked against a limit
		// below nothing, at nothing, and above it. Taking nothing is always an option under a maximum or
		// unconstrained count, which is what puts a limit at exactly 0: no picker reaches that on its own, since its
		// qualifier only opens an end it passes, and 0 is passed by every end.

		// No upper limit, with the lower one below nothing, at nothing, and above it.
		{
			"a disadvantage that may be taken alongside an open child puts the lower limit below nothing",
			criteria.AnyNumber, 0,
			[]*Trait{noCeiling(), settledAtMinus20()},
			"-20+", false,
		},
		{
			"a picker that need take nothing of what it offers puts the lower limit at nothing",
			criteria.AnyNumber, 0,
			[]*Trait{noCeiling()},
			"0+", false,
		},
		{
			"a settled child that must be taken alongside the open one puts the lower limit above nothing",
			criteria.EqualsNumber, 2,
			[]*Trait{noCeiling(), settledAt20()},
			"21+", false,
		},

		// No lower limit, with the upper one below nothing, at nothing, and above it.
		{
			"a disadvantage that must be taken alongside an open child puts the upper limit below nothing",
			criteria.EqualsNumber, 2,
			[]*Trait{noFloor(), settledAtMinus20()},
			"≤-21", false,
		},
		{
			"a picker that need take nothing of what it offers puts the upper limit at nothing",
			criteria.AnyNumber, 0,
			[]*Trait{noFloor()},
			"≤0", false,
		},
		{
			"an advantage that may be taken alongside the open child puts the upper limit above nothing",
			criteria.AnyNumber, 0,
			[]*Trait{noFloor(), settledAt20()},
			"≤20", false,
		},

		// A maximum count reaches the same two ends as an unconstrained one, since neither has to take anything.
		{
			"a maximum count over a child with no upper limit spans nothing to no limit at all",
			criteria.AtMostNumber, 1,
			[]*Trait{noCeiling()},
			"0+", false,
		},
		{
			"a maximum count over a child with no lower limit spans no limit at all to nothing",
			criteria.AtMostNumber, 1,
			[]*Trait{noFloor()},
			"≤0", false,
		},
		{
			"a settled child alongside the open one is left out of the end that takes nothing",
			criteria.AtMostNumber, 1,
			[]*Trait{noCeiling(), settledAt20()},
			"0+", false,
		},

		// Two children open on the same side leave the picker with nothing to order them by at that end, so the
		// other end is what decides between them. The lower limits of [1+, 30+] still order, and the upper limits of
		// [≤-1, ≤-30] do too.
		{
			"one of two children with no upper limit costs at least the cheaper of them",
			criteria.EqualsNumber, 1,
			[]*Trait{noCeiling(), noCeilingAt30()},
			"1+", false,
		},
		{
			"taking both adds their lower limits together",
			criteria.EqualsNumber, 2,
			[]*Trait{noCeiling(), noCeilingAt30()},
			"31+", false,
		},
		{
			"neither need be taken when the count does not require it",
			criteria.AnyNumber, 0,
			[]*Trait{noCeiling(), noCeilingAt30()},
			"0+", false,
		},
		{
			"one of two children with no lower limit costs at most the costlier of them",
			criteria.EqualsNumber, 1,
			[]*Trait{noFloor(), noFloorAt30()},
			"≤-1", false,
		},
		{
			"taking both adds their upper limits together",
			criteria.EqualsNumber, 2,
			[]*Trait{noFloor(), noFloorAt30()},
			"≤-31", false,
		},
		{
			"neither of those need be taken either",
			criteria.AnyNumber, 0,
			[]*Trait{noFloor(), noFloorAt30()},
			"≤0", false,
		},

		// The open child listed after a bounded one, rather than before it, which is the other way the two can be
		// handed to the ordering.
		{
			"an open child offered after a settled one is still the costliest thing there",
			criteria.EqualsNumber, 1,
			[]*Trait{settledAt20(), noCeiling()},
			"1+", false,
		},
		{
			"an open child offered after a settled one is still the cheapest thing there",
			criteria.EqualsNumber, 1,
			[]*Trait{settledAt20(), noFloor()},
			"≤20", false,
		},

		// Offering both halves at once is the only way to reach a range with no limit at either end, which renders as
		// a dash. However the count is constrained, neither end can be brought back within a limit.
		{
			"a picker offering a child open at each end has no limit at either",
			criteria.EqualsNumber, 1,
			[]*Trait{noCeiling(), noFloor()},
			"—", false,
		},
		{
			"taking both leaves it open at both ends as well",
			criteria.EqualsNumber, 2,
			[]*Trait{noCeiling(), noFloor()},
			"—", false,
		},
		{
			"an unconstrained count over them is open at both ends",
			criteria.AnyNumber, 0,
			[]*Trait{noCeiling(), noFloor()},
			"—", false,
		},
		{
			"a minimum count over them is open at both ends",
			criteria.AtLeastNumber, 1,
			[]*Trait{noCeiling(), noFloor()},
			"—", false,
		},
		{
			"a maximum count over them is open at both ends",
			criteria.AtMostNumber, 1,
			[]*Trait{noCeiling(), noFloor()},
			"—", false,
		},

		// Forbidding any selection is the one thing that closes an open end, since nothing that has one can be taken.
		{
			"a maximum count of none closes both ends, whatever the children left open",
			criteria.AtMostNumber, 0,
			[]*Trait{noCeiling(), noFloor()},
			"0", true,
		},
		{
			"an exact count of none does the same",
			criteria.EqualsNumber, 0,
			[]*Trait{noCeiling(), noFloor()},
			"0", true,
		},
	})
}

// TestPointsRangeForExactPointsPickerDirectly verifies the one branch of pointsRangeForPickerByPoints that no
// container reaches: an exact points picker is settled at its qualifier by settledPickerCost, which every PointsRange
// consults before walking any children, so nothing that arrives here carries that comparison. The branch has to
// answer anyway, and has to answer the same thing the short circuit does, since a change to either could make it the
// one that runs.
func TestPointsRangeForExactPointsPickerDirectly(t *testing.T) {
	c := check.New(t)

	exact := func(qualifier int) criteria.Number {
		var cq criteria.Number
		cq.Compare = criteria.EqualsNumber
		cq.Qualifier = fxp.FromInteger(qualifier)
		return cq
	}
	// Children that reach nowhere near the qualifier, one of them without an upper limit at all: none of it matters,
	// since the qualifier is the cost.
	children := []PointsRange{
		PointsRangeOf(fxp.Five),
		PointsRangeOf(fxp.FromInteger(-40)),
		pointsRangeAtLeast(fxp.Ten),
	}

	r := pointsRangeForPickerByPoints(exact(20), children)
	checkRange(c, 20, 20, r, "an exact points picker is worth its qualifier whatever its children can reach")
	c.True(r.IsSettled(), "an exact points picker is settled")

	r = pointsRangeForPickerByPoints(exact(-50), children)
	checkRange(c, -50, -50, r, "an exact points picker asking for disadvantages is worth its qualifier too")

	// The short circuit that normally answers first has to agree with it.
	var tp TemplatePicker
	tp.Type = picker.Points
	tp.Qualifier = exact(20)
	value, settled := settledPickerCost(tp)
	c.True(settled, "the short circuit settles an exact points picker without looking at its children")
	c.Equal(fxp.Twenty, value, "and settles it at the same cost the branch returns")
}

// TestPointsRangeForPointsPickerWithNoChildren verifies what a points picker authored with nothing to pick from
// reports. Nothing on offer is nothing to take a side, so the qualifier binds neither end and the container can only
// cost nothing -- the same answer a count picker gives when there is nothing to count (see
// TestPointsRangeForCountPicker), and the same one the picker keeps while every choice on offer carries no points of
// its own. An exact qualifier is the exception: it says what the container is worth without consulting its children
// at all.
func TestPointsRangeForPointsPickerWithNoChildren(t *testing.T) {
	c := check.New(t)

	r := newPickerRange(picker.Points, criteria.EqualsNumber, 20, nil)
	checkRange(c, 20, 20, r, "an exact points picker with nothing to pick from is worth what it asks for")
	c.True(r.IsSettled(), "and it is settled, since every way of meeting it costs the same")

	// The short circuit answers this one without looking at the children, so it cannot help but agree -- which is
	// the point: the branch that does look at them has to come back with the same answer.
	var tp TemplatePicker
	tp.Type = picker.Points
	tp.Qualifier.Compare = criteria.EqualsNumber
	tp.Qualifier.Qualifier = fxp.Twenty
	value, settled := settledPickerCost(tp)
	c.True(settled, "the short circuit settles an exact points picker with no children too")
	c.Equal(fxp.Twenty, value, "at the same cost the branch returns")

	r = newPickerRange(picker.Points, criteria.AtMostNumber, 20, nil)
	checkRange(c, 0, 0, r, "a maximum points picker with nothing to pick from costs nothing")
	c.True(r.IsSettled(), "and is settled, since there is nothing to spend the cap on")

	r = newPickerRange(picker.Points, criteria.AtLeastNumber, 20, nil)
	c.Equal("0", r.String(), "a minimum points picker with nothing to pick from costs nothing as well")
	c.True(r.IsSettled(), "and is settled too")

	// The same container once a single choice carrying no points is in place: the answers do not move, since a
	// choice that can only cost nothing leaves the picker as sideless as no choice at all.
	r = newPickerRange(picker.Points, criteria.AtMostNumber, 20, []int{0})
	checkRange(c, 0, 0, r, "a maximum points picker does not change once a choice carrying no points is added")
	r = newPickerRange(picker.Points, criteria.AtLeastNumber, 20, []int{0})
	c.Equal("0", r.String(), "nor does a minimum points picker")
}

// TestPointsRangeForInvalidPicker verifies what a container carrying a picker type the app does not know reports.
// Only a file written by something else can hold one, since the editor offers the valid types and nothing else, and a
// choice that cannot be presented is no choice at all: the container is worth the total of its children, exactly as
// one carrying no picker is.
func TestPointsRangeForInvalidPicker(t *testing.T) {
	c := check.New(t)

	parent := newPickerContainer(
		picker.Count, criteria.EqualsNumber,
		1,
		[]int{10, 20, -5},
	)
	parent.TemplatePicker.Type = picker.LastType + 1
	r := parent.PointsRange(nil)
	checkRange(c, 25, 25, r, "a container whose picker type is not a known one is worth the total of its children")
	c.True(r.IsSettled(), "and it is settled, since there is no choice left to make")
}

// TestPointsRangeForPickerIsNeverInverted sweeps every picker type and comparison over a spread of qualifiers and
// children, verifying the one invariant every branch has to hold: a range's minimum is never above its maximum. An
// inverted range renders as nonsense like "100~50" rather than failing outright, so nothing else catches it.
func TestPointsRangeForPickerIsNeverInverted(t *testing.T) {
	c := check.New(t)

	childSets := [][]int{
		{},
		{0, 0, 0},
		{10, 20, 40},
		{-10, -20, -40},
		{10, 20, 40, -5},
		{5, -5},
	}
	comparisons := []criteria.NumericComparison{
		criteria.AnyNumber,
		criteria.EqualsNumber,
		criteria.NotEqualsNumber,
		criteria.AtLeastNumber,
		criteria.AtMostNumber,
	}
	qualifiers := []int{-100, -20, -1, 0, 1, 2, 20, 100}
	for _, pt := range []picker.Type{picker.NotApplicable, picker.Count, picker.Points} {
		for _, compare := range comparisons {
			for _, qualifier := range qualifiers {
				for _, children := range childSets {
					r := newPickerRange(
						pt, compare,
						qualifier,
						children,
					)
					if r.Min == nil || r.Max == nil {
						continue // An open end cannot be on the wrong side of anything.
					}
					c.True(*r.Min <= *r.Max, fmt.Sprintf(
						"%v picker, %v %d over %v must not invert, got %s", pt, compare, qualifier, children, r,
					))
				}
			}
		}
	}
}

// TestPointsRangeNestedPickers verifies that a picker inside a picker composes, which 93 of the pickers in the master
// library do.
func TestPointsRangeNestedPickers(t *testing.T) {
	c := check.New(t)

	// An outer "pick 1" over two inner containers: one settled at 20 points, one spanning 5 to 40.
	checkRange(
		c, 5, 40,
		newNestedPickerRange(
			criteria.EqualsNumber, 1,
			newPickerContainer(
				picker.Points, criteria.EqualsNumber,
				20,
				[]int{5, 10, 20},
			),
			newPickerContainer(
				picker.Count, criteria.EqualsNumber,
				1,
				[]int{5, 40},
			),
		),
		"an outer picker spans the cheapest and costliest outcomes of the pickers inside it",
	)
}

// TestPointsRangeForPlainContainers verifies that a container presenting no choice reports the total of its children,
// settled, which is what it has always cost.
func TestPointsRangeForPlainContainers(t *testing.T) {
	c := check.New(t)

	parent := NewTrait(nil, nil, true)
	for _, points := range []int{10, 20, -5} {
		child := NewTrait(nil, parent, false)
		child.BasePoints = fxp.FromInteger(points)
		parent.Children = append(parent.Children, child)
	}
	r := parent.PointsRange(nil)
	checkRange(c, 25, 25, r, "a container presenting no choice is worth the total of its children")
	c.True(r.IsSettled(), "a container presenting no choice is settled")
	c.Equal(parent.AdjustedPoints(nil), *r.Min, "the range agrees with the point cost")

	parent.ContainerType = container.AlternativeAbilities
	r = parent.PointsRange(nil)
	// 20 at full cost, 10 and -5 at a fifth: 20 + 2 + -1.
	checkRange(c, 21, 21, r, "alternative abilities are worth what they cost")
	c.Equal(parent.AdjustedPoints(nil), *r.Min, "the range agrees with the point cost for alternative abilities")

	disabled := NewTrait(nil, nil, false)
	disabled.BasePoints = fxp.FromInteger(15)
	disabled.Disabled = true
	checkRange(c, 0, 0, disabled.PointsRange(nil), "a disabled trait costs nothing")
}

// TestAdjustedPointsCollapsesSettledPickers verifies that a container whose choices all cost the same reports that
// cost, rather than the total of children only some of which will ever be taken, and that one whose choices differ
// keeps reporting what it always has.
func TestAdjustedPointsCollapsesSettledPickers(t *testing.T) {
	c := check.New(t)

	// "Pick 20 points worth" from 75 points of children is worth 20, not 75.
	exact := newPickerContainer(
		picker.Points, criteria.EqualsNumber,
		20,
		[]int{5, 10, 20, 40},
	)
	c.Equal(fxp.FromInteger(20), exact.AdjustedPoints(nil),
		"a container asking for an exact number of points is worth that many")

	// "Pick 1" of three 10-point children is worth 10, not 30.
	identical := newPickerContainer(
		picker.Count, criteria.EqualsNumber,
		1,
		[]int{10, 10, 10},
	)
	c.Equal(fxp.FromInteger(10), identical.AdjustedPoints(nil),
		"a container asking for one of several identically priced children is worth one of them")

	// "Pick 1" of children that differ has no single answer, so the long-standing total is left alone.
	varied := newPickerContainer(
		picker.Count, criteria.EqualsNumber,
		1,
		[]int{10, 20, 40},
	)
	c.Equal(fxp.FromInteger(70), varied.AdjustedPoints(nil),
		"a container whose choices differ still reports the total of its children")

	// Stripping the choices -- what copying such a container anywhere but a template does -- returns it to the total.
	exact.TemplatePicker = TemplatePicker{}
	c.Equal(fxp.FromInteger(75), exact.AdjustedPoints(nil),
		"a container that no longer presents a choice is worth the total of its children again")
}

// TestPointsRangeFormatting verifies how a range renders, which is what every points column shows.
func TestPointsRangeFormatting(t *testing.T) {
	c := check.New(t)

	c.Equal("10", PointsRangeOf(fxp.Ten).String(), "a settled range renders as the bare cost")
	c.Equal("-1", PointsRangeOf(fxp.NegOne).String(), "a settled negative range renders as the bare cost")
	c.Equal("10~15", newPointsRange(fxp.Ten, fxp.Fifteen).String(), "a range renders both ends")
	c.Equal("-30~-20",
		newPointsRange(fxp.FromInteger(-30), fxp.FromInteger(-20)).String(),
		"a range of disadvantages stays readable, with no sign to confuse the separator for")
	c.Equal("10+", pointsRangeAtLeast(fxp.Ten).String(), "a range with no upper limit renders as a plus")
	c.Equal("≤15", pointsRangeAtMost(fxp.Fifteen).String(), "a range with no lower limit renders as a cap")
	c.Equal("0+", pointsRangeAtLeast(0).String(), "an end open above nothing still renders as a plus")
	c.Equal("≤0", pointsRangeAtMost(0).String(), "an end open below nothing still renders as a cap")
	c.Equal("—", PointsRange{}.String(), "a range with no limits at all renders as a dash")
	c.Equal("1,500~2,000",
		newPointsRange(fxp.FromInteger(1500), fxp.FromInteger(2000)).Comma(),
		"Comma renders both ends with separators")
}

// TestPointsLessFromString verifies that a points column sorts by the low end of whatever it shows, which a plain
// numeric comparison cannot do once a cell holds a range.
func TestPointsLessFromString(t *testing.T) {
	c := check.New(t)

	c.True(PointsLessFromString("9", "10"), "plain costs sort numerically, not as text")
	c.True(PointsLessFromString("10", "100"), "plain costs sort numerically past a single digit")
	c.True(PointsLessFromString("-5", "0"), "negative costs sort ahead of nothing")
	c.True(PointsLessFromString("1,500", "2,000"), "separators don't alter the order")

	c.True(PointsLessFromString("9~20", "10"), "a range sorts by its low end")
	c.False(PointsLessFromString("10", "9~20"), "a range sorts by its low end, from either side")
	c.True(PointsLessFromString("10", "10~15"), "a settled cost sorts ahead of a range starting at the same cost")
	c.True(PointsLessFromString("10~15", "20~25"), "ranges sort against each other by their low ends")
	c.True(PointsLessFromString("≤15", "-100"), "a range with no lower limit sorts ahead of everything")
	c.False(PointsLessFromString("-100", "≤15"), "a range with no lower limit sorts ahead of everything, from either side")
	c.True(PointsLessFromString("10", "10+"), "an unbounded upper end sorts after the same cost on its own")
	c.True(PointsLessFromString("0", "0+"), "the same holds where that cost is nothing")
	c.True(PointsLessFromString("≤0", "-100"), "a range open below nothing still sorts ahead of everything")

	c.True(PointsLessFromString("≤-100", "≤-5"), "ranges with no lower limit sort by their upper ends")
	c.False(PointsLessFromString("≤-5", "≤-100"), "ranges with no lower limit sort by their upper ends, from either side")
	c.True(PointsLessFromString("≤-5", "≤0"), "an upper end below nothing sorts ahead of one at nothing")
	c.True(PointsLessFromString("≤15", "—"), "a range with no limits at all sorts after one with an upper limit")
	c.True(PointsLessFromString("—", "-100"), "a range with no limits at all still sorts ahead of every lower limit")
	c.True(PointsLessFromString("10~15", "10+"), "a range with no upper limit sorts after one starting at the same cost")
	c.True(PointsLessFromString("10~15", "10~20"), "ranges starting at the same cost sort by their upper ends")
	c.True(PointsLessFromString("1,000~1,500", "1,000~2,000"), "separators don't alter the order of upper ends")
}

// TestRawPointsRangeLeavesOutBonuses verifies that the range a template picker counts a skill container by is built
// from the raw points of the skills inside it, just as a skill checked on its own is counted, rather than from the
// points the sheet's bonuses make of them.
func TestRawPointsRangeLeavesOutBonuses(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	bonus := NewSkillPointBonus()
	bonus.NameCriteria.Qualifier = "Guns"
	bonus.Amount = fxp.Two
	addTraitWithFeatures(e, "Gun Talent", bonus)

	guns := NewSkill(e, nil, true)
	guns.TemplatePicker.Type = picker.Count
	guns.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
	guns.TemplatePicker.Qualifier.Qualifier = fxp.One
	for _, points := range []int{1, 2} {
		child := NewSkill(e, guns, false)
		child.Name = "Guns"
		child.Points = fxp.FromInteger(points)
		guns.Children = append(guns.Children, child)
	}
	e.SetSkillList([]*Skill{guns})
	e.Recalculate()

	c.Equal("3~4", guns.PointsRange(nil).String(), "the range shown on a sheet includes its bonuses")
	c.Equal("1~2", guns.RawPointsRange().String(), "the range a picker counts leaves them out")
	c.Equal("2", guns.Children[1].RawPointsRange().String(), "a skill on its own is counted by its raw points")
}

// TestPointsCellForPickers verifies what the points column of a table shows for a container presenting choices: the
// range, explained in the tooltip, or the settled cost with nothing added to the tooltip.
func TestPointsCellForPickers(t *testing.T) {
	c := check.New(t)

	var data CellData
	newPickerContainer(
		picker.Count, criteria.EqualsNumber,
		1,
		[]int{10, 20, 40},
	).CellData(TraitPointsColumn, &data)
	c.Equal("10~40", data.Primary, "an unsettled container shows the range in the points column")
	c.Contains(data.Tooltip, "choices", "an unsettled cost explains itself in the tooltip")

	data = CellData{}
	newPickerContainer(
		picker.Points, criteria.EqualsNumber,
		20,
		[]int{5, 10, 40},
	).CellData(TraitPointsColumn, &data)
	c.Equal("20", data.Primary, "a settled container shows what it asks for")
	c.Equal("", data.Tooltip, "a settled cost has nothing to explain")

	// The same holds for skills and spells, which share the machinery.
	skill := NewSkill(nil, nil, true)
	skill.TemplatePicker.Type = picker.Count
	skill.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
	skill.TemplatePicker.Qualifier.Qualifier = fxp.One
	for _, points := range []int{1, 4} {
		child := NewSkill(nil, skill, false)
		child.Points = fxp.FromInteger(points)
		skill.Children = append(skill.Children, child)
	}
	data = CellData{}
	skill.CellData(SkillPointsColumn, &data)
	c.Equal("1~4", data.Primary, "a skill container presenting choices shows the range too")
}

// TestPointsRangeCanSatisfy verifies the question a picker dialog asks of a partly-made selection: could what is left
// to decide still satisfy the constraint?
func TestPointsRangeCanSatisfy(t *testing.T) {
	c := check.New(t)

	newCriteria := func(compare criteria.NumericComparison, qualifier int) criteria.Number {
		var n criteria.Number
		n.Compare = compare
		n.Qualifier = fxp.FromInteger(qualifier)
		return n
	}
	open := newPointsRange(fxp.Ten, fxp.FromInteger(40))

	c.True(open.CanSatisfy(newCriteria(criteria.EqualsNumber, 20)), "an exact cost within the range can still be met")
	c.False(open.CanSatisfy(newCriteria(criteria.EqualsNumber, 50)), "an exact cost beyond the range cannot be met")
	c.True(open.CanSatisfy(newCriteria(criteria.AtLeastNumber, 40)), "a minimum the range reaches can still be met")
	c.False(open.CanSatisfy(newCriteria(criteria.AtLeastNumber, 41)), "a minimum beyond the range cannot be met")
	c.True(open.CanSatisfy(newCriteria(criteria.AtMostNumber, 10)), "a maximum the range starts at can still be met")
	c.False(open.CanSatisfy(newCriteria(criteria.AtMostNumber, 9)), "a maximum below the range cannot be met")
	c.True(open.CanSatisfy(newCriteria(criteria.AnyNumber, 0)), "an unconstrained picker is always satisfied")

	// A settled range is the plain comparison, which is what every picker that holds no further choices does.
	settled := PointsRangeOf(fxp.Twenty)
	c.True(settled.CanSatisfy(newCriteria(criteria.EqualsNumber, 20)), "a settled cost meets an exact match")
	c.False(settled.CanSatisfy(newCriteria(criteria.EqualsNumber, 21)), "a settled cost fails an exact mismatch")
	nothing := PointsRangeOf(0)
	c.False(nothing.CanSatisfy(newCriteria(criteria.EqualsNumber, 1)), "picking nothing doesn't satisfy a picker")
	c.True(nothing.CanSatisfy(newCriteria(criteria.AtMostNumber, 5)), "picking nothing satisfies a maximum")

	// A range with no upper limit -- the everyday "pick at least N points" over children carrying no points -- can
	// still reach anything above its lower limit, but nothing below it.
	atLeast := pointsRangeAtLeast(fxp.Ten)
	c.True(atLeast.CanSatisfy(newCriteria(criteria.EqualsNumber, 1000)),
		"an exact cost above a range with no upper limit can still be met")
	c.False(atLeast.CanSatisfy(newCriteria(criteria.EqualsNumber, 5)),
		"an exact cost below the lower limit cannot be met")
	c.True(atLeast.CanSatisfy(newCriteria(criteria.AtLeastNumber, 1000)),
		"no minimum is out of reach of a range with no upper limit")
	c.True(atLeast.CanSatisfy(newCriteria(criteria.AtMostNumber, 10)),
		"a maximum the lower limit reaches can still be met")
	c.False(atLeast.CanSatisfy(newCriteria(criteria.AtMostNumber, 9)),
		"a maximum below the lower limit cannot be met")

	// The mirror of that: a range with no lower limit can reach anything below its upper limit.
	atMost := pointsRangeAtMost(fxp.Fifteen)
	c.True(atMost.CanSatisfy(newCriteria(criteria.EqualsNumber, -1000)),
		"an exact cost below a range with no lower limit can still be met")
	c.False(atMost.CanSatisfy(newCriteria(criteria.EqualsNumber, 20)),
		"an exact cost above the upper limit cannot be met")
	c.True(atMost.CanSatisfy(newCriteria(criteria.AtMostNumber, -1000)),
		"no maximum is out of reach of a range with no lower limit")
	c.True(atMost.CanSatisfy(newCriteria(criteria.AtLeastNumber, 15)),
		"a minimum the upper limit reaches can still be met")
	c.False(atMost.CanSatisfy(newCriteria(criteria.AtLeastNumber, 16)),
		"a minimum above the upper limit cannot be met")

	// A range with no limits at all cannot rule anything out.
	var unlimited PointsRange
	c.True(unlimited.CanSatisfy(newCriteria(criteria.EqualsNumber, 20)), "a range with no limits meets any exact cost")
	c.True(unlimited.CanSatisfy(newCriteria(criteria.AtLeastNumber, 20)), "a range with no limits meets any minimum")
	c.True(unlimited.CanSatisfy(newCriteria(criteria.AtMostNumber, 20)), "a range with no limits meets any maximum")

	// An unsettled range always holds something other than any single cost, however it is bounded.
	c.True(open.CanSatisfy(newCriteria(criteria.NotEqualsNumber, 20)),
		"a range spanning more than one cost can avoid any single one")
	c.True(atLeast.CanSatisfy(newCriteria(criteria.NotEqualsNumber, 10)),
		"a range with no upper limit can avoid any single cost")
	c.False(settled.CanSatisfy(newCriteria(criteria.NotEqualsNumber, 20)),
		"a settled cost cannot avoid itself")
}

// TestPointsRangeWithoutLimits verifies how a range with no limit at one end composes and compares, since nothing that
// is added to such an end brings it back within one.
func TestPointsRangeWithoutLimits(t *testing.T) {
	c := check.New(t)

	open := pointsRangeAtLeast(fxp.Ten)
	c.Nil(open.Max, "a range with no upper limit has none")
	c.False(open.IsSettled(), "a range with no upper limit is never settled")
	_, settled := open.Settled()
	c.False(settled, "a range with no upper limit has no single cost")

	sum := open.Add(PointsRangeOf(fxp.Five))
	c.NotNil(sum.Min, "adding a cost to a range with no upper limit still has a lower one")
	c.Equal(fxp.Fifteen, *sum.Min, "adding a cost raises the lower limit")
	c.Nil(sum.Max, "adding a cost to a range with no upper limit leaves it without one")

	// A child with no upper limit leaves the container that holds it without one either.
	parent := NewTrait(nil, nil, true)
	child := NewTrait(nil, parent, false)
	child.BasePoints = fxp.Ten
	parent.Children = append(parent.Children, child)
	c.True(parent.PointsRange(nil).IsSettled(), "a container of settled children is settled")
	c.Nil(sumPointsRanges([]PointsRange{PointsRangeOf(fxp.Ten), pointsRangeAtLeast(fxp.Five)}).Max,
		"a total that includes a cost with no upper limit has none")
}

// TestPointsRangeSign verifies which side of nothing a single range sits on. A points picker consults this to decide
// whether its qualifier constrains its children at all: a qualifier only binds children that reach toward it.
//
// A range is positive when nothing it can cost is less than nothing, and negative when nothing it can cost is more.
// An end with no limit is the side the range runs away toward, so one with no upper limit reaches above nothing and
// one with no lower limit reaches below it. A range that can only cost nothing is zero, which is no side of its own
// but agrees with either. A range holding costs on both sides is mixed, and a picker over such children has no side
// to take.
func TestPointsRangeSign(t *testing.T) {
	c := check.New(t)

	for _, tc := range []struct {
		msg  string
		r    PointsRange
		want PointsRangeSign
	}{
		// Positive: nothing the range can cost is less than nothing.
		{"a settled cost above nothing is positive", PointsRangeOf(fxp.FromInteger(20)), PointsRangePositive},
		{
			"a range between two costs above nothing is positive",
			newPointsRange(fxp.FromInteger(5), fxp.FromInteger(20)), PointsRangePositive,
		},
		{"a range reaching up from nothing is positive", newPointsRange(0, fxp.FromInteger(20)), PointsRangePositive},
		{"a range with no upper limit is positive", pointsRangeAtLeast(fxp.FromInteger(5)), PointsRangePositive},
		{"a range with no upper limit starting at nothing is positive", pointsRangeAtLeast(0), PointsRangePositive},

		// Zero: the range can only cost nothing, so it sits on neither side. It is what a child carrying no points
		// reports, and having no side of its own is what lets it keep the company of either -- see
		// TestSignForPointsRanges.
		{"a settled cost of nothing is zero", PointsRangeOf(0), PointsRangeZero},

		// Negative: nothing the range can cost is more than nothing.
		{"a settled cost below nothing is negative", PointsRangeOf(fxp.FromInteger(-20)), PointsRangeNegative},
		{
			"a range between two costs below nothing is negative",
			newPointsRange(fxp.FromInteger(-20), fxp.FromInteger(-5)), PointsRangeNegative,
		},
		{"a range with no lower limit is negative", pointsRangeAtMost(fxp.FromInteger(-5)), PointsRangeNegative},

		// A range capped at exactly nothing is the disadvantage package that need not be taken: every cost it can
		// reach returns points, and the one that does not returns none. That is the negative side, not both sides.
		{
			"a range reaching down from nothing is negative",
			newPointsRange(fxp.FromInteger(-20), 0), PointsRangeNegative,
		},
		{"a range with no lower limit capped at nothing is negative", pointsRangeAtMost(0), PointsRangeNegative},

		// Mixed: the range holds costs on both sides of nothing, so neither side describes it.
		{
			"a range straddling nothing is mixed",
			newPointsRange(fxp.FromInteger(-20), fxp.FromInteger(5)), PointsRangeMixed,
		},
		{
			"a range with no lower limit reaching above nothing is mixed",
			pointsRangeAtMost(fxp.FromInteger(5)), PointsRangeMixed,
		},
		{
			"a range with no upper limit reaching below nothing is mixed",
			pointsRangeAtLeast(fxp.FromInteger(-20)), PointsRangeMixed,
		},
		{"a range with no limit at either end is mixed", PointsRange{}, PointsRangeMixed},
	} {
		c.Equal(signName(tc.want), signName(tc.r.Sign()), tc.msg)
	}
}

// TestSignForPointsRanges verifies the side a whole set of children sits on, which is what a points picker actually
// asks. The set has a side only when every range in it agrees; one range on the other side, or one that reaches both
// ways at once, leaves the set with none. A range that can only cost nothing is the exception: it takes no side, so
// it contradicts neither, and a set of nothing but those is itself zero.
func TestSignForPointsRanges(t *testing.T) {
	c := check.New(t)

	advantage := PointsRangeOf(fxp.FromInteger(20))
	otherAdvantage := PointsRangeOf(fxp.FromInteger(5))
	disadvantage := PointsRangeOf(fxp.FromInteger(-20))
	otherDisadvantage := PointsRangeOf(fxp.FromInteger(-5))
	noPoints := PointsRangeOf(0)
	straddling := newPointsRange(fxp.FromInteger(-20), fxp.FromInteger(5))

	for _, tc := range []struct {
		msg    string
		ranges []PointsRange
		want   PointsRangeSign
	}{
		// A picker authored with no children at all has nothing to take a side, so the set is zero.
		{"no ranges at all are zero", nil, PointsRangeZero},

		{"a single advantage is positive", []PointsRange{advantage}, PointsRangePositive},
		{
			"advantages throughout are positive",
			[]PointsRange{advantage, otherAdvantage},
			PointsRangePositive,
		},
		{"a single disadvantage is negative", []PointsRange{disadvantage}, PointsRangeNegative},
		{
			"disadvantages throughout are negative",
			[]PointsRange{disadvantage, otherDisadvantage},
			PointsRangeNegative,
		},

		// One child on the other side is enough, whichever order the two arrive in.
		{
			"a disadvantage among advantages is mixed",
			[]PointsRange{advantage, otherAdvantage, disadvantage},
			PointsRangeMixed,
		},
		{
			"an advantage among disadvantages is mixed",
			[]PointsRange{disadvantage, otherDisadvantage, advantage},
			PointsRangeMixed,
		},

		// A child with no side of its own gives the set none either, since nothing it reaches can be ruled out.
		{"a single range straddling nothing is mixed", []PointsRange{straddling}, PointsRangeMixed},
		{
			"a range straddling nothing among advantages is mixed",
			[]PointsRange{advantage, straddling},
			PointsRangeMixed,
		},
		{
			"a range straddling nothing among disadvantages is mixed",
			[]PointsRange{disadvantage, straddling},
			PointsRangeMixed,
		},

		// Children carrying no points take no side, so they leave the side of the set to whatever else is in it. A
		// picker over skills or spells that carry none has no side at all, while one that offers them alongside a
		// disadvantage package is still a disadvantage package.
		{
			"children carrying no points are zero",
			[]PointsRange{noPoints, noPoints, noPoints},
			PointsRangeZero,
		},
		{
			"children carrying no points sit alongside advantages",
			[]PointsRange{advantage, noPoints},
			PointsRangePositive,
		},
		{
			"children carrying no points sit alongside disadvantages too",
			[]PointsRange{disadvantage, noPoints},
			PointsRangeNegative,
		},
	} {
		c.Equal(signName(tc.want), signName(SignForPointsRanges(tc.ranges...)), tc.msg)
	}
}

// signName names a sign, so that a failure above reads as the side it expected rather than as a number.
func signName(s PointsRangeSign) string {
	switch s {
	case PointsRangePositive:
		return "positive"
	case PointsRangeNegative:
		return "negative"
	case PointsRangeZero:
		return "zero"
	case PointsRangeMixed:
		return "mixed"
	default:
		return fmt.Sprintf("unknown(%d)", byte(s))
	}
}

// newPickerContainer creates a trait container carrying the given template choices, with one non-container child per
// supplied point cost.
func newPickerContainer(pt picker.Type, compare criteria.NumericComparison, qualifier int, childPoints []int) *Trait {
	parent := NewTrait(nil, nil, true)
	parent.TemplatePicker.Type = pt
	parent.TemplatePicker.Qualifier.Compare = compare
	parent.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(qualifier)
	for _, points := range childPoints {
		child := NewTrait(nil, parent, false)
		child.BasePoints = fxp.FromInteger(points)
		parent.Children = append(parent.Children, child)
	}
	return parent
}

// newPickerRange is newPickerContainer for the common case of only wanting the container's resulting range.
func newPickerRange(pt picker.Type, compare criteria.NumericComparison, qualifier int, childPoints []int) PointsRange {
	return newPickerContainer(pt, compare, qualifier, childPoints).PointsRange(nil)
}

// checkRange verifies both ends of a range against whole-number expectations, neither of which may be unlimited.
func checkRange(c check.Checker, minimum, maximum int, r PointsRange, msg string) {
	c.Helper()
	c.NotNil(r.Min, msg+" (has a lower limit)")
	c.NotNil(r.Max, msg+" (has an upper limit)")
	if r.Min == nil || r.Max == nil {
		return
	}
	c.Equal(fxp.FromInteger(minimum), *r.Min, msg+" (minimum)")
	c.Equal(fxp.FromInteger(maximum), *r.Max, msg+" (maximum)")
}

// pickerCase is one picker to check: the constraint it presents, the children it presents it over, the ends of the
// range that should result, and whether that leaves the container with a single cost. Both ends must be bounded; a
// picker whose range is open at one end is checked on its own, since there is more to say about it than a pair of
// numbers.
type pickerCase struct {
	msg         string
	compare     criteria.NumericComparison
	qualifier   int
	childPoints []int
	lower       int
	upper       int
	settled     bool
}

// checkPickerCases verifies the range of a picker of the given type built from each case, and what that range is
// worth when the case says there is nothing left to decide.
func checkPickerCases(c check.Checker, pt picker.Type, cases []pickerCase) {
	c.Helper()
	for _, tc := range cases {
		r := newPickerRange(pt, tc.compare, tc.qualifier, tc.childPoints)
		checkRange(c, tc.lower, tc.upper, r, tc.msg)
		value, settled := r.Settled()
		c.Equal(tc.settled, settled, tc.msg+" (settled)")
		if settled {
			c.Equal(fxp.FromInteger(tc.lower), value, tc.msg+" (settled cost)")
		}
	}
}

// openRangeCase is pickerCase for a picker whose range is open at an end, and so says what the range renders as --
// "20+", "≤-20", or "—" for one open at both ends -- in place of the pair of numbers that cannot describe it.
type openRangeCase struct {
	msg         string
	compare     criteria.NumericComparison
	qualifier   int
	childPoints []int
	want        string
}

// checkOpenRangeCases verifies the range of a picker of the given type built from each case. Nothing with an end left
// open is ever settled, which is checked alongside.
func checkOpenRangeCases(c check.Checker, pt picker.Type, cases []openRangeCase) {
	c.Helper()
	for _, tc := range cases {
		r := newPickerRange(pt, tc.compare, tc.qualifier, tc.childPoints)
		c.Equal(tc.want, r.String(), tc.msg)
		c.False(r.IsSettled(), tc.msg+" (settled)")
	}
}

// newNestedPickerRange returns the range of a count picker presenting the given containers as the choices it offers.
// A container is the only thing that produces a range of its own today -- no item node does yet -- so nesting one
// inside a picker is the only way to reach a range that is open at an end, or at both.
func newNestedPickerRange(compare criteria.NumericComparison, qualifier int, inner ...*Trait) PointsRange {
	outer := NewTrait(nil, nil, true)
	outer.TemplatePicker.Type = picker.Count
	outer.TemplatePicker.Qualifier.Compare = compare
	outer.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(qualifier)
	for _, one := range inner {
		one.SetParent(outer)
		outer.Children = append(outer.Children, one)
	}
	return outer.PointsRange(nil)
}

// openPickerCase is one nested picker whose range cannot be written as a pair of numbers, because it is open at one
// end or at both. Such a case says what the range renders as instead -- "16+", "≤-50", or "—" for one open at both
// ends -- and whether the container is left with a single cost, which nothing open ever is.
type openPickerCase struct {
	msg       string
	compare   criteria.NumericComparison
	qualifier int
	inner     []*Trait
	want      string
	settled   bool
}

// checkOpenPickerCases verifies the range of the nested picker each case describes.
func checkOpenPickerCases(c check.Checker, cases []openPickerCase) {
	c.Helper()
	for _, tc := range cases {
		r := newNestedPickerRange(tc.compare, tc.qualifier, tc.inner...)
		c.Equal(tc.want, r.String(), tc.msg)
		c.Equal(tc.settled, r.IsSettled(), tc.msg+" (settled)")
	}
}
