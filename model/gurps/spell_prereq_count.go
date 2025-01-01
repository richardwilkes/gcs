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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellcmp"
)

// CountPrereqsForSpell returns the number of prerequisites for the specified spell.
func CountPrereqsForSpell(spell *Spell, availableSpells []*Spell, nonSpellsCountAs int, useHighestInOr bool) int {
	return countPrereqsForList(spell.Prereq, availableSpells, nonSpellsCountAs, useHighestInOr)
}

func countPrereqsForList(list *PrereqList, availableSpells []*Spell, nonSpellsCountAs int, useHighestInOr bool) int {
	counts := make([]int, len(list.Prereqs))
	for i, prereq := range list.Prereqs {
		switch p := prereq.(type) {
		case *PrereqList:
			counts[i] = countPrereqsForList(p, availableSpells, nonSpellsCountAs, useHighestInOr)
		case *TraitPrereq:
			if p.Has {
				switch p.LevelCriteria.Compare {
				case criteria.EqualsNumber, criteria.AtLeastNumber:
					counts[i] = nonSpellsCountAs * max(fxp.As[int](p.LevelCriteria.Qualifier), 1)
				default:
					counts[i] = nonSpellsCountAs
				}
			}
		case *AttributePrereq:
			if p.Has {
				counts[i] = nonSpellsCountAs
			}
		case *ContainedQuantityPrereq:
			if p.Has {
				counts[i] = nonSpellsCountAs
			}
		case *ContainedWeightPrereq:
			if p.Has {
				counts[i] = nonSpellsCountAs
			}
		case *SkillPrereq:
			if p.Has {
				counts[i] = nonSpellsCountAs
			}
		case *SpellPrereq:
			if p.Has {
				switch p.QuantityCriteria.Compare {
				case criteria.EqualsNumber, criteria.AtLeastNumber:
					counts[i] = fxp.As[int](p.QuantityCriteria.Qualifier)
				default:
					counts[i] = 1
				}
				if counts[i] == 1 && p.SubType == spellcmp.Name && p.QualifierCriteria.Compare == criteria.IsText {
					Traverse(func(s *Spell) bool {
						if strings.EqualFold(s.NameWithReplacements(), p.QualifierCriteria.Qualifier) {
							counts[i] = 1 + countPrereqsForList(s.Prereq, availableSpells, nonSpellsCountAs, useHighestInOr)
							return true
						}
						return false
					}, false, true, availableSpells...)
				}
			}
		default:
			counts[i] = nonSpellsCountAs
		}
	}
	if list.All {
		total := 0
		for _, count := range counts {
			total += count
		}
		return total
	}
	result := 0
	for _, count := range counts {
		if result < count {
			result = count
		}
	}
	if !useHighestInOr {
		for _, count := range counts {
			if count > 0 && result > count {
				result = count
			}
		}
	}
	return result
}
