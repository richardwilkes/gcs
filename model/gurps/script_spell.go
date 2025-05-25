package gurps

import (
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/json"
)

type scriptSpell struct {
	Name              string         `json:"name"`
	Kind              string         `json:"kind"`
	Attribute         string         `json:"attribute,omitempty"`
	Difficulty        string         `json:"difficulty,omitempty"`
	College           []string       `json:"college,omitempty"`
	PowerSource       string         `json:"power_source,omitempty"`
	Class             string         `json:"spell_class,omitempty"`
	Resist            string         `json:"resist,omitempty"`
	CastingCost       string         `json:"casting_cost,omitempty"`
	MaintenanceCost   string         `json:"maintenance_cost,omitempty"`
	CastingTime       string         `json:"casting_time,omitempty"`
	Duration          string         `json:"duration,omitempty"`
	RitualSkillName   string         `json:"base_skill,omitempty"`
	Children          []*scriptSpell `json:"children,omitempty"`
	Tags              []string       `json:"tags,omitempty"`
	Level             int            `json:"level"`
	RelativeLevel     int            `json:"relative_level"`
	RitualPrereqCount int            `json:"prereq_count,omitempty"`
}

func newScriptSpell(entity *Entity, spell *Spell, includeChildren bool) *scriptSpell {
	s := scriptSpell{
		Name: spell.NameWithReplacements(),
		Tags: slices.Clone(spell.Tags),
	}
	if spell.Container() {
		s.Kind = groupKind
		if includeChildren {
			children := spell.NodeChildren()
			s.Children = make([]*scriptSpell, 0, len(children))
			for _, child := range children {
				if child.Enabled() {
					s.Children = append(s.Children, newScriptSpell(entity, child, true))
				}
			}
		}
	} else {
		if spell.IsRitualMagic() {
			s.Kind = "ritual magic spell"
		} else {
			s.Kind = "spell"
		}
		s.Attribute = spell.Difficulty.Attribute
		s.Difficulty = spell.Difficulty.Difficulty.Key()
		s.College = slices.Clone([]string(spell.College))
		s.PowerSource = spell.PowerSource
		s.Class = spell.Class
		s.Resist = spell.Resist
		s.CastingCost = spell.CastingCost
		s.MaintenanceCost = spell.MaintenanceCost
		s.CastingTime = spell.CastingTime
		s.Duration = spell.Duration
		s.RitualSkillName = spell.RitualSkillName
		s.RitualPrereqCount = spell.RitualPrereqCount
		spell.UpdateLevel()
		s.Level = fxp.As[int](spell.LevelData.Level)
		s.RelativeLevel = fxp.As[int](spell.LevelData.RelativeLevel)
	}
	return &s
}

func (t *scriptSpell) String() string {
	data, err := json.Marshal(t)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
