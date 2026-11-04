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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// A trait's levels are plain data that no disabling changes the meaning of, so the level commands stay available for a
// disabled trait: on a sheet enforcing trait prerequisites, decrementing is the one-step fix for a level above the
// maximum, which is what disabled the trait, and on a template, disabled alternatives are kept on purpose and still
// need their levels set.
func TestTraitLevelsStayAdjustableWhileDisabled(t *testing.T) {
	c := check.New(t)
	allowed := func(trait *gurps.Trait, increment bool) bool {
		_, ok := traitLevelExtractor(increment)(trait)
		return ok
	}

	entity := gurps.NewEntity()
	entity.SheetSettings.EnforceTraitPrereqs = true
	trait := gurps.NewTrait(entity, nil, false)
	trait.CanLevel = true
	trait.Levels = fxp.Four
	trait.MaxLevels = "3"
	entity.Traits = []*gurps.Trait{trait}
	entity.Recalculate()
	c.True(trait.DisabledByPrereqs(), "a level above the maximum disables the trait")
	c.True(allowed(trait, false), "decrementing the level of a trait its level disabled")
	c.True(allowed(trait, true), "incrementing the level of a trait its level disabled")

	trait.SetDisabled(true)
	c.True(allowed(trait, false), "decrementing the level of a trait the user disabled")
	c.True(allowed(trait, true), "incrementing the level of a trait the user disabled")

	tmpl := gurps.NewTemplate()
	alternative := gurps.NewTrait(tmpl, nil, false)
	alternative.CanLevel = true
	alternative.Levels = fxp.Two
	alternative.SetDisabled(true)
	tmpl.Traits = []*gurps.Trait{alternative}
	c.True(allowed(alternative, false), "decrementing the level of a disabled alternative on a template")
	c.True(allowed(alternative, true), "incrementing the level of a disabled alternative on a template")
}
