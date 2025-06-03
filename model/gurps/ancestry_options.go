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
	"bytes"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
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
}

// AncestryOptionsData holds the data that will be serialized for the AncestryOptions.
type AncestryOptionsData struct {
	Name              string                  `json:"name,omitempty"`
	HeightScript      string                  `json:"height_script,omitempty"`
	WeightScript      string                  `json:"weight_script,omitempty"`
	AgeScript         string                  `json:"age_script,omitempty"`
	HairOptions       []*WeightedStringOption `json:"hair_options,omitempty"`
	EyeOptions        []*WeightedStringOption `json:"eye_options,omitempty"`
	SkinOptions       []*WeightedStringOption `json:"skin_options,omitempty"`
	HandednessOptions []*WeightedStringOption `json:"handedness_options,omitempty"`
	NameGenerators    []string                `json:"name_generators,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (o *AncestryOptions) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	e := json.NewEncoder(&buffer)
	e.SetEscapeHTML(false)
	err := e.Encode(&o.AncestryOptionsData)
	return buffer.Bytes(), err
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *AncestryOptions) UnmarshalJSON(data []byte) error {
	var legacy struct {
		AncestryOptionsData
		// Old data fields
		HeightFormula string `json:"height_formula"`
		WeightFormula string `json:"weight_formula"`
		AgeFormula    string `json:"age_formula"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
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

// RandomHeight returns a randomized height.
func (o *AncestryOptions) RandomHeight(entity *Entity, not fxp.Length) fxp.Length {
	def := fxp.LengthFromInteger(defaultHeight, fxp.Inch)
	for range maximumRandomTries {
		value := fxp.Length(ResolveToNumber(entity, ScriptSelfProvider{}, o.HeightScript))
		if value <= 0 {
			value = def
		}
		if value != not {
			return value
		}
	}
	return def
}

// RandomWeight returns a randomized weight.
func (o *AncestryOptions) RandomWeight(entity *Entity, not fxp.Weight) fxp.Weight {
	def := fxp.WeightFromInteger(defaultWeight, fxp.Pound)
	for range maximumRandomTries {
		value := fxp.Weight(ResolveToNumber(entity, ScriptSelfProvider{}, o.WeightScript))
		if value <= 0 {
			value = def
		}
		if value != not {
			return value
		}
	}
	return def
}

// RandomAge returns a randomized age.
func (o *AncestryOptions) RandomAge(entity *Entity, not int) int {
	for range maximumRandomTries {
		age := fxp.As[int](ResolveToNumber(entity, ScriptSelfProvider{}, o.AgeScript))
		if age <= 0 {
			age = defaultAge
		}
		if age != not {
			return age
		}
	}
	return defaultAge
}

// RandomHair returns a randomized hair.
func (o *AncestryOptions) RandomHair(not string) string {
	if choice := ChooseWeightedStringOption(o.HairOptions, not); choice != "" {
		return choice
	}
	return defaultHair
}

// RandomEye returns a randomized eye.
func (o *AncestryOptions) RandomEye(not string) string {
	if choice := ChooseWeightedStringOption(o.EyeOptions, not); choice != "" {
		return choice
	}
	return defaultEye
}

// RandomSkin returns a randomized skin.
func (o *AncestryOptions) RandomSkin(not string) string {
	if choice := ChooseWeightedStringOption(o.SkinOptions, not); choice != "" {
		return choice
	}
	return defaultSkin
}

// RandomHandedness returns a randomized handedness.
func (o *AncestryOptions) RandomHandedness(not string) string {
	if choice := ChooseWeightedStringOption(o.HandednessOptions, not); choice != "" {
		return choice
	}
	return defaultHandedness
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
