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
	"fmt"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	uncheck "github.com/richardwilkes/unison/enums/check"
)

// findFeatureTypePopup returns the first feature-type switcher popup found anywhere beneath the given panel, or nil if
// there is none. It is used by the tests to drive a type change the same way a user's popup selection would.
func findFeatureTypePopup(p *unison.Panel) *unison.PopupMenu[feature.Type] {
	if popup, ok := p.Self.(*unison.PopupMenu[feature.Type]); ok {
		return popup
	}
	for _, child := range p.Children() {
		if popup := findFeatureTypePopup(child); popup != nil {
			return popup
		}
	}
	return nil
}

// switchFeatureType finds the type switcher inside the given feature row and invokes its callback to switch to newType,
// mirroring what happens when the user picks a different entry from the popup.
func switchFeatureType(c check.Checker, row *unison.Panel, types []feature.Type, newType feature.Type) {
	popup := findFeatureTypePopup(row)
	c.NotNil(popup, "expected a feature-type switcher in the row")
	index := slices.Index(types, newType)
	c.True(index >= 0, "the new feature type must be present in the switcher's list")
	popup.ChoiceMadeCallback(popup, index, newType)
}

// TestFeaturesPanelSwitchAwayFromSelectorOverride reproduces the bug where switching a feature's type away from the
// SelectorOverride ("Set the value of") entry deleted the entire Features section instead of just replacing that one
// row. The SelectorOverride row hosts its type switcher directly on the base row panel, so the old parent.Parent()
// removal walked all the way up to the features panel and removed everything.
func TestFeaturesPanelSwitchAwayFromSelectorOverride(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)

	override := gurps.NewSelectorOverride(selector.WeaponDamageType)
	override.SetOwner(owner)
	features := gurps.Features{override}

	panel := newFeaturesPanel(entity, owner, &features, false)
	// Give the features panel a parent so that an erroneous RemoveFromParent() on the panel itself would be observable
	// as the whole section vanishing, exactly as it does in a real editor.
	container := unison.NewPanel()
	container.AddChild(panel)

	// Sanity check: the panel holds the add button plus the single SelectorOverride row.
	c.Equal(2, len(panel.Children()), "expected add button + one feature row before the switch")

	row := panel.Children()[1]
	switchFeatureType(c, row, panel.featureTypesList(), feature.WeaponBonus)

	// The features panel must still be attached to its parent (the section did not vanish).
	c.True(slices.Contains(container.Children(), panel.AsPanel()), "the features section must not be removed")

	// The single feature is now a weapon damage bonus, and the panel still shows the add button plus one row.
	c.Equal(1, len(features), "there must still be exactly one feature")
	_, ok := features[0].(*gurps.WeaponBonus)
	c.True(ok, "the feature must have been replaced with a WeaponBonus")
	c.Equal(2, len(panel.Children()), "expected add button + one feature row after the switch")
}

// TestFeaturesPanelSwitchToSelectorOverride verifies the reverse direction: switching an ordinary feature to the
// SelectorOverride type replaces just that row and leaves the section intact.
func TestFeaturesPanelSwitchToSelectorOverride(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)

	bonus := gurps.NewAttributeBonus(gurps.StrengthID)
	bonus.SetOwner(owner)
	features := gurps.Features{bonus}

	panel := newFeaturesPanel(entity, owner, &features, false)
	container := unison.NewPanel()
	container.AddChild(panel)

	row := panel.Children()[1]
	switchFeatureType(c, row, panel.featureTypesList(), feature.SelectorOverride)

	c.Equal(1, len(container.Children()), "the features section must not be removed")
	c.Equal(1, len(features), "there must still be exactly one feature")
	_, ok := features[0].(*gurps.SelectorOverride)
	c.True(ok, "the feature must have been replaced with a SelectorOverride")
	c.Equal(2, len(panel.Children()), "expected add button + one feature row after the switch")
}

