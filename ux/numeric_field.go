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
	"unicode"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/mod"
)

// Beep plays the system beep sound when a field refuses typed or pasted input. Tests replace it with a no-op so that
// running them makes no noise.
var Beep = unison.Beep

// NumericField holds a numeric value that can be edited.
type NumericField[T xmath.Integer | xmath.Float] struct {
	undoableField[T]
	getPrototypes func(minValue, maxValue T) []T
	Format        func(T) string
	// DisplayFormat, if not nil, provides the text to show while the field does not have the keyboard focus. That text
	// is never parsed back into the value, since it may be a rounded, lossy rendering of it: the moment the field gains
	// the focus, or is handed new text by SetText or a new value by an assistive technology, Format's exact rendering
	// replaces it.
	DisplayFormat      func(T) string
	extract            func(s string) (T, error)
	validationTooltip  *unison.Panel
	savedTooltip       *unison.Panel
	minValue           T
	maxValue           T
	exception          T
	hasException       bool
	showingDisplayText bool
	// limitToMax has typed or pasted text that would put the value above the maximum refused; see refuseAboveMax.
	limitToMax bool
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

// NewUnitsField creates a new field that holds a fixed-point value carrying units, such as a length or a weight. The
// format and extract functions are expected to render and parse the value in the units the entity is using. Unlike the
// other numeric fields, keystrokes are not validated as they are typed, since the extractor rejects the text while a
// units suffix is being typed after the number.
func NewUnitsField[T ~int64](targetMgr *TargetMgr, targetKey, undoTitle string, get func() T, set func(T), format func(T) string, extract func(string) (T, error), minValue, maxValue T, noMinWidth bool) *NumericField[T] {
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
	// An edit made over text the field can't parse records the exact text of the value the field holds in its place,
	// since that text leaves the value as it is (see mustExtract): undoing the edit would otherwise put the text back
	// but leave the value the edit set.
	modified := f.ModifiedCallback
	f.ModifiedCallback = func(before, after *unison.FieldState) {
		if _, ok := f.extractValid(before.Text); !ok {
			exact := *before
			exact.Text = f.Format(f.last)
			before = &exact
		}
		modified(before, after)
	}
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

// PerformAccessibilityAction implements unison.AccessibilityActor. An assistive technology replacing the field's text
// does so as SetText does, with the exact text restored first, so that the "before" state the resulting modification
// records for undo holds the exact value. While display text is showing, a range of it means nothing in the exact text
// unless the two are the same, so replacing part of it is otherwise refused, and replacing all of it replaces all of
// the exact text. Without the focus, text that would leave the value as it is changes nothing, since the field is
// synced to show its value only once the value changes, and text the field can't parse, which leaves the value as it is
// too (see mustExtract), is refused rather than left showing as invalid, with no loss of the focus to come and put the
// value's text back.
func (f *NumericField[T]) PerformAccessibilityAction(req accessibility.ActionRequest) bool {
	if req.Action != accessibility.SetValue && req.Action != accessibility.ReplaceText {
		return f.Field.PerformAccessibilityAction(req)
	}
	exact := f.Text()
	if f.showingDisplayText {
		exact = f.Format(f.last)
		if req.Action == accessibility.ReplaceText {
			shown := f.Text()
			whole := req.Start <= 0 && req.End >= len([]rune(shown))
			if !whole && shown != exact {
				return false
			}
			if whole {
				req.Start, req.End = 0, len([]rune(exact))
			}
		}
	}
	if !f.hasFocus {
		text := req.Value
		if req.Action == accessibility.ReplaceText {
			runes := []rune(exact)
			start := min(max(req.Start, 0), len(runes))
			end := min(max(req.End, start), len(runes))
			text = string(runes[:start]) + req.Value + string(runes[end:])
		}
		text = strings.TrimSpace(text)
		if _, ok := f.extractValid(text); !ok {
			return false
		}
		if f.mustExtract(text) == f.last {
			return true
		}
	}
	f.restoreExactText()
	return f.Field.PerformAccessibilityAction(req)
}

func (f *NumericField[T]) gainedFocus() {
	f.restoreExactText()
	f.undoableField.gainedFocus()
}

func (f *NumericField[T]) lostFocus() {
	f.undoableField.lostFocus()
	f.showDisplayText()
}

// mustExtract returns the value the text holds, brought into the field's range unless it is the exception. Text the
// field can't parse, which shows as invalid, holds no value, so it is taken for the value the field holds: the value is
// left as it is, and losing the focus puts its text back.
func (f *NumericField[T]) mustExtract(s string) T {
	v, ok := f.extractValid(s)
	if !ok {
		return f.last
	}
	if f.hasException && v == f.exception {
		return v
	}
	return min(max(v, f.minValue), f.maxValue)
}

// extractValid returns the value the text holds, which may lie outside of the field's range, and whether the field can
// parse the text at all. Text with no digit, such as a sign or a decimal point by itself, which some extractors take
// for zero, is only the start of a number.
func (f *NumericField[T]) extractValid(s string) (T, bool) {
	s = strings.TrimSpace(s)
	v, err := f.extract(s)
	return v, err == nil && strings.ContainsAny(s, "0123456789")
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
	v, ok := f.extractValid(f.Text())
	if !ok {
		return i18n.Text("Invalid value")
	}
	if f.hasException && v == f.exception {
		return ""
	}
	if minimum := f.minValue; v < minimum {
		return i18n.Text("Value must be at least %s", f.Format(minimum))
	}
	if maximum := f.maxValue; v > maximum {
		return i18n.Text("Value must be no more than %s", f.Format(maximum))
	}
	return ""
}

func (f *NumericField[T]) runeTyped(ch rune) bool {
	if !unicode.IsControl(ch) {
		if f.minValue >= 0 && ch == '-' {
			Beep()
			return false
		}
		if text := strings.TrimSpace(string(f.RunesIfPasted([]rune{ch}))); text != "-" && text != "+" {
			if _, err := f.extract(text); err != nil || f.wouldExceedMax(string(ch)) {
				Beep()
				return false
			}
		}
	}
	return f.DefaultRuneTyped(ch)
}

// refuseAboveMax has the field refuse, with a beep, typed or pasted text that would put its value above the maximum,
// rather than show it as invalid, for a maximum that is a hard limit, such as the quantity of an item carried. Pasting
// is refused both by the Paste command and by its keyboard shortcut, which unison.Field.DefaultKeyDown handles itself
// for a field with no menu bar to route it through, such as one in a modal dialog.
func (f *NumericField[T]) refuseAboveMax() {
	f.limitToMax = true
	f.InstallCmdHandlers(unison.PasteItemID, func(_ any) bool { return f.CanPaste() }, func(_ any) {
		if f.wouldExceedMax(unison.ClipboardGetText()) {
			Beep()
			return
		}
		f.Paste()
	})
	keyDown := f.KeyDownCallback
	f.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, repeat bool) bool {
		if keyCode == unison.KeyV && mods&mod.NonSticky == mod.OSMenuCommand() && f.CanPaste() &&
			f.wouldExceedMax(unison.ClipboardGetText()) {
			Beep()
			return true
		}
		return keyDown(keyCode, mods, repeat)
	}
}

// wouldExceedMax returns true if the field refuses values above its maximum and inserting the text in place of the
// selection would put its value above it.
func (f *NumericField[T]) wouldExceedMax(text string) bool {
	if !f.limitToMax {
		return false
	}
	v, err := f.extract(strings.TrimSpace(string(f.RunesIfPasted([]rune(text)))))
	if err != nil || (f.hasException && v == f.exception) {
		return false
	}
	return v > f.maxValue
}

func (f *NumericField[T]) setWithoutUndo(state *unison.FieldState, focus bool) {
	// Whatever is about to show is exact text meant to be parsed -- it came from the user, from an undo, or from the
	// model by way of getData, never from DisplayFormat.
	f.showingDisplayText = false
	f.undoableField.setWithoutUndo(state, focus)
}

// Sync the field to the current value. While the field has the focus, the text in it is what the user is working on, so
// it is re-parsed rather than replaced, unless the field is disabled, in which case it shows the value's exact text;
// see undoableField.Sync.
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
