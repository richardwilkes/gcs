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
	"encoding/json"
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// WeaponDamageData holds the WeaponDamage data that is written to disk.
type WeaponDamageData struct {
	Type                      string       `json:"type"`
	StrengthType              stdmg.Option `json:"st,omitempty"`
	Leveled                   bool         `json:"leveled,omitempty"`
	StrengthMultiplier        fxp.Int      `json:"st_mul,omitempty"`
	Base                      *dice.Dice   `json:"base,omitempty"`
	BaseLeveled               *dice.Dice   `json:"base_leveled,omitempty"`
	ArmorDivisor              fxp.Int      `json:"armor_divisor,omitempty"`
	Fragmentation             *dice.Dice   `json:"fragmentation,omitempty"`
	FragmentationArmorDivisor fxp.Int      `json:"fragmentation_armor_divisor,omitempty"`
	FragmentationType         string       `json:"fragmentation_type,omitempty"`
	ModifierPerDie            fxp.Int      `json:"modifier_per_die,omitempty"`
}

// WeaponDamage holds the damage information for a weapon.
type WeaponDamage struct {
	WeaponDamageData
	Owner *Weapon
}

// Hash writes this object's contents into the hasher.
func (w *WeaponDamage) Hash(h hash.Hash) {
	if w == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.StringWithLen(h, w.Type)
	xhash.Num8(h, w.StrengthType)
	xhash.Bool(h, w.Leveled)
	xhash.Num64(h, w.StrengthMultiplier)
	w.Base.Hash(h)
	w.BaseLeveled.Hash(h)
	xhash.Num64(h, w.ArmorDivisor)
	w.Fragmentation.Hash(h)
	xhash.Num64(h, w.FragmentationArmorDivisor)
	xhash.StringWithLen(h, w.FragmentationType)
	xhash.Num64(h, w.ModifierPerDie)
}

// Clone creates a copy of this data.
func (w *WeaponDamage) Clone(owner *Weapon) *WeaponDamage {
	other := *w
	other.Owner = owner
	if other.Base != nil {
		d := *other.Base
		other.Base = &d
	}
	if other.BaseLeveled != nil {
		d := *other.BaseLeveled
		other.BaseLeveled = &d
	}
	if other.Fragmentation != nil {
		d := *other.Fragmentation
		other.Fragmentation = &d
	}
	return &other
}

