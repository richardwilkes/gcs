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
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/errs"
)

const (
	defaultHeight      = 64
	defaultWeight      = 140
	defaultAge         = 18
	defaultHair        = "Brown"
	defaultEye         = "Brown"
	defaultSkin        = "Brown"
	defaultHandedness  = "Right"
	maximumRandomTries = 5
)

// AncestryOptions holds options that may be randomized for an Entity's ancestry.
type AncestryOptions struct {
	AncestryOptionsData
	// KeyPrefix is a runtime-only key the editor uses to give this block's widgets a stable identity. It is never
	// written to disk: MarshalJSONTo serializes only the AncestryOptionsData.
	KeyPrefix string `json:"-"`
}

// AncestryOptionsData holds the data that will be serialized for the AncestryOptions.
type AncestryOptionsData struct {
	Name              string                  `json:"name,omitzero"`
	HeightScript      string                  `json:"height_script,omitzero"`
	WeightScript      string                  `json:"weight_script,omitzero"`
	AgeScript         string                  `json:"age_script,omitzero"`
	HairOptions       []*WeightedStringOption `json:"hair_options,omitempty"`
	EyeOptions        []*WeightedStringOption `json:"eye_options,omitempty"`
	SkinOptions       []*WeightedStringOption `json:"skin_options,omitempty"`
	HandednessOptions []*WeightedStringOption `json:"handedness_options,omitempty"`
	NameGenerators    []string                `json:"name_generators,omitempty"`
}

// MarshalJSONTo implements json.MarshalerTo.
func (o *AncestryOptions) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, &o.AncestryOptionsData)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (o *AncestryOptions) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var legacy struct {
		AncestryOptionsData
		// Old data fields
		HeightFormula string `json:"height_formula"`
		WeightFormula string `json:"weight_formula"`
		AgeFormula    string `json:"age_formula"`
	}
	if err := json.UnmarshalDecode(dec, &legacy); err != nil {
		return err
	}
	o.AncestryOptionsData = legacy.AncestryOptionsData
	if o.HeightScript == "" && legacy.HeightFormula != "" {
		o.HeightScript = ExprToScript(legacy.HeightFormula)
	}
	if o.WeightScript == "" && legacy.WeightFormula != "" {
		o.WeightScript = ExprToScript(legacy.WeightFormula)
	}
	if o.AgeScript == "" && legacy.AgeFormula != "" {
		o.AgeScript = ExprToScript(legacy.AgeFormula)
	}
	return nil
}

// Clone returns a deep copy of these options, or nil if these options are nil. Nil entries within the option lists are
// preserved as nil, so the copy is faithful to the original; Normalize is what removes them.
func (o *AncestryOptions) Clone() *AncestryOptions {
	if o == nil {
		return nil
	}
	clone := *o
	clone.HairOptions = cloneWeightedStringOptions(o.HairOptions)
	clone.EyeOptions = cloneWeightedStringOptions(o.EyeOptions)
	clone.SkinOptions = cloneWeightedStringOptions(o.SkinOptions)
	clone.HandednessOptions = cloneWeightedStringOptions(o.HandednessOptions)
	clone.NameGenerators = slices.Clone(o.NameGenerators)
	return &clone
}

// cloneWeightedStringOptions returns a deep copy of the list. A nil list stays nil and nil entries stay nil.
func cloneWeightedStringOptions(list []*WeightedStringOption) []*WeightedStringOption {
	if list == nil {
		return nil
	}
	clone := make([]*WeightedStringOption, len(list))
	for i, one := range list {
		clone[i] = one.Clone()
	}
	return clone
}

// ResetTargetKeyPrefixes assigns new key prefixes for all data within these options.
func (o *AncestryOptions) ResetTargetKeyPrefixes(prefixProvider func() string) {
	o.KeyPrefix = prefixProvider()
	for _, list := range [][]*WeightedStringOption{o.HairOptions, o.EyeOptions, o.SkinOptions, o.HandednessOptions} {
		for _, one := range list {
			if one != nil {
				one.KeyPrefix = prefixProvider()
			}
		}
	}
}

