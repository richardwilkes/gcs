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
	"encoding/json/v2"
	"hash"
	"maps"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/levelcap"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// SkillLevelCap holds the level of the skills or spells it selects to at most what its cap resolves to (see
// Entity.capLevel). The criteria match the item's own values, as a prerequisite's do: the specialization criteria
// applies only to skills and the college and power source criteria only to spells, each ignored for the other kind, so
// that a cap on skills and spells alike can still narrow either by name and tags. A container is never matched. Cap is
// never nil once the cap has been constructed or loaded. A target this version of GCS doesn't know selects nothing,
// and its key is written back out as it was read.
type SkillLevelCap struct {
	Target                 levelcap.Target `json:"target"`
	NameCriteria           criteria.Text   `json:"name,omitzero"`
	SpecializationCriteria criteria.Text   `json:"specialization,omitzero"`
	TagsCriteria           criteria.Text   `json:"tags,omitzero"`
	CollegeCriteria        criteria.Text   `json:"college,omitzero"`
	PowerSourceCriteria    criteria.Text   `json:"power_source,omitzero"`
	Cap                    *LevelCapGroup  `json:"cap"`
	unknownTarget          string
}

// skillLevelCapData avoids recursion into the JSON methods of SkillLevelCap.
type skillLevelCapData SkillLevelCap

// skillLevelCapJSON is a SkillLevelCap as JSON holds it, with the key of its target, if present, in place of the
// target.
type skillLevelCapJSON struct {
	Target *string `json:"target"`
	skillLevelCapData
}

// NewSkillLevelCap creates a new cap that applies to every skill and holds nothing to cap at yet.
func NewSkillLevelCap() *SkillLevelCap {
	c := &SkillLevelCap{Target: levelcap.Skills}
	c.EnsureValidity()
	return c
}

// Clone creates a deep copy of this cap.
func (c *SkillLevelCap) Clone() *SkillLevelCap {
	clone := *c
	if c.Cap != nil {
		clone.Cap = c.Cap.CloneAsLevelCapGroup(nil)
	}
	clone.EnsureValidity()
	return &clone
}

// EnsureValidity makes sure the target is valid and there is a root group, and walks the tree to set each node's type
// and parent.
func (c *SkillLevelCap) EnsureValidity() {
	c.Target = c.Target.EnsureValid()
	if c.Cap == nil {
		c.Cap = NewLevelCapGroup(nil)
	}
	ensureLevelCapValidity(c.Cap, nil)
}

// MarshalJSONTo implements json.MarshalerTo.
func (c *SkillLevelCap) MarshalJSONTo(enc *jsontext.Encoder) error {
	target := c.Target.Key()
	if c.Target == levelcap.UnknownTarget && c.unknownTarget != "" {
		target = c.unknownTarget
	}
	return json.MarshalEncode(enc, &skillLevelCapJSON{Target: &target, skillLevelCapData: skillLevelCapData(*c)})
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom. A cap without a target applies to skills.
func (c *SkillLevelCap) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var data skillLevelCapJSON
	if err := json.UnmarshalDecode(dec, &data); err != nil {
		return err
	}
	*c = SkillLevelCap(data.skillLevelCapData)
	if data.Target != nil {
		var known bool
		if c.Target, known = levelcap.ExtractKnownTarget(*data.Target); !known {
			c.unknownTarget = *data.Target
		}
	}
	c.EnsureValidity()
	return nil
}

// Hash writes this object's contents into the hasher.
func (c *SkillLevelCap) Hash(h hash.Hash) {
	if c == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, c.Target)
	if c.Target == levelcap.UnknownTarget {
		xhash.StringWithLen(h, c.unknownTarget)
	}
	c.NameCriteria.Hash(h)
	c.SpecializationCriteria.Hash(h)
	c.TagsCriteria.Hash(h)
	c.CollegeCriteria.Hash(h)
	c.PowerSourceCriteria.Hash(h)
	c.Cap.Hash(h)
}

