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
	"os"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/attack"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// TestWeaponJSONOmitsUnusedFields verifies that a weapon writes its parry and block only when it can parry and block,
// and never writes the fields of the other kind of weapon, even while it holds values for them.
func TestWeaponJSONOmitsUnusedFields(t *testing.T) {
	c := check.New(t)
	marshal := func(w *gurps.Weapon) string {
		c.Helper()
		data, err := jio.Marshal(w)
		c.NoError(err)
		return string(data)
	}

	melee := gurps.NewWeapon(nil, true)
	s := marshal(melee)
	c.NotContains(s, `"parry"`, "a melee weapon that can't parry writes no parry")
	c.NotContains(s, `"block"`, "a melee weapon that can't block writes no block")
	melee.Parry.CanParry = true
	melee.Block.CanBlock = true
	melee.Accuracy = attack.ParseAccuracy("3")
	s = marshal(melee)
	c.Contains(s, `"parry"`, "a melee weapon that can parry writes its parry")
	c.Contains(s, `"block"`, "a melee weapon that can block writes its block")
	c.NotContains(s, `"accuracy"`, "a melee weapon writes no accuracy")

	ranged := gurps.NewWeapon(nil, false)
	ranged.Parry.CanParry = true
	ranged.Block.CanBlock = true
	s = marshal(ranged)
	c.NotContains(s, `"parry"`, "a ranged weapon writes no parry")
	c.NotContains(s, `"block"`, "a ranged weapon writes no block")
}

// TestWeaponValidate verifies that validating a weapon drops the fields of the other kind of weapon, and makes both the
// accuracy and the rate of fire of a ranged weapon a jet when either one is.
func TestWeaponValidate(t *testing.T) {
	c := check.New(t)
	melee := gurps.NewWeapon(nil, true)
	melee.Accuracy = attack.ParseAccuracy("3")
	melee.Validate()
	c.Equal(attack.Accuracy{}, melee.Accuracy, "a melee weapon has no accuracy")

	ranged := gurps.NewWeapon(nil, false)
	ranged.Parry = attack.ParseParry("0")
	ranged.Validate()
	c.Equal(attack.Parry{}, ranged.Parry, "a ranged weapon has no parry")

	for _, accuracyJet := range []bool{true, false} {
		jet := gurps.NewWeapon(nil, false)
		jet.Accuracy.Jet = accuracyJet
		jet.RateOfFire.Jet = !accuracyJet
		jet.Validate()
		c.True(jet.Accuracy.Jet, "accuracy jet %v: the accuracy is a jet", accuracyJet)
		c.True(jet.RateOfFire.Jet, "accuracy jet %v: the rate of fire is a jet", accuracyJet)
	}
}

// TestExtractWeaponsOfType verifies that ExtractWeaponsOfType keeps the weapons of the kind asked for, in order,
// leaving out the hidden ones only when asked to.
func TestExtractWeaponsOfType(t *testing.T) {
	c := check.New(t)
	named := func(name string, melee bool) *gurps.Weapon {
		owner := gurps.NewTrait(nil, nil, false)
		owner.Name = name // A weapon prints as its owner's name
		return gurps.NewWeapon(owner, melee)
	}
	melee := named("Sword", true)
	ranged := named("Bow", false)
	hidden := named("Dagger", true)
	hidden.Hide = true
	list := []*gurps.Weapon{melee, ranged, hidden}
	c.Equal([]*gurps.Weapon{melee, hidden}, gurps.ExtractWeaponsOfType(true, false, list), "melee, hidden included")
	c.Equal([]*gurps.Weapon{melee}, gurps.ExtractWeaponsOfType(true, true, list), "melee, hidden left out")
	c.Equal([]*gurps.Weapon{ranged}, gurps.ExtractWeaponsOfType(false, true, list), "ranged")
}

// TestWeaponWithNullDefaultEntry verifies that a JSON null in a weapon's "defaults" array is dropped rather than
// dereferenced, mirroring the handling in Skill.UnmarshalJSONFrom. Such an entry decodes into a nil pointer without
// error, and the walkers over the defaults -- skill level resolution, hashing, nameable extraction -- dereference it.
func TestWeaponWithNullDefaultEntry(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	addDXSkill(e, "Broadsword", difficulty.Average, fxp.FromInteger(4)) // DX+1, i.e. level 11
	e.Recalculate()

	owner := gurps.NewTrait(e, nil, false)
	owner.Name = "Sword"
	var w gurps.Weapon
	c.NoError(jio.Unmarshal([]byte(`{"type":"melee_weapon","usage":"Swung",`+
		`"defaults":[null,{"type":"skill","name":{"compare":"is","qualifier":"Broadsword"},"modifier":-2}]}`), &w),
		"a weapon with a null default entry should load")
	c.Equal(1, len(w.Defaults), "the null entry was dropped and the usable one kept")
	w.Owner = owner
	owner.Weapons = []*gurps.Weapon{&w}
	e.Traits = append(e.Traits, owner)
	e.Recalculate()

	// The surviving default still resolves, so the null entry cost nothing but itself.
	c.Equal(fxp.FromInteger(9), w.SkillLevel(nil), "the weapon defaults to Broadsword-2")

	// The other walkers over the defaults must be equally safe.
	c.NotPanics(func() { gurps.Hash64(&w) }, "hashing must not panic")
	c.NotPanics(func() { w.HashDisplayedState() }, "hashing the displayed state must not panic")
	c.NotPanics(func() { w.FillWithNameableKeys(make(map[string]string), nil) },
		"nameable extraction must not panic")
	c.NotPanics(func() { w.Clone(gurps.LibraryFile{}, e, nil, gurps.Reference) }, "cloning must not panic")
}

// TestWeaponColumnHasData verifies that ColumnHasData reports hideable columns as empty until they hold meaningful data,
// and always reports non-hideable columns as having data. See issue #161.
func TestWeaponColumnHasData(t *testing.T) {
	c := check.New(t)

	// Non-hideable columns always report data, even for a blank weapon.
	melee := gurps.NewWeapon(nil, true)
	for _, columnID := range []int{
		gurps.WeaponDescriptionColumn,
		gurps.WeaponSLColumn,
		gurps.WeaponDamageColumn,
	} {
		c.True(melee.ColumnHasData(columnID), "column %d should always report data", columnID)
	}

	// The usage column is hideable when empty.
	c.False(melee.ColumnHasData(gurps.WeaponUsageColumn))
	melee.Usage = "Swung"
	c.True(melee.ColumnHasData(gurps.WeaponUsageColumn))

	// A blank ranged weapon (with rate of fire cleared) uses none of its hideable columns.
	ranged := gurps.NewWeapon(nil, false)
	ranged.RateOfFire = attack.RoF{}
	for _, columnID := range []int{
		gurps.WeaponAccColumn,
		gurps.WeaponRangeColumn,
		gurps.WeaponRoFColumn,
		gurps.WeaponShotsColumn,
		gurps.WeaponBulkColumn,
		gurps.WeaponRecoilColumn,
		gurps.WeaponSTColumn,
	} {
		c.False(ranged.ColumnHasData(columnID), "column %d should be empty for a blank ranged weapon", columnID)
	}

	// Populating each field flips the corresponding column to "has data".
	ranged.Accuracy.Base = fxp.Three
	c.True(ranged.ColumnHasData(gurps.WeaponAccColumn))
	ranged.Range.Max = fxp.FromInteger(100)
	c.True(ranged.ColumnHasData(gurps.WeaponRangeColumn))
	ranged.RateOfFire.Mode1.ShotsPerAttack = fxp.One
	c.True(ranged.ColumnHasData(gurps.WeaponRoFColumn))
	ranged.Shots.Count = fxp.FromInteger(8)
	c.True(ranged.ColumnHasData(gurps.WeaponShotsColumn))
	ranged.Bulk.Normal = fxp.FromInteger(-2)
	c.True(ranged.ColumnHasData(gurps.WeaponBulkColumn))
	ranged.Recoil.Shot = fxp.Two
	c.True(ranged.ColumnHasData(gurps.WeaponRecoilColumn))
	ranged.Strength.Min = fxp.FromInteger(10)
	c.True(ranged.ColumnHasData(gurps.WeaponSTColumn))

	// A jet's accuracy is shown, though it has no number.
	ranged.Accuracy = attack.Accuracy{Jet: true}
	c.True(ranged.ColumnHasData(gurps.WeaponAccColumn), "a jet has accuracy")

	// Melee-specific columns behave the same way.
	c.False(melee.ColumnHasData(gurps.WeaponParryColumn))
	c.False(melee.ColumnHasData(gurps.WeaponBlockColumn))
	melee.Parry.CanParry = true
	melee.Block.CanBlock = true
	c.True(melee.ColumnHasData(gurps.WeaponParryColumn))
	c.True(melee.ColumnHasData(gurps.WeaponBlockColumn))
	c.True(melee.ColumnHasData(gurps.WeaponReachColumn))
	melee.Reach = attack.Reach{}
	c.False(melee.ColumnHasData(gurps.WeaponReachColumn), "a melee weapon without reach")
}

