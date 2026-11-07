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
	"regexp"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// currentAdjustment is what the bonuses that alter a value's current value alone do to it: the amounts the current
// thresholds of the pools add to it, and the factors it is multiplied by, from those thresholds or from the items in
// effect. Its zero value changes nothing.
type currentAdjustment struct {
	bonus   fxp.Int
	factors []fxp.Int
}

// add folds the bonus in: a multiplier as a factor, anything else as an amount.
func (c *currentAdjustment) add(b *AttributeBonus) {
	if b.Multiply {
		c.factors = append(c.factors, b.Amount)
	} else {
		c.bonus += b.AdjustedAmount()
	}
}

// apply returns the value with the amount added and then multiplied by each factor in turn. With whole set, for a
// value that holds whole numbers, the amount is floored first, as the other bonuses such a value takes are, and each
// product is rounded to the nearest whole number, so that halving twice rounds up twice, as the rules do (B419).
func (c currentAdjustment) apply(value fxp.Int, whole bool) fxp.Int {
	return c.applyRounding(value, whole, whole)
}

// applyRounding is apply with the amount floored only when wholeAmount is set and each product rounded to the nearest
// whole number only when roundProducts is set.
func (c currentAdjustment) applyRounding(value fxp.Int, wholeAmount, roundProducts bool) fxp.Int {
	bonus := c.bonus
	if wholeAmount {
		bonus = bonus.Floor()
	}
	value += bonus
	for _, factor := range c.factors {
		value = value.Mul(factor)
		if roundProducts {
			value = value.Round()
		}
	}
	return value
}

// adjustmentKey identifies what a currentAdjustment applies to: an attribute ID, or MoveID, DodgeID, ParryID or
// BlockID standing in for one, along with the ST limitation; see AttributeBonus.ActualLimitation.
type adjustmentKey struct {
	id         string
	limitation stlimit.Option
}

// forEachThresholdFeatureList calls fn with the features the current threshold of each pool applies, leaving out those
// ThresholdIgnores flags, in the order the sheet lists the pools, along with the threshold and the pool's definition,
// for each pool whose threshold applies any. A pool's current threshold is only found when one of its thresholds has
// features, since finding it runs their scripts. While the prerequisites of an item carrying a grant are judged, the
// thresholds are those they see; see Attribute.CurrentThreshold. A threshold has no switch, so whether its features
// are switchable is ignored.
func (e *Entity) forEachThresholdFeatureList(fn func(t *PoolThreshold, def *AttributeDef, list Features)) {
	var defs *AttributeDefs
	for _, attr := range e.Attributes.List() {
		def := attr.AttributeDef()
		if def == nil || (def.Type != attribute.Pool && def.Type != attribute.PoolRef) ||
			!slices.ContainsFunc(def.Thresholds, func(t *PoolThreshold) bool { return len(t.Features) != 0 }) {
			continue
		}
		t := attr.CurrentThreshold()
		if t == nil || len(t.Features) == 0 {
			continue
		}
		if defs == nil {
			defs = AttributeDefsFor(e)
		}
		list := t.Features
		ignored := func(f Feature) bool { return ThresholdIgnores(defs, f) }
		if slices.ContainsFunc(list, ignored) {
			list = slices.DeleteFunc(slices.Clone(list), ignored)
		}
		if len(list) != 0 {
			fn(t, def, list)
		}
	}
}

