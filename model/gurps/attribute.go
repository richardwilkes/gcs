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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"hash"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

var _ Hashable = &Attribute{}

// AttributeData holds the Attribute data that is written to disk.
type AttributeData struct {
	AttrID     string  `json:"attr_id"`
	Adjustment fxp.Int `json:"adj"`
	Damage     fxp.Int `json:"damage,omitzero"`
	// Unknown holds the members this release does not know, so that they survive a load/save cycle unchanged.
	Unknown jsontext.Value `json:",embed"`
}

// Attribute holds the current state of an AttributeDef.
type Attribute struct {
	AttributeData
	Entity        *Entity `json:"-"`
	Bonus         fxp.Int `json:"-"`
	CostReduction fxp.Int `json:"-"`
	Order         int     `json:"-"`
}

// NewAttribute creates a new Attribute.
func NewAttribute(entity *Entity, attrID string, order int) *Attribute {
	return &Attribute{
		AttrID: attrID,
		Entity: entity,
		Order:  order,
	}
}

// MarshalJSONTo implements json.MarshalerTo.
func (a *Attribute) MarshalJSONTo(enc *jsontext.Encoder) error {
	if a.Entity != nil && !omitCalc(enc) {
		if def := a.AttributeDef(); def != nil {
			type calc struct {
				Value   fxp.Int `json:"value"`
				Current fxp.Int `json:"current"`
				Points  fxp.Int `json:"points"`
			}
			if def.IsSeparator() {
				return json.MarshalEncode(enc, &a.AttributeData)
			}
			data := struct {
				AttributeData
				Calc calc `json:"calc"`
			}{
				AttributeData: a.AttributeData,
				Calc: calc{
					Value:   a.Maximum(),
					Current: a.Current(),
					Points:  a.PointCost(),
				},
			}
			return json.MarshalEncode(enc, &data)
		}
	}
	return json.MarshalEncode(enc, &a.AttributeData)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (a *Attribute) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var data struct {
		AttributeData
		// Written for third parties and ignored on load, but named so that Unknown doesn't keep it.
		Calc jsontext.Value `json:"calc"`
	}
	if err := json.UnmarshalDecode(dec, &data); err != nil {
		return err
	}
	a.AttributeData = data.AttributeData
	return nil
}

// Clone returns a copy of this Attribute, bound to the given Entity.
func (a *Attribute) Clone(entity *Entity) *Attribute {
	clone := *a
	clone.Entity = entity
	clone.Unknown = slices.Clone(a.Unknown)
	return &clone
}

// ID returns the ID.
func (a *Attribute) ID() string {
	return a.AttrID
}

// SetID sets the ID, sanitizing it in the process, so the stored value may differ from what was passed in.
func (a *Attribute) SetID(value string) {
	a.AttrID = SanitizeID(value, false, ReservedIDs...)
}

// AttributeDef looks up the AttributeDef this Attribute references from the Entity. May return nil.
func (a *Attribute) AttributeDef() *AttributeDef {
	if a.Entity == nil {
		return nil
	}
	return a.Entity.SheetSettings.Attributes.Set[a.AttrID]
}

// Maximum returns the maximum value of a pool or the adjusted attribute value for other types.
func (a *Attribute) Maximum() fxp.Int {
	return a.maximumWithBonus(a.bonus())
}

// bonus returns Bonus or, in a computation at base, the total of the bonuses it sees given the attribute, which leaves
// out what the pools' thresholds put in effect (see Entity.atBaseWhile), and, while a pool's base or thresholds are
// found, the total without what the features those thresholds hold add to it; see Entity.thresholdFeaturesLeftOut.
func (a *Attribute) bonus() fxp.Int {
	e := a.Entity
	if e == nil || (e.atBase == 0 && !e.thresholdFeaturesLeftOut()) {
		return a.Bonus
	}
	bonus := e.AttributeBonusFor(a.AttrID, stlimit.None, nil)
	if def := a.AttributeDef(); def != nil && !def.AllowsDecimal() {
		bonus = bonus.Floor()
	}
	return bonus
}

