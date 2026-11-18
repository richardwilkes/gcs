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
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

// findFeaturesPanel returns the first features panel found anywhere beneath the given panel.
func findFeaturesPanel(p *unison.Panel) *featuresPanel {
	panel, _ := uxtest.FirstPanelOfType[*featuresPanel](p)
	return panel
}

// A non-container spell can be given features, along with the switch that governs the switchable ones; a spell
// container, which has no features of its own, gets neither.
func TestSpellEditorHasFeaturesPanelAndSwitch(t *testing.T) {
	t.Run("spell", func(t *testing.T) {
		c := check.New(t)
		sheet := newTestSheetForTemplate(t)
		spell := gurps.NewSpell(sheet.Entity(), nil, false)
		e, content := buildEditorContent(sheet, spell, initSpellEditor)
		c.NotNil(findFeaturesPanel(content), "expected a features panel in the spell editor")
		box := findCheckBoxTitled(content, i18n.Text("Switched On"))
		c.NotNil(box, "expected a Switched On checkbox in the spell editor")

		clickCheckBox(box, true)
		c.True(e.editorData.SwitchedOn, "checking the box must turn the switch on in the editor's data")
	})

	t.Run("spell container", func(t *testing.T) {
		c := check.New(t)
		sheet := newTestSheetForTemplate(t)
		spell := gurps.NewSpell(sheet.Entity(), nil, true)
		_, content := buildEditorContent(sheet, spell, initSpellEditor)
		c.Nil(findFeaturesPanel(content), "a spell container must not offer a features panel")
		c.Nil(findCheckBoxTitled(content, i18n.Text("Switched On")),
			"a spell container has nothing to switch, so it must not offer the checkbox")
	})
}

// panelBesideLabel returns the panel that follows the label bearing the given text among the panel's direct children,
// or nil if there is none.
func panelBesideLabel(p *unison.Panel, labelText string) *unison.Panel {
	children := p.Children()
	for i, child := range children {
		if label, ok := child.Self.(*unison.Label); ok && label.String() == labelText && i+1 < len(children) {
			return children[i+1]
		}
	}
	return nil
}

// A non-container spell gets a field for each part of its rituals text, each wired to its own part; a spell container
// has no rituals, so it gets none.
func TestSpellEditorEditsEachRitualOverride(t *testing.T) {
	t.Run("spell", func(t *testing.T) {
		c := check.New(t)
		sheet := newTestSheetForTemplate(t)
		spell := gurps.NewSpell(sheet.Entity(), nil, false)
		e, content := buildEditorContent(sheet, spell, initSpellEditor)
		row := panelBesideLabel(content, i18n.Text("Rituals"))
		if row == nil {
			t.Fatal("expected a Rituals row in the spell editor")
		}
		for _, one := range []struct {
			label string
			text  string
		}{
			{i18n.Text("Ritual"), "must speak"},
			{i18n.Text("Time"), "not reduced"},
			{i18n.Text("Cost"), "-"},
		} {
			field := stringFieldLabeled(row, one.label)
			if field == nil {
				t.Fatalf("expected a field after the %s label in the Rituals row", one.label)
			}
			field.SetText(one.text)
		}
		c.Equal(gurps.RitualOverrides{Ritual: "must speak", Time: "not reduced", Cost: "-"}, e.editorData.RitualOverrides,
			"each field edits its own part")
		c.True(e.isModified(), "editing an override modifies the editor")
	})

	t.Run("spell container", func(t *testing.T) {
		c := check.New(t)
		sheet := newTestSheetForTemplate(t)
		spell := gurps.NewSpell(sheet.Entity(), nil, true)
		_, content := buildEditorContent(sheet, spell, initSpellEditor)
		c.Nil(panelBesideLabel(content, i18n.Text("Rituals")), "a spell container has no rituals to override")
	})
}
