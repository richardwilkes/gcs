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
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

// TestAddAttributeDefAssignsUniqueKeyPrefix verifies that an attribute added with the "Add Attribute" toolbar button is
// given its own target key prefix. The panel builds each widget's reference key as the definition's KeyPrefix plus a
// suffix ("id", "name", "type", ...), so a definition added without a prefix produces bare keys that collide with those
// of the next one added. TargetMgr.Find returns the first match, which sent an undo of an edit made in the second new
// attribute to the first one's field, and restored focus to the wrong attribute's widget after a rebuild. The prefix
// must also be carried through the clones the undo edits hold, since applying one replaces the dockable's definitions
// and rebuilds the panels from them.
func TestAddAttributeDefAssignsUniqueKeyPrefix(t *testing.T) {
	c := check.New(t)
	d := newTestAttributeSettingsDockable()

	first := d.addAttributeDef()
	second := d.addAttributeDef()
	c.NotEqual("", first.KeyPrefix, "an added attribute must be given a target key prefix")
	c.NotEqual("", second.KeyPrefix, "an added attribute must be given a target key prefix")
	c.NotEqual(first.KeyPrefix, second.KeyPrefix, "two added attributes must not share a target key prefix")
	c.NotEqual(first.DefID, second.DefID, "two added attributes must not share an ID")
	restored, found := d.model.Clone().Set[second.DefID]
	c.True(found, "the added attribute must be present in the clone")
	c.Equal(second.KeyPrefix, restored.KeyPrefix, "cloning must preserve the target key prefix")

	// The prefixes must also be distinct from those of every pre-existing definition and pool threshold, since all of
	// their widgets live in the same dockable.
	seen := make(map[string]string)
	for _, def := range d.model.List(false) {
		c.NotEqual("", def.KeyPrefix, "attribute %q must have a target key prefix", def.DefID)
		if other, exists := seen[def.KeyPrefix]; exists {
			c.Equal("", def.KeyPrefix, "attributes %q and %q share the target key prefix %q", other, def.DefID,
				def.KeyPrefix)
		}
		seen[def.KeyPrefix] = def.DefID
		for i, threshold := range def.Thresholds {
			if other, exists := seen[threshold.KeyPrefix]; exists {
				c.Equal("", threshold.KeyPrefix, "%q threshold %d and %q share the target key prefix %q", def.DefID, i,
					other, threshold.KeyPrefix)
			}
			seen[threshold.KeyPrefix] = def.DefID
		}
	}
}

// TestAttributeDeleteButtonEnablement verifies that the "can't delete the last attribute" guard is recomputed whenever
// the panels are added or rebuilt, rather than only when one is deleted. Adding an attribute after deleting down to one
// used to leave the older attribute's delete button disabled forever, and any rebuild (type change, undo, load, reset)
// used to restore an enabled delete button to a lone attribute, as did filling its row again to pack a window of its
// own.
func TestAttributeDeleteButtonEnablement(t *testing.T) {
	c := check.New(t)
	d := newTestAttributeSettingsDockableFor(&gurps.AttributeDefs{Set: make(map[string]*gurps.AttributeDef)})
	first := d.addAttributeDef()
	second := d.addAttributeDef()
	initTestSettingsContent(&d.undoableSettingsDockable)
	c.Equal(2, len(d.content.Children()), "both attributes should have a panel")
	c.True(attrDeleteButtonEnabled(d, 0), "with two attributes, the first delete button is enabled")
	c.True(attrDeleteButtonEnabled(d, 1), "with two attributes, the second delete button is enabled")

	// Deleting one leaves a lone attribute that may not be deleted.
	attrDefPanel(d, 1).deleteAttrDef()
	c.Equal(1, len(d.content.Children()), "one attribute panel should remain")
	c.False(attrDeleteButtonEnabled(d, 0), "the last remaining attribute may not be deleted")

	// Adding another attribute with the toolbar button must re-enable the older attribute's delete button.
	third := d.addAttribute("Add Attribute").def
	c.Equal(2, len(d.content.Children()), "the added attribute should have a panel")
	c.True(attrDeleteButtonEnabled(d, 0), "adding an attribute re-enables the older delete button")
	c.True(attrDeleteButtonEnabled(d, 1), "the added attribute may be deleted")

	// A rebuild with a lone attribute must not hand back an enabled delete button.
	delete(d.model.Set, third.DefID)
	delete(d.model.Set, second.DefID)
	d.sync()
	c.Equal(1, len(d.content.Children()), "only the lone attribute should have a panel")
	c.Equal(first.DefID, attrDefPanel(d, 0).def.DefID, "the lone attribute should be the one left in the set")
	c.False(attrDeleteButtonEnabled(d, 0), "a rebuild with one attribute leaves its delete button disabled")

	// Nor must filling the lone row again to pack a window of its own around it expanded.
	d.showForOwnWindow(func() {})
	c.False(attrDeleteButtonEnabled(d, 0), "packing a window around the lone attribute leaves its delete button disabled")

	// A dockable opened with a single attribute starts with the guard in place.
	d2 := newTestAttributeSettingsDockableFor(&gurps.AttributeDefs{Set: make(map[string]*gurps.AttributeDef)})
	d2.addAttributeDef()
	initTestSettingsContent(&d2.undoableSettingsDockable)
	c.False(attrDeleteButtonEnabled(d2, 0), "a dockable opened with one attribute can't delete it")
}

