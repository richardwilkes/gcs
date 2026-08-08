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
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

func newTestMenuKeySettingsDockable() *menuKeySettingsDockable {
	registerKeyBindingsOnce.Do(registerActions)
	d := &menuKeySettingsDockable{}
	d.Self = d
	d.initContent(unison.NewPanel())
	return d
}

// buttonTitle returns the text shown on a button, which is empty for the drawable-only reset buttons.
func buttonTitle(b *unison.Button) string {
	if b.Text == nil {
		return ""
	}
	return b.Text.String()
}

// TestMenuKeySettingsRowLayout verifies the ordering of the widgets fill() adds for each binding. The per-binding
// reset button locates the button displaying the key binding by its position relative to itself, so the documented
// [key button, title label, reset button] ordering is a contract the reset path depends on.
func TestMenuKeySettingsRowLayout(t *testing.T) {
	c := check.New(t)
	d := newTestMenuKeySettingsDockable()
	bindings := gurps.CurrentBindings()
	c.True(len(bindings) > 1, "there must be more than one key binding to test with")
	children := d.content.Children()
	c.Equal(len(bindings)*3, len(children),
		"each binding must add exactly three widgets: the key button, the title label and the reset button")
	for i, binding := range bindings {
		keyButton, ok := children[i*3].Self.(*unison.Button)
		c.True(ok, "child %d must be the key binding button for %q", i*3, binding.ID)
		c.Equal(binding.KeyBinding.String(), buttonTitle(keyButton),
			"the key binding button for %q must show its key binding", binding.ID)
		label, ok := children[i*3+1].Self.(*unison.Label)
		c.True(ok, "child %d must be the title label for %q", i*3+1, binding.ID)
		c.Equal(binding.Action.Title, label.String(), "the label for %q must show the action title", binding.ID)
		resetButton, ok := children[i*3+2].Self.(*unison.Button)
		c.True(ok, "child %d must be the reset button for %q", i*3+2, binding.ID)
		c.Equal("", buttonTitle(resetButton), "the reset button for %q must be drawable-only", binding.ID)
	}
}

// TestMenuKeySettingsResetButtonTargetsItsOwnKeyButton verifies that the sibling lookup used by the reset button finds
// the key binding button of its own row. It once landed one position early, on the title label, so the type assertion
// always failed and the displayed key binding was never refreshed.
func TestMenuKeySettingsResetButtonTargetsItsOwnKeyButton(t *testing.T) {
	c := check.New(t)
	d := newTestMenuKeySettingsDockable()
	children := d.content.Children()
	for i := range gurps.CurrentBindings() {
		keyButton, ok := children[i*3].Self.(*unison.Button)
		c.True(ok, "child %d must be a key binding button", i*3)
		resetButton, ok := children[i*3+2].Self.(*unison.Button)
		c.True(ok, "child %d must be a reset button", i*3+2)
		c.Equal(keyButton, bindingButtonForResetButton(resetButton),
			"the reset button in row %d must target the key binding button of that same row", i)
	}
}

// TestBindingButtonForResetButtonWithoutRow verifies that the sibling lookup reports failure rather than panicking or
// returning an unrelated widget when the reset button is not part of a complete row.
func TestBindingButtonForResetButtonWithoutRow(t *testing.T) {
	c := check.New(t)

	orphan := unison.NewButton()
	c.Equal((*unison.Button)(nil), bindingButtonForResetButton(orphan), "a parentless button has no row")

	parent := unison.NewPanel()
	first := unison.NewButton()
	parent.AddChild(first)
	c.Equal((*unison.Button)(nil), bindingButtonForResetButton(first), "nothing precedes the first child")
	second := unison.NewButton()
	parent.AddChild(second)
	c.Equal((*unison.Button)(nil), bindingButtonForResetButton(second), "only one widget precedes the second child")
	third := unison.NewButton()
	parent.AddChild(third)
	c.Equal(first, bindingButtonForResetButton(third), "the third child's row starts at the first")

	// A row whose expected key binding slot holds something other than a button must not be returned.
	other := unison.NewPanel()
	other.AddChild(unison.NewLabel())
	other.AddChild(unison.NewLabel())
	trailing := unison.NewButton()
	other.AddChild(trailing)
	c.Equal((*unison.Button)(nil), bindingButtonForResetButton(trailing), "a non-button in the key slot yields nil")
}

// TestMenuKeySettingsResetUpdatesDisplayedBinding verifies that resetting a single binding refreshes the key shown on
// that binding's own button, and leaves the other rows alone.
func TestMenuKeySettingsResetUpdatesDisplayedBinding(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	bindings := gurps.CurrentBindings()
	c.True(len(bindings) > 1, "there must be more than one key binding to test with")

	// Use a binding other than the first so that an off-by-one lookup can't accidentally land on the right widget.
	target := bindings[1]
	defaultBinding := target.KeyBinding
	custom := unison.KeyBinding{KeyCode: unison.KeyF19, Modifiers: mod.Shift | mod.Command}
	c.NotEqual(defaultBinding.String(), custom.String(), "the test binding must differ from the default")
	g := gurps.GlobalSettings()
	g.KeyBindings.Set(target.ID, custom)
	g.KeyBindings.MakeCurrent()
	t.Cleanup(func() {
		g.KeyBindings.ResetOne(target.ID)
		g.KeyBindings.MakeCurrent()
	})

	d := newTestMenuKeySettingsDockable()
	current := gurps.CurrentBindings()
	i := slices.IndexFunc(current, func(b *gurps.Binding) bool { return b.ID == target.ID })
	c.True(i >= 0, "the target binding must be present")
	children := d.content.Children()
	keyButton, ok := children[i*3].Self.(*unison.Button)
	c.True(ok, "child %d must be a key binding button", i*3)
	resetButton, ok := children[i*3+2].Self.(*unison.Button)
	c.True(ok, "child %d must be a reset button", i*3+2)
	c.Equal(custom.String(), buttonTitle(keyButton), "the customized binding must be displayed before the reset")

	before := make([]string, len(children))
	for j, child := range children {
		if b, isButton := child.Self.(*unison.Button); isButton {
			before[j] = buttonTitle(b)
		}
	}

	d.resetBinding(current[i], resetButton)

	c.Equal(defaultBinding.String(), buttonTitle(keyButton), "the reset must refresh the displayed key binding")
	c.Equal(defaultBinding, g.KeyBindings.Current(target.ID), "the reset must restore the factory binding")
	for j, child := range children {
		if j == i*3 {
			continue
		}
		if b, isButton := child.Self.(*unison.Button); isButton {
			c.Equal(before[j], buttonTitle(b), "resetting one binding must not alter the button at index %d", j)
		}
	}
}
