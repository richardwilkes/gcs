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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var (
	_ json.Omitter     = WeaponRecoil{}
	_ json.Marshaler   = WeaponRecoil{}
	_ json.Unmarshaler = &(WeaponRecoil{})
)

// WeaponRecoil holds the recoil data for a weapon.
type WeaponRecoil struct {
	Shot fxp.Int
	Slug fxp.Int
}

// ParseWeaponRecoil parses a string into a WeaponRecoil.
func ParseWeaponRecoil(s string) WeaponRecoil {
	var wr WeaponRecoil
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", "")
	parts := strings.Split(s, "/")
	wr.Shot, _ = fxp.Extract(parts[0])
	if len(parts) > 1 {
		wr.Slug, _ = fxp.Extract(parts[1])
	}
	return wr
}

// ShouldOmit returns true if the data should be omitted from JSON output.
func (wr WeaponRecoil) ShouldOmit() bool {
	return wr == WeaponRecoil{}
}

// MarshalJSON marshals the data to JSON.
func (wr WeaponRecoil) MarshalJSON() ([]byte, error) {
	return json.Marshal(wr.String())
}

// UnmarshalJSON unmarshals the data from JSON.
func (wr *WeaponRecoil) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*wr = ParseWeaponRecoil(s)
	return nil
}

// Hash writes this object's contents into the hasher.
func (wr WeaponRecoil) Hash(h hash.Hash) {
	if wr.ShouldOmit() {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.Num64(h, wr.Shot)
	hashhelper.Num64(h, wr.Slug)
}

// Resolve any bonuses that apply.
func (wr WeaponRecoil) Resolve(w *Weapon, modifiersTooltip *xio.ByteBuffer) WeaponRecoil {
	result := wr
	// 0 means recoil isn't used; 1+ means it is.
	if wr.Shot > 0 || wr.Slug > 0 {
		var percent fxp.Int
		for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, feature.WeaponRecoilBonus) {
			amt := bonus.AdjustedAmountForWeapon(w)
			if bonus.Percent {
				percent += amt
			} else {
				result.Shot += amt
				result.Slug += amt
			}
		}
		if percent != 0 {
			result.Shot += result.Shot.Mul(percent).Div(fxp.Hundred).Trunc()
			result.Slug += result.Slug.Mul(percent).Div(fxp.Hundred).Trunc()
		}
		if wr.Shot > 0 {
			result.Shot = result.Shot.Max(fxp.One)
		} else {
			result.Shot = 0
		}
		if wr.Slug > 0 {
			result.Slug = result.Slug.Max(fxp.One)
		} else {
			result.Slug = 0
		}
	}
	result.Validate()
	return result
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Call .Resolve() prior to calling this method if you want the resolved values.
func (wr WeaponRecoil) String() string {
	if wr.Shot == 0 && wr.Slug == 0 {
		return ""
	}
	var buffer strings.Builder
	buffer.WriteString(wr.Shot.String())
	if wr.Slug != 0 {
		buffer.WriteByte('/')
		buffer.WriteString(wr.Slug.String())
	}
	return buffer.String()
}

// Tooltip returns a tooltip for the data, if any. Call .Resolve() prior to calling this method if you want the tooltip
// to be based on the resolved values.
func (wr WeaponRecoil) Tooltip() string {
	if wr.Shot != 0 && wr.Slug != 0 && wr.Shot != wr.Slug {
		return i18n.Text("First Recoil value is for shot, second is for slugs")
	}
	return ""
}

// Validate ensures that the data is valid.
func (wr *WeaponRecoil) Validate() {
	wr.Shot = wr.Shot.Max(0)
	wr.Slug = wr.Slug.Max(0)
}
