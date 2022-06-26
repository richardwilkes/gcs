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

package attribute

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	Unknown ThresholdOp = iota
	HalveMove
	HalveDodge
	HalveST
	LastThresholdOp = HalveST
)

var (
	// AllThresholdOp holds all possible values.
	AllThresholdOp = []ThresholdOp{
		Unknown,
		HalveMove,
		HalveDodge,
		HalveST,
	}
	thresholdOpData = []struct {
		key    string
		string string
		alt    string
	}{
		{
			key:    "unknown",
			string: i18n.Text("Unknown"),
			alt:    i18n.Text("Unknown"),
		},
		{
			key:    "halve_move",
			string: i18n.Text("Halve Move"),
			alt:    i18n.Text("Halve Move (round up)"),
		},
		{
			key:    "halve_dodge",
			string: i18n.Text("Halve Dodge"),
			alt:    i18n.Text("Halve Dodge (round up)"),
		},
		{
			key:    "halve_st",
			string: i18n.Text("Halve Strength"),
			alt:    i18n.Text("Halve Strength (round up; does not affect HP and damage)"),
		},
	}
)

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
	return thresholdOpData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum ThresholdOp) String() string {
	return thresholdOpData[enum.EnsureValid()].string
}

// AltString returns the alternate string.
func (enum ThresholdOp) AltString() string {
	return thresholdOpData[enum.EnsureValid()].alt
}

// ExtractThresholdOp extracts the value from a string.
func ExtractThresholdOp(str string) ThresholdOp {
	for i, one := range thresholdOpData {
		if strings.EqualFold(one.key, str) {
			return ThresholdOp(i)
		}
	}
	return 0
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
