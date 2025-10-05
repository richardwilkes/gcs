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
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
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

// SetLeveledOwner implements Bonus.
func (r *ReactionBonus) SetLeveledOwner(owner LeveledOwner) {
	r.LeveledOwner = owner
}

// AddToTooltip implements Bonus.
func (r *ReactionBonus) AddToTooltip(buffer *xbytes.InsertBuffer) {
	r.basicAddToTooltip(&r.LeveledAmount, buffer)
}

// Hash writes this object's contents into the hasher.
func (r *ReactionBonus) Hash(h hash.Hash) {
	if r == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, r.Type)
	xhash.StringWithLen(h, r.Situation)
	r.LeveledAmount.Hash(h)
}
