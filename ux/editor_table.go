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
	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/unison"
)

func newEditorTable[T model.NodeTypes](parent *unison.Panel, provider TableProvider[T]) *unison.Table[*Node[T]] {
	header, table := NewNodeTable[T](provider, unison.FieldFont)
	table.InstallCmdHandlers(OpenEditorItemID, func(_ any) bool { return table.HasSelection() },
		func(_ any) { provider.OpenEditor(unison.AncestorOrSelf[Rebuildable](table), table) })
	table.InstallCmdHandlers(OpenOnePageReferenceItemID,
		func(_ any) bool { return CanOpenPageRef(table) },
		func(_ any) { OpenPageRef(table) })
	table.InstallCmdHandlers(OpenEachPageReferenceItemID,
		func(_ any) bool { return CanOpenPageRef(table) },
		func(_ any) { OpenEachPageRef(table) })
	table.InstallCmdHandlers(unison.DeleteItemID,
		func(_ any) bool { return table.HasSelection() },
		func(_ any) { DeleteSelection(table) })
	table.InstallCmdHandlers(DuplicateItemID,
		func(_ any) bool { return table.HasSelection() },
		func(_ any) { DuplicateSelection(table) })
	InstallTableDropSupport(table, provider)
	table.SyncToModel()
	parent.AddChild(header)
	parent.AddChild(table)
	return table
}
