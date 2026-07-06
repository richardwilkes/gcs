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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/equipmentsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/maxusesmod"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xreflect"
)

var _ Bonus = &EquipmentMaxUsesBonus{}

// EquipmentMaxUsesBonus holds an adjustment to a piece of equipment's maximum uses. The Amount field encodes both the
// operation and the value: a plain number (e.g. "+1" or "-1") is an addition, a value with a trailing "%" (e.g. "10%")
// is a percentage adjustment, and a value with a leading or trailing "x" (e.g. "x2") is a multiplier.
type EquipmentMaxUsesBonus struct {
	Type          feature.Type      `json:"type"`
	SelectionType equipmentsel.Type `json:"selection_type"`
	PerLevel      bool              `json:"per_level,omitzero"`
	NameCriteria  criteria.Text     `json:"name,omitzero"`
	TagsCriteria  criteria.Text     `json:"tags,omitzero"`
	Amount        string            `json:"amount"`
	LeveledOwner  LeveledOwner      `json:"-"`
	BonusOwner    `json:"-"`
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
		nameable.Extract(e.NameCriteria.Qualifier, m, existing)
		nameable.Extract(e.TagsCriteria.Qualifier, m, existing)
	}
}

// Operation returns the type of adjustment encoded in the Amount.
func (e *EquipmentMaxUsesBonus) Operation() maxusesmod.Type {
	return maxusesmod.FromString(e.Amount)
}

// SetLeveledOwner implements Bonus.
func (e *EquipmentMaxUsesBonus) SetLeveledOwner(owner LeveledOwner) {
	e.LeveledOwner = owner
}

// AdjustedAmount implements Bonus. It returns the numeric portion of the Amount, scaled by the leveled owner's current
// level when this is a per-level bonus. A per-level bonus whose owner has no levels contributes nothing.
func (e *EquipmentMaxUsesBonus) AdjustedAmount() fxp.Int {
	value := e.Operation().ExtractValue(e.Amount)
	if e.PerLevel {
		if xreflect.IsNil(e.LeveledOwner) || !e.LeveledOwner.IsLeveled() {
			return 0
		}
		level := e.LeveledOwner.CurrentLevel()
		if level <= 0 {
			return 0
		}
		value = value.Mul(level)
	}
	return value
}

// AddToTooltip implements Bonus.
func (e *EquipmentMaxUsesBonus) AddToTooltip(buffer *xbytes.InsertBuffer) {
	if buffer != nil {
		buffer.WriteByte('\n')
		buffer.WriteString(e.parentName())
		buffer.WriteString(" [")
		buffer.WriteString(e.Operation().Format(e.AdjustedAmount()))
		buffer.WriteByte(']')
	}
}

// Hash writes this object's contents into the hasher.
func (e *EquipmentMaxUsesBonus) Hash(h hash.Hash) {
	if e == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, e.Type)
	xhash.Num8(h, e.SelectionType)
	e.NameCriteria.Hash(h)
	e.TagsCriteria.Hash(h)
	xhash.StringWithLen(h, e.Amount)
	xhash.Bool(h, e.PerLevel)
}
