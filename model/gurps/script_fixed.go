package gurps

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
)

type scriptFixed struct {
	MaxSafeMultiply fxp.Int
	MaxValue        fxp.Int
	MinValue        fxp.Int
	Zero            fxp.Int
	Tenth           fxp.Int
	Eighth          fxp.Int
	Quarter         fxp.Int
	Half            fxp.Int
	One             fxp.Int
	Two             fxp.Int
	Three           fxp.Int
	Four            fxp.Int
	Five            fxp.Int
	Ten             fxp.Int
	Hundred         fxp.Int
	Thousand        fxp.Int
}

func newScriptFixed() *scriptFixed {
	return &scriptFixed{
		MaxSafeMultiply: fxp.MaxSafeMultiply,
		MaxValue:        fxp.Max,
		MinValue:        fxp.Min,
		Zero:            0,
		Tenth:           fxp.Tenth,
		Eighth:          fxp.Eighth,
		Quarter:         fxp.Quarter,
		Half:            fxp.Half,
		One:             fxp.One,
		Two:             fxp.Two,
		Three:           fxp.Three,
		Four:            fxp.Four,
		Five:            fxp.Five,
		Ten:             fxp.Ten,
		Hundred:         fxp.Hundred,
		Thousand:        fxp.Thousand,
	}
}

func (f *scriptFixed) AsInt(value fxp.Int) int {
	return fxp.As[int](value)
}

func (f *scriptFixed) AsFloat(value fxp.Int) float64 {
	return fxp.As[float64](value)
}

func (f *scriptFixed) FromString(value string) fxp.Int {
	return fxp.FromStringForced(value)
}

func (f *scriptFixed) FromInt(value int) fxp.Int {
	return fxp.From(value)
}

func (f *scriptFixed) FromFloat(value float64) fxp.Int {
	return fxp.From(value)
}

func (f *scriptFixed) ApplyRounding(value fxp.Int, roundDown bool) fxp.Int {
	return fxp.ApplyRounding(value, roundDown)
}
