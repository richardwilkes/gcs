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
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/layoutnode"
	"github.com/richardwilkes/gcs/v5/model/gurps/gurpstest"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	uncheck "github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/role"
	"github.com/richardwilkes/unison/enums/side"
)

// TestNewItemCommandUsesTheLiveList checks that the "New Trait" command adds its item to the list on screen even after
// the sheet has replaced that list, as gaining the switch column forces. A command that captured the original list
// would create items in an orphan whose table can't find the undo manager, so the insertion wouldn't be undoable and
// the new row would be neither selected nor scrolled into view.
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

// TestSheetReroutesDropsToTheListOfTheirKind verifies that an item of each kind the sheet takes, dropped anywhere on
// it, is rerouted to the list of that kind.
func TestSheetReroutesDropsToTheListOfTheirKind(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	for _, one := range []struct {
		name string
		key  *uti.DataType
		list unison.Paneler
	}{
		{"equipment", equipmentDragKey, sheet.CarriedEquipment},
		{"a skill", skillDragKey, sheet.Skills},
		{"a spell", spellDragKey, sheet.Spells},
		{"a trait", traitDragKey, sheet.Traits},
		{"a note", noteDragKey, sheet.Notes},
	} {
		c.True(unison.AncestorIs(sheet.keyToPanel(one.key), one.list), "%s dropped on the sheet must go to its list",
			one.name)
	}
}

// TestSheetClearPortraitCommandIsUndoable verifies that the Clear Portrait command is on offer only while the sheet has
// a portrait, takes the portrait away, and can be undone and redone.
func TestSheetClearPortraitCommandIsUndoable(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	profile := &sheet.Entity().Profile
	p := sheet.AsPanel()
	c.False(p.CanPerformCmd(nil, ClearPortraitItemID), "there is no portrait to clear")
	var buf bytes.Buffer
	c.NoError(png.Encode(&buf, image.NewNRGBA(image.Rect(0, 0, 1, 1))))
	portrait := buf.Bytes()
	profile.SetPortraitData(portrait)
	c.True(p.CanPerformCmd(nil, ClearPortraitItemID), "a portrait can be cleared")

	p.PerformCmd(nil, ClearPortraitItemID)
	c.Equal(0, len(profile.PortraitData), "the command takes the portrait away")
	mgr := sheet.UndoManager()
	c.True(mgr.CanUndo(), "clearing the portrait must be undoable")
	mgr.Undo()
	c.Equal(portrait, profile.PortraitData, "undo puts the portrait back")
	mgr.Redo()
	c.Equal(0, len(profile.PortraitData), "redo takes it away again")
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

// TestSheetToolbarNotesUnsettledData verifies that the sheet's toolbar shows the notice that its data never settles
// while the entity says so (see gurps.Entity.Unsettled), whether the sheet was opened that way or edited into it, and
// takes the notice down again once the data settles. The notice sits at the end of the toolbar, after the search field
// that is given what is left of the row, level with the rest of the row.
func TestSheetToolbarNotesUnsettledData(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	c.Nil(sheet.unsettledNotice.Parent(), "a new sheet's data settles, so the notice is not shown")
	layout, ok := sheet.toolbar.Layout().(*flowFillLayout)
	c.True(ok, "the toolbar wraps its children")
	c.True(ok && layout.Fill == sheet.search.field, "with the search field given what is left of its row")

	entity.SheetSettings.EnforceTraitPrereqs = true
	ring := newContradictoryTraitRing(entity)
	entity.Traits = ring
	sheet.Rebuild(true)
	c.True(entity.Unsettled(), "precondition: the ring leaves the data unable to settle")
	c.Equal(sheet.toolbar.AsPanel(), sheet.unsettledNotice.Parent(), "the notice is shown once the data never settles")
	c.Equal(sheet.unsettledNotice.AsPanel(), sheet.toolbar.Children()[len(sheet.toolbar.Children())-1],
		"at the end of the toolbar")
	c.Equal(align.Middle, sheet.unsettledNotice.LayoutData(), "level with the rest of its row")

	ring[1].SetDisabled(true)
	sheet.MarkModified(nil)
	c.False(entity.Unsettled(), "precondition: disabling one of the ring settles the data")
	c.Nil(sheet.unsettledNotice.Parent(), "the notice is taken down once the data settles")

	// A sheet whose data never settles when it is opened shows the notice from the start.
	opened := gurps.NewEntity()
	opened.SheetSettings.EnforceTraitPrereqs = true
	opened.Traits = newContradictoryTraitRing(opened)
	opened.Recalculate()
	c.True(opened.Unsettled(), "precondition: the ring leaves the data unable to settle")
	openedSheet := NewSheet("opened"+gurps.SheetExt, opened)
	c.Equal(openedSheet.toolbar.AsPanel(), openedSheet.unsettledNotice.Parent(),
		"a sheet opened with data that never settles shows the notice from the start")
	c.Equal(openedSheet.unsettledNotice.AsPanel(), openedSheet.toolbar.Children()[len(openedSheet.toolbar.Children())-1],
		"at the end of the toolbar")
}

// newWeightScriptedSheet returns a sheet whose ST a script derives from the character's weight, a point per 15 lb, so
// that an edit to the weight -- a Description block field, flagged SkipDeepSync -- changes what the rest of the sheet
// shows, along with the weight field, to be named as the source of such an edit. The character starts at 150 lb, so
// ST 10, and carries 30 lb, so is lightly encumbered. The sheet enforces trait prerequisites.
func newWeightScriptedSheet(t *testing.T) (sheet *Sheet, weightField *WeightField) {
	t.Helper()
	c := check.New(t)
	sheet = newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.SheetSettings.EnforceTraitPrereqs = true
	entity.SheetSettings.Attributes.Set[gurps.StrengthID].Base = "entity.weightInPounds / 15"
	entity.Profile.Weight = fxp.WeightFromInteger(150, fxp.Pound)
	load := gurps.NewEquipment(entity, nil, false)
	load.Name = "Load"
	load.BaseWeight = "30 lb"
	entity.CarriedEquipment = append(entity.CarriedEquipment, load)
	sheet.Rebuild(true)
	panel := sheet.targetMgr.Find(descriptionPanelWeightFieldRefKey)
	c.NotNil(panel, "the sheet must have a weight field")
	weightField, ok := panel.Self.(*WeightField)
	c.True(ok, "the weight field must be a WeightField")
	_, skip := weightField.ClientData()[SkipDeepSync]
	c.True(skip, "the weight field must be flagged SkipDeepSync, or the tests using it check nothing")
	return sheet, weightField
}

// TestProfileEditThatChangesTheDerivedStateSyncsTheLists checks that an edit made through a field flagged
// SkipDeepSync is not spared the deep sync when the recalculation it prompts changes which rows a list has: here a
// script prerequisite reading the weight enables a trait, whose weapon the melee weapons list then has to show.
func TestProfileEditThatChangesTheDerivedStateSyncsTheLists(t *testing.T) {
	c := check.New(t)
	sheet, weightField := newWeightScriptedSheet(t)
	entity := sheet.Entity()
	for _, one := range []struct{ name, script string }{
		{name: "Club"},
		{name: "Heavy", script: `entity.weightInPounds > 300 ? "" : "Weighs too little"`},
	} {
		trait := gurps.NewTrait(entity, nil, false)
		trait.Name = one.name
		weapon := gurps.NewWeapon(trait, true)
		weapon.Usage = one.name
		trait.Weapons = append(trait.Weapons, weapon)
		if one.script != "" {
			trait.Prereq = gurps.NewPrereqList()
			script := gurps.NewScriptPrereq()
			script.Parent = trait.Prereq
			script.Script = one.script
			trait.Prereq.Prereqs = append(trait.Prereq.Prereqs, script)
		}
		entity.Traits = append(entity.Traits, trait)
	}
	sheet.Rebuild(true)
	heavy := entity.Traits[len(entity.Traits)-1]
	c.False(heavy.Enabled(), "the heavy trait must start disabled, or the check below means nothing")
	// The table only has the rows it was last synced to; RootRowCount would ask the model.
	before := sheet.MeleeWeapons.Table.LastRowIndex() + 1

	entity.Profile.Weight = fxp.WeightFromInteger(450, fxp.Pound)
	sheet.MarkModified(weightField)

	c.True(heavy.Enabled(), "the heavy trait must now be enabled, or the check below means nothing")
	c.Equal(before+1, sheet.MeleeWeapons.Table.LastRowIndex()+1, "the list shows the weapon of the trait the edit enabled")
}

// TestEditThatChangesWhetherADerivedListIsEmptyLaysOutThePage checks that an edit that gives one of the lists derived
// from the character its first row puts that list on the page, and that one taking its last row away takes it off:
// here a trait whose ST prerequisite the edit satisfies, and the melee weapon it grants.
func TestEditThatChangesWhetherADerivedListIsEmptyLaysOutThePage(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.SheetSettings.EnforceTraitPrereqs = true
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Strong Kick"
	weapon := gurps.NewWeapon(trait, true)
	weapon.Usage = "Kick"
	trait.Weapons = append(trait.Weapons, weapon)
	trait.Prereq = gurps.NewPrereqList()
	prereq := gurps.NewAttributePrereq(entity)
	prereq.Parent = trait.Prereq
	prereq.QualifierCriteria.Qualifier = fxp.Twelve
	trait.Prereq.Prereqs = append(trait.Prereq.Prereqs, prereq)
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)
	c.False(trait.Enabled(), "the trait must start disabled, or the checks below mean nothing")
	c.True(sheet.MeleeWeapons.Parent() == nil, "the empty melee list must start off the page, or the checks below mean nothing")

	entity.Attributes.Set[gurps.StrengthID].Adjustment = fxp.Two
	sheet.MarkModified(nil)
	c.True(trait.Enabled(), "the trait must now be enabled, or the checks below mean nothing")
	c.True(sheet.MeleeWeapons.Parent() != nil, "the melee list joins the page with the weapon the edit enabled")
	c.Equal(1, sheet.MeleeWeapons.Table.LastRowIndex()+1, "and shows it")

	entity.Attributes.Set[gurps.StrengthID].Adjustment = 0
	sheet.MarkModified(nil)
	c.False(trait.Enabled(), "the trait must be disabled again, or the check below means nothing")
	c.True(sheet.MeleeWeapons.Parent() == nil, "the melee list leaves the page once the edit takes its only weapon away")
}

