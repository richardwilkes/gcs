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
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

// stringPopupsIn returns every string-valued popup found anywhere beneath the given panel. Each criteria row on a
// default carries exactly one, so collecting them all lets a test count the rows without depending on where within the
// row's nested panels each one happens to sit.
func stringPopupsIn(p *unison.Panel) []*unison.PopupMenu[string] {
	if popup, ok := p.Self.(*unison.PopupMenu[string]); ok {
		return []*unison.PopupMenu[string]{popup}
	}
	var popups []*unison.PopupMenu[string]
	for _, child := range p.Children() {
		popups = append(popups, stringPopupsIn(child)...)
	}
	return popups
}

// popupItems returns the text of every entry a string popup offers, which is what distinguishes one criteria row's
// popup from another's.
func popupItems(popup *unison.PopupMenu[string]) []string {
	items := make([]string, 0, popup.ItemCount())
	for i := range popup.ItemCount() {
		if item, ok := popup.ItemAt(i); ok {
			items = append(items, item)
		}
	}
	return items
}

// findTagCriteriaPopup returns the sole tag criteria popup beneath the given panel, or nil if there is none. It is
// identified by the choices addTagCriteriaPanel builds for it. A row carrying more than one is a failure, since the
// extra would be a second, unattached way to edit the same criteria.
func findTagCriteriaPopup(c check.Checker, p *unison.Panel) *unison.PopupMenu[string] {
	choices := criteria.PrefixedStringComparisonChoices(i18n.Text("and at least one tag"), i18n.Text("and all tags"))
	var found []*unison.PopupMenu[string]
	for _, popup := range stringPopupsIn(p) {
		if slices.Equal(choices, popupItems(popup)) {
			found = append(found, popup)
		}
	}
	if len(found) == 0 {
		return nil
	}
	c.Equal(1, len(found), "a default must not hold more than one tag criteria popup")
	return found[0]
}

// findAttributeChoicePopup returns the first default-type switcher popup found anywhere beneath the given panel, or nil
// if there is none. It is used by the tests to change a default's type the same way a user's popup selection would.
func findAttributeChoicePopup(p *unison.Panel) *unison.PopupMenu[*gurps.AttributeChoice] {
	if popup, ok := p.Self.(*unison.PopupMenu[*gurps.AttributeChoice]); ok {
		return popup
	}
	for _, child := range p.Children() {
		if popup := findAttributeChoicePopup(child); popup != nil {
			return popup
		}
	}
	return nil
}

// selectPopupIndex selects the given index and then runs the popup's selection callback once, mirroring what happens
// when the user picks that entry. The callback is invoked directly rather than through SelectIndex, which wraps it in
// unison.SafeCall and would swallow a failure inside it, so it is taken off the popup while the selection is made.
func selectPopupIndex[T comparable](popup *unison.PopupMenu[T], index int) {
	callback := popup.SelectionChangedCallback
	popup.SelectionChangedCallback = nil
	popup.SelectIndex(index)
	popup.SelectionChangedCallback = callback
	callback(popup)
}

// qualifierFieldFor returns the qualifier field that sits beside the given criteria popup, or nil if there is none.
func qualifierFieldFor(popup *unison.PopupMenu[string]) *StringField {
	for _, child := range popup.Parent().Children() {
		if field, ok := child.Self.(*StringField); ok {
			return field
		}
	}
	return nil
}

