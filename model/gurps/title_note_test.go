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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellcmp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// newTitleNoteFor returns a title note feature holding the given text.
func newTitleNoteFor(text string) *TitleNote {
	note := NewTitleNote()
	note.Text = text
	return note
}

// TestTitleNoteJSONHashAndDescribe checks how a title note feature is saved, hashed, cloned and described.
func TestTitleNoteJSONHashAndDescribe(t *testing.T) {
	c := check.New(t)
	who := newTitleNoteFor("@Who@")
	first := newTitleNoteFor("@Who@")
	first.BeforeBuiltIn = true
	blank := newTitleNoteFor("  ")
	c.Equal(who.Text, first.Text, "precondition: the two notes differ only in the flag")
	c.False(who.BeforeBuiltIn, "precondition: the who note goes after the built-in notes")
	c.Equal("  ", blank.Text, "precondition: the blank note is only spaces")
	checkFeaturesJSONRoundTrip(c, Features{who, first, NewTitleNote()})
	c.NotEqual(Hash64(who), Hash64(first), "the flag is source data and is hashed")
	c.NotEqual(Hash64(who), Hash64(blank), "the text is source data and is hashed")
	keys := make(map[string]string)
	who.FillWithNameableKeys(keys, nil)
	_, found := keys["Who"]
	c.True(found, "the text's nameable key is found")
	clone, ok := who.Clone().(*TitleNote)
	c.True(ok)
	if ok {
		clone.Text = "Changed"
		c.Equal("@Who@", who.Text, "a clone is a copy")
	}
	data, err := jio.Marshal(Features{who, first})
	c.NoError(err)
	c.Equal(1, strings.Count(string(data), `"before_built_in":true`), "the flag is written only when set")
	var restored Features
	c.NoError(jio.Unmarshal([]byte(`[{"type":"title_note","text":"Pistol","switchable":true}]`), &restored))
	c.Equal(1, len(restored))
	c.False(restored[0].IsSwitchable(), "a switchable flag in the data is ignored")

	em := func(s string) string { return "[" + s + "]" }
	bob := map[string]string{"Who": "Bob"}
	c.Equal("Adds the title note [Bob]", who.Describe(nil, bob, em))
	c.Equal("Adds the title note [@Who@]", who.Describe(nil, nil, em), "a marker without a value is left as it is")
	c.Equal("Adds the title note [Bob] before the built-in notes", first.Describe(nil, bob, em))
	c.Equal("Adds an empty title note", blank.Describe(nil, nil, em))
}

// TestTraitStringIncludesTitleNotes checks that a trait's name, leveled or not, is followed by its title notes, with
// their replacements applied and blank ones left out, and that a disabled trait keeps them.
func TestTraitStringIncludesTitleNotes(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	trait := addTraitWithFeatures(e, "Allies", newTitleNoteFor("@Ally@"), newTitleNoteFor("  "))
	trait.Replacements = map[string]string{"Ally": "Bob"}
	c.False(trait.IsLeveled(), "precondition: the trait isn't leveled")
	c.Equal("Allies (Bob)", trait.String())
	c.Equal("Allies (Bob)", trait.StringWithSavedCalc())
	trait.CanLevel = true
	trait.Levels = fxp.Two
	c.Equal("Allies 2 (Bob)", trait.String())
	c.Equal("Allies 2 (Bob)", trait.StringWithSavedCalc())
	trait.Disabled = true
	c.Equal("Allies 2 (Bob)", trait.String(), "a disabled trait keeps its own title notes")
}

// TestTitleNotesIgnoreTheSwitch checks that a title note can't be switchable and is shown whatever the switch.
func TestTitleNotesIgnoreTheSwitch(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	note := newTitleNoteFor("Always")
	note.SetSwitchable(true)
	c.False(note.IsSwitchable(), "a title note can't be made switchable")
	trait, skill, spell, eqp := newSwitchableItemSet(e, note)
	for _, item := range []interface {
		fmt.Stringer
		FeatureSwitcher
	}{trait, skill, spell, eqp} {
		c.False(item.HasSwitchableFeatures(), "a title note doesn't give %s a switch", item)
		item.SetSwitchedOn(false)
		off := item.String()
		item.SetSwitchedOn(true)
		c.Equal(off, item.String(), "the switch doesn't change the name")
		c.True(strings.HasSuffix(off, " (Always)"), "the title note is shown: %s", off)
	}
}

