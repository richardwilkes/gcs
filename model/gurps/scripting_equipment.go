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

func deferredNewScriptEquipment(item *Equipment) ScriptSelfProvider {
	if item == nil {
		return ScriptSelfProvider{}
	}
	return ScriptSelfProvider{
		ID:       string(item.TID),
		Provider: func(r *goja.Runtime) any { return newScriptEquipment(r, item) },
	}
}

func newScriptEquipment(r *goja.Runtime, item *Equipment) *goja.Object {
	m := make(map[string]func() goja.Value)
	m["id"] = func() goja.Value { return r.ToValue(item.TID) }
	m["parentID"] = func() goja.Value {
		if item.parent == nil {
			return goja.Undefined()
		}
		return r.ToValue(item.parent.TID)
	}
	m["parent"] = func() goja.Value {
		if item.parent == nil {
			return goja.Undefined()
		}
		return newScriptEquipment(r, item.parent)
	}
	m["name"] = func() goja.Value { return r.ToValue(item.NameWithReplacements()) }
	m["techLevel"] = func() goja.Value { return r.ToValue(item.TechLevel) }
	m["legalityClass"] = func() goja.Value { return r.ToValue(item.LegalityClass) }
	m["quantity"] = func() goja.Value { return r.ToValue(fxp.AsFloat[float64](item.Quantity)) }
	m["level"] = func() goja.Value { return r.ToValue(fxp.AsFloat[float64](item.Level)) }
	m["uses"] = func() goja.Value { return r.ToValue(item.Uses) }
	m["maxUses"] = func() goja.Value { return r.ToValue(item.MaxUses) }
	m["value"] = func() goja.Value {
		return r.ToValue(fxp.AsFloat[float64](item.AdjustedValue()))
	}
	m["extendedValue"] = func() goja.Value {
		return r.ToValue(fxp.AsFloat[float64](item.ExtendedValue()))
	}
	m["weight"] = func() goja.Value {
		return r.ToValue(fxp.AsFloat[float64](fxp.Int(item.AdjustedWeight(false, fxp.Pound))))
	}
	m["extendedWeight"] = func() goja.Value {
		return r.ToValue(fxp.AsFloat[float64](fxp.Int(item.ExtendedWeight(false, fxp.Pound))))
	}
	m["weightIgnoredForSkills"] = func() goja.Value { return r.ToValue(item.WeightIgnoredForSkills) }
	m["equipped"] = func() goja.Value { return r.ToValue(item.Equipped) }
	m["container"] = func() goja.Value { return r.ToValue(item.Container()) }
	m["notes"] = func() goja.Value {
		return r.ToValue(item.SecondaryText(func(_ display.Option) bool { return true }))
	}
	m["tags"] = func() goja.Value { return r.ToValue(slices.Clone(item.Tags)) }
	m["find"] = func() goja.Value {
		return r.ToValue(func(call goja.FunctionCall) goja.Value {
			name := callArgAsString(call, 0)
			tag := callArgAsString(call, 1)
			return findScriptEquipment(r, name, tag, item.Children...)
		})
	}
	m["weapons"] = func() goja.Value {
		weapons := make([]*goja.Object, 0, len(item.Weapons))
		for _, w := range item.Weapons {
			weapons = append(weapons, newScriptWeapon(r, w))
		}
		return r.ToValue(weapons)
	}
	m["findWeapons"] = func() goja.Value {
		return r.ToValue(func(call goja.FunctionCall) goja.Value {
			melee := call.Argument(0).ToBoolean()
			name := callArgAsString(call, 1)
			usage := callArgAsString(call, 2)
			return matchWeapons(r, item.Weapons, name, usage, melee)
		})
	}
	m["findActiveModifier"] = func() goja.Value {
		return r.ToValue(func(call goja.FunctionCall) goja.Value {
			name := callArgAsTrimmedString(call, 0)
			mod := item.ActiveModifierFor(name)
			if mod == nil {
				return goja.Null()
			}
			return newScriptEquipmentModifier(r, mod)
		})
	}
	m["activeModifiers"] = func() goja.Value {
		mods := make([]*goja.Object, 0, len(item.Modifiers))
		Traverse(func(mod *EquipmentModifier) bool {
			mods = append(mods, newScriptEquipmentModifier(r, mod))
			return false
		}, true, true, item.Modifiers...)
		return r.ToValue(mods)
	}
	if item.Container() {
		m["children"] = func() goja.Value {
			children := make([]*goja.Object, 0, len(item.Children))
			for _, child := range item.Children {
				if child.Quantity > 0 {
					children = append(children, newScriptEquipment(r, child))
				}
			}
			return r.ToValue(children)
		}
	}
	return r.NewDynamicObject(NewScriptObject(r, m))
}

func findScriptEquipment(r *goja.Runtime, name, tag string, topLevelItems ...*Equipment) goja.Value {
	var items []*goja.Object
	Traverse(func(item *Equipment) bool {
		if item.Quantity > 0 {
			parent := item.parent
			for parent != nil {
				if parent.Quantity <= 0 {
					return false
				}
				parent = parent.parent
			}
			if (name == "" || strings.EqualFold(item.NameWithReplacements(), name)) && matchTag(tag, item.Tags) {
				items = append(items, newScriptEquipment(r, item))
			}
		}
		return false
	}, true, false, topLevelItems...)
	return r.ToValue(items)
}

func matchTag(tag string, tags []string) bool {
	if tag == "" {
		return true
	}
	for _, t := range tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}
