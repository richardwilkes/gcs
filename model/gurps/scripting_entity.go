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
	"strings"

	"github.com/dop251/goja"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/xrand"
)

func newScriptEntity(r *goja.Runtime, entity *Entity) *goja.Object {
	m := make(map[string]func() goja.Value)
	m["exists"] = func() goja.Value { return r.ToValue(entity != nil) }
	m["randomHeightInInches"] = func() goja.Value {
		// Returns a height in inches based on the given strength using the chart from B18.
		return r.ToValue(func(call goja.FunctionCall) goja.Value {
			st := int(call.Argument(0).ToInteger())
			rnd := xrand.New()
			return r.ToValue(68 + (st-10)*2 + (rnd.Intn(6) + 1) - (rnd.Intn(6) + 1))
		})
	}
	m["randomWeightInPounds"] = func() goja.Value {
		// Returns a weight in pounds based on the given strength using the chart from B18. Adjusts appropriately
		// for the traits Skinny, Overweight, Fat, and Very Fat, if present on the sheet. 'shift' causes a shift
		// towards a lighter value if negative and a heavier value if positive, similar to having one of the traits
		// Skinny, Overweight, Fat, and Very Fat applied, but is additive to them.
		return r.ToValue(func(call goja.FunctionCall) goja.Value {
			st := int(call.Argument(0).ToInteger())
			shift := int(call.Argument(1).ToInteger())
			shift += 3 // Average
			if entity != nil {
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
				}, true, false, entity.Traits...)
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
			rnd := xrand.New()
			mid := 145 + (st-10)*15
			deviation := mid/5 + 2
			return r.ToValue(((mid + rnd.Intn(deviation) - rnd.Intn(deviation)) * shift) / 3)
		})
	}
	if entity != nil {
		settings := SheetSettingsFor(entity)
		m["exists"] = func() goja.Value { return r.ToValue(true) }
		m["playerName"] = func() goja.Value { return r.ToValue(entity.Profile.PlayerName) }
		m["name"] = func() goja.Value { return r.ToValue(entity.Profile.Name) }
		m["title"] = func() goja.Value { return r.ToValue(entity.Profile.Title) }
		m["organization"] = func() goja.Value { return r.ToValue(entity.Profile.Organization) }
		m["religion"] = func() goja.Value { return r.ToValue(entity.Profile.Religion) }
		m["techLevel"] = func() goja.Value { return r.ToValue(entity.Profile.TechLevel) }
		m["gender"] = func() goja.Value { return r.ToValue(entity.Profile.Gender) }
		m["age"] = func() goja.Value { return r.ToValue(entity.Profile.Age) }
		m["birthday"] = func() goja.Value { return r.ToValue(entity.Profile.Birthday) }
		m["eyes"] = func() goja.Value { return r.ToValue(entity.Profile.Eyes) }
		m["hair"] = func() goja.Value { return r.ToValue(entity.Profile.Hair) }
		m["skin"] = func() goja.Value { return r.ToValue(entity.Profile.Skin) }
		m["handedness"] = func() goja.Value { return r.ToValue(entity.Profile.Handedness) }
		m["heightInInches"] = func() goja.Value { return r.ToValue(fxp.AsFloat[float64](fxp.Int(entity.Profile.Height))) }
		m["weightInPounds"] = func() goja.Value { return r.ToValue(fxp.AsFloat[float64](fxp.Int(entity.Profile.Weight))) }
		m["displayHeightUnits"] = func() goja.Value { return r.ToValue(settings.DefaultLengthUnits.Key()) }
		m["displayWeightUnits"] = func() goja.Value { return r.ToValue(settings.DefaultWeightUnits.Key()) }
		m["sizeModifier"] = func() goja.Value { return r.ToValue(entity.Profile.AdjustedSizeModifier()) }
		m["liftingStrength"] = func() goja.Value { return r.ToValue(fxp.AsInteger[int](entity.LiftingStrength())) }
		m["strikingStrength"] = func() goja.Value { return r.ToValue(fxp.AsInteger[int](entity.StrikingStrength())) }
		m["throwingStrength"] = func() goja.Value { return r.ToValue(fxp.AsInteger[int](entity.ThrowingStrength())) }
		m["extraDiceFromModifiers"] = func() goja.Value { return r.ToValue(settings.UseModifyingDicePlusAdds) }
		m["attributes"] = func() goja.Value {
			return r.ToValue(func(_ goja.FunctionCall) goja.Value {
				list := entity.Attributes.List()
				attrs := make([]*goja.Object, 0, len(list))
				for _, attr := range list {
					if def := attr.AttributeDef(); def != nil {
						if def.IsSeparator() {
							continue
						}
						attrs = append(attrs, newScriptAttribute(r, attr))
					}
				}
				return r.ToValue(attrs)
			})
		}
		m["attribute"] = func() goja.Value {
			return r.ToValue(func(call goja.FunctionCall) goja.Value {
				if attr := entity.Attributes.Find(callArgAsString(call, 0)); attr != nil {
					return r.ToValue(newScriptAttribute(r, attr))
				}
				return goja.Undefined()
			})
		}
		m["traits"] = func() goja.Value {
			return r.ToValue(func(_ goja.FunctionCall) goja.Value {
				traits := make([]*goja.Object, 0, len(entity.Traits))
				for _, trait := range entity.Traits {
					if trait.Enabled() {
						traits = append(traits, newScriptTrait(r, trait))
					}
				}
				return r.ToValue(traits)
			})
		}
		m["findTraits"] = func() goja.Value {
			return r.ToValue(func(call goja.FunctionCall) goja.Value {
				name := callArgAsString(call, 0)
				tag := callArgAsString(call, 1)
				return findScriptTraits(r, name, tag, entity.Traits...)
			})
		}
		m["hasTrait"] = func() goja.Value {
			return r.ToValue(func(call goja.FunctionCall) goja.Value {
				found := false
				name := callArgAsTrimmedString(call, 0)
				Traverse(func(t *Trait) bool {
					if strings.EqualFold(t.NameWithReplacements(), name) {
						found = true
						return true
					}
					return false
				}, true, false, entity.Traits...)
				return r.ToValue(found)
			})
		}
		m["traitLevel"] = func() goja.Value {
			return r.ToValue(func(call goja.FunctionCall) goja.Value {
				name := callArgAsTrimmedString(call, 0)
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
				}, true, true, entity.Traits...)
				return r.ToValue(fxp.AsFloat[float64](level))
			})
		}
		m["skills"] = func() goja.Value {
			return r.ToValue(func(_ goja.FunctionCall) goja.Value {
				skills := make([]*goja.Object, 0, len(entity.Skills))
				for _, skill := range entity.Skills {
					if skill.Enabled() {
						skills = append(skills, newScriptSkill(r, skill))
					}
				}
				return r.ToValue(skills)
			})
		}
		m["findSkills"] = func() goja.Value {
			return r.ToValue(func(call goja.FunctionCall) goja.Value {
				name := callArgAsString(call, 0)
				specialization := callArgAsString(call, 1)
				tag := callArgAsString(call, 2)
				return findScriptSkills(r, name, specialization, tag, entity.Skills...)
			})
		}
		m["skillLevel"] = func() goja.Value {
			return r.ToValue(func(call goja.FunctionCall) goja.Value {
				name := callArgAsTrimmedString(call, 0)
				specialization := callArgAsTrimmedString(call, 1)
				if entity.isSkillLevelResolutionExcluded(name, specialization) {
					return r.ToValue(0)
				}
				entity.registerSkillLevelResolutionExclusion(name, specialization)
				defer entity.unregisterSkillLevelResolutionExclusion(name, specialization)
				relative := call.Argument(2).ToBoolean()
				var level int
				Traverse(func(s *Skill) bool {
					if strings.EqualFold(s.NameWithReplacements(), name) &&
						strings.EqualFold(s.SpecializationWithReplacements(), specialization) {
						s.UpdateLevel()
						if relative {
							level = fxp.AsInteger[int](s.LevelData.RelativeLevel)
						} else {
							level = fxp.AsInteger[int](s.LevelData.Level)
						}
						return true
					}
					return false
				}, true, true, entity.Skills...)
				return r.ToValue(level)
			})
		}
		m["spells"] = func() goja.Value {
			return r.ToValue(func(_ goja.FunctionCall) goja.Value {
				spells := make([]*goja.Object, 0, len(entity.Spells))
				for _, spell := range entity.Spells {
					if spell.Enabled() {
						spells = append(spells, newScriptSpell(r, spell))
					}
				}
				return r.ToValue(spells)
			})
		}
		m["findSpells"] = func() goja.Value {
			return r.ToValue(func(call goja.FunctionCall) goja.Value {
				name := callArgAsString(call, 0)
				tag := callArgAsString(call, 1)
				return findScriptSpells(r, name, tag, entity.Spells...)
			})
		}
		m["equipment"] = func() goja.Value {
			return r.ToValue(func(_ goja.FunctionCall) goja.Value {
				items := make([]*goja.Object, 0, len(entity.CarriedEquipment))
				for _, item := range entity.CarriedEquipment {
					if item.Quantity > 0 {
						items = append(items, newScriptEquipment(r, item))
					}
				}
				return r.ToValue(items)
			})
		}
		m["findEquipment"] = func() goja.Value {
			return r.ToValue(func(call goja.FunctionCall) goja.Value {
				name := callArgAsString(call, 0)
				tag := callArgAsString(call, 1)
				return findScriptEquipment(r, name, tag, entity.CarriedEquipment...)
			})
		}
		m["encumbrance"] = func() goja.Value {
			return r.ToValue(func(_ goja.FunctionCall) goja.Value {
				return r.ToValue(newScriptEncumbrance(r, entity))
			})
		}
		m["currentEncumbrance"] = func() goja.Value {
			return r.ToValue(func(call goja.FunctionCall) goja.Value {
				forSkills := call.Argument(0).ToBoolean()
				returnMoveFactor := call.Argument(1).ToBoolean()
				level := int(entity.EncumbranceLevel(forSkills))
				if returnMoveFactor {
					return r.ToValue(fxp.AsFloat[float64](fxp.One - fxp.FromInteger(level).Mul(fxp.Two).Div(fxp.Ten)))
				}
				return r.ToValue(level)
			})
		}
		m["weaponDamage"] = func() goja.Value {
			return r.ToValue(func(call goja.FunctionCall) goja.Value {
				name := callArgAsString(call, 0)
				usage := callArgAsString(call, 1)
				for _, w := range entity.Weapons(true, false) {
					if strings.EqualFold(w.String(), name) && strings.EqualFold(w.UsageWithReplacements(), usage) {
						return r.ToValue(w.Damage.ResolvedDamage(nil))
					}
				}
				for _, w := range entity.Weapons(false, false) {
					if strings.EqualFold(w.String(), name) && strings.EqualFold(w.UsageWithReplacements(), usage) {
						return r.ToValue(w.Damage.ResolvedDamage(nil))
					}
				}
				return r.ToValue("")
			})
		}
	}
	return r.NewDynamicObject(NewScriptObject(r, m))
}
