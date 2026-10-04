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
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// stringFieldLabeled returns the string field that follows the label bearing the given text among the panel's direct
// children, or nil if there is none.
func stringFieldLabeled(p *unison.Panel, labelText string) *StringField {
	children := p.Children()
	for i, child := range children {
		if label, ok := child.Self.(*unison.Label); ok && label.String() == labelText && i+1 < len(children) {
			if field, ok2 := children[i+1].Self.(*StringField); ok2 {
				return field
			}
		}
	}
	return nil
}

// openSourceMenu clicks the editor's source button and returns the panels of the items of the menu that opens.
func openSourceMenu(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window, d unison.Dockable) []*unison.Panel {
	t.Helper()
	var button *unison.Button
	screen.Do(func() { button = buttonWithSVG(d.AsPanel(), svg.Database) })
	if button == nil {
		t.Fatal("the editor has no source button")
	}
	screen.Click(screen.PanelCenter(button))
	var items []*unison.Panel
	screen.Do(func() { items = slices.Clone(menuItemPanels(openMenuPopup(wnd))) })
	if len(items) == 0 {
		t.Fatal("the source button opened no menu")
	}
	return items
}

// TestEditorSourceMenu drives the source menu of a trait editor as a user would: the button names itself, the menu
// shows the trait's ID and source and how it compares to the library, choosing the ID copies it, a command that doesn't
// apply is disabled and does nothing, and a source cleared from the menu stays put when the editor's changes are
// discarded. A weapon's editor has no such menu, since a weapon has no source.
func TestEditorSourceMenu(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	trait := gurps.NewTrait(sheet.Entity(), nil, false)
	trait.Name = "Claws"
	trait.Source = testSource
	custom := gurps.NewTrait(sheet.Entity(), nil, false)
	custom.Name = "Fangs"
	var e, customEditor *editor[*gurps.Trait, *gurps.TraitEditData]
	var tip string
	screen.Do(func() {
		sheet.Entity().Traits = append(sheet.Entity().Traits, trait, custom)
		sheet.Rebuild(true)
		customEditor = EditTrait(sheet, custom)
		if button := buttonWithSVG(customEditor.AsPanel(), svg.Database); button != nil {
			tip = tooltipText(button.Tooltip)
		}
	})
	c.Equal("ID & Library Source", tip, "the source button is named by its tooltip")

	// clearSource has no check of its own, so only the menu keeps it from a trait without a source.
	items := openSourceMenu(t, screen, wnd, customEditor)
	c.Equal([]string{
		"ID: " + string(custom.ID()),
		srcstate.Custom.String(),
		"Sync with Source",
		"Clear Source",
	}, contextMenuTitles(t, screen, wnd))
	clearItem := items[len(items)-1]
	node := screen.AccessibilityNodeFor(clearItem)
	c.True(node != nil && node.Disabled, "a trait without a source has none to clear")
	screen.Click(screen.PanelCenter(clearItem))
	var modified bool
	screen.Do(func() { modified = customEditor.isModified() })
	c.False(modified, "choosing a disabled command does nothing")
	closeContextMenu(c, screen, wnd)
	closeEditorWithoutPrompt(t, screen, customEditor)

	screen.Do(func() { e = EditTrait(sheet, trait) })
	items = openSourceMenu(t, screen, wnd, e)
	c.Equal([]string{
		"ID: " + string(trait.ID()),
		"Source ID: " + string(testSource.TID),
		"Source Library: " + testSource.Library,
		"Source Path: " + testSource.Path,
		srcstate.Missing.String(),
		"Sync with Source",
		"Clear Source",
	}, contextMenuTitles(t, screen, wnd))
	node = screen.AccessibilityNodeFor(items[len(items)-2])
	c.True(node != nil && node.Disabled, "a source no library holds can't be synced with")
	closeContextMenu(c, screen, wnd)

	items = openSourceMenu(t, screen, wnd, e)
	screen.Click(screen.PanelCenter(items[0]))
	var clip string
	screen.Do(func() { clip = unison.ClipboardGetText() })
	c.Equal(string(trait.ID()), clip, "choosing the ID copies it")

	items = openSourceMenu(t, screen, wnd, e)
	screen.Click(screen.PanelCenter(items[len(items)-1]))
	var cancelEnabled bool
	screen.Do(func() {
		modified = e.isModified()
		cancelEnabled = e.cancelButton.Enabled()
	})
	c.True(modified, "clearing the source is a change to apply")
	c.True(cancelEnabled, "or to discard")
	c.Equal(testSource, trait.Source, "the trait keeps its source until the change is applied")
	screen.Click(screen.PanelCenter(e.cancelButton))
	var open bool
	screen.Do(func() {
		open = slices.ContainsFunc(AllDockables(), func(d unison.Dockable) bool { return d.AsPanel().Self == e })
	})
	c.False(open, "Discard closes the editor")
	c.Equal(testSource, trait.Source, "and the trait keeps its source")

	var weaponEditor unison.Dockable
	var hasButton bool
	screen.Do(func() {
		before := AllDockables()
		EditWeapon(sheet, gurps.NewWeapon(trait, true))
		for _, d := range AllDockables() {
			if !slices.Contains(before, d) {
				weaponEditor = d
			}
		}
		if weaponEditor != nil {
			hasButton = buttonWithSVG(weaponEditor.AsPanel(), svg.Database) != nil
		}
	})
	if weaponEditor == nil {
		t.Fatal("the weapon editor did not open")
	}
	c.False(hasButton, "a weapon has no source, so its editor has no source menu")
	if closer, isCloser := weaponEditor.(interface {
		unison.Dockable
		unison.TabCloser
	}); isCloser {
		closeEditorWithoutPrompt(t, screen, closer)
	}
}

