// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"encoding/json/v2"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// The saved_calc.gcs fixture records, in its "calc" objects, values that recalculating the sheet would not produce:
// Magery is level 3 with no bonus anywhere, yet is recorded as level 5, and the note's script yields 4, yet its text
// is recorded as "ST 42". A load that keeps the saved calc can therefore be told apart from one that derives the
// values. The Dormant trait is disabled, so its recorded level is zero. Unproven's prerequisites were unmet but not
// enforced, so it stayed in effect at its recorded level, while Barred and the Barred Group holding Sealed were
// disabled by the sheet for theirs, so their recorded level is zero.

// TestNewEntityFromFileWithSavedCalc verifies that the saved calc loader shows the values a sheet records without
// deriving anything, while the ordinary loader ignores them and derives its own.
func TestNewEntityFromFileWithSavedCalc(t *testing.T) {
	c := check.New(t)
	e, err := NewEntityFromFileWithSavedCalc(os.DirFS("testdata"), "saved_calc.gcs")
	c.NoError(err)
	c.Equal(6, len(e.Traits))
	c.Equal(2, len(e.Notes))
	c.Equal("Magery 5", e.Traits[0].StringWithSavedCalc())
	c.Equal("Combat Reflexes", e.Traits[1].StringWithSavedCalc())
	// A disabled trait shows the level it would have if enabled, as String does, not the recorded zero.
	c.Equal("Dormant 2", e.Traits[2].StringWithSavedCalc())
	c.Equal("Unproven 5", e.Traits[3].StringWithSavedCalc())
	c.Equal("Barred 2", e.Traits[4].StringWithSavedCalc())
	c.Equal("Sealed 4", e.Traits[5].Children[0].StringWithSavedCalc())
	c.Equal("ST 42", e.Notes[0].StringWithSavedCalc())
	c.Equal("Plain note", e.Notes[1].StringWithSavedCalc())
	// Nothing was derived: the items were never even attached to the entity.
	c.Nil(e.Traits[0].DataOwner())
	c.Nil(e.Notes[0].DataOwner())

	// The ordinary loader ignores the recorded values.
	e, err = NewEntityFromFile(os.DirFS("testdata"), "saved_calc.gcs")
	c.NoError(err)
	c.Equal("Magery 3", e.Traits[0].String())
	c.Equal("Magery 3", e.Traits[0].StringWithSavedCalc())
	c.Equal("ST <script>2+2</script>", e.Notes[0].StringWithSavedCalc())
	c.NotNil(e.Traits[0].DataOwner())
}

