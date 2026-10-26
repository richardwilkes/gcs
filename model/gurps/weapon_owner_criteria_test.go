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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// weaponNameSelectors holds a damage bonus, a skill bonus and a damage type override, each selecting weapons by name.
type weaponNameSelectors struct {
	damage   *WeaponBonus
	skill    *SkillBonus
	override *SelectorOverride
}

// addWeaponNameSelectors adds a trait carrying a weaponNameSelectors whose members select the weapons named name to
// the entity and returns it. The entity is not recalculated here.
func addWeaponNameSelectors(e *Entity, name string) *weaponNameSelectors {
	s := &weaponNameSelectors{
		damage:   NewWeaponBonus(feature.WeaponBonus),
		skill:    NewSkillBonus(),
		override: newDamageTypeOverride(wsel.WithName, "imp"),
	}
	s.damage.SelectionType = wsel.WithName
	s.damage.NameCriteria.Qualifier = name
	s.skill.SelectionType = skillsel.WeaponsWithName
	s.skill.NameCriteria.Qualifier = name
	s.override.NameCriteria.Qualifier = name
	addTraitWithFeatures(e, "Rules", s.damage, s.skill, s.override)
	return s
}

// owners returns the owner criteria of each selector.
func (s *weaponNameSelectors) owners() []*WeaponOwnerCriteria {
	return []*WeaponOwnerCriteria{
		&s.damage.WeaponOwnerCriteria, &s.skill.WeaponOwnerCriteria,
		&s.override.WeaponOwnerCriteria,
	}
}

// check verifies that each selector selects the weapon when want is true, and none do otherwise.
func (s *weaponNameSelectors) check(c check.Checker, e *Entity, w *Weapon, want bool, msg string) {
	c.Helper()
	usage := w.UsageWithReplacements()
	tags := w.Owner.TagList()
	var damages []*WeaponBonus
	e.addNamedWeaponBonusesFor(w, usage, w.Damage.resolvedType(nil), tags, func(bonus *WeaponBonus) { damages = append(damages, bonus) })
	c.Equal(want, slices.Contains(damages, s.damage), "damage bonus: "+msg)
	c.Equal(want, slices.Contains(e.NamedWeaponSkillBonusesFor(w, usage, w.Damage.resolvedType(nil), tags, nil), s.skill),
		"skill bonus: "+msg)
	c.Equal(want, resolvedDamageType(w, nil) == "imp", "selector override: "+msg)
}

// TestWeaponsSelectedByEquipmentName verifies that features selecting weapons by name match the name of a piece of
// equipment without its level, and that the owner level and TL criteria match its level and TL.
func TestWeaponsSelectedByEquipmentName(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	selectors := addWeaponNameSelectors(e, "Broadsword")
	sword := addCarriedEquipmentWithFeatures(e, "Broadsword")
	sword.Level = fxp.One
	sword.TechLevel = "3"
	w := newCrushingWeapon(sword, "Swung")
	sword.Weapons = []*Weapon{w}
	e.Recalculate()
	c.Equal("Broadsword 1", w.String(), "precondition: the weapon's name has the level")
	selectors.check(c, e, w, true, "the name matches without the level")
	for _, one := range selectors.owners() {
		one.OwnerLevelCriteria = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.One}
		one.OwnerTechLevelCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "3"}
	}
	selectors.check(c, e, w, true, "the level and TL match")
	c.Equal(3, selectors.override.OverrideSpecificity(), "the owner criteria in use count toward specificity")
	for _, one := range selectors.owners() {
		one.OwnerLevelCriteria.Qualifier = fxp.Two
	}
	selectors.check(c, e, w, false, "the level does not match")
	for _, one := range selectors.owners() {
		one.OwnerLevelCriteria = criteria.Number{}
		one.OwnerTechLevelCriteria.Qualifier = "4"
	}
	selectors.check(c, e, w, false, "the TL does not match")
}

// TestWeaponsSelectedByTraitName verifies that features selecting weapons by name match the name of a trait without its
// level, and that the owner level criterion matches the level of a leveled trait and 0 for any other.
func TestWeaponsSelectedByTraitName(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	selectors := addWeaponNameSelectors(e, "Innate Attack")
	attack := addTraitWithFeatures(e, "Innate Attack")
	attack.CanLevel = true
	attack.Levels = fxp.Two
	w := newCrushingWeapon(attack, "Burn")
	attack.Weapons = []*Weapon{w}
	e.Recalculate()
	c.Equal("Innate Attack 2", w.String(), "precondition: the weapon's name has the level")
	selectors.check(c, e, w, true, "the name matches without the level")
	for _, one := range selectors.owners() {
		one.OwnerLevelCriteria = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.Two}
	}
	selectors.check(c, e, w, true, "the level matches")
	attack.CanLevel = false
	e.Recalculate()
	c.Equal("Innate Attack", w.String(), "precondition: the trait is no longer leveled")
	selectors.check(c, e, w, false, "a trait that isn't leveled has level 0")
}

