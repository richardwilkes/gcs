// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"slices"
	"strings"

	"github.com/dop251/goja"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
)

func deferredNewScriptSpell(spell *Spell) ScriptSelfProvider {
	if spell == nil {
		return ScriptSelfProvider{}
	}
	return ScriptSelfProvider{
		ID:       string(spell.TID),
		Provider: func(r *goja.Runtime) any { return newScriptSpell(r, spell) },
	}
}

func newScriptSpell(r *goja.Runtime, spell *Spell) *goja.Object {
	m := make(map[string]func() goja.Value)
	m["id"] = func() goja.Value { return r.ToValue(spell.TID) }
	m["parentID"] = func() goja.Value {
		if spell.parent == nil {
			return goja.Undefined()
		}
		return r.ToValue(spell.parent.TID)
	}
	m["parent"] = func() goja.Value {
		if spell.parent == nil {
			return goja.Undefined()
		}
		return newScriptSpell(r, spell.parent)
	}
	m["name"] = func() goja.Value { return r.ToValue(spell.NameWithReplacements()) }
	m["notes"] = func() goja.Value {
		return r.ToValue(spell.SecondaryText(func(_ display.Option) bool { return true }))
	}
	m["tags"] = func() goja.Value { return r.ToValue(slices.Clone(spell.Tags)) }
	m["container"] = func() goja.Value { return r.ToValue(spell.Container()) }
	if spell.Container() {
		m["children"] = func() goja.Value {
			children := make([]*goja.Object, 0, len(spell.Children))
			for _, child := range spell.Children {
				children = append(children, newScriptSpell(r, child))
			}
			return r.ToValue(children)
		}
		m["find"] = func() goja.Value {
			return r.ToValue(func(call goja.FunctionCall) goja.Value {
				name := callArgAsString(call, 0)
				tag := callArgAsString(call, 1)
				return findScriptSpells(r, name, tag, spell.Children...)
			})
		}
	} else {
		m["techLevel"] = func() goja.Value {
			if spell.TechLevel != nil {
				return r.ToValue(*spell.TechLevel)
			}
			return goja.Undefined()
		}
		m["kind"] = func() goja.Value {
			if spell.IsRitualMagic() {
				return r.ToValue("ritual magic spell")
			}
			return r.ToValue("spell")
		}
		m["attribute"] = func() goja.Value { return r.ToValue(spell.Difficulty.Attribute) }
		m["difficulty"] = func() goja.Value { return r.ToValue(spell.Difficulty.Difficulty.String()) }
		m["points"] = func() goja.Value { return r.ToValue(fxp.AsInteger[int](spell.AdjustedPoints(nil))) }
		m["college"] = func() goja.Value { return r.ToValue(slices.Clone([]string(spell.College))) }
		m["powerSource"] = func() goja.Value { return r.ToValue(spell.PowerSourceWithReplacements()) }
		m["spellClass"] = func() goja.Value { return r.ToValue(spell.ClassWithReplacements()) }
		m["resist"] = func() goja.Value { return r.ToValue(spell.ResistWithReplacements()) }
		m["castingCost"] = func() goja.Value { return r.ToValue(spell.CastingCostWithReplacements()) }
		m["maintenanceCost"] = func() goja.Value { return r.ToValue(spell.MaintenanceCostWithReplacements()) }
		m["castingTime"] = func() goja.Value { return r.ToValue(spell.CastingTimeWithReplacements()) }
		m["duration"] = func() goja.Value { return r.ToValue(spell.DurationWithReplacements()) }
		m["item"] = func() goja.Value { return r.ToValue(spell.ItemWithReplacements()) }
		m["ritualSkillName"] = func() goja.Value { return r.ToValue(spell.RitualSkillNameWithReplacements()) }
		m["prereqCount"] = func() goja.Value { return r.ToValue(spell.PrereqCount) }
		m["level"] = func() goja.Value {
			spell.UpdateLevel()
			return r.ToValue(fxp.AsInteger[int](spell.LevelData.Level))
		}
		m["relativeLevel"] = func() goja.Value {
			spell.UpdateLevel()
			return r.ToValue(fxp.AsInteger[int](spell.LevelData.RelativeLevel))
		}
	}
	return r.NewDynamicObject(NewScriptObject(r, m))
}

func findScriptSpells(r *goja.Runtime, name, tag string, topLevelSpells ...*Spell) goja.Value {
	var spells []*goja.Object
	Traverse(func(spell *Spell) bool {
		if (name == "" || strings.EqualFold(spell.NameWithReplacements(), name)) && matchTag(tag, spell.Tags) {
			spells = append(spells, newScriptSpell(r, spell))
		}
		return false
	}, true, false, topLevelSpells...)
	return r.ToValue(spells)
}
