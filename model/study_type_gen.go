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
	SelfStudyType StudyType = iota
	JobStudyType
	TeacherStudyType
	IntensiveStudyType
	LastStudyType = IntensiveStudyType
)

// AllStudyType holds all possible values.
var AllStudyType = []StudyType{
	SelfStudyType,
	JobStudyType,
	TeacherStudyType,
	IntensiveStudyType,
}

// StudyType holds the type of study.
type StudyType byte

// EnsureValid ensures this is of a known value.
func (enum StudyType) EnsureValid() StudyType {
	if enum <= LastStudyType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum StudyType) Key() string {
	switch enum {
	case SelfStudyType:
		return "self"
	case JobStudyType:
		return "job"
	case TeacherStudyType:
		return "teacher"
	case IntensiveStudyType:
		return "intensive"
	default:
		return StudyType(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum StudyType) String() string {
	switch enum {
	case SelfStudyType:
		return i18n.Text("Self-Taught")
	case JobStudyType:
		return i18n.Text("On-the-Job Training")
	case TeacherStudyType:
		return i18n.Text("Professional Teacher")
	case IntensiveStudyType:
		return i18n.Text("Intensive Training")
	default:
		return StudyType(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum StudyType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *StudyType) UnmarshalText(text []byte) error {
	*enum = ExtractStudyType(string(text))
	return nil
}

// ExtractStudyType extracts the value from a string.
func ExtractStudyType(str string) StudyType {
	for _, enum := range AllStudyType {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
