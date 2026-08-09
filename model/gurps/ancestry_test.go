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

// TestRandomStringOptions verifies that randomizing the hair, eyes, skin and handedness always yields a value the
// ancestry actually defines. Excluding the current value from an ancestry that defines only one left a total weight of
// 0, so the hard-coded package default ("Brown"/"Right") was substituted for the ancestry's own lone choice.
func TestRandomStringOptions(t *testing.T) {
	c := check.New(t)

	opt := func(value string) *WeightedStringOption {
		return &WeightedStringOption{Weight: 1, Value: value}
	}

	single := &Ancestry{CommonOptions: &AncestryOptions{AncestryOptionsData: AncestryOptionsData{
		HairOptions:       []*WeightedStringOption{opt("Green")},
		EyeOptions:        []*WeightedStringOption{opt("Violet")},
		SkinOptions:       []*WeightedStringOption{opt("Blue")},
		HandednessOptions: []*WeightedStringOption{opt("Ambidextrous")},
	}}}
	c.Equal("Green", single.RandomHair("", "Green"), "the lone hair option is kept rather than replaced")
	c.Equal("Violet", single.RandomEyes("", "Violet"), "the lone eye option is kept rather than replaced")
	c.Equal("Blue", single.RandomSkin("", "Blue"), "the lone skin option is kept rather than replaced")
	c.Equal("Ambidextrous", single.RandomHandedness("", "Ambidextrous"),
		"the lone handedness option is kept rather than replaced")
	c.Equal("Green", single.RandomHair("", ""), "the lone hair option is chosen when nothing is excluded")

	// When an alternative exists, the current value is still excluded from the choice.
	pair := &Ancestry{CommonOptions: &AncestryOptions{AncestryOptionsData: AncestryOptionsData{
		HairOptions: []*WeightedStringOption{opt("Green"), opt("Blue")},
	}}}
	for range 20 {
		c.Equal("Blue", pair.RandomHair("", "Green"), "the alternative hair is chosen when one exists")
	}

	// An ancestry with no options at all still falls back to the package defaults.
	none := &Ancestry{}
	c.Equal(defaultHair, none.RandomHair("", "Green"), "no hair options falls back to the default")
	c.Equal(defaultEye, none.RandomEyes("", "Violet"), "no eye options falls back to the default")
	c.Equal(defaultSkin, none.RandomSkin("", "Blue"), "no skin options falls back to the default")
	c.Equal(defaultHandedness, none.RandomHandedness("", "Ambidextrous"),
		"no handedness options falls back to the default")

	// Options that can never be chosen (no weight) are treated as if they weren't there.
	zero := &AncestryOptions{AncestryOptionsData: AncestryOptionsData{
		HairOptions: []*WeightedStringOption{{Value: "Green"}},
	}}
	c.Equal(defaultHair, zero.RandomHair("Green"), "a weightless hair option falls back to the default")
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

// TestAncestryWithNullStringOptions verifies that a JSON null in one of the string option arrays is ignored rather than
// dereferenced. Randomizing the description (or creating a sheet with auto-fill enabled) reaches these lists and
// previously panicked on the nil entry.
func TestAncestryWithNullStringOptions(t *testing.T) {
	c := check.New(t)

	const data = `{
	"type": "ancestry",
	"version": 5,
	"name": "Human",
	"common_options": {
		"hair_options": [ null, { "weight": 1, "value": "Green" } ],
		"eye_options": [ null, { "weight": 1, "value": "Violet" } ],
		"skin_options": [ null, { "weight": 1, "value": "Blue" } ],
		"handedness_options": [ null, { "weight": 1, "value": "Ambidextrous" } ]
	}
}`
	fileSystem := fstest.MapFS{"Human.ancestry": {Data: []byte(data)}}
	a, err := NewAncestryFromFile(fileSystem, "Human.ancestry")
	c.NoError(err, "ancestry file with null string options should load")
	c.Equal(2, len(a.CommonOptions.HairOptions), "both hair options should be present")
	c.Nil(a.CommonOptions.HairOptions[0], "the first hair option should be nil")

	// Only the valid option may be chosen, no matter what is excluded.
	for range 20 {
		c.Equal("Green", a.RandomHair("", ""), "the only valid hair option is chosen")
		c.Equal("Green", a.RandomHair("", "Blue"), "the nil hair option is never chosen")
		c.Equal("Violet", a.RandomEyes("", ""), "the only valid eye option is chosen")
		c.Equal("Blue", a.RandomSkin("", ""), "the only valid skin option is chosen")
		c.Equal("Ambidextrous", a.RandomHandedness("", ""), "the only valid handedness option is chosen")
	}
	c.Equal("Green", a.RandomHair("", "Green"), "excluding the lone valid option keeps the current hair")

	// A list consisting solely of unusable entries falls back to the default rather than panicking.
	only := &AncestryOptions{AncestryOptionsData: AncestryOptionsData{
		HairOptions:       []*WeightedStringOption{nil},
		EyeOptions:        []*WeightedStringOption{nil, {Value: "Violet"}},
		SkinOptions:       []*WeightedStringOption{nil},
		HandednessOptions: []*WeightedStringOption{nil},
	}}
	c.Equal(defaultHair, only.RandomHair("Green"), "a nil-only hair list falls back to the default")
	c.Equal(defaultEye, only.RandomEye(""), "a nil plus weightless eye list falls back to the default")
	c.Equal(defaultSkin, only.RandomSkin(""), "a nil-only skin list falls back to the default")
	c.Equal(defaultHandedness, only.RandomHandedness(""), "a nil-only handedness list falls back to the default")
}
