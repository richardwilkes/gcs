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
	"fmt"
	"reflect"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xmath"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/slices"
)

var (
	lastFeatureTypeUsed = gurps.AttributeBonusFeatureType
	lastAttributeIDUsed = gurps.StrengthID
)

type featuresPanel struct {
	unison.Panel
	entity   *gurps.Entity
	owner    fmt.Stringer
	features *gurps.Features
}

func newFeaturesPanel(entity *gurps.Entity, owner fmt.Stringer, features *gurps.Features) *featuresPanel {
	p := &featuresPanel{
		entity:   entity,
		owner:    owner,
		features: features,
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	p.SetBorder(unison.NewCompoundBorder(
		&TitledBorder{
			Title: i18n.Text("Features"),
			Font:  unison.LabelFont,
		},
		unison.NewEmptyBorder(unison.NewUniformInsets(2))))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.ContentColor.Paint(gc, rect, unison.Fill))
	}
	addButton := unison.NewSVGButton(svg.CircledAdd)
	addButton.ClickCallback = func() {
		if created := p.createFeatureForType(lastFeatureTypeUsed); created != nil {
			*features = slices.Insert(*features, 0, created)
			p.insertFeaturePanel(1, created)
			unison.Ancestor[*unison.DockContainer](p).MarkForLayoutRecursively()
			MarkModified(p)
		}
	}
	p.AddChild(addButton)
	for i, one := range *features {
		p.insertFeaturePanel(i+1, one)
	}
	return p
}

func (p *featuresPanel) insertFeaturePanel(index int, f gurps.Feature) {
	var panel *unison.Panel
	switch one := f.(type) {
	case *gurps.AttributeBonus:
		panel = p.createAttributeBonusPanel(one)
	case *gurps.ConditionalModifierBonus:
		panel = p.createConditionalModifierPanel(one)
	case *gurps.ContainedWeightReduction:
		panel = p.createContainedWeightReductionPanel(one)
	case *gurps.CostReduction:
		panel = p.createCostReductionPanel(one)
	case *gurps.DRBonus:
		panel = p.createDRBonusPanel(one)
	case *gurps.ReactionBonus:
		panel = p.createReactionBonusPanel(one)
	case *gurps.SkillBonus:
		panel = p.createSkillBonusPanel(one)
	case *gurps.SkillPointBonus:
		panel = p.createSkillPointBonusPanel(one)
	case *gurps.SpellBonus:
		panel = p.createSpellBonusPanel(one)
	case *gurps.SpellPointBonus:
		panel = p.createSpellPointBonusPanel(one)
	case *gurps.WeaponBonus:
		panel = p.createWeaponDamageBonusPanel(one)
	default:
		jot.Warn(errs.Newf("unknown feature type: %s", reflect.TypeOf(f).String()))
		return
	}
	if panel != nil {
		panel.SetLayoutData(&unison.FlexLayoutData{
			HAlign: unison.FillAlignment,
			HGrab:  true,
		})
		p.AddChildAtIndex(panel, index)
	}
}

func (p *featuresPanel) createBasePanel(f gurps.Feature) *unison.Panel {
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HAlign:   unison.FillAlignment,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	deleteButton := unison.NewSVGButton(svg.Trash)
	deleteButton.ClickCallback = func() {
		if i := slices.IndexFunc(*p.features, func(elem gurps.Feature) bool { return elem == f }); i != -1 {
			*p.features = slices.Delete(*p.features, i, i+1)
		}
		panel.RemoveFromParent()
		unison.Ancestor[*unison.DockContainer](p).MarkForLayoutRecursively()
		MarkModified(p)
	}
	panel.AddChild(deleteButton)
	return panel
}

