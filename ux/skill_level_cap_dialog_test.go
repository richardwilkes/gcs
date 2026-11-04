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
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/levelcap"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// skillLevelCapDialogResult reports, once the cap editor is gone, whether showSkillLevelCapDialog has returned and
// whether it returned true.
type skillLevelCapDialogResult func() (returned, accepted bool)

// openSkillLevelCapDialog puts up the cap editor for the cap and returns its window, the dialog, the editor in it and
// what reports the result. The dialog runs a modal loop of its own, so it is put up from a task rather than through Do.
func openSkillLevelCapDialog(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window, lc *gurps.SkillLevelCap) (dialogWnd *unison.Window, dialog *unison.Dialog, p *skillLevelCapPanel, result skillLevelCapDialogResult) {
	t.Helper()
	var returned, accepted bool
	if !screen.Post(func() {
		accepted = showSkillLevelCapDialog("Skill Level Cap", nil, lc)
		returned = true
	}) {
		t.Fatal("the headless session has ended")
	}
	screen.Sync()
	dialogWnd, dialog = modalDialog(t, screen, wnd)
	screen.Do(func() { p, _ = uxtest.FirstPanelOfType[*skillLevelCapPanel](dialogWnd.Content()) })
	if p == nil {
		t.Fatal("the dialog holds no cap editor")
	}
	result = func() (wasReturned, wasAccepted bool) {
		screen.Do(func() { wasReturned, wasAccepted = returned, accepted })
		return wasReturned, wasAccepted
	}
	return dialogWnd, dialog, p, result
}

// TestSkillLevelCapDialogAcceptHeadless puts up the cap editor for a new cap, checking that it opens at least its
// minimum size with the focus on what the cap applies to, since a new cap compares no name; that OK stays off, with a
// notice that says why, until the cap has a value, and goes off again when undo takes the value away; and that OK
// returns true, leaving the edits, typing included, in the cap handed over.
func TestSkillLevelCapDialogAcceptHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	lc := gurps.NewSkillLevelCap()
	dialogWnd, dialog, p, result := openSkillLevelCapDialog(t, screen, wnd, lc)
	okButton := dialogButton(t, screen, dialog, unison.ModalResponseOK)
	type state struct {
		ok, notice bool
	}
	read := func() state {
		var s state
		screen.Do(func() {
			s.ok = okButton.Enabled()
			s.notice = slices.Contains(uxtest.LabelTexts(dialogWnd.Content()), "Add at least one cap value before saving.")
		})
		return s
	}
	var size geom.Size
	var focus string
	screen.Do(func() {
		size = dialogWnd.ContentRect().Size
		focus = focusedRefKey(dialogWnd)
	})
	c.True(size.Width >= skillLevelCapDialogMinWidth, "the editor opens at least %d wide, but is %v",
		skillLevelCapDialogMinWidth, size.Width)
	c.True(size.Height >= skillLevelCapDialogMinHeight, "the editor opens at least %d tall, but is %v",
		skillLevelCapDialogMinHeight, size.Height)
	c.Equal(levelCapMatcherKey+":target", focus, "with no name to compare, what the cap applies to takes the focus")
	c.Equal(state{notice: true}, read(), "a cap with no value can't be accepted, and the notice says so")

	screen.Do(func() { menuAction(p.treeAddEntries(p.levelCap.Cap, treeRootPath), "Fixed Level")() })
	c.Equal(state{ok: true}, read(), "a value lets the cap be accepted")
	var undoMgr *unison.UndoManager
	screen.Do(func() { undoMgr = unison.UndoManagerFor(p) })
	if undoMgr == nil {
		t.Fatal("the dialog has no undo manager")
	}
	screen.Do(undoMgr.Undo)
	c.Equal(state{notice: true}, read(), "undo taking the value away turns OK off again")
	screen.Do(undoMgr.Redo)
	c.Equal(state{ok: true}, read(), "and redo turns it back on")
	screen.Do(func() { focus = focusedRefKey(dialogWnd) })
	c.Equal("r.0:level", focus, "with the value open again and its level focused")
	screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
	screen.Type("14")
	screen.Click(screen.PanelCenter(okButton))
	returned, accepted := result()
	c.True(returned, "OK dismisses the editor")
	c.True(accepted, "accepting it")
	c.Equal(1, len(lc.Cap.Children), "the cap handed over holds the value")
	if len(lc.Cap.Children) == 1 {
		fixed, isFixed := lc.Cap.Children[0].(*gurps.FixedLevelCap)
		c.True(isFixed, "a fixed level")
		if isFixed {
			c.Equal(fxp.FromInteger(14), fixed.Level, "at the level typed")
		}
	}
}

