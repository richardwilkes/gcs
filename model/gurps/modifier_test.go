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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

func TestMergeReplacements(t *testing.T) {
	c := check.New(t)

	// A nil destination takes a copy of the source map, so that the two can then be changed independently.
	src := map[string]string{"Element": "Fire"}
	merged := mergeReplacements(nil, src)
	c.Equal(map[string]string{"Element": "Fire"}, merged)
	merged["Element"] = "Water"
	c.Equal("Fire", src["Element"], "the source map is not shared with the result")

	// Values the destination already holds win; new keys are added.
	dst := map[string]string{"Element": "Water"}
	c.Equal(map[string]string{"Element": "Water", "Target": "Orcs"},
		mergeReplacements(dst, map[string]string{"Element": "Fire", "Target": "Orcs"}))

	// An empty source leaves a nil destination nil rather than handing back an empty map.
	c.Nil(mergeReplacements(nil, map[string]string{}))
}

// TestModifierWithoutOwnerKeepsMarkersIntact verifies that a modifier not attached to a trait or piece of equipment
// displays its name and notes exactly as written, full marker syntax included, while an attached one resolves the
// markers against its owner's replacements.
func TestModifierWithoutOwnerKeepsMarkersIntact(t *testing.T) {
	c := check.New(t)
	const raw = "A @Element|Fire|Water@ spell"

	tm := NewTraitModifier(nil, nil, false)
	tm.Name = raw
	tm.LocalNotes = raw
	c.Equal(raw, tm.NameWithReplacements(), "an unattached trait modifier keeps its markers as written")
	c.Equal(raw, tm.LocalNotesWithReplacements(), "an unattached trait modifier keeps its markers as written")

	em := NewEquipmentModifier(nil, nil, false)
	em.Name = raw
	em.LocalNotes = raw
	c.Equal(raw, em.NameWithReplacements(), "an unattached equipment modifier keeps its markers as written")
	c.Equal(raw, em.LocalNotesWithReplacements(), "an unattached equipment modifier keeps its markers as written")

	// Once attached, the owner's replacements are applied and anything still unresolved is shown in compact form.
	trait := NewTrait(nil, nil, false)
	tm.SetTarget(trait)
	c.Equal("A @Element@ spell", tm.NameWithReplacements(), "an owner without replacements yields the compact form")
	trait.Replacements = map[string]string{"Element|Fire|Water": "Fire"}
	c.Equal("A Fire spell", tm.NameWithReplacements())
	c.Equal("A Fire spell", tm.LocalNotesWithReplacements())

	equipment := NewEquipment(nil, nil, false)
	em.SetTarget(equipment)
	c.Equal("A @Element@ spell", em.NameWithReplacements(), "an owner without replacements yields the compact form")
	equipment.Replacements = map[string]string{"Element|Fire|Water": "Water"}
	c.Equal("A Water spell", em.NameWithReplacements())
	c.Equal("A Water spell", em.LocalNotesWithReplacements())
}

// TestModifierContainersContributeNameableKeys verifies that container modifiers take part in nameable key
// collection, since their names and notes are displayed with replacements applied, while disabled leaf modifiers
// still do not.
func TestModifierContainersContributeNameableKeys(t *testing.T) {
	c := check.New(t)

	trait := NewTrait(nil, nil, false)
	trait.Name = "Ally"
	group := NewTraitModifier(nil, nil, true)
	group.Name = "@Group@ Options"
	group.LocalNotes = "Applies to @Target@"
	leaf := NewTraitModifier(nil, group, false)
	leaf.Name = "@Leaf@"
	off := NewTraitModifier(nil, group, false)
	off.Name = "@Off@"
	off.SetDisabled(true)
	group.Children = []*TraitModifier{leaf, off}
	trait.Modifiers = []*TraitModifier{group}
	keys := make(map[string]string)
	trait.FillWithNameableKeys(keys, nil)
	c.Equal(map[string]string{
		"Group":  nameable.Unset,
		"Target": nameable.Unset,
		"Leaf":   nameable.Unset,
	}, keys)

	equipment := NewEquipment(nil, nil, false)
	equipment.Name = "Sword"
	egroup := NewEquipmentModifier(nil, nil, true)
	egroup.Name = "@Group@ Options"
	egroup.LocalNotes = "Applies to @Target@"
	eleaf := NewEquipmentModifier(nil, egroup, false)
	eleaf.Name = "@Leaf@"
	eoff := NewEquipmentModifier(nil, egroup, false)
	eoff.Name = "@Off@"
	eoff.SetDisabled(true)
	egroup.Children = []*EquipmentModifier{eleaf, eoff}
	equipment.Modifiers = []*EquipmentModifier{egroup}
	keys = make(map[string]string)
	equipment.FillWithNameableKeys(keys, nil)
	c.Equal(map[string]string{
		"Group":  nameable.Unset,
		"Target": nameable.Unset,
		"Leaf":   nameable.Unset,
	}, keys)
}

