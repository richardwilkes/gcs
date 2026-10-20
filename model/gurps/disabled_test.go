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
	"encoding/json/v2"
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestDisabledItemsAreAsIfAbsent(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	_, skill, spell, eqp := newSwitchableItemSet(e, NewAttributeBonus(StrengthID))
	skill.Points = fxp.Four
	spell.Points = fxp.Four
	eqp.BaseWeight = "2 lb"
	eqp.BaseValue = "10"
	needsSkill := addTraitWithFeatures(e, "Needs Brawling")
	needsSkill.Prereq = NewPrereqList()
	addSkillPrereq(needsSkill.Prereq, "Brawling", fxp.One)
	e.Recalculate()
	c.Equal(fxp.Four, e.AttributeBonusFor(StrengthID, stlimit.None, nil))
	c.Equal(fxp.Four, e.PointsBreakdown().Skills)
	c.Equal(fxp.Four, e.PointsBreakdown().Spells)
	c.True(skill.LevelData.Level > 0)
	c.True(spell.LevelData.Level > 0)
	c.True(e.WeightCarried(false) > 0)
	c.Equal(fxp.Ten, e.WealthCarried())
	c.Equal(1, len(e.SkillNamed("Brawling", "", false, nil)))
	c.Equal("", needsSkill.UnsatisfiedReason)

	skill.SetDisabled(true)
	spell.SetDisabled(true)
	eqp.SetDisabled(true)
	e.Recalculate()
	c.Equal(fxp.One, e.AttributeBonusFor(StrengthID, stlimit.None, nil), "only the trait's bonus remains")
	c.Equal(fxp.Int(0), e.PointsBreakdown().Skills)
	c.Equal(fxp.Int(0), e.PointsBreakdown().Spells)
	c.Equal(fxp.Int(0), skill.LevelData.Level)
	c.Equal(fxp.Int(0), spell.LevelData.Level)
	c.Equal(fxp.Weight(0), e.WeightCarried(false))
	c.Equal(fxp.Int(0), e.WealthCarried())
	c.Equal(0, len(e.SkillNamed("Brawling", "", false, nil)))
	c.NotEqual("", needsSkill.UnsatisfiedReason)
}

func TestDisabledContainerDisablesItsContents(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	box := NewEquipment(e, nil, true)
	box.BaseWeight = "1 lb"
	box.BaseValue = "5"
	item := NewEquipment(e, box, false)
	item.BaseWeight = "3 lb"
	item.BaseValue = "10"
	other := NewEquipment(e, box, false)
	other.BaseWeight = "2 lb"
	other.BaseValue = "20"
	box.Children = []*Equipment{item, other}
	e.CarriedEquipment = append(e.CarriedEquipment, box)
	e.Recalculate()
	c.Equal(fxp.WeightFromInteger(6, fxp.Pound), e.WeightCarried(false))
	c.Equal(fxp.FromInteger(35), e.WealthCarried())

	item.SetDisabled(true)
	e.DiscardCaches() // The carried weight is cached until the caches are discarded, as a recalculation does.
	c.Equal(fxp.WeightFromInteger(3, fxp.Pound), e.WeightCarried(false), "a disabled item adds nothing to its container")
	c.Equal(fxp.WeightFromInteger(3, fxp.Pound), e.WeightCarried(true))
	c.Equal(fxp.FromInteger(25), e.WealthCarried())

	item.SetDisabled(false)
	box.SetDisabled(true)
	e.DiscardCaches()
	c.False(item.Enabled())
	c.False(item.ReallyEquipped())
	c.Equal(fxp.Weight(0), e.WeightCarried(false))
}

func TestDisabledItemsAreLeftOutOfTheListTotals(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	a := addCarriedEquipmentWithFeatures(e, "A")
	a.BaseWeight = "3 lb"
	a.BaseValue = "10"
	b := addCarriedEquipmentWithFeatures(e, "B")
	b.BaseWeight = "2 lb"
	b.BaseValue = "20"
	e.Recalculate()
	title, _ := equipmentTotalsTitle("Carried Equipment", e.CarriedEquipment, e.SheetSettings)
	c.Equal("Carried Equipment (5 lb; $30)", title)

	b.SetDisabled(true)
	e.Recalculate()
	title, _ = equipmentTotalsTitle("Carried Equipment", e.CarriedEquipment, e.SheetSettings)
	c.Equal("Carried Equipment (3 lb; $10)", title)
}

