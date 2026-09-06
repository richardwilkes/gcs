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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/behavior"
)

// TestAddAttributeDefAssignsUniqueKeyPrefix verifies that an attribute added with the "Add Attribute" toolbar button is
// given its own target key prefix. The panel builds each widget's reference key as the definition's KeyPrefix plus a
// suffix ("id", "name", "type", ...), so a definition added without a prefix produces bare keys that collide with those
// of the next one added. TargetMgr.Find returns the first match, which sent an undo of an edit made in the second new
// attribute to the first one's field, and restored focus to the wrong attribute's widget after a rebuild.
func TestAddAttributeDefAssignsUniqueKeyPrefix(t *testing.T) {
	c := check.New(t)
	d := newTestAttributeSettingsDockable()

	first := d.addAttributeDef()
	second := d.addAttributeDef()
	c.NotEqual("", first.KeyPrefix, "an added attribute must be given a target key prefix")
	c.NotEqual("", second.KeyPrefix, "an added attribute must be given a target key prefix")
	c.NotEqual(first.KeyPrefix, second.KeyPrefix, "two added attributes must not share a target key prefix")
	c.NotEqual(first.DefID, second.DefID, "two added attributes must not share an ID")

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

// TestAddAttributeDefKeyPrefixSurvivesUndoData verifies that the target key prefixes are carried through the clones the
// undo edits hold, since applying one replaces the dockable's definitions and rebuilds the panels from them.
func TestAddAttributeDefKeyPrefixSurvivesUndoData(t *testing.T) {
	c := check.New(t)
	d := newTestAttributeSettingsDockable()
	added := d.addAttributeDef()
	restored, exists := d.model.Clone().Set[added.DefID]
	c.True(exists, "the added attribute must be present in the clone")
	c.Equal(added.KeyPrefix, restored.KeyPrefix, "cloning must preserve the target key prefix")
}

// TestAttributeDeleteButtonEnablement verifies that the "can't delete the last attribute" guard is recomputed whenever
// the panels are added or rebuilt, rather than only when one is deleted. Adding an attribute after deleting down to one
// used to leave the older attribute's delete button disabled forever, and any rebuild (type change, undo, load, reset)
// used to restore an enabled delete button to a lone attribute.
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

// initTestSettingsContent builds a settings dockable's content the way Setup does, but with no dock and no window: the
// content is wrapped in a scroll panel that is itself a child of the dockable, so that sync() can find a scroll root and
// the target manager, which is rooted at the dockable, can find the widgets.
func initTestSettingsContent[T undoableSettingsModel](d *undoableSettingsDockable[T]) {
	content := unison.NewPanel()
	scroller := unison.NewScrollPanel()
	scroller.SetContent(content, behavior.Fill, behavior.Fill)
	d.AddChild(scroller)
	d.initContent(content)
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

// newTestAttributeSettingsDockable returns a dockable for the defaults that edits the factory attribute definitions, its
// content not yet built; see initTestSettingsContent.
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

// TestAttributeDefDragDropReorders verifies that dropping a dragged attribute definition moves it among the definitions,
// renumbers them to match, posts a single undo edit, rebuilds the panels and clears the drag state; and that a payload
// from another attribute editor, which is what a drag from another sheet's settings delivers, is ignored.
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
	pools := panelsOfType[*poolSettingsPanel](d.AsPanel())
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
	pools = panelsOfType[*poolSettingsPanel](d.AsPanel())
	c.Equal(1, len(pools), "the panels are rebuilt")
	c.Equal("Dead", panelsOfType[*thresholdSettingsPanel](pools[0].AsPanel())[0].threshold.State)

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
	rows := panelsOfType[*thresholdSettingsPanel](d.AsPanel())
	c.Equal(1, len(rows))
	c.False(rows[0].deleteButton.Enabled(), "the last threshold may not be deleted")

	rows[0].pool.addThreshold()
	c.Equal(2, len(d.model.Set["hp"].Thresholds), "the threshold is added to the pool")
	c.Equal("Undo Add Pool Threshold", d.undoMgr.UndoTitle())
	rows = panelsOfType[*thresholdSettingsPanel](d.AsPanel())
	c.Equal(2, len(rows), "the rows are rebuilt")
	c.True(rows[0].deleteButton.Enabled(), "with two thresholds, either may be deleted")
	c.True(rows[1].deleteButton.Enabled(), "with two thresholds, either may be deleted")
	c.NotEqual("", rows[1].threshold.KeyPrefix, "the added threshold is given a target key prefix")
	c.NotNil(d.targetMgr.Find(rows[1].threshold.KeyPrefix+"state"), "the added threshold's state field is registered")

	d.undoMgr.Undo()
	c.Equal([]string{"Reeling"}, thresholdStates(d.model.Set["hp"]), "undo removes the added threshold")
	c.False(d.undoMgr.CanUndo(), "the add is a single edit")
	rows = panelsOfType[*thresholdSettingsPanel](d.AsPanel())
	c.Equal(1, len(rows))
	c.False(rows[0].deleteButton.Enabled(), "undoing back to one threshold restores the guard")

	d.undoMgr.Redo()
	c.Equal(2, len(d.model.Set["hp"].Thresholds), "redo adds it back")
	rows = panelsOfType[*thresholdSettingsPanel](d.AsPanel())
	c.Equal(2, len(rows))

	rows[0].pool.deleteThreshold(rows[0])
	c.Equal(1, len(d.model.Set["hp"].Thresholds), "the threshold is deleted")
	c.Equal("", d.model.Set["hp"].Thresholds[0].State, "the first threshold is the one deleted")
	c.Equal("Undo Delete Pool Threshold", d.undoMgr.UndoTitle())
	rows = panelsOfType[*thresholdSettingsPanel](d.AsPanel())
	c.Equal(1, len(rows))
	c.False(rows[0].deleteButton.Enabled(), "deleting down to one threshold restores the guard")

	d.undoMgr.Undo()
	c.Equal(2, len(d.model.Set["hp"].Thresholds), "undo restores the deleted threshold")
	c.Equal("Reeling", d.model.Set["hp"].Thresholds[0].State, "in its original position")
	c.Equal([]string{"st", "hp"}, attrDefIDs(d.model), "the definitions are untouched throughout")
}

// testAttrDefs returns a set of integer attribute definitions with the given IDs, in that order.
func testAttrDefs(ids ...string) *gurps.AttributeDefs {
	defs := &gurps.AttributeDefs{Set: make(map[string]*gurps.AttributeDef)}
	for i, id := range ids {
		defs.Set[id] = testAttrDef(id, strings.ToUpper(id), i+1)
	}
	return defs
}

// attrDefIDs returns the IDs of the definitions in order.
func attrDefIDs(defs *gurps.AttributeDefs) []string {
	list := defs.List(false)
	ids := make([]string, len(list))
	for i, def := range list {
		ids[i] = def.DefID
	}
	return ids
}

// attrDefOrders returns the Order of the definitions in order.
func attrDefOrders(defs *gurps.AttributeDefs) []int {
	list := defs.List(false)
	orders := make([]int, len(list))
	for i, def := range list {
		orders[i] = def.Order
	}
	return orders
}

// thresholdStates returns the states of the definition's thresholds in order.
func thresholdStates(def *gurps.AttributeDef) []string {
	states := make([]string, len(def.Thresholds))
	for i, threshold := range def.Thresholds {
		states[i] = threshold.State
	}
	return states
}
