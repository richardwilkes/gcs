// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package emcost

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	Original Type = iota
	Base
	FinalBase
	Final
)

// LastType is the last valid value.
const LastType Type = Final

// Types holds all possible values.
var Types = []Type{
	Original,
	Base,
	FinalBase,
	Final,
}

// Type describes how an Equipment Modifier's cost is applied.
type Type byte

// EnsureValid ensures this is of a known value.
func (enum Type) EnsureValid() Type {
	if enum <= Final {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Type) Key() string {
	switch enum {
	case Original:
		return "to_original_cost"
	case Base:
		return "to_base_cost"
	case FinalBase:
		return "to_final_base_cost"
	case Final:
		return "to_final_cost"
	default:
		return Type(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Type) String() string {
	switch enum {
	case Original:
		return i18n.Text("to original cost")
	case Base:
		return i18n.Text("to base cost")
	case FinalBase:
		return i18n.Text("to final base cost")
	case Final:
		return i18n.Text("to final cost")
	default:
		return Type(0).String()
	}
}

// AltString returns the alternate string.
func (enum Type) AltString() string {
	switch enum {
	case Original:
		return i18n.Text("\"+5\", \"-5\", \"+10%\", \"-10%\", \"x3.2\"")
	case Base:
		return i18n.Text("\"x2\", \"+2 CF\", \"-0.2 CF\"")
	case FinalBase:
		return i18n.Text("\"+5\", \"-5\", \"+10%\", \"-10%\", \"x3.2\"")
	case Final:
		return i18n.Text("\"+5\", \"-5\", \"+10%\", \"-10%\", \"x3.2\"")
	default:
		return Type(0).AltString()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Type) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Type) UnmarshalText(text []byte) error {
	*enum = ExtractType(string(text))
	return nil
}

// ExtractType extracts the value from a string.
func ExtractType(str string) Type {
	for _, enum := range Types {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
