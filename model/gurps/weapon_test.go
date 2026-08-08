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
	"encoding/json/v2"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestWeaponParryAndBlockStorage(t *testing.T) {
	c := check.New(t)

	w := gurps.NewWeapon(nil, true)
	data, err := json.Marshal(w)
	c.NoError(err)
	s := string(data)
	c.NotContains(s, "parry")
	c.NotContains(s, "block")

	var loadedWeapon gurps.Weapon
	c.NoError(json.Unmarshal(data, &loadedWeapon))
	c.Equal(gurps.WeaponParry{}, loadedWeapon.Parry)
	c.Equal(gurps.WeaponBlock{}, loadedWeapon.Block)

	w.Parry.CanParry = true
	w.Block.CanBlock = true
	data, err = json.Marshal(&w)
	c.NoError(err)
	s = string(data)
	c.Contains(s, "parry")
	c.Contains(s, "block")

	w = gurps.NewWeapon(nil, false)
	data, err = json.Marshal(w)
	c.NoError(err)
	s = string(data)
	c.NotContains(s, "parry")
	c.NotContains(s, "block")

	loadedWeapon = gurps.Weapon{}
	c.NoError(json.Unmarshal(data, &loadedWeapon))
	c.Equal(gurps.WeaponParry{}, loadedWeapon.Parry)
	c.Equal(gurps.WeaponBlock{}, loadedWeapon.Block)
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
	ranged.RateOfFire = gurps.WeaponRoF{}
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

	// Melee-specific columns behave the same way.
	c.False(melee.ColumnHasData(gurps.WeaponParryColumn))
	c.False(melee.ColumnHasData(gurps.WeaponBlockColumn))
	melee.Parry.CanParry = true
	melee.Block.CanBlock = true
	c.True(melee.ColumnHasData(gurps.WeaponParryColumn))
	c.True(melee.ColumnHasData(gurps.WeaponBlockColumn))
}

// TestWeaponDamageWithNilOwner verifies that formatting and marshaling a weapon whose damage has no back-reference to
// its owning weapon does not panic. This can happen transiently on the table undo/redo deserialize path, where freshly
// deserialized weapons have not yet had their owners wired up. A non-dice (script) damage base is required to force
// resolution through the scripting path, which previously dereferenced the nil owner and panicked. See issue #1015.
func TestWeaponDamageWithNilOwner(t *testing.T) {
	c := check.New(t)

	w := gurps.NewWeapon(nil, true)
	w.Damage.Base = "$self.skillLevel" // A non-dice base forces resolution through the scripting path.
	w.Damage.Owner = nil               // Explicit: no back-reference to the owning weapon.

	c.NotPanics(func() {
		_ = w.Damage.String()
	})
	c.NotPanics(func() {
		_, err := json.Marshal(w)
		c.NoError(err)
	})
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

	sk := gurps.NewSkill(e, nil, false)
	sk.Name = "Cloak"
	sk.Difficulty.Attribute = gurps.DexterityID
	sk.Difficulty.Difficulty = difficulty.Average
	sk.Points = fxp.FromInteger(16) // DX+4, i.e. level 14 with the default DX of 10
	e.Skills = append(e.Skills, sk)

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
	w.Parry = gurps.WeaponParry{CanParry: true}
	w.Block = gurps.WeaponBlock{CanBlock: true}
	owner.Weapons = []*gurps.Weapon{w}
	e.Traits = append(e.Traits, owner)

	e.Recalculate()
	c.Equal(fxp.FromInteger(14), sk.LevelData.Level, "the Cloak skill should be at level 14")
	c.Equal(fxp.FromInteger(parryBonus), e.ParryBonus, "the entity should have a +%d parry bonus", parryBonus)
	c.Equal(fxp.FromInteger(blockBonus), e.BlockBonus, "the entity should have a +%d block bonus", blockBonus)
	return w
}