// TestWeaponsSelectedBySpellName verifies that features selecting weapons by name match the name of a spell without its
// TL, that the owner TL criterion matches its TL, and that the owner level criterion sees 0, not its skill level.
func TestWeaponsSelectedBySpellName(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	selectors := addWeaponNameSelectors(e, "Fireball")
	sp := NewSpell(e, nil, false)
	sp.Name = "Fireball"
	tl := "3"
	sp.TechLevel = &tl
	sp.Points = fxp.Four
	w := newCrushingWeapon(sp, "Thrown")
	sp.Weapons = []*Weapon{w}
	e.Spells = append(e.Spells, sp)
	e.Recalculate()
	c.Equal("Fireball/TL3", w.String(), "precondition: the weapon's name has the TL")
	selectors.check(c, e, w, true, "the name matches without the TL")
	for _, one := range selectors.owners() {
		one.OwnerTechLevelCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "3"}
	}
	selectors.check(c, e, w, true, "the TL matches")
	for _, one := range selectors.owners() {
		one.OwnerLevelCriteria = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.One}
	}
	selectors.check(c, e, w, false, "a spell has level 0")
}

// TestWeaponsSelectedByNameFromTheirOwner verifies that a weapon bonus selecting weapons by name applies to the weapons
// of its own owner by the owner's name and owner criteria, where the entity doesn't carry the bonus, as for equipment
// that isn't equipped.
func TestWeaponsSelectedByNameFromTheirOwner(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	bonus := NewWeaponBonus(feature.WeaponBonus)
	bonus.SelectionType = wsel.WithName
	bonus.NameCriteria.Qualifier = "Broadsword"
	bonus.OwnerLevelCriteria = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.One}
	sword := addCarriedEquipmentWithFeatures(e, "Broadsword", bonus)
	sword.Level = fxp.One
	sword.Equipped = false
	w := newCrushingWeapon(sword, "Swung")
	sword.Weapons = []*Weapon{w}
	e.Recalculate()
	var carried []*WeaponBonus
	e.addNamedWeaponBonusesFor(w, w.UsageWithReplacements(), w.Damage.resolvedType(nil), sword.TagList(),
		func(one *WeaponBonus) { carried = append(carried, one) })
	c.False(slices.Contains(carried, bonus), "precondition: the entity doesn't carry the bonus")
	collect := func() []*WeaponBonus {
		return w.collectWeaponBonuses(func() fxp.Int { return fxp.One }, nil, feature.WeaponBonus)
	}
	c.True(slices.Contains(collect(), bonus), "the name matches without the level")
	bonus.OwnerLevelCriteria.Qualifier = fxp.Two
	e.Recalculate()
	c.False(slices.Contains(collect(), bonus), "the level does not match")
}

// TestSkillBonusesSelectingWeaponsByNameFromTheirOwner verifies that the skill bonuses selecting weapons by name that
// an item and its modifiers carry reach the item's weapons whether or not the item is equipped, and are counted once.
func TestSkillBonusesSelectingWeaponsByNameFromTheirOwner(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	newBonus := func(amount fxp.Int) *SkillBonus {
		bonus := NewSkillBonus()
		bonus.SelectionType = skillsel.WeaponsWithName
		bonus.NameCriteria.Qualifier = "Broadsword"
		bonus.Amount = amount
		return bonus
	}
	bonus := newBonus(fxp.Two)
	modBonus := newBonus(fxp.One)
	sword := addCarriedEquipmentWithFeatures(e, "Broadsword", bonus)
	mod := NewEquipmentModifier(e, nil, false)
	mod.Name = "Fine"
	mod.Features = Features{modBonus}
	sword.AddModifiers(mod)
	w := newCrushingWeapon(sword, "Swung")
	sword.Weapons = []*Weapon{w}
	adjustment := func() fxp.Int {
		e.Recalculate()
		with := w.skillLevelBaseAdjustment(e, nil)
		bonus.Amount, modBonus.Amount = 0, 0
		without := w.skillLevelBaseAdjustment(e, nil)
		bonus.Amount, modBonus.Amount = fxp.Two, fxp.One
		return with - without
	}
	c.True(sword.Equipped, "precondition: new equipment starts out equipped")
	c.Equal(fxp.Three, adjustment(), "an equipped item's bonuses are counted once")
	sword.Equipped = false
	c.Equal(fxp.Three, adjustment(), "an unequipped item's bonuses still reach its own weapons")
}