// TestModifierSyncWithSource verifies that a modifier which has drifted from its library source pulls the synced
// fields back across, whether or not it has a data owner, and that containers do not pick up leaf-only data. The
// weapons it pulls across are its own copies, belonging to its trait and held by it.
func TestModifierSyncWithSource(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	libFile := LibraryFile{Library: "Test Library", Path: "Test" + TraitModifiersExt}

	// A leaf trait modifier whose local copy has drifted in name and cost, and lacks the source's weapon.
	source := NewTraitModifier(nil, nil, false)
	source.Name = "Reduced Time"
	source.CostAdj = "+20%"
	source.Tags = []string{"Enhancement"}
	source.Weapons = []*Weapon{NewWeapon(nil, true)}
	local := NewTraitModifier(e, nil, false)
	local.Name = "Reduced Time (old)"
	local.CostAdj = "+10%"
	local.Source = Source{LibraryFile: libFile, TID: source.TID}
	trait := NewTrait(e, nil, false)
	trait.AddModifiers(local)
	stubLibrarySources(t, e.SourceMatcher(), libFile, source)
	state, _ := e.SourceMatcher().Match(local)
	c.Equal(srcstate.Mismatched, state, "precondition: the local copy differs from its source")
	local.SyncWithSource()
	c.Equal("Reduced Time", local.Name)
	c.Equal("+20%", local.CostAdj)
	c.Equal([]string{"Enhancement"}, local.Tags)
	c.Equal(1, len(local.Weapons), "the source's weapon is synced")
	if len(local.Weapons) == 1 {
		w := local.Weapons[0]
		c.True(w != source.Weapons[0], "the synced weapon is a copy, not the source's")
		c.True(w.Owner == trait, "the synced weapon belongs to the local modifier's trait")
		c.True(w.HoldingModifier() == local, "the synced weapon is held by the local modifier")
	}
	source.Tags[0] = "changed"
	c.Equal([]string{"Enhancement"}, local.Tags, "synced tags must not alias the source's slice")
	state, _ = e.SourceMatcher().Match(local)
	c.Equal(srcstate.Matched, state, "after syncing, the local copy matches its source")

	// Without a data owner, the source is looked up in the shared matcher.
	isolateUnownedSrcMatcher(t)
	orphan := NewTraitModifier(nil, nil, false)
	orphan.Name = "Reduced Time (old)"
	orphan.Source = Source{LibraryFile: libFile, TID: source.TID}
	orphan.SyncWithSource()
	c.Equal("Reduced Time (old)", orphan.Name,
		"what the entity's matcher holds is no use to a modifier without a data owner")
	stubLibrarySources(t, &unownedSrcMatcher, libFile, source)
	orphan.SyncWithSource()
	c.Equal("Reduced Time", orphan.Name, "a modifier without a data owner is synced with what the shared matcher holds")

	// A container only picks up the common data.
	sourceContainer := NewEquipmentModifier(nil, nil, true)
	sourceContainer.Name = "Options"
	sourceContainer.PageRef = "B1"
	localContainer := NewEquipmentModifier(e, nil, true)
	localContainer.Name = "Options (old)"
	localContainer.Source = Source{LibraryFile: libFile, TID: sourceContainer.TID}
	stubLibrarySources(t, e.SourceMatcher(), libFile, sourceContainer)
	state, _ = e.SourceMatcher().Match(localContainer)
	c.Equal(srcstate.Mismatched, state, "precondition: the local container differs from its source")
	localContainer.SyncWithSource()
	c.Equal("Options", localContainer.Name)
	c.Equal("B1", localContainer.PageRef)
	c.Equal("", localContainer.CostAmount, "a container carries no leaf-only data")
}

