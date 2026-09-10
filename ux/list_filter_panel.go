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
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

const (
	// filterGroupColumns is the number of columns a group's own row holds: the buttons, the and/or label, the
	// must/must not popup and the all/any popup. The group's children follow, each spanning all of them, so the child
	// at position i within the group is the row panel's child at filterGroupColumns+i.
	filterGroupColumns = 4
	// filterRowIndent is how far, in pixels, each level of nesting indents a row.
	filterRowIndent = 20
)

// lastFilterFieldKeyUsed remembers, per list type key, the field the user last chose in a condition row, so the next
// condition starts out testing the same field. Session only.
var lastFilterFieldKeyUsed = make(map[string]string)

// listFilterPanel is the editor for a saved filter's tree of nodes. It holds the tree's root group row as its only
// child; every other row hangs off of that one, since a group's children are added to the group's own row panel.
type listFilterPanel struct {
	unison.Panel
	key      string
	filter   *gurps.ListFilter
	fields   []filterFieldInfo
	andOrMap map[gurps.FilterNode]*unison.Label
}

// newListFilterPanel creates the editor for the given filter. key is the list type the filter belongs to, used to
// remember the last field chosen, and fields are the fields that list type offers, in the order they are shown.
func newListFilterPanel(key string, filter *gurps.ListFilter, fields []filterFieldInfo) *listFilterPanel {
	p := &listFilterPanel{
		key:      key,
		filter:   filter,
		fields:   fields,
		andOrMap: make(map[gurps.FilterNode]*unison.Label),
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	// No titled border here, unlike the prerequisites and features sections of an editor: this panel is the whole of a
	// dialog, which supplies the title, so all it needs is a little breathing room.
	p.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(4)))
	root, _ := p.createGroupPanel(0, p.filter.Root)
	p.AddChild(root)
	return p
}

