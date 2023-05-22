/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xmath"
	"github.com/richardwilkes/unison"
)

// NumericField holds a numeric value that can be edited.
type NumericField[T xmath.Numeric] struct {
	*unison.Field
	targetMgr     *TargetMgr
	targetKey     string
	undoTitle     string
	getPrototypes func(min, max T) []T
	get           func() T
	set           func(T)
	Format        func(T) string
	extract       func(s string) (T, error)
	last          T
	min           T
	max           T
	exception     T
	hasException  bool
	useGet        bool
	marksModified bool
}

// NewNumericField creates a new field that formats its content.
func NewNumericField[T xmath.Numeric](targetMgr *TargetMgr, targetKey, undoTitle string, getPrototypes func(min, max T) []T, get func() T, set func(T), format func(T) string, extract func(s string) (T, error), min, max T) *NumericField[T] {
	f := newBaseNumericField(targetMgr, targetKey, undoTitle, getPrototypes, get, set, format, extract, min, max)
	f.adjustMinimumTextWidth()
	f.Sync()
	return f
}

// NewNumericFieldWithException creates a new field that formats its content and can hold an exceptional value (one
// outside of the min/max range.
func NewNumericFieldWithException[T xmath.Numeric](targetMgr *TargetMgr, targetKey, undoTitle string, getPrototypes func(min, max T) []T, get func() T, set func(T), format func(T) string, extract func(s string) (T, error), min, max, exception T) *NumericField[T] {
	f := newBaseNumericField(targetMgr, targetKey, undoTitle, getPrototypes, get, set, format, extract, min, max)
	f.exception = exception
	f.hasException = true
	f.adjustMinimumTextWidth()
	f.Sync()
	return f
}

func newBaseNumericField[T xmath.Numeric](targetMgr *TargetMgr, targetKey, undoTitle string, getPrototypes func(min, max T) []T, get func() T, set func(T), format func(T) string, extract func(s string) (T, error), min, max T) *NumericField[T] {
	f := &NumericField[T]{
		Field:         unison.NewField(),
		targetMgr:     targetMgr,
		targetKey:     targetKey,
		undoTitle:     undoTitle,
		getPrototypes: getPrototypes,
		get:           get,
		set:           set,
		Format:        format,
		extract:       extract,
		last:          get(),
		min:           min,
		max:           max,
		useGet:        true,
		marksModified: true,
	}
	f.Self = f
	f.LostFocusCallback = f.lostFocus
	f.RuneTypedCallback = f.runeTyped
	f.ModifiedCallback = f.modified
	f.ValidateCallback = f.validate
	if targetMgr != nil && targetKey != "" {
		f.RefKey = targetKey
	}
	return f
}

func (f *NumericField[T]) lostFocus() {
	f.useGet = true
	f.SetText(f.Format(f.mustExtract(f.Text())))
	f.DefaultFocusLost()
}

func (f *NumericField[T]) getData() string {
	if f.useGet {
		f.useGet = false
		return f.Format(f.get())
	}
	return f.Text()
}

func (f *NumericField[T]) mustExtract(s string) T {
	v, _ := f.extract(strings.TrimSpace(s)) //nolint:errcheck // Default value in case of error is acceptable
	if f.hasException && v == f.exception {
		return v
	}
	return xmath.Min(xmath.Max(v, f.min), f.max)
}

func (f *NumericField[T]) validate() bool {
	if text := f.tooltipTextForValidation(); text != "" {
		f.Tooltip = unison.NewTooltipWithText(text)
		return false
	}
	f.Tooltip = nil
	return true
}

func (f *NumericField[T]) tooltipTextForValidation() string {
	s := strings.TrimSpace(f.Text())
	v, err := f.extract(s)
	if err != nil || s == "-" || s == "+" {
		return i18n.Text("Invalid value")
	}
	if f.hasException && v == f.exception {
		return ""
	}
	if minimum := f.min; v < minimum {
		return fmt.Sprintf(i18n.Text("Value must be at least %s"), f.Format(minimum))
	}
	if maximum := f.max; v > maximum {
		return fmt.Sprintf(i18n.Text("Value must be no more than %s"), f.Format(maximum))
	}
	return ""
}

func (f *NumericField[T]) runeTyped(ch rune) bool {
	if !unicode.IsControl(ch) {
		if f.min >= 0 && ch == '-' {
			unison.Beep()
			return false
		}
		if text := strings.TrimSpace(string(f.RunesIfPasted([]rune{ch}))); text != "-" && text != "+" {
			if _, err := f.extract(text); err != nil {
				unison.Beep()
				return false
			}
		}
	}
	return f.DefaultRuneTyped(ch)
}

func (f *NumericField[T]) modified(before, after *unison.FieldState) {
	if f.CurrentUndoID() != unison.NoUndoID {
		if mgr := unison.UndoManagerFor(f); mgr != nil {
			undo := NewTargetUndo(f.targetMgr, f.targetKey, f.undoTitle, f.CurrentUndoID(),
				func(target *unison.Panel, data *unison.FieldState) {
					self := f
					if target != nil {
						if field, ok := target.Self.(*NumericField[T]); ok {
							self = field
						}
					}
					self.setWithoutUndo(data, true)
				}, before)
			undo.AfterData = after
			mgr.Add(undo)
		}
	}
	f.adjustForText()
}

func (f *NumericField[T]) adjustForText() {
	if v := f.mustExtract(f.Text()); f.last != v {
		f.last = v
		f.set(v)
		MarkForLayoutWithinDockable(f)
		if f.marksModified {
			MarkModified(f)
		}
	}
}

func (f *NumericField[T]) setWithoutUndo(state *unison.FieldState, focus bool) {
	f.ApplyFieldState(state)
	f.adjustForText()
	if focus {
		f.RequestFocus()
	}
	f.Validate()
}

// Sync the field to the current value.
func (f *NumericField[T]) Sync() {
	if !f.Focused() {
		f.useGet = true
	}
	state := f.GetFieldState()
	state.Text = f.getData()
	f.setWithoutUndo(state, false)
}

// HasException returns true if an exception value can be used.
func (f *NumericField[T]) HasException() bool {
	return f.hasException
}

// Exception returns the exception value.
func (f *NumericField[T]) Exception() T {
	return f.exception
}

// Min returns the minimum value allowed.
func (f *NumericField[T]) Min() T {
	return f.min
}

// Max returns the maximum value allowed.
func (f *NumericField[T]) Max() T {
	return f.max
}

// SetMinMax sets the minimum and maximum values and then adjusts the minimum text width, if a prototype function has
// been set.
func (f *NumericField[T]) SetMinMax(min, max T) {
	if f.min != min || f.max != max {
		f.min = min
		f.max = max
		f.adjustMinimumTextWidth()
	}
}

func (f *NumericField[T]) adjustMinimumTextWidth() {
	if f.getPrototypes != nil {
		prototypes := f.getPrototypes(f.min, f.max)
		candidates := make([]string, 0, len(prototypes))
		for _, v := range prototypes {
			candidates = append(candidates, f.Format(v))
		}
		f.SetMinimumTextWidthUsing(candidates...)
	}
}

// SetMarksModified sets whether this field will attempt to mark its ModifiableRoot as modified. Default is true.
func (f *NumericField[T]) SetMarksModified(marksModified bool) {
	f.marksModified = marksModified
}
