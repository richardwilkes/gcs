// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
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

// addWeaponPercentBonus returns value increased by the given percentage of itself, with the increment floored (i.e.
// rounded toward negative infinity, so a -1.5 increment becomes -2). A percent of 0 returns value unchanged. This is
// the standard way weapon stat percentage bonuses are applied.
func addWeaponPercentBonus(value, percent fxp.Int) fxp.Int {
	if percent == 0 {
		return value
	}
	return value + value.Mul(percent).Div(fxp.Hundred).Floor()
}

// weaponAdjustment accumulates the weapon bonuses of one feature type, keeping the flat amounts apart from the
// percentages so that they can be applied in the standard order: the flat amounts first, then the percentage of the
// adjusted value. The zero value is no adjustment.
type weaponAdjustment struct {
	flat    fxp.Int
	percent fxp.Int
}

// applyTo returns value with this adjustment applied.
func (a weaponAdjustment) applyTo(value fxp.Int) fxp.Int {
	return addWeaponPercentBonus(value+a.flat, a.percent)
}

// weaponAdjustments collects the weapon bonuses of the given feature types in a single pass and accumulates them into
// one weaponAdjustment per type. A type with no matching bonus is absent from the map, which reads as the zero
// weaponAdjustment and so applies no change.
func (w *Weapon) weaponAdjustments(dieCount dieCountFunc, tooltip *xbytes.InsertBuffer, types ...feature.Type) map[feature.Type]weaponAdjustment {
	adjustments := make(map[feature.Type]weaponAdjustment, len(types))
	for _, bonus := range w.collectWeaponBonuses(dieCount, tooltip, types...) {
		amt := bonus.AdjustedAmountForWeapon(w)
		adj := adjustments[bonus.Type]
		if bonus.Percent {
			adj.percent += amt
		} else {
			adj.flat += amt
		}
		adjustments[bonus.Type] = adj
	}
	return adjustments
}

// weaponAdjustment is weaponAdjustments for a single feature type.
func (w *Weapon) weaponAdjustment(dieCount dieCountFunc, tooltip *xbytes.InsertBuffer, featureType feature.Type) weaponAdjustment {
	return w.weaponAdjustments(dieCount, tooltip, featureType)[featureType]
}

// WeaponBonus holds the data for an adjustment to weapon stats.
type WeaponBonus struct {
	WeaponBonusData
}

// WeaponBonusData holds the persisted data for an adjustment to weapon stats.
type WeaponBonusData struct { //nolint:govet // The field alignment here is poor, but kept to reduce diffs in the data
	Type feature.Type `json:"type"`
	FeatureSwitch
	Percent                        bool            `json:"percent,omitzero"`
	SelectionType                  wsel.Type       `json:"selection_type"`
	SwitchType                     wswitch.Type    `json:"switch_type,omitzero"`
	SwitchTypeValue                bool            `json:"switch_type_value,omitzero"`
	NameCriteria                   criteria.Text   `json:"name,omitzero"`
	SpecializationCriteria         criteria.Text   `json:"specialization,omitzero"`
	OptionalSpecializationCriteria criteria.Text   `json:"optional_specialization,omitzero"`
	RelativeLevelCriteria          criteria.Number `json:"level,omitzero"`
	UsageCriteria                  criteria.Text   `json:"usage,omitzero"`
	TagsCriteria                   criteria.Text   `json:"tags,omitzero"`
	LeveledOwner                   LeveledOwner    `json:"-"`
	DieCount                       fxp.Int         `json:"-"`
	Amount                         fxp.Int         `json:"amount"`
	PerLevel                       bool            `json:"leveled,omitzero"`
	PerDie                         bool            `json:"per_die,omitzero"`
	BonusOwner                     `json:"-"`
}

