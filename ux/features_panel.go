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
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellmatch"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xslices"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

var (
	lastFeatureTypeUsed = feature.AttributeBonus
	lastAttributeIDUsed = gurps.StrengthID
)

type featuresPanel struct {
	unison.Panel
	entity               *gurps.Entity
	owner                fmt.Stringer
	features             *gurps.Features
	currentLocation      int
	forEquipmentModifier bool
}

func newFeaturesPanel(entity *gurps.Entity, owner fmt.Stringer, features *gurps.Features, forEquipmentModifier bool) *featuresPanel {
	p := &featuresPanel{
		entity:               entity,
		owner:                owner,
		features:             features,
		forEquipmentModifier: forEquipmentModifier,
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	p.SetBorder(unison.NewCompoundBorder(
		&TitledBorder{
			Title: i18n.Text("Features"),
			Font:  unison.LabelFont,
		},
		unison.NewEmptyBorder(geom.NewUniformInsets(2))))
	p.DrawCallback = func(gc *unison.Canvas, rect geom.Rect) {
		gc.DrawRect(rect, unison.ThemeSurface.Paint(gc, rect, paintstyle.Fill))
	}
	addButton := unison.NewSVGButton(svg.CircledAdd)
	addButton.ClickCallback = func() {
		if created := p.createFeatureForType(lastFeatureTypeUsed); created != nil {
			*features = slices.Insert(*features, 0, created)
			p.insertFeaturePanel(1, created)
			MarkRootAncestorForLayoutRecursively(p)
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
	var panel, focus unison.Paneler
	switch one := f.(type) {
	case *gurps.AttributeBonus:
		panel, focus = p.createAttributeBonusPanel(one)
	case *gurps.ConditionalModifierBonus:
		panel, focus = p.createConditionalModifierPanel(one)
	case *gurps.ContainedWeightReduction:
		panel, focus = p.createContainedWeightReductionPanel(one)
	case *gurps.CostReduction:
		panel, focus = p.createCostReductionPanel(one)
	case *gurps.DRBonus:
		panel, focus = p.createDRBonusPanel(one)
		// PassiveDefenseBonus is also handled as DRBonus
	case *gurps.ReactionBonus:
		panel, focus = p.createReactionBonusPanel(one)
	case *gurps.SkillBonus:
		panel, focus = p.createSkillBonusPanel(one)
	case *gurps.SkillPointBonus:
		panel, focus = p.createSkillPointBonusPanel(one)
	case *gurps.SpellBonus:
		panel, focus = p.createSpellBonusPanel(one)
	case *gurps.SpellPointBonus:
		panel, focus = p.createSpellPointBonusPanel(one)
	case *gurps.TraitBonus:
		panel, focus = p.createTraitBonusPanel(one)
	case *gurps.WeaponBonus:
		panel, focus = p.createWeaponBonusPanel(one)
	default:
		errs.Log(errs.New("unknown feature type"), "type", reflect.TypeOf(f).String())
		return
	}
	if panel != nil {
		panel.AsPanel().SetLayoutData(&unison.FlexLayoutData{
			HAlign: align.Fill,
			HGrab:  true,
		})
		p.AddChildAtIndex(panel, index)
		focus.AsPanel().RequestFocus()
	}
}

func (p *featuresPanel) createBasePanel(f gurps.Feature) *unison.Panel {
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HAlign:   align.Fill,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	deleteButton := unison.NewSVGButton(svg.Trash)
	deleteButton.ClickCallback = func() {
		if i := slices.IndexFunc(*p.features, func(elem gurps.Feature) bool { return elem == f }); i != -1 {
			*p.features = slices.Delete(*p.features, i, i+1)
		}
		panel.RemoveFromParent()
		MarkRootAncestorForLayoutRecursively(p)
		MarkModified(p)
	}
	panel.AddChild(deleteButton)
	return panel
}

func (p *featuresPanel) createAttributeBonusPanel(f *gurps.AttributeBonus) (main *unison.Panel, focus unison.Paneler) {
	panel := p.createBasePanel(f)
	focus = p.addLeveledModifierLine(panel, f, &f.LeveledAmount)
	panel.AddChild(unison.NewPanel())
	wrapper := unison.NewPanel()
	var limitationPopup *unison.PopupMenu[stlimit.Option]
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
	limitationPopup = addPopup(wrapper, stlimit.Options, &f.Limitation)
	adjustPopupBlank(limitationPopup, f.Attribute != gurps.StrengthID)
	p.addWrapperAtIndex(panel, wrapper, -1, true)
	return panel, focus
}

func (p *featuresPanel) createConditionalModifierPanel(f *gurps.ConditionalModifierBonus) (main *unison.Panel, focus unison.Paneler) {
	panel := p.createBasePanel(f)
	focus = p.addLeveledModifierLine(panel, f, &f.LeveledAmount)
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
		HAlign: align.Fill,
		HGrab:  true,
	})
	panel.AddChild(field)
	return panel, focus
}

func (p *featuresPanel) createDRBonusPanel(f *gurps.DRBonus) (main *unison.Panel, focus unison.Paneler) {
	panel := p.createBasePanel(f)
	focus = p.addLeveledModifierLine(panel, f, &f.LeveledAmount)
	panel.AddChild(unison.NewPanel())
	panel.AddChild(p.createHitLocationChoicesPanel(f))
	panel.AddChild(unison.NewPanel())
	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.AddChild(NewFieldLeadingLabel(i18n.Text("against"), false))
	field := NewStringField(nil, "", i18n.Text("Specialization"), func() string { return f.Specialization },
		func(value string) {
			f.Specialization = value
			f.Normalize()
			MarkModified(wrapper)
		})
	field.Watermark = gurps.AllID
	field.SetMinimumTextWidthUsing("Specialization")
	wrapper.AddChild(field)
	wrapper.AddChild(NewFieldTrailingLabel(i18n.Text("attacks"), false))
	panel.AddChild(wrapper)
	return panel, focus
}

func (p *featuresPanel) createHitLocationChoicesPanel(f *gurps.DRBonus) *unison.Panel {
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	popup := unison.NewPopupMenu[string]()
	if p.forEquipmentModifier {
		popup.AddItem(i18n.Text("to this armor"))
	}
	popup.AddItem(i18n.Text("to all locations"))
	popup.AddItem(i18n.Text("to these locations:"))
	if len(f.Locations) == 0 {
		p.currentLocation = 0
	} else {
		if slices.Contains(f.Locations, gurps.AllID) {
			p.currentLocation = 0
		} else {
			p.currentLocation = 1
		}
		if p.forEquipmentModifier {
			p.currentLocation++
		}
	}
	popup.SelectIndex(p.currentLocation)
	popup.SelectionChangedCallback = func(pop *unison.PopupMenu[string]) {
		p.currentLocation = pop.SelectedIndex()
		for len(panel.Children()) > 1 {
			panel.RemoveChildAtIndex(1)
		}
		switch {
		case p.isCurrentLocationThisArmor():
			f.Locations = nil
		case p.isCurrentLocationAll():
			f.Locations = []string{gurps.AllID}
		case p.isCurrentLocationList():
			f.Locations = slices.DeleteFunc(f.Locations, func(loc string) bool { return loc == gurps.AllID })
			if len(f.Locations) == 0 {
				f.Locations = []string{gurps.TorsoID}
			}
			panel.AddChild(p.createHitLocationsCheckBoxes(f))
		}
		MarkModified(panel)
		panel.MarkForLayoutRecursivelyUpward()
	}
	panel.AddChild(popup)
	if p.isCurrentLocationList() {
		f.Locations = slices.DeleteFunc(f.Locations, func(loc string) bool { return loc == gurps.AllID })
		panel.AddChild(p.createHitLocationsCheckBoxes(f))
	}
	return panel
}

func (p *featuresPanel) createHitLocationsCheckBoxes(f *gurps.DRBonus) *unison.Panel {
	panel := unison.NewPanel()
	const desiredColumns = 4
	panel.SetLayout(&unison.FlexLayout{
		Columns:      desiredColumns,
		HSpacing:     unison.StdHSpacing,
		VSpacing:     unison.StdVSpacing,
		EqualColumns: true,
	})
	panel.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: unison.StdHSpacing * 2}))
	bodyType := gurps.BodyFor(p.entity)
	existing := xslices.MapFromKeys(f.Locations, func(in string) string { return in })
	locs := bodyType.UniqueHitLocations(p.entity)
	boxes := make([]*unison.CheckBox, 0, len(existing)+len(locs))
	for _, loc := range locs {
		box := unison.NewCheckBox()
		box.SetTitle(loc.ChoiceName)
		box.State = check.FromBool(slices.Contains(f.Locations, loc.LocID))
		box.ClickCallback = func() { toggleHitLocation(box, f, loc.LocID) }
		delete(existing, loc.LocID)
		boxes = append(boxes, box)
	}
	const unknownKey = "unknown"
	for loc := range maps.Keys(existing) {
		box := unison.NewCheckBox()
		box.SetTitle(loc + "*")
		box.State = check.On
		box.ClientData()[unknownKey] = true
		box.ClickCallback = func() { toggleHitLocation(box, f, loc) }
		boxes = append(boxes, box)
	}
	xslices.ColumnSort(boxes, desiredColumns, func(a, b *unison.CheckBox) int {
		_, au := a.ClientData()[unknownKey]
		_, bu := b.ClientData()[unknownKey]
		if au != bu {
			if au {
				return 1
			}
			return -1
		}
		return xstrings.NaturalCmp(a.Text.String(), b.Text.String(), true)
	})
	for _, box := range boxes {
		panel.AddChild(box)
	}
	if len(existing) != 0 {
		label := unison.NewLabel()
		fd := label.Font.Descriptor()
		fd.Size *= 0.8
		label.Font = fd.Font()
		label.SetTitle(i18n.Text("* Locations not present in current body type"))
		label.SetLayoutData(&unison.FlexLayoutData{HSpan: desiredColumns})
		panel.AddChild(label)
	}
	return panel
}

