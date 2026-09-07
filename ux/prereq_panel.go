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
	"reflect"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellcmp"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

const (
	noAndOr = ""
	// samePowerSourceIndex is the position of the "is the same as this spell's" choice within the power source popup
	// menu, just after "is anything". The choice is only present when the prerequisite belongs to a spell.
	samePowerSourceIndex = 1
)

var lastPrereqTypeUsed = prereq.Trait

type prereqPanel struct {
	unison.Panel
	entity           *gurps.Entity
	root             **gurps.PrereqList
	permittedChoices []prereq.Type
	andOrMap         map[gurps.Prereq]*unison.Label
	ownerIsSpell     bool
}

func newPrereqPanel(entity *gurps.Entity, root **gurps.PrereqList, permittedChoices []prereq.Type, ownerIsSpell bool) *prereqPanel {
	p := &prereqPanel{
		entity:           entity,
		root:             root,
		permittedChoices: permittedChoices,
		andOrMap:         make(map[gurps.Prereq]*unison.Label),
		ownerIsSpell:     ownerIsSpell,
	}
	initTitledEditorSection(p, i18n.Text("Prerequisites"))
	list, _ := p.createPrereqListPanel(0, *root)
	p.AddChild(list)
	return p
}

func (p *prereqPanel) createPrereqListPanel(depth int, list *gurps.PrereqList) (main, focus unison.Paneler) {
	row := p.beginPrereqRow(depth, list, nil)
	_, focus = addNumericCriteriaPanel(row.panel, nil, "", i18n.Text("When the Tech Level"), i18n.Text("When Tech Level"),
		&list.WhenTL, 0, fxp.Twelve, 1, true, true)
	popup := addBoolPopup(row.panel, i18n.Text("requires all of:"), i18n.Text("requires at least one of:"), &list.All)
	callback := popup.SelectionChangedCallback
	popup.SelectionChangedCallback = func(pop *unison.PopupMenu[string]) {
		callback(pop)
		p.adjustAndOrForList(list)
	}
	row.finish()
	row.panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	for _, child := range list.Prereqs {
		p.addToList(row.panel, depth+1, -1, child)
	}
	return row.panel, focus
}

func (p *prereqPanel) addToList(parent *unison.Panel, depth, index int, child gurps.Prereq) {
	var panel, focus unison.Paneler
	switch one := child.(type) {
	case *gurps.PrereqList:
		panel, focus = p.createPrereqListPanel(depth, one)
	case *gurps.TraitPrereq:
		panel, focus = p.createTraitPrereqPanel(depth, one)
	case *gurps.AttributePrereq:
		panel, focus = p.createAttributePrereqPanel(depth, one)
	case *gurps.ContainedQuantityPrereq:
		panel, focus = p.createContainedQuantityPrereqPanel(depth, one)
	case *gurps.ContainedWeightPrereq:
		panel, focus = p.createContainedWeightPrereqPanel(depth, one)
	case *gurps.EquippedEquipmentPrereq:
		panel, focus = p.createEquippedEquipmentPrereqPanel(depth, one)
	case *gurps.SkillPrereq:
		panel, focus = p.createSkillPrereqPanel(depth, one)
	case *gurps.SpellPrereq:
		panel, focus = p.createSpellPrereqPanel(depth, one)
	case *gurps.ScriptPrereq:
		panel, focus = p.createScriptPrereqPanel(depth, one)
	case *gurps.UnknownPrereq:
		panel, focus = p.createUnknownPrereqPanel(depth, one)
	default:
		errs.Log(errs.New("unknown prerequisite type"), "type", reflect.TypeOf(child).String())
	}
	if panel != nil {
		columns := 1
		if parentLayout, ok := parent.Layout().(*unison.FlexLayout); ok {
			columns = parentLayout.Columns
		}
		panel.AsPanel().SetLayoutData(&unison.FlexLayoutData{
			HSpan:  columns,
			HAlign: align.Fill,
			HGrab:  true,
		})
		if index < 0 {
			parent.AddChild(panel)
		} else {
			parent.AddChildAtIndex(panel, columns+index)
		}
		focus.AsPanel().RequestFocus()
	}
}

