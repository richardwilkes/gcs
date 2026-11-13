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
	"slices"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xrand"
	"github.com/richardwilkes/unison"
)

const (
	// randomPickAttempts is how many times the picks of the choice they were asked for are drawn before giving up on
	// it. One draw can fail where another succeeds, since what a choice nested in an option comes to depends on its own
	// draw, which each attempt makes anew. A nested choice isn't retried itself, as that would multiply the attempts at
	// every level.
	randomPickAttempts = 32
	// randomSearchLimit caps the steps of the search for picks that come to an exact amount.
	randomSearchLimit = 100000
)

// randomPicks makes the random picks of the prompts of one transfer, which share it.
type randomPicks struct {
	// rnd is the source of the picks; nil uses xrand.New().
	rnd xrand.Randomizer
	// rest is set once the user asks for every pick still to come to be made at random.
	rest bool
}

// intn returns a random number from 0 to n-1. A nil receiver is fine.
func (r *randomPicks) intn(n int) int {
	if r == nil {
		return xrand.New().Intn(n)
	}
	if r.rnd == nil {
		r.rnd = xrand.New()
	}
	return r.rnd.Intn(n)
}

// remaining returns true if every pick still to come is to be made at random. A nil receiver is fine.
func (r *randomPicks) remaining() bool {
	return r != nil && r.rest
}

// shuffled returns a copy of the list in random order.
func shuffled[T any](r *randomPicks, list []T) []T {
	list = slices.Clone(list)
	for i := range len(list) - 1 {
		j := i + r.intn(len(list)-i)
		list[i], list[j] = list[j], list[i]
	}
	return list
}

// savedEntry is a map's entry for a key, or the lack of one.
type savedEntry[V any] struct {
	value V
	ok    bool
}

func saveEntry[K comparable, V any](m map[K]V, key K) savedEntry[V] {
	value, ok := m[key]
	return savedEntry[V]{value: value, ok: ok}
}

func (e savedEntry[V]) putBack(set func(V), remove func()) {
	if e.ok {
		set(e.value)
	} else {
		remove()
	}
}

// keep returns what puts back the answers, what choices are expected to come to and the enabled modifiers at or below
// the row. Unlike snapshot, it leaves the rest of the session alone and may be called more than once.
func (s *pickerSession[T]) keep(row T) (restore func()) {
	type saved struct {
		row                                            T
		chosen, pickerAnswered, modsAnswered, modsKept savedEntry[bool]
		expected                                       [3]savedEntry[gurps.NumericRange]
	}
	kinds := [3]picker.Type{picker.Points, picker.Value, picker.Weight}
	var rows []saved
	var undo []func()
	gurps.Traverse(func(one T) bool {
		entry := saved{
			row:            one,
			chosen:         saveEntry(s.chosen, one),
			pickerAnswered: saveEntry(s.pickerAnswered, one),
			modsAnswered:   saveEntry(s.modsAnswered, one),
			modsKept:       saveEntry(s.modsKept, one),
		}
		for i, kind := range kinds {
			entry.expected[i] = saveEntry(s.expected, pickerMeasureKey[T]{one, kind})
		}
		rows = append(rows, entry)
		switch item := any(one).(type) {
		case *gurps.Trait:
			undo = snapshotEnabled(undo, item.Modifiers)
		case *gurps.Equipment:
			undo = snapshotEnabled(undo, item.Modifiers)
		}
		return false
	}, false, false, row)
	return func() {
		for _, entry := range rows {
			one := entry.row
			entry.chosen.putBack(func(v bool) { s.chosen[one] = v }, func() { delete(s.chosen, one) })
			entry.pickerAnswered.putBack(func(v bool) { s.pickerAnswered[one] = v },
				func() { delete(s.pickerAnswered, one) })
			entry.modsAnswered.putBack(func(v bool) { s.modsAnswered[one] = v }, func() { delete(s.modsAnswered, one) })
			entry.modsKept.putBack(func(v bool) { s.modsKept[one] = v }, func() { delete(s.modsKept, one) })
			for i, kind := range kinds {
				key := pickerMeasureKey[T]{one, kind}
				entry.expected[i].putBack(func(v gurps.NumericRange) { s.expected[key] = v },
					func() { delete(s.expected, key) })
			}
		}
		for _, one := range undo {
			one()
		}
	}
}

