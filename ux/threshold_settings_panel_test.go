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
	expandAllAttributes(d)
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
// pools, that their sentences, collapsed or not, name an attribute as those definitions do, as does the popup of a
// bonus to an attribute the threshold ignores, and that a new attribute bonus names the attribute last chosen for a
// threshold, or the first offered when that one isn't.
func TestThresholdFeaturesOfferTheEditedAttributes(t *testing.T) {
	c := check.New(t)
	d := newTestThresholdDockable(testAttrDefs("st", "zz", "hp"), gurps.Features{gurps.NewAttributeBonus("zz")})
	p := thresholdFeaturesPanel(c, d)
	if p == nil {
		return
	}
	for _, view := range []struct {
		key       string
		text      string
		collapsed bool
	}{
		{key: sectionSummaryKey, text: "+1 to ZZ.", collapsed: true},
		{key: "0" + keySentence, text: "+1 to ZZ"},
	} {
		p.setView(rowsView{collapsed: view.collapsed})
		sentence, ok := p.FindRefKey(view.key).Self.(*sentenceButton)
		c.True(ok, "the panel shows the sentence keyed %s", view.key)
		if ok {
			c.Equal(view.text, sentence.plainText(), "collapsed: %v", view.collapsed)
		}
	}
	choices, _ := p.attributeChoices("", gurps.StrengthID)
	keys := make([]string, 0, len(choices))
	for _, one := range choices {
		keys = append(keys, one.Key)
	}
	c.Equal([]string{gurps.StrengthID, "zz", gurps.MoveID, gurps.DodgeID, gurps.ParryID, gurps.BlockID}, keys)

	*p.features = append(*p.features, gurps.NewAttributeBonus("hp"), gurps.NewAttributeBonus(gurps.SizeModifierID))
	for i, want := range []string{"to HP", "to Size Modifier"} {
		path := strconv.Itoa(i + 1)
		p.setView(rowsView{open: path})
		if popup := featuresPopup[*gurps.AttributeChoice](c, p, path+":attribute"); popup != nil {
			c.Equal(want, popup.Text(), "the threshold ignores the attribute, but names it")
		}
	}

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
				expandAllAttributes(d)
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

// TestThresholdFeaturesKeepTheirViewAcrossRebuilds checks that once the attribute settings rebuild their content, each
// threshold's features are shown as they were, collapsed or not and with the same row open, with the "as a %" of a
// weapon damage bonus still suspended, and that the focus held within one threshold's features stays within that
// threshold's rather than going to the widget with the same key in another's.
func TestThresholdFeaturesKeepTheirViewAcrossRebuilds(t *testing.T) {
	for _, tc := range []struct {
		name    string
		prepare func(d *attributeSettingsDockable) // what precedes the focus, if anything
		rebuild func(d *attributeSettingsDockable)
		focus   bool // whether the focus is expected to stay, rather than going where the rebuild puts it
	}{
		{name: "add threshold", rebuild: func(d *attributeSettingsDockable) {
			uxtest.PanelsOfType[*poolSettingsPanel](d.AsPanel())[0].addThreshold()
		}},
		{name: "undo", focus: true, prepare: func(d *attributeSettingsDockable) {
			uxtest.PanelsOfType[*poolSettingsPanel](d.AsPanel())[0].addThreshold()
		}, rebuild: func(d *attributeSettingsDockable) { d.undoMgr.Undo() }},
		{name: "type change", focus: true, rebuild: func(d *attributeSettingsDockable) {
			attrDefPanel(d, 0).applyAttributeType(attribute.Decimal)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			screen, _ := uxtest.StartHeadlessWorkspace(t, c)
			damage := gurps.NewWeaponBonus(feature.WeaponBonus)
			damage.Percent = true
			var d *attributeSettingsDockable
			screen.Do(func() {
				defs := testAttrDefs("st", "hp")
				def := defs.Set["hp"]
				def.Type = attribute.Pool
				def.Thresholds = []*gurps.PoolThreshold{
					{State: "Reeling", Features: gurps.Features{gurps.NewAttributeBonus(gurps.StrengthID)}},
					{State: "Tiring", Features: gurps.Features{damage}},
				}
				d = newAttributeSettingsDockable(nil, defs)
				d.TabTitle = "Attributes"
				d.TabIcon = svg.Attributes
				d.show()
				expandAllAttributes(d)
			})
			// thresholdFeatures returns the panels of the first and second thresholds' features.
			thresholdFeatures := func() (reeling, tiring *featuresPanel) {
				screen.Do(func() {
					if panels := uxtest.PanelsOfType[*featuresPanel](d.AsPanel()); len(panels) > 1 {
						reeling, tiring = panels[0], panels[1]
					}
				})
				c.NotNil(tiring, "each threshold has a features panel")
				return reeling, tiring
			}
			setAmount := func(p *featuresPanel, text string) {
				screen.Do(func() {
					p.setView(rowsView{open: "0"})
					field, ok := p.FindRefKey("0:amount").Self.(*StringField)
					c.True(ok, "the damage bonus amount is a text field")
					if ok {
						field.SetText(text)
					}
				})
			}
			_, p := thresholdFeatures()
			if p == nil {
				return
			}
			setAmount(p, "+1d")
			c.False(damage.Percent, "precondition: dice suspend the percentage")
			if tc.prepare != nil {
				screen.Do(func() { tc.prepare(d) })
			}
			reeling, p := thresholdFeatures()
			if p == nil {
				return
			}
			screen.Do(func() {
				reeling.setView(rowsView{}) // Expanded, as it doesn't start out.
				p.FindRefKey(sectionToggleKey).RequestFocus()
				tc.rebuild(d)
			})
			if reeling, p = thresholdFeatures(); p == nil {
				return
			}
			screen.Do(func() {
				c.Equal(rowsView{open: "0"}, p.view(), "the second threshold's features are shown as they were")
				c.Equal(rowsView{}, reeling.view(), "and so are the first's")
				if tc.focus {
					focus := d.Window().Focus()
					c.True(focus != nil && focus.RefKey == sectionToggleKey && unison.AncestorIsOrSelf(focus, p.AsPanel()),
						"the focus stays on the second threshold's features title bar")
				}
			})
			setAmount(p, "+10")
			bonus, ok := (*p.features)[0].(*gurps.WeaponBonus)
			c.True(ok && bonus.Percent, "the percentage comes back once the dice go")
		})
	}
}

// TestThresholdFeaturesKeepAPendingCriterionAcrossRebuilds checks that the chip of an optional criterion added to the
// open row of a threshold's features, which holds nothing yet, is still there once the attribute settings rebuild their
// content, with the focus still in it.
func TestThresholdFeaturesKeepAPendingCriterionAcrossRebuilds(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	var d *attributeSettingsDockable
	screen.Do(func() {
		defs := testAttrDefs("st", "hp")
		def := defs.Set["hp"]
		def.Type = attribute.Pool
		def.Thresholds = []*gurps.PoolThreshold{
			{State: "Reeling", Features: gurps.Features{gurps.NewConditionalModifierBonus()}},
		}
		d = newAttributeSettingsDockable(nil, defs)
		d.TabTitle = "Attributes"
		d.TabIcon = svg.Attributes
		d.show()
		expandAllAttributes(d)
		uxtest.PanelsOfType[*poolSettingsPanel](d.AsPanel())[0].addThreshold()
	})
	// inGroupField returns true if the focus is in the group field of the first threshold's open row.
	inGroupField := func() bool {
		var in bool
		screen.Do(func() {
			p := uxtest.PanelsOfType[*featuresPanel](d.AsPanel())[0]
			focus := d.Window().Focus()
			in = focus != nil && focus.RefKey == "0:group" && unison.AncestorIsOrSelf(focus, p.AsPanel())
		})
		return in
	}
	screen.Do(func() {
		p := uxtest.PanelsOfType[*featuresPanel](d.AsPanel())[0]
		p.setView(rowsView{open: "0"})
		add, ok := p.FindRefKey("0:add group").Self.(*unison.Button)
		c.True(ok, "the open row has a button that adds a group")
		if ok {
			add.ClickCallback()
		}
	})
	c.True(inGroupField(), "precondition: the focus is in the group field the button added")
	screen.Do(d.undoMgr.Undo)
	c.True(inGroupField(), "the group field is still there once undo rebuilds the content, with the focus in it")
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
