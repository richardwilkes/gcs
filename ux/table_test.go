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
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

// editingTableCmdIDs are the standard table commands that only an editable table offers.
var editingTableCmdIDs = []int{unison.DeleteItemID, DuplicateItemID, SyncWithSourceItemID, ClearSourceItemID}

// TestStandardTableCmdsReachableFromDockableFilterField verifies that the standard table commands of a list dockable
// are installed on the dockable rather than on its table, so that they still act on the selection while the filter
// field has the focus. Sync With Source and Clear Source used to be the odd ones out, installed on the table alone.
// Delete is left out because the field claims it for its own text.
func TestStandardTableCmdsReachableFromDockableFilterField(t *testing.T) {
	c := check.New(t)
	dockable := newTestTraitTableDockable()
	c.True(len(dockable.table.RootRows()) > 0, "the test list must have rows")
	ids := []int{OpenEditorItemID, DuplicateItemID, SyncWithSourceItemID, ClearSourceItemID}
	dockable.table.SelectByIndex(0)
	for _, id := range ids {
		c.True(dockable.filterField.CanPerformCmd(nil, id), "command %d must be reachable from the filter field", id)
	}
	dockable.table.ClearSelection()
	for _, id := range ids {
		c.False(dockable.filterField.CanPerformCmd(nil, id), "command %d must need a selection", id)
	}
}

// TestPageListEditingCmdsFollowOwner verifies that a page list with an owner offers the editing commands and that a
// read-only page list -- one derived from the rest of the sheet, which has no owner -- offers only the opening ones.
func TestPageListEditingCmdsFollowOwner(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Claws"
	trait.Weapons = []*gurps.Weapon{gurps.NewWeapon(trait, true)}
	entity.Traits = []*gurps.Trait{trait}

	traits := NewTraitsPageList(sheet, entity)
	c.True(len(traits.Table.RootRows()) > 0, "the traits list must show the trait")
	traits.Table.SelectByIndex(0)
	c.True(traits.Table.CanPerformCmd(nil, OpenEditorItemID), "an owned list must offer Open Editor")
	for _, id := range editingTableCmdIDs {
		c.True(traits.Table.CanPerformCmd(nil, id), "an owned list must offer command %d", id)
	}

	weapons := NewMeleeWeaponsPageList(entity)
	c.True(len(weapons.Table.RootRows()) > 0, "the weapons list must show the trait's weapon")
	weapons.Table.SelectByIndex(0)
	c.True(weapons.Table.CanPerformCmd(nil, OpenEditorItemID), "a read-only list must still offer Open Editor")
	for _, id := range editingTableCmdIDs {
		c.False(weapons.Table.CanPerformCmd(nil, id), "a read-only list must not offer command %d", id)
	}
}