// TestProfileEditReadByAScriptSyncsTheLists checks that an edit made through a field flagged SkipDeepSync is not spared
// the deep sync when a script reads what the field holds, even though the derived state is left as it was: here an
// equipment weight derived from the character's weight, whose total the carried equipment list's header shows. The
// blocks are brought up to date too: the ST field, in the secondary attributes block since a scripted base makes ST a
// secondary attribute, and the encumbrance block, whose current level is found only when it is synced.
func TestProfileEditReadByAScriptSyncsTheLists(t *testing.T) {
	c := check.New(t)
	sheet, weightField := newWeightScriptedSheet(t)
	entity := sheet.Entity()
	pack := gurps.NewEquipment(entity, nil, false)
	pack.Name = "Pack"
	pack.BaseWeight = "entity.weightInPounds / 10"
	entity.CarriedEquipment = append(entity.CarriedEquipment, pack)
	sheet.Rebuild(true)
	headerText := func() string {
		i := sheet.CarriedEquipment.Table.ColumnIndexForID(gurps.EquipmentDescriptionColumn)
		header, ok := sheet.CarriedEquipment.tableHeader.ColumnHeaders[i].(*PageTableColumnHeader[*gurps.Equipment])
		c.True(ok, "the description column must have a page header")
		return header.Text.String()
	}
	c.Equal("Carried Equipment (45 lb; $0)", headerText(), "the pack must start at 15 lb, or the checks below mean nothing")
	attrPanel, ok := sheet.blocks[gurps.BlockSecondaryAttributesKey].(*AttrPanel)
	c.True(ok, "the secondary attributes block must be an AttrPanel")
	stField, ok := attrPanel.valueFields[gurps.StrengthID].(*IntegerField)
	c.True(ok, "the ST field must be an IntegerField")
	encPanel, ok := sheet.blocks[gurps.BlockEncumbranceKey].(*EncumbrancePanel)
	c.True(ok, "the encumbrance block must be an EncumbrancePanel")
	c.Equal("10", stField.Text(), "ST must start at 10, or the checks below mean nothing")
	c.Equal(int(encumbrance.Medium), encPanel.current,
		"the 45 lb carried must start the character at medium encumbrance, or the checks below mean nothing")

	entity.Profile.Weight = fxp.WeightFromInteger(450, fxp.Pound)
	sheet.MarkModified(weightField)

	c.False(entity.DerivedStateChanged(), "the edit must leave the derived state as it was, or the checks below mean nothing")
	c.Equal("Carried Equipment (75 lb; $0)", headerText(), "the header shows the total the edit changed")
	c.Equal("30", stField.Text(), "the ST field shows the recalculated ST")
	c.Equal(int(encumbrance.No), encPanel.current, "the encumbrance block shows the recalculated level")
}

// TestProfileEditOnceNoScriptReadsSparesTheLists checks that an edit made through a field flagged SkipDeepSync is
// spared the syncing of the lists once no script reads the profile any longer, although one did: the reads are
// forgotten each time the sheet derives everything it shows afresh, not kept for as long as the sheet is open.
func TestProfileEditOnceNoScriptReadsSparesTheLists(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	pack := gurps.NewEquipment(entity, nil, false)
	pack.Name = "Pack"
	pack.BaseWeight = "entity.weightInPounds / 10"
	entity.CarriedEquipment = append(entity.CarriedEquipment, pack)
	sheet.Rebuild(true)
	c.True(entity.ProfileReadByScript(), "the pack's weight script must have read the profile, or the checks below mean nothing")
	panel := sheet.targetMgr.Find(identityPanelNameFieldRefKey)
	c.NotNil(panel, "the sheet must have a name field")
	nameField, ok := panel.Self.(*StringField)
	c.True(ok, "the name field must be a StringField")

	pack.BaseWeight = "15 lb"
	sheet.MarkModified(nil)
	c.False(entity.ProfileReadByScript(), "no script reads the profile once the weight is fixed")

	entity.Profile.Name = "Someone"
	sheet.MarkModified(nameField)

	c.False(entity.DerivedStateChanged(), "the edit must leave the derived state as it was, or the check below means nothing")
	c.NotEqual(currentDisplayStamp(), sheet.CarriedEquipment.syncedAtStamp(), "the list is not synced")
}

// TestProfileEditReadByAScriptOnlyWhenShownSyncsTheLists checks that edits made through a field flagged SkipDeepSync
// keep syncing the lists while a script that only the lists run, never the recalculation, reads what the field holds:
// here an equipment value derived from the character's weight, whose total the carried equipment list's header shows.
// The reads are forgotten before each deep sync, not before each recalculation, so that the sync's own reads count.
func TestProfileEditReadByAScriptOnlyWhenShownSyncsTheLists(t *testing.T) {
	c := check.New(t)
	sheet, weightField := newWeightScriptedSheet(t)
	entity := sheet.Entity()
	entity.SheetSettings.Attributes.Set[gurps.StrengthID].Base = "10"
	gem := gurps.NewEquipment(entity, nil, false)
	gem.Name = "Gem"
	gem.BaseValue = "entity.weightInPounds / 10"
	entity.CarriedEquipment = append(entity.CarriedEquipment, gem)
	sheet.Rebuild(true)
	headerText := func() string {
		i := sheet.CarriedEquipment.Table.ColumnIndexForID(gurps.EquipmentDescriptionColumn)
		header, ok := sheet.CarriedEquipment.tableHeader.ColumnHeaders[i].(*PageTableColumnHeader[*gurps.Equipment])
		c.True(ok, "the description column must have a page header")
		return header.Text.String()
	}
	c.Equal("Carried Equipment (30 lb; $15)", headerText(), "the gem must start at $15, or the checks below mean nothing")

	for i, weight := range []int{300, 450} {
		entity.Profile.Weight = fxp.WeightFromInteger(weight, fxp.Pound)
		sheet.MarkModified(weightField)
		c.False(entity.DerivedStateChanged(), "edit %d must leave the derived state as it was, or the check below means nothing", i+1)
		c.Equal(fmt.Sprintf("Carried Equipment (30 lb; $%d)", weight/10), headerText(), "the header shows the total edit %d changed", i+1)
	}
}

// TestSheetPaginatedToggle verifies that a sheet starts out showing pages or not as the general setting says, that
// togglePaginated flips the mode and latches the toolbar's pages button to match -- selected in its group with its base
// shown while on, unselected with its base hidden while off -- and that the Show as Pages menu item is checked while
// the sheet shows pages and unchecked when no sheet has the focus. Starting the mode from anything but the setting,
// leaving the button alone when toggling, or not setting the menu item's check state fails it.
func TestSheetPaginatedToggle(t *testing.T) {
	c := check.New(t)
	item := unison.NewInWindowMenuFactory().NewItem(ShowAsPagesItemID, "", unison.KeyBinding{}, nil, nil)
	for _, initial := range []bool{true, false} {
		uxtest.SwapForTest(t, &gurps.GlobalSettings().General.InitialSheetPaginated, initial)
		sheet := newTestSheetForTemplate(t)
		for i, want := range []bool{initial, !initial, initial} {
			if i != 0 {
				sheet.togglePaginated()
			}
			c.Equal(want, sheet.paginated, "initial %v, step %d: the mode", initial, i)
			c.Equal(want, sheet.pagesButtonGroup.Selected(sheet.pagesButton),
				"initial %v, step %d: the button is selected while on", initial, i)
			c.Equal(!want, sheet.pagesButton.HideBase, "initial %v, step %d: the button shows its base while on",
				initial, i)
			c.True(sheet.canTogglePaginated(item), "initial %v, step %d: the command is enabled", initial, i)
			c.Equal(uncheck.FromBool(want), item.CheckState(), "initial %v, step %d: the menu item's check", initial, i)
		}
	}
	item.SetCheckState(uncheck.On)
	c.False(showAsPagesAction.Enabled(item), "the action is disabled while no sheet has the focus")
	c.Equal(uncheck.Off, item.CheckState(), "and its menu item is unchecked")

	// The sheet is shown continuously while its layout is edited, whatever the choice, so the toggle is withheld then.
	uxtest.SwapForTest(t, &gurps.GlobalSettings().General.InitialSheetPaginated, true)
	sheet := newTestSheetForTemplate(t)
	sheet.toggleLayoutEditing()
	c.False(sheet.canTogglePaginated(item), "the command is withheld while the layout is being edited")
	c.False(sheet.pagesButton.Enabled(), "and so is the button")
	sheet.toggleLayoutEditing()
	c.True(sheet.canTogglePaginated(item), "the command comes back once the editing is over")
	c.True(sheet.pagesButton.Enabled(), "and so does the button")
}

