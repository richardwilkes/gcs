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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

// TestAttributeTypeChangeIsUndoable verifies that changing an attribute's type through its popup is a single undo edit
// of the whole set of definitions, so that undoing a change from a pool to a separator or to another type without
// thresholds brings back what the change cleared: the thresholds, and for a separator the fields it has no use for.
func TestAttributeTypeChangeIsUndoable(t *testing.T) {
	for _, attrType := range []attribute.Type{attribute.PoolSeparator, attribute.Integer} {
		t.Run(attrType.Key(), func(t *testing.T) {
			c := check.New(t)
			d := newTestAttributeSettingsDockable()
			initTestSettingsContent(&d.undoableSettingsDockable)
			expandAllAttributes(d)
			fp := d.model.Set[gurps.FatiguePointsID]
			c.Equal(attribute.Pool, fp.Type)
			c.True(len(fp.Thresholds) > 1, "the factory FP has several thresholds")
			before := fp.Clone()
			beforeHash := gurps.Hash64(d.model)
			rowsBefore := len(uxtest.PanelsOfType[*thresholdSettingsPanel](d.AsPanel()))

			widgetFor[*Popup[attribute.Type]](t, d, fp.KeyPrefix+"type").Select(attrType)
			fp = d.model.Set[gurps.FatiguePointsID]
			c.Equal(attrType, fp.Type, "the popup writes the model")
			if fp.IsSeparator() {
				c.Equal("", fp.Base, "a separator has no base")
			} else {
				c.Equal(before.Base, fp.Base, "only a separator loses its base")
			}
			c.Nil(fp.Thresholds, "a type other than a pool has no thresholds")
			c.Equal("Undo Attribute Type", d.undoMgr.UndoTitle())

			d.undoMgr.Undo()
			fp = d.model.Set[gurps.FatiguePointsID]
			c.Equal(attribute.Pool, fp.Type, "undo restores the type")
			c.Equal(before.FullName, fp.FullName, "undo restores the full name")
			c.Equal(before.Base, fp.Base, "undo restores the base")
			c.Equal(before.CostPerPoint, fp.CostPerPoint, "undo restores the cost per point")
			c.Equal(thresholdStates(before), thresholdStates(fp), "undo restores the thresholds")
			c.Equal(beforeHash, gurps.Hash64(d.model), "undo restores the definitions exactly")
			c.False(d.undoMgr.CanUndo(), "the type change is a single edit")
			c.Equal(rowsBefore, len(uxtest.PanelsOfType[*thresholdSettingsPanel](d.AsPanel())),
				"the panels are rebuilt from the restored definitions")

			d.undoMgr.Redo()
			c.Equal(attrType, d.model.Set[gurps.FatiguePointsID].Type, "redo changes the type again")
		})
	}
}

// TestAttributeRowsStartCollapsedToTheirHeader verifies that every attribute's row is built collapsed: a drag handle, a
// button column holding the delete button alone, and a header naming the definition by its ID, its name when it has
// one, and its type, with none of its fields, nor the thresholds of a pool, nor the button that adds one, so that
// nothing but deleting and rearranging can be done to a definition until its row is expanded.
func TestAttributeRowsStartCollapsedToTheirHeader(t *testing.T) {
	c := check.New(t)
	defs := testAttrDefs("st", "hp", "sep", "nameless")
	hp := defs.Set["hp"]
	hp.Type = attribute.Pool
	hp.Thresholds = []*gurps.PoolThreshold{{State: "Reeling"}}
	sep := defs.Set["sep"]
	sep.Type = attribute.PrimarySeparator
	sep.Name = "Primary"
	defs.Set["nameless"].Name = " "
	d := newTestAttributeSettingsDockableFor(defs)
	initTestSettingsContent(&d.undoableSettingsDockable)
	rows := d.content.Children()
	c.Equal(4, len(rows))
	for i, want := range []string{
		"st · ST · Integer", "hp · HP · Pool", "sep · Primary · Primary Separator", "nameless · Integer",
	} {
		parts := rows[i].Children()
		c.Equal(3, len(parts), "row %d: a drag handle, the buttons and the content", i)
		if len(parts) != 3 {
			continue
		}
		_, isHandle := parts[0].Self.(*DragHandle)
		c.True(isHandle, "row %d: the drag handle comes first", i)
		buttons := uxtest.PanelsOfType[*unison.Button](parts[1])
		c.Equal(1, len(buttons), "row %d: only the delete button is offered", i)
		if len(buttons) == 1 {
			c.Equal([]string{"Remove attribute"}, uxtest.LabelTexts(buttons[0].Tooltip), "row %d", i)
		}
		content := parts[2].Children()
		c.Equal(1, len(content), "row %d: the content is the header alone", i)
		if len(content) == 1 {
			c.Equal(role.DisclosureTriangle, content[0].Accessibility.Role, "row %d: the header is a disclosure", i)
			c.Equal(want, content[0].Accessibility.Name, "row %d: the header names the definition", i)
		}
	}
	c.Nil(d.targetMgr.Find(defs.Set["st"].KeyPrefix+"id"), "no field of a collapsed row is registered")
	c.Equal(0, len(uxtest.PanelsOfType[*poolSettingsPanel](d.AsPanel())), "a collapsed pool shows no thresholds")
	c.Equal(0, len(uxtest.PanelsOfType[*StringField](d.AsPanel())), "a collapsed row has no fields")
}