// UnmarshalJSON implements json.Unmarshaler.
func (w *WeaponDamage) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &w.WeaponDamageData); err != nil {
		return err
	}
	switch w.StrengthType {
	case stdmg.OldLeveledThrust:
		w.StrengthType = stdmg.Thrust
		w.Leveled = true
	case stdmg.OldLeveledSwing:
		w.StrengthType = stdmg.Swing
		w.Leveled = true
	}
	if w.StrengthMultiplier == 0 {
		w.StrengthMultiplier = fxp.One
	}
	if w.ArmorDivisor == 0 {
		w.ArmorDivisor = fxp.One
	}
	if w.Fragmentation != nil && w.FragmentationArmorDivisor == 0 {
		w.FragmentationArmorDivisor = fxp.One
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (w *WeaponDamage) MarshalJSON() ([]byte, error) {
	// A ST multiplier of 0 is not valid and 1 is very common, so suppress its output when 1.
	strengthMultiplier := w.StrengthMultiplier
	if strengthMultiplier == fxp.One {
		w.StrengthMultiplier = 0
	}
	// An armor divisor of 0 is not valid and 1 is very common, so suppress its output when 1.
	armorDivisor := w.ArmorDivisor
	if armorDivisor == fxp.One {
		w.ArmorDivisor = 0
	}
	fragArmorDivisor := w.FragmentationArmorDivisor
	if w.Fragmentation == nil {
		w.FragmentationArmorDivisor = 0
		w.FragmentationType = ""
	} else if w.FragmentationArmorDivisor == fxp.One {
		w.FragmentationArmorDivisor = 0
	}
	data, err := json.Marshal(&w.WeaponDamageData)
	w.StrengthMultiplier = strengthMultiplier
	w.ArmorDivisor = armorDivisor
	if w.Fragmentation != nil {
		w.FragmentationArmorDivisor = fragArmorDivisor
	}
	return data, err
}

func (w *WeaponDamage) String() string {
	var buffer strings.Builder
	if w.StrengthType != stdmg.None {
		buffer.WriteString(w.StrengthType.String())
		if w.Leveled {
			buffer.WriteString(i18n.Text(" (Leveled)"))
		}
	}
	convertMods := false
	if w.Owner != nil {
		convertMods = SheetSettingsFor(EntityFromNode(w.Owner)).UseModifyingDicePlusAdds
	}
	if w.Base != nil {
		if base := w.Base.StringExtra(convertMods); base != "0" {
			if buffer.Len() != 0 && base[0] != '+' && base[0] != '-' {
				buffer.WriteByte('+')
			}
			buffer.WriteString(base)
		}
	}
	if w.BaseLeveled != nil {
		if base := w.BaseLeveled.StringExtra(convertMods); base != "0" {
			if buffer.Len() != 0 && base[0] != '+' && base[0] != '-' {
				buffer.WriteByte('+')
			}
			buffer.WriteString(base)
			buffer.WriteString(i18n.Text(" per level "))
		}
	}
	if w.ArmorDivisor != fxp.One {
		buffer.WriteByte('(')
		buffer.WriteString(w.ArmorDivisor.String())
		buffer.WriteByte(')')
	}
	if w.ModifierPerDie != 0 {
		if buffer.Len() != 0 {
			buffer.WriteByte(' ')
		}
		buffer.WriteByte('(')
		buffer.WriteString(w.ModifierPerDie.StringWithSign())
		buffer.WriteString(i18n.Text(" per die)"))
	}
	if t := strings.TrimSpace(w.Type); t != "" {
		buffer.WriteByte(' ')
		buffer.WriteString(t)
	}
	if w.Fragmentation != nil {
		if frag := w.Fragmentation.StringExtra(convertMods); frag != "0" {
			buffer.WriteString(" [")
			buffer.WriteString(frag)
			if w.FragmentationArmorDivisor != fxp.One {
				buffer.WriteByte('(')
				buffer.WriteString(w.FragmentationArmorDivisor.String())
				buffer.WriteByte(')')
			}
			buffer.WriteByte(' ')
			buffer.WriteString(w.FragmentationType)
			buffer.WriteByte(']')
		}
	}
	return strings.TrimSpace(buffer.String())
}

// DamageTooltip returns a formatted tooltip for the damage.
func (w *WeaponDamage) DamageTooltip() string {
	var tooltip xbytes.InsertBuffer
	w.ResolvedDamage(&tooltip)
	if tooltip.Len() == 0 {
		return NoAdditionalModifiers()
	}
	return IncludesModifiersFrom() + tooltip.String()
}

// BaseDamageDice returns the base damage dice for this weapon (i.e. the dice before any bonuses are applied).
func (w *WeaponDamage) BaseDamageDice() *dice.Dice {
	if w.Owner == nil {
		return &dice.Dice{Sides: 6, Multiplier: 1}
	}
	entity := w.Owner.Entity()
	if entity == nil {
		return &dice.Dice{Sides: 6, Multiplier: 1}
	}
	maxST := w.Owner.Strength.Resolve(w.Owner, nil).Min.Mul(fxp.Three)
	var st fxp.Int
	if w.Owner.Owner != nil {
		st = w.Owner.Owner.RatedStrength()
	}
	if st == 0 {
		switch w.StrengthType {
		case stdmg.Thrust, stdmg.Swing:
			st = entity.StrikingStrength()
		case stdmg.LiftingThrust, stdmg.LiftingSwing:
			st = entity.LiftingStrength()
		case stdmg.TelekineticThrust, stdmg.TelekineticSwing:
			st = entity.TelekineticStrength()
		case stdmg.IQThrust, stdmg.IQSwing:
			st = entity.ResolveAttributeCurrent(IntelligenceID).Max(0).Floor()
		default:
			st = entity.ResolveAttributeCurrent(StrengthID).Max(0).Floor()
		}
	}
	var percentMin fxp.Int
	for _, bonus := range w.Owner.collectWeaponBonuses(1, nil, feature.WeaponEffectiveSTBonus) {
		amt := bonus.AdjustedAmountForWeapon(w.Owner)
		if bonus.Percent {
			percentMin += amt
		} else {
			st += amt
		}
	}
	if percentMin != 0 {
		st += st.Mul(percentMin).Div(fxp.Hundred).Floor()
	}
	if st < 0 {
		st = 0
	}
	if maxST > 0 && maxST < st {
		st = maxST
	}
	if w.StrengthMultiplier > 0 { // Just in case it somehow got set to 0
		st = st.Mul(w.StrengthMultiplier)
	}
	base := &dice.Dice{
		Sides:      6,
		Multiplier: 1,
	}
	if w.Base != nil {
		*base = *w.Base
	}
	levels := 0
	switch t := w.Owner.Owner.(type) {
	case *Trait:
		if t.IsLeveled() {
			levels = fxp.AsInteger[int](t.CurrentLevel())
		}
	case *Equipment:
		if t.IsLeveled() {
			levels = fxp.AsInteger[int](t.Level)
		}
	}
	if levels > 0 && w.BaseLeveled != nil {
		leveled := *w.BaseLeveled
		multiplyDice(levels, &leveled)
		base = addDice(base, &leveled)
	}
	intST := fxp.AsInteger[int](st)
	var stDamage *dice.Dice
	switch w.StrengthType {
	case stdmg.Thrust, stdmg.LiftingThrust, stdmg.TelekineticThrust, stdmg.IQThrust:
		stDamage = entity.ThrustFor(intST)
	case stdmg.Swing, stdmg.LiftingSwing, stdmg.TelekineticSwing, stdmg.IQSwing:
		stDamage = entity.SwingFor(intST)
	default:
		return base
	}
	if w.Leveled && levels >= 0 {
		multiplyDice(levels, stDamage)
	}
	base = addDice(base, stDamage)
	return base
}

// ResolvedDamage returns the damage, fully resolved for the user's sw or thr, if possible.
func (w *WeaponDamage) ResolvedDamage(tooltip *xbytes.InsertBuffer) string {
	if w.Owner == nil {
		return w.String()
	}
	entity := w.Owner.Entity()
	if entity == nil {
		return w.String()
	}
	base := w.BaseDamageDice()
	adjustForPhoenixFlame := entity.SheetSettings.DamageProgression == progression.PhoenixFlameD3 && base.Sides == 3
	var percentDamageBonus, percentDRDivisorBonus fxp.Int
	armorDivisor := w.ArmorDivisor
	for _, bonus := range w.Owner.collectWeaponBonuses(base.Count, tooltip, feature.WeaponBonus, feature.WeaponDRDivisorBonus) {
		switch bonus.Type {
		case feature.WeaponBonus:
			bonus.DieCount = fxp.FromInteger(base.Count)
			amt := bonus.AdjustedAmountForWeapon(w.Owner)
			if bonus.Percent {
				percentDamageBonus += amt
			} else {
				if adjustForPhoenixFlame {
					if bonus.PerLevel {
						amt = amt.Div(fxp.Two)
					}
					if bonus.PerDie {
						amt = amt.Div(fxp.Two)
					}
				}
				base.Modifier += fxp.AsInteger[int](amt)
			}
		case feature.WeaponDRDivisorBonus:
			amt := bonus.AdjustedAmountForWeapon(w.Owner)
			if bonus.Percent {
				percentDRDivisorBonus += amt
			} else {
				armorDivisor += amt
			}
		default:
		}
	}
	if w.ModifierPerDie != 0 {
		amt := w.ModifierPerDie.Mul(fxp.FromInteger(base.Count))
		if adjustForPhoenixFlame {
			amt = amt.Div(fxp.Two)
		}
		base.Modifier += fxp.AsInteger[int](amt)
	}
	if percentDamageBonus != 0 {
		base = adjustDiceForPercentBonus(base, percentDamageBonus)
	}
	if percentDRDivisorBonus != 0 {
		armorDivisor = armorDivisor.Mul(percentDRDivisorBonus).Div(fxp.Hundred)
	}
	var buffer strings.Builder
	if base.Count != 0 || base.Modifier != 0 {
		buffer.WriteString(base.StringExtra(entity.SheetSettings.UseModifyingDicePlusAdds))
	}
	if armorDivisor != fxp.One {
		buffer.WriteByte('(')
		buffer.WriteString(armorDivisor.String())
		buffer.WriteByte(')')
	}
	t := strings.TrimSpace(w.Type)
	if t != "" {
		if buffer.Len() != 0 {
			buffer.WriteByte(' ')
		}
		buffer.WriteString(t)
	}
	if w.Fragmentation != nil {
		if frag := w.Fragmentation.StringExtra(entity.SheetSettings.UseModifyingDicePlusAdds); frag != "0" {
			if buffer.Len() != 0 {
				buffer.WriteByte(' ')
			}
			buffer.WriteByte('[')
			buffer.WriteString(frag)
			if w.FragmentationArmorDivisor != fxp.One {
				buffer.WriteByte('(')
				buffer.WriteString(w.FragmentationArmorDivisor.String())
				buffer.WriteByte(')')
			}
			t = strings.TrimSpace(w.FragmentationType)
			if t != "" {
				buffer.WriteByte(' ')
				buffer.WriteString(t)
			}
			buffer.WriteByte(']')
		}
	}
	return buffer.String()
}

func multiplyDice(multiplier int, d *dice.Dice) {
	d.Count *= multiplier
	d.Modifier *= multiplier
	if d.Multiplier != 1 {
		d.Multiplier *= multiplier
	}
}

func addDice(left, right *dice.Dice) *dice.Dice {
	if left.Sides > 1 && right.Sides > 1 && left.Sides != right.Sides {
		sides := min(left.Sides, right.Sides)
		average := fxp.FromInteger(sides + 1).Div(fxp.Two)
		averageLeft := fxp.FromInteger(left.Count * (left.Sides + 1)).Div(fxp.Two).Mul(fxp.FromInteger(left.Multiplier))
		averageRight := fxp.FromInteger(right.Count * (right.Sides + 1)).Div(fxp.Two).Mul(fxp.FromInteger(right.Multiplier))
		averageBoth := averageLeft + averageRight
		return &dice.Dice{
			Count:      fxp.AsInteger[int](averageBoth.Div(average)),
			Sides:      sides,
			Modifier:   fxp.AsInteger[int](averageBoth.Mod(average).Round()) + left.Modifier + right.Modifier,
			Multiplier: 1,
		}
	}
	return &dice.Dice{
		Count:      left.Count + right.Count,
		Sides:      max(left.Sides, right.Sides),
		Modifier:   left.Modifier + right.Modifier,
		Multiplier: left.Multiplier + right.Multiplier - 1,
	}
}

func adjustDiceForPercentBonus(d *dice.Dice, percent fxp.Int) *dice.Dice {
	count := fxp.FromInteger(d.Count)
	modifier := fxp.FromInteger(d.Modifier)
	averagePerDie := fxp.FromInteger(d.Sides + 1).Div(fxp.Two)
	average := averagePerDie.Mul(count) + modifier
	modifier = modifier.Mul(fxp.Hundred + percent).Div(fxp.Hundred)
	if average < 0 {
		count = count.Mul(fxp.Hundred + percent).Div(fxp.Hundred).Max(0)
	} else {
		average = average.Mul(fxp.Hundred+percent).Div(fxp.Hundred) - modifier
		count = average.Div(averagePerDie).Floor().Max(0)
		modifier += (average - count.Mul(averagePerDie)).Round()
	}
	return &dice.Dice{
		Count:      fxp.AsInteger[int](count),
		Sides:      d.Sides,
		Modifier:   fxp.AsInteger[int](modifier),
		Multiplier: d.Multiplier,
	}
}