func toggleHitLocation(box *unison.CheckBox, f *gurps.DRBonus, loc string) {
	if box.State == check.On {
		f.Locations = append(f.Locations, loc)
		slices.Sort(f.Locations)
	} else {
		f.Locations = slices.DeleteFunc(f.Locations, func(in string) bool { return in == loc })
	}
	MarkModified(box)
}

func (p *featuresPanel) isCurrentLocationThisArmor() bool {
	return p.forEquipmentModifier && p.currentLocation == 0
}

func (p *featuresPanel) isCurrentLocationAll() bool {
	return (p.forEquipmentModifier && p.currentLocation == 1) || (!p.forEquipmentModifier && p.currentLocation == 0)
}

func (p *featuresPanel) isCurrentLocationList() bool {
	return (p.forEquipmentModifier && p.currentLocation == 2) || (!p.forEquipmentModifier && p.currentLocation == 1)
}

func (p *featuresPanel) createReactionBonusPanel(f *gurps.ReactionBonus) (main *unison.Panel, focus unison.Paneler) {
	panel := p.createBasePanel(f)
	focus = p.addLeveledModifierLine(panel, f, &f.LeveledAmount)
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
		HAlign: align.Fill,
		HGrab:  true,
	})
	panel.AddChild(field)
	return panel, focus
}

