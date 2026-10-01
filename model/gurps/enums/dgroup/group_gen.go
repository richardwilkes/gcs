// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package dgroup

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	CharacterSheets Group = iota
	CharacterTemplates
	LootSheets
	Editors
	Images
	Libraries
	Markdown
	PDFs
	Settings
	SubEditors
)

// DefaultGroup is the default value.
const DefaultGroup Group = CharacterSheets

// FirstGroup is the first valid value.
const FirstGroup Group = CharacterSheets

// LastGroup is the last valid value.
const LastGroup Group = SubEditors

// Groups holds all possible values.
var Groups = []Group{
	CharacterSheets,
	CharacterTemplates,
	LootSheets,
	Editors,
	Images,
	Libraries,
	Markdown,
	PDFs,
	Settings,
	SubEditors,
}

// Group holds the set of dockable groupings.
type Group byte

// EnsureValid returns this value if it is valid and DefaultGroup otherwise.
func (enum Group) EnsureValid() Group {
	if enum >= FirstGroup && enum <= LastGroup {
		return enum
	}
	return DefaultGroup
}

// Key returns the key used in serialization.
func (enum Group) Key() string {
	switch enum {
	case CharacterSheets:
		return "character_sheets"
	case CharacterTemplates:
		return "character_templates"
	case LootSheets:
		return "loot_sheets"
	case Editors:
		return "editors"
	case Images:
		return "images"
	case Libraries:
		return "libraries"
	case Markdown:
		return "markdown"
	case PDFs:
		return "pdfs"
	case Settings:
		return "settings"
	case SubEditors:
		return "sub-editors"
	default:
		return DefaultGroup.Key()
	}
}

// String implements fmt.Stringer.
func (enum Group) String() string {
	switch enum {
	case CharacterSheets:
		return i18n.Text(`Character Sheets`)
	case CharacterTemplates:
		return i18n.Text(`Character Templates`)
	case LootSheets:
		return i18n.Text(`Loot Sheets`)
	case Editors:
		return i18n.Text(`Editors`)
	case Images:
		return i18n.Text(`Images`)
	case Libraries:
		return i18n.Text(`Libraries`)
	case Markdown:
		return i18n.Text(`Markdown`)
	case PDFs:
		return i18n.Text(`PDFs`)
	case Settings:
		return i18n.Text(`Settings`)
	case SubEditors:
		return i18n.Text(`Sub-Editors`)
	default:
		return DefaultGroup.String()
	}
}

// MarshalText implements encoding.TextMarshaler.
func (enum Group) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (enum *Group) UnmarshalText(text []byte) error {
	*enum = ExtractGroup(string(text))
	return nil
}

// ExtractGroup returns the value whose key matches str, ignoring case, or DefaultGroup if none does.
func ExtractGroup(str string) Group {
	for _, enum := range Groups {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultGroup
}

// ExtractKnownGroup is like ExtractGroup, but also reports whether str was recognized, so a caller can tell an unknown
// key from the default.
func ExtractKnownGroup(str string) (value Group, known bool) {
	for _, enum := range Groups {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultGroup, false
}
