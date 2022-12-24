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
	NoneBonusLimitation BonusLimitation = iota
	StrikingOnlyBonusLimitation
	LiftingOnlyBonusLimitation
	ThrowingOnlyBonusLimitation
	LastBonusLimitation = ThrowingOnlyBonusLimitation
)

// AllBonusLimitation holds all possible values.
var AllBonusLimitation = []BonusLimitation{
	NoneBonusLimitation,
	StrikingOnlyBonusLimitation,
	LiftingOnlyBonusLimitation,
	ThrowingOnlyBonusLimitation,
}

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
	switch enum {
	case NoneBonusLimitation:
		return "none"
	case StrikingOnlyBonusLimitation:
		return "striking_only"
	case LiftingOnlyBonusLimitation:
		return "lifting_only"
	case ThrowingOnlyBonusLimitation:
		return "throwing_only"
	default:
		return BonusLimitation(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum BonusLimitation) String() string {
	switch enum {
	case NoneBonusLimitation:
		return ""
	case StrikingOnlyBonusLimitation:
		return i18n.Text("for striking only")
	case LiftingOnlyBonusLimitation:
		return i18n.Text("for lifting only")
	case ThrowingOnlyBonusLimitation:
		return i18n.Text("for throwing only")
	default:
		return BonusLimitation(0).String()
	}
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

// ExtractBonusLimitation extracts the value from a string.
func ExtractBonusLimitation(str string) BonusLimitation {
	for _, enum := range AllBonusLimitation {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
