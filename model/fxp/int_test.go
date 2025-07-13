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
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestApplyRounding(t *testing.T) {
	c := check.New(t)
	c.Equal(fxp.Two, fxp.ApplyRounding(fxp.OneAndAHalf, false))
	c.Equal(fxp.Two, fxp.ApplyRounding(fxp.OnePointTwo, false))
	c.Equal(fxp.One, fxp.ApplyRounding(fxp.OneAndAHalf, true))
	c.Equal(fxp.One, fxp.ApplyRounding(fxp.OnePointTwo, true))
	c.Equal(-fxp.Two, fxp.ApplyRounding(-fxp.OneAndAHalf, true))
	c.Equal(-fxp.Two, fxp.ApplyRounding(-fxp.OnePointTwo, true))
	c.Equal(-fxp.One, fxp.ApplyRounding(-fxp.OneAndAHalf, false))
	c.Equal(-fxp.One, fxp.ApplyRounding(-fxp.OnePointTwo, false))
	c.Equal(fxp.Two, fxp.ApplyRounding(fxp.Two, false))
	c.Equal(fxp.Two, fxp.ApplyRounding(fxp.Two, true))
}

func TestResetIfOutOfRange(t *testing.T) {
	c := check.New(t)
	c.Equal(15, fxp.ResetIfOutOfRange(5, 10, 20, 15))
	c.Equal(15, fxp.ResetIfOutOfRange(25, 10, 20, 15))
	c.Equal(15, fxp.ResetIfOutOfRange(15, 10, 20, 5))
	c.Equal(fxp.Fifteen, fxp.ResetIfOutOfRange(fxp.Five, fxp.Ten, fxp.Twenty, fxp.Fifteen))
	c.Equal(fxp.Fifteen, fxp.ResetIfOutOfRange(fxp.TwentyFive, fxp.Ten, fxp.Twenty, fxp.Fifteen))
	c.Equal(fxp.Fifteen, fxp.ResetIfOutOfRange(fxp.Fifteen, fxp.Ten, fxp.Twenty, fxp.Five))
}

func TestExtract(t *testing.T) {
	c := check.New(t)
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
		c.Equal(one.expectedValue, value, "test %d", i)
		c.Equal(one.expectedRemainder, remainder, "test %d", i)
	}
}
