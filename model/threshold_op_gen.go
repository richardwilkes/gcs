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

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	UnknownThresholdOp ThresholdOp = iota
	HalveMoveThresholdOp
	HalveDodgeThresholdOp
	HalveSTThresholdOp
	LastThresholdOp = HalveSTThresholdOp
)

// AllThresholdOp holds all possible values.
var AllThresholdOp = []ThresholdOp{
	UnknownThresholdOp,
	HalveMoveThresholdOp,
	HalveDodgeThresholdOp,
	HalveSTThresholdOp,
}

// ThresholdOp holds an operation to apply when a pool threshold is hit.
type ThresholdOp byte

// EnsureValid ensures this is of a known value.
func (enum ThresholdOp) EnsureValid() ThresholdOp {
	if enum <= LastThresholdOp {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum ThresholdOp) Key() string {
	switch enum {
	case UnknownThresholdOp:
		return "unknown"
	case HalveMoveThresholdOp:
		return "halve_move"
	case HalveDodgeThresholdOp:
		return "halve_dodge"
	case HalveSTThresholdOp:
		return "halve_st"
	default:
		return ThresholdOp(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum ThresholdOp) String() string {
	switch enum {
	case UnknownThresholdOp:
		return i18n.Text("Unknown")
	case HalveMoveThresholdOp:
		return i18n.Text("Halve Move")
	case HalveDodgeThresholdOp:
		return i18n.Text("Halve Dodge")
	case HalveSTThresholdOp:
		return i18n.Text("Halve Strength")
	default:
		return ThresholdOp(0).String()
	}
}

// AltString returns the alternate string.
func (enum ThresholdOp) AltString() string {
	switch enum {
	case UnknownThresholdOp:
		return i18n.Text("Unknown")
	case HalveMoveThresholdOp:
		return i18n.Text("Halve Move (round up)")
	case HalveDodgeThresholdOp:
		return i18n.Text("Halve Dodge (round up)")
	case HalveSTThresholdOp:
		return i18n.Text("Halve Strength (round up; does not affect HP and damage)")
	default:
		return ThresholdOp(0).AltString()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum ThresholdOp) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *ThresholdOp) UnmarshalText(text []byte) error {
	*enum = ExtractThresholdOp(string(text))
	return nil
}

// ExtractThresholdOp extracts the value from a string.
func ExtractThresholdOp(str string) ThresholdOp {
	for _, enum := range AllThresholdOp {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