// A new sheet puts the focus on its name field, not on the portrait laid out ahead of it, where typing would go
// nowhere, nor on the Identity heading, which can take the focus while a screen reader is listening. The same choice is
// made whenever the sheet's focus has to be found afresh, as when the layout editor closes.
func TestSheetOpensWithTheFocusOnItsName(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	focusKey := func() (key string) {
		screen.Do(func() {
			if focus := wnd.Focus(); focus != nil {
				key = focus.RefKey
			}
		})
		return key
	}
	c.Equal(identityPanelNameFieldRefKey, focusKey(), "the sheet opens with the focus on its name")
	var portrait *PortraitPanel
	screen.Do(func() { portrait, _ = uxtest.FirstPanelOfType[*PortraitPanel](sheet.AsPanel()) })
	if portrait == nil {
		t.Fatal("the character sheet must show the portrait")
	}
	var portraitFirst bool
	screen.Do(func() {
		portraitFirst = firstFocusableInSubtree(sheet.scroll.Content().AsPanel(), nil) == portrait.AsPanel()
	})
	c.True(portraitFirst, "the portrait is the first tab stop on the page, which is what makes the choice one")
	screen.Do(func() { FocusFirstContent(sheet.toolbar, sheet.scroll.Content()) })
	c.Equal(identityPanelNameFieldRefKey, focusKey(), "the focus found afresh goes to the name as well")

	// Under Linux's rules, which a headless session follows, the headings take the focus while a screen reader is
	// listening and come ahead of the name; they are passed over too.
	if screen.AccessibilityTree(wnd) == nil {
		t.Fatal("the window must be described")
	}
	headingIndex, nameIndex := -1, -1
	screen.Do(func() {
		for i, one := range uxtest.PanelsMatching(sheet.scroll.Content().AsPanel(), (*unison.Panel).Focusable) {
			switch {
			case one.Accessibility.Role == role.Heading && headingIndex < 0:
				headingIndex = i
			case one.RefKey == identityPanelNameFieldRefKey:
				nameIndex = i
			}
		}
	})
	c.True(headingIndex >= 0 && headingIndex < nameIndex, "a heading is now a tab stop ahead of the name, but the "+
		"first heading is tab stop %d and the name is tab stop %d", headingIndex, nameIndex)
	screen.Do(func() { FocusFirstContent(sheet.toolbar, sheet.scroll.Content()) })
	c.Equal(identityPanelNameFieldRefKey, focusKey(), "the focus found afresh still goes to the name")
}

// A tab stop without a reference key, such as the portrait, a heading or a label, keeps the focus through a change that
// leaves it in place. When the change removes it, the focus goes to the name field, as for a keyed field that is gone.
func TestSheetKeepsTheFocusOnKeylessTabStops(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	focus := func() (focus *unison.Panel) {
		screen.Do(func() { focus = wnd.Focus() })
		return focus
	}
	var portrait *PortraitPanel
	screen.Do(func() { portrait, _ = uxtest.FirstPanelOfType[*PortraitPanel](sheet.AsPanel()) })
	if portrait == nil {
		t.Fatal("the character sheet must show the portrait")
	}
	screen.Do(portrait.RequestFocus)
	c.Equal(portrait.AsPanel(), focus())
	screen.Do(func() { sheet.updatePortrait(nil) })
	c.Equal(portrait.AsPanel(), focus(), "changing the portrait leaves the focus on it")
	screen.Do(func() { sheet.Rebuild(false) })
	c.Equal(portrait.AsPanel(), focus(), "a rebuild leaves the focus on it")

	// Under Linux's rules a heading takes the focus once asking for the tree has activated accessibility. The Identity
	// block never rebuilds its contents, so its heading survives Rebuild(false).
	if screen.AccessibilityTree(wnd) == nil {
		t.Fatal("the window must be described")
	}
	var heading *unison.Panel
	screen.Do(func() {
		for _, one := range uxtest.PanelsMatching(sheet.AsPanel(), func(p *unison.Panel) bool {
			return p.Accessibility.Role == role.Heading && p.Accessibility.Name == "Identity"
		}) {
			heading = one
		}
	})
	if heading == nil {
		t.Fatal("the character sheet must show the Identity heading")
	}
	screen.Do(heading.RequestFocus)
	c.Equal(heading, focus(), "the heading takes the focus while a screen reader is listening")
	screen.Do(func() { sheet.Rebuild(false) })
	c.Equal(heading, focus(), "a rebuild leaves the focus on the heading")

	// An attribute block's rebuild replaces its rows and their read-only points fields, which are keyless tab stops
	// with static text in the Tab order (the name label is not, being the value field's caption).
	uxtest.FocusForReadingSetter(t, screen, wnd)(true)
	var attrs *AttrPanel
	var points *NonEditablePageField
	screen.Do(func() {
		for _, one := range uxtest.PanelsOfType[*AttrPanel](sheet.AsPanel()) {
			if fields := uxtest.PanelsOfType[*NonEditablePageField](one.AsPanel()); len(fields) > 0 {
				attrs = one
				points = fields[0]
				break
			}
		}
	})
	if points == nil {
		t.Fatal("the character sheet must show an attribute block with a points field")
	}
	screen.Do(points.RequestFocus)
	c.Equal(points.AsPanel(), focus(), "the points field takes the focus with static text in the Tab order")
	screen.Do(func() { attrs.rebuild(gurps.SheetSettingsFor(sheet.Entity()).Attributes) })
	var key string
	var stillThere bool
	screen.Do(func() {
		stillThere = points.Window() == wnd
		if f := wnd.Focus(); f != nil {
			key = f.RefKey
		}
	})
	c.False(stillThere, "the points field is gone with the rows")
	c.Equal(identityPanelNameFieldRefKey, key, "the focus goes to the name field")
}

// The benchmarks below measure what a character sheet with large lists costs to update, rebuild, draw and
// scroll, in a headless session. Run them with:
//
//	GOEXPERIMENT=simd go test ./ux -run '^$' -bench '^Benchmark(Sheet|PageList|NameablesPanel)' -benchmem -benchtime 2s \
//	    -count 6 -timeout 90m | tee baseline.txt
//
// For a quick look, add GCS_BENCH_SIZE=0.25 and -benchtime 5x. GCS_BENCH_FILE names a real sheet to measure instead
// of the generated one (see benchEntity); a benchmark of a list that sheet leaves off the page is skipped. Compare two
// runs with:
//
//	go run golang.org/x/perf/cmd/benchstat@latest before.txt after.txt
//
// Besides ns/op, which is the end-to-end time including the headless session's round trips, each benchmark reports
// ui-ns/op, the time the operation itself took on the UI thread, and draw-ns/op, the time the redraw it provoked
// took. Each benchmark starts one headless session and must therefore never run in parallel. The sheet is shown as one
// continuous page, except where a benchmark measures each of sheetViewModes in turn.

// benchFileEnvVar names the environment variable that, when set to the path of a character sheet, has the benchmarks
// measure that sheet instead of a generated one.
const benchFileEnvVar = "GCS_BENCH_FILE"

// sheetBench is a sheet open in a headless session, ready to be measured.
type sheetBench struct {
	screen *unison.HeadlessScreen
	wnd    *unison.Window
	sheet  *Sheet
	heapMB float64 // The heap in use once the sheet was open, which each measurement reports as heap-MB.
	draws  int     // How many times the window has drawn, counted on the UI thread; read it once the session is quiet.
}

// sheetViewModes are the ways a sheet can be shown, which the benchmarks of rebuilding and drawing measure in turn.
var sheetViewModes = []struct {
	name      string
	paginated bool
}{{"continuous", false}, {"pages", true}}

// benchEntity returns the character to measure: the sheet named by benchFileEnvVar, or a generated one of the size
// gurpstest.LargeEntitySizeFromEnv reports.
func benchEntity(b *testing.B) (entity *gurps.Entity, generated bool) {
	b.Helper()
	if path := os.Getenv(benchFileEnvVar); path != "" {
		e, err := gurps.NewEntityFromFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
		if err != nil {
			b.Fatalf("unable to load %s: %v", path, err)
		}
		return e, false
	}
	size, err := gurpstest.LargeEntitySizeFromEnv()
	if err != nil {
		b.Fatal(err)
	}
	return gurpstest.NewLargeEntity(&size), true
}

// newSheetBench starts a headless session, opens the character to measure in it as a sheet, checks that every list
// is on the page when the character was generated for that purpose, and performs one untimed update so that the
// first measurement doesn't pay for anything lazily built. It also records the heap in use once the sheet is open and
// starts counting the window's draws.
func newSheetBench(b *testing.B) *sheetBench {
	b.Helper()
	c := check.New(b)
	screen, wnd := uxtest.StartHeadlessWorkspace(b, c)
	uxtest.SwapForTest(b, &gurps.GlobalSettings().General.InitialSheetPaginated, false)
	sb := &sheetBench{screen: screen, wnd: wnd}
	entity, generated := benchEntity(b)
	var sheet *Sheet
	screen.Do(func() {
		content := wnd.Content()
		prior := content.DrawOverCallback
		content.DrawOverCallback = func(gc *unison.Canvas, rect geom.Rect) {
			if prior != nil {
				prior(gc, rect)
			}
			sb.draws++
		}
		sheet = NewSheet("bench"+gurps.SheetExt, entity)
		DisplayNewDockable(sheet)
	})
	if generated {
		for _, key := range gurps.AllBlockKeys {
			if !gurps.IsListBlockKey(key) {
				continue
			}
			list := sheet.list(key)
			if xreflect.IsNil(list) || !listAttachedToSheet(sheet, list) {
				b.Fatalf("the %s list is not on the page", key)
			}
		}
	}
	screen.Do(func() { sheet.MarkModified(sheet) })
	// Registered after the session's own cleanup, so that it runs first: the sheet is closed without a save prompt,
	// which a modified sheet would otherwise put up.
	b.Cleanup(func() {
		screen.Do(func() { sheet.markUnmodified() })
		uxtest.CloseEditorWithoutPrompt(b, screen, sheet)
	})
	runtime.GC()
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	sb.sheet = sheet
	sb.heapMB = float64(stats.HeapAlloc) / (1 << 20)
	return sb
}

