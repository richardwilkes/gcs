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
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// EditSkill displays the editor for an skill.
func EditSkill(owner Rebuildable, skill *model.Skill) {
	displayEditor[*model.Skill, *model.SkillEditData](owner, skill, svg.GCSSkills, nil, initSkillEditor)
}

func initSkillEditor(e *editor[*model.Skill, *model.SkillEditData], content *unison.Panel) func() {
	owner := e.owner.AsPanel().Self
	_, ownerIsSheet := owner.(*Sheet)
	_, ownerIsTemplate := owner.(*Template)
	addNameLabelAndField(content, &e.editorData.Name)
	isTechnique := strings.HasPrefix(e.target.Type, model.TechniqueID)
	if !e.target.Container() && !isTechnique {
		addSpecializationLabelAndField(content, &e.editorData.Specialization)
		addTechLevelRequired(content, &e.editorData.TechLevel, ownerIsSheet)
	}
	addNotesLabelAndField(content, &e.editorData.LocalNotes)
	addVTTNotesLabelAndField(content, &e.editorData.VTTNotes)
	addTagsLabelAndField(content, &e.editorData.Tags)
	if e.target.Container() {
		addTemplateChoices(content, nil, "", &e.editorData.TemplatePicker)
	} else {
		if isTechnique {
			wrapper := addFlowWrapper(content, i18n.Text("Defaults To"), 4)
			wrapper.SetLayoutData(&unison.FlexLayoutData{
				HAlign: unison.FillAlignment,
				HGrab:  true,
			})
			flags := model.TenFlag
			if isTechnique {
				flags |= model.SkillFlag + model.ParryFlag + model.BlockFlag + model.DodgeFlag
			}
			choices, attrChoice := model.AttributeChoices(e.target.Entity, "", flags, e.editorData.TechniqueDefault.DefaultType)
			attrChoicePopup := addPopup(wrapper, choices, &attrChoice)
			skillDefNameField := addStringField(wrapper, i18n.Text("Technique Default Skill Name"),
				i18n.Text("Skill Name"), &e.editorData.TechniqueDefault.Name)
			skillDefNameField.Watermark = i18n.Text("Skill")
			skillDefNameField.SetLayoutData(&unison.FlexLayoutData{
				HAlign: unison.FillAlignment,
				HGrab:  true,
			})
			skillDefSpecialtyField := addStringField(wrapper, i18n.Text("Technique Default Skill Specialization"),
				i18n.Text("Skill Specialization"), &e.editorData.TechniqueDefault.Specialization)
			skillDefSpecialtyField.Watermark = i18n.Text("Specialization")
			skillDefSpecialtyField.SetLayoutData(&unison.FlexLayoutData{
				HAlign: unison.FillAlignment,
				HGrab:  true,
			})
			lastWasSkillBased := model.DefaultTypeIsSkillBased(e.editorData.TechniqueDefault.DefaultType)
			if !lastWasSkillBased {
				skillDefNameField.RemoveFromParent()
				skillDefSpecialtyField.RemoveFromParent()
			}
			addDecimalField(wrapper, nil, "", i18n.Text("Technique Default Adjustment"),
				i18n.Text("Default Adjustment"), &e.editorData.TechniqueDefault.Modifier, -fxp.NinetyNine,
				fxp.NinetyNine)
			attrChoicePopup.SelectionCallback = func(_ int, item *model.AttributeChoice) {
				e.editorData.TechniqueDefault.DefaultType = item.Key
				if skillBased := model.DefaultTypeIsSkillBased(e.editorData.TechniqueDefault.DefaultType); skillBased != lastWasSkillBased {
					lastWasSkillBased = skillBased
					if skillBased {
						wrapper.AddChildAtIndex(skillDefNameField, len(wrapper.Children())-1)
						wrapper.AddChildAtIndex(skillDefSpecialtyField, len(wrapper.Children())-1)
					} else {
						skillDefNameField.RemoveFromParent()
						skillDefSpecialtyField.RemoveFromParent()
					}
				}
				MarkModified(content)
			}
			wrapper2 := addFlowWrapper(content, "", 2)
			limitField := NewDecimalField(nil, "", i18n.Text("Limit"),
				func() fxp.Int {
					if e.editorData.TechniqueLimitModifier != nil {
						return *e.editorData.TechniqueLimitModifier
					}
					return 0
				}, func(value fxp.Int) {
					if e.editorData.TechniqueLimitModifier != nil {
						*e.editorData.TechniqueLimitModifier = value
					}
					MarkModified(wrapper2)
				}, -fxp.NinetyNine, fxp.NinetyNine, false, false)
			wrapper2.AddChild(NewCheckBox(nil, "", i18n.Text("Cannot exceed default skill level by more than"),
				func() unison.CheckState {
					return unison.CheckStateFromBool(e.editorData.TechniqueLimitModifier != nil)
				},
				func(state unison.CheckState) {
					if state == unison.OnCheckState {
						if e.editorData.TechniqueLimitModifier == nil {
							var limit fxp.Int
							e.editorData.TechniqueLimitModifier = &limit
						}
						adjustFieldBlank(limitField, false)
					} else {
						e.editorData.TechniqueLimitModifier = nil
						adjustFieldBlank(limitField, true)
					}
				}))
			adjustFieldBlank(limitField, e.editorData.TechniqueLimitModifier == nil)
			wrapper2.AddChild(limitField)
			difficultyPopup := addLabelAndPopup(content, i18n.Text("Difficulty"), "", model.AllTechniqueDifficulty,
				&e.editorData.Difficulty.Difficulty)
			difficultyPopup.SelectionCallback = func(_ int, item model.Difficulty) {
				e.editorData.Difficulty.Difficulty = item
				if !ownerIsSheet && !ownerIsTemplate {
					if item == model.Hard {
						e.editorData.Points = fxp.Two
					} else {
						e.editorData.Points = fxp.One
					}
				}
				MarkModified(difficultyPopup)
			}
		} else {
			addDifficultyLabelAndFields(content, e.target.Entity, &e.editorData.Difficulty)
			encLabel := i18n.Text("Encumbrance Penalty")
			wrapper := addFlowWrapper(content, encLabel, 2)
			addDecimalField(wrapper, nil, "", encLabel, "", &e.editorData.EncumbrancePenaltyMultiplier, 0, fxp.Nine)
			wrapper.AddChild(NewFieldTrailingLabel(i18n.Text("times the current encumbrance level")))
		}

		if ownerIsSheet || ownerIsTemplate {
			pointsLabel := i18n.Text("Points")
			wrapper := addFlowWrapper(content, pointsLabel, 3)
			addDecimalField(wrapper, nil, "", pointsLabel, "", &e.editorData.Points, 0, fxp.MaxBasePoints)
			wrapper.AddChild(NewFieldInteriorLeadingLabel(i18n.Text("Level")))
			levelField := NewNonEditableField(func(field *NonEditableField) {
				points := model.AdjustedPointsForNonContainerSkillOrTechnique(e.target.Entity, e.editorData.Points,
					e.editorData.Name, e.editorData.Specialization, e.editorData.Tags, nil)
				var level model.Level
				if isTechnique {
					level = model.CalculateTechniqueLevel(e.target.Entity, e.editorData.Name,
						e.editorData.Specialization, e.editorData.Tags, e.editorData.TechniqueDefault,
						e.editorData.Difficulty.Difficulty, points, true, e.editorData.TechniqueLimitModifier)
				} else {
					level = model.CalculateSkillLevel(e.target.Entity, e.editorData.Name, e.editorData.Specialization,
						e.editorData.Tags, e.editorData.DefaultedFrom, e.editorData.Difficulty, points,
						e.editorData.EncumbrancePenaltyMultiplier)
				}
				lvl := level.Level.Trunc()
				if lvl <= 0 {
					field.Text = "-"
				} else {
					rsl := level.RelativeLevel
					if isTechnique {
						rsl += e.editorData.TechniqueDefault.Modifier
					}
					field.Text = lvl.String() + "/" + model.FormatRelativeSkill(e.target.Entity, e.target.Type,
						e.editorData.Difficulty, rsl)
				}
				field.MarkForLayoutAndRedraw()
			})
			insets := levelField.Border().Insets()
			levelField.SetLayoutData(&unison.FlexLayoutData{
				MinSize: unison.NewSize(levelField.Font.SimpleWidth((-fxp.MaxBasePoints*2).String())+insets.Left+insets.Right, 0),
			})
			wrapper.AddChild(levelField)
		}
	}
	addPageRefLabelAndField(content, &e.editorData.PageRef)
	if !e.target.Container() {
		content.AddChild(newPrereqPanel(e.target.Entity, &e.editorData.Prereq))
		content.AddChild(newDefaultsPanel(e.target.Entity, &e.editorData.Defaults))
		content.AddChild(newFeaturesPanel(e.target.Entity, e.target, &e.editorData.Features))
		for _, wt := range model.AllWeaponType {
			content.AddChild(newWeaponsPanel(e, e.target, wt, &e.editorData.Weapons))
		}
		content.AddChild(newStudyPanel(e.target.Entity, &e.editorData.Study))
	}
	return nil
}
