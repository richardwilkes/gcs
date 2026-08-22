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

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestTraitModifierCloneDoesNotShareReplacements verifies that a copy of a modifier still carrying the legacy
// replacements map gets its own map. setTrait() installs a modifier's replacements directly into the trait it is
// attached to when that trait has none of its own, so sharing the map let a later merge into the trait's map write into
// the library row the copy came from.
func TestTraitModifierCloneDoesNotShareReplacements(t *testing.T) {
	c := check.New(t)

	// A library row with legacy replacements, as it arrives from an .adm file written by an older version.
	library := NewTraitModifier(nil, nil, false)
	library.Name = "@Element@ Attunement"
	library.Replacements = map[string]string{"Element": "Fire"}

	// Dropping it onto a sheet's trait row clones it...
	dup := library.Clone(LibraryFile{}, nil, nil, Reference)
	c.Equal(map[string]string{"Element": "Fire"}, dup.Replacements, "the copy carries the replacements")

	// ...and the next attachment pass migrates the copy's map into the trait, which has none of its own.
	trait := NewTrait(nil, nil, false)
	dup.setTrait(trait)
	c.Nil(dup.Replacements, "the copy hands its map off to the trait")
	c.Equal(map[string]string{"Element": "Fire"}, trait.Replacements, "the trait picks up the replacements")

	// A second modifier attaching to the same trait takes setTrait's merge branch, writing into the map the trait now
	// holds. That must not reach back into the library row.
	second := NewTraitModifier(nil, nil, false)
	second.Replacements = map[string]string{"Aspect": "Heat"}
	second.setTrait(trait)
	c.Equal(map[string]string{"Element": "Fire", "Aspect": "Heat"}, trait.Replacements,
		"the trait merges in the second modifier's replacements")
	c.Equal(map[string]string{"Element": "Fire"}, library.Replacements, "the library row is left untouched")
}
