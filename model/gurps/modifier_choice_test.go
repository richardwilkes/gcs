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
	"hash"
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// newTraitModifierOption returns a trait modifier with the given name and cost adjustment, held by parent.
func newTraitModifierOption(parent *TraitModifier, name, cost string) *TraitModifier {
	m := NewTraitModifier(nil, parent, false)
	m.Name = name
	m.CostAdj = cost
	parent.Children = append(parent.Children, m)
	return m
}

// newTraitModifierChoiceWith returns a trait modifier choice holding an option for each cost adjustment, mandatory or
// not as asked, with none of them picked.
func newTraitModifierChoiceWith(mandatory bool, costs ...string) *TraitModifier {
	choice := NewTraitModifierChoice(nil, nil)
	choice.SetMandatoryChoice(mandatory)
	for _, cost := range costs {
		newTraitModifierOption(choice, "Option "+cost, cost).SetEnabled(false)
	}
	return choice
}

// TestModifierContainerKinds verifies that a new modifier container is a group, that a new modifier choice is a
// mandatory choice, and that each is named for its kind.
func TestModifierContainerKinds(t *testing.T) {
	c := check.New(t)
	group := NewTraitModifier(nil, nil, true)
	c.False(IsModifierChoice(group))
	c.Equal("Trait Modifier Group", group.Kind())
	c.Equal("Trait Modifier Group", group.Name)
	choice := NewTraitModifierChoice(nil, nil)
	c.True(IsModifierChoice(choice))
	c.True(IsMandatoryModifierChoice(choice))
	c.Equal("Trait Modifier Choice", choice.Kind())
	c.Equal("Trait Modifier Choice", choice.Name)
	c.Equal("Pick 1", ModifierChoiceDescription(choice))
	choice.SetMandatoryChoice(false)
	c.False(IsMandatoryModifierChoice(choice))
	c.Equal("Pick at most 1", ModifierChoiceDescription(choice))

	eqGroup := NewEquipmentModifier(nil, nil, true)
	c.Equal("Equipment Modifier Group", eqGroup.Kind())
	eqChoice := NewEquipmentModifierChoice(nil, nil)
	c.True(IsMandatoryModifierChoice(eqChoice))
	c.Equal("Equipment Modifier Choice", eqChoice.Kind())

	c.False(IsModifierChoice(NewTraitModifier(nil, nil, false)), "a modifier that isn't a container is never a choice")
	c.False(IsTemplateChoiceContainer(choice), "a modifier choice is not a template choice")
}

