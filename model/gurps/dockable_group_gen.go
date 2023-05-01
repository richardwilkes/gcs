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
	CharacterSheetsDockableGroup DockableGroup = iota
	CharacterTemplatesDockableGroup
	EditorsDockableGroup
	ImagesDockableGroup
	LibrariesDockableGroup
	MarkdownDockableGroup
	PDFsDockableGroup
	SettingsDockableGroup
	SubEditorsDockableGroup
	LastDockableGroup = SubEditorsDockableGroup
)

// AllDockableGroup holds all possible values.
var AllDockableGroup = []DockableGroup{
	CharacterSheetsDockableGroup,
	CharacterTemplatesDockableGroup,
	EditorsDockableGroup,
	ImagesDockableGroup,
	LibrariesDockableGroup,
	MarkdownDockableGroup,
	PDFsDockableGroup,
	SettingsDockableGroup,
	SubEditorsDockableGroup,
}

// DockableGroup holds the set of dockable groupings.
type DockableGroup byte

// EnsureValid ensures this is of a known value.
func (enum DockableGroup) EnsureValid() DockableGroup {
	if enum <= LastDockableGroup {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum DockableGroup) Key() string {
	switch enum {
	case CharacterSheetsDockableGroup:
		return "character_sheets"
	case CharacterTemplatesDockableGroup:
		return "character_templates"
	case EditorsDockableGroup:
		return "editors"
	case ImagesDockableGroup:
		return "images"
	case LibrariesDockableGroup:
		return "libraries"
	case MarkdownDockableGroup:
		return "markdown"
	case PDFsDockableGroup:
		return "pdfs"
	case SettingsDockableGroup:
		return "settings"
	case SubEditorsDockableGroup:
		return "sub-editors"
	default:
		return DockableGroup(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum DockableGroup) String() string {
	switch enum {
	case CharacterSheetsDockableGroup:
		return i18n.Text("Character Sheets")
	case CharacterTemplatesDockableGroup:
		return i18n.Text("Character Templates")
	case EditorsDockableGroup:
		return i18n.Text("Editors")
	case ImagesDockableGroup:
		return i18n.Text("Images")
	case LibrariesDockableGroup:
		return i18n.Text("Libraries")
	case MarkdownDockableGroup:
		return i18n.Text("Markdown")
	case PDFsDockableGroup:
		return i18n.Text("PDFs")
	case SettingsDockableGroup:
		return i18n.Text("Settings")
	case SubEditorsDockableGroup:
		return i18n.Text("Sub-Editors")
	default:
		return DockableGroup(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum DockableGroup) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *DockableGroup) UnmarshalText(text []byte) error {
	*enum = ExtractDockableGroup(string(text))
	return nil
}

// ExtractDockableGroup extracts the value from a string.
func ExtractDockableGroup(str string) DockableGroup {
	for _, enum := range AllDockableGroup {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