func (p *featuresPanel) createAttributeBonusPanel(f *gurps.AttributeBonus) *unison.Panel {
	panel := p.createBasePanel(f)
	p.addLeveledModifierLine(panel, f, &f.LeveledAmount)
	panel.AddChild(unison.NewPanel())
	wrapper := unison.NewPanel()
	var limitationPopup *unison.PopupMenu[gurps.BonusLimitation]
	attrChoicePopup := addAttributeChoicePopup(wrapper, p.entity, i18n.Text("to"), &f.Attribute,
		gurps.SizeFlag|gurps.DodgeFlag|gurps.ParryFlag|gurps.BlockFlag)
	callback := attrChoicePopup.SelectionChangedCallback
	attrChoicePopup.SelectionChangedCallback = func(popup *unison.PopupMenu[*gurps.AttributeChoice]) {
		if item, ok := popup.Selected(); ok {
			lastAttributeIDUsed = item.Key
			callback(popup)
			adjustPopupBlank(limitationPopup, f.Attribute != gurps.StrengthID)
		}
	}
	limitationPopup = addPopup(wrapper, gurps.AllBonusLimitation, &f.Limitation)
	adjustPopupBlank(limitationPopup, f.Attribute != gurps.StrengthID)
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  len(wrapper.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	panel.AddChild(wrapper)
	return panel
}

func (p *featuresPanel) createConditionalModifierPanel(f *gurps.ConditionalModifierBonus) *unison.Panel {
	panel := p.createBasePanel(f)
	p.addLeveledModifierLine(panel, f, &f.LeveledAmount)
	panel.AddChild(unison.NewPanel())
	watermark := i18n.Text("Triggering Condition")
	field := NewMultiLineStringField(nil, "", watermark, func() string { return f.Situation },
		func(value string) {
			f.Situation = value
			panel.MarkForLayoutAndRedraw()
			MarkModified(panel)
		})
	field.Watermark = watermark
	field.AutoScroll = false
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	panel.AddChild(field)
	return panel
}

func (p *featuresPanel) createDRBonusPanel(f *gurps.DRBonus) *unison.Panel {
	panel := p.createBasePanel(f)
	p.addLeveledModifierLine(panel, f, &f.LeveledAmount)

	panel.AddChild(unison.NewPanel())
	addHitLocationChoicePopup(panel, p.entity, i18n.Text("to the"), &f.Location)

	panel.AddChild(unison.NewPanel())
	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.AddChild(NewFieldLeadingLabel(i18n.Text("against")))
	field := NewStringField(nil, "", i18n.Text("Specialization"), func() string { return f.Specialization },
		func(value string) {
			f.Specialization = value
			f.Normalize()
			MarkModified(wrapper)
		})
	field.Watermark = gurps.AllID
	field.SetMinimumTextWidthUsing("Specialization")
	wrapper.AddChild(field)
	wrapper.AddChild(NewFieldTrailingLabel(i18n.Text("attacks")))
	panel.AddChild(wrapper)
	return panel
}

func (p *featuresPanel) createReactionBonusPanel(f *gurps.ReactionBonus) *unison.Panel {
	panel := p.createBasePanel(f)
	p.addLeveledModifierLine(panel, f, &f.LeveledAmount)
	panel.AddChild(unison.NewPanel())
	watermark := i18n.Text("from/to target group")
	field := NewMultiLineStringField(nil, "", watermark, func() string { return f.Situation },
		func(value string) {
			f.Situation = value
			panel.MarkForLayoutAndRedraw()
			MarkModified(panel)
		})
	field.Watermark = watermark
	field.AutoScroll = false
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	panel.AddChild(field)
	return panel
}

func (p *featuresPanel) createSkillBonusPanel(f *gurps.SkillBonus) *unison.Panel {
	panel := p.createBasePanel(f)
	p.addLeveledModifierLine(panel, f, &f.LeveledAmount)

	panel.AddChild(unison.NewPanel())
	wrapper := unison.NewPanel()
	var criteriaPopup *unison.PopupMenu[string]
	var criteriaField *StringField
	popup := addPopup(wrapper, gurps.AllSkillSelectionType, &f.SelectionType)
	popup.ChoiceMadeCallback = func(pop *unison.PopupMenu[gurps.SkillSelectionType], index int, item gurps.SkillSelectionType) {
		pop.SelectIndex(index)
		count := 4
		if f.SelectionType == gurps.ThisWeaponSkillSelectionType {
			count = 2
		}
		f.SelectionType = item
		adjustPopupBlank(criteriaPopup, f.SelectionType == gurps.ThisWeaponSkillSelectionType)
		adjustFieldBlank(criteriaField, f.SelectionType == gurps.ThisWeaponSkillSelectionType)
		i := panel.IndexOfChild(wrapper) + 1
		for j := count - 1; j >= 0; j-- {
			panel.RemoveChildAtIndex(i + j)
		}
		p.createSecondarySkillPanels(panel, i, f)
		unison.Ancestor[*unison.DockContainer](p).MarkForLayoutRecursively()
		MarkModified(p)
	}
	criteriaPopup, criteriaField = addStringCriteriaPanel(wrapper, "", "", i18n.Text("Name Qualifier"), &f.NameCriteria, 1, false)
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  len(wrapper.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
	})
	panel.AddChild(wrapper)
	adjustPopupBlank(criteriaPopup, f.SelectionType == gurps.ThisWeaponSkillSelectionType)
	adjustFieldBlank(criteriaField, f.SelectionType == gurps.ThisWeaponSkillSelectionType)

	p.createSecondarySkillPanels(panel, len(panel.Children()), f)
	return panel
}

