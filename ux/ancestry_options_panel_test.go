// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestAncestryEditorPanelFieldsBind(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	d := newTestAncestryEditorDockable(a)
	g := a.GenderOptions[0]
	integerFieldFor(t, d, g.KeyPrefix+"weight").SetText("7")
	c.Equal(7, g.Weight)
	stringFieldFor(t, d, g.Value.KeyPrefix+"name").SetText("Neuter")
	c.Equal("Neuter", g.Value.Name)
	stringFieldFor(t, d, a.CommonOptions.KeyPrefix+"height_script").SetText("entity.randomHeightInInches($st)")
	c.Equal("entity.randomHeightInInches($st)", a.CommonOptions.HeightScript)
	stringFieldFor(t, d, a.CommonOptions.KeyPrefix+"weight_script").SetText("entity.randomWeightInPounds($st)")
	c.Equal("entity.randomWeightInPounds($st)", a.CommonOptions.WeightScript)
	stringFieldFor(t, d, g.Value.KeyPrefix+"age_script").SetText(`dice.roll("1d10+20")`)
	c.Equal(`dice.roll("1d10+20")`, g.Value.AgeScript)
	c.Equal("", a.CommonOptions.AgeScript, "the gender's script is not the common one")
	c.True(d.Modified())
}

// Each gender's lists must edit that gender's options, not the common ones, and all four must be present for each.
func TestWeightedStringOptionsPerGenderListsBindToGender(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	d := newTestAncestryEditorDockable(a)
	g := a.GenderOptions[1].Value
	for _, list := range []*[]*gurps.WeightedStringOption{&g.HairOptions, &g.EyeOptions, &g.SkinOptions, &g.HandednessOptions} {
		listPanelFor(t, d, list).addOption()
	}
	g = d.model.GenderOptions[1].Value
	c.Equal(1, len(g.HairOptions))
	c.Equal(1, len(g.EyeOptions))
	c.Equal(1, len(g.SkinOptions))
	c.Equal(1, len(g.HandednessOptions))
	common := d.model.CommonOptions
	c.Equal(0, len(common.HairOptions)+len(common.EyeOptions)+len(common.SkinOptions)+len(common.HandednessOptions),
		"the common options are untouched")
	other := d.model.GenderOptions[0].Value
	c.Equal(0, len(other.HairOptions)+len(other.EyeOptions)+len(other.SkinOptions)+len(other.HandednessOptions),
		"the other gender is untouched")
}
