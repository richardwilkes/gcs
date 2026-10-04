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

func TestTitleNoteJSONRoundTrip(t *testing.T) {
	c := check.New(t)
	checkFeaturesJSONRoundTrip(c, Features{newTitleNoteFor("Pistol"), NewTitleNote()})
	var restored Features
	c.NoError(jio.Unmarshal([]byte(`[{"type":"title_note","text":"Pistol","switchable":true}]`), &restored))
	c.Equal(1, len(restored))
	c.False(restored[0].IsSwitchable(), "a switchable flag in the data is ignored")
}

func TestTraitStringIncludesTitleNotes(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	trait := addTraitWithFeatures(e, "Allies", newTitleNoteFor("@Ally@"), newTitleNoteFor("  "))
	trait.Replacements = map[string]string{"Ally": "Bob"}
	enabled := NewTraitModifier(e, nil, false)
	enabled.Name = "Frequency"
	enabled.Features = Features{newTitleNoteFor("15 or less")}
	disabled := NewTraitModifier(e, nil, false)
	disabled.Name = "Unused"
	disabled.Disabled = true
	disabled.Features = Features{newTitleNoteFor("Never shown")}
	trait.Modifiers = []*TraitModifier{enabled, disabled}
	c.True(enabled.Enabled(), "precondition: the first modifier is enabled")
	c.False(disabled.Enabled(), "precondition: the second modifier is disabled")

	c.Equal("Allies (Bob; 15 or less)", trait.String())

	var data CellData
	trait.SelfControl = selfctrl.CR12
	trait.CellData(TraitDescriptionColumn, &data)
	c.Equal("Allies (CR12; Bob; 15 or less)", data.Primary,
		"the self-control roll shares the parentheses with the title notes")
}

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

func TestSkillStringJoinsSpecializationsAndNotesWithSemicolons(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	guns := addTestSkill(e, "Guns", "Pistol", "", fxp.One)
	guns.OptionalSpecialization = "Revolver"
	guns.Features = Features{newTitleNoteFor("Custom")}
	c.Equal("Guns (Pistol; Revolver; Custom)", guns.String())
}

func TestEquipmentModifierTitleNotes(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	eqp := addCarriedEquipmentWithFeatures(e, "Sword")
	enabled := NewEquipmentModifier(e, nil, false)
	enabled.Features = Features{newTitleNoteFor("Fine")}
	disabled := NewEquipmentModifier(e, nil, false)
	disabled.Disabled = true
	disabled.Features = Features{newTitleNoteFor("Cheap")}
	eqp.Modifiers = []*EquipmentModifier{enabled, disabled}
	c.True(enabled.Enabled(), "precondition: the first modifier is enabled")
	c.False(disabled.Enabled(), "precondition: the second modifier is disabled")
	c.Equal("Sword (Fine)", eqp.String())
}

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
	traitPrereq.TitleNoteCriteria.Qualifier = "Alice, Bob"
	c.True(traitPrereq.Satisfied(e, nil, nil, "", nil), "trait matched by one of the listed title notes")
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

	eqpPrereq := NewEquippedEquipmentPrereq()
	eqpPrereq.NameCriteria.Qualifier = "Sword"
	eqpPrereq.TitleNoteCriteria.Compare = criteria.IsText
	eqpPrereq.TitleNoteCriteria.Qualifier = "Fine"
	c.True(eqpPrereq.Satisfied(e, nil, nil, "", nil), "equipment matched by its title note")
	eqpPrereq.TitleNoteCriteria.Qualifier = "Cheap"
	c.False(eqpPrereq.Satisfied(e, nil, nil, "", nil), "equipment with no matching title note")
}

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
	c.Equal(criteria.IsText, traitIs.TitleNoteCriteria.Compare, "precondition: the trait prereq matches by is")
	c.Equal(criteria.ContainsText, skillContains.TitleNoteCriteria.Compare, "precondition: the skill prereq contains")
	c.Equal(criteria.IsNotText, equippedNot.TitleNoteCriteria.Compare, "precondition: the equipment prereq is not")
	c.Equal(spellcmp.TitleNote, spellIs.SubType, "precondition: the spell prereq matches by title note")

	em := func(s string) string { return "[" + s + "]" }
	c.Equal("Has trait [Allies] with the title note [Bob]", traitIs.Describe(nil, nil, em))
	c.Equal(`Has skill [Guns] with a title note that contains "[Cus]" at level at least 0`,
		skillContains.Describe(nil, nil, em))
	c.Equal(`Has equipped equipment with all title notes that are not "[Cheap]"`, equippedNot.Describe(nil, nil, em))
	c.Equal("Knows at least 1 spell whose title note is [Blue]", spellIs.Describe(nil, nil, em))
}