// newEditedLibraryTrait saves a trait named "Claws", with a melee weapon, further altered by prepareLib if not nil, into
// a file of the user library, adds a copy sourced from that file to the sheet as "Claws (old)", further altered by
// prepare if not nil, and opens the copy in an editor.
func newEditedLibraryTrait(t *testing.T, c check.Checker, screen *unison.HeadlessScreen, sheet *Sheet, prepareLib, prepare func(trait *gurps.Trait)) (*editor[*gurps.Trait, *gurps.TraitEditData], *gurps.Trait) {
	t.Helper()
	user := gurps.GlobalSettings().Libraries.User()
	lib := gurps.NewTrait(nil, nil, false)
	lib.Name = "Claws"
	lib.Weapons = []*gurps.Weapon{gurps.NewWeapon(lib, true)}
	if prepareLib != nil {
		prepareLib(lib)
	}
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Traits/Test" + gurps.TraitsExt}
	c.NoError(gurps.SaveTraits([]*gurps.Trait{lib}, filepath.Join(user.Path(), filepath.FromSlash(libFile.Path))))
	var e *editor[*gurps.Trait, *gurps.TraitEditData]
	var local *gurps.Trait
	screen.Do(func() {
		entity := sheet.Entity()
		local = lib.Clone(libFile, entity, nil, gurps.Reference)
		local.Name = "Claws (old)"
		if prepare != nil {
			prepare(local)
		}
		entity.Traits = append(entity.Traits, local)
		sheet.Rebuild(true)
		e = EditTrait(sheet, local)
	})
	return e, local
}

// editorField returns the string field of the editor's content under the label, ending the test if there is none.
func editorField[N gurps.Node[N], D gurps.EditorData[N]](t *testing.T, screen *unison.HeadlessScreen, e *editor[N, D], label string) *StringField {
	t.Helper()
	var field *StringField
	screen.Do(func() { field = stringFieldLabeled(e.content, label) })
	if field == nil {
		t.Fatalf("the editor has no %s field", label)
	}
	return field
}

// chooseSyncWithSource chooses Sync with Source from the editor's source menu, without reading the menu's items as
// contextMenuTitles would, since that turns accessibility support on, after which clicking a button gives it the focus.
func chooseSyncWithSource(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window, d unison.Dockable) {
	t.Helper()
	items := openSourceMenu(t, screen, wnd, d)
	screen.Click(screen.PanelCenter(items[len(items)-2]))
}

