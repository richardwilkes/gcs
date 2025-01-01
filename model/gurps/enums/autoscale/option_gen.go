// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package autoscale

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	No Option = iota
	FitWidth
	FitPage
)

// LastOption is the last valid value.
const LastOption Option = FitPage

// Options holds all possible values.
var Options = []Option{
	No,
	FitWidth,
	FitPage,
}

// Option holds the possible auto-scaling options.
type Option byte

// EnsureValid ensures this is of a known value.
func (enum Option) EnsureValid() Option {
	if enum <= FitPage {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Option) Key() string {
	switch enum {
	case No:
		return "no"
	case FitWidth:
		return "fit_width"
	case FitPage:
		return "fit_page"
	default:
		return Option(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Option) String() string {
	switch enum {
	case No:
		return i18n.Text("No Auto-Scaling")
	case FitWidth:
		return i18n.Text("Fit Width")
	case FitPage:
		return i18n.Text("Fit Page")
	default:
		return Option(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Option) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Option) UnmarshalText(text []byte) error {
	*enum = ExtractOption(string(text))
	return nil
}

// ExtractOption extracts the value from a string.
func ExtractOption(str string) Option {
	for _, enum := range Options {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
