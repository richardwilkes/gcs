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
		parseOptionalDecimal, formatOptionalDecimal)
	f.HAlign = align.End
	f.RuneTypedCallback = f.runeTyped
	f.ValidateCallback = f.validate
	f.SetMinimumTextWidthUsing("-9999", "99999")
	f.Sync()
	return f
}

// parseOptionalDecimal reads the field's text. Text that is not a number is treated as none, as a number field treats
// it as zero; the field flags it as invalid meanwhile.
func parseOptionalDecimal(text string) optionalDecimal {
	value, err := fxp.FromString(strings.TrimSpace(text))
	if err != nil {
		return optionalDecimal{}
	}
	return optionalDecimal{value: value, present: true}
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

func (f *OptionalDecimalField) tooltipTextForValidation() string {
	text := strings.TrimSpace(f.Text())
	if text == "" {
		return f.problem(nil)
	}
	value, err := fxp.FromString(text)
	if err != nil || text == "-" || text == "+" {
		return i18n.Text("Invalid value")
	}
	return f.problem(&value)
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
