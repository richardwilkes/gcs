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

	s = gurps.ExprToScript(`roll(1d12+14,true)`)
	check.Equal(t, `dice.roll("1d12+14", true)`, s)

	s = gurps.ExprToScript(`roll(dice(1,12,14))`)
	check.Equal(t, `dice.roll(dice.from(1, 12, 14))`, s)

	s = gurps.ExprToScript(`add_dice(1d12+14,2d12-3)`)
	check.Equal(t, `dice.add("1d12+14", "2d12-3")`, s)

	s = gurps.ExprToScript(`(1 + 2) * 3 - max(1  /  2^3 + 2, if ( !$st, 3, 4)) + hello + if(world with me,x, y)`)
	check.Equal(t, `(1 + 2) * 3 - Math.max(1 / 2 ** 3 + 2, iff($st, 3, 4)) + "hello" + iff("world with me", "x", "y")`, s)

	s = gurps.ExprToScript(`$st+if(has_trait(Lifting ST),trait_level(Lifting ST),0)+if(skill_level(Sumo Wrestling)-$dx < 1,0,1)+if(skill_level(Sumo Wrestling)-$dx > 1,1,0)+if(has_trait(Wrestling Master),if(skill_level(Wrestling)-$dx < 1,0,skill_level(Wrestling)-$dx),if(skill_level(Wrestling)-$dx < 1,0,1)+if(skill_level(Wrestling)-$dx > 1,1,0))`)
	check.Equal(t, `$st + iff(entity.hasTrait("Lifting ST"), entity.traitLevel("Lifting ST"), 0) + iff(entity.skillLevel("Sumo Wrestling") - $dx < 1, 0, 1) + iff(entity.skillLevel("Sumo Wrestling") - $dx > 1, 1, 0) + iff(entity.hasTrait("Wrestling Master"), iff(entity.skillLevel("Wrestling") - $dx < 1, 0, entity.skillLevel("Wrestling") - $dx), iff(entity.skillLevel("Wrestling") - $dx < 1, 0, 1) + iff(entity.skillLevel("Wrestling") - $dx > 1, 1, 0))`, s)

	s = gurps.EmbeddedExprToScript(`Hello ||$name||!`)
	check.Equal(t, `Hello <script>$name</script>!`, s)

	s = gurps.EmbeddedExprToScript(`Hello ||$name||! A random weight for you is ||random_weight($st)||.`)
	check.Equal(t, `Hello <script>$name</script>! A random weight for you is <script>entity.randomWeightInPounds($st)</script>.`, s)

	s = gurps.EmbeddedExprToScript(`Hello ||$name||!
	A random weight for you is ||random_weight($st)||.`)
	check.Equal(t, `Hello <script>$name</script>!
	A random weight for you is <script>entity.randomWeightInPounds($st)</script>.`, s)

	s = gurps.ExprToScript(`trait_level("enhanced move (ground)")`)
	check.Equal(t, `entity.traitLevel("enhanced move (ground)")`, s)

	s = gurps.ExprToScript(`(2 * max(max($basic_move, floor(skill_level(jumping) / 2)), $st / 4) * (1 + max(0, trait_level("enhanced move (ground)"))) - 3) * enc(false, true) * if(trait_level("enhanced move (ground)")<1,2,1) * (2 ^ max(0, trait_level(super jump)))`)
	check.Equal(t, `(((2 * Math.max(Math.max($basic_move, Math.floor(entity.skillLevel("jumping") / 2)), $st / 4)) * (1 + Math.max(0, entity.traitLevel("enhanced move (ground)")))) - 3) * entity.currentEncumbrance(false, true) * iff(entity.traitLevel("enhanced move (ground)") < 1, 2, 1) * (2 ** Math.max(0, entity.traitLevel("super jump")))`, s)
}
