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

// showListFilterDialog puts up the modal editor for a saved filter and reports whether it was accepted. The filter is
// edited in place, so a caller that has to survive a cancel hands over a clone. except is the filter whose name the
// entered name may match without counting as a duplicate -- the one being edited -- and is nil for a new filter.
func showListFilterDialog(title, key string, filter *gurps.ListFilter, fields []filterFieldInfo, except *gurps.ListFilter) bool {
	content := unison.NewPanel()
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
	nameField.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	nameRow.AddChild(nameField)
	content.AddChild(nameRow)

	scroll := unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
	scroll.SetContent(newListFilterPanel(key, filter, fields), behavior.Fill, behavior.Fill)
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

	dialog, err := unison.NewDialog(unison.DefaultDialogTheme.QuestionIcon, unison.DefaultDialogTheme.QuestionIconInk,
		content, []*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfo()})
	if err != nil {
		reportUIError(i18n.Text("Unable to create the filter editor"), err)
		return false
	}
	dialog.Window().SetTitle(title)
	nameField.ValidateCallback = func() bool {
		name := strings.TrimSpace(filter.Name)
		valid := name != "" && !gurps.GlobalSettings().ListFilterNameInUse(key, name, except)
		dialog.Button(unison.ModalResponseOK).SetEnabled(valid)
		return valid
	}
	nameField.Validate() // Here to update the OK button.
	nameField.RequestFocus()
	if dialog.RunModal() != unison.ModalResponseOK {
		return false
	}
	filter.Name = strings.TrimSpace(filter.Name)
	return true
}
