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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/equipmentsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellmatch"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/traitsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/toolbox/v2/check"
)

// bracketed is the em func for the feature description tests, so they can see what is emphasized.
func bracketed(s string) string {
	return "[" + s + "]"
}

// TestFeatureDescribe verifies the description of each feature type, including which parts are passed through the
// emphasis func, which criteria are left out, and that a nameable marker shows its value where one is set and stays
// as it is where none is.
func TestFeatureDescribe(t *testing.T) {
	c := check.New(t)
	replacements := map[string]string{"Skill": "Streetwise"}
	skill := func(name string, amount fxp.Int) *gurps.SkillBonus {
		f := gurps.NewSkillBonus()
		f.NameCriteria.Qualifier = name
		f.Amount = amount
		return f
	}
	dr := func(amount fxp.Int, specialization string, locations ...string) *gurps.DRBonus {
		f := gurps.NewDRBonus()
		f.Amount = amount
		f.Specialization = specialization
		f.Locations = locations
		return f
	}
	weapon := func(featureType feature.Type, selection wsel.Type, name string) *gurps.WeaponBonus {
		f := gurps.NewWeaponBonus(featureType)
		f.SelectionType = selection
		f.NameCriteria.Qualifier = name
		return f
	}
	spell := func(match spellmatch.Type, name string) *gurps.SpellBonus {
		f := gurps.NewSpellBonus()
		f.SpellMatchType = match
		f.NameCriteria.Qualifier = name
		return f
	}

	skillPerLevel := skill("@Skill@", fxp.One)
	skillPerLevel.PerLevel = true
	skillUnset := skill("@Craft@", fxp.One)
	skillSpecialized := skill("Guns", fxp.One)
	skillSpecialized.SpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Pistol"}
	skillSpecialized.TagsCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Combat"}
	skillContains := skill("Sword", fxp.One)
	skillContains.NameCriteria.Compare = criteria.ContainsText
	skillMimicry := skill("Mimicry", fxp.One)
	skillMimicry.NameCriteria.Compare = criteria.ContainsText
	skillMimicry.SpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Animal Sounds"}
	skillClauses := skill("Guns", fxp.One)
	skillClauses.SpecializationCriteria = criteria.Text{Compare: criteria.StartsWithText, Qualifier: "Pis"}
	skillClauses.TagsCriteria = criteria.Text{Compare: criteria.DoesNotContainText, Qualifier: "Magic"}
	skillAny := skill("", fxp.One)
	skillAny.NameCriteria.Compare = criteria.AnyText
	skillThisWeapon := skill("", fxp.Two)
	skillThisWeapon.SelectionType = skillsel.ThisWeapon
	skillThisWeaponUsage := skill("", fxp.Two)
	skillThisWeaponUsage.SelectionType = skillsel.ThisWeapon
	skillThisWeaponUsage.SpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Thrown"}
	skillNamedWeapons := skill("Axe", fxp.One)
	skillNamedWeapons.SelectionType = skillsel.WeaponsWithName
	skillSwitchable := skill("Pickpocket", -fxp.Three)
	skillSwitchable.Switchable = true

	attribute := gurps.NewAttributeBonus(gurps.StrengthID)
	attribute.Amount = fxp.FromInteger(18)
	attribute.Limitation = stlimit.StrikingOnly
	attributeLimitationIgnored := gurps.NewAttributeBonus(gurps.DexterityID)
	attributeLimitationIgnored.Limitation = stlimit.StrikingOnly

	drOne := dr(fxp.Three, gurps.AllID, "skull")
	drAgainst := dr(-fxp.Two, "crushing", "vitals")
	drAll := dr(fxp.Two, gurps.AllID, gurps.AllID)
	drAll.PerLevel = true
	drSeveral := dr(fxp.One, gurps.AllID, "skull", "eye", "tail")
	drThisArmor := dr(fxp.One, gurps.AllID)
	drAgainstUnset := dr(fxp.One, "@Attack@", "torso")

	reaction := gurps.NewReactionBonus()
	reaction.Amount = -fxp.Two
	reaction.Situation = "from others except your own kind"
	reaction.Group = "Social Stigma"
	conditional := gurps.NewConditionalModifierBonus()
	conditional.Amount = -fxp.One
	conditional.Situation = "on any long task"
	conditionalNoSituation := gurps.NewConditionalModifierBonus()
	conditionalNoSituation.Situation = ""

	weaponPerDie := weapon(feature.WeaponBonus, wsel.ThisWeapon, "")
	weaponPerDie.Amount = -fxp.One
	weaponPerDie.PerDie = true
	weaponBySkill := weapon(feature.WeaponBonus, wsel.WithRequiredSkill, "Broadsword")
	weaponBySkill.PerDie = true
	weaponCriteria := weapon(feature.WeaponBonus, wsel.WithRequiredSkill, "Broadsword")
	weaponCriteria.SpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Rapier"}
	weaponCriteria.UsageCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Thrust"}
	weaponCriteria.RelativeLevelCriteria.Qualifier = fxp.Two
	weaponDice := weapon(feature.WeaponBonus, wsel.ThisWeapon, "")
	weaponDice.Dice, _ = gurps.ParseBonusDice("1d")
	weaponDice.Amount = fxp.Two
	weaponDice.Percent = true
	weaponPercent := weapon(feature.WeaponAccBonus, wsel.WithName, "Bow")
	weaponPercent.Amount = fxp.Ten
	weaponPercent.Percent = true
	weaponReach := weapon(feature.WeaponMaxReachBonus, wsel.ThisWeapon, "")
	weaponReach.PerLevel = true
	weaponMinST := weapon(feature.WeaponMinSTBonus, wsel.ThisWeapon, "")
	weaponMinST.Amount = -fxp.One
	weaponMinST.PerDie = true
	weaponSwitch := weapon(feature.WeaponSwitch, wsel.ThisWeapon, "")
	weaponSwitch.SwitchType = wswitch.Unbalanced
	weaponSwitch.SwitchTypeValue = true
	weaponSwitchNamed := weapon(feature.WeaponSwitch, wsel.WithName, "Spear")
	weaponSwitchNamed.SwitchType = wswitch.TwoHanded

	spellAll := gurps.NewSpellBonus()
	spellPowerSource := spell(spellmatch.PowerSource, "Elemental")
	spellPowerSource.PerLevel = true
	spellCollege := spell(spellmatch.CollegeName, "Fire")
	spellCollege.TagsCriteria = criteria.Text{Compare: criteria.ContainsText, Qualifier: "Ritual"}
	spellNotNamed := spell(spellmatch.Name, "Fire")
	spellNotNamed.NameCriteria.Compare = criteria.DoesNotContainText
	spellPoint := gurps.NewSpellPointBonus()
	spellPoint.SpellMatchType = spellmatch.Name
	spellPoint.NameCriteria.Qualifier = "Fireball"
	spellPoint.Amount = fxp.Two
	skillPoint := gurps.NewSkillPointBonus()
	skillPoint.NameCriteria.Qualifier = "@Skill@"
	trait := gurps.NewTraitBonus()
	trait.NameCriteria.Qualifier = "Magery"
	traitMaxLevel := gurps.NewTraitMaxLevelBonus()
	traitMaxLevelNamed := gurps.NewTraitMaxLevelBonus()
	traitMaxLevelNamed.SelectionType = traitsel.TraitWithName
	traitMaxLevelNamed.NameCriteria.Qualifier = "Magery"
	traitMaxLevelNamed.Amount = "10%"
	maxUses := gurps.NewEquipmentMaxUsesBonus()
	maxUses.SelectionType = equipmentsel.EquipmentWithName
	maxUses.NameCriteria.Compare = criteria.AnyText
	maxUses.Amount = "x2"
	maxUses.PerLevel = true

	costReduction := gurps.NewCostReduction(gurps.StrengthID)
	costReduction.Percentage = fxp.FromInteger(40)
	weightReduction := gurps.NewContainedWeightReduction()
	weightReduction.Reduction = "80%"
	override := gurps.NewSelectorOverride(selector.WeaponDamageType)
	override.Value = "cut"
	override.NameCriteria.Qualifier = "Axe"
	overrideFiltered := gurps.NewSelectorOverride(selector.WeaponDamageType)
	overrideFiltered.Value = "cut"
	overrideFiltered.NameCriteria.Qualifier = "Axe"
	overrideFiltered.UsageCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Swung"}
	overrideFiltered.TagsCriteria = criteria.Text{Compare: criteria.ContainsText, Qualifier: "Melee"}
	overrideTrait := gurps.NewSelectorOverride(selector.TraitFrequency)
	overrideTrait.NameCriteria.Compare = criteria.AnyText
	overrideTrait.Priority = 2
	unknown := gurps.NewUnknownFeature("x", []byte(`{"type":"x"}`))

	for _, one := range []struct {
		name     string
		feature  gurps.Feature
		expected string
	}{
		{"skill per level, marker set", skillPerLevel, "[+1] per level to skill [Streetwise]"},
		{"skill, marker unset", skillUnset, "[+1] to skill [@Craft@]"},
		{"skill specialized and tagged", skillSpecialized, "[+1] to skill [Guns] ([Pistol]) tagged [Combat]"},
		{"skill name contains", skillContains, `[+1] to skills whose name contains "[Sword]"`},
		{
			"skill name contains, specialized", skillMimicry,
			`[+1] to skills whose name contains "[Mimicry]" ([Animal Sounds])`,
		},
		{
			"skill clauses quoted", skillClauses,
			`[+1] to skill [Guns] with a specialization that starts with "[Pis]" with all tags that do not contain ` +
				`"[Magic]"`,
		},
		{"skill any name", skillAny, "[+1] to all skills"},
		{"skill this weapon", skillThisWeapon, "[+2] to this weapon's skill"},
		{
			"skill this weapon with usage", skillThisWeaponUsage,
			"[+2] to this weapon's skill when its usage is [Thrown]",
		},
		{"skill weapons with name", skillNamedWeapons, "[+1] to the skill of weapons named [Axe]"},
		{"skill switchable", skillSwitchable, "[-3] to skill [Pickpocket], only while switched on"},
		{"attribute limited", attribute, "[+18] to [ST] for striking only"},
		{"attribute limitation only for ST", attributeLimitationIgnored, "[+1] to [DX]"},
		{"DR one location", drOne, "[+3] DR to the [Skull]"},
		{"DR against", drAgainst, "[-2] DR to the [Vitals] against [crushing] attacks"},
		{"DR all locations", drAll, "[+2] per level DR to [all locations]"},
		{"DR several locations", drSeveral, "[+1] DR to [3 locations]: Skull, Eyes, tail"},
		{"DR this armor", drThisArmor, "[+1] DR to this armor"},
		{"DR against, marker unset", drAgainstUnset, "[+1] DR to the [Torso] against [@Attack@] attacks"},
		{"reaction", reaction, "[-2] to reactions from others except your own kind"},
		{"conditional", conditional, "[-1] on any long task"},
		{"conditional without situation", conditionalNoSituation, "[+1] as a conditional modifier"},
		{"weapon per die", weaponPerDie, "[-1] per die to this weapon's damage"},
		{"weapon by skill", weaponBySkill, "[+1] per die to the damage of weapons using [Broadsword]"},
		{
			"weapon by skill with criteria", weaponCriteria,
			"[+1] to the damage of weapons using [Broadsword] ([Rapier]) whose usage is [Thrust] whose relative " +
				"skill level is at least 2",
		},
		{"weapon dice suspend percent", weaponDice, "[+1d +2] to this weapon's damage"},
		{"weapon percent", weaponPercent, "[+10%] to the accuracy of weapons named [Bow]"},
		{"weapon per level", weaponReach, "[+1] per level to this weapon's maximum reach"},
		{"weapon minimum ST ignores per die", weaponMinST, "[-1] to this weapon's minimum ST"},
		{"weapon switch", weaponSwitch, "Sets this weapon's [Unbalanced] flag to [true]"},
		{"weapon switch named", weaponSwitchNamed, "Sets the [Two-handed] flag of weapons named [Spear] to [false]"},
		{"spell power source", spellPowerSource, "[+1] per level to spells of the [Elemental] power source"},
		{"spell all colleges", spellAll, "[+1] to all spells"},
		{
			"spell college tagged", spellCollege,
			`[+1] to spells of the [Fire] college with a tag that contains "[Ritual]"`,
		},
		{"spell name not matching", spellNotNamed, `[+1] to spells whose name does not contain "[Fire]"`},
		{"spell points", spellPoint, "[+2] points to spell [Fireball]"},
		{"skill point", skillPoint, "[+1] point to skill [Streetwise]"},
		{"trait", trait, "[+1] to the level of trait [Magery]"},
		{"trait maximum level", traitMaxLevel, "[+1] to the maximum level of this trait"},
		{"trait maximum level named", traitMaxLevelNamed, "[+10%] to the maximum level of trait [Magery]"},
		{"equipment maximum uses", maxUses, "[x2] per level to the maximum uses of all equipment"},
		{"cost reduction", costReduction, "Reduces the cost of [ST] by [40%]"},
		{"contained weight reduction", weightReduction, "Reduces the contained weight by [80%]"},
		{
			"selector override", override,
			"Sets [weapon damage type] to [cut] on weapons named [Axe] (priority 0)",
		},
		{
			"selector override with usage and tags", overrideFiltered,
			`Sets [weapon damage type] to [cut] on weapons named [Axe] whose usage is [Swung] with a tag that ` +
				`contains "[Melee]" (priority 0)`,
		},
		{
			"selector override with a value title", overrideTrait,
			"Sets [trait frequency of appearance] to [None required] on all traits (priority 2)",
		},
		{"unknown", unknown, `Unknown feature type "x"; it will be preserved, but ignored`},
	} {
		before := gurps.Hash64(one.feature)
		c.Equal(one.expected, one.feature.Describe(nil, replacements, bracketed), one.name)
		c.Equal(before, gurps.Hash64(one.feature), "%s: describing leaves the feature as it was", one.name)
	}
}

