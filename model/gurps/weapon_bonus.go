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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Bonus = &WeaponBonus{}

// WeaponBonus holds the data for an adjustment to weapon stats.
type WeaponBonus struct {
	WeaponBonusData
}

// WeaponBonusData holds the data for an adjustment to weapon stats which are persisted.
type WeaponBonusData struct { //nolint:govet // The field alignment here is poor, but kept to reduce diffs in the data
	Type                   feature.Type    `json:"type"`
	Percent                bool            `json:"percent,omitzero"`
	SelectionType          wsel.Type       `json:"selection_type"`
	SwitchType             wswitch.Type    `json:"switch_type,omitzero"`
	SwitchTypeValue        bool            `json:"switch_type_value,omitzero"`
	NameCriteria           criteria.Text   `json:"name,omitzero"`
	SpecializationCriteria criteria.Text   `json:"specialization,omitzero"`
	RelativeLevelCriteria  criteria.Number `json:"level,omitzero"`
	UsageCriteria          criteria.Text   `json:"usage,omitzero"`
	TagsCriteria           criteria.Text   `json:"tags,omitzero"`
	LeveledOwner           LeveledOwner    `json:"-"`
	DieCount               fxp.Int         `json:"-"`
	Amount                 fxp.Int         `json:"amount"`
	PerLevel               bool            `json:"leveled,omitzero"`
	PerDie                 bool            `json:"per_die,omitzero"`
	BonusOwner             `json:"-"`
}

// NewWeaponDamageBonus creates a new weapon damage bonus.
func NewWeaponDamageBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponBonus)
}

// NewWeaponDRDivisorBonus creates a new weapon DR divisor bonus.
func NewWeaponDRDivisorBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponDRDivisorBonus)
}

// NewWeaponEffectiveSTBonus creates a new weapon effective ST bonus.
func NewWeaponEffectiveSTBonus() *WeaponBonus {
	return newWeaponBonus(feature.WeaponEffectiveSTBonus)
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
	var w WeaponBonus
	w.Type = t
	w.SelectionType = wsel.WithRequiredSkill
	w.NameCriteria.Compare = criteria.IsText
	w.SpecializationCriteria.Compare = criteria.AnyText
	w.RelativeLevelCriteria.Compare = criteria.AtLeastNumber
	w.UsageCriteria.Compare = criteria.AnyText
	w.TagsCriteria.Compare = criteria.AnyText
	w.Amount = fxp.One
	return &w
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

// AdjustedAmountForWeapon returns the adjusted amount for the given weapon.
func (w *WeaponBonus) AdjustedAmountForWeapon(wpn *Weapon) fxp.Int {
	if w.Type == feature.WeaponMinSTBonus || w.Type == feature.WeaponEffectiveSTBonus {
		// Can't call BaseDamageDice() here because that would cause an infinite loop, so we just don't permit use of
		// the per-die feature for this bonus.
		w.DieCount = fxp.One
	} else {
		w.DieCount = fxp.FromInteger(wpn.Damage.BaseDamageDice().Count)
	}
	return w.AdjustedAmount()
}

// AdjustedAmount returns the amount, adjusted for level, if requested.
func (w *WeaponBonus) AdjustedAmount() fxp.Int {
	amt := w.Amount
	if w.PerDie {
		if w.DieCount < 0 {
			return 0
		}
		amt = amt.Mul(w.DieCount)
	}
	if w.PerLevel {
		leveledOwner := w.LeveledOwner
		if leveledOwner == nil {
			leveledOwner = w.DerivedLeveledOwner()
		}
		level := leveledOwner.CurrentLevel()
		if level < 0 {
			return 0
		}
		amt = amt.Mul(level)
	}
	return amt
}

// FillWithNameableKeys implements Feature.
func (w *WeaponBonus) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(w.SpecializationCriteria.Qualifier, m, existing)
	if w.SelectionType != wsel.ThisWeapon {
		nameable.Extract(w.NameCriteria.Qualifier, m, existing)
		nameable.Extract(w.UsageCriteria.Qualifier, m, existing)
		nameable.Extract(w.TagsCriteria.Qualifier, m, existing)
	}
}

// SetLeveledOwner implements Bonus.
func (w *WeaponBonus) SetLeveledOwner(owner LeveledOwner) {
	w.LeveledOwner = owner
}

// AddToTooltip implements Bonus.
func (w *WeaponBonus) AddToTooltip(buffer *xbytes.InsertBuffer) {
	if buffer != nil {
		var buf strings.Builder
		buf.WriteByte('\n')
		buf.WriteString(w.parentName())
		buf.WriteString(" [")
		if w.Type == feature.WeaponSwitch {
			fmt.Fprintf(&buf, "%v set to %v", w.SwitchType, w.SwitchTypeValue)
		} else {
			amt := w.Amount.StringWithSign()
			adjustedAmt := w.AdjustedAmount().StringWithSign()
			if w.Percent {
				amt += "%"
				adjustedAmt += "%"
			}
			switch {
			case w.PerDie && w.PerLevel:
				buf.WriteString(fmt.Sprintf(i18n.Text("%s (%s per die, per level)"), adjustedAmt, amt))
			case w.PerDie:
				buf.WriteString(fmt.Sprintf(i18n.Text("%s (%s per die)"), adjustedAmt, amt))
			case w.PerLevel:
				buf.WriteString(fmt.Sprintf(i18n.Text("%s (%s per level)"), adjustedAmt, amt))
			default:
				buf.WriteString(amt)
			}
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
			case feature.WeaponEffectiveSTBonus:
				buf.WriteString(i18n.Text("effective ST"))
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

// Hash writes this object's contents into the hasher.
func (w *WeaponBonus) Hash(h hash.Hash) {
	if w == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, w.Type)
	xhash.Bool(h, w.Percent)
	xhash.Num8(h, w.SelectionType)
	xhash.Num8(h, w.SwitchType)
	xhash.Bool(h, w.SwitchTypeValue)
	w.NameCriteria.Hash(h)
	w.SpecializationCriteria.Hash(h)
	w.RelativeLevelCriteria.Hash(h)
	w.UsageCriteria.Hash(h)
	w.TagsCriteria.Hash(h)
	xhash.Num64(h, w.Amount)
	xhash.Bool(h, w.PerLevel)
	xhash.Bool(h, w.PerDie)
}

// MarshalJSONTo implements json.MarshalerTo.
func (w *WeaponBonus) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, &w.WeaponBonusData)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (w *WeaponBonus) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var content struct {
		WeaponBonusData
		OldTagsCriteria criteria.Text `json:"category"`
		OldPerDie       bool          `json:"per_level"`
	}
	if err := json.UnmarshalDecode(dec, &content); err != nil {
		return err
	}
	w.WeaponBonusData = content.WeaponBonusData
	if !w.PerDie && content.OldPerDie {
		w.PerDie = true
	}
	if w.TagsCriteria.IsZero() && !content.OldTagsCriteria.IsZero() {
		w.TagsCriteria = content.OldTagsCriteria
	}
	return nil
}