// TestEditorSyncKeepsTheEditInProgress verifies that syncing from the source menu while a field holds the focus keeps
// what was typed into it when the sync leaves that field alone, and leaves the focus, the caret and the editor's
// scrolling as they were, so that typing carries on there. The field gives the focus up before it is replaced, while it
// is still part of the editor.
func TestEditorSyncKeepsTheEditInProgress(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	// Enlarged, so that the editor's content is bigger than its view, which can then be scrolled past the field.
	swapForTest(t, &gurps.GlobalSettings().General.InitialEditorUIScale, 300)
	e, local := newEditedLibraryTrait(t, c, screen, sheet, nil, nil)
	vttNotes := editorField(t, screen, e, "VTT Notes")
	// Brought into view to be clicked on, since at this size the editor opens scrolled past the start of its fields.
	screen.Do(vttNotes.ScrollIntoView)
	// Near its start, since at this size the middle of the field lies beyond the window's edge.
	screen.Click(screen.PanelPoint(vttNotes, geom.Point{X: 4, Y: 4}))
	var focused bool
	screen.Do(func() { focused = wnd.CurrentFocus() == vttNotes.AsPanel() })
	c.True(focused, "precondition: the click focuses the field")
	screen.Type("Typed")
	screen.KeyPress(unison.KeyLeft, 0)
	screen.KeyPress(unison.KeyLeft, 0)

	// Scrolled past the field, since giving the focus to a field out of view scrolls it into view, which putting the
	// focus back must not do.
	var h, v float32
	var inView bool
	var lostFocusWithinEditor []bool
	screen.Do(func() {
		e.scroll.SetPosition(0, e.content.FrameRect().Height)
		h, v = e.scroll.Position()
		inView = !visibleRect(vttNotes.AsPanel()).Empty()
		lostFocus := vttNotes.LostFocusCallback
		vttNotes.LostFocusCallback = func() {
			lostFocusWithinEditor = append(lostFocusWithinEditor, unison.AncestorIsOrSelf(vttNotes, e.content))
			lostFocus()
		}
	})
	c.False(inView, "precondition: the field is scrolled out of view")

	chooseSyncWithSource(t, screen, wnd, e)
	var name, notes string
	var modified, focusOnNotes bool
	var selStart, selEnd int
	var hAfter, vAfter float32
	screen.Do(func() {
		name = e.editorData.Name
		notes = e.editorData.VTTNotes
		modified = e.isModified()
		rebuilt := stringFieldLabeled(e.content, "VTT Notes")
		focusOnNotes = rebuilt != nil && rebuilt != vttNotes && wnd.CurrentFocus() == rebuilt.AsPanel()
		if rebuilt != nil {
			selStart, selEnd = rebuilt.Selection()
		}
		hAfter, vAfter = e.scroll.Position()
	})
	c.Equal("Claws", name, "the sync brings the library's name across")
	c.Equal("Typed", notes, "and keeps what was typed into a field it leaves alone")
	c.True(modified)
	c.True(len(lostFocusWithinEditor) != 0 && !slices.Contains(lostFocusWithinEditor, false),
		"the field gives up the focus while it is still part of the editor: %v", lostFocusWithinEditor)
	c.True(focusOnNotes, "the focus is on the rebuilt field that had it")
	c.Equal(3, selStart, "with the caret where it was")
	c.Equal(3, selEnd, "and nothing selected")
	c.Equal(h, hAfter, "the editor stays scrolled as it was")
	c.Equal(v, vAfter)
	c.Equal("Claws (old)", local.Name, "the trait is left alone until the changes are applied")

	screen.Type("X")
	screen.KeyPress(unison.KeyReturn, mod.OSMenuCommand())
	var state srcstate.Value
	screen.Do(func() { state, _ = gurps.MatchSource(local) })
	c.Equal("Claws", local.Name, "applying the changes syncs the trait")
	c.Equal("TypXed", local.VTTNotes, "typing carried on where it left off")
	c.Equal(srcstate.Matched, state)
	// Applying the changes modified the sheet, which is to be left unmodified (see startHeadlessWorkspace).
	screen.Do(sheet.markUnmodified)
}

