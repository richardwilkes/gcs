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
	"encoding/json/jsontext"
	"hash"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/levelcap"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xreflect"
)

var (
	_ LevelCap = &LevelCapGroup{}
	_ LevelCap = &FixedLevelCap{}
	_ LevelCap = &ReferenceLevelCap{}
	_ LevelCap = &UnknownLevelCap{}
)

// LevelCap is a node of the tree a SkillLevelCap holds its cap in: a fixed level, the level of a skill, spell or trait
// or the value of an attribute, each with an adjustment, a group that takes the lowest or highest of its children, or
// a node this version of GCS doesn't understand but preserves.
type LevelCap interface {
	// CapType returns the kind of node this is.
	CapType() levelcap.Type
	// ParentGroup returns the owning group, if any.
	ParentGroup() *LevelCapGroup
	// SetParentGroup sets the owning group.
	SetParentGroup(parent *LevelCapGroup)
	// Clone creates a deep copy of this node, owned by parent.
	Clone(parent *LevelCapGroup) LevelCap
	// Describe returns a plain-language description of the cap, naming attributes as the entity, which may be nil,
	// defines them. Names, qualifiers and numbers are passed through em, which may wrap them for emphasis; pass an
	// identity func for plain text.
	Describe(entity *Entity, em func(string) string) string
	// Hash writes this object's contents into the hasher.
	Hash(h hash.Hash)
	// resolve returns the level the node stands for on the resolver's sheet, or false when nothing it refers to is
	// there.
	resolve(r *levelCapResolver) (levelCapValue, bool)
}

// LevelCaps holds a list of level cap nodes.
type LevelCaps []LevelCap

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (l *LevelCaps) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	list, err := unmarshalTypedList(dec, levelcap.ExtractKnownType, allocLevelCap,
		func(kind string, raw jsontext.Value) LevelCap { return NewUnknownLevelCap(kind, raw) })
	if err != nil {
		return err
	}
	*l = list
	return nil
}

// allocLevelCap returns an empty LevelCap of the concrete type that represents capType, or nil if there is none.
func allocLevelCap(capType levelcap.Type) LevelCap {
	switch capType {
	case levelcap.Group:
		return &LevelCapGroup{}
	case levelcap.Fixed:
		return &FixedLevelCap{}
	case levelcap.Skill, levelcap.Spell, levelcap.Trait, levelcap.Attribute:
		return &ReferenceLevelCap{}
	default:
		return nil
	}
}

// ensureLevelCapValidity sets the node's parent and type, drops any nil children, and recurses.
func ensureLevelCapValidity(node LevelCap, parent *LevelCapGroup) {
	node.SetParentGroup(parent)
	switch n := node.(type) {
	case *LevelCapGroup:
		n.Type = levelcap.Group
		n.Children = slices.DeleteFunc(n.Children, func(child LevelCap) bool { return xreflect.IsNil(child) })
		for _, child := range n.Children {
			ensureLevelCapValidity(child, n)
		}
	case *FixedLevelCap:
		n.Type = levelcap.Fixed
	case *ReferenceLevelCap:
		n.normalize()
	}
}

// levelCapResolver is what resolving a cap reads the sheet with: the entity, the item whose prerequisites are being
// judged, if any (see Entity.judgingOr), and the excludes of the skill default chain the level being capped is part of,
// less the skill being capped (see Entity.capSkillLevel).
type levelCapResolver struct {
	entity   *Entity
	exclude  any
	excludes map[string]bool
}

// levelCapValue is a resolved cap: the level and, for a cap resting on something on the sheet, what it rests on, as
// the level's tooltip names it.
type levelCapValue struct {
	level  fxp.Int
	source string
}

// tooltipLine returns the line the tooltip of a level held to the cap gets, in the form of the bonus lines it holds.
func (v levelCapValue) tooltipLine(limit, delta fxp.Int) string {
	text := i18n.Text("\nLevel cap of %s", limit.String())
	if v.source != "" {
		text += " (" + v.source + ")"
	}
	return text + " [" + delta.StringWithSign() + "]"
}

// LevelCapGroup takes the lowest of its children, or the highest when Highest is set. A child that can't be resolved
// (see ReferenceLevelCap) is left out, and a group with no children, or none that resolve, can't be resolved itself,
// so it caps nothing.
type LevelCapGroup struct {
	Parent   *LevelCapGroup `json:"-"`
	Type     levelcap.Type  `json:"type"`
	Highest  bool           `json:"highest,omitzero"`
	Children LevelCaps      `json:"children,omitempty"`
	// Unknown holds the members this release does not know, so that they survive a load/save cycle unchanged.
	Unknown jsontext.Value `json:",embed"`
}