// TestAttributeSettingsTabTitle verifies that the character name is substituted into the tab title rather than being
// built into the string handed to i18n.Text, which would produce a per-character lookup key that no catalog entry can
// ever match.
func TestAttributeSettingsTabTitle(t *testing.T) {
	c := check.New(t)
	i18n.SetLocalizer(func(text string) string {
		if text == "Attributes: %s" {
			return "Attributes of %s"
		}
		return text
	})
	t.Cleanup(func() { i18n.SetLocalizer(nil) })

	entity := gurps.NewEntity()
	entity.Profile.Name = "Bob"
	c.Equal("Attributes of Bob", attributeSettingsTabTitle(&entityPanelForTest{entity: entity}),
		"the translated title must be used, with the name substituted into it")
	c.Equal("Default Attributes", attributeSettingsTabTitle(nil), "a nil owner yields the defaults title")
}

// TestMergeAttributeDefsPreservesImportedOrder verifies that merging imported attribute definitions into the existing
// ones appends the new ones in the order they had in the imported file. The merge used to range over the incoming map,
// whose iteration order is random, so importing more than one new attribute placed them in an arbitrary order in the
// dockable and on the sheet.
func TestMergeAttributeDefsPreservesImportedOrder(t *testing.T) {
	c := check.New(t)
	imported := []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta"}
	expected := append([]string{"st", "dx"}, imported...)
	// Run the merge repeatedly, since a single pass over a randomly ordered map may happen to come out sorted.
	for i := range 20 {
		existing := &gurps.AttributeDefs{Set: make(map[string]*gurps.AttributeDef)}
		existing.Set["st"] = testAttrDef("st", "Old ST", 1)
		existing.Set["dx"] = testAttrDef("dx", "Old DX", 2)
		incoming := &gurps.AttributeDefs{Set: make(map[string]*gurps.AttributeDef)}
		incoming.Set["dx"] = testAttrDef("dx", "New DX", 1)
		for j, id := range imported {
			incoming.Set[id] = testAttrDef(id, id, j+2)
		}

		merged := mergeAttributeDefs(existing, incoming)
		ids := make([]string, 0, len(merged.Set))
		for _, def := range merged.List(false) {
			ids = append(ids, def.DefID)
		}
		c.Equal(expected, ids, "pass %d: imported attributes must keep the order they had in the file", i)
		c.Equal("New DX", merged.Set["dx"].Name, "pass %d: an imported attribute must replace the existing one", i)
		c.Equal(2, merged.Set["dx"].Order, "pass %d: a replaced attribute must keep its existing position", i)
		c.Equal("Old ST", existing.Set["st"].Name, "pass %d: the existing definitions must not be altered", i)
	}
}

func testAttrDef(id, name string, order int) *gurps.AttributeDef {
	def := &gurps.AttributeDef{Order: order}
	def.DefID = id
	def.Name = name
	def.Type = attribute.Integer
	return def
}

// initTestSettingsContent builds a settings dockable's content as Setup does, but with no dock and no window. The
// content is wrapped in a scroll panel below the dockable, so that sync() can find a scroll root and the target
// manager, rooted at the dockable, can find the widgets.
func initTestSettingsContent[T undoableSettingsModel](d *undoableSettingsDockable[T]) {
	content := unison.NewPanel()
	scroller := unison.NewScrollPanel()
	scroller.SetContent(content, behavior.Fill, behavior.Fill)
	d.AddChild(scroller)
	d.initContent(content)
}

// expandAllAttributes expands every attribute's row with the toolbar's toggle, which does so while all of them are
// collapsed, as they are when the content is first built, for a test that reaches into the rows' fields or thresholds.
func expandAllAttributes(d *attributeSettingsDockable) {
	d.toggleAllExpanded()
}

func attrDefPanel(d *attributeSettingsDockable, index int) *attrDefSettingsPanel {
	children := d.content.Children()
	if index >= len(children) {
		return nil
	}
	panel, ok := children[index].Self.(*attrDefSettingsPanel)
	if !ok {
		return nil
	}
	return panel
}

func attrDeleteButtonEnabled(d *attributeSettingsDockable, index int) bool {
	panel := attrDefPanel(d, index)
	return panel != nil && panel.deleteButton.Enabled()
}

// newTestAttributeSettingsDockable returns a dockable for the defaults that edits the factory attribute definitions,
// its content not yet built.
func newTestAttributeSettingsDockable() *attributeSettingsDockable {
	return newTestAttributeSettingsDockableFor(gurps.FactoryAttributeDefs())
}

// newTestAttributeSettingsDockableFor returns a dockable for the defaults that edits the given definitions, its content
// not yet built. The undo manager is replaced by one that panics on an error, so that a test sees it.
func newTestAttributeSettingsDockableFor(defs *gurps.AttributeDefs) *attributeSettingsDockable {
	d := newAttributeSettingsDockable(nil, defs)
	d.undoMgr = unison.NewUndoManager(100, func(err error) { panic(err) })
	return d
}

