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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
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
		expected:       make(map[pickerMeasureKey[T]]gurps.NumericRange),
	}
	s.runPicker = s.showPicker
	gurps.Traverse(func(row T) bool {
		if gurps.IsTemplateChoiceContainer(row) {
			for _, kind := range []picker.Type{picker.Points, picker.Value, picker.Weight} {
				s.expected[pickerMeasureKey[T]{row, kind}] = gurps.PickerMeasureRange(row, kind, prompted, nil)
			}
		}
		return false
	}, false, false, rows...)
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
	return s.modsAnswered[row]
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
