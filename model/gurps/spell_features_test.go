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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellmatch"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// newSpellFeatureTestSkillBonus returns a skill bonus suitable for hanging off a spell, naming the given skill.
func newSpellFeatureTestSkillBonus(skillName string, amount fxp.Int) *SkillBonus {
	bonus := NewSkillBonus()
	bonus.NameCriteria.Compare = criteria.IsText
	bonus.NameCriteria.Qualifier = skillName
	bonus.Amount = amount
	return bonus
}

// TestSpellFeaturesRoundTrip verifies that the features a spell now carries are written and read back, including the
// switchable flag, and that a spell with none writes no "features" key at all.
func TestSpellFeaturesRoundTrip(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	spell := NewSpell(e, nil, false)
	spell.Name = "Fireball"
	data, err := jio.Marshal(spell)
	c.NoError(err, "a spell with no features should marshal")
	c.NotContains(string(data), `"features"`, "a spell with no features writes no features key")

	bonus := newSpellFeatureTestSkillBonus("Alchemy", fxp.Two)
	bonus.Switchable = true
	spell.Features = Features{bonus, NewAttributeBonus(StrengthID)}
	data, err = jio.Marshal(spell)
	c.NoError(err, "a spell with features should marshal")
	c.Contains(string(data), `"features"`, "a spell's features are written")
	c.Contains(string(data), `"switchable":true`, "a spell's switchable feature keeps its flag")

	var restored Spell
	c.NoError(jio.Unmarshal(data, &restored), "a spell with features should load")
	c.Equal(2, len(restored.Features), "both features come back")
	if len(restored.Features) == 2 {
		reloaded, ok := restored.Features[0].(*SkillBonus)
		c.True(ok, "the first feature is still a skill bonus")
		if ok {
			c.Equal("Alchemy", reloaded.NameCriteria.Qualifier, "the skill bonus keeps its criteria")
			c.Equal(fxp.Two, reloaded.Amount, "the skill bonus keeps its amount")
			c.True(reloaded.IsSwitchable(), "the skill bonus is still switchable")
		}
		c.False(restored.Features[1].IsSwitchable(), "the second feature is still always-on")
	}

	// The stored form is stable across a round-trip. The comparison is made over the form without the derived "calc"
	// values, since the reloaded spell has no owning entity to derive them from.
	stored, err := MarshalWithoutCalc(spell)
	c.NoError(err, "the spell should marshal without its derived values")
	again, err := MarshalWithoutCalc(&restored)
	c.NoError(err, "the reloaded spell should marshal without its derived values")
	c.Equal(string(stored), string(again), "the stored form is stable across a round-trip")
}

// TestSpellEditDataCopiesFeatures verifies that a spell editor works on its own copy of the features, in both
// directions, so that editing them in an editor doesn't reach back into the spell it was opened on.
func TestSpellEditDataCopiesFeatures(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	spell := NewSpell(e, nil, false)
	spell.Name = "Fireball"
	spell.Features = Features{newSpellFeatureTestSkillBonus("Alchemy", fxp.Two)}

	// What the editor is handed when it opens.
	var edit SpellEditData
	edit.CopyFrom(spell)
	c.Equal(1, len(edit.Features), "the copy holds the spell's features")
	c.False(edit.Features[0] == spell.Features[0], "the copy holds its own features, not the spell's")

	// Editing the copy leaves the original alone.
	editBonus := skillBonusAt(c, edit.Features, 0)
	spellBonus := skillBonusAt(c, spell.Features, 0)
	editBonus.NameCriteria.Qualifier = "Astronomy"
	editBonus.SetSwitchable(true)
	c.Equal("Alchemy", spellBonus.NameCriteria.Qualifier, "the original criteria is untouched")
	c.False(spellBonus.IsSwitchable(), "the original switchable flag is untouched")

	// Applying the editor's data hands the target its own copy in turn.
	target := NewSpell(e, nil, false)
	edit.ApplyTo(target)
	c.Equal(1, len(target.Features), "the applied data holds the edited features")
	targetBonus := skillBonusAt(c, target.Features, 0)
	c.Equal("Astronomy", targetBonus.NameCriteria.Qualifier, "the edit was applied")
	c.True(targetBonus.IsSwitchable(), "the switchable flag was applied")
	c.False(target.Features[0] == edit.Features[0], "the target holds its own features, not the editor's")
	targetBonus.Amount = fxp.Ten
	c.Equal(fxp.Two, editBonus.Amount, "editing the target doesn't reach back into the editor")
}

// skillBonusAt returns the feature at the given index as a *SkillBonus, failing the test if it isn't one.
func skillBonusAt(c check.Checker, features Features, index int) *SkillBonus {
	c.True(index < len(features), "the feature index must be in range")
	bonus, ok := features[index].(*SkillBonus)
	c.True(ok, "the feature must be a skill bonus")
	if !ok {
		return NewSkillBonus()
	}
	return bonus
}

