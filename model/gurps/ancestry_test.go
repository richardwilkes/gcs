// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestNewAncestryFromInvalidFile verifies that loading an ancestry file containing invalid JSON returns an error rather
// than crashing the application. See issue #1007, where a missing comma in an ancestry file caused GCS to exit when
// randomizing a value.
func TestNewAncestryFromInvalidFile(t *testing.T) {
	c := check.New(t)

	// A well-formed ancestry file loads without error.
	const valid = `{
	"type": "ancestry",
	"version": 5,
	"name": "Human"
}`
	fileSystem := fstest.MapFS{"Human.ancestry": {Data: []byte(valid)}}
	a, err := NewAncestryFromFile(fileSystem, "Human.ancestry")
	c.NoError(err, "valid ancestry file should load")
	c.Equal("Human", a.Name, "name should be loaded")

	// The same content with a missing comma (matching the report in issue #1007) must produce an error instead of
	// panicking or exiting.
	const invalid = `{
	"type": "ancestry"
	"version": 5,
	"name": "Human"
}`
	fileSystem = fstest.MapFS{"Human.ancestry": {Data: []byte(invalid)}}
	a, err = NewAncestryFromFile(fileSystem, "Human.ancestry")
	c.HasError(err, "invalid ancestry file should return an error, not crash")
	c.Equal((*Ancestry)(nil), a, "no ancestry should be returned on failure")
}

// TestRandomGender verifies that randomizing the gender always yields a usable value. Excluding the current gender from
// an ancestry that defines only one left a total weight of 0, so "" was returned and the sheet's randomize button
// blanked the field rather than leaving it alone.
func TestRandomGender(t *testing.T) {
	c := check.New(t)

	gender := func(name string) *WeightedAncestryOptions {
		return &WeightedAncestryOptions{
			Weight: 1,
			Value:  &AncestryOptions{AncestryOptionsData: AncestryOptionsData{Name: name}},
		}
	}

	single := &Ancestry{GenderOptions: []*WeightedAncestryOptions{gender("Female")}}
	c.Equal("Female", single.RandomGender("Female"), "the lone gender option is kept rather than blanked")
	c.Equal("Female", single.RandomGender(""), "the lone gender option is chosen when nothing is excluded")

	// When an alternative exists, the current gender is still excluded from the choice.
	pair := &Ancestry{GenderOptions: []*WeightedAncestryOptions{gender("Female"), gender("Male")}}
	for range 20 {
		c.Equal("Male", pair.RandomGender("Female"), "the alternative gender is chosen when one exists")
	}

	// An ancestry with no gender options at all preserves the incoming value.
	c.Equal("Female", (&Ancestry{}).RandomGender("Female"), "no gender options preserves the current gender")
}
