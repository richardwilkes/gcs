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
	"cmp"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

// NumericRange is the span of values something may end up having once every choice it defines has been made: the
// points an item costs, or the extended value or weight of a piece of equipment. On a character sheet every such
// choice has already been made, so a range there is always settled; in a library or on a template an item can still be
// worth "20 to 45 points", depending on what the player picks. A settled range displays as the bare value.
//
// Where a single number must stand for an unsettled range, as in a file's calc block, points and equipment differ for
// now; see singleValueOf.
type NumericRange struct {
	// Min is the least the item can cost, or nil if there is no lower limit.
	Min *fxp.Int
	// Max is the most the item can cost, or nil if there is no upper limit.
	Max *fxp.Int
}

// NumericRangeSign is an enum for the sign of a NumericRange.
type NumericRangeSign byte

const (
	// NumericRangePositive represents a NumericRange with no negative values and some positive ones.
	NumericRangePositive NumericRangeSign = iota
	// NumericRangeNegative represents a NumericRange with no positive values and some negative ones.
	NumericRangeNegative
	// NumericRangeZero represents a NumericRange where Min/Max are both zero.
	NumericRangeZero
	// NumericRangeMixed represents a NumericRange that spans both negative and positive values.
	NumericRangeMixed
)

// NumericRangeOf returns the settled range for an item whose cost is already known.
func NumericRangeOf(value fxp.Int) NumericRange {
	minimum := value
	maximum := value
	return NumericRange{Min: &minimum, Max: &maximum}
}

// newNumericRange returns the range between two known costs.
func newNumericRange(minimum, maximum fxp.Int) NumericRange {
	return NumericRange{Min: &minimum, Max: &maximum}
}

// numericRangeAtLeast returns the range of an item that costs at least the given amount, with no upper limit.
func numericRangeAtLeast(minimum fxp.Int) NumericRange {
	return NumericRange{Min: &minimum}
}

// numericRangeAtMost returns the range of an item that costs at most the given amount, with no lower limit.
func numericRangeAtMost(maximum fxp.Int) NumericRange {
	return NumericRange{Max: &maximum}
}

// IsSettled returns true if the range is a single known value, i.e. there is nothing left to decide.
func (r NumericRange) IsSettled() bool {
	return r.Min != nil && r.Max != nil && *r.Min == *r.Max
}

// Settled returns the single cost of a settled range, or false if the range is not settled.
func (r NumericRange) Settled() (value fxp.Int, settled bool) {
	if !r.IsSettled() {
		return 0, false
	}
	return *r.Min, true
}

// Sign returns whether the range is positive, negative, zero, or mixed.
func (r NumericRange) Sign() NumericRangeSign {
	if r.Min != nil && r.Max != nil {
		switch {
		case *r.Min == 0 && *r.Max == 0:
			return NumericRangeZero
		case *r.Min >= 0 && *r.Max > 0:
			return NumericRangePositive
		case *r.Min < 0 && *r.Max <= 0:
			return NumericRangeNegative
		}
	} else {
		switch {
		case r.Min != nil && r.Max == nil && *r.Min >= 0:
			return NumericRangePositive
		case r.Min == nil && r.Max != nil && *r.Max <= 0:
			return NumericRangeNegative
		}
	}
	return NumericRangeMixed
}

// Add returns the result of adding another range to this one. An end with no limit stays unlimited.
func (r NumericRange) Add(other NumericRange) NumericRange {
	var lower, upper numericBound
	return NumericRange{
		Min: lower.add(r.Min).add(other.Min).end(),
		Max: upper.add(r.Max).add(other.Max).end(),
	}
}

// CanSatisfy returns true if some cost within the range meets the criteria: for an open range, whether the choices left
// to make, as in a picker being filled in, could still come out right.
func (r NumericRange) CanSatisfy(n criteria.Number) bool {
	if value, settled := r.Settled(); settled {
		return n.Matches(value)
	}
	switch n.Compare.EnsureValid() {
	case criteria.EqualsNumber:
		return (r.Min == nil || *r.Min <= n.Qualifier) && (r.Max == nil || *r.Max >= n.Qualifier)
	case criteria.NotEqualsNumber:
		// An open range always holds something other than any single value.
		return true
	case criteria.AtLeastNumber:
		return r.Max == nil || *r.Max >= n.Qualifier
	case criteria.AtMostNumber:
		return r.Min == nil || *r.Min <= n.Qualifier
	default:
		return true
	}
}

