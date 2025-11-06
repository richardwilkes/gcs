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
	StrengthType              stdmg.Option `json:"st,omitzero"`
	Leveled                   bool         `json:"leveled,omitzero"`
	StrengthMultiplier        fxp.Int      `json:"st_mul,omitzero"`
	Base                      string       `json:"base,omitzero"`
	BaseLeveled               string       `json:"base_leveled,omitzero"`
	ArmorDivisor              fxp.Int      `json:"armor_divisor,omitzero"`
	Fragmentation             string       `json:"fragmentation,omitzero"`
	FragmentationArmorDivisor fxp.Int      `json:"fragmentation_armor_divisor,omitzero"`
	FragmentationType         string       `json:"fragmentation_type,omitzero"`
	ModifierPerDie            fxp.Int      `json:"modifier_per_die,omitzero"`
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
	xhash.StringWithLen(h, w.Base)
	xhash.StringWithLen(h, w.BaseLeveled)
	xhash.Num64(h, w.ArmorDivisor)
	xhash.StringWithLen(h, w.Fragmentation)
	xhash.Num64(h, w.FragmentationArmorDivisor)
	xhash.StringWithLen(h, w.FragmentationType)
	xhash.Num64(h, w.ModifierPerDie)
}

// Clone creates a copy of this data.
func (w *WeaponDamage) Clone(owner *Weapon) *WeaponDamage {
	other := *w
	other.Owner = owner
	return &other
}

