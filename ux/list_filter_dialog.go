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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/mod"
)

const (
	// listFilterDialogMinWidth and listFilterDialogMinHeight are the smallest the filter editor's scrolling area is
	// allowed to be. A filter's rows are wide and can nest deeply, so the dialog opens with room for a few levels
	// rather than packing itself around whatever the filter happens to hold at the moment.
	listFilterDialogMinWidth  = 700
	listFilterDialogMinHeight = 400
)

// showListFilterDialog puts up the modal editor for a saved filter and reports whether it was accepted. It can't be
// accepted until the filter has a name and at least one condition. The filter is edited in place, so a caller that has
// to survive a cancel hands over a clone. except is the filter whose name the entered name may match without counting
// as a duplicate -- the one being edited -- and is nil for a new filter.
func showListFilterDialog(title, key string, filter *gurps.ListFilter, fields []filterFieldInfo, except *gurps.ListFilter) bool {
	content := newListFilterDialogContent()
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})

	nameRow := unison.NewPanel()
	nameRow.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	nameRow.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	nameRow.AddChild(NewFieldLeadingLabel(i18n.Text("Name"), false))
	nameField := NewStringField(nil, "", i18n.Text("Name"), func() string { return filter.Name },
		func(s string) { filter.Name = s })
	nameField.Watermark = i18n.Text(`Give this filter a name, like "Allowed in My Campaign"`)
	nameField.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	nameRow.AddChild(nameField)
	content.AddChild(nameRow)

	scroll := unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
	panel := newListFilterPanel(key, filter, fields)
	scroll.SetContent(panel, behavior.Fill, behavior.Fill)
	scroll.BackgroundInk = unison.ThemeSurface
	// The dialog replaces the layout data of the panel it is handed, so the minimum size has to be asked for here,
	// on a child of that panel, rather than on the panel itself.
	scroll.SetLayoutData(&unison.FlexLayoutData{
		MinSize: geom.NewSize(listFilterDialogMinWidth, listFilterDialogMinHeight),
		HAlign:  align.Fill,
		VAlign:  align.Fill,
		HGrab:   true,
		VGrab:   true,
	})
	content.AddChild(scroll)

	dialog, err := unison.NewDialog(nil, nil, content,
		[]*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfo()})
	if err != nil {
		reportUIError(i18n.Text("Unable to create the filter editor"), err)
		return false
	}
	dialog.Window().SetTitle(title)
	// Rows are dragged within the dialog, which is a window of its own.
	dialog.Window().RegisterForDragTypes(listFilterDragKey)
	setMessage := addListFilterDialogMessage(dialog)
	// OK needs a name no other saved filter bears and something to test. A blank name gets no message, since the
	// watermark already says what is wanted.
	nameField.ValidateCallback = func() bool {
		name := strings.TrimSpace(filter.Name)
		inUse := name != "" && gurps.GlobalSettings().ListFilterNameInUse(key, name, except)
		empty := filter.Root.Empty()
		valid := name != "" && !inUse
		dialog.Button(unison.ModalResponseOK).SetEnabled(valid && !empty)
		var message string
		switch {
		case inUse:
			message = i18n.Text("Another saved filter already has this name.")
		case empty:
			message = i18n.Text("Add at least one filter before saving.")
		}
		setMessage(message)
		return valid
	}
	panel.rebuilt = nameField.Validate
	nameField.Validate() // Here to update the OK button and the message.
	// Undo and Redo find the dialog's undo manager wherever the focus is, even on a button outside the content.
	dialog.Window().ClientData()[windowUndoManagerKey] = content.undoMgr
	installListFilterDialogKeys(dialog.Window(), content, panel)
	nameField.RequestFocus()
	if dialog.RunModal() != unison.ModalResponseOK {
		return false
	}
	filter.Name = strings.TrimSpace(filter.Name)
	return true
}

// addListFilterDialogMessage puts an empty label for error messages at the start of the dialog's row of buttons and
// returns the function that sets its message, which is set out as a notice, or hidden when empty.
func addListFilterDialogMessage(dialog *unison.Dialog) func(message string) {
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

// listFilterDialogContent holds the filter editor's dialog content, and provides the undo manager its edits are
// recorded with.
type listFilterDialogContent struct {
	unison.Panel
	undoMgr *unison.UndoManager
}

func newListFilterDialogContent() *listFilterDialogContent {
	c := &listFilterDialogContent{undoMgr: unison.NewUndoManager(100, func(err error) { errs.Log(err) })}
	c.Self = c
	return c
}

// UndoManager implements unison.UndoManagerProvider.
func (c *listFilterDialogContent) UndoManager() *unison.UndoManager {
	return c.undoMgr
}

// dialogMenuTakesUndoKeys is true where the menu bar takes the key bindings of Undo and Redo while a dialog is up, as
// macOS's does, matching them by the character a key types. Elsewhere a dialog has no menu bar, so it takes them
// itself.
var dialogMenuTakesUndoKeys = runtime.GOOS == "darwin"

// installListFilterDialogKeys has Escape close the open condition before it cancels the dialog, wherever the focus is,
// since the name field and the buttons sit outside the panel of conditions, unless the focused field's suggestions are
// showing, when the field hides them with it instead. Where the menu bar doesn't take them, it also has the dialog's
// window handle the key bindings of Undo and Redo.
func installListFilterDialogKeys(wnd *unison.Window, content *listFilterDialogContent, panel *listFilterPanel) {
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
		case keyCode == unison.KeyEscape && noModifiersDown(mods) && panel.open != "" && !focusedSuggestionsShowing(wnd):
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
