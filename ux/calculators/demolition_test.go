// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calculators

import (
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestDemolitionCalculatorResults drives the demolition calculator in a headless workspace: the fields the mode and the
// explosive leave unused must be blank, and the worked example on BX415 must come out.
func TestDemolitionCalculatorResults(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	calc := openCalculator(t, screen)
	demolition := calc.demolition
	selectCalculatorTab(t, screen, calc, demolition)

	// BX415: a 6dx8 blast takes 16 lbs of TNT, or 20 lbs of dynamite, whose REF is 0.8.
	var explosivePopup *unison.PopupMenu[explosiveChoice]
	var refBlank, weightBlank, countBlank bool
	screen.Do(func() {
		explosivePopup, _ = uxtest.FirstPanelOfType[*unison.PopupMenu[explosiveChoice]](demolition.content)
		refBlank = !demolition.refField.Enabled() && demolition.refField.DrawOverCallback != nil
		weightBlank = !demolition.explosiveWeightField.Enabled() && demolition.explosiveWeightField.DrawOverCallback != nil
		countBlank = !demolition.blastCountField.Enabled() || demolition.blastCountField.DrawOverCallback != nil
	})
	if explosivePopup == nil {
		t.Fatal("the calculator must offer an explosive popup")
	}
	c.True(refBlank, "a preset explosive's REF field is disabled and blank, since what is typed in it is not used")
	c.True(weightBlank, "the weight is not used when working out the explosive a blast needs, so it is blank")
	c.False(countBlank, "the blast multiplier is in use, so it is editable and shown")
	uxtest.ChoosePopupItem(t, screen, wnd, explosivePopup, slices.IndexFunc(demolition.explosives,
		func(e explosiveChoice) bool { return e.title == "Dynamite" }))
	var damage, tnt, explosiveLabel, explosiveWeight string
	screen.Do(func() {
		demolition.blastCountField.SetText("8")
		damage = demolition.damageResult.String()
		tnt = demolition.tntResult.String()
		explosiveLabel = demolition.explosiveWeightLabel.String()
		explosiveWeight = demolition.explosiveWeightResult.String()
	})
	c.Equal("6dx8 cr ex", damage, "the blast is 6d times the multiplier")
	c.Equal("16 lb", tnt, "a 6dx8 blast takes (8x8)/4 lbs of TNT")
	c.Equal("Dynamite:", explosiveLabel, "the weight is labeled with the explosive it is of")
	c.Equal("20 lb", explosiveWeight, "the worked example on BX415 must come out at 20 lbs of dynamite")
	uxtest.CaptureScreen(t, c, screen, "demolition_calculator")

	uxtest.CloseEditorWithoutPrompt(t, screen, calc)
}
