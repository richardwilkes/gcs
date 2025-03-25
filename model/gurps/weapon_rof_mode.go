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
	"fmt"
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

// WeaponRoFMode holds the rate of fire data for one firing mode of a weapon.
type WeaponRoFMode struct {
	ShotsPerAttack             fxp.Int
	SecondaryProjectiles       fxp.Int
	FullAutoOnly               bool
	HighCyclicControlledBursts bool
}

// ParseWeaponRoFMode parses a string into a WeaponRoFMode.
func ParseWeaponRoFMode(s string) WeaponRoFMode {
	var wr WeaponRoFMode
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, ".", "x")
	wr.FullAutoOnly = strings.Contains(s, "!")
	s = strings.ReplaceAll(s, "!", "")
	wr.HighCyclicControlledBursts = strings.Contains(s, "#")
	s = strings.ReplaceAll(s, "#", "")
	s = strings.ReplaceAll(s, "×", "x")
	if strings.HasPrefix(s, "x") {
		s = "1" + s
	}
	parts := strings.Split(s, "x")
	wr.ShotsPerAttack, _ = fxp.Extract(s)
	if len(parts) > 1 {
		wr.SecondaryProjectiles, _ = fxp.Extract(parts[1])
	}
	wr.Validate()
	return wr
}

// Hash writes this object's contents into the hasher.
func (wr WeaponRoFMode) Hash(h hash.Hash) {
	hashhelper.Num64(h, wr.ShotsPerAttack)
	hashhelper.Num64(h, wr.SecondaryProjectiles)
	hashhelper.Bool(h, wr.FullAutoOnly)
	hashhelper.Bool(h, wr.HighCyclicControlledBursts)
}

// Resolve any bonuses that apply.
func (wr WeaponRoFMode) Resolve(w *Weapon, modifiersTooltip *xio.ByteBuffer, firstMode bool) WeaponRoFMode {
	result := wr
	var shotsFeature, secondaryFeature feature.Type
	if firstMode {
		shotsFeature = feature.WeaponRofMode1ShotsBonus
		secondaryFeature = feature.WeaponRofMode1SecondaryBonus
		result.FullAutoOnly = w.ResolveBoolFlag(wswitch.FullAuto1, wr.FullAutoOnly)
		result.HighCyclicControlledBursts = w.ResolveBoolFlag(wswitch.ControlledBursts1, wr.HighCyclicControlledBursts)
	} else {
		shotsFeature = feature.WeaponRofMode2ShotsBonus
		secondaryFeature = feature.WeaponRofMode2SecondaryBonus
		result.FullAutoOnly = w.ResolveBoolFlag(wswitch.FullAuto2, wr.FullAutoOnly)
		result.HighCyclicControlledBursts = w.ResolveBoolFlag(wswitch.ControlledBursts2, wr.HighCyclicControlledBursts)
	}
	var percentSPA, percentSP fxp.Int
	for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, shotsFeature, secondaryFeature) {
		amt := bonus.AdjustedAmountForWeapon(w)
		switch bonus.Type {
		case shotsFeature:
			if bonus.Percent {
				percentSPA += amt
			} else {
				result.ShotsPerAttack += amt
			}
		case secondaryFeature:
			if bonus.Percent {
				percentSP += amt
			} else {
				result.SecondaryProjectiles += amt
			}
		}
	}
	if percentSPA != 0 {
		result.ShotsPerAttack += result.ShotsPerAttack.Mul(percentSPA).Div(fxp.Hundred).Trunc()
	}
	if percentSP != 0 {
		result.SecondaryProjectiles += result.SecondaryProjectiles.Mul(percentSP).Div(fxp.Hundred).Trunc()
	}
	result.Validate()
	return result
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Call .Resolve() prior to calling this method if you want the resolved values.
func (wr WeaponRoFMode) String() string {
	if wr.ShotsPerAttack <= 0 {
		return ""
	}
	var buffer strings.Builder
	buffer.WriteString(wr.ShotsPerAttack.String())
	if wr.SecondaryProjectiles > 0 {
		buffer.WriteByte('x')
		buffer.WriteString(wr.SecondaryProjectiles.String())
	}
	if wr.FullAutoOnly {
		buffer.WriteByte('!')
	}
	if wr.HighCyclicControlledBursts {
		buffer.WriteByte('#')
	}
	return buffer.String()
}

// Tooltip returns a tooltip for the data, if any. Call .Resolve() prior to calling this method if you want the tooltip
// to be based on the resolved values.
func (wr WeaponRoFMode) Tooltip() string {
	if wr.ShotsPerAttack <= 0 || (wr.SecondaryProjectiles <= 0 && !wr.FullAutoOnly && !wr.HighCyclicControlledBursts) {
		return ""
	}
	var buffer strings.Builder
	if wr.SecondaryProjectiles > 0 {
		shotsText := i18n.Text("shots")
		if wr.ShotsPerAttack == fxp.One {
			shotsText = i18n.Text("shot")
		}
		projectilesText := i18n.Text("projectiles")
		if wr.SecondaryProjectiles == fxp.One {
			projectilesText = i18n.Text("projectile")
		}
		fmt.Fprintf(&buffer, i18n.Text("This weapon fires %v %s per attack and each shot releases %v smaller %s."), wr.ShotsPerAttack, shotsText, wr.SecondaryProjectiles, projectilesText)
	}
	if wr.FullAutoOnly {
		AppendStringOntoNewLine(&buffer, fmt.Sprintf(i18n.Text("!: This weapon can only fire on full automatic. Minimum RoF is %v."), wr.ShotsPerAttack.Div(fxp.Four).Ceil()))
	}
	if wr.HighCyclicControlledBursts {
		AppendStringOntoNewLine(&buffer, i18n.Text("#: This weapon can fire in high cyclic controlled bursts, reducing Recoil to 1."))
	}
	return buffer.String()
}

// Validate ensures that the data is valid.
func (wr *WeaponRoFMode) Validate() {
	wr.ShotsPerAttack = wr.ShotsPerAttack.Ceil().Max(0)
	if wr.ShotsPerAttack == 0 {
		wr.SecondaryProjectiles = 0
		wr.FullAutoOnly = false
		wr.HighCyclicControlledBursts = false
		return
	}
	wr.SecondaryProjectiles = wr.SecondaryProjectiles.Ceil().Max(0)
}
