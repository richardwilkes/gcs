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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// WeaponBulk holds the bulk data for a weapon.
type WeaponBulk struct {
	Normal          fxp.Int
	Giant           fxp.Int
	RetractingStock bool
}

// ParseWeaponBulk parses a string into a WeaponBulk.
func ParseWeaponBulk(s string) WeaponBulk {
	var wb WeaponBulk
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", "")
	wb.RetractingStock = strings.Contains(s, "*")
	parts := strings.Split(s, "/")
	wb.Normal, _ = fxp.Extract(parts[0])
	if len(parts) > 1 {
		wb.Giant, _ = fxp.Extract(parts[1])
	}
	wb.Validate()
	return wb
}

// IsZero implements json.isZero.
func (wb WeaponBulk) IsZero() bool {
	return wb == WeaponBulk{}
}

// MarshalJSONTo implements json.MarshalerTo.
func (wb WeaponBulk) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, wb.String())
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (wb *WeaponBulk) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var s string
	err := json.UnmarshalDecode(dec, &s)
	if err != nil {
		return err
	}
	*wb = ParseWeaponBulk(s)
	return nil
}

// Hash writes this object's contents into the hasher.
func (wb WeaponBulk) Hash(h hash.Hash) {
	xhash.Num64(h, wb.Normal)
	xhash.Num64(h, wb.Giant)
	xhash.Bool(h, wb.RetractingStock)
}

// Resolve any bonuses that apply.
func (wb WeaponBulk) Resolve(w *Weapon, modifiersTooltip *xbytes.InsertBuffer) WeaponBulk {
	result := wb
	var percent fxp.Int
	for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, feature.WeaponBulkBonus) {
		amt := bonus.AdjustedAmountForWeapon(w)
		if bonus.Percent {
			percent += amt
		} else {
			result.Normal += amt
			result.Giant += amt
		}
	}
	if percent != 0 {
		result.Normal += result.Normal.Mul(percent).Div(fxp.Hundred).Floor()
		result.Giant += result.Giant.Mul(percent).Div(fxp.Hundred).Floor()
	}
	result.Validate()
	return result
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Call .Resolve() prior to calling this method if you want the resolved values.
func (wb WeaponBulk) String() string {
	if wb.Normal >= 0 && wb.Giant >= 0 {
		return ""
	}
	var buffer strings.Builder
	buffer.WriteString(wb.Normal.String())
	if wb.Giant != 0 && wb.Giant != wb.Normal {
		buffer.WriteByte('/')
		buffer.WriteString(wb.Giant.String())
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
	if wb.Normal < 0 {
		wb.Normal += fxp.One
	}
	if wb.Giant < 0 {
		wb.Giant += fxp.One
	}
	wb.Validate()
	accuracy := w.Accuracy.Resolve(w, nil)
	accuracy.Base -= fxp.One
	accuracy.Validate()
	recoil := w.Recoil.Resolve(w, nil)
	if recoil.Shot > fxp.One {
		recoil.Shot += fxp.One
	}
	if recoil.Slug > fxp.One {
		recoil.Slug += fxp.One
	}
	recoil.Validate()
	minST := w.Strength.Resolve(w, nil)
	minST.Min = minST.Min.Mul(fxp.OnePointTwo).Ceil()
	minST.Validate()
	return fmt.Sprintf(i18n.Text("*: Has a retracting stock. With the stock folded, the weapon's stats change to Bulk %s, Accuracy %s, Recoil %s, and minimum ST %s. Folding or unfolding the stock takes one Ready maneuver."),
		wb.String(), accuracy.String(), recoil.String(), minST.String())
}

// Validate ensures that the data is valid.
func (wb *WeaponBulk) Validate() {
	wb.Normal = wb.Normal.Min(0)
	wb.Giant = wb.Giant.Min(0)
}
