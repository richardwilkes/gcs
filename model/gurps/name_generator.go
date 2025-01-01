// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"context"
	"io/fs"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/namegen"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/rpgtools/names"
	"github.com/richardwilkes/rpgtools/names/namesets/american"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xmath/rand"
)

var _ names.Namer = &NameGenerator{}

// NameGeneratorRef holds a reference to a NameGenerator.
type NameGeneratorRef struct {
	FileRef   *NamedFileRef
	generator *NameGenerator
}

// TrainingData is only valid when Type is not namegen.Compound. Only one will be used, and they are checked
// in the order listed here.
type TrainingData struct {
	BuiltIn    namegen.Builtin `json:"built_in_training_data,omitempty"`
	Weighted   map[string]int  `json:"weighted_training_data,omitempty"`
	Unweighted []string        `json:"training_data,omitempty"`
}

func (t *TrainingData) data() map[string]int {
	switch t.BuiltIn {
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
		if len(t.Weighted) != 0 {
			return t.Weighted
		}
		unweighted := make(map[string]int, len(t.Unweighted))
		for _, k := range t.Unweighted {
			unweighted[k] = 1
		}
		return unweighted
	}
}

func toUnweighted(data map[string]int) map[string]int {
	unweighted := make(map[string]int, len(data))
	for k := range data {
		unweighted[k] = 1
	}
	return unweighted
}

// NameGenerator holds the data necessary to create a Namer.
type NameGenerator struct {
	Type           namegen.Type     `json:"type"`
	NoLowered      bool             `json:"no_lowered,omitempty"`
	NoFirstToUpper bool             `json:"no_first_to_upper,omitempty"`
	Separator      string           `json:"separator,omitempty"` // Only valid for namegen.Compound
	Depth          int              `json:"depth,omitempty"`     // Only valid for namegen.MarkovLetter
	Compound       []*NameGenerator `json:"compound,omitempty"`  // Only valid for namegen.Compound
	TrainingData
	namer names.Namer
}

// AvailableNameGenerators scans the libraries and returns the available name generators.
func AvailableNameGenerators(libraries Libraries) []*NameGeneratorRef {
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
		return txt.NaturalCmp(a.FileRef.Name, b.FileRef.Name, true)
	})
	return list
}

// NewNameGeneratorFromFS creates a new NameGenerator from a file.
func NewNameGeneratorFromFS(fileSystem fs.FS, filePath string) (*NameGenerator, error) {
	var generator NameGenerator
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &generator); err != nil {
		return nil, err
	}
	if err := generator.createNamer(); err != nil {
		return nil, err
	}
	return &generator, nil
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
func (n *NameGenerator) GenerateNameWithRandomizer(rnd rand.Randomizer) string {
	return n.namer.GenerateNameWithRandomizer(rnd)
}

func (n *NameGenerator) createNamer() error {
	n.namer = nil
	if n.Type == namegen.Compound {
		if len(n.Compound) == 0 {
			return errs.New("no name generators specified for " + n.Type.String() + " generation type")
		}
		var namers []names.Namer
		for _, one := range n.Compound {
			if err := one.createNamer(); err != nil {
				return err
			}
			namers = append(namers, one.namer)
		}
		n.namer = names.NewCompoundNamer(n.Separator, !n.NoLowered, !n.NoFirstToUpper, namers...)
		return nil
	}
	data := n.data()
	if len(data) == 0 {
		return errs.New("invalid training data")
	}
	switch n.Type {
	case namegen.Simple:
		n.namer = names.NewSimpleNamer(data, !n.NoLowered, !n.NoFirstToUpper)
		return nil
	case namegen.MarkovLetter:
		depth := n.Depth
		if depth < 1 {
			depth = 3
		} else if depth > 5 {
			depth = 5
		}
		n.namer = names.NewMarkovLetterNamer(depth, data, !n.NoLowered, !n.NoFirstToUpper)
		return nil
	case namegen.MarkovRun:
		n.namer = names.NewMarkovRunNamer(data, !n.NoLowered, !n.NoFirstToUpper)
		return nil
	default:
		return errs.New("invalid name generator type")
	}
}