// NewLevelCapGroup creates a new, empty group that takes the lowest of its children.
func NewLevelCapGroup(parent *LevelCapGroup) *LevelCapGroup {
	return &LevelCapGroup{Parent: parent, Type: levelcap.Group}
}

// CapType implements LevelCap.
func (g *LevelCapGroup) CapType() levelcap.Type {
	return levelcap.Group
}

// ParentGroup implements LevelCap.
func (g *LevelCapGroup) ParentGroup() *LevelCapGroup {
	if g == nil {
		return nil
	}
	return g.Parent
}

// SetParentGroup implements LevelCap.
func (g *LevelCapGroup) SetParentGroup(parent *LevelCapGroup) {
	g.Parent = parent
}

// Clone implements LevelCap.
func (g *LevelCapGroup) Clone(parent *LevelCapGroup) LevelCap {
	return g.CloneAsLevelCapGroup(parent)
}

// CloneAsLevelCapGroup creates a deep copy of this group, owned by parent.
func (g *LevelCapGroup) CloneAsLevelCapGroup(parent *LevelCapGroup) *LevelCapGroup {
	clone := *g
	clone.Parent = parent
	clone.Unknown = slices.Clone(g.Unknown)
	clone.Children = make(LevelCaps, len(g.Children))
	for i, child := range g.Children {
		clone.Children[i] = child.Clone(&clone)
	}
	return &clone
}

// Hash implements LevelCap.
func (g *LevelCapGroup) Hash(h hash.Hash) {
	if g == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, levelcap.Group)
	xhash.Bool(h, g.Highest)
	hashList(h, g.Children)
}

// Empty returns true if nothing but groups lies within this group, at any depth, so that it caps nothing. A node this
// version of GCS doesn't recognize counts.
func (g *LevelCapGroup) Empty() bool {
	for _, child := range g.Children {
		if sub, ok := child.(*LevelCapGroup); !ok || !sub.Empty() {
			return false
		}
	}
	return true
}

// Describe implements LevelCap. A group of one reads as that one, and an empty group as no cap.
func (g *LevelCapGroup) Describe(entity *Entity, em func(string) string) string {
	parts := make([]string, 0, len(g.Children))
	for _, child := range g.Children {
		if sub, ok := child.(*LevelCapGroup); ok && sub.Empty() {
			continue
		}
		parts = append(parts, child.Describe(entity, em))
	}
	switch len(parts) {
	case 0:
		return i18n.Text("no cap")
	case 1:
		return parts[0]
	default:
	}
	list := i18n.Text("%s and %s", strings.Join(parts[:len(parts)-1], i18n.Text(", ")), parts[len(parts)-1])
	if g.Highest {
		return i18n.Text("the highest of %s", list)
	}
	return i18n.Text("the lowest of %s", list)
}

// resolve implements LevelCap. On a tie, the first child wins.
func (g *LevelCapGroup) resolve(r *levelCapResolver) (levelCapValue, bool) {
	var best levelCapValue
	found := false
	for _, child := range g.Children {
		v, ok := child.resolve(r)
		if !ok {
			continue
		}
		if !found || (g.Highest && v.level > best.level) || (!g.Highest && v.level < best.level) {
			best = v
			found = true
		}
	}
	return best, found
}

// FixedLevelCap is a cap at a set level.
type FixedLevelCap struct {
	Parent *LevelCapGroup `json:"-"`
	Type   levelcap.Type  `json:"type"`
	Level  fxp.Int        `json:"level"`
	// Unknown holds the members this release does not know, so that they survive a load/save cycle unchanged.
	Unknown jsontext.Value `json:",embed"`
}

// NewFixedLevelCap creates a new FixedLevelCap at the level.
func NewFixedLevelCap(parent *LevelCapGroup, level fxp.Int) *FixedLevelCap {
	return &FixedLevelCap{Parent: parent, Type: levelcap.Fixed, Level: level}
}

// CapType implements LevelCap.
func (c *FixedLevelCap) CapType() levelcap.Type {
	return levelcap.Fixed
}

// ParentGroup implements LevelCap.
func (c *FixedLevelCap) ParentGroup() *LevelCapGroup {
	return c.Parent
}

// SetParentGroup implements LevelCap.
func (c *FixedLevelCap) SetParentGroup(parent *LevelCapGroup) {
	c.Parent = parent
}