// TestPageTableSizesItsColumnsForANewWidthOrNewRows verifies that a frame change of any view of a page table sizes the
// table's columns again only when the width of the table's primary view or the heights of its rows are not what they
// were last sized for: a frame change that alters only a height, or that of a continuation of the list whatever its
// width, leaves them be. Sizing them on every frame change, letting a continuation size them for its own width, sizing
// them for the width alone, which leaves disclosed rows in columns too narrow for them, or for the width and the number
// of rows alone, which leaves an editor's columns as they were when what a row shows changed, fails it.
func TestPageTableSizesItsColumnsForANewWidthOrNewRows(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	container := gurps.NewTrait(entity, nil, true)
	container.Name = "Container"
	container.SetOpen(false)
	child := gurps.NewTrait(entity, container, false)
	child.Name = "Child"
	container.Children = []*gurps.Trait{child}
	entity.Traits = []*gurps.Trait{container}
	setFrame := func(table *unison.Table[*Node[*gurps.Trait]], width, height float32) {
		table.SetFrameRect(geom.NewRect(0, 0, width, height))
	}
	for _, one := range []struct {
		name   string
		change func(list *PageList[*gurps.Trait])
		sized  bool
	}{
		{
			name:   "a new height",
			change: func(list *PageList[*gurps.Trait]) { setFrame(list.Table, 800, 150) },
		},
		{
			name:   "a new width",
			change: func(list *PageList[*gurps.Trait]) { setFrame(list.Table, 700, 100) },
			sized:  true,
		},
		{
			name:   "a continuation's frame",
			change: func(list *PageList[*gurps.Trait]) { setFrame(list.continuation(0).Table, 500, 60) },
		},
		{
			name: "disclosed rows",
			change: func(list *PageList[*gurps.Trait]) {
				// unison fits the frame to the table as it syncs, which can alter its width, so the frame change is
				// held back until the frame is back at the width the columns were sized for.
				frameChanged := list.Table.FrameChangeCallback
				list.Table.FrameChangeCallback = nil
				list.SetDisclosureState(true)
				list.Table.FrameChangeCallback = frameChanged
				setFrame(list.Table, 800, 120)
			},
			sized: true,
		},
		{
			name: "a row grown taller",
			change: func(list *PageList[*gurps.Trait]) {
				// An editor's table is synced to its model without its columns being sized, unlike a page list (see
				// PageList.Sync), and the frame change is held back as for disclosed rows.
				frameChanged := list.Table.FrameChangeCallback
				list.Table.FrameChangeCallback = nil
				container.Name = strings.Repeat("A much longer name ", 40)
				list.Table.SyncToModel()
				list.Table.FrameChangeCallback = frameChanged
				setFrame(list.Table, 800, 150)
			},
			sized: true,
		},
	} {
		container.SetOpen(false)
		container.Name = "Container"
		list := NewTraitsPageList(sheet, entity)
		setFrame(list.Table, 800, 100)
		col := 0
		if col == list.Table.ColumnIndexForID(list.provider.ExcessWidthColumnID()) {
			col = 1
		}
		list.Table.Columns[col].Current = 1
		one.change(list)
		c.Equal(one.sized, list.Table.Columns[col].Current != 1, "%s: whether the columns are sized again", one.name)
	}
}

// watchedListOwner is a data owner without a source matcher, as an equipment list file is, that calls onSourceMatcher
// whenever it is asked for one.
type watchedListOwner struct {
	onSourceMatcher func()
}

func (w *watchedListOwner) OwningEntity() *gurps.Entity { return nil }

func (w *watchedListOwner) SourceMatcher() *gurps.SrcMatcher {
	w.onSourceMatcher()
	return nil
}

func (w *watchedListOwner) WeightUnit() fxp.WeightUnit { return fxp.Pound }

// TestSyncWithSourceForSelectionLoadsEachFileOnce verifies that syncing the selected rows of a library list loads the
// library file they are sourced from just once for the whole selection, rather than checking it again for each row, by
// taking the file away as soon as the first row has been synced, and that the file is checked again for a row matched
// once the sync is done.
func TestSyncWithSourceForSelectionLoadsEachFileOnce(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	_, user := uxtest.UseTestLibraries(t, c)
	RegisterKnownFileTypes()
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Test" + gurps.NotesExt}
	p := filepath.Join(user.Path(false), libFile.Path)
	libNotes := []*gurps.Note{gurps.NewNote(nil, nil, false), gurps.NewNote(nil, nil, false)}
	libNotes[0].MarkDown = "First"
	libNotes[1].MarkDown = "Second"
	c.NoError(gurps.SaveNotes(libNotes, p))
	owner := &watchedListOwner{}
	notes := make([]*gurps.Note, len(libNotes))
	for i, one := range libNotes {
		notes[i] = one.Clone(libFile, owner, nil, gurps.Reference)
		notes[i].MarkDown += " (old)"
	}
	// Each row asks its data owner for a matcher as it is synced, which is when a check of its file would be made.
	owner.onSourceMatcher = func() {
		if notes[0].MarkDown == "First" {
			if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
				t.Error(err)
			}
		}
	}
	dockable := NewNoteTableDockable("test"+gurps.NotesExt, notes)
	dockable.table.SelectAll()

	SyncWithSourceForSelection(dockable.table)
	_, err := os.Stat(p)
	c.True(os.IsNotExist(err), "precondition: the file was taken away during the sync")
	c.Equal("First", notes[0].MarkDown)
	c.Equal("Second", notes[1].MarkDown, "a row after the first is synced with the file as loaded for the selection")

	notes[0].MarkDown = "First (old)"
	state, _ := gurps.MatchSource(notes[0])
	c.Equal(srcstate.Missing, state, "the file is checked again for a row matched once the sync is done")
}

