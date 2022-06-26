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

package feature

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/nameables"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

var _ Bonus = &ConditionalModifier{}

// ConditionalModifier holds the data for a conditional modifier.
type ConditionalModifier struct {
	Type      Type         `json:"type"`
	Parent    fmt.Stringer `json:"-"`
	Situation string       `json:"situation,omitempty"`
	LeveledAmount
}

// NewConditionalModifierBonus creates a new ConditionalModifier.
func NewConditionalModifierBonus() *ConditionalModifier {
	return &ConditionalModifier{
		Type:          ConditionalModifierType,
		Situation:     i18n.Text("triggering condition"),
		LeveledAmount: LeveledAmount{Amount: fxp.One},
	}
}

// FeatureType implements Feature.
func (c *ConditionalModifier) FeatureType() Type {
	return c.Type
}

// Clone implements Feature.
func (c *ConditionalModifier) Clone() Feature {
	other := *c
	return &other
}

// FeatureMapKey implements Feature.
func (c *ConditionalModifier) FeatureMapKey() string {
	return ConditionalModifierType.Key()
}

// FillWithNameableKeys implements Feature.
func (c *ConditionalModifier) FillWithNameableKeys(m map[string]string) {
	nameables.Extract(c.Situation, m)
}

// ApplyNameableKeys implements Feature.
func (c *ConditionalModifier) ApplyNameableKeys(m map[string]string) {
	c.Situation = nameables.Apply(c.Situation, m)
}

// SetParent implements Bonus.
func (c *ConditionalModifier) SetParent(parent fmt.Stringer) {
	c.Parent = parent
}

// SetLevel implements Bonus.
func (c *ConditionalModifier) SetLevel(level fxp.Int) {
	c.Level = level
}

// AddToTooltip implements Bonus.
func (c *ConditionalModifier) AddToTooltip(buffer *xio.ByteBuffer) {
	basicAddToTooltip(c.Parent, &c.LeveledAmount, buffer)
}
