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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/toolbox/v2/check"
)

func addTestSkill(e *Entity, name, specialization, tl string, points fxp.Int) *Skill {
	sk := NewSkill(e, nil, false)
	sk.Name = name
	sk.Specialization = specialization
	sk.Difficulty.Attribute = IntelligenceID
	sk.Difficulty.Difficulty = difficulty.Average
	sk.Points = points
	if tl != "" {
		tlv := tl
		sk.TechLevel = &tlv
	}
	e.Skills = append(e.Skills, sk)
	return sk
}

func newWhenTLSkillDefault(name, specialization string, when criteria.Number) *SkillDefault {
	def := &SkillDefault{
		DefaultType: SkillID,
		Name:        criteria.Text{TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: name}},
		Modifier:    -fxp.Three,
		WhenTL:      when,
	}
	if specialization != "" {
		def.Specialization = criteria.Text{TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: specialization}}
	}
	return def
}

func numberCriteria(compare criteria.NumericComparison, qualifier fxp.Int) criteria.Number {
	return criteria.Number{NumberData: criteria.NumberData{Compare: compare, Qualifier: qualifier}}
}

// TestSkillDefaultWhenTL verifies that a skill default's WhenTL constraint is evaluated against the tech level of the
// skill being defaulted *from* (the matched skill), not the character's tech level. See issue #1040.
func TestSkillDefaultWhenTL(t *testing.T) {
	c := check.New(t)

	// Character is at TL5, but holds a TL3 Machinist and a TL5 Smith (Iron).
	e := NewEntity()
	e.Profile.TechLevel = "5"
	addTestSkill(e, "Machinist", "", "3", fxp.Four)
	addTestSkill(e, "Smith", "Iron", "5", fxp.Four)
	addTestSkill(e, "Carpentry", "", "", fxp.Four) // no tech level, so it falls back to the character's TL
	e.Recalculate()

	// The matched skill's TL fails the constraint, so the default must not resolve (the prior bug used the
	// character's TL of 5, which would have let these through).
	c.Equal(fxp.Min, newWhenTLSkillDefault("Machinist", "",
		numberCriteria(criteria.AtLeastNumber, fxp.Five)).SkillLevel(e, nil, true, nil, false),
		"Machinist (TL3) must fail 'when TL at least 5'")
	c.Equal(fxp.Min, newWhenTLSkillDefault("Smith", "Iron",
		numberCriteria(criteria.AtMostNumber, fxp.Four)).SkillLevel(e, nil, true, nil, false),
		"Smith/Iron (TL5) must fail 'when TL at most 4'")

	// The matched skill's TL satisfies the constraint, so the default resolves to a real level.
	c.NotEqual(fxp.Min, newWhenTLSkillDefault("Machinist", "",
		numberCriteria(criteria.AtMostNumber, fxp.Four)).SkillLevel(e, nil, true, nil, false),
		"Machinist (TL3) must satisfy 'when TL at most 4'")
	c.NotEqual(fxp.Min, newWhenTLSkillDefault("Smith", "Iron",
		numberCriteria(criteria.AtLeastNumber, fxp.Five)).SkillLevel(e, nil, true, nil, false),
		"Smith/Iron (TL5) must satisfy 'when TL at least 5'")

	// A default with no constraint always resolves against the matched skill.
	c.NotEqual(fxp.Min, newWhenTLSkillDefault("Machinist", "",
		criteria.Number{}).SkillLevel(e, nil, true, nil, false),
		"Machinist must resolve when there is no TL constraint")

	// SkillLevelFast (used for weapon defaults) must apply the same filtering.
	c.Equal(fxp.Min, newWhenTLSkillDefault("Machinist", "",
		numberCriteria(criteria.AtLeastNumber, fxp.Five)).SkillLevelFast(e, nil, true, nil, false),
		"SkillLevelFast: Machinist (TL3) must fail 'when TL at least 5'")
	c.NotEqual(fxp.Min, newWhenTLSkillDefault("Machinist", "",
		numberCriteria(criteria.AtMostNumber, fxp.Four)).SkillLevelFast(e, nil, true, nil, false),
		"SkillLevelFast: Machinist (TL3) must satisfy 'when TL at most 4'")

	// When the matched skill has no TL of its own, the constraint falls back to the character's TL (5 here).
	c.NotEqual(fxp.Min, newWhenTLSkillDefault("Carpentry", "",
		numberCriteria(criteria.AtLeastNumber, fxp.Five)).SkillLevel(e, nil, true, nil, false),
		"Carpentry (no TL) must use the character's TL of 5")
	c.Equal(fxp.Min, newWhenTLSkillDefault("Carpentry", "",
		numberCriteria(criteria.AtLeastNumber, fxp.Six)).SkillLevel(e, nil, true, nil, false),
		"Carpentry (no TL) must use the character's TL of 5, which fails 'at least 6'")
}

// TestSkillDefaultWhenTLChosenDefault reproduces issue #1040 end-to-end: an Armory (Body Armor) skill must not default
// to a skill whose tech level fails the default's WhenTL constraint, even after the character's TL is raised.
func TestSkillDefaultWhenTLChosenDefault(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	e.Profile.TechLevel = "5"
	addTestSkill(e, "Machinist", "", "3", fxp.Four)
	addTestSkill(e, "Smith", "Iron", "5", fxp.Four)

	armory := addTestSkill(e, "Armory", "Body Armor", "3", 0)
	armory.Defaults = []*SkillDefault{
		{
			DefaultType: IntelligenceID,
			Modifier:    -fxp.Five,
		},
		newWhenTLSkillDefault("Machinist", "", numberCriteria(criteria.AtLeastNumber, fxp.Five)),
		newWhenTLSkillDefault("Smith", "Iron", numberCriteria(criteria.AtMostNumber, fxp.Four)),
	}
	e.Recalculate()

	// Neither the TL3 Machinist (fails "at least 5") nor the TL5 Smith/Iron (fails "at most 4") is eligible, so the
	// Armory skill must fall back to its IQ-5 default rather than a skill default.
	c.NotNil(armory.DefaultedFrom, "Armory should have a resolved default")
	c.False(armory.DefaultedFrom.SkillBased(), "Armory must not default to a skill whose TL fails the constraint")
	c.Equal(IntelligenceID, armory.DefaultedFrom.DefaultType, "Armory must default to IQ")
}
