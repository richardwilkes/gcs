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
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
)

// pickerState is how a template choice stands, from best to worst.
type pickerState byte

const (
	// pickerOK means its own rule is met and nothing chosen from it has a problem.
	pickerOK pickerState = iota
	// pickerOpen means its picks could still meet its rule, or something chosen from it is still to be answered.
	pickerOpen
	// pickerWarning means its own rule is met, but a choice chosen from it is in error or warning.
	pickerWarning
	// pickerError means its own rule isn't met.
	pickerError
)

type pickerMeasureKey[T comparable] struct {
	row  T
	kind picker.Type
}

// pickerSession holds what has been answered for one part's template choices. Nothing dissolves until every dialog
// is done with.
type pickerSession[T gurps.Node[T]] struct {
	op             promptOperation
	prompted       bool
	chosen         map[T]bool
	pickerAnswered map[T]bool
	modsAnswered   map[T]bool
	// above holds the containers above a choice container, whose modifier choices count as they stand until asked.
	above map[T]bool
	// modPrompts holds the modifier prompt of each row with modifiers to ask about, as it stood at the start.
	modPrompts map[T]func(info *modifierPromptInfo) bool
	// expected holds what each choice container should come to by its rules alone, whatever is picked below it.
	expected map[pickerMeasureKey[T]]gurps.NumericRange
	// runPicker puts up the dialog for a choice container, returning how it was closed.
	runPicker func(row T, depth int) int
}

// newPickerSession returns a session for the rows. prompted says whether the modifier prompt follows (see
// gurps.PickerMeasureRange).
func newPickerSession[T gurps.Node[T]](op promptOperation, rows []T, prompted bool) *pickerSession[T] {
	s := &pickerSession[T]{
		op:             op,
		prompted:       prompted,
		chosen:         make(map[T]bool),
		pickerAnswered: make(map[T]bool),
		modsAnswered:   make(map[T]bool),
		above:          make(map[T]bool),
		modPrompts:     make(map[T]func(info *modifierPromptInfo) bool),
		expected:       make(map[pickerMeasureKey[T]]gurps.NumericRange),
	}
	s.runPicker = s.showPicker
	var containers []T
	gurps.Traverse(func(row T) bool {
		if gurps.IsTemplateChoiceContainer(row) {
			containers = append(containers, row)
			for parent := row.Parent(); !xreflect.IsNil(parent); parent = parent.Parent() {
				if !gurps.IsTemplateChoiceContainer(parent) {
					s.above[parent] = true
				}
			}
		}
		if prompted {
			// Choices are only resolved for a sheet, where picks are required.
			if ask := modifierPromptOf(row, true, false); ask != nil {
				s.modPrompts[row] = ask
			}
		}
		return false
	}, false, false, rows...)
	for _, row := range containers {
		for _, kind := range []picker.Type{picker.Points, picker.Value, picker.Weight} {
			s.expected[pickerMeasureKey[T]{row, kind}] = gurps.PickerMeasureRange(row, kind, prompted, s.taken)
		}
	}
	return s
}

// processRows replaces each choice container among the rows, however deep, with what was chosen from it, putting up
// the dialog for each one not already answered. Returns true for abort if one was canceled.
func (s *pickerSession[T]) processRows(rows []T) (revised []T, abort bool) {
	revised = make([]T, 0, len(rows))
	for _, one := range rows {
		result, cancel := s.processRow(one)
		if cancel {
			return nil, true
		}
		revised = append(revised, result...)
	}
	return revised, false
}

func (s *pickerSession[T]) processRow(row T) (revised []T, abort bool) {
	if !row.Container() {
		return []T{row}, false
	}
	if !gurps.IsTemplateChoiceContainer(row) {
		children, cancel := s.processRows(row.NodeChildren())
		if cancel {
			return nil, true
		}
		row.SetChildren(children)
		SetParents(children, row)
		return []T{row}, false
	}
	if !s.pickerAnswered[row] && s.runPicker(row, 0) == unison.ModalResponseCancel {
		return nil, true
	}
	var chosen []T
	for _, child := range row.NodeChildren() {
		if s.chosen[child] {
			chosen = append(chosen, child)
		}
	}
	if revised, abort = s.processRows(chosen); abort {
		return nil, true
	}
	SetParents(revised, row.Parent())
	return revised, false
}

