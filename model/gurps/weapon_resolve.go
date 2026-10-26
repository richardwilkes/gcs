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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/attack"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// ratedStrength returns the owner's rated ST, or 0 when the weapon has no owner or the owner has none.
func (w *Weapon) ratedStrength() fxp.Int {
	if w.Owner == nil {
		return 0
	}
	return w.Owner.RatedStrength()
}

// ResolvedStrength returns the weapon's strength with any bonuses that apply resolved. The result is kept as Weapon
// describes.
func (w *Weapon) ResolvedStrength(modifiersTooltip *xbytes.InsertBuffer) attack.Strength {
	return memoize(w, func(m *weaponMemo) *memoized[attack.Strength] { return &m.strength }, modifiersTooltip, w.resolveStrength)
}

func (w *Weapon) resolveStrength(modifiersTooltip *xbytes.InsertBuffer) attack.Strength {
	result := w.Strength
	result.Bipod = w.ResolveBoolFlag(wswitch.Bipod, result.Bipod)
	result.Mounted = w.ResolveBoolFlag(wswitch.Mounted, result.Mounted)
	result.MusketRest = w.ResolveBoolFlag(wswitch.MusketRest, result.MusketRest)
	result.TwoHanded = w.ResolveBoolFlag(wswitch.TwoHanded, result.TwoHanded)
	result.TwoHandedUnready = w.ResolveBoolFlag(wswitch.TwoHandedAndUnreadyAfterAttack, result.TwoHandedUnready)
	result.Min = w.weaponAdjustment(oneDieCount, modifiersTooltip, feature.WeaponMinSTBonus).applyTo(result.Min)
	result.Validate()
	return result
}

// ResolvedParry returns the weapon's parry with any bonuses that apply resolved. The result is kept as Weapon
// describes.
func (w *Weapon) ResolvedParry(modifiersTooltip *xbytes.InsertBuffer) attack.Parry {
	return memoize(w, func(m *weaponMemo) *memoized[attack.Parry] { return &m.parry }, modifiersTooltip, w.resolveParry)
}

func (w *Weapon) resolveParry(modifiersTooltip *xbytes.InsertBuffer) attack.Parry {
	result := w.Parry
	result.CanParry = w.ResolveBoolFlag(wswitch.CanParry, result.CanParry)
	if result.CanParry {
		result.Fencing = w.ResolveBoolFlag(wswitch.Fencing, result.Fencing)
		result.Unbalanced = w.ResolveBoolFlag(wswitch.Unbalanced, result.Unbalanced)
		if entity := w.Entity(); entity != nil {
			result.Modifier = w.resolveDefenseModifier(entity, modifiersTooltip, result.Modifier, ParryID,
				feature.WeaponParryBonus, entity.ParryBonusTooltip)
		}
	}
	result.Validate()
	return result
}

// ResolvedBlock returns the weapon's block with any bonuses that apply resolved. The result is kept as Weapon
// describes.
func (w *Weapon) ResolvedBlock(modifiersTooltip *xbytes.InsertBuffer) attack.Block {
	return memoize(w, func(m *weaponMemo) *memoized[attack.Block] { return &m.block }, modifiersTooltip, w.resolveBlock)
}

func (w *Weapon) resolveBlock(modifiersTooltip *xbytes.InsertBuffer) attack.Block {
	result := w.Block
	result.CanBlock = w.ResolveBoolFlag(wswitch.CanBlock, result.CanBlock)
	if result.CanBlock {
		if entity := w.Entity(); entity != nil {
			result.Modifier = w.resolveDefenseModifier(entity, modifiersTooltip, result.Modifier, BlockID,
				feature.WeaponBlockBonus, entity.BlockBonusTooltip)
		}
	}
	result.Validate()
	return result
}

// ResolvedReach returns the weapon's reach with any bonuses that apply resolved. The result is kept as Weapon
// describes.
func (w *Weapon) ResolvedReach(modifiersTooltip *xbytes.InsertBuffer) attack.Reach {
	return memoize(w, func(m *weaponMemo) *memoized[attack.Reach] { return &m.reach }, modifiersTooltip, w.resolveReach)
}

func (w *Weapon) resolveReach(modifiersTooltip *xbytes.InsertBuffer) attack.Reach {
	result := w.Reach
	result.CloseCombat = w.ResolveBoolFlag(wswitch.CloseCombat, result.CloseCombat)
	result.ChangeRequiresReady = w.ResolveBoolFlag(wswitch.ReachChangeRequiresReady, result.ChangeRequiresReady)
	adj := w.weaponAdjustments(w.baseDamageDieCount, modifiersTooltip, feature.WeaponMinReachBonus,
		feature.WeaponMaxReachBonus)
	result.Min = adj[feature.WeaponMinReachBonus].applyTo(result.Min)
	result.Max = adj[feature.WeaponMaxReachBonus].applyTo(result.Max)
	result.Validate()
	return result
}