// TestModifierHideNotes checks that a modifier with HideNotes set is listed in its owner's notes without its notes,
// while its full description and its weapon usage notes keep them.
func TestModifierHideNotes(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	allies := addTraitWithFeatures(e, "Allies")
	unwilling := NewTraitModifier(e, nil, false)
	unwilling.Name = "Unwilling"
	unwilling.LocalNotes = "Must be persuaded"
	unwilling.HideNotes = true
	unwilling.ShowNotesOnWeapon = true
	summonable := NewTraitModifier(e, nil, false)
	summonable.Name = "Summonable"
	summonable.LocalNotes = "Takes a turn"
	allies.SetModifiers([]*TraitModifier{unwilling, summonable})
	allies.Weapons = []*Weapon{NewWeapon(allies, true)}
	c.True(unwilling.Enabled(), "precondition: the unwilling modifier is enabled")
	c.True(summonable.Enabled(), "precondition: the summonable modifier is enabled")

	c.Equal("Unwilling; Summonable (Takes a turn)", allies.ModifierNotes(),
		"hidden notes are left out, while the modifier is still listed")
	c.Equal("Unwilling (Must be persuaded)", unwilling.FullDescription(), "the full description keeps hidden notes")
	c.Equal("Must be persuaded", allies.Weapons[0].Notes(), "weapon usage notes ignore HideNotes")

	sword := addCarriedEquipmentWithFeatures(e, "Broadsword")
	silvered := NewEquipmentModifier(e, nil, false)
	silvered.Name = "Silvered"
	silvered.LocalNotes = "Hurts werewolves"
	silvered.HideNotes = true
	sword.SetModifiers([]*EquipmentModifier{silvered})
	c.True(silvered.Enabled(), "precondition: the silvered modifier is enabled")

	c.Equal("Silvered", sword.ModifierNotes(), "equipment modifiers hide their notes too")
	c.Equal("Silvered (Hurts werewolves)", silvered.FullDescription(), "the full description keeps hidden notes")
}

// TestModifierHideNotesRoundTripAndHash checks that HideNotes and ShowInTitle on both kinds of modifier, and
// CostIgnoresLevel on a trait modifier, are each saved only when set, so that existing files re-save unchanged, load
// back, and are hashed, so that a library row differing only in one is reported as mismatched.
func TestModifierHideNotesRoundTripAndHash(t *testing.T) {
	c := check.New(t)
	traitMod := NewTraitModifier(nil, nil, false)
	traitMod.Name = "Unwilling"
	equipmentMod := NewEquipmentModifier(nil, nil, false)
	equipmentMod.Name = "Silvered"
	for _, one := range []struct {
		name string
		key  string
		set  func(bool)
		node Hashable
	}{
		{"trait modifier", "hide_notes", func(on bool) { traitMod.HideNotes = on }, traitMod},
		{"trait modifier", "show_in_title", func(on bool) { traitMod.ShowInTitle = on }, traitMod},
		{"trait modifier", "cost_ignores_level", func(on bool) { traitMod.CostIgnoresLevel = on }, traitMod},
		{"equipment modifier", "hide_notes", func(on bool) { equipmentMod.HideNotes = on }, equipmentMod},
		{"equipment modifier", "show_in_title", func(on bool) { equipmentMod.ShowInTitle = on }, equipmentMod},
	} {
		data, err := jio.Marshal(one.node)
		c.NoError(err)
		c.NotContains(string(data), one.key, "%s: an unset %s isn't saved", one.name, one.key)
		hash := Hash64(one.node)
		one.set(true)
		c.NotEqual(hash, Hash64(one.node), "%s: %s is source data and is hashed", one.name, one.key)
		data, err = jio.Marshal(one.node)
		c.NoError(err)
		c.Contains(string(data), `"`+one.key+`":true`, "%s", one.name)
	}

	data, err := jio.Marshal(traitMod)
	c.NoError(err)
	var restoredTraitMod TraitModifier
	c.NoError(jio.Unmarshal(data, &restoredTraitMod))
	c.True(restoredTraitMod.HideNotes && restoredTraitMod.ShowInTitle && restoredTraitMod.CostIgnoresLevel,
		"the trait modifier's flags load back")
	data, err = jio.Marshal(equipmentMod)
	c.NoError(err)
	var restoredEquipmentMod EquipmentModifier
	c.NoError(jio.Unmarshal(data, &restoredEquipmentMod))
	c.True(restoredEquipmentMod.HideNotes && restoredEquipmentMod.ShowInTitle,
		"the equipment modifier's flags load back")
}

