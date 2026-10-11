// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package attack

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// Shots holds the shots data for a weapon.
type Shots struct {
	Count               fxp.Int
	InChamber           fxp.Int
	Duration            fxp.Int
	ReloadTime          fxp.Int
	ReloadTimeIsPerShot bool
	Thrown              bool
}

// ParseShots parses a string into a Shots.
func ParseShots(s string) Shots {
	var ws Shots
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", "")
	if !strings.Contains(s, "fp") &&
		!strings.Contains(s, "hrs") &&
		!strings.Contains(s, "day") {
		ws.Thrown = strings.Contains(s, "t")
		if !strings.Contains(s, "spec") {
			ws.Count, s = fxp.Extract(s)
			if strings.HasPrefix(s, "+") {
				ws.InChamber, s = fxp.Extract(s)
			}
			if strings.HasPrefix(s, "x") {
				ws.Duration, s = fxp.Extract(s[1:])
			}
			// Locate the reload time rather than requiring it to be next, since the thrown marker or the duration's
			// trailing "s" may precede it.
			if i := strings.IndexByte(s, '('); i >= 0 {
				s = s[i+1:]
				ws.ReloadTime, _ = fxp.Extract(s)
				ws.ReloadTimeIsPerShot = strings.Contains(s, "i")
			}
		}
	}
	ws.Validate()
	return ws
}

// MarshalJSONTo implements json.MarshalerTo.
func (ws Shots) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, ws.String())
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (ws *Shots) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return jio.UnmarshalStringFromInfallible(dec, ws, ParseShots)
}

// IsZero reports whether json's omitzero option should omit this value.
func (ws Shots) IsZero() bool {
	return ws == Shots{}
}

// Hash writes this object's contents into the hasher.
func (ws Shots) Hash(h hash.Hash) {
	xhash.Num64(h, ws.Count)
	xhash.Num64(h, ws.InChamber)
	xhash.Num64(h, ws.Duration)
	xhash.Num64(h, ws.ReloadTime)
	xhash.Bool(h, ws.ReloadTimeIsPerShot)
	xhash.Bool(h, ws.Thrown)
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Use the weapon's ResolvedShots method first if you want the resolved values.
func (ws Shots) String() string {
	var buffer strings.Builder
	if ws.Thrown {
		buffer.WriteByte('T')
	} else {
		if ws.Count <= 0 && ws.InChamber <= 0 {
			return ""
		}
		buffer.WriteString(ws.Count.Max(0).String())
		if ws.InChamber > 0 {
			buffer.WriteByte('+')
			buffer.WriteString(ws.InChamber.String())
		}
		if ws.Duration > 0 {
			buffer.WriteByte('x')
			buffer.WriteString(ws.Duration.String())
			buffer.WriteByte('s')
		}
	}
	if ws.ReloadTime > 0 {
		buffer.WriteByte('(')
		buffer.WriteString(ws.ReloadTime.String())
		if ws.ReloadTimeIsPerShot {
			buffer.WriteByte('i')
		}
		buffer.WriteByte(')')
	}
	return buffer.String()
}

// Tooltip returns a tooltip for the data, if any. Use the weapon's ResolvedShots method first if you want the tooltip
// to be based on the resolved values.
func (ws Shots) Tooltip() string {
	if ws.ReloadTimeIsPerShot {
		return i18n.Text("i: Reload time is per shot")
	}
	return ""
}

// Validate ensures that the data is valid.
func (ws *Shots) Validate() {
	ws.ReloadTime = ws.ReloadTime.Max(0)
	if ws.Thrown {
		ws.Count = 0
		ws.InChamber = 0
		ws.Duration = 0
		return
	}
	ws.Count = ws.Count.Max(0)
	ws.InChamber = ws.InChamber.Max(0)
	if ws.Count == 0 && ws.InChamber == 0 {
		ws.Duration = 0
		ws.ReloadTime = 0
		return
	}
	ws.Duration = ws.Duration.Max(0)
}
