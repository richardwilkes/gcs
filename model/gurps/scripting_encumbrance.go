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
		return r.ToValue((fxp.One - fxp.FromInteger(int(level)).Mul(fxp.Two).Div(fxp.Ten)).AsFloat[float64]())
	}
	m["weightCarried"] = func() goja.Value {
		return r.ToValue(fxp.Int(entity.WeightCarried(false)).AsFloat[float64]())
	}
	m["maximumCarry"] = func() goja.Value {
		return r.ToValue(fxp.Int(entity.MaximumCarry(encumbrance.ExtraHeavy)).AsFloat[float64]())
	}
	m["basicLift"] = func() goja.Value {
		return r.ToValue(fxp.Int(entity.BasicLift()).AsFloat[float64]())
	}
	m["oneHandedLift"] = func() goja.Value {
		return r.ToValue(fxp.Int(entity.OneHandedLift()).AsFloat[float64]())
	}
	m["twoHandedLift"] = func() goja.Value {
		return r.ToValue(fxp.Int(entity.TwoHandedLift()).AsFloat[float64]())
	}
	m["shoveAndKnockOver"] = func() goja.Value {
		return r.ToValue(fxp.Int(entity.ShoveAndKnockOver()).AsFloat[float64]())
	}
	m["runningShoveAndKnockOver"] = func() goja.Value {
		return r.ToValue(fxp.Int(entity.RunningShoveAndKnockOver()).AsFloat[float64]())
	}
	m["carryOnBack"] = func() goja.Value {
		return r.ToValue(fxp.Int(entity.CarryOnBack()).AsFloat[float64]())
	}
	m["shiftSlightly"] = func() goja.Value {
		return r.ToValue(fxp.Int(entity.ShiftSlightly()).AsFloat[float64]())
	}
	return r.NewDynamicObject(NewScriptObject(r, m))
}
