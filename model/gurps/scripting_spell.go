// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"github.com/dop251/goja"
)

func deferredNewScriptSpell(spell *Spell) ScriptSelfProvider {
	return deferredScriptSelf(spell, newScriptSpell)
}

func newScriptSpell(r *goja.Runtime, spell *Spell) *goja.Object {
	m := make(map[string]func() goja.Value)
	addScriptNodeIdentity(r, m, spell, spell.Tags, newScriptSpell)
	m["name"] = func() goja.Value { return r.ToValue(spell.NameWithReplacements()) }
	m["notes"] = scriptNotes(r, spell)
	if spell.Container() {
		m["children"] = func() goja.Value { return scriptObjects(r, spell.Children, nil, newScriptSpell) }
		m["find"] = scriptNameTagFinder(r, func(name, tag string) goja.Value {
			return findScriptSpells(r, name, tag, spell.Children...)
		})
	} else {
		m["switchedOn"] = func() goja.Value { return r.ToValue(spell.SwitchedOn) }
		m["techLevel"] = func() goja.Value {
			if spell.TechLevel != nil {
				return r.ToValue(*spell.TechLevel)
			}
			return r.ToValue("")
		}
		m["kind"] = func() goja.Value {
			if spell.IsRitualMagic() {
				return r.ToValue("ritual magic spell")
			}
			return r.ToValue("spell")
		}
		m["attribute"] = func() goja.Value { return r.ToValue(spell.Difficulty.Attribute) }
		m["difficulty"] = func() goja.Value { return r.ToValue(spell.Difficulty.Difficulty.Key()) }
		m["points"] = func() goja.Value { return r.ToValue(spell.AdjustedPoints(nil).AsInteger[int]()) }
		m["college"] = func() goja.Value { return r.ToValue(spell.CollegeWithReplacements()) }
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
			return r.ToValue(scriptLevel(spell.LevelData))
		}
		m["relativeLevel"] = func() goja.Value {
			spell.UpdateLevel()
			return r.ToValue(scriptRelativeLevel(spell.LevelData))
		}
		addScriptWeapons(r, m, func() []*Weapon { return spell.Weapons })
	}
	return r.NewDynamicObject(NewScriptObject(r, m))
}

func findScriptSpells(r *goja.Runtime, name, tag string, topLevelSpells ...*Spell) goja.Value {
	return findScriptNodes(r, name, tag, scriptNodeKind[*Spell]{
		ctor:   newScriptSpell,
		nameOf: (*Spell).NameWithReplacements,
		tagsOf: func(spell *Spell) []string { return spell.Tags },
	}, nil, topLevelSpells...)
}
