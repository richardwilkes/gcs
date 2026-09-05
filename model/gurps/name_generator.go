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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"hash"
	"io/fs"
	"slices"
	"strings"
	"sync"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/namegen"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/rpgtools/names"
	"github.com/richardwilkes/rpgtools/names/namesets/american"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xrand"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

var (
	_ names.Namer = &NameGenerator{}
	_ Hashable    = &NameGenerator{}
)

// NameGeneratorRef holds a reference to a NameGenerator.
type NameGeneratorRef struct {
	FileRef   *NamedFileRef
	generator *NameGenerator
}

// NameGenerator holds the data necessary to create a Namer. Its JSON form is produced by MarshalJSONTo and consumed by
// UnmarshalJSONFrom, so the field tags that would otherwise shape it are absent here; see nameGeneratorData for the
// on-disk layout.
type NameGenerator struct {
	Type           namegen.Type
	NoLowered      bool
	NoFirstToUpper bool
	Separator      string           // Only valid for namegen.Compound
	Depth          int              // Only valid for namegen.MarkovLetter
	Compound       []*NameGenerator // Only valid for namegen.Compound
	// BuiltIn selects one of the built-in training data sets. When it is anything other than namegen.None, it is used
	// in preference to Entries. Only valid when Type is not namegen.Compound.
	BuiltIn namegen.Builtin
	// Entries is the custom training data: each name with its weight. It is written as "training_data" (a plain list,
	// preserving order) when every weight is 1 and as "weighted_training_data" (a map) otherwise. Only valid when Type
	// is not namegen.Compound.
	Entries []*WeightedStringOption
	// KeyPrefix is a runtime-only key the editor uses to give this generator's widgets a stable identity. It is never
	// written to disk.
	KeyPrefix string
	namer     names.Namer
}

// nameGeneratorData is the on-disk form of a NameGenerator. Files carry neither a version nor a wrapper, and this
// layout must not change, as it is what every existing .names file is written in. Of the two training data lists, only
// one is used: "weighted_training_data" when it is non-empty, otherwise "training_data".
type nameGeneratorData struct {
	Type           namegen.Type     `json:"type"`
	NoLowered      bool             `json:"no_lowered,omitzero"`
	NoFirstToUpper bool             `json:"no_first_to_upper,omitzero"`
	Separator      string           `json:"separator,omitzero"`
	Depth          int              `json:"depth,omitzero"`
	Compound       []*NameGenerator `json:"compound,omitempty"`
	BuiltIn        namegen.Builtin  `json:"built_in_training_data,omitzero"`
	Weighted       map[string]int   `json:"weighted_training_data,omitempty"`
	Unweighted     []string         `json:"training_data,omitempty"`
}

// AvailableNameGenerators scans the libraries and returns the available name generators.
func AvailableNameGenerators(libraries *Libraries) []*NameGeneratorRef {
	var list []*NameGeneratorRef
	seen := make(map[string]bool)
	for _, set := range ScanForNamedFileSets(embeddedFS, "embedded_data", true, libraries, NamesExt) {
		for _, one := range set.List {
			if seen[one.Name] {
				continue
			}
			seen[one.Name] = true
			list = append(list, &NameGeneratorRef{FileRef: one})
		}
	}
	slices.SortFunc(list, func(a, b *NameGeneratorRef) int {
		return xstrings.NaturalCmp(a.FileRef.Name, b.FileRef.Name, true)
	})
	return list
}

// NewNameGenerator returns a new, empty name generator of the simple type, ready to be filled in by an editor.
func NewNameGenerator() *NameGenerator {
	return &NameGenerator{Type: namegen.Simple}
}

// NewNameGeneratorFromFS creates a new NameGenerator from a file and prepares it to generate names. Use
// ReadNameGeneratorFromFS instead when the file is being opened for editing rather than for use, since a file that is
// incomplete (a compound generator with no children, say) is an error here but perfectly reasonable to edit.
func NewNameGeneratorFromFS(fileSystem fs.FS, filePath string) (*NameGenerator, error) {
	generator, err := ReadNameGeneratorFromFS(fileSystem, filePath)
	if err != nil {
		return nil, err
	}
	if err = generator.createNamer(); err != nil {
		return nil, err
	}
	return generator, nil
}