// TestSheetListContextMenuHeadless verifies that a list on a character sheet offers its context menu every way a user
// can ask for it. A right-click on one of several selected rows keeps them all selected for the menu's commands to act
// on, while a right-click or a screen reader's request on a row that is not selected makes it the selection alone, and
// either puts the focus on the list, since the menu's commands act on what holds it.
func TestSheetListContextMenuHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	screen.EnableAccessibility()
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	names := []string{"Alpha", "Beta", "Gamma"}
	var table *unison.Table[*Node[*gurps.Trait]]
	screen.Do(func() {
		entity := sheet.Entity()
		for _, name := range names {
			trait := gurps.NewTrait(entity, nil, false)
			trait.Name = name
			entity.Traits = append(entity.Traits, trait)
		}
		entity.Recalculate()
		sheet.Rebuild(true)
		table = sheet.Traits.Table
	})
	if table == nil {
		t.Fatal("the sheet must have a traits list")
	}
	selectedNames := func() []string {
		var selected []string
		screen.Do(func() {
			for _, row := range table.SelectedRows(false) {
				selected = append(selected, row.Data().Name)
			}
		})
		slices.Sort(selected)
		return selected
	}
	tableFocused := func() bool {
		var focused bool
		screen.Do(func() { focused = wnd.Focus() == table.AsPanel() })
		return focused
	}
	// rowPoint brings the named trait's row into view and returns the middle of its description cell on the screen,
	// clear of the check box at the front of the row that a click would toggle. Rows are found by name, since a new
	// sheet comes with a trait of its own.
	rowPoint := func(name string) geom.Point {
		t.Helper()
		row := -1
		screen.Do(func() {
			for i := 0; i <= table.LastRowIndex(); i++ {
				if table.RowFromIndex(i).Data().Name == name {
					row = i
					break
				}
			}
		})
		if row < 0 {
			t.Fatalf("the traits list must have a row for %s", name)
		}
		return cellScreenPoint(t, screen, wnd, table, row, gurps.TraitDescriptionColumn)
	}
	checkMenu := func(how string) {
		t.Helper()
		titles := contextMenuTitles(t, screen, wnd)
		c.NotNil(titles, "%s must open the list's menu", how)
		c.True(slices.Contains(titles, openEditorAction.Title), "the menu %s opens must offer %q; it offers %v", how,
			openEditorAction.Title, titles)
		c.True(slices.Contains(titles, duplicateAction.Title), "the menu %s opens must offer %q; it offers %v", how,
			duplicateAction.Title, titles)
	}

	screen.Click(rowPoint("Alpha"))
	c.Equal([]string{"Alpha"}, selectedNames(), "clicking a row selects it")
	c.True(tableFocused(), "clicking a row focuses the list")

	screen.KeyPress(unison.KeyF10, mod.Shift)
	checkMenu("shift+F10")
	c.Equal([]string{"Alpha"}, selectedNames(), "opening the menu leaves the selection alone")
	closeContextMenu(c, screen, wnd)
	screen.KeyPress(unison.KeyMenu, mod.None)
	checkMenu("the Menu key")
	closeContextMenu(c, screen, wnd)

	screen.ClickWith(rowPoint("Beta"), unison.ButtonLeft, mod.Shift)
	c.Equal([]string{"Alpha", "Beta"}, selectedNames(), "shift-clicking the next row extends the selection to it")
	screen.ClickWith(rowPoint("Beta"), unison.ButtonRight, mod.None)
	checkMenu("a right-click on a selected row")
	c.Equal([]string{"Alpha", "Beta"}, selectedNames(),
		"a right-click on one of several selected rows must keep them all selected for the menu to act on")
	closeContextMenu(c, screen, wnd)

	// A screen reader asks a row that is not selected for the menu while the focus is elsewhere. Only rows that can be
	// seen, are selected or hold the focus are described, so the row is brought into view first.
	var navigatorFocused bool
	screen.Do(func() {
		Workspace.Navigator.InitialFocus()
		navigatorFocused = wnd.Focus() == Workspace.Navigator.table.AsPanel()
	})
	c.True(navigatorFocused, "the test needs the focus to start out in the navigator")
	rowPoint("Gamma")
	tree := screen.AccessibilityTree(wnd)
	if tree == nil {
		t.Fatal("the workspace window must be described")
	}
	tableNode := screen.AccessibilityNodeFor(table)
	if tableNode == nil {
		t.Fatal("the traits list must be described")
	}
	var gamma *accessibility.Node
	for _, id := range tableNode.Children {
		if node := tree.Node(id); node != nil && node.Role == role.Row && strings.Contains(node.Name, "Gamma") {
			gamma = node
		}
	}
	if gamma == nil {
		t.Fatal("the row for Gamma must be described")
	}
	c.False(gamma.Selected, "the Gamma row must not start out selected")
	c.True(gamma.Actions.Has(accessibility.ShowContextMenu), "a row of the list must offer its context menu")
	c.True(screen.PerformAccessibilityAction(accessibility.ActionRequest{
		Node:   gamma.ID,
		Action: accessibility.ShowContextMenu,
	}), "the row must carry out the request for its context menu")
	checkMenu("a screen reader's request")
	c.Equal([]string{"Gamma"}, selectedNames(), "asking a row that is not selected for the menu makes it the selection")
	c.True(tableFocused(), "asking a row for the menu puts the focus on the list its commands act on")
	var menuNodes int
	screen.AccessibilityTree(wnd).Walk(func(n *accessibility.Node) bool {
		if n.Role == role.Menu {
			menuNodes++
		}
		return true
	})
	c.Equal(1, menuNodes, "the open menu must be described to the screen reader")
	closeContextMenu(c, screen, wnd)

	// A right-click on a row outside a selection of several, made while the focus is elsewhere, makes that row the
	// whole selection and puts the focus on the list, although unison delivers the list no mouse event for it.
	screen.ClickWith(rowPoint("Beta"), unison.ButtonLeft, mod.Shift)
	c.Equal([]string{"Beta", "Gamma"}, selectedNames(), "shift-clicking the row before extends the selection to it")
	screen.Do(func() {
		Workspace.Navigator.InitialFocus()
		navigatorFocused = wnd.Focus() == Workspace.Navigator.table.AsPanel()
	})
	c.True(navigatorFocused, "the test needs the focus to be in the navigator again")
	screen.ClickWith(rowPoint("Alpha"), unison.ButtonRight, mod.None)
	checkMenu("a right-click on a row that is not selected")
	c.Equal([]string{"Alpha"}, selectedNames(), "a right-click on a row that is not selected makes it the selection")
	c.True(tableFocused(), "a right-click on a row puts the focus on the list its menu's commands act on")
	closeContextMenu(c, screen, wnd)
}