// TestPrereqsMatchTitleNotes checks that each prerequisite matches title notes, and only title notes.
func TestPrereqsMatchTitleNotes(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	allies := addTraitWithFeatures(e, "Allies", newTitleNoteFor("Bob"))
	guns := addTestSkill(e, "Guns", "Pistol", "", fxp.One)
	guns.Features = Features{newTitleNoteFor("Custom")}
	fireball := addTestSpell(e, "Fireball", fxp.One)
	fireball.Features = Features{newTitleNoteFor("Blue")}
	sword := addCarriedEquipmentWithFeatures(e, "Sword", newTitleNoteFor("Fine"))
	c.Equal("Allies (Bob)", allies.String(), "precondition: the trait has its title note")
	c.Equal("Guns (Pistol; Custom)", guns.String(), "precondition: the skill has its title note")
	c.Equal("Fireball (Blue)", fireball.String(), "precondition: the spell has its title note")
	c.True(sword.ReallyEquipped(), "precondition: the sword is equipped")
	c.Equal("Sword (Fine)", sword.String(), "precondition: the equipment has its title note")

	traitPrereq := NewTraitPrereq()
	traitPrereq.NameCriteria.Qualifier = "Allies"
	traitPrereq.LevelCriteria.Compare = criteria.AnyNumber
	traitPrereq.TitleNoteCriteria.Compare = criteria.IsText
	traitPrereq.TitleNoteCriteria.Qualifier = "Alice"
	var tooltip xbytes.InsertBuffer
	c.False(traitPrereq.Satisfied(e, nil, &tooltip, "", nil), "trait with no matching title note")
	c.Equal(`Has trait Allies with the title note Alice`, tooltip.String())

	skillPrereq := NewSkillPrereq()
	skillPrereq.NameCriteria.Qualifier = "Guns"
	skillPrereq.LevelCriteria.Compare = criteria.AnyNumber
	skillPrereq.TitleNoteCriteria.Compare = criteria.IsText
	skillPrereq.TitleNoteCriteria.Qualifier = "Custom"
	c.True(skillPrereq.Satisfied(e, nil, nil, "", nil), "skill matched by its title note")
	skillPrereq.TitleNoteCriteria.Qualifier = "Pistol"
	c.False(skillPrereq.Satisfied(e, nil, nil, "", nil), "the specialization is not a title note")

	spellPrereq := NewSpellPrereq()
	spellPrereq.SubType = spellcmp.TitleNote
	spellPrereq.SamePowerSource = false
	spellPrereq.QualifierCriteria.Qualifier = "Blue"
	c.True(spellPrereq.Satisfied(e, nil, nil, "", nil), "spell matched by its title note")
	spellPrereq.QualifierCriteria.Qualifier = "Red"
	c.False(spellPrereq.Satisfied(e, nil, nil, "", nil), "spell with no matching title note")
	spellPrereq.QualifierCriteria.Compare = criteria.IsNotText
	spellPrereq.QualifierCriteria.Qualifier = "Blue"
	c.False(spellPrereq.Satisfied(e, nil, nil, "", nil), "a not check must hold for every title note, as on traits")

	eqpPrereq := NewEquippedEquipmentPrereq()
	eqpPrereq.NameCriteria.Qualifier = "Sword"
	eqpPrereq.TitleNoteCriteria.Compare = criteria.IsText
	eqpPrereq.TitleNoteCriteria.Qualifier = "Fine"
	c.True(eqpPrereq.Satisfied(e, nil, nil, "", nil), "equipment matched by its title note")
	eqpPrereq.TitleNoteCriteria.Qualifier = "Cheap"
	c.False(eqpPrereq.Satisfied(e, nil, nil, "", nil), "equipment with no matching title note")
}