func TestTraitStringWithSavedCalcIncludesTitleNotes(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	allies := addTraitWithFeatures(e, "Allies", newTitleNoteFor("Bob"))
	c.Equal("Allies (Bob)", allies.String(), "precondition: the trait has its title note")
	c.Equal("Allies (Bob)", allies.StringWithSavedCalc())
}

func TestTitleNoteCriteriaComparisons(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	allies := addTraitWithFeatures(e, "Allies", newTitleNoteFor("Bob"), newTitleNoteFor("15 or less"))
	c.Equal("Allies (Bob; 15 or less)", allies.String(), "precondition: the trait has both title notes")
	for _, one := range []struct {
		compare   criteria.StringComparison
		qualifier string
		expected  bool
	}{
		{criteria.IsText, "bob", true},
		{criteria.IsText, "Alice, Bob", true},
		{criteria.IsText, "Alice", false},
		{criteria.IsNotText, "Alice", true},
		{criteria.IsNotText, "Alice, Bob", false},
		{criteria.ContainsText, "or le", true},
		{criteria.ContainsText, "Ali", false},
		{criteria.DoesNotContainText, "Ali", true},
		{criteria.DoesNotContainText, "ob", false},
		{criteria.StartsWithText, "15", true},
		{criteria.StartsWithText, "less", false},
		{criteria.DoesNotStartWithText, "Ali", true},
		{criteria.DoesNotStartWithText, "B", false},
		{criteria.EndsWithText, "less", true},
		{criteria.EndsWithText, "15", false},
		{criteria.DoesNotEndWithText, "ce", true},
		{criteria.DoesNotEndWithText, "ob", false},
	} {
		p := NewTraitPrereq()
		p.NameCriteria.Qualifier = "Allies"
		p.LevelCriteria.Compare = criteria.AnyNumber
		p.TitleNoteCriteria.Compare = one.compare
		p.TitleNoteCriteria.Qualifier = one.qualifier
		c.Equal(one.expected, p.Satisfied(e, nil, nil, "", nil), "%s %q", one.compare.Key(), one.qualifier)
	}
}

// newTitleTraitModifier returns an enabled trait modifier with the given name and short name that shows in its owner's
// title.
func newTitleTraitModifier(e *Entity, name, shortName string) *TraitModifier {
	mod := NewTraitModifier(e, nil, false)
	mod.Name = name
	mod.ShortName = shortName
	mod.ShowInTitle = true
	return mod
}

func TestTraitModifiersInTheTitle(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	allies := addTraitWithFeatures(e, "Allies", newTitleNoteFor("@Who@"))
	allies.Replacements = map[string]string{"Who": "Bob", "Pct": "25%"}
	pointTotal := newTitleTraitModifier(e, "25% of your starting points", "Built on @Pct@")
	frequency := newTitleTraitModifier(e, "Appears quite often (12-)", "12 or less")
	frequency.Features = Features{newTitleNoteFor("Loyal")}
	unnamed := newTitleTraitModifier(e, "Summonable", "")
	unwilling := NewTraitModifier(e, nil, false)
	unwilling.Name = "Unwilling"
	unwilling.ShortName = "Grudging"
	off := newTitleTraitModifier(e, "Appears constantly", "Constantly")
	off.Disabled = true
	allies.SetModifiers([]*TraitModifier{pointTotal, frequency, unnamed, unwilling, off})
	c.True(pointTotal.ShowsInTitle(), "precondition: the point total modifier shows in the title")
	c.True(frequency.ShowsInTitle(), "precondition: the frequency modifier shows in the title")
	c.True(unnamed.ShowsInTitle(), "precondition: the modifier without a short name shows in the title")
	c.False(unwilling.ShowsInTitle(), "precondition: the unwilling modifier stays in the notes")
	c.False(off.Enabled(), "precondition: the constantly modifier is disabled")

	c.Equal("Allies (Bob; Built on 25%; 12 or less; Loyal; Summonable)", allies.String(),
		"title modifiers follow the trait's own title notes in modifier order, each before its own title notes")
	c.Equal("Grudging", allies.ModifierNotes(), "the notes keep only the modifier not shown in the title, by its short name")
	c.Equal("Unwilling", unwilling.FullDescription(), "the full description keeps the full name")

	p := NewTraitPrereq()
	p.NameCriteria.Qualifier = "Allies"
	p.LevelCriteria.Compare = criteria.AnyNumber
	p.TitleNoteCriteria.Compare = criteria.IsText
	p.TitleNoteCriteria.Qualifier = "12 or less"
	c.True(p.Satisfied(e, nil, nil, "", nil), "a modifier in the title counts as a title note")
	p.TitleNoteCriteria.Qualifier = "Grudging"
	c.False(p.Satisfied(e, nil, nil, "", nil), "a modifier in the notes does not")
}

