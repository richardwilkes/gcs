package gurps

import (
	"strings"
)

type scriptEntity struct {
	entity *Entity
}

func (e *scriptEntity) Exists() bool {
	return e.entity != nil
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
			attrs = append(attrs, newScriptAttribute(attr))
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
		return newScriptAttribute(attr)
	}
	return nil
}

// Traits returns a hierarchical list of enabled traits.
func (e *scriptEntity) Traits() []*scriptTrait {
	if e.entity == nil {
		return nil
	}
	traits := make([]*scriptTrait, 0, len(e.entity.Traits))
	for _, trait := range e.entity.Traits {
		if trait.Enabled() {
			traits = append(traits, newScriptTrait(trait, true))
		}
	}
	return traits
}

// Trait returns all traits with the given name.
func (e *scriptEntity) Trait(name string, includeEnabledChildren bool) []*scriptTrait {
	if e.entity == nil {
		return nil
	}
	var result []*scriptTrait
	Traverse(func(t *Trait) bool {
		if strings.EqualFold(t.NameWithReplacements(), name) {
			result = append(result, newScriptTrait(t, includeEnabledChildren))
		}
		return false
	}, true, false, e.entity.Traits...)
	return result
}

// Skills returns a hierarchical list of skills.
func (e *scriptEntity) Skills() []*scriptSkill {
	if e.entity == nil {
		return nil
	}
	skills := make([]*scriptSkill, 0, len(e.entity.Skills))
	for _, skill := range e.entity.Skills {
		if skill.Enabled() {
			skills = append(skills, newScriptSkill(e.entity, skill, true))
		}
	}
	return skills
}

// Skill returns all skills with the given name.
func (e *scriptEntity) Skill(name, specialization string, includeChildren bool) []*scriptSkill {
	if e.entity == nil {
		return nil
	}
	var result []*scriptSkill
	Traverse(func(t *Skill) bool {
		if strings.EqualFold(t.NameWithReplacements(), name) &&
			strings.EqualFold(t.SpecializationWithReplacements(), specialization) {
			result = append(result, newScriptSkill(e.entity, t, includeChildren))
		}
		return false
	}, true, false, e.entity.Skills...)
	return result
}

// Spells returns a hierarchical list of spells.
func (e *scriptEntity) Spells() []*scriptSpell {
	if e.entity == nil {
		return nil
	}
	spells := make([]*scriptSpell, 0, len(e.entity.Spells))
	for _, spell := range e.entity.Spells {
		if spell.Enabled() {
			spells = append(spells, newScriptSpell(e.entity, spell, true))
		}
	}
	return spells
}

// Spell returns all spells with the given name.
func (e *scriptEntity) Spell(name string, includeChildren bool) []*scriptSpell {
	if e.entity == nil {
		return nil
	}
	var result []*scriptSpell
	Traverse(func(t *Spell) bool {
		if strings.EqualFold(t.NameWithReplacements(), name) {
			result = append(result, newScriptSpell(e.entity, t, includeChildren))
		}
		return false
	}, true, false, e.entity.Spells...)
	return result
}
