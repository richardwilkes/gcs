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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
)

func newScriptEncumbrance(r *goja.Runtime, entity *Entity) *goja.Object {
	m := make(map[string]func() goja.Value)
	m["levelName"] = func() goja.Value {
		return r.ToValue(strings.ReplaceAll(entity.EncumbranceLevel(false).Key(), "_", " "))
	}
	m["level"] = func() goja.Value { return r.ToValue(int(entity.EncumbranceLevel(false))) }
	m["levelForSkills"] = func() goja.Value { return r.ToValue(int(entity.EncumbranceLevel(true))) }
	m["moveFactor"] = func() goja.Value {
		level := entity.EncumbranceLevel(false)
		return r.ToValue(fxp.AsFloat[float64](fxp.One - fxp.FromInteger(int(level)).Mul(fxp.Two).Div(fxp.Ten)))
	}
	m["weightCarried"] = func() goja.Value {
		return r.ToValue(fxp.AsFloat[float64](fxp.Int(entity.WeightCarried(false))))
	}
	m["maximumCarry"] = func() goja.Value {
		return r.ToValue(fxp.AsFloat[float64](fxp.Int(entity.MaximumCarry(encumbrance.ExtraHeavy))))
	}
	m["basicLift"] = func() goja.Value {
		return r.ToValue(fxp.AsFloat[float64](fxp.Int(entity.BasicLift())))
	}
	m["oneHandedLift"] = func() goja.Value {
		return r.ToValue(fxp.AsFloat[float64](fxp.Int(entity.OneHandedLift())))
	}
	m["twoHandedLift"] = func() goja.Value {
		return r.ToValue(fxp.AsFloat[float64](fxp.Int(entity.TwoHandedLift())))
	}
	m["shoveAndKnockOver"] = func() goja.Value {
		return r.ToValue(fxp.AsFloat[float64](fxp.Int(entity.ShoveAndKnockOver())))
	}
	m["runningShoveAndKnockOver"] = func() goja.Value {
		return r.ToValue(fxp.AsFloat[float64](fxp.Int(entity.RunningShoveAndKnockOver())))
	}
	m["carryOnBack"] = func() goja.Value {
		return r.ToValue(fxp.AsFloat[float64](fxp.Int(entity.CarryOnBack())))
	}
	m["shiftSlightly"] = func() goja.Value {
		return r.ToValue(fxp.AsFloat[float64](fxp.Int(entity.ShiftSlightly())))
	}
	return r.NewDynamicObject(NewScriptObject(r, m))
}
