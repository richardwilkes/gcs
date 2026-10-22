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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// newShortTraitModifier returns an enabled trait modifier with the given name and short name.
func newShortTraitModifier(e *Entity, name, shortName string) *TraitModifier {
	mod := NewTraitModifier(e, nil, false)
	mod.Name = name
	mod.ShortName = shortName
	return mod
}

func TestModifierShortNamesInOwnerNotes(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	allies := addTraitWithFeatures(e, "Allies")
	allies.Replacements = map[string]string{"Pct": "25%"}
	pointTotal := newShortTraitModifier(e, "25% of your starting points", "Built on @Pct@")
	blank := newShortTraitModifier(e, "Summonable", "  ")
	unwilling := newShortTraitModifier(e, "Unwilling", "")
	divisor := newShortTraitModifier(e, "Armor Divisor", "AD")
	divisor.Levels = fxp.Two
	allies.SetModifiers([]*TraitModifier{pointTotal, blank, unwilling, divisor})
	c.True(divisor.IsLeveled(), "precondition: the armor divisor modifier is leveled")

	c.Equal("Built on 25%; Summonable; Unwilling; AD 2", allies.ModifierNotes(),
		"the notes name each modifier by its short name, or its name when it has none, followed by any level")
	c.Equal("Armor Divisor 2", divisor.String(), "String keeps the full name")
	c.Equal("25% of your starting points", pointTotal.FullDescription(), "the full description keeps the full name")

	sword := addCarriedEquipmentWithFeatures(e, "Broadsword")
	fine := NewEquipmentModifier(e, nil, false)
	fine.Name = "Fine Quality"
	fine.ShortName = "Fine"
	sword.SetModifiers([]*EquipmentModifier{fine})
	c.Equal("Fine", sword.ModifierNotes(), "equipment modifiers are named by their short name too")
}

func TestModifierShortNamesInTitleNotes(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	attack := addTraitWithFeatures(e, "Extra Attack")
	multiple := newShortTraitModifier(e, "Multiple Attacks", "Multi")
	multiple.Levels = fxp.Two
	multiple.ShowInTitle = true
	plain := newShortTraitModifier(e, "Single Shot", "")
	plain.ShowInTitle = true
	attack.SetModifiers([]*TraitModifier{multiple, plain})
	c.True(multiple.IsLeveled(), "precondition: the multiple attacks modifier is leveled")

	c.Equal("Extra Attack (Multi 2; Single Shot)", attack.String(),
		"the title names each modifier by its short name, or its name when it has none, followed by any level")
	c.Equal("", attack.ModifierNotes(), "modifiers shown in the title stay out of the notes")

	sword := addCarriedEquipmentWithFeatures(e, "Broadsword")
	fine := NewEquipmentModifier(e, nil, false)
	fine.Name = "Fine Quality"
	fine.ShortName = "Fine"
	fine.ShowInTitle = true
	sword.SetModifiers([]*EquipmentModifier{fine})
	c.Equal("Broadsword (Fine)", sword.String(), "equipment modifiers are named by their short name in the title too")
}

func TestModifierShortNameRoundTripAndHash(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	plain := NewTraitModifier(e, nil, false)
	plain.Name = "Appears quite often (12-)"
	short := NewTraitModifier(e, nil, false)
	short.Name = plain.Name
	short.TID = plain.TID
	short.ShortName = "12 or less"
	c.NotEqual(Hash64(plain), Hash64(short), "the short name is source data and is hashed")

	data, err := jio.Marshal(short)
	c.NoError(err)
	c.Contains(string(data), `"short_name":"12 or less"`)
	var restored TraitModifier
	c.NoError(jio.Unmarshal(data, &restored))
	c.Equal("12 or less", restored.ShortName)
	data, err = jio.Marshal(plain)
	c.NoError(err)
	c.NotContains(string(data), "short_name")
}

