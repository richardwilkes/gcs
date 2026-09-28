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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/unison/enums/align"
)

// The value and weight ranges of equipment reuse NumericRange, which is a range of fxp.Int with the picker arithmetic
// already worked out; a weight is an fxp.Int in canonical units. As with points, a range is only ever unsettled on a
// template, the one place a choice can still be left to make, so everywhere else it is the same single value
// ExtendedValue or ExtendedWeight reports.

// ExtendedValueRange returns the span of extended values this equipment may end up having once every choice within it
// has been made.
func (e *Equipment) ExtendedValueRange() NumericRange {
	if e.Quantity <= 0 {
		return NumericRangeOf(0)
	}
	if !e.Container() {
		return NumericRangeOf(e.ExtendedValue())
	}
	children := make([]NumericRange, len(e.Children))
	for i, one := range e.Children {
		children[i] = one.ExtendedValueRange()
	}
	contents := equipmentContentsRange(e, picker.Value, children)
	return scaleNumericRange(NumericRangeOf(e.AdjustedValue()).Add(contents), e.Quantity)
}

// ExtendedWeightRange returns the span of extended weights this equipment may end up having once every choice within
// it has been made.
func (e *Equipment) ExtendedWeightRange(defUnits fxp.WeightUnit) NumericRange {
	if e.Quantity <= 0 {
		return NumericRangeOf(0)
	}
	if !e.Container() {
		return NumericRangeOf(fxp.Int(e.ExtendedWeight(false, defUnits)))
	}
	children := make([]NumericRange, len(e.Children))
	for i, one := range e.Children {
		children[i] = one.ExtendedWeightRange(defUnits)
	}
	contents := equipmentContentsRange(e, picker.Weight, children)
	reduction := containedWeightReductionFor(e, defUnits, e.Modifiers, e.Features)
	reduce := func(end *fxp.Int) *fxp.Int {
		if end == nil {
			return nil
		}
		value := fxp.Int(reduction.apply(fxp.Weight(*end)))
		return &value
	}
	contents = NumericRange{Min: reduce(contents.Min), Max: reduce(contents.Max)}
	base := WeightAdjustedForModifiers(e, e.ResolvedBaseWeight(), e.Modifiers, defUnits)
	return scaleNumericRange(NumericRangeOf(fxp.Int(base)).Add(contents), e.Quantity)
}

// equipmentContentsRange returns the range of what the container holds, measured as the given picker type measures it,
// given the ranges of its children. A container that isn't a choice holds all of its children.
//
// A choice picked by count takes its options as they are, so its range is worked out just as it is for points. A
// choice picked by the same measure is bounded by its qualifier directly, just as a points choice is by its own. A
// choice picked by the other measure lets the quantity of each option be raised while picking, so there is no upper
// limit to what it may hold, unless nothing it offers has anything to raise.
func equipmentContentsRange(e *Equipment, measure picker.Type, children []NumericRange) NumericRange {
	if !IsTemplateChoiceContainer(e) {
		return sumNumericRanges(children)
	}
	switch e.TemplatePicker.Type {
	case picker.Count:
		return rangeForPickerByCount(e.TemplatePicker.Qualifier, children)
	case measure:
		return rangeForPickerByMeasure(e.TemplatePicker.Qualifier, children)
	default:
		if SignForNumericRanges(children...) == NumericRangeZero {
			return NumericRangeOf(0)
		}
		return numericRangeAtLeast(0)
	}
}

// lowerEndOf returns the least of the range, a value or weight never being less than nothing.
func lowerEndOf(r NumericRange) fxp.Int {
	if r.Min == nil {
		return 0
	}
	return max(*r.Min, 0)
}

// scaleNumericRange returns the range multiplied by the given positive quantity.
func scaleNumericRange(r NumericRange, quantity fxp.Int) NumericRange {
	scale := func(end *fxp.Int) *fxp.Int {
		if end == nil {
			return nil
		}
		value := end.Mul(quantity)
		return &value
	}
	return NumericRange{Min: scale(r.Min), Max: scale(r.Max)}
}

// FormatValueRange renders a range of values, using format to render each end of it.
func FormatValueRange(r NumericRange, format func(fxp.Int) string) string {
	return r.format(format)
}

