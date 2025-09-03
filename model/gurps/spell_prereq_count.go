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
	"fmt"
	"maps"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellcmp"
)

// CountPrereqsForSpell returns the number of prerequisites for the specified spell.
func CountPrereqsForSpell(spell *Spell, allSpells []*Spell) int {
	collect := make(map[string]int)
	collectPrereqsForPrereq(spell.Prereq, allSpells, collect)
	return countPrereqSet(collect)
}

func collectPrereqsForPrereq(one Prereq, allSpells []*Spell, collect map[string]int) {
	switch p := one.(type) {
	case *PrereqList:
		if p.All {
			for _, one := range p.Prereqs {
				collectPrereqsForPrereq(one, allSpells, collect)
			}
		} else {
			var current map[string]int
			for _, one := range p.Prereqs {
				set := make(map[string]int)
				maps.Copy(set, collect)
				collectPrereqsForPrereq(one, allSpells, set)
				if current == nil || countPrereqSet(set) < countPrereqSet(current) {
					current = set
				}
			}
			maps.Copy(collect, current)
		}
	case *TraitPrereq:
		if p.Has {
			switch p.LevelCriteria.Compare {
			case criteria.EqualsNumber, criteria.AtLeastNumber:
				levels := fxp.AsInteger[int](p.LevelCriteria.Qualifier)
				base := 1
				if p.NameCriteria.Qualifier == "Magery" {
					base = 0
				}
				for i := base; i <= levels; i++ {
					collect[fmt.Sprintf("T:%s %d", strings.ToLower(p.NameCriteria.Qualifier), i)] = 1
				}
			default:
				collect["T:"+strings.ToLower(p.NameCriteria.Qualifier)] = 1
			}
		}
	case *AttributePrereq:
		if p.Has {
			collect["A:"+strings.ToLower(p.Which+p.CombinedWith)] = 1
		}
	case *SkillPrereq:
		if p.Has {
			collect["Sk:"+strings.ToLower(p.NameCriteria.Qualifier)] = 1
		}
	case *SpellPrereq:
		if p.Has {
			count := 1
			if p.QuantityCriteria.Compare == criteria.EqualsNumber ||
				p.QuantityCriteria.Compare == criteria.AtLeastNumber {
				count = fxp.AsInteger[int](p.QuantityCriteria.Qualifier)
			}
			key := "Sp:" + strings.ToLower(p.QualifierCriteria.Qualifier)
			needTraverse := count == 1 && p.SubType == spellcmp.Name && p.QualifierCriteria.Compare == criteria.IsText
			if v, ok := collect[key]; ok {
				if v < count {
					collect[key] = count
				}
				needTraverse = false
			} else {
				collect[key] = count
			}
			if needTraverse {
				Traverse(func(s *Spell) bool {
					if !strings.EqualFold(s.Name, p.QualifierCriteria.Qualifier) {
						return false
					}
					collectPrereqsForPrereq(s.Prereq, allSpells, collect)
					return true
				}, false, true, allSpells...)
			}
		}
	case *ScriptPrereq:
		count := 1
		script := strings.ToLower(p.Script)
		if revised, found := strings.CutPrefix(script, "// prereq count:"); found {
			if n, err := strconv.Atoi(strings.TrimSpace(strings.SplitN(revised, "\n", 2)[0])); err == nil {
				count = n
			}
		}
		collect["Sc:"+script] = count
	}
}

func countPrereqSet(set map[string]int) int {
	var total int
	for _, count := range set {
		total += count
	}
	return total
}