// TestWeaponHashDisplayedStateIncludesHide verifies that the hidden state participates in the displayed-state hash.
// Without it, a hidden weapon and an otherwise display-identical visible one collapse into a single entry in the
// weapon maps built by Entity.Weapons, making one of them unreachable in the views where the Hide toggle is shown.
func TestWeaponHashDisplayedStateIncludesHide(t *testing.T) {
	c := check.New(t)

	trait := gurps.NewTrait(gurps.NewEntity(), nil, false)
	trait.Name = "Claws"
	w1 := gurps.NewWeapon(trait, true)
	w2 := gurps.NewWeapon(trait, true)
	c.Equal(w1.HashDisplayedState(), w2.HashDisplayedState(),
		"identically displayed weapons should hash the same")

	w2.Hide = true
	c.NotEqual(w1.HashDisplayedState(), w2.HashDisplayedState(),
		"weapons differing only in their hidden state should hash differently")
}

// TestWeaponHideIsNotSourceData verifies that the hidden state stays out of the source-data hash used by library
// syncing. Hide is a local display choice, so toggling it must not make a node report as out of sync with its library
// source, nor may a sync overwrite it.
func TestWeaponHideIsNotSourceData(t *testing.T) {
	c := check.New(t)

	trait := gurps.NewTrait(gurps.NewEntity(), nil, false)
	trait.Name = "Claws"
	w := gurps.NewWeapon(trait, true)
	trait.Weapons = []*gurps.Weapon{w}

	weaponHash := gurps.Hash64(&w.WeaponData)
	traitHash := gurps.Hash64(trait)

	w.Hide = true
	c.Equal(weaponHash, gurps.Hash64(&w.WeaponData), "hiding a weapon should not alter its source-data hash")
	c.Equal(traitHash, gurps.Hash64(trait), "hiding a weapon should not alter its owner's source-data hash")
}

// addDXSkill adds a DX-based skill of the given difficulty, points and tags to the entity's skills and returns it. The
// entity is not recalculated.
func addDXSkill(e *gurps.Entity, name string, diff difficulty.Level, points fxp.Int, tags ...string) *gurps.Skill {
	sk := gurps.NewSkill(e, nil, false)
	sk.Name = name
	sk.Tags = tags
	sk.Difficulty.Attribute = gurps.DexterityID
	sk.Difficulty.Difficulty = diff
	sk.Points = points
	e.Skills = append(e.Skills, sk)
	return sk
}

// newWeaponWithBonuses builds an entity with a trait that owns a single weapon of the requested type and carries the
// given weapon bonuses, each scoped to that weapon.
func newWeaponWithBonuses(melee bool, bonuses ...*gurps.WeaponBonus) *gurps.Weapon {
	e := gurps.NewEntity()
	owner := gurps.NewTrait(e, nil, false)
	owner.Name = "Gadget"
	w := gurps.NewWeapon(owner, melee)
	owner.Weapons = []*gurps.Weapon{w}
	for _, bonus := range bonuses {
		bonus.SelectionType = wsel.ThisWeapon
		owner.Features = append(owner.Features, bonus)
	}
	e.Traits = append(e.Traits, owner)
	e.Recalculate()
	return w
}

// newSwitchProbeWeapon builds a weapon of the requested kind carrying the given switch bonus, with every field a
// weapon switch could act on populated so that flipping the switch has something visible to change.
func newSwitchProbeWeapon(melee bool, sw *gurps.WeaponBonus) *gurps.Weapon {
	w := newWeaponWithBonuses(melee, sw)
	w.Strength = attack.ParseStrength("10")
	if melee {
		w.Parry = attack.ParseParry("0")
		w.Block = attack.ParseBlock("0")
		w.Reach = attack.ParseReach("1-2")
	} else {
		w.Accuracy = attack.ParseAccuracy("3")
		w.Range = attack.ParseRange("100/1000")
		w.RateOfFire = attack.ParseRoF("3/2")
		w.Shots = attack.ParseShots("20(3)")
		w.Bulk = attack.ParseBulk("-4")
	}
	return w
}

// weaponSwitchProbe renders every resolved field a weapon switch could affect, so a test can tell whether flipping a
// switch changed anything at all. Both weapon kinds are rendered; the fields belonging to the other kind are simply
// empty.
func weaponSwitchProbe(w *gurps.Weapon) string {
	return strings.Join([]string{
		w.ResolvedParry(nil).String(),
		w.ResolvedBlock(nil).String(),
		w.ResolvedReach(nil).String(),
		w.ResolvedAccuracy(nil).String(),
		w.ResolvedRange(nil).String(false),
		w.ResolvedRateOfFire(nil).String(),
		w.ResolvedShots(nil).String(),
		w.ResolvedBulk(nil).String(),
		w.ResolvedStrength(nil).String(),
	}, "|")
}

// TestEveryWeaponSwitchIsResolved verifies that each wswitch value the feature editor offers actually reaches a
// ResolveBoolFlag() call. Every value is presented to the user (ux/features_panel.go offers wswitch.Types[1:]), so one
// with no resolution path is a feature that silently does nothing -- which is what RetractingStock was. A switch is
// considered wired if setting it changes some resolved field on either a melee or a ranged weapon.
func TestEveryWeaponSwitchIsResolved(t *testing.T) {
	c := check.New(t)
	for _, switchType := range wswitch.Types[1:] {
		wired := false
		for _, melee := range []bool{true, false} {
			sw := gurps.NewWeaponBonus(feature.WeaponSwitch)
			sw.SwitchType = switchType
			sw.SwitchTypeValue = true
			w := newSwitchProbeWeapon(melee, sw)
			on := weaponSwitchProbe(w)
			sw.SwitchTypeValue = false
			w.Entity().DiscardCaches()
			if on != weaponSwitchProbe(w) {
				wired = true
			}
		}
		c.True(wired, "weapon switch %q is offered by the editor but never resolved", switchType.Key())
	}
}

