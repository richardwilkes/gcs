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
	"hash"
	"io/fs"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
)

var _ Hashable = &Ancestry{}

// DefaultAncestry holds the name of the default ancestry.
const DefaultAncestry = "Human"

// Ancestry holds details necessary to generate ancestry-specific customizations.
type Ancestry struct {
	Name          string                     `json:"name,omitzero"`
	CommonOptions *AncestryOptions           `json:"common_options,omitzero"`
	GenderOptions []*WeightedAncestryOptions `json:"gender_options,omitempty"`
	// KeyPrefix is a runtime-only key the editor uses to give the ancestry's widgets a stable identity. It is never
	// written to disk.
	KeyPrefix string `json:"-"`
}

type ancestryData struct {
	Version int `json:"version"`
	Ancestry
}

// AvailableAncestries scans the libraries and returns the available ancestries.
func AvailableAncestries(libraries *Libraries) []*NamedFileSet {
	return ScanForNamedFileSets(embeddedFS, "embedded_data", true, libraries, AncestryExt)
}

// LookupAncestry an Ancestry by name.
func LookupAncestry(name string, libraries *Libraries) *Ancestry {
	for _, lib := range AvailableAncestries(libraries) {
		for _, one := range lib.List {
			if one.Name == name {
				if a, err := NewAncestryFromFile(one.FileSystem, one.FilePath); err != nil {
					errs.Log(err, "path", one.FilePath)
				} else {
					return a
				}
			}
		}
	}
	return nil
}

// NewAncestry returns a new ancestry with the customary scaffolding: no name, an empty set of common options, and two
// equally weighted genders with no overrides of their own. The gender names are data matched against Profile.Gender,
// so they are not localized.
func NewAncestry() *Ancestry {
	return &Ancestry{
		CommonOptions: &AncestryOptions{},
		GenderOptions: []*WeightedAncestryOptions{
			{Weight: 1, Value: &AncestryOptions{Name: "Male"}},
			{Weight: 1, Value: &AncestryOptions{Name: "Female"}},
		},
	}
}

// NewAncestryFromFile creates a new Ancestry from a file.
func NewAncestryFromFile(fileSystem fs.FS, filePath string) (*Ancestry, error) {
	var ancestry ancestryData
	if err := jio.LoadFromFile(fileSystem, filePath, &ancestry); err != nil {
		return nil, err
	}
	if ancestry.Version == 0 { // for some older files
		ancestry.Version = jio.MinimumDataVersion
	}
	if err := jio.CheckVersion(ancestry.Version); err != nil {
		return nil, err
	}
	if ancestry.Name == "" {
		ancestry.Name = xfilepath.BaseName(filePath)
	}
	return &ancestry.Ancestry, nil
}

// Save writes the Ancestry to the file as JSON.
func (a *Ancestry) Save(filePath string) error {
	return jio.SaveToFile(filePath, &ancestryData{
		Version:  jio.CurrentDataVersion,
		Ancestry: *a,
	})
}

// Clone returns a deep copy of this ancestry, or nil if this ancestry is nil. The runtime-only KeyPrefix fields are
// copied along with everything else. Nil entries within the lists are preserved as nil, so the copy is faithful to the
// original; Normalize is what removes them.
func (a *Ancestry) Clone() *Ancestry {
	if a == nil {
		return nil
	}
	clone := *a
	clone.CommonOptions = a.CommonOptions.Clone()
	if a.GenderOptions != nil {
		clone.GenderOptions = make([]*WeightedAncestryOptions, len(a.GenderOptions))
		for i, one := range a.GenderOptions {
			clone.GenderOptions[i] = one.Clone()
		}
	}
	return &clone
}

// Hash writes this object's contents into the hasher. The hash is of the ancestry's own JSON content, without the
// version wrapper that Save adds around it, so it answers exactly "has the file content changed", and the runtime-only
// KeyPrefix fields never take part.
func (a *Ancestry) Hash(h hash.Hash) {
	HashJSON(h, a)
}

// ResetTargetKeyPrefixes assigns new key prefixes for all data within this Ancestry. Nil entries are skipped.
func (a *Ancestry) ResetTargetKeyPrefixes(prefixProvider func() string) {
	a.KeyPrefix = prefixProvider()
	if a.CommonOptions != nil {
		a.CommonOptions.ResetTargetKeyPrefixes(prefixProvider)
	}
	for _, one := range a.GenderOptions {
		if one != nil {
			one.KeyPrefix = prefixProvider()
			if one.Value != nil {
				one.Value.ResetTargetKeyPrefixes(prefixProvider)
			}
		}
	}
}

