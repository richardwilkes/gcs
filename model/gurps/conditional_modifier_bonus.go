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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ situationKeyed = &ConditionalModifierBonus{}

// ConditionalModifierBonus holds the data for a conditional modifier bonus.
type ConditionalModifierBonus struct {
	Type feature.Type `json:"type"`
	FeatureSwitch
	situationBonus
}

// NewConditionalModifierBonus creates a new ConditionalModifierBonus.
func NewConditionalModifierBonus() *ConditionalModifierBonus {
	var c ConditionalModifierBonus
	c.Type = feature.ConditionalModifier
	c.Situation = i18n.Text("triggering condition")
	c.Amount = fxp.One
	return &c
}

// FeatureType implements Feature.
func (c *ConditionalModifierBonus) FeatureType() feature.Type {
	return c.Type
}

// Clone implements Feature.
func (c *ConditionalModifierBonus) Clone() Feature {
	other := *c
	return &other
}

// Hash writes this object's contents into the hasher.
func (c *ConditionalModifierBonus) Hash(h hash.Hash) {
	if c == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, c.Type)
	xhash.Bool(h, c.Switchable)
	c.hashSituation(h)
}
