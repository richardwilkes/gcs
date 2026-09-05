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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/equipmentsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/maxusesmod"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Bonus = &EquipmentMaxUsesBonus{}

// EquipmentMaxUsesBonus holds an adjustment to a piece of equipment's maximum uses. See MaxUsesModAmount for how the
// adjustment is encoded.
type EquipmentMaxUsesBonus struct {
	Type feature.Type `json:"type"`
	FeatureSwitch
	SelectionType equipmentsel.Type `json:"selection_type"`
	NameCriteria  criteria.Text     `json:"name,omitzero"`
	TagsCriteria  criteria.Text     `json:"tags,omitzero"`
	MaxUsesModAmount
	BonusOwner `json:"-"`
}

// NewEquipmentMaxUsesBonus creates a new EquipmentMaxUsesBonus.
func NewEquipmentMaxUsesBonus() *EquipmentMaxUsesBonus {
	var e EquipmentMaxUsesBonus
	e.Type = feature.EquipmentMaxUsesBonus
	e.SelectionType = equipmentsel.ThisEquipment
	e.NameCriteria.Compare = criteria.IsText
	e.TagsCriteria.Compare = criteria.AnyText
	e.Amount = maxusesmod.Normalize("+1")
	return &e
}

// FeatureType implements Feature.
func (e *EquipmentMaxUsesBonus) FeatureType() feature.Type {
	return e.Type
}

// Clone implements Feature.
func (e *EquipmentMaxUsesBonus) Clone() Feature {
	other := *e
	return &other
}

// FillWithNameableKeys implements Feature.
func (e *EquipmentMaxUsesBonus) FillWithNameableKeys(m, existing map[string]string) {
	if e.SelectionType == equipmentsel.EquipmentWithName {
		nameable.Extract(
			m, existing,
			e.NameCriteria.Qualifier,
			e.TagsCriteria.Qualifier,
		)
	}
}

// AddToTooltip implements Bonus.
func (e *EquipmentMaxUsesBonus) AddToTooltip(buffer *xbytes.InsertBuffer) {
	e.addToTooltip(e.parentName(), buffer)
}

// Hash writes this object's contents into the hasher.
func (e *EquipmentMaxUsesBonus) Hash(h hash.Hash) {
	if e == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, e.Type)
	xhash.Bool(h, e.Switchable)
	xhash.Num8(h, e.SelectionType)
	e.NameCriteria.Hash(h)
	e.TagsCriteria.Hash(h)
	e.MaxUsesModAmount.Hash(h)
}
