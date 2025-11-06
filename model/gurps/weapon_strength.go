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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// WeaponStrength holds the minimum strength data for a weapon.
type WeaponStrength struct {
	Min              fxp.Int
	Bipod            bool
	Mounted          bool
	MusketRest       bool
	TwoHanded        bool
	TwoHandedUnready bool
}

// ParseWeaponStrength parses a string into a WeaponStrength.
func ParseWeaponStrength(s string) WeaponStrength {
	var ws WeaponStrength
	s = strings.ReplaceAll(s, " ", "")
	if s != "" {
		s = strings.ToLower(s)
		ws.Bipod = strings.Contains(s, "b")
		ws.Mounted = strings.Contains(s, "m")
		ws.MusketRest = strings.Contains(s, "r")
		ws.TwoHanded = strings.Contains(s, "†") || strings.Contains(s, "*") // bad input in some files had * instead of †
		ws.TwoHandedUnready = strings.Contains(s, "‡")
		ws.Min, _ = fxp.Extract(s)
		ws.Validate()
	}
	return ws
}

// MarshalJSONTo implements json.MarshalerTo.
func (ws WeaponStrength) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, ws.String())
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (ws *WeaponStrength) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var s string
	if err := json.UnmarshalDecode(dec, &s); err != nil {
		return err
	}
	*ws = ParseWeaponStrength(s)
	return nil
}

// IsZero implements json.isZero.
func (ws WeaponStrength) IsZero() bool {
	return ws == WeaponStrength{}
}

// Hash writes this object's contents into the hasher.
func (ws WeaponStrength) Hash(h hash.Hash) {
	xhash.Num64(h, ws.Min)
	xhash.Bool(h, ws.Bipod)
	xhash.Bool(h, ws.Mounted)
	xhash.Bool(h, ws.MusketRest)
	xhash.Bool(h, ws.TwoHanded)
	xhash.Bool(h, ws.TwoHandedUnready)
}

// Resolve any bonuses that apply.
func (ws WeaponStrength) Resolve(w *Weapon, modifiersTooltip *xbytes.InsertBuffer) WeaponStrength {
	result := ws
	result.Bipod = w.ResolveBoolFlag(wswitch.Bipod, result.Bipod)
	result.Mounted = w.ResolveBoolFlag(wswitch.Mounted, result.Mounted)
	result.MusketRest = w.ResolveBoolFlag(wswitch.MusketRest, result.MusketRest)
	result.TwoHanded = w.ResolveBoolFlag(wswitch.TwoHanded, result.TwoHanded)
	result.TwoHandedUnready = w.ResolveBoolFlag(wswitch.TwoHandedAndUnreadyAfterAttack, result.TwoHandedUnready)
	var percentMin fxp.Int
	for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, feature.WeaponMinSTBonus) {
		amt := bonus.AdjustedAmountForWeapon(w)
		if bonus.Percent {
			percentMin += amt
		} else {
			result.Min += amt
		}
	}
	if percentMin != 0 {
		result.Min += result.Min.Mul(percentMin).Div(fxp.Hundred).Floor()
	}
	result.Validate()
	return result
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Call .Resolve() prior to calling this method if you want the resolved values.
func (ws WeaponStrength) String() string {
	var buffer strings.Builder
	if ws.Min > 0 {
		buffer.WriteString(ws.Min.String())
	}
	if ws.Bipod {
		buffer.WriteByte('B')
	}
	if ws.Mounted {
		buffer.WriteByte('M')
	}
	if ws.MusketRest {
		buffer.WriteByte('R')
	}
	if ws.TwoHanded || ws.TwoHandedUnready {
		if ws.TwoHandedUnready {
			buffer.WriteRune('‡')
		} else {
			buffer.WriteRune('†')
		}
	}
	return buffer.String()
}

// Tooltip returns a tooltip for the data, if any. Call .Resolve() prior to calling this method if you want the tooltip
// to be based on the resolved values.
func (ws WeaponStrength) Tooltip(w *Weapon) string {
	var tooltip strings.Builder
	if w.Owner != nil {
		if st := w.Owner.RatedStrength(); st > 0 {
			fmt.Fprintf(&tooltip, i18n.Text("The weapon has a rated ST of %v, which is used instead of the user's ST for all calculations except minimum ST."), st)
		}
	}
	if ws.Min > 0 {
		if tooltip.Len() != 0 {
			tooltip.WriteString("\n\n")
		}
		fmt.Fprintf(&tooltip, i18n.Text("The weapon has a minimum ST of %v. If your ST is less than this, you will suffer a -1 to weapon skill per point of ST you lack and lose one extra FP at the end of any fight that lasts long enough to cost FP."), ws.Min)
	}
	if ws.Bipod {
		if tooltip.Len() != 0 {
			tooltip.WriteString("\n\n")
		}
		tooltip.WriteString(i18n.Text("B: Has an attached bipod. When used from a prone position, "))
		reducedST := ws.Min.Mul(fxp.Two).Div(fxp.Three).Ceil()
		if reducedST > 0 && reducedST != ws.Min {
			fmt.Fprintf(&tooltip, i18n.Text("reduces the ST requirement to %v and "), reducedST)
		}
		tooltip.WriteString(i18n.Text("treats the attack as braced (add +1 to Accuracy)."))
	}
	if ws.Mounted {
		if tooltip.Len() != 0 {
			tooltip.WriteString("\n\n")
		}
		tooltip.WriteString(i18n.Text("M: Mounted. Ignore listed ST and Bulk when firing from its mount. Takes at least 3 Ready maneuvers to unmount or remount the weapon."))
	}
	if ws.MusketRest {
		if tooltip.Len() != 0 {
			tooltip.WriteString("\n\n")
		}
		tooltip.WriteString(i18n.Text("R: Uses a Musket Rest. Any aimed shot fired while stationary and standing up is automatically braced (add +1 to Accuracy)."))
	}
	if ws.TwoHanded || ws.TwoHandedUnready {
		if tooltip.Len() != 0 {
			tooltip.WriteString("\n\n")
		}
		if ws.TwoHandedUnready {
			fmt.Fprintf(&tooltip, i18n.Text("‡: Requires two hands and becomes unready after you attack with it. If you have at least ST %v, you can used it two-handed without it becoming unready. If you have at least ST %v, you can use it one-handed with no readiness penalty."), ws.Min.Mul(fxp.OneAndAHalf).Ceil(), ws.Min.Mul(fxp.Three).Ceil())
		} else {
			fmt.Fprintf(&tooltip, i18n.Text("†: Requires two hands. If you have at least ST %v, you can use it one-handed, but it becomes unready after you attack with it. If you have at least ST %v, you can use it one-handed with no readiness penalty."), ws.Min.Mul(fxp.OneAndAHalf).Ceil(), ws.Min.Mul(fxp.Two).Ceil())
		}
	}
	return tooltip.String()
}

// Validate ensures that the data is valid.
func (ws *WeaponStrength) Validate() {
	ws.Min = ws.Min.Max(0)
	if ws.TwoHanded && ws.TwoHandedUnready {
		ws.TwoHanded = false
	}
}
