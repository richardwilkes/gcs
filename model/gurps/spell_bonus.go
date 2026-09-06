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

var _ Bonus = &SpellBonus{}

// SpellBonus holds the data for a bonus to a spell.
type SpellBonus struct {
	SpellBonusData
}

// SpellBonusData holds the persisted data shared by SpellBonus and SpellPointBonus, which differ only in the feature
// type they carry, along with the behavior that depends on nothing but that data.
type SpellBonusData struct {
	Type feature.Type `json:"type"`
	FeatureSwitch
	SpellMatchType spellmatch.Type `json:"match"`
	NameCriteria   criteria.Text   `json:"name,omitzero"`
	TagsCriteria   criteria.Text   `json:"tags,omitzero"`
	LeveledAmount
	BonusOwner `json:"-"`
}

// NewSpellBonus creates a new SpellBonus.
func NewSpellBonus() *SpellBonus {
	return &SpellBonus{SpellBonusData: newSpellBonusData(feature.SpellBonus)}
}

// newSpellBonusData returns the initial data for a spell bonus of the given feature type: one that applies to all
// colleges and whose amount is one.
func newSpellBonusData(featureType feature.Type) SpellBonusData {
	var s SpellBonusData
	s.Type = featureType
	s.SpellMatchType = spellmatch.AllColleges
	s.NameCriteria.Compare = criteria.IsText
	s.TagsCriteria.Compare = criteria.AnyText
	s.Amount = fxp.One
	return s
}

// FeatureType implements Feature.
func (s *SpellBonusData) FeatureType() feature.Type {
	return s.Type
}

// FillWithNameableKeys implements Feature.
func (s *SpellBonusData) FillWithNameableKeys(m, existing map[string]string) {
	if s.SpellMatchType != spellmatch.AllColleges {
		nameable.Extract(m, existing, s.NameCriteria.Qualifier)
	}
	nameable.Extract(m, existing, s.TagsCriteria.Qualifier)
}

// SetLeveledOwner implements Bonus.
func (s *SpellBonusData) SetLeveledOwner(owner LeveledOwner) {
	s.LeveledOwner = owner
}

// AddToTooltip implements Bonus.
func (s *SpellBonusData) AddToTooltip(buffer *xbytes.InsertBuffer) {
	s.basicAddToTooltip(&s.LeveledAmount, buffer)
}

// MatchesSpell returns true if this bonus applies to the spell with the given name, power source, colleges and
// tags, according to its match type.
func (s *SpellBonusData) MatchesSpell(replacements map[string]string, name, powerSource string, colleges, tags []string) bool {
	return s.TagsCriteria.MatchesList(replacements, tags...) &&
		s.SpellMatchType.MatchForType(s.NameCriteria, replacements, name, powerSource, colleges)
}

// Hash writes the data into the hasher.
func (s *SpellBonusData) Hash(h hash.Hash) {
	xhash.Num8(h, s.Type)
	xhash.Bool(h, s.Switchable)
	xhash.Num8(h, s.SpellMatchType)
	s.NameCriteria.Hash(h)
	s.TagsCriteria.Hash(h)
	s.LeveledAmount.Hash(h)
}

// Clone implements Feature.
func (s *SpellBonus) Clone() Feature {
	return clonePtr(s)
}

// Hash writes this object's contents into the hasher.
func (s *SpellBonus) Hash(h hash.Hash) {
	if s == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	s.SpellBonusData.Hash(h)
}

// MarshalJSONTo implements json.MarshalerTo.
func (s *SpellBonus) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, &s.SpellBonusData)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (s *SpellBonus) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return unmarshalWithLegacyTags(dec, &s.SpellBonusData, &s.TagsCriteria)
}