func TestPrereqModifierNameMatchesShortName(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	allies := addTraitWithFeatures(e, "Allies")
	frequency := newShortTraitModifier(e, "Appears quite often (12-)", "12 or less")
	unwilling := newShortTraitModifier(e, "Unwilling", "")
	allies.SetModifiers([]*TraitModifier{frequency})
	c.True(frequency.Enabled(), "precondition: the frequency modifier is enabled")

	for _, one := range []struct {
		compare   criteria.StringComparison
		qualifier string
		expected  bool
	}{
		{criteria.IsText, "Appears quite often (12-)", true},
		{criteria.IsText, "12 or less", true},
		{criteria.ContainsText, "quite often", true},
		{criteria.ContainsText, "or less", true},
		{criteria.IsText, "Unwilling", false},
		{criteria.IsNotText, "Unwilling", true},
		{criteria.IsNotText, "12 or less", false},
		{criteria.IsNotText, "Appears quite often (12-)", false},
	} {
		p := NewTraitPrereq()
		p.NameCriteria.Qualifier = "Allies"
		p.LevelCriteria.Compare = criteria.AnyNumber
		p.ModifierNameCriteria.Compare = one.compare
		p.ModifierNameCriteria.Qualifier = one.qualifier
		c.Equal(one.expected, p.Satisfied(e, nil, nil, "", nil), "%s %q", one.compare.Key(), one.qualifier)
	}

	allies.SetModifiers([]*TraitModifier{frequency, unwilling})
	both := NewTraitPrereq()
	both.NameCriteria.Qualifier = "Allies"
	both.LevelCriteria.Compare = criteria.AnyNumber
	both.ModifierNameCriteria.Compare = criteria.IsNotText
	both.ModifierNameCriteria.Qualifier = "12 or less"
	c.True(both.Satisfied(e, nil, nil, "", nil), "another modifier whose names both differ meets a not check")

	sword := addCarriedEquipmentWithFeatures(e, "Broadsword")
	fine := NewEquipmentModifier(e, nil, false)
	fine.Name = "Fine Quality"
	fine.ShortName = "Fine"
	sword.SetModifiers([]*EquipmentModifier{fine})
	c.True(sword.ReallyEquipped(), "precondition: the sword is equipped")
	c.True(fine.Enabled(), "precondition: the fine modifier is enabled")
	eqpPrereq := NewEquippedEquipmentPrereq()
	eqpPrereq.NameCriteria.Qualifier = "Broadsword"
	eqpPrereq.ModifierNameCriteria.Compare = criteria.IsText
	eqpPrereq.ModifierNameCriteria.Qualifier = "Fine"
	c.True(eqpPrereq.Satisfied(e, nil, nil, "", nil), "equipment modifier matched by its short name")
	eqpPrereq.ModifierNameCriteria.Compare = criteria.IsNotText
	c.False(eqpPrereq.Satisfied(e, nil, nil, "", nil), "a not check must hold for both equipment modifier names")
}

func TestActiveModifierForMatchesShortName(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	allies := addTraitWithFeatures(e, "Allies")
	frequency := newShortTraitModifier(e, "Appears quite often (12-)", "12 or less")
	unwilling := newShortTraitModifier(e, "Unwilling", "")
	allies.SetModifiers([]*TraitModifier{frequency, unwilling})
	c.True(frequency.Enabled() && unwilling.Enabled(), "precondition: both modifiers are enabled")
	c.True(allies.ActiveModifierFor("Appears quite often (12-)") == frequency, "found by its full name")
	c.True(allies.ActiveModifierFor("12 OR LESS") == frequency, "found by its short name, ignoring case")
	c.True(allies.ActiveModifierFor("Unwilling") == unwilling, "a modifier without a short name is found by name")
	c.Nil(allies.ActiveModifierFor(""), "an empty name doesn't match an empty short name")

	sword := addCarriedEquipmentWithFeatures(e, "Broadsword")
	fine := NewEquipmentModifier(e, nil, false)
	fine.Name = "Fine Quality"
	fine.ShortName = "Fine"
	sword.SetModifiers([]*EquipmentModifier{fine})
	c.True(fine.Enabled(), "precondition: the fine modifier is enabled")
	c.True(sword.ActiveModifierFor("fine") == fine, "equipment modifiers are found by short name too")
}

func TestLinkedModifierCodeInShortName(t *testing.T) {
	c := check.New(t)
	trait, mod := linkedTrait()
	mod.ShortName = "@focus@"
	trait.Replacements = map[string]string{"style": "Judo"}
	c.Equal("Judo", mod.ShortNameWithReplacements(), "a linked code in the short name takes the owner's answer")

	m := make(map[string]string)
	trait.FillWithNameableKeys(m, nil)
	_, asked := m["focus"]
	c.False(asked, "the linked code in the short name is asked as the trait's")
}
