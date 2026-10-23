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
	"bytes"
	"encoding/json/v2"
	"hash"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/eqcontainer"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// recordingHasher captures what would be hashed, so that the tests can inspect it. Only the io.Writer portion of
// hash.Hash is used by HashJSON.
type recordingHasher struct {
	hash.Hash
	buffer bytes.Buffer
}

func (r *recordingHasher) Write(data []byte) (int, error) {
	return r.buffer.Write(data)
}

// hashedJSON returns the JSON that HashJSON would feed into a hasher for the given object.
func hashedJSON(t *testing.T, in any) string {
	t.Helper()
	var rec recordingHasher
	HashJSON(&rec, in)
	return rec.buffer.String()
}

// newPopulatedEntity returns an entity holding at least one of every kind of object that publishes derived values, so
// that a marshal of it exercises every one of their marshalers.
func newPopulatedEntity() *Entity {
	e := NewEntity()

	trait := NewTrait(e, nil, false)
	trait.Name = "Claws"
	trait.LocalNotes = "Sharp. <script>1 + 1</script>"
	modifier := NewTraitModifier(e, nil, false)
	modifier.Name = "Long"
	modifier.LocalNotes = "Longer. <script>2 + 2</script>"
	trait.Modifiers = []*TraitModifier{modifier}
	trait.Weapons = []*Weapon{NewWeapon(trait, true), NewWeapon(trait, false)}
	e.Traits = append(e.Traits, trait)

	skill := NewSkill(e, nil, false)
	skill.Name = "Brawling"
	skill.Points = fxp.Four
	e.Skills = append(e.Skills, skill)

	spell := NewSpell(e, nil, false)
	spell.Name = "Light"
	spell.Points = fxp.Four
	e.Spells = append(e.Spells, spell)

	equipment := NewEquipment(e, nil, false)
	equipment.Name = "Backpack"
	equipment.LocalNotes = "Roomy. <script>3 + 3</script>"
	equipmentModifier := NewEquipmentModifier(e, nil, false)
	equipmentModifier.Name = "Reinforced"
	equipmentModifier.LocalNotes = "Sturdy. <script>4 + 4</script>"
	equipment.Modifiers = []*EquipmentModifier{equipmentModifier}
	equipment.Weapons = []*Weapon{NewWeapon(equipment, true), NewWeapon(equipment, false)}
	e.CarriedEquipment = append(e.CarriedEquipment, equipment)

	note := NewNote(e, nil, false)
	note.MarkDown = "Remember. <script>5 + 5</script>"
	e.Notes = append(e.Notes, note)

	e.Recalculate()
	return e
}

// TestHashedFormHasNoDerivedValues verifies that nothing an entity holds publishes its derived values into a hash. Each
// kind of object decides that for itself, from the marker on the marshal, so this covers every one of them at once: a
// marshaler that ignored the marker would put values produced by running the data's scripts back into the hash, and a
// script that exceeds the permitted per-script execution time resolves to 0 rather than its real value, which would
// make an untouched document report itself as modified whenever the machine was briefly too busy.
func TestHashedFormHasNoDerivedValues(t *testing.T) {
	c := check.New(t)
	e := newPopulatedEntity()

	saved, err := jio.Marshal(e, json.Deterministic(true))
	c.NoError(err, "the entity marshals")
	c.True(strings.Contains(string(saved), `"calc":`), "the saved form still publishes its derived values")

	hashed := hashedJSON(t, e)
	c.False(strings.Contains(hashed, `"calc":`), "the hashed form has none of them: %s", hashed)

	// What is left must still be well-formed JSON, since it is what the hash is computed over.
	var round map[string]any
	c.NoError(jio.Unmarshal([]byte(hashed), &round), "the hashed form is valid JSON")

	// The hash must still notice an edit.
	before := Hash64(e)
	e.Profile.Name = "Changed"
	c.NotEqual(before, Hash64(e), "editing the entity changes its hash")
}

