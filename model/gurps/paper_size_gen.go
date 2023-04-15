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
	LetterPaperSize PaperSize = iota
	LegalPaperSize
	TabloidPaperSize
	A0PaperSize
	A1PaperSize
	A2PaperSize
	A3PaperSize
	A4PaperSize
	A5PaperSize
	A6PaperSize
	LastPaperSize = A6PaperSize
)

// AllPaperSize holds all possible values.
var AllPaperSize = []PaperSize{
	LetterPaperSize,
	LegalPaperSize,
	TabloidPaperSize,
	A0PaperSize,
	A1PaperSize,
	A2PaperSize,
	A3PaperSize,
	A4PaperSize,
	A5PaperSize,
	A6PaperSize,
}

// PaperSize holds a standard paper dimension.
type PaperSize byte

// EnsureValid ensures this is of a known value.
func (enum PaperSize) EnsureValid() PaperSize {
	if enum <= LastPaperSize {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum PaperSize) Key() string {
	switch enum {
	case LetterPaperSize:
		return "letter"
	case LegalPaperSize:
		return "legal"
	case TabloidPaperSize:
		return "tabloid"
	case A0PaperSize:
		return "a0"
	case A1PaperSize:
		return "a1"
	case A2PaperSize:
		return "a2"
	case A3PaperSize:
		return "a3"
	case A4PaperSize:
		return "a4"
	case A5PaperSize:
		return "a5"
	case A6PaperSize:
		return "a6"
	default:
		return PaperSize(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum PaperSize) String() string {
	switch enum {
	case LetterPaperSize:
		return i18n.Text("Letter")
	case LegalPaperSize:
		return i18n.Text("Legal")
	case TabloidPaperSize:
		return i18n.Text("Tabloid")
	case A0PaperSize:
		return i18n.Text("A0")
	case A1PaperSize:
		return i18n.Text("A1")
	case A2PaperSize:
		return i18n.Text("A2")
	case A3PaperSize:
		return i18n.Text("A3")
	case A4PaperSize:
		return i18n.Text("A4")
	case A5PaperSize:
		return i18n.Text("A5")
	case A6PaperSize:
		return i18n.Text("A6")
	default:
		return PaperSize(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum PaperSize) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *PaperSize) UnmarshalText(text []byte) error {
	*enum = ExtractPaperSize(string(text))
	return nil
}

// ExtractPaperSize extracts the value from a string.
func ExtractPaperSize(str string) PaperSize {
	for _, enum := range AllPaperSize {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
