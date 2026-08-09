// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"encoding/json/v2"
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/toolbox/v2/check"
)

func newTraitNamed(e *Entity, name string) *Trait {
	t := NewTrait(e, nil, false)
	t.Name = name
	return t
}

func TestEntityHasTraitNamed(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.Traits = append(e.Traits, newTraitNamed(e, "Magery"))

	c.True(e.HasTraitNamed("Magery"), "exact match")
	c.True(e.HasTraitNamed(" magery "), "case-insensitive and trimmed match")
	c.False(e.HasTraitNamed("Combat Reflexes"), "absent trait")
	c.False(e.HasTraitNamed(""), "empty name never matches")
	c.False((*Entity)(nil).HasTraitNamed("Magery"), "nil entity never matches")

	disabled := newTraitNamed(e, "Sense of Duty")
	disabled.Disabled = true
	e.Traits = append(e.Traits, disabled)
	c.False(e.HasTraitNamed("Sense of Duty"), "disabled traits are ignored")

	// A trait nested under a disabled container is also ignored.
	container := NewTrait(e, nil, true)
	container.Name = "Package"
	container.Disabled = true
	child := newTraitNamed(e, "Racial Memory")
	child.parent = container
	container.Children = append(container.Children, child)
	e.Traits = append(e.Traits, container)
	c.False(e.HasTraitNamed("Racial Memory"), "traits under a disabled container are ignored")
}

func TestAttributeDefEffectivePlacement(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.Traits = append(e.Traits, newTraitNamed(e, "Magery"))

	def := &AttributeDef{AttributeDefData: AttributeDefData{
		DefID: "sm",
		Type:  attribute.Integer,
		Base:  "10",
	}}

	// A plain Hidden placement stays hidden regardless of traits.
	def.Placement = attribute.Hidden
	c.Equal(attribute.Hidden, def.EffectivePlacement(e), "hidden with no trait override")
	c.Equal(attribute.Hidden, def.EffectivePlacement(nil), "hidden with nil entity")

	// With a trait override that matches, the alternate placement is used.
	def.PlacementTrait = "Magery"
	def.PlacementWhenPresent = attribute.Secondary
	c.Equal(attribute.Secondary, def.EffectivePlacement(e), "override applies when trait present")
	c.Equal(attribute.Hidden, def.EffectivePlacement(nil), "override needs an entity")

	// When the trait is absent, it remains hidden.
	def.PlacementTrait = "Combat Reflexes"
	c.Equal(attribute.Hidden, def.EffectivePlacement(e), "override ignored when trait absent")

	// An empty trait name keeps the original (hidden) behavior even if a present-placement is set.
	def.PlacementTrait = ""
	c.Equal(attribute.Hidden, def.EffectivePlacement(e), "empty trait name keeps it hidden")

	// A non-hidden placement never consults the trait override.
	def.Placement = attribute.Primary
	def.PlacementTrait = "Magery"
	c.Equal(attribute.Primary, def.EffectivePlacement(e), "non-hidden placement ignores override")
}

func TestAttributeDefPlacementResolution(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	def := &AttributeDef{AttributeDefData: AttributeDefData{
		DefID:                "sm",
		Type:                 attribute.Integer,
		Base:                 "10",
		Placement:            attribute.Hidden,
		PlacementTrait:       "Magery",
		PlacementWhenPresent: attribute.Primary,
	}}

	// Without the trait, the attribute is hidden, so it is not relevant to any displayed kind. (Relevant is the gate
	// used when laying out the sheet.)
	c.False(def.Relevant(e, PrimaryAttrKind), "hidden: not relevant as primary")
	c.False(def.Relevant(e, SecondaryAttrKind), "hidden: not relevant as secondary")

	// Adding the trait reveals it with the requested placement.
	e.Traits = append(e.Traits, newTraitNamed(e, "Magery"))
	c.True(def.Relevant(e, PrimaryAttrKind), "revealed: relevant as primary")
	c.False(def.Relevant(e, SecondaryAttrKind), "revealed: not relevant as secondary")
	c.True(def.Primary(e), "revealed: primary")
	c.False(def.Secondary(e), "revealed: not secondary")
	c.Equal(PrimaryAttrKind, def.Kind(e), "revealed: primary kind")

	// Switching the alternate placement to Secondary moves it to the secondary section.
	def.PlacementWhenPresent = attribute.Secondary
	c.True(def.Relevant(e, SecondaryAttrKind), "revealed as secondary: relevant as secondary")
	c.False(def.Relevant(e, PrimaryAttrKind), "revealed as secondary: not relevant as primary")
	c.Equal(SecondaryAttrKind, def.Kind(e), "revealed as secondary: secondary kind")
}

// TestAttributeDefsWithNullRows verifies that JSON null entries in an attribute definition list (or in a pool's
// threshold list) are dropped rather than dereferenced. json/v2 decodes a null array entry into a nil pointer without
// error, so a hand-edited or corrupt .attr file — or the SheetSettings.Attributes embedded in a .gcs sheet or a
// settings file — used to crash GCS with a nil-pointer panic instead of loading.
func TestAttributeDefsWithNullRows(t *testing.T) {
	c := check.New(t)

	const data = `{
	"version": 5,
	"rows": [
		{"id":"st","type":"integer","name":"ST","base":"10","cost_per_point":10},
		null,
		{"id":"hp","type":"pool","name":"HP","base":"$st","cost_per_point":2,"thresholds":[
			null,
			{"state":"Dead","value":"-$self*5","ops":["halve_move"]}
		]}
	]
}`
	defs, err := NewAttributeDefsFromFile(fstest.MapFS{"Test.attr": {Data: []byte(data)}}, "Test.attr")
	c.NoError(err, "a file with a null row should load")
	c.Equal(2, len(defs.Set), "the null row should be dropped")
	list := defs.List(false)
	c.Equal("st", list[0].ID(), "first surviving row")
	c.Equal(1, list[0].Order, "orders remain sequential with no gap for the dropped row")
	c.Equal("hp", list[1].ID(), "second surviving row")
	c.Equal(2, list[1].Order, "orders remain sequential with no gap for the dropped row")
	c.Equal(1, len(list[1].Thresholds), "the null threshold should be dropped")
	c.Equal("Dead", list[1].Thresholds[0].State, "the surviving threshold is intact")
	c.Equal("-$hp*5", list[1].Thresholds[0].Value, "$self is still resolved in the surviving threshold")

	// Cloning and hashing must not trip over the sanitized data either.
	c.NotNil(defs.Clone(), "clone works")
	c.Equal(Hash64(defs), Hash64(defs.Clone()), "hash of clone matches")
}

// TestAttributesWithNullRows verifies that a JSON null entry in an entity's attribute list is dropped rather than
// dereferenced while decoding.
func TestAttributesWithNullRows(t *testing.T) {
	c := check.New(t)

	var attrs Attributes
	c.NoError(json.Unmarshal([]byte(`[{"attr_id":"st","adj":2},null,{"attr_id":"dx","adj":1}]`), &attrs),
		"a null attribute entry should not panic or error")
	c.Equal(2, len(attrs.Set), "the null entry should be dropped")
	list := attrs.List()
	c.Equal("st", list[0].ID(), "first surviving attribute")
	c.Equal(0, list[0].Order, "orders remain sequential with no gap for the dropped entry")
	c.Equal("dx", list[1].ID(), "second surviving attribute")
	c.Equal(1, list[1].Order, "orders remain sequential with no gap for the dropped entry")
}
