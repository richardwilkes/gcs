// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

// Package themeset holds the machinery shared by the theme color and theme font settings: a set of values keyed by ID
// that can capture the live theme, be applied back to it, and serialize in a stable, factory-defined order.
package themeset

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
)

// Entry is one themed value. The live theme is a list of entries whose values the settings UI edits in place. A Set
// captures those values so they can be written out and applies them back once loaded.
type Entry[V any] interface {
	// Key returns the ID of the entry, which is used as its JSON key.
	Key() string
	// Value returns the entry's current value.
	Value() V
	// SetValue replaces the entry's current value.
	SetValue(v V)
}

// Provider supplies the entries a Set works with. It is a type parameter of Set rather than a field so that a zero
// value Set is fully usable, which the global settings rely upon on a first run, when there is no settings file to
// load. Implementations must be usable as their zero value.
type Provider[V any, E Entry[V]] interface {
	// Current returns the live entries.
	Current() []E
	// Factory returns entries holding the original, unmodified values. Their order is the order the keys are written
	// in.
	Factory() []E
	// Applied is called after MakeCurrent has written its values into the live entries.
	Applied()
}

// Set holds a set of themed values keyed by ID. It exists for serialization: the settings UI edits the live entries in
// place and never touches a Set, so anything that represents the live theme must call CaptureCurrent before it is
// written, or those edits are lost.
type Set[V any, E Entry[V], P Provider[V, E]] struct {
	data map[string]V
}

// MarshalJSONTo implements json.MarshalerTo. This writes the receiver's own values, not the live theme; see the note on
// CaptureCurrent. The factory list drives the iteration so that the keys are written in a stable, meaningful order. A
// value the receiver doesn't define falls back to the factory value, matching what UnmarshalJSONFrom fills in for a
// missing key.
func (s *Set[V, E, P]) MarshalJSONTo(enc *jsontext.Encoder) error {
	if err := enc.WriteToken(jsontext.BeginObject); err != nil {
		return err
	}
	var p P
	for _, one := range p.Factory() {
		if err := enc.WriteToken(jsontext.String(one.Key())); err != nil {
			return err
		}
		v, ok := s.data[one.Key()]
		if !ok {
			v = one.Value()
		}
		if err := json.MarshalEncode(enc, v); err != nil {
			return err
		}
	}
	return enc.WriteToken(jsontext.EndObject)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom. Keys the data doesn't define are filled in with the factory
// values, so that the result is always complete, even when an error is returned.
func (s *Set[V, E, P]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	s.data = nil
	err := json.UnmarshalDecode(dec, &s.data)
	if err != nil {
		s.data = nil
	}
	var p P
	s.ensureData()
	for _, one := range p.Factory() {
		if _, ok := s.data[one.Key()]; !ok {
			s.data[one.Key()] = one.Value()
		}
	}
	return err
}

// CaptureCurrent copies the live theme into this object so that a subsequent save writes it. The settings UI mutates
// the live entries in place and never touches this object, so anything that represents the live theme has to call this
// before saving or those edits are lost.
func (s *Set[V, E, P]) CaptureCurrent() {
	var p P
	s.assign(p.Current())
}

// MakeCurrent applies these values to the live theme, then invokes the provider's Applied hook. Values this object
// doesn't define are left as they are.
func (s *Set[V, E, P]) MakeCurrent() {
	var p P
	for _, one := range p.Current() {
		if v, ok := s.data[one.Key()]; ok {
			one.SetValue(v)
		}
	}
	p.Applied()
}

// Reset to factory defaults.
func (s *Set[V, E, P]) Reset() {
	var p P
	s.assign(p.Factory())
}

// ResetOne resets one value by ID to its factory default. Unknown IDs are ignored.
func (s *Set[V, E, P]) ResetOne(id string) {
	var p P
	for _, one := range p.Factory() {
		if one.Key() == id {
			s.ensureData()
			s.data[id] = one.Value()
			return
		}
	}
}

func (s *Set[V, E, P]) assign(entries []E) {
	s.ensureData()
	for _, one := range entries {
		s.data[one.Key()] = one.Value()
	}
}

// ensureData allocates the map if this object has never been through UnmarshalJSONFrom -- the first-run case, where no
// settings file exists.
func (s *Set[V, E, P]) ensureData() {
	if s.data == nil {
		var p P
		s.data = make(map[string]V, len(p.Factory()))
	}
}
