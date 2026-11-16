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

// A marker with several segments takes the first as its whole label, colons included, and the other non-empty ones as
// options, each listed once, except that "?" sets AllowEmpty and "*" sets FreeForm, which is also set when no options
// remain. tt(...) segments anywhere in it give its tooltip, joined by line breaks, keeping colons and turning `\|`,
// `\n` and `\\` into '|', a line break and '\'. `\|` also unescapes to a literal '|' in the label and in any option,
// the last one included, while `\n` outside a tooltip keeps its two literal characters.
func TestNewMarkerWithSegments(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		raw        string
		label      string
		options    []string
		tooltip    string
		allowEmpty bool
		freeForm   bool
	}{
		{raw: "Element|Fire|Water|Earth", label: "Element", options: []string{"Fire", "Water", "Earth"}},
		// A stray "||" (e.g. from hand-edited data) or a trailing '|' is dropped rather than becoming a blank option.
		{raw: "Element||Fire||Water|", label: "Element", options: []string{"Fire", "Water"}},
		{raw: "Element|Fire|Water|Fire", label: "Element", options: []string{"Fire", "Water"}},
		// The legacy "Label: text" forms apply only to a single segment.
		{
			raw:     "Time: HH:MM|Morning|Afternoon|Evening",
			label:   "Time: HH:MM",
			options: []string{"Morning", "Afternoon", "Evening"},
		},
		{
			raw:        "Element|Fire|Water|?|*",
			label:      "Element",
			options:    []string{"Fire", "Water"},
			allowEmpty: true,
			freeForm:   true,
		},
		{raw: "Element|?", label: "Element", allowEmpty: true, freeForm: true},
		{raw: "Element|*", label: "Element", freeForm: true},
		{
			raw:     "Element|tt(Pick the affinity)|Fire|Water",
			label:   "Element",
			options: []string{"Fire", "Water"},
			tooltip: "Pick the affinity",
		},
		{
			raw:     "Element|tt(Ratio is 1:2)|Fire|Water",
			label:   "Element",
			options: []string{"Fire", "Water"},
			tooltip: "Ratio is 1:2",
		},
		{
			raw:        "Element|Fire|tt(Pick one)|Water|?",
			label:      "Element",
			options:    []string{"Fire", "Water"},
			tooltip:    "Pick one",
			allowEmpty: true,
		},
		// The form Key gives a marker with a tooltip, as read back from a saved key.
		{raw: "Element|Fire|tt(Choose)", label: "Element", options: []string{"Fire"}, tooltip: "Choose"},
		{
			raw:     "Element|tt(First line)|tt(Second line)|Fire",
			label:   "Element",
			options: []string{"Fire"},
			tooltip: "First line\nSecond line",
		},
		{
			raw:     `Element|tt(Choose one\|Two)|Fire`,
			label:   "Element",
			options: []string{"Fire"},
			tooltip: "Choose one|Two",
		},
		{
			raw:     `Element|tt(Line one\nLine two)|Fire`,
			label:   "Element",
			options: []string{"Fire"},
			tooltip: "Line one\nLine two",
		},
		// The tooltip is unescaped in a single pass, so an escaped backslash followed by a literal 'n' -- as in a
		// Windows-style path -- stays a backslash rather than collapsing into a line break.
		{raw: `Element|tt(Path\\name)|Fire`, label: "Element", options: []string{"Fire"}, tooltip: `Path\name`},
		{raw: `Fire\|Water|Hot|Cold`, label: "Fire|Water", options: []string{"Hot", "Cold"}},
		{raw: `Element|Hot\|Cold|Water`, label: "Element", options: []string{"Hot|Cold", "Water"}},
		// ExtractSegments appends the segment after the last '|' once its loop ends.
		{raw: `Element|Water|Hot\|Cold`, label: "Element", options: []string{"Water", "Hot|Cold"}},
		{raw: `Multi\nLine|Fire|Water`, label: `Multi\nLine`, options: []string{"Fire", "Water"}},
	} {
		marker, ok := nameable.NewMarker(tc.raw)
		c.True(ok, tc.raw)
		c.Equal(tc.raw, marker.Raw, tc.raw)
		c.False(marker.Legacy, tc.raw)
		c.Equal(tc.label, marker.Label, tc.raw)
		if len(tc.options) == 0 {
			c.Equal(0, len(marker.Options), tc.raw)
		} else {
			c.Equal(tc.options, marker.Options, tc.raw)
		}
		c.Equal(tc.tooltip, marker.Tooltip, tc.raw)
		c.Equal(tc.allowEmpty, marker.AllowEmpty, tc.raw)
		c.Equal(tc.freeForm, marker.FreeForm, tc.raw)
	}
}

