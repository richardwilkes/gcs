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
	"strconv"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// focusForReadingSetter returns a function that sets the FocusForReading general setting, passes it to unison and
// describes the window, which also activates accessibility. Both are restored when the test ends.
func focusForReadingSetter(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window) func(enabled bool) {
	t.Helper()
	gs := gurps.GlobalSettings().General
	swapForTest(t, &gs.FocusForReading, false)
	// Restored on the UI thread: from elsewhere, with accessibility active, unison queues a task to describe the
	// windows again, and a task still queued at Stop runs in the next test.
	saved := unison.FocusForReading()
	t.Cleanup(func() { screen.Do(func() { unison.SetFocusForReading(saved) }) })
	return func(enabled bool) {
		t.Helper()
		screen.Do(func() {
			gs.FocusForReading = enabled
			gs.UpdateFocusForReading()
		})
		if screen.AccessibilityTree(wnd) == nil {
			t.Fatal("the window must be described")
		}
	}
}

// With a screen reader listening and the FocusForReading setting on, a disabled control, such as the hiking
// calculator's Move while a sheet supplies it, is a tab stop. It stays disabled otherwise: a screen reader is told it
// is unavailable, typing does not change it, and while focused it keeps showing the value it is given.
func TestDisabledControlsJoinTheTabOrderForScreenReaders(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	setFocusForReading := focusForReadingSetter(t, screen, wnd)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	screen.Do(func() { DisplayCalculator(sheet) })
	calc := soleEditor[*Calculator](t, screen, func(d unison.Dockable) bool {
		_, isCalculator := d.AsPanel().Self.(*Calculator)
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
	screen.Do(func() { target = firstContentFocusTarget(unison.NewPanel(), hiking.content) })
	c.NotNil(target, "the calculator holds controls that can be used")
	if target != nil {
		c.True(target.Enabled(), "the focus must start on a control that can be used, not on a %T", target.Self)
	}

	setFocusForReading(false)
	c.False(focusable(move), "a disabled field must not be a tab stop once the setting is off again")
	c.True(focusable(rest))
}

// firstContentFocusTarget prefers a control that can be used, in the content and then the toolbar, over a disabled one,
// and picks a disabled control only when nothing else can take the focus, so a screen reader has somewhere to start.
func TestFirstFocusLeavesDisabledControlsUntilLast(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	setFocusForReading := focusForReadingSetter(t, screen, wnd)
	var holder, toolbar, content *unison.Panel
	var disabledField, disabledButton, button *unison.Panel
	var toolbarDisabledButton, toolbarButton *unison.Panel
	screen.Do(func() {
		holder = unison.NewPanel()
		toolbar = unison.NewPanel()
		content = unison.NewPanel()
		holder.AddChild(toolbar)
		holder.AddChild(content)
		wnd.Content().AddChild(holder)
		disabled := func(p unison.Paneler) *unison.Panel {
			p.AsPanel().SetEnabled(false)
			return p.AsPanel()
		}
		disabledField = disabled(unison.NewField())
		disabledButton = disabled(unison.NewButton())
		button = unison.NewButton().AsPanel()
		toolbarDisabledButton = disabled(unison.NewButton())
		toolbarButton = unison.NewButton().AsPanel()
	})
	t.Cleanup(func() { screen.Do(holder.RemoveFromParent) })
	target := func(toolbarChildren, contentChildren []*unison.Panel) (result *unison.Panel) {
		screen.Do(func() {
			toolbar.RemoveAllChildren()
			content.RemoveAllChildren()
			for _, one := range toolbarChildren {
				toolbar.AddChild(one)
			}
			for _, one := range contentChildren {
				content.AddChild(one)
			}
			result = firstContentFocusTarget(toolbar, content)
		})
		return result
	}
	all := []*unison.Panel{disabledField, disabledButton, button}
	bar := []*unison.Panel{toolbarDisabledButton, toolbarButton}

	setFocusForReading(true)
	c.Equal(button, target(bar, all), "a button that can be used comes ahead of the disabled controls before it")
	c.Equal(toolbarButton, target(bar, all[:2]),
		"with nothing in the content that can be used, the toolbar's first button that can be is next")
	c.Equal(disabledField, target(bar[:1], all[:2]),
		"with nothing anywhere that can be used, the first disabled control of the content is where to start")
	c.Equal(toolbarDisabledButton, target(bar[:1], nil), "and failing that the toolbar's")

	// With the setting off, a disabled control takes no focus at all.
	setFocusForReading(false)
	c.Equal(button, target(bar, all))
	c.Equal(toolbarButton, target(bar, all[:2]))
	c.Nil(target(bar[:1], all[:2]), "nothing can take the focus")
}