// TestAttributeDefDragDropReorders verifies that dropping a dragged attribute definition moves it among the
// definitions, renumbers them to match, posts a single undo edit, rebuilds the panels and clears the drag state; and
// that a payload from another attribute editor, which is what a drag from another sheet's settings delivers, is
// ignored.
func TestAttributeDefDragDropReorders(t *testing.T) {
	c := check.New(t)
	d := newTestAttributeSettingsDockableFor(testAttrDefs("st", "dx", "iq"))
	initTestSettingsContent(&d.undoableSettingsDockable)
	rows := d.content.Children()
	c.Equal(3, len(rows))
	dd := dragDataForRow(t, rows[0])
	c.Equal("Attribute Definition Drag", dd.title)

	beginDragOver(&d.rowDragState, d.content, 3)
	d.dataDragDrop(geom.Point{}, dd)
	c.Equal([]string{"dx", "iq", "st"}, attrDefIDs(d.model))
	c.Equal([]int{0, 1, 2}, attrDefOrders(d.model), "the definitions are renumbered to match their new positions")
	c.False(d.inDragOver, "the drag state is cleared")
	c.Equal(-1, d.dragInsert)
	c.Nil(d.dragTarget)
	c.True(d.Modified())
	c.Equal("dx", attrDefPanel(d, 0).def.DefID, "the panels are rebuilt in the new order")

	d.undoMgr.Undo()
	c.Equal([]string{"st", "dx", "iq"}, attrDefIDs(d.model), "undo restores the order")
	c.False(d.undoMgr.CanUndo(), "the drop is a single edit")
	d.undoMgr.Redo()
	c.Equal([]string{"dx", "iq", "st"}, attrDefIDs(d.model))

	// Dropping a definition just below itself leaves everything as it is.
	beginDragOver(&d.rowDragState, d.content, 1)
	d.dataDragDrop(geom.Point{}, dragDataForRow(t, d.content.Children()[0]))
	c.Equal([]string{"dx", "iq", "st"}, attrDefIDs(d.model))
	c.False(d.undoMgr.CanRedo(), "a drop is what was last done")
	d.undoMgr.Undo()
	c.Equal([]string{"st", "dx", "iq"}, attrDefIDs(d.model), "a no-op drop posted no edit, so undo reaches the real drop")

	other := newTestAttributeSettingsDockableFor(testAttrDefs("st", "dx", "iq"))
	initTestSettingsContent(&other.undoableSettingsDockable)
	foreign := dragDataForRow(t, other.content.Children()[2])
	beginDragOver(&d.rowDragState, d.content, 0)
	d.dataDragDrop(geom.Point{}, foreign)
	c.Equal([]string{"st", "dx", "iq"}, attrDefIDs(d.model), "another editor's payload is ignored")
	c.Equal([]string{"st", "dx", "iq"}, attrDefIDs(other.model))
	c.False(d.inDragOver)
}

// TestPoolThresholdDragDropReorders verifies that dropping a dragged pool threshold moves it within the thresholds of
// the pool that owns it, undoably, leaving the definitions themselves in place.
func TestPoolThresholdDragDropReorders(t *testing.T) {
	c := check.New(t)
	defs := testAttrDefs("st", "hp")
	pool := defs.Set["hp"]
	pool.Type = attribute.Pool
	for _, state := range []string{"Reeling", "Collapse", "Dead"} {
		pool.Thresholds = append(pool.Thresholds, &gurps.PoolThreshold{State: state})
	}
	d := newTestAttributeSettingsDockableFor(defs)
	initTestSettingsContent(&d.undoableSettingsDockable)
	expandAllAttributes(d)
	pools := uxtest.PanelsOfType[*poolSettingsPanel](d.AsPanel())
	c.Equal(1, len(pools), "only the pool attribute has a threshold list")
	rows := pools[0].Children()
	c.Equal(3, len(rows))
	dd := dragDataForRow(t, rows[2])
	c.Equal("Pool Threshold Drag", dd.title)

	beginDragOver(&d.rowDragState, pools[0].AsPanel(), 0)
	d.dataDragDrop(geom.Point{}, dd)
	c.Equal([]string{"Dead", "Reeling", "Collapse"}, thresholdStates(d.model.Set["hp"]))
	c.Equal([]string{"st", "hp"}, attrDefIDs(d.model), "the definitions keep their order")
	c.False(d.inDragOver, "the drag state is cleared")
	pools = uxtest.PanelsOfType[*poolSettingsPanel](d.AsPanel())
	c.Equal(1, len(pools), "the panels are rebuilt")
	c.Equal("Dead", uxtest.PanelsOfType[*thresholdSettingsPanel](pools[0].AsPanel())[0].threshold.State)

	d.undoMgr.Undo()
	c.Equal([]string{"Reeling", "Collapse", "Dead"}, thresholdStates(d.model.Set["hp"]), "undo restores the order")
	c.False(d.undoMgr.CanUndo(), "the drop is a single edit")
}

