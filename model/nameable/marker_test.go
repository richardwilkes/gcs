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

func TestParseMarkerPlainKeyFallsBackToLegacy(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker("Weapon Name")
	c.True(ok)
	c.True(marker.Legacy)
	c.Equal("Weapon Name", marker.Label)
}

func TestParseMarkerBasic(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker("Element|Fire|Water|Earth")
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
	marker, ok := nameable.ParseMarker("Element||Fire||Water|")
	c.True(ok)
	c.Equal("Element", marker.Label)
	c.Equal([]string{"Fire", "Water"}, marker.Options)
}

func TestParseMarkerWithTooltip(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker("Element|tt(Pick the affinity)|Fire|Water")
	c.True(ok)
	c.Equal("Element", marker.Label)
	c.Equal("Pick the affinity", marker.Tooltip)
	c.Equal([]string{"Fire", "Water"}, marker.Options)
}

func TestParseMarkerLabelMayContainColon(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker("Time: HH:MM|Morning|Afternoon|Evening")
	c.True(ok)
	c.Equal("Time: HH:MM", marker.Label)
	c.Equal("", marker.Tooltip)
	c.Equal([]string{"Morning", "Afternoon", "Evening"}, marker.Options)
}

func TestParseMarkerTooltipMayContainColon(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker("Element|tt(Ratio is 1:2)|Fire|Water")
	c.True(ok)
	c.Equal("Element", marker.Label)
	c.Equal("Ratio is 1:2", marker.Tooltip)
}

func TestParseMarkerTooltipAnywhereInSegments(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker("Element|Fire|tt(Pick one)|Water|?")
	c.True(ok)
	c.Equal("Pick one", marker.Tooltip)
	c.Equal([]string{"Fire", "Water"}, marker.Options)
	c.True(marker.AllowEmpty)
}

func TestParseMarkerMultipleTooltipLinesJoined(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker("Element|tt(First line)|tt(Second line)|Fire")
	c.True(ok)
	c.Equal("First line\nSecond line", marker.Tooltip)
	c.Equal([]string{"Fire"}, marker.Options)
}

func TestParseMarkerEscapedPipeInLabel(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker(`Fire\|Water|Hot|Cold`)
	c.True(ok)
	c.Equal("Fire|Water", marker.Label)
	c.Equal([]string{"Hot", "Cold"}, marker.Options)
}

func TestParseMarkerEscapedPipeInOption(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker(`Element|Hot\|Cold|Water`)
	c.True(ok)
	c.Equal([]string{"Hot|Cold", "Water"}, marker.Options)
}

func TestParseMarkerEscapedPipeInTrailingOption(t *testing.T) {
	c := check.New(t)
	// The trailing segment (after the last '|', or the whole string when there's no '|' at all) must be unescaped
	// and trimmed the same as every other segment.
	marker, ok := nameable.ParseMarker(`Element|Water|Hot\|Cold`)
	c.True(ok)
	c.Equal([]string{"Water", "Hot|Cold"}, marker.Options)
}

func TestParseMarkerTrimsWhitespaceInBareLabel(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker("  Weapon Name  ")
	c.True(ok)
	c.True(marker.Legacy)
	c.Equal("Weapon Name", marker.Label)
}

func TestParseMarkerEscapedPipeInTooltip(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker(`Element|tt(Choose one\|Two)|Fire`)
	c.True(ok)
	c.Equal("Choose one|Two", marker.Tooltip)
	c.Equal([]string{"Fire"}, marker.Options)
}

func TestParseMarkerEscapedNewlineInTooltip(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker(`Element|tt(Line one\nLine two)|Fire`)
	c.True(ok)
	c.Equal("Line one\nLine two", marker.Tooltip)
}

func TestParseMarkerEscapedNewlineInLabel(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker(`Multi\nLine|Fire|Water`)
	c.True(ok)
	c.Equal("Multi\nLine", marker.Label)
}

func TestParseMarkerAllowEmptyAndFreeForm(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker("Element|Fire|Water|?|*")
	c.True(ok)
	c.True(marker.AllowEmpty)
	c.True(marker.FreeForm)
	c.Equal([]string{"Fire", "Water"}, marker.Options)
}

func TestParseMarkerFreeFormOnlyNoOptions(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker("Element|*")
	c.True(ok)
	c.True(marker.FreeForm)
	c.Equal(0, len(marker.Options))
}

func TestParseMarkerEmptyLabelFails(t *testing.T) {
	c := check.New(t)
	_, ok := nameable.ParseMarker("|Fire|Water")
	c.False(ok)
}

func TestParseMarkerNoOptionsImpliesFreeForm(t *testing.T) {
	c := check.New(t)
	// No literal options and no explicit FreeFormToken -- there's nothing else it could mean, so FreeForm is turned
	// on automatically instead of failing.
	marker, ok := nameable.ParseMarker("Element|?")
	c.True(ok)
	c.True(marker.FreeForm)
	c.True(marker.AllowEmpty)
	c.Equal(0, len(marker.Options))
}

func TestParseMarkerNoPipeFallsBackToLegacy(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker("Element:Fire:Water")
	c.True(ok)
	c.True(marker.Legacy)
}

func TestParseMarkerSingleSegmentFallsBackToLegacy(t *testing.T) {
	c := check.New(t)
	marker, ok := nameable.ParseMarker("Element")
	c.True(ok)
	c.True(marker.Legacy)
	c.Equal("Element", marker.Label)
}

func TestDefaultValue(t *testing.T) {
	c := check.New(t)

	marker, ok := nameable.ParseMarker("Weapon Name")
	c.True(ok)
	c.Equal("Weapon Name", marker.DefaultValue())

	marker, ok = nameable.ParseMarker("Element|Fire|Water")
	c.True(ok)
	c.Equal("Fire", marker.DefaultValue())

	marker, ok = nameable.ParseMarker("Element|Fire|Water|?")
	c.True(ok)
	c.Equal("", marker.DefaultValue())
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
