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
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var skillBasedDefaultTypes = map[string]bool{
	SkillID: true,
	ParryID: true,
	BlockID: true,
}

// SkillDefault holds data for a Skill default.
type SkillDefault struct {
	DefaultType    string          `json:"type"`
	Name           criteria.Text   `json:"name,omitzero"`
	Specialization criteria.Text   `json:"specialization,omitzero"`
	Tags           criteria.Text   `json:"tags,omitzero"`
	Modifier       fxp.Int         `json:"modifier,omitzero"`
	Level          fxp.Int         `json:"level,omitzero"`
	AdjLevel       fxp.Int         `json:"adjusted_level,omitzero"`
	Points         fxp.Int         `json:"points,omitzero"`
	WhenTL         criteria.Number `json:"when_tl,omitzero"`
}

// migrateStringToCriteriaText loads a criteria.Text that may have been written in the old format, where it was a plain
// string rather than an object. The raw value was captured by the caller rather than decoded in place, so the caller's
// options have to be handed back in for the re-parse to be made under the same terms as the decode it came from.
func migrateStringToCriteriaText(raw jsontext.Value, dst *criteria.Text, opts json.Options) error {
	if len(raw) == 0 {
		return nil
	}
	if raw[0] == '"' {
		var str string
		if err := json.Unmarshal(raw, &str, opts); err != nil {
			return err
		}
		if str != "" {
			dst.Compare = criteria.IsText
			dst.Qualifier = str
		}
		return nil
	}
	return json.Unmarshal(raw, dst, opts)
}

// DefaultTypeIsSkillBased returns true if the SkillDefault type is Skill-based.
func DefaultTypeIsSkillBased(skillDefaultType string) bool {
	return skillBasedDefaultTypes[normalizeDefaultType(skillDefaultType)]
}

// normalizeDefaultType reduces a SkillDefault type to the form the IDs are written in. Only SetType() sanitizes the
// type, so a data file not written by GCS may hold something like "Parry" or " dx ", which every classifier here
// (DefaultTypeIsSkillBased, SkillBased, FullName) already accepts. Level resolution has to agree with them, or such a
// default is displayed and treated as a parry/block/skill default everywhere except where its level is computed.
func normalizeDefaultType(skillDefaultType string) string {
	return strings.ToLower(strings.TrimSpace(skillDefaultType))
}

// CloneWithoutLevelOrPoints creates a copy, but without the level or points set.
func (s *SkillDefault) CloneWithoutLevelOrPoints() *SkillDefault {
	clone := *s
	clone.Level = 0
	clone.AdjLevel = 0
	clone.Points = 0
	return &clone
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (s *SkillDefault) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var localData struct {
		DefaultType    string          `json:"type"`
		Name           jsontext.Value  `json:"name,omitzero"`
		Specialization jsontext.Value  `json:"specialization,omitzero"`
		Tags           criteria.Text   `json:"tags,omitzero"`
		Modifier       fxp.Int         `json:"modifier,omitzero"`
		Level          fxp.Int         `json:"level,omitzero"`
		AdjLevel       fxp.Int         `json:"adjusted_level,omitzero"`
		Points         fxp.Int         `json:"points,omitzero"`
		WhenTL         criteria.Number `json:"when_tl,omitzero"`
	}
	if err := json.UnmarshalDecode(dec, &localData); err != nil {
		return err
	}
	s.DefaultType = localData.DefaultType
	s.Tags = localData.Tags
	s.Modifier = localData.Modifier
	s.Level = localData.Level
	s.AdjLevel = localData.AdjLevel
	s.Points = localData.Points
	s.WhenTL = localData.WhenTL
	if err := migrateStringToCriteriaText(localData.Name, &s.Name, dec.Options()); err != nil {
		return err
	}
	return migrateStringToCriteriaText(localData.Specialization, &s.Specialization, dec.Options())
}

// Equivalent returns true if this can be considered equivalent to other.
func (s *SkillDefault) Equivalent(replacements map[string]string, other *SkillDefault) bool {
	return other != nil &&
		s.Type() == other.Type() &&
		s.Modifier == other.Modifier &&
		s.WhenTL == other.WhenTL &&
		s.NameWithReplacements(replacements) == other.NameWithReplacements(replacements) &&
		s.SpecializationWithReplacements(replacements) == other.SpecializationWithReplacements(replacements) &&
		sameTagCriteria(s.Tags, other.Tags, replacements)
}