// createGroupPanel creates the row for a group: its buttons, its and/or label, the popup that inverts it and the popup
// that says whether all of its children or any one of them has to match. The rows of the group's children are added
// after those, each spanning all of the row's columns.
func (p *listFilterPanel) createGroupPanel(depth int, group *gurps.FilterGroup) (main, focus unison.Paneler) {
	row := p.beginFilterRow(depth, group)
	addNotPopup(row, &group.Not)
	popup := addBoolPopup(row, i18n.Text("match all of:"), i18n.Text("match any of:"), &group.All)
	callback := popup.SelectionChangedCallback
	popup.SelectionChangedCallback = func(pop *unison.PopupMenu[string]) {
		callback(pop)
		// Switching between all and any changes what the children's and/or labels say.
		p.adjustAndOrForGroup(group)
	}
	row.SetLayout(&unison.FlexLayout{
		Columns:  filterGroupColumns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	row.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	for _, child := range group.Children {
		p.addToGroup(row, depth+1, -1, child)
	}
	return row, popup
}

// createConditionPanel creates the row for a condition: its buttons, its and/or label, the popup that inverts it, the
// popup that chooses the field to test and, for every kind of field but a yes/no one, the criteria the field is
// compared against.
func (p *listFilterPanel) createConditionPanel(depth int, cond *gurps.FilterCondition) (main, focus unison.Paneler) {
	// A filter written by a newer version of GCS may name a field this version has never heard of. Rather than drop
	// the condition, point it at the first field and clear the criteria, so that what the editor shows is what the
	// condition now does.
	if p.fieldIndex(cond.Field) == -1 && len(p.fields) != 0 {
		cond.Field = p.fields[0].key
		resetFilterConditionCriteria(cond)
	}
	row := p.beginFilterRow(depth, cond)
	addNotPopup(row, &cond.Not)
	popup := p.addFieldPopup(row, cond)
	p.addConditionCriteria(row, cond)
	setFilterRowLayout(row)
	return row, popup
}

// createUnknownNodePanel creates the row for a node this version of GCS doesn't understand. No editing is offered,
// since we have no idea what the data means, but the row is shown so the node is visible and can be deleted
// deliberately.
func (p *listFilterPanel) createUnknownNodePanel(depth int, node *gurps.UnknownFilterNode) (main, focus unison.Paneler) {
	row := p.beginFilterRow(depth, node)
	label := NewFieldLeadingLabel(fmt.Sprintf(i18n.Text("Unknown filter node type %q; it will be preserved, but never matches"),
		node.Kind), false)
	label.Tooltip = newWrappedTooltip(i18n.Text("This was most likely created by a newer version of GCS. Its original data will be written back out unchanged when this filter is saved."))
	row.AddChild(label)
	setFilterRowLayout(row)
	return row, row
}

// beginFilterRow starts a row with the parts every row has: the buttons, then the label that joins the row to the one
// ahead of it. That label always sits right after the buttons, even while it has nothing to say. This differs from a
// prerequisite row, which parks an empty and/or label at the end of the row and moves it forward when it gains text:
// a condition row throws away and rebuilds everything after its field popup whenever the field changes, so a label
// parked at the end would not survive that.
func (p *listFilterPanel) beginFilterRow(depth int, node gurps.FilterNode) *unison.Panel {
	row := unison.NewPanel()
	p.createButtonsPanel(row, depth, node)
	p.addAndOr(row, node)
	return row
}

// createButtonsPanel adds the row's leading column of buttons, indented to show how deeply the node is nested. Only a
// group can have things added to it, and only a node that has a parent can be deleted, which leaves the root group
// with nothing but the two add buttons.
func (p *listFilterPanel) createButtonsPanel(row *unison.Panel, depth int, node gurps.FilterNode) {
	buttons := unison.NewPanel()
	buttons.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: float32(depth * filterRowIndent)}))
	row.AddChild(buttons)
	if group, ok := node.(*gurps.FilterGroup); ok {
		addConditionButton := unison.NewSVGButton(unison.CircledAddSVG)
		addConditionButton.Tooltip = newWrappedTooltip(i18n.Text("Add a condition"))
		addConditionButton.ClickCallback = func() {
			p.insertIntoGroup(row, depth, group, gurps.NewFilterCondition(group, p.defaultFieldKey()))
		}
		buttons.AddChild(addConditionButton)

		addGroupButton := unison.NewSVGButton(svg.CircledVerticalEllipsis)
		addGroupButton.Tooltip = newWrappedTooltip(i18n.Text("Add a group"))
		addGroupButton.ClickCallback = func() {
			p.insertIntoGroup(row, depth, group, gurps.NewFilterGroup(group))
		}
		buttons.AddChild(addGroupButton)
	}
	if parentGroup := node.ParentGroup(); parentGroup != nil {
		deleteButton := unison.NewSVGButton(unison.TrashSVG)
		deleteButton.ClickCallback = func() {
			p.removeFromGroup(row, parentGroup, node)
		}
		buttons.AddChild(deleteButton)
	}
	buttons.SetLayout(&unison.FlexLayout{
		Columns: len(buttons.Children()),
	})
}

// insertIntoGroup puts child at the head of the group's children and its row at the head of the group's rows. New
// nodes go first so that the one just added is where the eye already is, right beneath the button that made it.
func (p *listFilterPanel) insertIntoGroup(row *unison.Panel, depth int, group *gurps.FilterGroup, child gurps.FilterNode) {
	group.Children = slices.Insert(group.Children, 0, child)
	p.addToGroup(row, depth+1, 0, child)
	p.adjustAndOrForGroup(group)
	MarkModified(p)
}

// removeFromGroup takes the node out of the group and its row out of the panel.
func (p *listFilterPanel) removeFromGroup(row *unison.Panel, group *gurps.FilterGroup, node gurps.FilterNode) {
	if i := slices.IndexFunc(group.Children, func(elem gurps.FilterNode) bool { return elem == node }); i != -1 {
		group.Children = slices.Delete(group.Children, i, i+1)
	}
	delete(p.andOrMap, node)
	row.RemoveFromParent()
	p.adjustAndOrForGroup(group)
	MarkModified(p)
}

// addToGroup builds the row for a child of a group and adds it to the group's row panel, either at the given position
// among the group's children or, when index is negative, at the end.
func (p *listFilterPanel) addToGroup(parent *unison.Panel, depth, index int, child gurps.FilterNode) {
	var panel, focus unison.Paneler
	switch one := child.(type) {
	case *gurps.FilterGroup:
		panel, focus = p.createGroupPanel(depth, one)
	case *gurps.FilterCondition:
		panel, focus = p.createConditionPanel(depth, one)
	case *gurps.UnknownFilterNode:
		panel, focus = p.createUnknownNodePanel(depth, one)
	default:
		errs.Log(errs.New("unknown filter node type"), "type", reflect.TypeOf(child).String())
		return
	}
	panel.AsPanel().SetLayoutData(&unison.FlexLayoutData{
		HSpan:  filterGroupColumns,
		HAlign: align.Fill,
		HGrab:  true,
	})
	if index < 0 {
		parent.AddChild(panel)
	} else {
		parent.AddChildAtIndex(panel, filterGroupColumns+index)
	}
	focus.AsPanel().RequestFocus()
}

