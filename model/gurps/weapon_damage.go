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
	"cmp"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"hash"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
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
	// Marshal a copy so that suppressing "default" values for output doesn't mutate the receiver.
	data := w.WeaponDamageData
	// A ST multiplier of 0 is not valid and 1 is very common, so suppress its output when 1.
	if data.StrengthMultiplier == fxp.One {
		data.StrengthMultiplier = 0
	}
	// An armor divisor of 0 is not valid and 1 is very common, so suppress its output when 1.
	if data.ArmorDivisor == fxp.One {
		data.ArmorDivisor = 0
	}
	data.Fragmentation = strings.TrimSpace(data.Fragmentation)
	if data.Fragmentation == "" {
		data.FragmentationArmorDivisor = 0
		data.FragmentationType = ""
	} else if data.FragmentationArmorDivisor == fxp.One {
		data.FragmentationArmorDivisor = 0
	}
	return json.MarshalEncode(enc, &data)
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

// resolvedStrengthType returns the strength basis after applying any matching selector override. Passing a tooltip
// records the contest; BaseDamageDice resolves with a nil tooltip for the computed dice, so the entry is written once.
func (w *WeaponDamage) resolvedStrengthType(tooltip *xbytes.InsertBuffer) stdmg.Option {
	if w.Owner == nil {
		return w.StrengthType
	}
	return stdmg.ExtractOption(w.Owner.ResolveSelector(selector.WeaponDamageStrengthBasis, w.StrengthType.Key(),
		tooltip))
}

// resolvedDamageString returns a string-valued damage field after applying any matching selector override.
func (w *WeaponDamage) resolvedDamageString(field selector.Field, base string, tooltip *xbytes.InsertBuffer) string {
	if w.Owner == nil {
		return base
	}
	return w.Owner.ResolveSelector(field, base, tooltip)
}

// resolvedDamageNumeric returns a numeric damage field after applying any matching selector override. The override
// value is carried as a string; an unparsable winner falls back to the base value.
func (w *WeaponDamage) resolvedDamageNumeric(field selector.Field, base fxp.Int, tooltip *xbytes.InsertBuffer) fxp.Int {
	if w.Owner == nil {
		return base
	}
	if v, err := fxp.FromString(w.Owner.ResolveSelector(field, base.String(), tooltip)); err == nil {
		return v
	}
	return base
}

// surfaceBaseDamageOverrides records, in the tooltip, the contest for every override that BaseDamageDice consumes with
// a nil tooltip, so those entries are written exactly once here rather than on every internal call.
func (w *WeaponDamage) surfaceBaseDamageOverrides(tooltip *xbytes.InsertBuffer) {
	w.resolvedStrengthType(tooltip)
	w.resolvedDamageString(selector.WeaponBaseDamageDice, w.Base, tooltip)
	w.resolvedDamageString(selector.WeaponBaseDamageDicePerLevel, w.BaseLeveled, tooltip)
	w.resolvedDamageNumeric(selector.WeaponDamageStrengthMultiplier, w.StrengthMultiplier, tooltip)
}