// showAs shows the sheet as pages or as one continuous page.
func (sb *sheetBench) showAs(paginated bool) {
	sb.screen.Do(func() {
		if sb.sheet.paginated != paginated {
			sb.sheet.togglePaginated()
		}
	})
}

// reportHeap reports the heap in use once the sheet was open. It has to be reported after the timed loop, since
// resetting the timer discards the metrics reported before it.
func (sb *sheetBench) reportHeap(b *testing.B) {
	b.ReportMetric(sb.heapMB, "heap-MB")
}

// drawSince returns the time of the window's last draw if it has drawn since it had drawn seen times, or 0 if it has
// not, and updates seen to the current count. Call it once the session is quiet.
func (sb *sheetBench) drawSince(seen *int) time.Duration {
	if sb.draws == *seen {
		return 0
	}
	*seen = sb.draws
	return sb.wnd.LastDrawDuration()
}

// run measures fn, which runs on the UI thread once per iteration, reporting the time it took there as ui-ns/op and
// the time of the redraw it provoked, if any, as draw-ns/op alongside the usual end-to-end ns/op.
func (sb *sheetBench) run(b *testing.B, fn func()) {
	b.Helper()
	var ui, draw time.Duration
	var seen int
	sb.screen.Do(func() { seen = sb.draws })
	n := 0
	b.ResetTimer()
	for b.Loop() {
		sb.screen.Do(func() {
			// Do waits for the session to go quiet, so a draw since the last look is the redraw the previous iteration
			// provoked.
			draw += sb.drawSince(&seen)
			start := time.Now()
			fn()
			ui += time.Since(start)
		})
		n++
	}
	b.StopTimer()
	sb.screen.Do(func() { draw += sb.drawSince(&seen) })
	b.ReportMetric(float64(ui.Nanoseconds())/float64(n), "ui-ns/op")
	b.ReportMetric(float64(draw.Nanoseconds())/float64(n), "draw-ns/op")
	sb.reportHeap(b)
}

// scrollTo brings the panel into view, so that drawing and scrolling measure the part of the page that holds it. It
// skips the benchmark when the panel isn't on the page, as is the case on a sheet loaded through benchFileEnvVar for an
// empty weapons list or a list its layout omits.
func (sb *sheetBench) scrollTo(b *testing.B, p unison.Paneler) {
	b.Helper()
	var onPage bool
	sb.screen.Do(func() {
		if onPage = listAttachedToSheet(sb.sheet, p); onPage {
			p.AsPanel().ScrollIntoView()
		}
	})
	if !onPage {
		b.Skip("not on the page")
	}
}

func BenchmarkSheetMarkModified(b *testing.B) {
	sb := newSheetBench(b)
	sb.run(b, func() { sb.sheet.MarkModified(sb.sheet) })
}

func BenchmarkSheetRebuild(b *testing.B) {
	sb := newSheetBench(b)
	for _, mode := range sheetViewModes {
		for _, one := range []struct {
			name string
			full bool
		}{{"full", true}, {"partial", false}} {
			b.Run(mode.name+"/"+one.name, func(b *testing.B) {
				sb.showAs(mode.paginated)
				sb.run(b, func() { sb.sheet.Rebuild(one.full) })
			})
		}
	}
}

func BenchmarkSheetRedraw(b *testing.B) {
	sb := newSheetBench(b)
	for _, mode := range sheetViewModes {
		for _, one := range []struct {
			name  string
			panel unison.Paneler
		}{{"top", sb.sheet.page}, {"melee", sb.sheet.MeleeWeapons}, {"spells", sb.sheet.Spells}} {
			b.Run(mode.name+"/"+one.name, func(b *testing.B) {
				sb.showAs(mode.paginated)
				sb.scrollTo(b, one.panel)
				sb.run(b, func() {
					sb.wnd.MarkForRedraw()
					sb.wnd.FlushDrawing()
				})
			})
		}
	}
}

// BenchmarkSheetScroll measures wheel scrolling over each of the lists that are costly to draw: twenty steps down
// and twenty back up per iteration, reporting the time per step.
func BenchmarkSheetScroll(b *testing.B) {
	sb := newSheetBench(b)
	const steps = 20
	for _, one := range []struct {
		name  string
		panel unison.Paneler
	}{
		{"melee", sb.sheet.MeleeWeapons},
		{"ranged", sb.sheet.RangedWeapons},
		{"skills", sb.sheet.Skills},
		{"spells", sb.sheet.Spells},
	} {
		b.Run(one.name, func(b *testing.B) {
			sb.scrollTo(b, one.panel)
			pt := sb.screen.PanelCenter(sb.sheet.scroll)
			var draw time.Duration
			seen := sb.draws
			n := 0
			b.ResetTimer()
			for b.Loop() {
				// Wheel waits for the session to go quiet, so a draw since the last look is the one its step provoked.
				for range steps {
					sb.screen.Wheel(pt, geom.Point{Y: -3}, 0)
					draw += sb.drawSince(&seen)
				}
				for range steps {
					sb.screen.Wheel(pt, geom.Point{Y: 3}, 0)
					draw += sb.drawSince(&seen)
				}
				n++
			}
			b.StopTimer()
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(n*2*steps), "ns/step")
			b.ReportMetric(float64(draw.Nanoseconds())/float64(n*2*steps), "draw-ns/step")
			sb.reportHeap(b)
		})
	}
}

// BenchmarkPageListSync measures syncing each of the sheet's lists to its model, which is the bulk of an update.
func BenchmarkPageListSync(b *testing.B) {
	sb := newSheetBench(b)
	for _, key := range gurps.AllBlockKeys {
		if !gurps.IsListBlockKey(key) {
			continue
		}
		list := sb.sheet.list(key)
		if xreflect.IsNil(list) {
			continue
		}
		syncer, ok := list.(Syncer)
		if !ok {
			b.Fatalf("the %s list is not a Syncer", key)
		}
		b.Run(key, func(b *testing.B) {
			sb.run(b, syncer.Sync)
		})
	}
}

// BenchmarkSheetSearch measures the search refresh every update performs: with the field empty, with text in it and its
// matches showing, and with the same text but the matches hidden, which must cost next to nothing.
func BenchmarkSheetSearch(b *testing.B) {
	sb := newSheetBench(b)
	search := sb.sheet.search
	for _, one := range []struct {
		name   string
		text   string
		hidden bool
	}{{name: "empty"}, {name: "text", text: "spell"}, {name: "hidden", text: "spell", hidden: true}} {
		b.Run(one.name, func(b *testing.B) {
			sb.screen.Do(func() {
				search.field.RequestFocus()
				search.field.SetText(one.text)
				if one.hidden {
					search.suggestions.Hide()
				}
			})
			sb.run(b, search.Refresh)
		})
	}
}

// BenchmarkNameablesPanel measures the Nameables section of the editor of a trait on the sheet whose modifiers share
// and are linked to its nameables: "sync" is the check it makes after every change in the editor, and "rebuild" builds
// its rows again, as each change to them does. The section isn't in a window, so both are timed on the UI thread
// alone, with nothing drawn.
func BenchmarkNameablesPanel(b *testing.B) {
	sb := newSheetBench(b)
	var p *nameablesPanel
	sb.screen.Do(func() {
		gurps.Traverse(func(t *gurps.Trait) bool {
			if len(t.NameableLinks) == 0 {
				return false
			}
			_, content := buildEditorContent(sb.sheet, t, initTraitEditor)
			p, _ = uxtest.FirstPanelOfType[*nameablesPanel](content)
			return true
		}, false, true, sb.sheet.entity.Traits...)
	})
	if p == nil {
		b.Skip("no trait on the sheet links its modifiers' nameables")
	}
	sb.screen.Do(func() {
		if p.collapse.collapsed {
			p.collapse.toggle()
		}
	})
	b.Run("sync", func(b *testing.B) {
		sb.screen.Do(func() {
			for b.Loop() {
				p.Sync()
			}
		})
	})
	b.Run("rebuild", func(b *testing.B) {
		sb.screen.Do(func() {
			for b.Loop() {
				p.RemoveAllChildren()
				p.build()
			}
		})
	})
}

// TestSheetBenchSkipsAListNotOnThePage verifies that a benchmark of a list that a loaded sheet leaves off the page,
// such as the ranged weapons of a character without any, is skipped rather than measuring whatever part of the page is
// in view.
func TestSheetBenchSkipsAListNotOnThePage(t *testing.T) {
	c := check.New(t)
	path := filepath.Join(t.TempDir(), "empty"+gurps.SheetExt)
	c.NoError(gurps.NewEntity().Save(path))
	t.Setenv(benchFileEnvVar, path)
	var opened, skipped bool
	testing.Benchmark(func(b *testing.B) {
		sb := newSheetBench(b)
		opened = true
		b.Run("ranged", func(b *testing.B) {
			defer func() { skipped = b.Skipped() }()
			sb.scrollTo(b, sb.sheet.RangedWeapons)
		})
	})
	c.True(opened, "the sheet must open")
	c.True(skipped, "a list that is not on the page must not be measured")
}

