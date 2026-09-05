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
	"fmt"
	"os"
	"path/filepath"
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
			Value:  &AncestryOptions{Name: name},
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

	single := &Ancestry{CommonOptions: &AncestryOptions{
		HairOptions:       []*WeightedStringOption{opt("Green")},
		EyeOptions:        []*WeightedStringOption{opt("Violet")},
		SkinOptions:       []*WeightedStringOption{opt("Blue")},
		HandednessOptions: []*WeightedStringOption{opt("Ambidextrous")},
	}}
	c.Equal("Green", single.RandomHair("", "Green"), "the lone hair option is kept rather than replaced")
	c.Equal("Violet", single.RandomEyes("", "Violet"), "the lone eye option is kept rather than replaced")
	c.Equal("Blue", single.RandomSkin("", "Blue"), "the lone skin option is kept rather than replaced")
	c.Equal("Ambidextrous", single.RandomHandedness("", "Ambidextrous"),
		"the lone handedness option is kept rather than replaced")
	c.Equal("Green", single.RandomHair("", ""), "the lone hair option is chosen when nothing is excluded")

	// When an alternative exists, the current value is still excluded from the choice.
	pair := &Ancestry{CommonOptions: &AncestryOptions{
		HairOptions: []*WeightedStringOption{opt("Green"), opt("Blue")},
	}}
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
	zero := &AncestryOptions{
		HairOptions: []*WeightedStringOption{{Value: "Green"}},
	}
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
	only := &AncestryOptions{
		HairOptions:       []*WeightedStringOption{nil},
		EyeOptions:        []*WeightedStringOption{nil, {Value: "Violet"}},
		SkinOptions:       []*WeightedStringOption{nil},
		HandednessOptions: []*WeightedStringOption{nil},
	}
	c.Equal(defaultHair, only.RandomHair("Green"), "a nil-only hair list falls back to the default")
	c.Equal(defaultEye, only.RandomEye(""), "a nil plus weightless eye list falls back to the default")
	c.Equal(defaultSkin, only.RandomSkin(""), "a nil-only skin list falls back to the default")
	c.Equal(defaultHandedness, only.RandomHandedness(""), "a nil-only handedness list falls back to the default")
}

// populatedAncestry returns an ancestry with every level filled in -- common options with all four weighted lists and
// name generators, plus two genders that carry options of their own -- for the tests that need to reach each of them.
func populatedAncestry() *Ancestry {
	return &Ancestry{
		Name: "Elf",
		CommonOptions: &AncestryOptions{
			HeightScript:      "60 + 6",
			WeightScript:      "100 + 20",
			AgeScript:         "18 + 100",
			HairOptions:       []*WeightedStringOption{entry(3, "Blond"), entry(1, "Silver")},
			EyeOptions:        []*WeightedStringOption{entry(2, "Green"), entry(1, "Violet")},
			SkinOptions:       []*WeightedStringOption{entry(1, "Fair")},
			HandednessOptions: []*WeightedStringOption{entry(9, "Right"), entry(1, "Left")},
			NameGenerators:    []string{"Elf First", "Elf Last"},
		},
		GenderOptions: []*WeightedAncestryOptions{
			{Weight: 1, Value: &AncestryOptions{Name: "Male", EyeOptions: []*WeightedStringOption{entry(1, "Gray")}}},
			{Weight: 1, Value: &AncestryOptions{Name: "Female", HairOptions: []*WeightedStringOption{entry(1, "Red")}}},
		},
	}
}

// keyPrefixCounter returns a prefix provider that hands out "p<n>" for successive n, starting after start, along with a
// pointer to the count of prefixes it has produced.
func keyPrefixCounter(start int) (provider func() string, count *int) {
	n := 0
	return func() string {
		n++
		return fmt.Sprintf("p%d", start+n)
	}, &n
}

