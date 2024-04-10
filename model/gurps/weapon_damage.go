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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

// WeaponDamageData holds the WeaponDamage data that is written to disk.
type WeaponDamageData struct {
	Type                      string       `json:"type"`
	StrengthType              stdmg.Option `json:"st,omitempty"`
	Base                      *dice.Dice   `json:"base,omitempty"`
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

// Clone creates a copy of this data.
func (w *WeaponDamage) Clone(owner *Weapon) *WeaponDamage {
	other := *w
	other.Owner = owner
	if other.Base != nil {
		d := *other.Base
		other.Base = &d
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
	}
	convertMods := false
	if w.Owner != nil {
		convertMods = SheetSettingsFor(w.Owner.Entity()).UseModifyingDicePlusAdds
	}
	if w.Base != nil {
		if base := w.Base.StringExtra(convertMods); base != "0" {
			if buffer.Len() != 0 && base[0] != '+' && base[0] != '-' {
				buffer.WriteByte('+')
			}
			buffer.WriteString(base)
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
	var tooltip xio.ByteBuffer
	w.ResolvedDamage(&tooltip)
	if tooltip.Len() == 0 {
		return noAdditionalModifiers()
	}
	return includesModifiersFrom() + tooltip.String()
}

// BaseDamageDice returns the base damage dice for this weapon (i.e. the dice before any bonuses are applied).
func (w *WeaponDamage) BaseDamageDice() *dice.Dice {
	if w.Owner == nil {
		return &dice.Dice{Sides: 6, Multiplier: 1}
	}
	pc := w.Owner.PC()
	if pc == nil {
		return &dice.Dice{Sides: 6, Multiplier: 1}
	}
	maxST := w.Owner.Strength.Resolve(w.Owner, nil).Min.Mul(fxp.Three)
	var st fxp.Int
	if w.Owner.Owner != nil {
		st = w.Owner.Owner.RatedStrength()
	}
	if st == 0 {
		st = pc.StrikingStrength()
	}
	if maxST > 0 && maxST < st {
		st = maxST
	}
	base := &dice.Dice{
		Sides:      6,
		Multiplier: 1,
	}
	if w.Base != nil {
		*base = *w.Base
	}
	t, tOK := w.Owner.Owner.(*Trait)
	if tOK && t.IsLeveled() {
		multiplyDice(fxp.As[int](t.Levels), base)
	}
	intST := fxp.As[int](st)
	switch w.StrengthType {
	case stdmg.Thrust:
		base = addDice(base, pc.ThrustFor(intST))
	case stdmg.LeveledThrust:
		thrust := pc.ThrustFor(intST)
		if tOK && t.IsLeveled() {
			multiplyDice(fxp.As[int](t.Levels), thrust)
		}
		base = addDice(base, thrust)
	case stdmg.Swing:
		base = addDice(base, pc.SwingFor(intST))
	case stdmg.LeveledSwing:
		swing := pc.SwingFor(intST)
		if tOK && t.IsLeveled() {
			multiplyDice(fxp.As[int](t.Levels), swing)
		}
		base = addDice(base, swing)
	default:
	}
	return base
}

// ResolvedDamage returns the damage, fully resolved for the user's sw or thr, if possible.
func (w *WeaponDamage) ResolvedDamage(tooltip *xio.ByteBuffer) string {
	base := w.BaseDamageDice()
	if base.Count == 0 && base.Modifier == 0 {
		return w.String()
	}
	pc := w.Owner.PC()
	adjustForPhoenixFlame := pc.SheetSettings.DamageProgression == progression.PhoenixFlameD3 && base.Sides == 3
	var percentDamageBonus, percentDRDivisorBonus fxp.Int
	armorDivisor := w.ArmorDivisor
	for _, bonus := range w.Owner.collectWeaponBonuses(base.Count, tooltip, feature.WeaponBonus, feature.WeaponDRDivisorBonus) {
		switch bonus.Type {
		case feature.WeaponBonus:
			bonus.WeaponLeveledAmount.DieCount = fxp.From(base.Count)
			amt := bonus.AdjustedAmountForWeapon(w.Owner)
			if bonus.Percent {
				percentDamageBonus += amt
			} else {
				if adjustForPhoenixFlame {
					if bonus.WeaponLeveledAmount.PerLevel {
						amt = amt.Div(fxp.Two)
					}
					if bonus.WeaponLeveledAmount.PerDie {
						amt = amt.Div(fxp.Two)
					}
				}
				base.Modifier += fxp.As[int](amt)
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
		amt := w.ModifierPerDie.Mul(fxp.From(base.Count))
		if adjustForPhoenixFlame {
			amt = amt.Div(fxp.Two)
		}
		base.Modifier += fxp.As[int](amt)
	}
	if percentDamageBonus != 0 {
		base = adjustDiceForPercentBonus(base, percentDamageBonus)
	}
	if percentDRDivisorBonus != 0 {
		armorDivisor = armorDivisor.Mul(percentDRDivisorBonus).Div(fxp.Hundred)
	}
	var buffer strings.Builder
	if base.Count != 0 || base.Modifier != 0 {
		buffer.WriteString(base.StringExtra(pc.SheetSettings.UseModifyingDicePlusAdds))
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
		if frag := w.Fragmentation.StringExtra(pc.SheetSettings.UseModifyingDicePlusAdds); frag != "0" {
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
		average := fxp.From(sides + 1).Div(fxp.Two)
		averageLeft := fxp.From(left.Count * (left.Sides + 1)).Div(fxp.Two).Mul(fxp.From(left.Multiplier))
		averageRight := fxp.From(right.Count * (right.Sides + 1)).Div(fxp.Two).Mul(fxp.From(right.Multiplier))
		averageBoth := averageLeft + averageRight
		return &dice.Dice{
			Count:      fxp.As[int](averageBoth.Div(average)),
			Sides:      sides,
			Modifier:   fxp.As[int](averageBoth.Mod(average).Round()) + left.Modifier + right.Modifier,
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
	count := fxp.From(d.Count)
	modifier := fxp.From(d.Modifier)
	averagePerDie := fxp.From(d.Sides + 1).Div(fxp.Two)
	average := averagePerDie.Mul(count) + modifier
	modifier = modifier.Mul(fxp.Hundred + percent).Div(fxp.Hundred)
	if average < 0 {
		count = count.Mul(fxp.Hundred + percent).Div(fxp.Hundred).Max(0)
	} else {
		average = average.Mul(fxp.Hundred+percent).Div(fxp.Hundred) - modifier
		count = average.Div(averagePerDie).Trunc().Max(0)
		modifier += (average - count.Mul(averagePerDie)).Round()
	}
	return &dice.Dice{
		Count:      fxp.As[int](count),
		Sides:      d.Sides,
		Modifier:   fxp.As[int](modifier),
		Multiplier: d.Multiplier,
	}
}