// BaseDamageDice returns the base damage dice for this weapon (i.e. the dice before any bonuses are applied).
func (w *WeaponDamage) BaseDamageDice() dice.Dice {
	if w.Owner == nil {
		return dice.Dice{Sides: 6, Multiplier: 1}
	}
	entity := w.Owner.Entity()
	if entity == nil {
		return dice.Dice{Sides: 6, Multiplier: 1}
	}
	strengthType := w.resolvedStrengthType(nil)
	st := w.Owner.effectiveStrength(func(entity *Entity) fxp.Int {
		switch strengthType {
		case stdmg.Thrust, stdmg.Swing:
			return entity.StrikingStrength()
		case stdmg.LiftingThrust, stdmg.LiftingSwing:
			return entity.LiftingStrength()
		case stdmg.TelekineticThrust, stdmg.TelekineticSwing:
			return entity.TelekineticStrength()
		case stdmg.IQThrust, stdmg.IQSwing:
			return entity.ResolveAttributeCurrent(IntelligenceID).Max(0).Floor()
		default:
			return entity.ResolveAttributeCurrent(StrengthID).Max(0).Floor()
		}
	}, nil)
	if strengthMultiplier := w.resolvedDamageNumeric(selector.WeaponDamageStrengthMultiplier, w.StrengthMultiplier, nil); strengthMultiplier > 0 { // Just in case it somehow got set to 0
		st = st.Mul(strengthMultiplier)
	}
	base := dice.Dice{
		Sides:      6,
		Multiplier: 1,
	}
	baseSub := false
	if baseSpec := w.resolvedDamageString(selector.WeaponBaseDamageDice, w.Base, nil); baseSpec != "" {
		base, baseSub = w.resolveDiceSpec(baseSpec)
	}
	levels := 0
	switch t := w.Owner.Owner.(type) {
	case *Trait:
		if t.IsLeveled() {
			levels = t.CurrentLevel().AsInteger[int]()
		}
	case *Equipment:
		if t.IsLeveled() {
			levels = t.Level.AsInteger[int]()
		}
	}
	if baseLeveledSpec := w.resolvedDamageString(selector.WeaponBaseDamageDicePerLevel, w.BaseLeveled, nil); levels > 0 && baseLeveledSpec != "" {
		leveled, leveledSub := w.resolveDiceSpec(baseLeveledSpec)
		leveled = multiplyDice(levels, leveled)
		base, baseSub = addDice(base, leveled, baseSub, leveledSub)
	}
	intST := st.AsInteger[int]()
	var stDamage dice.Dice
	switch strengthType {
	case stdmg.Thrust, stdmg.LiftingThrust, stdmg.TelekineticThrust, stdmg.IQThrust:
		stDamage = entity.ThrustFor(intST)
	case stdmg.Swing, stdmg.LiftingSwing, stdmg.TelekineticSwing, stdmg.IQSwing:
		stDamage = entity.SwingFor(intST)
	default:
		return base
	}
	if w.Leveled && levels >= 0 {
		stDamage = multiplyDice(levels, stDamage)
	}
	base, baseSub = addDice(base, stDamage, baseSub, false)
	if baseSub {
		// Still negative, so return 0 damage.
		base = dice.Dice{Sides: 6, Multiplier: 1}
	}
	return base
}

// ResolvedWeaponDamage is a weapon's damage after the wielder's ST, the bonuses that apply to it and any selector
// overrides have all been folded in. ResolvedDamage formats it, but callers that need the pieces themselves -- the
// explosion calculator, which wants the dice and the fragmentation dice -- work from this rather than parsing the
// formatted string back apart.
type ResolvedWeaponDamage struct {
	Type                      string    // The damage type, e.g. "cr ex".
	FragmentationType         string    // The damage type of the fragments, e.g. "cut".
	Dice                      dice.Dice // The damage dice.
	Fragmentation             dice.Dice // The fragmentation dice; only meaningful when HasFragmentation is true.
	ArmorDivisor              fxp.Int   // The armor divisor; 1 means none.
	FragmentationArmorDivisor fxp.Int   // The armor divisor for the fragments; 1 means none.
	HasFragmentation          bool      // Whether a fragmentation bracket would be printed for this damage.
	// useModifyingDicePlusAdds is captured from the entity's sheet settings so that String() cannot format the dice
	// differently than the entity the damage was resolved against would.
	useModifyingDicePlusAdds bool
}

