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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
)

// TestEditorListPanelRowGrowsWithItem checks that a row of an editor's modifier table grows, and the table with it,
// when the modifier's own editor gives it a name that wraps.
func TestEditorListPanelRowGrowsWithItem(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Claws"
	modifier := gurps.NewTraitModifier(entity, nil, false)
	modifier.Name = "Short"
	trait.Modifiers = []*gurps.TraitModifier{modifier}
	var e *editor[*gurps.Trait, *gurps.TraitEditData]
	var panel *traitModifiersPanel
	screen.Do(func() {
		entity.Traits = append(entity.Traits, trait)
		sheet.Rebuild(true)
		e = EditTrait(sheet, trait)
		panel, ok = uxtest.FirstPanelOfType[*traitModifiersPanel](e.AsPanel())
	})
	c.True(ok, "expected a trait modifiers panel in the trait editor")
	table := panel.table
	var before, after, frameHeight float32
	var pref geom.Size
	screen.Do(func() {
		before = table.RowHeights()[0]
		// What applying the modifier's editor does: change the editor's copy, then rebuild the editor.
		e.editorData.Modifiers[0].Name = strings.Repeat("A modifier name long enough to wrap ", 8)
		rebuildAsModified(e, true)
	})
	screen.Do(func() {
		after = table.RowHeights()[0]
		frameHeight = table.FrameRect().Height
		_, pref, _ = table.Sizes(geom.Size{})
	})
	c.True(after > before, "the row must grow to hold the wrapped name")
	c.Equal(pref.Height, frameHeight, "the table must grow to hold the taller row")
}
