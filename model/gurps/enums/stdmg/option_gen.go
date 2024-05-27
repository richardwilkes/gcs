// Code generated from "enum.go.tmpl" - DO NOT EDIT.

/*
 * Copyright ©1998-2024 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package stdmg

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	None Option = iota
	Thrust
	LeveledThrust
	Swing
	LeveledSwing
)

// LastOption is the last valid value.
const LastOption Option = LeveledSwing

// Options holds all possible values.
var Options = []Option{
	None,
	Thrust,
	LeveledThrust,
	Swing,
	LeveledSwing,
}

// Option holds the type of strength dice to add to damage.
type Option byte

// EnsureValid ensures this is of a known value.
func (enum Option) EnsureValid() Option {
	if enum <= LeveledSwing {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Option) Key() string {
	switch enum {
	case None:
		return "none"
	case Thrust:
		return "thr"
	case LeveledThrust:
		return "thr_leveled"
	case Swing:
		return "sw"
	case LeveledSwing:
		return "sw_leveled"
	default:
		return Option(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Option) String() string {
	switch enum {
	case None:
		return i18n.Text("None")
	case Thrust:
		return "thr"
	case LeveledThrust:
		return i18n.Text("thr (leveled)")
	case Swing:
		return "sw"
	case LeveledSwing:
		return i18n.Text("sw (leveled)")
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
