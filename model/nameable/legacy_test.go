// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package nameable_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestParseMarkerRealCombo(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker("Element|Fire|Water")
	c.True(ok)
	c.False(marker.Legacy)
	c.Equal("Element|Fire|Water", marker.Raw)
	c.Equal("Element", marker.Label)
	c.Equal([]string{"Fire", "Water"}, marker.Options)
	c.False(marker.AllowEmpty)
	c.False(marker.FreeForm)
}

func TestParseMarkerRawPreservedAcrossAllFallbackTiers(t *testing.T) {
	c := check.New(t)
	for _, raw := range []string{
		"Rare: Acceleration, Altitude Sickness, etc.", // tier 1
		"Class: Mammalia", // tier 2
		"Weapon Name",     // tier 3
	} {
		marker, ok := nameable.ParseMarker(raw)
		c.True(marker.Legacy)
		c.True(ok)
		c.Equal(raw, marker.Raw)
	}
}

// Tier 1: "Label: item, item[, item...], etc." -- real examples pulled from the GCS master library scan.

func TestParseMarkerTier1ExampleList(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker("Rare: Acceleration, Altitude Sickness, Bends, Seasickness, Space Sickness, Nanomachines, etc.")
	c.True(ok)
	c.True(marker.Legacy)
	c.Equal("Rare", marker.Label)
	c.Equal([]string{"Acceleration", "Altitude Sickness", "Bends", "Seasickness", "Space Sickness", "Nanomachines"},
		marker.Options)
	c.Equal("Acceleration, Altitude Sickness, Bends, Seasickness, Space Sickness, Nanomachines, etc.", marker.Tooltip)
	c.True(marker.AllowEmpty)
	c.True(marker.FreeForm)
}

func TestParseMarkerTier1NoTrailingPeriod(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker("Skill: Fast-Draw, Smith, etc")
	c.True(ok)
	c.True(marker.Legacy)
	c.Equal("Skill", marker.Label)
	c.Equal([]string{"Fast-Draw", "Smith"}, marker.Options)
	c.Equal("Fast-Draw, Smith, etc", marker.Tooltip)
	c.True(marker.AllowEmpty)
	c.True(marker.FreeForm)
}

func TestParseMarkerTier1RequiresTwoOrMoreItems(t *testing.T) {
	c := check.New(t)
	// Only one item before "etc." isn't a strong enough signal -- falls through to tier 2 instead.
	marker, ok := nameable.ParseMarker("Absurd: chocolate, etc.")
	c.True(ok)
	c.True(marker.Legacy)
	c.Equal("Absurd", marker.Label)
	c.Equal("chocolate, etc.", marker.Tooltip)
	c.Equal(0, len(marker.Options))
}

func TestParseMarkerTier1RequiresTrailingEtc(t *testing.T) {
	c := check.New(t)
	// Two items but no "etc." reads as a closed, exhaustive set (e.g. a fixed enum), not an example list -- falls
	// through to tier 2 instead of being marked free-form with those as options.
	marker, ok := nameable.ParseMarker("Type: Mental, Physical, Magical or Chi")
	c.True(ok)
	c.True(marker.Legacy)
	c.Equal("Type", marker.Label)
	c.Equal("Mental, Physical, Magical or Chi", marker.Tooltip)
	c.Equal(0, len(marker.Options))
	c.True(marker.AllowEmpty)
	c.True(marker.FreeForm)
}

// Tier 2: "Label: anything" -- real examples pulled from the GCS master library scan.

func TestParseMarkerTier2SingleValue(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker("Class: Mammalia")
	c.True(ok)
	c.True(marker.Legacy)
	c.Equal("Class", marker.Label)
	c.Equal("Mammalia", marker.Tooltip)
	c.Equal(0, len(marker.Options))
	c.True(marker.AllowEmpty)
	c.True(marker.FreeForm)
}

func TestParseMarkerTier2EmbeddedCommaFallsThroughFromTier1(t *testing.T) {
	c := check.New(t)
	// This one does end in "etc." and has 2+ comma-separated segments, but a comma embedded inside a quoted phrase
	// throws off tier 1's item splitting -- it's expected to land safely in tier 2 (a tooltip) rather than being
	// mis-split into garbage options.
	marker, ok := nameable.ParseMarker(`Habit: Witless witticisms, calling people "nuncle," abuse of custard, etc.`)
	c.True(ok)
	c.True(marker.Legacy)
	c.Equal("Habit", marker.Label)
	c.Equal(`Witless witticisms, calling people "nuncle," abuse of custard, etc.`, marker.Tooltip)
	c.Equal(0, len(marker.Options))
}

func TestParseMarkerTier2LabelLengthBoundary(t *testing.T) {
	c := check.New(t)
	// A leading segment longer than 40 characters doesn't read as a short label -- falls through to tier 3.
	marker, ok := nameable.ParseMarker("This leading segment is far too long to read as a label: short tail")
	c.True(ok)
	c.True(marker.Legacy)
	c.Equal("This leading segment is far too long to read as a label: short tail", marker.Label)
	c.Equal("", marker.Tooltip)
}

// Tier 3: no colon-shape at all -- today's behavior, unchanged except for AllowEmpty/FreeForm now being set.

func TestParseMarkerTier3PlainLabel(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker("Weapon Name")
	c.True(ok)
	c.True(marker.Legacy)
	c.Equal("Weapon Name", marker.Label)
	c.Equal("", marker.Tooltip)
	c.Equal(0, len(marker.Options))
	c.True(marker.AllowEmpty)
	c.True(marker.FreeForm)
}

func TestParseMarkerNeverFails(t *testing.T) {
	c := check.New(t)
	for _, raw := range []string{"@@@", ":", "::::", "no colon here at all"} {
		marker, ok := nameable.ParseMarker(raw)
		c.True(ok)
		c.True(marker.Legacy)
		c.Equal(raw, marker.Label)
	}
}

func TestParseMarkerEmptyRawFails(t *testing.T) {
	c := check.New(t)
	_, ok := nameable.ParseMarker("")
	c.False(ok)
}
