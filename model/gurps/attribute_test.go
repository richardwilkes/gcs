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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/toolbox/v2/check"
)

// newAttrBonusTrait creates a trait carrying a single bonus to the given attribute with the given amount and
// limitation, and adds it to the entity.
func newAttrBonusTrait(e *Entity, name, attrID string, amount fxp.Int, limitation stlimit.Option) *Trait {
	bonus := NewAttributeBonus(attrID)
	bonus.Amount = amount
	bonus.Limitation = limitation
	return addTraitWithFeatures(e, name, bonus)
}

// TestAttributeBonusTooltipListsSources verifies that the tooltip names each source of an attribute bonus along with
// its amount, in the order the entity processes them, and that it is empty when nothing contributes.
func TestAttributeBonusTooltipListsSources(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	st := e.Attributes.Set[StrengthID]
	e.Recalculate()
	c.Equal("", st.BonusTooltip(), "no bonuses means no tooltip")

	trait := newAttrBonusTrait(e, "Increased Strength", StrengthID, fxp.Two, stlimit.None)

	// An equipment modifier's bonus is attributed to both the item and the modifier.
	modBonus := NewAttributeBonus(StrengthID)
	modBonus.Amount = fxp.One
	mod := NewEquipmentModifier(e, nil, false)
	mod.Name = "AR software"
	mod.Features = Features{modBonus}
	eqp := addCarriedEquipmentWithFeatures(e, "Smart Gloves")
	eqp.AddModifiers(mod)

	e.Recalculate()
	c.Equal("Includes modifiers from:\nIncreased Strength [+2]\nSmart Gloves (AR software) [+1]", st.BonusTooltip(),
		"each active bonus is listed with its source and amount")
	c.Equal(fxp.Three, st.Bonus, "the tooltip reflects the bonus that was applied")

	trait.SetDisabled(true)
	e.Recalculate()
	c.Equal("Includes modifiers from:\nSmart Gloves (AR software) [+1]", st.BonusTooltip(),
		"a disabled trait's bonus is not listed")

	eqp.Equipped = false
	e.Recalculate()
	c.Equal("", st.BonusTooltip(), "an unequipped item's bonus is not listed")

	trait.SetDisabled(false)
	e.Recalculate()
	c.Equal("Includes modifiers from:\nIncreased Strength [+2]", st.BonusTooltip(),
		"a re-enabled trait's bonus is listed again, while the unequipped item's stays out")
	c.Equal(fxp.Two, st.Bonus, "the tooltip reflects the bonus that was applied")

	// Bonuses for other attributes never leak into the tooltip.
	c.Equal("", e.Attributes.Set[DexterityID].BonusTooltip(), "an unrelated attribute has no tooltip")
}

// TestAttributeBonusTooltipTraitModifierSource verifies that a bonus granted by one of a trait's modifiers names both
// the trait and the modifier as its source, just as an equipment modifier's bonus does, and that disabling the
// modifier takes the bonus away entirely.
func TestAttributeBonusTooltipTraitModifierSource(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	st := e.Attributes.Set[StrengthID]
	trait := newAttrBonusTrait(e, "Increased Strength", StrengthID, fxp.Two, stlimit.None)

	modBonus := NewAttributeBonus(StrengthID)
	modBonus.Amount = fxp.One
	mod := NewTraitModifier(e, nil, false)
	mod.Name = "Cybernetic"
	mod.Features = Features{modBonus}
	trait.AddModifiers(mod)

	e.Recalculate()
	c.Equal("Includes modifiers from:\nIncreased Strength [+2]\nIncreased Strength (Cybernetic) [+1]",
		st.BonusTooltip(), "a trait modifier's bonus is attributed to both the trait and the modifier")
	c.Equal(fxp.Three, st.Bonus, "the tooltip reflects the bonus that was applied")

	mod.SetDisabled(true)
	e.Recalculate()
	c.Equal("Includes modifiers from:\nIncreased Strength [+2]", st.BonusTooltip(),
		"a disabled modifier's bonus is not listed")
	c.Equal(fxp.Two, st.Bonus, "a disabled modifier's bonus is not applied")
}