// sameTagCriteria reports whether two tag criteria select the same skills, judged the way Hash and IsZero judge them:
// every "is anything" criteria is the same as every other, whatever qualifier it may still carry from an earlier
// selection and whatever unknown comparison it may hold, and otherwise the comparison and the qualifier (after
// replacements) must both agree.
func sameTagCriteria(a, b criteria.Text, replacements map[string]string) bool {
	if a.IsZero() && b.IsZero() {
		return true
	}
	return a.Compare.EnsureValid() == b.Compare.EnsureValid() &&
		nameable.Apply(a.Qualifier, replacements) == nameable.Apply(b.Qualifier, replacements)
}

// Type returns the type of the SkillDefault, normalized to the form the IDs are written in.
func (s *SkillDefault) Type() string {
	return normalizeDefaultType(s.DefaultType)
}

// SetType sets the type of the SkillDefault.
func (s *SkillDefault) SetType(t string) {
	s.DefaultType = SanitizeID(t, true)
}

// FullName returns the full name of the skill to default from. A default that names a skill outright is described by
// that name, with the specialization in parentheses when it names one; any other selection is spelled out criteria by
// criteria, in the words the editor uses for them, so that nothing the default matches on goes unmentioned.
func (s *SkillDefault) FullName(entity *Entity, replacements map[string]string) string {
	if !s.SkillBased() {
		return ResolveAttributeName(entity, s.Type())
	}
	var buffer strings.Builder
	if s.namesSkill(replacements) {
		buffer.WriteString(s.NameWithReplacements(replacements))
		if spec := s.SpecializationWithReplacements(replacements); spec != "" && !s.Specialization.IsZero() {
			buffer.WriteString(" (")
			buffer.WriteString(spec)
			buffer.WriteByte(')')
		}
	} else {
		buffer.WriteString(i18n.Text("any skill"))
		var clauses []string
		if !s.Name.IsZero() {
			prefix := i18n.Text("whose name")
			clauses = append(clauses, s.Name.StringWithPrefix(replacements, prefix, prefix))
		}
		if !s.Specialization.IsZero() {
			prefix := i18n.Text("whose specialization")
			clauses = append(clauses, s.Specialization.StringWithPrefix(replacements, prefix, prefix))
		}
		if !s.Tags.IsZero() {
			prefix, notPrefix := i18n.Text("at least one tag"), i18n.Text("all tags")
			if len(clauses) == 0 {
				prefix, notPrefix = i18n.Text("where at least one tag"), i18n.Text("where all tags")
			}
			clauses = append(clauses, s.Tags.StringWithPrefix(replacements, prefix, notPrefix))
		}
		for i, clause := range clauses {
			if i == 0 {
				buffer.WriteByte(' ')
			} else {
				buffer.WriteString(i18n.Text(" and "))
			}
			buffer.WriteString(clause)
		}
	}
	switch s.Type() {
	case DodgeID:
		buffer.WriteString(i18n.Text(" Dodge"))
	case ParryID:
		buffer.WriteString(i18n.Text(" Parry"))
	case BlockID:
		buffer.WriteString(i18n.Text(" Block"))
	}
	return buffer.String()
}

// namesSkill reports whether this default names a skill outright: its name criteria is "is" some name, its
// specialization criteria is either "is anything" or "is" some specialization (an empty one meaning a skill without
// one), and it asks nothing of the tags. That is the ordinary kind of default, and the one FullName describes by the
// name alone.
func (s *SkillDefault) namesSkill(replacements map[string]string) bool {
	return s.Name.Compare.EnsureValid() == criteria.IsText &&
		s.NameWithReplacements(replacements) != "" &&
		(s.Specialization.IsZero() || s.Specialization.Compare.EnsureValid() == criteria.IsText) &&
		s.Tags.IsZero()
}

// NameWithReplacements returns the name of the skill to default from with any nameable keys replaced.
func (s *SkillDefault) NameWithReplacements(replacements map[string]string) string {
	return nameable.Apply(s.Name.Qualifier, replacements)
}

// SpecializationWithReplacements returns the specialization of the skill to default from with any nameable keys
// replaced.
func (s *SkillDefault) SpecializationWithReplacements(replacements map[string]string) string {
	return nameable.Apply(s.Specialization.Qualifier, replacements)
}

// FillWithNameableKeys adds any nameable keys found in this SkillDefault to the provided map.
func (s *SkillDefault) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(
		m, existing,
		s.Name.Qualifier,
		s.Specialization.Qualifier,
		s.Tags.Qualifier,
	)
}

