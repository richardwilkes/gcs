// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package fxp_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestDecimalPlaceDefaultIsAsNeeded pins the enum's zero value to AsNeeded, which is what makes a NumberFormat's zero
// value -- and therefore every sheet settings file written before the setting existed -- format exactly as before.
func TestDecimalPlaceDefaultIsAsNeeded(t *testing.T) {
	c := check.New(t)
	c.Equal(fxp.AsNeeded, fxp.DecimalPlace(0))
	c.Equal(fxp.AsNeeded, fxp.DefaultDecimalPlace)
	c.Equal(fxp.AsNeeded, fxp.ExtractDecimalPlace("bogus"))
	c.Equal(fxp.AsNeeded, fxp.DecimalPlace(200).EnsureValid())
	c.Equal(fxp.NumberFormat{}, fxp.NumberFormat{Places: fxp.DecimalPlace(200)}.EnsureValid())
}

// TestNumberFormatRound verifies rounding to each number of decimal places, including that it rounds half away from
// zero in both directions and that AsNeeded, FourPlaces and a Places out of range that has not been through EnsureValid
// leave the value alone. The last matters apart from Format, since LengthUnit.FormatWith calls Round directly for feet
// and inches.
func TestNumberFormatRound(t *testing.T) {
	c := check.New(t)
	for _, d := range []struct {
		places   fxp.DecimalPlace
		in       string
		expected string
	}{
		{places: fxp.AsNeeded, in: "7.5127", expected: "7.5127"},
		{places: fxp.FourPlaces, in: "7.5127", expected: "7.5127"},
		{places: fxp.ThreePlaces, in: "7.5127", expected: "7.513"},
		{places: fxp.TwoPlaces, in: "7.5127", expected: "7.51"},
		{places: fxp.OnePlace, in: "7.5127", expected: "7.5"},
		{places: fxp.ZeroPlaces, in: "7.5127", expected: "8"},
		{places: fxp.ZeroPlaces, in: "7.4999", expected: "7"},
		{places: fxp.ZeroPlaces, in: "7.5", expected: "8"},
		{places: fxp.TwoPlaces, in: "7.505", expected: "7.51"},
		{places: fxp.TwoPlaces, in: "7.5049", expected: "7.5"},
		{places: fxp.TwoPlaces, in: "-7.505", expected: "-7.51"},
		{places: fxp.TwoPlaces, in: "-7.5049", expected: "-7.5"},
		{places: fxp.ZeroPlaces, in: "-7.5", expected: "-8"},
		{places: fxp.ZeroPlaces, in: "-0.4999", expected: "0"},
		{places: fxp.TwoPlaces, in: "0", expected: "0"},
		{places: fxp.TwoPlaces, in: "1234.9999", expected: "1235"},
		{places: fxp.DecimalPlace(200), in: "1234.5678", expected: "1234.5678"},
		{places: fxp.LastDecimalPlace + 1, in: "1234.5678", expected: "1234.5678"},
	} {
		f := fxp.NumberFormat{Places: d.places}
		c.Equal(d.expected, f.Round(fxp.FromStringForced(d.in)).String(), "rounding %s to %s places (Places %d)", d.in,
			d.places.Key(), d.places)
	}
}

// TestNumberFormatRoundExtremes verifies that values whose rounded result would not be representable are returned
// unchanged rather than overflowing.
func TestNumberFormatRoundExtremes(t *testing.T) {
	c := check.New(t)
	for _, places := range fxp.DecimalPlaces {
		f := fxp.NumberFormat{Places: places}
		c.Equal(fxp.Max, f.Round(fxp.Max), "Max at %s places", places.Key())
		c.Equal(fxp.Min, f.Round(fxp.Min), "Min at %s places", places.Key())
	}
}

