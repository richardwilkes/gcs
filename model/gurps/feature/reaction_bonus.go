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

var _ Bonus = &ReactionBonus{}

// ReactionBonus holds a modifier due to a reaction.
type ReactionBonus struct {
	Type      Type         `json:"type"`
	Parent    fmt.Stringer `json:"-"`
	Situation string       `json:"situation,omitempty"`
	LeveledAmount
}

// NewReactionBonus creates a new ReactionBonus.
func NewReactionBonus() *ReactionBonus {
	return &ReactionBonus{
		Type:          ReactionBonusType,
		Situation:     i18n.Text("from others"),
		LeveledAmount: LeveledAmount{Amount: fxp.One},
	}
}

// FeatureType implements Feature.
func (r *ReactionBonus) FeatureType() Type {
	return r.Type
}

// Clone implements Feature.
func (r *ReactionBonus) Clone() Feature {
	other := *r
	return &other
}

// FeatureMapKey implements Feature.
func (r *ReactionBonus) FeatureMapKey() string {
	return "reaction"
}

// FillWithNameableKeys implements Feature.
func (r *ReactionBonus) FillWithNameableKeys(m map[string]string) {
	nameables.Extract(r.Situation, m)
}

// ApplyNameableKeys implements Feature.
func (r *ReactionBonus) ApplyNameableKeys(m map[string]string) {
	r.Situation = nameables.Apply(r.Situation, m)
}

// SetParent implements Bonus.
func (r *ReactionBonus) SetParent(parent fmt.Stringer) {
	r.Parent = parent
}

// SetLevel implements Bonus.
func (r *ReactionBonus) SetLevel(level fxp.Int) {
	r.Level = level
}

// AddToTooltip implements Bonus.
func (r *ReactionBonus) AddToTooltip(buffer *xio.ByteBuffer) {
	basicAddToTooltip(r.Parent, &r.LeveledAmount, buffer)
}