// maximumWithBonus is Maximum with the given bonus in place of the attribute's own.
func (a *Attribute) maximumWithBonus(bonus fxp.Int) fxp.Int {
	def := a.AttributeDef()
	if def == nil || def.IsSeparator() {
		return 0
	}
	maximum := def.BaseValue(a) + a.Adjustment + bonus
	if !def.AllowsDecimal() {
		maximum = maximum.Floor()
	}
	return maximum
}

// SetMaximum sets the maximum value.
func (a *Attribute) SetMaximum(value fxp.Int) {
	if a.Maximum() == value {
		return
	}
	if def := a.AttributeDef(); def != nil && !def.IsSeparator() {
		a.Adjustment = value - (def.BaseValue(a) + a.Bonus)
	}
}

// BonusTooltip returns a description of the bonuses currently being applied to this attribute, or an empty string if
// there are none. In each section, the bonuses that alter only the current value follow those that make up the
// maximum; a pool, which they don't alter, lists none. The first section is followed, for an attribute whose base reads
// the current values of others, as "$id.current" does anywhere and "$id" does for one standing in for Striking ST,
// Lifting ST, Throwing ST, Dodge or Move, by those that alter only the current values of those attributes, under a
// heading naming each; see Entity.addSourceAdjustments. Strength bonuses limited to striking, lifting or throwing are
// listed in sections of their own, since they don't alter the attribute's value. An attribute standing in for Striking
// ST, Lifting ST or Throwing ST lists in the first section, after its own, those ST is given under the same
// limitation, which its current value takes; see Current.
func (a *Attribute) BonusTooltip() string {
	if a.Entity == nil {
		return ""
	}
	def := a.AttributeDef()
	pool := def != nil && (def.Type == attribute.Pool || def.Type == attribute.PoolRef)
	add := func(limitation stlimit.Option, buffer *xbytes.InsertBuffer) {
		a.Entity.AttributeBonusFor(a.AttrID, limitation, buffer)
		if !pool {
			a.Entity.currentAdjustmentFor(a.AttrID, limitation, buffer)
		}
	}
	var buffer xbytes.InsertBuffer
	add(stlimit.None, &buffer)
	if limitation := a.strengthLimitation(); limitation != stlimit.None {
		a.Entity.AttributeBonusFor(StrengthID, limitation, &buffer)
		a.Entity.currentAdjustmentFor(StrengthID, limitation, &buffer)
	}
	a.Entity.addSourceAdjustments(a.AttrID, stlimit.None, &buffer)
	if a.AttrID == StrengthID {
		var limited xbytes.InsertBuffer
		for _, limitation := range stlimit.Options {
			if limitation == stlimit.None {
				continue
			}
			limited.Reset()
			add(limitation, &limited)
			if limited.Len() != 0 {
				buffer.WriteByte('\n')
				buffer.WriteString(xstrings.FirstToUpper(limitation.String()))
				buffer.WriteByte(':')
				buffer.WriteString(limited.String())
			}
		}
	}
	if buffer.Len() == 0 {
		return ""
	}
	return IncludesModifiersFrom() + ":" + buffer.String()
}

// NameMatches checks if the name matches this Attribute's name or full name. Case is ignored.
func (a *Attribute) NameMatches(name string) bool {
	def := a.AttributeDef()
	if def == nil || def.IsSeparator() {
		return false
	}
	return strings.EqualFold(def.Name, name) || strings.EqualFold(def.FullName, name)
}