// TestPoolThresholdAddAndDeleteAreUndoable verifies that adding a threshold to a pool and deleting one from it are each
// recorded as a single undo edit of the whole set of definitions, that the rows are rebuilt to match after each step,
// and that the guard against deleting the last threshold follows the count.
func TestPoolThresholdAddAndDeleteAreUndoable(t *testing.T) {
	c := check.New(t)
	defs := testAttrDefs("st", "hp")
	pool := defs.Set["hp"]
	pool.Type = attribute.Pool
	pool.Thresholds = []*gurps.PoolThreshold{{State: "Reeling", KeyPrefix: "r"}}
	d := newTestAttributeSettingsDockableFor(defs)
	initTestSettingsContent(&d.undoableSettingsDockable)
	expandAllAttributes(d)
	rows := uxtest.PanelsOfType[*thresholdSettingsPanel](d.AsPanel())
	c.Equal(1, len(rows))
	c.False(rows[0].deleteButton.Enabled(), "the last threshold may not be deleted")

	rows[0].pool.addThreshold()
	c.Equal(2, len(d.model.Set["hp"].Thresholds), "the threshold is added to the pool")
	c.Equal("Undo Add Pool Threshold", d.undoMgr.UndoTitle())
	rows = uxtest.PanelsOfType[*thresholdSettingsPanel](d.AsPanel())
	c.Equal(2, len(rows), "the rows are rebuilt")
	c.True(rows[0].deleteButton.Enabled(), "with two thresholds, either may be deleted")
	c.True(rows[1].deleteButton.Enabled(), "with two thresholds, either may be deleted")
	c.NotEqual("", rows[1].threshold.KeyPrefix, "the added threshold is given a target key prefix")
	c.NotNil(d.targetMgr.Find(rows[1].threshold.KeyPrefix+"state"), "the added threshold's state field is registered")

	d.undoMgr.Undo()
	c.Equal([]string{"Reeling"}, thresholdStates(d.model.Set["hp"]), "undo removes the added threshold")
	c.False(d.undoMgr.CanUndo(), "the add is a single edit")
	rows = uxtest.PanelsOfType[*thresholdSettingsPanel](d.AsPanel())
	c.Equal(1, len(rows))
	c.False(rows[0].deleteButton.Enabled(), "undoing back to one threshold restores the guard")

	d.undoMgr.Redo()
	c.Equal(2, len(d.model.Set["hp"].Thresholds), "redo adds it back")
	rows = uxtest.PanelsOfType[*thresholdSettingsPanel](d.AsPanel())
	c.Equal(2, len(rows))

	rows[0].pool.deleteThreshold(rows[0])
	c.Equal(1, len(d.model.Set["hp"].Thresholds), "the threshold is deleted")
	c.Equal("", d.model.Set["hp"].Thresholds[0].State, "the first threshold is the one deleted")
	c.Equal("Undo Delete Pool Threshold", d.undoMgr.UndoTitle())
	rows = uxtest.PanelsOfType[*thresholdSettingsPanel](d.AsPanel())
	c.Equal(1, len(rows))
	c.False(rows[0].deleteButton.Enabled(), "deleting down to one threshold restores the guard")

	d.undoMgr.Undo()
	c.Equal(2, len(d.model.Set["hp"].Thresholds), "undo restores the deleted threshold")
	c.Equal("Reeling", d.model.Set["hp"].Thresholds[0].State, "in its original position")
	c.Equal([]string{"st", "hp"}, attrDefIDs(d.model), "the definitions are untouched throughout")
}

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

// TestThresholdFeaturesFollowAttributeEdits verifies that the features of a threshold name an attribute as the
// definitions being edited do once the attribute is renamed, given another ID, deleted or added, none of which rebuilds
// the content: in the summary of those naming it, and in the attribute popup of an open row, even of a feature naming
// another attribute; and that the field edited, which lies outside them, keeps the focus as they are rebuilt.
func TestThresholdFeaturesFollowAttributeEdits(t *testing.T) {
	const ignored = "+1 to zz; a pool threshold ignores it."
	setText := func(c check.Checker, p *attrDefSettingsPanel, key, text string) {
		field, ok := p.FindRefKey(p.def.KeyPrefix + key).Self.(*StringField)
		c.True(ok, "the attribute has a field keyed %s", key)
		if ok {
			field.SetText(text)
		}
	}
	rename := func(c check.Checker, p *attrDefSettingsPanel) { setText(c, p, "name", "Zest") }
	summary := func(c check.Checker, p *featuresPanel) string {
		sentence, ok := p.FindRefKey(sectionSummaryKey).Self.(*sentenceButton)
		c.True(ok, "the collapsed features show a summary")
		if !ok {
			return ""
		}
		return sentence.plainText()
	}
	for _, tc := range []struct {
		name   string
		bonus  string // the attribute the threshold's only feature names
		view   rowsView
		focus  string // the key, after the attribute's prefix, of the field that has the focus during the edit, if any
		edit   func(c check.Checker, p *attrDefSettingsPanel)
		shown  func(c check.Checker, p *featuresPanel) string
		before string
		after  string
	}{
		{
			name: "short name", bonus: "zz", view: rowsView{collapsed: true}, focus: "name", edit: rename,
			shown: summary, before: "+1 to ZZ.", after: "+1 to Zest.",
		},
		{
			name: "id", bonus: "zz", view: rowsView{collapsed: true}, focus: "id", shown: summary, before: "+1 to ZZ.",
			after: ignored, edit: func(c check.Checker, p *attrDefSettingsPanel) { setText(c, p, "id", "zy") },
		},
		{
			name: "delete", bonus: "zz", view: rowsView{collapsed: true}, shown: summary, before: "+1 to ZZ.",
			after: ignored, edit: func(_ check.Checker, p *attrDefSettingsPanel) { p.deleteAttrDef() },
		},
		{
			name: "add", bonus: "a", view: rowsView{collapsed: true}, shown: summary,
			// The new attribute has no name yet, so its ID names it.
			before: "+1 to a; a pool threshold ignores it.", after: "+1 to a.",
			edit: func(_ check.Checker, p *attrDefSettingsPanel) { p.dockable.addAttribute("Add Attribute") },
		},
		{
			name: "open row", bonus: gurps.StrengthID, view: rowsView{open: "0"}, focus: "name", edit: rename,
			shown: func(c check.Checker, p *featuresPanel) string {
				popup := featuresPopup[*gurps.AttributeChoice](c, p, "0:attribute")
				if popup == nil {
					return ""
				}
				titles := make([]string, 0, popup.ItemCount())
				for i := range popup.ItemCount() {
					if item, ok := popup.ItemAt(i); ok {
						titles = append(titles, item.Title)
					}
				}
				return strings.Join(titles, ", ")
			},
			before: "to ST, to ZZ, to Move, to Dodge, to Parry, to Block",
			after:  "to ST, to Zest, to Move, to Dodge, to Parry, to Block",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
			var p *featuresPanel
			var focus *unison.Panel
			screen.Do(func() {
				defs := testAttrDefs("st", "zz", "hp")
				hp := defs.Set["hp"]
				hp.Type = attribute.Pool
				hp.Thresholds = []*gurps.PoolThreshold{
					{State: "Reeling", Features: gurps.Features{gurps.NewAttributeBonus(tc.bonus)}},
				}
				d := newAttributeSettingsDockable(nil, defs)
				d.TabTitle = "Attributes"
				d.TabIcon = svg.Attributes
				d.show()
				expandAllAttributes(d)
				if p = thresholdFeaturesPanel(c, d); p == nil {
					return
				}
				p.setView(tc.view)
				c.Equal(tc.before, tc.shown(c, p), "precondition: the features name the attribute")
				zz := attrDefPanel(d, 1)
				c.True(zz != nil && zz.def.DefID == "zz", "precondition: the second attribute is the one edited")
				if zz == nil {
					return
				}
				if tc.focus != "" {
					if focus = zz.FindRefKey(zz.def.KeyPrefix + tc.focus); focus != nil {
						focus.RequestFocus()
					}
					c.True(focus != nil && wnd.Focus() == focus, "precondition: the field edited has the focus")
				}
				tc.edit(c, zz)
			})
			if p != nil {
				screen.Do(func() {
					c.Equal(tc.after, tc.shown(c, p), "the features follow the edit")
					if focus != nil {
						c.True(wnd.Focus() == focus, "the field edited keeps the focus")
					}
				})
			}
		})
	}
}