// TestSkillLevelCapDialogCancelHeadless puts up the cap editor on a copy of a cap, as its callers do, checking that the
// focus starts on the name the cap compares, and that Cancel returns false, which has the caller leave the cap it made
// the copy from as it was, though the copy was changed.
func TestSkillLevelCapDialogCancelHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	original := gurps.NewSkillLevelCap()
	original.NameCriteria = criteria.Text{Compare: criteria.ContainsText, Qualifier: "Magic"}
	original.Cap.Children = gurps.LevelCaps{gurps.NewFixedLevelCap(original.Cap, fxp.Twelve)}
	before := gurps.Hash64(original)
	edited := original.Clone()
	dialogWnd, dialog, p, result := openSkillLevelCapDialog(t, screen, wnd, edited)
	var focus string
	screen.Do(func() { focus = focusedRefKey(dialogWnd) })
	c.Equal(levelCapMatcherKey+":name", focus, "the name the cap compares takes the focus")
	screen.Do(func() {
		if popup, ok := refAs[*unison.PopupMenu[levelcap.Target]](t, p.AsPanel(), levelCapMatcherKey+":target"); ok {
			popup.Select(levelcap.Spells)
		}
	})
	screen.Do(func() { menuAction(p.treeAddEntries(p.levelCap.Cap, treeRootPath), "Skill Level")() })
	screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseCancel)))
	returned, accepted := result()
	c.True(returned, "Cancel dismisses the editor")
	c.False(accepted, "without accepting it")
	c.Equal(levelcap.Spells, edited.Target, "the copy took the edits")
	c.Equal(2, len(edited.Cap.Children))
	c.Equal(before, gurps.Hash64(original), "the cap the copy was made from is as it was")
}

// TestSkillLevelCapDialogKeysHeadless checks that, where the dialog has no menu bar to take them, the key bindings of
// Undo and Redo work in the cap editor with the focus on a button outside the rows, and that Escape closes the open row
// before it cancels the dialog.
func TestSkillLevelCapDialogKeysHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &dialogMenuTakesUndoKeys, false)
	lc := gurps.NewSkillLevelCap()
	lc.Cap.Children = gurps.LevelCaps{gurps.NewFixedLevelCap(lc.Cap, fxp.Twelve)}
	_, dialog, p, result := openSkillLevelCapDialog(t, screen, wnd, lc)
	okButton := dialogButton(t, screen, dialog, unison.ModalResponseOK)
	kind := func() levelcap.Type {
		var k levelcap.Type
		screen.Do(func() { k = p.node("r.0").CapType() })
		return k
	}
	open := func() string {
		var path string
		screen.Do(func() { path = p.open })
		return path
	}
	windows := func() int {
		var n int
		screen.Do(func() { n = len(unison.Windows()) })
		return n
	}
	focusOK := func() { screen.Do(func() { okButton.RequestFocus() }) }

	screen.Do(func() { p.toggle("r.0") })
	screen.Do(func() {
		if popup, ok := refAs[*unison.PopupMenu[levelcap.Type]](t, p.AsPanel(), "r.0:type"); ok {
			popup.Select(levelcap.Skill)
		}
	})
	c.Equal(levelcap.Skill, kind(), "precondition: the value has been switched to a skill's level")
	focusOK()
	screen.KeyPress(unison.KeyZ, mod.OSMenuCommand())
	c.Equal(levelcap.Fixed, kind(), "the Undo key undoes with OK focused")
	focusOK()
	screen.KeyPress(unison.KeyY, mod.OSMenuCommand())
	c.Equal(levelcap.Skill, kind(), "as the Redo key redoes")

	c.Equal("r.0", open(), "precondition: the row is open")
	focusOK()
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal("", open(), "Escape closes the open row")
	c.Equal(2, windows(), "leaving the dialog up")
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal(1, windows(), "with no row open, Escape cancels the dialog")
	returned, accepted := result()
	c.True(returned, "which has returned")
	c.False(accepted, "without accepting it")
}

// TestSkillLevelCapDialogDragHeadless drags a value by its sentence within the cap editor's dialog, which is a window
// of its own, so the drop lands only because the dialog takes the rows' drag type.
func TestSkillLevelCapDialogDragHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	lc := gurps.NewSkillLevelCap()
	first := gurps.NewFixedLevelCap(lc.Cap, fxp.Twelve)
	second := gurps.NewFixedLevelCap(lc.Cap, fxp.FromInteger(14))
	lc.Cap.Children = gurps.LevelCaps{first, second}
	_, _, p, result := openSkillLevelCapDialog(t, screen, wnd, lc)
	var sentence, target *unison.Panel
	var above geom.Point
	screen.Do(func() {
		sentence, _ = refAs[*unison.Panel](t, p.AsPanel(), "r.1"+keySentence)
		if more, ok := refAs[*unison.Panel](t, p.AsPanel(), "r.0"+keyMore); ok {
			target = more.Parent()
			above = geom.NewPoint(40, target.FrameRect().Height*0.2)
		}
	})
	if sentence == nil || target == nil {
		return
	}
	screen.Drag(screen.PanelCenter(sentence), screen.PanelPoint(target, above), 10)
	screen.Do(func() {
		c.True(p.node("r.0") == second, "the dragged value goes first")
		c.True(p.node("r.1") == first, "ahead of the other")
	})
	screen.KeyPress(unison.KeyEscape, mod.None)
	returned, _ := result()
	c.True(returned, "Escape dismisses the editor")
}
