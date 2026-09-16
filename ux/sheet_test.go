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
	"strconv"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// TestMarkModifiedRecalculates checks that telling the sheet something changed brings the entity's derived state up to
// date, since everything MarkModified goes on to do displays that state. It used to be refreshed only as a side effect
// of the tab asking whether the sheet had unsaved changes, which recalculated the entity on its way to hashing it;
// hashing no longer does that, since recalculating rewrites part of what gets saved.
func TestMarkModifiedRecalculates(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	skill := gurps.NewSkill(entity, nil, false)
	skill.Name = "Brawling"
	skill.Points = fxp.Four
	entity.Skills = append(entity.Skills, skill)
	entity.Recalculate()
	before := skill.LevelData.Level

	entity.Attributes.Set["dx"].Adjustment = fxp.Four
	sheet.MarkModified(nil)

	c.Equal(before+fxp.Four, skill.LevelData.Level, "the skill level must reflect the raised attribute")
}

// TestNewItemCommandUsesTheLiveList checks that the "New Trait" command adds its item to the list the user is looking
// at, even after the sheet has had to replace that list. A list can only change its set of columns by being built anew
// -- which the arrival of the first switchable feature forces, since it brings the switch column in -- so a command
// that captured the list when the sheet was created would afterwards be creating items in an orphan: the model would
// gain the trait, but an orphaned table can't find the undo manager, so the insertion wouldn't be undoable and the
// user's next undo would silently take back the edit before it instead, and the new row would be neither selected nor
// scrolled into view in the list that is actually on screen.
func TestNewItemCommandUsesTheLiveList(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.Traits = nil
	sheet.Rebuild(true)
	stale := sheet.Traits
	mgr := sheet.UndoManager()
	c.NotNil(mgr, "the sheet must have an undo manager")
	c.False(mgr.CanUndo(), "nothing has been done yet")

	// Giving the sheet its first switchable feature brings in the switch column, which replaces the traits list.
	claws := newSwitchableTrait(entity, "Claws")
	entity.Traits = []*gurps.Trait{claws}
	sheet.Rebuild(true)
	c.NotEqual(stale, sheet.Traits, "gaining the switch column must have replaced the traits list")

	sheet.AsPanel().PerformCmd(nil, NewTraitItemID)
	c.Equal(2, len(entity.Traits), "the command must have added a trait to the entity")
	added := entity.Traits[1]
	c.NotEqual(claws.ID(), added.ID(), "the added trait must be the new one, not the one that was already there")
	c.Equal(2, sheet.Traits.Table.RootRowCount(), "the new row must be in the list that is on screen")
	c.True(sheet.Traits.Table.CopySelectionMap()[added.ID()],
		"the new row must be selected in the list that is on screen")

	c.True(mgr.CanUndo(), "creating an item must be undoable")
	mgr.Undo()
	c.Equal(1, len(entity.Traits), "undo must take the new trait back out of the entity")
	c.Equal(claws.ID(), entity.Traits[0].ID(), "undo must leave the trait that was already there alone")
	c.Equal(1, sheet.Traits.Table.RootRowCount(), "undo must take the row back out of the list that is on screen")
}

// TestSheetListsFollowTheCanonicalBlockOrder checks that the sheet's lists come back in the canonical block order, that
// the lookup by key the layout editor uses yields the same list for each, and that it yields nothing for a key that
// isn't a list's.
func TestSheetListsFollowTheCanonicalBlockOrder(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	expected := []unison.Paneler{
		sheet.Reactions,
		sheet.ConditionalModifiers,
		sheet.MeleeWeapons,
		sheet.RangedWeapons,
		sheet.Traits,
		sheet.Skills,
		sheet.Spells,
		sheet.CarriedEquipment,
		sheet.OtherEquipment,
		sheet.Notes,
	}
	lists := sheet.lists()
	c.Equal(len(expected), len(lists), "every list is present")
	i := 0
	for _, key := range gurps.AllBlockKeys {
		if !gurps.IsListBlockKey(key) {
			c.Nil(sheet.list(key), "%s is not a list", key)
			continue
		}
		c.Equal(expected[i].AsPanel(), lists[i].AsPanel(), "the list for %s is in its canonical place", key)
		c.Equal(expected[i].AsPanel(), sheet.blockPanel(key).AsPanel(), "and is what the key maps to")
		i++
	}
}

