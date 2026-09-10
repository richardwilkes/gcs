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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

// newTestListFilterPanel creates a filter editor over the given filter, with the session's memory of the field last
// used swapped out for an empty one so that one test cannot influence another. The panel has no window above it and
// nothing that implements ModifiableRoot, so every MarkModified the editor makes finds nothing to mark; a test that
// drives the editor and never panics is therefore also proof that the editor works outside of a document.
func newTestListFilterPanel(t *testing.T, key string, fields []filterFieldInfo, filter *gurps.ListFilter) *listFilterPanel {
	t.Helper()
	swapForTest(t, &lastFilterFieldKeyUsed, make(map[string]string))
	return newListFilterPanel(key, filter, fields)
}

// filterRootRow returns the row of the filter's root group, which is the editor's only child.
func filterRootRow(p *listFilterPanel) *unison.Panel {
	return p.Children()[0]
}

// filterChildRow returns the row of the group row's child at the given position. A group's children follow its own
// fixed columns within its row panel.
func filterChildRow(groupRow *unison.Panel, index int) *unison.Panel {
	return groupRow.Children()[filterGroupColumns+index]
}

// filterRowButtons returns the buttons of a filter row, which live in the row's first child.
func filterRowButtons(row *unison.Panel) []*unison.Button {
	return panelsOfType[*unison.Button](row.Children()[0])
}

// filterRowAndOr returns the text of the label that joins a filter row to the one ahead of it. It always sits right
// after the buttons, whether or not it has anything to say.
func filterRowAndOr(row *unison.Panel) string {
	if label, ok := row.Children()[1].Self.(*unison.Label); ok {
		return label.String()
	}
	return "<not a label>"
}

// filterFieldPopup returns the popup that chooses the field a condition row tests, which follows the buttons, the
// and/or label and the must/must not popup.
func filterFieldPopup(row *unison.Panel) *unison.PopupMenu[string] {
	if popup, ok := row.Children()[3].Self.(*unison.PopupMenu[string]); ok {
		return popup
	}
	return nil
}

// chooseFilterField picks the field with the given key in a condition row's field popup, exactly as a choice made from
// the popup menu would.
func chooseFilterField(c check.Checker, p *listFilterPanel, row *unison.Panel, key string) {
	index := p.fieldIndex(key)
	c.True(index >= 0, "the %q field must be among those offered", key)
	popup := filterFieldPopup(row)
	c.NotNil(popup, "a condition row must hold a field popup")
	if index < 0 || popup == nil {
		return
	}
	item, _ := popup.ItemAt(index)
	popup.ChoiceMadeCallback(popup, index, item)
}

// filterRowColumns returns the number of columns a filter row lays its children out in.
func filterRowColumns(row *unison.Panel) int {
	if layout, ok := row.Layout().(*unison.FlexLayout); ok {
		return layout.Columns
	}
	return -1
}

