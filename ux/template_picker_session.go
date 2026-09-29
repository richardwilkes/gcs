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
	"maps"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/xreflect"
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
	if state == pickerError {
		return state
	}
	for _, child := range container.NodeChildren() {
		if !s.chosen[child] {
			continue
		}
		if gurps.IsTemplateChoiceContainer(child) && s.hasPicks(child) && s.state(child) >= pickerWarning {
			return pickerWarning
		}
		if !s.resolved(child) {
			state = pickerOpen
		}
	}
	return state
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
	target := formatPickerTotal(container, tp.Type, gurps.NumericRangeOf(tp.Qualifier.Qualifier))
	switch tp.Qualifier.Compare.EnsureValid() {
	case criteria.AtLeastNumber:
		target = "≥" + target
	case criteria.AtMostNumber:
		target = "≤" + target
	case criteria.NotEqualsNumber:
		target = "≠" + target
	default:
	}
	return formatPickerTotal(container, tp.Type, s.total(container, tp.Type)) + " / " + target
}

// templatePicker returns the picker of a choice container.
func templatePicker[T gurps.Node[T]](container T) gurps.TemplatePicker {
	if tpp, ok := any(container).(gurps.TemplatePickerProvider); ok {
		_, tp := tpp.TemplatePickerData()
		return *tp
	}
	return gurps.TemplatePicker{}
}
