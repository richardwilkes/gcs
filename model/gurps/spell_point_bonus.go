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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellmatch"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var _ Bonus = &SpellPointBonus{}

// SpellPointBonus holds an adjustment to a spell's points.
type SpellPointBonus struct {
	Type           feature.Type    `json:"type"`
	SpellMatchType spellmatch.Type `json:"match"`
	NameCriteria   criteria.Text   `json:"name,omitempty"`
	TagsCriteria   criteria.Text   `json:"tags,alt=category,omitempty"`
	LeveledAmount
	BonusOwner
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

// SetLevel implements Bonus.
func (s *SpellPointBonus) SetLevel(level fxp.Int) {
	s.Level = level
}

// AddToTooltip implements Bonus.
func (s *SpellPointBonus) AddToTooltip(buffer *xio.ByteBuffer) {
	s.basicAddToTooltip(&s.LeveledAmount, buffer)
}

// MatchForType returns true if this spell bonus matches the data for its match type.
func (s *SpellPointBonus) MatchForType(replacements map[string]string, name, powerSource string, colleges []string) bool {
	return s.SpellMatchType.MatchForType(s.NameCriteria, replacements, name, powerSource, colleges)
}

// Hash writes this object's contents into the hasher.
func (s *SpellPointBonus) Hash(h hash.Hash) {
	if s == nil {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.Num8(h, s.Type)
	hashhelper.Num8(h, s.SpellMatchType)
	s.NameCriteria.Hash(h)
	s.TagsCriteria.Hash(h)
	s.LeveledAmount.Hash(h)
}
