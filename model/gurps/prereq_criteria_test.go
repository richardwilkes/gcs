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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellcmp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// isText returns a criterion that is met by text that is the qualifier.
func isText(qualifier string) criteria.Text {
	return criteria.Text{Compare: criteria.IsText, Qualifier: qualifier}
}

// atLeast returns a criterion that is met by a number of at least the qualifier.
func atLeast(qualifier int) criteria.Number {
	return criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.FromInteger(qualifier)}
}

// TestTraitPrereqCriteria checks the points and tags of a trait prerequisite, and its modifier criteria, which a single
// enabled modifier has to meet together, containers left out.
func TestTraitPrereqCriteria(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	magery := gurps.NewTrait(e, nil, false)
	magery.Name = "Magery"
	magery.BasePoints = fxp.FromInteger(15)
	magery.Tags = []string{"Mental"}
	oneCollege := gurps.NewTraitModifier(e, nil, false)
	oneCollege.Name = "One College"
	oneCollege.LocalNotes = "Fire"
	oneCollege.Tags = []string{"Limitation"}
	water := gurps.NewTraitModifier(e, nil, false)
	water.Name = "Water Only"
	water.LocalNotes = "Water"
	solar := gurps.NewTraitModifier(e, nil, false)
	solar.Name = "Solar"
	solar.Tags = []string{"Limitation"}
	solar.SetDisabled(true)
	group := gurps.NewTraitModifier(e, nil, true)
	group.Name = "Group"
	magery.Modifiers = []*gurps.TraitModifier{oneCollege, water, solar, group}
	e.Traits = []*gurps.Trait{magery}
	c.Equal(fxp.FromInteger(15), magery.AdjustedPoints(nil), "precondition: Magery is worth 15 points")
	c.True(oneCollege.Enabled(), "precondition: One College is enabled")
	c.True(water.Enabled(), "precondition: Water Only is enabled")
	c.False(solar.Enabled(), "precondition: Solar is disabled")
	c.True(group.Container(), "precondition: Group is a container")

	for _, one := range []struct {
		name  string
		setup func(p *gurps.TraitPrereq)
		want  bool
	}{
		{"points met", func(p *gurps.TraitPrereq) { p.PointsCriteria = atLeast(15) }, true},
		{"points unmet", func(p *gurps.TraitPrereq) { p.PointsCriteria = atLeast(16) }, false},
		{"tag met", func(p *gurps.TraitPrereq) { p.TagsCriteria = isText("Mental") }, true},
		{"tag unmet", func(p *gurps.TraitPrereq) { p.TagsCriteria = isText("Physical") }, false},
		{"modifier name", func(p *gurps.TraitPrereq) { p.ModifierNameCriteria = isText("One College") }, true},
		{"disabled modifier name", func(p *gurps.TraitPrereq) { p.ModifierNameCriteria = isText("Solar") }, false},
		{"container modifier name", func(p *gurps.TraitPrereq) { p.ModifierNameCriteria = isText("Group") }, false},
		{"modifier notes", func(p *gurps.TraitPrereq) { p.ModifierNotesCriteria = isText("Fire") }, true},
		{"modifier tag", func(p *gurps.TraitPrereq) { p.ModifierTagsCriteria = isText("Limitation") }, true},
		{"one modifier meets all", func(p *gurps.TraitPrereq) {
			p.ModifierNameCriteria = isText("One College")
			p.ModifierTagsCriteria = isText("Limitation")
		}, true},
		{"no one modifier meets all", func(p *gurps.TraitPrereq) {
			p.ModifierNameCriteria = isText("One College")
			p.ModifierNotesCriteria = isText("Water")
		}, false},
	} {
		p := namedTrait("Magery")
		one.setup(p)
		c.Equal(one.want, p.Satisfied(e, nil, nil, "", nil), one.name)
	}
}

