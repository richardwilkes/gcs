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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

// PointsRange is the span of point costs an item may end up costing once every choice it defines has been made. On a
// character sheet every such choice has already been made, so a range there is always settled; it is in a library or on
// a template that an item can still be worth "20 to 45 points", depending on what the player picks.
//
// A settled range -- one whose minimum and maximum are the same known value -- is what the vast majority of items
// report, and it displays exactly as the bare point cost always has.
type PointsRange struct {
	// Min is the least the item can cost, or nil if there is no lower limit.
	Min *fxp.Int
	// Max is the most the item can cost, or nil if there is no upper limit.
	Max *fxp.Int
}

// PointsRangeSign is an enum for the sign of a PointsRange
type PointsRangeSign byte

const (
	// PointsRangePositive represents a PointsRange where Min/Max are both positive
	PointsRangePositive PointsRangeSign = iota
	// PointsRangeNegative represents a PointsRange where Min/Max are both negative
	PointsRangeNegative
	// PointsRangeZero represents a PointsRange where Min/Max are both zero
	PointsRangeZero
	// PointsRangeMixed represents a PointsRange where Min/Max are split between positive and negative
	PointsRangeMixed
)

// PointsRangeOf returns the settled range for an item whose cost is already known.
func PointsRangeOf(value fxp.Int) PointsRange {
	minimum := value
	maximum := value
	return PointsRange{Min: &minimum, Max: &maximum}
}

// newPointsRange returns the range between two known costs.
func newPointsRange(minimum, maximum fxp.Int) PointsRange {
	return PointsRange{Min: &minimum, Max: &maximum}
}

// pointsRangeAtLeast returns the range of an item that costs at least the given amount, with no upper limit.
func pointsRangeAtLeast(minimum fxp.Int) PointsRange {
	return PointsRange{Min: &minimum}
}

// pointsRangeAtMost returns the range of an item that costs at most the given amount, with no lower limit.
func pointsRangeAtMost(maximum fxp.Int) PointsRange {
	return PointsRange{Max: &maximum}
}

// IsSettled returns true if the range is a single known value, i.e. there is nothing left to decide.
func (r PointsRange) IsSettled() bool {
	return r.Min != nil && r.Max != nil && *r.Min == *r.Max
}

// Settled returns the single cost of a settled range. The second return is false for a range that still has something
// left to decide, in which case there is no single cost to return.
func (r PointsRange) Settled() (value fxp.Int, settled bool) {
	if !r.IsSettled() {
		return 0, false
	}
	return *r.Min, true
}

// Sign returns if the range is positive, negative, zero, or mixed
func (r PointsRange) Sign() PointsRangeSign {
	if r.Min != nil && r.Max != nil {
		switch {
		case *r.Min == 0 && *r.Max == 0:
			return PointsRangeZero
		case *r.Min >= 0 && *r.Max > 0:
			return PointsRangePositive
		case *r.Min < 0 && *r.Max <= 0:
			return PointsRangeNegative
		}
	} else {
		switch {
		case r.Min != nil && r.Max == nil && *r.Min >= 0:
			return PointsRangePositive
		case r.Min == nil && r.Max != nil && *r.Max <= 0:
			return PointsRangeNegative
		}
	}
	return PointsRangeMixed
}

// Add returns the result of adding another range to this one. An end with no limit stays that way, since nothing that
// can be added to it brings it back within one.
func (r PointsRange) Add(other PointsRange) PointsRange {
	var lower, upper pointsBound
	return PointsRange{
		Min: lower.add(r.Min).add(other.Min).end(),
		Max: upper.add(r.Max).add(other.Max).end(),
	}
}

// CanSatisfy returns true if some cost within the range meets the criteria. For a settled range this is the plain
// comparison; for one that is still open it asks whether the choices left to make could still come out right, which is
// what a picker in the middle of being filled in needs to know.
func (r PointsRange) CanSatisfy(n criteria.Number) bool {
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

func (r PointsRange) String() string {
	return r.format(func(value fxp.Int) string { return value.String() })
}

// Comma returns the same text as String, but with commas inserted into the numbers.
func (r PointsRange) Comma() string {
	return r.format(func(value fxp.Int) string { return value.Comma() })
}

// format renders the range, using f to render each end of it. A settled range renders as the bare number, so that
// everything which isn't a choice looks exactly as it always has.
func (r PointsRange) format(f func(fxp.Int) string) string {
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
		return f(*r.Min) + rangeSeparator + f(*r.Max)
	}
}

