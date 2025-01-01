// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package paper

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	Portrait Orientation = iota
	Landscape
)

// LastOrientation is the last valid value.
const LastOrientation Orientation = Landscape

// Orientations holds all possible values.
var Orientations = []Orientation{
	Portrait,
	Landscape,
}

// Orientation holds the orientation of the page.
type Orientation byte

// EnsureValid ensures this is of a known value.
func (enum Orientation) EnsureValid() Orientation {
	if enum <= Landscape {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Orientation) Key() string {
	switch enum {
	case Portrait:
		return "portrait"
	case Landscape:
		return "landscape"
	default:
		return Orientation(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Orientation) String() string {
	switch enum {
	case Portrait:
		return i18n.Text("Portrait")
	case Landscape:
		return i18n.Text("Landscape")
	default:
		return Orientation(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Orientation) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Orientation) UnmarshalText(text []byte) error {
	*enum = ExtractOrientation(string(text))
	return nil
}

// ExtractOrientation extracts the value from a string.
func ExtractOrientation(str string) Orientation {
	for _, enum := range Orientations {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