// TestTemplateAndLootHashesHaveNoDerivedValues covers the other documents whose whole contents are hashed to decide
// whether they have unsaved changes.
func TestTemplateAndLootHashesHaveNoDerivedValues(t *testing.T) {
	c := check.New(t)
	e := newPopulatedEntity()

	template := NewTemplate()
	template.Traits = e.Traits
	template.Skills = e.Skills
	template.Spells = e.Spells
	template.Equipment = e.CarriedEquipment
	template.Notes = e.Notes
	template.EnsureAttachments()
	c.False(strings.Contains(hashedJSON(t, template), `"calc":`), "a template's hashed form has no derived values")

	loot := NewLoot()
	loot.Equipment = e.CarriedEquipment
	loot.EnsureAttachments()
	c.False(strings.Contains(hashedJSON(t, loot), `"calc":`), "a loot sheet's hashed form has no derived values")
}

// TestHashingDoesNotRecalculate verifies that hashing an entity leaves it alone. Recalculating updates the
// "defaulted_from" of every skill, which is written to disk, from values the scripts in the data produce -- so if
// hashing recalculated, asking an untouched character sheet whether it had unsaved changes could be what gave it some.
func TestHashingDoesNotRecalculate(t *testing.T) {
	c := check.New(t)
	e := newPopulatedEntity()
	// A skill that defaults from an attribute whose own value comes from a script.
	e.Skills[0].Defaults = []*SkillDefault{{DefaultType: "per"}}
	e.Recalculate()
	c.NotNil(e.Skills[0].DefaultedFrom, "the test requires the skill to have defaulted from something")
	defaultedFrom := *e.Skills[0].DefaultedFrom
	before := Hash64(e)

	// Hash it while nothing can resolve, as it would be on a machine too busy for the scripts to finish in time.
	SuppressScriptResolveErrorLogging(func() {
		e.scriptResolvingDepth = maximumAllowedResolvingDepth
		defer func() { e.scriptResolvingDepth = 0 }()
		c.Equal(before, Hash64(e), "the hash must not change")
	})
	c.Equal(defaultedFrom, *e.Skills[0].DefaultedFrom, "and the entity must be untouched")
}

// TestRecalculateKeepsWhatAStoppedScriptCannotReplace verifies that recalculating an entity while its scripts are being
// stopped part way through leaves the values derived from them alone. Whether a script gets stopped depends on how busy
// the machine is, and the skill's "defaulted_from" is written to disk, so committing what was computed without the
// script's real result would put a value on disk that varies with the load on the machine that wrote it.
func TestRecalculateKeepsWhatAStoppedScriptCannotReplace(t *testing.T) {
	c := check.New(t)
	e := newPopulatedEntity()
	// A skill that defaults from an attribute whose own value comes from a script.
	e.Skills[0].Defaults = []*SkillDefault{{DefaultType: "per"}}
	e.Recalculate()
	c.NotNil(e.Skills[0].DefaultedFrom, "the test requires the skill to have defaulted from something")
	defaultedFrom := *e.Skills[0].DefaultedFrom
	level := e.Skills[0].LevelData
	spellLevel := e.Spells[0].LevelData
	c.NotEqual(fxp.Int(0), defaultedFrom.AdjLevel, "the test requires a non-zero value to be at risk")

	SuppressScriptResolveErrorLogging(func() {
		e.scriptResolvingDepth = maximumAllowedResolvingDepth
		defer func() { e.scriptResolvingDepth = 0 }()
		e.Recalculate()
	})

	c.Equal(defaultedFrom, *e.Skills[0].DefaultedFrom, "the skill's recorded default must be left alone")
	c.Equal(level, e.Skills[0].LevelData, "and so must its level")
	c.Equal(spellLevel, e.Spells[0].LevelData, "and the spell's level")

	// Once the scripts can finish again, a recalculation is free to update them as usual.
	e.Skills[0].Points = fxp.Twelve
	e.Recalculate()
	c.NotEqual(level, e.Skills[0].LevelData, "a recalculation that gets through its scripts still updates the level")
}

// TestCompletedScriptErrorsStillCount verifies that the guard above is confined to scripts that were stopped. A script
// that runs to completion and throws is wrong in the same way on every machine and every run, so its result is used as
// any other would be; declining it would leave a document that could never record what its own data says.
func TestCompletedScriptErrorsStillCount(t *testing.T) {
	c := check.New(t)
	e := newPopulatedEntity()
	e.Skills[0].Defaults = []*SkillDefault{{DefaultType: "per"}}
	e.Recalculate()
	before := e.Skills[0].DefaultedFrom.AdjLevel

	// "per" is defined in terms of IQ; make that definition throw every time it is evaluated.
	e.SheetSettings.Attributes.Set["iq"].Base = "nonexistent.reference"
	SuppressScriptResolveErrorLogging(func() { e.Recalculate() })

	c.NotEqual(before, e.Skills[0].DefaultedFrom.AdjLevel, "a failing script's result must still be recorded")
}

