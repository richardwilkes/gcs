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
	"strconv"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// newTestThresholdDockable returns an attribute settings dockable, with its toolbar and content built, for the given
// definitions, with the one with the ID hp made a pool whose only threshold holds the features.
func newTestThresholdDockable(defs *gurps.AttributeDefs, features gurps.Features) *attributeSettingsDockable {
	def := defs.Set["hp"]
	def.Type = attribute.Pool
	threshold := &gurps.PoolThreshold{}
	threshold.State = "Reeling"
	threshold.Features = features
	def.Thresholds = []*gurps.PoolThreshold{threshold}
	d := newTestAttributeSettingsDockableFor(defs)
	d.addToStartToolbar(unison.NewPanel())
	initTestSettingsContent(&d.undoableSettingsDockable)
	return d
}

// thresholdFeaturesPanel returns the panel that edits the features of the dockable's only threshold.
func thresholdFeaturesPanel(c check.Checker, d *attributeSettingsDockable) *featuresPanel {
	panels := uxtest.PanelsOfType[*featuresPanel](d.AsPanel())
	c.Equal(1, len(panels), "the threshold has a features panel")
	if len(panels) == 0 {
		return nil
	}
	return panels[0]
}

// TestThresholdFeaturesUndoAfterRebuild checks that adding a feature to a threshold is undone in the model the dockable
// holds once a change to its structure has rebuilt its content, rather than in the threshold the panel that added it
// showed, which leaves the dockable unmodified.
func TestThresholdFeaturesUndoAfterRebuild(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &lastThresholdFeatureTypeUsed, feature.AttributeBonus)
	var d *attributeSettingsDockable
	screen.Do(func() {
		d = newTestThresholdDockable(testAttrDefs("st", "hp"), nil)
		p := thresholdFeaturesPanel(c, d)
		if p == nil {
			return
		}
		add, ok := p.FindRefKey(featureAddKey).Self.(*unison.Button)
		c.True(ok, "the features panel has an add button")
		if ok {
			add.ClickCallback()
		}
	})
	if d == nil {
		return
	}
	thresholds := func() []*gurps.PoolThreshold { return d.model.Set["hp"].Thresholds }
	c.Equal(1, len(thresholds()[0].Features), "the feature is added to the threshold")
	c.Equal("Undo Add Feature", d.undoMgr.UndoTitle())
	c.True(d.applyButton.Enabled(), "and the dockable is marked modified")

	screen.Do(func() { uxtest.PanelsOfType[*poolSettingsPanel](d.AsPanel())[0].addThreshold() })
	c.Equal(2, len(thresholds()), "precondition: a threshold is added, rebuilding the content")
	screen.Do(d.undoMgr.Undo)
	screen.Do(d.undoMgr.Undo)
	c.Equal(1, len(thresholds()))
	c.Equal(0, len(thresholds()[0].Features), "undo removes the feature from the model the dockable holds")
	c.False(d.Modified(), "leaving it as it was opened")
	c.False(d.applyButton.Enabled(), "and marked so")
}

// TestThresholdFeaturesOfferTheEditedAttributes checks that a threshold's features offer the attributes being edited,
// even one the sheet and the defaults lack, with Move, Dodge, Parry and Block, but neither the size modifier nor the
// pools, and that a new attribute bonus names the attribute last chosen for a threshold, or the first offered when that
// one isn't.
func TestThresholdFeaturesOfferTheEditedAttributes(t *testing.T) {
	c := check.New(t)
	d := newTestThresholdDockable(testAttrDefs("st", "zz", "hp"), nil)
	p := thresholdFeaturesPanel(c, d)
	if p == nil {
		return
	}
	choices, _ := p.attributeChoices("", gurps.StrengthID)
	keys := make([]string, 0, len(choices))
	for _, one := range choices {
		keys = append(keys, one.Key)
	}
	c.Equal([]string{gurps.StrengthID, "zz", gurps.MoveID, gurps.DodgeID, gurps.ParryID, gurps.BlockID}, keys)

	uxtest.SwapForTest(t, &lastAttributeIDUsed, gurps.SizeModifierID)
	uxtest.SwapForTest(t, &lastThresholdAttributeIDUsed, "zz")
	for _, want := range []struct {
		last string
		id   string
	}{
		{last: "zz", id: "zz"},
		{last: gurps.SizeModifierID, id: gurps.StrengthID},
	} {
		lastThresholdAttributeIDUsed = want.last
		bonus, ok := p.createFeatureForType(feature.AttributeBonus).(*gurps.AttributeBonus)
		c.True(ok, "an attribute bonus is created")
		if ok {
			c.Equal(want.id, bonus.Attribute, "with %s last chosen", want.last)
		}
	}
}

