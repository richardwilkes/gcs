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
	"encoding/json/jsontext"
	"hash"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/filternode"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

// newTestListFilterPanel creates a filter editor over the given filter, with the session's memory of the field last
// used swapped out for an empty one so that one test cannot influence another. The panel has no window above it and
// nothing that implements ModifiableRoot, so every MarkModified the criteria widgets it shares with the editors make
// finds nothing to mark; a test that drives the editor and never panics is therefore also proof that the editor works
// outside of a document.
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
// and/or label and the must/must not popup, or nil when the row has none, as a row the editor can only preserve has.
func filterFieldPopup(row *unison.Panel) *unison.PopupMenu[string] {
	children := row.Children()
	if len(children) < 4 {
		return nil
	}
	if popup, ok := children[3].Self.(*unison.PopupMenu[string]); ok {
		return popup
	}
	return nil
}

// filterNotPopup returns the popup that inverts a row's node. Every row has one, group and condition alike, right
// after the buttons and the and/or label.
func filterNotPopup(row *unison.Panel) *unison.PopupMenu[string] {
	if popup, ok := row.Children()[2].Self.(*unison.PopupMenu[string]); ok {
		return popup
	}
	return nil
}

// filterRowIndentOf returns how far a row's buttons are indented, which is what shows how deeply the row's node is
// nested. A row with no border on its buttons yields -1, so that a missing indent fails the comparison readably.
func filterRowIndentOf(row *unison.Panel) float32 {
	if border := row.Children()[0].Border(); border != nil {
		return border.Insets().Left
	}
	return -1
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
	second := filter.Root.Children[0]
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
	c.True(filter.Root.Children[0] == second,
		"the row whose button was clicked is the first condition's, so the one added second must be what is left")
	c.Equal(filterGroupColumns+1, len(root.Children()), "deleting must take the row out of the group's row")
	c.Equal("", filterRowAndOr(filterChildRow(root, 0)), "the sole remaining row must have no and/or text")
}

