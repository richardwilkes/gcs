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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellmatch"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// spellBonusLike is what SpellBonus and SpellPointBonus have in common for the purposes of the tests below.
type spellBonusLike interface {
	Bonus
	MatchesSpell(replacements map[string]string, name, powerSource string, colleges, tags []string) bool
}

// TestSpellBonusTypesShareData verifies that a spell bonus and a spell point bonus, which share their persisted data,
// each keep their own feature type through construction, cloning and a JSON round trip, and match spells alike.
func TestSpellBonusTypesShareData(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name  string
		bonus spellBonusLike
		typ   feature.Type
	}{
		{name: "spell bonus", bonus: NewSpellBonus(), typ: feature.SpellBonus},
		{name: "spell point bonus", bonus: NewSpellPointBonus(), typ: feature.SpellPointBonus},
	} {
		c.Equal(tc.typ, tc.bonus.FeatureType(), "%s: constructor sets the type", tc.name)
		c.Equal(fxp.One, tc.bonus.AdjustedAmount(), "%s: constructor sets the default amount", tc.name)
		c.True(tc.bonus.MatchesSpell(nil, "Fireball", "Arcane", []string{"Fire"}, nil),
			"%s: constructor matches all colleges", tc.name)

		clone := tc.bonus.Clone()
		c.Equal(tc.typ, clone.FeatureType(), "%s: clone keeps the type", tc.name)
		c.Equal(any(tc.bonus), any(clone), "%s: clone is equal to the original", tc.name)
		c.True(any(tc.bonus) != any(clone), "%s: clone is a distinct object", tc.name)

		data, err := jio.Marshal(Features{tc.bonus})
		c.NoError(err, "%s: should marshal", tc.name)
		c.Contains(string(data), `"type":"`+tc.typ.Key()+`"`, "%s: type written under its key", tc.name)

		var loaded Features
		c.NoError(jio.Unmarshal([]byte(`[{"type":"`+tc.typ.Key()+`","match":"college_name","name":{"compare":"is","qualifier":"Fire"},"tags":{"compare":"is","qualifier":"Magical"},"amount":2}]`),
			&loaded), "%s: should load", tc.name)
		c.Equal(1, len(loaded), "%s: a single feature should load", tc.name)
		if len(loaded) != 1 {
			continue
		}
		restored, ok := loaded[0].(spellBonusLike)
		c.True(ok, "%s: loads as a spell bonus", tc.name)
		if !ok {
			continue
		}
		c.Equal(tc.typ, restored.FeatureType(), "%s: type survives the round trip", tc.name)
		c.Equal(fxp.Two, restored.AdjustedAmount(), "%s: amount survives the round trip", tc.name)
		c.True(restored.MatchesSpell(nil, "Fireball", "Arcane", []string{"Fire"}, []string{"Magical"}),
			"%s: matches a spell in the named college with the named tag", tc.name)
		c.False(restored.MatchesSpell(nil, "Fireball", "Arcane", []string{"Water"}, []string{"Magical"}),
			"%s: does not match a spell in another college", tc.name)
		c.False(restored.MatchesSpell(nil, "Fireball", "Arcane", []string{"Fire"}, nil),
			"%s: does not match a spell without the named tag", tc.name)
	}

	// The two types differ by nothing but their feature type, and their hashes must reflect only that difference.
	spell := NewSpellBonus()
	points := NewSpellPointBonus()
	c.NotEqual(Hash64(spell), Hash64(points), "the feature type distinguishes the hashes of otherwise equal bonuses")
	spell.SpellMatchType = spellmatch.Name
	spell.NameCriteria = textCriteria(criteria.IsText, "Fireball")
	points.SpellMatchType = spellmatch.Name
	points.NameCriteria = textCriteria(criteria.IsText, "Fireball")
	points.Type = feature.SpellBonus
	c.Equal(Hash64(spell), Hash64(points), "equal data hashes equally regardless of the Go type")
	var nilSpell *SpellBonus
	var nilPoints *SpellPointBonus
	c.Equal(Hash64(nilSpell), Hash64(nilPoints), "nil bonuses hash alike")
	c.NotEqual(Hash64(nilSpell), Hash64(spell), "a nil bonus hashes unlike a real one")
}