// ModifierAsString returns the modifier as a string suitable for appending.
func (s *SkillDefault) ModifierAsString() string {
	if s.Modifier != 0 {
		return s.Modifier.StringWithSign()
	}
	return ""
}

// SkillBased returns true if the Type() is Skill-based.
func (s *SkillDefault) SkillBased() bool {
	return skillBasedDefaultTypes[s.Type()]
}

// SkillLevel returns the base skill level for this SkillDefault.
func (s *SkillDefault) SkillLevel(entity *Entity, replacements map[string]string, requirePoints bool, excludes map[string]bool, ruleOf20 bool) fxp.Int {
	switch s.Type() {
	case ParryID:
		best := s.best(entity, replacements, requirePoints, excludes)
		if best != fxp.Min {
			best = best.Div(fxp.Two).Floor() + fxp.Three + entity.ParryBonus
		}
		return s.finalLevel(best)
	case BlockID:
		best := s.best(entity, replacements, requirePoints, excludes)
		if best != fxp.Min {
			best = best.Div(fxp.Two).Floor() + fxp.Three + entity.BlockBonus
		}
		return s.finalLevel(best)
	case SkillID:
		return s.finalLevel(s.best(entity, replacements, requirePoints, excludes))
	default:
		return s.SkillLevelFast(entity, replacements, requirePoints, excludes, ruleOf20)
	}
}

// isTLPermitted reports whether the WhenTL constraint (if any) is satisfied. skillTL is the tech level of the skill the
// default is resolving against; when empty (the skill has no tech level, or the default isn't skill-based), it falls
// back to the entity's tech level.
func (s *SkillDefault) isTLPermitted(entity *Entity, skillTL string) bool {
	if s.WhenTL.Compare == criteria.AnyNumber {
		return true
	}
	tlStr := skillTL
	if tlStr == "" {
		if entity == nil {
			return true
		}
		tlStr = entity.Profile.TechLevel
	}
	tl, _, _ := ExtractTechLevel(tlStr)
	if tl < 0 {
		tl = 0
	}
	return s.WhenTL.Compare.Matches(s.WhenTL.Qualifier, tl)
}

// matchingSkills returns the skills this SkillDefault selects. On top of the name and specialization criteria, the tag
// criteria (when one has been set) filters the result: the qualifier may hold a comma-separated list of tags, where a
// positive comparison (is, contains, starts with, ...) matches a skill having any one tag that matches any one of the
// qualifiers, while a negative one (is not, does not contain, ...) requires every tag of the skill to fail against
// every qualifier. A skill carrying no tags at all counts as having a single, empty tag.
func (s *SkillDefault) matchingSkills(entity *Entity, replacements map[string]string, requirePoints bool, excludes map[string]bool) []*Skill {
	if entity == nil {
		return nil
	}
	list := entity.SkillMatching(s.Name, s.Specialization, replacements, requirePoints, excludes)
	if s.Tags.IsZero() {
		return list
	}
	return slices.DeleteFunc(list, func(sk *Skill) bool {
		return !s.Tags.MatchesList(replacements, sk.Tags...)
	})
}

// bestMatchingSkill returns the highest-level skill this SkillDefault selects, or nil if it selects none.
func (s *SkillDefault) bestMatchingSkill(entity *Entity, replacements map[string]string, requirePoints bool, excludes map[string]bool) *Skill {
	return BestSkillIn(s.matchingSkills(entity, replacements, requirePoints, excludes), excludes)
}

// bestFastMatchingSkill returns the skill bestFast() scores this SkillDefault by: the one with the highest already-
// calculated level among those the default selects and whose tech level satisfies the WhenTL constraint, with the
// first one encountered winning a tie.
func (s *SkillDefault) bestFastMatchingSkill(entity *Entity, replacements map[string]string, requirePoints bool, excludes map[string]bool) *Skill {
	var best *Skill
	level := fxp.Min
	for _, sk := range s.matchingSkills(entity, replacements, requirePoints, excludes) {
		if !s.isTLPermitted(entity, sk.TL()) {
			continue
		}
		if level < sk.LevelData.Level {
			best = sk
			level = sk.LevelData.Level
		}
	}
	return best
}

func (s *SkillDefault) best(entity *Entity, replacements map[string]string, requirePoints bool, excludes map[string]bool) fxp.Int {
	best := fxp.Min
	for _, sk := range s.matchingSkills(entity, replacements, requirePoints, excludes) {
		if !s.isTLPermitted(entity, sk.TL()) {
			continue
		}
		if best < sk.LevelData.Level {
			level := sk.CalculateLevel(excludes).Level
			if best < level {
				best = level
			}
		}
	}
	return best
}