// TestFeaturesPanelSwitchMiddleFeature guards the index-based removal: with several features present, switching the
// type of one in the middle must replace only that row and leave the others (and their order) untouched.
func TestFeaturesPanelSwitchMiddleFeature(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)

	first := gurps.NewAttributeBonus(gurps.StrengthID)
	first.SetOwner(owner)
	middle := gurps.NewSelectorOverride(selector.WeaponDamageType)
	middle.SetOwner(owner)
	last := gurps.NewSkillBonus()
	last.SetOwner(owner)
	features := gurps.Features{first, middle, last}

	panel := newFeaturesPanel(entity, owner, &features, false)
	container := unison.NewPanel()
	container.AddChild(panel)

	// Feature at list index 1 lives at child index 2 (child 0 is the add button).
	row := panel.Children()[2]
	switchFeatureType(c, row, panel.featureTypesList(), feature.WeaponBonus)

	c.Equal(1, len(container.Children()), "the features section must not be removed")
	c.Equal(3, len(features), "the feature count must be unchanged")
	c.Equal(gurps.Feature(first), features[0], "the first feature must be untouched")
	c.Equal(gurps.Feature(last), features[2], "the last feature must be untouched")
	_, ok := features[1].(*gurps.WeaponBonus)
	c.True(ok, "the middle feature must have been replaced with a WeaponBonus")
	c.Equal(4, len(panel.Children()), "expected add button + three feature rows after the switch")
}

// switchableCheckBoxes returns every "switchable" checkbox found anywhere beneath the given panel. Like
// findFeatureTypePopup, it lets the tests reach a widget whose position within the row varies from one feature type to
// the next, and returning all of them rather than just the first lets a test insist that a row has exactly one.
func switchableCheckBoxes(p *unison.Panel) []*CheckBox {
	if box, ok := p.Self.(*CheckBox); ok && box.Text.String() == i18n.Text("switchable") {
		return []*CheckBox{box}
	}
	var boxes []*CheckBox
	for _, child := range p.Children() {
		boxes = append(boxes, switchableCheckBoxes(child)...)
	}
	return boxes
}

// findSwitchableCheckBox returns the sole "switchable" checkbox beneath the given panel, or nil if there is none. A row
// carrying more than one is a failure, since the extra would be an unattached duplicate of the same flag.
func findSwitchableCheckBox(c check.Checker, p *unison.Panel) *CheckBox {
	boxes := switchableCheckBoxes(p)
	if len(boxes) == 0 {
		return nil
	}
	c.Equal(1, len(boxes), "a row must not hold more than one switchable checkbox")
	return boxes[0]
}

// clickCheckBox puts the checkbox into the given state and runs its click callback, mirroring what a user's click on it
// does.
func clickCheckBox(box *CheckBox, on bool) {
	box.State = uncheck.FromBool(on)
	box.ClickCallback()
}

// TestFeaturesPanelSwitchableCheckBoxTogglesTheFlag verifies that the "switchable" checkbox on a feature row is wired
// to the feature's own flag in both directions.
func TestFeaturesPanelSwitchableCheckBoxTogglesTheFlag(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)

	bonus := gurps.NewAttributeBonus(gurps.StrengthID)
	bonus.SetOwner(owner)
	features := gurps.Features{bonus}

	panel := newFeaturesPanel(entity, owner, &features, false)
	box := findSwitchableCheckBox(c, panel.Children()[1])
	c.NotNil(box, "expected a switchable checkbox in the attribute bonus row")
	c.Equal(uncheck.Off, box.State, "a feature that isn't switchable must start out unchecked")
	c.False(bonus.IsSwitchable(), "a new attribute bonus must not be switchable")

	clickCheckBox(box, true)
	c.True(bonus.IsSwitchable(), "checking the box must mark the feature as switchable")

	clickCheckBox(box, false)
	c.False(bonus.IsSwitchable(), "clearing the box must mark the feature as not switchable")
}

