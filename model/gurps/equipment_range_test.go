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
	"encoding/json/v2"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emcost"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emweight"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/check"
)

// expectedExtendedValue works out the extended value of equipment presenting no choice the way it has always been
// worked out, independently of the ranges it now comes from.
func expectedExtendedValue(e *Equipment) fxp.Int {
	if e.Quantity <= 0 {
		return 0
	}
	value := e.AdjustedValue()
	for _, one := range e.Children {
		value += expectedExtendedValue(one)
	}
	return value.Mul(e.Quantity)
}

// expectedExtendedWeight works out the extended weight of equipment presenting no choice the way it has always been
// worked out, independently of the ranges it now comes from.
func expectedExtendedWeight(e *Equipment) fxp.Weight {
	if e.Quantity <= 0 {
		return 0
	}
	var contents fxp.Int
	for _, one := range e.Children {
		contents += fxp.Int(expectedExtendedWeight(one))
	}
	if len(e.Children) != 0 {
		reduction := containedWeightReductionFor(e, fxp.Pound, e.Modifiers, e.Features)
		contents = fxp.Int(reduction.apply(fxp.Weight(contents)))
	}
	return fxp.Weight((fxp.Int(e.AdjustedWeight(false, fxp.Pound)) + contents).Mul(e.Quantity))
}

// TestEquipmentTotalsAgreeWithTheirRanges verifies, for equipment presenting no choice, that the extended value and
// weight worked out from the ranges are exactly what they always were, that the ranges are settled on them, and that
// the editor's own working of the extended weight agrees, across modifiers, levels, quantities, nesting, groups and
// contained weight reductions, switched on and off.
func TestEquipmentTotalsAgreeWithTheirRanges(t *testing.T) {
	c := check.New(t)
	perLevel := func(parent *Equipment) *Equipment {
		one := newEquipmentItem("Sword", "100", "3 lb")
		one.Level = fxp.FromInteger(2)
		one.Quantity = fxp.FromInteger(3)
		mod := NewEquipmentModifier(nil, nil, false)
		mod.CostType = emcost.Original
		mod.CostAmount = "+10"
		mod.CostIsPerLevel = true
		mod.WeightType = emweight.Original
		mod.WeightAmount = "+1 lb"
		mod.WeightIsPerLevel = true
		one.SetModifiers([]*EquipmentModifier{mod})
		if parent != nil {
			one.SetParent(parent)
			parent.Children = append(parent.Children, one)
		}
		return one
	}
	reducing := func(reduction string, switchable, switchedOn bool) *Equipment {
		bag := NewEquipment(nil, nil, true)
		bag.BaseValue = "25"
		bag.BaseWeight = "2 lb"
		bag.Quantity = fxp.FromInteger(2)
		bag.SwitchedOn = switchedOn
		cwr := NewContainedWeightReduction()
		cwr.Reduction = reduction
		cwr.Switchable = switchable
		bag.Features = Features{cwr}
		perLevel(bag)
		item(bag, "Rope", "5", "2 lb")
		none := item(bag, "Torch", "3", "1 lb")
		none.Quantity = 0
		inner := NewEquipment(nil, bag, true)
		inner.BaseWeight = "0.5 lb"
		item(inner, "Coin", "1", "0.02 lb").Quantity = fxp.FromInteger(40)
		bag.Children = append(bag.Children, inner)
		return bag
	}
	group := NewEquipmentGroup(nil, nil)
	item(group, "Knife", "40", "1 lb")
	bagInGroup := reducing("25%", false, false)
	bagInGroup.SetParent(group)
	group.Children = append(group.Children, bagInGroup)
	fractional := newEquipmentItem("Flour", "2", "1 lb")
	fractional.Quantity = fxp.FromStringForced("2.5")

	for _, one := range []struct {
		name string
		eqp  *Equipment
	}{
		{"per-level modifiers", perLevel(nil)},
		{"percentage reduction", reducing("50%", false, false)},
		{"fixed reduction", reducing("2 lb", false, false)},
		{"full reduction", reducing("100%", false, false)},
		{"switchable reduction, switched off", reducing("50%", true, false)},
		{"switchable reduction, switched on", reducing("50%", true, true)},
		{"group", group},
		{"fractional quantity", fractional},
	} {
		value, settled := one.eqp.ExtendedValueRange().Settled()
		c.True(settled, "%s: the value range must be settled", one.name)
		c.Equal(expectedExtendedValue(one.eqp), value, "%s: the value range", one.name)
		c.Equal(expectedExtendedValue(one.eqp), one.eqp.ExtendedValue(), "%s: the extended value", one.name)
		weight, settled := one.eqp.ExtendedWeightRange(fxp.Pound).Settled()
		c.True(settled, "%s: the weight range must be settled", one.name)
		c.Equal(fxp.Int(expectedExtendedWeight(one.eqp)), weight, "%s: the weight range", one.name)
		c.Equal(expectedExtendedWeight(one.eqp), one.eqp.ExtendedWeight(false, fxp.Pound), "%s: the extended weight",
			one.name)
		c.Equal(one.eqp.ExtendedWeight(false, fxp.Pound), ExtendedWeightAdjustedForModifiers(one.eqp, fxp.Pound,
			one.eqp.Quantity, one.eqp.ResolvedBaseWeight(), one.eqp.Modifiers, one.eqp.Features, one.eqp.Children,
			false, false), "%s: the editor's working of the extended weight", one.name)
	}

	// The weight that counts toward the encumbrance for skills leaves out equipped equipment marked "Ignore weight for
	// skills", which the ranges don't, so it is worked out on its own.
	pack := NewEquipment(nil, nil, true)
	pack.BaseWeight = "3 lb"
	armor := item(pack, "Armor", "100", "20 lb")
	armor.WeightIgnoredForSkills = true
	item(pack, "Rope", "5", "2 lb")
	c.Equal(fxp.Weight(fxp.FromInteger(25)), pack.ExtendedWeight(false, fxp.Pound))
	c.Equal(fxp.Weight(fxp.FromInteger(5)), pack.ExtendedWeight(true, fxp.Pound),
		"the armor's weight must not count for skills")
}

