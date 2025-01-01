// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellcmp"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var _ Prereq = &SpellPrereq{}

// SpellPrereq holds a prerequisite for a spell.
type SpellPrereq struct {
	Parent            *PrereqList     `json:"-"`
	Type              prereq.Type     `json:"type"`
	SubType           spellcmp.Type   `json:"sub_type"`
	Has               bool            `json:"has"`
	QualifierCriteria criteria.Text   `json:"qualifier,omitempty"`
	QuantityCriteria  criteria.Number `json:"quantity,omitempty"`
}

// NewSpellPrereq creates a new SpellPrereq.
func NewSpellPrereq() *SpellPrereq {
	var p SpellPrereq
	p.Type = prereq.Spell
	p.SubType = spellcmp.Name
	p.QualifierCriteria.Compare = criteria.IsText
	p.QuantityCriteria.Compare = criteria.AtLeastNumber
	p.QuantityCriteria.Qualifier = fxp.One
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
		nameable.Extract(p.QualifierCriteria.Qualifier, m, existing)
	}
}

// Satisfied implements Prereq.
func (p *SpellPrereq) Satisfied(entity *Entity, exclude any, tooltip *xio.ByteBuffer, prefix string, _ *bool) bool {
	var replacements map[string]string
	if na, ok := exclude.(nameable.Accesser); ok {
		replacements = na.NameableReplacements()
	}
	var techLevel *string
	if sp, ok := exclude.(*Spell); ok {
		techLevel = sp.TechLevel
	}
	count := 0
	colleges := make(map[string]bool)
	Traverse(func(sp *Spell) bool {
		if exclude == sp || sp.AdjustedPoints(nil) == 0 {
			return false
		}
		if techLevel != nil && sp.TechLevel != nil && *techLevel != *sp.TechLevel {
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
	satisfied := p.QuantityCriteria.Matches(fxp.From(count))
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
			tooltip.WriteString(i18n.Text("of any kind"))
		case spellcmp.CollegeCount:
			tooltip.WriteString(i18n.Text("from different colleges"))
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
		}
	}
	return satisfied
}

// Hash writes this object's contents into the hasher.
func (p *SpellPrereq) Hash(h hash.Hash) {
	if p == nil {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.Num8(h, p.Type)
	hashhelper.Num8(h, p.SubType)
	hashhelper.Bool(h, p.Has)
	p.QualifierCriteria.Hash(h)
	p.QuantityCriteria.Hash(h)
}
