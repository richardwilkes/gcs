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

// ConvertableContainer defines the types that the container conversion can work on.
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
	c.Owner.Rebuild(true)
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
	p := paneler.AsPanel()
	p.InstallCmdHandlers(ConvertToContainerItemID,
		func(_ any) bool { return CanConvertToContainer(table) },
		func(_ any) { ConvertToContainer(owner, table) })
	p.InstallCmdHandlers(ConvertToNonContainerItemID,
		func(_ any) bool { return CanConvertToNonContainer(table) },
		func(_ any) { ConvertToNonContainer(owner, table) })
}

// CanConvertToContainer returns true if the table's current selection has a row that can be converted to a container.
func CanConvertToContainer[T gurps.Node[T]](table *unison.Table[*Node[T]]) bool {
	for _, row := range table.SelectedRows(false) {
		if data, ok := any(row.Data()).(ConvertableContainer); ok && !xreflect.IsNil(data) && data.CanConvertToFromContainer() && !data.Container() {
			return true
		}
	}
	return false
}

// CanConvertToNonContainer returns true if the table's current selection has a row that can be converted to a
// non-container.
func CanConvertToNonContainer[T gurps.Node[T]](table *unison.Table[*Node[T]]) bool {
	for _, row := range table.SelectedRows(false) {
		if data, ok := any(row.Data()).(ConvertableContainer); ok && !xreflect.IsNil(data) && data.CanConvertToFromContainer() && data.Container() {
			return true
		}
	}
	return false
}

// ConvertToContainer converts any selected rows to containers, if possible.
func ConvertToContainer[T gurps.Node[T]](owner Rebuildable, table *unison.Table[*Node[T]]) {
	before := &containerConversionList{Owner: owner}
	after := &containerConversionList{Owner: owner}
	for _, row := range table.SelectedRows(false) {
		if data, ok := any(row.Data()).(ConvertableContainer); ok && !xreflect.IsNil(data) && data.CanConvertToFromContainer() && !data.Container() {
			before.List = append(before.List, newContainerConversion(data, false))
			after.List = append(after.List, newContainerConversion(data, true))
			data.ConvertToContainer()
		}
	}
	if len(before.List) > 0 {
		if mgr := unison.UndoManagerFor(table); mgr != nil {
			mgr.Add(&unison.UndoEdit[*containerConversionList]{
				ID:         unison.NextUndoID(),
				EditName:   convertToContainerAction.Title,
				UndoFunc:   func(edit *unison.UndoEdit[*containerConversionList]) { edit.BeforeData.Apply() },
				RedoFunc:   func(edit *unison.UndoEdit[*containerConversionList]) { edit.AfterData.Apply() },
				BeforeData: before,
				AfterData:  after,
			})
		}
		owner.Rebuild(true)
	}
}

// ConvertToNonContainer converts any selected rows to non-containers, if possible.
func ConvertToNonContainer[T gurps.Node[T]](owner Rebuildable, table *unison.Table[*Node[T]]) {
	before := &containerConversionList{Owner: owner}
	after := &containerConversionList{Owner: owner}
	for _, row := range table.SelectedRows(false) {
		if data, ok := any(row.Data()).(ConvertableContainer); ok && !xreflect.IsNil(data) && data.CanConvertToFromContainer() && data.Container() {
			before.List = append(before.List, newContainerConversion(data, true))
			after.List = append(after.List, newContainerConversion(data, false))
			data.ConvertToNonContainer()
		}
	}
	if len(before.List) > 0 {
		if mgr := unison.UndoManagerFor(table); mgr != nil {
			mgr.Add(&unison.UndoEdit[*containerConversionList]{
				ID:         unison.NextUndoID(),
				EditName:   convertToNonContainerAction.Title,
				UndoFunc:   func(edit *unison.UndoEdit[*containerConversionList]) { edit.BeforeData.Apply() },
				RedoFunc:   func(edit *unison.UndoEdit[*containerConversionList]) { edit.AfterData.Apply() },
				BeforeData: before,
				AfterData:  after,
			})
		}
		owner.Rebuild(true)
	}
}
