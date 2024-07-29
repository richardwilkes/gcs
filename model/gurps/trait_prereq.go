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

var _ Prereq = &TraitPrereq{}

// TraitPrereq holds a prereq against a Trait.
type TraitPrereq struct {
	Parent        *PrereqList     `json:"-"`
	Type          prereq.Type     `json:"type"`
	Has           bool            `json:"has"`
	NameCriteria  StringCriteria  `json:"name,omitempty"`
	LevelCriteria NumericCriteria `json:"level,omitempty"`
	NotesCriteria StringCriteria  `json:"notes,omitempty"`
}

// NewTraitPrereq creates a new TraitPrereq.
func NewTraitPrereq() *TraitPrereq {
	var p TraitPrereq
	p.Type = prereq.Trait
	p.NameCriteria.Compare = IsString
	p.LevelCriteria.Compare = AtLeastNumber
	p.NotesCriteria.Compare = AnyString
	p.Has = true
	return &p
}

// PrereqType implements Prereq.
func (p *TraitPrereq) PrereqType() prereq.Type {
	return p.Type
}

// ParentList implements Prereq.
func (p *TraitPrereq) ParentList() *PrereqList {
	return p.Parent
}

// Clone implements Prereq.
func (p *TraitPrereq) Clone(parent *PrereqList) Prereq {
	clone := *p
	clone.Parent = parent
	return &clone
}

// FillWithNameableKeys implements Prereq.
func (p *TraitPrereq) FillWithNameableKeys(m, existing map[string]string) {
	ExtractNameables(p.NameCriteria.Qualifier, m, existing)
	ExtractNameables(p.NotesCriteria.Qualifier, m, existing)
}

// Satisfied implements Prereq.
func (p *TraitPrereq) Satisfied(entity *Entity, exclude any, tooltip *xio.ByteBuffer, prefix string, _ *bool) bool {
	var replacements map[string]string
	if na, ok := exclude.(NameableAccesser); ok {
		replacements = na.NameableReplacements()
	}
	satisfied := false
	Traverse(func(t *Trait) bool {
		if exclude == t || !p.NameCriteria.Matches(replacements, t.NameWithReplacements()) {
			return false
		}
		notes := t.Notes()
		if modNotes := t.ModifierNotes(); modNotes != "" {
			notes += "\n" + modNotes
		}
		if !p.NotesCriteria.Matches(replacements, notes) {
			return false
		}
		var levels fxp.Int
		if t.IsLeveled() {
			levels = t.Levels.Max(0)
		}
		satisfied = p.LevelCriteria.Matches(levels)
		return satisfied
	}, true, false, entity.Traits...)
	if !p.Has {
		satisfied = !satisfied
	}
	if !satisfied && tooltip != nil {
		tooltip.WriteString(prefix)
		tooltip.WriteString(HasText(p.Has))
		tooltip.WriteString(i18n.Text(" a trait whose name "))
		tooltip.WriteString(p.NameCriteria.String(replacements))
		if p.NotesCriteria.Compare != AnyString {
			tooltip.WriteString(i18n.Text(", notes "))
			tooltip.WriteString(p.NotesCriteria.String(replacements))
			tooltip.WriteByte(',')
		}
		tooltip.WriteString(i18n.Text(" and level "))
		tooltip.WriteString(p.LevelCriteria.String())
	}
	return satisfied
}

// Hash writes this object's contents into the hasher.
func (p *TraitPrereq) Hash(h hash.Hash) {
	if p == nil {
		return
	}
	_ = binary.Write(h, binary.LittleEndian, p.Type)
	_ = binary.Write(h, binary.LittleEndian, p.Has)
	p.NameCriteria.Hash(h)
	p.LevelCriteria.Hash(h)
	p.NotesCriteria.Hash(h)
}