// NewWeaponBonus creates a new weapon bonus of the given type, which must satisfy feature.Type.IsWeaponBonus.
func NewWeaponBonus(t feature.Type) *WeaponBonus {
	var w WeaponBonus
	w.Type = t
	w.SelectionType = wsel.WithRequiredSkill
	w.NameCriteria.Compare = criteria.IsText
	w.SpecializationCriteria.Compare = criteria.AnyText
	w.OptionalSpecializationCriteria.Compare = criteria.AnyText
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
	return clonePtr(w)
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

// AdjustedAmount returns the amount, adjusted for the die count and level when the bonus is per-die or per-level.
func (w *WeaponBonus) AdjustedAmount() fxp.Int {
	return w.adjustedAmount(w.DieCount, w.LeveledOwner)
}

// resolveDieCount returns the die count to use for this bonus, only asking the supplier for it when the bonus actually
// scales per die, since resolving it means computing the weapon's base damage dice.
func (w *WeaponBonus) resolveDieCount(dieCount dieCountFunc) fxp.Int {
	if !w.PerDie {
		return fxp.One
	}
	return dieCount()
}

// adjustedAmount returns the amount adjusted for the given die count and leveled owner. Taking these as parameters
// rather than reading the DieCount/LeveledOwner scratch fields lets callers compute an amount without mutating the
// shared bonus, which is not safe when the bonus may be read concurrently.
func (w *WeaponBonus) adjustedAmount(dieCount fxp.Int, leveledOwner LeveledOwner) fxp.Int {
	amt := w.Amount
	if w.PerDie {
		if dieCount < 0 {
			return 0
		}
		amt = amt.Mul(dieCount)
	}
	if w.PerLevel {
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
	nameable.Extract(
		m, existing,
		w.SpecializationCriteria.Qualifier,
		w.OptionalSpecializationCriteria.Qualifier,
	)
	if w.SelectionType != wsel.ThisWeapon {
		nameable.Extract(
			m, existing,
			w.NameCriteria.Qualifier,
			w.UsageCriteria.Qualifier,
			w.TagsCriteria.Qualifier,
		)
	}
}

// SetLeveledOwner implements Bonus.
func (w *WeaponBonus) SetLeveledOwner(owner LeveledOwner) {
	w.LeveledOwner = owner
}

// AddToTooltip implements Bonus.
func (w *WeaponBonus) AddToTooltip(buffer *xbytes.InsertBuffer) {
	w.addToTooltip(w.AdjustedAmount(), buffer)
}

// addToTooltip writes the tooltip using a pre-computed adjusted amount, so it has no dependence on the mutable
// DieCount/LeveledOwner scratch fields.
func (w *WeaponBonus) addToTooltip(adjustedAmount fxp.Int, buffer *xbytes.InsertBuffer) {
	if buffer != nil {
		var buf strings.Builder
		buf.WriteByte('\n')
		buf.WriteString(w.parentName())
		buf.WriteString(" [")
		if w.Type == feature.WeaponSwitch {
			fmt.Fprintf(&buf, i18n.Text("%v set to %v"), w.SwitchType, w.SwitchTypeValue)
		} else {
			amt := w.Amount.StringWithSign()
			adjustedAmt := adjustedAmount.StringWithSign()
			if w.Percent {
				amt += "%"
				adjustedAmt += "%"
			}
			switch {
			case w.PerDie && w.PerLevel:
				fmt.Fprintf(&buf, i18n.Text("%s (%s per die, per level)"), adjustedAmt, amt)
			case w.PerDie:
				fmt.Fprintf(&buf, i18n.Text("%s (%s per die)"), adjustedAmt, amt)
			case w.PerLevel:
				fmt.Fprintf(&buf, i18n.Text("%s (%s per level)"), adjustedAmt, amt)
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
				buf.WriteString(i18n.Text("armor divisor"))
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
	xhash.Bool(h, w.Switchable)
	xhash.Bool(h, w.Percent)
	xhash.Num8(h, w.SelectionType)
	xhash.Num8(h, w.SwitchType)
	xhash.Bool(h, w.SwitchTypeValue)
	w.NameCriteria.Hash(h)
	w.SpecializationCriteria.Hash(h)
	w.OptionalSpecializationCriteria.Hash(h)
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
		OldPerDie bool `json:"per_level"`
	}
	if err := unmarshalWithLegacyTags(dec, &content, &content.TagsCriteria); err != nil {
		return err
	}
	w.WeaponBonusData = content.WeaponBonusData
	if !w.PerDie && content.OldPerDie {
		w.PerDie = true
	}
	return nil
}
