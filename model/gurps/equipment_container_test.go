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
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/eqcontainer"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

func newEquipmentItem(name, value, weight string) *Equipment {
	e := NewEquipment(nil, nil, false)
	e.Name = name
	e.BaseValue = value
	e.BaseWeight = weight
	return e
}

// newEquipmentChoice returns an equipment choice container holding an option for each name.
func newEquipmentChoice(names ...string) *Equipment {
	choice := NewEquipmentChoiceContainer(nil, nil)
	for _, name := range names {
		option := newEquipmentItem(name, "10", "1 lb")
		option.SetParent(choice)
		choice.Children = append(choice.Children, option)
	}
	return choice
}

// TestEquipmentContainerTypeDefaultsToPhysical verifies that a container written without a container type, as every
// container was before groups existed, loads as a physical container keeping everything it had, and that only a group
// writes its type out.
func TestEquipmentContainerTypeDefaultsToPhysical(t *testing.T) {
	c := check.New(t)
	var legacy Equipment
	c.NoError(json.Unmarshal([]byte(`{"id":"`+string(NewEquipment(nil, nil, true).TID)+
		`","description":"Backpack","quantity":2,"base_value":"60","base_weight":"3 lb","tech_level":"3"}`),
		&legacy))
	c.True(legacy.IsPhysicalContainer(), "a container with no type must be a physical container")
	c.Equal(fxp.FromInteger(2), legacy.Quantity)
	c.Equal("60", legacy.BaseValue)
	c.Equal("3 lb", legacy.BaseWeight)
	c.Equal("3", legacy.TechLevel)

	data, err := json.Marshal(&legacy)
	c.NoError(err)
	c.False(strings.Contains(string(data), `"container_type"`), "a physical container must not write its type")

	group := NewEquipmentGroup(nil, nil)
	data, err = json.Marshal(group)
	c.NoError(err)
	c.True(strings.Contains(string(data), `"container_type":"group"`), "a group must write its type")
	var loaded Equipment
	c.NoError(json.Unmarshal(data, &loaded))
	c.True(loaded.IsGroup(), "a group must load as one")
}

// TestEquipmentKinds verifies that each kind of equipment is named after what it is.
func TestEquipmentKinds(t *testing.T) {
	c := check.New(t)
	c.Equal("Equipment", NewEquipment(nil, nil, false).Kind())
	c.Equal("Equipment Container", NewEquipment(nil, nil, true).Kind())
	group := NewEquipmentGroup(nil, nil)
	c.Equal("Equipment Group", group.Kind())
	c.Equal("Equipment Group", group.Name)
	choice := NewEquipmentChoiceContainer(nil, nil)
	c.Equal("Equipment Choice", choice.Kind())
	c.Equal("Equipment Choice", choice.Name)
}

