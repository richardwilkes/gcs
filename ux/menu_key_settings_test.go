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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// newTestMenuKeySettingsDockable builds the view's toolbar and content without placing it in a dock.
func newTestMenuKeySettingsDockable() *menuKeySettingsDockable {
	registerKeyBindingsOnce.Do(registerActions)
	d := &menuKeySettingsDockable{}
	d.Self = d
	d.addToStartToolbar(unison.NewPanel())
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

// restoreKeyBindings puts the bindings with the IDs back to their factory defaults when the test ends.
func restoreKeyBindings(t *testing.T, ids ...string) {
	t.Cleanup(func() {
		g := gurps.GlobalSettings()
		for _, id := range ids {
			g.KeyBindings.ResetOne(id)
		}
		g.KeyBindings.MakeCurrent()
	})
}

// TestMenuKeyFilterFieldNarrowsRows verifies that typing in the toolbar's filter field rebuilds the rows with only the
// bindings it matches, in only the sections that keep any, and that clearing it brings every row back.
func TestMenuKeyFilterFieldNarrowsRows(t *testing.T) {
	c := check.New(t)
	d := newTestMenuKeySettingsDockable()
	all := len(gurps.CurrentBindings())
	c.Equal(all, len(d.rows), "every binding has a row before anything is typed")
	d.filterField.SetText("Select All")
	c.Equal(1, len(d.rows), "only the matching binding has a row")
	_, ok := d.rows["select.all"]
	c.True(ok, "the matching binding's row is the one kept")
	c.Equal(1, len(d.content.Children()), "only the section holding the match is shown")
	d.filterField.SetText("")
	c.Equal(all, len(d.rows), "clearing the filter brings every row back")
}

// TestMenuKeySectionColumnsAlignAcrossBlocks verifies that a section's labels all end at the same place, those in a
// submenu's block being narrower by the block's indent, and that its key buttons share the width of the widest, so that
// the key and reset columns of the block line up with those of the section's own rows despite being laid out in a grid
// of their own.
func TestMenuKeySectionColumnsAlignAcrossBlocks(t *testing.T) {
	c := check.New(t)
	d := newTestMenuKeySettingsDockable()
	own := d.rows["open"]            // In the File section's own grid
	inBlock := d.rows["new.eqm.lib"] // In its New block, with its widest label
	block := inBlock.label.Parent()
	c.True(own.label.Parent() != block, "the rows must be in different grids")
	indent := block.Border().Insets().Left
	c.True(indent > 0, "the block must be inset")
	var labelEnd float32
	for _, row := range d.rows {
		_, pref, _ := row.label.Sizes(geom.Size{})
		switch row.label.Parent() {
		case own.label.Parent():
			labelEnd = max(labelEnd, pref.Width)
		case block:
			labelEnd = max(labelEnd, indent+pref.Width)
		}
	}
	minWidth := func(p unison.Paneler) float32 {
		data, ok := p.AsPanel().LayoutData().(*unison.FlexLayoutData)
		c.True(ok, "the row's widgets must carry flex layout data")
		return data.MinSize.Width
	}
	c.Equal(labelEnd, minWidth(own.label), "a label in the section's own grid reaches the common end")
	c.Equal(labelEnd-indent, minWidth(inBlock.label), "a label in the block reaches it too, from its inset start")
	c.True(minWidth(own.keyButton) > 0, "the key buttons are widened too")
	c.Equal(minWidth(own.keyButton), minWidth(inBlock.keyButton), "to the same width in and out of the block")
}

// TestMenuKeyRowShowsCustomizationAndResets verifies that a row's reset button is enabled only while its binding differs
// from the factory default, that the key button shows the current key, or says there is none, and that resetting the
// row restores the default.
func TestMenuKeyRowShowsCustomizationAndResets(t *testing.T) {
	c := check.New(t)
	d := newTestMenuKeySettingsDockable()
	g := gurps.GlobalSettings()
	// Save has a default key and Add as Grant has none, so between them the rows show both states from the start.
	const id = "save"
	restoreKeyBindings(t, id)
	row := d.rows[id]
	factory := gurps.FactoryKeyBinding(id)
	c.False(factory.IsZero(), "the test needs a binding with a default key")
	c.Equal(factory.String(), buttonTitle(row.keyButton), "the key button shows the default key")
	c.False(row.resetButton.Enabled(), "a binding at its default has nothing to reset")
	unbound := d.rows["add.as.grant"]
	c.True(unbound.binding.KeyBinding.IsZero(), "the test needs a binding with no default key")
	c.Equal("None", buttonTitle(unbound.keyButton), "a binding with no key says so")
	c.False(unbound.resetButton.Enabled(), "a binding at its default has nothing to reset, even when that is no key")

	custom := unison.KeyBinding{KeyCode: unison.KeyF19, Modifiers: mod.Shift | mod.Command}
	d.applyBinding(row, custom)
	c.Equal(custom, g.KeyBindings.Current(id), "the key is applied to the settings")
	c.Equal(custom.String(), buttonTitle(row.keyButton), "the key button shows the new key")
	c.True(row.resetButton.Enabled(), "a customized binding can be reset")

	d.applyBinding(row, unison.KeyBinding{})
	c.Equal("None", buttonTitle(row.keyButton), "a cleared binding says it has no key")
	c.True(row.resetButton.Enabled(), "a cleared binding can be reset")

	d.resetBinding(row)
	c.Equal(factory, g.KeyBindings.Current(id), "the reset restores the factory default")
	c.Equal(factory.String(), buttonTitle(row.keyButton), "the key button shows the default key again")
	c.False(row.resetButton.Enabled(), "the reset binding has nothing to reset")
}

// TestMenuKeyResetRestoresEveryBinding verifies that resetting the view puts every binding back to its factory default,
// in the settings and in the menus, and rebuilds the rows to show it.
func TestMenuKeyResetRestoresEveryBinding(t *testing.T) {
	c := check.New(t)
	d := newTestMenuKeySettingsDockable()
	g := gurps.GlobalSettings()
	ids := []string{"save", "copy"}
	restoreKeyBindings(t, ids...)
	d.applyBinding(d.rows[ids[0]], unison.KeyBinding{KeyCode: unison.KeyF19, Modifiers: mod.Shift | mod.Command})
	d.applyBinding(d.rows[ids[1]], unison.KeyBinding{})

	d.reset()
	inMenus := make(map[string]unison.KeyBinding)
	for _, b := range gurps.CurrentBindings() {
		inMenus[b.ID] = b.KeyBinding
	}
	for _, id := range ids {
		factory := gurps.FactoryKeyBinding(id)
		c.Equal(factory, g.KeyBindings.Current(id), "%s: the settings hold the default", id)
		c.Equal(factory, inMenus[id], "%s: the menus use the default", id)
		c.Equal(factory.String(), buttonTitle(d.rows[id].keyButton), "%s: the key button shows the default", id)
		c.False(d.rows[id].resetButton.Enabled(), "%s: there is nothing left to reset", id)
	}
}

// TestMenuKeyApplyBindingTakesKeyFromOtherBindings verifies that the dialog warns of the other bindings a key is set to,
// naming them, that applying the key takes it from them and updates their rows, and that neither the binding's own
// key nor an empty one counts as a conflict.
func TestMenuKeyApplyBindingTakesKeyFromOtherBindings(t *testing.T) {
	c := check.New(t)
	d := newTestMenuKeySettingsDockable()
	g := gurps.GlobalSettings()
	restoreKeyBindings(t, "copy", "paste")
	copyRow := d.rows["copy"]
	pasteRow := d.rows["paste"]
	copyKey := copyRow.binding.KeyBinding
	c.False(copyKey.IsZero(), "the test needs a binding with a default key")

	others := conflictingBindings("paste", copyKey)
	c.Equal(1, len(others), "one other binding has the key")
	c.Equal("copy", others[0].ID, "the binding that has the key is named")
	c.Equal("Already used by Copy. Set will remove it from there.", conflictWarning("paste", copyKey))
	c.Nil(conflictingBindings("copy", copyKey), "a binding does not conflict with itself")
	c.Nil(conflictingBindings("paste", unison.KeyBinding{}), "no key conflicts with nothing")
	c.Equal("", conflictWarning("copy", copyKey), "nothing to warn of")

	d.applyBinding(pasteRow, copyKey)
	c.Equal(copyKey, g.KeyBindings.Current("paste"), "the key is applied")
	c.True(g.KeyBindings.Current("copy").IsZero(), "the key is taken from the binding that had it")
	c.Equal("None", buttonTitle(copyRow.keyButton), "the row that lost the key shows it has none")
	c.True(copyRow.resetButton.Enabled(), "the row that lost the key can be reset")
	c.Equal(copyKey.String(), buttonTitle(pasteRow.keyButton), "the row given the key shows it")
}

// TestMenuKeyFreedKeyReturnsToItsFactoryBinding verifies that a binding whose factory key the user gave another binding
// shows no key, and that its row shows the factory key again once that other binding lets the key go.
func TestMenuKeyFreedKeyReturnsToItsFactoryBinding(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	g := gurps.GlobalSettings()
	restoreKeyBindings(t, "save", "add.as.grant")
	saveKey := gurps.FactoryKeyBinding("save")
	c.False(saveKey.IsZero(), "the test needs a binding with a default key")
	g.KeyBindings.Set("add.as.grant", saveKey) // As if chosen before Save had the key
	g.KeyBindings.MakeCurrent()
	d := newTestMenuKeySettingsDockable()
	c.Equal("None", buttonTitle(d.rows["save"].keyButton), "the factory key the user gave another binding isn't used")

	d.applyBinding(d.rows["add.as.grant"], unison.KeyBinding{})
	c.Equal(saveKey, g.KeyBindings.Current("save"), "the factory key is used once freed")
	c.Equal(saveKey.String(), buttonTitle(d.rows["save"].keyButton), "and its row shows it")
}

// TestClassifyCapturedKey verifies which key presses the capture dialog takes as the new combination, which it waits
// through, and which it leaves to the dialog itself.
func TestClassifyCapturedKey(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name string
		key  unison.KeyCode
		mods mod.Modifiers
		want capturedKeyUse
	}{
		{name: "letter with modifier", key: unison.KeyA, mods: mod.Control, want: keyIsCombination},
		{name: "function key alone", key: unison.KeyF5, want: keyIsCombination},
		{name: "left shift alone", key: unison.KeyLShift, mods: mod.Shift, want: keyIsModifier},
		{name: "right command alone", key: unison.KeyRCommand, mods: mod.Command, want: keyIsModifier},
		{name: "caps lock alone", key: unison.KeyCapsLock, mods: mod.CapsLock, want: keyIsModifier},
		{name: "bare escape", key: unison.KeyEscape, want: keyDrivesDialog},
		{name: "escape with caps lock on", key: unison.KeyEscape, mods: mod.CapsLock, want: keyDrivesDialog},
		{name: "escape with modifier", key: unison.KeyEscape, mods: mod.Control, want: keyIsCombination},
		{name: "bare return", key: unison.KeyReturn, want: keyDrivesDialog},
		{name: "bare numeric keypad enter", key: unison.KeyNumPadEnter, want: keyDrivesDialog},
		{name: "return with modifier", key: unison.KeyReturn, mods: mod.Option, want: keyIsCombination},
		{name: "bare tab", key: unison.KeyTab, want: keyDrivesDialog},
		{name: "shift tab", key: unison.KeyTab, mods: mod.Shift, want: keyDrivesDialog},
		{name: "tab with modifier", key: unison.KeyTab, mods: mod.Control, want: keyIsCombination},
	} {
		c.Equal(one.want, classifyCapturedKey(one.key, one.mods), one.name)
	}
}
