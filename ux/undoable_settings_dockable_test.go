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
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestUndoableSettingsButtonsCloseTheirSubEditorsFirst verifies that the Apply and Discard buttons of the attribute
// settings close the editor open on the item a threshold grants before they settle the settings' own changes, since
// that editor edits an object within them: Apply asks whether to save the item's changes and applies them along with
// the rest, Discard asks whether to discard them instead, and canceling at the prompt applies nothing and leaves both
// open, the settings still prompting on close.
func TestUndoableSettingsButtonsCloseTheirSubEditorsFirst(t *testing.T) {
	isTraitEditor := func(one unison.Dockable) bool {
		_, ok := one.AsPanel().Self.(*editor[*gurps.Trait, *gurps.TraitEditData])
		return ok
	}
	for _, tc := range []struct {
		name    string
		apply   bool   // whether Apply is clicked, rather than Discard
		answer  int    // the response given at the item editor's prompt
		closed  bool   // whether the settings and the item editor close
		applied bool   // whether the settings' changes reach the defaults
		item    string // the name of the item the defaults' grant gives afterward
	}{
		{name: "apply", apply: true, answer: unison.ModalResponseOK, closed: true, applied: true, item: "Dark Vision"},
		{name: "apply canceled", apply: true, answer: unison.ModalResponseCancel, item: "Night Vision"},
		{name: "discard", answer: unison.ModalResponseOK, closed: true, item: "Night Vision"},
		{name: "discard canceled", answer: unison.ModalResponseCancel, item: "Night Vision"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
			global := gurps.GlobalSettings()
			t.Cleanup(gurps.SyncGlobalSheetSettings) // Runs after the swap below is undone.
			uxtest.SwapForTest(t, &global.Sheet, global.Sheet.Clone(nil))
			g := gurps.NewGrant(feature.GrantTrait)
			g.SetItem(newGrantTestTrait("Night Vision", 0), gurps.LibraryFile{})
			var d *attributeSettingsDockable
			screen.Do(func() {
				defs := testAttrDefs("st", "hp")
				def := defs.Set["hp"]
				def.Type = attribute.Pool
				def.Thresholds = []*gurps.PoolThreshold{{State: "Reeling", Features: gurps.Features{g}}}
				global.Sheet.Attributes = defs.Clone()
				d = newAttributeSettingsDockable(nil, defs)
				d.TabTitle = "Attributes"
				d.TabIcon = svg.Attributes
				d.show()
				expandAllAttributes(d)
				p := thresholdFeaturesPanel(c, d)
				p.editGrantedItem(grantAt(t, p, 0))
			})
			item := uxtest.SoleEditor[*editor[*gurps.Trait, *gurps.TraitEditData]](t, screen, isTraitEditor)
			button := d.cancelButton
			if tc.apply {
				button = d.applyButton
			}
			screen.Do(func() {
				item.editorData.Name = "Dark Vision"
				item.MarkModified(nil)
				d.model.Set["st"].Name = "Brawn"
				d.MarkModified(nil)
				c.True(button.Enabled(), "precondition: the settings have changes to settle")
			})
			screen.Post(button.ClickCallback)
			_, dialog := modalDialog(t, screen, wnd)
			var offersSave bool
			screen.Do(func() { offersSave = dialog.Button(unison.ModalResponseDiscard) != nil })
			c.Equal(tc.apply, offersSave, "the item's editor asks whether to save its changes, not to discard them: %v",
				tc.apply)
			screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, tc.answer)))
			var open, prompts bool
			var editors int
			var attrName, itemName string
			screen.Do(func() {
				open = slices.ContainsFunc(AllDockables(), func(one unison.Dockable) bool {
					return one.AsPanel().Self == d
				})
				prompts = d.promptForSave
				editors = len(AllMatchingDockables(isTraitEditor))
				attrName = global.Sheet.Attributes.Set["st"].Name
				if one, ok := global.Sheet.Attributes.Set["hp"].Thresholds[0].Features[0].(*gurps.Grant); ok {
					itemName = one.ItemName()
				}
			})
			c.Equal(!tc.closed, open, "the settings stay open: %v", !tc.closed)
			c.True(tc.closed || prompts, "and, if open, still prompt when closed")
			if tc.closed {
				c.Equal(0, editors, "the item's editor closes")
			} else {
				c.Equal(1, editors, "the item's editor stays open")
			}
			if tc.applied {
				c.Equal("Brawn", attrName, "the settings' changes are applied")
			} else {
				c.Equal("ST", attrName, "the settings' changes are not applied")
			}
			c.Equal(tc.item, itemName, "the defaults' grant gives the item as it now is")
		})
	}
}