// TestEquipmentGroupHasNothingOfItsOwn verifies that a group holds none of the data that would make it a piece of
// equipment in its own right, and so is worth and weighs exactly what its contents do.
func TestEquipmentGroupHasNothingOfItsOwn(t *testing.T) {
	c := check.New(t)
	group := NewEquipmentGroup(nil, nil)
	c.Equal("", group.LegalityClass, "a new group must not start out with a legality class")
	group.Quantity = fxp.FromInteger(3)
	group.BaseValue = "100"
	group.BaseWeight = "5 lb"
	group.TechLevel = "4"
	group.LegalityClass = "2"
	group.MaxUses = 3
	group.Uses = 2
	group.RatedST = fxp.FromInteger(10)
	group.Level = fxp.One
	group.WeightIgnoredForSkills = true
	group.SwitchedOn = true
	group.Preconfigured = true
	group.Prereq = NewPrereqList()
	group.Features = Features{NewContainedWeightReduction()}
	group.Modifiers = []*EquipmentModifier{NewEquipmentModifier(nil, nil, false)}
	group.Weapons = []*Weapon{NewWeapon(group, true)}
	for _, one := range []*Equipment{newEquipmentItem("Rope", "10", "2 lb"), newEquipmentItem("Torch", "3", "1 lb")} {
		one.SetParent(group)
		group.Children = append(group.Children, one)
	}
	group.ClearUnusedFieldsForType()

	c.Equal(fxp.One, group.Quantity, "a group's quantity must always be one")
	c.Equal("", group.BaseValue)
	c.Equal("", group.BaseWeight)
	c.Equal("", group.TechLevel)
	c.Equal("", group.LegalityClass)
	c.Equal(0, group.MaxUses)
	c.Equal(0, group.Uses)
	c.Equal(fxp.Int(0), group.RatedST)
	c.Equal(fxp.Int(0), group.Level)
	c.False(group.WeightIgnoredForSkills)
	c.False(group.SwitchedOn)
	c.False(group.Preconfigured)
	c.True(group.Prereq.IsZero())
	c.Equal(0, len(group.Features))
	c.Equal(0, len(group.Modifiers))
	c.Equal(0, len(group.Weapons))
	c.Equal(2, len(group.Children), "a group must keep its contents")

	c.Equal(fxp.FromInteger(13), group.ExtendedValue(), "a group must be worth what its contents are")
	c.Equal(fxp.Weight(fxp.FromInteger(3)), group.ExtendedWeight(false, fxp.Pound),
		"a group must weigh what its contents do")
	c.False(group.RequiresTL(), "a group has no tech level of its own")
	c.False(group.CanPreconfigureContainer(), "a group has no modifiers of its own to preconfigure")
	c.True(NewEquipment(nil, nil, true).CanPreconfigureContainer(), "a physical container may still be preconfigured")
}

// TestEquipmentGroupConversion verifies that a physical container and a group can be converted into each other, that
// converting to a group reports and removes what a group can't hold, and that a choice container can't skip straight
// to either a physical container or a piece of equipment that isn't a container.
func TestEquipmentGroupConversion(t *testing.T) {
	c := check.New(t)
	backpack := NewEquipment(nil, nil, true)
	backpack.Name = "Backpack"
	c.True(backpack.CanConvertToGroup())
	c.False(backpack.CanConvertToPhysicalContainer(), "a physical container already is one")
	c.Equal(0, len(backpack.GroupConversionLosses()), "a new container has nothing worth a warning to lose")
	backpack.LegalityClass = "2"
	c.Equal([]string{"legality class"}, backpack.GroupConversionLosses(), "a changed legality class is worth one")
	backpack.BaseValue = "60"
	backpack.Modifiers = []*EquipmentModifier{NewEquipmentModifier(nil, nil, false)}
	c.Equal([]string{"value", "legality class", "modifiers"}, backpack.GroupConversionLosses())

	backpack.ConvertToGroup()
	c.True(backpack.IsGroup())
	c.Equal("", backpack.BaseValue)
	c.Equal(0, len(backpack.Modifiers))
	c.False(backpack.CanConvertToGroup(), "a group already is one")

	c.True(backpack.CanConvertToPhysicalContainer())
	backpack.ConvertToPhysicalContainer()
	c.True(backpack.IsPhysicalContainer())
	c.Equal("4", backpack.LegalityClass, "a container must get back the legality class new equipment starts with")

	c.False(NewEquipment(nil, nil, false).CanConvertToGroup(), "only a container may become a group")

	choice := NewEquipmentChoiceContainer(nil, nil)
	c.False(choice.CanConvertToPhysicalContainer(), "a choice container must become a group first")
	c.False(choice.CanConvertToGroup(), "a choice container becomes a group by losing its choices")
	c.False(choice.CanConvertToFromContainer(), "a choice container must not become a plain piece of equipment")
	c.True(NewEquipmentGroup(nil, nil).CanConvertToFromContainer(), "an empty group may become a plain item")
}