// A single segment never fails to parse and always allows empty and free-form input, with its trimmed text as Raw.
// "Label: item, item[, item...], etc.", with or without the final period, is a legacy marker with its items, each
// listed once, as options and the text after the colon as its tooltip. Any other "Label: text" is a legacy marker with
// that text as its tooltip and no options. Either needs a label of at most 40 characters before the colon. Anything
// else is a current-form marker labeled with its whole text, with no tooltip or options.
func TestNewMarkerSingleSegment(t *testing.T) {
	c := check.New(t)
	label40 := strings.Repeat("A", 40)
	label41 := strings.Repeat("A", 41)
	for _, tc := range []struct {
		raw     string
		legacy  bool
		label   string
		options []string
		tooltip string
	}{
		// Real examples pulled from the GCS master library scan.
		{
			raw:    "Rare: Acceleration, Altitude Sickness, Bends, Seasickness, Space Sickness, Nanomachines, etc.",
			legacy: true,
			label:  "Rare",
			options: []string{
				"Acceleration", "Altitude Sickness", "Bends", "Seasickness", "Space Sickness", "Nanomachines",
			},
			tooltip: "Acceleration, Altitude Sickness, Bends, Seasickness, Space Sickness, Nanomachines, etc.",
		},
		{
			raw:     "Skill: Fast-Draw, Smith, etc",
			legacy:  true,
			label:   "Skill",
			options: []string{"Fast-Draw", "Smith"},
			tooltip: "Fast-Draw, Smith, etc",
		},
		{
			raw:     "Skill: Smith, Fast-Draw, Smith, etc.",
			legacy:  true,
			label:   "Skill",
			options: []string{"Smith", "Fast-Draw"},
			tooltip: "Smith, Fast-Draw, Smith, etc.",
		},
		{raw: "Class: Mammalia", legacy: true, label: "Class", tooltip: "Mammalia"},
		// A single item before "etc." isn't a strong enough signal for an example list.
		{raw: "Absurd: chocolate, etc.", legacy: true, label: "Absurd", tooltip: "chocolate, etc."},
		// Items with no "etc." read as a closed, exhaustive set (e.g. a fixed enum), not as examples.
		{
			raw:     "Type: Mental, Physical, Magical or Chi",
			legacy:  true,
			label:   "Type",
			tooltip: "Mental, Physical, Magical or Chi",
		},
		// A comma inside a quoted phrase throws off the item splitting of an example list, so this safely reads as
		// plain "Label: text" rather than being mis-split into garbage options.
		{
			raw:     `Habit: Witless witticisms, calling people "nuncle," abuse of custard, etc.`,
			legacy:  true,
			label:   "Habit",
			tooltip: `Witless witticisms, calling people "nuncle," abuse of custard, etc.`,
		},
		// A label of exactly 40 characters is still a legacy label, while one of 41 is not, so the whole text becomes
		// the label of a current-form marker.
		{raw: label40 + ": tail", legacy: true, label: label40, tooltip: "tail"},
		{raw: label41 + ": tail", label: label41 + ": tail"},
		{
			raw:     label40 + ": B, C, etc.",
			legacy:  true,
			label:   label40,
			options: []string{"B", "C"},
			tooltip: "B, C, etc.",
		},
		{raw: label41 + ": B, C, etc.", label: label41 + ": B, C, etc."},
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
		c.Equal(strings.TrimSpace(tc.raw), marker.Raw, tc.raw)
		c.Equal(tc.legacy, marker.Legacy, tc.raw)
		c.Equal(tc.label, marker.Label, tc.raw)
		if len(tc.options) == 0 {
			c.Equal(0, len(marker.Options), tc.raw)
		} else {
			c.Equal(tc.options, marker.Options, tc.raw)
		}
		c.Equal(tc.tooltip, marker.Tooltip, tc.raw)
		c.True(marker.AllowEmpty, tc.raw)
		c.True(marker.FreeForm, tc.raw)
	}
}

// A marker with an empty label, or with no text at all, fails to parse.
func TestNewMarkerFailsWithoutLabel(t *testing.T) {
	c := check.New(t)
	for _, raw := range []string{"|Fire|Water", ""} {
		_, ok := nameable.NewMarker(raw)
		c.False(ok, "%q", raw)
	}
}

// A marker keys on its label, options, '*' and '?', in that order whatever their order in its text, so markers that
// differ only in segment order share a key. A bare label keys as itself and a legacy marker on its raw text, so the
// replacements of sheets saved before keys were normalized keep resolving.
func TestKeyNormalization(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		raw  string
		want string
	}{
		{raw: "Element|Fire|Water|?|*", want: "Element|Fire|Water|*|?"},
		{raw: "Element|Fire|Water", want: "Element|Fire|Water"},
		{raw: "Element|Fire|Water|?", want: "Element|Fire|Water|?"},
		{raw: "Element|?|Fire|Water", want: "Element|Fire|Water|?"},
		{raw: "Weapon Name", want: "Weapon Name"},
		// Real, unmodified markers from the GCS master library, whatever options and tooltip the parse took from them.
		{
			raw:  "Rare: Acceleration, Altitude Sickness, Bends, Seasickness, Space Sickness, Nanomachines, etc.",
			want: "Rare: Acceleration, Altitude Sickness, Bends, Seasickness, Space Sickness, Nanomachines, etc.",
		},
		{raw: "Class: Mammalia", want: "Class: Mammalia"},
	} {
		m, ok := nameable.NewMarker(tc.raw)
		c.True(ok, tc.raw)
		c.Equal(tc.want, m.Key(), tc.raw)
	}
}

// Key, tested apart from NewMarker's unescaping on markers built directly, gives the tooltip its own segment after the
// label or last option, leaves the tooltip's parens unescaped, since "tt(" and ")" are stripped as a fixed wrapper, and
// escapes '|' and '\' in the label and options, and those and line breaks in the tooltip.
func TestKeyTooltipAndEscaping(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		marker nameable.Marker
		want   string
	}{
		{marker: nameable.Marker{Label: "Element", Tooltip: "Choose"}, want: "Element|tt(Choose)"},
		{
			marker: nameable.Marker{Label: "Element", Options: []string{"Fire"}, Tooltip: "Choose"},
			want:   "Element|Fire|tt(Choose)",
		},
		{marker: nameable.Marker{Label: "Element", Tooltip: "a (b) c"}, want: "Element|tt(a (b) c)"},
		{marker: nameable.Marker{Label: "Element", Tooltip: "a|b\\c\nd"}, want: `Element|tt(a\|b\\c\nd)`},
		{marker: nameable.Marker{Label: "A|B", Options: []string{`C\D`}}, want: `A\|B|C\\D`},
	} {
		c.Equal(tc.want, tc.marker.Key())
	}
}