func (r NumericRange) String() string {
	return r.format(func(value fxp.Int) string { return value.String() })
}

// Comma returns the same text as String, but with commas inserted into the numbers.
func (r NumericRange) Comma() string {
	return r.format(func(value fxp.Int) string { return value.Comma() })
}

// format renders the range, using f to render each end. A settled range, or one whose ends render the same (as they can
// when f rounds), renders as the bare number.
func (r NumericRange) format(f func(fxp.Int) string) string {
	switch {
	case r.Min == nil && r.Max == nil:
		return noLimitsAtAll
	case r.Min == nil:
		return unboundedMinPrefix + f(*r.Max)
	case r.Max == nil:
		return f(*r.Min) + unboundedMaxSuffix
	case *r.Min == *r.Max:
		return f(*r.Min)
	default:
		lower := f(*r.Min)
		upper := f(*r.Max)
		if lower == upper {
			return lower
		}
		return lower + rangeSeparator + upper
	}
}

// numericBound is one end of a range while it is being worked out: a running total, plus whether anything with no limit
// has gone into it, which nothing added afterwards can take back.
type numericBound struct {
	total     fxp.Int
	unlimited bool
}

// add folds one end of another range into this one.
func (b numericBound) add(value *fxp.Int) numericBound {
	if value == nil {
		b.unlimited = true
		return b
	}
	b.total += *value
	return b
}

// end returns the end of a range this bound describes.
func (b numericBound) end() *fxp.Int {
	if b.unlimited {
		return nil
	}
	total := b.total
	return &total
}

// sumNumericRanges returns the total of the ranges, which is what a container that presents no choice costs: everything
// inside it is taken.
func sumNumericRanges(ranges []NumericRange) NumericRange {
	var lower, upper numericBound
	for _, one := range ranges {
		lower = lower.add(one.Min)
		upper = upper.add(one.Max)
	}
	return NumericRange{Min: lower.end(), Max: upper.end()}
}

// rangeForPickerByCount returns the range of a container whose picker constrains how many of its children are
// taken. These cases are exact: the cheapest way to satisfy "pick 3" is the 3 cheapest children, and a child that costs
// less than nothing is always worth taking when the picker allows more to be taken.
func rangeForPickerByCount(cq criteria.Number, children []NumericRange) NumericRange {
	if len(children) == 0 {
		return NumericRangeOf(0)
	}

	compare := cq.Compare.EnsureValid()
	count := max(min(cq.Qualifier.AsInteger[int](), len(children)), 0)

	if count == 0 {
		switch compare {
		case criteria.EqualsNumber, criteria.AtMostNumber:
			return NumericRangeOf(0)
		case criteria.AtLeastNumber:
			compare = criteria.AnyNumber
		}
	}

	cheapestFirst := byMin(children)
	costliestFirst := byMax(children)

	switch compare {
	case criteria.EqualsNumber: // Expects an exact number of selections
		// Min totals the Mins of the {count} cheapest children; Max totals the Maxes of the {count} costliest.
		return NumericRange{
			Min: totalOfMins(cheapestFirst[:count]).end(),
			Max: totalOfMaxes(costliestFirst[:count]).end(),
		}
	case criteria.AtLeastNumber:
		// {count} is a floor, not a ceiling, so each end takes the {count} children that suit it and then whatever of
		// the rest still helps it: Min adds the remaining children under 0, Max those over 0.
		return NumericRange{
			Min: totalOfMins(cheapestFirst[:count]).add(sumUnder(cheapestFirst[count:]).end()).end(),
			Max: totalOfMaxes(costliestFirst[:count]).add(sumOver(costliestFirst[count:]).end()).end(),
		}
	case criteria.AtMostNumber:
		// Only what helps each end is taken, and no more of it than the picker allows.
		return NumericRange{
			Min: sumUnder(cheapestFirst[:count]).end(),
			Max: sumOver(costliestFirst[:count]).end(),
		}
	default: // Any number, or an "is not" that no selection is meaningfully constrained by.
		return NumericRange{
			Min: sumUnder(cheapestFirst).end(),
			Max: sumOver(costliestFirst).end(),
		}
	}
}

