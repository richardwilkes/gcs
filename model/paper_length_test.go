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

func TestRealLengthConversion(t *testing.T) {
	assert.Equal(t, `0.25 in`, model.PaperLength{Length: 0.25, Units: model.InchPaperUnits}.String())
	assert.Equal(t, float32(18), model.PaperLength{Length: 0.25, Units: model.InchPaperUnits}.Pixels())
	assert.Equal(t, `1 in`, model.PaperLength{Length: 1, Units: model.InchPaperUnits}.String())
	assert.Equal(t, float32(72), model.PaperLength{Length: 1, Units: model.InchPaperUnits}.Pixels())
	assert.Equal(t, `15 in`, model.PaperLength{Length: 15, Units: model.InchPaperUnits}.String())
	assert.Equal(t, float32(1080), model.PaperLength{Length: 15, Units: model.InchPaperUnits}.Pixels())
	assert.Equal(t, "1 cm", model.PaperLength{Length: 1, Units: model.CentimeterPaperUnits}.String())
	assert.Equal(t, float32(28.3464566929), model.PaperLength{Length: 1, Units: model.CentimeterPaperUnits}.Pixels())
	assert.Equal(t, "1 mm", model.PaperLength{Length: 1, Units: model.MillimeterPaperUnits}.String())
	assert.Equal(t, float32(2.8346456693), model.PaperLength{Length: 1, Units: model.MillimeterPaperUnits}.Pixels())
}