// processThresholdFeatures collects the features of the pools' current thresholds once those of the items are in place,
// since a pool's maximum, and so its threshold, rests on the bonuses the items give, and records what those thresholds
// and every multiplier do to each value; see recordCurrentAdjustments. The thresholds are found while the values carry
// what the items' multipliers do to them alone, so that a pool whose base or thresholds read a value is found with the
// value the sheet shows unless the thresholds alter it, in which case it is found without that alteration rather than
// with what the previous pass left. The caches are discarded before the thresholds are found and once the adjustments
// are in place, since what the scripts cached was computed without them. Found outside a recalculation, a pool whose
// thresholds alter what it reads may be found at another threshold. A trait bonus a threshold holds raises the level
// the per-level bonuses of the trait read, whenever they are read, save those given a pool or the size modifier (see
// AttributeBonusFor), so the totals totalBonuses recorded are recorded again once one is collected. The parry and
// block tooltips are built here, so that they name these bonuses too. A grant a threshold holds is collected here as
// well, and so is in effect while the pool is at the threshold; the rows it puts on the sheet are there whether or not
// it is, as for any grant (see forEachGrant).
func (e *Entity) processThresholdFeatures() {
	if e.recordCurrentAdjustments() {
		e.discardCaches(true)
	}
	traitBonuses := len(e.features.traitBonuses)
	e.collectThresholdFeatures(&e.features)
	if len(e.features.traitBonuses) != traitBonuses {
		e.totalBonuses()
	}
	e.recordCurrentAdjustments()
	var tooltip xbytes.InsertBuffer
	e.AttributeBonusFor(ParryID, stlimit.None, &tooltip)
	e.currentAdjustmentFor(ParryID, stlimit.None, &tooltip)
	e.ParryBonusTooltip = tooltip.String()
	tooltip.Reset()
	e.AttributeBonusFor(BlockID, stlimit.None, &tooltip)
	e.currentAdjustmentFor(BlockID, stlimit.None, &tooltip)
	e.BlockBonusTooltip = tooltip.String()
	e.discardCaches(true)
}

// recordCurrentAdjustments records what the bonuses collected so far that alter current values alone do to each value,
// passing over, as AttributeBonusFor does, those bonusExcluded leaves out, and returns whether they do anything; see
// currentAdjustment.
func (e *Entity) recordCurrentAdjustments() bool {
	adjustments := make(map[adjustmentKey]currentAdjustment)
	for _, one := range e.features.currentAttributeBonuses {
		if e.bonusExcluded(one) {
			continue
		}
		key := adjustmentKey{id: one.Attribute, limitation: one.ActualLimitation()}
		adj := adjustments[key]
		adj.add(one)
		adjustments[key] = adj
	}
	e.currentAdjustments = adjustments
	return len(adjustments) != 0
}

// currentAdjustmentFor returns what the bonuses that alter current values alone do to the value with the given ID and
// limitation (see currentAdjustment), adding each to the tooltip, which may be nil, and passing over those
// bonusExcluded leaves out. While the prerequisites of an item carrying a grant are judged, the bonuses of the
// thresholds are taken from the thresholds as those prerequisites see the pools, and none of theirs while those
// thresholds are being found, as the sheet's own are found with the items' multipliers alone; see bonusFeatures,
// Attribute.CurrentThreshold and processThresholdFeatures. A computation at base gets none at all; see atBaseWhile.
func (e *Entity) currentAdjustmentFor(id string, limitation stlimit.Option, tooltip *xbytes.InsertBuffer) currentAdjustment {
	var adj currentAdjustment
	if e.atBase > 0 {
		return adj
	}
	for _, one := range e.bonusFeatures().currentAttributeBonuses {
		if one.Attribute == id && one.ActualLimitation() == limitation && !e.bonusExcluded(one) {
			adj.add(one)
			one.AddToTooltip(tooltip)
		}
	}
	return adj
}

// collectThresholdFeatures collects into the feature lists the features the pools' current thresholds apply; see
// forEachThresholdFeatureList.
func (e *Entity) collectThresholdFeatures(into *features) {
	e.forEachThresholdFeatureList(func(t *PoolThreshold, def *AttributeDef, list Features) {
		for _, f := range list {
			e.processFeature(into, t, def, f, nil)
		}
	})
}

// The views of the sheet's feature lists that bonusFeatures builds while the prerequisites of an item carrying a grant
// are judged; see Entity.judgedFeatures.
const (
	featuresWhileSearching = iota
	featuresAsJudged
	featuresAtBase
	featureViews
)