// rangeForPickerByMeasure returns the range of a container whose picker constrains the total of the quantity the range
// measures: the points spent on its children, or the value or weight of the equipment taken from them. The qualifier
// thus bounds the container's total directly, with no need to search for a subset that adds up to it.
//
// Only which side of nothing the children fall on is consulted. Taking nothing is always an option, so an end the
// qualifier leaves unconstrained is nothing on the side facing zero and open on the other. An open end stays open even
// when the children cannot reach the qualifier: such a picker offers a leveled trait whose cost can be raised while
// picking, or a skill or spell whose points are assigned there (see ux.pickerRowPointEditor). A qualifier on the far
// side of nothing from every child is an invalid picker, even where every pick would happen to meet it, and is left
// open at both ends so that it stands out rather than passing for an unconstrained one. Children on both sides of
// nothing are left open at both ends as well.
//
// Children that can only cost nothing -- and a picker with nothing to pick from -- leave the qualifier nothing to bind,
// so the container costs nothing until something on offer can cost something. An exact qualifier is the exception,
// since it says what the container is worth without consulting its children.
func rangeForPickerByMeasure(cq criteria.Number, children []NumericRange) NumericRange {
	compare := cq.Compare.EnsureValid()
	if compare == criteria.EqualsNumber {
		return NumericRangeOf(cq.Qualifier)
	}
	switch SignForNumericRanges(children...) {
	case NumericRangePositive:
		if cq.Qualifier < 0 {
			break
		}
		switch compare {
		case criteria.AtLeastNumber:
			return numericRangeAtLeast(cq.Qualifier)
		case criteria.AtMostNumber:
			return newNumericRange(0, cq.Qualifier)
		default:
			return numericRangeAtLeast(0)
		}
	case NumericRangeNegative:
		if cq.Qualifier > 0 {
			break
		}
		switch compare {
		case criteria.AtLeastNumber:
			return newNumericRange(cq.Qualifier, 0)
		case criteria.AtMostNumber:
			return numericRangeAtMost(cq.Qualifier)
		default:
			return numericRangeAtMost(0)
		}
	case NumericRangeZero:
		return NumericRangeOf(0)
	}

	// Children on both sides of nothing, or a qualifier on the far side of nothing from them.
	return NumericRange{}
}

// byMin returns the ranges ordered from cheapest to costliest, those with no lower limit first.
func byMin(ranges []NumericRange) []NumericRange {
	sorted := slices.Clone(ranges)
	slices.SortStableFunc(sorted, func(a, b NumericRange) int {
		if (a.Min == nil) != (b.Min == nil) {
			if a.Min == nil {
				return -1
			}
			return 1
		}
		if a.Min == nil {
			return 0
		}
		return cmp.Compare(*a.Min, *b.Min)
	})
	return sorted
}

// byMax returns the ranges ordered from costliest to cheapest, those with no upper limit first.
func byMax(ranges []NumericRange) []NumericRange {
	sorted := slices.Clone(ranges)
	slices.SortStableFunc(sorted, func(a, b NumericRange) int {
		if (a.Max == nil) != (b.Max == nil) {
			if a.Max == nil {
				return -1
			}
			return 1
		}
		if a.Max == nil {
			return 0
		}
		return cmp.Compare(*b.Max, *a.Max)
	})
	return sorted
}

// totalOfMins returns the total of the least the given children can cost.
func totalOfMins(ranges []NumericRange) numericBound {
	var result numericBound
	for _, one := range ranges {
		result = result.add(one.Min)
	}
	return result
}

// totalOfMaxes returns the total of the most the given children can cost.
func totalOfMaxes(ranges []NumericRange) numericBound {
	var result numericBound
	for _, one := range ranges {
		result = result.add(one.Max)
	}
	return result
}

// sumUnder returns the total of the given children that cost less than nothing, which is the cheapest any pick that
// may leave them out can be. The children must already be ordered cheapest first.
func sumUnder(cheapestFirst []NumericRange) numericBound {
	var result numericBound
	for _, one := range cheapestFirst {
		if one.Min == nil {
			result = result.add(nil)
			continue
		}
		if *one.Min >= 0 {
			break // Ordered cheapest first, so nothing past this one costs less than nothing either.
		}
		result = result.add(one.Min)
	}
	return result
}