// TestDuplicatingAnOptionKeepsThePick verifies that an option duplicated within a choice in an editor arrives turned
// off, so that the choice keeps its pick, and modifies the item its original does with the same data owner. What a
// structural undo gives back still belongs to the item: the pick is still locked, and a weapon still has its owner.
func TestDuplicatingAnOptionKeepsThePick(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Modifiers = []*gurps.TraitModifier{newTraitModifierChoiceFor(entity, true, []string{"A", "B"}, "A")}
	trait.Weapons = []*gurps.Weapon{gurps.NewWeapon(trait, true)}
	e, content := buildEditorContent(sheet, trait, initTraitEditor)
	panel, ok := uxtest.FirstPanelOfType[*traitModifiersPanel](content)
	c.True(ok)
	table := panel.table
	choice := e.editorData.Modifiers[0]
	table.SetSelectionMap(map[tid.TID]bool{choice.Children[0].ID(): true})
	DuplicateSelection(table)
	c.Equal(3, len(choice.Children))
	c.True(choice.Children[0].Enabled(), "the pick the choice had is kept")
	c.False(choice.Children[1].Enabled(), "the duplicate arrives turned off")
	c.Equal(trait, choice.Children[1].Target(), "the duplicate modifies the trait its original does")
	c.True(gurps.IsOnSheet(choice.Children[1]))

	unison.UndoManagerFor(table).Undo()
	table = liveTable(table)
	options := table.RootRows()[0].Data().Children
	c.Equal(2, len(options), "undo takes the duplicate away")
	c.False(options[1].Enabled())
	pick := options[0]
	c.True(pick.Enabled(), "undo keeps the pick")
	c.True(gurps.IsOnSheet(pick), "the restored modifiers keep the sheet as their owner")
	c.Equal(trait, pick.Target(), "the restored modifiers keep modifying the trait")
	table.SetSelectionMap(map[tid.TID]bool{pick.ID(): true})
	c.False(table.CanPerformCmd(nil, ToggleStateItemID), "the pick is still locked")
	adjustModifierEnabled(e, table, pick, false)
	c.True(pick.Enabled(), "and can't be turned off")

	weapons, ok := uxtest.FirstPanelOfType[*weaponsPanel](content)
	c.True(ok)
	weaponTable := weapons.table
	weaponTable.SetSelectionMap(map[tid.TID]bool{e.editorData.Weapons[0].ID(): true})
	DuplicateSelection(weaponTable)
	unison.UndoManagerFor(weaponTable).Undo()
	c.Equal(1, len(e.editorData.Weapons))
	c.Equal(gurps.WeaponOwner(trait), e.editorData.Weapons[0].Owner, "the restored weapon keeps its owner")
}

