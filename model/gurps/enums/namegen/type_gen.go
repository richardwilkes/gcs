// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package namegen

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
)

// Possible values.
const (
	Simple Type = iota
	MarkovLetter
	MarkovRun
	Compound
)

// LastType is the last valid value.
const LastType Type = Compound

// Types holds all possible values.
var Types = []Type{
	Simple,
	MarkovLetter,
	MarkovRun,
	Compound,
}

// Type holds a name generation type.
type Type byte

// EnsureValid ensures this is of a known value.
func (enum Type) EnsureValid() Type {
	if enum <= Compound {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Type) Key() string {
	switch enum {
	case Simple:
		return "simple"
	case MarkovLetter:
		return "markov_letter"
	case MarkovRun:
		return "markov_run"
	case Compound:
		return "compound"
	default:
		return Type(0).Key()
	}
}

func (enum Type) oldKeys() []string {
	switch enum {
	case Simple:
		return nil
	case MarkovLetter:
		return []string{"markov_chain"}
	case MarkovRun:
		return nil
	case Compound:
		return nil
	default:
		return Type(0).oldKeys()
	}
}

// String implements fmt.Stringer.
func (enum Type) String() string {
	switch enum {
	case Simple:
		return i18n.Text("Simple")
	case MarkovLetter:
		return i18n.Text("Markov Letter")
	case MarkovRun:
		return i18n.Text("Markov Run")
	case Compound:
		return i18n.Text("Compound")
	default:
		return Type(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Type) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Type) UnmarshalText(text []byte) error {
	*enum = ExtractType(string(text))
	return nil
}

// ExtractType extracts the value from a string.
func ExtractType(str string) Type {
	for _, enum := range Types {
		if strings.EqualFold(enum.Key(), str) || txt.CaselessSliceContains(enum.oldKeys(), str) {
			return enum
		}
	}
	return 0
}
