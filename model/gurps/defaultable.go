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
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xreflect"
)

// defaultable is an item whose level may rest on one of its defaults, and that the defaults of other items may select.
type defaultable interface {
	fmt.Stringer
	DataOwner() DataOwner
	Enabled() bool
	TL() string
	NameWithReplacements() string
	NameableReplacements() map[string]string
	AdjustedPoints(tooltip *xbytes.InsertBuffer) fxp.Int
	WildcardDifficulty() bool
	UpdateLevel() bool
	CanSwapDefaults() bool
	// levelData returns the level the last recalculation recorded.
	levelData() *Level
	// levelExcluding returns the level computed afresh, with the items the excludes name (see excludeKey) left out of
	// those it may rest on.
	levelExcluding(excludes map[string]bool) Level
	// excludeKey returns the name of the item in a set of excludes.
	excludeKey() string
	// ownBonus returns the bonus the item's computed level holds from the bonuses that select it.
	ownBonus() fxp.Int
	// ownPowerSource returns the power source of the item, a spell, that a default of its own whose power source is to
	// be the same as that spell's selects by (see SkillDefault.withPowerSourceOf), or nil for an item that is no spell.
	ownPowerSource() *string
	// picksDefault returns true if the item picks which of its defaults its level rests on.
	picksDefault() bool
	// declaredDefaults returns the defaults the item's data holds.
	declaredDefaults() []*SkillDefault
	// syntheticDefaults returns the defaults the item has without declaring them.
	syntheticDefaults() []*SkillDefault
	// recordedDefault returns the default the item recorded as the one its level rests on, or nil. A Ritual Magic spell
	// records the best of its defaults even while its ritual skills give it a level that one doesn't beat; see
	// restingDefault.
	recordedDefault() *SkillDefault
	recordDefault(def *SkillDefault)
	// restingDefault returns the default the item recorded, or nil when it has none or the last update of its level
	// found that its level doesn't rest on that one.
	restingDefault() *SkillDefault
	// levelRestsOn returns true if the item's level would rest on the default, its level set (see setDefaultPoints),
	// were that the one it recorded.
	levelRestsOn(def *SkillDefault) bool
	// defaultBaseline returns the level, rounded down, that 1 point buys the item, from which the points a default's
	// level stands for are measured, and whether those points add to the item's own; see setDefaultPoints.
	defaultBaseline() (baseline fxp.Int, stacks bool)
}

// asSkill returns the item as a skill, or nil when it is not one.
func asSkill(h defaultable) *Skill {
	if sk, ok := h.(*Skill); ok {
		return sk
	}
	return nil
}

// entityOf returns the entity that owns the item, or nil.
func entityOf(h defaultable) *Entity {
	if owner := h.DataOwner(); !xreflect.IsNil(owner) {
		return owner.OwningEntity()
	}
	return nil
}

// updateLevel is the shared body of UpdateLevel: it records the default the item's level rests on and that level,
// returning true if either has changed.
func updateLevel(h defaultable) bool {
	level := h.levelData()
	saved := *level
	if !h.Enabled() {
		*level = Level{}
		return saved != *level
	}
	savedDefault := h.recordedDefault()
	var computed Level
	if anyScriptAbandonedDuring(entityOf(h), func() {
		// Set the default before computing the level, since levelExcluding reads it from the item. Computing against
		// the previous pass's default would leave the level a pass behind, and since Entity.Recalculate stops once the
		// levels settle, that pass would never come.
		h.recordDefault(bestDefaultWithPoints(h, nil))
		computed = h.levelExcluding(nil)
	}) {
		// A script was stopped before it could answer, so both values came from a stand-in. Keep the old ones, which
		// are merely out of date rather than wrong -- the recorded default is written to disk, so it would be wrong
		// there too. The next recalculation that gets through its scripts puts both right.
		h.recordDefault(savedDefault)
		return false
	}
	*level = computed
	// A change to the default is reported too, since other items read it: an item judging whether it may default to
	// this one, or a technique judging whether this one is usable as its base.
	return saved != *level || !sameSkillDefault(savedDefault, h.recordedDefault())
}

