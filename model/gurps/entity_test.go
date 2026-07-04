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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestEntityAttributeBonus(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	c.Equal(fxp.Ten, e.Attributes.Current("st"), "ST default")
	c.Equal(e.SwingFor(10), e.Swing(), "Swing default")
	c.Equal(fxp.WeightFromInteger(20, fxp.Pound), e.BasicLift(), "Basic Lift default")
	c.Equal(fxp.Int(0), e.ThrowingStrengthBonus, "Throwing ST Bonus default")

	bonus := NewAttributeBonus("st")
	trait := NewTrait(e, nil, false)
	trait.Features = append(trait.Features, bonus)
	e.Traits = append(e.Traits, trait)
	e.Recalculate()
	c.Equal(fxp.Eleven, e.Attributes.Current("st"), "ST; simple +1 bonus")

	bonus.PerLevel = true
	e.Recalculate()
	c.Equal(fxp.Ten, e.Attributes.Current("st"), "ST; leveled +1 bonus, but no levels")

	trait.CanLevel = true
	trait.Levels = fxp.Three
	e.Recalculate()
	c.Equal(fxp.Thirteen, e.Attributes.Current("st"), "ST; leveled +1 bonus, with 3 levels")

	bonus.Limitation = stlimit.StrikingOnly
	e.Recalculate()
	c.Equal(fxp.Ten, e.Attributes.Current("st"), "ST; leveled +1 bonus, with 3 levels, for striking only")
	c.Equal(e.SwingFor(13), e.Swing(), "Swing; leveled +1 bonus, with 3 levels, for striking only")

	bonus.Limitation = stlimit.LiftingOnly
	e.Recalculate()
	c.Equal(fxp.Ten, e.Attributes.Current("st"), "ST; leveled +1 bonus, with 3 levels, for lifting only")
	c.Equal(fxp.WeightFromInteger(34, fxp.Pound), e.BasicLift(), "Basic Lift; leveled +1 bonus, with 3 levels, for lifting only")

	bonus.Limitation = stlimit.ThrowingOnly
	e.Recalculate()
	c.Equal(fxp.Ten, e.Attributes.Current("st"), "ST; leveled +1 bonus, with 3 levels, for throwing only")
	c.Equal(fxp.Three, e.ThrowingStrengthBonus, "Throwing ST Bonus; leveled +1 bonus, with 3 levels, for throwing only")
}

func TestEntityHideZeroValueConditionalModifiers(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	// A situation whose contributing bonuses cancel out to a total of zero.
	addConditionalModifier(e, "cancels out", fxp.One)
	addConditionalModifier(e, "cancels out", fxp.NegOne)
	// A situation with a non-zero total.
	addConditionalModifier(e, "still applies", fxp.Two)

	e.SheetSettings.HideZeroValueConditionalMods = false
	mods := e.ConditionalModifiers()
	c.Equal(2, len(mods), "both situations listed when zero-value modifiers are shown")

	e.SheetSettings.HideZeroValueConditionalMods = true
	mods = e.ConditionalModifiers()
	c.Equal(1, len(mods), "zero-total situation omitted when hidden")
	if len(mods) == 1 {
		c.Equal("still applies", mods[0].From, "the remaining modifier is the non-zero one")
		c.Equal(fxp.Two, mods[0].Total(), "the remaining modifier retains its total")
	}
}

func addConditionalModifier(e *Entity, situation string, amt fxp.Int) {
	bonus := NewConditionalModifierBonus()
	bonus.Situation = situation
	bonus.Amount = amt
	trait := NewTrait(e, nil, false)
	trait.Features = append(trait.Features, bonus)
	e.Traits = append(e.Traits, trait)
}