// Normalize drops the null entries a hand-written file may contain from each of the weighted option lists. The
// randomizers tolerate them, but an editor needs every pointer it renders to be non-nil.
func (o *AncestryOptions) Normalize() {
	o.HairOptions = dropNilWeightedStringOptions(o.HairOptions)
	o.EyeOptions = dropNilWeightedStringOptions(o.EyeOptions)
	o.SkinOptions = dropNilWeightedStringOptions(o.SkinOptions)
	o.HandednessOptions = dropNilWeightedStringOptions(o.HandednessOptions)
}

func dropNilWeightedStringOptions(list []*WeightedStringOption) []*WeightedStringOption {
	return slices.DeleteFunc(list, func(o *WeightedStringOption) bool { return o == nil })
}

// RandomHeight returns a randomized height.
func (o *AncestryOptions) RandomHeight(entity *Entity, not fxp.Length) fxp.Length {
	def := fxp.LengthFromInteger(defaultHeight, fxp.Inch)
	i := 0
	for {
		value := fxp.Length(ResolveToNumber(entity, ScriptSelfProvider{}, o.HeightScript))
		if value <= 0 {
			value = def
		}
		i++
		if value != not || i >= maximumRandomTries {
			return value
		}
	}
}

// RandomWeight returns a randomized weight.
func (o *AncestryOptions) RandomWeight(entity *Entity, not fxp.Weight) fxp.Weight {
	def := fxp.WeightFromInteger(defaultWeight, fxp.Pound)
	i := 0
	for {
		value := fxp.Weight(ResolveToNumber(entity, ScriptSelfProvider{}, o.WeightScript))
		if value <= 0 {
			value = def
		}
		i++
		if value != not || i >= maximumRandomTries {
			return value
		}
	}
}

// RandomAge returns a randomized age.
func (o *AncestryOptions) RandomAge(entity *Entity, not int) int {
	i := 0
	for {
		age := ResolveToNumber(entity, ScriptSelfProvider{}, o.AgeScript).AsInteger[int]()
		if age <= 0 {
			age = defaultAge
		}
		i++
		if age != not || i >= maximumRandomTries {
			return age
		}
	}
}

// RandomHair returns a randomized hair.
func (o *AncestryOptions) RandomHair(not string) string {
	return chooseWeightedStringOption(o.HairOptions, not, defaultHair)
}

// RandomEye returns a randomized eye.
func (o *AncestryOptions) RandomEye(not string) string {
	return chooseWeightedStringOption(o.EyeOptions, not, defaultEye)
}

// RandomSkin returns a randomized skin.
func (o *AncestryOptions) RandomSkin(not string) string {
	return chooseWeightedStringOption(o.SkinOptions, not, defaultSkin)
}

// RandomHandedness returns a randomized handedness.
func (o *AncestryOptions) RandomHandedness(not string) string {
	return chooseWeightedStringOption(o.HandednessOptions, not, defaultHandedness)
}

// chooseWeightedStringOption randomly selects one of the options, preferring one other than 'not'. If excluding 'not'
// leaves nothing to choose from -- as happens when the options consist solely of the value being replaced -- a second,
// unfiltered pass is made, so that the ancestry's own choice is kept rather than being replaced by def, a value the
// ancestry may not define at all. def is only used when there are no usable options.
func chooseWeightedStringOption(options []*WeightedStringOption, not, def string) string {
	if choice := ChooseWeightedStringOption(options, not); choice != "" {
		return choice
	}
	if not != "" {
		if choice := ChooseWeightedStringOption(options, ""); choice != "" {
			return choice
		}
	}
	return def
}

// RandomName returns a randomized name.
func (o *AncestryOptions) RandomName(nameGeneratorRefs []*NameGeneratorRef) string {
	m := make(map[string]*NameGeneratorRef)
	for _, one := range nameGeneratorRefs {
		m[one.FileRef.Name] = one
	}
	var buffer strings.Builder
	for _, one := range o.NameGenerators {
		if ref, ok := m[one]; ok {
			if generator, err := ref.Generator(); err != nil {
				errs.Log(err)
			} else {
				if name := strings.TrimSpace(generator.GenerateName()); name != "" {
					if buffer.Len() != 0 {
						buffer.WriteByte(' ')
					}
					buffer.WriteString(name)
				}
			}
		}
	}
	return buffer.String()
}
