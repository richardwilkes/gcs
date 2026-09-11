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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// The saved_calc.gcs fixture records, in its "calc" objects, values that recalculating the sheet would not produce:
// Magery is level 3 with no bonus anywhere, yet is recorded as level 5, and the note's script yields 4, yet its text
// is recorded as "ST 42". A load that keeps the saved calc can therefore be told apart from one that derives the
// values. The Dormant trait is disabled, so its recorded level is zero.

func TestNewEntityFromFileWithSavedCalc(t *testing.T) {
	c := check.New(t)
	e, err := NewEntityFromFileWithSavedCalc(os.DirFS("testdata"), "saved_calc.gcs")
	c.NoError(err)
	c.Equal(3, len(e.Traits))
	c.Equal(2, len(e.Notes))
	c.Equal("Magery 5", e.Traits[0].StringWithSavedCalc())
	c.Equal("Combat Reflexes", e.Traits[1].StringWithSavedCalc())
	// A disabled trait shows the level it would have if enabled, as String does, not the recorded zero.
	c.Equal("Dormant 2", e.Traits[2].StringWithSavedCalc())
	c.Equal("ST 42", e.Notes[0].StringWithSavedCalc())
	c.Equal("Plain note", e.Notes[1].StringWithSavedCalc())
	// Nothing was derived: the items were never even attached to the entity.
	c.Nil(e.Traits[0].DataOwner())
	c.Nil(e.Notes[0].DataOwner())
}

func TestNewEntityFromFileIgnoresSavedCalc(t *testing.T) {
	c := check.New(t)
	e, err := NewEntityFromFile(os.DirFS("testdata"), "saved_calc.gcs")
	c.NoError(err)
	c.Equal("Magery 3", e.Traits[0].String())
	c.Equal("Magery 3", e.Traits[0].StringWithSavedCalc())
	c.Equal("ST <script>2+2</script>", e.Notes[0].StringWithSavedCalc())
	c.NotNil(e.Traits[0].DataOwner())
}

func TestStringWithSavedCalcFallsBackWhenNothingUsableWasRecorded(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	data := fmt.Appendf(nil, `{
	"version": %d,
	"profile": {"name": "Plain"},
	"traits": [{"name": "Magery", "levels": 3, "can_level": true, "calc": "not an object"}],
	"notes": [{"markdown": "ST <script>2+2</script>"}]
}`, jio.CurrentDataVersion)
	c.NoError(os.WriteFile(filepath.Join(dir, "plain.gcs"), data, 0o600))
	e, err := NewEntityFromFileWithSavedCalc(os.DirFS(dir), "plain.gcs")
	c.NoError(err)
	c.Equal("Magery 3", e.Traits[0].StringWithSavedCalc())
	c.Equal("ST <script>2+2</script>", e.Notes[0].StringWithSavedCalc())
}

func TestNewNotesFromFileWithSavedCalc(t *testing.T) {
	c := check.New(t)
	notes, err := NewNotesFromFileWithSavedCalc(os.DirFS("testdata"), "saved_calc.not")
	c.NoError(err)
	c.Equal(2, len(notes))
	c.Equal("ST 42", notes[0].StringWithSavedCalc())
	c.Equal("Plain note", notes[1].StringWithSavedCalc())
	// The ordinary loader ignores the recorded text.
	notes, err = NewNotesFromFile(os.DirFS("testdata"), "saved_calc.not")
	c.NoError(err)
	c.Equal("ST <script>2+2</script>", notes[0].StringWithSavedCalc())
}

func TestNewTemplateFromFileWithSavedCalc(t *testing.T) {
	c := check.New(t)
	tmpl, err := NewTemplateFromFileWithSavedCalc(os.DirFS("testdata"), "saved_calc.gct")
	c.NoError(err)
	c.Equal(1, len(tmpl.Traits))
	c.Equal(2, len(tmpl.Notes))
	c.Equal("Magery 5", tmpl.Traits[0].StringWithSavedCalc())
	c.Equal("ST 42", tmpl.Notes[0].StringWithSavedCalc())
	c.Equal("Plain note", tmpl.Notes[1].StringWithSavedCalc())
	tmpl, err = NewTemplateFromFile(os.DirFS("testdata"), "saved_calc.gct")
	c.NoError(err)
	c.Equal("Magery 3", tmpl.Traits[0].StringWithSavedCalc())
	c.Equal("ST <script>2+2</script>", tmpl.Notes[0].StringWithSavedCalc())
}

func TestNewLootFromFileWithSavedCalc(t *testing.T) {
	c := check.New(t)
	loot, err := NewLootFromFileWithSavedCalc(os.DirFS("testdata"), "saved_calc.loot")
	c.NoError(err)
	c.Equal(2, len(loot.Notes))
	c.Equal("ST 42", loot.Notes[0].StringWithSavedCalc())
	c.Equal("Plain note", loot.Notes[1].StringWithSavedCalc())
	loot, err = NewLootFromFile(os.DirFS("testdata"), "saved_calc.loot")
	c.NoError(err)
	c.Equal("ST <script>2+2</script>", loot.Notes[0].StringWithSavedCalc())
}