// TestWeaponSkillLevelIgnoresDefenseTypeDefaults verifies that a parry- or block-type weapon default does not become
// the weapon's attack skill. The one Defaults list feeds the attack, parry, and block calculations alike, and a
// defense-type default is there for the defenses: the Tonfa in the High Tech library carries "Brawling Parry" and
// "Karate Parry" alongside its real attack defaults. Scoring such a default by the level of the skill it names would
// let a high Karate win the attack, so it stays scored by its halved defense level, which keeps it out of the way.
func TestWeaponSkillLevelIgnoresDefenseTypeDefaults(t *testing.T) {
	c := check.New(t)
	w := newDefenseTestWeapon(c)
	e := w.Entity()

	// Stand in for the Tonfa: a weaker skill the weapon really attacks with, alongside the strong Cloak-14 that the
	// weapon only names for its parry.
	tonfa := gurps.NewSkill(e, nil, false)
	tonfa.Name = "Tonfa"
	tonfa.Difficulty.Attribute = gurps.DexterityID
	tonfa.Difficulty.Difficulty = difficulty.Average
	tonfa.Points = fxp.FromInteger(12) // DX+3, i.e. level 13
	e.Skills = append(e.Skills, tonfa)
	e.Recalculate()

	tonfaDefault := newDefenseTestDefault(gurps.SkillID)
	tonfaDefault.Name.Qualifier = "Tonfa"
	w.Defaults = []*gurps.SkillDefault{tonfaDefault, newDefenseTestDefault(gurps.ParryID)}

	// The attack stays on Tonfa-13. Cloak's parry level of 11 loses the max, as intended; Cloak-14 would have won it.
	c.Equal(fxp.FromInteger(13), w.SkillLevel(nil), "a parry-type default must not take over the attack skill")

	// The parry, meanwhile, is the better of Tonfa's (13/2 + 3 + 1) and Cloak's (14/2 + 3 + 1).
	c.Equal("11", w.Parry.Resolve(w, nil).String(), "the parry-type default still feeds the parry")
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

// TestNewWeaponIsCurrentSubVersion verifies that a freshly created weapon is stamped with the current data sub-version,
// so that the data fix-ups applied to pre-sub-versioning data are not run against it. Without the stamp, the first
// SetOwner call on a weapon owned by a leveled trait or piece of equipment moved the base damage into the per-level
// slot, turning a new weapon's "1d" into "1d per level".
func TestNewWeaponIsCurrentSubVersion(t *testing.T) {
	c := check.New(t)

	e := gurps.NewEntity()
	trait := gurps.NewTrait(e, nil, false)
	trait.Name = "Innate Attack"
	trait.CanLevel = true
	trait.Levels = fxp.Three
	c.True(trait.IsLeveled())

	w := gurps.NewWeapon(trait, false)
	c.Equal("1d", w.Damage.Base)
	trait.Weapons = []*gurps.Weapon{w}
	e.Traits = append(e.Traits, trait)
	e.Recalculate()

	c.Equal("1d", w.Damage.Base, "a new weapon's base damage must not be migrated into the per-level slot")
	c.Equal("", w.Damage.BaseLeveled, "a new weapon must not acquire per-level base damage")
}

// TestPreSubVersionWeaponDamageIsMigrated verifies that weapon data written before sub-versioning existed still has its
// base damage moved into the per-level slot when its owner is leveled. This is the migration that
// TestNewWeaponIsCurrentSubVersion must not trigger for new weapons.
func TestPreSubVersionWeaponDamageIsMigrated(t *testing.T) {
	c := check.New(t)

	e := gurps.NewEntity()
	trait := gurps.NewTrait(e, nil, false)
	trait.Name = "Innate Attack"
	trait.CanLevel = true
	trait.Levels = fxp.Three

	w := gurps.NewWeapon(trait, false)
	w.SubVersion = 0 // Simulate data saved prior to sub-versioning.
	trait.Weapons = []*gurps.Weapon{w}
	e.Traits = append(e.Traits, trait)
	e.Recalculate()

	c.Equal("", w.Damage.Base)
	c.Equal("1d", w.Damage.BaseLeveled, "pre-sub-versioning damage on a leveled owner must migrate to per-level")
}
