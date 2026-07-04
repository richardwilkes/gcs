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

// addTestSpell creates a non-container spell owned by the entity, giving it the supplied name and points, then appends
// it to the entity's spell list.
func addTestSpell(e *Entity, name string, points fxp.Int) *Spell {
	s := NewSpell(e, nil, false)
	s.Name = name
	s.Points = points
	e.Spells = append(e.Spells, s)
	return s
}

// TestSpellAdjustedRelativeLevel verifies that AdjustedRelativeLevel returns the cached relative level for a positive
// level spell (matching the freshly-computed level), and fxp.Min for containers and unowned spells. This guards the
// switch from recomputing the whole level via CalculateLevel to reading the cached LevelData.
func TestSpellAdjustedRelativeLevel(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	s := addTestSpell(e, "Fireball", fxp.Four)
	e.Recalculate()

	// Precondition: the spell must resolve to a positive level so AdjustedRelativeLevel takes the non-Min branch.
	c.True(s.LevelData.Level > 0, "precondition: the spell must have a positive level")

	// The cached relative level must be returned, and it must match a fresh recomputation (proving the cheaper cached
	// read is equivalent to the old CalculateLevel path).
	c.Equal(s.LevelData.RelativeLevel, s.AdjustedRelativeLevel())
	c.Equal(s.CalculateLevel().RelativeLevel, s.AdjustedRelativeLevel())

	// A container spell has no meaningful relative level.
	container := NewSpell(e, nil, true)
	c.Equal(fxp.Min, container.AdjustedRelativeLevel())

	// A spell with no owning entity has no relative level either.
	orphan := NewSpell(nil, nil, false)
	orphan.Name = "Fireball"
	orphan.Points = fxp.Four
	c.Equal(fxp.Min, orphan.AdjustedRelativeLevel())
}