// TestFeaturesPanelSwitchableCheckBoxReflectsExistingFlag verifies that a feature loaded as switchable shows a checked
// box rather than an empty one.
func TestFeaturesPanelSwitchableCheckBoxReflectsExistingFlag(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)

	bonus := gurps.NewAttributeBonus(gurps.StrengthID)
	bonus.SetOwner(owner)
	bonus.SetSwitchable(true)
	features := gurps.Features{bonus}

	panel := newFeaturesPanel(entity, owner, &features, false)
	box := findSwitchableCheckBox(c, panel.Children()[1])
	c.NotNil(box, "expected a switchable checkbox in the attribute bonus row")
	c.Equal(uncheck.On, box.State, "a switchable feature must show a checked box")
}

// TestFeaturesPanelSwitchPreservesSwitchable verifies that changing a feature's type carries the switchable flag over
// to the replacement feature, so that a user retyping a feature doesn't silently lose the switch that governs it.
func TestFeaturesPanelSwitchPreservesSwitchable(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)

	bonus := gurps.NewAttributeBonus(gurps.StrengthID)
	bonus.SetOwner(owner)
	bonus.SetSwitchable(true)
	features := gurps.Features{bonus}

	panel := newFeaturesPanel(entity, owner, &features, false)
	switchFeatureType(c, panel.Children()[1], panel.featureTypesList(), feature.SkillBonus)

	c.Equal(1, len(features), "there must still be exactly one feature")
	_, ok := features[0].(*gurps.SkillBonus)
	c.True(ok, "the feature must have been replaced with a SkillBonus")
	c.True(features[0].IsSwitchable(), "the replacement feature must still be switchable")

	box := findSwitchableCheckBox(c, panel.Children()[1])
	c.NotNil(box, "expected a switchable checkbox in the replacement row")
	c.Equal(uncheck.On, box.State, "the replacement row's checkbox must show the preserved flag")
}

// TestFeaturesPanelUnknownFeatureHasNoSwitchableCheckBox verifies that a feature this version of GCS doesn't understand
// offers no switchable checkbox. Its raw data is preserved verbatim, so there is nothing here that could be switched
// and any such flag in the data must not be second-guessed.
func TestFeaturesPanelUnknownFeatureHasNoSwitchableCheckBox(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)

	unknown := gurps.NewUnknownFeature("bogus", jsontext.Value(`{"type":"bogus","switchable":true}`))
	features := gurps.Features{unknown}

	panel := newFeaturesPanel(entity, owner, &features, false)
	c.Equal(2, len(panel.Children()), "expected add button + one feature row")
	c.Nil(findSwitchableCheckBox(c, panel.Children()[1]), "an unknown feature must not offer a switchable checkbox")
	c.False(unknown.IsSwitchable(), "an unknown feature is never switchable")
}

