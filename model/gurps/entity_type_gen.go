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
	PC             EntityType = iota
	LastEntityType            = PC
)

var (
	// AllEntityType holds all possible values.
	AllEntityType = []EntityType{
		PC,
	}
	entityTypeData = []struct {
		key    string
		string string
	}{
		{
			key:    "character",
			string: i18n.Text("PC"),
		},
	}
)

// EntityType holds the type of an Entity.
type EntityType byte

// EnsureValid ensures this is of a known value.
func (enum EntityType) EnsureValid() EntityType {
	if enum <= LastEntityType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum EntityType) Key() string {
	return entityTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum EntityType) String() string {
	return entityTypeData[enum.EnsureValid()].string
}

// ExtractEntityType extracts the value from a string.
func ExtractEntityType(str string) EntityType {
	for i, one := range entityTypeData {
		if strings.EqualFold(one.key, str) {
			return EntityType(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum EntityType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *EntityType) UnmarshalText(text []byte) error {
	*enum = ExtractEntityType(string(text))
	return nil
}
