// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package attribute

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	Automatic Placement = iota
	Primary
	Secondary
	Hidden
)

// LastPlacement is the last valid value.
const LastPlacement Placement = Hidden

// Placements holds all possible values.
var Placements = []Placement{
	Automatic,
	Primary,
	Secondary,
	Hidden,
}

// Placement determines the placement of the attribute on the sheet.
type Placement byte

// EnsureValid ensures this is of a known value.
func (enum Placement) EnsureValid() Placement {
	if enum <= Hidden {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Placement) Key() string {
	switch enum {
	case Automatic:
		return "automatic"
	case Primary:
		return "primary"
	case Secondary:
		return "secondary"
	case Hidden:
		return "hidden"
	default:
		return Placement(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Placement) String() string {
	switch enum {
	case Automatic:
		return i18n.Text("Automatic")
	case Primary:
		return i18n.Text("Primary")
	case Secondary:
		return i18n.Text("Secondary")
	case Hidden:
		return i18n.Text("Hidden")
	default:
		return Placement(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Placement) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Placement) UnmarshalText(text []byte) error {
	*enum = ExtractPlacement(string(text))
	return nil
}

// ExtractPlacement extracts the value from a string.
func ExtractPlacement(str string) Placement {
	for _, enum := range Placements {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