// TestSaveRecalculatesFirst verifies that saving brings the derived values the file publishes up to date, since
// marshaling an entity doesn't recalculate it.
func TestSaveRecalculatesFirst(t *testing.T) {
	c := check.New(t)
	e := newPopulatedEntity()
	e.Skills[0].Points = fxp.Twelve // Changed behind the sheet's back, so the recorded level is now stale.
	stale := e.Skills[0].LevelData.Level

	dir := t.TempDir()
	c.NoError(e.Save(filepath.Join(dir, "test"+SheetExt)), "the entity saves")

	c.NotEqual(stale, e.Skills[0].LevelData.Level, "saving must have brought the skill level up to date")
	saved, err := os.ReadFile(filepath.Join(dir, "test"+SheetExt))
	c.NoError(err, "the saved file reads back")
	c.True(strings.Contains(string(saved), `"level":`), "the saved file publishes the level it just computed")
}

// TestEditDataMarshalsWithoutDerivedValues verifies that the form an open editor compares -- the marshal of the data it
// was opened with against the marshal of the data it holds now -- carries none of the derived values and does not move
// when the scripts that produce them cannot resolve. The two copies an editor holds have separate script caches, so a
// script that blew its time budget while one of them was being marshaled but not the other used to be enough to make an
// editor with no edits in it claim unsaved changes.
func TestEditDataMarshalsWithoutDerivedValues(t *testing.T) {
	c := check.New(t)
	e := newPopulatedEntity()
	var data TraitEditData
	data.CopyFrom(e.Traits[len(e.Traits)-1]) // The one newPopulatedEntity adds, after those a new entity comes with.
	c.NotEqual(0, len(data.Weapons), "the test requires the edit data to hold weapons")
	c.NotEqual(0, len(data.Modifiers), "the test requires the edit data to hold modifiers")

	saved, err := jio.Marshal(&data, json.Deterministic(true))
	c.NoError(err, "the edit data marshals")
	c.True(strings.Contains(string(saved), `"calc":`), "an ordinary marshal of it still publishes derived values")

	compared, err := MarshalWithoutCalc(&data)
	c.NoError(err, "the edit data marshals without its derived values")
	c.False(strings.Contains(string(compared), `"calc":`), "the compared form has none of them: %s", compared)

	// Marshal it again while nothing can resolve, as it would be on a machine too busy for the scripts to finish in
	// time. The bytes the comparison is made over must be exactly the same ones.
	SuppressScriptResolveErrorLogging(func() {
		e.scriptResolvingDepth = maximumAllowedResolvingDepth
		defer func() { e.scriptResolvingDepth = 0 }()
		stopped, err2 := MarshalWithoutCalc(&data)
		c.NoError(err2, "the edit data marshals while its scripts are being stopped")
		c.Equal(string(compared), string(stopped), "a failure to resolve the scripts must not alter the bytes")
	})
}