// TestListFilterPanelAddsGroups verifies that the add-group button creates a sub-group owned by the group it was added
// to, that the sub-group's row offers everything a group's row offers plus the delete button the root lacks, that what
// goes into it is owned by it, indented one level further and joined by its own all/any popup rather than the root's,
// and that deleting it takes every row below it, and every and/or label they had, along with it.
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
	subButtons := filterRowButtons(subRow)
	c.Equal(3, len(subButtons), "a sub-group may be added to and deleted")
	if len(subButtons) != 3 {
		return
	}
	c.Equal(float32(filterRowIndent), filterRowIndentOf(subRow), "a row one level down is indented one level")

	// What is added to the sub-group belongs to it, not to the root group, and sits one level deeper still.
	subButtons[0].ClickCallback()
	subButtons[0].ClickCallback()
	c.Equal(2, len(group.Children), "both conditions must go into the sub-group")
	c.Equal(1, len(filter.Root.Children), "and none of them into the root group")
	for i, child := range group.Children {
		c.True(child.ParentGroup() == group, "condition %d must be owned by the sub-group it was added to", i)
	}
	c.Equal(filterGroupColumns+2, len(subRow.Children()), "the sub-group's row must hold the rows of both conditions")
	c.Equal(float32(2*filterRowIndent), filterRowIndentOf(filterChildRow(subRow, 0)),
		"a row two levels down is indented two levels")
	c.Equal(float32(2*filterRowIndent), filterRowIndentOf(filterChildRow(subRow, 1)),
		"whichever of the two levels down it is")

	// The rows within the sub-group are joined by the sub-group's own all/any popup, not by the root group's.
	c.Equal("", filterRowAndOr(filterChildRow(subRow, 0)), "the first row of the sub-group must have no and/or text")
	c.Equal(i18n.Text("and"), filterRowAndOr(filterChildRow(subRow, 1)),
		"the second row of a sub-group that matches all of its children must read \"and\"")
	allPopup, ok2 := subRow.Children()[3].Self.(*unison.PopupMenu[string])
	c.True(ok2, "the sub-group's row must hold an all/any popup")
	if !ok2 {
		return
	}
	selectPopupIndex(allPopup, 1)
	c.False(group.All, "picking the second choice must make the sub-group match any of its children")
	c.True(filter.Root.All, "and must leave the root group alone")
	c.Equal(i18n.Text("or"), filterRowAndOr(filterChildRow(subRow, 1)),
		"the second row of a sub-group that matches any of its children must read \"or\"")

	// A group may be nested within the sub-group, and a condition within that, each indented one level further.
	subButtons[1].ClickCallback()
	nested, ok3 := group.Children[0].(*gurps.FilterGroup)
	c.True(ok3, "the added node must be a group")
	if !ok3 {
		return
	}
	c.True(nested.ParentGroup() == group, "the nested group must be owned by the sub-group it was added to")
	nestedRow := filterChildRow(subRow, 0)
	c.Equal(float32(2*filterRowIndent), filterRowIndentOf(nestedRow), "the nested group's row is two levels down")
	filterRowButtons(nestedRow)[0].ClickCallback()
	c.Equal(1, len(nested.Children), "the condition must go into the nested group")
	c.Equal(filterGroupColumns+1, len(nestedRow.Children()), "and its row into the nested group's row")
	c.Equal(float32(3*filterRowIndent), filterRowIndentOf(filterChildRow(nestedRow, 0)),
		"a row three levels down is indented three levels")

	// Deleting the sub-group takes everything below it along, and the and/or label of every one of those rows is
	// forgotten with it, leaving only the root group's own.
	c.Equal(6, len(panel.andOrMap), "every row built so far has an and/or label of its own")
	subButtons[2].ClickCallback()
	c.Equal(0, len(filter.Root.Children), "deleting the sub-group must take it out of the root group")
	c.Equal(filterGroupColumns, len(root.Children()),
		"and its row, along with the rows of everything it held, out of the root group's row")
	c.Equal(1, len(panel.andOrMap), "only the root group's own and/or label may be left behind")
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

// TestListFilterPanelSameFieldChoiceChangesNothing verifies that picking the field a condition already tests leaves
// both the criteria the user filled in and the widgets holding them exactly as they were, rather than resetting the
// one and rebuilding the other.
func TestListFilterPanelSameFieldChoiceChangesNothing(t *testing.T) {
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
	text := criteria.Text{Compare: criteria.IsText, Qualifier: "Sword"}
	cond.Text = text
	before := slices.Clone(row.Children())

	chooseFilterField(c, panel, row, cond.Field)
	c.Equal("name", cond.Field, "the condition must still test the field it did")
	c.Equal(text, cond.Text, "re-picking the field a condition already tests must leave the criteria alone")
	after := row.Children()
	c.Equal(len(before), len(after), "and must leave the row with the children it had")
	if len(before) == len(after) {
		for i := range before {
			c.True(before[i] == after[i], "the row's widgets must be the same ones, but child %d was rebuilt", i)
		}
	}
	_, remembered := lastFilterFieldKeyUsed["adq"]
	c.False(remembered, "a choice that changes nothing must not be remembered as the field last used")
}

