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

func TestGURPSLengthConversion(t *testing.T) {
	assert.Equal(t, `1"`, model.FeetAndInches.Format(model.LengthFromInteger(1, model.Inch)))
	assert.Equal(t, `1'3"`, model.FeetAndInches.Format(model.LengthFromInteger(15, model.Inch)))
	assert.Equal(t, "2.5 cm", model.Centimeter.Format(model.LengthFromStringForced("2.5", model.Centimeter)))
	assert.Equal(t, "37.5 cm", model.Centimeter.Format(model.LengthFromStringForced("37.5", model.Centimeter)))

	w, err := model.LengthFromString("1", model.Inch)
	assert.NoError(t, err)
	assert.Equal(t, `1"`, model.FeetAndInches.Format(w))
	w, err = model.LengthFromString(`6'         2"`, model.Inch)
	assert.NoError(t, err)
	assert.Equal(t, `6'2"`, model.FeetAndInches.Format(w))
	w, err = model.LengthFromString(" +32   yd  ", model.Inch)
	assert.NoError(t, err)
	assert.Equal(t, "96'", model.FeetAndInches.Format(w))
	w, err = model.LengthFromString("0.5m", model.Inch)
	assert.NoError(t, err)
	assert.Equal(t, "50 cm", model.Centimeter.Format(w))
	w, err = model.LengthFromString("1cm", model.Inch)
	assert.NoError(t, err)
	assert.Equal(t, "1 cm", model.Centimeter.Format(w))
}
