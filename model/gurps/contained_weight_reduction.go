/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

var _ Feature = &ContainedWeightReduction{}

// ContainedWeightReduction holds the data for a weight reduction that can be applied to a container's contents.
type ContainedWeightReduction struct {
	Type      FeatureType `json:"type"`
	Reduction string      `json:"reduction"`
}

// NewContainedWeightReduction creates a new ContainedWeightReduction.
func NewContainedWeightReduction() *ContainedWeightReduction {
	return &ContainedWeightReduction{
		Type:      ContainedWeightReductionFeatureType,
		Reduction: "0%",
	}
}

// FeatureType implements Feature.
func (c *ContainedWeightReduction) FeatureType() FeatureType {
	return c.Type
}

// Clone implements Feature.
func (c *ContainedWeightReduction) Clone() Feature {
	other := *c
	return &other
}

// FillWithNameableKeys implements Feature.
func (c *ContainedWeightReduction) FillWithNameableKeys(_ map[string]string) {
}

// ApplyNameableKeys implements Feature.
func (c *ContainedWeightReduction) ApplyNameableKeys(_ map[string]string) {
}

// IsPercentageReduction returns true if this is a percentage reduction and not a fixed amount.
func (c *ContainedWeightReduction) IsPercentageReduction() bool {
	return strings.HasSuffix(c.Reduction, "%")
}

// PercentageReduction returns the percentage (where 1% is 1, not 0.01) the weight should be reduced by. Will return 0
// if this is not a percentage.
func (c *ContainedWeightReduction) PercentageReduction() fxp.Int {
	if !c.IsPercentageReduction() {
		return 0
	}
	return fxp.FromStringForced(c.Reduction[:len(c.Reduction)-1])
}

// FixedReduction returns the fixed amount the weight should be reduced by. Will return 0 if this is a percentage.
func (c *ContainedWeightReduction) FixedReduction(defUnits fxp.WeightUnits) fxp.Weight {
	if c.IsPercentageReduction() {
		return 0
	}
	return fxp.WeightFromStringForced(c.Reduction, defUnits)
}

// ExtractContainedWeightReduction extracts the weight reduction (which may be a weight or a percentage) and returns
// a sanitized result. If 'err' is not nil, then the input was bad. Even in that case, however, a valid string is
// returned.
func ExtractContainedWeightReduction(s string, defUnits fxp.WeightUnits) (string, error) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "%") {
		v, err := fxp.FromString(strings.TrimSpace(s[:len(s)-1]))
		return v.String() + "%", err
	}
	w, err := fxp.WeightFromString(s, defUnits)
	return defUnits.Format(w), err
}