// TestAttributeRowHeaderTogglesTheRow verifies that a click on a row's header, Space, Return or Enter while it has the
// focus, or a screen reader's press or request to expand or collapse it, expands the row to the definition's fields,
// the add threshold button and the thresholds of a pool, and collapses it again, with the header saying which; that
// the focus stays on the header, built anew with the row; and that none of it is an edit, to be undone or applied.
func TestAttributeRowHeaderTogglesTheRow(t *testing.T) {
	type act func(screen *unison.HeadlessScreen, header *unison.Panel, expand bool)
	key := func(code unison.KeyCode) act {
		return func(screen *unison.HeadlessScreen, _ *unison.Panel, _ bool) { screen.KeyPress(code, mod.None) }
	}
	ask := func(action func(expand bool) accessibility.Action) act {
		return func(screen *unison.HeadlessScreen, header *unison.Panel, expand bool) {
			node := screen.AccessibilityNodeFor(header)
			if node == nil {
				return
			}
			screen.PerformAccessibilityAction(accessibility.ActionRequest{Node: node.ID, Action: action(expand)})
		}
	}
	for _, tc := range []struct {
		name  string
		act   act
		focus bool // whether the focus is expected on the header afterward, as a key or an action leaves it
	}{
		{name: "click", act: func(screen *unison.HeadlessScreen, header *unison.Panel, _ bool) {
			screen.Click(screen.PanelCenter(header))
		}},
		{name: "space", act: key(unison.KeySpace), focus: true},
		{name: "return", act: key(unison.KeyReturn), focus: true},
		{name: "enter", act: key(unison.KeyNumPadEnter), focus: true},
		{name: "press", act: ask(func(_ bool) accessibility.Action { return accessibility.Press }), focus: true},
		{name: "expand and collapse", act: ask(func(expand bool) accessibility.Action {
			if expand {
				return accessibility.Expand
			}
			return accessibility.Collapse
		}), focus: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
			screen.EnableAccessibility()
			var d *attributeSettingsDockable
			screen.Do(func() {
				defs := testAttrDefs("st", "hp")
				hp := defs.Set["hp"]
				hp.Type = attribute.Pool
				hp.Thresholds = []*gurps.PoolThreshold{{State: "Reeling"}}
				d = newAttributeSettingsDockable(nil, defs)
				d.TabTitle = "Attributes"
				d.TabIcon = svg.Attributes
				d.show()
			})
			var header *unison.Panel
			screen.Do(func() {
				header = attrRowHeader(t, d, "hp")
				header.RequestFocus()
				c.True(wnd.Focus() == header, "precondition: the header has the focus")
			})
			node := screen.AccessibilityNodeFor(header)
			c.NotNil(node, "the header is described")
			if node != nil {
				c.True(node.Expandable && !node.Expanded, "the header is described as collapsed")
			}
			unchanged := func(when string) {
				c.False(d.undoMgr.CanUndo(), "%s: nothing is undone", when)
				c.False(d.Modified(), "%s: nothing is modified", when)
				c.False(d.applyButton.Enabled(), "%s: there is nothing to apply", when)
			}
			for _, expand := range []bool{true, false} {
				tc.act(screen, header, expand)
				screen.Do(func() {
					c.Equal(expand, attrRowExpanded(d, "hp"), "the row is expanded: %v", expand)
					buttons := uxtest.PanelsOfType[*unison.Button](d.rows()[1].Children()[1])
					pools := uxtest.PanelsOfType[*poolSettingsPanel](d.AsPanel())
					if expand {
						c.Equal(2, len(buttons), "the threshold can be added to while the row is expanded")
						c.Equal(1, len(pools), "the thresholds are shown while the row is expanded")
					} else {
						c.Equal(1, len(buttons), "only deletion is offered while the row is collapsed")
						c.Equal(0, len(pools), "the thresholds are hidden while the row is collapsed")
					}
					header = attrRowHeader(t, d, "hp")
					c.Equal("hp · HP · Pool", header.Accessibility.Name, "the header still names the definition")
					if tc.focus {
						c.True(wnd.Focus() == header, "the focus stays on the header; it is on %T", wnd.Focus())
					}
				})
				node = screen.AccessibilityNodeFor(header)
				c.NotNil(node, "the header is described")
				if node != nil {
					c.Equal(expand, node.Expanded, "the header is described as expanded: %v", expand)
				}
				unchanged(fmt.Sprintf("expanded %v", expand))
			}
		})
	}
}

