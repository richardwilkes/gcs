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

// TestJumpingCalculatorResults drives the jumping calculator inside a headless workspace the way a user would: it works
// the basic example on BX352 and the extra effort of BX357 with numbers typed in after switching the source to Manual.
func TestJumpingCalculatorResults(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet := openNewCharacterSheet(t, screen)
	calc := openCalculator(t, screen)
	jumping := calc.jumping
	selectCalculatorTab(t, screen, calc, jumping)

	// The extra effort penalty is the calculator's only integer field besides the Jumping skill's.
	var extraEffortField *ux.IntegerField
	var runningStart string
	screen.Do(func() {
		for _, one := range uxtest.PanelsOfType[*ux.IntegerField](jumping.content) {
			if one != jumping.jumpingSkillField {
				extraEffortField = one
			}
		}
		runningStart = jumping.runningStartLabel.String()
	})
	if extraEffortField == nil {
		t.Fatal("the calculator must offer an extra effort penalty field")
	}
	c.Equal("yard running start", runningStart, "a sheet in yards measures the running start in yards")
	uxtest.CaptureScreen(t, c, screen, "jumping_calculator")

	results := func() (high, broad string, notes []string) {
		screen.Do(func() {
			high = jumping.highJumpResult.String()
			broad = jumping.broadJumpResult.String()
			notes = slices.DeleteFunc(uxtest.LabelTexts(jumping.notes), func(text string) bool { return text == "•" })
		})
		return high, broad, notes
	}

	// BX352: with Basic Move 5 and ST 10, a standing high jump is 6×5−10 = 20 inches and a broad jump 2×5−3 = 7 feet.
	uxtest.ChoosePopupItem(t, screen, wnd, jumping.source.popup, 0)
	screen.Do(func() {
		jumping.basicMoveField.SetText("5")
		jumping.liftingSTField.SetText("10")
		jumping.weightField.SetText("150")
	})
	high, broad, _ := results()
	c.Equal("1 foot, 8 inches", high, "the worked example on BX352 must come out at 20 inches")
	c.Equal("2 yards, 1 foot", broad, "the worked example on BX352 must come out at 7 feet")

	// BX357: extra effort at -2 adds 10% to the distances, and the note says what a failure leaves.
	screen.Do(func() { extraEffortField.SetText("-2") })
	high, broad, notes := results()
	c.Equal("1 foot, 10 inches", high, "extra effort at -2 must add 10% to the high jump")
	c.Equal("2 yards, 1 foot, 8 inches", broad, "extra effort at -2 must add 10% to the broad jump")
	c.Equal([]string{
		"The extra effort takes a Will roll, or a Will-based Jumping roll if that is better, at -2 for the +10% shown, and costs 1 FP whether it succeeds or fails. A failure leaves the jump as it would be without it: 1 foot, 8 inches high and 2 yards, 1 foot broad. A critical failure costs 1 HP of injury to a foot or leg instead and the jump fails, and on a natural 18 a HT roll is needed as well to avoid a temporary Crippled Leg (B357).",
	}, notes, "extra effort must be explained, with what a failure leaves")

	uxtest.CloseEditorWithoutPrompt(t, screen, calc)
	uxtest.CloseEditorWithoutPrompt(t, screen, sheet)
}
