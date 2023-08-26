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
	"github.com/richardwilkes/toolbox/check"
)

func TestRealLengthConversion(t *testing.T) {
	check.Equal(t, `0.25 in`, gurps.PaperLength{Length: 0.25, Units: gurps.InchPaperUnits}.String())
	check.Equal(t, float32(18), gurps.PaperLength{Length: 0.25, Units: gurps.InchPaperUnits}.Pixels())
	check.Equal(t, `1 in`, gurps.PaperLength{Length: 1, Units: gurps.InchPaperUnits}.String())
	check.Equal(t, float32(72), gurps.PaperLength{Length: 1, Units: gurps.InchPaperUnits}.Pixels())
	check.Equal(t, `15 in`, gurps.PaperLength{Length: 15, Units: gurps.InchPaperUnits}.String())
	check.Equal(t, float32(1080), gurps.PaperLength{Length: 15, Units: gurps.InchPaperUnits}.Pixels())
	check.Equal(t, "1 cm", gurps.PaperLength{Length: 1, Units: gurps.CentimeterPaperUnits}.String())
	check.Equal(t, float32(28.3464566929), gurps.PaperLength{Length: 1, Units: gurps.CentimeterPaperUnits}.Pixels())
	check.Equal(t, "1 mm", gurps.PaperLength{Length: 1, Units: gurps.MillimeterPaperUnits}.String())
	check.Equal(t, float32(2.8346456693), gurps.PaperLength{Length: 1, Units: gurps.MillimeterPaperUnits}.Pixels())
}
