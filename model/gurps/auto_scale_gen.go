// Code generated from "enum.go.tmpl" - DO NOT EDIT.

/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
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
	NoAutoScale AutoScale = iota
	FitWidthAutoScale
	FitPageAutoScale
	LastAutoScale = FitPageAutoScale
)

// AllAutoScale holds all possible values.
var AllAutoScale = []AutoScale{
	NoAutoScale,
	FitWidthAutoScale,
	FitPageAutoScale,
}

// AutoScale holds the possible auto-scaling options.
type AutoScale byte

// EnsureValid ensures this is of a known value.
func (enum AutoScale) EnsureValid() AutoScale {
	if enum <= LastAutoScale {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum AutoScale) Key() string {
	switch enum {
	case NoAutoScale:
		return "no"
	case FitWidthAutoScale:
		return "fit_width"
	case FitPageAutoScale:
		return "fit_page"
	default:
		return AutoScale(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum AutoScale) String() string {
	switch enum {
	case NoAutoScale:
		return i18n.Text("No Auto-Scaling")
	case FitWidthAutoScale:
		return i18n.Text("Fit Width")
	case FitPageAutoScale:
		return i18n.Text("Fit Page")
	default:
		return AutoScale(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum AutoScale) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *AutoScale) UnmarshalText(text []byte) error {
	*enum = ExtractAutoScale(string(text))
	return nil
}

// ExtractAutoScale extracts the value from a string.
func ExtractAutoScale(str string) AutoScale {
	for _, enum := range AllAutoScale {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
