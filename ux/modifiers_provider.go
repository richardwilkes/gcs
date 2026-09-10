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
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/unison"
)

// modifierProviderSpec captures what differs between the trait modifier and equipment modifier table providers. Both
// tables show the same shape of list: an enabled column and a library source column that only an editor's table has,
// around a fixed run of columns led by the description.
type modifierProviderSpec[T gurps.Node[T]] struct {
	refKey            string
	dragKey           *uti.DataType
	dragSVG           *unison.SVG
	singular          string
	plural            string
	enabledColumn     int
	descriptionColumn int
	columns           []int // The fixed run of columns, description first.
	libSrcColumn      int
	newItem           func(owner gurps.DataOwner, parent T, container bool) T
	edit              func(owner Rebuildable, item T)
	menuActions       []*unison.Action
	filterKey         string
	filterFields      func() []*gurps.FilterField[T]
}

type modifiersProvider[T gurps.Node[T]] struct {
	listProvider[T]
	spec      modifierProviderSpec[T]
	forEditor bool
}

func newModifiersProvider[T gurps.Node[T]](owner gurps.DataOwnerProvider, list func() []T, setList func([]T), headerData func(columnID int) gurps.HeaderData, forEditor bool, spec modifierProviderSpec[T]) *modifiersProvider[T] {
	p := &modifiersProvider[T]{spec: spec, forEditor: forEditor}
	p.listProvider = listProvider[T]{
		dataOwner:    owner,
		list:         list,
		setList:      setList,
		columnIDs:    p.ColumnIDs,
		headerData:   headerData,
		newItem:      spec.newItem,
		edit:         spec.edit,
		filterKey:    spec.filterKey,
		filterFields: spec.filterFields,
	}
	return p
}

func (p *modifiersProvider[T]) RefKey() string {
	return p.spec.refKey
}

func (p *modifiersProvider[T]) DragKey() *uti.DataType {
	return p.spec.dragKey
}

func (p *modifiersProvider[T]) DragSVG() *unison.SVG {
	return p.spec.dragSVG
}

func (p *modifiersProvider[T]) ItemNames() (singular, plural string) {
	return p.spec.singular, p.spec.plural
}

func (p *modifiersProvider[T]) ColumnIDs() []int {
	columnIDs := make([]int, 0, len(p.spec.columns)+2)
	if p.forEditor {
		columnIDs = append(columnIDs, p.spec.enabledColumn)
	}
	columnIDs = append(columnIDs, p.spec.columns...)
	if p.forEditor {
		columnIDs = append(columnIDs, p.spec.libSrcColumn)
	}
	return columnIDs
}

func (p *modifiersProvider[T]) HierarchyColumnID() int {
	return p.spec.descriptionColumn
}

func (p *modifiersProvider[T]) ExcessWidthColumnID() int {
	return p.spec.descriptionColumn
}

func (p *modifiersProvider[T]) ContextMenuItems() []ContextMenuItem {
	items := make([]ContextMenuItem, 0, len(p.spec.menuActions))
	for _, action := range p.spec.menuActions {
		items = append(items, contextMenuItemFor(action))
	}
	return AppendDefaultContextMenuItems(items)
}