// addFieldPopup adds the popup that chooses which field a condition tests. Changing the field discards the criteria
// that went with the old one and rebuilds the trailing part of the row, since a field of another kind needs another
// kind of criteria.
func (p *listFilterPanel) addFieldPopup(row *unison.Panel, cond *gurps.FilterCondition) *unison.PopupMenu[string] {
	titles := make([]string, len(p.fields))
	for i, info := range p.fields {
		titles[i] = info.title
	}
	popup := unison.NewPopupMenu[string]()
	popup.AddItem(titles...)
	popup.SelectIndex(p.fieldIndex(cond.Field))
	popup.Tooltip = newWrappedTooltip(i18n.Text("The field to test"))
	popup.ChoiceMadeCallback = func(pop *unison.PopupMenu[string], index int, _ string) {
		pop.SelectIndex(index)
		if index < 0 || index >= len(p.fields) || p.fields[index].key == cond.Field {
			return
		}
		cond.Field = p.fields[index].key
		lastFilterFieldKeyUsed[p.key] = cond.Field
		resetFilterConditionCriteria(cond)
		first := row.IndexOfChild(pop) + 1
		for i := len(row.Children()) - 1; i >= first; i-- {
			row.RemoveChildAtIndex(i)
		}
		p.addConditionCriteria(row, cond)
		setFilterRowLayout(row)
		markListFilterPanelForLayout(p)
		MarkModified(p)
	}
	row.AddChild(popup)
	return popup
}

// addConditionCriteria adds the criteria the condition's field is compared against, which is whichever one goes with
// the field's kind. A yes/no field has nothing to compare against, so it gets none: the must/must not popup ahead of
// the field already says all there is to say about it.
func (p *listFilterPanel) addConditionCriteria(row *unison.Panel, cond *gurps.FilterCondition) {
	info, ok := p.fieldInfo(cond.Field)
	if !ok {
		return
	}
	// The criteria follow the field's title, so that a row reads "must have a name that is ...".
	prefix := i18n.Text("that")
	switch info.kind {
	case gurps.FilterFieldText:
		addStringCriteriaPanel(row, prefix, prefix, i18n.Text("Text Qualifier"), &cond.Text, 1, false)
	case gurps.FilterFieldList:
		addListCriteriaPanel(row, &cond.Text)
	case gurps.FilterFieldNumber:
		addNumericCriteriaPanel(row, nil, "", prefix, i18n.Text("Number Qualifier"), &cond.Number, fxp.Min, fxp.Max, 1,
			false, false)
	case gurps.FilterFieldWeight:
		// A weight criteria adds its parts directly to what it is given, so it needs a panel of its own to sit in.
		addWeightCriteriaPanel(newCriteriaPanel(row, 1, false), nil, "", prefix, nil, &cond.Weight)
	case gurps.FilterFieldBool:
		// Nothing to add. A yes/no field is satisfied by the value being true, and whether that is what the condition
		// wants is what the must/must not popup ahead of the field says.
	}
}

// addListCriteriaPanel adds the criteria for a field holding a list of values, such as tags. A list is matched value
// by value, so the comparison reads "where at least one" or "where all" rather than the plain "that" a single value
// gets.
func addListCriteriaPanel(parent *unison.Panel, text *criteria.Text) (*unison.PopupMenu[string], *StringField) {
	popup, field := addStringCriteriaPanel(parent, i18n.Text("where at least one"), i18n.Text("where all"),
		i18n.Text("List Qualifier"), text, 1, false)
	field.Tooltip = newWrappedTooltip(i18n.Text(`Separate multiple values with commas to match any one of them, e.g. "Sword, Axe"`))
	return popup, field
}

