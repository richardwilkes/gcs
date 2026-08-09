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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/toolbox/v2/check"
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
	for _, def := range d.defs.List(false) {
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
	restored, exists := d.defs.Clone().Set[added.DefID]
	c.True(exists, "the added attribute must be present in the clone")
	c.Equal(added.KeyPrefix, restored.KeyPrefix, "cloning must preserve the target key prefix")
}

// TestAttributeDeleteButtonEnablement verifies that the "can't delete the last attribute" guard is recomputed whenever
// the panels are added or rebuilt, rather than only when one is deleted. Adding an attribute after deleting down to one
// used to leave the older attribute's delete button disabled forever, and any rebuild (type change, undo, load, reset)
// used to restore an enabled delete button to a lone attribute.
func TestAttributeDeleteButtonEnablement(t *testing.T) {
	c := check.New(t)
	d := newTestAttributeSettingsDockable()
	d.defs = &gurps.AttributeDefs{Set: make(map[string]*gurps.AttributeDef)}
	first := d.addAttributeDef()
	second := d.addAttributeDef()
	initTestAttributeContent(d)
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
	delete(d.defs.Set, third.DefID)
	delete(d.defs.Set, second.DefID)
	d.sync()
	c.Equal(1, len(d.content.Children()), "only the lone attribute should have a panel")
	c.Equal(first.DefID, attrDefPanel(d, 0).def.DefID, "the lone attribute should be the one left in the set")
	c.False(attrDeleteButtonEnabled(d, 0), "a rebuild with one attribute leaves its delete button disabled")

	// A dockable opened with a single attribute starts with the guard in place.
	d2 := newTestAttributeSettingsDockable()
	d2.defs = &gurps.AttributeDefs{Set: make(map[string]*gurps.AttributeDef)}
	d2.addAttributeDef()
	initTestAttributeContent(d2)
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

// initTestAttributeContent builds the dockable's content the way Setup does, inside a scroll panel, since sync() saves
// and restores the scroll position.
func initTestAttributeContent(d *attributeSettingsDockable) {
	content := unison.NewPanel()
	scroller := unison.NewScrollPanel()
	scroller.SetContent(content, behavior.Fill, behavior.Fill)
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

func newTestAttributeSettingsDockable() *attributeSettingsDockable {
	d := &attributeSettingsDockable{defs: gurps.FactoryAttributeDefs()}
	d.Self = d
	d.undoMgr = unison.NewUndoManager(100, func(_ error) {})
	d.targetMgr = NewTargetMgr(d)
	d.defs.ResetTargetKeyPrefixes(d.targetMgr.NextPrefix)
	return d
}
