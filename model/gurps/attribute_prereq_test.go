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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestAttributePrereqNeedsTheAttributes checks that an attribute prerequisite's comparison is met only when the sheet
// defines each attribute it names, whatever the comparison, so a "Does not have" one is met when either is missing,
// and that the size modifier and the Dodge, which the editor offers alongside the attributes, are compared at their
// values, the Dodge being the current one whether or not the sheet has a dodge attribute, while a parry or block, which
// it no longer offers, counts as missing.
func TestAttributePrereqNeedsTheAttributes(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	e.Attributes.Set[gurps.HitPointsID].Damage = fxp.FromInteger(13)
	e.Recalculate()
	c.Equal(fxp.FromInteger(-3), e.ResolveAttributeCurrent(gurps.HitPointsID), "precondition: HP is at -3")
	c.Equal(fxp.Min, e.ResolveAttributeCurrent("sanity"), "precondition: the sheet has no Sanity")
	c.Nil(e.ResolveAttribute(gurps.DodgeID), "precondition: the sheet has no dodge attribute")
	e.Profile.SizeModifier = 2
	pack := gurps.NewEquipment(e, nil, false)
	pack.BaseWeight = "30 lb"
	e.CarriedEquipment = []*gurps.Equipment{pack}
	c.Equal(encumbrance.Light, e.EncumbranceLevel(false), "precondition: the pack is a Light load")
	attr := func(has bool, which, combined string, compare criteria.NumericComparison,
		qualifier int,
	) *gurps.AttributePrereq {
		p := gurps.NewAttributePrereq(e)
		p.Has = has
		p.Which = which
		p.CombinedWith = combined
		p.QualifierCriteria = criteria.Number{Compare: compare, Qualifier: fxp.FromInteger(qualifier)}
		return p
	}
	for _, one := range []struct {
		name   string
		prereq *gurps.AttributePrereq
		want   bool
	}{
		{"defined", attr(true, gurps.HitPointsID, "", criteria.AtLeastNumber, -3), true},
		{"defined pair", attr(true, gurps.HitPointsID, gurps.StrengthID, criteria.AtLeastNumber, 7), true},
		{"missing at most", attr(true, "sanity", "", criteria.AtMostNumber, 10), false},
		{"missing at any value", attr(true, "sanity", "", criteria.AnyNumber, 0), false},
		{"missing plus negative", attr(true, "sanity", gurps.HitPointsID, criteria.AtLeastNumber, 10), false},
		{"negative plus missing", attr(true, gurps.HitPointsID, "sanity", criteria.AtLeastNumber, 10), false},
		{"both missing", attr(true, "sanity", "luck", criteria.AtLeastNumber, 0), false},
		{"does not have missing", attr(false, "sanity", "", criteria.AtLeastNumber, 10), true},
		{"does not have missing pair", attr(false, "sanity", gurps.HitPointsID, criteria.AtLeastNumber, 10), true},
		{"size modifier", attr(true, gurps.SizeModifierID, "", criteria.EqualsNumber, 2), true},
		{"does not have size modifier", attr(false, gurps.SizeModifierID, "", criteria.AtLeastNumber, 1), false},
		{"ST plus size modifier", attr(true, gurps.StrengthID, gurps.SizeModifierID, criteria.EqualsNumber, 12), true},
		{"dodge", attr(true, gurps.DodgeID, "", criteria.EqualsNumber, 3), true}, // 8/2 at HP -3, -1 for Light
		{"parry", attr(true, gurps.ParryID, "", criteria.AnyNumber, 0), false},
		{"block", attr(true, gurps.BlockID, "", criteria.AnyNumber, 0), false},
	} {
		c.Equal(one.want, one.prereq.Satisfied(e, nil, nil, "", nil), one.name)
	}
	e.SheetSettings.Attributes.Set[gurps.DodgeID] = &gurps.AttributeDef{
		DefID: gurps.DodgeID,
		Type:  attribute.Integer,
		Name:  "Dodge",
		Base:  "$basic_speed + 3",
	}
	e.Attributes.Set[gurps.DodgeID] = gurps.NewAttribute(e, gurps.DodgeID, len(e.Attributes.Set))
	e.Recalculate()
	c.Equal(fxp.Eight, e.Attributes.Maximum(gurps.DodgeID), "precondition: the dodge attribute is 8")
	c.Equal(fxp.Four, e.ResolveAttributeCurrent(gurps.DodgeID), "precondition: its current value is halved at HP -3")
	c.True(attr(true, gurps.DodgeID, "", criteria.EqualsNumber, 3).Satisfied(e, nil, nil, "", nil),
		"a dodge attribute is halved and penalized as the Dodge without one is")
}
