// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package levelcap

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	Skills Target = iota
	Spells
	SkillsAndSpells
)

// DefaultTarget is the default value.
const DefaultTarget Target = Skills

// FirstTarget is the first valid value.
const FirstTarget Target = Skills

// LastTarget is the last valid value.
const LastTarget Target = SkillsAndSpells

// Targets holds all possible values.
var Targets = []Target{
	Skills,
	Spells,
	SkillsAndSpells,
}

// Target holds the kind of item a SkillLevelCap applies to.
type Target byte

// EnsureValid returns this value if it is valid and DefaultTarget otherwise.
func (enum Target) EnsureValid() Target {
	if enum >= FirstTarget && enum <= LastTarget {
		return enum
	}
	return DefaultTarget
}

// Key returns the key used in serialization.
func (enum Target) Key() string {
	switch enum {
	case Skills:
		return "skills"
	case Spells:
		return "spells"
	case SkillsAndSpells:
		return "skills_and_spells"
	default:
		return DefaultTarget.Key()
	}
}

// String implements fmt.Stringer.
func (enum Target) String() string {
	switch enum {
	case Skills:
		return i18n.Text(`Skills`)
	case Spells:
		return i18n.Text(`Spells`)
	case SkillsAndSpells:
		return i18n.Text(`Skills and spells`)
	default:
		return DefaultTarget.String()
	}
}

// MarshalText implements encoding.TextMarshaler.
func (enum Target) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler. A key that matches no value is an error.
func (enum *Target) UnmarshalText(text []byte) error {
	v, known := ExtractKnownTarget(string(text))
	if !known {
		return errs.Newf("unknown levelcap target %q", string(text))
	}
	*enum = v
	return nil
}

// ExtractTarget returns the value whose key matches str, ignoring case, or DefaultTarget if none does.
func ExtractTarget(str string) Target {
	for _, enum := range Targets {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultTarget
}

// ExtractKnownTarget is like ExtractTarget, but also reports whether str was recognized, so a caller can tell an
// unknown key from the default.
func ExtractKnownTarget(str string) (value Target, known bool) {
	for _, enum := range Targets {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultTarget, false
}
