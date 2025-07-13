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
