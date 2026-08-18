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
	"slices"
	"strings"

	"github.com/dop251/goja"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
)

func deferredNewScriptSkill(skill *Skill) ScriptSelfProvider {
	if skill == nil {
		return ScriptSelfProvider{}
	}
	return ScriptSelfProvider{
		ID:       string(skill.TID),
		Provider: func(r *goja.Runtime) any { return newScriptSkill(r, skill) },
	}
}

func newScriptSkill(r *goja.Runtime, skill *Skill) *goja.Object {
	m := make(map[string]func() goja.Value)
	m["id"] = func() goja.Value { return r.ToValue(string(skill.TID)) }
	m["parentID"] = func() goja.Value {
		if skill.parent == nil {
			return goja.Undefined()
		}
		return r.ToValue(string(skill.parent.TID))
	}
	m["parent"] = func() goja.Value {
		if skill.parent == nil {
			return goja.Undefined()
		}
		return newScriptSkill(r, skill.parent)
	}
	m["name"] = func() goja.Value { return r.ToValue(skill.NameWithReplacements()) }
	m["notes"] = func() goja.Value {
		return r.ToValue(skill.SecondaryText(func(_ display.Option) bool { return true }))
	}
	m["tags"] = func() goja.Value { return r.ToValue(slices.Clone(skill.Tags)) }
	m["container"] = func() goja.Value { return r.ToValue(skill.Container()) }
	m["switchedOn"] = func() goja.Value { return r.ToValue(skill.SwitchedOn) }
	if skill.Container() {
		m["children"] = func() goja.Value {
			children := make([]*goja.Object, 0, len(skill.Children))
			for _, child := range skill.Children {
				if child.Enabled() {
					children = append(children, newScriptSkill(r, child))
				}
			}
			return r.ToValue(children)
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
		m["points"] = func() goja.Value { return r.ToValue(fxp.AsInteger[int](skill.AdjustedPoints(nil))) }
		m["level"] = func() goja.Value {
			entity := EntityFromNode(skill)
			if entity == nil {
				return r.ToValue(0)
			}
			updateSkillLevelForScript(entity, skill)
			return r.ToValue(scriptLevel(skill.LevelData))
		}
		m["relativeLevel"] = func() goja.Value {
			entity := EntityFromNode(skill)
			if entity == nil {
				return r.ToValue(0)
			}
			updateSkillLevelForScript(entity, skill)
			return r.ToValue(scriptRelativeLevel(skill.LevelData))
		}
		m["weapons"] = func() goja.Value {
			weapons := make([]*goja.Object, 0, len(skill.Weapons))
			for _, w := range skill.Weapons {
				weapons = append(weapons, newScriptWeapon(r, w))
			}
			return r.ToValue(weapons)
		}
		m["findWeapons"] = func() goja.Value {
			return r.ToValue(func(call goja.FunctionCall) goja.Value {
				melee := call.Argument(0).ToBoolean()
				name := callArgAsString(call, 1)
				usage := callArgAsString(call, 2)
				return matchWeapons(r, skill.Weapons, name, usage, melee)
			})
		}
	}
	return r.NewDynamicObject(NewScriptObject(r, m))
}

// updateSkillLevelForScript recalculates the skill's level, guarding against a script that resolves a skill's level via
// itself. The guard is removed with a defer so that a panic inside UpdateLevel (recovered further up by xos.SafeCall)
// can't leave the exclusion registered on the entity, which would permanently freeze that skill's script-visible level.
func updateSkillLevelForScript(entity *Entity, skill *Skill) {
	name := skill.NameWithReplacements()
	specialization := skill.SpecializationWithReplacements()
	optionalSpecialization := skill.OptionalSpecializationWithReplacements()
	if entity.isSkillLevelResolutionExcluded(name, specialization, optionalSpecialization) {
		return
	}
	entity.registerSkillLevelResolutionExclusion(name, specialization, optionalSpecialization)
	defer entity.unregisterSkillLevelResolutionExclusion(name, specialization, optionalSpecialization)
	skill.UpdateLevel()
}

func findScriptSkills(r *goja.Runtime, name, specialization, tag string, topLevelSkills ...*Skill) goja.Value {
	var skills []*goja.Object
	Traverse(func(skill *Skill) bool {
		if (name == "" || strings.EqualFold(skill.NameWithReplacements(), name)) &&
			(specialization == "" || strings.EqualFold(skill.SpecializationWithReplacements(), specialization) ||
				strings.EqualFold(skill.OptionalSpecializationWithReplacements(), specialization)) &&
			matchTag(tag, skill.Tags) {
			skills = append(skills, newScriptSkill(r, skill))
		}
		return false
	}, true, false, topLevelSkills...)
	return r.ToValue(skills)
}