func TestDisabledIsClearedOffASheet(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	skill := NewSkill(e, nil, false)
	skill.SetDisabled(true)
	spell := NewSpell(e, nil, false)
	spell.SetDisabled(true)
	eqp := NewEquipment(e, nil, false)
	eqp.SetDisabled(true)
	c.True(skill.Clone(LibraryFile{}, e, nil, Copy).IsDisabled())
	c.True(spell.Clone(LibraryFile{}, e, nil, Copy).IsDisabled())
	c.True(eqp.Clone(LibraryFile{}, e, nil, Copy).IsDisabled())
	c.False(skill.Clone(LibraryFile{}, nil, nil, Copy).IsDisabled())
	c.False(spell.Clone(LibraryFile{}, nil, nil, Copy).IsDisabled())
	c.False(eqp.Clone(LibraryFile{}, nil, nil, Copy).IsDisabled())
}

func TestDisabledItemsHaveNoUnsatisfiedReason(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	_, skill, spell, eqp := newSwitchableItemSet(e)
	box := NewEquipment(e, nil, true)
	boxed := NewEquipment(e, box, false)
	box.Children = []*Equipment{boxed}
	e.CarriedEquipment = append(e.CarriedEquipment, box)
	requireMissingEquipment := func() *PrereqList {
		list := NewPrereqList()
		p := NewEquippedEquipmentPrereq()
		p.Parent = list
		p.NameCriteria.Qualifier = "Lockpicks"
		list.Prereqs = append(list.Prereqs, p)
		return list
	}
	skill.Prereq = requireMissingEquipment()
	spell.Prereq = requireMissingEquipment()
	eqp.Prereq = requireMissingEquipment()
	boxed.Prereq = requireMissingEquipment()
	e.Recalculate()
	c.NotEqual("", skill.UnsatisfiedReason)
	c.NotEqual("", spell.UnsatisfiedReason)
	c.NotEqual("", eqp.UnsatisfiedReason)
	c.NotEqual("", boxed.UnsatisfiedReason)
	c.True(skill.takesEquipmentPenalty)
	c.True(spell.takesEquipmentPenalty)

	skill.SetDisabled(true)
	spell.SetDisabled(true)
	eqp.SetDisabled(true)
	box.SetDisabled(true)
	e.Recalculate()
	c.Equal("", skill.UnsatisfiedReason)
	c.Equal("", spell.UnsatisfiedReason)
	c.Equal("", eqp.UnsatisfiedReason)
	c.Equal("", boxed.UnsatisfiedReason, "an item in a disabled container is not judged")
	c.False(skill.takesEquipmentPenalty)
	c.False(spell.takesEquipmentPenalty)

	skill.SetDisabled(false)
	spell.SetDisabled(false)
	eqp.SetDisabled(false)
	box.SetDisabled(false)
	e.Recalculate()
	c.NotEqual("", skill.UnsatisfiedReason)
	c.NotEqual("", spell.UnsatisfiedReason)
	c.NotEqual("", eqp.UnsatisfiedReason)
	c.NotEqual("", boxed.UnsatisfiedReason)
}

func TestDisabledSkillsAndSpellsShowNoLevel(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	_, skill, spell, _ := newSwitchableItemSet(e)
	skill.Points = fxp.Four
	spell.Points = fxp.Four
	e.Recalculate()
	var data CellData
	skill.CellData(SkillLevelColumn, &data)
	c.NotEqual("-", data.Primary)
	spell.CellData(SpellLevelColumn, &data)
	c.NotEqual("-", data.Primary)

	skill.SetDisabled(true)
	spell.SetDisabled(true)
	e.Recalculate()
	data = CellData{}
	skill.CellData(SkillLevelColumn, &data)
	c.Equal("-", data.Primary)
	c.Equal("", data.Tooltip)
	data = CellData{}
	spell.CellData(SpellLevelColumn, &data)
	c.Equal("-", data.Primary)
	c.Equal("", data.Tooltip)
}

