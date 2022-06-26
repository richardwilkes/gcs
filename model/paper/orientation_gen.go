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

package paper

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	Portrait Orientation = iota
	Landscape
	LastOrientation = Landscape
)

var (
	// AllOrientation holds all possible values.
	AllOrientation = []Orientation{
		Portrait,
		Landscape,
	}
	orientationData = []struct {
		key    string
		string string
	}{
		{
			key:    "portrait",
			string: i18n.Text("Portrait"),
		},
		{
			key:    "landscape",
			string: i18n.Text("Landscape"),
		},
	}
)

// Orientation holds the orientation of the page.
type Orientation byte

// EnsureValid ensures this is of a known value.
func (enum Orientation) EnsureValid() Orientation {
	if enum <= LastOrientation {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Orientation) Key() string {
	return orientationData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum Orientation) String() string {
	return orientationData[enum.EnsureValid()].string
}

// ExtractOrientation extracts the value from a string.
func ExtractOrientation(str string) Orientation {
	for i, one := range orientationData {
		if strings.EqualFold(one.key, str) {
			return Orientation(i)
		}
	}
	return 0
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
