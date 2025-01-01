// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package container

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
)

// Possible values.
const (
	Group Type = iota
	AlternativeAbilities
	Ancestry
	Attributes
	MetaTrait
)

// LastType is the last valid value.
const LastType Type = MetaTrait

// Types holds all possible values.
var Types = []Type{
	Group,
	AlternativeAbilities,
	Ancestry,
	Attributes,
	MetaTrait,
}

// Type holds the type of a trait container.
type Type byte

// EnsureValid ensures this is of a known value.
func (enum Type) EnsureValid() Type {
	if enum <= MetaTrait {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Type) Key() string {
	switch enum {
	case Group:
		return "group"
	case AlternativeAbilities:
		return "alternative_abilities"
	case Ancestry:
		return "ancestry"
	case Attributes:
		return "attributes"
	case MetaTrait:
		return "meta_trait"
	default:
		return Type(0).Key()
	}
}

func (enum Type) oldKeys() []string {
	switch enum {
	case Group:
		return nil
	case AlternativeAbilities:
		return nil
	case Ancestry:
		return []string{"race"}
	case Attributes:
		return nil
	case MetaTrait:
		return nil
	default:
		return Type(0).oldKeys()
	}
}

// String implements fmt.Stringer.
func (enum Type) String() string {
	switch enum {
	case Group:
		return i18n.Text("Group")
	case AlternativeAbilities:
		return i18n.Text("Alternative Abilities")
	case Ancestry:
		return i18n.Text("Ancestry")
	case Attributes:
		return i18n.Text("Attributes")
	case MetaTrait:
		return i18n.Text("Meta-Trait")
	default:
		return Type(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Type) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Type) UnmarshalText(text []byte) error {
	*enum = ExtractType(string(text))
	return nil
}

// ExtractType extracts the value from a string.
func ExtractType(str string) Type {
	for _, enum := range Types {
		if strings.EqualFold(enum.Key(), str) || txt.CaselessSliceContains(enum.oldKeys(), str) {
			return enum
		}
	}
	return 0
}
