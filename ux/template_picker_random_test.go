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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/check"
)

// firstPick is a randomizer that always gives the first of what it is asked to pick from, which leaves a shuffled list
// in its order and takes nothing at even odds.
type firstPick struct{}

func (firstPick) Intn(int) int { return 0 }

// countedPicks is firstPick that counts how often it is asked.
type countedPicks struct{ n int }

func (r *countedPicks) Intn(int) int {
	r.n++
	return 0
}

// newRandomChoiceTrait returns a trait with the name and cost in points.
func newRandomChoiceTrait(name string, points int) *gurps.Trait {
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = name
	trait.BasePoints = fxp.FromInteger(points)
	return trait
}

// newRandomChoice returns a choice container with the rule and options given.
func newRandomChoice(pt picker.Type, compare criteria.NumericComparison, qualifier int, options ...*gurps.Trait) *gurps.Trait {
	choice := gurps.NewTrait(nil, nil, true)
	choice.TemplatePicker.Type = pt
	choice.TemplatePicker.Qualifier.Compare = compare
	choice.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(qualifier)
	choice.Children = options
	SetParents(options, choice)
	return choice
}

// newRandomSession returns a session over the choice that picks the first of whatever it picks from.
func newRandomSession(choice *gurps.Trait, prompted bool) *pickerSession[*gurps.Trait] {
	s := newPickerSession(promptOperation{}, []*gurps.Trait{choice}, prompted)
	s.random.rnd = firstPick{}
	return s
}

// pickedNames returns the names of the container's options that are checked, in order.
func pickedNames(s *pickerSession[*gurps.Trait], container *gurps.Trait) string {
	return rowNames(s.picks(container, func(*gurps.Trait) bool { return true }))
}

// TestPickerSessionPicksRandomly verifies how much is picked at random under each kind of rule, that what is already
// checked is kept, and that nothing changes when the rule can't be met.
func TestPickerSessionPicksRandomly(t *testing.T) {
	for _, one := range []struct {
		name      string
		pt        picker.Type
		compare   criteria.NumericComparison
		qualifier int
		points    []int
		checked   string
		expected  string
		failed    bool
	}{
		{
			name: "an exact amount is searched for", pt: picker.Points, compare: criteria.EqualsNumber, qualifier: 10,
			points: []int{8, 5, 5}, expected: "b, c",
		},
		{
			name: "at least stops once reached", pt: picker.Points, compare: criteria.AtLeastNumber, qualifier: 10,
			points: []int{3, 8, 4}, expected: "a, b",
		},
		{
			name: "at most passes over what doesn't fit", pt: picker.Points, compare: criteria.AtMostNumber, qualifier: 10,
			points: []int{8, 5, 2}, expected: "a, c",
		},
		{
			name: "at most takes what a later pick made room for", pt: picker.Points, compare: criteria.AtMostNumber,
			qualifier: 10, points: []int{8, 6, -5}, expected: "a, b, c",
		},
		{
			name: "at most takes what together brings the picks within it", pt: picker.Points,
			compare: criteria.AtMostNumber, qualifier: 10, points: []int{20, -6, -6}, checked: "a", expected: "a, b, c",
		},
		{
			name: "at least a negative amount fills up to its size", pt: picker.Points, compare: criteria.AtLeastNumber,
			qualifier: -15, points: []int{-10, -10, -5}, expected: "a, c",
		},
		{
			name: "at most a negative amount stops once its size is reached", pt: picker.Points,
			compare: criteria.AtMostNumber, qualifier: -15, points: []int{-10, -10, -5}, expected: "a, b",
		},
		{
			name: "anything but takes one more when the picks come to the amount ruled out", pt: picker.Points,
			compare: criteria.NotEqualsNumber, qualifier: 0, points: []int{5, 3}, expected: "a",
		},
		{
			name: "what is checked is kept", pt: picker.Count, compare: criteria.EqualsNumber, qualifier: 2,
			points: []int{1, 1, 1}, checked: "c", expected: "a, c",
		},
		{
			name: "a rule that can't be met changes nothing", pt: picker.Count, compare: criteria.EqualsNumber,
			qualifier: 3, points: []int{1, 1}, checked: "a", expected: "a", failed: true,
		},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			options := make([]*gurps.Trait, len(one.points))
			for i, points := range one.points {
				options[i] = newRandomChoiceTrait(string(rune('a'+i)), points)
			}
			choice := newRandomChoice(one.pt, one.compare, one.qualifier, options...)
			s := newRandomSession(choice, false)
			for _, option := range options {
				if option.Name == one.checked {
					s.chosen[option] = true
				}
			}
			c.Equal(!one.failed, s.pickRandomly(choice))
			c.Equal(one.expected, pickedNames(s, choice))
			c.Equal(one.failed, strings.HasPrefix(s.hint(choice).text, "No random picks meet this rule."))
		})
	}
}

// TestPickerSessionRandomPicksDrawNestedChoicesOnce verifies that a choice below the one the picks were asked for is
// drawn at most once per attempt at the latter, so that one that can't be met isn't retried at every level.
func TestPickerSessionRandomPicksDrawNestedChoicesOnce(t *testing.T) {
	c := check.New(t)
	// shuffles counts the draws one attempt shuffling the options of every choice takes.
	shuffles := 0
	var build func(depth int) *gurps.Trait
	build = func(depth int) *gurps.Trait {
		if depth == 0 {
			return newRandomChoice(picker.Count, criteria.EqualsNumber, 2, newRandomChoiceTrait("x", 1))
		}
		units := make([]*gurps.Trait, 3)
		for i := range units {
			units[i] = gurps.NewTrait(nil, nil, true)
			units[i].Children = []*gurps.Trait{build(depth - 1)}
			SetParents(units[i].Children, units[i])
		}
		shuffles += len(units) - 1
		return newRandomChoice(picker.Count, criteria.EqualsNumber, 1, units...)
	}
	choice := build(3)
	s := newRandomSession(choice, false)
	draws := &countedPicks{}
	s.random.rnd = draws
	c.False(s.pickRandomly(choice))
	c.True(draws.n <= randomPickAttempts*shuffles, "%d draws", draws.n)
}

