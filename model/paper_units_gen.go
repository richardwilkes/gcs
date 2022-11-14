// Code generated from "enum.go.tmpl" - DO NOT EDIT.

/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package model

import (
	"strings"
)

// Possible values.
const (
	InchPaperUnits PaperUnits = iota
	CentimeterPaperUnits
	MillimeterPaperUnits
	LastPaperUnits = MillimeterPaperUnits
)

var (
	// AllPaperUnits holds all possible values.
	AllPaperUnits = []PaperUnits{
		InchPaperUnits,
		CentimeterPaperUnits,
		MillimeterPaperUnits,
	}
	paperUnitsData = []struct {
		key    string
		string string
	}{
		{
			key:    "in",
			string: "in",
		},
		{
			key:    "cm",
			string: "cm",
		},
		{
			key:    "mm",
			string: "mm",
		},
	}
)

// PaperUnits holds the real-world length unit type.
type PaperUnits byte

// EnsureValid ensures this is of a known value.
func (enum PaperUnits) EnsureValid() PaperUnits {
	if enum <= LastPaperUnits {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum PaperUnits) Key() string {
	return paperUnitsData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum PaperUnits) String() string {
	return paperUnitsData[enum.EnsureValid()].string
}

// ExtractPaperUnits extracts the value from a string.
func ExtractPaperUnits(str string) PaperUnits {
	for i, one := range paperUnitsData {
		if strings.EqualFold(one.key, str) {
			return PaperUnits(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum PaperUnits) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *PaperUnits) UnmarshalText(text []byte) error {
	*enum = ExtractPaperUnits(string(text))
	return nil
}
