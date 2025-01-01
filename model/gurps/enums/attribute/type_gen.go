// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package attribute

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	Integer Type = iota
	IntegerRef
	Decimal
	DecimalRef
	Pool
	PoolRef
	PrimarySeparator
	SecondarySeparator
	PoolSeparator
)

// LastType is the last valid value.
const LastType Type = PoolSeparator

// Types holds all possible values.
var Types = []Type{
	Integer,
	IntegerRef,
	Decimal,
	DecimalRef,
	Pool,
	PoolRef,
	PrimarySeparator,
	SecondarySeparator,
	PoolSeparator,
}

// Type holds the type of an attribute definition.
type Type byte

// EnsureValid ensures this is of a known value.
func (enum Type) EnsureValid() Type {
	if enum <= PoolSeparator {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Type) Key() string {
	switch enum {
	case Integer:
		return "integer"
	case IntegerRef:
		return "integer_ref"
	case Decimal:
		return "decimal"
	case DecimalRef:
		return "decimal_ref"
	case Pool:
		return "pool"
	case PoolRef:
		return "pool_ref"
	case PrimarySeparator:
		return "primary_separator"
	case SecondarySeparator:
		return "secondary_separator"
	case PoolSeparator:
		return "pool_separator"
	default:
		return Type(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Type) String() string {
	switch enum {
	case Integer:
		return i18n.Text("Integer")
	case IntegerRef:
		return i18n.Text("Integer (Display Only)")
	case Decimal:
		return i18n.Text("Decimal")
	case DecimalRef:
		return i18n.Text("Decimal (Display Only)")
	case Pool:
		return i18n.Text("Pool")
	case PoolRef:
		return i18n.Text("Pool (Display Only for Maximum)")
	case PrimarySeparator:
		return i18n.Text("Primary Separator")
	case SecondarySeparator:
		return i18n.Text("Secondary Separator")
	case PoolSeparator:
		return i18n.Text("Pool Separator")
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
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
