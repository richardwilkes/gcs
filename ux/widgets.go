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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// Rebuildable defines the methods a rebuildable panel should provide.
type Rebuildable interface {
	unison.Paneler
	fmt.Stringer
	Rebuild(full bool)
}

// Syncer should be called to sync an object's UI state to its model.
type Syncer interface {
	Sync()
}

// DeepSync does a depth-first traversal of the panel and all of its descendents and calls Sync() on any Syncer objects
// it finds.
func DeepSync(panel unison.Paneler) {
	p := panel.AsPanel()
	for _, child := range p.Children() {
		DeepSync(child)
	}
	if syncer, ok := p.Self.(Syncer); ok {
		syncer.Sync()
	}
}

// ModifiableRoot marks the root of a modifable tree of components, typically a Dockable.
type ModifiableRoot interface {
	MarkModified(src unison.Paneler)
}

// MarkModified looks for a ModifiableRoot, starting at the panel. If found, it then called MarkModified() on it.
func MarkModified(panel unison.Paneler) {
	gurps.DiscardGlobalResolveCache()
	p := panel.AsPanel()
	for p != nil {
		if modifiable, ok := p.Self.(ModifiableRoot); ok {
			modifiable.MarkModified(panel)
			break
		}
		p = p.Parent()
	}
}

func addSourceFields(parent *unison.Panel, source *gurps.SourcedID) {
	parent.AddChild(NewFieldLeadingLabel(i18n.Text("ID"), false))
	parent.AddChild(NewNonEditableField(func(f *NonEditableField) {
		f.SetTitle(string(source.TID))
	}))
	parent.AddChild(NewFieldLeadingLabel(i18n.Text("Source ID"), false))
	parent.AddChild(NewNonEditableField(func(f *NonEditableField) {
		f.SetTitle(string(source.Source.TID))
	}))
	parent.AddChild(NewFieldLeadingLabel(i18n.Text("Source Library"), false))
	parent.AddChild(NewNonEditableField(func(f *NonEditableField) {
		f.SetTitle(source.Source.Library)
	}))
	parent.AddChild(NewFieldLeadingLabel(i18n.Text("Source Path"), false))
	parent.AddChild(NewNonEditableField(func(f *NonEditableField) {
		f.SetTitle(source.Source.Path)
	}))
}

func addNameLabelAndField(parent *unison.Panel, fieldData *string) {
	addLabelAndStringField(parent, i18n.Text("Name"), "", fieldData)
}

func addSpecializationLabelAndField(parent *unison.Panel, fieldData *string) {
	addLabelAndStringField(parent, i18n.Text("Specialization"), "", fieldData)
}

func addPageRefLabelAndField(parent *unison.Panel, fieldData *string) {
	addLabelAndStringField(parent, i18n.Text("Page Reference"), gurps.PageRefTooltip(), fieldData)
}

func addPageRefHighlightLabelAndField(parent *unison.Panel, fieldData *string) {
	addLabelAndStringField(parent, i18n.Text("Page Highlight"),
		i18n.Text(`A snippet of text to highlight on the page when opening the page reference; only needed if the default behavior isn't highlighting the expected area`),
		fieldData)
}

func addNotesLabelAndField(parent *unison.Panel, fieldData *string) {
	labelText := i18n.Text("Notes")
	label := NewFieldLeadingLabel(labelText, false)
	parent.AddChild(label)
	addScriptField(parent, nil, "", labelText,
		i18n.Text("These notes may have scripts embedded in them by wrapping each script in <script>your script goes here</script> tags."),
		func() string { return *fieldData },
		func(value string) {
			*fieldData = value
			parent.MarkForLayoutAndRedraw()
			MarkModified(parent)
		})
}

func addVTTNotesLabelAndField(parent *unison.Panel, fieldData *string) {
	addLabelAndMultiLineStringField(parent, i18n.Text("VTT Notes"),
		i18n.Text("Any notes for VTT use; see the instructions for your VTT to determine if/how these can be used"),
		fieldData)
}

