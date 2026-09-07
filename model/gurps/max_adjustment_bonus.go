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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/maxusesmod"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// maxAdjustmentBonusData is the part shared by the bonuses that adjust a maximum: the maximum uses of a piece of
// equipment (EquipmentMaxUsesBonus) and the maximum level of a trait (TraitMaxLevelBonus). S is the selection enum
// that says whether the bonus applies to the item it is attached to or to the items matching the name and tag
// criteria. The embedding types inline it untagged, so the JSON layout is theirs.
type maxAdjustmentBonusData[S ~byte] struct {
	Type feature.Type `json:"type"`
	FeatureSwitch
	SelectionType S             `json:"selection_type"`
	NameCriteria  criteria.Text `json:"name,omitzero"`
	TagsCriteria  criteria.Text `json:"tags,omitzero"`
	MaxUsesModAmount
	BonusOwner `json:"-"`
}

// newMaxAdjustmentBonusData returns the data for a new bonus of the given feature type: one that applies to the item it
// is attached to and adds one to its maximum.
func newMaxAdjustmentBonusData[S ~byte](featureType feature.Type, this S) maxAdjustmentBonusData[S] {
	var d maxAdjustmentBonusData[S]
	d.Type = featureType
	d.SelectionType = this
	d.NameCriteria.Compare = criteria.IsText
	d.TagsCriteria.Compare = criteria.AnyText
	d.Amount = maxusesmod.Normalize("+1")
	return d
}

// FeatureType implements Feature.
func (d *maxAdjustmentBonusData[S]) FeatureType() feature.Type {
	return d.Type
}

// AddToTooltip implements Bonus.
func (d *maxAdjustmentBonusData[S]) AddToTooltip(buffer *xbytes.InsertBuffer) {
	d.addToTooltip(d.parentName(), buffer)
}

// adjustmentData returns the shared data, which is how addMaxAdjustmentsFrom picks out the bonuses for its selection
// enum from a feature list.
func (d *maxAdjustmentBonusData[S]) adjustmentData() *maxAdjustmentBonusData[S] {
	return d
}

// fillWithNameableKeysWhen extracts the nameable keys from the name and tag criteria when the selection is withName,
// the only one that consults them.
func (d *maxAdjustmentBonusData[S]) fillWithNameableKeysWhen(m, existing map[string]string, withName S) {
	if d.SelectionType == withName {
		nameable.Extract(
			m, existing,
			d.NameCriteria.Qualifier,
			d.TagsCriteria.Qualifier,
		)
	}
}

// Hash writes the fields into the hash. The embedding types guard against a nil receiver before calling it.
func (d *maxAdjustmentBonusData[S]) Hash(h hash.Hash) {
	xhash.Num8(h, d.Type)
	xhash.Bool(h, d.Switchable)
	xhash.Num8(h, d.SelectionType)
	d.NameCriteria.Hash(h)
	d.TagsCriteria.Hash(h)
	d.MaxUsesModAmount.Hash(h)
}

// maxAdjustment accumulates the bonuses that adjust a maximum and applies them to the base value. The additions are
// applied first, then the percentages, then the multipliers.
type maxAdjustment struct {
	addition   fxp.Int
	percentage fxp.Int
	multiplier fxp.Int
	have       bool
}

func newMaxAdjustment() maxAdjustment {
	return maxAdjustment{multiplier: fxp.One}
}

func (m *maxAdjustment) add(bonus *MaxUsesModAmount) {
	m.have = true
	amount := bonus.AdjustedAmount()
	switch bonus.Operation() {
	case maxusesmod.Percentage:
		m.percentage += amount
	case maxusesmod.Multiplier:
		if amount <= 0 {
			amount = fxp.One
		}
		m.multiplier = m.multiplier.Mul(amount)
	default: // maxusesmod.Addition
		m.addition += amount
	}
}

// apply returns the base adjusted by the accumulated bonuses, floored at zero. A base of zero means there is no
// maximum, and it is returned as is. Bonuses adjust an existing cap, so one must never be allowed to manufacture a cap
// from a base of zero -- that would turn a bonus meant to raise a limit into one that imposes it.
func (m *maxAdjustment) apply(base fxp.Int) fxp.Int {
	if !m.have || base <= 0 {
		return base.Max(0)
	}
	result := base + m.addition
	result += result.Mul(m.percentage).Div(fxp.Hundred)
	result = result.Mul(m.multiplier)
	return result.Max(0)
}

// addMaxAdjustmentsFrom folds into adj each bonus in the feature list whose selection enum is S and whose selection is
// this, the value meaning the item the bonus is attached to, after making leveledOwner the node whose level drives a
// per-level amount.
func addMaxAdjustmentsFrom[S ~byte](adj *maxAdjustment, features Features, this S, leveledOwner LeveledOwner) {
	for _, f := range features {
		if bonus, ok := f.(interface {
			adjustmentData() *maxAdjustmentBonusData[S]
		}); ok {
			if data := bonus.adjustmentData(); data.SelectionType == this {
				data.SetLeveledOwner(leveledOwner)
				adj.add(&data.MaxUsesModAmount)
			}
		}
	}
}