// TestEditorSyncPutsTheFocusBack verifies where the focus goes after a sync that changes what held it: a field whose
// text the sync replaced gets it back with all of its text selected, and when the control that held it is gone, as the
// popup of a modifier choice is once the sync has made the choice a group, the nearest control gets it.
func TestEditorSyncPutsTheFocusBack(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	e, _ := newEditedLibraryTrait(t, c, screen, sheet, nil, nil)
	nameField := editorField(t, screen, e, "Name")
	screen.Click(screen.PanelCenter(nameField))
	screen.KeyPress(unison.KeyEnd, 0)
	var focused bool
	var selStart, selEnd int
	screen.Do(func() {
		focused = wnd.CurrentFocus() == nameField.AsPanel()
		selStart, selEnd = nameField.Selection()
	})
	c.True(focused, "precondition: the click focuses the field")
	c.Equal(len("Claws (old)"), selStart, "precondition: the caret is at the end of the name")
	c.Equal(selStart, selEnd, "precondition: nothing is selected")

	chooseSyncWithSource(t, screen, wnd, e)
	var text string
	screen.Do(func() {
		rebuilt := stringFieldLabeled(e.content, "Name")
		focused = rebuilt != nil && rebuilt != nameField && wnd.CurrentFocus() == rebuilt.AsPanel()
		if rebuilt != nil {
			text = rebuilt.Text()
			selStart, selEnd = rebuilt.Selection()
		}
	})
	c.Equal("Claws", text, "the sync replaces the name")
	c.True(focused, "the focus is on the rebuilt field that had it")
	c.Equal(0, selStart, "with all of what the sync brought in selected")
	c.Equal(len("Claws"), selEnd)
	screen.Click(screen.PanelCenter(e.cancelButton))

	// A choice whose library copy is a group, with notes long enough to put what follows them out of view.
	user := gurps.GlobalSettings().Libraries.User()
	lib := gurps.NewTraitModifier(nil, nil, true)
	lib.Name = "Options"
	lib.LocalNotes = strings.Repeat("A line of notes.\n", 80)
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Modifiers/Test" + gurps.TraitModifiersExt}
	c.NoError(gurps.SaveTraitModifiers([]*gurps.TraitModifier{lib},
		filepath.Join(user.Path(), filepath.FromSlash(libFile.Path))))
	isChoiceEditor := func(d unison.Dockable) bool {
		_, isEditor := d.AsPanel().Self.(*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData])
		return isEditor
	}
	screen.Do(func() {
		entity := sheet.Entity()
		choice := gurps.NewTraitModifierChoice(entity, nil)
		choice.Name = lib.Name
		choice.LocalNotes = lib.LocalNotes
		choice.Source = gurps.Source{LibraryFile: libFile, TID: lib.TID}
		trait := gurps.NewTrait(entity, nil, false)
		trait.Modifiers = []*gurps.TraitModifier{choice}
		gurps.AttachModifiers(trait, trait.Modifiers)
		entity.Traits = append(entity.Traits, trait)
		sheet.Rebuild(true)
		EditTraitModifier(sheet, choice)
	})
	choiceEditor := soleEditor[*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]](t, screen, isChoiceEditor)
	var popup *unison.Panel
	var popupInView bool
	screen.Do(func() {
		if popups := panelsOfType[*unison.PopupMenu[string]](choiceEditor.content); len(popups) == 1 {
			popup = popups[0].AsPanel()
			popup.RequestFocus()
			focused = wnd.CurrentFocus() == popup
			choiceEditor.scroll.SetPosition(0, 0)
			popupInView = !visibleRect(popup).Empty()
		}
	})
	if popup == nil {
		t.Fatal("the choice's editor must have the popup that says what the choice asks for")
	}
	c.True(focused, "precondition: the popup holds the focus")
	c.False(popupInView, "precondition: the popup is scrolled out of view")

	chooseSyncWithSource(t, screen, wnd, choiceEditor)
	var isChoice, focusOnTags, tagsInView bool
	screen.Do(func() {
		isChoice = choiceEditor.editorData.IsChoice()
		if tags := stringFieldLabeled(choiceEditor.content, "Tags"); tags != nil {
			focusOnTags = wnd.CurrentFocus() == tags.AsPanel()
			tagsInView = fullyVisible(tags.AsPanel())
		}
	})
	c.False(isChoice, "the sync makes the choice a group, whose editor has no such popup")
	c.True(focusOnTags, "the focus goes to the field that is now where the popup was")
	c.True(tagsInView, "which is scrolled into view")
	screen.Click(screen.PanelCenter(choiceEditor.cancelButton))
}