// sameSkillDefault returns true if the two defaults, either of which may be nil, hold the same values.
func sameSkillDefault(a, b *SkillDefault) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// bestDefaultWithPoints is bestDefault with the points of the default it returns set (see setDefaultPoints), or nil
// for an item that picks none.
func bestDefaultWithPoints(h defaultable, excluded *SkillDefault) *SkillDefault {
	if !h.picksDefault() {
		return nil
	}
	best := bestDefault(h, excluded)
	if best != nil {
		setDefaultPoints(h, best)
	}
	return best
}

// defaultWithEdits returns the default onSheet, a skill or spell, would rest on once the edits held by edited, a copy
// of it, are applied, as the next recalculation picks it (see updateLevel), with its points set, or nil for none. When
// defaultsEdited, applying the edits drops the default onSheet recorded (see SkillEditData.ApplyTo), so the copy picks
// one afresh. The defaults are picked with onSheet found as edited (see Entity.matchLevelCapsAs), so that one leading
// back to it is passed over. With no sheet, nothing picks one, so the copy keeps what it recorded.
func defaultWithEdits(onSheet, edited defaultable, defaultsEdited bool) *SkillDefault {
	if defaultsEdited {
		edited.recordDefault(nil)
	}
	e := entityOf(onSheet)
	if e == nil {
		return edited.recordedDefault()
	}
	defer e.matchLevelCapsAs(onSheet, edited)()
	return bestDefaultWithPoints(edited, nil)
}

// keptDefaultWithPoints is bestDefaultWithPoints with the default the item recorded kept, when it resolves, whether
// or not the item has the points it keeps it for; see keepsDefault.
func keptDefaultWithPoints(h defaultable) *SkillDefault {
	if !h.picksDefault() || entityOf(h) == nil {
		return nil
	}
	kept := bestDefaultKeeping(h, nil, map[string]bool{h.excludeKey(): true},
		h.recordedDefault() != nil && len(h.declaredDefaults()) > 0)
	if kept != nil {
		setDefaultPoints(h, kept)
	}
	return kept
}

// setDefaultPoints sets the default's adjusted level to its level rounded down, and its points to those that level
// stands for in the item, measured from its baseline (see defaultable.defaultBaseline), or to none for an item whose
// own points don't add to them.
func setDefaultPoints(h defaultable, def *SkillDefault) {
	baseline, stacks := h.defaultBaseline()
	level := def.Level.Floor()
	def.AdjLevel = level
	switch {
	case !stacks:
		def.Points = 0
	case level == baseline:
		def.Points = fxp.One
	case level == baseline+fxp.One:
		def.Points = fxp.Two
	case level > baseline+fxp.One:
		def.Points = fxp.Four.Mul(level - (baseline + fxp.One))
	default:
		def.Points = -level.Max(0)
	}
}

// bestDefault returns the best of the item's defaults other than excluded, with the levels taken with the item
// excluded, or nil.
func bestDefault(h defaultable, excluded *SkillDefault) *SkillDefault {
	return bestDefaultWith(h, excluded, map[string]bool{h.excludeKey(): true})
}

// keepsDefault returns true if the item keeps the default it recorded for as long as that resolves. An item with
// points that already records which default it uses keeps it, rather than silently jumping to whichever default
// currently yields the highest level; the user changes it explicitly via "Swap Defaults". Items with no points, those
// whose points don't add to those of their default (see defaultable.defaultBaseline), so that it never holds any of
// theirs, and those whose defaults are purely synthetic (such as a skill's automatic -2 default to a sibling's
// optional specialization), always track the dynamically-best default.
func keepsDefault(h defaultable) bool {
	if h.recordedDefault() == nil || h.AdjustedPoints(nil) <= 0 || len(h.declaredDefaults()) == 0 {
		return false
	}
	_, stacks := h.defaultBaseline()
	return stacks
}