func addUserDescLabelAndField(parent *unison.Panel, fieldData *string) {
	addLabelAndMultiLineStringField(parent, i18n.Text("User Description"),
		i18n.Text("Additional notes for your own reference. These only exist in character sheets and will be removed if transferred to a data list or template"),
		fieldData)
}

func addTechLevelRequired(parent *unison.Panel, fieldData **string, ownerIsSheet bool) {
	tl := i18n.Text("Tech Level")
	var field *StringField
	wrapper := addFlowWrapper(parent, tl, 2)
	field = NewStringField(nil, "", tl, func() string {
		if *fieldData == nil {
			return ""
		}
		return **fieldData
	}, func(value string) {
		if *fieldData == nil {
			return
		}
		**fieldData = value
		MarkModified(parent)
	})
	tip := gurps.TechLevelInfo()
	if !ownerIsSheet {
		tip = txt.Wrap("", i18n.Text("Leave field blank to auto-populate with the character's TL when added to a character sheet."), 60) + "\n\n" + tip
	}
	field.Tooltip = newWrappedTooltip(tip)
	if *fieldData == nil {
		field.SetEnabled(false)
	}
	field.SetMinimumTextWidthUsing("12^")
	wrapper.AddChild(field)
	parent = wrapper
	last := *fieldData
	required := last != nil
	parent.AddChild(NewCheckBox(nil, "", i18n.Text("Required"),
		func() check.Enum { return check.FromBool(required) },
		func(state check.Enum) {
			if required = state == check.On; required {
				if last == nil {
					var data string
					last = &data
				}
				*fieldData = last
				if field != nil {
					field.SetEnabled(true)
				}
			} else {
				last = *fieldData
				*fieldData = nil
				if field != nil {
					field.SetEnabled(false)
				}
			}
		}))
}

func addAttributeChoicePopup(parent *unison.Panel, entity *gurps.Entity, prefix string, fieldData *string, flags gurps.AttributeFlags) *unison.PopupMenu[*gurps.AttributeChoice] {
	choices, current := gurps.AttributeChoices(entity, prefix, flags, *fieldData)
	popup := addPopup(parent, choices, &current)
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[*gurps.AttributeChoice]) {
		if choice, ok := p.Selected(); ok {
			*fieldData = choice.Key
			MarkModified(parent)
		}
	}
	return popup
}

func addDifficultyLabelAndFields(parent *unison.Panel, entity *gurps.Entity, attrDiff *gurps.AttributeDifficulty) {
	wrapper := addFlowWrapper(parent, i18n.Text("Difficulty"), 3)
	addAttributeChoicePopup(wrapper, entity, "", &attrDiff.Attribute, gurps.TenFlag)
	wrapper.AddChild(NewFieldTrailingLabel("/", false))
	addPopup(wrapper, difficulty.Levels, &attrDiff.Difficulty)
}

func addTagsLabelAndField(parent *unison.Panel, fieldData *[]string) {
	addLabelAndListField(parent, i18n.Text("Tags"), i18n.Text("tags"), fieldData)
}

func addLabelAndListField(parent *unison.Panel, labelText, pluralForTooltip string, fieldData *[]string) {
	tooltip := fmt.Sprintf(i18n.Text("Separate multiple %s with commas"), pluralForTooltip)
	label := NewFieldLeadingLabel(labelText, false)
	if tooltip != "" {
		label.Tooltip = newWrappedTooltip(tooltip)
	}
	parent.AddChild(label)
	field := NewMultiLineStringField(nil, "", labelText,
		func() string { return gurps.CombineTags(*fieldData) },
		func(value string) {
			*fieldData = gurps.ExtractTags(value)
			parent.MarkForLayoutAndRedraw()
			MarkModified(parent)
		})
	if tooltip != "" {
		field.Tooltip = newWrappedTooltip(tooltip)
	}
	field.AutoScroll = false
	parent.AddChild(field)
}

func addLabelAndStringField(parent *unison.Panel, labelText, tooltip string, fieldData *string) *StringField {
	label := NewFieldLeadingLabel(labelText, false)
	if tooltip != "" {
		label.Tooltip = newWrappedTooltip(tooltip)
	}
	parent.AddChild(label)
	return addStringField(parent, labelText, tooltip, fieldData)
}