// TestEquipmentChoiceConversion verifies that only a group may become a choice container, giving up its VTT notes,
// tags, unequipped state and library source, and that converting back leaves a group.
func TestEquipmentChoiceConversion(t *testing.T) {
	c := check.New(t)
	c.False(CanConvertToTemplateChoiceContainer(NewEquipment(nil, nil, true)),
		"a physical container must become a group first")
	c.False(CanConvertToTemplateChoiceContainer(NewEquipment(nil, nil, false)), "only a container may become one")

	group := NewEquipmentGroup(nil, nil)
	c.True(CanConvertToTemplateChoiceContainer(group))
	c.Equal(0, len(TemplateChoiceConversionLosses(group)), "a bare group loses nothing")
	group.Equipped = false
	group.VTTNotes = "vtt"
	group.Tags = []string{"Gear"}
	group.Source = Source{Library: "lib", Path: "group.eqp", TID: group.ID()}
	c.Equal([]string{"VTT notes", "tags", "unequipped state", "library source"}, TemplateChoiceConversionLosses(group))

	ConvertToTemplateChoiceContainer(group)
	c.True(IsTemplateChoiceContainer(group))
	c.True(group.Equipped, "a choice container must be left equipped")
	c.Equal("", group.VTTNotes)
	c.Equal(0, len(group.Tags))
	c.True(group.Source.IsZero())
	c.Equal("Pick 1", group.TemplatePicker.String())

	ConvertFromTemplateChoiceContainer(group)
	c.False(IsTemplateChoiceContainer(group))
	c.True(group.IsGroup(), "it must be left a group")
}

// TestEquipmentChoiceContainers verifies that an equipment choice container is picked by count, value or weight, and
// takes part in finding and clearing template picker data like any other.
func TestEquipmentChoiceContainers(t *testing.T) {
	c := check.New(t)
	choice := newEquipmentChoice("Sword", "Axe")
	c.True(IsTemplateChoiceContainer(choice))
	c.True(choice.IsGroup(), "a choice container must be a group")
	types, _ := choice.TemplatePickerData()
	c.Equal([]picker.Type{picker.NotApplicable, picker.Count, picker.Value, picker.Weight}, types,
		"equipment may be picked by count, value or weight, but not by points")

	outer := NewEquipmentGroup(nil, nil)
	choice.SetParent(outer)
	outer.Children = []*Equipment{choice}
	c.True(HasTemplatePickerData(outer), "a nested choice must be found")
	ClearTemplatePickerData(outer)
	c.False(HasTemplatePickerData(outer), "the nested choice must be cleared")
	c.True(choice.IsGroup(), "a cleared choice must be left a group")
}

// TestTemplateLoadGroupsEquipmentChoiceContainers verifies that loading a template turns an equipment choice container
// that isn't a group into one, clearing what only a physical container holds.
func TestTemplateLoadGroupsEquipmentChoiceContainers(t *testing.T) {
	c := check.New(t)
	choice := newEquipmentChoice("Sword", "Axe")
	choice.ContainerType = eqcontainer.Container
	choice.BaseValue = "50"
	choice.VTTNotes = "vtt"
	choice.Tags = []string{"Weapon"}
	choice.Source = Source{Library: "lib", Path: "choice.eqp", TID: choice.ID()}
	backpack := NewEquipment(nil, nil, true)
	backpack.BaseValue = "60"
	tmpl := NewTemplate()
	tmpl.Equipment = []*Equipment{choice, backpack}

	data, err := json.Marshal(tmpl)
	c.NoError(err)
	var loaded Template
	c.NoError(json.Unmarshal(data, &loaded))
	c.Equal(2, len(loaded.Equipment))
	c.True(loaded.Equipment[0].IsGroup(), "a choice container must become a group")
	c.Equal("", loaded.Equipment[0].BaseValue, "a choice container must not keep a value of its own")
	c.Equal("", loaded.Equipment[0].VTTNotes, "a choice container must not keep VTT notes")
	c.Equal(0, len(loaded.Equipment[0].Tags), "a choice container must not keep tags")
	c.True(loaded.Equipment[0].Source.IsZero(), "a choice container must not keep a source")
	c.True(IsTemplateChoiceContainer(loaded.Equipment[0]), "the choices themselves must survive")
	c.True(loaded.Equipment[1].IsPhysicalContainer(), "a container without choices keeps its type")
	c.Equal("60", loaded.Equipment[1].BaseValue)
}

