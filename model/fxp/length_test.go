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

func TestGURPSLengthConversion(t *testing.T) {
	c := check.New(t)
	c.Equal(`1"`, fxp.FeetAndInches.Format(fxp.LengthFromInteger(1, fxp.Inch)))
	c.Equal(`1'3"`, fxp.FeetAndInches.Format(fxp.LengthFromInteger(15, fxp.Inch)))
	c.Equal("2.5 cm", fxp.Centimeter.Format(fxp.LengthFromStringForced("2.5", fxp.Centimeter)))
	c.Equal("37.5 cm", fxp.Centimeter.Format(fxp.LengthFromStringForced("37.5", fxp.Centimeter)))

	w, err := fxp.LengthFromString("1", fxp.Inch)
	c.NoError(err)
	c.Equal(`1"`, fxp.FeetAndInches.Format(w))
	w, err = fxp.LengthFromString(`6'         2"`, fxp.Inch)
	c.NoError(err)
	c.Equal(`6'2"`, fxp.FeetAndInches.Format(w))
	w, err = fxp.LengthFromString(" +32   yd  ", fxp.Inch)
	c.NoError(err)
	c.Equal("96'", fxp.FeetAndInches.Format(w))
	w, err = fxp.LengthFromString("0.5m", fxp.Inch)
	c.NoError(err)
	c.Equal("50 cm", fxp.Centimeter.Format(w))
	w, err = fxp.LengthFromString("1cm", fxp.Inch)
	c.NoError(err)
	c.Equal("1 cm", fxp.Centimeter.Format(w))
}
