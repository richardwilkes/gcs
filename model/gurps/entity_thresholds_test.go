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
	"cmp"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantadjust"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
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
// ignores, such as a reaction per level, which never does. A trait bonus raises the level that the trait's per-level
// attribute bonus reads, which the attribute's value follows, save that of a pool or the size modifier, which a
// threshold never alters, even through an attribute a pool's base reads, so that the pool stays at the threshold whose
// features apply.
func TestEntityThresholdFeaturesTakeEffect(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	brawling := addSkill(e, "Brawling", fxp.One)
	stPerLevel := newAttrBonus(StrengthID, fxp.One)
	stPerLevel.PerLevel = true
	fpPerLevel := newAttrBonus(FatiguePointsID, fxp.One)
	fpPerLevel.PerLevel = true
	smPerLevel := newAttrBonus(SizeModifierID, fxp.One)
	smPerLevel.PerLevel = true
	lifting := addTraitWithFeatures(e, "Lifting", stPerLevel, fpPerLevel, smPerLevel)
	lifting.CanLevel = true
	lifting.Levels = fxp.Two
	st := e.Attributes.Set[StrengthID]
	tired := poolThreshold(e, FatiguePointsID, "Tired")
	c.NotNil(tired)
	reaction := NewReactionBonus()
	reaction.Situation = "from everyone"
	reaction.Amount = -fxp.One
	perLevel := NewReactionBonus()
	perLevel.Situation = "from strangers"
	perLevel.Amount = -fxp.One
	perLevel.PerLevel = true
	tired.Features = append(tired.Features, newSkillBonusTo("Brawling", fxp.Two), reaction, perLevel,
		newTraitBonusTo("Lifting", fxp.Two))
	e.Recalculate()
	level := brawling.CalculateLevel(nil).Level
	c.Equal(0, len(e.Reactions()), "no reaction while Tiring")
	c.Equal(fxp.Twelve, st.Maximum(), "ST from Lifting 2 while Tiring")

	fp := e.Attributes.Set[FatiguePointsID]
	fp.Damage = fxp.Nine
	e.Recalculate()
	c.Equal(level+fxp.Two, brawling.CalculateLevel(nil).Level, "the skill bonus takes effect while Tired")
	c.Equal(fxp.Four, lifting.CurrentLevel(), "the trait bonus takes effect while Tired")
	c.Equal(fxp.FromInteger(14), st.Maximum(), "ST follows the level the trait bonus raises")
	c.Equal(fxp.Twelve, fp.Maximum(), "FP doesn't follow it")
	c.Equal(fxp.Twelve, e.Attributes.Maximum(HitPointsID), "nor does HP, whose base reads ST")
	c.Equal("Tired", fp.CurrentThreshold().State, "so FP stays at the threshold whose features apply")
	c.Equal(2, e.Profile.AdjustedSizeModifier(), "nor does the size modifier")
	c.Equal(fxp.FromInteger(14), st.Current())
	c.Equal("Includes modifiers from:\nLifting 4 [+4 (+1 per level)]\nFor lifting only:\nTired (FP) [×0.5]",
		st.BonusTooltip())
	reactions := e.Reactions()
	c.Equal(1, len(reactions), "the reaction takes effect while Tired")
	c.Equal("from everyone", reactions[0].From)
	c.Equal([]string{"from Tired (FP)"}, reactions[0].Sources, "the threshold and its pool are the source")
}

