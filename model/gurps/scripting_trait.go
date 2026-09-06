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
	"strings"

	"github.com/dop251/goja"
)

func deferredNewScriptTrait(trait *Trait) ScriptSelfProvider {
	return deferredScriptSelf(trait, newScriptTrait)
}

func newScriptTrait(r *goja.Runtime, trait *Trait) *goja.Object {
	m := make(map[string]func() goja.Value)
	addScriptNodeIdentity(r, m, trait, trait.Tags, newScriptTrait)
	m["name"] = func() goja.Value { return r.ToValue(trait.NameWithReplacements()) }
	m["notes"] = scriptNotes(r, trait)
	m["switchedOn"] = func() goja.Value { return r.ToValue(trait.SwitchedOn) }
	m["points"] = func() goja.Value { return r.ToValue(trait.AdjustedPoints().AsFloat[float64]()) }
	m["selfControl"] = func() goja.Value { return r.ToValue(trait.ResolvedSelfControl(nil).Number()) }
	m["selfControlAdjustment"] = func() goja.Value {
		return r.ToValue(trait.ResolvedSelfControlAdjustment(nil).Key())
	}
	m["frequency"] = func() goja.Value { return r.ToValue(trait.ResolvedFrequency(nil).Number()) }
	if trait.Container() {
		m["kind"] = func() goja.Value { return r.ToValue(strings.ReplaceAll(trait.ContainerType.Key(), "_", " ")) }
		m["children"] = func() goja.Value { return scriptObjects(r, trait.Children, (*Trait).Enabled, newScriptTrait) }
		m["find"] = scriptNameTagFinder(r, func(name, tag string) goja.Value {
			return findScriptTraits(r, name, tag, trait.Children...)
		})
	} else {
		if trait.CanLevel {
			m["level"] = func() goja.Value {
				return r.ToValue(trait.CurrentLevel().AsFloat[float64]())
			}
		}
		addScriptWeapons(r, m, func() []*Weapon { return trait.Weapons })
	}
	addScriptActiveModifiers(r, m, trait.ActiveModifierFor, func() []*TraitModifier { return trait.Modifiers },
		newScriptTraitModifier)
	return r.NewDynamicObject(NewScriptObject(r, m))
}

func findScriptTraits(r *goja.Runtime, name, tag string, topLevelTraits ...*Trait) goja.Value {
	return findScriptNodes(r, name, tag, scriptNodeKind[*Trait]{
		ctor:   newScriptTrait,
		nameOf: (*Trait).NameWithReplacements,
		tagsOf: func(trait *Trait) []string { return trait.Tags },
	}, nil, topLevelTraits...)
}
