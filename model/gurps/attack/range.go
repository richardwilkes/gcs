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
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// Range holds the range data for a weapon.
type Range struct {
	HalfDamage    fxp.Int
	Min           fxp.Int
	Max           fxp.Int
	MusclePowered bool
	InMiles       bool
}

// ParseRange parses a string into a Range.
func ParseRange(s string) Range {
	var wr Range
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
// The unresolved display string is persisted so that the "x" prefixes (e.g. "x10/x100") survive: ParseRange
// detects them on load to restore the MusclePowered flag.
func (wr Range) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, wr.String(false))
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (wr *Range) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return jio.UnmarshalStringFromInfallible(dec, wr, ParseRange)
}

// IsZero reports whether json's omitzero option should omit this value.
func (wr Range) IsZero() bool {
	return wr == Range{}
}

// Hash writes this object's contents into the hasher.
func (wr Range) Hash(h hash.Hash) {
	xhash.Num64(h, wr.HalfDamage)
	xhash.Num64(h, wr.Min)
	xhash.Num64(h, wr.Max)
	xhash.Bool(h, wr.MusclePowered)
	xhash.Bool(h, wr.InMiles)
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Use the weapon's ResolvedRange method first if you want the resolved values.
//
// When MusclePowered is set, passing false for musclePowerIsResolved prefixes the range values with "x" (the GURPS
// notation for "multiply by ST"); pass true once the ranges have already been multiplied by ST so the "x" is omitted.
func (wr Range) String(musclePowerIsResolved bool) string {
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
func (wr *Range) Validate() {
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
