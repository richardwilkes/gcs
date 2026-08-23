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
	"strings"
	"testing"

	"github.com/dop251/goja"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/equipmentsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/traitsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// newSwitchableAttributeBonuses returns a pair of ST bonuses to attach to the same owner: a +1 that always applies and
// a switchable +2 that only applies while the owner's switch is on. Every entity-level gating test uses this pair, so
// the ST bonus the entity reports says exactly which of the two were let through: 1 = off, 3 = on.
func newSwitchableAttributeBonuses() Features {
	always := NewAttributeBonus(StrengthID)
	always.Amount = fxp.One
	switched := NewAttributeBonus(StrengthID)
	switched.Amount = fxp.Two
	switched.Switchable = true
	return Features{always, switched}
}

// TestFeaturesActiveAndAnySwitchable covers the filtering primitive every gated call site is built on.
func TestFeaturesActiveAndAnySwitchable(t *testing.T) {
	c := check.New(t)

	// Nothing to filter.
	var nilFeatures Features
	c.False(nilFeatures.AnySwitchable(), "a nil list has nothing switchable")
	c.Equal(0, len(nilFeatures.Active(false)), "a nil list filters to nothing")
	c.Equal(0, len(nilFeatures.Active(true)), "a nil list filters to nothing")
	empty := Features{}
	c.False(empty.AnySwitchable(), "an empty list has nothing switchable")
	c.Equal(0, len(empty.Active(false)), "an empty list filters to nothing")

	// A list with nothing switchable is handed back as-is, without allocating a copy.
	plain := Features{NewAttributeBonus(StrengthID), NewSkillBonus()}
	c.False(plain.AnySwitchable(), "no feature is switchable")
	for _, on := range []bool{false, true} {
		active := plain.Active(on)
		c.Equal(len(plain), len(active), "switchedOn=%v: every feature is active", on)
		c.True(&plain[0] == &active[0], "switchedOn=%v: the receiver itself is returned", on)
	}

	// A list with something switchable is also handed back as-is while the switch is on.
	first := NewAttributeBonus(StrengthID)
	second := NewSkillBonus()
	second.Switchable = true
	third := NewSpellBonus()
	fourth := NewReactionBonus()
	fourth.Switchable = true
	mixed := Features{first, second, third, fourth}
	c.True(mixed.AnySwitchable(), "one of the features is switchable")
	active := mixed.Active(true)
	c.Equal(len(mixed), len(active), "everything is active while the switch is on")
	c.True(&mixed[0] == &active[0], "the receiver itself is returned while the switch is on")

	// While the switch is off, only the non-switchable features survive, in their original order.
	active = mixed.Active(false)
	c.Equal(2, len(active), "only the non-switchable features are active while the switch is off")
	c.True(Feature(first) == active[0], "the first non-switchable feature keeps its place")
	c.True(Feature(third) == active[1], "the second non-switchable feature keeps its place")
	c.Equal(4, len(mixed), "the receiver is left alone")
}

// TestEveryFeatureTypeCarriesTheSwitchableFlag is a guard over the whole feature catalog: every type the feature editor
// offers must be able to store, load, save and hash the switchable flag. A type that embedded FeatureSwitch but forgot
// it in its Hash would let a switchable feature and an always-on one collapse together, and a type that never embedded
// it at all would silently discard the user's choice on the next save.
func TestEveryFeatureTypeCarriesTheSwitchableFlag(t *testing.T) {
	c := check.New(t)
	for _, ft := range feature.SelectableTypes {
		key := ft.Key()
		var loaded Features
		c.NoError(jio.Unmarshal([]byte(`[{"type":"`+key+`","switchable":true}]`), &loaded), "%s: should load", key)
		c.Equal(1, len(loaded), "%s: a single feature should load", key)
		if len(loaded) != 1 {
			continue
		}
		_, unknown := loaded[0].(*UnknownFeature)
		c.False(unknown, "%s: should load as a known feature type, not an UnknownFeature", key)
		c.True(loaded[0].IsSwitchable(), "%s: the switchable flag should survive the load", key)

		data, err := jio.Marshal(loaded)
		c.NoError(err, "%s: should marshal", key)
		c.Contains(string(data), `"switchable":true`, "%s: the switchable flag should be written out", key)
		switchedHash := Hash64(loaded[0])

		// The same feature with the flag cleared must not write the key at all, and must hash differently.
		loaded[0].SetSwitchable(false)
		c.False(loaded[0].IsSwitchable(), "%s: the flag should clear", key)
		data, err = jio.Marshal(loaded)
		c.NoError(err, "%s: should marshal", key)
		c.NotContains(string(data), "switchable", "%s: a non-switchable feature should omit the key", key)
		c.NotEqual(switchedHash, Hash64(loaded[0]), "%s: the switchable flag should participate in the hash", key)

		// A feature loaded without the key is not switchable.
		var plain Features
		c.NoError(jio.Unmarshal([]byte(`[{"type":"`+key+`"}]`), &plain), "%s: should load without the key", key)
		if len(plain) == 1 {
			c.False(plain[0].IsSwitchable(), "%s: a feature without the key is not switchable", key)
		}
	}
}

// TestUnknownFeatureIgnoresButPreservesSwitchable verifies that a feature type this build doesn't recognize is neither
// treated as switchable (it has no effect to switch) nor stripped of a "switchable" flag written by a newer version of
// GCS.
func TestUnknownFeatureIgnoresButPreservesSwitchable(t *testing.T) {
	c := check.New(t)
	data, err := jio.Marshal([]map[string]any{{
		"type":       "future_bonus_from_a_newer_gcs",
		"switchable": true,
		"amount":     3,
	}})
	c.NoError(err, "could not prepare test data")
	var features Features
	c.NoError(jio.Unmarshal(data, &features), "an unknown feature type must still load")
	c.Equal(1, len(features), "the unknown feature should be present")
	unknown, ok := features[0].(*UnknownFeature)
	c.True(ok, "the feature should load as an UnknownFeature")
	c.False(unknown.IsSwitchable(), "an unknown feature is never switchable")

	// Setting the flag is a no-op, since the raw data is what gets written back out.
	unknown.SetSwitchable(true)
	c.False(unknown.IsSwitchable(), "setting the flag on an unknown feature changes nothing")

	out, err := jio.Marshal(features)
	c.NoError(err, "the unknown feature should marshal")
	c.Equal(string(data), string(out), "the unknown feature must be reproduced byte for byte, switchable flag included")

	// The switchable state of an unknown feature must also stay out of the way of the switch machinery.
	c.False(features.AnySwitchable(), "an unknown feature doesn't make a list switchable")
	c.Equal(1, len(features.Active(false)), "an unknown feature is never filtered out")
}

// TestFeatureHashIncludesSwitchable verifies that two otherwise identical features that differ only in their switchable
// state hash differently, so that editing the switch marks the document as changed.
func TestFeatureHashIncludesSwitchable(t *testing.T) {
	c := check.New(t)
	always := NewAttributeBonus(StrengthID)
	switched := NewAttributeBonus(StrengthID)
	c.Equal(Hash64(always), Hash64(switched), "identical bonuses hash the same")
	switched.Switchable = true
	c.NotEqual(Hash64(always), Hash64(switched), "bonuses differing only in the switchable flag hash differently")
}

