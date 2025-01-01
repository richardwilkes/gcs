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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var _ Prereq = &PrereqList{}

// PrereqList holds a prereq that contains a list of prerequisites.
type PrereqList struct {
	Parent  *PrereqList     `json:"-"`
	Type    prereq.Type     `json:"type"`
	All     bool            `json:"all"`
	WhenTL  criteria.Number `json:"when_tl,omitempty"`
	Prereqs Prereqs         `json:"prereqs,omitempty"`
}

// NewPrereqList creates a new PrereqList.
func NewPrereqList() *PrereqList {
	return &PrereqList{
		Type: prereq.List,
		All:  true,
	}
}

// ShouldOmit implements json.Omitter.
func (p *PrereqList) ShouldOmit() bool {
	return p == nil || len(p.Prereqs) == 0
}

// PrereqType implements Prereq.
func (p *PrereqList) PrereqType() prereq.Type {
	return p.Type
}

// ParentList implements Prereq.
func (p *PrereqList) ParentList() *PrereqList {
	if p == nil {
		return nil
	}
	return p.Parent
}

// Clone implements Prereq.
func (p *PrereqList) Clone(parent *PrereqList) Prereq {
	return p.CloneAsPrereqList(parent)
}

// CloneAsPrereqList clones this prereq list.
func (p *PrereqList) CloneAsPrereqList(parent *PrereqList) *PrereqList {
	clone := *p
	clone.Parent = parent
	clone.Prereqs = make(Prereqs, len(p.Prereqs))
	for i := range p.Prereqs {
		clone.Prereqs[i] = p.Prereqs[i].Clone(&clone)
	}
	return &clone
}

// CloneResolvingEmpty clones this prereq list. If the result would be nil and it isn't a container, a new, empty, list
// is created. If the result would not be nil but pruneIfEmpty is true and calling ShouldOmit() on it would return true,
// then nil is returned.
func (p *PrereqList) CloneResolvingEmpty(isContainer, pruneIfEmpty bool) *PrereqList {
	if p != nil {
		if pruneIfEmpty && p.ShouldOmit() {
			return nil
		}
		return p.CloneAsPrereqList(nil)
	}
	if isContainer {
		return nil
	}
	return NewPrereqList()
}

// FillWithNameableKeys implements Prereq.
func (p *PrereqList) FillWithNameableKeys(m, existing map[string]string) {
	for _, one := range p.Prereqs {
		one.FillWithNameableKeys(m, existing)
	}
}

// Satisfied implements Prereq.
func (p *PrereqList) Satisfied(entity *Entity, exclude any, buffer *xio.ByteBuffer, prefix string, hasEquipmentPenalty *bool) bool {
	if p.WhenTL.Compare != criteria.AnyNumber {
		tl, _, _ := ExtractTechLevel(entity.Profile.TechLevel)
		if tl < 0 {
			tl = 0
		}
		if !p.WhenTL.Compare.Matches(p.WhenTL.Qualifier, tl) {
			return true
		}
	}
	count := 0
	var local *xio.ByteBuffer
	if buffer != nil {
		local = &xio.ByteBuffer{}
	}
	eqpPenalty := false
	for _, one := range p.Prereqs {
		if one.Satisfied(entity, exclude, local, prefix, &eqpPenalty) {
			count++
		}
	}
	if local != nil && local.Len() != 0 {
		indented := strings.ReplaceAll(local.String(), "\n", "\n\u00a0\u00a0\u00a0")
		local = &xio.ByteBuffer{}
		local.WriteString(indented)
	}
	satisfied := count == len(p.Prereqs) || (!p.All && count > 0)
	if !satisfied {
		if eqpPenalty {
			*hasEquipmentPenalty = eqpPenalty
		}
		if buffer != nil && local != nil {
			buffer.WriteString(prefix)
			if p.All {
				buffer.WriteString(i18n.Text("Requires all of:"))
			} else {
				buffer.WriteString(i18n.Text("Requires at least one of:"))
			}
			buffer.WriteString(local.String())
		}
	}
	return satisfied
}

// Hash writes this object's contents into the hasher.
func (p *PrereqList) Hash(h hash.Hash) {
	if p.ShouldOmit() {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.Num8(h, p.Type)
	hashhelper.Bool(h, p.All)
	p.WhenTL.Hash(h)
	hashhelper.Num64(h, len(p.Prereqs))
	for _, one := range p.Prereqs {
		one.Hash(h)
	}
}
