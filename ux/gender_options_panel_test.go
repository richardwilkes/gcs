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

// TestAncestryEditorPanelRemoveGender verifies that a gender can be removed from the middle of the list, that the last
// one may be removed too, and that removals are undoable.
func TestAncestryEditorPanelRemoveGender(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	a.GenderOptions = append(a.GenderOptions, &gurps.WeightedAncestryOptions{
		Weight: 1,
		Value:  &gurps.AncestryOptions{Name: "Other"},
	})
	d := newTestAncestryEditorDockable(a)
	c.Equal([]string{"Male", "Female", "Other"}, genderNames(d.model))

	genderRows(d)[1].removeGender()
	c.Equal([]string{"Male", "Other"}, genderNames(d.model), "the middle gender is removed")
	c.Equal(2, len(genderRows(d)))

	genderRows(d)[0].removeGender()
	genderRows(d)[0].removeGender()
	c.Equal(0, len(d.model.GenderOptions), "the last gender may be removed")
	c.Equal(0, len(genderRows(d)))
	c.NotNil(rootAncestryPanel(t, d).genders, "the empty gender container remains")

	d.undoMgr.Undo()
	c.Equal([]string{"Other"}, genderNames(d.model), "undo restores the last removal")
	d.undoMgr.Undo()
	d.undoMgr.Undo()
	c.Equal([]string{"Male", "Female", "Other"}, genderNames(d.model), "each removal is its own undo")
	c.False(d.undoMgr.CanUndo())
}
