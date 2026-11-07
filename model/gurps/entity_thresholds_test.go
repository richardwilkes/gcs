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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestCurrentAdjustmentApply verifies the order and rounding of a current adjustment: the amount is added first,
// floored for a whole value, and then each factor is applied in turn, rounding to the nearest whole number after each,
// so that halving twice rounds up twice.
func TestCurrentAdjustmentApply(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name  string
		adj   currentAdjustment
		value fxp.Int
		whole bool
		want  fxp.Int
	}{
		{"zero value", currentAdjustment{}, fxp.Ten, true, fxp.Ten},
		{"amount added", currentAdjustment{bonus: fxp.Two}, fxp.Ten, true, fxp.Twelve},
		{"fraction of an amount dropped for a whole value", currentAdjustment{bonus: fxp.OneAndAHalf}, fxp.Ten, true, fxp.Eleven},
		{"fraction of an amount kept otherwise", currentAdjustment{bonus: fxp.OneAndAHalf}, fxp.Ten, false, fxp.FromStringForced("11.5")},
		{"amount before factor", currentAdjustment{bonus: fxp.Two, factors: []fxp.Int{fxp.Half}}, fxp.Ten, true, fxp.Six},
		{"halving rounds up", currentAdjustment{factors: []fxp.Int{fxp.Half}}, fxp.Seven, true, fxp.Four},
		{"halving twice rounds up twice", currentAdjustment{factors: []fxp.Int{fxp.Half, fxp.Half}}, fxp.Five, true, fxp.Two},
		{"halving twice rounds up twice again", currentAdjustment{factors: []fxp.Int{fxp.Half, fxp.Half}}, fxp.Nine, true, fxp.Three},
		{"fraction of a product kept otherwise", currentAdjustment{factors: []fxp.Int{fxp.Half}}, fxp.Five, false, fxp.TwoAndAHalf},
		{"a tenth off rounds to nearest", currentAdjustment{factors: []fxp.Int{fxp.FromStringForced("0.9")}}, fxp.FromInteger(16), true, fxp.FromInteger(14)},
		{"factor of zero", currentAdjustment{factors: []fxp.Int{0}}, fxp.Ten, true, 0},
	} {
		c.Equal(one.want, one.adj.apply(one.value, one.whole), one.name)
	}
}

// TestEntityFactoryThresholdsHalveMoveDodgeAndLiftingST verifies that the factory thresholds halve Move and Dodge,
// rounding up, once for Tired and once for Reeling, and halve the ST used for lifting alone while Tired, leaving the
// attributes themselves and the damage alone.
func TestEntityFactoryThresholdsHalveMoveDodgeAndLiftingST(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name  string
		fp    fxp.Int
		hp    fxp.Int
		move  int
		dodge int
		lift  int
	}{
		{"rested and healthy", 0, 0, 5, 8, 20},
		{"tired", fxp.Seven, 0, 3, 4, 5},
		{"reeling", 0, fxp.Seven, 3, 4, 20},
		{"tired and reeling", fxp.Seven, fxp.Seven, 2, 2, 5},
	} {
		e := NewEntity()
		e.Attributes.Set[FatiguePointsID].Damage = one.fp
		e.Attributes.Set[HitPointsID].Damage = one.hp
		e.Recalculate()
		c.Equal(one.move, e.Move(encumbrance.No), "%s: Move", one.name)
		c.Equal(one.dodge, e.Dodge(encumbrance.No), "%s: Dodge", one.name)
		c.Equal(fxp.WeightFromInteger(one.lift, fxp.Pound), e.BasicLift(), "%s: Basic Lift", one.name)
		c.Equal(fxp.Ten, e.Attributes.Current(StrengthID), "%s: ST itself", one.name)
		c.Equal(fxp.Five, e.Attributes.Current(BasicMoveID), "%s: Basic Move itself", one.name)
		c.Equal(e.SwingFor(10), e.Swing(), "%s: damage", one.name)
	}
}

// TestEntityThresholdFeaturesTakeEffect verifies that the other kinds of feature a threshold holds take effect while
// the pool is at it, naming the threshold and its pool as their source, and not otherwise, save one the threshold
// ignores, such as a reaction per level, which never does.
func TestEntityThresholdFeaturesTakeEffect(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	brawling := addSkill(e, "Brawling", fxp.One)
	tired := poolThreshold(e, FatiguePointsID, "Tired")
	c.NotNil(tired)
	reaction := NewReactionBonus()
	reaction.Situation = "from everyone"
	reaction.Amount = -fxp.One
	perLevel := NewReactionBonus()
	perLevel.Situation = "from strangers"
	perLevel.Amount = -fxp.One
	perLevel.PerLevel = true
	tired.Features = append(tired.Features, newSkillBonusTo("Brawling", fxp.Two), reaction, perLevel)
	e.Recalculate()
	level := brawling.CalculateLevel(nil).Level
	c.Equal(0, len(e.Reactions()), "no reaction while Tiring")

	e.Attributes.Set[FatiguePointsID].Damage = fxp.Seven
	e.Recalculate()
	c.Equal(level+fxp.Two, brawling.CalculateLevel(nil).Level, "the skill bonus takes effect while Tired")
	reactions := e.Reactions()
	c.Equal(1, len(reactions), "the reaction takes effect while Tired")
	c.Equal("from everyone", reactions[0].From)
	c.Equal([]string{"from Tired (FP)"}, reactions[0].Sources, "the threshold and its pool are the source")
}

// TestEntityMoveBonus verifies that a bonus to Move adds to Basic Move while the sheet defines no move attribute, and
// goes to that attribute once it does.
func TestEntityMoveBonus(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	bonus := NewAttributeBonus(MoveID)
	addTraitWithFeatures(e, "Quick", bonus)
	e.Recalculate()
	c.Equal(fxp.One, e.MoveBonus, "the bonus is held apart while no move attribute is defined")
	c.Equal(6, e.Move(encumbrance.No), "and added to Basic Move")

	e.SheetSettings.Attributes.Set[MoveID] = &AttributeDef{
		DefID: MoveID,
		Type:  attribute.Integer,
		Name:  "Move",
		Base:  "$basic_move",
	}
	e.Attributes.Set[MoveID] = NewAttribute(e, MoveID, len(e.Attributes.Set))
	e.Recalculate()
	c.Equal(fxp.Int(0), e.MoveBonus, "a move attribute carries the bonus itself")
	c.Equal(fxp.Six, e.Attributes.Maximum(MoveID))
	c.Equal(6, e.Move(encumbrance.No))
}