func (p *featuresPanel) createSecondarySkillPanels(parent *unison.Panel, index int, f *gurps.SkillBonus) {
	parent.AddChildAtIndex(unison.NewPanel(), index)
	index++
	wrapper := unison.NewPanel()
	switch f.SelectionType {
	case gurps.NameSkillSelectionType:
		addSpecializationCriteriaPanel(wrapper, &f.SpecializationCriteria, 1, false)
	case gurps.ThisWeaponSkillSelectionType, gurps.WeaponsWithNameSkillSelectionType:
		prefix := i18n.Text("and whose usage")
		addStringCriteriaPanel(wrapper, prefix, prefix, i18n.Text("Usage Qualifier"), &f.SpecializationCriteria, 1, false)
	default:
		jot.Errorf("unknown selection type: %v", f.SelectionType)
	}
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  len(wrapper.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
	})
	parent.AddChildAtIndex(wrapper, index)
	index++

	if f.SelectionType != gurps.ThisWeaponSkillSelectionType {
		parent.AddChildAtIndex(unison.NewPanel(), index)
		index++
		wrapper = unison.NewPanel()
		addTagCriteriaPanel(wrapper, &f.TagsCriteria, 1, false)
		wrapper.SetLayout(&unison.FlexLayout{
			Columns:  len(wrapper.Children()),
			HSpacing: unison.StdHSpacing,
			VSpacing: unison.StdVSpacing,
		})
		wrapper.SetLayoutData(&unison.FlexLayoutData{
			HAlign: unison.FillAlignment,
		})
		parent.AddChildAtIndex(wrapper, index)
	}
}

func (p *featuresPanel) createSkillPointBonusPanel(f *gurps.SkillPointBonus) *unison.Panel {
	panel := p.createBasePanel(f)
	p.addLeveledModifierLine(panel, f, &f.LeveledAmount)
	prefix := i18n.Text("to skills whose name")
	addStringCriteriaPanel(panel, prefix, prefix, i18n.Text("Name Qualifier"), &f.NameCriteria, 1, true)
	addSpecializationCriteriaPanel(panel, &f.SpecializationCriteria, 1, true)
	addTagCriteriaPanel(panel, &f.TagsCriteria, 1, true)
	return panel
}

func (p *featuresPanel) createSpellBonusPanel(f *gurps.SpellBonus) *unison.Panel {
	panel := p.createBasePanel(f)
	p.addLeveledModifierLine(panel, f, &f.LeveledAmount)
	panel.AddChild(unison.NewPanel())
	wrapper := unison.NewPanel()
	var criteriaPopup *unison.PopupMenu[string]
	var criteriaField *StringField
	popup := addPopup(wrapper, gurps.AllSpellMatchType, &f.SpellMatchType)
	popup.ChoiceMadeCallback = func(pop *unison.PopupMenu[gurps.SpellMatchType], index int, item gurps.SpellMatchType) {
		pop.SelectIndex(index)
		f.SpellMatchType = item
		adjustPopupBlank(criteriaPopup, f.SpellMatchType == gurps.AllCollegesSpellMatchType)
		adjustFieldBlank(criteriaField, f.SpellMatchType == gurps.AllCollegesSpellMatchType)
		MarkModified(p)
	}
	criteriaPopup, criteriaField = addStringCriteriaPanel(wrapper, "", "", i18n.Text("Name Qualifier"), &f.NameCriteria, 1, false)
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  len(wrapper.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
	})
	panel.AddChild(wrapper)
	adjustPopupBlank(criteriaPopup, f.SpellMatchType == gurps.AllCollegesSpellMatchType)
	adjustFieldBlank(criteriaField, f.SpellMatchType == gurps.AllCollegesSpellMatchType)

	addTagCriteriaPanel(panel, &f.TagsCriteria, 1, true)
	return panel
}

