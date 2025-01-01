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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var _ Bonus = &AttributeBonus{}

// AttributeBonus holds the data for a bonus to an attribute.
type AttributeBonus struct {
	Type       feature.Type   `json:"type"`
	Limitation stlimit.Option `json:"limitation,omitempty"`
	Attribute  string         `json:"attribute"`
	LeveledAmount
	BonusOwner
}

// NewAttributeBonus creates a new AttributeBonus.
func NewAttributeBonus(attrID string) *AttributeBonus {
	return &AttributeBonus{
		Type:          feature.AttributeBonus,
		Attribute:     attrID,
		Limitation:    stlimit.None,
		LeveledAmount: LeveledAmount{Amount: fxp.One},
	}
}

// ActualLimitation returns the actual limitation, if any. This is needed in case the limitation is set to something
// other than none when the attribute is not ST.
func (a *AttributeBonus) ActualLimitation() stlimit.Option {
	if a.Attribute == StrengthID {
		return a.Limitation
	}
	return stlimit.None
}

// FeatureType implements Feature.
func (a *AttributeBonus) FeatureType() feature.Type {
	return a.Type
}

// Clone implements Feature.
func (a *AttributeBonus) Clone() Feature {
	other := *a
	return &other
}

// FillWithNameableKeys implements Feature.
func (a *AttributeBonus) FillWithNameableKeys(_, _ map[string]string) {
}

// SetLevel implements Bonus.
func (a *AttributeBonus) SetLevel(level fxp.Int) {
	a.Level = level
}

// AddToTooltip implements Bonus.
func (a *AttributeBonus) AddToTooltip(buffer *xio.ByteBuffer) {
	a.basicAddToTooltip(&a.LeveledAmount, buffer)
}

// Hash writes this object's contents into the hasher.
func (a *AttributeBonus) Hash(h hash.Hash) {
	if a == nil {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.Num8(h, a.Type)
	hashhelper.Num8(h, a.Limitation)
	hashhelper.String(h, a.Attribute)
	a.LeveledAmount.Hash(h)
}