// TestPickerSessionRandomPicksSettleWhatTheyTake verifies that a random pick has its modifier choice made, so that it
// comes to a single amount, that an option costed but not picked is left as it was, that what a choice below a pick
// already checked leaves open is settled while its answer is kept, that one whose answer can't be settled is what the
// failure is put down to, and that a pick whose mandatory modifier choices are already made keeps them and is still
// asked about its modifiers.
func TestPickerSessionRandomPicksSettleWhatTheyTake(t *testing.T) {
	c := check.New(t)
	modified := newRandomChoiceTrait("modified", 10)
	modifierChoice := newTraitModifierChoiceFor(nil, true, []string{"+10", "+5"})
	for _, option := range modifierChoice.Children {
		option.CostAdj = option.Name
	}
	modified.AddModifiers(modifierChoice)
	x := newRandomChoiceTrait("x", 5)
	nested := newRandomChoice(picker.Count, criteria.EqualsNumber, 1, x, newRandomChoiceTrait("y", 5))
	unit := gurps.NewTrait(nil, nil, true)
	unit.Name = "unit"
	unit.Children = []*gurps.Trait{nested}
	SetParents(unit.Children, unit)
	plain := newRandomChoiceTrait("plain", 10)
	plainChoice := newTraitModifierChoiceFor(nil, true, []string{"A", "B"})
	plain.AddModifiers(plainChoice)
	choice := newRandomChoice(picker.Points, criteria.EqualsNumber, 20, modified, unit, plain)
	s := newRandomSession(choice, true)
	c.True(s.pickRandomly(choice))
	c.Equal("modified", pickedNames(s, choice), "its 15~20 points became the 20 asked for")
	c.True(modifierChoice.Children[0].Enabled())
	c.False(modifierChoice.Children[1].Enabled())
	c.True(s.modsAnswered[modified], "so it isn't asked about its modifiers again")
	c.Equal(pickerOK, s.state(choice))
	c.False(s.pickerAnswered[nested], "the choice in the unit passed over is still to be asked")
	c.False(s.chosen[x])
	c.False(s.modsAnswered[plain], "the option passed over is still to be asked about its modifiers")
	c.False(plainChoice.Children[0].Enabled(), "and has its modifier pick put back")

	open := newRandomChoiceTrait("open", 5)
	openChoice := newTraitModifierChoiceFor(nil, true, []string{"+10", "+5"})
	open.AddModifiers(openChoice)
	nested = newRandomChoice(picker.Count, criteria.EqualsNumber, 1, open, newRandomChoiceTrait("closed", 5))
	nested.Name = "nested"
	choice = newRandomChoice(picker.Count, criteria.EqualsNumber, 1, nested, newRandomChoiceTrait("plain", 10))
	s = newRandomSession(choice, true)
	// The nested choice was answered with Override, leaving the modifier choice of its pick to be made.
	s.chosen[nested], s.pickerAnswered[nested], s.chosen[open] = true, true, true
	c.True(s.pickRandomly(choice))
	c.Equal("nested", pickedNames(s, choice))
	c.Equal("open", pickedNames(s, nested), "an answered choice keeps its picks")
	c.True(openChoice.Children[0].Enabled(), "what they left open was settled")
	c.Equal(pickerOK, s.state(choice))

	five, ten := newRandomChoiceTrait("five", 5), newRandomChoiceTrait("ten", 10)
	nested = newRandomChoice(picker.Count, criteria.EqualsNumber, 1, five, ten)
	unit = gurps.NewTrait(nil, nil, true)
	unit.Name = "unit"
	unit.Children = []*gurps.Trait{nested}
	SetParents(unit.Children, unit)
	choice = newRandomChoice(picker.Points, criteria.AtLeastNumber, 10, unit, newRandomChoiceTrait("x", 10))
	s = newRandomSession(choice, false)
	// The nested choice was answered with Override, keeping both of its options, which miss its rule.
	s.chosen[unit], s.chosen[nested], s.pickerAnswered[nested] = true, true, true
	s.chosen[five], s.chosen[ten] = true, true
	c.False(s.pickRandomly(choice))
	c.HasPrefix(s.hint(choice).text, "What is already checked can't be settled at random. Needs attention below: unit.",
		"the rule is met, so it isn't what is blamed")

	made := newRandomChoiceTrait("made", 10)
	madeChoice := newTraitModifierChoiceFor(nil, true, []string{"Low", "High"}, "High")
	made.AddModifiers(madeChoice, newTraitModifierChoiceFor(nil, false, []string{"Loud", "Quiet"}))
	choice = newRandomChoice(picker.Count, criteria.EqualsNumber, 1, made)
	s = newRandomSession(choice, true)
	c.True(s.pickRandomly(choice))
	c.Equal(pickerOK, s.state(choice), "a pick whose modifier choice is already made comes to a single amount")
	c.True(madeChoice.Children[1].Enabled(), "and keeps that pick")
	c.Equal([]*gurps.Trait{made}, modifierTargets([]*gurps.Trait{made}, true, s.modsAnswered),
		"its modifier prompt asks for no pick, so it is still put up, as it is outside a picker")
}
