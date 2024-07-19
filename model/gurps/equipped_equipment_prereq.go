// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"encoding/binary"
	"fmt"
	"hash"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

var _ Prereq = &EquippedEquipmentPrereq{}

// EquippedEquipmentPrereq holds a prerequisite for an equipped piece of equipment.
type EquippedEquipmentPrereq struct {
	Parent       *PrereqList    `json:"-"`
	Type         prereq.Type    `json:"type"`
	NameCriteria StringCriteria `json:"name,omitempty"`
	TagsCriteria StringCriteria `json:"tags,omitempty"`
}

// NewEquippedEquipmentPrereq creates a new EquippedEquipmentPrereq.
func NewEquippedEquipmentPrereq() *EquippedEquipmentPrereq {
	return &EquippedEquipmentPrereq{
		Type: prereq.EquippedEquipment,
		NameCriteria: StringCriteria{
			StringCriteriaData: StringCriteriaData{
				Compare: IsString,
			},
		},
		TagsCriteria: StringCriteria{
			StringCriteriaData: StringCriteriaData{
				Compare: AnyString,
			},
		},
	}
}

// PrereqType implements Prereq.
func (e *EquippedEquipmentPrereq) PrereqType() prereq.Type {
	return e.Type
}

// ParentList implements Prereq.
func (e *EquippedEquipmentPrereq) ParentList() *PrereqList {
	return e.Parent
}

// Clone implements Prereq.
func (e *EquippedEquipmentPrereq) Clone(parent *PrereqList) Prereq {
	clone := *e
	clone.Parent = parent
	return &clone
}

// FillWithNameableKeys implements Prereq.
func (e *EquippedEquipmentPrereq) FillWithNameableKeys(m, existing map[string]string) {
	ExtractNameables(e.NameCriteria.Qualifier, m, existing)
	ExtractNameables(e.TagsCriteria.Qualifier, m, existing)
}

// Satisfied implements Prereq.
func (e *EquippedEquipmentPrereq) Satisfied(entity *Entity, exclude any, tooltip *xio.ByteBuffer, prefix string, hasEquipmentPenalty *bool) bool {
	var replacements map[string]string
	if na, ok := exclude.(NameableAccesser); ok {
		replacements = na.NameableReplacements()
	}
	satisfied := false
	Traverse(func(eqp *Equipment) bool {
		satisfied = exclude != eqp && eqp.Equipped && eqp.Quantity > 0 &&
			e.NameCriteria.Matches(replacements, eqp.NameWithReplacements()) &&
			e.TagsCriteria.MatchesList(replacements, eqp.Tags...)
		return satisfied
	}, false, false, entity.CarriedEquipment...)
	if !satisfied {
		*hasEquipmentPenalty = true
		if tooltip != nil {
			fmt.Fprintf(tooltip, i18n.Text("%sHas equipment which is equipped and whose name %s %s"),
				prefix, e.NameCriteria.String(replacements),
				e.TagsCriteria.StringWithPrefix(replacements, i18n.Text("and at least one tag"), i18n.Text("and all tags")))
		}
	}
	return satisfied
}

// Hash writes this object's contents into the hasher.
func (e *EquippedEquipmentPrereq) Hash(h hash.Hash) {
	if e == nil {
		return
	}
	_ = binary.Write(h, binary.LittleEndian, e.Type)
	e.NameCriteria.Hash(h)
	e.TagsCriteria.Hash(h)
}
