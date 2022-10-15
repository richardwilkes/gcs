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
	NotApplicableTemplatePickerType TemplatePickerType = iota
	CountTemplatePickerType
	PointsTemplatePickerType
	LastTemplatePickerType = PointsTemplatePickerType
)

var (
	// AllTemplatePickerType holds all possible values.
	AllTemplatePickerType = []TemplatePickerType{
		NotApplicableTemplatePickerType,
		CountTemplatePickerType,
		PointsTemplatePickerType,
	}
	templatePickerTypeData = []struct {
		key    string
		string string
	}{
		{
			key:    "not_applicable",
			string: i18n.Text("Not Applicable"),
		},
		{
			key:    "count",
			string: i18n.Text("Count"),
		},
		{
			key:    "points",
			string: i18n.Text("Points"),
		},
	}
)

// TemplatePickerType holds the type of template picker.
type TemplatePickerType byte

// EnsureValid ensures this is of a known value.
func (enum TemplatePickerType) EnsureValid() TemplatePickerType {
	if enum <= LastTemplatePickerType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum TemplatePickerType) Key() string {
	return templatePickerTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum TemplatePickerType) String() string {
	return templatePickerTypeData[enum.EnsureValid()].string
}

// ExtractTemplatePickerType extracts the value from a string.
func ExtractTemplatePickerType(str string) TemplatePickerType {
	for i, one := range templatePickerTypeData {
		if strings.EqualFold(one.key, str) {
			return TemplatePickerType(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum TemplatePickerType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *TemplatePickerType) UnmarshalText(text []byte) error {
	*enum = ExtractTemplatePickerType(string(text))
	return nil
}
