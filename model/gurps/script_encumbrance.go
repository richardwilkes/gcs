package gurps

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/json"
)

type scriptEncumbrance struct {
	LevelName                string     `json:"levelName"`
	Level                    int        `json:"level"`
	LevelForSkills           int        `json:"levelForSkills"`
	MoveFactor               fxp.Int    `json:"moveFactor"`
	WeightCarried            fxp.Weight `json:"weightCarried"`
	MaximumCarry             fxp.Weight `json:"maximumCarry"`
	BasicLift                fxp.Weight `json:"basicLift"`
	OneHandedLift            fxp.Weight `json:"oneHandedLift"`
	TwoHandedLift            fxp.Weight `json:"twoHandedLift"`
	ShoveAndKnockOver        fxp.Weight `json:"shoveAndKnockOver"`
	RunningShoveAndKnockOver fxp.Weight `json:"runningShoveAndKnockOver"`
	CarryOnBack              fxp.Weight `json:"carryOnBack"`
	ShiftSlightly            fxp.Weight `json:"shiftSlightly"`
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
		MoveFactor:               fxp.One - fxp.From(int(level)).Mul(fxp.Two).Div(fxp.Ten),
		WeightCarried:            entity.WeightCarried(false),
		MaximumCarry:             entity.MaximumCarry(encumbrance.ExtraHeavy),
		BasicLift:                entity.BasicLift(),
		OneHandedLift:            entity.OneHandedLift(),
		TwoHandedLift:            entity.TwoHandedLift(),
		ShoveAndKnockOver:        entity.ShoveAndKnockOver(),
		RunningShoveAndKnockOver: entity.RunningShoveAndKnockOver(),
		CarryOnBack:              entity.CarryOnBack(),
		ShiftSlightly:            entity.ShiftSlightly(),
	}
}

func (e scriptEncumbrance) String() string {
	data, err := json.Marshal(e)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
