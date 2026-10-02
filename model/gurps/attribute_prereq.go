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
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Prereq = &AttributePrereq{}

// AttributePrereq holds a prerequisite for an attribute.
type AttributePrereq struct {
	Parent            *PrereqList     `json:"-"`
	Type              prereq.Type     `json:"type"`
	Has               bool            `json:"has"`
	CombinedWith      string          `json:"combined_with,omitzero"`
	QualifierCriteria criteria.Number `json:"qualifier,omitzero"`
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
func (p *AttributePrereq) Satisfied(entity *Entity, _ any, tooltip *xbytes.InsertBuffer, prefix string, _ *bool) bool {
	if entity == nil {
		return true
	}
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
		tooltip.WriteString(p.Describe(entity, nil, plainText))
	}
	return satisfied
}

// Describe implements Prereq. Attributes are named as the entity, which may be nil, defines them.
func (p *AttributePrereq) Describe(entity *Entity, _ map[string]string, em func(string) string) string {
	names := attributeTitle(entity, p.Which)
	if p.CombinedWith != "" {
		names += "+" + attributeTitle(entity, p.CombinedWith)
	}
	text := HasText(p.Has) + " " + em(names)
	if p.QualifierCriteria.Compare != criteria.AnyNumber {
		text += " " + p.QualifierCriteria.AltString()
	}
	return text
}

// attributeTitle returns the title the attribute choices give key, or key itself when it isn't one of them.
func attributeTitle(entity *Entity, key string) string {
	// No choice has an empty key, so the unrecognized choice for it is always the last one, which is skipped.
	choices, _ := AttributeChoices(entity, "", SizeFlag|DodgeFlag|ParryFlag|BlockFlag, "")
	for _, choice := range choices[:len(choices)-1] {
		if choice.Key == key {
			return choice.Title
		}
	}
	return key
}

// Hash writes this object's contents into the hasher.
func (p *AttributePrereq) Hash(h hash.Hash) {
	if p == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, p.Type)
	xhash.Bool(h, p.Has)
	xhash.StringWithLen(h, p.CombinedWith)
	p.QualifierCriteria.Hash(h)
	xhash.StringWithLen(h, p.Which)
}