func newWrappedTooltip(tooltip string) *unison.Panel {
	return unison.NewTooltipWithText(wrapTextForTooltip(tooltip))
}

func newWrappedTooltipWithSecondaryText(primary, secondary string) *unison.Panel {
	return unison.NewTooltipWithSecondaryText(wrapTextForTooltip(primary), wrapTextForTooltip(secondary))
}

func wrapTextForTooltip(tooltip string) string {
	return strings.ReplaceAll(txt.Wrap("", strings.ReplaceAll(tooltip, " ", "␣"), 80), "␣", " ")
}

func addStringField(parent *unison.Panel, labelText, tooltip string, fieldData *string) *StringField {
	field := NewStringField(nil, "", labelText,
		func() string { return *fieldData },
		func(value string) {
			*fieldData = value
			MarkModified(parent)
		})
	if tooltip != "" {
		field.Tooltip = newWrappedTooltip(tooltip)
	}
	parent.AddChild(field)
	return field
}

func addLabelAndMultiLineStringField(parent *unison.Panel, labelText, tooltip string, fieldData *string) {
	label := NewFieldLeadingLabel(labelText, false)
	if tooltip != "" {
		label.Tooltip = newWrappedTooltip(tooltip)
	}
	parent.AddChild(label)
	field := NewMultiLineStringField(nil, "", labelText,
		func() string { return *fieldData },
		func(value string) {
			*fieldData = value
			parent.MarkForLayoutAndRedraw()
			MarkModified(parent)
		})
	if tooltip != "" {
		field.Tooltip = newWrappedTooltip(tooltip)
	}
	field.AutoScroll = false
	parent.AddChild(field)
}

func addIntegerField(parent *unison.Panel, targetMgr *TargetMgr, targetKey, labelText, tooltip string, fieldData *int, minValue, maxValue int) *IntegerField {
	field := NewIntegerField(targetMgr, targetKey, labelText,
		func() int { return *fieldData },
		func(value int) {
			*fieldData = value
			MarkModified(parent)
		}, minValue, maxValue, false, false)
	if tooltip != "" {
		field.Tooltip = newWrappedTooltip(tooltip)
	}
	parent.AddChild(field)
	return field
}

func addLabelAndDecimalField(parent *unison.Panel, targetMgr *TargetMgr, targetKey, labelText, tooltip string, fieldData *fxp.Int, minValue, maxValue fxp.Int) *DecimalField {
	label := NewFieldLeadingLabel(labelText, false)
	if tooltip != "" {
		label.Tooltip = newWrappedTooltip(tooltip)
	}
	parent.AddChild(label)
	return addDecimalField(parent, targetMgr, targetKey, labelText, tooltip, fieldData, minValue, maxValue)
}

func addDecimalField(parent *unison.Panel, targetMgr *TargetMgr, targetKey, labelText, tooltip string, fieldData *fxp.Int, minValue, maxValue fxp.Int) *DecimalField {
	field := NewDecimalField(targetMgr, targetKey, labelText,
		func() fxp.Int { return *fieldData },
		func(value fxp.Int) {
			*fieldData = value
			MarkModified(parent)
		}, minValue, maxValue, false, false)
	if tooltip != "" {
		field.Tooltip = newWrappedTooltip(tooltip)
	}
	parent.AddChild(field)
	return field
}

func addDecimalFieldWithSign(parent *unison.Panel, targetMgr *TargetMgr, targetKey, labelText, tooltip string, fieldData *fxp.Int, minValue, maxValue fxp.Int) *DecimalField {
	field := NewDecimalField(targetMgr, targetKey, labelText,
		func() fxp.Int { return *fieldData },
		func(value fxp.Int) {
			*fieldData = value
			MarkModified(parent)
		}, minValue, maxValue, true, false)
	if tooltip != "" {
		field.Tooltip = newWrappedTooltip(tooltip)
	}
	parent.AddChild(field)
	return field
}

