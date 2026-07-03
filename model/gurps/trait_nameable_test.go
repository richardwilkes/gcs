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

// TestTraitLevelBonusUsesResolvedName verifies that a TraitBonus aimed at a leveled trait's resolved
// (replacement-applied) name adjusts the trait's level. internalCurrentLevel previously matched the bonus against the
// raw stored name (containing @placeholders@), so a bonus targeting the resolved name never applied (issue #4).
func TestTraitLevelBonusUsesResolvedName(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	// A leveled trait with a nameable name that resolves to "Talent (Rifle)".
	talent := NewTrait(e, nil, false)
	talent.Name = "Talent (@which@)"
	talent.Replacements = map[string]string{"which": "Rifle"}
	talent.CanLevel = true
	talent.Levels = fxp.Two

	// A separate trait that grants +1 level to "Talent (Rifle)".
	source := NewTrait(e, nil, false)
	source.Name = "Bonus Source"
	bonus := NewTraitBonus()
	bonus.NameCriteria.Qualifier = "Talent (Rifle)"
	source.Features = append(source.Features, bonus)

	e.Traits = append(e.Traits, talent, source)
	e.Recalculate()
	c.Equal(fxp.Three, talent.CurrentLevel(), "a TraitBonus on the resolved name should raise level 2 -> 3")

	// A bonus qualified by the unresolved raw name must not match the resolved trait.
	bonus.NameCriteria.Qualifier = "Talent (@which@)"
	e.Recalculate()
	c.Equal(fxp.Two, talent.CurrentLevel(), "a TraitBonus on the raw @placeholder@ name must not apply")
}