// TestListFilterPanelNegationPopups verifies that the must/must not popup of a row shows its own node's negation and
// writes back to that node alone, on a group as well as on a condition.
func TestListFilterPanelNegationPopups(t *testing.T) {
	c := check.New(t)
	filter := gurps.NewListFilter("Test")
	filter.Root.Not = true
	negated := gurps.NewFilterCondition(filter.Root, "name")
	negated.Not = true
	plain := gurps.NewFilterCondition(filter.Root, "name")
	filter.Root.Children = append(filter.Root.Children, negated, plain)
	panel := newTestListFilterPanel(t, "adq", filterFieldInfos(gurps.TraitFilterFields()), filter)
	root := filterRootRow(panel)

	rootNot := filterNotPopup(root)
	negatedNot := filterNotPopup(filterChildRow(root, 0))
	plainNot := filterNotPopup(filterChildRow(root, 1))
	c.NotNil(rootNot, "a group's row must hold a must/must not popup")
	c.NotNil(negatedNot, "a condition's row must hold one as well")
	c.NotNil(plainNot, "whether or not the condition is negated")
	if rootNot == nil || negatedNot == nil || plainNot == nil {
		return
	}
	c.Equal(1, rootNot.SelectedIndex(), "a group that must not match must show the second choice")
	c.Equal(1, negatedNot.SelectedIndex(), "and so must a condition that must not match")
	c.Equal(0, plainNot.SelectedIndex(), "while a condition that must match shows the first")

	// Each popup writes back to the node it belongs to, and to no other.
	selectPopupIndex(rootNot, 0)
	c.False(filter.Root.Not, "picking the first choice must make the group one that has to match")
	c.True(negated.Not, "and must leave the conditions below it alone")
	selectPopupIndex(plainNot, 1)
	c.True(plain.Not, "picking the second choice must invert the condition")
	c.True(negated.Not, "and must leave the other condition alone")
	selectPopupIndex(negatedNot, 0)
	c.False(negated.Not, "picking the first choice must put a negated condition back to one that has to match")
	c.True(plain.Not, "and must leave the other condition alone")
}

// TestListFilterPanelPreservesUnknownField verifies that a condition naming a field this version of GCS doesn't have --
// which a filter written by a newer version may hold -- is shown as something the editor cannot represent and left
// exactly as it was, rather than being quietly turned into a condition on a field this version does have.
func TestListFilterPanelPreservesUnknownField(t *testing.T) {
	c := check.New(t)
	filter := gurps.NewListFilter("Test")
	cond := gurps.NewFilterCondition(filter.Root, "not_a_field_this_version_knows")
	cond.Not = true
	text := criteria.Text{Compare: criteria.IsText, Qualifier: "Sword"}
	cond.Text = text
	filter.Root.Children = append(filter.Root.Children, cond)
	panel := newTestListFilterPanel(t, "adq", filterFieldInfos(gurps.TraitFilterFields()), filter)
	root := filterRootRow(panel)
	row := filterChildRow(root, 0)

	c.Equal("not_a_field_this_version_knows", cond.Field, "a condition on an unknown field must keep the field it names")
	c.Equal(text, cond.Text,
		"along with the criteria that went with it, which a newer version of GCS still understands")
	c.True(cond.Not, "and its negation")
	c.Nil(filterFieldPopup(row), "a condition the editor can only preserve must offer no field popup")
	c.Equal(3, len(row.Children()),
		"a preserved row holds its buttons, its and/or label and the label saying what it is, and nothing more")
	checkPreservedFilterRow(c, row, cond.Field, i18n.Text("Delete this condition"))

	// Nothing can be altered, but the condition can still be deleted deliberately.
	filterRowButtons(row)[0].ClickCallback()
	c.Equal(0, len(filter.Root.Children), "deleting must take the condition out of the group")
	c.Equal(filterGroupColumns, len(root.Children()), "and its row out of the group's row")
}

// TestListFilterPanelPreservesUnknownNode verifies that a node of a kind this version of GCS doesn't understand -- one
// a newer version wrote -- is shown as a row saying so, holds nothing that could alter it, keeps the data it was
// loaded with, and can be deleted deliberately.
func TestListFilterPanelPreservesUnknownNode(t *testing.T) {
	c := check.New(t)
	const data = `{"type":"future_node","weird":[1,2]}`
	filter := gurps.NewListFilter("Test")
	node := gurps.NewUnknownFilterNode("future_node", jsontext.Value(data))
	filter.Root.Children = append(filter.Root.Children, node)
	filter.EnsureValidity()
	panel := newTestListFilterPanel(t, "adq", filterFieldInfos(gurps.TraitFilterFields()), filter)
	root := filterRootRow(panel)
	row := filterChildRow(root, 0)

	c.Equal(3, len(row.Children()),
		"a preserved row holds its buttons, its and/or label and the label saying what it is, and nothing more")
	checkPreservedFilterRow(c, row, "future_node", i18n.Text("Delete this node"))
	c.Equal(data, node.Data.String(), "the node's original data must be left exactly as it was")

	filterRowButtons(row)[0].ClickCallback()
	c.Equal(0, len(filter.Root.Children), "deleting must take the node out of the group")
	c.Equal(filterGroupColumns, len(root.Children()), "and its row out of the group's row")
}

