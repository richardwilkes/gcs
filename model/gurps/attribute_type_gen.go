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
	IntegerAttributeType AttributeType = iota
	IntegerRefAttributeType
	DecimalAttributeType
	DecimalRefAttributeType
	PoolAttributeType
	PrimarySeparatorAttributeType
	SecondarySeparatorAttributeType
	PoolSeparatorAttributeType
	LastAttributeType = PoolSeparatorAttributeType
)

// AllAttributeType holds all possible values.
var AllAttributeType = []AttributeType{
	IntegerAttributeType,
	IntegerRefAttributeType,
	DecimalAttributeType,
	DecimalRefAttributeType,
	PoolAttributeType,
	PrimarySeparatorAttributeType,
	SecondarySeparatorAttributeType,
	PoolSeparatorAttributeType,
}

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
	switch enum {
	case IntegerAttributeType:
		return "integer"
	case IntegerRefAttributeType:
		return "integer_ref"
	case DecimalAttributeType:
		return "decimal"
	case DecimalRefAttributeType:
		return "decimal_ref"
	case PoolAttributeType:
		return "pool"
	case PrimarySeparatorAttributeType:
		return "primary_separator"
	case SecondarySeparatorAttributeType:
		return "secondary_separator"
	case PoolSeparatorAttributeType:
		return "pool_separator"
	default:
		return AttributeType(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum AttributeType) String() string {
	switch enum {
	case IntegerAttributeType:
		return i18n.Text("Integer")
	case IntegerRefAttributeType:
		return i18n.Text("Integer (Display Only)")
	case DecimalAttributeType:
		return i18n.Text("Decimal")
	case DecimalRefAttributeType:
		return i18n.Text("Decimal (Display Only)")
	case PoolAttributeType:
		return i18n.Text("Pool")
	case PrimarySeparatorAttributeType:
		return i18n.Text("Primary Separator")
	case SecondarySeparatorAttributeType:
		return i18n.Text("Secondary Separator")
	case PoolSeparatorAttributeType:
		return i18n.Text("Pool Separator")
	default:
		return AttributeType(0).String()
	}
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

// ExtractAttributeType extracts the value from a string.
func ExtractAttributeType(str string) AttributeType {
	for _, enum := range AllAttributeType {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
