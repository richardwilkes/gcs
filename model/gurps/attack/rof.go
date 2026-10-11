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

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// RoF holds the rate of fire data for a weapon.
type RoF struct {
	Mode1 RoFMode
	Mode2 RoFMode
	Jet   bool
}

// ParseRoF parses a string into a RoF.
func ParseRoF(s string) RoF {
	var wr RoF
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ToLower(s)
	if strings.Contains(s, "jet") {
		wr.Jet = true
	} else {
		parts := strings.Split(s, "/")
		wr.Mode1 = ParseRoFMode(parts[0])
		if len(parts) > 1 {
			wr.Mode2 = ParseRoFMode(parts[1])
		}
	}
	wr.Validate()
	return wr
}

// MarshalJSONTo implements json.MarshalerTo.
func (wr RoF) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, wr.String())
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (wr *RoF) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return jio.UnmarshalStringFromInfallible(dec, wr, ParseRoF)
}

// IsZero reports whether json's omitzero option should omit this value.
func (wr RoF) IsZero() bool {
	return wr == RoF{}
}

// Hash writes this object's contents into the hasher.
func (wr RoF) Hash(h hash.Hash) {
	wr.Mode1.Hash(h)
	wr.Mode2.Hash(h)
	xhash.Bool(h, wr.Jet)
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Use the weapon's ResolvedRateOfFire method first if you want the resolved values.
func (wr RoF) String() string {
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

// Tooltip returns a tooltip for the data, if any. Use the weapon's ResolvedRateOfFire method first if you want the
// tooltip to be based on the resolved values.
func (wr RoF) Tooltip() string {
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
func (wr *RoF) Validate() {
	if wr.Jet {
		wr.Mode1 = RoFMode{}
		wr.Mode2 = RoFMode{}
		return
	}
	wr.Mode1.Validate()
	wr.Mode2.Validate()
}
