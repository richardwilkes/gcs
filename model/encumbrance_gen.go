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
	NoEncumbrance Encumbrance = iota
	LightEncumbrance
	MediumEncumbrance
	HeavyEncumbrance
	ExtraHeavyEncumbrance
	LastEncumbrance = ExtraHeavyEncumbrance
)

// AllEncumbrance holds all possible values.
var AllEncumbrance = []Encumbrance{
	NoEncumbrance,
	LightEncumbrance,
	MediumEncumbrance,
	HeavyEncumbrance,
	ExtraHeavyEncumbrance,
}

// Encumbrance holds the encumbrance level.
type Encumbrance byte

// EnsureValid ensures this is of a known value.
func (enum Encumbrance) EnsureValid() Encumbrance {
	if enum <= LastEncumbrance {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Encumbrance) Key() string {
	switch enum {
	case NoEncumbrance:
		return "none"
	case LightEncumbrance:
		return "light"
	case MediumEncumbrance:
		return "medium"
	case HeavyEncumbrance:
		return "heavy"
	case ExtraHeavyEncumbrance:
		return "extra_heavy"
	default:
		return Encumbrance(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Encumbrance) String() string {
	switch enum {
	case NoEncumbrance:
		return i18n.Text("None")
	case LightEncumbrance:
		return i18n.Text("Light")
	case MediumEncumbrance:
		return i18n.Text("Medium")
	case HeavyEncumbrance:
		return i18n.Text("Heavy")
	case ExtraHeavyEncumbrance:
		return i18n.Text("X-Heavy")
	default:
		return Encumbrance(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Encumbrance) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Encumbrance) UnmarshalText(text []byte) error {
	*enum = ExtractEncumbrance(string(text))
	return nil
}

// ExtractEncumbrance extracts the value from a string.
func ExtractEncumbrance(str string) Encumbrance {
	for _, enum := range AllEncumbrance {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
