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
