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
	"bytes"
	"encoding/json/jsontext"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestNodeSyncDataIsInlined verifies that embedding NodeSyncData leaves the on-disk format of each node type as it was
// when the fields were declared directly: the shared members are top-level keys of the node, in their original order,
// rather than nested under an object of their own, and they survive a round trip.
func TestNodeSyncDataIsInlined(t *testing.T) {
	c := check.New(t)
	fill := func(n *NodeSyncData) {
		n.Name = "Name"
		n.PageRef = "B1"
		n.PageRefHighlight = "highlight"
		n.LocalNotes = "notes"
		n.Tags = []string{"one", "two"}
	}
	skill := NewSkill(nil, nil, false)
	fill(&skill.SkillSyncData)
	spell := NewSpell(nil, nil, false)
	fill(&spell.SpellSyncData)
	trait := NewTrait(nil, nil, false)
	fill(&trait.NodeSyncData)
	traitMod := NewTraitModifier(nil, nil, false)
	fill(&traitMod.TraitModifierSyncData)
	eqpMod := NewEquipmentModifier(nil, nil, false)
	fill(&eqpMod.EquipmentModifierSyncData)
	expected := []string{"name", "reference", "reference_highlight", "local_notes", "tags"}
	for _, tc := range []struct {
		name    string
		node    any
		restore func([]byte) (*NodeSyncData, error)
	}{
		{"skill", skill, reloadInto(func(s *Skill) *NodeSyncData { return &s.SkillSyncData })},
		{"spell", spell, reloadInto(func(s *Spell) *NodeSyncData { return &s.SpellSyncData })},
		{"trait", trait, reloadInto(func(tr *Trait) *NodeSyncData { return &tr.NodeSyncData })},
		{"trait modifier", traitMod, reloadInto(func(m *TraitModifier) *NodeSyncData {
			return &m.TraitModifierSyncData
		})},
		{"equipment modifier", eqpMod, reloadInto(func(m *EquipmentModifier) *NodeSyncData {
			return &m.EquipmentModifierSyncData
		})},
	} {
		data, err := jio.Marshal(tc.node)
		c.NoError(err, tc.name)
		keys := topLevelKeys(t, data)
		c.False(slices.ContainsFunc(keys, func(k string) bool { return strings.HasSuffix(k, "SyncData") }),
			"%s: the shared fields must not be nested", tc.name)
		var shared []string
		for _, key := range keys {
			if slices.Contains(expected, key) {
				shared = append(shared, key)
			}
		}
		c.Equal(expected, shared, "%s: the shared fields are top-level keys in declaration order", tc.name)
		restored, err := tc.restore(data)
		c.NoError(err, tc.name)
		c.Equal(NodeSyncData{
			Name:             "Name",
			PageRef:          "B1",
			PageRefHighlight: "highlight",
			LocalNotes:       "notes",
			Tags:             []string{"one", "two"},
		}, *restored, "%s: the shared fields round-trip", tc.name)
	}
}

// topLevelKeys returns the member names of the JSON object in data, in the order they appear.
func topLevelKeys(t *testing.T, data []byte) []string {
	t.Helper()
	c := check.New(t)
	dec := jsontext.NewDecoder(bytes.NewReader(data))
	tok, err := dec.ReadToken()
	c.NoError(err)
	c.Equal(jsontext.BeginObject, tok, "a JSON object")
	var keys []string
	for dec.PeekKind() != '}' {
		tok, err = dec.ReadToken()
		c.NoError(err)
		keys = append(keys, tok.String())
		c.NoError(dec.SkipValue())
	}
	return keys
}
