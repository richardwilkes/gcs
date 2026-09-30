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
	"time"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
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

// newEquipmentModifierChoiceWith returns a mandatory equipment modifier choice holding an option for each pair of cost
// and weight adjustments, with none of them picked.
func newEquipmentModifierChoiceWith(options ...[2]string) *EquipmentModifier {
	choice := NewEquipmentModifierChoice(nil, nil)
	for _, one := range options {
		option := NewEquipmentModifier(nil, choice, false)
		option.CostAmount = one[0]
		option.WeightAmount = one[1]
		option.SetEnabled(false)
		choice.Children = append(choice.Children, option)
	}
	return choice
}

// linkToLibrarySource makes source the library version of the modifier whose Source is local, as the entity's source
// matcher sees it.
func linkToLibrarySource[T interface {
	Hashable
	ID() tid.TID
}](entity *Entity, local *Source, source T) {
	libFile := LibraryFile{Library: "Test Library", Path: "Test"}
	*local = Source{LibraryFile: libFile, TID: source.ID()}
	entity.SourceMatcher().libHashes = map[LibraryFile]libSrcData{
		libFile: {dataHashes: map[tid.TID]HashAndData{source.ID(): {Hash: Hash64(source), Data: source}}},
	}
}

// hashOf is a Hashable made of a function, for working out what a hash is expected to be.
type hashOf func(h hash.Hash)

func (f hashOf) Hash(h hash.Hash) { f(h) }

// TestModifierContainerKinds verifies that a new modifier container is a group and a new modifier choice a mandatory
// one, that only a group converts to a choice and back, and that a group hashes as a modifier container always has, so
// that copies on existing sheets don't show as out of step with their library after an upgrade.
func TestModifierContainerKinds(t *testing.T) {
	c := check.New(t)
	group := NewTraitModifier(nil, nil, true)
	c.False(IsModifierChoice(group))
	c.Equal(Hash64(hashOf(func(h hash.Hash) {
		group.TraitModifierSyncData.hash(h)
		xhash.Num8(h, uint8(255))
	})), Hash64(group), "a group hashes as a container always has")
	c.True(CanConvertToModifierChoice(group))
	ConvertToModifierChoice(group)
	c.True(IsMandatoryModifierChoice(group))
	c.Equal("Pick 1", ModifierChoiceDescription(group))
	c.False(CanConvertToModifierChoice(group), "a choice is already a choice")
	c.False(IsTemplateChoiceContainer(group), "a modifier choice is not a template choice")
	mandatory := Hash64(group)
	group.SetMandatoryChoice(false)
	c.Equal("Pick at most 1", ModifierChoiceDescription(group))
	c.NotEqual(mandatory, Hash64(group), "what a choice asks for is part of its hash")
	ConvertFromModifierChoice(group)
	c.False(IsModifierChoice(group))

	c.True(IsMandatoryModifierChoice(NewTraitModifierChoice(nil, nil)))
	c.True(IsMandatoryModifierChoice(NewEquipmentModifierChoice(nil, nil)))
	c.False(IsModifierChoice(NewTraitModifier(nil, nil, false)), "a modifier that isn't a container is never a choice")
	c.False(CanConvertToModifierChoice(NewEquipmentModifier(nil, nil, false)))
	c.True(CanConvertToModifierChoice(NewEquipmentModifier(nil, nil, true)))
	c.False(CanConvertToModifierChoice(NewTrait(nil, nil, true)), "only a modifier container can become one")
}

