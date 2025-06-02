package gurps

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/json"
)

type scriptWeight struct {
	Pound    fxp.WeightUnit `json:"pound"`
	PoundAlt fxp.WeightUnit `json:"poundAlt"`
	Ounce    fxp.WeightUnit `json:"ounce"`
	Ton      fxp.WeightUnit `json:"ton"`
	TonAlt   fxp.WeightUnit `json:"tonAlt"`
	Kilogram fxp.WeightUnit `json:"kilogram"`
	Gram     fxp.WeightUnit `json:"gram"`
}

func newScriptWeight() *scriptWeight {
	return &scriptWeight{
		Pound:    fxp.Pound,
		PoundAlt: fxp.PoundAlt,
		Ounce:    fxp.Ounce,
		Ton:      fxp.Ton,
		TonAlt:   fxp.TonAlt,
		Kilogram: fxp.Kilogram,
		Gram:     fxp.Gram,
	}
}

func (l *scriptWeight) UnitsFromString(value string) fxp.WeightUnit {
	return fxp.ExtractWeightUnit(value)
}

func (l *scriptWeight) FromString(value string, defaultUnits fxp.WeightUnit) fxp.Weight {
	return fxp.WeightFromStringForced(value, defaultUnits)
}

func (l *scriptWeight) FromInteger(value int, units fxp.WeightUnit) fxp.Weight {
	return fxp.WeightFromInteger(value, units)
}

func (l *scriptWeight) FromFloat(value float64, units fxp.WeightUnit) fxp.Weight {
	return fxp.WeightFromFixed(fxp.From(value), units)
}

func (l *scriptWeight) FromFixed(value fxp.Int, units fxp.WeightUnit) fxp.Weight {
	return fxp.WeightFromFixed(value, units)
}

func (l *scriptWeight) AsFixedPounds(value fxp.Weight) fxp.Int {
	return fxp.Int(value)
}

func (l *scriptWeight) String() string {
	data, err := json.Marshal(l)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
