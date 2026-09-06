// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// WeaponRange holds the range data for a weapon.
type WeaponRange struct {
	HalfDamage    fxp.Int
	Min           fxp.Int
	Max           fxp.Int
	MusclePowered bool
	InMiles       bool
}

// ParseWeaponRange parses a string into a WeaponRange.
func ParseWeaponRange(s string) WeaponRange {
	var wr WeaponRange
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, fxp.MultiplicationSign, "x")
	if !strings.Contains(s, "sight") &&
		!strings.Contains(s, "spec") &&
		!strings.Contains(s, "skill") &&
		!strings.Contains(s, "point") &&
		!strings.Contains(s, "pbaoe") &&
		!strings.HasPrefix(s, "b") {
		s = strings.ReplaceAll(s, ",max", "/")
		s = strings.ReplaceAll(s, "max", "")
		s = strings.ReplaceAll(s, "1/2d", "")
		wr.MusclePowered = strings.Contains(s, "x")
		s = strings.ReplaceAll(s, "x", "")
		s = strings.ReplaceAll(s, "st", "")
		s = strings.ReplaceAll(s, "c/", "")
		wr.InMiles = strings.Contains(s, "mi")
		s = strings.ReplaceAll(s, "mi.", "")
		s = strings.ReplaceAll(s, "mi", "")
		s = strings.ReplaceAll(s, ",", "")
		parts := strings.Split(s, "/")
		if len(parts) > 1 {
			wr.HalfDamage, _ = fxp.Extract(parts[0])
			parts[0] = parts[1]
		}
		parts = strings.Split(parts[0], "-")
		if len(parts) > 1 {
			wr.Min, _ = fxp.Extract(parts[0])
			wr.Max, _ = fxp.Extract(parts[1])
		} else {
			wr.Max, _ = fxp.Extract(parts[0])
		}
		wr.Validate()
	}
	return wr
}

// MarshalJSONTo implements json.MarshalerTo.
//
// The data is persisted as the unresolved display string. Passing false for musclePowerIsResolved is deliberate and
// load-bearing: it emits the "x" prefixes (e.g. "x10/x100") that encode the MusclePowered flag, which ParseWeaponRange
// detects on load to restore that flag. Passing true here would silently drop MusclePowered on a save/reload
// round-trip.
func (wr WeaponRange) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, wr.String(false))
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (wr *WeaponRange) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return jio.UnmarshalStringFromInfallible(dec, wr, ParseWeaponRange)
}

// IsZero implements json.isZero.
func (wr WeaponRange) IsZero() bool {
	return wr == WeaponRange{}
}

// Hash writes this object's contents into the hasher.
func (wr WeaponRange) Hash(h hash.Hash) {
	xhash.Num64(h, wr.HalfDamage)
	xhash.Num64(h, wr.Min)
	xhash.Num64(h, wr.Max)
	xhash.Bool(h, wr.MusclePowered)
	xhash.Bool(h, wr.InMiles)
}

// Resolve any bonuses that apply.
func (wr WeaponRange) Resolve(w *Weapon, modifiersTooltip *xbytes.InsertBuffer) WeaponRange {
	result := wr
	result.MusclePowered = w.ResolveBoolFlag(wswitch.MusclePowered, result.MusclePowered)
	result.InMiles = w.ResolveBoolFlag(wswitch.RangeInMiles, result.InMiles)
	if result.MusclePowered {
		var st fxp.Int
		maxST := w.Strength.Resolve(w, nil).Min.Mul(fxp.Three)
		if w.Owner != nil {
			st = w.Owner.RatedStrength()
		}
		if st == 0 {
			if entity := w.Entity(); entity != nil {
				switch w.Damage.resolvedStrengthType(nil) {
				case stdmg.TelekineticThrust, stdmg.TelekineticSwing:
					st = entity.TelekineticStrength()
				case stdmg.IQThrust, stdmg.IQSwing:
					st = entity.ResolveAttributeCurrent(IntelligenceID).Max(0).Floor()
				default:
					st = entity.ThrowingStrength()
				}
			}
		}
		st = max(w.weaponAdjustment(oneDieCount, modifiersTooltip, feature.WeaponEffectiveSTBonus).applyTo(st), 0)
		if maxST > 0 && maxST < st {
			st = maxST
		}
		if st > 0 {
			result.HalfDamage = result.HalfDamage.Mul(st).Floor().Max(0)
			result.Min = result.Min.Mul(st).Floor().Max(0)
			result.Max = result.Max.Mul(st).Floor().Max(0)
		}
	}
	adj := w.weaponAdjustments(w.baseDamageDieCount, modifiersTooltip, feature.WeaponHalfDamageRangeBonus,
		feature.WeaponMinRangeBonus, feature.WeaponMaxRangeBonus)
	result.HalfDamage = adj[feature.WeaponHalfDamageRangeBonus].applyTo(result.HalfDamage)
	result.Min = adj[feature.WeaponMinRangeBonus].applyTo(result.Min)
	result.Max = adj[feature.WeaponMaxRangeBonus].applyTo(result.Max)
	result.Validate()
	return result
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Call .Resolve() prior to calling this method if you want the resolved values.
//
// When MusclePowered is set, passing false for musclePowerIsResolved prefixes the range values with "x" (the GURPS
// notation for "multiply by ST"); pass true once the ranges have already been multiplied by ST so the "x" is omitted.
// ParseWeaponRange relies on the "x" prefix to recover the MusclePowered flag, so persistence must stringify with
// false.
func (wr WeaponRange) String(musclePowerIsResolved bool) string {
	var buffer strings.Builder
	if wr.HalfDamage != 0 {
		if wr.MusclePowered && !musclePowerIsResolved {
			buffer.WriteByte('x')
		}
		buffer.WriteString(wr.HalfDamage.Comma())
		buffer.WriteByte('/')
	}
	if wr.Min != 0 || wr.Max != 0 {
		if wr.Min != 0 && wr.Min != wr.Max {
			if wr.MusclePowered && !musclePowerIsResolved {
				buffer.WriteByte('x')
			}
			buffer.WriteString(wr.Min.Comma())
			buffer.WriteByte('-')
		}
		if wr.MusclePowered && !musclePowerIsResolved {
			buffer.WriteByte('x')
		}
		buffer.WriteString(wr.Max.Comma())
	}
	if wr.InMiles && buffer.Len() != 0 {
		buffer.WriteByte(' ')
		buffer.WriteString(fxp.Mile.String())
	}
	return buffer.String()
}

// Validate ensures that the data is valid.
func (wr *WeaponRange) Validate() {
	wr.HalfDamage = wr.HalfDamage.Max(0)
	wr.Min = wr.Min.Max(0)
	wr.Max = wr.Max.Max(0)
	if wr.Min > wr.Max {
		wr.Min, wr.Max = wr.Max, wr.Min
	}
	if wr.HalfDamage < wr.Min || wr.HalfDamage >= wr.Max {
		wr.HalfDamage = 0
	}
}
