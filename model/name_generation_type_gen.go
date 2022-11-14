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

var (
	// AllNameGenerationType holds all possible values.
	AllNameGenerationType = []NameGenerationType{
		SimpleNameGenerationType,
		MarkovChainNameGenerationType,
	}
	nameGenerationTypeData = []struct {
		key    string
		string string
	}{
		{
			key:    "simple",
			string: i18n.Text("Simple"),
		},
		{
			key:    "markov_chain",
			string: i18n.Text("Markov Chain"),
		},
	}
)

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
	return nameGenerationTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum NameGenerationType) String() string {
	return nameGenerationTypeData[enum.EnsureValid()].string
}

// ExtractNameGenerationType extracts the value from a string.
func ExtractNameGenerationType(str string) NameGenerationType {
	for i, one := range nameGenerationTypeData {
		if strings.EqualFold(one.key, str) {
			return NameGenerationType(i)
		}
	}
	return 0
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
