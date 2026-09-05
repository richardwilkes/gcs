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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/namegen"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// entry builds a weighted string option: a name generator's training data entry, or one of an ancestry's choices.
func entry(weight int, value string) *WeightedStringOption {
	return &WeightedStringOption{Weight: weight, Value: value}
}

// saveAndReload writes the generator to a file in a temporary directory, then reads it back without preparing it for
// use. The JSON text of the file is returned alongside the reloaded generator, so tests can assert on the on-disk form.
func saveAndReload(t *testing.T, c check.Checker, n *NameGenerator) (reloaded *NameGenerator, text string) {
	dir := t.TempDir()
	const base = "Test.names"
	c.NoError(n.Save(filepath.Join(dir, base)), "save should succeed")
	data, err := os.ReadFile(filepath.Join(dir, base))
	c.NoError(err, "the saved file should be readable")
	reloaded, err = ReadNameGeneratorFromFS(os.DirFS(dir), base)
	c.NoError(err, "the saved file should load")
	text = string(data)
	c.NotContains(text, "KeyPrefix", "the runtime-only key prefix must never be written")
	c.NotContains(text, "key_prefix", "the runtime-only key prefix must never be written")
	return reloaded, text
}

// TestNameGeneratorRoundTripUnweighted verifies that entries whose weights are all 1 are written as the plain
// "training_data" list, in order, and read back in that same order.
func TestNameGeneratorRoundTripUnweighted(t *testing.T) {
	c := check.New(t)
	n := &NameGenerator{
		Type:    namegen.Simple,
		Entries: []*WeightedStringOption{entry(1, "Zed"), entry(1, "alpha"), entry(1, "Mike")},
	}
	n.KeyPrefix = "root-prefix"
	n.Entries[0].KeyPrefix = "entry-prefix"
	reloaded, text := saveAndReload(t, c, n)
	c.Contains(text, `"training_data"`, "unit weights are written as the plain list")
	c.NotContains(text, `"weighted_training_data"`, "unit weights do not need the weighted map")
	c.NotContains(text, `"built_in_training_data"`, "no built-in set was selected")
	c.Equal(namegen.Simple, reloaded.Type)
	c.Equal(3, len(reloaded.Entries), "all entries should survive the round trip")
	for i, expected := range []string{"Zed", "alpha", "Mike"} {
		c.Equal(expected, reloaded.Entries[i].Value, "file order is preserved for entry %d", i)
		c.Equal(1, reloaded.Entries[i].Weight, "weight is 1 for entry %d", i)
	}
}

// TestNameGeneratorRoundTripWeighted verifies that entries with mixed weights are written as the
// "weighted_training_data" map and read back sorted by name with their weights intact, and that a name appearing more
// than once has its weights added together.
func TestNameGeneratorRoundTripWeighted(t *testing.T) {
	c := check.New(t)
	n := &NameGenerator{
		Type:    namegen.MarkovRun,
		Entries: []*WeightedStringOption{entry(2, "Zed"), entry(1, "alpha"), entry(3, "Beta"), entry(7, "Zed")},
	}
	reloaded, text := saveAndReload(t, c, n)
	c.Contains(text, `"weighted_training_data"`, "mixed weights need the weighted map")
	c.NotContains(text, `"training_data"`, "mixed weights cannot be expressed as the plain list")
	c.Contains(text, `"Zed": 9`, "the duplicate name is written once, with the sum of its weights")
	c.Equal(namegen.MarkovRun, reloaded.Type)
	c.Equal(3, len(reloaded.Entries), "the duplicate name collapses into one entry")
	expected := []*WeightedStringOption{entry(1, "alpha"), entry(3, "Beta"), entry(9, "Zed")}
	for i, one := range expected {
		c.Equal(one.Value, reloaded.Entries[i].Value, "entries are sorted by name (%d)", i)
		c.Equal(one.Weight, reloaded.Entries[i].Weight, "weight is preserved for %s", one.Value)
	}
}

