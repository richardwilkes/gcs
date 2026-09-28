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
	// stateBefore and stateAfter hold what the conversion changes outside of the data an editor edits, such as the kind
	// of an equipment container, for a target that keeps any (see containerConversionStateKeeper).
	stateBefore any
	stateAfter  any
}

type choiceConversionList[T gurps.Node[T], D gurps.EditorData[T]] struct {
	owner Rebuildable
	list  []*choiceConversion[T, D]
}

// apply puts the targets back the way they were before the conversion when undo is true, and the way they were after it
// otherwise. Any editor open on a target is discarded first: it was opened on the other kind of container, so it shows
// the wrong fields, and applying it would silently convert the target back. Its pending changes are dropped rather than
// applied, since they were made to a state the target is no longer in, and asking would put up a prompt in the middle
// of an undo.
func (c *choiceConversionList[T, D]) apply(undo bool) {
	ids := make(map[tid.TID]bool, len(c.list))
	for _, one := range c.list {
		ids[one.target.ID()] = true
	}
	discardEditorsFor(ids)
	for _, one := range c.list {
		keeper, keeps := any(one.target).(containerConversionStateKeeper)
		if undo {
			one.before.ApplyTo(one.target)
			restoreSource(one.target, one.sourceBefore, one.sourceAfter)
			if keeps {
				keeper.RestoreContainerConversionState(one.stateBefore)
			}
		} else {
			one.after.ApplyTo(one.target)
			restoreSource(one.target, one.sourceAfter, one.sourceBefore)
			if keeps {
				keeper.RestoreContainerConversionState(one.stateAfter)
			}
		}
	}
	rebuildAsModified(c.owner, true)
}

// containerKind identifies the kind of container a conversion turns a container into.
type containerKind int

const (
	// choiceContainerKind is a template choice container, which only a template may hold.
	choiceContainerKind containerKind = iota
	// groupContainerKind is a group: a container that only organizes what it holds.
	groupContainerKind
)

// installChoiceConversionHandlers installs the commands that convert the selected group containers to template choice
// containers and back. Only a template may hold choices, so only a template's lists install these, which keeps the
// commands disabled, and out of the context menus, everywhere else.
func installChoiceConversionHandlers[T gurps.Node[T], D gurps.EditorData[T]](p *PageList[T], owner Rebuildable) {
	if _, ok := owner.(*Template); !ok {
		return
	}
	installContainerKindConversionHandler[T, D](p, p.Table, owner, ConvertToChoiceContainerItemID, choiceContainerKind)
	installContainerKindConversionHandler[T, D](p, p.Table, owner, ConvertToGroupContainerItemID, groupContainerKind)
}

// installEquipmentContainerConversionHandlers installs the commands that convert equipment containers to a choice and
// to a group. Unlike a choice, a group may be held by anything, so "Convert to Group" is installed for every equipment
// list, turning a physical container into one as well as a choice. Turning a group into a physical container is left
// to "Convert to Container" (see InstallContainerConversionHandlers).
func installEquipmentContainerConversionHandlers(paneler unison.Paneler, table *unison.Table[*Node[*gurps.Equipment]], owner Rebuildable) {
	if _, ok := owner.(*Template); ok {
		installContainerKindConversionHandler[*gurps.Equipment, *gurps.EquipmentEditData](paneler, table, owner,
			ConvertToChoiceContainerItemID, choiceContainerKind)
	}
	installContainerKindConversionHandler[*gurps.Equipment, *gurps.EquipmentEditData](paneler, table, owner,
		ConvertToGroupContainerItemID, groupContainerKind)
}

func installContainerKindConversionHandler[T gurps.Node[T], D gurps.EditorData[T]](paneler unison.Paneler, table *unison.Table[*Node[T]], owner Rebuildable, id int, kind containerKind) {
	paneler.AsPanel().InstallCmdHandlers(id,
		func(_ any) bool { return len(containerKindConvertibleSelection(table, kind)) != 0 },
		func(_ any) { convertSelectedContainers[T, D](owner, table, kind) })
}

// convertSelectedContainers converts the selected rows that can be converted to the given kind, after closing any
// editor open on them and warning about any data the conversion discards, recording a single undo edit for the whole
// selection.
func convertSelectedContainers[T gurps.Node[T], D gurps.EditorData[T]](owner Rebuildable, table *unison.Table[*Node[T]], kind containerKind) {
	targets := containerKindConvertibleSelection(table, kind)
	if len(targets) == 0 {
		return
	}
	live, ok := closeEditorsBeforeConversion(table, targets)
	if !ok {
		return
	}
	if edits := convertContainerKinds[T, D](owner, live, kind); edits != nil {
		addContainerKindUndo(live, kind, edits.apply)
		rebuildAsModified(owner, true)
	}
}

// canConvertContainerKind returns true if the data is a container that can be converted to the given kind.
func canConvertContainerKind[T gurps.Node[T]](data T, kind containerKind) bool {
	switch kind {
	case choiceContainerKind:
		return gurps.CanConvertToTemplateChoiceContainer(data)
	case groupContainerKind:
		return gurps.CanConvertToGroupContainer(data)
	default:
		return false
	}
}

// convertContainerKind converts the data, which canConvertContainerKind said could be, to the given kind.
func convertContainerKind[T gurps.Node[T]](data T, kind containerKind) {
	switch kind {
	case choiceContainerKind:
		gurps.ConvertToTemplateChoiceContainer(data)
	case groupContainerKind:
		gurps.ConvertToGroupContainer(data)
	}
}

