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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/maxusesmod"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/traitsel"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Bonus = &TraitMaxLevelBonus{}

// TraitMaxLevelBonus holds an adjustment to a trait's maximum level. See MaxUsesModAmount for how the adjustment is
// encoded.
type TraitMaxLevelBonus struct {
	Type feature.Type `json:"type"`
	FeatureSwitch
	SelectionType traitsel.Type `json:"selection_type"`
	NameCriteria  criteria.Text `json:"name,omitzero"`
	TagsCriteria  criteria.Text `json:"tags,omitzero"`
	MaxUsesModAmount
	BonusOwner `json:"-"`
}

// NewTraitMaxLevelBonus creates a new TraitMaxLevelBonus.
func NewTraitMaxLevelBonus() *TraitMaxLevelBonus {
	var t TraitMaxLevelBonus
	t.Type = feature.TraitMaxLevelBonus
	t.SelectionType = traitsel.ThisTrait
	t.NameCriteria.Compare = criteria.IsText
	t.TagsCriteria.Compare = criteria.AnyText
	t.Amount = maxusesmod.Normalize("+1")
	return &t
}

// FeatureType implements Feature.
func (t *TraitMaxLevelBonus) FeatureType() feature.Type {
	return t.Type
}

// Clone implements Feature.
func (t *TraitMaxLevelBonus) Clone() Feature {
	other := *t
	return &other
}

// FillWithNameableKeys implements Feature.
func (t *TraitMaxLevelBonus) FillWithNameableKeys(m, existing map[string]string) {
	if t.SelectionType == traitsel.TraitWithName {
		nameable.Extract(
			m, existing,
			t.NameCriteria.Qualifier,
			t.TagsCriteria.Qualifier,
		)
	}
}

// AddToTooltip implements Bonus.
func (t *TraitMaxLevelBonus) AddToTooltip(buffer *xbytes.InsertBuffer) {
	t.addToTooltip(t.parentName(), buffer)
}

// Hash writes this object's contents into the hasher.
func (t *TraitMaxLevelBonus) Hash(h hash.Hash) {
	if t == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, t.Type)
	xhash.Bool(h, t.Switchable)
	xhash.Num8(h, t.SelectionType)
	t.NameCriteria.Hash(h)
	t.TagsCriteria.Hash(h)
	t.MaxUsesModAmount.Hash(h)
}