// TestSheetSettingsUpdatedSyncsTheSheetOnlyOnce verifies that a change to a sheet's settings, which used to report
// itself more than once, syncs the sheet a single time while still bumping its modification timestamp, and that a
// change to another entity's settings leaves the sheet alone.
func TestSheetSettingsUpdatedSyncsTheSheetOnlyOnce(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	for _, fullRebuild := range []bool{false, true} {
		entity.ModifiedOn = jio.Time{}
		counter := installSyncCounter(sheet)
		sheet.SheetSettingsUpdated(entity, gurps.SheetSettingsAllParts, fullRebuild)
		c.Equal(1, counter.count, "a sheet settings change must sync the sheet exactly once (fullRebuild=%v)",
			fullRebuild)
		c.NotEqual(jio.Time{}, entity.ModifiedOn,
			"a sheet settings change must bump the modification timestamp (fullRebuild=%v)", fullRebuild)
		sheet.AsPanel().RemoveChild(counter)
	}

	// Settings for some other entity are none of this sheet's business.
	counter := installSyncCounter(sheet)
	sheet.SheetSettingsUpdated(gurps.NewEntity(), gurps.SheetSettingsAllParts, true)
	c.Equal(0, counter.count, "another entity's settings change must not touch the sheet")
}

// syncCountingProvider counts the syncs of the page list it is installed in: PageList.Sync is the only caller of
// SyncHeader.
type syncCountingProvider[T gurps.Node[T]] struct {
	TableProvider[T]
	syncs int
}

func (p *syncCountingProvider[T]) SyncHeader(headers []unison.TableColumnHeader[*Node[T]]) {
	p.syncs++
	p.TableProvider.SyncHeader(headers)
}

// A full rebuild syncs every list it keeps as it lays the page out, and the deep sync that follows used to sync each
// of them a second time, which on a sheet with large lists was half the cost of the rebuild. A list is synced once
// per update, whichever kind of update it is.
func TestFullRebuildSyncsEachListOnce(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Kept"
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)
	counter := &syncCountingProvider[*gurps.Trait]{TableProvider: sheet.Traits.provider}
	sheet.Traits.provider = counter
	list := sheet.Traits

	sheet.Rebuild(true)
	c.Equal(list, sheet.Traits, "the list must have been kept, or the count means nothing")
	c.Equal(1, counter.syncs, "a full rebuild syncs a list it keeps once")
	counter.syncs = 0
	sheet.Rebuild(false)
	c.Equal(1, counter.syncs, "a partial rebuild syncs the list once")
	counter.syncs = 0
	sheet.MarkModified(sheet)
	c.Equal(1, counter.syncs, "marking the sheet modified syncs the list once")

	data := gurps.NewTemplate()
	data.Traits = []*gurps.Trait{gurps.NewTrait(nil, nil, false)}
	template := newTestTemplateDockable("Syncs", data)
	templateCounter := &syncCountingProvider[*gurps.Trait]{TableProvider: template.Traits.provider}
	template.Traits.provider = templateCounter
	templateList := template.Traits
	template.Rebuild(true)
	c.Equal(templateList, template.Traits, "the template's list must have been kept")
	c.Equal(1, templateCounter.syncs, "a full rebuild of a template syncs a list it keeps once")
}

// TestSheetPageMatchesTheLayout checks that the sheet's page holds one panel per band of the layout that has anything
// to show, and that the blocks that aren't lists are the panels the sheet keeps rather than new ones.
func TestSheetPageMatchesTheLayout(t *testing.T) {
	c := check.New(t)
	sheet := newContinuousTestSheet(t)
	layout := sheet.Entity().SheetSettings.Layout

	c.Equal(nonEmptyBandCount(sheet, layout), len(sheet.page.Children()),
		"every band that has something to show, and only those, must be on the page")
	c.True(len(sheet.page.Children()) < len(layout.Root.Children),
		"a character with no content has bands with nothing to show")
	c.Equal(12, len(sheet.blocks), "every block that isn't a list is built and kept")
	for key, block := range sheet.blocks {
		c.NotNil(block.AsPanel().Parent(), "the %s block must be on the page", key)
	}
	c.Nil(sheet.Reactions.AsPanel().Parent(), "an empty reactions list must not be placed")
	c.NotNil(sheet.MeleeWeapons.AsPanel().Parent(), "a list with rows must be placed")
}

// nonEmptyBandCount returns how many of the layout's bands have something to show for the given sheet.
func nonEmptyBandCount(sheet *Sheet, layout *gurps.SheetLayout) int {
	count := 0
	for _, band := range layout.Root.Children {
		if bandHasContent(sheet, band) {
			count++
		}
	}
	return count
}

func bandHasContent(sheet *Sheet, node *gurps.SheetLayoutNode) bool {
	if node.Type == layoutnode.Block {
		return !isEmptyDerivedList(sheet, node.Key)
	}
	for _, child := range node.Children {
		if bandHasContent(sheet, child) {
			return true
		}
	}
	return false
}

// TestSheetHidingAndShowingABlock checks that hiding a block takes it off the page without taking it away from the
// sheet, that nothing reaching for it afterwards is handed something nobody can see, and that showing it brings it
// back.
func TestSheetHidingAndShowingABlock(t *testing.T) {
	c := check.New(t)
	sheet := newContinuousTestSheet(t)
	layout := sheet.Entity().SheetSettings.Layout
	notes := sheet.Notes
	c.NotNil(notes.AsPanel().Parent(), "the notes list starts out on the page")
	c.NotNil(sheet.keyToPanel(noteDragKey), "a note can be dropped onto the sheet")

	c.True(layout.Hide(gurps.BlockNotesKey))
	sheet.Rebuild(true)
	c.NotNil(sheet.Notes, "the sheet keeps its notes list even when the layout doesn't show it")
	c.Nil(sheet.Notes.AsPanel().Parent(), "a list that isn't on the page has no parent")
	c.Nil(sheet.keyToPanel(noteDragKey), "a list that isn't on the page is not a drop target")
	c.NotNil(sheet.keyToPanel(traitDragKey), "the lists that are still on the page remain drop targets")
	for _, child := range sheet.page.Children() {
		c.NotEqual(sheet.Notes.AsPanel(), child, "the notes list must not be on the page")
	}

	c.True(layout.Show(gurps.BlockNotesKey))
	sheet.Rebuild(true)
	c.NotNil(sheet.Notes.AsPanel().Parent(), "showing the block puts the list back on the page")
	c.NotNil(sheet.keyToPanel(noteDragKey), "and makes it a drop target again")
}

// TestSheetHidingABlockThatIsNotAList checks that a block the sheet keeps a panel for is taken off the page when it is
// hidden, and that the panel itself survives to be put back.
func TestSheetHidingABlockThatIsNotAList(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	layout := sheet.Entity().SheetSettings.Layout
	portrait := sheet.blocks[gurps.BlockPortraitKey]
	c.NotNil(portrait)

	c.True(layout.Hide(gurps.BlockPortraitKey))
	sheet.Rebuild(true)
	c.Equal(portrait, sheet.blocks[gurps.BlockPortraitKey], "the panel is kept")
	c.Nil(portrait.AsPanel().Parent(), "but it isn't on the page")

	c.True(layout.Show(gurps.BlockPortraitKey))
	sheet.Rebuild(true)
	c.NotNil(portrait.AsPanel().Parent(), "showing it puts the very same panel back")
	c.Equal(portrait, sheet.blocks[gurps.BlockPortraitKey])
}

// TestSwapDefaultsFollowsTheLayout checks that Swap Defaults is disabled while the layout leaves the skill or spell
// list off the page, where the list keeps its selection but has no undo manager, so the command would swap the hidden
// items' defaults where the user can't see and can't undo.
func TestSwapDefaultsFollowsTheLayout(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	newSkill := func(name, defaultsTo string, points fxp.Int) *gurps.Skill {
		sk := gurps.NewSkill(entity, nil, false)
		sk.Name = name
		sk.Points = points
		sk.Defaults = []*gurps.SkillDefault{{
			DefaultType: gurps.SkillID,
			Name:        criteria.Text{TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: defaultsTo}},
			Modifier:    -fxp.Four,
		}}
		return sk
	}
	broadsword := newSkill("Broadsword", "Rapier", fxp.Four)
	rapier := newSkill("Rapier", "Broadsword", fxp.Eight)
	entity.Skills = []*gurps.Skill{broadsword, rapier}
	entity.Recalculate()
	sheet.Rebuild(true)
	c.NotNil(broadsword.DefaultedFrom, "precondition: Broadsword defaults to Rapier")
	sheet.Skills.Table.SetSelectionMap(map[tid.TID]bool{broadsword.TID: true})
	c.True(sheet.CanPerformCmd(sheet, SwapDefaultsItemID), "the defaults can be swapped while the list is on the page")

	layout := entity.SheetSettings.Layout
	c.True(layout.Hide(gurps.BlockSkillsKey))
	sheet.Rebuild(true)
	c.True(sheet.Skills.Table.HasSelection(), "precondition: the hidden list keeps its selection")
	c.False(sheet.CanPerformCmd(sheet, SwapDefaultsItemID), "the command is disabled while the list is off the page")
	c.True(layout.Show(gurps.BlockSkillsKey))
	sheet.Rebuild(true)
	c.True(sheet.CanPerformCmd(sheet, SwapDefaultsItemID), "the command is enabled again once the list is back")

	// The spell list is acted on in the same way when the skill list has nothing selected.
	lore := gurps.NewSpell(entity, nil, false)
	lore.Name = "Lore"
	lore.Points = fxp.Four
	lore.Defaults = []*gurps.SkillDefault{
		{DefaultType: gurps.IntelligenceID, Modifier: -fxp.Two},
		{DefaultType: gurps.DexterityID, Modifier: -fxp.Two},
	}
	entity.Spells = []*gurps.Spell{lore}
	entity.Recalculate()
	sheet.Rebuild(true)
	sheet.Skills.Table.ClearSelection()
	sheet.Spells.Table.SetSelectionMap(map[tid.TID]bool{lore.TID: true})
	c.True(sheet.CanPerformCmd(sheet, SwapDefaultsItemID), "the defaults of a selected spell can be swapped")
	c.True(layout.Hide(gurps.BlockSpellsKey))
	sheet.Rebuild(true)
	c.True(sheet.Spells.Table.HasSelection(), "precondition: the hidden spell list keeps its selection")
	c.False(sheet.CanPerformCmd(sheet, SwapDefaultsItemID), "but not while the spell list is off the page")
}