// TestModifierChoiceReadsUnsupportedFormsAsTheNearest verifies that a choice held in a form this version doesn't
// support is read, and described, as the nearer of the two it does, an exact count as exactly one and any other as at
// most one, yet kept as it is through a save, so that a newer version's choice isn't lost, until the choice is edited.
// Neither a modifier that isn't a container nor a group keeps picker data, nor a container its VTT notes.
func TestModifierChoiceReadsUnsupportedFormsAsTheNearest(t *testing.T) {
	c := check.New(t)
	load := func(container bool, tp TemplatePicker) *TraitModifier {
		t.Helper()
		pickerData, err := json.Marshal(tp)
		c.NoError(err)
		var m TraitModifier
		c.NoError(json.Unmarshal([]byte(`{"id":"`+string(NewTraitModifier(nil, nil, container).TID)+
			`","name":"Choice","vtt_notes":"notes","choice":`+string(pickerData)+`}`), &m))
		return &m
	}
	number := func(compare criteria.NumericComparison, qualifier int) criteria.Number {
		return criteria.Number{Compare: compare, Qualifier: fxp.FromInteger(qualifier)}
	}

	pickThree := TemplatePicker{Type: picker.Count, Qualifier: number(criteria.EqualsNumber, 3)}
	exact := load(true, pickThree)
	c.True(IsMandatoryModifierChoice(exact), "an exact count is read as exactly one")
	c.Equal("Pick 1", ModifierChoiceDescription(exact), "and described as the form it is read as")
	c.Equal(pickThree, exact.Choice, "the form it was held in is kept")
	c.Equal("", exact.VTTNotes, "a container keeps no VTT notes")

	byPoints := load(true, TemplatePicker{Type: picker.Points, Qualifier: number(criteria.AtLeastNumber, 5)})
	c.True(IsModifierChoice(byPoints))
	c.False(IsMandatoryModifierChoice(byPoints), "any other count is read as at most one")
	c.Equal("Pick at most 1", ModifierChoiceDescription(byPoints))

	group := load(true, TemplatePicker{Qualifier: number(criteria.AtMostNumber, 2)})
	c.False(IsModifierChoice(group))
	c.Equal(TemplatePicker{}, group.Choice, "a group keeps no stray picker data")

	leaf := load(false, newModifierChoicePicker(true))
	c.Equal(TemplatePicker{}, leaf.Choice, "a modifier that isn't a container keeps no picker data")
	c.Equal("notes", leaf.VTTNotes)

	data, err := json.Marshal(exact)
	c.NoError(err)
	var reloaded TraitModifier
	c.NoError(json.Unmarshal(data, &reloaded))
	c.Equal(pickThree, reloaded.Choice, "a form this version doesn't support survives a save")

	exact.SetMandatoryChoice(true)
	c.Equal(newModifierChoicePicker(true), exact.Choice, "editing the choice rewrites it in a supported form")
}

// TestModifierLibraryKeepsItsChoices verifies that a modifier library keeps its choices when saved and loaded again,
// unlike template picker data, which only a template may hold and which loading any other list removes.
func TestModifierLibraryKeepsItsChoices(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	c.NoError(SaveTraitModifiers([]*TraitModifier{newTraitModifierChoiceWith(false, "+5")},
		filepath.Join(dir, "mods"+TraitModifiersExt)))
	loaded, err := NewTraitModifiersFromFile(os.DirFS(dir), "mods"+TraitModifiersExt)
	c.NoError(err)
	c.Equal(1, len(loaded))
	c.True(IsModifierChoice(loaded[0]))
	c.False(IsMandatoryModifierChoice(loaded[0]))
	c.Equal(1, len(ModifierChoiceOptions(loaded[0])))
}

// TestModifierChoiceOptions verifies that the options of a choice are the modifiers beneath it, through any group, but
// not those of a choice within it, which has its own.
func TestModifierChoiceOptions(t *testing.T) {
	c := check.New(t)
	choice := newTraitModifierChoiceWith(true, "+5")
	group := NewTraitModifier(nil, choice, true)
	choice.Children = append(choice.Children, group)
	inGroup := newTraitModifierOption(group, "In Group", "+1")
	inner := NewTraitModifierChoice(nil, choice)
	choice.Children = append(choice.Children, inner)
	innerOption := newTraitModifierOption(inner, "Inner", "+2")

	c.Equal([]*TraitModifier{choice.Children[0], inGroup}, ModifierChoiceOptions(choice))
	c.Equal([]*TraitModifier{innerOption}, ModifierChoiceOptions(inner))
	owner, ok := ModifierChoiceFor(inGroup)
	c.True(ok)
	c.Equal(choice, owner, "a group within a choice only organizes its options")
	owner, ok = ModifierChoiceFor(innerOption)
	c.True(ok)
	c.Equal(inner, owner, "a choice within a choice has its own options")
	_, ok = ModifierChoiceFor(group)
	c.False(ok, "a container is never an option")
	_, ok = ModifierChoiceFor(newTraitModifierOption(NewTraitModifier(nil, nil, true), "Loose", "+1"))
	c.False(ok, "a modifier in a plain group is no option")
	c.Nil(ModifierChoiceOptions(group))
}

