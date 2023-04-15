/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/stretchr/testify/assert"
)

func TestWeightConversion(t *testing.T) {
	assert.Equal(t, "1 lb", gurps.Pound.Format(gurps.WeightFromInteger(1, gurps.Pound)))
	assert.Equal(t, "15 lb", gurps.Pound.Format(gurps.WeightFromInteger(15, gurps.Pound)))
	assert.Equal(t, "0.5 kg", gurps.Kilogram.Format(gurps.WeightFromInteger(1, gurps.Pound)))
	assert.Equal(t, "7.5 kg", gurps.Kilogram.Format(gurps.WeightFromInteger(15, gurps.Pound)))

	w, err := gurps.WeightFromString("1", gurps.Pound)
	assert.NoError(t, err)
	assert.Equal(t, "1 lb", w.String())
	w, err = gurps.WeightFromString("1", gurps.Kilogram)
	assert.NoError(t, err)
	assert.Equal(t, "2 lb", w.String())
	w, err = gurps.WeightFromString("22.34 lb", gurps.Pound)
	assert.NoError(t, err)
	assert.Equal(t, "22.34 lb", w.String())
	w, err = gurps.WeightFromString(" +22.34   lb  ", gurps.Pound)
	assert.NoError(t, err)
	assert.Equal(t, "22.34 lb", w.String())
	w, err = gurps.WeightFromString("0.5kg", gurps.Pound)
	assert.NoError(t, err)
	assert.Equal(t, "0.5 kg", gurps.Kilogram.Format(w))
	w, err = gurps.WeightFromString(" 15.25 kg ", gurps.Pound)
	assert.NoError(t, err)
	assert.Equal(t, "15.25 kg", gurps.Kilogram.Format(w))
}
