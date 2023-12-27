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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

var (
	_ json.Omitter     = WeaponBulk{}
	_ json.Marshaler   = WeaponBulk{}
	_ json.Unmarshaler = &(WeaponBulk{})
)

// WeaponBulk holds the bulk data for a weapon.
type WeaponBulk struct {
	NormalBulk      fxp.Int
	GiantBulk       fxp.Int
	RetractingStock bool
}

// ParseWeaponBulk parses a string into a WeaponBulk.
func ParseWeaponBulk(s string) WeaponBulk {
	var wb WeaponBulk
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", "")
	wb.RetractingStock = strings.Contains(s, "*")
	parts := strings.Split(s, "/")
	wb.NormalBulk, _ = fxp.Extract(parts[0])
	if len(parts) > 1 {
		wb.GiantBulk, _ = fxp.Extract(parts[1])
	}
	wb.Validate()
	return wb
}

// ShouldOmit returns true if the data should be omitted from JSON output.
func (wb WeaponBulk) ShouldOmit() bool {
	return wb == WeaponBulk{}
}

// MarshalJSON marshals the data to JSON.
func (wb WeaponBulk) MarshalJSON() ([]byte, error) {
	return json.Marshal(wb.String())
}

// UnmarshalJSON unmarshals the data from JSON.
func (wb *WeaponBulk) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*wb = ParseWeaponBulk(s)
	return nil
}

// nolint:errcheck // Not checking errors on writes to a bytes.Buffer
func (wb WeaponBulk) hash(h hash.Hash32) {
	_ = binary.Write(h, binary.LittleEndian, wb.NormalBulk)
	_ = binary.Write(h, binary.LittleEndian, wb.GiantBulk)
	_ = binary.Write(h, binary.LittleEndian, wb.RetractingStock)
}

// Resolve any bonuses that apply.
func (wb WeaponBulk) Resolve(w *Weapon, modifiersTooltip *xio.ByteBuffer) WeaponBulk {
	result := wb
	for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, feature.WeaponBulkBonus) {
		result.NormalBulk += bonus.AdjustedAmount()
		result.GiantBulk += bonus.AdjustedAmount()
	}
	result.Validate()
	return result
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Call .Resolve() prior to calling this method if you want the resolved values.
func (wb WeaponBulk) String() string {
	if wb.NormalBulk >= 0 && wb.GiantBulk >= 0 {
		return ""
	}
	var buffer strings.Builder
	buffer.WriteString(wb.NormalBulk.String())
	if wb.GiantBulk != 0 && wb.GiantBulk != wb.NormalBulk {
		buffer.WriteByte('/')
		buffer.WriteString(wb.GiantBulk.String())
	}
	if wb.RetractingStock {
		buffer.WriteByte('*')
	}
	return buffer.String()
}

// Tooltip returns a tooltip for the data, if any. Call .Resolve() prior to calling this method if you want the tooltip
// to be based on the resolved values.
func (wb WeaponBulk) Tooltip(w *Weapon) string {
	if !wb.RetractingStock {
		return ""
	}
	if wb.NormalBulk < 0 {
		wb.NormalBulk += fxp.One
	}
	if wb.GiantBulk < 0 {
		wb.GiantBulk += fxp.One
	}
	wb.Validate()
	accuracy := w.Accuracy.Resolve(w, nil)
	accuracy.Base -= fxp.One
	accuracy.Validate()
	recoil := w.Recoil.Resolve(w, nil)
	if recoil.ShotRecoil > fxp.One {
		recoil.ShotRecoil += fxp.One
	}
	if recoil.SlugRecoil > fxp.One {
		recoil.SlugRecoil += fxp.One
	}
	recoil.Validate()
	minST := w.Strength.Resolve(w, nil)
	minST.Minimum = minST.Minimum.Mul(fxp.OnePointTwo).Ceil()
	minST.Validate()
	return fmt.Sprintf(i18n.Text("Has a retracting stock. With the stock folded, the weapon's stats change to Bulk %s, Accuracy %s, Recoil %s, and minimum ST %s. Folding or unfolding the stock takes one Ready maneuver."),
		wb.String(), accuracy.String(), recoil.String(), minST.String())
}

// Validate ensures that the data is valid.
func (wb *WeaponBulk) Validate() {
	wb.NormalBulk = wb.NormalBulk.Min(0)
	wb.GiantBulk = wb.GiantBulk.Min(0)
}
