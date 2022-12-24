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
	NotApplicableTemplatePickerType TemplatePickerType = iota
	CountTemplatePickerType
	PointsTemplatePickerType
	LastTemplatePickerType = PointsTemplatePickerType
)

// AllTemplatePickerType holds all possible values.
var AllTemplatePickerType = []TemplatePickerType{
	NotApplicableTemplatePickerType,
	CountTemplatePickerType,
	PointsTemplatePickerType,
}

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
	switch enum {
	case NotApplicableTemplatePickerType:
		return "not_applicable"
	case CountTemplatePickerType:
		return "count"
	case PointsTemplatePickerType:
		return "points"
	default:
		return TemplatePickerType(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum TemplatePickerType) String() string {
	switch enum {
	case NotApplicableTemplatePickerType:
		return i18n.Text("Not Applicable")
	case CountTemplatePickerType:
		return i18n.Text("Count")
	case PointsTemplatePickerType:
		return i18n.Text("Points")
	default:
		return TemplatePickerType(0).String()
	}
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

// ExtractTemplatePickerType extracts the value from a string.
func ExtractTemplatePickerType(str string) TemplatePickerType {
	for _, enum := range AllTemplatePickerType {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
