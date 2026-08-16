// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
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

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	Portrait Orientation = iota
	Landscape
)

// DefaultOrientation is the default value.
const DefaultOrientation Orientation = Portrait

// FirstOrientation is the first valid value.
const FirstOrientation Orientation = Portrait

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
	if enum >= FirstOrientation && enum <= LastOrientation {
		return enum
	}
	return DefaultOrientation
}

// Key returns the key used in serialization.
func (enum Orientation) Key() string {
	switch enum {
	case Portrait:
		return "portrait"
	case Landscape:
		return "landscape"
	default:
		return DefaultOrientation.Key()
	}
}

// String implements fmt.Stringer.
func (enum Orientation) String() string {
	switch enum {
	case Portrait:
		return i18n.Text(`Portrait`)
	case Landscape:
		return i18n.Text(`Landscape`)
	default:
		return DefaultOrientation.String()
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
	return DefaultOrientation
}

// ExtractKnownOrientation extracts the value from a string, reporting whether the string was actually recognized.
//
// Unlike ExtractOrientation, which quietly maps anything it doesn't recognize onto the first value, this permits a
// caller that is dispatching on the type to detect unknown types.
func ExtractKnownOrientation(str string) (value Orientation, known bool) {
	for _, enum := range Orientations {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultOrientation, false
}
