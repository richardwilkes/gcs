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

package model_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model"
	"github.com/stretchr/testify/assert"
)

func TestWeightConversion(t *testing.T) {
	assert.Equal(t, "1 lb", model.Pound.Format(model.WeightFromInteger(1, model.Pound)))
	assert.Equal(t, "15 lb", model.Pound.Format(model.WeightFromInteger(15, model.Pound)))
	assert.Equal(t, "0.5 kg", model.Kilogram.Format(model.WeightFromInteger(1, model.Pound)))
	assert.Equal(t, "7.5 kg", model.Kilogram.Format(model.WeightFromInteger(15, model.Pound)))

	w, err := model.WeightFromString("1", model.Pound)
	assert.NoError(t, err)
	assert.Equal(t, "1 lb", w.String())
	w, err = model.WeightFromString("1", model.Kilogram)
	assert.NoError(t, err)
	assert.Equal(t, "2 lb", w.String())
	w, err = model.WeightFromString("22.34 lb", model.Pound)
	assert.NoError(t, err)
	assert.Equal(t, "22.34 lb", w.String())
	w, err = model.WeightFromString(" +22.34   lb  ", model.Pound)
	assert.NoError(t, err)
	assert.Equal(t, "22.34 lb", w.String())
	w, err = model.WeightFromString("0.5kg", model.Pound)
	assert.NoError(t, err)
	assert.Equal(t, "0.5 kg", model.Kilogram.Format(w))
	w, err = model.WeightFromString(" 15.25 kg ", model.Pound)
	assert.NoError(t, err)
	assert.Equal(t, "15.25 kg", model.Kilogram.Format(w))
}
