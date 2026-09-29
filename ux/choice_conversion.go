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

// choiceConversion holds a node's state on either side of a conversion to or from a template choice container or a
// modifier choice, so that the conversion can be undone and redone.
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
	// optionsBefore and optionsAfter hold the enabled states of the modifiers in the targets' trees, which converting a
	// modifier container may change. They are taken once for the whole conversion, since targets may share a tree.
	optionsBefore map[gurps.GeneralModifier]bool
	optionsAfter  map[gurps.GeneralModifier]bool
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
		if undo {
			one.before.ApplyTo(one.target)
			restoreSource(one.target, one.sourceBefore, one.sourceAfter)
		} else {
			one.after.ApplyTo(one.target)
			restoreSource(one.target, one.sourceAfter, one.sourceBefore)
		}
	}
	if undo {
		restoreModifierEnabledStates(c.optionsBefore)
	} else {
		restoreModifierEnabledStates(c.optionsAfter)
	}
	rebuildAsModified(c.owner, true)
}

// containerKind identifies the kind of container a conversion turns a container into.
type containerKind int

const (
	// choiceContainerKind is a template choice container, which only a template may hold, or a modifier choice.
	choiceContainerKind containerKind = iota
	// groupContainerKind is a group: a container that only organizes what it holds.
	groupContainerKind
	// physicalContainerKind is a piece of equipment, such as a backpack, that holds other equipment. Only equipment
	// has these.
	physicalContainerKind
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

// installEquipmentContainerConversionHandlers installs the commands that convert equipment among its kinds of
// container. Unlike a choice, a group and a physical container may be held by anything, so these are installed for
// every equipment list, after the ones that convert an item to a container and back, since "Convert to Container" is
// extended here to also turn a group into a physical container.
func installEquipmentContainerConversionHandlers(paneler unison.Paneler, table *unison.Table[*Node[*gurps.Equipment]], owner Rebuildable) {
	if _, ok := owner.(*Template); ok {
		installContainerKindConversionHandler[*gurps.Equipment, *gurps.EquipmentEditData](paneler, table, owner,
			ConvertToChoiceContainerItemID, choiceContainerKind)
	}
	installContainerKindConversionHandler[*gurps.Equipment, *gurps.EquipmentEditData](paneler, table, owner,
		ConvertToGroupContainerItemID, groupContainerKind)
	paneler.AsPanel().InstallCmdHandlers(ConvertToContainerItemID,
		func(_ any) bool {
			return CanConvertToContainer(table) ||
				len(containerKindConvertibleSelection(table, physicalContainerKind)) != 0
		},
		func(_ any) { convertEquipmentToPhysicalContainers(owner, table) })
}

// installModifierChoiceConversionHandlers installs the commands that convert the selected modifier groups to modifier
// choices and back. Unlike a template choice, a modifier choice may be held anywhere modifiers are, so every modifier
// table installs these: a library list's and the one in a trait or equipment editor. The owner is asked for when a
// command runs, since an editor's table is built before it is attached to the editor that owns it.
func installModifierChoiceConversionHandlers[T gurps.Node[T], D gurps.EditorData[T]](paneler unison.Paneler, table *unison.Table[*Node[T]], owner func() Rebuildable) {
	for _, kind := range []containerKind{choiceContainerKind, groupContainerKind} {
		id := ConvertToChoiceContainerItemID
		if kind == groupContainerKind {
			id = ConvertToGroupContainerItemID
		}
		paneler.AsPanel().InstallCmdHandlers(id,
			func(_ any) bool { return len(containerKindConvertibleSelection(table, kind)) != 0 },
			func(_ any) { convertSelectedContainers[T, D](owner(), table, kind) })
	}
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
		return gurps.CanConvertToTemplateChoiceContainer(data) || gurps.CanConvertToModifierChoice(data)
	case groupContainerKind:
		if gurps.IsTemplateChoiceContainer(data) || gurps.IsModifierChoice(data) {
			return true
		}
		eqp, ok := any(data).(*gurps.Equipment)
		return ok && eqp.CanConvertToGroup()
	case physicalContainerKind:
		eqp, ok := any(data).(*gurps.Equipment)
		return ok && eqp.CanConvertToPhysicalContainer()
	default:
		return false
	}
}