// newDefenseTestWeapon builds an entity holding a "Cloak" skill at level 14 and a +1 bonus to both parry and block,
// along with a trait carrying a melee weapon that can both parry and block. The weapon has no defaults; the caller
// supplies whichever ones the test needs.
func newDefenseTestWeapon(c check.Checker) *gurps.Weapon {
	return newDefenseTestWeaponWithBonuses(c, 1, 1)
}

// newDefenseTestWeaponWithBonuses is newDefenseTestWeapon with the parry and block bonuses stated separately, so that a
// test can tell which of the two a defense calculation actually applied.
func newDefenseTestWeaponWithBonuses(c check.Checker, parryBonus, blockBonus int) *gurps.Weapon {
	e := gurps.NewEntity()

	// DX+4, i.e. level 14 with the default DX of 10
	sk := addDXSkill(e, "Cloak", difficulty.Average, fxp.FromInteger(16))

	bonuses := gurps.NewTrait(e, nil, false)
	bonuses.Name = "Improved Defenses"
	for _, one := range []struct {
		attrID string
		amount int
	}{{gurps.ParryID, parryBonus}, {gurps.BlockID, blockBonus}} {
		bonus := gurps.NewAttributeBonus(one.attrID)
		bonus.Amount = fxp.FromInteger(one.amount)
		bonuses.Features = append(bonuses.Features, bonus)
	}
	e.Traits = append(e.Traits, bonuses)

	owner := gurps.NewTrait(e, nil, false)
	owner.Name = "Cloak"
	w := gurps.NewWeapon(owner, true)
	w.Parry = attack.Parry{CanParry: true}
	w.Block = attack.Block{CanBlock: true}
	owner.Weapons = []*gurps.Weapon{w}
	e.Traits = append(e.Traits, owner)

	e.Recalculate()
	c.Equal(fxp.FromInteger(14), sk.LevelData.Level, "the Cloak skill should be at level 14")
	c.Equal(fxp.FromInteger(parryBonus), e.ParryBonus, "the entity should have a +%d parry bonus", parryBonus)
	c.Equal(fxp.FromInteger(blockBonus), e.BlockBonus, "the entity should have a +%d block bonus", blockBonus)
	return w
}

// TestWeaponSkillLevelIgnoresDefenseTypeDefaults verifies that a parry- or block-type weapon default does not become
// the weapon's attack skill. Such a default is there for the defenses (the Tonfa in the High Tech library carries
// "Brawling Parry" and "Karate Parry" alongside its real attack defaults), so it stays scored by its halved defense
// level rather than by the level of the skill it names, without the multipliers the defense takes.
func TestWeaponSkillLevelIgnoresDefenseTypeDefaults(t *testing.T) {
	c := check.New(t)
	w := newDefenseTestWeapon(c)
	e := w.Entity()

	// Stand in for the Tonfa: a weaker skill the weapon really attacks with, alongside the strong Cloak-14 that the
	// weapon only names for its parry.
	addDXSkill(e, "Tonfa", difficulty.Average, fxp.FromInteger(12)) // DX+3, i.e. level 13
	e.Recalculate()

	tonfaDefault := newDefenseTestDefault(gurps.SkillID)
	tonfaDefault.Name.Qualifier = "Tonfa"
	w.Defaults = []*gurps.SkillDefault{tonfaDefault, newDefenseTestDefault(gurps.ParryID)}

	// The attack stays on Tonfa-13. Cloak's parry level of 11 loses the max, as intended; Cloak-14 would have won it.
	c.Equal(fxp.FromInteger(13), w.SkillLevel(nil), "a parry-type default must not take over the attack skill")

	// The parry, meanwhile, is the better of Tonfa's (13/2 + 3 + 1) and Cloak's (14/2 + 3 + 1).
	c.Equal("11", w.ResolvedParry(nil).String(), "the parry-type default still feeds the parry")

	doubled := gurps.NewTrait(e, nil, false)
	doubled.Name = "Doubled Parry"
	multiplier := gurps.NewAttributeBonus(gurps.ParryID)
	multiplier.Multiply = true
	multiplier.Amount = fxp.Two
	doubled.Features = gurps.Features{multiplier}
	e.Traits = append(e.Traits, doubled)
	e.Recalculate()
	c.Equal(fxp.FromInteger(13), w.SkillLevel(nil), "a multiplied parry must not take over the attack skill either")
	c.Equal("22", w.ResolvedParry(nil).String(), "while the parry is multiplied")
}

// newDefenseTestDefault creates a skill default of the given type that matches the "Cloak" skill created by
// newDefenseTestWeapon.
func newDefenseTestDefault(defaultType string) *gurps.SkillDefault {
	return &gurps.SkillDefault{
		DefaultType: defaultType,
		Name: criteria.Text{
			TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: "Cloak"},
		},
	}
}

// TestTboneProgressionSTDamageContribution verifies that a Tbone damage progression still contributes its ST-based
// damage when combined with a base damage that uses dice with a different number of sides. addDice weights each
// operand's average by its dice Multiplier, so a progression that left Multiplier at 0 silently erased the entire ST
// contribution.
func TestTboneProgressionSTDamageContribution(t *testing.T) {
	c := check.New(t)
	for i, one := range []struct {
		prog     progression.Option
		expected string
	}{
		{progression.BasicSet, "2d3 cr"},
		{progression.Tbone1, "2d3+2 cr"},
		{progression.Tbone1Clean, "2d3+2 cr"},
		{progression.Tbone2, "2d3 cr"},
		{progression.Tbone2Clean, "2d3 cr"},
	} {
		e := gurps.NewEntity()
		e.SheetSettings.DamageProgression = one.prog
		trait := gurps.NewTrait(e, nil, false)
		trait.Name = "Claws"
		w := gurps.NewWeapon(trait, true)
		w.Damage.Base = "1d3"
		trait.Weapons = append(trait.Weapons, w)
		e.Traits = append(e.Traits, trait)
		e.Recalculate()
		c.Equal(one.expected, w.Damage.ResolvedDamage(nil), "test %d: %s", i, one.prog.Key())
	}
}

// TestPreSubVersionWeaponDamageIsMigrated verifies that weapon data written before sub-versioning existed still has its
// base damage moved into the per-level slot when its owner is leveled, while a freshly created weapon is stamped with
// the current data sub-version so that this migration is not run against it. Without the stamp, the first SetOwner
// call on a weapon owned by a leveled trait or piece of equipment moved the base damage into the per-level slot,
// turning a new weapon's "1d" into "1d per level".
func TestPreSubVersionWeaponDamageIsMigrated(t *testing.T) {
	for _, tc := range []struct {
		name          string
		preSubVersion bool
		base          string
		baseLeveled   string
		baseMsg       string
		baseLevelMsg  string
	}{
		{
			name:         "new weapon",
			base:         "1d",
			baseMsg:      "a new weapon's base damage must not be migrated into the per-level slot",
			baseLevelMsg: "a new weapon must not acquire per-level base damage",
		},
		{
			name:          "pre-sub-versioning data",
			preSubVersion: true,
			baseLeveled:   "1d",
			baseMsg:       "migrated damage must leave the base slot",
			baseLevelMsg:  "pre-sub-versioning damage on a leveled owner must migrate to per-level",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)

			e := gurps.NewEntity()
			trait := gurps.NewTrait(e, nil, false)
			trait.Name = "Innate Attack"
			trait.CanLevel = true
			trait.Levels = fxp.Three
			c.True(trait.IsLeveled())

			w := gurps.NewWeapon(trait, false)
			c.Equal("1d", w.Damage.Base)
			if tc.preSubVersion {
				w.SubVersion = 0 // Simulate data saved prior to sub-versioning.
			}
			trait.Weapons = []*gurps.Weapon{w}
			e.Traits = append(e.Traits, trait)
			e.Recalculate()

			c.Equal(tc.base, w.Damage.Base, tc.baseMsg)
			c.Equal(tc.baseLeveled, w.Damage.BaseLeveled, tc.baseLevelMsg)
		})
	}
}

