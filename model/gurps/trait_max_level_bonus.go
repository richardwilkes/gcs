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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/traitsel"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xreflect"
)

var _ Bonus = &TraitMaxLevelBonus{}

// TraitMaxLevelBonus holds an adjustment to a trait's maximum level. The Amount field encodes both the operation and
// the value: a plain number (e.g. "+1" or "-1") is an addition, a value with a trailing "%" (e.g. "10%") is a
// percentage adjustment, and a value with a leading or trailing "x" (e.g. "x2") is a multiplier.
type TraitMaxLevelBonus struct {
	Type          feature.Type  `json:"type"`
	SelectionType traitsel.Type `json:"selection_type"`
	PerLevel      bool          `json:"per_level,omitzero"`
	NameCriteria  criteria.Text `json:"name,omitzero"`
	TagsCriteria  criteria.Text `json:"tags,omitzero"`
	Amount        string        `json:"amount"`
	LeveledOwner  LeveledOwner  `json:"-"`
	BonusOwner    `json:"-"`
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
		nameable.Extract(t.NameCriteria.Qualifier, m, existing)
		nameable.Extract(t.TagsCriteria.Qualifier, m, existing)
	}
}

// Operation returns the type of adjustment encoded in the Amount.
func (t *TraitMaxLevelBonus) Operation() maxusesmod.Type {
	return maxusesmod.FromString(t.Amount)
}

// SetLeveledOwner implements Bonus.
func (t *TraitMaxLevelBonus) SetLeveledOwner(owner LeveledOwner) {
	t.LeveledOwner = owner
}

// AdjustedAmount implements Bonus. It returns the numeric portion of the Amount, scaled by the leveled owner's current
// level when this is a per-level bonus. A per-level bonus whose owner has no levels contributes nothing.
func (t *TraitMaxLevelBonus) AdjustedAmount() fxp.Int {
	value := t.Operation().ExtractValue(t.Amount)
	if t.PerLevel {
		if xreflect.IsNil(t.LeveledOwner) || !t.LeveledOwner.IsLeveled() {
			return 0
		}
		level := t.LeveledOwner.CurrentLevel()
		if level <= 0 {
			return 0
		}
		value = value.Mul(level)
	}
	return value
}

// AddToTooltip implements Bonus.
func (t *TraitMaxLevelBonus) AddToTooltip(buffer *xbytes.InsertBuffer) {
	if buffer != nil {
		buffer.WriteByte('\n')
		buffer.WriteString(t.parentName())
		buffer.WriteString(" [")
		buffer.WriteString(t.Operation().Format(t.AdjustedAmount()))
		buffer.WriteByte(']')
	}
}

// Hash writes this object's contents into the hasher.
func (t *TraitMaxLevelBonus) Hash(h hash.Hash) {
	if t == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, t.Type)
	xhash.Num8(h, t.SelectionType)
	t.NameCriteria.Hash(h)
	t.TagsCriteria.Hash(h)
	xhash.StringWithLen(h, t.Amount)
	xhash.Bool(h, t.PerLevel)
}