// collectKeyPrefixes walks the ancestry in a fixed order and returns the KeyPrefix of every non-nil node: the ancestry,
// the common options and each of their list entries, then each gender entry, its value, and that value's list entries.
func collectKeyPrefixes(a *Ancestry) []string {
	var prefixes []string
	options := func(o *AncestryOptions) {
		if o == nil {
			return
		}
		prefixes = append(prefixes, o.KeyPrefix)
		for _, list := range [][]*WeightedStringOption{o.HairOptions, o.EyeOptions, o.SkinOptions, o.HandednessOptions} {
			for _, one := range list {
				if one != nil {
					prefixes = append(prefixes, one.KeyPrefix)
				}
			}
		}
	}
	prefixes = append(prefixes, a.KeyPrefix)
	options(a.CommonOptions)
	for _, one := range a.GenderOptions {
		if one != nil {
			prefixes = append(prefixes, one.KeyPrefix)
			options(one.Value)
		}
	}
	return prefixes
}

// TestNewAncestryDefaults verifies that a freshly created ancestry carries the scaffolding an editor expects, and that
// it is already in normalized form.
func TestNewAncestryDefaults(t *testing.T) {
	c := check.New(t)

	a := NewAncestry()
	c.Equal("", a.Name, "a new ancestry has no name")
	c.NotNil(a.CommonOptions, "a new ancestry has a common options block")
	c.Equal(2, len(a.GenderOptions), "a new ancestry has two genders")
	c.NotNil(a.GenderOptions[0].Value, "the first gender has a value")
	c.NotNil(a.GenderOptions[1].Value, "the second gender has a value")
	c.Equal("Male", a.GenderOptions[0].Value.Name, "the first gender is Male")
	c.Equal("Female", a.GenderOptions[1].Value.Name, "the second gender is Female")
	c.Equal(1, a.GenderOptions[0].Weight, "the first gender has a weight of 1")
	c.Equal(1, a.GenderOptions[1].Weight, "the second gender has a weight of 1")

	before := Hash64(a)
	a.Normalize()
	c.Equal(before, Hash64(a), "normalizing a new ancestry changes nothing")
}

// TestAncestryCloneIsDeep verifies that Clone copies every level of the ancestry, so that editing the clone leaves the
// original untouched, while still carrying the runtime-only key prefixes across and faithfully preserving nil entries.
func TestAncestryCloneIsDeep(t *testing.T) {
	c := check.New(t)

	original := populatedAncestry()
	provider, _ := keyPrefixCounter(0)
	original.ResetTargetKeyPrefixes(provider)
	clone := original.Clone()
	c.Equal(Hash64(original), Hash64(clone), "the clone has the same content as the original")
	c.Equal(collectKeyPrefixes(original), collectKeyPrefixes(clone), "the clone carries the same key prefixes")

	// Mutate every level of the clone.
	clone.Name = "Dwarf"
	clone.CommonOptions.HairOptions[0].Value = "Black"
	clone.CommonOptions.NameGenerators[0] = "Dwarf First"
	clone.GenderOptions[0].Weight = 7
	clone.GenderOptions[0].Value.Name = "Other"
	clone.GenderOptions[0].Value.EyeOptions[0].Weight = 5

	// None of it reached the original.
	c.Equal("Elf", original.Name, "the original's name is unchanged")
	c.Equal("Blond", original.CommonOptions.HairOptions[0].Value, "the original's hair option is unchanged")
	c.Equal("Elf First", original.CommonOptions.NameGenerators[0], "the original's name generator is unchanged")
	c.Equal(1, original.GenderOptions[0].Weight, "the original's gender weight is unchanged")
	c.Equal("Male", original.GenderOptions[0].Value.Name, "the original's gender name is unchanged")
	c.Equal(1, original.GenderOptions[0].Value.EyeOptions[0].Weight, "the original's gender eye option is unchanged")
	c.Equal(Hash64(populatedAncestry()), Hash64(original), "the original hashes the same as a fresh copy")

	// Mutating the clone did not disturb its key prefixes either.
	c.Equal(collectKeyPrefixes(original), collectKeyPrefixes(clone), "key prefixes survive editing the clone")

	c.Nil((*Ancestry)(nil).Clone(), "cloning a nil ancestry yields nil")

	// Nil entries and nil lists are preserved as-is rather than being dropped or materialized.
	sparse := &Ancestry{
		CommonOptions: &AncestryOptions{
			HairOptions: []*WeightedStringOption{nil, {Weight: 1, Value: "Green"}},
		},
		GenderOptions: []*WeightedAncestryOptions{nil, {Weight: 1}},
	}
	sparseClone := sparse.Clone()
	c.Equal(2, len(sparseClone.CommonOptions.HairOptions), "the hair list keeps its length")
	c.Nil(sparseClone.CommonOptions.HairOptions[0], "a nil hair entry stays nil")
	c.Equal("Green", sparseClone.CommonOptions.HairOptions[1].Value, "the real hair entry is copied")
	c.Nil(sparseClone.CommonOptions.EyeOptions, "a nil eye list stays nil")
	c.Nil(sparseClone.CommonOptions.NameGenerators, "a nil name generator list stays nil")
	c.Equal(2, len(sparseClone.GenderOptions), "the gender list keeps its length")
	c.Nil(sparseClone.GenderOptions[0], "a nil gender entry stays nil")
	c.Nil(sparseClone.GenderOptions[1].Value, "a valueless gender entry stays valueless")
	c.Nil((&Ancestry{}).Clone().CommonOptions, "nil common options stay nil")
	c.Nil((&Ancestry{}).Clone().GenderOptions, "a nil gender list stays nil")
}

