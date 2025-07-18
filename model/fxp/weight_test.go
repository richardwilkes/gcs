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

func TestWeightConversion(t *testing.T) {
	c := check.New(t)
	c.Equal("1 lb", fxp.Pound.Format(fxp.WeightFromInteger(1, fxp.Pound)))
	c.Equal("15 lb", fxp.Pound.Format(fxp.WeightFromInteger(15, fxp.Pound)))
	c.Equal("0.5 kg", fxp.Kilogram.Format(fxp.WeightFromInteger(1, fxp.Pound)))
	c.Equal("7.5 kg", fxp.Kilogram.Format(fxp.WeightFromInteger(15, fxp.Pound)))

	w, err := fxp.WeightFromString("1", fxp.Pound)
	c.NoError(err)
	c.Equal("1 lb", w.String())
	w, err = fxp.WeightFromString("1", fxp.Kilogram)
	c.NoError(err)
	c.Equal("2 lb", w.String())
	w, err = fxp.WeightFromString("22.34 lb", fxp.Pound)
	c.NoError(err)
	c.Equal("22.34 lb", w.String())
	w, err = fxp.WeightFromString(" +22.34   lb  ", fxp.Pound)
	c.NoError(err)
	c.Equal("22.34 lb", w.String())
	w, err = fxp.WeightFromString("0.5kg", fxp.Pound)
	c.NoError(err)
	c.Equal("0.5 kg", fxp.Kilogram.Format(w))
	w, err = fxp.WeightFromString(" 15.25 kg ", fxp.Pound)
	c.NoError(err)
	c.Equal("15.25 kg", fxp.Kilogram.Format(w))
}
