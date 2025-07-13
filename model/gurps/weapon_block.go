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
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var (
	_ json.Marshaler   = WeaponBlock{}
	_ json.Unmarshaler = &(WeaponBlock{})
)

// WeaponBlock holds the block data for a weapon.
type WeaponBlock struct {
	CanBlock bool
	Modifier fxp.Int
}

// ParseWeaponBlock parses a string into a WeaponBlock.
func ParseWeaponBlock(s string) WeaponBlock {
	var wb WeaponBlock
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	if s != "" && s != "-" && s != "–" && !strings.Contains(s, "no") {
		wb.CanBlock = true
		wb.Modifier, _ = fxp.Extract(s)
	}
	wb.Validate()
	return wb
}

// IsZero implements json.isZero.
func (wb WeaponBlock) IsZero() bool {
	return !wb.CanBlock
}

// MarshalJSON marshals the data to JSON.
func (wb WeaponBlock) MarshalJSON() ([]byte, error) {
	return json.Marshal(wb.String())
}

// UnmarshalJSON unmarshals the data from JSON.
func (wb *WeaponBlock) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*wb = ParseWeaponBlock(s)
	return nil
}

// Hash writes this object's contents into the hasher.
func (wb WeaponBlock) Hash(h hash.Hash) {
	xhash.Bool(h, wb.CanBlock)
	xhash.Num64(h, wb.Modifier)
}

// Resolve any bonuses that apply.
func (wb WeaponBlock) Resolve(w *Weapon, modifiersTooltip *xbytes.InsertBuffer) WeaponBlock {
	result := wb
	result.CanBlock = w.ResolveBoolFlag(wswitch.CanBlock, result.CanBlock)
	if result.CanBlock {
		if entity := w.Entity(); entity != nil {
			var primaryTooltip *xbytes.InsertBuffer
			if modifiersTooltip != nil {
				primaryTooltip = &xbytes.InsertBuffer{}
			}
			preAdj := w.skillLevelBaseAdjustment(entity, primaryTooltip)
			postAdj := w.skillLevelPostAdjustment(entity, primaryTooltip)
			best := fxp.Min
			for _, def := range w.Defaults {
				level := def.SkillLevelFast(entity, w.NameableReplacements(), false, nil, true)
				if level == fxp.Min {
					continue
				}
				level += preAdj
				if def.Type() != BlockID {
					level = level.Div(fxp.Two).Floor()
				}
				level += postAdj
				if best < level {
					best = level
				}
			}
			if best != fxp.Min {
				AppendBufferOntoNewLine(modifiersTooltip, primaryTooltip)
				result.Modifier += fxp.Three + best + entity.BlockBonus
				AppendStringOntoNewLine(modifiersTooltip, entity.BlockBonusTooltip)
				var percentModifier fxp.Int
				for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, feature.WeaponBlockBonus) {
					amt := bonus.AdjustedAmountForWeapon(w)
					if bonus.Percent {
						percentModifier += amt
					} else {
						result.Modifier += amt
					}
				}
				if percentModifier != 0 {
					result.Modifier += result.Modifier.Mul(percentModifier).Div(fxp.Hundred).Floor()
				}
				result.Modifier = result.Modifier.Max(0).Floor()
			} else {
				result.Modifier = 0
			}
		}
	}
	result.Validate()
	return result
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Call .Resolve() prior to calling this method if you want the resolved values.
func (wb WeaponBlock) String() string {
	if !wb.CanBlock {
		return "No"
	}
	return wb.Modifier.String()
}

// Validate ensures that the data is valid.
func (wb *WeaponBlock) Validate() {
	if !wb.CanBlock {
		wb.Modifier = 0
	}
}