// TestModifierChoiceConversion verifies that a group converts to a mandatory choice and back, and that nothing else
// does.
func TestModifierChoiceConversion(t *testing.T) {
	c := check.New(t)
	group := NewEquipmentModifier(nil, nil, true)
	c.True(CanConvertToModifierChoice(group))
	ConvertToModifierChoice(group)
	c.True(IsMandatoryModifierChoice(group))
	c.False(CanConvertToModifierChoice(group), "a choice is already a choice")
	before := Hash64(group)
	group.SetMandatoryChoice(false)
	c.NotEqual(before, Hash64(group), "what a choice asks for is part of its sync data")
	ConvertFromModifierChoice(group)
	c.False(IsModifierChoice(group))
	c.False(CanConvertToModifierChoice(NewEquipmentModifier(nil, nil, false)))
	c.False(CanConvertToModifierChoice(NewTrait(nil, nil, true)), "only a modifier container can become one")
}

// TestTraitPointsRangeWithMandatoryModifierChoice verifies that a trait whose mandatory modifier choice is still to be
// made reports the range of costs its options give it, and counts as the least of them wherever a single cost is
// needed. On a character sheet every choice has been made. A preconfigured trait takes a choice that already has its
// pick as made, but still has one without a pick to make. An optional choice never opens the cost.
func TestTraitPointsRangeWithMandatoryModifierChoice(t *testing.T) {
	c := check.New(t)
	newTrait := func(owner DataOwner, mandatory bool) *Trait {
		trait := NewTrait(owner, nil, false)
		trait.BasePoints = fxp.FromInteger(10)
		trait.AddModifiers(newTraitModifierChoiceWith(mandatory, "+5", "+10"))
		return trait
	}

	trait := newTrait(nil, true)
	c.Equal("15~20", trait.PointsRange(nil).String(), "each option is costed on its own")
	c.Equal(fxp.FromInteger(15), trait.AdjustedPoints(nil), "a single cost is the least the choice may come to")
	trait.Modifiers[0].Children[1].SetEnabled(true)
	c.Equal("15~20", trait.PointsRange(nil).String(), "a pick made outside a sheet is only a default until asked")

	trait.Preconfigured = true
	c.Equal("20", trait.PointsRange(nil).String(), "a preconfigured trait takes the pick already made")
	c.Equal(fxp.FromInteger(20), trait.AdjustedPoints(nil))
	trait.Modifiers[0].Children[1].SetEnabled(false)
	c.Equal("15~20", trait.PointsRange(nil).String(), "an unresolved choice is still to be made when preconfigured")

	c.Equal("10", newTrait(nil, false).PointsRange(nil).String(), "an optional choice doesn't open the cost")
	c.Equal("10", newTrait(NewEntity(), true).PointsRange(nil).String(), "on a sheet every choice has been made")

	// A choice held by a container applies to the traits inside it too.
	parent := NewTrait(nil, nil, true)
	parent.AddModifiers(newTraitModifierChoiceWith(true, "+1", "+3"))
	child := NewTrait(nil, parent, false)
	child.BasePoints = fxp.FromInteger(10)
	parent.Children = []*Trait{child}
	c.Equal("11~13", child.PointsRange(nil).String())
	c.Equal("11~13", parent.PointsRange(nil).String())
	c.Equal(fxp.FromInteger(11), parent.AdjustedPoints(nil))

	// The container is asked about its own modifiers, so whether their choice counts as made is up to its preconfigured
	// mark, not the child's.
	parent.Modifiers[0].Children[0].SetEnabled(true)
	child.Preconfigured = true
	c.Equal("11~13", child.PointsRange(nil).String(), "a preconfigured child doesn't settle its container's choice")
	parent.Preconfigured = true
	child.Preconfigured = false
	c.Equal("11", child.PointsRange(nil).String(), "a preconfigured container settles its own choice")
}