func (p *featuresPanel) createSkillBonusPanel(f *gurps.SkillBonus) (main *unison.Panel, focus unison.Paneler) {
	panel := p.createBasePanel(f)
	focus = p.addLeveledModifierLine(panel, f, &f.LeveledAmount)
	panel.AddChild(unison.NewPanel())
	wrapper := unison.NewPanel()
	var criteriaPopup *unison.PopupMenu[string]
	var criteriaField *StringField
	popup := addPopup(wrapper, skillsel.Types, &f.SelectionType)
	popup.ChoiceMadeCallback = func(pop *unison.PopupMenu[skillsel.Type], index int, item skillsel.Type) {
		pop.SelectIndex(index)
		f.SelectionType = item
		adjustPopupBlank(criteriaPopup, f.SelectionType == skillsel.ThisWeapon)
		adjustFieldBlank(criteriaField, f.SelectionType == skillsel.ThisWeapon)
		i := panel.IndexOfChild(wrapper) + 1
		for j := len(panel.Children()) - 1; j >= i; j-- {
			panel.RemoveChildAtIndex(j)
		}
		p.createSecondarySkillPanels(panel, i, f)
		MarkRootAncestorForLayoutRecursively(p)
		MarkModified(p)
	}
	criteriaPopup, criteriaField = addStringCriteriaPanel(wrapper, "", "", i18n.Text("Name Qualifier"), &f.NameCriteria, 1, false)
	p.addWrapperAtIndex(panel, wrapper, -1, false)
	adjustPopupBlank(criteriaPopup, f.SelectionType == skillsel.ThisWeapon)
	adjustFieldBlank(criteriaField, f.SelectionType == skillsel.ThisWeapon)
	p.createSecondarySkillPanels(panel, len(panel.Children()), f)
	return panel, focus
}