// TestNameGeneratorDuplicateNamesAddWeights verifies that a name listed more than once has its weights added together
// everywhere: in the training data a namer is built from, in the weighted map written to disk, and so across a save and
// reload in either on-disk form, which therefore preserves the distribution the entries describe.
func TestNameGeneratorDuplicateNamesAddWeights(t *testing.T) {
	c := check.New(t)

	weighted := &NameGenerator{
		Type:    namegen.Simple,
		Entries: []*WeightedStringOption{entry(2, "Zed"), entry(1, "alpha"), entry(3, "Zed"), entry(0, "Zed")},
	}
	data := weighted.data()
	c.Equal(map[string]int{"Zed": 5, "alpha": 1}, data,
		"a repeated name's weights add together, a weightless one adds nothing")
	reloaded, text := saveAndReload(t, c, weighted)
	c.Contains(text, `"Zed": 5`, "the weighted map carries the sum")
	c.Equal(2, len(reloaded.Entries), "the repeated name reads back as a single entry")
	c.Equal(data, reloaded.data(), "the distribution survives the weighted form")

	// Sampling draws on the summed weights, so both names, and nothing else, can come up.
	samples, err := weighted.SampleNames(20)
	c.NoError(err)
	for _, one := range samples {
		c.True(one == "Zed" || one == "Alpha", "samples come from the training names: %q", one)
	}

	// The plain list keeps every occurrence, which amounts to the same thing.
	unweighted := &NameGenerator{
		Type:    namegen.Simple,
		Entries: []*WeightedStringOption{entry(1, "Zed"), entry(1, "alpha"), entry(1, "Zed")},
	}
	data = unweighted.data()
	c.Equal(map[string]int{"Zed": 2, "alpha": 1}, data, "each occurrence in a plain list counts once")
	reloaded, text = saveAndReload(t, c, unweighted)
	c.Contains(text, `"training_data"`, "unit weights are still written as the plain list")
	c.Equal(3, len(reloaded.Entries), "the plain list keeps every occurrence")
	c.Equal(data, reloaded.data(), "the distribution survives the plain list form")
}

// TestNameGeneratorRoundTripBuiltIn verifies that a generator using a built-in training data set writes only the
// built-in key and no lists.
func TestNameGeneratorRoundTripBuiltIn(t *testing.T) {
	c := check.New(t)
	n := &NameGenerator{Type: namegen.Simple, BuiltIn: namegen.AmericanLast}
	reloaded, text := saveAndReload(t, c, n)
	c.Contains(text, `"built_in_training_data": "american_last"`, "the built-in selection is written by key")
	c.NotContains(text, `"training_data"`, "no plain list is written")
	c.NotContains(text, `"weighted_training_data"`, "no weighted map is written")
	c.Equal(namegen.AmericanLast, reloaded.BuiltIn)
	c.Equal(0, len(reloaded.Entries), "no entries are read back")
}

// TestNameGeneratorRoundTripCompound verifies that a compound generator with a separator and two children round-trips
// with the children's entries intact.
func TestNameGeneratorRoundTripCompound(t *testing.T) {
	c := check.New(t)
	n := &NameGenerator{
		Type:      namegen.Compound,
		Separator: "-",
		NoLowered: true,
		Compound: []*NameGenerator{
			{Type: namegen.Simple, Entries: []*WeightedStringOption{entry(1, "Alpha"), entry(1, "Beta")}},
			{Type: namegen.MarkovLetter, Depth: 2, Entries: []*WeightedStringOption{entry(2, "Gamma"), entry(1, "Delta")}},
		},
	}
	reloaded, text := saveAndReload(t, c, n)
	c.Contains(text, `"separator": "-"`, "the separator is written")
	c.Contains(text, `"compound"`, "the children are written")
	c.Equal(namegen.Compound, reloaded.Type)
	c.Equal("-", reloaded.Separator)
	c.True(reloaded.NoLowered, "flags round-trip")
	c.Equal(2, len(reloaded.Compound), "both children should be present")
	first := reloaded.Compound[0]
	c.Equal(namegen.Simple, first.Type)
	c.Equal(2, len(first.Entries))
	c.Equal("Alpha", first.Entries[0].Value)
	c.Equal("Beta", first.Entries[1].Value)
	second := reloaded.Compound[1]
	c.Equal(namegen.MarkovLetter, second.Type)
	c.Equal(2, second.Depth)
	c.Equal(2, len(second.Entries))
	c.Equal("Delta", second.Entries[0].Value, "weighted entries come back sorted by name")
	c.Equal(1, second.Entries[0].Weight)
	c.Equal("Gamma", second.Entries[1].Value)
	c.Equal(2, second.Entries[1].Weight)
}

