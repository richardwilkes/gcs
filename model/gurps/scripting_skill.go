// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
		ID:       skill.TID,
		Provider: func(r *goja.Runtime) any { return newScriptSkill(r, skill) },
	}
}

func newScriptSkill(r *goja.Runtime, skill *Skill) *goja.Object {
	m := make(map[string]func() goja.Value)
	m["id"] = func() goja.Value { return r.ToValue(skill.TID) }
	m["parentID"] = func() goja.Value {
		if skill.parent != nil {
			return r.ToValue(skill.parent.TID)
		}
		return r.ToValue("")
	}
	m["name"] = func() goja.Value { return r.ToValue(skill.NameWithReplacements()) }
	m["specialization"] = func() goja.Value { return r.ToValue(skill.SpecializationWithReplacements()) }
	m["notes"] = func() goja.Value {
		return r.ToValue(skill.SecondaryText(func(_ display.Option) bool { return true }))
	}
	m["kind"] = func() goja.Value {
		if skill.Container() {
			return r.ToValue("")
		}
		if skill.IsTechnique() {
			return r.ToValue("technique")
		}
		return r.ToValue("skill")
	}
	m["techLevel"] = func() goja.Value {
		if skill.TechLevel != nil {
			return r.ToValue(*skill.TechLevel)
		}
		return r.ToValue("")
	}
	m["tags"] = func() goja.Value { return r.ToValue(slices.Clone(skill.Tags)) }
	m["container"] = func() goja.Value { return r.ToValue(skill.Container()) }
	m["hasChildren"] = func() goja.Value { return r.ToValue(skill.HasChildren()) }
	if skill.Container() {
		m["children"] = func() goja.Value {
			return r.ToValue(func(_ goja.FunctionCall) goja.Value {
				children := make([]*goja.Object, 0, len(skill.Children))
				for _, child := range skill.Children {
					if child.Enabled() {
						children = append(children, newScriptSkill(r, child))
					}
				}
				return r.ToValue(children)
			})
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
		m["attribute"] = func() goja.Value { return r.ToValue(skill.Difficulty.Attribute) }
		m["difficulty"] = func() goja.Value { return r.ToValue(skill.Difficulty.Key()) }
		m["points"] = func() goja.Value { return r.ToValue(fxp.AsInteger[int](skill.AdjustedPoints(nil))) }
		m["level"] = func() goja.Value {
			return r.ToValue(func(_ goja.FunctionCall) goja.Value {
				entity := EntityFromNode(skill)
				if entity == nil {
					return r.ToValue(0)
				}
				if !entity.isSkillLevelResolutionExcluded(skill.Name, skill.Specialization) {
					entity.registerSkillLevelResolutionExclusion(skill.Name, skill.Specialization)
					skill.UpdateLevel()
					entity.unregisterSkillLevelResolutionExclusion(skill.Name, skill.Specialization)
				}
				return r.ToValue(fxp.AsInteger[int](skill.LevelData.Level))
			})
		}
		m["relativeLevel"] = func() goja.Value {
			return r.ToValue(func(_ goja.FunctionCall) goja.Value {
				entity := EntityFromNode(skill)
				if entity == nil {
					return r.ToValue(0)
				}
				if !entity.isSkillLevelResolutionExcluded(skill.Name, skill.Specialization) {
					entity.registerSkillLevelResolutionExclusion(skill.Name, skill.Specialization)
					skill.UpdateLevel()
					entity.unregisterSkillLevelResolutionExclusion(skill.Name, skill.Specialization)
				}
				return r.ToValue(fxp.AsInteger[int](skill.LevelData.RelativeLevel))
			})
		}
	}
	return r.NewDynamicObject(NewScriptObject(r, m))
}

func findScriptSkills(r *goja.Runtime, name, specialization, tag string, topLevelSkills ...*Skill) goja.Value {
	var skills []*goja.Object
	Traverse(func(skill *Skill) bool {
		if (name == "" || strings.EqualFold(skill.NameWithReplacements(), name)) &&
			(specialization == "" || strings.EqualFold(skill.SpecializationWithReplacements(), specialization)) &&
			matchTag(tag, skill.Tags) {
			skills = append(skills, newScriptSkill(r, skill))
		}
		return false
	}, true, false, topLevelSkills...)
	return r.ToValue(skills)
}
