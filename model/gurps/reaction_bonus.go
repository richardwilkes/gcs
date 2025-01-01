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

var _ Bonus = &ReactionBonus{}

// ReactionBonus holds a modifier due to a reaction.
type ReactionBonus struct {
	Type      feature.Type `json:"type"`
	Situation string       `json:"situation,omitempty"`
	LeveledAmount
	BonusOwner
}

// NewReactionBonus creates a new ReactionBonus.
func NewReactionBonus() *ReactionBonus {
	return &ReactionBonus{
		Type:          feature.ReactionBonus,
		Situation:     i18n.Text("from others"),
		LeveledAmount: LeveledAmount{Amount: fxp.One},
	}
}

// FeatureType implements Feature.
func (r *ReactionBonus) FeatureType() feature.Type {
	return r.Type
}

// Clone implements Feature.
func (r *ReactionBonus) Clone() Feature {
	other := *r
	return &other
}

// FillWithNameableKeys implements Feature.
func (r *ReactionBonus) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(r.Situation, m, existing)
}

// SetLevel implements Bonus.
func (r *ReactionBonus) SetLevel(level fxp.Int) {
	r.Level = level
}

// AddToTooltip implements Bonus.
func (r *ReactionBonus) AddToTooltip(buffer *xio.ByteBuffer) {
	r.basicAddToTooltip(&r.LeveledAmount, buffer)
}

// Hash writes this object's contents into the hasher.
func (r *ReactionBonus) Hash(h hash.Hash) {
	if r == nil {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.Num8(h, r.Type)
	hashhelper.String(h, r.Situation)
	r.LeveledAmount.Hash(h)
}