// TestSpellHashIncludesFeatures verifies that a spell's features participate in the source-data hash, so that a spell
// whose features differ from its library source is reported as out of sync.
func TestSpellHashIncludesFeatures(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	spell := NewSpell(e, nil, false)
	spell.Name = "Fireball"
	before := Hash64(spell)

	spell.Features = Features{newSpellFeatureTestSkillBonus("Alchemy", fxp.Two)}
	withFeature := Hash64(spell)
	c.NotEqual(before, withFeature, "adding a feature changes the spell's hash")

	spell.Features[0].SetSwitchable(true)
	c.NotEqual(withFeature, Hash64(spell), "making the feature switchable changes the spell's hash")

	spell.Features = nil
	c.Equal(before, Hash64(spell), "removing the feature restores the original hash")
}

// TestSpellSyncWithSourceClonesFeatures verifies that syncing a spell with its library source brings the source's
// features across as clones, so that later edits to one don't reach into the other. The library data is injected
// directly into the source matcher, which is what a real sync gets from the library file it loaded.
func TestSpellSyncWithSourceClonesFeatures(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	// Stand in for the spell as it appears in a library file: same identity, but carrying a feature.
	source := NewSpell(nil, nil, false)
	source.Name = "Fireball"
	bonus := newSpellFeatureTestSkillBonus("Alchemy", fxp.Two)
	bonus.Switchable = true
	source.Features = Features{bonus}

	local := NewSpell(e, nil, false)
	local.Name = "Fireball"
	e.Spells = append(e.Spells, local)
	libFile := LibraryFile{Library: "Test Library", Path: "Test.spl"}
	local.Source = Source{LibraryFile: libFile, TID: source.TID}
	e.SourceMatcher().libHashes = map[LibraryFile]libSrcData{
		libFile: {dataHashes: map[tid.TID]HashAndData{source.TID: {Hash: Hash64(source), Data: source}}},
	}

	c.Equal(0, len(local.Features), "precondition: the local copy starts with no features")
	local.SyncWithSource()
	c.Equal(1, len(local.Features), "the sync brings the source's features across")
	if len(local.Features) == 1 {
		c.True(local.Features[0].IsSwitchable(), "the synced feature keeps its switchable flag")
		c.False(local.Features[0] == source.Features[0], "the synced feature is a clone, not the source's own")
		skillBonusAt(c, local.Features, 0).Amount = fxp.Ten
		c.Equal(fxp.Two, skillBonusAt(c, source.Features, 0).Amount, "editing the synced copy leaves the source alone")
	}
}

// TestSpellFeaturesProcessedByEntity verifies that a spell's features reach the entity at all, and that the switchable
// ones among them honor the spell's switch.
func TestSpellFeaturesProcessedByEntity(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	always := NewAttributeBonus(StrengthID)
	always.Amount = fxp.One
	switched := NewAttributeBonus(StrengthID)
	switched.Amount = fxp.Two
	switched.Switchable = true

	spell := NewSpell(e, nil, false)
	spell.Name = "Might"
	spell.Points = fxp.Four
	spell.Features = Features{always, switched}
	e.Spells = append(e.Spells, spell)
	e.Recalculate()

	c.Equal(fxp.One, e.AttributeBonusFor(StrengthID, stlimit.None, nil),
		"a spell's always-on feature applies, its switchable one doesn't while the switch is off")

	spell.SetSwitchedOn(true)
	e.Recalculate()
	c.Equal(fxp.Three, e.AttributeBonusFor(StrengthID, stlimit.None, nil),
		"both of a spell's features apply while its switch is on")

	// A container spell's features never apply, since containers are skipped by the traversal.
	e = NewEntity()
	container := NewSpell(e, nil, true)
	container.Name = "Fire College"
	container.Features = Features{NewAttributeBonus(StrengthID)}
	e.Spells = append(e.Spells, container)
	e.Recalculate()
	c.Equal(fxp.Int(0), e.AttributeBonusFor(StrengthID, stlimit.None, nil),
		"a container spell's features are not applied")
}

// TestSpellConditionalModifierFromFeatures verifies that a spell participates in the reaction and conditional modifier
// traversal, which is a separate walk from the one that feeds processFeatures.
func TestSpellConditionalModifierFromFeatures(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	reaction := NewReactionBonus()
	reaction.Situation = "from the faithful"
	reaction.Amount = fxp.Two
	condition := NewConditionalModifierBonus()
	condition.Situation = "when casting at night"
	condition.Amount = fxp.One
	condition.Switchable = true

	spell := NewSpell(e, nil, false)
	spell.Name = "Bless"
	spell.Points = fxp.Four
	spell.Features = Features{reaction, condition}
	e.Spells = append(e.Spells, spell)
	e.Recalculate()

	reactions := e.Reactions()
	c.Equal(1, len(reactions), "a spell's reaction bonus is collected")
	if len(reactions) == 1 {
		c.Equal("from the faithful", reactions[0].From, "the reaction names the spell's situation")
		c.Equal(fxp.Two, reactions[0].Total(), "the reaction carries the spell's amount")
	}
	c.Equal(0, len(e.ConditionalModifiers()), "the switchable conditional modifier is held back while the switch is off")

	spell.SetSwitchedOn(true)
	e.Recalculate()
	mods := e.ConditionalModifiers()
	c.Equal(1, len(mods), "the switchable conditional modifier appears once the switch is on")
	if len(mods) == 1 {
		c.Equal("when casting at night", mods[0].From, "the modifier names the spell's situation")
	}
}