// TestAttributeRowHeaderNamesTheAttributeAsEdited verifies that a row's header follows the edits made in the row: to
// the ID, to the name, including emptying it, which leaves it out, and to the type, and that undoing an edit does too.
func TestAttributeRowHeaderNamesTheAttributeAsEdited(t *testing.T) {
	c := check.New(t)
	d := newTestAttributeSettingsDockableFor(testAttrDefs("st", "dx"))
	initTestSettingsContent(&d.undoableSettingsDockable)
	askAttrRow(t, d, "st", true)
	prefix := d.model.Set["st"].KeyPrefix
	for _, tc := range []struct {
		name string
		id   string // the ID the definition has after the edit
		edit func()
		want string
	}{
		{name: "id", id: "zz", want: "zz · ST · Integer", edit: func() {
			widgetFor[*StringField](t, d, prefix+"id").SetText("zz")
		}},
		{name: "name", id: "zz", want: "zz · Zest · Integer", edit: func() {
			widgetFor[*StringField](t, d, prefix+"name").SetText("Zest")
		}},
		{name: "no name", id: "zz", want: "zz · Integer", edit: func() {
			widgetFor[*StringField](t, d, prefix+"name").SetText("")
		}},
		{name: "type", id: "zz", want: "zz · Decimal", edit: func() {
			widgetFor[*Popup[attribute.Type]](t, d, prefix+"type").Select(attribute.Decimal)
		}},
		{name: "undo", id: "zz", want: "zz · Integer", edit: func() { d.undoMgr.Undo() }},
	} {
		tc.edit()
		c.Equal(tc.want, attrRowHeader(t, d, tc.id).Accessibility.Name, "%s: the header follows the edit", tc.name)
	}
}