// TestItemHashIgnoresSwitchedOn verifies that the on/off state of an item's switch stays out of its source-data hash.
// The state is a local choice made on the sheet, so flipping it must not make the item report as out of sync with the
// library it came from, nor may a sync overwrite it.
func TestItemHashIgnoresSwitchedOn(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	trait := NewTrait(e, nil, false)
	trait.Name = "Gadget"
	trait.Features = Features{NewAttributeBonus(StrengthID)}
	skill := NewSkill(e, nil, false)
	skill.Name = "Brawling"
	skill.Features = Features{NewAttributeBonus(StrengthID)}
	spell := NewSpell(e, nil, false)
	spell.Name = "Fireball"
	spell.Features = Features{NewAttributeBonus(StrengthID)}
	eqp := NewEquipment(e, nil, false)
	eqp.Name = "Amulet"
	eqp.Features = Features{NewAttributeBonus(StrengthID)}

	for _, one := range []struct {
		name     string
		hashable Hashable
		switcher FeatureSwitcher
	}{
		{name: "trait", hashable: trait, switcher: trait},
		{name: "skill", hashable: skill, switcher: skill},
		{name: "spell", hashable: spell, switcher: spell},
		{name: "equipment", hashable: eqp, switcher: eqp},
	} {
		before := Hash64(one.hashable)
		one.switcher.SetSwitchedOn(true)
		c.Equal(before, Hash64(one.hashable), "%s: flipping the switch must not alter the source-data hash", one.name)
		c.True(one.switcher.IsSwitchedOn(), "%s: the switch is on", one.name)
		one.switcher.SetSwitchedOn(false)
		c.False(one.switcher.IsSwitchedOn(), "%s: the switch is off again", one.name)
	}
}

// TestSwitchedOnJSONRoundTrip verifies that the switch state of each kind of item is persisted, and only when it is on.
func TestSwitchedOnJSONRoundTrip(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	for _, one := range []struct {
		name   string
		make   func() (any, FeatureSwitcher)
		reload func(data []byte) (FeatureSwitcher, error)
	}{
		{
			name: "trait",
			make: func() (any, FeatureSwitcher) {
				t := NewTrait(e, nil, false)
				t.Name = "Gadget"
				return t, t
			},
			reload: func(data []byte) (FeatureSwitcher, error) {
				var t Trait
				return &t, jio.Unmarshal(data, &t)
			},
		},
		{
			name: "skill",
			make: func() (any, FeatureSwitcher) {
				s := NewSkill(e, nil, false)
				s.Name = "Brawling"
				return s, s
			},
			reload: func(data []byte) (FeatureSwitcher, error) {
				var s Skill
				return &s, jio.Unmarshal(data, &s)
			},
		},
		{
			name: "spell",
			make: func() (any, FeatureSwitcher) {
				s := NewSpell(e, nil, false)
				s.Name = "Fireball"
				return s, s
			},
			reload: func(data []byte) (FeatureSwitcher, error) {
				var s Spell
				return &s, jio.Unmarshal(data, &s)
			},
		},
		{
			name: "equipment",
			make: func() (any, FeatureSwitcher) {
				eqp := NewEquipment(e, nil, false)
				eqp.Name = "Amulet"
				return eqp, eqp
			},
			reload: func(data []byte) (FeatureSwitcher, error) {
				var eqp Equipment
				return &eqp, jio.Unmarshal(data, &eqp)
			},
		},
	} {
		item, switcher := one.make()

		// A switch that is off is not written at all.
		data, err := jio.Marshal(item)
		c.NoError(err, "%s: should marshal", one.name)
		c.NotContains(string(data), "switched_on", "%s: an off switch is not written", one.name)
		restored, err := one.reload(data)
		c.NoError(err, "%s: should load", one.name)
		c.False(restored.IsSwitchedOn(), "%s: the loaded switch is off", one.name)

		// A switch that is on is written and comes back on.
		switcher.SetSwitchedOn(true)
		data, err = jio.Marshal(item)
		c.NoError(err, "%s: should marshal", one.name)
		c.Contains(string(data), `"switched_on":true`, "%s: an on switch is written", one.name)
		restored, err = one.reload(data)
		c.NoError(err, "%s: should load", one.name)
		c.True(restored.IsSwitchedOn(), "%s: the loaded switch is on", one.name)
	}
}

// TestHasSwitchableFeatures verifies which items report that they have something to switch, since that is what decides
// whether the switch cell appears in the tables at all.
func TestHasSwitchableFeatures(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	switchableBonus := func() *AttributeBonus {
		bonus := NewAttributeBonus(StrengthID)
		bonus.Switchable = true
		return bonus
	}

	// A non-container trait with a switchable feature of its own.
	trait := NewTrait(e, nil, false)
	c.False(trait.HasSwitchableFeatures(), "a trait with no features has nothing to switch")
	trait.Features = Features{NewAttributeBonus(StrengthID)}
	c.False(trait.HasSwitchableFeatures(), "a trait whose features are all always-on has nothing to switch")
	trait.Features = append(trait.Features, switchableBonus())
	c.True(trait.HasSwitchableFeatures(), "a trait with a switchable feature has something to switch")

	// A container trait's own features never apply, so they don't count -- but its modifiers' do.
	container := NewTrait(e, nil, true)
	container.Features = Features{switchableBonus()}
	c.False(container.HasSwitchableFeatures(), "a container's own features never apply, so they aren't switchable")
	mod := NewTraitModifier(e, nil, false)
	mod.Features = Features{switchableBonus()}
	container.Modifiers = []*TraitModifier{mod}
	c.True(container.HasSwitchableFeatures(), "an enabled modifier's switchable feature counts")
	mod.Disabled = true
	c.False(container.HasSwitchableFeatures(), "a disabled modifier's switchable feature doesn't count")
	mod.Disabled = false

	// A modifier with nothing switchable ahead of one that has something must not mask it.
	plainMod := NewTraitModifier(e, nil, false)
	plainMod.Features = Features{NewAttributeBonus(StrengthID)}
	container.Modifiers = []*TraitModifier{plainMod, mod}
	c.True(container.HasSwitchableFeatures(), "a switchable modifier later in the list is still found")

	// Skills and spells carry only their own features.
	skill := NewSkill(e, nil, false)
	c.False(skill.HasSwitchableFeatures(), "a skill with no features has nothing to switch")
	skill.Features = Features{switchableBonus()}
	c.True(skill.HasSwitchableFeatures(), "a skill with a switchable feature has something to switch")

	spell := NewSpell(e, nil, false)
	c.False(spell.HasSwitchableFeatures(), "a spell with no features has nothing to switch")
	spell.Features = Features{switchableBonus()}
	c.True(spell.HasSwitchableFeatures(), "a spell with a switchable feature has something to switch")

	// A skill or spell container's features never apply (see Entity.processFeatures), so even features that were left
	// on one -- ClearUnusedFieldsForType strips them, but a caller that sets them directly bypasses that -- give it
	// nothing to switch.
	skillContainer := NewSkill(e, nil, true)
	skillContainer.Features = Features{switchableBonus()}
	c.False(skillContainer.HasSwitchableFeatures(), "a skill container's features never apply, so they aren't switchable")
	spellContainer := NewSpell(e, nil, true)
	spellContainer.Features = Features{switchableBonus()}
	c.False(spellContainer.HasSwitchableFeatures(), "a spell container's features never apply, so they aren't switchable")

	// Equipment counts both its own features and those of its enabled modifiers.
	eqp := NewEquipment(e, nil, false)
	c.False(eqp.HasSwitchableFeatures(), "equipment with no features has nothing to switch")
	eqp.Features = Features{switchableBonus()}
	c.True(eqp.HasSwitchableFeatures(), "equipment with a switchable feature has something to switch")

	eqp = NewEquipment(e, nil, false)
	eqpMod := NewEquipmentModifier(e, nil, false)
	eqpMod.Features = Features{switchableBonus()}
	eqp.Modifiers = []*EquipmentModifier{eqpMod}
	c.True(eqp.HasSwitchableFeatures(), "an enabled equipment modifier's switchable feature counts")
	eqpMod.Disabled = true
	c.False(eqp.HasSwitchableFeatures(), "a disabled equipment modifier's switchable feature doesn't count")
}

