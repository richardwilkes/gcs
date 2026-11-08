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
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/align"
)

// optionalDecimal is the value an OptionalDecimalField edits: a fixed-point number, or none at all. It stands in for
// the *fxp.Int the model holds, since an undoable field compares values to tell whether they changed, and two pointers
// to the same number are not equal.
type optionalDecimal struct {
	value   fxp.Int
	present bool
}

func optionalDecimalOf(value *fxp.Int) optionalDecimal {
	if value == nil {
		return optionalDecimal{}
	}
	return optionalDecimal{value: *value, present: true}
}

func (d optionalDecimal) pointer() *fxp.Int {
	if !d.present {
		return nil
	}
	return &d.value
}

// OptionalDecimalField is a field that holds a fixed-point number which may be left blank, meaning there is none.
type OptionalDecimalField struct {
	undoableField[optionalDecimal]
	// Problem, if not nil, returns the reason a value that parses is still not acceptable, or an empty string if it is.
	Problem           func(value *fxp.Int) string
	validationTooltip *unison.Panel
	savedTooltip      *unison.Panel
}

// NewOptionalDecimalField creates a new field that holds a fixed-point number which may be left blank.
func NewOptionalDecimalField(targetMgr *TargetMgr, targetKey, undoTitle string, get func() *fxp.Int, set func(*fxp.Int)) *OptionalDecimalField {
	f := &OptionalDecimalField{}
	f.init(f, unison.NewField(), targetMgr, targetKey, undoTitle,
		func() optionalDecimal { return optionalDecimalOf(get()) },
		func(value optionalDecimal) { set(value.pointer()) },
		f.mustParse, formatOptionalDecimal)
	// An edit made over text the field can't parse records the text of the value the field holds in its place, since
	// that text leaves the value as it is (see mustParse): undoing the edit would otherwise put the text back but leave
	// the value the edit set.
	modified := f.ModifiedCallback
	f.ModifiedCallback = func(before, after *unison.FieldState) {
		if _, ok := parseOptionalDecimal(before.Text); !ok {
			exact := *before
			exact.Text = formatOptionalDecimal(f.last)
			before = &exact
		}
		modified(before, after)
	}
	f.HAlign = align.End
	f.RuneTypedCallback = f.runeTyped
	f.ValidateCallback = f.validate
	f.SetMinimumTextWidthUsing("-9999", "99999")
	f.Sync()
	return f
}

// mustParse returns the value the text holds. Text the field can't parse, which shows as invalid, holds no value, so it
// is taken for the value the field holds: the value is left as it is, and losing the focus puts its text back.
func (f *OptionalDecimalField) mustParse(text string) optionalDecimal {
	if value, ok := parseOptionalDecimal(text); ok {
		return value
	}
	return f.last
}

// parseOptionalDecimal returns the value the text holds, which is none for blank text, and whether the field can parse
// the text at all. Text with no digit, such as a sign or a decimal point by itself, which fxp.FromString takes for
// zero, is only the start of a number.
func parseOptionalDecimal(text string) (optionalDecimal, bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return optionalDecimal{}, true
	}
	value, err := fxp.FromString(text)
	if err != nil || !strings.ContainsAny(text, "0123456789") {
		return optionalDecimal{}, false
	}
	return optionalDecimal{value: value, present: true}, true
}

func formatOptionalDecimal(value optionalDecimal) string {
	if !value.present {
		return ""
	}
	return value.value.String()
}

func (f *OptionalDecimalField) runeTyped(ch rune) bool {
	if !unicode.IsControl(ch) {
		if text := strings.TrimSpace(string(f.RunesIfPasted([]rune{ch}))); text != "" && text != "-" && text != "+" {
			if _, err := fxp.FromString(text); err != nil {
				Beep()
				return false
			}
		}
	}
	return f.DefaultRuneTyped(ch)
}

// PerformAccessibilityAction implements unison.AccessibilityActor. Without the focus, text that would leave the value
// as it is changes nothing, since the field is synced to show its value only once the value changes, and text the
// field can't parse, which leaves the value as it is too (see mustParse), is refused rather than left showing as
// invalid, with no loss of the focus to come and put the value's text back.
func (f *OptionalDecimalField) PerformAccessibilityAction(req accessibility.ActionRequest) bool {
	if !f.hasFocus && (req.Action == accessibility.SetValue || req.Action == accessibility.ReplaceText) {
		text := req.Value
		if req.Action == accessibility.ReplaceText {
			runes := []rune(f.Text())
			start := min(max(req.Start, 0), len(runes))
			end := min(max(req.End, start), len(runes))
			text = string(runes[:start]) + req.Value + string(runes[end:])
		}
		value, ok := parseOptionalDecimal(text)
		if !ok {
			return false
		}
		if value == f.last {
			return true
		}
	}
	return f.Field.PerformAccessibilityAction(req)
}

func (f *OptionalDecimalField) tooltipTextForValidation() string {
	value, ok := parseOptionalDecimal(f.Text())
	if !ok {
		return i18n.Text("Invalid value")
	}
	return f.problem(value.pointer())
}

func (f *OptionalDecimalField) problem(value *fxp.Int) string {
	if f.Problem == nil {
		return ""
	}
	return f.Problem(value)
}

func (f *OptionalDecimalField) validate() bool {
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
