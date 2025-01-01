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
	"fmt"
	"hash"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var _ Prereq = &EquippedEquipmentPrereq{}

// EquippedEquipmentPrereq holds a prerequisite for an equipped piece of equipment.
type EquippedEquipmentPrereq struct {
	Parent       *PrereqList   `json:"-"`
	Type         prereq.Type   `json:"type"`
	NameCriteria criteria.Text `json:"name,omitempty"`
	TagsCriteria criteria.Text `json:"tags,omitempty"`
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
	nameable.Extract(p.NameCriteria.Qualifier, m, existing)
	nameable.Extract(p.TagsCriteria.Qualifier, m, existing)
}

// Satisfied implements Prereq.
func (p *EquippedEquipmentPrereq) Satisfied(entity *Entity, exclude any, tooltip *xio.ByteBuffer, prefix string, hasEquipmentPenalty *bool) bool {
	var replacements map[string]string
	if na, ok := exclude.(nameable.Accesser); ok {
		replacements = na.NameableReplacements()
	}
	satisfied := false
	Traverse(func(eqp *Equipment) bool {
		satisfied = exclude != eqp && eqp.Equipped && eqp.Quantity > 0 &&
			p.NameCriteria.Matches(replacements, eqp.NameWithReplacements()) &&
			p.TagsCriteria.MatchesList(replacements, eqp.Tags...)
		return satisfied
	}, false, false, entity.CarriedEquipment...)
	if !satisfied {
		*hasEquipmentPenalty = true
		if tooltip != nil {
			fmt.Fprintf(tooltip, i18n.Text("%sHas equipment which is equipped and whose name %s %s"),
				prefix, p.NameCriteria.String(replacements),
				p.TagsCriteria.StringWithPrefix(replacements, i18n.Text("and at least one tag"), i18n.Text("and all tags")))
		}
	}
	return satisfied
}

// Hash writes this object's contents into the hasher.
func (p *EquippedEquipmentPrereq) Hash(h hash.Hash) {
	if p == nil {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.Num8(h, p.Type)
	p.NameCriteria.Hash(h)
	p.TagsCriteria.Hash(h)
}
