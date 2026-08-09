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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// newEditorForTrait returns an editor holding the two copies of a trait's edit data that displayEditor would give it,
// without building any of the editor's UI. The trait belongs to an entity and carries weapons and a script-bearing
// modifier, so the derived values both copies would otherwise publish come from running the scripts in the data.
func newEditorForTrait() *editor[*gurps.Trait, *gurps.TraitEditData] {
	entity := gurps.NewEntity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Claws"
	trait.LocalNotes = "Sharp. <script>1 + 1</script>"
	modifier := gurps.NewTraitModifier(entity, nil, false)
	modifier.Name = "Long"
	modifier.LocalNotes = "Longer. <script>2 + 2</script>"
	trait.Modifiers = []*gurps.TraitModifier{modifier}
	trait.Weapons = []*gurps.Weapon{gurps.NewWeapon(trait, true), gurps.NewWeapon(trait, false)}
	entity.Traits = append(entity.Traits, trait)
	entity.Recalculate()

	e := &editor[*gurps.Trait, *gurps.TraitEditData]{target: trait}
	e.beforeData = &gurps.TraitEditData{}
	e.beforeData.CopyFrom(trait)
	e.editorData = &gurps.TraitEditData{}
	e.editorData.CopyFrom(trait)
	return e
}

// TestEditorIsModifiedFollowsTheDataAlone verifies that an editor reports unsaved changes when, and only when, its data
// has actually been edited. The two copies it compares are separate clones with separate script caches, and the derived
// values used to be part of the comparison: a script that exceeded the permitted per-script execution time while one
// copy was being marshaled but not the other made an untouched editor enable Apply and Cancel and prompt to save on the
// way out, purely because the machine had been busy for a moment.
func TestEditorIsModifiedFollowsTheDataAlone(t *testing.T) {
	c := check.New(t)
	e := newEditorForTrait()
	c.False(e.isModified(), "an editor whose two copies of the data agree must report no changes")

	e.editorData.Name = "Fangs"
	c.True(e.isModified(), "editing a field must be reported as a change")

	e.editorData.Name = "Claws"
	c.False(e.isModified(), "putting the original value back must clear the change")
}
