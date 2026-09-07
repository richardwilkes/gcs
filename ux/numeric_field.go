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
	"strings"
	"unicode"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
)

// NumericField holds a numeric value that can be edited.
type NumericField[T xmath.Integer | xmath.Float] struct {
	undoableField[T]
	getPrototypes func(minValue, maxValue T) []T
	Format        func(T) string
	// DisplayFormat, if not nil, provides the text to show while the field does not have the keyboard focus. That text
	// is never parsed back into the value, since it may be a rounded, lossy rendering of it: the moment the field gains
	// the focus, or is handed new text by SetText, Format's exact rendering replaces it.
	DisplayFormat      func(T) string
	extract            func(s string) (T, error)
	validationTooltip  *unison.Panel
	savedTooltip       *unison.Panel
	minValue           T
	maxValue           T
	exception          T
	hasException       bool
	showingDisplayText bool
}

// NewNumericField creates a new field that formats its content.
func NewNumericField[T xmath.Integer | xmath.Float](targetMgr *TargetMgr, targetKey, undoTitle string, getPrototypes func(minValue, maxValue T) []T, get func() T, set func(T), format func(T) string, extract func(s string) (T, error), minValue, maxValue T) *NumericField[T] {
	f := newBaseNumericField(targetMgr, targetKey, undoTitle, getPrototypes, get, set, format, extract, minValue, maxValue)
	f.adjustMinimumTextWidth()
	f.Sync()
	return f
}

// NewNumericFieldWithException creates a new field that formats its content and can hold an exceptional value (one
// outside of the minimum/maximum range).
func NewNumericFieldWithException[T xmath.Integer | xmath.Float](targetMgr *TargetMgr, targetKey, undoTitle string, getPrototypes func(minValue, maxValue T) []T, get func() T, set func(T), format func(T) string, extract func(s string) (T, error), minValue, maxValue, exception T) *NumericField[T] {
	f := newBaseNumericField(targetMgr, targetKey, undoTitle, getPrototypes, get, set, format, extract, minValue, maxValue)
	f.exception = exception
	f.hasException = true
	f.adjustMinimumTextWidth()
	f.Sync()
	return f
}

// newUnitsField creates a new field that holds a fixed-point value carrying units, such as a length or a weight. The
// format and extract functions are expected to render and parse the value in the units the entity is using. Unlike the
// other numeric fields, keystrokes are not validated as they are typed, since the extractor rejects the text while a
// units suffix is being typed after the number.
func newUnitsField[T ~int64](targetMgr *TargetMgr, targetKey, undoTitle string, get func() T, set func(T), format func(T) string, extract func(string) (T, error), minValue, maxValue T, noMinWidth bool) *NumericField[T] {
	var getPrototypes func(minValue, maxValue T) []T
	if !noMinWidth {
		getPrototypes = fixedPointPrototypes[T]
	}
	f := NewNumericField(targetMgr, targetKey, undoTitle, getPrototypes, get, set, format, extract, minValue, maxValue)
	f.RuneTypedCallback = f.DefaultRuneTyped
	return f
}

// fixedPointPrototypes returns the values whose renderings a fixed-point field is sized to fit: the minimum and maximum
// with every decimal place filled in, and the widest single-digit value. The unbounded fxp.Min and fxp.Max are treated
// as -1 and 1, since a field wide enough to show them would be enormous.
func fixedPointPrototypes[T ~int64](minValue, maxValue T) []T {
	if minValue == T(fxp.Min) {
		minValue = T(-fxp.One)
	}
	minValue = T(fxp.Int(minValue).Floor() + fxp.One - 1)
	if maxValue == T(fxp.Max) {
		maxValue = T(fxp.One)
	}
	maxValue = T(fxp.Int(maxValue).Floor() + fxp.One - 1)
	return []T{minValue, T(fxp.Two - 1), maxValue}
}

func newBaseNumericField[T xmath.Integer | xmath.Float](targetMgr *TargetMgr, targetKey, undoTitle string, getPrototypes func(minValue, maxValue T) []T, get func() T, set func(T), format func(T) string, extract func(s string) (T, error), minValue, maxValue T) *NumericField[T] {
	f := &NumericField[T]{
		getPrototypes: getPrototypes,
		Format:        format,
		extract:       extract,
		minValue:      minValue,
		maxValue:      maxValue,
	}
	// Format is called through a closure rather than handed over directly, so that a Format installed later is used.
	f.init(f, unison.NewField(), targetMgr, targetKey, undoTitle, get, set, f.mustExtract,
		func(v T) string { return f.Format(v) })
	f.RuneTypedCallback = f.runeTyped
	f.ValidateCallback = f.validate
	return f
}

// showDisplayText replaces the field's text with the display rendering of the current value. Nothing is done when no
// DisplayFormat has been supplied, in which case the exact text is always what shows.
func (f *NumericField[T]) showDisplayText() {
	if f.DisplayFormat == nil {
		return
	}
	f.replaceText(f.DisplayFormat(f.get()))
	f.last = f.get()
	f.showingDisplayText = true
}