// ResolveDamage returns the damage, fully resolved for the user's sw or thr, or nil if there is no owning weapon or no
// entity to resolve it against. If 'tooltip' isn't nil, it is updated with the details of every bonus and selector
// override that was consulted.
func (w *WeaponDamage) ResolveDamage(tooltip *xbytes.InsertBuffer) *ResolvedWeaponDamage {
	if w.Owner == nil {
		return nil
	}
	entity := w.Owner.Entity()
	if entity == nil {
		return nil
	}
	base := w.BaseDamageDice()
	w.surfaceBaseDamageOverrides(tooltip)
	adjustForPhoenixFlame := entity.SheetSettings.DamageProgression == progression.PhoenixFlameD3 && base.Sides == 3
	var percentDamageBonus, percentDRDivisorBonus fxp.Int
	var bonusDice []BonusDice // The dice contributed by the damage bonuses
	armorDivisor := w.resolvedDamageNumeric(selector.WeaponArmorDivisor, w.ArmorDivisor, tooltip)
	dieCount := fxp.FromInteger(base.Count) // Already resolved here, so hand it over rather than resolving it again
	for _, bonus := range w.Owner.collectWeaponBonuses(func() fxp.Int { return dieCount }, tooltip, feature.WeaponBonus,
		feature.WeaponDRDivisorBonus) {
		switch bonus.Type {
		case feature.WeaponBonus:
			d, amt := bonus.AdjustedForWeapon(w.Owner)
			if bonus.Percent {
				percentDamageBonus += amt
			} else {
				if adjustForPhoenixFlame {
					amt = halveForPhoenixFlame(bonus, amt)
					d.Modifier = halveForPhoenixFlame(bonus, fxp.FromInteger(d.Modifier)).AsInteger[int]()
				}
				base.Modifier += amt.AsInteger[int]()
				if !d.IsZero() {
					bonusDice = append(bonusDice, d)
				}
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
	if modifierPerDie := w.resolvedDamageNumeric(selector.WeaponDamagePerDieModifier, w.ModifierPerDie, tooltip); modifierPerDie != 0 {
		amt := modifierPerDie.Mul(fxp.FromInteger(base.Count))
		if adjustForPhoenixFlame {
			amt = amt.Div(fxp.Two)
		}
		base.Modifier += amt.AsInteger[int]()
	}
	base = addBonusDice(base, bonusDice)
	if percentDamageBonus != 0 {
		base = adjustDiceForPercentBonus(base, percentDamageBonus)
	}
	if percentDRDivisorBonus != 0 {
		armorDivisor += armorDivisor.Mul(percentDRDivisorBonus).Div(fxp.Hundred)
	}
	resolved := ResolvedWeaponDamage{
		Dice:                     base,
		ArmorDivisor:             armorDivisor,
		useModifyingDicePlusAdds: entity.SheetSettings.UseModifyingDicePlusAdds,
	}
	resolved.Type = strings.TrimSpace(w.Owner.ResolveSelector(selector.WeaponDamageType, w.Type, tooltip))
	if fragSpec := w.resolvedDamageString(selector.WeaponFragmentationDice, w.Fragmentation, tooltip); fragSpec != "" {
		d, sub := w.resolveDiceSpec(fragSpec)
		if sub {
			// Negative fragmentation doesn't make sense, so ignore it.
			d = dice.Dice{Sides: 6, Multiplier: 1}
		}
		// Dice that format as "0" produce no fragmentation bracket at all, so the formatted form is what decides
		// whether there is any fragmentation, just as it decides whether the bracket is printed.
		if FormatDice(d, resolved.useModifyingDicePlusAdds) != "0" {
			resolved.HasFragmentation = true
			resolved.Fragmentation = d
			resolved.FragmentationArmorDivisor = w.resolvedDamageNumeric(selector.WeaponFragmentationArmorDivisor,
				w.FragmentationArmorDivisor, tooltip)
			resolved.FragmentationType = strings.TrimSpace(w.Owner.ResolveSelector(selector.WeaponFragmentationType,
				w.FragmentationType, tooltip))
		}
	}
	return &resolved
}

// String returns the resolved damage formatted the way it appears on a sheet: the dice, the armor divisor in
// parentheses if it isn't 1, the damage type, and the fragmentation in brackets.
func (r *ResolvedWeaponDamage) String() string {
	var buffer strings.Builder
	if r.Dice.Count != 0 || r.Dice.Modifier != 0 {
		buffer.WriteString(FormatDice(r.Dice, r.useModifyingDicePlusAdds))
	}
	if r.ArmorDivisor != fxp.One {
		buffer.WriteByte('(')
		buffer.WriteString(r.ArmorDivisor.String())
		buffer.WriteByte(')')
	}
	if r.Type != "" {
		if buffer.Len() != 0 {
			buffer.WriteByte(' ')
		}
		buffer.WriteString(r.Type)
	}
	if r.HasFragmentation {
		if buffer.Len() != 0 {
			buffer.WriteByte(' ')
		}
		buffer.WriteByte('[')
		buffer.WriteString(FormatDice(r.Fragmentation, r.useModifyingDicePlusAdds))
		if r.FragmentationArmorDivisor != fxp.One {
			buffer.WriteByte('(')
			buffer.WriteString(r.FragmentationArmorDivisor.String())
			buffer.WriteByte(')')
		}
		if r.FragmentationType != "" {
			buffer.WriteByte(' ')
			buffer.WriteString(r.FragmentationType)
		}
		buffer.WriteByte(']')
	}
	return buffer.String()
}

// IsExplosive reports whether this damage is an explosion for the purposes of the explosion rules (BX414): its damage
// type carries the Explosion modifier, or it throws fragments.
func (r *ResolvedWeaponDamage) IsExplosive() bool {
	return IsExplosiveDamageType(r.Type) || r.HasFragmentation
}

// ResolvedDamage returns the damage, fully resolved for the user's sw or thr, if possible.
func (w *WeaponDamage) ResolvedDamage(tooltip *xbytes.InsertBuffer) string {
	if r := w.ResolveDamage(tooltip); r != nil {
		return r.String()
	}
	return w.String()
}

// multiplyDice returns the dice scaled by the given multiplier. Dice evaluate as ((sum of Count dice) + Modifier) *
// Multiplier, so scaling just Count and Modifier scales the whole result; scaling Multiplier as well would apply the
// multiplier a second time.
func multiplyDice(multiplier int, d dice.Dice) dice.Dice {
	d.Count *= multiplier
	d.Modifier *= multiplier
	return d
}

// halveForPhoenixFlame halves a damage bonus amount once for each of the per-level and per-die options the bonus has
// set, which is how those bonuses are adjusted for the Phoenix Flame D3 progression, whose base damage has twice as
// many dice as the standard one. A dice bonus has only its modifier halved this way, so that "1d+2 per level" stays
// in step with "+2 per level"; the dice themselves are left alone, since a die is a die whatever the progression.
func halveForPhoenixFlame(bonus *WeaponBonus, amt fxp.Int) fxp.Int {
	if bonus.PerLevel {
		amt = amt.Div(fxp.Two)
	}
	if bonus.PerDie {
		amt = amt.Div(fxp.Two)
	}
	return amt
}

// addBonusDice folds the dice contributed by the damage bonuses into the base dice. Adding dice whose sides differ
// averages the two together, which rounds, so the dice are added in the order that averages as late and as rarely as
// possible: those with the base's sides come first, since adding them is exact, and the rest are grouped by their
// sides, each group summed exactly before it is averaged into the total. The order is fixed, so the result does not
// depend on the order the bonuses were collected in. Dice that take the total below zero leave no damage at all, as
// they do for the base damage itself.
func addBonusDice(base dice.Dice, bonusDice []BonusDice) dice.Dice {
	for i := range bonusDice {
		bonusDice[i].Dice = Roller.Normalize(bonusDice[i].Dice)
	}
	slices.SortFunc(bonusDice, func(a, b BonusDice) int {
		return cmp.Or(
			cmp.Compare(boolToInt(a.Sides != base.Sides), boolToInt(b.Sides != base.Sides)),
			compareBonusDice(a, b),
		)
	})
	var sub bool
	for i := 0; i < len(bonusDice); {
		group := dice.Dice{Sides: bonusDice[i].Sides, Multiplier: 1}
		var groupSub bool
		for ; i < len(bonusDice) && bonusDice[i].Sides == group.Sides; i++ {
			group, groupSub = addDice(group, bonusDice[i].Dice, groupSub, bonusDice[i].Sub)
		}
		base, sub = addDice(base, group, sub, groupSub)
	}
	if sub {
		return dice.Dice{Sides: 6, Multiplier: 1}
	}
	return base
}

func addDice(left, right dice.Dice, leftSub, rightSub bool) (d dice.Dice, sub bool) {
	if leftSub {
		left.Count = -left.Count
	}
	if rightSub {
		right.Count = -right.Count
	}
	if left.Sides > 1 && right.Sides > 1 && left.Sides != right.Sides {
		sides := min(left.Sides, right.Sides)
		average := fxp.FromInteger(sides + 1).Div(fxp.Two)
		averageLeft := fxp.FromInteger(left.Count * (left.Sides + 1)).Div(fxp.Two).Mul(fxp.FromInteger(left.Multiplier))
		averageRight := fxp.FromInteger(right.Count * (right.Sides + 1)).Div(fxp.Two).Mul(fxp.FromInteger(right.Multiplier))
		averageBoth := averageLeft + averageRight
		d = dice.Dice{
			Count:      averageBoth.Div(average).AsInteger[int](),
			Sides:      sides,
			Modifier:   averageBoth.Mod(average).Round().AsInteger[int]() + left.Modifier + right.Modifier,
			Multiplier: 1,
		}
	} else {
		d = dice.Dice{
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

func adjustDiceForPercentBonus(d dice.Dice, percent fxp.Int) dice.Dice {
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
	return dice.Dice{
		Count:      count.AsInteger[int](),
		Sides:      d.Sides,
		Modifier:   modifier.AsInteger[int](),
		Multiplier: d.Multiplier,
	}
}

func (w *WeaponDamage) formatDiceWithSub(s string, convertMods, following bool) string {
	d, sub := w.resolveDiceSpec(s)
	if base := FormatDice(d, convertMods); base != "0" {
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

func (w *WeaponDamage) resolveDiceSpec(s string) (d dice.Dice, sub bool) {
	var ok bool
	if d, sub, ok = parsePotentialDiceSpec(s); ok {
		return d, sub
	}
	var entity *Entity
	if w.Owner != nil {
		entity = w.Owner.Entity()
	}
	value := ResolveScript(entity, deferredNewScriptWeapon(w.Owner), s)
	value = strings.TrimPrefix(strings.TrimSpace(value), "+")
	if sub = isDiceSubtraction(value); sub {
		value = value[1:]
	}
	return Roller.Parse(value), sub
}

func parsePotentialDiceSpec(s string) (d dice.Dice, sub, ok bool) {
	spec := strings.TrimLeft(strings.TrimSpace(s), "+")
	if sub = isDiceSubtraction(spec); sub {
		spec = spec[1:]
	}
	if !isCompleteDiceSpec(spec) {
		return dice.Dice{}, false, false
	}
	d = Roller.Parse(spec)
	if spec == "0" || spec == "-0" {
		return d, false, true
	}
	empty := Roller.Normalize(dice.Dice{})
	if d != empty {
		return d, sub, true
	}
	return dice.Dice{}, false, false
}

// isCompleteDiceSpec reports whether the entire string is consumed by the dice grammar, i.e. an optional count,
// followed by an optional die marker and number of sides, followed by an optional signed modifier, followed by an
// optional multiplier. The dice parser stops at the first character it can't use and silently discards the remainder,
// so anything with leftover text has to be treated as a script expression instead: without this check, "2*self.level"
// would become a flat +2 and "1d+self.level" would become 1d, with the script never evaluated.
func isCompleteDiceSpec(s string) bool {
	i := skipDiceDigits(s, 0)
	if i < len(s) && (s[i] == 'd' || s[i] == 'D') {
		i = skipDiceDigits(s, i+1)
	}
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i = skipDiceDigits(s, i+1)
	}
	if i < len(s) && (s[i] == 'x' || s[i] == 'X') {
		i = skipDiceDigits(s, i+1)
	}
	return i == len(s)
}

func skipDiceDigits(s string, i int) int {
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	return i
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
