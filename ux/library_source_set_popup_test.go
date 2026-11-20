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
)

// TestLibrarySourceSetPopupCommands verifies the source set popup's commands, with the editor and the confirmation
// stubbed: New adds the set and shows it; a rename through Edit is recorded as the name of the set last used; Edit
// moving the set to another list type switches the views showing it to that type; Save As from the set of every file
// starts from one whole-library entry per library, under a name numbered apart from the title of the set of every file,
// and shows the saved set; and Delete has its view fall back to every file of the type, while another view showing the
// set keeps its sources as an unsaved set. Two views are needed to see the change reach the other, so the workspace is
// started; with no listing, the views read no files.
func TestLibrarySourceSetPopupCommands(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	isolateLibraryFilterSettings(t)
	uxtest.SwapForTest(t, &libraryFilterListing, func() *navigatorListing { return nil })
	settings := gurps.GlobalSettings()
	libs := settings.Libraries
	u := libs.User().Key()
	m := libs.Master().Key()
	spellsKey := gurps.ListFilterKeyForExtension(gurps.SpellsExt)

	var first, second *LibraryFilterDockable
	screen.Do(func() {
		first = openLibraryFilter(librarySkillsKey, nil, false)
		second = openLibraryFilter(librarySkillsKey, nil, false)
	})
	t.Cleanup(func() {
		screen.Do(func() {
			first.AttemptClose()
			second.AttemptClose()
		})
	})
	// command makes the choice the index picks in the view's popup, as its menu would.
	command := func(d *LibraryFilterDockable, index func(p *librarySourceSetPopup) int) {
		screen.Do(func() {
			p := d.content.sourceSetPopup()
			p.rebuildItems()
			i := index(p)
			item, _ := p.popup.ItemAt(i)
			p.popup.ChoiceMadeCallback(p.popup, i, item)
		})
	}
	// The editor accepts what it is given, recorded in given, after making the change, if any.
	var given *gurps.LibrarySourceSet
	var change func(key *string, set *gurps.LibrarySourceSet)
	uxtest.SwapForTest(t, &showLibrarySourceSetEditor,
		func(_ string, key *string, set, _ *gurps.LibrarySourceSet) bool {
			given = set.Clone()
			if change != nil {
				change(key, set)
			}
			return true
		})

	change = func(_ *string, set *gurps.LibrarySourceSet) {
		set.Name = "Mine"
		set.Entries = []gurps.LibraryFile{{Library: u, Path: "Skills"}}
	}
	command(first, func(p *librarySourceSetPopup) int { return p.newIndex })
	var mine *gurps.LibrarySourceSet
	screen.Do(func() {
		sets := settings.LibrarySourceSetsFor(librarySkillsKey)
		c.Equal(1, len(sets), "New adds the set")
		if len(sets) == 1 {
			mine = sets[0]
		}
		c.True(mine != nil && first.set == mine && !first.adHoc, "the view shows the new set")
		second.chooseSet(librarySkillsKey, mine, false)
	})

	change = func(_ *string, set *gurps.LibrarySourceSet) { set.Name = "Ours" }
	command(first, func(p *librarySourceSetPopup) int { return p.editIndex })
	c.Equal("Ours", settings.LibraryFilter.LastSets[librarySkillsKey], "the set is last used under its new name")

	change = func(key *string, _ *gurps.LibrarySourceSet) { *key = spellsKey }
	command(first, func(p *librarySourceSetPopup) int { return p.editIndex })
	screen.Do(func() {
		c.Equal(0, len(settings.LibrarySourceSetsFor(librarySkillsKey)), "the set leaves its old type")
		c.Equal([]*gurps.LibrarySourceSet{mine}, settings.LibrarySourceSetsFor(spellsKey), "and joins the new one")
		for _, d := range []*LibraryFilterDockable{first, second} {
			c.Equal(spellsKey, d.kind.key, "a view showing the set switches type with it")
			c.True(d.set == mine && !d.adHoc, "and goes on showing the set")
		}
	})

	change = nil
	screen.Do(func() { second.chooseSet(spellsKey, nil, false) })
	command(second, func(p *librarySourceSetPopup) int { return p.saveAsIndex })
	screen.Do(func() {
		c.Equal("All Spells (2)", given.Name,
			"Save As suggests a name for what the view shows, numbered apart from the set of every file")
		c.Equal(gurps.NormalizeLibrarySourceEntries([]gurps.LibraryFile{{Library: m}, {Library: u}}), given.Entries,
			"Save As from every file starts from each whole library")
		c.True(second.set != nil && second.set != mine && !second.adHoc && second.set.Name == "All Spells (2)",
			"the view shows the saved set")
		second.chooseSet(spellsKey, mine, false)
	})

	var confirmed string
	uxtest.SwapForTest(t, &confirmLibrarySourceSetDeletion, func(name string) bool {
		confirmed = name
		return true
	})
	command(first, func(p *librarySourceSetPopup) int { return p.deleteIndex })
	screen.Do(func() {
		c.Equal("Ours", confirmed, "Delete asks first")
		c.Equal(1, len(settings.LibrarySourceSetsFor(spellsKey)), "the set is removed")
		c.True(first.set == nil && !first.adHoc && first.kind.key == spellsKey,
			"its view falls back to every file of its type")
		c.True(second.adHoc && second.set != nil && second.kind.key == spellsKey,
			"another view showing it keeps it as an unsaved set")
		if second.set != nil {
			c.Equal([]gurps.LibraryFile{{Library: u, Path: "Skills"}}, second.set.Entries, "with the same sources")
		}
		c.Equal("Selected Spells", second.Title())
	})
}