// MarshalJSONTo implements json.MarshalerTo.
func (w *WeaponDamage) MarshalJSONTo(enc *jsontext.Encoder) error {
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
	w.Fragmentation = strings.TrimSpace(w.Fragmentation)
	if w.Fragmentation == "" {
		w.FragmentationArmorDivisor = 0
		w.FragmentationType = ""
	} else if w.FragmentationArmorDivisor == fxp.One {
		w.FragmentationArmorDivisor = 0
	}
	// An armor divisor of 0 is not valid and 1 is very common, so suppress its output when 1.
	fragArmorDivisor := w.FragmentationArmorDivisor
	if fragArmorDivisor == fxp.One {
		fragArmorDivisor = 0
	}
	err := json.MarshalEncode(enc, &w.WeaponDamageData)
	w.StrengthMultiplier = strengthMultiplier
	w.ArmorDivisor = armorDivisor
	w.FragmentationArmorDivisor = fragArmorDivisor
	return err
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (w *WeaponDamage) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if err := json.UnmarshalDecode(dec, &w.WeaponDamageData); err != nil {
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
	w.Fragmentation = strings.TrimSpace(w.Fragmentation)
	if w.Fragmentation != "" && w.FragmentationArmorDivisor == 0 {
		w.FragmentationArmorDivisor = fxp.One
	}
	return nil
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
	if w.Base != "" {
		buffer.WriteString(w.formatDiceWithSub(w.Base, convertMods, buffer.Len() != 0))
	}
	if w.BaseLeveled != "" {
		if s := w.formatDiceWithSub(w.BaseLeveled, convertMods, buffer.Len() != 0); s != "" {
			buffer.WriteString(s)
			buffer.WriteByte(' ')
			buffer.WriteString(i18n.Text("per level"))
			buffer.WriteByte(' ')
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
	if w.Fragmentation != "" {
		if s := w.formatDiceWithSub(w.Fragmentation, convertMods, false); s != "" {
			buffer.WriteString(" [")
			buffer.WriteString(s)
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
	baseSub := false
	if w.Base != "" {
		base, baseSub = w.resolveDiceSpec(w.Base)
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
	if levels > 0 && w.BaseLeveled != "" {
		leveled, leveledSub := w.resolveDiceSpec(w.BaseLeveled)
		multiplyDice(levels, leveled)
		base, baseSub = addDice(base, leveled, baseSub, leveledSub)
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
	base, baseSub = addDice(base, stDamage, baseSub, false)
	if baseSub {
		// Still negative, so return 0 damage.
		base = &dice.Dice{Sides: 6, Multiplier: 1}
	}
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
	if w.Fragmentation != "" {
		d, sub := w.resolveDiceSpec(w.Fragmentation)
		if sub {
			// Negative fragmentation doesn't make sense, so ignore it.
			d = &dice.Dice{Sides: 6, Multiplier: 1}
		}
		if frag := d.StringExtra(entity.SheetSettings.UseModifyingDicePlusAdds); frag != "0" {
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

func addDice(left, right *dice.Dice, leftSub, rightSub bool) (d *dice.Dice, sub bool) {
	if leftSub {
		left.Count = -left.Count
	}
	if rightSub {
		right.Count = -right.Count
	}
	defer func() {
		if leftSub {
			left.Count = -left.Count
		}
		if rightSub {
			right.Count = -right.Count
		}
	}()
	if left.Sides > 1 && right.Sides > 1 && left.Sides != right.Sides {
		sides := min(left.Sides, right.Sides)
		average := fxp.FromInteger(sides + 1).Div(fxp.Two)
		averageLeft := fxp.FromInteger(left.Count * (left.Sides + 1)).Div(fxp.Two).Mul(fxp.FromInteger(left.Multiplier))
		averageRight := fxp.FromInteger(right.Count * (right.Sides + 1)).Div(fxp.Two).Mul(fxp.FromInteger(right.Multiplier))
		averageBoth := averageLeft + averageRight
		d = &dice.Dice{
			Count:      fxp.AsInteger[int](averageBoth.Div(average)),
			Sides:      sides,
			Modifier:   fxp.AsInteger[int](averageBoth.Mod(average).Round()) + left.Modifier + right.Modifier,
			Multiplier: 1,
		}
	} else {
		d = &dice.Dice{
			Count:      left.Count + right.Count,
			Sides:      max(left.Sides, right.Sides),
			Modifier:   left.Modifier + right.Modifier,
			Multiplier: left.Multiplier + right.Multiplier - 1,
		}
	}
	if d.Count < 0 {
		d.Count = -d.Count
		sub = true
	}
	return d, sub
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

func (w *WeaponDamage) formatDiceWithSub(s string, convertMods, following bool) string {
	d, sub := w.resolveDiceSpec(s)
	if base := d.StringExtra(convertMods); base != "0" {
		if sub {
			return "-" + base
		}
		if following && base[0] != '+' && base[0] != '-' {
			return "+" + base
		}
		return base
	}
	return ""
}

func (w *WeaponDamage) resolveDiceSpec(s string) (d *dice.Dice, sub bool) {
	if d, sub = parsePotentialDiceSpec(s); d != nil {
		return d, sub
	}
	value := ResolveScript(w.Owner.Entity(), deferredNewScriptWeapon(w.Owner), s)
	value = strings.TrimPrefix(strings.TrimSpace(value), "+")
	if sub = isDiceSubtraction(value); sub {
		value = value[1:]
	}
	return dice.New(value), sub
}

func parsePotentialDiceSpec(s string) (d *dice.Dice, sub bool) {
	spec := strings.TrimLeft(strings.TrimSpace(s), "+")
	if !strings.Contains(spec, " ") {
		if sub = isDiceSubtraction(spec); sub {
			spec = spec[1:]
		}
		d = dice.New(spec)
		if spec == "0" || spec == "-0" {
			return d, false
		}
		empty := dice.Dice{}
		empty.Normalize()
		if *d != empty {
			return d, sub
		}
	}
	return nil, false
}

func isDiceSubtraction(s string) bool {
	if strings.HasPrefix(s, "-") {
		i := 1
		for i < len(s) && (s[i] >= '0' && s[i] <= '9') {
			i++
		}
		if i < len(s) && (s[i] == 'd' || s[i] == 'D') {
			return true
		}
	}
	return false
}