func (p *featuresPanel) createSpellPointBonusPanel(f *gurps.SpellPointBonus) *unison.Panel {
	panel := p.createBasePanel(f)
	p.addLeveledModifierLine(panel, f, &f.LeveledAmount)
	panel.AddChild(unison.NewPanel())
	wrapper := unison.NewPanel()
	var criteriaPopup *unison.PopupMenu[string]
	var criteriaField *StringField
	popup := addPopup(wrapper, gurps.AllSpellMatchType, &f.SpellMatchType)
	popup.ChoiceMadeCallback = func(pop *unison.PopupMenu[gurps.SpellMatchType], index int, item gurps.SpellMatchType) {
		pop.SelectIndex(index)
		f.SpellMatchType = item
		adjustPopupBlank(criteriaPopup, f.SpellMatchType == gurps.AllCollegesSpellMatchType)
		adjustFieldBlank(criteriaField, f.SpellMatchType == gurps.AllCollegesSpellMatchType)
		MarkModified(p)
	}
	criteriaPopup, criteriaField = addStringCriteriaPanel(wrapper, "", "", i18n.Text("Name Qualifier"), &f.NameCriteria, 1, false)
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  len(wrapper.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
	})
	panel.AddChild(wrapper)
	adjustPopupBlank(criteriaPopup, f.SpellMatchType == gurps.AllCollegesSpellMatchType)
	adjustFieldBlank(criteriaField, f.SpellMatchType == gurps.AllCollegesSpellMatchType)

	addTagCriteriaPanel(panel, &f.TagsCriteria, 1, true)
	return panel
}

func (p *featuresPanel) createWeaponDamageBonusPanel(f *gurps.WeaponBonus) *unison.Panel {
	panel := p.createBasePanel(f)
	p.addLeveledModifierLine(panel, f, &f.LeveledAmount)

	panel.AddChild(unison.NewPanel())
	wrapper := unison.NewPanel()
	var criteriaPopup *unison.PopupMenu[string]
	var criteriaField *StringField
	popup := addPopup(wrapper, gurps.AllWeaponSelectionType, &f.SelectionType)
	popup.ChoiceMadeCallback = func(pop *unison.PopupMenu[gurps.WeaponSelectionType], index int, item gurps.WeaponSelectionType) {
		pop.SelectIndex(index)
		var count int
		switch f.SelectionType {
		case gurps.WithRequiredSkillWeaponSelectionType:
			count = 6
		case gurps.ThisWeaponWeaponSelectionType:
			count = 2
		case gurps.WithNameWeaponSelectionType:
			count = 4
		}
		f.SelectionType = item
		adjustPopupBlank(criteriaPopup, f.SelectionType == gurps.ThisWeaponWeaponSelectionType)
		adjustFieldBlank(criteriaField, f.SelectionType == gurps.ThisWeaponWeaponSelectionType)
		i := panel.IndexOfChild(wrapper) + 1
		for j := count - 1; j >= 0; j-- {
			panel.RemoveChildAtIndex(i + j)
		}
		p.createSecondaryWeaponPanels(panel, i, f)
		unison.Ancestor[*unison.DockContainer](p).MarkForLayoutRecursively()
		MarkModified(p)
	}
	criteriaPopup, criteriaField = addStringCriteriaPanel(wrapper, "", "", i18n.Text("Name Qualifier"), &f.NameCriteria, 1, false)
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  len(wrapper.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
	})
	panel.AddChild(wrapper)
	adjustPopupBlank(criteriaPopup, f.SelectionType == gurps.ThisWeaponWeaponSelectionType)
	adjustFieldBlank(criteriaField, f.SelectionType == gurps.ThisWeaponWeaponSelectionType)

	p.createSecondaryWeaponPanels(panel, len(panel.Children()), f)
	return panel
}

