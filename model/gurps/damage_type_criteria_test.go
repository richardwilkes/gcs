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
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// damageTypeSelectors holds a damage bonus, a skill bonus and an armor divisor override that select weapons as the
// selection says, each narrowed by the same damage type criterion. A selection by required skill has no skill bonus,
// since a skill bonus can't select weapons that way.
type damageTypeSelectors struct {
	damage   *WeaponBonus
	skill    *SkillBonus
	override *SelectorOverride
}

// newDamageTypeSelectors returns selectors selecting the weapons named, or using the skill named, name, each narrowed
// to the damage types the criterion matches.
func newDamageTypeSelectors(selection wsel.Type, name string, damageType criteria.Text) *damageTypeSelectors {
	s := &damageTypeSelectors{
		damage:   NewWeaponBonus(feature.WeaponBonus),
		override: NewSelectorOverride(selector.WeaponArmorDivisor),
	}
	s.damage.SelectionType = selection
	s.damage.NameCriteria.Qualifier = name
	s.damage.DamageTypeCriteria = damageType
	s.override.SelectionType = selection
	s.override.Value = "2"
	s.override.NameCriteria.Qualifier = name
	s.override.DamageTypeCriteria = damageType
	switch selection {
	case wsel.ThisWeapon, wsel.WithName:
		s.skill = NewSkillBonus()
		s.skill.SelectionType = skillsel.ThisWeapon
		if selection == wsel.WithName {
			s.skill.SelectionType = skillsel.WeaponsWithName
		}
		s.skill.NameCriteria.Qualifier = name
		s.skill.DamageTypeCriteria = damageType
		s.skill.Amount = fxp.Two
	default:
	}
	return s
}

// features returns the selectors as the features of an item.
func (s *damageTypeSelectors) features() Features {
	features := Features{s.damage, s.override}
	if s.skill != nil {
		features = append(features, s.skill)
	}
	return features
}

// check verifies that each selector selects the weapon when want is true, and none do otherwise. skillLevel is the
// level of the skill the weapon is used with, which the skill bonus raises by 2.
func (s *damageTypeSelectors) check(c check.Checker, w *Weapon, skillLevel fxp.Int, want bool, msg string) {
	c.Helper()
	bonuses := w.collectWeaponBonuses(oneDieCount, nil, feature.WeaponBonus)
	c.Equal(want, slices.Contains(bonuses, s.damage), "damage bonus: "+msg)
	if s.skill != nil {
		expected := skillLevel
		if want {
			expected += fxp.Two
		}
		c.Equal(expected, w.SkillLevel(nil), "skill bonus: "+msg)
	}
	expected := "1"
	if want {
		expected = "2"
	}
	c.Equal(expected, w.ResolveSelector(selector.WeaponArmorDivisor, "1", nil), "selector override: "+msg)
}

// newDamageTypeTestWeapon returns a melee weapon of the owner used with the Broadsword skill that does the given type
// of damage. It is not added to the owner.
func newDamageTypeTestWeapon(owner WeaponOwner, damageType string) *Weapon {
	w := newCrushingWeapon(owner, "Swung")
	w.Damage.Type = damageType
	w.Defaults = []*SkillDefault{newSkillDefaultTo("Broadsword", "", true, 0)}
	return w
}