// TestEntityGatesSwitchableFeatures is the core of the feature: for each place a feature can be attached, a switchable
// bonus must be ignored by the entity while the owning item's switch is off and applied while it is on, without
// disturbing the non-switchable bonus sitting beside it. Features carried by a modifier follow the switch of the trait
// or piece of equipment the modifier belongs to; nothing is set on the modifier itself.
func TestEntityGatesSwitchableFeatures(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name  string
		setup func(e *Entity, features Features) FeatureSwitcher
	}{
		{
			name: "trait",
			setup: func(e *Entity, features Features) FeatureSwitcher {
				trait := NewTrait(e, nil, false)
				trait.Name = "Gadget"
				trait.Features = features
				e.Traits = append(e.Traits, trait)
				return trait
			},
		},
		{
			name: "trait modifier",
			setup: func(e *Entity, features Features) FeatureSwitcher {
				trait := NewTrait(e, nil, false)
				trait.Name = "Gadget"
				mod := NewTraitModifier(e, nil, false)
				mod.Name = "Enhanced"
				mod.Features = features
				trait.Modifiers = []*TraitModifier{mod}
				e.Traits = append(e.Traits, trait)
				return trait
			},
		},
		{
			name: "skill",
			setup: func(e *Entity, features Features) FeatureSwitcher {
				skill := NewSkill(e, nil, false)
				skill.Name = "Brawling"
				skill.Points = fxp.Four
				skill.Features = features
				e.Skills = append(e.Skills, skill)
				return skill
			},
		},
		{
			name: "spell",
			setup: func(e *Entity, features Features) FeatureSwitcher {
				spell := NewSpell(e, nil, false)
				spell.Name = "Fireball"
				spell.Points = fxp.Four
				spell.Features = features
				e.Spells = append(e.Spells, spell)
				return spell
			},
		},
		{
			name: "carried equipment",
			setup: func(e *Entity, features Features) FeatureSwitcher {
				eqp := NewEquipment(e, nil, false)
				eqp.Name = "Amulet"
				eqp.Features = features
				e.CarriedEquipment = append(e.CarriedEquipment, eqp)
				return eqp
			},
		},
		{
			name: "equipment modifier",
			setup: func(e *Entity, features Features) FeatureSwitcher {
				eqp := NewEquipment(e, nil, false)
				eqp.Name = "Amulet"
				mod := NewEquipmentModifier(e, nil, false)
				mod.Name = "Blessed"
				mod.Features = features
				eqp.Modifiers = []*EquipmentModifier{mod}
				e.CarriedEquipment = append(e.CarriedEquipment, eqp)
				return eqp
			},
		},
	} {
		e := NewEntity()
		switcher := tc.setup(e, newSwitchableAttributeBonuses())
		c.True(switcher.HasSwitchableFeatures(), "%s: the item reports something to switch", tc.name)
		c.False(switcher.IsSwitchedOn(), "%s: a new item's switch starts off", tc.name)

		e.Recalculate()
		c.Equal(fxp.One, e.AttributeBonusFor(StrengthID, stlimit.None, nil),
			"%s: only the always-on bonus applies while the switch is off", tc.name)
		c.Equal(fxp.Eleven, e.Attributes.Current(StrengthID), "%s: ST reflects only the always-on bonus", tc.name)

		switcher.SetSwitchedOn(true)
		e.Recalculate()
		c.Equal(fxp.Three, e.AttributeBonusFor(StrengthID, stlimit.None, nil),
			"%s: both bonuses apply while the switch is on", tc.name)
		c.Equal(fxp.Thirteen, e.Attributes.Current(StrengthID), "%s: ST reflects both bonuses", tc.name)

		switcher.SetSwitchedOn(false)
		e.Recalculate()
		c.Equal(fxp.One, e.AttributeBonusFor(StrengthID, stlimit.None, nil),
			"%s: switching back off drops the switchable bonus again", tc.name)
	}
}

// TestSwitchableSkillBonus verifies the gating with a bonus whose effect is visible on another node rather than on an
// attribute, confirming the switch is applied where the features are collected rather than where they are consumed.
func TestSwitchableSkillBonus(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	skill := NewSkill(e, nil, false)
	skill.Name = "Brawling"
	skill.Points = fxp.Four
	e.Skills = append(e.Skills, skill)

	always := NewSkillBonus()
	always.NameCriteria.Qualifier = "Brawling"
	always.Amount = fxp.One
	switched := NewSkillBonus()
	switched.NameCriteria.Qualifier = "Brawling"
	switched.Amount = fxp.Two
	switched.Switchable = true

	trait := NewTrait(e, nil, false)
	trait.Name = "Gadget"
	trait.Features = Features{always, switched}
	e.Traits = append(e.Traits, trait)
	e.Recalculate()

	off := skill.LevelData.Level
	c.Equal(fxp.One, e.SkillBonusFor("Brawling", "", "", nil, nil),
		"only the always-on skill bonus applies while the switch is off")

	trait.SetSwitchedOn(true)
	e.Recalculate()
	c.Equal(fxp.Three, e.SkillBonusFor("Brawling", "", "", nil, nil),
		"both skill bonuses apply while the switch is on")
	c.Equal(off+fxp.Two, skill.LevelData.Level, "the skill's level rises by the switchable bonus")
}

// TestSwitchableReactionsAndConditionalModifiers verifies the gating on the separate traversal that builds the
// reactions and conditional modifiers lists, which collects features without going through processFeatures.
func TestSwitchableReactionsAndConditionalModifiers(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	alwaysReaction := NewReactionBonus()
	alwaysReaction.Situation = "always"
	alwaysReaction.Amount = fxp.One
	switchedReaction := NewReactionBonus()
	switchedReaction.Situation = "switched"
	switchedReaction.Amount = fxp.Two
	switchedReaction.Switchable = true

	alwaysMod := NewConditionalModifierBonus()
	alwaysMod.Situation = "always"
	alwaysMod.Amount = fxp.One
	switchedMod := NewConditionalModifierBonus()
	switchedMod.Situation = "switched"
	switchedMod.Amount = fxp.Two
	switchedMod.Switchable = true

	trait := NewTrait(e, nil, false)
	trait.Name = "Gadget"
	trait.Features = Features{alwaysReaction, switchedReaction, alwaysMod, switchedMod}
	e.Traits = append(e.Traits, trait)
	e.Recalculate()

	situations := func(list []*ConditionalModifier) []string {
		result := make([]string, 0, len(list))
		for _, one := range list {
			result = append(result, one.From)
		}
		return result
	}

	c.Equal([]string{"always"}, situations(e.Reactions()),
		"only the always-on reaction is listed while the switch is off")
	c.Equal([]string{"always"}, situations(e.ConditionalModifiers()),
		"only the always-on conditional modifier is listed while the switch is off")

	trait.SetSwitchedOn(true)
	e.Recalculate()
	c.Equal([]string{"always", "switched"}, situations(e.Reactions()),
		"both reactions are listed while the switch is on")
	c.Equal([]string{"always", "switched"}, situations(e.ConditionalModifiers()),
		"both conditional modifiers are listed while the switch is on")
}

// TestSwitchableTraitMaxLevelBonus verifies the gating of a "to this trait" maximum level bonus, which the trait
// resolves locally rather than through the entity, on both the trait's own features and its modifiers'.
func TestSwitchableTraitMaxLevelBonus(t *testing.T) {
	c := check.New(t)

	newSwitchableMaxLevelBonus := func(amount string) *TraitMaxLevelBonus {
		bonus := NewTraitMaxLevelBonus()
		bonus.SelectionType = traitsel.ThisTrait
		bonus.Amount = amount
		bonus.Switchable = true
		return bonus
	}

	trait := NewTrait(nil, nil, false)
	trait.CanLevel = true
	trait.MaxLevels = "10"
	trait.Features = Features{newSwitchableMaxLevelBonus("+3")}
	c.Equal(fxp.FromInteger(10), trait.ResolvedMaxLevels(), "a switchable bonus is ignored while the switch is off")
	trait.SetSwitchedOn(true)
	c.Equal(fxp.FromInteger(13), trait.ResolvedMaxLevels(), "a switchable bonus applies while the switch is on")

	// The same bonus, carried by a modifier, follows the trait's switch.
	trait = NewTrait(nil, nil, false)
	trait.CanLevel = true
	trait.MaxLevels = "10"
	mod := NewTraitModifier(nil, nil, false)
	mod.Features = Features{newSwitchableMaxLevelBonus("+3")}
	trait.Modifiers = []*TraitModifier{mod}
	c.Equal(fxp.FromInteger(10), trait.ResolvedMaxLevels(),
		"a modifier's switchable bonus is ignored while the trait's switch is off")
	trait.SetSwitchedOn(true)
	c.Equal(fxp.FromInteger(13), trait.ResolvedMaxLevels(),
		"a modifier's switchable bonus applies while the trait's switch is on")
}