// TestTraitPrereqContainerModifiers checks that a trait's modifier criteria see the enabled modifiers of the containers
// holding it, which apply to it.
func TestTraitPrereqContainerModifiers(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	container := gurps.NewTrait(e, nil, true)
	container.Name = "Breath Weapons"
	cone := gurps.NewTraitModifier(e, nil, false)
	cone.Name = "Cone"
	jet := gurps.NewTraitModifier(e, nil, false)
	jet.Name = "Jet"
	jet.SetDisabled(true)
	container.Modifiers = []*gurps.TraitModifier{cone, jet}
	breath := gurps.NewTrait(e, container, false)
	breath.Name = "Innate Attack"
	container.Children = []*gurps.Trait{breath}
	e.Traits = []*gurps.Trait{container}
	c.True(cone.Enabled(), "precondition: Cone is enabled")
	c.False(jet.Enabled(), "precondition: Jet is disabled")
	c.Equal(0, len(breath.Modifiers), "precondition: the trait has no modifiers of its own")

	p := namedTrait("Innate Attack")
	p.ModifierNameCriteria = isText("Cone")
	c.True(p.Satisfied(e, nil, nil, "", nil), "the container's enabled modifier applies to the trait")
	p.ModifierNameCriteria = isText("Jet")
	c.False(p.Satisfied(e, nil, nil, "", nil), "its disabled one doesn't")
}

// TestSkillPrereqCriteria checks the points and tags of a skill prerequisite.
func TestSkillPrereqCriteria(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	broadsword := gurps.NewSkill(e, nil, false)
	broadsword.Name = "Broadsword"
	broadsword.Points = fxp.Four
	broadsword.Tags = []string{"Combat"}
	e.Skills = []*gurps.Skill{broadsword}
	c.Equal(fxp.Four, broadsword.AdjustedPoints(nil), "precondition: Broadsword has 4 points")

	for _, one := range []struct {
		name  string
		setup func(p *gurps.SkillPrereq)
		want  bool
	}{
		{"points met", func(p *gurps.SkillPrereq) { p.PointsCriteria = atLeast(4) }, true},
		{"points unmet", func(p *gurps.SkillPrereq) { p.PointsCriteria = atLeast(8) }, false},
		{"tag met", func(p *gurps.SkillPrereq) { p.TagsCriteria = isText("Combat") }, true},
		{"tag unmet", func(p *gurps.SkillPrereq) { p.TagsCriteria = isText("Social") }, false},
	} {
		p := gurps.NewSkillPrereq()
		p.NameCriteria = isText("Broadsword")
		one.setup(p)
		c.Equal(one.want, p.Satisfied(e, nil, nil, "", nil), one.name)
	}
}

// TestSpellPrereqLevel checks that only spells at the level a spell prerequisite asks for count toward its quantity.
func TestSpellPrereqLevel(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	fireSpell := func(name string, level int) *gurps.Spell {
		sp := gurps.NewSpell(e, nil, false)
		sp.Name = name
		sp.College = gurps.CollegeList{"Fire"}
		sp.Points = fxp.One
		sp.LevelData.Level = fxp.FromInteger(level)
		e.Spells = append(e.Spells, sp)
		return sp
	}
	ignite := fireSpell("Ignite Fire", 12)
	create := fireSpell("Create Fire", 15)
	fireball := fireSpell("Fireball", 16)
	c.Equal(fxp.Twelve, ignite.LevelData.Level, "precondition: Ignite Fire is at level 12")
	c.Equal(fxp.FromInteger(15), create.LevelData.Level, "precondition: Create Fire is at level 15")
	c.Equal(fxp.FromInteger(16), fireball.LevelData.Level, "precondition: Fireball is at level 16")

	p := gurps.NewSpellPrereq()
	p.SubType = spellcmp.College
	p.QualifierCriteria = isText("Fire")
	p.SamePowerSource = false
	p.QuantityCriteria = atLeast(2)
	c.True(p.Satisfied(e, nil, nil, "", nil), "two spells count at any level")
	p.LevelCriteria = atLeast(15)
	c.True(p.Satisfied(e, nil, nil, "", nil), "two spells are at level 15 or more")
	p.LevelCriteria = atLeast(16)
	c.False(p.Satisfied(e, nil, nil, "", nil), "only one spell is at level 16 or more")
	c.Equal("Knows at least 2 spells whose college is Fire at level at least 16",
		p.Describe(nil, nil, func(s string) string { return s }))
}

