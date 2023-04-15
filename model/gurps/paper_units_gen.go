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

package gurps

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

// AllPaperUnits holds all possible values.
var AllPaperUnits = []PaperUnits{
	InchPaperUnits,
	CentimeterPaperUnits,
	MillimeterPaperUnits,
}

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
	switch enum {
	case InchPaperUnits:
		return "in"
	case CentimeterPaperUnits:
		return "cm"
	case MillimeterPaperUnits:
		return "mm"
	default:
		return PaperUnits(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum PaperUnits) String() string {
	switch enum {
	case InchPaperUnits:
		return "in"
	case CentimeterPaperUnits:
		return "cm"
	case MillimeterPaperUnits:
		return "mm"
	default:
		return PaperUnits(0).String()
	}
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

// ExtractPaperUnits extracts the value from a string.
func ExtractPaperUnits(str string) PaperUnits {
	for _, enum := range AllPaperUnits {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
