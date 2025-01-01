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
	"hash"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var _ Bonus = &ConditionalModifierBonus{}

// ConditionalModifierBonus holds the data for a conditional modifier bonus.
type ConditionalModifierBonus struct {
	Type      feature.Type `json:"type"`
	Situation string       `json:"situation,omitempty"`
	LeveledAmount
	BonusOwner
}

// NewConditionalModifierBonus creates a new ConditionalModifierBonus.
func NewConditionalModifierBonus() *ConditionalModifierBonus {
	return &ConditionalModifierBonus{
		Type:          feature.ConditionalModifier,
		Situation:     i18n.Text("triggering condition"),
		LeveledAmount: LeveledAmount{Amount: fxp.One},
	}
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

// FillWithNameableKeys implements Feature.
func (c *ConditionalModifierBonus) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(c.Situation, m, existing)
}

// SetLevel implements Bonus.
func (c *ConditionalModifierBonus) SetLevel(level fxp.Int) {
	c.Level = level
}

// AddToTooltip implements Bonus.
func (c *ConditionalModifierBonus) AddToTooltip(buffer *xio.ByteBuffer) {
	c.basicAddToTooltip(&c.LeveledAmount, buffer)
}

// Hash writes this object's contents into the hasher.
func (c *ConditionalModifierBonus) Hash(h hash.Hash) {
	if c == nil {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.Num8(h, c.Type)
	hashhelper.String(h, c.Situation)
	c.LeveledAmount.Hash(h)
}
