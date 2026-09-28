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
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

// ConvertableContainer defines the methods a node must implement to be convertible to and from a container. Rows are
// matched against it at runtime, so nodes that do not implement it are simply skipped.
type ConvertableContainer interface {
	Container() bool
	CanConvertToFromContainer() bool
	ConvertToContainer()
	ConvertToNonContainer()
}

// containerConversionStateKeeper is implemented by nodes whose conversion to or from a container changes more than
// their kind, such as equipment, whose kind of container and legality class change too, so that undoing and redoing
// a conversion can put that back as well.
type containerConversionStateKeeper interface {
	ContainerConversionState() any
	RestoreContainerConversionState(state any)
}

type containerConversionList struct {
	Owner Rebuildable
	List  []*containerConversion
	// ids holds the IDs of the converted rows, both before and after their conversion, since converting a row to or from
	// a container changes its ID's kind, and an editor may be open on either.
	ids map[tid.TID]bool
}

// Apply carries out the conversions. Any editor open on one of the rows is discarded first: it was opened on the other
// kind of row, so it shows the wrong fields, and applying it would undo the conversion again. Its pending changes are
// dropped rather than applied, since they were made to a state the row is no longer in.
func (c *containerConversionList) Apply() {
	discardEditorsFor(c.ids)
	for _, one := range c.List {
		one.Apply()
	}
	rebuildAsModified(c.Owner, true)
}

type containerConversion struct {
	Target      ConvertableContainer
	ToContainer bool
	// state, when hasState is true, is what the target is left holding once converted (see
	// containerConversionStateKeeper).
	state    any
	hasState bool
}

func newContainerConversion(target ConvertableContainer, toContainer bool) *containerConversion {
	return &containerConversion{
		Target:      target,
		ToContainer: toContainer,
	}
}

func (c *containerConversion) Apply() {
	if c.ToContainer {
		c.Target.ConvertToContainer()
	} else {
		c.Target.ConvertToNonContainer()
	}
	if c.hasState {
		if keeper, ok := c.Target.(containerConversionStateKeeper); ok {
			keeper.RestoreContainerConversionState(c.state)
		}
	}
}

// InstallContainerConversionHandlers installs the to & from container conversion handlers.
func InstallContainerConversionHandlers[T gurps.Node[T]](paneler unison.Paneler, owner Rebuildable, table *unison.Table[*Node[T]]) {
	var zero T
	if _, ok := any(zero).(ConvertableContainer); ok {
		p := paneler.AsPanel()
		p.InstallCmdHandlers(ConvertToContainerItemID,
			func(_ any) bool { return CanConvertToContainer(table) },
			func(_ any) { ConvertToContainer(owner, table) })
		p.InstallCmdHandlers(ConvertToNonContainerItemID,
			func(_ any) bool { return CanConvertToNonContainer(table) },
			func(_ any) { ConvertToNonContainer(owner, table) })
	}
}

// CanConvertToContainer returns true if the table's current selection has a row that can be converted to a container.
func CanConvertToContainer[T gurps.Node[T]](table *unison.Table[*Node[T]]) bool {
	return canConvertContainers(table, true)
}

// CanConvertToNonContainer returns true if the table's current selection has a row that can be converted to a
// non-container.
func CanConvertToNonContainer[T gurps.Node[T]](table *unison.Table[*Node[T]]) bool {
	return canConvertContainers(table, false)
}

// ConvertToContainer converts any selected rows to containers, if possible.
func ConvertToContainer[T gurps.Node[T]](owner Rebuildable, table *unison.Table[*Node[T]]) {
	convertContainers(owner, table, true)
}

// ConvertToNonContainer converts any selected rows to non-containers, if possible.
func ConvertToNonContainer[T gurps.Node[T]](owner Rebuildable, table *unison.Table[*Node[T]]) {
	convertContainers(owner, table, false)
}

// convertibleSelection returns the selected rows' data that can be converted in the given direction: to a container
// when toContainer is true, to a non-container otherwise.
func convertibleSelection[T gurps.Node[T]](table *unison.Table[*Node[T]], toContainer bool) []T {
	var list []T
	for _, row := range table.SelectedRows(false) {
		data := row.Data()
		if c, ok := any(data).(ConvertableContainer); ok && !xreflect.IsNil(data) &&
			c.CanConvertToFromContainer() && c.Container() != toContainer {
			list = append(list, data)
		}
	}
	return list
}

// canConvertContainers returns true if the table's current selection has a row that can be converted in the given
// direction.
func canConvertContainers[T gurps.Node[T]](table *unison.Table[*Node[T]], toContainer bool) bool {
	return len(convertibleSelection(table, toContainer)) > 0
}

// convertContainers converts any selected rows in the given direction, if possible, recording a single undo edit for
// the whole selection. Any editor open on one of the rows is closed first, since the fields it shows depend on whether
// the row is a container (see closeEditorsBeforeConversion).
func convertContainers[T gurps.Node[T]](owner Rebuildable, table *unison.Table[*Node[T]], toContainer bool) {
	targets := convertibleSelection(table, toContainer)
	if len(targets) == 0 {
		return
	}
	table, ok := closeEditorsBeforeConversion(table, targets)
	if !ok {
		return
	}
	before, after := convertContainersWithoutUndo(owner, table, toContainer)
	if before == nil {
		return
	}
	if mgr := unison.UndoManagerFor(table); mgr != nil {
		action := convertToNonContainerAction
		if toContainer {
			action = convertToContainerAction
		}
		mgr.Add(&unison.UndoEdit[*containerConversionList]{
			ID:         unison.NextUndoID(),
			EditName:   action.Title,
			UndoFunc:   func(edit *unison.UndoEdit[*containerConversionList]) { edit.BeforeData.Apply() },
			RedoFunc:   func(edit *unison.UndoEdit[*containerConversionList]) { edit.AfterData.Apply() },
			BeforeData: before,
			AfterData:  after,
		})
	}
	rebuildAsModified(owner, true)
}

// convertContainersWithoutUndo converts any selected rows in the given direction, if possible, returning the
// conversions that undo and redo it, or nils if nothing was converted. Recording the undo edit and rebuilding are left
// to the caller.
func convertContainersWithoutUndo[T gurps.Node[T]](owner Rebuildable, table *unison.Table[*Node[T]], toContainer bool) (before, after *containerConversionList) {
	targets := convertibleSelection(table, toContainer)
	if len(targets) == 0 {
		return nil, nil
	}
	ids := make(map[tid.TID]bool, len(targets))
	before = &containerConversionList{Owner: owner, ids: ids}
	after = &containerConversionList{Owner: owner, ids: ids}
	for _, data := range targets {
		ids[data.ID()] = true
		target := any(data).(ConvertableContainer) //nolint:errcheck // convertibleSelection checked this
		undo := newContainerConversion(target, !toContainer)
		redo := newContainerConversion(target, toContainer)
		keeper, keeps := target.(containerConversionStateKeeper)
		if keeps {
			undo.state, undo.hasState = keeper.ContainerConversionState(), true
		}
		redo.Apply()
		ids[data.ID()] = true
		if keeps {
			redo.state, redo.hasState = keeper.ContainerConversionState(), true
		}
		before.List = append(before.List, undo)
		after.List = append(after.List, redo)
	}
	return before, after
}