// newRangeTestChoice returns an equipment choice picked as given from two options, one worth $10 and weighing 1 lb, the
// other worth $30 and weighing 3 lb.
func newRangeTestChoice(pickerType picker.Type, compare criteria.NumericComparison, qualifier fxp.Int) *Equipment {
	choice := NewEquipmentChoiceContainer(nil, nil)
	choice.TemplatePicker.Type = pickerType
	choice.TemplatePicker.Qualifier.Compare = compare
	choice.TemplatePicker.Qualifier.Qualifier = qualifier
	for _, one := range []*Equipment{newEquipmentItem("Cheap", "10", "1 lb"), newEquipmentItem("Dear", "30", "3 lb")} {
		one.SetParent(choice)
		choice.Children = append(choice.Children, one)
	}
	return choice
}

// TestEquipmentChoiceDescriptions verifies how a choice made by value or weight describes itself.
func TestEquipmentChoiceDescriptions(t *testing.T) {
	c := check.New(t)
	c.Equal("Pick at most $1,500 worth",
		newRangeTestChoice(picker.Value, criteria.AtMostNumber, fxp.FromInteger(1500)).TemplatePicker.String())
	units := SheetSettingsFor(nil).DefaultWeightUnits
	c.Equal("Pick at least "+units.Format(fxp.Weight(fxp.FromInteger(10)))+" in weight",
		newRangeTestChoice(picker.Weight, criteria.AtLeastNumber, fxp.FromInteger(10)).TemplatePicker.String())
}