// Current returns the current value. That of a pool is its maximum less the damage taken. That of any other attribute
// is its maximum with the amounts the pools' current thresholds give it added, then multiplied by each multiplier given
// it in turn, rounded to whole after each unless the attribute allows decimals and stands in for neither Dodge nor
// Move, which the sheet uses as whole numbers; see Entity.processThresholdFeatures.
// That of one standing in for Striking ST, Lifting ST or Throwing ST then takes the bonuses ST is given under the same
// limitation, as the ST the sheet works out without it does, so that it is the ST in use; see Entity.limitedStrength.
// Scripts read it as "$id.current"; "$id", the attributes derived from it and its point cost use the maximum, except
// that the current value of an attribute standing in for Striking ST, Lifting ST, Throwing ST, Dodge or Move starts
// from its base with each attribute other than a pool the base reads as "$id" giving its current value, as the value
// the sheet works out without such an attribute starts from the current value of ST, Basic Speed or Basic Move; see
// baseReadsCurrents. A pool's current value is what it has left, so the base reads its maximum there still.
func (a *Attribute) Current() fxp.Int {
	return a.currentWith(a.bonus(), false, true)
}

// currentWith is Current with the given bonus in place of the attribute's own and, with forPrereqs set, with the pools'
// thresholds as the prerequisites being judged see them; see Entity.currentAdjustmentFor. With limitedAmounts unset, an
// attribute standing in for Striking ST, Lifting ST or Throwing ST takes, of the bonuses ST is given under the same
// limitation, the multipliers alone; see Entity.LiftingStrengthForThrowing.
func (a *Attribute) currentWith(bonus fxp.Int, forPrereqs, limitedAmounts bool) fxp.Int {
	def := a.AttributeDef()
	if def == nil || def.IsSeparator() {
		return 0
	}
	if def.Type == attribute.Pool || def.Type == attribute.PoolRef {
		return a.maximumWithBonus(bonus) - a.Damage
	}
	var value fxp.Int
	if baseReadsCurrents(a.AttrID) {
		value = a.maximumReadingCurrents(bonus)
	} else {
		value = a.maximumWithBonus(bonus)
	}
	var adj currentAdjustment
	if forPrereqs {
		adj = a.Entity.currentAdjustmentFor(a.AttrID, stlimit.None, nil)
	} else {
		adj = a.Entity.sheetAdjustment(a.AttrID, stlimit.None)
	}
	// Move and Dodge are whole wherever the sheet uses them, so a product for one standing in for either is rounded even
	// when it allows decimals, so that a halving rounds up as it does without the attribute.
	decimal := def.AllowsDecimal()
	value = adj.applyRounding(value, !decimal, !decimal || a.AttrID == DodgeID || a.AttrID == MoveID)
	if limitation := strengthLimitationOf(a.AttrID); limitation != stlimit.None {
		value = a.Entity.limitedStrength(value, limitation, forPrereqs, limitedAmounts)
	}
	return value
}

// strengthLimitation returns the limitation of the ST bonuses the attribute's current value takes (see Current): that
// of the Striking ST, Lifting ST or Throwing ST it stands in for, unless it is a pool, and stlimit.None otherwise.
func (a *Attribute) strengthLimitation() stlimit.Option {
	def := a.AttributeDef()
	if def == nil || def.IsSeparator() || def.Type == attribute.Pool || def.Type == attribute.PoolRef {
		return stlimit.None
	}
	return strengthLimitationOf(a.AttrID)
}

// strengthLimitationOf returns the limitation of the ST bonuses that the Striking ST, Lifting ST or Throwing ST an
// attribute with the given ID stands in for takes, or stlimit.None for any other ID.
func strengthLimitationOf(attrID string) stlimit.Option {
	switch attrID {
	case StrikingStrengthID:
		return stlimit.StrikingOnly
	case LiftingStrengthID:
		return stlimit.LiftingOnly
	case ThrowingStrengthID:
		return stlimit.ThrowingOnly
	default:
		return stlimit.None
	}
}

// baseReadsCurrents reports whether the attribute with the given ID stands in for a value that the sheet works out,
// when it doesn't define the attribute, from the current value of another: Striking ST, Lifting ST and Throwing ST from
// ST's, Dodge from Basic Speed's and Move from Basic Move's. The base such an attribute's current value starts from reads
// each attribute other than a pool as "$id" for its current value; see Current.
func baseReadsCurrents(attrID string) bool {
	switch attrID {
	case StrikingStrengthID, LiftingStrengthID, ThrowingStrengthID, DodgeID, MoveID:
		return true
	default:
		return false
	}
}

