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
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
)

// CloneDetails returns nil for nil, and otherwise a copy that can be changed, options included, without changing the
// original.
func TestCloneDetails(t *testing.T) {
	c := check.New(t)
	c.Nil(nameable.CloneDetails(nil))
	original := map[string]nameable.Detail{
		"style": {Label: "Martial arts style", Options: []string{"Judo", "Karate"}, FreeForm: true},
	}
	clone := nameable.CloneDetails(original)
	c.Equal(original, clone)
	clone["style"].Options[0] = "Sumo"
	clone["weapon"] = nameable.Detail{NotBlank: true}
	c.Equal(map[string]nameable.Detail{
		"style": {Label: "Martial arts style", Options: []string{"Judo", "Karate"}, FreeForm: true},
	}, original)
}

// Describe says what a nameable asks for, passing its code, name and options through em: its label, or its code when
// it has none, its options as suggestions when it is free-form or as a choice otherwise, listing at most four, and
// whether it takes any text or can't be blank.
func TestDetailDescribe(t *testing.T) {
	c := check.New(t)
	plain := func(s string) string { return s }
	bold := func(s string) string { return "*" + s + "*" }
	for _, tc := range []struct {
		detail nameable.Detail
		code   string
		em     func(string) string
		want   string
	}{
		{
			detail: nameable.Detail{
				Label:    "Martial arts style",
				Options:  []string{"Judo", "Karate", "Wrestling", "Boxing"},
				FreeForm: true,
			},
			code: "style",
			em:   bold,
			want: "*@style@* asks for *Martial arts style*, suggesting *Judo*, *Karate*, *Wrestling* or *Boxing*",
		},
		{
			detail: nameable.Detail{
				Label:   "Martial arts style",
				Options: []string{"Judo", "Karate", "Wrestling", "Boxing", "Sumo"},
			},
			code: "style",
			em:   plain,
			want: "@style@ asks for Martial arts style, one of Judo, Karate, Wrestling or 2 more",
		},
		{
			detail: nameable.Detail{Label: "Hand", Options: []string{"Left", "Right"}, FreeForm: true, NotBlank: true},
			code:   "hand",
			em:     plain,
			want:   "@hand@ asks for Hand, suggesting Left or Right, and can't be blank",
		},
		{
			detail: nameable.Detail{Label: "Hand", Options: []string{"Left", "Right"}, NotBlank: true},
			code:   "hand",
			em:     plain,
			want:   "@hand@ asks for Hand, one of Left or Right, and can't be blank",
		},
		{
			detail: nameable.Detail{Label: "Hand", Options: []string{"Left"}},
			code:   "hand",
			em:     plain,
			want:   "@hand@ asks for Hand, one of Left",
		},
		// A nameable with no label is named by its code.
		{
			detail: nameable.Detail{FreeForm: true, NotBlank: true},
			code:   "weapon",
			em:     plain,
			want:   "@weapon@ asks for weapon as any text, which can't be blank",
		},
		{
			detail: nameable.Detail{Label: "Hand", FreeForm: true},
			code:   "hand",
			em:     plain,
			want:   "@hand@ asks for Hand as any text",
		},
		{
			detail: nameable.Detail{Label: "Nothing"},
			code:   "x",
			em:     plain,
			want:   "@x@ asks for Nothing, with nothing to choose from",
		},
	} {
		c.Equal(tc.want, tc.detail.Describe(tc.code, tc.em))
	}
}

// Extract records each marker that parses under its key, once however often it appears, in the map it is given or a
// new one when that is nil, with the value the replacements hold for that key, or Unset when they hold none.
func TestExtract(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		nameables    map[string]string
		replacements map[string]string
		in           []string
		want         map[string]string
	}{
		{
			nameables: map[string]string{"kept": "as is"},
			in:        []string{"A @Element|Fire|Water@ spell"},
			want:      map[string]string{"kept": "as is", "Element|Fire|Water": nameable.Unset},
		},
		{
			in:   []string{"A @Element|Fire|Water@ spell"},
			want: map[string]string{"Element|Fire|Water": nameable.Unset},
		},
		{
			replacements: map[string]string{"Element|Fire|Water": "Water"},
			in:           []string{"A @Element|Fire|Water@ spell"},
			want:         map[string]string{"Element|Fire|Water": "Water"},
		},
		// "|x" has no label, so it is not a marker.
		{
			in:   []string{"No markers", "@Skill@ and @Skill@", "@|x@ @Attack@"},
			want: map[string]string{"Skill": nameable.Unset, "Attack": nameable.Unset},
		},
	} {
		got := nameable.Extract(tc.nameables, tc.replacements, tc.in...)
		c.Equal(tc.want, got, "%q", tc.in)
		if tc.nameables != nil {
			c.Equal(tc.want, tc.nameables, "the map given for %q", tc.in)
		}
	}
}

