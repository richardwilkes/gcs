package gurps

import (
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/tid"
)

type scriptSkill struct {
	entity         *Entity
	skill          *Skill
	ID             tid.TID
	ParentID       tid.TID
	Name           string
	Specialization string
	Kind           string
	TechLevel      string
	Attribute      string
	Difficulty     string
	children       []*scriptSkill
	Tags           []string
	Points         int
	level          int
	relativeLevel  int
	Container      bool
	HasChildren    bool
	cachedChildren bool
	cachedLevels   bool
}

func deferredNewScriptSkill(entity *Entity, skill *Skill) ScriptSelfProvider {
	if skill == nil {
		return ScriptSelfProvider{}
	}
	return ScriptSelfProvider{
		ID:       skill.TID,
		Provider: func() any { return newScriptSkill(entity, skill) },
	}
}

func newScriptSkill(entity *Entity, skill *Skill) *scriptSkill {
	var parentID tid.TID
	if skill.parent != nil {
		parentID = skill.parent.TID
	}
	var tl string
	if skill.TechLevel != nil {
		tl = *skill.TechLevel
	}
	s := scriptSkill{
		entity:         entity,
		skill:          skill,
		ID:             skill.TID,
		ParentID:       parentID,
		Name:           skill.NameWithReplacements(),
		Specialization: skill.SpecializationWithReplacements(),
		Tags:           slices.Clone(skill.Tags),
		TechLevel:      tl,
		Container:      skill.Container(),
		HasChildren:    skill.HasChildren(),
	}
	if !skill.Container() {
		if skill.IsTechnique() {
			s.Kind = "technique"
		} else {
			s.Kind = "skill"
		}
		s.Attribute = skill.Difficulty.Attribute
		s.Difficulty = skill.Difficulty.Difficulty.Key()
		s.Points = fxp.As[int](skill.AdjustedPoints(nil))
	}
	return &s
}

func (s *scriptSkill) Level() int {
	s.ensureCachedLevels()
	return s.level
}

func (s *scriptSkill) RelativeLevel() int {
	s.ensureCachedLevels()
	return s.relativeLevel
}

func (s *scriptSkill) ensureCachedLevels() {
	if s.cachedLevels {
		return
	}
	if !s.entity.isSkillLevelResolutionExcluded(s.Name, s.Specialization) {
		s.entity.registerSkillLevelResolutionExclusion(s.Name, s.Specialization)
		s.skill.UpdateLevel()
		s.entity.unregisterSkillLevelResolutionExclusion(s.Name, s.Specialization)
		s.level = fxp.As[int](s.skill.LevelData.Level)
		s.relativeLevel = fxp.As[int](s.skill.LevelData.RelativeLevel)
	}
}

func (s *scriptSkill) Children() []*scriptSkill {
	if !s.cachedChildren {
		s.cachedChildren = true
		if len(s.skill.Children) != 0 {
			s.children = make([]*scriptSkill, 0, len(s.skill.Children))
			for _, child := range s.skill.Children {
				if child.Enabled() {
					s.children = append(s.children, newScriptSkill(s.entity, child))
				}
			}
		}
	}
	return s.children
}

func (s *scriptSkill) Find(name, specialization, tag string) []*scriptSkill {
	if !s.skill.Container() {
		return nil
	}
	return findScriptSkills(s.entity, name, specialization, tag, s.skill.Children...)
}

func findScriptSkills(entity *Entity, name, specialization, tag string, topLevelSkills ...*Skill) []*scriptSkill {
	var skills []*scriptSkill
	Traverse(func(skill *Skill) bool {
		if (name == "" || strings.EqualFold(skill.NameWithReplacements(), name)) &&
			(specialization == "" || strings.EqualFold(skill.SpecializationWithReplacements(), specialization)) &&
			matchTag(tag, skill.Tags) {
			skills = append(skills, newScriptSkill(entity, skill))
		}
		return false
	}, true, false, topLevelSkills...)
	return skills
}