// TestAttributeBonusTooltipNonStrengthLimitation verifies that a limitation left over on a bonus to something other
// than ST is ignored. The feature editor doesn't clear the limitation when the attribute is switched away from ST, so
// such a bonus must still be applied and listed under the plain heading, not under a section only ST has.
func TestAttributeBonusTooltipNonStrengthLimitation(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	dx := e.Attributes.Set[DexterityID]
	newAttrBonusTrait(e, "Nimble Fingers", DexterityID, fxp.Two, stlimit.LiftingOnly)
	e.Recalculate()
	c.Equal("Includes modifiers from:\nNimble Fingers [+2]", dx.BonusTooltip(),
		"a limitation on a non-ST bonus doesn't move it into a section of its own")
	c.Equal(fxp.Twelve, dx.Maximum(), "a limitation on a non-ST bonus doesn't stop it from being applied")
}

// TestAttributeBonusTooltipLimitedStrength verifies that ST bonuses limited to striking, lifting or throwing are listed
// in sections of their own, in that order, since they do not contribute to the attribute's value.
func TestAttributeBonusTooltipLimitedStrength(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	st := e.Attributes.Set[StrengthID]
	newAttrBonusTrait(e, "Increased Strength", StrengthID, fxp.Two, stlimit.None)
	newAttrBonusTrait(e, "Lifting Harness", StrengthID, fxp.Three, stlimit.LiftingOnly)
	e.Recalculate()
	c.Equal("Includes modifiers from:\nIncreased Strength [+2]\nFor lifting only:\nLifting Harness [+3]",
		st.BonusTooltip(), "a limited bonus is listed under its own heading")
	c.Equal(fxp.Twelve, st.Maximum(), "a limited bonus does not change the attribute's value")

	// Sections follow the limitation order, not the order the traits were added in.
	e = NewEntity()
	st = e.Attributes.Set[StrengthID]
	newAttrBonusTrait(e, "Throwing Arm", StrengthID, fxp.One, stlimit.ThrowingOnly)
	newAttrBonusTrait(e, "Lifting Harness", StrengthID, fxp.Two, stlimit.LiftingOnly)
	newAttrBonusTrait(e, "Striking Power", StrengthID, fxp.Three, stlimit.StrikingOnly)
	e.Recalculate()
	c.Equal("Includes modifiers from:\nFor striking only:\nStriking Power [+3]\nFor lifting only:\nLifting Harness [+2]\nFor throwing only:\nThrowing Arm [+1]",
		st.BonusTooltip(), "limited bonuses are grouped by limitation in a fixed order")
	c.Equal(fxp.Ten, st.Maximum(), "limited bonuses do not change the attribute's value")
}

// TestAttributeBonusTooltipPerLevel verifies that a per-level bonus shows both the total and the per-level amount, and
// that the leveled trait is named with its level.
func TestAttributeBonusTooltipPerLevel(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	bonus := NewAttributeBonus(StrengthID)
	bonus.Amount = fxp.One
	bonus.PerLevel = true
	trait := addTraitWithFeatures(e, "Visual Enhancement", bonus)
	trait.CanLevel = true
	trait.Levels = fxp.Three
	e.Recalculate()
	c.Equal("Includes modifiers from:\nVisual Enhancement 3 [+3 (+1 per level)]",
		e.Attributes.Set[StrengthID].BonusTooltip(), "a per-level bonus shows its total and its rate")
}

// TestAttributeBonusTooltipWithoutEntity verifies that an attribute that is not attached to an entity reports no
// bonuses rather than failing.
func TestAttributeBonusTooltipWithoutEntity(t *testing.T) {
	c := check.New(t)
	c.Equal("", NewAttribute(nil, StrengthID, 0).BonusTooltip(), "a detached attribute has no tooltip")
}

