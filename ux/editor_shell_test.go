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

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// shellTestEditor is the least an editor built on editorShell can be: it supplies the title, whether there are changes,
// and how they are applied, and counts what the shell asks of it.
type shellTestEditor struct {
	editorShell
	modified bool
	applied  int
	closed   int
}

func newShellTestEditor() *shellTestEditor {
	e := &shellTestEditor{}
	e.Self = e
	return e
}

func (e *shellTestEditor) Title() string { return "Shell Test" }

func (e *shellTestEditor) Modified() bool { return e.enableApplyAndCancel(e.modified) }

func (e *shellTestEditor) isModified() bool { return e.modified }

func (e *shellTestEditor) apply() { e.applied++ }

func (e *shellTestEditor) AttemptClose() bool {
	if !e.confirmClose(e.isModified, e.apply) {
		return false
	}
	e.closed++
	return true
}

// TestEditorShellContentKeysDriveTheButtons verifies that Cmd-Return and Escape within an editor's content stand in for
// its Apply and Discard buttons -- doing nothing while those are disabled, since there is nothing to apply or discard --
// that neither button prompts on the way out, and that other keys are left to the content.
func TestEditorShellContentKeysDriveTheButtons(t *testing.T) {
	c := check.New(t)
	e := newShellTestEditor()
	content := e.newContentPanel(2)
	c.NotNil(e.scroll, "the content panel comes with the scroll panel that holds it")
	c.Equal(content, e.scroll.Content().AsPanel(), "and the scroll panel holds the content")
	toolbar := newToolbar()
	e.addApplyAndCancelButtons(toolbar, e.apply)
	e.applyButton.ClickAnimationTime = 0
	e.cancelButton.ClickAnimationTime = 0
	c.Equal([]*unison.Panel{e.applyButton.AsPanel(), e.cancelButton.AsPanel()}, toolbar.Children())
	e.promptForSave = true

	c.False(content.KeyDownCallback(unison.KeyA, 0, false), "a key the shell does not handle is left to the content")
	c.True(content.KeyDownCallback(unison.KeyReturn, mod.OSMenuCommand(), false), "Cmd-Return is always consumed")
	c.True(content.KeyDownCallback(unison.KeyEscape, 0, false), "as is Escape")
	c.Equal(0, e.applied, "but with nothing to apply, neither does anything")
	c.Equal(0, e.closed)
	c.True(e.promptForSave)
	c.False(content.KeyDownCallback(unison.KeyEscape, mod.Shift, false), "Escape with a modifier is not the shortcut")

	e.modified = true
	c.True(e.Modified(), "reporting changes enables the buttons")
	c.True(e.applyButton.Enabled())
	c.True(e.cancelButton.Enabled())

	c.True(content.KeyDownCallback(unison.KeyNumPadEnter, mod.OSMenuCommand(), false))
	c.Equal(1, e.applied, "Cmd-Enter applies the changes")
	c.Equal(1, e.closed, "and closes the editor")
	c.False(e.promptForSave, "without prompting, since the user has just said what to do with the changes")

	e.promptForSave = true
	c.True(content.KeyDownCallback(unison.KeyEscape, 0, false))
	c.Equal(1, e.applied, "Escape applies nothing")
	c.Equal(2, e.closed, "but closes the editor")
	c.False(e.promptForSave, "again without prompting")

	e.modified = false
	c.False(e.Modified(), "reporting no changes disables the buttons again")
	c.False(e.applyButton.Enabled())
	c.False(e.cancelButton.Enabled())
}

// TestEditorShellConfirmCloseAsksOnlyWhenThereIsSomethingToAsk verifies the two cases in which closing an editor must
// not put up the save prompt: when a button has already settled what happens to the changes, and when there are none.
// Whether there are any is not even asked in the first case, since finding out can be costly.
func TestEditorShellConfirmCloseAsksOnlyWhenThereIsSomethingToAsk(t *testing.T) {
	c := check.New(t)
	e := newShellTestEditor()
	asked, applied := 0, 0
	isModified := func() bool {
		asked++
		return true
	}
	apply := func() { applied++ }

	e.promptForSave = false
	c.True(e.confirmClose(isModified, apply), "with the prompt settled, closing proceeds")
	c.Equal(0, asked, "without asking whether there are changes")
	c.Equal(0, applied, "or applying any")

	e.promptForSave = true
	c.True(e.confirmClose(func() bool { return false }, apply), "with no changes, closing proceeds")
	c.Equal(0, applied, "without applying anything")
}

