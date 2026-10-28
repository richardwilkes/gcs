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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantadjust"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
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

// levelPreview returns the text of the editor's level field, which sits beside the Level label.
func levelPreview(content *unison.Panel) string {
	for _, f := range uxtest.PanelsOfType[*NonEditableField](content) {
		if parent := f.Parent(); parent != nil {
			for _, sibling := range parent.Children() {
				if l, ok := sibling.Self.(*unison.Label); ok && l.Text.String() == "Level" {
					return f.Text.String()
				}
			}
		}
	}
	return ""
}

// TestEditorLevelPreviewIncludesGrants verifies that the skill and spell editors' level previews include what grants
// gave the item, a technique included, as the sheet's level does, so that a granted row's editor shows its level rather
// than none, that they read the level a Keep Highest or Keep Lowest grant keeps as the sheet reads it: without the
// skill's default, held to a technique's limit, and with its relative level, and that they merge the grants for the
// item as edited, as applying the edits does.
func TestEditorLevelPreviewIncludesGrants(t *testing.T) {
	// grantOn puts on the sheet a trait carrying a grant of the item, which adjusts it as given.
	grantOn := func(e *gurps.Entity, name string, item any, adjust grantadjust.Kind, amount fxp.Int,
		merge grantmerge.Mode,
	) {
		trait := gurps.NewTrait(e, nil, false)
		trait.Name = name
		grantType := feature.GrantSkill
		if _, ok := item.(*gurps.Spell); ok {
			grantType = feature.GrantSpell
		}
		g := gurps.NewGrant(grantType)
		g.SetItem(item, gurps.LibraryFile{})
		g.Adjust, g.Amount, g.Merge = adjust, amount, merge
		trait.Features = gurps.Features{g}
		e.Traits = append(e.Traits, trait)
	}
	addSkill := func(e *gurps.Entity, name string, diff gurps.AttributeDifficulty, points fxp.Int) *gurps.Skill {
		s := gurps.NewSkill(e, nil, false)
		s.Name = name
		s.Difficulty = diff
		s.Points = points
		e.Skills = append(e.Skills, s)
		return s
	}
	addTechnique := func(e *gurps.Entity, name, skill string, modifier fxp.Int, limit *fxp.Int) *gurps.Skill {
		s := gurps.NewTechnique(e, nil, skill)
		s.Name = name
		s.TechniqueDefault.Modifier = modifier
		s.TechniqueLimitModifier = limit
		e.Skills = append(e.Skills, s)
		return s
	}
	easyDX := gurps.AttributeDifficulty{Attribute: gurps.DexterityID, Difficulty: difficulty.Easy}
	hardIQ := gurps.AttributeDifficulty{Attribute: gurps.IntelligenceID, Difficulty: difficulty.Hard}
	wildDX := gurps.AttributeDifficulty{Attribute: gurps.DexterityID, Difficulty: difficulty.Wildcard}
	// skillItem returns a skill, not on the sheet, for a grant to give.
	skillItem := func(name string, diff gurps.AttributeDifficulty) *gurps.Skill {
		s := newTestSkill(name, 0, nil)
		s.Difficulty = diff
		return s
	}
	for _, tc := range []struct {
		name string
		// build sets up the sheet and returns the skill or spell whose editor is checked, or nil for the row a grant
		// put on the sheet.
		build func(e *gurps.Entity) any
		// typed, when not nil, is the points typed into the editor, diff, when not nil, the difficulty chosen there,
		// and optSpec, for a skill, the optional specialization typed there, which the sheet's level is then read with.
		typed   *fxp.Int
		diff    *difficulty.Level
		optSpec string
		want    string
	}{
		{name: "granted skill", want: "11/DX+1", build: func(e *gurps.Entity) any {
			grantOn(e, "Training", newTestSkill("Brawling", fxp.One, nil), grantadjust.AddLevels, fxp.Three,
				grantmerge.Ignore)
			return nil
		}},
		{name: "granted spell", want: "10/IQ", build: func(e *gurps.Entity) any {
			grantOn(e, "Training", newTestSpell("Light", fxp.One, nil), grantadjust.AddLevels, fxp.Three,
				grantmerge.Ignore)
			return nil
		}},
		{
			// The 2 points that buy 11 are not added to the 4 the default to Karate-2 at 12 stands for.
			name: "skill kept below its default", want: "11/DX+1", build: func(e *gurps.Entity) any {
				addSkill(e, "Karate", easyDX, fxp.Twelve)
				brawling := addSkill(e, "Brawling", easyDX, 0)
				brawling.Defaults = []*gurps.SkillDefault{{
					DefaultType: gurps.SkillID,
					Name:        criteria.Text{Compare: criteria.IsText, Qualifier: "Karate"},
					Modifier:    -fxp.Two,
				}}
				grantOn(e, "Training", brawling, grantadjust.SetLevel, fxp.Eleven, grantmerge.KeepLowest)
				return brawling
			},
		},
		{
			// The points taken away leave the skill no level of its own, so it reads as the level kept.
			name: "skill kept at a level it lacks", want: "12/IQ+2", build: func(e *gurps.Entity) any {
				alchemy := addSkill(e, "Alchemy", hardIQ, fxp.One)
				grantOn(e, "Training", alchemy, grantadjust.SetLevel, fxp.Twelve, grantmerge.KeepHighest)
				grantOn(e, "Curse", alchemy, grantadjust.AddPoints, -fxp.One, grantmerge.Ignore)
				return alchemy
			},
		},
		{name: "spell kept at a level it lacks", want: "12/IQ+2", build: func(e *gurps.Entity) any {
			light := gurps.NewSpell(e, nil, false)
			light.Name = "Light"
			light.Points = fxp.One
			e.Spells = append(e.Spells, light)
			grantOn(e, "Training", light, grantadjust.SetLevel, fxp.Twelve, grantmerge.KeepHighest)
			grantOn(e, "Curse", light, grantadjust.AddPoints, -fxp.One, grantmerge.Ignore)
			return light
		}},
		{
			// With a level of its own, the skill takes the levels added rather than the 2 points that buy them.
			name: "points typed into a skill a grant adds levels to", typed: new(fxp.One), want: "12/DX+2",
			build: func(e *gurps.Entity) any {
				brawling := addSkill(e, "Brawling", easyDX, 0)
				grantOn(e, "Training", brawling, grantadjust.AddLevels, fxp.Two, grantmerge.Ignore)
				return brawling
			},
		},
		{
			// With no level of its own, the skill takes the 2 points that buy the levels added rather than nothing.
			name: "points taken from a skill a grant adds levels to", typed: new(fxp.Int(0)), want: "11/DX+1",
			build: func(e *gurps.Entity) any {
				brawling := addSkill(e, "Brawling", easyDX, fxp.One)
				grantOn(e, "Training", brawling, grantadjust.AddLevels, fxp.Two, grantmerge.Ignore)
				return brawling
			},
		},
		{
			name: "points typed into a spell a grant adds levels to", typed: new(fxp.One), want: "10/IQ",
			build: func(e *gurps.Entity) any {
				light := gurps.NewSpell(e, nil, false)
				light.Name = "Light"
				light.Points = 0
				e.Spells = append(e.Spells, light)
				grantOn(e, "Training", light, grantadjust.AddLevels, fxp.Two, grantmerge.Ignore)
				return light
			},
		},
		{
			// The optional specialization makes the skill IQ/A, at which the 8 points that buy the level kept are
			// priced, rather than the 12 that buy it at IQ/H and IQ+3 at IQ/A.
			name: "optional specialization typed into a skill a grant keeps at a level", optSpec: "Potions",
			want: "12/IQ+2", build: func(e *gurps.Entity) any {
				alchemy := addSkill(e, "Alchemy", hardIQ, fxp.One)
				grantOn(e, "Training", alchemy, grantadjust.SetLevel, fxp.Twelve, grantmerge.KeepHighest)
				return alchemy
			},
		},
		{
			// Another difficulty makes the skill another item than the grant's, which then puts its own on the sheet
			// rather than add levels to this one.
			name: "difficulty chosen for a skill a grant adds levels to", diff: new(difficulty.Average), want: "9/DX-1",
			build: func(e *gurps.Entity) any {
				brawling := addSkill(e, "Brawling", easyDX, fxp.One)
				grantOn(e, "Training", brawling, grantadjust.AddLevels, fxp.Two, grantmerge.Ignore)
				return brawling
			},
		},
		{
			name: "difficulty chosen for a spell a grant adds levels to", diff: new(difficulty.VeryHard),
			want: "7/IQ-3", build: func(e *gurps.Entity) any {
				light := gurps.NewSpell(e, nil, false)
				light.Name = "Light"
				light.Points = fxp.One
				e.Spells = append(e.Spells, light)
				grantOn(e, "Training", light, grantadjust.AddLevels, fxp.Two, grantmerge.Ignore)
				return light
			},
		},
		{
			// Another difficulty makes the skill the item a grant put on the sheet, which then merges into it rather
			// than keep its own row, giving the 6 points that buy the levels added to one with no level of its own.
			name: "difficulty chosen for a skill a grant puts on the sheet", diff: new(difficulty.Wildcard),
			want: "8/DX-2", build: func(e *gurps.Entity) any {
				brawling := addSkill(e, "Brawling", easyDX, fxp.One)
				grantOn(e, "Training", skillItem("Brawling", wildDX), grantadjust.AddLevels, fxp.Two, grantmerge.Ignore)
				return brawling
			},
		},
		{
			// The grant, which can't put its item on the sheet, takes a level from the skill once it is that item.
			name: "difficulty chosen for a skill a grant takes levels from", diff: new(difficulty.Average),
			want: "8/DX-2", build: func(e *gurps.Entity) any {
				brawling := addSkill(e, "Brawling", easyDX, fxp.One)
				grantOn(e, "Curse", skillItem("Brawling", gurps.AttributeDifficulty{
					Attribute:  gurps.DexterityID,
					Difficulty: difficulty.Average,
				}), grantadjust.AddLevels, -fxp.One, grantmerge.Ignore)
				return brawling
			},
		},
		{
			// The grant keeps the row it put on the sheet, which the user disabled, rather than merge into the skill.
			name: "difficulty chosen for a skill whose grant's own row is disabled", diff: new(difficulty.Wildcard),
			want: "7/DX-3", build: func(e *gurps.Entity) any {
				brawling := addSkill(e, "Brawling", easyDX, fxp.Three)
				grantOn(e, "Training", skillItem("Brawling", wildDX), grantadjust.AddLevels, fxp.Two, grantmerge.Ignore)
				e.Recalculate()
				for _, one := range e.Skills {
					if gurps.IsGranted(one) {
						one.SetDisabled(true)
					}
				}
				return brawling
			},
		},
		{
			// The grant stays merged into the row of the item listed ahead of the skill.
			name: "difficulty chosen for a skill listed after a row a grant adds levels to",
			diff: new(difficulty.Wildcard), want: "7/DX-3", build: func(e *gurps.Entity) any {
				addSkill(e, "Brawling", wildDX, fxp.Three)
				brawling := addSkill(e, "Brawling", easyDX, fxp.Three)
				grantOn(e, "Training", skillItem("Brawling", wildDX), grantadjust.AddLevels, fxp.Two, grantmerge.Ignore)
				return brawling
			},
		},
		{
			// The grant moves to the skill from the row of the item listed after it.
			name: "difficulty chosen for a skill listed ahead of a row a grant adds levels to",
			diff: new(difficulty.Wildcard), want: "9/DX-1", build: func(e *gurps.Entity) any {
				brawling := addSkill(e, "Brawling", easyDX, fxp.Three)
				addSkill(e, "Brawling", wildDX, fxp.Three)
				grantOn(e, "Training", skillItem("Brawling", wildDX), grantadjust.AddLevels, fxp.Two, grantmerge.Ignore)
				return brawling
			},
		},
		{
			// The levels added count from the 11 the technique's 1 point buys over its default of Karate-2 at 10.
			name: "technique a grant adds levels to", want: "14/+2", build: func(e *gurps.Entity) any {
				addSkill(e, "Karate", easyDX, fxp.Four)
				kick := addTechnique(e, "Kick", "Karate", -fxp.Two, nil)
				grantOn(e, "Training", kick, grantadjust.AddLevels, fxp.Three, grantmerge.Ignore)
				return kick
			},
		},
		{
			// The level kept is held to the technique's limit of Karate+1 at 13.
			name: "technique a grant keeps at a level above its limit", want: "13/+1", build: func(e *gurps.Entity) any {
				addSkill(e, "Karate", easyDX, fxp.Four)
				kick := addTechnique(e, "Kick", "Karate", -fxp.Two, new(fxp.One))
				grantOn(e, "Training", kick, grantadjust.SetLevel, fxp.Fifteen, grantmerge.KeepHighest)
				return kick
			},
		},
		{
			// The optional specialization, which a technique has only from a file, makes the Hard technique Average,
			// so its 3 points buy +3 over its default of Karate-2 at 10 rather than +2.
			name: "hard technique with an optional specialization a grant adds levels to", want: "14/+2",
			build: func(e *gurps.Entity) any {
				addSkill(e, "Karate", easyDX, fxp.Four)
				kick := addTechnique(e, "Kick", "Karate", -fxp.Two, nil)
				kick.Difficulty.Difficulty = difficulty.Hard
				kick.Points = fxp.Three
				kick.OptionalSpecialization = "Spinning"
				grantOn(e, "Training", kick, grantadjust.AddLevels, fxp.One, grantmerge.Ignore)
				return kick
			},
		},
		{
			name: "difficulty chosen for a spell a grant puts on the sheet", diff: new(difficulty.VeryHard),
			want: "9/IQ-1", build: func(e *gurps.Entity) any {
				light := gurps.NewSpell(e, nil, false)
				light.Name = "Light"
				light.Points = fxp.One
				e.Spells = append(e.Spells, light)
				granted := newTestSpell("Light", 0, nil)
				granted.Difficulty.Difficulty = difficulty.VeryHard
				grantOn(e, "Training", granted, grantadjust.AddLevels, fxp.Two, grantmerge.Ignore)
				return light
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			sheet := newTestSheetForTemplate(t)
			entity := sheet.Entity()
			target := tc.build(entity)
			entity.Recalculate()
			if target == nil {
				for _, one := range entity.Skills {
					if gurps.IsGranted(one) {
						target = one
					}
				}
				for _, one := range entity.Spells {
					if gurps.IsGranted(one) {
						target = one
					}
				}
				if target == nil {
					t.Fatal("precondition: the grant puts its item on the sheet")
				}
			}
			var preview string
			var onSheet func() string
			switch item := target.(type) {
			case *gurps.Skill:
				onSheet = func() string {
					return item.LevelData.Level.Floor().String() + "/" +
						gurps.FormatRelativeSkill(entity, item.IsTechnique(), item.Difficulty, item.AdjustedRelativeLevel())
				}
				e, content := buildEditorContent(sheet, item, initSkillEditor)
				if tc.typed != nil {
					e.editorData.Points = *tc.typed
				}
				if tc.diff != nil {
					e.editorData.Difficulty.Difficulty = *tc.diff
				}
				if tc.optSpec != "" {
					e.editorData.OptionalSpecialization = tc.optSpec
				}
				DeepSync(e)
				if tc.typed != nil {
					item.Points = *tc.typed
				}
				if tc.diff != nil {
					item.Difficulty.Difficulty = *tc.diff
				}
				if tc.optSpec != "" {
					item.OptionalSpecialization = tc.optSpec
				}
				preview = levelPreview(content)
			case *gurps.Spell:
				onSheet = func() string {
					return item.LevelData.Level.Floor().String() + "/" +
						gurps.FormatRelativeSkill(entity, false, item.Difficulty, item.LevelData.RelativeLevel)
				}
				e, content := buildEditorContent(sheet, item, initSpellEditor)
				if tc.typed != nil {
					e.editorData.Points = *tc.typed
				}
				if tc.diff != nil {
					e.editorData.Difficulty.Difficulty = *tc.diff
				}
				DeepSync(e)
				if tc.typed != nil {
					item.Points = *tc.typed
				}
				if tc.diff != nil {
					item.Difficulty.Difficulty = *tc.diff
				}
				preview = levelPreview(content)
			}
			entity.Recalculate()
			c.Equal(tc.want, onSheet(), "precondition: the sheet's level")
			c.Equal(tc.want, preview, "the editor previews the level the sheet shows")
		})
	}
}