// TestEditorSyncKeepsTheFocusInPlaceAsContentGrows verifies that when a sync grows the content ahead of the field that
// holds the focus, as longer notes from the library do, the editor scrolls along with the field, which stays where it
// was within the view rather than being pushed out of it.
func TestEditorSyncKeepsTheFocusInPlaceAsContentGrows(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	e, _ := newEditedLibraryTrait(t, c, screen, sheet, func(lib *gurps.Trait) {
		lib.LocalNotes = strings.Repeat("A line of notes.\n", 150)
	}, func(local *gurps.Trait) {
		local.LocalNotes = ""
		// Tall enough that the content doesn't fit the view even before the sync, so that it can be scrolled by all
		// of what it then grows by.
		local.UserDesc = strings.Repeat("A line of my own.\n", 20)
	})
	tags := editorField(t, screen, e, "Tags")
	screen.Do(tags.ScrollIntoView)
	screen.Click(screen.PanelCenter(tags))
	var focused bool
	var before geom.Rect
	screen.Do(func() {
		focused = wnd.CurrentFocus() == tags.AsPanel()
		before = visibleRect(tags.AsPanel())
	})
	c.True(focused, "precondition: the click focuses the field")

	chooseSyncWithSource(t, screen, wnd, e)
	var notes string
	var focusOnTags bool
	var v, viewHeight float32
	var after, whole geom.Rect
	screen.Do(func() {
		notes = e.editorData.LocalNotes
		if rebuilt := stringFieldLabeled(e.content, "Tags"); rebuilt != nil {
			focusOnTags = rebuilt != tags && wnd.CurrentFocus() == rebuilt.AsPanel()
			after = visibleRect(rebuilt.AsPanel())
			whole = rebuilt.RectToRoot(rebuilt.ContentRect(false))
		}
		_, v = e.scroll.Position()
		viewHeight = e.scroll.ContentView().ContentRect(false).Height
	})
	c.Equal(150, strings.Count(notes, "\n"), "the sync brings the library's notes across")
	c.True(focusOnTags, "the focus is on the rebuilt field that had it")
	c.True(v > viewHeight, "the notes push the field down by more than the view's height, and the editor scrolls "+
		"along with it: scrolled to %v, in a view %v high", v, viewHeight)
	c.True(nearlyWithin(whole, after), "so all of the field is in view: %v within %v", whole, after)
	c.True(nearlyWithin(before, after) && nearlyWithin(after, before),
		"right where it was: %v, was %v", after, before)
	screen.Click(screen.PanelCenter(e.cancelButton))
}

