/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/xio"
)

var _ Bonus = &SpellPointBonus{}

// SpellPointBonus holds an adjustment to a spell's points.
type SpellPointBonus struct {
	Type           FeatureType    `json:"type"`
	SpellMatchType SpellMatchType `json:"match"`
	NameCriteria   StringCriteria `json:"name,omitempty"`
	TagsCriteria   StringCriteria `json:"tags,alt=category,omitempty"`
	LeveledAmount
	BonusOwner
}

// NewSpellPointBonus creates a new SpellPointBonus.
func NewSpellPointBonus() *SpellPointBonus {
	return &SpellPointBonus{
		Type:           SpellPointBonusFeatureType,
		SpellMatchType: AllCollegesSpellMatchType,
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
func (s *SpellPointBonus) FeatureType() FeatureType {
	return s.Type
}

// Clone implements Feature.
func (s *SpellPointBonus) Clone() Feature {
	other := *s
	return &other
}

// FillWithNameableKeys implements Feature.
func (s *SpellPointBonus) FillWithNameableKeys(m map[string]string) {
	if s.SpellMatchType != AllCollegesSpellMatchType {
		Extract(s.NameCriteria.Qualifier, m)
	}
	Extract(s.TagsCriteria.Qualifier, m)
}

// ApplyNameableKeys implements Feature.
func (s *SpellPointBonus) ApplyNameableKeys(m map[string]string) {
	if s.SpellMatchType != AllCollegesSpellMatchType {
		s.NameCriteria.Qualifier = Apply(s.NameCriteria.Qualifier, m)
	}
	s.TagsCriteria.Qualifier = Apply(s.TagsCriteria.Qualifier, m)
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