// TestNameGeneratorLegacyFiles verifies that each of the existing on-disk forms still loads and still produces a
// working namer.
func TestNameGeneratorLegacyFiles(t *testing.T) {
	c := check.New(t)
	fileSystem := fstest.MapFS{
		"weighted.names": {Data: []byte(`{
	"type": "markov_letter",
	"weighted_training_data": {
		"Zelda": 3,
		"Anna": 1,
		"maria": 2
	}
}`)},
		"unweighted.names": {Data: []byte(`{
	"type": "simple",
	"training_data": [
		"Zelda",
		"Anna",
		"Maria"
	]
}`)},
	}

	weighted, err := ReadNameGeneratorFromFS(fileSystem, "weighted.names")
	c.NoError(err, "a weighted file loads")
	c.Equal(namegen.MarkovLetter, weighted.Type)
	c.Equal(3, len(weighted.Entries))
	c.Equal("Anna", weighted.Entries[0].Value)
	c.Equal(1, weighted.Entries[0].Weight)
	c.Equal("maria", weighted.Entries[1].Value)
	c.Equal(2, weighted.Entries[1].Weight)
	c.Equal("Zelda", weighted.Entries[2].Value)
	c.Equal(3, weighted.Entries[2].Weight)

	unweighted, err := ReadNameGeneratorFromFS(fileSystem, "unweighted.names")
	c.NoError(err, "an unweighted file loads")
	c.Equal(3, len(unweighted.Entries))
	for i, expected := range []string{"Zelda", "Anna", "Maria"} {
		c.Equal(expected, unweighted.Entries[i].Value, "file order is preserved (%d)", i)
		c.Equal(1, unweighted.Entries[i].Weight)
	}

	embedded, err := ReadNameGeneratorFromFS(embeddedFS, "embedded_data/Human Last.names")
	c.NoError(err, "the embedded file loads")
	c.Equal(namegen.UnweightedAmericanLast, embedded.BuiltIn)
	c.Equal(0, len(embedded.Entries))

	for _, path := range []string{"weighted.names", "unweighted.names"} {
		var generator *NameGenerator
		generator, err = NewNameGeneratorFromFS(fileSystem, path)
		c.NoError(err, "%s should produce a working generator", path)
		c.NotEqual("", generator.GenerateName(), "%s should generate a name", path)
	}
	generator, err := NewNameGeneratorFromFS(embeddedFS, "embedded_data/Human Last.names")
	c.NoError(err, "the embedded file should produce a working generator")
	c.NotEqual("", generator.GenerateName(), "the embedded file should generate a name")
}

// sampleCompoundGenerator builds a compound generator with two children for the cloning and hashing tests.
func sampleCompoundGenerator() *NameGenerator {
	return &NameGenerator{
		Type:      namegen.Compound,
		Separator: " ",
		KeyPrefix: "root",
		Compound: []*NameGenerator{
			{
				Type:      namegen.Simple,
				KeyPrefix: "first",
				Entries: []*WeightedStringOption{
					{Weight: 1, Value: "Alpha", KeyPrefix: "first-alpha"},
					{Weight: 2, Value: "Beta", KeyPrefix: "first-beta"},
				},
			},
			{
				Type:      namegen.MarkovLetter,
				Depth:     2,
				KeyPrefix: "second",
				Entries:   []*WeightedStringOption{{Weight: 1, Value: "Gamma", KeyPrefix: "second-gamma"}},
			},
		},
	}
}

