package gurps

import "github.com/richardwilkes/gcs/v5/model/fxp"

type scriptLength struct {
	FeetAndInches fxp.LengthUnit
	Inch          fxp.LengthUnit
	Feet          fxp.LengthUnit
	Yard          fxp.LengthUnit
	Mile          fxp.LengthUnit
	Centimeter    fxp.LengthUnit
	Kilometer     fxp.LengthUnit
	Meter         fxp.LengthUnit
}

func newScriptLength() *scriptLength {
	return &scriptLength{
		FeetAndInches: fxp.FeetAndInches,
		Inch:          fxp.Inch,
		Feet:          fxp.Feet,
		Yard:          fxp.Yard,
		Mile:          fxp.Mile,
		Centimeter:    fxp.Centimeter,
		Kilometer:     fxp.Kilometer,
		Meter:         fxp.Meter,
	}
}

func (l *scriptLength) UnitsFromString(value string) fxp.LengthUnit {
	return fxp.ExtractLengthUnit(value)
}

func (l *scriptLength) FromString(value string, defaultUnits fxp.LengthUnit) fxp.Length {
	return fxp.LengthFromStringForced(value, defaultUnits)
}

func (l *scriptLength) FromInteger(value int, units fxp.LengthUnit) fxp.Length {
	return fxp.LengthFromInteger(value, units)
}

func (l *scriptLength) FromFixed(value fxp.Int, units fxp.LengthUnit) fxp.Length {
	return fxp.LengthFromInteger(value, units)
}

func (l *scriptLength) AsFixed(value fxp.Length) fxp.Int {
	return fxp.Int(value)
}