// bestDefaultWith is bestDefault with the levels of the defaults taken with the given excludes, which hold the item.
func bestDefaultWith(h defaultable, excluded *SkillDefault, excludes map[string]bool) *SkillDefault {
	if entityOf(h) == nil {
		return nil
	}
	return bestDefaultKeeping(h, excluded, excludes, excluded == nil && keepsDefault(h))
}

// bestDefaultKeeping is bestDefaultWith with the default the item recorded preferred, when it resolves, if
// keepCurrent is set.
func bestDefaultKeeping(h defaultable, excluded *SkillDefault, excludes map[string]bool, keepCurrent bool) *SkillDefault {
	replacements := h.NameableReplacements()
	var bestDef, preferredDef *SkillDefault
	best := fxp.Min
	for _, def := range resolveToSpecificDefaults(h) {
		// Skip the excluded default and any that circle back through this item.
		if def.Equivalent(replacements, excluded) || inDefaultChain(h, def, make(map[defaultable]bool)) {
			continue
		}
		level := calcSkillDefaultLevel(h, def, excludes)
		if level == fxp.Min {
			continue
		}
		if best < level {
			best = level
			bestDef = def.CloneWithoutLevelOrPoints()
			bestDef.Level = level
		}
		if keepCurrent && def.Equivalent(replacements, h.recordedDefault()) {
			preferredDef = def.CloneWithoutLevelOrPoints()
			preferredDef.Level = level
		}
	}
	if preferredDef != nil {
		return preferredDef
	}
	return bestDef
}

// resolvableDefaults returns the distinct, currently-resolvable defaults the item could use, ordered from the highest
// resulting level to the lowest. Candidates that currently default back to the item are included: the user may choose
// one, and swapToNextDefault breaks the resulting cycle so the chosen item re-resolves to its own best default, and
// such a candidate's level is taken as that leaves it; see chosenDefaultLevel.
func resolvableDefaults(h defaultable) []*SkillDefault {
	if entityOf(h) == nil {
		return nil
	}
	replacements := h.NameableReplacements()
	excludes := map[string]bool{h.excludeKey(): true}
	var result []*SkillDefault
	for _, def := range resolveToSpecificDefaults(h) {
		level := chosenDefaultLevel(h, def, excludes)
		if level == fxp.Min {
			continue
		}
		dup := false
		for _, existing := range result {
			if existing.Equivalent(replacements, def) {
				dup = true
				break
			}
		}
		if dup {
			continue
		}
		clone := def.CloneWithoutLevelOrPoints()
		clone.Level = level
		result = append(result, clone)
	}
	slices.SortStableFunc(result, func(a, b *SkillDefault) int {
		return cmp.Compare(b.Level, a.Level)
	})
	return result
}

// chosenDefaultLevel returns the level the default gives the item, with the levels taken with the excludes, once
// swapToNextDefault has chosen it: when the default leads back to the item, the item it selects drops the default it
// recorded and picks afresh the best of its defaults, passing over those that now lead back to it through the item, so
// that its level no longer holds what the item gave it.
func chosenDefaultLevel(h defaultable, def *SkillDefault, excludes map[string]bool) fxp.Int {
	if inDefaultChain(h, def, make(map[defaultable]bool)) {
		if target := defaultTargetOf(entityOf(h), h, def, true); target != nil && target != h {
			// The defaults are recorded as the swap leaves them only while the level is found.
			saved, savedTarget := h.recordedDefault(), target.recordedDefault()
			defer func() {
				h.recordDefault(saved)
				target.recordDefault(savedTarget)
			}()
			h.recordDefault(def)
			target.recordDefault(nil)
			target.recordDefault(bestDefaultWithPoints(target, nil))
		}
	}
	return calcSkillDefaultLevel(h, def, excludes)
}

