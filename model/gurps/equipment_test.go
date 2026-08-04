// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestEquipmentRatedStrength verifies that equipment reports its user-settable rated ST. Unlike traits, skills and
// spells -- whose RatedStrength always returns 0 -- equipment genuinely carries one, and weapon strength calculations
// depend on getting it back.
func TestEquipmentRatedStrength(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	equipment := NewEquipment(e, nil, false)
	c.Equal(fxp.Int(0), equipment.RatedStrength(), "equipment with no rated ST reports 0")
	equipment.RatedST = fxp.FromInteger(12)
	c.Equal(fxp.FromInteger(12), equipment.RatedStrength(), "equipment reports the rated ST it was given")
}