// AppliesToSkill returns true if the cap selects the skill.
func (c *SkillLevelCap) AppliesToSkill(sk *Skill) bool {
	return !sk.Container() && c.applies(skillLevelCapSubject(sk))
}

// AppliesToSpell returns true if the cap selects the spell.
func (c *SkillLevelCap) AppliesToSpell(sp *Spell) bool {
	return !sp.Container() && c.applies(spellLevelCapSubject(sp))
}

// Describe returns a plain-language description of the cap, such as `Skills whose name contains "Magic" are capped at
// the lowest of 12 and skill Thaumatology -2`, naming attributes as the entity, which may be nil, defines them. Names,
// qualifiers and numbers are passed through em, which may wrap them for emphasis; pass an identity func for plain
// text.
func (c *SkillLevelCap) Describe(entity *Entity, em func(string) string) string {
	return i18n.Text("%s are capped at %s", c.DescribeSubject(em), c.Cap.Describe(entity, em))
}

// DescribeSubject returns the part of Describe that says what the cap applies to, such as `Spells of the Fire college`.
func (c *SkillLevelCap) DescribeSubject(em func(string) string) string {
	text := c.Target.String()
	switch c.NameCriteria.Compare.EnsureValid() {
	case criteria.AnyText:
	case criteria.IsText:
		text += i18n.Text(" named ") + describeName(c.NameCriteria, nil, em)
	default:
		text += i18n.Text(" whose name ") + describeText(c.NameCriteria, nil, em)
	}
	if c.Target.MatchesSkills() {
		text += describeSpecialization(c.SpecializationCriteria, criteria.Text{}, nil, em)
	}
	text += describeTags(c.TagsCriteria, nil, em)
	if c.Target.MatchesSpells() {
		if c.CollegeCriteria.Compare.EnsureValid() != criteria.AnyText {
			text += describeSpellList(c.CollegeCriteria, true, i18n.Text(" whose college "), i18n.Text(" without colleges"),
				i18n.Text(" with at least one college"), nil, em)
		}
		if c.PowerSourceCriteria.Compare.EnsureValid() != criteria.AnyText {
			text += i18n.Text(" whose power source ") + describeText(c.PowerSourceCriteria, nil, em)
		}
	}
	return text
}

// levelCapSubject is what a cap is matched against: the values of a skill or spell, resolved once rather than once per
// cap.
type levelCapSubject struct {
	spell                  bool
	name                   string
	specialization         string
	optionalSpecialization string
	powerSource            string
	tags                   []string
	colleges               []string
}

func skillLevelCapSubject(sk *Skill) *levelCapSubject {
	return &levelCapSubject{
		name:                   sk.NameWithReplacements(),
		specialization:         sk.SpecializationWithReplacements(),
		optionalSpecialization: sk.OptionalSpecializationWithReplacements(),
		tags:                   sk.Tags,
	}
}

func spellLevelCapSubject(sp *Spell) *levelCapSubject {
	return &levelCapSubject{
		spell:       true,
		name:        sp.NameWithReplacements(),
		powerSource: sp.PowerSourceWithReplacements(),
		tags:        sp.Tags,
		colleges:    sp.CollegeWithReplacements(),
	}
}

// applies returns true if the cap selects the subject. The criteria hold no nameable markers, so they are matched
// without replacements.
func (c *SkillLevelCap) applies(s *levelCapSubject) bool {
	if s.spell {
		if !c.Target.MatchesSpells() {
			return false
		}
	} else if !c.Target.MatchesSkills() {
		return false
	}
	if !c.NameCriteria.Matches(nil, s.name) || !c.TagsCriteria.MatchesList(nil, s.tags...) {
		return false
	}
	if s.spell {
		return c.CollegeCriteria.MatchesList(nil, s.colleges...) && c.PowerSourceCriteria.Matches(nil, s.powerSource)
	}
	return skillSpecializationMatches(c.SpecializationCriteria, nil, s.specialization, s.optionalSpecialization)
}