// TestThresholdFeatureRowKeysStayApartFromAttributeKeys checks that the reference key of a widget of an attribute,
// which starts with the attribute's target key prefix, doesn't find the like-named widget of an open feature row of a
// threshold shown ahead of it, whose path is the row's index.
func TestThresholdFeatureRowKeysStayApartFromAttributeKeys(t *testing.T) {
	c := check.New(t)
	features := make(gurps.Features, 4)
	for i := range features {
		features[i] = gurps.NewTraitBonus()
	}
	d := newTestThresholdDockable(testAttrDefs("hp", "st"), features)
	p := thresholdFeaturesPanel(c, d)
	if p == nil {
		return
	}
	st := d.model.Set[gurps.StrengthID]
	index, err := strconv.Atoi(strings.Trim(st.KeyPrefix, "#:"))
	c.NoError(err, "the target key prefix holds a number")
	c.True(index >= 0 && index < len(features), "precondition: the number %d is the index of a row", index)
	p.setView(rowsView{open: strconv.Itoa(index)})
	c.NotNil(p.FindRefKey(strconv.Itoa(index)+":name"), "precondition: the open row has a name field")
	stPanel := attrDefPanel(d, 1)
	c.True(stPanel != nil && stPanel.def == st, "precondition: the attribute's panel follows the pool's")
	if stPanel == nil {
		return
	}
	field := stPanel.FindRefKey(st.KeyPrefix + "name")
	c.NotNil(field, "precondition: the attribute has a name field")
	c.True(d.targetMgr.Find(st.KeyPrefix+"name") == field, "the attribute's key finds its own name field")
}

// TestThresholdGrantItemEditorFollowsItsThreshold checks that the editor open on the item of a threshold's grant
// follows the threshold across a change that replaces the dockable's model with a copy, so that what it applies lands
// in the model the dockable holds and a click on the grant brings it forward rather than opening another, and that it
// closes, without asking, once the change takes the threshold away, after which undoing the change and then what the
// editor applied restores the item in the model the dockable holds.
func TestThresholdGrantItemEditorFollowsItsThreshold(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(d *attributeSettingsDockable)
		kept   bool // whether the threshold is still there after the change
	}{
		{name: "undo", kept: true, change: func(d *attributeSettingsDockable) {
			uxtest.PanelsOfType[*poolSettingsPanel](d.AsPanel())[0].addThreshold()
			d.undoMgr.Undo()
		}},
		{name: "reset", change: func(d *attributeSettingsDockable) { d.reset() }},
		{name: "delete attribute", change: func(d *attributeSettingsDockable) {
			for _, one := range uxtest.PanelsOfType[*attrDefSettingsPanel](d.AsPanel()) {
				if one.def.DefID == "hp" {
					one.deleteAttrDef()
				}
			}
		}},
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
				p := thresholdFeaturesPanel(c, d)
				p.editGrantedItem(grantAt(t, p, 0))
			})
			// itemName returns the name of the item given by the threshold's grant in the model the dockable holds.
			itemName := func() string {
				var name string
				screen.Do(func() {
					def := d.model.Set["hp"]
					if def == nil || len(def.Thresholds) == 0 || len(def.Thresholds[0].Features) == 0 {
						return
					}
					if one, isGrant := def.Thresholds[0].Features[0].(*gurps.Grant); isGrant {
						name = one.ItemName()
					}
				})
				return name
			}
			nested := uxtest.SoleEditor[*editor[*gurps.Trait, *gurps.TraitEditData]](t, screen, isTraitEditor)
			screen.Do(func() {
				nested.editorData.Name = "Dark Vision"
				nested.apply()
			})
			c.Equal("Dark Vision", itemName(), "precondition: the editor applies to the grant")
			screen.Do(func() {
				nested.editorData.Name = "Blindsight"
				nested.MarkModified(nil)
				tc.change(d)
			})
			var open, windows int
			screen.Do(func() {
				open = len(AllMatchingDockables(isTraitEditor))
				windows = len(unison.Windows())
			})
			if !tc.kept {
				c.Equal(0, open, "the editor closes along with the threshold")
				c.Equal(1, windows, "nothing asks about its changes: %v", wnd.Title())
				screen.Do(d.undoMgr.Undo)
				c.Equal("Dark Vision", itemName(), "precondition: undo brings the threshold back")
				screen.Do(d.undoMgr.Undo)
				c.Equal("Night Vision", itemName(), "undoing what the editor applied restores the item")
				return
			}
			c.Equal(1, open, "the editor stays open on the threshold's grant")
			screen.Do(nested.apply)
			c.Equal("Blindsight", itemName(), "the edit lands in the model the dockable holds")
			screen.Do(func() {
				p := thresholdFeaturesPanel(c, d)
				p.editGrantedItem(grantAt(t, p, 0))
				open = len(AllMatchingDockables(isTraitEditor))
			})
			c.Equal(1, open, "a click on the grant brings its editor forward rather than opening another")
		})
	}
}
