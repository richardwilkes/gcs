/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gid"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// EditSpell displays the editor for an spell.
func EditSpell(owner Rebuildable, spell *model.Spell) {
	displayEditor[*model.Spell, *model.SpellEditData](owner, spell, svg.GCSSpells, initSpellEditor)
}

func initSpellEditor(e *editor[*model.Spell, *model.SpellEditData], content *unison.Panel) func() {
	owner := e.owner.AsPanel().Self
	_, ownerIsSheet := owner.(*Sheet)
	_, ownerIsTemplate := owner.(*Template)
	isRitualMagic := strings.HasPrefix(e.target.Type, gid.RitualMagicSpell)
	addNameLabelAndField(content, &e.editorData.Name)
	if !e.target.Container() {
		addTechLevelRequired(content, &e.editorData.TechLevel, ownerIsSheet)
		addLabelAndListField(content, i18n.Text("College"), i18n.Text("colleges"), (*[]string)(&e.editorData.College))
		addLabelAndStringField(content, i18n.Text("Class"), "", &e.editorData.Class)
		addLabelAndStringField(content, i18n.Text("Power Source"), "", &e.editorData.PowerSource)
		if isRitualMagic {
			addLabelAndStringField(content, i18n.Text("Base Skill"), "", &e.editorData.RitualSkillName)
			wrapper := addFlowWrapper(content, i18n.Text("Difficulty"), 3)
			addPopup(wrapper, model.AllTechniqueDifficulty, &e.editorData.Difficulty.Difficulty)
			prereqCount := i18n.Text("Prerequisite Count")
			wrapper.AddChild(NewFieldInteriorLeadingLabel(prereqCount))
			addIntegerField(wrapper, nil, "", prereqCount, "", &e.editorData.RitualPrereqCount, 0, 99)
		} else {
			addDifficultyLabelAndFields(content, e.target.Entity, &e.editorData.Difficulty)
		}
		if ownerIsSheet || ownerIsTemplate {
			pointsLabel := i18n.Text("Points")
			wrapper := addFlowWrapper(content, pointsLabel, 3)
			addDecimalField(wrapper, nil, "", pointsLabel, "", &e.editorData.Points, 0, fxp.MaxBasePoints)
			wrapper.AddChild(NewFieldInteriorLeadingLabel(i18n.Text("Level")))
			levelField := NewNonEditableField(func(field *NonEditableField) {
				points := model.AdjustedPointsForNonContainerSpell(e.target.Entity, e.editorData.Points,
					e.editorData.Name, e.editorData.PowerSource, e.editorData.College, e.editorData.Tags, nil)
				var level model.Level
				if isRitualMagic {
					level = model.CalculateRitualMagicSpellLevel(e.target.Entity, e.editorData.Name,
						e.editorData.PowerSource, e.editorData.RitualSkillName, e.editorData.RitualPrereqCount,
						e.editorData.College, e.editorData.Tags, e.editorData.Difficulty, points)
				} else {
					level = model.CalculateSpellLevel(e.target.Entity, e.editorData.Name, e.editorData.PowerSource,
						e.editorData.College, e.editorData.Tags, e.editorData.Difficulty, points)
				}
				lvl := level.Level.Trunc()
				if lvl <= 0 {
					field.Text = "-"
				} else {
					field.Text = lvl.String() + "/" + model.FormatRelativeSkill(e.target.Entity, e.target.Type,
						e.editorData.Difficulty, level.RelativeLevel)
				}
				field.MarkForLayoutAndRedraw()
			})
			insets := levelField.Border().Insets()
			levelField.SetLayoutData(&unison.FlexLayoutData{
				MinSize: unison.NewSize(levelField.Font.SimpleWidth((-fxp.MaxBasePoints*2).String())+insets.Left+insets.Right, 0),
			})
			wrapper.AddChild(levelField)
		}
		addLabelAndStringField(content, i18n.Text("Resistance"), "", &e.editorData.Resist)
		addLabelAndStringField(content, i18n.Text("Casting Cost"), "", &e.editorData.CastingCost)
		addLabelAndStringField(content, i18n.Text("Maintenance Cost"), "", &e.editorData.MaintenanceCost)
		addLabelAndStringField(content, i18n.Text("Casting Time"), "", &e.editorData.CastingTime)
		addLabelAndStringField(content, i18n.Text("Casting Duration"), "", &e.editorData.Duration)
	}
	addNotesLabelAndField(content, &e.editorData.LocalNotes)
	addVTTNotesLabelAndField(content, &e.editorData.VTTNotes)
	addTagsLabelAndField(content, &e.editorData.Tags)
	if e.target.Container() {
		addTemplateChoices(content, nil, "", &e.editorData.TemplatePicker)
	}
	addPageRefLabelAndField(content, &e.editorData.PageRef)
	if !e.target.Container() {
		content.AddChild(newPrereqPanel(e.target.Entity, &e.editorData.Prereq))
		for _, wt := range model.AllWeaponType {
			content.AddChild(newWeaponsPanel(e, e.target, wt, &e.editorData.Weapons))
		}
		content.AddChild(newStudyPanel(e.target.Entity, &e.editorData.Study))
	}
	return nil
}