// alternateDefaultsAvailable returns true if the item has more than one default it could use, allowing the user to
// choose among them via "Swap Defaults". Only items that declare their own defaults are swappable; purely synthetic
// defaults track the best default automatically.
func alternateDefaultsAvailable(h defaultable) bool {
	return h.CanSwapDefaults() && len(h.declaredDefaults()) > 0 && len(resolvableDefaults(h)) > 1
}

// CanSwapDefaultsOf returns true if "Swap Defaults" can change the default of the skill or spell: it can swap its own
// (see CanSwapDefaults), and either has more than one default it could use, or could trade places with the skill or
// spell it defaults to, or one defaulting to it, so that the other defaults to it instead (see canTradePlaces). A
// technique, a container, an item a grant put on the sheet, one without points and a Ritual Magic spell can't swap
// their own.
func CanSwapDefaultsOf[T *Skill | *Spell](item T) bool {
	h := defaultable(item)
	return h.CanSwapDefaults() && (alternateDefaultsAvailable(h) || canSwapDefaultsWith(h, defaultTarget(h)) ||
		bestSwappablePartner(h) != nil)
}

// SwapDefaultsOf changes the default of each of the skills or spells in turn as "Swap Defaults" does, when it can swap
// its own: to the next of its defaults, when it has more than one it could use, or otherwise, so that two items
// defaulting one to the other trade places, it drops its default and either the skill or spell it defaulted to, when
// the two could trade places (see canTradePlaces), takes its default to it instead (see tradePlaces), or else the best
// one defaulting to it that could picks its own default afresh (see swapDefaults). Each swap leaves alone the items an
// earlier one was made for, along with the other skill or spell whose default each of those changed: such an item is
// passed over, as is one whose swap would change the default of such an item, since it would undo the earlier swap, as
// when both of two items defaulting to each other are given.
func SwapDefaultsOf[T *Skill | *Spell](items ...T) {
	changed := make(map[defaultable]bool)
	for _, item := range items {
		swapDefaultsOf(defaultable(item), changed)
	}
}

// swapDefaultsOf is SwapDefaultsOf for one item, unless it or the other skill or spell its swap would change the
// default of is in changed, to which it adds both once it has swapped.
func swapDefaultsOf(h defaultable, changed map[defaultable]bool) {
	if changed[h] || !h.CanSwapDefaults() {
		return
	}
	var other defaultable
	var swap func()
	if alternateDefaultsAvailable(h) {
		// Moving to the next default clears the default of the item it selects, when that leads back; see
		// swapToNextDefault.
		if next := nextDefault(h); inDefaultChain(h, next, make(map[defaultable]bool)) {
			if target := defaultTargetOf(entityOf(h), h, next, true); target != h {
				other = target
			}
		}
		swap = func() { swapToNextDefault(h) }
	} else if target := defaultTarget(h); canSwapDefaultsWith(h, target) {
		other = target
		swap = func() { tradePlaces(h, target) }
	} else if partner := bestSwappablePartner(h); partner != nil {
		other = partner
		swap = func() {
			h.recordDefault(nil)
			swapDefaults(partner)
		}
	} else {
		return
	}
	if other != nil && changed[other] {
		return
	}
	swap()
	changed[h] = true
	if other != nil {
		changed[other] = true
	}
}

// canSwapDefaultsWith returns true if the item can swap its own default (see CanSwapDefaults) and could trade places
// with other, the skill or spell it defaults to (see canTradePlaces). A nil other has none.
func canSwapDefaultsWith(h, other defaultable) bool {
	return other != nil && h.CanSwapDefaults() && canTradePlaces(h, other)
}

