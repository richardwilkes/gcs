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
	Text CellType = iota
	Toggle
	PageRef
	LastCellType = PageRef
)

var (
	// AllCellType holds all possible values.
	AllCellType = []CellType{
		Text,
		Toggle,
		PageRef,
	}
	cellTypeData = []struct {
		key    string
		string string
	}{
		{
			key:    "text",
			string: i18n.Text("Text"),
		},
		{
			key:    "toggle",
			string: i18n.Text("Toggle"),
		},
		{
			key:    "page_ref",
			string: i18n.Text("Page Ref"),
		},
	}
)

// CellType holds the type of table cell.
type CellType byte

// EnsureValid ensures this is of a known value.
func (enum CellType) EnsureValid() CellType {
	if enum <= LastCellType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum CellType) Key() string {
	return cellTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum CellType) String() string {
	return cellTypeData[enum.EnsureValid()].string
}

// ExtractCellType extracts the value from a string.
func ExtractCellType(str string) CellType {
	for i, one := range cellTypeData {
		if strings.EqualFold(one.key, str) {
			return CellType(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum CellType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *CellType) UnmarshalText(text []byte) error {
	*enum = ExtractCellType(string(text))
	return nil
}
