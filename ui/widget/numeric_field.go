/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package widget

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
	undoID        int64
	undoTitle     string
	getPrototypes func(min, max T) []T
	get           func() T
	set           func(T)
	Format        func(T) string
	extract       func(s string) (T, error)
	last          T
	min           T
	max           T
	prototype     T
	useGet        bool
	marksModified bool
	inUndo        bool
}

// NewNumericField creates a new field that formats its content.
func NewNumericField[T xmath.Numeric](undoTitle string, getPrototypes func(min, max T) []T, get func() T, set func(T), format func(T) string, extract func(s string) (T, error), min, max T) *NumericField[T] {
	f := &NumericField[T]{
		Field:         unison.NewField(),
		undoID:        unison.NextUndoID(),
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
	f.adjustMinimumTextWidth()
	f.Sync()
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

func (f *NumericField[T]) modified() {
	text := f.Text()
	if !f.inUndo && f.undoID != unison.NoUndoID {
		if mgr := unison.UndoManagerFor(f); mgr != nil {
			mgr.Add(&unison.UndoEdit[string]{
				ID:       f.undoID,
				EditName: f.undoTitle,
				UndoFunc: func(e *unison.UndoEdit[string]) { f.setWithoutUndo(e.BeforeData, true) },
				RedoFunc: func(e *unison.UndoEdit[string]) { f.setWithoutUndo(e.AfterData, true) },
				AbsorbFunc: func(e *unison.UndoEdit[string], other unison.Undoable) bool {
					if e2, ok := other.(*unison.UndoEdit[string]); ok && e2.ID == f.undoID {
						e.AfterData = e2.AfterData
						return true
					}
					return false
				},
				BeforeData: f.getData(),
				AfterData:  text,
			})
		}
	}
	if v := f.mustExtract(f.Text()); f.last != v {
		f.last = v
		f.set(v)
		MarkForLayoutWithinDockable(f)
		if f.marksModified {
			MarkModified(f)
		}
	}
}

func (f *NumericField[T]) setWithoutUndo(text string, focus bool) {
	f.inUndo = true
	f.SetText(text)
	f.inUndo = false
	if focus {
		f.RequestFocus()
		f.SelectAll()
	}
}

// Sync the field to the current value.
func (f *NumericField[T]) Sync() {
	if !f.Focused() {
		f.useGet = true
	}
	f.setWithoutUndo(f.getData(), false)
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