func (p *featuresPanel) createSecondarySkillPanels(parent *unison.Panel, index int, f *gurps.SkillBonus) {
	var wrapper *unison.Panel
	wrapper, index = p.prepareNewWrapper(parent, index)
	switch f.SelectionType {
	case skillsel.Name:
		addSpecializationCriteriaPanel(wrapper, &f.SpecializationCriteria, 1, false)
	case skillsel.ThisWeapon, skillsel.WeaponsWithName:
		prefix := i18n.Text("and whose usage")
		addStringCriteriaPanel(wrapper, prefix, prefix, i18n.Text("Usage Qualifier"), &f.SpecializationCriteria, 1, false)
	default:
		errs.Log(errs.New("unknown selection type"), "type", int(f.SelectionType))
	}
	index = p.addWrapperAtIndex(parent, wrapper, index, false)
	if f.SelectionType != skillsel.ThisWeapon {
		wrapper, index = p.prepareNewWrapper(parent, index)
		addTagCriteriaPanel(wrapper, &f.TagsCriteria, 1, false)
		p.addWrapperAtIndex(parent, wrapper, index, false)
	}
}

func (p *featuresPanel) createSkillPointBonusPanel(f *gurps.SkillPointBonus) (main *unison.Panel, focus unison.Paneler) {
	panel := p.createBasePanel(f)
	focus = p.addLeveledModifierLine(panel, f, &f.LeveledAmount)
	prefix := i18n.Text("to skills whose name")
	addStringCriteriaPanel(panel, prefix, prefix, i18n.Text("Name Qualifier"), &f.NameCriteria, 1, true)
	addSpecializationCriteriaPanel(panel, &f.SpecializationCriteria, 1, true)
	addTagCriteriaPanel(panel, &f.TagsCriteria, 1, true)
	return panel, focus
}

func (p *featuresPanel) createSpellBonusPanel(f *gurps.SpellBonus) (main *unison.Panel, focus unison.Paneler) {
	panel := p.createBasePanel(f)
	focus = p.addLeveledModifierLine(panel, f, &f.LeveledAmount)
	panel.AddChild(unison.NewPanel())
	wrapper := unison.NewPanel()
	var criteriaPopup *unison.PopupMenu[string]
	var criteriaField *StringField
	popup := addPopup(wrapper, spellmatch.Types, &f.SpellMatchType)
	popup.ChoiceMadeCallback = func(pop *unison.PopupMenu[spellmatch.Type], index int, item spellmatch.Type) {
		pop.SelectIndex(index)
		f.SpellMatchType = item
		adjustPopupBlank(criteriaPopup, f.SpellMatchType == spellmatch.AllColleges)
		adjustFieldBlank(criteriaField, f.SpellMatchType == spellmatch.AllColleges)
		MarkModified(p)
	}
	criteriaPopup, criteriaField = addStringCriteriaPanel(wrapper, "", "", i18n.Text("Name Qualifier"), &f.NameCriteria, 1, false)
	p.addWrapperAtIndex(panel, wrapper, -1, false)
	adjustPopupBlank(criteriaPopup, f.SpellMatchType == spellmatch.AllColleges)
	adjustFieldBlank(criteriaField, f.SpellMatchType == spellmatch.AllColleges)
	addTagCriteriaPanel(panel, &f.TagsCriteria, 1, true)
	return panel, focus
}

