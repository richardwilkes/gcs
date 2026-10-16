// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

import (
	"context"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// Item is an opaque host object. It must be comparable (in practice a pointer): pools, memoization and hits key on it.
type Item any

// Host supplies everything an evaluation needs to know about the data it runs against. The engine never retains a
// Host or an Item beyond the call it was given them in.
type Host interface {
	// Collection returns every member of a named collection, self included; the engine leaves self out unless asked to
	// keep it. Entity collections (traits, skills, spells, equipped...) may use self, such as for guards; relative
	// collections (children, modifiers...) belong to subject. ok is false for an unknown key or when no entity is
	// available (a filter's host), which fails the evaluation.
	Collection(key string, self, subject Item) (items []Item, ok bool)
	// Field returns a field of item. ok is false when the key names no field of the item's kind, which fails the
	// evaluation. A field the item's kind has but that is unset for this item (a nil tech level) returns ok true with
	// Kind ValueAbsent.
	Field(item Item, key string) (v Value, ok bool)
	// Attribute returns the current value of an attribute.
	Attribute(key string) (fxp.Int, bool)
	// Entity returns a character-level number, such as tech_level or sm.
	Entity(key string) (fxp.Int, bool)
	// Replacements returns the nameable replacements for qualifiers evaluated with this owner and subject (prereqs and
	// bonuses: the owner's; overrides: the candidate's; filters: nil). Cached per (self, subject) within one
	// evaluation.
	Replacements(self, subject Item) map[string]string
	// RunScript runs a script node; memoized per (node, self, subject) within one evaluation. It should stop and
	// return Failed once ctx is done.
	RunScript(ctx context.Context, s *Script, self, subject Item) (State, string)
}

// FieldResolver is optionally implemented by a Host. of is "" for the root subject, otherwise the key of the enclosing
// collection, which a Matcher knows statically from the check's position. The returned func must behave exactly like
// Host.Field.
type FieldResolver interface {
	FieldFunc(of, key string) (func(Item) (Value, bool), bool)
}

// ValueKind identifies the kind of data a Value holds. It is never persisted.
type ValueKind byte

// Possible ValueKind values.
const (
	ValueAbsent ValueKind = iota
	ValueText
	ValueList
	ValueNumber
	ValueWeight
	ValueBool
)

// String implements fmt.Stringer.
func (k ValueKind) String() string {
	switch k {
	case ValueAbsent:
		return "absent"
	case ValueText:
		return "text"
	case ValueList:
		return "list"
	case ValueNumber:
		return "number"
	case ValueWeight:
		return "weight"
	case ValueBool:
		return "bool"
	default:
		return "unknown"
	}
}

// Value is a field value reported by a Host.
type Value struct {
	Kind   ValueKind
	Text   string
	List   []string
	Number fxp.Int // Weight values are in pounds
	Bool   bool
}
