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
	"strings"

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
	NameCriteria           StringCriteria      `json:"name,omitempty"`
	SpecializationCriteria StringCriteria      `json:"specialization,omitempty"`
	RelativeLevelCriteria  NumericCriteria     `json:"level,omitempty"`
	TagsCriteria           StringCriteria      `json:"tags,alt=category,omitempty"`
	LeveledAmount
	BonusOwner
}

// NewWeaponDamageBonus creates a new weapon damage bonus.
func NewWeaponDamageBonus() *WeaponBonus {
	return newWeaponBonus(WeaponBonusFeatureType)
}

// NewWeaponDRDivisorBonus creates a new weapon DR divisor bonus.
func NewWeaponDRDivisorBonus() *WeaponBonus {
	return newWeaponBonus(WeaponDRDivisorBonusFeatureType)
}

// NewWeaponMinSTBonus creates a new weapon minimum ST bonus.
func NewWeaponMinSTBonus() *WeaponBonus {
	return newWeaponBonus(WeaponMinSTBonusFeatureType)
}

// NewWeaponAccBonus creates a new weapon accuracy bonus.
func NewWeaponAccBonus() *WeaponBonus {
	return newWeaponBonus(WeaponAccBonusFeatureType)
}

// NewWeaponScopeAccBonus creates a new weapon scope accuracy bonus.
func NewWeaponScopeAccBonus() *WeaponBonus {
	return newWeaponBonus(WeaponScopeAccBonusFeatureType)
}

// NewWeaponBulkBonus creates a new weapon bulk bonus.
func NewWeaponBulkBonus() *WeaponBonus {
	return newWeaponBonus(WeaponBulkBonusFeatureType)
}

// NewWeaponRecoilBonus creates a new weapon recoil bonus.
func NewWeaponRecoilBonus() *WeaponBonus {
	return newWeaponBonus(WeaponRecoilBonusFeatureType)
}

func newWeaponBonus(t FeatureType) *WeaponBonus {
	return &WeaponBonus{
		Type:          t,
		SelectionType: WithRequiredSkillWeaponSelectionType,
		NameCriteria: StringCriteria{
			StringCriteriaData: StringCriteriaData{
				Compare: IsString,
			},
		},
		SpecializationCriteria: StringCriteria{
			StringCriteriaData: StringCriteriaData{
				Compare: AnyString,
			},
		},
		RelativeLevelCriteria: NumericCriteria{
			NumericCriteriaData: NumericCriteriaData{
				Compare: AtLeastNumber,
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

// SetLevel implements Bonus.
func (w *WeaponBonus) SetLevel(level fxp.Int) {
	w.Level = level
}

// AddToTooltip implements Bonus.
func (w *WeaponBonus) AddToTooltip(buffer *xio.ByteBuffer) {
	if buffer != nil {
		var buf strings.Builder
		buf.WriteByte('\n')
		buf.WriteString(w.parentName())
		buf.WriteString(" [")
		switch w.Type {
		case WeaponBonusFeatureType:
			buf.WriteString(w.LeveledAmount.Format(w.Percent, i18n.Text("die")))
			buf.WriteString(i18n.Text(" to damage"))
		case WeaponAccBonusFeatureType:
			buf.WriteString(w.LeveledAmount.FormatWithLevel(w.Percent))
			buf.WriteString(i18n.Text(" to weapon accuracy"))
		case WeaponScopeAccBonusFeatureType:
			buf.WriteString(w.LeveledAmount.FormatWithLevel(w.Percent))
			buf.WriteString(i18n.Text(" to scope accuracy"))
		case WeaponDRDivisorBonusFeatureType:
			buf.WriteString(w.LeveledAmount.FormatWithLevel(w.Percent))
			buf.WriteString(i18n.Text(" to DR divisor"))
		case WeaponMinSTBonusFeatureType:
			buf.WriteString(w.LeveledAmount.FormatWithLevel(w.Percent))
			buf.WriteString(i18n.Text(" to minimum ST"))
		case WeaponBulkBonusFeatureType:
			buf.WriteString(w.LeveledAmount.FormatWithLevel(w.Percent))
			buf.WriteString(i18n.Text(" to bulk"))
		case WeaponRecoilBonusFeatureType:
			buf.WriteString(w.LeveledAmount.FormatWithLevel(w.Percent))
			buf.WriteString(i18n.Text(" to recoil"))
		default:
			return
		}
		buf.WriteByte(']')
		buffer.WriteString(buf.String())
	}
}
