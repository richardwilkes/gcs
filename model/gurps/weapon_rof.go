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
	"hash"
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

// WeaponRoF holds the rate of fire data for a weapon.
type WeaponRoF struct {
	Mode1 WeaponRoFMode
	Mode2 WeaponRoFMode
}

// ParseWeaponRoF parses a string into a WeaponRoF.
func ParseWeaponRoF(s string) WeaponRoF {
	var wr WeaponRoF
	parts := strings.Split(s, "/")
	wr.Mode1 = ParseWeaponRoFMode(parts[0])
	if len(parts) > 1 {
		wr.Mode2 = ParseWeaponRoFMode(parts[1])
	}
	return wr
}

// nolint:errcheck // Not checking errors on writes to a bytes.Buffer
func (wr WeaponRoF) hash(h hash.Hash32) {
	wr.Mode1.hash(h)
	wr.Mode2.hash(h)
}

// Resolve any bonuses that apply.
func (wr WeaponRoF) Resolve(w *Weapon, modifiersTooltip *xio.ByteBuffer) WeaponRoF {
	var buf1, buf2 xio.ByteBuffer
	result := WeaponRoF{
		Mode1: wr.Mode1.Resolve(w, &buf1, true),
		Mode2: wr.Mode2.Resolve(w, &buf2, false),
	}
	if buf1.Len() != 0 {
		modifiersTooltip.WriteString(i18n.Text("First mode:\n"))
		modifiersTooltip.WriteString(buf1.String())
	}
	if buf2.Len() != 0 {
		if buf1.Len() != 0 {
			modifiersTooltip.WriteString("\n\n")
		}
		modifiersTooltip.WriteString(i18n.Text("Second mode:\n"))
		modifiersTooltip.WriteString(buf2.String())
	}
	return result
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Call .Resolve() prior to calling this method if you want the resolved values.
func (wr WeaponRoF) String(w *Weapon) string {
	if w.ResolveBoolFlag(JetWeaponSwitchType, w.Jet) {
		return i18n.Text("Jet")
	}
	s1 := wr.Mode1.String()
	s2 := wr.Mode2.String()
	if s1 == "" {
		return s2
	}
	if s2 != "" {
		return s1 + "/" + s2
	}
	return s1
}

// Tooltip returns a tooltip for the data, if any. Call .Resolve() prior to calling this method if you want the tooltip
// to be based on the resolved values.
func (wr WeaponRoF) Tooltip(w *Weapon) string {
	if w.ResolveBoolFlag(JetWeaponSwitchType, w.Jet) {
		return ""
	}
	var buffer strings.Builder
	t1 := wr.Mode1.Tooltip()
	t2 := wr.Mode2.Tooltip()
	if t1 != "" {
		buffer.WriteString(i18n.Text("First mode:\n"))
		buffer.WriteString(t1)
	}
	if t2 != "" {
		if t1 != "" {
			buffer.WriteString("\n\n")
		}
		buffer.WriteString(i18n.Text("Second mode:\n"))
		buffer.WriteString(t2)
	}
	return buffer.String()
}

// Validate ensures that the data is valid.
func (wr *WeaponRoF) Validate() {
	wr.Mode1.Validate()
	wr.Mode2.Validate()
}
