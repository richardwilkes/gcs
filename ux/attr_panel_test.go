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
)

// TestAttrPanelHashReflectsTraitDrivenVisibility verifies that the panel's hash changes when a trait is added, removed,
// enabled, or disabled in a way that reveals or hides an attribute. Without folding the trait-driven placement into the
// hash, Sync would not rebuild the panel and the revealed/hidden attribute would not update on the sheet.
func TestAttrPanelHashReflectsTraitDrivenVisibility(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()

	// A hidden secondary attribute that becomes visible when the character has the "Magery" trait.
	defs := &gurps.AttributeDefs{Set: map[string]*gurps.AttributeDef{
		"sm": {AttributeDefData: gurps.AttributeDefData{
			DefID:                "sm",
			Type:                 attribute.Integer,
			Base:                 "10",
			Placement:            attribute.Hidden,
			PlacementTrait:       "Magery",
			PlacementWhenPresent: attribute.Secondary,
		}},
	}}

	panel := &AttrPanel{entity: e, kind: gurps.SecondaryAttrKind}
	hidden := panel.computeHash(defs)

	// Adding the matching trait must change the hash so the panel rebuilds and reveals the attribute.
	trait := gurps.NewTrait(e, nil, false)
	trait.Name = "Magery"
	e.Traits = append(e.Traits, trait)
	revealed := panel.computeHash(defs)
	c.NotEqual(hidden, revealed, "adding the trait must change the hash")

	// Disabling the trait must return the hash to the hidden state.
	trait.Disabled = true
	c.Equal(hidden, panel.computeHash(defs), "disabling the trait must restore the hidden hash")

	// Re-enabling it reveals the attribute again.
	trait.Disabled = false
	c.Equal(revealed, panel.computeHash(defs), "re-enabling the trait must restore the revealed hash")

	// Removing the trait returns to the hidden state.
	e.Traits = nil
	c.Equal(hidden, panel.computeHash(defs), "removing the trait must restore the hidden hash")
}
