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
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
)

// templatePickerCopyAction is what a copy of rows carrying template choices does when it reaches its destination. The
// choices attached to a container can only be managed on a template, so everywhere else either makes them at once or
// lets them go.
type templatePickerCopyAction uint8

const (
	// templatePickerCopyRefuse abandons the copy, which is what declining to give up the choices amounts to.
	templatePickerCopyRefuse templatePickerCopyAction = iota
	// templatePickerCopyKeep carries the choices over untouched. Only a template can do this.
	templatePickerCopyKeep
	// templatePickerCopyResolve turns the copy into a partial template application: the same choice dialogs an applied
	// template presents are shown, and each container carrying choices is replaced by what was picked.
	templatePickerCopyResolve
	// templatePickerCopyStrip removes the choices from the copy, leaving the rows otherwise as they were.
	templatePickerCopyStrip
)

// pendingTemplatePickerAction records what the drop about to be processed is to do about the template choices it
// carries. It is set by the guard installed on the table's drop callback and consumed by the did-drop callback, which
// run one after the other on the UI thread as part of a single drop.
var pendingTemplatePickerAction templatePickerCopyAction

// templatePickerStripDialog asks whether a copy that cannot keep its template choices should go ahead without them.
// Held in a variable so that headless tests, which have no way to respond to a modal dialog, can substitute their own
// answer.
var templatePickerStripDialog = confirmTemplatePickerStrip

// rowsHaveTemplatePickers reports whether any of the rows, or any row beneath them, is a container carrying template
// choices.
func rowsHaveTemplatePickers[T gurps.Node[T]](rows []T) bool {
	for _, row := range rows {
		if !row.Container() {
			continue
		}
		if provider, ok := any(row).(gurps.TemplatePickerProvider); ok {
			if _, picker := provider.TemplatePickerData(); picker != nil && !picker.IsZero() {
				return true
			}
		}
		if rowsHaveTemplatePickers(row.NodeChildren()) {
			return true
		}
	}
	return false
}

// stripTemplatePickers removes the template choices from the rows and everything beneath them, leaving them otherwise
// untouched.
func stripTemplatePickers[T gurps.Node[T]](rows []T) {
	for _, row := range rows {
		if !row.Container() {
			continue
		}
		if provider, ok := any(row).(gurps.TemplatePickerProvider); ok {
			if _, picker := provider.TemplatePickerData(); picker != nil {
				*picker = gurps.TemplatePicker{}
			}
		}
		stripTemplatePickers(row.NodeChildren())
	}
}

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

// guardTemplatePickerDrop reports whether a drop into the table may proceed, and records what the did-drop callback is
// to do about the template choices it carries once the rows have landed (see decideTemplatePickerCopy). A move is
// always let through: the providers only allow one within a single dockable (see their DropShouldMoveData), so it adds
// rows to no other document.
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
	if !rowsHaveTemplatePickers(ExtractNodeDataFromList(data.Rows)) {
		return true
	}
	pendingTemplatePickerAction = decideTemplatePickerCopy(table)
	return pendingTemplatePickerAction != templatePickerCopyRefuse
}

// resolveTemplatePickers replaces the rows that have just been added to the table with the result of making their
// template choices, the same way applying a template does, leaving the rows around them where they are. Returns false
// if a choice was canceled, in which case the added rows are removed again, leaving the table as it was before they
// arrived. The choice processing is passed in so that headless tests, which have no way to respond to the dialogs it
// would otherwise present, can substitute their own.
func resolveTemplatePickers[T gurps.Node[T]](table *unison.Table[*Node[T]], added []*Node[T], process func(rows []*Node[T]) (revised []*Node[T], abort bool)) bool {
	if table == nil || len(added) == 0 {
		return true
	}
	// The added rows may not all share a parent -- nothing about a drop requires it -- so they are processed and
	// spliced back a parent at a time. Order is preserved so that the outcome doesn't depend on map iteration.
	type siblingGroup struct {
		parent *Node[T]
		rows   []*Node[T]
	}
	var groups []*siblingGroup
	byParent := make(map[*Node[T]]*siblingGroup)
	for _, row := range added {
		group, ok := byParent[row.Parent()]
		if !ok {
			group = &siblingGroup{parent: row.Parent()}
			byParent[row.Parent()] = group
			groups = append(groups, group)
		}
		group.rows = append(group.rows, row)
	}
	// Every group is resolved before any of them is put back, since canceling any one of the choices abandons the
	// whole operation and must leave nothing of it behind.
	revised := make([][]*Node[T], len(groups))
	for i, group := range groups {
		result, abort := process(group.rows)
		if abort {
			for _, g := range groups {
				replaceChildRows(table, g.parent, g.rows, nil)
			}
			table.ClearSelection()
			table.SyncToModel()
			return false
		}
		revised[i] = result
	}
	selection := make(map[tid.TID]bool)
	for i, group := range groups {
		replaceChildRows(table, group.parent, group.rows, revised[i])
		for _, row := range revised[i] {
			selection[row.ID()] = true
		}
	}
	table.SyncToModel()
	table.SetSelectionMap(selection)
	return true
}

// replaceChildRows replaces the given rows, which must all be children of parent -- or root rows, when parent is nil
// -- with the replacements, which land where the first of the rows was.
func replaceChildRows[T gurps.Node[T]](table *unison.Table[*Node[T]], parent *Node[T], rows, replacements []*Node[T]) {
	siblings := table.RootRows()
	if parent != nil {
		siblings = parent.Children()
	}
	removing := make(map[tid.TID]bool, len(rows))
	for _, row := range rows {
		removing[row.ID()] = true
	}
	revised := make([]*Node[T], 0, len(siblings)-len(rows)+len(replacements))
	inserted := false
	for _, sibling := range siblings {
		if removing[sibling.ID()] {
			if !inserted {
				revised = append(revised, replacements...)
				inserted = true
			}
			continue
		}
		revised = append(revised, sibling)
	}
	if !inserted {
		revised = append(revised, replacements...)
	}
	if parent == nil {
		table.SetRootRows(revised)
	} else {
		parent.SetChildren(revised)
	}
}
