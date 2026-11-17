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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison/enums/align"
)

// TestForSortDistinguishesSwitchStates verifies that the on, off and nothing-to-switch states of a switch column sort
// as three distinct groups (see switchOffSortValue). Off once sorted as empty, like a toggle, interleaving those rows
// with the ones that have no switch.
func TestForSortDistinguishesSwitchStates(t *testing.T) {
	c := check.New(t)
	on := (&CellData{Type: cell.Switch, Checked: true}).ForSort()
	off := (&CellData{Type: cell.Switch}).ForSort()
	nothingToSwitch := (&CellData{}).ForSort()

	c.NotEqual(on, off, "an on switch must not sort as an off switch")
	c.NotEqual(off, nothingToSwitch, "an off switch must not sort as a row with nothing to switch")
	c.NotEqual(on, nothingToSwitch, "an on switch must not sort as a row with nothing to switch")
	c.Equal("", nothingToSwitch, "a row with nothing to switch must sort as nothing")
	c.NotEqual("", off, "an off switch must sort as something")

	// The table sorts with xstrings.NaturalLess, so check the grouping the chosen values actually produce.
	c.True(xstrings.NaturalLess(nothingToSwitch, off, true), "rows with nothing to switch sort ahead of off switches")
	c.True(xstrings.NaturalLess(off, on, true), "off switches sort ahead of on switches")

	// A search of the sheet matches against the same strings, so the off state must not be something that could be
	// typed into the search field while looking for something else.
	c.NotContains(off, "-", "an off switch must not be matched by a search for an ASCII hyphen")

	// Toggles remain a two-state cell: they are present on every row of their column, so there is no third state to
	// distinguish and the off state stays empty.
	c.Equal(on, (&CellData{Type: cell.Toggle, Checked: true}).ForSort(), "an on toggle sorts as an on switch does")
	c.Equal("", (&CellData{Type: cell.Toggle}).ForSort(), "an off toggle sorts as nothing")
}

// TestForSortMatchesTheShownText verifies that a cell sorts and searches by the text it shows: a text cell by both of
// its lines, a page reference or Markdown cell by its primary text alone, and the tags cell fillTagsCell makes by its
// tags.
func TestForSortMatchesTheShownText(t *testing.T) {
	c := check.New(t)
	var tags CellData
	fillTagsCell(&tags, []string{"Alpha", "Beta"})
	c.Equal(cell.Tags, tags.Type, "fillTagsCell makes a tags cell")
	for _, one := range []struct {
		data CellData
		want string
	}{
		{CellData{Type: cell.Text, Primary: "Name"}, "Name"},
		{CellData{Type: cell.Text, Primary: "Name", Secondary: "Notes"}, "Name\nNotes"},
		{CellData{Type: cell.PageRef, Primary: "B12", Secondary: "Name"}, "B12"},
		{tags, "Alpha, Beta"},
		{CellData{Type: cell.Markdown, Primary: "**Bold**"}, "**Bold**"},
	} {
		c.Equal(one.want, one.data.ForSort(), "a %v cell that should sort as %q", one.data.Type, one.want)
	}
}

// TestJoinTooltipBlocks verifies that two tooltip blocks are set apart by a rule, with a blank line ahead of it so that
// Markdown doesn't make a heading of the first block's last line, and that an empty block adds nothing.
func TestJoinTooltipBlocks(t *testing.T) {
	c := check.New(t)
	c.Equal("", JoinTooltipBlocks("", ""))
	c.Equal("first", JoinTooltipBlocks("first", ""))
	c.Equal("second", JoinTooltipBlocks("", "second"))
	c.Equal("first\n\n---\nsecond", JoinTooltipBlocks("first", "second"))
}

// TestSharedHeaderData verifies that the icon columns the node types share show their icons, carry the names a screen
// reader announces for them, and explain themselves in their tooltips.
func TestSharedHeaderData(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		header  HeaderData
		icon    string
		name    string
		tooltip string
	}{
		{pageRefHeaderData(), HeaderBookmark, "Page Reference", PageRefTooltip()},
		{libSrcHeaderData(), HeaderDatabase, "Library Source", LibSrcTooltip()},
		{switchHeaderData(), HeaderSwitch, "Switched On", SwitchHeaderTooltip()},
		{enabledHeaderData(), HeaderCheckmark, "Enabled", ModifierEnabledTooltip()},
	} {
		c.Equal(one.icon, one.header.Title)
		c.Equal(one.name, one.header.Name, "the name of the %s column", one.icon)
		c.Equal(one.tooltip, one.header.Detail, "the tooltip of the %s column", one.icon)
	}
}

