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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellmatch"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Bonus = &SpellPointBonus{}

// SpellPointBonus holds an adjustment to a spell's points.
type SpellPointBonus struct {
	SpellPointBonusData
}

// SpellPointBonusData holds an adjustment to a spell's points which is persisted.
type SpellPointBonusData struct {
	Type           feature.Type    `json:"type"`
	SpellMatchType spellmatch.Type `json:"match"`
	NameCriteria   criteria.Text   `json:"name,omitzero"`
	TagsCriteria   criteria.Text   `json:"tags,omitzero"`
	LeveledAmount
	BonusOwner `json:"-"`
}

// NewSpellPointBonus creates a new SpellPointBonus.
func NewSpellPointBonus() *SpellPointBonus {
	var s SpellPointBonus
	s.Type = feature.SpellPointBonus
	s.SpellMatchType = spellmatch.AllColleges
	s.NameCriteria.Compare = criteria.IsText
	s.TagsCriteria.Compare = criteria.AnyText
	s.Amount = fxp.One
	return &s
}

// FeatureType implements Feature.
func (s *SpellPointBonus) FeatureType() feature.Type {
	return s.Type
}

// Clone implements Feature.
func (s *SpellPointBonus) Clone() Feature {
	other := *s
	return &other
}

// FillWithNameableKeys implements Feature.
func (s *SpellPointBonus) FillWithNameableKeys(m, existing map[string]string) {
	if s.SpellMatchType != spellmatch.AllColleges {
		nameable.Extract(s.NameCriteria.Qualifier, m, existing)
	}
	nameable.Extract(s.TagsCriteria.Qualifier, m, existing)
}

// SetLeveledOwner implements Bonus.
func (s *SpellPointBonus) SetLeveledOwner(owner LeveledOwner) {
	s.LeveledOwner = owner
}

// AddToTooltip implements Bonus.
func (s *SpellPointBonus) AddToTooltip(buffer *xbytes.InsertBuffer) {
	s.basicAddToTooltip(&s.LeveledAmount, buffer)
}

// MatchForType returns true if this spell bonus matches the data for its match type.
func (s *SpellPointBonus) MatchForType(replacements map[string]string, name, powerSource string, colleges []string) bool {
	return s.SpellMatchType.MatchForType(s.NameCriteria, replacements, name, powerSource, colleges)
}

// Hash writes this object's contents into the hasher.
func (s *SpellPointBonus) Hash(h hash.Hash) {
	if s == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, s.Type)
	xhash.Num8(h, s.SpellMatchType)
	s.NameCriteria.Hash(h)
	s.TagsCriteria.Hash(h)
	s.LeveledAmount.Hash(h)
}

// MarshalJSONTo implements json.MarshalerTo.
func (s *SpellPointBonus) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, &s.SpellPointBonusData)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (s *SpellPointBonus) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var content struct {
		SpellPointBonusData
		OldTagsCriteria criteria.Text `json:"category"`
	}
	if err := json.UnmarshalDecode(dec, &content); err != nil {
		return err
	}
	s.SpellPointBonusData = content.SpellPointBonusData
	if s.TagsCriteria.IsZero() && !content.OldTagsCriteria.IsZero() {
		s.TagsCriteria = content.OldTagsCriteria
	}
	return nil
}
