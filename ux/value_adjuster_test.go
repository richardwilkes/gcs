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

// TestOwnerRecalculatesSkipsSheetAlreadyUpdating verifies that an edit only leaves the recalculation to the owner when
// the owner is actually going to perform one. A sheet that is already inside an update pass is not, since
// Sheet.MarkModified does nothing at all while awaitingUpdate is set.
func TestOwnerRecalculatesSkipsSheetAlreadyUpdating(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()

	c.True(ownerRecalculates(sheet, entity), "a sheet recalculates its own entity when marked as modified")
	sheet.awaitingUpdate = true
	c.False(ownerRecalculates(sheet, entity), "a sheet already inside an update pass won't recalculate again")
	c.False(ownerRecalculates(sheet.Traits, entity), "a panel within that sheet resolves to the same answer")
	sheet.awaitingUpdate = false
	c.True(ownerRecalculates(sheet, entity), "the sheet recalculates again once the update pass is done")
}

// TestValueAdjustmentRecalculatesDuringUpdatePass verifies that an adjustment made while the sheet is already inside an
// update pass still brings the derived state up to date. Sheet.MarkModified is a no-op in that case, so an adjustment
// that left the recalculation to it would silently drop it, leaving the sheet showing stale levels and points until
// some unrelated later edit happened to trigger a pass of its own.
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

// TestToggleDisabledUpdatesTheSheetOnlyOnce verifies that toggling a trait's enablement pays for the sheet's update
// exactly once. The change alters more of the sheet than the traits list -- a trait that is turned off takes its
// features, weapons, reactions and conditional modifiers out of play with it, and whether the lists showing those are
// carried on the page at all is decided only when the sheet creates its lists -- so the owner is rebuilt. Marking it
// as modified first would recalculate the entity, re-sync every table, refresh the search results and reacquire the
// focus, and the rebuild would then immediately do all of it over again; on a sheet with many rows that work is the
// entire cost of the edit. The rebuild has to bump the modification timestamp itself, since that is the one thing
// marking as modified would have done that it doesn't.
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
	c.True(trait.Disabled, "the trait must have been turned off")
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
	c.False(trait.Disabled, "undo must turn the trait back on")
	c.Equal(1, counter.count, "the undo must update the sheet exactly once as well")
	c.Equal(boosted, skill.LevelData.Level, "undo must bring the bonus back into play")
	c.True(listAttachedToSheet(sheet, sheet.MeleeWeapons), "undo must put the melee weapons list back on the page")
	c.NotEqual(jio.Time{}, entity.ModifiedOn, "the undo must bump the modification timestamp, too")
}

// TestOwnerRebuildRecalculatesIgnoresTheUpdatePass verifies the difference between the two "will the owner recalculate
// this entity on its own" checks. Marking a sheet as modified does nothing at all while the sheet is already inside an
// update pass, so an edit that left the recalculation to it then would drop it rather than defer it; rebuilding is
// never suppressed, so an edit that rebuilds can always leave the recalculation to the rebuild. Anything that isn't
// the sheet holding the entity leaves it to the caller either way.
func TestOwnerRebuildRecalculatesIgnoresTheUpdatePass(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()

	c.True(ownerRebuildRecalculates(sheet, entity), "a sheet recalculates its own entity when it is rebuilt")
	c.True(ownerRebuildRecalculates(sheet.Traits, entity), "a panel within the sheet resolves to the sheet")

	sheet.awaitingUpdate = true
	c.False(ownerRecalculates(sheet, entity),
		"a sheet already inside an update pass won't recalculate when it is marked as modified")
	c.True(ownerRebuildRecalculates(sheet, entity), "a rebuild recalculates even inside an update pass")
	sheet.awaitingUpdate = false

	c.False(ownerRebuildRecalculates(sheet, gurps.NewEntity()), "another entity isn't the sheet's to recalculate")
	c.False(ownerRebuildRecalculates(sheet, nil), "there is no entity to recalculate")
	c.False(ownerRebuildRecalculates(nil, entity), "there is no owner to do the recalculation")
	c.False(ownerRebuildRecalculates(NewTemplate("test"+gurps.TemplatesExt, gurps.NewTemplate()), entity),
		"a template doesn't recalculate an entity when it is rebuilt")
}