// Clone implements LevelCap.
func (c *FixedLevelCap) Clone(parent *LevelCapGroup) LevelCap {
	clone := *c
	clone.Parent = parent
	clone.Unknown = slices.Clone(c.Unknown)
	return &clone
}

// Hash implements LevelCap.
func (c *FixedLevelCap) Hash(h hash.Hash) {
	if c == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, levelcap.Fixed)
	xhash.Num64(h, c.Level)
}

// Describe implements LevelCap.
func (c *FixedLevelCap) Describe(_ *Entity, em func(string) string) string {
	return em(c.Level.String())
}

// resolve implements LevelCap.
func (c *FixedLevelCap) resolve(_ *levelCapResolver) (levelCapValue, bool) {
	return levelCapValue{level: c.Level}, true
}

// ReferenceLevelCap is the level of a skill, spell or trait, or the value of an attribute, plus an adjustment. A
// skill, spell or trait is picked by its name, and a skill by its specialization as well, and the highest level among
// those matching is taken. With nothing matching, or an attribute the sheet doesn't have, or nothing matching that has
// a level, the level is taken to be 0, so that a cap such as "Thaumatology -2" holds spells down on a sheet without
// Thaumatology rather than letting them go. One whose own caps are being resolved, or whose level a reference is
// computing, is passed over (see Entity.capLevel), so that no cap rests on a level that rests on it, and a skill the
// default chain of the level being capped excludes is too, as is a technique left with no level for resting on one;
// when nothing else matching has a level, the cap can't be resolved, and is left out.
type ReferenceLevelCap struct {
	Parent                 *LevelCapGroup `json:"-"`
	Type                   levelcap.Type  `json:"type"`
	NameCriteria           criteria.Text  `json:"name,omitzero"`
	SpecializationCriteria criteria.Text  `json:"specialization,omitzero"`
	// Attribute is the ID of the attribute an attribute cap reads, or SizeModifierID or DodgeID, as an AttributePrereq
	// names them.
	Attribute string  `json:"attribute,omitzero"`
	Modifier  fxp.Int `json:"modifier,omitzero"`
	// Unknown holds the members this release does not know, so that they survive a load/save cycle unchanged.
	Unknown jsontext.Value `json:",embed"`
}

// NewReferenceLevelCap creates a new ReferenceLevelCap of the type, which must be Skill, Spell, Trait or Attribute: one
// naming an item outright, or reading the entity's DX. 'entity' may be nil.
func NewReferenceLevelCap(entity *Entity, parent *LevelCapGroup, capType levelcap.Type) *ReferenceLevelCap {
	c := &ReferenceLevelCap{Parent: parent, Type: capType}
	c.normalize()
	if c.Type == levelcap.Attribute {
		c.Attribute = AttributeIDFor(entity, DexterityID)
	} else {
		c.NameCriteria.Compare = criteria.IsText
	}
	return c
}

// normalize makes the type one of the reference types, Skill when it isn't, and clears the data a cap of that type
// makes no use of, so that nothing lingers unseen to be written out or to tell two equivalent caps apart.
func (c *ReferenceLevelCap) normalize() {
	switch c.Type {
	case levelcap.Skill, levelcap.Spell, levelcap.Trait, levelcap.Attribute:
	default:
		c.Type = levelcap.Skill
	}
	if c.Type != levelcap.Skill {
		c.SpecializationCriteria = criteria.Text{}
	}
	if c.Type == levelcap.Attribute {
		c.NameCriteria = criteria.Text{}
	} else {
		c.Attribute = ""
	}
}

// CapType implements LevelCap.
func (c *ReferenceLevelCap) CapType() levelcap.Type {
	return c.Type
}

// ParentGroup implements LevelCap.
func (c *ReferenceLevelCap) ParentGroup() *LevelCapGroup {
	return c.Parent
}

// SetParentGroup implements LevelCap.
func (c *ReferenceLevelCap) SetParentGroup(parent *LevelCapGroup) {
	c.Parent = parent
}

// Clone implements LevelCap.
func (c *ReferenceLevelCap) Clone(parent *LevelCapGroup) LevelCap {
	clone := *c
	clone.Parent = parent
	clone.Unknown = slices.Clone(c.Unknown)
	return &clone
}

// Hash implements LevelCap.
func (c *ReferenceLevelCap) Hash(h hash.Hash) {
	if c == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, c.Type)
	c.NameCriteria.Hash(h)
	c.SpecializationCriteria.Hash(h)
	xhash.StringWithLen(h, c.Attribute)
	xhash.Num64(h, c.Modifier)
}

