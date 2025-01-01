// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"cmp"
	"context"
	"io/fs"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
)

var (
	factoryBindings              = make(map[string]*Binding)
	_               json.Omitter = &KeyBindings{}
)

// KeyBindings holds a set of key bindings.
type KeyBindings struct {
	data map[string]unison.KeyBinding
}

// Binding holds a single key binding.
type Binding struct {
	ID         string
	KeyBinding unison.KeyBinding
	Action     *unison.Action
}

// RegisterKeyBinding register a keybinding.
func RegisterKeyBinding(id string, action *unison.Action) {
	if _, exists := factoryBindings[id]; exists {
		return
	}
	factoryBindings[id] = &Binding{
		ID:         id,
		KeyBinding: action.KeyBinding,
		Action:     action,
	}
}

// CurrentBindings returns a sorted list with the current bindings.
func CurrentBindings() []*Binding {
	list := make([]*Binding, 0, len(factoryBindings))
	for _, v := range factoryBindings {
		list = append(list, &Binding{
			ID:         v.ID,
			KeyBinding: v.Action.KeyBinding,
			Action:     v.Action,
		})
	}
	slices.SortFunc(list, func(a, b *Binding) int {
		result := txt.NaturalCmp(a.Action.Title, b.Action.Title, true)
		if result == 0 {
			result = cmp.Compare(a.ID, b.ID)
		}
		return result
	})
	return list
}

// NewKeyBindingsFromFS creates a new set of key bindings from a file. Any missing values will be filled in with
// defaults.
func NewKeyBindingsFromFS(fileSystem fs.FS, filePath string) (*KeyBindings, error) {
	var b KeyBindings
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// ShouldOmit implements json.Omitter.
func (b *KeyBindings) ShouldOmit() bool {
	for k, v := range b.data {
		if info, ok := factoryBindings[k]; ok && v != info.KeyBinding {
			return false
		}
	}
	return true
}

// Save writes the Fonts to the file as JSON.
func (b *KeyBindings) Save(filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, b)
}

// MarshalJSON implements json.Marshaler.
func (b *KeyBindings) MarshalJSON() ([]byte, error) {
	data := make(map[string]unison.KeyBinding, len(b.data))
	for k, v := range b.data {
		if info, ok := factoryBindings[k]; ok && info.KeyBinding != v {
			data[k] = v
		}
	}
	return json.Marshal(&data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *KeyBindings) UnmarshalJSON(data []byte) error {
	m := make(map[string]unison.KeyBinding, len(factoryBindings))
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	b.data = m
	return nil
}

// MakeCurrent applies these key bindings to the current key bindings set.
func (b *KeyBindings) MakeCurrent() {
	var actions []*unison.Action
	for k, v := range factoryBindings {
		current, ok := b.data[k]
		if !ok {
			current = v.KeyBinding
		}
		if v.Action.KeyBinding != current {
			v.Action.KeyBinding = current
			actions = append(actions, v.Action)
		}
	}
	if len(actions) != 0 {
		factory := unison.DefaultMenuFactory()
		for _, w := range unison.Windows() {
			if bar := factory.BarForWindowNoCreate(w); !toolbox.IsNil(bar) {
				for _, a := range actions {
					if item := bar.Item(a.ID); item != nil {
						item.SetKeyBinding(a.KeyBinding)
					}
				}
				if factory.BarIsPerWindow() {
					break
				}
			}
		}
	}
}

// Current returns the binding for the given ID.
func (b *KeyBindings) Current(id string) unison.KeyBinding {
	if f, ok := factoryBindings[id]; ok {
		if c, ok2 := b.data[id]; ok2 {
			return c
		}
		return f.KeyBinding
	}
	return unison.KeyBinding{}
}

// Set the binding for the given ID.
func (b *KeyBindings) Set(id string, binding unison.KeyBinding) {
	if f, ok := factoryBindings[id]; ok {
		if b.data == nil {
			b.data = make(map[string]unison.KeyBinding, len(factoryBindings))
		}
		if f.KeyBinding != binding {
			b.data[id] = binding
		} else {
			delete(b.data, id)
		}
	}
}

// Reset to factory defaults.
func (b *KeyBindings) Reset() {
	b.data = nil
}

// ResetOne resets one font by ID to factory defaults.
func (b *KeyBindings) ResetOne(id string) {
	if b.data != nil {
		delete(b.data, id)
	}
}