func (s *pickerSession[T]) taken(row T) bool {
	return s.modsAnswered[row] || s.above[row]
}

// modTargets returns the row and the rows below it that have modifiers to ask about.
func (s *pickerSession[T]) modTargets(row T) []T {
	var targets []T
	gurps.Traverse(func(one T) bool {
		if s.modPrompts[one] != nil {
			targets = append(targets, one)
		}
		return false
	}, false, false, row)
	return targets
}

// modsOpen returns true if the row or a row below it has a mandatory modifier choice still to be made.
func (s *pickerSession[T]) modsOpen(row T) bool {
	return slices.ContainsFunc(s.modTargets(row), func(one T) bool {
		return gurps.HasOpenModifierChoice(one, s.prompted, s.taken)
	})
}

// chooseModifiers puts up the modifier prompts for the row now rather than after the choices. Confirming counts them
// as answered and checks the row; canceling puts everything back as it was.
func (s *pickerSession[T]) chooseModifiers(row T) {
	restore := s.snapshot(row)
	targets := s.modTargets(row)
	for i, target := range targets {
		if s.modPrompts[target](&modifierPromptInfo{
			op:           s.op,
			name:         target.String(),
			location:     rowLocation(target),
			step:         i + 1,
			steps:        len(targets),
			requirePicks: true,
			early: &earlyModifierPrompt{
				cost:          func() string { return target.String() + ": " + s.costText(target) },
				preconfigured: gurps.IsNodePreconfigured(target),
			},
		}) {
			restore()
			return
		}
		s.modsAnswered[target] = true
	}
	s.chosen[row] = true
}

// choosePicks puts up the dialog for a choice container picked from another now rather than after the outer OK.
// Confirming counts it as answered and checks the row, unless Override kept nothing, which leaves it to be asked later.
// Canceling puts everything back as it was.
func (s *pickerSession[T]) choosePicks(row T, depth int) {
	restore := s.snapshot(row)
	switch s.runPicker(row, depth+1) {
	case unison.ModalResponseOK:
	case unison.ModalResponseUserBase:
		if !s.hasPicks(row) {
			// Even if answered before, it is asked later.
			delete(s.pickerAnswered, row)
			return
		}
	default:
		restore()
		return
	}
	s.pickerAnswered[row] = true
	s.chosen[row] = true
}

// clear takes back the container's own picks, leaving the answers below them alone.
func (s *pickerSession[T]) clear(container T) {
	for _, child := range container.NodeChildren() {
		delete(s.chosen, child)
	}
}

// costText returns what the row costs, counting its modifier picks as made.
func (s *pickerSession[T]) costText(row T) string {
	taken := func(one T) bool { return one == row || s.taken(one) }
	if _, ok := any(row).(*gurps.Equipment); ok {
		return formatPickerTotal(row, picker.Value, gurps.PickerMeasureRange(row, picker.Value, true, taken)) + ", " +
			formatPickerTotal(row, picker.Weight, gurps.PickerMeasureRange(row, picker.Weight, true, taken))
	}
	return pointsText(gurps.PickerMeasureRange(row, picker.Points, true, taken))
}

// snapshot returns what puts back everything a popup for the row can change: the answers, and the levels, points,
// quantities and enabled modifiers at or below the row.
func (s *pickerSession[T]) snapshot(row T) (restore func()) {
	chosen, pickerAnswered, modsAnswered := maps.Clone(s.chosen), maps.Clone(s.pickerAnswered), maps.Clone(s.modsAnswered)
	var undo []func()
	gurps.Traverse(func(one T) bool {
		switch item := any(one).(type) {
		case *gurps.Trait:
			levels := item.Levels
			undo = append(snapshotEnabled(undo, item.Modifiers), func() { item.Levels = levels })
		case *gurps.Equipment:
			quantity := item.Quantity
			undo = append(snapshotEnabled(undo, item.Modifiers), func() { item.Quantity = quantity })
		case gurps.RawPointsAdjuster:
			points := item.RawPoints()
			undo = append(undo, func() { item.SetRawPoints(points) })
		}
		return false
	}, false, false, row)
	return func() {
		s.chosen, s.pickerAnswered, s.modsAnswered = chosen, pickerAnswered, modsAnswered
		for _, one := range undo {
			one()
		}
	}
}