// TestAncestryHashIgnoresKeyPrefixAndTracksContent verifies that the hash is a function of the file content alone: the
// runtime-only key prefixes never affect it, while a change at any level of the data does.
func TestAncestryHashIgnoresKeyPrefixAndTracksContent(t *testing.T) {
	c := check.New(t)

	a := populatedAncestry()
	base := Hash64(a)
	c.Equal(base, Hash64(a.Clone()), "a clone hashes the same as the original")

	provider, _ := keyPrefixCounter(1000)
	a.ResetTargetKeyPrefixes(provider)
	c.Equal(base, Hash64(a), "assigning key prefixes does not change the hash")
	provider, _ = keyPrefixCounter(5000)
	a.ResetTargetKeyPrefixes(provider)
	c.Equal(base, Hash64(a), "assigning different key prefixes does not change the hash")

	changes := []struct {
		name   string
		mutate func(*Ancestry)
	}{
		{"name", func(a *Ancestry) { a.Name = "Dwarf" }},
		{"common hair weight", func(a *Ancestry) { a.CommonOptions.HairOptions[0].Weight++ }},
		{"nested value string", func(a *Ancestry) { a.GenderOptions[0].Value.EyeOptions[0].Value = "Amber" }},
		{"gender name", func(a *Ancestry) { a.GenderOptions[1].Value.Name = "Other" }},
		{"name generator", func(a *Ancestry) {
			a.CommonOptions.NameGenerators = append(a.CommonOptions.NameGenerators, "Elf Title")
		}},
	}
	for _, one := range changes {
		edited := a.Clone()
		one.mutate(edited)
		c.NotEqual(base, Hash64(edited), "changing the %s changes the hash", one.name)
	}
}

// TestAncestryResetTargetKeyPrefixesAreUnique verifies that every node in the ancestry receives its own prefix from the
// provider, exactly once each, and that nil entries are skipped rather than dereferenced.
func TestAncestryResetTargetKeyPrefixesAreUnique(t *testing.T) {
	c := check.New(t)

	a := populatedAncestry()
	a.CommonOptions.HairOptions = append(a.CommonOptions.HairOptions, nil)
	a.GenderOptions = append(a.GenderOptions, nil)

	provider, count := keyPrefixCounter(0)
	c.NotPanics(func() { a.ResetTargetKeyPrefixes(provider) }, "nil entries are skipped without panicking")

	prefixes := collectKeyPrefixes(a)
	// The ancestry, its common options with 2+2+1+2 entries, and two genders each with a value holding 1 entry.
	const nodeCount = 1 + (1 + 7) + 2*(1+1+1)
	c.Equal(nodeCount, len(prefixes), "every non-nil node was visited")
	c.Equal(nodeCount, *count, "the provider was called exactly once per node")
	set := make(map[string]bool, len(prefixes))
	for _, one := range prefixes {
		c.NotEqual("", one, "every node received a prefix")
		set[one] = true
	}
	c.Equal(nodeCount, len(set), "every node received a distinct prefix")
	c.Nil(a.CommonOptions.HairOptions[2], "the nil hair entry is still nil")
	c.Nil(a.GenderOptions[2], "the nil gender entry is still nil")
}

