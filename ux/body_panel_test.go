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
	"github.com/richardwilkes/unison"
)

// TestBodyPanelNotesRefKeysAreUnique verifies that each hit location's notes field gets a reference key of its own,
// even though the factory Humanoid body uses "leg" and "arm" twice each, so the fields can't be keyed by location ID.
// TargetMgr.Find resolves a key to the first panel that matches, so when the Right and Left Leg shared the key
// "body:leg", typing in one of them rebuilt the panel and restored focus to the other -- writing the rest of the typed
// text into the wrong location's notes -- and undoing an edit applied the restored text to the wrong location too.
func TestBodyPanelNotesRefKeysAreUnique(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	body := gurps.FactoryBody()
	counts := make(map[string]int)
	countLocationIDs(body, counts)
	c.Equal(2, counts["leg"], "the factory body has two locations with the ID \"leg\"")
	c.Equal(2, counts["arm"], "the factory body has two locations with the ID \"arm\"")
	entity.SheetSettings.BodyType = body
	panel := NewBodyPanel(entity, NewTargetMgr(unison.NewPanel()))

	isNotesField := func(p *unison.Panel) bool { return strings.HasPrefix(p.RefKey, "body:") }
	fields := uxtest.PanelsMatching(panel.AsPanel(), isNotesField)
	keys := make([]string, 0, len(fields))
	for _, field := range fields {
		keys = append(keys, field.RefKey)
	}
	c.Equal(countLocations(body), len(keys), "every hit location must contribute a notes field")
	seen := make(map[string]bool, len(keys))
	for _, key := range keys {
		c.False(seen[key], "reference key %q is used by more than one hit location notes field", key)
		seen[key] = true
	}
}

// TestBodyLocationRefKeyFollowsIndexPath verifies that the reference key is derived from the location's position within
// the body rather than its ID, so that sibling locations sharing an ID are still told apart.
func TestBodyLocationRefKeyFollowsIndexPath(t *testing.T) {
	c := check.New(t)
	c.Equal("body:0", bodyLocationRefKey([]int{0}), "a top-level location is keyed by its index")
	c.Equal("body:3:1", bodyLocationRefKey([]int{3, 1}), "a nested location is keyed by its full index path")
	c.NotEqual(bodyLocationRefKey([]int{1}), bodyLocationRefKey([]int{11}),
		"index paths must not run together into the same key")
}

func countLocationIDs(body *gurps.Body, counts map[string]int) {
	for _, location := range body.Locations {
		counts[location.ID()]++
		if location.SubTable != nil {
			countLocationIDs(location.SubTable, counts)
		}
	}
}

func countLocations(body *gurps.Body) int {
	count := 0
	for _, location := range body.Locations {
		count++
		if location.SubTable != nil {
			count += countLocations(location.SubTable)
		}
	}
	return count
}

