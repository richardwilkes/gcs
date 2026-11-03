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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
)

const (
	// skillLevelCapDialogMinWidth and skillLevelCapDialogMinHeight are the smallest the cap editor's scrolling area is
	// allowed to be, so that it opens with room for a few levels of groups rather than fitted to what the cap holds.
	skillLevelCapDialogMinWidth  = 700
	skillLevelCapDialogMinHeight = 360
)

// showSkillLevelCapEditor puts up the editor for one skill level cap and reports whether it was accepted; see
// showSkillLevelCapDialog. It is a variable so that tests can stand in for the modal dialog.
var showSkillLevelCapEditor = showSkillLevelCapDialog

// showSkillLevelCapDialog puts up the modal editor for a skill level cap, titled title, and reports whether it was
// accepted. It can't be accepted until the cap has at least one value. The cap is edited in place, so a caller that has
// to survive a cancel hands over a clone. entity, which may be nil, defines the attributes the cap's values can rest
// on.
func showSkillLevelCapDialog(title string, entity *gurps.Entity, levelCap *gurps.SkillLevelCap) bool {
	content := newSentenceDialogContent()
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	scroll := unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
	panel := newSkillLevelCapPanel(entity, levelCap)
	scroll.SetContent(panel, behavior.Fill, behavior.Fill)
	scroll.BackgroundInk = unison.ThemeSurface
	// The dialog replaces the layout data of the panel it is handed, so the minimum size has to be asked for here, on a
	// child of that panel, rather than on the panel itself.
	scroll.SetLayoutData(&unison.FlexLayoutData{
		MinSize: geom.NewSize(skillLevelCapDialogMinWidth, skillLevelCapDialogMinHeight),
		HAlign:  align.Fill,
		VAlign:  align.Fill,
		HGrab:   true,
		VGrab:   true,
	})
	content.AddChild(scroll)

	dialog, err := unison.NewDialog(nil, nil, content,
		[]*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfo()})
	if err != nil {
		reportUIError(i18n.Text("Unable to create the skill level cap editor"), err)
		return false
	}
	dialog.Window().SetTitle(title)
	// Rows are dragged within the dialog, which is a window of its own.
	dialog.Window().RegisterForDragTypes(levelCapDragKey)
	setMessage := addDialogNotice(dialog)
	// OK needs a value to cap at; a cap with none would cap nothing.
	validate := func() {
		empty := levelCap.Cap.Empty()
		dialog.Button(unison.ModalResponseOK).SetEnabled(!empty)
		var message string
		if empty {
			message = i18n.Text("Add at least one cap value before saving.")
		}
		setMessage(message)
	}
	panel.rebuilt = validate
	validate()
	// Undo and Redo find the dialog's undo manager wherever the focus is, even on a button outside the content.
	dialog.Window().ClientData()[windowUndoManagerKey] = content.undoMgr
	installSentenceDialogKeys(dialog.Window(), content, panel)
	// The name has a field only while it is compared with something; otherwise the popup that starts the sentence takes
	// the focus.
	focus := panel.FindRefKey(levelCapMatcherKey + ":name")
	if focus == nil {
		focus = panel.FindRefKey(levelCapMatcherKey + ":target")
	}
	if focus != nil {
		focus.RequestFocus()
	}
	return dialog.RunModal() == unison.ModalResponseOK
}
