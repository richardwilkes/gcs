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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellcmp"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Prereq = &SpellPrereq{}

// SpellPrereq holds a prerequisite for a spell.
type SpellPrereq struct {
	Parent              *PrereqList     `json:"-"`
	Type                prereq.Type     `json:"type"`
	SubType             spellcmp.Type   `json:"sub_type"`
	Has                 bool            `json:"has"`
	QualifierCriteria   criteria.Text   `json:"qualifier,omitzero"`
	QuantityCriteria    criteria.Number `json:"quantity,omitzero"`
	SamePowerSource     bool            `json:"same_power_source,omitzero"`
	PowerSourceCriteria criteria.Text   `json:"power_source,omitzero"`
}

// NewSpellPrereq creates a new SpellPrereq.
func NewSpellPrereq() *SpellPrereq {
	var p SpellPrereq
	p.Type = prereq.Spell
	p.SubType = spellcmp.Name
	p.QualifierCriteria.Compare = criteria.IsText
	p.QuantityCriteria.Compare = criteria.AtLeastNumber
	p.QuantityCriteria.Qualifier = fxp.One
	p.SamePowerSource = true
	p.PowerSourceCriteria.Compare = criteria.AnyText
	p.Has = true
	return &p
}

// PrereqType implements Prereq.
func (p *SpellPrereq) PrereqType() prereq.Type {
	return p.Type
}

// ParentList implements Prereq.
func (p *SpellPrereq) ParentList() *PrereqList {
	return p.Parent
}

// Clone implements Prereq.
func (p *SpellPrereq) Clone(parent *PrereqList) Prereq {
	clone := *p
	clone.Parent = parent
	return &clone
}

// FillWithNameableKeys implements Prereq.
func (p *SpellPrereq) FillWithNameableKeys(m, existing map[string]string) {
	if p.SubType.UsesStringCriteria() {
		nameable.Extract(m, existing, p.QualifierCriteria.Qualifier)
	}
	if !p.SamePowerSource {
		nameable.Extract(m, existing, p.PowerSourceCriteria.Qualifier)
	}
}

// powerSourceMatches returns true if a candidate spell's power source passes this prerequisite's filter. For the "same
// as this spell" form, ownerPowerSource is the owning spell's resolved power source, or nil when the prerequisite does
// not belong to a spell, in which case nothing matches.
func (p *SpellPrereq) powerSourceMatches(replacements map[string]string, candidate string,
	ownerPowerSource *string,
) bool {
	if p.SamePowerSource {
		return ownerPowerSource != nil && strings.EqualFold(candidate, *ownerPowerSource)
	}
	return p.PowerSourceCriteria.Matches(replacements, candidate)
}

// hasPowerSourceFilter returns true if this prerequisite restricts which power sources may satisfy it.
func (p *SpellPrereq) hasPowerSourceFilter() bool {
	return p.SamePowerSource || p.PowerSourceCriteria.Compare != criteria.AnyText
}

// powerSourceDescription returns the "is ..." text for the tooltip. For the "same as this spell" form, the owning
// spell's resolved power source is named as well, so that a spell whose power source is blank or spelled differently
// from the spells it expects to match can be diagnosed from the tooltip alone.
func (p *SpellPrereq) powerSourceDescription(replacements map[string]string, ownerPowerSource *string) string {
	if !p.SamePowerSource {
		return p.PowerSourceCriteria.String(replacements)
	}
	desc := i18n.Text("is the same as this spell's")
	if ownerPowerSource != nil {
		desc += ` ("` + *ownerPowerSource + `")`
	}
	return desc
}

// writePowerSourceTooltip appends the power source portion of the tooltip, introduced by leadIn, to text that already
// describes the spells being looked for. Nothing is written when no power source filter has been set.
func (p *SpellPrereq) writePowerSourceTooltip(tooltip *xbytes.InsertBuffer, leadIn string,
	replacements map[string]string, ownerPowerSource *string,
) {
	if p.hasPowerSourceFilter() {
		tooltip.WriteString(leadIn)
		tooltip.WriteString(p.powerSourceDescription(replacements, ownerPowerSource))
	}
}

