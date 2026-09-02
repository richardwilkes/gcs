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

// newAttrBonusTrait creates a trait carrying a single bonus to the given attribute with the given amount and
// limitation, and adds it to the entity.
func newAttrBonusTrait(e *Entity, name, attrID string, amount fxp.Int, limitation stlimit.Option) *Trait {
	bonus := NewAttributeBonus(attrID)
	bonus.Amount = amount
	bonus.Limitation = limitation
	trait := NewTrait(e, nil, false)
	trait.Name = name
	trait.Features = Features{bonus}
	e.Traits = append(e.Traits, trait)
	return trait
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
	eqp := NewEquipment(e, nil, false)
	eqp.Name = "Smart Gloves"
	eqp.Modifiers = []*EquipmentModifier{mod}
	e.CarriedEquipment = append(e.CarriedEquipment, eqp)

	e.Recalculate()
	c.Equal("Includes modifiers from:\nIncreased Strength [+2]\nSmart Gloves (AR software) [+1]", st.BonusTooltip(),
		"each active bonus is listed with its source and amount")
	c.Equal(fxp.Three, st.Bonus, "the tooltip reflects the bonus that was applied")

	trait.Disabled = true
	e.Recalculate()
	c.Equal("Includes modifiers from:\nSmart Gloves (AR software) [+1]", st.BonusTooltip(),
		"a disabled trait's bonus is not listed")

	eqp.Equipped = false
	e.Recalculate()
	c.Equal("", st.BonusTooltip(), "an unequipped item's bonus is not listed")

	trait.Disabled = false
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
	trait.Modifiers = []*TraitModifier{mod}

	e.Recalculate()
	c.Equal("Includes modifiers from:\nIncreased Strength [+2]\nIncreased Strength (Cybernetic) [+1]",
		st.BonusTooltip(), "a trait modifier's bonus is attributed to both the trait and the modifier")
	c.Equal(fxp.Three, st.Bonus, "the tooltip reflects the bonus that was applied")

	mod.Disabled = true
	e.Recalculate()
	c.Equal("Includes modifiers from:\nIncreased Strength [+2]", st.BonusTooltip(),
		"a disabled modifier's bonus is not listed")
	c.Equal(fxp.Two, st.Bonus, "a disabled modifier's bonus is not applied")
}

// TestAttributeBonusTooltipNonStrengthLimitation verifies that a limitation left over on a bonus to something other
// than ST is ignored: the feature editor never clears the limitation when the attribute is switched away from ST, so
// such a bonus must still be applied and listed under the plain heading rather than being filed away under -- or
// dropped along with -- a section that only ST has.
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
	trait := NewTrait(e, nil, false)
	trait.Name = "Visual Enhancement"
	trait.CanLevel = true
	trait.Levels = fxp.Three
	trait.Features = Features{bonus}
	e.Traits = append(e.Traits, trait)
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
