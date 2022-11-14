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

	measure2 "github.com/richardwilkes/gcs/v5/model/measure"
	"github.com/stretchr/testify/assert"
)

func TestGURPSLengthConversion(t *testing.T) {
	assert.Equal(t, `1"`, measure2.FeetAndInches.Format(measure2.LengthFromInteger(1, measure2.Inch)))
	assert.Equal(t, `1'3"`, measure2.FeetAndInches.Format(measure2.LengthFromInteger(15, measure2.Inch)))
	assert.Equal(t, "2.5 cm", measure2.Centimeter.Format(measure2.LengthFromStringForced("2.5", measure2.Centimeter)))
	assert.Equal(t, "37.5 cm", measure2.Centimeter.Format(measure2.LengthFromStringForced("37.5", measure2.Centimeter)))

	w, err := measure2.LengthFromString("1", measure2.Inch)
	assert.NoError(t, err)
	assert.Equal(t, `1"`, measure2.FeetAndInches.Format(w))
	w, err = measure2.LengthFromString(`6'         2"`, measure2.Inch)
	assert.NoError(t, err)
	assert.Equal(t, `6'2"`, measure2.FeetAndInches.Format(w))
	w, err = measure2.LengthFromString(" +32   yd  ", measure2.Inch)
	assert.NoError(t, err)
	assert.Equal(t, "96'", measure2.FeetAndInches.Format(w))
	w, err = measure2.LengthFromString("0.5m", measure2.Inch)
	assert.NoError(t, err)
	assert.Equal(t, "50 cm", measure2.Centimeter.Format(w))
	w, err = measure2.LengthFromString("1cm", measure2.Inch)
	assert.NoError(t, err)
	assert.Equal(t, "1 cm", measure2.Centimeter.Format(w))
}