// TestStringWithSavedCalcMatchesStringAtSave verifies that a trait loaded with its saved calc reads as String read it
// when the sheet was saved, whether its recorded level is zero because the trait is at level zero, or because it was
// disabled or not in effect, in which case String shows the level it would have. A granted trait shows the level the
// sheet recorded, as a bought trait does, since whether its grant was in effect is known only from that record.
func TestStringWithSavedCalcMatchesStringAtSave(t *testing.T) {
	for _, tc := range []struct {
		name string
		// atSave is what String reads for the trait when the sheet is saved.
		atSave string
		// add puts the trait on the sheet and returns it, and find returns the same trait in the sheet loaded from the
		// file, or, when nil, it is the last trait at the top of the list.
		add  func(e *Entity) *Trait
		find func(loaded *Entity) *Trait
	}{
		{name: "enabled at level zero", atSave: "Acute Vision 0", add: func(e *Entity) *Trait {
			bonus := NewTraitBonus()
			bonus.NameCriteria.Qualifier = "Acute Vision"
			bonus.Amount = -fxp.Two
			addTraitWithFeatures(e, "Dulled", bonus)
			vision := addTrait(e, newGrantLeveledTrait("Acute Vision", fxp.Two))
			vision.Levels = fxp.Two
			return vision
		}},
		{name: "granted, with a bonus", atSave: "Night Vision 3", add: func(e *Entity) *Trait {
			bonus := NewTraitBonus()
			bonus.NameCriteria.Qualifier = "Night Vision"
			addTraitWithFeatures(e, "Boost", bonus)
			addTraitWithFeatures(e, "Racial",
				newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.Two, grantmerge.Ignore))
			e.Recalculate()
			return grantedOf(e.Traits)[0]
		}},
		{name: "granted but not in effect", atSave: "Night Vision 2", add: func(e *Entity) *Trait {
			goggles := addCarriedEquipmentWithFeatures(e, "Goggles",
				newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.Two, grantmerge.Ignore))
			goggles.Equipped = false
			e.Recalculate()
			return grantedOf(e.Traits)[0]
		}},
		{name: "disabled by unmet prerequisites", atSave: "Magery 3", add: func(e *Entity) *Trait {
			e.SheetSettings.EnforceTraitPrereqs = true
			magery := addTrait(e, newGrantLeveledTrait("Magery", fxp.Ten))
			magery.Levels = fxp.Three
			magery.Prereq = newPrereqListForTrait("Magic Aptitude", true)
			return magery
		}},
		{name: "unmet prerequisites not enforced, with a bonus", atSave: "Magery 4", add: func(e *Entity) *Trait {
			bonus := NewTraitBonus()
			bonus.NameCriteria.Qualifier = "Magery"
			addTraitWithFeatures(e, "Boost", bonus)
			magery := addTrait(e, newGrantLeveledTrait("Magery", fxp.Ten))
			magery.Levels = fxp.Three
			magery.Prereq = newPrereqListForTrait("Magic Aptitude", true)
			return magery
		}},
		{name: "disabled by a bonus", atSave: "Magery 4", add: func(e *Entity) *Trait {
			bonus := NewTraitBonus()
			bonus.NameCriteria.Qualifier = "Magery"
			addTraitWithFeatures(e, "Boost", bonus)
			magery := addTrait(e, newGrantLeveledTrait("Magery", fxp.Ten))
			magery.Levels = fxp.Three
			magery.SetDisabled(true)
			return magery
		}},
		{name: "in a container disabled by unmet prerequisites", atSave: "Magery 3", add: func(e *Entity) *Trait {
			e.SheetSettings.EnforceTraitPrereqs = true
			group := NewTrait(e, nil, true)
			group.Name = "Group"
			group.Prereq = newPrereqListForTrait("Magic Aptitude", true)
			magery := newGrantLeveledTrait("Magery", fxp.Ten)
			magery.Levels = fxp.Three
			magery.SetParent(group)
			group.Children = []*Trait{magery}
			addTrait(e, group)
			return magery
		}, find: func(loaded *Entity) *Trait { return loaded.Traits[len(loaded.Traits)-1].Children[0] }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			trait := tc.add(e)
			e.Recalculate()
			data, err := json.Marshal(e)
			c.NoError(err)
			c.Equal(tc.atSave, trait.String(), "precondition: what String reads when the sheet is saved")
			dir := t.TempDir()
			c.NoError(os.WriteFile(filepath.Join(dir, "sheet.gcs"), data, 0o600))
			loaded, err := NewEntityFromFileWithSavedCalc(os.DirFS(dir), "sheet.gcs")
			c.NoError(err)
			found := loaded.Traits[len(loaded.Traits)-1]
			if tc.find != nil {
				found = tc.find(loaded)
			}
			c.Equal(trait.String(), found.StringWithSavedCalc())
		})
	}
}

