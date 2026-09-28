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
	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// settledPickerCost returns what a container carrying the given picker is worth no matter which of its children are
// picked, when that can be known without looking at them at all: "pick 20 points worth" is worth 20 whichever way it
// is satisfied. It is the most common picker there is, and answering it without descending into the children spares
// every render and every sort comparison that walk.
func settledPickerCost(tp TemplatePicker) (value fxp.Int, settled bool) {
	if tp.Type == picker.Points && tp.Qualifier.Compare.EnsureValid() == criteria.EqualsNumber {
		return tp.Qualifier.Qualifier, true
	}
	return 0, false
}

// pointsRangeNode is a constraint for a Node whose cost can be asked for, which is every node a template picker can
// appear on.
type pointsRangeNode[T pointsRangeNode[T]] interface {
	Node[T]
	AdjustedPoints(tooltip *xbytes.InsertBuffer) fxp.Int
	PointsRange(tooltip *xbytes.InsertBuffer) NumericRange
}

// childPointsRanges returns the range of every child. A container's own range, and the total it falls back to when
// that range is left unsettled, are both worked out from these, so they are only ever built once per container.
func childPointsRanges[T pointsRangeNode[T]](children []T) []NumericRange {
	ranges := make([]NumericRange, len(children))
	for i, one := range children {
		ranges[i] = one.PointsRange(nil)
	}
	return ranges
}

// rawPointsRangeNode is a constraint for a Node whose cost before any bonus can be asked for, which is every skill and
// spell.
type rawPointsRangeNode[T rawPointsRangeNode[T]] interface {
	Node[T]
	RawPointsRange() NumericRange
}

// containerRawPointsRange returns the range of a container carrying the given picker, worked out as PointsRange works
// it out, but from what its children cost before any bonus the sheet they are on grants them.
func containerRawPointsRange[T rawPointsRangeNode[T]](tp TemplatePicker, children []T) NumericRange {
	if value, settled := settledPickerCost(tp); settled {
		return NumericRangeOf(value)
	}
	ranges := make([]NumericRange, len(children))
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
func totalOfAdjustedPoints[T pointsRangeNode[T]](children []T, ranges []NumericRange) fxp.Int {
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
func pointsRangeForPicker(tp TemplatePicker, children []NumericRange) NumericRange {
	switch tp.Type {
	case picker.Count:
		return rangeForPickerByCount(tp.Qualifier, children)
	case picker.Points:
		return rangeForPickerByMeasure(tp.Qualifier, children)
	default:
		return sumNumericRanges(children)
	}
}

// PointsLessFromString orders the text of two point costs, which is all a table column has to sort by. A range sorts
// by its lower end, then by its upper end, ties falling to the text itself, so that "10" comes ahead of "10~15", which
// comes ahead of "10+", and all of them come ahead of "20". An end with no limit sorts beyond every finite one. A plain
// numeric comparison cannot be used: it reads the whole string, so every range would come back as zero and sort as
// equal.
func PointsLessFromString(a, b string) bool {
	return rangeLessFromString(a, b, extractSortEnd)
}

// extractSortEnd reads one end of a rendered point cost.
func extractSortEnd(text string) *fxp.Int {
	value, _ := fxp.Extract(text) // Extract reads the commas a rendered cost may carry
	return &value
}