// TestModifierChoiceNormalizesOnLoad verifies that loading rewrites a choice in the nearer supported form, an exact
// count of one or more as mandatory and any other as optional, and settles its picks, and that neither a group nor a
// modifier that isn't a container keeps picker data, nor a container its VTT notes.
func TestModifierChoiceNormalizesOnLoad(t *testing.T) {
	c := check.New(t)
	load := func(container bool, tp TemplatePicker) *TraitModifier {
		t.Helper()
		m := NewTraitModifier(nil, nil, container)
		if container {
			newTraitModifierOption(m, "A", "+1")
			newTraitModifierOption(m, "B", "+1")
		}
		data, err := json.Marshal(m)
		c.NoError(err)
		pickerData, err := json.Marshal(tp)
		c.NoError(err)
		data = append(data[:len(data)-1], `,"vtt_notes":"notes","choice":`+string(pickerData)+"}"...)
		var loaded TraitModifier
		c.NoError(json.Unmarshal(data, &loaded))
		return &loaded
	}
	number := func(compare criteria.NumericComparison, qualifier int) criteria.Number {
		return criteria.Number{Compare: compare, Qualifier: fxp.FromInteger(qualifier)}
	}
	for tp, mandatory := range map[TemplatePicker]bool{
		{Type: picker.Count, Qualifier: number(criteria.EqualsNumber, 3)}:   true,
		{Type: picker.Count, Qualifier: number(criteria.AtLeastNumber, 1)}:  false,
		{Type: picker.Count, Qualifier: number(criteria.EqualsNumber, 0)}:   false,
		{Type: picker.Points, Qualifier: number(criteria.EqualsNumber, 10)}: false,
	} {
		choice := load(true, tp)
		c.Equal(newModifierChoicePicker(mandatory), choice.Choice)
		c.True(choice.Children[0].Enabled())
		c.False(choice.Children[1].Enabled(), "its picks are settled")
		c.Equal("", choice.VTTNotes, "a container keeps no VTT notes")
	}
	group := load(true, TemplatePicker{Qualifier: number(criteria.AtMostNumber, 2)})
	c.Equal(TemplatePicker{}, group.Choice, "a group keeps no stray picker data")
	leaf := load(false, newModifierChoicePicker(true))
	c.Equal(TemplatePicker{}, leaf.Choice, "a modifier that isn't a container keeps no picker data")
	c.Equal("notes", leaf.VTTNotes)
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

// newTraitWithPoints returns a trait that isn't a container, costing the given points, held by parent.
func newTraitWithPoints(parent *Trait, points int) *Trait {
	trait := NewTrait(nil, parent, false)
	trait.BasePoints = fxp.FromInteger(points)
	parent.Children = append(parent.Children, trait)
	return trait
}

// newTraitContainerWithChoice returns a trait container of the given type whose modifiers are a mandatory choice
// holding an option for each cost adjustment, with none of them picked, and which holds a trait costing each of the
// given points.
func newTraitContainerWithChoice(containerType container.Type, costs []string, points ...int) *Trait {
	parent := NewTrait(nil, nil, true)
	parent.ContainerType = containerType
	parent.AddModifiers(newTraitModifierChoiceWith(true, costs...))
	for _, one := range points {
		newTraitWithPoints(parent, one)
	}
	return parent
}

// TestTraitContainerChoiceIsCostedAsAWhole verifies that a mandatory choice among a container's own modifiers is made
// once for the whole container, each way of making it costing every trait inside with the same pick, rather than each
// trait being costed with whichever pick suits it. Wherever a single cost is needed, the container counts as the least
// of those.
func TestTraitContainerChoiceIsCostedAsAWhole(t *testing.T) {
	c := check.New(t)
	costs := []string{"-50%", "+50%"}
	group := newTraitContainerWithChoice(container.Group, costs, 10, -10)
	c.Equal("0", group.PointsRange(nil).String(), "either pick makes the two cancel out")
	c.Equal(fxp.Int(0), group.AdjustedPoints(nil))
	c.Equal("5~15", group.Children[0].PointsRange(nil).String(), "each trait alone still shows its own range")

	abilities := newTraitContainerWithChoice(container.AlternativeAbilities, costs, 10, -10)
	c.Equal("4~12", abilities.PointsRange(nil).String())
	c.Equal(fxp.FromInteger(4), abilities.AdjustedPoints(nil))

	abilities.Modifiers[0].Children[1].SetEnabled(true)
	c.Equal("4~12", abilities.PointsRange(nil).String(), "a pick made outside a sheet is only a default until asked")
	abilities.Preconfigured = true
	c.Equal("12", abilities.PointsRange(nil).String(), "a preconfigured container takes the pick already made")
	c.Equal(fxp.FromInteger(12), abilities.AdjustedPoints(nil))

	// A container within a container makes its own choice for each way the outer one's is made. The four ways cost -3,
	// 6, -2 and 8, the inner container coming to 6, 15, 6 and 16 of them.
	outer := newTraitContainerWithChoice(container.Group, []string{"+1", "+2"})
	inner := newTraitContainerWithChoice(container.AlternativeAbilities, costs, 10, -10)
	inner.SetParent(outer)
	outer.Children = append(outer.Children, inner)
	newTraitWithPoints(outer, -10)
	c.Equal("-3~8", outer.PointsRange(nil).String())
	c.Equal(fxp.FromInteger(-3), outer.AdjustedPoints(nil))
	c.Equal("6~16", inner.PointsRange(nil).String(), "the inner container alone works through both choices")
	outer.Preconfigured = true
	outer.Modifiers[0].Children[1].SetEnabled(true)
	c.Equal("-2~8", outer.PointsRange(nil).String(), "a preconfigured outer container settles only its own choice")
	c.Equal("6~16", inner.PointsRange(nil).String())
}

// TestTraitContainerChoicesPastTheCap verifies that the ways of making the open choices of a container and those of a
// trait inside it multiply together, and that past the cap every choice within the container counts as made with the
// picks it has, even for a trait inside that would be within the cap alone.
func TestTraitContainerChoicesPastTheCap(t *testing.T) {
	c := check.New(t)
	build := func(containerChoices int) *Trait {
		parent := NewTrait(nil, nil, true)
		for range containerChoices {
			parent.AddModifiers(newTraitModifierChoiceWith(true, "+0", "+1"))
		}
		newTraitWithPoints(parent, 10).AddModifiers(newTraitModifierChoiceWith(true, "+0", "+1"))
		newTraitWithPoints(parent, 10)
		return parent
	}
	within := build(11)
	c.Equal("20~43", within.PointsRange(nil).String(), "2048 ways by 2 is within the cap")
	past := build(12)
	c.Equal("10~22", past.Children[1].PointsRange(nil).String(), "4096 ways alone is within it")
	c.Equal("20", past.PointsRange(nil).String(), "but 4096 ways by 2 is past it")
	c.Equal(fxp.FromInteger(20), past.AdjustedPoints(nil))
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
// a choice already had over options that have just arrived.
func TestSettleModifierChoices(t *testing.T) {
	c := check.New(t)
	choice := newTraitModifierChoiceWith(false, "+1", "+2", "+3")
	for _, one := range choice.Children {
		one.SetEnabled(true)
	}
	arrived := choice.Children[0]
	c.True(SettleModifierChoices(func(m *TraitModifier) bool { return m == arrived }, choice))
	c.Equal([]bool{false, true, false},
		[]bool{choice.Children[0].Enabled(), choice.Children[1].Enabled(), choice.Children[2].Enabled()},
		"the first option that was already there is kept")
	c.False(SettleModifierChoices(nil, choice), "a settled choice has nothing to settle")
}

// TestEquipmentRangesWithMandatoryModifierChoice verifies that equipment whose mandatory modifier choice is still to be
// made reports the range of values and weights its options give it, both for one of it and scaled by its quantity.
func TestEquipmentRangesWithMandatoryModifierChoice(t *testing.T) {
	c := check.New(t)
	eqp := newEquipmentItem("Sword", "100", "3 lb")
	eqp.Quantity = fxp.FromInteger(2)
	choice := newEquipmentModifierChoiceWith([2]string{"+50", "+1 lb"}, [2]string{"+100", "+2 lb"})
	eqp.AddModifiers(choice)

	c.Equal("150~200", FormatValueRange(eqp.adjustedValueRange(), fxp.Int.Comma), "one of it alone is open too")
	c.Equal("4~5 lb", FormatWeightRange(eqp.adjustedWeightRange(fxp.Pound), fxp.Pound.Format))
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
	pack.AddModifiers(newEquipmentModifierChoiceWith([2]string{"+1", ""}, [2]string{"+2", ""}))
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

// TestPromptedRanges verifies that the prompted ranges cost an item as the modifier prompt will see it, the same on a
// sheet as off one, and that an item taken counts as preconfigured.
func TestPromptedRanges(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	newTrait := func(owner DataOwner) *Trait {
		trait := NewTrait(owner, nil, false)
		trait.BasePoints = fxp.FromInteger(10)
		trait.AddModifiers(newTraitModifierChoiceWith(true, "+5", "+10"))
		return trait
	}
	all := func(*Trait) bool { return true }
	for _, owner := range []DataOwner{nil, entity} {
		trait := newTrait(owner)
		c.Equal("15~20", trait.PromptedPointsRange(nil).String())
		c.Equal("15~20", trait.PromptedPointsRange(all).String(), "a choice with no pick is still open when taken")
		trait.Modifiers[0].Children[1].SetEnabled(true)
		c.Equal("15~20", trait.PromptedPointsRange(nil).String(), "a pick is only a default until asked")
		c.Equal("20", trait.PromptedPointsRange(all).String(), "a taken trait keeps its pick")
		trait.Preconfigured = true
		c.Equal("20", trait.PromptedPointsRange(nil).String(), "as does a preconfigured one")
	}
	c.Equal("10", newTrait(entity).PointsRange(nil).String(), "on a sheet the choice counts as made")

	// A choice inherited from a container is taken with the container, not the trait asked about.
	parent := NewTrait(entity, nil, true)
	parent.AddModifiers(newTraitModifierChoiceWith(true, "+1", "+3"))
	parent.Modifiers[0].Children[0].SetEnabled(true)
	child := newTraitWithPoints(parent, 10)
	child.SetDataOwner(entity)
	c.Equal("11~13", parent.PromptedPointsRange(nil).String())
	c.Equal("11~13", child.PromptedPointsRange(func(one *Trait) bool { return one == child }).String())
	c.Equal("11", child.PromptedPointsRange(func(one *Trait) bool { return one == parent }).String())
	parent.Modifiers[0].Children[0].SetEnabled(false)
	c.Equal("10", child.PromptedPointsRange(func(one *Trait) bool { return one == parent }).String(),
		"the container answers it for the trait, picked or not")
	c.Equal("11~13", parent.PromptedPointsRange(func(one *Trait) bool { return one == parent }).String())

	taken := func(*Equipment) bool { return true }
	for _, owner := range []DataOwner{nil, entity} {
		eqp := newEquipmentItem("Sword", "100", "3 lb")
		eqp.SetDataOwner(owner)
		eqp.Quantity = fxp.FromInteger(2)
		choice := newEquipmentModifierChoiceWith([2]string{"+50", "+1 lb"}, [2]string{"+100", "+2 lb"})
		eqp.AddModifiers(choice)
		choice.Children[1].SetEnabled(true)
		c.Equal("300~400", FormatValueRange(eqp.PromptedExtendedValueRange(nil), fxp.Int.Comma))
		c.Equal("8~10 lb", FormatWeightRange(eqp.PromptedExtendedWeightRange(fxp.Pound, nil), fxp.Pound.Format))
		c.Equal("400", FormatValueRange(eqp.PromptedExtendedValueRange(taken), fxp.Int.Comma))
		c.Equal("10 lb", FormatWeightRange(eqp.PromptedExtendedWeightRange(fxp.Pound, taken), fxp.Pound.Format))
		if owner != nil {
			c.Equal("400", FormatValueRange(eqp.ExtendedValueRange(), fxp.Int.Comma), "on a sheet the pick counts")
		}
	}
}

// newPerPoundModifier returns an equipment modifier adding the given cost per pound.
func newPerPoundModifier(cost string) *EquipmentModifier {
	m := NewEquipmentModifier(nil, nil, false)
	m.CostAmount = cost
	m.CostIsPerPound = true
	return m
}

// TestPerPoundCostWithOpenChoices verifies that a cost per pound is by the least the equipment's own weight may come
// to, seen as its value is, and that costing it doesn't enumerate the weight for each way of making the choices.
func TestPerPoundCostWithOpenChoices(t *testing.T) {
	c := check.New(t)
	for _, owner := range []DataOwner{nil, NewEntity()} {
		eqp := newEquipmentItem("Thing", "100", "3 lb")
		eqp.SetDataOwner(owner)
		eqp.AddModifiers(newPerPoundModifier("+1"),
			newEquipmentModifierChoiceWith([2]string{"+1", "+1 lb"}, [2]string{"+2", "+2 lb"}))
		c.Equal("105~106", FormatValueRange(PickerMeasureRange(eqp, picker.Value, true, nil), fxp.Int.Comma))
	}

	eqp := newEquipmentItem("Thing", "100", "3 lb")
	eqp.AddModifiers(newPerPoundModifier("+1"))
	for range 12 {
		eqp.AddModifiers(newEquipmentModifierChoiceWith([2]string{"+1", "+1 lb"}, [2]string{"+2", "+2 lb"}))
	}
	start := time.Now()
	c.Equal("127~139", FormatValueRange(eqp.ExtendedValueRange(), fxp.Int.Comma))
	c.True(time.Since(start) < 10*time.Second, "costing it must not multiply the ways of making the choices")
}

// TestTemplateChoiceOfEquipmentWithAnOpenModifierChoice verifies that a template choice of equipment counts an option
// whose mandatory modifier choice is still to be made by the range its modifiers give it.
func TestTemplateChoiceOfEquipmentWithAnOpenModifierChoice(t *testing.T) {
	c := check.New(t)
	pick := newEquipmentChoice("Rope")
	eqp := newEquipmentItem("Sword", "100", "3 lb")
	eqp.Quantity = fxp.FromInteger(2)
	eqp.AddModifiers(newEquipmentModifierChoiceWith([2]string{"+50", "+1 lb"}, [2]string{"+100", "+2 lb"}))
	eqp.SetParent(pick)
	pick.Children = append(pick.Children, eqp)
	c.Equal("10~400", FormatValueRange(pick.ExtendedValueRange(), fxp.Int.Comma))
	c.Equal("1~10 lb", FormatWeightRange(pick.ExtendedWeightRange(fxp.Pound), fxp.Pound.Format))
	c.Equal(fxp.FromInteger(150), eqp.ExtendedValueOfJustOne(), "one of it counts as the least it may come to")
}

// TestWeightIgnoredForSkillsWithAnOpenChoice verifies that equipment whose weight is ignored for skills weighs nothing
// for them while a choice is still to be made, whether it is a mandatory choice among its own modifiers or a template
// choice holding it, just as it does once the choice is made.
func TestWeightIgnoredForSkillsWithAnOpenChoice(t *testing.T) {
	c := check.New(t)
	eqp := newEquipmentItem("Armor", "100", "3 lb")
	eqp.Equipped = true
	eqp.WeightIgnoredForSkills = true
	c.Equal(fxp.Weight(0), eqp.ExtendedWeight(true, fxp.Pound))
	eqp.AddModifiers(newEquipmentModifierChoiceWith([2]string{"", "+1 lb"}, [2]string{"", "+2 lb"}))
	c.Equal(fxp.Weight(0), eqp.ExtendedWeight(true, fxp.Pound), "an open modifier choice doesn't change that")
	c.Equal(fxp.Weight(fxp.FromInteger(4)), eqp.ExtendedWeight(false, fxp.Pound))

	pick := newEquipmentChoice()
	eqp.SetParent(pick)
	pick.Children = []*Equipment{eqp}
	c.Equal(fxp.Weight(0), pick.ExtendedWeight(true, fxp.Pound), "nor does a template choice holding it")
	c.Equal(fxp.Weight(fxp.FromInteger(4)), pick.ExtendedWeight(false, fxp.Pound))
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

// TestModifierEnabledChanges verifies the changes worked out for setting modifiers' enabled states: turning an option
// on turns the choice's other options off, the later of two turned on together winning, a modifier whose state doesn't
// change is left out, and so is a request to turn off the pick of a mandatory choice on a sheet.
func TestModifierEnabledChanges(t *testing.T) {
	c := check.New(t)
	choice := newTraitModifierChoiceWith(false, "+1", "+2", "+3")
	a, b, cc := choice.Children[0], choice.Children[1], choice.Children[2]
	a.SetEnabled(true)
	targets, enabled := ModifierEnabledChanges([]*TraitModifier{b, cc}, func(*TraitModifier) bool { return true })
	c.Equal([]*TraitModifier{a, cc}, targets, "b is turned on and off again, so it doesn't change")
	c.Equal(map[*TraitModifier]bool{a: false, cc: true}, enabled)
	targets, _ = ModifierEnabledChanges([]*TraitModifier{a}, func(*TraitModifier) bool { return true })
	c.Equal(0, len(targets), "a modifier already in the state asked for doesn't change")

	trait := NewTrait(NewEntity(), nil, false)
	onSheet := newTraitModifierChoiceWith(true, "+1")
	trait.AddModifiers(onSheet)
	onSheet.Children[0].SetEnabled(true)
	targets, _ = ModifierEnabledChanges(onSheet.Children, func(*TraitModifier) bool { return false })
	c.Equal(0, len(targets), "the pick of a mandatory choice on a sheet can't be turned off")
}

// TestSyncSettlesAModifierChoice verifies that a group which its library source has since made a choice gets what the
// choice asks for but not its options, and keeps no more than one of its options enabled once synced. One with none
// enabled is left without a pick, flagged on a sheet as any other.
func TestSyncSettlesAModifierChoice(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	source := newTraitModifierChoiceWith(true, "+5")
	for _, enabled := range []bool{true, false} {
		local := NewTraitModifier(entity, nil, true)
		a := NewTraitModifier(entity, local, false)
		b := NewTraitModifier(entity, local, false)
		a.SetEnabled(enabled)
		b.SetEnabled(enabled)
		local.Children = []*TraitModifier{a, b}
		linkToLibrarySource(entity, &local.Source, source)
		local.SyncWithSource()
		c.True(IsMandatoryModifierChoice(local), "the sync brings the choice across")
		c.Equal([]*TraitModifier{a, b}, local.Children, "but not its options")
		c.Equal(enabled, a.Enabled(), "the first enabled option is kept")
		c.False(b.Enabled(), "and no other")
		var data CellData
		local.CellData(TraitModifierDescriptionColumn, &data)
		c.Equal(!enabled, data.ChoiceRequired, "a choice left without a pick is flagged")
	}
}

// TestSyncSettlesTheOuterChoice verifies that a choice within a choice which its library source has since made a group
// gives its options to the choice around it, which keeps the pick it had.
func TestSyncSettlesTheOuterChoice(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	outer := NewEquipmentModifierChoice(entity, nil)
	outer.SetMandatoryChoice(false)
	p := NewEquipmentModifier(entity, outer, false)
	inner := NewEquipmentModifierChoice(entity, outer)
	q := NewEquipmentModifier(entity, inner, false)
	inner.Children = []*EquipmentModifier{q}
	outer.Children = []*EquipmentModifier{p, inner}
	linkToLibrarySource(entity, &inner.Source, NewEquipmentModifier(nil, nil, true))
	inner.SyncWithSource()
	c.False(IsModifierChoice(inner), "the sync makes the inner choice a group")
	c.True(p.Enabled(), "the outer choice keeps its pick")
	c.False(q.Enabled(), "the option that came with the group is turned off")
}

// TestContainedWeightReductionOfEachOption verifies that each option of a mandatory choice on a container reduces the
// weight of its contents by its own reduction, paired with its own weight, rather than the least weight of one option
// being paired with the greatest reduction of another.
func TestContainedWeightReductionOfEachOption(t *testing.T) {
	c := check.New(t)
	pack := NewEquipment(nil, nil, true)
	pack.BaseWeight = "1 lb"
	choice := newEquipmentModifierChoiceWith([2]string{"", "+1 lb"}, [2]string{"", ""})
	reduction := NewContainedWeightReduction()
	reduction.Reduction = "50%"
	choice.Children[0].Features = Features{reduction}
	pack.AddModifiers(choice)
	item := newEquipmentItem("Rock", "0", "10 lb")
	item.SetParent(pack)
	pack.Children = []*Equipment{item}
	c.Equal("7~11 lb", FormatWeightRange(pack.ExtendedWeightRange(fxp.Pound), fxp.Pound.Format))
	c.Equal(fxp.Weight(fxp.FromInteger(7)), pack.ExtendedWeight(false, fxp.Pound))
}

// TestEditorCopyAsksAboutInheritedChoicesOnTheContainer verifies that a copy of a trait, as an editor works on, whose
// own modifiers still belong to the original, takes a choice it inherits as made when the container above it is
// preconfigured, and its own choice as made when the copy is.
func TestEditorCopyAsksAboutInheritedChoicesOnTheContainer(t *testing.T) {
	c := check.New(t)
	parent := newTraitContainerWithChoice(container.Group, []string{"+1", "+3"})
	parent.Modifiers[0].Children[1].SetEnabled(true)
	parent.Preconfigured = true
	child := newTraitWithPoints(parent, 10)
	child.AddModifiers(newTraitModifierChoiceWith(true, "+5", "+10"))
	child.Modifiers[0].Children[1].SetEnabled(true)
	editorCopy := child.Clone(LibraryFile{}, child.DataOwner(), parent, Copy)
	editorCopy.TraitEditData = child.TraitEditData
	c.Equal(child, editorCopy.Modifiers[0].Target(), "the copy's modifiers still belong to the original")
	c.Equal("18~23", editorCopy.PointsRange(nil).String(), "the container's choice is taken as made")
	editorCopy.Preconfigured = true
	c.Equal("23", editorCopy.PointsRange(nil).String(), "and the copy's own once it is preconfigured")
}

// TestPastTheCapTheCurrentPicksCount verifies that with more ways of making the open choices than are worked through,
// the choices count as made with the picks they have, so that every figure agrees with every other.
func TestPastTheCapTheCurrentPicksCount(t *testing.T) {
	c := check.New(t)
	eqp := newEquipmentItem("Crate", "100", "10 lb")
	eqp.Quantity = fxp.FromInteger(2)
	for range 13 {
		eqp.AddModifiers(newEquipmentModifierChoiceWith([2]string{"+1", "+1 lb"}, [2]string{"+1", "+1 lb"}))
	}
	c.Equal("10 lb", FormatWeightRange(eqp.adjustedWeightRange(fxp.Pound), fxp.Pound.Format))
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