// TestStringWithSavedCalcFallback verifies when a trait or note loaded with its saved calc falls back to its own
// unadjusted level or text: for a calc that fails to decode, even if part of it would, a missing one, or a granted
// trait's zero current level from a file that predates the recorded level, but not for a bought trait's recorded zero.
func TestStringWithSavedCalcFallback(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	data := fmt.Appendf(nil, `{
	"version": %d,
	"profile": {"name": "Plain"},
	"traits": [
		{"name": "Magery", "levels": 3, "can_level": true, "calc": {"level": 5, "unsatisfied_reason": 7}},
		{"name": "Night Vision", "levels": 2, "can_level": true, "granted_by": {"grantor": %q, "key": %q, "hash": "x"},
			"calc": {"current_level": 0}},
		{"name": "Acute Vision", "levels": 2, "can_level": true, "calc": {"current_level": 0}}
	],
	"notes": [{"markdown": "ST <script>2+2</script>"}]
}`, jio.CurrentDataVersion, tid.MustNewTID(kinds.Trait), tid.MustNewTID(kinds.Trait))
	c.NoError(os.WriteFile(filepath.Join(dir, "plain.gcs"), data, 0o600))
	e, err := NewEntityFromFileWithSavedCalc(os.DirFS(dir), "plain.gcs")
	c.NoError(err)
	c.Equal("Magery 3", e.Traits[0].StringWithSavedCalc(), "no part of a calc that fails to decode is used")
	c.Equal("Night Vision 2", e.Traits[1].StringWithSavedCalc(), "a granted trait's grant may not have been in effect")
	c.Equal("Acute Vision 0", e.Traits[2].StringWithSavedCalc(), "a bought trait's recorded zero is its level")
	c.Equal("ST <script>2+2</script>", e.Notes[0].StringWithSavedCalc())
}

// TestNewNotesFromFileWithSavedCalc verifies that the note list loader keeping the saved calc shows the text the file
// recorded, while the ordinary one does not.
func TestNewNotesFromFileWithSavedCalc(t *testing.T) {
	c := check.New(t)
	notes, err := NewNotesFromFileWithSavedCalc(os.DirFS("testdata"), "saved_calc.not")
	c.NoError(err)
	c.Equal(2, len(notes))
	c.Equal("ST 42", notes[0].StringWithSavedCalc())
	// The ordinary loader ignores the recorded text.
	notes, err = NewNotesFromFile(os.DirFS("testdata"), "saved_calc.not")
	c.NoError(err)
	c.Equal("ST <script>2+2</script>", notes[0].StringWithSavedCalc())
}

// TestNewTemplateFromFileWithSavedCalc verifies that the template loader keeping the saved calc shows the levels and
// text the file recorded, while the ordinary one does not.
func TestNewTemplateFromFileWithSavedCalc(t *testing.T) {
	c := check.New(t)
	tmpl, err := NewTemplateFromFileWithSavedCalc(os.DirFS("testdata"), "saved_calc.gct")
	c.NoError(err)
	c.Equal(1, len(tmpl.Traits))
	c.Equal(2, len(tmpl.Notes))
	c.Equal("Magery 5", tmpl.Traits[0].StringWithSavedCalc())
	c.Equal("ST 42", tmpl.Notes[0].StringWithSavedCalc())
	tmpl, err = NewTemplateFromFile(os.DirFS("testdata"), "saved_calc.gct")
	c.NoError(err)
	c.Equal("Magery 3", tmpl.Traits[0].StringWithSavedCalc())
	c.Equal("ST <script>2+2</script>", tmpl.Notes[0].StringWithSavedCalc())
}

// TestNewLootFromFileWithSavedCalc verifies that the loot sheet loader keeping the saved calc shows the text the file
// recorded, while the ordinary one does not.
func TestNewLootFromFileWithSavedCalc(t *testing.T) {
	c := check.New(t)
	loot, err := NewLootFromFileWithSavedCalc(os.DirFS("testdata"), "saved_calc.loot")
	c.NoError(err)
	c.Equal(2, len(loot.Notes))
	c.Equal("ST 42", loot.Notes[0].StringWithSavedCalc())
	loot, err = NewLootFromFile(os.DirFS("testdata"), "saved_calc.loot")
	c.NoError(err)
	c.Equal("ST <script>2+2</script>", loot.Notes[0].StringWithSavedCalc())
}