// TestAttributeCurrentCarriesThresholdEffects verifies that the bonuses of a pool's current threshold, and multipliers
// wherever they come from, alter an attribute's current value alone: its maximum, which scripts read as "$id" and the
// attributes derived from it rest on, stays as it is, amounts are added before factors are applied, the result is
// rounded to whole unless the attribute allows decimals, a pool is left alone, and nothing changes while the pool is
// at another threshold.
func TestAttributeCurrentCarriesThresholdEffects(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name     string
		attrID   string
		features Features
		fromItem bool
		damage   fxp.Int
		maximum  fxp.Int
		current  fxp.Int
	}{
		{"amount added", DexterityID, Features{NewAttributeBonus(DexterityID)}, false, fxp.Seven, fxp.Ten, fxp.Eleven},
		{
			"amount before factor", DexterityID,
			Features{newAttrBonus(DexterityID, fxp.Two), newMultiplier(DexterityID, fxp.Half)},
			false, fxp.Seven,
			fxp.Ten, fxp.Six,
		},
		{"product rounded for a whole attribute", DexterityID, Features{newMultiplier(DexterityID, fxp.Half)}, false, fxp.Seven, fxp.Ten, fxp.Five},
		{"fraction kept for a decimal attribute", BasicSpeedID, Features{newMultiplier(BasicSpeedID, fxp.Half)}, false, fxp.Seven, fxp.Five, fxp.TwoAndAHalf},
		{"multiplier from an item", StrengthID, Features{newMultiplier(StrengthID, fxp.Two)}, true, 0, fxp.Ten, fxp.FromInteger(20)},
		{"pool left alone", HitPointsID, Features{newMultiplier(HitPointsID, fxp.Half), NewAttributeBonus(HitPointsID)}, false, fxp.Seven, fxp.Ten, fxp.Ten},
		{"not at the threshold", DexterityID, Features{newMultiplier(DexterityID, fxp.Half)}, false, fxp.Six, fxp.Ten, fxp.Ten},
		{"derived from another's maximum", "will", Features{newMultiplier(IntelligenceID, fxp.Half)}, false, fxp.Seven, fxp.Ten, fxp.Ten},
	} {
		e := NewEntity()
		if one.fromItem {
			addTraitWithFeatures(e, "Item", one.features...)
		} else {
			tired := poolThreshold(e, FatiguePointsID, "Tired")
			tired.Features = append(tired.Features, one.features...)
		}
		e.Attributes.Set[FatiguePointsID].Damage = one.damage
		e.Recalculate()
		attr := e.Attributes.Set[one.attrID]
		c.Equal(one.maximum, attr.Maximum(), "%s: maximum", one.name)
		c.Equal(one.current, attr.Current(), "%s: current", one.name)
		c.Equal(one.maximum.String(), e.ResolveVariable(one.attrID), "%s: $%s", one.name, one.attrID)
		c.Equal(one.current.String(), e.ResolveVariable(one.attrID+".current"), "%s: $%s.current", one.name, one.attrID)
	}
}

// TestAttributeCurrentOfStandInReadsCurrents verifies that the current value of an attribute standing in for Striking
// ST, Lifting ST, Dodge or Move rests on the current values of the attributes its base reads as "$id", but not on those
// that their own bases read, so that the sheet gets from it what it gets without it, while its maximum, which scripts
// read as "$id", rests on their maximums.
func TestAttributeCurrentOfStandInReadsCurrents(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name    string
		attrID  string
		base    string
		source  string
		value   func(e *Entity) int
		maximum fxp.Int
		current fxp.Int
	}{
		{
			"striking ST", StrikingStrengthID, "$st", StrengthID,
			func(e *Entity) int { return e.StrikingStrength().AsInteger[int]() }, fxp.Ten, fxp.Five,
		},
		{
			"lifting ST", LiftingStrengthID, "$st", StrengthID,
			func(e *Entity) int { return e.LiftingStrength().AsInteger[int]() }, fxp.Ten, fxp.Five,
		},
		{
			"dodge", DodgeID, "Math.floor($basic_speed) + 3", BasicSpeedID,
			func(e *Entity) int { return e.Dodge(encumbrance.No) }, fxp.Eight, fxp.Three,
		},
		{
			"dodge, from a Basic Speed derived from DX's maximum", DodgeID, "Math.floor($basic_speed) + 3", DexterityID,
			func(e *Entity) int { return e.Dodge(encumbrance.No) }, fxp.Eight, fxp.Four,
		},
		{
			"move", MoveID, "$basic_move", BasicMoveID,
			func(e *Entity) int { return e.Move(encumbrance.No) }, fxp.Five, fxp.Two,
		},
	} {
		e := NewEntity()
		tired := poolThreshold(e, FatiguePointsID, "Tired")
		tired.Features = append(tired.Features, newMultiplier(one.source, fxp.Half))
		e.Attributes.Set[FatiguePointsID].Damage = fxp.Seven
		e.Recalculate()
		want := one.value(e)
		e.SheetSettings.Attributes.Set[one.attrID] = &AttributeDef{
			DefID: one.attrID,
			Type:  attribute.Integer,
			Name:  one.name,
			Base:  one.base,
		}
		e.Attributes.Set[one.attrID] = NewAttribute(e, one.attrID, len(e.Attributes.Set))
		e.Recalculate()
		attr := e.Attributes.Set[one.attrID]
		c.Equal(one.maximum, attr.Maximum(), "%s: maximum", one.name)
		c.Equal(one.current, attr.Current(), "%s: current", one.name)
		c.Equal(want, one.value(e), "%s: as without the attribute", one.name)
		c.Equal(one.maximum.String(), e.ResolveVariable(one.attrID), "%s: $%s", one.name, one.attrID)
		c.Equal(one.current.String(), e.ResolveVariable(one.attrID+".current"), "%s: $%s.current", one.name, one.attrID)
	}
}

