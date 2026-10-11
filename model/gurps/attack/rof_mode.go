// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package attack

import (
	"fmt"
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/txtutil"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// RoFMode holds the rate of fire data for one firing mode of a weapon.
type RoFMode struct {
	ShotsPerAttack             fxp.Int
	SecondaryProjectiles       fxp.Int
	FullAutoOnly               bool
	HighCyclicControlledBursts bool
}

// ParseRoFMode parses a string into a RoFMode.
func ParseRoFMode(s string) RoFMode {
	var wr RoFMode
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, ".", "x")
	wr.FullAutoOnly = strings.Contains(s, "!")
	s = strings.ReplaceAll(s, "!", "")
	wr.HighCyclicControlledBursts = strings.Contains(s, "#")
	s = strings.ReplaceAll(s, "#", "")
	s = strings.ReplaceAll(s, fxp.MultiplicationSign, "x")
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
func (wr RoFMode) Hash(h hash.Hash) {
	xhash.Num64(h, wr.ShotsPerAttack)
	xhash.Num64(h, wr.SecondaryProjectiles)
	xhash.Bool(h, wr.FullAutoOnly)
	xhash.Bool(h, wr.HighCyclicControlledBursts)
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Use the weapon's ResolvedRateOfFire method first if you want the resolved values.
func (wr RoFMode) String() string {
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

// Tooltip returns a tooltip for the data, if any. Use the weapon's ResolvedRateOfFire method first if you want the
// tooltip to be based on the resolved values.
func (wr RoFMode) Tooltip() string {
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
		txtutil.AppendStringOntoNewLine(&buffer, i18n.Text("!: This weapon can only fire on full automatic. Minimum RoF is %v.", wr.ShotsPerAttack.Div(fxp.Four).Ceil()))
	}
	if wr.HighCyclicControlledBursts {
		txtutil.AppendStringOntoNewLine(&buffer, i18n.Text("#: This weapon can fire in high cyclic controlled bursts, reducing Recoil to 1."))
	}
	return buffer.String()
}

// Validate ensures that the data is valid.
func (wr *RoFMode) Validate() {
	wr.ShotsPerAttack = wr.ShotsPerAttack.Ceil().Max(0)
	if wr.ShotsPerAttack == 0 {
		wr.SecondaryProjectiles = 0
		wr.FullAutoOnly = false
		wr.HighCyclicControlledBursts = false
		return
	}
	wr.SecondaryProjectiles = wr.SecondaryProjectiles.Ceil().Max(0)
}
