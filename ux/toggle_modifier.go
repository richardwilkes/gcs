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
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

// modifierExtractor yields a selected row's modifier when its enablement can be toggled. Containers are left out: a
// container is always enabled and shows no checkmark cell, so leaving them out is what disables the menu item for a
// container-only selection instead of letting it register an edit that changes nothing.
func modifierExtractor[T gurps.Node[T]](node T) (gurps.GeneralModifier, bool) {
	if xreflect.IsNil(node) {
		// Container() would dereference the node, so the nil check has to come first.
		return nil, false
	}
	m, ok := any(node).(gurps.GeneralModifier)
	return m, ok && !m.Container()
}

// modifierToggleUndoTitle returns the undo title to use for a table of the given kind of modifier. The Edit menu shows
// this title next to Undo and Redo, so it says which kind of modifier was toggled rather than just "modifier", which is
// what the separate per-kind code that preceded this did.
func modifierToggleUndoTitle[T gurps.Node[T]]() string {
	var zero T
	switch any(zero).(type) {
	case *gurps.TraitModifier:
		return i18n.Text("Toggle Trait Modifier")
	case *gurps.EquipmentModifier:
		return i18n.Text("Toggle Equipment Modifier")
	default:
		return i18n.Text("Toggle Modifier")
	}
}

func canToggleModifierEnabled[T gurps.Node[T]](table *unison.Table[*Node[T]]) bool {
	return canAdjustSelection(table, modifierExtractor[T])
}

// toggleModifierEnabled flips the enabled state of each selected modifier. Unlike a trait or a piece of equipment,
// turning a modifier on or off changes only what its owner is worth and what it grants, never which lists the owner
// shows, so the owner is merely marked as modified rather than rebuilt.
func toggleModifierEnabled[T gurps.Node[T]](owner Rebuildable, table *unison.Table[*Node[T]]) {
	adjustSelection(modifierToggleUndoTitle[T](), owner, table, modifierExtractor[T], gurps.GeneralModifier.Enabled,
		gurps.GeneralModifier.SetEnabled,
		func(m gurps.GeneralModifier) { m.SetEnabled(!m.Enabled()) },
		true, false)
}

// adjustModifierEnabled sets the enabled state of a single modifier. This is the form the checkmark cell uses, so that
// a click and the command register the same kind of undoable edit and report the change the same way.
func adjustModifierEnabled[T gurps.Node[T]](owner Rebuildable, undoSource unison.Paneler, node T, on bool) {
	m, ok := modifierExtractor(node)
	if !ok {
		return
	}
	adjustTargets(modifierToggleUndoTitle[T](), owner, undoSource, gurps.EntityFromNode(node),
		[]gurps.GeneralModifier{m}, gurps.GeneralModifier.Enabled, gurps.GeneralModifier.SetEnabled,
		func(m gurps.GeneralModifier) { m.SetEnabled(on) },
		false)
}

func installToggleModifierEnabledHandler[T gurps.Node[T]](table *unison.Table[*Node[T]]) {
	table.InstallCmdHandlers(ToggleStateItemID,
		func(_ any) bool { return canToggleModifierEnabled(table) },
		func(_ any) { toggleModifierEnabled(table.AncestorOrSelf[Rebuildable](), table) })
}
