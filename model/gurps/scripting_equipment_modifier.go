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
	"slices"

	"github.com/dop251/goja"
)

func deferredNewScriptEquipmentModifier(mod *EquipmentModifier) ScriptSelfProvider {
	return deferredScriptSelf(mod, newScriptEquipmentModifier)
}

func newScriptEquipmentModifier(r *goja.Runtime, mod *EquipmentModifier) *goja.Object {
	m := make(map[string]func() goja.Value)
	m["id"] = func() goja.Value { return r.ToValue(string(mod.TID)) }
	m["attachedTo"] = func() goja.Value {
		if mod.equipment == nil {
			return goja.Undefined()
		}
		return newScriptEquipment(r, mod.equipment)
	}
	m["name"] = func() goja.Value { return r.ToValue(mod.NameWithReplacements()) }
	m["techLevel"] = func() goja.Value { return r.ToValue(mod.TechLevel) }
	m["tags"] = func() goja.Value { return r.ToValue(slices.Clone(mod.Tags)) }
	m["notes"] = scriptNotes(r, mod)
	return r.NewDynamicObject(NewScriptObject(r, m))
}
