// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package threshold

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	Unknown Op = iota
	HalveMove
	HalveDodge
	HalveST
)

// LastOp is the last valid value.
const LastOp Op = HalveST

// Ops holds all possible values.
var Ops = []Op{
	Unknown,
	HalveMove,
	HalveDodge,
	HalveST,
}

// Op holds an operation to apply when a pool threshold is hit.
type Op byte

// EnsureValid ensures this is of a known value.
func (enum Op) EnsureValid() Op {
	if enum <= HalveST {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Op) Key() string {
	switch enum {
	case Unknown:
		return "unknown"
	case HalveMove:
		return "halve_move"
	case HalveDodge:
		return "halve_dodge"
	case HalveST:
		return "halve_st"
	default:
		return Op(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Op) String() string {
	switch enum {
	case Unknown:
		return i18n.Text("Unknown")
	case HalveMove:
		return i18n.Text("Halve Move")
	case HalveDodge:
		return i18n.Text("Halve Dodge")
	case HalveST:
		return i18n.Text("Halve Strength")
	default:
		return Op(0).String()
	}
}

// AltString returns the alternate string.
func (enum Op) AltString() string {
	switch enum {
	case Unknown:
		return i18n.Text("Unknown")
	case HalveMove:
		return i18n.Text("Halve Move (round up)")
	case HalveDodge:
		return i18n.Text("Halve Dodge (round up)")
	case HalveST:
		return i18n.Text("Halve Strength (round up; does not affect HP and damage)")
	default:
		return Op(0).AltString()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Op) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Op) UnmarshalText(text []byte) error {
	*enum = ExtractOp(string(text))
	return nil
}

// ExtractOp extracts the value from a string.
func ExtractOp(str string) Op {
	for _, enum := range Ops {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