// sumOver returns the total of the given children that cost more than nothing, which is the costliest any pick that
// may leave them out can be. The children must already be ordered costliest first.
func sumOver(costliestFirst []NumericRange) numericBound {
	var result numericBound
	for _, one := range costliestFirst {
		if one.Max == nil {
			result = result.add(nil)
			continue
		}
		if *one.Max <= 0 {
			break // Ordered costliest first, so nothing past this one costs more than nothing either.
		}
		result = result.add(one.Max)
	}
	return result
}

// SignForNumericRanges returns the sign across a slice of ranges.
func SignForNumericRanges(ranges ...NumericRange) NumericRangeSign {
	var positive bool
	var negative bool
	for _, r := range ranges {
		switch r.Sign() {
		case NumericRangePositive:
			if negative {
				return NumericRangeMixed
			}
			positive = true
		case NumericRangeNegative:
			if positive {
				return NumericRangeMixed
			}
			negative = true
		case NumericRangeMixed:
			return NumericRangeMixed
		}
	}

	if negative {
		return NumericRangeNegative
	}
	if positive {
		return NumericRangePositive
	}

	// Every range agreed on nothing, or there were no ranges at all.
	return NumericRangeZero
}

// How a range is punctuated. These are symbols, not prose, so they are not run through i18n: a translation that
// reordered the ends would silently break PointsLessFromString, which reads them back out of the rendered text.
const (
	// rangeSeparator parts the two ends of a range. A dash would be unreadable between negative ends ("-30--20") and
	// mistaken for a sign between positive ones, and spacing it out costs more width than a points column can spare, so
	// a character that can never be read as a sign is used instead.
	rangeSeparator = "~"
	// unboundedMinPrefix marks a range with no lower limit -- "no more than this much". PointsLessFromString looks for
	// it to order such a range ahead of every finite one.
	unboundedMinPrefix = "≤"
	// unboundedMaxSuffix marks a range with no upper limit -- "this much or more".
	unboundedMaxSuffix = "+"
	// noLimitsAtAll is a range with a limit at neither end, which there is no number to show for.
	noLimitsAtAll = "—"
)

// rangeLessFromString orders the text of two ranges as PointsLessFromString does, using extract to read each end of
// them back out.
func rangeLessFromString(a, b string, extract func(text string) *fxp.Int) bool {
	aKey := rangeSortKeyOf(a, extract)
	bKey := rangeSortKeyOf(b, extract)
	if result := compareSortEnds(aKey.lower, bKey.lower, -1); result != 0 {
		return result < 0
	}
	if result := compareSortEnds(aKey.upper, bKey.upper, 1); result != 0 {
		return result < 0
	}
	return xstrings.NaturalLess(a, b, true)
}

// rangeSortKey is the two ends of a rendered range, each nil where that end has no limit.
type rangeSortKey struct {
	lower *fxp.Int
	upper *fxp.Int
}

// rangeSortKeyOf reads the ends back out of a rendered range, using extract to read each of them.
func rangeSortKeyOf(text string, extract func(text string) *fxp.Int) rangeSortKey {
	text = strings.TrimSpace(text)
	switch {
	case text == noLimitsAtAll:
		return rangeSortKey{}
	case strings.HasPrefix(text, unboundedMinPrefix):
		return rangeSortKey{upper: extract(strings.TrimPrefix(text, unboundedMinPrefix))}
	case strings.HasSuffix(text, unboundedMaxSuffix):
		return rangeSortKey{lower: extract(strings.TrimSuffix(text, unboundedMaxSuffix))}
	}
	lower, upper, found := strings.Cut(text, rangeSeparator)
	key := rangeSortKey{lower: extract(lower)}
	if found {
		key.upper = extract(upper)
	} else {
		key.upper = key.lower
	}
	return key
}

// compareSortEnds compares the same end of two rendered point costs. An end with no limit sorts to the side given by
// unlimited: -1 ahead of every finite end, 1 after them.
func compareSortEnds(a, b *fxp.Int, unlimited int) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return unlimited
	case b == nil:
		return -unlimited
	default:
		return cmp.Compare(*a, *b)
	}
}