// Normalize makes the ancestry safe to edit: it drops the null entries a hand-written file may contain, gives the
// ancestry a common options block if it has none, and gives every gender option a value if it lacks one. The
// randomizers tolerate all of those, but an editor needs every pointer it renders to be non-nil. This is deliberately
// not called when loading a file, so that the load path keeps tolerating such files as-is.
func (a *Ancestry) Normalize() {
	a.GenderOptions = slices.DeleteFunc(a.GenderOptions, func(o *WeightedAncestryOptions) bool { return o == nil })
	for _, one := range a.GenderOptions {
		if one.Value == nil {
			one.Value = &AncestryOptions{}
		}
	}
	if a.CommonOptions == nil {
		a.CommonOptions = &AncestryOptions{}
	}
	a.CommonOptions.Normalize()
	for _, one := range a.GenderOptions {
		one.Value.Normalize()
	}
}

// RandomGender returns a randomized gender.
func (a *Ancestry) RandomGender(not string) string {
	if choice := ChooseWeightedAncestryOptions(a.GenderOptions, func(o *AncestryOptions) bool {
		return o.Name == not
	}); choice != nil {
		return choice.Name
	}
	if choice := ChooseWeightedAncestryOptions(a.GenderOptions, nil); choice != nil {
		return choice.Name
	}
	return not
}

// GenderedOptions returns the options for the specified gender, or nil.
func (a *Ancestry) GenderedOptions(gender string) *AncestryOptions {
	gender = strings.TrimSpace(gender)
	for _, one := range a.GenderOptions {
		if one != nil && one.Value != nil && strings.EqualFold(one.Value.Name, gender) {
			return one.Value
		}
	}
	return nil
}

// RandomHeight returns a randomized height.
func (a *Ancestry) RandomHeight(entity *Entity, gender string, not fxp.Length) fxp.Length {
	if options := a.GenderedOptions(gender); options != nil && options.HeightScript != "" {
		return options.RandomHeight(entity, not)
	}
	if a.CommonOptions != nil && a.CommonOptions.HeightScript != "" {
		return a.CommonOptions.RandomHeight(entity, not)
	}
	return fxp.LengthFromInteger(defaultHeight, fxp.Inch)
}

// RandomWeight returns a randomized weight.
func (a *Ancestry) RandomWeight(entity *Entity, gender string, not fxp.Weight) fxp.Weight {
	if options := a.GenderedOptions(gender); options != nil && options.WeightScript != "" {
		return options.RandomWeight(entity, not)
	}
	if a.CommonOptions != nil && a.CommonOptions.WeightScript != "" {
		return a.CommonOptions.RandomWeight(entity, not)
	}
	return fxp.WeightFromInteger(defaultWeight, fxp.Pound)
}

// RandomAge returns a randomized age.
func (a *Ancestry) RandomAge(entity *Entity, gender string, not int) int {
	if options := a.GenderedOptions(gender); options != nil && options.AgeScript != "" {
		return options.RandomAge(entity, not)
	}
	if a.CommonOptions != nil && a.CommonOptions.AgeScript != "" {
		return a.CommonOptions.RandomAge(entity, not)
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

// ActiveAncestries returns a list of Ancestry nodes that are enabled in the given Trait nodes and their descendants.
func ActiveAncestries(list []*Trait) []*Ancestry {
	var ancestries []*Ancestry
	libraries := GlobalSettings().Libraries
	Traverse(func(t *Trait) bool {
		if t.Container() && t.ContainerType == container.Ancestry && t.Enabled() {
			if anc := LookupAncestry(t.Ancestry, libraries); anc != nil {
				ancestries = append(ancestries, anc)
			}
		}
		return false
	}, true, false, list...)
	return ancestries
}

// ActiveAncestryTraits returns the Traits that have Ancestry data and are enabled within the given traits or their
// descendants.
func ActiveAncestryTraits(list []*Trait) []*Trait {
	var result []*Trait
	libraries := GlobalSettings().Libraries
	Traverse(func(t *Trait) bool {
		if t.Container() && t.ContainerType == container.Ancestry && t.Enabled() {
			if ancestry := LookupAncestry(t.Ancestry, libraries); ancestry != nil {
				result = append(result, t)
			}
		}
		return false
	}, true, false, list...)
	return result
}
