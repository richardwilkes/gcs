// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"encoding/binary"
	"hash"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellmatch"
	"github.com/richardwilkes/toolbox/xio"
)

var _ Bonus = &SpellPointBonus{}

// SpellPointBonus holds an adjustment to a spell's points.
type SpellPointBonus struct {
	Type           feature.Type    `json:"type"`
	SpellMatchType spellmatch.Type `json:"match"`
	NameCriteria   StringCriteria  `json:"name,omitempty"`
	TagsCriteria   StringCriteria  `json:"tags,alt=category,omitempty"`
	LeveledAmount
	BonusOwner
}

// NewSpellPointBonus creates a new SpellPointBonus.
func NewSpellPointBonus() *SpellPointBonus {
	return &SpellPointBonus{
		Type:           feature.SpellPointBonus,
		SpellMatchType: spellmatch.AllColleges,
		NameCriteria: StringCriteria{
			StringCriteriaData: StringCriteriaData{
				Compare: IsString,
			},
		},
		TagsCriteria: StringCriteria{
			StringCriteriaData: StringCriteriaData{
				Compare: AnyString,
			},
		},
		LeveledAmount: LeveledAmount{Amount: fxp.One},
	}
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
func (s *SpellPointBonus) FillWithNameableKeys(m map[string]string) {
	if s.SpellMatchType != spellmatch.AllColleges {
		ExtractNameables(s.NameCriteria.Qualifier, m)
	}
	ExtractNameables(s.TagsCriteria.Qualifier, m)
}

// ApplyNameableKeys implements Feature.
func (s *SpellPointBonus) ApplyNameableKeys(m map[string]string) {
	if s.SpellMatchType != spellmatch.AllColleges {
		s.NameCriteria.Qualifier = ApplyNameables(s.NameCriteria.Qualifier, m)
	}
	s.TagsCriteria.Qualifier = ApplyNameables(s.TagsCriteria.Qualifier, m)
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
func (s *SpellPointBonus) MatchForType(name, powerSource string, colleges []string) bool {
	return s.SpellMatchType.MatchForType(s.NameCriteria, name, powerSource, colleges)
}

// Hash writes this object's contents into the hasher.
func (s *SpellPointBonus) Hash(h hash.Hash) {
	if s == nil {
		return
	}
	_ = binary.Write(h, binary.LittleEndian, s.Type)
	_ = binary.Write(h, binary.LittleEndian, s.SpellMatchType)
	s.NameCriteria.Hash(h)
	s.TagsCriteria.Hash(h)
	s.LeveledAmount.Hash(h)
}
