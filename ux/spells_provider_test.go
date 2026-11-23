// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// TestDroppedSpellPicksItsDefaultFirst checks that a spell dropped on the sheet picks its default ahead of the items
// already there, as a dropped skill does, so that one dropped beside a skill the two of which default to each other
// rests on the skill, rather than the skill, which is updated first, resting on the spell.
func TestDroppedSpellPicksItsDefaultFirst(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	thaumatology := gurps.NewSkill(entity, nil, false)
	thaumatology.Name = "Thaumatology"
	thaumatology.Points = fxp.Four
	thaumatology.Defaults = []*gurps.SkillDefault{{
		DefaultType: gurps.SpellID,
		Name:        criteria.Text{Compare: criteria.IsText, Qualifier: "Lore"},
		Modifier:    -fxp.Two,
	}}
	entity.Skills = []*gurps.Skill{thaumatology}
	sheet.Rebuild(true)
	lore := gurps.NewSpell(entity, nil, false)
	lore.Name = "Lore"
	lore.Points = fxp.Four
	lore.Defaults = []*gurps.SkillDefault{{
		DefaultType: gurps.SkillID,
		Name:        criteria.Text{Compare: criteria.IsText, Qualifier: "Thaumatology"},
		Modifier:    -fxp.Two,
	}}
	entity.Spells = []*gurps.Spell{lore}
	table := sheet.Spells.Table
	table.SyncToModel()
	table.SetSelectionMap(map[tid.TID]bool{lore.TID: true})
	provider, ok := any(table.Model).(TableProvider[*gurps.Spell])
	if !ok {
		t.Fatal("the spells table has no provider")
	}
	provider.ProcessDropData(nil, table)
	entity.Recalculate()
	c.NotNil(lore.DefaultedFrom, "the dropped spell rests on the skill")
	c.Nil(thaumatology.DefaultedFrom, "which so can't rest on the spell")
}
