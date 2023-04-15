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
	NoneNameData NameData = iota
	AmericanMaleNameData
	AmericanFemaleNameData
	AmericanLastNameData
	UnweightedAmericanMaleNameData
	UnweightedAmericanFemaleNameData
	UnweightedAmericanLastNameData
	LastNameData = UnweightedAmericanLastNameData
)

// AllNameData holds all possible values.
var AllNameData = []NameData{
	NoneNameData,
	AmericanMaleNameData,
	AmericanFemaleNameData,
	AmericanLastNameData,
	UnweightedAmericanMaleNameData,
	UnweightedAmericanFemaleNameData,
	UnweightedAmericanLastNameData,
}

// NameData holds a built-in name data type.
type NameData byte

// EnsureValid ensures this is of a known value.
func (enum NameData) EnsureValid() NameData {
	if enum <= LastNameData {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum NameData) Key() string {
	switch enum {
	case NoneNameData:
		return "none"
	case AmericanMaleNameData:
		return "american_male"
	case AmericanFemaleNameData:
		return "american_female"
	case AmericanLastNameData:
		return "american_last"
	case UnweightedAmericanMaleNameData:
		return "unweighted_american_male"
	case UnweightedAmericanFemaleNameData:
		return "unweighted_american_female"
	case UnweightedAmericanLastNameData:
		return "unweighted_american_last"
	default:
		return NameData(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum NameData) String() string {
	switch enum {
	case NoneNameData:
		return i18n.Text("None")
	case AmericanMaleNameData:
		return i18n.Text("American Male")
	case AmericanFemaleNameData:
		return i18n.Text("American Female")
	case AmericanLastNameData:
		return i18n.Text("American Last")
	case UnweightedAmericanMaleNameData:
		return i18n.Text("Unweighted American Male")
	case UnweightedAmericanFemaleNameData:
		return i18n.Text("Unweighted American Female")
	case UnweightedAmericanLastNameData:
		return i18n.Text("Unweighted American Last")
	default:
		return NameData(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum NameData) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *NameData) UnmarshalText(text []byte) error {
	*enum = ExtractNameData(string(text))
	return nil
}

// ExtractNameData extracts the value from a string.
func ExtractNameData(str string) NameData {
	for _, enum := range AllNameData {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