// bonusFeatures returns the feature lists the bonuses are read from: those processFeatures collected, or, while the
// prerequisites of an item carrying a grant are judged, those with the features of the thresholds the sheet's pools are
// at replaced by those of the thresholds as those prerequisites see the pools, so that the prerequisites never see a
// threshold that what the grant put on the sheet brought the sheet to, along with those of the items the grants of
// those thresholds put in effect (see collectGrantedFeaturesAsJudged), or with none at all while those thresholds are
// being found, as the sheet's own are found without them; see Attribute.CurrentThreshold. A computation at base sees
// none either, since they alter only the sheet's current state, nor those of the items the thresholds' grants alone put
// in effect; see atBaseWhile and featuresAtBase. Each is built once while the same prerequisites are judged, or the
// same computation at base runs, since their view of the sheet holds still meanwhile; see JudgingPrereqsOf. Nothing
// reads the grants they hold, which are not those the sheet applies: the sheet's are left out, and those the view as
// judged collects along with the features of the thresholds and of the items their grants give go unused.
func (e *Entity) bonusFeatures() *features {
	if !e.carriesGrant(e.judging) {
		if e.atBase == 0 {
			return &e.features
		}
		if e.baseFeatures == nil {
			base := e.featuresAtBase()
			e.baseFeatures = &base
		}
		return e.baseFeatures
	}
	view := featuresAsJudged
	switch {
	case e.judgedThresholdSearches > 0:
		view = featuresWhileSearching
	case e.atBase > 0:
		view = featuresAtBase
	}
	if judged := e.judgedFeatures[view]; judged != nil {
		return judged
	}
	var judged features
	if view == featuresAtBase {
		judged = e.featuresAtBase()
	} else {
		judged = featuresWithout(&e.features, ownedByThreshold)
	}
	judged.grants = nil
	if view == featuresAsJudged {
		e.collectThresholdFeatures(&judged)
		// Finding what the grants of the thresholds give may read bonuses, which then see the lists as built so far.
		e.judgedFeatures[view] = &judged
		e.collectGrantedFeaturesAsJudged(&judged)
	}
	e.judgedFeatures[view] = &judged
	return &judged
}

// featuresAtBase returns a copy of the sheet's feature lists as a computation at base sees them (see atBaseWhile):
// without the features of the pools' thresholds, nor those of the items in effect only because the thresholds' grants
// put them there; see rowsGrantedByThresholdsAlone. Each list is clipped, so that adding to the copy never writes into
// the original.
func (e *Entity) featuresAtBase() features {
	rows := e.rowsGrantedByThresholdsAlone()
	return featuresWithout(&e.features, func(owner fmt.Stringer) bool {
		if ownedByThreshold(owner) {
			return true
		}
		holder, ok := owner.(grantHolder)
		return ok && rows[holder]
	})
}

// rowsGrantedByThresholdsAlone returns the items grants put on the sheet whose features the sheet's lists hold that a
// grant of a pool's threshold puts in effect and no other grant gave anything when the lists were collected, since one
// that gives an item nothing, such as a level set below what the points of its first level buy, leaves it out of effect
// without the threshold's (see grantPlan.reaches). What the other grants gave is taken from the lists rather than
// priced afresh, since pricing a level set may read the bonuses at base, which rest on this; see
// features.plantedElsewhere.
func (e *Entity) rowsGrantedByThresholdsAlone() map[grantHolder]bool {
	var rows map[grantHolder]bool
	for i := range e.grantPlans {
		plan := &e.grantPlans[i]
		row := plan.row()
		if !ownedByThreshold(plan.source.owner) || !e.features.grantedRows[row] || e.features.plantedElsewhere[row] ||
			!plan.grant.putsItemOnSheet() || !e.grantInEffect(plan) {
			continue
		}
		if rows == nil {
			rows = make(map[grantHolder]bool)
		}
		rows[row] = true
	}
	return rows
}

// ownedByThreshold returns true if the owner of a feature is a pool's threshold.
func ownedByThreshold(owner fmt.Stringer) bool {
	_, ok := owner.(*PoolThreshold)
	return ok
}

// featuresWithout returns a copy of the feature lists without the features whose owners dropped reports. Each list is
// clipped, so that adding to the copy never writes into the original.
func featuresWithout(list *features, dropped func(owner fmt.Stringer) bool) features {
	result := *list
	result.attributeBonuses = withoutFeaturesOf(result.attributeBonuses, dropped)
	result.costReductions = withoutFeaturesOf(result.costReductions, dropped)
	result.drBonuses = withoutFeaturesOf(result.drBonuses, dropped)
	result.maxUsesBonuses = withoutFeaturesOf(result.maxUsesBonuses, dropped)
	result.skillBonuses = withoutFeaturesOf(result.skillBonuses, dropped)
	result.skillPointBonuses = withoutFeaturesOf(result.skillPointBonuses, dropped)
	result.spellBonuses = withoutFeaturesOf(result.spellBonuses, dropped)
	result.spellPointBonuses = withoutFeaturesOf(result.spellPointBonuses, dropped)
	result.traitBonuses = withoutFeaturesOf(result.traitBonuses, dropped)
	result.traitMaxLevelBonuses = withoutFeaturesOf(result.traitMaxLevelBonuses, dropped)
	result.weaponBonuses = withoutFeaturesOf(result.weaponBonuses, dropped)
	result.selectorOverrides = withoutFeaturesOf(result.selectorOverrides, dropped)
	result.currentAttributeBonuses = withoutFeaturesOf(result.currentAttributeBonuses, dropped)
	return result
}

