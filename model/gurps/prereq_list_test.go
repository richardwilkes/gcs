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
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// TestPrereqListSkipsListsThatDoNotApply verifies that a list whose tech level condition doesn't match the sheet's, or
// that has nothing in it that applies, is left out of its parent's check, and that one left out at the top is met.
func TestPrereqListSkipsListsThatDoNotApply(t *testing.T) {
	e := NewEntity()
	e.Profile.TechLevel = "3"
	list := func(all bool, children ...Prereq) *PrereqList {
		p := NewPrereqList()
		p.All = all
		p.Prereqs = children
		return p
	}
	skipped := func(children ...Prereq) *PrereqList {
		p := list(true, children...)
		p.WhenTL = numberCriteria(criteria.AtLeastNumber, fxp.Nine)
		return p
	}
	attr := func(which string, atLeast fxp.Int) Prereq {
		p := NewAttributePrereq(e)
		p.Which = which
		p.QualifierCriteria.Qualifier = atLeast
		return p
	}
	met := func() Prereq { return attr(StrengthID, fxp.Ten) }
	unmet := func() Prereq { return attr(DexterityID, fxp.Twenty) }
	hidden := func() Prereq { return attr(IntelligenceID, fxp.Twenty) }
	equipment := func() Prereq {
		p := NewEquippedEquipmentPrereq()
		p.NameCriteria.Qualifier = "Longarm"
		return p
	}
	const (
		needsDX   = "\n- Has DX at least 20"
		needsGear = "\n- Has Longarm equipped"
	)
	for _, one := range []struct {
		name    string
		list    *PrereqList
		met     bool
		text    string
		penalty bool
	}{
		{"any of: skipped group and unmet item", list(false, skipped(hidden()), unmet()), false, needsDX, false},
		{"any of: skipped group and met item", list(false, skipped(hidden()), met()), true, "", false},
		{"all of: skipped group and met item", list(true, skipped(hidden()), met()), true, "", false},
		{"all of: skipped group and unmet item", list(true, skipped(hidden()), unmet()), false, needsDX, false},
		{"nested lists all skipped", list(false, list(true, skipped(hidden())), skipped(unmet())), true, "", false},
		{"any of: empty group and unmet item", list(false, list(true), unmet()), false, needsDX, false},
		{"empty top level", list(true), true, "", false},
		{"any of: skipped gear and met item", list(false, skipped(equipment()), met()), true, "", false},
		{"any of: skipped gear and unmet item", list(false, skipped(equipment()), unmet()), false, needsDX, false},
		{"any of: skipped group and unmet gear", list(false, skipped(hidden()), equipment()), false, needsGear, true},
		{"all of: unmet gear and met item", list(true, equipment(), met()), false, needsGear, true},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			var buffer xbytes.InsertBuffer
			penalty := false
			c.Equal(one.met, one.list.Satisfied(e, nil, &buffer, "\n- ", &penalty))
			c.Equal(one.text, buffer.String())
			c.Equal(one.penalty, penalty)
		})
	}
}

// TestPrereqListHasNothingToCheck verifies that a list has nothing to check when it holds nothing but lists that have
// nothing to check.
func TestPrereqListHasNothingToCheck(t *testing.T) {
	list := func(children ...Prereq) *PrereqList {
		p := NewPrereqList()
		p.Prereqs = children
		return p
	}
	for _, one := range []struct {
		name string
		list *PrereqList
		want bool
	}{
		{"nil", nil, true},
		{"empty", list(), true},
		{"nested empty lists", list(list(), list(list()), (*PrereqList)(nil)), true},
		{"nil entries", list(nil, list(nil)), true},
		{"item", list(NewTraitPrereq()), false},
		{"nested item", list(list(), list(list(NewTraitPrereq()))), false},
	} {
		check.New(t).Equal(one.want, one.list.HasNothingToCheck(), one.name)
	}
}

