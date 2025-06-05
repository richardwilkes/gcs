package gurps

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

type scriptMeasurement struct{}

func (s scriptMeasurement) FormatLength(inches float64, toUnits string) string {
	return fxp.ExtractLengthUnit(toUnits).Format(fxp.Length(fxp.From(inches)))
}

func (s scriptMeasurement) LengthToInches(value float64, fromUnits string) float64 {
	return fxp.As[float64](fxp.ExtractLengthUnit(fromUnits).ToInches(fxp.From(value)))
}

func (s scriptMeasurement) StringLengthToInches(str, defaultUnits string) float64 {
	length, err := fxp.LengthFromString(str, fxp.ExtractLengthUnit(defaultUnits))
	if err != nil {
		return 0
	}
	return fxp.As[float64](fxp.Int(length))
}

func (s scriptMeasurement) FormatWeight(pounds float64, toUnits string) string {
	return fxp.ExtractWeightUnit(toUnits).Format(fxp.Weight(fxp.From(pounds)))
}

func (s scriptMeasurement) WeightToPounds(value float64, fromUnits string) float64 {
	return fxp.As[float64](fxp.ExtractWeightUnit(fromUnits).ToPounds(fxp.From(value)))
}

func (s scriptMeasurement) StringWeightToPounds(str, defaultUnits string) float64 {
	weight, err := fxp.WeightFromString(str, fxp.ExtractWeightUnit(defaultUnits))
	if err != nil {
		return 0
	}
	return fxp.As[float64](fxp.Int(weight))
}

func (s scriptMeasurement) RangeModifier(yards float64) int {
	return -ssrtInchesToValue(fxp.Yard.ToInches(fxp.From(yards)), false)
}

func (s scriptMeasurement) SizeModifier(yards float64) int {
	return ssrtInchesToValue(fxp.Yard.ToInches(fxp.From(yards)), true)
}

func (s scriptMeasurement) Modifier(length float64, units string, forSize bool) int {
	l, err := fxp.LengthFromString(fmt.Sprintf("%v %s", length, units), fxp.Yard)
	if err != nil {
		return 0
	}
	result := ssrtInchesToValue(fxp.Int(l), forSize)
	if !forSize {
		result = -result
	}
	return result
}

func (s scriptMeasurement) ModifierToYards(ssrtValue int) float64 {
	return fxp.As[float64](ssrtToYards(ssrtValue))
}

func ssrtInchesToValue(inches fxp.Int, allowNegative bool) int {
	yards := inches.Div(fxp.ThirtySix)
	if allowNegative {
		switch {
		case inches <= fxp.One.Div(fxp.Five):
			return -15
		case inches <= fxp.One.Div(fxp.Three):
			return -14
		case inches <= fxp.Half:
			return -13
		case inches <= fxp.Two.Div(fxp.Three):
			return -12
		case inches <= fxp.One:
			return -11
		case inches <= fxp.OneAndAHalf:
			return -10
		case inches <= fxp.Two:
			return -9
		case inches <= fxp.Three:
			return -8
		case inches <= fxp.Five:
			return -7
		case inches <= fxp.Eight:
			return -6
		}
		feet := inches.Div(fxp.Twelve)
		switch {
		case feet <= fxp.One:
			return -5
		case feet <= fxp.OneAndAHalf:
			return -4
		case feet <= fxp.Two:
			return -3
		case yards <= fxp.One:
			return -2
		case yards <= fxp.OneAndAHalf:
			return -1
		}
	}
	if yards <= fxp.Two {
		return 0
	}
	amt := 0
	for yards > fxp.Ten {
		yards = yards.Div(fxp.Ten)
		amt += 6
	}
	switch {
	case yards > fxp.Seven:
		return amt + 4
	case yards > fxp.Five:
		return amt + 3
	case yards > fxp.Three:
		return amt + 2
	case yards > fxp.Two:
		return amt + 1
	case yards > fxp.OneAndAHalf:
		return amt
	default:
		return amt - 1
	}
}

func ssrtToYards(value int) fxp.Int {
	if value < -15 {
		value = -15
	}
	switch value {
	case -15:
		return fxp.One.Div(fxp.Five).Div(fxp.ThirtySix)
	case -14:
		return fxp.One.Div(fxp.Three).Div(fxp.ThirtySix)
	case -13:
		return fxp.Half.Div(fxp.ThirtySix)
	case -12:
		return fxp.Two.Div(fxp.Three).Div(fxp.ThirtySix)
	case -11:
		return fxp.One.Div(fxp.ThirtySix)
	case -10:
		return fxp.OneAndAHalf.Div(fxp.ThirtySix)
	case -9:
		return fxp.Two.Div(fxp.ThirtySix)
	case -8:
		return fxp.Three.Div(fxp.ThirtySix)
	case -7:
		return fxp.Five.Div(fxp.ThirtySix)
	case -6:
		return fxp.Eight.Div(fxp.ThirtySix)
	case -5:
		return fxp.One.Div(fxp.Three)
	case -4:
		return fxp.OneAndAHalf.Div(fxp.Three)
	case -3:
		return fxp.Two.Div(fxp.Three)
	case -2:
		return fxp.One
	case -1:
		return fxp.OneAndAHalf
	case 0:
		return fxp.Two
	case 1:
		return fxp.Three
	case 2:
		return fxp.Five
	case 3:
		return fxp.Seven
	}
	value -= 4
	multiplier := fxp.One
	for range value / 6 {
		multiplier = multiplier.Mul(fxp.Ten)
	}
	var v fxp.Int
	switch value % 6 {
	case 0:
		v = fxp.Ten
	case 1:
		v = fxp.Fifteen
	case 2:
		v = fxp.Twenty
	case 3:
		v = fxp.Thirty
	case 4:
		v = fxp.Fifty
	case 5:
		v = fxp.Seventy
	}
	return v.Mul(multiplier)
}
