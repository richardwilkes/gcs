package gurps

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/xmath/rand"
)

type scriptEntity struct {
	entity                 *Entity
	PlayerName             string
	Name                   string
	Title                  string
	Organization           string
	Religion               string
	TechLevel              string
	Gender                 string
	Age                    string
	Birthday               string
	Eyes                   string
	Hair                   string
	Skin                   string
	Handedness             string
	DisplayHeightUnits     string
	DisplayWeightUnits     string
	HeightInInches         float64
	WeightInPounds         float64
	SizeModifier           int
	ExtraDiceFromModifiers bool
	Exists                 bool
}

func newScriptEntity(entity *Entity) *scriptEntity {
	settings := SheetSettingsFor(entity)
	e := &scriptEntity{
		DisplayHeightUnits:     settings.DefaultLengthUnits.Key(),
		DisplayWeightUnits:     settings.DefaultWeightUnits.Key(),
		ExtraDiceFromModifiers: settings.UseModifyingDicePlusAdds,
	}
	if entity != nil {
		e.entity = entity
		e.PlayerName = entity.Profile.PlayerName
		e.Name = entity.Profile.Name
		e.Title = entity.Profile.Title
		e.Organization = entity.Profile.Organization
		e.Religion = entity.Profile.Religion
		e.TechLevel = entity.Profile.TechLevel
		e.Gender = entity.Profile.Gender
		e.Age = entity.Profile.Age
		e.Birthday = entity.Profile.Birthday
		e.Eyes = entity.Profile.Eyes
		e.Hair = entity.Profile.Hair
		e.Skin = entity.Profile.Skin
		e.Handedness = entity.Profile.Handedness
		e.HeightInInches = fxp.As[float64](fxp.Int(entity.Profile.Height))
		e.WeightInPounds = fxp.As[float64](fxp.Int(entity.Profile.Weight))
		e.SizeModifier = entity.Profile.AdjustedSizeModifier()
		e.Exists = true
	}
	return e
}

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

func (e *scriptEntity) Attribute(idOrName string) *scriptAttribute {
	if e.entity == nil {
		return nil
	}
	if attr := e.entity.Attributes.Find(idOrName); attr != nil {
		return newScriptAttribute(attr)
	}
	return nil
}

func (e *scriptEntity) Traits() []*scriptTrait {
	if e.entity == nil {
		return nil
	}
	traits := make([]*scriptTrait, 0, len(e.entity.Traits))
	for _, trait := range e.entity.Traits {
		if trait.Enabled() {
			traits = append(traits, newScriptTrait(trait))
		}
	}
	return traits
}

func (e *scriptEntity) FindTraits(name, tag string) []*scriptTrait {
	if e.entity == nil {
		return nil
	}
	return findScriptTraits(name, tag, e.entity.Traits...)
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

func (e *scriptEntity) Skills() []*scriptSkill {
	if e.entity == nil {
		return nil
	}
	skills := make([]*scriptSkill, 0, len(e.entity.Skills))
	for _, skill := range e.entity.Skills {
		if skill.Enabled() {
			skills = append(skills, newScriptSkill(e.entity, skill))
		}
	}
	return skills
}

func (e *scriptEntity) FindSkills(name, specialization, tag string) []*scriptSkill {
	if e.entity == nil {
		return nil
	}
	return findScriptSkills(e.entity, name, specialization, tag, e.entity.Skills...)
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

func (e *scriptEntity) Spells() []*scriptSpell {
	if e.entity == nil {
		return nil
	}
	spells := make([]*scriptSpell, 0, len(e.entity.Spells))
	for _, spell := range e.entity.Spells {
		if spell.Enabled() {
			spells = append(spells, newScriptSpell(e.entity, spell))
		}
	}
	return spells
}

func (e *scriptEntity) FindSpells(name, tag string) []*scriptSpell {
	if e.entity == nil {
		return nil
	}
	return findScriptSpells(e.entity, name, tag, e.entity.Spells...)
}

func (e *scriptEntity) Equipment() []*scriptEquipment {
	if e.entity == nil {
		return nil
	}
	items := make([]*scriptEquipment, 0, len(e.entity.CarriedEquipment))
	for _, item := range e.entity.CarriedEquipment {
		if item.Quantity > 0 {
			items = append(items, newScriptEquipment(item))
		}
	}
	return items
}

func (e *scriptEntity) FindEquipment(name, tag string) []*scriptEquipment {
	if e.entity == nil {
		return nil
	}
	return findScriptEquipment(name, tag, e.entity.CarriedEquipment...)
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

// RandomHeightInInches returns a height in inches based on the given strength using the chart from B18.
func (e *scriptEntity) RandomHeightInInches(st int) int {
	r := rand.NewCryptoRand()
	return 68 + (st-10)*2 + (r.Intn(6) + 1) - (r.Intn(6) + 1)
}

// RandomWeightInPounds returns a weight in pounds based on the given strength using the chart from B18. Adjusts
// appropriately for the traits Skinny, Overweight, Fat, and Very Fat, if present on the sheet. 'shift' causes a shift
// towards a lighter value if negative and a heavier value if positive, similar to having one of the traits Skinny,
// Overweight, Fat, and Very Fat applied, but is additive to them.
func (e *scriptEntity) RandomWeightInPounds(st, shift int) int {
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
