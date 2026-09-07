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
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/equipmentsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/maxusesmod"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellmatch"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/traitsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xslices"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
)

var (
	lastFeatureTypeUsed   = feature.AttributeBonus
	lastAttributeIDUsed   = gurps.StrengthID
	lastSelectorFieldUsed = selector.WeaponDamageType
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
	initTitledEditorSection(p, i18n.Text("Features"))
	p.AddChild(newSectionAddButton(p, func() bool {
		created := p.createFeatureForType(lastFeatureTypeUsed)
		if created == nil {
			return false
		}
		*features = slices.Insert(*features, 0, created)
		p.insertFeaturePanel(1, created)
		return true
	}))
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
	case *gurps.EquipmentMaxUsesBonus:
		panel, focus = p.createEquipmentMaxUsesBonusPanel(one)
	case *gurps.DRBonus:
		panel, focus = p.createDRBonusPanel(one)
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
	case *gurps.TraitMaxLevelBonus:
		panel, focus = p.createTraitMaxLevelBonusPanel(one)
	case *gurps.WeaponBonus:
		panel, focus = p.createWeaponBonusPanel(one)
	case *gurps.SelectorOverride:
		panel, focus = p.createSelectorOverridePanel(one)
	case *gurps.UnknownFeature:
		panel, focus = p.createUnknownFeaturePanel(one)
	default:
		errs.Log(errs.New("unknown feature type"), "type", reflect.TypeOf(f).String())
		return
	}
	if panel != nil {
		_, isSelectorOverride := f.(*gurps.SelectorOverride)
		panel.AsPanel().SetLayoutData(&unison.FlexLayoutData{
			HAlign: align.Fill,
			HGrab:  !isSelectorOverride,
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
	deleteButton := unison.NewSVGButton(unison.TrashSVG)
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

// addSelectionCriteriaRow adds the row shared by the bonuses that pick their targets by a selection type: a popup
// offering the types, followed by the name criteria. The name criteria are blanked while the chosen type needs no name
// (blank reports that, e.g. for a "this weapon" choice that targets the owner itself), and the name field is also
// blanked while its comparison accepts anything. When secondary is non-nil, it is called to add the rows that depend on
// the chosen type after the selection row, and those rows are torn down and rebuilt whenever the choice changes. It is
// a plain function rather than a method because methods cannot have type parameters.
func addSelectionCriteriaRow[E comparable](p *featuresPanel, panel *unison.Panel, types []E, sel *E, blank func(E) bool, nameCriteria *criteria.Text, secondary func(parent *unison.Panel, index int)) {
	panel.AddChild(unison.NewPanel())
	wrapper := unison.NewPanel()
	var criteriaPopup *unison.PopupMenu[string]
	var criteriaField *StringField
	adjust := func() {
		noName := blank(*sel)
		adjustPopupBlank(criteriaPopup, noName)
		adjustFieldBlank(criteriaField, noName || nameCriteria.IsZero())
	}
	popup := addPopup(wrapper, types, sel)
	popup.ChoiceMadeCallback = func(pop *unison.PopupMenu[E], index int, item E) {
		pop.SelectIndex(index)
		*sel = item
		adjust()
		if secondary != nil {
			i := panel.IndexOfChild(wrapper) + 1
			for j := len(panel.Children()) - 1; j >= i; j-- {
				panel.RemoveChildAtIndex(j)
			}
			secondary(panel, i)
			MarkRootAncestorForLayoutRecursively(p)
		}
		MarkModified(p)
	}
	criteriaPopup, criteriaField = addStringCriteriaPanel(wrapper, "", "", i18n.Text("Name Qualifier"), nameCriteria, 1, false)
	p.addWrapperAtIndex(panel, wrapper, -1, false)
	adjust()
	if secondary != nil {
		secondary(panel, len(panel.Children()))
	}
}

func (p *featuresPanel) createSkillBonusPanel(f *gurps.SkillBonus) (main *unison.Panel, focus unison.Paneler) {
	panel := p.createBasePanel(f)
	focus = p.addLeveledModifierLine(panel, f, &f.LeveledAmount)
	addSelectionCriteriaRow(p, panel, skillsel.Types, &f.SelectionType,
		func(t skillsel.Type) bool { return t == skillsel.ThisWeapon }, &f.NameCriteria,
		func(parent *unison.Panel, index int) { p.createSecondarySkillPanels(parent, index, f) })
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

// maxAdjustmentBonusSpec describes the parts of a maximum uses / maximum level adjustment bonus panel that differ
// between the equipment and trait variants. Both features carry the same name and tag criteria and the same
// MaxUsesModAmount; only the selection enum and the amount field's label vary.
type maxAdjustmentBonusSpec[E comparable] struct {
	feature     gurps.Feature
	types       []E
	selection   *E
	this        E // The "this item" choice, which needs no criteria at all.
	withName    E // The "items with name" choice, which adds the tag criteria row.
	name        *criteria.Text
	tags        *criteria.Text
	amount      *gurps.MaxUsesModAmount
	amountLabel string
}

func (p *featuresPanel) createEquipmentMaxUsesBonusPanel(f *gurps.EquipmentMaxUsesBonus) (main *unison.Panel, focus unison.Paneler) {
	return createMaxAdjustmentBonusPanel(p, maxAdjustmentBonusSpec[equipmentsel.Type]{
		feature:     f,
		types:       equipmentsel.Types,
		selection:   &f.SelectionType,
		this:        equipmentsel.ThisEquipment,
		withName:    equipmentsel.EquipmentWithName,
		name:        &f.NameCriteria,
		tags:        &f.TagsCriteria,
		amount:      &f.MaxUsesModAmount,
		amountLabel: i18n.Text("Maximum Uses Adjustment"),
	})
}

func (p *featuresPanel) createTraitMaxLevelBonusPanel(f *gurps.TraitMaxLevelBonus) (main *unison.Panel, focus unison.Paneler) {
	return createMaxAdjustmentBonusPanel(p, maxAdjustmentBonusSpec[traitsel.Type]{
		feature:     f,
		types:       traitsel.Types,
		selection:   &f.SelectionType,
		this:        traitsel.ThisTrait,
		withName:    traitsel.TraitWithName,
		name:        &f.NameCriteria,
		tags:        &f.TagsCriteria,
		amount:      &f.MaxUsesModAmount,
		amountLabel: i18n.Text("Maximum Level Adjustment"),
	})
}

// createMaxAdjustmentBonusPanel builds the panel for a maximum uses / maximum level adjustment bonus. It is a plain
// function rather than a method because methods cannot have type parameters.
func createMaxAdjustmentBonusPanel[E comparable](p *featuresPanel, spec maxAdjustmentBonusSpec[E]) (main *unison.Panel, focus unison.Paneler) {
	panel := p.createBasePanel(spec.feature)
	focus = p.addMaxAdjustmentModifierLine(panel, spec.feature, spec.amount, spec.amountLabel)
	addSelectionCriteriaRow(p, panel, spec.types, spec.selection, func(t E) bool { return t == spec.this }, spec.name,
		func(parent *unison.Panel, index int) { createSecondaryMaxAdjustmentPanels(p, parent, index, spec) })
	return panel, focus
}

func (p *featuresPanel) addMaxAdjustmentModifierLine(parent *unison.Panel, f gurps.Feature, amount *gurps.MaxUsesModAmount, label string) *StringField {
	panel := unison.NewPanel()
	p.addTypeSwitcher(panel, f)
	field := NewStringField(nil, "", label,
		func() string { return amount.Amount },
		func(value string) {
			amount.Amount = maxusesmod.Normalize(value)
			MarkModified(panel)
		})
	field.SetMinimumTextWidthUsing("-1,000,000")
	field.Tooltip = newWrappedTooltip(i18n.Text(`Enter a number, percentage or multiplier, e.g. "-1", "10%" or "x2"`))
	panel.AddChild(field)
	addCheckBox(panel, i18n.Text("per level"), &amount.PerLevel)
	addSwitchableCheckBox(panel, f)
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

func createSecondaryMaxAdjustmentPanels[E comparable](p *featuresPanel, parent *unison.Panel, index int, spec maxAdjustmentBonusSpec[E]) {
	if *spec.selection == spec.withName {
		var wrapper *unison.Panel
		wrapper, index = p.prepareNewWrapper(parent, index)
		addTagCriteriaPanel(wrapper, spec.tags, 1, false)
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
	return p.createSpellMatchBonusPanel(f, &f.SpellMatchType, &f.NameCriteria, &f.TagsCriteria, &f.LeveledAmount)
}

func (p *featuresPanel) createSpellPointBonusPanel(f *gurps.SpellPointBonus) (main *unison.Panel, focus unison.Paneler) {
	return p.createSpellMatchBonusPanel(f, &f.SpellMatchType, &f.NameCriteria, &f.TagsCriteria, &f.LeveledAmount)
}

// createSpellMatchBonusPanel builds the panel shared by the spell bonus and the spell point bonus, which pick the
// spells they apply to the same way: by a match type, a name to match against it and the spells' tags.
func (p *featuresPanel) createSpellMatchBonusPanel(f gurps.Feature, matchType *spellmatch.Type, name, tags *criteria.Text, amount *gurps.LeveledAmount) (main *unison.Panel, focus unison.Paneler) {
	panel := p.createBasePanel(f)
	focus = p.addLeveledModifierLine(panel, f, amount)
	addSelectionCriteriaRow(p, panel, spellmatch.Types, matchType,
		func(t spellmatch.Type) bool { return t == spellmatch.AllColleges }, name, nil)
	addTagCriteriaPanel(panel, tags, 1, true)
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
	addSelectionCriteriaRow(p, panel, wsel.Types, &f.SelectionType,
		func(t wsel.Type) bool { return t == wsel.ThisWeapon }, &f.NameCriteria,
		func(parent *unison.Panel, index int) { p.createSecondaryWeaponPanels(parent, index, f) })
	return panel, focus
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
	addSwitchableCheckBox(wrapper, f)
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
	choice := choices[max(min((f.Percentage.AsInteger[int]()/5)-1, 15), 0)]
	pop := addPopup(wrapper, choices, &choice)
	pop.ChoiceMadeCallback = func(popup *unison.PopupMenu[string], index int, _ string) {
		popup.SelectIndex(index)
		f.Percentage = fxp.FromInteger((index + 1) * 5)
		MarkModified(wrapper)
	}
	addSwitchableCheckBox(wrapper, f)
	p.addWrapperAtIndex(panel, wrapper, -1, true)
	return panel, pop
}

func (p *featuresPanel) addLeveledModifierLine(parent *unison.Panel, f gurps.Feature, amount *gurps.LeveledAmount) *DecimalField {
	panel := unison.NewPanel()
	p.addTypeSwitcher(panel, f)
	field, _ := addLeveledAmountPanel(panel, nil, "", i18n.Text("per level"), amount)
	addSwitchableCheckBox(panel, f)
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
		switcher.SetLayoutData(&unison.FlexLayoutData{HSpan: 4})
		if wb.SwitchType == wswitch.NotSwitched {
			wb.SwitchType = wswitch.Types[1]
		}
		spacer := unison.NewPanel()
		spacer.SetLayoutData(&unison.FlexLayoutData{SizeHint: geom.Size{Width: 16}})
		wrapper.AddChild(spacer)
		addPopup(wrapper, wswitch.Types[1:], &wb.SwitchType)
		focus = addBoolPopup(wrapper, i18n.Text("to true"), i18n.Text("to false"), &wb.SwitchTypeValue)
		// The checkbox belongs inside the wrapper, at the end of its second row, so that it sits snugly after the last
		// control as it does on every other feature row. Beside the wrapper instead, it would be top-aligned against a
		// two-row neighbor and pinned to the far right edge, since the wrapper's column absorbs all of the slack.
		addSwitchableCheckBox(wrapper, wb)
		wrapper.SetLayout(&unison.FlexLayout{
			Columns:  4,
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
			// Can't allow the per-die option for MinST bonuses, since that would cause an infinite loop on resolution.
			addCheckBox(panel, i18n.Text("per die"), &wb.PerDie)
		}
		addCheckBox(panel, i18n.Text("as a %"), &wb.Percent)
		addSwitchableCheckBox(panel, wb)
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

func (p *featuresPanel) createSelectorOverridePanel(f *gurps.SelectorOverride) (main *unison.Panel, focus unison.Paneler) {
	panel := p.createBasePanel(f)
	wrapper := unison.NewPanel()
	p.addTypeSwitcher(wrapper, f)
	addSwitchableCheckBox(wrapper, f)
	p.addWrapperAtIndex(panel, wrapper, -1, false)
	panel.AddChild(unison.NewPanel())
	focus = p.addSelectorOverrideLine(panel, f)
	if gurps.SelectorFieldDescriptorFor(f.Field).Scope == gurps.SelectorScopeTrait {
		prefix := i18n.Text("to traits whose name")
		addStringCriteriaPanel(panel, prefix, prefix, i18n.Text("Name Qualifier"), &f.NameCriteria, 1, true)
		addTagCriteriaPanel(panel, &f.TagsCriteria, 1, true)
	} else {
		prefix := i18n.Text("to weapons whose name")
		addStringCriteriaPanel(panel, prefix, prefix, i18n.Text("Name Qualifier"), &f.NameCriteria, 1, true)
		addUsageCriteriaPanel(panel, &f.UsageCriteria, 1, true)
		addTagCriteriaPanel(panel, &f.TagsCriteria, 1, true)
	}
	return panel, focus
}

// addSelectorOverrideLine builds the line "[field] to [value] priority [n]". Changing the field rebuilds the row, since
// a different field may have a different set of valid values (and thus a different value editor).
func (p *featuresPanel) addSelectorOverrideLine(parent *unison.Panel, f *gurps.SelectorOverride) unison.Paneler {
	panel := unison.NewPanel()
	fieldPopup := addPopup(panel, selector.Fields, &f.Field)
	fieldPopup.ChoiceMadeCallback = func(pop *unison.PopupMenu[selector.Field], index int, item selector.Field) {
		pop.SelectIndex(index)
		f.Field = item
		lastSelectorFieldUsed = item
		d := gurps.SelectorFieldDescriptorFor(item)
		if len(d.SuggestedStates) != 0 {
			f.Value = d.SuggestedStates[0]
		} else {
			f.Value = ""
		}
		if d.Scope == gurps.SelectorScopeTrait {
			// The usage row is hidden for trait-scoped fields, so clear any criterion the user set while the field was
			// weapon-scoped rather than leaving an invisible one behind in the data.
			f.UsageCriteria = criteria.Text{Compare: criteria.AnyText}
		}
		p.rebuildFeaturePanel(f)
	}
	panel.AddChild(NewFieldLeadingLabel(i18n.Text("to"), false))
	focus := p.addSelectorValueEditor(panel, f)
	panel.AddChild(NewFieldLeadingLabel(i18n.Text("with priority"), false))
	panel.AddChild(NewIntegerField(nil, "", i18n.Text("Priority"),
		func() int { return f.Priority },
		func(value int) {
			f.Priority = value
			MarkModified(panel)
		}, -99, 99, false, false))
	panel.SetLayout(&unison.FlexLayout{
		Columns:  len(panel.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
	})
	parent.AddChild(panel)
	return focus
}

// addSelectorValueEditor adds the value editor appropriate to the field: a popup when the field is constrained to a
// known set of states, or a free-form string field (with the suggested states offered in a tooltip) otherwise.
func (p *featuresPanel) addSelectorValueEditor(parent *unison.Panel, f *gurps.SelectorOverride) unison.Paneler {
	d := gurps.SelectorFieldDescriptorFor(f.Field)
	if len(d.SuggestedStates) != 0 && !d.FreeForm {
		if !slices.Contains(d.SuggestedStates, f.Value) {
			f.Value = d.SuggestedStates[0]
		}
		// A constrained field shows a popup of human labels while storing the canonical value behind each one.
		popup := unison.NewPopupMenu[string]()
		for _, state := range d.SuggestedStates {
			if d.StateTitle != nil {
				popup.AddItem(d.StateTitle(state))
			} else {
				popup.AddItem(state)
			}
		}
		popup.SelectIndex(slices.Index(d.SuggestedStates, f.Value))
		popup.SelectionChangedCallback = func(pm *unison.PopupMenu[string]) {
			if i := pm.SelectedIndex(); i >= 0 && i < len(d.SuggestedStates) {
				f.Value = d.SuggestedStates[i]
				MarkModified(parent)
			}
		}
		parent.AddChild(popup)
		return popup
	}
	field := NewStringField(nil, "", i18n.Text("Value"),
		func() string { return f.Value },
		func(value string) {
			f.Value = value
			MarkModified(parent)
		})
	// Give the field a minimum width so it can't collapse to nothing as the panel narrows, and let it grab the slack so
	// it flexes rather than being crushed by its neighbors.
	field.SetMinimumTextWidthUsing("impaling")
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	if len(d.SuggestedStates) != 0 {
		field.Tooltip = newWrappedTooltip(fmt.Sprintf(i18n.Text("Suggested values: %s"), strings.Join(d.SuggestedStates, ", ")))
	}
	if d.Validate != nil {
		field.ValidateCallback = func() bool { return d.Validate(field.Text()) }
	}
	parent.AddChild(field)
	return field
}

// createUnknownFeaturePanel creates the panel for a feature this version of GCS doesn't understand. No editing is
// offered, since we have no idea what the data means, but the row is shown so the feature's presence is visible and it
// can be deleted deliberately. No type switcher is present, as switching the type would throw away the original data.
func (p *featuresPanel) createUnknownFeaturePanel(f *gurps.UnknownFeature) (main *unison.Panel, focus unison.Paneler) {
	panel := p.createBasePanel(f)
	label := NewFieldLeadingLabel(fmt.Sprintf(i18n.Text("Unknown feature type %q; it will be preserved, but ignored"),
		f.Kind), false)
	label.Tooltip = newWrappedTooltip(i18n.Text("This was most likely created by a newer version of GCS. Its original data will be written back out unchanged when this file is saved."))
	label.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	panel.AddChild(label)
	return panel, panel
}

// rebuildFeaturePanel replaces the on-screen panel for the given feature in place, keeping the same feature object. It
// is used when an edit (such as changing a selector's field) changes which sub-widgets the row needs.
func (p *featuresPanel) rebuildFeaturePanel(f gurps.Feature) {
	i := slices.IndexFunc(*p.features, func(one gurps.Feature) bool { return one == f })
	if i < 0 {
		return
	}
	p.RemoveChildAtIndex(i + 1) // child 0 is the add button; feature at list index i is at child index i+1
	p.insertFeaturePanel(i+1, f)
	MarkRootAncestorForLayoutRecursively(p)
	MarkModified(p)
}

func (p *featuresPanel) featureTypesList() []feature.Type {
	if e, ok := p.owner.(*gurps.Equipment); ok && e.Container() {
		return feature.SelectableTypes
	}
	return feature.SelectableTypesWithoutContainedWeightReduction
}

// addSwitchableCheckBox adds the checkbox that marks a feature as switchable, i.e. one that only takes effect while the
// owning item's switch is on.
func addSwitchableCheckBox(parent *unison.Panel, f gurps.Feature) *CheckBox {
	checkBox := NewCheckBox(nil, "", i18n.Text("switchable"),
		func() check.Enum { return check.FromBool(f.IsSwitchable()) },
		func(state check.Enum) { f.SetSwitchable(state == check.On) })
	checkBox.Tooltip = newWrappedTooltip(gurps.SwitchableTooltip())
	parent.AddChild(checkBox)
	return checkBox
}

func (p *featuresPanel) addTypeSwitcher(parent *unison.Panel, f gurps.Feature) *unison.PopupMenu[feature.Type] {
	currentType := f.FeatureType()
	popup := addPopup(parent, p.featureTypesList(), &currentType)
	popup.ChoiceMadeCallback = func(pop *unison.PopupMenu[feature.Type], index int, item feature.Type) {
		pop.SelectIndex(index)
		if newFeature := p.createFeatureForType(item); newFeature != nil {
			list := *p.features
			i := slices.IndexFunc(list, func(one gurps.Feature) bool { return one == f })
			if i < 0 {
				return
			}
			lastFeatureTypeUsed = item
			newFeature.SetSwitchable(f.IsSwitchable())
			// Remove the old row by its list position rather than via parent.Parent(): the type switcher isn't nested
			// at a fixed depth for every feature type, so walking up from the switcher could reach the whole features
			// panel and delete the entire section.
			p.RemoveChildAtIndex(i + 1) // child 0 is the add button; feature at list index i is at child index i+1
			list[i] = newFeature
			p.insertFeaturePanel(i+1, newFeature)
			MarkRootAncestorForLayoutRecursively(p)
			MarkModified(p)
		}
	}
	return popup
}

func (p *featuresPanel) createFeatureForType(featureType feature.Type) gurps.Feature {
	if featureType.IsWeaponBonus() {
		bonus := gurps.NewWeaponBonus(featureType)
		bonus.SetOwner(p.owner)
		return bonus
	}
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
	case feature.EquipmentMaxUsesBonus:
		bonus = gurps.NewEquipmentMaxUsesBonus()
	case feature.DRBonus:
		bonus = gurps.NewDRBonus()
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
	case feature.TraitMaxLevelBonus:
		bonus = gurps.NewTraitMaxLevelBonus()
	case feature.SelectorOverride:
		override := gurps.NewSelectorOverride(lastSelectorFieldUsed)
		override.SetOwner(p.owner)
		return override
	default:
		errs.Log(errs.New("unknown feature type"), "type", featureType.Key())
		return nil
	}
	bonus.SetOwner(p.owner)
	return bonus
}