// TestEditorShellReturnsToPreviousWithoutADockable verifies that returning to where an editor was opened from is a
// no-op when nothing was recorded, as when the editor was opened while no dockable was current.
func TestEditorShellReturnsToPreviousWithoutADockable(t *testing.T) {
	c := check.New(t)
	e := newShellTestEditor()
	c.Nil(e.returnToPrevious(), "nothing recorded, nothing to return to")
	e.previousDockable = newShellTestEditor()
	c.Nil(e.returnToPrevious(), "a dockable that is not in a dock container cannot be made current")
}

// isPointsEditor reports whether a dockable is a points editor.
func isPointsEditor(d unison.Dockable) bool {
	_, ok := d.AsPanel().Self.(*pointsEditor)
	return ok
}

// openPointsEditorForSheet focuses the sheet's name field, so that there is a focus to restore, and opens the points
// editor for the sheet's entity from it, the way the edit button beside the point total does. It returns the editor.
func openPointsEditorForSheet(t *testing.T, screen *unison.HeadlessScreen, sheet *Sheet) *pointsEditor {
	t.Helper()
	screen.Do(func() {
		if field := sheet.AsPanel().FindRefKey(identityPanelNameFieldRefKey); field != nil {
			field.RequestFocus()
		}
		displayPointsEditor(sheet, sheet.Entity())
	})
	return soleEditor[*pointsEditor](t, screen, isPointsEditor)
}

// checkReturnedToSheet verifies that, with the points editor gone, the sheet is current in its dock container again and
// its name field holds the keyboard focus once more, and that no prompt was left behind.
func checkReturnedToSheet(t *testing.T, c check.Checker, screen *unison.HeadlessScreen, wnd *unison.Window, sheet *Sheet) {
	t.Helper()
	var editors, windows int
	var sheetCurrent bool
	var focusKey string
	screen.Do(func() {
		editors = len(AllMatchingDockables(isPointsEditor))
		windows = len(unison.Windows())
		if dc := unison.Ancestor[*unison.DockContainer](sheet); dc != nil {
			if cur := dc.CurrentDockable(); cur != nil {
				sheetCurrent = cur.AsPanel().Self == sheet
			}
		}
		if focus := wnd.Focus(); focus != nil {
			focusKey = focus.RefKey
		}
	})
	c.Equal(0, editors, "the points editor has closed")
	c.Equal(1, windows, "without leaving a prompt behind")
	c.True(sheetCurrent, "the sheet is current in its dock container again")
	c.Equal(identityPanelNameFieldRefKey, focusKey, "and its name field has the focus back")
}