// TestSwitchableEquipmentMaxUsesBonus verifies the gating of a "to this equipment" maximum uses bonus, which the
// equipment resolves locally, on both its own features and its modifiers'.
func TestSwitchableEquipmentMaxUsesBonus(t *testing.T) {
	c := check.New(t)

	newSwitchableMaxUsesBonus := func(amount string) *EquipmentMaxUsesBonus {
		bonus := NewEquipmentMaxUsesBonus()
		bonus.SelectionType = equipmentsel.ThisEquipment
		bonus.Amount = amount
		bonus.Switchable = true
		return bonus
	}

	eqp := NewEquipment(nil, nil, false)
	eqp.MaxUses = 10
	eqp.Features = Features{newSwitchableMaxUsesBonus("+3")}
	c.Equal(10, eqp.ResolvedMaxUses(), "a switchable bonus is ignored while the switch is off")
	eqp.SetSwitchedOn(true)
	c.Equal(13, eqp.ResolvedMaxUses(), "a switchable bonus applies while the switch is on")

	eqp = NewEquipment(nil, nil, false)
	eqp.MaxUses = 10
	mod := NewEquipmentModifier(nil, nil, false)
	mod.Features = Features{newSwitchableMaxUsesBonus("+3")}
	eqp.Modifiers = []*EquipmentModifier{mod}
	c.Equal(10, eqp.ResolvedMaxUses(),
		"a modifier's switchable bonus is ignored while the equipment's switch is off")
	eqp.SetSwitchedOn(true)
	c.Equal(13, eqp.ResolvedMaxUses(),
		"a modifier's switchable bonus applies while the equipment's switch is on")
}

// TestSwitchableContainedWeightReduction verifies the gating of a contained weight reduction. Unlike most feature
// collection, this one runs for equipment that isn't equipped and even for equipment kept in the "other" list, so both
// lists are exercised.
func TestSwitchableContainedWeightReduction(t *testing.T) {
	c := check.New(t)

	build := func(e *Entity, carried bool) *Equipment {
		container := NewEquipment(e, nil, true)
		container.Name = "Bag of Holding"
		reduction := NewContainedWeightReduction()
		reduction.Reduction = "50%"
		reduction.Switchable = true
		container.Features = Features{reduction}
		child := NewEquipment(e, container, false)
		child.Name = "Anvil"
		child.BaseWeight = "10 lb"
		container.Children = []*Equipment{child}
		if carried {
			e.CarriedEquipment = append(e.CarriedEquipment, container)
		} else {
			e.OtherEquipment = append(e.OtherEquipment, container)
		}
		return container
	}

	for _, carried := range []bool{true, false} {
		where := "other"
		if carried {
			where = "carried"
		}
		e := NewEntity()
		container := build(e, carried)
		e.Recalculate()
		c.Equal(fxp.WeightFromInteger(10, fxp.Pound), container.ExtendedWeight(false, fxp.Pound),
			"%s: a switchable reduction is ignored while the switch is off", where)

		container.SetSwitchedOn(true)
		e.Recalculate()
		c.Equal(fxp.WeightFromInteger(5, fxp.Pound), container.ExtendedWeight(false, fxp.Pound),
			"%s: a switchable reduction applies while the switch is on", where)
	}

	// A reduction carried by a modifier follows the equipment's switch too.
	e := NewEntity()
	container := NewEquipment(e, nil, true)
	container.Name = "Bag of Holding"
	child := NewEquipment(e, container, false)
	child.Name = "Anvil"
	child.BaseWeight = "10 lb"
	container.Children = []*Equipment{child}
	reduction := NewContainedWeightReduction()
	reduction.Reduction = "50%"
	reduction.Switchable = true
	mod := NewEquipmentModifier(e, nil, false)
	mod.Features = Features{reduction}
	container.Modifiers = []*EquipmentModifier{mod}
	e.CarriedEquipment = append(e.CarriedEquipment, container)
	e.Recalculate()
	c.Equal(fxp.WeightFromInteger(10, fxp.Pound), container.ExtendedWeight(false, fxp.Pound),
		"a modifier's switchable reduction is ignored while the equipment's switch is off")
	container.SetSwitchedOn(true)
	e.Recalculate()
	c.Equal(fxp.WeightFromInteger(5, fxp.Pound), container.ExtendedWeight(false, fxp.Pound),
		"a modifier's switchable reduction applies while the equipment's switch is on")
}

// TestSwitchableWeaponLocalBonuses verifies the gating of the bonuses a weapon collects directly from its owner rather
// than from the entity's collected feature lists: a "to this weapon" weapon bonus and a "to this weapon" skill bonus.
func TestSwitchableWeaponLocalBonuses(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	accBonus := NewWeaponAccBonus()
	accBonus.SelectionType = wsel.ThisWeapon
	accBonus.Amount = fxp.Two
	accBonus.Switchable = true

	skillBonus := NewSkillBonus()
	skillBonus.SelectionType = skillsel.ThisWeapon
	skillBonus.Amount = fxp.Two
	skillBonus.Switchable = true

	trait := NewTrait(e, nil, false)
	trait.Name = "Gadget"
	trait.Features = Features{accBonus, skillBonus}
	w := NewWeapon(trait, false)
	w.Accuracy = ParseWeaponAccuracy("3")
	w.Defaults = []*SkillDefault{{DefaultType: DexterityID}}
	trait.Weapons = []*Weapon{w}
	e.Traits = append(e.Traits, trait)
	e.Recalculate()

	c.Equal("3", w.Accuracy.Resolve(w, nil).String(), "a switchable weapon bonus is ignored while the switch is off")
	c.Equal(fxp.Ten, w.SkillLevel(nil), "a switchable skill bonus is ignored while the switch is off")

	trait.SetSwitchedOn(true)
	e.Recalculate()
	c.Equal("5", w.Accuracy.Resolve(w, nil).String(), "a switchable weapon bonus applies while the switch is on")
	c.Equal(fxp.Twelve, w.SkillLevel(nil), "a switchable skill bonus applies while the switch is on")
}

// TestSwitchableWeaponBonusFromModifier verifies that weapon bonuses carried by a trait modifier follow the trait's
// switch, which is the loop the weapon runs over its owner's modifiers.
func TestSwitchableWeaponBonusFromModifier(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	accBonus := NewWeaponAccBonus()
	accBonus.SelectionType = wsel.ThisWeapon
	accBonus.Amount = fxp.Two
	accBonus.Switchable = true

	trait := NewTrait(e, nil, false)
	trait.Name = "Gadget"
	mod := NewTraitModifier(e, nil, false)
	mod.Name = "Scoped"
	mod.Features = Features{accBonus}
	trait.Modifiers = []*TraitModifier{mod}
	w := NewWeapon(trait, false)
	w.Accuracy = ParseWeaponAccuracy("3")
	trait.Weapons = []*Weapon{w}
	e.Traits = append(e.Traits, trait)
	e.Recalculate()

	c.Equal("3", w.Accuracy.Resolve(w, nil).String(),
		"a modifier's switchable weapon bonus is ignored while the trait's switch is off")
	trait.SetSwitchedOn(true)
	e.Recalculate()
	c.Equal("5", w.Accuracy.Resolve(w, nil).String(),
		"a modifier's switchable weapon bonus applies while the trait's switch is on")
}

