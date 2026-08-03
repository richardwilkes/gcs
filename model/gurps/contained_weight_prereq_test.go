// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// newTestEquipment creates a piece of equipment with the given base weight and quantity, attaching it to the parent, if
// one was provided.
func newTestEquipment(parent *gurps.Equipment, container bool, baseWeight string, qty fxp.Int) *gurps.Equipment {
	e := gurps.NewEquipment(nil, parent, container)
	e.BaseWeight = baseWeight
	e.Quantity = qty
	if parent != nil {
		parent.Children = append(parent.Children, e)
	}
	return e
}

// A container's contained weight is the weight it holds, which is independent of how many copies of the container there
// are. Deriving it as ExtendedWeight-AdjustedWeight instead yielded S+qty*C for qty > 1 and -S for qty == 0.
func TestContainedWeightIsIndependentOfQuantity(t *testing.T) {
	c := check.New(t)
	units := fxp.Pound
	container := newTestEquipment(nil, true, "5 lb", fxp.One)
	newTestEquipment(container, false, "3 lb", fxp.One)
	expected := fxp.WeightFromInteger(3, units)
	for _, qty := range []fxp.Int{fxp.One, fxp.Two, fxp.Ten, 0, -fxp.One} {
		container.Quantity = qty
		c.Equal(expected, container.ContainedWeight(false, units), "quantity %v", qty)
	}
}

// The extended weight itself must still scale with the quantity: (self + contents) * qty, and 0 for qty <= 0.
func TestExtendedWeightStillScalesWithQuantity(t *testing.T) {
	c := check.New(t)
	units := fxp.Pound
	container := newTestEquipment(nil, true, "5 lb", fxp.One)
	newTestEquipment(container, false, "3 lb", fxp.One)
	c.Equal(fxp.WeightFromInteger(8, units), container.ExtendedWeight(false, units))
	container.Quantity = fxp.Two
	c.Equal(fxp.WeightFromInteger(16, units), container.ExtendedWeight(false, units))
	container.Quantity = 0
	c.Equal(fxp.Weight(0), container.ExtendedWeight(false, units))
}

// Contained weight reductions are applied to the contents, and only once, regardless of the quantity.
func TestContainedWeightHonorsReductions(t *testing.T) {
	c := check.New(t)
	units := fxp.Pound
	container := newTestEquipment(nil, true, "5 lb", fxp.Two)
	newTestEquipment(container, false, "3 lb", fxp.Two) // 6 lb of contents
	percentage := gurps.NewContainedWeightReduction()
	percentage.Reduction = "50%"
	fixed := gurps.NewContainedWeightReduction()
	fixed.Reduction = "1 lb"
	container.Features = gurps.Features{percentage, fixed}
	c.Equal(fxp.WeightFromInteger(2, units), container.ContainedWeight(false, units))

	// A reduction larger than the contents clamps to zero rather than going negative.
	fixed.Reduction = "20 lb"
	c.Equal(fxp.Weight(0), container.ContainedWeight(false, units))
}

// A non-container carries nothing, so the weight criteria is skipped entirely and the prereq is satisfied.
func TestContainedWeightPrereqNonContainer(t *testing.T) {
	c := check.New(t)
	e := newTestEquipment(nil, false, "3 lb", fxp.One)
	p := gurps.NewContainedWeightPrereq(nil)
	c.True(p.Satisfied(nil, e, nil, "", nil))
	p.Has = false
	c.False(p.Satisfied(nil, e, nil, "", nil))
}

// The prereq comparison must not shift as the container's quantity changes.
func TestContainedWeightPrereqIgnoresQuantity(t *testing.T) {
	c := check.New(t)
	container := newTestEquipment(nil, true, "5 lb", fxp.One)
	newTestEquipment(container, false, "3 lb", fxp.One)

	atMost5 := gurps.NewContainedWeightPrereq(nil) // at most 5 lb
	c.Equal(criteria.AtMostNumber, atMost5.WeightCriteria.Compare)
	atLeast1 := gurps.NewContainedWeightPrereq(nil)
	atLeast1.WeightCriteria.Compare = criteria.AtLeastNumber
	atLeast1.WeightCriteria.Qualifier = fxp.WeightFromInteger(1, fxp.Pound)

	// 3 lb of contents: at most 5 lb holds and at least 1 lb holds, no matter the quantity.
	for _, qty := range []fxp.Int{fxp.One, fxp.Two, fxp.Ten, 0} {
		container.Quantity = qty
		c.True(atMost5.Satisfied(nil, container, nil, "", nil), "at most 5 lb, quantity %v", qty)
		c.True(atLeast1.Satisfied(nil, container, nil, "", nil), "at least 1 lb, quantity %v", qty)
	}

	// A container that really is over the limit still fails, regardless of the quantity.
	container.Children[0].BaseWeight = "9 lb"
	for _, qty := range []fxp.Int{fxp.One, fxp.Two, 0} {
		container.Quantity = qty
		c.False(atMost5.Satisfied(nil, container, nil, "", nil), "at most 5 lb, quantity %v", qty)
	}
}
