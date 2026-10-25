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
	"strings"

	"github.com/dop251/goja"
)

func deferredNewScriptSkill(skill *Skill) ScriptSelfProvider {
	return deferredScriptSelf(skill, newScriptSkill)
}

func newScriptSkill(r *goja.Runtime, skill *Skill) *goja.Object {
	m := make(map[string]func() goja.Value)
	addScriptNodeIdentity(r, m, skill, skill.Tags, newScriptSkill)
	m["name"] = func() goja.Value { return r.ToValue(skill.NameWithReplacements()) }
	m["notes"] = scriptNotes(r, skill)
	if skill.Container() {
		m["children"] = func() goja.Value {
			return scriptObjects(r, skill.Children, scriptSees((*Skill).Enabled), newScriptSkill)
		}
		m["find"] = func() goja.Value {
			return r.ToValue(func(call goja.FunctionCall) goja.Value {
				name := callArgAsString(call, 0)
				specialization := callArgAsString(call, 1)
				tag := callArgAsString(call, 2)
				return findScriptSkills(r, name, specialization, tag, skill.Children...)
			})
		}
	} else {
		m["switchedOn"] = func() goja.Value { return r.ToValue(skill.SwitchedOn) }
		m["specialization"] = func() goja.Value { return r.ToValue(skill.SpecializationWithReplacements()) }
		m["optionalSpecialization"] = func() goja.Value {
			return r.ToValue(skill.OptionalSpecializationWithReplacements())
		}
		m["techLevel"] = func() goja.Value {
			if skill.TechLevel != nil {
				return r.ToValue(*skill.TechLevel)
			}
			return r.ToValue("")
		}
		m["kind"] = func() goja.Value {
			if skill.IsTechnique() {
				return r.ToValue("technique")
			}
			return r.ToValue("skill")
		}
		m["attribute"] = func() goja.Value { return r.ToValue(skill.Difficulty.Attribute) }
		m["difficulty"] = func() goja.Value { return r.ToValue(skill.Difficulty.Difficulty.Key()) }
		m["points"] = func() goja.Value { return r.ToValue(skill.PointsCost().AsInteger[int]()) }
		m["level"] = func() goja.Value {
			entity := EntityFromNode(skill)
			if entity == nil {
				return r.ToValue(0)
			}
			return r.ToValue(scriptLevel(updateSkillLevelForScript(entity, skill)))
		}
		m["relativeLevel"] = func() goja.Value {
			entity := EntityFromNode(skill)
			if entity == nil {
				return r.ToValue(0)
			}
			return r.ToValue(scriptRelativeLevel(updateSkillLevelForScript(entity, skill)))
		}
		addScriptWeapons(r, m, func() []*Weapon { return skill.Weapons })
	}
	return r.NewDynamicObject(NewScriptObject(r, m))
}

// updateSkillLevelForScript returns the skill's level for a script, as scriptSkillLevel gives it, guarding against a
// script that resolves a skill's level via itself, which gets the level recorded. The guard is removed with a defer so
// that a panic inside UpdateLevel (recovered further up by xos.SafeCall) can't leave the exclusion registered on the
// entity, which would permanently freeze that skill's script-visible level.
func updateSkillLevelForScript(entity *Entity, skill *Skill) Level {
	name := skill.NameWithReplacements()
	specialization := skill.SpecializationWithReplacements()
	optionalSpecialization := skill.OptionalSpecializationWithReplacements()
	if entity.isSkillLevelResolutionExcluded(name, specialization, optionalSpecialization) {
		return skill.LevelData
	}
	entity.registerSkillLevelResolutionExclusion(name, specialization, optionalSpecialization)
	defer entity.unregisterSkillLevelResolutionExclusion(name, specialization, optionalSpecialization)
	return scriptSkillLevel(entity, skill)
}

// scriptSkillLevel returns the skill's level for a script, recalculated so that it is current. While the script runs
// for the prerequisites of an item carrying grants, the level is computed as those prerequisites see it, without what
// the grants gave the skill, and is not recorded, since the sheet has the level with them; see Skill.levelForPrereqsOf.
func scriptSkillLevel(entity *Entity, skill *Skill) Level {
	if entity.carriesGrant(entity.judging) {
		return skill.calculateLevelWith(entity.judging, skill.bestDefaultWithPoints(nil), nil)
	}
	skill.UpdateLevel()
	return skill.LevelData
}

// findScriptSkills is findScriptNodes for skills, which additionally match on a specialization: a skill's required or
// optional specialization must equal the one asked for, unless none was.
func findScriptSkills(r *goja.Runtime, name, specialization, tag string, topLevelSkills ...*Skill) goja.Value {
	return findScriptNodes(r, name, tag, scriptNodeKind[*Skill]{
		ctor:   newScriptSkill,
		nameOf: (*Skill).NameWithReplacements,
		tagsOf: func(skill *Skill) []string { return skill.Tags },
	}, func(skill *Skill) bool {
		return specialization == "" || strings.EqualFold(skill.SpecializationWithReplacements(), specialization) ||
			strings.EqualFold(skill.OptionalSpecializationWithReplacements(), specialization)
	}, topLevelSkills...)
}
