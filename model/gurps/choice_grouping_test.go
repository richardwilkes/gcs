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
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/eqcontainer"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// newTraitGroup returns a trait group holding the children, picked from separately when separately is true.
func newTraitGroup(name string, separately bool, children ...*Trait) *Trait {
	group := NewTrait(nil, nil, true)
	group.Name = name
	group.PickSeparately = separately
	group.Children = children
	for _, child := range children {
		child.SetParent(group)
	}
	return group
}

// newOrganizedChoice returns "Pick 1" of a, then b, c and d in organizing groups, the inner one nested, then e.
func newOrganizedChoice() (choice, outer, inner *Trait) {
	choice = newTemplateChoiceTrait("Choice", "a")
	inner = newTraitGroup("inner", true, NewTrait(nil, nil, false), NewTrait(nil, nil, false))
	inner.Children[0].Name, inner.Children[1].Name = "c", "d"
	b := NewTrait(nil, nil, false)
	b.Name = "b"
	outer = newTraitGroup("outer", true, b, inner)
	e := NewTrait(nil, choice, false)
	e.Name = "e"
	outer.SetParent(choice)
	choice.Children = append(choice.Children, outer, e)
	return choice, outer, inner
}

func TestPickSeparatelyPersistence(t *testing.T) {
	c := check.New(t)
	group := newTraitGroup("Group", false)
	data, err := json.Marshal(group)
	c.NoError(err)
	c.NotContains(string(data), "pick_separately", "the key is left out when off")
	hash := Hash64(group)

	group.PickSeparately = true
	c.Equal(hash, Hash64(group), "the flag isn't source data")
	data, err = json.Marshal(group)
	c.NoError(err)
	c.Contains(string(data), `"pick_separately":true`)
	var loaded Trait
	c.NoError(json.Unmarshal(data, &loaded))
	c.True(loaded.PickSeparately)
	c.True(group.Clone(LibraryFile{}, nil, nil, Copy).PickSeparately, "a clone keeps it")

	e := NewEntity()
	libFile := LibraryFile{Library: "Test Library", Path: "Test" + TraitsExt}
	source := newTraitGroup("Source", false)
	stubLibrarySources(t, e.SourceMatcher(), libFile, source)
	local := newTraitGroup("Local", true)
	local.SetDataOwner(e)
	local.Source = Source{LibraryFile: libFile, TID: source.TID}
	state, _ := e.SourceMatcher().Match(local)
	c.Equal(srcstate.Mismatched, state)
	local.SyncWithSource()
	c.Equal("Source", local.Name)
	c.True(local.PickSeparately, "syncing leaves it alone")
}

func TestPickSeparatelyClearing(t *testing.T) {
	c := check.New(t)
	choice, outer, inner := newOrganizedChoice()
	ClearTemplatePickerData(choice)
	c.False(outer.PickSeparately)
	c.False(inner.PickSeparately)

	choice, outer, inner = newOrganizedChoice()
	list, err := json.Marshal(&listData[*Trait]{Version: jio.CurrentDataVersion, Rows: []*Trait{choice}})
	c.NoError(err)
	rows, err := NewTraitsFromFile(fstest.MapFS{"list.adq": &fstest.MapFile{Data: list}}, "list.adq")
	c.NoError(err)
	c.False(rows[0].Children[1].PickSeparately, "a list doesn't keep it")

	outer.TemplatePicker = newTemplateChoicePicker()
	normalizeTemplateChoiceContainer(outer)
	c.False(outer.PickSeparately, "a choice is never picked from separately")

	inner.ContainerType = container.MetaTrait
	inner.ClearUnusedFieldsForType()
	c.False(inner.PickSeparately, "only a group can be")
	leaf := NewTrait(nil, nil, false)
	leaf.PickSeparately = true
	leaf.ClearUnusedFieldsForType()
	c.False(leaf.PickSeparately)
	skill := NewSkill(nil, nil, false)
	skill.PickSeparately = true
	skill.ClearUnusedFieldsForType()
	c.False(skill.PickSeparately)
	eqp := NewEquipment(nil, nil, true)
	eqp.PickSeparately = true
	eqp.ClearUnusedFieldsForType()
	c.False(eqp.PickSeparately, "a physical container is picked as a unit")
	group := NewEquipmentGroup(nil, nil)
	group.PickSeparately = true
	group.ClearUnusedFieldsForType()
	c.True(group.PickSeparately)
	group.ConvertToPhysicalContainer()
	c.False(group.PickSeparately, "nor once converted to one")
}

