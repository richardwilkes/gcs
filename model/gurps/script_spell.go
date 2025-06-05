package gurps

import (
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/tid"
)

type scriptSpell struct {
	entity            *Entity
	spell             *Spell
	ID                tid.TID
	ParentID          tid.TID
	Name              string
	Kind              string
	TechLevel         string
	Attribute         string
	Difficulty        string
	College           []string
	PowerSource       string
	SpellClass        string
	Resist            string
	CastingCost       string
	MaintenanceCost   string
	CastingTime       string
	Duration          string
	RitualSkillName   string
	children          []*scriptSpell
	Tags              []string
	RitualPrereqCount int
	Points            int
	level             int
	relativeLevel     int
	Container         bool
	HasChildren       bool
	cachedChildren    bool
	cachedLevels      bool
}

func deferredNewScriptSpell(entity *Entity, spell *Spell) ScriptSelfProvider {
	if spell == nil {
		return ScriptSelfProvider{}
	}
	return ScriptSelfProvider{
		ID:       spell.TID,
		Provider: func() any { return newScriptSpell(entity, spell) },
	}
}

func newScriptSpell(entity *Entity, spell *Spell) *scriptSpell {
	var parentID tid.TID
	if spell.parent != nil {
		parentID = spell.parent.TID
	}
	var tl string
	if spell.TechLevel != nil {
		tl = *spell.TechLevel
	}
	s := scriptSpell{
		entity:      entity,
		spell:       spell,
		ID:          spell.TID,
		ParentID:    parentID,
		Name:        spell.NameWithReplacements(),
		Tags:        slices.Clone(spell.Tags),
		TechLevel:   tl,
		Container:   spell.Container(),
		HasChildren: spell.HasChildren(),
	}
	if !spell.Container() {
		if spell.IsRitualMagic() {
			s.Kind = "ritual magic spell"
		} else {
			s.Kind = "spell"
		}
		s.Attribute = spell.Difficulty.Attribute
		s.Difficulty = spell.Difficulty.Difficulty.Key()
		s.College = slices.Clone([]string(spell.College))
		s.PowerSource = spell.PowerSource
		s.SpellClass = spell.Class
		s.Resist = spell.Resist
		s.CastingCost = spell.CastingCost
		s.MaintenanceCost = spell.MaintenanceCost
		s.CastingTime = spell.CastingTime
		s.Duration = spell.Duration
		s.RitualSkillName = spell.RitualSkillName
		s.RitualPrereqCount = spell.RitualPrereqCount
		s.Points = fxp.As[int](spell.AdjustedPoints(nil))
	}
	return &s
}

func (s *scriptSpell) Level() int {
	s.ensureCachedLevels()
	return s.level
}

func (s *scriptSpell) RelativeLevel() int {
	s.ensureCachedLevels()
	return s.relativeLevel
}

func (s *scriptSpell) ensureCachedLevels() {
	if s.cachedLevels {
		return
	}
	s.spell.UpdateLevel()
	s.level = fxp.As[int](s.spell.LevelData.Level)
	s.relativeLevel = fxp.As[int](s.spell.LevelData.RelativeLevel)
}

func (s *scriptSpell) Children() []*scriptSpell {
	if !s.cachedChildren {
		s.cachedChildren = true
		if len(s.spell.Children) != 0 {
			s.children = make([]*scriptSpell, 0, len(s.spell.Children))
			for _, child := range s.spell.Children {
				if child.Enabled() {
					s.children = append(s.children, newScriptSpell(s.entity, child))
				}
			}
		}
	}
	return s.children
}

func (s *scriptSpell) Find(name, tag string) []*scriptSpell {
	if !s.spell.Container() {
		return nil
	}
	return findScriptSpells(s.entity, name, tag, s.spell.Children...)
}

func findScriptSpells(entity *Entity, name, tag string, topLevelSpells ...*Spell) []*scriptSpell {
	var spells []*scriptSpell
	Traverse(func(spell *Spell) bool {
		if (name == "" || strings.EqualFold(spell.NameWithReplacements(), name)) && matchTag(tag, spell.Tags) {
			spells = append(spells, newScriptSpell(entity, spell))
		}
		return false
	}, true, false, topLevelSpells...)
	return spells
}
