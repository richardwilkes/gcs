package gurps

import "github.com/richardwilkes/gcs/v5/model/fxp"

type scriptWeight struct {
	Pound    fxp.WeightUnit
	PoundAlt fxp.WeightUnit
	Ounce    fxp.WeightUnit
	Ton      fxp.WeightUnit
	TonAlt   fxp.WeightUnit
	Kilogram fxp.WeightUnit
	Gram     fxp.WeightUnit
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

func (l *scriptWeight) FromFixed(value fxp.Int, units fxp.WeightUnit) fxp.Weight {
	return fxp.WeightFromInteger(value, units)
}

func (l *scriptWeight) AsFixed(value fxp.Weight) fxp.Int {
	return fxp.Int(value)
}