// TestSwitchableThisArmorDRBonus verifies that the "this armor" DR expansion sees only the equipment's active features.
// The expansion resolves the locations to apply to by scanning the equipment's other DR bonuses, so switching those off
// must leave the "this armor" bonus with nothing to attach to, even though the bonus itself is always on.
func TestSwitchableThisArmorDRBonus(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	located := NewDRBonus()
	located.Locations = []string{TorsoID}
	located.Specialization = AllID
	located.Amount = fxp.Four
	located.Switchable = true

	thisArmor := NewDRBonus()
	thisArmor.Locations = nil // No locations, i.e. "this armor".
	thisArmor.Specialization = AllID
	thisArmor.Amount = fxp.One

	eqp := NewEquipment(e, nil, false)
	eqp.Name = "Mail Hauberk"
	eqp.Features = Features{located, thisArmor}
	e.CarriedEquipment = append(e.CarriedEquipment, eqp)
	e.Recalculate()

	c.Equal(0, e.AddDRBonusesFor(TorsoID, nil, nil)[AllID],
		"with the switch off there is no located DR, so the 'this armor' bonus has nothing to expand onto")

	eqp.SetSwitchedOn(true)
	e.Recalculate()
	c.Equal(5, e.AddDRBonusesFor(TorsoID, nil, nil)[AllID],
		"with the switch on the located DR applies and the 'this armor' bonus expands onto it")
}

// TestSwitchableThisArmorDRBonusFromModifier verifies that the DR bonuses of the equipment's modifiers, which also
// contribute locations to the "this armor" expansion, are gated by the equipment's switch there just as they are when
// features are collected: a modifier's switchable located DR bonus only becomes something for the "this armor" bonus
// to expand onto while the equipment's switch is on.
func TestSwitchableThisArmorDRBonusFromModifier(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	alwaysLocated := NewDRBonus()
	alwaysLocated.Locations = []string{TorsoID}
	alwaysLocated.Specialization = AllID
	alwaysLocated.Amount = fxp.Two

	thisArmor := NewDRBonus()
	thisArmor.Locations = nil // No locations, i.e. "this armor".
	thisArmor.Specialization = AllID
	thisArmor.Amount = fxp.One

	modLocated := NewDRBonus()
	modLocated.Locations = []string{"arm"}
	modLocated.Specialization = AllID
	modLocated.Amount = fxp.Four
	modLocated.Switchable = true

	eqp := NewEquipment(e, nil, false)
	eqp.Name = "Mail Hauberk"
	eqp.Features = Features{alwaysLocated, thisArmor}
	mod := NewEquipmentModifier(e, nil, false)
	mod.Name = "Sleeves"
	mod.Features = Features{modLocated}
	eqp.Modifiers = []*EquipmentModifier{mod}
	e.CarriedEquipment = append(e.CarriedEquipment, eqp)
	e.Recalculate()

	c.Equal(3, e.AddDRBonusesFor(TorsoID, nil, nil)[AllID],
		"the equipment's own located DR still receives the 'this armor' bonus while the switch is off")
	c.Equal(0, e.AddDRBonusesFor("arm", nil, nil)[AllID],
		"with the switch off the modifier's located DR doesn't apply, so the 'this armor' bonus doesn't reach it")

	eqp.SetSwitchedOn(true)
	e.Recalculate()
	c.Equal(3, e.AddDRBonusesFor(TorsoID, nil, nil)[AllID],
		"the equipment's own located DR is unaffected by the switch")
	c.Equal(5, e.AddDRBonusesFor("arm", nil, nil)[AllID],
		"with the switch on the modifier's located DR applies and the 'this armor' bonus expands onto it")
}

// TestScriptSwitchedOn verifies that the switch state of each kind of item is readable from the scripts the data owner
// writes, so that a script can react to it.
func TestScriptSwitchedOn(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	trait := NewTrait(e, nil, false)
	trait.Name = "Gadget"
	e.Traits = append(e.Traits, trait)
	skill := NewSkill(e, nil, false)
	skill.Name = "Brawling"
	e.Skills = append(e.Skills, skill)
	spell := NewSpell(e, nil, false)
	spell.Name = "Fireball"
	e.Spells = append(e.Spells, spell)
	eqp := NewEquipment(e, nil, false)
	eqp.Name = "Amulet"
	e.CarriedEquipment = append(e.CarriedEquipment, eqp)
	e.Recalculate()

	for _, one := range []struct {
		name     string
		arg      ScriptArg
		switcher FeatureSwitcher
	}{
		{
			name:     "trait",
			arg:      ScriptArg{Name: "self", Value: func(r *goja.Runtime) any { return newScriptTrait(r, trait) }},
			switcher: trait,
		},
		{
			name:     "skill",
			arg:      ScriptArg{Name: "self", Value: func(r *goja.Runtime) any { return newScriptSkill(r, skill) }},
			switcher: skill,
		},
		{
			name:     "spell",
			arg:      ScriptArg{Name: "self", Value: func(r *goja.Runtime) any { return newScriptSpell(r, spell) }},
			switcher: spell,
		},
		{
			name:     "equipment",
			arg:      ScriptArg{Name: "self", Value: func(r *goja.Runtime) any { return newScriptEquipment(r, eqp) }},
			switcher: eqp,
		},
	} {
		v, err := runScript(0, "String(self.switchedOn)", one.arg)
		c.NoError(err, "%s: the script should run", one.name)
		c.Equal("false", v, "%s: an off switch reads as false", one.name)

		one.switcher.SetSwitchedOn(true)
		v, err = runScript(0, "String(self.switchedOn)", one.arg)
		c.NoError(err, "%s: the script should run", one.name)
		c.Equal("true", v, "%s: an on switch reads as true", one.name)
		one.switcher.SetSwitchedOn(false)
	}
}

// TestLegacyExportSkipsSwitchedOffDRBonuses verifies that the legacy text export's hit-location armor list is built
// from the equipment's active features, so a switched-off DR bonus doesn't show up in an exported sheet.
func TestLegacyExportSkipsSwitchedOffDRBonuses(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	bonus := NewDRBonus()
	bonus.Locations = []string{TorsoID}
	bonus.Specialization = AllID
	bonus.Amount = fxp.Four
	bonus.Switchable = true

	eqp := NewEquipment(e, nil, false)
	eqp.Name = "Mail Hauberk"
	eqp.Features = Features{bonus}
	e.CarriedEquipment = append(e.CarriedEquipment, eqp)
	e.Recalculate()

	ex := &legacyExporter{entity: e}
	location := e.SheetSettings.BodyType.LookupLocationByID(e, TorsoID)
	c.NotNil(location, "the torso location should exist")

	c.Equal(0, len(ex.hitLocationEquipment(location)),
		"a switched-off DR bonus contributes no armor to the export")
	eqp.SetSwitchedOn(true)
	c.True(strings.Contains(strings.Join(ex.hitLocationEquipment(location), ","), "Mail Hauberk"),
		"a switched-on DR bonus contributes its equipment to the export")
}

