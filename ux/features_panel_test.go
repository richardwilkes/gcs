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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
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

// TestFeaturesPanelSwitchMiddleFeature guards the index-based removal: with several features present, switching the type
// of one in the middle must replace only that row and leave the others (and their order) untouched.
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