// TestHashStartsWithSyncKind verifies that a node of each type hashes differently once its TID is given another kind,
// even though everything else it holds stays the same, except for plain equipment and a physical container, and a note
// and a note container, which hash the same so that one can be converted into the other.
func TestHashStartsWithSyncKind(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name string
		pair func() (Hashable, Hashable)
		same bool
	}{
		{name: "trait vs trait container", pair: func() (Hashable, Hashable) {
			trait := NewTrait(nil, nil, false)
			other := trait.Clone(LibraryFile{}, nil, nil, Copy)
			other.TID = tid.TID(kinds.TraitContainer) + trait.TID[1:]
			return trait, other
		}},
		{name: "trait modifier vs trait modifier container", pair: func() (Hashable, Hashable) {
			modifier := NewTraitModifier(nil, nil, false)
			other := modifier.Clone(LibraryFile{}, nil, nil, Copy)
			other.TID = tid.TID(kinds.TraitModifierContainer) + modifier.TID[1:]
			return modifier, other
		}},
		{name: "skill vs technique", pair: func() (Hashable, Hashable) {
			skill := NewSkill(nil, nil, false)
			other := skill.Clone(LibraryFile{}, nil, nil, Copy)
			other.TID = tid.TID(kinds.Technique) + skill.TID[1:]
			return skill, other
		}},
		{name: "skill vs skill container", pair: func() (Hashable, Hashable) {
			skill := NewSkill(nil, nil, false)
			other := skill.Clone(LibraryFile{}, nil, nil, Copy)
			other.TID = tid.TID(kinds.SkillContainer) + skill.TID[1:]
			return skill, other
		}},
		{name: "spell vs ritual magic spell", pair: func() (Hashable, Hashable) {
			spell := NewSpell(nil, nil, false)
			other := spell.Clone(LibraryFile{}, nil, nil, Copy)
			other.TID = tid.TID(kinds.RitualMagicSpell) + spell.TID[1:]
			return spell, other
		}},
		{name: "spell vs spell container", pair: func() (Hashable, Hashable) {
			spell := NewSpell(nil, nil, false)
			other := spell.Clone(LibraryFile{}, nil, nil, Copy)
			other.TID = tid.TID(kinds.SpellContainer) + spell.TID[1:]
			return spell, other
		}},
		{name: "equipment vs physical container", same: true, pair: func() (Hashable, Hashable) {
			eqp := NewEquipment(nil, nil, false)
			other := eqp.Clone(LibraryFile{}, nil, nil, Copy)
			other.TID = tid.TID(kinds.EquipmentContainer) + eqp.TID[1:]
			return eqp, other
		}},
		{name: "equipment vs group", pair: func() (Hashable, Hashable) {
			eqp := NewEquipment(nil, nil, false)
			other := eqp.Clone(LibraryFile{}, nil, nil, Copy)
			other.TID = tid.TID(kinds.EquipmentContainer) + eqp.TID[1:]
			other.ContainerType = eqcontainer.Group
			return eqp, other
		}},
		{name: "equipment modifier vs equipment modifier container", pair: func() (Hashable, Hashable) {
			modifier := NewEquipmentModifier(nil, nil, false)
			other := modifier.Clone(LibraryFile{}, nil, nil, Copy)
			other.TID = tid.TID(kinds.EquipmentModifierContainer) + modifier.TID[1:]
			return modifier, other
		}},
		{name: "note vs note container", same: true, pair: func() (Hashable, Hashable) {
			note := NewNote(nil, nil, false)
			other := note.Clone(LibraryFile{}, nil, nil, Copy)
			other.TID = tid.TID(kinds.NoteContainer) + note.TID[1:]
			return note, other
		}},
		{name: "conditional modifier vs conditional modifier container", pair: func() (Hashable, Hashable) {
			modifier := NewConditionalModifier("Source", "From", fxp.One)
			other := modifier.Clone(LibraryFile{}, nil, nil, Copy)
			other.TID = tid.TID(kinds.ConditionalModifierContainer) + modifier.TID[1:]
			return modifier, other
		}},
		{name: "reaction vs reaction container", pair: func() (Hashable, Hashable) {
			reaction := NewReaction("Source", "From", fxp.One)
			other := reaction.Clone(LibraryFile{}, nil, nil, Copy)
			other.TID = tid.TID(kinds.ReactionContainer) + reaction.TID[1:]
			return reaction, other
		}},
		{name: "reaction vs conditional modifier", pair: func() (Hashable, Hashable) {
			return NewReaction("Source", "From", fxp.One), NewConditionalModifier("Source", "From", fxp.One)
		}},
		{name: "melee vs ranged weapon", pair: func() (Hashable, Hashable) {
			melee := NewWeapon(nil, true)
			other := NewWeapon(nil, false)
			other.WeaponData = melee.WeaponData
			other.TID = tid.TID(kinds.WeaponRanged) + melee.TID[1:]
			return melee, other
		}},
	} {
		a, b := one.pair()
		if one.same {
			c.Equal(Hash64(a), Hash64(b), one.name+" hash the same")
		} else {
			c.NotEqual(Hash64(a), Hash64(b), one.name+" hash differently")
		}
	}
}
