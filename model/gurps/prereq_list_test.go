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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// TestPrereqListSkipsListsThatDoNotApply verifies that a list whose tech level condition doesn't match the sheet's, or
// that has nothing in it that applies, is left out of its parent's check, and that one left out at the top is met.
func TestPrereqListSkipsListsThatDoNotApply(t *testing.T) {
	e := NewEntity()
	e.Profile.TechLevel = "3"
	list := func(all bool, children ...Prereq) *PrereqList {
		p := NewPrereqList()
		p.All = all
		p.Prereqs = children
		return p
	}
	skipped := func(children ...Prereq) *PrereqList {
		p := list(true, children...)
		p.WhenTL = numberCriteria(criteria.AtLeastNumber, fxp.Nine)
		return p
	}
	attr := func(which string, atLeast fxp.Int) Prereq {
		p := NewAttributePrereq(e)
		p.Which = which
		p.QualifierCriteria.Qualifier = atLeast
		return p
	}
	met := func() Prereq { return attr(StrengthID, fxp.Ten) }
	unmet := func() Prereq { return attr(DexterityID, fxp.Twenty) }
	hidden := func() Prereq { return attr(IntelligenceID, fxp.Twenty) }
	equipment := func() Prereq {
		p := NewEquippedEquipmentPrereq()
		p.NameCriteria.Qualifier = "Longarm"
		return p
	}
	const (
		needsDX   = "\n- Has DX at least 20"
		needsGear = "\n- Has Longarm equipped"
	)
	for _, one := range []struct {
		name    string
		list    *PrereqList
		met     bool
		text    string
		penalty bool
	}{
		{"any of: skipped group and unmet item", list(false, skipped(hidden()), unmet()), false, needsDX, false},
		{"any of: skipped group and met item", list(false, skipped(hidden()), met()), true, "", false},
		{"all of: skipped group and met item", list(true, skipped(hidden()), met()), true, "", false},
		{"all of: skipped group and unmet item", list(true, skipped(hidden()), unmet()), false, needsDX, false},
		{"nested lists all skipped", list(false, list(true, skipped(hidden())), skipped(unmet())), true, "", false},
		{"any of: empty group and unmet item", list(false, list(true), unmet()), false, needsDX, false},
		{"empty top level", list(true), true, "", false},
		{"any of: skipped gear and met item", list(false, skipped(equipment()), met()), true, "", false},
		{"any of: skipped gear and unmet item", list(false, skipped(equipment()), unmet()), false, needsDX, false},
		{"any of: skipped group and unmet gear", list(false, skipped(hidden()), equipment()), false, needsGear, true},
		{"all of: unmet gear and met item", list(true, equipment(), met()), false, needsGear, true},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			var buffer xbytes.InsertBuffer
			penalty := false
			c.Equal(one.met, one.list.Satisfied(e, nil, &buffer, "\n- ", &penalty))
			c.Equal(one.text, buffer.String())
			c.Equal(one.penalty, penalty)
		})
	}
}
