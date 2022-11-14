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

var (
	// AllPaperSize holds all possible values.
	AllPaperSize = []PaperSize{
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
	paperSizeData = []struct {
		key    string
		string string
	}{
		{
			key:    "letter",
			string: i18n.Text("Letter"),
		},
		{
			key:    "legal",
			string: i18n.Text("Legal"),
		},
		{
			key:    "tabloid",
			string: i18n.Text("Tabloid"),
		},
		{
			key:    "a0",
			string: i18n.Text("A0"),
		},
		{
			key:    "a1",
			string: i18n.Text("A1"),
		},
		{
			key:    "a2",
			string: i18n.Text("A2"),
		},
		{
			key:    "a3",
			string: i18n.Text("A3"),
		},
		{
			key:    "a4",
			string: i18n.Text("A4"),
		},
		{
			key:    "a5",
			string: i18n.Text("A5"),
		},
		{
			key:    "a6",
			string: i18n.Text("A6"),
		},
	}
)

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
	return paperSizeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum PaperSize) String() string {
	return paperSizeData[enum.EnsureValid()].string
}

// ExtractPaperSize extracts the value from a string.
func ExtractPaperSize(str string) PaperSize {
	for i, one := range paperSizeData {
		if strings.EqualFold(one.key, str) {
			return PaperSize(i)
		}
	}
	return 0
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
