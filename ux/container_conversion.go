/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// ConvertableNodeTypes defines the types that the container conversion can work on.
type ConvertableNodeTypes interface {
	*gurps.Equipment | *gurps.Note
	Container() bool
	GetType() string
	SetType(string)
	HasChildren() bool
}

type containerConversionList[T ConvertableNodeTypes] struct {
	Owner Rebuildable
	List  []*containerConversion[T]
}

func (c *containerConversionList[T]) Apply() {
	for _, one := range c.List {
		one.Apply()
	}
	c.Owner.Rebuild(true)
}

type containerConversion[T ConvertableNodeTypes] struct {
	Target T
	Type   string
}

func newContainerConversion[T ConvertableNodeTypes](target T) *containerConversion[T] {
	return &containerConversion[T]{
		Target: target,
		Type:   target.GetType(),
	}
}

func (c *containerConversion[T]) Apply() {
	c.Target.SetType(c.Type)
}

// InstallContainerConversionHandlers installs the to & from container conversion handlers.
func InstallContainerConversionHandlers[T ConvertableNodeTypes](paneler unison.Paneler, owner Rebuildable, table *unison.Table[*Node[T]]) {
	p := paneler.AsPanel()
	p.InstallCmdHandlers(ConvertToContainerItemID,
		func(_ any) bool { return CanConvertToContainer(table) },
		func(_ any) { ConvertToContainer(owner, table) })
	p.InstallCmdHandlers(ConvertToNonContainerItemID,
		func(_ any) bool { return CanConvertToNonContainer(table) },
		func(_ any) { ConvertToNonContainer(owner, table) })
}

// CanConvertToContainer returns true if the table's current selection has a row that can be converted to a container.
func CanConvertToContainer[T ConvertableNodeTypes](table *unison.Table[*Node[T]]) bool {
	for _, row := range table.SelectedRows(false) {
		if data := row.Data(); data != nil && !data.Container() {
			return true
		}
	}
	return false
}

// CanConvertToNonContainer returns true if the table's current selection has a row that can be converted to a
// non-container.
func CanConvertToNonContainer[T ConvertableNodeTypes](table *unison.Table[*Node[T]]) bool {
	for _, row := range table.SelectedRows(false) {
		if data := row.Data(); data != nil && data.Container() && !data.HasChildren() {
			return true
		}
	}
	return false
}

// ConvertToContainer converts any selected rows to containers, if possible.
func ConvertToContainer[T ConvertableNodeTypes](owner Rebuildable, table *unison.Table[*Node[T]]) {
	before := &containerConversionList[T]{Owner: owner}
	after := &containerConversionList[T]{Owner: owner}
	for _, row := range table.SelectedRows(false) {
		if data := row.Data(); data != nil && !data.Container() {
			before.List = append(before.List, newContainerConversion(data))
			data.SetType(data.GetType() + gurps.ContainerKeyPostfix)
			after.List = append(after.List, newContainerConversion(data))
		}
	}
	if len(before.List) > 0 {
		if mgr := unison.UndoManagerFor(table); mgr != nil {
			mgr.Add(&unison.UndoEdit[*containerConversionList[T]]{
				ID:         unison.NextUndoID(),
				EditName:   i18n.Text("Convert to Container"),
				UndoFunc:   func(edit *unison.UndoEdit[*containerConversionList[T]]) { edit.BeforeData.Apply() },
				RedoFunc:   func(edit *unison.UndoEdit[*containerConversionList[T]]) { edit.AfterData.Apply() },
				BeforeData: before,
				AfterData:  after,
			})
		}
		owner.Rebuild(true)
	}
}

// ConvertToNonContainer converts any selected rows to non-containers, if possible.
func ConvertToNonContainer[T ConvertableNodeTypes](owner Rebuildable, table *unison.Table[*Node[T]]) {
	before := &containerConversionList[T]{Owner: owner}
	after := &containerConversionList[T]{Owner: owner}
	for _, row := range table.SelectedRows(false) {
		if data := row.Data(); data != nil && data.Container() && !data.HasChildren() {
			before.List = append(before.List, newContainerConversion(data))
			data.SetType(strings.TrimSuffix(data.GetType(), gurps.ContainerKeyPostfix))
			after.List = append(after.List, newContainerConversion(data))
		}
	}
	if len(before.List) > 0 {
		if mgr := unison.UndoManagerFor(table); mgr != nil {
			mgr.Add(&unison.UndoEdit[*containerConversionList[T]]{
				ID:         unison.NextUndoID(),
				EditName:   i18n.Text("Convert to Non-Container"),
				UndoFunc:   func(edit *unison.UndoEdit[*containerConversionList[T]]) { edit.BeforeData.Apply() },
				RedoFunc:   func(edit *unison.UndoEdit[*containerConversionList[T]]) { edit.AfterData.Apply() },
				BeforeData: before,
				AfterData:  after,
			})
		}
		owner.Rebuild(true)
	}
}
