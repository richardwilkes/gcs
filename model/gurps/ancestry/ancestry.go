/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ancestry

import (
	"context"
	"io/fs"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps/measure"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/eval"
	"github.com/richardwilkes/toolbox/log/jot"
	xfs "github.com/richardwilkes/toolbox/xio/fs"
)

// Default holds the name of the default ancestry.
const Default = "Human"

// Ancestry holds details necessary to generate ancestry-specific customizations.
type Ancestry struct {
	Name          string                     `json:"name,omitempty"`
	CommonOptions *Options                   `json:"common_options,omitempty"`
	GenderOptions []*WeightedAncestryOptions `json:"gender_options,omitempty"`
}

// AvailableAncestries scans the libraries and returns the available ancestries.
func AvailableAncestries(libraries library.Libraries) []*library.NamedFileSet {
	return library.ScanForNamedFileSets(embeddedFS, "data", true, libraries, ".ancestry")
}

// Lookup an Ancestry by name.
func Lookup(name string, libraries library.Libraries) *Ancestry {
	for _, lib := range AvailableAncestries(libraries) {
		for _, one := range lib.List {
			if one.Name == name {
				if a, err := NewAncestryFromFile(one.FileSystem, one.FilePath); err != nil {
					jot.Warn(err)
				} else {
					return a
				}
			}
		}
	}
	return nil
}

// NewAncestryFromFile creates a new Ancestry from a file.
func NewAncestryFromFile(fileSystem fs.FS, filePath string) (*Ancestry, error) {
	var ancestry Ancestry
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &ancestry); err != nil {
		return nil, err
	}
	if ancestry.Name == "" {
		ancestry.Name = xfs.BaseName(filePath)
	}
	return &ancestry, nil
}

// Save writes the Ancestry to the file as JSON.
func (a *Ancestry) Save(filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, a)
}

// RandomGender returns a randomized gender.
func (a *Ancestry) RandomGender(not string) string {
	if choice := ChooseWeightedAncestryOptions(a.GenderOptions, func(o *Options) bool {
		return o.Name == not
	}); choice != nil {
		return choice.Name
	}
	return ""
}

// GenderedOptions returns the options for the specified gender, or nil.
func (a *Ancestry) GenderedOptions(gender string) *Options {
	gender = strings.TrimSpace(gender)
	for _, one := range a.GenderOptions {
		if strings.EqualFold(one.Value.Name, gender) {
			return one.Value
		}
	}
	return nil
}

// RandomHeight returns a randomized height.
func (a *Ancestry) RandomHeight(resolver eval.VariableResolver, gender string, not measure.Length) measure.Length {
	if options := a.GenderedOptions(gender); options != nil && options.HeightFormula != "" {
		return options.RandomHeight(resolver, not)
	}
	if a.CommonOptions != nil && a.CommonOptions.HeightFormula != "" {
		return a.CommonOptions.RandomHeight(resolver, not)
	}
	return measure.LengthFromInteger(defaultHeight, measure.Inch)
}

// RandomWeight returns a randomized weight.
func (a *Ancestry) RandomWeight(resolver eval.VariableResolver, gender string, not measure.Weight) measure.Weight {
	if options := a.GenderedOptions(gender); options != nil && options.WeightFormula != "" {
		return options.RandomWeight(resolver, not)
	}
	if a.CommonOptions != nil && a.CommonOptions.WeightFormula != "" {
		return a.CommonOptions.RandomWeight(resolver, not)
	}
	return measure.WeightFromInteger(defaultWeight, measure.Pound)
}

// RandomAge returns a randomized age.
func (a *Ancestry) RandomAge(resolver eval.VariableResolver, gender string, not int) int {
	if options := a.GenderedOptions(gender); options != nil && options.AgeFormula != "" {
		return options.RandomAge(resolver, not)
	}
	if a.CommonOptions != nil && a.CommonOptions.AgeFormula != "" {
		return a.CommonOptions.RandomAge(resolver, not)
	}
	return defaultAge
}

// RandomHair returns a randomized hair.
func (a *Ancestry) RandomHair(gender, not string) string {
	if options := a.GenderedOptions(gender); options != nil && len(options.HairOptions) != 0 {
		return options.RandomHair(not)
	}
	if a.CommonOptions != nil && len(a.CommonOptions.HairOptions) != 0 {
		return a.CommonOptions.RandomHair(not)
	}
	return defaultHair
}

// RandomEyes returns a randomized eyes.
func (a *Ancestry) RandomEyes(gender, not string) string {
	if options := a.GenderedOptions(gender); options != nil && len(options.EyeOptions) != 0 {
		return options.RandomEye(not)
	}
	if a.CommonOptions != nil && len(a.CommonOptions.EyeOptions) != 0 {
		return a.CommonOptions.RandomEye(not)
	}
	return defaultEye
}

// RandomSkin returns a randomized skin.
func (a *Ancestry) RandomSkin(gender, not string) string {
	if options := a.GenderedOptions(gender); options != nil && len(options.SkinOptions) != 0 {
		return options.RandomSkin(not)
	}
	if a.CommonOptions != nil && len(a.CommonOptions.SkinOptions) != 0 {
		return a.CommonOptions.RandomSkin(not)
	}
	return defaultSkin
}

// RandomHandedness returns a randomized handedness.
func (a *Ancestry) RandomHandedness(gender, not string) string {
	if options := a.GenderedOptions(gender); options != nil && len(options.HandednessOptions) != 0 {
		return options.RandomHandedness(not)
	}
	if a.CommonOptions != nil && len(a.CommonOptions.HandednessOptions) != 0 {
		return a.CommonOptions.RandomHandedness(not)
	}
	return defaultHandedness
}

// RandomName returns a randomized name.
func (a *Ancestry) RandomName(nameGeneratorRefs []*NameGeneratorRef, gender string) string {
	if options := a.GenderedOptions(gender); options != nil && len(options.NameGenerators) != 0 {
		return options.RandomName(nameGeneratorRefs)
	}
	if a.CommonOptions != nil && len(a.CommonOptions.NameGenerators) != 0 {
		return a.CommonOptions.RandomName(nameGeneratorRefs)
	}
	return ""
}
