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
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestBodyHash verifies that a body's hash changes with any of its data, its locations' included, at any depth, since
// the sheet's body block is rebuilt only when it does.
func TestBodyHash(t *testing.T) {
	c := check.New(t)
	original := loadTestBody(c, armAndHandBody)
	if original == nil {
		return
	}
	base := Hash64(original)
	for _, tc := range []struct {
		name string
		edit func(b *Body)
	}{
		{"name", func(b *Body) { b.Name = "Other" }},
		{"roll", func(b *Body) { b.Roll = dice.Dice{Count: 2, Sides: 6} }},
		{"location added", func(b *Body) { b.AddLocation(NewHitLocation(nil, "")) }},
		{"ID", func(b *Body) { b.Locations[0].LocID = "leg" }},
		{"choice name", func(b *Body) { b.Locations[0].ChoiceName = "Leg" }},
		{"table name", func(b *Body) { b.Locations[0].TableName = "Leg" }},
		{"slots", func(b *Body) { b.Locations[0].Slots = 2 }},
		{"hit penalty", func(b *Body) { b.Locations[0].HitPenalty = -2 }},
		{"DR bonus", func(b *Body) { b.Locations[0].DRBonus = 3 }},
		{"description", func(b *Body) { b.Locations[0].Description = "Limb" }},
		{"notes", func(b *Body) { b.Locations[0].Notes = "Limb" }},
		{"sub-table removed", func(b *Body) { b.Locations[0].SubTable = nil }},
		{"sub-table location", func(b *Body) { b.Locations[0].SubTable.Locations[0].HitPenalty = -4 }},
	} {
		edited := original.Clone(nil, nil)
		tc.edit(edited)
		c.NotEqual(base, Hash64(edited), "changing the %s changes the hash", tc.name)
	}
}

// TestNewBodyFromFile verifies that a body file loads in either layout, the current one or the older one that wraps the
// body in "hit_locations", and that the descriptions, sub-tables' included, are rewrapped only for files older than
// version 4.
func TestNewBodyFromFile(t *testing.T) {
	const long = "An attack that misses by 1 hits the torso instead, and only some attacks can target this location."
	locations := `[{"id":"arm","choice_name":"Arm","table_name":"Arm","slots":1,"description":"` + long + `",` +
		`"sub_table":{"roll":"1d","locations":[{"id":"hand","choice_name":"Hand","table_name":"Hand","slots":1,` +
		`"description":"` + long + `"}]}}]`
	for _, tc := range []struct {
		name      string
		data      string
		rewrapped bool
	}{
		{"version 4", `{"version":4,"name":"Test","roll":"3d","locations":` + locations + `}`, false},
		{"version 3", `{"version":3,"name":"Test","roll":"3d","locations":` + locations + `}`, true},
		{
			"hit_locations wrapper",
			`{"version":3,"hit_locations":{"name":"Test","roll":"3d","locations":` + locations + `}}`, true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			body := loadTestBody(c, tc.data)
			if body == nil {
				return
			}
			c.Equal("Test", body.Name)
			for _, id := range []string{"arm", "hand"} {
				loc := body.LookupLocationByID(nil, id)
				c.NotNil(loc, "%s is loaded", id)
				if loc == nil {
					continue
				}
				c.Equal(long, strings.ReplaceAll(loc.Description, "\n", " "), "%s keeps its description", id)
				c.Equal(tc.rewrapped, strings.Contains(loc.Description, "\n"), "%s is rewrapped", id)
			}
		})
	}
}

// TestBodyUniqueHitLocations verifies that a body lists each location ID once, sub-tables' included, ordered by choice
// name and then by ID.
func TestBodyUniqueHitLocations(t *testing.T) {
	c := check.New(t)
	const hand = `{"roll":"1d","locations":[{"id":"hand","choice_name":"Hand","table_name":"Hand","slots":1}]}`
	body := loadTestBody(c, `{"version":5,"name":"Test","roll":"3d","locations":[`+
		`{"id":"right_arm","choice_name":"Arm","table_name":"Right Arm","slots":1,"sub_table":`+hand+`},`+
		`{"id":"left_arm","choice_name":"Arm","table_name":"Left Arm","slots":1,"sub_table":`+hand+`},`+
		`{"id":"eye","choice_name":"Eyes","table_name":"Eyes"}]}`)
	if body == nil {
		return
	}
	want := []string{"left_arm", "right_arm", "eye", "hand"}
	// The locations come from a map, so an unsorted list could match by chance; repeat to make that vanishingly
	// unlikely.
	for range 20 {
		var ids []string
		for _, loc := range body.UniqueHitLocations(nil) {
			ids = append(ids, loc.LocID)
		}
		if !slices.Equal(want, ids) {
			c.Equal(want, ids)
			break
		}
	}
}