// TestEntityPoolBaseLeavesOutThresholdTraitBonuses verifies that a pool's base leaves out what the trait bonuses its
// thresholds hold add to anything it reads that the levels of the traits move, such as the level of a default, Dodge,
// a defense, a skill or spell or a relative skill level, rather than read the values the sheet recorded with them, so
// that the pool stays at the threshold whose features apply, and that nothing found for the base then is recorded on
// the sheet or kept for the next time it is found.
func TestEntityPoolBaseLeavesOutThresholdTraitBonuses(t *testing.T) {
	c := check.New(t)
	perLevel := func(bonus Feature) Feature {
		switch b := bonus.(type) {
		case *AttributeBonus:
			b.PerLevel = true
		case *SkillBonus:
			b.PerLevel = true
		case *SpellBonus:
			b.PerLevel = true
		}
		return bonus
	}
	defenseDefault := func(id string) func(e *Entity) Feature {
		return func(e *Entity) Feature {
			addSkill(e, "Sword", fxp.One)
			def := newSkillDefaultTo("Sword", "", true, 0)
			def.DefaultType = id
			addSkill(e, "Riposte", 0).Defaults = []*SkillDefault{def}
			return perLevel(newAttrBonus(id, fxp.One))
		}
	}
	// Each setup puts on the sheet what the base reads and returns the per-level feature of Agile, which has levels (1
	// unless given) and to which Drained holds a trait bonus of held (+4 unless given). evade, when set, is the level
	// the sheet gives Evade while Drained.
	for _, one := range []struct {
		name   string
		base   string
		want   fxp.Int
		levels fxp.Int
		held   fxp.Int
		evade  fxp.Int
		setup  func(e *Entity) Feature
	}{
		{
			name: "a skill defaulting to Dodge", base: `entity.skillLevel("Evade")`, want: fxp.Nine,
			setup: func(e *Entity) Feature {
				addSkill(e, "Evade", 0).Defaults = []*SkillDefault{{DefaultType: DodgeID}}
				return perLevel(newAttrBonus(DodgeID, fxp.One))
			},
		},
		{
			// While Drained, the sheet's Dodge 13 makes Evade's default stand for 8 of the 12 points that set it at 14,
			// and the base's Dodge 9 for none.
			name: "a skill defaulting to Dodge that a grant sets at a level", base: `entity.skillLevel("Evade")`,
			want: fxp.FromInteger(14), evade: fxp.FromInteger(14), setup: func(e *Entity) Feature {
				evade := newEasySkill("Evade", 0)
				evade.Defaults = []*SkillDefault{{DefaultType: DodgeID}}
				addTraitWithFeatures(e, "Training",
					newSkillGrantOf(evade, grantadjust.SetLevel, fxp.FromInteger(14), grantmerge.Ignore))
				return perLevel(newAttrBonus(DodgeID, fxp.One))
			},
		},
		{
			name: "a skill defaulting to a parry", base: `entity.skillLevel("Riposte")`, want: fxp.Nine,
			setup: defenseDefault(ParryID),
		},
		{
			name: "a skill defaulting to a block", base: `entity.skillLevel("Riposte")`, want: fxp.Nine,
			setup: defenseDefault(BlockID),
		},
		{
			name: "a skill defaulting to the parry of the best of several skills", base: `entity.skillLevel("Riposte")`,
			want: fxp.Ten, levels: fxp.Five, held: -fxp.Four, setup: func(e *Entity) Feature {
				addSkill(e, "Sword", fxp.Four).Difficulty.Attribute = IntelligenceID
				addSkill(e, "Sword", fxp.One)
				def := newSkillDefaultTo("Sword", "", true, 0)
				def.DefaultType = ParryID
				addSkill(e, "Riposte", 0).Defaults = []*SkillDefault{def}
				return perLevel(newAttrBonus(DexterityID, fxp.One))
			},
		},
		{
			name: "a skill defaulting to an attribute", base: `entity.skillLevel("Juggling")`, want: fxp.Nine,
			setup: func(e *Entity) Feature {
				addSkill(e, "Juggling", 0).Defaults = []*SkillDefault{{DefaultType: DexterityID, Modifier: -fxp.Two}}
				return perLevel(newAttrBonus(DexterityID, fxp.One))
			},
		},
		{
			name: "a skill a per-level bonus of a skill raises", base: `entity.skillLevel("Trance")`, want: fxp.FromInteger(21),
			setup: func(e *Entity) Feature {
				addSkill(e, "Meditation", fxp.One).Features = Features{perLevel(newSkillBonusTo("Trance", fxp.One))}
				addSkill(e, "Trance", fxp.One)
				return perLevel(newSkillBonusTo("Meditation", fxp.One))
			},
		},
		{
			name: "a skill a per-level bonus of a spell raises", base: `entity.skillLevel("Trance")`, want: fxp.FromInteger(19),
			setup: func(e *Entity) Feature {
				addTestSpell(e, "Light", fxp.One).Features = Features{perLevel(newSkillBonusTo("Trance", fxp.One))}
				addSkill(e, "Trance", fxp.One)
				return perLevel(NewSpellBonus())
			},
		},
		{
			name: "a weapon's level", base: `entity.weapons[0].level`, want: fxp.Eleven,
			setup: func(e *Entity) Feature {
				addSkill(e, "Sword", fxp.One)
				blade := addTraitWithFeatures(e, "Blade")
				swing := NewWeapon(blade, true)
				swing.Defaults = []*SkillDefault{newSkillDefaultTo("Sword", "", true, 0)}
				blade.Weapons = []*Weapon{swing}
				return perLevel(newAttrBonus(DexterityID, fxp.One))
			},
		},
		{
			name: "a weapon bonus selecting by relative skill level", want: fxp.Ten,
			base: `entity.weaponDamage("Fist", "Punch") == "1d-2 cr" ? 10 : 12`,
			setup: func(e *Entity) Feature {
				addSkill(e, "Brawling", fxp.One)
				bonus := NewWeaponBonus(feature.WeaponBonus)
				bonus.NameCriteria.Qualifier = "Brawling"
				bonus.RelativeLevelCriteria = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.Three}
				bonus.Amount = fxp.Five
				fist := addTraitWithFeatures(e, "Fist", bonus)
				punch := NewWeapon(fist, true)
				punch.Usage = "Punch"
				punch.Defaults = []*SkillDefault{newSkillDefaultTo("Brawling", "", true, 0)}
				fist.Weapons = []*Weapon{punch}
				return perLevel(newSkillBonusTo("Brawling", fxp.One))
			},
		},
	} {
		e := NewEntity()
		agile := addTraitWithFeatures(e, "Agile", one.setup(e))
		agile.CanLevel = true
		agile.Levels = cmp.Or(one.levels, fxp.One)
		held := cmp.Or(one.held, fxp.Four)
		drained := NewPoolThreshold()
		drained.State = "Drained"
		drained.Value = "5"
		drained.Features = Features{newTraitBonusTo("Agile", held)}
		fresh := NewPoolThreshold()
		fresh.State = "Fresh"
		fresh.Value = "$focus"
		e.SheetSettings.Attributes.Set["focus"] = &AttributeDef{
			DefID:      "focus",
			Type:       attribute.Pool,
			Name:       "Focus",
			Base:       one.base,
			Thresholds: []*PoolThreshold{drained, fresh},
		}
		focus := NewAttribute(e, "focus", len(e.Attributes.Set))
		e.Attributes.Set["focus"] = focus
		e.Recalculate()
		c.Equal(one.want, focus.Maximum(), "%s: the base while Fresh", one.name)
		focus.Damage = one.want - fxp.One
		e.Recalculate()
		c.Equal("Drained", focus.CurrentThreshold().State, "%s", one.name)
		c.Equal(agile.Levels+held, agile.CurrentLevel(), "%s: the trait bonus takes effect while Drained", one.name)
		c.Equal(one.want, focus.Maximum(), "%s: the base doesn't follow it", one.name)
		if one.evade != 0 {
			c.Equal(one.evade, e.SkillNamed("Evade", "", false, nil)[0].LevelData.Level, "%s: the sheet's Evade",
				one.name)
		}
		agile.Levels += fxp.One
		focus.Damage = 0
		e.Recalculate()
		raised := focus.Maximum()
		focus.Damage = raised - fxp.One
		e.Recalculate()
		c.Equal("Drained", focus.CurrentThreshold().State, "%s: once the trait gains a level", one.name)
		c.Equal(raised, focus.Maximum(), "%s: the base once the trait gains a level", one.name)
	}
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

