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

package gurps

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	Portrait PaperOrientation = iota
	Landscape
	LastPaperOrientation = Landscape
)

// AllPaperOrientation holds all possible values.
var AllPaperOrientation = []PaperOrientation{
	Portrait,
	Landscape,
}

// PaperOrientation holds the orientation of the page.
type PaperOrientation byte

// EnsureValid ensures this is of a known value.
func (enum PaperOrientation) EnsureValid() PaperOrientation {
	if enum <= LastPaperOrientation {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum PaperOrientation) Key() string {
	switch enum {
	case Portrait:
		return "portrait"
	case Landscape:
		return "landscape"
	default:
		return PaperOrientation(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum PaperOrientation) String() string {
	switch enum {
	case Portrait:
		return i18n.Text("Portrait")
	case Landscape:
		return i18n.Text("Landscape")
	default:
		return PaperOrientation(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum PaperOrientation) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *PaperOrientation) UnmarshalText(text []byte) error {
	*enum = ExtractPaperOrientation(string(text))
	return nil
}

// ExtractPaperOrientation extracts the value from a string.
func ExtractPaperOrientation(str string) PaperOrientation {
	for _, enum := range AllPaperOrientation {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