func TestLeveledTraitModifierInTheTitle(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	trait := addTraitWithFeatures(e, "Innate Attack")
	mod := newTitleTraitModifier(e, "Armor Divisor", "AD")
	mod.Levels = fxp.Two
	trait.SetModifiers([]*TraitModifier{mod})
	c.True(mod.IsLeveled(), "precondition: the modifier is leveled")
	c.Equal("AD 2", mod.CompactName())
	c.Equal("Armor Divisor 2", mod.String(), "String keeps the full name")
	c.Equal("Innate Attack (AD 2)", trait.String())
}

func TestEquipmentModifiersInTheTitle(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sword := addCarriedEquipmentWithFeatures(e, "Broadsword")
	fine := NewEquipmentModifier(e, nil, false)
	fine.Name = "Fine Quality"
	fine.ShortName = "Fine"
	fine.ShowInTitle = true
	silver := NewEquipmentModifier(e, nil, false)
	silver.Name = "Silver Coated"
	silver.ShortName = "Silvered"
	sword.SetModifiers([]*EquipmentModifier{fine, silver})
	c.True(fine.ShowsInTitle(), "precondition: the fine modifier shows in the title")
	c.False(silver.ShowsInTitle(), "precondition: the silver modifier stays in the notes")
	c.Equal("Broadsword (Fine)", sword.String())
	c.Equal("Silvered", sword.ModifierNotes())
}

func TestModifierTitleFieldsRoundTripAndHash(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	plain := NewTraitModifier(e, nil, false)
	plain.Name = "Appears quite often (12-)"
	titled := NewTraitModifier(e, nil, false)
	titled.Name = plain.Name
	titled.TID = plain.TID
	titled.ShortName = "12 or less"
	titled.ShowInTitle = true
	c.Equal(plain.Name, titled.Name, "precondition: the two modifiers differ only in the new fields")
	c.NotEqual(Hash64(plain), Hash64(titled), "the new fields are source data and are hashed")

	data, err := jio.Marshal(titled)
	c.NoError(err)
	c.Contains(string(data), `"short_name":"12 or less"`)
	c.Contains(string(data), `"show_in_title":true`)
	var restored TraitModifier
	c.NoError(jio.Unmarshal(data, &restored))
	c.Equal("12 or less", restored.ShortName)
	c.True(restored.ShowInTitle)
	data, err = jio.Marshal(plain)
	c.NoError(err)
	c.NotContains(string(data), "short_name")
	c.NotContains(string(data), "show_in_title")
}

