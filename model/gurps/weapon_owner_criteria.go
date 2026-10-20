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
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

// WeaponOwnerCriteria holds the criteria a feature selecting weapons by name applies to the qualifiers of each
// weapon's owner other than its name. The level is that of a leveled trait or piece of equipment, and 0 for any other
// owner, including a skill or spell. A missing TL or specialization compares as empty.
type WeaponOwnerCriteria struct {
	OwnerLevelCriteria                  criteria.Number `json:"owner_level,omitzero"`
	OwnerTechLevelCriteria              criteria.Text   `json:"owner_tech_level,omitzero"`
	OwnerSpecializationCriteria         criteria.Text   `json:"owner_specialization,omitzero"`
	OwnerOptionalSpecializationCriteria criteria.Text   `json:"owner_optional_specialization,omitzero"`
}

// matches returns true if the criteria match the owner.
func (c *WeaponOwnerCriteria) matches(replacements map[string]string, owner WeaponOwner) bool {
	var level fxp.Int
	var tl, specialization, optionalSpecialization string
	switch one := owner.(type) {
	case *Trait:
		// Only worked out when needed, since the current level of a trait takes its bonuses into account.
		if c.OwnerLevelCriteria.Compare != criteria.AnyNumber {
			level = one.CurrentLevel()
		}
	case *Equipment:
		level = one.CurrentLevel()
		tl = one.TL()
	case *Skill:
		tl = one.TL()
		specialization = one.SpecializationWithReplacements()
		optionalSpecialization = one.OptionalSpecializationWithReplacements()
	case *Spell:
		tl = one.TL()
	}
	return c.OwnerLevelCriteria.Matches(level) &&
		c.OwnerTechLevelCriteria.Matches(replacements, tl) &&
		c.OwnerSpecializationCriteria.Matches(replacements, specialization) &&
		c.OwnerOptionalSpecializationCriteria.Matches(replacements, optionalSpecialization)
}

// fillWithNameableKeys adds any nameable keys found in the criteria to the provided map.
func (c *WeaponOwnerCriteria) fillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(m, existing, c.OwnerTechLevelCriteria.Qualifier, c.OwnerSpecializationCriteria.Qualifier,
		c.OwnerOptionalSpecializationCriteria.Qualifier)
}

// specificity returns the number of criteria in use.
func (c *WeaponOwnerCriteria) specificity() int {
	count := 0
	for _, one := range []criteria.Text{
		c.OwnerTechLevelCriteria, c.OwnerSpecializationCriteria,
		c.OwnerOptionalSpecializationCriteria,
	} {
		if !one.IsZero() {
			count++
		}
	}
	if !c.OwnerLevelCriteria.IsZero() {
		count++
	}
	return count
}

// describe returns the clauses describing the criteria in use, e.g. ` whose owner's level is at least 2`.
func (c *WeaponOwnerCriteria) describe(replacements map[string]string, em func(string) string) string {
	var text string
	if c.OwnerLevelCriteria.Compare != criteria.AnyNumber {
		text = i18n.Text(" whose owner's level ") + c.OwnerLevelCriteria.String()
	}
	// An empty "is" picks only an owner without the qualifier, as for a skill's specialization.
	whose := func(t criteria.Text, whose, without string) string {
		if t.Compare == criteria.IsText && strings.TrimSpace(nameable.Apply(t.Qualifier, replacements)) == "" {
			return " " + without
		}
		return describeWhose(whose, t, replacements, em)
	}
	return text + whose(c.OwnerTechLevelCriteria, i18n.Text("whose owner's tech level"),
		i18n.Text("whose owner has no tech level")) +
		whose(c.OwnerSpecializationCriteria, i18n.Text("whose owner's specialization"),
			i18n.Text("whose owner has no specialization")) +
		whose(c.OwnerOptionalSpecializationCriteria, i18n.Text("whose owner's optional specialization"),
			i18n.Text("whose owner has no optional specialization"))
}

// Hash writes this object's contents into the hasher.
func (c *WeaponOwnerCriteria) Hash(h hash.Hash) {
	c.OwnerLevelCriteria.Hash(h)
	c.OwnerTechLevelCriteria.Hash(h)
	c.OwnerSpecializationCriteria.Hash(h)
	c.OwnerOptionalSpecializationCriteria.Hash(h)
}