// TestUnequippedItemBonusesNameTheirOwner verifies that the skill and weapon bonuses an item and its modifiers carry
// reach the item's weapons with their owner set, where the entity never processed them, as for equipment that isn't
// equipped when loaded: a per-level amount scales by the item's level and the tooltip names the item, not "Unknown".
func TestUnequippedItemBonusesNameTheirOwner(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	newSkillBonus := func(sel skillsel.Type) *SkillBonus {
		bonus := NewSkillBonus()
		bonus.SelectionType = sel
		bonus.NameCriteria.Qualifier = "Broadsword"
		bonus.Amount = fxp.Two
		bonus.PerLevel = true
		return bonus
	}
	newWeaponBonus := func(sel wsel.Type) *WeaponBonus {
		bonus := NewWeaponBonus(feature.WeaponBonus)
		bonus.SelectionType = sel
		bonus.NameCriteria.Qualifier = "Broadsword"
		bonus.Amount = fxp.Two
		bonus.PerLevel = true
		return bonus
	}
	sword := NewEquipment(e, nil, false)
	sword.Name = "Broadsword"
	sword.Level = fxp.Two
	sword.Equipped = false
	sword.Features = Features{
		newSkillBonus(skillsel.ThisWeapon), newSkillBonus(skillsel.WeaponsWithName),
		newWeaponBonus(wsel.ThisWeapon), newWeaponBonus(wsel.WithName),
	}
	mod := NewEquipmentModifier(e, nil, false)
	mod.Name = "Fine"
	mod.Features = Features{newSkillBonus(skillsel.ThisWeapon), newWeaponBonus(wsel.ThisWeapon)}
	sword.AddModifiers(mod)
	w := newCrushingWeapon(sword, "Swung")
	sword.Weapons = []*Weapon{w}
	e.CarriedEquipment = append(e.CarriedEquipment, sword)
	e.Recalculate()
	c.False(sword.ReallyEquipped(), "precondition: the sword isn't equipped")
	var tooltip xbytes.InsertBuffer
	c.Equal(fxp.Twelve, w.skillLevelBaseAdjustment(e, &tooltip), "three skill bonuses of +2 per level at level 2")
	c.Equal(2, strings.Count(tooltip.String(), "\nBroadsword 2 [+4 (+2 per level)]"), "the item's own skill bonuses name it")
	c.Equal(1, strings.Count(tooltip.String(), "\nBroadsword 2 (Fine) [+4 (+2 per level)]"), "the modifier's skill bonus names both")
	c.NotContains(tooltip.String(), "Unknown")
	tooltip.Reset()
	var total fxp.Int
	bonuses := w.collectWeaponBonuses(oneDieCount, &tooltip, feature.WeaponBonus)
	for _, bonus := range bonuses {
		total += bonus.AdjustedAmountForWeapon(w)
	}
	c.Equal(3, len(bonuses), "the item's two weapon bonuses and the modifier's")
	c.Equal(fxp.Twelve, total, "three weapon bonuses of +2 per level at level 2")
	c.Equal(2, strings.Count(tooltip.String(), "\nBroadsword 2 [+4 (+2 per level) to damage]"), "the item's own weapon bonuses name it")
	c.Equal(1, strings.Count(tooltip.String(), "\nBroadsword 2 (Fine) [+4 (+2 per level) to damage]"), "the modifier's weapon bonus names both")
	c.NotContains(tooltip.String(), "Unknown")
}

