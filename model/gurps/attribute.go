// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/threshold"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var _ Hashable = &Attribute{}

// AttributeData holds the Attribute data that is written to disk.
type AttributeData struct {
	AttrID     string  `json:"attr_id"`
	Adjustment fxp.Int `json:"adj"`
	Damage     fxp.Int `json:"damage,omitempty"`
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
		AttributeData: AttributeData{
			AttrID: attrID,
		},
		Entity: entity,
		Order:  order,
	}
}

// MarshalJSON implements json.Marshaler.
func (a *Attribute) MarshalJSON() ([]byte, error) {
	if a.Entity != nil {
		if def := a.AttributeDef(); def != nil {
			type calc struct {
				Value   fxp.Int  `json:"value"`
				Current *fxp.Int `json:"current,omitempty"`
				Points  fxp.Int  `json:"points"`
			}
			if def.IsSeparator() {
				return json.Marshal(&a.AttributeData)
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
			return json.Marshal(&data)
		}
	}
	return json.Marshal(&a.AttributeData)
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Attribute) UnmarshalJSON(data []byte) error {
	a.AttributeData = AttributeData{}
	return json.Unmarshal(data, &a.AttributeData)
}

// Clone a copy of this.
func (a *Attribute) Clone(entity *Entity) *Attribute {
	clone := *a
	clone.Entity = entity
	return &clone
}

// ID returns the ID.
func (a *Attribute) ID() string {
	return a.AttrID
}

// SetID sets the ID, sanitizing it in the process (i.e. it may be changed from what you set -- read it back if you want
// to be sure of what it gets set to.
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
	def := a.AttributeDef()
	if def == nil || def.IsSeparator() {
		return 0
	}
	maximum := def.BaseValue(a.Entity) + a.Adjustment + a.Bonus
	if !def.AllowsDecimal() {
		maximum = maximum.Trunc()
	}
	return maximum
}

// SetMaximum sets the maximum value.
func (a *Attribute) SetMaximum(value fxp.Int) {
	if a.Maximum() == value {
		return
	}
	if def := a.AttributeDef(); def != nil && !def.IsSeparator() {
		a.Adjustment = value - (def.BaseValue(a.Entity) + a.Bonus)
	}
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
	def := a.AttributeDef()
	if def == nil || def.IsSeparator() {
		return 0
	}
	maximum := a.Maximum()
	if def.Type != attribute.Pool && def.Type != attribute.PoolRef {
		return maximum
	}
	return maximum - a.Damage
}

// CurrentThreshold return the current PoolThreshold, if any.
func (a *Attribute) CurrentThreshold() *PoolThreshold {
	def := a.AttributeDef()
	if def == nil || def.IsSeparator() {
		return nil
	}
	if len(def.Thresholds) != 0 {
		cur := a.Current()
		for _, t := range def.Thresholds {
			if cur <= t.Threshold(a.Entity) {
				return t
			}
		}
	}
	return nil
}

// PointCost returns the number of points spent on this Attribute.
func (a *Attribute) PointCost() fxp.Int {
	def := a.AttributeDef()
	if def == nil || def.IsSeparator() {
		return 0
	}
	var sm int
	if a.Entity != nil {
		sm = a.Entity.Profile.AdjustedSizeModifier()
	}
	return def.ComputeCost(a.Entity, a.Adjustment, a.CostReduction, sm)
}

// IsThresholdOpMet if the given ThresholdOp is met.
func IsThresholdOpMet(op threshold.Op, attributes *Attributes) bool {
	for _, one := range attributes.Set {
		if t := one.CurrentThreshold(); t != nil && t.ContainsOp(op) {
			return true
		}
	}
	return false
}

// CountThresholdOpMet counts the number of times the given ThresholdOp is met.
func CountThresholdOpMet(op threshold.Op, attributes *Attributes) int {
	total := 0
	for _, one := range attributes.Set {
		if t := one.CurrentThreshold(); t != nil && t.ContainsOp(op) {
			total++
		}
	}
	return total
}

// Hash writes this object's contents into the hasher.
func (a *Attribute) Hash(h hash.Hash) {
	hashhelper.String(h, a.AttrID)
	hashhelper.Num64(h, a.Adjustment)
	hashhelper.Num64(h, a.Damage)
	hashhelper.Num64(h, a.Bonus)
	hashhelper.Num64(h, a.CostReduction)
	hashhelper.Num64(h, a.Order)
	if def := a.AttributeDef(); def != nil {
		def.Hash(h)
	} else {
		hashhelper.Num8(h, uint8(255))
	}
}
