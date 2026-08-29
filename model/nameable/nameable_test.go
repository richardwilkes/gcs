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

func TestApplyResolvedCombo(t *testing.T) {
	c := check.New(t)
	m := map[string]string{"Element|Fire|Water|?": "Fire"}
	c.Equal("A Fire spell", nameable.Apply("A @Element|Fire|Water|?@ spell", m))
}

func TestApplyUnresolvedAllowEmptyComboFallsBackToLabel(t *testing.T) {
	c := check.New(t)
	// Simulates the persisted state after Reduce omits an optional combo left at its "none" state: the key never
	// makes it into m, so Apply must not leave the full raw "@Element|Fire|Water|?@" markup in displayed text, but
	// must still keep the '@' wrapper so the value visibly reads as an unresolved nameable.
	m := map[string]string{}
	c.Equal("A @Element@ spell", nameable.Apply("A @Element|Fire|Water|?@ spell", m))
}

func TestApplyUnresolvedRequiredComboFallsBackToLabel(t *testing.T) {
	c := check.New(t)
	// A required combo can still be missing from m -- e.g. cleared via the "trash" action so a template can defer
	// providing it until application -- and must fall back the same as an optional one left at "none".
	m := map[string]string{}
	c.Equal("A @Element@ spell", nameable.Apply("A @Element|Fire|Water@ spell", m))
}

func TestApplyUnresolvedPlainKeyStaysRaw(t *testing.T) {
	c := check.New(t)
	m := map[string]string{}
	c.Equal("A @Weapon Name@ spell", nameable.Apply("A @Weapon Name@ spell", m))
}

func TestApplyUnresolvedLegacyExampleListFallsBackToShortLabel(t *testing.T) {
	c := check.New(t)
	// A real, unmodified marker pulled from the GCS master library (the Resistant trait). Unresolved, it should
	// collapse to its tier-1-parsed label rather than dumping the whole raw example list into the display.
	m := map[string]string{}
	raw := "Resistant to @Rare: Acceleration, Altitude Sickness, Bends, Seasickness, Space Sickness, Nanomachines, " +
		"etc.@ (+3)"
	c.Equal("Resistant to @Rare@ (+3)", nameable.Apply(raw, m))
}

func TestApplyUnresolvedLegacyLabelWithColonFallsBackToShortLabel(t *testing.T) {
	c := check.New(t)
	// A real, unmodified marker pulled from the GCS master library. No safe options list, but tier 2 still shortens
	// the display down to the label instead of the full "Class: Mammalia" text.
	m := map[string]string{}
	c.Equal("A @Class@ trait", nameable.Apply("A @Class: Mammalia@ trait", m))
}

func TestApplyToListUnresolvedAllowEmptyComboFallsBackToLabel(t *testing.T) {
	c := check.New(t)
	m := map[string]string{}
	got := nameable.ApplyToList([]string{"A @Element|Fire|Water|?@ spell"}, m)
	c.Equal([]string{"A @Element@ spell"}, got)
}

func TestApplyUnresolvedMarkersWithSharedLabelDoNotBleed(t *testing.T) {
	c := check.New(t)
	m := map[string]string{"Who": "The King"}
	for range 50 {
		c.Equal("Patron (@Who@) and The King", nameable.Apply("Patron (@Who: A deity@) and @Who@", m))
		if t.Failed() {
			// Stop at the first bad iteration instead of repeating the same failure 500 times in the test output.
			break
		}
	}
}

func TestApplyToListUnresolvedMarkersWithSharedLabelDoNotBleedAcrossEntries(t *testing.T) {
	c := check.New(t)
	m := map[string]string{"Habit": "Nail-biting"}
	for range 50 {
		got := nameable.ApplyToList([]string{
			"@Habit: Curtness, Ranting, Scowling, etc.@",
			"@Habit@",
		}, m)
		c.Equal([]string{
			"@Habit@",
			"Nail-biting",
		}, got)
		if t.Failed() {
			// Stop at the first bad iteration instead of repeating the same failure 500 times in the test output.
			break
		}
	}
}

func TestApplyToListEmptyInputReturnsNil(t *testing.T) {
	c := check.New(t)
	c.Equal([]string(nil), nameable.ApplyToList(nil, nil))
}

func TestApplyMalformedMarkerWrittenBackVerbatim(t *testing.T) {
	c := check.New(t)
	// A leading '|' makes the label segment empty, which NewMarker rejects. Since the marker doesn't parse, its
	// original span is written back out unchanged rather than being treated as resolved or unresolved.
	m := map[string]string{}
	c.Equal("A @|Fire|Water@ spell", nameable.Apply("A @|Fire|Water@ spell", m))
}

func TestApplyToListRetainsMalformedReplacementKeyWithoutAffectingOthers(t *testing.T) {
	c := check.New(t)
	// A replacements key that itself fails to parse as a marker (here, "") is retained as-is while building the
	// normalized-key lookup map -- it simply never matches any real marker's key, and doesn't disturb resolution of
	// the other, well-formed entries.
	m := map[string]string{"": "unused", "Element|Fire|Water": "Fire"}
	got := nameable.ApplyToList([]string{"A @Element|Fire|Water@ spell"}, m)
	c.Equal([]string{"A Fire spell"}, got)
}

func TestApplyToListResolvesSameMarkerAcrossMultipleEntries(t *testing.T) {
	c := check.New(t)
	// ApplyToList extracts its full marker set from every entry in the list up front, so a marker resolved via one
	// entry's key must still resolve correctly when it recurs in another entry, and an entry with no markers at all
	// must pass through untouched.
	m := map[string]string{"Element|Fire|Water": "Fire"}
	got := nameable.ApplyToList([]string{
		"A @Element|Fire|Water@ spell",
		"No markers here",
		"Another @Element|Fire|Water@ spell, and an @Element|Fire|Water|?@ one too",
	}, m)
	c.Equal([]string{
		"A Fire spell",
		"No markers here",
		"Another Fire spell, and an @Element@ one too",
	}, got)
}

func TestReduceEmptyNameablesReturnsNil(t *testing.T) {
	c := check.New(t)
	c.Equal(map[string]string(nil), nameable.Reduce(nil, map[string]string{"Habit": "Nail-biting"}))
}

func TestReduceOmitsReplacementsNotInNameables(t *testing.T) {
	c := check.New(t)
	nameables := map[string]string{"Habit": "Bad Habit"}
	replacements := map[string]string{"Habit": "Nail-biting", "Weapon Name": "Excalibur"}
	c.Equal(map[string]string{"Habit": "Nail-biting"}, nameable.Reduce(nameables, replacements))
}

func TestReduceStoresUnderNormalizedKeyEvenWhenReplacementKeyIsUnnormalized(t *testing.T) {
	c := check.New(t)
	// The replacements map is keyed on the raw, legacy marker text (as it would have been stored before marker
	// normalization existed), while nameables (as produced by Extract) is keyed on the normalized form. Reduce must
	// store the value under the normalized key -- not the raw replacement key -- or the entry silently fails to
	// resolve later since Apply/ApplyToList look up replacements by normalized key.
	const raw = "Class: Mammalia"
	marker, ok := nameable.NewMarker(raw)
	c.True(ok)
	normalizedKey := marker.Key()
	c.NotEqual(raw, normalizedKey)

	nameables := map[string]string{normalizedKey: "Mammalia"}
	replacements := map[string]string{raw: "Mammalia"}
	c.Equal(map[string]string{normalizedKey: "Mammalia"}, nameable.Reduce(nameables, replacements))
}
