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
// options, except that "?" sets AllowEmpty and "*" sets FreeForm, which is also set when no options remain.
func TestParseMarkerBasic(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		raw        string
		label      string
		options    []string
		allowEmpty bool
		freeForm   bool
	}{
		{raw: "Element|Fire|Water|Earth", label: "Element", options: []string{"Fire", "Water", "Earth"}},
		// A stray "||" (e.g. from hand-edited data) or a trailing '|' is dropped rather than becoming a blank option.
		{raw: "Element||Fire||Water|", label: "Element", options: []string{"Fire", "Water"}},
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
	} {
		marker, ok := nameable.NewMarker(tc.raw)
		c.True(ok, tc.raw)
		c.Equal(tc.raw, marker.Raw, tc.raw)
		c.Equal(tc.label, marker.Label, tc.raw)
		c.Equal("", marker.Tooltip, tc.raw)
		if len(tc.options) == 0 {
			c.Equal(0, len(marker.Options), tc.raw)
		} else {
			c.Equal(tc.options, marker.Options, tc.raw)
		}
		c.Equal(tc.allowEmpty, marker.AllowEmpty, tc.raw)
		c.Equal(tc.freeForm, marker.FreeForm, tc.raw)
	}
}

// tt(...) segments anywhere in a marker give its tooltip, joined by line breaks, keeping colons and turning `\|`, `\n`
// and `\\` into '|', a line break and '\', while the other segments still give the options and flags.
func TestParseMarkerWithTooltip(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		raw        string
		tooltip    string
		options    []string
		allowEmpty bool
	}{
		{
			raw:     "Element|tt(Pick the affinity)|Fire|Water",
			tooltip: "Pick the affinity",
			options: []string{"Fire", "Water"},
		},
		{raw: "Element|tt(Ratio is 1:2)|Fire|Water", tooltip: "Ratio is 1:2", options: []string{"Fire", "Water"}},
		{
			raw:        "Element|Fire|tt(Pick one)|Water|?",
			tooltip:    "Pick one",
			options:    []string{"Fire", "Water"},
			allowEmpty: true,
		},
		{
			raw:     "Element|tt(First line)|tt(Second line)|Fire",
			tooltip: "First line\nSecond line",
			options: []string{"Fire"},
		},
		{raw: `Element|tt(Choose one\|Two)|Fire`, tooltip: "Choose one|Two", options: []string{"Fire"}},
		{raw: `Element|tt(Line one\nLine two)|Fire`, tooltip: "Line one\nLine two", options: []string{"Fire"}},
		// The tooltip is unescaped in a single pass, so an escaped backslash followed by a literal 'n' -- as in a
		// Windows-style path -- stays a backslash rather than collapsing into a line break.
		{raw: `Element|tt(Path\\name)|Fire`, tooltip: `Path\name`, options: []string{"Fire"}},
	} {
		marker, ok := nameable.NewMarker(tc.raw)
		c.True(ok, tc.raw)
		c.Equal("Element", marker.Label, tc.raw)
		c.Equal(tc.tooltip, marker.Tooltip, tc.raw)
		c.Equal(tc.options, marker.Options, tc.raw)
		c.Equal(tc.allowEmpty, marker.AllowEmpty, tc.raw)
	}
}

// `\|` unescapes to a literal '|' in the label and in any option, the last one included, while `\n` outside a tooltip
// keeps its two literal characters.
func TestParseMarkerEscapedPipeInLabel(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		raw     string
		label   string
		options []string
	}{
		{raw: `Fire\|Water|Hot|Cold`, label: "Fire|Water", options: []string{"Hot", "Cold"}},
		{raw: `Element|Hot\|Cold|Water`, label: "Element", options: []string{"Hot|Cold", "Water"}},
		// ExtractSegments appends the segment after the last '|' once its loop ends.
		{raw: `Element|Water|Hot\|Cold`, label: "Element", options: []string{"Water", "Hot|Cold"}},
		{raw: `Multi\nLine|Fire|Water`, label: `Multi\nLine`, options: []string{"Fire", "Water"}},
	} {
		marker, ok := nameable.NewMarker(tc.raw)
		c.True(ok, tc.raw)
		c.Equal(tc.label, marker.Label, tc.raw)
		c.Equal(tc.options, marker.Options, tc.raw)
	}
}

// A marker with an empty label, or with no text at all, fails to parse.
func TestParseMarkerEmptyLabelFails(t *testing.T) {
	c := check.New(t)
	for _, raw := range []string{"|Fire|Water", ""} {
		_, ok := nameable.NewMarker(raw)
		c.False(ok, "%q", raw)
	}
}

// A marker keys on its label, options, '*' and '?', in that order whatever their order in its text, so markers that
// differ only in segment order share a key. A bare label keys as itself and a legacy marker on its raw text, so the
// replacements of sheets saved before keys were normalized keep resolving.
func TestKeyIncludesAllowEmptyAndFreeFormTokens(t *testing.T) {
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
// escapes '|' and '\' in the label and options.
func TestKeyTooltipIsSeparatedFromPrecedingContentBySegmentDelimiter(t *testing.T) {
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
		{marker: nameable.Marker{Label: "A|B", Options: []string{`C\D`}}, want: `A\|B|C\\D`},
	} {
		c.Equal(tc.want, tc.marker.Key())
	}
}

// Feeding a key back through NewMarker -- exactly what happens when a key is persisted and reloaded -- must keep the
// tooltip rather than absorbing it into the preceding option.
func TestKeyRoundTripsLosslessThroughNewMarker(t *testing.T) {
	c := check.New(t)
	m, ok := nameable.NewMarker("Element|tt(Choose)|Fire")
	c.True(ok)
	reparsed, ok := nameable.NewMarker(m.Key())
	c.True(ok)
	c.Equal([]string{"Fire"}, reparsed.Options)
	c.Equal("Choose", reparsed.Tooltip)
}

// Extract records each marker under its key, in the map it is given or a new one when that is nil, with the value the
// replacements hold for that key, or Unset when they hold none.
func TestExtractPipeDelimited(t *testing.T) {
	c := check.New(t)
	m := make(map[string]string)
	nameable.Extract(m, nil, "A @Element|Fire|Water@ spell")
	c.Equal(nameable.Unset, m["Element|Fire|Water"])

	m = make(map[string]string)
	nameable.Extract(m, nil, "A @Element|Fire|Water|?@ spell")
	c.Equal(nameable.Unset, m["Element|Fire|Water|?"])

	m = nameable.Extract(nil, nil, "A @Element|Fire|Water@ spell")
	c.Equal(nameable.Unset, m["Element|Fire|Water"])

	m = nameable.Extract(nil, map[string]string{"Element|Fire|Water": "Water"}, "A @Element|Fire|Water@ spell")
	c.Equal("Water", m["Element|Fire|Water"])
}

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
