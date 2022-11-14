/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
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
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/xio"
)

var _ Bonus = &SpellBonus{}

// SpellBonus holds the data for a bonus to a spell.
type SpellBonus struct {
	Type           FeatureType     `json:"type"`
	SpellMatchType SpellMatchType  `json:"match"`
	NameCriteria   criteria.String `json:"name,omitempty"`
	TagsCriteria   criteria.String `json:"tags,alt=category,omitempty"`
	LeveledAmount
	owner fmt.Stringer
}

// NewSpellBonus creates a new SpellBonus.
func NewSpellBonus() *SpellBonus {
	return &SpellBonus{
		Type:           SpellBonusFeatureType,
		SpellMatchType: AllCollegesSpellMatchType,
		NameCriteria: criteria.String{
			StringData: criteria.StringData{
				Compare: criteria.Is,
			},
		},
		TagsCriteria: criteria.String{
			StringData: criteria.StringData{
				Compare: criteria.Any,
			},
		},
		LeveledAmount: LeveledAmount{Amount: fxp.One},
	}
}

// FeatureType implements Feature.
func (s *SpellBonus) FeatureType() FeatureType {
	return s.Type
}

// Clone implements Feature.
func (s *SpellBonus) Clone() Feature {
	other := *s
	return &other
}

// FillWithNameableKeys implements Feature.
func (s *SpellBonus) FillWithNameableKeys(m map[string]string) {
	if s.SpellMatchType != AllCollegesSpellMatchType {
		Extract(s.NameCriteria.Qualifier, m)
	}
	Extract(s.TagsCriteria.Qualifier, m)
}

// ApplyNameableKeys implements Feature.
func (s *SpellBonus) ApplyNameableKeys(m map[string]string) {
	if s.SpellMatchType != AllCollegesSpellMatchType {
		s.NameCriteria.Qualifier = Apply(s.NameCriteria.Qualifier, m)
	}
	s.TagsCriteria.Qualifier = Apply(s.TagsCriteria.Qualifier, m)
}

// Owner implements Bonus.
func (s *SpellBonus) Owner() fmt.Stringer {
	return s.owner
}

// SetOwner implements Bonus.
func (s *SpellBonus) SetOwner(owner fmt.Stringer) {
	s.owner = owner
}

// SetLevel implements Bonus.
func (s *SpellBonus) SetLevel(level fxp.Int) {
	s.Level = level
}

// AddToTooltip implements Bonus.
func (s *SpellBonus) AddToTooltip(buffer *xio.ByteBuffer) {
	basicAddToTooltip(s.owner, &s.LeveledAmount, buffer)
}

// MatchForType returns true if this spell bonus matches the data for its match type.
func (s *SpellBonus) MatchForType(name, powerSource string, colleges []string) bool {
	return s.SpellMatchType.MatchForType(s.NameCriteria, name, powerSource, colleges)
}