// TestNumberFormatFormat verifies the rendered text, including thousands separators and the optional padding of the
// fractional part with trailing zeros. A Places out of range that has not been through EnsureValid, as happens if a
// format reaches formatting before SheetSettings.EnsureValidity() has been applied, must behave exactly as AsNeeded
// does, never padding.
func TestNumberFormatFormat(t *testing.T) {
	c := check.New(t)
	for _, d := range []struct {
		format   fxp.NumberFormat
		in       string
		expected string
	}{
		{format: fxp.NumberFormat{}, in: "1234.5678", expected: "1,234.5678"},
		{format: fxp.NumberFormat{PadWithZeros: true}, in: "1234.5", expected: "1,234.5"},
		{format: fxp.NumberFormat{Places: fxp.TwoPlaces}, in: "1234.5678", expected: "1,234.57"},
		{format: fxp.NumberFormat{Places: fxp.TwoPlaces}, in: "7.5", expected: "7.5"},
		{format: fxp.NumberFormat{Places: fxp.TwoPlaces, PadWithZeros: true}, in: "7.5", expected: "7.50"},
		{format: fxp.NumberFormat{Places: fxp.TwoPlaces, PadWithZeros: true}, in: "7", expected: "7.00"},
		{format: fxp.NumberFormat{Places: fxp.TwoPlaces, PadWithZeros: true}, in: "-7", expected: "-7.00"},
		{format: fxp.NumberFormat{Places: fxp.TwoPlaces, PadWithZeros: true}, in: "0", expected: "0.00"},
		{format: fxp.NumberFormat{Places: fxp.TwoPlaces, PadWithZeros: true}, in: "1234.5678", expected: "1,234.57"},
		{format: fxp.NumberFormat{Places: fxp.ZeroPlaces, PadWithZeros: true}, in: "7.5", expected: "8"},
		{format: fxp.NumberFormat{Places: fxp.FourPlaces, PadWithZeros: true}, in: "7.5", expected: "7.5000"},
		{format: fxp.NumberFormat{Places: fxp.OnePlace, PadWithZeros: true}, in: "7", expected: "7.0"},
		{format: fxp.NumberFormat{Places: fxp.ThreePlaces, PadWithZeros: true}, in: "7.12", expected: "7.120"},
		{format: fxp.NumberFormat{Places: fxp.DecimalPlace(200)}, in: "1234.5678", expected: "1,234.5678"},
		{format: fxp.NumberFormat{Places: fxp.DecimalPlace(200)}, in: "7.5", expected: "7.5"},
		{format: fxp.NumberFormat{Places: fxp.DecimalPlace(200), PadWithZeros: true}, in: "1234.5678", expected: "1,234.5678"},
		{format: fxp.NumberFormat{Places: fxp.DecimalPlace(200), PadWithZeros: true}, in: "7.5", expected: "7.5"},
		{format: fxp.NumberFormat{Places: fxp.LastDecimalPlace + 1, PadWithZeros: true}, in: "1234.5678", expected: "1,234.5678"},
		{format: fxp.NumberFormat{Places: fxp.LastDecimalPlace + 1, PadWithZeros: true}, in: "7.5", expected: "7.5"},
	} {
		c.Equal(d.expected, d.format.Format(fxp.FromStringForced(d.in)), "formatting %s with %+v", d.in, d.format)
	}
}

// TestUnitFormatMatchesZeroNumberFormat verifies that the plain Format methods produce exactly what FormatWith does
// with a zero NumberFormat for every unit, which is what guarantees no change in behavior where no display preference
// has been expressed.
func TestUnitFormatMatchesZeroNumberFormat(t *testing.T) {
	c := check.New(t)
	for _, in := range []string{"0", "1", "-1", "7.5127", "-7.5127", "1234.5", "24.6952", "12345678.9"} {
		weight := fxp.Weight(fxp.FromStringForced(in))
		for _, unit := range fxp.WeightUnits {
			c.Equal(unit.Format(weight), unit.FormatWith(weight, fxp.NumberFormat{}), "%s lb as %s", in, unit.Key())
		}
		length := fxp.Length(fxp.FromStringForced(in))
		for _, unit := range fxp.LengthUnits {
			c.Equal(unit.Format(length), unit.FormatWith(length, fxp.NumberFormat{}), "%s in as %s", in, unit.Key())
		}
	}
}