// convertContainerKind converts the data, which canConvertContainerKind said could be, to the given kind.
func convertContainerKind[T gurps.Node[T]](data T, kind containerKind) {
	switch kind {
	case choiceContainerKind:
		if gurps.CanConvertToModifierChoice(data) {
			gurps.ConvertToModifierChoice(data)
		} else {
			gurps.ConvertToTemplateChoiceContainer(data)
		}
	case groupContainerKind:
		switch {
		case gurps.IsTemplateChoiceContainer(data):
			gurps.ConvertFromTemplateChoiceContainer(data)
		case gurps.IsModifierChoice(data):
			gurps.ConvertFromModifierChoice(data)
		default:
			if eqp, ok := any(data).(*gurps.Equipment); ok {
				eqp.ConvertToGroup()
			}
		}
	case physicalContainerKind:
		if eqp, ok := any(data).(*gurps.Equipment); ok {
			eqp.ConvertToPhysicalContainer()
		}
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
	edits := &choiceConversionList[T, D]{owner: owner, optionsBefore: modifierEnabledStates(targets)}
	for _, target := range targets {
		conv := &choiceConversion[T, D]{
			target:       target,
			before:       newData(target),
			sourceBefore: target.GetSource(),
		}
		convertContainerKind(target, kind)
		conv.after = newData(target)
		conv.sourceAfter = target.GetSource()
		edits.list = append(edits.list, conv)
	}
	edits.optionsAfter = modifierEnabledStates(targets)
	return edits
}

// addContainerKindUndo records a single undo edit for a conversion to the given kind, which apply undoes or redoes.
func addContainerKindUndo[T gurps.Node[T]](table *unison.Table[*Node[T]], kind containerKind, apply func(undo bool)) {
	mgr := unison.UndoManagerFor(table)
	if mgr == nil {
		return
	}
	var action *unison.Action
	switch kind {
	case choiceContainerKind:
		action = convertToChoiceContainerAction
	case groupContainerKind:
		action = convertToGroupContainerAction
	default:
		action = convertToContainerAction
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

// convertEquipmentToPhysicalContainers carries out "Convert to Container" for an equipment list: the selected items
// become containers, as they do for any list, and the selected groups become physical containers, all recorded as a
// single undo edit. As with the other conversions, any editor open on the rows is closed first.
func convertEquipmentToPhysicalContainers(owner Rebuildable, table *unison.Table[*Node[*gurps.Equipment]]) {
	var targets []*gurps.Equipment
	for _, row := range table.SelectedRows(false) {
		if data := row.Data(); !xreflect.IsNil(data) &&
			((!data.Container() && data.CanConvertToFromContainer()) || data.CanConvertToPhysicalContainer()) {
			targets = append(targets, data)
		}
	}
	if len(targets) == 0 {
		return
	}
	table, ok := closeEditorsBeforeConversion(table, targets)
	if !ok {
		return
	}
	before, after := convertContainersWithoutUndo(owner, table, true)
	groups := convertContainerKinds[*gurps.Equipment, *gurps.EquipmentEditData](owner, table, physicalContainerKind)
	if before == nil && groups == nil {
		return
	}
	addContainerKindUndo(table, physicalContainerKind, func(undo bool) {
		if groups != nil {
			groups.apply(undo)
		}
		if before != nil {
			if undo {
				before.Apply()
			} else {
				after.Apply()
			}
		}
	})
	rebuildAsModified(owner, true)
}

// confirmContainerKindConversion asks whether to go ahead with a conversion that discards data, returning true if it
// should proceed. Converting a choice container to a group always discards the choices. Converting to a choice
// container, or a physical container to a group, discards whatever the new kind can't hold, and needs no confirmation
// when there is nothing of the sort, which is always so for a modifier group, since a modifier choice holds everything
// it does. Converting a group to a physical container discards nothing.
func confirmContainerKindConversion[T gurps.Node[T]](targets []T, kind containerKind) bool {
	var losses strings.Builder
	removesChoices := false
	for _, target := range targets {
		var list []string
		switch kind {
		case choiceContainerKind:
			list = gurps.TemplateChoiceConversionLosses(target)
		case groupContainerKind:
			if gurps.IsTemplateChoiceContainer(target) || gurps.IsModifierChoice(target) {
				removesChoices = true
			} else if eqp, ok := any(target).(*gurps.Equipment); ok {
				list = eqp.GroupConversionLosses()
			}
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

// modifierEnabledStates returns the enabled state of each modifier in the trees holding the nodes.
func modifierEnabledStates[T gurps.Node[T]](nodes []T) map[gurps.GeneralModifier]bool {
	states := make(map[gurps.GeneralModifier]bool)
	var zero T
	if _, ok := any(zero).(gurps.GeneralModifier); !ok {
		return states
	}
	for _, node := range nodes {
		for parent := node.Parent(); !xreflect.IsNil(parent); parent = node.Parent() {
			node = parent
		}
		gurps.Traverse(func(one T) bool {
			if gm, ok := any(one).(gurps.GeneralModifier); ok {
				states[gm] = gm.Enabled()
			}
			return false
		}, false, true, node)
	}
	return states
}

// restoreModifierEnabledStates puts back the enabled states modifierEnabledStates returned.
func restoreModifierEnabledStates(states map[gurps.GeneralModifier]bool) {
	for gm, enabled := range states {
		gm.SetEnabled(enabled)
	}
}