// TestListFilterPanelAddsAndRemovesConditions verifies that the add button puts a new condition at the head of both
// the group and its row, that the and/or labels reflect where each condition sits and how the group combines them,
// and that the delete button takes a condition out of both.
func TestListFilterPanelAddsAndRemovesConditions(t *testing.T) {
	c := check.New(t)
	filter := gurps.NewListFilter("Test")
	panel := newTestListFilterPanel(t, "adq", filterFieldInfos(gurps.TraitFilterFields()), filter)
	root := filterRootRow(panel)

	buttons := filterRowButtons(root)
	c.Equal(2, len(buttons), "the root group may be added to, but not deleted")
	if len(buttons) != 2 {
		return
	}
	buttons[0].ClickCallback()
	c.Equal(1, len(filter.Root.Children), "adding a condition must put it into the group")
	first, ok := filter.Root.Children[0].(*gurps.FilterCondition)
	c.True(ok, "the added node must be a condition")
	if !ok {
		return
	}
	c.Equal("name", first.Field, "a new condition must start out on the first field")
	c.Equal(filterGroupColumns+1, len(root.Children()), "the condition's row must be added to the group's row")

	// A second condition goes ahead of the first, so that what was just added is right beneath the button that made it.
	buttons[0].ClickCallback()
	c.Equal(2, len(filter.Root.Children), "the group must now hold both conditions")
	c.True(filter.Root.Children[1] == gurps.FilterNode(first), "the earlier condition must be pushed down")
	c.Equal(filterGroupColumns+2, len(root.Children()), "the group's row must now hold both rows")

	// The first of the two has nothing ahead of it to be joined to; the second does.
	c.Equal("", filterRowAndOr(filterChildRow(root, 0)), "the first row must have no and/or text")
	c.Equal(i18n.Text("and"), filterRowAndOr(filterChildRow(root, 1)),
		"the second row of a group that matches all of its children must read \"and\"")

	// Switching the group from all to any changes what joins its children.
	allPopup, ok := root.Children()[3].Self.(*unison.PopupMenu[string])
	c.True(ok, "the group's row must hold an all/any popup")
	if !ok {
		return
	}
	selectPopupIndex(allPopup, 1)
	c.False(filter.Root.All, "picking the second choice must make the group match any of its children")
	c.Equal("", filterRowAndOr(filterChildRow(root, 0)), "the first row must still have no and/or text")
	c.Equal(i18n.Text("or"), filterRowAndOr(filterChildRow(root, 1)),
		"the second row of a group that matches any of its children must read \"or\"")

	// Deleting takes the condition out of the group and its row out of the panel.
	secondRow := filterChildRow(root, 1)
	rowButtons := filterRowButtons(secondRow)
	c.Equal(1, len(rowButtons), "a condition row offers only the delete button")
	if len(rowButtons) != 1 {
		return
	}
	rowButtons[0].ClickCallback()
	c.Equal(1, len(filter.Root.Children), "deleting must take the condition out of the group")
	c.True(filter.Root.Children[0] != gurps.FilterNode(first), "the deleted condition must be the one that was removed")
	c.Equal(filterGroupColumns+1, len(root.Children()), "deleting must take the row out of the group's row")
	c.Equal("", filterRowAndOr(filterChildRow(root, 0)), "the sole remaining row must have no and/or text")
}

// TestListFilterPanelAddsGroups verifies that the add-group button creates a sub-group owned by the group it was added
// to, and that the sub-group's row offers everything a group's row offers plus the delete button the root lacks.
func TestListFilterPanelAddsGroups(t *testing.T) {
	c := check.New(t)
	filter := gurps.NewListFilter("Test")
	panel := newTestListFilterPanel(t, "adq", filterFieldInfos(gurps.TraitFilterFields()), filter)
	root := filterRootRow(panel)

	buttons := filterRowButtons(root)
	c.Equal(2, len(buttons), "the root group may be added to, but not deleted")
	if len(buttons) != 2 {
		return
	}
	buttons[1].ClickCallback()
	c.Equal(1, len(filter.Root.Children), "adding a group must put it into the root group")
	group, ok := filter.Root.Children[0].(*gurps.FilterGroup)
	c.True(ok, "the added node must be a group")
	if !ok {
		return
	}
	c.True(group.ParentGroup() == filter.Root, "the new group must be owned by the group it was added to")
	c.True(group.All, "a new group must start out matching all of its children")

	subRow := filterChildRow(root, 0)
	c.Equal(filterGroupColumns, len(subRow.Children()), "a group's row starts out with only its own columns")
	c.Equal(3, len(filterRowButtons(subRow)), "a sub-group may be added to and deleted")
}