// Describe implements LevelCap, as "skill Thaumatology -2", "spell Enchantment -2", "trait Magery +10" or "IQ +2".
func (c *ReferenceLevelCap) Describe(entity *Entity, em func(string) string) string {
	text := c.describeTarget(entity, em)
	if c.Modifier != 0 {
		text += " " + em(c.Modifier.StringWithSign())
	}
	return text
}

// describeTarget returns Describe without the modifier: what the cap rests on.
func (c *ReferenceLevelCap) describeTarget(entity *Entity, em func(string) string) string {
	switch c.Type {
	case levelcap.Spell:
		return i18n.Text("spell %s", describeName(c.NameCriteria, nil, em))
	case levelcap.Trait:
		return i18n.Text("trait %s", describeName(c.NameCriteria, nil, em))
	case levelcap.Attribute:
		return em(attributeTitle(entity, c.Attribute))
	default:
		return i18n.Text("skill %s", describeName(c.NameCriteria, nil, em)) +
			describeSpecialization(c.SpecializationCriteria, criteria.Text{}, nil, em)
	}
}

// levelCapMemoKey identifies what a reference resolved to with a view of the sheet: the reference, the item whose
// prerequisites were being judged, the excludes of the default chain, whether it was resolved at base (see
// Entity.atBaseWhile), whether it was resolved while a pool's threshold was found for those prerequisites (see
// Attribute.CurrentThreshold) and whether it was resolved while a pool's base or thresholds were found without the
// features those thresholds hold (see Entity.thresholdFeaturesLeftOut); see Entity.levelCapMemo.
type levelCapMemoKey struct {
	ref       *ReferenceLevelCap
	exclude   any
	excludes  string
	base      bool
	searching bool
	untouched bool
}

// levelCapMemoEntry is what a reference resolved to, or that it couldn't be, along with the items matched while it was,
// by it or by any reference resolved for it. It holds only while none of them is resolving (see
// Entity.levelCapsResolving), since one that is would be passed over.
type levelCapMemoEntry struct {
	value      levelCapValue
	ok         bool
	candidates []any
}