// createUnknownPrereqPanel creates the panel for a prerequisite this version of GCS doesn't understand. No editing is
// offered, since we have no idea what the data means, but the row is shown so that the presence of the prerequisite is
// visible and it can be deleted deliberately. Note that no type switcher is present, as switching the type would throw
// away the original data.
func (p *prereqPanel) createUnknownPrereqPanel(depth int, pr *gurps.UnknownPrereq) (main, focus unison.Paneler) {
	row := p.beginPrereqRow(depth, pr, nil)
	label := NewFieldLeadingLabel(fmt.Sprintf(i18n.Text("Unknown prerequisite type %q; it will be preserved, but is never satisfied"),
		pr.Kind), false)
	label.Tooltip = newWrappedTooltip(i18n.Text("This was most likely created by a newer version of GCS. Its original data will be written back out unchanged when this file is saved."))
	row.panel.AddChild(label)
	row.finish()
	return row.panel, row.panel
}

func (p *prereqPanel) createButtonsPanel(parent *unison.Panel, depth int, data gurps.Prereq) {
	buttons := unison.NewPanel()
	buttons.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: float32(depth * 20)}))
	parent.AddChild(buttons)
	if prereqList, ok := data.(*gurps.PrereqList); ok {
		addPrereqButton := unison.NewSVGButton(unison.CircledAddSVG)
		addPrereqButton.ClickCallback = func() {
			if created := p.createPrereqForType(lastPrereqTypeUsed, prereqList); created != nil {
				prereqList.Prereqs = slices.Insert(prereqList.Prereqs, 0, created)
				p.addToList(parent, depth+1, 0, created)
				p.adjustAndOrForList(prereqList)
				MarkRootAncestorForLayoutRecursively(p)
				MarkModified(p)
			}
		}
		buttons.AddChild(addPrereqButton)

		addPrereqListButton := unison.NewSVGButton(svg.CircledVerticalEllipsis)
		addPrereqListButton.ClickCallback = func() {
			newList := gurps.NewPrereqList()
			newList.Parent = prereqList
			prereqList.Prereqs = slices.Insert(prereqList.Prereqs, 0, gurps.Prereq(newList))
			p.addToList(parent, depth+1, 0, newList)
			p.adjustAndOrForList(prereqList)
			MarkRootAncestorForLayoutRecursively(p)
			MarkModified(p)
		}
		buttons.AddChild(addPrereqListButton)
	}
	parentList := data.ParentList()
	if parentList != nil {
		deleteButton := unison.NewSVGButton(unison.TrashSVG)
		deleteButton.ClickCallback = func() {
			delete(p.andOrMap, data)
			if i := slices.IndexFunc(parentList.Prereqs, func(elem gurps.Prereq) bool { return elem == data }); i != -1 {
				parentList.Prereqs = slices.Delete(parentList.Prereqs, i, i+1)
			}
			parent.RemoveFromParent()
			p.adjustAndOrForList(parentList)
			MarkRootAncestorForLayoutRecursively(p)
			MarkModified(p)
		}
		buttons.AddChild(deleteButton)
	}
	buttons.SetLayout(&unison.FlexLayout{
		Columns: len(buttons.Children()),
	})
}

func (p *prereqPanel) addAndOr(parent *unison.Panel, data gurps.Prereq) {
	label := NewFieldLeadingLabel(andOrText(data), false)
	parent.AddChild(label)
	p.andOrMap[data] = label
}

func (p *prereqPanel) adjustAndOrForList(list *gurps.PrereqList) {
	for _, one := range list.Prereqs {
		p.adjustAndOr(one)
	}
	p.MarkForLayoutRecursively()
}

func (p *prereqPanel) adjustAndOr(data gurps.Prereq) {
	if label, ok := p.andOrMap[data]; ok {
		if text := andOrText(data); text != label.Text.String() {
			parent := label.Parent()
			label.RemoveFromParent()
			label.SetTitle(text)
			i := 1
			if text == noAndOr {
				if parentLayout, ok2 := parent.Layout().(*unison.FlexLayout); ok2 {
					i = parentLayout.Columns - 1
				}
			}
			parent.AddChildAtIndex(label, i)
		}
	}
}

func andOrText(pr gurps.Prereq) string {
	list := pr.ParentList()
	if list == nil || len(list.Prereqs) < 2 || list.Prereqs[0] == pr {
		return noAndOr
	}
	if list.All {
		return i18n.Text("and")
	}
	return i18n.Text("or")
}