// TestSheetRebuildCarriesTheSelectionToAReplacedList checks that a row selected in a list is still selected after a
// rebuild that had to replace the list, since the selection is put back into the list that is on screen rather than the
// orphan it replaced.
func TestSheetRebuildCarriesTheSelectionToAReplacedList(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	plain := gurps.NewTrait(entity, nil, false)
	plain.Name = "Plain"
	entity.Traits = []*gurps.Trait{plain}
	sheet.Rebuild(true)
	stale := sheet.Traits
	stale.Table.SetSelectionMap(map[tid.TID]bool{plain.ID(): true})

	// Giving the sheet its first switchable feature brings in the switch column, which replaces the traits list.
	entity.Traits = append(entity.Traits, newSwitchableTrait(entity, "Claws"))
	sheet.Rebuild(true)
	c.NotEqual(stale, sheet.Traits, "gaining the switch column must have replaced the traits list")
	c.True(sheet.Traits.Table.CopySelectionMap()[plain.ID()],
		"the selection must have followed the row into the replacement")
}

// newContradictoryTraitRing returns two traits that each require the other's absence, so that a sheet enforcing trait
// prerequisites finds that its data never settles while both are enabled.
func newContradictoryTraitRing(entity *gurps.Entity) []*gurps.Trait {
	ring := make([]*gurps.Trait, 2)
	for i := range ring {
		ring[i] = gurps.NewTrait(entity, nil, false)
		ring[i].Name = "Ring " + strconv.Itoa(i+1)
		list := gurps.NewPrereqList()
		prereq := gurps.NewTraitPrereq()
		prereq.Parent = list
		prereq.Has = false
		prereq.NameCriteria.Qualifier = "Ring " + strconv.Itoa((i+1)%len(ring)+1)
		list.Prereqs = append(list.Prereqs, prereq)
		ring[i].Prereq = list
	}
	return ring
}

// toolbarColumns returns how many children the toolbar's layout was finished for; see finishToolbarLayout.
func toolbarColumns(c check.Checker, toolbar *unison.Panel) int {
	c.Helper()
	layout, ok := toolbar.Layout().(*unison.FlexLayout)
	c.True(ok, "the toolbar has a flex layout")
	return layout.Columns
}

// TestSheetToolbarNotesUnsettledData verifies that the sheet's toolbar shows the notice that its data never settles
// while the entity says so (see gurps.Entity.Unsettled), whether the sheet was opened that way or edited into it, and
// takes the notice down again once the data settles, with the toolbar's layout kept in step with the children it has
// either way.
func TestSheetToolbarNotesUnsettledData(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	c.Nil(sheet.unsettledNotice.Parent(), "a new sheet's data settles, so the notice is not shown")
	c.Equal(len(sheet.toolbar.Children()), toolbarColumns(c, sheet.toolbar), "and the toolbar lays out its children")

	entity.SheetSettings.EnforceTraitPrereqs = true
	ring := newContradictoryTraitRing(entity)
	entity.Traits = ring
	sheet.Rebuild(true)
	c.True(entity.Unsettled(), "precondition: the ring leaves the data unable to settle")
	c.Equal(sheet.toolbar.AsPanel(), sheet.unsettledNotice.Parent(), "the notice is shown once the data never settles")
	c.Equal(len(sheet.toolbar.Children()), toolbarColumns(c, sheet.toolbar), "and the toolbar lays out the notice too")
	c.Equal(sheet.unsettledNotice.AsPanel(), sheet.toolbar.Children()[len(sheet.toolbar.Children())-1],
		"at the end of the toolbar")

	ring[1].Disabled = true
	sheet.MarkModified(nil)
	c.False(entity.Unsettled(), "precondition: disabling one of the ring settles the data")
	c.Nil(sheet.unsettledNotice.Parent(), "the notice is taken down once the data settles")
	c.Equal(len(sheet.toolbar.Children()), toolbarColumns(c, sheet.toolbar), "and the toolbar lays out what is left")

	// A sheet whose data never settles when it is opened shows the notice from the start.
	opened := gurps.NewEntity()
	opened.SheetSettings.EnforceTraitPrereqs = true
	opened.Traits = newContradictoryTraitRing(opened)
	opened.Recalculate()
	c.True(opened.Unsettled(), "precondition: the ring leaves the data unable to settle")
	openedSheet := NewSheet("opened"+gurps.SheetExt, opened)
	c.Equal(openedSheet.toolbar.AsPanel(), openedSheet.unsettledNotice.Parent(),
		"a sheet opened with data that never settles shows the notice from the start")
	c.Equal(len(openedSheet.toolbar.Children()), toolbarColumns(c, openedSheet.toolbar),
		"and its toolbar lays out the notice too")
}