// TestEquipmentPrereqs checks equipped and carried equipment prerequisites, either of which may be negated, and that
// only an unmet one to have equipment equipped takes the missing equipment penalty.
func TestEquipmentPrereqs(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	sword := gurps.NewEquipment(e, nil, false)
	sword.Name = "Sword"
	sword.Equipped = true
	kit := gurps.NewEquipment(e, nil, false)
	kit.Name = "Kit"
	kit.Equipped = false
	spare := gurps.NewEquipment(e, nil, false)
	spare.Name = "Spare"
	spare.Equipped = false
	spare.SetDisabled(true)
	used := gurps.NewEquipment(e, nil, false)
	used.Name = "Used"
	used.Equipped = false
	used.Quantity = 0
	box := gurps.NewEquipment(e, nil, true)
	box.Name = "Box"
	box.Equipped = false
	box.Quantity = 0
	boxed := gurps.NewEquipment(e, box, false)
	boxed.Name = "Boxed"
	boxed.Equipped = false
	box.Children = []*gurps.Equipment{boxed}
	e.CarriedEquipment = []*gurps.Equipment{sword, kit, spare, used, box}
	c.True(sword.ReallyEquipped(), "precondition: the sword is equipped")
	c.False(kit.ReallyEquipped(), "precondition: the kit is not equipped")
	c.True(kit.Enabled(), "precondition: the kit is enabled")
	c.False(spare.Enabled(), "precondition: the spare is disabled")
	c.Equal(fxp.Int(0), used.Quantity, "precondition: none of the used item is left")
	c.Equal(fxp.One, boxed.Quantity, "precondition: one boxed item is in the box")
	c.Equal(fxp.Int(0), box.Quantity, "precondition: none of the box is left")

	for _, one := range []struct {
		name    string
		carried bool
		item    string
		has     bool
		want    bool
		penalty bool
	}{
		{"equipped sword", false, "Sword", true, true, false},
		{"equipped kit", false, "Kit", true, false, true},
		{"not equipped sword", false, "Sword", false, false, false},
		{"not equipped kit", false, "Kit", false, true, false},
		{"carried kit", true, "Kit", true, true, false},
		{"carried sword", true, "Sword", true, false, false},
		{"carried disabled spare", true, "Spare", true, false, false},
		{"carried used up", true, "Used", true, false, false},
		{"carried in a box that is used up", true, "Boxed", true, false, false},
		{"not carried kit", true, "Kit", false, false, false},
	} {
		p := gurps.NewEquippedEquipmentPrereq()
		if one.carried {
			p = gurps.NewCarriedEquipmentPrereq()
		}
		p.NameCriteria = isText(one.item)
		p.Has = one.has
		var penalty bool
		c.Equal(one.want, p.Satisfied(e, nil, nil, "", &penalty), one.name)
		c.Equal(one.penalty, penalty, "%s: penalty", one.name)
	}

	em := func(s string) string { return s }
	carried := gurps.NewCarriedEquipmentPrereq()
	carried.NameCriteria = isText("Kit")
	c.Equal("Has Kit carried but not equipped", carried.Describe(nil, nil, em))
	carried.Has = false
	carried.NameCriteria.Compare = criteria.AnyText
	c.Equal("Does not have any equipment carried but not equipped", carried.Describe(nil, nil, em))
	carried.NameCriteria = criteria.Text{Compare: criteria.ContainsText, Qualifier: "Ki"}
	c.Equal(`Does not have unequipped carried equipment whose name contains "Ki"`, carried.Describe(nil, nil, em))

	negated := gurps.NewPrereqList()
	negated.All = false
	negated.Not = true
	met := namedTrait("Z")
	met.Has = false
	shield := gurps.NewEquippedEquipmentPrereq()
	shield.NameCriteria = isText("Shield")
	negated.Prereqs = gurps.Prereqs{met, shield}
	var penalty bool
	c.False(negated.Satisfied(e, nil, nil, "", &penalty), "none of a met prerequisite is unmet")
	c.False(penalty, "an unmet equipment prerequisite in a negated list takes no penalty")
}

