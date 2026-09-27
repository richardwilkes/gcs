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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

// choiceConversion holds a node's state on either side of a conversion to or from a template choice container, so that
// the conversion can be undone and redone.
type choiceConversion[T gurps.Node[T], D gurps.EditorData[T]] struct {
	target       T
	before       D
	after        D
	sourceBefore gurps.Source
	sourceAfter  gurps.Source
}

type choiceConversionList[T gurps.Node[T], D gurps.EditorData[T]] struct {
	owner Rebuildable
	list  []*choiceConversion[T, D]
}

func (c *choiceConversionList[T, D]) apply(undo bool) {
	for _, one := range c.list {
		if undo {
			one.before.ApplyTo(one.target)
			restoreSource(one.target, one.sourceBefore, one.sourceAfter)
		} else {
			one.after.ApplyTo(one.target)
			restoreSource(one.target, one.sourceAfter, one.sourceBefore)
		}
	}
	rebuildAsModified(c.owner, true)
}

// installChoiceConversionHandlers installs the commands that convert the selected group containers to template choice
// containers and back. Only a template may hold choices, so only a template's lists install these, which keeps the
// commands disabled, and out of the context menus, everywhere else.
func installChoiceConversionHandlers[T gurps.Node[T], D gurps.EditorData[T]](p *PageList[T], owner Rebuildable) {
	if _, ok := owner.(*Template); !ok {
		return
	}
	p.InstallCmdHandlers(ConvertToChoiceContainerItemID,
		func(_ any) bool { return len(choiceConvertibleSelection(p.Table, true)) != 0 },
		func(_ any) { convertChoiceContainers[T, D](owner, p.Table, true) })
	p.InstallCmdHandlers(ConvertToGroupContainerItemID,
		func(_ any) bool { return len(choiceConvertibleSelection(p.Table, false)) != 0 },
		func(_ any) { convertChoiceContainers[T, D](owner, p.Table, false) })
}

// choiceConvertibleSelection returns the selected rows' data that can be converted to template choice containers when
// toChoice is true, or back to group containers otherwise.
func choiceConvertibleSelection[T gurps.Node[T]](table *unison.Table[*Node[T]], toChoice bool) []T {
	var list []T
	for _, row := range table.SelectedRows(false) {
		data := row.Data()
		if xreflect.IsNil(data) {
			continue
		}
		if (toChoice && gurps.CanConvertToTemplateChoiceContainer(data)) ||
			(!toChoice && gurps.IsTemplateChoiceContainer(data)) {
			list = append(list, data)
		}
	}
	return list
}

// convertChoiceContainers converts the selected rows that can be converted in the given direction, after warning about
// any data the conversion discards, recording a single undo edit for the whole selection. Any editor open on one of the
// rows is closed first, since the fields it shows depend on whether the row is a choice container.
func convertChoiceContainers[T gurps.Node[T], D gurps.EditorData[T]](owner Rebuildable, table *unison.Table[*Node[T]], toChoice bool) {
	targets := choiceConvertibleSelection(table, toChoice)
	if len(targets) == 0 || !confirmChoiceConversion(targets, toChoice) {
		return
	}
	ids := make(map[tid.TID]bool, len(targets))
	for _, target := range targets {
		ids[target.ID()] = true
	}
	if !CloseID(ids) {
		return
	}
	newData := newEditorData[T, D]
	edits := &choiceConversionList[T, D]{owner: owner}
	for _, target := range targets {
		conv := &choiceConversion[T, D]{
			target:       target,
			before:       newData(target),
			sourceBefore: target.GetSource(),
		}
		if toChoice {
			gurps.ConvertToTemplateChoiceContainer(target)
		} else {
			gurps.ConvertFromTemplateChoiceContainer(target)
		}
		conv.after = newData(target)
		conv.sourceAfter = target.GetSource()
		edits.list = append(edits.list, conv)
	}
	if mgr := unison.UndoManagerFor(table); mgr != nil {
		action := convertToGroupContainerAction
		if toChoice {
			action = convertToChoiceContainerAction
		}
		mgr.Add(&unison.UndoEdit[*choiceConversionList[T, D]]{
			ID:         unison.NextUndoID(),
			EditName:   action.Title,
			UndoFunc:   func(edit *unison.UndoEdit[*choiceConversionList[T, D]]) { edit.BeforeData.apply(true) },
			RedoFunc:   func(edit *unison.UndoEdit[*choiceConversionList[T, D]]) { edit.AfterData.apply(false) },
			BeforeData: edits,
			AfterData:  edits,
		})
	}
	rebuildAsModified(owner, true)
}

// confirmChoiceConversion asks whether to go ahead with a conversion that discards data, returning true if it should
// proceed. Converting to a group container always discards the choices. Converting to a choice container discards
// whatever a choice container can't hold, and needs no confirmation when there is nothing of the sort.
func confirmChoiceConversion[T gurps.Node[T]](targets []T, toChoice bool) bool {
	if !toChoice {
		return askToConvertChoiceContainers(i18n.Text("Converting to a group removes the choices"),
			i18n.Text("The choices are removed, leaving the groups and their contents otherwise as they are."))
	}
	var buffer strings.Builder
	for _, target := range targets {
		if losses := gurps.TemplateChoiceConversionLosses(target); len(losses) != 0 {
			if buffer.Len() != 0 {
				buffer.WriteByte('\n')
			}
			fmt.Fprintf(&buffer, "%s: %s", target.String(), strings.Join(losses, ", "))
		}
	}
	if buffer.Len() == 0 {
		return true
	}
	return askToConvertChoiceContainers(i18n.Text("Converting to a choice removes some data"),
		i18n.Text("A choice can't hold the following, so it will be removed:")+"\n"+buffer.String())
}

// askToConvertChoiceContainers asks whether to go ahead with a conversion that discards what the message describes. It
// is held in a variable so that tests can substitute a non-interactive implementation.
var askToConvertChoiceContainers = func(title, message string) bool {
	convert := unison.NewOKButtonInfo()
	convert.Title = i18n.Text("Convert")
	dialog, err := unison.NewDialog(unison.DefaultDialogTheme.WarningIcon, unison.DefaultDialogTheme.WarningIconInk,
		unison.NewMessagePanel(title, message), []*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), convert})
	if err != nil {
		errs.Log(err)
		return false
	}
	return dialog.RunModal() == unison.ModalResponseOK
}
