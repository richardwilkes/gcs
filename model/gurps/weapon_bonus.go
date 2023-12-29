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
	"fmt"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

var _ Bonus = &WeaponBonus{}

// WeaponBonus holds the data for an adjustment to weapon damage.
type WeaponBonus struct {
	Type                   feature.Type    `json:"type"`
	Percent                bool            `json:"percent,omitempty"`
	SwitchTypeValue        bool            `json:"switch_type_value,omitempty"`
	SelectionType          wsel.Type       `json:"selection_type"`
	SwitchType             wswitch.Type    `json:"switch_type,omitempty"`
	NameCriteria           StringCriteria  `json:"name,omitempty"`
	SpecializationCriteria StringCriteria  `json:"specialization,omitempty"`
	RelativeLevelCriteria  NumericCriteria `json:"level,omitempty"`
	UsageCriteria          StringCriteria  `json:"usage,omitempty"`
	TagsCriteria           StringCriteria  `json:"tags,alt=category,omitempty"`
	WeaponLeveledAmount
	BonusOwner
}

// NewWeaponDamageBonus creates a new weapon damage bonus.
func NewWeaponDamageBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponBonus)
}

// NewWeaponDRDivisorBonus creates a new weapon DR divisor bonus.
func NewWeaponDRDivisorBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponDRDivisorBonus)
}

// NewWeaponMinSTBonus creates a new weapon minimum ST bonus.
func NewWeaponMinSTBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponMinSTBonus)
}

// NewWeaponMinReachBonus creates a new weapon minimum reach bonus.
func NewWeaponMinReachBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponMinReachBonus)
}

// NewWeaponMaxReachBonus creates a new weapon maximum reach bonus.
func NewWeaponMaxReachBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponMaxReachBonus)
}

// NewWeaponHalfDamageRangeBonus creates a new weapon half-damage range bonus.
func NewWeaponHalfDamageRangeBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponHalfDamageRangeBonus)
}

// NewWeaponMinRangeBonus creates a new weapon minimum range bonus.
func NewWeaponMinRangeBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponMinRangeBonus)
}

// NewWeaponMaxRangeBonus creates a new weapon maximum range bonus.
func NewWeaponMaxRangeBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponMaxRangeBonus)
}

// NewWeaponAccBonus creates a new weapon accuracy bonus.
func NewWeaponAccBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponAccBonus)
}

// NewWeaponScopeAccBonus creates a new weapon scope accuracy bonus.
func NewWeaponScopeAccBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponScopeAccBonus)
}

// NewWeaponBulkBonus creates a new weapon bulk bonus.
func NewWeaponBulkBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponBulkBonus)
}

// NewWeaponRecoilBonus creates a new weapon recoil bonus.
func NewWeaponRecoilBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponRecoilBonus)
}

// NewWeaponParryBonus creates a new weapon parry bonus.
func NewWeaponParryBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponParryBonus)
}

// NewWeaponBlockBonus creates a new weapon block bonus.
func NewWeaponBlockBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponBlockBonus)
}

// NewWeaponRofMode1ShotsBonus creates a new weapon rate of fire mode 1 shots per attack bonus.
func NewWeaponRofMode1ShotsBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponRofMode1ShotsBonus)
}

// NewWeaponRofMode1SecondaryBonus creates a new weapon rate of fire mode 1 secondary projectile bonus.
func NewWeaponRofMode1SecondaryBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponRofMode1SecondaryBonus)
}

// NewWeaponRofMode2ShotsBonus creates a new weapon rate of fire mode 2 shots per attack bonus.
func NewWeaponRofMode2ShotsBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponRofMode2ShotsBonus)
}

// NewWeaponRofMode2SecondaryBonus creates a new weapon rate of fire mode 2 secondary projectile bonus.
func NewWeaponRofMode2SecondaryBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponRofMode2SecondaryBonus)
}

// NewWeaponNonChamberShotsBonus creates a new weapon non-chamber shots bonus.
func NewWeaponNonChamberShotsBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponNonChamberShotsBonus)
}

// NewWeaponChamberShotsBonus creates a new weapon chamber shots bonus.
func NewWeaponChamberShotsBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponChamberShotsBonus)
}

// NewWeaponShotDurationBonus creates a new weapon shot duration bonus.
func NewWeaponShotDurationBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponShotDurationBonus)
}

// NewWeaponReloadTimeBonus creates a new weapon reload time bonus.
func NewWeaponReloadTimeBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponReloadTimeBonus)
}

// NewWeaponSwitchBonus creates a new weapon switch bonus.
func NewWeaponSwitchBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponSwitch)
}

func newWeaponBonus(t feature.Type) *WeaponBonus {
	return &WeaponBonus{
		Type:          t,
		SelectionType: wsel.WithRequiredSkill,
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
		UsageCriteria: StringCriteria{
			StringCriteriaData: StringCriteriaData{
				Compare: AnyString,
			},
		},
		TagsCriteria: StringCriteria{
			StringCriteriaData: StringCriteriaData{
				Compare: AnyString,
			},
		},
		WeaponLeveledAmount: WeaponLeveledAmount{Amount: fxp.One},
	}
}

