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
	"github.com/richardwilkes/toolbox/txt"
)

// Possible values.
const (
	GroupContainerType ContainerType = iota
	AlternativeAbilitiesContainerType
	AncestryContainerType
	AttributesContainerType
	MetaTraitContainerType
	LastContainerType = MetaTraitContainerType
)

// AllContainerType holds all possible values.
var AllContainerType = []ContainerType{
	GroupContainerType,
	AlternativeAbilitiesContainerType,
	AncestryContainerType,
	AttributesContainerType,
	MetaTraitContainerType,
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
	case AlternativeAbilitiesContainerType:
		return "alternative_abilities"
	case AncestryContainerType:
		return "ancestry"
	case AttributesContainerType:
		return "attributes"
	case MetaTraitContainerType:
		return "meta_trait"
	default:
		return ContainerType(0).Key()
	}
}

func (enum ContainerType) oldKeys() []string {
	switch enum {
	case GroupContainerType:
		return nil
	case AlternativeAbilitiesContainerType:
		return nil
	case AncestryContainerType:
		return []string{"race"}
	case AttributesContainerType:
		return nil
	case MetaTraitContainerType:
		return nil
	default:
		return ContainerType(0).oldKeys()
	}
}

// String implements fmt.Stringer.
func (enum ContainerType) String() string {
	switch enum {
	case GroupContainerType:
		return i18n.Text("Group")
	case AlternativeAbilitiesContainerType:
		return i18n.Text("Alternative Abilities")
	case AncestryContainerType:
		return i18n.Text("Ancestry")
	case AttributesContainerType:
		return i18n.Text("Attributes")
	case MetaTraitContainerType:
		return i18n.Text("Meta-Trait")
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
		if strings.EqualFold(enum.Key(), str) || txt.CaselessSliceContains(enum.oldKeys(), str) {
			return enum
		}
	}
	return 0
}
