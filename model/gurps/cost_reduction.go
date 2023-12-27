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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
)

var _ Feature = &CostReduction{}

// CostReduction holds the data for a cost reduction.
type CostReduction struct {
	Type       feature.Type `json:"type"`
	Attribute  string       `json:"attribute,omitempty"`
	Percentage fxp.Int      `json:"percentage,omitempty"`
}

// NewCostReduction creates a new CostReduction.
func NewCostReduction(attrID string) *CostReduction {
	return &CostReduction{
		Type:       feature.CostReduction,
		Attribute:  attrID,
		Percentage: fxp.Forty,
	}
}

// FeatureType implements Feature.
func (c *CostReduction) FeatureType() feature.Type {
	return c.Type
}

// Clone implements Feature.
func (c *CostReduction) Clone() Feature {
	other := *c
	return &other
}

// FillWithNameableKeys implements Feature.
func (c *CostReduction) FillWithNameableKeys(_ map[string]string) {
}

// ApplyNameableKeys implements Feature.
func (c *CostReduction) ApplyNameableKeys(_ map[string]string) {
}
