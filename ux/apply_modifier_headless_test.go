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
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

// rowInView reports whether the row at the given index of the table is within the part of the table that can be seen,
// judged top to bottom only, since a row is as wide as its table and the table may be wider than the view.
func rowInView[T gurps.Node[T]](table *unison.Table[*Node[T]], row int) bool {
	frame := table.RectToRoot(table.RowFrame(row))
	visible := visibleRect(table.AsPanel())
	return frame.Y >= visible.Y && frame.Bottom() <= visible.Bottom()
}

// TestApplyModifierHeadless drives the Apply Modifier command through its menus and dialogs with a sheet, a template
// and a trait modifier library open: the command is in the modifier list's context menu but not a trait list's; the
// destination prompt allows one choice, with the first selected so that Return alone accepts it; the target prompt
// names each trait with its container and allows several; canceling either prompt changes nothing; accepting attaches
// an enabled copy of the selected modifier to the chosen traits, opens their container and focuses the traits list
// with them selected and in view; Undo and Redo take the copies off and put them back, container included; and with
// only the sheet open the command goes straight to the target prompt.
func TestApplyModifierHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	screen.EnableAccessibility()
	forbidModifierPrompts(t)

	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	traitNames := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon"}
	var delta *gurps.Trait
	screen.Do(func() {
		entity := sheet.Entity()
		entity.Traits = nil
		for _, name := range traitNames[:3] {
			trait := gurps.NewTrait(entity, nil, false)
			trait.Name = name
			entity.Traits = append(entity.Traits, trait)
		}
		// A closed container holding a trait, so that the prompt has a nested row to name and the command has a
		// container to open.
		delta = gurps.NewTrait(entity, nil, true)
		delta.Name = "Delta"
		delta.SetOpen(false)
		epsilon := gurps.NewTrait(entity, delta, false)
		epsilon.Name = "Epsilon"
		delta.Children = []*gurps.Trait{epsilon}
		entity.Traits = append(entity.Traits, delta)
		entity.Recalculate()
		sheet.Rebuild(true)
		// A new sheet counts as modified until saved, so it is marked unmodified and its timestamp cleared for the test
		// to tell whether the command changed and reported anything.
		sheet.markUnmodified()
		entity.ModifiedOn = jio.Time{}
	})
	template, ok := openedByAction(t, screen, newCharacterTemplateAction).(*Template)
	if !ok {
		t.Fatal("New Character Template must open a template")
	}
	screen.Do(func() {
		// A template with nothing in it has nothing to apply a modifier to, so it would not be offered.
		trait := gurps.NewTrait(template.template, nil, false)
		trait.Name = "Template Trait"
		template.template.Traits = []*gurps.Trait{trait}
		template.Rebuild(true)
		template.markUnmodified() // So that closing it later asks nothing.
	})
	ranged := gurps.NewTraitModifier(nil, nil, false)
	ranged.Name = "Ranged"
	ranged.Disabled = true
	melee := gurps.NewTraitModifier(nil, nil, false)
	melee.Name = "Melee"
	var library *TableDockable[*gurps.TraitModifier]
	screen.Do(func() {
		library = NewTraitModifierTableDockable(filepath.Join(t.TempDir(), "mods"+gurps.TraitModifiersExt),
			[]*gurps.TraitModifier{ranged, melee})
		DisplayNewDockable(library)
	})
	if library == nil {
		t.Fatal("the modifier library could not be opened")
	}
	// traitModifierNames returns the modifier names of the sheet's trait with the given name, failing the test if there
	// is none.
	traitModifierNames := func(name string) []string {
		t.Helper()
		var names []string
		found := false
		screen.Do(func() {
			gurps.Traverse(func(trait *gurps.Trait) bool {
				if trait.Name == name {
					names = appliedModifierNames(trait.Modifiers)
					found = true
				}
				return found
			}, false, false, sheet.Entity().Traits...)
		})
		if !found {
			t.Fatalf("the sheet has no trait named %q", name)
		}
		return names
	}
	nothingApplied := func(why string) {
		t.Helper()
		var canUndo, modified bool
		var modifiedOn jio.Time
		screen.Do(func() {
			canUndo = sheet.undoMgr.CanUndo()
			modified = sheet.Modified()
			modifiedOn = sheet.Entity().ModifiedOn
		})
		c.False(canUndo, "%s: there must be nothing to undo", why)
		c.False(modified, "%s: the sheet must be left unmodified", why)
		c.Equal(jio.Time{}, modifiedOn, "%s: nothing must have been reported to the sheet", why)
		for _, name := range traitNames {
			c.Equal(0, len(traitModifierNames(name)), "%s: %s must have no modifiers", why, name)
		}
	}
	windowCount := func() int {
		var count int
		screen.Do(func() { count = len(unison.Windows()) })
		return count
	}
	// selectModifier clicks the library's first row, Ranged, which selects it alone and puts the focus on the list.
	selectModifier := func() {
		t.Helper()
		screen.Click(cellScreenPoint(t, screen, wnd, library.table, 0, gurps.TraitModifierDescriptionColumn))
		var enabled bool
		screen.Do(func() {
			enabled = wnd.Focus() == library.table.AsPanel() && library.table.HasSelection() &&
				applyModifierAction.Enabled(nil)
		})
		c.True(enabled, "with a modifier selected and a destination open, the command must be available")
	}
	// destinationPrompt chooses the command from the Edit menu and returns the destination prompt it puts up, along
	// with the prompt's list and the row of the sheet in it.
	destinationPrompt := func() (dialogWnd *unison.Window, dialog *unison.Dialog, list *unison.List[FileBackedDockable], sheetRow int) {
		t.Helper()
		selectModifier()
		chooseMenuBarItem(t, screen, wnd, "Edit", applyModifierAction.Title)
		dialogWnd, dialog = modalDialog(t, screen, wnd)
		sheetRow = -1
		screen.Do(func() {
			if lists := panelsOfType[*unison.List[FileBackedDockable]](dialogWnd.Content()); len(lists) == 1 {
				list = lists[0]
				for i := range list.Count() {
					if list.DataAtIndex(i).AsPanel() == sheet.AsPanel() {
						sheetRow = i
					}
				}
			}
		})
		if list == nil {
			t.Fatal("the destination prompt must hold one list of destinations")
		}
		if sheetRow == -1 {
			t.Fatal("the destination prompt must offer the sheet")
		}
		return dialogWnd, dialog, list, sheetRow
	}
	// chooseSheet picks the sheet in the destination prompt and accepts it with Return.
	chooseSheet := func(dialogWnd *unison.Window, _ *unison.Dialog, list *unison.List[FileBackedDockable], sheetRow int) {
		t.Helper()
		var rowPt geom.Point
		screen.Do(func() { rowPt = screenPoint(dialogWnd, list.RectToRoot(list.RowRect(sheetRow)).Center()) })
		screen.Click(rowPt)
		var selected int
		screen.Do(func() { selected = list.Selection.FirstSet() })
		c.Equal(sheetRow, selected, "clicking the sheet's row must select it")
		screen.KeyPress(unison.KeyReturn, mod.None)
	}
	// targetPromptFor returns the target prompt that is up and its list, checking that its header names the given
	// destination.
	targetPromptFor := func(title string) (*unison.Window, *unison.Dialog, *unison.List[modifierTargetChoice[*gurps.Trait]]) {
		t.Helper()
		dialogWnd, dialog := modalDialog(t, screen, wnd)
		var targets *unison.List[modifierTargetChoice[*gurps.Trait]]
		var headerFound bool
		screen.Do(func() {
			if lists := panelsOfType[*unison.List[modifierTargetChoice[*gurps.Trait]]](dialogWnd.Content()); len(lists) == 1 {
				targets = lists[0]
			}
			for _, label := range panelsOfType[*unison.Label](dialogWnd.Content()) {
				if strings.Contains(label.String(), "Choose the traits in "+title) {
					headerFound = true
				}
			}
		})
		if targets == nil {
			t.Fatal("the target prompt must hold one list of traits")
		}
		c.True(headerFound, "the prompt's header must name the destination that was chosen, %s", title)
		return dialogWnd, dialog, targets
	}
	// targetPrompt returns the sheet's target prompt, which must be the one showing.
	targetPrompt := func() (*unison.Window, *unison.Dialog, *unison.List[modifierTargetChoice[*gurps.Trait]]) {
		t.Helper()
		return targetPromptFor(sheet.Title())
	}
	// sheetState records what the command leaves behind on the sheet.
	type sheetState struct {
		selected  []string
		deltaOpen bool
		focused   bool
	}
	currentSheetState := func() sheetState {
		var s sheetState
		screen.Do(func() {
			s.deltaOpen = delta.IsOpen()
			s.focused = wnd.Focus() == sheet.Traits.Table.AsPanel()
			for _, row := range sheet.Traits.Table.SelectedRows(false) {
				s.selected = append(s.selected, row.Data().Name)
			}
		})
		return s
	}

	// The command is offered where the modifier is, and not on a list it can't work from.
	selectModifier()
	screen.KeyPress(unison.KeyF10, mod.Shift)
	c.True(slices.Contains(contextMenuTitles(t, screen, wnd), applyModifierAction.Title),
		"the modifier list's context menu must offer the command")
	closeContextMenu(c, screen, wnd)
	screen.Click(cellScreenPoint(t, screen, wnd, sheet.Traits.Table, 0, gurps.TraitDescriptionColumn))
	screen.KeyPress(unison.KeyF10, mod.Shift)
	traitsMenu := contextMenuTitles(t, screen, wnd)
	c.True(len(traitsMenu) != 0, "the traits list must have opened a context menu for the check to mean anything")
	c.False(slices.Contains(traitsMenu, applyModifierAction.Title), "a list of traits has no modifier to apply")
	closeContextMenu(c, screen, wnd)

	// The destination prompt offers the sheet and the template, one choosable, with the first already selected so that
	// Return alone accepts it. Canceling either prompt changes nothing.
	dialogWnd, _, list, _ := destinationPrompt()
	var titles []string
	var multiple, headerFound bool
	var preselected int
	screen.Do(func() {
		multiple = list.AllowMultipleSelection()
		preselected = list.Selection.FirstSet()
		for i := range list.Count() {
			titles = append(titles, list.DataAtIndex(i).Title())
		}
		for _, label := range panelsOfType[*unison.Label](dialogWnd.Content()) {
			if label.String() == "Choose a destination:" {
				headerFound = true
			}
		}
	})
	c.False(multiple, "only one destination may be chosen")
	c.Equal(0, preselected, "the first destination starts out selected")
	c.True(headerFound, "the prompt asks for a destination")
	first := titles[0]
	slices.Sort(titles)
	wanted := []string{sheet.Title(), template.Title()}
	slices.Sort(wanted)
	c.Equal(wanted, titles, "the prompt offers everything open that holds traits, and nothing else")
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal(1, windowCount(), "the prompt has been dismissed")
	nothingApplied("canceling the destination prompt")
	destinationPrompt()
	screen.KeyPress(unison.KeyReturn, mod.None)
	targetPromptFor(first)
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal(1, windowCount(), "the prompt has been dismissed")
	nothingApplied("canceling the target prompt Return led to")

	// The target prompt offers the sheet's traits by name, the nested one with its container, and lets several be
	// chosen. Canceling it changes nothing.
	chooseSheet(destinationPrompt())
	dialogWnd, _, targets := targetPrompt()
	var labels []string
	screen.Do(func() {
		multiple = targets.AllowMultipleSelection()
		for i := range targets.Count() {
			labels = append(labels, targets.DataAtIndex(i).String())
		}
	})
	c.True(multiple, "more than one target may be chosen")
	c.Equal([]string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon (in Delta)"}, labels,
		"the prompt offers the sheet's traits, naming the container a nested one sits in")
	// A row hands its text over to the label described within it, so what is heard for a row is the label's text.
	var spoken []string
	tree := screen.AccessibilityTree(dialogWnd)
	tree.Walk(func(n *accessibility.Node) bool {
		if n.Role == role.ListItem {
			text := n.Name
			for _, id := range tree.UnignoredChildren(n.ID) {
				if child := tree.Node(id); child != nil {
					text += child.Name
				}
			}
			spoken = append(spoken, text)
		}
		return true
	})
	c.Equal(labels, spoken, "a screen reader hears each row by its label, container and all")
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal(1, windowCount(), "the prompt has been dismissed")
	nothingApplied("canceling the target prompt")

	// Accepting the target prompt with Alpha and Epsilon chosen applies the selected modifier to both, opening Delta to
	// show Epsilon.
	chooseSheet(destinationPrompt())
	_, _, targets = targetPrompt()
	screen.Do(func() { targets.Select(false, 0, 4) })
	screen.KeyPress(unison.KeyReturn, mod.None)
	var modified, enabled bool
	var modifiedOn jio.Time
	var undoTitle string
	var selectedInView, selectedRows int
	screen.Do(func() {
		modified = sheet.Modified()
		modifiedOn = sheet.Entity().ModifiedOn
		undoTitle = sheet.undoMgr.UndoTitle()
		for _, trait := range sheet.Entity().Traits {
			if trait.Name == "Alpha" && len(trait.Modifiers) == 1 {
				enabled = trait.Modifiers[0].Enabled()
			}
		}
		table := sheet.Traits.Table
		for row := range table.LastRowIndex() + 1 {
			if table.IsRowSelected(row) {
				selectedRows++
				if rowInView(table, row) {
					selectedInView++
				}
			}
		}
	})
	c.Equal(1, windowCount(), "the prompt has been dismissed")
	c.Equal([]string{"Ranged"}, traitModifierNames("Alpha"), "only the selected modifier is applied")
	c.Equal(0, len(traitModifierNames("Beta")), "Beta was not chosen")
	c.Equal(0, len(traitModifierNames("Gamma")), "Gamma was not chosen")
	c.Equal(0, len(traitModifierNames("Delta")), "Delta was not chosen")
	c.Equal([]string{"Ranged"}, traitModifierNames("Epsilon"), "only the selected modifier is applied")
	c.True(enabled, "the applied copy is switched on although the library's modifier is off")
	c.True(modified, "the sheet has been changed")
	c.NotEqual(jio.Time{}, modifiedOn, "the change has been reported to the sheet")
	c.Equal(sheetState{selected: []string{"Alpha", "Epsilon"}, deltaOpen: true, focused: true}, currentSheetState(),
		"the container is opened and the focus lands on the sheet's traits list with the targets selected")
	c.Equal(2, selectedRows, "both targets must be selected for the check to mean anything")
	c.Equal(selectedRows, selectedInView, "both targets are in view")
	c.Equal("Undo "+applyModifierAction.Title, undoTitle)
	var templateModifiers int
	screen.Do(func() { templateModifiers = len(template.template.Traits[0].Modifiers) })
	c.Equal(0, templateModifiers, "the template, which was not chosen, is left alone")

	// Edit > Undo undoes the command in one step, closing the container it opened, and Edit > Redo puts everything back
	// as the command left it.
	chooseMenuBarItem(t, screen, wnd, "Edit", "Undo "+applyModifierAction.Title)
	var canUndo, canRedo bool
	screen.Do(func() {
		canUndo = sheet.undoMgr.CanUndo()
		canRedo = sheet.undoMgr.CanRedo()
	})
	c.Equal(0, len(traitModifierNames("Alpha")), "undo must take the modifier back off")
	c.Equal(0, len(traitModifierNames("Epsilon")), "undo must take the modifier back off")
	c.False(currentSheetState().deltaOpen, "undo must close the container the command opened")
	c.False(canUndo, "one command must make exactly one edit")
	c.True(canRedo, "the command must be redoable")
	chooseMenuBarItem(t, screen, wnd, "Edit", "Redo "+applyModifierAction.Title)
	c.Equal([]string{"Ranged"}, traitModifierNames("Alpha"), "redo must put the modifier back")
	c.Equal([]string{"Ranged"}, traitModifierNames("Epsilon"), "redo must put the modifier back")
	c.Equal(sheetState{selected: []string{"Alpha", "Epsilon"}, deltaOpen: true, focused: true}, currentSheetState(),
		"redo must show the targets as the command did")
	chooseMenuBarItem(t, screen, wnd, "Edit", "Undo "+applyModifierAction.Title)
	var stillModified bool
	screen.Do(func() {
		stillModified = sheet.Modified()
		sheet.Entity().ModifiedOn = jio.Time{} // Undo puts the data back, not the timestamp the command's report set.
	})
	c.False(stillModified, "undoing the command must leave the sheet as it was")

	// With the template closed, the sheet is the only destination, so the command asks for its traits straight away.
	closeEditorWithoutPrompt(t, screen, template)
	var destinations int
	screen.Do(func() { destinations = len(modifierDestinations(traitModifierTargetKind(), AllDockables())) })
	c.Equal(1, destinations, "only the sheet must be left as a destination for the check to mean anything")
	selectModifier()
	chooseMenuBarItem(t, screen, wnd, "Edit", applyModifierAction.Title)
	targetPrompt()
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal(1, windowCount(), "the prompt has been dismissed")
	nothingApplied("canceling the target prompt reached directly")
}

