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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

// EditSpell displays the editor for a spell.
func EditSpell(owner Rebuildable, spell *gurps.Spell) *editor[*gurps.Spell, *gurps.SpellEditData] {
	return displayEditor(owner, spell, svg.GCSSpells, "md:User%20Guide/Spells", nil, initSpellEditor, nil)
}

func initSpellEditor(e *editor[*gurps.Spell, *gurps.SpellEditData], content *unison.Panel) func() {
	owner := e.owner.AsPanel().Self
	_, ownerIsSheet := owner.(*Sheet)
	_, ownerIsTemplate := owner.(*Template)
	entity := gurps.EntityFromNode(e.target)
	addNameLabelAndField(content, &e.editorData.Name)
	addNotesLabelAndField(content, &e.editorData.LocalNotes)
	// A choice container dissolves into the options chosen from it when the template is applied, so it has no use for
	// the VTT notes or tags, which only matter to a spell that reaches a sheet.
	choice := gurps.IsTemplateChoiceContainer(e.target)
	if !choice {
		addVTTNotesLabelAndField(content, &e.editorData.VTTNotes)
	}
	if !e.target.Container() {
		addTechLevelRequired(content, &e.editorData.TechLevel, ownerIsSheet)
		addLabelAndListField(content, i18n.Text("College"), i18n.Text("colleges"), (*[]string)(&e.editorData.College))
		addLabelAndStringField(content, i18n.Text("Class"), "", &e.editorData.Class)
		addLabelAndStringField(content, i18n.Text("Power Source"), "", &e.editorData.PowerSource)
		prereqCount := i18n.Text("Prerequisite Count")
		if e.target.IsRitualMagic() {
			addLabelAndStringField(content, i18n.Text("Base Skill"), "", &e.editorData.RitualSkillName)
			wrapper, diffLabel := addFlowWrapper(content, i18n.Text("Difficulty"), 3)
			labelControl(addPopup(wrapper, difficulty.TechniqueLevels, &e.editorData.Difficulty.Difficulty), diffLabel)
			wrapper.AddChild(NewFieldInteriorLeadingLabel(prereqCount, false))
			addIntegerField(wrapper, nil, "", prereqCount, "", &e.editorData.PrereqCount, 0, 999)
		} else {
			addDifficultyLabelAndFields(content, entity, &e.editorData.Difficulty)
			addLabelAndIntegerField(content, nil, "", prereqCount, "", &e.editorData.PrereqCount, 0, 999)
		}
		if ownerIsSheet || ownerIsTemplate {
			pointsLabel := i18n.Text("Points")
			wrapper, _ := addFlowWrapper(content, pointsLabel, 3)
			addDecimalField(wrapper, nil, "", pointsLabel, "", &e.editorData.Points, 0, fxp.MaxBasePoints, false)
			wrapper.AddChild(NewFieldInteriorLeadingLabel(i18n.Text("Level"), false))
			levelField := NewNonEditableField(func(field *NonEditableField) {
				replacements := e.target.NameableReplacements()
				localName := nameable.Apply(e.editorData.Name, replacements)
				localPowerSource := nameable.Apply(e.editorData.PowerSource, replacements)
				localColleges := nameable.ApplyToList(e.editorData.College, replacements)
				// What grants give the spell as edited counts, as it does on the sheet once the edits are applied.
				points := e.target.PointsWithGrants(e.editorData, gurps.AdjustedPointsForNonContainerSpell(entity,
					e.editorData.Points, localName, localPowerSource, localColleges, e.editorData.Tags, nil))
				levelAt := func(points fxp.Int) gurps.Level {
					if e.target.IsRitualMagic() {
						return gurps.CalculateRitualMagicSpellLevel(entity, localName, localPowerSource,
							nameable.Apply(e.editorData.RitualSkillName, replacements),
							e.editorData.PrereqCount, localColleges, e.editorData.Tags, e.editorData.Difficulty,
							points)
					}
					return gurps.CalculateSpellLevel(entity, localName, localPowerSource, localColleges,
						e.editorData.Tags, e.editorData.Difficulty, points)
				}
				level := e.target.LevelWithGrants(e.editorData, levelAt(points), func(tierPoints fxp.Int) gurps.Level {
					return levelAt(gurps.AdjustedPointsForNonContainerSpell(entity, tierPoints, localName,
						localPowerSource, localColleges, e.editorData.Tags, nil))
				})
				lvl := level.Level.Floor()
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
				MinSize: geom.NewSize(levelField.Font.SimpleWidth((-fxp.MaxBasePoints*2).String())+insets.Left+insets.Right, 0),
			})
			wrapper.AddChild(levelField)
		}
		addLabelAndStringField(content, i18n.Text("Resistance"), "", &e.editorData.Resist)
		addLabelAndMultiLineStringField(content, i18n.Text("Casting Cost"), "", &e.editorData.CastingCost)
		addLabelAndMultiLineStringField(content, i18n.Text("Maintenance Cost"), "", &e.editorData.MaintenanceCost)
		addLabelAndMultiLineStringField(content, i18n.Text("Casting Time"), "", &e.editorData.CastingTime)
		addLabelAndMultiLineStringField(content, i18n.Text("Duration"), "", &e.editorData.Duration)
		addLabelAndMultiLineStringField(content, i18n.Text("Item"), "", &e.editorData.Item)
		addRitualOverrideFields(content, &e.editorData.RitualOverrides)
	}
	if !choice {
		addTagsLabelAndField(content, &e.editorData.Tags)
	}
	addPreconfigurable(e, content)
	addPickSeparately(content, e.target, &e.editorData.PickSeparately)
	addChoices(e, content)
	if !e.target.Container() {
		addSwitchedOnCheckBox(content, &e.editorData.SwitchedOn)
	}
	addPageRefLabelsAndFields(content, &e.editorData.PageRef, &e.editorData.PageRefHighlight, e.pageRefHighlight)
	if !e.target.Container() {
		d := e.editorData
		content.AddChild(newNameablesPanel(&d.Nameables, func() map[string]int {
			return nameableUses(append([]string{
				d.Name, d.LocalNotes, d.PowerSource, d.Class, d.Resist, d.CastingCost, d.MaintenanceCost, d.CastingTime,
				d.Duration, d.Item, d.RitualOverrides.Ritual, d.RitualOverrides.Time, d.RitualOverrides.Cost,
				d.RitualSkillName,
			}, d.College...), d.Prereq, d.Features, d.Weapons)
		}, func() map[string]nameable.Detail { return grantedNameableDefinitions(d.Features) }, nil, nil, nil))
		content.AddChild(newPrereqPanel(entity, &e.editorData.Prereq, prereq.TypesForNonEquipment, true).
			withReplacementsFrom(e.editorData))
		content.AddChild(newFeaturesPanel(entity, e.target, &e.editorData.Features, false).
			withReplacementsFrom(e.editorData))
		e.meleeWeapons = newWeaponsPanel(e, e.target, true, &e.editorData.Weapons)
		content.AddChild(e.meleeWeapons)
		e.rangedWeapons = newWeaponsPanel(e, e.target, false, &e.editorData.Weapons)
		content.AddChild(e.rangedWeapons)
		if entity != nil {
			content.AddChild(newStudyPanel(&e.editorData.StudyHoursNeeded, &e.editorData.Study))
		}
	}
	return nil
}