// SkillLevelCaps holds the skill level caps of a SheetSettings (see SheetSettingsData.SkillLevelCaps).
type SkillLevelCaps []*SkillLevelCap

// UnmarshalJSONFrom implements json.UnmarshalerFrom. An entry that can't be read is logged and left out rather than
// failing the whole document, which would set a settings file aside, or keep a sheet from opening, over one cap.
func (l *SkillLevelCaps) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	*l = nil
	if dec.PeekKind() != '[' {
		value, err := dec.ReadValue()
		if err != nil {
			return err
		}
		if value.Kind() != 'n' {
			errs.Log(errs.New("ignoring the skill level caps, which aren't a list"))
		}
		return nil
	}
	if _, err := dec.ReadToken(); err != nil {
		return err
	}
	for dec.PeekKind() != ']' {
		value, err := dec.ReadValue()
		if err != nil {
			return err
		}
		var one *SkillLevelCap
		if err = jio.Unmarshal(value, &one); err != nil {
			errs.Log(errs.NewWithCause("ignoring a skill level cap that can't be read", err))
			continue
		}
		if one != nil {
			*l = append(*l, one)
		}
	}
	_, err := dec.ReadToken()
	return err
}

// Clone returns a deep copy of the list, or nil when it is empty.
func (l SkillLevelCaps) Clone() SkillLevelCaps {
	if len(l) == 0 {
		return nil
	}
	clone := make(SkillLevelCaps, 0, len(l))
	for _, one := range l {
		if one != nil {
			clone = append(clone, one.Clone())
		}
	}
	return clone
}

// EnsureValidity drops the entries that are nil and makes the rest valid.
func (l *SkillLevelCaps) EnsureValidity() {
	*l = slices.DeleteFunc(*l, func(one *SkillLevelCap) bool { return one == nil })
	for _, one := range *l {
		one.EnsureValidity()
	}
}

// Hash writes this object's contents into the hasher.
func (l SkillLevelCaps) Hash(h hash.Hash) {
	hashList(h, l)
}

// capSkillLevel returns the skill's level held to the sheet's level caps (see capLevel), the penalty being the
// encumbrance penalty CalculateSkillLevel folds into the level of a skill that isn't a technique. matched is the skill
// as it is matched against the caps and its penalty taken from, which may be a copy holding edits (see
// LevelWithGrants), and self the skill on the sheet, for which matched stands in while the caps are resolved, as it
// would once the edits are applied; see matchLevelCapsAs. The caps are resolved without the skill in the excludes, where
// a technique resting on it puts it (see techniqueBaseLevel), as they are when its level is computed for the sheet,
// since it is passed over for having its caps resolved instead, so that a level they rest on may still rest on it.
func (e *Entity) capSkillLevel(matched, self *Skill, exclude any, level Level, excludes map[string]bool) Level {
	if e == nil || e.SheetSettings == nil || len(e.SheetSettings.SkillLevelCaps) == 0 || level.Level == fxp.Min ||
		matched.Container() {
		return level
	}
	if matched != self {
		defer e.matchLevelCapsAs(self, matched)()
	}
	if name := matched.String(); excludes[name] {
		excludes = maps.Clone(excludes)
		delete(excludes, name)
	}
	var penalty fxp.Int
	if !matched.IsTechnique() && matched.EncumbrancePenaltyMultiplier != 0 {
		penalty = e.EncumbranceLevel(true).Penalty().Mul(matched.EncumbrancePenaltyMultiplier)
	}
	return e.capLevel(matched, skillLevelCapSubject(matched), exclude, level, penalty, excludes)
}

// capSpellLevel is capSkillLevel for a spell, except that a copy holding edits only changes what the references match,
// since no skill's or spell's level rests on a spell's.
func (e *Entity) capSpellLevel(matched, self *Spell, exclude any, level Level) Level {
	if e == nil || e.SheetSettings == nil || len(e.SheetSettings.SkillLevelCaps) == 0 || level.Level == fxp.Min ||
		matched.Container() {
		return level
	}
	if matched != self {
		defer e.matchLevelCapsAs(self, matched)()
	}
	return e.capLevel(self, spellLevelCapSubject(matched), exclude, level, 0, nil)
}

