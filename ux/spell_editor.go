// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// EditSpell displays the editor for an spell.
func EditSpell(owner Rebuildable, spell *gurps.Spell) {
	displayEditor(owner, spell, svg.GCSSpells, "md:Help/Interface/Spell", nil,
		initSpellEditor, nil)
}

func initSpellEditor(e *editor[*gurps.Spell, *gurps.SpellEditData], content *unison.Panel) func() {
	owner := e.owner.AsPanel().Self
	_, ownerIsSheet := owner.(*Sheet)
	_, ownerIsTemplate := owner.(*Template)
	addNameLabelAndField(content, &e.editorData.Name)
	entity := gurps.EntityFromNode(e.target)
	if !e.target.Container() {
		addTechLevelRequired(content, &e.editorData.TechLevel, ownerIsSheet)
		addLabelAndListField(content, i18n.Text("College"), i18n.Text("colleges"), (*[]string)(&e.editorData.College))
		addLabelAndStringField(content, i18n.Text("Class"), "", &e.editorData.Class)
		addLabelAndStringField(content, i18n.Text("Power Source"), "", &e.editorData.PowerSource)
		if e.target.IsRitualMagic() {
			addLabelAndStringField(content, i18n.Text("Base Skill"), "", &e.editorData.RitualSkillName)
			wrapper := addFlowWrapper(content, i18n.Text("Difficulty"), 3)
			addPopup(wrapper, difficulty.TechniqueLevels, &e.editorData.Difficulty.Difficulty)
			prereqCount := i18n.Text("Prerequisite Count")
			wrapper.AddChild(NewFieldInteriorLeadingLabel(prereqCount, false))
			addIntegerField(wrapper, nil, "", prereqCount, "", &e.editorData.RitualPrereqCount, 0, 99)
		} else {
			addDifficultyLabelAndFields(content, entity, &e.editorData.Difficulty)
		}
		if ownerIsSheet || ownerIsTemplate {
			pointsLabel := i18n.Text("Points")
			wrapper := addFlowWrapper(content, pointsLabel, 3)
			addDecimalField(wrapper, nil, "", pointsLabel, "", &e.editorData.Points, 0, fxp.MaxBasePoints)
			wrapper.AddChild(NewFieldInteriorLeadingLabel(i18n.Text("Level"), false))
			levelField := NewNonEditableField(func(field *NonEditableField) {
				replacements := e.target.NameableReplacements()
				localName := nameable.Apply(e.editorData.Name, replacements)
				localPowerSource := nameable.Apply(e.editorData.PowerSource, replacements)
				localColleges := nameable.ApplyToList(e.editorData.College, replacements)
				points := gurps.AdjustedPointsForNonContainerSpell(entity, e.editorData.Points, localName,
					localPowerSource, localColleges, e.editorData.Tags, nil)
				var level gurps.Level
				if e.target.IsRitualMagic() {
					level = gurps.CalculateRitualMagicSpellLevel(entity, localName, localPowerSource,
						nameable.Apply(e.editorData.RitualSkillName, replacements),
						e.editorData.RitualPrereqCount, localColleges, e.editorData.Tags, e.editorData.Difficulty,
						points)
				} else {
					level = gurps.CalculateSpellLevel(entity, localName, localPowerSource, localColleges,
						e.editorData.Tags, e.editorData.Difficulty, points)
				}
				lvl := level.Level.Trunc()
				if lvl <= 0 {
					field.SetTitle("-")
				} else {
					field.SetTitle(lvl.String() + "/" + gurps.FormatRelativeSkill(entity,
						e.target.IsRitualMagic(), e.editorData.Difficulty, level.RelativeLevel))
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
	addPageRefHighlightLabelAndField(content, &e.editorData.PageRefHighlight)
	addSourceFields(content, &e.target.SourcedID)
	if !e.target.Container() {
		content.AddChild(newPrereqPanel(entity, &e.editorData.Prereq, prereq.TypesForNonEquipment))
		content.AddChild(newWeaponsPanel(e, e.target, true, &e.editorData.Weapons))
		content.AddChild(newWeaponsPanel(e, e.target, false, &e.editorData.Weapons))
		content.AddChild(newStudyPanel(entity, &e.editorData.StudyHoursNeeded, &e.editorData.Study))
	}
	return nil
}