func (p *featuresPanel) createSecondaryWeaponPanels(parent *unison.Panel, index int, f *gurps.WeaponBonus) {
	parent.AddChildAtIndex(unison.NewPanel(), index)
	index++
	wrapper := unison.NewPanel()
	switch f.SelectionType {
	case gurps.WithRequiredSkillWeaponSelectionType:
		addSpecializationCriteriaPanel(wrapper, &f.SpecializationCriteria, 1, false)
	case gurps.ThisWeaponWeaponSelectionType, gurps.WithNameWeaponSelectionType:
		prefix := i18n.Text("and whose usage")
		addStringCriteriaPanel(wrapper, prefix, prefix, i18n.Text("Usage Qualifier"), &f.SpecializationCriteria, 1, false)
	default:
		jot.Errorf("unknown selection type: %v", f.SelectionType)
	}
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  len(wrapper.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
	})
	parent.AddChildAtIndex(wrapper, index)
	index++

	if f.SelectionType != gurps.ThisWeaponWeaponSelectionType {
		parent.AddChildAtIndex(unison.NewPanel(), index)
		index++
		wrapper = unison.NewPanel()
		addTagCriteriaPanel(wrapper, &f.TagsCriteria, 1, false)
		wrapper.SetLayout(&unison.FlexLayout{
			Columns:  len(wrapper.Children()),
			HSpacing: unison.StdHSpacing,
			VSpacing: unison.StdVSpacing,
		})
		wrapper.SetLayoutData(&unison.FlexLayoutData{
			HAlign: unison.FillAlignment,
		})
		parent.AddChildAtIndex(wrapper, index)
		index++

		if f.SelectionType != gurps.WithNameWeaponSelectionType {
			parent.AddChildAtIndex(unison.NewPanel(), index)
			index++
			wrapper = unison.NewPanel()
			addNumericCriteriaPanel(wrapper, nil, "", i18n.Text("and whose relative skill level"),
				i18n.Text("Level Qualifier"), &f.RelativeLevelCriteria, -fxp.Thousand, fxp.Thousand, 1, true, false)
			wrapper.SetLayout(&unison.FlexLayout{
				Columns:  len(wrapper.Children()),
				HSpacing: unison.StdHSpacing,
				VSpacing: unison.StdVSpacing,
			})
			wrapper.SetLayoutData(&unison.FlexLayoutData{
				HAlign: unison.FillAlignment,
			})
			parent.AddChildAtIndex(wrapper, index)
		}
	}
}

func (p *featuresPanel) createContainedWeightReductionPanel(f *gurps.ContainedWeightReduction) *unison.Panel {
	panel := p.createBasePanel(f)
	wrapper := unison.NewPanel()
	p.addTypeSwitcher(wrapper, f)

	field := NewStringField(nil, "", i18n.Text("Contained Weight Reduction"),
		func() string { return f.Reduction },
		func(value string) {
			f.Reduction, _ = gurps.ExtractContainedWeightReduction(value, gurps.SheetSettingsFor(p.entity).DefaultWeightUnits) //nolint:errcheck // A valid value is always returned
			MarkModified(wrapper)
		})
	field.SetMinimumTextWidthUsing("1,000 lb")
	field.Tooltip = unison.NewTooltipWithText(i18n.Text(`Enter a weight or percentage, e.g. "2 lb" or "5%"`))
	field.ValidateCallback = func() bool {
		_, err := gurps.ExtractContainedWeightReduction(field.Text(), gurps.SheetSettingsFor(p.entity).DefaultWeightUnits)
		return err == nil
	}
	wrapper.AddChild(field)

	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  len(wrapper.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
	})
	panel.AddChild(wrapper)
	return panel
}

