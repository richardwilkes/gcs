// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package namegen

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	None Builtin = iota
	AmericanMale
	AmericanFemale
	AmericanLast
	UnweightedAmericanMale
	UnweightedAmericanFemale
	UnweightedAmericanLast
)

// DefaultBuiltin is the default value.
const DefaultBuiltin Builtin = None

// FirstBuiltin is the first valid value.
const FirstBuiltin Builtin = None

// LastBuiltin is the last valid value.
const LastBuiltin Builtin = UnweightedAmericanLast

// Builtins holds all possible values.
var Builtins = []Builtin{
	None,
	AmericanMale,
	AmericanFemale,
	AmericanLast,
	UnweightedAmericanMale,
	UnweightedAmericanFemale,
	UnweightedAmericanLast,
}

// Builtin holds a built-in name data type.
type Builtin byte

// EnsureValid returns this value if it is valid and DefaultBuiltin otherwise.
func (enum Builtin) EnsureValid() Builtin {
	if enum >= FirstBuiltin && enum <= LastBuiltin {
		return enum
	}
	return DefaultBuiltin
}

// Key returns the key used in serialization.
func (enum Builtin) Key() string {
	switch enum {
	case None:
		return "none"
	case AmericanMale:
		return "american_male"
	case AmericanFemale:
		return "american_female"
	case AmericanLast:
		return "american_last"
	case UnweightedAmericanMale:
		return "unweighted_american_male"
	case UnweightedAmericanFemale:
		return "unweighted_american_female"
	case UnweightedAmericanLast:
		return "unweighted_american_last"
	default:
		return DefaultBuiltin.Key()
	}
}

// String implements fmt.Stringer.
func (enum Builtin) String() string {
	switch enum {
	case None:
		return i18n.Text(`None`)
	case AmericanMale:
		return i18n.Text(`American Male`)
	case AmericanFemale:
		return i18n.Text(`American Female`)
	case AmericanLast:
		return i18n.Text(`American Last`)
	case UnweightedAmericanMale:
		return i18n.Text(`Unweighted American Male`)
	case UnweightedAmericanFemale:
		return i18n.Text(`Unweighted American Female`)
	case UnweightedAmericanLast:
		return i18n.Text(`Unweighted American Last`)
	default:
		return DefaultBuiltin.String()
	}
}

// MarshalText implements encoding.TextMarshaler.
func (enum Builtin) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (enum *Builtin) UnmarshalText(text []byte) error {
	*enum = ExtractBuiltin(string(text))
	return nil
}

// ExtractBuiltin returns the value whose key matches str, ignoring case, or DefaultBuiltin if none does.
func ExtractBuiltin(str string) Builtin {
	for _, enum := range Builtins {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultBuiltin
}

// ExtractKnownBuiltin is like ExtractBuiltin, but also reports whether str was recognized, so a caller can tell an
// unknown key from the default.
func ExtractKnownBuiltin(str string) (value Builtin, known bool) {
	for _, enum := range Builtins {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultBuiltin, false
}