func TestPrereqModifierCriteria(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	allies := addTraitWithFeatures(e, "Allies")
	frequency := newTitleTraitModifier(e, "Appears quite often (12-)", "12 or less")
	unwilling := NewTraitModifier(e, nil, false)
	unwilling.Name = "Unwilling"
	group := NewTraitModifier(e, nil, true)
	group.Name = "Frequency of Appearance"
	off := NewTraitModifier(e, group, false)
	off.Name = "Appears constantly"
	off.Disabled = true
	group.Children = []*TraitModifier{off}
	allies.SetModifiers([]*TraitModifier{frequency, unwilling, group})
	c.True(frequency.ShowsInTitle(), "precondition: the frequency modifier shows in the title")
	c.False(unwilling.ShowsInTitle(), "precondition: the unwilling modifier stays in the notes")
	c.True(group.Container(), "precondition: the group is a container")
	c.False(off.Enabled(), "precondition: the constantly modifier is disabled")
	c.Equal([]string{"Appears quite often (12-)", "12 or less", "Unwilling"}, allies.ModifierNames())

	for _, one := range []struct {
		compare   criteria.StringComparison
		qualifier string
		expected  bool
	}{
		{criteria.IsText, "Appears quite often (12-)", true},
		{criteria.IsText, "12 or less", true},
		{criteria.IsText, "unwilling", true},
		{criteria.IsText, "Appears constantly", false},
		{criteria.IsText, "Frequency of Appearance", false},
		{criteria.IsText, "Summonable, Unwilling", true},
		{criteria.ContainsText, "quite often", true},
		{criteria.IsNotText, "Summonable", true},
		{criteria.IsNotText, "12 or less", false},
	} {
		p := NewTraitPrereq()
		p.NameCriteria.Qualifier = "Allies"
		p.LevelCriteria.Compare = criteria.AnyNumber
		p.ModifierCriteria.Compare = one.compare
		p.ModifierCriteria.Qualifier = one.qualifier
		c.Equal(one.expected, p.Satisfied(e, nil, nil, "", nil), "%s %q", one.compare.Key(), one.qualifier)
	}

	both := NewTraitPrereq()
	both.NameCriteria.Qualifier = "Allies"
	both.LevelCriteria.Compare = criteria.AnyNumber
	both.ModifierCriteria.Compare = criteria.IsText
	both.ModifierCriteria.Qualifier = "Appears quite often (12-)"
	both.TitleNoteCriteria.Compare = criteria.IsText
	both.TitleNoteCriteria.Qualifier = "12 or less"
	c.True(both.Satisfied(e, nil, nil, "", nil), "a modifier in the title matches both a modifier and a title note check")

	sword := addCarriedEquipmentWithFeatures(e, "Broadsword")
	fine := NewEquipmentModifier(e, nil, false)
	fine.Name = "Fine Quality"
	fine.ShortName = "Fine"
	sword.SetModifiers([]*EquipmentModifier{fine})
	c.True(sword.ReallyEquipped(), "precondition: the sword is equipped")
	eqp := NewEquippedEquipmentPrereq()
	eqp.NameCriteria.Qualifier = "Broadsword"
	eqp.ModifierCriteria.Compare = criteria.IsText
	eqp.ModifierCriteria.Qualifier = "Fine"
	c.True(eqp.Satisfied(e, nil, nil, "", nil), "equipment matched by a modifier's short name")
	eqp.ModifierCriteria.Qualifier = "Cheap"
	c.False(eqp.Satisfied(e, nil, nil, "", nil), "equipment with no matching modifier")
}

func TestPrereqDescribeModifiers(t *testing.T) {
	c := check.New(t)
	traitIs := NewTraitPrereq()
	traitIs.NameCriteria.Qualifier = "Allies"
	traitIs.ModifierCriteria.Compare = criteria.IsText
	traitIs.ModifierCriteria.Qualifier = "Unwilling"
	equippedContains := NewEquippedEquipmentPrereq()
	equippedContains.NameCriteria.Compare = criteria.AnyText
	equippedContains.ModifierCriteria.Compare = criteria.ContainsText
	equippedContains.ModifierCriteria.Qualifier = "Fine"
	c.Equal(criteria.IsText, traitIs.ModifierCriteria.Compare, "precondition: the trait prereq matches by is")
	c.Equal(criteria.ContainsText, equippedContains.ModifierCriteria.Compare,
		"precondition: the equipment prereq contains")

	em := func(s string) string { return "[" + s + "]" }
	c.Equal("Has trait [Allies] with the modifier [Unwilling]", traitIs.Describe(nil, nil, em))
	c.Equal(`Has equipped equipment with a modifier that contains "[Fine]"`, equippedContains.Describe(nil, nil, em))
}