// TestNewPrereqCriteriaHashAndKeys checks that the hash follows the new fields, and that their nameable keys are found.
func TestNewPrereqCriteriaHashAndKeys(t *testing.T) {
	c := check.New(t)
	list := gurps.NewPrereqList()
	list.Prereqs = gurps.Prereqs{namedTrait("Magery")}
	base := gurps.Hash64(list)
	list.Not = true
	c.NotEqual(base, gurps.Hash64(list), "negating a list changes its hash")
	equipped := gurps.NewEquippedEquipmentPrereq()
	base = gurps.Hash64(equipped)
	equipped.Has = false
	c.NotEqual(base, gurps.Hash64(equipped), "negating an equipment prerequisite changes its hash")

	trait := namedTrait("Magery")
	trait.TagsCriteria = isText("@Tag@")
	trait.ModifierNameCriteria = isText("@Name@")
	trait.ModifierNotesCriteria = isText("@Notes@")
	trait.ModifierTagsCriteria = isText("@ModTag@")
	keys := make(map[string]string)
	trait.FillWithNameableKeys(keys, nil)
	for _, key := range []string{"Tag", "Name", "Notes", "ModTag"} {
		_, found := keys[key]
		c.True(found, "the trait's %s key is found", key)
	}
	skill := gurps.NewSkillPrereq()
	skill.TagsCriteria = isText("@SkillTag@")
	keys = make(map[string]string)
	skill.FillWithNameableKeys(keys, nil)
	_, found := keys["SkillTag"]
	c.True(found, "the skill's tag key is found")
}

// TestEquipmentPrereqHasOnDisk checks that an equipment prerequisite saved before it could be negated loads as one to
// have, that a negated one keeps its negation, and that the carried type loads as an equipment prerequisite.
func TestEquipmentPrereqHasOnDisk(t *testing.T) {
	c := check.New(t)
	var list gurps.PrereqList
	c.NoError(jio.Unmarshal([]byte(`{"type":"prereq_list","all":true,"prereqs":[
		{"type":"equipped_equipment","name":{"compare":"is","qualifier":"Sword"}},
		{"type":"equipped_equipment","has":false},
		{"type":"carried_equipment","has":true}]}`), &list), "the list should load")
	c.Equal(3, len(list.Prereqs), "every prerequisite loads")
	for i, want := range []struct {
		kind prereq.Type
		has  bool
	}{{prereq.EquippedEquipment, true}, {prereq.EquippedEquipment, false}, {prereq.CarriedEquipment, true}} {
		p, ok := list.Prereqs[i].(*gurps.EquippedEquipmentPrereq)
		c.True(ok, "prerequisite %d is an equipment prerequisite", i)
		if ok {
			c.Equal(want.kind, p.Type, "prerequisite %d type", i)
			c.Equal(want.has, p.Has, "prerequisite %d has", i)
		}
	}
	data, err := jio.Marshal(gurps.NewEquippedEquipmentPrereq())
	c.NoError(err, "a new equipment prerequisite should marshal")
	c.Contains(string(data), `"has":true`, "has is always written")
}