// TestWeaponEditorPreservesID verifies that applying an editor's copy back onto the weapon it was taken from leaves the
// weapon's identity alone. A Weapon acts as its own editor data, so the copy the editor holds carries a TID of its own,
// but the weapon's TID is persisted as its "id" and table selections are keyed on it, so neither applying an edit nor
// undoing or redoing one may change it. The damage of the copy and of the weapon must also stay their own, rather than
// belong to the clone each was made from.
func TestWeaponEditorPreservesID(t *testing.T) {
	c := check.New(t)

	w := gurps.NewWeapon(nil, true)
	w.Usage = "Swung"
	w.ClonedFromTID = gurps.NewWeapon(nil, true).TID // Stand in for the weapon this one was cloned from.
	originalTID := w.TID
	originalClonedFromTID := w.ClonedFromTID

	// The editor holds two copies: one to restore on undo and one the user edits.
	var beforeData gurps.Weapon
	beforeData.CopyFrom(w)
	var editorData gurps.Weapon
	editorData.CopyFrom(w)
	c.True(editorData.Damage.Owner == &editorData, "the copy's damage belongs to the copy")
	editorData.Usage = "Thrust"

	editorData.ApplyTo(w)
	c.Equal("Thrust", w.Usage, "the edit was applied")
	c.True(w.Damage.Owner == w, "the applied damage belongs to the weapon")
	c.Equal(originalTID, w.TID, "applying an edit doesn't change the weapon's ID")
	c.Equal(originalClonedFromTID, w.ClonedFromTID, "applying an edit doesn't change the weapon's lineage")

	beforeData.ApplyTo(w)
	c.Equal("Swung", w.Usage, "the edit was undone")
	c.Equal(originalTID, w.TID, "undoing an edit doesn't change the weapon's ID")
	c.Equal(originalClonedFromTID, w.ClonedFromTID, "undoing an edit doesn't change the weapon's lineage")

	editorData.ApplyTo(w)
	c.Equal("Thrust", w.Usage, "the edit was redone")
	c.Equal(originalTID, w.TID, "redoing an edit doesn't change the weapon's ID")
	c.Equal(originalClonedFromTID, w.ClonedFromTID, "redoing an edit doesn't change the weapon's lineage")
}

// TestWeaponPerDieBonusTooltip verifies that the tooltip written for a per-die weapon bonus reports the amount that is
// actually applied, which scales by the weapon's base damage dice, rather than always reporting the single-die amount.
// A per-die minimum ST bonus is the exception: resolving the base damage dice requires the minimum ST, so those dice
// can't be consulted for it, and AdjustedAmountForWeapon deliberately pins it to one die; its tooltip must say so too.
func TestWeaponPerDieBonusTooltip(t *testing.T) {
	c := check.New(t)

	acc := gurps.NewWeaponBonus(feature.WeaponAccBonus)
	acc.PerDie = true
	bulk := gurps.NewWeaponBonus(feature.WeaponBulkBonus)
	bulk.PerDie = true
	bulk.Amount = fxp.NegOne
	minST := gurps.NewWeaponBonus(feature.WeaponMinSTBonus)
	minST.PerDie = true
	w := newWeaponWithBonuses(false, acc, bulk, minST)
	w.Damage.Base = "3d"
	w.Accuracy = attack.ParseAccuracy("3")
	w.Bulk = attack.ParseBulk("-4")
	w.Strength = attack.ParseStrength("10")

	var accTooltip xbytes.InsertBuffer
	c.Equal("6", w.ResolvedAccuracy(&accTooltip).String(), "+1 per die on a 3d weapon raises Acc 3 to 6")
	c.Contains(accTooltip.String(), "+3 (+1 per die) to accuracy", "the tooltip reports the amount applied")

	// The same holds for the other stats, which share the collection that writes these tooltips.
	var bulkTooltip xbytes.InsertBuffer
	c.Equal("-7", w.ResolvedBulk(&bulkTooltip).String(), "-1 per die on a 3d weapon takes Bulk -4 to -7")
	c.Contains(bulkTooltip.String(), "-3 (-1 per die) to bulk", "the tooltip reports the amount applied")

	var stTooltip xbytes.InsertBuffer
	c.Equal("11", w.ResolvedStrength(&stTooltip).String(), "a per-die minimum ST bonus counts as a single die")
	c.Contains(stTooltip.String(), "+1 (+1 per die) to minimum ST", "the tooltip reports the amount applied")
}

// TestWeaponPerLevelBonusFromTraitModifier verifies that a per-level weapon bonus carried by one of a trait's
// modifiers scales with the modifier's level rather than the trait's, and that the tooltip reports the same amount
// that is applied. A modifier with no levels of its own contributes nothing, even when the trait it hangs off of is
// leveled; the tooltip previously claimed the trait's level while the value stayed put.
func TestWeaponPerLevelBonusFromTraitModifier(t *testing.T) {
	c := check.New(t)

	e := gurps.NewEntity()
	trait := gurps.NewTrait(e, nil, false)
	trait.Name = "Innate Attack"
	trait.CanLevel = true
	trait.Levels = fxp.Three
	acc := gurps.NewWeaponBonus(feature.WeaponAccBonus)
	acc.SelectionType = wsel.ThisWeapon
	acc.PerLevel = true
	mod := gurps.NewTraitModifier(e, nil, false)
	mod.Name = "Targeting"
	mod.Features = gurps.Features{acc}
	trait.Modifiers = []*gurps.TraitModifier{mod}
	w := gurps.NewWeapon(trait, false)
	w.Accuracy = attack.ParseAccuracy("3")
	trait.Weapons = []*gurps.Weapon{w}
	e.Traits = append(e.Traits, trait)
	e.Recalculate()

	var tooltip xbytes.InsertBuffer
	c.Equal("3", w.ResolvedAccuracy(&tooltip).String(),
		"a per-level bonus on a modifier with no levels adds nothing, no matter the trait's level")
	c.Equal("\nInnate Attack 3 (Targeting) [+0 (+1 per level) to accuracy]", tooltip.String(),
		"the tooltip reports the amount applied rather than scaling by the trait's level")

	// Giving the modifier levels of its own is what makes the bonus count.
	mod.Levels = fxp.Two
	e.Recalculate()
	tooltip.Reset()
	c.Equal("5", w.ResolvedAccuracy(&tooltip).String(), "+1 per level on a level 2 modifier raises Acc 3 to 5")
	c.Equal("\nInnate Attack 3 (Targeting 2) [+2 (+1 per level) to accuracy]", tooltip.String(),
		"the tooltip names the modifier with its level and agrees with the amount applied")

	// A "use level from owner" modifier's bonus scales by the trait's level, which the trait already shows.
	mod.Levels = 0
	mod.UseLevelFromTrait = true
	e.Recalculate()
	tooltip.Reset()
	c.Equal("6", w.ResolvedAccuracy(&tooltip).String(),
		"a modifier that takes the trait's level of 3 raises Acc 3 to 6")
	c.Equal("\nInnate Attack 3 (Targeting) [+3 (+1 per level) to accuracy]", tooltip.String(),
		"the tooltip agrees with the amount applied when the level comes from the trait")
}

