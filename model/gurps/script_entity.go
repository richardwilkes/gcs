package gurps

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

type scriptEntity struct {
	entity *Entity
}

func (e *scriptEntity) Exists() bool {
	return e.entity != nil
}

// HasTrait checks if the entity has a trait with the given name.
func (e *scriptEntity) HasTrait(traitName string) bool {
	if e.entity == nil {
		return false
	}
	found := false
	Traverse(func(t *Trait) bool {
		if strings.EqualFold(t.NameWithReplacements(), traitName) {
			found = true
			return true
		}
		return false
	}, true, false, e.entity.Traits...)
	return found
}

// TraitLevel returns the level of the trait with the given name, or -1 if not found or not leveled.
func (e *scriptEntity) TraitLevel(traitName string) fxp.Int {
	if e.entity == nil {
		return -fxp.One
	}
	levels := -fxp.One
	Traverse(func(t *Trait) bool {
		if strings.EqualFold(t.NameWithReplacements(), traitName) {
			if t.IsLeveled() {
				if levels == -fxp.One {
					levels = t.Levels
				} else {
					levels += t.Levels
				}
			}
		}
		return false
	}, true, false, e.entity.Traits...)
	return levels
}

// SkillLevel returns the level of the skill with the given name, or 0 if not found.
func (e *scriptEntity) SkillLevel(name, specialization string, relative bool) int {
	if e.entity == nil {
		return 0
	}
	if e.entity.isSkillLevelResolutionExcluded(name, specialization) {
		return 0
	}
	e.entity.registerSkillLevelResolutionExclusion(name, specialization)
	defer e.entity.unregisterSkillLevelResolutionExclusion(name, specialization)
	var level int
	Traverse(func(s *Skill) bool {
		if strings.EqualFold(s.NameWithReplacements(), name) &&
			strings.EqualFold(s.SpecializationWithReplacements(), specialization) {
			s.UpdateLevel()
			if relative {
				level = fxp.As[int](s.LevelData.RelativeLevel)
			} else {
				level = fxp.As[int](s.LevelData.Level)
			}
			return true
		}
		return false
	}, true, true, e.entity.Skills...)
	return level
}

// Attributes returns a list of available attributes.
func (e *scriptEntity) Attributes() []*scriptAttribute {
	if e.entity == nil {
		return nil
	}
	list := e.entity.Attributes.List()
	attrs := make([]*scriptAttribute, 0, len(list))
	for _, attr := range list {
		if def := attr.AttributeDef(); def != nil {
			if def.IsSeparator() {
				continue
			}
			attrs = append(attrs, &scriptAttribute{attr: attr})
		}
	}
	return attrs
}

// Attribute returns the attribute with the given ID or name, or nil if there is no match.
func (e *scriptEntity) Attribute(idOrName string) *scriptAttribute {
	if e.entity == nil {
		return nil
	}
	if attr := e.entity.Attributes.Find(idOrName); attr != nil {
		return &scriptAttribute{attr: attr}
	}
	return nil
}
