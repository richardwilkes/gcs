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
	IntegerAttributeType AttributeType = iota
	DecimalAttributeType
	PoolAttributeType
	PrimarySeparatorAttributeType
	SecondarySeparatorAttributeType
	PoolSeparatorAttributeType
	LastAttributeType = PoolSeparatorAttributeType
)

var (
	// AllAttributeType holds all possible values.
	AllAttributeType = []AttributeType{
		IntegerAttributeType,
		DecimalAttributeType,
		PoolAttributeType,
		PrimarySeparatorAttributeType,
		SecondarySeparatorAttributeType,
		PoolSeparatorAttributeType,
	}
	attributeTypeData = []struct {
		key    string
		string string
	}{
		{
			key:    "integer",
			string: i18n.Text("Integer"),
		},
		{
			key:    "decimal",
			string: i18n.Text("Decimal"),
		},
		{
			key:    "pool",
			string: i18n.Text("Pool"),
		},
		{
			key:    "primary_separator",
			string: i18n.Text("Primary Separator"),
		},
		{
			key:    "secondary_separator",
			string: i18n.Text("Secondary Separator"),
		},
		{
			key:    "pool_separator",
			string: i18n.Text("Pool Separator"),
		},
	}
)

// AttributeType holds the type of an attribute definition.
type AttributeType byte

// EnsureValid ensures this is of a known value.
func (enum AttributeType) EnsureValid() AttributeType {
	if enum <= LastAttributeType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum AttributeType) Key() string {
	return attributeTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum AttributeType) String() string {
	return attributeTypeData[enum.EnsureValid()].string
}

// ExtractAttributeType extracts the value from a string.
func ExtractAttributeType(str string) AttributeType {
	for i, one := range attributeTypeData {
		if strings.EqualFold(one.key, str) {
			return AttributeType(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum AttributeType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *AttributeType) UnmarshalText(text []byte) error {
	*enum = ExtractAttributeType(string(text))
	return nil
}