// pickRandomly adds picks made at random to those the container already has until its rule is met, settling whatever
// they leave open below them the same way (see settle). Returns false, changing nothing, if it found no picks that
// do, which the container's hint then tells of, along with whether it was the picks already made that were in the way.
func (s *pickerSession[T]) pickRandomly(container T) bool {
	ok, stuck := s.drawPicks(container, randomPickAttempts)
	switch {
	case ok:
		delete(s.randomFailed, container)
	case stuck:
		s.randomFailed[container] = i18n.Text("What is already checked can't be settled at random.")
	default:
		s.randomFailed[container] = i18n.Text("No random picks meet this rule.")
	}
	return ok
}

// drawPicks is pickRandomly without the record of a failure, giving up after the given number of attempts. When it
// does, stuck is true if no attempt got past settling the options already checked.
func (s *pickerSession[T]) drawPicks(container T, attempts int) (ok, stuck bool) {
	restore := s.keep(container)
	stuck = true
	for range attempts {
		drawn, held := s.drawPicksOnce(container)
		if drawn && s.state(container) == pickerOK {
			return true, false
		}
		stuck = stuck && held
		restore()
	}
	return false, stuck
}

// settle answers what the option leaves open, so that it comes to a single amount: its unmade mandatory modifier
// choices are given a random option each, those already made counting as they stand, and each choice it is or holds
// has its picks made at random if it isn't answered yet, or what they leave open settled the same way if it is. Returns
// false if that leaves anything open.
func (s *pickerSession[T]) settle(option T) bool {
	choices := []T{option}
	if !gurps.IsTemplateChoiceContainer(option) {
		choices = s.nestedChoices(option)
		// Outermost first, since a container's answer counts for the rows inside it.
		for _, target := range s.modTargets(option) {
			if gurps.HasOpenModifierChoice(target, s.prompted, s.taken) {
				if answerModifiersRandomly(target, true, false, s.random.intn) {
					s.modsAnswered[target] = true
				} else {
					// Its prompt asks for no pick, so it is still put up, as it is outside a picker.
					s.modsKept[target] = true
				}
			}
		}
		// The choices below are now expected to come to what their rules give with the modifiers as they stand.
		s.expect(option)
	}
	for _, choice := range choices {
		if s.pickerAnswered[choice] {
			// Its answer is kept as it is; only what its picks leave open is settled.
			for _, pick := range s.picks(choice, func(T) bool { return true }) {
				if !s.settle(pick) {
					return false
				}
			}
			continue
		}
		if ok, _ := s.drawPicks(choice, 1); !ok {
			return false
		}
		s.pickerAnswered[choice] = true
		if choice != option {
			s.chosen[choice] = true
		}
	}
	return s.resolved(option)
}

// drawPicksOnce makes one try at the container's picks, returning false for drawn if it couldn't make them, with held
// true if the options already checked couldn't be settled. The session is left changed either way.
func (s *pickerSession[T]) drawPicksOnce(container T) (drawn, held bool) {
	tp := templatePicker(container)
	var candidates []T
	for _, option := range gurps.TemplateChoiceOptions(container) {
		switch {
		case !s.chosen[option]:
			candidates = append(candidates, option)
		case !s.settle(option):
			return false, true
		}
	}
	total, settled := s.total(container, tp.Type).Settled()
	if !settled {
		return false, true
	}
	candidates = shuffled(s.random, candidates)
	costs := make([]fxp.Int, len(candidates))
	var undo []func()
	take := func(i int) bool {
		s.chosen[candidates[i]] = true
		return true
	}
	if tp.Type == picker.Count {
		// Every option counts the same, so only those taken need settling.
		for i := range costs {
			costs[i] = fxp.One
		}
		take = func(i int) bool {
			restore := s.keep(candidates[i])
			s.chosen[candidates[i]] = true
			if s.settle(candidates[i]) {
				return true
			}
			restore()
			return false
		}
	} else {
		// What an option counts isn't known until it is settled, so all are, and those not taken are put back after.
		n := 0
		for _, option := range candidates {
			restore := s.keep(option)
			if s.settle(option) {
				if cost, ok := s.actual(option, tp.Type).Settled(); ok {
					candidates[n], costs[n] = option, cost
					undo = append(undo, restore)
					n++
					continue
				}
			}
			restore()
		}
		candidates, costs = candidates[:n], costs[:n]
	}
	compare, target := tp.Qualifier.Compare.EnsureValid(), tp.Qualifier.Qualifier
	if target < 0 {
		// Negative points are judged by how many are asked for, so everything is mirrored.
		target, total = -target, -total
		for i := range costs {
			costs[i] = -costs[i]
		}
		switch compare {
		case criteria.AtLeastNumber:
			compare = criteria.AtMostNumber
		case criteria.AtMostNumber:
			compare = criteria.AtLeastNumber
		default:
		}
	}
	// tried holds the candidates that were taken or couldn't be.
	tried := make([]bool, len(candidates))
	taken := make([]bool, len(candidates))
	add := func(i int) bool {
		if tried[i] {
			return false
		}
		tried[i] = true
		if !take(i) {
			return false
		}
		taken[i] = true
		total += costs[i]
		return true
	}
	switch {
	case compare == criteria.EqualsNumber && tp.Type != picker.Count:
		picks, ok := exactSubset(costs, target-total, randomSearchLimit)
		if !ok {
			return false, false
		}
		for _, i := range picks {
			add(i)
		}
	case compare == criteria.EqualsNumber || compare == criteria.AtLeastNumber:
		// Just enough to reach the target.
		for i := range candidates {
			if total >= target {
				break
			}
			if costs[i] >= 0 {
				add(i)
			}
		}
	case compare == criteria.AtMostNumber:
		// What takes away, until the picks are within the target, then as much as fits, going around again when
		// anything was added, as that may have made room for more.
		for i := 0; i < len(candidates) && total > target; i++ {
			if costs[i] < 0 {
				add(i)
			}
		}
		for added := true; added; {
			added = false
			for i := range candidates {
				if total+costs[i] <= target && add(i) {
					added = true
				}
			}
		}
	default:
		// Any amount will do, so each is taken at even odds, and one more if that came to the one amount ruled out.
		for i := range candidates {
			if s.random.intn(2) == 1 {
				add(i)
			}
		}
		for i := 0; i < len(candidates) && !compare.Matches(target, total); i++ {
			if costs[i] != 0 {
				add(i)
			}
		}
	}
	for i, restore := range undo {
		if !taken[i] {
			restore()
		}
	}
	return true, false
}