// TestAnyModifierSwitchable covers the modifier scan that decides whether an item with no switchable features of its
// own still gets a switch. The answer must not depend on where the switchable modifier sits in the list or how deep it
// sits in the container tree, and it must ignore both disabled modifiers and the containers' own features, matching
// what Entity.processFeatures actually applies.
func TestAnyModifierSwitchable(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	plain := func() *TraitModifier {
		mod := NewTraitModifier(e, nil, false)
		mod.Features = Features{NewAttributeBonus(StrengthID)}
		return mod
	}
	switchable := func() *TraitModifier {
		mod := NewTraitModifier(e, nil, false)
		bonus := NewAttributeBonus(StrengthID)
		bonus.Switchable = true
		mod.Features = Features{bonus}
		return mod
	}
	disabled := func() *TraitModifier {
		mod := switchable()
		mod.Disabled = true
		return mod
	}
	container := func(children ...*TraitModifier) *TraitModifier {
		mod := NewTraitModifier(e, nil, true)
		bonus := NewAttributeBonus(StrengthID)
		bonus.Switchable = true
		mod.Features = Features{bonus} // A container's own features never apply, so they must not be counted.
		mod.Children = children
		return mod
	}
	for _, tc := range []struct {
		name      string
		modifiers []*TraitModifier
		expected  bool
	}{
		{name: "no modifiers at all"},
		{name: "nothing switchable", modifiers: []*TraitModifier{plain(), plain()}},
		{name: "only a disabled modifier is switchable", modifiers: []*TraitModifier{plain(), disabled()}},
		{name: "only a container's own features are switchable", modifiers: []*TraitModifier{container()}},
		{name: "switchable first", modifiers: []*TraitModifier{switchable(), plain()}, expected: true},
		{name: "switchable last", modifiers: []*TraitModifier{plain(), plain(), switchable()}, expected: true},
		{
			name:      "switchable between plain ones",
			modifiers: []*TraitModifier{plain(), switchable(), plain()},
			expected:  true,
		},
		{
			name:      "switchable followed by a disabled one",
			modifiers: []*TraitModifier{switchable(), disabled()},
			expected:  true,
		},
		{
			name:      "switchable nested in a container",
			modifiers: []*TraitModifier{plain(), container(plain(), switchable())},
			expected:  true,
		},
		{
			name:      "switchable nested two containers deep",
			modifiers: []*TraitModifier{container(container(plain(), switchable()))},
			expected:  true,
		},
	} {
		c.Equal(tc.expected, anyModifierSwitchable(tc.modifiers, func(mod *TraitModifier) Features {
			return mod.Features
		}), tc.name)
	}
}

// TestContainerSwitchIsClearedForSkillsAndSpells verifies that a switch state doesn't linger on an item that can never
// present a switch. SwitchedOn lives in the edit data shared by containers and non-containers, but a container skill or
// spell reports nothing to switch, so an on-state that got set on one -- or that was left behind when a non-container
// was turned into a container -- would be written to disk with no way for the user to reach it again. Trait and
// equipment containers keep theirs: both can genuinely contribute switchable features.
func TestContainerSwitchIsClearedForSkillsAndSpells(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	type switchableItem interface {
		FeatureSwitcher
		ClearUnusedFieldsForType()
	}
	for _, tc := range []struct {
		name    string
		make    func(container bool) switchableItem
		cleared bool
	}{
		{name: "skill", make: func(container bool) switchableItem { return NewSkill(e, nil, container) }, cleared: true},
		{name: "spell", make: func(container bool) switchableItem { return NewSpell(e, nil, container) }, cleared: true},
		{name: "trait", make: func(container bool) switchableItem { return NewTrait(e, nil, container) }},
		{name: "equipment", make: func(container bool) switchableItem { return NewEquipment(e, nil, container) }},
	} {
		// A non-container's switch is never touched, no matter the type.
		item := tc.make(false)
		item.SetSwitchedOn(true)
		item.ClearUnusedFieldsForType()
		c.True(item.IsSwitchedOn(), "%s: a non-container keeps its switch", tc.name)

		// A container's switch is cleared only where it could never be reached again.
		item = tc.make(true)
		item.SetSwitchedOn(true)
		item.ClearUnusedFieldsForType()
		c.Equal(!tc.cleared, item.IsSwitchedOn(), "%s container: switch state after clearing the unused fields",
			tc.name)

		// The same has to hold through a save, which is where a stale state would otherwise reach the file.
		item = tc.make(true)
		item.SetSwitchedOn(true)
		data, err := jio.Marshal(item)
		c.NoError(err, "%s container: should marshal", tc.name)
		if tc.cleared {
			c.NotContains(string(data), "switched_on", "%s container: a stale switch is not written", tc.name)
		} else {
			c.Contains(string(data), `"switched_on":true`, "%s container: the switch is written", tc.name)
		}
	}
}