func (p *prereqPanel) addPrereqTypeSwitcher(parent *unison.Panel, depth int, pr gurps.Prereq) {
	prereqType := pr.PrereqType()
	popup := addPopup(parent, p.permittedChoices, &prereqType)
	popup.SelectionChangedCallback = func(pop *unison.PopupMenu[prereq.Type]) {
		if item, ok := pop.Selected(); ok {
			parentList := pr.ParentList()
			if newPrereq := p.createPrereqForType(item, parentList); newPrereq != nil {
				lastPrereqTypeUsed = item
				parentOfParent := parent.Parent()
				parent.RemoveFromParent()
				list := parentList.Prereqs
				i := slices.IndexFunc(list, func(one gurps.Prereq) bool { return one == pr })
				list[i] = newPrereq
				p.addToList(parentOfParent, depth, i, newPrereq)
				MarkRootAncestorForLayoutRecursively(p)
				MarkModified(p)
			}
		}
	}
}

func (p *prereqPanel) createPrereqForType(prereqType prereq.Type, parentList *gurps.PrereqList) gurps.Prereq {
	switch prereqType {
	case prereq.List:
		one := gurps.NewPrereqList()
		one.Parent = parentList
		return one
	case prereq.Trait:
		one := gurps.NewTraitPrereq()
		one.Parent = parentList
		return one
	case prereq.Attribute:
		one := gurps.NewAttributePrereq(p.entity)
		one.Parent = parentList
		return one
	case prereq.ContainedQuantity:
		one := gurps.NewContainedQuantityPrereq()
		one.Parent = parentList
		return one
	case prereq.ContainedWeight:
		one := gurps.NewContainedWeightPrereq(p.entity)
		one.Parent = parentList
		return one
	case prereq.EquippedEquipment:
		one := gurps.NewEquippedEquipmentPrereq()
		one.Parent = parentList
		return one
	case prereq.Skill:
		one := gurps.NewSkillPrereq()
		one.Parent = parentList
		return one
	case prereq.Spell:
		one := gurps.NewSpellPrereq()
		one.Parent = parentList
		// Matching the owning spell's power source only makes sense for a prerequisite that belongs to a spell.
		one.SamePowerSource = p.ownerIsSpell
		return one
	case prereq.Script:
		one := gurps.NewScriptPrereq()
		one.Parent = parentList
		return one
	default:
		errs.Log(errs.New("unknown prerequisite type"), "type", prereqType.Key())
		return nil
	}
}

// prereqRow is the leading row of a prerequisite panel while it is being assembled. beginPrereqRow adds the parts every
// row starts with, the caller adds those particular to the prerequisite type, and finish adds the trailing and/or
// label and sets the layout.
type prereqRow struct {
	panel   *unison.Panel
	owner   *prereqPanel
	pr      gurps.Prereq
	depth   int
	inFront bool
}

// beginPrereqRow starts the leading row of a prerequisite panel with the buttons, the and/or label when it has text,
// and the "has" popup when has is not nil. An and/or label without text is instead added at the end of the row by
// finish, from where adjustAndOr moves it to the front should it gain text later.
func (p *prereqPanel) beginPrereqRow(depth int, pr gurps.Prereq, has *bool) *prereqRow {
	row := &prereqRow{
		panel:   unison.NewPanel(),
		owner:   p,
		pr:      pr,
		depth:   depth,
		inFront: andOrText(pr) != noAndOr,
	}
	p.createButtonsPanel(row.panel, depth, pr)
	if row.inFront {
		p.addAndOr(row.panel, pr)
	}
	if has != nil {
		addHasPopup(row.panel, has)
	}
	return row
}

// addTypeSwitcher adds the popup that switches the prerequisite to a different type.
func (r *prereqRow) addTypeSwitcher() {
	r.owner.addPrereqTypeSwitcher(r.panel, r.depth, r.pr)
}

// finish adds the and/or label if it was not placed in front, sets the row's layout to one column per child and returns
// that column count, so that callers can span subsequent rows across the columns after the buttons.
func (r *prereqRow) finish() (columns int) {
	if !r.inFront {
		r.owner.addAndOr(r.panel, r.pr)
	}
	columns = len(r.panel.Children())
	r.panel.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	return columns
}

