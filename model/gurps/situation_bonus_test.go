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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestSituationBonusPersistsAndHashesThroughEmbedding guards the shared situationBonus embedding: the situation and
// amount must still be written and read under their original keys even though they now live in an unexported
// embedded struct, and both must still participate in the hash.
func TestSituationBonusPersistsAndHashesThroughEmbedding(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name  string
		bonus situationKeyed
		typ   feature.Type
	}{
		{name: "reaction", bonus: NewReactionBonus(), typ: feature.ReactionBonus},
		{name: "conditional modifier", bonus: NewConditionalModifierBonus(), typ: feature.ConditionalModifier},
	} {
		c.Equal(tc.typ, tc.bonus.FeatureType(), "%s: constructor sets the type", tc.name)
		c.Equal(fxp.One, tc.bonus.AdjustedAmount(), "%s: constructor sets the default amount", tc.name)
		c.NotEqual("", tc.bonus.situation(), "%s: constructor sets a default situation", tc.name)

		data, err := jio.Marshal(Features{tc.bonus})
		c.NoError(err, "%s: should marshal", tc.name)
		c.Contains(string(data), `"situation":"`+tc.bonus.situation()+`"`, "%s: situation written under its key", tc.name)
		c.Contains(string(data), `"amount":1`, "%s: amount written under its key", tc.name)

		var loaded Features
		c.NoError(jio.Unmarshal([]byte(`[{"type":"`+tc.typ.Key()+`","situation":"@Target@","amount":3,"per_level":true}]`),
			&loaded), "%s: should load", tc.name)
		c.Equal(1, len(loaded), "%s: a single feature should load", tc.name)
		if len(loaded) != 1 {
			continue
		}
		restored, ok := loaded[0].(situationKeyed)
		c.True(ok, "%s: loads as a situation-keyed bonus", tc.name)
		if !ok {
			continue
		}
		c.Equal(tc.typ, restored.FeatureType(), "%s: type survives the round trip", tc.name)
		c.Equal("@Target@", restored.situation(), "%s: situation survives the round trip", tc.name)

		keys := make(map[string]string)
		restored.FillWithNameableKeys(keys, nil)
		_, present := keys["Target"]
		c.True(present, "%s: nameable keys are extracted from the situation", tc.name)

		before := Hash64(restored)
		clone, ok := restored.Clone().(situationKeyed)
		c.True(ok, "%s: clone keeps the concrete type", tc.name)
		if !ok {
			continue
		}
		c.Equal(before, Hash64(clone), "%s: a clone hashes identically", tc.name)
		switch b := clone.(type) {
		case *ReactionBonus:
			b.Situation = "changed"
		case *ConditionalModifierBonus:
			b.Situation = "changed"
		}
		c.NotEqual(before, Hash64(clone), "%s: the situation participates in the hash", tc.name)
		c.Equal(before, Hash64(restored), "%s: changing the clone leaves the original alone", tc.name)
	}
}

// TestSituationModifiersApplyOwnerReplacements verifies that the shared collector resolves nameable keys in the
// situation using the owning trait's replacements, so that two traits filling "@Target@" in differently produce
// separate entries while two filling it in identically accumulate into one.
func TestSituationModifiersApplyOwnerReplacements(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	newPair := func(amt fxp.Int) (*ReactionBonus, *ConditionalModifierBonus) {
		r := NewReactionBonus()
		r.Situation = "from @Target@"
		r.Amount = amt
		m := NewConditionalModifierBonus()
		m.Situation = "when facing @Target@"
		m.Amount = amt
		return r, m
	}
	r1, m1 := newPair(fxp.One)
	addTraitWithFeatures(e, "Charisma", r1, m1).Replacements = map[string]string{"Target": "elves"}
	r2, m2 := newPair(fxp.Two)
	addTraitWithFeatures(e, "Voice", r2, m2).Replacements = map[string]string{"Target": "elves"}
	r3, m3 := newPair(fxp.Three)
	addTraitWithFeatures(e, "Odious", r3, m3).Replacements = map[string]string{"Target": "dwarves"}
	e.Recalculate()

	totals := func(list []*ConditionalModifier) map[string]fxp.Int {
		result := make(map[string]fxp.Int, len(list))
		for _, one := range list {
			result[one.From] = one.Total()
		}
		return result
	}
	c.Equal(map[string]fxp.Int{"from elves": fxp.Three, "from dwarves": fxp.Three}, totals(e.Reactions()),
		"reactions are keyed by the resolved situation and accumulate per key")
	c.Equal(map[string]fxp.Int{"when facing elves": fxp.Three, "when facing dwarves": fxp.Three},
		totals(e.ConditionalModifiers()),
		"conditional modifiers are keyed by the resolved situation and accumulate per key")
}
