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

var (
	// AllStudyType holds all possible values.
	AllStudyType = []StudyType{
		SelfStudyType,
		JobStudyType,
		TeacherStudyType,
		IntensiveStudyType,
	}
	studyTypeData = []struct {
		key    string
		string string
	}{
		{
			key:    "self",
			string: i18n.Text("Self-Taught"),
		},
		{
			key:    "job",
			string: i18n.Text("On-the-Job Training"),
		},
		{
			key:    "teacher",
			string: i18n.Text("Professional Teacher"),
		},
		{
			key:    "intensive",
			string: i18n.Text("Intensive Training"),
		},
	}
)

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
	return studyTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum StudyType) String() string {
	return studyTypeData[enum.EnsureValid()].string
}

// ExtractStudyType extracts the value from a string.
func ExtractStudyType(str string) StudyType {
	for i, one := range studyTypeData {
		if strings.EqualFold(one.key, str) {
			return StudyType(i)
		}
	}
	return 0
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