// TestWeightUnitFormatWith verifies that the rounding is applied after conversion to the unit, so the requested number
// of decimal places is what is shown regardless of the unit chosen.
func TestWeightUnitFormatWith(t *testing.T) {
	c := check.New(t)
	two := fxp.NumberFormat{Places: fxp.TwoPlaces}
	twoPadded := fxp.NumberFormat{Places: fxp.TwoPlaces, PadWithZeros: true}
	whole := fxp.NumberFormat{Places: fxp.ZeroPlaces}
	w := fxp.Weight(fxp.FromStringForced("24.6952"))
	c.Equal("24.7 lb", fxp.Pound.FormatWith(w, two))
	c.Equal("24.70 lb", fxp.Pound.FormatWith(w, twoPadded))
	c.Equal("25 lb", fxp.Pound.FormatWith(w, whole))
	c.Equal("12.35 kg", fxp.Kilogram.FormatWith(w, two))
	c.Equal("12 kg", fxp.Kilogram.FormatWith(w, whole))
	c.Equal("395.12 oz", fxp.Ounce.FormatWith(w, two))
	c.Equal("12,347.6 g", fxp.Gram.FormatWith(w, two))
	c.Equal("0.01 tn", fxp.Ton.FormatWith(w, two))
	c.Equal("0.0123 tn", fxp.Ton.FormatWith(w, fxp.NumberFormat{}))
	c.Equal("7.50 lb", fxp.Pound.FormatWith(fxp.Weight(fxp.FromStringForced("7.5")), twoPadded))
}

// TestLengthUnitFormatWith verifies the rounding of lengths, in particular that for feet and inches the rounding is
// done on the total inches so that an inches remainder which rounds up to a whole foot carries into the feet, that
// padding applies to the inches part, and that a zero inches remainder is still omitted.
func TestLengthUnitFormatWith(t *testing.T) {
	c := check.New(t)
	two := fxp.NumberFormat{Places: fxp.TwoPlaces}
	twoPadded := fxp.NumberFormat{Places: fxp.TwoPlaces, PadWithZeros: true}
	whole := fxp.NumberFormat{Places: fxp.ZeroPlaces}
	for _, d := range []struct {
		unit     fxp.LengthUnit
		format   fxp.NumberFormat
		inches   string
		expected string
	}{
		{unit: fxp.FeetAndInches, format: whole, inches: "71.6", expected: "6'"},
		{unit: fxp.FeetAndInches, format: whole, inches: "71.4", expected: `5'11"`},
		{unit: fxp.FeetAndInches, format: whole, inches: "-71.6", expected: "-6'"},
		{unit: fxp.FeetAndInches, format: whole, inches: "0.4", expected: "0'"},
		{unit: fxp.FeetAndInches, format: whole, inches: "-0.4", expected: "0'"},
		{unit: fxp.FeetAndInches, format: two, inches: "71.4567", expected: `5'11.46"`},
		{unit: fxp.FeetAndInches, format: twoPadded, inches: "71.5", expected: `5'11.50"`},
		{unit: fxp.FeetAndInches, format: twoPadded, inches: "72", expected: "6'"},
		{unit: fxp.FeetAndInches, format: twoPadded, inches: "0", expected: "0'"},
		{unit: fxp.FeetAndInches, format: twoPadded, inches: "6", expected: `6.00"`},
		{unit: fxp.FeetAndInches, format: fxp.NumberFormat{}, inches: "71.4567", expected: `5'11.4567"`},
		{unit: fxp.Inch, format: whole, inches: "71.6", expected: "72 in"},
		{unit: fxp.Centimeter, format: whole, inches: "68.98", expected: "172 cm"},
		{unit: fxp.Centimeter, format: two, inches: "68.98", expected: "172.45 cm"},
		{unit: fxp.Meter, format: two, inches: "68.98", expected: "1.72 m"},
		{unit: fxp.Meter, format: twoPadded, inches: "68", expected: "1.70 m"},
		{unit: fxp.Feet, format: whole, inches: "71.6", expected: "6 ft"},
	} {
		c.Equal(d.expected, d.unit.FormatWith(fxp.Length(fxp.FromStringForced(d.inches)), d.format),
			"formatting %s inches as %s with %+v", d.inches, d.unit.Key(), d.format)
	}
}
