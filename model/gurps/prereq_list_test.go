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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// namedTrait returns a trait prerequisite for the name.
func namedTrait(name string) *TraitPrereq {
	p := NewTraitPrereq()
	p.NameCriteria.Qualifier = name
	return p
}

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

// TestPrereqListEvaluate verifies the result of a list and of each prerequisite within it: an unmet child decides an
// "all of" list and a met one an "any of" list, even beside a failed one, a list not decided that way has failed when
// any child it doesn't leave out has failed, a list skipped by its tech level skips everything in it, nested lists
// included, without running its scripts, a list leaves out its skipped children and is skipped when nothing is left,
// each script runs once, and without an entity a list is met and nothing is visited.
func TestPrereqListEvaluate(t *testing.T) {
	negated := func(p *PrereqList) *PrereqList {
		p.Not = true
		return p
	}
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
	later := list(true, broken, list(true, stuck))
	later.WhenTL = numberCriteria(criteria.AtLeastNumber, fxp.Ten)
	for _, one := range []struct {
		name string
		list *PrereqList
		want CheckResult
	}{
		{"all of, all met", list(true, script("true"), script("true")), CheckMet},
		{"all of, one unmet", list(true, script("true"), script("false")), CheckUnmet},
		{"any of, one met", list(false, script("false"), script("true")), CheckMet},
		{"any of, none met", list(false, script("false"), script("false")), CheckUnmet},
		{"all of, one failed", list(true, script("true"), script("nope(")), CheckFailed},
		{"all of, one failed and one unmet", list(true, script("false"), script("nope(")), CheckUnmet},
		{"all of, one unmet after one failed", list(true, script("nope("), script("false")), CheckUnmet},
		{"all of, a failed group and one unmet", list(true, list(false, script("nope(")), script("false")), CheckUnmet},
		{"any of, one met and one failed", list(false, script("true"), script("nope(")), CheckMet},
		{"any of, one met after one failed", list(false, script("nope("), script("true")), CheckMet},
		{"any of, one failed and one unmet", list(false, script("false"), script("nope(")), CheckFailed},
		{"failed nested two deep", list(true, list(false, list(true, script("nope(")))), CheckFailed},
		{
			"any of, a failed branch and a met one",
			list(false, list(true, script("nope(")), list(true, script("true"))), CheckMet,
		},
		{"skipped by its tech level, with one that would fail", later, CheckSkipped},
		{"any of, a skipped group and one unmet", list(false, later, script("false")), CheckUnmet},
		{"any of, only skipped groups", list(false, later, list(true, list(true))), CheckSkipped},
		{"any of, an empty group and one unmet", list(false, list(true), script("false")), CheckUnmet},
		{"empty", list(true), CheckSkipped},
		{"none of, none met", negated(list(false, script("false"), script("false"))), CheckMet},
		{"none of, one met", negated(list(false, script("false"), script("true"))), CheckUnmet},
		{"not all of, all met", negated(list(true, script("true"), script("true"))), CheckUnmet},
		{"not all of, one unmet", negated(list(true, script("true"), script("false"))), CheckMet},
		{"none of, one failed", negated(list(false, script("false"), script("nope("))), CheckFailed},
		{"none of, empty", negated(list(false)), CheckSkipped},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			visits := make(map[Prereq]int)
			var result CheckResult
			SuppressScriptResolveErrorLogging(func() {
				result = one.list.Evaluate(e, nil, func(node Prereq, nodeResult CheckResult, _ string) {
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
	visitedWithoutEntity := false
	c.Equal(CheckMet, list(true, namedTrait("Magery")).Evaluate(nil, nil, func(Prereq, CheckResult, string) {
		visitedWithoutEntity = true
	}), "without an entity, a list is met")
	c.False(visitedWithoutEntity, "and nothing in it is visited")
	visited := make(map[Prereq]CheckResult)
	var reason string
	root := list(true, later, script("true"), broken.Clone(nil))
	SuppressScriptResolveErrorLogging(func() {
		root.Evaluate(e, nil, func(node Prereq, result CheckResult, why string) {
			visited[node] = result
			if node == root.Prereqs[2] {
				reason = why
			}
		})
	})
	c.Equal(map[Prereq]CheckResult{
		later: CheckSkipped, broken: CheckSkipped, later.Prereqs[1]: CheckSkipped, stuck: CheckSkipped,
		root.Prereqs[1]: CheckMet, root.Prereqs[2]: CheckFailed, root: CheckFailed,
	}, visited, "each prerequisite is visited with its own result")
	c.Contains(reason, "SyntaxError", "a script that couldn't run gives its error")

	// Whether a skipped list's scripts ran shows in the count of scripts stopped before they could answer, since a
	// script that runs out of time is counted each time it is resolved, even when its result comes from the cache. The
	// time limit is lowered for this evaluation alone, in which the only script that runs is the one meant to run out
	// of it: a legitimate script has run out of so short a limit on a slow runner.
	prev := scriptExecTimeLimitOverride.Load()
	SetScriptExecTimeLimitForTesting(fxp.FromStringForced("0.01"))
	defer scriptExecTimeLimitOverride.Store(prev)
	stopped := abandonedScripts(e)
	root = list(true, later, stuck.Clone(nil))
	visited = make(map[Prereq]CheckResult)
	SuppressScriptResolveErrorLogging(func() {
		root.Evaluate(e, nil, func(node Prereq, result CheckResult, _ string) { visited[node] = result })
	})
	c.Equal(map[Prereq]CheckResult{
		later: CheckSkipped, broken: CheckSkipped, later.Prereqs[1]: CheckSkipped, stuck: CheckSkipped,
		root.Prereqs[1]: CheckFailed, root: CheckFailed,
	}, visited, "a script that runs out of time has failed")
	c.Equal(int64(1), abandonedScripts(e)-stopped, "each script runs once, and none in a skipped list does")
}

// TestNewPrereqCriteriaHashAndKeys checks that the hash follows the new fields, and that their nameable keys are found.
func TestNewPrereqCriteriaHashAndKeys(t *testing.T) {
	c := check.New(t)
	list := newPrereqListRequiringTrait("Magery")
	base := Hash64(list)
	list.Not = true
	c.NotEqual(base, Hash64(list), "negating a list changes its hash")
	equipped := NewEquippedEquipmentPrereq()
	base = Hash64(equipped)
	equipped.Not = true
	c.NotEqual(base, Hash64(equipped), "negating an equipment prerequisite changes its hash")

	trait := NewTraitPrereq()
	trait.NameCriteria.Qualifier = "Magery"
	trait.TagsCriteria = textCriteria(criteria.IsText, "@Tag@")
	trait.ModifierNameCriteria = textCriteria(criteria.IsText, "@Name@")
	trait.ModifierNotesCriteria = textCriteria(criteria.IsText, "@Notes@")
	trait.ModifierTagsCriteria = textCriteria(criteria.IsText, "@ModTag@")
	keys := make(map[string]string)
	trait.FillWithNameableKeys(keys, nil)
	for _, key := range []string{"Tag", "Name", "Notes", "ModTag"} {
		_, found := keys[key]
		c.True(found, "the trait's %s key is found", key)
	}
	skill := NewSkillPrereq()
	skill.TagsCriteria = textCriteria(criteria.IsText, "@SkillTag@")
	keys = make(map[string]string)
	skill.FillWithNameableKeys(keys, nil)
	_, found := keys["SkillTag"]
	c.True(found, "the skill's tag key is found")
}

// TestPrereqListDescribe verifies that a list joins its children to match its mode, parenthesizes nested lists of more
// than one child, notes a tech level condition, and hands its entity to its children, so that a weight is given in the
// entity's units.
func TestPrereqListDescribe(t *testing.T) {
	c := check.New(t)
	root := NewPrereqList()
	root.Prereqs = append(root.Prereqs, namedTrait("Magery"))
	anyOf := NewPrereqList()
	anyOf.All = false
	anyOf.WhenTL.Compare = criteria.AtLeastNumber
	anyOf.WhenTL.Qualifier = fxp.Three
	anyOf.Prereqs = append(anyOf.Prereqs, namedTrait("Luck"), namedTrait("DX"))
	single := NewPrereqList()
	single.Prereqs = append(single.Prereqs, namedTrait("Fearlessness"))
	root.Prereqs = append(root.Prereqs, anyOf, single)
	c.Equal("Has trait Magery and (has trait Luck or has trait DX) (only when TL at least 3) and has trait Fearlessness",
		root.Describe(nil, nil, func(s string) string { return s }))
	c.Equal("", NewPrereqList().Describe(nil, nil, func(s string) string { return s }))
	noneOf := &PrereqList{Not: true, Prereqs: Prereqs{namedTrait("Luck"), namedTrait("DX")}}
	c.Equal("None of (has trait Luck or has trait DX)", noneOf.Describe(nil, nil, func(s string) string { return s }))
	allOf := &PrereqList{All: true, Prereqs: Prereqs{namedTrait("Luck"), namedTrait("DX")}}
	noneOfAll := &PrereqList{Not: true, Prereqs: Prereqs{allOf}}
	c.Equal("Not all of (has trait Luck and has trait DX)",
		noneOfAll.Describe(nil, nil, func(s string) string { return s }),
		"a negated list holding only an all of group says not all of its items, without parenthesizing them twice")
	anyOfLuck := &PrereqList{Prereqs: Prereqs{namedTrait("Luck"), namedTrait("DX")}}
	holdingAnyOf := &PrereqList{All: true, Prereqs: Prereqs{anyOfLuck}}
	c.Equal("(Has trait Luck or has trait DX)", holdingAnyOf.Describe(nil, nil, func(s string) string { return s }),
		"a list that isn't negated keeps the parentheses of a lone group")
	notAllOfAny := &PrereqList{All: true, Not: true, Prereqs: Prereqs{anyOfLuck}}
	c.Equal("None of (has trait Luck or has trait DX)",
		notAllOfAny.Describe(nil, nil, func(s string) string { return s }),
		"and one holding only an any of group says none of its items")
	allOfAtTL := &PrereqList{
		All: true, WhenTL: criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.Three},
		Prereqs: Prereqs{namedTrait("Luck"), namedTrait("DX")},
	}
	noneOfAtTL := &PrereqList{Not: true, Prereqs: Prereqs{allOfAtTL}}
	c.Equal("None of ((has trait Luck and has trait DX) (only when TL at least 3))",
		noneOfAtTL.Describe(nil, nil, func(s string) string { return s }),
		"but keeps the group's own parentheses when its tech level condition follows them")

	named := NewScriptPrereq()
	named.Name = "Elf ancestry"
	unnamed := NewScriptPrereq()
	for _, one := range []struct {
		name     string
		prereqs  Prereqs
		expected string
	}{
		{"a named script", Prereqs{namedTrait("Magery"), named}, "Has trait Magery and Elf ancestry"},
		{"an unnamed script", Prereqs{namedTrait("Magery"), unnamed}, "Has trait Magery and a custom check"},
		{
			"a named script leading a group",
			Prereqs{namedTrait("Magery"), &PrereqList{All: true, Prereqs: Prereqs{named, namedTrait("Luck")}}},
			"Has trait Magery and (Elf ancestry and has trait Luck)",
		},
		{"an empty group", Prereqs{namedTrait("Magery"), &PrereqList{}}, "Has trait Magery"},
		{"an empty group first", Prereqs{&PrereqList{}, namedTrait("Magery")}, "Has trait Magery"},
		{
			"an empty group with a tech level",
			Prereqs{namedTrait("Magery"), &PrereqList{WhenTL: criteria.Number{Compare: criteria.AtLeastNumber}}},
			"Has trait Magery",
		},
		{
			"a none of group",
			Prereqs{namedTrait("Magery"), &PrereqList{Not: true, Prereqs: Prereqs{namedTrait("Luck"), namedTrait("DX")}}},
			"Has trait Magery and none of (has trait Luck or has trait DX)",
		},
		{
			"a not all of group",
			Prereqs{namedTrait("Magery"), &PrereqList{All: true, Not: true, Prereqs: Prereqs{namedTrait("Luck")}}},
			"Has trait Magery and not all of (has trait Luck)",
		},
		{
			"a group holding an empty group",
			Prereqs{namedTrait("Magery"), &PrereqList{Prereqs: Prereqs{&PrereqList{}, namedTrait("Luck")}}},
			"Has trait Magery and has trait Luck",
		},
	} {
		list := NewPrereqList()
		list.Prereqs = one.prereqs
		c.Equal(one.expected, list.Describe(nil, nil, func(s string) string { return s }), one.name)
	}

	entity := NewEntity()
	entity.SheetSettings.DefaultWeightUnits = fxp.Kilogram
	weighed := NewPrereqList()
	weighed.Prereqs = Prereqs{namedTrait("Magery"), NewContainedWeightPrereq(nil)}
	c.Equal("Has trait Magery and has a contained weight of at most 2.5 kg",
		weighed.Describe(entity, nil, func(s string) string { return s }), "a weight is given in the entity's units")
}

// TestPrereqListFailureText verifies how a list lays out what is unmet: a single unmet item collapses to that item,
// nested lists with their parent's mode are flattened into it, lists that don't apply at the tech level are left out,
// an "any of" list names every option, and a negated list names what is met as the item judged names it.
func TestPrereqListFailureText(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	entity.Profile.TechLevel = "3"
	met := namedTrait("Z")
	met.Has = false
	list := func(all bool, children ...Prereq) *PrereqList {
		p := NewPrereqList()
		p.All = all
		p.Prereqs = children
		return p
	}
	not := func(all bool, children ...Prereq) *PrereqList {
		p := list(all, children...)
		p.Not = true
		return p
	}
	met2 := namedTrait("Y")
	met2.Has = false
	atTL5 := list(true, namedTrait("B"))
	atTL5.WhenTL.Compare = criteria.AtLeastNumber
	atTL5.WhenTL.Qualifier = fxp.Five
	a, b, cc, d := namedTrait("A"), namedTrait("B"), namedTrait("C"), namedTrait("D")

	for _, one := range []struct {
		name     string
		list     *PrereqList
		expected string
	}{
		{"all of", list(true, a, b), "\n- Has trait A\n- Has trait B"},
		{"all of, one unmet", list(true, met, a), "\n- Has trait A"},
		{"all of, nested all of", list(true, a, list(true, b, cc)), "\n- Has trait A\n- Has trait B\n- Has trait C"},
		{
			"all of, nested any of", list(true, a, list(false, b, cc)),
			"\n- Has trait A\n- Requires at least one of:\n\t- Has trait B\n\t- Has trait C",
		},
		{"all of, nested any of with one option", list(true, a, list(false, b)), "\n- Has trait A\n- Has trait B"},
		{
			"all of, nested three deep", list(true, a, list(false, b, list(true, cc, d))),
			"\n- Has trait A\n- Requires at least one of:\n\t- Has trait B\n\t- Requires all of:\n\t\t- Has trait C" +
				"\n\t\t- Has trait D",
		},
		{"all of, skipped for TL", list(true, a, atTL5), "\n- Has trait A"},
		{"any of", list(false, a, b), "\n- Requires at least one of:\n\t- Has trait A\n\t- Has trait B"},
		{"any of with one option", list(false, a), "\n- Has trait A"},
		{
			"any of, nested any of", list(false, a, list(false, b, cc)),
			"\n- Requires at least one of:\n\t- Has trait A\n\t- Has trait B\n\t- Has trait C",
		},
		{
			"any of, nested all of with one unmet", list(false, a, list(true, met, b)),
			"\n- Requires at least one of:\n\t- Has trait A\n\t- Has trait B",
		},
		{"none of, one met", not(false, a, met), "\n- Requires none of:\n\t- Does not have trait Z"},
		{
			"not all of, all met", not(true, met, met2),
			"\n- Requires not all of:\n\t- Does not have trait Z\n\t- Does not have trait Y",
		},
		{
			"all of, nested none of", list(true, a, not(false, met)),
			"\n- Has trait A\n- Requires none of:\n\t- Does not have trait Z",
		},
	} {
		var buffer xbytes.InsertBuffer
		c.False(one.list.Satisfied(entity, nil, &buffer, "\n- ", nil), one.name)
		c.Equal(one.expected, buffer.String(), one.name)
	}
	c.True(list(false, a, met).Satisfied(entity, nil, nil, "\n- ", nil), "an any of list with a met option is met")
	c.True(not(false, a, b).Satisfied(entity, nil, nil, "\n- ", nil), "a none of list with nothing met is met")
	foe := namedTrait("@Foe@")
	foe.Has = false
	owner := NewTrait(entity, nil, false)
	owner.Replacements = map[string]string{"Foe": "Z"}
	var named xbytes.InsertBuffer
	c.False(not(false, foe).Satisfied(entity, owner, &named, "\n- ", nil), "a none of list with a met item is unmet")
	c.Equal("\n- Requires none of:\n\t- Does not have trait Z", named.String(), "the met item takes its owner's names")
	broken := NewScriptPrereq()
	broken.Script = "nope("
	var buffer xbytes.InsertBuffer
	SuppressScriptResolveErrorLogging(func() {
		c.False(not(false, a, broken).Satisfied(entity, nil, &buffer, "\n- ", nil), "a none of list that fails is unmet")
	})
	heading := "\n- Requires none of:"
	c.True(strings.HasPrefix(buffer.String(), heading+"\n\t- "), "a failed none of list says why: %q", buffer.String())
	c.False(strings.Contains(buffer.String(), "Has trait A"), "but not what was unmet")
	buffer = xbytes.InsertBuffer{}
	SuppressScriptResolveErrorLogging(func() {
		c.False(not(false, list(false, broken, a)).Satisfied(entity, nil, &buffer, "\n- ", nil),
			"a none of list holding a failed any of list is unmet")
	})
	c.True(strings.Contains(buffer.String(), "\n\t- Requires at least one of:"),
		"a failed list of the same mode keeps its own heading: %q", buffer.String())
	c.False(strings.Contains(buffer.String(), "\n\t- Has trait A"),
		"so what was unmet in it isn't listed as forbidden: %q", buffer.String())

	future := NewUnknownPrereq("future", []byte(`{"type":"future"}`))
	buffer = xbytes.InsertBuffer{}
	c.False(not(false, future).Satisfied(entity, nil, &buffer, "\n- ", nil),
		"a negated list can't make a prerequisite it doesn't know into a pass")
	c.True(strings.Contains(buffer.String(), "future"), "and says what it couldn't check: %q", buffer.String())
	c.Equal(CheckFailed, not(false, future).Evaluate(entity, nil, nil), "it fails rather than being unmet")
}

// TestPrereqListAppliesAt verifies that a list applies only at the tech levels its condition allows, and always when
// there is no condition or no entity, and that it applies with its parents only where each of them does as well.
func TestPrereqListAppliesAt(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	entity.Profile.TechLevel = "3"
	p := NewPrereqList()
	c.True(p.AppliesAt(entity), "no condition")
	p.WhenTL.Compare = criteria.AtLeastNumber
	p.WhenTL.Qualifier = fxp.Five
	c.False(p.AppliesAt(entity), "TL 3 is below 5")
	c.True(p.AppliesAt(nil), "no entity")
	child := NewPrereqList()
	child.Parent = p
	c.True(child.AppliesAt(entity), "a nested list without a condition applies on its own")
	c.False(child.AppliesWithParentsAt(entity), "but not with a parent that doesn't apply")
	entity.Profile.TechLevel = "5^"
	c.True(p.AppliesAt(entity), "TL 5^ is at least 5")
	c.True(child.AppliesWithParentsAt(entity), "and with its parent once that applies")
	child.WhenTL = numberCriteria(criteria.AtLeastNumber, fxp.Nine)
	c.False(child.AppliesWithParentsAt(entity), "but not when its own condition isn't met")
}

// TestPrereqListCloneResolvingEmpty verifies that a nil list becomes a new, empty one unless it belongs to a container,
// that an empty list is cloned unless it is to be pruned, and that a list holding something is always cloned.
func TestPrereqListCloneResolvingEmpty(t *testing.T) {
	c := check.New(t)
	empty := NewPrereqList()
	full := newPrereqListRequiringTrait("Magery")
	for _, one := range []struct {
		name        string
		list        *PrereqList
		isContainer bool
		prune       bool
		wantNil     bool
		wantItems   int
	}{
		{"nil, even when pruning", nil, false, true, false, 0},
		{"nil, for a container", nil, true, false, true, 0},
		{"empty", empty, false, false, false, 0},
		{"empty, pruned", empty, false, true, true, 0},
		{"not empty, pruned", full, false, true, false, 1},
	} {
		got := one.list.CloneResolvingEmpty(one.isContainer, one.prune)
		c.Equal(one.wantNil, got == nil, one.name)
		if got != nil {
			c.True(got != one.list, "%s: a new list", one.name)
			c.Equal(one.wantItems, len(got.Prereqs), "%s: what it holds", one.name)
		}
	}
}

// TestEquipmentPenaltyThroughNegation checks that equipment a prerequisite list requires to be equipped, through any
// number of negations, earns the missing equipment penalty when it isn't, even when the list it is in has failed, and
// that equipment it forbids never does.
func TestEquipmentPenaltyThroughNegation(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	shield := NewEquipment(e, nil, false)
	shield.Name = "Shield"
	shield.Equipped = true
	e.CarriedEquipment = []*Equipment{shield}
	c.True(shield.ReallyEquipped(), "precondition: the shield is equipped")
	equipped := func(name string, not bool) *EquippedEquipmentPrereq {
		p := NewEquippedEquipmentPrereq()
		p.NameCriteria = textCriteria(criteria.IsText, name)
		p.Not = not
		return p
	}
	carried := func(name string, not bool) *EquippedEquipmentPrereq {
		p := NewCarriedEquipmentPrereq()
		p.NameCriteria = textCriteria(criteria.IsText, name)
		p.Not = not
		return p
	}
	list := func(all, not bool, children ...Prereq) *PrereqList {
		p := NewPrereqList()
		p.All = all
		p.Not = not
		p.Prereqs = children
		return p
	}
	noneOf := func(children ...Prereq) *PrereqList { return list(false, true, children...) }
	allOf := func(children ...Prereq) *PrereqList { return list(true, false, children...) }
	anyOf := func(children ...Prereq) *PrereqList { return list(false, false, children...) }
	met := namedTrait("Z")
	met.Has = false
	for _, one := range []struct {
		name      string
		list      *PrereqList
		satisfied bool
		penalty   bool
	}{
		{"none of not having a longarm", noneOf(equipped("Longarm", true)), false, true},
		{"none of none of having a longarm", noneOf(noneOf(equipped("Longarm", false))), false, true},
		{"none of having the shield", noneOf(equipped("Shield", false)), false, false},
		{"none of none of not having the shield", noneOf(noneOf(equipped("Shield", true))), false, false},
		{"none of not having the shield", noneOf(equipped("Shield", true)), true, false},
		{"none of not having a kit carried", noneOf(carried("Kit", true)), false, false},
		{"none of a met item and having a longarm", noneOf(met, equipped("Longarm", false)), false, false},
		{
			"none of a failed all of not having a longarm",
			noneOf(allOf(equipped("Longarm", true), NewUnknownPrereq("future", []byte(`{"type":"future"}`)))),
			false, true,
		},
		{
			"any of having a longarm and a failure",
			anyOf(equipped("Longarm", false), NewUnknownPrereq("future", []byte(`{"type":"future"}`))),
			false, true,
		},
	} {
		root := NewPrereqList()
		root.Prereqs = Prereqs{one.list}
		var penalty bool
		c.Equal(one.satisfied, root.Satisfied(e, nil, nil, "", &penalty), "%s: satisfied", one.name)
		c.Equal(one.penalty, penalty, "%s: penalty", one.name)
	}
}
