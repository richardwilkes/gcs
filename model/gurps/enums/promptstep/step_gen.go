// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package promptstep

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	None Step = iota
	Destination
	Targets
	Choice
	Level
	Points
	Quantity
	RemoveChoices
	Modifiers
	Substitutions
	Ancestry
	Randomize
)

// DefaultStep is the default value.
const DefaultStep Step = None

// FirstStep is the first valid value.
const FirstStep Step = None

// LastStep is the last valid value.
const LastStep Step = Randomize

// Steps holds all possible values.
var Steps = []Step{
	None,
	Destination,
	Targets,
	Choice,
	Level,
	Points,
	Quantity,
	RemoveChoices,
	Modifiers,
	Substitutions,
	Ancestry,
	Randomize,
}

// Step identifies the step of an operation a prompt is for, such as settling the modifiers of rows being applied.
type Step byte

// EnsureValid ensures this is of a known value.
func (enum Step) EnsureValid() Step {
	if enum >= FirstStep && enum <= LastStep {
		return enum
	}
	return DefaultStep
}

// Key returns the key used in serialization.
func (enum Step) Key() string {
	switch enum {
	case None:
		return "none"
	case Destination:
		return "destination"
	case Targets:
		return "targets"
	case Choice:
		return "choice"
	case Level:
		return "level"
	case Points:
		return "points"
	case Quantity:
		return "quantity"
	case RemoveChoices:
		return "remove_choices"
	case Modifiers:
		return "modifiers"
	case Substitutions:
		return "substitutions"
	case Ancestry:
		return "ancestry"
	case Randomize:
		return "randomize"
	default:
		return DefaultStep.Key()
	}
}

// String implements fmt.Stringer.
func (enum Step) String() string {
	switch enum {
	case None:
		return ``
	case Destination:
		return i18n.Text(`Destination`)
	case Targets:
		return i18n.Text(`Targets`)
	case Choice:
		return i18n.Text(`Choice`)
	case Level:
		return i18n.Text(`Level`)
	case Points:
		return i18n.Text(`Points`)
	case Quantity:
		return i18n.Text(`Quantity`)
	case RemoveChoices:
		return i18n.Text(`Remove Choices`)
	case Modifiers:
		return i18n.Text(`Modifiers`)
	case Substitutions:
		return i18n.Text(`Substitutions`)
	case Ancestry:
		return i18n.Text(`Ancestry`)
	case Randomize:
		return i18n.Text(`Randomize`)
	default:
		return DefaultStep.String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Step) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Step) UnmarshalText(text []byte) error {
	*enum = ExtractStep(string(text))
	return nil
}

// ExtractStep extracts the value from a string.
func ExtractStep(str string) Step {
	for _, enum := range Steps {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultStep
}

// ExtractKnownStep extracts the value from a string, reporting whether the string was actually recognized.
//
// Unlike ExtractStep, which quietly maps anything it doesn't recognize onto the default value, this permits a caller
// that is dispatching on the type to detect unknown types.
func ExtractKnownStep(str string) (value Step, known bool) {
	for _, enum := range Steps {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultStep, false
}
