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
	"fmt"
	"hash"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Prereq = &PrereqList{}

// PrereqList holds a prereq that contains a list of prerequisites.
type PrereqList struct {
	Parent  *PrereqList     `json:"-"`
	Type    prereq.Type     `json:"type"`
	All     bool            `json:"all"`
	WhenTL  criteria.Number `json:"when_tl,omitzero"`
	Prereqs Prereqs         `json:"prereqs,omitempty"`
}

// NewPrereqList creates a new PrereqList.
func NewPrereqList() *PrereqList {
	return &PrereqList{
		Type: prereq.List,
		All:  true,
	}
}

// IsZero reports whether json's omitzero option should omit this value.
func (p *PrereqList) IsZero() bool {
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

// CloneResolvingEmpty clones this prereq list. A nil list becomes a new, empty one unless isContainer is true, and an
// empty list becomes nil if pruneIfEmpty is true.
func (p *PrereqList) CloneResolvingEmpty(isContainer, pruneIfEmpty bool) *PrereqList {
	if p != nil {
		if pruneIfEmpty && p.IsZero() {
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

// Satisfied implements Prereq. hasEquipmentPenalty, if not nil, is set to true when this list is unsatisfied and an
// unmet equipped-equipment prerequisite is among what made it so: one of its own, or one reached through nested lists
// that are each unsatisfied. An "all of" list that also fails for some other prerequisite sets it, while an unmet
// equipment prerequisite inside a satisfied nested list, or in a list that does not apply at the sheet's tech level,
// never does.
func (p *PrereqList) Satisfied(entity *Entity, exclude any, buffer *xbytes.InsertBuffer, prefix string, hasEquipmentPenalty *bool) bool {
	if entity == nil {
		return true
	}
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
	var local *xbytes.InsertBuffer
	if buffer != nil {
		local = &xbytes.InsertBuffer{}
	}
	eqpPenalty := false
	for _, one := range p.Prereqs {
		if one.Satisfied(entity, exclude, local, prefix, &eqpPenalty) {
			count++
		}
	}
	if local != nil && local.Len() != 0 {
		indented := strings.ReplaceAll(local.String(), "\n", "\n\t")
		local = &xbytes.InsertBuffer{}
		local.WriteString(indented)
	}
	satisfied := count == len(p.Prereqs) || (!p.All && count > 0)
	if !satisfied {
		if eqpPenalty && hasEquipmentPenalty != nil {
			*hasEquipmentPenalty = true
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

// Describe implements Prereq. The children are joined with "and" or "or" to match the list's mode, a nested list with
// more than one child is parenthesized, and a tech level condition is noted at the end.
func (p *PrereqList) Describe(replacements map[string]string, em func(string) string) string {
	return p.describeChildren(replacements, em) + p.describeWhenTL()
}

func (p *PrereqList) describeChildren(replacements map[string]string, em func(string) string) string {
	parts := make([]string, 0, len(p.Prereqs))
	for i, one := range p.Prereqs {
		list, isList := one.(*PrereqList)
		var text string
		if isList {
			text = list.describeChildren(replacements, em)
		} else {
			text = one.Describe(replacements, em)
		}
		if i != 0 {
			text = lowerFirst(text)
		}
		if isList {
			if len(list.Prereqs) > 1 {
				text = "(" + text + ")"
			}
			text += list.describeWhenTL()
		}
		parts = append(parts, text)
	}
	if p.All {
		return strings.Join(parts, i18n.Text(" and "))
	}
	return strings.Join(parts, i18n.Text(" or "))
}

func (p *PrereqList) describeWhenTL() string {
	if p.WhenTL.Compare == criteria.AnyNumber {
		return ""
	}
	return fmt.Sprintf(i18n.Text(" (only when TL %s)"), p.WhenTL.AltString())
}

// lowerFirst lowercases the first letter of text when it begins a word that continues in lowercase, so a description
// can follow a joining word without lowering a name or an acronym.
func lowerFirst(text string) string {
	r, size := utf8.DecodeRuneInString(text)
	if next, _ := utf8.DecodeRuneInString(text[size:]); unicode.IsUpper(r) && (unicode.IsLower(next) || next == '\'') {
		return string(unicode.ToLower(r)) + text[size:]
	}
	return text
}

// Hash writes this object's contents into the hasher.
func (p *PrereqList) Hash(h hash.Hash) {
	if p.IsZero() {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, p.Type)
	xhash.Bool(h, p.All)
	p.WhenTL.Hash(h)
	hashList(h, p.Prereqs)
}
