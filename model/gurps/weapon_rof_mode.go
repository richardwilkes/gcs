/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps

import (
	"encoding/binary"
	"fmt"
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

// WeaponRoFMode holds the rate of fire data for one firing mode of a weapon.
type WeaponRoFMode struct {
	ShotsPerAttack             fxp.Int `json:"shots_per_attack,omitempty"`
	SecondaryProjectiles       fxp.Int `json:"secondary_projectiles,omitempty"`
	FullAutoOnly               bool    `json:"full_auto_only,omitempty"`
	HighCyclicControlledBursts bool    `json:"high_cyclic_controlled_bursts,omitempty"`
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

// nolint:errcheck // Not checking errors on writes to a bytes.Buffer
func (wr WeaponRoFMode) hash(h hash.Hash32) {
	_ = binary.Write(h, binary.LittleEndian, wr.ShotsPerAttack)
	_ = binary.Write(h, binary.LittleEndian, wr.SecondaryProjectiles)
	_ = binary.Write(h, binary.LittleEndian, wr.FullAutoOnly)
	_ = binary.Write(h, binary.LittleEndian, wr.HighCyclicControlledBursts)
}

// Resolve any bonuses that apply.
func (wr WeaponRoFMode) Resolve(w *Weapon, modifiersTooltip *xio.ByteBuffer, firstMode bool) WeaponRoFMode {
	if w.ResolveBoolFlag(JetWeaponSwitchType, w.Jet) {
		return WeaponRoFMode{}
	}
	result := wr
	var shotsFeature, secondaryFeature FeatureType
	if firstMode {
		shotsFeature = WeaponRofMode1ShotsBonusFeatureType
		secondaryFeature = WeaponRofMode1SecondaryBonusFeatureType
		result.FullAutoOnly = w.ResolveBoolFlag(FullAuto1WeaponSwitchType, wr.FullAutoOnly)
		result.HighCyclicControlledBursts = w.ResolveBoolFlag(ControlledBursts1WeaponSwitchType, wr.HighCyclicControlledBursts)
	} else {
		shotsFeature = WeaponRofMode2ShotsBonusFeatureType
		secondaryFeature = WeaponRofMode2SecondaryBonusFeatureType
		result.FullAutoOnly = w.ResolveBoolFlag(FullAuto2WeaponSwitchType, wr.FullAutoOnly)
		result.HighCyclicControlledBursts = w.ResolveBoolFlag(ControlledBursts2WeaponSwitchType, wr.HighCyclicControlledBursts)
	}
	for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, shotsFeature, secondaryFeature) {
		if bonus.Type == shotsFeature {
			result.ShotsPerAttack += bonus.AdjustedAmount()
		} else if bonus.Type == secondaryFeature {
			result.SecondaryProjectiles += bonus.AdjustedAmount()
		}
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
		AppendStringOntoNewLine(&buffer, fmt.Sprintf(i18n.Text("This weapon can only fire on full automatic. Minimum RoF is %v."), wr.ShotsPerAttack.Div(fxp.Four).Ceil()))
	}
	if wr.HighCyclicControlledBursts {
		AppendStringOntoNewLine(&buffer, i18n.Text("This weapon can fire in high cyclic controlled bursts, reducing Recoil to 1."))
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
