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

func TestGURPSLengthConversion(t *testing.T) {
	assert.Equal(t, `1"`, measure.FeetAndInches.Format(measure.LengthFromInteger(1, measure.Inch)))
	assert.Equal(t, `1'3"`, measure.FeetAndInches.Format(measure.LengthFromInteger(15, measure.Inch)))
	assert.Equal(t, "2.5 cm", measure.Centimeter.Format(measure.LengthFromStringForced("2.5", measure.Centimeter)))
	assert.Equal(t, "37.5 cm", measure.Centimeter.Format(measure.LengthFromStringForced("37.5", measure.Centimeter)))

	w, err := measure.LengthFromString("1", measure.Inch)
	assert.NoError(t, err)
	assert.Equal(t, `1"`, measure.FeetAndInches.Format(w))
	w, err = measure.LengthFromString(`6'         2"`, measure.Inch)
	assert.NoError(t, err)
	assert.Equal(t, `6'2"`, measure.FeetAndInches.Format(w))
	w, err = measure.LengthFromString(" +32   yd  ", measure.Inch)
	assert.NoError(t, err)
	assert.Equal(t, "96'", measure.FeetAndInches.Format(w))
	w, err = measure.LengthFromString("0.5m", measure.Inch)
	assert.NoError(t, err)
	assert.Equal(t, "50 cm", measure.Centimeter.Format(w))
	w, err = measure.LengthFromString("1cm", measure.Inch)
	assert.NoError(t, err)
	assert.Equal(t, "1 cm", measure.Centimeter.Format(w))
}
