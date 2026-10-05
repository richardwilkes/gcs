// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// equipmentFilterLookup looks the fields up among the equipment list's, as the filter editor does.
func equipmentFilterLookup(key string) (title string, kind gurps.FilterFieldKind, plural, ok bool) {
	if field := findFilterField(gurps.EquipmentFilterFields(), key); field != nil {
		return field.Title, field.Kind, field.Plural, true
	}
	return "", 0, false, false
}

// bracket marks what Describe emphasizes.
func bracket(s string) string {
	return "[" + s + "]"
}

// negated returns the condition with its Not set.
func negated(cond *gurps.FilterCondition) *gurps.FilterCondition {
	cond.Not = true
	return cond
}

// describeCase is a condition and what Describe should say of it.
type describeCase struct {
	cond *gurps.FilterCondition
	want string
}

func TestFilterConditionDescribe(t *testing.T) {
	c := check.New(t)
	for i, one := range []describeCase{
		{newTextFilterCondition("name", criteria.ContainsText, "sword"), `Must have a name that contains "[sword]"`},
		{newTextFilterCondition("name", criteria.IsText, "Axe"), `Must have a name that is [Axe]`},
		{newTextFilterCondition("name", criteria.IsText, ""), `Must have a name that is ""`},
		{newTextFilterCondition("name", criteria.IsText, "   "), `Must have a name that is ""`},
		{newTextFilterCondition("name", criteria.IsText, " Axe"), `Must have a name that is "[ Axe]"`},
		{newTextFilterCondition("name", criteria.IsText, "Axe, Hand"), `Must have a name that is "[Axe, Hand]"`},
		{
			newTextFilterCondition("name", criteria.ContainsText, "Sword, Axe"),
			`Must have a name that contains "[Sword, Axe]"`,
		},
		{newTextFilterCondition("name", criteria.AnyText, ""), `Must have a name that is anything`},
		{
			newTextFilterCondition("name", criteria.StringComparison(200), "x"),
			`Must have a name that is anything`,
		},
		{
			negated(newTextFilterCondition("notes", criteria.StartsWithText, "Cheap")),
			`Must not have notes that start with "[Cheap]"`,
		},
		{newTextFilterCondition("tags", criteria.IsText, "Shield"), `Must have tags where at least one is [Shield]`},
		{
			newTextFilterCondition("tags", criteria.DoesNotContainText, "Melee"),
			`Must have tags where none contains "[Melee]"`,
		},
		{
			negated(newTextFilterCondition("tags", criteria.IsNotText, "Shield")),
			`Must not have tags where none is [Shield]`,
		},
		{newTextFilterCondition("tags", criteria.ContainsText, ""), `Must have tags where at least one contains ""`},
		{
			newTextFilterCondition("tags", criteria.IsText, "Sword, Axe"),
			`Must have tags where at least one is [Sword] or [Axe]`,
		},
		{
			newTextFilterCondition("tags", criteria.ContainsText, "Sword, Axe, ,Bow"),
			`Must have tags where at least one contains "[Sword]", "[Axe]" or "[Bow]"`,
		},
		{
			negated(newTextFilterCondition("tags", criteria.StartsWithText, "Mel")),
			`Must not have tags where at least one starts with "[Mel]"`,
		},
		{
			newTextFilterCondition("tags", criteria.IsNotText, "Sword, Axe"),
			`Must have tags where none is [Sword] or [Axe]`,
		},
		{
			negated(newTextFilterCondition("tags", criteria.DoesNotContainText, "Sword,Axe")),
			`Must not have tags where none contains "[Sword]" or "[Axe]"`,
		},
		{newTextFilterCondition("tags", criteria.AnyText, ""), `Must have tags that are anything`},
		{newTextFilterCondition("tags", criteria.IsText, ""), `Must not have tags`},
		{newTextFilterCondition("tags", criteria.IsText, " , "), `Must not have tags`},
		{negated(newTextFilterCondition("tags", criteria.IsText, "")), `Must have tags`},
		{
			newNumberFilterCondition("cost", criteria.AtLeastNumber, fxp.FromInteger(1000)),
			`Must have a cost that is at least [1,000]`,
		},
		{newNumberFilterCondition("cost", criteria.AnyNumber, 0), `Must have a cost that is anything`},
		{
			negated(newNumberFilterCondition("cost", criteria.NotEqualsNumber, fxp.FromInteger(10))),
			`Must not have a cost that is not [10]`,
		},
		{
			newWeightFilterCondition("weight", criteria.AtMostNumber, fxp.WeightFromInteger(5, fxp.Pound)),
			`Must have a weight that is at most [5 lb]`,
		},
		{
			negated(newWeightFilterCondition("weight", criteria.AtLeastNumber, fxp.WeightFromInteger(2, fxp.Pound))),
			`Must not have a weight that is at least [2 lb]`,
		},
		{gurps.NewFilterCondition(nil, "container"), `Must be a container`},
		{negated(gurps.NewFilterCondition(nil, "container")), `Must not be a container`},
		{
			gurps.NewFilterCondition(nil, "future_field"),
			`Condition on unknown field "future_field"; it will be preserved, but never matches`,
		},
		{
			negated(gurps.NewFilterCondition(nil, "future_field")),
			`Condition on unknown field "future_field"; it will be preserved, but never matches`,
		},
	} {
		c.Equal(one.want, one.cond.Describe(equipmentFilterLookup, fxp.Pound, bracket), "case %d", i)
	}
}

