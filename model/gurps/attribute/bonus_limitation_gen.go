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
	None BonusLimitation = iota
	StrikingOnly
	LiftingOnly
	ThrowingOnly
	LastBonusLimitation = ThrowingOnly
)

var (
	// AllBonusLimitation holds all possible values.
	AllBonusLimitation = []BonusLimitation{
		None,
		StrikingOnly,
		LiftingOnly,
		ThrowingOnly,
	}
	bonusLimitationData = []struct {
		key    string
		string string
	}{
		{
			key:    "none",
			string: "",
		},
		{
			key:    "striking_only",
			string: i18n.Text("for striking only"),
		},
		{
			key:    "lifting_only",
			string: i18n.Text("for lifting only"),
		},
		{
			key:    "throwing_only",
			string: i18n.Text("for throwing only"),
		},
	}
)

// BonusLimitation holds a limitation for an AttributeBonus.
type BonusLimitation byte

// EnsureValid ensures this is of a known value.
func (enum BonusLimitation) EnsureValid() BonusLimitation {
	if enum <= LastBonusLimitation {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum BonusLimitation) Key() string {
	return bonusLimitationData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum BonusLimitation) String() string {
	return bonusLimitationData[enum.EnsureValid()].string
}

// ExtractBonusLimitation extracts the value from a string.
func ExtractBonusLimitation(str string) BonusLimitation {
	for i, one := range bonusLimitationData {
		if strings.EqualFold(one.key, str) {
			return BonusLimitation(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum BonusLimitation) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *BonusLimitation) UnmarshalText(text []byte) error {
	*enum = ExtractBonusLimitation(string(text))
	return nil
}