// addRitualOverrideFields adds a row with a field for each part of the rituals text a spell's description can fix in
// place of what its level gives.
func addRitualOverrideFields(parent *unison.Panel, overrides *gurps.RitualOverrides) {
	tip := i18n.Text("Leave blank to use the standard text for the spell's level, enter \"-\" to leave this part out, or enter the text to show in its place at every level")
	wrapper, label := addFillWrapper(parent, i18n.Text("Rituals"), 6)
	label.Tooltip = newWrappedTooltip(tip)
	ritual := i18n.Text("Ritual")
	wrapper.AddChild(NewFieldLeadingLabel(ritual, false))
	addStringField(wrapper, ritual, tip, &overrides.Ritual)
	// The time and cost fields follow "Time" and "Cost" labels, which would leave a screen reader and the Undo menu
	// unable to tell them from the casting time and cost, so they are named in full.
	wrapper.AddChild(NewFieldInteriorLeadingLabel(i18n.Text("Time"), false))
	text := i18n.Text("Ritual Time")
	addStringField(wrapper, text, tip, &overrides.Time).Accessibility.Name = text
	wrapper.AddChild(NewFieldInteriorLeadingLabel(i18n.Text("Cost"), false))
	text = i18n.Text("Ritual Cost")
	addStringField(wrapper, text, tip, &overrides.Cost).Accessibility.Name = text
}
