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
	"runtime"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/mod"
)

// addDialogNotice puts an empty label for error messages at the start of the dialog's row of buttons and returns the
// function that sets its message, which is set out as a notice, or hidden when empty.
func addDialogNotice(dialog *unison.Dialog) func(message string) {
	buttons := dialog.Button(unison.ModalResponseOK).Parent()
	parent := buttons.Parent()
	index := parent.IndexOfChild(buttons)
	buttons.RemoveFromParent()
	row := unison.NewPanel()
	row.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
	})
	row.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	label, setNotice := newNoticeLabel()
	label.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Start,
		VAlign: align.Middle,
		HGrab:  true,
	})
	row.AddChild(label)
	row.AddChild(buttons)
	parent.AddChildAtIndex(row, index)
	return func(message string) {
		if message == "" {
			setNotice("", nil, unison.ThemeOnSurface)
		} else {
			setNotice(message, unison.ThemeError, unison.ThemeOnError)
		}
		row.MarkForLayoutAndRedraw()
	}
}

// sentenceDialogContent holds the content of a dialog that edits a panel of sentence rows, and provides the undo
// manager its edits are recorded with.
type sentenceDialogContent struct {
	unison.Panel
	undoMgr *unison.UndoManager
}

func newSentenceDialogContent() *sentenceDialogContent {
	c := &sentenceDialogContent{undoMgr: unison.NewUndoManager(100, func(err error) { errs.Log(err) })}
	c.Self = c
	return c
}

// UndoManager implements unison.UndoManagerProvider.
func (c *sentenceDialogContent) UndoManager() *unison.UndoManager {
	return c.undoMgr
}

// dialogMenuTakesUndoKeys is true where the menu bar takes the key bindings of Undo and Redo while a dialog is up, as
// macOS's does, matching them by the character a key types. Elsewhere a dialog has no menu bar, so it takes them
// itself.
var dialogMenuTakesUndoKeys = runtime.GOOS == "darwin"

// rowCloser is a panel of rows, one of which may be open, such as a sentenceRows.
type rowCloser interface {
	hasOpenRow() bool
	closeRow()
}

// installSentenceDialogKeys has Escape close the open row of the panel of rows before it cancels the dialog, wherever
// the focus is, even on the buttons or another field outside that panel, unless the focused field's suggestions are
// showing, when the field hides them with it instead. Where the menu bar doesn't take them, it also has the dialog's
// window handle the key bindings of Undo and Redo.
func installSentenceDialogKeys(wnd *unison.Window, content *sentenceDialogContent, panel rowCloser) {
	wnd.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, repeat bool) bool {
		// A held Escape or Return acts once, rather than closing a row and then canceling or accepting the dialog.
		if repeat && (keyCode == unison.KeyEscape || keyCode == unison.KeyReturn || keyCode == unison.KeyNumPadEnter) {
			return true
		}
		// A cleared key binding matches no key.
		bound := func(action *unison.Action) bool {
			return !dialogMenuTakesUndoKeys && action != nil && action.KeyBinding.KeyCode != 0 &&
				action.KeyBinding.KeyCode == keyCode && action.KeyBinding.Modifiers == mods&mod.NonSticky
		}
		switch {
		case bound(undoAction):
			if content.undoMgr.CanUndo() {
				content.undoMgr.Undo()
			}
		case bound(redoAction):
			if content.undoMgr.CanRedo() {
				content.undoMgr.Redo()
			}
		case keyCode == unison.KeyEscape && noModifiersDown(mods) && panel.hasOpenRow() && !focusedSuggestionsShowing(wnd):
			panel.closeRow()
		default:
			return false
		}
		return true
	}
}

// focusedSuggestionsShowing returns true while the field with the focus in wnd is showing suggestions.
func focusedSuggestionsShowing(wnd *unison.Window) bool {
	if focus := wnd.CurrentFocus(); focus != nil {
		if field, ok := focus.Self.(*StringField); ok {
			return field.suggestionsShowing()
		}
	}
	return false
}