// TestEquipmentSwitchCellDimming verifies when the equipment switch cell is drawn dimmed. Dimming means "throwing this
// switch would change nothing right now", so it may only happen when every switchable feature the item contributes is
// one that reaches the character solely through the entity's collection pass over really-equipped carried equipment.
// The switch column is shown for the other equipment list as well as the carried one, and the equipped flag says
// nothing useful there -- new equipment starts out equipped and nothing clears the flag when an item is created in or
// moved to that list -- so both the list the item is rooted in and its equipped state have to be taken into account.
// Either way the cell stays a live switch; the dimming is purely visual.
func TestEquipmentSwitchCellDimming(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	attributeBonus := func() Feature {
		bonus := NewAttributeBonus(StrengthID)
		bonus.Switchable = true
		return bonus
	}
	weightReduction := func() Feature {
		reduction := NewContainedWeightReduction()
		reduction.Reduction = "50%"
		reduction.Switchable = true
		return reduction
	}
	maxUsesBonus := func(sel equipmentsel.Type) Feature {
		bonus := NewEquipmentMaxUsesBonus()
		bonus.SelectionType = sel
		bonus.Amount = "+3"
		bonus.Switchable = true
		return bonus
	}
	weaponAccBonus := func(sel wsel.Type) Feature {
		bonus := NewWeaponAccBonus()
		bonus.SelectionType = sel
		bonus.Amount = fxp.Two
		bonus.Switchable = true
		return bonus
	}
	weaponSkillBonus := func(sel skillsel.Type) Feature {
		bonus := NewSkillBonus()
		bonus.SelectionType = sel
		bonus.Amount = fxp.Two
		bonus.Switchable = true
		return bonus
	}
	// Both constructors put what they build into the entity's carried equipment list, since that is where an item has
	// to live for the character to collect anything from it at all.
	item := func(features ...Feature) *Equipment {
		eqp := NewEquipment(e, nil, false)
		eqp.Name = "Amulet"
		eqp.Equipped = false
		eqp.Features = features
		e.CarriedEquipment = append(e.CarriedEquipment, eqp)
		return eqp
	}
	withWeapon := func(eqp *Equipment) *Equipment {
		w := NewWeapon(eqp, false)
		w.Defaults = []*SkillDefault{{DefaultType: DexterityID}}
		eqp.Weapons = []*Weapon{w}
		return eqp
	}
	fullBag := func(features ...Feature) *Equipment {
		bag := NewEquipment(e, nil, true)
		bag.Name = "Bag"
		bag.Equipped = false
		bag.Features = features
		child := NewEquipment(e, bag, false)
		child.Name = "Anvil"
		child.BaseWeight = "10 lb"
		bag.Children = []*Equipment{child}
		e.CarriedEquipment = append(e.CarriedEquipment, bag)
		return bag
	}
	// inOther moves a top-level item the constructors above placed in the carried list over to the other equipment
	// list, leaving its equipped flag alone.
	inOther := func(eqp *Equipment) *Equipment {
		e.CarriedEquipment = slices.DeleteFunc(e.CarriedEquipment, func(one *Equipment) bool { return one == eqp })
		e.OtherEquipment = append(e.OtherEquipment, eqp)
		return eqp
	}
	nested := func(parent *Equipment, features ...Feature) *Equipment {
		child := NewEquipment(e, parent, false)
		child.Name = "Ring"
		child.Features = features
		parent.Children = append(parent.Children, child)
		return child
	}

	for _, tc := range []struct {
		name  string
		build func() *Equipment
		dim   bool
	}{
		{
			name: "carried, equipped item with an entity-collected feature",
			build: func() *Equipment {
				eqp := item(attributeBonus())
				eqp.Equipped = true
				return eqp
			},
		},
		{
			name:  "unequipped item with an entity-collected feature",
			build: func() *Equipment { return item(attributeBonus()) },
			dim:   true,
		},
		{
			// Nothing clears the equipped flag when an item is created in or moved to the other equipment list, so
			// the flag can't be trusted there: the character collects nothing from that list.
			name: "equipped item in the other list with an entity-collected feature",
			build: func() *Equipment {
				eqp := item(attributeBonus())
				eqp.Equipped = true
				return inOther(eqp)
			},
			dim: true,
		},
		{
			name: "equipped container in the other list with a contained weight reduction",
			build: func() *Equipment {
				bag := fullBag(weightReduction())
				bag.Equipped = true
				return inOther(bag)
			},
		},
		{
			name: "equipped item in the other list with a 'this equipment' max uses bonus",
			build: func() *Equipment {
				eqp := item(maxUsesBonus(equipmentsel.ThisEquipment))
				eqp.Equipped = true
				return inOther(eqp)
			},
		},
		{
			// Which list an item belongs to is decided by the root of the tree it lives in, so an equipped item
			// tucked inside a container in the other list is no more use to the character than the container is.
			name: "equipped child of an equipped container in the other list",
			build: func() *Equipment {
				bag := fullBag()
				bag.Equipped = true
				return nested(inOther(bag), attributeBonus())
			},
			dim: true,
		},
		{
			name: "equipped child of an equipped carried container",
			build: func() *Equipment {
				bag := fullBag()
				bag.Equipped = true
				return nested(bag, attributeBonus())
			},
		},
		{
			// A library list, a template and a loot sheet have no owning entity and therefore no carried/other split,
			// so the equipped state alone decides there. The switch column is never shown outside a character sheet,
			// so this only has to stay sensible, not visible.
			name: "equipped item with no owning entity",
			build: func() *Equipment {
				eqp := NewEquipment(nil, nil, false)
				eqp.Features = Features{attributeBonus()}
				return eqp
			},
		},
		{
			name: "unequipped item with no owning entity",
			build: func() *Equipment {
				eqp := NewEquipment(nil, nil, false)
				eqp.Equipped = false
				eqp.Features = Features{attributeBonus()}
				return eqp
			},
			dim: true,
		},
		{
			name:  "unequipped container with a contained weight reduction",
			build: func() *Equipment { return fullBag(weightReduction()) },
		},
		{
			name:  "unequipped empty container with a contained weight reduction",
			build: func() *Equipment { return item(weightReduction()) },
			dim:   true,
		},
		{
			name: "unequipped container whose modifier supplies the weight reduction",
			build: func() *Equipment {
				bag := fullBag()
				mod := NewEquipmentModifier(e, nil, false)
				mod.Features = Features{weightReduction()}
				bag.Modifiers = []*EquipmentModifier{mod}
				return bag
			},
		},
		{
			name:  "unequipped item with a 'this equipment' max uses bonus",
			build: func() *Equipment { return item(maxUsesBonus(equipmentsel.ThisEquipment)) },
		},
		{
			name:  "unequipped item with a 'named equipment' max uses bonus",
			build: func() *Equipment { return item(maxUsesBonus(equipmentsel.EquipmentWithName)) },
			dim:   true,
		},
		{
			name:  "unequipped weapon with a 'this weapon' bonus",
			build: func() *Equipment { return withWeapon(item(weaponAccBonus(wsel.ThisWeapon))) },
		},
		{
			name:  "unequipped item with a 'this weapon' bonus but no weapon",
			build: func() *Equipment { return item(weaponAccBonus(wsel.ThisWeapon)) },
			dim:   true,
		},
		{
			// The weapon reads its owner's active features directly, so a "named weapon" bonus can still land on the
			// item's own weapon while the item is unequipped.
			name:  "unequipped weapon with a 'named weapon' bonus",
			build: func() *Equipment { return withWeapon(item(weaponAccBonus(wsel.WithName))) },
		},
		{
			name:  "unequipped item with a 'named weapon' bonus but no weapon",
			build: func() *Equipment { return item(weaponAccBonus(wsel.WithName)) },
			dim:   true,
		},
		{
			name:  "unequipped weapon with a 'weapon with required skill' bonus",
			build: func() *Equipment { return withWeapon(item(weaponAccBonus(wsel.WithRequiredSkill))) },
			dim:   true,
		},
		{
			name:  "unequipped weapon with a 'this weapon' skill bonus",
			build: func() *Equipment { return withWeapon(item(weaponSkillBonus(skillsel.ThisWeapon))) },
		},
		{
			name:  "unequipped weapon with a named skill bonus",
			build: func() *Equipment { return withWeapon(item(weaponSkillBonus(skillsel.Name))) },
			dim:   true,
		},
		{
			name: "unequipped item with both an entity-collected feature and a locally resolved one",
			build: func() *Equipment {
				bag := fullBag(attributeBonus(), weightReduction())
				return bag
			},
		},
		{
			name: "item with zero quantity is not really equipped",
			build: func() *Equipment {
				eqp := item(attributeBonus())
				eqp.Equipped = true
				eqp.Quantity = 0
				return eqp
			},
			dim: true,
		},
	} {
		eqp := tc.build()
		var data CellData
		eqp.CellData(EquipmentSwitchColumn, &data)
		c.Equal(cell.Switch, data.Type, "%s: the cell is still a switch", tc.name)
		c.Equal(tc.dim, data.Dim, "%s: dim state", tc.name)

		// Dimming is purely visual: the cell must still report the switch's state so it stays clickable.
		eqp.SetSwitchedOn(true)
		data = CellData{}
		eqp.CellData(EquipmentSwitchColumn, &data)
		c.Equal(cell.Switch, data.Type, "%s: the cell is still a switch while on", tc.name)
		c.True(data.Checked, "%s: the cell reports the switch as on", tc.name)
		c.Equal(tc.dim, data.Dim, "%s: dim state doesn't depend on the switch's own state", tc.name)
	}

	// An item with nothing to switch gets no switch cell at all, dimmed or otherwise.
	var data CellData
	item(NewAttributeBonus(StrengthID)).CellData(EquipmentSwitchColumn, &data)
	c.NotEqual(cell.Switch, data.Type, "an item with nothing switchable gets no switch cell")
}

