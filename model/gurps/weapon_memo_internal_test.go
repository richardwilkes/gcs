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
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/toolbox/v2/check"
)

// newMemoTestWeapon returns a melee weapon with a minimum ST of 10 on the only trait of a recalculated entity.
func newMemoTestWeapon() (*Entity, *Trait, *Weapon) {
	e := NewEntity()
	e.Traits = nil // Drop the natural attacks
	owner := NewTrait(e, nil, false)
	owner.Name = "Claws"
	w := NewWeapon(owner, true)
	w.SetOwner(owner)
	w.Strength.Min = fxp.Ten
	owner.Weapons = []*Weapon{w}
	e.Traits = append(e.Traits, owner)
	e.Recalculate()
	return e, owner, w
}

func TestWeaponMemoFingerprintCoversWeaponData(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name   string
		change func(w *Weapon)
	}{
		{"usage", func(w *Weapon) { w.Usage = "Swung" }},
		{"usage notes", func(w *Weapon) { w.UsageNotes = "Only at night" }},
		{"minimum ST", func(w *Weapon) { w.Strength.Min += fxp.One }},
		{"base damage", func(w *Weapon) { w.Damage.Base = "2d" }},
		{"damage type", func(w *Weapon) { w.Damage.Type = "cut" }},
		{"maximum reach", func(w *Weapon) { w.Reach.Max += fxp.One }},
		{"parry modifier", func(w *Weapon) { w.Parry.Modifier += fxp.One }},
		{"can block", func(w *Weapon) { w.Block.CanBlock = !w.Block.CanBlock }},
		{"accuracy", func(w *Weapon) { w.Accuracy.Base += fxp.One }},
		{"maximum range", func(w *Weapon) { w.Range.Max += fxp.One }},
		{"shots per attack", func(w *Weapon) { w.RateOfFire.Mode1.ShotsPerAttack += fxp.One }},
		{"shots", func(w *Weapon) { w.Shots.Count += fxp.One }},
		{"bulk", func(w *Weapon) { w.Bulk.Normal -= fxp.One }},
		{"recoil", func(w *Weapon) { w.Recoil.Shot += fxp.One }},
		{"defaults", func(w *Weapon) {
			w.Defaults = append(w.Defaults, &SkillDefault{DefaultType: DexterityID, Modifier: -fxp.Four})
		}},
		{"not equipped", func(w *Weapon) { w.NotEquipped = !w.NotEquipped }},
		{"not carried", func(w *Weapon) { w.NotCarried = !w.NotCarried }},
	} {
		_, _, w := newMemoTestWeapon()
		c.NotNil(w.Entity(), one.name)
		before := w.memoFingerprint()
		one.change(w)
		c.NotEqual(before, w.memoFingerprint(), "a change to the %s must change the fingerprint", one.name)
	}
	// Resolution never reads whether the weapon is hidden, so hiding it needn't, and doesn't, set its memo aside.
	_, _, w := newMemoTestWeapon()
	before := w.memoFingerprint()
	w.Hide = !w.Hide
	c.Equal(before, w.memoFingerprint())
}

func TestWeaponMemoReset(t *testing.T) {
	c := check.New(t)
	e, owner, w := newMemoTestWeapon()
	c.Equal(fxp.Ten, w.ResolvedStrength(nil).Min)
	c.Equal(1, len(e.weaponMemos), "the resolution is memoized")

	bonus := NewWeaponBonus(feature.WeaponMinSTBonus)
	bonus.SelectionType = wsel.ThisWeapon
	bonus.Amount = fxp.Two
	owner.Features = append(owner.Features, bonus)
	c.Equal(fxp.Ten, w.ResolvedStrength(nil).Min, "the memo holds the bonuses found before the change")
	e.DiscardCaches()
	c.Equal(fxp.Twelve, w.ResolvedStrength(nil).Min, "discarding the caches brings the new bonus in")

	e.Recalculate()
	c.Equal(0, len(e.weaponMemos), "nothing is memoized during a recalculation")

	// Partway through a recalculation, nothing resolved and no list of weapons is kept.
	e.recalculating++
	c.Equal(fxp.Twelve, w.ResolvedStrength(nil).Min)
	c.Equal([]*Weapon{w}, e.Weapons(true, false, true))
	c.Equal(0, len(e.weaponMemos), "nothing is memoized while recalculating")
	c.Equal(uint8(0), e.weaponListsCached, "no weapon list is kept while recalculating")
	e.recalculating--

	// Nor is a list kept while the prerequisites of an item carrying a grant are judged, since they see the sheet
	// without what the grant gives.
	training := addTraitWithFeatures(e, "Training",
		newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.One, grantmerge.Ignore))
	e.Recalculate()
	e.JudgingPrereqsOf(training, func() { c.Equal([]*Weapon{w}, e.Weapons(true, false, true)) })
	c.Equal(uint8(0), e.weaponListsCached, "no weapon list is kept while judging those prerequisites")
}