func TestFilterConditionDescribesWeightInUnits(t *testing.T) {
	c := check.New(t)
	cond := newWeightFilterCondition("weight", criteria.EqualsNumber, fxp.WeightFromInteger(2, fxp.Kilogram))
	c.Equal("Must have a weight that is 2 kg", cond.Describe(equipmentFilterLookup, fxp.Kilogram,
		func(s string) string { return s }))
}

// TestFilterConditionDescribesPluralFields checks that what follows a plural title agrees with it, for each kind of
// field that has a comparison: notes and points are plural in the lists that have them, and no list type has a plural
// weight, so one stands in for it.
func TestFilterConditionDescribesPluralFields(t *testing.T) {
	c := check.New(t)
	lookup := func(key string) (title string, kind gurps.FilterFieldKind, plural, ok bool) {
		switch key {
		case "weights":
			return "have weights", gurps.FilterFieldWeight, true, true
		default:
			if field := findFilterField(gurps.TraitFilterFields(), key); field != nil {
				return field.Title, field.Kind, field.Plural, true
			}
			return "", 0, false, false
		}
	}
	for i, one := range []describeCase{
		{newTextFilterCondition("notes", criteria.ContainsText, "cheap"), `Must have notes that contain "[cheap]"`},
		{newTextFilterCondition("notes", criteria.IsText, "Rare"), `Must have notes that are [Rare]`},
		{newTextFilterCondition("notes", criteria.IsText, "Rare "), `Must have notes that are "[Rare ]"`},
		{newTextFilterCondition("notes", criteria.AnyText, ""), `Must have notes that are anything`},
		{
			negated(newTextFilterCondition("notes", criteria.DoesNotEndWithText, "x")),
			`Must not have notes that do not end with "[x]"`,
		},
		{
			newNumberFilterCondition("points", criteria.AtLeastNumber, fxp.FromInteger(5)),
			`Must have points that are at least [5]`,
		},
		{newNumberFilterCondition("levels", criteria.AnyNumber, 0), `Must have levels that are anything`},
		{
			negated(newWeightFilterCondition("weights", criteria.AtMostNumber, fxp.WeightFromInteger(3, fxp.Pound))),
			`Must not have weights that are at most [3 lb]`,
		},
		{newWeightFilterCondition("weights", criteria.AnyNumber, 0), `Must have weights that are anything`},
		{newTextFilterCondition("tags", criteria.IsText, "Rare"), `Must have tags where at least one is [Rare]`},
	} {
		c.Equal(one.want, one.cond.Describe(lookup, fxp.Pound, bracket), "case %d", i)
	}
}

func TestFilterFieldPluralTitles(t *testing.T) {
	c := check.New(t)
	plural := map[string]bool{
		"have notes": true, "have points": true, "have levels": true, "have tags": true,
		"have colleges": true,
	}
	verify := func(title string, isPlural bool) {
		c.Equal(plural[title], isPlural, "%q is marked plural only if it is", title)
	}
	for _, f := range gurps.TraitFilterFields() {
		verify(f.Title, f.Plural)
	}
	for _, f := range gurps.TraitModifierFilterFields() {
		verify(f.Title, f.Plural)
	}
	for _, f := range gurps.SkillFilterFields() {
		verify(f.Title, f.Plural)
	}
	for _, f := range gurps.SpellFilterFields() {
		verify(f.Title, f.Plural)
	}
	for _, f := range gurps.EquipmentFilterFields() {
		verify(f.Title, f.Plural)
	}
	for _, f := range gurps.EquipmentModifierFilterFields() {
		verify(f.Title, f.Plural)
	}
	for _, f := range gurps.NoteFilterFields() {
		verify(f.Title, f.Plural)
	}
}

func TestUnknownFilterNodeDescribe(t *testing.T) {
	c := check.New(t)
	node := gurps.NewUnknownFilterNode("sparkle", []byte(`{"type":"sparkle"}`))
	c.Equal(`Unknown filter node type "sparkle"; it will be preserved, but never matches`, node.Describe())
}

// TestFilterTextFieldKeepsCommas checks that a comma in the qualifier of a single text field is part of the text, as
// its sentence quotes it, rather than splitting it into several values as a list's does.
func TestFilterTextFieldKeepsCommas(t *testing.T) {
	c := check.New(t)
	f := gurps.NewListFilter("Commas")
	f.Root.Children = gurps.FilterNodes{newTextFilterCondition("name", criteria.IsText, "Axe, Hand")}
	f.EnsureValidity()
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Axe, Hand"
	c.True(matchesListFilter(f, gurps.TraitFilterFields(), trait), "the whole text matches")
	trait.Name = "Axe"
	c.False(matchesListFilter(f, gurps.TraitFilterFields(), trait), "a part of it doesn't")
}
