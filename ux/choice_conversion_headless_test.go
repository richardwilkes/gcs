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
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// TestChoiceConversionClosesEditors verifies that no editor is left open on a container whose kind has changed under
// it. Converting closes the editor first, and undo and redo discard an editor opened on the other kind of container,
// since applying it would silently convert the container back and throw away what the undo or redo restored.
func TestChoiceConversionClosesEditors(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &askToConvertChoiceContainers, func(_, _ string) bool { return true })
	group := gurps.NewTrait(nil, nil, true)
	group.Name = "Advantages"
	group.Tags = []string{"Advantage"}
	data := gurps.NewTemplate()
	data.Traits = []*gurps.Trait{group}
	var template *Template
	screen.Do(func() {
		template = NewTemplate("test"+gurps.TemplatesExt, data)
		DisplayNewDockable(template)
	})
	editorsOpen := func() int {
		var count int
		screen.Do(func() {
			count = len(AllMatchingDockables(func(d unison.Dockable) bool {
				e, ok := d.(*editor[*gurps.Trait, *gurps.TraitEditData])
				return ok && e.target == group
			}))
		})
		return count
	}
	var mgr *unison.UndoManager
	screen.Do(func() {
		mgr = unison.UndoManagerFor(template.Traits.Table)
		EditTrait(template, group)
	})
	c.NotNil(mgr, "the template must have an undo manager")
	c.Equal(1, editorsOpen(), "the group's editor must be open")

	screen.Do(func() {
		table := template.Traits.Table
		table.SetSelectionMap(map[tid.TID]bool{group.ID(): true})
		convertSelectedContainers[*gurps.Trait, *gurps.TraitEditData](template, table, choiceContainerKind)
	})
	c.True(gurps.IsTemplateChoiceContainer(group), "the group must have become a choice")
	c.Equal(0, editorsOpen(), "converting must close the group's editor")

	screen.Do(func() { EditTrait(template, group) })
	c.Equal(1, editorsOpen(), "the choice's editor must be open")
	screen.Do(mgr.Undo)
	c.False(gurps.IsTemplateChoiceContainer(group), "undo must turn the choice back into a group")
	c.Equal([]string{"Advantage"}, group.Tags, "undo must restore what the conversion removed")
	c.Equal(0, editorsOpen(), "undo must discard the editor opened on the choice")

	screen.Do(func() { EditTrait(template, group) })
	c.Equal(1, editorsOpen(), "the group's editor must be open")
	screen.Do(mgr.Redo)
	c.True(gurps.IsTemplateChoiceContainer(group), "redo must convert the group again")
	c.Equal(0, editorsOpen(), "redo must discard the editor opened on the group")
}