// TestUnresolvedModifierChoiceOnASheet verifies that a mandatory choice with no pick is flagged on a character sheet,
// both on the item and on the choice itself, and only there, and that the pick of one on a sheet can't be turned off,
// while anything else can.
func TestUnresolvedModifierChoiceOnASheet(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	trait := NewTrait(entity, nil, false)
	choice := newTraitModifierChoiceWith(true, "+5", "+10")
	choice.Name = "Size"
	trait.AddModifiers(choice)
	c.False(ModifierChoiceIsResolved(choice))
	c.Equal([]*TraitModifier{choice}, UnresolvedModifierChoices(trait.Modifiers...))
	var data CellData
	trait.CellData(TraitDescriptionColumn, &data)
	c.Equal("A modifier must be picked for each of these choices: Size", data.UnresolvedChoice)
	data = CellData{}
	choice.CellData(TraitModifierDescriptionColumn, &data)
	c.True(data.ChoiceRequired, "the choice itself is marked as required too")
	c.Equal("One of these modifiers must be picked.", data.Tooltip)

	library := NewTrait(nil, nil, false)
	library.AddModifiers(newTraitModifierChoiceWith(true, "+5"))
	data = CellData{}
	library.CellData(TraitDescriptionColumn, &data)
	c.Equal("", data.UnresolvedChoice, "an unresolved choice is allowed anywhere but on a sheet")
	data = CellData{}
	library.Modifiers[0].CellData(TraitModifierDescriptionColumn, &data)
	c.False(data.ChoiceRequired, "and so isn't marked as required there")

	pick := choice.Children[0]
	pick.SetEnabled(true)
	c.True(ModifierChoiceIsResolved(choice))
	data = CellData{}
	trait.CellData(TraitDescriptionColumn, &data)
	c.Equal("", data.UnresolvedChoice)
	data = CellData{}
	choice.CellData(TraitModifierDescriptionColumn, &data)
	c.False(data.ChoiceRequired, "a choice with its pick is no longer required")
	c.True(IsLockedModifierChoiceSelection(pick), "the pick of a mandatory choice on a sheet can't be turned off")
	c.False(IsLockedModifierChoiceSelection(choice.Children[1]), "an option that isn't picked can be")
	c.False(IsLockedModifierChoiceSelection(library.Modifiers[0].Children[0]), "off a sheet anything can be")
	optional := newTraitModifierChoiceWith(false, "+1")
	trait.AddModifiers(optional)
	optional.Children[0].SetEnabled(true)
	c.False(IsLockedModifierChoiceSelection(optional.Children[0]), "an optional choice may be left without a pick")
	c.True(ModifierChoiceIsResolved(NewTraitModifierChoice(nil, nil)), "a choice with no options has nothing to pick")
}

// TestSettleModifierChoices verifies that settling keeps no more than one option of a choice enabled, keeping the pick
// a choice already had over options that have just arrived, and that loading and converting to a choice settle too.
func TestSettleModifierChoices(t *testing.T) {
	c := check.New(t)
	choice := newTraitModifierChoiceWith(false, "+1", "+2", "+3")
	enabled := func() []bool {
		return []bool{choice.Children[0].Enabled(), choice.Children[1].Enabled(), choice.Children[2].Enabled()}
	}
	for _, one := range choice.Children {
		one.SetEnabled(true)
	}
	arrived := choice.Children[0]
	c.True(SettleModifierChoices(func(m *TraitModifier) bool { return m == arrived }, choice))
	c.Equal([]bool{false, true, false}, enabled(), "the first option that was already there is kept")
	c.False(SettleModifierChoices(nil, choice), "a settled choice has nothing to settle")

	for _, one := range choice.Children {
		one.SetEnabled(true)
	}
	data, err := json.Marshal(choice)
	c.NoError(err)
	var loaded TraitModifier
	c.NoError(json.Unmarshal(data, &loaded))
	c.True(loaded.Children[0].Enabled(), "loading keeps the first option enabled")
	c.False(loaded.Children[1].Enabled() || loaded.Children[2].Enabled(), "and no other")

	group := NewTraitModifier(nil, nil, true)
	a := NewTraitModifier(nil, group, false)
	b := NewTraitModifier(nil, group, false)
	group.Children = []*TraitModifier{a, b}
	ConvertToModifierChoice(group)
	c.True(a.Enabled())
	c.False(b.Enabled(), "a group that becomes a choice keeps only its first enabled option")

	entity := NewEntity()
	onSheet := NewTraitModifier(entity, nil, true)
	first := NewTraitModifier(entity, onSheet, false)
	first.SetEnabled(false)
	onSheet.Children = []*TraitModifier{first}
	ConvertToModifierChoice(onSheet)
	c.True(first.Enabled(), "on a sheet a new mandatory choice gets its first option picked")
}

