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
	// SpellPointsID holds the ID for spell point lookups.
	SpellPointsID = "spell.points"
	// SpellCollegePointsID holds the ID for spell college point lookups.
	SpellCollegePointsID = "spell.college.points"
	// SpellPowerSourcePointsID holds the ID for spell power source point lookups.
	SpellPowerSourcePointsID = "spell.power_source.points"
)

var _ Bonus = &SpellPointBonus{}

// SpellPointBonus holds an adjustment to a spell's points.
type SpellPointBonus struct {
	Type           Type            `json:"type"`
	SpellMatchType spell.MatchType `json:"match"`
	NameCriteria   criteria.String `json:"name,omitempty"`
	TagsCriteria   criteria.String `json:"tags,alt=category,omitempty"`
	LeveledAmount
	owner fmt.Stringer
}

// NewSpellPointBonus creates a new SpellPointBonus.
func NewSpellPointBonus() *SpellPointBonus {
	return &SpellPointBonus{
		Type:           SpellPointBonusType,
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
func (s *SpellPointBonus) FeatureType() Type {
	return s.Type
}

// Clone implements Feature.
func (s *SpellPointBonus) Clone() Feature {
	other := *s
	return &other
}

// FeatureMapKey implements Feature.
func (s *SpellPointBonus) FeatureMapKey() string {
	if s.TagsCriteria.Compare != criteria.Any {
		return SpellPointsID + "*"
	}
	switch s.SpellMatchType {
	case spell.AllColleges:
		return SpellCollegePointsID
	case spell.CollegeName:
		return s.buildKey(SpellCollegePointsID)
	case spell.PowerSource:
		return s.buildKey(SpellPowerSourcePointsID)
	case spell.Spell:
		return s.buildKey(SpellPointsID)
	default:
		jot.Fatal(1, "invalid match type: ", s.SpellMatchType)
		return ""
	}
}

func (s *SpellPointBonus) buildKey(prefix string) string {
	if s.NameCriteria.Compare == criteria.Is {
		return prefix + "/" + s.NameCriteria.Qualifier
	}
	return prefix + "*"
}

// FillWithNameableKeys implements Feature.
func (s *SpellPointBonus) FillWithNameableKeys(m map[string]string) {
	if s.SpellMatchType != spell.AllColleges {
		nameables.Extract(s.NameCriteria.Qualifier, m)
	}
	nameables.Extract(s.TagsCriteria.Qualifier, m)
}

// ApplyNameableKeys implements Feature.
func (s *SpellPointBonus) ApplyNameableKeys(m map[string]string) {
	if s.SpellMatchType != spell.AllColleges {
		s.NameCriteria.Qualifier = nameables.Apply(s.NameCriteria.Qualifier, m)
	}
	s.TagsCriteria.Qualifier = nameables.Apply(s.TagsCriteria.Qualifier, m)
}

// Owner implements Bonus.
func (s *SpellPointBonus) Owner() fmt.Stringer {
	return s.owner
}

// SetOwner implements Bonus.
func (s *SpellPointBonus) SetOwner(owner fmt.Stringer) {
	s.owner = owner
}

// SetLevel implements Bonus.
func (s *SpellPointBonus) SetLevel(level fxp.Int) {
	s.Level = level
}

// AddToTooltip implements Bonus.
func (s *SpellPointBonus) AddToTooltip(buffer *xio.ByteBuffer) {
	basicAddToTooltip(s.owner, &s.LeveledAmount, buffer)
}