// testAttrDefs returns a set of integer attribute definitions with the given IDs, in that order.
func testAttrDefs(ids ...string) *gurps.AttributeDefs {
	defs := &gurps.AttributeDefs{Set: make(map[string]*gurps.AttributeDef)}
	for i, id := range ids {
		defs.Set[id] = testAttrDef(id, strings.ToUpper(id), i+1)
	}
	return defs
}

func attrDefIDs(defs *gurps.AttributeDefs) []string {
	list := defs.List(false)
	ids := make([]string, len(list))
	for i, def := range list {
		ids[i] = def.DefID
	}
	return ids
}

func attrDefOrders(defs *gurps.AttributeDefs) []int {
	list := defs.List(false)
	orders := make([]int, len(list))
	for i, def := range list {
		orders[i] = def.Order
	}
	return orders
}

func thresholdStates(def *gurps.AttributeDef) []string {
	states := make([]string, len(def.Thresholds))
	for i, threshold := range def.Thresholds {
		states[i] = threshold.State
	}
	return states
}

// TestApplyingDefaultAttributesReportsTheChange verifies that applying the default attribute definitions reports the
// change to the display. A library list has no entity, so its difficulty column shows the default definitions' names,
// and it keeps what it derived until a change is reported.
func TestApplyingDefaultAttributesReportsTheChange(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	global := gurps.GlobalSettings()
	t.Cleanup(gurps.SyncGlobalSheetSettings) // Runs after the swap below is undone.
	uxtest.SwapForTest(t, &global.Sheet, global.Sheet.Clone(nil))

	list := NewSkillTableDockable("test"+gurps.SkillsExt, []*gurps.Skill{newTestSkill("Brawling", fxp.One, nil)})
	col := list.table.ColumnIndexForID(gurps.SkillDifficultyColumn)
	c.True(col >= 0, "the skills table must have a difficulty column")
	row := list.table.RowFromIndex(0)
	c.NotNil(row, "the skills table must show the skill")
	difficulty := func() []string {
		var texts []string
		screen.Do(func() {
			texts = uxtest.LabelTexts(row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel())
		})
		return texts
	}
	c.Equal([]string{"DX/A"}, difficulty(), "the difficulty cell must show the attribute's default name")

	screen.Do(func() { ShowAttributeSettings(nil) })
	d := uxtest.SoleEditor[*attributeSettingsDockable](t, screen, func(d unison.Dockable) bool {
		s, ok := d.AsPanel().Self.(*attributeSettingsDockable)
		return ok && s.owner == nil
	})
	screen.Do(func() {
		d.model.Set[gurps.DexterityID].Name = "Dex"
		d.apply()
	})
	c.Equal("Dex", gurps.ResolveAttributeName(nil, gurps.DexterityID), "the defaults must hold the new name")
	c.True(list.table.RowFromIndex(0) == row, "nothing may have synced the table, or this test covers nothing")
	c.Equal([]string{"Dex/A"}, difficulty(), "the same row must show the new name once the defaults are applied")
}