func snapshotEnabled[M gurps.Node[M]](undo []func(), modifiers []M) []func() {
	gurps.Traverse(func(one M) bool {
		if gm, ok := any(one).(gurps.GeneralModifier); ok {
			enabled := gm.Enabled()
			undo = append(undo, func() { gm.SetEnabled(enabled) })
		}
		return false
	}, false, false, modifiers...)
	return undo
}

func (s *pickerSession[T]) hasPicks(container T) bool {
	return slices.ContainsFunc(container.NodeChildren(), func(child T) bool { return s.chosen[child] })
}

// actual returns what the row counts toward a choice made by kind: a choice container with picks, what they come to
// (so a count overridden past its number costs every pick); one without, what its rules expect; anything else, its
// range with the modifiers answered so far.
func (s *pickerSession[T]) actual(row T, kind picker.Type) gurps.NumericRange {
	if kind != picker.Count && gurps.IsTemplateChoiceContainer(row) && s.hasPicks(row) {
		return s.total(row, kind)
	}
	if r, ok := s.expected[pickerMeasureKey[T]{row, kind}]; ok {
		return r
	}
	return gurps.PickerMeasureRange(row, kind, s.prompted, s.taken)
}

// total returns what the container's picks come to toward a choice made by kind.
func (s *pickerSession[T]) total(container T, kind picker.Type) gurps.NumericRange {
	total := gurps.NumericRangeOf(0)
	for _, child := range container.NodeChildren() {
		if s.chosen[child] {
			total = total.Add(s.actual(child, kind))
		}
	}
	return total
}

// ownState judges the container's picks against its own rule alone.
func (s *pickerSession[T]) ownState(container T) pickerState {
	tp := templatePicker(container)
	total := s.total(container, tp.Type)
	switch {
	case !total.CanSatisfy(tp.Qualifier):
		return pickerError
	case total.IsSettled():
		return pickerOK
	default:
		return pickerOpen
	}
}

// state judges the container: its own error, then trouble in a choice picked from it, then anything still open.
func (s *pickerSession[T]) state(container T) pickerState {
	state := s.ownState(container)
	switch {
	case state == pickerError:
	case len(s.troubled(container)) != 0:
		state = pickerWarning
	case len(s.unresolved(container)) != 0:
		state = pickerOpen
	}
	return state
}

// troubled returns the choices picked from the container that are in error or warning themselves.
func (s *pickerSession[T]) troubled(container T) []T {
	return s.picks(container, func(child T) bool {
		return gurps.IsTemplateChoiceContainer(child) && s.hasPicks(child) && s.state(child) >= pickerWarning
	})
}

// unresolved returns the container's picks that have something left to answer.
func (s *pickerSession[T]) unresolved(container T) []T {
	return s.picks(container, func(child T) bool { return !s.resolved(child) })
}

// picks returns the container's picks that match.
func (s *pickerSession[T]) picks(container T, match func(T) bool) []T {
	var list []T
	for _, child := range container.NodeChildren() {
		if s.chosen[child] && match(child) {
			list = append(list, child)
		}
	}
	return list
}

// resolved returns true if nothing is left to answer for the row.
func (s *pickerSession[T]) resolved(row T) bool {
	if gurps.IsTemplateChoiceContainer(row) {
		return s.state(row) == pickerOK
	}
	return !gurps.HasOpenModifierChoice(row, s.prompted, s.taken)
}

// pillText returns what the container's picks come to against its target, as in "40~70 / 60" or "65 / ≥60".
func (s *pickerSession[T]) pillText(container T) string {
	tp := templatePicker(container)
	return formatPickerTotal(container, tp.Type, s.total(container, tp.Type)) + " / " +
		pickerTarget(container, formatPickerTotal[T])
}