// Satisfied implements Prereq.
func (p *SpellPrereq) Satisfied(entity *Entity, exclude any, tooltip *xbytes.InsertBuffer, prefix string, _ *bool) bool {
	if entity == nil {
		return true
	}
	var replacements map[string]string
	if na, ok := exclude.(nameable.Accesser); ok {
		replacements = na.NameableReplacements()
	}
	var techLevel *string
	var ownerName string
	var ownerPowerSource *string
	excludeSpell, isSpell := exclude.(*Spell)
	if isSpell {
		techLevel = excludeSpell.TechLevel
		// Resolve the owning spell's name and power source once here rather than once per candidate spell, since a
		// nameable marker in either would otherwise be re-applied for every spell on the sheet.
		ownerName = excludeSpell.NameWithReplacements()
		resolved := excludeSpell.PowerSourceWithReplacements()
		ownerPowerSource = &resolved
	}
	filterPowerSource := p.hasPowerSourceFilter()
	count := 0
	colleges := make(map[string]bool)
	Traverse(func(sp *Spell) bool {
		if exclude == sp || sp.AdjustedPoints(nil) == 0 {
			return false
		}
		// Don't count a spell that, in turn, directly requires the spell being checked, since doing so would create a
		// circular prerequisite relationship (see GitHub issue #737).
		if excludeSpell != nil && spellDirectlyRequiresNamed(sp, ownerName, *ownerPowerSource) {
			return false
		}
		if techLevel != nil && sp.TechLevel != nil && *techLevel != *sp.TechLevel {
			return false
		}
		if filterPowerSource && !p.powerSourceMatches(replacements, sp.PowerSourceWithReplacements(), ownerPowerSource) {
			return false
		}
		switch p.SubType {
		case spellcmp.Name:
			if p.QualifierCriteria.Matches(replacements, sp.NameWithReplacements()) {
				count++
			}
		case spellcmp.Tag:
			for _, one := range sp.Tags {
				if p.QualifierCriteria.Matches(replacements, one) {
					count++
					break
				}
			}
		case spellcmp.College:
			for _, one := range sp.CollegeWithReplacements() {
				if p.QualifierCriteria.Matches(replacements, one) {
					count++
					break
				}
			}
		case spellcmp.CollegeCount:
			for _, one := range sp.CollegeWithReplacements() {
				colleges[one] = true
			}
		case spellcmp.Any:
			count++
		}
		return false
	}, false, true, entity.Spells...)
	if p.SubType == spellcmp.CollegeCount {
		count = len(colleges)
	}
	satisfied := p.QuantityCriteria.Matches(fxp.FromInteger(count))
	if !p.Has {
		satisfied = !satisfied
	}
	if !satisfied && tooltip != nil {
		tooltip.WriteString(prefix)
		tooltip.WriteString(HasText(p.Has))
		tooltip.WriteByte(' ')
		tooltip.WriteString(p.QuantityCriteria.AltString())
		if p.QuantityCriteria.Qualifier == fxp.One {
			tooltip.WriteString(i18n.Text(" spell "))
		} else {
			tooltip.WriteString(i18n.Text(" spells "))
		}
		switch p.SubType {
		case spellcmp.Any:
			if filterPowerSource {
				p.writePowerSourceTooltip(tooltip, i18n.Text("whose power source "), replacements, ownerPowerSource)
			} else {
				tooltip.WriteString(i18n.Text("of any kind"))
			}
		case spellcmp.CollegeCount:
			tooltip.WriteString(i18n.Text("from different colleges"))
			p.writePowerSourceTooltip(tooltip, i18n.Text(" whose power source "), replacements, ownerPowerSource)
		default:
			switch p.SubType {
			case spellcmp.Name:
				tooltip.WriteString(i18n.Text("whose name "))
			case spellcmp.Tag:
				tooltip.WriteString(i18n.Text("whose tag "))
			case spellcmp.College:
				tooltip.WriteString(i18n.Text("whose college "))
			}
			tooltip.WriteString(p.QualifierCriteria.String(replacements))
			p.writePowerSourceTooltip(tooltip, i18n.Text(" and whose power source "), replacements, ownerPowerSource)
		}
	}
	return satisfied
}

// spellDirectlyRequires returns true if the candidate spell directly lists a spell-by-name prerequisite that matches
// the target spell. This is used to prevent a spell from being counted as a prerequisite for a spell that it, in turn,
// requires, which would otherwise create a circular prerequisite relationship.
func spellDirectlyRequires(candidate, target *Spell) bool {
	if candidate == nil || target == nil {
		return false
	}
	return spellDirectlyRequiresNamed(candidate, target.NameWithReplacements(), target.PowerSourceWithReplacements())
}

// spellDirectlyRequiresNamed is spellDirectlyRequires with the target spell's name and power source already resolved,
// for callers that check many candidates against the same target.
func spellDirectlyRequiresNamed(candidate *Spell, targetName, targetPowerSource string) bool {
	if candidate == nil || candidate.Prereq == nil {
		return false
	}
	candidatePowerSource := candidate.PowerSourceWithReplacements()
	return prereqListRequiresSpell(candidate.Prereq, candidate.NameableReplacements(), &candidatePowerSource, targetName,
		targetPowerSource)
}

// prereqListRequiresSpell returns true if the given prereq list contains a spell-by-name prerequisite (at any nesting
// depth) that is required (Has), matches the target spell's name, and whose power source filter, if any, the target
// spell's power source passes. The replacements and owner power source are those of the spell the prereq list belongs
// to: its qualifiers are resolved against the former, and a "same power source as this spell" filter compares against
// the latter.
func prereqListRequiresSpell(list *PrereqList, replacements map[string]string, ownerPowerSource *string, targetName,
	targetPowerSource string,
) bool {
	if list == nil {
		return false
	}
	for _, one := range list.Prereqs {
		switch p := one.(type) {
		case *PrereqList:
			if prereqListRequiresSpell(p, replacements, ownerPowerSource, targetName, targetPowerSource) {
				return true
			}
		case *SpellPrereq:
			if p.Has && p.SubType == spellcmp.Name && p.QualifierCriteria.Matches(replacements, targetName) &&
				p.powerSourceMatches(replacements, targetPowerSource, ownerPowerSource) {
				return true
			}
		}
	}
	return false
}

// Hash writes this object's contents into the hasher.
func (p *SpellPrereq) Hash(h hash.Hash) {
	if p == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, p.Type)
	xhash.Num8(h, p.SubType)
	xhash.Bool(h, p.Has)
	p.QualifierCriteria.Hash(h)
	p.QuantityCriteria.Hash(h)
	xhash.Bool(h, p.SamePowerSource)
	if !p.SamePowerSource {
		p.PowerSourceCriteria.Hash(h)
	}
}
