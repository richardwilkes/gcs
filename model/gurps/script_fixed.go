package gurps

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/json"
)

type scriptFixed struct {
	MaxSafeMultiply fxp.Int `json:"maxSafeMultiply"`
	MaxValue        fxp.Int `json:"maxValue"`
	MinValue        fxp.Int `json:"minValue"`
	Zero            fxp.Int `json:"zero"`
	Tenth           fxp.Int `json:"tenth"`
	Eighth          fxp.Int `json:"eighth"`
	Quarter         fxp.Int `json:"quarter"`
	Half            fxp.Int `json:"half"`
	One             fxp.Int `json:"one"`
	Two             fxp.Int `json:"two"`
	Three           fxp.Int `json:"three"`
	Four            fxp.Int `json:"four"`
	Five            fxp.Int `json:"five"`
	Ten             fxp.Int `json:"ten"`
	Hundred         fxp.Int `json:"hundred"`
	Thousand        fxp.Int `json:"thousand"`
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

func (f *scriptFixed) AsInteger(value fxp.Int) int {
	return fxp.As[int](value)
}

func (f *scriptFixed) AsFloat(value fxp.Int) float64 {
	return fxp.As[float64](value)
}

func (f *scriptFixed) FromString(value string) fxp.Int {
	return fxp.FromStringForced(value)
}

func (f *scriptFixed) FromInteger(value int) fxp.Int {
	return fxp.From(value)
}

func (f *scriptFixed) FromFloat(value float64) fxp.Int {
	return fxp.From(value)
}

func (f *scriptFixed) ApplyRounding(value fxp.Int, roundDown bool) fxp.Int {
	return fxp.ApplyRounding(value, roundDown)
}

func (f *scriptFixed) String() string {
	data, err := json.Marshal(f)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