// collectGrantedFeaturesAsJudged collects into the feature lists the features of the items grants put on the sheet
// that are in effect as the prerequisites being judged see the pools' thresholds, when that differs from what the sheet
// has (see Entity.grantsAsJudged), and whose features the lists don't hold already, as processFeatures collects those
// of the items in effect. The features of those the lists hold that are out of effect as those prerequisites see it
// are left out as they are read instead; see grantedOutOfEffectAsJudged.
func (e *Entity) collectGrantedFeaturesAsJudged(into *features) {
	if e.thresholdGrantsAsJudged() == nil {
		return
	}
	seen := make(map[grantHolder]bool)
	for i := range e.grantPlans {
		row := e.grantPlans[i].row()
		if seen[row] || into.grantedRows[row] || !IsGranted(row) {
			continue
		}
		seen[row] = true
		if e.grantsAsJudged(row) == row.grants() {
			continue
		}
		switch one := row.(type) {
		case *Trait:
			if one.Enabled() {
				forEachActiveTraitFeatureList(one, func(mod fmt.Stringer, leveled LeveledOwner, list Features) {
					for _, f := range list {
						e.processFeature(into, one, mod, f, leveled)
					}
				})
				e.collectMerchantPenalty(into, one)
			}
		case *Skill:
			if one.Enabled() {
				for _, f := range one.ActiveFeatures() {
					e.processFeature(into, one, nil, f, one)
				}
			}
		case *Spell:
			if one.Enabled() {
				for _, f := range one.ActiveFeatures() {
					e.processFeature(into, one, nil, f, one)
				}
			}
		}
	}
}

// withoutFeaturesOf returns the list without the features whose owners dropped reports, clipped, so that adding to what
// it returns never writes into the list.
func withoutFeaturesOf[T interface{ Owner() fmt.Stringer }](list []T, dropped func(owner fmt.Stringer) bool) []T {
	drop := func(one T) bool { return dropped(one.Owner()) }
	if !slices.ContainsFunc(list, drop) {
		return slices.Clip(list)
	}
	return slices.DeleteFunc(slices.Clone(list), drop)
}

// sheetAdjustment returns what the current thresholds and multipliers do to the value with the given ID and
// limitation, as processThresholdFeatures recorded it, which is nothing for a value they leave alone, or for any value
// to a computation at base; see atBaseWhile.
func (e *Entity) sheetAdjustment(id string, limitation stlimit.Option) currentAdjustment {
	if e == nil || e.currentAdjustments == nil || e.atBase > 0 {
		return currentAdjustment{}
	}
	return e.currentAdjustments[adjustmentKey{id: id, limitation: limitation}]
}

// adjustmentFor is sheetAdjustment, except while the prerequisites of an item carrying a grant are judged, when it is
// what currentAdjustmentFor gives those prerequisites.
func (e *Entity) adjustmentFor(id string, limitation stlimit.Option) currentAdjustment {
	if e.carriesGrant(e.judging) {
		return e.currentAdjustmentFor(id, limitation, nil)
	}
	return e.sheetAdjustment(id, limitation)
}

// CurrentAdjustmentTooltip returns a line for each bonus that goes into the current value of the attribute with the
// given ID, or of the Move, Dodge, Parry or Block standing in for one, under the given ST limitation, such as
// "Tired (FP) [×0.5]": those the items add to it, then those that alter only its current value, then, under a heading
// naming each, those that alter only the current values of the attributes it is figured from (see
// addSourceAdjustments). It returns an empty string when there are none.
func (e *Entity) CurrentAdjustmentTooltip(id string, limitation stlimit.Option) string {
	if e == nil {
		return ""
	}
	var buffer xbytes.InsertBuffer
	e.AttributeBonusFor(id, limitation, &buffer)
	e.currentAdjustmentFor(id, limitation, &buffer)
	e.addSourceAdjustments(id, limitation, &buffer)
	return strings.TrimPrefix(buffer.String(), "\n")
}