// pointsBound is one end of a range while it is being worked out: a running total, plus whether anything with no limit
// has gone into it, which nothing added afterwards can take back.
type pointsBound struct {
	total     fxp.Int
	unlimited bool
}

// add folds one end of another range into this one.
func (b pointsBound) add(value *fxp.Int) pointsBound {
	if value == nil {
		b.unlimited = true
		return b
	}
	b.total += *value
	return b
}

// end returns the end of a range this bound describes.
func (b pointsBound) end() *fxp.Int {
	if b.unlimited {
		return nil
	}
	total := b.total
	return &total
}

// settledPickerCost returns what a container carrying the given picker is worth no matter which of its children are
// picked, when that can be known without looking at them at all: "pick 20 points worth" is worth 20 whichever way it
// is satisfied. It is the most common picker there is, and answering it without descending into the children spares
// every render and every sort comparison that walk.
func settledPickerCost(tp TemplatePicker) (value fxp.Int, settled bool) {
	if tp.Type.EnsureValid() == picker.Points && tp.Qualifier.Compare.EnsureValid() == criteria.EqualsNumber {
		return tp.Qualifier.Qualifier, true
	}
	return 0, false
}

// pointsRangeNode is a constraint for a Node whose cost can be asked for, which is every node a template picker can
// appear on.
type pointsRangeNode[T pointsRangeNode[T]] interface {
	Node[T]
	AdjustedPoints(tooltip *xbytes.InsertBuffer) fxp.Int
	PointsRange(tooltip *xbytes.InsertBuffer) PointsRange
}

// childPointsRanges returns the range of every child. A container's own range, and the total it falls back to when
// that range is left unsettled, are both worked out from these, so they are only ever built once per container.
func childPointsRanges[T pointsRangeNode[T]](children []T) []PointsRange {
	ranges := make([]PointsRange, len(children))
	for i, one := range children {
		ranges[i] = one.PointsRange(nil)
	}
	return ranges
}

// rawPointsRangeNode is a constraint for a Node whose cost before any bonus can be asked for, which is every skill and
// spell.
type rawPointsRangeNode[T rawPointsRangeNode[T]] interface {
	Node[T]
	RawPointsRange() PointsRange
}

// containerRawPointsRange returns the range of a container carrying the given picker, worked out as PointsRange works
// it out, but from what its children cost before any bonus the sheet they are on grants them.
func containerRawPointsRange[T rawPointsRangeNode[T]](tp TemplatePicker, children []T) PointsRange {
	if value, settled := settledPickerCost(tp); settled {
		return PointsRangeOf(value)
	}
	ranges := make([]PointsRange, len(children))
	for i, one := range children {
		ranges[i] = one.RawPointsRange()
	}
	return pointsRangeForPicker(tp, ranges)
}

// pickerContainerPoints returns what a container carrying the given picker is worth, which is never what its children
// add up to, since only some of them will be taken. When every way of making the choice costs the same -- "pick 20
// points worth", most often -- that is what it is worth. When they don't, there is no single answer, and the total of
// the children is left as the answer AdjustedPoints has always given here, with PointsRange holding the one that can
// be relied upon.
//
// The children are walked once. The range that comes out of that walk answers both questions: whether the choice has
// a single cost after all, and, when it doesn't, what the children add up to.
func pickerContainerPoints[T pointsRangeNode[T]](tp TemplatePicker, children []T) fxp.Int {
	if value, settled := settledPickerCost(tp); settled {
		return value
	}
	ranges := childPointsRanges(children)
	if value, settled := pointsRangeForPicker(tp, ranges).Settled(); settled {
		return value
	}
	return totalOfAdjustedPoints(children, ranges)
}