// TestBodyPanelHeadingFollowsTheBodyName verifies that the body block is named for the body, both when it is built
// and as the name changes, and that a named body is given a heading of that name, which comes first and takes the
// title strip, while an unnamed body has none, so that its border keeps the strip; a heading as well would push the
// header row down a strip.
func TestBodyPanelHeadingFollowsTheBodyName(t *testing.T) {
	for _, names := range [][]string{
		{"Humanoid", "Custom", ""}, // Built with a heading, which is renamed and then removed.
		{"", "Named", ""},          // Built without a heading, which is added and then removed.
		{"Humanoid", "", "Custom"}, // Built with a heading, which is removed and then added again.
	} {
		t.Run(strings.Join(names, " to "), func(t *testing.T) {
			c := check.New(t)
			entity := gurps.NewEntity()
			body := gurps.FactoryBody()
			body.Name = names[0]
			entity.SheetSettings.BodyType = body
			panel := NewBodyPanel(entity, NewTargetMgr(unison.NewPanel()))
			strip := panel.titledBorder.TitleHeight() + 1
			for i, name := range names {
				if i != 0 {
					body.Name = name
					panel.sync(true)
				}
				_, pref, _ := panel.Sizes(geom.Size{})
				panel.SetFrameRect(geom.NewRect(0, 0, pref.Width, pref.Height))
				panel.ValidateLayout()
				c.Equal(name, panel.Accessibility.Name)
				c.Equal(name, panel.titledBorder.Title)
				if name == "" {
					c.Nil(panel.heading, "an unnamed body has no heading")
					c.Nil(blockHeading(panel.AsPanel()))
					c.False(panel.titledBorder.HeadingInContent, "the border keeps the title strip")
					c.Equal(strip, panel.titledBorder.Insets().Top)
					c.Equal(panel.header.AsPanel(), panel.Children()[0], "the header row comes first")
					c.Equal(strip, panel.header.FrameRect().Y, "the header row lies right under the title strip")
					continue
				}
				if panel.heading == nil {
					t.Fatalf("the body named %q has a heading", name)
				}
				c.Equal(panel.heading, blockHeading(panel.AsPanel()))
				c.Equal(name, panel.heading.Accessibility.Name, "the heading follows the body's name")
				c.True(panel.titledBorder.HeadingInContent, "the border leaves the title strip to the heading")
				c.Equal(float32(1), panel.titledBorder.Insets().Top)
				c.Equal(panel.heading, panel.Children()[0], "the heading comes first")
				c.Equal(panel.header.AsPanel(), panel.Children()[1], "the header row follows it")
				c.Equal(panel.titledBorder.TitleStrip(panel.FrameRect().Size), panel.heading.FrameRect())
				c.Equal(strip, panel.header.FrameRect().Y, "the header row lies right under the title strip, as before")
			}
		})
	}
}

// TestBodyPanelSubTableNameShowsAndHidesItsLocations verifies that a click on the name of a hit location with a
// sub-table hides the sub-table's locations, which start out shown, and that another click shows them again.
func TestBodyPanelSubTableNameShowsAndHidesItsLocations(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	body := &gurps.Body{Name: "Test", Roll: gurps.Roller.Parse("1d")}
	for _, id := range []string{"a", "b"} {
		loc := gurps.NewHitLocation(entity, "")
		loc.LocID = id
		loc.Slots = 3
		body.AddLocation(loc)
	}
	sub := &gurps.Body{Roll: gurps.Roller.Parse("1d")}
	sub.SetOwningLocation(body.Locations[0])
	body.Locations[0].SubTable = sub
	for _, id := range []string{"x", "y"} {
		loc := gurps.NewHitLocation(entity, "")
		loc.LocID = id
		loc.Slots = 3
		sub.AddLocation(loc)
	}
	body.Update(entity)
	entity.SheetSettings.BodyType = body
	var panel *BodyPanel
	screen.Do(func() {
		sheet := NewSheet("test"+gurps.SheetExt, entity)
		DisplayNewDockable(sheet)
		if panels := uxtest.PanelsOfType[*BodyPanel](sheet.AsPanel()); len(panels) == 1 {
			panel = panels[0]
		}
	})
	if panel == nil {
		t.Fatal("the sheet shows the body")
	}
	// shown returns the reference keys of the notes fields the panel shows, one for each location it lists.
	shown := func() []string {
		var keys []string
		screen.Do(func() {
			for _, field := range uxtest.PanelsMatching(panel.AsPanel(), func(p *unison.Panel) bool {
				return strings.HasPrefix(p.RefKey, "body:")
			}) {
				keys = append(keys, field.RefKey)
			}
		})
		return keys
	}
	all := []string{"body:0", "body:0:0", "body:0:1", "body:1"}
	c.Equal(all, shown(), "the sub-table's locations start out shown")
	for _, want := range [][]string{{"body:0", "body:1"}, all} {
		screen.Do(func() {
			buttons := uxtest.PanelsOfType[*unison.Button](panel.AsPanel())
			c.Equal(1, len(buttons), "only the location with the sub-table is a button")
			if len(buttons) == 1 {
				buttons[0].ClickCallback()
			}
		})
		c.Equal(want, shown(), "a click on the location's name shows or hides its sub-table")
	}
}
