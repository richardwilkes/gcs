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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
)

func deferredNewScriptTraitModifier(mod *TraitModifier) ScriptSelfProvider {
	if mod == nil {
		return ScriptSelfProvider{}
	}
	return ScriptSelfProvider{
		ID:       string(mod.TID),
		Provider: func(r *goja.Runtime) any { return newScriptTraitModifier(r, mod) },
	}
}

func newScriptTraitModifier(r *goja.Runtime, mod *TraitModifier) *goja.Object {
	m := make(map[string]func() goja.Value)
	m["id"] = func() goja.Value { return r.ToValue(mod.TID) }
	m["attachedTo"] = func() goja.Value {
		if mod.trait == nil {
			return goja.Undefined()
		}
		return newScriptTrait(r, mod.trait)
	}
	m["name"] = func() goja.Value { return r.ToValue(mod.NameWithReplacements()) }
	m["level"] = func() goja.Value {
		if mod.IsLeveled() {
			return r.ToValue(fxp.AsFloat[float64](mod.RawCurrentLevel()))
		}
		return goja.Undefined()
	}
	m["tags"] = func() goja.Value { return r.ToValue(slices.Clone(mod.Tags)) }
	m["notes"] = func() goja.Value {
		return r.ToValue(mod.SecondaryText(func(_ display.Option) bool { return true }))
	}
	return r.NewDynamicObject(NewScriptObject(r, m))
}
