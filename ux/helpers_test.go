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

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

func TestNoModifiersDown(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name string
		mods mod.Modifiers
		want bool
	}{
		{name: "none", mods: mod.None, want: true},
		// A latched CapsLock or NumLock must be ignored so that plain keypresses are still recognized.
		{name: "caps lock latched", mods: mod.CapsLock, want: true},
		{name: "num lock latched", mods: mod.NumLock, want: true},
		{name: "both lock keys latched", mods: mod.CapsLock | mod.NumLock, want: true},
		// Any non-sticky modifier being down means it is not a plain keypress.
		{name: "shift", mods: mod.Shift, want: false},
		{name: "control", mods: mod.Control, want: false},
		{name: "option", mods: mod.Option, want: false},
		{name: "command", mods: mod.Command, want: false},
		// A real modifier still counts even when combined with a latched lock key.
		{name: "shift with caps lock", mods: mod.Shift | mod.CapsLock, want: false},
		{name: "command with num lock", mods: mod.Command | mod.NumLock, want: false},
	} {
		c.Equal(one.want, noModifiersDown(one.mods), one.name)
	}
}

// newFocusablePanel returns a plain, focusable, non-button panel usable in a headless test.
func newFocusablePanel() *unison.Panel {
	p := unison.NewPanel()
	p.SetFocusable(true)
	return p
}

// TestFirstContentFocusTarget verifies that the focus target for a newly-opened dockable is chosen deterministically
// from the content and toolbar subtrees, preferring the first non-button focusable widget in the content, then any
// focusable widget in the content, and finally a focusable widget in the toolbar.
func TestFirstContentFocusTarget(t *testing.T) {
	c := check.New(t)

	t.Run("prefers first non-button focusable in content", func(_ *testing.T) {
		toolbar := unison.NewPanel()
		toolbar.AddChild(unison.NewButton())
		content := unison.NewPanel()
		content.AddChild(unison.NewButton())
		field := newFocusablePanel()
		content.AddChild(field)
		content.AddChild(newFocusablePanel())
		c.Equal(field, firstContentFocusTarget(toolbar, content))
	})

	t.Run("falls back to first focusable button in content", func(_ *testing.T) {
		toolbar := unison.NewPanel()
		toolbar.AddChild(newFocusablePanel())
		content := unison.NewPanel()
		firstButton := unison.NewButton()
		content.AddChild(firstButton)
		content.AddChild(unison.NewButton())
		c.Equal(firstButton.AsPanel(), firstContentFocusTarget(toolbar, content))
	})

	t.Run("falls back to toolbar when content has no focusables", func(_ *testing.T) {
		toolbar := unison.NewPanel()
		toolbar.AddChild(unison.NewPanel()) // not focusable
		tbTarget := newFocusablePanel()
		toolbar.AddChild(tbTarget)
		content := unison.NewPanel()
		content.AddChild(unison.NewPanel()) // not focusable
		c.Equal(tbTarget, firstContentFocusTarget(toolbar, content))
	})

	t.Run("returns nil when nothing is focusable", func(_ *testing.T) {
		toolbar := unison.NewPanel()
		toolbar.AddChild(unison.NewPanel())
		content := unison.NewPanel()
		content.AddChild(unison.NewPanel())
		c.Equal((*unison.Panel)(nil), firstContentFocusTarget(toolbar, content))
	})

	t.Run("walks content depth-first in pre-order", func(_ *testing.T) {
		toolbar := unison.NewPanel()
		content := unison.NewPanel()
		branch := unison.NewPanel() // not focusable itself, but holds the earliest focusable
		deep := newFocusablePanel()
		branch.AddChild(deep)
		content.AddChild(branch)
		content.AddChild(newFocusablePanel()) // later in pre-order than deep
		c.Equal(deep, firstContentFocusTarget(toolbar, content))
	})

	t.Run("skips hidden and disabled widgets", func(_ *testing.T) {
		toolbar := unison.NewPanel()
		content := unison.NewPanel()
		hidden := newFocusablePanel()
		hidden.Hidden = true
		content.AddChild(hidden)
		disabled := newFocusablePanel()
		disabled.SetEnabled(false)
		content.AddChild(disabled)
		reachable := newFocusablePanel()
		content.AddChild(reachable)
		c.Equal(reachable, firstContentFocusTarget(toolbar, content))
	})

	t.Run("returns the content root when it is itself focusable", func(_ *testing.T) {
		toolbar := unison.NewPanel()
		content := newFocusablePanel()
		content.AddChild(newFocusablePanel())
		c.Equal(content, firstContentFocusTarget(toolbar, content))
	})

	t.Run("descends past a focusable button root during the non-button scan", func(_ *testing.T) {
		toolbar := unison.NewPanel()
		content := unison.NewPanel()
		buttonRoot := unison.NewButton()
		field := newFocusablePanel()
		buttonRoot.AddChild(field)
		content.AddChild(buttonRoot)
		c.Equal(field, firstContentFocusTarget(toolbar, content))
	})
}
