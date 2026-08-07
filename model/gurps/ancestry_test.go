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

// TestAncestryWithValuelessGenderOption verifies that a gender option whose "value" object is missing from the file is
// ignored rather than dereferenced. Randomizing an entity's description (or creating a sheet with auto-fill enabled)
// reaches both RandomGender and GenderedOptions, which previously panicked on the nil Value.
func TestAncestryWithValuelessGenderOption(t *testing.T) {
	c := check.New(t)

	const data = `{
	"type": "ancestry",
	"version": 5,
	"name": "Human",
	"gender_options": [
		{ "weight": 1 },
		{ "weight": 1, "value": { "name": "Female" } }
	]
}`
	fileSystem := fstest.MapFS{"Human.ancestry": {Data: []byte(data)}}
	a, err := NewAncestryFromFile(fileSystem, "Human.ancestry")
	c.NoError(err, "ancestry file with a valueless gender option should load")
	c.Equal(2, len(a.GenderOptions), "both gender options should be present")
	c.Nil(a.GenderOptions[0].Value, "the first gender option should have no value")

	// Only the option that actually has a value may be chosen, no matter what is excluded.
	for range 20 {
		c.Equal("Female", a.RandomGender(""), "the only valued gender option is chosen")
		c.Equal("Female", a.RandomGender("Male"), "the valueless gender option is never chosen")
	}
	c.Equal("Female", a.RandomGender("Female"), "excluding the lone valued option keeps the current gender")

	c.NotNil(a.GenderedOptions("Female"), "options for a valued gender are found")
	c.Nil(a.GenderedOptions(""), "a valueless gender option is not matched by an empty gender")
	c.Nil(a.GenderedOptions("Male"), "an unknown gender has no options")

	// The randomizers that consult the gender options must fall back to their defaults rather than panicking.
	c.Equal(defaultHair, a.RandomHair("", ""), "hair falls back to the default")
	c.Equal(defaultEye, a.RandomEyes("", ""), "eyes fall back to the default")
	c.Equal(defaultSkin, a.RandomSkin("", ""), "skin falls back to the default")
	c.Equal(defaultHandedness, a.RandomHandedness("", ""), "handedness falls back to the default")
	c.Equal(defaultAge, a.RandomAge(nil, "", 0), "age falls back to the default")

	// A completely nil entry in the list (from a JSON null) must be tolerated as well.
	a.GenderOptions = append(a.GenderOptions, nil)
	c.Equal("Female", a.RandomGender(""), "a nil gender option is ignored")
	c.Nil(a.GenderedOptions("Male"), "a nil gender option is not matched")
}