// TestCollapsingAnAttributeRowClosesItsGrantItemEditorsFirst checks that collapsing a pool's row, by its header or by
// the toolbar's toggle, first closes the editor open on the item one of its thresholds grants, asking whether to save
// its changes as applying the settings does, and leaves the row expanded when that is canceled; and that collapsing
// another row asks nothing and leaves the editor open.
func TestCollapsingAnAttributeRowClosesItsGrantItemEditorsFirst(t *testing.T) {
	for _, tc := range []struct {
		name      string
		id        string // the row collapsed, or "" for the toolbar's toggle
		prompts   bool   // whether the editor is expected to ask about its changes
		answer    int    // the response to the prompt
		collapsed bool   // whether the row of hp ends up collapsed
		closed    bool   // whether the editor ends up closed
		item      string // the name of the item the grant then gives
	}{
		{name: "cancel", id: "hp", prompts: true, answer: unison.ModalResponseCancel, item: "Night Vision"},
		{
			name: "save", id: "hp", prompts: true, answer: unison.ModalResponseOK, collapsed: true, closed: true,
			item: "Dark Vision",
		},
		{
			name: "discard", id: "hp", prompts: true, answer: unison.ModalResponseDiscard, collapsed: true, closed: true,
			item: "Night Vision",
		},
		{name: "toolbar cancel", prompts: true, answer: unison.ModalResponseCancel, item: "Night Vision"},
		{name: "another row", id: "st", item: "Night Vision"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
			g := gurps.NewGrant(feature.GrantTrait)
			g.SetItem(newGrantTestTrait("Night Vision", 0), gurps.LibraryFile{})
			isTraitEditor := func(one unison.Dockable) bool {
				_, ok := one.AsPanel().Self.(*editor[*gurps.Trait, *gurps.TraitEditData])
				return ok
			}
			var d *attributeSettingsDockable
			screen.Do(func() {
				defs := testAttrDefs("st", "hp")
				def := defs.Set["hp"]
				def.Type = attribute.Pool
				def.Thresholds = []*gurps.PoolThreshold{{State: "Reeling", Features: gurps.Features{g}}}
				d = newAttributeSettingsDockable(nil, defs)
				d.TabTitle = "Attributes"
				d.TabIcon = svg.Attributes
				d.show()
				expandAllAttributes(d)
				p := thresholdFeaturesPanel(c, d)
				p.editGrantedItem(grantAt(t, p, 0))
			})
			nested := uxtest.SoleEditor[*editor[*gurps.Trait, *gurps.TraitEditData]](t, screen, isTraitEditor)
			screen.Do(func() {
				nested.editorData.Name = "Dark Vision"
				nested.MarkModified(nil)
			})
			screen.Post(func() {
				if tc.id == "" {
					toolbarButtonWithTip(t, d.toolbar, toggleAttributesTip).ClickCallback()
				} else {
					askAttrRow(t, d, tc.id, false)
				}
			})
			if tc.prompts {
				_, dialog := modalDialog(t, screen, wnd)
				screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, tc.answer)))
			} else {
				var dialog *unison.Dialog
				screen.Do(func() {
					for _, w := range unison.Windows() {
						if one, ok := w.ClientData()[unison.DialogClientDataKey].(*unison.Dialog); ok {
							dialog = one
						}
					}
				})
				c.Nil(dialog, "collapsing a row without the editor's grant asks nothing")
				if dialog != nil {
					screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseCancel)))
				}
			}
			screen.Do(func() {
				c.Equal(tc.collapsed, attrRowExpanded(d, "hp") == false, "the pool's row is collapsed: %v", tc.collapsed)
				c.Equal(tc.id == "st", !attrRowExpanded(d, "st"), "only the row asked to collapse without the editor does")
				c.Equal(!tc.closed, len(AllMatchingDockables(isTraitEditor)) == 1, "the editor is open: %v", !tc.closed)
				c.Equal(1, len(unison.Windows()), "no prompt is left up")
				var item string
				if one, ok := d.model.Set["hp"].Thresholds[0].Features[0].(*gurps.Grant); ok {
					item = one.ItemName()
				}
				c.Equal(tc.item, item, "the grant gives the item the answer leaves it")
				c.Equal(tc.item == "Dark Vision", d.applyButton.Enabled(), "saving the item changes the settings")
			})
		})
	}
}

// TestAttributeIDMustBeAnUnusedValidID verifies that an attribute's ID may be changed to one that no other attribute
// has and that is already in the form IDs take, in any case and with surrounding spaces, or kept as it is, but not
// emptied, nor changed to another attribute's ID, a reserved one or one that would need sanitizing.
func TestAttributeIDMustBeAnUnusedValidID(t *testing.T) {
	c := check.New(t)
	d := newTestAttributeSettingsDockableFor(testAttrDefs("st", "dx"))
	initTestSettingsContent(&d.undoableSettingsDockable)
	p := attrDefPanel(d, 0)
	if p == nil || p.def.DefID != "st" {
		t.Fatal("the first row edits ST")
	}
	for _, tc := range []struct {
		id    string
		valid bool
	}{
		{id: "st", valid: true},
		{id: " St ", valid: true},
		{id: "zz", valid: true},
		{id: "dx"},
		{id: ""},
		{id: "z z"},
		{id: "1z"},
		{id: gurps.SkillID},
	} {
		c.Equal(tc.valid, p.validateAttrID(tc.id), "%q is accepted: %v", tc.id, tc.valid)
	}
}

