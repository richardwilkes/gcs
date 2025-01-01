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

func TestWeightConversion(t *testing.T) {
	check.Equal(t, "1 lb", fxp.Pound.Format(fxp.WeightFromInteger(1, fxp.Pound)))
	check.Equal(t, "15 lb", fxp.Pound.Format(fxp.WeightFromInteger(15, fxp.Pound)))
	check.Equal(t, "0.5 kg", fxp.Kilogram.Format(fxp.WeightFromInteger(1, fxp.Pound)))
	check.Equal(t, "7.5 kg", fxp.Kilogram.Format(fxp.WeightFromInteger(15, fxp.Pound)))

	w, err := fxp.WeightFromString("1", fxp.Pound)
	check.NoError(t, err)
	check.Equal(t, "1 lb", w.String())
	w, err = fxp.WeightFromString("1", fxp.Kilogram)
	check.NoError(t, err)
	check.Equal(t, "2 lb", w.String())
	w, err = fxp.WeightFromString("22.34 lb", fxp.Pound)
	check.NoError(t, err)
	check.Equal(t, "22.34 lb", w.String())
	w, err = fxp.WeightFromString(" +22.34   lb  ", fxp.Pound)
	check.NoError(t, err)
	check.Equal(t, "22.34 lb", w.String())
	w, err = fxp.WeightFromString("0.5kg", fxp.Pound)
	check.NoError(t, err)
	check.Equal(t, "0.5 kg", fxp.Kilogram.Format(w))
	w, err = fxp.WeightFromString(" 15.25 kg ", fxp.Pound)
	check.NoError(t, err)
	check.Equal(t, "15.25 kg", fxp.Kilogram.Format(w))
}