// bestSwappablePartner returns the skill or spell of the highest level on the item's sheet that defaults to it and that
// it could trade places with (see canTradePlaces), whether or not that skill or spell can swap its own, when the item
// can swap its own (see CanSwapDefaults), or nil.
func bestSwappablePartner(h defaultable) defaultable {
	e := entityOf(h)
	if e == nil || !h.CanSwapDefaults() {
		return nil
	}
	var best defaultable
	var bestLevel fxp.Int
	consider := func(other defaultable) {
		if defaultTarget(other) == h && canTradePlaces(other, h) {
			if level := other.levelExcluding(nil).Level; best == nil || bestLevel < level {
				best = other
				bestLevel = level
			}
		}
	}
	Traverse(func(sk *Skill) bool {
		consider(sk)
		return false
	}, true, true, e.Skills...)
	Traverse(func(sp *Spell) bool {
		consider(sp)
		return false
	}, true, true, e.Spells...)
	return best
}

// canTradePlaces returns true if lower, a skill or spell defaulting to upper, could trade places with it, so that
// upper's level rests on lower instead and stays there: neither is an item a grant put on the sheet, since such a one
// changes only with its grant, and, were lower to record the best of its other defaults and upper its default to lower,
// as a swap leaves them, the default upper then picks (see bestDefaultWithPoints) would select lower and be one upper's
// level rests on (see defaultable.levelRestsOn). It would not where lower has no level without upper, nor where upper
// is a Ritual Magic spell, which always takes its best default, that would take another or whose ritual skills would
// outrank its default to lower; the next recalculation would put the two back where they were.
func canTradePlaces(lower, upper defaultable) bool {
	e := entityOf(upper)
	if e == nil || IsGranted(lower) || IsGranted(upper) {
		return false
	}
	toLower := defaultTo(upper, lower)
	if toLower == nil {
		return false
	}
	// The defaults are recorded as a swap leaves them only while the one upper picks is found.
	savedLower, savedUpper := lower.recordedDefault(), upper.recordedDefault()
	defer func() {
		lower.recordDefault(savedLower)
		upper.recordDefault(savedUpper)
	}()
	lower.recordDefault(bestDefaultWithPoints(lower, savedLower))
	upper.recordDefault(toLower)
	def := bestDefaultWithPoints(upper, nil)
	return def != nil && defaultTargetOf(e, upper, def, true) == lower && upper.levelRestsOn(def)
}

// tradePlaces has lower, a skill or spell defaulting to upper that could trade places with it (see canTradePlaces), do
// so: lower drops its default and upper records its default to lower, and the levels of both are updated, lower's
// first, so that lower picks afresh the best of its other defaults, the one to upper now leading back to it.
func tradePlaces(lower, upper defaultable) {
	lower.recordDefault(nil)
	upper.recordDefault(defaultTo(upper, lower))
	lower.UpdateLevel()
	upper.UpdateLevel()
}

// nextDefault returns the default after the one the item uses among those it could use (see resolvableDefaults),
// wrapping around, or nil when it has fewer than two.
func nextDefault(h defaultable) *SkillDefault {
	candidates := resolvableDefaults(h)
	if len(candidates) < 2 {
		return nil
	}
	idx := -1
	if current := h.recordedDefault(); current != nil {
		replacements := h.NameableReplacements()
		for i, def := range candidates {
			if def.Equivalent(replacements, current) {
				idx = i
				break
			}
		}
	}
	return candidates[(idx+1)%len(candidates)]
}

// swapToNextDefault changes the default the item uses to the next available one (see nextDefault), preserving the
// choice until the user changes it again.
func swapToNextDefault(h defaultable) {
	chosen := nextDefault(h)
	if chosen == nil {
		return
	}
	h.recordDefault(chosen)
	// If the newly chosen item currently defaults back to this one (directly or transitively), clear its default so
	// this explicit choice wins; the recalculation below resolves it to its own best remaining default, keeping the two
	// from defaulting to each other.
	if inDefaultChain(h, chosen, make(map[defaultable]bool)) {
		if target := defaultTarget(h); target != nil && target != h {
			target.recordDefault(nil)
		}
	}
	if e := entityOf(h); e != nil {
		e.Recalculate()
	}
}