func TestPickSeparatelyHasEffect(t *testing.T) {
	c := check.New(t)
	choice, outer, inner := newOrganizedChoice()
	c.False(CanPickSeparately(choice))
	c.False(CanPickSeparately(choice.Children[0]))
	c.True(PickSeparatelyHasEffect(outer))
	c.True(IsOrganizingGroup(outer))
	c.True(IsOrganizingGroup(inner), "through an organizing group")

	outer.PickSeparately = false
	c.True(PickSeparatelyHasEffect(outer))
	c.False(IsOrganizingGroup(outer))
	c.False(PickSeparatelyHasEffect(inner), "not inside a unit group")
	c.False(IsOrganizingGroup(inner))

	inner.SetParent(nil)
	c.False(PickSeparatelyHasEffect(inner), "not outside a choice")
	c.False(IsOrganizingGroup(inner))

	c.True(CanPickSeparately(NewSkill(nil, nil, true)))
	c.False(CanPickSeparately(NewSkillChoiceContainer(nil, nil)))
	c.True(CanPickSeparately(NewSpell(nil, nil, true)))
	c.True(CanPickSeparately(NewEquipmentGroup(nil, nil)))
	c.False(CanPickSeparately(NewEquipment(nil, nil, true)))
	c.False(CanPickSeparately(NewNote(nil, nil, true)))
	var none *Trait
	c.False(IsOrganizingGroup(none))
	c.Equal(eqcontainer.Group, NewEquipmentChoiceContainer(nil, nil).ContainerType)
	c.False(CanPickSeparately(NewEquipmentChoiceContainer(nil, nil)))
}

func TestTemplateChoiceOptions(t *testing.T) {
	c := check.New(t)
	choice, outer, _ := newOrganizedChoice()
	c.Equal([]string{"a", "b", "c", "d", "e"}, traitNames(TemplateChoiceOptions(choice)))
	c.Equal([]string{"a", "outer", "e"}, traitNames(choice.Children), "the children are left alone")
	c.Equal([]string{"b", "inner"}, traitNames(TemplateChoiceOptions(outer)), "only a choice offers options")

	outer.PickSeparately = false
	options := TemplateChoiceOptions(choice)
	c.Equal([]string{"a", "outer", "e"}, traitNames(options))
	c.True(&options[0] == &choice.Children[0], "the children themselves come back")
	c.Equal(0.0, testing.AllocsPerRun(10, func() { TemplateChoiceOptions(choice) }))
}

func traitNames(traits []*Trait) []string {
	names := make([]string, len(traits))
	for i, one := range traits {
		names[i] = one.Name
	}
	return names
}

func newCostedTrait(points int) *Trait {
	trait := NewTrait(nil, nil, false)
	trait.BasePoints = fxp.FromInteger(points)
	return trait
}

func addOptions[T Node[T]](choice T, options ...T) T {
	choice.SetChildren(append(choice.NodeChildren(), options...))
	for _, one := range options {
		one.SetParent(choice)
	}
	return choice
}

func TestOrganizedChoiceCosts(t *testing.T) {
	c := check.New(t)
	a := newTraitGroup("A", true, newCostedTrait(10), newCostedTrait(20))
	b := newTraitGroup("B", true, newCostedTrait(5), newTraitGroup("C", true, newCostedTrait(40)))
	choice := addOptions(newPickerContainer(picker.Count, criteria.EqualsNumber, 1, []int{8}), a, b)
	c.Equal("5~40", choice.PointsRange(nil).String(), "options count one by one across the groups")
	a.PickSeparately, b.PickSeparately = false, false
	c.Equal("8~45", choice.PointsRange(nil).String(), "a unit group is one option")

	var data CellData
	a.CellData(TraitPointsColumn, &data)
	c.Equal("30", data.Primary)
	a.PickSeparately = true
	data = CellData{}
	a.CellData(TraitPointsColumn, &data)
	c.Equal("", data.Primary, "an organizing group shows no cost")
	a.SetParent(nil)
	data = CellData{}
	a.CellData(TraitPointsColumn, &data)
	c.Equal("30", data.Primary, "the flag means nothing outside a choice")
	c.Equal("30", a.PointsRange(nil).String())

	unit := newTraitGroup("Unit", false, newTraitGroup("D", true, newCostedTrait(10), newCostedTrait(20)))
	choice = addOptions(newPickerContainer(picker.Count, criteria.EqualsNumber, 1, []int{5}), unit)
	c.Equal("5~30", choice.PointsRange(nil).String(), "nor inside a unit group")

	choice = addOptions(newPickerContainer(picker.Points, criteria.AtLeastNumber, 20, nil),
		newTraitGroup("E", true, newTraitGroup("F", true, newCostedTrait(-10)), newCostedTrait(15)))
	c.Equal(newPickerRange(picker.Points, criteria.AtLeastNumber, 20, []int{-10, 15}), choice.PointsRange(nil),
		"a points choice sees the options in nested groups")

	enhanced := newTraitGroup("G", true, newCostedTrait(10), newCostedTrait(20))
	mod := NewTraitModifier(nil, nil, false)
	mod.CostAdj = "+100%"
	enhanced.AddModifiers(mod)
	choice = addOptions(newPickerContainer(picker.Count, criteria.EqualsNumber, 1, []int{5}), enhanced)
	c.Equal("5~40", choice.PointsRange(nil).String(), "the options cost with the group's modifiers")
}

