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
	"github.com/richardwilkes/unison/enums/role"
)

// TestFactoryBodyHasDuplicateLocationIDs documents the reason the hit location notes fields can't be keyed by location
// ID: the factory Humanoid body uses "leg" and "arm" twice each, so every default character sheet has two locations
// with each of those IDs.
func TestFactoryBodyHasDuplicateLocationIDs(t *testing.T) {
	c := check.New(t)
	counts := make(map[string]int)
	countLocationIDs(gurps.FactoryBody(), counts)
	c.Equal(2, counts["leg"], "the factory body has two locations with the ID \"leg\"")
	c.Equal(2, counts["arm"], "the factory body has two locations with the ID \"arm\"")
}

// TestBodyPanelNotesRefKeysAreUnique verifies that each hit location's notes field gets a reference key of its own.
// TargetMgr.Find resolves a key to the first panel that matches, so when the Right and Left Leg shared the key
// "body:leg", typing in one of them rebuilt the panel and restored focus to the other -- writing the rest of the typed
// text into the wrong location's notes -- and undoing an edit applied the restored text to the wrong location too.
func TestBodyPanelNotesRefKeysAreUnique(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	body := gurps.FactoryBody()
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

func TestBodyPanelHeadingFollowsTheBodyName(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	body := gurps.FactoryBody()
	entity.SheetSettings.BodyType = body
	panel := NewBodyPanel(entity, NewTargetMgr(unison.NewPanel()))
	c.Equal(role.Heading, panel.Children()[0].Accessibility.Role, "the heading comes first")
	c.Equal(body.Name, panel.Children()[0].Accessibility.Name)
	c.Equal(body.Name, panel.Accessibility.Name)
	c.Equal(body.Name, panel.titledBorder.Title)

	body.Name = "Custom"
	panel.sync(true)
	c.Equal(role.Heading, panel.Children()[0].Accessibility.Role, "the heading comes first after a rebuild")
	c.Equal("Custom", panel.Children()[0].Accessibility.Name, "the heading follows the body's name")
	c.Equal("Custom", panel.Accessibility.Name)
	c.Equal("Custom", panel.titledBorder.Title)
}

// An unnamed body has no heading, so its border keeps the title strip; a heading as well would push the header row down
// a strip.
func TestBodyPanelWithoutANameHasNoHeading(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	body := gurps.FactoryBody()
	body.Name = ""
	entity.SheetSettings.BodyType = body
	panel := NewBodyPanel(entity, NewTargetMgr(unison.NewPanel()))
	strip := panel.titledBorder.TitleHeight() + 1
	layOut := func() {
		_, pref, _ := panel.Sizes(geom.Size{})
		panel.SetFrameRect(geom.NewRect(0, 0, pref.Width, pref.Height))
		panel.ValidateLayout()
	}
	verify := func(name string) {
		t.Helper()
		layOut()
		c.Equal(name, panel.Accessibility.Name)
		c.Equal(name, panel.titledBorder.Title)
		if name == "" {
			c.Nil(panel.heading, "an unnamed body has no heading")
			c.Nil(blockHeading(panel.AsPanel()))
			c.False(panel.titledBorder.HeadingInContent, "the border keeps the title strip")
			c.Equal(strip, panel.titledBorder.Insets().Top)
			c.Equal(panel.header.AsPanel(), panel.Children()[0], "the header row comes first")
			c.Equal(strip, panel.header.FrameRect().Y, "the header row lies right under the title strip")
			return
		}
		if panel.heading == nil {
			t.Fatal("a named body has a heading")
		}
		c.Equal(panel.heading, blockHeading(panel.AsPanel()))
		c.Equal(name, panel.heading.Accessibility.Name)
		c.True(panel.titledBorder.HeadingInContent, "the border leaves the title strip to the heading")
		c.Equal(float32(1), panel.titledBorder.Insets().Top)
		c.Equal(panel.heading, panel.Children()[0], "the heading comes first")
		c.Equal(panel.header.AsPanel(), panel.Children()[1], "the header row follows it")
		c.Equal(panel.titledBorder.TitleStrip(panel.FrameRect().Size), panel.heading.FrameRect())
		c.Equal(strip, panel.header.FrameRect().Y, "the header row lies right under the title strip, as before")
	}
	verify("")

	body.Name = "Named"
	panel.sync(true)
	verify("Named")

	body.Name = ""
	panel.sync(true)
	verify("")
}