// TestEditorSyncClosesItsSubEditors verifies that syncing from the source menu first closes the editors open on the
// modifiers and weapons within the editor's data, which the sync replaces, and syncs nothing when one of them is kept
// open, as canceling its prompt to save does. Closing a sub-editor gives the focus to the list it was opened from and
// scrolls the editor to that list, yet the focus, its selection and the editor's scrolling all end up as they were when
// the sync was chosen.
func TestEditorSyncClosesItsSubEditors(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	// Enlarged, so that the lists at the end of the editor are out of view while a field near its start is in view.
	swapForTest(t, &gurps.GlobalSettings().General.InitialEditorUIScale, 300)
	e, _ := newEditedLibraryTrait(t, c, screen, sheet, nil, func(local *gurps.Trait) {
		local.Weapons = []*gurps.Weapon{gurps.NewWeapon(local, true)}
		local.Weapons[0].Usage = "Thrust"
		modifier := gurps.NewTraitModifier(local.DataOwner(), nil, false)
		modifier.Name = "Long"
		local.Modifiers = []*gurps.TraitModifier{modifier}
		gurps.AttachModifiers(local, local.Modifiers)
	})
	isWeaponEditor := func(d unison.Dockable) bool {
		_, isEditor := d.AsPanel().Self.(*editor[*gurps.Weapon, *gurps.Weapon])
		return isEditor
	}
	isModifierEditor := func(d unison.Dockable) bool {
		_, isEditor := d.AsPanel().Self.(*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData])
		return isEditor
	}
	// Opened as from the list of weapons, which holds the focus as the weapon's editor opens.
	screen.Do(func() {
		e.meleeWeapons.table.RequestFocus()
		EditWeapon(e, e.editorData.Weapons[0])
	})
	weaponEditor := soleEditor[*editor[*gurps.Weapon, *gurps.Weapon]](t, screen, isWeaponEditor)
	screen.Do(func() {
		weaponEditor.editorData.Usage = "Swung"
		weaponEditor.MarkModified(nil)
	})
	vttNotes := editorField(t, screen, e, "VTT Notes")
	// Brought into view to be clicked on, which leaves the list of weapons out of view.
	screen.Do(vttNotes.ScrollIntoView)
	screen.Click(screen.PanelPoint(vttNotes, geom.Point{X: 4, Y: 4}))
	screen.Type("Typed")
	// A selection made backwards, whose start is the end that moves as it is extended.
	screen.KeyPress(unison.KeyLeft, mod.Shift)
	screen.KeyPress(unison.KeyLeft, mod.Shift)
	var focused, listInView bool
	var h, v float32
	screen.Do(func() {
		focused = wnd.CurrentFocus() == vttNotes.AsPanel()
		listInView = !visibleRect(e.meleeWeapons.table.AsPanel()).Empty()
		h, v = e.scroll.Position()
	})
	c.True(focused, "precondition: the field holds the focus")
	c.False(listInView, "precondition: the list of weapons is out of view")

	// Canceling the prompt to save the weapon's changes keeps its editor open, so nothing is synced.
	chooseSyncWithSource(t, screen, wnd, e)
	_, dialog := modalDialog(t, screen, wnd)
	var button *unison.Button
	screen.Do(func() { button = dialog.Button(unison.ModalResponseCancel) })
	if button == nil {
		t.Fatal("the save prompt has no cancel button")
	}
	screen.Click(screen.PanelCenter(button))
	var name string
	var weaponEditors, selStart, selEnd int
	var hAfter, vAfter float32
	screen.Do(func() {
		name = e.editorData.Name
		weaponEditors = len(AllMatchingDockables(isWeaponEditor))
		focused = wnd.CurrentFocus() == vttNotes.AsPanel()
		selStart, selEnd = vttNotes.Selection()
		hAfter, vAfter = e.scroll.Position()
	})
	c.Equal(1, weaponEditors, "canceling keeps the weapon's editor open")
	c.Equal("Claws (old)", name, "so nothing is synced")
	c.True(focused, "and the focus is back on the field that had it")
	c.Equal(3, selStart, "with the selection as it was")
	c.Equal(5, selEnd)
	c.Equal(h, hAfter, "and the editor scrolled as it was")
	c.Equal(v, vAfter)
	screen.KeyPress(unison.KeyLeft, mod.Shift)
	screen.Do(func() {
		selStart, selEnd = vttNotes.Selection()
		h, v = e.scroll.Position()
	})
	c.Equal(2, selStart, "the selection is still extended from its start")
	c.Equal(5, selEnd)

	// Discarding the weapon's changes lets its editor close, and the sync go ahead.
	chooseSyncWithSource(t, screen, wnd, e)
	_, dialog = modalDialog(t, screen, wnd)
	screen.Do(func() { button = dialog.Button(unison.ModalResponseDiscard) })
	if button == nil {
		t.Fatal("the save prompt has no discard button")
	}
	screen.Click(screen.PanelCenter(button))
	rebuilt := editorField(t, screen, e, "VTT Notes")
	var notes string
	screen.Do(func() {
		name = e.editorData.Name
		notes = e.editorData.VTTNotes
		weaponEditors = len(AllMatchingDockables(isWeaponEditor))
		focused = rebuilt != vttNotes && wnd.CurrentFocus() == rebuilt.AsPanel()
		selStart, selEnd = rebuilt.Selection()
		listInView = !visibleRect(e.meleeWeapons.table.AsPanel()).Empty()
		hAfter, vAfter = e.scroll.Position()
	})
	c.Equal(0, weaponEditors, "the weapon's editor is closed")
	c.Equal("Claws", name, "and the trait synced")
	c.Equal("Typed", notes)
	c.True(focused, "the focus is on the rebuilt field that had it, not the list the weapon's editor gave it to")
	c.Equal(2, selStart, "with the selection as it was")
	c.Equal(5, selEnd)
	c.Equal(h, hAfter, "the editor is scrolled as it was, not to the list the weapon's editor was opened from")
	c.Equal(v, vAfter)
	c.False(listInView)
	screen.KeyPress(unison.KeyLeft, mod.Shift)
	screen.Do(func() { selStart, selEnd = rebuilt.Selection() })
	c.Equal(1, selStart, "the rebuilt field's selection is extended from its start as well")
	c.Equal(5, selEnd)

	// A focus that isn't within the content, as the source button's is once a screen reader has clicked it, is put back
	// too, rather than left on the list a closed sub-editor gave it to, and the editor is scrolled as it was.
	var sourceButton *unison.Button
	var modifiers *traitModifiersPanel
	screen.Do(func() {
		e.editorData.Name = "Claws (changed)"
		e.meleeWeapons.table.RequestFocus()
		EditWeapon(e, e.editorData.Weapons[0])
		if panels := panelsOfType[*traitModifiersPanel](e.content); len(panels) == 1 {
			modifiers = panels[0]
			modifiers.table.RequestFocus()
			EditTraitModifier(e, e.editorData.Modifiers[0])
		}
		if sourceButton = buttonWithSVG(e.AsPanel(), svg.Database); sourceButton != nil {
			sourceButton.RequestFocus()
			focused = wnd.CurrentFocus() == sourceButton.AsPanel()
		}
		e.scroll.SetPosition(0, 0)
		listInView = !visibleRect(e.meleeWeapons.table.AsPanel()).Empty()
	})
	if sourceButton == nil {
		t.Fatal("the editor has no source button")
	}
	if modifiers == nil {
		t.Fatal("the editor has no list of modifiers")
	}
	c.True(focused, "precondition: the source button holds the focus")
	c.False(listInView, "precondition: the list of weapons is out of view")
	soleEditor[*editor[*gurps.Weapon, *gurps.Weapon]](t, screen, isWeaponEditor)
	soleEditor[*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]](t, screen, isModifierEditor)
	chooseSyncWithSource(t, screen, wnd, e)
	var modifierEditors int
	screen.Do(func() {
		name = e.editorData.Name
		weaponEditors = len(AllMatchingDockables(isWeaponEditor))
		modifierEditors = len(AllMatchingDockables(isModifierEditor))
		focused = wnd.CurrentFocus() == sourceButton.AsPanel()
		hAfter, vAfter = e.scroll.Position()
	})
	c.Equal(0, weaponEditors, "an unchanged weapon's editor closes without asking")
	c.Equal(0, modifierEditors, "as does an unchanged modifier's")
	c.Equal("Claws", name, "and the trait is synced")
	c.True(focused, "the focus is back on the source button")
	c.Equal(float32(0), hAfter, "and the editor scrolled as it was")
	c.Equal(float32(0), vAfter)

	// A field outside the content, as the toolbar's scale field is, gets its caret back along with the focus.
	var scaleField *PercentageField
	screen.Do(func() {
		e.editorData.Name = "Claws (changed)"
		e.meleeWeapons.table.RequestFocus()
		EditWeapon(e, e.editorData.Weapons[0])
		if fields := panelsOfType[*PercentageField](e.Children()[0]); len(fields) == 1 {
			scaleField = fields[0]
			scaleField.RequestFocus()
			scaleField.SetSelection(1, 1)
			focused = wnd.CurrentFocus() == scaleField.AsPanel()
		}
	})
	if scaleField == nil {
		t.Fatal("the editor's toolbar has no scale field")
	}
	c.True(focused, "precondition: the scale field holds the focus")
	chooseSyncWithSource(t, screen, wnd, e)
	screen.Do(func() {
		name = e.editorData.Name
		weaponEditors = len(AllMatchingDockables(isWeaponEditor))
		focused = wnd.CurrentFocus() == scaleField.AsPanel()
		selStart, selEnd = scaleField.Selection()
	})
	c.Equal(0, weaponEditors, "the weapon's editor is closed")
	c.Equal("Claws", name, "and the trait synced")
	c.True(focused, "the focus is back on the scale field")
	c.Equal(1, selStart, "with the caret where it was, rather than all of its text selected")
	c.Equal(1, selEnd)
	screen.Click(screen.PanelCenter(e.cancelButton))
}

