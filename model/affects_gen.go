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

package model

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	TotalAffects Affects = iota
	BaseOnlyAffects
	LevelsOnlyAffects
	LastAffects = LevelsOnlyAffects
)

// AllAffects holds all possible values.
var AllAffects = []Affects{
	TotalAffects,
	BaseOnlyAffects,
	LevelsOnlyAffects,
}

// Affects describes how a TraitModifier affects the point cost.
type Affects byte

// EnsureValid ensures this is of a known value.
func (enum Affects) EnsureValid() Affects {
	if enum <= LastAffects {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Affects) Key() string {
	switch enum {
	case TotalAffects:
		return "total"
	case BaseOnlyAffects:
		return "base_only"
	case LevelsOnlyAffects:
		return "levels_only"
	default:
		return Affects(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Affects) String() string {
	switch enum {
	case TotalAffects:
		return i18n.Text("to cost")
	case BaseOnlyAffects:
		return i18n.Text("to base cost only")
	case LevelsOnlyAffects:
		return i18n.Text("to leveled cost only")
	default:
		return Affects(0).String()
	}
}

// AltString returns the alternate string.
func (enum Affects) AltString() string {
	switch enum {
	case TotalAffects:
		return ""
	case BaseOnlyAffects:
		return i18n.Text("(base only)")
	case LevelsOnlyAffects:
		return i18n.Text("(levels only)")
	default:
		return Affects(0).AltString()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Affects) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Affects) UnmarshalText(text []byte) error {
	*enum = ExtractAffects(string(text))
	return nil
}

// ExtractAffects extracts the value from a string.
func ExtractAffects(str string) Affects {
	for _, enum := range AllAffects {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
