// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"cmp"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
)

// keyBindingEntries maps the ID of each registered key binding to its action and factory default.
var keyBindingEntries = make(map[string]*keyBindingEntry)

// keyBindingEntry holds a single key binding.
type keyBindingEntry struct {
	ID         string
	KeyBinding unison.KeyBinding
	Action     *unison.Action
}

// registerKeyBinding registers an action whose key binding the user may change, with the binding it has now as the
// factory default. A second registration of the same ID is ignored.
func registerKeyBinding(id string, action *unison.Action) {
	if _, exists := keyBindingEntries[id]; exists {
		return
	}
	keyBindingEntries[id] = &keyBindingEntry{
		ID:         id,
		KeyBinding: action.KeyBinding,
		Action:     action,
	}
	gurps.RegisterKeyBinding(id, action.KeyBinding.Key())
}

// currentKeyBindings returns a sorted list with the current bindings.
func currentKeyBindings() []*keyBindingEntry {
	list := make([]*keyBindingEntry, 0, len(keyBindingEntries))
	for _, v := range keyBindingEntries {
		list = append(list, &keyBindingEntry{
			ID:         v.ID,
			KeyBinding: v.Action.KeyBinding,
			Action:     v.Action,
		})
	}
	slices.SortFunc(list, func(a, b *keyBindingEntry) int {
		result := xstrings.NaturalCmp(a.Action.Title, b.Action.Title, true)
		if result == 0 {
			result = cmp.Compare(a.ID, b.ID)
		}
		return result
	})
	return list
}

// applyKeyBindings applies the key bindings to their actions and to the menu items that show them. Each binding is also
// stored back in its canonical text form, since the model compares bindings by their text and would otherwise take a
// factory default that was spelled differently for a change, and the bindings for IDs no action registered are dropped,
// since nothing could show or reset them.
func applyKeyBindings(b *gurps.KeyBindings) {
	b.Prune()
	var actions []*unison.Action
	for id, v := range keyBindingEntries {
		current := unison.KeyBindingFromKey(b.Current(id))
		b.Set(id, current.Key())
		if v.Action.KeyBinding != current {
			v.Action.KeyBinding = current
			actions = append(actions, v.Action)
		}
	}
	if len(actions) != 0 {
		factory := unison.DefaultMenuFactory()
		for _, w := range unison.Windows() {
			if bar := factory.BarForWindowNoCreate(w); !xreflect.IsNil(bar) {
				for _, a := range actions {
					if item := bar.Item(a.ID); item != nil {
						item.SetKeyBinding(a.KeyBinding)
					}
				}
				if !factory.BarIsPerWindow() {
					break
				}
			}
		}
	}
}