// TestNameGeneratorClone verifies that Clone is deep, preserves the key prefixes, and is nil-safe.
func TestNameGeneratorClone(t *testing.T) {
	c := check.New(t)
	c.Nil((*NameGenerator)(nil).Clone(), "cloning nil yields nil")

	original := sampleCompoundGenerator()
	c.NoError(original.createNamer(), "the sample should be usable")
	clone := original.Clone()
	c.Nil(clone.namer, "the clone does not inherit the prepared namer")
	c.Equal("root", clone.KeyPrefix, "the key prefix is preserved")
	c.Equal("first", clone.Compound[0].KeyPrefix, "child key prefixes are preserved")
	c.Equal("first-beta", clone.Compound[0].Entries[1].KeyPrefix, "entry key prefixes are preserved")
	c.True(clone.Compound[0] != original.Compound[0], "children are copied, not shared")
	c.True(clone.Compound[0].Entries[0] != original.Compound[0].Entries[0], "entries are copied, not shared")
	c.Equal(Hash64(original), Hash64(clone), "the clone hashes the same as the original")

	clone.Compound[0].Entries[0].Value = "Changed"
	clone.Compound[1].Depth = 5
	clone.Compound = append(clone.Compound, &NameGenerator{Type: namegen.Simple})
	c.Equal("Alpha", original.Compound[0].Entries[0].Value, "changing a clone's entry leaves the original alone")
	c.Equal(2, original.Compound[1].Depth, "changing a clone's child leaves the original alone")
	c.Equal(2, len(original.Compound), "adding to a clone leaves the original alone")
}

// TestNameGeneratorHash verifies that the hash ignores the runtime-only key prefixes and reflects every piece of data
// that is written to disk.
func TestNameGeneratorHash(t *testing.T) {
	c := check.New(t)
	base := Hash64(sampleCompoundGenerator())

	prefixed := sampleCompoundGenerator()
	provider, _ := keyPrefixCounter(0)
	prefixed.ResetTargetKeyPrefixes(provider)
	c.Equal(base, Hash64(prefixed), "key prefixes do not take part in the hash")

	changes := map[string]func(n *NameGenerator){
		"type":        func(n *NameGenerator) { n.Compound[0].Type = namegen.MarkovRun },
		"flag":        func(n *NameGenerator) { n.NoFirstToUpper = true },
		"separator":   func(n *NameGenerator) { n.Separator = "-" },
		"depth":       func(n *NameGenerator) { n.Compound[1].Depth = 4 },
		"weight":      func(n *NameGenerator) { n.Compound[0].Entries[0].Weight = 5 },
		"value":       func(n *NameGenerator) { n.Compound[0].Entries[1].Value = "Zeta" },
		"child added": func(n *NameGenerator) { n.Compound = append(n.Compound, &NameGenerator{Type: namegen.Simple}) },
		"built-in":    func(n *NameGenerator) { n.Compound[0].BuiltIn = namegen.AmericanMale },
	}
	for what, change := range changes {
		n := sampleCompoundGenerator()
		change(n)
		c.NotEqual(base, Hash64(n), "changing the %s changes the hash", what)
	}
}

// TestNameGeneratorKeyPrefixesAndNormalize verifies that ResetTargetKeyPrefixes reaches every piece of data with a
// unique prefix while skipping nil entries and children, and that Normalize removes those nils.
func TestNameGeneratorKeyPrefixesAndNormalize(t *testing.T) {
	c := check.New(t)
	n := sampleCompoundGenerator()
	n.Compound[0].Entries = append(n.Compound[0].Entries, nil)
	n.Compound = append(n.Compound, nil, &NameGenerator{
		Type: namegen.Compound,
		Compound: []*NameGenerator{
			nil,
			{Type: namegen.Simple, Entries: []*WeightedStringOption{nil, entry(1, "Deep")}},
		},
	})

	provider, count := keyPrefixCounter(0)
	n.ResetTargetKeyPrefixes(provider)
	seen := make(map[string]bool)
	var collect func(n *NameGenerator)
	collect = func(n *NameGenerator) {
		c.NotEqual("", n.KeyPrefix, "every generator gets a prefix")
		c.False(seen[n.KeyPrefix], "prefix %s is unique", n.KeyPrefix)
		seen[n.KeyPrefix] = true
		for _, one := range n.Entries {
			if one != nil {
				c.NotEqual("", one.KeyPrefix, "every entry gets a prefix")
				c.False(seen[one.KeyPrefix], "prefix %s is unique", one.KeyPrefix)
				seen[one.KeyPrefix] = true
			}
		}
		for _, one := range n.Compound {
			if one != nil {
				collect(one)
			}
		}
	}
	collect(n)
	// The root, its three non-nil children, the nested child, and the four non-nil entries across them.
	c.Equal(9, len(seen), "each piece of data gets exactly one prefix")
	c.Equal(9, *count, "the provider is called once per piece of data")

	n.Normalize()
	c.Equal(3, len(n.Compound), "the nil child is dropped")
	c.Equal(2, len(n.Compound[0].Entries), "the nil entry is dropped")
	c.Equal(1, len(n.Compound[2].Compound), "nested nil children are dropped")
	c.Equal(1, len(n.Compound[2].Compound[0].Entries), "nested nil entries are dropped")
	c.Equal("Deep", n.Compound[2].Compound[0].Entries[0].Value)
}

