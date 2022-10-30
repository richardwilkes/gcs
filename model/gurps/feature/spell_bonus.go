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

package feature

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/nameables"
	"github.com/richardwilkes/gcs/v5/model/gurps/spell"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xio"
)

const (
	// SpellNameID holds the ID for spell name lookups.
	SpellNameID = "spell.name"
	// SpellCollegeID holds the ID for spell college name lookups.
	SpellCollegeID = "spell.college"
	// SpellPowerSourceID holds the ID for spell power source name lookups.
	SpellPowerSourceID = "spell.power_source"
)

var _ Bonus = &SpellBonus{}

// SpellBonus holds the data for a bonus to a spell.
type SpellBonus struct {
	Type           Type            `json:"type"`
	SpellMatchType spell.MatchType `json:"match"`
	NameCriteria   criteria.String `json:"name,omitempty"`
	TagsCriteria   criteria.String `json:"tags,alt=category,omitempty"`
	LeveledAmount
	owner fmt.Stringer
}

// NewSpellBonus creates a new SpellBonus.
func NewSpellBonus() *SpellBonus {
	return &SpellBonus{
		Type:           SpellBonusType,
		SpellMatchType: spell.AllColleges,
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
func (s *SpellBonus) FeatureType() Type {
	return s.Type
}

// Clone implements Feature.
func (s *SpellBonus) Clone() Feature {
	other := *s
	return &other
}

// FeatureMapKey implements Feature.
func (s *SpellBonus) FeatureMapKey() string {
	if s.TagsCriteria.Compare != criteria.Any {
		return SpellNameID + "*"
	}
	switch s.SpellMatchType {
	case spell.AllColleges:
		return SpellCollegeID
	case spell.CollegeName:
		return s.buildKey(SpellCollegeID)
	case spell.PowerSource:
		return s.buildKey(SpellPowerSourceID)
	case spell.Spell:
		return s.buildKey(SpellNameID)
	default:
		jot.Fatal(1, "invalid match type: ", s.SpellMatchType)
		return ""
	}
}

func (s *SpellBonus) buildKey(prefix string) string {
	if s.NameCriteria.Compare == criteria.Is {
		return prefix + "/" + s.NameCriteria.Qualifier
	}
	return prefix + "*"
}

// FillWithNameableKeys implements Feature.
func (s *SpellBonus) FillWithNameableKeys(m map[string]string) {
	if s.SpellMatchType != spell.AllColleges {
		nameables.Extract(s.NameCriteria.Qualifier, m)
	}
	nameables.Extract(s.TagsCriteria.Qualifier, m)
}

// ApplyNameableKeys implements Feature.
func (s *SpellBonus) ApplyNameableKeys(m map[string]string) {
	if s.SpellMatchType != spell.AllColleges {
		s.NameCriteria.Qualifier = nameables.Apply(s.NameCriteria.Qualifier, m)
	}
	s.TagsCriteria.Qualifier = nameables.Apply(s.TagsCriteria.Qualifier, m)
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