// TestApplyingDefaultAttributesSyncsTheListsWithoutAnEntity verifies that applying the default attribute definitions
// syncs the open templates and library lists, which show the default definitions' names, so that their columns and
// rows are sized for the new names and redrawn with them, and leaves the open loot sheets, which show no attributes,
// alone.
func TestApplyingDefaultAttributesSyncsTheListsWithoutAnEntity(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	global := gurps.GlobalSettings()
	t.Cleanup(gurps.SyncGlobalSheetSettings) // Runs after the swap below is undone.
	uxtest.SwapForTest(t, &global.Sheet, global.Sheet.Clone(nil))

	data := gurps.NewTemplate()
	data.Skills = []*gurps.Skill{newTestSkill("Brawling", fxp.One, nil)}
	var template *Template
	var list *TableDockable[*gurps.Skill]
	var loot *LootSheet
	screen.Do(func() {
		template = newTestTemplateDockable("Template", data)
		DisplayNewDockable(template)
		template.markUnmodified() // So that closing it asks nothing.
		list = NewSkillTableDockable(filepath.Join(t.TempDir(), "test"+gurps.SkillsExt),
			[]*gurps.Skill{newTestSkill("Brawling", fxp.One, nil)})
		DisplayNewDockable(list)
		loot = NewLootSheet("test"+gurps.LootExt, gurps.NewLoot())
		DisplayNewDockable(loot)
		loot.markUnmodified()
	})
	c.NotNil(template, "the template must open")
	c.NotNil(list, "the library list must open")
	c.NotNil(loot, "the loot sheet must open")
	if template == nil || list == nil || loot == nil {
		return
	}
	row := list.table.RowFromIndex(0)
	c.NotNil(row, "the library list must show the skill")
	lootStamp := loot.Equipment.syncedAtStamp()

	screen.Do(func() { ShowAttributeSettings(nil) })
	d := uxtest.SoleEditor[*attributeSettingsDockable](t, screen, func(d unison.Dockable) bool {
		s, ok := d.AsPanel().Self.(*attributeSettingsDockable)
		return ok && s.owner == nil
	})
	before := currentDisplayStamp()
	screen.Do(func() {
		d.model.Set[gurps.DexterityID].Name = "Dexterity"
		d.apply()
	})
	c.True(template.Skills.syncedAtStamp() > before, "the template's skills list must be synced")
	c.True(list.table.RowFromIndex(0) != row, "the library list must be synced")
	c.Equal(lootStamp, loot.Equipment.syncedAtStamp(), "the loot sheet shows no attributes, so it must be left alone")
}

// attrRowHeader returns the header of the row of the definition with the ID: the disclosure that heads the row's
// content, failing the test when the row has none.
func attrRowHeader(t *testing.T, d *attributeSettingsDockable, id string) *unison.Panel {
	t.Helper()
	for _, row := range d.rows() {
		if row.def.DefID != id {
			continue
		}
		if parts := row.Children(); len(parts) == 3 {
			if content := parts[2].Children(); len(content) != 0 &&
				content[0].Accessibility.Role == role.DisclosureTriangle {
				return content[0]
			}
		}
		t.Fatalf("the row of %q has no header", id)
	}
	t.Fatalf("no row shows %q", id)
	return nil
}

// askAttrRow asks the header of the row of the definition with the ID to expand or collapse the row, as a screen
// reader would.
func askAttrRow(t *testing.T, d *attributeSettingsDockable, id string, expand bool) {
	t.Helper()
	action := accessibility.Collapse
	if expand {
		action = accessibility.Expand
	}
	attrRowHeader(t, d, id).Accessibility.ActionCallback(accessibility.ActionRequest{Action: action})
}

// attrRowExpanded reports whether the row of the definition with the ID shows its fields, by the ID field being
// registered with the dockable.
func attrRowExpanded(d *attributeSettingsDockable, id string) bool {
	return d.targetMgr.Find(d.model.Set[id].KeyPrefix+"id") != nil
}

// toolbarButtonWithTip returns the toolbar's button with the tooltip, failing the test if there is none.
func toolbarButtonWithTip(t *testing.T, toolbar *unison.Panel, tip string) *unison.Button {
	t.Helper()
	for _, b := range uxtest.PanelsOfType[*unison.Button](toolbar) {
		if b.Tooltip != nil && slices.Contains(uxtest.LabelTexts(b.Tooltip), tip) {
			return b
		}
	}
	t.Fatalf("the toolbar has no button with the tooltip %q", tip)
	return nil
}

const toggleAttributesTip = "Expands/collapses all attribute definitions"

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

// TestAttributeRowExpansionSurvivesRebuildsButNotReplacement verifies that an expanded row stays expanded, and the
// collapsed ones collapsed, when the content is rebuilt from the same definitions, as a type change, undo and redo,
// a drag, adding a threshold and bringing back a deleted definition do, while replacing the definitions, as resetting
// and loading do, collapses every row, even one whose definition has the same ID as before.
func TestAttributeRowExpansionSurvivesRebuildsButNotReplacement(t *testing.T) {
	changeType := func(d *attributeSettingsDockable) { attrDefPanel(d, 0).applyAttributeType(attribute.Decimal) }
	for _, tc := range []struct {
		name         string
		change       func(d *attributeSettingsDockable)
		allCollapsed bool
	}{
		{name: "type change", change: changeType},
		{name: "undo", change: func(d *attributeSettingsDockable) {
			changeType(d)
			d.undoMgr.Undo()
		}},
		{name: "redo", change: func(d *attributeSettingsDockable) {
			changeType(d)
			d.undoMgr.Undo()
			d.undoMgr.Redo()
		}},
		{name: "drag", change: func(d *attributeSettingsDockable) {
			d.reorderRows("Drag", func() bool { return d.moveAttributeDef(d.model.Set["hp"], 0) })
		}},
		{name: "add threshold", change: func(d *attributeSettingsDockable) {
			uxtest.PanelsOfType[*poolSettingsPanel](d.AsPanel())[0].addThreshold()
		}},
		{name: "delete and undo", change: func(d *attributeSettingsDockable) {
			attrDefPanel(d, 2).deleteAttrDef()
			d.undoMgr.Undo()
		}},
		{name: "reset", allCollapsed: true, change: (*attributeSettingsDockable).reset},
		{name: "load", allCollapsed: true, change: func(d *attributeSettingsDockable) {
			d.replaceModel("Load Attributes", d.model.Clone())
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			defs := testAttrDefs("st", "dx", "hp")
			hp := defs.Set["hp"]
			hp.Type = attribute.Pool
			hp.Thresholds = []*gurps.PoolThreshold{{State: "Reeling"}}
			d := newTestAttributeSettingsDockableFor(defs)
			initTestSettingsContent(&d.undoableSettingsDockable)
			askAttrRow(t, d, "hp", true)
			c.True(attrRowExpanded(d, "hp"), "precondition: the pool's row is expanded")
			c.False(attrRowExpanded(d, "st"), "precondition: the other rows are collapsed")

			tc.change(d)
			c.NotNil(d.model.Set["hp"], "precondition: a definition with the ID is still there")
			c.Equal(!tc.allCollapsed, attrRowExpanded(d, "hp"), "the pool's row is expanded after the change")
			c.False(attrRowExpanded(d, "st"), "the first row is collapsed after the change")
			c.False(attrRowExpanded(d, "dx"), "the second row is collapsed after the change")
		})
	}
}

