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
	"strconv"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// With a screen reader listening and the FocusForReading setting on, a disabled control, such as the hiking
// calculator's Move while a sheet supplies it, is a tab stop. It stays disabled otherwise: a screen reader is told it
// is unavailable, typing does not change it, and while focused it keeps showing the value it is given.
func TestDisabledControlsJoinTheTabOrderForScreenReaders(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	setFocusForReading := uxtest.FocusForReadingSetter(t, screen, wnd)
	sheet := uxtest.OpenNewCharacterSheet(t, screen)
	screen.Do(func() { Display(sheet) })
	calc := uxtest.SoleEditor[*Dockable](t, screen, func(d unison.Dockable) bool {
		_, isCalculator := d.AsPanel().Self.(*Dockable)
		return isCalculator
	})
	hiking := calc.hiking
	selectCalculatorTab(t, screen, calc, hiking)
	move := hiking.moveField
	rest := hiking.restField
	focusable := func(p unison.Paneler) (result bool) {
		screen.Do(func() { result = p.AsPanel().Focusable() })
		return result
	}
	enabled := func(p unison.Paneler) (result bool) {
		screen.Do(func() { result = p.AsPanel().Enabled() })
		return result
	}
	focus := func() (focus *unison.Panel) {
		screen.Do(func() { focus = wnd.Focus() })
		return focus
	}
	text := func() (result string) {
		screen.Do(func() { result = move.Text() })
		return result
	}
	c.False(enabled(move), "the sheet supplies the Move, so its field must be disabled")
	c.True(enabled(rest), "the rest is not the sheet's to supply, so its field must be enabled")

	setFocusForReading(false)
	c.False(focusable(move), "a disabled field must not be a tab stop while the setting is off")
	c.True(focusable(rest), "a field that can be used is always a tab stop")

	setFocusForReading(true)
	c.True(focusable(move), "a disabled field must be a tab stop while the setting is on")
	c.True(focusable(rest))
	c.False(enabled(move), "taking the focus must not enable the field")
	node := screen.AccessibilityNodeFor(move)
	if node == nil {
		t.Fatal("the Move field must be described")
	}
	c.True(node.Disabled, "a screen reader must be told the field is unavailable")
	before := text()
	c.NotEqual("", before, "the sheet must have supplied a Move")
	screen.Do(move.RequestFocus)
	c.Equal(move.AsPanel(), focus(), "the disabled field takes the focus when asked to")
	screen.Type("9")
	c.Equal(before, text(), "nothing typed on a disabled field may change it")
	c.Equal(move.AsPanel(), focus(), "and the focus stays where it was")
	var want string
	screen.Do(func() {
		entity := sheet.Entity()
		attr := entity.Attributes.Set[gurps.BasicMoveID]
		attr.SetMaximum(attr.Maximum() + fxp.Three)
		entity.Recalculate()
		hiking.sheetChanged(sheet)
		want = strconv.Itoa(hiking.move)
	})
	c.NotEqual(before, want, "the sheet's Move must have changed")
	c.Equal(want, text(), "a disabled field holding the focus must show the number it is given")
	c.Equal(move.AsPanel(), focus())

	screen.Do(rest.RequestFocus)
	c.Equal(rest.AsPanel(), focus())
	c.Equal(want, text(), "the focus leaving a disabled field must leave it as it was")
	var got string
	screen.Do(func() { got = strconv.Itoa(hiking.move) })
	c.Equal(want, got, "and the number behind it too")

	var target *unison.Panel
	screen.Do(func() {
		ux.FocusFirstContent(unison.NewPanel(), hiking.content)
		target = wnd.Focus()
	})
	c.NotNil(target, "the calculator holds controls that can be used")
	if target != nil {
		c.True(target.Enabled(), "the focus must start on a control that can be used, not on a %T", target.Self)
	}

	setFocusForReading(false)
	c.False(focusable(move), "a disabled field must not be a tab stop once the setting is off again")
	c.True(focusable(rest))
}
