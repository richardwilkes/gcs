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
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/zeebo/xxh3"
)

// maximumWeaponMemos caps the memos an entity holds between recalculations. A sheet's own weapons number in the
// dozens; the cap only matters for the copies the item editors make, which belong to the entity too, and it is
// enforced by starting over rather than by evicting, since a recalculation starts over anyway.
const maximumWeaponMemos = 4096

// weaponMemosEnabled lets a test compare the memoized results, and the weapon lists Entity.Weapons keeps, with those
// computed afresh.
var weaponMemosEnabled = true

// memoized is a resolved value along with the text resolving it added to the modifiers tooltip, so that a later call
// asking for the tooltip gets exactly the text the resolution produced.
type memoized[T any] struct {
	value     T
	tooltip   string
	done      bool
	resolving bool
}

// requiredSkill names a skill a weapon is used with; see Weapon.forEachRequiredSkill.
type requiredSkill struct {
	name           string
	specialization string
}

// weaponMemo holds what resolving a weapon's fields has found for the current state of the entity's derived data:
// the skills the weapon is used with, every weapon bonus that applies to it, and each resolved field once it has been
// resolved. Resolving a field collects the bonuses, which walks the skills several times over, and a weapon's fields
// are resolved over and over -- once per column of its row per sync of its table, again for the hash that tells the
// weapons apart, again on every draw -- so without the memo a sheet with many weapons spent most of its time here. A
// memo lives in the entity's weaponMemos, which discardCaches drops, so it lasts until the next recalculation or
// DiscardCaches. Weapon.memo replaces it sooner only when the weapon's own data changes or it is given another owner;
// a change to the owner's data, such as its rated ST or features, or to the rest of the entity is not seen until then.
type weaponMemo struct {
	owner          WeaponOwner
	fingerprint    uint64
	requiredSkills []requiredSkill
	bonuses        []*WeaponBonus
	hasRequired    bool
	hasBonuses     bool
	skillLevel     memoized[fxp.Int]
	strength       memoized[attack.Strength]
	parry          memoized[attack.Parry]
	block          memoized[attack.Block]
	reach          memoized[attack.Reach]
	accuracy       memoized[attack.Accuracy]
	rng            memoized[attack.Range]
	rateOfFire     memoized[attack.RoF]
	shots          memoized[attack.Shots]
	bulk           memoized[attack.Bulk]
	recoil         memoized[attack.Recoil]
	damage         memoized[*ResolvedWeaponDamage]
	baseDice       memoized[dice.Dice]
}

// weaponCachesUsable returns true if what is resolved about the entity's weapons, and the lists of them Weapons
// builds, may be kept and served from what was kept. Neither may be while the entity is recalculating, since its
// derived state changes from one step of a pass to the next, so a value kept early in a pass could be served stale
// later in it, nor while the prerequisites of an item carrying a grant are judged, since the weapons then resolve as
// those prerequisites see the sheet, without what the grant gives (see JudgingPrereqsOf).
func (e *Entity) weaponCachesUsable() bool {
	return weaponMemosEnabled && e.recalculating == 0 && !e.carriesGrant(e.judging)
}

// memo returns the memo for the weapon's current data and owner, or nil when nothing may be memoized: a weapon with
// no entity has nothing to resolve against, and one whose entity's caches can't be used may not keep it (see
// Entity.weaponCachesUsable).
func (w *Weapon) memo() *weaponMemo {
	entity := w.Entity()
	if entity == nil || !entity.weaponCachesUsable() {
		return nil
	}
	fingerprint := w.memoFingerprint()
	if m, ok := entity.weaponMemos[w]; ok && m.owner == w.Owner && m.fingerprint == fingerprint {
		return m
	}
	if entity.weaponMemos == nil || len(entity.weaponMemos) >= maximumWeaponMemos {
		entity.weaponMemos = make(map[*Weapon]*weaponMemo)
	}
	m := &weaponMemo{owner: w.Owner, fingerprint: fingerprint}
	entity.weaponMemos[w] = m
	return m
}

// memoFingerprint hashes everything of the weapon itself that resolving its fields reads: its data, and the carried
// and equipped flags the entity sets on it, which String reports in the minimum-ST tooltip and HashDisplayedState.
func (w *Weapon) memoFingerprint() uint64 {
	h := xxh3.New()
	w.Hash(h)
	xhash.Bool(h, w.NotEquipped)
	xhash.Bool(h, w.NotCarried)
	return h.Sum64()
}

// memoize returns the value in the memo's slot, resolving it first if it isn't there yet, and adds the text that
// resolution produced to the tooltip, if one was given. A resolution that may be kept runs with a buffer, so that the
// text is on hand for whichever later call wants it. No resolved value depends on the tooltip, but some resolutions
// read its length to separate what they add from what was already there, so a call with a buffer that isn't empty
// resolves afresh into it, as does one for a weapon that can't be memoized. What is resolved while a resolution comes
// back around to a slot still being resolved isn't kept; see resolveSlot.
func memoize[T any](w *Weapon, slot func(*weaponMemo) *memoized[T], tooltip *xbytes.InsertBuffer, resolve func(*xbytes.InsertBuffer) T) T {
	m := w.memo()
	if m == nil || (tooltip != nil && tooltip.Len() != 0) {
		return resolve(tooltip)
	}
	s := slot(m)
	if !s.done {
		entity := w.Entity()
		if s.resolving {
			entity.circularWeaponResolution = true
			return resolve(tooltip)
		}
		var buffer xbytes.InsertBuffer
		value, keep := resolveSlot(entity, s, &buffer, resolve)
		if !keep {
			if tooltip != nil && buffer.Len() != 0 {
				tooltip.WriteString(buffer.String())
			}
			return value
		}
		s.value = value
		s.tooltip = buffer.String()
		s.done = true
	}
	if tooltip != nil && s.tooltip != "" {
		tooltip.WriteString(s.tooltip)
	}
	return s.value
}

// resolveSlot resolves the slot's value into buffer and reports whether it may be kept. Only circular data, such as a
// damage script that reads the weapon's own damage, comes back around to a slot still being resolved. What the
// resolutions in progress then find depends on how far around the circle each got before the script depth limit cut
// it off, so no resolution that finishes from then on is kept, up to and including the outermost one.
func resolveSlot[T any](entity *Entity, s *memoized[T], buffer *xbytes.InsertBuffer, resolve func(*xbytes.InsertBuffer) T) (value T, keep bool) {
	s.resolving = true
	entity.weaponResolutions++
	defer func() {
		s.resolving = false
		entity.weaponResolutions--
		if entity.weaponResolutions == 0 {
			entity.circularWeaponResolution = false
		}
	}()
	value = resolve(buffer)
	return value, !entity.circularWeaponResolution
}