// pickerText is a piece of the picker dialog's text, colored by the state it tells of, with a tooltip saying why.
type pickerText struct {
	text, tip string
	state     pickerState
}

// detail returns what follows the row's name: the modifiers picked for it, or for a choice, what was picked from it or
// its rule, flagged when its picks miss that rule or have trouble below them.
func (s *pickerSession[T]) detail(row T) pickerText {
	if !gurps.IsTemplateChoiceContainer(row) {
		preconfigured := gurps.IsNodePreconfigured(row)
		names := pickedModifierNames(row)
		if len(names) == 0 || (!preconfigured && !s.modsAnswered[row]) {
			return pickerText{}
		}
		t := pickerText{text: " [" + strings.Join(names, ", ") + "]"}
		if preconfigured {
			t.tip = i18n.Text("Preconfigured")
		}
		return t
	}
	tp := templatePicker(row)
	t := pickerText{text: " (" + xstrings.FirstToLower(tp.StringWithUnits(pickerWeightUnits(row))) + ")"}
	if !s.hasPicks(row) {
		return t
	}
	if tp.Type == picker.Count && s.ownState(row) == pickerOK {
		t.text = ": " + rowNames(s.picks(row, func(T) bool { return true }))
	}
	switch state := s.state(row); state {
	case pickerError:
		t.tip, t.state = s.ruleMiss(row), state
	case pickerWarning:
		t.tip = fmt.Sprintf(i18n.Text("Something picked below needs attention: %s."), rowNames(s.troubled(row)))
		t.state = state
	default:
	}
	return t
}

// ruleMiss returns why the container's picks miss its rule.
func (s *pickerSession[T]) ruleMiss(container T) string {
	tp := templatePicker(container)
	total := s.total(container, tp.Type)
	target := pickerTarget(container, pickerMeasureText[T])
	if tp.Type == picker.Count {
		return fmt.Sprintf(i18n.Text("%s picked, but this asks for %s."), total.Comma(), target)
	}
	return fmt.Sprintf(i18n.Text("The picks come to %s, but this asks for %s."),
		pickerMeasureText(container, tp.Type, total), target)
}

// cost returns what the row comes to by kind: red, with what was expected, when a choice's picks come to more or less;
// amber while it is still a range. A row that costs no points shows none.
func (s *pickerSession[T]) cost(row T, kind picker.Type) pickerText {
	actual := s.actual(row, kind)
	t := pickerText{text: pickerMeasureText(row, kind, actual)}
	if !actual.IsSettled() {
		t.state = pickerOpen
	}
	if kind == picker.Points {
		if value, settled := actual.Settled(); settled && value == 0 {
			return pickerText{}
		}
		t.text = " [" + t.text + "]"
	}
	if expected, ok := s.expected[pickerMeasureKey[T]{row, kind}]; ok && s.hasPicks(row) {
		text := pickerMeasureText(row, kind, expected)
		switch {
		case actual.Min != nil && expected.Max != nil && *actual.Min > *expected.Max:
			t.tip, t.state = fmt.Sprintf(i18n.Text("Over: expected %s."), text), pickerError
		case actual.Max != nil && expected.Min != nil && *actual.Max < *expected.Min:
			t.tip, t.state = fmt.Sprintf(i18n.Text("Under: expected %s."), text), pickerError
		default:
		}
	}
	return t
}