func (p *featuresPanel) createSpellPointBonusPanel(f *gurps.SpellPointBonus) (main *unison.Panel, focus unison.Paneler) {
	panel := p.createBasePanel(f)
	focus = p.addLeveledModifierLine(panel, f, &f.LeveledAmount)
	wrapper, _ := p.prepareNewWrapper(panel, -1)
	var criteriaPopup *unison.PopupMenu[string]
	var criteriaField *StringField
	popup := addPopup(wrapper, spellmatch.Types, &f.SpellMatchType)
	popup.ChoiceMadeCallback = func(pop *unison.PopupMenu[spellmatch.Type], index int, item spellmatch.Type) {
		pop.SelectIndex(index)
		f.SpellMatchType = item
		adjustPopupBlank(criteriaPopup, f.SpellMatchType == spellmatch.AllColleges)
		adjustFieldBlank(criteriaField, f.SpellMatchType == spellmatch.AllColleges)
		MarkModified(p)
	}
	criteriaPopup, criteriaField = addStringCriteriaPanel(wrapper, "", "", i18n.Text("Name Qualifier"), &f.NameCriteria, 1, false)
	p.addWrapperAtIndex(panel, wrapper, -1, false)
	adjustPopupBlank(criteriaPopup, f.SpellMatchType == spellmatch.AllColleges)
	adjustFieldBlank(criteriaField, f.SpellMatchType == spellmatch.AllColleges)
	addTagCriteriaPanel(panel, &f.TagsCriteria, 1, true)
	return panel, focus
}

func (p *featuresPanel) createTraitBonusPanel(f *gurps.TraitBonus) (main *unison.Panel, focus unison.Paneler) {
	panel := p.createBasePanel(f)
	focus = p.addLeveledModifierLine(panel, f, &f.LeveledAmount)
	prefix := i18n.Text("to traits whose name")
	addStringCriteriaPanel(panel, prefix, prefix, i18n.Text("Name Qualifier"), &f.NameCriteria, 1, true)
	addTagCriteriaPanel(panel, &f.TagsCriteria, 1, true)
	return panel, focus
}

func (p *featuresPanel) createWeaponBonusPanel(f *gurps.WeaponBonus) (main *unison.Panel, focus unison.Paneler) {
	panel := p.createBasePanel(f)
	_, focus = p.addWeaponLeveledModifierLine(panel, f)
	panel.AddChild(unison.NewPanel())
	wrapper := unison.NewPanel()
	var criteriaPopup *unison.PopupMenu[string]
	var criteriaField *StringField
	popup := addPopup(wrapper, wsel.Types, &f.SelectionType)
	popup.ChoiceMadeCallback = func(pop *unison.PopupMenu[wsel.Type], index int, item wsel.Type) {
		pop.SelectIndex(index)
		f.SelectionType = item
		p.adjustCriteriaPopupAndField(f, criteriaPopup, criteriaField)
		i := panel.IndexOfChild(wrapper) + 1
		for j := len(panel.Children()) - 1; j >= i; j-- {
			panel.RemoveChildAtIndex(j)
		}
		p.createSecondaryWeaponPanels(panel, i, f)
		MarkRootAncestorForLayoutRecursively(p)
		MarkModified(p)
	}
	criteriaPopup, criteriaField = addStringCriteriaPanel(wrapper, "", "", i18n.Text("Name Qualifier"), &f.NameCriteria, 1, false)
	p.addWrapperAtIndex(panel, wrapper, -1, false)
	p.adjustCriteriaPopupAndField(f, criteriaPopup, criteriaField)
	p.createSecondaryWeaponPanels(panel, len(panel.Children()), f)
	return panel, focus
}

func (p *featuresPanel) adjustCriteriaPopupAndField(f *gurps.WeaponBonus, criteriaPopup *unison.PopupMenu[string], criteriaField *StringField) {
	blank := f.SelectionType == wsel.ThisWeapon
	if !blank {
		blank = criteria.AllStringComparisons[criteriaPopup.SelectedIndex()] == criteria.AnyText
	}
	adjustPopupBlank(criteriaPopup, f.SelectionType == wsel.ThisWeapon)
	adjustFieldBlank(criteriaField, blank)
}

