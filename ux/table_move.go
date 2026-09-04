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
	"slices"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// MoveDirection identifies one of the commands that reposition the selected rows of a table from the keyboard. Each is
// the equivalent of a single drag: up or down by one place among the row's siblings, out of the container the row is
// in to sit just above that container, or into the container sitting just below the row to become its first child.
// Moving out and then back in returns a row to where it was only if it was the container's first child: out puts the
// row directly above its container no matter where inside it the row was, and in always puts it at the front.
type MoveDirection int

// Possible values for MoveDirection.
const (
	MoveUp MoveDirection = iota
	MoveDown
	MoveOutOfContainer
	MoveIntoContainer
)

// Title returns the Title of the command, which also serves as the name of the undo edit it records.
func (d MoveDirection) Title() string {
	switch d {
	case MoveDown:
		return i18n.Text("Move Down")
	case MoveOutOfContainer:
		return i18n.Text("Move Out of Container")
	case MoveIntoContainer:
		return i18n.Text("Move Into Container")
	default:
		return i18n.Text("Move Up")
	}
}

// CmdID returns the command ID the direction is routed through.
func (d MoveDirection) CmdID() int {
	switch d {
	case MoveDown:
		return MoveDownItemID
	case MoveOutOfContainer:
		return MoveOutOfContainerItemID
	case MoveIntoContainer:
		return MoveIntoContainerItemID
	default:
		return MoveUpItemID
	}
}

// InstallMoveSelectionHandlers installs the command handlers for the keyboard repositioning commands on a table. They
// go on every table that supports rearranging its rows by drag and drop, since they are the keyboard equivalents of
// those drags, and on the table itself rather than its owner so that a rebuild which replaces the table brings them
// along with the new one.
func InstallMoveSelectionHandlers[T gurps.Node[T]](table *unison.Table[*Node[T]]) {
	for _, dir := range []MoveDirection{MoveUp, MoveDown, MoveOutOfContainer, MoveIntoContainer} {
		table.InstallCmdHandlers(dir.CmdID(),
			func(_ any) bool { return CanMoveSelection(table, dir) },
			func(_ any) { MoveSelection(table, dir) })
	}
}

// CanMoveSelection returns true if MoveSelection would reposition at least one of the selected rows.
func CanMoveSelection[T gurps.Node[T]](table *unison.Table[*Node[T]], dir MoveDirection) bool {
	provider, items, selected, ok := selectionToMove(table)
	if !ok {
		return false
	}
	for _, item := range items {
		if canMove(provider, item, dir, selected) {
			return true
		}
	}
	return false
}