// newTagDefaultTestWeapon builds an entity holding a Combat-tagged Broadsword skill, a trait granting +2 damage to any
// weapon whose required skill is named Broadsword, and a trait-owned melee weapon whose only default is the one given.
func newTagDefaultTestWeapon(c check.Checker, def *gurps.SkillDefault) *gurps.Weapon {
	e := gurps.NewEntity()

	// DX+1, i.e. level 11 with the default DX of 10
	sk := addDXSkill(e, "Broadsword", difficulty.Average, fxp.Four, "Combat")

	master := gurps.NewTrait(e, nil, false)
	master.Name = "Weapon Master"
	bonus := gurps.NewWeaponBonus(feature.WeaponBonus)
	bonus.SelectionType = wsel.WithRequiredSkill
	bonus.NameCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Broadsword"}
	bonus.Amount = fxp.Two
	master.Features = append(master.Features, bonus)
	e.Traits = append(e.Traits, master)

	owner := gurps.NewTrait(e, nil, false)
	owner.Name = "Sword"
	w := gurps.NewWeapon(owner, true)
	w.Damage.StrengthType = stdmg.None // keep the resolved damage to just the base and the bonus
	w.Damage.Base = "1d"
	w.Defaults = []*gurps.SkillDefault{def}
	owner.Weapons = []*gurps.Weapon{w}
	e.Traits = append(e.Traits, owner)

	e.Recalculate()
	c.Equal(fxp.Eleven, sk.LevelData.Level, "the Broadsword skill should be at level 11")
	return w
}

// TestWeaponTagDefaultCollectsSkillBonuses verifies that a weapon whose only skill default selects its skill by tag
// still collects the weapon bonuses that are qualified by the name of the skill it attacks with. Such a default names
// no skill outright, so the qualifier text it carries is empty and the bonus would be missed unless the default is
// resolved to the skill it actually scored. See issue #1112.
func TestWeaponTagDefaultCollectsSkillBonuses(t *testing.T) {
	c := check.New(t)

	tagged := newTagDefaultTestWeapon(c, &gurps.SkillDefault{
		DefaultType: gurps.SkillID,
		Tags:        criteria.Text{Compare: criteria.IsText, Qualifier: "Combat"},
	})
	// The control: the same weapon, with a default that names the skill outright.
	named := newTagDefaultTestWeapon(c, &gurps.SkillDefault{
		DefaultType: gurps.SkillID,
		Name:        criteria.Text{Compare: criteria.IsText, Qualifier: "Broadsword"},
	})

	c.Equal(fxp.Eleven, tagged.SkillLevel(nil), "a tag default must attack with the skill it selects")
	c.Equal(named.SkillLevel(nil), tagged.SkillLevel(nil),
		"a tag default must attack at the same level as a default naming the same skill")
	c.Equal("1d+2 cr", tagged.Damage.ResolvedDamage(nil),
		"a tag default must collect the +2 granted to weapons whose required skill is Broadsword")
	c.Equal(named.Damage.ResolvedDamage(nil), tagged.Damage.ResolvedDamage(nil),
		"a tag default must collect the same skill-qualified bonuses as a default naming the same skill")
}

// TestWeaponBonusWithRequiredSkillAndNoSkill verifies how a weapon bonus selecting weapons by required skill treats a
// weapon for which no skill-based default resolves: the unfilled name criterion a new bonus starts with selects
// nothing, while a name criterion of anything selects it.
func TestWeaponBonusWithRequiredSkillAndNoSkill(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	master := gurps.NewTrait(e, nil, false)
	master.Name = "Weapon Master"
	bonus := gurps.NewWeaponBonus(feature.WeaponBonus) // Selects by required skill, with the name left blank
	master.Features = gurps.Features{bonus}
	e.Traits = append(e.Traits, master)
	owner := gurps.NewTrait(e, nil, false)
	owner.Name = "Natural Attacks"
	byAttribute := gurps.NewWeapon(owner, true)
	byAttribute.Usage = "Punch"
	byAttribute.Damage.StrengthType = stdmg.None
	byAttribute.Damage.Base = "1d"
	byAttribute.Defaults = []*gurps.SkillDefault{{DefaultType: gurps.DexterityID}}
	byAbsentSkill := gurps.NewWeapon(owner, true)
	byAbsentSkill.Usage = "Kick"
	byAbsentSkill.Damage.StrengthType = stdmg.None
	byAbsentSkill.Damage.Base = "1d"
	byAbsentSkill.Defaults = []*gurps.SkillDefault{{
		DefaultType: gurps.SkillID,
		Name:        criteria.Text{Compare: criteria.IsText, Qualifier: "Brawling"},
	}}
	owner.Weapons = []*gurps.Weapon{byAttribute, byAbsentSkill}
	e.Traits = append(e.Traits, owner)
	e.Recalculate()
	c.Equal("1d cr", byAttribute.Damage.ResolvedDamage(nil),
		"a bonus whose skill name is left blank does not apply to a weapon whose defaults are all attribute-based")
	c.Equal("1d cr", byAbsentSkill.Damage.ResolvedDamage(nil),
		"nor to one whose skill default names a skill the character lacks")

	bonus.NameCriteria = criteria.Text{Compare: criteria.AnyText}
	e.DiscardCaches()
	c.Equal("1d+1 cr", byAttribute.Damage.ResolvedDamage(nil), "a skill name of anything applies to both")
	c.Equal("1d+1 cr", byAbsentSkill.Damage.ResolvedDamage(nil), "a skill name of anything applies to both")
}

// TestWeaponDefaultsToSpell verifies that an attack can rest on a spell: at the level of the best spell its default
// selects plus the modifier, for its parry as for its attack, and with that spell as the skill it is used with, by name
// and relative level, for the bonuses selecting weapons by it. The attack belongs to the Arcane Fireball, so a default
// whose power source is to be the same as this spell's selects that one, not the better Divine one.
func TestWeaponDefaultsToSpell(t *testing.T) {
	same := &gurps.SkillDefault{
		DefaultType:     gurps.SpellID,
		Name:            criteria.Text{Compare: criteria.IsText, Qualifier: "Fireball"},
		SamePowerSource: true,
		Modifier:        -fxp.One,
	}
	anySource := *same
	anySource.SamePowerSource = false
	for _, tc := range []struct {
		name   string
		def    *gurps.SkillDefault
		level  fxp.Int
		parry  string
		damage string
	}{
		{name: "any power source", def: &anySource, level: fxp.Thirteen, parry: "9", damage: "1d+2 burn"},
		{name: "the same power source as this spell", def: same, level: fxp.Eleven, parry: "8", damage: "1d+2 burn"},
	} {
		c := check.New(t)
		e := gurps.NewEntity()
		var arcane *gurps.Spell
		for _, one := range []struct {
			powerSource string
			points      fxp.Int
		}{{"Arcane", fxp.Twelve}, {"Divine", fxp.FromInteger(20)}} { // 12 and 14, IQ+2 and IQ+4
			sp := gurps.NewSpell(e, nil, false)
			sp.Name = "Fireball"
			sp.PowerSource = one.powerSource
			sp.Points = one.points
			e.Spells = append(e.Spells, sp)
			if arcane == nil {
				arcane = sp
			}
		}
		master := gurps.NewTrait(e, nil, false)
		master.Name = "Fire Adept"
		bonus := gurps.NewWeaponBonus(feature.WeaponBonus)
		bonus.SelectionType = wsel.WithRequiredSkill
		bonus.NameCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Fireball"}
		bonus.RelativeLevelCriteria = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.Two}
		bonus.Amount = fxp.Two
		master.Features = gurps.Features{bonus}
		e.Traits = append(e.Traits, master)
		w := gurps.NewWeapon(arcane, true)
		w.Usage = "Flaming Touch"
		w.Damage.StrengthType = stdmg.None
		w.Damage.Base = "1d"
		w.Damage.Type = "burn"
		w.Parry = attack.Parry{CanParry: true}
		w.Defaults = []*gurps.SkillDefault{tc.def}
		arcane.Weapons = []*gurps.Weapon{w}
		e.Recalculate()
		c.Equal(fxp.Twelve, arcane.LevelData.Level, "%s: precondition: the Arcane Fireball is at 12", tc.name)
		c.Equal(tc.level, w.SkillLevel(nil), "%s: the attack's level", tc.name)
		c.Equal(tc.parry, w.ResolvedParry(nil).String(), "%s: the parry", tc.name)
		c.Equal(tc.damage, w.Damage.ResolvedDamage(nil), "%s: the bonus to weapons whose required skill is Fireball",
			tc.name)
	}
}

