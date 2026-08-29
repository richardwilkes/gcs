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

func TestParseMarkerPlainKeyIsCurrentFormNotLegacy(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.NewMarker("Weapon Name")
	c.True(ok)
	c.False(marker.Legacy)
	c.Equal("Weapon Name", marker.Label)
}

func TestParseMarkerBasic(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.NewMarker("Element|Fire|Water|Earth")
	c.True(ok)
	c.Equal("Element|Fire|Water|Earth", marker.Raw)
	c.Equal("Element", marker.Label)
	c.Equal("", marker.Tooltip)
	c.Equal([]string{"Fire", "Water", "Earth"}, marker.Options)
	c.False(marker.AllowEmpty)
	c.False(marker.FreeForm)
}

func TestParseMarkerSkipsEmptyNonLabelSegments(t *testing.T) {
	c := check.New(t)
	// A stray "||" (e.g. from hand-edited data, or a trailing "|" before the closing '@') shouldn't surface as a
	// blank option, tooltip, or token -- it's silently dropped rather than treated as ErrEmptyLabel.
	marker, ok := nameable.NewMarker("Element||Fire||Water|")
	c.True(ok)
	c.Equal("Element", marker.Label)
	c.Equal([]string{"Fire", "Water"}, marker.Options)
}

func TestParseMarkerWithTooltip(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.NewMarker("Element|tt(Pick the affinity)|Fire|Water")
	c.True(ok)
	c.Equal("Element", marker.Label)
	c.Equal("Pick the affinity", marker.Tooltip)
	c.Equal([]string{"Fire", "Water"}, marker.Options)
}

func TestParseMarkerLabelMayContainColon(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.NewMarker("Time: HH:MM|Morning|Afternoon|Evening")
	c.True(ok)
	c.Equal("Time: HH:MM", marker.Label)
	c.Equal("", marker.Tooltip)
	c.Equal([]string{"Morning", "Afternoon", "Evening"}, marker.Options)
}

func TestParseMarkerTooltipMayContainColon(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.NewMarker("Element|tt(Ratio is 1:2)|Fire|Water")
	c.True(ok)
	c.Equal("Element", marker.Label)
	c.Equal("Ratio is 1:2", marker.Tooltip)
}

func TestParseMarkerTooltipAnywhereInSegments(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.NewMarker("Element|Fire|tt(Pick one)|Water|?")
	c.True(ok)
	c.Equal("Pick one", marker.Tooltip)
	c.Equal([]string{"Fire", "Water"}, marker.Options)
	c.True(marker.AllowEmpty)
}

func TestParseMarkerMultipleTooltipLinesJoined(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.NewMarker("Element|tt(First line)|tt(Second line)|Fire")
	c.True(ok)
	c.Equal("First line\nSecond line", marker.Tooltip)
	c.Equal([]string{"Fire"}, marker.Options)
}

func TestParseMarkerEscapedPipeInLabel(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.NewMarker(`Fire\|Water|Hot|Cold`)
	c.True(ok)
	c.Equal("Fire|Water", marker.Label)
	c.Equal([]string{"Hot", "Cold"}, marker.Options)
}

func TestParseMarkerEscapedPipeInOption(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.NewMarker(`Element|Hot\|Cold|Water`)
	c.True(ok)
	c.Equal([]string{"Hot|Cold", "Water"}, marker.Options)
}

func TestParseMarkerEscapedPipeInTrailingOption(t *testing.T) {
	c := check.New(t)
	// The trailing segment (after the last '|', or the whole string when there's no '|' at all) must be unescaped
	// and trimmed the same as every other segment.
	marker, ok := nameable.NewMarker(`Element|Water|Hot\|Cold`)
	c.True(ok)
	c.Equal([]string{"Water", "Hot|Cold"}, marker.Options)
}

func TestParseMarkerTrimsWhitespaceInBareLabel(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.NewMarker("  Weapon Name  ")
	c.True(ok)
	c.False(marker.Legacy)
	c.Equal("Weapon Name", marker.Label)
}

func TestParseMarkerEscapedPipeInTooltip(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.NewMarker(`Element|tt(Choose one\|Two)|Fire`)
	c.True(ok)
	c.Equal("Choose one|Two", marker.Tooltip)
	c.Equal([]string{"Fire"}, marker.Options)
}

func TestParseMarkerEscapedNewlineInTooltip(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.NewMarker(`Element|tt(Line one\nLine two)|Fire`)
	c.True(ok)
	c.Equal("Line one\nLine two", marker.Tooltip)
}