// TestSpellWeaponBonusFromSpellFeatures verifies that a weapon owned by a spell picks up the "to this weapon" bonuses
// carried by the spell's own features, which is what made Spell an ordinary WeaponOwner rather than one that always
// reported an empty feature list.
func TestSpellWeaponBonusFromSpellFeatures(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	accBonus := NewWeaponAccBonus()
	accBonus.SelectionType = wsel.ThisWeapon
	accBonus.Amount = fxp.Two

	spell := NewSpell(e, nil, false)
	spell.Name = "Fireball"
	spell.Points = fxp.Four
	spell.Features = Features{accBonus}
	w := NewWeapon(spell, false)
	w.Accuracy = ParseWeaponAccuracy("3")
	spell.Weapons = []*Weapon{w}
	e.Spells = append(e.Spells, spell)
	e.Recalculate()
	c.Equal("5", w.Accuracy.Resolve(w, nil).String(), "the spell's own weapon bonus reaches its weapon")

	// And the switch gates it, just as it does for a trait or a piece of equipment.
	accBonus.Switchable = true
	e.Recalculate()
	c.Equal("3", w.Accuracy.Resolve(w, nil).String(), "a switchable bonus is held back while the spell's switch is off")
	spell.SetSwitchedOn(true)
	e.Recalculate()
	c.Equal("5", w.Accuracy.Resolve(w, nil).String(), "a switchable bonus applies while the spell's switch is on")
}

// TestSpellBonusFromItsOwnSpell verifies that a spell carrying a bonus that names itself resolves to a sane level
// rather than recursing forever. The spell's level is what the bonus is collected during, so a bonus pointing back at
// its own spell is the shortest cycle the machinery has to survive.
func TestSpellBonusFromItsOwnSpell(t *testing.T) {
	c := check.New(t)

	// A control entity, to establish the level the spell would have without any bonus.
	control := NewEntity()
	baseline := NewSpell(control, nil, false)
	baseline.Name = "Fireball"
	baseline.Points = fxp.Four
	control.Spells = append(control.Spells, baseline)
	control.Recalculate()
	c.True(baseline.LevelData.Level > 0, "precondition: the spell resolves to a positive level")

	e := NewEntity()
	spell := NewSpell(e, nil, false)
	spell.Name = "Fireball"
	spell.Points = fxp.Four
	bonus := NewSpellBonus()
	bonus.SpellMatchType = spellmatch.Name
	bonus.NameCriteria.Compare = criteria.IsText
	bonus.NameCriteria.Qualifier = "Fireball"
	bonus.Amount = fxp.Two
	spell.Features = Features{bonus}
	e.Spells = append(e.Spells, spell)

	c.NotPanics(func() { e.Recalculate() }, "a self-referencing spell bonus must not blow the stack")
	c.Equal(baseline.LevelData.Level+fxp.Two, spell.LevelData.Level, "the spell's own bonus raises its own level")

	// The per-level form reads the spell's own cached level rather than recomputing it, so it also terminates.
	bonus.PerLevel = true
	c.NotPanics(func() { e.Recalculate() }, "a per-level self-referencing spell bonus must not blow the stack")
	c.True(spell.LevelData.Level > 0, "the spell still resolves to a positive level")
	c.NotEqual(fxp.Min, spell.LevelData.Level, "the spell's level stays resolvable")
}

// TestSpellFillWithNameableKeysIncludesFeatures verifies that the substitution keys used inside a spell's features are
// offered to the user along with the ones in the spell's own text, so a spell can carry a feature whose target is
// filled in when the spell is added to a sheet.
func TestSpellFillWithNameableKeysIncludesFeatures(t *testing.T) {
	c := check.New(t)

	spell := NewSpell(nil, nil, false)
	spell.Name = "Enhance @what@"
	bonus := newSpellFeatureTestSkillBonus("@skill@", fxp.Two)
	bonus.SpecializationCriteria.Compare = criteria.IsText
	bonus.SpecializationCriteria.Qualifier = "@spec@"
	spell.Features = Features{bonus}

	m := make(map[string]string)
	spell.FillWithNameableKeys(m, nil)
	c.Equal(map[string]string{
		"what":  "what",
		"skill": "skill",
		"spec":  "spec",
	}, m, "the keys from the spell's features are offered along with its own")
}