func TestWeaponMemoOfCircularDataIsOrderIndependent(t *testing.T) {
	c := check.New(t)
	type resolved struct{ damage, rng string }
	resolve := func(rangeFirst, memos bool) resolved {
		restore := SetWeaponMemosEnabled(memos)
		defer restore()
		e := NewEntity()
		e.Traits = nil // Drop the natural attacks
		owner := NewTrait(e, nil, false)
		owner.Name = "Spit"
		w := NewWeapon(owner, false)
		w.SetOwner(owner)
		w.Damage.StrengthType = stdmg.None
		// The base damage reads the damage it is part of, adding a die each time around, until the script depth limit
		// cuts the circle off. A bonus to the maximum range brings the base damage into resolving the range too.
		w.Damage.Base = `(parseInt(self.damage) || 0) + 1 + "d"`
		bonus := NewWeaponBonus(feature.WeaponMaxRangeBonus)
		bonus.SelectionType = wsel.ThisWeapon
		bonus.Amount = fxp.One
		owner.Features = append(owner.Features, bonus)
		owner.Weapons = []*Weapon{w}
		e.Traits = append(e.Traits, owner)
		e.Recalculate()
		e.DiscardCaches()
		var result resolved
		if rangeFirst {
			result.rng = w.ResolvedRange(nil).String(true)
			result.damage = w.Damage.ResolvedDamage(nil)
		} else {
			result.damage = w.Damage.ResolvedDamage(nil)
			result.rng = w.ResolvedRange(nil).String(true)
		}
		c.Equal(0, e.weaponResolutions)
		c.False(e.circularWeaponResolution)
		c.Equal(result, resolved{damage: w.Damage.ResolvedDamage(nil), rng: w.ResolvedRange(nil).String(true)},
			"memos %v, range first %v: asking again", memos, rangeFirst)
		return result
	}
	want := resolve(false, false)
	c.Equal(want, resolve(true, false), "without memos, the order of resolution doesn't matter")
	c.Equal(want, resolve(false, true), "damage first")
	c.Equal(want, resolve(true, true), "range first")
}

func TestEntityWeaponsOfCircularDataMatchUncached(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name       string
		damage     string
		usageNotes string
	}{
		// The base damage reads the damage it is part of and how many weapons are listed, so whether the two copies are
		// told apart depends on how far around the circle each resolution got before the script depth limit cut it off.
		{name: "damage", damage: `((parseInt(self.damage) || 0) + entity.weapons.length) % 4 + 1 + "d"`},
		// The same, through text that isn't memoized.
		{
			name:       "usage notes",
			damage:     "1d",
			usageNotes: `<script>(parseInt(self.usageNotes) || 0) + entity.weapons.length</script>`,
		},
	} {
		// describeLists returns, for each of two calls to Entity.Weapons, which copies it listed and what they resolve
		// to. With askFirst at 0 or more, that copy is resolved first, so that its scripts ask for the list from inside
		// the circle.
		describeLists := func(memos bool, askFirst int) []string {
			restore := SetWeaponMemosEnabled(memos)
			defer restore()
			e := NewEntity()
			e.Traits = nil // Drop the natural attacks
			copies := make([]*Weapon, 2)
			for i := range copies {
				owner := NewTrait(e, nil, false)
				owner.Name = "Spit"
				w := NewWeapon(owner, true)
				w.SetOwner(owner)
				w.Damage.StrengthType = stdmg.None
				w.Damage.Base = one.damage
				w.UsageNotes = one.usageNotes
				owner.Weapons = []*Weapon{w}
				e.Traits = append(e.Traits, owner)
				copies[i] = w
			}
			e.Recalculate()
			e.DiscardCaches()
			if askFirst >= 0 {
				copies[askFirst].HashDisplayedState()
			}
			lists := make([]string, 0, 2)
			for range 2 {
				listed := e.Weapons(true, false, true)
				var buffer strings.Builder
				for i, w := range copies {
					if slices.Contains(listed, w) {
						fmt.Fprintf(&buffer, "[%d: %s; %s]", i, w.Damage.ResolvedDamage(nil), w.ResolveUsageNotes())
					}
				}
				lists = append(lists, buffer.String())
			}
			return lists
		}
		for askFirst := -1; askFirst < 2; askFirst++ {
			c.Equal(describeLists(false, askFirst), describeLists(true, askFirst), "%s, asking first for copy %d",
				one.name, askFirst)
		}
	}
}
