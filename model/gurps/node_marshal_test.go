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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestMarshalNodeDataEnvelope verifies the envelope every node type writes through marshalNodeData: the data fields
// come first and "calc" last, a nil calc is left out while an empty one that is always wanted is still written, the
// hashed form has no calc at all, and each type keeps its own calc key.
func TestMarshalNodeDataEnvelope(t *testing.T) {
	c := check.New(t)
	marshaled := func(in any) string {
		data, err := jio.Marshal(in)
		c.NoError(err)
		return string(data)
	}

	// A note records its resolved text under its own key, and only when the text resolves to something else.
	note := NewNote(nil, nil, false)
	note.MarkDown = "Owes @who@ a favor"
	c.False(strings.Contains(marshaled(note), `"calc"`), "a note whose text resolves to itself has no calc")
	note.Replacements = map[string]string{"who": "Baron Hale"}
	out := marshaled(note)
	c.True(strings.Contains(out, `"calc":{"resolved_text":"Owes Baron Hale a favor"}`), "resolved text: %s", out)
	c.True(strings.Index(out, `"markdown"`) < strings.Index(out, `"calc"`), "the data fields precede calc: %s", out)
	c.False(strings.Contains(hashedJSON(t, note), `"calc"`), "the hashed form has no calc")
	var restored Note
	c.NoError(jio.Unmarshal([]byte(out), &restored))
	c.Equal(note.MarkDown, restored.MarkDown, "the data fields survive a round-trip")
	c.Equal(note.Replacements, restored.Replacements, "the replacements survive a round-trip")

	// The modifiers share the resolved-notes calc, and likewise drop it when there is nothing to say. Their
	// replacements are those of the item they belong to.
	tm := NewTraitModifier(nil, nil, false)
	tm.LocalNotes = "Only while @when@"
	c.False(strings.Contains(marshaled(tm), `"calc"`), "a trait modifier with unchanged notes has no calc")
	trait := NewTrait(nil, nil, false)
	trait.Replacements = map[string]string{"when": "raining"}
	tm.setTrait(trait)
	out = marshaled(tm)
	c.True(strings.Contains(out, `"calc":{"resolved_notes":"Only while raining"}`), "trait modifier notes: %s", out)
	em := NewEquipmentModifier(nil, nil, false)
	em.LocalNotes = "Only while @when@"
	eqp := NewEquipment(nil, nil, false)
	eqp.Replacements = map[string]string{"when": "raining"}
	em.setEquipment(eqp)
	out = marshaled(em)
	c.True(strings.Contains(out, `"calc":{"resolved_notes":"Only while raining"}`), "equipment modifier notes: %s", out)

	// A trait and a piece of equipment always publish their calc, since it always has values in it.
	trait = NewTrait(nil, nil, false)
	out = marshaled(trait)
	c.True(strings.Contains(out, `"calc":{"points":0}`), "a trait always writes its points: %s", out)
	c.False(strings.Contains(out, `"resolved_notes"`), "unchanged notes are not recorded: %s", out)
	c.False(strings.Contains(hashedJSON(t, trait), `"calc"`), "the hashed trait has no calc")
	eqp = NewEquipment(nil, nil, false)
	out = marshaled(eqp)
	c.True(strings.Contains(out, `"calc":{"value":0,"extended_value":0,"weight":"0 lb","extended_weight":"0 lb"}`),
		"equipment always writes its totals: %s", out)
	c.False(strings.Contains(hashedJSON(t, eqp), `"calc"`), "the hashed equipment has no calc")
}
