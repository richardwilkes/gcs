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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/threshold"
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
				Value   fxp.Int  `json:"value"`
				Current *fxp.Int `json:"current,omitzero"`
				Points  fxp.Int  `json:"points"`
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
					Value:  a.Maximum(),
					Points: a.PointCost(),
				},
			}
			if def.Type == attribute.Pool || def.Type == attribute.PoolRef {
				current := a.Current()
				data.Calc.Current = &current
			}
			return json.MarshalEncode(enc, &data)
		}
	}
	return json.MarshalEncode(enc, &a.AttributeData)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (a *Attribute) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	a.AttributeData = AttributeData{}
	return json.UnmarshalDecode(dec, &a.AttributeData)
}

// Clone returns a copy of this Attribute, bound to the given Entity.
func (a *Attribute) Clone(entity *Entity) *Attribute {
	clone := *a
	clone.Entity = entity
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
	return a.maximumWithBonus(a.Bonus)
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
// there are none. Strength bonuses limited to striking, lifting or throwing are listed in sections of their own, since
// they don't alter the attribute's value.
func (a *Attribute) BonusTooltip() string {
	if a.Entity == nil {
		return ""
	}
	var buffer xbytes.InsertBuffer
	a.Entity.AttributeBonusFor(a.AttrID, stlimit.None, &buffer)
	if a.AttrID == StrengthID {
		var limited xbytes.InsertBuffer
		for _, limitation := range stlimit.Options {
			if limitation == stlimit.None {
				continue
			}
			limited.Reset()
			a.Entity.AttributeBonusFor(a.AttrID, limitation, &limited)
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

// Current returns the current value. Same as .Maximum() if not a pool.
func (a *Attribute) Current() fxp.Int {
	return a.currentWithBonus(a.Bonus)
}

// currentWithBonus is Current with the given bonus in place of the attribute's own.
func (a *Attribute) currentWithBonus(bonus fxp.Int) fxp.Int {
	def := a.AttributeDef()
	if def == nil || def.IsSeparator() {
		return 0
	}
	maximum := a.maximumWithBonus(bonus)
	if def.Type != attribute.Pool && def.Type != attribute.PoolRef {
		return maximum
	}
	return maximum - a.Damage
}

// CurrentThreshold returns the current PoolThreshold, if any. While the prerequisites of an item carrying a grant are
// judged, it is the one they see: the pool's current value is taken without the bonuses from what the grant put on the
// sheet, and the scripts giving the thresholds run as those prerequisites see the sheet. It is then found only once
// while they are judged, since their view of the sheet holds still meanwhile and the Basic Lift every skill level with
// an encumbrance penalty reads asks for it; see Entity.JudgingPrereqsOf.
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
	cur := e.attributeForPrereqs(a, true)
	t := func() *PoolThreshold {
		e.scriptsForPrereqs++
		defer func() { e.scriptsForPrereqs-- }()
		return a.thresholdFor(def, cur)
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
// the sheet; see Entity.JudgingPrereqsOf.
func (a *Attribute) PointCost() fxp.Int {
	def := a.AttributeDef()
	if def == nil || def.IsSeparator() {
		return 0
	}
	var sm int
	costReduction := a.CostReduction
	if e := a.Entity; e != nil {
		if e.carriesGrant(e.judging) {
			sm = e.sizeModifierForPrereqs()
			costReduction = e.CostReductionFor(a.AttrID)
		} else {
			sm = e.Profile.AdjustedSizeModifier()
		}
	}
	return def.ComputeCost(a.Entity, a.Adjustment, costReduction, sm)
}

// IsThresholdOpMet returns true if any attribute's current threshold contains the given op.
func IsThresholdOpMet(op threshold.Op, attributes *Attributes) bool {
	for _, one := range attributes.Set {
		if one.thresholdOpMet(op) {
			return true
		}
	}
	return false
}

// CountThresholdOpMet returns how many attributes have a current threshold containing the given op.
func CountThresholdOpMet(op threshold.Op, attributes *Attributes) int {
	total := 0
	for _, one := range attributes.Set {
		if one.thresholdOpMet(op) {
			total++
		}
	}
	return total
}

// thresholdOpMet returns true if the attribute's current threshold contains the given op. The current threshold is
// only found when one of the attribute's thresholds contains the op, since finding it runs their scripts.
func (a *Attribute) thresholdOpMet(op threshold.Op) bool {
	def := a.AttributeDef()
	if def == nil || !slices.ContainsFunc(def.Thresholds, func(t *PoolThreshold) bool { return t.ContainsOp(op) }) {
		return false
	}
	t := a.CurrentThreshold()
	return t != nil && t.ContainsOp(op)
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
