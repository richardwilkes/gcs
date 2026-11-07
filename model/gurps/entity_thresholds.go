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

// processThresholdFeatures collects the features of the pools' current thresholds once those of the items are in
// place, since a pool's maximum, and so its threshold, rests on the bonuses the items give, and records what those
// thresholds and every multiplier do to each value; see currentAdjustment. The thresholds are found while no value
// carries such an adjustment, so a pool whose base or thresholds read a value they alter is found without the
// alteration rather than with what the previous pass left, and the caches are discarded once the adjustments are in
// place, since what those scripts cached was computed without them. Found outside a recalculation, such a pool's
// threshold may differ. The parry and block tooltips are built here, so that they name these bonuses too. A grant a
// threshold holds is collected here as well, and so is in effect while the pool is at the threshold; the rows it puts on
// the sheet are there whether or not it is, as for any grant (see forEachGrant).
func (e *Entity) processThresholdFeatures() {
	e.forEachThresholdFeatureList(func(t *PoolThreshold, def *AttributeDef, list Features) {
		for _, f := range list {
			if f.FeatureType().AllowedInThreshold() {
				e.processFeature(t, def, f, nil)
			}
		}
	})
	adjustments := make(map[adjustmentKey]currentAdjustment)
	for _, one := range e.features.currentAttributeBonuses {
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
// thresholds are taken from the thresholds as those prerequisites see the pools, and none at all while those
// thresholds are being found, as the sheet's own are found without any; see JudgingPrereqsOf,
// Attribute.CurrentThreshold and processThresholdFeatures.
func (e *Entity) currentAdjustmentFor(id string, limitation stlimit.Option, tooltip *xbytes.InsertBuffer) currentAdjustment {
	var adj currentAdjustment
	judging := e.carriesGrant(e.judging)
	if judging && e.judgedThresholdSearches > 0 {
		return adj
	}
	add := func(one *AttributeBonus) {
		if one.Attribute == id && one.ActualLimitation() == limitation && !e.bonusExcluded(one) {
			adj.add(one)
			one.AddToTooltip(tooltip)
		}
	}
	for _, one := range e.features.currentAttributeBonuses {
		if judging {
			if _, ok := one.Owner().(*PoolThreshold); ok {
				continue
			}
		}
		add(one)
	}
	if judging {
		e.forEachThresholdFeatureList(func(t *PoolThreshold, def *AttributeDef, list Features) {
			for _, f := range list {
				if one, ok := f.(*AttributeBonus); ok {
					one.SetOwner(t)
					one.SetSubOwner(def)
					add(one)
				}
			}
		})
	}
	return adj
}

// sheetAdjustment returns what the current thresholds and multipliers do to the value with the given ID and
// limitation, as processThresholdFeatures recorded it, which is nothing for a value they leave alone.
func (e *Entity) sheetAdjustment(id string, limitation stlimit.Option) currentAdjustment {
	if e == nil || e.currentAdjustments == nil {
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

// CurrentAdjustmentTooltip returns a line for each bonus that alters only the current value of the attribute with the
// given ID, or of the Move, Dodge, Parry or Block standing in for one, under the given ST limitation, such as
// "Tired (FP) [×0.5]", or an empty string when there are none.
func (e *Entity) CurrentAdjustmentTooltip(id string, limitation stlimit.Option) string {
	if e == nil {
		return ""
	}
	var buffer xbytes.InsertBuffer
	e.currentAdjustmentFor(id, limitation, &buffer)
	return strings.TrimPrefix(buffer.String(), "\n")
}
