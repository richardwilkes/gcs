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
	SettingsDockableGroup DockableGroup = iota
	EditorsDockableGroup
	SubEditorsDockableGroup
	LastDockableGroup = SubEditorsDockableGroup
)

// AllDockableGroup holds all possible values.
var AllDockableGroup = []DockableGroup{
	SettingsDockableGroup,
	EditorsDockableGroup,
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
	case SettingsDockableGroup:
		return "settings"
	case EditorsDockableGroup:
		return "editors"
	case SubEditorsDockableGroup:
		return "sub-editors"
	default:
		return DockableGroup(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum DockableGroup) String() string {
	switch enum {
	case SettingsDockableGroup:
		return i18n.Text("Settings")
	case EditorsDockableGroup:
		return i18n.Text("Editors")
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