// MoveSelection repositions the selected rows of the table one step in the given direction, as a drag of each of them
// would, then reports the change by rebuilding the table's owner. Up and down move a row one place among its siblings;
// out moves it from its container to the place just above that container; into moves it from just above a container
// to be that container's first child. The selection stays on the moved rows and the change is recorded as a single
// undo edit. Nothing at all happens, and no edit is recorded, when none of the selected rows can move.
//
// Each selected row is moved independently, so a selection spread over several containers moves within each of them.
// Rows that are selected together keep their order relative to one another: a row doesn't move up past a selected
// sibling, so a run of selected siblings moves as a block, and one that has reached the top of its list stops there,
// holding back any selected siblings directly below it. The same holds in the other direction, and a run of selected
// siblings above a container enters it as a block, in order. A container moves with everything in it, so the
// descendants of a selected container are left alone even when they are selected as well, and nothing is moved into
// a selected container.
//
// A drag within a table finishes by giving the provider a look at the rows that landed -- for skills and spells that
// fills in a blank tech level from the entity and refreshes their levels -- and by clearing the Preconfigured flag on
// them anywhere but in a template (see didDropCallback), so a move does the same to stay interchangeable with one. The
// prompts for modifiers and nameables that a drop may go on to raise are left out: they configure rows arriving on a
// sheet, which a move never brings.
func MoveSelection[T gurps.Node[T]](table *unison.Table[*Node[T]], dir MoveDirection) {
	provider, items, selected, ok := selectionToMove(table)
	if !ok {
		return
	}
	// The data has to be captured before anything moves, but the edit is only recorded once something has, which the
	// loop below finds out as it goes; a selection none of whose rows can move therefore costs one serialization and
	// nothing more, rather than a full pass over it up front to establish that.
	var undo *unison.UndoEdit[*TableUndoEditData[T]]
	var opened []T
	mgr := unison.UndoManagerFor(table)
	if mgr != nil {
		undo = &unison.UndoEdit[*TableUndoEditData[T]]{
			ID:       unison.NextUndoID(),
			EditName: dir.Title(),
			// The containers the move opened, collected by the loop below, are closed again before undo puts the data
			// back, so that undo leaves no trace, and reopened before redo does, so that the rows redo moves into them
			// are showing and can be selected. Their open state is kept in the global settings under their IDs, so it
			// makes no difference that the objects themselves are replaced when the data is deserialized.
			UndoFunc: func(e *unison.UndoEdit[*TableUndoEditData[T]]) {
				setContainersOpen(opened, false)
				e.BeforeData.Apply()
			},
			RedoFunc: func(e *unison.UndoEdit[*TableUndoEditData[T]]) {
				setContainersOpen(opened, true)
				e.AfterData.Apply()
			},
			AbsorbFunc: func(_ *unison.UndoEdit[*TableUndoEditData[T]], _ unison.Undoable) bool { return false },
			BeforeData: NewTableUndoEditData(table),
		}
	}
	if dir == MoveDown || dir == MoveIntoContainer {
		// These two act on what sits directly below a row, so the rows are taken from the bottom up: that way a
		// selected sibling below has already moved on before the row above it looks at what is beneath it. Moving
		// down, a run of selected siblings then moves as a block instead of the top one being blocked by the one
		// below it; moving into a container, each row of the run finds the container directly beneath it in turn and
		// goes in ahead of the one that just did, so the run arrives in order.
		slices.Reverse(items)
	}
	moved := false
	for _, item := range items {
		if !canMove(provider, item, dir, selected) {
			continue
		}
		switch dir {
		case MoveUp:
			ok = swapWithSibling(provider, item, -1)
		case MoveDown:
			ok = swapWithSibling(provider, item, 1)
		case MoveOutOfContainer:
			ok = moveOutOfParent(provider, item)
		case MoveIntoContainer:
			var container T
			if container, ok = moveIntoNextSibling(provider, item); ok && !container.IsOpen() {
				// Opened so that the row doesn't vanish from view -- and out of the selection, which can only hold rows
				// that are showing -- the moment it is moved.
				container.SetOpen(true)
				opened = append(opened, container)
			}
		}
		if ok {
			moved = true
		}
	}
	if !moved {
		return
	}
	table.SyncToModel()
	table.SetSelectionMap(selected)
	provider.ProcessDropData(table, table)
	clearPreconfiguredFlag(table, nil)
	table.ScrollRowCellIntoView(table.LastSelectedRowIndex(), 0)
	table.ScrollRowCellIntoView(table.FirstSelectedRowIndex(), 0)
	if mgr != nil && undo != nil {
		undo.AfterData = NewTableUndoEditData(table)
		mgr.Add(undo)
	}
	rebuildAsModified(table.AncestorOrSelf[Rebuildable](), true)
}

// selectionToMove returns what the move commands work from: the table's provider, the data behind the selected rows,
// in table order, and the selection itself, which decides which siblings a row may move past. It reports false when the
// table has no provider, has nothing selected, or is showing search results, whose flat list has no order to rearrange.
func selectionToMove[T gurps.Node[T]](table *unison.Table[*Node[T]]) (provider TableProvider[T], items []T,
	selected map[tid.TID]bool, ok bool,
) {
	if provider, ok = any(table.Model).(TableProvider[T]); !ok || !HasSelectionAndNotFiltered(table) {
		return nil, nil, nil, false
	}
	var zero T
	rows := table.SelectedRows(true)
	items = make([]T, 0, len(rows))
	for _, row := range rows {
		if item := row.Data(); item != zero {
			items = append(items, item)
		}
	}
	return provider, items, table.CopySelectionMap(), true
}

// setContainersOpen opens or closes each of the containers.
func setContainersOpen[T gurps.Node[T]](containers []T, open bool) {
	for _, container := range containers {
		container.SetOpen(open)
	}
}

