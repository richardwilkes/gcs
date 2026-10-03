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
	"github.com/richardwilkes/toolbox/v2/xreflect"
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

// SetParentList implements Prereq.
func (p *PrereqList) SetParentList(list *PrereqList) {
	p.Parent = list
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

// HasNothingToCheck returns true if this list, which may be nil, holds no prerequisite other than nil entries and lists
// that themselves have nothing to check.
func (p *PrereqList) HasNothingToCheck() bool {
	if p != nil {
		for _, one := range p.Prereqs {
			if xreflect.IsNil(one) {
				continue
			}
			if list, ok := one.(*PrereqList); !ok || !list.HasNothingToCheck() {
				return false
			}
		}
	}
	return true
}

// AppliesAt returns true if this list applies at the tech level of the entity, which may be nil. A list with no tech
// level condition always applies, as does any list when there is no entity.
func (p *PrereqList) AppliesAt(entity *Entity) bool {
	if entity == nil || p.WhenTL.Compare == criteria.AnyNumber {
		return true
	}
	tl, _, _ := ExtractTechLevel(entity.Profile.TechLevel)
	return p.WhenTL.Compare.Matches(p.WhenTL.Qualifier, max(tl, 0))
}

// AppliesWithParentsAt returns true if this list and the lists holding it apply at the tech level of the entity, which
// may be nil.
func (p *PrereqList) AppliesWithParentsAt(entity *Entity) bool {
	for ; p != nil; p = p.Parent {
		if !p.AppliesAt(entity) {
			return false
		}
	}
	return true
}

// PrereqResult is the outcome of checking a prerequisite against a sheet.
type PrereqResult uint8

// Possible PrereqResult values. A skipped prerequisite is left out of the check, and a failed one couldn't be checked,
// because a script that decides it couldn't run.
const (
	PrereqMet PrereqResult = iota
	PrereqUnmet
	PrereqSkipped
	PrereqFailed
)

// Evaluate checks this list against the entity for the item given as exclude, returning its result. A list is skipped,
// along with everything in it, when it does not apply at the sheet's tech level, and is also skipped when it has
// nothing in it that isn't skipped. Its parent leaves it out, so an "all of" list is met when the rest are all met and
// an "any of" list when any of the rest is. A list that isn't met has failed when any of the rest has, and is otherwise
// unmet. visit, if not nil, is called with the result of each prerequisite in the list and then of the list itself,
// along with the reason a script gives, which is the error when it couldn't run. Each script runs at most once, and
// none in a skipped list does. Without an entity, the list is met and nothing is visited.
func (p *PrereqList) Evaluate(entity *Entity, exclude any, visit func(one Prereq, result PrereqResult, reason string)) PrereqResult {
	result, _ := p.evaluate(&prereqEvaluation{entity: entity, exclude: exclude, visit: visit}, nil, nil, p.All)
	return result
}

// Satisfied implements Prereq. A list is satisfied when Evaluate finds it met or skipped. hasEquipmentPenalty, if not
// nil, is set to true when this list is unsatisfied and an unmet equipped-equipment prerequisite is among what made it
// so: one of its own, or one reached through nested lists that are each unsatisfied. An "all of" list that also fails
// for some other prerequisite sets it, while an unmet equipment prerequisite inside a satisfied or skipped nested list
// never does.
//
// The text written to buffer lists what is unmet. A skipped list writes nothing, a list with a single unmet item
// writes just that item, and a nested list with the same mode as its parent writes its items alongside its parent's.
// Any other list writes a heading with its items indented beneath it. This list is treated as though its caller were
// an "all of" list, so the items of an "all of" list are written without a heading.
func (p *PrereqList) Satisfied(entity *Entity, exclude any, buffer *xbytes.InsertBuffer, prefix string, hasEquipmentPenalty *bool) bool {
	result, _ := p.evaluate(&prereqEvaluation{entity: entity, exclude: exclude, prefix: prefix}, buffer,
		hasEquipmentPenalty, p.All)
	return result == PrereqMet || result == PrereqSkipped
}

// prereqEvaluation holds what stays the same throughout an evaluation of a prerequisite list.
type prereqEvaluation struct {
	entity  *Entity
	exclude any
	visit   func(one Prereq, result PrereqResult, reason string)
	prefix  string
}

// evaluate is Evaluate, also writing the text and setting the equipment penalty that Satisfied describes, and returning
// how many items of text it wrote at the level of the prefix. flatten requests that the unmet items be written at that
// level rather than under a heading.
func (p *PrereqList) evaluate(ev *prereqEvaluation, buffer *xbytes.InsertBuffer, hasEquipmentPenalty *bool, flatten bool) (result PrereqResult, items int) {
	if ev.entity == nil {
		return PrereqMet, 0
	}
	if !p.AppliesAt(ev.entity) {
		p.visitSkipped(ev.visit)
		return PrereqSkipped, 0
	}
	met, applicable, failed := 0, 0, false
	var local *xbytes.InsertBuffer
	if buffer != nil {
		local = &xbytes.InsertBuffer{}
	}
	eqpPenalty := false
	for _, one := range p.Prereqs {
		childResult := PrereqMet
		if list, ok := one.(*PrereqList); ok {
			var n int
			childResult, n = list.evaluate(ev, local, &eqpPenalty, list.All == p.All)
			items += n
		} else {
			var reason string
			if script, isScript := one.(*ScriptPrereq); isScript {
				childResult, reason = script.evaluate(ev.entity, ev.exclude, local, ev.prefix)
			} else if !one.Satisfied(ev.entity, ev.exclude, local, ev.prefix, &eqpPenalty) {
				childResult = PrereqUnmet
			}
			if childResult != PrereqMet {
				items++
			}
			if ev.visit != nil {
				ev.visit(one, childResult, reason)
			}
		}
		switch childResult {
		case PrereqSkipped:
			continue
		case PrereqMet:
			met++
		case PrereqFailed:
			failed = true
		default:
		}
		applicable++
	}
	switch {
	case applicable == 0:
		result = PrereqSkipped
	case met == applicable || (!p.All && met > 0):
		result = PrereqMet
	case failed:
		result = PrereqFailed
	default:
		result = PrereqUnmet
	}
	if ev.visit != nil {
		ev.visit(p, result, "")
	}
	if result == PrereqMet || result == PrereqSkipped {
		return result, 0
	}
	if eqpPenalty && hasEquipmentPenalty != nil {
		*hasEquipmentPenalty = true
	}
	if buffer == nil {
		return result, 0
	}
	if flatten || items == 1 {
		buffer.WriteString(local.String())
		return result, items
	}
	buffer.WriteString(ev.prefix)
	if p.All {
		buffer.WriteString(i18n.Text("Requires all of:"))
	} else {
		buffer.WriteString(i18n.Text("Requires at least one of:"))
	}
	buffer.WriteString(strings.ReplaceAll(local.String(), "\n", "\n\t"))
	return result, 1
}

// visitSkipped calls visit, if not nil, with each prerequisite in this list and then the list itself as skipped.
func (p *PrereqList) visitSkipped(visit func(one Prereq, result PrereqResult, reason string)) {
	if visit == nil {
		return
	}
	for _, one := range p.Prereqs {
		if list, ok := one.(*PrereqList); ok {
			list.visitSkipped(visit)
		} else {
			visit(one, PrereqSkipped, "")
		}
	}
	visit(p, PrereqSkipped, "")
}

// Describe implements Prereq. The children are joined with "and" or "or" to match the list's mode, a nested list that
// joins more than one is parenthesized, a tech level condition is noted at the end, and empty lists are left out.
func (p *PrereqList) Describe(entity *Entity, replacements map[string]string, em func(string) string) string {
	text, _ := p.describeChildren(entity, replacements, em, false)
	return text + p.describeWhenTL()
}

// describeChildren joins the descriptions of the children, returning how many it joined. When lower is true, the first
// of them follows a joining word, as every other one does, and so begins in lowercase unless it is a name someone
// wrote.
func (p *PrereqList) describeChildren(entity *Entity, replacements map[string]string, em func(string) string, lower bool) (text string, count int) {
	parts := make([]string, 0, len(p.Prereqs))
	for _, one := range p.Prereqs {
		lower = lower || len(parts) != 0
		if list, isList := one.(*PrereqList); isList {
			var n int
			if text, n = list.describeChildren(entity, replacements, em, lower); n == 0 {
				continue
			}
			if n > 1 {
				text = "(" + text + ")"
			}
			text += list.describeWhenTL()
		} else {
			text = one.Describe(entity, replacements, em)
			if script, isScript := one.(*ScriptPrereq); lower && (!isScript || script.ResolvedName(replacements) == "") {
				text = lowerFirst(text)
			}
		}
		parts = append(parts, text)
	}
	joiner := i18n.Text(" or ")
	if p.All {
		joiner = i18n.Text(" and ")
	}
	return strings.Join(parts, joiner), len(parts)
}

// describeWhenTL returns the note that ends a description of this list when it has a tech level condition, or an empty
// string when it has none.
func (p *PrereqList) describeWhenTL() string {
	if p.WhenTL.Compare == criteria.AnyNumber {
		return ""
	}
	return fmt.Sprintf(i18n.Text(" (only when TL %s)"), p.WhenTL.AltString())
}

// lowerFirst lowercases the first letter of text when it begins a word that continues in lowercase or is a single
// letter, so a description can follow a joining word without lowering a name or an acronym.
func lowerFirst(text string) string {
	r, size := utf8.DecodeRuneInString(text)
	if next, _ := utf8.DecodeRuneInString(text[size:]); unicode.IsUpper(r) && (unicode.IsLower(next) || next == '\'' || next == ' ') {
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