// linkedTrait returns a trait whose own text uses @style@, with a modifier whose @focus@ is linked to it and whose
// @extra@ is not.
func linkedTrait() (*Trait, *TraitModifier) {
	trait := NewTrait(nil, nil, false)
	trait.Name = "Trained in @style@"
	mod := NewTraitModifier(nil, nil, false)
	mod.Name = "@focus@ focus"
	mod.LocalNotes = "with @extra@"
	trait.Modifiers = []*TraitModifier{mod}
	mod.SetTarget(trait)
	trait.NameableLinks = map[tid.TID]map[string]string{mod.TID: {"focus": "style"}}
	return trait, mod
}

// TestLinkedModifierCodeIsAskedAsOwnerCode verifies that a modifier's linked code is asked for, answered and kept as
// its owner's code, and that its linked and unlinked codes both resolve through the owner's replacements, in the text
// of the modifier's own weapon as in the rest of it.
func TestLinkedModifierCodeIsAskedAsOwnerCode(t *testing.T) {
	c := check.New(t)
	trait, mod := linkedTrait()
	w := NewWeapon(nil, true)
	w.Usage = "@focus@ strike"
	w.UsageNotes = "@stance@"
	mod.Weapons = []*Weapon{w}
	mod.SetTarget(trait)
	trait.Replacements = map[string]string{"style": "Judo"}
	m := make(map[string]string)
	trait.FillWithNameableKeys(m, nil)
	c.Equal(map[string]string{"style": "Judo", "extra": nameable.Unset, "stance": nameable.Unset}, m,
		"the linked code is asked as the trait's, and the weapon's codes are asked too")

	// An answer given for the trait's code is the one the linked code shows.
	m["style"] = "Karate"
	m["extra"] = "zeal"
	m["stance"] = "Horse stance"
	trait.ApplyNameableKeys(m)
	c.Equal("Karate focus", mod.NameWithReplacements())
	c.Equal("with zeal", mod.LocalNotesWithReplacements(), "an unlinked code resolves through the owner too")
	c.Equal("Karate strike", w.UsageWithReplacements(), "the modifier's weapon resolves the linked code as it does")
	c.Equal("Horse stance", w.UsageNotesWithReplacements(), "and an unlinked one")
	c.Equal(map[string]string{"style": "Karate", "extra": "zeal", "stance": "Horse stance"}, trait.Replacements,
		"nothing is kept under the modifier's own code")
}

func TestCloneRekeysNameableLinks(t *testing.T) {
	c := check.New(t)
	trait, mod := linkedTrait()
	trait.Replacements = map[string]string{"style": "Judo"}
	clone := trait.Clone(LibraryFile{}, nil, nil, Reference)
	c.Equal(1, len(clone.Modifiers))
	if len(clone.Modifiers) != 1 {
		return
	}
	cloned := clone.Modifiers[0]
	c.NotEqual(mod.TID, cloned.TID, "a reference copy gets a fresh ID")
	c.Equal(map[tid.TID]map[string]string{cloned.TID: {"focus": "style"}}, clone.NameableLinks)
	c.Equal("Judo focus", cloned.NameWithReplacements())

	clone.NameableLinks[cloned.TID]["focus"] = "other"
	c.Equal("style", trait.NameableLinks[mod.TID]["focus"], "the original's links are not shared")
}

func TestLinkedEquipmentModifierCode(t *testing.T) {
	c := check.New(t)
	eqp := NewEquipment(nil, nil, false)
	eqp.Name = "Sword of @metal@"
	mod := NewEquipmentModifier(nil, nil, false)
	mod.Name = "@material@ inlay"
	eqp.SetModifiers([]*EquipmentModifier{mod})
	eqp.NameableLinks = map[tid.TID]map[string]string{mod.TID: {"material": "metal"}}
	eqp.Replacements = map[string]string{"metal": "Silver"}
	c.Equal("Silver inlay", mod.NameWithReplacements())
	m := make(map[string]string)
	eqp.FillWithNameableKeys(m, nil)
	c.Equal(map[string]string{"metal": "Silver"}, m, "the linked code is asked as the equipment's")
}

