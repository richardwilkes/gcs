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
	"maps"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/toolbox/v2/check"
)

// nameableKeysOf returns the nameable keys the feature contributes, sorted.
func nameableKeysOf(f Feature) []string {
	m := make(map[string]string)
	f.FillWithNameableKeys(m, nil)
	return slices.Sorted(maps.Keys(m))
}

// TestSkillBonusNameableKeysBySelection verifies that only the criteria a skill bonus's selection consults contribute
// nameable keys. The specialization criterion holds the usage when the bonus selects weapons.
func TestSkillBonusNameableKeysBySelection(t *testing.T) {
	c := check.New(t)
	b := NewSkillBonus()
	b.NameCriteria.Qualifier = "@Name@"
	b.SpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "@Spec@"}
	b.OptionalSpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "@Opt@"}
	b.DamageTypeCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "@Type@"}
	b.TagsCriteria = criteria.Text{Compare: criteria.ContainsText, Qualifier: "@Tag@"}
	b.OwnerTechLevelCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "@Owner@"}
	for _, one := range []struct {
		selection skillsel.Type
		keys      []string
	}{
		{skillsel.Name, []string{"Name", "Opt", "Spec", "Tag"}},
		{skillsel.ThisWeapon, []string{"Spec", "Type"}},
		{skillsel.WeaponsWithName, []string{"Name", "Owner", "Spec", "Tag", "Type"}},
	} {
		b.SelectionType = one.selection
		c.Equal(one.keys, nameableKeysOf(b), "selecting %v", one.selection)
	}
}

// TestWeaponBonusNameableKeysBySelection verifies that only the criteria a weapon bonus's selection consults contribute
// nameable keys: the specialization criterion holds the usage except when selecting by required skill, where the usage
// criterion does, and the optional specialization is never consulted.
func TestWeaponBonusNameableKeysBySelection(t *testing.T) {
	c := check.New(t)
	b := NewWeaponBonus(feature.WeaponBonus)
	b.NameCriteria.Qualifier = "@Name@"
	b.SpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "@Spec@"}
	b.OptionalSpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "@Opt@"}
	b.UsageCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "@Usage@"}
	b.DamageTypeCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "@Type@"}
	b.TagsCriteria = criteria.Text{Compare: criteria.ContainsText, Qualifier: "@Tag@"}
	b.OwnerTechLevelCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "@Owner@"}
	for _, one := range []struct {
		selection wsel.Type
		keys      []string
	}{
		{wsel.WithRequiredSkill, []string{"Name", "Spec", "Tag", "Type", "Usage"}},
		{wsel.WithName, []string{"Name", "Owner", "Spec", "Tag", "Type"}},
		{wsel.ThisWeapon, []string{"Spec", "Type"}},
	} {
		b.SelectionType = one.selection
		c.Equal(one.keys, nameableKeysOf(b), "selecting %v", one.selection)
	}
}
