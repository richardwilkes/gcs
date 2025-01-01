// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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

	"github.com/richardwilkes/toolbox/i18n"
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

// EnsureValid ensures this is of a known value.
func (enum Builtin) EnsureValid() Builtin {
	if enum <= UnweightedAmericanLast {
		return enum
	}
	return 0
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
		return Builtin(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Builtin) String() string {
	switch enum {
	case None:
		return i18n.Text("None")
	case AmericanMale:
		return i18n.Text("American Male")
	case AmericanFemale:
		return i18n.Text("American Female")
	case AmericanLast:
		return i18n.Text("American Last")
	case UnweightedAmericanMale:
		return i18n.Text("Unweighted American Male")
	case UnweightedAmericanFemale:
		return i18n.Text("Unweighted American Female")
	case UnweightedAmericanLast:
		return i18n.Text("Unweighted American Last")
	default:
		return Builtin(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Builtin) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Builtin) UnmarshalText(text []byte) error {
	*enum = ExtractBuiltin(string(text))
	return nil
}

// ExtractBuiltin extracts the value from a string.
func ExtractBuiltin(str string) Builtin {
	for _, enum := range Builtins {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
