/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/check"
)

func TestEntityAttributeBonus(t *testing.T) {
	entity := NewEntity(PC)
	check.Equal(t, fxp.Ten, entity.Attributes.Current("st"), "ST default")
	check.Equal(t, entity.SwingFor(10), entity.Swing(), "Swing default")
	check.Equal(t, fxp.WeightFromInteger(20, fxp.Pound), entity.BasicLift(), "Basic Lift default")
	check.Equal(t, fxp.From(0), entity.ThrowingStrengthBonus, "Throwing ST Bonus default")

	bonus := NewAttributeBonus("st")
	trait := NewTrait(entity, nil, false)
	trait.Features = append(trait.Features, bonus)
	entity.Traits = append(entity.Traits, trait)
	entity.Recalculate()
	check.Equal(t, fxp.From(11), entity.Attributes.Current("st"), "ST; simple +1 bonus")

	bonus.PerLevel = true
	entity.Recalculate()
	check.Equal(t, fxp.Ten, entity.Attributes.Current("st"), "ST; leveled +1 bonus, but no levels")

	trait.CanLevel = true
	trait.Levels = fxp.Three
	entity.Recalculate()
	check.Equal(t, fxp.From(13), entity.Attributes.Current("st"), "ST; leveled +1 bonus, with 3 levels")

	bonus.Limitation = StrikingOnlyBonusLimitation
	entity.Recalculate()
	check.Equal(t, fxp.Ten, entity.Attributes.Current("st"), "ST; leveled +1 bonus, with 3 levels, for striking only")
	check.Equal(t, entity.SwingFor(13), entity.Swing(), "Swing; leveled +1 bonus, with 3 levels, for striking only")

	bonus.Limitation = LiftingOnlyBonusLimitation
	entity.Recalculate()
	check.Equal(t, fxp.Ten, entity.Attributes.Current("st"), "ST; leveled +1 bonus, with 3 levels, for lifting only")
	check.Equal(t, fxp.WeightFromInteger(34, fxp.Pound), entity.BasicLift(), "Basic Lift; leveled +1 bonus, with 3 levels, for lifting only")

	bonus.Limitation = ThrowingOnlyBonusLimitation
	entity.Recalculate()
	check.Equal(t, fxp.Ten, entity.Attributes.Current("st"), "ST; leveled +1 bonus, with 3 levels, for throwing only")
	check.Equal(t, fxp.From(3), entity.ThrowingStrengthBonus, "Throwing ST Bonus; leveled +1 bonus, with 3 levels, for throwing only")
}
