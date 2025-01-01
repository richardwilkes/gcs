// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package paper

import (
	"strconv"
	"strings"

	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
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
	text = strings.TrimLeft(strings.TrimSpace(text), "+")
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

// MarshalJSON implements json.Marshaler.
func (l Length) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (l *Length) UnmarshalJSON(in []byte) error {
	var s string
	if err := json.Unmarshal(in, &s); err != nil {
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
	if l.Length < 0 {
		l.Length = 0
	}
}
