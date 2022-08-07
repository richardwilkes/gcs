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

package editors

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ui/widget/ntable"
	"github.com/richardwilkes/gcs/v5/ui/workspace/settings"
	"github.com/richardwilkes/unison"
)

// CanOpenPageRef returns true if the current selection on the table has a page reference.
func CanOpenPageRef[T gurps.NodeTypes](table *unison.Table[*ntable.Node[T]]) bool {
	for _, row := range table.SelectedRows(false) {
		var data gurps.CellData
		gurps.AsNode(row.Data()).CellData(gurps.PageRefCellAlias, &data)
		if len(settings.ExtractPageReferences(data.Primary)) != 0 {
			return true
		}
	}
	return false
}

// OpenPageRef opens the first page reference on each selected item in the table.
func OpenPageRef[T gurps.NodeTypes](table *unison.Table[*ntable.Node[T]]) {
	promptCtx := make(map[string]bool)
	for _, row := range table.SelectedRows(false) {
		var data gurps.CellData
		gurps.AsNode(row.Data()).CellData(gurps.PageRefCellAlias, &data)
		for _, one := range settings.ExtractPageReferences(data.Primary) {
			if settings.OpenPageReference(table.Window(), one, data.Secondary, promptCtx) {
				return
			}
		}
	}
}

// OpenEachPageRef opens the all page references on each selected item in the table.
func OpenEachPageRef[T gurps.NodeTypes](table *unison.Table[*ntable.Node[T]]) {
	promptCtx := make(map[string]bool)
	for _, row := range table.SelectedRows(false) {
		var data gurps.CellData
		gurps.AsNode(row.Data()).CellData(gurps.PageRefCellAlias, &data)
		for _, one := range settings.ExtractPageReferences(data.Primary) {
			if settings.OpenPageReference(table.Window(), one, data.Secondary, promptCtx) {
				return
			}
		}
	}
}