// TestFeaturesPanelWeaponSwitchRowPlacesCheckBoxLast guards the layout of the weapon "switch" row, the one row whose
// controls live in a nested two-row wrapper. The switchable checkbox belongs at the end of the wrapper's second row,
// right after the last control, as it is on every other row. As a sibling of the wrapper instead, it was top-aligned
// against the wrapper's two rows and pinned to the far right edge, since the wrapper's column takes up all the slack.
func TestFeaturesPanelWeaponSwitchRowPlacesCheckBoxLast(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)

	bonus := gurps.NewWeaponSwitchBonus()
	bonus.SetOwner(owner)
	features := gurps.Features{bonus}

	panel := newFeaturesPanel(entity, owner, &features, false)
	boxes := switchableCheckBoxes(panel.Children()[1])
	c.Equal(1, len(boxes), "expected exactly one switchable checkbox in the row")
	if len(boxes) != 1 {
		return
	}
	box := boxes[0]
	wrapper := box.Parent()
	c.NotNil(wrapper, "the checkbox must live inside the wrapper holding the switch controls")

	children := wrapper.Children()
	c.Equal(5, len(children), "the wrapper holds the type switcher, the indent spacer, both popups and the checkbox")
	c.True(children[len(children)-1] == box.AsPanel(), "the checkbox must be the last widget in the wrapper")
	_, ok := children[len(children)-2].Self.(*unison.PopupMenu[string])
	c.True(ok, "the checkbox must sit immediately after the bool popup")

	layout, ok := wrapper.Layout().(*unison.FlexLayout)
	c.True(ok, "the wrapper must use a flex layout")
	c.Equal(4, layout.Columns, "the wrapper's second row holds the spacer, both popups and the checkbox")
	data, ok := children[0].LayoutData().(*unison.FlexLayoutData)
	c.True(ok, "the type switcher must carry flex layout data")
	c.Equal(layout.Columns, data.HSpan, "the type switcher must span the whole first row")

	// The wrapper is the only thing on the line, so the line's column count must be 1: anything else beside the wrapper
	// would be shoved aside by the column the wrapper's HGrab expands.
	line := wrapper.Parent()
	c.NotNil(line, "the wrapper must be attached to the line panel")
	c.Equal(1, len(line.Children()), "the wrapper must be the only widget on the line")
	lineLayout, ok := line.Layout().(*unison.FlexLayout)
	c.True(ok, "the line panel must use a flex layout")
	c.Equal(len(line.Children()), lineLayout.Columns, "the line's column count must match its child count")
}

// TestFeaturesPanelSwitchableCheckBoxOnEveryRowType verifies that every feature type that builds its own first row --
// rather than going through the shared leveled-amount line -- gets exactly one switchable checkbox, and that the
// checkbox is wired to that feature.
func TestFeaturesPanelSwitchableCheckBoxOnEveryRowType(t *testing.T) {
	entity := gurps.NewEntity()
	trait := gurps.NewTrait(entity, nil, false)
	equipmentContainer := gurps.NewEquipment(entity, nil, true)
	for _, one := range []struct {
		name    string
		owner   fmt.Stringer
		feature gurps.Feature
	}{
		{name: "weapon switch bonus", owner: trait, feature: gurps.NewWeaponSwitchBonus()},
		{name: "weapon damage bonus", owner: trait, feature: gurps.NewWeaponDamageBonus()},
		{name: "contained weight reduction", owner: equipmentContainer, feature: gurps.NewContainedWeightReduction()},
		{name: "cost reduction", owner: trait, feature: gurps.NewCostReduction(gurps.StrengthID)},
		{name: "selector override", owner: trait, feature: gurps.NewSelectorOverride(selector.WeaponDamageType)},
		{name: "equipment max uses bonus", owner: trait, feature: gurps.NewEquipmentMaxUsesBonus()},
		{name: "trait max level bonus", owner: trait, feature: gurps.NewTraitMaxLevelBonus()},
		{name: "DR bonus", owner: trait, feature: gurps.NewDRBonus()},
		{name: "attribute bonus", owner: trait, feature: gurps.NewAttributeBonus(gurps.StrengthID)},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			if bonus, ok := one.feature.(gurps.Bonus); ok {
				bonus.SetOwner(one.owner)
			}
			features := gurps.Features{one.feature}
			panel := newFeaturesPanel(entity, one.owner, &features, false)
			c.Equal(2, len(panel.Children()), "expected add button + one feature row")
			boxes := switchableCheckBoxes(panel.Children()[1])
			c.Equal(1, len(boxes), "expected exactly one switchable checkbox in the row")
			if len(boxes) != 1 {
				return
			}
			box := boxes[0]
			c.Equal(uncheck.Off, box.State, "the feature starts out not switchable")

			clickCheckBox(box, true)
			c.True(one.feature.IsSwitchable(), "checking the box must mark this feature as switchable")

			clickCheckBox(box, false)
			c.False(one.feature.IsSwitchable(), "clearing the box must mark this feature as not switchable")
		})
	}
}