// FeatureType implements Feature.
func (w *WeaponBonus) FeatureType() feature.Type {
	return w.Type
}

// Clone implements Feature.
func (w *WeaponBonus) Clone() Feature {
	other := *w
	return &other
}

// AdjustedAmount implements Bonus.
func (w *WeaponBonus) AdjustedAmount() fxp.Int {
	return w.WeaponLeveledAmount.AdjustedAmount()
}

// AdjustedAmountForWeapon returns the adjusted amount for the given weapon.
func (w *WeaponBonus) AdjustedAmountForWeapon(wpn *Weapon) fxp.Int {
	w.WeaponLeveledAmount.DieCount = fxp.From(wpn.Damage.BaseDamageDice().Count)
	return w.WeaponLeveledAmount.AdjustedAmount()
}

// FillWithNameableKeys implements Feature.
func (w *WeaponBonus) FillWithNameableKeys(m map[string]string) {
	Extract(w.SpecializationCriteria.Qualifier, m)
	if w.SelectionType != wsel.ThisWeapon {
		Extract(w.NameCriteria.Qualifier, m)
		Extract(w.SpecializationCriteria.Qualifier, m)
		Extract(w.UsageCriteria.Qualifier, m)
		Extract(w.TagsCriteria.Qualifier, m)
	}
}

// ApplyNameableKeys implements Feature.
func (w *WeaponBonus) ApplyNameableKeys(m map[string]string) {
	w.SpecializationCriteria.Qualifier = Apply(w.SpecializationCriteria.Qualifier, m)
	if w.SelectionType != wsel.ThisWeapon {
		w.NameCriteria.Qualifier = Apply(w.NameCriteria.Qualifier, m)
		w.SpecializationCriteria.Qualifier = Apply(w.SpecializationCriteria.Qualifier, m)
		w.UsageCriteria.Qualifier = Apply(w.UsageCriteria.Qualifier, m)
		w.TagsCriteria.Qualifier = Apply(w.TagsCriteria.Qualifier, m)
	}
}

// SetLevel implements Bonus.
func (w *WeaponBonus) SetLevel(level fxp.Int) {
	w.WeaponLeveledAmount.Level = level
}

// AddToTooltip implements Bonus.
func (w *WeaponBonus) AddToTooltip(buffer *xio.ByteBuffer) {
	if buffer != nil {
		var buf strings.Builder
		buf.WriteByte('\n')
		buf.WriteString(w.parentName())
		buf.WriteString(" [")
		if w.Type == feature.WeaponSwitch {
			fmt.Fprintf(&buf, "%v set to %v", w.SwitchType, w.SwitchTypeValue)
		} else {
			buf.WriteString(w.WeaponLeveledAmount.Format(w.Percent))
			buf.WriteString(i18n.Text(" to "))
			switch w.Type {
			case feature.WeaponBonus:
				buf.WriteString(i18n.Text("damage"))
			case feature.WeaponAccBonus:
				buf.WriteString(i18n.Text("weapon accuracy"))
			case feature.WeaponScopeAccBonus:
				buf.WriteString(i18n.Text("scope accuracy"))
			case feature.WeaponDRDivisorBonus:
				buf.WriteString(i18n.Text("DR divisor"))
			case feature.WeaponMinSTBonus:
				buf.WriteString(i18n.Text("minimum ST"))
			case feature.WeaponMinReachBonus:
				buf.WriteString(i18n.Text("minimum reach"))
			case feature.WeaponMaxReachBonus:
				buf.WriteString(i18n.Text("maximum reach"))
			case feature.WeaponHalfDamageRangeBonus:
				buf.WriteString(i18n.Text("half-damage range"))
			case feature.WeaponMinRangeBonus:
				buf.WriteString(i18n.Text("minimum range"))
			case feature.WeaponMaxRangeBonus:
				buf.WriteString(i18n.Text("maximum range"))
			case feature.WeaponBulkBonus:
				buf.WriteString(i18n.Text("bulk"))
			case feature.WeaponRecoilBonus:
				buf.WriteString(i18n.Text("recoil"))
			case feature.WeaponParryBonus:
				buf.WriteString(i18n.Text("parry"))
			case feature.WeaponBlockBonus:
				buf.WriteString(i18n.Text("block"))
			case feature.WeaponRofMode1ShotsBonus, feature.WeaponRofMode2ShotsBonus:
				buf.WriteString(i18n.Text("shots per attack"))
			case feature.WeaponRofMode1SecondaryBonus, feature.WeaponRofMode2SecondaryBonus:
				buf.WriteString(i18n.Text("secondary projectiles"))
			case feature.WeaponNonChamberShotsBonus:
				buf.WriteString(i18n.Text("non-chamber shots"))
			case feature.WeaponChamberShotsBonus:
				buf.WriteString(i18n.Text("chamber shots"))
			case feature.WeaponShotDurationBonus:
				buf.WriteString(i18n.Text("shot duration"))
			case feature.WeaponReloadTimeBonus:
				buf.WriteString(i18n.Text("reload time"))
			default:
			}
		}
		buf.WriteByte(']')
		buffer.WriteString(buf.String())
	}
}