// TestPrereqDescribeTitleNotes checks how each prerequisite describes its title note criteria.
func TestPrereqDescribeTitleNotes(t *testing.T) {
	c := check.New(t)
	traitIs := NewTraitPrereq()
	traitIs.NameCriteria.Qualifier = "Allies"
	traitIs.TitleNoteCriteria.Compare = criteria.IsText
	traitIs.TitleNoteCriteria.Qualifier = "Bob"
	skillContains := NewSkillPrereq()
	skillContains.NameCriteria.Qualifier = "Guns"
	skillContains.TitleNoteCriteria.Compare = criteria.ContainsText
	skillContains.TitleNoteCriteria.Qualifier = "Cus"
	equippedNot := NewEquippedEquipmentPrereq()
	equippedNot.NameCriteria.Compare = criteria.AnyText
	equippedNot.TitleNoteCriteria.Compare = criteria.IsNotText
	equippedNot.TitleNoteCriteria.Qualifier = "Cheap"
	spellIs := NewSpellPrereq()
	spellIs.SubType = spellcmp.TitleNote
	spellIs.SamePowerSource = false
	spellIs.QualifierCriteria.Qualifier = "Blue"
	c.Equal(criteria.IsText, traitIs.TitleNoteCriteria.Compare, "precondition: the trait prereq checks is")
	c.Equal(criteria.ContainsText, skillContains.TitleNoteCriteria.Compare,
		"precondition: the skill prereq checks contains")
	c.Equal(criteria.IsNotText, equippedNot.TitleNoteCriteria.Compare,
		"precondition: the equipment prereq checks is not")
	c.Equal(spellcmp.TitleNote, spellIs.SubType, "precondition: the spell prereq checks title notes")

	em := func(s string) string { return "[" + s + "]" }
	c.Equal("Has trait [Allies] with the title note [Bob]", traitIs.Describe(nil, nil, em))
	traitIs.TitleNoteCriteria.Qualifier = "@Who@"
	c.Equal("Has trait [Allies] with the title note [Bob]",
		traitIs.Describe(nil, map[string]string{"Who": "Bob"}, em), "replacements are applied")
	c.Equal(`Has skill [Guns] with a title note that contains "[Cus]" at level at least 0`,
		skillContains.Describe(nil, nil, em))
	c.Equal(`Has equipped equipment with all title notes that are not "[Cheap]"`, equippedNot.Describe(nil, nil, em))
	c.Equal("Knows at least 1 spell whose title note is [Blue]", spellIs.Describe(nil, nil, em))

	traitIs.TitleNoteCriteria.Qualifier = " "
	c.Equal("Has trait [Allies] without title notes", traitIs.Describe(nil, nil, em),
		"an empty is matches only an item without title notes")
	equippedNot.TitleNoteCriteria.Qualifier = ""
	c.Equal("Has equipped equipment with at least one title note", equippedNot.Describe(nil, nil, em),
		"an empty is not matches any item with a title note")
}

// newTitleTraitModifier returns an enabled trait modifier with the given name that shows in its owner's title.
func newTitleTraitModifier(e *Entity, name string) *TraitModifier {
	mod := NewTraitModifier(e, nil, false)
	mod.Name = name
	mod.ShowInTitle = true
	return mod
}

// TestTraitModifiersInTheTitle checks that the modifiers shown in a trait's title follow its own title notes in
// modifier order, each named as it is shown and followed by its own title notes, and leave the trait's notes unless
// they have notes to show there.
func TestTraitModifiersInTheTitle(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	allies := addTraitWithFeatures(e, "Allies", newTitleNoteFor("@Who@"))
	allies.Replacements = map[string]string{"Who": "Bob", "Pct": "25%", "Empty": ""}
	pointTotal := newTitleTraitModifier(e, "Built on @Pct@")
	pointTotal.LocalNotes = "Rounded down"
	pointTotal.HideNotes = true
	frequency := newTitleTraitModifier(e, "12 or less")
	frequency.LocalNotes = "Usually at night"
	frequency.Features = Features{newTitleNoteFor("@Who@ again")}
	divisor := newTitleTraitModifier(e, "Armor Divisor")
	divisor.Levels = fxp.Two
	blank := newTitleTraitModifier(e, "@Empty@")
	unwilling := NewTraitModifier(e, nil, false)
	unwilling.Name = "Unwilling"
	off := newTitleTraitModifier(e, "Constantly")
	off.Disabled = true
	allies.SetModifiers([]*TraitModifier{pointTotal, frequency, divisor, blank, unwilling, off})
	c.True(pointTotal.ShowsInTitle() && pointTotal.HideNotes,
		"precondition: the point total modifier shows in the title and hides its notes")
	c.True(frequency.ShowsInTitle(), "precondition: the frequency modifier shows in the title")
	c.True(divisor.ShowsInTitle() && divisor.IsLeveled(), "precondition: the leveled divisor shows in the title")
	c.True(blank.ShowsInTitle(), "precondition: the modifier with a blank name shows in the title")
	c.False(unwilling.ShowsInTitle(), "precondition: the unwilling modifier stays in the notes")
	c.False(off.Enabled(), "precondition: the constantly modifier is disabled")

	c.Equal("Allies (Bob; Built on 25%; 12 or less; Bob again; Armor Divisor 2)", allies.String())
	c.Equal("12 or less (Usually at night); Unwilling", allies.ModifierNotes(),
		"a title modifier with notes still shows them, and one whose notes are hidden or empty is left out")

	p := NewTraitPrereq()
	p.NameCriteria.Qualifier = "Allies"
	p.LevelCriteria.Compare = criteria.AnyNumber
	p.TitleNoteCriteria.Compare = criteria.IsText
	p.TitleNoteCriteria.Qualifier = "12 or less"
	c.True(p.Satisfied(e, nil, nil, "", nil), "a modifier in the title counts as a title note")
	p.ModifierNameCriteria.Compare = criteria.IsText
	p.ModifierNameCriteria.Qualifier = "12 or less"
	c.True(p.Satisfied(e, nil, nil, "", nil), "and as a modifier, matching both checks at once")
	p.ModifierNameCriteria = criteria.Text{}
	p.TitleNoteCriteria.Qualifier = "Armor Divisor 2"
	c.True(p.Satisfied(e, nil, nil, "", nil), "a leveled modifier is matched as it is shown, with its level")
	p.TitleNoteCriteria.Qualifier = "Armor Divisor"
	c.False(p.Satisfied(e, nil, nil, "", nil), "and not without it")
	p.TitleNoteCriteria.Qualifier = "Unwilling"
	c.False(p.Satisfied(e, nil, nil, "", nil), "a modifier in the notes is not a title note")
}

