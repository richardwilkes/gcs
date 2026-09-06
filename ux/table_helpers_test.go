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
	"github.com/richardwilkes/unison"
)

// newProviderTable returns a table for the provider, with the provider told about it and its root rows loaded into
// it: the wiring NewNodeTable does, minus the header, columns and layout that a test driving the provider directly has
// no use for.
func newProviderTable[T gurps.Node[T]](provider TableProvider[T]) *unison.Table[*Node[T]] {
	table := unison.NewTable(provider)
	provider.SetTable(table)
	table.SetRootRows(provider.RootRows())
	return table
}

// newDragData wraps data as the drag data a table drop hands over: nodes belonging to a throw-away table of their own,
// standing in for the list they were dragged out of.
func newDragData[T gurps.Node[T]](data ...T) *unison.TableDragData[*Node[T]] {
	table := unison.NewTable(&unison.SimpleTableModel[*Node[T]]{})
	rows := make([]*Node[T], len(data))
	for i, one := range data {
		rows[i] = NewNode(table, nil, one, false)
	}
	return &unison.TableDragData[*Node[T]]{Table: table, Rows: rows}
}

// altDrop performs an alternate drop of dropped onto the rows at rowIndexes of the table support belongs to.
func altDrop[T gurps.Node[T]](support *AltDropSupport, rowIndexes []int, dropped ...T) {
	support.Drop(rowIndexes, newDragData(dropped...))
}
