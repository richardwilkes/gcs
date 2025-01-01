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

var _ Prereq = &AttributePrereq{}

// AttributePrereq holds a prerequisite for an attribute.
type AttributePrereq struct {
	Parent            *PrereqList     `json:"-"`
	Type              prereq.Type     `json:"type"`
	Has               bool            `json:"has"`
	CombinedWith      string          `json:"combined_with,omitempty"`
	QualifierCriteria criteria.Number `json:"qualifier,omitempty"`
	Which             string          `json:"which"`
}

// NewAttributePrereq creates a new AttributePrereq. 'entity' may be nil.
func NewAttributePrereq(entity *Entity) *AttributePrereq {
	var p AttributePrereq
	p.Type = prereq.Attribute
	p.QualifierCriteria.Compare = criteria.AtLeastNumber
	p.QualifierCriteria.Qualifier = fxp.Ten
	p.Which = AttributeIDFor(entity, StrengthID)
	p.Has = true
	return &p
}

// PrereqType implements Prereq.
func (p *AttributePrereq) PrereqType() prereq.Type {
	return p.Type
}

// ParentList implements Prereq.
func (p *AttributePrereq) ParentList() *PrereqList {
	return p.Parent
}

// Clone implements Prereq.
func (p *AttributePrereq) Clone(parent *PrereqList) Prereq {
	clone := *p
	clone.Parent = parent
	return &clone
}

// FillWithNameableKeys implements Prereq.
func (p *AttributePrereq) FillWithNameableKeys(_, _ map[string]string) {
}

// Satisfied implements Prereq.
func (p *AttributePrereq) Satisfied(entity *Entity, _ any, tooltip *xio.ByteBuffer, prefix string, _ *bool) bool {
	value := entity.ResolveAttributeCurrent(p.Which)
	if p.CombinedWith != "" {
		value += entity.ResolveAttributeCurrent(p.CombinedWith)
	}
	satisfied := p.QualifierCriteria.Matches(value)
	if !p.Has {
		satisfied = !satisfied
	}
	if !satisfied && tooltip != nil {
		tooltip.WriteString(prefix)
		tooltip.WriteString(HasText(p.Has))
		tooltip.WriteByte(' ')
		tooltip.WriteString(entity.ResolveAttributeName(p.Which))
		if p.CombinedWith != "" {
			tooltip.WriteByte('+')
			tooltip.WriteString(entity.ResolveAttributeName(p.CombinedWith))
		}
		tooltip.WriteString(i18n.Text(" which "))
		tooltip.WriteString(p.QualifierCriteria.String())
	}
	return satisfied
}

// Hash writes this object's contents into the hasher.
func (p *AttributePrereq) Hash(h hash.Hash) {
	if p == nil {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.Num8(h, p.Type)
	hashhelper.Bool(h, p.Has)
	hashhelper.String(h, p.CombinedWith)
	p.QualifierCriteria.Hash(h)
	hashhelper.String(h, p.Which)
}