func addWeightField(parent *unison.Panel, targetMgr *TargetMgr, targetKey, labelText, tooltip string, entity *gurps.Entity, fieldData *fxp.Weight, noMinWidth bool) *WeightField {
	field := NewWeightField(targetMgr, targetKey, labelText, entity,
		func() fxp.Weight { return *fieldData },
		func(value fxp.Weight) {
			*fieldData = value
			MarkModified(parent)
		}, 0, fxp.Weight(fxp.Max), noMinWidth)
	if tooltip != "" {
		field.Tooltip = newWrappedTooltip(tooltip)
	}
	parent.AddChild(field)
	return field
}

func addCheckBox(parent *unison.Panel, labelText string, fieldData *bool) *CheckBox {
	checkBox := NewCheckBox(nil, "", labelText,
		func() check.Enum { return check.FromBool(*fieldData) },
		func(state check.Enum) { *fieldData = state == check.On })
	parent.AddChild(checkBox)
	return checkBox
}

func addInvertedCheckBox(parent *unison.Panel, labelText string, fieldData *bool) *CheckBox {
	checkBox := NewCheckBox(nil, "", labelText,
		func() check.Enum { return check.FromBool(!*fieldData) },
		func(state check.Enum) { *fieldData = state == check.Off })
	parent.AddChild(checkBox)
	return checkBox
}

func addFlowWrapper(parent *unison.Panel, labelText string, count int) *unison.Panel {
	parent.AddChild(NewFieldLeadingLabel(labelText, false))
	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  count,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		VAlign:   align.Middle,
	})
	parent.AddChild(wrapper)
	return wrapper
}

func addFillWrapper(parent *unison.Panel, labelText string, count int) *unison.Panel {
	wrapper := addFlowWrapper(parent, labelText, count)
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	return wrapper
}

func addNullableDice(parent *unison.Panel, labelText, tooltip string, fieldData **dice.Dice, useSign bool) *StringField {
	var data string
	if *fieldData != nil {
		data = (*fieldData).String()
	}
	field := NewStringField(nil, "", labelText,
		func() string {
			if useSign && data != "" && !strings.HasPrefix(data, "+") && !strings.HasPrefix(data, "-") {
				return "+" + data
			}
			return data
		},
		func(value string) {
			data = strings.TrimPrefix(strings.TrimSpace(value), "+")
			if value == "" {
				*fieldData = nil
			} else {
				*fieldData = dice.New(data)
			}
			MarkModified(parent)
		})
	if tooltip != "" {
		field.Tooltip = newWrappedTooltip(tooltip)
	}
	parent.AddChild(field)
	return field
}

func addLabelAndPopup[T comparable](parent *unison.Panel, labelText, tooltip string, choices []T, fieldData *T) *unison.PopupMenu[T] {
	label := NewFieldLeadingLabel(labelText, false)
	if tooltip != "" {
		label.Tooltip = newWrappedTooltip(tooltip)
	}
	parent.AddChild(label)
	return addPopup(parent, choices, fieldData)
}

func addPopup[T comparable](parent *unison.Panel, choices []T, fieldData *T) *unison.PopupMenu[T] {
	popup := unison.NewPopupMenu[T]()
	for _, one := range choices {
		popup.AddItem(one)
	}
	popup.Select(*fieldData)
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[T]) {
		if item, ok := p.Selected(); ok {
			*fieldData = item
			MarkModified(parent)
		}
	}
	parent.AddChild(popup)
	return popup
}

func addBoolPopup(parent *unison.Panel, trueChoice, falseChoice string, fieldData *bool) *unison.PopupMenu[string] {
	popup := unison.NewPopupMenu[string]()
	popup.AddItem(trueChoice)
	popup.AddItem(falseChoice)
	if *fieldData {
		popup.SelectIndex(0)
	} else {
		popup.SelectIndex(1)
	}
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[string]) {
		*fieldData = p.SelectedIndex() == 0
		MarkModified(parent)
	}
	parent.AddChild(popup)
	return popup
}

func addHasPopup(parent *unison.Panel, has *bool) {
	addBoolPopup(parent, i18n.Text("has"), i18n.Text("doesn't have"), has)
}