func TestParseMarkerLiteralBackslashNInLabelIsNotANewline(t *testing.T) {
	c := check.New(t)
	// The '\n' escape-to-newline conversion is only recognized inside a tooltip segment. A label containing the
	// literal two-character sequence '\n' keeps it as-is instead of turning it into a real newline.
	marker, ok := nameable.NewMarker(`Multi\nLine|Fire|Water`)
	c.True(ok)
	c.Equal(`Multi\nLine`, marker.Label)
}

func TestParseMarkerTooltipEscapedBackslashFollowedByNIsNotNewline(t *testing.T) {
	c := check.New(t)
	// The tooltip segment unescapes "\\" to "\" before converting a literal "\n" to a real newline, so an escaped
	// backslash that happens to be followed by a literal 'n' -- as in a Windows-style path -- collapses into a line
	// break and eats the 'n', instead of staying the single backslash the author escaped.
	marker, ok := nameable.NewMarker(`Element|tt(Path\\name)|Fire`)
	c.True(ok)
	c.Equal(`Path\name`, marker.Tooltip)
}

func TestParseMarkerAllowEmptyAndFreeForm(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.NewMarker("Element|Fire|Water|?|*")
	c.True(ok)
	c.True(marker.AllowEmpty)
	c.True(marker.FreeForm)
	c.Equal([]string{"Fire", "Water"}, marker.Options)
}

func TestParseMarkerFreeFormOnlyNoOptions(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.NewMarker("Element|*")
	c.True(ok)
	c.True(marker.FreeForm)
	c.Equal(0, len(marker.Options))
}

func TestParseMarkerEmptyLabelFails(t *testing.T) {
	c := check.New(t)
	_, ok := nameable.NewMarker("|Fire|Water")
	c.False(ok)
}

func TestParseMarkerNoOptionsImpliesFreeForm(t *testing.T) {
	c := check.New(t)
	// No literal options and no explicit FreeFormToken -- there's nothing else it could mean, so FreeForm is turned
	// on automatically instead of failing.
	marker, ok := nameable.NewMarker("Element|?")
	c.True(ok)
	c.True(marker.FreeForm)
	c.True(marker.AllowEmpty)
	c.Equal(0, len(marker.Options))
}

func TestParseMarkerNoPipeIsCurrentFormNotLegacy(t *testing.T) {
	c := check.New(t)
	// No space after the colons, so neither legacy regex matches -- falls to the tier-3 current form.
	marker, ok := nameable.NewMarker("Element:Fire:Water")
	c.True(ok)
	c.False(marker.Legacy)
}

func TestParseMarkerSingleSegmentIsCurrentFormNotLegacy(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.NewMarker("Element")
	c.True(ok)
	c.False(marker.Legacy)
	c.Equal("Element", marker.Label)
}

func TestDefaultValue(t *testing.T) {
	c := check.New(t)

	marker, ok := nameable.NewMarker("Weapon Name")
	c.True(ok)
	c.Equal("Weapon Name", marker.DefaultValue())

	marker, ok = nameable.NewMarker("Element|Fire|Water")
	c.True(ok)
	c.Equal("Fire", marker.DefaultValue())

	marker, ok = nameable.NewMarker("Element|Fire|Water|?")
	c.True(ok)
	c.Equal("", marker.DefaultValue())
}

func TestKeyPlainLegacyLabel(t *testing.T) {
	c := check.New(t)
	m, ok := nameable.NewMarker("Weapon Name")
	c.True(ok)
	// A bare, single-segment legacy marker short-circuits to just its (escaped) label -- no '*'/'?' tokens -- so
	// that a plain "@Weapon Name@" keeps the same Replacements key it always has, rather than becoming
	// "Weapon Name|*|?" and silently orphaning every character sheet saved before that normalization existed.
	c.Equal("Weapon Name", m.Key())
}

func TestKeyLegacyExampleListMarkerDoesNotShortCircuit(t *testing.T) {
	c := check.New(t)
	// A real, unmodified marker pulled from the GCS master library (the Resistant trait). legacyExampleListPattern
	// gives it both Options and a Tooltip, so it must NOT hit the plain-legacy short circuit in Key() -- that only
	// applies when a legacy marker has neither. It gets the full options/flags/tooltip key like any new-format
	// marker would.
	raw := "Rare: Acceleration, Altitude Sickness, Bends, Seasickness, Space Sickness, Nanomachines, etc."
	m, ok := nameable.NewMarker(raw)
	c.True(ok)
	c.True(m.Legacy)
	c.Equal(
		"Rare|Acceleration|Altitude Sickness|Bends|Seasickness|Space Sickness|Nanomachines|*|?|"+
			"tt(Acceleration, Altitude Sickness, Bends, Seasickness, Space Sickness, Nanomachines, etc.)",
		m.Key(),
	)
}