// TestHeaderDataNamesWhatIsNotReadable walks every header data function, so that a new icon column, or one titled with
// an abbreviation of up to three characters, is caught if it lacks a name.
func TestHeaderDataNamesWhatIsNotReadable(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	functions := map[string]func(columnID int) HeaderData{
		"conditional modifiers": ConditionalModifiersHeaderData,
		"reaction modifiers":    ReactionModifiersHeaderData,
		"equipment modifiers":   EquipmentModifierHeaderData,
		"notes":                 NotesHeaderData,
		"skills":                SkillsHeaderData,
		"trait modifiers":       TraitModifierHeaderData,
		"spells":                SpellsHeaderData,
		"traits":                TraitsHeaderData,
		"melee weapons":         func(columnID int) HeaderData { return WeaponHeaderData(columnID, true, false) },
		"ranged weapons":        func(columnID int) HeaderData { return WeaponHeaderData(columnID, false, false) },
		"page melee weapons":    func(columnID int) HeaderData { return WeaponHeaderData(columnID, true, true) },
		"carried equipment":     func(columnID int) HeaderData { return EquipmentHeaderData(columnID, entity, true, false) },
		"other equipment":       func(columnID int) HeaderData { return EquipmentHeaderData(columnID, entity, false, true) },
	}
	abbreviations := map[string]string{
		"±":    "Modifier",
		"#":    "Quantity",
		"TL":   "Tech Level",
		"LC":   "Legality Class",
		"Diff": "Difficulty",
		"SL":   "Skill Level",
		"RSL":  "Relative Skill Level",
		"Pts":  "Points",
		"P#":   "Prerequisite Count",
		"ST":   "Minimum Strength",
		"Acc":  "Accuracy Bonus",
		"RoF":  "Rate of Fire",
	}
	seen := make(map[string]bool)
	for name, f := range functions {
		columns := 0
		// IDs that name no column, such as a retired one, have no title. 64 is well past the last ID of any kind.
		for columnID := range 64 {
			data := f(columnID)
			if data.Title == "" {
				continue
			}
			columns++
			switch {
			case data.TitleIsImageKey:
				c.NotEqual("", data.Name, "%s column %d, an icon, must carry a name", name, columnID)
			case len([]rune(data.Title)) <= 3:
				c.NotEqual("", data.Name, "%s column %d, titled %q, must carry the name it stands for", name, columnID,
					data.Title)
			}
			if expected, ok := abbreviations[data.Title]; ok {
				seen[data.Title] = true
				c.Equal(expected, data.Name, "%s column %d, titled %q, must be named for what it stands for", name,
					columnID, data.Title)
				c.Equal(expected, data.Detail, "%s column %d, titled %q, must explain itself in its tooltip", name,
					columnID, data.Title)
			}
			if data.Name != "" {
				c.True(data.TitleIsImageKey || abbreviations[data.Title] != "",
					"%s column %d, titled %q, is a word and needs no name", name, columnID, data.Title)
			}
		}
		c.True(columns > 1, "%s must have columns", name)
	}
	for title := range abbreviations {
		c.True(seen[title], "the abbreviation %q must be in use somewhere", title)
	}
}

// TestFillPageRefCell verifies that a page reference cell gives the text to look for on the page as the highlight when
// there is one and otherwise as what the fallback produces, which is only computed when it is needed.
func TestFillPageRefCell(t *testing.T) {
	c := check.New(t)
	calls := 0
	fallback := func() string {
		calls++
		return "fallback"
	}
	var data CellData
	fillPageRefCell(&data, "B12", "highlight", fallback)
	c.Equal(cell.PageRef, data.Type)
	c.Equal("B12", data.Primary)
	c.Equal("highlight", data.Secondary, "the highlight wins when there is one")
	c.Equal(0, calls, "the fallback is not evaluated when the highlight is used")
	data = CellData{}
	fillPageRefCell(&data, "B12", "", fallback)
	c.Equal("fallback", data.Secondary, "the fallback is used when there is no highlight")
	c.Equal(1, calls)
}

// TestFillLibSrcCell verifies that a library source cell shows how the node compares to its library source, with the
// source's details in the tooltip when it has one, and shows nothing when there is no owner to ask.
func TestFillLibSrcCell(t *testing.T) {
	c := check.New(t)
	trait := NewTrait(nil, nil, false)
	var data CellData
	fillLibSrcCell(&data, nil, trait)
	c.Equal(cell.Text, data.Type)
	c.Equal(align.Middle, data.Alignment)
	c.Equal("", data.Primary, "without an owner there is nobody to ask about the source")
	c.Equal("", data.Tooltip)

	entity := NewEntity()
	data = CellData{}
	fillLibSrcCell(&data, entity, trait)
	c.Equal(srcstate.Custom.AltString(), data.Primary)
	c.Equal(srcstate.Custom.String(), data.Tooltip, "custom data has no source details to show")

	trait.Source = Source{Library: "Test Library", Path: "Test" + TraitsExt, TID: "SourceTID"}
	data = CellData{}
	fillLibSrcCell(&data, entity, trait)
	c.Equal(srcstate.Missing.AltString(), data.Primary)
	for _, detail := range []string{srcstate.Missing.String(), "Test Library", "Test" + TraitsExt, "SourceTID"} {
		c.Contains(data.Tooltip, detail, "the tooltip gives the state and the source's details")
	}
}

// TestLevelCellsShowRecordedLevel verifies that the level cells show the level the last recalculation recorded rather
// than computing one of their own: computing a level walks the skills list, which for a sheet with hundreds of spells
// was most of what every update cost, and the recorded level is what every other reader of a level, from the weapons
// to the prerequisites, already uses.
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