// hint returns the line under the container's list: how far its picks are off, what needs attention below, and what
// can be done about it.
func (s *pickerSession[T]) hint(container T) pickerText {
	tp := templatePicker(container)
	t := pickerText{state: s.state(container)}
	var parts []string
	if s.ownState(container) == pickerError {
		total, target := s.total(container, tp.Type), tp.Qualifier.Qualifier
		switch {
		case tp.Type == picker.Count:
			parts = append(parts, s.ruleMiss(container))
		case total.Min != nil && *total.Min > target:
			over := total.Add(gurps.NumericRangeOf(-target))
			parts = append(parts, fmt.Sprintf(i18n.Text("Over by %s."), pickerMeasureText(container, tp.Type, over)))
		case total.Max != nil && *total.Max < target:
			short := gurps.NumericRange{Min: new(target - *total.Max)}
			if total.Min != nil {
				short.Max = new(target - *total.Min)
			}
			parts = append(parts, fmt.Sprintf(i18n.Text("%s short."), pickerMeasureText(container, tp.Type, short)))
		default:
		}
	}
	if troubled := s.troubled(container); len(troubled) != 0 {
		parts = append(parts, fmt.Sprintf(i18n.Text("Needs attention below: %s."), rowNames(troubled)))
	}
	open := len(s.unresolved(container))
	switch {
	case t.state >= pickerWarning:
		parts = append(parts, i18n.Text("Override to keep it anyway."))
	case open == 1:
		parts = append(parts, i18n.Text("1 pick still depends on choices below. Choose it now to fix its cost, or later when the template is applied."))
	case open > 1:
		parts = append(parts, fmt.Sprintf(i18n.Text("%d picks still depend on choices below. Choose them now to fix the cost, or later when the template is applied."), open))
	case t.state == pickerOpen:
		parts = append(parts, i18n.Text("Could still meet the rule."))
	default:
		parts = append(parts, i18n.Text("Every pick has a fixed cost."))
	}
	t.text = strings.Join(parts, " ")
	return t
}

// tip returns what the state means, for the pill.
func (state pickerState) tip() string {
	switch state {
	case pickerOK:
		return i18n.Text("The rule is met.")
	case pickerOpen:
		return i18n.Text("Could still meet the rule once the choices below are made.")
	case pickerWarning:
		return i18n.Text("The rule is met, but something picked below needs attention.")
	default:
		return i18n.Text("The picks don't meet the rule.")
	}
}

// pickerTarget returns the container's target, formatted by format and marked with how it compares, as in "≥60".
func pickerTarget[T gurps.Node[T]](container T, format func(T, picker.Type, gurps.NumericRange) string) string {
	tp := templatePicker(container)
	target := format(container, tp.Type, gurps.NumericRangeOf(tp.Qualifier.Qualifier))
	switch tp.Qualifier.Compare.EnsureValid() {
	case criteria.AtLeastNumber:
		return "≥" + target
	case criteria.AtMostNumber:
		return "≤" + target
	case criteria.NotEqualsNumber:
		return "≠" + target
	default:
		return target
	}
}

// pickerMeasureText returns the measure as text, points being "5 points" rather than "5".
func pickerMeasureText[T gurps.Node[T]](row T, kind picker.Type, r gurps.NumericRange) string {
	if kind == picker.Points {
		return pointsText(r)
	}
	return formatPickerTotal(row, kind, r)
}

func rowNames[T gurps.Node[T]](rows []T) string {
	names := make([]string, len(rows))
	for i, row := range rows {
		names[i] = row.String()
	}
	return strings.Join(names, ", ")
}

// pickedModifierNames returns the names of the options picked in the row's modifier choices.
func pickedModifierNames[T gurps.Node[T]](row T) []string {
	switch item := any(row).(type) {
	case *gurps.Trait:
		return pickedOptionNames(item.Modifiers)
	case *gurps.Equipment:
		return pickedOptionNames(item.Modifiers)
	default:
		return nil
	}
}

func pickedOptionNames[M gurps.Node[M]](modifiers []M) []string {
	var names []string
	gurps.Traverse(func(m M) bool {
		if gm, ok := any(m).(gurps.GeneralModifier); ok {
			if _, isOption := gurps.ModifierChoiceFor(m); isOption {
				names = append(names, gm.NameWithReplacements())
			}
		}
		return false
	}, true, true, modifiers...)
	return names
}

// templatePicker returns the picker of a choice container.
func templatePicker[T gurps.Node[T]](container T) gurps.TemplatePicker {
	if tpp, ok := any(container).(gurps.TemplatePickerProvider); ok {
		_, tp := tpp.TemplatePickerData()
		return *tp
	}
	return gurps.TemplatePicker{}
}