// canMove returns true if the item would be moved in the given direction. A row moves out of a container whenever it
// is in one, and into one whenever the sibling directly below it is a container that isn't itself selected. It moves up
// or down whenever there is an unselected sibling on that side of it: the selected siblings between the two travel
// along with it, so only an unselected one gives it somewhere to go.
func canMove[T gurps.Node[T]](provider TableProvider[T], item T, dir MoveDirection, selected map[tid.TID]bool) bool {
	var zero T
	if dir == MoveOutOfContainer {
		return item.Parent() != zero
	}
	siblings, index := siblingsOf(provider, item)
	if index == -1 {
		return false
	}
	if dir == MoveIntoContainer {
		if index+1 >= len(siblings) {
			return false
		}
		next := siblings[index+1]
		return next.Container() && !selected[next.ID()]
	}
	var candidates []T
	if dir == MoveUp {
		candidates = siblings[:index]
	} else {
		candidates = siblings[index+1:]
	}
	for _, sibling := range candidates {
		if !selected[sibling.ID()] {
			return true
		}
	}
	return false
}

// swapWithSibling exchanges the item with the sibling the given number of places away from it, reporting whether it
// did. The caller has already established (see canMove) that the row can move in that direction, and, given the order
// MoveSelection takes the rows in, that the sibling next to it on that side is by then an unselected one.
func swapWithSibling[T gurps.Node[T]](provider TableProvider[T], item T, offset int) bool {
	siblings, index := siblingsOf(provider, item)
	other := index + offset
	if index == -1 || other < 0 || other >= len(siblings) {
		return false
	}
	siblings = slices.Clone(siblings)
	siblings[index], siblings[other] = siblings[other], siblings[index]
	setSiblings(provider, item.Parent(), siblings)
	return true
}

// moveOutOfParent takes the item out of its container and puts it into the list that container is in, immediately
// above the container, reporting whether it did. Rows taken out of the same container one after another in
// top-to-bottom order therefore end up above it in the order they had inside it. Like the other movers, it leaves
// everything alone if the item isn't where its parent says it is, rather than inserting it above the container without
// having removed it from anywhere.
func moveOutOfParent[T gurps.Node[T]](provider TableProvider[T], item T) bool {
	var zero T
	parent := item.Parent()
	if parent == zero {
		return false
	}
	children := parent.NodeChildren()
	index := slices.Index(children, item)
	if index == -1 {
		return false
	}
	parent.SetChildren(slices.Delete(slices.Clone(children), index, index+1))
	list, parentIndex := siblingsOf(provider, parent)
	list = slices.Clone(list)
	if parentIndex == -1 {
		parentIndex = len(list)
	}
	grandparent := parent.Parent()
	setSiblings(provider, grandparent, slices.Insert(list, parentIndex, item))
	item.SetParent(grandparent)
	return true
}

// moveIntoNextSibling takes the item out of the list it is in and makes it the first child of the container that was
// directly below it, returning that container along with whether the move was made. The caller has already established
// (see canMove) that there is such a container.
func moveIntoNextSibling[T gurps.Node[T]](provider TableProvider[T], item T) (container T, moved bool) {
	siblings, index := siblingsOf(provider, item)
	if index == -1 || index+1 >= len(siblings) {
		return container, false
	}
	container = siblings[index+1]
	setSiblings(provider, item.Parent(), slices.Delete(slices.Clone(siblings), index, index+1))
	container.SetChildren(slices.Insert(slices.Clone(container.NodeChildren()), 0, item))
	item.SetParent(container)
	return container, true
}

// siblingsOf returns the list the item sits in -- its parent's children, or the provider's top-level data for a row
// with no parent -- along with the item's index in that list, or -1 if it isn't there. The list is the provider's own,
// so it must be cloned before being altered.
func siblingsOf[T gurps.Node[T]](provider TableProvider[T], item T) (siblings []T, index int) {
	var zero T
	if parent := item.Parent(); parent == zero {
		siblings = provider.RootData()
	} else {
		siblings = parent.NodeChildren()
	}
	return siblings, slices.Index(siblings, item)
}

// setSiblings stores a list back where siblingsOf got it from: the children of the given parent, or the provider's
// top-level data when there is no parent.
func setSiblings[T gurps.Node[T]](provider TableProvider[T], parent T, list []T) {
	var zero T
	if parent == zero {
		provider.SetRootData(list)
	} else {
		parent.SetChildren(list)
	}
}
