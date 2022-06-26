/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package measure_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/measure"
	"github.com/stretchr/testify/assert"
)

func TestWeightConversion(t *testing.T) {
	assert.Equal(t, "1 lb", measure.Pound.Format(measure.WeightFromInteger(1, measure.Pound)))
	assert.Equal(t, "15 lb", measure.Pound.Format(measure.WeightFromInteger(15, measure.Pound)))
	assert.Equal(t, "0.5 kg", measure.Kilogram.Format(measure.WeightFromInteger(1, measure.Pound)))
	assert.Equal(t, "7.5 kg", measure.Kilogram.Format(measure.WeightFromInteger(15, measure.Pound)))

	w, err := measure.WeightFromString("1", measure.Pound)
	assert.NoError(t, err)
	assert.Equal(t, "1 lb", w.String())
	w, err = measure.WeightFromString("1", measure.Kilogram)
	assert.NoError(t, err)
	assert.Equal(t, "2 lb", w.String())
	w, err = measure.WeightFromString("22.34 lb", measure.Pound)
	assert.NoError(t, err)
	assert.Equal(t, "22.34 lb", w.String())
	w, err = measure.WeightFromString(" +22.34   lb  ", measure.Pound)
	assert.NoError(t, err)
	assert.Equal(t, "22.34 lb", w.String())
	w, err = measure.WeightFromString("0.5kg", measure.Pound)
	assert.NoError(t, err)
	assert.Equal(t, "0.5 kg", measure.Kilogram.Format(w))
	w, err = measure.WeightFromString(" 15.25 kg ", measure.Pound)
	assert.NoError(t, err)
	assert.Equal(t, "15.25 kg", measure.Kilogram.Format(w))
}