func adjustFieldBlank(field unison.Paneler, blank bool) {
	panel := field.AsPanel()
	panel.SetEnabled(!blank)
	if blank {
		panel.DrawOverCallback = func(gc *unison.Canvas, _ unison.Rect) {
			var ink unison.Ink
			if f, ok := panel.Self.(*unison.Field); ok {
				ink = f.BackgroundInk
			} else {
				ink = unison.DefaultFieldTheme.BackgroundInk
			}
			r := panel.ContentRect(false)
			gc.DrawRect(r, ink.Paint(gc, r, paintstyle.Fill))
		}
	} else {
		panel.DrawOverCallback = nil
	}
}

func adjustPopupBlank[T comparable](popup *unison.PopupMenu[T], blank bool) {
	popup.SetEnabled(!blank)
	if blank {
		popup.DrawOverCallback = func(gc *unison.Canvas, _ unison.Rect) {
			unison.DrawRoundedRectBase(gc, popup.ContentRect(false), popup.CornerRadius, 1, popup.BackgroundInk, popup.EdgeInk)
		}
	} else {
		popup.DrawOverCallback = nil
	}
}

func addNameCriteriaPanel(parent *unison.Panel, strCriteria *criteria.Text, hSpan int, includeEmptyFiller bool) {
	prefix := i18n.Text("whose name")
	addStringCriteriaPanel(parent, prefix, prefix, i18n.Text("Name Qualifier"), strCriteria, hSpan, includeEmptyFiller)
}

func addSpecializationCriteriaPanel(parent *unison.Panel, strCriteria *criteria.Text, hSpan int, includeEmptyFiller bool) {
	prefix := i18n.Text("and whose specialization")
	addStringCriteriaPanel(parent, prefix, prefix, i18n.Text("Specialization Qualifier"), strCriteria, hSpan, includeEmptyFiller)
}

func addUsageCriteriaPanel(parent *unison.Panel, strCriteria *criteria.Text, hSpan int, includeEmptyFiller bool) {
	prefix := i18n.Text("and whose usage")
	addStringCriteriaPanel(parent, prefix, prefix, i18n.Text("Usage Qualifier"), strCriteria, hSpan, includeEmptyFiller)
}

func addTagCriteriaPanel(parent *unison.Panel, strCriteria *criteria.Text, hSpan int, includeEmptyFiller bool) {
	addStringCriteriaPanel(parent, i18n.Text("and at least one tag"), i18n.Text("and all tags"), i18n.Text("Tag Qualifier"), strCriteria, hSpan, includeEmptyFiller)
}

func addNotesCriteriaPanel(parent *unison.Panel, strCriteria *criteria.Text, hSpan int, includeEmptyFiller bool) {
	prefix := i18n.Text("and whose notes")
	addStringCriteriaPanel(parent, prefix, prefix, i18n.Text("Notes Qualifier"), strCriteria, hSpan, includeEmptyFiller)
}

func addStringCriteriaPanel(parent *unison.Panel, prefix, notPrefix, undoTitle string, strCriteria *criteria.Text, hSpan int, includeEmptyFiller bool) (*unison.PopupMenu[string], *StringField) {
	if includeEmptyFiller {
		parent.AddChild(unison.NewPanel())
	}
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		VAlign:   align.Middle,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  hSpan,
		HAlign: align.Fill,
		HGrab:  true,
	})
	var criteriaField *StringField
	popup := unison.NewPopupMenu[string]()
	for _, one := range criteria.PrefixedStringComparisonChoices(prefix, notPrefix) {
		popup.AddItem(one)
	}
	popup.SelectIndex(criteria.ExtractStringComparisonIndex(string(strCriteria.Compare)))
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[string]) {
		strCriteria.Compare = criteria.AllStringComparisons[p.SelectedIndex()]
		adjustFieldBlank(criteriaField, strCriteria.Compare == criteria.AnyText)
		MarkModified(panel)
	}
	panel.AddChild(popup)
	criteriaField = addStringField(panel, undoTitle, "", &strCriteria.Qualifier)
	adjustFieldBlank(criteriaField, strCriteria.Compare == criteria.AnyText)
	parent.AddChild(panel)
	return popup, criteriaField
}

