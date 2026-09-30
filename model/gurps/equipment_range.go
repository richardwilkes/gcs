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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/unison/enums/align"
)

// The value and weight ranges of equipment reuse NumericRange, which is a range of fxp.Int with the picker arithmetic
// already worked out; a weight is an fxp.Int in canonical units. As with points, a range is never unsettled on a sheet,
// where every choice has been made, so there it is the same single value ExtendedValue or ExtendedWeight reports.

// ExtendedValueRange returns the span of extended values this equipment may end up having once every choice within it
// has been made.
func (e *Equipment) ExtendedValueRange() NumericRange {
	return equipmentValue().rangeOf(e, e.Quantity)
}

// ExtendedWeightRange returns the span of extended weights this equipment may end up having once every choice within
// it has been made.
func (e *Equipment) ExtendedWeightRange(defUnits fxp.WeightUnit) NumericRange {
	return e.extendedWeightRange(false, defUnits)
}

// extendedWeightRange is ExtendedWeightRange, counting only the weight that counts for skills when forSkills is true.
func (e *Equipment) extendedWeightRange(forSkills bool, defUnits fxp.WeightUnit) NumericRange {
	return equipmentWeight(forSkills, defUnits).rangeOf(e, e.Quantity)
}

// adjustedValueRange returns the span of values one of this equipment may have, not counting what it holds, while a
// mandatory choice among its modifiers is open.
func (e *Equipment) adjustedValueRange() NumericRange {
	return equipmentValue().oneOf(e, nil)
}

// adjustedWeightRange is adjustedValueRange for weight.
func (e *Equipment) adjustedWeightRange(defUnits fxp.WeightUnit) NumericRange {
	return equipmentWeight(false, defUnits).oneOf(e, nil)
}

// singleValueOf returns the one value a range stands for where a single number is needed: the value itself when the
// range is settled, and otherwise the least it may come to, a value or weight never being less than nothing. This is
// how an equipment choice yet to be made, or an open mandatory modifier choice, counts in the calc block written to a
// file, in scripts and in exports. A template choice of points counts differently, as the total of its options (see
// pickerContainerPoints), since that is what points always reported; the two are to be brought together once template
// choices can be left open outside of templates.
func singleValueOf(r NumericRange) fxp.Int {
	if value, settled := r.Settled(); settled {
		return value
	}
	return lowerEndOf(r)
}

// equipmentMeasure is a quantity a choice of equipment may be made by, other than a count: its value or its weight.
type equipmentMeasure struct {
	kind picker.Type
	// own returns the measure of a single piece of the equipment with the given modifiers, not counting anything it
	// holds. weight is its own adjusted weight, which only a value needs (see CostMultiplier).
	own func(e *Equipment, modifiers []*EquipmentModifier, weight fxp.Int) fxp.Int
	// reduce returns the measure of what a single piece of the equipment with the given modifiers holds, given the
	// range of it before any reduction the equipment makes to it.
	reduce func(e *Equipment, modifiers []*EquipmentModifier, contents NumericRange) NumericRange
	// view says how open modifier choices are costed.
	view choiceView
}

// seenAs returns the measure with open modifier choices costed as view says.
func (m equipmentMeasure) seenAs(view choiceView) equipmentMeasure {
	m.view = view
	return m
}

// equipmentValue returns the measure of equipment's value.
func equipmentValue() equipmentMeasure {
	return equipmentMeasure{
		kind: picker.Value,
		own: func(e *Equipment, modifiers []*EquipmentModifier, weight fxp.Int) fxp.Int {
			return ValueAdjustedForModifiers(e, e.ResolvedBaseValue(), weight, modifiers)
		},
		reduce: func(_ *Equipment, _ []*EquipmentModifier, contents NumericRange) NumericRange { return contents },
	}
}

// equipmentWeight returns the measure of equipment's weight, a weight with no units given being in defUnits. When
// forSkills is true, it measures only the weight that counts for skills.
func equipmentWeight(forSkills bool, defUnits fxp.WeightUnit) equipmentMeasure {
	return equipmentMeasure{
		kind: picker.Weight,
		own: func(e *Equipment, modifiers []*EquipmentModifier, _ fxp.Int) fxp.Int {
			if forSkills && e.WeightIgnoredForSkills && e.ReallyEquipped() {
				return 0
			}
			return fxp.Int(WeightAdjustedForModifiers(e, e.ResolvedBaseWeight(), modifiers, defUnits))
		},
		reduce: func(e *Equipment, modifiers []*EquipmentModifier, contents NumericRange) NumericRange {
			reduction := containedWeightReductionFor(e, defUnits, modifiers, e.Features)
			reduce := func(end *fxp.Int) *fxp.Int {
				if end == nil {
					// A reduction that takes away everything leaves nothing, however much there was to begin with.
					if !reduction.removesEverything() {
						return nil
					}
					var nothing fxp.Int
					return &nothing
				}
				value := fxp.Int(reduction.apply(fxp.Weight(*end)))
				return &value
			}
			return NumericRange{Min: reduce(contents.Min), Max: reduce(contents.Max)}
		},
	}
}