// levelCapEdit is a skill or spell on the sheet and a copy of it holding an editor's edits, along with the rows grants
// put on the sheet that applying the edits takes off it, the skills or spells grants would put there once the edits
// are applied and what the grants would then give the other rows they merge into, when that changes; see
// Entity.matchLevelCapsAs.
type levelCapEdit struct {
	onSheet       any
	edited        any
	mergedAway    map[any]bool
	leftBehind    map[any]*grantState
	plantedSkills []*Skill
	plantedSpells []*Spell
}

// matchLevelCapsAs makes the references of the level caps match onSheet, a skill or spell on the sheet, as edited, a
// copy of it holding an editor's edits, until the returned func is called; see ReferenceLevelCap.resolve. A skill is
// found as edited by every lookup of the skills by name (see skillEdit), and the defaults of the skills are picked,
// and their levels computed, afresh rather than read as recorded (see Skill.defaultExcept and SkillDefault.best), so
// that a level resting on it, as a technique's or a default's does, rests on it as edited. The defaults picked while
// the edits are previewed are kept apart from any picked before. The rows of the grants that merge into it as edited
// in place of their own, which applying the edits takes off the sheet, are found neither by those lookups nor by the
// references; see takenOffByEdit. The rows the grants merging into it now would put on the sheet in its place are
// found by both, after the rows on the sheet. The other rows the grants merge into now, or would then, have their
// levels computed with those that would then merge into them, as they would once the edits are applied; see
// grantsMovedByEdit and grantsAsEdited.
func (e *Entity) matchLevelCapsAs(onSheet, edited any) (restore func()) {
	saved, savedDefaults := e.levelCapEdit, e.judgedDefaults
	edit := levelCapEdit{onSheet: onSheet, edited: edited}
	switch one := onSheet.(type) {
	case *Skill:
		if copyOf, ok := edited.(*Skill); ok {
			edit.mergedAway, edit.leftBehind, edit.plantedSkills = grantsMovedByEdit(one, copyOf)
		}
	case *Spell:
		if copyOf, ok := edited.(*Spell); ok {
			edit.mergedAway, edit.leftBehind, edit.plantedSpells = grantsMovedByEdit(one, copyOf)
		}
	}
	e.levelCapEdit, e.judgedDefaults = edit, nil
	return func() { e.levelCapEdit, e.judgedDefaults = saved, savedDefaults }
}

// takenOffByEdit returns true if the item is a row a grant put on the sheet that applying the edits an editor previews
// takes off it, as the grant merges into the item edited instead; see matchLevelCapsAs.
func (e *Entity) takenOffByEdit(item any) bool {
	return e != nil && e.levelCapEdit.mergedAway[item]
}

// grantsAsEdited returns the grant state the level and points of the skill or spell holder are computed with: while an
// editor previews edits that change which grants merge into it, the state those that would then merge into it give it,
// and otherwise its own; see matchLevelCapsAs.
func (e *Entity) grantsAsEdited(holder grantHolder) *grantState {
	if e != nil {
		if state := e.levelCapEdit.leftBehind[holder]; state != nil {
			return state
		}
	}
	return holder.grants()
}

// previewingSkillEdit returns true while an editor previews the level of a skill as edited; see matchLevelCapsAs.
func (e *Entity) previewingSkillEdit() bool {
	onSheet, _ := e.skillEdit()
	return onSheet != nil
}

// skillEdit returns the skill on the sheet an editor previews the level of, and the copy of it holding the edits, which
// stands in for it, or nils when there is none; see matchLevelCapsAs.
func (e *Entity) skillEdit() (onSheet, edited *Skill) {
	if e != nil {
		var ok bool
		if onSheet, ok = e.levelCapEdit.onSheet.(*Skill); ok {
			if edited, ok = e.levelCapEdit.edited.(*Skill); ok {
				return onSheet, edited
			}
		}
	}
	return nil, nil
}