// ResolvedAccuracy returns the weapon's accuracy with any bonuses that apply resolved. The result is kept as Weapon
// describes.
func (w *Weapon) ResolvedAccuracy(modifiersTooltip *xbytes.InsertBuffer) attack.Accuracy {
	return memoize(w, func(m *weaponMemo) *memoized[attack.Accuracy] { return &m.accuracy }, modifiersTooltip, w.resolveAccuracy)
}

func (w *Weapon) resolveAccuracy(modifiersTooltip *xbytes.InsertBuffer) attack.Accuracy {
	result := w.Accuracy
	result.Jet = w.ResolveBoolFlag(wswitch.Jet, result.Jet)
	if !result.Jet {
		if entity := w.Entity(); entity != nil {
			adj := w.weaponAdjustments(w.baseDamageDieCount, modifiersTooltip, feature.WeaponAccBonus,
				feature.WeaponScopeAccBonus)
			result.Base = adj[feature.WeaponAccBonus].applyTo(result.Base)
			result.Scope = adj[feature.WeaponScopeAccBonus].applyTo(result.Scope)
		}
	}
	result.Validate()
	return result
}

// ResolvedRange returns the weapon's range with any bonuses that apply resolved. The result is kept as Weapon
// describes.
func (w *Weapon) ResolvedRange(modifiersTooltip *xbytes.InsertBuffer) attack.Range {
	return memoize(w, func(m *weaponMemo) *memoized[attack.Range] { return &m.rng }, modifiersTooltip, w.resolveRange)
}

func (w *Weapon) resolveRange(modifiersTooltip *xbytes.InsertBuffer) attack.Range {
	result := w.Range
	result.MusclePowered = w.ResolveBoolFlag(wswitch.MusclePowered, result.MusclePowered)
	result.InMiles = w.ResolveBoolFlag(wswitch.RangeInMiles, result.InMiles)
	if result.MusclePowered {
		st := w.effectiveStrength(func(entity *Entity) fxp.Int {
			switch w.Damage.resolvedStrengthType(nil) {
			case stdmg.TelekineticThrust, stdmg.TelekineticSwing:
				return entity.TelekineticStrength()
			case stdmg.IQThrust, stdmg.IQSwing:
				return entity.resolveAttributeCurrentForPrereqs(IntelligenceID).Max(0).Floor()
			default:
				return entity.ThrowingStrength()
			}
		}, modifiersTooltip)
		if st > 0 {
			result.HalfDamage = result.HalfDamage.Mul(st).Floor().Max(0)
			result.Min = result.Min.Mul(st).Floor().Max(0)
			result.Max = result.Max.Mul(st).Floor().Max(0)
		}
	}
	adj := w.weaponAdjustments(w.baseDamageDieCount, modifiersTooltip, feature.WeaponHalfDamageRangeBonus,
		feature.WeaponMinRangeBonus, feature.WeaponMaxRangeBonus)
	result.HalfDamage = adj[feature.WeaponHalfDamageRangeBonus].applyTo(result.HalfDamage)
	result.Min = adj[feature.WeaponMinRangeBonus].applyTo(result.Min)
	result.Max = adj[feature.WeaponMaxRangeBonus].applyTo(result.Max)
	result.Validate()
	return result
}

// ResolvedRateOfFire returns the weapon's rate of fire with any bonuses that apply resolved. The result is kept as
// Weapon describes.
func (w *Weapon) ResolvedRateOfFire(modifiersTooltip *xbytes.InsertBuffer) attack.RoF {
	return memoize(w, func(m *weaponMemo) *memoized[attack.RoF] { return &m.rateOfFire }, modifiersTooltip, w.resolveRateOfFire)
}

func (w *Weapon) resolveRateOfFire(modifiersTooltip *xbytes.InsertBuffer) attack.RoF {
	result := w.RateOfFire
	result.Jet = w.ResolveBoolFlag(wswitch.Jet, result.Jet)
	if !result.Jet {
		var buf1, buf2 xbytes.InsertBuffer
		result.Mode1 = w.resolvedRoFMode(result.Mode1, &buf1, true)
		result.Mode2 = w.resolvedRoFMode(result.Mode2, &buf2, false)
		if modifiersTooltip != nil {
			if buf1.Len() != 0 {
				if buf2.Len() != 0 {
					modifiersTooltip.WriteString(i18n.Text("First mode:\n"))
				}
				modifiersTooltip.WriteString(buf1.String())
			}
			if buf2.Len() != 0 {
				if buf1.Len() != 0 {
					modifiersTooltip.WriteString(i18n.Text("\n\nSecond mode:\n"))
				}
				modifiersTooltip.WriteString(buf2.String())
			}
		}
	}
	result.Validate()
	return result
}

