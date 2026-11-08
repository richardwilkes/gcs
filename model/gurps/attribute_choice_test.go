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
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestAttributeChoicesForMoveAndPools verifies that MoveFlag adds a stand-in for Move only while the definitions list
// no move attribute, which is listed whatever the flags, and that NoPoolsFlag leaves the pools out, a dodge or move
// pool included, with a stand-in for Dodge or Move in its place.
func TestAttributeChoicesForMoveAndPools(t *testing.T) {
	c := check.New(t)
	keys := func(defs *AttributeDefs, flags AttributeFlags) []string {
		choices, _ := AttributeChoicesFor(defs, "", flags, StrengthID)
		result := make([]string, 0, len(choices))
		for _, one := range choices {
			result = append(result, one.Key)
		}
		return result
	}
	defs := FactoryAttributeDefs()
	c.True(slices.Contains(keys(defs, MoveFlag), MoveID), "a stand-in for Move with the flag")
	c.False(slices.Contains(keys(defs, 0), MoveID), "none without it")
	c.True(slices.Contains(keys(defs, 0), FatiguePointsID), "the pools are listed")
	c.False(slices.Contains(keys(defs, NoPoolsFlag), FatiguePointsID), "unless left out")
	c.True(slices.Contains(keys(defs, NoPoolsFlag), StrengthID), "which leaves the others alone")

	defs.Set[MoveID] = &AttributeDef{DefID: MoveID, Type: attribute.Integer, Name: "Move", Base: "$basic_move", Order: len(defs.Set)}
	list := keys(defs, MoveFlag)
	c.Equal(1, len(slices.DeleteFunc(slices.Clone(list), func(k string) bool { return k != MoveID })),
		"a defined move attribute is listed once, with no stand-in beside it")
	c.True(slices.Contains(keys(defs, 0), MoveID), "and listed without the flag")

	defs.Set[MoveID].Type = attribute.Pool
	defs.Set[MoveID].Name = "Move Pool"
	defs.Set[DodgeID] = &AttributeDef{DefID: DodgeID, Type: attribute.Pool, Name: "Dodge Pool", Base: "8", Order: len(defs.Set)}
	choices, _ := AttributeChoicesFor(defs, "", ThresholdAttributeFlags, StrengthID)
	titles := make(map[string]string)
	for _, one := range choices {
		titles[one.Key] = one.Title
	}
	c.Equal("Move", titles[MoveID], "a move pool is left out, with a stand-in for Move in its place")
	c.Equal("Dodge", titles[DodgeID], "and so is a dodge pool, with a stand-in for Dodge")
}
