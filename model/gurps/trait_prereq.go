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
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Prereq = &TraitPrereq{}

// TraitPrereq holds a prereq against a Trait.
type TraitPrereq struct {
	Parent         *PrereqList     `json:"-"`
	Type           prereq.Type     `json:"type"`
	Has            bool            `json:"has"`
	NameCriteria   criteria.Text   `json:"name,omitzero"`
	LevelCriteria  criteria.Number `json:"level,omitzero"`
	PointsCriteria criteria.Number `json:"points,omitzero"`
	TagsCriteria   criteria.Text   `json:"tags,omitzero"`
	NotesCriteria  criteria.Text   `json:"notes,omitzero"`
	ModifierCriteria
	TitleNoteCriteria criteria.Text `json:"title_note,omitzero"`
}

// NewTraitPrereq creates a new TraitPrereq.
func NewTraitPrereq() *TraitPrereq {
	var p TraitPrereq
	p.Type = prereq.Trait
	p.NameCriteria.Compare = criteria.IsText
	p.LevelCriteria.Compare = criteria.AtLeastNumber
	p.NotesCriteria.Compare = criteria.AnyText
	p.TitleNoteCriteria.Compare = criteria.AnyText
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
		p.TitleNoteCriteria.Qualifier,
	)
	p.fillWithModifierNameableKeys(m, existing)
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
		if excludedFromPrereqs(exclude, t) || !p.NameCriteria.Matches(replacements, t.NameWithReplacements()) ||
			// The points are those the trait would cost if bought, so that a granted trait, which costs nothing, is
			// judged as the same trait bought would be, as a granted skill is by the points its grant gives it.
			(p.PointsCriteria.Compare != criteria.AnyNumber && !p.PointsCriteria.Matches(t.ClassifyingPoints())) ||
			(p.TagsCriteria.Compare != criteria.AnyText && !p.TagsCriteria.MatchesList(replacements, t.Tags...)) ||
			(p.TitleNoteCriteria.Compare != criteria.AnyText &&
				!p.TitleNoteCriteria.MatchesEach(replacements, t.TitleNotes()...)) ||
			// A container's modifiers apply to the traits it holds.
			(p.hasModifierCriteria() && !modifiersMeet(&p.ModifierCriteria, replacements, t.AllModifiers())) {
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
			levels = t.levelForPrereqsOf(exclude)
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
	return text + p.describeModifier(replacements, em) + describeTitleNote(p.TitleNoteCriteria, false, replacements, em)
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
	p.hashModifier(h)
	p.TitleNoteCriteria.Hash(h)
}
