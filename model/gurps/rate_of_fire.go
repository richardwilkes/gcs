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

// RateOfFire holds the rate of fire data for one firing mode of a weapon.
type RateOfFire struct {
	ShotsPerAttack             fxp.Int `json:"shots_per_attack,omitempty"`
	SecondaryProjectiles       fxp.Int `json:"secondary_projectiles,omitempty"`
	FullAutoOnly               bool    `json:"full_auto_only,omitempty"`
	HighCyclicControlledBursts bool    `json:"high_cyclic_controlled_bursts,omitempty"`
}

// ShouldOmit returns true if this RoF should be omitted from the JSON.
func (r RateOfFire) ShouldOmit() bool {
	return r == (RateOfFire{})
}

// nolint:errcheck // Not checking errors on writes to a bytes.Buffer
func (r RateOfFire) hash(h hash.Hash32) {
	_ = binary.Write(h, binary.LittleEndian, r.ShotsPerAttack)
	_ = binary.Write(h, binary.LittleEndian, r.SecondaryProjectiles)
	_ = binary.Write(h, binary.LittleEndian, r.FullAutoOnly)
	_ = binary.Write(h, binary.LittleEndian, r.HighCyclicControlledBursts)
}

func (r *RateOfFire) parseOldRateOfFire(s string) {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, ".", "x") // Fix some faulty input that exists in the old files
	r.FullAutoOnly = strings.Contains(s, "!")
	s = strings.ReplaceAll(s, "!", "")
	r.HighCyclicControlledBursts = strings.Contains(s, "#")
	s = strings.ReplaceAll(s, "#", "")
	s = strings.ReplaceAll(s, "×", "x") // Fix some more faulty input that exists in the old files
	if strings.HasPrefix(s, "x") {
		s = "1" + s // Fix some more faulty input that exists in the old files
	}
	parts := strings.Split(s, "x")
	r.ShotsPerAttack, _ = fxp.Extract(s)
	if len(parts) > 1 {
		r.SecondaryProjectiles, _ = fxp.Extract(parts[1])
	}
}

// Combined returns a string combining the RoF data.
func (r RateOfFire) Combined(tooltip *xio.ByteBuffer) string {
	spa := r.ShotsPerAttack.Ceil()
	if spa <= 0 {
		return ""
	}
	var buffer strings.Builder
	buffer.WriteString(spa.String())
	sp := r.SecondaryProjectiles.Ceil()
	if sp > 0 {
		buffer.WriteByte('x')
		buffer.WriteString(sp.String())
		shotsText := i18n.Text("shots")
		if spa == fxp.One {
			shotsText = i18n.Text("shot")
		}
		projectilesText := i18n.Text("projectiles")
		if sp == fxp.One {
			projectilesText = i18n.Text("projectile")
		}
		AppendStringOntoNewLine(tooltip, fmt.Sprintf(i18n.Text("This weapon fires %v %s per attack and each shot releases %v smaller %s."), spa, shotsText, sp, projectilesText))
	}
	if r.FullAutoOnly {
		buffer.WriteByte('!')
		AppendStringOntoNewLine(tooltip, fmt.Sprintf(i18n.Text("This weapon can only fire on full automatic. Minimum RoF is %v."), r.ShotsPerAttack.Div(fxp.Four).Ceil()))
	}
	if r.HighCyclicControlledBursts {
		buffer.WriteByte('#')
		AppendStringOntoNewLine(tooltip, i18n.Text("This weapon can fire in high cyclic controlled bursts, reducing Recoil to 1."))
	}
	return buffer.String()
}