// ReadNameGeneratorFromFS loads a NameGenerator from a file without preparing it to generate names, so that files
// whose definitions are incomplete can still be opened for editing. Normalize is not called, so null entries the file
// may contain are preserved as-is; an editor should call it before binding to the data. Use SampleNames to try the
// definition out, or NewNameGeneratorFromFS to load a generator for use.
func ReadNameGeneratorFromFS(fileSystem fs.FS, filePath string) (*NameGenerator, error) {
	var generator NameGenerator
	if err := jio.LoadFromFile(fileSystem, filePath, &generator); err != nil {
		return nil, err
	}
	return &generator, nil
}

// Save writes the NameGenerator to the file as JSON. Matching the existing files, there is no version wrapper.
func (n *NameGenerator) Save(filePath string) error {
	return jio.SaveToFile(filePath, n)
}

// MarshalJSONTo implements json.MarshalerTo. Entries are written as "training_data" when every weight is 1 and as
// "weighted_training_data" otherwise; in the weighted form, the weights of a name that appears more than once add
// together, which is also what the plain list's repeated names amount to. Nil entries are skipped, and no training data
// key is written when there are none to write.
func (n *NameGenerator) MarshalJSONTo(enc *jsontext.Encoder) error {
	data := nameGeneratorData{
		Type:           n.Type,
		NoLowered:      n.NoLowered,
		NoFirstToUpper: n.NoFirstToUpper,
		Separator:      n.Separator,
		Depth:          n.Depth,
		Compound:       n.Compound,
		BuiltIn:        n.BuiltIn,
	}
	entries := make([]*WeightedStringOption, 0, len(n.Entries))
	allUnit := true
	for _, one := range n.Entries {
		if one != nil {
			entries = append(entries, one)
			if one.Weight != 1 {
				allUnit = false
			}
		}
	}
	switch {
	case len(entries) == 0:
	case allUnit:
		data.Unweighted = make([]string, len(entries))
		for i, one := range entries {
			data.Unweighted[i] = one.Value
		}
	default:
		data.Weighted = make(map[string]int, len(entries))
		for _, one := range entries {
			data.Weighted[one.Value] += one.Weight
		}
	}
	return json.MarshalEncode(enc, &data)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom. A non-empty "weighted_training_data" map becomes the Entries,
// sorted by name since the map carries no order of its own; otherwise "training_data" becomes the Entries, each with a
// weight of 1, in file order. The runtime-only KeyPrefix is left alone.
func (n *NameGenerator) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var data nameGeneratorData
	if err := json.UnmarshalDecode(dec, &data); err != nil {
		return err
	}
	n.Type = data.Type
	n.NoLowered = data.NoLowered
	n.NoFirstToUpper = data.NoFirstToUpper
	n.Separator = data.Separator
	n.Depth = data.Depth
	n.Compound = data.Compound
	n.BuiltIn = data.BuiltIn
	n.namer = nil
	switch {
	case len(data.Weighted) != 0:
		n.Entries = make([]*WeightedStringOption, 0, len(data.Weighted))
		for k, v := range data.Weighted {
			n.Entries = append(n.Entries, &WeightedStringOption{Weight: v, Value: k})
		}
		slices.SortFunc(n.Entries, func(a, b *WeightedStringOption) int {
			return xstrings.NaturalCmp(a.Value, b.Value, true)
		})
	case len(data.Unweighted) != 0:
		n.Entries = make([]*WeightedStringOption, len(data.Unweighted))
		for i, one := range data.Unweighted {
			n.Entries[i] = &WeightedStringOption{Weight: 1, Value: one}
		}
	default:
		n.Entries = nil
	}
	return nil
}

// Hash writes this object's contents into the hasher. The hash is of the JSON content that Save would write, so it
// answers exactly "has the file content changed", and the runtime-only KeyPrefix fields never take part.
func (n *NameGenerator) Hash(h hash.Hash) {
	HashJSON(h, n)
}

