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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// The level cells show the level the last recalculation recorded rather than computing one of their own: computing
// a level walks the skills list, which for a sheet with hundreds of spells was most of what every update cost, and
// the recorded level is what every other reader of a level, from the weapons to the prerequisites, already uses.
func TestLevelCellsShowRecordedLevel(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.SheetSettings.ShowSpellAdj = true
	sk := addTestSkill(e, "Broadsword", "", "", fxp.Four)
	sk.Difficulty.Attribute = DexterityID
	sp := addTestSpell(e, "Fireball", fxp.Four)
	spellBonus := NewSpellBonus()
	spellBonus.Amount = fxp.Two
	addTraitWithFeatures(e, "Magery", spellBonus)
	e.Recalculate()

	levelCell := func(node interface{ CellData(int, *CellData) }, column int) CellData {
		var data CellData
		node.CellData(column, &data)
		return data
	}
	c.Equal(sk.LevelData.LevelAsString(false), levelCell(sk, SkillLevelColumn).Primary)
	c.Equal(sp.LevelData.LevelAsString(false), levelCell(sp, SpellLevelColumn).Primary)
	skillBefore := levelCell(sk, SkillLevelColumn).Primary
	spellBefore := levelCell(sp, SpellLevelColumn).Primary
	ritualsBefore := sp.Rituals()
	tooltipBefore := levelCell(sp, SpellRelativeLevelColumn).Tooltip
	c.NotEqual("-", skillBefore)
	c.NotEqual("-", spellBefore)
	c.NotEqual("", sp.LevelData.Tooltip, "the spell bonus is recorded in the level's tooltip")
	c.Equal(IncludesModifiersFrom()+":"+sp.LevelData.Tooltip, tooltipBefore)

	// A change that hasn't been recalculated yet isn't shown...
	sk.Points = fxp.FromInteger(24)
	sp.Points = fxp.FromInteger(24)
	spellBonus.Amount = fxp.Three
	c.Equal(skillBefore, levelCell(sk, SkillLevelColumn).Primary, "the skill cell shows the recorded level")
	c.Equal(spellBefore, levelCell(sp, SpellLevelColumn).Primary, "the spell cell shows the recorded level")
	c.Equal(ritualsBefore, sp.Rituals(), "the rituals follow the recorded level")
	c.Equal(tooltipBefore, levelCell(sp, SpellRelativeLevelColumn).Tooltip,
		"the relative level cell carries the recorded tooltip")

	// ...until it is.
	e.Recalculate()
	c.NotEqual(skillBefore, levelCell(sk, SkillLevelColumn).Primary, "the skill cell shows the new level")
	c.NotEqual(spellBefore, levelCell(sp, SpellLevelColumn).Primary, "the spell cell shows the new level")
	c.Equal(sk.LevelData.LevelAsString(false), levelCell(sk, SkillLevelColumn).Primary)
	c.Equal(sp.LevelData.LevelAsString(false), levelCell(sp, SpellLevelColumn).Primary)
	c.NotEqual(ritualsBefore, sp.Rituals(), "the rituals follow the new level")
	c.NotEqual(tooltipBefore, levelCell(sp, SpellRelativeLevelColumn).Tooltip, "the tooltip follows the new bonus")
	c.Equal(IncludesModifiersFrom()+":"+sp.LevelData.Tooltip, levelCell(sp, SpellRelativeLevelColumn).Tooltip)

	// A disabled spell has no level to show, but its rituals are still those of the level it would have.
	enabledRituals := sp.Rituals()
	sp.Disabled = true
	e.Recalculate()
	c.Equal("-", levelCell(sp, SpellLevelColumn).Primary)
	c.Equal(fxp.Int(0), sp.LevelData.Level)
	c.Equal(enabledRituals, sp.Rituals(), "a disabled spell reports the rituals of the level it would have")
}
