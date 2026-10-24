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
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestCountPrereqsForSpellUnderNegation verifies what CountPrereqsForSpell counts: a required spell along with its own
// prerequisites, but nothing for a missing list (a typed-nil Prereq), a nil entry, or an "any of" option with nothing
// to check. A negated list counts what it holds turned around: what it rules out isn't counted, traits and other
// prerequisites turn around as spells do, it asks for every one or any one of its children turned around, a script
// under it is left out, a spell it requires brings its own prerequisites along, and a second negation turns it back.
func TestCountPrereqsForSpellUnderNegation(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	ignite := addTestSpell(e, "Ignite Fire", fxp.One)
	light := addTestSpell(e, "Light", fxp.One)
	addSpellNamePrereq(light, "Ignite Fire")
	target := addTestSpell(e, "Target", fxp.One)
	c.Nil(ignite.Prereq, "precondition: Ignite Fire has no prerequisites")
	knowing := func(name string) *SpellPrereq {
		p := NewSpellPrereq()
		p.QualifierCriteria.Qualifier = name
		return p
	}
	notKnowing := func(name string) *SpellPrereq {
		p := knowing(name)
		p.Has = false
		return p
	}
	list := func(all bool, children ...Prereq) *PrereqList {
		p := NewPrereqList()
		p.All = all
		p.Prereqs = children
		return p
	}
	negated := func(all bool, children ...Prereq) *PrereqList {
		p := list(all, children...)
		p.Not = true
		return p
	}
	trait := NewTraitPrereq()
	trait.NameCriteria.Qualifier = "Magery"
	trait.Has = false
	trait.LevelCriteria.Compare = criteria.AnyNumber
	script := NewScriptPrereq()
	script.Script = "true"
	nothingToCheck := list(true, NewPrereqList(), (*PrereqList)(nil), nil)
	for _, one := range []struct {
		name string
		list *PrereqList
		want int
	}{
		{"a spell without prerequisites counts none", nil, 0},
		{"a required spell without prerequisites counts just itself", list(true, knowing("Ignite Fire")), 1},
		{"a required spell brings its own", list(true, knowing("Light")), 2},
		{"a nil entry is skipped", list(true, knowing("Ignite Fire"), nil), 1},
		{
			"an \"any of\" option with nothing to check isn't taken as the cheapest",
			list(false, knowing("Ignite Fire"), NewPrereqList(), nothingToCheck),
			1,
		},
		{"what a negated list rules out isn't counted", negated(true, knowing("Ignite Fire")), 0},
		{"none of none of knowing a spell requires it", negated(true, negated(true, knowing("Ignite Fire"))), 1},
		{"none of not knowing a spell requires it", negated(true, notKnowing("Ignite Fire")), 1},
		{"none of not having a trait requires it", negated(false, trait), 1},
		{"none of not knowing two spells requires both", negated(false, notKnowing("Ignite Fire"), notKnowing("Other")), 2},
		{"not all of not knowing two spells requires either", negated(true, notKnowing("Ignite Fire"), notKnowing("Other")), 1},
		{"a script under a negation isn't counted", negated(false, script), 0},
		{"a spell required through a negation brings its own", negated(false, notKnowing("Light")), 2},
	} {
		target.Prereq = one.list
		c.Equal(one.want, CountPrereqsForSpell(target, e.Spells), one.name)
	}
}