// exactSubset returns the indexes of costs that come to exactly target, preferring the earlier ones, or false if there
// are none or finding them took more than limit steps.
func exactSubset(costs []fxp.Int, target fxp.Int, limit int) ([]int, bool) {
	// most[i] and least[i] are the most and the least that costs[i:] can add.
	most := make([]fxp.Int, len(costs)+1)
	least := make([]fxp.Int, len(costs)+1)
	for i := len(costs) - 1; i >= 0; i-- {
		most[i] = most[i+1] + max(costs[i], 0)
		least[i] = least[i+1] + min(costs[i], 0)
	}
	var picks []int
	var search func(i int, need fxp.Int) bool
	search = func(i int, need fxp.Int) bool {
		if need == 0 {
			return true
		}
		if need > most[i] || need < least[i] || limit <= 0 {
			return false
		}
		limit--
		picks = append(picks, i)
		if search(i+1, need-costs[i]) {
			return true
		}
		picks = picks[:len(picks)-1]
		return search(i+1, need)
	}
	if !search(0, target) {
		return nil, false
	}
	return picks, true
}

// The response codes of the buttons that make a prompt's picks at random, which follow those of its Override and Clear
// Selections buttons: autoPickResponse for its own picks and autoPickAllResponse for every pick still to come as well.
const (
	autoPickResponse = unison.ModalResponseUserBase + 2 + iota
	autoPickAllResponse
)

// withAutoPickButtons returns the buttons with those that make a prompt's picks at random ahead of them: Auto Pick and,
// withAll, Auto Pick All. setupAutoPickButtons finishes them once the dialog is made.
func withAutoPickButtons(buttons []*unison.DialogButtonInfo, withAll bool) []*unison.DialogButtonInfo {
	auto := []*unison.DialogButtonInfo{{Title: i18n.Text("Auto Pick"), ResponseCode: autoPickResponse}}
	if withAll {
		auto = append(auto, &unison.DialogButtonInfo{
			Title:        i18n.Text("Auto Pick All"),
			ResponseCode: autoPickAllResponse,
		})
	}
	return append(auto, buttons...)
}

// setupAutoPickButtons gives the dialog's buttons from withAutoPickButtons their tooltips and has them call pick, with
// whether Auto Pick All was the one pressed, rather than close the dialog. Returns the Auto Pick button.
func setupAutoPickButtons(dialog *unison.Dialog, pick func(all bool)) *unison.Button {
	one := dialog.Button(autoPickResponse)
	one.Tooltip = newWrappedTooltip(i18n.Text("Make what is still to be picked here at random, within its rule, and move on"))
	one.ClickCallback = func() { pick(false) }
	if all := dialog.Button(autoPickAllResponse); all != nil {
		all.Tooltip = newWrappedTooltip(i18n.Text("Make what is still to be picked here, and every pick still to come, at random, each within its rule"))
		all.ClickCallback = func() { pick(true) }
	}
	return one
}
