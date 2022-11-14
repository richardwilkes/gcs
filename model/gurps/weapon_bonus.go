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
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

var _ Bonus = &WeaponBonus{}

// WeaponBonus holds the data for an adjustment to weapon damage.
type WeaponBonus struct {
	Type                   FeatureType         `json:"type"`
	Percent                bool                `json:"percent,omitempty"`
	SelectionType          WeaponSelectionType `json:"selection_type"`
	NameCriteria           criteria.String     `json:"name,omitempty"`
	SpecializationCriteria criteria.String     `json:"specialization,omitempty"`
	RelativeLevelCriteria  criteria.Numeric    `json:"level,omitempty"`
	TagsCriteria           criteria.String     `json:"tags,alt=category,omitempty"`
	LeveledAmount
	owner fmt.Stringer
}

// NewWeaponDamageBonus creates a new weapon damage bonus.
func NewWeaponDamageBonus() *WeaponBonus {
	return newWeaponDamageBonus(WeaponBonusFeatureType)
}

// NewWeaponDRDivisorBonus creates a new weapon DR divisor bonus.
func NewWeaponDRDivisorBonus() *WeaponBonus {
	return newWeaponDamageBonus(WeaponDRDivisorBonusFeatureType)
}

func newWeaponDamageBonus(t FeatureType) *WeaponBonus {
	return &WeaponBonus{
		Type:          t,
		SelectionType: WithRequiredSkillWeaponSelectionType,
		NameCriteria: criteria.String{
			StringData: criteria.StringData{
				Compare: criteria.Is,
			},
		},
		SpecializationCriteria: criteria.String{
			StringData: criteria.StringData{
				Compare: criteria.Any,
			},
		},
		RelativeLevelCriteria: criteria.Numeric{
			NumericData: criteria.NumericData{
				Compare: criteria.AtLeast,
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
func (w *WeaponBonus) FeatureType() FeatureType {
	return w.Type
}

// Clone implements Feature.
func (w *WeaponBonus) Clone() Feature {
	other := *w
	return &other
}

// FillWithNameableKeys implements Feature.
func (w *WeaponBonus) FillWithNameableKeys(m map[string]string) {
	Extract(w.SpecializationCriteria.Qualifier, m)
	if w.SelectionType != ThisWeaponWeaponSelectionType {
		Extract(w.NameCriteria.Qualifier, m)
		Extract(w.SpecializationCriteria.Qualifier, m)
		Extract(w.TagsCriteria.Qualifier, m)
	}
}

// ApplyNameableKeys implements Feature.
func (w *WeaponBonus) ApplyNameableKeys(m map[string]string) {
	w.SpecializationCriteria.Qualifier = Apply(w.SpecializationCriteria.Qualifier, m)
	if w.SelectionType != ThisWeaponWeaponSelectionType {
		w.NameCriteria.Qualifier = Apply(w.NameCriteria.Qualifier, m)
		w.SpecializationCriteria.Qualifier = Apply(w.SpecializationCriteria.Qualifier, m)
		w.TagsCriteria.Qualifier = Apply(w.TagsCriteria.Qualifier, m)
	}
}

// Owner implements Bonus.
func (w *WeaponBonus) Owner() fmt.Stringer {
	return w.owner
}

// SetOwner implements Bonus.
func (w *WeaponBonus) SetOwner(owner fmt.Stringer) {
	w.owner = owner
}

// SetLevel implements Bonus.
func (w *WeaponBonus) SetLevel(level fxp.Int) {
	w.Level = level
}

// AddToTooltip implements Bonus.
func (w *WeaponBonus) AddToTooltip(buffer *xio.ByteBuffer) {
	if buffer != nil {
		buffer.WriteByte('\n')
		buffer.WriteString(parentName(w.owner))
		buffer.WriteString(" [")
		if w.Type == WeaponBonusFeatureType {
			buffer.WriteString(w.LeveledAmount.Format(w.Percent, i18n.Text("die")))
			buffer.WriteString(i18n.Text(" to damage"))
		} else {
			buffer.WriteString(w.LeveledAmount.FormatWithLevel(w.Percent))
			buffer.WriteString(i18n.Text(" to DR divisor"))
		}
		buffer.WriteByte(']')
	}
}