func TestDisabledSkillsAndSpellsCostNothing(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	skills := NewSkill(e, nil, true)
	brawling := NewSkill(e, skills, false)
	brawling.Points = fxp.Four
	karate := NewSkill(e, skills, false)
	karate.Points = fxp.FromInteger(8)
	skills.Children = []*Skill{brawling, karate}
	e.Skills = append(e.Skills, skills)
	spells := NewSpell(e, nil, true)
	light := NewSpell(e, spells, false)
	light.Points = fxp.One
	fireball := NewSpell(e, spells, false)
	fireball.Points = fxp.Two
	spells.Children = []*Spell{light, fireball}
	e.Spells = append(e.Spells, spells)
	e.Recalculate()
	c.Equal(fxp.FromInteger(12), skills.AdjustedPoints(nil))
	c.Equal(fxp.Three, spells.AdjustedPoints(nil))

	karate.SetDisabled(true)
	fireball.SetDisabled(true)
	e.Recalculate()
	c.Equal(fxp.Int(0), karate.AdjustedPoints(nil))
	c.Equal(fxp.Int(0), fireball.AdjustedPoints(nil))
	c.Equal(fxp.Four, skills.AdjustedPoints(nil))
	c.Equal(fxp.One, spells.AdjustedPoints(nil))
	c.Equal(NumericRangeOf(fxp.Four), skills.PointsRange(nil))
	c.Equal(e.PointsBreakdown().Skills, skills.AdjustedPoints(nil))
	c.Equal(e.PointsBreakdown().Spells, spells.AdjustedPoints(nil))
}

func TestLoadingOffASheetClearsDisabled(t *testing.T) {
	c := check.New(t)
	trait := NewTrait(nil, nil, false)
	trait.SetDisabled(true)
	skill := NewSkill(nil, nil, false)
	skill.SetDisabled(true)
	spell := NewSpell(nil, nil, false)
	spell.SetDisabled(true)
	eqp := NewEquipment(nil, nil, false)
	eqp.SetDisabled(true)

	entity := NewEntity()
	entity.Skills = []*Skill{skill}
	entity.Spells = []*Spell{spell}
	entity.CarriedEquipment = []*Equipment{eqp}
	data, err := json.Marshal(entity)
	c.NoError(err)
	var loadedEntity Entity
	c.NoError(json.Unmarshal(data, &loadedEntity))
	c.True(loadedEntity.Skills[0].IsDisabled(), "a character sheet keeps a disabled skill")
	c.True(loadedEntity.Spells[0].IsDisabled(), "a character sheet keeps a disabled spell")
	c.True(loadedEntity.CarriedEquipment[0].IsDisabled(), "a character sheet keeps disabled equipment")

	template := NewTemplate()
	template.Traits = []*Trait{trait}
	template.Skills = []*Skill{skill}
	template.Spells = []*Spell{spell}
	template.Equipment = []*Equipment{eqp}
	data, err = json.Marshal(template)
	c.NoError(err)
	var loadedTemplate Template
	c.NoError(json.Unmarshal(data, &loadedTemplate))
	c.True(loadedTemplate.Traits[0].IsDisabled(), "a template keeps a disabled trait")
	c.False(loadedTemplate.Skills[0].IsDisabled(), "a template must not keep a disabled skill")
	c.False(loadedTemplate.Spells[0].IsDisabled(), "a template must not keep a disabled spell")
	c.False(loadedTemplate.Equipment[0].IsDisabled(), "a template must not keep disabled equipment")

	loot := NewLoot()
	loot.Equipment = []*Equipment{eqp}
	data, err = json.Marshal(loot)
	c.NoError(err)
	var loadedLoot Loot
	c.NoError(json.Unmarshal(data, &loadedLoot))
	c.False(loadedLoot.Equipment[0].IsDisabled(), "a loot sheet must not keep disabled equipment")

	list, err := json.Marshal(&listData[*Skill]{Version: jio.CurrentDataVersion, Rows: []*Skill{skill}})
	c.NoError(err)
	skills, err := NewSkillsFromFile(fstest.MapFS{"list.skl": &fstest.MapFile{Data: list}}, "list.skl")
	c.NoError(err)
	c.False(skills[0].IsDisabled(), "a skill list must not keep a disabled skill")
	list, err = json.Marshal(&listData[*Trait]{Version: jio.CurrentDataVersion, Rows: []*Trait{trait}})
	c.NoError(err)
	traits, err := NewTraitsFromFile(fstest.MapFS{"list.adq": &fstest.MapFile{Data: list}}, "list.adq")
	c.NoError(err)
	c.True(traits[0].IsDisabled(), "a trait list keeps a disabled trait")
}

