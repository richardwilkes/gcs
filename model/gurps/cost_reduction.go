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
	"hash"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Feature = &CostReduction{}

// CostReduction holds the data for a cost reduction.
type CostReduction struct {
	Type feature.Type `json:"type"`
	FeatureSwitch
	Attribute  string  `json:"attribute,omitzero"`
	Percentage fxp.Int `json:"percentage,omitzero"`
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
	return clonePtr(c)
}

// FillWithNameableKeys implements Feature.
func (c *CostReduction) FillWithNameableKeys(_, _ map[string]string) {
}

// Describe implements Feature. The attribute is named as the entity, which may be nil, defines it.
func (c *CostReduction) Describe(entity *Entity, _ map[string]string, em func(string) string) string {
	return describeSwitchable(c.Switchable, fmt.Sprintf(i18n.Text("Reduces the cost of %s by %s"),
		em(attributeTitle(entity, c.Attribute)), em(c.Percentage.String()+"%")))
}

// Hash writes this object's contents into the hasher.
func (c *CostReduction) Hash(h hash.Hash) {
	if c == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, c.Type)
	xhash.Bool(h, c.Switchable)
	xhash.StringWithLen(h, c.Attribute)
	xhash.Num64(h, c.Percentage)
}
