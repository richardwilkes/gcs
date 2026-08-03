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
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestWeaponParryAndBlockStorage(t *testing.T) {
	c := check.New(t)
	const (
		parryKey = "parry"
		blockKey = "block"
	)

	w := gurps.NewWeapon(nil, true)
	data, err := json.Marshal(w)
	c.NoError(err)
	s := string(data)
	c.NotContains(s, parryKey)
	c.NotContains(s, blockKey)

	var loadedWeapon gurps.Weapon
	c.NoError(json.Unmarshal(data, &loadedWeapon))
	c.Equal(gurps.WeaponParry{}, loadedWeapon.Parry)
	c.Equal(gurps.WeaponBlock{}, loadedWeapon.Block)

	w.Parry.CanParry = true
	w.Block.CanBlock = true
	data, err = json.Marshal(&w)
	c.NoError(err)
	s = string(data)
	c.Contains(s, parryKey)
	c.Contains(s, blockKey)

	w = gurps.NewWeapon(nil, false)
	data, err = json.Marshal(w)
	c.NoError(err)
	s = string(data)
	c.NotContains(s, parryKey)
	c.NotContains(s, blockKey)

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

// defenseTestSkillName is the name of the skill the defense test weapon and its defaults are built around.
const defenseTestSkillName = "Cloak"

// newDefenseTestWeapon builds an entity holding a "Cloak" skill at level 14 and a +1 bonus to both parry and block,
// along with a trait carrying a melee weapon that can both parry and block. The weapon has no defaults; the caller
// supplies whichever ones the test needs.
func newDefenseTestWeapon(c check.Checker) *gurps.Weapon {
	e := gurps.NewEntity()

	sk := gurps.NewSkill(e, nil, false)
	sk.Name = defenseTestSkillName
	sk.Difficulty.Attribute = gurps.DexterityID
	sk.Difficulty.Difficulty = difficulty.Average
	sk.Points = fxp.FromInteger(16) // DX+4, i.e. level 14 with the default DX of 10
	e.Skills = append(e.Skills, sk)

	bonuses := gurps.NewTrait(e, nil, false)
	bonuses.Name = "Improved Defenses"
	for _, attrID := range []string{gurps.ParryID, gurps.BlockID} {
		bonus := gurps.NewAttributeBonus(attrID)
		bonus.Amount = fxp.One
		bonuses.Features = append(bonuses.Features, bonus)
	}
	e.Traits = append(e.Traits, bonuses)

	owner := gurps.NewTrait(e, nil, false)
	owner.Name = defenseTestSkillName
	w := gurps.NewWeapon(owner, true)
	w.Parry = gurps.WeaponParry{CanParry: true}
	w.Block = gurps.WeaponBlock{CanBlock: true}
	owner.Weapons = []*gurps.Weapon{w}
	e.Traits = append(e.Traits, owner)

	e.Recalculate()
	c.Equal(fxp.FromInteger(14), sk.LevelData.Level, "the Cloak skill should be at level 14")
	c.Equal(fxp.One, e.ParryBonus, "the entity should have a +1 parry bonus")
	c.Equal(fxp.One, e.BlockBonus, "the entity should have a +1 block bonus")
	return w
}

// newDefenseTestDefault creates a skill default of the given type that matches the "Cloak" skill created by
// newDefenseTestWeapon.
func newDefenseTestDefault(defaultType string) *gurps.SkillDefault {
	return &gurps.SkillDefault{
		DefaultType: defaultType,
		Name: criteria.Text{
			TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: defenseTestSkillName},
		},
	}
}
