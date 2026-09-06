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

type containerConversionList struct {
	Owner Rebuildable
	List  []*containerConversion
}

func (c *containerConversionList) Apply() {
	for _, one := range c.List {
		one.Apply()
	}
	rebuildAsModified(c.Owner, true)
}

type containerConversion struct {
	Target      ConvertableContainer
	ToContainer bool
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
func convertibleSelection[T gurps.Node[T]](table *unison.Table[*Node[T]], toContainer bool) []ConvertableContainer {
	var list []ConvertableContainer
	for _, row := range table.SelectedRows(false) {
		if data, ok := any(row.Data()).(ConvertableContainer); ok && !xreflect.IsNil(data) &&
			data.CanConvertToFromContainer() && data.Container() != toContainer {
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
// the whole selection.
func convertContainers[T gurps.Node[T]](owner Rebuildable, table *unison.Table[*Node[T]], toContainer bool) {
	targets := convertibleSelection(table, toContainer)
	if len(targets) == 0 {
		return
	}
	before := &containerConversionList{Owner: owner}
	after := &containerConversionList{Owner: owner}
	for _, data := range targets {
		conv := newContainerConversion(data, toContainer)
		before.List = append(before.List, newContainerConversion(data, !toContainer))
		after.List = append(after.List, conv)
		conv.Apply()
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
