package gurps_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/check"
)

func TestExprToScript(t *testing.T) {
	s := gurps.ExprToScript("-$fp")
	check.Equal(t, `-$fp`, s)

	s = gurps.ExprToScript(`roll(1d12+14)`)
	check.Equal(t, `dice.roll("1d12+14")`, s)

	s = gurps.ExprToScript(`roll(dice(1,12,14))`)
	check.Equal(t, `dice.roll(dice.from(1, 12, 14))`, s)

	s = gurps.ExprToScript(`(1 + 2) * 3 - max(1  /  2^3 + 2, if ( !$st, 3, 4)) + hello + if(world with me,x, y)`)
	check.Equal(t, `(1 + 2) * 3 - Math.max(1 / 2 ** 3 + 2, iff($st, 3, 4)) + "hello" + iff("world with me", "x", "y")`, s)

	s = gurps.ExprToScript(`$st+if(has_trait(Lifting ST),trait_level(Lifting ST),0)+if(skill_level(Sumo Wrestling)-$dx < 1,0,1)+if(skill_level(Sumo Wrestling)-$dx > 1,1,0)+if(has_trait(Wrestling Master),if(skill_level(Wrestling)-$dx < 1,0,skill_level(Wrestling)-$dx),if(skill_level(Wrestling)-$dx < 1,0,1)+if(skill_level(Wrestling)-$dx > 1,1,0))`)
	check.Equal(t, `$st + iff(entity.hasTrait("Lifting ST"), entity.traitLevel("Lifting ST"), 0) + iff(entity.skillLevel("Sumo Wrestling") - $dx < 1, 0, 1) + iff(entity.skillLevel("Sumo Wrestling") - $dx > 1, 1, 0) + iff(entity.hasTrait("Wrestling Master"), iff(entity.skillLevel("Wrestling") - $dx < 1, 0, entity.skillLevel("Wrestling") - $dx), iff(entity.skillLevel("Wrestling") - $dx < 1, 0, 1) + iff(entity.skillLevel("Wrestling") - $dx > 1, 1, 0))`, s)

	s = gurps.EmbeddedExprToScript(`Hello ||$name||!`)
	check.Equal(t, `Hello ||$name||!`, s)

	s = gurps.EmbeddedExprToScript(`Hello ||$name||! A random weight for you is ||random_weight($st)||.`)
	check.Equal(t, `Hello ||$name||! A random weight for you is ||entity.randomWeight($st)||.`, s)

	s = gurps.EmbeddedExprToScript(`Hello ||$name||!
A random weight for you is ||random_weight($st)||.`)
	check.Equal(t, `Hello ||$name||!
A random weight for you is ||entity.randomWeight($st)||.`, s)
}