// TestSwapDefaultsActsOnTheFocusedList checks that Swap Defaults acts on the skill or spell list holding the focus,
// rather than on the first with a selection, and so does nothing while that list has no selection.
func TestSwapDefaultsActsOnTheFocusedList(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	// Each has IQ-2 and DX-2 to move between, both at 8, and rests on the first.
	defaults := func() []*gurps.SkillDefault {
		return []*gurps.SkillDefault{
			{DefaultType: gurps.IntelligenceID, Modifier: -fxp.Two},
			{DefaultType: gurps.DexterityID, Modifier: -fxp.Two},
		}
	}
	var skill *gurps.Skill
	var spell *gurps.Spell
	screen.Do(func() {
		entity := sheet.Entity()
		skill = gurps.NewSkill(entity, nil, false)
		skill.Name = "Thaumatology"
		skill.Points = fxp.Four
		skill.Defaults = defaults()
		spell = gurps.NewSpell(entity, nil, false)
		spell.Name = "Lore"
		spell.Points = fxp.Four
		spell.Defaults = defaults()
		entity.Skills = []*gurps.Skill{skill}
		entity.Spells = []*gurps.Spell{spell}
		sheet.Rebuild(true)
		sheet.Skills.Table.SetSelectionMap(map[tid.TID]bool{skill.TID: true})
		sheet.Spells.Table.SetSelectionMap(map[tid.TID]bool{spell.TID: true})
		sheet.Spells.Table.RequestFocus()
	})
	c.Equal(gurps.IntelligenceID, skill.DefaultedFrom.Type(), "precondition: the skill rests on IQ")
	c.Equal(gurps.IntelligenceID, spell.DefaultedFrom.Type(), "precondition: the spell rests on IQ")
	screen.Do(func() {
		c.True(sheet.CanPerformCmd(sheet, SwapDefaultsItemID), "the spell's defaults can be swapped")
		sheet.PerformCmd(sheet, SwapDefaultsItemID)
	})
	c.Equal(gurps.DexterityID, spell.DefaultedFrom.Type(), "the spell list holding the focus is acted on")
	c.Equal(gurps.IntelligenceID, skill.DefaultedFrom.Type(), "rather than the skill list")
	screen.Do(func() {
		sheet.Spells.Table.ClearSelection()
		c.False(sheet.CanPerformCmd(sheet, SwapDefaultsItemID), "nor is the skill list acted on with no spell selected")
	})
}

// TestSwapDefaultsUndoRestoresBothLists checks that undoing a swap between a skill and a spell, which changes the
// defaults of both, puts both back.
func TestSwapDefaultsUndoRestoresBothLists(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	// Thaumatology and Lore default to each other, and the skill, updated first, rests on the spell.
	thaumatology := gurps.NewSkill(entity, nil, false)
	thaumatology.Name = "Thaumatology"
	thaumatology.Points = fxp.Four
	thaumatology.Defaults = []*gurps.SkillDefault{{
		DefaultType: gurps.SpellID,
		Name:        criteria.Text{Compare: criteria.IsText, Qualifier: "Lore"},
		Modifier:    -fxp.Two,
	}}
	lore := gurps.NewSpell(entity, nil, false)
	lore.Name = "Lore"
	lore.Points = fxp.Four
	lore.Defaults = []*gurps.SkillDefault{{
		DefaultType: gurps.SkillID,
		Name:        criteria.Text{Compare: criteria.IsText, Qualifier: "Thaumatology"},
		Modifier:    -fxp.Two,
	}}
	entity.Skills = []*gurps.Skill{thaumatology}
	entity.Spells = []*gurps.Spell{lore}
	sheet.Rebuild(true)
	restsOn := func() (skill, spell string) {
		if def := entity.Skills[0].DefaultedFrom; def != nil {
			skill = def.NameWithReplacements(nil)
		}
		if def := entity.Spells[0].DefaultedFrom; def != nil {
			spell = def.NameWithReplacements(nil)
		}
		return skill, spell
	}
	skill, spell := restsOn()
	c.True(skill == "Lore" && spell == "", "precondition: the skill rests on the spell")
	sheet.Skills.Table.SetSelectionMap(map[tid.TID]bool{thaumatology.TID: true})
	sheet.PerformCmd(sheet, SwapDefaultsItemID)
	skill, spell = restsOn()
	c.True(skill == "" && spell == "Thaumatology", "precondition: the swap has the spell rest on the skill")
	sheet.UndoManager().Undo()
	skill, spell = restsOn()
	c.Equal("Lore", skill, "undo has the skill rest on the spell again")
	c.Equal("", spell, "and the spell on nothing")
}

// TestHideAndShowBlock checks that taking a block off the sheet and putting it back are each one undoable edit that
// updates the sheet exactly once, and that both can still be undone once editing mode has been left.
func TestHideAndShowBlock(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetForLayoutEditing(t)
	layout := sheet.Entity().SheetSettings.Layout
	counter := installSyncCounter(sheet)
	mgr := sheet.UndoManager()
	before := layout.VisibleKeys()

	counter.count = 0
	sheet.hideLayoutBlock(gurps.BlockNotesKey)
	c.Equal(1, counter.count, "hiding a block must sync the sheet exactly once")
	c.False(layout.Contains(gurps.BlockNotesKey), "the hidden block must be out of the tree")
	c.Equal([]string{gurps.BlockNotesKey}, layout.HiddenKeys(), "the hidden block must be recorded as hidden")
	c.Nil(sheet.blockPanel(gurps.BlockNotesKey).AsPanel().Parent(), "the hidden block must be off the page")
	c.Nil(editor.ensureRegions().leafFor(gurps.BlockNotesKey), "the hidden block must have no region")

	counter.count = 0
	sheet.showLayoutBlock(gurps.BlockNotesKey)
	c.Equal(1, counter.count, "showing a block must sync the sheet exactly once")
	c.True(layout.Contains(gurps.BlockNotesKey), "the shown block must be back in the tree")
	c.Equal(0, len(layout.HiddenKeys()), "nothing must be left hidden")
	c.Equal(2, undoEditCount(mgr), "hiding and showing a block must be two undoable edits")

	// An undo can be asked for from the Edit menu long after editing mode was left, so it has to work from there.
	sheet.toggleLayoutEditing()
	c.False(sheet.layoutEditing(), "editing mode must have been left")

	mgr.Undo()
	c.False(sheet.Entity().SheetSettings.Layout.Contains(gurps.BlockNotesKey),
		"undoing the show must take the block back off the sheet")
	mgr.Undo()
	c.Equal(before, sheet.Entity().SheetSettings.Layout.VisibleKeys(),
		"undoing the hide must put the sheet back the way it was")
	mgr.Redo()
	c.False(sheet.Entity().SheetSettings.Layout.Contains(gurps.BlockNotesKey),
		"redoing the hide must take the block off the sheet again")
	mgr.Redo()
	c.Equal(before, sheet.Entity().SheetSettings.Layout.VisibleKeys(), "redoing the show must put the block back")
}

// TestHideBlockThatIsAlreadyHiddenDoesNothing checks that an edit changing nothing is neither recorded nor shown.
func TestHideBlockThatIsAlreadyHiddenDoesNothing(t *testing.T) {
	c := check.New(t)
	sheet, _ := newTestSheetForLayoutEditing(t)
	counter := installSyncCounter(sheet)
	sheet.hideLayoutBlock(gurps.BlockNotesKey)

	counter.count = 0
	sheet.hideLayoutBlock(gurps.BlockNotesKey)
	c.Equal(0, counter.count, "hiding a block that is already hidden must not sync the sheet")
	c.Equal(1, undoEditCount(sheet.UndoManager()), "hiding a block that is already hidden must not be recorded")
}

// TestResetLayoutRestoresTheDefault checks that resetting returns the sheet to the layout new sheets are given, as one
// undoable edit that updates the sheet exactly once.
func TestResetLayoutRestoresTheDefault(t *testing.T) {
	c := check.New(t)
	sheet, _ := newTestSheetForLayoutEditing(t)
	counter := installSyncCounter(sheet)
	mgr := sheet.UndoManager()
	original := gurps.Hash64(sheet.Entity().SheetSettings.Layout)
	sheet.hideLayoutBlock(gurps.BlockNotesKey)
	c.NotEqual(original, gurps.Hash64(sheet.Entity().SheetSettings.Layout), "the layout must have been altered")

	counter.count = 0
	sheet.resetLayout()
	c.Equal(1, counter.count, "resetting the layout must sync the sheet exactly once")
	c.Equal(gurps.Hash64(gurps.GlobalSettings().Sheet.Layout), gurps.Hash64(sheet.Entity().SheetSettings.Layout),
		"resetting must give the sheet the layout new sheets are given")

	mgr.Undo()
	c.False(sheet.Entity().SheetSettings.Layout.Contains(gurps.BlockNotesKey),
		"undoing the reset must take the block back off the sheet")
	mgr.Redo()
	c.Equal(original, gurps.Hash64(sheet.Entity().SheetSettings.Layout), "redoing the reset must restore the default")
	c.Equal(2, undoEditCount(mgr), "resetting the layout must be one more undoable edit")
}

