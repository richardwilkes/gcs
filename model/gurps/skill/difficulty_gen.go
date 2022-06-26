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

package skill

import (
	"strings"
)

// Possible values.
const (
	Easy Difficulty = iota
	Average
	Hard
	VeryHard
	Wildcard
	LastDifficulty = Wildcard
)

var (
	// AllDifficulty holds all possible values.
	AllDifficulty = []Difficulty{
		Easy,
		Average,
		Hard,
		VeryHard,
		Wildcard,
	}
	difficultyData = []struct {
		key    string
		string string
	}{
		{
			key:    "e",
			string: "E",
		},
		{
			key:    "a",
			string: "A",
		},
		{
			key:    "h",
			string: "H",
		},
		{
			key:    "vh",
			string: "VH",
		},
		{
			key:    "w",
			string: "W",
		},
	}
)

// Difficulty holds the difficulty level of a skill.
type Difficulty byte

// EnsureValid ensures this is of a known value.
func (enum Difficulty) EnsureValid() Difficulty {
	if enum <= LastDifficulty {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Difficulty) Key() string {
	return difficultyData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum Difficulty) String() string {
	return difficultyData[enum.EnsureValid()].string
}

// ExtractDifficulty extracts the value from a string.
func ExtractDifficulty(str string) Difficulty {
	for i, one := range difficultyData {
		if strings.EqualFold(one.key, str) {
			return Difficulty(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Difficulty) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Difficulty) UnmarshalText(text []byte) error {
	*enum = ExtractDifficulty(string(text))
	return nil
}
