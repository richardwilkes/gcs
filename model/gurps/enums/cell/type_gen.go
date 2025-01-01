// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package cell

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	Text Type = iota
	Tags
	Toggle
	PageRef
	Markdown
)

// LastType is the last valid value.
const LastType Type = Markdown

// Types holds all possible values.
var Types = []Type{
	Text,
	Tags,
	Toggle,
	PageRef,
	Markdown,
}

// Type holds the type of table cell.
type Type byte

// EnsureValid ensures this is of a known value.
func (enum Type) EnsureValid() Type {
	if enum <= Markdown {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Type) Key() string {
	switch enum {
	case Text:
		return "text"
	case Tags:
		return "tags"
	case Toggle:
		return "toggle"
	case PageRef:
		return "page_ref"
	case Markdown:
		return "markdown"
	default:
		return Type(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Type) String() string {
	switch enum {
	case Text:
		return i18n.Text("Text")
	case Tags:
		return i18n.Text("Tags")
	case Toggle:
		return i18n.Text("Toggle")
	case PageRef:
		return i18n.Text("Page Ref")
	case Markdown:
		return i18n.Text("Markdown")
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
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
