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
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// BonusDice is the dice of a weapon damage bonus: a complete dice specification, such as "1d", "2d3+1" or "2d+1x3",
// which is added to the weapon's damage the same way a base damage specification is. When Sub is set, the dice
// themselves are taken away instead, but only the dice: as in the base damage notation, the modifier and multiplier
// still apply as written, so "-1d+2" removes one die and adds 2, and "-2dx3" removes two dice and then multiplies the
// whole total by 3. The zero value is no dice. In text, it is the specification with a leading "-" when Sub is set.
type BonusDice struct {
	dice.Dice
	Sub bool
}

// ParseBonusDice parses a dice specification, optionally preceded by a sign, into a BonusDice. It reports false when
// the whole of the text is not a dice specification with at least one die in it. A bare number, such as "2" or "-2",
// is a flat amount rather than dice, so it is rejected here and left to ParseWeaponDamageBonus, which keeps every
// BonusDice either empty or carrying dice. A zero specification, such as "0", is no dice.
func ParseBonusDice(text string) (BonusDice, bool) {
	d, sub, ok := parsePotentialDiceSpec(text)
	if !ok {
		return BonusDice{}, false
	}
	b := BonusDice{Dice: d, Sub: sub}
	if b.IsZero() {
		return BonusDice{}, true
	}
	if b.Count == 0 {
		return BonusDice{}, false
	}
	return b, true
}

// IsZero returns true if there are no dice and no modifier, i.e. nothing to add to the damage. Parsed dice always have
// a count, so for them this is the same as having no dice; only scaling can leave a modifier behind without any.
func (b BonusDice) IsZero() bool {
	return b.Count == 0 && b.Modifier == 0
}

// String returns the text form of the dice, in the notation the weapon damage uses, which is empty when there are
// none.
func (b BonusDice) String() string {
	if b.IsZero() {
		return ""
	}
	s := Roller.Format(b.Dice)
	if b.Sub {
		return "-" + s
	}
	return s
}

// StringWithSign returns the text form of the dice, with a leading "+" when they are added rather than taken away.
func (b BonusDice) StringWithSign() string {
	if b.IsZero() || b.Sub {
		return b.String()
	}
	return "+" + b.String()
}

// MarshalText implements encoding.TextMarshaler.
func (b BonusDice) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler. The whole of the text must be a dice specification; empty text is
// no dice.
func (b *BonusDice) UnmarshalText(text []byte) error {
	if strings.TrimSpace(string(text)) == "" {
		*b = BonusDice{}
		return nil
	}
	d, ok := ParseBonusDice(string(text))
	if !ok {
		return errs.Newf("invalid bonus dice: %q", string(text))
	}
	*b = d
	return nil
}

// Hash writes this object's contents into the hasher.
func (b BonusDice) Hash(h hash.Hash) {
	b.Dice.Hash(h)
	xhash.Bool(h, b.Sub)
}

// scaled returns the dice with their count and modifier multiplied by the given factor, dropping any fraction, which
// is how a per-die or per-level bonus is scaled. The multiplier is left alone, since scaling it as well would apply the
// factor twice.
func (b BonusDice) scaled(factor fxp.Int) BonusDice {
	if b.IsZero() {
		return BonusDice{}
	}
	b.Count = fxp.FromInteger(b.Count).Mul(factor).AsInteger[int]()
	b.Modifier = fxp.FromInteger(b.Modifier).Mul(factor).AsInteger[int]()
	if b.IsZero() {
		return BonusDice{}
	}
	return b
}

// compareBonusDice orders bonus dice by their sides and then their remaining parts, so that a set of them is always
// added to a weapon's damage in the same order, since adding dice with differing sides involves rounding that can
// depend on the order.
func compareBonusDice(a, b BonusDice) int {
	return cmp.Or(
		cmp.Compare(a.Sides, b.Sides),
		cmp.Compare(a.Count, b.Count),
		cmp.Compare(a.Modifier, b.Modifier),
		cmp.Compare(a.Multiplier, b.Multiplier),
		cmp.Compare(boolToInt(a.Sub), boolToInt(b.Sub)),
	)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ParseWeaponDamageBonus parses the text of a weapon damage bonus, which may be a signed number, such as "+2" or
// "-0.5", or a complete dice specification, optionally preceded by a sign, such as "+1d", "-1d", "1d+2" or "2d+1x3",
// into its dice and flat parts. A number is always the flat part, so only text with dice in it produces dice. It
// reports false for text in any other form.
func ParseWeaponDamageBonus(text string) (d BonusDice, amount fxp.Int, ok bool) {
	text = strings.TrimSpace(text)
	if v, err := fxp.FromString(text); err == nil {
		return BonusDice{}, v, true
	}
	if d, ok = ParseBonusDice(text); ok {
		return d, 0, true
	}
	return BonusDice{}, 0, false
}

// FormatWeaponDamageBonus returns the text form of a weapon damage bonus with the given dice and flat parts, always
// carrying a leading sign: "+2", "-1d" or "+2d+1x3". Both parts are shown when both are present, which only a
// hand-edited file can produce, since they are added to the damage together.
func FormatWeaponDamageBonus(d BonusDice, amount fxp.Int) string {
	if d.IsZero() {
		return amount.StringWithSign()
	}
	if amount == 0 {
		return d.StringWithSign()
	}
	return d.StringWithSign() + " " + amount.StringWithSign()
}
