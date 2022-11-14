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

package gurps

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/nameables"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

var _ Bonus = &ConditionalModifierBonus{}

// ConditionalModifierBonus holds the data for a conditional modifier bonus.
type ConditionalModifierBonus struct {
	Type      FeatureType `json:"type"`
	Situation string      `json:"situation,omitempty"`
	LeveledAmount
	owner fmt.Stringer
}

// NewConditionalModifierBonus creates a new ConditionalModifierBonus.
func NewConditionalModifierBonus() *ConditionalModifierBonus {
	return &ConditionalModifierBonus{
		Type:          ConditionalModifierFeatureType,
		Situation:     i18n.Text("triggering condition"),
		LeveledAmount: LeveledAmount{Amount: fxp.One},
	}
}

// FeatureType implements Feature.
func (c *ConditionalModifierBonus) FeatureType() FeatureType {
	return c.Type
}

// Clone implements Feature.
func (c *ConditionalModifierBonus) Clone() Feature {
	other := *c
	return &other
}

// FillWithNameableKeys implements Feature.
func (c *ConditionalModifierBonus) FillWithNameableKeys(m map[string]string) {
	nameables.Extract(c.Situation, m)
}

// ApplyNameableKeys implements Feature.
func (c *ConditionalModifierBonus) ApplyNameableKeys(m map[string]string) {
	c.Situation = nameables.Apply(c.Situation, m)
}

// Owner implements Bonus.
func (c *ConditionalModifierBonus) Owner() fmt.Stringer {
	return c.owner
}

// SetOwner implements Bonus.
func (c *ConditionalModifierBonus) SetOwner(owner fmt.Stringer) {
	c.owner = owner
}

// SetLevel implements Bonus.
func (c *ConditionalModifierBonus) SetLevel(level fxp.Int) {
	c.Level = level
}

// AddToTooltip implements Bonus.
func (c *ConditionalModifierBonus) AddToTooltip(buffer *xio.ByteBuffer) {
	basicAddToTooltip(c.owner, &c.LeveledAmount, buffer)
}