// TestFeatureDescribeNamesLocationsByBody verifies that a DR bonus names its locations as the body type of the entity
// names them, and falls back to the ID of a location that body lacks.
func TestFeatureDescribeNamesLocationsByBody(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	foreleg := gurps.NewHitLocation(entity, "")
	foreleg.LocID = "foreleg"
	foreleg.ChoiceName = "Forelegs"
	body := &gurps.Body{Locations: []*gurps.HitLocation{foreleg}}
	body.Update(entity)
	entity.SheetSettings.BodyType = body
	c.Nil(body.LookupLocationByID(entity, "skull"), "the body has no skull")
	f := gurps.NewDRBonus()
	f.Locations = []string{"foreleg", "skull"}
	c.Equal("[+1] DR to [2 locations]: Forelegs, skull", f.Describe(entity, nil, bracketed))
}

// TestFeatureDescribeNamesEveryWeaponBonus verifies that each weapon bonus type names what it changes.
func TestFeatureDescribeNamesEveryWeaponBonus(t *testing.T) {
	c := check.New(t)
	for _, one := range feature.Types {
		if !one.IsWeaponBonus() || one == feature.WeaponSwitch {
			continue
		}
		f := gurps.NewWeaponBonus(one)
		f.SelectionType = wsel.ThisWeapon
		text := f.Describe(nil, nil, bracketed)
		c.True(strings.HasPrefix(text, "[+1] to this weapon's "), "%s: %s", one.Key(), text)
		c.False(strings.HasSuffix(text, "'s "), "%s names what it changes", one.Key())
	}
}
