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

var _ situationKeyed = &ReactionBonus{}

// ReactionBonus holds a modifier due to a reaction.
type ReactionBonus struct {
	Type feature.Type `json:"type"`
	FeatureSwitch
	situationBonus
}

// NewReactionBonus creates a new ReactionBonus.
func NewReactionBonus() *ReactionBonus {
	var r ReactionBonus
	r.Type = feature.ReactionBonus
	r.Situation = i18n.Text("from others")
	r.Amount = fxp.One
	return &r
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

// Hash writes this object's contents into the hasher.
func (r *ReactionBonus) Hash(h hash.Hash) {
	if r == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, r.Type)
	xhash.Bool(h, r.Switchable)
	r.hashSituation(h)
}
