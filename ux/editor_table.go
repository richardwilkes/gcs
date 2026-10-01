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
	installStandardTableCmdHandlers(parent, table, provider,
		func() Rebuildable { return table.AncestorOrSelf[Rebuildable]() }, true)
	// Toggle State is installed here rather than in NewNodeTable because only the editor tables carry the checkmark
	// columns it flips: modifier tables show the enabled column only when built for an editor, and weapon tables show
	// the Hide column only when not built for a page. The owner is resolved inside the execute closure, since this
	// table's parent is not yet attached to the editor that owns it.
	switch t := any(table).(type) {
	case *unison.Table[*Node[*gurps.TraitModifier]]:
		installToggleModifierEnabledHandler(t)
		installModifierChoiceConversionHandlers[*gurps.TraitModifier, *gurps.TraitModifierEditData](t, t)
	case *unison.Table[*Node[*gurps.EquipmentModifier]]:
		installToggleModifierEnabledHandler(t)
		installModifierChoiceConversionHandlers[*gurps.EquipmentModifier, *gurps.EquipmentModifierEditData](t, t)
	case *unison.Table[*Node[*gurps.Weapon]]:
		installToggleHiddenHandler(t)
	}
	InstallTableDropSupport(table, provider)
	table.SyncToModel()
	parent.AddChild(header)
	parent.AddChild(table)
	return table
}
