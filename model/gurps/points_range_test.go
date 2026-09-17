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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/check"
)

// newPickerContainer creates a trait container carrying the given template choices, with one non-container child per
// supplied point cost.
func newPickerContainer(pt picker.Type, compare criteria.NumericComparison, qualifier int, childPoints ...int) *Trait {
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

// TestPointsRangeForCountPicker verifies the range of a container that asks for a number of its children to be picked,
// for each way that number can be constrained.
func TestPointsRangeForCountPicker(t *testing.T) {
	c := check.New(t)

	// Pick exactly 2 of 5, 10, 20 and 40: the 2 cheapest are 5 + 10, the 2 costliest are 40 + 20.
	checkRange(c, 15, 60, newPickerContainer(picker.Count, criteria.EqualsNumber, 2, 5, 10, 20, 40).PointsRange(nil),
		"picking an exact count spans the cheapest and costliest ways to make it")

	// Pick exactly 1 of 10 and 10: every way of choosing costs the same, so there is nothing left to decide.
	settled := newPickerContainer(picker.Count, criteria.EqualsNumber, 1, 10, 10).PointsRange(nil)
	checkRange(c, 10, 10, settled, "a count picker whose children all cost the same is settled")
	value, isSettled := settled.Settled()
	c.True(isSettled, "a count picker whose children all cost the same is settled")
	c.Equal(fxp.Ten, value, "a count picker whose children all cost the same is worth one of them")

	// Pick any number of 5, 10 and -20: taking nothing but the disadvantage is cheapest, taking both advantages is
	// costliest.
	checkRange(c, -20, 15, newPickerContainer(picker.Count, criteria.AnyNumber, 0, 5, 10, -20).PointsRange(nil),
		"an unconstrained count takes every disadvantage at the low end and every advantage at the high end")

	// Pick at least 2 of 10, 20 and -5: the cheapest pair is -5 + 10, with nothing left worth taking; the costliest
	// pair is 20 + 10, again with nothing left worth taking.
	checkRange(c, 5, 30, newPickerContainer(picker.Count, criteria.AtLeastNumber, 2, 10, 20, -5).PointsRange(nil),
		"a minimum count has to be taken even when what is left would be better")

	// Pick at least 1 of 10 and -5: the -5 alone is cheapest, and the 10 plus the -5 that must also be taken... is not
	// required, since only one is, so the 10 alone is costliest.
	checkRange(c, -5, 10, newPickerContainer(picker.Count, criteria.AtLeastNumber, 1, 10, -5).PointsRange(nil),
		"beyond the required count, only what helps each end is taken")

	// Pick at most 2 of 10, 20, 40 and -5: the -5 alone is cheapest, and the 2 costliest are 40 + 20.
	checkRange(c, -5, 60, newPickerContainer(picker.Count, criteria.AtMostNumber, 2, 10, 20, 40, -5).PointsRange(nil),
		"a maximum count limits how much can be taken at either end")

	// A count larger than the number of children cannot ask for more than there is.
	checkRange(c, 30, 30, newPickerContainer(picker.Count, criteria.EqualsNumber, 9, 10, 20).PointsRange(nil),
		"a count exceeding the child count takes every child")

	// A container that presents a choice among nothing costs nothing.
	checkRange(c, 0, 0, newPickerContainer(picker.Count, criteria.EqualsNumber, 1).PointsRange(nil),
		"a picker with no children costs nothing")
}

// TestPointsRangeForPointsPicker verifies the range of a container that asks for an amount of points to be picked,
// which the qualifier constrains directly.
func TestPointsRangeForPointsPicker(t *testing.T) {
	c := check.New(t)

	// Pick exactly 20 points worth: whatever is chosen, 20 points is what it costs.
	exact := newPickerContainer(picker.Points, criteria.EqualsNumber, 20, 5, 10, 20, 40).PointsRange(nil)
	checkRange(c, 20, 20, exact, "an exact points picker is worth what it asks for")
	c.True(exact.IsSettled(), "an exact points picker is settled")

	// The same holds for the disadvantage packages templates use: "pick -50 points worth".
	checkRange(c, -50, -50, newPickerContainer(picker.Points, criteria.EqualsNumber, -50, -10, -20, -40).PointsRange(nil),
		"an exact points picker asking for disadvantages is worth what it asks for")

	// Pick at least 20 points from 10, 20 and 40: 20 is the least that satisfies it, 70 is everything.
	checkRange(c, 20, 70, newPickerContainer(picker.Points, criteria.AtLeastNumber, 20, 10, 20, 40).PointsRange(nil),
		"a minimum points picker spans what it asks for up to everything worth taking")

	// Pick at least 100 points from children whose own costs only add up to 50: those costs set no ceiling on what the
	// container is worth, since the points of a skill or spell picked from it are assigned while picking. This is the
	// shape of 13 of the 23 "at least" pickers in the master library, whose children carry no points at all.
	noCeiling := newPickerContainer(picker.Points, criteria.AtLeastNumber, 100, 10, 40).PointsRange(nil)
	c.NotNil(noCeiling.Min, "a minimum points picker its children cannot bound has a lower limit")
	c.Equal(fxp.FromInteger(100), *noCeiling.Min, "it costs at least what it asks for")
	c.Nil(noCeiling.Max, "a minimum points picker its children cannot bound has no upper limit")
	c.False(noCeiling.IsSettled(), "a cost with no upper limit is never settled")
	c.Equal("100+", noCeiling.String(), "it renders as its minimum and a plus")

	// A container of 0-point skills is the everyday form of that, and it must not report a settled cost.
	skills := NewSkill(nil, nil, true)
	skills.TemplatePicker.Type = picker.Points
	skills.TemplatePicker.Qualifier.Compare = criteria.AtLeastNumber
	skills.TemplatePicker.Qualifier.Qualifier = fxp.Sixteen
	for range 3 {
		child := NewSkill(nil, skills, false)
		child.Points = 0
		skills.Children = append(skills.Children, child)
	}
	c.Equal("16+", skills.PointsRange(nil).String(), "a points picker over skills carrying no points has no ceiling")
	c.Equal(fxp.Int(0), skills.AdjustedPoints(nil),
		"and it is left reporting the total of its children, since there is no single cost to collapse to")

	// Pick at most 40 points from 10, 20, 40 and -5: the disadvantage alone is cheapest, and the cap is the most that
	// may be spent even though there is more on offer.
	checkRange(c, -5, 40, newPickerContainer(picker.Points, criteria.AtMostNumber, 40, 10, 20, 40, -5).PointsRange(nil),
		"a maximum points picker is capped by what it allows")

	// Pick at most 100 points from children that only add up to 30: what is there is the real limit.
	checkRange(c, 0, 30, newPickerContainer(picker.Points, criteria.AtMostNumber, 100, 10, 20).PointsRange(nil),
		"a maximum points picker allowing more than there is is limited by what there is")
}

// TestPointsRangeNestedPickers verifies that a picker inside a picker composes, which 93 of the pickers in the master
// library do.
func TestPointsRangeNestedPickers(t *testing.T) {
	c := check.New(t)

	// An outer "pick 1" over two inner containers: one settled at 20 points, one spanning 5 to 40.
	outer := NewTrait(nil, nil, true)
	outer.TemplatePicker.Type = picker.Count
	outer.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
	outer.TemplatePicker.Qualifier.Qualifier = fxp.One
	for _, inner := range []*Trait{
		newPickerContainer(picker.Points, criteria.EqualsNumber, 20, 5, 10, 20),
		newPickerContainer(picker.Count, criteria.EqualsNumber, 1, 5, 40),
	} {
		inner.SetParent(outer)
		outer.Children = append(outer.Children, inner)
	}

	checkRange(c, 5, 40, outer.PointsRange(nil),
		"an outer picker spans the cheapest and costliest outcomes of the pickers inside it")
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
	exact := newPickerContainer(picker.Points, criteria.EqualsNumber, 20, 5, 10, 20, 40)
	c.Equal(fxp.FromInteger(20), exact.AdjustedPoints(nil),
		"a container asking for an exact number of points is worth that many")

	// "Pick 1" of three 10-point children is worth 10, not 30.
	identical := newPickerContainer(picker.Count, criteria.EqualsNumber, 1, 10, 10, 10)
	c.Equal(fxp.FromInteger(10), identical.AdjustedPoints(nil),
		"a container asking for one of several identically priced children is worth one of them")

	// "Pick 1" of children that differ has no single answer, so the long-standing total is left alone.
	varied := newPickerContainer(picker.Count, criteria.EqualsNumber, 1, 10, 20, 40)
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
}

// TestPointsCellForPickers verifies what the points column of a table shows for a container presenting choices: the
// range, explained in the tooltip, or the settled cost with nothing added to the tooltip.
func TestPointsCellForPickers(t *testing.T) {
	c := check.New(t)

	var data CellData
	newPickerContainer(picker.Count, criteria.EqualsNumber, 1, 10, 20, 40).CellData(TraitPointsColumn, &data)
	c.Equal("10~40", data.Primary, "an unsettled container shows the range in the points column")
	c.Contains(data.Tooltip, "choices", "an unsettled cost explains itself in the tooltip")

	data = CellData{}
	newPickerContainer(picker.Points, criteria.EqualsNumber, 20, 5, 10, 40).CellData(TraitPointsColumn, &data)
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
