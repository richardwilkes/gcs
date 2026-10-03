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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Prereq = &ContainedQuantityPrereq{}

// ContainedQuantityPrereq holds a prerequisite for an equipment contained quantity.
type ContainedQuantityPrereq struct {
	Parent            *PrereqList     `json:"-"`
	Type              prereq.Type     `json:"type"`
	Has               bool            `json:"has"`
	QualifierCriteria criteria.Number `json:"qualifier,omitzero"`
}

// NewContainedQuantityPrereq creates a new ContainedQuantityPrereq.
func NewContainedQuantityPrereq() *ContainedQuantityPrereq {
	var p ContainedQuantityPrereq
	p.Type = prereq.ContainedQuantity
	p.QualifierCriteria.Compare = criteria.AtMostNumber
	p.QualifierCriteria.Qualifier = fxp.One
	p.Has = true
	return &p
}

// PrereqType implements Prereq.
func (p *ContainedQuantityPrereq) PrereqType() prereq.Type {
	return p.Type
}

// ParentList implements Prereq.
func (p *ContainedQuantityPrereq) ParentList() *PrereqList {
	return p.Parent
}

// SetParentList implements Prereq.
func (p *ContainedQuantityPrereq) SetParentList(list *PrereqList) {
	p.Parent = list
}

// Clone implements Prereq.
func (p *ContainedQuantityPrereq) Clone(parent *PrereqList) Prereq {
	clone := *p
	clone.Parent = parent
	return &clone
}

// FillWithNameableKeys implements Prereq.
func (p *ContainedQuantityPrereq) FillWithNameableKeys(_, _ map[string]string) {
}

// Satisfied implements Prereq.
func (p *ContainedQuantityPrereq) Satisfied(_ *Entity, exclude any, tooltip *xbytes.InsertBuffer, prefix string, _ *bool) bool {
	satisfied := false
	if eqp, ok := exclude.(*Equipment); ok {
		if satisfied = !eqp.Container(); !satisfied {
			satisfied = p.QualifierCriteria.Matches(containedQuantity(eqp.Children))
		}
	}
	if !p.Has {
		satisfied = !satisfied
	}
	if !satisfied && tooltip != nil {
		tooltip.WriteString(prefix)
		tooltip.WriteString(p.Describe(nil, nil, plainText))
	}
	return satisfied
}

// Describe implements Prereq.
func (p *ContainedQuantityPrereq) Describe(_ *Entity, _ map[string]string, _ func(string) string) string {
	text := HasText(p.Has) + i18n.Text(" a contained quantity")
	if p.QualifierCriteria.Compare != criteria.AnyNumber {
		text += i18n.Text(" of ") + p.QualifierCriteria.AltString()
	}
	return text
}

// Hash writes this object's contents into the hasher.
func (p *ContainedQuantityPrereq) Hash(h hash.Hash) {
	if p == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, p.Type)
	xhash.Bool(h, p.Has)
	p.QualifierCriteria.Hash(h)
}

// containedQuantity returns how many pieces of equipment the children amount to. A group only organizes what it holds,
// so what it holds is counted in its place, the same way the weight of what it holds counts toward the weight a
// container holds.
func containedQuantity(children []*Equipment) fxp.Int {
	var qty fxp.Int
	for _, child := range children {
		if child.IsGroup() {
			qty += containedQuantity(child.Children)
		} else {
			qty += child.Quantity
		}
	}
	return qty
}