// resolve implements LevelCap. The levels are computed as the prerequisites being judged, if any, see them, rather
// than read as recorded, so that a cap resting on a skill the cap itself holds down doesn't fall a level with every
// pass; see Entity.capLevel. During a recalculation pass, what a reference resolves to is kept and reused for the
// same view of the sheet, since a cap on every skill resolves its references once per skill. The answer is kept with
// every item matched while it is resolved, by it or by any reference resolved for it, and only when none of them is
// still resolving once it is done, which is when none was passed over for resolving before it began; it is reused
// only while none is resolving, since one that is would be passed over.
func (c *ReferenceLevelCap) resolve(r *levelCapResolver) (levelCapValue, bool) {
	e := r.entity
	if c.Type == levelcap.Attribute {
		return c.resolved(attributePrereqValue(e, c.Attribute), attributeTitle(e, c.Attribute), false)
	}
	var key levelCapMemoKey
	var candidates []any
	resolving := func(item any) bool { return e.levelCapsResolving[item] }
	memo := e.recalculating > 0
	if memo {
		key = levelCapMemoKey{
			ref:       c,
			exclude:   r.exclude,
			excludes:  excludesKey(r.excludes),
			base:      e.atBase > 0,
			searching: e.judgedThresholdSearches > 0,
			untouched: e.thresholdFeaturesLeftOut(),
		}
		if entry, ok := e.levelCapMemo[key]; ok && !slices.ContainsFunc(entry.candidates, resolving) {
			if e.levelCapCandidates != nil {
				*e.levelCapCandidates = append(*e.levelCapCandidates, entry.candidates...)
			}
			return entry.value, entry.ok
		}
		outer := e.levelCapCandidates
		e.levelCapCandidates = &candidates
		defer func() {
			e.levelCapCandidates = outer
			if outer != nil {
				*outer = append(*outer, candidates...)
			}
		}()
	}
	best := fxp.Min
	var source string
	skipped := false
	consider := func(item any, level fxp.Int, name string) {
		candidates = append(candidates, item)
		if level != fxp.Min && (source == "" || level > best) {
			best = level
			source = name
		}
	}
	// A skill or spell passed over while not in effect, such as the row a grant puts on the sheet while what the grant
	// gives it is priced (see grantPlan.row), is still passed over rather than taken as absent, so that a level computed
	// with it passed over is the same whether or not it is in effect yet. The spells are found as SkillMatching finds
	// the skills while an editor previews edits.
	switch c.Type {
	case levelcap.Spell:
		visit := func(sp *Spell) bool {
			if (!e.levelCapsResolving[sp] && !sp.Enabled()) || e.takenOffByEdit(sp) {
				return false
			}
			matched := sp
			if edited, ok := e.levelCapEdit.edited.(*Spell); ok && e.levelCapEdit.onSheet == any(sp) {
				matched = edited
			}
			if !grantedFor(e.judging, sp) && c.NameCriteria.Matches(nil, matched.NameWithReplacements()) {
				if e.levelCapsResolving[sp] {
					candidates = append(candidates, sp)
					skipped = true
				} else {
					consider(sp, levelResolving(e, sp, func() Level {
						return sp.calculateLevelExcept(r.exclude, sp.adjustedPointsExcept(r.exclude, nil))
					}), sp.String())
				}
			}
			return false
		}
		Traverse(visit, false, true, e.Spells...)
		if len(e.levelCapEdit.plantedSpells) != 0 {
			Traverse(visit, false, true, e.levelCapEdit.plantedSpells...)
		}
	case levelcap.Trait:
		Traverse(func(t *Trait) bool {
			if !e.excludedWhileJudging(t) && c.NameCriteria.Matches(nil, t.NameWithReplacements()) {
				consider(t, t.levelForPrereqsOf(r.exclude), t.String())
			}
			return false
		}, true, true, e.Traits...)
	default:
		for _, sk := range e.SkillMatching(c.NameCriteria, c.SpecializationCriteria, nil, false, nil) {
			switch {
			case len(r.excludes) != 0 && r.excludes[sk.String()]:
				skipped = true
			case e.levelCapsResolving[sk]:
				candidates = append(candidates, sk)
				skipped = true
			default:
				level := levelResolving(e, sk, func() Level { return sk.calculateLevelExcept(r.exclude, r.excludes) })
				if level == fxp.Min && baseExcluded(e, sk, r.excludes) {
					// It rests on a skill the default chain excludes, so it is passed over as that skill is.
					candidates = append(candidates, sk)
					skipped = true
				} else {
					consider(sk, level, sk.String())
				}
			}
		}
		name := c.NameCriteria.Matcher(nil)
		specialization := c.SpecializationCriteria.Matcher(nil)
		for item := range e.levelCapsResolving {
			if sk, ok := item.(*Skill); ok && !sk.Enabled() && name.Matches(sk.NameWithReplacements()) &&
				skillSpecializationMatcherMatches(specialization, sk.SpecializationWithReplacements(),
					sk.OptionalSpecializationWithReplacements()) && !grantedFor(e.judging, sk) {
				candidates = append(candidates, sk)
				skipped = true
			}
		}
	}
	value, ok := c.resolved(best, source, skipped && best == fxp.Min)
	if memo && !slices.ContainsFunc(candidates, resolving) {
		if e.levelCapMemo == nil {
			e.levelCapMemo = make(map[levelCapMemoKey]levelCapMemoEntry)
		}
		e.levelCapMemo[key] = levelCapMemoEntry{value: value, ok: ok, candidates: candidates}
	}
	return value, ok
}

// baseExcluded returns true if the skill is a technique whose default matches a skill the excludes hold, so that its
// level, computed with them, can't rest on that skill, whose level the default chain is computing.
func baseExcluded(e *Entity, sk *Skill, excludes map[string]bool) bool {
	if len(excludes) == 0 || !sk.restsOnSkill() {
		return false
	}
	for _, base := range sk.TechniqueDefault.matchingSkills(e, sk.Replacements, true, nil) {
		if excludes[base.String()] {
			return true
		}
	}
	return false
}

// matchingAny returns the skills or spells on the sheet that a reference to a skill or spell matches, whether or not
// they are in effect, or none for a reference of another type.
func (c *ReferenceLevelCap) matchingAny(e *Entity) []any {
	var list []any
	switch c.Type {
	case levelcap.Skill:
		name := c.NameCriteria.Matcher(nil)
		specialization := c.SpecializationCriteria.Matcher(nil)
		Traverse(func(sk *Skill) bool {
			if name.Matches(sk.NameWithReplacements()) && skillSpecializationMatcherMatches(specialization,
				sk.SpecializationWithReplacements(), sk.OptionalSpecializationWithReplacements()) {
				list = append(list, sk)
			}
			return false
		}, false, true, e.Skills...)
	case levelcap.Spell:
		Traverse(func(sp *Spell) bool {
			if c.NameCriteria.Matches(nil, sp.NameWithReplacements()) {
				list = append(list, sp)
			}
			return false
		}, false, true, e.Spells...)
	default:
	}
	return list
}

