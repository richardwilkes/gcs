package gurps_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/check"
)

func TestExprToScript(t *testing.T) {
	s, err := gurps.ExprToScript(`1 + 2 * 3 - max(1  /  2^3 + 2, if ( !$st, 3, 4)) + hello + if(world with me,x, y)`)
	check.NoError(t, err)
	check.Equal(t, `1 + 2 * 3 - Math.max(1 / 2 ** 3 + 2, iff($st, 3, 4)) + "hello" + iff("world with me", "x", "y")`, s)

	s, err = gurps.ExprToScript(`$st+if(has_trait(Lifting ST),trait_level(Lifting ST),0)+if(skill_level(Sumo Wrestling)-$dx < 1,0,1)+if(skill_level(Sumo Wrestling)-$dx > 1,1,0)+if(has_trait(Wrestling Master),if(skill_level(Wrestling)-$dx < 1,0,skill_level(Wrestling)-$dx),if(skill_level(Wrestling)-$dx < 1,0,1)+if(skill_level(Wrestling)-$dx > 1,1,0))`)
	check.NoError(t, err)
	check.Equal(t, `$st + iff(has_trait("Lifting ST"), trait_level("Lifting ST"), 0) + iff(skill_level("Sumo Wrestling") - $dx < 1, 0, 1) + iff(skill_level("Sumo Wrestling") - $dx > 1, 1, 0) + iff(has_trait("Wrestling Master"), iff(skill_level("Wrestling") - $dx < 1, 0, skill_level("Wrestling") - $dx), iff(skill_level("Wrestling") - $dx < 1, 0, 1) + iff(skill_level("Wrestling") - $dx > 1, 1, 0))`, s)
}