func TestKeyLegacyLabeledMarkerDoesNotShortCircuit(t *testing.T) {
	c := check.New(t)
	// A real, unmodified marker pulled from the GCS master library. legacyLabeledPattern gives it a Tooltip (the
	// text after the colon) with no Options, so -- same as the example-list case above -- it must not hit the
	// plain-legacy short circuit, since that requires an empty Tooltip too.
	m, ok := nameable.NewMarker("Class: Mammalia")
	c.True(ok)
	c.True(m.Legacy)
	c.Equal("Class|*|?|tt(Mammalia)", m.Key())
}

func TestKeyLabelWithOptionsNoFlags(t *testing.T) {
	c := check.New(t)
	m, ok := nameable.NewMarker("Element|Fire|Water")
	c.True(ok)
	c.Equal("Element|Fire|Water", m.Key())
}

func TestKeyIncludesAllowEmptyAndFreeFormTokens(t *testing.T) {
	c := check.New(t)
	m, ok := nameable.NewMarker("Element|Fire|Water|?|*")
	c.True(ok)
	// Tokens are written in a fixed order -- label, then options, then '*', then '?' -- regardless of where they
	// appeared in raw, each as its own pipe-delimited segment.
	c.Equal("Element|Fire|Water|*|?", m.Key())
}

func TestKeyNormalizesEquivalentMarkersToTheSameKey(t *testing.T) {
	c := check.New(t)
	// This is the whole point of Key(): two raw marker texts that differ only in segment order are semantically the
	// same marker and must collapse to one entry in ExtractMarkers and one lookup slot in Extract/Apply.
	a, ok := nameable.NewMarker("Element|Fire|Water|?")
	c.True(ok)
	b, ok := nameable.NewMarker("Element|?|Fire|Water")
	c.True(ok)
	c.Equal(a.Key(), b.Key())
	c.Equal("Element|Fire|Water|?", a.Key())
}

func TestKeyEscapesSegmentDelimiterAndEscapeRuneInLabelAndOptions(t *testing.T) {
	c := check.New(t)
	// Constructed directly rather than via NewMarker, to test Key()'s own escaping in isolation from whatever
	// NewMarker's parse-time unescaping does with the source text.
	m := nameable.Marker{Label: "A|B", Options: []string{`C\D`}}
	c.Equal(`A\|B|C\\D`, m.Key())
}

func TestKeyDoesNotEscapeParensInTooltip(t *testing.T) {
	c := check.New(t)
	// The tooltip segment is delimited by the '|' segment delimiter, not by matching parens -- the "tt(" prefix and
	// ")" suffix are stripped as a fixed-length wrapper around the whole segment, so inner parens are inert and
	// don't need escaping.
	m := nameable.Marker{Label: "Element", Tooltip: "a (b) c"}
	c.Contains(m.Key(), "a (b) c")
}

func TestKeyTooltipIsSeparatedFromPrecedingContentBySegmentDelimiter(t *testing.T) {
	c := check.New(t)
	// Regression test: Key() used to write the tooltip's "tt(" prefix with no leading '|', gluing it directly onto
	// whatever came before it -- the label when there are no options, or the last option when there are. Fixed now.
	m1 := nameable.Marker{Label: "Element", Tooltip: "Choose"}
	c.Equal("Element|tt(Choose)", m1.Key())
	m2 := nameable.Marker{Label: "Element", Options: []string{"Fire"}, Tooltip: "Choose"}
	c.Equal("Element|Fire|tt(Choose)", m2.Key())
}

func TestKeyRoundTripsLosslessThroughNewMarker(t *testing.T) {
	c := check.New(t)
	// Regression test: before the tooltip-separator fix above, the produced key wasn't valid Marker syntax (the
	// tooltip segment wasn't actually a separate segment), so feeding it back through NewMarker -- exactly what
	// happens when a key is persisted and reloaded -- silently absorbed the tooltip into the preceding option and
	// lost it. Fixed now: the round trip is lossless.
	m, ok := nameable.NewMarker("Element|tt(Choose)|Fire")
	c.True(ok)
	reparsed, ok := nameable.NewMarker(m.Key())
	c.True(ok)
	c.Equal([]string{"Fire"}, reparsed.Options)
	c.Equal("Choose", reparsed.Tooltip)
}

