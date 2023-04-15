/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps

import (
	"fmt"

	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

var _ Prereq = &EquippedEquipmentPrereq{}

// EquippedEquipmentPrereq holds a prerequisite for an equipped piece of equipment.
type EquippedEquipmentPrereq struct {
	Parent       *PrereqList    `json:"-"`
	Type         PrereqType     `json:"type"`
	NameCriteria StringCriteria `json:"name,omitempty"`
}

// NewEquippedEquipmentPrereq creates a new EquippedEquipmentPrereq.
func NewEquippedEquipmentPrereq() *EquippedEquipmentPrereq {
	return &EquippedEquipmentPrereq{
		Type: EquippedEquipmentPrereqType,
		NameCriteria: StringCriteria{
			StringCriteriaData: StringCriteriaData{
				Compare: IsString,
			},
		},
	}
}

// PrereqType implements Prereq.
func (e *EquippedEquipmentPrereq) PrereqType() PrereqType {
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
func (e *EquippedEquipmentPrereq) FillWithNameableKeys(m map[string]string) {
	Extract(e.NameCriteria.Qualifier, m)
}

// ApplyNameableKeys implements Prereq.
func (e *EquippedEquipmentPrereq) ApplyNameableKeys(m map[string]string) {
	e.NameCriteria.Qualifier = Apply(e.NameCriteria.Qualifier, m)
}

// Satisfied implements Prereq.
func (e *EquippedEquipmentPrereq) Satisfied(entity *Entity, exclude any, tooltip *xio.ByteBuffer, prefix string, hasEquipmentPenalty *bool) bool {
	satisfied := false
	Traverse(func(eqp *Equipment) bool {
		satisfied = exclude != eqp && eqp.Equipped && eqp.Quantity > 0 && e.NameCriteria.Matches(eqp.Name)
		return satisfied
	}, false, false, entity.CarriedEquipment...)
	if !satisfied {
		*hasEquipmentPenalty = true
		if tooltip != nil {
			fmt.Fprintf(tooltip, i18n.Text("%sHas equipment which is equipped and whose name %s"), prefix, e.NameCriteria.String())
		}
	}
	return satisfied
}