// A replacements map loaded from an old file may be keyed on an unnormalized ordering of a current-format marker's
// segments, while nameables (as produced by Extract) is always keyed on the normalized form. Normalize must rewrite the
// value to live under the normalized key -- not the raw key -- or the entry would silently fail to resolve later,
// since Reduce/Missing/Apply/ApplyToList all expect replacements to be normalized already. (Legacy markers are exempt
// from this: their key is always their raw text, unchanged -- see TestKeyNormalization.) A key that does not parse as a
// marker is kept as it is, and an empty map yields nil.
func TestNormalize(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		replacements map[string]string
		want         map[string]string
	}{
		{
			replacements: map[string]string{"Element|?|Fire|Water": "Fire"},
			want:         map[string]string{"Element|Fire|Water|?": "Fire"},
		},
		// A leading '|' leaves the label empty, which NewMarker rejects.
		{replacements: map[string]string{"|Fire|Water": "Fire"}, want: map[string]string{"|Fire|Water": "Fire"}},
		{replacements: map[string]string{}, want: nil},
	} {
		got := nameable.Normalize(tc.replacements)
		c.Equal(tc.want, got, "got %#v from %v", got, tc.replacements)
	}
}

// Uses counts every reference to each marker key, within a string and across all the strings it is given.
func TestUsesCountsEachReference(t *testing.T) {
	c := check.New(t)
	c.Equal(map[string]int{"style": 3, "weapon": 1},
		nameable.Uses("A @style@ with a @weapon@, in @style@", "More @style@", "none"))
}

// Apply replaces a marker whose key has a replacement, an empty one included, whatever the order of the marker's
// segments. An unresolved current-form marker, whether or not it allows empty, is shown as just "@Label@" rather than
// its raw markup, keeping the '@' wrapper so it still visibly reads as an unresolved nameable, as is an unresolved
// legacy example list, while an unresolved legacy "Label: text" marker keeps its full text, even when a marker with the
// same label resolves. A marker that does not parse is written back verbatim, and an escaped '@' delimits nothing and
// loses its escape. Rows that share a string apply it resolved before unresolved, since no result may stick to the
// string's cached parse.
func TestApply(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		in           string
		replacements map[string]string
		want         string
	}{
		{in: "@Skill@", replacements: map[string]string{"Skill": "Broadsword"}, want: "Broadsword"},
		{in: "@Skill@", replacements: map[string]string{"Skill": ""}, want: ""},
		{in: "@Skill@", replacements: nil, want: "@Skill@"},
		{
			in:           "Targeted Attack (@Skill@ @Attack@/Neck)",
			replacements: map[string]string{"Skill": "Broadsword", "Attack": "Swung"},
			want:         "Targeted Attack (Broadsword Swung/Neck)",
		},
		{
			in:           "Targeted Attack (@Skill@ @Attack@/Neck)",
			replacements: map[string]string{"Skill": "", "Attack": ""},
			want:         "Targeted Attack ( /Neck)",
		},
		{
			in:           "Targeted Attack (@Skill@ @Attack@/Neck)",
			replacements: map[string]string{"Skill": "Broadsword"},
			want:         "Targeted Attack (Broadsword @Attack@/Neck)",
		},
		{
			in:           "Targeted Attack (@Skill@ @Attack@/Neck)",
			replacements: nil,
			want:         "Targeted Attack (@Skill@ @Attack@/Neck)",
		},
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
			in:           "A @Element|?|Fire|Water@ spell",
			replacements: map[string]string{"Element|Fire|Water|?": "Fire"},
			want:         "A Fire spell",
		},
		// A replacement under the empty key, which literal text carries when compiled, replaces nothing.
		{
			in:           "Another @Element|Fire|Water@ spell, and an @Element|Fire|Water|?@ one too",
			replacements: map[string]string{"Element|Fire|Water": "Fire", "": "unused"},
			want:         "Another Fire spell, and an @Element@ one too",
		},
		{
			in:           "unclosed @Skill",
			replacements: map[string]string{"Skill": "Broadsword", "": "unused"},
			want:         "unclosed @Skill",
		},
		{
			in:           "Patron (@Who: A deity@) and @Who@",
			replacements: map[string]string{"Who": "The King"},
			want:         "Patron (@Who: A deity@) and The King",
		},
		{
			in:           "@Habit: Curtness, Ranting, Scowling, etc.@ and @Habit@",
			replacements: map[string]string{"Habit": "Nail-biting"},
			want:         "@Habit@ and Nail-biting",
		},
		// A leading '|' leaves the label empty, which NewMarker rejects.
		{in: "A @|Fire|Water@ spell", replacements: map[string]string{}, want: "A @|Fire|Water@ spell"},
		{in: `\@Skill@`, replacements: map[string]string{"Skill": "Broadsword"}, want: "@Skill@"},
		{
			in:           `escaped inside @Sk\@ill@`,
			replacements: map[string]string{"Sk@ill": "Escaped"},
			want:         "escaped inside Escaped",
		},
		{in: `escaped inside @Sk\@ill@`, replacements: nil, want: "escaped inside @Sk@ill@"},
		// The key keeps the escape that the label shown for an unresolved marker drops.
		{in: `@Label\|with pipe@`, replacements: map[string]string{`Label\|with pipe`: "Piped"}, want: "Piped"},
		{
			in:           `@Label\|with pipe@`,
			replacements: map[string]string{"Label|with pipe": "Piped"},
			want:         "@Label|with pipe@",
		},
		{in: "ünïcödé @Spéciälty@ ✓", replacements: map[string]string{"Spéciälty": "Feu"}, want: "ünïcödé Feu ✓"},
	} {
		// The second pass applies the compiled text from the cache.
		for range 2 {
			c.Equal(tc.want, nameable.Apply(tc.in, tc.replacements), "%q with %v", tc.in, tc.replacements)
		}
	}
}

