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
	"slices"
	"strings"

	"github.com/dop251/goja"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
)

func deferredNewScriptTrait(trait *Trait) ScriptSelfProvider {
	if trait == nil {
		return ScriptSelfProvider{}
	}
	return ScriptSelfProvider{
		ID:       string(trait.TID),
		Provider: func(r *goja.Runtime) any { return newScriptTrait(r, trait) },
	}
}

func newScriptTrait(r *goja.Runtime, trait *Trait) *goja.Object {
	m := make(map[string]func() goja.Value)
	m["id"] = func() goja.Value { return r.ToValue(trait.TID) }
	m["parentID"] = func() goja.Value {
		if trait.parent == nil {
			return goja.Undefined()
		}
		return r.ToValue(trait.parent.TID)
	}
	m["parent"] = func() goja.Value {
		if trait.parent == nil {
			return goja.Undefined()
		}
		return newScriptTrait(r, trait.parent)
	}
	m["name"] = func() goja.Value { return r.ToValue(trait.NameWithReplacements()) }
	m["notes"] = func() goja.Value {
		return r.ToValue(trait.SecondaryText(func(_ display.Option) bool { return true }))
	}
	m["tags"] = func() goja.Value { return r.ToValue(slices.Clone(trait.Tags)) }
	m["container"] = func() goja.Value { return r.ToValue(trait.Container()) }
	if trait.Container() {
		m["kind"] = func() goja.Value { return r.ToValue(strings.ReplaceAll(trait.ContainerType.Key(), "_", " ")) }
		m["children"] = func() goja.Value {
			children := make([]*goja.Object, 0, len(trait.Children))
			for _, child := range trait.Children {
				if child.Enabled() {
					children = append(children, newScriptTrait(r, child))
				}
			}
			return r.ToValue(children)
		}
		m["find"] = func() goja.Value {
			return r.ToValue(func(call goja.FunctionCall) goja.Value {
				name := callArgAsString(call, 0)
				tag := callArgAsString(call, 1)
				return findScriptTraits(r, name, tag, trait.Children...)
			})
		}
	} else if trait.CanLevel {
		m["level"] = func() goja.Value {
			return r.ToValue(fxp.AsFloat[float64](trait.Levels))
		}
	}
	m["activeModifierNamed"] = func() goja.Value {
		return r.ToValue(func(call goja.FunctionCall) goja.Value {
			name := callArgAsString(call, 0)
			mod := trait.ActiveModifierFor(name)
			if mod == nil {
				return goja.Null()
			}
			return newScriptTraitModifier(r, mod)
		})
	}
	m["activeModifiers"] = func() goja.Value {
		mods := make([]*goja.Object, 0, len(trait.Modifiers))
		Traverse(func(mod *TraitModifier) bool {
			mods = append(mods, newScriptTraitModifier(r, mod))
			return false
		}, true, true, trait.Modifiers...)
		return r.ToValue(mods)
	}
	return r.NewDynamicObject(NewScriptObject(r, m))
}

func findScriptTraits(r *goja.Runtime, name, tag string, topLevelTraits ...*Trait) goja.Value {
	var traits []*goja.Object
	Traverse(func(trait *Trait) bool {
		if (name == "" || strings.EqualFold(trait.NameWithReplacements(), name)) && matchTag(tag, trait.Tags) {
			traits = append(traits, newScriptTrait(r, trait))
		}
		return false
	}, true, false, topLevelTraits...)
	return r.ToValue(traits)
}