// TestWeaponDamageTypeCriteria verifies that the damage type criterion of a weapon bonus, a skill bonus and a selector
// override narrows the weapons each selects, in every selection, to those whose damage type matches any of the
// comma-separated types it names.
func TestWeaponDamageTypeCriteria(t *testing.T) {
	c := check.New(t)
	for _, selection := range []wsel.Type{wsel.ThisWeapon, wsel.WithName, wsel.WithRequiredSkill} {
		for _, one := range []struct {
			name     string
			criteria criteria.Text
			wantCut  bool
			wantCr   bool
		}{
			{"is anything", criteria.Text{Compare: criteria.AnyText}, true, true},
			{"is cut", criteria.Text{Compare: criteria.IsText, Qualifier: "cut"}, true, false},
			{"is imp or cut", criteria.Text{Compare: criteria.IsText, Qualifier: "imp, cut"}, true, false},
			{"is imp", criteria.Text{Compare: criteria.IsText, Qualifier: "imp"}, false, false},
			{"is not cut", criteria.Text{Compare: criteria.IsNotText, Qualifier: "cut"}, false, true},
			{"starts with c", criteria.Text{Compare: criteria.StartsWithText, Qualifier: "c"}, true, true},
		} {
			msg := selection.String() + ", " + one.name
			e := NewEntity()
			sk := addAverageDXSkill(e, "Broadsword", fxp.Four)
			name := "Gadget"
			if selection == wsel.WithRequiredSkill {
				name = "Broadsword"
			}
			selectors := newDamageTypeSelectors(selection, name, one.criteria)
			gadget := addTraitWithFeatures(e, "Gadget")
			if selection == wsel.ThisWeapon {
				gadget.Features = selectors.features()
			} else {
				addTraitWithFeatures(e, "Rules", selectors.features()...)
			}
			cut := newDamageTypeTestWeapon(gadget, "cut")
			cr := newDamageTypeTestWeapon(gadget, "cr")
			gadget.Weapons = []*Weapon{cut, cr}
			e.Recalculate()
			selectors.check(c, cut, sk.LevelData.Level, one.wantCut, msg+": cutting weapon")
			selectors.check(c, cr, sk.LevelData.Level, one.wantCr, msg+": crushing weapon")
		}
	}
}

// TestWeaponDamageTypeCriteriaMatchesResolvedType verifies that a weapon bonus, a skill bonus and a selector override
// of another field match the damage type as the overrides resolve it, while an override of the damage type itself
// matches the weapon's own, since resolving the type consults it.
func TestWeaponDamageTypeCriteriaMatchesResolvedType(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sk := addAverageDXSkill(e, "Broadsword", fxp.Four)
	isCut := criteria.Text{Compare: criteria.IsText, Qualifier: "cut"}
	selectors := newDamageTypeSelectors(wsel.WithName, "Gadget", isCut)
	toCut := newDamageTypeOverride(wsel.WithName, "cut")
	toCut.NameCriteria.Qualifier = "Gadget"
	cutToImp := newDamageTypeOverride(wsel.WithName, "imp")
	cutToImp.NameCriteria.Qualifier = "Gadget"
	cutToImp.DamageTypeCriteria = isCut
	cutToImp.Priority = 1
	addTraitWithFeatures(e, "Rules", append(selectors.features(), toCut, cutToImp)...)
	gadget := addTraitWithFeatures(e, "Gadget")
	w := newDamageTypeTestWeapon(gadget, "cr")
	gadget.Weapons = []*Weapon{w}
	e.Recalculate()
	c.Equal("cut", resolvedDamageType(w, nil),
		"an override of the damage type matches the weapon's own type, so the one for cutting weapons is passed over")
	bonuses := w.collectWeaponBonuses(oneDieCount, nil, feature.WeaponBonus)
	c.True(slices.Contains(bonuses, selectors.damage), "a weapon bonus matches the resolved damage type")
	c.Equal(sk.LevelData.Level+fxp.Two, w.SkillLevel(nil), "as does a skill bonus")
	c.Equal("2", w.ResolveSelector(selector.WeaponArmorDivisor, "1", nil), "as does an override of another field")
}

// TestWeaponDamageTypeCriteriaJSON verifies that the damage type criterion is stored under "damage_type", left out
// while it is anything, and survives a round-trip.
func TestWeaponDamageTypeCriteriaJSON(t *testing.T) {
	c := check.New(t)
	selectors := newDamageTypeSelectors(wsel.WithName, "Gadget", criteria.Text{Compare: criteria.AnyText})
	features := selectors.features()
	stored := func() []map[string]any {
		data, err := jio.Marshal(features)
		c.NoError(err)
		var list []map[string]any
		c.NoError(jio.Unmarshal(data, &list))
		return list
	}
	for _, one := range stored() {
		_, exists := one["damage_type"]
		c.False(exists, "%v leaves out a damage type that is anything", one["type"])
	}
	selectors.damage.DamageTypeCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "cut"}
	selectors.skill.DamageTypeCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "cut"}
	selectors.override.DamageTypeCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "cut"}
	checkFeaturesJSONRoundTrip(c, features)
	for _, one := range stored() {
		_, exists := one["damage_type"]
		c.True(exists, "%v stores a damage type that is set", one["type"])
	}
}