// TestEquipmentRangesWithMandatoryModifierChoice verifies that equipment whose mandatory modifier choice is still to be
// made reports the range of values and weights its options give it, both for one of it and scaled by its quantity.
func TestEquipmentRangesWithMandatoryModifierChoice(t *testing.T) {
	c := check.New(t)
	newOption := func(parent *EquipmentModifier, cost, weight string) {
		m := NewEquipmentModifier(nil, parent, false)
		m.CostAmount = cost
		m.WeightAmount = weight
		m.SetEnabled(false)
		parent.Children = append(parent.Children, m)
	}
	eqp := newEquipmentItem("Sword", "100", "3 lb")
	eqp.Quantity = fxp.FromInteger(2)
	choice := NewEquipmentModifierChoice(nil, nil)
	newOption(choice, "+50", "+1 lb")
	newOption(choice, "+100", "+2 lb")
	eqp.AddModifiers(choice)

	c.Equal("150~200", FormatValueRange(eqp.AdjustedValueRange(), fxp.Int.Comma), "one of it alone is open too")
	c.Equal("4~5 lb", FormatWeightRange(eqp.AdjustedWeightRange(fxp.Pound), fxp.Pound.Format))
	c.Equal("300~400", FormatValueRange(eqp.ExtendedValueRange(), fxp.Int.Comma))
	c.Equal("8~10 lb", FormatWeightRange(eqp.ExtendedWeightRange(fxp.Pound), fxp.Pound.Format))
	c.Equal(fxp.FromInteger(150), eqp.AdjustedValue(), "a single value is the least the choice may come to")
	c.Equal(fxp.Weight(fxp.FromInteger(4)), eqp.AdjustedWeight(false, fxp.Pound))
	c.Equal(fxp.FromInteger(300), eqp.ExtendedValue())
	c.Equal(fxp.Weight(fxp.FromInteger(8)), eqp.ExtendedWeight(false, fxp.Pound))
	var data CellData
	eqp.CellData(EquipmentCostColumn, &data)
	c.Equal("150~200", data.Primary, "the value column shows the range")
	data = CellData{}
	eqp.CellData(EquipmentWeightColumn, &data)
	c.Equal("4~5 lb", data.Primary, "the weight column shows the range")

	// A container's own modifiers are costed with its contents as they are.
	pack := NewEquipment(nil, nil, true)
	pack.BaseValue = "10"
	packChoice := NewEquipmentModifierChoice(nil, nil)
	newOption(packChoice, "+1", "")
	newOption(packChoice, "+2", "")
	pack.AddModifiers(packChoice)
	item := newEquipmentItem("Rope", "5", "1 lb")
	item.SetParent(pack)
	pack.Children = []*Equipment{item}
	c.Equal("16~17", FormatValueRange(pack.ExtendedValueRange(), fxp.Int.Comma))
	c.Equal(fxp.FromInteger(16), pack.ExtendedValue())

	eqp.Preconfigured = true
	c.Equal("300~400", FormatValueRange(eqp.ExtendedValueRange(), fxp.Int.Comma),
		"a preconfigured item still has an unresolved choice to make")
	choice.Children[0].SetEnabled(true)
	c.Equal("300", FormatValueRange(eqp.ExtendedValueRange(), fxp.Int.Comma), "and takes a pick already made")
	c.Equal("8 lb", FormatWeightRange(eqp.ExtendedWeightRange(fxp.Pound), fxp.Pound.Format))
	c.Equal(fxp.FromInteger(300), eqp.ExtendedValue(), "once made, the options enabled count as they always have")
}