// sheetLayoutGivenToNewEntity returns the block layout an entity with no embedded sheet settings is given when it is
// unmarshaled. It comes from the published snapshot of the global sheet settings rather than the live settings, so it
// is what shows whether a change to the default layout was published.
func sheetLayoutGivenToNewEntity(t *testing.T) *gurps.SheetLayout {
	t.Helper()
	var e gurps.Entity
	e.DiscardCaches()
	check.New(t).NoError(jio.Unmarshal(fmt.Appendf(nil, `{"version":%d}`, jio.CurrentDataVersion), &e))
	return e.SheetSettings.Layout
}

// restoreDefaultSheetLayout puts the given layout back as the default for new sheets when the test finishes,
// republishing it, since the default the test installed was published as well.
func restoreDefaultSheetLayout(t *testing.T, layout *gurps.SheetLayout) {
	t.Helper()
	t.Cleanup(func() {
		gurps.GlobalSettings().Sheet.Layout = layout
		gurps.SyncGlobalSheetSettings()
	})
}

// TestUseLayoutAsDefaultStoresACopy checks that making a sheet's layout the default for new sheets stores a copy of it
// rather than the sheet's own, publishes it for the sheets opened from then on, and tells everything that lays itself
// out from the defaults about the change.
func TestUseLayoutAsDefaultStoresACopy(t *testing.T) {
	c := check.New(t)
	global := gurps.GlobalSettings()
	restoreDefaultSheetLayout(t, global.Sheet.Layout)

	sheet, _ := newTestSheetForLayoutEditing(t)
	recorder := &sheetSettingsRecorder{}
	recorder.Self = recorder
	Workspace.DocumentDock.DockTo(recorder, nil, side.Left)
	sheet.hideLayoutBlock(gurps.BlockNotesKey)
	recorder.entities = nil
	recorder.updates = nil
	c.True(sheetLayoutGivenToNewEntity(t).Contains(gurps.BlockNotesKey),
		"the published default must start out showing the block the sheet hides, or the test proves nothing")

	sheet.useLayoutAsDefault()

	layout := sheet.Entity().SheetSettings.Layout
	c.Equal(gurps.Hash64(layout), gurps.Hash64(global.Sheet.Layout), "the default must match the sheet's layout")
	c.True(layout != global.Sheet.Layout, "the default must be a copy, not the sheet's own layout")
	c.Equal(gurps.Hash64(layout), gurps.Hash64(sheetLayoutGivenToNewEntity(t)),
		"a sheet without embedded settings opened from now on must be given the new default")
	c.Equal(1, len(recorder.updates), "everything that lays itself out from the defaults must be told once")
	c.True(recorder.updates[0], "the notification must ask for a full rebuild")
	c.Nil(recorder.entities[0], "a nil entity is what says the change was to the defaults")
	c.Equal(gurps.SheetSettingsLayoutPart, recorder.parts[0], "the notification must say only the layout changed")

	// Neither of the two can be changed by way of the other.
	global.Sheet.Layout.Hide(gurps.BlockSpellsKey)
	c.True(layout.Contains(gurps.BlockSpellsKey), "changing the default must not change the sheet")
}

// TestResetDefaultLayoutRestoresTheFactoryLayout checks that resetting the default layout puts the factory layout in
// place for new sheets and publishes it for the sheets opened from then on, leaves the sheet's own layout alone, and
// tells everything that lays itself out from the defaults about the change.
func TestResetDefaultLayoutRestoresTheFactoryLayout(t *testing.T) {
	c := check.New(t)
	global := gurps.GlobalSettings()
	restoreDefaultSheetLayout(t, global.Sheet.Layout)

	sheet, _ := newTestSheetForLayoutEditing(t)
	recorder := &sheetSettingsRecorder{}
	recorder.Self = recorder
	Workspace.DocumentDock.DockTo(recorder, nil, side.Left)
	sheet.hideLayoutBlock(gurps.BlockNotesKey)
	sheet.useLayoutAsDefault()
	c.False(global.Sheet.Layout.Contains(gurps.BlockNotesKey), "the default must start out altered")
	c.False(sheetLayoutGivenToNewEntity(t).Contains(gurps.BlockNotesKey),
		"the published default must start out altered too, or the test proves nothing")
	recorder.entities = nil
	recorder.updates = nil

	sheet.resetDefaultLayout()

	c.Equal(gurps.Hash64(gurps.FactorySheetLayout()), gurps.Hash64(global.Sheet.Layout),
		"the default must be the factory layout again")
	c.Equal(gurps.Hash64(gurps.FactorySheetLayout()), gurps.Hash64(sheetLayoutGivenToNewEntity(t)),
		"a sheet without embedded settings opened from now on must be given the factory layout again")
	c.False(sheet.Entity().SheetSettings.Layout.Contains(gurps.BlockNotesKey),
		"the sheet's own layout must be left alone")
	c.Equal(1, len(recorder.updates), "everything that lays itself out from the defaults must be told once")
	c.True(recorder.updates[0], "the notification must ask for a full rebuild")
	c.Nil(recorder.entities[0], "a nil entity is what says the change was to the defaults")
	c.Equal(gurps.SheetSettingsLayoutPart, recorder.parts[0], "the notification must say only the layout changed")
}

// TestHiddenListsOfferNoCommands checks that the commands that add to a list, or move equipment into it, are disabled
// while the layout leaves the list off the page, since they would otherwise change a list the user can't see without
// recording an undo edit, and are enabled again once the list is shown.
func TestHiddenListsOfferNoCommands(t *testing.T) {
	c := check.New(t)
	sheet, _ := newTestSheetForLayoutEditing(t)
	entity := sheet.Entity()
	carried := gurps.NewEquipment(entity, nil, false)
	carried.Name = "Rope"
	entity.CarriedEquipment = append(entity.CarriedEquipment, carried)
	other := gurps.NewEquipment(entity, nil, false)
	other.Name = "Tent"
	entity.OtherEquipment = append(entity.OtherEquipment, other)
	entity.Recalculate()
	sheet.Rebuild(true)
	sheet.CarriedEquipment.Table.SelectByIndex(0)
	sheet.OtherEquipment.Table.SelectByIndex(0)
	type gates struct {
		newOther, newOtherContainer, newOtherGroup, moveToOther         bool
		newCarried, newCarriedContainer, newCarriedGroup, moveToCarried bool
		newTrait                                                        bool
	}
	readGates := func() gates {
		return gates{
			newOther:            sheet.CanPerformCmd(nil, NewOtherEquipmentItemID),
			newOtherContainer:   sheet.CanPerformCmd(nil, NewOtherEquipmentContainerItemID),
			newOtherGroup:       sheet.CanPerformCmd(nil, NewOtherEquipmentGroupItemID),
			moveToOther:         sheet.CarriedEquipment.CanPerformCmd(nil, MoveToOtherEquipmentItemID),
			newCarried:          sheet.CanPerformCmd(nil, NewCarriedEquipmentItemID),
			newCarriedContainer: sheet.CanPerformCmd(nil, NewCarriedEquipmentContainerItemID),
			newCarriedGroup:     sheet.CanPerformCmd(nil, NewCarriedEquipmentGroupItemID),
			moveToCarried:       sheet.OtherEquipment.CanPerformCmd(nil, MoveToCarriedEquipmentItemID),
			newTrait:            sheet.CanPerformCmd(nil, NewTraitItemID),
		}
	}
	all := gates{true, true, true, true, true, true, true, true, true}
	c.Equal(all, readGates(), "every command is enabled while every list is shown")

	sheet.hideLayoutBlock(gurps.BlockOtherEquipmentKey)
	c.Equal(gates{false, false, false, false, true, true, true, true, true}, readGates(),
		"hiding the other equipment disables adding to it and moving equipment into it")
	sheet.showLayoutBlock(gurps.BlockOtherEquipmentKey)
	c.Equal(all, readGates(), "showing it enables them again")

	sheet.hideLayoutBlock(gurps.BlockEquipmentKey)
	c.Equal(gates{true, true, true, true, false, false, false, false, true}, readGates(),
		"hiding the carried equipment disables adding to it and moving equipment into it")
	sheet.showLayoutBlock(gurps.BlockEquipmentKey)
	c.Equal(all, readGates())

	sheet.hideLayoutBlock(gurps.BlockTraitsKey)
	c.False(readGates().newTrait, "hiding the traits disables adding to them")
}

