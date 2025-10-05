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
	"strings"

	"github.com/dop251/goja"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/xreflect"
)

func newScriptWeapon(r *goja.Runtime, w *Weapon) *goja.Object {
	m := make(map[string]func() goja.Value)
	m["id"] = func() goja.Value { return r.ToValue(w.TID) }
	m["attachedTo"] = func() goja.Value {
		if xreflect.IsNil(w.Owner) {
			return goja.Undefined()
		}
		switch owner := w.Owner.(type) {
		case *Equipment:
			return newScriptEquipment(r, owner)
		case *Trait:
			return newScriptTrait(r, owner)
		}
		return goja.Undefined()
	}
	m["name"] = func() goja.Value { return r.ToValue(w.String()) }
	m["notes"] = func() goja.Value { return r.ToValue(w.Notes()) }
	m["usage"] = func() goja.Value { return r.ToValue(w.UsageWithReplacements()) }
	m["usageNotes"] = func() goja.Value { return r.ToValue(w.UsageNotesWithReplacements()) }
	m["level"] = func() goja.Value { return r.ToValue(fxp.AsInteger[int](w.SkillLevel(nil).Max(0))) }
	m["damage"] = func() goja.Value { return r.ToValue(w.Damage.ResolvedDamage(nil)) }
	m["strength"] = func() goja.Value { return r.ToValue(w.Strength.Resolve(w, nil).String()) }
	m["hidden"] = func() goja.Value { return r.ToValue(w.Hide) }
	if w.IsMelee() {
		m["kind"] = func() goja.Value { return r.ToValue("melee") }
		m["block"] = func() goja.Value { return r.ToValue(w.Block.Resolve(w, nil).String()) }
		m["parry"] = func() goja.Value { return r.ToValue(w.Parry.Resolve(w, nil).String()) }
		m["reach"] = func() goja.Value { return r.ToValue(w.Reach.Resolve(w, nil).String()) }
	} else {
		m["kind"] = func() goja.Value { return r.ToValue("ranged") }
		m["accuracy"] = func() goja.Value { return r.ToValue(w.Accuracy.Resolve(w, nil).String()) }
		m["bulk"] = func() goja.Value { return r.ToValue(w.Bulk.Resolve(w, nil).String()) }
		m["range"] = func() goja.Value { return r.ToValue(w.Range.Resolve(w, nil).String(w.musclePowerIsResolved())) }
		m["rateOfFire"] = func() goja.Value { return r.ToValue(w.RateOfFire.Resolve(w, nil).String()) }
		m["recoil"] = func() goja.Value { return r.ToValue(w.Recoil.Resolve(w, nil).String()) }
		m["shots"] = func() goja.Value { return r.ToValue(w.Shots.Resolve(w, nil).String()) }
	}
	return r.NewDynamicObject(NewScriptObject(r, m))
}

func matchWeapons(r *goja.Runtime, weapons []*Weapon, name, usage string, melee bool) goja.Value {
	var result []*goja.Object
	for _, w := range weapons {
		if melee == w.IsMelee() &&
			(name == "" || strings.EqualFold(w.String(), name)) &&
			(usage == "" || strings.EqualFold(w.UsageWithReplacements(), usage)) {
			result = append(result, newScriptWeapon(r, w))
		}
	}
	return r.ToValue(result)
}