// TestApplyModifierHeadlessRaisesTheDestinationsWindow verifies that the command, chosen from a modifier library in a
// window of its own, ends with the destination's window in front and the focus on the list that received the modifier,
// so that the next Undo goes to the sheet rather than the library.
func TestApplyModifierHeadlessRaisesTheDestinationsWindow(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	forbidModifierPrompts(t)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	screen.Do(func() {
		entity := sheet.Entity()
		trait := gurps.NewTrait(entity, nil, false)
		trait.Name = "Alpha"
		entity.Traits = []*gurps.Trait{trait}
		entity.Recalculate()
		sheet.Rebuild(true)
		sheet.markUnmodified()
	})
	ranged := gurps.NewTraitModifier(nil, nil, false)
	ranged.Name = "Ranged"
	var library *TableDockable[*gurps.TraitModifier]
	var libraryWnd *unison.Window
	screen.Do(func() {
		library = NewTraitModifierTableDockable(filepath.Join(t.TempDir(), "mods"+gurps.TraitModifiersExt),
			[]*gurps.TraitModifier{ranged})
		var err error
		if libraryWnd, err = NewWindowForDockable(library, dgroup.Libraries); err != nil {
			t.Errorf("unable to open the library in a window of its own: %v", err)
			return
		}
		// A window packed around one modifier is too small to hold its Edit menu's popup, which would then open
		// past its edge, over the workspace window, where a click on an item lands in the workspace instead.
		libraryWnd.SetFrameRect(geom.NewRect(200, 100, 900, 700))
		libraryWnd.ToFront()
	})
	if libraryWnd == nil {
		t.Fatal("the modifier library could not be opened in a window of its own")
	}
	alphaModifiers := func() []string {
		var names []string
		screen.Do(func() { names = appliedModifierNames(sheet.Entity().Traits[0].Modifiers) })
		return names
	}

	screen.Click(cellScreenPoint(t, screen, libraryWnd, library.table, 0, gurps.TraitModifierDescriptionColumn))
	c.Equal(libraryWnd, screen.FocusedWindow(), "the library's window must hold the focus for the check to mean anything")
	var enabled bool
	screen.Do(func() { enabled = library.table.HasSelection() && applyModifierAction.Enabled(nil) })
	c.True(enabled, "with a modifier selected and the sheet open, the command must be available")

	// The sheet being the only destination, the command asks for its traits straight away.
	chooseMenuBarItem(t, screen, libraryWnd, "Edit", applyModifierAction.Title)
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	var picked bool
	screen.Do(func() {
		if lists := panelsOfType[*unison.List[modifierTargetChoice[*gurps.Trait]]](dialogWnd.Content()); len(lists) == 1 {
			lists[0].Select(false, 0)
			picked = true
		}
	})
	if !picked {
		t.Fatal("the target prompt must hold one list of the sheet's traits")
	}
	screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseOK)))
	c.Equal([]string{"Ranged"}, alphaModifiers(), "the modifier must have been applied")
	c.Equal(wnd, screen.FocusedWindow(), "the sheet's window must have been brought to the front")
	var focused bool
	screen.Do(func() { focused = wnd.Focus() == sheet.Traits.Table.AsPanel() })
	c.True(focused, "the focus must land on the sheet's traits list")

	// The focus being on the sheet, its window's Edit > Undo undoes the command.
	chooseMenuBarItem(t, screen, wnd, "Edit", "Undo "+applyModifierAction.Title)
	c.Equal(0, len(alphaModifiers()), "undo must take the modifier back off")
	screen.Do(func() { libraryWnd.Dispose() })
}

