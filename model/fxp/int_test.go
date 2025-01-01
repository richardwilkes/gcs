// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/toolbox/check"
)

func TestApplyRounding(t *testing.T) {
	check.Equal(t, fxp.Two, fxp.ApplyRounding(fxp.OneAndAHalf, false))
	check.Equal(t, fxp.Two, fxp.ApplyRounding(fxp.OnePointTwo, false))
	check.Equal(t, fxp.One, fxp.ApplyRounding(fxp.OneAndAHalf, true))
	check.Equal(t, fxp.One, fxp.ApplyRounding(fxp.OnePointTwo, true))
	check.Equal(t, -fxp.Two, fxp.ApplyRounding(-fxp.OneAndAHalf, true))
	check.Equal(t, -fxp.Two, fxp.ApplyRounding(-fxp.OnePointTwo, true))
	check.Equal(t, -fxp.One, fxp.ApplyRounding(-fxp.OneAndAHalf, false))
	check.Equal(t, -fxp.One, fxp.ApplyRounding(-fxp.OnePointTwo, false))
	check.Equal(t, fxp.Two, fxp.ApplyRounding(fxp.Two, false))
	check.Equal(t, fxp.Two, fxp.ApplyRounding(fxp.Two, true))
}

func TestResetIfOutOfRange(t *testing.T) {
	check.Equal(t, 15, fxp.ResetIfOutOfRange(5, 10, 20, 15))
	check.Equal(t, 15, fxp.ResetIfOutOfRange(25, 10, 20, 15))
	check.Equal(t, 15, fxp.ResetIfOutOfRange(15, 10, 20, 5))
	check.Equal(t, fxp.Fifteen, fxp.ResetIfOutOfRange(fxp.Five, fxp.Ten, fxp.Twenty, fxp.Fifteen))
	check.Equal(t, fxp.Fifteen, fxp.ResetIfOutOfRange(fxp.TwentyFive, fxp.Ten, fxp.Twenty, fxp.Fifteen))
	check.Equal(t, fxp.Fifteen, fxp.ResetIfOutOfRange(fxp.Fifteen, fxp.Ten, fxp.Twenty, fxp.Five))
}

func TestExtract(t *testing.T) {
	for i, one := range []struct {
		input             string
		expectedValue     fxp.Int
		expectedRemainder string
	}{
		{" 24abc", fxp.TwentyFour, "abc"},
		{"24abc", fxp.TwentyFour, "abc"},
		{"24 abc", fxp.TwentyFour, " abc"},
		{"0.125abc", fxp.Eighth, "abc"},
		{"-0.125abc", -fxp.Eighth, "abc"},
		{"-24abc", -fxp.TwentyFour, "abc"},
		{"+24abc", fxp.TwentyFour, "abc"},
		{"abc", fxp.Int(0), "abc"},
		{"", fxp.Int(0), ""},
	} {
		value, remainder := fxp.Extract(one.input)
		check.Equal(t, one.expectedValue, value, "test %d", i)
		check.Equal(t, one.expectedRemainder, remainder, "test %d", i)
	}
}
