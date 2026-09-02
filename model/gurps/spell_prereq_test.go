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
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// addTestCollegeSpell creates a non-container spell owned by the entity with the given name and college, one point, then
// appends it to the entity's spell list.
func addTestCollegeSpell(e *Entity, name, college string) *Spell {
	s := addTestSpell(e, name, fxp.One)
	s.College = CollegeList{college}
	return s
}

// addSpellNamePrereq attaches a "requires a spell named target" prerequisite to the supplied spell and returns the
// prerequisite that was created.
func addSpellNamePrereq(s *Spell, target string) *SpellPrereq {
	p := NewSpellPrereq()
	p.SubType = spellcmp.Name
	p.QualifierCriteria.Compare = criteria.IsText
	p.QualifierCriteria.Qualifier = target
	s.Prereq = NewPrereqList()
	s.Prereq.Prereqs = append(s.Prereq.Prereqs, p)
	p.Parent = s.Prereq
	return p
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

// TestSpellPrereqNilEntity verifies that a nil entity is treated as satisfied rather than panicking, matching every
// other prereq implementation.
func TestSpellPrereqNilEntity(t *testing.T) {
	c := check.New(t)
	for _, subType := range []spellcmp.Type{
		spellcmp.Name,
		spellcmp.Tag,
		spellcmp.College,
		spellcmp.CollegeCount,
		spellcmp.Any,
	} {
		p := NewSpellPrereq()
		p.SubType = subType
		p.QualifierCriteria.Compare = criteria.IsText
		p.QualifierCriteria.Qualifier = "Wisdom"
		var tooltip xbytes.InsertBuffer
		c.NotPanics(func() {
			c.True(p.Satisfied(nil, nil, &tooltip, "", nil), "%v: a nil entity must be treated as satisfied", subType)
		}, "%v: a nil entity must not panic", subType)
		c.Equal("", tooltip.String(), "%v: no tooltip should be written for a nil entity", subType)
	}
}

// addTestPowerSourceSpell creates a non-container spell owned by the entity with the given name and power source, one
// point, then appends it to the entity's spell list.
func addTestPowerSourceSpell(e *Entity, name, powerSource string) *Spell {
	s := addTestSpell(e, name, fxp.One)
	s.PowerSource = powerSource
	return s
}

// TestSpellPrereqSamePowerSource verifies that the default "same power source as this spell" filter counts only spells
// whose power source matches the owning spell's, per B77. See GitHub issue #1093.
func TestSpellPrereqSamePowerSource(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	owner := addTestPowerSourceSpell(e, "Ignite Fire", "Arcane")
	addTestPowerSourceSpell(e, "Fireball", "Arcane")
	addTestPowerSourceSpell(e, "Fireball", "Clerical")

	p := NewSpellPrereq()
	p.QualifierCriteria.Qualifier = "Fireball"
	c.True(p.Satisfied(e, owner, nil, "", nil), "the Arcane Fireball satisfies an Arcane spell's prerequisite")

	p.QuantityCriteria.Qualifier = fxp.Two
	c.False(p.Satisfied(e, owner, nil, "", nil), "the Clerical Fireball must not count toward an Arcane spell")

	// The same prerequisite on a Clerical spell finds the Clerical Fireball instead.
	owner.PowerSource = "Clerical"
	p.QuantityCriteria.Qualifier = fxp.One
	c.True(p.Satisfied(e, owner, nil, "", nil), "the Clerical Fireball satisfies a Clerical spell's prerequisite")
	p.QuantityCriteria.Qualifier = fxp.Two
	c.False(p.Satisfied(e, owner, nil, "", nil), "the Arcane Fireball must not count toward a Clerical spell")

	// A spell with no power source matches only other spells with no power source.
	owner.PowerSource = ""
	p.QuantityCriteria.Qualifier = fxp.One
	c.False(p.Satisfied(e, owner, nil, "", nil),
		"a spell with no power source must not be satisfied by spells that have one")
	addTestPowerSourceSpell(e, "Fireball", "")
	c.True(p.Satisfied(e, owner, nil, "", nil),
		"a spell with no power source is satisfied by another spell with no power source")
}

// TestSpellPrereqSamePowerSourceNonSpellOwner verifies that a "same power source as this spell" filter attached to
// something that isn't a spell matches nothing: the prerequisite is unsatisfied and its tooltip names the filter that
// went unmet, while the negated form of the same prerequisite is satisfied.
func TestSpellPrereqSamePowerSourceNonSpellOwner(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	addTestPowerSourceSpell(e, "Fireball", "Arcane")
	owner := NewTrait(e, nil, false)

	p := NewSpellPrereq()
	p.QualifierCriteria.Qualifier = "Fireball"
	var tooltip xbytes.InsertBuffer
	c.False(p.Satisfied(e, owner, &tooltip, "", nil), "a non-spell owner has no power source to match against")
	c.Equal(`Has at least 1 spell whose name is "Fireball" and whose power source is the same as this spell's`,
		tooltip.String(), "the tooltip must name the unmet power source filter")

	p.Has = false
	c.True(p.Satisfied(e, owner, nil, "", nil), "the negated form of the same prerequisite is satisfied")
}

// TestSpellPrereqPowerSourceCriteria verifies that an explicit power source criterion filters the spells that count,
// and that a prerequisite with no power source filter at all counts them all.
func TestSpellPrereqPowerSourceCriteria(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	owner := addTestPowerSourceSpell(e, "Ignite Fire", "Arcane")
	addTestPowerSourceSpell(e, "Fireball", "Arcane")
	addTestPowerSourceSpell(e, "Fireball", "Clerical")

	p := NewSpellPrereq()
	p.QualifierCriteria.Qualifier = "Fireball"
	p.SamePowerSource = false
	p.PowerSourceCriteria.Compare = criteria.IsText
	p.PowerSourceCriteria.Qualifier = "Arcane"
	c.True(p.Satisfied(e, owner, nil, "", nil), "the Arcane Fireball satisfies an explicit Arcane power source")
	p.QuantityCriteria.Qualifier = fxp.Two
	c.False(p.Satisfied(e, owner, nil, "", nil), "the Clerical Fireball must not count as an Arcane one")

	// With no power source filter, both spells count, as they did before this option existed.
	p.PowerSourceCriteria.Compare = criteria.AnyText
	c.True(p.Satisfied(e, owner, nil, "", nil), "without a power source filter, both Fireballs count")
}

// TestSpellPrereqPowerSourceWithAnyAndCollegeCount verifies that the power source filter applies to the "any" and
// "college count" sub-types as well, with the college count only gathering colleges from spells that pass the filter.
func TestSpellPrereqPowerSourceWithAnyAndCollegeCount(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	owner := addTestCollegeSpell(e, "Ignite Fire", "Fire")
	addTestCollegeSpell(e, "Flame Jet", "Fire")
	bless := addTestCollegeSpell(e, "Bless", "Holy")
	bless.PowerSource = "Clerical"
	cure := addTestCollegeSpell(e, "Cure Disease", "Healing")
	cure.PowerSource = "Clerical"

	for _, subType := range []spellcmp.Type{spellcmp.Any, spellcmp.CollegeCount} {
		// Only the one other Arcane spell (or the one college it belongs to) counts for an Arcane spell.
		p := NewSpellPrereq()
		p.SubType = subType
		c.True(p.Satisfied(e, owner, nil, "", nil), "%v: the other Arcane spell counts", subType)
		p.QuantityCriteria.Qualifier = fxp.Two
		c.False(p.Satisfied(e, owner, nil, "", nil), "%v: the Clerical spells must not count", subType)

		// An explicit Clerical criterion counts the two Clerical spells (or their two colleges) instead.
		p = NewSpellPrereq()
		p.SubType = subType
		p.SamePowerSource = false
		p.PowerSourceCriteria.Compare = criteria.IsText
		p.PowerSourceCriteria.Qualifier = "Clerical"
		p.QuantityCriteria.Qualifier = fxp.Two
		c.True(p.Satisfied(e, owner, nil, "", nil), "%v: both Clerical spells count", subType)
		p.QuantityCriteria.Qualifier = fxp.Three
		c.False(p.Satisfied(e, owner, nil, "", nil), "%v: the Arcane spells must not count", subType)
	}
}

// addTestFireSpell creates a non-container spell owned by the entity with the given name and power source, one point,
// belonging to the "Fire" college and tagged "Fire", then appends it to the entity's spell list.
func addTestFireSpell(e *Entity, name, powerSource string) *Spell {
	s := addTestPowerSourceSpell(e, name, powerSource)
	s.College = CollegeList{"Fire"}
	s.Tags = []string{"Fire"}
	return s
}

// TestSpellPrereqPowerSourceWithTagAndCollege verifies that the power source filter applies to the "tag" and "college"
// sub-types as well, which count a spell when any one of its tags or colleges matches the qualifier.
func TestSpellPrereqPowerSourceWithTagAndCollege(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	owner := addTestFireSpell(e, "Ignite Fire", "Arcane")
	addTestFireSpell(e, "Flame Jet", "Arcane")
	addTestFireSpell(e, "Holy Fire", "Clerical")
	addTestFireSpell(e, "Sacred Flame", "Clerical")

	for _, subType := range []spellcmp.Type{spellcmp.Tag, spellcmp.College} {
		// Only the one other Arcane spell counts for an Arcane spell.
		p := NewSpellPrereq()
		p.SubType = subType
		p.QualifierCriteria.Qualifier = "Fire"
		c.True(p.Satisfied(e, owner, nil, "", nil), "%v: the other Arcane spell counts", subType)
		p.QuantityCriteria.Qualifier = fxp.Two
		c.False(p.Satisfied(e, owner, nil, "", nil), "%v: the Clerical spells must not count", subType)

		// An explicit Clerical criterion counts the two Clerical spells instead.
		p.SamePowerSource = false
		p.PowerSourceCriteria.Compare = criteria.IsText
		p.PowerSourceCriteria.Qualifier = "Clerical"
		c.True(p.Satisfied(e, owner, nil, "", nil), "%v: both Clerical spells count", subType)
		p.QuantityCriteria.Qualifier = fxp.Three
		c.False(p.Satisfied(e, owner, nil, "", nil), "%v: the Arcane spells must not count", subType)
	}
}

// TestSpellPrereqPowerSourceTooltip verifies the text produced for an unsatisfied prerequisite, with and without a
// power source filter. The prerequisite has no owning spell here, so the "same as this spell's" form cannot name a
// power source; TestSpellPrereqSamePowerSourceTooltipNamesOwner covers the case where it can.
func TestSpellPrereqPowerSourceTooltip(t *testing.T) {
	c := check.New(t)

	for _, one := range []struct {
		name      string
		subType   spellcmp.Type
		qualifier string
		quantity  fxp.Int
		same      bool
		compare   criteria.StringComparison
		source    string
		expected  string
	}{
		{
			name:      "name without a power source filter",
			subType:   spellcmp.Name,
			qualifier: "Fireball",
			quantity:  fxp.One,
			expected:  `Has at least 1 spell whose name is "Fireball"`,
		},
		{
			name:      "name with the same power source",
			subType:   spellcmp.Name,
			qualifier: "Fireball",
			quantity:  fxp.One,
			same:      true,
			expected:  `Has at least 1 spell whose name is "Fireball" and whose power source is the same as this spell's`,
		},
		{
			name:      "name with an explicit power source",
			subType:   spellcmp.Name,
			qualifier: "Fireball",
			quantity:  fxp.One,
			compare:   criteria.IsText,
			source:    "Clerical",
			expected:  `Has at least 1 spell whose name is "Fireball" and whose power source is "Clerical"`,
		},
		{
			name:      "tag with the same power source",
			subType:   spellcmp.Tag,
			qualifier: "Fire",
			quantity:  fxp.One,
			same:      true,
			expected:  `Has at least 1 spell whose tag is "Fire" and whose power source is the same as this spell's`,
		},
		{
			name:      "college with an explicit power source",
			subType:   spellcmp.College,
			qualifier: "Fire",
			quantity:  fxp.Two,
			compare:   criteria.IsText,
			source:    "Clerical",
			expected:  `Has at least 2 spells whose college is "Fire" and whose power source is "Clerical"`,
		},
		{
			name:     "any with the same power source",
			subType:  spellcmp.Any,
			quantity: fxp.Three,
			same:     true,
			expected: "Has at least 3 spells whose power source is the same as this spell's",
		},
		{
			name:     "any without a power source filter",
			subType:  spellcmp.Any,
			quantity: fxp.Three,
			expected: "Has at least 3 spells of any kind",
		},
		{
			name:     "college count with an explicit power source",
			subType:  spellcmp.CollegeCount,
			quantity: fxp.Three,
			compare:  criteria.IsText,
			source:   "Clerical",
			expected: `Has at least 3 spells from different colleges whose power source is "Clerical"`,
		},
		{
			name:     "college count without a power source filter",
			subType:  spellcmp.CollegeCount,
			quantity: fxp.Three,
			expected: "Has at least 3 spells from different colleges",
		},
	} {
		p := NewSpellPrereq()
		p.SubType = one.subType
		p.QualifierCriteria.Qualifier = one.qualifier
		p.QuantityCriteria.Qualifier = one.quantity
		p.SamePowerSource = one.same
		p.PowerSourceCriteria.Compare = one.compare
		p.PowerSourceCriteria.Qualifier = one.source
		var tooltip xbytes.InsertBuffer
		c.False(p.Satisfied(NewEntity(), nil, &tooltip, "", nil), "%s: an entity with no spells cannot satisfy this",
			one.name)
		c.Equal(one.expected, tooltip.String(), "%s: unexpected tooltip", one.name)
	}
}

// TestSpellPrereqSamePowerSourceTooltipNamesOwner verifies that the "same as this spell's" tooltip names the owning
// spell's resolved power source, so that a spell whose power source is blank or spelled differently from the spells it
// expects to match can be diagnosed from the tooltip alone. See GitHub issue #1093.
func TestSpellPrereqSamePowerSourceTooltipNamesOwner(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	owner := addTestPowerSourceSpell(e, "Ignite Fire", "Arcane")
	addTestPowerSourceSpell(e, "Fireball", "Magical")

	p := NewSpellPrereq()
	p.QualifierCriteria.Qualifier = "Fireball"
	var tooltip xbytes.InsertBuffer
	c.False(p.Satisfied(e, owner, &tooltip, "", nil), "a Magical Fireball must not count toward an Arcane spell")
	c.Equal(`Has at least 1 spell whose name is "Fireball" and whose power source is the same as this spell's ("Arcane")`,
		tooltip.String(), "the tooltip must name the owning spell's power source")

	owner.PowerSource = ""
	tooltip.Reset()
	c.False(p.Satisfied(e, owner, &tooltip, "", nil),
		"a Magical Fireball must not count toward a spell with no power source")
	c.Equal(`Has at least 1 spell whose name is "Fireball" and whose power source is the same as this spell's ("")`,
		tooltip.String(), "the tooltip must show that the owning spell has no power source")
}

// TestSpellPrereqCircularRespectsPowerSource verifies that the circular prerequisite guard added for GitHub issue #737
// now takes the power source filter into account: a spell that requires a differently-sourced spell of the same name is
// not treated as requiring this one, so it can still count toward this spell's own prerequisites, provided those accept
// its power source.
func TestSpellPrereqCircularRespectsPowerSource(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	arcaneFireball := addTestPowerSourceSpell(e, "Fireball", "Arcane")
	clericalFireball := addTestPowerSourceSpell(e, "Fireball", "Clerical")
	holy := addTestPowerSourceSpell(e, "Holy Wrath", "Clerical")
	p := addSpellNamePrereq(holy, "Fireball")

	// "Holy Wrath" requires a Clerical "Fireball", so it does not require the Arcane one.
	c.False(spellDirectlyRequires(holy, arcaneFireball), "the Arcane Fireball is not required by Holy Wrath")
	c.True(spellDirectlyRequires(holy, clericalFireball), "the Clerical Fireball is required by Holy Wrath")

	// So Holy Wrath still counts toward the Arcane Fireball's own prerequisites, but not toward the Clerical
	// Fireball's. For that to matter, those prerequisites have to accept a Clerical spell in the first place, so give
	// them no power source filter of their own, as a prerequisite from a file written before the filter existed has.
	arcaneReq := addSpellNamePrereq(arcaneFireball, "Holy Wrath")
	arcaneReq.SamePowerSource = false
	c.True(arcaneReq.Satisfied(e, arcaneFireball, nil, "", nil),
		"Holy Wrath does not require the Arcane Fireball, so it counts toward the Arcane Fireball's prerequisite")
	clericalReq := addSpellNamePrereq(clericalFireball, "Holy Wrath")
	clericalReq.SamePowerSource = false
	c.False(clericalReq.Satisfied(e, clericalFireball, nil, "", nil),
		"Holy Wrath requires the Clerical Fireball, so it must not count toward the Clerical Fireball's prerequisite")

	// The explicit form of the filter behaves the same way.
	p.SamePowerSource = false
	p.PowerSourceCriteria.Compare = criteria.IsText
	p.PowerSourceCriteria.Qualifier = "Clerical"
	c.False(spellDirectlyRequires(holy, arcaneFireball),
		"the Arcane Fireball is not required by an explicitly Clerical prerequisite")
	c.True(spellDirectlyRequires(holy, clericalFireball),
		"the Clerical Fireball is required by an explicitly Clerical prerequisite")

	// With no power source filter, the original guard still excludes the spell.
	p.PowerSourceCriteria.Compare = criteria.AnyText
	p.PowerSourceCriteria.Qualifier = ""
	c.True(spellDirectlyRequires(holy, arcaneFireball),
		"without a power source filter, any Fireball is required by Holy Wrath")
}

// TestSpellPrereqPowerSourceNameableReplacements verifies that substitution markers are resolved while matching: an
// owning spell whose own power source is a marker compares its resolved value, a marker in the power source qualifier
// is resolved through the owning spell's replacements, and a candidate spell's marker is resolved through its own.
func TestSpellPrereqPowerSourceNameableReplacements(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	owner := addTestPowerSourceSpell(e, "Ignite Fire", "@source@")
	owner.Replacements = map[string]string{"source": "Clerical"}
	addTestPowerSourceSpell(e, "Fireball", "Arcane")
	clericalFireball := addTestPowerSourceSpell(e, "Fireball", "@holy@")
	clericalFireball.Replacements = map[string]string{"holy": "Clerical"}

	// The owner's power source resolves to Clerical, so only the Clerical Fireball is the same as it.
	p := NewSpellPrereq()
	p.QualifierCriteria.Qualifier = "Fireball"
	c.True(p.Satisfied(e, owner, nil, "", nil), "the Clerical Fireball matches the owner's resolved power source")
	p.QuantityCriteria.Qualifier = fxp.Two
	var tooltip xbytes.InsertBuffer
	c.False(p.Satisfied(e, owner, &tooltip, "", nil),
		"the Arcane Fireball must not match the owner's resolved power source")
	c.Equal(`Has at least 2 spells whose name is "Fireball" and whose power source is the same as this spell's `+
		`("Clerical")`, tooltip.String(), "the tooltip must name the owner's resolved power source")

	// A marker in the power source qualifier is resolved through the owner's replacements.
	p = NewSpellPrereq()
	p.QualifierCriteria.Qualifier = "Fireball"
	p.SamePowerSource = false
	p.PowerSourceCriteria.Compare = criteria.IsText
	p.PowerSourceCriteria.Qualifier = "@source@"
	c.True(p.Satisfied(e, owner, nil, "", nil), "the Clerical Fireball matches the resolved qualifier")
	p.QuantityCriteria.Qualifier = fxp.Two
	tooltip.Reset()
	c.False(p.Satisfied(e, owner, &tooltip, "", nil), "the Arcane Fireball must not match the resolved qualifier")
	c.Equal(`Has at least 2 spells whose name is "Fireball" and whose power source is "Clerical"`, tooltip.String(),
		"the tooltip must show the resolved qualifier")

	// Changing the owner's replacement changes which spell matches.
	owner.Replacements["source"] = "Arcane"
	p.QuantityCriteria.Qualifier = fxp.One
	c.True(p.Satisfied(e, owner, nil, "", nil), "the Arcane Fireball matches the re-resolved qualifier")
	p.PowerSourceCriteria.Qualifier = "Arcane"
	p.PowerSourceCriteria.Compare = criteria.IsNotText
	c.True(p.Satisfied(e, owner, nil, "", nil),
		"the Clerical Fireball's power source resolves through its own replacements")
	p.QuantityCriteria.Qualifier = fxp.Two
	c.False(p.Satisfied(e, owner, nil, "", nil),
		"the Clerical Fireball is the only spell whose power source is not Arcane")
}

// TestSpellPrereqPowerSourceDefaultsAndNameableKeys verifies the defaults for a newly created spell prerequisite and
// that a substitution marker in the power source qualifier is extracted regardless of the sub-type.
func TestSpellPrereqPowerSourceDefaultsAndNameableKeys(t *testing.T) {
	c := check.New(t)

	p := NewSpellPrereq()
	c.True(p.SamePowerSource, "a new spell prereq defaults to matching the owning spell's power source")
	c.Equal(criteria.AnyText, p.PowerSourceCriteria.Compare, "a new spell prereq has no explicit power source criteria")

	for _, subType := range []spellcmp.Type{spellcmp.Name, spellcmp.Any, spellcmp.CollegeCount} {
		p = NewSpellPrereq()
		p.SubType = subType
		p.SamePowerSource = false
		p.PowerSourceCriteria.Compare = criteria.IsText
		p.PowerSourceCriteria.Qualifier = "@source@"
		m := make(map[string]string)
		p.FillWithNameableKeys(m, nil)
		c.Equal(nameable.Unset, m["source"], "%v: the power source qualifier's marker must be extracted", subType)
	}
}

// TestSpellPrereqPowerSourceJSON verifies that the power source fields are omitted from saved files when unset, survive
// a round trip when set, and that a file written before this option existed loads with no power source filter.
func TestSpellPrereqPowerSourceJSON(t *testing.T) {
	c := check.New(t)

	// Nothing is written when no power source filter is present.
	p := NewSpellPrereq()
	p.SamePowerSource = false
	p.QualifierCriteria.Qualifier = "Fireball"
	data, err := jio.Marshal(Prereqs{p})
	c.NoError(err, "the prereq should marshal")
	c.NotContains(string(data), "same_power_source", "an unset same power source flag should not be written")
	c.NotContains(string(data), "power_source", "an unset power source criteria should not be written")

	// The "same as this spell's" form survives a round trip.
	p.SamePowerSource = true
	data, err = jio.Marshal(Prereqs{p})
	c.NoError(err, "the prereq should marshal")
	c.Contains(string(data), `"same_power_source":true`, "the same power source flag should be written")
	c.True(loadSpellPrereq(c, data).SamePowerSource, "the same power source flag should load")

	// So does an explicit criterion.
	p.SamePowerSource = false
	p.PowerSourceCriteria.Compare = criteria.IsText
	p.PowerSourceCriteria.Qualifier = "Arcane"
	data, err = jio.Marshal(Prereqs{p})
	c.NoError(err, "the prereq should marshal")
	c.Contains(string(data), `"power_source":{"compare":"is","qualifier":"Arcane"}`,
		"the power source criteria should be written")
	loaded := loadSpellPrereq(c, data)
	c.False(loaded.SamePowerSource, "the same power source flag should not have been set")
	c.Equal(criteria.IsText, loaded.PowerSourceCriteria.Compare, "the power source comparison should load")
	c.Equal("Arcane", loaded.PowerSourceCriteria.Qualifier, "the power source qualifier should load")

	// A file written before this option existed carries neither key and loads with no power source filter.
	loaded = loadSpellPrereq(c, []byte(`[{"type":"spell_prereq","sub_type":"name","has":true,`+
		`"qualifier":{"compare":"is","qualifier":"Fireball"},"quantity":{"compare":"at_least","qualifier":1}}]`))
	c.False(loaded.SamePowerSource, "an old file must not gain the same power source filter")
	c.Equal(criteria.AnyText, loaded.PowerSourceCriteria.Compare, "an old file must not gain a power source criteria")
	c.False(loaded.hasPowerSourceFilter(), "an old file's prereq has no power source filter at all")
}

// loadSpellPrereq unmarshals a one-element prereq list and returns the spell prereq it holds.
func loadSpellPrereq(c check.Checker, data []byte) *SpellPrereq {
	var prereqs Prereqs
	c.NoError(jio.Unmarshal(data, &prereqs), "the prereqs should unmarshal")
	c.Equal(1, len(prereqs), "exactly one prereq should be present")
	p, ok := prereqs[0].(*SpellPrereq)
	c.True(ok, "the prereq should be a spell prereq")
	return p
}

// TestSpellPrereqPowerSourceHash verifies that the power source fields participate in the hash, so that a change to
// them is seen as a change to the prerequisite.
func TestSpellPrereqPowerSourceHash(t *testing.T) {
	c := check.New(t)

	none := NewSpellPrereq()
	none.SamePowerSource = false
	same := NewSpellPrereq()
	explicit := NewSpellPrereq()
	explicit.SamePowerSource = false
	explicit.PowerSourceCriteria.Compare = criteria.IsText
	explicit.PowerSourceCriteria.Qualifier = "Arcane"

	c.NotEqual(Hash64(none), Hash64(same), "the same power source flag must affect the hash")
	c.NotEqual(Hash64(none), Hash64(explicit), "the power source criteria must affect the hash")
	c.NotEqual(Hash64(same), Hash64(explicit), "the two forms of the filter must hash differently")

	other := NewSpellPrereq()
	other.SamePowerSource = false
	other.PowerSourceCriteria.Compare = criteria.IsText
	other.PowerSourceCriteria.Qualifier = "Arcane"
	c.Equal(Hash64(explicit), Hash64(other), "two identical prereqs must hash the same")
}
