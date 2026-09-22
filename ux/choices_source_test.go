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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// pickTwo returns template picker data asking for two of whatever the container holds.
func pickTwo() gurps.TemplatePicker {
	var tp gurps.TemplatePicker
	tp.Type = picker.Count
	tp.Qualifier.Compare = criteria.EqualsNumber
	tp.Qualifier.Qualifier = fxp.FromInteger(2)
	return tp
}

// newSourcedPickerEditor returns a template holding one sourced container, along with an editor over that container
// set up the way displayEditor would leave it. The editor's UI is not built; only the two copies of the edit data it
// compares, which is all the apply path reads.
func newSourcedPickerEditor(t *testing.T) (*Template, *gurps.Trait, *editor[*gurps.Trait, *gurps.TraitEditData]) {
	t.Helper()
	data := gurps.NewTemplate()
	container := gurps.NewTrait(nil, nil, true)
	container.Name = "Racial Package"
	container.Source.LibraryFile = gurps.LibraryFile{Library: "Master Library", Path: "Traits/Racial.trait"}
	container.Source.TID = container.TID
	data.Traits = []*gurps.Trait{container}
	template := newTestTemplateDockable("Choices", data)

	e := &editor[*gurps.Trait, *gurps.TraitEditData]{owner: template, target: container}
	e.beforeData = &gurps.TraitEditData{}
	e.beforeData.CopyFrom(container)
	e.editorData = &gurps.TraitEditData{}
	e.editorData.CopyFrom(container)
	return template, container, e
}

// TestTemplatePickerAdded verifies what counts as gaining choices. Only the crossing from none to some matters, and
// the comparison is on meaning rather than on the struct, since a picker that has been turned off keeps whatever
// qualifier was last typed into it.
func TestTemplatePickerAdded(t *testing.T) {
	c := check.New(t)
	unchanged := &gurps.TraitEditData{}
	edited := &gurps.TraitEditData{}
	c.False(templatePickerAdded(unchanged, edited), "two untouched containers must not read as gaining choices")

	edited.TemplatePicker = pickTwo()
	c.True(templatePickerAdded(unchanged, edited), "authoring choices onto a container with none must read as gaining them")
	c.False(templatePickerAdded(edited, unchanged), "removing authored choices must not read as gaining them")

	retyped := &gurps.TraitEditData{}
	retyped.TemplatePicker = pickTwo()
	retyped.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(3)
	c.False(templatePickerAdded(edited, retyped), "altering choices already there must not read as gaining them")

	off := &gurps.TraitEditData{}
	off.TemplatePicker = pickTwo()
	off.TemplatePicker.Type = picker.NotApplicable
	c.False(templatePickerAdded(unchanged, off),
		"a picker turned off must match an unset one, whatever qualifier it still carries")
	c.True(templatePickerAdded(off, edited), "a picker turned off is the same starting point as an unset one")
}

// TestAuthoringChoicesClearsTheSource verifies that authoring picker data onto a sourced container drops its link to
// the library item it was copied from, and that undoing the edit puts the link back.
//
// Both halves matter. Left claiming a source, the container reads as out of sync with it -- the picker data is part of
// the hash the two are compared on -- and Sync With Source, the remedy offered for that, assigns the source's whole
// container block over it and takes the authored choices away. And the source lives on the node rather than in the
// editor's data, so replaying the edit data on undo cannot restore it: without the separate capture the link would be
// gone for good, saved file and all.
func TestAuthoringChoicesClearsTheSource(t *testing.T) {
	c := check.New(t)
	template, container, e := newSourcedPickerEditor(t)
	mgr := unison.UndoManagerFor(template)
	c.NotNil(mgr, "the template must have an undo manager")
	before := container.GetSource()
	c.False(before.IsZero(), "the container must start out sourced")

	e.editorData.TemplatePicker = pickTwo()
	e.applyEdit()

	c.Equal("Pick 2", container.TemplatePicker.String(), "the authored choices must be applied")
	c.True(container.GetSource().IsZero(), "a container given choices must stop claiming a source")

	mgr.Undo()
	c.Equal("", container.TemplatePicker.String(), "undo must take the choices back off")
	c.Equal(before, container.GetSource(), "undo must put the source back")

	mgr.Redo()
	c.Equal("Pick 2", container.TemplatePicker.String(), "redo must put the choices back")
	c.True(container.GetSource().IsZero(), "redo must clear the source again")
}

// TestEditingSomethingElseKeepsTheSource verifies that the source survives every other kind of edit. Diverging from a
// library item is ordinary: the mismatch marker and Sync With Source exist for it. Only the choices are special, and
// only because syncing them away is destructive.
func TestEditingSomethingElseKeepsTheSource(t *testing.T) {
	c := check.New(t)
	_, container, e := newSourcedPickerEditor(t)
	before := container.GetSource()

	e.editorData.Name = "Renamed Package"
	e.applyEdit()

	c.Equal("Renamed Package", container.Name, "the edit must be applied")
	c.Equal(before, container.GetSource(), "renaming must leave the source alone")
}
