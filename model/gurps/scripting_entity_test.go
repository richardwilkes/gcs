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

// entity.wealthCarried and entity.wealthNotCarried sum the extended value of the carried and other equipment lists,
// respectively, matching Entity.WealthCarried and Entity.WealthNotCarried.
func TestScriptEntityWealth(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	carried := NewEquipment(e, nil, false)
	carried.Name = "Sword"
	carried.BaseValue = "10"
	carried.Quantity = fxp.Two
	e.CarriedEquipment = []*Equipment{carried}

	other := NewEquipment(e, nil, false)
	other.Name = "Spare Armor"
	other.BaseValue = "50"
	other.Quantity = fxp.One
	e.OtherEquipment = []*Equipment{other}
	e.Recalculate()

	resolve := func(expr string) string { return ResolveScript(e, ScriptSelfProvider{}, expr) }
	c.Equal("20", resolve("entity.wealthCarried.toString()"))
	c.Equal("50", resolve("entity.wealthNotCarried.toString()"))
	c.Equal(fxp.FromFloat[float64](20), e.WealthCarried())
	c.Equal(fxp.FromFloat[float64](50), e.WealthNotCarried())
}
