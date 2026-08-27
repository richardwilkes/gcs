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