// TestAttributeToolbarToggleExpandsOnlyWhenAllAreCollapsed verifies that the toolbar's toggle expands every row while
// all of them are collapsed, and otherwise collapses them all, whichever rows are expanded, and that doing so is no
// edit.
func TestAttributeToolbarToggleExpandsOnlyWhenAllAreCollapsed(t *testing.T) {
	for _, tc := range []struct {
		name     string
		expanded []string
		expands  bool
	}{
		{name: "none expanded", expands: true},
		{name: "first expanded", expanded: []string{"st"}},
		{name: "another expanded", expanded: []string{"hp"}},
		{name: "all expanded", expanded: []string{"st", "dx", "hp"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			d := newTestAttributeSettingsDockableFor(testAttrDefs("st", "dx", "hp"))
			d.addToStartToolbar(unison.NewPanel())
			initTestSettingsContent(&d.undoableSettingsDockable)
			for _, id := range tc.expanded {
				askAttrRow(t, d, id, true)
			}
			toolbarButtonWithTip(t, d.toolbar, toggleAttributesTip).ClickCallback()
			for _, id := range []string{"st", "dx", "hp"} {
				c.Equal(tc.expands, attrRowExpanded(d, id), "%s is expanded", id)
			}
			c.False(d.undoMgr.CanUndo(), "nothing is undone")
			c.False(d.Modified(), "nothing is modified")
		})
	}
}

// TestAddedAttributeStartsExpandedWithFocusInItsID verifies that a definition added with the toolbar's button, alone
// among the rows, starts out expanded, with the focus in its ID field, since it is there to be edited.
func TestAddedAttributeStartsExpandedWithFocusInItsID(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	screen.Do(func() {
		d := newAttributeSettingsDockable(nil, testAttrDefs("st", "dx"))
		d.TabTitle = "Attributes"
		d.TabIcon = svg.Attributes
		d.show()
		toolbarButtonWithTip(t, d.toolbar, "Add Attribute").ClickCallback()
		rows := d.rows()
		c.Equal(3, len(rows), "the definition has a row")
		if len(rows) != 3 {
			return
		}
		added := rows[2].def
		c.True(attrRowExpanded(d, added.DefID), "the added definition's row is expanded")
		c.True(wnd.Focus() == d.targetMgr.Find(added.KeyPrefix+"id"), "the focus is in its ID field; it is on %T",
			wnd.Focus())
		c.False(attrRowExpanded(d, "st"), "the other rows stay collapsed")
		c.False(attrRowExpanded(d, "dx"), "the other rows stay collapsed")
	})
}

// TestCollapsingAnAttributeRowMovesTheFocusToItsHeader verifies that collapsing a row while one of its fields has the
// focus, by its header or by the toolbar's toggle, hands the focus straight to the row's header rather than leaving it
// on a field no longer shown. Collapsing a later row as well used to hand the focus to the window's first tab stop on
// the way, since asking where the focus was seeded it afresh once the first row's field had gone.
func TestCollapsingAnAttributeRowMovesTheFocusToItsHeader(t *testing.T) {
	for _, tc := range []struct {
		name     string
		collapse func(t *testing.T, d *attributeSettingsDockable)
	}{
		{name: "header", collapse: func(t *testing.T, d *attributeSettingsDockable) { askAttrRow(t, d, "st", false) }},
		{name: "toolbar", collapse: func(t *testing.T, d *attributeSettingsDockable) {
			toolbarButtonWithTip(t, d.toolbar, toggleAttributesTip).ClickCallback()
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
			screen.Do(func() {
				d := newAttributeSettingsDockable(nil, testAttrDefs("st", "dx"))
				d.TabTitle = "Attributes"
				d.TabIcon = svg.Attributes
				d.show()
				askAttrRow(t, d, "st", true)
				askAttrRow(t, d, "dx", true)
				field := d.targetMgr.Find(d.model.Set["st"].KeyPrefix + "name")
				c.NotNil(field, "precondition: the expanded row has a name field")
				if field == nil {
					return
				}
				field.RequestFocus()
				c.True(wnd.Focus() == field, "precondition: the field has the focus")
				var focused []*unison.Panel
				wnd.Content().FocusChangeInHierarchyCallback = func(_, to *unison.Panel) { focused = append(focused, to) }
				tc.collapse(t, d)
				wnd.Content().FocusChangeInHierarchyCallback = nil
				c.False(attrRowExpanded(d, "st"), "the row is collapsed")
				header := attrRowHeader(t, d, "st")
				c.True(wnd.Focus() == header, "the focus is on the row's header; it is on %T", wnd.Focus())
				c.NotEqual(0, len(focused), "precondition: the focus changes are seen")
				for _, one := range focused {
					c.True(one == header, "the focus goes nowhere but the header; it went to %T", one.Self)
				}
			})
		})
	}
}