// addSourceAdjustments adds to the tooltip, which may be nil, the bonuses that alter only the current value of each
// attribute whose current value the value with the given ID and limitation is figured from (see currentSources), and in
// turn of each attribute those are figured from, under a heading naming the attribute. Each attribute is listed once,
// and a pool, which they don't alter, not at all.
func (e *Entity) addSourceAdjustments(id string, limitation stlimit.Option, tooltip *xbytes.InsertBuffer) {
	if tooltip == nil {
		return
	}
	seen := make(map[string]bool)
	if limitation == stlimit.None {
		seen[id] = true
	}
	var walk func(id string, limitation stlimit.Option)
	walk = func(id string, limitation stlimit.Option) {
		for _, source := range e.currentSources(id, limitation) {
			def := e.ResolveAttributeDef(source)
			if seen[source] || def == nil || def.IsSeparator() || def.Type == attribute.Pool ||
				def.Type == attribute.PoolRef {
				continue
			}
			seen[source] = true
			var lines xbytes.InsertBuffer
			e.currentAdjustmentFor(source, stlimit.None, &lines)
			if lines.Len() != 0 {
				tooltip.WriteByte('\n')
				tooltip.WriteString(def.String())
				tooltip.WriteByte(':')
				tooltip.WriteString(lines.String())
			}
			walk(source, stlimit.None)
		}
	}
	walk(id, limitation)
}

// scriptAttributeRef matches an attribute read in a script as "$id", capturing the ID, or as entity.attribute("id"),
// capturing the ID or name given it in whichever kind of quotes holds it (see Attributes.Find), along with the member
// read from it, such as "current" or "maximum", capturing its name when there is one.
var scriptAttributeRef = regexp.MustCompile(`(?:\$(\w+)|\bentity\s*\.\s*attribute\s*\(\s*(?:"([^"]*)"|'([^']*)'|` +
	"`([^`]*)`" + `)\s*\))(?:\s*\??\.\s*(\w+))?`)

// currentSources returns the IDs of the attributes whose current values, rather than their maximums, the value with the
// given ID and limitation is figured from: for ST under a limitation, the attribute standing in for that ST or, without
// one, ST (see strengthFor); for Move and Dodge without an attribute standing in for them, Basic Move and Basic Speed
// (see Move and Dodge); for an attribute, each attribute its base reads as "$id.current" or
// entity.attribute("id").current, and for one other than a pool standing in for any of those, each other than a pool it
// reads as "$id" or entity.attribute("id") too, either of which gives the current value there (see Attribute.Current),
// but not as "$id.maximum" or entity.attribute("id").maximum; and for anything else, none.
func (e *Entity) currentSources(id string, limitation stlimit.Option) []string {
	if limitation != stlimit.None {
		var standIn string
		switch limitation {
		case stlimit.StrikingOnly:
			standIn = StrikingStrengthID
		case stlimit.LiftingOnly:
			standIn = LiftingStrengthID
		default:
			standIn = ThrowingStrengthID
		}
		if e.ResolveAttribute(standIn) != nil {
			return []string{standIn}
		}
		return []string{StrengthID}
	}
	if e.ResolveAttribute(id) == nil {
		switch id {
		case MoveID:
			return []string{BasicMoveID}
		case DodgeID:
			return []string{BasicSpeedID}
		default:
			return nil
		}
	}
	def := e.ResolveAttributeDef(id)
	if def == nil {
		return nil
	}
	readsCurrents := baseReadsCurrents(id) && def.Type != attribute.Pool && def.Type != attribute.PoolRef
	var ids []string
	for _, match := range scriptAttributeRef.FindAllStringSubmatch(def.Base, -1) {
		bare := false
		switch match[5] {
		case "current":
		case "", "valueOf":
			if !readsCurrents {
				continue
			}
			bare = true
		default:
			continue
		}
		var attr *Attribute
		if match[1] != "" {
			attr = e.ResolveAttribute(match[1])
		} else if idOrName := match[2] + match[3] + match[4]; idOrName != "" {
			attr = e.Attributes.Find(idOrName)
		}
		if attr == nil || slices.Contains(ids, attr.AttrID) {
			continue
		}
		if source := attr.AttributeDef(); bare && source != nil &&
			(source.Type == attribute.Pool || source.Type == attribute.PoolRef) {
			continue
		}
		ids = append(ids, attr.AttrID)
	}
	return ids
}