// newCrossbowTestWeapon builds an entity whose ST is 10 for striking and 12 for lifting, holding a Crossbow skill
// tagged "Missile Weapon" at level 12, and returns a muscle-powered ranged weapon with a minimum ST of 12 whose only
// default is the one given. The weapon is penalized 2 for its minimum ST unless it is recognized as a crossbow, which
// is drawn with lifting ST.
func newCrossbowTestWeapon(c check.Checker, def *gurps.SkillDefault) *gurps.Weapon {
	e := gurps.NewEntity()

	// DX+2, i.e. level 12 with the default DX of 10
	sk := addDXSkill(e, "Crossbow", difficulty.Easy, fxp.Four, "Missile Weapon")

	lifter := gurps.NewTrait(e, nil, false)
	lifter.Name = "Lifting ST 2"
	bonus := gurps.NewAttributeBonus(gurps.StrengthID)
	bonus.Amount = fxp.Two
	bonus.Limitation = stlimit.LiftingOnly
	lifter.Features = gurps.Features{bonus}
	e.Traits = append(e.Traits, lifter)

	owner := gurps.NewTrait(e, nil, false)
	owner.Name = "Crossbow"
	w := gurps.NewWeapon(owner, false)
	w.Range.MusclePowered = true
	w.Strength.Min = fxp.Twelve
	w.Defaults = []*gurps.SkillDefault{def}
	owner.Weapons = []*gurps.Weapon{w}
	e.Traits = append(e.Traits, owner)

	e.Recalculate()
	c.Equal(fxp.Twelve, sk.LevelData.Level, "the Crossbow skill should be at level 12")
	c.Equal(fxp.Ten, e.StrikingStrength(), "striking ST should be the unmodified 10")
	c.Equal(fxp.Twelve, e.LiftingStrength(), "lifting ST should carry the +2")
	return w
}

// TestWeaponTagDefaultToCrossbowUsesLiftingST verifies that a muscle-powered ranged weapon whose default reaches the
// Crossbow skill by tag is drawn with lifting ST, as one naming Crossbow outright is, rather than being penalized
// against striking ST for its minimum ST.
func TestWeaponTagDefaultToCrossbowUsesLiftingST(t *testing.T) {
	c := check.New(t)

	named := newCrossbowTestWeapon(c, &gurps.SkillDefault{
		DefaultType: gurps.SkillID,
		Name:        criteria.Text{Compare: criteria.IsText, Qualifier: "Crossbow"},
	})
	c.Equal(fxp.Twelve, named.SkillLevel(nil), "the control: lifting ST 12 meets the minimum ST, so there is no penalty")

	tagged := newCrossbowTestWeapon(c, &gurps.SkillDefault{
		DefaultType: gurps.SkillID,
		Tags:        criteria.Text{Compare: criteria.IsText, Qualifier: "Missile Weapon"},
	})
	c.Equal(named.SkillLevel(nil), tagged.SkillLevel(nil),
		"a default that reaches Crossbow by tag must be drawn with lifting ST as well")

	// A default names the skill however its "is" matches it, ignoring case and surrounding space, whether or not the
	// character has the skill, in which case the weapon falls to its DX default.
	for _, name := range []string{"Crossbow", "crossbow", " Crossbow "} {
		spelled := newCrossbowTestWeapon(c, &gurps.SkillDefault{
			DefaultType: gurps.SkillID,
			Name:        criteria.Text{Compare: criteria.IsText, Qualifier: name},
		})
		spelled.Defaults = append(spelled.Defaults, &gurps.SkillDefault{DefaultType: gurps.DexterityID, Modifier: -fxp.Four})
		e := spelled.Entity()
		e.Skills = nil
		e.Recalculate()
		c.Equal(fxp.Six, spelled.SkillLevel(nil), "%q names the Crossbow skill, so lifting ST meets the minimum", name)
	}

	// A default that reaches some other skill is still drawn with striking ST, and takes the penalty.
	bow := newCrossbowTestWeapon(c, &gurps.SkillDefault{
		DefaultType: gurps.SkillID,
		Tags:        criteria.Text{Compare: criteria.IsText, Qualifier: "Archery"},
	})
	addDXSkill(bow.Entity(), "Bow", difficulty.Easy, fxp.Four, "Archery")
	bow.Entity().Recalculate()
	c.Equal(fxp.Ten, bow.SkillLevel(nil), "a bow is drawn with striking ST 10, and is penalized 2 for its minimum ST")
}

// TestWeaponSkillLevelUsesResolvedMusclePower verifies that the ST a ranged weapon's minimum ST is checked against
// follows a weapon switch for muscle power, as the weapon's range does, rather than the weapon's own flag.
func TestWeaponSkillLevelUsesResolvedMusclePower(t *testing.T) {
	c := check.New(t)
	for _, musclePowered := range []bool{true, false} {
		bow := newCrossbowTestWeapon(c, &gurps.SkillDefault{
			DefaultType: gurps.SkillID,
			Tags:        criteria.Text{Compare: criteria.IsText, Qualifier: "Archery"},
		})
		e := bow.Entity()
		addDXSkill(e, "Bow", difficulty.Easy, fxp.Four, "Archery")
		bow.Range.MusclePowered = !musclePowered
		sw := gurps.NewWeaponBonus(feature.WeaponSwitch)
		sw.SelectionType = wsel.ThisWeapon
		sw.SwitchType = wswitch.MusclePowered
		sw.SwitchTypeValue = musclePowered
		owner, ok := bow.Owner.(*gurps.Trait)
		c.True(ok, "the test weapon should be owned by a trait")
		owner.Features = append(owner.Features, sw)
		e.Recalculate()
		c.Equal(musclePowered, bow.ResolvedRange(nil).MusclePowered, "the switch decides the range's muscle power")
		if musclePowered {
			c.Equal(fxp.Ten, bow.SkillLevel(nil),
				"a bow switched to muscle power is drawn with striking ST 10, and is penalized 2 for its minimum ST")
		} else {
			c.Equal(fxp.Twelve, bow.SkillLevel(nil),
				"a bow switched off muscle power is checked against lifting ST 12, which meets its minimum ST")
		}
	}
}

