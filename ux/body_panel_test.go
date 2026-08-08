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
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
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

	var keys []string
	collectRefKeys(panel.AsPanel(), "body:", &keys)
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

func collectRefKeys(panel *unison.Panel, prefix string, keys *[]string) {
	if strings.HasPrefix(panel.RefKey, prefix) {
		*keys = append(*keys, panel.RefKey)
	}
	for _, child := range panel.Children() {
		collectRefKeys(child, prefix, keys)
	}
}
