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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
)

type scriptEncumbrance struct {
	LevelName                string  `json:"levelName"`
	Level                    int     `json:"level"`
	LevelForSkills           int     `json:"levelForSkills"`
	MoveFactor               float64 `json:"moveFactor"`
	WeightCarried            float64 `json:"weightCarried"`
	MaximumCarry             float64 `json:"maximumCarry"`
	BasicLift                float64 `json:"basicLift"`
	OneHandedLift            float64 `json:"oneHandedLift"`
	TwoHandedLift            float64 `json:"twoHandedLift"`
	ShoveAndKnockOver        float64 `json:"shoveAndKnockOver"`
	RunningShoveAndKnockOver float64 `json:"runningShoveAndKnockOver"`
	CarryOnBack              float64 `json:"carryOnBack"`
	ShiftSlightly            float64 `json:"shiftSlightly"`
}

func newScriptEncumbrance(entity *Entity) scriptEncumbrance {
	if entity == nil {
		return scriptEncumbrance{}
	}
	level := entity.EncumbranceLevel(false)
	return scriptEncumbrance{
		LevelName:                strings.ReplaceAll(level.Key(), "_", " "),
		Level:                    int(level),
		LevelForSkills:           int(entity.EncumbranceLevel(true)),
		MoveFactor:               fxp.AsFloat[float64](fxp.One - fxp.FromInteger(int(level)).Mul(fxp.Two).Div(fxp.Ten)),
		WeightCarried:            fxp.AsFloat[float64](fxp.Int(entity.WeightCarried(false))),
		MaximumCarry:             fxp.AsFloat[float64](fxp.Int(entity.MaximumCarry(encumbrance.ExtraHeavy))),
		BasicLift:                fxp.AsFloat[float64](fxp.Int(entity.BasicLift())),
		OneHandedLift:            fxp.AsFloat[float64](fxp.Int(entity.OneHandedLift())),
		TwoHandedLift:            fxp.AsFloat[float64](fxp.Int(entity.TwoHandedLift())),
		ShoveAndKnockOver:        fxp.AsFloat[float64](fxp.Int(entity.ShoveAndKnockOver())),
		RunningShoveAndKnockOver: fxp.AsFloat[float64](fxp.Int(entity.RunningShoveAndKnockOver())),
		CarryOnBack:              fxp.AsFloat[float64](fxp.Int(entity.CarryOnBack())),
		ShiftSlightly:            fxp.AsFloat[float64](fxp.Int(entity.ShiftSlightly())),
	}
}
