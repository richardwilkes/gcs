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
