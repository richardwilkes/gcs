// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestWeaponBlock(t *testing.T) {
	c := check.New(t)
	for i, s := range []string{
		"0",
		"-1",
		"10",
		"No",
	} {
		c.Equal(s, gurps.ParseWeaponBlock(s).String(), "test %d", i)
	}

	cases := []struct {
		input    string
		expected string
	}{
		{"", "No"},
		{"-", "No"},
		{"+0", "0"},
		{"+1", "1"},
	}
	for i, one := range cases {
		c.Equal(one.expected, gurps.ParseWeaponBlock(one.input).String(), "test %d", i)
	}
}

// TestWeaponBlockResolveBlockTypeDefault verifies that a block-type weapon default does not have the +3 and the
// entity's block bonus applied to it twice. SkillLevelFast() already turns a block-type default into a block level, so
// a weapon defaulting to "Cloak Block" must resolve to the same value as one defaulting to the "Cloak" skill itself.
func TestWeaponBlockResolveBlockTypeDefault(t *testing.T) {
	c := check.New(t)
	w := newDefenseTestWeapon(c)

	// A skill-type default halves the skill level, then adds the +3 and the entity's block bonus: 14/2 + 3 + 1.
	w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(gurps.SkillID)}
	c.Equal("11", w.Block.Resolve(w, nil).String(), "skill-type default")

	// A block-type default has already been through that conversion, so it must yield the same block, not 15.
	w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(gurps.BlockID)}
	c.Equal("11", w.Block.Resolve(w, nil).String(), "block-type default")

	// Both kinds of default together still resolve to the same block.
	w.Defaults = []*gurps.SkillDefault{
		newDefenseTestDefault(gurps.SkillID),
		newDefenseTestDefault(gurps.BlockID),
	}
	c.Equal("11", w.Block.Resolve(w, nil).String(), "both defaults")

	// The weapon's own block modifier and the default's modifier still apply on top.
	w.Block.Modifier = fxp.Two
	w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(gurps.BlockID)}
	w.Defaults[0].Modifier = -fxp.One
	c.Equal("12", w.Block.Resolve(w, nil).String(), "block-type default with modifiers")
}