// capLevel returns the level held to the lowest of the sheet's level caps that select the subject and resolve, with
// the relative level lowered by as much and a line saying so added to the tooltip, or the level as it is when none do
// or it is already no higher. A cap never raises a level. penalty, the encumbrance penalty the level carries, if any,
// is left out of the level compared with the caps and still applies to the level held. self, the item on the sheet,
// is noted as having its caps resolved while they are, so that a cap resting on its level, or on a level that rests
// on it, passes it over (see ReferenceLevelCap.resolve) rather than recursing without end; the caps are still resolved
// when its level is reached again that way, so that a cap resting on nothing in the cycle still holds. The levels the
// caps rest on are computed afresh, with the excludes of the default chain, rather than read as recorded, so that the
// result is a function of the sheet's data alone and settles in a pass.
func (e *Entity) capLevel(self any, subject *levelCapSubject, exclude any, level Level, penalty fxp.Int, excludes map[string]bool) Level {
	if !e.levelCapsResolving[self] {
		if e.levelCapsResolving == nil {
			e.levelCapsResolving = make(map[any]bool)
		}
		e.levelCapsResolving[self] = true
		defer delete(e.levelCapsResolving, self)
	}
	r := levelCapResolver{entity: e, exclude: e.judgingOr(exclude), excludes: excludes}
	var lowest levelCapValue
	found := false
	for _, c := range e.SheetSettings.SkillLevelCaps {
		if c == nil || c.Cap == nil || !c.applies(subject) {
			continue
		}
		if v, ok := c.Cap.resolve(&r); ok && (!found || v.level < lowest.level) {
			lowest = v
			found = true
		}
	}
	if !found {
		return level
	}
	limit := lowest.level.Floor()
	if limit >= level.Level-penalty {
		return level
	}
	delta := limit - (level.Level - penalty)
	level.Level = limit + penalty
	level.RelativeLevel += delta
	level.Tooltip += lowest.tooltipLine(limit, delta)
	return level
}

// hasLevelCaps returns true if the sheet has level caps.
func (e *Entity) hasLevelCaps() bool {
	return e != nil && e.SheetSettings != nil && len(e.SheetSettings.SkillLevelCaps) != 0
}

// levelCapRefsHolding returns the references to a skill or spell of the sheet's level caps that select the skill or
// spell, or none for anything else.
func (e *Entity) levelCapRefsHolding(item any) []*ReferenceLevelCap {
	if !e.hasLevelCaps() {
		return nil
	}
	var subject *levelCapSubject
	switch one := item.(type) {
	case *Skill:
		if !one.Container() {
			subject = skillLevelCapSubject(one)
		}
	case *Spell:
		if !one.Container() {
			subject = spellLevelCapSubject(one)
		}
	}
	if subject == nil {
		return nil
	}
	var refs []*ReferenceLevelCap
	for _, c := range e.SheetSettings.SkillLevelCaps {
		if c != nil && c.Cap != nil && c.applies(subject) {
			refs = appendSkillAndSpellRefs(refs, c.Cap)
		}
	}
	return refs
}

// passOver notes the item, a technique or Ritual Magic spell as the references of the level caps find it (see
// ReferenceLevelCap.resolve), as having its caps resolved, unless it already is, until the returned func is called,
// while the level of the skill it rests on is computed. A cap met on the way that rests on the item, as one holding
// that skill may, then passes it over, as it does while the item's own caps are resolved (see capLevel), rather than
// holding that skill, and with it the item, to the item's level computed afresh. A nil item passes nothing over.
func (e *Entity) passOver(item any) (restore func()) {
	if item == nil || !e.hasLevelCaps() || e.levelCapsResolving[item] {
		return func() {}
	}
	if e.levelCapsResolving == nil {
		e.levelCapsResolving = make(map[any]bool)
	}
	e.levelCapsResolving[item] = true
	return func() { delete(e.levelCapsResolving, item) }
}
