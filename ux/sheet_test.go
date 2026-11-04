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
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/gcs/v5/model/gurps/gurpstest"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	uncheck "github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/role"
)

// TestMarkModifiedRecalculates checks that telling the sheet something changed brings the entity's derived state up to
// date, since everything MarkModified goes on to do displays that state, and hashing the entity for the tab's unsaved
// changes check no longer recalculates it.
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

// TestSheetBenchSkipsAListNotOnThePage verifies that a benchmark of a list that a loaded sheet leaves off the page, such
// as the ranged weapons of a character without any, is skipped rather than measuring whatever part of the page is in
// view.
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
