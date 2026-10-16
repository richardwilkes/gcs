// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	AggregateNone Aggregate = iota
	AggregateCount
	AggregateSum
	AggregateMax
)

// DefaultAggregate is the default value.
const DefaultAggregate Aggregate = AggregateNone

// FirstAggregate is the first valid value.
const FirstAggregate Aggregate = AggregateNone

// LastAggregate is the last valid value.
const LastAggregate Aggregate = AggregateMax

// Aggregates holds all possible values.
var Aggregates = []Aggregate{
	AggregateNone,
	AggregateCount,
	AggregateSum,
	AggregateMax,
}

// Aggregate holds the way an operand combines the members of a collection.
type Aggregate byte

// EnsureValid returns this value if it is valid and DefaultAggregate otherwise.
func (enum Aggregate) EnsureValid() Aggregate {
	if enum >= FirstAggregate && enum <= LastAggregate {
		return enum
	}
	return DefaultAggregate
}

// Key returns the key used in serialization.
func (enum Aggregate) Key() string {
	switch enum {
	case AggregateNone:
		return ""
	case AggregateCount:
		return "count"
	case AggregateSum:
		return "sum"
	case AggregateMax:
		return "max"
	default:
		return DefaultAggregate.Key()
	}
}

// String implements fmt.Stringer.
func (enum Aggregate) String() string {
	switch enum {
	case AggregateNone:
		return ``
	case AggregateCount:
		return i18n.Text(`count`)
	case AggregateSum:
		return i18n.Text(`sum`)
	case AggregateMax:
		return i18n.Text(`maximum`)
	default:
		return DefaultAggregate.String()
	}
}

// MarshalText implements encoding.TextMarshaler.
func (enum Aggregate) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (enum *Aggregate) UnmarshalText(text []byte) error {
	*enum = ExtractAggregate(string(text))
	return nil
}

// ExtractAggregate returns the value whose key matches str, ignoring case, or DefaultAggregate if none does.
func ExtractAggregate(str string) Aggregate {
	for _, enum := range Aggregates {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultAggregate
}

// ExtractKnownAggregate is like ExtractAggregate, but also reports whether str was recognized, so a caller can tell an
// unknown key from the default.
func ExtractKnownAggregate(str string) (value Aggregate, known bool) {
	for _, enum := range Aggregates {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultAggregate, false
}