// TestNameGeneratorSampleNames verifies that SampleNames produces names from each generator type, reports unusable
// definitions with an error, and never leaves a namer behind on the generator it was called on.
func TestNameGeneratorSampleNames(t *testing.T) {
	c := check.New(t)

	simple := &NameGenerator{Type: namegen.Simple, Entries: []*WeightedStringOption{entry(1, "ALICE"), entry(3, "bob")}}
	samples, err := simple.SampleNames(10)
	c.NoError(err)
	c.Nil(simple.namer, "sampling leaves the generator untouched")
	c.Equal(10, len(samples))
	for _, one := range samples {
		c.True(one == "Alice" || one == "Bob", "simple names are the entries with case adjusted: %q", one)
	}
	simple.NoLowered = true
	simple.NoFirstToUpper = true
	samples, err = simple.SampleNames(10)
	c.NoError(err)
	for _, one := range samples {
		c.True(one == "ALICE" || one == "bob", "simple names are the entries as-is when the flags are set: %q", one)
	}

	// An entry with no weight or no name, such as a row an editor has just added, contributes nothing.
	padded := &NameGenerator{
		Type:    namegen.Simple,
		Entries: []*WeightedStringOption{entry(1, "Carol"), entry(0, "Dave"), entry(1, ""), entry(1, " \t ")},
	}
	samples, err = padded.SampleNames(10)
	c.NoError(err)
	for _, one := range samples {
		c.Equal("Carol", one, "zero-weight, empty and whitespace-only entries are skipped")
	}
	// A name that is nothing but whitespace is no name at all: the namer would trim it away, so it must not count as
	// training data, or the samples would come back blank rather than with an explanation of what is missing.
	for _, kind := range []namegen.Type{namegen.Simple, namegen.MarkovLetter, namegen.MarkovRun} {
		blank := &NameGenerator{Type: kind, Entries: []*WeightedStringOption{entry(1, ""), entry(1, " \t ")}}
		samples, err = blank.SampleNames(1)
		c.HasError(err, "%s with only empty and whitespace entries is the same as no training data", kind.Key())
		c.Nil(samples, "%s produced samples without training data", kind.Key())
		if err != nil {
			c.Contains(err.Error(), "no training data has been provided", "%s says what is missing", kind.Key())
		}
	}

	entries := []*WeightedStringOption{
		entry(1, "Alexander"), entry(1, "Benjamin"), entry(2, "Catherine"), entry(1, "Dominic"), entry(1, "Eleanor"),
	}
	for _, kind := range []namegen.Type{namegen.MarkovLetter, namegen.MarkovRun} {
		n := &NameGenerator{Type: kind, Entries: entries}
		samples, err = n.SampleNames(10)
		c.NoError(err, "%s should generate names", kind.Key())
		c.Nil(n.namer, "sampling leaves the %s generator untouched", kind.Key())
		c.Equal(10, len(samples))
		for _, one := range samples {
			c.NotEqual("", one, "%s names are not empty", kind.Key())
		}
	}

	compound := &NameGenerator{
		Type:      namegen.Compound,
		Separator: "-",
		Compound: []*NameGenerator{
			{Type: namegen.Simple, Entries: []*WeightedStringOption{entry(1, "Alpha")}},
			{Type: namegen.Simple, Entries: []*WeightedStringOption{entry(1, "Beta")}},
		},
	}
	samples, err = compound.SampleNames(10)
	c.NoError(err)
	c.Nil(compound.namer, "sampling leaves the compound generator untouched")
	c.Nil(compound.Compound[0].namer, "sampling leaves the children untouched")
	c.Equal(10, len(samples))
	for _, one := range samples {
		c.Contains(one, "-", "compound names contain the separator")
	}

	emptyCompound := &NameGenerator{Type: namegen.Compound}
	samples, err = emptyCompound.SampleNames(10)
	c.HasError(err, "a compound generator with no children cannot generate names")
	c.Nil(samples)
	c.Nil(emptyCompound.namer, "a failed sample leaves no namer behind")

	emptySimple := &NameGenerator{Type: namegen.Simple}
	samples, err = emptySimple.SampleNames(10)
	c.HasError(err, "a simple generator with no entries cannot generate names")
	c.Nil(samples)
	c.Nil(emptySimple.namer, "a failed sample leaves no namer behind")

	// A broken child is reported by position, so the message tells the user where to look.
	broken := &NameGenerator{
		Type: namegen.Compound,
		Compound: []*NameGenerator{
			{Type: namegen.Simple, Entries: []*WeightedStringOption{entry(1, "Alpha")}},
			{Type: namegen.Simple},
		},
	}
	_, err = broken.SampleNames(1)
	c.HasError(err)
	c.Contains(err.Error(), "compound entry #2", "the failing child is identified")

	// Reading an incomplete file still works, even though preparing it for use does not.
	fileSystem := fstest.MapFS{"empty.names": {Data: []byte(`{"type": "compound"}`)}}
	_, err = ReadNameGeneratorFromFS(fileSystem, "empty.names")
	c.NoError(err, "an incomplete file can be opened for editing")
	_, err = NewNameGeneratorFromFS(fileSystem, "empty.names")
	c.HasError(err, "an incomplete file cannot be loaded for use")
}