func TestExtractPipeDelimited(t *testing.T) {
	c := check.New(t)
	m := make(map[string]string)
	nameable.Extract(m, nil, "A @Element|Fire|Water@ spell")
	c.Equal("Fire", m["Element|Fire|Water"])

	m = make(map[string]string)
	nameable.Extract(m, nil, "A @Element|Fire|Water|?@ spell")
	c.Equal("", m["Element|Fire|Water|?"])
}

func TestExtractWithNilTargetCreatesMap(t *testing.T) {
	c := check.New(t)
	got := nameable.Extract(nil, nil, "A @Element|Fire|Water@ spell")
	c.Equal("Fire", got["Element|Fire|Water"])
}

func TestExtractUsesExistingValueWhenKeyMatches(t *testing.T) {
	c := check.New(t)
	existing := map[string]string{"Element|Fire|Water": "Water"}
	got := nameable.Extract(nil, existing, "A @Element|Fire|Water@ spell")
	c.Equal("Water", got["Element|Fire|Water"])
}

func TestExtractMarkersEscapedBackslashBeforeAtIsNotEscapedAt(t *testing.T) {
	c := check.New(t)
	// Two literal backslashes immediately before '@' form one escaped-backslash pair (per the same pairing rule
	// splitSegments uses for '|'), leaving the '@' itself unescaped and free to open a marker.
	markers := nameable.ExtractMarkers(`A\\@Element|Fire|Water@ B`)
	c.Equal(1, len(markers))
	_, ok := markers["Element|Fire|Water"]
	c.True(ok)
}

func TestReduceKeepsAllowEmptyEmptyValue(t *testing.T) {
	c := check.New(t)
	// "Empty" (present, value "") and "not set" (key absent) are distinct states -- Reduce must not conflate them by
	// dropping a deliberately-stored empty value.
	needed := map[string]string{"Element|Fire|Water|?": ""}
	replacements := map[string]string{"Element|Fire|Water|?": ""}
	reduced := nameable.Reduce(needed, replacements)
	v, has := reduced["Element|Fire|Water|?"]
	c.True(has)
	c.Equal("", v)
}

func TestReduceOmitsKeyNotInNeeded(t *testing.T) {
	c := check.New(t)
	// "Not set" is represented purely by absence from replacements to begin with -- Reduce's only remaining job is
	// pruning orphaned entries whose marker no longer appears in needed (e.g. after the source text changed).
	needed := map[string]string{}
	replacements := map[string]string{"Element|Fire|Water|?": "Fire"}
	reduced := nameable.Reduce(needed, replacements)
	_, has := reduced["Element|Fire|Water|?"]
	c.False(has)
}

func TestReduceKeepsRequiredEmptyValue(t *testing.T) {
	c := check.New(t)
	// Reduce is a pure key-membership filter -- it does no normalization of its own, so needed and replacements
	// must already agree on key format. A bare legacy label like "Weapon Name" keeps its plain key (see
	// TestKeyPlainLegacyLabel), matching what's stored for any character sheet saved before Marker.Key() existed.
	needed := map[string]string{"Weapon Name": ""}
	replacements := map[string]string{"Weapon Name": ""}
	reduced := nameable.Reduce(needed, replacements)
	_, has := reduced["Weapon Name"]
	c.True(has)
}

func TestReduceKeepsAllowEmptyChosenValue(t *testing.T) {
	c := check.New(t)
	needed := map[string]string{"Element|Fire|Water|?": "Fire"}
	replacements := map[string]string{"Element|Fire|Water|?": "Fire"}
	reduced := nameable.Reduce(needed, replacements)
	c.Equal("Fire", reduced["Element|Fire|Water|?"])
}

func TestReduceRetainsMalformedReplacementKeyButHasNoEffect(t *testing.T) {
	c := check.New(t)
	// A replacements key that itself fails to parse as a marker (here, "") is retained as-is rather than dropped,
	// but simply never matches any real marker's key, so it has no effect on the result here.
	needed := map[string]string{"Element|Fire|Water": ""}
	replacements := map[string]string{"": "unused", "Element|Fire|Water": "Fire"}
	reduced := nameable.Reduce(needed, replacements)
	c.Equal(1, len(reduced))
	c.Equal("Fire", reduced["Element|Fire|Water"])
}