// TestWeaponOwnerCriteriaHashAndNameables verifies that each owner criterion takes part in the hash of each feature
// selecting weapons by name, and that the nameable keys in the text criteria are offered for substitution.
func TestWeaponOwnerCriteriaHashAndNameables(t *testing.T) {
	c := check.New(t)
	selectors := addWeaponNameSelectors(NewEntity(), "Guns")
	features := []interface {
		Hashable
		FillWithNameableKeys(m, existing map[string]string)
	}{selectors.damage, selectors.skill, selectors.override}
	for i, one := range selectors.owners() {
		f := features[i]
		hashes := map[uint64]bool{Hash64(f): true}
		one.OwnerLevelCriteria = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.One}
		hashes[Hash64(f)] = true
		one.OwnerTechLevelCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "@TL@"}
		hashes[Hash64(f)] = true
		one.OwnerSpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "@Gun@"}
		hashes[Hash64(f)] = true
		one.OwnerOptionalSpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "@Kind@"}
		hashes[Hash64(f)] = true
		one.OwnerTitleNoteCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "@Note@"}
		hashes[Hash64(f)] = true
		c.Equal(6, len(hashes), "%T: each owner criterion changes the hash", f)
		m := make(map[string]string)
		f.FillWithNameableKeys(m, nil)
		for _, key := range []string{"TL", "Gun", "Kind", "Note"} {
			_, exists := m[key]
			c.True(exists, "%T offers %s", f, key)
		}
	}
}

// TestWeaponsSelectedBySkillName verifies that features selecting weapons by name match the name of a skill without
// its TL or specializations, and that the owner TL and specialization criteria match those, with the skill's
// substitutions applied.
func TestWeaponsSelectedBySkillName(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	selectors := addWeaponNameSelectors(e, "Beam Weapons")
	sk := NewSkill(e, nil, false)
	sk.Name = "Beam Weapons"
	sk.Specialization = "@Kind@"
	sk.Replacements = map[string]string{"Kind": "Pistol"}
	sk.OptionalSpecialization = "Laser"
	tl := "9"
	sk.TechLevel = &tl
	w := newCrushingWeapon(sk, "Shot")
	sk.Weapons = []*Weapon{w}
	e.Skills = append(e.Skills, sk)
	e.Recalculate()
	c.Equal("Beam Weapons/TL9 (Pistol; Laser)", w.String(),
		"precondition: the weapon's name has the TL and specializations")
	selectors.check(c, e, w, true, "the name matches without the TL or specializations")
	for _, one := range selectors.owners() {
		one.OwnerTechLevelCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "9"}
		one.OwnerSpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Pistol"}
		one.OwnerOptionalSpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Laser"}
	}
	selectors.check(c, e, w, true, "the TL and specializations match")
	for _, one := range selectors.owners() {
		one.OwnerSpecializationCriteria.Qualifier = "Rifle"
	}
	selectors.check(c, e, w, false, "the specialization does not match")
	for _, one := range selectors.owners() {
		one.OwnerSpecializationCriteria = criteria.Text{}
		one.OwnerOptionalSpecializationCriteria.Qualifier = "Maser"
	}
	selectors.check(c, e, w, false, "the optional specialization does not match")
}

// TestWeaponOwnerCriteriaJSON verifies that the owner criteria are stored as top-level keys of the feature, left out
// while unset, and survive a round-trip.
func TestWeaponOwnerCriteriaJSON(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	selectors := addWeaponNameSelectors(e, "Broadsword")
	features := Features{selectors.damage, selectors.skill, selectors.override}
	keys := []string{
		"owner_level", "owner_tech_level", "owner_specialization", "owner_optional_specialization", "owner_title_note",
	}
	stored := func() []map[string]any {
		data, err := jio.Marshal(features)
		c.NoError(err)
		var list []map[string]any
		c.NoError(jio.Unmarshal(data, &list))
		return list
	}
	for _, one := range stored() {
		for _, key := range keys {
			_, exists := one[key]
			c.False(exists, "%v leaves out an unset %s", one["type"], key)
		}
	}
	for _, one := range selectors.owners() {
		one.OwnerLevelCriteria = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.One}
		one.OwnerTechLevelCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "3"}
		one.OwnerSpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Pistol"}
		one.OwnerOptionalSpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Laser"}
		one.OwnerTitleNoteCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Fine"}
	}
	checkFeaturesJSONRoundTrip(c, features)
	for _, one := range stored() {
		for _, key := range keys {
			_, exists := one[key]
			c.True(exists, "%v stores %s at the top level", one["type"], key)
		}
	}
}

