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
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
)

// pendingTemplatePickerAction records what the drop about to be processed is to do about the template choices it
// carries. It is set by the guard installed on the table's drop callback and consumed by the did-drop callback, which
// run one after the other on the UI thread as part of a single drop.
//
// Only templatePickerCopyStrip is ever handed on this way. The guard settles the other three itself: a refusal and a
// partial template application both turn the drop away, and keeping the choices is what an unguarded drop already
// does.
var pendingTemplatePickerAction templatePickerCopyAction

// templatePickerStripDialog asks whether a copy that cannot keep its template choices should go ahead without them.
// Held in a variable so that headless tests, which have no way to respond to a modal dialog, can substitute their own
// answer.
var templatePickerStripDialog = confirmTemplatePickerStrip

// decideTemplatePickerCopy reports what a copy of rows carrying template choices onto the given destination is to do
// about them. Only a template takes them as they are, since nothing else provides a way to manage them once they have
// arrived. A character sheet makes the choices on the spot, which is what applying a template does. Anywhere else --
// a library list, most of all -- can only take the rows without the choices, which is confirmed here before it does.
func decideTemplatePickerCopy(dst unison.Paneler) templatePickerCopyAction {
	switch unison.AncestorOrSelf[unison.Dockable](dst).(type) {
	case *Template:
		return templatePickerCopyKeep
	case *Sheet:
		return templatePickerCopyResolve
	case *LootSheet:
		// Nothing carries template choices onto a loot sheet today: a loot sheet holds equipment and notes, and
		// neither can carry them. Once equipment can, both consumers of this answer -- copySelectionTo and
		// guardTemplatePickerDrop -- have to be able to run a partial application against a loot sheet, which today
		// they cannot: each looks for a *Sheet and quietly does nothing when it finds none.
		return templatePickerCopyResolve
	default:
		if templatePickerStripDialog() {
			return templatePickerCopyStrip
		}
		return templatePickerCopyRefuse
	}
}

// confirmTemplatePickerStrip warns that the template choices cannot come along and asks whether to copy the rows
// without them, returning true if that is what the user wants.
func confirmTemplatePickerStrip() bool {
	dialog, err := unison.NewDialog(unison.DefaultDialogTheme.WarningIcon, unison.DefaultDialogTheme.WarningIconInk,
		unison.NewMessagePanel(i18n.Text("Template choices cannot be kept here"),
			i18n.Text(`One or more of the containers being copied has template choices attached to it.
Only a template provides a way to manage those choices, so they will be
removed from the copy. Everything else is copied unchanged.`)),
		[]*unison.DialogButtonInfo{
			unison.NewCancelButtonInfo(),
			{
				Title:        i18n.Text("Copy Without Choices"),
				ResponseCode: unison.ModalResponseOK,
			},
		})
	if err != nil {
		errs.Log(err)
		return false
	}
	return dialog.RunModal() == unison.ModalResponseOK
}

// guardTemplatePickerDrop reports whether a drop into the table may proceed, deciding what is to become of the
// template choices it carries before unison has inserted anything (see decideTemplatePickerCopy). A move is always let
// through: the providers only allow one within a single dockable (see their DropShouldMoveData), so it adds rows to no
// other document.
//
// A drop onto a sheet is performed here rather than let through, since it is a partial template application and the
// rows unison would insert are the unresolved ones. Everything the drop would have done is done by the template
// pipeline instead, so the drop itself is turned away.
func guardTemplatePickerDrop[T gurps.Node[T]](table *unison.Table[*Node[T]], provider TableProvider[T], di drag.Info) bool {
	pendingTemplatePickerAction = templatePickerCopyKeep
	if !table.Enabled() || table.IsFiltered() || !di.HasDataType(provider.DragKey().UTI) {
		// The drop is going to be turned away regardless (see unison's TableDrop.CanAcceptDropCallback), so there is
		// nothing to say about the choices it carries.
		return true
	}
	data, ok := draggedTableData.(*unison.TableDragData[*Node[T]])
	if !ok || data.Table == table {
		return true
	}
	if provider.DropShouldMoveData(data.Table, table) {
		return true
	}
	if !gurps.HasTemplatePickerData(ExtractNodeDataFromList(data.Rows)...) {
		return true
	}
	pendingTemplatePickerAction = decideTemplatePickerCopy(table)
	switch pendingTemplatePickerAction {
	case templatePickerCopyResolve:
		// Handled in full here, so nothing is left for the did-drop callback to act on. A canceled choice leaves the
		// sheet untouched, which is the same outcome as a refusal from the drop's point of view.
		pendingTemplatePickerAction = templatePickerCopyKeep
		if sheet := unison.AncestorOrSelf[*Sheet](table); sheet != nil {
			applyPartialTemplate(sheet, data.Rows)
		}
		return false
	case templatePickerCopyRefuse:
		return false
	default:
		return true
	}
}
