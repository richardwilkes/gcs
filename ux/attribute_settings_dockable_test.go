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
	"github.com/richardwilkes/toolbox/v2/check"
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

func newTestAttributeSettingsDockable() *attributeSettingsDockable {
	d := &attributeSettingsDockable{defs: gurps.FactoryAttributeDefs()}
	d.Self = d
	d.targetMgr = NewTargetMgr(d)
	d.defs.ResetTargetKeyPrefixes(d.targetMgr.NextPrefix)
	return d
}