// TestEquipmentModifiersInTheTitle checks the modifiers shown in a piece of equipment's title, including one in a
// group, and that a modifier's note placed before the built-in notes leads the equipment's own.
func TestEquipmentModifiersInTheTitle(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sword := addCarriedEquipmentWithFeatures(e, "Broadsword", newTitleNoteFor("Own"))
	sword.Replacements = map[string]string{"Feel": "Balanced"}
	first := newTitleNoteFor("First")
	first.BeforeBuiltIn = true
	fine := NewEquipmentModifier(e, nil, false)
	fine.Name = "Fine"
	fine.ShowInTitle = true
	fine.Features = Features{newTitleNoteFor("@Feel@"), first}
	silver := NewEquipmentModifier(e, nil, false)
	silver.Name = "Silvered"
	cheap := NewEquipmentModifier(e, nil, false)
	cheap.Name = "Cheap"
	cheap.ShowInTitle = true
	cheap.Disabled = true
	cheap.Features = Features{newTitleNoteFor("Brittle")}
	group := NewEquipmentModifier(e, nil, true)
	engraved := NewEquipmentModifier(e, group, false)
	engraved.Name = "Engraved"
	engraved.ShowInTitle = true
	group.Children = []*EquipmentModifier{engraved}
	sword.SetModifiers([]*EquipmentModifier{fine, silver, cheap, group})
	c.True(group.Container() && group.Enabled(), "precondition: the group is an enabled container")
	c.True(fine.ShowsInTitle(), "precondition: the fine modifier shows in the title")
	c.False(silver.ShowsInTitle(), "precondition: the silver modifier stays in the notes")
	c.False(cheap.Enabled(), "precondition: the cheap modifier is disabled")
	c.True(first.BeforeBuiltIn, "precondition: the first note goes before the built-in notes")
	c.Equal("Broadsword (First; Own; Fine; Balanced; Engraved)", sword.String(),
		"a disabled modifier adds nothing, while one in a group does")
	c.Equal("Silvered", sword.ModifierNotes())
}

// TestTitleNotesBeforeBuiltIn checks that the notes placed before the built-in notes lead the parentheses, ahead of a
// skill's specializations and a trait's self-control roll, for skills, spells and traits.
func TestTitleNotesBeforeBuiltIn(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	first := func() *TitleNote {
		note := newTitleNoteFor("First")
		note.BeforeBuiltIn = true
		return note
	}

	guns := addTestSkill(e, "Guns", "Pistol", "", fxp.One)
	guns.OptionalSpecialization = "Revolver"
	guns.Features = Features{newTitleNoteFor("Custom"), first()}
	c.Equal("Revolver", guns.OptionalSpecialization, "precondition: the skill has an optional specialization")
	c.Equal("Guns (First; Pistol; Revolver; Custom)", guns.String())
	c.Equal([]string{"First", "Custom"}, guns.TitleNotes(), "the notes placed first lead the list")

	fireball := addTestSpell(e, "Fireball", fxp.One)
	fireball.Features = Features{newTitleNoteFor("Late"), first()}
	c.Equal(2, len(fireball.Features), "precondition: the spell has both notes")
	c.Equal("Fireball (First; Late)", fireball.String())

	trait := addTraitWithFeatures(e, "Allies", newTitleNoteFor("Own"))
	mod := newTitleTraitModifier(e, "Frequency")
	mod.Features = Features{first()}
	trait.Modifiers = []*TraitModifier{mod}
	trait.SelfControl = selfctrl.CR12
	c.True(mod.ShowsInTitle(), "precondition: the frequency modifier shows in the title")
	var data CellData
	trait.CellData(TraitDescriptionColumn, &data)
	c.Equal("Allies (First; CR12; Own; Frequency)", data.Primary,
		"a modifier's note placed first goes before the self-control roll, and the modifier's name stays after")
	c.Equal("Allies (First; Own; Frequency)", trait.String())
}