// TestPrereqListEvaluate verifies the result of a list and of each prerequisite within it: a list that isn't met has
// failed when any child it doesn't leave out has failed, a list skipped by its tech level skips everything in it
// without running its scripts, a list leaves out its skipped children and is skipped when nothing is left, and each
// script runs once.
func TestPrereqListEvaluate(t *testing.T) {
	e := NewEntity()
	e.Profile.TechLevel = "3"
	script := func(text string) *ScriptPrereq {
		p := NewScriptPrereq()
		p.Script = text
		return p
	}
	list := func(all bool, children ...Prereq) *PrereqList {
		p := NewPrereqList()
		p.All = all
		p.Prereqs = children
		return p
	}
	broken := script("nope(")
	stuck := script("for (;;);")
	later := list(true, broken, stuck)
	later.WhenTL = numberCriteria(criteria.AtLeastNumber, fxp.Ten)
	for _, one := range []struct {
		name string
		list *PrereqList
		want PrereqResult
	}{
		{"all of, all met", list(true, script("true"), script("true")), PrereqMet},
		{"all of, one unmet", list(true, script("true"), script("false")), PrereqUnmet},
		{"any of, one met", list(false, script("false"), script("true")), PrereqMet},
		{"any of, none met", list(false, script("false"), script("false")), PrereqUnmet},
		{"all of, one failed", list(true, script("true"), script("nope(")), PrereqFailed},
		{"all of, one failed and one unmet", list(true, script("false"), script("nope(")), PrereqFailed},
		{"any of, one met and one failed", list(false, script("true"), script("nope(")), PrereqMet},
		{"any of, one failed and one unmet", list(false, script("false"), script("nope(")), PrereqFailed},
		{"failed nested two deep", list(true, list(false, list(true, script("nope(")))), PrereqFailed},
		{
			"any of, a failed branch and a met one",
			list(false, list(true, script("nope(")), list(true, script("true"))), PrereqMet,
		},
		{"skipped by its tech level, with one that would fail", later, PrereqSkipped},
		{"any of, a skipped group and one unmet", list(false, later, script("false")), PrereqUnmet},
		{"any of, only skipped groups", list(false, later, list(true, list(true))), PrereqSkipped},
		{"any of, an empty group and one unmet", list(false, list(true), script("false")), PrereqUnmet},
		{"empty", list(true), PrereqSkipped},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			visits := make(map[Prereq]int)
			var result PrereqResult
			SuppressScriptResolveErrorLogging(func() {
				result = one.list.Evaluate(e, nil, func(node Prereq, nodeResult PrereqResult, _ string) {
					visits[node]++
					if node == one.list {
						c.Equal(one.want, nodeResult, "the visited result of the list")
					}
				})
			})
			c.Equal(one.want, result)
			c.Equal(1, visits[one.list], "the list is visited")
			for node, n := range visits {
				c.Equal(1, n, "%T is visited once", node)
			}
		})
	}
	c := check.New(t)
	visited := make(map[Prereq]PrereqResult)
	var reason string
	root := list(true, later, script("true"), broken.Clone(nil), stuck.Clone(nil))
	// A script that runs out of time is counted each time it is resolved, even when its result comes from the cache.
	prev := scriptExecTimeLimitOverride.Load()
	SetScriptExecTimeLimitForTesting(fxp.FromStringForced("0.01"))
	defer scriptExecTimeLimitOverride.Store(prev)
	stopped := abandonedScripts(e)
	SuppressScriptResolveErrorLogging(func() {
		root.Evaluate(e, nil, func(node Prereq, result PrereqResult, why string) {
			visited[node] = result
			if node == root.Prereqs[2] {
				reason = why
			}
		})
	})
	c.Equal(map[Prereq]PrereqResult{
		later: PrereqSkipped, broken: PrereqSkipped, stuck: PrereqSkipped, root.Prereqs[1]: PrereqMet,
		root.Prereqs[2]: PrereqFailed, root.Prereqs[3]: PrereqFailed, root: PrereqFailed,
	}, visited, "each prerequisite is visited with its own result")
	c.Equal(int64(1), abandonedScripts(e)-stopped, "each script runs once, and none in a skipped list does")
	c.Contains(reason, "SyntaxError", "a script that couldn't run gives its error")
}