// maximumReadingCurrents is maximumWithBonus with the script giving the base reading each attribute other than a pool
// as "$id" for its current value rather than its maximum; see Entity.scriptReadsCurrents. The scripts that one runs in
// turn, including those giving the bases of the attributes it reads, read maximums as usual.
func (a *Attribute) maximumReadingCurrents(bonus fxp.Int) fxp.Int {
	e := a.Entity
	saved := e.currentsDepth
	e.currentsDepth = e.scriptResolvingDepth + 1
	defer func() { e.currentsDepth = saved }()
	return a.maximumWithBonus(bonus)
}

// CurrentThreshold returns the current PoolThreshold, if any. While the prerequisites of an item carrying a grant are
// judged, it is the one they see: the pool's current value is taken without the bonuses from what the grant put on the
// sheet, the scripts giving the thresholds run as those prerequisites see the sheet, and the values they read carry no
// current adjustment but what the items' multipliers do, as the sheet's own search sees (see
// Entity.processThresholdFeatures). It is then found only once while they are judged, since their view of the sheet
// holds still meanwhile and the Basic Lift every skill level with an encumbrance penalty reads asks for it; see
// Entity.JudgingPrereqsOf.
func (a *Attribute) CurrentThreshold() *PoolThreshold {
	def := a.AttributeDef()
	if def == nil || def.IsSeparator() || len(def.Thresholds) == 0 {
		return nil
	}
	e := a.Entity
	if !e.carriesGrant(e.judging) {
		return a.thresholdFor(def, a.Current())
	}
	if t, ok := e.judgedThresholds[a]; ok {
		return t
	}
	t := func() *PoolThreshold {
		e.judgedThresholdSearches++
		e.scriptsForPrereqs++
		defer func() {
			e.scriptsForPrereqs--
			e.judgedThresholdSearches--
		}()
		return a.thresholdFor(def, e.attributeForPrereqs(a, true))
	}()
	if e.judgedThresholds == nil {
		e.judgedThresholds = make(map[*Attribute]*PoolThreshold)
	}
	e.judgedThresholds[a] = t
	return t
}

// thresholdFor returns the first of the definition's thresholds that the given current value of the pool lies at or
// below, or nil if there is none.
func (a *Attribute) thresholdFor(def *AttributeDef, cur fxp.Int) *PoolThreshold {
	for _, t := range def.Thresholds {
		if cur <= t.Threshold(a) {
			return t
		}
	}
	return nil
}

// PointCost returns the number of points spent on this Attribute. While the prerequisites of an item carrying a grant
// are judged, it is the cost they see, without the size modifier bonuses and cost reductions from what the grant put on
// the sheet (see Entity.JudgingPrereqsOf), and in a computation at base, the cost with the size modifier and cost
// reductions it sees; see Entity.sizeModifier and Entity.atBaseWhile.
func (a *Attribute) PointCost() fxp.Int {
	def := a.AttributeDef()
	if def == nil || def.IsSeparator() {
		return 0
	}
	var sm int
	costReduction := a.CostReduction
	if e := a.Entity; e != nil {
		sm = e.sizeModifier()
		if e.carriesGrant(e.judging) || e.atBase > 0 {
			costReduction = e.CostReductionFor(a.AttrID)
		}
	}
	return def.ComputeCost(a.Entity, a.Adjustment, costReduction, sm)
}

// Hash writes this object's contents into the hasher.
func (a *Attribute) Hash(h hash.Hash) {
	xhash.StringWithLen(h, a.AttrID)
	xhash.Num64(h, a.Adjustment)
	xhash.Num64(h, a.Damage)
	xhash.Num64(h, a.Bonus)
	xhash.Num64(h, a.CostReduction)
	xhash.Num64(h, a.Order)
	if def := a.AttributeDef(); def != nil {
		def.Hash(h)
	} else {
		xhash.Num8(h, uint8(255))
	}
}
