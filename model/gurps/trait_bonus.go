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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"hash"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Bonus = &TraitBonus{}

// TraitBonus holds an adjustment to a trait.
type TraitBonus struct {
	TraitBonusData
}

// TraitBonusData holds an adjustment to a trait which is persisted.
type TraitBonusData struct {
	Type         feature.Type  `json:"type"`
	NameCriteria criteria.Text `json:"name,omitzero"`
	TagsCriteria criteria.Text `json:"tags,omitzero"`
	LeveledAmount
	BonusOwner `json:"-"`
}

// NewTraitBonus creates a new TraitBonus.
func NewTraitBonus() *TraitBonus {
	var s TraitBonus
	s.Type = feature.TraitBonus
	s.NameCriteria.Compare = criteria.IsText
	s.TagsCriteria.Compare = criteria.AnyText
	s.Amount = fxp.One
	return &s
}

// FeatureType implements Feature.
func (s *TraitBonus) FeatureType() feature.Type {
	return s.Type
}

// Clone implements Feature.
func (s *TraitBonus) Clone() Feature {
	other := *s
	return &other
}

// FillWithNameableKeys implements Feature.
func (s *TraitBonus) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(s.NameCriteria.Qualifier, m, existing)
	nameable.Extract(s.TagsCriteria.Qualifier, m, existing)
}

// SetLeveledOwner implements Bonus.
func (s *TraitBonus) SetLeveledOwner(owner LeveledOwner) {
	s.LeveledOwner = owner
}

// AddToTooltip implements Bonus.
func (s *TraitBonus) AddToTooltip(buffer *xbytes.InsertBuffer) {
	s.basicAddToTooltip(&s.LeveledAmount, buffer)
}

// Hash writes this object's contents into the hasher.
func (s *TraitBonus) Hash(h hash.Hash) {
	if s == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, s.Type)
	s.NameCriteria.Hash(h)
	s.TagsCriteria.Hash(h)
	s.LeveledAmount.Hash(h)
}

// MarshalJSONTo implements json.MarshalerTo.
func (s *TraitBonus) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, &s.TraitBonusData)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (s *TraitBonus) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var content TraitBonusData
	if err := json.UnmarshalDecode(dec, &content); err != nil {
		return err
	}
	s.TraitBonusData = content
	return nil
}
