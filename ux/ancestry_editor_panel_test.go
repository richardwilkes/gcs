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

func rootAncestryPanel(t *testing.T, d *ancestryEditorDockable) *ancestryEditorPanel {
	t.Helper()
	roots := panelsOfType[*ancestryEditorPanel](d.AsPanel())
	if len(roots) != 1 {
		t.Fatalf("expected exactly one root panel, found %d", len(roots))
	}
	return roots[0]
}

func genderNames(a *gurps.Ancestry) []string {
	names := make([]string, 0, len(a.GenderOptions))
	for _, g := range a.GenderOptions {
		names = append(names, g.Value.Name)
	}
	return names
}

// TestAncestryEditorPanelRowsMirrorGenders verifies that there is one gender row per gender, in order, and that the
// fields that belong to a gender -- its weight and name -- appear only in gender rows, while the fields every options
// block has appear in both.
func TestAncestryEditorPanelRowsMirrorGenders(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	d := newTestAncestryEditorDockable(a)
	rows := genderRows(d)
	c.Equal(len(a.GenderOptions), len(rows), "one row per gender")
	for i, row := range rows {
		c.True(row.gender == a.GenderOptions[i], "row %d shows gender %d", i, i)
		c.NotNil(d.targetMgr.Find(a.GenderOptions[i].KeyPrefix+"weight"), "gender %d has a weight field", i)
		c.NotNil(d.targetMgr.Find(a.GenderOptions[i].Value.KeyPrefix+"name"), "gender %d has a name field", i)
		c.NotNil(d.targetMgr.Find(a.GenderOptions[i].Value.KeyPrefix+"age_script"), "gender %d has script fields", i)
	}
	common := a.CommonOptions
	c.Nil(d.targetMgr.Find(common.KeyPrefix+"weight"), "the common options have no weight")
	c.Nil(d.targetMgr.Find(common.KeyPrefix+"name"), "the common options have no name")
	c.NotNil(d.targetMgr.Find(common.KeyPrefix+"height_script"), "the common options have script fields")
	c.NotNil(d.targetMgr.Find(a.KeyPrefix+"name"), "the ancestry has a name field")
	c.NotNil(rootAncestryPanel(t, d).genders, "the gender container exists even when there are no genders")
}

// TestAncestryEditorPanelFieldsBind verifies that the gender and script fields write to the model.
func TestAncestryEditorPanelFieldsBind(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	d := newTestAncestryEditorDockable(a)
	g := a.GenderOptions[0]
	integerFieldFor(t, d, g.KeyPrefix+"weight").SetText("7")
	c.Equal(7, g.Weight)
	stringFieldFor(t, d, g.Value.KeyPrefix+"name").SetText("Neuter")
	c.Equal("Neuter", g.Value.Name)
	stringFieldFor(t, d, a.CommonOptions.KeyPrefix+"height_script").SetText("entity.randomHeightInInches($st)")
	c.Equal("entity.randomHeightInInches($st)", a.CommonOptions.HeightScript)
	stringFieldFor(t, d, a.CommonOptions.KeyPrefix+"weight_script").SetText("entity.randomWeightInPounds($st)")
	c.Equal("entity.randomWeightInPounds($st)", a.CommonOptions.WeightScript)
	stringFieldFor(t, d, g.Value.KeyPrefix+"age_script").SetText(`dice.roll("1d10+20")`)
	c.Equal(`dice.roll("1d10+20")`, g.Value.AgeScript)
	c.Equal("", a.CommonOptions.AgeScript, "the gender's script is not the common one")
	c.True(d.Modified())
}

// TestAncestryEditorPanelAddGender verifies that adding a gender appends an equally weighted gender with an empty
// options block of its own, with key prefixes distinct from every other gender's, rebuilds the rows, and can be undone
// and redone.
func TestAncestryEditorPanelAddGender(t *testing.T) {
	c := check.New(t)
	d := newTestAncestryEditorDockable(gurps.NewAncestry())
	rootAncestryPanel(t, d).addGender()
	c.Equal(3, len(d.model.GenderOptions))
	added := d.model.GenderOptions[2]
	c.Equal(1, added.Weight)
	c.NotNil(added.Value, "the new gender has an options block")
	c.Equal("", added.Value.Name, "the new gender starts unnamed")
	prefixes := make(map[string]bool)
	for _, g := range d.model.GenderOptions {
		for _, prefix := range []string{g.KeyPrefix, g.Value.KeyPrefix} {
			c.NotEqual("", prefix, "every gender and its options have a key prefix")
			c.False(prefixes[prefix], "key prefix %q is used more than once", prefix)
			prefixes[prefix] = true
		}
	}
	c.Equal(3, len(genderRows(d)), "the rows are rebuilt")
	c.NotNil(d.targetMgr.Find(added.Value.KeyPrefix+"name"), "the new gender's name field exists")
	c.True(d.Modified())
	c.True(d.undoMgr.CanUndo())

	d.undoMgr.Undo()
	c.Equal(2, len(d.model.GenderOptions), "undo removes the gender")
	c.Equal(2, len(genderRows(d)))
	c.False(d.Modified())

	d.undoMgr.Redo()
	c.Equal(3, len(d.model.GenderOptions), "redo adds it back")
	c.Equal(3, len(genderRows(d)))
}

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