// TestPerPoundCostFollowsEachOptionsWeight verifies that a cost per pound is worked out from the weight each way of
// making a mandatory choice gives, rather than from the weight the equipment would have with the lightest option.
func TestPerPoundCostFollowsEachOptionsWeight(t *testing.T) {
	c := check.New(t)
	eqp := newEquipmentItem("Crate", "10", "1 lb")
	perPound := NewEquipmentModifier(nil, nil, false)
	perPound.CostAmount = "+1"
	perPound.CostIsPerPound = true
	choice := NewEquipmentModifierChoice(nil, nil)
	for _, weight := range []string{"+1 lb", "+9 lb"} {
		option := NewEquipmentModifier(nil, choice, false)
		option.WeightAmount = weight
		option.SetEnabled(false)
		choice.Children = append(choice.Children, option)
	}
	eqp.AddModifiers(perPound, choice)
	c.Equal("2~10 lb", FormatWeightRange(eqp.AdjustedWeightRange(fxp.Pound), fxp.Pound.Format))
	c.Equal("12~20", FormatValueRange(eqp.AdjustedValueRange(), fxp.Int.Comma),
		"the heavier option costs more per pound")
}

// TestLootIsASheet verifies that equipment on a loot sheet has its modifier choices made, just as it would on a
// character sheet, since both ask for them on arrival.
func TestLootIsASheet(t *testing.T) {
	c := check.New(t)
	loot := NewLoot()
	eqp := NewEquipment(loot, nil, false)
	eqp.BaseValue = "100"
	choice := NewEquipmentModifierChoice(loot, nil)
	option := NewEquipmentModifier(loot, choice, false)
	option.CostAmount = "+50"
	choice.Children = []*EquipmentModifier{option}
	eqp.AddModifiers(choice)
	c.True(IsOnSheet(eqp))
	c.True(IsOnSheet(option))
	c.Equal("150", FormatValueRange(eqp.ExtendedValueRange(), fxp.Int.Comma), "the choice is made, so settled")
	c.True(IsLockedModifierChoiceSelection(option))
	c.False(IsOnSheet(NewEquipment(nil, nil, false)))
}

// TestUnnestingAChoiceKeepsTheOuterPick verifies that a choice within a choice turned back into a group gives its
// options to the choice around it, which keeps the pick it had.
func TestUnnestingAChoiceKeepsTheOuterPick(t *testing.T) {
	c := check.New(t)
	outer := newTraitModifierChoiceWith(false, "+1")
	outer.Children[0].SetEnabled(true)
	inner := NewTraitModifierChoice(nil, outer)
	outer.Children = append(outer.Children, inner)
	innerOption := newTraitModifierOption(inner, "Inner", "+2")
	ConvertFromModifierChoice(inner)
	c.Equal([]*TraitModifier{outer.Children[0], innerOption}, ModifierChoiceOptions(outer))
	c.True(outer.Children[0].Enabled(), "the outer choice keeps its pick")
	c.False(innerOption.Enabled(), "the option that came with the group is turned off")
}