// TestDefaultsPanelSkillDefaultHasTagRow verifies that a skill default offers a tag criteria row, that the row is bound
// to the default's own Tags criteria, and that the row goes away when the default is no longer a skill default, taking
// what it held with it along with the name and specialization criteria.
func TestDefaultsPanelSkillDefaultHasTagRow(t *testing.T) {
	c := check.New(t)
	def := &gurps.SkillDefault{
		DefaultType:    gurps.SkillID,
		Name:           criteria.Text{Compare: criteria.IsText, Qualifier: "Broadsword"},
		Specialization: criteria.Text{Compare: criteria.IsText, Qualifier: "Fencing"},
	}
	defs := []*gurps.SkillDefault{def}
	panel := newDefaultsPanel(gurps.NewEntity(), &defs)

	// Child 0 is the add button, so the sole default's row is child 1.
	c.Equal(2, len(panel.Children()), "expected add button + one default row")
	row := panel.Children()[1]

	popup := findTagCriteriaPopup(c, row)
	c.NotNil(popup, "expected a tag criteria popup in the skill default row")
	if popup == nil {
		return
	}

	// Picking "is" must land on the default's own Tags criteria, proving the row is bound to it rather than to a
	// throwaway.
	index := slices.Index(criteria.AllStringComparisons, criteria.IsText)
	c.True(index >= 0, "the \"is\" comparison must be present in the comparison list")
	selectPopupIndex(popup, index)
	c.Equal(criteria.IsText, def.Tags.Compare, "selecting \"is\" must set the default's tag comparison")

	// Switching the default's type to an attribute must take the tag row away with the rest of the skill criteria.
	typePopup := findAttributeChoicePopup(row)
	c.NotNil(typePopup, "expected a default-type popup in the row")
	if typePopup == nil {
		return
	}
	dexIndex := -1
	for i := range typePopup.ItemCount() {
		if choice, ok := typePopup.ItemAt(i); ok && choice.Key == gurps.DexterityID {
			dexIndex = i
			break
		}
	}
	c.True(dexIndex >= 0, "the type popup must offer DX")
	selectPopupIndex(typePopup, dexIndex)
	c.Equal(gurps.DexterityID, def.DefaultType, "selecting DX must change the default's type")
	c.Nil(findTagCriteriaPopup(c, row), "an attribute default must not offer a tag criteria row")
	// Nothing consults the skill criteria on an attribute default, and there is no longer any way to see or edit
	// them, so leaving them behind would only make the default look different from one that never had them.
	c.True(def.Tags.IsZero(), "switching to an attribute default must clear the tag criteria")
	c.True(def.Name.IsZero(), "switching to an attribute default must clear the name criteria")
	c.True(def.Specialization.IsZero(), "switching to an attribute default must clear the specialization criteria")
}

// TestDefaultsPanelUnknownComparisonBlanksQualifier verifies that a criteria whose comparison is not a known one, which
// matches everything just as "is anything" does, has its qualifier field blanked out the way "is anything" does. The
// popup already lands on the "is anything" entry for it; a qualifier field left editable beside that would take text
// that matching ignores.
func TestDefaultsPanelUnknownComparisonBlanksQualifier(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	for _, one := range []struct {
		compare criteria.StringComparison
		blank   bool
	}{
		{compare: criteria.AnyText, blank: true},
		{compare: "any", blank: true},
		{compare: criteria.IsText, blank: false},
	} {
		def := &gurps.SkillDefault{
			DefaultType: gurps.SkillID,
			Tags:        criteria.Text{Compare: one.compare, Qualifier: "Combat"},
		}
		defs := []*gurps.SkillDefault{def}
		row := newDefaultsPanel(entity, &defs).Children()[1]
		popup := findTagCriteriaPopup(c, row)
		c.NotNil(popup, "expected a tag criteria popup for comparison %q", one.compare)
		if popup == nil {
			continue
		}
		field := qualifierFieldFor(popup)
		c.True(field != nil, "expected a qualifier field beside the tag criteria popup for comparison %q", one.compare)
		if field == nil {
			continue
		}
		c.Equal(!one.blank, field.Enabled(), "qualifier field enabled state for comparison %q", one.compare)
	}
}

// TestDefaultsPanelUnnormalizedSkillTypeShowsCriteriaRows verifies that a hand-edited default whose type is written as
// "Skill" gets the same criteria rows as one written as "skill". The model treats the type case-insensitively
// everywhere else, so the editor must not be the one place that leaves such a default with no way to edit it.
func TestDefaultsPanelUnnormalizedSkillTypeShowsCriteriaRows(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()

	normalized := []*gurps.SkillDefault{{DefaultType: gurps.SkillID}}
	normalizedRow := newDefaultsPanel(entity, &normalized).Children()[1]
	// The name, specialization and tag criteria rows, plus the "when the Tech Level" row.
	c.Equal(4, len(stringPopupsIn(normalizedRow)), "expected four criteria rows on a skill default")

	unnormalized := []*gurps.SkillDefault{{DefaultType: "Skill"}}
	unnormalizedRow := newDefaultsPanel(entity, &unnormalized).Children()[1]
	c.Equal(len(stringPopupsIn(normalizedRow)), len(stringPopupsIn(unnormalizedRow)),
		"a default typed \"Skill\" must show the same criteria rows as one typed \"skill\"")
	c.NotNil(findTagCriteriaPopup(c, unnormalizedRow), "a default typed \"Skill\" must offer a tag criteria row")
}