// addIndentedSubRow adds a row beneath the leading row of a prerequisite panel that is indented past the buttons column
// and spans the remaining columns, calls populate to fill it in, then sets its layout to one column per child. When
// fill is true, the sub-row stretches to the width of the panel so that a field within it can grow.
func addIndentedSubRow(parent *unison.Panel, columns int, fill bool, populate func(subRow *unison.Panel)) {
	parent.AddChild(unison.NewPanel())
	subRow := unison.NewPanel()
	data := &unison.FlexLayoutData{HSpan: columns - 1}
	if fill {
		data.HAlign = align.Fill
		data.HGrab = true
	}
	subRow.SetLayoutData(data)
	parent.AddChild(subRow)
	populate(subRow)
	subRow.SetLayout(&unison.FlexLayout{
		Columns:  len(subRow.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
}

func (p *prereqPanel) createTraitPrereqPanel(depth int, pr *gurps.TraitPrereq) (main, focus unison.Paneler) {
	row := p.beginPrereqRow(depth, pr, &pr.Has)
	row.addTypeSwitcher()
	columns := row.finish()
	_, focus = addNameCriteriaPanel(row.panel, &pr.NameCriteria, columns-1, true)
	addNotesCriteriaPanel(row.panel, &pr.NotesCriteria, columns-1, true)
	addLevelCriteriaPanel(row.panel, nil, "", &pr.LevelCriteria, columns-1, true)
	return row.panel, focus
}

func (p *prereqPanel) createAttributePrereqPanel(depth int, pr *gurps.AttributePrereq) (main, focus unison.Paneler) {
	row := p.beginPrereqRow(depth, pr, &pr.Has)
	row.addTypeSwitcher()
	addIndentedSubRow(row.panel, row.finish(), false, func(subRow *unison.Panel) {
		extra := gurps.SizeFlag | gurps.DodgeFlag | gurps.ParryFlag | gurps.BlockFlag
		addAttributeChoicePopup(subRow, p.entity, noAndOr, &pr.Which, extra)
		addAttributeChoicePopup(subRow, p.entity, i18n.Text("combined with"), &pr.CombinedWith, extra|gurps.BlankFlag)
		_, focus = addNumericCriteriaPanel(subRow, nil, "", i18n.Text("which"), i18n.Text("Attribute Qualifier"),
			&pr.QualifierCriteria, fxp.Min, fxp.Max, 1, false, false)
	})
	return row.panel, focus
}

func (p *prereqPanel) createContainedQuantityPrereqPanel(depth int, pr *gurps.ContainedQuantityPrereq) (main, focus unison.Paneler) {
	row := p.beginPrereqRow(depth, pr, &pr.Has)
	row.addTypeSwitcher()
	_, focus = addQuantityCriteriaPanel(row.panel, nil, "", &pr.QualifierCriteria)
	row.finish()
	return row.panel, focus
}

func (p *prereqPanel) createContainedWeightPrereqPanel(depth int, pr *gurps.ContainedWeightPrereq) (main, focus unison.Paneler) {
	row := p.beginPrereqRow(depth, pr, &pr.Has)
	row.addTypeSwitcher()
	addIndentedSubRow(row.panel, row.finish(), false, func(subRow *unison.Panel) {
		_, focus = addWeightCriteriaPanel(subRow, nil, "", p.entity, &pr.WeightCriteria)
	})
	return row.panel, focus
}

func (p *prereqPanel) createEquippedEquipmentPrereqPanel(depth int, pr *gurps.EquippedEquipmentPrereq) (main, focus unison.Paneler) {
	row := p.beginPrereqRow(depth, pr, nil)
	row.addTypeSwitcher()
	columns := row.finish()
	_, focus = addNameCriteriaPanel(row.panel, &pr.NameCriteria, columns-1, true)
	addTagCriteriaPanel(row.panel, &pr.TagsCriteria, columns-1, true)
	return row.panel, focus
}

func (p *prereqPanel) createSkillPrereqPanel(depth int, pr *gurps.SkillPrereq) (main, focus unison.Paneler) {
	row := p.beginPrereqRow(depth, pr, &pr.Has)
	row.addTypeSwitcher()
	columns := row.finish()
	_, focus = addNameCriteriaPanel(row.panel, &pr.NameCriteria, columns-1, true)
	addSpecializationCriteriaPanel(row.panel, &pr.SpecializationCriteria, columns-1, true)
	addLevelCriteriaPanel(row.panel, nil, "", &pr.LevelCriteria, columns-1, true)
	return row.panel, focus
}

func (p *prereqPanel) createSpellPrereqPanel(depth int, pr *gurps.SpellPrereq) (main, focus unison.Paneler) {
	row := p.beginPrereqRow(depth, pr, &pr.Has)
	_, focus = addQuantityCriteriaPanel(row.panel, nil, "", &pr.QuantityCriteria)
	row.addTypeSwitcher()
	columns := row.finish()
	addIndentedSubRow(row.panel, columns, true, func(subRow *unison.Panel) {
		subTypePopup := addPopup(subRow, spellcmp.Types, &pr.SubType)
		popup, field := addStringCriteriaPanel(subRow, "", "", i18n.Text("Spell Qualifier"), &pr.QualifierCriteria, 1, false)
		// Neither "any" nor a college count has a qualifier to match against.
		adjustQualifier := func() {
			blank := pr.SubType == spellcmp.Any || pr.SubType == spellcmp.CollegeCount
			adjustPopupBlank(popup, blank)
			adjustFieldBlank(field, blank)
		}
		savedCallback := subTypePopup.SelectionChangedCallback
		subTypePopup.SelectionChangedCallback = func(pop *unison.PopupMenu[spellcmp.Type]) {
			savedCallback(pop)
			adjustQualifier()
		}
		adjustQualifier()
		if field.Enabled() {
			focus = field
		}
	})
	p.addPowerSourceCriteriaPanel(row.panel, pr, columns-1)
	return row.panel, focus
}

// addPowerSourceCriteriaPanel adds the row that restricts which power sources may satisfy a spell prerequisite. When
// the prerequisite belongs to a spell, an extra choice for matching that spell's own power source is offered.
func (p *prereqPanel) addPowerSourceCriteriaPanel(parent *unison.Panel, pr *gurps.SpellPrereq, hSpan int) {
	// Only a spell can ask for its own power source, and when it does, any explicit criteria is ignored. Neither of the
	// other states can be produced here, but a file edited by hand can hold them, so bring the prerequisite into line
	// before building the popup: otherwise the popup would show something other than what the prerequisite does.
	pr.SamePowerSource = pr.SamePowerSource && p.ownerIsSpell
	if pr.SamePowerSource {
		pr.PowerSourceCriteria.Compare = criteria.AnyText
	}
	panel := newCriteriaPanel(parent, hSpan, true)
	prefix := i18n.Text("and whose power source")
	choices := criteria.PrefixedStringComparisonChoices(prefix, prefix)
	if p.ownerIsSpell {
		choices = slices.Insert(choices, samePowerSourceIndex, prefix+" "+i18n.Text("is the same as this spell's"))
	}
	var criteriaField *StringField
	popup := newComparisonPopup(choices, p.powerSourceComparisonIndex(pr))
	popup.SelectionChangedCallback = func(pop *unison.PopupMenu[string]) {
		i := pop.SelectedIndex()
		pr.SamePowerSource = p.ownerIsSpell && i == samePowerSourceIndex
		if pr.SamePowerSource {
			pr.PowerSourceCriteria.Compare = criteria.AnyText
		} else {
			if p.ownerIsSpell && i > samePowerSourceIndex {
				i--
			}
			pr.PowerSourceCriteria.Compare = criteria.StringComparisons[i]
		}
		adjustFieldBlank(criteriaField, pr.SamePowerSource || pr.PowerSourceCriteria.Compare == criteria.AnyText)
		MarkModified(panel)
	}
	panel.AddChild(popup)
	criteriaField = addStringField(panel, i18n.Text("Power Source Qualifier"), "", &pr.PowerSourceCriteria.Qualifier)
	adjustFieldBlank(criteriaField, pr.SamePowerSource || pr.PowerSourceCriteria.Compare == criteria.AnyText)
}

// powerSourceComparisonIndex returns the index of the power source popup choice that reflects the prerequisite's
// current state, accounting for the extra choice that is inserted when the prerequisite belongs to a spell.
func (p *prereqPanel) powerSourceComparisonIndex(pr *gurps.SpellPrereq) int {
	if p.ownerIsSpell && pr.SamePowerSource {
		return samePowerSourceIndex
	}
	i := int(pr.PowerSourceCriteria.Compare.EnsureValid())
	if p.ownerIsSpell && i >= samePowerSourceIndex {
		i++
	}
	return i
}

func (p *prereqPanel) createScriptPrereqPanel(depth int, pr *gurps.ScriptPrereq) (main, focus unison.Paneler) {
	row := p.beginPrereqRow(depth, pr, nil)
	row.addTypeSwitcher()
	focus = addScriptField(row.panel, nil, "", i18n.Text("Prereq Script"),
		i18n.Text("The script should return text describing the missing prerequisite or an empty string if the prerequisite has been met"),
		func() string { return pr.Script }, func(text string) { pr.Script = text }, true)
	row.finish()
	return row.panel, focus
}