// TestEntityPoolThresholdSeesItemMultipliers verifies that the sheet, and the prerequisites of an item carrying a
// grant, find the threshold of a pool whose base reads an attribute's current value with the multipliers the items give
// that attribute in place, as the pool's state and value are shown, so that the features applied are those of the
// threshold shown.
func TestEntityPoolThresholdSeesItemMultipliers(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTraitWithFeatures(e, "Hardy", newMultiplier(HealthID, fxp.Two))
	training := addTraitWithFeatures(e, "Training", newTraitGrant(newGrantLeveledTrait("Skilled", fxp.One), fxp.One,
		grantmerge.Ignore))
	training.Prereq = NewPrereqList()
	dx := NewAttributePrereq(e)
	dx.Parent = training.Prereq
	dx.Which = DexterityID
	training.Prereq.Prereqs = append(training.Prereq.Prereqs, dx)
	weak := NewPoolThreshold()
	weak.State = "Weak"
	weak.Value = "5"
	weak.Features = Features{newAttrBonus(DexterityID, -fxp.Three)}
	ok := NewPoolThreshold()
	ok.State = "OK"
	ok.Value = "$stam"
	e.SheetSettings.Attributes.Set["stam"] = &AttributeDef{
		DefID:      "stam",
		Type:       attribute.Pool,
		Name:       "Stam",
		Base:       "$ht.current",
		Thresholds: []*PoolThreshold{weak, ok},
	}
	stam := NewAttribute(e, "stam", len(e.Attributes.Set))
	stam.Damage = fxp.Seven
	e.Attributes.Set["stam"] = stam
	e.Recalculate()
	c.Equal(fxp.Thirteen, stam.Current(), "Stam rests on HT's current value of 20")
	c.Equal("OK", stam.CurrentThreshold().State)
	c.Equal(fxp.Ten, e.Attributes.Current(DexterityID), "Weak's penalty is not applied while the pool shows OK")
	c.Equal("", training.UnsatisfiedReason, "nor while the prerequisites see the pool")
}