// TestAttributePlacementHiddenUnlessATraitIsPresent verifies that choosing the Hidden placement for an attribute asks
// for the trait that reveals it and the placement it then takes, each edited in the definition, that a pool, which is
// either shown automatically or hidden, is offered nothing else, revealed or not, and that choosing another placement
// asks for neither again.
func TestAttributePlacementHiddenUnlessATraitIsPresent(t *testing.T) {
	c := check.New(t)
	defs := testAttrDefs("st", "hp")
	defs.Set["hp"].Type = attribute.Pool
	defs.Set["hp"].Thresholds = []*gurps.PoolThreshold{{State: "Reeling"}}
	d := newTestAttributeSettingsDockableFor(defs)
	initTestSettingsContent(&d.undoableSettingsDockable)
	expandAllAttributes(d)
	st := d.model.Set["st"]
	hp := d.model.Set["hp"]
	placement := func(def *gurps.AttributeDef) *Popup[attribute.Placement] {
		return widgetFor[*Popup[attribute.Placement]](t, d, def.KeyPrefix+"placement")
	}
	c.Equal(len(attribute.Placements), placement(st).ItemCount(), "every placement is offered")
	c.Equal(2, placement(hp).ItemCount(), "a pool is offered automatic placement and hiding alone")
	c.Nil(d.targetMgr.Find(st.KeyPrefix+"placement_trait"), "no trait is asked for while the attribute is shown")

	placement(st).Select(attribute.Hidden)
	c.Equal(attribute.Hidden, st.Placement, "the attribute is hidden")
	widgetFor[*StringField](t, d, st.KeyPrefix+"placement_trait").SetText("Magery")
	c.Equal("Magery", st.PlacementTrait, "the trait that reveals it is asked for")
	whenPresent := widgetFor[*Popup[attribute.Placement]](t, d, st.KeyPrefix+"placement_when_present")
	c.True(whenPresent.Enabled(), "the placement it is revealed in may be chosen")
	whenPresent.Select(attribute.Secondary)
	c.Equal(attribute.Secondary, st.PlacementWhenPresent, "and is edited in the definition")

	placement(hp).Select(attribute.Hidden)
	whenPresent = widgetFor[*Popup[attribute.Placement]](t, d, hp.KeyPrefix+"placement_when_present")
	c.False(whenPresent.Enabled(), "a revealed pool is shown automatically")
	c.Equal(1, whenPresent.ItemCount(), "and offered nothing else")

	placement(st).Select(attribute.Primary)
	c.Equal(attribute.Primary, st.Placement)
	c.Nil(d.targetMgr.Find(st.KeyPrefix+"placement_trait"), "another placement asks for no trait")
}

// TestAttributeTypeChangeToAPoolGivesItAThreshold verifies that an attribute made a pool is given a threshold to edit,
// while a pool made a pool of the other kind keeps the thresholds it has.
func TestAttributeTypeChangeToAPoolGivesItAThreshold(t *testing.T) {
	for _, tc := range []struct {
		id     string
		typ    attribute.Type
		states []string
	}{
		{id: "st", typ: attribute.Pool, states: []string{""}},
		{id: "st", typ: attribute.PoolRef, states: []string{""}},
		{id: "hp", typ: attribute.PoolRef, states: []string{"Reeling", "Collapse"}},
	} {
		t.Run(tc.id+" to "+tc.typ.Key(), func(t *testing.T) {
			c := check.New(t)
			defs := testAttrDefs("st", "hp")
			defs.Set["hp"].Type = attribute.Pool
			defs.Set["hp"].Thresholds = []*gurps.PoolThreshold{{State: "Reeling"}, {State: "Collapse"}}
			d := newTestAttributeSettingsDockableFor(defs)
			initTestSettingsContent(&d.undoableSettingsDockable)
			expandAllAttributes(d)
			widgetFor[*Popup[attribute.Type]](t, d, d.model.Set[tc.id].KeyPrefix+"type").Select(tc.typ)
			def := d.model.Set[tc.id]
			c.Equal(tc.typ, def.Type, "the popup writes the model")
			c.Equal(len(tc.states), len(def.Thresholds), "the pool has thresholds")
			c.Equal(tc.states, thresholdStates(def), "the pool has thresholds")
			for _, threshold := range def.Thresholds {
				c.NotEqual("", threshold.KeyPrefix, "each threshold has a target key prefix")
				c.NotNil(d.targetMgr.Find(threshold.KeyPrefix+"state"), "each threshold's state field is registered")
			}
		})
	}
}
