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
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var _ Prereq = &ContainedWeightPrereq{}

// ContainedWeightPrereq holds a prerequisite for an equipment contained weight.
type ContainedWeightPrereq struct {
	Parent         *PrereqList     `json:"-"`
	Type           prereq.Type     `json:"type"`
	Has            bool            `json:"has"`
	WeightCriteria criteria.Weight `json:"qualifier,omitempty"`
}

// NewContainedWeightPrereq creates a new ContainedWeightPrereq.
func NewContainedWeightPrereq(entity *Entity) *ContainedWeightPrereq {
	var p ContainedWeightPrereq
	p.Type = prereq.ContainedWeight
	p.WeightCriteria.Compare = criteria.AtMostNumber
	p.WeightCriteria.Qualifier = fxp.WeightFromInteger(5, SheetSettingsFor(entity).DefaultWeightUnits)
	p.Has = true
	return &p
}

// PrereqType implements Prereq.
func (p *ContainedWeightPrereq) PrereqType() prereq.Type {
	return p.Type
}

// ParentList implements Prereq.
func (p *ContainedWeightPrereq) ParentList() *PrereqList {
	return p.Parent
}

// Clone implements Prereq.
func (p *ContainedWeightPrereq) Clone(parent *PrereqList) Prereq {
	clone := *p
	clone.Parent = parent
	return &clone
}

// FillWithNameableKeys implements Prereq.
func (p *ContainedWeightPrereq) FillWithNameableKeys(_, _ map[string]string) {
}

// Satisfied implements Prereq.
func (p *ContainedWeightPrereq) Satisfied(entity *Entity, exclude any, tooltip *xio.ByteBuffer, prefix string, _ *bool) bool {
	satisfied := false
	if eqp, ok := exclude.(*Equipment); ok {
		if satisfied = !eqp.Container(); !satisfied {
			units := SheetSettingsFor(entity).DefaultWeightUnits
			weight := eqp.ExtendedWeight(false, units) - eqp.AdjustedWeight(false, units)
			satisfied = p.WeightCriteria.Matches(weight)
		}
	}
	if !p.Has {
		satisfied = !satisfied
	}
	if !satisfied && tooltip != nil {
		tooltip.WriteString(prefix)
		tooltip.WriteString(HasText(p.Has))
		tooltip.WriteString(i18n.Text(" a contained weight which "))
		tooltip.WriteString(p.WeightCriteria.String())
	}
	return satisfied
}

// Hash writes this object's contents into the hasher.
func (p *ContainedWeightPrereq) Hash(h hash.Hash) {
	if p == nil {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.Num8(h, p.Type)
	hashhelper.Bool(h, p.Has)
	p.WeightCriteria.Hash(h)
}