func (p *featuresPanel) createSecondaryWeaponPanels(parent *unison.Panel, index int, f *gurps.WeaponBonus) {
	var wrapper *unison.Panel
	wrapper, index = p.prepareNewWrapper(parent, index)
	switch f.SelectionType {
	case wsel.WithRequiredSkill:
		addSpecializationCriteriaPanel(wrapper, &f.SpecializationCriteria, 1, false)
		wrapper, index = p.prepareNewWrapper(parent, p.addWrapperAtIndex(parent, wrapper, index, false))
		addUsageCriteriaPanel(wrapper, &f.UsageCriteria, 1, false)
	case wsel.ThisWeapon, wsel.WithName:
		addUsageCriteriaPanel(wrapper, &f.SpecializationCriteria, 1, false)
	default:
		errs.Log(errs.New("unknown selection type"), "type", int(f.SelectionType))
	}
	index = p.addWrapperAtIndex(parent, wrapper, index, false)

	if f.SelectionType != wsel.ThisWeapon {
		wrapper, index = p.prepareNewWrapper(parent, index)
		addTagCriteriaPanel(wrapper, &f.TagsCriteria, 1, false)
		index = p.addWrapperAtIndex(parent, wrapper, index, false)
		if f.SelectionType != wsel.WithName {
			wrapper, index = p.prepareNewWrapper(parent, index)
			addNumericCriteriaPanel(wrapper, nil, "", i18n.Text("and whose relative skill level"),
				i18n.Text("Level Qualifier"), &f.RelativeLevelCriteria, -fxp.Thousand, fxp.Thousand, 1, true, false)
			p.addWrapperAtIndex(parent, wrapper, index, false)
		}
	}
}

func (p *featuresPanel) prepareNewWrapper(parent *unison.Panel, index int) (wrapper *unison.Panel, newIndex int) {
	parent.AddChildAtIndex(unison.NewPanel(), index)
	index++
	return unison.NewPanel(), index
}

func (p *featuresPanel) addWrapperAtIndex(parent, wrapper *unison.Panel, index int, hgrab bool) int {
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  len(wrapper.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  hgrab,
	})
	parent.AddChildAtIndex(wrapper, index)
	index++
	return index
}

func (p *featuresPanel) createContainedWeightReductionPanel(f *gurps.ContainedWeightReduction) (main *unison.Panel, focus unison.Paneler) {
	panel := p.createBasePanel(f)
	wrapper := unison.NewPanel()
	p.addTypeSwitcher(wrapper, f)
	field := NewStringField(nil, "", i18n.Text("Contained Weight Reduction"),
		func() string { return f.Reduction },
		func(value string) {
			//nolint:errcheck // A valid value is always returned
			f.Reduction, _ = gurps.ExtractContainedWeightReduction(value,
				gurps.SheetSettingsFor(p.entity).DefaultWeightUnits)
			MarkModified(wrapper)
		})
	field.SetMinimumTextWidthUsing("1,000 lb")
	field.Tooltip = newWrappedTooltip(i18n.Text(`Enter a weight or percentage, e.g. "2 lb" or "5%"`))
	field.ValidateCallback = func() bool {
		_, err := gurps.ExtractContainedWeightReduction(field.Text(), gurps.SheetSettingsFor(p.entity).DefaultWeightUnits)
		return err == nil
	}
	wrapper.AddChild(field)
	p.addWrapperAtIndex(panel, wrapper, -1, false)
	return panel, field
}

func (p *featuresPanel) createCostReductionPanel(f *gurps.CostReduction) (main *unison.Panel, focus unison.Paneler) {
	panel := p.createBasePanel(f)
	wrapper := unison.NewPanel()
	p.addTypeSwitcher(wrapper, f)
	addAttributeChoicePopup(wrapper, p.entity, "", &f.Attribute, gurps.SizeFlag|gurps.DodgeFlag|gurps.ParryFlag|gurps.BlockFlag)
	choices := make([]string, 0, 16)
	for i := 5; i <= 80; i += 5 {
		choices = append(choices, fmt.Sprintf(i18n.Text("by %d%%"), i))
	}
	choice := choices[max(min((fxp.AsInteger[int](f.Percentage)/5)-1, 15), 0)]
	pop := addPopup(wrapper, choices, &choice)
	pop.ChoiceMadeCallback = func(popup *unison.PopupMenu[string], index int, _ string) {
		popup.SelectIndex(index)
		f.Percentage = fxp.FromInteger((index + 1) * 5)
		MarkModified(wrapper)
	}
	p.addWrapperAtIndex(panel, wrapper, -1, true)
	return panel, pop
}

