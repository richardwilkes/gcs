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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/calculator"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestScatterCalculatorResults drives the scatter calculator in a headless workspace: the worked example on BX414 must
// come out, both for a plain miss and for one whose margin is squared.
func TestScatterCalculatorResults(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	calc := openCalculator(t, screen)
	scatter := calc.scatter
	selectCalculatorTab(t, screen, calc, scatter)

	// BX414: an attack that misses by 3 at 10 yards scatters 3 yards, and one that misses by 3 with the miss squared
	// would scatter 9 but is limited to half the distance.
	var causePopup *unison.PopupMenu[calculator.ScatterCause]
	var scattered string
	screen.Do(func() {
		causePopup, _ = uxtest.FirstPanelOfType[*unison.PopupMenu[calculator.ScatterCause]](scatter.content)
		scatter.marginField.SetText("3")
		scattered = scatter.result.String()
	})
	if causePopup == nil {
		t.Fatal("the calculator must offer a cause popup")
	}
	c.Equal("3 yards", scattered, "a miss scatters by its margin")
	uxtest.ChoosePopupItem(t, screen, wnd, causePopup, 1)
	screen.Do(func() { scattered = scatter.result.String() })
	c.Equal("5 yards (limited to half the distance)", scattered, "a squared miss is limited to half the distance")
	uxtest.CaptureScreen(t, c, screen, "scatter_calculator")

	uxtest.CloseEditorWithoutPrompt(t, screen, calc)
}