// TestRangesWhoseEndsRenderAlike verifies that a range whose ends render the same, as they can when a display format
// rounds them, is shown as that one rendering rather than as a range from it to itself.
func TestRangesWhoseEndsRenderAlike(t *testing.T) {
	c := check.New(t)
	r := newNumericRange(fxp.FromStringForced("1.2"), fxp.FromStringForced("1.4"))
	whole := func(value fxp.Int) string { return value.Round().String() }
	c.Equal("1", FormatValueRange(r, whole))
	c.Equal("1 lb", FormatWeightRange(r, func(w fxp.Weight) string { return whole(fxp.Int(w)) + " lb" }))
	c.Equal("1.2~1.4 lb", FormatWeightRange(r, fxp.Pound.Format), "ends that render differently remain a range")
}

// TestDisabledGroupIsLeftOutOfAnEquipmentChoice verifies that the contents of a disabled group picked from separately
// are left out of the choice's range, just as a disabled option sitting directly in the choice is.
func TestDisabledGroupIsLeftOutOfAnEquipmentChoice(t *testing.T) {
	c := check.New(t)
	choice := NewEquipmentChoiceContainer(nil, nil)
	choice.TemplatePicker.Type = picker.Count
	choice.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
	choice.TemplatePicker.Qualifier.Qualifier = fxp.One
	group := NewEquipmentGroup(nil, choice)
	group.PickSeparately = true
	item(group, "Rope", "10", "1 lb")
	choice.Children = append(choice.Children, group)
	item(choice, "Knife", "20", "2 lb")
	c.Equal("10~20", choice.ExtendedValueRange().String())
	c.Equal("1~2 lb", FormatWeightRange(choice.ExtendedWeightRange(fxp.Pound), fxp.Pound.Format))

	group.SetDisabled(true)
	c.Equal("20", choice.ExtendedValueRange().String(), "the contents of a disabled group are not options")
	c.Equal("2 lb", FormatWeightRange(choice.ExtendedWeightRange(fxp.Pound), fxp.Pound.Format))
}

