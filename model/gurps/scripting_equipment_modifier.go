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

	"github.com/dop251/goja"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
)

func deferredNewScriptEquipmentModifier(mod *EquipmentModifier) ScriptSelfProvider {
	if mod == nil {
		return ScriptSelfProvider{}
	}
	return ScriptSelfProvider{
		ID:       string(mod.TID),
		Provider: func(r *goja.Runtime) any { return newScriptEquipmentModifier(r, mod) },
	}
}

func newScriptEquipmentModifier(r *goja.Runtime, mod *EquipmentModifier) *goja.Object {
	m := make(map[string]func() goja.Value)
	m["id"] = func() goja.Value { return r.ToValue(mod.TID) }
	m["attachedTo"] = func() goja.Value {
		if mod.equipment == nil {
			return goja.Undefined()
		}
		return newScriptEquipment(r, mod.equipment)
	}
	m["name"] = func() goja.Value { return r.ToValue(mod.NameWithReplacements()) }
	m["techLevel"] = func() goja.Value { return r.ToValue(mod.TechLevel) }
	m["tags"] = func() goja.Value { return r.ToValue(slices.Clone(mod.Tags)) }
	m["notes"] = func() goja.Value {
		return r.ToValue(mod.SecondaryText(func(_ display.Option) bool { return true }))
	}
	return r.NewDynamicObject(NewScriptObject(r, m))
}
