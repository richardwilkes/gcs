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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestExprToScript(t *testing.T) {
	c := check.New(t)
	s := gurps.ExprToScript("-$fp")
	c.Equal(`-$fp`, s)

	s = gurps.ExprToScript(`roll(1d12+14)`)
	c.Equal(`dice.roll("1d12+14")`, s)

	s = gurps.ExprToScript(`roll(1d12+14,true)`)
	c.Equal(`dice.roll("1d12+14", true)`, s)

	s = gurps.ExprToScript(`roll(dice(1,12,14))`)
	c.Equal(`dice.roll(dice.from(1, 12, 14))`, s)

	s = gurps.ExprToScript(`add_dice(1d12+14,2d12-3)`)
	c.Equal(`dice.add("1d12+14", "2d12-3")`, s)

	s = gurps.ExprToScript(`(1 + 2) * 3 - max(1  /  2^3 + 2, if ( !$st, 3, 4)) + hello + if(world with me,x, y)`)
	c.Equal(`(1 + 2) * 3 - Math.max(1 / 2 ** 3 + 2, iff(!$st, 3, 4)) + "hello" + iff("world with me", "x", "y")`, s)

	s = gurps.ExprToScript(`$st+if(has_trait(Lifting ST),trait_level(Lifting ST),0)+if(skill_level(Sumo Wrestling)-$dx < 1,0,1)+if(skill_level(Sumo Wrestling)-$dx > 1,1,0)+if(has_trait(Wrestling Master),if(skill_level(Wrestling)-$dx < 1,0,skill_level(Wrestling)-$dx),if(skill_level(Wrestling)-$dx < 1,0,1)+if(skill_level(Wrestling)-$dx > 1,1,0))`)
	c.Equal(`$st + iff(entity.hasTrait("Lifting ST"), entity.traitLevel("Lifting ST"), 0) + iff(entity.skillLevel("Sumo Wrestling") - $dx < 1, 0, 1) + iff(entity.skillLevel("Sumo Wrestling") - $dx > 1, 1, 0) + iff(entity.hasTrait("Wrestling Master"), iff(entity.skillLevel("Wrestling") - $dx < 1, 0, entity.skillLevel("Wrestling") - $dx), iff(entity.skillLevel("Wrestling") - $dx < 1, 0, 1) + iff(entity.skillLevel("Wrestling") - $dx > 1, 1, 0))`, s)

	s = gurps.EmbeddedExprToScript(`Hello ||$name||!`)
	c.Equal(`Hello <script>$name</script>!`, s)

	s = gurps.EmbeddedExprToScript(`Hello ||$name||! A random weight for you is ||random_weight($st)||.`)
	c.Equal(`Hello <script>$name</script>! A random weight for you is <script>entity.randomWeightInPounds($st)</script>.`, s)

	s = gurps.EmbeddedExprToScript(`Hello ||$name||!
	A random weight for you is ||random_weight($st)||.`)
	c.Equal(`Hello <script>$name</script>!
	A random weight for you is <script>entity.randomWeightInPounds($st)</script>.`, s)

	s = gurps.ExprToScript(`trait_level("enhanced move (ground)")`)
	c.Equal(`entity.traitLevel("enhanced move (ground)")`, s)

	s = gurps.ExprToScript(`(2 * max(max($basic_move, floor(skill_level(jumping) / 2)), $st / 4) * (1 + max(0, trait_level("enhanced move (ground)"))) - 3) * enc(false, true) * if(trait_level("enhanced move (ground)")<1,2,1) * (2 ^ max(0, trait_level(super jump)))`)
	c.Equal(`(((2 * Math.max(Math.max($basic_move, Math.floor(entity.skillLevel("jumping") / 2)), $st / 4)) * (1 + Math.max(0, entity.traitLevel("enhanced move (ground)")))) - 3) * entity.currentEncumbrance(false, true) * iff(entity.traitLevel("enhanced move (ground)") < 1, 2, 1) * (2 ** Math.max(0, entity.traitLevel("super jump")))`, s)
}

func TestExprToScriptUnaryOperators(t *testing.T) {
	c := check.New(t)
	for i, one := range []struct {
		in  string
		out string
	}{
		// A unary operator applied to a parenthesized group must be emitted as a prefix, not a suffix.
		{in: `!($st > 10)`, out: `!($st > 10)`},
		{in: `-(2 + 3)`, out: `-(2 + 3)`},
		{in: `-(2 + 3) - 4`, out: `-(2 + 3) - 4`},
		{in: `!(1 + 2) * 3`, out: `!(1 + 2) * 3`},
		{in: `!($st)`, out: `!$st`},

		// A unary operator applied to a function call must be retained.
		{in: `!has_trait(Magery)`, out: `!entity.hasTrait("Magery")`},
		{in: `-max(1,2)`, out: `-Math.max(1, 2)`},
		{in: `-roll(1d6)`, out: `-dice.roll("1d6")`},
		{in: `!enc(false,true)`, out: `!entity.currentEncumbrance(false, true)`},
		{in: `if(!has_trait(Magery), 1, 0)`, out: `iff(!entity.hasTrait("Magery"), 1, 0)`},

		// A unary operator is recognized anywhere an operand doesn't precede it, not just at index 0.
		{in: `-$fp`, out: `-$fp`},
		{in: `if( !$st, 3, 4)`, out: `iff(!$st, 3, 4)`},
		{in: `$st > -1`, out: `$st > -1`},
		{in: `$st>=-1`, out: `$st >= -1`},
		{in: `2 * -3`, out: `2 * -3`},
		{in: `2 - -3`, out: `2 - -3`},
		{in: `max(1,2) * -3`, out: `Math.max(1, 2) * -3`},
		{in: `max(-1, 2)`, out: `Math.max(-1, 2)`},
		{in: `if(-$st < 0, 1, 0)`, out: `iff(-$st < 0, 1, 0)`},

		// The hack that permits negative exponents on floating point numbers must still work.
		{in: `1.2e-2 + 1`, out: `1.2e-2 + 1`},
	} {
		c.Equal(one.out, gurps.ExprToScript(one.in), "index %d: %s", i, one.in)
	}
}

func TestExprToScriptMalformed(t *testing.T) {
	c := check.New(t)
	// Malformed expressions must be left unchanged rather than panicking or emitting garbage.
	for i, one := range []string{
		`(*`,
		`(+)`,
		`(1 + 2`,
		`max((),1)`,
		`max( ,1)`,
		`-()`,
		`--3`,
		`max(unknown_function(1),2)`,
	} {
		c.Equal(one, gurps.ExprToScript(one), "index %d: %s", i, one)
	}
}
