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

// Apply replaces a marker whose key has a replacement. An unresolved current-form marker, whether or not it allows
// empty, is shown as just "@Label@" rather than its raw markup, keeping the '@' wrapper so it still visibly reads as an
// unresolved nameable, while an unresolved legacy "Label: text" marker keeps its full text, even when a marker with
// the same label resolves. A marker that does not parse is written back verbatim.
func TestApplyUnresolvedAllowEmptyComboFallsBackToLabel(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		in           string
		replacements map[string]string
		want         string
	}{
		// The persisted state after Reduce omits an optional combo left at its "none" state.
		{in: "A @Element|Fire|Water|?@ spell", replacements: map[string]string{}, want: "A @Element@ spell"},
		// A required combo can be missing too, e.g. cleared via the "trash" action so a template can defer providing it
		// until application.
		{in: "A @Element|Fire|Water@ spell", replacements: map[string]string{}, want: "A @Element@ spell"},
		{
			in:           "A @Element|Fire|Water|?@ spell",
			replacements: map[string]string{"Element|Fire|Water|?": "Fire"},
			want:         "A Fire spell",
		},
		{
			in:           "Patron (@Who: A deity@) and @Who@",
			replacements: map[string]string{"Who": "The King"},
			want:         "Patron (@Who: A deity@) and The King",
		},
		// A leading '|' leaves the label empty, which NewMarker rejects.
		{in: "A @|Fire|Water@ spell", replacements: map[string]string{}, want: "A @|Fire|Water@ spell"},
	} {
		// The second pass applies the compiled text from the cache.
		for range 2 {
			c.Equal(tc.want, nameable.Apply(tc.in, tc.replacements), tc.in)
		}
	}
}

// A marker resolved in one entry must still resolve correctly when it recurs in another entry, and an entry with no
// markers at all must pass through untouched. A "Habit" replacement resolves "@Habit@" without resolving an unresolved
// legacy example list with that label, which shows as "@Habit@", and a replacement under the empty key, which literal
// text carries when compiled, replaces nothing.
func TestApplyToListResolvesSameMarkerAcrossMultipleEntries(t *testing.T) {
	c := check.New(t)
	m := map[string]string{"Element|Fire|Water": "Fire", "Habit": "Nail-biting", "": "unused"}
	got := nameable.ApplyToList([]string{
		"A @Element|Fire|Water@ spell",
		"No markers here",
		"Another @Element|Fire|Water@ spell, and an @Element|Fire|Water|?@ one too",
		"@Habit: Curtness, Ranting, Scowling, etc.@",
		"@Habit@",
	}, m)
	c.Equal([]string{
		"A Fire spell",
		"No markers here",
		"Another Fire spell, and an @Element@ one too",
		"@Habit@",
		"Nail-biting",
	}, got)
}

// Reduce keeps the replacements whose key is in nameables, an empty value included, since "empty" and "not set" are
// distinct states, drops those still Unset, and returns nil, not an empty map, when nameables is empty.
func TestReduceOmitsReplacementsNotInNameables(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		nameables    map[string]string
		replacements map[string]string
		want         map[string]string
	}{
		{
			nameables:    map[string]string{"Habit": "Bad Habit"},
			replacements: map[string]string{"Habit": "Nail-biting", "Weapon Name": "Excalibur"},
			want:         map[string]string{"Habit": "Nail-biting"},
		},
		{
			nameables:    map[string]string{"Habit": "Bad Habit", "Weapon Name": "A weapon name"},
			replacements: map[string]string{"Habit": "Nail-biting", "Weapon Name": nameable.Unset},
			want:         map[string]string{"Habit": "Nail-biting"},
		},
		{
			nameables:    map[string]string{"Element|Fire|Water|?": ""},
			replacements: map[string]string{"Element|Fire|Water|?": ""},
			want:         map[string]string{"Element|Fire|Water|?": ""},
		},
		{nameables: nil, replacements: map[string]string{"Habit": "Nail-biting"}, want: nil},
	} {
		got := nameable.Reduce(tc.nameables, tc.replacements)
		c.Equal(tc.want, got, "got %#v from nameables %v", got, tc.nameables)
	}
}

// A replacements map loaded from an old file may be keyed on an unnormalized ordering of a current-format marker's
// segments, while nameables (as produced by Extract) is always keyed on the normalized form. Normalize must rewrite the
// value to live under the normalized key -- not the raw key -- or the entry would silently fail to resolve later,
// since Reduce/Missing/Apply/ApplyToList all expect replacements to be normalized already. (Legacy markers are exempt
// from this: their key is always their raw text, unchanged -- see TestKeyIncludesAllowEmptyAndFreeFormTokens.)
func TestNormalizeRewritesUnnormalizedReplacementKey(t *testing.T) {
	c := check.New(t)
	const raw = "Element|?|Fire|Water"
	marker, ok := nameable.NewMarker(raw)
	c.True(ok)
	normalizedKey := marker.Key()
	c.NotEqual(raw, normalizedKey)

	replacements := map[string]string{raw: "Fire"}
	c.Equal(map[string]string{normalizedKey: "Fire"}, nameable.Normalize(replacements))
}

// Uses counts every reference to each marker key across all the strings it is given.
func TestUsesCountsEachReference(t *testing.T) {
	c := check.New(t)
	c.Equal(map[string]int{"style": 2, "weapon": 1}, nameable.Uses("A @style@ with a @weapon@", "More @style@", "none"))
}

// Describe says what a nameable asks for, passing its code, name and options through em: its label, or its code when
// it has none, its options as suggestions when it is free-form or as a choice otherwise, listing at most four, and
// whether it takes any text or can't be blank.
func TestDetailDescribe(t *testing.T) {
	c := check.New(t)
	plain := func(s string) string { return s }
	bold := func(s string) string { return "*" + s + "*" }
	style := nameable.Detail{
		Label: "Martial arts style", Options: []string{"Judo", "Karate", "Wrestling", "Boxing"},
		FreeForm: true,
	}
	c.Equal("*@style@* asks for *Martial arts style*, suggesting *Judo*, *Karate*, *Wrestling* or *Boxing*",
		style.Describe("style", bold))
	style.Options = append(style.Options, "Sumo")
	style.FreeForm = false
	c.Equal("@style@ asks for Martial arts style, one of Judo, Karate, Wrestling or 2 more", style.Describe("style", plain))
	weapon := nameable.Detail{FreeForm: true, NotBlank: true}
	c.Equal("@weapon@ asks for weapon as any text, which can't be blank", weapon.Describe("weapon", plain),
		"a nameable with no name is named by its code")
	only := nameable.Detail{Label: "Hand", Options: []string{"Left"}}
	c.Equal("@hand@ asks for Hand, one of Left", only.Describe("hand", plain))
	empty := nameable.Detail{Label: "Nothing"}
	c.Equal("@x@ asks for Nothing, with nothing to choose from", empty.Describe("x", plain))
}