// TestWeaponSkillLevelUsesResolvedStrengthBasis verifies that the minimum-ST skill penalty is computed against the
// attribute the weapon's *resolved* damage strength basis names, so a WeaponDamageStrengthBasis override is honored.
func TestWeaponSkillLevelUsesResolvedStrengthBasis(t *testing.T) {
	c := check.New(t)
	dxDefault := func() []*gurps.SkillDefault {
		return []*gurps.SkillDefault{{DefaultType: gurps.DexterityID}}
	}

	// Without an override, the thrust basis measures the weapon's ST 14 requirement against striking ST 14, so there is
	// no penalty and the weapon sits at its DX default of 10.
	w := newStrengthBasisWeapon(true, 5, "")
	w.Strength = attack.ParseStrength("14")
	w.Defaults = dxDefault()
	c.Equal(fxp.Ten, w.SkillLevel(nil), "striking ST 14 meets the weapon's minimum ST of 14")

	// Retargeting the strength basis to telekinetic must measure that same requirement against TK 5, a 9-point shortfall.
	w = newStrengthBasisWeapon(true, 5, stdmg.TelekineticThrust.Key())
	w.Strength = attack.ParseStrength("14")
	w.Defaults = dxDefault()
	c.Equal(fxp.One, w.SkillLevel(nil), "TK 5 against a minimum ST of 14 costs 9 levels")

	var tooltip xbytes.InsertBuffer
	w.SkillLevel(&tooltip)
	c.Contains(tooltip.String(), "-9", "the tooltip reports the penalty actually applied")
	c.Contains(tooltip.String(), "minimum ST requirement", "the penalty is attributed to the minimum ST requirement")
}

// TestWeaponSkillLevelFloorsAtZero verifies that a weapon whose best default resolves below 0 has a skill level of 0.
func TestWeaponSkillLevelFloorsAtZero(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	owner := gurps.NewTrait(e, nil, false)
	owner.Name = "Sword"
	w := gurps.NewWeapon(owner, true)
	w.Defaults = []*gurps.SkillDefault{{DefaultType: gurps.DexterityID, Modifier: -fxp.Twelve}} // DX 10 - 12
	owner.Weapons = []*gurps.Weapon{w}
	e.Traits = append(e.Traits, owner)
	e.Recalculate()
	c.Equal(fxp.Int(0), w.SkillLevel(nil))
}

// TestWeaponBonusesFollowATechniqueToItsBaseSkill verifies that a weapon whose best default is a technique also
// collects the weapon bonuses qualified by the skill the technique is based on, judged by that skill's relative level.
func TestWeaponBonusesFollowATechniqueToItsBaseSkill(t *testing.T) {
	c := check.New(t)
	e, err := gurps.NewEntityFromFile(os.DirFS("testdata"), "issue767.gcs")
	c.NoError(err)
	var kick *gurps.Weapon
	gurps.Traverse(func(tr *gurps.Trait) bool {
		for _, w := range tr.Weapons {
			if w.UsageWithReplacements() == "Kick" {
				kick = w
			}
		}
		return false
	}, false, false, e.Traits...)
	c.NotNil(kick)
	// Brawling is at DX+2 (RSL +2), so its "+1 per die, RSL >= 2" weapon damage bonus must apply even though the
	// best skill default for the Kick is the Kicking technique (RSL -1 vs Brawling). 1d-2 base + 1 = 1d-1.
	c.Equal("1d-1 cr", kick.Damage.ResolvedDamage(nil))
}

// TestWeaponNotesFromOwnerModifiers verifies that a weapon's notes gather the notes of every enabled, non-container
// modifier of its owner that asks to be shown on the weapon, at any depth, ahead of the weapon's own usage notes,
// whether the owner is a trait or a piece of equipment, and whether the weapon is the owner's own or one of its
// modifiers'.
func TestWeaponNotesFromOwnerModifiers(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()

	trait := gurps.NewTrait(e, nil, false)
	trait.Name = "Claws"
	shown := gurps.NewTraitModifier(e, nil, false)
	shown.Name = "Sharp"
	shown.LocalNotes = " Cutting damage "
	shown.ShowNotesOnWeapon = true
	hidden := gurps.NewTraitModifier(e, nil, false)
	hidden.Name = "Long"
	hidden.LocalNotes = "Reach 2"
	disabled := gurps.NewTraitModifier(e, nil, false)
	disabled.Name = "Poisoned"
	disabled.LocalNotes = "Follow-up toxic"
	disabled.ShowNotesOnWeapon = true
	disabled.SetDisabled(true)
	nested := gurps.NewTraitModifier(e, nil, false)
	nested.Name = "Retractable"
	nested.LocalNotes = "Takes a Ready"
	nested.ShowNotesOnWeapon = true
	container := gurps.NewTraitModifier(e, nil, true)
	container.Name = "Options"
	container.LocalNotes = "Never shown: a container's own notes don't apply"
	container.ShowNotesOnWeapon = true
	container.Children = []*gurps.TraitModifier{nested}
	trait.Modifiers = []*gurps.TraitModifier{shown, hidden, disabled, container}
	claws := gurps.NewWeapon(trait, true)
	claws.UsageNotes = "Usage"
	trait.Weapons = []*gurps.Weapon{claws}
	e.Traits = append(e.Traits, trait)
	c.Equal("Cutting damage\nTakes a Ready\nUsage", claws.Notes())

	// A modifier's own weapon belongs to the trait as well, so it gathers the same notes, the holder's own included.
	slash := gurps.NewWeapon(nil, true)
	slash.UsageNotes = "Slash"
	shown.Weapons = []*gurps.Weapon{slash}
	shown.SetTarget(trait)
	c.Equal("Cutting damage\nTakes a Ready\nSlash", slash.Notes())

	eqp := gurps.NewEquipment(e, nil, false)
	eqp.Name = "Sword"
	fine := gurps.NewEquipmentModifier(e, nil, false)
	fine.Name = "Fine"
	fine.LocalNotes = "+1 damage"
	fine.ShowNotesOnWeapon = true
	silvered := gurps.NewEquipmentModifier(e, nil, false)
	silvered.Name = "Silvered"
	silvered.LocalNotes = "Hurts werewolves"
	off := gurps.NewEquipmentModifier(e, nil, false)
	off.Name = "Poisoned"
	off.LocalNotes = "Follow-up toxic"
	off.ShowNotesOnWeapon = true
	off.SetDisabled(true)
	eqp.Modifiers = []*gurps.EquipmentModifier{fine, silvered, off}
	sword := gurps.NewWeapon(eqp, true)
	eqp.Weapons = []*gurps.Weapon{sword}
	e.CarriedEquipment = append(e.CarriedEquipment, eqp)
	c.Equal("+1 damage", sword.Notes(), "a weapon without usage notes has just the modifier notes")

	orphan := gurps.NewWeapon(nil, true)
	orphan.UsageNotes = "Usage"
	c.Equal("Usage", orphan.Notes(), "a weapon with no owner has just its usage notes")
}