func (p *featuresPanel) addLeveledModifierLine(parent *unison.Panel, f gurps.Feature, amount *gurps.LeveledAmount) *DecimalField {
	panel := unison.NewPanel()
	p.addTypeSwitcher(panel, f)
	field, _ := addLeveledAmountPanel(panel, nil, "", i18n.Text("per level"), amount)
	panel.SetLayout(&unison.FlexLayout{
		Columns:  len(panel.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	parent.AddChild(panel)
	return field
}

func (p *featuresPanel) addWeaponLeveledModifierLine(parent *unison.Panel, wb *gurps.WeaponBonus) (main *unison.Panel, focus unison.Paneler) {
	panel := unison.NewPanel()
	switcher := p.addTypeSwitcher(panel, wb)
	if wb.Type == feature.WeaponSwitch {
		wrapper := unison.NewPanel()
		wrapper.AddChild(switcher)
		switcher.SetLayoutData(&unison.FlexLayoutData{HSpan: 3})
		if wb.SwitchType == wswitch.NotSwitched {
			wb.SwitchType = wswitch.Types[1]
		}
		spacer := unison.NewPanel()
		spacer.SetLayoutData(&unison.FlexLayoutData{SizeHint: geom.Size{Width: 16}})
		wrapper.AddChild(spacer)
		addPopup(wrapper, wswitch.Types[1:], &wb.SwitchType)
		focus = addBoolPopup(wrapper, i18n.Text("to true"), i18n.Text("to false"), &wb.SwitchTypeValue)
		wrapper.SetLayout(&unison.FlexLayout{
			Columns:  3,
			HSpacing: unison.StdHSpacing,
			VSpacing: unison.StdVSpacing,
		})
		wrapper.SetLayoutData(&unison.FlexLayoutData{
			HAlign: align.Fill,
			HGrab:  true,
		})
		panel.AddChild(wrapper)
	} else {
		field := NewDecimalField(nil, "", i18n.Text("Amount"),
			func() fxp.Int { return wb.Amount },
			func(value fxp.Int) {
				wb.Amount = value
				MarkModified(panel)
			}, fxp.Min, fxp.Max, true, false)
		focus = field
		panel.AddChild(field)
		addCheckBox(panel, i18n.Text("per level"), &wb.PerLevel)
		if wb.Type != feature.WeaponMinSTBonus && wb.Type != feature.WeaponEffectiveSTBonus {
			// Can't allow the per-die option for MinST bonuses, since that would cause an infinite loop on
			// resolution.
			addCheckBox(panel, i18n.Text("per die"), &wb.PerDie)
		}
		addCheckBox(panel, i18n.Text("as a %"), &wb.Percent)
	}
	panel.SetLayout(&unison.FlexLayout{
		Columns:  len(panel.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	parent.AddChild(panel)
	return panel, focus
}

func (p *featuresPanel) featureTypesList() []feature.Type {
	if e, ok := p.owner.(*gurps.Equipment); ok && e.Container() {
		return feature.Types
	}
	return feature.TypesWithoutContainedWeightReduction
}

func (p *featuresPanel) addTypeSwitcher(parent *unison.Panel, f gurps.Feature) *unison.PopupMenu[feature.Type] {
	currentType := f.FeatureType()
	popup := addPopup(parent, p.featureTypesList(), &currentType)
	popup.ChoiceMadeCallback = func(pop *unison.PopupMenu[feature.Type], index int, item feature.Type) {
		pop.SelectIndex(index)
		if newFeature := p.createFeatureForType(item); newFeature != nil {
			lastFeatureTypeUsed = item
			parent.Parent().RemoveFromParent()
			list := *p.features
			i := slices.IndexFunc(list, func(one gurps.Feature) bool { return one == f })
			list[i] = newFeature
			p.insertFeaturePanel(i+1, newFeature)
			MarkRootAncestorForLayoutRecursively(p)
			MarkModified(p)
		}
	}
	return popup
}

func (p *featuresPanel) createFeatureForType(featureType feature.Type) gurps.Feature {
	var bonus gurps.Bonus
	switch featureType {
	case feature.AttributeBonus:
		bonus = gurps.NewAttributeBonus(lastAttributeIDUsed)
	case feature.ConditionalModifier:
		bonus = gurps.NewConditionalModifierBonus()
	case feature.ContainedWeightReduction:
		return gurps.NewContainedWeightReduction()
	case feature.CostReduction:
		return gurps.NewCostReduction(lastAttributeIDUsed)
	case feature.DRBonus:
		bonus = gurps.NewDRBonus()
	case feature.PassiveDefenseBonus:
		bonus = gurps.NewPassiveDefenseBonus()
	case feature.ReactionBonus:
		bonus = gurps.NewReactionBonus()
	case feature.SkillBonus:
		bonus = gurps.NewSkillBonus()
	case feature.SkillPointBonus:
		bonus = gurps.NewSkillPointBonus()
	case feature.SpellBonus:
		bonus = gurps.NewSpellBonus()
	case feature.SpellPointBonus:
		bonus = gurps.NewSpellPointBonus()
	case feature.TraitBonus:
		bonus = gurps.NewTraitBonus()
	case feature.WeaponBonus:
		bonus = gurps.NewWeaponDamageBonus()
	case feature.WeaponAccBonus:
		bonus = gurps.NewWeaponAccBonus()
	case feature.WeaponScopeAccBonus:
		bonus = gurps.NewWeaponScopeAccBonus()
	case feature.WeaponDRDivisorBonus:
		bonus = gurps.NewWeaponDRDivisorBonus()
	case feature.WeaponEffectiveSTBonus:
		bonus = gurps.NewWeaponEffectiveSTBonus()
	case feature.WeaponMinSTBonus:
		bonus = gurps.NewWeaponMinSTBonus()
	case feature.WeaponMinReachBonus:
		bonus = gurps.NewWeaponMinReachBonus()
	case feature.WeaponMaxReachBonus:
		bonus = gurps.NewWeaponMaxReachBonus()
	case feature.WeaponHalfDamageRangeBonus:
		bonus = gurps.NewWeaponHalfDamageRangeBonus()
	case feature.WeaponMinRangeBonus:
		bonus = gurps.NewWeaponMinRangeBonus()
	case feature.WeaponMaxRangeBonus:
		bonus = gurps.NewWeaponMaxRangeBonus()
	case feature.WeaponBulkBonus:
		bonus = gurps.NewWeaponBulkBonus()
	case feature.WeaponRecoilBonus:
		bonus = gurps.NewWeaponRecoilBonus()
	case feature.WeaponParryBonus:
		bonus = gurps.NewWeaponParryBonus()
	case feature.WeaponBlockBonus:
		bonus = gurps.NewWeaponBlockBonus()
	case feature.WeaponRofMode1ShotsBonus:
		bonus = gurps.NewWeaponRofMode1ShotsBonus()
	case feature.WeaponRofMode1SecondaryBonus:
		bonus = gurps.NewWeaponRofMode1SecondaryBonus()
	case feature.WeaponRofMode2ShotsBonus:
		bonus = gurps.NewWeaponRofMode2ShotsBonus()
	case feature.WeaponRofMode2SecondaryBonus:
		bonus = gurps.NewWeaponRofMode2SecondaryBonus()
	case feature.WeaponNonChamberShotsBonus:
		bonus = gurps.NewWeaponNonChamberShotsBonus()
	case feature.WeaponChamberShotsBonus:
		bonus = gurps.NewWeaponChamberShotsBonus()
	case feature.WeaponShotDurationBonus:
		bonus = gurps.NewWeaponShotDurationBonus()
	case feature.WeaponReloadTimeBonus:
		bonus = gurps.NewWeaponReloadTimeBonus()
	case feature.WeaponSwitch:
		bonus = gurps.NewWeaponSwitchBonus()
	default:
		errs.Log(errs.New("unknown feature type"), "type", featureType.Key())
		return nil
	}
	bonus.SetOwner(p.owner)
	return bonus
}