// Clone returns a deep copy of this generator, or nil if this generator is nil. The runtime-only KeyPrefix fields are
// copied along with everything else, while the prepared namer is not: the copy has none until createNamer is called
// on it. Nil entries and nil compound children are preserved as nil, so the copy is faithful to the original;
// Normalize is what removes them.
func (n *NameGenerator) Clone() *NameGenerator {
	if n == nil {
		return nil
	}
	clone := *n
	clone.namer = nil
	if n.Compound != nil {
		clone.Compound = make([]*NameGenerator, len(n.Compound))
		for i, one := range n.Compound {
			clone.Compound[i] = one.Clone()
		}
	}
	clone.Entries = cloneWeightedStringOptions(n.Entries)
	return &clone
}

// ResetTargetKeyPrefixes assigns new key prefixes to this generator, each of its entries, and each of its compound
// children, recursively. Nil entries and children are skipped.
func (n *NameGenerator) ResetTargetKeyPrefixes(prefixProvider func() string) {
	n.KeyPrefix = prefixProvider()
	for _, one := range n.Entries {
		if one != nil {
			one.KeyPrefix = prefixProvider()
		}
	}
	for _, one := range n.Compound {
		if one != nil {
			one.ResetTargetKeyPrefixes(prefixProvider)
		}
	}
}

// Normalize makes the generator safe to edit: it drops the null entries and null compound children a hand-written file
// may contain, recursively. Name generation tolerates them, but an editor needs every pointer it renders to be non-nil.
// This is deliberately not called when loading a file, so that the load path keeps tolerating such files as-is.
func (n *NameGenerator) Normalize() {
	n.Entries = dropNilWeightedStringOptions(n.Entries)
	n.Compound = slices.DeleteFunc(n.Compound, func(one *NameGenerator) bool { return one == nil })
	for _, one := range n.Compound {
		one.Normalize()
	}
}

// Generator returns the NameGenerator, loading it if needed.
func (n *NameGeneratorRef) Generator() (*NameGenerator, error) {
	if n.generator == nil {
		var err error
		if n.generator, err = NewNameGeneratorFromFS(n.FileRef.FileSystem, n.FileRef.FilePath); err != nil {
			return nil, err
		}
	}
	return n.generator, nil
}

// GenerateName generates a new random name.
func (n *NameGenerator) GenerateName() string {
	return n.namer.GenerateName()
}

// GenerateNameWithRandomizer generates a new random name using the specified randomizer.
func (n *NameGenerator) GenerateNameWithRandomizer(rnd xrand.Randomizer) string {
	return n.namer.GenerateNameWithRandomizer(rnd)
}

// SampleNames generates count names from the current definition, for trying it out while editing. The work is done on
// a clone, so this generator is left exactly as it was: no namer is prepared on it, and a definition that is only
// partly edited never leaves a stale one behind. An error describes what keeps the definition from generating names,
// in terms suitable for showing to the user.
func (n *NameGenerator) SampleNames(count int) ([]string, error) {
	clone := n.Clone()
	if err := clone.createNamer(); err != nil {
		return nil, err
	}
	result := make([]string, 0, max(count, 0))
	for range count {
		result = append(result, clone.GenerateName())
	}
	return result, nil
}

// data returns the training data to build a namer from: the built-in set when one is selected, otherwise the Entries,
// with the weights of a name that appears more than once added together. Nil entries are skipped.
func (n *NameGenerator) data() map[string]int {
	switch n.BuiltIn {
	case namegen.AmericanMale:
		return american.Male()
	case namegen.AmericanFemale:
		return american.Female()
	case namegen.AmericanLast:
		return american.Last()
	case namegen.UnweightedAmericanMale:
		return toUnweighted(american.Male())
	case namegen.UnweightedAmericanFemale:
		return toUnweighted(american.Female())
	case namegen.UnweightedAmericanLast:
		return toUnweighted(american.Last())
	default:
		// An entry with no weight or no name (whitespace alone counts as no name) contributes nothing, which is what an
		// editor's freshly added, not yet filled in row should do. The weights of a name that appears more than once
		// add together.
		data := make(map[string]int, len(n.Entries))
		for _, one := range n.Entries {
			if one.Valid() && strings.TrimSpace(one.Value) != "" {
				data[one.Value] += one.Weight
			}
		}
		return data
	}
}