// swapDefaults clears the default the item recorded and, when its best default rests on another item, records the best
// of its defaults other than the one it had and updates the levels of both.
func swapDefaults(h defaultable) {
	def := h.recordedDefault()
	h.recordDefault(nil)
	if e := entityOf(h); e != nil {
		if target := defaultTargetOf(e, h, bestDefault(h, nil), true); target != nil {
			h.recordDefault(bestDefaultWithPoints(h, def))
			target.UpdateLevel()
			h.UpdateLevel()
		}
	}
}

// calcSkillDefaultLevel returns the level the default gives the item, with the levels taken with the excludes, or
// fxp.Min when it doesn't resolve: without the bonus the item a skill- or spell-based default rests on has from the
// bonuses that select it (see SkillDefault.skillLevelLess), and with the rule of 20 for an item that picks its default.
func calcSkillDefaultLevel(h defaultable, def *SkillDefault, excludes map[string]bool) fxp.Int {
	return def.skillLevelLess(entityOf(h), h.NameableReplacements(), true, excludes, h.picksDefault(), true)
}

// inDefaultChain returns true if the item is among those the default selects, or those the recorded defaults their
// levels rest on (see defaultable.restingDefault) select in turn, skills and spells alike, passing over the items in
// lookedAt, to which those it walks through are added.
func inDefaultChain(h defaultable, def *SkillDefault, lookedAt map[defaultable]bool) bool {
	e := entityOf(h)
	if e == nil || def == nil || !def.SkillOrSpellBased() {
		return false
	}
	for _, one := range def.matchingTargets(e, h.NameableReplacements(), true, nil) {
		if one == h {
			return true
		}
		if !lookedAt[one] {
			lookedAt[one] = true
			if inDefaultChain(h, one.restingDefault(), lookedAt) {
				return true
			}
		}
	}
	return false
}

// resolveToSpecificDefaults returns the item's declared defaults, each one whose power source is to be the same as
// the item's taking the item's (see SkillDefault.withPowerSourceOf), and each skill- or spell-based one replaced on a
// sheet by a copy for every item it selects, pinned to that item, followed by its synthetic defaults.
func resolveToSpecificDefaults(h defaultable) []*SkillDefault {
	e := entityOf(h)
	declared := h.declaredDefaults()
	replacements := h.NameableReplacements()
	result := make([]*SkillDefault, 0, len(declared))
	for _, def := range declared {
		if def == nil {
			continue
		}
		if def.SamePowerSource {
			def = def.withPowerSourceOf(h.ownPowerSource())
		}
		if e == nil || !def.SkillOrSpellBased() {
			result = append(result, def)
			continue
		}
		for _, one := range def.matchingTargets(e, replacements, true, map[string]bool{h.excludeKey(): true}) {
			local := *def
			// The copy is pinned to the matched item by name, a skill by specialization and a spell by power source, so
			// the tag criteria has already done its job. Leaving it on would carry it into the persisted recorded
			// default and make an expansion of a tag default look different from a plain default to the same item, so
			// that neither the swap list nor the "keep the current default" check could pair them.
			local.Tags = criteria.Text{}
			local.Name.Compare = criteria.IsText
			local.Name.Qualifier = one.NameWithReplacements()
			switch target := one.(type) {
			case *Skill:
				local.Specialization.Compare = criteria.IsText
				if spec := target.SpecializationWithReplacements(); spec != "" {
					local.Specialization.Qualifier = spec
				} else {
					local.Specialization.Qualifier = target.OptionalSpecializationWithReplacements()
				}
				if target.OptionalSpecializationWithReplacements() != "" &&
					def.Specialization.Matches(replacements, target.SpecializationWithReplacements()) {
					local.Modifier -= fxp.Two
				}
			case *Spell:
				local.SamePowerSource = false
				local.PowerSource = criteria.Text{Compare: criteria.IsText, Qualifier: target.PowerSourceWithReplacements()}
			}
			result = append(result, &local)
		}
	}
	return append(result, h.syntheticDefaults()...)
}