// TestNewNameGenerator verifies the scaffolding a fresh generator starts with and its minimal serialized form.
func TestNewNameGenerator(t *testing.T) {
	c := check.New(t)
	n := NewNameGenerator()
	c.Equal(namegen.Simple, n.Type)
	c.Equal(0, len(n.Entries))
	c.Equal(0, len(n.Compound))
	c.Equal(namegen.None, n.BuiltIn)
	data, err := jio.Marshal(n)
	c.NoError(err)
	compact := strings.Join(strings.Fields(string(data)), "")
	c.Equal(`{"type":"simple"}`, compact, "a fresh generator writes only its type")
}

// TestNameGeneratorReusesTrainedNamer verifies that a generator built for a definition identical to one built before
// reuses its namer rather than training another, whichever generator asks, while any change to the definition trains
// a new one. The editor relies on this to keep a compound of built-in sets responsive as it is edited.
func TestNameGeneratorReusesTrainedNamer(t *testing.T) {
	c := check.New(t)
	a := &NameGenerator{Type: namegen.Simple, Entries: []*WeightedStringOption{entry(1, "Alice"), entry(2, "Bob")}}
	c.NoError(a.createNamer())
	b := a.Clone()
	b.KeyPrefix = "ignored"
	c.NoError(b.createNamer())
	c.True(a.namer == b.namer, "an identical definition shares the namer, whatever its key prefix")

	b.Entries[1].Weight = 3
	c.NoError(b.createNamer())
	c.True(a.namer != b.namer, "a changed definition trains a namer of its own")

	// A compound assembles its namer afresh each time, from the leaves' shared namers.
	compound := &NameGenerator{Type: namegen.Compound, Separator: " ", Compound: []*NameGenerator{a.Clone(), b.Clone()}}
	c.NoError(compound.createNamer())
	c.True(compound.Compound[0].namer == a.namer)
	c.True(compound.Compound[1].namer == b.namer)
	names, err := compound.SampleNames(3)
	c.NoError(err)
	c.Equal(3, len(names))
}