// TestEquipmentPrereqModifiers checks the modifier criteria of an equipment prerequisite, which a single enabled
// modifier has to meet together, and that they are saved and described alongside the others.
func TestEquipmentPrereqModifiers(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	sword := gurps.NewEquipment(e, nil, false)
	sword.Name = "Sword"
	sword.Equipped = true
	fine := gurps.NewEquipmentModifier(e, nil, false)
	fine.Name = "Fine"
	fine.LocalNotes = "Edge"
	fine.Tags = []string{"Quality"}
	ornate := gurps.NewEquipmentModifier(e, nil, false)
	ornate.Name = "Ornate"
	ornate.LocalNotes = "Gold"
	cheap := gurps.NewEquipmentModifier(e, nil, false)
	cheap.Name = "Cheap"
	cheap.SetDisabled(true)
	sword.Modifiers = []*gurps.EquipmentModifier{fine, ornate, cheap}
	e.CarriedEquipment = []*gurps.Equipment{sword}
	c.True(sword.ReallyEquipped(), "precondition: the sword is equipped")
	c.True(fine.Enabled(), "precondition: Fine is enabled")
	c.True(ornate.Enabled(), "precondition: Ornate is enabled")
	c.False(cheap.Enabled(), "precondition: Cheap is disabled")

	for _, one := range []struct {
		name  string
		setup func(p *gurps.EquippedEquipmentPrereq)
		want  bool
	}{
		{"modifier name", func(p *gurps.EquippedEquipmentPrereq) { p.ModifierNameCriteria = isText("Fine") }, true},
		{"disabled modifier name", func(p *gurps.EquippedEquipmentPrereq) { p.ModifierNameCriteria = isText("Cheap") }, false},
		{"modifier notes", func(p *gurps.EquippedEquipmentPrereq) { p.ModifierNotesCriteria = isText("Gold") }, true},
		{"modifier tag", func(p *gurps.EquippedEquipmentPrereq) { p.ModifierTagsCriteria = isText("Quality") }, true},
		{"one modifier meets all", func(p *gurps.EquippedEquipmentPrereq) {
			p.ModifierNameCriteria = isText("Fine")
			p.ModifierNotesCriteria = isText("Edge")
		}, true},
		{"no one modifier meets all", func(p *gurps.EquippedEquipmentPrereq) {
			p.ModifierNameCriteria = isText("Fine")
			p.ModifierNotesCriteria = isText("Gold")
		}, false},
	} {
		p := gurps.NewEquippedEquipmentPrereq()
		p.NameCriteria = isText("Sword")
		one.setup(p)
		c.Equal(one.want, p.Satisfied(e, nil, nil, "", nil), one.name)
	}

	p := gurps.NewEquippedEquipmentPrereq()
	p.NameCriteria = isText("Sword")
	p.ModifierNameCriteria = isText("Fine")
	p.ModifierTagsCriteria = isText("Quality")
	c.Equal("Has Sword equipped with a modifier whose name is Fine and tagged Quality",
		p.Describe(nil, nil, func(s string) string { return s }))
	data, err := jio.Marshal(p)
	c.NoError(err, "the prerequisite should marshal")
	c.Contains(string(data), `"modifier_name":{"compare":"is","qualifier":"Fine"}`,
		"the modifier criteria are saved beside the others")
	var restored gurps.EquippedEquipmentPrereq
	c.NoError(jio.Unmarshal(data, &restored), "the prerequisite should unmarshal")
	c.Equal(p.ModifierCriteria, restored.ModifierCriteria, "the modifier criteria survive a round trip")

	base := gurps.Hash64(p)
	p.ModifierNotesCriteria = isText("Edge")
	c.NotEqual(base, gurps.Hash64(p), "the modifier criteria change the hash")
	p.ModifierNameCriteria = isText("@Name@")
	keys := make(map[string]string)
	p.FillWithNameableKeys(keys, nil)
	_, found := keys["Name"]
	c.True(found, "the modifier name's key is found")
}