// TestTakingThePickAwayLeavesTheChoiceFlagged verifies that deleting the pick of a mandatory choice on a sheet, or
// moving it out of the choice, leaves the choice without a pick, flagged as required, rather than picking another
// option in its place.
func TestTakingThePickAwayLeavesTheChoiceFlagged(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	required := func(choice *gurps.TraitModifier) bool {
		var data gurps.CellData
		choice.CellData(gurps.TraitModifierDescriptionColumn, &data)
		return data.ChoiceRequired
	}
	for _, remove := range []func(table *unison.Table[*Node[*gurps.TraitModifier]]){
		func(table *unison.Table[*Node[*gurps.TraitModifier]]) { DeleteSelection(table, true) },
		func(table *unison.Table[*Node[*gurps.TraitModifier]]) {
			table.PerformCmd(nil, MoveOutOfContainerItemID)
		},
	} {
		e, _, table := traitEditorOnSheet(t, sheet, newTraitModifierChoiceFor(entity, true, []string{"A", "B"}, "A"))
		choice := e.editorData.Modifiers[len(e.editorData.Modifiers)-1]
		c.False(required(choice))
		table.SetSelectionMap(map[tid.TID]bool{choice.Children[0].ID(): true})
		remove(table)
		choice = e.editorData.Modifiers[len(e.editorData.Modifiers)-1]
		c.Equal(1, len(choice.Children))
		c.False(choice.Children[0].Enabled(), "B isn't picked in A's place")
		c.True(required(choice), "the choice is flagged as required")
	}
}

// TestRowEditsReportThemselvesOnce verifies that the row-level edits that report themselves by rebuilding the sheet --
// deleting and duplicating rows, here -- do so exactly once and bump the modification timestamp as they go, which a
// bare rebuild used to leave out.
func TestRowEditsReportThemselvesOnce(t *testing.T) {
	c := check.New(t)
	sheet, trait := newSheetWithSwitchableTrait(t)
	entity := sheet.Entity()
	before := len(entity.Traits)

	sheet.Traits.Table.SetSelectionMap(map[tid.TID]bool{trait.ID(): true})
	entity.ModifiedOn = jio.Time{}
	counter := installSyncCounter(sheet)
	DuplicateSelection(sheet.Traits.Table)
	c.Equal(before+1, len(entity.Traits), "duplicating must add a row")
	c.Equal(1, counter.count, "duplicating must sync the sheet exactly once")
	c.NotEqual(jio.Time{}, entity.ModifiedOn, "duplicating must bump the modification timestamp")

	sheet.Traits.Table.SetSelectionMap(map[tid.TID]bool{trait.ID(): true})
	entity.ModifiedOn = jio.Time{}
	counter.count = 0
	DeleteSelection(sheet.Traits.Table, true)
	c.Equal(before, len(entity.Traits), "deleting must remove the row")
	c.Equal(1, counter.count, "deleting must sync the sheet exactly once")
	c.NotEqual(jio.Time{}, entity.ModifiedOn, "deleting must bump the modification timestamp")
}