func TestModifierNotesWithTitleAndHideNotes(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	allies := addTraitWithFeatures(e, "Allies", newTitleNoteFor("Bob"))
	titledWithNotes := newTitleTraitModifier(e, "Appears quite often (12-)", "12 or less")
	titledWithNotes.LocalNotes = "Usually at night"
	titledHidden := newTitleTraitModifier(e, "25% of your starting points", "Built on 25%")
	titledHidden.LocalNotes = "Rounded down"
	titledHidden.HideNotes = true
	plainHidden := NewTraitModifier(e, nil, false)
	plainHidden.Name = "Unwilling"
	plainHidden.LocalNotes = "Must be persuaded"
	plainHidden.HideNotes = true
	plain := NewTraitModifier(e, nil, false)
	plain.Name = "Summonable"
	plain.LocalNotes = "Takes a turn"
	allies.SetModifiers([]*TraitModifier{titledWithNotes, titledHidden, plainHidden, plain})
	c.True(titledWithNotes.ShowsInTitle(), "precondition: the frequency modifier shows in the title")
	c.True(titledHidden.ShowsInTitle(), "precondition: the point total modifier shows in the title")
	c.False(plainHidden.ShowsInTitle(), "precondition: the unwilling modifier stays in the notes")
	c.False(plain.ShowsInTitle(), "precondition: the summonable modifier stays in the notes")

	c.Equal("Allies (Bob; 12 or less; Built on 25%)", allies.String())
	c.Equal("12 or less (Usually at night); Unwilling; Summonable (Takes a turn)", allies.ModifierNotes(),
		"a title modifier with notes still shows them; hidden notes are left out, and so is a title modifier with none")
	c.Equal("Unwilling (Must be persuaded)", plainHidden.FullDescription(), "the full description keeps hidden notes")

	hidden := Hash64(plainHidden)
	plainHidden.HideNotes = false
	c.NotEqual(hidden, Hash64(plainHidden), "hiding notes is source data and is hashed")
	plainHidden.HideNotes = true
	data, err := jio.Marshal(plainHidden)
	c.NoError(err)
	c.Contains(string(data), `"hide_notes":true`)
}

func TestActiveModifierForMatchesShortName(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	allies := addTraitWithFeatures(e, "Allies")
	frequency := newTitleTraitModifier(e, "Appears quite often (12-)", "12 or less")
	unwilling := NewTraitModifier(e, nil, false)
	unwilling.Name = "Unwilling"
	allies.SetModifiers([]*TraitModifier{frequency, unwilling})
	c.Equal("12 or less", frequency.ShortName, "precondition: the frequency modifier has a short name")
	c.Equal("", unwilling.ShortName, "precondition: the unwilling modifier has none")
	c.True(allies.ActiveModifierFor("Appears quite often (12-)") == frequency, "found by its full name")
	c.True(allies.ActiveModifierFor("12 OR LESS") == frequency, "found by its short name, ignoring case")
	c.True(allies.ActiveModifierFor("Unwilling") == unwilling, "a modifier without a short name is found by name")
	c.Nil(allies.ActiveModifierFor(""), "an empty name doesn't match an empty short name")

	sword := addCarriedEquipmentWithFeatures(e, "Broadsword")
	fine := NewEquipmentModifier(e, nil, false)
	fine.Name = "Fine Quality"
	fine.ShortName = "Fine"
	sword.SetModifiers([]*EquipmentModifier{fine})
	c.Equal("Fine", fine.ShortName, "precondition: the fine modifier has a short name")
	c.True(sword.ActiveModifierFor("fine") == fine, "equipment modifiers are found by short name too")
}

func TestTitleNoteDescribe(t *testing.T) {
	c := check.New(t)
	em := func(s string) string { return "[" + s + "]" }
	who := newTitleNoteFor("@Who@")
	blank := newTitleNoteFor("  ")
	c.Equal("@Who@", who.Text, "precondition: the note holds a marker")
	c.Equal("  ", blank.Text, "precondition: the other note is blank")
	c.Equal("Adds the title note [Bob]", who.Describe(nil, map[string]string{"Who": "Bob"}, em))
	c.Equal("Adds the title note [@Who@]", who.Describe(nil, nil, em), "a marker without a value is left as it is")
	c.Equal("Adds an empty title note", blank.Describe(nil, nil, em))
}
