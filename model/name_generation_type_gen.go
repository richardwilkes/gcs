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
	SimpleNameGenerationType NameGenerationType = iota
	MarkovChainNameGenerationType
	LastNameGenerationType = MarkovChainNameGenerationType
)

// AllNameGenerationType holds all possible values.
var AllNameGenerationType = []NameGenerationType{
	SimpleNameGenerationType,
	MarkovChainNameGenerationType,
}

// NameGenerationType holds a name generation type.
type NameGenerationType byte

// EnsureValid ensures this is of a known value.
func (enum NameGenerationType) EnsureValid() NameGenerationType {
	if enum <= LastNameGenerationType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum NameGenerationType) Key() string {
	switch enum {
	case SimpleNameGenerationType:
		return "simple"
	case MarkovChainNameGenerationType:
		return "markov_chain"
	default:
		return NameGenerationType(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum NameGenerationType) String() string {
	switch enum {
	case SimpleNameGenerationType:
		return i18n.Text("Simple")
	case MarkovChainNameGenerationType:
		return i18n.Text("Markov Chain")
	default:
		return NameGenerationType(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum NameGenerationType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *NameGenerationType) UnmarshalText(text []byte) error {
	*enum = ExtractNameGenerationType(string(text))
	return nil
}

// ExtractNameGenerationType extracts the value from a string.
func ExtractNameGenerationType(str string) NameGenerationType {
	for _, enum := range AllNameGenerationType {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