// TestWeaponEffectiveStrength verifies the ST a weapon is used at, which its damage and its muscle-powered range
// share: the entity's ST by default, the owner's rated ST instead when it has one, raised by any effective ST bonus
// and capped at three times the weapon's minimum ST.
func TestWeaponEffectiveStrength(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity() // ST 10
	eqp := gurps.NewEquipment(e, nil, false)
	eqp.Name = "Sling"
	w := gurps.NewWeapon(eqp, false)
	w.Damage.Base = "" // Just the ST-based damage, so the ST used is plain to see
	w.Damage.StrengthType = stdmg.Thrust
	w.Range = attack.ParseRange("x1/x2")
	eqp.Weapons = []*gurps.Weapon{w}
	e.CarriedEquipment = append(e.CarriedEquipment, eqp)
	e.Recalculate()
	c.True(w.Range.MusclePowered)

	expect := func(expectedDamage, expectedRange, msg string) {
		c.Helper()
		c.Equal(expectedDamage, w.Damage.ResolvedDamage(nil), msg)
		c.Equal(expectedRange, w.ResolvedRange(nil).String(true), msg)
	}
	expect("1d-2 cr", "10/20", "the entity's ST 10 is used when there is no rated ST")

	eqp.RatedST = fxp.FromInteger(14)
	e.DiscardCaches()
	expect("1d cr", "14/28", "the owner's rated ST replaces the entity's")

	bonus := gurps.NewWeaponBonus(feature.WeaponEffectiveSTBonus)
	bonus.SelectionType = wsel.ThisWeapon
	bonus.Amount = fxp.Two
	eqp.Features = gurps.Features{bonus}
	e.DiscardCaches()
	expect("1d+1 cr", "16/32", "an effective ST bonus raises the rated ST")

	w.Strength = attack.ParseStrength("5")
	expect("1d+1 cr", "15/30", "the effective ST is capped at three times the weapon's minimum ST")

	var tooltip xbytes.InsertBuffer
	w.ResolvedRange(&tooltip)
	c.Contains(tooltip.String(), "+2 to effective ST", "range resolution explains the effective ST bonus")
}

// TestCloneWeaponsBelongToTheirNewHolder verifies that the weapon copies made when a trait, skill, spell or piece of
// equipment is cloned -- or synced from its library source -- belong to the clone, those of its modifiers included,
// which are held by the clone's modifier. Weapon.Clone() keeps the original's owner and holding modifier, so every
// caller must supply the new ones; a copy left pointing at the original resolves its nameable replacements (and its
// skill defaults and minimum ST) through it, so a duplicated row's weapons followed the original.
func TestCloneWeaponsBelongToTheirNewHolder(t *testing.T) {
	c := check.New(t)

	// Each case builds a node named "@kind@ Attack" holding one weapon, clones it, then diverges the clone's
	// replacements. The clone's weapon must read the clone's map.
	for _, one := range []struct {
		name  string
		build func() (source, clone any, sourceWeapon, cloneWeapon *gurps.Weapon)
		// holder returns the modifier of the node that holds its weapon, or is nil when the node holds it itself.
		holder func(node any) gurps.WeaponHoldingModifier
	}{
		{
			name: "trait",
			build: func() (any, any, *gurps.Weapon, *gurps.Weapon) {
				n := gurps.NewTrait(nil, nil, false)
				n.Name = "@kind@ Attack"
				return cloneWithWeapon(n, true, func(t *gurps.Trait) *[]*gurps.Weapon { return &t.Weapons })
			},
		},
		{
			name: "skill",
			build: func() (any, any, *gurps.Weapon, *gurps.Weapon) {
				n := gurps.NewSkill(nil, nil, false)
				n.Name = "@kind@ Attack"
				return cloneWithWeapon(n, true, func(s *gurps.Skill) *[]*gurps.Weapon { return &s.Weapons })
			},
		},
		{
			name: "spell",
			build: func() (any, any, *gurps.Weapon, *gurps.Weapon) {
				n := gurps.NewSpell(nil, nil, false)
				n.Name = "@kind@ Attack"
				return cloneWithWeapon(n, false, func(s *gurps.Spell) *[]*gurps.Weapon { return &s.Weapons })
			},
		},
		{
			name: "equipment",
			build: func() (any, any, *gurps.Weapon, *gurps.Weapon) {
				n := gurps.NewEquipment(nil, nil, false)
				n.Name = "@kind@ Attack"
				return cloneWithWeapon(n, true, func(e *gurps.Equipment) *[]*gurps.Weapon { return &e.Weapons })
			},
		},
		{
			name: "trait modifier",
			build: func() (any, any, *gurps.Weapon, *gurps.Weapon) {
				n := gurps.NewTrait(nil, nil, false)
				n.Name = "@kind@ Attack"
				n.AddModifiers(gurps.NewTraitModifier(nil, nil, false))
				return cloneWithWeapon(n, true, func(t *gurps.Trait) *[]*gurps.Weapon { return &t.Modifiers[0].Weapons })
			},
			holder: func(node any) gurps.WeaponHoldingModifier {
				if n, ok := node.(*gurps.Trait); ok {
					return n.Modifiers[0]
				}
				return nil
			},
		},
		{
			name: "equipment modifier",
			build: func() (any, any, *gurps.Weapon, *gurps.Weapon) {
				n := gurps.NewEquipment(nil, nil, false)
				n.Name = "@kind@ Attack"
				n.AddModifiers(gurps.NewEquipmentModifier(nil, nil, false))
				return cloneWithWeapon(n, true, func(e *gurps.Equipment) *[]*gurps.Weapon { return &e.Modifiers[0].Weapons })
			},
			holder: func(node any) gurps.WeaponHoldingModifier {
				if n, ok := node.(*gurps.Equipment); ok {
					return n.Modifiers[0]
				}
				return nil
			},
		},
	} {
		source, clone, sourceWeapon, cloneWeapon := one.build()
		c.True(cloneWeapon.Owner == clone, "%s: the cloned weapon belongs to the clone", one.name)
		c.True(sourceWeapon.Owner == source, "%s: the original's weapon still belongs to the original", one.name)
		var sourceHolder, cloneHolder gurps.WeaponHoldingModifier
		if one.holder != nil {
			sourceHolder, cloneHolder = one.holder(source), one.holder(clone)
		}
		c.True(cloneWeapon.HoldingModifier() == cloneHolder,
			"%s: the cloned weapon is held by the clone's modifier, if any", one.name)
		c.True(sourceWeapon.HoldingModifier() == sourceHolder, "%s: the original's weapon keeps its holder", one.name)
		c.Equal(map[string]string{"kind": "Ice"}, cloneWeapon.NameableReplacements(),
			"%s: the cloned weapon resolves against the clone's replacements", one.name)
		c.Equal(map[string]string{"kind": "Fire"}, sourceWeapon.NameableReplacements(),
			"%s: the original's weapon is unaffected", one.name)
		c.True(cloneWeapon.Damage.Owner == cloneWeapon, "%s: the cloned weapon's damage belongs to it", one.name)
	}
}

// cloneWithWeapon gives n a "Fire" replacement for "@kind@" and one weapon, then clones n and diverges the clone's
// replacement to "Ice", so the two weapons must resolve differently. The Weapons slice is a plain field on every node
// type rather than an interface method, so it needs the accessor, which may reach into the node's modifier.
func cloneWithWeapon[T interface {
	gurps.Node[T]
	gurps.WeaponOwner
	nameable.Setter
}](n T, melee bool, weapons func(T) *[]*gurps.Weapon) (source, clone any, sourceWeapon, cloneWeapon *gurps.Weapon) {
	n.SetNameableReplacements(map[string]string{"kind": "Fire"})
	*weapons(n) = []*gurps.Weapon{gurps.NewWeapon(n, melee)}
	n.SetDataOwner(nil)
	dup := n.Clone(gurps.LibraryFile{}, nil, nil, gurps.Reference)
	dup.NameableReplacements()["kind"] = "Ice"
	return n, dup, (*weapons(n))[0], (*weapons(dup))[0]
}