func addLevelCriteriaPanel(parent *unison.Panel, targetMgr *TargetMgr, targetKey string, numCriteria *criteria.Number, hSpan int, includeEmptyFiller bool) {
	addNumericCriteriaPanel(parent, targetMgr, targetKey, i18n.Text("and whose level"), i18n.Text("Level Qualifier"),
		numCriteria, 0, fxp.Thousand, hSpan, false, includeEmptyFiller)
}

func addNumericCriteriaPanel(parent *unison.Panel, targetMgr *TargetMgr, targetKey, prefix, undoTitle string, numCriteria *criteria.Number, minValue, maxValue fxp.Int, hSpan int, integerOnly, includeEmptyFiller bool) (popup *unison.PopupMenu[string], field unison.Paneler) {
	if includeEmptyFiller {
		parent.AddChild(unison.NewPanel())
	}
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		VAlign:   align.Middle,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  hSpan,
		HAlign: align.Fill,
		HGrab:  true,
	})
	popup = unison.NewPopupMenu[string]()
	for _, one := range criteria.PrefixedNumericComparisonChoices(prefix) {
		popup.AddItem(one)
	}
	popup.SelectIndex(criteria.ExtractNumericComparisonIndex(string(numCriteria.Compare)))
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[string]) {
		numCriteria.Compare = criteria.AllNumericComparisons[p.SelectedIndex()]
		adjustFieldBlank(field, numCriteria.Compare == criteria.AnyNumber)
		MarkModified(panel)
	}
	panel.AddChild(popup)
	if integerOnly {
		field = NewIntegerField(targetMgr, targetKey, undoTitle,
			func() int { return fxp.As[int](numCriteria.Qualifier) },
			func(value int) {
				numCriteria.Qualifier = fxp.From(value)
				MarkModified(panel)
			}, fxp.As[int](minValue), fxp.As[int](maxValue), false, false)
		panel.AddChild(field)
	} else {
		field = addDecimalField(panel, targetMgr, targetKey, undoTitle, "", &numCriteria.Qualifier, minValue, maxValue)
	}
	adjustFieldBlank(field, numCriteria.Compare == criteria.AnyNumber)
	parent.AddChild(panel)
	return popup, field
}

