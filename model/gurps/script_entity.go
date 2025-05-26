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
	var traits []*scriptTrait
	Traverse(func(t *Trait) bool {
		if strings.EqualFold(t.NameWithReplacements(), name) {
			traits = append(traits, newScriptTrait(t, includeEnabledChildren))
		}
		return false
	}, true, false, e.entity.Traits...)
	return traits
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
	var skills []*scriptSkill
	Traverse(func(t *Skill) bool {
		if strings.EqualFold(t.NameWithReplacements(), name) &&
			strings.EqualFold(t.SpecializationWithReplacements(), specialization) {
			skills = append(skills, newScriptSkill(e.entity, t, includeChildren))
		}
		return false
	}, true, false, e.entity.Skills...)
	return skills
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
	var spells []*scriptSpell
	Traverse(func(t *Spell) bool {
		if strings.EqualFold(t.NameWithReplacements(), name) {
			spells = append(spells, newScriptSpell(e.entity, t, includeChildren))
		}
		return false
	}, true, false, e.entity.Spells...)
	return spells
}

// Items returns a hierarchical list of carried equipment.
func (e *scriptEntity) Items() []*scriptEquipment {
	if e.entity == nil {
		return nil
	}
	items := make([]*scriptEquipment, 0, len(e.entity.CarriedEquipment))
	for _, item := range e.entity.CarriedEquipment {
		items = append(items, newScriptEquipment(e.entity, item, true))
	}
	return items
}

// Item returns all carried equipment with the given name.
func (e *scriptEntity) Item(name string, includeChildren bool) []*scriptEquipment {
	if e.entity == nil {
		return nil
	}
	var items []*scriptEquipment
	Traverse(func(t *Equipment) bool {
		if strings.EqualFold(t.NameWithReplacements(), name) {
			items = append(items, newScriptEquipment(e.entity, t, includeChildren))
		}
		return false
	}, true, false, e.entity.CarriedEquipment...)
	return items
}

func (e *scriptEntity) Encumbrance() scriptEncumbrance {
	return newScriptEncumbrance(e.entity)
}

func (e *scriptEntity) WeightUnits() fxp.WeightUnit {
	return SheetSettingsFor(e.entity).DefaultWeightUnits
}

func (e *scriptEntity) SizeModifier() int {
	if e.entity == nil {
		return 0
	}
	return e.entity.Profile.AdjustedSizeModifier()
}
