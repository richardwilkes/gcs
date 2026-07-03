// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestHikingTimeInDays verifies the hiking travel-time calculation, including the 0 Move case that previously divided by
// zero and crashed the Calculator the moment a non-zero "Distance to Cover" was entered.
func TestHikingTimeInDays(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name            string
		distanceToCover fxp.Int
		distancePerDay  fxp.Int
		wantDays        fxp.Int
		wantOK          bool
	}{
		// Regression: distancePerDay of 0 (Move resolved to 0) must not divide by zero, even with distance to cover.
		{name: "0 move with distance to cover", distanceToCover: fxp.FromInteger(100), distancePerDay: 0, wantDays: 0, wantOK: false},
		{name: "0 move with no distance to cover", distanceToCover: 0, distancePerDay: 0, wantDays: 0, wantOK: false},
		// Nothing to cover is already "there": 0 days, and no division hazard.
		{name: "no distance to cover", distanceToCover: 0, distancePerDay: fxp.FromInteger(20), wantDays: 0, wantOK: true},
		// 100 miles to cover at 20 miles/day -> 5 days exactly.
		{name: "even multiple", distanceToCover: fxp.FromInteger(100), distancePerDay: fxp.FromInteger(20), wantDays: fxp.FromInteger(5), wantOK: true},
		// Rounds to a tenth of a day: 10 / 3 = 3.333... -> 3.3 days.
		{name: "rounds to tenths", distanceToCover: fxp.FromInteger(10), distancePerDay: fxp.FromInteger(3), wantDays: fxp.FromStringForced("3.3"), wantOK: true},
	} {
		days, ok := hikingTimeInDays(one.distanceToCover, one.distancePerDay)
		c.Equal(one.wantOK, ok, one.name)
		c.Equal(one.wantDays, days, one.name)
	}
}