// totalOfAdjustedPoints returns the total the children cost, given the ranges already worked out for them. A settled
// range is exactly what AdjustedPoints reports for that child, so only a child still presenting a choice of its own
// has to be walked a second time.
func totalOfAdjustedPoints[T pointsRangeNode[T]](children []T, ranges []PointsRange) fxp.Int {
	var total fxp.Int
	for i, one := range children {
		if value, settled := ranges[i].Settled(); settled {
			total += value
			continue
		}
		total += one.AdjustedPoints(nil)
	}
	return total
}

// sumPointsRanges returns the total of the ranges, which is what a container that presents no choice costs: everything
// inside it is taken.
func sumPointsRanges(ranges []PointsRange) PointsRange {
	var lower, upper pointsBound
	for _, one := range ranges {
		lower = lower.add(one.Min)
		upper = upper.add(one.Max)
	}
	return PointsRange{Min: lower.end(), Max: upper.end()}
}

// pointsRangeForPicker returns the range of costs a container carrying the given template picker may end up being
// worth, given the ranges of the children that may be picked from.
//
// The count cases are exact: the cheapest way to satisfy "pick 3" is the 3 cheapest children, and a child that costs
// less than nothing is always worth taking when the picker allows more to be taken. The points cases lean on the fact
// that a points picker measures the very quantity it constrains -- the cost of what is picked -- so the qualifier
// bounds the container's total directly, with no need to search for a subset that adds up to it. The cost of that
// shortcut is that achievability isn't checked: "pick at least 7 points" from children worth 5 and 10 reports a
// minimum of 7, where the cheapest satisfying pick is really 10. Reporting the constraint the picker states is both
// cheaper and closer to how the picker describes itself.
func pointsRangeForPicker(tp TemplatePicker, children []PointsRange) PointsRange {
	switch tp.Type.EnsureValid() {
	case picker.Count:
		return pointsRangeForPickerByCount(tp.Qualifier, children)
	case picker.Points:
		return pointsRangeForPickerByPoints(tp.Qualifier, children)
	default:
		return sumPointsRanges(children)
	}
}

// pointsRangeForPickerByCount returns the range of a container whose picker constrains how many of its children are
// taken. These cases are exact: the cheapest way to satisfy "pick 3" is the 3 cheapest children, and a child that costs
// less than nothing is always worth taking when the picker allows more to be taken.
func pointsRangeForPickerByCount(cq criteria.Number, children []PointsRange) PointsRange {
	// a picker authored with zero children will always be 0 points for count-based pickers
	if len(children) == 0 {
		return PointsRangeOf(0)
	}

	compare := cq.Compare.EnsureValid()
	count := max(min(cq.Qualifier.AsInteger[int](), len(children)), 0)

	if count == 0 {
		switch compare {
		case criteria.EqualsNumber, criteria.AtMostNumber:
			return PointsRangeOf(0)
		case criteria.AtLeastNumber:
			compare = criteria.AnyNumber
		}
	}

	cheapestFirst := byMin(children)
	costliestFirst := byMax(children)

	switch compare {
	case criteria.EqualsNumber: // Expects an exact number of selections
		// To get a valid range for an exact number of selections, we generate two sums.
		// The first sum is the `Min` side and is the `Min` sum from the *cheapest* {count} children.
		// The second sum is the `Max` side and is the `Max` sum from the *most expensive* {count} children.
		// Both sums respect unbounded ranges
		return PointsRange{
			Min: totalOfMins(cheapestFirst[:count]).end(),
			Max: totalOfMaxes(costliestFirst[:count]).end(),
		}
	case criteria.AtLeastNumber:
		// To get a valid range for an number of selections or more, we generate two sums. {count} is a floor on how
		// many are taken, not a ceiling, so each sum takes the {count} children that suit its end and then whatever
		// of the rest still helps it.
		// The first sum is the `Min` side and is the `Min` from the *cheapest* {count} children, plus the `Min` from
		// the sum of the *remaining* children under 0 points.
		// The second sum is the `Max` side and is the `Max` from the *most expensive* {count} children, plus the
		// `Max` from the sum of the *remaining* children over 0 points.
		// Both sums respect unbounded ranges
		return PointsRange{
			Min: totalOfMins(cheapestFirst[:count]).add(sumUnder(cheapestFirst[count:]).end()).end(),
			Max: totalOfMaxes(costliestFirst[:count]).add(sumOver(costliestFirst[count:]).end()).end(),
		}
	case criteria.AtMostNumber:
		// Only what helps each end is taken, and no more of it than the picker allows.
		return PointsRange{
			Min: sumUnder(cheapestFirst[:count]).end(),
			Max: sumOver(costliestFirst[:count]).end(),
		}
	default: // Any number, or an "is not" that no selection is meaningfully constrained by.
		return PointsRange{
			Min: sumUnder(cheapestFirst).end(),
			Max: sumOver(costliestFirst).end(),
		}
	}
}

