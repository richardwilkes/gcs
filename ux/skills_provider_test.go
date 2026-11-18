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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// TestMovingGrantedRowLeavesItAlone checks that what a move does to the rows moved, filling in an empty tech level and
// recalculating the level, leaves a granted row alone, since it changes only with its grant.
func TestMovingGrantedRowLeavesItAlone(t *testing.T) {
	c := check.New(t)
	sheet, _, granted, _ := newTestSheetWithGrant(t)
	sheet.Entity().Profile.TechLevel = "3"
	techLevel := ""
	granted.TechLevel = &techLevel
	table := sheet.Skills.Table
	table.SetSelectionMap(map[tid.TID]bool{granted.ID(): true})
	provider, ok := any(table.Model).(TableProvider[*gurps.Skill])
	if !ok {
		t.Fatal("the skills table has no provider")
	}
	provider.ProcessDropData(nil, table)
	c.Equal("", *granted.TechLevel, "the granted row's tech level is left as its grant made it")
}