// replaceText puts the given text into the field without reporting a modification, applying the field state directly
// rather than going through SetText, so that text which must never be parsed back into the value can be shown. Since
// nothing else will ask for one, a layout is requested when the text actually changes: the field is sized to its text,
// and the display rendering and the exact text will usually differ in width.
func (f *NumericField[T]) replaceText(text string) {
	state := f.GetFieldState()
	if state.Text == text {
		return
	}
	state.Text = text
	f.ApplyFieldState(state)
	MarkForLayoutWithinDockable(f)
}

// restoreExactText puts the exact rendering of the value the field is showing back into it, if display text is
// showing. This must be called before anything that will parse the field's text or capture it as an undo state, so
// that the lossy display rendering never becomes the value.
//
// The value restored is last, which showDisplayText recorded as it rendered, rather than whatever the model holds at
// this moment. The two differ only when the model has been changed behind the field's back and the field is then told
// its new text -- which is exactly what the Description block's height and weight randomizers do. Rendering the model's
// new value here would leave the field already holding the text it is about to be given, so the assignment would be a
// no-op, no modification would be reported, and the randomization would go unrecorded and thus be impossible to undo.
func (f *NumericField[T]) restoreExactText() {
	if !f.showingDisplayText {
		return
	}
	f.showingDisplayText = false
	f.replaceText(f.Format(f.last))
}

// SetText sets the content of the field. The exact text is restored first, so that the "before" state the resulting
// modification records for undo holds the exact value rather than a rounded display rendering of it.
func (f *NumericField[T]) SetText(text string) {
	f.restoreExactText()
	f.Field.SetText(text)
}

func (f *NumericField[T]) gainedFocus() {
	f.restoreExactText()
	f.undoableField.gainedFocus()
}

func (f *NumericField[T]) lostFocus() {
	f.undoableField.lostFocus()
	f.showDisplayText()
}

func (f *NumericField[T]) mustExtract(s string) T {
	v, _ := f.extract(strings.TrimSpace(s)) //nolint:errcheck // Default value in case of error is acceptable
	if f.hasException && v == f.exception {
		return v
	}
	return min(max(v, f.minValue), f.maxValue)
}

func (f *NumericField[T]) validate() bool {
	if text := f.tooltipTextForValidation(); text != "" {
		if f.Tooltip != f.validationTooltip {
			f.savedTooltip = f.Tooltip
		}
		f.validationTooltip = newWrappedTooltip(text)
		f.Tooltip = f.validationTooltip
		return false
	}
	if f.validationTooltip != nil {
		if f.Tooltip == f.validationTooltip {
			f.Tooltip = f.savedTooltip
		}
		f.validationTooltip = nil
		f.savedTooltip = nil
	}
	return true
}

// SetBaseTooltip sets the tooltip to show when the field's content is valid. If the tooltip explaining why the content
// is currently invalid is showing, the new tooltip is remembered and installed in its place once the content becomes
// valid again, rather than replacing the explanation while the user still needs to see it.
func (f *NumericField[T]) SetBaseTooltip(tip *unison.Panel) {
	if f.validationTooltip != nil && f.Tooltip == f.validationTooltip {
		f.savedTooltip = tip
		return
	}
	f.Tooltip = tip
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
	if minimum := f.minValue; v < minimum {
		return fmt.Sprintf(i18n.Text("Value must be at least %s"), f.Format(minimum))
	}
	if maximum := f.maxValue; v > maximum {
		return fmt.Sprintf(i18n.Text("Value must be no more than %s"), f.Format(maximum))
	}
	return ""
}

func (f *NumericField[T]) runeTyped(ch rune) bool {
	if !unicode.IsControl(ch) {
		if f.minValue >= 0 && ch == '-' {
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

func (f *NumericField[T]) setWithoutUndo(state *unison.FieldState, focus bool) {
	// Whatever is about to show is exact text meant to be parsed -- it came from the user, from an undo, or from the
	// model by way of getData, never from DisplayFormat.
	f.showingDisplayText = false
	f.undoableField.setWithoutUndo(state, focus)
}

// Sync the field to the current value. While the field has the focus, the text in it is what the user is working on,
// so it is re-parsed rather than replaced.
func (f *NumericField[T]) Sync() {
	if !f.hasFocus && f.DisplayFormat != nil {
		// The value is to come from the model and the field isn't being edited, so show the display rendering of it
		// rather than the exact text. No modification can be reported for it, so there is nothing else to do. useGet
		// is cleared for the same reason getData clears it: a fetch made while the field has the focus must re-parse
		// the field's text rather than fetch the value again.
		f.useGet = false
		f.showDisplayText()
		f.Validate()
		return
	}
	f.undoableField.Sync()
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
	return f.minValue
}

// Max returns the maximum value allowed.
func (f *NumericField[T]) Max() T {
	return f.maxValue
}

// SetMinMax sets the minimum and maximum values and then adjusts the minimum text width, if a prototype function has
// been set.
func (f *NumericField[T]) SetMinMax(minValue, maxValue T) {
	if f.minValue != minValue || f.maxValue != maxValue {
		f.minValue = minValue
		f.maxValue = maxValue
		f.adjustMinimumTextWidth()
	}
}

func (f *NumericField[T]) adjustMinimumTextWidth() {
	if f.getPrototypes != nil {
		prototypes := f.getPrototypes(f.minValue, f.maxValue)
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