// TestEquipmentChoiceRanges verifies the value and weight a choice may come to under each kind of picker, and that a
// container holding one adds its own value and quantity to that.
func TestEquipmentChoiceRanges(t *testing.T) {
	c := check.New(t)
	lb := fxp.FromInteger[int64]

	byCount := newRangeTestChoice(picker.Count, criteria.EqualsNumber, fxp.One)
	c.Equal("10~30", byCount.ExtendedValueRange().String(), "one option by count costs the cheapest to the dearest")
	c.Equal(newNumericRange(lb(1), lb(3)), byCount.ExtendedWeightRange(fxp.Pound))

	byValue := newRangeTestChoice(picker.Value, criteria.AtMostNumber, fxp.FromInteger(50))
	c.Equal("0~50", byValue.ExtendedValueRange().String(), "a value choice is bounded by its qualifier")
	c.Equal("0+", byValue.ExtendedWeightRange(fxp.Pound).String(),
		"quantities may be raised while picking, so the weight has no upper limit")

	byWeight := newRangeTestChoice(picker.Weight, criteria.AtLeastNumber, lb(10))
	c.Equal("10+", byWeight.ExtendedWeightRange(fxp.Pound).String(), "a weight choice is bounded by its qualifier")
	c.Equal("0+", byWeight.ExtendedValueRange().String())

	backpack := NewEquipment(nil, nil, true)
	backpack.BaseValue = "5"
	backpack.Quantity = fxp.FromInteger(2)
	byCount.SetParent(backpack)
	backpack.Children = []*Equipment{byCount}
	c.Equal("30~70", backpack.ExtendedValueRange().String(), "the container's own value and quantity apply")

	tmpl := NewTemplate()
	tmpl.Equipment = []*Equipment{backpack, newEquipmentItem("Knife", "40", "1 lb")}
	title, _ := equipmentTotalsTitle("Equipment", tmpl.Equipment, SheetSettingsFor(nil))
	c.True(strings.Contains(title, "$70~110"), "the list's totals must be ranges too: %s", title)
}

func TestEquipmentRangeSorting(t *testing.T) {
	c := check.New(t)
	c.True(ValueRangeLessFromString("10", "10~30"))
	c.True(ValueRangeLessFromString("10~30", "20"))
	c.True(ValueRangeLessFromString("1,000", "2,000~3,000"))
	less := WeightRangeLessFromStringFunc(fxp.Pound)
	c.True(less("1 lb", "1~3 lb"))
	c.True(less("1~3 lb", "2 lb"))
	c.True(less("1 lb", "1 lb~3 lb"), "a range with the units on both ends must still be read")
	c.True(less("500 g~2 kg", "1 kg"), "ends in different units must each be read in their own")
	c.False(less("0 lb+", "≤5 lb"), "a range with no lower limit sorts first")
}

// TestFormatWeightRange verifies that a weight range gives its units once when both ends share them, and on each end
// otherwise.
func TestFormatWeightRange(t *testing.T) {
	c := check.New(t)
	lb := fxp.FromInteger[int64]
	c.Equal("3~5 lb", FormatWeightRange(newNumericRange(lb(3), lb(5)), fxp.Pound.Format))
	c.Equal("4 lb", FormatWeightRange(NumericRangeOf(lb(4)), fxp.Pound.Format), "a settled range is a single weight")
	c.Equal("0 lb+", FormatWeightRange(numericRangeAtLeast(0), fxp.Pound.Format))
	c.Equal("≤10 lb", FormatWeightRange(numericRangeAtMost(lb(10)), fxp.Pound.Format))
	mixed := func(w fxp.Weight) string {
		if w < fxp.Weight(lb(1)) {
			return fxp.Ounce.Format(w)
		}
		return fxp.Pound.Format(w)
	}
	c.Equal("8 oz~2 lb", FormatWeightRange(newNumericRange(fxp.Half, lb(2)), mixed),
		"ends in different units must each keep their own")
}

// TestEquipmentChoiceCountsAsItsLowerEnd verifies that a choice yet to be made counts as the least it may come to
// wherever a single value or weight is needed, including what is written into a template's file, and so does anything
// holding one.
func TestEquipmentChoiceCountsAsItsLowerEnd(t *testing.T) {
	c := check.New(t)
	choice := newRangeTestChoice(picker.Count, criteria.EqualsNumber, fxp.One)
	c.Equal(fxp.FromInteger(10), choice.ExtendedValue(), "the cheapest option is the least the choice may come to")
	c.Equal(fxp.Weight(fxp.One), choice.ExtendedWeight(false, fxp.Pound))

	backpack := NewEquipment(nil, nil, true)
	backpack.BaseValue = "5"
	backpack.BaseWeight = "2 lb"
	choice.SetParent(backpack)
	backpack.Children = []*Equipment{choice}
	c.Equal(fxp.FromInteger(15), backpack.ExtendedValue())
	c.Equal(fxp.FromInteger(15), backpack.ExtendedValueOfJustOne())
	c.Equal(fxp.Weight(fxp.FromInteger(3)), backpack.ExtendedWeight(false, fxp.Pound))

	tmpl := NewTemplate()
	tmpl.Equipment = []*Equipment{backpack}
	data, err := json.Marshal(tmpl)
	c.NoError(err)
	c.True(strings.Contains(string(data), `"extended_value":15,`),
		"the file must record the least the backpack may come to: %s", string(data))
}

