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
	"hash/fnv"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/maxusesmod"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

type testLeveledOwner struct {
	level   fxp.Int
	leveled bool
}

func (o *testLeveledOwner) IsLeveled() bool       { return o.leveled }
func (o *testLeveledOwner) CurrentLevel() fxp.Int { return o.level }

// TestMaxUsesModAmount covers the shared amount logic used by both EquipmentMaxUsesBonus and TraitMaxLevelBonus: the
// operation implied by the text, per-level scaling (including the no-owner, non-leveled and zero-level guards), the
// tooltip line and the hash.
func TestMaxUsesModAmount(t *testing.T) {
	c := check.New(t)

	for _, one := range []struct {
		amount string
		op     maxusesmod.Type
		value  fxp.Int
		format string
	}{
		{amount: "+2", op: maxusesmod.Addition, value: fxp.Two, format: "+2"},
		{amount: "-3", op: maxusesmod.Addition, value: -fxp.Three, format: "-3"},
		{amount: "50%", op: maxusesmod.Percentage, value: fxp.FromInteger(50), format: "+50%"},
		{amount: "x2", op: maxusesmod.Multiplier, value: fxp.Two, format: "x2"},
	} {
		m := MaxUsesModAmount{Amount: one.amount}
		c.Equal(one.op, m.Operation(), one.amount)
		c.Equal(one.value, m.AdjustedAmount(), one.amount)
		c.Equal(one.format, m.Format(), one.amount)
	}

	// Per-level scaling requires a leveled owner with a positive level.
	m := MaxUsesModAmount{Amount: "+2", PerLevel: true}
	c.Equal(fxp.Int(0), m.AdjustedAmount(), "per-level with no owner contributes nothing")
	m.SetLeveledOwner(&testLeveledOwner{level: fxp.Three})
	c.Equal(fxp.Int(0), m.AdjustedAmount(), "per-level with a non-leveled owner contributes nothing")
	m.SetLeveledOwner(&testLeveledOwner{leveled: true})
	c.Equal(fxp.Int(0), m.AdjustedAmount(), "per-level with a zero-level owner contributes nothing")
	m.SetLeveledOwner(&testLeveledOwner{leveled: true, level: fxp.Three})
	c.Equal(fxp.Six, m.AdjustedAmount(), "per-level scales by the owner's level")
	c.Equal("+6", m.Format(), "formatted per-level amount")
	m.PerLevel = false
	c.Equal(fxp.Two, m.AdjustedAmount(), "a fixed amount ignores the owner's level")

	// Tooltip: a nil buffer is tolerated, otherwise a line is appended.
	m.addToTooltip("Source", nil)
	var buffer xbytes.InsertBuffer
	m.addToTooltip("Source", &buffer)
	c.Equal("\nSource [+2]", buffer.String())

	// Hash: both the amount and the per-level flag participate.
	hashOf := func(m *MaxUsesModAmount) uint64 {
		h := fnv.New64a()
		m.Hash(h)
		return h.Sum64()
	}
	base := hashOf(&MaxUsesModAmount{Amount: "+2"})
	c.Equal(base, hashOf(&MaxUsesModAmount{Amount: "+2"}), "equal contents hash equally")
	c.NotEqual(base, hashOf(&MaxUsesModAmount{Amount: "+3"}), "amount participates in the hash")
	c.NotEqual(base, hashOf(&MaxUsesModAmount{Amount: "+2", PerLevel: true}), "per-level participates in the hash")
	c.NotEqual(base, hashOf(nil), "nil hashes distinctly")
}
