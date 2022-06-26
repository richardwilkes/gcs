/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/attribute"
	"github.com/richardwilkes/gcs/v5/model/id"
	"github.com/richardwilkes/json"
)

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
			if def.Type == attribute.Pool {
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
	a.AttrID = id.Sanitize(value, false, ReservedIDs...)
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
	if def == nil {
		return 0
	}
	max := def.BaseValue(a.Entity) + a.Adjustment + a.Bonus
	if def.Type != attribute.Decimal {
		max = max.Trunc()
	}
	return max
}

// SetMaximum sets the maximum value.
func (a *Attribute) SetMaximum(value fxp.Int) {
	if a.Maximum() == value {
		return
	}
	if def := a.AttributeDef(); def != nil {
		a.Adjustment = value - (def.BaseValue(a.Entity) + a.Bonus)
	}
}

// Current returns the current value. Same as .Maximum() if not a pool.
func (a *Attribute) Current() fxp.Int {
	max := a.Maximum()
	def := a.AttributeDef()
	if def == nil || def.Type != attribute.Pool {
		return max
	}
	return max - a.Damage
}

// CurrentThreshold return the current PoolThreshold, if any.
func (a *Attribute) CurrentThreshold() *PoolThreshold {
	def := a.AttributeDef()
	if def == nil {
		return nil
	}
	max := a.Maximum()
	cur := a.Current()
	for _, threshold := range def.Thresholds {
		if cur <= threshold.Threshold(max) {
			return threshold
		}
	}
	return nil
}

// PointCost returns the number of points spent on this Attribute.
func (a *Attribute) PointCost() fxp.Int {
	def := a.AttributeDef()
	if def == nil {
		return 0
	}
	var sm int
	if a.Entity != nil {
		sm = a.Entity.Profile.AdjustedSizeModifier()
	}
	return def.ComputeCost(a.Entity, a.Adjustment, a.CostReduction, sm)
}

// IsThresholdOpMet if the given ThresholdOp is met.
func IsThresholdOpMet(op attribute.ThresholdOp, attributes *Attributes) bool {
	for _, one := range attributes.Set {
		if threshold := one.CurrentThreshold(); threshold != nil && threshold.ContainsOp(op) {
			return true
		}
	}
	return false
}

// CountThresholdOpMet counts the number of times the given ThresholdOp is met.
func CountThresholdOpMet(op attribute.ThresholdOp, attributes *Attributes) int {
	total := 0
	for _, one := range attributes.Set {
		if threshold := one.CurrentThreshold(); threshold != nil && threshold.ContainsOp(op) {
			total++
		}
	}
	return total
}