// TestUndoReachesAnEditMadeInACollapsedAttributeRow verifies that undoing an edit made in a row that has since been
// collapsed, of a field of the definition or of the features of one of its thresholds, expands the row and lands in
// the definitions the dockable holds, even once another edit has replaced them with a copy, and leaves the dockable
// as unmodified as the undo makes it.
func TestUndoReachesAnEditMadeInACollapsedAttributeRow(t *testing.T) {
	rename := func(t *testing.T, d *attributeSettingsDockable) {
		t.Helper()
		askAttrRow(t, d, "st", true)
		widgetFor[*StringField](t, d, d.model.Set["st"].KeyPrefix+"name").SetText("Brawn")
		askAttrRow(t, d, "st", false)
	}
	for _, tc := range []struct {
		name string
		id   string // the row edited
		edit func(t *testing.T, d *attributeSettingsDockable)
		undo func(d *attributeSettingsDockable)
	}{
		{name: "field", id: "st", edit: rename, undo: func(d *attributeSettingsDockable) { d.undoMgr.Undo() }},
		{
			name: "field after a copy", id: "st",
			edit: func(t *testing.T, d *attributeSettingsDockable) {
				rename(t, d)
				attrDefPanel(d, 1).applyAttributeType(attribute.Decimal) // Replaces the definitions with a copy.
			},
			undo: func(d *attributeSettingsDockable) {
				d.undoMgr.Undo()
				d.undoMgr.Undo()
			},
		},
		{
			name: "features", id: "hp",
			edit: func(t *testing.T, d *attributeSettingsDockable) {
				askAttrRow(t, d, "hp", true)
				p := uxtest.PanelsOfType[*featuresPanel](d.AsPanel())[0]
				add, ok := p.FindRefKey(featureAddKey).Self.(*unison.Button)
				if ok {
					add.ClickCallback()
				}
				askAttrRow(t, d, "hp", false)
			},
			undo: func(d *attributeSettingsDockable) { d.undoMgr.Undo() },
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			screen, _ := uxtest.StartHeadlessWorkspace(t, c)
			screen.Do(func() {
				defs := testAttrDefs("st", "dx", "hp")
				hp := defs.Set["hp"]
				hp.Type = attribute.Pool
				hp.Thresholds = []*gurps.PoolThreshold{{State: "Reeling"}}
				d := newAttributeSettingsDockable(nil, defs)
				d.TabTitle = "Attributes"
				d.TabIcon = svg.Attributes
				d.show()
				before := gurps.Hash64(d.model)
				tc.edit(t, d)
				c.True(d.undoMgr.CanUndo(), "precondition: the edit is recorded")
				c.NotEqual(before, gurps.Hash64(d.model), "precondition: the edit changes the definitions")
				c.True(d.applyButton.Enabled(), "precondition: the edit is there to be applied")
				c.False(attrRowExpanded(d, tc.id), "precondition: the row is collapsed")

				tc.undo(d)
				c.Equal(before, gurps.Hash64(d.model), "undo restores the definitions the dockable holds")
				c.True(attrRowExpanded(d, tc.id), "undo expands the row it reaches into")
				c.False(d.applyButton.Enabled(), "the dockable shows as unmodified")
				c.Equal("st · ST · Integer", attrRowHeader(t, d, "st").Accessibility.Name,
					"the header names the definition as restored")
			})
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

// TestAttributeSettingsWindowFitsAnExpandedRow verifies that the window the attribute settings are given when they are
// taken out of the workspace has room for a typical row expanded. The rows start out collapsed, and the window used to
// be packed around their headers alone, leaving too little room to edit a definition in.
func TestAttributeSettingsWindowFitsAnExpandedRow(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	screen.Do(func() { ShowAttributeSettings(nil) })
	d := uxtest.SoleEditor[*attributeSettingsDockable](t, screen, func(d unison.Dockable) bool {
		s, ok := d.AsPanel().Self.(*attributeSettingsDockable)
		return ok && s.owner == nil
	})
	screen.Do(func() {
		wnd, err := MoveDockableToWindow(d)
		c.NoError(err)
		c.NotNil(wnd, "the settings are given a window of their own")
		if wnd == nil {
			return
		}
		c.False(attrRowExpanded(d, gurps.StrengthID), "precondition: the rows are collapsed")
		askAttrRow(t, d, gurps.StrengthID, true)
		_, pref, _ := d.Sizes(geom.Size{})
		content := wnd.ContentRect()
		c.True(pref.Width <= content.Width, "the window is wide enough for a row expanded: %v within %v", pref, content)
		c.True(pref.Height <= content.Height, "the window is tall enough for a row expanded: %v within %v", pref,
			content)
		c.True(d.AttemptClose(), "the window closes with the unmodified settings")
	})
}
