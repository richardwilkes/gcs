// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package paper

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"math"
	"strconv"
	"strings"

	"github.com/richardwilkes/toolbox/v2/errs"
)

// Length contains a real-world length value with an attached units.
type Length struct {
	Length float64
	Units  Unit
}

// LengthFromString creates a new Length. May have any of the known unit suffixes or no notation at all, in which case
// inch is used.
func LengthFromString(text string) Length {
	length, err := ParseLengthFromString(text)
	if err != nil {
		length.Length = 0
	}
	return length
}

// ParseLengthFromString parses a Length from the text. May have any of the known unit suffixes or no notation at all,
// in which case inch is used.
func ParseLengthFromString(text string) (length Length, err error) {
	text = strings.ToLower(strings.TrimLeft(strings.TrimSpace(text), "+"))
	for _, unit := range Units {
		if strings.HasSuffix(text, unit.Key()) {
			length.Units = unit
			text = strings.TrimSpace(strings.TrimSuffix(text, unit.Key()))
			break
		}
	}
	if length.Length, err = strconv.ParseFloat(text, 64); err != nil {
		return length, errs.NewWithCause("invalid value", err)
	}
	// strconv.ParseFloat accepts "inf", "infinity" and "nan" (and any leading '+' has already been trimmed), none of
	// which are < 0, so they must be rejected explicitly. Letting one through poisons the page layout arithmetic with
	// a non-finite number that then survives a save/load round trip.
	if math.IsInf(length.Length, 0) || math.IsNaN(length.Length) {
		length.Length = 0
		return length, errs.New("value must be a finite number")
	}
	if length.Length < 0 {
		return length, errs.New("value must be zero or greater")
	}
	return length, nil
}

func (l Length) String() string {
	return strconv.FormatFloat(l.Length, 'f', -1, 64) + " " + l.Units.Key()
}

// CSSString returns a CSS-compatible version of the value.
func (l Length) CSSString() string {
	return strconv.FormatFloat(l.Length, 'f', -1, 64) + l.Units.Key()
}

// Pixels returns the number of 72-pixels-per-inch pixels this represents.
func (l Length) Pixels() float32 {
	return l.Units.ToPixels(l.Length)
}

// MarshalJSONTo implements json.MarshalerTo.
func (l Length) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, l.String())
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (l *Length) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var s string
	if err := json.UnmarshalDecode(dec, &s); err != nil {
		return err
	}
	var err error
	if *l, err = ParseLengthFromString(s); err != nil {
		return err
	}
	return nil
}

// EnsureValidity checks the current settings for validity and if they aren't valid, makes them so.
func (l *Length) EnsureValidity() {
	l.Units = l.Units.EnsureValid()
	// Neither +Inf nor NaN is < 0, so they have to be checked for separately.
	if l.Length < 0 || math.IsInf(l.Length, 0) || math.IsNaN(l.Length) {
		l.Length = 0
	}
}