// defaultTargetOf returns the item whose level the default, one of h's, uses on the entity: the best one it selects
// whose tech level satisfies its WhenTL, or nil.
func defaultTargetOf(e *Entity, h defaultable, def *SkillDefault, requirePoints bool) defaultable {
	if e == nil || def == nil || !def.SkillOrSpellBased() {
		return nil
	}
	if def.SamePowerSource {
		def = def.withPowerSourceOf(h.ownPowerSource())
	}
	target, _ := def.best(e, h.NameableReplacements(), requirePoints, nil)
	return target
}

// defaultTarget returns the item h currently defaults to, or nil: the one the default it recorded uses or, for a
// technique, which picks none, the skill its own default uses.
func defaultTarget(h defaultable) defaultable {
	def := h.recordedDefault()
	if sk, ok := h.(*Skill); ok && sk.IsTechnique() {
		if !sk.restsOnSkill() {
			return nil
		}
		def = sk.TechniqueDefault
	}
	return defaultTargetOf(entityOf(h), h, def, true)
}

// hasDefaultTo returns true if the set of possible defaults of the item includes the target; see defaultTo.
func hasDefaultTo(h, target defaultable) bool {
	return defaultTo(h, target) != nil
}

// defaultTo returns the first of the item's possible defaults (see resolveToSpecificDefaults) to the target, or nil:
// for a skill, a skill-based default naming it and selecting its specialization, and for a spell, a spell-based default
// naming it and selecting its power source.
func defaultTo(h, target defaultable) *SkillDefault {
	replacements := h.NameableReplacements()
	for _, def := range resolveToSpecificDefaults(h) {
		if def.NameWithReplacements(replacements) != target.NameWithReplacements() {
			continue
		}
		switch one := target.(type) {
		case *Skill:
			if def.SkillBased() && skillSpecializationMatches(def.Specialization, replacements,
				one.SpecializationWithReplacements(), one.OptionalSpecializationWithReplacements()) {
				return def
			}
		case *Spell:
			if def.SpellBased() && !def.SamePowerSource &&
				def.PowerSource.Matches(replacements, one.PowerSourceWithReplacements()) {
				return def
			}
		}
	}
	return nil
}

// defaultNote returns the note naming what the item defaults to, such as "Default: Broadsword-2", or "" when its
// recorded default rests on nothing.
func defaultNote(h defaultable) string {
	def := h.recordedDefault()
	if def == nil {
		return ""
	}
	target := defaultTarget(h)
	if target == nil {
		return ""
	}
	return i18n.Text("Default: ") + target.String() + def.ModifierAsString()
}