// TestScriptWeaponsMatchOwnerName verifies that a script weapon's name is its owner's name without its level or title
// notes, that its title keeps them, and that the scripts finding weapons by name match the name or the title, with
// weaponDamage taking a weapon whose title matches over one before it whose name does.
func TestScriptWeaponsMatchOwnerName(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	attack := addTraitWithFeatures(e, "Innate Attack", newTitleNoteFor("Fire"))
	attack.CanLevel = true
	attack.Levels = fxp.Two
	w := newCrushingWeapon(attack, "Burn")
	attack.Weapons = []*Weapon{w}
	e.Recalculate()
	c.Equal("Innate Attack 2 (Fire)", w.String(), "precondition: the weapon's name has the level and title note")
	resolve := func(expr string) string { return ResolveScript(e, ScriptSelfProvider{}, expr) }
	c.Equal("Innate Attack", resolve(`entity.weapons[0].name`))
	c.Equal("Innate Attack 2 (Fire)", resolve(`entity.weapons[0].title`))
	c.Equal("1", resolve(`entity.findWeapons(true, entity.weapons[0].name).length.toString()`))
	c.Equal("1", resolve(`entity.findWeapons(true, entity.weapons[0].title).length.toString()`))
	c.Equal("1d-2 cr", resolve(`entity.weaponDamage("Innate Attack", "Burn")`))
	c.Equal("1d-2 cr", resolve(`entity.weaponDamage("Innate Attack 2 (Fire)", "Burn")`))

	gun := func(specialization, damage string) *Skill {
		s := NewSkill(e, nil, false)
		s.Name = "Guns"
		s.Specialization = specialization
		tl := "8"
		s.TechLevel = &tl
		shoot := NewWeapon(s, false)
		shoot.Usage = "Shoot"
		shoot.Damage.Base = damage
		shoot.Damage.Type = "pi"
		s.Weapons = []*Weapon{shoot}
		e.Skills = append(e.Skills, s)
		return s
	}
	pistol := gun("Pistol", "1d+2")
	gun("Rifle", "7d")
	whip := NewWeapon(pistol, true)
	pistol.Weapons = append(pistol.Weapons, whip)
	builtIn := addTraitWithFeatures(e, "Guns")
	builtIn.Weapons = []*Weapon{NewWeapon(builtIn, false)}
	builtIn.Weapons[0].Damage.Base = "2d"
	e.Recalculate()
	c.Equal("1d+2 pi", resolve(`entity.weaponDamage("Guns/TL8 (Pistol)", "Shoot")`))
	c.Equal("7d pi", resolve(`entity.weaponDamage("Guns/TL8 (Rifle)", "Shoot")`))
	c.Equal("1", resolve(`entity.findWeapons(false, "Guns/TL8 (Rifle)").length.toString()`))
	c.Equal("2d cr", resolve(`entity.weaponDamage("Guns", "")`),
		"the built-in gun's title is Guns, though the pistol's melee weapon, listed first, is a Guns weapon too")
}

// TestOwnerlessWeaponName verifies that a weapon without an owner has an empty name and title, in scripts too, and that
// matching it by name doesn't panic.
func TestOwnerlessWeaponName(t *testing.T) {
	c := check.New(t)
	w := NewWeapon(nil, true)
	c.Equal("", w.Name())
	c.Equal("", w.Title())
	c.False(w.matchesName(criteria.Text{Compare: criteria.IsText, Qualifier: "Sword"}, &WeaponOwnerCriteria{}, nil))
	c.True(w.matchesName(criteria.Text{}, &WeaponOwnerCriteria{}, nil), "an empty criterion still matches")
	e := NewEntity()
	c.Equal("", ResolveScript(e, deferredNewScriptWeapon(w), "self.name"))
	c.Equal("", ResolveScript(e, deferredNewScriptWeapon(w), "self.title"))
}

// TestScriptWeaponAttachedTo verifies that a script weapon's attachedTo is its owner, whatever kind of owner it has.
func TestScriptWeaponAttachedTo(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	trait := addTraitWithFeatures(e, "Claws")
	eqp := NewEquipment(e, nil, false)
	eqp.Name = "Sword"
	sk := NewSkill(e, nil, false)
	sk.Name = "Karate"
	sp := NewSpell(e, nil, false)
	sp.Name = "Fireball"
	for _, owner := range []WeaponOwner{trait, eqp, sk, sp} {
		w := NewWeapon(owner, true)
		c.Equal(owner.NameWithReplacements(), ResolveScript(e, deferredNewScriptWeapon(w), "self.attachedTo.name"))
	}
	c.Equal("true", ResolveScript(e, deferredNewScriptWeapon(NewWeapon(nil, true)), "self.attachedTo === undefined"),
		"a weapon without an owner isn't attached to anything")
}