// pointsRangeForPickerByPoints returns the range of a container whose picker constrains how many points are spent on
// its children. These cases lean on the fact that a points picker measures the very quantity it constrains -- the cost
// of what is picked -- so the qualifier bounds the container's total directly, with no need to search for a subset that
// adds up to it.
//
// What the children can reach still matters at both ends. Taking nothing is always an option, so the cheapest pick can
// never cost more than nothing and the costliest can never cost less; the qualifier binds only the end it constrains,
// and only as far as the children allow. A qualifier the children cannot reach leaves its end open rather than
// contradicting the other one: such a picker offers a leveled trait whose cost can be raised while picking, or a skill
// or spell whose points are assigned there (see ux.pickerRowPointEditor).
//
// Children that can only cost nothing -- and a picker authored with nothing to pick from -- leave it no side to take
// at all, and a qualifier with no side to bind is not what the container is worth: it costs nothing until something
// on offer can cost something. An exact qualifier is the exception, since it says what the container is worth without
// consulting its children.
func pointsRangeForPickerByPoints(cq criteria.Number, children []PointsRange) PointsRange {
	compare := cq.Compare.EnsureValid()
	if compare == criteria.EqualsNumber {
		return PointsRangeOf(cq.Qualifier)
	}
	switch SignForPointsRanges(children...) {
	case PointsRangePositive:
		if cq.Qualifier < 0 {
			break
		}
		switch compare {
		case criteria.AtLeastNumber:
			return pointsRangeAtLeast(cq.Qualifier)
		case criteria.AtMostNumber:
			return newPointsRange(0, cq.Qualifier)
		default:
			return pointsRangeAtLeast(0)
		}
	case PointsRangeNegative:
		if cq.Qualifier > 0 {
			break
		}
		switch compare {
		case criteria.AtLeastNumber:
			return newPointsRange(cq.Qualifier, 0)
		case criteria.AtMostNumber:
			return pointsRangeAtMost(cq.Qualifier)
		default:
			return pointsRangeAtMost(0)
		}
	case PointsRangeZero:
		return PointsRangeOf(0)
	}

	// This is the degenerate case and we return a fully unbounded range - ideally this never happens (but it could)
	return PointsRange{}
}

