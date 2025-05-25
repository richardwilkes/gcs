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

// AttributeIDs returns a list of available attribute IDs.
func (e *scriptEntity) AttributeIDs() []string {
	if e.entity == nil {
		return nil
	}
	attrs := e.entity.Attributes.List()
	ids := make([]string, 0, len(attrs))
	for _, attr := range attrs {
		if def := attr.AttributeDef(); def != nil {
			if def.IsSeparator() {
				continue
			}
			ids = append(ids, def.DefID)
		}
	}
	return ids
}

// CurrentAttributeValue resolves the given attribute ID to its current value.
func (e *scriptEntity) CurrentAttributeValue(attrID string) (value fxp.Int, exists bool) {
	if e.entity == nil {
		return 0, false
	}
	if value = e.entity.Attributes.Current(attrID); value == fxp.Min {
		return 0, false
	}
	return value, true
}
