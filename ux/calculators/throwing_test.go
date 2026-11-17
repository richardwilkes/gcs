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

	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestThrowingCalculatorResults drives the throwing calculator inside a headless workspace the way a user would: with
// no sheet to draw on, it works the example on BX355 and the extra effort of BX357 with numbers typed in, and checks
// that Throwing Art adds to the distance and damage.
func TestThrowingCalculatorResults(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	calc := openCalculator(t, screen)
	throwing := calc.throwing
	selectCalculatorTab(t, screen, calc, throwing)

	// The extra effort penalty is the calculator's only integer field.
	var extraEffortField *ux.IntegerField
	screen.Do(func() { extraEffortField, _ = uxtest.FirstPanelOfType[*ux.IntegerField](throwing.content) })
	if extraEffortField == nil {
		t.Fatal("the calculator must offer an extra effort penalty field")
	}
	results := func() (distance, damage string, notes []string) {
		screen.Do(func() {
			distance = throwing.distanceResult.String()
			damage = throwing.damageResult.String()
			notes = slices.DeleteFunc(uxtest.LabelTexts(throwing.notes), func(text string) bool { return text == "•" })
		})
		return distance, damage, notes
	}

	// BX355: ST 10 throws a 1 lb object, a twentieth of its Basic Lift of 20, 3.5×ST = 35 yards, for thrust−2 per die:
	// 1d−2 becomes 1d−4.
	screen.Do(func() {
		throwing.stField.SetText("10")
		throwing.strikingSTField.SetText("10")
		throwing.weightField.SetText("1")
	})
	distance, damage, _ := results()
	c.Equal("35 yards", distance, "the worked example on BX355 must come out at 35 yards")
	c.Equal("1d-4", damage, "a light object does thrust−2 per die")
	uxtest.CaptureScreen(t, c, screen, "throwing_calculator")

	// BX357: extra effort at -2 raises the ST for both distance and damage by 10%, to 11: 11 × 3.5 = 38.5 yards, and
	// thrust for ST 11 is 1d-1, less 2 for the light object.
	screen.Do(func() { extraEffortField.SetText("-2") })
	distance, damage, notes := results()
	c.Equal("38 yards, 1 foot, 6 inches", distance, "extra effort at -2 must add 10% to the ST for distance")
	c.Equal("1d-3", damage, "extra effort at -2 must add 10% to the ST for damage")
	c.Equal([]string{
		"The extra effort takes a Will roll, or a Will-based Throwing roll if that is better, at -2 for the +10% ST shown, and costs 1 FP whether it succeeds or fails. A failure leaves the throw as it would be without it: 35 yards for 1d-4. A critical failure costs 1 HP of injury instead and the throw fails, and on a natural 18 a HT roll is needed as well to avoid a temporary disadvantage (B357).",
	}, notes, "extra effort must be explained, with what a failure leaves")
	screen.Do(func() { extraEffortField.SetText("0") })

	// Throwing Art at DX+1 adds 2 to the ST for distance and 2 per die to the damage: 12×3.5 = 42 yards and 1d−2.
	uxtest.ChoosePopupItem(t, screen, wnd, throwing.throwingArtPopup, 2)
	distance, damage, _ = results()
	c.Equal("42 yards", distance, "Throwing Art at DX+1 must add 2 to the ST the distance is worked out from")
	c.Equal("1d-2", damage, "Throwing Art at DX+1 must add 2 per die to the damage")

	uxtest.CloseEditorWithoutPrompt(t, screen, calc)
}
