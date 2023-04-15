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
	GroupContainerType ContainerType = iota
	MetaTraitContainerType
	RaceContainerType
	AlternativeAbilitiesContainerType
	AttributesContainerType
	LastContainerType = AttributesContainerType
)

// AllContainerType holds all possible values.
var AllContainerType = []ContainerType{
	GroupContainerType,
	MetaTraitContainerType,
	RaceContainerType,
	AlternativeAbilitiesContainerType,
	AttributesContainerType,
}

// ContainerType holds the type of a trait container.
type ContainerType byte

// EnsureValid ensures this is of a known value.
func (enum ContainerType) EnsureValid() ContainerType {
	if enum <= LastContainerType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum ContainerType) Key() string {
	switch enum {
	case GroupContainerType:
		return "group"
	case MetaTraitContainerType:
		return "meta_trait"
	case RaceContainerType:
		return "race"
	case AlternativeAbilitiesContainerType:
		return "alternative_abilities"
	case AttributesContainerType:
		return "attributes"
	default:
		return ContainerType(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum ContainerType) String() string {
	switch enum {
	case GroupContainerType:
		return i18n.Text("Group")
	case MetaTraitContainerType:
		return i18n.Text("Meta-Trait")
	case RaceContainerType:
		return i18n.Text("Race")
	case AlternativeAbilitiesContainerType:
		return i18n.Text("Alternative Abilities")
	case AttributesContainerType:
		return i18n.Text("Attributes")
	default:
		return ContainerType(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum ContainerType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *ContainerType) UnmarshalText(text []byte) error {
	*enum = ExtractContainerType(string(text))
	return nil
}

// ExtractContainerType extracts the value from a string.
func ExtractContainerType(str string) ContainerType {
	for _, enum := range AllContainerType {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