// TestPrereqTitleNoteCriteriaHashJSONAndKeys checks that each prerequisite's title note criteria are hashed, saved and
// searched for nameable keys.
func TestPrereqTitleNoteCriteriaHashJSONAndKeys(t *testing.T) {
	c := check.New(t)
	trait := NewTraitPrereq()
	skill := NewSkillPrereq()
	eqp := NewEquippedEquipmentPrereq()
	for _, one := range []struct {
		prereq   Prereq
		criteria *criteria.Text
	}{{trait, &trait.TitleNoteCriteria}, {skill, &skill.TitleNoteCriteria}, {eqp, &eqp.TitleNoteCriteria}} {
		c.Equal(criteria.AnyText, one.criteria.Compare, "precondition: %T accepts any title note", one.prereq)
		hash := Hash64(one.prereq)
		one.criteria.Compare = criteria.IsText
		one.criteria.Qualifier = "@Who@"
		c.NotEqual(hash, Hash64(one.prereq), "the title note criteria is hashed")
		data, err := jio.Marshal(one.prereq)
		c.NoError(err)
		c.Contains(string(data), `"title_note":{"compare":"is","qualifier":"@Who@"}`)
		keys := make(map[string]string)
		one.prereq.FillWithNameableKeys(keys, nil)
		_, found := keys["Who"]
		c.True(found, "the title note criteria's nameable key is found")
	}
}

// TestWeaponsSelectedByOwnerTitleNote verifies that features selecting weapons by name match the owner's name without
// its title notes, and their owner title note criteria against each of the owner's title notes.
func TestWeaponsSelectedByOwnerTitleNote(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	selectors := addWeaponNameSelectors(e, "Broadsword")
	sword := addCarriedEquipmentWithFeatures(e, "Broadsword", newTitleNoteFor("Fine"), newTitleNoteFor("Balanced"))
	w := newCrushingWeapon(sword, "Swung")
	sword.Weapons = []*Weapon{w}
	e.Recalculate()
	c.Equal("Broadsword (Fine; Balanced)", w.String(), "precondition: the weapon's name has the title notes")
	selectors.check(c, e, w, true, "the name matches without the title notes")
	for _, one := range selectors.owners() {
		one.OwnerTitleNoteCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Balanced"}
	}
	selectors.check(c, e, w, true, "one of the title notes matches")
	c.Contains(selectors.damage.Describe(e, nil, func(s string) string { return s }),
		"weapons named Broadsword whose owner has the title note Balanced")
	for _, one := range selectors.owners() {
		one.OwnerTitleNoteCriteria.Qualifier = "Cheap"
	}
	selectors.check(c, e, w, false, "no title note matches")
	c.Equal(1, selectors.damage.specificity(), "the owner title note criterion counts toward specificity")
}

// TestSavedCalcTitleModifiers checks that a modifier shown in the title of an item loaded with the saved calc, which
// is never recalculated, still names itself with its owner's replacements.
func TestSavedCalcTitleModifiers(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	allies := addTraitWithFeatures(e, "Allies")
	allies.Replacements = map[string]string{"Who": "Bob"}
	allies.SetModifiers([]*TraitModifier{newTitleTraitModifier(e, "With @Who@")})
	sword := addCarriedEquipmentWithFeatures(e, "Sword")
	sword.Replacements = map[string]string{"Quality": "Fine"}
	fine := NewEquipmentModifier(e, nil, false)
	fine.Name = "@Quality@"
	fine.ShowInTitle = true
	sword.SetModifiers([]*EquipmentModifier{fine})
	c.Equal("Allies (With Bob)", allies.String(), "precondition: the trait's title names its modifier")
	c.Equal("Sword (Fine)", sword.String(), "precondition: the sword's title names its modifier")

	dir := t.TempDir()
	c.NoError(e.Save(filepath.Join(dir, "titles.gcs")))
	loaded, err := NewEntityFromFileWithSavedCalc(os.DirFS(dir), "titles.gcs")
	c.NoError(err)
	loadedAllies := loaded.Traits[len(loaded.Traits)-1]
	c.Equal("Allies", loadedAllies.Name, "precondition: the last trait is Allies")
	c.Equal("Allies (With Bob)", loadedAllies.StringWithSavedCalc())
	c.Equal("Sword (Fine)", loaded.CarriedEquipment[0].String())
}
