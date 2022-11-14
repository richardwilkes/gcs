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
	"github.com/richardwilkes/toolbox/txt"
)

// Possible values.
const (
	ListPrereqType PrereqType = iota
	TraitPrereqType
	AttributePrereqType
	ContainedQuantityPrereqType
	ContainedWeightPrereqType
	EquippedEquipmentPrereqType
	SkillPrereqType
	SpellPrereqType
	LastPrereqType = SpellPrereqType
)

var (
	// AllPrereqType holds all possible values.
	AllPrereqType = []PrereqType{
		ListPrereqType,
		TraitPrereqType,
		AttributePrereqType,
		ContainedQuantityPrereqType,
		ContainedWeightPrereqType,
		EquippedEquipmentPrereqType,
		SkillPrereqType,
		SpellPrereqType,
	}
	prereqTypeData = []struct {
		key     string
		oldKeys []string
		string  string
	}{
		{
			key:    "prereq_list",
			string: i18n.Text("a list"),
		},
		{
			key:     "trait_prereq",
			oldKeys: []string{"advantage_prereq"},
			string:  i18n.Text("a trait"),
		},
		{
			key:    "attribute_prereq",
			string: i18n.Text("the attribute"),
		},
		{
			key:    "contained_quantity_prereq",
			string: i18n.Text("a contained quantity of"),
		},
		{
			key:    "contained_weight_prereq",
			string: i18n.Text("a contained weight"),
		},
		{
			key:    "equipped_equipment",
			string: i18n.Text("has equipped equipment"),
		},
		{
			key:    "skill_prereq",
			string: i18n.Text("a skill"),
		},
		{
			key:    "spell_prereq",
			string: i18n.Text("spell(s)"),
		},
	}
)

// PrereqType holds the type of a Prereq.
type PrereqType byte

// EnsureValid ensures this is of a known value.
func (enum PrereqType) EnsureValid() PrereqType {
	if enum <= LastPrereqType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum PrereqType) Key() string {
	return prereqTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum PrereqType) String() string {
	return prereqTypeData[enum.EnsureValid()].string
}

// ExtractPrereqType extracts the value from a string.
func ExtractPrereqType(str string) PrereqType {
	for i, one := range prereqTypeData {
		if strings.EqualFold(one.key, str) || txt.CaselessSliceContains(one.oldKeys, str) {
			return PrereqType(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum PrereqType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *PrereqType) UnmarshalText(text []byte) error {
	*enum = ExtractPrereqType(string(text))
	return nil
}