func TestLinkedModifierCodeReachesItsFeatures(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	skill := NewSkill(e, nil, false)
	skill.Name = "Broadsword"
	skill.Points = fxp.Four
	e.Skills = append(e.Skills, skill)

	// A trait modifier whose skill bonus names its skill through a code linked to the trait's, and whose bonuses to the
	// trait's weapons name their usage through it.
	traitMod := NewTraitModifier(e, nil, false)
	traitMod.Name = "Focused on @skill@"
	usage := criteria.Text{Compare: criteria.IsText, Qualifier: "@skill@"}
	weaponBonus := NewWeaponBonus(feature.WeaponBonus)
	weaponBonus.SelectionType = wsel.ThisWeapon
	weaponBonus.SpecializationCriteria = usage
	weaponSkillBonus := NewSkillBonus()
	weaponSkillBonus.SelectionType = skillsel.ThisWeapon
	weaponSkillBonus.SpecializationCriteria = usage
	weaponSkillBonus.Amount = fxp.Two
	traitMod.Features = Features{newSkillBonusTo("@skill@", fxp.Two), weaponBonus, weaponSkillBonus}
	trait := addTraitWithFeatures(e, "Weapon Master (@weapon@)")
	trait.SetModifiers([]*TraitModifier{traitMod})
	trait.NameableLinks = map[tid.TID]map[string]string{traitMod.TID: {"skill": "weapon"}}
	trait.Replacements = map[string]string{"weapon": "Broadsword"}
	w := newCrushingWeapon(trait, "@weapon@")
	w.Defaults = []*SkillDefault{newSkillDefaultTo("Broadsword", "", true, 0)}
	trait.Weapons = []*Weapon{w}

	// Likewise for an equipment modifier.
	eqpMod := NewEquipmentModifier(e, nil, false)
	eqpMod.Name = "Balanced for @use@"
	eqpMod.Features = Features{newSkillBonusTo("@use@", fxp.One)}
	eqp := addCarriedEquipmentWithFeatures(e, "Sword of @skill@")
	eqp.SetModifiers([]*EquipmentModifier{eqpMod})
	eqp.NameableLinks = map[tid.TID]map[string]string{eqpMod.TID: {"use": "skill"}}
	eqp.Replacements = map[string]string{"skill": "Broadsword"}
	eqp.Equipped = true

	e.Recalculate()
	c.Equal("Focused on Broadsword", traitMod.NameWithReplacements())
	c.Equal("Balanced for Broadsword", eqpMod.NameWithReplacements())
	c.Equal(fxp.Three, e.SkillBonusFor("Broadsword", "", "", nil, nil),
		"a modifier's bonus resolves its linked code through its owner, as its name does")
	c.True(slices.Contains(w.collectWeaponBonuses(oneDieCount, nil, feature.WeaponBonus), weaponBonus),
		"so does a modifier's weapon bonus to the weapons of its owner")
	c.Equal(skill.LevelData.Level+fxp.Two, w.SkillLevel(nil), "and its skill bonus to those weapons")
}

func TestApplyDropsSelfLinks(t *testing.T) {
	c := check.New(t)
	trait, mod := linkedTrait()
	trait.NameableLinks[mod.TID]["extra"] = "style"
	var edit TraitEditData
	edit.CopyFrom(trait)
	// Renaming the trait's code to the modifier's in the editor leaves the link pointing at itself.
	edit.NameableLinks[mod.TID]["focus"] = "focus"
	edit.ApplyTo(trait)
	c.Equal(map[tid.TID]map[string]string{mod.TID: {"extra": "style"}}, trait.NameableLinks,
		"a link of a code to itself means the same as none, so it is not kept")

	edit.CopyFrom(trait)
	edit.NameableLinks[mod.TID]["extra"] = "extra"
	edit.ApplyTo(trait)
	c.Nil(trait.NameableLinks, "a modifier left with no links has no entry")
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
