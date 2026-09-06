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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

// newTestPrereqList returns a list, requiring all of its members, that holds one prerequisite of every kind, in the
// order of the prereq.Type enumeration, followed by a nested list.
func newTestPrereqList(entity *gurps.Entity) *gurps.PrereqList {
	root := gurps.NewPrereqList()
	for _, one := range []gurps.Prereq{
		gurps.NewTraitPrereq(),
		gurps.NewAttributePrereq(entity),
		gurps.NewContainedQuantityPrereq(),
		gurps.NewContainedWeightPrereq(entity),
		gurps.NewEquippedEquipmentPrereq(),
		gurps.NewSkillPrereq(),
		gurps.NewSpellPrereq(),
		gurps.NewScriptPrereq(),
		gurps.NewUnknownPrereq("future", []byte(`{"type":"future"}`)),
		gurps.NewPrereqList(),
	} {
		root.Prereqs = append(root.Prereqs, one.Clone(root))
	}
	return root
}

// prereqRows returns the panels for the members of the list shown by the given list panel, which follow the children
// of its own leading row.
func prereqRows(listPanel *unison.Panel) []*unison.Panel {
	layout, ok := listPanel.Layout().(*unison.FlexLayout)
	if !ok {
		return nil
	}
	return listPanel.Children()[layout.Columns:]
}

// TestPrereqRowsShareTheirLeadingRowShape builds a prerequisite panel holding one of every kind of prerequisite and
// checks the shape every row builder shares: the leading row is laid out with one column per child, the and/or label
// sits at the end of the leading row for the first member of a list and just after the buttons for the others, and any
// sub-row is indented past the buttons column and spans the rest.
func TestPrereqRowsShareTheirLeadingRowShape(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	root := newTestPrereqList(entity)
	p := newPrereqPanel(entity, &root, prereq.TypesForNonEquipment, true)
	c.Equal(1, len(p.Children()), "the panel holds the root list")
	listPanel := p.Children()[0]
	rows := prereqRows(listPanel)
	c.Equal(len(root.Prereqs), len(rows), "one row per member of the list")
	for i, row := range rows {
		pr := root.Prereqs[i]
		name := pr.PrereqType().Key()
		layout, ok := row.Layout().(*unison.FlexLayout)
		c.True(ok, "%s: the row uses a flex layout", name)
		columns := layout.Columns
		c.True(columns > 1 && columns <= len(row.Children()), "%s: the leading row spans %d columns", name, columns)
		label, ok := p.andOrMap[pr]
		c.True(ok, "%s: the row has an and/or label", name)
		if i == 0 {
			c.Equal(columns-1, row.IndexOfChild(label), "%s: the first member's label trails its leading row", name)
			c.Equal("", label.Text.String(), "%s: the first member has no and/or text", name)
		} else {
			c.Equal(1, row.IndexOfChild(label), "%s: a later member's label follows the buttons", name)
			c.Equal(i18n.Text("and"), label.Text.String(), "%s: a later member of an 'all' list shows 'and'", name)
		}
		switch pr.PrereqType() {
		case prereq.Attribute, prereq.ContainedWeight, prereq.Spell:
			c.True(len(row.Children()) >= columns+2, "%s: the row has a sub-row", name)
			c.Equal(0, len(row.Children()[columns].Children()), "%s: an empty filler precedes the sub-row", name)
			subRow := row.Children()[columns+1]
			data, ok2 := subRow.LayoutData().(*unison.FlexLayoutData)
			c.True(ok2, "%s: the sub-row has flex layout data", name)
			c.Equal(columns-1, data.HSpan, "%s: the sub-row spans the columns after the buttons", name)
			c.Equal(pr.PrereqType() == prereq.Spell, data.HGrab, "%s: only the spell sub-row grows", name)
			subLayout, ok3 := subRow.Layout().(*unison.FlexLayout)
			c.True(ok3, "%s: the sub-row uses a flex layout", name)
			c.Equal(len(subRow.Children()), subLayout.Columns, "%s: the sub-row has one column per child", name)
		case prereq.List:
			c.Equal(0, len(prereqRows(row)), "%s: the nested list is empty", name)
		default:
		}
	}
}

// TestPrereqRowAndOrLabelsFollowDeletion verifies that the leading row's layout is what adjustAndOr relies on: after
// the first member of a list is deleted, the label of the member that becomes first is moved to the end of its leading
// row and emptied, and the others keep theirs in place.
func TestPrereqRowAndOrLabelsFollowDeletion(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	root := newTestPrereqList(entity)
	p := newPrereqPanel(entity, &root, prereq.TypesForNonEquipment, true)
	listPanel := p.Children()[0]
	first := root.Prereqs[0]
	second := root.Prereqs[1]
	rows := prereqRows(listPanel)
	buttons := rows[0].Children()[0]
	c.Equal(1, len(buttons.Children()), "a non-list row offers only the delete button")
	deleteButton, ok := buttons.Children()[0].Self.(*unison.Button)
	c.True(ok, "the delete button is a button")
	deleteButton.ClickCallback()

	c.Equal(len(root.Prereqs), len(prereqRows(listPanel)), "the deleted member's row is gone")
	c.Equal(second, root.Prereqs[0], "the second member is now first")
	_, ok = p.andOrMap[first]
	c.False(ok, "the deleted member's label is forgotten")
	row := prereqRows(listPanel)[0]
	layout, ok := row.Layout().(*unison.FlexLayout)
	c.True(ok, "the row uses a flex layout")
	label := p.andOrMap[second]
	c.Equal(layout.Columns-1, row.IndexOfChild(label), "the new first member's label trails its leading row")
	c.Equal("", label.Text.String(), "the new first member has no and/or text")
	label = p.andOrMap[root.Prereqs[1]]
	c.Equal(1, prereqRows(listPanel)[1].IndexOfChild(label), "the next member's label still follows the buttons")
	c.Equal(i18n.Text("and"), label.Text.String())
}

// TestSpellPrereqPowerSourcePopupOffersTheOwningSpell verifies that the power source popup gains the "same as this
// spell's" choice only when the prerequisite belongs to a spell, and that the selection reflects the prerequisite's
// state either way.
func TestSpellPrereqPowerSourcePopupOffersTheOwningSpell(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	for _, ownerIsSpell := range []bool{true, false} {
		root := gurps.NewPrereqList()
		pr := gurps.NewSpellPrereq()
		pr.Parent = root
		pr.SamePowerSource = true
		root.Prereqs = append(root.Prereqs, pr)
		p := newPrereqPanel(entity, &root, prereq.TypesForNonEquipment, ownerIsSpell)
		row := prereqRows(p.Children()[0])[0]
		powerSource := row.Children()[len(row.Children())-1]
		popup, ok := powerSource.Children()[0].Self.(*unison.PopupMenu[string])
		c.True(ok, "ownerIsSpell=%v: the power source panel starts with its popup", ownerIsSpell)
		if ownerIsSpell {
			c.Equal(len(criteria.StringComparisons)+1, popup.ItemCount(), "a spell's prerequisite gains a choice")
			c.Equal(samePowerSourceIndex, popup.SelectedIndex(), "the extra choice reflects SamePowerSource")
			c.True(pr.SamePowerSource)
		} else {
			c.Equal(len(criteria.StringComparisons), popup.ItemCount(), "other owners get the plain comparisons")
			c.Equal(int(criteria.AnyText), popup.SelectedIndex(), "with no owning spell, any power source will do")
			c.False(pr.SamePowerSource, "a prerequisite that does not belong to a spell cannot match its power source")
		}
	}
}
