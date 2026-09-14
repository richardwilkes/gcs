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
	if len(children) == 0 {
		return PointsRangeOf(0)
	}
	compare := tp.Qualifier.Compare.EnsureValid()
	switch tp.Type.EnsureValid() {
	case picker.Count:
		count := max(min(tp.Qualifier.Qualifier.AsInteger[int](), len(children)), 0)
		cheapestFirst := byMin(children)
		costliestFirst := byMax(children)
		switch compare {
		case criteria.EqualsNumber:
			return PointsRange{
				Min: totalOfMins(cheapestFirst[:count]).end(),
				Max: totalOfMaxes(costliestFirst[:count]).end(),
			}
		case criteria.AtLeastNumber:
			// The required count has to be taken whether it helps or not -- "pick at least 2" of a 10-point and a
			// -5-point child is worth 5, not 10 -- and beyond it, anything that costs less than nothing still lowers
			// the total while anything that costs more than nothing still raises it.
			lower := totalOfMins(cheapestFirst[:count])
			for _, one := range cheapestFirst[count:] {
				if one.Min == nil || *one.Min < 0 {
					lower = lower.add(one.Min)
				}
			}
			upper := totalOfMaxes(costliestFirst[:count])
			for _, one := range costliestFirst[count:] {
				if one.Max == nil || *one.Max > 0 {
					upper = upper.add(one.Max)
				}
			}
			return PointsRange{Min: lower.end(), Max: upper.end()}
		case criteria.AtMostNumber:
			// Only what helps each end is taken, and no more of it than the picker allows.
			return PointsRange{
				Min: negativeCount(cheapestFirst, count).end(),
				Max: positiveCount(costliestFirst, count).end(),
			}
		default: // Any number, or an "is not" that no selection is meaningfully constrained by.
			return PointsRange{
				Min: everythingNegative(children).end(),
				Max: everythingPositive(children).end(),
			}
		}
	case picker.Points:
		qualifier := tp.Qualifier.Qualifier
		switch compare {
		case criteria.EqualsNumber:
			return PointsRangeOf(qualifier)
		case criteria.AtLeastNumber:
			upper := everythingPositive(children)
			// Children that between them cost less than the picker asks for set no ceiling on it. That isn't a
			// malformed template: the children of such a picker are typically skills or spells carrying no points at
			// all, whose cost is assigned while picking (see ux.pickerRowPointEditor), and a leveled trait's cost can
			// be raised there too. With nothing to bound it, the container is worth what it asks for or more.
			if !upper.unlimited && upper.total < qualifier {
				return pointsRangeAtLeast(qualifier)
			}
			return PointsRange{Min: &qualifier, Max: upper.end()}
		case criteria.AtMostNumber:
			upper := everythingPositive(children)
			if !upper.unlimited && upper.total > qualifier {
				upper.total = qualifier
			}
			return PointsRange{Min: everythingNegative(children).end(), Max: upper.end()}
		default: // Any number, or an "is not" that no selection is meaningfully constrained by.
			return PointsRange{
				Min: everythingNegative(children).end(),
				Max: everythingPositive(children).end(),
			}
		}
	default:
		return sumPointsRanges(children)
	}
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

// negativeCount returns the total of up to count children that cost less than nothing, which is the cheapest a picker
// allowing at most that many can be. The children must already be ordered cheapest first.
func negativeCount(cheapestFirst []PointsRange, count int) pointsBound {
	var result pointsBound
	for _, one := range cheapestFirst[:count] {
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

// positiveCount returns the total of up to count children that cost more than nothing, which is the costliest a picker
// allowing at most that many can be. The children must already be ordered costliest first.
func positiveCount(costliestFirst []PointsRange, count int) pointsBound {
	var result pointsBound
	for _, one := range costliestFirst[:count] {
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

// everythingNegative returns the total of every child that costs less than nothing.
func everythingNegative(ranges []PointsRange) pointsBound {
	var result pointsBound
	for _, one := range ranges {
		if one.Min == nil || *one.Min < 0 {
			result = result.add(one.Min)
		}
	}
	return result
}

// everythingPositive returns the total of every child that costs more than nothing.
func everythingPositive(ranges []PointsRange) pointsBound {
	var result pointsBound
	for _, one := range ranges {
		if one.Max == nil || *one.Max > 0 {
			result = result.add(one.Max)
		}
	}
	return result
}

// How a range is punctuated. These are symbols rather than prose, so they are not run through i18n: a translated
// "%s%s" would be a catalog key with no content to translate, and reordering its two ends would silently break
// PointsLessFromString, which reads the lower end back out of the rendered text.
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
// by its lower end, ties falling to the text itself, so that "10" comes ahead of "10~15" and both come ahead of "20".
// A plain numeric comparison cannot be used: it reads the whole string, so every range would come back as zero and
// sort as equal.
func PointsLessFromString(a, b string) bool {
	aValue, aUnlimited := pointsSortValue(a)
	bValue, bUnlimited := pointsSortValue(b)
	if aUnlimited != bUnlimited {
		return aUnlimited
	}
	if !aUnlimited && aValue != bValue {
		return aValue < bValue
	}
	return xstrings.NaturalLess(a, b, true)
}

// pointsSortValue returns the lower end of a rendered point cost, along with whether that end has no limit, in which
// case nothing finite sorts ahead of it.
func pointsSortValue(text string) (value fxp.Int, unlimited bool) {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, unboundedMinPrefix) || text == noLimitsAtAll {
		return 0, true
	}
	value, _ = fxp.Extract(strings.ReplaceAll(text, ",", ""))
	return value, false
}