func TestOrganizedChoiceCostsForOtherTypes(t *testing.T) {
	c := check.New(t)
	skills := NewSkillChoiceContainer(nil, nil)
	group := NewSkill(nil, nil, true)
	group.PickSeparately = true
	for _, points := range []int{1, 2} {
		skill := NewSkill(nil, nil, false)
		skill.Points = fxp.FromInteger(points)
		addOptions(group, skill)
	}
	four := NewSkill(nil, nil, false)
	four.Points = fxp.Four
	addOptions(skills, group, four)
	c.Equal("1~4", skills.RawPointsRange().String())
	c.Equal("1~4", skills.PointsRange(nil).String())

	spells := NewSpellChoiceContainer(nil, nil)
	spellGroup := NewSpell(nil, nil, true)
	spellGroup.PickSeparately = true
	for _, points := range []int{1, 2} {
		spell := NewSpell(nil, nil, false)
		spell.Points = fxp.FromInteger(points)
		addOptions(spellGroup, spell)
	}
	fourSpell := NewSpell(nil, nil, false)
	fourSpell.Points = fxp.Four
	addOptions(spells, spellGroup, fourSpell)
	c.Equal("1~4", spells.RawPointsRange().String())
	c.Equal("1~4", spells.PointsRange(nil).String())

	eqp := NewEquipmentChoiceContainer(nil, nil)
	eqp.TemplatePicker.Type = picker.Value
	eqp.TemplatePicker.Qualifier.Compare = criteria.AtMostNumber
	eqp.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(50)
	kit := NewEquipmentGroup(nil, nil)
	addOptions(eqp, addOptions(kit, newEquipmentItem("Rope", "30", "1 lb")))
	c.Equal("0~30", eqp.ExtendedValueRange().String(), "a unit group can't be raised")
	kit.PickSeparately = true
	c.Equal("0~50", eqp.ExtendedValueRange().String(), "what it holds can")

	eqp.TemplatePicker.Type = picker.Weight
	eqp.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(5)
	kit.PickSeparately = false
	c.Equal("0~1", eqp.ExtendedWeightRange(fxp.Pound).String(), "by weight, a unit group can't be raised")
	kit.PickSeparately = true
	c.Equal("0~5", eqp.ExtendedWeightRange(fxp.Pound).String(), "what it holds can")
}

func TestOrganizingGroupModifierChoiceIsMadeOnce(t *testing.T) {
	c := check.New(t)
	group := newTraitContainerWithChoice(container.Group, []string{"-50%", "+50%"}, 10, -10)
	group.PickSeparately = true
	both := addOptions(newPickerContainer(picker.Count, criteria.EqualsNumber, 2, nil), group)
	c.Equal("0", both.PointsRange(nil).String(), "either pick makes the two cancel out")
	c.Equal(fxp.Int(0), both.AdjustedPoints(nil))

	group = newTraitContainerWithChoice(container.Group, []string{"-50%", "+50%"}, 10, -10)
	group.PickSeparately = true
	one := addOptions(newPickerContainer(picker.Count, criteria.EqualsNumber, 1, nil), group)
	c.Equal("-15~15", one.PointsRange(nil).String())
	c.Equal(fxp.Int(0), one.AdjustedPoints(nil), "the least of the ways the group's choice is made")

	nested := newTraitContainerWithChoice(container.Group, []string{"-50%", "+50%"}, 10, -10)
	nested.PickSeparately = true
	outer := newTraitGroup("outer", true, nested)
	c.Equal("0", addOptions(newPickerContainer(picker.Count, criteria.EqualsNumber, 2, nil), outer).PointsRange(nil).String(),
		"a group nested in another")
}