// TestRevealModifierTargetsFocusesTheFirstListWithASelection verifies that with targets in both of a sheet's equipment
// lists, the focus lands on the first list, with its target in view.
func TestRevealModifierTargetsFocusesTheFirstListWithASelection(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	forbidModifierPrompts(t)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var rope, coin *gurps.Equipment
	screen.Do(func() {
		entity := sheet.Entity()
		rope = gurps.NewEquipment(entity, nil, false)
		rope.Name = "Rope"
		entity.CarriedEquipment = []*gurps.Equipment{rope}
		coin = gurps.NewEquipment(entity, nil, false)
		coin.Name = "Coin"
		entity.OtherEquipment = []*gurps.Equipment{coin}
		entity.Recalculate()
		sheet.Rebuild(true)
		sheet.markUnmodified()
	})
	fine := gurps.NewEquipmentModifier(nil, nil, false)
	fine.Name = "Fine"
	var applied bool
	screen.Do(func() {
		tables := []*unison.Table[*Node[*gurps.Equipment]]{sheet.CarriedEquipment.Table, sheet.OtherEquipment.Table}
		applied = applyModifiersTo(tables, []*gurps.Equipment{rope, coin}, []*gurps.EquipmentModifier{fine},
			gurps.LibraryFile{})
		revealModifierTargets(sheet, tables)
	})
	c.True(applied, "the modifier must have been applied")
	selectedNames := func(table *unison.Table[*Node[*gurps.Equipment]]) []string {
		rows := table.SelectedRows(false)
		names := make([]string, 0, len(rows))
		for _, row := range rows {
			names = append(names, row.Data().Name)
		}
		return names
	}
	var carriedSelected, otherSelected []string
	var carriedFocused, otherFocused, inView bool
	screen.Do(func() {
		carried, other := sheet.CarriedEquipment.Table, sheet.OtherEquipment.Table
		carriedSelected = selectedNames(carried)
		otherSelected = selectedNames(other)
		carriedFocused = wnd.Focus() == carried.AsPanel()
		otherFocused = wnd.Focus() == other.AsPanel()
		if row := carried.FirstSelectedRowIndex(); row != -1 {
			inView = rowInView(carried, row)
		}
	})
	c.Equal([]string{"Rope"}, carriedSelected, "the carried target must be selected")
	c.Equal([]string{"Coin"}, otherSelected, "the other target must be selected")
	c.True(carriedFocused, "the focus must land on the first list holding a target")
	c.False(otherFocused, "the focus must not move on to the second list")
	c.True(inView, "the first list's target must be in view")
	screen.Do(func() { sheet.undoMgr.Undo() })
}