// TestListFilterPanelFieldChangeRebuildsCriteria verifies that choosing another field on a condition throws away the
// criteria that went with the old one and builds whatever the new one needs: a comma-aware text field for a list, a
// decimal field for a number, and nothing at all for a yes/no field.
func TestListFilterPanelFieldChangeRebuildsCriteria(t *testing.T) {
	c := check.New(t)
	filter := gurps.NewListFilter("Test")
	panel := newTestListFilterPanel(t, "adq", filterFieldInfos(gurps.TraitFilterFields()), filter)
	root := filterRootRow(panel)
	filterRowButtons(root)[0].ClickCallback()
	cond, ok := filter.Root.Children[0].(*gurps.FilterCondition)
	c.True(ok, "the added node must be a condition")
	if !ok {
		return
	}
	row := filterChildRow(root, 0)
	cond.Text = criteria.Text{Compare: criteria.IsText, Qualifier: "Sword"}

	chooseFilterField(c, panel, row, "tags")
	c.Equal("tags", cond.Field, "picking the tags field must change what the condition tests")
	c.True(cond.Text.IsZero(), "changing the field must reset the criteria that went with the old one")
	c.Equal("tags", lastFilterFieldKeyUsed["adq"], "the field just chosen must be remembered for this list type")
	textFields := panelsOfType[*StringField](row)
	c.Equal(1, len(textFields), "a list field must offer exactly one qualifier field")
	if len(textFields) == 1 {
		c.True(strings.Contains(tooltipText(textFields[0].Tooltip), "commas"),
			"a list field's qualifier must say that commas separate the values")
	}

	// The next condition added starts out on the field last chosen rather than back at the first one.
	filterRowButtons(root)[0].ClickCallback()
	next, ok2 := filter.Root.Children[0].(*gurps.FilterCondition)
	c.True(ok2, "the added node must be a condition")
	if ok2 {
		c.Equal("tags", next.Field, "a new condition must start out on the field last chosen")
	}

	// A yes/no field needs no criteria at all, so nothing follows the field popup.
	row = filterChildRow(root, 1)
	chooseFilterField(c, panel, row, "container")
	c.Equal("container", cond.Field, "picking the container field must change what the condition tests")
	c.Equal(4, len(row.Children()), "a yes/no field must leave nothing after the field popup")
	c.Equal(len(row.Children()), filterRowColumns(row), "a row lays its children out in one column each")

	// A number field gets a decimal qualifier.
	chooseFilterField(c, panel, row, "points")
	c.Equal("points", cond.Field, "picking the points field must change what the condition tests")
	c.Equal(1, len(panelsOfType[*DecimalField](row)), "a number field must offer exactly one decimal qualifier")
	c.Equal(0, len(panelsOfType[*StringField](row)), "a number field must not leave a text qualifier behind")
	c.Equal(len(row.Children()), filterRowColumns(row), "a row lays its children out in one column each")
}

// TestListFilterPanelWeightFieldForEquipment verifies that a condition on an equipment list's weight field offers a
// weight qualifier, which is the one kind of criteria no other list type has.
func TestListFilterPanelWeightFieldForEquipment(t *testing.T) {
	c := check.New(t)
	filter := gurps.NewListFilter("Test")
	panel := newTestListFilterPanel(t, "eqp", filterFieldInfos(gurps.EquipmentFilterFields()), filter)
	root := filterRootRow(panel)
	filterRowButtons(root)[0].ClickCallback()
	cond, ok := filter.Root.Children[0].(*gurps.FilterCondition)
	c.True(ok, "the added node must be a condition")
	if !ok {
		return
	}
	row := filterChildRow(root, 0)

	chooseFilterField(c, panel, row, "weight")
	c.Equal("weight", cond.Field, "picking the weight field must change what the condition tests")
	c.Equal(1, len(panelsOfType[*WeightField](row)), "a weight field must offer exactly one weight qualifier")
}

// TestListFilterPanelCoercesUnknownField verifies that a condition naming a field this version of GCS doesn't have --
// which a filter written by a newer version may hold -- is brought onto the first field rather than shown as
// something the editor cannot represent.
func TestListFilterPanelCoercesUnknownField(t *testing.T) {
	c := check.New(t)
	filter := gurps.NewListFilter("Test")
	cond := gurps.NewFilterCondition(filter.Root, "not_a_field_this_version_knows")
	cond.Text = criteria.Text{Compare: criteria.IsText, Qualifier: "Sword"}
	filter.Root.Children = append(filter.Root.Children, cond)
	panel := newTestListFilterPanel(t, "adq", filterFieldInfos(gurps.TraitFilterFields()), filter)

	c.Equal("name", cond.Field, "a condition on an unknown field must be brought onto the first field")
	c.True(cond.Text.IsZero(), "the criteria of a coerced condition must be reset")
	popup := filterFieldPopup(filterChildRow(filterRootRow(panel), 0))
	c.NotNil(popup, "a condition row must hold a field popup")
	if popup != nil {
		c.Equal(0, popup.SelectedIndex(), "the field popup must show the field the condition was brought onto")
	}
}