// newAttrBonus returns an attribute bonus adding the given amount to the given attribute.
func newAttrBonus(attrID string, amount fxp.Int) *AttributeBonus {
	bonus := NewAttributeBonus(attrID)
	bonus.Amount = amount
	return bonus
}

// TestAttributeBonusTooltipThresholdSources verifies that the bonuses of a pool's current threshold are listed after
// the items' in each section, named for the threshold and its pool, that a pool lists none, and that the tooltip of a
// value figured from the current values of attributes, such as lifting ST, Move, Dodge or an attribute standing in for
// one, lists after its own bonuses those altering only the current value of each such attribute, under its name.
func TestAttributeBonusTooltipThresholdSources(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	st := e.Attributes.Set[StrengthID]
	newAttrBonusTrait(e, "Lifting Harness", StrengthID, fxp.Three, stlimit.LiftingOnly)
	newAttrBonusTrait(e, "Fleet", MoveID, fxp.Two, stlimit.None)
	tired := poolThreshold(e, FatiguePointsID, "Tired")
	tired.Features = append(tired.Features, newAttrBonus(StrengthID, -fxp.One), newMultiplier(BasicMoveID, fxp.Half),
		newMultiplier(BasicSpeedID, fxp.Half))
	hp := NewAttributeBonus(HitPointsID)
	hp.Amount = -fxp.Two
	tired.Features = append(tired.Features, hp)
	e.SheetSettings.Attributes.Set[StrikingStrengthID] = &AttributeDef{
		DefID: StrikingStrengthID,
		Type:  attribute.Integer,
		Name:  "Striking ST",
		Base:  "$st",
	}
	e.Attributes.Set[StrikingStrengthID] = NewAttribute(e, StrikingStrengthID, len(e.Attributes.Set))
	e.Attributes.Set[FatiguePointsID].Damage = fxp.Seven
	e.Recalculate()
	c.Equal("Includes modifiers from:\nTired (FP) [-1]\nFor lifting only:\nLifting Harness [+3]\nTired (FP) [×0.5]",
		st.BonusTooltip(), "the threshold's bonuses follow the items' in each section")
	c.Equal("", e.Attributes.Set[HitPointsID].BonusTooltip(), "a pool lists no threshold bonus, which doesn't alter it")
	c.Equal("Includes modifiers from:\nST:\nTired (FP) [-1]", e.Attributes.Set[StrikingStrengthID].BonusTooltip(),
		"a stand-in names what alters the current value of the attribute its base reads")
	c.Equal("Lifting Harness [+3]\nTired (FP) [×0.5]\nST:\nTired (FP) [-1]",
		e.CurrentAdjustmentTooltip(StrengthID, stlimit.LiftingOnly))
	c.Equal("Fleet [+2]\nTired (FP) [×0.5]\nBasic Move:\nTired (FP) [×0.5]",
		e.CurrentAdjustmentTooltip(MoveID, stlimit.None))
	c.Equal("Tired (FP) [×0.5]\nBasic Speed:\nTired (FP) [×0.5]", e.CurrentAdjustmentTooltip(DodgeID, stlimit.None))
	c.Equal("", e.CurrentAdjustmentTooltip(DexterityID, stlimit.None))
}
