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

	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestEveryCalculatorControlHasAnAccessibleName opens the calculators with a character sheet active and fails for
// every control on any of their tabs that an assistive technology would be handed with no name; see
// uxtest.AXNameAudit. ux's own audit covers the rest of the workspace.
func TestEveryCalculatorControlHasAnAccessibleName(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	audit := uxtest.NewAXNameAudit(t, screen, wnd)
	sheet := uxtest.OpenNewCharacterSheet(t, screen)
	calc, isCalc := audit.Open(func() { Display(sheet) }).(*Dockable)
	if !isCalc {
		t.Fatal("the calculators did not open")
	}
	for i := range calc.tabs {
		var title string
		screen.Do(func() {
			calc.tabBar.SelectTab(i)
			title = calc.tabs[i].title()
		})
		audit.Check("calculator: "+title, calc)
	}
	audit.Report()
}
