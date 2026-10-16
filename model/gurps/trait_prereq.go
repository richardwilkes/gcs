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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Prereq = &TraitPrereq{}

// TraitPrereq holds a prereq against a Trait. Its modifier criteria are met by an enabled modifier that meets them all.
type TraitPrereq struct {
	Parent                *PrereqList     `json:"-"`
	Type                  prereq.Type     `json:"type"`
	Has                   bool            `json:"has"`
	NameCriteria          criteria.Text   `json:"name,omitzero"`
	LevelCriteria         criteria.Number `json:"level,omitzero"`
	PointsCriteria        criteria.Number `json:"points,omitzero"`
	TagsCriteria          criteria.Text   `json:"tags,omitzero"`
	NotesCriteria         criteria.Text   `json:"notes,omitzero"`
	ModifierNameCriteria  criteria.Text   `json:"modifier_name,omitzero"`
	ModifierNotesCriteria criteria.Text   `json:"modifier_notes,omitzero"`
	ModifierTagsCriteria  criteria.Text   `json:"modifier_tags,omitzero"`
}

// NewTraitPrereq creates a new TraitPrereq.
func NewTraitPrereq() *TraitPrereq {
	var p TraitPrereq
	p.Type = prereq.Trait
	p.NameCriteria.Compare = criteria.IsText
	p.LevelCriteria.Compare = criteria.AtLeastNumber
	p.NotesCriteria.Compare = criteria.AnyText
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

// SetParentList implements Prereq.
func (p *TraitPrereq) SetParentList(list *PrereqList) {
	p.Parent = list
}

// Clone implements Prereq.
func (p *TraitPrereq) Clone(parent *PrereqList) Prereq {
	clone := *p
	clone.Parent = parent
	return &clone
}

// FillWithNameableKeys implements Prereq.
func (p *TraitPrereq) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(
		m, existing,
		p.NameCriteria.Qualifier,
		p.TagsCriteria.Qualifier,
		p.NotesCriteria.Qualifier,
		p.ModifierNameCriteria.Qualifier,
		p.ModifierNotesCriteria.Qualifier,
		p.ModifierTagsCriteria.Qualifier,
	)
}

// Satisfied implements Prereq.
func (p *TraitPrereq) Satisfied(entity *Entity, exclude any, tooltip *xbytes.InsertBuffer, prefix string, _ *bool) bool {
	if entity == nil {
		return true
	}
	var replacements map[string]string
	if na, ok := exclude.(nameable.Accesser); ok {
		replacements = na.NameableReplacements()
	}
	satisfied := false
	Traverse(func(t *Trait) bool {
		if exclude == t || !p.NameCriteria.Matches(replacements, t.NameWithReplacements()) ||
			(p.PointsCriteria.Compare != criteria.AnyNumber && !p.PointsCriteria.Matches(t.AdjustedPoints(nil))) ||
			(p.TagsCriteria.Compare != criteria.AnyText && !p.TagsCriteria.MatchesList(replacements, t.Tags...)) ||
			!p.modifierMatches(replacements, t.Modifiers) {
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
			levels = t.CurrentLevel()
		}
		satisfied = p.LevelCriteria.Matches(levels)
		return satisfied
	}, true, false, entity.Traits...)
	if !p.Has {
		satisfied = !satisfied
	}
	if !satisfied && tooltip != nil {
		tooltip.WriteString(prefix)
		tooltip.WriteString(p.Describe(entity, replacements, plainText))
	}
	return satisfied
}

// Describe implements Prereq. A level of "at least 0" or less is left out, since every trait has at least that.
func (p *TraitPrereq) Describe(_ *Entity, replacements map[string]string, em func(string) string) string {
	text := HasText(p.Has) + i18n.Text(" trait ") + describeName(p.NameCriteria, replacements, em)
	if p.LevelCriteria.Compare != criteria.AnyNumber &&
		(p.LevelCriteria.Compare != criteria.AtLeastNumber || p.LevelCriteria.Qualifier > 0) {
		text += i18n.Text(" at level ") + p.LevelCriteria.AltString()
	}
	if p.PointsCriteria.Compare != criteria.AnyNumber {
		text += i18n.Text(" whose point total is ") + p.PointsCriteria.AltString()
	}
	text += describeTags(p.TagsCriteria, replacements, em)
	if p.NotesCriteria.Compare != criteria.AnyText {
		text += i18n.Text(" whose notes ") + describeText(p.NotesCriteria, replacements, em)
	}
	if p.hasModifierCriteria() {
		var parts []string
		for _, part := range []string{
			describeWhose(i18n.Text("whose name"), p.ModifierNameCriteria, replacements, em),
			describeWhose(i18n.Text("whose notes"), p.ModifierNotesCriteria, replacements, em),
			describeTags(p.ModifierTagsCriteria, replacements, em),
		} {
			if part != "" {
				parts = append(parts, strings.TrimSpace(part))
			}
		}
		text += i18n.Text(" with a modifier ") + strings.Join(parts, i18n.Text(" and "))
	}
	return text
}

func (p *TraitPrereq) hasModifierCriteria() bool {
	return p.ModifierNameCriteria.Compare != criteria.AnyText || p.ModifierNotesCriteria.Compare != criteria.AnyText ||
		p.ModifierTagsCriteria.Compare != criteria.AnyText
}

// modifierMatches returns true if there are no modifier criteria, or one of the enabled modifiers, containers left
// out, meets them all.
func (p *TraitPrereq) modifierMatches(replacements map[string]string, modifiers []*TraitModifier) bool {
	if !p.hasModifierCriteria() {
		return true
	}
	found := false
	Traverse(func(mod *TraitModifier) bool {
		found = p.ModifierNameCriteria.Matches(replacements, mod.NameWithReplacements()) &&
			p.ModifierNotesCriteria.Matches(replacements, mod.LocalNotesWithReplacements()) &&
			p.ModifierTagsCriteria.MatchesList(replacements, mod.Tags...)
		return found
	}, true, true, modifiers...)
	return found
}

// Hash writes this object's contents into the hasher.
func (p *TraitPrereq) Hash(h hash.Hash) {
	if p == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, p.Type)
	xhash.Bool(h, p.Has)
	p.NameCriteria.Hash(h)
	p.LevelCriteria.Hash(h)
	p.PointsCriteria.Hash(h)
	p.TagsCriteria.Hash(h)
	p.NotesCriteria.Hash(h)
	p.ModifierNameCriteria.Hash(h)
	p.ModifierNotesCriteria.Hash(h)
	p.ModifierTagsCriteria.Hash(h)
}
