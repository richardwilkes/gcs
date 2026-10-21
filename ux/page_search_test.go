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
	"fmt"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// matchLabels returns the matches as "name | title" strings.
func matchLabels(matches []*pageSearchMatch) []string {
	labels := make([]string, len(matches))
	for i, m := range matches {
		labels[i] = m.name + " | " + m.title
	}
	return labels
}

// newFindableTrait returns a trait of the given name owned by the entity.
func newFindableTrait(entity *gurps.Entity, name string) *gurps.Trait {
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = name
	return trait
}

// The matches are labeled with what the row's hierarchy column shows -- not column 0, which on a sheet is the equipped
// box for carried equipment -- cut to its first line, since a note's String() is its whole text, and carry the title of
// the list they were found in, in block order. A template's and a loot sheet's equipment lists are headed plainly
// "Equipment", as their blocks are.
func TestPageSearchMatchesCarryTheirListTitles(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.Traits = []*gurps.Trait{newFindableTrait(entity, "Findable Trait")}
	skill := gurps.NewSkill(entity, nil, false)
	skill.Name = "Findable Skill"
	entity.Skills = []*gurps.Skill{skill}
	sword := gurps.NewEquipment(entity, nil, false)
	sword.Name = "Findable Sword"
	entity.CarriedEquipment = []*gurps.Equipment{sword}
	note := gurps.NewNote(entity, nil, false)
	note.MarkDown = "# Findable note\nwith a second line"
	entity.Notes = []*gurps.Note{note}
	sheet.Rebuild(true)
	c.Equal([]string{
		"Findable Trait | Traits",
		"Findable Skill | Skills",
		"Findable Sword | Carried Equipment",
		"Findable note | Notes",
	}, matchLabels(sheet.search.findMatches("findable")))

	data := gurps.NewTemplate()
	gear := gurps.NewEquipment(nil, nil, false)
	gear.Name = "Findable Gear"
	data.Equipment = []*gurps.Equipment{gear}
	template := newTestTemplateDockable("Search", data)
	c.Equal([]string{"Findable Gear | Equipment"}, matchLabels(template.search.findMatches("findable")))

	loot := newTestLootSheet(t)
	lootGear := gurps.NewEquipment(loot.loot, nil, false)
	lootGear.Name = "Findable Loot"
	loot.loot.Equipment = []*gurps.Equipment{lootGear}
	lootNote := gurps.NewNote(loot.loot, nil, false)
	lootNote.MarkDown = "Findable loot note"
	loot.loot.Notes = []*gurps.Note{lootNote}
	loot.Rebuild(true)
	c.Equal([]string{"Findable Loot | Equipment", "Findable loot note | Notes"},
		matchLabels(loot.search.findMatches("findable")))
}

// A long label is cut short, and a weapon, whose description column shows its owner, adds its usage so that one
// weapon's attacks can be told apart, even when its owner's name is cut short.
func TestPageSearchLabels(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	blade := newFindableTrait(entity, "Findable Blade")
	swing := gurps.NewWeapon(blade, true)
	swing.Usage = "Swung"
	thrust := gurps.NewWeapon(blade, true)
	thrust.Usage = "Thrust"
	blade.Weapons = []*gurps.Weapon{swing, thrust}
	longName := "Findable " + strings.Repeat("y", maxPageSearchLabelLength)
	cutName := "Findable " + strings.Repeat("y", maxPageSearchLabelLength-len("Findable ")) + "…"
	longBlade := newFindableTrait(entity, longName)
	longSwing := gurps.NewWeapon(longBlade, true)
	longSwing.Usage = "Swung"
	longThrust := gurps.NewWeapon(longBlade, true)
	longThrust.Usage = "Thrust"
	longBlade.Weapons = []*gurps.Weapon{longSwing, longThrust}
	entity.Traits = []*gurps.Trait{blade, longBlade}
	note := gurps.NewNote(entity, nil, false)
	note.MarkDown = "Findable " + strings.Repeat("x", maxPageSearchLabelLength)
	entity.Notes = []*gurps.Note{note}
	sheet.Rebuild(true)
	c.Equal([]string{
		"Findable Blade (Swung) | Melee Attacks",
		"Findable Blade (Thrust) | Melee Attacks",
		cutName + " (Swung) | Melee Attacks",
		cutName + " (Thrust) | Melee Attacks",
		"Findable Blade | Traits",
		cutName + " | Traits",
		"Findable " + strings.Repeat("x", maxPageSearchLabelLength-len("Findable ")) + "… | Notes",
	}, matchLabels(sheet.search.findMatches("findable")))
}

// A row of the dropdown shows the match's name followed by the title of its list, and is named for both to an
// assistive technology.
func TestPageSearchCellShowsTheListTitle(t *testing.T) {
	c := check.New(t)
	list := unison.NewList[*pageSearchMatch]()
	match := &pageSearchMatch{name: "Findable", title: "Traits"}
	label, ok := (&pageSearchCellFactory{}).CreateCell(list, match, 0, unison.Black, unison.White, false, false).(*unison.Label)
	if !ok {
		t.Fatal("the cell must be a label")
	}
	c.Equal("Findable  Traits", label.String())
	c.Equal("Findable, Traits", label.Accessibility.Name)
}

// The lists are only walked while the matches can be seen: not for an empty search, not for text put in the field while
// it lacks the focus, which can't show them, and not by the refresh every update performs while they are hidden. The
// walk has every row of every list derive its children and the data of its cells afresh, which was most of what an
// update of a sheet with large lists cost while nobody was searching. Nothing is selected until a match is chosen.
func TestPageSearchDoesNotWalkWhileHidden(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.Traits = []*gurps.Trait{newFindableTrait(entity, "Findable")}
	sheet.Rebuild(true)
	search := sheet.search
	walks := 0
	findMatches := search.findMatches
	search.findMatches = func(text string) []*pageSearchMatch {
		walks++
		return findMatches(text)
	}

	search.Refresh()
	c.Equal(0, walks, "an empty search must not walk the lists")

	search.field.SetText("findable")
	c.Equal(0, walks, "text set while the field lacks the focus must not walk the lists")
	c.Equal(0, search.suggestions.List().Count(), "nothing is on offer")
	c.False(search.suggestions.Showing())

	search.Refresh()
	c.Equal(0, walks, "a refresh while the matches are hidden must not walk the lists")
	c.Equal(0, sheet.Traits.Table.SelectionCount(), "nothing is selected until a match is chosen")

	c.Equal(1, len(findMatches("findable")), "the walk itself finds the trait")
	c.Equal(0, len(findMatches("")), "but nothing for an empty search")
}

// The walk stops once there are as many matches as the dropdown offers.
func TestPageSearchCapsTheMatches(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	for i := range maxPageSearchMatches + 50 {
		entity.Traits = append(entity.Traits, newFindableTrait(entity, fmt.Sprintf("Findable %d", i)))
	}
	sheet.Rebuild(true)
	c.Equal(maxPageSearchMatches, len(sheet.search.findMatches("findable")))
}