// byMin returns the ranges ordered from cheapest to costliest, one with no lower limit coming first, since nothing is
// cheaper than something with no limit to how cheap it is.
func byMin(ranges []PointsRange) []PointsRange {
	sorted := slices.Clone(ranges)
	slices.SortStableFunc(sorted, func(a, b PointsRange) int {
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

// byMax returns the ranges ordered from costliest to cheapest, one with no upper limit coming first, since nothing is
// costlier than something with no limit to how costly it is.
func byMax(ranges []PointsRange) []PointsRange {
	sorted := slices.Clone(ranges)
	slices.SortStableFunc(sorted, func(a, b PointsRange) int {
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
func totalOfMins(ranges []PointsRange) pointsBound {
	var result pointsBound
	for _, one := range ranges {
		result = result.add(one.Min)
	}
	return result
}

// totalOfMaxes returns the total of the most the given children can cost.
func totalOfMaxes(ranges []PointsRange) pointsBound {
	var result pointsBound
	for _, one := range ranges {
		result = result.add(one.Max)
	}
	return result
}

// sumUnder returns the total of the given children that cost less than nothing, which is the cheapest any pick that
// may leave them out can be. The children must already be ordered cheapest first.
func sumUnder(cheapestFirst []PointsRange) pointsBound {
	var result pointsBound
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
func sumOver(costliestFirst []PointsRange) pointsBound {
	var result pointsBound
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

// SignForPointsRanges returns the sign across a slice of ranges
func SignForPointsRanges(ranges ...PointsRange) PointsRangeSign {
	var positive bool
	var negative bool
	for _, r := range ranges {
		switch r.Sign() {
		case PointsRangePositive:
			if negative {
				return PointsRangeMixed
			}
			positive = true
		case PointsRangeNegative:
			if positive {
				return PointsRangeMixed
			}
			negative = true
		case PointsRangeMixed:
			return PointsRangeMixed
		}
	}

	if negative {
		return PointsRangeNegative
	}
	if positive {
		return PointsRangePositive
	}

	// Every range agreed on nothing, or there were no ranges at all.
	return PointsRangeZero
}

// How a range is punctuated. These are symbols rather than prose, so they are not run through i18n: a translated
// "%s%s" would be a catalog key with no content to translate, and reordering its two ends would silently break
// PointsLessFromString, which reads the ends back out of the rendered text.
const (
	// rangeSeparator parts the two ends of a range. Either end can be negative, and a dash between two of them --
	// "-30--20" -- is unreadable, while a dash between two positive costs is easily taken for a single negative one.
	// Spacing the dash out solves both and costs more width than a points column can spare, so a character that can
	// never be read as a sign is used instead, and needs no spaces at all.
	rangeSeparator = "~"
	// unboundedMinPrefix marks a range with no lower limit -- "no more than this much". PointsLessFromString looks for
	// it to order such a range ahead of every finite one.
	unboundedMinPrefix = "≤"
	// unboundedMaxSuffix marks a range with no upper limit -- "this much or more".
	unboundedMaxSuffix = "+"
	// noLimitsAtAll is a range with a limit at neither end, which there is no number to show for.
	noLimitsAtAll = "—"
)

// PointsLessFromString orders the text of two point costs, which is all a table column has to sort by. A range sorts
// by its lower end, then by its upper end, ties falling to the text itself, so that "10" comes ahead of "10~15", which
// comes ahead of "10+", and all of them come ahead of "20". An end with no limit sorts beyond every finite one. A plain
// numeric comparison cannot be used: it reads the whole string, so every range would come back as zero and sort as
// equal.
func PointsLessFromString(a, b string) bool {
	aKey := pointsSortKeyOf(a)
	bKey := pointsSortKeyOf(b)
	if result := compareSortEnds(aKey.lower, bKey.lower, -1); result != 0 {
		return result < 0
	}
	if result := compareSortEnds(aKey.upper, bKey.upper, 1); result != 0 {
		return result < 0
	}
	return xstrings.NaturalLess(a, b, true)
}

// pointsSortKey is the two ends of a rendered point cost, each nil where that end has no limit.
type pointsSortKey struct {
	lower *fxp.Int
	upper *fxp.Int
}

// pointsSortKeyOf reads the ends back out of a rendered point cost.
func pointsSortKeyOf(text string) pointsSortKey {
	text = strings.TrimSpace(text)
	switch {
	case text == noLimitsAtAll:
		return pointsSortKey{}
	case strings.HasPrefix(text, unboundedMinPrefix):
		return pointsSortKey{upper: extractSortEnd(strings.TrimPrefix(text, unboundedMinPrefix))}
	case strings.HasSuffix(text, unboundedMaxSuffix):
		return pointsSortKey{lower: extractSortEnd(strings.TrimSuffix(text, unboundedMaxSuffix))}
	}
	lower, upper, found := strings.Cut(text, rangeSeparator)
	key := pointsSortKey{lower: extractSortEnd(lower)}
	if found {
		key.upper = extractSortEnd(upper)
	} else {
		key.upper = key.lower
	}
	return key
}

// extractSortEnd reads one end of a rendered point cost.
func extractSortEnd(text string) *fxp.Int {
	value, _ := fxp.Extract(text) // Extract reads the commas a rendered cost may carry
	return &value
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
