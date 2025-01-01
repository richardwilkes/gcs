// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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

// CanOpenPageRef returns true if the current selection on the table has a page reference.
func CanOpenPageRef[T gurps.NodeTypes](table *unison.Table[*Node[T]]) bool {
	for _, row := range table.SelectedRows(false) {
		var data gurps.CellData
		gurps.AsNode(row.Data()).CellData(gurps.PageRefCellAlias, &data)
		if len(ExtractPageReferences(data.Primary)) != 0 {
			return true
		}
	}
	return false
}

// OpenPageRef opens the first page reference on each selected item in the table.
func OpenPageRef[T gurps.NodeTypes](table *unison.Table[*Node[T]]) {
	promptCtx := make(map[string]bool)
	for _, row := range table.SelectedRows(false) {
		var data gurps.CellData
		gurps.AsNode(row.Data()).CellData(gurps.PageRefCellAlias, &data)
		for _, one := range ExtractPageReferences(data.Primary) {
			OpenPageReference(one, data.Secondary, promptCtx)
			break
		}
	}
}

// OpenEachPageRef opens the all page references on each selected item in the table.
func OpenEachPageRef[T gurps.NodeTypes](table *unison.Table[*Node[T]]) {
	promptCtx := make(map[string]bool)
	for _, row := range table.SelectedRows(false) {
		var data gurps.CellData
		gurps.AsNode(row.Data()).CellData(gurps.PageRefCellAlias, &data)
		for _, one := range ExtractPageReferences(data.Primary) {
			if OpenPageReference(one, data.Secondary, promptCtx) {
				return
			}
		}
	}
}
