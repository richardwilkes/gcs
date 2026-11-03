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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
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
	content := newSentenceDialogContent()
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
	setMessage := addDialogNotice(dialog)
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
	installSentenceDialogKeys(dialog.Window(), content, panel)
	nameField.RequestFocus()
	if dialog.RunModal() != unison.ModalResponseOK {
		return false
	}
	filter.Name = strings.TrimSpace(filter.Name)
	return true
}