// defaultExcept returns the default the item recorded or, when exclude (nil, or the item whose prerequisites are being
// judged; see Entity.judgingOr) carries a grant and the item picks its default, a default picked afresh as those
// prerequisites see the sheet, since the recorded one, or the lack of one, was picked, and its level taken, with what
// the grants did to the items it rests on. It is picked afresh too while an editor previews the level of a skill or
// spell, since the recorded one was picked with that item as it is on the sheet rather than as edited (see
// Entity.matchLevelCapsAs), at base, since it was picked, and its level taken, with what alters current values alone
// (see Entity.atBaseWhile), and while a pool's base or thresholds are found without the features those thresholds hold,
// since it was picked, and its level taken, with what they add (see Entity.thresholdFeaturesLeftOut). It is picked as
// bestDefault picks, but with the levels taken with the item added to the excludes given rather than with it alone
// excluded, so that a chain of defaults never comes back to an item it passes through: the recorded default, if the
// item keeps it (see keepsDefault), it resolves and it does not lead back to the item, and otherwise the best of the
// item's defaults. Each pick is kept while those prerequisites are judged, the edits previewed, the computation at base
// runs, or the pool's base or thresholds are found, apart from those found otherwise, and reused only for the same item
// with the same excludes and the same view of the sheet, which leaves out what the thresholds do while a pool's
// threshold is found for those prerequisites, and the features they hold while a pool's base or thresholds are found
// (see judgedDefaultKey), so that the answers are those of picking afresh every time, while a pick beneath an item
// weighing several defaults is not made again for each of them, and so on down every chain. The picks are still as many
// as the distinct excludes an item is reached with, which double with each member of a set of items with points that
// default to one another, so judging stays fast only while such a set stays small.
func defaultExcept(h defaultable, exclude any, excludes map[string]bool) *SkillDefault {
	e := entityOf(h)
	atBase := e != nil && e.atBase > 0
	recorded := h.recordedDefault()
	if !h.picksDefault() || (!e.carriesGrant(e.judgingOr(exclude)) && !e.previewingEdit() && !atBase &&
		!e.thresholdFeaturesLeftOut()) {
		return recorded
	}
	excludes = maps.Clone(excludes)
	if excludes == nil {
		excludes = make(map[string]bool)
	}
	excludes[h.excludeKey()] = true
	key := newJudgedDefaultKey(h, excludes)
	baseKey := baseDefaultKey{judgedDefaultKey: key, judged: e.judgingOr(exclude)}
	if atBase {
		if def, ok := e.baseDefaults[baseKey]; ok {
			return def
		}
	} else if def, ok := e.judgedDefaults[key]; ok {
		return def
	}
	var def, excluded *SkillDefault
	if keepsDefault(h) && !inDefaultChain(h, recorded, make(map[defaultable]bool)) {
		if level := calcSkillDefaultLevel(h, recorded, excludes); level != fxp.Min {
			def = recorded.CloneWithoutLevelOrPoints()
			def.Level = level
		} else {
			excluded = recorded
		}
	}
	if def == nil {
		def = bestDefaultWith(h, excluded, excludes)
	}
	if def != nil {
		setDefaultPoints(h, def)
	}
	if atBase {
		if e.baseDefaults == nil {
			e.baseDefaults = make(map[baseDefaultKey]*SkillDefault)
		}
		e.baseDefaults[baseKey] = def
	} else {
		if e.judgedDefaults == nil {
			e.judgedDefaults = make(map[judgedDefaultKey]*SkillDefault)
		}
		e.judgedDefaults[key] = def
	}
	return def
}

// judgedDefaultKey identifies the default defaultExcept picks for an item with the levels taken with a set of
// excludes, which it holds as their names sorted and joined, whether it is picked while a pool's threshold is found
// for the prerequisites being judged, which see the sheet without what the thresholds do meanwhile (see
// Attribute.CurrentThreshold), and whether it is picked while a pool's base or thresholds are found without the
// features those thresholds hold (see Entity.thresholdFeaturesLeftOut).
type judgedDefaultKey struct {
	item      defaultable
	excludes  string
	searching bool
	untouched bool
}

// baseDefaultKey identifies the default defaultExcept picks for an item at base, which rests as well on the item whose
// prerequisites are being judged, if any.
type baseDefaultKey struct {
	judgedDefaultKey
	judged any
}

// newJudgedDefaultKey returns the key of the default the item picks with the levels taken with the given excludes, as
// the sheet is seen now.
func newJudgedDefaultKey(h defaultable, excludes map[string]bool) judgedDefaultKey {
	e := entityOf(h)
	return judgedDefaultKey{
		item:      h,
		excludes:  excludesKey(excludes),
		searching: e != nil && e.judgedThresholdSearches > 0,
		untouched: e.thresholdFeaturesLeftOut(),
	}
}

// excludesKey returns the excludes as a string that is the same for any two sets of the same names: the names sorted
// and joined.
func excludesKey(excludes map[string]bool) string {
	names := make([]string, 0, len(excludes))
	for name, excluded := range excludes {
		if excluded {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return strings.Join(names, "\x00")
}