// rangeOf returns the range of the measure of the given quantity of the equipment, what it holds included.
func (m equipmentMeasure) rangeOf(e *Equipment, quantity fxp.Int) NumericRange {
	if quantity <= 0 {
		return NumericRangeOf(0)
	}
	var contents *NumericRange
	if e.Container() {
		r := m.contentsOf(e)
		contents = &r
	}
	return scaleNumericRange(m.oneOf(e, contents), quantity)
}

// oneOf returns the range of the measure of a single piece of the equipment over each way of making the open mandatory
// choices among its modifiers, which may also change how much of what it holds is reduced. contents is the range of
// what it holds before that reduction, or nil to leave what it holds out.
func (m equipmentMeasure) oneOf(e *Equipment, contents *NumericRange) NumericRange {
	var weight fxp.Int
	if m.kind == picker.Value {
		// Worked out once, seen as the value is, rather than for each way of making the choices.
		units := SheetSettingsFor(EntityFromNode(e)).DefaultWeightUnits
		weight = lowerEndOf(equipmentWeight(false, units).seenAs(m.view).oneOf(e, nil))
	}
	eval := func(modifiers []*EquipmentModifier) NumericRange {
		r := NumericRangeOf(m.own(e, modifiers, weight))
		if contents != nil {
			r = r.Add(m.reduce(e, modifiers, *contents))
		}
		return r
	}
	if r, open := modifierChoiceRange(e, e.Modifiers, nil, m.view, eval); open {
		return r
	}
	return eval(e.Modifiers)
}

// contentsOf returns the range of the measure of what a single piece of the container holds, before any reduction the
// container makes to it. A container that isn't a choice holds all of its children.
//
// A choice picked by count takes its options as they are, so its range is worked out just as it is for points.
//
// A choice picked by a measure lets the quantity of each option with a quantity of its own be raised while picking, so
// such an option can always add more, so long as a single one of it measures anything, whatever quantity it starts out
// with. A choice picked by this measure is then bounded by its qualifier directly, just as a points choice is by its
// own; when no option can be raised, it is also held to what its options can come to together, taken or left. A choice
// picked by the other measure has no upper limit to this one once any option can be raised, and otherwise may hold
// anything from none of its options to all of them.
func (m equipmentMeasure) contentsOf(e *Equipment) NumericRange {
	children := make([]NumericRange, len(e.Children))
	for i, one := range e.Children {
		children[i] = m.rangeOf(one, one.Quantity)
	}
	if !IsTemplateChoiceContainer(e) {
		return sumNumericRanges(children)
	}
	if e.TemplatePicker.Type == picker.Count {
		return rangeForPickerByCount(e.TemplatePicker.Qualifier, children)
	}
	// What each option could add is what a single one of it measures when it can be raised, and what it measures as it
	// stands otherwise.
	potential := make([]NumericRange, len(e.Children))
	raisable := false
	for i, one := range e.Children {
		potential[i] = children[i]
		if !one.IsGroup() {
			if unit := m.rangeOf(one, fxp.One); SignForNumericRanges(unit) != NumericRangeZero {
				potential[i] = unit
				raisable = true
			}
		}
	}
	reachable := rangeForPickerByCount(criteria.Number{Compare: criteria.AnyNumber}, children)
	if e.TemplatePicker.Type != m.kind {
		if raisable {
			return numericRangeAtLeast(0)
		}
		return reachable
	}
	bounded := rangeForPickerByMeasure(e.TemplatePicker.Qualifier, potential)
	if raisable {
		return bounded
	}
	return capNumericRange(bounded, reachable)
}

// capNumericRange returns the part of the range within the limits of another. A range with no part within them, as a
// qualifier nothing can reach gives, is returned as it is: whether a choice can be satisfied is not a range's concern.
func capNumericRange(r, limits NumericRange) NumericRange {
	lower := r.Min
	if lower == nil || (limits.Min != nil && *limits.Min > *lower) {
		lower = limits.Min
	}
	upper := r.Max
	if upper == nil || (limits.Max != nil && *limits.Max < *upper) {
		upper = limits.Max
	}
	if lower != nil && upper != nil && *lower > *upper {
		return r
	}
	return NumericRange{Min: lower, Max: upper}
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
		if lower == upper {
			return upper
		}
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
