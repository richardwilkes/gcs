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
	"errors"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestParseSyntaxPlainKeyFails(t *testing.T) {
	c := check.New(t)
	_, err := nameable.ParseSyntax("Weapon Name")
	c.True(errors.Is(err, nameable.ErrNoSyntax))
}

func TestParseSyntaxBasic(t *testing.T) {
	c := check.New(t)
	marker, err := nameable.ParseSyntax("Element|Fire|Water|Earth")
	c.NoError(err)
	c.Equal("Element|Fire|Water|Earth", marker.Raw)
	c.Equal("Element", marker.Label)
	c.Equal("", marker.Tooltip)
	c.Equal([]string{"Fire", "Water", "Earth"}, marker.Options)
	c.False(marker.AllowEmpty)
	c.False(marker.FreeForm)
}

func TestParseSyntaxWithTooltip(t *testing.T) {
	c := check.New(t)
	marker, err := nameable.ParseSyntax("Element|tt(Pick the affinity)|Fire|Water")
	c.NoError(err)
	c.Equal("Element", marker.Label)
	c.Equal("Pick the affinity", marker.Tooltip)
	c.Equal([]string{"Fire", "Water"}, marker.Options)
}

func TestParseSyntaxLabelMayContainColon(t *testing.T) {
	c := check.New(t)
	marker, err := nameable.ParseSyntax("Time: HH:MM|Morning|Afternoon|Evening")
	c.NoError(err)
	c.Equal("Time: HH:MM", marker.Label)
	c.Equal("", marker.Tooltip)
	c.Equal([]string{"Morning", "Afternoon", "Evening"}, marker.Options)
}

func TestParseSyntaxTooltipMayContainColon(t *testing.T) {
	c := check.New(t)
	marker, err := nameable.ParseSyntax("Element|tt(Ratio is 1:2)|Fire|Water")
	c.NoError(err)
	c.Equal("Element", marker.Label)
	c.Equal("Ratio is 1:2", marker.Tooltip)
}

func TestParseSyntaxTooltipAnywhereInSegments(t *testing.T) {
	c := check.New(t)
	marker, err := nameable.ParseSyntax("Element|Fire|tt(Pick one)|Water|?")
	c.NoError(err)
	c.Equal("Pick one", marker.Tooltip)
	c.Equal([]string{"Fire", "Water"}, marker.Options)
	c.True(marker.AllowEmpty)
}

func TestParseSyntaxMultipleTooltipLinesJoined(t *testing.T) {
	c := check.New(t)
	marker, err := nameable.ParseSyntax("Element|tt(First line)|tt(Second line)|Fire")
	c.NoError(err)
	c.Equal("First line\nSecond line", marker.Tooltip)
	c.Equal([]string{"Fire"}, marker.Options)
}

func TestParseSyntaxEscapedPipeInLabel(t *testing.T) {
	c := check.New(t)
	marker, err := nameable.ParseSyntax(`Fire\|Water|Hot|Cold`)
	c.NoError(err)
	c.Equal("Fire|Water", marker.Label)
	c.Equal([]string{"Hot", "Cold"}, marker.Options)
}

func TestParseSyntaxEscapedPipeInOption(t *testing.T) {
	c := check.New(t)
	marker, err := nameable.ParseSyntax(`Element|Hot\|Cold|Water`)
	c.NoError(err)
	c.Equal([]string{"Hot|Cold", "Water"}, marker.Options)
}

func TestParseSyntaxEscapedPipeInTooltip(t *testing.T) {
	c := check.New(t)
	marker, err := nameable.ParseSyntax(`Element|tt(Choose one\|Two)|Fire`)
	c.NoError(err)
	c.Equal("Choose one|Two", marker.Tooltip)
	c.Equal([]string{"Fire"}, marker.Options)
}

func TestParseSyntaxEscapedNewlineInTooltip(t *testing.T) {
	c := check.New(t)
	marker, err := nameable.ParseSyntax(`Element|tt(Line one\nLine two)|Fire`)
	c.NoError(err)
	c.Equal("Line one\nLine two", marker.Tooltip)
}

func TestParseSyntaxEscapedNewlineInLabel(t *testing.T) {
	c := check.New(t)
	marker, err := nameable.ParseSyntax(`Multi\nLine|Fire|Water`)
	c.NoError(err)
	c.Equal("Multi\nLine", marker.Label)
}

func TestParseSyntaxAllowEmptyAndFreeForm(t *testing.T) {
	c := check.New(t)
	marker, err := nameable.ParseSyntax("Element|Fire|Water|?|*")
	c.NoError(err)
	c.True(marker.AllowEmpty)
	c.True(marker.FreeForm)
	c.Equal([]string{"Fire", "Water"}, marker.Options)
}

func TestParseSyntaxFreeFormOnlyNoOptions(t *testing.T) {
	c := check.New(t)
	marker, err := nameable.ParseSyntax("Element|*")
	c.NoError(err)
	c.True(marker.FreeForm)
	c.Equal(0, len(marker.Options))
}

func TestParseSyntaxEmptyLabelFails(t *testing.T) {
	c := check.New(t)
	_, err := nameable.ParseSyntax("|Fire|Water")
	c.True(errors.Is(err, nameable.ErrEmptyLabel))
}

func TestParseSyntaxNoOptionsNoFreeFormFails(t *testing.T) {
	c := check.New(t)
	_, err := nameable.ParseSyntax("Element|?")
	c.True(errors.Is(err, nameable.ErrNoOptions))
}

func TestParseSyntaxNoPipeFails(t *testing.T) {
	c := check.New(t)
	_, err := nameable.ParseSyntax("Element:Fire:Water")
	c.True(errors.Is(err, nameable.ErrNoSyntax))
}

func TestParseSyntaxSingleSegmentFails(t *testing.T) {
	c := check.New(t)
	_, err := nameable.ParseSyntax("Element")
	c.True(errors.Is(err, nameable.ErrNoSyntax))
}

func TestDefaultValue(t *testing.T) {
	c := check.New(t)
	c.Equal("Weapon Name", nameable.DefaultValue("Weapon Name"))
	c.Equal("Fire", nameable.DefaultValue("Element|Fire|Water"))
	c.Equal("", nameable.DefaultValue("Element|Fire|Water|?"))
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
