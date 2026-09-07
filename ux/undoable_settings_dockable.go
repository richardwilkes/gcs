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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

// undoableSettingsModel is what an undoableSettingsDockable needs of the model it edits.
type undoableSettingsModel interface {
	gurps.Hashable
	Save(filePath string) error
	ResetTargetKeyPrefixes(prefixProvider func() string)
}

// undoableSettingsSpec describes one kind of per-sheet settings editor: where its help lives, how to copy its model,
// how to build its content, how to apply the edited model to what it came from, and any extra toolbar buttons. It is
// all that distinguishes the attribute editor from the body type editor.
type undoableSettingsSpec[T undoableSettingsModel] struct {
	// helpLink is the link the toolbar's Help button opens.
	helpLink string
	// clone returns a deep copy of a model.
	clone func(T) T
	// buildContent fills the editor's content panel from its model. It is called for the initial build and again by
	// every sync().
	buildContent func()
	// apply copies the edited model back to the settings it came from.
	apply func()
	// extraToolbar, which may be nil, adds buttons to the toolbar after the apply and cancel buttons.
	extraToolbar func(toolbar *unison.Panel)
}

// undoableSettingsDockable is the common part of the editors for settings that belong to a sheet or to the defaults --
// the attribute and body type editors. Each edits a copy of the settings, applying it back only when asked to, or when
// closing with changes pending and the user agrees; the cancel button discards the copy. Every change to the model
// goes through an undo edit that captures the whole model, so that a structural change is as undoable as a typed one.
type undoableSettingsDockable[T undoableSettingsModel] struct {
	SettingsDockable
	structuralEditorBase[T]
	spec         undoableSettingsSpec[T]
	applyButton  *unison.Button
	cancelButton *unison.Button
	// promptForSave is cleared by the apply and cancel buttons, which have already settled what happens to pending
	// changes, so that closing then asks nothing.
	promptForSave bool
}

// init readies the dockable to edit the model, which is given its own target key prefixes and recorded as the state
// the dockable is compared against. The outer dockable calls it after setting Self, then sets what is particular to it
// -- the tab title and icon, the file extensions, the loader and the resetter -- before calling show.
func (d *undoableSettingsDockable[T]) init(editor rowDragEditor, spec undoableSettingsSpec[T], model T) {
	d.spec = spec
	d.promptForSave = true
	d.initEditor(editor, spec.clone, spec.buildContent)
	model.ResetTargetKeyPrefixes(d.targetMgr.NextPrefix)
	d.setModel(model)
	d.Saver = func(filePath string) error { return d.model.Save(filePath) }
	d.ModifiedCallback = d.modified
	d.WillCloseCallback = d.willClose
}

// show builds the toolbar and content and places the dockable in the dock.
func (d *undoableSettingsDockable[T]) show() {
	d.Setup(d.addToStartToolbar, nil, d.initContent)
}

func (d *undoableSettingsDockable[T]) modified() bool {
	modified := d.modelModified()
	if d.applyButton != nil {
		d.applyButton.SetEnabled(modified)
		d.cancelButton.SetEnabled(modified)
	}
	return modified
}

func (d *undoableSettingsDockable[T]) willClose() bool {
	if d.promptForSave && d.modelModified() {
		switch unison.YesNoCancelDialog(fmt.Sprintf(i18n.Text("Apply changes made to\n%s?"), d.Title()), "") {
		case unison.ModalResponseDiscard:
		case unison.ModalResponseOK:
			d.spec.apply()
		case unison.ModalResponseCancel:
			return false
		}
	}
	return true
}

func (d *undoableSettingsDockable[T]) addToStartToolbar(toolbar *unison.Panel) {
	d.toolbar = toolbar

	addHelpButton(toolbar, d.spec.helpLink)

	d.applyButton, d.cancelButton = newApplyCancelButtons(toolbar, false,
		func() bool {
			d.spec.apply()
			return true
		},
		func() {
			d.promptForSave = false
			d.AttemptClose()
		})

	if d.spec.extraToolbar != nil {
		d.spec.extraToolbar(toolbar)
	}
}

// replaceModel replaces the model with another, undoably, first giving the new model's widgets their own target key
// prefixes. It is what loading a file and resetting do.
func (d *undoableSettingsDockable[T]) replaceModel(title string, model T) {
	model.ResetTargetKeyPrefixes(d.targetMgr.NextPrefix)
	d.editStructure(title, func() { d.model = model }, "")
}