// TestEquipmentRangeThroughAFullReduction verifies that a container whose contents weigh nothing, however much they
// weigh, settles at its own weight even when a choice inside it has no upper limit to its weight.
func TestEquipmentRangeThroughAFullReduction(t *testing.T) {
	c := check.New(t)
	bag := NewEquipment(nil, nil, true)
	bag.BaseWeight = "1 lb"
	cwr := NewContainedWeightReduction()
	cwr.Reduction = "100%"
	bag.Features = Features{cwr}
	choice := newRangeTestChoice(picker.Value, criteria.AtMostNumber, fxp.FromInteger(50))
	choice.SetParent(bag)
	bag.Children = []*Equipment{choice}
	c.Equal("0+", choice.ExtendedWeightRange(fxp.Pound).String(), "precondition: the choice's weight has no upper limit")
	c.Equal("1 lb", FormatWeightRange(bag.ExtendedWeightRange(fxp.Pound), fxp.Pound.Format))
}

// TestEquipmentChoiceOtherMeasureWithGroups verifies that a choice picked by one measure has an upper limit to the other
// when every option with something to raise is a group, whose quantity can't be raised while picking, and none when an
// option's quantity may be raised, even one starting out with a quantity of nothing.
func TestEquipmentChoiceOtherMeasureWithGroups(t *testing.T) {
	c := check.New(t)
	choice := NewEquipmentChoiceContainer(nil, nil)
	choice.TemplatePicker.Type = picker.Value
	choice.TemplatePicker.Qualifier.Compare = criteria.AtMostNumber
	choice.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(100)
	for _, weight := range []string{"2 lb", "3 lb"} {
		group := NewEquipmentGroup(nil, choice)
		item(group, "Gear", "10", weight)
		choice.Children = append(choice.Children, group)
	}
	free := newEquipmentItem("Advice", "0", "0 lb")
	free.SetParent(choice)
	choice.Children = append(choice.Children, free)
	c.Equal("0~5 lb", FormatWeightRange(choice.ExtendedWeightRange(fxp.Pound), fxp.Pound.Format),
		"groups can only be taken or left, and an option weighing nothing adds nothing however many are taken")

	rope := newEquipmentItem("Rope", "5", "2 lb")
	rope.SetParent(choice)
	choice.Children = append(choice.Children, rope)
	c.Equal("0 lb+", FormatWeightRange(choice.ExtendedWeightRange(fxp.Pound), fxp.Pound.Format),
		"an option whose quantity may be raised leaves no upper limit")

	atLeast := NewEquipmentChoiceContainer(nil, nil)
	atLeast.TemplatePicker.Type = picker.Value
	atLeast.TemplatePicker.Qualifier.Compare = criteria.AtLeastNumber
	atLeast.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(50)
	item(atLeast, "Rope", "5", "2 lb").Quantity = 0
	c.Equal("50+", atLeast.ExtendedValueRange().String(), "any pick must come to at least the qualifier")
	c.Equal("0 lb+", FormatWeightRange(atLeast.ExtendedWeightRange(fxp.Pound), fxp.Pound.Format),
		"an option with a quantity of nothing may still be raised, so there is no upper limit to the weight")
}

// TestEquipmentChoiceCappedAtWhatItsOptionsReach verifies that a choice made by value, none of whose options can be
// raised while picking, is held to what its options can come to together, while a qualifier they can't reach at all
// still gives the range it states.
func TestEquipmentChoiceCappedAtWhatItsOptionsReach(t *testing.T) {
	c := check.New(t)
	rangeFor := func(compare criteria.NumericComparison, qualifier int64) string {
		choice := NewEquipmentChoiceContainer(nil, nil)
		choice.TemplatePicker.Type = picker.Value
		choice.TemplatePicker.Qualifier.Compare = compare
		choice.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(qualifier)
		for _, value := range []string{"20", "30"} {
			group := NewEquipmentGroup(nil, choice)
			item(group, "Gear", value, "1 lb")
			choice.Children = append(choice.Children, group)
		}
		return choice.ExtendedValueRange().String()
	}
	c.Equal("0~50", rangeFor(criteria.AnyNumber, 0))
	c.Equal("0~50", rangeFor(criteria.AtMostNumber, 100))
	c.Equal("10~50", rangeFor(criteria.AtLeastNumber, 10))
	c.Equal("100+", rangeFor(criteria.AtLeastNumber, 100), "a qualifier nothing can reach still gives its own range")
}
