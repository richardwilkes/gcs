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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Prereq = &EquippedEquipmentPrereq{}

// EquippedEquipmentPrereq holds a prerequisite for a piece of carried equipment that is equipped or, when its type is
// prereq.CarriedEquipment, carried but not equipped. Not requires that there be no such equipment instead.
type EquippedEquipmentPrereq struct {
	Parent       *PrereqList   `json:"-"`
	Type         prereq.Type   `json:"type"`
	Not          bool          `json:"not,omitzero"`
	NameCriteria criteria.Text `json:"name,omitzero"`
	TagsCriteria criteria.Text `json:"tags,omitzero"`
	ModifierCriteria
	TitleNoteCriteria criteria.Text `json:"title_note,omitzero"`
}

// NewEquippedEquipmentPrereq creates a new EquippedEquipmentPrereq.
func NewEquippedEquipmentPrereq() *EquippedEquipmentPrereq {
	var p EquippedEquipmentPrereq
	p.Type = prereq.EquippedEquipment
	p.NameCriteria.Compare = criteria.IsText
	p.TagsCriteria.Compare = criteria.AnyText
	p.TitleNoteCriteria.Compare = criteria.AnyText
	return &p
}

// NewCarriedEquipmentPrereq creates a new EquippedEquipmentPrereq for equipment that is carried but not equipped.
func NewCarriedEquipmentPrereq() *EquippedEquipmentPrereq {
	p := NewEquippedEquipmentPrereq()
	p.Type = prereq.CarriedEquipment
	return p
}

// PrereqType implements Prereq.
func (p *EquippedEquipmentPrereq) PrereqType() prereq.Type {
	return p.Type
}

// ParentList implements Prereq.
func (p *EquippedEquipmentPrereq) ParentList() *PrereqList {
	return p.Parent
}

// SetParentList implements Prereq.
func (p *EquippedEquipmentPrereq) SetParentList(list *PrereqList) {
	p.Parent = list
}

// Clone implements Prereq.
func (p *EquippedEquipmentPrereq) Clone(parent *PrereqList) Prereq {
	clone := *p
	clone.Parent = parent
	return &clone
}

// FillWithNameableKeys implements Prereq.
func (p *EquippedEquipmentPrereq) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(
		m, existing,
		p.NameCriteria.Qualifier,
		p.TagsCriteria.Qualifier,
		p.TitleNoteCriteria.Qualifier,
	)
	p.fillWithModifierNameableKeys(m, existing)
}

// Satisfied implements Prereq. Only an unmet prerequisite to have equipment equipped sets hasEquipmentPenalty.
func (p *EquippedEquipmentPrereq) Satisfied(entity *Entity, exclude any, tooltip *xbytes.InsertBuffer, prefix string, hasEquipmentPenalty *bool) bool {
	if entity == nil {
		return true
	}
	var replacements map[string]string
	if na, ok := exclude.(nameable.Accesser); ok {
		replacements = na.NameableReplacements()
	}
	carried := p.Type == prereq.CarriedEquipment
	satisfied := false
	Traverse(func(eqp *Equipment) bool {
		held := eqp.ReallyEquipped()
		if carried {
			held = !held && eqp.Enabled()
			for one := eqp; held && one != nil; one = one.parent {
				held = one.Quantity > 0
			}
		}
		satisfied = exclude != eqp && held && p.NameCriteria.Matches(replacements, eqp.NameWithReplacements()) &&
			p.TagsCriteria.MatchesList(replacements, eqp.Tags...) &&
			modifiersMeet(&p.ModifierCriteria, replacements, eqp.Modifiers) &&
			(p.TitleNoteCriteria.Compare == criteria.AnyText ||
				p.TitleNoteCriteria.MatchesEach(replacements, eqp.TitleNotes()...))
		return satisfied
	}, false, false, entity.CarriedEquipment...)
	if p.Not {
		satisfied = !satisfied
	}
	if !satisfied {
		if !p.Not && !carried && hasEquipmentPenalty != nil {
			*hasEquipmentPenalty = true
		}
		if tooltip != nil {
			tooltip.WriteString(prefix)
			tooltip.WriteString(p.Describe(entity, replacements, plainText))
		}
	}
	return satisfied
}

// Describe implements Prereq.
func (p *EquippedEquipmentPrereq) Describe(_ *Entity, replacements map[string]string, em func(string) string) string {
	has := HasText(!p.Not)
	carried := p.Type == prereq.CarriedEquipment
	var text string
	switch {
	case p.NameCriteria.Compare == criteria.IsText && strings.TrimSpace(nameable.Apply(p.NameCriteria.Qualifier, replacements)) != "":
		name := describeName(p.NameCriteria, replacements, em)
		if carried {
			text = i18n.Text("%s %s carried but not equipped", has, name)
		} else {
			text = i18n.Text("%s %s equipped", has, name)
		}
	case p.NameCriteria.Compare != criteria.AnyText:
		whose := describeText(p.NameCriteria, replacements, em)
		if carried {
			text = i18n.Text("%s unequipped carried equipment whose name %s", has, whose)
		} else {
			text = i18n.Text("%s equipped equipment whose name %s", has, whose)
		}
	case p.TagsCriteria.Compare != criteria.AnyText || p.TitleNoteCriteria.Compare != criteria.AnyText:
		if carried {
			text = i18n.Text("%s unequipped carried equipment", has)
		} else {
			text = i18n.Text("%s equipped equipment", has)
		}
	case carried:
		text = i18n.Text("%s any equipment carried but not equipped", has)
	default:
		text = i18n.Text("%s any equipment equipped", has)
	}
	return text + describeTags(p.TagsCriteria, replacements, em) + p.describeModifier(replacements, em) +
		describeTitleNote(p.TitleNoteCriteria, false, replacements, em)
}

// Hash writes this object's contents into the hasher.
func (p *EquippedEquipmentPrereq) Hash(h hash.Hash) {
	if p == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, p.Type)
	xhash.Bool(h, p.Not)
	p.NameCriteria.Hash(h)
	p.TagsCriteria.Hash(h)
	p.hashModifier(h)
	p.TitleNoteCriteria.Hash(h)
}
