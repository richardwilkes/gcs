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

// TestCountPrereqsForSpellWithNoPrereqs verifies that a spell with no "prereqs" key -- which leaves Prereq as a nil
// *PrereqList and therefore arrives as a typed-nil Prereq interface -- counts as zero rather than panicking.
func TestCountPrereqsForSpellWithNoPrereqs(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	light := addTestSpell(e, "Light", fxp.One)
	c.Nil(light.Prereq, "a freshly created spell must have no prerequisite list")
	count := -1
	c.NotPanics(func() { count = CountPrereqsForSpell(light, e.Spells) },
		"a spell without prerequisites must not panic")
	c.Equal(0, count)
}

// TestCountPrereqsForSpellTraversesIntoSpellWithNoPrereqs verifies that the recursive traversal into a required spell
// tolerates that spell having no prerequisites of its own.
func TestCountPrereqsForSpellTraversesIntoSpellWithNoPrereqs(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTestSpell(e, "Light", fxp.One)
	continualLight := addTestSpell(e, "Continual Light", fxp.One)
	addSpellNamePrereq(continualLight, "Light")
	count := -1
	c.NotPanics(func() { count = CountPrereqsForSpell(continualLight, e.Spells) },
		"traversal into a spell without prerequisites must not panic")
	c.Equal(1, count, "only the required spell itself should be counted")
}

// TestCountPrereqsForSpellChainedPrereqs verifies that the nil guard doesn't short-circuit the normal recursive count.
func TestCountPrereqsForSpellChainedPrereqs(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTestSpell(e, "Ignite Fire", fxp.One)
	light := addTestSpell(e, "Light", fxp.One)
	addSpellNamePrereq(light, "Ignite Fire")
	continualLight := addTestSpell(e, "Continual Light", fxp.One)
	addSpellNamePrereq(continualLight, "Light")
	c.Equal(2, CountPrereqsForSpell(continualLight, e.Spells),
		"both Light and its own Ignite Fire prerequisite should be counted")
}

// TestCountPrereqsForSpellWithNilEntryInList verifies that a nil entry inside a prerequisite list is skipped rather
// than panicking.
func TestCountPrereqsForSpellWithNilEntryInList(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTestSpell(e, "Light", fxp.One)
	continualLight := addTestSpell(e, "Continual Light", fxp.One)
	addSpellNamePrereq(continualLight, "Light")
	continualLight.Prereq.Prereqs = append(continualLight.Prereq.Prereqs, nil)
	count := -1
	c.NotPanics(func() { count = CountPrereqsForSpell(continualLight, e.Spells) },
		"a nil entry within a prerequisite list must not panic")
	c.Equal(1, count)
}

// TestCountPrereqsForSpellSkipsNegatedList verifies that what a negated list holds is not counted, since it is ruled
// out rather than required.
func TestCountPrereqsForSpellSkipsNegatedList(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTestSpell(e, "Light", fxp.One)
	continualLight := addTestSpell(e, "Continual Light", fxp.One)
	addSpellNamePrereq(continualLight, "Light")
	c.Equal(1, CountPrereqsForSpell(continualLight, e.Spells), "precondition: Light is counted")
	continualLight.Prereq.Not = true
	c.Equal(0, CountPrereqsForSpell(continualLight, e.Spells))
}

// TestCountPrereqsForSpellSeesThroughDoubleNegation verifies that a spell required through two negations is counted,
// whether they are two negated lists or a negated list holding a prerequisite not to know it.
func TestCountPrereqsForSpellSeesThroughDoubleNegation(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTestSpell(e, "Light", fxp.One)
	continualLight := addTestSpell(e, "Continual Light", fxp.One)
	light := addSpellNamePrereq(continualLight, "Light")
	c.Equal(1, CountPrereqsForSpell(continualLight, e.Spells), "precondition: Light is counted")
	inner := NewPrereqList()
	inner.Not = true
	inner.Prereqs = Prereqs{light}
	continualLight.Prereq.Prereqs = Prereqs{inner}
	continualLight.Prereq.Not = true
	c.Equal(1, CountPrereqsForSpell(continualLight, e.Spells), "none of none of knowing Light requires it")
	light.Has = false
	continualLight.Prereq.Prereqs = Prereqs{light}
	c.Equal(1, CountPrereqsForSpell(continualLight, e.Spells), "none of not knowing Light requires it")
}

// TestCountPrereqsForSpellUnderNegation verifies how a negated list counts what it holds: traits and other
// prerequisites turn around as spells do, a negated list asks for every one or any one of its children turned around,
// a script under a negation is left out, and a spell required through a negation brings its own prerequisites along.
func TestCountPrereqsForSpellUnderNegation(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	ignite := addTestSpell(e, "Ignite Fire", fxp.One)
	light := addTestSpell(e, "Light", fxp.One)
	addSpellNamePrereq(light, "Ignite Fire")
	target := addTestSpell(e, "Target", fxp.One)
	c.Nil(ignite.Prereq, "precondition: Ignite Fire has no prerequisites")
	notKnowing := func(name string) *SpellPrereq {
		p := NewSpellPrereq()
		p.QualifierCriteria.Qualifier = name
		p.Has = false
		return p
	}
	negated := func(all bool, children ...Prereq) *PrereqList {
		p := NewPrereqList()
		p.All = all
		p.Not = true
		p.Prereqs = children
		return p
	}
	trait := NewTraitPrereq()
	trait.NameCriteria.Qualifier = "Magery"
	trait.Has = false
	trait.LevelCriteria.Compare = criteria.AnyNumber
	script := NewScriptPrereq()
	script.Script = "true"
	for _, one := range []struct {
		name string
		list *PrereqList
		want int
	}{
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

// TestCountPrereqsForSpellSkipsEmptyOption verifies that a list in an "any of" list with nothing to check, whether
// empty or holding only such lists and nil entries, isn't taken as its cheapest option.
func TestCountPrereqsForSpellSkipsEmptyOption(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTestSpell(e, "Light", fxp.One)
	continualLight := addTestSpell(e, "Continual Light", fxp.One)
	addSpellNamePrereq(continualLight, "Light")
	continualLight.Prereq.All = false
	nested := NewPrereqList()
	nested.Prereqs = Prereqs{NewPrereqList(), (*PrereqList)(nil), nil}
	continualLight.Prereq.Prereqs = append(continualLight.Prereq.Prereqs, NewPrereqList(), nested)
	c.Equal(1, CountPrereqsForSpell(continualLight, e.Spells))
}