// TestSetDefaultSheetLayoutRebuildsOnlyWhatShowsIt checks that setting the default block layout rebuilds the open
// templates, which lay themselves out from it, and nothing else: a library list and a loot sheet have no layout, and
// used to be fully rebuilt all the same.
func TestSetDefaultSheetLayoutRebuildsOnlyWhatShowsIt(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	global := gurps.GlobalSettings()
	restoreDefaultSheetLayout(t, global.Sheet.Layout)

	var sheet *Sheet
	var template *Template
	var loot *LootSheet
	var list *TableDockable[*gurps.Trait]
	screen.Do(func() {
		sheet = NewSheet("test"+gurps.SheetExt, gurps.NewEntity())
		DisplayNewDockable(sheet)
		sheet.markUnmodified() // So that closing it asks nothing.
		template = newTestTemplateDockable("Template", gurps.NewTemplate())
		DisplayNewDockable(template)
		template.markUnmodified()
		loot = NewLootSheet("test"+gurps.LootExt, gurps.NewLoot())
		DisplayNewDockable(loot)
		loot.markUnmodified()
		list = NewTraitTableDockable(filepath.Join(t.TempDir(), "test"+gurps.TraitsExt),
			[]*gurps.Trait{newFilterTestTrait("Acute Vision", "Physical")})
		DisplayNewDockable(list)
	})
	row := list.table.RowFromIndex(0)
	c.NotNil(row, "the library list must show its trait")
	lootStamp := loot.Equipment.syncedAtStamp()
	before := currentDisplayStamp()

	screen.Do(func() {
		sheet.Entity().SheetSettings.Layout.Hide(gurps.BlockNotesKey)
		sheet.useLayoutAsDefault()
	})

	c.False(global.Sheet.Layout.Contains(gurps.BlockNotesKey), "the default must be the sheet's layout")
	c.True(template.Traits.syncedAtStamp() > before,
		"a template lays itself out from the default layout, so it must be rebuilt")
	c.Equal(lootStamp, loot.Equipment.syncedAtStamp(), "a loot sheet has no layout, so it must be left alone")
	c.True(list.table.RowFromIndex(0) == row, "a library list has no layout, so it must be left alone")
}

// closeBuyDockable closes the sheet's buy dockable, if it has one, abandoning the purchase.
func closeBuyDockable(sheet *Sheet) {
	if d := buyDockableFor(sheet); d != nil {
		d.discardAndClose()
	}
}

// TestBuyingLeavesTheSheetEditable verifies that a sheet stays editable while its buy dockable is open: the commands that
// add to the carried equipment list stay enabled, the list stays the carried equipment block's panel for the layout
// editor, equipment dropped on the sheet goes to the dockable, and hiding the block neither ends buying nor takes the
// dockable out of the sheet's drop rerouting.
func TestBuyingLeavesTheSheetEditable(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	newCarriedIDs := []int{
		NewCarriedEquipmentItemID,
		NewCarriedEquipmentContainerItemID,
		NewCarriedEquipmentGroupItemID,
	}
	type gates struct {
		moveToCarried bool
		newCarried    []bool
	}
	readGates := func() gates {
		var g gates
		g.moveToCarried = sheet.OtherEquipment.CanPerformCmd(nil, MoveToCarriedEquipmentItemID)
		for _, id := range newCarriedIDs {
			g.newCarried = append(g.newCarried, sheet.CanPerformCmd(nil, id))
		}
		return g
	}
	var before, during, after gates
	var buyEnabled, blockPanelIsList, dropTargetIsDockable, editorFindsList, buyingAfterHide, dropTargetAfterHide bool
	screen.Do(func() {
		entity := sheet.Entity()
		eqp := gurps.NewEquipment(entity, nil, false)
		eqp.Name = "Rope"
		entity.OtherEquipment = append(entity.OtherEquipment, eqp)
		entity.Recalculate()
		sheet.Rebuild(true)
		sheet.OtherEquipment.Table.SelectByIndex(0)
		before = readGates()
		buyEnabled = buyEquipmentAction.EnabledCallback(buyEquipmentAction, nil)

		d := showBuyDockable(sheet)
		during = readGates()
		blockPanelIsList = sheet.blockPanel(gurps.BlockEquipmentKey) == unison.Paneler(sheet.CarriedEquipment)
		dropTargetIsDockable = sheet.keyToPanel(equipmentDragKey) == d.AsPanel()
		sheet.toggleLayoutEditing()
		if sheet.layoutEditing() {
			editorFindsList = sheet.layoutEditor.ensureRegions().leafFor(gurps.BlockEquipmentKey) != nil
			sheet.toggleLayoutEditing()
		}
		sheet.hideLayoutBlock(gurps.BlockEquipmentKey)
		buyingAfterHide = buyDockableFor(sheet) == d
		dropTargetAfterHide = sheet.keyToPanel(equipmentDragKey) == d.AsPanel()
		sheet.UndoManager().Undo()
		closeBuyDockable(sheet)
		after = readGates()
	})
	c.True(buyEnabled, "Buy Equipment must be enabled for a new sheet")
	c.True(before.moveToCarried, "moving the selected other equipment to the carried list must be enabled")
	c.Equal([]bool{true, true, true}, before.newCarried, "the New Carried Equipment commands must be enabled")
	c.Equal(before, during, "buying must leave the commands that add to the carried equipment list enabled")
	c.True(blockPanelIsList, "the carried equipment list must stay the block's panel while buying")
	c.True(dropTargetIsDockable, "equipment dropped on the sheet must go to the buy dockable")
	c.True(editorFindsList, "the layout editor must find the carried equipment block while buying")
	c.True(buyingAfterHide, "hiding the carried equipment block must not end buying")
	c.True(dropTargetAfterHide, "equipment dropped on the sheet must still go to the buy dockable with the list hidden")
	c.Equal(before, after, "closing the dockable must leave everything as it was")
}

// TestDropOnAPaginatedSheetGoesToTheEndOfTheList verifies that an item dropped anywhere on a sheet whose list goes on
// over several pages is rerouted to the view of the list showing its last row, so that it lands at the end of the
// list. Rerouting it to the list's first view, which ends on an earlier page, would insert it there instead.
func TestDropOnAPaginatedSheetGoesToTheEndOfTheList(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet := openPaginatedSheet(t, screen, func(entity *gurps.Entity) []*gurps.Trait { return plainTraits(entity, 150) })
	table := sheet.Traits.Table
	var target, last *unison.Panel
	screen.Do(func() {
		target = sheet.keyToPanel(traitDragKey)
		last = table.ViewForRow(table.LastRowIndex()).AsPanel()
	})
	c.True(last != table.AsPanel(), "the list's last row must be on a later page than its first")
	c.True(target == last, "the drop must go to the view showing the list's last row")
}

// TestTogglingPagesKeepsTheFocus verifies that switching a sheet between pages and one continuous page leaves the
// keyboard focus where it was, whether on a field or on a list going on over several pages, where it stays on the view
// showing the selected row, and that the continuous page is the only one, with no list going on past it. Rebuilding
// without putting the focus back, putting it on another view of the list, or leaving pages or continuations behind,
// fails it.
func TestTogglingPagesKeepsTheFocus(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet := openPaginatedSheet(t, screen, func(entity *gurps.Entity) []*gurps.Trait { return plainTraits(entity, 150) })
	field := func() *unison.Panel { return sheet.AsPanel().FindRefKey(identityPanelNameFieldRefKey) }
	for _, one := range []struct {
		name   string
		focus  func() *unison.Panel
		holder func() *unison.Panel
	}{
		{name: "field", focus: field, holder: field},
		{
			name: "list",
			focus: func() *unison.Panel {
				table := sheet.Traits.Table
				views := table.Views()
				view := views[len(views)-1]
				start, _ := view.CurrentDrawRowRange()
				table.SelectByIndex(start)
				return view.AsPanel()
			},
			holder: func() *unison.Panel {
				table := sheet.Traits.Table
				return table.ViewForRow(table.FirstSelectedRowIndex()).AsPanel()
			},
		},
	} {
		var ref *FocusRef
		screen.Do(func() {
			if !sheet.paginated {
				sheet.togglePaginated()
			}
			if target := one.focus(); target != nil {
				target.RequestFocus()
			}
			ref = sheet.targetMgr.CurrentFocusRef()
		})
		if ref == nil || ref.Key == "" {
			t.Fatalf("%s: the focus must be on the sheet, on a panel with a reference key", one.name)
		}
		for _, paginated := range []bool{false, true} {
			var after *FocusRef
			var pages, views int
			var onHolder bool
			screen.Do(func() {
				sheet.togglePaginated()
				after = sheet.targetMgr.CurrentFocusRef()
				pages = len(sheet.grid.pages)
				views = len(sheet.Traits.Table.Views())
				onHolder = wnd.Focus() == one.holder()
			})
			c.Equal(paginated, sheet.paginated, "%s: the mode must have switched", one.name)
			c.NotNil(after, "%s: the focus must stay on the sheet (paginated %v)", one.name, paginated)
			if after != nil {
				c.Equal(ref.Key, after.Key, "%s: the focus must stay where it was (paginated %v)", one.name, paginated)
			}
			c.True(onHolder, "%s: the focus must be on the panel showing what it was on (paginated %v)", one.name,
				paginated)
			if paginated {
				c.True(pages > 1, "%s: the pages must come back", one.name)
			} else {
				c.Equal(1, pages, "%s: the sheet must show one continuous page", one.name)
				c.Equal(1, views, "%s: no list may go on past the continuous page", one.name)
			}
		}
	}
}

// TestLeavingTheLayoutEditorKeepsThePlaceOnTheSheet verifies that a sheet shown as pages and scrolled well down is left
// about as far down once its layout has been edited, rather than scrolled to wherever the editor hands the focus back.
// Bringing the focus into view as the pages come back fails it.
func TestLeavingTheLayoutEditorKeepsThePlaceOnTheSheet(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet := openPaginatedSheet(t, screen, func(entity *gurps.Entity) []*gurps.Trait { return plainTraits(entity, 300) })
	var fraction float32
	screen.Do(func() {
		sheet.scroll.SetPosition(0, sheet.scroll.Bar(false).MaxValue())
		sheet.toggleLayoutEditing()
		sheet.toggleLayoutEditing()
		fraction = sheet.scrollFraction()
	})
	c.True(fraction > 0.5, "the sheet must still be scrolled well down, not %v of the way", fraction)
}