func toUnweighted(data map[string]int) map[string]int {
	unweighted := make(map[string]int, len(data))
	for k := range data {
		unweighted[k] = 1
	}
	return unweighted
}

func (n *NameGenerator) createNamer() error {
	return n.createNamerAt("")
}

// createNamerAt builds the namer for this generator, reusing the one already built for an identical definition when
// there is one; see leafNamerCache. 'where' names this generator's position within the enclosing compound generators,
// and is empty for the top-level one; it prefixes each error message so that the message points at the entry that
// needs attention.
func (n *NameGenerator) createNamerAt(where string) error {
	n.namer = nil
	if n.Type == namegen.Compound {
		namers := make([]names.Namer, 0, len(n.Compound))
		for i, one := range n.Compound {
			if one == nil {
				continue
			}
			childWhere := fmt.Sprintf(i18n.Text("compound entry #%d"), i+1)
			if where != "" {
				childWhere = where + ", " + childWhere
			}
			if err := one.createNamerAt(childWhere); err != nil {
				return err
			}
			namers = append(namers, one.namer)
		}
		if len(namers) == 0 {
			return namerError(where,
				i18n.Text("a compound name generator needs at least one name generator to combine"))
		}
		n.namer = names.NewCompoundNamer(n.Separator, !n.NoLowered, !n.NoFirstToUpper, namers...)
		return nil
	}
	key := Hash64(n)
	if namer := cachedLeafNamer(key); namer != nil {
		n.namer = namer
		return nil
	}
	data := n.data()
	if len(data) == 0 {
		return namerError(where, i18n.Text("no training data has been provided"))
	}
	switch n.Type {
	case namegen.Simple:
		n.namer = names.NewSimpleNamer(data, !n.NoLowered, !n.NoFirstToUpper)
	case namegen.MarkovLetter:
		depth := n.Depth
		if depth < 1 {
			depth = 3
		} else if depth > 5 {
			depth = 5
		}
		n.namer = names.NewMarkovLetterNamer(depth, data, !n.NoLowered, !n.NoFirstToUpper)
	case namegen.MarkovRun:
		n.namer = names.NewMarkovRunNamer(data, !n.NoLowered, !n.NoFirstToUpper)
	default:
		return namerError(where, i18n.Text("invalid name generator type"))
	}
	storeLeafNamer(key, n.namer)
	return nil
}

// leafNamerCacheLimit is how many namers leafNamerCache holds before it is emptied. The editor produces a new
// definition with every keystroke, so without a bound the cache would grow for as long as a session lasts.
const leafNamerCacheLimit = 64

// leafNamerCache holds the namers built for generators other than compound ones, keyed by the hash of the definition
// each was built from. Training a namer on one of the built-in sets takes tens of milliseconds, and the editor builds
// namers afresh for every change to the definition it shows, most of which -- a separator, a case option, a compound
// entry being added -- leave the definition of the expensive leaf as it was; the cache lets those changes reuse the
// namer already trained. A namer is immutable once built, so it can be shared freely.
var (
	leafNamerCacheLock sync.Mutex
	leafNamerCache     = make(map[uint64]names.Namer)
)

// cachedLeafNamer returns the namer built for the definition with the given hash, or nil if there is none.
func cachedLeafNamer(key uint64) names.Namer {
	leafNamerCacheLock.Lock()
	defer leafNamerCacheLock.Unlock()
	return leafNamerCache[key]
}

// storeLeafNamer records the namer built for the definition with the given hash, emptying the cache first if it is
// full.
func storeLeafNamer(key uint64, namer names.Namer) {
	leafNamerCacheLock.Lock()
	defer leafNamerCacheLock.Unlock()
	if len(leafNamerCache) >= leafNamerCacheLimit {
		clear(leafNamerCache)
	}
	leafNamerCache[key] = namer
}

// namerError returns an error carrying msg, prefixed by where when it is not empty.
func namerError(where, msg string) error {
	if where != "" {
		return errs.New(where + ": " + msg)
	}
	return errs.New(msg)
}
