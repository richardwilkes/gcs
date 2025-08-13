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
	"encoding/json"
	"hash"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellmatch"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Bonus = &SpellBonus{}

// SpellBonus holds the data for a bonus to a spell.
type SpellBonus struct {
	SpellBonusData
}

// SpellBonusData holds the data for a bonus to a spell which is persisted.
type SpellBonusData struct {
	Type           feature.Type    `json:"type"`
	SpellMatchType spellmatch.Type `json:"match"`
	NameCriteria   criteria.Text   `json:"name,omitzero"`
	TagsCriteria   criteria.Text   `json:"tags,omitzero"`
	LeveledAmount
	BonusOwner
}

// NewSpellBonus creates a new SpellBonus.
func NewSpellBonus() *SpellBonus {
	var s SpellBonus
	s.Type = feature.SpellBonus
	s.SpellMatchType = spellmatch.AllColleges
	s.NameCriteria.Compare = criteria.IsText
	s.TagsCriteria.Compare = criteria.AnyText
	s.Amount = fxp.One
	return &s
}

// FeatureType implements Feature.
func (s *SpellBonus) FeatureType() feature.Type {
	return s.Type
}

// Clone implements Feature.
func (s *SpellBonus) Clone() Feature {
	other := *s
	return &other
}

// FillWithNameableKeys implements Feature.
func (s *SpellBonus) FillWithNameableKeys(m, existing map[string]string) {
	if s.SpellMatchType != spellmatch.AllColleges {
		nameable.Extract(s.NameCriteria.Qualifier, m, existing)
	}
	nameable.Extract(s.TagsCriteria.Qualifier, m, existing)
}

// SetLevel implements Bonus.
func (s *SpellBonus) SetLevel(level fxp.Int) {
	s.Level = level
}

// AddToTooltip implements Bonus.
func (s *SpellBonus) AddToTooltip(buffer *xbytes.InsertBuffer) {
	s.basicAddToTooltip(&s.LeveledAmount, buffer)
}

// MatchForType returns true if this spell bonus matches the data for its match type.
func (s *SpellBonus) MatchForType(replacements map[string]string, name, powerSource string, colleges []string) bool {
	return s.SpellMatchType.MatchForType(s.NameCriteria, replacements, name, powerSource, colleges)
}

// Hash writes this object's contents into the hasher.
func (s *SpellBonus) Hash(h hash.Hash) {
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

// MarshalJSON implements json.Marshaler.
func (s *SpellBonus) MarshalJSON() ([]byte, error) {
	return json.Marshal(&s.SpellBonusData)
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *SpellBonus) UnmarshalJSON(data []byte) error {
	var content struct {
		SpellBonusData
		OldTagsCriteria criteria.Text `json:"category"`
	}
	if err := json.Unmarshal(data, &content); err != nil {
		return err
	}
	s.SpellBonusData = content.SpellBonusData
	if s.TagsCriteria.IsZero() && !content.OldTagsCriteria.IsZero() {
		s.TagsCriteria = content.OldTagsCriteria
	}
	return nil
}