// FormatWeightRange renders a range of weights, using format to render each end of it. When both ends are in the same
// units, as they almost always are, the units are only given once, after the upper end: "3~5 lb".
func FormatWeightRange(r NumericRange, format func(fxp.Weight) string) string {
	if r.Min != nil && r.Max != nil && *r.Min != *r.Max {
		lower := format(fxp.Weight(*r.Min))
		upper := format(fxp.Weight(*r.Max))
		if number, units, found := strings.Cut(lower, " "); found {
			if _, upperUnits, upperFound := strings.Cut(upper, " "); upperFound && units == upperUnits {
				return number + rangeSeparator + upper
			}
		}
	}
	return r.format(func(value fxp.Int) string { return format(fxp.Weight(value)) })
}

// ValueRangeLessFromString orders the text of two values, either of which may be a range, as rendered by
// FormatValueRange.
func ValueRangeLessFromString(a, b string) bool {
	return rangeLessFromString(a, b, func(text string) *fxp.Int {
		value := fxp.FromStringForced(text)
		return &value
	})
}

// WeightRangeLessFromStringFunc returns a func that orders the text of two weights, either of which may be a range, as
// rendered by FormatWeightRange.
func WeightRangeLessFromStringFunc(units fxp.WeightUnit) func(a, b string) bool {
	return func(a, b string) bool {
		return rangeLessFromString(expandWeightRange(a), expandWeightRange(b), func(text string) *fxp.Int {
			value := fxp.Int(fxp.WeightFromStringForced(text, units))
			return &value
		})
	}
}

// expandWeightRange gives the lower end of a rendered weight range the units FormatWeightRange leaves off it, so that
// "3~5 lb" reads back as "3 lb~5 lb". Anything else is returned as it is.
func expandWeightRange(text string) string {
	lower, upper, found := strings.Cut(text, rangeSeparator)
	if !found || strings.Contains(lower, " ") {
		return text
	}
	if _, units, hasUnits := strings.Cut(upper, " "); hasUnits {
		return lower + " " + units + rangeSeparator + upper
	}
	return text
}

// valueRangeCellData fills in the cell data for a range of monetary values, just as valueCellData does for a single
// one, which is what a settled range is shown as.
func (e *Equipment) valueRangeCellData(data *CellData, r NumericRange) {
	if value, settled := r.Settled(); settled {
		e.valueCellData(data, value)
		return
	}
	data.Type = cell.Text
	data.Alignment = align.End
	data.Primary = FormatValueRange(r, fxp.Int.Comma)
	if data.ForPage {
		if text := FormatValueRange(r, SheetSettingsFor(EntityFromNode(e)).FormatEquipmentValue); text != data.Primary {
			data.Tooltip = data.Primary
			data.Primary = text
		}
	}
}

// weightRangeCellData fills in the cell data for a range of weights, just as weightCellData does for a single one,
// which is what a settled range is shown as.
func (e *Equipment) weightRangeCellData(data *CellData, weigh func(defUnits fxp.WeightUnit) NumericRange) {
	settings := SheetSettingsFor(EntityFromNode(e))
	r := weigh(settings.DefaultWeightUnits)
	if value, settled := r.Settled(); settled {
		e.weightCellData(data, func(_ bool, _ fxp.WeightUnit) fxp.Weight { return fxp.Weight(value) })
		return
	}
	data.Type = cell.Text
	data.Alignment = align.End
	data.Primary = FormatWeightRange(r, settings.DefaultWeightUnits.Format)
	if data.ForPage {
		if text := FormatWeightRange(r, settings.FormatEquipmentWeight); text != data.Primary {
			data.Tooltip = data.Primary
			data.Primary = text
		}
	}
}

// equipmentRangeTotals returns the total extended weight and value of the equipment, as ranges.
func equipmentRangeTotals(list []*Equipment, defUnits fxp.WeightUnit) (weight, value NumericRange) {
	weight = NumericRangeOf(0)
	value = NumericRangeOf(0)
	for _, one := range list {
		weight = weight.Add(one.ExtendedWeightRange(defUnits))
		value = value.Add(one.ExtendedValueRange())
	}
	return weight, value
}