// TestKeepModifierChoiceRules verifies the rules applied after a modifier's enabled state is set directly, as its
// editor does: an option turned on becomes the pick, and on a sheet the pick of a mandatory choice can't be turned off.
func TestKeepModifierChoiceRules(t *testing.T) {
	c := check.New(t)
	choice := newTraitModifierChoiceWith(true, "+1", "+2")
	a, b := choice.Children[0], choice.Children[1]
	a.SetEnabled(true)
	b.SetEnabled(true)
	KeepModifierChoiceRules(b, false)
	c.False(a.Enabled(), "turning b on makes it the pick")
	c.True(b.Enabled())
	b.SetEnabled(false)
	KeepModifierChoiceRules(b, true)
	c.False(b.Enabled(), "off a sheet a mandatory choice may be left without its pick")

	trait := NewTrait(NewEntity(), nil, false)
	onSheet := newTraitModifierChoiceWith(true, "+1")
	trait.AddModifiers(onSheet)
	pick := onSheet.Children[0]
	pick.SetEnabled(false)
	KeepModifierChoiceRules(pick, true)
	c.True(pick.Enabled(), "on a sheet the pick of a mandatory choice is turned back on")
}

// TestModifierEnabledChanges verifies the changes worked out for setting modifiers' enabled states: turning an option
// on turns the choice's other options off, the later of two turned on together winning, and a request to turn off the
// pick of a mandatory choice on a sheet is left out.
func TestModifierEnabledChanges(t *testing.T) {
	c := check.New(t)
	choice := newTraitModifierChoiceWith(false, "+1", "+2", "+3")
	a, b, cc := choice.Children[0], choice.Children[1], choice.Children[2]
	a.SetEnabled(true)
	targets, enabled := ModifierEnabledChanges([]*TraitModifier{b, cc}, func(*TraitModifier) bool { return true })
	c.Equal([]*TraitModifier{a, b, cc}, targets)
	c.Equal(map[*TraitModifier]bool{a: false, b: false, cc: true}, enabled)

	trait := NewTrait(NewEntity(), nil, false)
	onSheet := newTraitModifierChoiceWith(true, "+1")
	trait.AddModifiers(onSheet)
	onSheet.Children[0].SetEnabled(true)
	targets, _ = ModifierEnabledChanges(onSheet.Children, func(*TraitModifier) bool { return false })
	c.Equal(0, len(targets), "the pick of a mandatory choice on a sheet can't be turned off")
}

// TestSyncSettlesAModifierChoice verifies that a group which its library source has since made a choice keeps no more
// than one of its options enabled once synced.
func TestSyncSettlesAModifierChoice(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	libFile := LibraryFile{Library: "Test Library", Path: "Test" + TraitModifiersExt}
	source := NewTraitModifierChoice(nil, nil)
	local := NewTraitModifier(entity, nil, true)
	a := NewTraitModifier(entity, local, false)
	b := NewTraitModifier(entity, local, false)
	local.Children = []*TraitModifier{a, b}
	local.Source = Source{LibraryFile: libFile, TID: source.TID}
	entity.SourceMatcher().libHashes = map[LibraryFile]libSrcData{
		libFile: {dataHashes: map[tid.TID]HashAndData{source.TID: {Hash: Hash64(source), Data: source}}},
	}
	local.SyncWithSource()
	c.True(IsMandatoryModifierChoice(local), "the sync brings the choice across")
	c.True(a.Enabled(), "the first enabled option is kept")
	c.False(b.Enabled(), "and no other")
}

// TestLoadingKeepsThePicksOfUnsupportedForms verifies that loading keeps a choice held in a supported form to one
// enabled option, but leaves the options of one held in a form a newer version may allow several picks for as they are.
func TestLoadingKeepsThePicksOfUnsupportedForms(t *testing.T) {
	c := check.New(t)
	load := func(tp TemplatePicker) *TraitModifier {
		t.Helper()
		choice := NewTraitModifierChoice(nil, nil)
		choice.Choice = tp
		for range 2 {
			newTraitModifierOption(choice, "Option", "+1")
		}
		data, err := json.Marshal(choice)
		c.NoError(err)
		var loaded TraitModifier
		c.NoError(json.Unmarshal(data, &loaded))
		return &loaded
	}
	atLeastOne := TemplatePicker{Type: picker.Count, Qualifier: criteria.Number{
		Compare:   criteria.AtLeastNumber,
		Qualifier: fxp.One,
	}}
	loaded := load(atLeastOne)
	c.True(loaded.Children[0].Enabled() && loaded.Children[1].Enabled(), "a form this version doesn't support keeps its picks")
	c.Equal(atLeastOne, loaded.Choice)
	loaded = load(newModifierChoicePicker(false))
	c.True(loaded.Children[0].Enabled())
	c.False(loaded.Children[1].Enabled(), "a supported form keeps no more than one")
}