// appendSkillAndSpellRefs appends the references to a skill or spell within the node, at any depth, to refs.
func appendSkillAndSpellRefs(refs []*ReferenceLevelCap, node LevelCap) []*ReferenceLevelCap {
	switch n := node.(type) {
	case *LevelCapGroup:
		for _, child := range n.Children {
			refs = appendSkillAndSpellRefs(refs, child)
		}
	case *ReferenceLevelCap:
		if n.Type == levelcap.Skill || n.Type == levelcap.Spell {
			refs = append(refs, n)
		}
	}
	return refs
}

// levelResolving returns the level calc computes for the item, which is noted as having its caps resolved while it
// does, so that a cap met on the way passes it over rather than resting on the level being computed, which may rest
// on that cap, or computing it again without end.
func levelResolving(e *Entity, item any, calc func() Level) fxp.Int {
	if e.levelCapsResolving == nil {
		e.levelCapsResolving = make(map[any]bool)
	}
	e.levelCapsResolving[item] = true
	defer delete(e.levelCapsResolving, item)
	return calc().Level
}

// resolved returns the cap's value for the level of what it rests on, named by source, which is taken to be 0 when
// there is nothing to rest on (fxp.Min or an empty source), or false when onlyCyclic says that what it would rest on
// was passed over for having its caps resolved or being excluded by the default chain, or resting on a skill that is.
func (c *ReferenceLevelCap) resolved(level fxp.Int, source string, onlyCyclic bool) (levelCapValue, bool) {
	if onlyCyclic {
		return levelCapValue{}, false
	}
	if level == fxp.Min || source == "" {
		level = 0
		source = i18n.Text("no %s", c.describeTarget(nil, plainText))
	}
	if c.Modifier != 0 {
		source += " " + c.Modifier.StringWithSign()
	}
	return levelCapValue{level: level + c.Modifier, source: source}, true
}

// UnknownLevelCap holds a level cap node whose type this version of GCS doesn't recognize. It caps nothing, and its
// original data is written back out unchanged so that a newer version can still use it.
type UnknownLevelCap struct {
	// Parent is the owning group.
	Parent *LevelCapGroup
	// Kind is the unrecognized value that was found in the "type" field.
	Kind string
	// Data is the original JSON for the node, exactly as it was read.
	Data jsontext.Value
}

// NewUnknownLevelCap creates a new UnknownLevelCap holding a copy of the passed-in data.
func NewUnknownLevelCap(kind string, data jsontext.Value) *UnknownLevelCap {
	return &UnknownLevelCap{Kind: kind, Data: slices.Clone(data)}
}

// CapType implements LevelCap.
func (n *UnknownLevelCap) CapType() levelcap.Type {
	return levelcap.Unknown
}

// ParentGroup implements LevelCap.
func (n *UnknownLevelCap) ParentGroup() *LevelCapGroup {
	return n.Parent
}

// SetParentGroup implements LevelCap.
func (n *UnknownLevelCap) SetParentGroup(parent *LevelCapGroup) {
	n.Parent = parent
}

// Clone implements LevelCap.
func (n *UnknownLevelCap) Clone(parent *LevelCapGroup) LevelCap {
	clone := NewUnknownLevelCap(n.Kind, n.Data)
	clone.Parent = parent
	return clone
}

// MarshalJSONTo implements json.MarshalerTo. The original data is written back out as-is.
func (n *UnknownLevelCap) MarshalJSONTo(enc *jsontext.Encoder) error {
	return enc.WriteValue(n.Data)
}

// Hash implements LevelCap.
func (n *UnknownLevelCap) Hash(h hash.Hash) {
	if n == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, levelcap.Unknown)
	xhash.StringWithLen(h, n.Kind)
	xhash.BytesWithLen(h, n.Data)
}

// Describe implements LevelCap.
func (n *UnknownLevelCap) Describe(_ *Entity, em func(string) string) string {
	return i18n.Text(`an unknown kind of level cap ("%s") that needs a newer version of GCS`, em(n.Kind))
}

// resolve implements LevelCap. An unknown cap is never resolved.
func (n *UnknownLevelCap) resolve(_ *levelCapResolver) (levelCapValue, bool) {
	return levelCapValue{}, false
}