// TestPointsEditorOpensBesideTheSheetAndReturnsToIt verifies that the points editor is docked with the editors, next to
// the sheet it was opened from rather than in the sheet's own tab group, records what it edits and where it was opened
// from, and then, whichever way it is closed -- discarding with Escape, applying with Cmd-Return, or closing the tab and
// discarding at the prompt -- makes the sheet current again and hands the focus back to the field that had it.
func TestPointsEditorOpensBesideTheSheetAndReturnsToIt(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	entity := sheet.Entity()
	records := len(entity.PointsRecord)

	e := openPointsEditorForSheet(t, screen, sheet)
	var previous unison.Dockable
	var focusKey string
	var associatedID any
	var group any
	var editorCurrent, sameContainer, promptForSave, applyEnabled, cancelEnabled, focusInEditor bool
	screen.Do(func() {
		previous = e.previousDockable
		focusKey = e.previousFocusKey
		associatedID = e.ClientData()[AssociatedIDKey]
		group = e.ClientData()[dockGroupClientDataKey]
		promptForSave = e.promptForSave
		applyEnabled = e.applyButton.Enabled()
		cancelEnabled = e.cancelButton.Enabled()
		dc := unison.Ancestor[*unison.DockContainer](e)
		if dc != nil {
			if cur := dc.CurrentDockable(); cur != nil {
				editorCurrent = cur.AsPanel().Self == e
			}
			sameContainer = dc == unison.Ancestor[*unison.DockContainer](sheet)
		}
		if focus := wnd.Focus(); focus != nil {
			focusInEditor = unison.AncestorIsOrSelf(focus, e)
		}
	})
	c.True(previous != nil && previous.AsPanel().Self == sheet, "the editor records the sheet as where it was opened from")
	c.Equal(identityPanelNameFieldRefKey, focusKey, "along with the field that had the focus")
	c.Equal(any(entity.ID), associatedID, "and what it edits, so that closing the sheet closes it too")
	c.Equal(any(dgroup.Editors), group, "it is grouped with the editors")
	c.False(sameContainer, "which are docked beside the sheet, not in its tab group")
	c.True(editorCurrent, "and is brought to the front")
	c.True(focusInEditor, "with the focus within it")
	c.True(promptForSave, "closing it prompts to save until a button settles the changes")
	c.False(applyEnabled, "there is nothing to apply yet")
	c.False(cancelEnabled, "or to discard")

	// Add an entry, which is a change to discard, then press Escape, which stands in for the Discard button.
	addEntry := buttonWithTooltip(e.AsPanel(), i18n.Text("Add Entry"))
	if addEntry == nil {
		t.Fatal("the points editor has no Add Entry button")
	}
	screen.Click(screen.PanelCenter(addEntry))
	var modified bool
	screen.Do(func() {
		modified = e.Modified()
		applyEnabled = e.applyButton.Enabled()
		cancelEnabled = e.cancelButton.Enabled()
	})
	c.True(modified, "adding an entry is a change")
	c.True(applyEnabled, "that can be applied")
	c.True(cancelEnabled, "or discarded")
	screen.KeyPress(unison.KeyEscape, 0)
	checkReturnedToSheet(t, c, screen, wnd, sheet)
	c.Equal(records, len(entity.PointsRecord), "discarding leaves the entity's points record alone")

	// Add an entry again, then press Cmd-Return, which stands in for the Apply button.
	e = openPointsEditorForSheet(t, screen, sheet)
	addEntry = buttonWithTooltip(e.AsPanel(), i18n.Text("Add Entry"))
	if addEntry == nil {
		t.Fatal("the reopened points editor has no Add Entry button")
	}
	screen.Click(screen.PanelCenter(addEntry))
	screen.KeyPress(unison.KeyReturn, mod.OSMenuCommand())
	checkReturnedToSheet(t, c, screen, wnd, sheet)
	c.Equal(records+1, len(entity.PointsRecord), "applying adds the entry to the entity's points record")

	// Add an entry once more, then close the tab, which prompts, and discard at the prompt. The close runs a modal
	// prompt, so it is posted rather than run through Do, which would wait for it to return.
	e = openPointsEditorForSheet(t, screen, sheet)
	addEntry = buttonWithTooltip(e.AsPanel(), i18n.Text("Add Entry"))
	if addEntry == nil {
		t.Fatal("the reopened points editor has no Add Entry button")
	}
	screen.Click(screen.PanelCenter(addEntry))
	screen.Post(func() { e.AttemptClose() })
	screen.Sync()
	_, dialog := modalDialog(t, screen, wnd)
	var discard *unison.Button
	screen.Do(func() { discard = dialog.Button(unison.ModalResponseDiscard) })
	if discard == nil {
		t.Fatal("the save prompt has no discard button")
	}
	screen.Click(screen.PanelCenter(discard))
	checkReturnedToSheet(t, c, screen, wnd, sheet)
	c.Equal(records+1, len(entity.PointsRecord), "discarding at the prompt leaves the points record as it was")

	// Opening the editor a second time for the same sheet brings the existing one to the front instead.
	e = openPointsEditorForSheet(t, screen, sheet)
	screen.Do(func() { displayPointsEditor(sheet, entity) })
	c.Equal(e, soleEditor[*pointsEditor](t, screen, isPointsEditor),
		"asking for the editor again activates the one that is open instead of opening another")
	closeEditorWithoutPrompt(t, screen, e)
}