// TestDisabledGroupIsLeftOutOfAnEquipmentChoice verifies that the contents of a disabled group picked from separately
// are left out of the choice's range, just as a disabled option sitting directly in the choice is.
func TestDisabledGroupIsLeftOutOfAnEquipmentChoice(t *testing.T) {
	c := check.New(t)
	choice := NewEquipmentChoiceContainer(nil, nil)
	choice.TemplatePicker.Type = picker.Count
	choice.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
	choice.TemplatePicker.Qualifier.Qualifier = fxp.One
	group := NewEquipmentGroup(nil, choice)
	group.PickSeparately = true
	item(group, "Rope", "10", "1 lb")
	choice.Children = append(choice.Children, group)
	item(choice, "Knife", "20", "2 lb")
	c.Equal("10~20", choice.ExtendedValueRange().String())
	c.Equal("1~2 lb", FormatWeightRange(choice.ExtendedWeightRange(fxp.Pound), fxp.Pound.Format))

	group.SetDisabled(true)
	c.Equal("20", choice.ExtendedValueRange().String(), "the contents of a disabled group are not options")
	c.Equal("2 lb", FormatWeightRange(choice.ExtendedWeightRange(fxp.Pound), fxp.Pound.Format))
}

// TestDisabledContainersWithPickersCostNothing verifies that a disabled skill or spell container reports no points
// through its ranges either, whatever its picker would be worth, so the points column agrees with the points total.
func TestDisabledContainersWithPickersCostNothing(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	skills := NewSkill(e, nil, true)
	skills.TemplatePicker.Type = picker.Points
	skills.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
	skills.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(20)
	brawling := NewSkill(e, skills, false)
	brawling.Points = fxp.Four
	skills.Children = []*Skill{brawling}
	e.Skills = append(e.Skills, skills)
	spells := NewSpell(e, nil, true)
	spells.TemplatePicker = skills.TemplatePicker
	light := NewSpell(e, spells, false)
	light.Points = fxp.One
	spells.Children = []*Spell{light}
	e.Spells = append(e.Spells, spells)
	e.Recalculate()
	twenty := NumericRangeOf(fxp.FromInteger(20))
	c.Equal(twenty, skills.PointsRange(nil))
	c.Equal(twenty, skills.RawPointsRange())
	c.Equal(twenty, spells.PointsRange(nil))
	c.Equal(twenty, spells.RawPointsRange())

	skills.SetDisabled(true)
	spells.SetDisabled(true)
	e.Recalculate()
	none := NumericRangeOf(0)
	c.Equal(none, skills.PointsRange(nil), "a disabled skill container's range holds its AdjustedPoints")
	c.Equal(none, skills.RawPointsRange())
	c.Equal(none, brawling.RawPointsRange(), "a skill under a disabled container is counted as nothing")
	c.Equal(none, spells.PointsRange(nil), "a disabled spell container's range holds its AdjustedPoints")
	c.Equal(none, spells.RawPointsRange())
	c.Equal(none, light.RawPointsRange(), "a spell under a disabled container is counted as nothing")

	// A disabled child of an enabled container is counted as nothing, as AdjustedPoints counts it.
	skills.SetDisabled(false)
	skills.TemplatePicker = TemplatePicker{}
	brawling.SetDisabled(true)
	c.Equal(none, skills.PointsRange(nil))
	c.Equal(none, skills.RawPointsRange())
}
