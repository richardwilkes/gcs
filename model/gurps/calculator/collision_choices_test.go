// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calculator

import (
	"fmt"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// checkNamed verifies that every entry of a choice table has a name and shows it.
func checkNamed[T fmt.Stringer](c check.Checker, what string, items []T, names func(T) string) {
	c.True(len(items) > 0, "the %s table must have entries", what)
	for i, one := range items {
		c.True(names(one) != "", "%s entry %d must be named", what, i)
		c.Equal(names(one), one.String(), "%s entry %d must show its name", what, i)
	}
}

// TestCollisionChoices verifies the collision tables: their order, which the calculator's indexes depend on, and the
// facts each entry carries.
func TestCollisionChoices(t *testing.T) {
	c := check.New(t)

	shapes := CollisionShapes()
	checkNamed(c, "shape", shapes, func(s CollisionShape) string { return s.Name })
	c.Equal([]CollisionShape{
		{Name: shapes[0].Name, DamageType: "cr"},
		{Name: shapes[1].Name, DamageType: "pi", HalfDamage: true},
		{Name: shapes[2].Name, DamageType: "cut", HalfDamage: true},
		{Name: shapes[3].Name, DamageType: "imp", HalfDamage: true},
	}, shapes, "blunt does full crushing damage; the other shapes do half damage of their own type")

	surfaces := CollisionSurfaces()
	checkNamed(c, "surface", surfaces, func(s CollisionSurface) string { return s.Name })
	c.Equal(4, len(surfaces), "there are four kinds of surface")
	c.True(surfaces[0].Hard, "the first surface is the hard one a sudden stop uses")
	c.True(surfaces[2].Elastic, "the third is elastic")
	c.True(surfaces[3].Water, "the fourth is water")
	for i, s := range surfaces {
		c.True(!s.Hard || !s.Elastic && !s.Water, "surface %d must be one kind", i)
	}

	terminals := TerminalVelocityChoices()
	checkNamed(c, "terminal velocity", terminals, func(t TerminalVelocityChoice) string { return t.Name })
	c.Equal([]fxp.Int{fxp.Sixty, fxp.Hundred, fxp.FromInteger(200), 0, 0},
		[]fxp.Int{terminals[0].BaseVelocity, terminals[1].BaseVelocity, terminals[2].BaseVelocity, terminals[3].BaseVelocity, terminals[4].BaseVelocity},
		"the terminal velocities are 60, 100 and 200, then none")
	for i, one := range terminals {
		c.Equal(i == len(terminals)-1, one.Custom, "only the last terminal velocity is custom")
	}

	angles := CollisionAngleChoices()
	checkNamed(c, "angle", angles, func(a CollisionAngleChoice) string { return a.Name })
	c.Equal([]CollisionAngle{HeadOnCollision, RearEndCollision, SideOnCollision},
		[]CollisionAngle{angles[0].Angle, angles[1].Angle, angles[2].Angle}, "the angles are offered in the order of the rules")

	restraints := CollisionRestraints()
	checkNamed(c, "restraint", restraints, func(r CollisionRestraint) string { return r.Name })
	c.Equal([]int{0, 5, 10}, []int{restraints[0].DR, restraints[1].DR, restraints[2].DR}, "a seatbelt is DR 5 and an airbag DR 10")
}