func addWeightCriteriaPanel(parent *unison.Panel, targetMgr *TargetMgr, targetKey string, entity *gurps.Entity, weightCriteria *criteria.Weight) {
	popup := unison.NewPopupMenu[string]()
	for _, one := range criteria.PrefixedNumericComparisonChoices(i18n.Text("which")) {
		popup.AddItem(one)
	}
	popup.SelectIndex(criteria.ExtractNumericComparisonIndex(string(weightCriteria.Compare)))
	parent.AddChild(popup)
	field := addWeightField(parent, targetMgr, targetKey, i18n.Text("Weight Qualifier"), "", entity,
		&weightCriteria.Qualifier, false)
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[string]) {
		weightCriteria.Compare = criteria.AllNumericComparisons[p.SelectedIndex()]
		adjustFieldBlank(field, weightCriteria.Compare == criteria.AnyNumber)
		MarkModified(parent)
	}
	adjustFieldBlank(field, weightCriteria.Compare == criteria.AnyNumber)
	parent.SetLayout(&unison.FlexLayout{
		Columns:  len(parent.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
}

func addQuantityCriteriaPanel(parent *unison.Panel, targetMgr *TargetMgr, targetKey string, numCriteria *criteria.Number) {
	choices := []string{
		i18n.Text("exactly"),
		i18n.Text("at least"),
		i18n.Text("at most"),
	}
	var numType string
	switch numCriteria.Compare {
	case criteria.AtLeastNumber:
		numType = choices[1]
	case criteria.AtMostNumber:
		numType = choices[2]
	default:
		numType = choices[0]
	}
	popup := unison.NewPopupMenu[string]()
	for _, one := range choices {
		popup.AddItem(one)
	}
	popup.Select(numType)
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[string]) {
		switch p.SelectedIndex() {
		case 0:
			numCriteria.Compare = criteria.EqualsNumber
		case 1:
			numCriteria.Compare = criteria.AtLeastNumber
		case 2:
			numCriteria.Compare = criteria.AtMostNumber
		}
		MarkModified(parent)
	}
	parent.AddChild(popup)
	parent.AddChild(NewIntegerField(targetMgr, targetKey, i18n.Text("Quantity Criteria"),
		func() int { return fxp.As[int](numCriteria.Qualifier) },
		func(value int) {
			numCriteria.Qualifier = fxp.From(value)
			MarkModified(parent)
		}, 0, 9999, false, false))
}

func addLeveledAmountPanel(parent *unison.Panel, targetMgr *TargetMgr, targetKey, title string, amount *gurps.LeveledAmount) {
	parent.AddChild(NewDecimalField(targetMgr, targetKey, i18n.Text("Amount"),
		func() fxp.Int { return amount.Amount },
		func(value fxp.Int) {
			amount.Amount = value
			MarkModified(parent)
		}, fxp.Min, fxp.Max, true, false))
	addCheckBox(parent, title, &amount.PerLevel)
}

func addTemplateChoices(parent *unison.Panel, targetmgr *TargetMgr, targetKey string, tp **gurps.TemplatePicker) {
	if *tp == nil {
		*tp = &gurps.TemplatePicker{}
	}
	last := (*tp).Type
	wrapper := addFlowWrapper(parent, i18n.Text("Template Choices"), 3)
	templatePickerTypePopup := addPopup(wrapper, picker.Types, &(*tp).Type)
	text := i18n.Text("Template Choice Quantifier")
	popup, field := addNumericCriteriaPanel(wrapper, targetmgr, targetKey, "", text, &(*tp).Qualifier, fxp.Min,
		fxp.Max, 1, false, false)
	templatePickerTypePopup.SelectionChangedCallback = func(p *unison.PopupMenu[picker.Type]) {
		if item, ok := p.Selected(); ok {
			(*tp).Type = item
			if last == picker.NotApplicable && item != picker.NotApplicable {
				(*tp).Qualifier.Qualifier = fxp.One
				if syncer, ok2 := field.(Syncer); ok2 {
					syncer.Sync()
				}
			}
			last = item
			adjustFieldBlank(field, item == picker.NotApplicable || (*tp).Qualifier.Compare == criteria.AnyNumber)
			adjustPopupBlank(popup, item == picker.NotApplicable)
			MarkModified(parent)
		}
	}
	adjustFieldBlank(field, (*tp).Type == picker.NotApplicable)
}

func addScriptField(parent *unison.Panel, targetMgr *TargetMgr, targetKey, undoTitle, tooltip string, get func() string, set func(string)) *StringField {
	field := NewMultiLineStringField(targetMgr, targetKey, undoTitle, get, set)
	field.AutoScroll = false
	field.SetMinimumTextWidthUsing("floor($basic_speed)")
	field.Tooltip = newWrappedTooltip(tooltip)
	scriptHelpButton := unison.NewSVGButton(svg.Script)
	scriptHelpButton.ClickCallback = func() { HandleLink(nil, "md:Help/Scripting Guide") }
	scriptHelpButton.Tooltip = newWrappedTooltip(i18n.Text("Scripting Guide"))
	parent.AddChild(WrapWithSpan(1, field, scriptHelpButton))
	return field
}

// WrapWithSpan wraps a number of children with a single panel that request to fill in span number of columns.
func WrapWithSpan(span int, children ...unison.Paneler) *unison.Panel {
	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  len(children),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  span,
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	for _, child := range children {
		wrapper.AddChild(child)
	}
	return wrapper
}

// NewSVGButtonForFont creates a new SVG button with the given font and a size adjustment.
func NewSVGButtonForFont(svgData *unison.SVG, font unison.Font, sizeAdjust float32) *unison.Button {
	b := unison.NewButton()
	b.ButtonTheme = unison.DefaultButtonTheme
	b.Font = font
	b.DrawableOnlyVMargin = 1
	b.DrawableOnlyHMargin = 1
	b.HideBase = true
	baseline := font.Baseline() + sizeAdjust
	b.Drawable = &unison.DrawableSVG{
		SVG:  svgData,
		Size: unison.NewSize(baseline, baseline).Ceil(),
	}
	return b
}
