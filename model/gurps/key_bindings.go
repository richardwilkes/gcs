// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"io/fs"
	"maps"

	"github.com/richardwilkes/gcs/v5/model/jio"
)

// factoryBindings maps the ID of each registered key binding to its factory default.
var factoryBindings = make(map[string]string)

// KeyBindings holds a set of key bindings. Each binding is held in the text form the user interface serializes it to,
// which the model never interprets, so two bindings are the same only when their text is identical. Bindings for IDs
// that aren't registered are carried along as they are, since a run without the user interface, such as a file
// conversion, registers none and must not lose them; the user interface drops them with Prune once it has registered
// its own.
type KeyBindings struct {
	data map[string]string
}

// RegisterKeyBinding registers the factory default for a key binding. A second registration of the same ID is ignored.
func RegisterKeyBinding(id, binding string) {
	if _, exists := factoryBindings[id]; !exists {
		factoryBindings[id] = binding
	}
}

// NewKeyBindingsFromFS creates a new set of key bindings from a file. Any missing values will be filled in with
// defaults.
func NewKeyBindingsFromFS(fileSystem fs.FS, filePath string) (*KeyBindings, error) {
	return jio.LoadNew[KeyBindings](fileSystem, filePath)
}

// IsZero reports whether json's omitzero option should omit this value, which is when every binding is at its factory
// default.
func (b *KeyBindings) IsZero() bool {
	for k, v := range b.data {
		if factory, ok := factoryBindings[k]; !ok || v != factory {
			return false
		}
	}
	return true
}

// Save writes the key bindings to the file as JSON.
func (b *KeyBindings) Save(filePath string) error {
	return jio.SaveToFile(filePath, b)
}

// MarshalJSONTo implements json.MarshalerTo. Only the bindings that differ from their factory default are written.
func (b *KeyBindings) MarshalJSONTo(enc *jsontext.Encoder) error {
	data := make(map[string]string, len(b.data))
	for k, v := range b.data {
		if factory, ok := factoryBindings[k]; !ok || factory != v {
			data[k] = v
		}
	}
	return json.MarshalEncode(enc, &data)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (b *KeyBindings) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	m := make(map[string]string, len(factoryBindings))
	if err := json.UnmarshalDecode(dec, &m); err != nil {
		return err
	}
	b.data = m
	return nil
}

// Current returns the binding for the given ID, or an empty string if the ID isn't registered.
func (b *KeyBindings) Current(id string) string {
	if factory, ok := factoryBindings[id]; ok {
		if c, ok2 := b.data[id]; ok2 {
			return c
		}
		return factory
	}
	return ""
}

// Set the binding for the given ID.
func (b *KeyBindings) Set(id, binding string) {
	if factory, ok := factoryBindings[id]; ok {
		if b.data == nil {
			b.data = make(map[string]string, len(factoryBindings))
		}
		if factory != binding {
			b.data[id] = binding
		} else {
			delete(b.data, id)
		}
	}
}

// Reset all bindings to the factory defaults.
func (b *KeyBindings) Reset() {
	b.data = nil
}

// Prune drops the bindings for IDs that aren't registered. The user interface calls it once every binding is
// registered, so that the IDs of actions that no longer exist don't linger in the settings and the files they are
// exported to, where nothing lists them and a later action given the same ID would silently take on the binding.
func (b *KeyBindings) Prune() {
	maps.DeleteFunc(b.data, func(id, _ string) bool {
		_, ok := factoryBindings[id]
		return !ok
	})
}

// ResetOne resets one key binding by ID to the factory default.
func (b *KeyBindings) ResetOne(id string) {
	if b.data != nil {
		delete(b.data, id)
	}
}