// ApplyLinked gives a marker whose key the links name another key for that other key's replacement, whether the marker
// is the whole string or part of it, and leaves an unlinked marker with its own.
func TestApplyLinked(t *testing.T) {
	c := check.New(t)
	replacements := map[string]string{"hand": "Left", "style": "Judo", "weapon": "Sword"}
	links := map[string]string{"weapon": "style"}
	c.Equal("Judo", nameable.ApplyLinked("@weapon@", replacements, links))
	c.Equal("Left: Judo", nameable.ApplyLinked("@hand@: @weapon@", replacements, links))
}

// ApplyToList applies each string in turn, keeping their order, and returns nil for an empty list.
func TestApplyToList(t *testing.T) {
	c := check.New(t)
	replacements := map[string]string{"Skill": "Broadsword"}
	c.Equal([]string{"Broadsword", "No markers here", "@Attack@"},
		nameable.ApplyToList([]string{"@Skill@", "No markers here", "@Attack@"}, replacements))
	c.Nil(nameable.ApplyToList(nil, replacements))
	c.Nil(nameable.ApplyToList([]string{}, replacements))
}

// Reduce keeps the replacements whose key is in nameables, an empty value included, since "empty" and "not set" are
// distinct states, drops those still Unset, and returns nil, not an empty map, when nameables is empty or nothing is
// kept.
func TestReduce(t *testing.T) {
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
		{
			nameables:    map[string]string{"Habit": "Bad Habit"},
			replacements: map[string]string{"Weapon Name": "Excalibur"},
			want:         nil,
		},
	} {
		got := nameable.Reduce(tc.nameables, tc.replacements)
		c.Equal(tc.want, got, "got %#v from nameables %v", got, tc.nameables)
	}
}

// Missing returns the nameable keys the replacements have no entry for, an empty replacement being an entry, which is
// all of them when there are no replacements.
func TestMissing(t *testing.T) {
	c := check.New(t)
	nameables := map[string]string{"Habit": nameable.Unset, "Weapon Name": nameable.Unset}
	for _, tc := range []struct {
		replacements map[string]string
		want         []string
	}{
		{replacements: map[string]string{"Habit": "", "Element": "Fire"}, want: []string{"Weapon Name"}},
		{replacements: nil, want: []string{"Habit", "Weapon Name"}},
		{replacements: map[string]string{"Habit": "Nail-biting", "Weapon Name": "Excalibur"}, want: nil},
	} {
		got := nameable.Missing(nameables, tc.replacements)
		slices.Sort(got)
		if len(tc.want) == 0 {
			c.Equal(0, len(got), "got %v with %v", got, tc.replacements)
		} else {
			c.Equal(tc.want, got, "with %v", tc.replacements)
		}
	}
}

// BenchmarkApply measures replacing the markers in the kinds of strings a sheet's names hold. Run it with:
//
//	GOEXPERIMENT=simd go test ./model/nameable -run '^$' -bench BenchmarkApply -benchmem
func BenchmarkApply(b *testing.B) {
	replacements := map[string]string{"Skill": "Broadsword", "Attack": "Swung", "Specialty": "Fire"}
	for _, one := range []struct {
		name string
		text string
	}{
		{"plain", "Broadsword"},
		{"marker", "Ritual Magic (@Specialty@)"},
		{"multiple", "Targeted Attack (@Skill@ @Attack@/Neck)"},
		{"legacy", "@Skill: Broadsword, Shortsword, etc.@"},
		{"unfilled", "Hobby Skill (@Hobby@)"},
		{"options", "@Skill|Broadsword|Shortsword|*@"},
	} {
		b.Run(one.name, func(b *testing.B) {
			for b.Loop() {
				nameable.Apply(one.text, replacements)
			}
		})
	}
}