// checkPreservedFilterRow verifies that a row the editor can only preserve says what it holds, explains why nothing
// can be done to it, and offers the delete button alone.
func checkPreservedFilterRow(c check.Checker, row *unison.Panel, mentions, deleteTooltip string) {
	label, ok := row.Children()[2].Self.(*unison.Label)
	c.True(ok, "a preserved row must end with the label that says what it holds")
	if ok {
		c.Contains(label.String(), mentions, "the label must name what was found")
		// The tooltip is compared in the wrapped form every tooltip is built with, since that is what the labels
		// holding it were given.
		c.Equal(wrapTextForTooltip(preservedFilterNodeTooltip()), tooltipText(label.Tooltip),
			"the label must explain that what it stands for is kept as it is")
	}
	buttons := filterRowButtons(row)
	c.Equal(1, len(buttons), "a preserved row offers only the delete button, since there is nothing to add to or edit")
	if len(buttons) == 1 {
		c.Equal(deleteTooltip, tooltipText(buttons[0].Tooltip), "which must say what it would delete")
	}
}

// TestListFilterPanelSkipsNodeItCannotBuild verifies that a node the editor has no case for at all is passed over
// without a row, rather than bringing the editor down, and is left in the filter so that saving writes it back out.
func TestListFilterPanelSkipsNodeItCannotBuild(t *testing.T) {
	c := check.New(t)
	filter := gurps.NewListFilter("Test")
	filter.Root.Children = append(filter.Root.Children, &testUnhandledFilterNode{})
	filter.EnsureValidity()
	var panel *listFilterPanel
	c.NotPanics(func() {
		panel = newTestListFilterPanel(t, "adq", filterFieldInfos(gurps.TraitFilterFields()), filter)
	}, "a node the editor has no case for must not bring the editor down")
	if panel == nil {
		return
	}

	c.Equal(filterGroupColumns, len(filterRootRow(panel).Children()),
		"a node the editor has no case for must be passed over without a row")
	c.Equal(1, len(filter.Root.Children), "but must be left in the group it belongs to")
	c.True(filter.Root.Children[0].ParentGroup() == filter.Root, "still owned by that group")
}

// testUnhandledFilterNode is a filter node of a kind the editor has no case for. The editor has one for every kind the
// model can produce, so a stand-in is needed to reach the fallback that logs the node and builds no row for it.
type testUnhandledFilterNode struct {
	parent *gurps.FilterGroup
}

// NodeType implements gurps.FilterNode.
func (n *testUnhandledFilterNode) NodeType() filternode.Type {
	return filternode.Unknown
}

// ParentGroup implements gurps.FilterNode.
func (n *testUnhandledFilterNode) ParentGroup() *gurps.FilterGroup {
	return n.parent
}

// SetParentGroup implements gurps.FilterNode.
func (n *testUnhandledFilterNode) SetParentGroup(parent *gurps.FilterGroup) {
	n.parent = parent
}

// Clone implements gurps.FilterNode.
func (n *testUnhandledFilterNode) Clone(parent *gurps.FilterGroup) gurps.FilterNode {
	return &testUnhandledFilterNode{parent: parent}
}

// Hash implements gurps.FilterNode.
func (n *testUnhandledFilterNode) Hash(_ hash.Hash) {
}
