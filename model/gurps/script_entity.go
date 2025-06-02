package gurps

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/xmath/rand"
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
	Traverse(func(trait *Trait) bool {
		if strings.EqualFold(trait.NameWithReplacements(), name) && trait.Enabled() {
			traits = append(traits, newScriptTrait(trait, includeEnabledChildren))
		}
		return false
	}, true, false, e.entity.Traits...)
	return traits
}

func (e *scriptEntity) HasTrait(name string) bool {
	if e.entity == nil {
		return false
	}
	found := false
	name = strings.TrimSpace(name)
	Traverse(func(t *Trait) bool {
		if strings.EqualFold(t.NameWithReplacements(), name) {
			found = true
			return true
		}
		return false
	}, true, false, e.entity.Traits...)
	return found
}

func (e *scriptEntity) TraitLevel(name string) float64 {
	if e.entity == nil {
		return -1
	}
	name = strings.TrimSpace(name)
	level := -fxp.One
	Traverse(func(t *Trait) bool {
		if strings.EqualFold(t.NameWithReplacements(), name) {
			if t.IsLeveled() {
				if level == -fxp.One {
					level = t.Levels
				} else {
					level += t.Levels
				}
			}
		}
		return false
	}, true, true, e.entity.Traits...)
	return fxp.As[float64](level)
}

// Skills returns a hierarchical list of skills.
func (e *scriptEntity) Skills() []*scriptSkill {
	if e.entity == nil {
		return nil
	}
	skills := make([]*scriptSkill, 0, len(e.entity.Skills))
	for _, skill := range e.entity.Skills {
		skills = append(skills, newScriptSkill(e.entity, skill, true))
	}
	return skills
}

// Skill returns all skills with the given name.
func (e *scriptEntity) Skill(name, specialization string, includeChildren bool) []*scriptSkill {
	if e.entity == nil {
		return nil
	}
	var skills []*scriptSkill
	Traverse(func(skill *Skill) bool {
		if strings.EqualFold(skill.NameWithReplacements(), name) &&
			strings.EqualFold(skill.SpecializationWithReplacements(), specialization) {
			skills = append(skills, newScriptSkill(e.entity, skill, includeChildren))
		}
		return false
	}, true, false, e.entity.Skills...)
	return skills
}

func (e *scriptEntity) SkillLevel(name, specialization string, relative bool) int {
	if e.entity == nil {
		return 0
	}
	name = strings.TrimSpace(name)
	specialization = strings.TrimSpace(specialization)
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

// Spells returns a hierarchical list of spells.
func (e *scriptEntity) Spells() []*scriptSpell {
	if e.entity == nil {
		return nil
	}
	spells := make([]*scriptSpell, 0, len(e.entity.Spells))
	for _, spell := range e.entity.Spells {
		spells = append(spells, newScriptSpell(e.entity, spell, true))
	}
	return spells
}

// Spell returns all spells with the given name.
func (e *scriptEntity) Spell(name string, includeChildren bool) []*scriptSpell {
	if e.entity == nil {
		return nil
	}
	var spells []*scriptSpell
	Traverse(func(spell *Spell) bool {
		if strings.EqualFold(spell.NameWithReplacements(), name) {
			spells = append(spells, newScriptSpell(e.entity, spell, includeChildren))
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
		if item.Quantity > 0 {
			items = append(items, newScriptEquipment(e.entity, item, true))
		}
	}
	return items
}

// Item returns all carried equipment with the given name.
func (e *scriptEntity) Item(name string, includeChildren bool) []*scriptEquipment {
	if e.entity == nil {
		return nil
	}
	var items []*scriptEquipment
	Traverse(func(item *Equipment) bool {
		if item.Quantity > 0 && strings.EqualFold(item.NameWithReplacements(), name) {
			items = append(items, newScriptEquipment(e.entity, item, includeChildren))
		}
		return false
	}, true, false, e.entity.CarriedEquipment...)
	return items
}

func (e *scriptEntity) Encumbrance() scriptEncumbrance {
	return newScriptEncumbrance(e.entity)
}

func (e *scriptEntity) CurrentEncumbrance(forSkills, returnMoveFactor bool) float64 {
	if e.entity == nil {
		return 0
	}
	level := int(e.entity.EncumbranceLevel(forSkills))
	if returnMoveFactor {
		return fxp.As[float64](fxp.One - fxp.From(level).Mul(fxp.Two).Div(fxp.Ten))
	}
	return float64(level)
}

func (e *scriptEntity) WeightUnits() fxp.WeightUnit {
	return SheetSettingsFor(e.entity).DefaultWeightUnits
}

func (e *scriptEntity) ExtraDiceFromModifiers() bool {
	return SheetSettingsFor(e.entity).UseModifyingDicePlusAdds
}

func (e *scriptEntity) SizeModifier() int {
	if e.entity == nil {
		return 0
	}
	return e.entity.Profile.AdjustedSizeModifier()
}

func (e *scriptEntity) WeaponDamage(name, usage string) string {
	if e.entity == nil {
		return ""
	}
	for _, w := range e.entity.Weapons(true, false) {
		if strings.EqualFold(w.String(), name) && strings.EqualFold(w.UsageWithReplacements(), usage) {
			return w.Damage.ResolvedDamage(nil)
		}
	}
	for _, w := range e.entity.Weapons(false, false) {
		if strings.EqualFold(w.String(), name) && strings.EqualFold(w.UsageWithReplacements(), usage) {
			return w.Damage.ResolvedDamage(nil)
		}
	}
	return ""
}

// RandomHeight returns a height in inches based on the given strength using the chart from B18.
func (e *scriptEntity) RandomHeight(st int) int {
	r := rand.NewCryptoRand()
	return 68 + (st-10)*2 + (r.Intn(6) + 1) - (r.Intn(6) + 1)
}

// RandomWeight returns a weight in pounds based on the given strength using the chart from B18. Adjusts appropriately
// for the traits Skinny, Overweight, Fat, and Very Fat, if present on the sheet. 'shift' causes a shift towards a
// lighter value if negative and a heavier value if positive, similar to having one of the traits Skinny, Overweight,
// Fat, and Very Fat applied, but is additive to them.
func (e *scriptEntity) RandomWeight(st, shift int) int {
	shift += 3 // Average
	if e.entity != nil {
		skinny := false
		overweight := false
		fat := false
		veryFat := false
		Traverse(func(t *Trait) bool {
			switch strings.ToLower(t.NameWithReplacements()) {
			case "skinny":
				skinny = true
			case "overweight":
				overweight = true
			case "fat":
				fat = true
			case "very Fat":
				veryFat = true
			}
			return false
		}, true, false, e.entity.Traits...)
		switch {
		case skinny:
			shift--
		case veryFat:
			shift += 3
		case fat:
			shift += 2
		case overweight:
			shift++
		}
	}
	r := rand.NewCryptoRand()
	mid := 145 + (st-10)*15
	deviation := mid/5 + 2
	return ((mid + r.Intn(deviation) - r.Intn(deviation)) * shift) / 3
}