// containerKindConvertibleSelection returns the selected rows' data that can be converted to the given kind.
func containerKindConvertibleSelection[T gurps.Node[T]](table *unison.Table[*Node[T]], kind containerKind) []T {
	var list []T
	for _, row := range table.SelectedRows(false) {
		data := row.Data()
		if !xreflect.IsNil(data) && canConvertContainerKind(data, kind) {
			list = append(list, data)
		}
	}
	return list
}

// closeEditorsBeforeConversion closes any editor open on the targets of a conversion, since the fields it shows depend
// on the kind of container its row is. That has to happen before the warning is worked out, since closing an editor can
// apply its pending changes, which the warning must cover, and can even leave a row no longer convertible, so the
// targets must be looked at again afterward, in the table that is live once the editors are closed. It returns that
// table, or false if closing an editor was canceled.
func closeEditorsBeforeConversion[T gurps.Node[T]](table *unison.Table[*Node[T]], targets []T) (*unison.Table[*Node[T]], bool) {
	ids := make(map[tid.TID]bool, len(targets))
	for _, target := range targets {
		ids[target.ID()] = true
	}
	if !CloseID(ids) {
		return nil, false
	}
	return liveTable(table), true
}

// convertContainerKinds converts the selected rows that can be converted to the given kind, after warning about any
// data the conversion discards, returning what is needed to undo and redo it, or nil if nothing was converted. Any
// editor open on the rows must already have been closed (see closeEditorsBeforeConversion).
func convertContainerKinds[T gurps.Node[T], D gurps.EditorData[T]](owner Rebuildable, table *unison.Table[*Node[T]], kind containerKind) *choiceConversionList[T, D] {
	targets := containerKindConvertibleSelection(table, kind)
	if len(targets) == 0 || !confirmContainerKindConversion(targets, kind) {
		return nil
	}
	newData := newEditorData[T, D]
	edits := &choiceConversionList[T, D]{owner: owner}
	for _, target := range targets {
		conv := &choiceConversion[T, D]{
			target:       target,
			before:       newData(target),
			sourceBefore: target.GetSource(),
		}
		keeper, keeps := any(target).(containerConversionStateKeeper)
		if keeps {
			conv.stateBefore = keeper.ContainerConversionState()
		}
		convertContainerKind(target, kind)
		conv.after = newData(target)
		conv.sourceAfter = target.GetSource()
		if keeps {
			conv.stateAfter = keeper.ContainerConversionState()
		}
		edits.list = append(edits.list, conv)
	}
	return edits
}

// addContainerKindUndo records a single undo edit for a conversion to the given kind, which apply undoes or redoes.
func addContainerKindUndo[T gurps.Node[T]](table *unison.Table[*Node[T]], kind containerKind, apply func(undo bool)) {
	mgr := unison.UndoManagerFor(table)
	if mgr == nil {
		return
	}
	action := convertToGroupContainerAction
	if kind == choiceContainerKind {
		action = convertToChoiceContainerAction
	}
	mgr.Add(&unison.UndoEdit[func(bool)]{
		ID:         unison.NextUndoID(),
		EditName:   action.Title,
		UndoFunc:   func(edit *unison.UndoEdit[func(bool)]) { edit.BeforeData(true) },
		RedoFunc:   func(edit *unison.UndoEdit[func(bool)]) { edit.AfterData(false) },
		BeforeData: apply,
		AfterData:  apply,
	})
}

// confirmContainerKindConversion asks whether to go ahead with a conversion that discards data, returning true if it
// should proceed. Converting a choice container to a group always discards the choices. Converting to a choice
// container, or a physical container to a group, discards whatever the new kind can't hold, and needs no confirmation
// when there is nothing of the sort.
func confirmContainerKindConversion[T gurps.Node[T]](targets []T, kind containerKind) bool {
	var losses strings.Builder
	removesChoices := false
	for _, target := range targets {
		var list []string
		switch kind {
		case choiceContainerKind:
			list = gurps.TemplateChoiceConversionLosses(target)
		case groupContainerKind:
			removesChoices = removesChoices || gurps.IsTemplateChoiceContainer(target)
			list = gurps.GroupConversionLosses(target)
		default:
		}
		if len(list) != 0 {
			if losses.Len() != 0 {
				losses.WriteByte('\n')
			}
			fmt.Fprintf(&losses, "%s: %s", target.String(), strings.Join(list, ", "))
		}
	}
	switch {
	case removesChoices && losses.Len() == 0:
		return askToConvertChoiceContainers(i18n.Text("Converting to a group removes the choices"),
			i18n.Text("The choices are removed, leaving the groups and their contents otherwise as they are."))
	case losses.Len() == 0:
		return true
	case kind == choiceContainerKind:
		return askToConvertChoiceContainers(i18n.Text("Converting to a choice removes some data"),
			i18n.Text("A choice can't hold the following, so it will be removed:")+"\n"+losses.String())
	default:
		message := i18n.Text("A group can't hold the following, so it will be removed:") + "\n" + losses.String()
		if removesChoices {
			message += "\n" + i18n.Text("The choices will be removed as well.")
		}
		return askToConvertChoiceContainers(i18n.Text("Converting to a group removes some data"), message)
	}
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
