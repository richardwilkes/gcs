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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
)

// Every fallback tier keeps the marker text as Raw, and only the two legacy tiers mark it Legacy.
func TestParseMarkerRawPreservedAcrossAllFallbackTiers(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		raw    string
		legacy bool
	}{
		{raw: "Rare: Acceleration, Altitude Sickness, etc.", legacy: true}, // tier 1
		{raw: "Class: Mammalia", legacy: true},                             // tier 2
		{raw: "Weapon Name", legacy: false},                                // tier 3 -- a blessed current form now
	} {
		marker, ok := nameable.NewMarker(tc.raw)
		c.True(ok)
		c.Equal(tc.legacy, marker.Legacy)
		c.Equal(tc.raw, marker.Raw)
	}
}

// Tier 1: "Label: item, item[, item...], etc." -- real examples pulled from the GCS master library scan.

// A single "Label: item, item[, item...], etc." segment, with or without the final period, is a legacy marker that
// allows empty and free-form input, with its items as options and the text after the colon as its tooltip.
func TestParseMarkerTier1ExampleList(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		raw     string
		label   string
		options []string
		tooltip string
	}{
		{
			raw:   "Rare: Acceleration, Altitude Sickness, Bends, Seasickness, Space Sickness, Nanomachines, etc.",
			label: "Rare",
			options: []string{
				"Acceleration", "Altitude Sickness", "Bends", "Seasickness", "Space Sickness", "Nanomachines",
			},
			tooltip: "Acceleration, Altitude Sickness, Bends, Seasickness, Space Sickness, Nanomachines, etc.",
		},
		{
			raw:     "Skill: Fast-Draw, Smith, etc",
			label:   "Skill",
			options: []string{"Fast-Draw", "Smith"},
			tooltip: "Fast-Draw, Smith, etc",
		},
	} {
		marker, ok := nameable.NewMarker(tc.raw)
		c.True(ok, tc.raw)
		c.True(marker.Legacy, tc.raw)
		c.Equal(tc.label, marker.Label, tc.raw)
		c.Equal(tc.options, marker.Options, tc.raw)
		c.Equal(tc.tooltip, marker.Tooltip, tc.raw)
		c.True(marker.AllowEmpty, tc.raw)
		c.True(marker.FreeForm, tc.raw)
	}
}

// Tier 2: "Label: anything" -- real examples pulled from the GCS master library scan.

// A single "Label: text" segment that is not a tier-1 example list is a legacy marker that allows empty and free-form
// input, with the text after the colon as its tooltip and no options.
func TestParseMarkerTier2SingleValue(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		raw     string
		label   string
		tooltip string
	}{
		{raw: "Class: Mammalia", label: "Class", tooltip: "Mammalia"},
		// A single item before "etc." isn't a strong enough signal for an example list.
		{raw: "Absurd: chocolate, etc.", label: "Absurd", tooltip: "chocolate, etc."},
		// Items with no "etc." read as a closed, exhaustive set (e.g. a fixed enum), not as examples.
		{raw: "Type: Mental, Physical, Magical or Chi", label: "Type", tooltip: "Mental, Physical, Magical or Chi"},
		// A comma inside a quoted phrase throws off tier 1's item splitting, so this lands safely in tier 2 rather
		// than being mis-split into garbage options.
		{
			raw:     `Habit: Witless witticisms, calling people "nuncle," abuse of custard, etc.`,
			label:   "Habit",
			tooltip: `Witless witticisms, calling people "nuncle," abuse of custard, etc.`,
		},
	} {
		marker, ok := nameable.NewMarker(tc.raw)
		c.True(ok, tc.raw)
		c.True(marker.Legacy, tc.raw)
		c.Equal(tc.label, marker.Label, tc.raw)
		c.Equal(tc.tooltip, marker.Tooltip, tc.raw)
		c.Equal(0, len(marker.Options), tc.raw)
		c.True(marker.AllowEmpty, tc.raw)
		c.True(marker.FreeForm, tc.raw)
	}
}

// The leading segment before the colon must be 1-40 characters to read as a short label. Exactly 40 stays in tier 2;
// 41 falls through to tier 3 (no longer tagged Legacy -- it's a blessed current form), taking the whole string as the
// label instead of splitting it into a label and tooltip.
func TestParseMarkerTier2LabelLengthBoundary(t *testing.T) {
	c := check.New(t)
	at40 := strings.Repeat("A", 40) + ": tail"
	marker, ok := nameable.NewMarker(at40)
	c.True(ok)
	c.True(marker.Legacy)
	c.Equal(strings.Repeat("A", 40), marker.Label)
	c.Equal("tail", marker.Tooltip)

	at41 := strings.Repeat("A", 41) + ": tail"
	marker, ok = nameable.NewMarker(at41)
	c.True(ok)
	c.False(marker.Legacy)
	c.Equal(at41, marker.Label)
	c.Equal("", marker.Tooltip)
}

// Tier 3: no colon-shape at all -- a blessed current form (not Legacy), with AllowEmpty/FreeForm set.

// A single segment matching neither legacy pattern never fails to parse: it is a current-form marker that allows empty
// and free-form input, labeled with its trimmed text, with no tooltip or options.
func TestParseMarkerTier3PlainLabel(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		raw   string
		label string
	}{
		{raw: "Weapon Name", label: "Weapon Name"},
		{raw: "  Weapon Name  ", label: "Weapon Name"},
		// No space follows the colons, so neither legacy pattern matches.
		{raw: "Element:Fire:Water", label: "Element:Fire:Water"},
		{raw: "@@@", label: "@@@"},
		{raw: ":", label: ":"},
		{raw: "::::", label: "::::"},
		{raw: "no colon here at all", label: "no colon here at all"},
	} {
		marker, ok := nameable.NewMarker(tc.raw)
		c.True(ok, tc.raw)
		c.False(marker.Legacy, tc.raw)
		c.Equal(tc.label, marker.Label, tc.raw)
		c.Equal("", marker.Tooltip, tc.raw)
		c.Equal(0, len(marker.Options), tc.raw)
		c.True(marker.AllowEmpty, tc.raw)
		c.True(marker.FreeForm, tc.raw)
	}
}
