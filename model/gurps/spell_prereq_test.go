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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellcmp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// addTestCollegeSpell creates a non-container spell owned by the entity with the given name and college, one point, then
// appends it to the entity's spell list.
func addTestCollegeSpell(e *Entity, name, college string) *Spell {
	s := addTestSpell(e, name, fxp.One)
	s.College = CollegeList{college}
	return s
}

// addSpellNamePrereq attaches a "requires a spell named target" prerequisite to the supplied spell.
func addSpellNamePrereq(s *Spell, target string) {
	p := NewSpellPrereq()
	p.SubType = spellcmp.Name
	p.QualifierCriteria.Compare = criteria.IsText
	p.QualifierCriteria.Qualifier = target
	s.Prereq = NewPrereqList()
	s.Prereq.Prereqs = append(s.Prereq.Prereqs, p)
	p.Parent = s.Prereq
}

// TestSpellPrereqCircularNotCounted verifies that a spell which itself requires the spell being checked is not counted
// toward that spell's own college prerequisite, which would otherwise create a circular prerequisite relationship.
// See GitHub issue #737.
func TestSpellPrereqCircularNotCounted(t *testing.T) {
	c := check.New(t)

	e := NewEntity()

	// "Wisdom" requires at least 5 spells whose college contains "Mind Control".
	wisdom := addTestCollegeSpell(e, "Wisdom", "Mind Control")
	collegeReq := NewSpellPrereq()
	collegeReq.SubType = spellcmp.College
	collegeReq.QualifierCriteria.Compare = criteria.ContainsText
	collegeReq.QualifierCriteria.Qualifier = "Mind Control"
	collegeReq.QuantityCriteria.Compare = criteria.AtLeastNumber
	collegeReq.QuantityCriteria.Qualifier = fxp.FromInteger(5)
	wisdom.Prereq = NewPrereqList()
	wisdom.Prereq.Prereqs = append(wisdom.Prereq.Prereqs, collegeReq)
	collegeReq.Parent = wisdom.Prereq

	// "Boost Intelligence" is a Mind Control spell that directly requires "Wisdom".
	boost := addTestCollegeSpell(e, "Boost Intelligence", "Mind Control")
	addSpellNamePrereq(boost, "Wisdom")

	// Four additional, non-circular Mind Control spells.
	addTestCollegeSpell(e, "Mind A", "Mind Control")
	addTestCollegeSpell(e, "Mind B", "Mind Control")
	addTestCollegeSpell(e, "Mind C", "Mind Control")
	addTestCollegeSpell(e, "Mind D", "Mind Control")

	// The direct-prerequisite helper must recognize the circular relationship.
	c.True(spellDirectlyRequires(boost, wisdom), "Boost Intelligence directly requires Wisdom")
	c.False(spellDirectlyRequires(e.Spells[2], wisdom), "Mind A does not require Wisdom")

	// There are five Mind Control spells besides Wisdom (Boost Intelligence + four plain ones), but Boost Intelligence
	// must not be counted because it requires Wisdom. That leaves only four, so the "at least 5" requirement is not met.
	c.False(collegeReq.Satisfied(e, wisdom, nil, "", nil),
		"Wisdom's college prerequisite must not be satisfied once the circular spell is excluded")

	// Adding a fifth non-circular Mind Control spell brings the count back up to five and satisfies the requirement.
	addTestCollegeSpell(e, "Mind E", "Mind Control")
	c.True(collegeReq.Satisfied(e, wisdom, nil, "", nil),
		"Wisdom's college prerequisite must be satisfied once a fifth non-circular spell is present")
}

// TestSpellPrereqNestedCircularNotCounted verifies that a circular relationship expressed inside a nested prereq list is
// still detected.
func TestSpellPrereqNestedCircularNotCounted(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	wisdom := addTestCollegeSpell(e, "Wisdom", "Mind Control")
	boost := addTestCollegeSpell(e, "Boost Intelligence", "Mind Control")

	// Nest the spell-name prerequisite one level down inside an "any of" sub-list.
	inner := NewPrereqList()
	inner.All = false
	namePrereq := NewSpellPrereq()
	namePrereq.SubType = spellcmp.Name
	namePrereq.QualifierCriteria.Compare = criteria.IsText
	namePrereq.QualifierCriteria.Qualifier = "Wisdom"
	inner.Prereqs = append(inner.Prereqs, namePrereq)
	namePrereq.Parent = inner
	boost.Prereq = NewPrereqList()
	boost.Prereq.Prereqs = append(boost.Prereq.Prereqs, inner)
	inner.Parent = boost.Prereq

	c.True(spellDirectlyRequires(boost, wisdom), "nested spell-name prerequisite must be detected")
}