func (p *featuresPanel) createCostReductionPanel(f *gurps.CostReduction) *unison.Panel {
	panel := p.createBasePanel(f)
	wrapper := unison.NewPanel()
	p.addTypeSwitcher(wrapper, f)
	addAttributeChoicePopup(wrapper, p.entity, "", &f.Attribute, gurps.SizeFlag|gurps.DodgeFlag|gurps.ParryFlag|gurps.BlockFlag)
	choices := make([]string, 0, 16)
	for i := 5; i <= 80; i += 5 {
		choices = append(choices, fmt.Sprintf(i18n.Text("by %d%%"), i))
	}
	choice := choices[xmath.Max(xmath.Min((fxp.As[int](f.Percentage)/5)-1, 15), 0)]
	addPopup(wrapper, choices, &choice).ChoiceMadeCallback = func(popup *unison.PopupMenu[string], index int, item string) {
		popup.SelectIndex(index)
		f.Percentage = fxp.From[int]((index + 1) * 5)
		MarkModified(wrapper)
	}
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  len(wrapper.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	panel.AddChild(wrapper)
	return panel
}

func (p *featuresPanel) addLeveledModifierLine(parent *unison.Panel, f gurps.Feature, amount *gurps.LeveledAmount) {
	panel := unison.NewPanel()
	p.addTypeSwitcher(panel, f)
	switch ft := f.(type) {
	case *gurps.WeaponBonus:
		var title string
		if ft.Type == gurps.WeaponBonusFeatureType {
			title = i18n.Text("per die")
		} else {
			title = i18n.Text("per level")
		}
		addLeveledAmountPanel(panel, nil, "", title, amount)
		addCheckBox(panel, i18n.Text("as a percentage"), &ft.Percent)
	default:
		addLeveledAmountPanel(panel, nil, "", i18n.Text("per level"), amount)
	}
	panel.SetLayout(&unison.FlexLayout{
		Columns:  len(panel.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	parent.AddChild(panel)
}

func (p *featuresPanel) featureTypesList() []gurps.FeatureType {
	if e, ok := p.owner.(*gurps.Equipment); ok && e.Container() {
		return gurps.AllFeatureType
	}
	return gurps.AllFeatureTypesWithoutContainedWeightType
}

func (p *featuresPanel) addTypeSwitcher(parent *unison.Panel, f gurps.Feature) {
	currentType := f.FeatureType()
	popup := addPopup(parent, p.featureTypesList(), &currentType)
	popup.ChoiceMadeCallback = func(pop *unison.PopupMenu[gurps.FeatureType], index int, item gurps.FeatureType) {
		pop.SelectIndex(index)
		if newFeature := p.createFeatureForType(item); newFeature != nil {
			lastFeatureTypeUsed = item
			parent.Parent().RemoveFromParent()
			list := *p.features
			i := slices.IndexFunc(list, func(one gurps.Feature) bool { return one == f })
			list[i] = newFeature
			p.insertFeaturePanel(i+1, newFeature)
			unison.Ancestor[*unison.DockContainer](p).MarkForLayoutRecursively()
			MarkModified(p)
		}
	}
}

func (p *featuresPanel) createFeatureForType(featureType gurps.FeatureType) gurps.Feature {
	var bonus gurps.Bonus
	switch featureType {
	case gurps.AttributeBonusFeatureType:
		bonus = gurps.NewAttributeBonus(lastAttributeIDUsed)
	case gurps.ConditionalModifierFeatureType:
		bonus = gurps.NewConditionalModifierBonus()
	case gurps.ContainedWeightReductionFeatureType:
		return gurps.NewContainedWeightReduction()
	case gurps.CostReductionFeatureType:
		return gurps.NewCostReduction(lastAttributeIDUsed)
	case gurps.DRBonusFeatureType:
		bonus = gurps.NewDRBonus()
	case gurps.ReactionBonusFeatureType:
		bonus = gurps.NewReactionBonus()
	case gurps.SkillBonusFeatureType:
		bonus = gurps.NewSkillBonus()
	case gurps.SkillPointBonusFeatureType:
		bonus = gurps.NewSkillPointBonus()
	case gurps.SpellBonusFeatureType:
		bonus = gurps.NewSpellBonus()
	case gurps.SpellPointBonusFeatureType:
		bonus = gurps.NewSpellPointBonus()
	case gurps.WeaponBonusFeatureType:
		bonus = gurps.NewWeaponDamageBonus()
	case gurps.WeaponDRDivisorBonusFeatureType:
		bonus = gurps.NewWeaponDRDivisorBonus()
	default:
		jot.Warn(errs.Newf("unknown feature type: %s", featureType.Key()))
		return nil
	}
	bonus.SetOwner(p.owner)
	return bonus
}
