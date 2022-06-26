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

package trait

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	Group ContainerType = iota
	MetaTrait
	Race
	AlternativeAbilities
	LastContainerType = AlternativeAbilities
)

var (
	// AllContainerType holds all possible values.
	AllContainerType = []ContainerType{
		Group,
		MetaTrait,
		Race,
		AlternativeAbilities,
	}
	containerTypeData = []struct {
		key    string
		string string
	}{
		{
			key:    "group",
			string: i18n.Text("Group"),
		},
		{
			key:    "meta_trait",
			string: i18n.Text("Meta-Trait"),
		},
		{
			key:    "race",
			string: i18n.Text("Race"),
		},
		{
			key:    "alternative_abilities",
			string: i18n.Text("Alternative Abilities"),
		},
	}
)

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
	return containerTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum ContainerType) String() string {
	return containerTypeData[enum.EnsureValid()].string
}

// ExtractContainerType extracts the value from a string.
func ExtractContainerType(str string) ContainerType {
	for i, one := range containerTypeData {
		if strings.EqualFold(one.key, str) {
			return ContainerType(i)
		}
	}
	return 0
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
