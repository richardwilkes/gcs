// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package fxp

import "strings"

// This file holds the grammar shared by the modifier amounts users type for equipment modifier costs and weights, trait
// modifier costs, and maximum uses / maximum level adjustments: a plain number ("+2", "-1") is an addition, a trailing
// "%" marks a percentage ("+10%"), and a leading or trailing "x" marks a multiplier ("x2", "2x"). Classification is
// case-insensitive and ignores surrounding whitespace, and the Unicode multiplication sign is accepted in place of
// "x". The markers match the keys of the enum values that use this grammar.

const (
	// MultiplicationSign is the Unicode multiplication sign, which users may type in place of the ASCII "x" when
	// entering a multiplier.
	MultiplicationSign = "×"
	multiplierMarker   = "x"
	percentMarker      = "%"
	// multiplierLeaders holds every rune that may lead a multiplier value. Since classification is case-insensitive
	// and also accepts the Unicode multiplication sign, extraction must strip all of these forms.
	multiplierLeaders = "xX" + MultiplicationSign
)

// StripMultiplierLeaders returns s with surrounding whitespace and any leading multiplier markers ("x", "X" or "×")
// removed, leaving the numeric portion ready for extraction.
func StripMultiplierLeaders(s string) string {
	return strings.TrimLeft(strings.TrimSpace(s), multiplierLeaders)
}

// HasMultiplierPrefix reports whether s begins with a multiplier marker ("x" or "×"), ignoring surrounding whitespace
// and case.
func HasMultiplierPrefix(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.HasPrefix(s, multiplierMarker) || strings.HasPrefix(s, MultiplicationSign)
}

// HasMultiplierMarker reports whether s begins or ends with a multiplier marker ("x" or "×"), ignoring surrounding
// whitespace and case.
func HasMultiplierMarker(s string) bool {
	if HasMultiplierPrefix(s) {
		return true
	}
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.HasSuffix(s, multiplierMarker) || strings.HasSuffix(s, MultiplicationSign)
}

// HasPercentSuffix reports whether s ends with a percent sign, ignoring surrounding whitespace.
func HasPercentSuffix(s string) bool {
	return strings.HasSuffix(strings.TrimSpace(s), percentMarker)
}

// ExtractModifierValue extracts the numeric portion of a modifier amount such as "+2", "-10%" or "x3", ignoring any
// multiplier marker. When multiplier is true, a non-positive value is treated as 1, since multiplying by zero or a
// negative amount is never what was intended.
func ExtractModifierValue(s string, multiplier bool) Int {
	v, _ := Extract(StripMultiplierLeaders(s))
	if multiplier && v <= 0 {
		v = One
	}
	return v
}
