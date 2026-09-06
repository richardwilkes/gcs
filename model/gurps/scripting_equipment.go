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
	"github.com/dop251/goja"
	"github.com/richardwilkes/gcs/v5/model/fxp"
)

func deferredNewScriptEquipment(item *Equipment) ScriptSelfProvider {
	return deferredScriptSelf(item, newScriptEquipment)
}

func newScriptEquipment(r *goja.Runtime, item *Equipment) *goja.Object {
	m := make(map[string]func() goja.Value)
	addScriptNodeIdentity(r, m, item, item.Tags, newScriptEquipment)
	m["name"] = func() goja.Value { return r.ToValue(item.NameWithReplacements()) }
	m["techLevel"] = func() goja.Value { return r.ToValue(item.TechLevel) }
	m["legalityClass"] = func() goja.Value { return r.ToValue(item.LegalityClass) }
	m["quantity"] = func() goja.Value { return r.ToValue(item.Quantity.AsFloat[float64]()) }
	m["level"] = func() goja.Value { return r.ToValue(item.Level.AsFloat[float64]()) }
	m["uses"] = func() goja.Value { return r.ToValue(item.ResolvedUses()) }
	m["maxUses"] = func() goja.Value { return r.ToValue(item.ResolvedMaxUses()) }
	m["value"] = func() goja.Value {
		return r.ToValue(item.AdjustedValue().AsFloat[float64]())
	}
	m["extendedValue"] = func() goja.Value {
		return r.ToValue(item.ExtendedValue().AsFloat[float64]())
	}
	m["weight"] = func() goja.Value {
		return r.ToValue(fxp.Int(item.AdjustedWeight(false, fxp.Pound)).AsFloat[float64]())
	}
	m["extendedWeight"] = func() goja.Value {
		return r.ToValue(fxp.Int(item.ExtendedWeight(false, fxp.Pound)).AsFloat[float64]())
	}
	m["weightIgnoredForSkills"] = func() goja.Value { return r.ToValue(item.WeightIgnoredForSkills) }
	// To a script, "equipped" means "affecting the character", so an item in the other equipment list is never
	// equipped: the character collects nothing from that list, and the equipped flag is meaningless there, since
	// nothing clears it when an item is created in or moved to it.
	m["equipped"] = func() goja.Value { return r.ToValue(item.IsCarried() && item.ReallyEquipped()) }
	m["switchedOn"] = func() goja.Value { return r.ToValue(item.SwitchedOn) }
	m["notes"] = scriptNotes(r, item)
	m["find"] = scriptNameTagFinder(r, func(name, tag string) goja.Value {
		return findScriptEquipment(r, name, tag, item.Children...)
	})
	addScriptWeapons(r, m, func() []*Weapon { return item.Weapons })
	addScriptActiveModifiers(r, m, item.ActiveModifierFor, func() []*EquipmentModifier { return item.Modifiers },
		newScriptEquipmentModifier)
	if item.Container() {
		m["children"] = func() goja.Value { return scriptObjects(r, item.Children, hasQuantity, newScriptEquipment) }
	}
	return r.NewDynamicObject(NewScriptObject(r, m))
}

// hasQuantity reports whether the item has a positive quantity, which is what makes it present as far as a script is
// concerned.
func hasQuantity(item *Equipment) bool {
	return item.Quantity > 0
}

// hasQuantityThroughAncestors reports whether the item and every container it sits in have a positive quantity.
// Traverse descends into a container regardless of its quantity, so a search has to check the chain itself.
func hasQuantityThroughAncestors(item *Equipment) bool {
	for ; item != nil; item = item.parent {
		if !hasQuantity(item) {
			return false
		}
	}
	return true
}

func findScriptEquipment(r *goja.Runtime, name, tag string, topLevelItems ...*Equipment) goja.Value {
	return findScriptNodes(r, name, tag, scriptNodeKind[*Equipment]{
		ctor:   newScriptEquipment,
		nameOf: (*Equipment).NameWithReplacements,
		tagsOf: func(item *Equipment) []string { return item.Tags },
	}, hasQuantityThroughAncestors, topLevelItems...)
}