// resolvedRoFMode returns mode with any bonuses that apply resolved. firstMode selects which mode's switches and
// bonuses are consulted.
func (w *Weapon) resolvedRoFMode(mode attack.RoFMode, modifiersTooltip *xbytes.InsertBuffer, firstMode bool) attack.RoFMode {
	result := mode
	var shotsFeature, secondaryFeature feature.Type
	if firstMode {
		shotsFeature = feature.WeaponRofMode1ShotsBonus
		secondaryFeature = feature.WeaponRofMode1SecondaryBonus
		result.FullAutoOnly = w.ResolveBoolFlag(wswitch.FullAuto1, mode.FullAutoOnly)
		result.HighCyclicControlledBursts = w.ResolveBoolFlag(wswitch.ControlledBursts1, mode.HighCyclicControlledBursts)
	} else {
		shotsFeature = feature.WeaponRofMode2ShotsBonus
		secondaryFeature = feature.WeaponRofMode2SecondaryBonus
		result.FullAutoOnly = w.ResolveBoolFlag(wswitch.FullAuto2, mode.FullAutoOnly)
		result.HighCyclicControlledBursts = w.ResolveBoolFlag(wswitch.ControlledBursts2, mode.HighCyclicControlledBursts)
	}
	adj := w.weaponAdjustments(w.baseDamageDieCount, modifiersTooltip, shotsFeature, secondaryFeature)
	result.ShotsPerAttack = adj[shotsFeature].applyTo(result.ShotsPerAttack)
	result.SecondaryProjectiles = adj[secondaryFeature].applyTo(result.SecondaryProjectiles)
	result.Validate()
	return result
}

// ResolvedShots returns the weapon's shots with any bonuses that apply resolved. The result is kept as Weapon
// describes.
func (w *Weapon) ResolvedShots(modifiersTooltip *xbytes.InsertBuffer) attack.Shots {
	return memoize(w, func(m *weaponMemo) *memoized[attack.Shots] { return &m.shots }, modifiersTooltip, w.resolveShots)
}

func (w *Weapon) resolveShots(modifiersTooltip *xbytes.InsertBuffer) attack.Shots {
	result := w.Shots
	result.ReloadTimeIsPerShot = w.ResolveBoolFlag(wswitch.ReloadTimeIsPerShot, result.ReloadTimeIsPerShot)
	result.Thrown = w.ResolveBoolFlag(wswitch.Thrown, result.Thrown)
	adj := w.weaponAdjustments(w.baseDamageDieCount, modifiersTooltip, feature.WeaponNonChamberShotsBonus,
		feature.WeaponChamberShotsBonus, feature.WeaponShotDurationBonus, feature.WeaponReloadTimeBonus)
	result.Count = adj[feature.WeaponNonChamberShotsBonus].applyTo(result.Count)
	result.InChamber = adj[feature.WeaponChamberShotsBonus].applyTo(result.InChamber)
	result.Duration = adj[feature.WeaponShotDurationBonus].applyTo(result.Duration)
	result.ReloadTime = adj[feature.WeaponReloadTimeBonus].applyTo(result.ReloadTime)
	result.Validate()
	return result
}

// ResolvedBulk returns the weapon's bulk with any bonuses that apply resolved. The result is kept as Weapon describes.
func (w *Weapon) ResolvedBulk(modifiersTooltip *xbytes.InsertBuffer) attack.Bulk {
	return memoize(w, func(m *weaponMemo) *memoized[attack.Bulk] { return &m.bulk }, modifiersTooltip, w.resolveBulk)
}

func (w *Weapon) resolveBulk(modifiersTooltip *xbytes.InsertBuffer) attack.Bulk {
	result := w.Bulk
	result.RetractingStock = w.ResolveBoolFlag(wswitch.RetractingStock, result.RetractingStock)
	adj := w.weaponAdjustment(w.baseDamageDieCount, modifiersTooltip, feature.WeaponBulkBonus)
	result.Normal = adj.applyTo(result.Normal)
	if w.Bulk.Giant != 0 { // 0 means there is no separate giant bulk, so don't let the bonuses materialize one.
		result.Giant = adj.applyTo(result.Giant)
	}
	result.Validate()
	return result
}

// ResolvedRecoil returns the weapon's recoil with any bonuses that apply resolved. The result is kept as Weapon
// describes.
func (w *Weapon) ResolvedRecoil(modifiersTooltip *xbytes.InsertBuffer) attack.Recoil {
	return memoize(w, func(m *weaponMemo) *memoized[attack.Recoil] { return &m.recoil }, modifiersTooltip, w.resolveRecoil)
}

func (w *Weapon) resolveRecoil(modifiersTooltip *xbytes.InsertBuffer) attack.Recoil {
	result := w.Recoil
	// 0 means recoil isn't used; 1+ means it is.
	if w.Recoil.Shot > 0 || w.Recoil.Slug > 0 {
		adj := w.weaponAdjustment(w.baseDamageDieCount, modifiersTooltip, feature.WeaponRecoilBonus)
		result.Shot = adj.applyTo(result.Shot)
		result.Slug = adj.applyTo(result.Slug)
		if w.Recoil.Shot > 0 {
			result.Shot = result.Shot.Max(fxp.One)
		} else {
			result.Shot = 0
		}
		if w.Recoil.Slug > 0 {
			result.Slug = result.Slug.Max(fxp.One)
		} else {
			result.Slug = 0
		}
	}
	result.Validate()
	return result
}