// TestPastTheCapTheCurrentPicksCount verifies that with more ways of making the open choices than are worked through,
// the choices count as made with the picks they have, so that every figure agrees with every other.
func TestPastTheCapTheCurrentPicksCount(t *testing.T) {
	c := check.New(t)
	eqp := newEquipmentItem("Crate", "100", "10 lb")
	eqp.Quantity = fxp.FromInteger(2)
	for range 13 {
		choice := NewEquipmentModifierChoice(nil, nil)
		for range 2 {
			option := NewEquipmentModifier(nil, choice, false)
			option.WeightAmount = "+1 lb"
			option.CostAmount = "+1"
			option.SetEnabled(false)
			choice.Children = append(choice.Children, option)
		}
		eqp.AddModifiers(choice)
	}
	c.Equal("10 lb", FormatWeightRange(eqp.AdjustedWeightRange(fxp.Pound), fxp.Pound.Format))
	c.Equal(fxp.Weight(fxp.FromInteger(10)), eqp.AdjustedWeight(false, fxp.Pound))
	c.Equal(fxp.Weight(fxp.FromInteger(20)), eqp.ExtendedWeight(false, fxp.Pound))
	c.Equal("200", FormatValueRange(eqp.ExtendedValueRange(), fxp.Int.Comma))
	c.Equal(fxp.FromInteger(200), eqp.ExtendedValue())
}

// TestConvertingAGroupWithinAChoiceOnASheet verifies that a group within a mandatory choice on a sheet, holding that
// choice's pick, which becomes a choice of its own, leaves the choice around it with a pick: the first of the options
// left to it.
func TestConvertingAGroupWithinAChoiceOnASheet(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	outer := NewTraitModifierChoice(entity, nil)
	group := NewTraitModifier(entity, outer, true)
	pick := NewTraitModifier(entity, group, false)
	group.Children = []*TraitModifier{pick}
	other := NewTraitModifier(entity, outer, false)
	other.SetEnabled(false)
	outer.Children = []*TraitModifier{group, other}
	trait := NewTrait(entity, nil, false)
	trait.AddModifiers(outer)
	c.True(ModifierChoiceIsResolved(outer))
	ConvertToModifierChoice(group)
	c.True(pick.Enabled(), "the new choice keeps the pick it holds")
	c.True(other.Enabled(), "the choice around it gets a pick of its own")
	c.True(ModifierChoiceIsResolved(outer))
}

// hashOf is a Hashable made of a function, for working out what a hash is expected to be.
type hashOf func(h hash.Hash)

func (f hashOf) Hash(h hash.Hash) { f(h) }

// TestModifierContainerHashes verifies that a group hashes just as a modifier container always has, so that copies in
// existing sheets don't show as out of step with their library after an upgrade, and that what a choice asks for is
// part of its hash.
func TestModifierContainerHashes(t *testing.T) {
	c := check.New(t)
	group := NewTraitModifier(nil, nil, true)
	group.Name = "Options"
	c.Equal(Hash64(hashOf(func(h hash.Hash) {
		group.TraitModifierSyncData.hash(h)
		xhash.Num8(h, uint8(255))
	})), Hash64(group), "a group hashes as a container always has")
	choice := NewTraitModifierChoice(nil, nil)
	mandatory := Hash64(choice)
	choice.SetMandatoryChoice(false)
	c.NotEqual(mandatory, Hash64(choice), "mandatory and optional choices differ")
}
