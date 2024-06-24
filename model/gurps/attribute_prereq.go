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
	"hash"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

var _ Prereq = &AttributePrereq{}

// AttributePrereq holds a prerequisite for an attribute.
type AttributePrereq struct {
	Parent            *PrereqList     `json:"-"`
	Type              prereq.Type     `json:"type"`
	Has               bool            `json:"has"`
	CombinedWith      string          `json:"combined_with,omitempty"`
	QualifierCriteria NumericCriteria `json:"qualifier,omitempty"`
	Which             string          `json:"which"`
}

// NewAttributePrereq creates a new AttributePrereq. 'entity' may be nil.
func NewAttributePrereq(entity *Entity) *AttributePrereq {
	return &AttributePrereq{
		Type: prereq.Attribute,
		QualifierCriteria: NumericCriteria{
			NumericCriteriaData: NumericCriteriaData{
				Compare:   AtLeastNumber,
				Qualifier: fxp.Ten,
			},
		},
		Which: AttributeIDFor(entity, StrengthID),
		Has:   true,
	}
}

// PrereqType implements Prereq.
func (a *AttributePrereq) PrereqType() prereq.Type {
	return a.Type
}

// ParentList implements Prereq.
func (a *AttributePrereq) ParentList() *PrereqList {
	return a.Parent
}

// Clone implements Prereq.
func (a *AttributePrereq) Clone(parent *PrereqList) Prereq {
	clone := *a
	clone.Parent = parent
	return &clone
}

// FillWithNameableKeys implements Prereq.
func (a *AttributePrereq) FillWithNameableKeys(_ map[string]string) {
}

// ApplyNameableKeys implements Prereq.
func (a *AttributePrereq) ApplyNameableKeys(_ map[string]string) {
}

// Satisfied implements Prereq.
func (a *AttributePrereq) Satisfied(entity *Entity, _ any, tooltip *xio.ByteBuffer, prefix string, _ *bool) bool {
	value := entity.ResolveAttributeCurrent(a.Which)
	if a.CombinedWith != "" {
		value += entity.ResolveAttributeCurrent(a.CombinedWith)
	}
	satisfied := a.QualifierCriteria.Matches(value)
	if !a.Has {
		satisfied = !satisfied
	}
	if !satisfied && tooltip != nil {
		tooltip.WriteString(prefix)
		tooltip.WriteString(HasText(a.Has))
		tooltip.WriteByte(' ')
		tooltip.WriteString(entity.ResolveAttributeName(a.Which))
		if a.CombinedWith != "" {
			tooltip.WriteByte('+')
			tooltip.WriteString(entity.ResolveAttributeName(a.CombinedWith))
		}
		tooltip.WriteString(i18n.Text(" which "))
		tooltip.WriteString(a.QualifierCriteria.String())
	}
	return satisfied
}

// Hash writes this object's contents into the hasher.
func (a *AttributePrereq) Hash(h hash.Hash) {
	if a == nil {
		return
	}
	_ = binary.Write(h, binary.LittleEndian, a.Type)
	_ = binary.Write(h, binary.LittleEndian, a.Has)
	_, _ = h.Write([]byte(a.CombinedWith))
	a.QualifierCriteria.Hash(h)
	_, _ = h.Write([]byte(a.Which))
}
