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
	TextCellType CellType = iota
	TagsCellType
	ToggleCellType
	PageRefCellType
	MarkdownCellType
	LastCellType = MarkdownCellType
)

// AllCellType holds all possible values.
var AllCellType = []CellType{
	TextCellType,
	TagsCellType,
	ToggleCellType,
	PageRefCellType,
	MarkdownCellType,
}

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
	switch enum {
	case TextCellType:
		return "text"
	case TagsCellType:
		return "tags"
	case ToggleCellType:
		return "toggle"
	case PageRefCellType:
		return "page_ref"
	case MarkdownCellType:
		return "markdown"
	default:
		return CellType(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum CellType) String() string {
	switch enum {
	case TextCellType:
		return i18n.Text("Text")
	case TagsCellType:
		return i18n.Text("Tags")
	case ToggleCellType:
		return i18n.Text("Toggle")
	case PageRefCellType:
		return i18n.Text("Page Ref")
	case MarkdownCellType:
		return i18n.Text("Markdown")
	default:
		return CellType(0).String()
	}
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

// ExtractCellType extracts the value from a string.
func ExtractCellType(str string) CellType {
	for _, enum := range AllCellType {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
