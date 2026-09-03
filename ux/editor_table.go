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

func newEditorTable[T gurps.Node[T]](parent *unison.Panel, provider TableProvider[T]) *unison.Table[*Node[T]] {
	header, table := NewNodeTable(provider, unison.FieldFont)
	table.InstallCmdHandlers(OpenEditorItemID, func(_ any) bool { return table.HasSelection() },
		func(_ any) { provider.OpenEditor(table.AncestorOrSelf[Rebuildable](), table) })
	table.InstallCmdHandlers(OpenOnePageReferenceItemID,
		func(_ any) bool { return CanOpenPageRef(table) },
		func(_ any) { OpenPageRef(table) })
	table.InstallCmdHandlers(OpenEachPageReferenceItemID,
		func(_ any) bool { return CanOpenPageRef(table) },
		func(_ any) { OpenEachPageRef(table) })
	table.InstallCmdHandlers(unison.DeleteItemID,
		func(_ any) bool { return HasSelectionAndNotFiltered(table) },
		func(_ any) { DeleteSelection(table, true) })
	table.InstallCmdHandlers(DuplicateItemID,
		func(_ any) bool { return HasSelectionAndNotFiltered(table) },
		func(_ any) { DuplicateSelection(table) })
	table.InstallCmdHandlers(SyncWithSourceItemID,
		func(_ any) bool { return HasSelectionAndNotFiltered(table) },
		func(_ any) { SyncWithSourceForSelection(table) })
	table.InstallCmdHandlers(ClearSourceItemID,
		func(_ any) bool { return HasSelectionAndNotFiltered(table) },
		func(_ any) { ClearSourceFromSelection(table) })
	// Toggle State belongs here rather than in NewNodeTable because the editor tables are the only ones that carry the
	// checkmark columns it flips: the modifier tables show the enabled column only when built for an editor, and the
	// weapon tables show the Hide column only when they aren't built for a page. Installing it any higher would offer
	// the command on the library modifier lists and the sheet's own weapon lists, where there is nothing to toggle. The
	// owner is resolved inside the execute closure rather than here, since the panel this table is being added to has
	// not yet been attached to the editor that owns it.
	switch t := any(table).(type) {
	case *unison.Table[*Node[*gurps.TraitModifier]]:
		installToggleModifierEnabledHandler(t)
	case *unison.Table[*Node[*gurps.EquipmentModifier]]:
		installToggleModifierEnabledHandler(t)
	case *unison.Table[*Node[*gurps.Weapon]]:
		installToggleHiddenHandler(t)
	}
	InstallTableDropSupport(table, provider)
	table.SyncToModel()
	parent.AddChild(header)
	parent.AddChild(table)
	return table
}
