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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestTechniqueEditorKeepsAttributeDefaultNormalized verifies that the editor of a technique whose default isn't
// skill-based gives that default no name criteria, which would leave the editor modified before anything had been
// edited, and that switching the default to a skill and back leaves nothing of the skill behind.
func TestTechniqueEditorKeepsAttributeDefaultNormalized(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	technique := gurps.NewTechnique(entity, nil, "Karate")
	technique.Name = "Neck Snap"
	technique.TechniqueDefault = &gurps.SkillDefault{DefaultType: gurps.DexterityID, Modifier: -fxp.Four}
	entity.Skills = append(entity.Skills, technique)
	entity.Recalculate()
	e, content := buildEditorContent(sheet, technique, initSkillEditor)
	c.True(e.editorData.TechniqueDefault.Name.IsZero(), "an attribute default is given no name criteria")
	c.False(e.isModified(), "so the editor opens unmodified")

	var popup *unison.PopupMenu[*gurps.AttributeChoice]
	for _, one := range uxtest.PanelsOfType[*unison.PopupMenu[*gurps.AttributeChoice]](content) {
		if item, ok := one.Selected(); ok && item != nil && item.Key == gurps.DexterityID {
			popup = one
			break
		}
	}
	if popup == nil {
		t.Fatal("the technique editor has no default type popup")
	}
	choose := func(key string) {
		t.Helper()
		for i := range popup.ItemCount() {
			if item, ok := popup.ItemAt(i); ok && item != nil && item.Key == key {
				selectPopupIndex(popup, i)
				return
			}
		}
		t.Fatalf("the default type popup has no %s entry", key)
	}
	nameField := func() *StringField {
		for _, field := range uxtest.PanelsOfType[*StringField](content) {
			if field.Watermark == "Skill" {
				return field
			}
		}
		return nil
	}
	c.Nil(nameField(), "an attribute default has no skill name field")

	choose(gurps.SkillID)
	c.Equal(criteria.IsText, e.editorData.TechniqueDefault.Name.Compare, "a skill default names its skill")
	field := nameField()
	if field == nil {
		t.Fatal("a skill default has no skill name field")
	}
	field.SetText("Judo")
	c.Equal("Judo", e.editorData.TechniqueDefault.Name.Qualifier, "precondition: the field edits the name criteria")
	c.True(e.isModified(), "precondition: the technique now differs")

	choose(gurps.DexterityID)
	c.True(e.editorData.TechniqueDefault.Name.IsZero(), "switching back drops the name criteria")
	c.Nil(nameField(), "and takes the skill name field away")
	c.False(e.isModified(), "leaving the technique as it started")

	choose(gurps.SkillID)
	field = nameField()
	if field == nil {
		t.Fatal("a skill default has no skill name field")
	}
	c.Equal("", field.Text(), "the skill name field doesn't bring back the name that was dropped")
	c.Equal("", e.editorData.TechniqueDefault.Name.Qualifier)
}

// TestTechniqueEditorOpensUnmodifiedWithoutADefault verifies that the editor of a technique that has no default, as one
// loaded from a file with no "default" key has none, gives itself a skill default to edit without that counting as a
// change, while an edit to that default still does.
func TestTechniqueEditorOpensUnmodifiedWithoutADefault(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	technique := gurps.NewTechnique(entity, nil, "Karate")
	technique.Name = "Neck Snap"
	technique.TechniqueDefault = nil
	entity.Skills = append(entity.Skills, technique)
	entity.Recalculate()
	e, content := buildEditorContent(sheet, technique, initSkillEditor)
	def := e.editorData.TechniqueDefault
	if def == nil {
		t.Fatal("the editor has no default to edit")
	}
	c.Equal(gurps.SkillID, def.DefaultType, "the editor is given a skill default to edit")
	c.False(e.isModified(), "which leaves the editor unmodified")
	c.Nil(technique.TechniqueDefault, "and the technique alone")

	var field *StringField
	for _, one := range uxtest.PanelsOfType[*StringField](content) {
		if one.Watermark == "Skill" {
			field = one
			break
		}
	}
	if field == nil {
		t.Fatal("a skill default has no skill name field")
	}
	field.SetText("Karate")
	c.Equal("Karate", def.Name.Qualifier, "precondition: the field edits the default the editor was given")
	c.True(e.isModified(), "naming the skill is a change")
	field.SetText("")
	c.False(e.isModified(), "that taking the name back out undoes")
}
