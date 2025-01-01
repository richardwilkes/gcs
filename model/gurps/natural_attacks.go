// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/i18n"
)

// NewNaturalAttacks creates a new "Natural Attacks" trait.
func NewNaturalAttacks(entity *Entity, parent *Trait) *Trait {
	a := NewTrait(entity, parent, false)
	a.Name = i18n.Text("Natural Attacks")
	a.PageRef = "B271"
	a.Weapons = []*Weapon{newBite(a), newPunch(a), newKick(a)}
	return a
}

func newBite(owner WeaponOwner) *Weapon {
	bite := NewWeapon(owner, true)
	bite.Usage = i18n.Text("Bite")
	bite.Reach.CloseCombat = true
	bite.Reach.Min = 0
	bite.Reach.Max = 0
	bite.Defaults = []*SkillDefault{
		{
			DefaultType: DexterityID,
		},
		{
			DefaultType: SkillID,
			Name:        "Brawling",
		},
	}
	bite.Damage.Type = "cr"
	bite.Damage.StrengthType = stdmg.Thrust
	bite.Damage.Base = &dice.Dice{
		Sides:      6,
		Modifier:   -1,
		Multiplier: 1,
	}
	bite.Damage.ArmorDivisor = fxp.One
	bite.Damage.FragmentationArmorDivisor = fxp.One
	bite.Damage.Owner = bite
	return bite
}

func newPunch(owner WeaponOwner) *Weapon {
	punch := NewWeapon(owner, true)
	punch.Usage = i18n.Text("Punch")
	punch.Parry.CanParry = true
	punch.Reach.CloseCombat = true
	punch.Reach.Min = 0
	punch.Reach.Max = 0
	punch.Defaults = []*SkillDefault{
		{
			DefaultType: DexterityID,
		},
		{
			DefaultType: SkillID,
			Name:        "Boxing",
		},
		{
			DefaultType: SkillID,
			Name:        "Brawling",
		},
		{
			DefaultType: SkillID,
			Name:        "Karate",
		},
	}
	punch.Damage.Type = "cr"
	punch.Damage.StrengthType = stdmg.Thrust
	punch.Damage.Base = &dice.Dice{
		Sides:      6,
		Modifier:   -1,
		Multiplier: 1,
	}
	punch.Damage.ArmorDivisor = fxp.One
	punch.Damage.FragmentationArmorDivisor = fxp.One
	punch.Damage.Owner = punch
	return punch
}

func newKick(owner WeaponOwner) *Weapon {
	kick := NewWeapon(owner, true)
	kick.Usage = i18n.Text("Kick")
	kick.Reach.CloseCombat = true
	kick.Reach.Min = fxp.One
	kick.Reach.Max = fxp.One
	kick.Defaults = []*SkillDefault{
		{
			DefaultType: DexterityID,
			Modifier:    -fxp.Two,
		},
		{
			DefaultType: SkillID,
			Name:        "Brawling",
			Modifier:    -fxp.Two,
		},
		{
			DefaultType: SkillID,
			Name:        "Kicking",
		},
		{
			DefaultType: SkillID,
			Name:        "Karate",
			Modifier:    -fxp.Two,
		},
	}
	kick.Damage.Type = "cr"
	kick.Damage.StrengthType = stdmg.Thrust
	kick.Damage.ArmorDivisor = fxp.One
	kick.Damage.FragmentationArmorDivisor = fxp.One
	kick.Damage.Owner = kick
	return kick
}