// TestAncestryNormalize verifies that Normalize turns a hand-written file's tolerated gaps -- null list entries, a
// gender without a value, a missing common options block -- into the fully populated structure an editor requires,
// without discarding any real data.
func TestAncestryNormalize(t *testing.T) {
	c := check.New(t)

	const data = `{
	"type": "ancestry",
	"version": 5,
	"name": "Human",
	"gender_options": [
		{ "weight": 3 },
		null,
		{
			"weight": 1,
			"value": {
				"name": "Female",
				"hair_options": [ null, { "weight": 1, "value": "Green" }, null ],
				"eye_options": [ null ]
			}
		}
	]
}`
	fileSystem := fstest.MapFS{"Human.ancestry": {Data: []byte(data)}}
	a, err := NewAncestryFromFile(fileSystem, "Human.ancestry")
	c.NoError(err, "ancestry file should load")

	// Loading alone leaves the gaps in place; the load path must keep tolerating such files as-is.
	c.Nil(a.CommonOptions, "loading does not add common options")
	c.Equal(3, len(a.GenderOptions), "loading keeps the null gender entry")
	c.Nil(a.GenderOptions[1], "the null gender entry is nil")
	c.Nil(a.GenderOptions[0].Value, "the valueless gender entry has no value")

	a.Normalize()
	c.NotNil(a.CommonOptions, "normalizing adds a common options block")
	c.Equal(2, len(a.GenderOptions), "normalizing drops exactly the null gender entry")
	for i, one := range a.GenderOptions {
		c.NotNil(one, "gender entry %d is present", i)
		c.NotNil(one.Value, "gender entry %d has a value", i)
	}
	c.Equal(3, a.GenderOptions[0].Weight, "the valueless gender keeps its weight")
	c.Equal("", a.GenderOptions[0].Value.Name, "the valueless gender gains an empty value")
	female := a.GenderOptions[1].Value
	c.Equal("Female", female.Name, "the valued gender keeps its data")
	c.Equal(1, len(female.HairOptions), "the null hair entries are dropped")
	c.Equal("Green", female.HairOptions[0].Value, "the real hair entry is kept")
	c.Equal(0, len(female.EyeOptions), "a list of only nulls becomes empty")
	for _, list := range [][]*WeightedStringOption{
		female.HairOptions, female.EyeOptions, female.SkinOptions, female.HandednessOptions,
	} {
		for _, one := range list {
			c.NotNil(one, "no nil entries remain")
		}
	}

	// Normalizing again is a no-op.
	before := Hash64(a)
	a.Normalize()
	c.Equal(before, Hash64(a), "normalizing is idempotent")
}

// TestAncestrySaveLoadRoundTrip verifies that an ancestry written by Save reads back with identical content, and that
// the file carries the current data version but none of the runtime-only key prefixes.
func TestAncestrySaveLoadRoundTrip(t *testing.T) {
	c := check.New(t)

	a := populatedAncestry()
	provider, _ := keyPrefixCounter(0)
	a.ResetTargetKeyPrefixes(provider)
	dir := t.TempDir()
	path := filepath.Join(dir, "Elf.ancestry")
	c.NoError(a.Save(path), "saving the ancestry succeeds")

	data, err := os.ReadFile(path)
	c.NoError(err, "the saved file can be read")
	c.Contains(string(data), `"version": 5`, "the file carries the current data version")
	c.NotContains(string(data), "KeyPrefix", "the runtime-only key prefixes are not written")
	c.NotContains(string(data), "p1", "no key prefix value leaks into the file")

	loaded, err := NewAncestryFromFile(os.DirFS(dir), "Elf.ancestry")
	c.NoError(err, "the saved file loads")
	c.Equal(Hash64(a), Hash64(loaded), "the loaded ancestry has the same content")
	c.Equal("", loaded.KeyPrefix, "a loaded ancestry has no key prefix")
}
