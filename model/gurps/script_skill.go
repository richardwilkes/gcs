package gurps

import (
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

const groupKind = "group"

type scriptSkill struct {
	Name           string         `json:"name"`
	Specialization string         `json:"specialization,omitempty"`
	Kind           string         `json:"kind"`
	Attribute      string         `json:"attribute,omitempty"`
	Difficulty     string         `json:"difficulty,omitempty"`
	Children       []*scriptSkill `json:"children,omitempty"`
	Tags           []string       `json:"tags,omitempty"`
	Level          int            `json:"level"`
	RelativeLevel  int            `json:"relative_level"`
}

func deferredNewScriptSkill(entity *Entity, skill *Skill) func() any {
	if skill == nil {
		return nil
	}
	return func() any {
		return newScriptSkill(entity, skill, true)
	}
}

func newScriptSkill(entity *Entity, skill *Skill, includeChildren bool) *scriptSkill {
	s := scriptSkill{
		Name:           skill.NameWithReplacements(),
		Specialization: skill.SpecializationWithReplacements(),
		Tags:           slices.Clone(skill.Tags),
	}
	if skill.Container() {
		s.Kind = groupKind
		if includeChildren {
			children := skill.NodeChildren()
			s.Children = make([]*scriptSkill, 0, len(children))
			for _, child := range children {
				if child.Enabled() {
					s.Children = append(s.Children, newScriptSkill(entity, child, true))
				}
			}
		}
	} else {
		if skill.IsTechnique() {
			s.Kind = "technique"
		} else {
			s.Kind = "skill"
		}
		s.Attribute = skill.Difficulty.Attribute
		s.Difficulty = skill.Difficulty.Difficulty.Key()
		if !entity.isSkillLevelResolutionExcluded(s.Name, s.Specialization) {
			entity.registerSkillLevelResolutionExclusion(s.Name, s.Specialization)
			skill.UpdateLevel()
			entity.unregisterSkillLevelResolutionExclusion(s.Name, s.Specialization)
			s.Level = fxp.As[int](skill.LevelData.Level)
			s.RelativeLevel = fxp.As[int](skill.LevelData.RelativeLevel)
		}
	}
	return &s
}
