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
	"reflect"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

// newEditorForTrait returns an editor holding the two copies of a trait's edit data that displayEditor would give it,
// without building any of the editor's UI. The trait belongs to an entity and carries weapons and a script-bearing
// modifier, so the derived values both copies would otherwise publish come from running the scripts in the data.
func newEditorForTrait() *editor[*gurps.Trait, *gurps.TraitEditData] {
	entity := gurps.NewEntity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Claws"
	trait.LocalNotes = "Sharp. <script>1 + 1</script>"
	modifier := gurps.NewTraitModifier(entity, nil, false)
	modifier.Name = "Long"
	modifier.LocalNotes = "Longer. <script>2 + 2</script>"
	trait.Modifiers = []*gurps.TraitModifier{modifier}
	trait.Weapons = []*gurps.Weapon{gurps.NewWeapon(trait, true), gurps.NewWeapon(trait, false)}
	entity.Traits = append(entity.Traits, trait)
	entity.Recalculate()

	e := &editor[*gurps.Trait, *gurps.TraitEditData]{target: trait}
	e.beforeData = &gurps.TraitEditData{}
	e.beforeData.CopyFrom(trait)
	e.editorData = &gurps.TraitEditData{}
	e.editorData.CopyFrom(trait)
	return e
}

// TestEditorIsModifiedFollowsTheDataAlone verifies that an editor reports unsaved changes when, and only when, its data
// has actually been edited. The two copies it compares are separate clones with separate script caches, and the derived
// values used to be part of the comparison: a script that exceeded the permitted per-script execution time while one
// copy was being marshaled but not the other made an untouched editor enable Apply and Cancel and prompt to save on the
// way out, purely because the machine had been busy for a moment.
func TestEditorIsModifiedFollowsTheDataAlone(t *testing.T) {
	c := check.New(t)
	e := newEditorForTrait()
	c.False(e.isModified(), "an editor whose two copies of the data agree must report no changes")

	e.editorData.Name = "Fangs"
	c.True(e.isModified(), "editing a field must be reported as a change")

	e.editorData.Name = "Claws"
	c.False(e.isModified(), "putting the original value back must clear the change")
}

// buildEditorContent fills in the content panel that displayEditor would hand to the given init function, without any
// of the docking machinery that needs a window. It returns both the editor and its content, so that a test can check
// what a widget in the content does to the editor's copy of the data.
func buildEditorContent[N gurps.Node[N], D gurps.EditorData[N]](owner Rebuildable, target N,
	initContent func(*editor[N, D], *unison.Panel) func(),
) (*editor[N, D], *unison.Panel) {
	e := &editor[N, D]{owner: owner, target: target}
	e.Self = e
	e.undoMgr = unison.NewUndoManager(100, func(_ error) {})
	e.SetLayout(&unison.FlexLayout{Columns: 1})
	reflect.ValueOf(&e.beforeData).Elem().Set(reflect.New(reflect.TypeOf(e.beforeData).Elem()))
	e.beforeData.CopyFrom(target)
	reflect.ValueOf(&e.editorData).Elem().Set(reflect.New(reflect.TypeOf(e.editorData).Elem()))
	e.editorData.CopyFrom(target)
	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	e.AddChild(content)
	initContent(e, content)
	return e, content
}

// findCheckBoxTitled returns the first checkbox bearing the given title found anywhere beneath the given panel, or nil
// if there is none.
func findCheckBoxTitled(p *unison.Panel, title string) *CheckBox {
	if box, ok := p.Self.(*CheckBox); ok && box.Text.String() == title {
		return box
	}
	for _, child := range p.Children() {
		if box := findCheckBoxTitled(child, title); box != nil {
			return box
		}
	}
	return nil
}

// findFeaturesPanel returns the first features panel found anywhere beneath the given panel, or nil if there is none.
func findFeaturesPanel(p *unison.Panel) *featuresPanel {
	if panel, ok := p.Self.(*featuresPanel); ok {
		return panel
	}
	for _, child := range p.Children() {
		if panel := findFeaturesPanel(child); panel != nil {
			return panel
		}
	}
	return nil
}

// TestTraitEditorHasSwitchedOnCheckBox verifies that both container and non-container traits offer the "Switched On"
// checkbox, since a container's modifiers can hold switchable features even though the container itself has no
// features of its own, and that the checkbox is wired to the editor's copy of the data.
func TestTraitEditorHasSwitchedOnCheckBox(t *testing.T) {
	for _, isContainer := range []bool{false, true} {
		name := "trait"
		if isContainer {
			name = "trait container"
		}
		t.Run(name, func(t *testing.T) {
			c := check.New(t)
			sheet := newTestSheetForTemplate(t)
			trait := gurps.NewTrait(sheet.Entity(), nil, isContainer)
			e, content := buildEditorContent(sheet, trait, initTraitEditor)
			box := findCheckBoxTitled(content, i18n.Text("Switched On"))
			c.NotNil(box, "expected a Switched On checkbox in the editor")
			c.False(e.editorData.SwitchedOn, "the switch starts out off")

			clickCheckBox(box, true)
			c.True(e.editorData.SwitchedOn, "checking the box must turn the switch on in the editor's data")
			c.False(trait.SwitchedOn, "the target must not be touched until the edit is applied")

			clickCheckBox(box, false)
			c.False(e.editorData.SwitchedOn, "clearing the box must turn the switch back off")
		})
	}
}

// TestSpellEditorHasFeaturesPanelAndSwitch verifies that a non-container spell can now be given features, along with
// the switch that governs the switchable ones, and that a spell container -- which has no features of its own -- gets
// neither.
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

// TestSkillEditorSwitchedOnCheckBoxOnlyForNonContainers verifies that only a skill that can hold features offers the
// switch.
func TestSkillEditorSwitchedOnCheckBoxOnlyForNonContainers(t *testing.T) {
	t.Run("skill", func(t *testing.T) {
		c := check.New(t)
		sheet := newTestSheetForTemplate(t)
		skill := gurps.NewSkill(sheet.Entity(), nil, false)
		e, content := buildEditorContent(sheet, skill, initSkillEditor)
		box := findCheckBoxTitled(content, i18n.Text("Switched On"))
		c.NotNil(box, "expected a Switched On checkbox in the skill editor")

		clickCheckBox(box, true)
		c.True(e.editorData.SwitchedOn, "checking the box must turn the switch on in the editor's data")
	})

	t.Run("skill container", func(t *testing.T) {
		c := check.New(t)
		sheet := newTestSheetForTemplate(t)
		skill := gurps.NewSkill(sheet.Entity(), nil, true)
		_, content := buildEditorContent(sheet, skill, initSkillEditor)
		c.Nil(findCheckBoxTitled(content, i18n.Text("Switched On")),
			"a skill container has nothing to switch, so it must not offer the checkbox")
	})
}