// TestRevealModifierTargetsScrollsToATargetBelowTheOldEndOfTheList verifies that a target far down inside a closed
// container is scrolled into view, which needs the layout validated after opening the container, or the scroll would
// be clamped to the old end of the list.
func TestRevealModifierTargetsScrollsToATargetBelowTheOldEndOfTheList(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	forbidModifierPrompts(t)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var last *gurps.Trait
	var lastRowBefore int
	screen.Do(func() {
		entity := sheet.Entity()
		entity.Traits = nil
		for i := range 10 {
			trait := gurps.NewTrait(entity, nil, false)
			trait.Name = fmt.Sprintf("Trait %d", i)
			entity.Traits = append(entity.Traits, trait)
		}
		bag := gurps.NewTrait(entity, nil, true)
		bag.Name = "Bag"
		bag.SetOpen(false)
		for i := range 60 {
			last = gurps.NewTrait(entity, bag, false)
			last.Name = fmt.Sprintf("Item %d", i)
			bag.Children = append(bag.Children, last)
		}
		entity.Traits = append(entity.Traits, bag)
		entity.Recalculate()
		sheet.Rebuild(true)
		sheet.markUnmodified()
		lastRowBefore = sheet.Traits.Table.LastRowIndex()
	})
	c.Equal(10, lastRowBefore, "the list must start out showing the ten traits and the closed container")
	ranged := gurps.NewTraitModifier(nil, nil, false)
	ranged.Name = "Ranged"
	var applied bool
	screen.Do(func() {
		tables := []*unison.Table[*Node[*gurps.Trait]]{sheet.Traits.Table}
		applied = applyModifiersTo(tables, []*gurps.Trait{last}, []*gurps.TraitModifier{ranged}, gurps.LibraryFile{})
		revealModifierTargets(sheet, tables)
	})
	c.True(applied, "the modifier must have been applied")
	// The state is read in a second round, once the scroll the reveal asked for has been laid out.
	var row int
	var inView bool
	screen.Do(func() {
		live := sheet.Traits.Table
		row = live.FirstSelectedRowIndex()
		if row != -1 {
			inView = rowInView(live, row)
		}
	})
	c.Equal(70, row, "the last item in the container must be the row selected")
	c.True(inView, "the selected target must have been scrolled into view")
	screen.Do(func() { sheet.undoMgr.Undo() })
}
