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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// Sheet.MarkModified is a no-op while the sheet is inside an update pass, so an adjustment that left the
// recalculation to it would silently drop it, leaving the sheet showing stale levels and points until some unrelated
// later edit happened to trigger a pass of its own.
func TestValueAdjustmentRecalculatesDuringUpdatePass(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	skill := gurps.NewSkill(entity, nil, false)
	skill.Name = "Brawling"
	skill.Points = fxp.Four
	entity.Skills = append(entity.Skills, skill)
	entity.Recalculate()
	before := skill.LevelData.Level

	sheet.awaitingUpdate = true
	t.Cleanup(func() { sheet.awaitingUpdate = false })
	adjustment := &snapshotList[*gurps.Attribute, fxp.Int]{
		owner:  sheet,
		entity: entity,
		set:    func(attr *gurps.Attribute, value fxp.Int) { attr.Adjustment = value },
		list: []valueSnapshot[*gurps.Attribute, fxp.Int]{
			{target: entity.Attributes.Set["dx"], value: fxp.Four},
		},
	}
	adjustment.apply()

	c.Equal(before+fxp.Four, skill.LevelData.Level, "the skill level must reflect the raised attribute")
}

// Toggling a trait's enablement alters more of the sheet than the traits list -- a trait that is turned off takes its
// features, weapons, reactions and conditional modifiers out of play with it, and whether the lists showing those are
// carried on the page at all is decided only when the sheet creates its lists -- so the owner is rebuilt. Marking it
// as modified first would recalculate the entity, re-sync every table, refresh the search results and reacquire the
// focus, all of which the rebuild would then immediately do again. The rebuild has to bump the modification timestamp
// itself, since that is the one thing marking as modified would have done that it doesn't.
func TestToggleDisabledUpdatesTheSheetOnlyOnce(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	skill := gurps.NewSkill(entity, nil, false)
	skill.Name = "Brawling"
	skill.Points = fxp.Four
	entity.Skills = []*gurps.Skill{skill}
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Claws"
	bonus := gurps.NewSkillBonus()
	bonus.NameCriteria.Qualifier = skill.Name
	bonus.SetOwner(trait)
	trait.Features = gurps.Features{bonus}
	trait.Weapons = []*gurps.Weapon{gurps.NewWeapon(trait, true)}
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)
	table := sheet.Traits.Table
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	boosted := skill.LevelData.Level
	c.True(listAttachedToSheet(sheet, sheet.MeleeWeapons), "the melee weapons list starts out on the page")

	table.SetSelectionMap(map[tid.TID]bool{trait.ID(): true})
	counter := installSyncCounter(sheet)
	entity.ModifiedOn = jio.Time{}
	toggleDisabled(sheet, table)
	c.True(trait.IsDisabled(), "the trait must have been turned off")
	c.Equal(1, counter.count, "toggling enablement must update the sheet exactly once")
	c.Equal(boosted-fxp.One, skill.LevelData.Level,
		"the entity must have been recalculated with the trait's bonus out of play")
	c.False(listAttachedToSheet(sheet, sheet.MeleeWeapons),
		"the rebuild must take the now-empty melee weapons list off the page")
	c.NotEqual(jio.Time{}, entity.ModifiedOn, "an edit that rebuilds must still bump the modification timestamp")

	counter.count = 0
	entity.ModifiedOn = jio.Time{}
	c.True(mgr.CanUndo(), "toggling enablement must be undoable")
	mgr.Undo()
	c.False(trait.IsDisabled(), "undo must turn the trait back on")
	c.Equal(1, counter.count, "the undo must update the sheet exactly once as well")
	c.Equal(boosted, skill.LevelData.Level, "undo must bring the bonus back into play")
	c.True(listAttachedToSheet(sheet, sheet.MeleeWeapons), "undo must put the melee weapons list back on the page")
	c.NotEqual(jio.Time{}, entity.ModifiedOn, "the undo must bump the modification timestamp, too")
}

// The checks that keep a single edit from recalculating the entity twice: only the sheet holding the entity
// recalculates it on its own, when marked as modified or rebuilt. Marking it as modified does nothing while it is
// inside an update pass, so an edit that left the recalculation to it then would drop it rather than defer it, while
// rebuilding is never suppressed. Anything else leaves the recalculation to the caller.
func TestOwnerRecalculatesOnlyForItsOwnSheet(t *testing.T) {
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	for _, tc := range []struct {
		name     string
		owner    unison.Paneler
		entity   *gurps.Entity
		updating bool
		modified bool // What ownerRecalculates returns.
		rebuilt  bool // What ownerRebuildRecalculates returns.
	}{
		{name: "sheet", owner: sheet, entity: entity, modified: true, rebuilt: true},
		{name: "panel within the sheet", owner: sheet.Traits, entity: entity, modified: true, rebuilt: true},
		{name: "sheet in an update pass", owner: sheet, entity: entity, updating: true, rebuilt: true},
		{name: "panel in an update pass", owner: sheet.Traits, entity: entity, updating: true, rebuilt: true},
		{name: "another entity", owner: sheet, entity: gurps.NewEntity()},
		{name: "no entity", owner: sheet},
		{name: "no owner", entity: entity},
		{name: "panel outside a modifiable root", owner: unison.NewPanel(), entity: entity},
		{name: "template", owner: NewTemplate("test"+gurps.TemplatesExt, gurps.NewTemplate()), entity: entity},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			sheet.awaitingUpdate = tc.updating
			t.Cleanup(func() { sheet.awaitingUpdate = false })
			c.Equal(tc.modified, ownerRecalculates(tc.owner, tc.entity), "when marked as modified")
			c.Equal(tc.rebuilt, ownerRebuildRecalculates(tc.owner, tc.entity), "when rebuilt")
		})
	}
}
