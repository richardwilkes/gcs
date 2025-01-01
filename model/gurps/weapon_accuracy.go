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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var (
	_ json.Omitter     = WeaponAccuracy{}
	_ json.Marshaler   = WeaponAccuracy{}
	_ json.Unmarshaler = &(WeaponAccuracy{})
)

// WeaponAccuracy holds the accuracy data for a weapon.
type WeaponAccuracy struct {
	Base  fxp.Int
	Scope fxp.Int
	Jet   bool
}

// ParseWeaponAccuracy parses a string into a WeaponAccuracy.
func ParseWeaponAccuracy(s string) WeaponAccuracy {
	var wa WeaponAccuracy
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ToLower(s)
	if strings.Contains(s, "jet") {
		wa.Jet = true
	} else {
		s = strings.TrimPrefix(s, "+")
		parts := strings.Split(s, "+")
		wa.Base, _ = fxp.Extract(parts[0])
		if len(parts) > 1 {
			wa.Scope, _ = fxp.Extract(parts[1])
		}
	}
	wa.Validate()
	return wa
}

// ShouldOmit returns true if the data should be omitted from JSON output.
func (wa WeaponAccuracy) ShouldOmit() bool {
	return wa == WeaponAccuracy{}
}

// MarshalJSON marshals the data to JSON.
func (wa WeaponAccuracy) MarshalJSON() ([]byte, error) {
	return json.Marshal(wa.String())
}

// UnmarshalJSON unmarshals the data from JSON.
func (wa *WeaponAccuracy) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*wa = ParseWeaponAccuracy(s)
	return nil
}

// Hash writes this object's contents into the hasher.
func (wa WeaponAccuracy) Hash(h hash.Hash) {
	if wa.ShouldOmit() {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.Num64(h, wa.Base)
	hashhelper.Num64(h, wa.Scope)
	hashhelper.Bool(h, wa.Jet)
}

// Resolve any bonuses that apply.
func (wa WeaponAccuracy) Resolve(w *Weapon, modifiersTooltip *xio.ByteBuffer) WeaponAccuracy {
	result := wa
	result.Jet = w.ResolveBoolFlag(wswitch.Jet, result.Jet)
	if !result.Jet {
		if entity := w.Entity(); entity != nil {
			var percentBase, percentScope fxp.Int
			for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, feature.WeaponAccBonus, feature.WeaponScopeAccBonus) {
				amt := bonus.AdjustedAmountForWeapon(w)
				switch bonus.Type {
				case feature.WeaponAccBonus:
					if bonus.Percent {
						percentBase += amt
					} else {
						result.Base += amt
					}
				case feature.WeaponScopeAccBonus:
					if bonus.Percent {
						percentScope += amt
					} else {
						result.Scope += amt
					}
				default:
				}
			}
			if percentBase != 0 {
				result.Base += result.Base.Mul(percentBase).Div(fxp.Hundred).Trunc()
			}
			if percentScope != 0 {
				result.Scope += result.Scope.Mul(percentScope).Div(fxp.Hundred).Trunc()
			}
		}
	}
	result.Validate()
	return result
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Call .Resolve() prior to calling this method if you want the resolved values.
func (wa WeaponAccuracy) String() string {
	if wa.Jet {
		return "Jet" // Not localized, since it is part of the data
	}
	if wa.Scope != 0 {
		return wa.Base.String() + wa.Scope.StringWithSign()
	}
	return wa.Base.String()
}

// Validate ensures that the data is valid.
func (wa *WeaponAccuracy) Validate() {
	if wa.Jet {
		wa.Base = 0
		wa.Scope = 0
		return
	}
	wa.Base = wa.Base.Max(0)
	wa.Scope = wa.Scope.Max(0)
}
