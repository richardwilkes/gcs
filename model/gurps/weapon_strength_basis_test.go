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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// newStrengthBasisWeapon builds a ST 14 entity carrying a "Telekinesis" trait at the given level and a trait that owns
// a single thrust-based weapon. When overrideTo is non-empty, the owning trait also carries a WeaponDamageStrengthBasis
// override, matching any weapon, that retargets the weapon's strength basis to that value.
func newStrengthBasisWeapon(melee bool, tkLevels int, overrideTo string) *gurps.Weapon {
	e := gurps.NewEntity()
	e.Attributes.Set[gurps.StrengthID].Adjustment = fxp.Four // ST 14
	owner := gurps.NewTrait(e, nil, false)
	owner.Name = "Gadget"
	w := gurps.NewWeapon(owner, melee)
	w.Damage.StrengthType = stdmg.Thrust
	owner.Weapons = []*gurps.Weapon{w}
	if overrideTo != "" {
		override := gurps.NewSelectorOverride(selector.WeaponDamageStrengthBasis)
		override.Value = overrideTo
		override.NameCriteria.Compare = criteria.AnyText
		owner.Features = append(owner.Features, override)
	}
	e.Traits = append(e.Traits, owner)
	if tkLevels != 0 {
		tk := gurps.NewTrait(e, nil, false)
		tk.Name = "Telekinesis"
		tk.CanLevel = true
		tk.Levels = fxp.FromInteger(tkLevels)
		e.Traits = append(e.Traits, tk)
	}
	e.Recalculate()
	return w
}

// TestWeaponRangeUsesResolvedStrengthBasis verifies that a muscle-powered range is multiplied by the strength the
// weapon's *resolved* damage strength basis names. A WeaponDamageStrengthBasis override that retargets the weapon to
// telekinetic strength must move the range along with the damage; reading the raw basis here left the two disagreeing
// about the weapon's power source (damage from TK, range from throwing ST).
func TestWeaponRangeUsesResolvedStrengthBasis(t *testing.T) {
	c := check.New(t)

	// Without an override, the thrust basis falls through to throwing ST, which is the character's ST of 14.
	w := newStrengthBasisWeapon(false, 5, "")
	w.Range = gurps.ParseWeaponRange("x10/x15")
	c.Equal("140/210", w.Range.Resolve(w, nil).String(true), "the unoverridden range comes from throwing ST 14")

	// Retargeting the strength basis to telekinetic must drive the range from TK 5 instead.
	w = newStrengthBasisWeapon(false, 5, stdmg.TelekineticThrust.Key())
	w.Range = gurps.ParseWeaponRange("x10/x15")
	c.Equal("50/75", w.Range.Resolve(w, nil).String(true), "an override to telekinetic drives the range from TK 5")

	// The IQ basis is honored through the same path.
	w = newStrengthBasisWeapon(false, 5, stdmg.IQThrust.Key())
	w.Range = gurps.ParseWeaponRange("x10/x15")
	c.Equal("100/150", w.Range.Resolve(w, nil).String(true), "an override to IQ drives the range from IQ 10")
}

// TestWeaponSkillLevelUsesResolvedStrengthBasis verifies that the minimum-ST skill penalty is computed against the
// attribute the weapon's *resolved* damage strength basis names. Reading the raw basis measured the weapon's minimum ST
// against the wrong attribute whenever a WeaponDamageStrengthBasis override applied.
func TestWeaponSkillLevelUsesResolvedStrengthBasis(t *testing.T) {
	c := check.New(t)
	dxDefault := func() []*gurps.SkillDefault {
		return []*gurps.SkillDefault{{DefaultType: gurps.DexterityID}}
	}

	// Without an override, the thrust basis measures the weapon's ST 14 requirement against striking ST 14, so there is
	// no penalty and the weapon sits at its DX-10 default.
	w := newStrengthBasisWeapon(true, 5, "")
	w.Strength = gurps.ParseWeaponStrength("14")
	w.Defaults = dxDefault()
	c.Equal(fxp.Ten, w.SkillLevel(nil), "striking ST 14 meets the weapon's minimum ST of 14")

	// Retargeting the strength basis to telekinetic must measure that same requirement against TK 5, a 9-point shortfall.
	w = newStrengthBasisWeapon(true, 5, stdmg.TelekineticThrust.Key())
	w.Strength = gurps.ParseWeaponStrength("14")
	w.Defaults = dxDefault()
	c.Equal(fxp.One, w.SkillLevel(nil), "TK 5 against a minimum ST of 14 costs 9 levels")

	var tooltip xbytes.InsertBuffer
	w.SkillLevel(&tooltip)
	c.Contains(tooltip.String(), "-9", "the tooltip reports the penalty actually applied")
	c.Contains(tooltip.String(), "minimum ST requirement", "the penalty is attributed to the minimum ST requirement")
}
