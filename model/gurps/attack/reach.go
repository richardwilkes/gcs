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

// Reach holds the reach data for a weapon.
type Reach struct {
	Min                 fxp.Int
	Max                 fxp.Int
	CloseCombat         bool
	ChangeRequiresReady bool
}

// ParseReach parses a string into a Reach.
func ParseReach(s string) Reach {
	var wr Reach
	s = strings.ReplaceAll(s, " ", "")
	if s != "" {
		s = strings.ToLower(s)
		if !strings.Contains(s, "spec") {
			wr.CloseCombat = strings.Contains(s, "c")
			wr.ChangeRequiresReady = strings.Contains(s, "*")
			// "C-5" runs from close combat out to 5, so its number is only the maximum and Validate fills in a minimum
			// of 1. Elsewhere ("C,1", "C,2-3") the marker stands apart from the reach, so the first number is the
			// minimum. The markers are removed so they aren't read as a reach component of 0.
			closeCombatRange := strings.Contains(s, "c-")
			s = strings.NewReplacer("*", "", "c", "").Replace(s)
			takeMin := !closeCombatRange
			for one := range strings.SplitSeq(strings.ReplaceAll(s, "-", ","), ",") {
				reach, remainder := fxp.Extract(one)
				if remainder == one {
					continue // Holds no number, e.g. an empty component left behind by a separator.
				}
				if takeMin {
					wr.Min = reach
					takeMin = false
				}
				wr.Max = max(wr.Max, reach)
			}
			wr.Validate()
		}
	}
	return wr
}

// MarshalJSONTo implements json.MarshalerTo.
func (wr Reach) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, wr.String())
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (wr *Reach) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return jio.UnmarshalStringFromInfallible(dec, wr, ParseReach)
}

// IsZero reports whether json's omitzero option should omit this value.
func (wr Reach) IsZero() bool {
	return wr == Reach{}
}

// Hash writes this object's contents into the hasher.
func (wr Reach) Hash(h hash.Hash) {
	xhash.Num64(h, wr.Min)
	xhash.Num64(h, wr.Max)
	xhash.Bool(h, wr.CloseCombat)
	xhash.Bool(h, wr.ChangeRequiresReady)
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Use the weapon's ResolvedReach method first if you want the resolved values.
func (wr Reach) String() string {
	var buffer strings.Builder
	if wr.CloseCombat {
		buffer.WriteByte('C')
	}
	if wr.Min != 0 || wr.Max != 0 {
		if buffer.Len() != 0 {
			buffer.WriteByte(',')
		}
		buffer.WriteString(wr.Min.String())
		if wr.Min != wr.Max {
			buffer.WriteByte('-')
			buffer.WriteString(wr.Max.String())
		}
	}
	if wr.ChangeRequiresReady {
		buffer.WriteByte('*')
	}
	return buffer.String()
}

// Tooltip returns a tooltip for the data, if any. Use the weapon's ResolvedReach method first if you want the tooltip
// to be based on the resolved values.
func (wr Reach) Tooltip() string {
	if wr.ChangeRequiresReady {
		return i18n.Text("*: Changing reach requires a Ready maneuver.")
	}
	return ""
}

// Validate ensures that the data is valid.
func (wr *Reach) Validate() {
	wr.Min = wr.Min.Max(0)
	wr.Max = wr.Max.Max(0)
	if wr.Min == 0 && wr.Max != 0 {
		wr.Min = fxp.One
	} else if wr.Min != 0 && wr.Max == 0 {
		wr.Max = wr.Min
	}
	wr.Max = max(wr.Max, wr.Min)
}