// TestTraitSwitchCellDimming verifies when the trait switch cell is drawn dimmed. A disabled trait -- or one inside a
// disabled container -- is in the same position as unequipped equipment: every collection pass over the traits skips
// the ones that aren't enabled, so its features, its modifiers' features, its weapons and its reactions and
// conditional modifiers are all left out, and even the one thing a trait resolves from its own switchable features,
// ResolvedMaxLevels, is only consulted for enabled traits. Throwing the switch therefore changes nothing the user can
// see, which is exactly what dimming means. As with equipment, the cell stays a live switch; the dimming is purely
// visual.
func TestTraitSwitchCellDimming(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	attributeBonus := func() Feature {
		bonus := NewAttributeBonus(StrengthID)
		bonus.Switchable = true
		return bonus
	}
	maxLevelBonus := func() Feature {
		bonus := NewTraitMaxLevelBonus()
		bonus.SelectionType = traitsel.ThisTrait
		bonus.Amount = "+2"
		bonus.Switchable = true
		return bonus
	}
	trait := func(features ...Feature) *Trait {
		one := NewTrait(e, nil, false)
		one.Name = "Gadget"
		one.Features = features
		e.Traits = append(e.Traits, one)
		return one
	}
	inContainer := func(one *Trait) *Trait {
		parent := NewTrait(e, nil, true)
		parent.Name = "Gadgets"
		e.Traits = slices.DeleteFunc(e.Traits, func(other *Trait) bool { return other == one })
		one.SetParent(parent)
		parent.Children = []*Trait{one}
		e.Traits = append(e.Traits, parent)
		return one
	}

	for _, tc := range []struct {
		name  string
		build func() *Trait
		dim   bool
	}{
		{
			name:  "enabled trait",
			build: func() *Trait { return trait(attributeBonus()) },
		},
		{
			name: "disabled trait",
			build: func() *Trait {
				one := trait(attributeBonus())
				one.Disabled = true
				return one
			},
			dim: true,
		},
		{
			name: "enabled trait inside an enabled container",
			build: func() *Trait {
				return inContainer(trait(attributeBonus()))
			},
		},
		{
			name: "enabled trait inside a disabled container",
			build: func() *Trait {
				one := inContainer(trait(attributeBonus()))
				one.Parent().Disabled = true
				return one
			},
			dim: true,
		},
		{
			name: "disabled trait whose switchable feature comes from a modifier",
			build: func() *Trait {
				one := trait()
				mod := NewTraitModifier(e, nil, false)
				mod.Name = "Blessed"
				mod.Features = Features{attributeBonus()}
				one.Modifiers = []*TraitModifier{mod}
				one.Disabled = true
				return one
			},
			dim: true,
		},
		{
			// ResolvedMaxLevels applies a switchable "to this trait" maximum level bonus without asking the entity
			// for anything, but the only thing that consults it for a sheet's traits skips the disabled ones, so the
			// switch is still inert here.
			name: "disabled trait with a 'this trait' maximum level bonus",
			build: func() *Trait {
				one := trait(maxLevelBonus())
				one.CanLevel = true
				one.MaxLevels = "3"
				one.Levels = fxp.One
				one.Disabled = true
				return one
			},
			dim: true,
		},
	} {
		one := tc.build()
		var data CellData
		one.CellData(TraitSwitchColumn, &data)
		c.Equal(cell.Switch, data.Type, "%s: the cell is still a switch", tc.name)
		c.Equal(tc.dim, data.Dim, "%s: dim state", tc.name)

		// Dimming is purely visual: the cell must still report the switch's state so it stays clickable.
		one.SetSwitchedOn(true)
		data = CellData{}
		one.CellData(TraitSwitchColumn, &data)
		c.Equal(cell.Switch, data.Type, "%s: the cell is still a switch while on", tc.name)
		c.True(data.Checked, "%s: the cell reports the switch as on", tc.name)
		c.Equal(tc.dim, data.Dim, "%s: dim state doesn't depend on the switch's own state", tc.name)
	}

	// A trait with nothing to switch gets no switch cell at all, dimmed or otherwise.
	var data CellData
	trait(NewAttributeBonus(StrengthID)).CellData(TraitSwitchColumn, &data)
	c.NotEqual(cell.Switch, data.Type, "a trait with nothing switchable gets no switch cell")
}

// TestThisArmorDRBonusExpansionKeepsSwitchable verifies that the copy of a "this armor" DR bonus that the entity emits
// onto the armor's locations still describes itself the way the original does: a switchable original yields a
// switchable copy. The flag has no bearing on whether the copy applies -- the original has already passed the switch
// gate by the time it is expanded -- but anything that later asks the collected bonuses whether they are switchable
// should get the truthful answer.
func TestThisArmorDRBonusExpansionKeepsSwitchable(t *testing.T) {
	c := check.New(t)
	for _, switchable := range []bool{false, true} {
		e := NewEntity()
		located := NewDRBonus()
		located.Locations = []string{TorsoID}
		located.Specialization = AllID
		located.Amount = fxp.Four
		thisArmor := NewDRBonus()
		thisArmor.Locations = nil // No locations, i.e. "this armor".
		thisArmor.Specialization = "crushing"
		thisArmor.Amount = fxp.One
		thisArmor.Switchable = switchable
		eqp := NewEquipment(e, nil, false)
		eqp.Name = "Mail Hauberk"
		eqp.Features = Features{located, thisArmor}
		eqp.SwitchedOn = true
		e.CarriedEquipment = append(e.CarriedEquipment, eqp)
		e.Recalculate()

		var expanded *DRBonus
		for _, bonus := range e.features.drBonuses {
			if bonus.Specialization == "crushing" {
				expanded = bonus
			}
		}
		c.NotNil(expanded, "the 'this armor' bonus must have been expanded onto the armor's locations")
		c.Equal([]string{TorsoID}, expanded.Locations, "the expansion must cover the armor's located DR")
		c.Equal(switchable, expanded.IsSwitchable(),
			"the expansion must carry the switchable flag of the bonus it was made from")
	}
}

// TestSpellIsALeveledOwner verifies the leveled-owner behavior a spell's features rely on: a per-level feature on a
// spell scales with the spell's level, exactly as one on a skill does with the skill's, and a spell container --
// which never has features that apply -- reports itself as not leveled.
func TestSpellIsALeveledOwner(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	spell := NewSpell(e, nil, false)
	spell.Name = "Fireball"
	spell.Points = fxp.Four
	bonus := NewAttributeBonus(StrengthID)
	bonus.Amount = fxp.One
	bonus.PerLevel = true
	spell.Features = Features{bonus}
	e.Spells = append(e.Spells, spell)
	e.Recalculate()

	c.True(spell.IsLeveled(), "a spell is a leveled owner")
	c.True(spell.LevelData.Level > 0, "precondition: the spell resolves to a positive level")
	c.Equal(spell.LevelData.Level, spell.CurrentLevel(), "a spell's current level is its resolved level")
	c.Equal(spell.LevelData.Level, e.AttributeBonusFor(StrengthID, stlimit.None, nil),
		"a per-level bonus on a spell must be multiplied by the spell's level")

	container := NewSpell(e, nil, true)
	c.False(container.IsLeveled(), "a spell container is not a leveled owner")
}

// TestSwitchCellTooltipMentionsCascadeOnlyForContainers verifies that the switch cell's tooltip only offers the
// Option/Alt-click cascade for a row that has contents to cascade to.
func TestSwitchCellTooltipMentionsCascadeOnlyForContainers(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	switchable := func() *AttributeBonus {
		bonus := NewAttributeBonus(StrengthID)
		bonus.Switchable = true
		return bonus
	}
	tooltipFor := func(node interface{ CellData(int, *CellData) }, column int) string {
		var data CellData
		node.CellData(column, &data)
		c.Equal(cell.Switch, data.Type, "the row must have a switch cell")
		return data.Tooltip
	}
	c.False(strings.Contains(SwitchCellTooltip(false), "Option/Alt"), "a leaf's tooltip must not offer the cascade")
	c.True(strings.Contains(SwitchCellTooltip(true), "Option/Alt"), "a container's tooltip must offer the cascade")

	trait := NewTrait(e, nil, false)
	trait.Features = Features{switchable()}
	c.Equal(SwitchCellTooltip(false), tooltipFor(trait, TraitSwitchColumn), "a trait's tooltip must not offer the cascade")
	traitContainer := NewTrait(e, nil, true)
	mod := NewTraitModifier(e, nil, false)
	mod.Features = Features{switchable()}
	traitContainer.Modifiers = []*TraitModifier{mod}
	c.Equal(SwitchCellTooltip(true), tooltipFor(traitContainer, TraitSwitchColumn),
		"a trait container's tooltip must offer the cascade")

	eqp := NewEquipment(e, nil, false)
	eqp.Features = Features{switchable()}
	c.Equal(SwitchCellTooltip(false), tooltipFor(eqp, EquipmentSwitchColumn),
		"a piece of equipment's tooltip must not offer the cascade")
	eqpContainer := NewEquipment(e, nil, true)
	eqpContainer.Features = Features{switchable()}
	c.Equal(SwitchCellTooltip(true), tooltipFor(eqpContainer, EquipmentSwitchColumn),
		"an equipment container's tooltip must offer the cascade")

	skill := NewSkill(e, nil, false)
	skill.Features = Features{switchable()}
	c.Equal(SwitchCellTooltip(false), tooltipFor(skill, SkillSwitchColumn), "a skill's tooltip must not offer the cascade")
	spell := NewSpell(e, nil, false)
	spell.Features = Features{switchable()}
	c.Equal(SwitchCellTooltip(false), tooltipFor(spell, SpellSwitchColumn), "a spell's tooltip must not offer the cascade")
}
