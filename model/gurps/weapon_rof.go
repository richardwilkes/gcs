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
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var (
	_ json.Omitter     = WeaponRoF{}
	_ json.Marshaler   = WeaponRoF{}
	_ json.Unmarshaler = &(WeaponRoF{})
)

// WeaponRoF holds the rate of fire data for a weapon.
type WeaponRoF struct {
	Mode1 WeaponRoFMode
	Mode2 WeaponRoFMode
	Jet   bool
}

// ParseWeaponRoF parses a string into a WeaponRoF.
func ParseWeaponRoF(s string) WeaponRoF {
	var wr WeaponRoF
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ToLower(s)
	if strings.Contains(s, "jet") {
		wr.Jet = true
	} else {
		parts := strings.Split(s, "/")
		wr.Mode1 = ParseWeaponRoFMode(parts[0])
		if len(parts) > 1 {
			wr.Mode2 = ParseWeaponRoFMode(parts[1])
		}
	}
	wr.Validate()
	return wr
}

// ShouldOmit returns true if the data should be omitted from JSON output.
func (wr WeaponRoF) ShouldOmit() bool {
	return wr == WeaponRoF{}
}

// MarshalJSON marshals the data to JSON.
func (wr WeaponRoF) MarshalJSON() ([]byte, error) {
	return json.Marshal(wr.String())
}

// UnmarshalJSON unmarshals the data from JSON.
func (wr *WeaponRoF) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*wr = ParseWeaponRoF(s)
	return nil
}

// Hash writes this object's contents into the hasher.
func (wr WeaponRoF) Hash(h hash.Hash) {
	if wr.ShouldOmit() {
		hashhelper.Num8(h, uint8(255))
		return
	}
	wr.Mode1.Hash(h)
	wr.Mode2.Hash(h)
	hashhelper.Bool(h, wr.Jet)
}

// Resolve any bonuses that apply.
func (wr WeaponRoF) Resolve(w *Weapon, modifiersTooltip *xio.ByteBuffer) WeaponRoF {
	result := wr
	result.Jet = w.ResolveBoolFlag(wswitch.Jet, result.Jet)
	if !result.Jet {
		var buf1, buf2 xio.ByteBuffer
		result.Mode1 = result.Mode1.Resolve(w, &buf1, true)
		result.Mode2 = result.Mode2.Resolve(w, &buf2, false)
		if modifiersTooltip != nil {
			if buf1.Len() != 0 {
				if buf2.Len() != 0 {
					modifiersTooltip.WriteString(i18n.Text("First mode:\n"))
				}
				modifiersTooltip.WriteString(buf1.String())
			}
			if buf2.Len() != 0 {
				if buf1.Len() != 0 {
					modifiersTooltip.WriteString(i18n.Text("\n\nSecond mode:\n"))
				}
				modifiersTooltip.WriteString(buf2.String())
			}
		}
	}
	result.Validate()
	return result
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Call .Resolve() prior to calling this method if you want the resolved values.
func (wr WeaponRoF) String() string {
	if wr.Jet {
		return "Jet" // Not localized, since it is part of the data
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
func (wr WeaponRoF) Tooltip() string {
	if wr.Jet {
		return ""
	}
	var buffer strings.Builder
	t1 := wr.Mode1.Tooltip()
	t2 := wr.Mode2.Tooltip()
	if t1 != "" {
		if t2 != "" {
			buffer.WriteString(i18n.Text("First mode:\n"))
		}
		buffer.WriteString(t1)
	}
	if t2 != "" {
		if t1 != "" {
			buffer.WriteString(i18n.Text("\n\nSecond mode:\n"))
		}
		buffer.WriteString(t2)
	}
	return buffer.String()
}

// Validate ensures that the data is valid.
func (wr *WeaponRoF) Validate() {
	if wr.Jet {
		wr.Mode1 = WeaponRoFMode{}
		wr.Mode2 = WeaponRoFMode{}
		return
	}
	wr.Mode1.Validate()
	wr.Mode2.Validate()
}