// TestEquipmentRangesMatchSettledTotals verifies that equipment presenting no choice has a settled value and weight
// range that is exactly its extended value and weight, contained weight reductions and quantity included.
func TestEquipmentRangesMatchSettledTotals(t *testing.T) {
	c := check.New(t)
	bag := NewEquipment(nil, nil, true)
	bag.BaseValue = "5"
	bag.BaseWeight = "2 lb"
	bag.Quantity = fxp.FromInteger(3)
	cwr := NewContainedWeightReduction()
	cwr.Reduction = "25%"
	bag.Features = Features{cwr}
	for _, one := range []*Equipment{newEquipmentItem("Rope", "10", "3 lb"), newEquipmentItem("Torch", "3", "1 lb")} {
		one.SetParent(bag)
		bag.Children = append(bag.Children, one)
	}
	value, settled := bag.ExtendedValueRange().Settled()
	c.True(settled)
	c.Equal(bag.ExtendedValue(), value)
	weight, settled := bag.ExtendedWeightRange(fxp.Pound).Settled()
	c.True(settled)
	c.Equal(fxp.Int(bag.ExtendedWeight(false, fxp.Pound)), weight)
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

// TestEquipmentSyncToGroupClearsWhatAGroupCantHold verifies that a physical container synced to a library entry that
// has since become a group drops what a group can't hold, rather than keeping it hidden until the next save.
func TestEquipmentSyncToGroupClearsWhatAGroupCantHold(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	libFile := LibraryFile{Library: "Test Library", Path: "Test" + EquipmentExt}
	source := NewEquipmentGroup(nil, nil)
	source.Name = "Kit"
	stubLibrarySources(t, e.SourceMatcher(), libFile, source)
	local := NewEquipment(e, nil, true)
	local.Name = "Kit"
	local.Quantity = fxp.FromInteger(2)
	local.RatedST = fxp.FromInteger(10)
	local.Modifiers = []*EquipmentModifier{NewEquipmentModifier(e, nil, false)}
	local.Source = Source{LibraryFile: libFile, TID: source.TID}
	state, _ := e.SourceMatcher().Match(local)
	c.Equal(srcstate.Mismatched, state, "precondition: the local copy differs from its source")

	local.SyncWithSource()
	c.True(local.IsGroup(), "the container must have become a group, as its source is")
	c.Equal(fxp.One, local.Quantity, "a group's quantity must be one")
	c.Equal(fxp.Int(0), local.RatedST)
	c.Equal(0, len(local.Modifiers))
}

// TestEquipmentChoiceDescribesWeightInGivenUnits verifies that a choice made by weight can describe itself in the units
// of the sheet it is being picked for.
func TestEquipmentChoiceDescribesWeightInGivenUnits(t *testing.T) {
	c := check.New(t)
	tp := newRangeTestChoice(picker.Weight, criteria.AtMostNumber, fxp.FromInteger(10)).TemplatePicker
	c.Equal("Pick at most "+fxp.Kilogram.Format(fxp.Weight(fxp.FromInteger(10)))+" in weight",
		tp.StringWithUnits(fxp.Kilogram))
	c.Equal(tp.StringWithUnits(SheetSettingsFor(nil).DefaultWeightUnits), tp.String())
}

// TestEquipmentGroupToItemGetsDefaultLegalityClass verifies that a group converted to a piece of equipment gets the
// legality class new equipment starts out with, while a physical container keeps whatever it had.
func TestEquipmentGroupToItemGetsDefaultLegalityClass(t *testing.T) {
	c := check.New(t)
	group := NewEquipmentGroup(nil, nil)
	group.ConvertToNonContainer()
	c.False(group.Container())
	c.Equal("4", group.LegalityClass)

	backpack := NewEquipment(nil, nil, true)
	backpack.LegalityClass = ""
	backpack.ConvertToNonContainer()
	c.Equal("", backpack.LegalityClass, "a legality class the user cleared must be left alone")
}

// TestLoadingOutsideATemplateClearsEquipmentPickerData verifies that a character sheet, a loot sheet and an equipment
// list all have any equipment template picker data removed when loaded, leaving the options alone.
func TestLoadingOutsideATemplateClearsEquipmentPickerData(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	entity.CarriedEquipment = []*Equipment{newEquipmentChoice("Sword", "Axe")}
	entity.OtherEquipment = []*Equipment{newEquipmentChoice("Rope", "Chain")}
	data, err := json.Marshal(entity)
	c.NoError(err)
	var loadedEntity Entity
	c.NoError(json.Unmarshal(data, &loadedEntity))
	c.False(HasTemplatePickerData(loadedEntity.CarriedEquipment...), "carried equipment must not keep picker data")
	c.False(HasTemplatePickerData(loadedEntity.OtherEquipment...), "other equipment must not keep picker data")
	c.Equal(2, len(loadedEntity.CarriedEquipment[0].Children), "the options must be left alone")

	loot := NewLoot()
	loot.Equipment = []*Equipment{newEquipmentChoice("Sword", "Axe")}
	data, err = json.Marshal(loot)
	c.NoError(err)
	var loadedLoot Loot
	c.NoError(json.Unmarshal(data, &loadedLoot))
	c.False(HasTemplatePickerData(loadedLoot.Equipment...), "a loot sheet must not keep picker data")

	list, err := json.Marshal(&listData[*Equipment]{
		Version: jio.CurrentDataVersion,
		Rows:    []*Equipment{newEquipmentChoice("Sword", "Axe")},
	})
	c.NoError(err)
	rows, err := NewEquipmentFromFile(fstest.MapFS{"list.eqp": &fstest.MapFile{Data: list}}, "list.eqp")
	c.NoError(err)
	c.False(HasTemplatePickerData(rows...), "an equipment list must not keep picker data")
}

// TestCanTakeModifiers verifies which rows modifiers may be attached to: anything but a template choice container or
// an equipment group.
func TestCanTakeModifiers(t *testing.T) {
	c := check.New(t)
	c.True(CanTakeModifiers(NewEquipment(nil, nil, false)))
	c.True(CanTakeModifiers(NewEquipment(nil, nil, true)), "a physical container keeps modifiers of its own")
	c.False(CanTakeModifiers(NewEquipmentGroup(nil, nil)), "a group keeps nothing of its own, modifiers included")
	c.False(CanTakeModifiers(NewEquipmentChoiceContainer(nil, nil)))
	c.True(CanTakeModifiers(NewTrait(nil, nil, true)))
	c.False(CanTakeModifiers(NewTraitChoiceContainer(nil, nil)))
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

// TestEquipmentGroupRoundTripsThroughAnItem verifies that a group converted to a plain item and back comes back as a
// physical container, since the item it became is a piece of equipment in its own right, and that the item keeps no
// trace of having been a group.
func TestEquipmentGroupRoundTripsThroughAnItem(t *testing.T) {
	c := check.New(t)
	group := NewEquipmentGroup(nil, nil)
	group.ConvertToNonContainer()
	c.Equal(eqcontainer.Container, group.ContainerType, "the item must not keep the group's container type")
	group.BaseValue = "10"
	group.ConvertToContainer()
	c.True(group.IsPhysicalContainer(), "an item must become a physical container")
	c.Equal("10", group.BaseValue, "the value the item was given must be kept")
	c.Equal(fxp.FromInteger(10), group.ExtendedValue())

	state := group.ContainerConversionState()
	group.ConvertToGroup()
	group.RestoreContainerConversionState(state)
	c.True(group.IsPhysicalContainer(), "restoring the state must put the container type back")
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
// when every option with something to raise is a group, whose quantity can't be raised while picking.
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
}

// item adds a piece of equipment with the given name, value and weight to the parent.
func item(parent *Equipment, name, value, weight string) *Equipment {
	one := newEquipmentItem(name, value, weight)
	one.SetParent(parent)
	parent.Children = append(parent.Children, one)
	return one
}

// TestEquipmentChoiceOptionWithNoQuantity verifies that an option starting out with a quantity of nothing still counts
// as able to add to a choice made by value or weight, since its quantity may be raised while picking.
func TestEquipmentChoiceOptionWithNoQuantity(t *testing.T) {
	c := check.New(t)
	choice := NewEquipmentChoiceContainer(nil, nil)
	choice.TemplatePicker.Type = picker.Value
	choice.TemplatePicker.Qualifier.Compare = criteria.AtLeastNumber
	choice.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(50)
	rope := item(choice, "Rope", "5", "2 lb")
	rope.Quantity = 0
	c.Equal("50+", choice.ExtendedValueRange().String(), "any pick must come to at least the qualifier")
	c.Equal("0 lb+", FormatWeightRange(choice.ExtendedWeightRange(fxp.Pound), fxp.Pound.Format),
		"the rope's quantity may be raised, so there is no upper limit to the weight")
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

// TestEditorDataCantChangeTheKindOfContainer verifies that data an editor took from a container before the container
// was converted to another kind can't put the old kind back when applied, since the kind isn't among the data an editor
// edits.
func TestEditorDataCantChangeTheKindOfContainer(t *testing.T) {
	c := check.New(t)
	backpack := NewEquipment(nil, nil, true)
	var data EquipmentEditData
	data.CopyFrom(backpack)
	backpack.ConvertToGroup()
	data.ApplyTo(backpack)
	c.True(backpack.IsGroup(), "applying the stale data must leave the group a group")

	clone := backpack.Clone(LibraryFile{}, nil, nil, Copy)
	c.True(clone.IsGroup(), "a clone must keep the kind of container")
}
