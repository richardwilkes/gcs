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
	"fmt"
	"hash"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Prereq = &EquippedEquipmentPrereq{}

// EquippedEquipmentPrereq holds a prerequisite for an equipped piece of equipment.
type EquippedEquipmentPrereq struct {
	Parent       *PrereqList   `json:"-"`
	Type         prereq.Type   `json:"type"`
	NameCriteria criteria.Text `json:"name,omitzero"`
	TagsCriteria criteria.Text `json:"tags,omitzero"`
}

// NewEquippedEquipmentPrereq creates a new EquippedEquipmentPrereq.
func NewEquippedEquipmentPrereq() *EquippedEquipmentPrereq {
	var p EquippedEquipmentPrereq
	p.Type = prereq.EquippedEquipment
	p.NameCriteria.Compare = criteria.IsText
	p.TagsCriteria.Compare = criteria.AnyText
	return &p
}

// PrereqType implements Prereq.
func (p *EquippedEquipmentPrereq) PrereqType() prereq.Type {
	return p.Type
}

// ParentList implements Prereq.
func (p *EquippedEquipmentPrereq) ParentList() *PrereqList {
	return p.Parent
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
	)
}

// Satisfied implements Prereq.
func (p *EquippedEquipmentPrereq) Satisfied(entity *Entity, exclude any, tooltip *xbytes.InsertBuffer, prefix string, hasEquipmentPenalty *bool) bool {
	if entity == nil {
		return true
	}
	var replacements map[string]string
	if na, ok := exclude.(nameable.Accesser); ok {
		replacements = na.NameableReplacements()
	}
	satisfied := false
	Traverse(func(eqp *Equipment) bool {
		satisfied = exclude != eqp && eqp.ReallyEquipped() &&
			p.NameCriteria.Matches(replacements, eqp.NameWithReplacements()) &&
			p.TagsCriteria.MatchesList(replacements, eqp.Tags...)
		return satisfied
	}, false, false, entity.CarriedEquipment...)
	if !satisfied {
		if hasEquipmentPenalty != nil {
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
	tags := p.TagsCriteria.Compare != criteria.AnyText
	var text string
	switch {
	case p.NameCriteria.Compare == criteria.IsText && p.NameCriteria.Qualifier != "":
		text = fmt.Sprintf(i18n.Text("Has %s equipped"), em(nameable.Apply(p.NameCriteria.Qualifier, replacements)))
	case p.NameCriteria.Compare != criteria.AnyText:
		text = i18n.Text("Has equipped equipment whose name ") + describeText(p.NameCriteria, replacements, em)
	case tags:
		text = i18n.Text("Has equipped equipment")
	default:
		return i18n.Text("Has any equipment equipped")
	}
	if tags {
		q := nameable.Apply(p.TagsCriteria.Qualifier, replacements)
		if q != "" {
			q = em(q)
		}
		if p.TagsCriteria.Compare == criteria.IsText && q != "" {
			text += i18n.Text(" tagged ") + q
		} else {
			text += " " + p.TagsCriteria.Compare.DescribeWithPrefix(i18n.Text("with a tag that"),
				i18n.Text("with all tags that"), q)
		}
	}
	return text
}

// Hash writes this object's contents into the hasher.
func (p *EquippedEquipmentPrereq) Hash(h hash.Hash) {
	if p == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, p.Type)
	p.NameCriteria.Hash(h)
	p.TagsCriteria.Hash(h)
}
