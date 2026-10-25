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
	"unicode"
	"unicode/utf8"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xreflect"
)

var _ Prereq = &PrereqList{}

// PrereqList holds a prereq that contains a list of prerequisites. Not swaps met and unmet, making "all of" into "not
// all of" and "any of" into "none of".
type PrereqList struct {
	Parent  *PrereqList     `json:"-"`
	Type    prereq.Type     `json:"type"`
	All     bool            `json:"all"`
	Not     bool            `json:"not,omitzero"`
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

// Evaluate checks this list against the entity for the item given as exclude, returning its result. A list is skipped,
// along with everything in it, when it does not apply at the sheet's tech level, and is also skipped when it has
// nothing in it that isn't skipped. Its parent leaves it out. Of the rest, one unmet item makes an "all of" list unmet,
// and one met item makes an "any of" list met, even when others failed. Otherwise the list has failed when any of the
// rest has, and is met or unmet as they all are. A negated list then swaps met and unmet, leaving skipped and failed
// alone. A prerequisite of a type this version of GCS doesn't know has failed, so that negating it can't satisfy it.
// visit, if not nil, is called with the result of each prerequisite in the list and then of the list itself, along with
// the reason a script gives, which is the error when it couldn't run. Each script runs at most once, and none in a
// skipped list does. Without an entity, the list is met and nothing is visited.
func (p *PrereqList) Evaluate(entity *Entity, exclude any, visit func(one Prereq, result CheckResult, reason string)) CheckResult {
	var result CheckResult
	entity.JudgingPrereqsOf(exclude, func() {
		result, _ = p.evaluate(&prereqEvaluation{entity: entity, exclude: exclude, visit: visit}, nil, nil, p.All, false)
	})
	return result
}

// Satisfied implements Prereq. A list is satisfied when Evaluate finds it met or skipped. hasEquipmentPenalty, if not
// nil, is set to true when this list is unsatisfied partly because equipment it requires to be equipped isn't: an unmet
// prerequisite to have it equipped, or, inside a negated list, a met prerequisite not to, reached through nested lists
// that each count against the one holding them. An "all of" list that also fails for some other prerequisite sets it,
// while one inside a nested list that doesn't count against its parent never does.
//
// The text written to buffer lists what is unmet. A skipped list writes nothing, a list with a single unmet item writes
// just that item, and a nested list with the same mode as its parent writes its items alongside its parent's. Any other
// list writes a heading with its items indented beneath it, which for a negated list are those that are met, or those
// that failed when the list did. This list is treated as though its caller were an "all of" list, so the items of an
// "all of" list are written without a heading.
func (p *PrereqList) Satisfied(entity *Entity, exclude any, buffer *xbytes.InsertBuffer, prefix string, hasEquipmentPenalty *bool) bool {
	var result CheckResult
	entity.JudgingPrereqsOf(exclude, func() {
		result, _ = p.evaluate(&prereqEvaluation{entity: entity, exclude: exclude, prefix: prefix}, buffer,
			hasEquipmentPenalty, p.All, false)
	})
	return result == CheckMet || result == CheckSkipped
}

// prereqEvaluation holds what stays the same throughout an evaluation of a prerequisite list.
type prereqEvaluation struct {
	entity  *Entity
	exclude any
	visit   func(one Prereq, result CheckResult, reason string)
	prefix  string
}

// evaluate is Evaluate, also writing the text and setting the equipment penalty that Satisfied describes, and returning
// how many items of text it wrote at the level of the prefix. flatten requests that the unmet items of a list that
// isn't negated be written at that level rather than under a heading. negated is set when the lists holding this one
// are negated an odd number of times, which turns what each prerequisite asks for around: a met one then counts
// against the outermost list, and one that doesn't have equipment equipped asks that it be equipped.
func (p *PrereqList) evaluate(ev *prereqEvaluation, buffer *xbytes.InsertBuffer, hasEquipmentPenalty *bool, flatten, negated bool) (result CheckResult, items int) {
	if ev.entity == nil {
		return CheckMet, 0
	}
	if !p.AppliesAt(ev.entity) {
		p.visitSkipped(ev.visit)
		return CheckSkipped, 0
	}
	var tally checkTally
	var local *xbytes.InsertBuffer
	if buffer != nil {
		local = &xbytes.InsertBuffer{}
	}
	var met []Prereq
	eqpPenalty := false
	inner := negated != p.Not
	for _, one := range p.Prereqs {
		out := local
		if p.Not && local != nil {
			// What is unmet doesn't count against a negated list, so only what failed is kept.
			out = &xbytes.InsertBuffer{}
		}
		childResult := CheckMet
		if list, ok := one.(*PrereqList); ok {
			var n int
			childResult, n = list.evaluate(ev, out, &eqpPenalty, list.All == p.All && !p.Not, inner)
			items += n
		} else {
			var reason string
			penalty := &eqpPenalty
			if inner {
				penalty = nil
			}
			switch typed := one.(type) {
			case *ScriptPrereq:
				childResult, reason = typed.evaluate(ev.entity, ev.exclude, out, ev.prefix)
			case *UnknownPrereq:
				typed.Satisfied(ev.entity, ev.exclude, out, ev.prefix, nil)
				childResult = CheckFailed
			default:
				if !one.Satisfied(ev.entity, ev.exclude, out, ev.prefix, penalty) {
					childResult = CheckUnmet
				}
			}
			// Negated, one that doesn't have equipment equipped asks that it be, and is unmet when it is met.
			if eqp, isEqp := one.(*EquippedEquipmentPrereq); isEqp && inner && childResult == CheckMet && eqp.Not &&
				eqp.Type == prereq.EquippedEquipment {
				eqpPenalty = true
			}
			if childResult != CheckMet {
				items++
			}
			if ev.visit != nil {
				ev.visit(one, childResult, reason)
			}
		}
		if p.Not && local != nil {
			switch childResult {
			case CheckMet:
				met = append(met, one)
			case CheckFailed:
				local.WriteString(out.String())
			default:
			}
		}
		tally.add(childResult)
	}
	result = tally.combine(p.All).negate(p.Not)
	if ev.visit != nil {
		ev.visit(p, result, "")
	}
	// The penalty goes up when this list counts against the lists holding it: when it has failed, which no negation
	// changes, when it is unmet and they don't negate it, or when it is met and they do.
	if eqpPenalty && hasEquipmentPenalty != nil && result != CheckSkipped &&
		(result == CheckFailed || (result == CheckMet) == negated) {
		*hasEquipmentPenalty = true
	}
	if result == CheckMet || result == CheckSkipped {
		return result, 0
	}
	if buffer == nil {
		return result, 0
	}
	if p.Not {
		buffer.WriteString(ev.prefix)
		if p.All {
			buffer.WriteString(i18n.Text("Requires not all of:"))
		} else {
			buffer.WriteString(i18n.Text("Requires none of:"))
		}
		if result == CheckFailed {
			buffer.WriteString(strings.ReplaceAll(local.String(), "\n", "\n\t"))
			return result, 1
		}
		var replacements map[string]string
		if na, ok := ev.exclude.(nameable.Accesser); ok {
			replacements = na.NameableReplacements()
		}
		for _, one := range met {
			buffer.WriteString(strings.ReplaceAll(ev.prefix+one.Describe(ev.entity, replacements, plainText), "\n",
				"\n\t"))
		}
		return result, 1
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
func (p *PrereqList) visitSkipped(visit func(one Prereq, result CheckResult, reason string)) {
	if visit == nil {
		return
	}
	for _, one := range p.Prereqs {
		if list, ok := one.(*PrereqList); ok {
			list.visitSkipped(visit)
		} else {
			visit(one, CheckSkipped, "")
		}
	}
	visit(p, CheckSkipped, "")
}

// Describe implements Prereq. The children are joined with "and" or "or" to match the list's mode, a nested list that
// joins more than one is parenthesized, a negated list says so, a tech level condition is noted at the end, and empty
// lists are left out.
func (p *PrereqList) Describe(entity *Entity, replacements map[string]string, em func(string) string) string {
	text, n, all := p.describeChildren(entity, replacements, em, p.Not)
	if n != 0 {
		text = p.describeNot(text, all)
	}
	return text + p.describeWhenTL()
}

// describeChildren joins the descriptions of the children, returning how many it joined and whether they are joined
// with "and". When lower is true, the first of them follows a joining word, as every other one does, and so begins in
// lowercase unless it is a name someone wrote. A negated list that holds only a nested list joining several items
// joins those items instead, in that list's mode, since negating one thing is the same whichever way it is combined.
func (p *PrereqList) describeChildren(entity *Entity, replacements map[string]string, em func(string) string, lower bool) (text string, count int, all bool) {
	parts := make([]string, 0, len(p.Prereqs))
	// bare is the text of a nested list that joins more than one item, without its parentheses, and bareAll its mode.
	var bare string
	var bareAll bool
	for _, one := range p.Prereqs {
		lower = lower || len(parts) != 0
		if list, isList := one.(*PrereqList); isList {
			var n int
			var listAll bool
			if text, n, listAll = list.describeChildren(entity, replacements, em, lower || list.Not); n == 0 {
				continue
			}
			switch {
			case list.Not:
				if text = list.describeNot(text, listAll); lower {
					text = lowerFirst(text)
				}
			case n > 1:
				if list.WhenTL.Compare == criteria.AnyNumber {
					bare, bareAll = text, listAll
				}
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
	if p.Not && len(parts) == 1 && bare != "" {
		return bare, 1, bareAll
	}
	joiner := i18n.Text(" or ")
	if p.All {
		joiner = i18n.Text(" and ")
	}
	return strings.Join(parts, joiner), len(parts), p.All
}

// describeNot returns the text of the children of a negated list, joined with "and" when all is set, as the list says
// them, or unchanged when it isn't one.
func (p *PrereqList) describeNot(text string, all bool) string {
	switch {
	case !p.Not:
		return text
	case all:
		return i18n.Text("Not all of (%s)", text)
	default:
		return i18n.Text("None of (%s)", text)
	}
}

// describeWhenTL returns the note that ends a description of this list when it has a tech level condition, or an empty
// string when it has none.
func (p *PrereqList) describeWhenTL() string {
	if p.WhenTL.Compare == criteria.AnyNumber {
		return ""
	}
	return i18n.Text(" (only when TL %s)", p.WhenTL.AltString())
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
	xhash.Bool(h, p.Not)
	p.WhenTL.Hash(h)
	hashList(h, p.Prereqs)
}
