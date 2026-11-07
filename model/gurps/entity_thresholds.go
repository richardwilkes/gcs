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
	bonus := c.bonus
	if whole {
		bonus = bonus.Floor()
	}
	value += bonus
	for _, factor := range c.factors {
		value = value.Mul(factor)
		if whole {
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

// forEachThresholdFeatureList calls fn with the features of the current threshold of each pool that has any, in the
// order the sheet lists the pools, along with the threshold and the pool's definition. A pool's current threshold is
// only found when one of its thresholds has features, since finding it runs their scripts. While the prerequisites of
// an item carrying a grant are judged, the thresholds are those they see; see Attribute.CurrentThreshold. A threshold
// has no switch, so whether its features are switchable is ignored.
func (e *Entity) forEachThresholdFeatureList(fn func(t *PoolThreshold, def *AttributeDef, list Features)) {
	for _, attr := range e.Attributes.List() {
		def := attr.AttributeDef()
		if def == nil || (def.Type != attribute.Pool && def.Type != attribute.PoolRef) ||
			!slices.ContainsFunc(def.Thresholds, func(t *PoolThreshold) bool { return len(t.Features) != 0 }) {
			continue
		}
		if t := attr.CurrentThreshold(); t != nil && len(t.Features) != 0 {
			fn(t, def, t.Features)
		}
	}
}

// processThresholdFeatures collects the features of the pools' current thresholds once those of the items are in place,
// since a pool's maximum, and so its threshold, rests on the bonuses the items give, and records what those thresholds
// and every multiplier do to each value, passing over, as AttributeBonusFor does, the bonuses bonusExcluded leaves out;
// see currentAdjustment. The thresholds are found while no value carries such an adjustment, so a pool whose base or
// thresholds read a value they alter is found without the alteration rather than with what the previous pass left, and
// the caches are discarded once the adjustments are in place, since what those scripts cached was computed without
// them. Found outside a recalculation, such a pool's threshold may differ. The parry and block tooltips are built here,
// so that they name these bonuses too. A grant a threshold holds is collected here as well, and so is in effect while
// the pool is at the threshold; the rows it puts on the sheet are there whether or not it is, as for any grant (see
// forEachGrant).
func (e *Entity) processThresholdFeatures() {
	e.collectThresholdFeatures(&e.features)
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

// currentAdjustmentFor returns what the bonuses that alter current values alone do to the value with the given ID and
// limitation (see currentAdjustment), adding each to the tooltip, which may be nil, and passing over those
// bonusExcluded leaves out. While the prerequisites of an item carrying a grant are judged, the bonuses of the
// thresholds are taken from the thresholds as those prerequisites see the pools, and none at all, nor any multiplier,
// while those thresholds are being found, as the sheet's own are found without any; see bonusFeatures,
// Attribute.CurrentThreshold and processThresholdFeatures. A computation at base gets none at all; see atBaseWhile.
func (e *Entity) currentAdjustmentFor(id string, limitation stlimit.Option, tooltip *xbytes.InsertBuffer) currentAdjustment {
	var adj currentAdjustment
	if e.atBase > 0 || (e.judgedThresholdSearches > 0 && e.carriesGrant(e.judging)) {
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

// collectThresholdFeatures collects into the feature lists the features of the pools' current thresholds that a
// threshold may hold; see forEachThresholdFeatureList.
func (e *Entity) collectThresholdFeatures(into *features) {
	e.forEachThresholdFeatureList(func(t *PoolThreshold, def *AttributeDef, list Features) {
		for _, f := range list {
			if f.FeatureType().AllowedInThreshold() {
				e.processFeature(into, t, def, f, nil)
			}
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
// none either, since they alter only the sheet's current state; see atBaseWhile. Each is built once while the same
// prerequisites are judged, or the same computation at base runs, since their view of the sheet holds still meanwhile;
// see JudgingPrereqsOf. The grants they hold are not those the sheet applies, and so are left out.
func (e *Entity) bonusFeatures() *features {
	if !e.carriesGrant(e.judging) {
		if e.atBase == 0 {
			return &e.features
		}
		if e.baseFeatures == nil {
			base := featuresWithoutThresholds(&e.features)
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
	judged := featuresWithoutThresholds(&e.features)
	judged.grants = nil
	if view != featuresWhileSearching {
		if view == featuresAsJudged {
			e.collectThresholdFeatures(&judged)
		}
		// Finding what the grants of the thresholds give may read bonuses, which then see the lists as built so far.
		e.judgedFeatures[view] = &judged
		e.collectGrantedFeaturesAsJudged(&judged)
	}
	e.judgedFeatures[view] = &judged
	return &judged
}

// featuresWithoutThresholds returns a copy of the feature lists without the features of the pools' thresholds, each
// list clipped, so that adding to the copy never writes into the original.
func featuresWithoutThresholds(list *features) features {
	result := *list
	result.attributeBonuses = withoutThresholdFeatures(result.attributeBonuses)
	result.costReductions = slices.Clip(result.costReductions)
	result.drBonuses = withoutThresholdFeatures(result.drBonuses)
	result.maxUsesBonuses = withoutThresholdFeatures(result.maxUsesBonuses)
	result.skillBonuses = withoutThresholdFeatures(result.skillBonuses)
	result.skillPointBonuses = withoutThresholdFeatures(result.skillPointBonuses)
	result.spellBonuses = withoutThresholdFeatures(result.spellBonuses)
	result.spellPointBonuses = withoutThresholdFeatures(result.spellPointBonuses)
	result.traitBonuses = withoutThresholdFeatures(result.traitBonuses)
	result.traitMaxLevelBonuses = withoutThresholdFeatures(result.traitMaxLevelBonuses)
	result.weaponBonuses = withoutThresholdFeatures(result.weaponBonuses)
	result.selectorOverrides = withoutThresholdFeatures(result.selectorOverrides)
	result.currentAttributeBonuses = withoutThresholdFeatures(result.currentAttributeBonuses)
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
			if !one.Container() && one.Enabled() {
				for _, f := range one.ActiveFeatures() {
					e.processFeature(into, one, nil, f, one)
				}
			}
		case *Spell:
			if !one.Container() && one.Enabled() {
				for _, f := range one.ActiveFeatures() {
					e.processFeature(into, one, nil, f, one)
				}
			}
		}
	}
}

// withoutThresholdFeatures returns the list without the features a pool's threshold holds, clipped, so that adding to
// what it returns never writes into the list.
func withoutThresholdFeatures[T interface{ Owner() fmt.Stringer }](list []T) []T {
	fromThreshold := func(one T) bool {
		_, ok := one.Owner().(*PoolThreshold)
		return ok
	}
	if !slices.ContainsFunc(list, fromThreshold) {
		return slices.Clip(list)
	}
	return slices.DeleteFunc(slices.Clone(list), fromThreshold)
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
// capturing the ID or name given it in whichever kind of quotes holds it; see Attributes.Find.
var scriptAttributeRef = regexp.MustCompile(`\$(\w+)|\bentity\s*\.\s*attribute\s*\(\s*(?:"([^"]*)"|'([^']*)'|` +
	"`([^`]*)`" + `)\s*\)`)

// currentSources returns the IDs of the attributes whose current values, rather than their maximums, the value with the
// given ID and limitation is figured from: for ST under a limitation, the attribute standing in for that ST or, without
// one, ST (see strengthFor); for Move and Dodge without an attribute standing in for them, Basic Move and Basic Speed
// (see Move and Dodge); for an attribute other than a pool standing in for any of those, each attribute its base reads
// as "$id" or entity.attribute("id"), either of which gives the current value there (see Attribute.Current); and for
// anything else, none.
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
	if def == nil || def.Type == attribute.Pool || def.Type == attribute.PoolRef || !baseReadsCurrents(id) {
		return nil
	}
	var ids []string
	for _, match := range scriptAttributeRef.FindAllStringSubmatch(def.Base, -1) {
		var attr *Attribute
		if match[1] != "" {
			attr = e.ResolveAttribute(match[1])
		} else if idOrName := match[2] + match[3] + match[4]; idOrName != "" {
			attr = e.Attributes.Find(idOrName)
		}
		if attr != nil && !slices.Contains(ids, attr.AttrID) {
			ids = append(ids, attr.AttrID)
		}
	}
	return ids
}
