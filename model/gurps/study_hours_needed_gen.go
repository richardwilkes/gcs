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

// Code generated from "enum.go.tmpl" - DO NOT EDIT.

package gurps

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	StandardStudyHoursNeeded StudyHoursNeeded = iota
	Level1StudyHoursNeeded
	Level2StudyHoursNeeded
	Level3StudyHoursNeeded
	Level4StudyHoursNeeded
	LastStudyHoursNeeded = Level4StudyHoursNeeded
)

// AllStudyHoursNeeded holds all possible values.
var AllStudyHoursNeeded = []StudyHoursNeeded{
	StandardStudyHoursNeeded,
	Level1StudyHoursNeeded,
	Level2StudyHoursNeeded,
	Level3StudyHoursNeeded,
	Level4StudyHoursNeeded,
}

// StudyHoursNeeded holds the number of study hours required per point.
type StudyHoursNeeded byte

// EnsureValid ensures this is of a known value.
func (enum StudyHoursNeeded) EnsureValid() StudyHoursNeeded {
	if enum <= LastStudyHoursNeeded {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum StudyHoursNeeded) Key() string {
	switch enum {
	case StandardStudyHoursNeeded:
		return ""
	case Level1StudyHoursNeeded:
		return "180"
	case Level2StudyHoursNeeded:
		return "160"
	case Level3StudyHoursNeeded:
		return "140"
	case Level4StudyHoursNeeded:
		return "120"
	default:
		return StudyHoursNeeded(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum StudyHoursNeeded) String() string {
	switch enum {
	case StandardStudyHoursNeeded:
		return i18n.Text("Standard")
	case Level1StudyHoursNeeded:
		return i18n.Text("Reduction for Talent level 1")
	case Level2StudyHoursNeeded:
		return i18n.Text("Reduction for Talent level 2")
	case Level3StudyHoursNeeded:
		return i18n.Text("Reduction for Talent level 3")
	case Level4StudyHoursNeeded:
		return i18n.Text("Reduction for Talent level 4")
	default:
		return StudyHoursNeeded(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum StudyHoursNeeded) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *StudyHoursNeeded) UnmarshalText(text []byte) error {
	*enum = ExtractStudyHoursNeeded(string(text))
	return nil
}

// ExtractStudyHoursNeeded extracts the value from a string.
func ExtractStudyHoursNeeded(str string) StudyHoursNeeded {
	for _, enum := range AllStudyHoursNeeded {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
