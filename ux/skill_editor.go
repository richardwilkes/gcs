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
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
)

// EditSkill displays the editor for an skill.
func EditSkill(owner Rebuildable, skill *gurps.Skill) {
	displayEditor(owner, skill, svg.GCSSkills, "md:Help/Interface/Skill", nil, initSkillEditor, nil)
}

func initSkillEditor(e *editor[*gurps.Skill, *gurps.SkillEditData], content *unison.Panel) func() {
	owner := e.owner.AsPanel().Self
	_, ownerIsSheet := owner.(*Sheet)
	_, ownerIsTemplate := owner.(*Template)
	addNameLabelAndField(content, &e.editorData.Name)
	if !e.target.Container() && !e.target.IsTechnique() {
		addSpecializationLabelAndField(content, &e.editorData.Specialization)
		addTechLevelRequired(content, &e.editorData.TechLevel, ownerIsSheet)
	}
	addNotesLabelAndField(content, &e.editorData.LocalNotes)
	addVTTNotesLabelAndField(content, &e.editorData.VTTNotes)
	addTagsLabelAndField(content, &e.editorData.Tags)
	entity := gurps.EntityFromNode(e.target)
	if e.target.Container() {
		addTemplateChoices(content, nil, "", &e.editorData.TemplatePicker)
	} else {
		if e.target.IsTechnique() {
			wrapper := addFlowWrapper(content, i18n.Text("Defaults To"), 4)
			wrapper.SetLayoutData(&unison.FlexLayoutData{
				HAlign: align.Fill,
				HGrab:  true,
			})
			choices, attrChoice := gurps.AttributeChoices(entity, "",
				gurps.TenFlag|gurps.SkillFlag|gurps.ParryFlag|gurps.BlockFlag|gurps.DodgeFlag,
				e.editorData.TechniqueDefault.DefaultType)
			attrChoicePopup := addPopup(wrapper, choices, &attrChoice)
			skillDefNameField := addStringField(wrapper, i18n.Text("Technique Default Skill Name"),
				i18n.Text("Skill Name"), &e.editorData.TechniqueDefault.Name)
			skillDefNameField.Watermark = i18n.Text("Skill")
			skillDefNameField.SetLayoutData(&unison.FlexLayoutData{
				HAlign: align.Fill,
				HGrab:  true,
			})
			skillDefSpecialtyField := addStringField(wrapper, i18n.Text("Technique Default Skill Specialization"),
				i18n.Text("Skill Specialization"), &e.editorData.TechniqueDefault.Specialization)
			skillDefSpecialtyField.Watermark = i18n.Text("Specialization")
			skillDefSpecialtyField.SetLayoutData(&unison.FlexLayoutData{
				HAlign: align.Fill,
				HGrab:  true,
			})
			lastWasSkillBased := gurps.DefaultTypeIsSkillBased(e.editorData.TechniqueDefault.DefaultType)
			if !lastWasSkillBased {
				skillDefNameField.RemoveFromParent()
				skillDefSpecialtyField.RemoveFromParent()
			}
			addDecimalField(wrapper, nil, "", i18n.Text("Technique Default Adjustment"),
				i18n.Text("Default Adjustment"), &e.editorData.TechniqueDefault.Modifier, -fxp.NinetyNine,
				fxp.NinetyNine)
			attrChoicePopup.SelectionChangedCallback = func(popup *unison.PopupMenu[*gurps.AttributeChoice]) {
				if item, ok := popup.Selected(); ok {
					e.editorData.TechniqueDefault.DefaultType = item.Key
					if skillBased := gurps.DefaultTypeIsSkillBased(e.editorData.TechniqueDefault.DefaultType); skillBased != lastWasSkillBased {
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
				func() check.Enum {
					return check.FromBool(e.editorData.TechniqueLimitModifier != nil)
				},
				func(state check.Enum) {
					if state == check.On {
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
			difficultyPopup := addLabelAndPopup(content, i18n.Text("Difficulty"), "", difficulty.TechniqueLevels,
				&e.editorData.Difficulty.Difficulty)
			difficultyPopup.SelectionChangedCallback = func(popup *unison.PopupMenu[difficulty.Level]) {
				if item, ok := popup.Selected(); ok {
					e.editorData.Difficulty.Difficulty = item
					if !ownerIsSheet && !ownerIsTemplate {
						if item == difficulty.Hard {
							e.editorData.Points = fxp.Two
						} else {
							e.editorData.Points = fxp.One
						}
					}
					MarkModified(difficultyPopup)
				}
			}
		} else {
			addDifficultyLabelAndFields(content, entity, &e.editorData.Difficulty)
			encLabel := i18n.Text("Encumbrance Penalty")
			wrapper := addFlowWrapper(content, encLabel, 2)
			addDecimalField(wrapper, nil, "", encLabel, "", &e.editorData.EncumbrancePenaltyMultiplier, 0, fxp.Nine)
			wrapper.AddChild(NewFieldTrailingLabel(i18n.Text("times the current encumbrance level"), false))
		}

		if ownerIsSheet || ownerIsTemplate {
			pointsLabel := i18n.Text("Points")
			wrapper := addFlowWrapper(content, pointsLabel, 3)
			addDecimalField(wrapper, nil, "", pointsLabel, "", &e.editorData.Points, 0, fxp.MaxBasePoints)
			wrapper.AddChild(NewFieldInteriorLeadingLabel(i18n.Text("Level"), false))
			levelField := NewNonEditableField(func(field *NonEditableField) {
				localName := nameable.Apply(e.editorData.Name, e.target.Replacements)
				localSpec := nameable.Apply(e.editorData.Specialization, e.target.Replacements)
				points := gurps.AdjustedPointsForNonContainerSkillOrTechnique(entity, e.editorData.Points, localName,
					localSpec, e.editorData.Tags, nil)
				var level gurps.Level
				if e.target.IsTechnique() {
					level = gurps.CalculateTechniqueLevel(entity, e.target.NameableReplacements(), localName, localSpec,
						e.editorData.Tags, e.editorData.TechniqueDefault, e.editorData.Difficulty.Difficulty, points,
						true, e.editorData.TechniqueLimitModifier, nil)
				} else {
					level = gurps.CalculateSkillLevel(entity, localName, localSpec, e.editorData.Tags,
						e.editorData.DefaultedFrom, e.editorData.Difficulty, points,
						e.editorData.EncumbrancePenaltyMultiplier)
				}
				lvl := level.Level.Trunc()
				if lvl <= 0 {
					field.SetTitle("-")
				} else {
					rsl := level.RelativeLevel
					if e.target.IsTechnique() {
						rsl += e.editorData.TechniqueDefault.Modifier
					}
					field.SetTitle(lvl.String() + "/" + gurps.FormatRelativeSkill(entity,
						e.target.IsTechnique(), e.editorData.Difficulty, rsl))
				}
				field.MarkForLayoutAndRedraw()
			})
			insets := levelField.Border().Insets()
			levelField.SetLayoutData(&unison.FlexLayoutData{
				MinSize: unison.NewSize(levelField.Font.SimpleWidth((-fxp.MaxBasePoints*2).String())+insets.Left+
					insets.Right, 0),
			})
			wrapper.AddChild(levelField)
		}
	}
	addPageRefLabelAndField(content, &e.editorData.PageRef)
	addPageRefHighlightLabelAndField(content, &e.editorData.PageRefHighlight)
	addSourceFields(content, &e.target.SourcedID)
	if !e.target.Container() {
		content.AddChild(newPrereqPanel(entity, &e.editorData.Prereq, prereq.TypesForNonEquipment))
		if !e.target.IsTechnique() {
			content.AddChild(newDefaultsPanel(entity, &e.editorData.Defaults))
		}
		content.AddChild(newFeaturesPanel(entity, e.target, &e.editorData.Features, false))
		content.AddChild(newWeaponsPanel(e, e.target, true, &e.editorData.Weapons))
		content.AddChild(newWeaponsPanel(e, e.target, false, &e.editorData.Weapons))
		content.AddChild(newStudyPanel(entity, &e.editorData.StudyHoursNeeded, &e.editorData.Study))
	}
	return nil
}