// addNotPopup adds the popup that inverts a node's result. The value is stored as the negative, so that the common
// case -- a node that has to match -- is the zero value and stays out of the JSON, which is why the second choice is
// the one that sets it.
func addNotPopup(parent *unison.Panel, not *bool) *unison.PopupMenu[string] {
	popup := unison.NewPopupMenu[string]()
	popup.AddItem(i18n.Text("must"))
	popup.AddItem(i18n.Text("must not"))
	if *not {
		popup.SelectIndex(1)
	} else {
		popup.SelectIndex(0)
	}
	popup.SelectionChangedCallback = func(pop *unison.PopupMenu[string]) {
		*not = pop.SelectedIndex() == 1
		MarkModified(parent)
	}
	parent.AddChild(popup)
	return popup
}

// addAndOr adds the label that joins a row to the one ahead of it and remembers it, so that it can be brought into
// line whenever the group it belongs to changes.
func (p *listFilterPanel) addAndOr(row *unison.Panel, node gurps.FilterNode) {
	label := NewFieldLeadingLabel(filterAndOrText(node), false)
	row.AddChild(label)
	p.andOrMap[node] = label
}

// adjustAndOrForGroup brings the and/or labels of every one of the group's children into line and lays the editor out
// again, which is needed whether or not any of them changed, since the caller has just changed the group.
func (p *listFilterPanel) adjustAndOrForGroup(group *gurps.FilterGroup) {
	for _, child := range group.Children {
		p.adjustAndOr(child)
	}
	markListFilterPanelForLayout(p)
}

// adjustAndOr updates the node's and/or label. The label never moves, so unlike a prerequisite's there is nothing to
// do beyond retitling it.
func (p *listFilterPanel) adjustAndOr(node gurps.FilterNode) {
	if label, ok := p.andOrMap[node]; ok {
		if text := filterAndOrText(node); text != label.String() {
			label.SetTitle(text)
		}
	}
}

// filterAndOrText returns the text of the label that joins the node to its siblings ahead of it.
func filterAndOrText(node gurps.FilterNode) string {
	group := node.ParentGroup()
	if group == nil {
		return noAndOr
	}
	return joiningText(len(group.Children), group.Children[0] == node, group.All)
}

// setFilterRowLayout gives a row one column per child. A condition row's trailing criteria are thrown away and rebuilt
// whenever the field being tested changes, so the count has to be set again each time that happens.
func setFilterRowLayout(row *unison.Panel) {
	row.SetLayout(&unison.FlexLayout{
		Columns:  len(row.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
}

// markListFilterPanelForLayout lays the editor out again from top to bottom after a structural change.
// MarkRootAncestorForLayoutRecursively, which the editors in the workspace use for this, is a no-op here: it looks for
// an ancestor dock container or dockable, and the filter editor lives in a dialog, which is neither.
func markListFilterPanelForLayout(p unison.Paneler) {
	panel := p.AsPanel()
	panel.MarkForLayoutRecursively()
	panel.MarkForLayoutRecursivelyUpward()
	panel.MarkForRedraw()
}

// resetFilterConditionCriteria puts the condition's criteria back to the defaults that accept anything. Only the one
// that goes with the field's kind is ever consulted, so leaving the others as they were would keep values around that
// nothing shows, nothing uses, and yet would still be written back out to disk.
func resetFilterConditionCriteria(cond *gurps.FilterCondition) {
	cond.Text = criteria.Text{}
	cond.Number = criteria.Number{}
	cond.Weight = criteria.Weight{}
}

// defaultFieldKey returns the key of the field a newly added condition should start out testing: the one last chosen
// for this list type, if it is still among the fields, and otherwise the first field.
func (p *listFilterPanel) defaultFieldKey() string {
	if len(p.fields) == 0 {
		return ""
	}
	if key, ok := lastFilterFieldKeyUsed[p.key]; ok && p.fieldIndex(key) != -1 {
		return key
	}
	return p.fields[0].key
}

// fieldIndex returns the position of the field with the given key, or -1 if there is none.
func (p *listFilterPanel) fieldIndex(key string) int {
	return slices.IndexFunc(p.fields, func(info filterFieldInfo) bool { return info.key == key })
}

// fieldInfo returns the field with the given key, if there is one.
func (p *listFilterPanel) fieldInfo(key string) (info filterFieldInfo, ok bool) {
	if i := p.fieldIndex(key); i != -1 {
		return p.fields[i], true
	}
	return filterFieldInfo{}, false
}