// TestEditorSyncTakesTheFocusFromAClosedSubEditor verifies that syncing while the focus is within a sub-editor, which
// the sync closes, leaves the focus within the editor, where Escape and Cmd-Return still reach it: on the list the
// sub-editor was opened from, as rebuilt, or on the editor's first control when closing the sub-editor gave the focus
// to something else.
func TestEditorSyncTakesTheFocusFromAClosedSubEditor(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	e, _ := newEditedLibraryTrait(t, c, screen, sheet, nil, func(local *gurps.Trait) {
		local.Weapons = []*gurps.Weapon{gurps.NewWeapon(local, true)}
		local.Weapons[0].Usage = "Thrust"
	})
	isWeaponEditor := func(d unison.Dockable) bool {
		_, isEditor := d.AsPanel().Self.(*editor[*gurps.Weapon, *gurps.Weapon])
		return isEditor
	}
	// typeIntoWeaponEditor types into the first field of the one weapon editor open, leaving the focus there.
	typeIntoWeaponEditor := func() {
		t.Helper()
		weaponEditor := soleEditor[*editor[*gurps.Weapon, *gurps.Weapon]](t, screen, isWeaponEditor)
		var field *StringField
		screen.Do(func() {
			if fields := panelsOfType[*StringField](weaponEditor.content); len(fields) != 0 {
				field = fields[0]
			}
		})
		if field == nil {
			t.Fatal("the weapon's editor has no field to type into")
		}
		screen.Click(screen.PanelCenter(field))
		screen.Type("Swung")
		var focused, modified bool
		screen.Do(func() {
			focused = wnd.CurrentFocus() == field.AsPanel()
			modified = weaponEditor.isModified()
		})
		c.True(focused, "precondition: the focus is within the weapon's editor")
		c.True(modified, "precondition: which has changes to save")
	}
	// syncDiscardingWeaponChanges chooses Sync with Source, then Discard from the weapon editor's prompt to save.
	syncDiscardingWeaponChanges := func() {
		t.Helper()
		chooseSyncWithSource(t, screen, wnd, e)
		_, dialog := modalDialog(t, screen, wnd)
		var button *unison.Button
		screen.Do(func() { button = dialog.Button(unison.ModalResponseDiscard) })
		if button == nil {
			t.Fatal("the save prompt has no discard button")
		}
		screen.Click(screen.PanelCenter(button))
	}

	// Opened as from the list of weapons, which holds the focus as the weapon's editor opens.
	var oldTable *unison.Panel
	screen.Do(func() {
		oldTable = e.meleeWeapons.table.AsPanel()
		oldTable.RequestFocus()
		EditWeapon(e, e.editorData.Weapons[0])
	})
	typeIntoWeaponEditor()
	syncDiscardingWeaponChanges()
	var name string
	var weaponEditors int
	var focusOnList, listInView bool
	screen.Do(func() {
		name = e.editorData.Name
		weaponEditors = len(AllMatchingDockables(isWeaponEditor))
		table := e.meleeWeapons.table.AsPanel()
		focusOnList = table != oldTable && wnd.CurrentFocus() == table
		listInView = !visibleRect(table).Empty()
	})
	c.Equal(0, weaponEditors, "the weapon's editor is closed")
	c.Equal("Claws", name, "and the trait synced")
	c.True(focusOnList, "the focus is on the list the weapon's editor was opened from, as rebuilt")
	c.True(listInView, "which is in view")

	// Opened while the sheet holds the focus, which is what closing the weapon's editor then gives it back to.
	screen.Do(func() {
		e.editorData.Name = "Claws (changed)"
		wnd.SetFocus(sheet)
		EditWeapon(e, e.editorData.Weapons[0])
	})
	typeIntoWeaponEditor()
	syncDiscardingWeaponChanges()
	var focusInContent bool
	screen.Do(func() {
		name = e.editorData.Name
		weaponEditors = len(AllMatchingDockables(isWeaponEditor))
		focus := wnd.CurrentFocus()
		focusInContent = focus != nil && unison.AncestorIsOrSelf(focus, e.content)
	})
	c.Equal(0, weaponEditors, "the weapon's editor is closed")
	c.Equal("Claws", name, "and the trait synced")
	c.True(focusInContent, "the focus is within the editor's content, not on the sheet")

	screen.KeyPress(unison.KeyEscape, 0)
	var open bool
	screen.Do(func() {
		open = slices.ContainsFunc(AllDockables(), func(d unison.Dockable) bool { return d.AsPanel().Self == e })
	})
	c.False(open, "Escape reaches the editor, which discards its changes and closes")
}