// SkillLevelFast returns the base skill level for this SkillDefault.
func (s *SkillDefault) SkillLevelFast(entity *Entity, replacements map[string]string, requirePoints bool, excludes map[string]bool, ruleOf20 bool) fxp.Int {
	switch s.Type() {
	case DodgeID:
		if !s.isTLPermitted(entity, "") {
			return fxp.Min
		}
		level := entity.Dodge(entity.EncumbranceLevel(false))
		if ruleOf20 && level > 20 {
			level = 20
		}
		return s.finalLevel(fxp.FromInteger(level))
	case ParryID:
		best := s.bestFast(entity, replacements, requirePoints, excludes)
		if best != fxp.Min {
			best = best.Div(fxp.Two).Floor() + fxp.Three + entity.ParryBonus
		}
		return s.finalLevel(best)
	case BlockID:
		best := s.bestFast(entity, replacements, requirePoints, excludes)
		if best != fxp.Min {
			best = best.Div(fxp.Two).Floor() + fxp.Three + entity.BlockBonus
		}
		return s.finalLevel(best)
	case SkillID:
		return s.finalLevel(s.bestFast(entity, replacements, requirePoints, excludes))
	default:
		if !s.isTLPermitted(entity, "") {
			return fxp.Min
		}
		level := entity.ResolveAttributeCurrent(s.Type())
		if level != fxp.Min {
			if ruleOf20 {
				level = level.Min(fxp.Twenty)
			}
			if entity.SheetSettings.UseHalfStatDefaults {
				level = level.Div(fxp.Two).Floor() + fxp.Five
			}
		}
		return s.finalLevel(level)
	}
}

// asDefense returns this SkillDefault re-pointed at the given defense (ParryID or BlockID) when it names the other
// defense, and the receiver unchanged otherwise. A "Cloak Parry" default names the Cloak skill; the parry conversion
// SkillLevelFast() applies to it is the wrong one when the caller is computing a block, and halving the resulting
// parry level a second time would fold the parry bonus into the block on top of that. Re-pointing it makes such a
// default contribute exactly what a plain skill default to the same skill would, with the defense being computed
// supplying its own +3 and its own bonus. Note that this is only right between the two defenses -- the weapon's attack
// skill deliberately leaves a defense-type default alone; see Weapon.SkillLevel.
func (s *SkillDefault) asDefense(defenseID string) *SkillDefault {
	switch s.Type() {
	case ParryID, BlockID:
		if s.Type() != defenseID {
			repointed := *s
			repointed.DefaultType = defenseID
			return &repointed
		}
	}
	return s
}

// defenseLevelFast returns the defense level for a defense-type default (one whose Type() is ParryID or BlockID),
// folding skillAdj into the named skill's level before the halving that turns it into a defense level and adding
// defenseBonus afterwards. SkillLevelFast() performs that halving itself, so a caller holding a skill-level adjustment
// -- a minimum-ST penalty or a bonus aimed at this weapon's skill -- cannot simply add it to that result: at defense
// scale the adjustment would count for twice what it does on the skill-type default path.
func (s *SkillDefault) defenseLevelFast(entity *Entity, replacements map[string]string, skillAdj, defenseBonus fxp.Int) fxp.Int {
	best := s.bestFast(entity, replacements, false, nil)
	if best == fxp.Min {
		return fxp.Min
	}
	return s.finalLevel((best + skillAdj).Div(fxp.Two).Floor() + fxp.Three + defenseBonus)
}

func (s *SkillDefault) bestFast(entity *Entity, replacements map[string]string, requirePoints bool, excludes map[string]bool) fxp.Int {
	if sk := s.bestFastMatchingSkill(entity, replacements, requirePoints, excludes); sk != nil {
		return sk.LevelData.Level
	}
	return fxp.Min
}

func (s *SkillDefault) finalLevel(level fxp.Int) fxp.Int {
	if level != fxp.Min {
		level += s.Modifier
	}
	return level
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (s *SkillDefault) Hash(h hash.Hash) {
	xhash.StringWithLen(h, s.DefaultType)
	xhash.Num64(h, s.Modifier)
	s.Name.Hash(h)
	s.Specialization.Hash(h)
	if !s.WhenTL.IsZero() {
		// Only hash this when its not the default, so that old files don't suddenly become marked as modified.
		s.WhenTL.Hash(h)
	}
	if !s.Tags.IsZero() {
		// Only hash this when its not the default, so that old files don't suddenly become marked as modified.
		s.Tags.Hash(h)
	}
}
